package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/axwfae/clawlcm"
	"github.com/axwfae/clawlcm/logger"
)

var (
	version        = "0.8.2"
	commit         = "cli-fix"
	date           = time.Now().Format("2006-01-02")
	defaultCfgPath = "./data/config.json"
)

type ConfigFile struct {
	Database struct {
		Path string `json:"path"`
	} `json:"database"`
	LLM struct {
		Model     string `json:"model"`
		Provider  string `json:"provider"`
		APIKey    string `json:"apiKey"`
		BaseURL   string `json:"baseURL"`
		TimeoutMs int    `json:"timeoutMs"`
	} `json:"llm"`
	Context struct {
		Threshold                        float64 `json:"threshold"`
		FreshTailCount                   int     `json:"freshTailCount"`
		UseCJKTokenizer                  bool    `json:"useCJKTokenizer"`
		CondensedMinFanout               int     `json:"condensedMinFanout"`
		IncrementalMaxDepth              int     `json:"incrementalMaxDepth"`
		ProactiveThresholdCompactionMode string  `json:"proactiveThresholdCompactionMode"`
		MaintenanceDebtEnabled           bool    `json:"maintenanceDebtEnabled"`
		MaintenanceDebtThreshold         int     `json:"maintenanceDebtThreshold"`
		LargeFilesDir                    string  `json:"largeFilesDir"`
		CacheAwareCompaction             bool    `json:"cacheAwareCompaction"`
	} `json:"context"`
	Session struct {
		IgnoreSessionPatterns    []string `json:"ignoreSessionPatterns"`
		StatelessSessionPatterns []string `json:"statelessSessionPatterns"`
		SkipStatelessSessions    bool     `json:"skipStatelessSessions"`
	} `json:"session"`
	Enabled bool `json:"enabled"`
	Verbose bool `json:"verbose"`
}

func createDefaultConfig(path string) error {
	defaultConfig := ConfigFile{}
	defaultConfig.Database.Path = "./data/clawlcm.db"
	defaultConfig.LLM.Provider = "openai"
	defaultConfig.LLM.TimeoutMs = 120000
	defaultConfig.Context.Threshold = 0.75
	defaultConfig.Context.FreshTailCount = 8
	defaultConfig.Context.UseCJKTokenizer = true
	defaultConfig.Context.CondensedMinFanout = 4
	defaultConfig.Context.IncrementalMaxDepth = 1
	defaultConfig.Context.ProactiveThresholdCompactionMode = "deferred"
	defaultConfig.Context.MaintenanceDebtEnabled = true
	defaultConfig.Context.MaintenanceDebtThreshold = 50000
	defaultConfig.Context.LargeFilesDir = "./data/large_files"
	defaultConfig.Context.CacheAwareCompaction = false
	defaultConfig.Enabled = true
	defaultConfig.Verbose = false

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func loadConfig(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create config directory: %v", err)
			}
			if err := createDefaultConfig(path); err != nil {
				return nil, fmt.Errorf("failed to create default config: %v", err)
			}
			data, err = os.ReadFile(path)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var config ConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	dbPath := flag.String("db", "", "Database path (overrides config)")
	configFile := flag.String("config", "", "Config file path")
	enable := flag.Bool("enable", true, "Enable LCM")
	verbose := flag.Bool("v", false, "Verbose output")
	showVersion := flag.Bool("version", false, "Show version")

	llmModel := flag.String("llm-model", "", "LLM model")
	llmProvider := flag.String("llm-provider", "", "LLM provider")
	llmApiKey := flag.String("llm-api-key", "", "LLM API key")
	llmBaseURL := flag.String("llm-base-url", "", "LLM base URL")
	llmTimeout := flag.Int("llm-timeout", 0, "LLM timeout in ms")

	sessionKey := flag.String("session-key", "", "Session key")
	sessionID := flag.String("session-id", "", "Session ID")
	messages := flag.String("messages", "[]", "JSON messages")
	role := flag.String("role", "user", "Message role")
	content := flag.String("content", "", "Message content")
	tokenBudget := flag.Int("token-budget", 128000, "Token budget")
	force := flag.Bool("force", false, "Force")
	pattern := flag.String("pattern", "", "Search pattern")
	mode := flag.String("mode", "full_text", "Mode")
	scope := flag.String("scope", "all", "Scope")
	allConv := flag.Bool("all", false, "All conversations")
	limit := flag.Int("limit", 20, "Limit")
	sort := flag.String("sort", "desc", "Sort")
	id := flag.String("id", "", "ID")
	summaryIDs := flag.String("summary-ids", "", "Summary IDs")
	query := flag.String("query", "", "Query")
	maxDepth := flag.Int("max-depth", 3, "Max depth")
	includeMsgs := flag.Bool("include-messages", false, "Include messages")
	maintOp := flag.String("maint-op", "gc", "Operation: gc, vacuum, backup, doctor, clean, rotate")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "clawlcm - Lossless Context Management v0.6.0\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  bootstrap  Initialize conversation\n")
		fmt.Fprintf(os.Stderr, "  ingest     Add message\n")
		fmt.Fprintf(os.Stderr, "  assemble   Get context\n")
		fmt.Fprintf(os.Stderr, "  compact    Trigger compaction\n")
		fmt.Fprintf(os.Stderr, "  grep       Search\n")
		fmt.Fprintf(os.Stderr, "  describe   Describe summary\n")
		fmt.Fprintf(os.Stderr, "  expand     Expand summary\n")
		fmt.Fprintf(os.Stderr, "  maintain   Run maintenance (gc|vacuum|backup|doctor|clean|rotate)\n")
		fmt.Fprintf(os.Stderr, "  tui        Interactive TUI mode\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("clawlcm version %s (commit: %s, date: %s)\n", version, commit, date)
		return
	}

	if flag.NArg() < 1 {
		runDemoTest(dbPath, configFile, enable, verbose, llmModel, llmProvider, llmApiKey, llmBaseURL, llmTimeout)
		return
	}

	cmd := flag.Arg(0)

	remainingArgs := flag.Args()[1:]
	if cmd == "bootstrap" || cmd == "ingest" {
		for i, arg := range remainingArgs {
			if (arg == "--session-key" || arg == "-session-key") && i+1 < len(remainingArgs) {
				*sessionKey = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--session-key=") {
				*sessionKey = strings.TrimPrefix(arg, "--session-key=")
			}
			if (arg == "--content" || arg == "-content") && i+1 < len(remainingArgs) {
				*content = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--content=") {
				*content = strings.TrimPrefix(arg, "--content=")
			}
			if (arg == "--role" || arg == "-role") && i+1 < len(remainingArgs) {
				*role = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--role=") {
				*role = strings.TrimPrefix(arg, "--role=")
			}
			if (arg == "--messages" || arg == "-messages") && i+1 < len(remainingArgs) {
				*messages = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--messages=") {
				*messages = strings.TrimPrefix(arg, "--messages=")
			}
			if (arg == "--session-id" || arg == "-session-id") && i+1 < len(remainingArgs) {
				*sessionID = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--session-id=") {
				*sessionID = strings.TrimPrefix(arg, "--session-id=")
			}
		}
	} else {
		for i, arg := range remainingArgs {
			if (arg == "--session-key" || arg == "-session-key") && i+1 < len(remainingArgs) {
				*sessionKey = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--session-key=") {
				*sessionKey = strings.TrimPrefix(arg, "--session-key=")
			}
			if strings.HasPrefix(arg, "-session-key=") {
				*sessionKey = strings.TrimPrefix(arg, "-session-key=")
			}
			if arg == "--force" || arg == "-force" {
				*force = true
			}
			if (arg == "--pattern" || arg == "-pattern") && i+1 < len(remainingArgs) && cmd == "grep" {
				*pattern = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--pattern=") && cmd == "grep" {
				*pattern = strings.TrimPrefix(arg, "--pattern=")
			}
			// Fix: Parse --id for describe command
			if (arg == "--id" || arg == "-id") && i+1 < len(remainingArgs) && cmd == "describe" {
				*id = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--id=") {
				*id = strings.TrimPrefix(arg, "--id=")
			}
			if strings.HasPrefix(arg, "-id=") {
				*id = strings.TrimPrefix(arg, "-id=")
			}
			// Fix: Parse --summary-ids for expand command
			if (arg == "--summary-ids" || arg == "-summary-ids") && i+1 < len(remainingArgs) && cmd == "expand" {
				*summaryIDs = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--summary-ids=") {
				*summaryIDs = strings.TrimPrefix(arg, "--summary-ids=")
			}
			if strings.HasPrefix(arg, "-summary-ids=") {
				*summaryIDs = strings.TrimPrefix(arg, "-summary-ids=")
			}
			// Fix: Parse --query for expand command
			if (arg == "--query" || arg == "-query") && i+1 < len(remainingArgs) && cmd == "expand" {
				*query = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "--query=") {
				*query = strings.TrimPrefix(arg, "--query=")
			}
			// Fix: Parse --max-depth for expand command
			if (arg == "--max-depth" || arg == "-max-depth") && i+1 < len(remainingArgs) && cmd == "expand" {
				if val, err := strconv.Atoi(remainingArgs[i+1]); err == nil {
					*maxDepth = val
				}
			}
			if strings.HasPrefix(arg, "--max-depth=") {
				if val, err := strconv.Atoi(strings.TrimPrefix(arg, "--max-depth=")); err == nil {
					*maxDepth = val
				}
			}
			// Fix: Parse -maint-op (single dash) for maintain command
			if (arg == "-maint-op" || arg == "--maint-op") && i+1 < len(remainingArgs) && cmd == "maintain" {
				*maintOp = remainingArgs[i+1]
			}
			if strings.HasPrefix(arg, "-maint-op=") && cmd == "maintain" {
				*maintOp = strings.TrimPrefix(arg, "-maint-op=")
			}
			if strings.HasPrefix(arg, "--maint-op=") && cmd == "maintain" {
				*maintOp = strings.TrimPrefix(arg, "--maint-op=")
			}
		}
	}

	// For maintain command, manually parse --op/-maint-op from remaining args
	if cmd == "maintain" {
		for i := 1; i < len(flag.Args()); i++ {
			if (flag.Args()[i] == "--op" || flag.Args()[i] == "-op" || flag.Args()[i] == "-maint-op" || flag.Args()[i] == "--maint-op") && i+1 < len(flag.Args()) {
				*maintOp = flag.Args()[i+1]
				break
			}
			if strings.HasPrefix(flag.Args()[i], "--op=") {
				*maintOp = strings.TrimPrefix(flag.Args()[i], "--op=")
				break
			}
			if strings.HasPrefix(flag.Args()[i], "-op=") {
				*maintOp = strings.TrimPrefix(flag.Args()[i], "-op=")
				break
			}
			if strings.HasPrefix(flag.Args()[i], "-maint-op=") {
				*maintOp = strings.TrimPrefix(flag.Args()[i], "-maint-op=")
				break
			}
			if strings.HasPrefix(flag.Args()[i], "--maint-op=") {
				*maintOp = strings.TrimPrefix(flag.Args()[i], "--maint-op=")
				break
			}
		}
	}

	cfgPath := defaultCfgPath
	if *configFile != "" {
		cfgPath = *configFile
	}

	cfgJSON, err := loadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	cfg := clawlcm.DefaultConfig()

	if *dbPath != "" {
		cfg.DatabasePath = *dbPath
	} else if cfgJSON.Database.Path != "" {
		cfg.DatabasePath = cfgJSON.Database.Path
	}

	cfg.Enabled = *enable
	cfg.Verbose = *verbose

	if *llmModel != "" {
		cfg.SummaryModel = *llmModel
	} else if cfgJSON.LLM.Model != "" {
		cfg.SummaryModel = cfgJSON.LLM.Model
	}

	if *llmProvider != "" {
		cfg.SummaryProvider = *llmProvider
	} else if cfgJSON.LLM.Provider != "" {
		cfg.SummaryProvider = cfgJSON.LLM.Provider
	}

	if *llmApiKey != "" {
		cfg.SummaryAPIKey = *llmApiKey
	} else if cfgJSON.LLM.APIKey != "" {
		cfg.SummaryAPIKey = cfgJSON.LLM.APIKey
	}

	if *llmBaseURL != "" {
		cfg.SummaryBaseURL = *llmBaseURL
	} else if cfgJSON.LLM.BaseURL != "" {
		cfg.SummaryBaseURL = cfgJSON.LLM.BaseURL
	}

	if *llmTimeout > 0 {
		cfg.SummaryTimeoutMs = *llmTimeout
	} else if cfgJSON.LLM.TimeoutMs > 0 {
		cfg.SummaryTimeoutMs = cfgJSON.LLM.TimeoutMs
	}

	if cfgJSON.Context.Threshold > 0 {
		cfg.ContextThreshold = cfgJSON.Context.Threshold
	}

	if cfgJSON.Context.FreshTailCount > 0 {
		cfg.FreshTailCount = cfgJSON.Context.FreshTailCount
	}

	cfg.UseCJKTokenizer = cfgJSON.Context.UseCJKTokenizer

	if cfgJSON.Context.CondensedMinFanout > 0 {
		cfg.CondensedMinFanout = cfgJSON.Context.CondensedMinFanout
	}

	if cfgJSON.Context.IncrementalMaxDepth != 0 {
		cfg.IncrementalMaxDepth = cfgJSON.Context.IncrementalMaxDepth
	}

	cfg.IgnoreSessionPatterns = cfgJSON.Session.IgnoreSessionPatterns
	cfg.StatelessSessionPatterns = cfgJSON.Session.StatelessSessionPatterns
	cfg.SkipStatelessSessions = cfgJSON.Session.SkipStatelessSessions

	log := logger.New()
	if *verbose || cfg.Verbose {
		log.Info("Starting LCM Engine")
		log.Info(fmt.Sprintf("Database: %s", cfg.DatabasePath))
	}

	engine, err := clawlcm.NewEngine(cfg, log)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create engine: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	switch cmd {
	case "bootstrap":
		handleBootstrap(ctx, engine, sessionKey, sessionID, messages)
	case "ingest":
		handleIngest(ctx, engine, sessionKey, role, content)
	case "assemble":
		handleAssemble(ctx, engine, sessionKey, tokenBudget)
	case "compact":
		handleCompact(ctx, engine, sessionKey, force)
	case "grep":
		handleGrep(ctx, engine, sessionKey, pattern, mode, scope, allConv, limit, sort)
	case "describe":
		handleDescribe(ctx, engine, sessionKey, id, allConv)
	case "expand":
		handleExpand(ctx, engine, sessionKey, summaryIDs, query, maxDepth, includeMsgs)
	case "maintain":
		handleMaintain(ctx, engine, sessionKey, maintOp)
	case "tui":
		handleTUI(ctx, engine)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func runDemoTest(dbPath, configFile *string, enable, verbose *bool, llmModel, llmProvider, llmApiKey, llmBaseURL *string, llmTimeout *int) {
	cfgPath := defaultCfgPath
	if *configFile != "" {
		cfgPath = *configFile
	}

	cfgJSON, err := loadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	cfg := clawlcm.DefaultConfig()

	if *dbPath != "" {
		cfg.DatabasePath = *dbPath
	} else if cfgJSON.Database.Path != "" {
		cfg.DatabasePath = cfgJSON.Database.Path
	}

	cfg.Enabled = *enable
	cfg.Verbose = *verbose

	if *llmModel != "" {
		cfg.SummaryModel = *llmModel
	} else if cfgJSON.LLM.Model != "" {
		cfg.SummaryModel = cfgJSON.LLM.Model
	}

	if *llmProvider != "" {
		cfg.SummaryProvider = *llmProvider
	} else if cfgJSON.LLM.Provider != "" {
		cfg.SummaryProvider = cfgJSON.LLM.Provider
	}

	if *llmApiKey != "" {
		cfg.SummaryAPIKey = *llmApiKey
	} else if cfgJSON.LLM.APIKey != "" {
		cfg.SummaryAPIKey = cfgJSON.LLM.APIKey
	}

	if *llmBaseURL != "" {
		cfg.SummaryBaseURL = *llmBaseURL
	} else if cfgJSON.LLM.BaseURL != "" {
		cfg.SummaryBaseURL = cfgJSON.LLM.BaseURL
	}

	if *llmTimeout > 0 {
		cfg.SummaryTimeoutMs = *llmTimeout
	} else if cfgJSON.LLM.TimeoutMs > 0 {
		cfg.SummaryTimeoutMs = cfgJSON.LLM.TimeoutMs
	}

	if cfgJSON.Context.Threshold > 0 {
		cfg.ContextThreshold = cfgJSON.Context.Threshold
	}

	if cfgJSON.Context.FreshTailCount > 0 {
		cfg.FreshTailCount = cfgJSON.Context.FreshTailCount
	}

	cfg.UseCJKTokenizer = cfgJSON.Context.UseCJKTokenizer

	if cfgJSON.Context.CondensedMinFanout > 0 {
		cfg.CondensedMinFanout = cfgJSON.Context.CondensedMinFanout
	}

	if cfgJSON.Context.IncrementalMaxDepth != 0 {
		cfg.IncrementalMaxDepth = cfgJSON.Context.IncrementalMaxDepth
	}

	cfg.IgnoreSessionPatterns = cfgJSON.Session.IgnoreSessionPatterns
	cfg.StatelessSessionPatterns = cfgJSON.Session.StatelessSessionPatterns
	cfg.SkipStatelessSessions = cfgJSON.Session.SkipStatelessSessions

	log := logger.New()
	if *verbose || cfg.Verbose {
		log.Info("Starting LCM Engine")
		log.Info(fmt.Sprintf("Database: %s", cfg.DatabasePath))
		log.Info(fmt.Sprintf("Enabled: %v", cfg.Enabled))
		if cfg.SummaryModel != "" {
			log.Info(fmt.Sprintf("LLM: %s @ %s", cfg.SummaryModel, cfg.SummaryBaseURL))
		}
		log.Info(fmt.Sprintf("IncrementalMaxDepth: %d", cfg.IncrementalMaxDepth))
		log.Info(fmt.Sprintf("CondensedMinFanout: %d", cfg.CondensedMinFanout))
	}

	engine, err := clawlcm.NewEngine(cfg, log)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create engine: %v", err))
		os.Exit(1)
	}

	info := engine.Info()
	log.Info(fmt.Sprintf("Engine: %s v%s", info.Name, info.Version))

	ctx := context.Background()
	log.Info("=== Running condensed summarization test ===")
	testSession := "test:novel:condensed"

	// Step 1: Bootstrap
	log.Info("Step 1: Bootstrap session")
	_, err = engine.Bootstrap(ctx, clawlcm.BootstrapRequest{
		SessionKey:  testSession,
		SessionID:   "novel-001",
		TokenBudget: 128000,
	})
	if err != nil {
		log.Error(fmt.Sprintf("Bootstrap failed: %v", err))
		os.Exit(1)
	}
	log.Info("Bootstrap OK")

	// Step 2: Ingest novel files
	log.Info("Step 2: Ingest novel files...")
	files := []string{
		"/home/ubuntu/workspace/《庆余年》_qinkan.net.txt",
	}
	totalBatches := 0
	maxBatches := 150 // Limit to prevent timeout
	for _, fpath := range files {
		if totalBatches >= maxBatches {
			break
		}
		content, err := os.ReadFile(fpath)
		if err != nil {
			log.Error(fmt.Sprintf("Read %s failed: %v", fpath, err))
			continue
		}
		lines := strings.Split(string(content), "\n")
		batchSize := 50
		for i := 0; i < len(lines) && totalBatches < maxBatches; i += batchSize {
			end := i + batchSize
			if end > len(lines) {
				end = len(lines)
			}
			batch := strings.Join(lines[i:end], "\n")
			if batch == "" {
				continue
			}
			_, err := engine.Ingest(ctx, clawlcm.IngestRequest{
				SessionKey:  testSession,
				TokenBudget: 128000,
				Message:     clawlcm.Message{Role: "user", Content: batch},
			})
			if err != nil {
				log.Error(fmt.Sprintf("Ingest failed: %v", err))
			}
			totalBatches++
			if totalBatches%25 == 0 {
				log.Info(fmt.Sprintf("Ingested %d batches", totalBatches))
			}
		}
	}
	log.Info(fmt.Sprintf("Total batches ingested: %d", totalBatches))

	// Step 3: Leaf compression
	log.Info("Step 3: Leaf compression (maxDepth=0)")
	for i := 0; i < 15; i++ {
		r, err := engine.Compact(ctx, clawlcm.CompactRequest{
			SessionKey: testSession,
			MaxDepth:   0,
			Force:      true,
		})
		if err != nil {
			log.Error(fmt.Sprintf("Compact error: %v", err))
			break
		}
		if !r.ActionTaken {
			log.Info("No leaf compression")
			break
		}
		log.Info(fmt.Sprintf("Leaf %d: before=%d, after=%d, summary=%d, condensed=%v",
			i+1, r.TokensBefore, r.TokensAfter, r.SummaryID, r.Condensed))
	}

	// Step 4: Check summaries
	asm, err := engine.Assemble(ctx, clawlcm.AssembleRequest{
		SessionKey:  testSession,
		TokenBudget: 128000,
	})
	if err != nil {
		log.Error(fmt.Sprintf("Assemble error: %v", err))
	} else {
		log.Info(fmt.Sprintf("Assembled: messages=%d, tokens=%d, raw=%d, summaries=%d",
			len(asm.Messages), asm.EstimatedTokens, asm.Stats.RawMessageCount, asm.Stats.SummaryCount))
	}

	// Step 5: Condensed compression
	log.Info("Step 4: Condensed compression (maxDepth=1)")
	for i := 0; i < 10; i++ {
		r, err := engine.Compact(ctx, clawlcm.CompactRequest{
			SessionKey: testSession,
			MaxDepth:   1,
		})
		if err != nil {
			log.Error(fmt.Sprintf("Compact error: %v", err))
			break
		}
		if !r.ActionTaken {
			log.Info("No condensed compression")
			break
		}
		log.Info(fmt.Sprintf("Condensed %d: before=%d, after=%d, summary=%d, condensed=%v",
			i+1, r.TokensBefore, r.TokensAfter, r.SummaryID, r.Condensed))
	}

	// Final check
	asm, err = engine.Assemble(ctx, clawlcm.AssembleRequest{
		SessionKey:  testSession,
		TokenBudget: 128000,
	})
	if err != nil {
		log.Error(fmt.Sprintf("Final assemble error: %v", err))
	} else {
		log.Info(fmt.Sprintf("Final: messages=%d, tokens=%d, raw=%d, summaries=%d",
			len(asm.Messages), asm.EstimatedTokens, asm.Stats.RawMessageCount, asm.Stats.SummaryCount))
	}

	log.Info("Condensed test completed!")
}

func handleBootstrap(ctx context.Context, engine *clawlcm.Engine, sessionKey, sessionID, messages *string) {
	if *sessionKey == "" {
		fmt.Fprintf(os.Stderr, "Error: --session-key is required\n")
		os.Exit(1)
	}

	var msgs []clawlcm.Message
	json.Unmarshal([]byte(*messages), &msgs)

	resp, err := engine.Bootstrap(ctx, clawlcm.BootstrapRequest{
		SessionKey:  *sessionKey,
		SessionID:   *sessionID,
		TokenBudget: 128000,
		Messages:    msgs,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func handleIngest(ctx context.Context, engine *clawlcm.Engine, sessionKey, role, content *string) {
	if *sessionKey == "" || *content == "" {
		fmt.Fprintf(os.Stderr, "Error: --session-key and --content are required\n")
		os.Exit(1)
	}

	resp, err := engine.Ingest(ctx, clawlcm.IngestRequest{
		SessionKey:  *sessionKey,
		TokenBudget: 128000,
		Message:     clawlcm.Message{Role: *role, Content: *content},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ingest failed: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func handleAssemble(ctx context.Context, engine *clawlcm.Engine, sessionKey *string, tokenBudget *int) {
	if *sessionKey == "" {
		fmt.Fprintf(os.Stderr, "Error: --session-key is required\n")
		os.Exit(1)
	}

	resp, err := engine.Assemble(ctx, clawlcm.AssembleRequest{
		SessionKey:  *sessionKey,
		TokenBudget: *tokenBudget,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assemble failed: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func handleCompact(ctx context.Context, engine *clawlcm.Engine, sessionKey *string, force *bool) {
	if *sessionKey == "" {
		fmt.Fprintf(os.Stderr, "Error: --session-key is required\n")
		os.Exit(1)
	}

	resp, err := engine.Compact(ctx, clawlcm.CompactRequest{
		SessionKey: *sessionKey,
		Force:      *force,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compact failed: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func handleGrep(ctx context.Context, engine *clawlcm.Engine, sessionKey, pattern, mode, scope *string, allConv *bool, limit *int, sort *string) {
	if *sessionKey == "" && !*allConv {
		fmt.Fprintf(os.Stderr, "Error: --session-key or --all is required\n")
		os.Exit(1)
	}

	resp, err := engine.Grep(ctx, clawlcm.GrepRequest{
		SessionKey:       *sessionKey,
		Pattern:          *pattern,
		Mode:             *mode,
		Scope:            *scope,
		AllConversations: *allConv,
		Limit:            *limit,
		Sort:             *sort,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Grep failed: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func handleDescribe(ctx context.Context, engine *clawlcm.Engine, sessionKey, id *string, allConv *bool) {
	if *sessionKey == "" && !*allConv {
		fmt.Fprintf(os.Stderr, "Error: --session-key or --all is required\n")
		os.Exit(1)
	}

	if *id == "" {
		fmt.Fprintf(os.Stderr, "Error: --id is required\n")
		os.Exit(1)
	}

	resp, err := engine.Describe(ctx, clawlcm.DescribeRequest{
		SessionKey:       *sessionKey,
		ID:               *id,
		AllConversations: *allConv,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Describe failed: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func handleExpand(ctx context.Context, engine *clawlcm.Engine, sessionKey, summaryIDs, query *string, maxDepth *int, includeMsgs *bool) {
	if *sessionKey == "" {
		fmt.Fprintf(os.Stderr, "Error: --session-key is required\n")
		os.Exit(1)
	}

	if *summaryIDs == "" {
		fmt.Fprintf(os.Stderr, "Error: --summary-ids is required\n")
		os.Exit(1)
	}

	var ids []string
	for _, s := range strings.Split(*summaryIDs, ",") {
		if s = strings.TrimSpace(s); s != "" {
			ids = append(ids, s)
		}
	}

	resp, err := engine.Expand(ctx, clawlcm.ExpandRequest{
		SessionKey:      *sessionKey,
		SummaryIDs:      ids,
		Query:           *query,
		MaxDepth:        *maxDepth,
		IncludeMessages: *includeMsgs,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Expand failed: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func handleMaintain(ctx context.Context, engine *clawlcm.Engine, sessionKey, maintOp *string) {
	var req clawlcm.MaintainRequest
	if *sessionKey != "" {
		req.SessionKey = new(string)
		*req.SessionKey = *sessionKey
	}
	req.Operation = *maintOp

	resp, err := engine.Maintain(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Maintain failed: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func handleTUI(ctx context.Context, engine *clawlcm.Engine) {
	fmt.Println("clawlcm TUI - Interactive Mode")
	fmt.Println("================================")
	fmt.Println("Available commands:")
	fmt.Println("  sessions   - List all conversations")
	fmt.Println("  inspect    - Inspect a conversation")
	fmt.Println("  compact    - Force compact a conversation")
	fmt.Println("  doctor     - Run health check")
	fmt.Println("  backup     - Create backup")
	fmt.Println("  quit       - Exit TUI")
	fmt.Println("")

	convs, err := engine.Grep(ctx, clawlcm.GrepRequest{AllConversations: true, Limit: 100})
	if err != nil || !convs.Success {
		fmt.Println("No conversations found")
		return
	}

	fmt.Printf("Found %d conversations:\n", convs.TotalMatches)
	for i, m := range convs.Matches[:min(10, len(convs.Matches))] {
		fmt.Printf("  %d. conv_id=%d, role=%s, content=%s...\n", i+1, m.ConversationID, m.Role, m.Content[:min(50, len(m.Content))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
