# 安塔蘭戰略入侵通知文字稽核（2026-08-27）

## 證據基線

- 輸入、工具、位址基準與原版玩法證據沿用
  [`antaran-periodic-invasion-audit-20260825.md`](antaran-periodic-invasion-audit-20260825.md)：
  `Orion2.exe` 1.31，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`，
  IDA Pro 9.4，IDA linear／DOS4GW LE object #1。
- **已證實**：`sub_643A0` 建立最多五艘 owner 8 艦隊，`sub_64311`／`sub_645A4`
  選擇目標殖民星；抵達後進一般航行／戰鬥鏈。這些證據支持通知中的目標星、出發與戰鬥結果
  資料，不支持任何現行中英文句子是原版逐字。
- **已證實（稽核前 remake source）**：`internal/shell/antaran_invasion.go` 在四個規則分支直接
  組合繁中與英文成品句子，再由回合摘要及 INFO 摘要選取其中一份。規則狀態因此同時承擔
  語言與玩家文案責任。
- **已證實（稽核前 remake source）**：熱座席位快照保存兩份成品字串；一般 JSON 存檔刻意
  排除這類本回合顯示暫態，因此不存在需要維持的公開存檔欄位格式。

## 結論與證據等級

- 安塔蘭通知改保存型別化結果：通知種類、雙語星名、ETA、我方損失與是否擊退。
- 中英文句型與警示符號放入 `assets/i18n/ui.json`；兩個 UI 消費端共用同一個格式化入口。
- 出發、AI 星系交戰、未設防抵達、玩家守軍擊退／未擊退五種顯示分支是 **remake adapter**。
  原版逐字、通知排序與是否對所有帝國公開仍為 **未知**，不得由現行畫面或綠測試升格。
- 本切片不改 `AntaranInvasionState`、目標權重、ETA 近似、owner 8 戰力映射或快速戰鬥。

