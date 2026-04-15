---
name: clawlcm
description: Lossless Context Management - LLM-driven summarization, BM25 retrieval, DAG compression, CJK support. Based on lossless-claw-enhanced.
metadata: {nanobot:{emoji:🧠,requires:{bins:[clawlcm]},install:[{id:manual,kind:binary,label:Copy clawlcm binary to skill bin directory}]}}
---

# clawlcm Skill

> **Version**: v0.8.1 | **Updated**: 2026-04-15 | **Reference**: [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced) | **Based on**: picolcm

Enhanced ultra-lightweight lossless context management system based on [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced).

## What is LCM?

When conversations exceed the model's context window, traditional methods truncate old messages. LCM uses a DAG-structured summarization system that preserves every message while keeping the active context within the model's token limit.

```
Workflow:
1. Persist   ──→  Store all messages in SQLite
2. Summarize ──→  Compress old messages into Leaf Summary
3. Condense  ──→  Merge multiple Leaves into higher-level nodes
4. Assemble ──→  Summary + Fresh Tail = Full Context
```

## ⚠️ Important: Configuration Required Before Running

Before using clawlcm, **you must configure LLM settings** in `config.json`, otherwise it won't work!

### Config File `config.json`

```json
{
  "database": {
    "path": "./data/clawlcm.db"
  },
  "llm": {
    "model": "",
    "provider": "openai",
    "apiKey": "",
    "baseURL": "",
    "timeoutMs": 120000
  },
  "context": {
    "threshold": 0.75,
    "freshTailCount": 8,
    "useCJKTokenizer": true,
    "maxDepth": 8
  },
  "enabled": true,
  "verbose": false
}
```

> ⚠️ **Note**: Do NOT include `/v1` at the end of `baseURL`, code will automatically append `/v1/chat/completions`

### Or Override via Command Line

```bash
./bin/clawlcm --llm-model <YOUR_MODEL> --llm-api-key <YOUR_KEY> --llm-base-url <URL> ...
```

> **⚠️ Please do not hardcode sensitive info in config, use env vars or CLI args instead**

## Directory Structure

```
clawlcm/
├── bin/clawlcm           # Main binary
├── data/                 # Data directory
│   ├── config.json       # Config file
│   └── clawlcm.db        # SQLite database
└── SKILL_en.md          # Skill definition (English)
```

### Data Directory

- **Config**: `./data/config.json` (auto-created default)
- **Database**: `./data/clawlcm.db`
- Recommend mapping `data/` to external volume for data persistence

## DAG Structure

```
         [Message 1-15]                   ← Original messages (compressed)
               │
               ▼
        [Leaf Summary]                    ← Leaf summary (depth=0)
               │
               ▼
    ┌──────────┴──────────┐
    ▼                     ▼
[Leaf A]             [Leaf B]           ← Multiple Leaves condensed
    │                     │
    └──────────┬──────────┘
               ▼
       [Condensed Summary]              ← Condensed summary (depth=1)
               │
               ▼
    [Fresh Tail: Last N messages]        ← Protected fresh tail
```

- **Leaf Summary**: Compressed summary of original messages
- **Condensed Summary**: Further condensation of multiple Leaf summaries
- **Fresh Tail**: Protect last N messages from compression

## Use Cases

Use this skill when:
- Need LLM-driven summarization (Leaf + Condensed)
- Need BM25 relevance retrieval
- Multi-level compression (DAG depth tracking)
- Chinese/Japanese/Korean text support
- Need enhanced features:
  - Accurate CJK Token estimation (1.5x CJK, 2x Emoji)
  - Auth Error filtering
  - Session Rotation detection
  - Empty message skip

## Quick Start

```bash
# Check version
./bin/clawlcm --version

# Test LLM (please replace with your configuration)
./bin/clawlcm -v
```

## Operations

### bootstrap
Initialize conversation and load existing messages.

```bash
./bin/clawlcm bootstrap \
  --session-key user:chat:123 \
  --session-id uuid-123 \
  --token-budget 128000 \
  --messages '[{"role":"user","content":"Hello"},{"role":"assistant","content":"Hi"}]'
```

### ingest
Add new message to conversation.

```bash
./bin/clawlcm ingest \
  --session-key user:chat:123 \
  --session-id uuid-123 \
  --role user \
  --content "Explain Go goroutines"
```

### assemble
Assemble context (summary + fresh tail).

```bash
./bin/clawlcm assemble \
  --session-key user:chat:123 \
  --token-budget 128000
```

### compact
Trigger LLM summarization (create Leaf summary).

```bash
./bin/clawlcm compact \
  --session-key user:chat:123 \
  --force
```

### grep ⭐ NEW
BM25 search messages.

```bash
# Search single conversation
./bin/clawlcm grep \
  --session-key user:chat:123 \
  --pattern "keyword"

# Search all conversations
./bin/clawlcm grep \
  --all \
  --pattern "keyword" \
  --limit 20
```

### describe ⭐ NEW
Describe summary details.

```bash
./bin/clawlcm describe \
  --session-key user:chat:123 \
  --id summary-id
```

### expand ⭐ NEW
Expand summary content.

```bash
./bin/clawlcm expand \
  --session-key user:chat:123 \
  --summary-ids "id1,id2" \
  --query "query content" \
  --max-depth 3
```

### maintain ⭐ NEW OPERATIONS
Execute maintenance tasks.

```bash
# Garbage collection
./bin/clawlcm maintain --session-key user:chat:123 --op gc

# Database vacuum
./bin/clawlcm maintain --op vacuum

# Create backup
./bin/clawlcm maintain --op backup

# Health check
./bin/clawlcm maintain --op doctor

# Clean large files
./bin/clawlcm maintain --op clean

# Session rotation
./bin/clawlcm maintain --op rotate
```

### tui ⚠️ STUB (Not Recommended)
Interactive TUI mode (stub implementation, not available yet).

> **Note**: TUI is marked as "Not Recommended", CLI already supports all operations.

```bash
# This command is currently not available
./bin/clawlcm tui
```

## Configuration

### Config File Parameters

#### database

| Parameter | Default | Description |
|-----------|---------|-------------|
| `database.path` | `./data/clawlcm.db` | Database path |

#### llm (LLM Config) ⭐ NEW

| Parameter | Default | Description |
|-----------|---------|-------------|
| `llm.model` | - | LLM model **Required** |
| `llm.provider` | `openai` | LLM provider |
| `llm.apiKey` | - | API key |
| `llm.baseURL` | - | API endpoint (without /v1) **Required** |
| `llm.timeoutMs` | 120000 | Request timeout (ms) |

#### context (Context Compression) ⭐ NEW FIELDS

| Parameter | Default | Description |
|-----------|---------|-------------|
| `context.threshold` | 0.75 | Compression threshold (0.0-1.0) |
| `context.freshTailCount` | 8 | Protected recent message count |
| `context.useCJKTokenizer` | true | Enable Chinese tokenization |
| `context.maxDepth` | 8 | Max DAG depth (deprecated, use incrementalMaxDepth) |
| `context.condensedMinFanout` | 4 | **⭐ NEW** Min fanout for Leaf condensation |
| `context.incrementalMaxDepth` | 1 | **⭐ NEW** Incremental max depth |
| `context.proactiveThresholdCompactionMode` | `deferred` | **⭐ NEW** Proactive mode (deferred/immediate) |
| `context.maintenanceDebtEnabled` | true | **⭐ NEW** Maintenance debt enabled |
| `context.maintenanceDebtThreshold` | 50000 | **⭐ NEW** Maintenance debt threshold |
| `context.largeFilesDir` | `./data/large_files` | **⭐ NEW** Large files directory |
| `context.cacheAwareCompaction` | false | **⭐ NEW** Cache-aware compaction |
| `context.leafChunkTokens` | 20000 | Leaf compression chunk size |

#### session (Session) ⭐ NEW

| Parameter | Default | Description |
|-----------|---------|-------------|
| `session.ignoreSessionPatterns` | [] | **⭐ NEW** Ignore session patterns |
| `session.statelessSessionPatterns` | [] | **⭐ NEW** Stateless session patterns |
| `session.skipStatelessSessions` | false | **⭐ NEW** Skip stateless sessions |

#### Top-level

| Parameter | Default | Description |
|-----------|---------|-------------|
| `enabled` | true | Enable LCM |
| `verbose` | false | Verbose output |

### Command Line Parameters

#### Common

| Parameter | Description |
|------------|-------------|
| `--config` | Config file path |
| `--db` | Database path |
| `--llm-model` | LLM model |
| `--llm-provider` | LLM provider |
| `--llm-api-key` | API key |
| `--llm-base-url` | API endpoint |
| `--llm-timeout` | Request timeout (ms) |
| `-v` | Verbose output |
| `--version` | Show version |

#### Session

| Parameter | Description |
|------------|-------------|
| `--session-key` | Session key (required) |
| `--session-id` | Session ID |
| `--token-budget` | Token budget (default 128000) |
| `--force` | Force execution |

### Recommended Configuration

```json
{
  "llm": {
    "model": "minimax_m2.5",
    "provider": "openai",
    "apiKey": "",
    "baseURL": "http://YOUR_LLM_SERVER:PORT",
    "timeoutMs": 120000
  },
  "context": {
    "threshold": 0.75,
    "freshTailCount": 64,
    "leafChunkTokens": 20000,
    "maxDepth": 8
  }
}
```

> **⚠️ Security Note**: Replace sensitive info (apiKey, baseURL) with actual values, do not commit to version control

- **freshTailCount=64**: Protect last 64 messages for better conversation continuity
- **leafChunkTokens=20000**: Control Leaf compression chunk size
- **threshold=0.75**: Trigger compression when context reaches 75%
- **maxDepth=8**: Maximum DAG compression depth

## Feature Matrix

| Feature | Description |
|---------|-------------|
| Message Persistence | ✅ SQLite storage |
| Context Assembly | ✅ Summary + Fresh Tail |
| Leaf Summary | ✅ (LLM-driven) |
| Condensed Summary | ✅ (DAG) |
| BM25 Retrieval | ✅ |
| Chinese Tokenization | ✅ |
| DAG Depth Tracking | ✅ |
| CJK Token Estimation | ✅ 1.5x CJK, 2x Emoji |
| Auth Error Filtering | ✅ |
| Session Rotation | ✅ |
| Empty Message Skip | ✅ |

## Output Format

JSON format for programmatic use.

## Technical Specs

| Metric | Value |
|--------|-------|
| Target RAM | <50 MB (excluding LLM) |
| Dependencies | SQLite + GORM |
| Binary Size | ~15 MB |
| Language | Go 1.22+ |

## Source

Based on [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced), this is the LCM (Lossless Context Management) plugin for OpenClaw.
This is an enhanced version of picolcm with multiple fixes and optimizations.

## References

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - Original project
- [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced) - Enhanced version
- [LCM Paper](https://papers.voltropy.com/LCM) - Technical paper