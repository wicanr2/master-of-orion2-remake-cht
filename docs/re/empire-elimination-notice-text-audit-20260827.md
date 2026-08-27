# 帝國滅亡事件 29 通知文字稽核（2026-08-27）

## 證據基線

- **已證實**：`docs/re/event-status-broadcasts-audit-20260825.md` 以 IDA Pro 9.4 證實
  `sub_E4EB3 @ IDA linear 0xE4EB3` 掃描原先 active、現已沒有有效殖民地的帝國，並由
  `sub_233AB @ 0x233AB` 建立事件 29。原版 record 保存帝國 slot 與 state 1。
- 輸入為 `Orion2.exe` 1.31，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`；工具為
  IDA Pro 9.4，位址空間為 IDA linear、DOS/4GW LE object #1。
- **已證實（稽核前 remake source）**：`detectEmpireEliminationBroadcasts` 已保存
  `TargetKind/TargetIndex/TargetName` 與可存檔 `EmpireAlive` 去重狀態，但同時建立中英文成品句子。
- **已證實（remake source）**：事件畫面、回合摘要與 INFO 已共用 `eventReportMessageText`，
  事件 29 可沿同一 typed 顯示入口遷移。

## 實作結論與限制

- 事件 29 只需單一帝國 target 即可重建目前通知；規則層不再保存成品句子。
- 原版會從四種文案中隨機選一種；現行 remake 只有單一等義摘要。模板列為 **remake 介面轉接**，
  不宣稱已復原原版四種逐字文案或其亂數消費位置。
- 本切片不改 active→inactive 觸發、清理順序、去重陣列、事件佇列或存檔格式。
- `Message`／`MessageEN` 仍保留給其餘事件與舊存檔；事件 29 typed 欄位缺失時走舊訊息 fallback。

