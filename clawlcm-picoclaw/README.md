# clawlcm-picoclaw

clawlcm 的 skill 部署目錄，基於 lossless-claw-enhanced 增強版。

## 目錄結構

```
clawlcm-picoclaw/
├── README.md            # 本文件
├── clawlcm/            # Skill 部署包
│   ├── SKILL.md        # Skill 定義 (中文)
│   ├── SKILL_en.md    # Skill 定義 (英文)
│   ├── bin/
│   │   └── clawlcm   # 主程式 (~15MB)
│   └── data/
│       └── config.json # 配置模板
└── ../clawlcm/        # 源碼 (上級目錄)
```

## 對照 picolcm-picoclaw

| 項目 | picolcm-picoclaw | clawlcm-picoclaw |
|------|-----------------|-----------------|
| 版本 | v0.3.1 | **v0.8.1** |
| 基於 | lossless-claw | **lossless-claw-enhanced** |
| Token 估算 | 標準 | **1.5x CJK, 2x Emoji** |
| Auth Error | - | **✅ 過濾** |
| Session Rotation | - | **✅ 檢測** |
| 空訊息 | - | **✅ 跳過** |
| MaxDepth | - | **✅ 8** |

## 使用方法

參考 `clawlcm/SKILL.md` 或 `clawlcm/SKILL_en.md`。

## 部署到 Container

```bash
# 複製到 workspace
mkdir -p /app/workspace/skills/lcm
cp -r clawlcm/* /app/workspace/skills/lcm/

# 配置 LLM
vim /app/workspace/skills/lcm/data/config.json
```

## 版本信息

- **Version**: v0.8.1
- **更新日**: 2026-04-15
- **Commit**: enhanced-fix
- **參考**: [lossless-claw-enhanced](https://github.com/win4r/lossless-claw-enhanced)