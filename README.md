# clawlcm - 無損上下文管理

clawlcm 是基於 lossless-claw 的無損上下文管理系統的 Go 實現。

## 專案資訊

| 項目 | 值 |
|------|------|
| 專案名稱 | clawlcm |
| 基於 | lossless-claw v0.9.1 + lossless-claw-enhanced |
| 當前版本 | v0.8.9 |
| 語言 | Go 1.22+ |

## 概述

LCM (Lossless Context Management) 是一個專為大型語言模型 (LLM) 應用設計的無損上下文管理系統。當對話超過模型的上下文窗口時，傳統方法會截斷舊訊息。LCM 採用 DAG 結構的摘要系統，保留每條訊息，同時將活躍上下文保持在模型的 token 限制內。

```
工作流程:
1. 持久化  ──→  所有訊息存入 SQLite
2. 摘要    ──→  舊訊息壓縮成 Leaf Summary  
3. 凝聚    ──→  多個 Leaf 凝聚成更高層節點
4. 組裝    ──→  摘要 + 新鮮尾部 = 完整上下文
```

## 核心功能

| 功能 | 說明 | 狀態 |
|------|------|------|
| 訊息持久化 | SQLite 儲存對話歷史 | ✅ |
| 上下文組裝 | 智能組裝摘要 + 新鮮尾部 | ✅ |
| Leaf 摘要 | LLM 驅動的摘要生成 | ✅ |
| Condensed 摘要 | DAG 多層凝聚 | ✅ |
| BM25 檢索 | 相關性搜索 | ✅ |
| 中文分詞 | 支援中文/日文/韓文 | ✅ |
| Grep 命令 | BM25 全文檢索 | ✅ |
| Describe 命令 | 描述摘要詳情 | ✅ |
| Expand 命令 | 展開摘要內容 | ✅ |
| Maintain | 維護工具 (gc/vacuum/backup) | ✅ |

## 技術規格

| 指標 | 數值 |
|------|------|
| 目標 RAM | <50 MB (不含 LLM) |
| 依賴 | SQLite + GORM |
| 編譯大小 | ~15 MB |
| 語言 | Go 1.22+ |

## 命令詳解

### bootstrap

初始化對話並可選載入初始訊息。

```bash
./clawlcm bootstrap \
  --session-key "user:chat:123" \
  --session-id "uuid-123" \
  --token-budget 128000 \
  --messages '[{"role":"user","content":"你好"}]'
```

**參數說明:**
| 參數 | 說明 | 預設值 |
|------|------|--------|
| `--session-key` | 會話鍵 (格式: user:chat:123) | - |
| `--session-id` | 會話 ID (UUID) | 自動生成 |
| `--token-budget` | Token 預算 | 128000 |
| `--messages` | JSON 訊息陣列 | [] |

---

### ingest

新增訊息到對話中。

```bash
./clawlcm ingest \
  --session-key "user:chat:123" \
  --role user \
  --content "什麼是 Go 語言？"
```

**參數說明:**
| 參數 | 說明 | 預設值 |
|------|------|--------|
| `--session-key` | 會話鍵 | - |
| `--role` | 角色 | user |
| `--content` | 訊息內容 | - |

---

### assemble

組裝上下文 (摘要 + 新鮮尾部)。

```bash
./clawlcm assemble \
  --session-key "user:chat:123" \
  --token-budget 128000
```

**參數說明:**
| 參數 | 說明 | 預設值 |
|------|------|--------|
| `--session-key` | 會話鍵 | - |
| `--token-budget` | Token 預算 | 128000 |

---

### compact

觸發 LLM 摘要 (建立 Leaf 摘要)。

```bash
./clawlcm compact \
  --session-key "user:chat:123" \
  --force
```

**參數說明:**
| 參數 | 說明 | 預設值 |
|------|------|--------|
| `--session-key` | 會話鍵 | - |
| `--force` | 強制壓縮 | false |

---

### grep ⭐ 重要

BM25 全文檢索。

```bash
# 搜尋單一会話
./clawlcm grep \
  --session-key "user:chat:123" \
  --pattern "Go 語言"

# 搜尋所有會話
./clawlcm grep \
  --all \
  --pattern "error" \
  --limit 20

# 使用正規表達式
./clawlcm grep \
  --session-key "user:chat:123" \
  --pattern "error.*fail" \
  --mode regex
```

**參數說明:**
| 參數 | 說明 | 預設值 |
|------|------|--------|
| `--session-key` | 會話鍵 | - |
| `--all` | 搜尋所有會話 | false |
| `--pattern` | **搜尋模式** | - |
| `--mode` | 搜尋模式 (full_text/regex) | full_text |
| `--scope` | 搜尋範圍 (all/messages/summaries) | all |
| `--limit` | 結果數量 | 20 |
| `--sort` | 排序 (desc/asc) | desc |

> **注意**: `grep` 使用 `--pattern` 不是 `--query`！

---

### describe

描述摘要詳情。

```bash
./clawlcm describe \
  --session-key "user:chat:123" \
  --id 1
```

**參數說明:**
| 參數 | 說明 | 預設值 |
|------|------|--------|
| `--session-key` | 會話鍵 | - |
| `--id` | 摘要 ID | - |
| `--all` | 所有會話 | false |

---

### expand

展開摘要內容。

```bash
./clawlcm expand \
  --session-key "user:chat:123" \
  --summary-ids "1,2,3" \
  --query "Go 語言的特性" \
  --max-depth 3 \
  --include-messages
```

**參數說明:**
| 參數 | 說明 | 預設值 |
|------|------|--------|
| `--session-key` | 會話鍵 | - |
| `--summary-ids` | 摘要 ID (逗號分隔) | - |
| `--query` | 查詢內容 | - |
| `--max-depth` | 最大深度 | 3 |
| `--include-messages` | 包含訊息 | false |

> **注意**: `expand` 使用 `--query` 不是 `--pattern`！

---

### maintain

執行維護任務。

```bash
# 垃圾回收
./clawlcm maintain --session-key "user:chat:123" --maint-op gc

# 資料庫優化
./clawlcm maintain --op vacuum

# 創建備份
./clawlcm maintain --op backup

# 健康檢查
./clawlcm maintain --op doctor

# 清理大文件
./clawlcm maintain --op clean

# Session Rotation
./clawlcm maintain --op rotate
```

**參數說明:**
| 參數 | 說明 | 預設值 |
|------|------|--------|
| `--session-key` | 會話鍵 | - |
| `--maint-op` | 維護操作 | gc |
| `--all` | 所有會話 | false |

**維護操作:**
| 操作 | 說明 |
|------|------|
| `gc` | 垃圾回收 (清理孤立摘要) |
| `vacuum` | 資料庫優化 |
| `backup` | 創建備份 |
| `doctor` | 健康檢查 |
| `clean` | 清理大文件 |
| `rotate` | Session Rotation |

---

## 配置

配置文件 `./data/config.json`：

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

> ⚠️ **注意**: `baseURL` 不要包含 `/v1` 结尾

### 配置參數

#### llm

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `llm.model` | - | LLM 模型 **必填** |
| `llm.provider` | `openai` | LLM provider |
| `llm.apiKey` | - | API 金鑰 |
| `llm.baseURL` | - | API 端點 (不含 /v1) **必填** |
| `llm.timeoutMs` | 120000 | 超時 (毫秒) |

#### context

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `context.threshold` | 0.75 | 壓縮閾值 (0.0-1.0) |
| `context.freshTailCount` | 8 | 保護的最近訊息數 |
| `context.useCJKTokenizer` | true | 啟用中文分詞 |
| `context.condensedMinFanout` | 4 | Leaf 凝聚最小子節點數 |
| `context.incrementalMaxDepth` | 1 | 遞進壓縮最大深度 |
| `context.proactiveThresholdCompactionMode` | `deferred` | 主動壓縮模式 (deferred/immediate) |
| `context.maintenanceDebtEnabled` | true | 維護Debt啟用 |
| `context.maintenanceDebtThreshold` | 50000 | 維護Debt閾值 |
| `context.leafChunkTokens` | 20000 | Leaf 壓縮區塊大小 |

#### session

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `session.ignoreSessionPatterns` | [] | 忽略會話模式 |
| `session.statelessSessionPatterns` | [] | 無狀態會話模式 |
| `session.skipStatelessSessions` | false | 跳過無狀態會話 |

### CLI 全域參數

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

## 推薦配置

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

- **freshTailCount=8**: 保護最後 8 條訊息
- **incrementalMaxDepth=1**: 遞進壓縮最大深度 (超過此深度不再進行更高層凝聚)
- **condensedMinFanout=4**: Leaf 凝聚最少需要 4 個子節點
- **leafChunkTokens=20000**: 控制 Leaf 壓縮區塊大小
- **threshold=0.75**: 當上下文達到 75% 時觸發壓縮

## 常見錯誤

| 錯誤訊息 | 原因 | 解決方式 |
|---------|------|----------|
| `session-key is required` | 未指定會話 | 添加 `--session-key user:chat:123` |
| `Error: --id is required` | describe 未指定 ID | 添加 `--id 1` |
| `Error: --summary-ids is required` | expand 未指定摘要 ID | 添加 `--summary-ids 1` |
| `grep` 返回 0 結果 | 使用了 `--query` | 確認使用 `--pattern` (不是 `--query`) |

## 專案結構

```
clawlcm/
├── cmd/              # CLI 入口
├── engine.go         # 引擎核心
├── store/            # 數據存儲
├── retrieval/        # BM25 檢索
├── tokenizer/        # 中文分詞
├── llm/              # LLM 客戶端
├── db/               # 資料庫連接
├── types/            # 類型定義
├── logger/           # 日誌
├── docker/           # Docker 配置
└── Makefile          # 構建腳本
```

## 版本歷史

| 版本 | 日期 | 變更 |
|------|------|------|
| v0.8.9 | 2026-04-17 | 修復 --help 版本號、Keywords 填充、Tokenizer CJK、Grep |
| v0.8.8 | 2026-04-17 | 修復 maintainGC、JSON unmarshal、氣泡排序優化 |
| v0.8.7 | 2026-04-16 | 修復 grep -mode 參數解析 |

## 參考

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - 原始項目
- [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced) - 增強版本
- [LCM Paper](https://papers.voltropy.com/LCM) - 技術論文

## 授權

MIT License