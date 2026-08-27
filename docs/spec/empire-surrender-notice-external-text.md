# 帝國投降事件 34 外部通知規格

## 資料契約

- `queueEmpireSurrender` 只寫 `EventID=34`、投降者 `TargetKind/TargetIndex/TargetName` 與接收者
  `SecondaryTargetKind/SecondaryTargetIndex/SecondaryTargetName`；不得寫 `Message`／`MessageEN`。
- pending surrender 與事件 report 必須同時入列，資產仍由下一個 surrender consumer 延後接收。
- 其他事件與舊存檔仍可保存成品訊息，本切片不得以全域刪欄破壞相容性。

## 顯示契約

- `eventReportMessageText` 是事件畫面、回合摘要與 INFO 摘要的共同入口。
- `EventID=34` 使用 `event.status.surrender` JSON 模板；合法 AI target 依 `RaceIndex` 顯示目前語言
  的原版種族名，玩家與熱座名稱保留玩家自訂名稱，資料無法解析時使用 report fallback 或
  `common.unknown_empire`。
- 非事件 34 依語言沿用 `Message`／`MessageEN`；英文欄缺值時安全回退 `Message`。

## 驗證契約

- AI→AI、AI→玩家、AI→熱座與缺 target 的事件 34 通知均有測試；英文 AI 名不得洩漏繁中種族名。
- 三個 UI 消費端不得直接選 `MessageEN`，必須呼叫共同入口。
- 規則來源靜態測試禁止投降通知中英文成品句子重新出現。
- 既有 pending／延後轉移／快照測試、聚焦 UI 測試、雙語畫廊、`git diff --check`、擁有權與
  Docker 清理均須通過。
