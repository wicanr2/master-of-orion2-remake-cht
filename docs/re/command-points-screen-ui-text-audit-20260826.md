# 指揮點數視窗 IDA 與文案稽核（2026-08-26）

## 輸入與位址契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4，`metapc`；本文位址為 IDA DOS/4GW image 線性位址。
- 非破壞性匯出器：`tools/ida/audit_command_points_screen_ui.py`。它保留 raw `sub_*`、函式邊界、caller、bytes、指令與資料參照，不改寫 IDA 內名稱。

## 已證實

- `sub_8BAB9 @ 0x8BAB9..0x8BB39` 是畫面外層。`0x8BAD5..0x8BAEA` 建立 `0..639 × 0..479` 欄位並傳入 ESC 字串；`0x8BAF9..0x8BB09` 清除同一區域，`0x8BB0E` 重畫 `sub_84E9D`，`0x8BB24` 直接呼叫 `sub_E2644`。
- `sub_E2644 @ 0xE2644..0xE2671` 才是外層直接呼叫的指揮點數摘要 wrapper。它在 `0xE265C` 呼叫 `sub_E2000`組內容，並在 `0xE266C` 跳到 `loc_DDF24`顯示。舊文件把外層寫成直接呼叫 `sub_E2000`，已精緻修正。
- `sub_E2000 @ 0xE2000..0xE2644` 是格式化內容的大型函式；它的直接 caller 包含 `sub_E2644 @ 0xE265C` 與 `sub_E2710 @ 0xE2A56`。
- `0xDDF24` 不是獨立函式入口，而是 `sub_DDEFB @ 0xDDEFB..0xDDF2C` 內呼叫 `sub_A5EB2` 的尾端位置。因此只保留 raw `loc_DDF24`，不再將它命名為「泛用訊息視窗函式」。

## 強推論、remake 選擇與未知

- 執行檔符號 `_starting_command_points_msg`、`_total_command_points_msg`、`_total_command_point(s)_used_msg` 與 `_command_summary_msg` 支持目前欄位組成；現行中英文譯文不是原版逐字證據，列為 remake 文案。
- 面板 `(150,130,340,236)` 與「淨餘／超額懲罰」是 remake 可讀性擴充，不冒稱原版逐像素布局。懲罰值使用已有手冊與引擎證據的每點 10 BC 契約。
- `sub_E2000` 內的完整原版格式化字串與行列尚未逐項解出；現有證據已足以實作玩家可見摘要，不為逐字相同繼續擴張 RE。
