# 外交會談請求燈：排列、動畫與遮罩稽核

## 輸入與方法

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址是 IDA linear、DOS/4GW 映像。
- 可重生匯出：`tools/ida/audit_diplomacy_request_lights.py`。腳本唯讀分析 `/tmp` 資料庫
  副本，保留原始函式名、位址、指令及 bytes。

## 已證實

`sub_FA795 @ 0xFA795..0xFA79B` 只執行
`A0 54 B0 1A 00: mov al, byte_1AB054` 後返回；`sub_83D06` 是唯一直接 caller。
因此 `byte_1AB054` 是請求燈繪製端使用的八位遮罩，但「每個 bit 的完整 AI 觸發政策」不由此
函式證明。

`sub_83D06 @ 0x83D06..0x83DEA` 有兩個直接 caller。它逐一掃描
`di < word_199998`，以 `mask >> di & 1` 決定是否繪製該帝國：

- `0x83D35..0x83D46`：以帝國 stride `0xEA9` 讀 `player+0x26`；
- `0x83D4E..0x83D67`：`x = 0x1FA - drawnCount * word[dword_19C128+0]`；
- `0x83D71`：從該種族動畫物件 `+4` 讀一個 word，精確語意維持未知；
- `0x83D78..0x83D98`：用 `byte_19C148[race]` 選 frame，於 `(x,5)` 繪製；
- `0x83D9D..0x83DB8`：frame 加一，達動畫物件 `+6` 的 frame count 後歸零；
- `0x83DCE`：只有實際畫出一盞燈才增加 `drawnCount`。

## 勘誤

舊 Go 註解與測試把 `0x1FA` 說成「最右一盞燈的右緣」，並實作
`x=506-(n+1)*width`。原始指令先把 `drawnCount=0` 代入，再於 `0x83D8C` 把算出的值直接
當圖片 x 座標，因此第一盞燈的**左緣**是 506；正確公式是 `x=506-n*width`。排列步距統一
取動畫物件 0 的寬度，不是逐種族動畫各自取寬。

## Remake 邊界

- 原版使用逐種族動畫；現行資產來源仍未知。remake 的 22×16 旗色方塊、來意色與單字 glyph
  是明標的資訊轉接，不是原版圖像還原。
- `war／trade／alliance` 是 remake typed reason。原版遮罩只證明「哪些帝國亮燈」，不證明
  燈本身向玩家揭露來意；因此 glyph 不得列為原版已證實行為。
- 點擊方塊後直接進入該對手外交畫面是 remake 正常玩家路徑；本函式只提供繪製證據，沒有
  原版 widget 熱區證據。

