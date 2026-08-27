# 銀河議會回合通知外部文字規格

## 資料契約

- `CouncilNotice` 以列舉區分候選不足、等待真人投票、玩家當選、AI 當選待回應及無人過門檻。
- 結構只保存屆次、候選帝國索引／fallback 名稱、兩方得票、總票與明確 `WinnerSlot`；不得保存
  中英文成品句子，也不得由玩家是否為候選人反推哪一位 AI 當選。
- `advanceCouncil` 每次執行先清空 `LastCouncilNotice`；`RespondToCouncilVote` 依結果覆寫同一 typed
  notice。議會是全銀河新聞，不隨熱座席位切換，沿用既有生命週期。

## 顯示契約

- `councilNoticeText` 是唯一格式化入口，所有固定句型取自 `assets/i18n/ui.json`。
- 候選索引 `-1` 顯示外部 `common.you`；合法 AI 索引依 `RaceIndex` 取得該語言種族名，無法解析
  時才使用 notice 保存的 fallback 名稱，仍無名稱則使用 `common.unknown_empire`。
- INFO 回合摘要只讀 typed notice；空 notice 不產生訊息列。

## 驗證契約

- 五種通知均有繁中與英文格式測試；英文候選名不得洩漏繁中種族名。
- 議會排程、真人投票、勝負與熱座共同新聞測試改查 typed notice，不以文字內容判定規則。
- 靜態測試禁止 `internal/shell/council.go` 再宣告 `LastCouncil` 或內嵌五種玩家通知句型。
- 聚焦規則／UI 測試、雙語正常畫廊、`git diff --check`、檔案擁有權與 Docker 清理均須通過。
