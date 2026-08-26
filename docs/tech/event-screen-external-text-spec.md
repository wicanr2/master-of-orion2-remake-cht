# 事件／勘查快報外部文案與版面規格

## 文案契約

- `eventscreen.go` 只保存 `event.*` 語意鍵、座標、顏色、typed report 與資產索引；固定玩家文案
  全部由 `assets/i18n/ui.json` 提供。
- 必要鍵：`event.transition.summary`、`event.header.gnn`、`event.header.survey`、
  `event.tag.alert`、`event.tag.good_news`、`event.tag.discovery`、`event.button.continue`。
- 事件／發現名稱和敘述是 session 的動態玩法資料。畫面只依語言挑選已保存的中英文欄位，
  不在 UI 裡重新發明事件內容。

## 資產與 fallback

- 可解碼 `EVENTS.LBX#0/#1` 時，`#0` 提供調色盤，`#1` 的 31 個 delta 幀以累積方式播放。
- `LastEventReport.EventID` 只有在 `0 <= ID < 36` 且 `ID+2 < archive count` 時才可載入插圖；
  插圖繪於原版已證實的 `(320,14)`。負數測試 fixture 不得索引資產；畫廊以事件 0
  的外部雙語固定報告驗證 `ID+2` 插圖，但不進玩法規則。
- 缺檔、缺 palette、解碼失敗、錯誤 count 或非法 ID 時，必須退回自繪安全面板；不可 panic、
  警告洗版或用版本常數假定 archive shape。
- 原版動畫速率尚未證實；每 3 tick 一幀明標 remake timing approximation，不列為 parity。

## 文字安全框

- fallback 台標：`(60,110,520,26)`，左右各 12px；單行省略。
- fallback 標題：`(76,148,488,24)`；標記與標題合成後單行省略。
- fallback 正文：`(76,184,488,156)`，`lineH=20`，最多 7 行，最後一行省略。
- 原版背景模式的報告底板置於下方不遮擋主播與 `(320,14,157,125)` 插圖；標題、正文與按鈕
  各有獨立安全框。未取得原版中文 baseline 前，不把此底板位置宣稱為逐像素原版。
- 「繼續」按鈕文字框與 `(270,372,100,24)` 可見按鈕同中心；點擊則依原版接受整個
  640×480，不以按鈕矩形縮小操作範圍。
- 所有文字只能透過 `textSafeRect` 繪製；`eventscreen.go` 禁止直接呼叫 `fnt.Draw*`。

## 驗收

- 七個固定鍵在繁中／英文均存在且不是 key fallback。
- 最長合理中英文標題、標記、正文與按鈕在 runtime bitmap font 下不越過各自安全框。
- 原版資產模式與 fallback 模式各有測試；正常事件、勘查報告與負數事件 ID 各抽一條。
- 中文畫廊與正版資料畫廊各抽查 `05_event.png`；單元測試只證 remake 自洽，不取代原版畫面證據。
