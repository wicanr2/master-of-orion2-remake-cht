# 種族選擇外部文案與版面規格

## 資料分層

`raceSelectList` 每筆只保留穩定 `key`、`portrait`、`shellIdx`。顯示名稱、帝國形容詞與能力摘要
使用 `race.select.race.<key>.*` 查詢 `assets/i18n/ui.json`；不得把中文、英文或翻譯後字串當成
列表識別欄位。

`shell.Races` 繼續保存規則所需索引與既有名稱相容欄位，但種族選擇畫面不得讀其中的 `Desc`／
`EnDesc` 組裝玩家文字。外交舊存檔名稱比對可比對 shell 相容名稱與外部雙語名稱，不能依賴
翻譯字串的固定語言。

## 外部鍵

- `race.select.title`、`race.select.cancel`、`race.select.transition.new_game`。
- `race.select.empire_name`：單一 `%s`，填入 `.adjective`。
- `race.select.race.<key>.name`：按鈕／資訊框顯示名；英文為原版複數族名。
- `race.select.race.<key>.adjective`：帝國名稱使用的單數形容詞。
- `race.select.race.<key>.description`：一至兩行 adapter 能力摘要；Custom 說明亦走同一鍵形。

所有 14 筆必須具備中英文三欄，格式模板參數數量與型別一致。

## 安全框

- 中文標題安全框與 `RACESEL.LBX#33` 的 219×29 矩形共用中心。
- 14 顆按鈕文字框由 123×45 熱區內縮 6px推導，中心不得漂移；英文模式不覆蓋烘字。
- 164×40 adapter 資訊框分成 18px 名稱列與 20px 單行摘要區；摘要使用精簡規則值並須完整
  顯示，不接受兩行字墨互相擠壓或以省略號隱藏規則。
- CANCEL 的 120×28 文字框由 adapter 熱區內縮 4px 推導，文字須完整顯示。
- 所有欄位同時限制寬高；文案不得改變原版按鈕位置、hover 熱區或新局狀態。

## 驗證

- IDA raw 位址、14 個欄位、資產索引及 ESC 輸入有可重生證據。
- catalog 14×3 鍵與畫面靜態鍵雙語齊全；帝國格式不產生 `%!`。
- `raceselect.go` 不呼叫 `.tr(`，且不內嵌玩家句子；按鈕／CANCEL 中英文經 bitmap glyph 量測。
- 中英文畫廊各 35 張，抽查 `02_raceselect.png` 的標題、14 顆按鈕、資訊框與 CANCEL。
