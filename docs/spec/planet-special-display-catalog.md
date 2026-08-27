# 行星特殊物產與勘查報告外部文字規格

## 資料契約

- `PlanetSpecialTextKey` 只回穩定語意鍵；`NoSpecial`、越界值回空鍵。
- `planetSpecialLabel` 以 `ui.json` 解析語意鍵；三個玩家畫面共用此入口。
- 殖民地與行星列表的特殊物產列使用 `planet.special.marked`，不得自行拼接 glyph。
- `SystemDiscovery` 新資料以 `Special`、`BCGained`、`Population`、`ColonyIdx`、`LeaderGot`、
  `TechTopics` 表達；規則層不得組合玩家句子。
- UI 依 typed 結果選擇 `event.discovery.*` 模板。舊 Name／Message 欄位只在 typed 資料不足時
  顯示，以維持舊存檔可讀性。

## 驗證契約

- 有名稱的特殊物產在兩種語言均須解析；無特殊物產與越界值不得顯示語意鍵。
- BC、失散殖民地成功／失敗、領袖成功／滿額、科技成功／無候選八條結果都有外部模板。
- 英文科技清單不得含中文頓號；繁中科技名稱須經 `tech.json`。
- `planet_special.go` 不得保存中英文顯示名稱；`discovery.go` 不得保存勘查報告句子。
- 雙語畫廊抽查星圖、殖民地、行星列表與勘查報告的安全框；畫廊固定報告只證版面。
