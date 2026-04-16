---
name: clawlcm
description: 無損上下文管理 - LLM 驅動摘要、BM25 檢索、DAG 壓縮，支援中文。基於 lossless-claw-enhanced 增強版。
metadata: {nanobot:{emoji:🧠,requires:{bins:[clawlcm]},install:[{id:manual,kind:binary,label:將 clawlcm 二進制複製到 skill bin 目錄}],version:"v0.8.7"}}
---

# clawlcm Skill

> **版本**: v0.8.7 | **更新日**: 2026-04-16 | **基於**: lossless-claw v0.9.1 + lossless-claw-enhanced | **移植工具**: OpenCode + Oh-My-OpenAgent + MiniMax M2.5

基於 **lossless-claw v0.9.1** 加上 **lossless-claw-enhanced** 修復點移植的超輕量級無損上下文管理系統。

### 移植來源

1. **lossless-claw v0.9.1** - 原始專案 (Martian-Engineering)
2. **lossless-claw-enhanced** - CJK Token 修復 + 上游 Bug 修復 (win4r fork)

## 什麼是 LCM？

當對話超過模型的上下文窗口時，傳統方法會截斷舊訊息。LCM 採用 DAG 結構的摘要系統，保留每條訊息，同時將活躍上下文保持在模型的 token 限制內。

```
工作流程:
1. 持久化  ──→  所有訊息存入 SQLite
2. 摘要    ──→  舊訊息壓縮成 Leaf Summary  
3. 凝聚    ──→  多個 Leaf 凝聚成更高層節點
4. 組裝    ──→  摘要 + 新鮮尾部 = 完整上下文
```

## ⚠️ 重要：運行前必須配置

> **在使用 clawlcm 之前，必須完成以下兩個步驟：**
> 1. **配置 LLM** - 在 `config.json` 中設置 LLM 服務器（模型、API Key、Base URL）
> 2. **執行 bootstrap** - 首次運行前執行 bootstrap 命令創建空資料庫

### 步驟 1：配置文件 `config.json`

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
    "largeFilesDir": "",
    "maxDepth": 8
  },
  "enabled": true,
  "verbose": false
}
```

> ⚠️ **注意**: `baseURL` 不要包含 `/v1` 结尾，代码会自动添加 `/v1/chat/completions`

### 步驟 2：執行 bootstrap 創建資料庫

```bash
# 執行 bootstrap 創建空資料庫（自動創建 data/clawlcm.db）
./bin/clawlcm bootstrap --session-key "user:default:001"
```

> **⚡ 注意**：首次運行時，資料庫不存在會自動創建在 `data/` 目錄下。

### 或透過命令列參數覆蓋

```bash
./bin/clawlcm --llm-model <YOUR_MODEL> --llm-api-key <YOUR_KEY> --llm-base-url <URL> ...
```

> **⚠️ 请勿在配置文件中硬编码敏感信息，建议使用环境变量或命令行参数**

## 目錄結構

```
clawlcm/
├── bin/clawlcm           # 主程式 (符號連結到 /usr/local/bin/)
├── data/                 # 數據目錄 (可執行檔的上一級目錄)
│   ├── config.json       # 配置文件
│   ├── clawlcm.db       # SQLite 數據庫
│   └── large_files/      # 大文件外置存儲
└── SKILL.md             # Skill 定義
```

### 數據目錄

- **配置檔**: `data/config.json` (自動創建預設)
- **數據庫**: `data/clawlcm.db` (自動創建)
- **大文件**: `data/large_files/` (外置存儲)
- 首次執行 `bootstrap` 或其他命令時，若資料庫不存在會自動創建
- 建議將 `data/` 目錄映射到外部 volume，確保數據持久化

> **⚡ 路徑說明**：所有路徑都基於**可執行檔所在目錄**，無論從哪個目錄運行。
> 支援符號連結，符號連結時使用連結所在目錄：
> ```bash
> # 符號連結後，在任意目錄運行都會將數據放在正確位置
> ln -s /path/to/clawlcm/bin/clawlcm /usr/local/bin/clawlcm
> clawlcm bootstrap  # 數據會寫入 /path/to/clawlcm/data/
> ```

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
               │
               ▼
     [Fresh Tail: 最後 N 條訊息]        ← 受保護的新鮮尾部
```

- **Leaf Summary**: 原始訊息的壓縮摘要
- **Condensed Summary**: 多個 Leaf 摘要的進一步凝聚
- **Fresh Tail**: 保護最近 N 條訊息不被壓縮

## 使用場景

使用此技能當：
- 需要 LLM 驅動的摘要功能 (Leaf + Condensed)
- 需要 BM25 相關性檢索
- 多層壓縮 (DAG 深度追蹤)
- 中文/日文/韓文文本支援
- 需要增強功能：
  - CJK Token 精確估算 (1.5x CJK, 2x Emoji)
  - Auth Error 錯誤過濾
  - Session Rotation 檢測
  - 空訊息跳過

## 快速開始

```bash
# 檢查版本
./bin/clawlcm --version

# 測試 LLM (請替換為您的配置)
./bin/clawlcm -v
```

## 操作說明

### bootstrap ⭐ 新增內容支持
初始化對話並載入現有訊息。

```bash
./bin/clawlcm bootstrap \
  --session-key user:chat:123 \
  --session-id uuid-123 \
  --token-budget 128000 \
  --messages '[{role:user,content:你好},{role:assistant,content:嗨}]'
```

### ingest
新增訊息到對話中。

```bash
./bin/clawlcm ingest \
  --session-key user:chat:123 \
  --session-id uuid-123 \
  --role user \
  --content "解釋 Go goroutines"
```

### assemble
組裝上下文 (摘要 + 新鮮尾部)。

```bash
./bin/clawlcm assemble \
  --session-key user:chat:123 \
  --token-budget 128000
```

### compact
觸發 LLM 摘要 (建立 Leaf 摘要)。

```bash
./bin/clawlcm compact \
  --session-key user:chat:123 \
  --force
```

### grep ⭐ 新增
BM25 檢索訊息。

```bash
# 搜尋單一会話
./bin/clawlcm grep \
  --session-key user:chat:123 \
  --pattern "關鍵詞"

# 搜尋所有會話
./bin/clawlcm grep \
  --all \
  --pattern "關鍵詞" \
  --limit 20
```

### describe ⭐ 新增
描述摘要詳情。

```bash
./bin/clawlcm describe \
  --session-key user:chat:123 \
  --id 摘要ID
```

### expand ⭐ 新增
展開摘要內容。

```bash
./bin/clawlcm expand \
  --session-key user:chat:123 \
  --summary-ids "id1,id2" \
  --query "查詢內容" \
  --max-depth 3
```

### maintain ⭐ 新增多操作
執行維護任務。

```bash
# 垃圾回收
./bin/clawlcm maintain --session-key user:chat:123 --op gc

# 數據庫優化
./bin/clawlcm maintain --op vacuum

# 創建備份
./bin/clawlcm maintain --op backup

# 健康檢查
./bin/clawlcm maintain --op doctor

# 清理大文件
./bin/clawlcm maintain --op clean

# Session Rotation
./bin/clawlcm maintain --op rotate
```

### tui ⚠️ 空實現 (不推薦)
互動式 TUI 模式 (空殼實現，暫不可用)。

> **注意**: TUI 功能已被標記為「不建議實作」，CLI 已完整支援所有操作。

```bash
# 此命令暫時不可用
./bin/clawlcm tui
```

## 配置說明

### 配置文件參數

#### database (資料庫)

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `database.path` | `""` | 資料庫路徑 (空值時使用 `data/clawlcm.db`) |

#### llm (LLM 配置) ⭐ 新增

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `llm.model` | - | LLM 模型 **必填** |
| `llm.provider` | `openai` | LLM provider |
| `llm.apiKey` | - | API 金鑰 |
| `llm.baseURL` | - | API 端點 (不含 /v1) **必填** |
| `llm.timeoutMs` | 120000 | 請求超時 (毫秒) |

#### context (上下文壓縮) ⭐ 新增多項

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `context.threshold` | 0.75 | 壓縮閾值 (0.0-1.0) |
| `context.freshTailCount` | 8 | 保護的最近訊息數 |
| `context.useCJKTokenizer` | true | 啟用中文分詞 |
| `context.maxDepth` | 8 | 最大 DAG 深度 (已棄用，改用 incrementalMaxDepth) |
| `context.condensedMinFanout` | 4 | **⭐ 新增** Leaf 凝聚最小子節點數 |
| `context.incrementalMaxDepth` | 1 | **⭐ 新增** 遞進壓縮最大深度 |
| `context.proactiveThresholdCompactionMode` | `deferred` | **⭐ 新增** 主動壓縮模式 (deferred/immediate) |
| `context.maintenanceDebtEnabled` | true | **⭐ 新增** 維護Debt啟用 |
| `context.maintenanceDebtThreshold` | 50000 | **⭐ 新增** 維護Debt閾值 |
| `context.largeFilesDir` | `""` | 大文件目錄 (空值時使用 `data/large_files/`) |
| `context.cacheAwareCompaction` | false | **⭐ 新增** 緩存感知壓縮 |
| `context.leafChunkTokens` | 20000 | Leaf 壓縮區塊大小 |

#### session (對話會話) ⭐ 新增

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `session.ignoreSessionPatterns` | [] | **⭐ 新增** 忽略會話模式 |
| `session.statelessSessionPatterns` | [] | **⭐ 新增** 無狀態會話模式 |
| `session.skipStatelessSessions` | false | **⭐ 新增** 跳過無狀態會話 |

#### 頂層配置

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `enabled` | true | 啟用 LCM |
| `verbose` | false | 詳細輸出 |

### 命令列參數

#### 通用參數

| 參數 | 說明 |
|------|------|
| `--config` | 配置文件路徑 |
| `--db` | 資料庫路徑 |
| `--llm-model` | LLM 模型 |
| `--llm-provider` | LLM provider |
| `--llm-api-key` | API 金鑰 |
| `--llm-base-url` | API 端點 |
| `--llm-timeout` | 請求超時 (毫秒) |
| `-v` | 詳細輸出 |
| `--version` | 顯示版本 |

#### session 參數

| 參數 | 說明 |
|------|------|
| `--session-key` | 會話鍵 (必需) |
| `--session-id` | 會話 ID |
| `--token-budget` | Token 預算 (預設 128000) |
| `--force` | 強制執行 |

### 推薦配置

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

> **⚠️ 安全提示**: 请将敏感信息 (apiKey, baseURL) 替换为实际值，不要提交到版本控制

- **freshTailCount=64**: 保護最後 64 條訊息，提供更好的對話連貫性
- **leafChunkTokens=20000**: 控制 Leaf 壓縮區塊大小
- **threshold=0.75**: 當上下文達到 75% 時觸發壓縮
- **maxDepth=8**: 最大 DAG 壓縮深度

## 功能對照

| 功能 | 說明 |
|------|------|
| 訊息持久化 | ✅ SQLite 儲存 |
| 上下文組裝 | ✅ 摘要 + 新鮮尾部 |
| Leaf 摘要 | ✅ (LLM) |
| Condensed 摘要 | ✅ (DAG) |
| BM25 檢索 | ✅ |
| 中文分詞 | ✅ |
| DAG 深度追蹤 | ✅ |
| CJK Token 估算 | ✅ 1.5x CJK, 2x Emoji |
| Auth Error 過濾 | ✅ |
| Session Rotation | ✅ |
| 空訊息跳過 | ✅ |

## 輸出格式

JSON 格式輸出供程式使用。

## 技術規格

| 指標 | 數值 |
|------|------|
| 目標 RAM | <50 MB (不含 LLM) |
| 依賴 | SQLite + GORM |
| 二進制大小 | ~15 MB |
| 語言 | Go 1.22+ |

## 來源

基於 [lossless-claw v0.9.1](https://github.com/Martian-Engineering/lossless-claw/releases/tag/v0.9.1) + [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced) 增強。
這是 OpenClaw 的 LCM (Lossless Context Management) 插件。

## 參考

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - 原始項目
- [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced) - 增強版本
- [LCM Paper](https://papers.voltropy.com/LCM) - 技術論文