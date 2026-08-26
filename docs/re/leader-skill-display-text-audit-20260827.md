# 領袖技能顯示文字稽核（2026-08-27）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4／IDAPython，映像 `ida-pro-9.4-idapython:py312-v1`；位址是 IDA linear、DOS/4GW LE object #1。
- 本輪以既有 `.i64` 的 `/tmp` 可寫副本重跑 `tools/ida/audit_random_officer_recruitment.py`；原始資料庫唯讀掛載，未改名或覆寫註記。
- 英文技能名稱另由 `GAME_MANUAL.pdf` p.135–137 錨定；繁中譯名屬 remake 顯示資料，不冒稱原版 executable 內含中文。

## 已證實

1. 領袖通用技能與專屬技能是數字 ID；HERODATA 的每項技能各佔 2 bit，值為 0／1／2 階。顯示文字不是規則識別鍵。
2. 艦艇軍官與殖民地領袖分別使用 captain／administrator 專屬技能段，再接通用技能段；既有 `LeaderSkillIDsFor` 順序與原版技能欄一致。
3. 招募、技能效果及 `.GAM` 匯入已有 typed ID 消費端；中文或英文名稱只需要在顯示邊界選擇，不應寫在 Go 的規則表內。

## 工程結論

- 27 個技能的原版 ID 留在 Go；穩定鍵、中英文名稱移到 `internal/gamedata/leader_skills.json`。
- 無技能領袖的「艦長／Ship Officer」與「行政官／Administrator」不是技能，放在 `assets/i18n/ui.json`，避免再次和 Commando 技能混成同一識別字串。
- 本輪沒有追查 Win95 字型或文字繪製 API；依專案停止線，這些平台內部行為不影響技能資料與玩家可見輸出契約。
