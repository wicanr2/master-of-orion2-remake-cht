# `Net_Next_Turn` 玩家文案與版面證據稽核（2026-08-27）

## 問題

確認等待其他網路玩家的畫面是否已有足夠原版證據，可在不改變鎖步規則的前提下，
將固定玩家文案移出 Go，並釐清哪些版面仍只能標為 remake 近似。

## 輸入與工具

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 資料庫：原始 `.i64` 的一次性副本，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4、IDAPython；位址空間為 IDA DOS/4GW 線性位址
- 匯出器：`tools/ida/audit_net_next_turn_ui.py`

## 非破壞性函式索引

| 附加語意 | 原始定位與名稱 | 邊界 | 證據等級 |
|---|---|---|---|
| 回合等待外層 | `sub_FC470 @ 0xFC470` | `0xFC470..0xFC6A5` | 已證實 |
| 面板 loader | `sub_F3E42 @ 0xF3E42` | `0xF3E42..0xF3FC6` | 已證實 |
| 輸入／玩家欄位 builder | `sub_EFCEA @ 0xEFCEA` | `0xEFCEA..0xEFE7A` | 已證實 |
| 畫面 renderer | `sub_F1075 @ 0xF1075` | `0xF1075..0xF166E` | 已證實 |
| 玩家配色 helper | `sub_F31BB @ 0xF31BB` | `0xF31BB..0xF33AE` | 已證實 |
| 聊天輸入迴圈 | `sub_F55A4 @ 0xF55A4` | `0xF55A4..0xF5681` | 已證實 |

原始名稱、位址、指令 bytes、caller 與資料參照均保留於本輪 JSON 匯出；附加語意不覆蓋
IDA 資料庫名稱。

## 結論

- **已證實**：loader 依 `MULTIGM.LBX#42/#43/#40` 尺寸置中三塊面板；builder 使用
  `+0xBB` 計算輸入列位置、高 `0x11`，玩家列迴圈步距為 `0x19`。
- **已證實**：renderer 讀聊天記錄計數欄 `+0x47C`，speaker `8` 走 GNN 分支；
  `sub_F55A4` 直接呼叫 renderer 並以全域輸入 buffer 是否為空控制送出。
- **未證實**：玩家列第一列的精確 y 錨點仍藏於 window／runtime 欄位資料流；現行
  `nntRowFirst=104` 只能維持既有 remake 版面近似。
- **未證實**：原版字串內容尚未在本輪匯出成可對照 catalog；繁中及現代狀態指紋／
  分岔警告只可標為等義介面轉接。
- **過期斷言**：`netNextTurnDemo` 註解稱「連線流程 UI 還沒做」已被正式
  `networkWaitScreen`、共同快照與兩階段鎖步流程推翻；該畫面目前只是畫廊 adapter，
  不是多人功能尚未接線的證據。

