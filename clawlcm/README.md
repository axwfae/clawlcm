# clawlcm - 無損上下文管理

clawlcm 是基於 picolcm 的無損上下文管理系統，旨在逐步對比並追上 [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) 的功能。

## 專案資訊

| 項目 | 值 |
|------|------|
| 專案名稱 | clawlcm |
| 基於 | picolcm |
| 原始專案 | [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) |
| 當前版本 | 0.3.0 |

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
| Condensed 摘要 | 多層壓縮 | ❌ 開發中 |
| Agent Tools | lcm_grep/expand 等 | ❌ 規劃中 |

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

# 初始化對話
./clawlcm bootstrap \
  --session-key "user:chat:1" \
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

# 觸發 LLM 摘要
./clawlcm compact \
  --session-key "user:chat:1" \
  --force
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
    "baseURL": "http://localhost:18869",
    "timeoutMs": 120000
  },
  "context": {
    "threshold": 0.75,
    "freshTailCount": 64,
    "useCJKTokenizer": true
  }
}
```

> ⚠️ **注意**: `baseURL` 不要包含 `/v1` 结尾

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

### 主要差距

1. **無 Condensed 摘要**: 無法將多個 Leaf 凝聚成高層節點
2. **無 Agent Tools**: 代理無法深入歷史訊息
3. **無會話過濾**: 無法根據模式排除會話
4. **無維護工具**: 無法進行健康檢查和修復

## 參考

- [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) - 原始項目
- [LCM Paper](https://papers.voltropy.com/LCM) - 技術論文

## 授權

MIT License
