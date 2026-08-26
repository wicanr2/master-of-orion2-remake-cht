# 種族選擇玩家流程與文案邊界稽核（2026-08-27）

## 問題

種族選擇已採原版左右配置，但 14 族中英文名稱、形容詞、自撰能力摘要、標題、取消與帝國名稱
仍內嵌於 Go。現有文件也同時保留外部符號的舊位址與修正位址，必須先固定 raw 定位，才能判定
哪些欄位是原版畫面、哪些是 remake 補充。

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址為 DOS/4GW 映像的 IDA linear address。
- 可重生匯出：`tools/ida/audit_race_selection_ui.py`；保留 raw 名稱、函式邊界、caller、指令
  bytes、資產字串與交叉參照。

## 已證實

- 原始函式邊界是 loader `sub_5BC74 @ 0x5BC74..0x5BD97`、draw
  `sub_5BD97 @ 0x5BD97..0x5BEE3` 與外層 `sub_5C510 @ 0x5C510..0x5CF37`。
  外層由熱座／新局等四個 caller 進入，並在四條分支呼叫同一 loader。
- `sub_5C510` 的 raw `0x5C5DF..0x5C639` 建立 14 個按鈕：
  `y = 0x5A + 0x30 × (i mod 7)`，`x = 0x15F + 0x7E × (i div 7)`；
  `sub_11523B` 接收每筆 `dword_19B7A4[i]` 與按鈕狀態 `word_19B820[i]`。
- `0x5C5BE..0x5C5C8` 明確以單位元字串 `0x1B` 呼叫 `sub_114C72`，原版取消輸入是 ESC，
  不是可見按鈕。
- `sub_5BD97` 在 `0x5BDCF` 取 `RACESEL.LBX#33` 並以 `(366,52)` 畫標題；
  `0x5BE50`／`0x5BE74` 以 `15 + raceIndex` 取肖像，後一路以 `(54,63)` 畫圖。
- `RACESEL.LBX` 的字串交叉參照集中在 loader／draw 與相鄰 helper；正版資產結構交叉驗證為
  `#1..14` 123×45 雙幀按鈕、`#15..28` 290×322 肖像。

## 符號勘誤與分級

- 外部 `func_names.txt` 把三個名稱放在 `0x5BD97／0x5C20E／0x5CF37`；修正後的
  `symbols_fixed.tsv` 與 IDA 邊界則是 `0x5BC74／0x5BD97／0x5C510`。本文件保留 raw 位址，
  不以衝突名稱取代定位。
- 左肖像、右側 2×7 按鈕、標題資產、hover 選擇與 ESC 取消為**已證實**。
- 肖像下方的 164×40 能力摘要與可見 CANCEL 按鈕是 **remake adapter**；原版同區文字只在
  特定多人占用分支出現，不能把 adapter 當成原版精確版面。
- 13 族與 Custom 的英文按鈕文字烘在正版資產；繁中名稱與 adapter 摘要是外部顯示資料，
  不冒稱原版逐字內容。

## Remake 映射

- `raceSelectList` 只保存穩定 race key、肖像 ID 與 `shellIdx`，不保存玩家文字。
- 名稱、英文形容詞、摘要、標題、取消、帝國格式與轉場由 `assets/i18n/ui.json` 提供。
- 英文模式仍露出原版按鈕與標題烘字；外部英文名稱只用於 adapter 資訊、帝國預設名及資料匹配。
- 按鈕、標題、能力摘要與 CANCEL 皆使用固定雙軸安全框；玩法設定與種族加成不變。
