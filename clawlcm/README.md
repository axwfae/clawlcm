# clawlcm - 無損上下文管理

clawlcm 是基於 lossless-claw v0.9.1 + lossless-claw-enhanced 移植的無損上下文管理系統。

> **版本**: v0.8.4 | **更新日**: 2026-04-16 | **移植工具**: OpenCode + Oh-My-OpenAgent + MiniMax M2.5

---

## 專案資訊

| 項目 | 值 |
|------|------|
| 專案名稱 | clawlcm |
| 基於 | lossless-claw v0.9.1 + lossless-claw-enhanced |
| 原始專案 | lossless-claw |
| 當前版本 | **v0.8.4** |

---

## 移植說明

本專案使用以下工具移植：

- **OpenCode** - AI 開發環境
- **Oh-My-OpenAgent** - AI Agent 編排框架
- **MiniMax M2.5** - LLM 模型

---

## 概述

LCM (Lossless Context Management) 是一個專為大型語言模型 (LLM) 應用設計的無損上下文管理系統。當對話超過模型的上下文窗口時，傳統方法會截斷舊訊息。LCM 採用 DAG 結構的摘要系統，保留每條訊息，同時將活躍上下文保持在模型的 token 限制內。

---

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
| CJK Token 估算 | 1.5x CJK, 2x Emoji | ✅ |
| Auth Error 過濾 | 防止 false-positive 錯誤 | ✅ |
| Session Rotation | 對話分割檢測 | ✅ |
| 空訊息跳過 | 跳過 empty assistant | ✅ |
| CLI --flags 修復 | 支援 command --flags 格式 | ✅ |

---

## 技術規格

| 指標 | 數值 |
|------|------|
| 目標 RAM | <50 MB (不含 LLM) |
| 依賴 | SQLite + GORM |
| 編譯大小 | ~15 MB |
| 語言 | Go 1.22+ |

---

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

---

## 配置

配置文件 `data/config.json`：

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

> ⚠️ **注意**: `baseURL` 不要包含 `/v1` 结尾
> **安全提示**: 敏感信息 (apiKey, baseURL) 請使用環境變量或命令行參數，不要提交到版本控制

---

## 專案結構

```
clawlcm/
├── bin/clawlcm       # 主程式 (符號連結到 /usr/local/bin/)
├── data/             # 數據目錄 (相對於可執行檔)
│   ├── config.json   # 配置文件
│   ├── clawlcm.db   # SQLite 數據庫
│   └── large_files/  # 大文件外置存儲
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
├── Makefile          # 構建腳本
└── README.md         # 本文件
```

---

## 自包含路徑

所有路徑都基於**可執行檔所在目錄的上一級**，無論從哪個目錄運行：
- 數據庫：`data/clawlcm.db`
- 配置檔：`data/config.json`
- 大文件：`data/large_files/`

```bash
# 符號連結後，在任意目錄運行都會將數據放在正確位置
ln -s /path/to/clawlcm/bin/clawlcm /usr/local/bin/clawlcm
clawlcm --help  # 數據會寫入 /path/to/clawlcm/data/
```

---

## CLI 命令

| 命令 | 說明 |
|------|------|
| `bootstrap` | 初始化對話 |
| `ingest` | 添加訊息 |
| `assemble` | 組裝上下文 |
| `compact` | 觸發壓縮 |
| `grep` | BM25 檢索 |
| `describe` | 描述摘要 |
| `expand` | 展開摘要 |
| `maintain --op` | 維護操作 (gc/vacuum/backup/doctor/clean/rotate) |
| `tui` | ⚠️ 空實現 (不推薦) |

---

## 不建議實作

1. **TUI (交互式終端界面)**

   - **原因**: 
   - CLI 已有完整命令支援
   - 目標是作為 OpenClaw 插件運行，非獨立 TUI 用途

2. **FTS5 (SQLite 全文搜索)**

   - **原因**:
   - 已有 BM25 檢索功能
   - FTS5 會增加複雜度
   - 建議使用外部檢索方案

---

## 版本歷史

| 版本 | 日期 | 變更 |
|------|------|------|
| v0.8.4 | 2026-04-16 | 統一文檔一致性，移除 ./ 前綴 |
| v0.8.3 | 2026-04-16 | 自包含路徑 + 修復 largeFilesDir + 符號連結支援 |
| v0.8.2 | 2026-04-16 | CLI 修復 (describe/expand/maintain 參數) |
| v0.8.1 | 2026-04-15 | 增強版發布 (CJK Token, Auth Error, Session Rotation, 空訊息, CLI --flags) |
| v0.8.0 | 2026-04-14 | 移植 lossless-claw 功能 (Deferred, Maintenance Debt, LargeFiles, Rotate, etc.) |

---

## 參考

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - 原始項目
- [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced) - 增強版本
- [LCM Paper](https://papers.voltropy.com/LCM) - 技術論文
- [OpenCode](https://opencode.com) - AI 開發環境
- [Oh-My-OpenAgent](https://github.com/oh-my-openagent) - AI Agent 編排框架

---

## 授權

MIT License