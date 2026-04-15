# Lossless-Claw-Enhanced 對比分析報告

## 概述

本文檔對比 `lossless-claw-enhanced` (https://github.com/win4r/lossless-claw-enhanced) 的四個核心修復點與目前 `clawlcm` (v0.8.0) 的實作狀態。

> **移植說明**：本程式使用 [OpenCode](https://opencode.ai) + [Oh-My-OpenAgent](https://github.com/OhMyOpenCode/oh-my-openagent) + [MiniMax M2.5](https://www.minimaxi.com) LLM 進行移植開源。

---

## 修復點對照表

| # | 修復點 | Lossless-Claw-Enhanced | clawlcm 狀態 | 緊急性 |
|---|-------|---------------------|---------------|-------------|-------|
| 1 | CJK Token 估算 | `CJK = chars * 1.5` | ✅ `CJK = chars * 3/2` | 已修復 |
| 2 | Auth 錯誤 false-positive | `stripAuthErrors()` | ✅ 已新增 | 已修復 |
| 3 | Session Rotation | 偵測 `/reset` 後觸發 | ✅ 已清理 debt | 已修復 |
| 4 | 空訊息跳過 | 跳過 empty/aborted | ✅ 已新增 | 已修復 |

---

## 詳細修復記錄

### 1. CJK Token 估算 ✅

**Enhancement 原文說明：**
```
原始 bug: Math.ceil(text.length / 4) = tokens
- 每個 CJK 字元應該是 ~1.5 tokens

Enhancement 修復:
CJK: chars * 1.5
Emoji/Supplementary: chars * 2.0
```

**clawlcm 修復後 (tokenizer/tokenizer.go:92-112)：**
```go
func EstimateTokens(text string) int {
    runeCount := 0
    isCJK := false
    isEmoji := false
    for _, r := range text {
        if r >= 0x4E00 && r <= 0x9FFF {
            isCJK = true
        }
        if r >= 0x1F300 && r <= 0x1F9FF {
            isEmoji = true
        }
        if r >= 0x10000 && r <= 0x1FFFF {
            isEmoji = true
        }
        runeCount++
    }

    if isCJK {
        return (runeCount * 3) / 2  // ✅ 1.5x
    }
    if isEmoji {
        return runeCount * 2        // ✅ 2.0x
    }
    return (runeCount / 5) * 4 / 3
}
```

**差異比較：**
| 文字 | Enhancement | clawlcm (修復前) | clawlcm (修復後) |
|------|-------------|------------------|------------------|------------------|
| "這個項目" (4字) | 6 tokens | 2 tokens | **6 tokens** ✅ |
| "測試" (2字) | 3 tokens | 1 token | **3 tokens** ✅ |
| 😀 (1 emoji) | 2 tokens | ~1 token | **2 tokens** ✅ |

---

### 2. Auth 錯誤 False-Positive ✅

**Enhancement 說明 (PR #178)：**
當對話中提到 "401 errors"、"API keys" 等關鍵字時，summarizer 會誤判為認證失敗，導致 compaction 中斷。

**clawlcm 修復後 (llm/client.go:129-143)：**
```go
func stripAuthErrors(text string) string {
    lines := strings.Split(text, "\n")
    var filtered []string
    for _, line := range lines {
        lower := strings.ToLower(line)
        if strings.Contains(lower, "http") && strings.Contains(lower, "error") {
            continue
        }
        if (strings.Contains(lower, "401") || strings.Contains(lower, "403") || 
            strings.Contains(lower, "api key")) &&
            !strings.Contains(lower, "discussing") && !strings.Contains(lower, "about") {
            continue
        }
        filtered = append(filtered, line)
    }
    return strings.Join(filtered, "\n")
}
```

---

### 3. Session Rotation 偵測 ✅

**Enhancement 說明 (PR #190)：**
當使用 `/reset` 或 rotate 指令後，compaction 不會在新 session 觸發，導致 context 無限成長。

**clawlcm 修復後 (engine.go:904-958)：**
```go
func (e *Engine) maintainRotate(sessionKey *string, resp *types.MaintainResponse) error {
    // ... 旋轉對話會清除維護 debt
    if e.maintenanceDebt != nil {
        e.maintenanceDebt.ClearDebt(conv.ID)
        e.log.Info("Maintenance debt cleared after rotation")  // ✅
    }
    // ...
}
```

---

### 4. 空訊息跳過 ✅

**Enhancement 說明 (PR #172)：**
API 500 錯誤會產生空訊息，這些訊息累積會造成 feedback loop，永久破壞 agent。

**clawlcm 修復後 (engine.go:332-343)：**
```go
func (e *Engine) Ingest(ctx context.Context, req types.IngestRequest) (*types.IngestResponse, error) {
    // ...
    if req.Message.Role == "assistant" && strings.TrimSpace(req.Message.Content) == "" {
        e.log.Info("Skipping empty assistant message")
        return &types.IngestResponse{
            MessageID:     0,
            Ordinal:       0,
            TokenCount:    0,
            ShouldCompact: false,
        }, nil
    }
    // ...
}
```

---

## 參考來源

- **原始增強**：https://github.com/win4r/lossless-claw-enhanced
- **上游專案**：https://github.com/Martian-Engineering/lossless-claw
- **LCM 論文**：https://papers.voltropy.com/LCM

---

## 移植工具

本程式使用以下工具進行移植：
- **OpenCode**：AI 程式碼編輯環境
- **Oh-My-OpenAgent**：Sisyphus AI Agent 框架
- **MiniMax M2.5**：大型語言模型

---

## 總結

四個修復點已全部移植完成：

| # | 修復點 | 檔案 | 狀態 |
|---|-------|------|------|
| 1 | CJK Token | tokenizer/tokenizer.go | ✅ |
| 2 | Auth Errors | llm/client.go | ✅ |
| 3 | Session Rotation | engine.go | ✅ |
| 4 | Empty Messages | engine.go | ✅ |