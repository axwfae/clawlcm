package types

import "time"

// ── Message Types ──────────────────────────────────────────────────────────────

type Message struct {
	Role    string `json:"role"`    // "user" | "assistant" | "system"
	Content string `json:"content"` // Message content
}

// ── Conversation Types ──────────────────────────────────────────────────────────

type Conversation struct {
	ID           int64     `json:"id"`
	SessionKey   string    `json:"sessionKey"`
	SessionID    string    `json:"sessionID"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	MessageCount int       `json:"messageCount"`
	TokenCount   int       `json:"tokenCount"`
}

type ConversationStore interface {
	Create(sessionKey, sessionID string) (int64, error)
	GetBySessionKey(sessionKey string) (*Conversation, error)
	GetByID(id int64) (*Conversation, error)
	Delete(id int64) error
	UpdateStats(id int64, messageCount, tokenCount int) error
}

// ── Message Types ──────────────────────────────────────────────────────────────

type MessageRecord struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversationId"`
	Ordinal        int       `json:"ordinal"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	TokenCount     int       `json:"tokenCount"`
	CreatedAt      time.Time `json:"createdAt"`
}

type MessageStore interface {
	Create(conversationID int64, ordinal int, role, content string, tokenCount int) (int64, error)
	GetByConversation(conversationID int64, limit, offset int) ([]MessageRecord, error)
	GetByID(id int64) (*MessageRecord, error)
	Count(conversationID int64) (int, error)
	GetAllTokens(conversationID int64) (int, error)
}

// ── Summary Types (DAG) ────────────────────────────────────────────────────────

type SummaryType string

const (
	SummaryTypeLeaf      SummaryType = "leaf"      // Raw messages → summary
	SummaryTypeCondensed SummaryType = "condensed" // Summary → summary
)

type SummaryRecord struct {
	ID             int64       `json:"id"`
	ConversationID int64       `json:"conversationId"`
	SummaryType    SummaryType `json:"summaryType"`
	Depth          int         `json:"depth"`        // DAG depth: 0=leaf, 1+=condensed
	Content        string      `json:"content"`      // Summary text
	TokenCount     int         `json:"tokenCount"`   // Summary token count
	SourceTokens   int         `json:"sourceTokens"` // Sum of source item tokens
	Ordinal        int         `json:"ordinal"`      // Ordinal position in conversation
	CreatedAt      time.Time   `json:"createdAt"`
	ParentIDs      []int64     `json:"parentIds"` // Parent summary IDs
	SourceIDs      []int64     `json:"sourceIds"` // Source message IDs
}

type SummaryStore interface {
	Create(conversationID int64, summaryType SummaryType, depth int, content string, tokenCount, sourceTokens, ordinal int, parentIDs, sourceIDs []int64) (int64, error)
	GetByConversation(conversationID int64) ([]SummaryRecord, error)
	GetByID(id int64) (*SummaryRecord, error)
	GetByDepth(conversationID int64, depth int) ([]SummaryRecord, error)
	GetLeafSummaries(conversationID int64) ([]SummaryRecord, error)
	GetCondensedSummaries(conversationID int64) ([]SummaryRecord, error)
	Delete(id int64) error
	DeleteByConversation(conversationID int64) error
}

// ── Context Item Types ─────────────────────────────────────────────────────────

type ContextItemType string

const (
	ContextItemMessage ContextItemType = "message"
	ContextItemSummary ContextItemType = "summary"
)

type ContextItemRecord struct {
	ID             int64           `json:"id"`
	ConversationID int64           `json:"conversationId"`
	ItemType       ContextItemType `json:"itemType"`
	ItemID         int64           `json:"itemId"`     // MessageID or SummaryID
	Ordinal        int             `json:"ordinal"`    // Position in assembled context
	TokenCount     int             `json:"tokenCount"` // Token count
	Keywords       string          `json:"keywords"`   // Extracted keywords (BM25)
	CreatedAt      time.Time       `json:"createdAt"`
}

type ContextItemStore interface {
	Create(conversationID int64, itemType ContextItemType, itemID int64, ordinal, tokenCount int, keywords string) (int64, error)
	GetByConversation(conversationID int64) ([]ContextItemRecord, error)
	Delete(id int64) error
	DeleteByConversation(conversationID int64) error
}

// ── Retrieval Types ─────────────────────────────────────────────────────────────

type RetrievalResult struct {
	Item     *ContextItemRecord
	Score    float64
	Messages []Message // For summaries: expanded messages
}

type RetrievalRequest struct {
	SessionKey string
	Query      string
	MaxResults int
	MinScore   float64
}

// ── Config Types ────────────────────────────────────────────────────────────────

type Config struct {
	Enabled               bool
	DatabasePath          string
	ContextThreshold      float64
	FreshTailCount        int
	LeafChunkTokens       int
	LeafTargetTokens      int
	CondensedTargetTokens int
	CondensedMinFanout    int
	IncrementalMaxDepth   int
	MaxRounds             int

	// Compression modes
	ProactiveThresholdCompactionMode string // "inline" | "deferred"
	MaintenanceDebtEnabled           bool   // Track maintenance debt in deferred mode
	MaintenanceDebtThreshold         int    // Max compaction debt before forced run
	CacheAwareCompaction             bool
	CacheTTLSeconds                  int
	CacheThroughputThreshold         int

	// Large files externalization
	LargeFilesDir string // Directory for externalized large files

	// Session management
	IgnoreSessionPatterns    []string
	StatelessSessionPatterns []string
	SkipStatelessSessions    bool

	// LLM configuration
	SummaryModel     string
	SummaryProvider  string
	SummaryAPIKey    string
	SummaryBaseURL   string
	SummaryTimeoutMs int

	// Token estimation
	UseCJKTokenizer bool

	// Logging
	Verbose bool
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:                          true,
		DatabasePath:                     "data/clawlcm.db",
		ContextThreshold:                 0.75,
		FreshTailCount:                   8,
		LeafChunkTokens:                  20000,
		LeafTargetTokens:                 2400,
		CondensedTargetTokens:            2000,
		CondensedMinFanout:               4,
		IncrementalMaxDepth:              1,
		MaxRounds:                        10,
		ProactiveThresholdCompactionMode: "deferred",
		MaintenanceDebtEnabled:           true,
		MaintenanceDebtThreshold:         50000,
		LargeFilesDir:                    "data/large_files",
		CacheAwareCompaction:             false,
		CacheTTLSeconds:                  3600,
		CacheThroughputThreshold:         10000,
		SummaryTimeoutMs:                 60000,
		UseCJKTokenizer:                  true,
		Verbose:                          false,
	}
}

type EngineInfo struct {
	ID      string
	Name    string
	Version string
}

type BootstrapRequest struct {
	SessionKey  string
	SessionID   string
	TokenBudget int
	Messages    []Message
}

type BootstrapResponse struct {
	ConversationID int64
	Messages       []Message
	TokenCount     int
}

type IngestRequest struct {
	SessionKey  string
	SessionID   string
	TokenBudget int
	Message     Message
}

type IngestResponse struct {
	MessageID     int64
	Ordinal       int
	TokenCount    int
	ShouldCompact bool
}

type AssembleRequest struct {
	SessionKey  string
	TokenBudget int
	Prompt      string
}

type AssembleResponse struct {
	Messages        []Message
	EstimatedTokens int
	SystemPrompt    string
	Stats           Stats
}

type Stats struct {
	RawMessageCount int
	SummaryCount    int
}

type CompactRequest struct {
	SessionKey string
	Force      bool
	MaxDepth   int
}

type CompactResponse struct {
	ActionTaken  bool
	TokensBefore int
	TokensAfter  int
	SummaryID    int64
	Condensed    bool
}

type MaintainRequest struct {
	SessionKey *string
	Operation  string // "gc" | "vacuum" | "backup" | "doctor" | "clean" | "rotate"
}

type MaintainResponse struct {
	Changed    bool
	BytesFreed int64
	Rewritten  int
	Errors     []string
}

type RotateRequest struct {
	SessionKey    string
	CopySummaries bool
}

type RotateResponse struct {
	Success   bool
	OldConvID int64
	NewConvID int64
	MsgMoved  int
	Errors    []string
}

// ── Agent Tool Types ─────────────────────────────────────────────────────────────

type ToolRequest struct {
	SessionKey       string
	AllConversations bool
	TokenCap         int
}

type ToolResponse struct {
	Success bool
	Error   string
	Data    interface{}
}

type GrepRequest struct {
	SessionKey       string
	Pattern          string
	Mode             string // "regex" | "full_text"
	Scope            string // "messages" | "summaries" | "all"
	AllConversations bool
	Since            string
	Before           string
	Limit            int
	Sort             string // "asc" | "desc"
	TokenCap         int
}

type GrepResponse struct {
	Success      bool
	Error        string
	MatchCount   int
	Matches      []GrepMatch
	TotalMatches int
}

type GrepMatch struct {
	ConversationID int64
	SessionKey     string
	MessageID      int64
	Ordinal        int
	Role           string
	Content        string
	Score          float64
}

type DescribeRequest struct {
	SessionKey       string
	ID               string // "sum_xxx" or file_xxx
	AllConversations bool
	TokenCap         int
}

type DescribeResponse struct {
	Success     bool
	Error       string
	SummaryID   int64
	SummaryType string
	Depth       int
	Content     string
	TokenCount  int
	SourceCount int
	CreatedAt   string
}

type ExpandRequest struct {
	SessionKey       string
	SummaryIDs       []string
	Query            string
	MaxDepth         int
	TokenCap         int
	IncludeMessages  bool
	AllConversations bool
}

type ExpandResponse struct {
	Success     bool
	Error       string
	Expanded    []ExpandedSummary
	TotalTokens int
}

type ExpandedSummary struct {
	SummaryID   int64
	Depth       int
	Content     string
	SourceItems []ExpandedSource
}

type ExpandedSource struct {
	Type    string // "message" | "summary"
	ID      int64
	Content string
}
