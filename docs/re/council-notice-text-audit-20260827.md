# 銀河議會回合通知文字稽核（2026-08-27）

## 證據基線

- **已證實**：`Council_Votes_ @ IDA linear 0x15EBC`、`Vote_Check_ @ 0x16021`、真人三選一
  `sub_1633C @ 0x1633C` 與當選門檻 `sub_15DF8 @ 0x15DF8` 的資料流已由
  `council-voting-audit-20260824.md` 閉合。輸入為 `mastori2/Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`；工具為
  IDA Pro 9.4.0.260610，位址空間為 IDA linear。
- **已證實（稽核前 remake source）**：`internal/shell/council.go` 在候選不足、等待真人投票、玩家當選、
  AI 當選待回應及無人過門檻五條分支中，以 `fmt.Sprintf` 建立繁中 `LastCouncil` 成品句子。
  `cmd/moo2/infosubscreens.go` 不分語言直接顯示該字串，因此英文 INFO 回合摘要會洩漏中文。
- **已證實（稽核前 remake source）**：議會畫面本身的固定文案已由 `council.*` JSON 鍵提供；本缺口只在
  回合通知資料契約，不改議會投票畫面、排程或公式。

## 結論與證據邊界

- 原版位址證據證實選舉狀態與結果，不證實 remake 回合摘要的逐字句型；五種通知模板均標為
  **remake 介面轉接**。
- 規則層應只留下屆次、通知種類、候選索引／名稱、當選 slot、得票及總票等 typed 結果。是否有通知以
  `LastCouncilNotice != nil` 判定，不保留平行成品字串。
- 玩家稱呼、未知帝國 fallback 與雙語句型由 UI catalog 決定；原始規則與既有存檔欄位不變。
