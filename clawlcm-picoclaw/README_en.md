# clawlcm-picoclaw (English)

This document explains how to use clawlcm as a Skill in a Container environment.

> **Version**: v0.8.2 | **Updated**: 2026-04-16 | **Based on**: lossless-claw v0.9.1 + lossless-claw-enhanced | **Porting Tools**: OpenCode + Oh-My-OpenAgent + MiniMax M2.5

---

## Overview

clawlcm is an ultra-lightweight lossless context management system based on **lossless-claw v0.9.1** plus **lossless-claw-enhanced** fixes.

When conversations exceed the model's context window, traditional methods truncate old messages. LCM uses a DAG-structured summarization system that preserves every message while keeping the active context within the model's token limit.

```
Workflow:
1. Persist   ──→  Store all messages in SQLite
2. Summarize ──→  Compress old messages into Leaf Summary
3. Condense  ──→  Merge multiple Leaves into higher-level nodes
4. Assemble ──→  Summary + Fresh Tail = Full Context
```

---

## Porting Source

clawlcm was ported from:

1. **lossless-claw v0.9.1** - Original project (Martian-Engineering)
2. **lossless-claw-enhanced** - CJK Token fixes + upstream bug fixes (win4r fork)

This project was ported using:

- **OpenCode** - AI Development Environment
- **Oh-My-OpenAgent** - AI Agent Orchestration Framework
- **MiniMax M2.5** - LLM Model

---

## Files to Copy

| File | Path | Description |
|------|------|------|
| Binary | `skill/bin/clawlcm` | Main binary |
| Skill Definition | `skill/SKILL.md` | Skill configuration |
| Config Example | `skill/data/config.json` | Configuration file template |

---

## Deployment Steps

### 1. Prepare Binary

Build from source:

```bash
cd clawlcm
make build
cp clawlcm ../clawlcm-picoclaw/clawlcm/bin/
chmod +x ../clawlcm-picoclaw/clawlcm/bin/clawlcm
```

### 2. Configure config.json

Edit `clawlcm/data/config.json`:

```json
{
  "database": {
    "path": "./data/clawlcm.db"
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
    "freshTailCount": 64,
    "useCJKTokenizer": true,
    "condensedMinFanout": 4,
    "incrementalMaxDepth": 1
  }
}
```

> ⚠️ **Note**: Do NOT include `/v1` at the end of `baseURL`

### 3. Run bootstrap to Create Database

```bash
cd clawlcm/data
../bin/clawlcm bootstrap --session-key "user:default:001"
```

> ⚡ **Important**: Must run bootstrap before first use to create database and session structure!

### 4. Copy to Workspace

In container:

```bash
mkdir -p /app/workspace/skills/lcm
cp -r /path/to/clawlcm/* /app/workspace/skills/lcm/
```

---

## Comparison

### vs lossless-claw

| Item | lossless-claw | clawlcm |
|------|---------------|---------|
| Version | v0.9.1 | **v0.8.1** |
| CJK Token | - | **1.5x CJK, 2x Emoji** |
| Auth Error Filter | - | **✅** |
| Session Rotation | - | **✅** |
| Empty Message Skip | - | **✅** |

### vs picolcm-picoclaw

| Item | picolcm-picoclaw | clawlcm-picoclaw |
|------|-----------------|-----------------|
| Version | v0.3.1 | **v0.8.1** |
| Based on | lossless-claw | **lossless-claw v0.9.1 + enhanced** |

---

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

---

## Usage

### Initialize Conversation

```bash
./bin/clawlcm bootstrap \
  --session-key "user:chat:123" \
  --session-id "uuid-123" \
  --token-budget 128000 \
  --messages '[{"role":"user","content":"Hello"}]'
```

### Add Message

```bash
./bin/clawlcm ingest \
  --session-key "user:chat:123" \
  --role user \
  --content "What is Go?"
```

### Assemble Context

```bash
./bin/clawlcm assemble \
  --session-key "user:chat:123" \
  --token-budget 128000
```

### Trigger LLM Summarization

```bash
./bin/clawlcm compact \
  --session-key "user:chat:123" \
  --force
```

### Maintenance

```bash
./bin/clawlcm maintain \
  --session-key "user:chat:123"
```

---

## New Features

### CLI Commands

| Command | Description |
|---------|-----------|
| `grep` | BM25 message search |
| `describe` | Describe summary details |
| `expand` | Expand summary content |
| `maintain --op` | Maintenance operations (gc/vacuum/backup/doctor/clean/rotate) |

---

## Configuration

### Config File Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `database.path` | `./data/clawlcm.db` | Database path |
| `llm.model` | - | LLM model **Required** |
| `llm.baseURL` | - | API endpoint (without /v1) **Required** |
| `llm.apiKey` | - | API key |
| `context.threshold` | 0.75 | Compression threshold (0.0-1.0) |
| `context.freshTailCount` | 8 | Protected recent message count |
| `context.useCJKTokenizer` | true | Enable Chinese tokenization |
| `context.condensedMinFanout` | 4 | Min fanout for Leaf condensation |
| `context.incrementalMaxDepth` | 1 | Incremental max depth |
| `context.maintenanceDebtEnabled` | true | Maintenance debt enabled |

### Recommended Configuration

```json
{
  "context": {
    "threshold": 0.75,
    "freshTailCount": 64,
    "leafChunkTokens": 20000,
    "maxDepth": 8
  }
}
```

---

## Docker Volume Configuration

```yaml
services:
  app:
    image: app:latest
    volumes:
      - ./skills:/app/workspace/skills
      - lcm-data:/app/data
    environment:
      - TZ=UTC

volumes:
  lcm-data:
```

---

## File List

```
clawlcm-picoclaw/
├── README.md              # This file
├── README_en.md          # English version
└── clawlcm/              # Skill deployment package
    ├── SKILL.md          # Skill definition
    ├── SKILL_en.md      # Skill definition (English)
    ├── data/
    │   └── config.json   # Configuration file
    └── bin/
        └── clawlcm       # Main binary (~15MB)
```

---

## Version History

| Version | Date | Changes |
|---------|------|--------|
| v0.8.2 | 2026-04-16 | CLI Fix (describe/expand/maintain params) |
| v0.8.1 | 2026-04-15 | Enhanced release (lossless-claw-enhanced) |
| v0.3.0 | 2026-04-14 | Removed incompatible providers, simplified LLM interface |
| v0.2.0 | 2026-04-14 | JSON config loading |
| v0.1.0 | 2026-04-14 | Initial version |

---

## Test Commands

```bash
# Check version
./bin/clawlcm --version

# Test run
./bin/clawlcm -v
```

---

## Features

| Feature | Description |
|---------|-----------|
| Message Persistence | ✅ SQLite storage |
| Context Assembly | ✅ Summary + Fresh Tail |
| Leaf Summary | ✅ (LLM-driven) |
| Condensed Summary | ✅ (DAG multi-level) |
| BM25 Retrieval | ✅ |
| Chinese Tokenization | ✅ |
| DAG Depth Tracking | ✅ |
| CJK Token Estimation | ✅ 1.5x CJK, 2x Emoji |
| Auth Error Filter | ✅ |
| Session Rotation | ✅ |
| Empty Message Skip | ✅ |
| CLI --flags Fix | ✅ |

---

## Technical Specs

| Metric | Value |
|--------|-------|
| Target RAM | <50 MB (excluding LLM) |
| Dependencies | SQLite + GORM |
| Binary Size | ~15 MB |
| Language | Go 1.22+ |

---

## References

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - Original project
- [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced) - Enhanced version
- [LCM Paper](https://papers.voltropy.com/LCM) - Technical paper
- [OpenCode](https://opencode.com) - AI Development Environment
- [Oh-My-OpenAgent](https://github.com/oh-my-openagent) - AI Agent Orchestration Framework