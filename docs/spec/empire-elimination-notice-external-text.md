# 帝國滅亡事件 29 外部通知規格

## 資料契約

- active 帝國轉為無有效殖民地時，只排入 `EventID=29` 與該帝國的
  `TargetKind/TargetIndex/TargetName`；不得建立 `Message`／`MessageEN`。
- `StatusBroadcast.EmpireAlive` 仍是唯一去重狀態；同一帝國不得因訊息外部化而重播。
- 舊存檔與其他事件的成品訊息欄保持相容，不在本切片全域刪除。

## 顯示契約

- `eventReportMessageText` 遇到具 typed target 的事件 29 時，使用
  `event.status.empire_eliminated` JSON 模板。
- AI 名稱依 `RaceIndex` 顯示當前語言；玩家與熱座保留自訂名稱；無法解析時依序使用 report
  fallback 與 `common.unknown_empire`。
- 事件 29 缺少所有 typed target 欄位時，視為舊存檔並沿用 `Message`／`MessageEN`。

## 驗證契約

- AI、玩家／熱座、非法 target 與舊存檔均有顯示測試；英文 AI 名不得洩漏繁中。
- 規則測試證明事件 29 report 的成品訊息為空、target 正確且第二次掃描不重播。
- 靜態測試禁止事件 29 中英文成品句子回到 `events_broadcast.go`。
- 聚焦測試、雙語正常畫廊、`git diff --check`、擁有權與 Docker 清理均須通過。

