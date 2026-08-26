# 最終得分外部文案、原版背景與版面規格

## 文案契約

- `hiscore.go` 只保存座標、顏色、typed `VictoryCondition`、`hiscore.*` 鍵及分數值；固定玩家
  文案全部由 `assets/i18n/ui.json` 提供。
- 必要鍵包括：背景標題、勝／敗標題、三種勝利原因、未知原因、勝／敗摘要格式、九個分數列、
  繼續提示及回合摘要轉場。
- `internal/shell.ScoreLine` 保存穩定 `TextKey`，不得再保存中文 `Label`。UI 才依目前語言解析
  `TextKey`；規則層不能反向依翻譯文字作判斷。

## 資產與 fallback

- `SCORE.LBX#0` 只有在 archive count 至少 1、影像恰為 640×480、有可用內嵌 palette 且能
  解碼時才作主要背景。
- 原版背景模式不得再鋪 640×480 黑遮罩或 460×360 自製外框；只清理內部黑色內容區，保留
  外圍機械框、雕像與底部控制列。
- 英文模式保留烘入背景的 `HALL OF FAME`；繁中模式在文字牌內安全覆蓋為
  `hiscore.header.hall_of_fame`。
- 缺資產時沿用現有 `TURNSUM.LBX` 自繪面板，所有文案與安全框政策相同；不得 panic 或警告洗版。
- `SCORE.LBX#1..14` 在 race icon writer／consumer 未閉合前不任意顯示。

## 安全框與輸入

- 原版背景內容區限制在約 `(136,122)..(504,414)`。勝負標題、摘要、九列分數與提示分別有
  不重疊安全框；分數 label 由 x=140 起，符合 IDA 已證實的原版起點。
- label 與數值分欄；數值靠右且不能進入 label 欄。九列含總分必須在底部提示前結束。
- fallback 保留現有 `(90,60,460,360)` 面板，但同樣透過 `textSafeRect` 收束。
- 背景標題只在繁中覆蓋，覆蓋框不得超出上方文字牌；英文不可重畫第二份標題。
- 畫面任意位置點擊均可繼續；可見提示不是唯一熱區。按鈕／提示文字與它所屬安全框共用中心。
- `hiscore.go` 禁止直接呼叫 `fnt.Draw*`；靠右數值應由共用安全框 helper 完成。

## 驗收

- 所有 `hiscore.*` 鍵繁中／英文皆存在，九個 `ScoreLine.TextKey` 可解析。
- 最長勝敗摘要、32-bit 分數與雙語列名在 runtime bitmap font 下不越框。
- 原版背景／fallback、勝／敗與未知 victory reason 各有測試；整張畫面點擊有輸入測試。
- 正版資產及 fallback 畫廊各產出 35/35，人工檢查 `11_hiscore.png` 的上方文字牌、九列、總分、
  提示與框架。這是 remake 視覺驗收，不等於 Hall of Fame 排名資料逐值 parity。

