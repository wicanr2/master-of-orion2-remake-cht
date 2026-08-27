# 回合摘要外部文案與安全版面規格

## 資料契約

- `internal/shell` 只回傳經濟數值、serious 結果及 `BuildNotice`；不得組玩家句子。
- `BuildNotice` 種類固定為一般完工、改裝完成、改裝取消報廢及人造行星完成，參數為殖民地
  索引與建造／艦艇／行星名稱。
- 所有固定文案及格式模板放在外部 `ui.json`，兩種語言的格式參數數量與順序必須相同。
- 建造項目用 `buildItemLabel` 轉譯；動態艦名與行星名不得複製進文案目錄。

## 版面契約

- 標題、關閉按鈕沿用 `misc.json` overlay；轉場名稱使用 `turnsummary.transition.galaxy`。
- 四條基礎經濟列使用不相交的 `textSafeRect`；動態區固定 x=40、y=168、寬 320、底界 306。
- 動態訊息以 19px 行距換行。內容超過可用行數時，只保留可見列，末列以實際字型量測後加
  省略號；不得靠 painter clipping 遮住溢位。
- 所有動態來源共用同一個追加器，禁止直接 append 無界 `extraText`。

## 驗收

- `turnSummary()` 來源切片不得含 `.tr(` 或固定玩家句子；`LastBuilt` 不得是 `[]string`。
- catalog 鍵齊全、格式參數契約一致，typed 建造通知四種均有雙語格式化測試。
- 最壞情境（破產、飢荒、叛亂、研究、多項完工、事件、安塔蘭、突襲）產生的每一列都在
  動態安全框內，且不侵入關閉按鈕。
- `cmd/moo2`、`internal/shell` 全套測試通過，繁中與英文畫廊 `06_turnsummary.png` 實圖抽查。
