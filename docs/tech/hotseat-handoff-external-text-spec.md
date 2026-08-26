# 熱座交接畫面外部文案與安全框規格

## 定位

熱座交接是 remake 的同機隱私轉接畫面。它不是已證實的原版
`sub_628E2` 逐像素複製；原版證據邊界見
[`hotseat-handoff-ui-audit-20260827.md`](../re/hotseat-handoff-ui-audit-20260827.md)。

## 文案契約

Go 只能保存下列 `ui.json` 語意鍵，不得保存中英文玩家句子：

| 鍵 | 用途 |
|---|---|
| `hotseat.handoff.title` | 畫面標題 |
| `hotseat.handoff.next_player` | 下一位玩家模板，單一 `%s` |
| `hotseat.handoff.seat_position` | 目前席位／總席位模板，兩個 `%d` |
| `hotseat.handoff.instruction` | 隱私交接操作說明 |
| `hotseat.handoff.note.resolve` | 全員下令後的結算提示 |
| `hotseat.handoff.button.take_over` | 接手按鈕 |
| `hotseat.transition.galaxy` | 接手後返回星圖的轉場標籤 |
| `hiscore.transition.screen` | 勝負後進入最終得分的轉場標籤 |

玩家名稱是存檔資料，經 `hotseatNameLabel` 做語言安全顯示後才填入模板。note 不接受任意已翻譯
句子，只接受空鍵或已知 catalog 鍵。

## 幾何契約

- 邏輯畫布固定 640×480；自繪視窗採 360×230 並整數置中。
- 標題、玩家列、席位列、兩行操作說明、兩行提示與按鈕分別具有獨立
  `textSafeRect`，高度不足時省略而不跨列。
- TAKE OVER 文字安全框與 110×30 點擊熱區共用相同整數中心。
- 畫面仍先全黑，再畫面板；上一席星圖不得透出。

## 驗證

- 雙語 catalog 鍵非空，格式參數可正常展開。
- 最長合理玩家名、英文說明與繁中提示在 bitmap runtime 字型下不超出各自安全框。
- 靜態測試禁止 `hotseat.go` 出現 `.tr(`、直接字型繪製及本規格列出的玩家句子。
- 熱座席位輪替、最後一席結算、`cmd/moo2` 測試與 35 張中文畫廊必須通過。
