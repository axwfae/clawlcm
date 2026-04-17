---
name: clawlcm
description: Lossless Context Management - LLM-driven summarization, BM25 retrieval, DAG compression, CJK support.
metadata: {nanobot:{emoji:🧠,requires:{bins:[clawlcm]},install:[{id:manual,kind:binary,label:Copy clawlcm binary to skill bin directory}],version:"v0.8.9"}}
---

# clawlcm Skill

> **Version**: v0.8.9 | **Updated**: 2026-04-17

## Installation

Copy `clawlcm` directory to picoclaw skills:

```bash
cp -r clawlcm /path/to/picoclaw/workspace/skills/
```

## File Structure

```
clawlcm/
└── clawlcm/
    ├── bin/
    │   └── clawlcm          # Binary
    ├── data/
    │   └── config.json     # Config
    └── SKILL.md         # Skill definition
```

## Configuration

Edit `clawlcm/data/config.json` to set LLM server.

See `clawlcm/SKILL.md` for details.