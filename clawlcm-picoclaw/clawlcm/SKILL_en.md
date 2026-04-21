---
name: clawlcm
description: Lossless Context Management - LLM-driven summarization, BM25 retrieval, DAG compression, CJK support. Based on lossless-claw-enhanced.
metadata: {"nanobot":{"emoji":"🧠","requires":{"bins":["clawlcm"]},"install":[{"id":"manual","kind":"binary","label":"Copy clawlcm binary to skill bin directory"}],"version":"v0.8.9"}}
---

# clawlcm Skill

> **Version**: v0.8.9 | **Updated**: 2026-04-17 | **Based on**: lossless-claw v0.9.1 + lossless-claw-enhanced

## What is LCM?

LCM (Lossless Context Management) preserves every message while keeping active context within token limits through DAG-structured summarization.

```
Workflow:
1. Persistence  ──→  Messages stored in SQLite
2. Summarization  ──→  Compress to Leaf Summary
3. Condensation  ──→  Multiple Leaves condensed
4. Assembly  ──→  Summary + Fresh Tail = Context
```

## ⚠️ Important: Required Before Use

> **Before using clawlcm, complete these steps:**
> 1. **Configure LLM** - Set LLM server in `config.json` (model, API Key, Base URL)
> 2. **Run bootstrap** - Execute bootstrap command to create database

### Step 1: Config File `config.json`

```json
{
  "database": { "path": "" },
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
    "largeFilesDir": "",
    "cacheAwareCompaction": false,
    "leafChunkTokens": 20000
  },
  "session": {
    "ignoreSessionPatterns": [],
    "statelessSessionPatterns": [],
    "skipStatelessSessions": false
  },
  "enabled": true,
  "verbose": false
}
```

> ⚠️ **Note**: Do NOT include actual API keys or URLs in version control.

### Step 2: Initialize Database

```bash
cd clawlcm/data
../bin/clawlcm bootstrap --session-key "user:default:001"
```

## Commands

### bootstrap

Initialize conversation.

```bash
./bin/clawlcm bootstrap \
  --session-key user:chat:123 \
  --session-id uuid-123 \
  --token-budget 128000 \
  --messages '[{"role":"user","content":"Hello"}]'
```

### ingest

Add message.

```bash
./bin/clawlcm ingest \
  --session-key user:chat:123 \
  --role user \
  --content "Explain Go goroutines"
```

### assemble

Assemble context.

```bash
./bin/clawlcm assemble \
  --session-key user:chat:123 \
  --token-budget 128000
```

### compact

Trigger LLM summarization.

```bash
./bin/clawlcm compact \
  --session-key user:chat:123 \
  --force
```

### grep ⭐ Important

BM25 search.

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

> **Note**: Use `--pattern` for grep, NOT `--query`!

### describe

Describe summary.

```bash
./bin/clawlcm describe \
  --session-key user:chat:123 \
  --id 1
```

### expand

Expand summary.

```bash
./bin/clawlcm expand \
  --session-key user:chat:123 \
  --summary-ids "1" \
  --query "query content" \
  --max-depth 3
```

> **Note**: Use `--query` for expand!

### maintain

Run maintenance.

```bash
# GC
./bin/clawlcm maintain --session-key user:chat:123 --maint-op gc

# Vacuum
./bin/clawlcm maintain --op vacuum

# Backup
./bin/clawlcm maintain --op backup

# Doctor
./bin/clawlcm maintain --op doctor

# Clean
./bin/clawlcm maintain --op clean

# Rotate
./bin/clawlcm maintain --op rotate
```

## Config Parameters

### llm

| Parameter | Default | Description |
|-----------|---------|-------------|
| `llm.model` | - | LLM model **Required** |
| `llm.provider` | `openai` | LLM provider |
| `llm.apiKey` | - | API key |
| `llm.baseURL` | - | API endpoint (no /v1) **Required** |
| `llm.timeoutMs` | 120000 | Timeout (ms) |

### context

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
| `context.largeFilesDir` | `""` | Large files directory |
| `context.cacheAwareCompaction` | false | Cache-aware compaction |
| `context.leafChunkTokens` | 20000 | Leaf chunk size |

### session

| Parameter | Default | Description |
|-----------|---------|-------------|
| `session.ignoreSessionPatterns` | [] | Ignore session patterns |
| `session.statelessSessionPatterns` | [] | Stateless session patterns |
| `session.skipStatelessSessions` | false | Skip stateless sessions |

### Top-level

| Parameter | Default | Description |
|-----------|---------|-------------|
| `enabled` | true | Enable LCM |
| `verbose` | false | Verbose output |

### CLI Parameters

#### Global

| Parameter | Description |
|-----------|-------------|
| `--config` | Config file path |
| `--db` | Database path |
| `--llm-model` | LLM model |
| `--llm-api-key` | API key |
| `--llm-base-url` | API endpoint |
| `--llm-timeout` | Timeout (ms) |
| `-v` | Verbose |
| `--version` | Show version |

#### Session

| Parameter | Description |
|-----------|-------------|
| `--session-key` | Session key (required) |
| `--session-id` | Session ID |
| `--token-budget` | Token budget (default 128000) |
| `--force` | Force |

### Recommended Config

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

> ⚠️ **Security Tip**: Replace sensitive values (apiKey, baseURL) with actual values, do NOT commit to version control.

- **freshTailCount=8**: Protect last 8 messages
- **incrementalMaxDepth=1**: Max condensation depth
- **condensedMinFanout=4**: Min children for Leaf condensation
- **leafChunkTokens=20000**: Leaf chunk size
- **threshold=0.75**: Trigger compaction at 75%

## Feature Comparison

| Feature | Description |
|--------|-------------|
| Message Persistence | ✅ SQLite storage |
| Context Assembly | ✅ Summary + Fresh Tail |
| Leaf Summary | ✅ (LLM-driven) |
| Condensed Summary | ✅ (DAG) |
| BM25 Search | ✅ |
| Chinese Tokenizer | ✅ |
| Grep Command | ✅ Full-text search |
| Describe Command | ✅ Summary details |
| Expand Command | ✅ Expand content |
| Maintain Tools | ✅ gc/vacuum/backup/doctor/clean/rotate |

## Common Errors

| Error | Cause | Solution |
|------|------|----------|
| `session-key is required` | No session specified | Add `--session-key user:chat:123` |
| `Error: --id is required` | describe missing ID | Add `--id 1` |
| `grep` returns 0 results | Used `--query` | Use `--pattern` for grep |

## Version History

| Version | Date | Changes |
|--------|------|---------|
| v0.8.9 | 2026-04-17 | Fix --help version, Keywords fill, Tokenizer CJK, Grep |
| v0.8.8 | 2026-04-17 | Fix maintainGC, JSON unmarshal, bubble sort |
| v0.8.7 | 2026-04-16 | Fix grep -mode parameter |
| v0.8.6 | 2026-04-16 | Regex search support |
| v0.8.5 | 2026-04-16 | Fix maintain backup/vacuum/clean |
| v0.8.1 | 2026-04-15 | Enhanced release |

## Reference

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw)
- [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced)
- [LCM Paper](https://papers.voltropy.com/LCM)

## License

MIT License
