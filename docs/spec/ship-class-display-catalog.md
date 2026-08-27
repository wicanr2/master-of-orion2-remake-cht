# 艦級顯示名稱外部目錄規格

## 顯示契約

- 六個戰鬥艦體與四個支援艦各有穩定 `ship.class.*` 語意鍵；中英文值只存在 `assets/i18n/ui.json`。
- `shipClassZH` 與 `Ship.Class` 保持既有規則／存檔鍵，不得直接回傳到玩家畫面。
- 未登記艦級使用 `ship.class.unknown`，不得於英文模式洩漏中文規則鍵。
- 艦艇設計標題及 CLEAR／CANCEL／BUILD 使用 `shipdesign.title` 與 `shipdesign.button.*`。
- 艦艇設計熱區 action 使用 typed 索引 `hull:<0..5>`，不得依顯示語言或顯示名稱判斷。

## 驗證契約

- 十種艦級與未知 fallback 在繁中、英文均須解析成非語意鍵文字。
- 艦隊、改裝與艦艇設計共用同一 `shipClassLabel`。
- `shipDesign()` 來源切片不得內嵌 Ship Design、六個英文艦體名或 Clear／Cancel／Build。
- 雙語艦隊名冊與艦艇設計畫面須以實際字型確認文字留在既有安全框。
