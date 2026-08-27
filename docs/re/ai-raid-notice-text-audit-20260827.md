# AI 殖民地突襲通知文字稽核（2026-08-27）

## 證據基線

- 原版 1.31 的目標估值鏈已有位址證據：`Colony_Worth_To_Player_ @ 0xD2CAE`、
  `Enemy_Colony_Worth_To_Player_ @ 0xD8D11`、`Proximity_Worth_To_Player_ @ 0xD2AEA`；
  AI 艦隊航行／抵達接線的歷史與限制見 `docs/re/01-gap-report.md` 第 47 項。
- **已證實**：現行 `AIRaidReport` 已保存 AI 名、星名、殖民地索引、是否擊退、人口／BC／建築／
  艦力損失；這些欄位足以重建目前兩種玩家通知。
- **已證實（稽核前 remake source）**：`internal/shell/ai_attack.go` 仍以 `fmt.Sprintf` 組合
  `Message`／`MessageEN`，再複製中文句子到 `GameSession.LastRaid`。回合摘要與 INFO 各自選擇
  中文成品或英文成品，翻譯責任因此滲入規則層。
- **已證實（稽核前 remake source）**：被摧毀建築保存的是既有規則名稱；英文成品句子直接
  插入該名稱，可能把中文建築名帶入英文通知。

## 證據限制與實作結論

- 原版證據支持目標估值與艦隊抵達，不支持現行 `aiRaidGraceTurns`、損失公式或通知逐字；
  擊退／突破句型維持 **remake adapter**。
- `AIRaidReport` 保留 typed 結果，移除 `Message`／`MessageEN`；`GameSession.LastRaid` 亦移除，
  是否有突襲一律以 `LastRaidReport != nil` 判定。
- UI 使用 `raid.notice.*` 外部模板；建築名在顯示層經既有 building catalog 翻譯，避免英文模式
  洩漏中文規則鍵。
- 本切片不改出兵意願、航程、目標估值、防禦、損失、建築摧毀或存檔節奏。

