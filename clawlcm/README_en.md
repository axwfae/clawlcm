# clawlcm - Lossless Context Management

clawlcm is a lossless context management system based on lossless-claw v0.9.1 + lossless-claw-enhanced.

> **Version**: v0.8.7 | **Updated**: 2026-04-16 | **Porting Tools**: OpenCode + Oh-My-OpenAgent + MiniMax M2.5

---

## Project Info

| Item | Value |
|------|-------|
| Project Name | clawlcm |
| Based on | lossless-claw v0.9.1 + lossless-claw-enhanced |
| Original Project | lossless-claw |
| Current Version | **v0.8.4** |

---

## Porting Note

This project was ported using:

- **OpenCode** - AI Development Environment
- **Oh-My-OpenAgent** - AI Agent Orchestration Framework
- **MiniMax M2.5** - LLM Model

---

## Overview

LCM (Lossless Context Management) is a lossless context management system designed for Large Language Model (LLM) applications. When a conversation grows beyond the model's context window, traditional methods truncate old messages. LCM uses a DAG-based summarization system that preserves every message while keeping active context within model token limits.

---

## Core Features

| Feature | Description | Status |
|---------|-------------|--------|
| Message Persistence | SQLite storage for conversation history | ✅ |
| Context Assembly | Smart assembly of summaries + fresh tail | ✅ |
| LLM Summarization | Leaf/Condensed summary generation | ✅ |
| BM25 Retrieval | Relevance search | ✅ |
| CJK Tokenization | Chinese/Japanese/Korean support | ✅ |
| Deferred Mode | Deferred/inline compaction mode | ✅ |
| Maintenance Debt | Maintenance debt tracking | ✅ |
| LargeFiles Dir | External large file storage | ✅ |
| Rotate Command | Session rotation/rewrite | ✅ |
| Session Filter | ignore/stateless patterns | ✅ |
| Agent Tools | lcm_grep, lcm_expand tools | ✅ |
| Expand Query | Query-based DAG expansion | ✅ |
| CJK Token Estimation | 1.5x CJK, 2x Emoji | ✅ |
| Auth Error Filter | Prevent false-positive errors | ✅ |
| Session Rotation | Session rotation detection | ✅ |
| Empty Message Skip | Skip empty assistant messages | ✅ |
| CLI --flags Fix | Support command --flags format | ✅ |

---

## Technical Specifications

| Metric | Value |
|--------|-------|
| Target RAM | <50 MB (excluding LLM) |
| Dependencies | SQLite + GORM |
| Binary Size | ~15 MB |
| Language | Go 1.22+ |

---

## Quick Start

```bash
# Build
make build

# Initialize conversation (command --flags)
./clawlcm bootstrap \
  --session-key "user:chat:1" \
  --messages '[{"role":"user","content":"Hello"}]'

# Or use --flags before command
./clawlcm --session-key "user:chat:1" bootstrap \
  --messages '[{"role":"user","content":"Hello"}]'

# Add message
./clawlcm ingest \
  --session-key "user:chat:1" \
  --role user \
  --content "What is Go language?"

# Assemble context
./clawlcm assemble \
  --session-key "user:chat:1" \
  --token-budget 128000

# Or use --flags before command
./clawlcm --session-key "user:chat:1" assemble --token-budget 128000

# Trigger LLM summarization
./clawlcm compact \
  --session-key "user:chat:1" \
  --force

# Or use --flags before command
./clawlcm --session-key "user:chat:1" compact --force
```

---

## Configuration

Edit `data/config.json`:

```json
{
  "database": {
    "path": ""
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
    "condensedMinFanout": 4,
    "incrementalMaxDepth": 1,
    "proactiveThresholdCompactionMode": "deferred",
    "maintenanceDebtEnabled": true,
    "maintenanceDebtThreshold": 50000,
    "largeFilesDir": "",
    "cacheAwareCompaction": false
  },
  "session": {
    "ignoreSessionPatterns": [],
    "statelessSessionPatterns": [],
    "skipStatelessSessions": false
  }
}
```

> ⚠️ **Note**: Do NOT include `/v1` suffix in `baseURL`
> **Security Note**: Sensitive info (apiKey, baseURL) should use env vars or CLI args, do not commit to version control

---

## Project Structure

```
clawlcm/
├── bin/clawlcm       # Main binary (symlink to /usr/local/bin/)
├── data/             # Data directory (relative to executable)
│   ├── config.json   # Config file
│   ├── clawlcm.db    # SQLite database
│   └── large_files/  # Large file external storage
├── cmd/              # CLI entry
├── engine.go         # Core engine
├── store/            # Data storage
├── retrieval/        # BM25 retrieval
├── tokenizer/        # Chinese tokenization
├── llm/              # LLM client
├── db/               # Database connection
├── types/            # Type definitions
├── logger/           # Logging
├── docker/           # Docker configuration
├── Makefile          # Build scripts
└── README_en.md      # This file
```

---

## Self-Contained Paths

All paths are based on **parent directory of executable location**, regardless of where you run:
- Database: `data/clawlcm.db`
- Config: `data/config.json`
- Large Files: `data/large_files/`

```bash
# After symlink, data is always placed in the correct location
ln -s /path/to/clawlcm/bin/clawlcm /usr/local/bin/clawlcm
clawlcm --help  # Data will be in /path/to/clawlcm/data/
```

---

## CLI Commands

| Command | Description |
|---------|-----------|
| `bootstrap` | Initialize conversation |
| `ingest` | Add message |
| `assemble` | Assemble context |
| `compact` | Trigger compaction |
| `grep` | BM25 search |
| `describe` | Describe summary |
| `expand` | Expand summary |
| `maintain --op` | Maintenance (gc/vacuum/backup/doctor/clean/rotate) |
| `tui` | ⚠️ Stub (not recommended) |

---

## Version History

| Version | Date | Changes |
|---------|------|--------|
| v0.8.7 | 2026-04-16 | Fix grep -mode parameter parsing |
| v0.8.6 | 2026-04-16 | Implement regex search, fix unused variables |
| v0.8.5 | 2026-04-16 | Fix maintain stub code (backup/vacuum/clean now actually execute) |
| v0.8.4 | 2026-04-16 | Unified documentation consistency, remove ./ prefix |
| v0.8.3 | 2026-04-16 | Self-contained paths + fix largeFilesDir + symlink support |
| v0.8.2 | 2026-04-16 | CLI Fix (describe/expand/maintain params) |
| v0.8.1 | 2026-04-15 | Enhanced release (CJK Token, Auth Error, Session Rotation, Empty Message, CLI --flags) |
| v0.8.0 | 2026-04-14 | Port lossless-claw features (Deferred, Maintenance Debt, LargeFiles, Rotate, etc.) |

---

## Not Recommended

1. **TUI (Interactive Terminal)**
   - **Reason**: 
   - Complete CLI commands already
   - Designed as OpenClaw plugin

2. **FTS5 (Full-Text Search)**
   - **Reason**:
   - BM25 retrieval already implemented
   - FTS5 adds complexity

---

## References

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - Original project
- [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced) - Enhanced version
- [LCM Paper](https://papers.voltropy.com/LCM) - Technical paper
- [OpenCode](https://opencode.com) - AI Development Environment
- [Oh-My-OpenAgent](https://github.com/oh-my-openagent) - AI Agent Orchestration Framework

---

## License

MIT License