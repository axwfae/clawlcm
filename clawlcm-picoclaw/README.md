---
name: clawlcm
description: 無損上下文管理 - LLM 驅動摘要、BM25 檢索、DAG 壓縮，支援中文。基於 lossless-claw-enhanced 增強版。
metadata: {nanobot:{emoji:🧠,requires:{bins:[clawlcm]},install:[{id:manual,kind:binary,label:將 clawlcm 二進制複製到 skill bin 目錄}],version:"v0.8.9"}}
---

# clawlcm Skill

> **版本**: v0.8.9 | **更新日**: 2026-04-17

## 安裝說明

將 `clawlcm` 目錄複製到 picoclaw 的 skills 目錄：

```bash
cp -r clawlcm /path/to/picoclaw/workspace/skills/
```

## 檔案結構

```
clawlcm/
└── clawlcm/
    ├── bin/
    │   └── clawlcm          # 二進制執行檔
    ├── data/
    │   └── config.json     # 配置文件
    └── SKILL.md         # Skill 定義
```

## 配置

編輯 `clawlcm/data/config.json`，設置 LLM 服務器。

更多說明請參考 `clawlcm/SKILL.md`。