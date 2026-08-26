# 對局內遊戲選單 IDA 與文案稽核（2026-08-26）

## 輸入與位址契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4，`metapc`；本文位址為 IDA DOS/4GW image 線性位址。
- 非破壞性匯出器：`tools/ida/audit_game_menu_popup_ui.py`；匯出保留 raw `sub_*`、邊界、caller、bytes、指令與資料參照。

## 已證實

- `sub_8012F @ 0x8012F..0x803D0` 是完整 Game Popup 外層。它呼叫
  `sub_7EA5C` 載入圖片，並依四路 switch 分流到 `sub_7DD41`主選單、
  `sub_7E00F`設定、`sub_7DA76`載入與 `sub_7E154`儲存。
- `sub_7DD41 @ 0x7DD41..0x7E00F` 是對局內主選單輸入迴圈，它在
  `0x7DFBB` 呼叫 `sub_7F701`繪製，並在音量欄命中時分別呼叫
  `sub_80892`與 `sub_80918`。
- `sub_7F701 @ 0x7F701..0x7FA28` 是主選單繪製函式；直接 caller 包含
  `sub_7DD41`、`sub_7ED66`、`sub_7EDF2` 與 `sub_8012F`。
- `sub_7EA5C @ 0x7EA5C..0x7ED4D` 從 `GAME.LBX` 依序載入資產 `0..7`；
  前八個寫入連續指標槽，對應背景、六顆按鈕與音量條。後續仍載入
  `0x1D`、`9..`等儲存／設定共用圖片，不可把整支函式誤寫為「只載主選單八張」。
- `sub_7E00F @ 0x7E00F..0x7E154` 與 `sub_7FA28 @ 0x7FA28..0x8011F` 證實原版
  SETTINGS 是 Game Popup 內的實際設定分頁，不是單一遷移連線開關。
- 外部符號所列 `Print_Options_To_Bitmap_ @ 0x7EDB1` 在目前 IDA 資料庫不是
  函式邊界；相鄰 `sub_7ED66` 精確結束於 `0x7EDB1`。本輪不以符號名強行建函式。

## 強推論與 remake 邊界

- openorion2 `MainMenuWindow::initWidgets` 給出視窗 `(144,25)`、六顆按鈕相對
  座標及 `GAME.LBX#0..6`資產對應；它與 IDA loader `#0..7` 連續鏈獨立吻合，
  列為強推論。
- 按鈕英文烘在圖上；現行繁中譯文、缺資產 fallback 英文與遇錯訊息均是
  remake 文案，不冒稱原版逐字資源。
- 現行 SETTINGS 按鈕在同一視窗展開「遷移連線」是已標註的 remake 擴充。
  原版 `sub_7FA28` 的完整設定欄位、輸入與回寫仍需獨立垂直切片，不屬於本輪文案搬遷。

