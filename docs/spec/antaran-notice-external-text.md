# 安塔蘭戰略入侵通知外部文字規格

## 資料契約

- `AntaranNotice` 只保存 `Kind`、`StarName`／`StarNameEN`、`ETA`、`ShipsLost` 與
  `Repelled`；規則層不得保存玩家句子或語言分隔符。
- `advanceAntares` 每個世界回合開始先清空通知；成功出兵或艦隊抵達時寫入一筆型別化結果。
- 熱座切換必須讓通知隨目標席位保存；切回原席位後不得遺失剛完成的結果。
- UI 以穩定 `antaran.notice.*` 鍵從 `ui.json` 格式化。英文模式優先 `StarNameEN`，缺值時
  安全回退 `StarName`。

## 顯示契約

- 顯示分支包含：已出發、抵達 AI 星系交戰、抵達未設防星系、玩家守軍擊退、玩家守軍未擊退。
- 回合摘要與 INFO 摘要必須共用 `antaranNoticeText`，不得各自保存第二份句型。
- `Kind` 未知時使用外部 `antaran.notice.unknown`，不得把語意鍵直接顯示給玩家。
- 既有回合摘要動態列與 INFO 事件列繼續使用雙軸安全框及其截斷策略；本切片不新增自由繪字。

## 驗證契約

- 五個分支在繁中與英文均能產生非鍵值文字，動態數值與英文星名正確。
- `internal/shell/antaran_invasion.go` 不得再出現中英文通知句子或 `fmt.Sprintf` 玩家文案。
- 重要摘要 gate 讀型別化通知；熱座席位往返保持通知。
- 聚焦測試涵蓋規則、顯示、熱座與來源防回歸；雙語畫廊抽查回合摘要／INFO 安全框。

