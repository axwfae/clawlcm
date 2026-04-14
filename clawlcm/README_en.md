# LCM - Lossless Context Management

Ultra-lightweight lossless context management system based on [lossless-claw](https://github.com/Martian-Engineering/lossless-claw).

## Overview

LCM (Lossless Context Management) is a lossless context management system designed for Large Language Model (LLM) applications. When a conversation grows beyond the model's context window, traditional methods truncate old messages. LCM uses a DAG-based summarization system that preserves every message while keeping active context within model token limits.

## Core Features

| Feature | Description |
|---------|-------------|
| Message Persistence | SQLite storage for conversation history |
| Context Assembly | Smart assembly of summaries + fresh tail |
| LLM Summarization | Leaf/Condensed summary generation |
| BM25 Retrieval | Relevance search |
| Chinese Tokenization | Chinese/Japanese/Korean support |
| DAG Compression | Multi-layer compression tracking |

## How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                    LCM Workflow                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Persist        ──→  All messages stored in SQLite      │
│         ↓                                                    │
│  2. Summarize      ──→  Old messages → Leaf Summary       │
│         ↓                                                    │
│  3. Condense       ──→  Multiple Leafs → Higher-level node  │
│         ↓                                                    │
│  4. Assemble       ──→  Summaries + Fresh Tail = Context   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### DAG Structure

```
         [Message 1-15]                    ← Raw messages (compressed)
               │
               ▼
        [Leaf Summary]                     ← Leaf summary (D=0)
               │
               ▼
    ┌──────────┴──────────┐
    ▼                     ▼
[Leaf A]             [Leaf B]              ← Multiple leaves condensed
    │                     │
    └──────────┬──────────┘
               ▼
       [Condensed Summary]                 ← Condensed summary (D=1)
               │
               ▼
     [Fresh Tail: last N messages]          ← Protected fresh tail
```

- **Leaf Summary**: Compressed summary of raw messages
- **Condensed Summary**: Further condensation of multiple Leaf summaries
- **Fresh Tail**: Last N messages protected from compression

## Technical Specifications

| Metric | Value |
|--------|-------|
| Target RAM | <50 MB (excluding LLM) |
| Dependencies | SQLite + GORM |
| Binary Size | ~15 MB |
| Language | Go 1.22+ |

## Quick Start

### Build

```bash
make build
```

### Configuration

Edit `./data/config.json`:

```json
{
  "database": {
    "path": "./data/picolcm.db"
  },
  "llm": {
    "model": "minimax_m2.5",
    "provider": "openai",
    "apiKey": "your-api-key",
    "baseURL": "http://your-llm-server:18869",
    "timeoutMs": 120000
  },
  "context": {
    "threshold": 0.75,
    "freshTailCount": 8,
    "useCJKTokenizer": true
  }
}
```

> ⚠️ **Note**: Do NOT include `/v1` suffix in `baseURL`. The code automatically appends `/v1/chat/completions`

### Basic Operations

```bash
# Initialize conversation
./picolcm bootstrap \
  --session-key "user:chat:1" \
  --messages '[{"role":"user","content":"Hello"}]'

# Add message
./picolcm ingest \
  --session-key "user:chat:1" \
  --role user \
  --content "What is Go language?"

# Assemble context
./picolcm assemble \
  --session-key "user:chat:1" \
  --token-budget 128000

# Trigger LLM summarization
./picolcm compact \
  --session-key "user:chat:1" \
  --force

# Maintenance (GC, optimization)
./picolcm maintain \
  --session-key "user:chat:1"
```

## Configuration Reference

### Config Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `threshold` | 0.75 | Context threshold that triggers compaction (0.0-1.0) |
| `freshTailCount` | 8 | Number of recent messages protected from compaction |
| `leafChunkTokens` | 20000 | Max tokens per Leaf chunk before summarization |
| `leafTargetTokens` | 2400 | Target token count for Leaf summaries |
| `condensedTargetTokens` | 2000 | Target token count for Condensed summaries |
| `useCJKTokenizer` | true | Enable Chinese tokenization |

### CLI Flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path (default: `./data/config.json`) |
| `--db` | Database path |
| `--llm-model` | LLM model |
| `--llm-api-key` | API key |
| `--llm-base-url` | API endpoint (without /v1) |
| `-v` | Verbose output |
| `--version` | Show version |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `PICOLCM_DB_PATH` | Database path |
| `PICOLCM_SUMMARY_MODEL` | Summarization model |
| `PICOLCM_SUMMARY_BASE_URL` | Summarization API endpoint |
| `PICOLCM_USE_CJK_TOKENIZER` | Enable Chinese tokenization |
| `PICOLCM_CONTEXT_THRESHOLD` | Compaction threshold |
| `PICOLCM_FRESH_TAIL_COUNT` | Protected recent messages count |

## Docker Deployment

```bash
# Build
make docker-build

# Run
make docker-run
```

Or use docker-compose:

```bash
make podman-compose-up
```

## Recommended Configuration

```json
{
  "context": {
    "threshold": 0.75,
    "freshTailCount": 64,
    "leafChunkTokens": 20000
  }
}
```

- **freshTailCount=64**: Protects last 64 messages for better conversation continuity
- **leafChunkTokens=20000**: Controls Leaf compression chunk size
- **threshold=0.75**: Triggers compaction when context reaches 75% of window

## Project Structure

```
picolcm/
├── cmd/picolcm/      # CLI entry point
├── engine.go         # Core engine
├── store/            # Data storage
│   └── store.go      # SQLite operations
├── retrieval/        # BM25 retrieval
│   └── bm25.go       # BM25 algorithm implementation
├── tokenizer/       # Chinese tokenization
│   └── tokenizer.go  # Tokenizer
├── llm/              # LLM client
│   └── client.go     # OpenAI compatible client
├── db/               # Database
│   └── connection.go # GORM connection
├── types/            # Type definitions
│   └── types.go      # Config and request/response types
├── logger/           # Logging
│   └── logger.go    # Logger implementation
├── docker/           # Docker configuration
│   ├── Dockerfile
│   └── docker-compose.yml
└── Makefile          # Build scripts
```

## Testing

```bash
make test
```

## License

MIT License

## References

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - Original project
- [LCM Paper](https://papers.voltropy.com/LCM) - Technical paper
