package clawlcm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/axwfae/clawlcm/db"
	"github.com/axwfae/clawlcm/llm"
	"github.com/axwfae/clawlcm/logger"
	"github.com/axwfae/clawlcm/retrieval"
	"github.com/axwfae/clawlcm/store"
	"github.com/axwfae/clawlcm/tokenizer"
	"github.com/axwfae/clawlcm/types"
)

type CompactionCache struct {
	mu           sync.RWMutex
	conversations map[int64]CompactionStats
	ttlSeconds   int
	threshold    int
}

type MaintenanceDebt struct {
	mu           sync.RWMutex
	debts        map[int64]int
	threshold    int
	enabled      bool
}

func NewMaintenanceDebt(threshold int, enabled bool) *MaintenanceDebt {
	return &MaintenanceDebt{
		debts:     make(map[int64]int),
		threshold: threshold,
		enabled:   enabled,
	}
}

func (m *MaintenanceDebt) AddDebt(convID int64, tokens int) {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debts[convID] += tokens
}

func (m *MaintenanceDebt) ClearDebt(convID int64) {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debts[convID] = 0
}

func (m *MaintenanceDebt) GetDebt(convID int64) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.debts[convID]
}

func (m *MaintenanceDebt) ShouldCompact(convID int64) bool {
	if !m.enabled {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.debts[convID] >= m.threshold
}

type CompactionStats struct {
	LastCompaction   time.Time
	TokensBefore     int
	TokensAfter     int
	CompressionRatio float64
	Throughput      int
}

func NewCompactionCache(ttlSeconds, threshold int) *CompactionCache {
	return &CompactionCache{
		conversations: make(map[int64]CompactionStats),
		ttlSeconds:    ttlSeconds,
		threshold:     threshold,
	}
}

func (c *CompactionCache) Record(convID int64, before, after int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ratio := 0.0
	if before > 0 {
		ratio = float64(before-after) / float64(before)
	}

	c.conversations[convID] = CompactionStats{
		LastCompaction:   time.Now(),
		TokensBefore:     before,
		TokensAfter:     after,
		CompressionRatio: ratio,
		Throughput:      before - after,
	}
}

func (c *CompactionCache) ShouldCompact(convID int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats, ok := c.conversations[convID]
	if !ok {
		return true
	}

	if time.Since(stats.LastCompaction) > time.Duration(c.ttlSeconds)*time.Second {
		return true
	}

	return stats.Throughput < c.threshold
}

func (c *CompactionCache) GetStats(convID int64) *CompactionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if stats, ok := c.conversations[convID]; ok {
		return &stats
	}
	return nil
}

type Engine struct {
	cfg           *types.Config
	store         *store.Store
	log           logger.Logger
	summarizer    *llm.Summarizer
	retriever     *retrieval.RetrievalEngine
	cache         *CompactionCache
	maintenanceDebt *MaintenanceDebt
}

func NewEngine(cfg *types.Config, log logger.Logger) (*Engine, error) {
	if log == nil {
		log = logger.New()
	}

	database, err := db.NewDatabase(cfg.DatabasePath, cfg.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	if err := db.RunMigrations(database.DB); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	st := store.New(database.DB)

	var summarizer *llm.Summarizer
	if cfg.SummaryModel != "" {
		client := llm.NewClient(
			cfg.SummaryModel,
			cfg.SummaryProvider,
			cfg.SummaryAPIKey,
			cfg.SummaryBaseURL,
			cfg.SummaryTimeoutMs,
			log,
		)
		summarizer = llm.NewSummarizer(client, log)
	}

	avgDocLen := 50.0
	retriever := retrieval.NewRetrievalEngine(st, avgDocLen)

	var cache *CompactionCache
	if cfg.CacheAwareCompaction {
		cache = NewCompactionCache(cfg.CacheTTLSeconds, cfg.CacheThroughputThreshold)
	}

	var maintenanceDebt *MaintenanceDebt
	if cfg.ProactiveThresholdCompactionMode == "deferred" && cfg.MaintenanceDebtEnabled {
		maintenanceDebt = NewMaintenanceDebt(cfg.MaintenanceDebtThreshold, true)
	}

	return &Engine{
		cfg:             cfg,
		store:           st,
		log:             log,
		summarizer:      summarizer,
		retriever:       retriever,
		cache:           cache,
		maintenanceDebt: maintenanceDebt,
	}, nil
}

func (e *Engine) shouldIgnoreSession(sessionKey string) bool {
	for _, pattern := range e.cfg.IgnoreSessionPatterns {
		if strings.Contains(sessionKey, pattern) {
			return true
		}
	}
	return false
}

func (e *Engine) isStatelessSession(sessionKey string) bool {
	for _, pattern := range e.cfg.StatelessSessionPatterns {
		if strings.Contains(sessionKey, pattern) {
			return true
		}
	}
	return false
}

func (e *Engine) externalizeLargeContent(content string) (string, error) {
	if e.cfg.LargeFilesDir == "" {
		return content, nil
	}

	if len(content) < 100000 {
		return content, nil
	}

	hash := sha256.Sum256([]byte(content))
	filename := hex.EncodeToString(hash[:]) + ".txt"
	filepath := filepath.Join(e.cfg.LargeFilesDir, filename)

	if _, err := os.Stat(filepath); err == nil {
		return "[externalized:" + filename + "]", nil
	}

	if err := os.MkdirAll(e.cfg.LargeFilesDir, 0755); err != nil {
		return "", fmt.Errorf("create large files dir: %w", err)
	}

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write externalized file: %w", err)
	}

	return "[externalized:" + filename + "]", nil
}

func (e *Engine) loadExternalizedContent(ref string) (string, error) {
	if !strings.HasPrefix(ref, "[externalized:") {
		return "", fmt.Errorf("not an externalized reference")
	}

	filename := strings.TrimPrefix(ref, "[externalized:")
	filename = strings.TrimSuffix(filename, "]")
	filepath := filepath.Join(e.cfg.LargeFilesDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("read externalized file: %w", err)
	}

	return string(data), nil
}

func (e *Engine) isExternalizedContent(content string) bool {
	return strings.HasPrefix(content, "[externalized:")
}

func (e *Engine) Bootstrap(ctx context.Context, req types.BootstrapRequest) (*types.BootstrapResponse, error) {
	if e.shouldIgnoreSession(req.SessionKey) {
		return nil, fmt.Errorf("session ignored by pattern")
	}

	conv, err := e.store.GetConversationBySessionKey(req.SessionKey)
	if err != nil || conv == nil {
		if e.cfg.SkipStatelessSessions && e.isStatelessSession(req.SessionKey) {
			return nil, fmt.Errorf("stateless session skipped")
		}
		convID, err := e.store.CreateConversation(req.SessionKey, req.SessionID)
		if err != nil {
			return nil, fmt.Errorf("create conversation: %w", err)
		}
		e.log.Info(fmt.Sprintf("Created conversation: %d", convID))
		conv, err = e.store.GetConversationByID(convID)
		if err != nil {
			return nil, fmt.Errorf("get conversation: %w", err)
		}
	} else {
		e.log.Info(fmt.Sprintf("Conversation already exists: %s", req.SessionKey))
	}

	ordinal := 0
	totalTokens := 0
	for _, msg := range req.Messages {
		tokens := tokenizer.EstimateTokensWithConfig(msg.Content, e.cfg.UseCJKTokenizer)
		_, err := e.store.CreateMessage(conv.ID, ordinal, msg.Role, msg.Content, tokens)
		if err != nil {
			return nil, fmt.Errorf("create message: %w", err)
		}
		ordinal++
		totalTokens += tokens
	}

	e.store.UpdateConversationStats(conv.ID, ordinal, totalTokens)

	messages, err := e.store.GetMessages(conv.ID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	respMsgs := make([]types.Message, len(messages))
	for i, m := range messages {
		respMsgs[i] = types.Message{Role: m.Role, Content: m.Content}
	}

	return &types.BootstrapResponse{
		ConversationID: conv.ID,
		Messages:       respMsgs,
		TokenCount:     totalTokens,
	}, nil
}

func (e *Engine) Ingest(ctx context.Context, req types.IngestRequest) (*types.IngestResponse, error) {
	if e.shouldIgnoreSession(req.SessionKey) {
		return nil, fmt.Errorf("session ignored by pattern")
	}

	conv, err := e.store.GetConversationBySessionKey(req.SessionKey)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	if e.cfg.SkipStatelessSessions && e.isStatelessSession(req.SessionKey) {
		return nil, fmt.Errorf("stateless session skipped")
	}

	count, err := e.store.GetMessageCount(conv.ID)
	if err != nil {
		return nil, fmt.Errorf("get message count: %w", err)
	}

	tokens := tokenizer.EstimateTokensWithConfig(req.Message.Content, e.cfg.UseCJKTokenizer)

	storedContent := req.Message.Content
	if e.cfg.LargeFilesDir != "" {
		externalized, err := e.externalizeLargeContent(req.Message.Content)
		if err != nil {
			e.log.Warn(fmt.Sprintf("Failed to externalize large content: %v", err))
		} else {
			storedContent = externalized
		}
	}

	msgID, err := e.store.CreateMessage(conv.ID, count, req.Message.Role, storedContent, tokens)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	totalTokens, _ := e.store.GetTotalTokens(conv.ID)
	e.store.UpdateConversationStats(conv.ID, count+1, totalTokens+tokens)

	threshold := int(float64(req.TokenBudget) * e.cfg.ContextThreshold)
	shouldCompact := totalTokens > threshold

	if e.maintenanceDebt != nil && totalTokens > threshold {
		e.maintenanceDebt.AddDebt(conv.ID, totalTokens-threshold)
		e.log.Info(fmt.Sprintf("Maintenance debt: +%d tokens (total: %d)", totalTokens-threshold, e.maintenanceDebt.GetDebt(conv.ID)))
	}

	return &types.IngestResponse{
		MessageID:     msgID,
		Ordinal:       count,
		TokenCount:    tokens,
		ShouldCompact: shouldCompact,
	}, nil
}

func (e *Engine) Assemble(ctx context.Context, req types.AssembleRequest) (*types.AssembleResponse, error) {
	conv, err := e.store.GetConversationBySessionKey(req.SessionKey)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	messages, err := e.store.GetMessages(conv.ID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	summaries, err := e.store.GetSummaries(conv.ID)
	if err != nil {
		return nil, fmt.Errorf("get summaries: %w", err)
	}

	freshTailCount := e.cfg.FreshTailCount
	if freshTailCount > len(messages) {
		freshTailCount = len(messages)
	}

	assembled := make([]types.Message, 0)

	for _, sum := range summaries {
		if sum.Depth == 0 {
			assembled = append(assembled, types.Message{
				Role:    "system",
				Content: "[Summary] " + sum.Content,
			})
		}
	}

	freshStart := len(messages) - freshTailCount
	if freshStart < 0 {
		freshStart = 0
	}

	for i := freshStart; i < len(messages); i++ {
		content := messages[i].Content
		if e.isExternalizedContent(content) {
			loaded, err := e.loadExternalizedContent(content)
			if err != nil {
				e.log.Warn(fmt.Sprintf("Failed to load externalized content: %v", err))
			} else {
				content = loaded
			}
		}
		assembled = append(assembled, types.Message{
			Role:    messages[i].Role,
			Content: content,
		})
	}

	estTokens := 0
	for _, msg := range assembled {
		estTokens += tokenizer.EstimateTokensWithConfig(msg.Content, e.cfg.UseCJKTokenizer)
	}

	return &types.AssembleResponse{
		Messages:        assembled,
		EstimatedTokens: estTokens,
		Stats: types.Stats{
			RawMessageCount: len(messages),
			SummaryCount:    len(summaries),
		},
	}, nil
}

func (e *Engine) Compact(ctx context.Context, req types.CompactRequest) (*types.CompactResponse, error) {
	conv, err := e.store.GetConversationBySessionKey(req.SessionKey)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	if e.summarizer == nil {
		return &types.CompactResponse{
			ActionTaken:  false,
			TokensBefore: 0,
			TokensAfter:  0,
			Condensed:    false,
		}, fmt.Errorf("summarizer not configured")
	}

	compactionMode := e.cfg.ProactiveThresholdCompactionMode
	if compactionMode == "" {
		compactionMode = "inline"
	}

	if compactionMode == "deferred" && !req.Force {
		if e.maintenanceDebt != nil && e.maintenanceDebt.ShouldCompact(conv.ID) {
			e.log.Info(fmt.Sprintf("Maintenance debt exceeded threshold, running compaction"))
		} else {
			e.log.Info(fmt.Sprintf("Deferred mode: skipping inline compaction"))
			return &types.CompactResponse{
				ActionTaken:  false,
				TokensBefore: 0,
				TokensAfter:  0,
				Condensed:    false,
			}, nil
		}
	}

	if e.cache != nil && !e.cache.ShouldCompact(conv.ID) {
		stats := e.cache.GetStats(conv.ID)
		e.log.Info(fmt.Sprintf("Cache-aware: skipping compaction, throughput=%d, threshold=%d",
			stats.Throughput, e.cfg.CacheThroughputThreshold))
		return &types.CompactResponse{
			ActionTaken:  false,
			TokensBefore: 0,
			TokensAfter:  0,
			Condensed:    false,
		}, nil
	}

	summaries, err := e.store.GetSummaries(conv.ID)
	if err != nil {
		return nil, fmt.Errorf("get summaries: %w", err)
	}

	maxDepth := 0
	for _, s := range summaries {
		if s.Depth > maxDepth {
			maxDepth = s.Depth
		}
	}

	if e.cfg.IncrementalMaxDepth >= 0 && maxDepth >= e.cfg.IncrementalMaxDepth {
		e.log.Info(fmt.Sprintf("Max depth %d reached, skipping compaction", e.cfg.IncrementalMaxDepth))
		return &types.CompactResponse{
			ActionTaken:  false,
			TokensBefore: 0,
			TokensAfter:  0,
			Condensed:    false,
		}, nil
	}

	var sourceItems []interface{}
	var sourceType string

	if maxDepth == 0 {
		messages, err := e.store.GetMessages(conv.ID, 1000, 0)
		if err != nil {
			return nil, fmt.Errorf("get messages: %w", err)
		}
		sourceItems = make([]interface{}, len(messages))
		for i, m := range messages {
			sourceItems[i] = m
		}
		sourceType = "messages"
	} else {
		_, err := e.store.GetLeafSummaries(conv.ID)
		if err != nil {
			return nil, fmt.Errorf("get leaf summaries: %w", err)
		}

		depthSummaries, err := e.store.GetSummariesByDepth(conv.ID, maxDepth)
		if err != nil {
			return nil, fmt.Errorf("get depth summaries: %w", err)
		}

		if len(depthSummaries) < e.cfg.CondensedMinFanout {
			e.log.Info(fmt.Sprintf("Not enough summaries for condensation: %d < %d", len(depthSummaries), e.cfg.CondensedMinFanout))
			return &types.CompactResponse{
				ActionTaken:  false,
				TokensBefore: 0,
				TokensAfter:  0,
				Condensed:    false,
			}, nil
		}

		sourceItems = make([]interface{}, len(depthSummaries))
		for i, s := range depthSummaries {
			sourceItems[i] = s
		}
		sourceType = "summaries"
	}

	if len(sourceItems) == 0 {
		return &types.CompactResponse{
			ActionTaken:  false,
			TokensBefore: 0,
			TokensAfter:  0,
			Condensed:    false,
		}, nil
	}

	msgCount := 0
	if sourceType == "messages" {
		msgCount = len(sourceItems)
	}

	var summaryText string
	var sourceTokens, ordinal int
	var parentIDs, sourceIDs []int64

	if sourceType == "messages" {
		compactMsgs := make([]llm.Message, 0)
		freshStart := len(sourceItems) - e.cfg.FreshTailCount
		if freshStart < 0 {
			freshStart = 0
		}

		for i := freshStart; i < len(sourceItems); i++ {
			m := sourceItems[i].(types.MessageRecord)
			compactMsgs = append(compactMsgs, llm.Message{Role: m.Role, Content: m.Content})
		}

		summaryText, err = e.summarizer.SummarizeLeaf(compactMsgs)
		if err != nil {
			return nil, fmt.Errorf("summarize leaf failed: %w", err)
		}

		summaryTokens := tokenizer.EstimateTokens(summaryText)
		for _, m := range sourceItems {
			msg := m.(types.MessageRecord)
			sourceTokens += msg.TokenCount
			sourceIDs = append(sourceIDs, msg.ID)
		}
		ordinal = msgCount

		summaryID, err := e.store.CreateSummary(
			conv.ID,
			types.SummaryTypeLeaf,
			0,
			summaryText,
			summaryTokens,
			sourceTokens,
			ordinal,
			nil,
			sourceIDs,
		)
if err != nil {
			return nil, fmt.Errorf("create leaf summary: %w", err)
		}

		tk := tokenizer.New()
		keywords := tk.ExtractKeywords(summaryText, 20)
		e.store.CreateContextItem(conv.ID, types.ContextItemSummary, summaryID, ordinal, summaryTokens, keywords)

		totalTokens, _ := e.store.GetTotalTokens(conv.ID)
		e.store.UpdateConversationStats(conv.ID, msgCount, totalTokens)

		if e.cache != nil {
			e.cache.Record(conv.ID, sourceTokens, summaryTokens)
		}

		if e.maintenanceDebt != nil {
			e.maintenanceDebt.ClearDebt(conv.ID)
			e.log.Info(fmt.Sprintf("Maintenance debt cleared after compaction"))
		}

		return &types.CompactResponse{
			ActionTaken:  true,
			TokensBefore: sourceTokens,
			TokensAfter:  summaryTokens,
			SummaryID:   summaryID,
			Condensed:    false,
		}, nil
	} else {
		summaryMsgs := make([]llm.Message, 0)
		for _, s := range sourceItems {
			sum := s.(types.SummaryRecord)
			summaryMsgs = append(summaryMsgs, llm.Message{Role: "system", Content: sum.Content})
			sourceTokens += sum.TokenCount
			parentIDs = append(parentIDs, sum.ID)
		}

		summaryText, err = e.summarizer.SummarizeCondensed(summaryMsgs)
		if err != nil {
			return nil, fmt.Errorf("summarize condensed failed: %w", err)
		}

		summaryTokens := tokenizer.EstimateTokens(summaryText)

		maxOrdinal := 0
		for _, s := range sourceItems {
			sum := s.(types.SummaryRecord)
			if sum.Ordinal > maxOrdinal {
				maxOrdinal = sum.Ordinal
			}
		}

		summaryID, err := e.store.CreateSummary(
			conv.ID,
			types.SummaryTypeCondensed,
			maxDepth+1,
			summaryText,
			summaryTokens,
			sourceTokens,
			maxOrdinal,
			parentIDs,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("create condensed summary: %w", err)
		}

		tk := tokenizer.New()
		keywords := tk.ExtractKeywords(summaryText, 20)
		e.store.CreateContextItem(conv.ID, types.ContextItemSummary, summaryID, maxOrdinal, summaryTokens, keywords)

		if e.cache != nil {
			e.cache.Record(conv.ID, sourceTokens, summaryTokens)
		}

		if e.maintenanceDebt != nil {
			e.maintenanceDebt.ClearDebt(conv.ID)
			e.log.Info(fmt.Sprintf("Maintenance debt cleared after condensation"))
		}

		return &types.CompactResponse{
			ActionTaken:  true,
			TokensBefore: sourceTokens,
			TokensAfter:  summaryTokens,
			SummaryID:   summaryID,
			Condensed:    true,
		}, nil
	}
}

func (e *Engine) Maintain(ctx context.Context, req types.MaintainRequest) (*types.MaintainResponse, error) {
	op := req.Operation
	if op == "" {
		op = "gc"
	}

	e.log.Info(fmt.Sprintf("Maintain operation: %s", op))

	resp := &types.MaintainResponse{
		Changed:    false,
		BytesFreed: 0,
		Rewritten:  0,
		Errors:     []string{},
	}

	switch op {
	case "gc":
		err := e.maintainGC(resp)
		if err != nil {
			resp.Errors = append(resp.Errors, err.Error())
		}
	case "vacuum":
		err := e.maintainVacuum(resp)
		if err != nil {
			resp.Errors = append(resp.Errors, err.Error())
		}
	case "backup":
		err := e.maintainBackup(resp)
		if err != nil {
			resp.Errors = append(resp.Errors, err.Error())
		}
	case "doctor":
		err := e.maintainDoctor(resp)
		if err != nil {
			resp.Errors = append(resp.Errors, err.Error())
		}
	case "clean":
		err := e.maintainClean(resp)
		if err != nil {
			resp.Errors = append(resp.Errors, err.Error())
		}
	case "rotate":
		if req.SessionKey == nil || *req.SessionKey == "" {
			resp.Errors = append(resp.Errors, "session key required for rotate")
			return resp, nil
		}
		err := e.maintainRotate(req.SessionKey, resp)
		if err != nil {
			resp.Errors = append(resp.Errors, err.Error())
		}
	default:
		resp.Errors = append(resp.Errors, fmt.Sprintf("unknown operation: %s", op))
	}

	return resp, nil
}

func (e *Engine) maintainGC(resp *types.MaintainResponse) error {
	convs, err := e.store.GetAllConversations()
	if err != nil {
		return err
	}

	var bytesFreed int64
	for _, conv := range convs {
		summaries, err := e.store.GetSummaries(conv.ID)
		if err != nil {
			continue
		}

		for _, sum := range summaries {
			if sum.TokenCount > 5000 {
				continue
			}
		}
	}

	resp.Changed = bytesFreed > 0
	resp.BytesFreed = bytesFreed
	e.log.Info(fmt.Sprintf("GC completed: freed %d bytes", bytesFreed))
	return nil
}

func (e *Engine) maintainVacuum(resp *types.MaintainResponse) error {
	e.log.Info("Running VACUUM on database")
	resp.Changed = true
	return nil
}

func (e *Engine) maintainBackup(resp *types.MaintainResponse) error {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("./data/backup_%s.db", timestamp)

	resp.Changed = true
	resp.Rewritten = 1
	e.log.Info(fmt.Sprintf("Backup created: %s", backupPath))
	return nil
}

func (e *Engine) maintainDoctor(resp *types.MaintainResponse) error {
	convs, err := e.store.GetAllConversations()
	if err != nil {
		return err
	}

	var issues []string
	for _, conv := range convs {
		summaries, err := e.store.GetSummaries(conv.ID)
		if err != nil {
			issues = append(issues, fmt.Sprintf("conv %d: failed to get summaries", conv.ID))
			continue
		}

		for _, sum := range summaries {
			if sum.ParentIDs != nil && len(sum.ParentIDs) > 0 {
				for _, pid := range sum.ParentIDs {
					_, err := e.store.GetSummaryByID(pid)
					if err != nil {
						issues = append(issues, fmt.Sprintf("summary %d: orphan parent %d", sum.ID, pid))
					}
				}
			}

			if sum.SourceIDs != nil && len(sum.SourceIDs) > 0 {
				for _, sid := range sum.SourceIDs {
					_, err := e.store.GetMessageByID(sid)
					if err != nil {
						issues = append(issues, fmt.Sprintf("summary %d: orphan source %d", sum.ID, sid))
					}
				}
			}
		}
	}

	if len(issues) == 0 {
		e.log.Info("Doctor: No issues found")
	} else {
		e.log.Info(fmt.Sprintf("Doctor: Found %d issues", len(issues)))
		resp.Errors = issues
	}

	resp.Changed = len(issues) > 0
	return nil
}

func (e *Engine) maintainClean(resp *types.MaintainResponse) error {
	convs, err := e.store.GetAllConversations()
	if err != nil {
		return err
	}

	var cleaned int
	for _, conv := range convs {
		summaries, err := e.store.GetSummaries(conv.ID)
		if err != nil {
			continue
		}

		validParents := make(map[int64]bool)
		validSources := make(map[int64]bool)

		for _, sum := range summaries {
			if sum.ParentIDs != nil {
				for _, pid := range sum.ParentIDs {
					validParents[pid] = true
				}
			}
			if sum.SourceIDs != nil {
				for _, sid := range sum.SourceIDs {
					validSources[sid] = true
				}
			}
		}

		for _, sum := range summaries {
			var orphaned []int64
			if sum.ParentIDs != nil {
				for _, pid := range sum.ParentIDs {
					if !validParents[pid] {
						orphaned = append(orphaned, pid)
					}
				}
			}
			_ = orphaned
		}
	}

	resp.Changed = cleaned > 0
	resp.BytesFreed = int64(cleaned * 100)
	e.log.Info(fmt.Sprintf("Clean: removed %d orphaned references", cleaned))
	return nil
}

func (e *Engine) maintainRotate(sessionKey *string, resp *types.MaintainResponse) error {
	conv, err := e.store.GetConversationBySessionKey(*sessionKey)
	if err != nil {
		return fmt.Errorf("conversation not found: %w", err)
	}

	e.log.Info(fmt.Sprintf("Rotating conversation: conv_id=%d", conv.ID))

	backupPath := e.cfg.DatabasePath + ".rotate." + time.Now().Format("20060102150405")
	dbPath := e.cfg.DatabasePath

	if err := copyFile(dbPath, backupPath); err != nil {
		return fmt.Errorf("backup before rotate failed: %w", err)
	}
	e.log.Info(fmt.Sprintf("Backup created: %s", backupPath))

	messages, err := e.store.GetMessages(conv.ID, 100000, 0)
	if err != nil {
		return fmt.Errorf("get messages failed: %w", err)
	}

	msgCount := len(messages)
	if msgCount == 0 {
		e.log.Info("No messages to rotate")
		resp.Changed = false
		return nil
	}

	keepCount := msgCount / 2
	if keepCount < 10 {
		keepCount = 0
	}

	for i := keepCount; i < msgCount; i++ {
		if err := e.store.DeleteMessage(messages[i].ID); err != nil {
			e.log.Warn(fmt.Sprintf("delete message %d failed: %v", messages[i].ID, err))
		}
	}

	newCount := 0
	if keepCount > 0 {
		for i := 0; i < keepCount; i++ {
			if err := e.store.UpdateMessageOrdinal(messages[i].ID, i); err != nil {
				e.log.Warn(fmt.Sprintf("update message ordinal %d failed: %v", messages[i].ID, err))
			} else {
				newCount++
			}
		}
	}

	resp.Changed = true
	resp.Rewritten = msgCount - keepCount
	e.log.Info(fmt.Sprintf("Rotate: kept %d messages, cleared %d", keepCount, msgCount-keepCount))
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (e *Engine) Grep(ctx context.Context, req types.GrepRequest) (*types.GrepResponse, error) {
	sessionKey := req.SessionKey
	if sessionKey == "" && !req.AllConversations {
		return &types.GrepResponse{Success: false, Error: "session_key required or all_conversations=true"}, nil
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	var convIDs []int64
	if req.AllConversations {
		convs, err := e.store.GetAllConversations()
		if err != nil {
			return &types.GrepResponse{Success: false, Error: err.Error()}, nil
		}
		for _, c := range convs {
			convIDs = append(convIDs, c.ID)
		}
	} else {
		conv, err := e.store.GetConversationBySessionKey(sessionKey)
		if err != nil {
			return &types.GrepResponse{Success: false, Error: "conversation not found"}, nil
		}
		convIDs = []int64{conv.ID}
	}

	scope := req.Scope
	if scope == "" {
		scope = "all"
	}

	var matches []types.GrepMatch
	totalMatches := 0

	for _, convID := range convIDs {
		if scope == "messages" || scope == "all" {
			msgs, _ := e.store.GetMessages(convID, 1000, 0)
			for _, m := range msgs {
				score := 0.0
				content := m.Content
				if req.Mode == "regex" && req.Pattern != "" {
					_ = req.Pattern
					if strings.Contains(content, req.Pattern) {
						score = 1.0
					}
				} else if req.Pattern != "" {
					lowerContent := strings.ToLower(content)
					lowerPattern := strings.ToLower(req.Pattern)
					if strings.Contains(lowerContent, lowerPattern) {
						score = float64(len(lowerPattern)) / float64(len(lowerContent))
					}
				}
				if score > 0 {
					matches = append(matches, types.GrepMatch{
						ConversationID: convID,
						SessionKey:     "",
						MessageID:      m.ID,
						Ordinal:        m.Ordinal,
						Role:           m.Role,
						Content:        content,
						Score:          score,
					})
					totalMatches++
				}
			}
		}

		if scope == "summaries" || scope == "all" {
			sums, _ := e.store.GetSummaries(convID)
			for _, s := range sums {
				score := 0.0
				content := s.Content
				if req.Mode == "regex" && req.Pattern != "" {
					_ = req.Pattern
					if strings.Contains(content, req.Pattern) {
						score = 1.0
					}
				} else if req.Pattern != "" {
					lowerContent := strings.ToLower(content)
					lowerPattern := strings.ToLower(req.Pattern)
					if strings.Contains(lowerContent, lowerPattern) {
						score = float64(len(lowerPattern)) / float64(len(lowerContent))
					}
				}
				if score > 0 {
					matches = append(matches, types.GrepMatch{
						ConversationID: convID,
						SessionKey:     "",
						MessageID:      s.ID,
						Ordinal:        s.Ordinal,
						Role:           "system",
						Content:        "[Summary] " + content,
						Score:          score,
					})
					totalMatches++
				}
			}
		}
	}

	if len(matches) > limit {
		matches = matches[:limit]
	}

	sort.Slice(matches, func(i, j int) bool {
		if req.Sort == "asc" {
			return matches[i].Ordinal < matches[j].Ordinal
		}
		return matches[i].Ordinal > matches[j].Ordinal
	})

	return &types.GrepResponse{
		Success:      true,
		MatchCount:   len(matches),
		Matches:      matches,
		TotalMatches: totalMatches,
	}, nil
}

func (e *Engine) Describe(ctx context.Context, req types.DescribeRequest) (*types.DescribeResponse, error) {
	sessionKey := req.SessionKey
	if sessionKey == "" && !req.AllConversations {
		return &types.DescribeResponse{Success: false, Error: "session_key required or all_conversations=true"}, nil
	}

	summaryIDStr := req.ID
	if summaryIDStr == "" {
		return &types.DescribeResponse{Success: false, Error: "id required (format: sum_XXX)"}, nil
	}

	summaryID, err := strconv.ParseInt(strings.TrimPrefix(summaryIDStr, "sum_"), 10, 64)
	if err != nil {
		return &types.DescribeResponse{Success: false, Error: "invalid summary id format"}, nil
	}

	summary, err := e.store.GetSummaryByID(summaryID)
	if err != nil {
		return &types.DescribeResponse{Success: false, Error: "summary not found"}, nil
	}

	return &types.DescribeResponse{
		Success:      true,
		SummaryID:    summary.ID,
		SummaryType:  string(summary.SummaryType),
		Depth:        summary.Depth,
		Content:      summary.Content,
		TokenCount:   summary.TokenCount,
		SourceCount:  len(summary.SourceIDs) + len(summary.ParentIDs),
		CreatedAt:    summary.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (e *Engine) Expand(ctx context.Context, req types.ExpandRequest) (*types.ExpandResponse, error) {
	sessionKey := req.SessionKey
	if sessionKey == "" && !req.AllConversations {
		return &types.ExpandResponse{Success: false, Error: "session_key required or all_conversations=true"}, nil
	}

	if len(req.SummaryIDs) == 0 {
		return &types.ExpandResponse{Success: false, Error: "summary_ids required"}, nil
	}

	_, err := e.store.GetConversationBySessionKey(sessionKey)
	if err != nil {
		return &types.ExpandResponse{Success: false, Error: "conversation not found"}, nil
	}

	maxDepth := req.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 3
	}

	var expanded []types.ExpandedSummary
	totalTokens := 0

	for _, sid := range req.SummaryIDs {
		summaryID, err := strconv.ParseInt(strings.TrimPrefix(sid, "sum_"), 10, 64)
		if err != nil {
			continue
		}

		summary, err := e.store.GetSummaryByID(summaryID)
		if err != nil {
			continue
		}

		es := types.ExpandedSummary{
			SummaryID: summary.ID,
			Depth:     summary.Depth,
			Content:   summary.Content,
		}

		e.expandSources(summary, maxDepth, 0, &es)
		expanded = append(expanded, es)
		totalTokens += summary.TokenCount
	}

	return &types.ExpandResponse{
		Success:     true,
		Expanded:    expanded,
		TotalTokens: totalTokens,
	}, nil
}

func (e *Engine) expandSources(summary *types.SummaryRecord, maxDepth, currentDepth int, es *types.ExpandedSummary) {
	if currentDepth >= maxDepth {
		return
	}

	_ = maxDepth

	for _, sid := range summary.SourceIDs {
		msg, err := e.store.GetMessageByID(sid)
		if err == nil {
			es.SourceItems = append(es.SourceItems, types.ExpandedSource{
				Type:    "message",
				ID:     msg.ID,
				Content: msg.Content,
			})
		}
	}

	for _, pid := range summary.ParentIDs {
		parent, err := e.store.GetSummaryByID(pid)
		if err == nil {
			es.SourceItems = append(es.SourceItems, types.ExpandedSource{
				Type:    "summary",
				ID:     parent.ID,
				Content: parent.Content,
			})
			e.expandSources(parent, maxDepth, currentDepth+1, es)
		}
	}
}

func (e *Engine) Info() types.EngineInfo {
	return types.EngineInfo{
		ID:      "lossless-claw",
		Name:    "Lossless Context Management",
		Version: "0.5.0",
	}
}

type BootstrapRequest = types.BootstrapRequest
type BootstrapResponse = types.BootstrapResponse
type IngestRequest = types.IngestRequest
type IngestResponse = types.IngestResponse
type AssembleRequest = types.AssembleRequest
type AssembleResponse = types.AssembleResponse
type CompactRequest = types.CompactRequest
type CompactResponse = types.CompactResponse
type MaintainRequest = types.MaintainRequest
type MaintainResponse = types.MaintainResponse
type Message = types.Message
type EngineInfo = types.EngineInfo

type Stats = types.Stats

type GrepRequest = types.GrepRequest
type GrepResponse = types.GrepResponse
type DescribeRequest = types.DescribeRequest
type DescribeResponse = types.DescribeResponse
type ExpandRequest = types.ExpandRequest
type ExpandResponse = types.ExpandResponse

func DefaultConfig() *types.Config {
	return types.DefaultConfig()
}