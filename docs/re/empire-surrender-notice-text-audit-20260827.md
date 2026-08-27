# 帝國投降事件 34 通知文字稽核（2026-08-27）

## 證據基線

- **已證實**：`docs/re/empire-surrender-audit-20260825.md` 以 IDA Pro 9.4 證實
  `sub_E4D06 @ IDA linear 0xE4D06` 先寫 pending surrender、`sub_233D2 @ 0x233D2`
  建立事件 34，後續 `sub_E4DC9 @ 0xE4DC9` 才呼叫 `sub_E4B5F @ 0xE4B5F` 接收資產。
  輸入 `Orion2.exe` SHA-256 為
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`；位址空間為
  IDA linear、DOS/4GW LE object #1。
- **已證實（稽核前 remake source）**：`queueEmpireSurrender` 已把投降者與接收者保存為
  `EventReport.Target*`／`SecondaryTarget*`，但同時以 `fmt.Sprintf` 建立中英文完整句子。
- **已證實（稽核前 remake source）**：事件畫面、回合摘要與 INFO 摘要各自直接選
  `Message`／`MessageEN`；若事件 34 改成 typed-only，三處必須共用同一格式化入口。

## 實作結論與限制

- 事件 34 已有足夠的雙 target kind／index 作 typed 顯示輸入；規則層不再建立成品句子。
- 原版證據證實事件種類與兩帝國關係，不證實 remake 的「帝國接收程序即將開始」逐字；外部模板
  維持 **remake 介面轉接**，不升格為原版原文。
- `EventReport.Message`／`MessageEN` 暫時保留給其餘尚未遷移事件與舊存檔；事件 34 的 UI
  優先依 typed target 格式化，資料不足時才安全回退既有欄位。
- 本切片不改自動投降近似 gate、事件佇列、延後 consumer、資產轉移或存檔格式。

