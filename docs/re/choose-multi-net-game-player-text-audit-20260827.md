# 區網對局選擇畫面玩家文案稽核（2026-08-27）

## 問題與範圍

`cmd/moo2/choosemultinetgame.go` 已依原版 `Choose_Multi_Network_Game_Screen_` 建立十列清單，
但標題、空清單提示、直接位址、取消、連線錯誤與畫廊示範對局名仍是 Go 內的玩家文字。
本輪只處理這張清單；後續席位名冊 `choosenetplyrs.go` 另有原版欄位與現代 roster 狀態，獨立稽核。

## 輸入與工具

- `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 唯讀來源 `Orion2.exe.i64` SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；稽核使用暫存副本。
- IDA Pro 9.4／IDAPython，`metapc`，DOS/4GW 映像的 IDA 線性位址。
- 可重生匯出：`tools/ida/audit_choose_multi_net_game_ui.py`。匯出保存原始名稱、位址、bytes、
  caller；不修改 `.i64` 名稱。

## 已證實

1. `sub_F0C8E @ 0xF0C8E..0xF0E17` 是清單外層，caller 為
   `sub_FAEDB @ 0xFAFBC`；它直接呼叫以下 loader、field builder 與 draw。
2. `sub_F40D3 @ 0xF40D3..0xF41AD` 是 loader，caller `0xF0D10`。它載入資產
   `0x29 @ 0xF411F`、字型 `0x101 @ 0xF413E`，並以 `0x280／0x1E0／0x51／0x25`
   在 `0xF4162..0xF419E` 計算面板位置；`[window+0x1E8]=0`、`+0x1E9=1` 初始化脈動狀態。
3. `sub_EFF87 @ 0xEFF87..0xF009A` 是 field builder，caller `0xF0D28`。十列欄位使用
   起點 `0x40 @ 0xEFFC5`、x 偏移 `0x26／0x190 @ 0xEFFDA／0xEFFF9`、高度
   `0x16 @ 0xEFFFE`、列距 `0x1B @ 0xF000E`；按鈕偏移為
   `0xBF／0x158 @ 0xF005D／0xF0075`。
4. `sub_F1AF4 @ 0xF1AF4..0xF1CE6` 是 draw，caller `0xF0DE5`。它讀取選中索引
   `[+0x101]` 與欄位陣列 `[+0xA7]`，並在 `0xF1C4D..0xF1C9C` 讓 `[+0x1E8]`
   於 `-3..4` 間依 `[+0x1E9]` 方向來回；逐列高度／步距仍是 `0x16／0x1B`。
5. `sub_F5777 @ 0xF5777..0xF5883` 是對局改名流程，caller `sub_F4760 @ 0xF4784`。
   `edx=8 @ 0xF580F` 傳入輸入 helper，後續 `strcmp @ 0xF5835` 逐一比對既有名稱，重名時
   顯示錯誤後重新輸入。現行 remake 主機名稱彈窗已使用同一 `netplay.GameNameMax=8`；清單本身
   只負責選擇／加入，不在此頁改名。

## remake 對映與證據等級

- **已證實**：479×384 面板定位公式、十列 362×22 熱區、27px 列距、底部按鈕座標、
  選中亮度狀態、名稱上限 8 與重名檢查流程。
- **強推論**：現行 row 內的名稱／位址／人數分欄，是依原版列熱區與 draw 起點切出的安全區；
  原版各欄精確寬度沒有從本輪指令中獨立證實。
- **remake 轉接設計**：UDP LAN discovery 取代 IPX 服務公告；直接輸入 TCP 位址是原版沒有的
  額外入口。它必須放在十列清單外並明確標註，不得冒稱原版 widget。
- **未知／非阻塞**：原版底部按鈕資產的精確寬高仍是量圖值；此項不影響已證實的左上座標。

## 驗收邊界

Go 只保存 UI 語意鍵與動態對局資料；固定雙語文案及三個畫廊示範對局名由 `ui.json` 供應。
標題、三個 row 欄、空清單兩列、直接位址、取消與錯誤列都必須有雙軸安全框。英文有正版面板時
保留 `JOIN NETWORK GAME SETUP` 烘字；繁中與缺面板 fallback 才重繪標題。這不證明 UDP discovery
等同原版 IPX 協定內部。
