# clawlcm - 無損上下文管理

clawlcm 是基於 picolcm 的無損上下文管理系統，旨在逐步對比並追上 [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) 的功能。

## 專案資訊

| 項目 | 值 |
|------|------|
| 專案名稱 | clawlcm |
| 基於 | picolcm |
| 原始專案 | [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) |
| 當前版本 | **v0.8.1** |

## 概述

LCM (Lossless Context Management) 是一個專為大型語言模型 (LLM) 應用設計的無損上下文管理系統。當對話超過模型的上下文窗口時，傳統方法會截斷舊訊息。LCM 採用 DAG 結構的摘要系統，保留每條訊息，同時將活躍上下文保持在模型的 token 限制內。

## 核心功能

| 功能 | 說明 | 狀態 |
|------|------|------|
| 訊息持久化 | SQLite 儲存對話歷史 | ✅ |
| 上下文組裝 | 智能組裝摘要 + 新鮮尾部 | ✅ |
| Leaf 摘要 | LLM 驅動的摘要生成 | ✅ |
| BM25 檢索 | 相關性搜索 | ✅ |
| 中文分詞 | 支援中文/日文/韓文 | ✅ |
| Deferred 壓縮 | deferred/inline 模式切換 | ✅ |
| Maintenance Debt | 維護債務追蹤 | ✅ |
| LargeFilesDir | 大文件外置存儲 | ✅ |
| rotate 命令 | 對話分割重寫 | ✅ |
| Session 過濾 | ignore/stateless patterns | ✅ |
| 完整維護工具 | gc/vacuum/backup/doctor/clean/rotate | ✅ |
| Agent Tools | lcm_grep/lcm_describe/lcm_expand | ✅ |
| Expand Query | 基於查詢的 DAG 擴展 | ✅ |

## 技術規格

| 指標 | 數值 |
|------|------|
| 目標 RAM | <50 MB (不含 LLM) |
| 依賴 | SQLite + GORM |
| 編譯大小 | ~15 MB |
| 語言 | Go 1.22+ |

## 快速開始

```bash
# 編譯
make build

# 初始化對話 (command --flags)
./clawlcm bootstrap \
  --session-key "user:chat:1" \
  --messages '[{"role":"user","content":"你好"}]'

# 或使用 --flags 在命令之前
./clawlcm --session-key "user:chat:1" bootstrap \
  --messages '[{"role":"user","content":"你好"}]'

# 添加訊息
./clawlcm ingest \
  --session-key "user:chat:1" \
  --role user \
  --content "什麼是 Go 語言？"

# 組裝上下文
./clawlcm assemble \
  --session-key "user:chat:1" \
  --token-budget 128000

# 或使用 --flags 在命令之前
./clawlcm --session-key "user:chat:1" assemble --token-budget 128000

# 觸發 LLM 摘要
./clawlcm compact \
  --session-key "user:chat:1" \
  --force

# 或使用 --flags 在命令之前
./clawlcm --session-key "user:chat:1" compact --force
```

## 配置

配置文件 `./data/config.json`：

```json
{
  "database": {
    "path": "./data/clawlcm.db"
  },
  "llm": {
    "model": "minimax_m2.5",
    "provider": "openai",
    "apiKey": "your-api-key",
    "baseURL": "http://your-llm-server:PORT",
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
    "largeFilesDir": "./data/large_files",
    "cacheAwareCompaction": false
  },
  "session": {
    "ignoreSessionPatterns": [],
    "statelessSessionPatterns": [],
    "skipStatelessSessions": false
  }
}
```

> ⚠️ **注意**: `baseURL` 不要包含 `/v1` 结尾
> **v0.6.0 新增**: `proactiveThresholdCompactionMode`, `maintenanceDebtEnabled`, `largeFilesDir` 等配置项

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
├── FEATURE_DIFF.md   # 功能差異報告
└── Makefile          # 構建腳本
```

## 功能差異

詳細功能對比請參考 [FEATURE_DIFF.md](./FEATURE_DIFF.md)。

### 已移植功能 (v0.8.0)

1. ✅ **Deferred 壓縮**: proactiveThresholdCompactionMode
2. ✅ **維護債務**: maintenanceDebt 追蹤
3. ✅ **大文件外置**: largeFilesDir
4. ✅ **rotate 命令**: 對話分割重寫
5. ✅ **Session 過濾**: ignore/stateless patterns
6. ✅ **Agent Tools**: lcm_grep, lcm_describe, lcm_expand
7. ✅ **Expand Query**: 基於查詢的 DAG 擴展，支持 summary_ids 和 query 兩種模式
8. ✅ **Condensed 摘要**: 多 Leaf 凝聚邏輯 (v0.8.1)
9. ✅ **CJK Token**: 修正 CJK/Emoji 估算公式 (v0.8.1)
10. ✅ **Auth Error 過濾**: 防止 false-positive 錯誤 (v0.8.1)
11. ✅ **空訊息過濾**: 跳過 empty assistant 訊息 (v0.8.1)
12. ✅ **CLI 修復**: 支援 command --flags 格式 (v0.8.1)

### 不建議實作

1. **TUI (交互式終端界面)**

   - **原因**: 
   - CLI 已有完整命令支援 (bootstrap/ingest/assemble/compact/grep/describe/expand/maintain)
   - 目標是作為 OpenClaw 插件運行，非獨立 TUI 用途
   - 可透過 OpenClaw 的 TUI 介面使用

2. **FTS5 (SQLite 全文搜索)**

   - **原因**:
   - 已有 BM25 檢索功能 (retrieval/bm25.go)
   - FTS5 會增加 SQLite 依賴和複雜度
   - BM25 對向量搜索場景已足夠
   - 建議使用外部檢索方案 (如向量資料庫) 進行進階搜索

## 參考

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - 原始項目
- [LCM Paper](https://papers.voltropy.com/LCM) - 技術論文

## 授權

MIT License
