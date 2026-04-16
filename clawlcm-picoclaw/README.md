# clawlcm-picoclaw

本文檔說明如何在 Container 環境中使用 clawlcm 作為 Skill。

> **版本**: v0.8.2 | **更新日**: 2026-04-16 | **基於**: lossless-claw v0.9.1 + lossless-claw-enhanced | **移植工具**: OpenCode + Oh-My-OpenAgent + MiniMax M2.5

---

## 概述

clawlcm 是基於 **lossless-claw v0.9.1** 加上 **lossless-claw-enhanced** 修復點移植的超輕量級無損上下文管理系統。

當對話超過模型的上下文窗口時，傳統方法會截斷舊訊息。LCM 採用 DAG 結構的摘要系統，保留每條訊息，同時將活躍上下文保持在模型的 token 限制內。

```
工作流程:
1. 持久化  ──→  所有訊息存入 SQLite
2. 摘要    ──→  舊訊息壓縮成 Leaf Summary  
3. 凝聚    ──→  多個 Leaf 凝聚成更高層節點
4. 組裝    ──→  摘要 + 新鮮尾部 = 完整上下文
```

---

## 移植來源

clawlcm 基於以下專案移植：

1. **lossless-claw v0.9.1** - 原始專案 (Martian-Engineering)
2. **lossless-claw-enhanced** - CJK Token 修復 + 上游 Bug 修復 (win4r fork)

本專案使用以下工具移植：

- **OpenCode** - AI 開發環境
- **Oh-My-OpenAgent** - AI Agent 編排框架
- **MiniMax M2.5** - LLM 模型

---

## 需要複製的文件

| 文件 | 路徑 | 說明 |
|------|------|------|
| 二進制 | `skill/bin/clawlcm` | 主程式 |
| Skill 定義 | `skill/SKILL.md` | Skill 配置 |
| 配置示例 | `skill/data/config.json` | 配置文件示例 |

---

## 部署步驟

### 1. 準備二進制文件

從源碼編譯：

```bash
cd clawlcm
make build
cp clawlcm ../clawlcm-picoclaw/clawlcm/bin/
chmod +x ../clawlcm-picoclaw/clawlcm/bin/clawlcm
```

### 2. 配置 config.json

編輯 `clawlcm/data/config.json`：

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

> ⚠️ **注意**: `baseURL` 不要包含 `/v1` 结尾

### 3. 複製到 workspace

在容器中：

```bash
mkdir -p /app/workspace/skills/lcm
cp -r /path/to/clawlcm/* /app/workspace/skills/lcm/
```

---

## 對照

### 與 lossless-claw 對照

| 項目 | lossless-claw | clawlcm |
|------|---------------|---------|
| 版本 | v0.9.1 | **v0.8.1** |
| CJK Token | - | **1.5x CJK, 2x Emoji** |
| Auth Error | - | **✅ 過濾** |
| Session Rotation | - | **✅ 檢測** |
| 空訊息 | - | **✅ 跳過** |

### 與 picolcm-picoclaw 對照

| 項目 | picolcm-picoclaw | clawlcm-picoclaw |
|------|-----------------|-----------------|
| 版本 | v0.3.1 | **v0.8.1** |
| 基於 | lossless-claw | **lossless-claw v0.9.1 + enhanced** |
| Token 估算 | 標準 | **1.5x CJK, 2x Emoji** |
| Auth Error 過濾 | - | **✅** |
| Session Rotation | - | **✅** |
| 空訊息跳過 | - | **✅** |
| MaxDepth | - | **✅** |
| CLI --flags 修復 | - | **✅** |

---

## DAG 結構

```
         [訊息 1-15]                    ← 原始訊息 (已壓縮)
               │
               ▼
        [Leaf Summary]                 ← Leaf 摘要 (深度=0)
               │
               ▼
    ┌──────────┴──────────┐
    ▼                     ▼
[Leaf A]             [Leaf B]          ← 多個 Leaf 凝聚
    │                     │
    └──────────┬──────────┘
               ▼
       [Condensed Summary]             ← 凝聚摘要 (深度=1)
               ���
               ▼
     [Fresh Tail: 最後 N 條訊息]        ← 受保護的新鮮尾部
```

---

## 使用方法

### 初始化對話

```bash
./bin/clawlcm bootstrap \
  --session-key "user:chat:123" \
  --session-id "uuid-123" \
  --token-budget 128000 \
  --messages '[{"role":"user","content":"你好"}]'
```

### 添加訊息

```bash
./bin/clawlcm ingest \
  --session-key "user:chat:123" \
  --role user \
  --content "什麼是 Go？"
```

### 組裝上下文

```bash
./bin/clawlcm assemble \
  --session-key "user:chat:123" \
  --token-budget 128000
```

### 觸發 LLM 摘要

```bash
./bin/clawlcm compact \
  --session-key "user:chat:123" \
  --force
```

### 維護

```bash
./bin/clawlcm maintain \
  --session-key "user:chat:123"
```

---

## 新增功能

### CLI 命令

| 命令 | 說明 |
|------|------|
| `grep` | BM25 檢索訊息 |
| `describe` | 描述摘要詳情 |
| `expand` | 展開摘要內容 |
| `maintain --op` | 維護操作 (gc/vacuum/backup/doctor/clean/rotate) |

---

## 配置說明

### 配置文件參數

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `database.path` | `./data/clawlcm.db` | 資料庫路徑 |
| `llm.model` | - | LLM 模型 **必填** |
| `llm.baseURL` | - | API 端點 (不含 /v1) **必填** |
| `llm.apiKey` | - | API 金鑰 |
| `context.threshold` | 0.75 | 壓縮閾值 (0.0-1.0) |
| `context.freshTailCount` | 8 | 保護的最近訊息數 |
| `context.useCJKTokenizer` | true | 啟用中文分詞 |
| `context.condensedMinFanout` | 4 | Leaf 凝聚最小子節點數 |
| `context.incrementalMaxDepth` | 1 | 遞進壓縮最大深度 |
| `context.maintenanceDebtEnabled` | true | 維護Debt啟用 |

### 推薦配置

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

## Docker Volume 配置

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

## 文件清單

```
clawlcm-picoclaw/
├── README.md              # 本文件
└── clawlcm/              # Skill 部署包
    ├── SKILL.md          # Skill 定義
    ├── SKILL_en.md      # Skill 定義 (英文)
    ├── data/
    │   └── config.json   # 配置文件
    └── bin/
        └── clawlcm       # 主程式 (~15MB)
```

---

## 版本信息

| 版本 | 日期 | 變更 |
|------|------|------|
| v0.8.2 | 2026-04-16 | CLI 修復 (describe/expand/maintain 參數) |
| v0.8.1 | 2026-04-15 | 增強版發布 (lossless-claw-enhanced) |
| v0.3.0 | 2026-04-14 | 移除不兼容 provider，簡化 LLM 接口 |
| v0.2.0 | 2026-04-14 | JSON 配置加載 |
| v0.1.0 | 2026-04-14 | 初始版本 |

---

## 測試命令

```bash
# 檢查版本
./bin/clawlcm --version

# 測試運行
./bin/clawlcm -v
```

---

## 功能說明

| 功能 | 說明 |
|------|------|
| 訊息持久化 | ✅ SQLite 儲存 |
| 上下文組裝 | ✅ 摘要 + 新鮮尾部 |
| Leaf 摘要 | ✅ (LLM 驅動) |
| Condensed 摘要 | ✅ (DAG 多層) |
| BM25 檢索 | ✅ |
| 中文分詞 | ✅ |
| DAG 深度追蹤 | ✅ |
| CJK Token 估算 | ✅ 1.5x CJK, 2x Emoji |
| Auth Error 過濾 | ✅ |
| Session Rotation | ✅ |
| 空訊息跳過 | ✅ |
| CLI --flags 修復 | ✅ |

---

## 技術規格

| 指標 | 數值 |
|------|------|
| 目標 RAM | <50 MB (不含 LLM) |
| 依賴 | SQLite + GORM |
| 二進制大小 | ~15 MB |
| 語言 | Go 1.22+ |

---

## 參考

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - 原始項目
- [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced) - 增強版本
- [LCM Paper](https://papers.voltropy.com/LCM) - 技術論文
- [OpenCode](https://opencode.com) - AI 開發環境
- [Oh-My-OpenAgent](https://github.com/oh-my-openagent) - AI Agent 編排框架