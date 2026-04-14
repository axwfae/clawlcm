# clawlcm - 功能差異分析報告

## 專案資訊

| 項目 | 值 |
|------|------|
| 專案名稱 | clawlcm |
| 基於 | picolcm (lossless-claw 移植) |
| 當前版本 | 0.4.0 |
| 原始專案 | [lossless-claw](https://github.com/Martian-Engineering/lossless-claw) |

---

## 功能對比總覽

### 已實現功能 ✅

| 功能 | clawlcm | lossless-claw | 說明 |
|------|---------|---------------|------|
| 訊息持久化 | ✅ | ✅ | SQLite 儲存對話歷史 |
| 上下文組裝 | ✅ | ✅ | 摘要 + 新鮮尾部 |
| Leaf 摘要 | ✅ | ✅ | LLM 驅動的摘要生成 |
| DAG 壓縮追蹤 | ⚠️ 基礎 | ✅ 完整 | DAG 結構支援 |
| BM25 檢索 | ✅ | ✅ | 相關性搜索 |
| 中文分詞 | ✅ | ✅ | CJK 支援 |
| JSON 配置 | ✅ | ✅ | 配置文件支援 |

### 缺失功能 ❌

| 功能 | 說明 | 優先級 |
|------|------|--------|
| **Condensed 摘要** | 多個 Leaf 凝聚成更高層節點 | 高 |
| **增量壓縮** | incrementalMaxDepth 控制 | 高 |
| **Agent Tools** | lcm_grep, lcm_describe, lcm_expand_query | 高 |
| **Session 過濾** | ignoreSessionPatterns | 中 |
| **Stateless Sessions** | 唯讀會話支援 | 中 |
| **Transcript GC** | 轉錄本垃圾回收 | 低 |
| **延遲壓縮** | proactiveThresholdCompactionMode | 低 |
| **大文件攔截** | 大文件獨立存儲 | 低 |
| **快取感知壓縮** | cacheAwareCompaction | 低 |
| **TUI 工具** | 交互式終端界面 | 低 |

---

## 詳細功能差異

### 1. 壓縮引擎 (Compaction Engine)

#### lossless-claw (完整實現)
```
- Leaf 壓縮 (單層摘要) ✅
- Condensed 壓縮 (多層凝聚) ✅
- 增量壓縮 (incrementalMaxDepth: 0=leaf only, 1=+1 condense, -1=unlimited) ✅
- 延遲壓縮模式 (deferred/inline) ✅
- 快取感知壓縮 (cache-aware) ✅
- 三級昇級 (Normal → Aggressive → Fallback) ✅
- Budget-targeted compaction ✅
```

#### clawlcm (當前)
```
- Leaf 壓縮 (單層摘要) ✅
- Condensed 壓縮 ❌
- 增量壓縮 ❌
- 延遲壓縮 ❌
- 快取感知 ❌
```

**差距**: 無法將多個 Leaf 摘要進一步凝聚成更高層節點，導致長期對話時摘要數量過多。

---

### 2. Agent Tools (代理工具)

#### lossless-claw 提供的工具
| 工具 | 功能 | 狀態 |
|------|------|------|
| `lcm_grep` | 全文檢索 (regex/full-text) | ✅ |
| `lcm_describe` | 描述對話摘要結構 | ✅ |
| `lcm_expand` | 展開摘要查看原始訊息 (低級) | ✅ |
| `lcm_expand_query` | 代理查詢，深入摘要獲取細節 | ✅ |

**工具參數詳解**:
- **lcm_grep**: pattern, mode (regex/full_text), scope, conversationId, allConversations, since/before, limit, sort
- **lcm_describe**: id (sum_xxx/file_xxx), conversationId, allConversations, tokenCap
- **lcm_expand**: summaryIds, query, maxDepth, tokenCap, includeMessages, conversationId
- **lcm_expand_query**: summaryIds, query, prompt, maxTokens, tokenCap, conversationId

#### clawlcm (當前)
```
- BM25 檢索 ✅ (基礎實現)
- 無專用工具 ❌
```

**差距**: 缺乏讓代理深入歷史訊息的工具，無法從壓縮的摘要中恢復原始詳細資訊。

---

### 3. 會話管理 (Session Management)

#### lossless-claw
```
- ignoreSessionPatterns: 排除特定會話 ✅
- statelessSessionPatterns: 唯讀會話 ✅
- skipStatelessSessions: 跳過唯寫 ✅
- newSessionRetainDepth: /new 命令保留深度 ✅
- Session reset semantics: /new vs /reset vs /lcm rotate ✅
- /lcm rotate: 轉錄本重寫 ✅
```

#### clawlcm (當前)
```
- 基礎會話創建/讀取 ✅
- 無模式匹配過濾 ❌
```

---

### 4. 維護任務 (Maintenance)

#### lossless-claw
```
- /lcm: 顯示版本、狀態、DB 路徑 ✅
- /lcm backup: 建立時間戳備份 ✅
- /lcm rotate: 轉錄本重寫 ✅
- /lcm doctor: 掃描損壞摘要 ✅
- /lcm doctor clean: 清理孤立數據 ✅
- /lcm status: 顯示插件狀態 ✅
- transcriptGcEnabled: 轉錄本 GC ✅
```

#### clawlcm (當前)
```
- maintain 命令 (空實現) ❌
- 無健康檢查 ❌
- 無備份功能 ❌
```

---

### 5. TUI 工具

#### lossless-claw
```
- lcm-tui repair: 掃描並修復損壞摘要
- lcm-tui rewrite: 重新摘要
- lcm-tui dissolve: 撤銷凝聚
- lcm-tui transplant: 跨會話複製 DAG
- lcm-tui backfill: 導入歷史 JSONL
- lcm-tui prompts: 管理深度提示模板
```

#### clawlcm
```
- 無 TUI ❌
```

---

### 6. 配置參數

#### lossless-claw 完整配置 (50+ 參數)

| 類別 | 參數 |
|------|------|
| **核心** | enabled, databasePath, freshTailCount, contextThreshold |
| **壓縮** | leafChunkTokens, leafTargetTokens, condensedTargetTokens, condensedMinFanout, incrementalMaxDepth |
| **快取** | cacheAwareCompaction.enabled, cacheTTLSeconds, cacheThroughputThreshold |
| **模型** | summaryModel, summaryProvider, expansionModel, expansionProvider |
| **會話** | ignoreSessionPatterns, statelessSessionPatterns, skipStatelessSessions |
| **維護** | transcriptGcEnabled, proactiveThresholdCompactionMode |
| **安全** | subagent.allowModelOverride, subagent.allowedModels |
| **大文件** | largeFileThresholdTokens, largeFileSummaryModel |
| **超時** | delegationTimeoutMs, summaryTimeoutMs |

#### clawlcm 當前配置 (15 參數)

| 參數 | 預設值 |
|------|--------|
| DatabasePath | ./data/clawlcm.db |
| Enabled | true |
| ContextThreshold | 0.75 |
| FreshTailCount | 8 |
| LeafChunkTokens | 20000 |
| LeafTargetTokens | 2400 |
| CondensedTargetTokens | 2000 |
| UseCJKTokenizer | true |
| SummaryModel | - |
| SummaryProvider | openai |
| SummaryBaseURL | - |
| SummaryTimeoutMs | 60000 |

**差距**: 缺少 35+ 個配置參數。

---

### 7. 命令列界面

#### lossless-claw 命令 (Slash Commands)

| 命令 | 功能 |
|------|------|
| `/lcm` | 顯示版本、狀態、DB 路徑 |
| `/lcm backup` | 建立時間戳備份 |
| `/lcm rotate` | 重寫轉錄本 |
| `/lcm doctor` | 掃描損壞摘要 |
| `/lcm doctor clean` | 清理孤立數據 |
| `/lcm status` | 顯示插件狀態 |
| `/lossless` | /lcm 別名 |

#### clawlcm 命令

| 命令 | 功能 |
|------|------|
| `bootstrap` | 初始化對話 ✅ |
| `ingest` | 添加訊息 ✅ |
| `assemble` | 組裝上下文 ✅ |
| `compact` | 觸發壓縮 ✅ |
| `maintain` | 維護 (空) ⚠️ |
| `--version` | 顯示版本 ✅ |

---

### 8. 專案結構

#### lossless-claw 結構
```
lossless-claw/
├── src/
│   ├── engine.ts           # ContextEngine 實現
│   ├── assembler.ts       # 上下文組裝
│   ├── compaction.ts       # 壓縮引擎
│   ├── summarize.ts       # 摘要生成
│   ├── retrieval.ts       # 檢索引擎
│   ├── expansion.ts       # DAG 展開
│   ├── integrity.ts       # 完整性檢查
│   ├── large-files.ts     # 大文件處理
│   └── ...
├── tools/
│   ├── lcm-grep-tool.ts
│   ├── lcm-describe-tool.ts
│   ├── lcm-expand-tool.ts
│   └── lcm-expand-query-tool.ts
├── tui/                    # Go 實現的 TUI
│   ├── main.go
│   ├── repair.go
│   ├── rewrite.go
│   └── ...
├── docs/
│   ├── agent-tools.md
│   ├── architecture.md
│   ├── configuration.md
│   ├── tui.md
│   └── fts5.md
└── ...
```

#### clawlcm 結構
```
clawlcm/
├── cmd/clawlcm/      # CLI
├── engine.go         # 引擎核心
├── store/            # 數據存儲
├── retrieval/        # BM25
├── tokenizer/        # 分詞
├── llm/              # LLM 客戶端
├── db/               # 資料庫
├── types/            # 類型定義
├── logger/          # 日誌
└── docker/          # Docker 配置
```

---

### 9. 特殊功能

#### lossless-claw 特殊功能
| 功能 | 說明 |
|------|------|
| **FTS5** | 可選的 SQLite 全文搜索支援 |
| **大文件處理** | 攔截大文件，獨立存儲探索摘要 |
| **認證層級** | 三層認證 cascade (auth profiles, env, config) |
| **Session 協調** | 啟動時的 bootstrap reconciliation |
| **操作序列化** | 每個會話的 promise queue |
| **深度提示模板** | 4 個深度特定模板 (leaf, condensed-d1, d2, d3) |

#### clawlcm
```
- 無 FTS5 ❌
- 無大文件處理 ❌
- 基礎認證 ❌
```

---

## 功能開發路線圖

### Phase 1: 核心壓縮 (高優先級)

| 功能 | 預估工作量 | 說明 |
|------|-----------|------|
| Condensed 摘要 | 2-3 天 | 多 Leaf 凝聚邏輯 |
| 增量壓縮控制 | 1 天 | incrementalMaxDepth |

### Phase 2: Agent Tools (高優先級)

| 功能 | 預估工作量 | 說明 |
|------|-----------|------|
| lcm_grep 工具 | 1 天 | 基於現有 BM25 |
| lcm_describe 工具 | 1 天 | 摘要結構描述 |
| lcm_expand 工具 | 2 天 | 展開摘要獲取細節 |

### Phase 3: 會話管理 (中優先級)

| 功能 | 預估工作量 | 說明 |
|------|-----------|------|
| ignoreSessionPatterns | 1 天 | 模式匹配 |
| statelessSessions | 1 天 | 唯讀會話 |

### Phase 4: 維護工具 (中優先級)

| 功能 | 預估工作量 | 說明 |
|------|-----------|------|
| /lcm doctor | 1 天 | 健康檢查 |
| /lcm backup | 0.5 天 | 備份功能 |
| maintain 實現 | 1 天 | GC 和優化 |

### Phase 5: 高級功能 (低優先級)

| 功能 | 預估工作量 | 說明 |
|------|-----------|------|
| TUI 工具 | 3-5 天 | 交互式界面 |
| 大文件攔截 | 2 天 | 文件處理 |
| 快取感知壓縮 | 1-2 天 | 優化 |
| Transcript GC | 1 天 | 清理 |

---

## 總結

### 完整度評估

| 類別 | 完整度 |
|------|--------|
| 基礎功能 | 60% |
| 壓縮引擎 | 40% |
| 代理工具 | 10% |
| 會話管理 | 30% |
| 維護工具 | 10% |
| 配置參數 | 30% |
| TUI/特殊功能 | 0% |

### 主要差距

1. **無 Condensed 摘要**: 無法將多個 Leaf 凝聚成高層節點
2. **無 Agent Tools**: 代理無法深入歷史訊息
3. **無會話過濾**: 無法根據模式排除會話
4. **無維護工具**: 無法進行健康檢查和修復
5. **配置參數不完整**: 缺少 35+ 個進階參數
6. **無 TUI**: 缺乏交互式界面

### 參考文檔

- [lossless-claw README](https://github.com/Martian-Engineering/lossless-claw/blob/d51abb9cabeea0d5875c4f53dced7d0167e0ba06/README.md)
- [lossless-claw Configuration](https://github.com/Martian-Engineering/lossless-claw/blob/d51abb9cabeea0d5875c4f53dced7d0167e0ba06/docs/configuration.md)
- [lossless-claw Architecture](https://github.com/Martian-Engineering/lossless-claw/blob/d51abb9cabeea0d5875c4f53dced7d0167e0ba06/docs/architecture.md)
- [lossless-claw Agent Tools](https://github.com/Martian-Engineering/lossless-claw/blob/d51abb9cabeea0d5875c4f53dced7d0167e0ba06/docs/agent-tools.md)
- [lossless-claw TUI](https://github.com/Martian-Engineering/lossless-claw/blob/d51abb9cabeea0d5875c4f53dced7d0167e0ba06/docs/tui.md)
