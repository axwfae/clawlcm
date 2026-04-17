# clawlcm - Lossless Context Management

clawlcm is a Go implementation of a lossless context management system based on lossless-claw.

## Project Info

| Item | Value |
|------|-------|
| Project Name | clawlcm |
| Based on | lossless-claw v0.9.1 + lossless-claw-enhanced |
| Current Version | v0.8.9 |
| Language | Go 1.22+ |

## Overview

LCM (Lossless Context Management) is a lossless context management system designed for large language model (LLM) applications. When conversations exceed the model's context window, traditional methods truncate old messages. LCM uses a DAG-structured summarization system that preserves every message while keeping the active context within the model's token limit.

```
Workflow:
1. Persistence  ──→  Messages stored in SQLite
2. Summarization  ──→  Messages compressed to Leaf Summary
3. Condensation  ──→  Multiple Leaves condensed to higher-level nodes
4. Assembly  ──→  Summary + Fresh Tail = Complete Context
```

## Core Features

| Feature | Description | Status |
|--------|-------------|--------|
| Message Persistence | SQLite storage | ✅ |
| Context Assembly | Summary + Fresh Tail | ✅ |
| Leaf Summary | LLM-driven summarization | ✅ |
| Condensed Summary | DAG multi-level condensation | ✅ |
| BM25 Search | Relevance search | ✅ |
| Chinese Tokenizer | CJK support | ✅ |
| Grep Command | BM25 full-text search | ✅ |
| Describe Command | Summary details | ✅ |
| Expand Command | Expand summary content | ✅ |
| Maintain | Maintenance tools | ✅ |

## Technical Specs

| Metric | Value |
|-------|-------|
| Target RAM | <50 MB (without LLM) |
| Dependencies | SQLite + GORM |
| Binary Size | ~15 MB |
| Language | Go 1.22+ |

## Commands

### bootstrap

Initialize conversation with optional initial messages.

```bash
./clawlcm bootstrap \
  --session-key "user:chat:123" \
  --session-id "uuid-123" \
  --token-budget 128000 \
  --messages '[{"role":"user","content":"Hello"}]'
```

**Parameters:**
| Parameter | Description | Default |
|-----------|-------------|---------|
| `--session-key` | Session key (format: user:chat:123) | - |
| `--session-id` | Session ID (UUID) | auto-generated |
| `--token-budget` | Token budget | 128000 |
| `--messages` | JSON message array | [] |

---

### ingest

Add message to conversation.

```bash
./clawlcm ingest \
  --session-key "user:chat:123" \
  --role user \
  --content "What is Go language?"
```

**Parameters:**
| Parameter | Description | Default |
|-----------|-------------|---------|
| `--session-key` | Session key | - |
| `--role` | Role | user |
| `--content` | Message content | - |

---

### assemble

Assemble context (summary + fresh tail).

```bash
./clawlcm assemble \
  --session-key "user:chat:123" \
  --token-budget 128000
```

**Parameters:**
| Parameter | Description | Default |
|-----------|-------------|---------|
| `--session-key` | Session key | - |
| `--token-budget` | Token budget | 128000 |

---

### compact

Trigger LLM summarization.

```bash
./clawlcm compact \
  --session-key "user:chat:123" \
  --force
```

**Parameters:**
| Parameter | Description | Default |
|-----------|-------------|---------|
| `--session-key` | Session key | - |
| `--force` | Force compaction | false |

---

### grep ⭐ Important

BM25 full-text search.

```bash
# Search single conversation
./clawlcm grep \
  --session-key "user:chat:123" \
  --pattern "Go language"

# Search all conversations
./clawlcm grep \
  --all \
  --pattern "error" \
  --limit 20

# Use regex
./clawlcm grep \
  --session-key "user:chat:123" \
  --pattern "error.*fail" \
  --mode regex
```

**Parameters:**
| Parameter | Description | Default |
|-----------|-------------|---------|
| `--session-key` | Session key | - |
| `--all` | Search all sessions | false |
| `--pattern` | **Search pattern** | - |
| `--mode` | Search mode (full_text/regex) | full_text |
| `--scope` | Scope (all/messages/summaries) | all |
| `--limit` | Result count | 20 |
| `--sort` | Sort order (desc/asc) | desc |

> **Note**: Use `--pattern` for grep, NOT `--query`!

---

### describe

Describe summary details.

```bash
./clawlcm describe \
  --session-key "user:chat:123" \
  --id 1
```

**Parameters:**
| Parameter | Description | Default |
|-----------|-------------|---------|
| `--session-key` | Session key | - |
| `--id` | Summary ID | - |
| `--all` | All sessions | false |

---

### expand

Expand summary content.

```bash
./clawlcm expand \
  --session-key "user:chat:123" \
  --summary-ids "1,2,3" \
  --query "Go language features" \
  --max-depth 3 \
  --include-messages
```

**Parameters:**
| Parameter | Description | Default |
|-----------|-------------|---------|
| `--session-key` | Session key | - |
| `--summary-ids` | Summary IDs (comma-separated) | - |
| `--query` | Query content | - |
| `--max-depth` | Max depth | 3 |
| `--include-messages` | Include messages | false |

> **Note**: Use `--query` for expand, NOT `--pattern`!

---

### maintain

Run maintenance tasks.

```bash
# GC
./clawlcm maintain --session-key "user:chat:123" --maint-op gc

# Vacuum
./clawlcm maintain --op vacuum

# Backup
./clawlcm maintain --op backup

# Doctor
./clawlcm maintain --op doctor

# Clean
./clawlcm maintain --op clean

# Rotate
./clawlcm maintain --op rotate
```

**Parameters:**
| Parameter | Description | Default |
|-----------|-------------|---------|
| `--session-key` | Session key | - |
| `--maint-op` | Maintenance operation | gc |
| `--all` | All sessions | false |

**Operations:**
| Operation | Description |
|-----------|-------------|
| `gc` | Garbage collection (orphaned summaries) |
| `vacuum` | Database optimization |
| `backup` | Create backup |
| `doctor` | Health check |
| `clean` | Clean large files |
| `rotate` | Session Rotation |

---

## Configuration

Config file `./data/config.json`:

```json
{
  "database": {
    "path": ""
  },
  "llm": {
    "model": "minimax_m2.5",
    "provider": "openai",
    "apiKey": "",
    "baseURL": "http://YOUR_LLM_SERVER:PORT",
    "timeoutMs": 120000
  },
  "context": {
    "threshold": 0.75,
    "freshTailCount": 8,
    "useCJKTokenizer": true,
    "condensedMinFanout": 4,
    "incrementalMaxDepth": 1,
    "proactiveThresholdCompactionMode": "deferred",
    "maintenanceDebtEnabled": true,
    "maintenanceDebtThreshold": 50000,
    "leafChunkTokens": 20000
  }
}
```

> ⚠️ **Note**: `baseURL` must NOT include `/v1` suffix

### Config Parameters

#### llm

| Parameter | Default | Description |
|-----------|---------|-------------|
| `llm.model` | - | LLM model **Required** |
| `llm.provider` | `openai` | LLM provider |
| `llm.apiKey` | - | API key |
| `llm.baseURL` | - | API endpoint (no /v1) **Required** |
| `llm.timeoutMs` | 120000 | Timeout (ms) |

#### context

| Parameter | Default | Description |
|-----------|---------|-------------|
| `context.threshold` | 0.75 | Compaction threshold (0.0-1.0) |
| `context.freshTailCount` | 8 | Protected recent messages |
| `context.useCJKTokenizer` | true | Enable Chinese tokenization |
| `context.condensedMinFanout` | 4 | Min children for condensation |
| `context.incrementalMaxDepth` | 1 | Max incremental depth |
| `context.proactiveThresholdCompactionMode` | `deferred` | Compaction mode (deferred/immediate) |
| `context.maintenanceDebtEnabled` | true | Maintenance debt enabled |
| `context.maintenanceDebtThreshold` | 50000 | Maintenance debt threshold |
| `context.leafChunkTokens` | 20000 | Leaf chunk size |

#### session

| Parameter | Default | Description |
|-----------|---------|-------------|
| `session.ignoreSessionPatterns` | [] | Ignore session patterns |
| `session.statelessSessionPatterns` | [] | Stateless session patterns |
| `session.skipStatelessSessions` | false | Skip stateless sessions |

### CLI Global Parameters

| Parameter | Description |
|-----------|-------------|
| `--config` | Config file path |
| `--db` | Database path |
| `--llm-model` | LLM model |
| `--llm-provider` | LLM provider |
| `--llm-api-key` | API key |
| `--llm-base-url` | API endpoint |
| `--llm-timeout` | Request timeout (ms) |
| `-v` | Verbose output |
| `--version` | Show version |

## Recommended Config

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
    "freshTailCount": 8,
    "incrementalMaxDepth": 1,
    "condensedMinFanout": 4,
    "leafChunkTokens": 20000
  }
}
```

- **freshTailCount=8**: Protect last 8 messages
- **incrementalMaxDepth=1**: Max condensation depth
- **condensedMinFanout=4**: Min children for Leaf condensation
- **leafChunkTokens=20000**: Leaf chunk size
- **threshold=0.75**: Trigger compaction at 75%

## Common Errors

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `session-key is required` | No session specified | Add `--session-key user:chat:123` |
| `Error: --id is required` | describe missing ID | Add `--id 1` |
| `Error: --summary-ids is required` | expand missing summary ID | Add `--summary-ids 1` |
| `grep` returns 0 results | Used `--query` | Use `--pattern` for grep |

## Project Structure

```
clawlcm/
├── cmd/              # CLI entry
├── engine.go         # Engine core
├── store/            # Data store
├── retrieval/        # BM25 retrieval
├── tokenizer/        # Chinese tokenization
├── llm/              # LLM client
├── db/               # Database connection
├── types/            # Type definitions
├── logger/           # Logger
├── docker/           # Docker config
└── Makefile          # Build script
```

## Version History

| Version | Date | Changes |
|--------|------|---------|
| v0.8.9 | 2026-04-17 | Fix --help version, Keywords fill, Tokenizer CJK, Grep |
| v0.8.8 | 2026-04-17 | Fix maintainGC, JSON unmarshal, bubble sort |
| v0.8.7 | 2026-04-16 | Fix grep -mode parameter |

## Reference

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw)
- [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced)
- [LCM Paper](https://papers.voltropy.com/LCM)

## License

MIT License