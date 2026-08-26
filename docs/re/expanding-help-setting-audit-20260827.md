# Expanding Help 設定稽核（2026-08-27）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；工具為 IDA Pro 9.4／IDAPython，位址均為 DOS/4GW 映像的 IDA linear address。
- 可重生匯出：`tools/ida/audit_game_menu_popup_ui.py`。匯出保留 raw 函式名、位址、指令 bytes 與交叉參照；本文件才附加語意與證據等級。

## 已證實

- 設定 byte 是 `byte_199BE0 @ 0x199BE0`。`sub_127E1 @ 0x127E1` 在 `0x12810` 寫入預設 0；`sub_7EFEF @ 0x7EFEF` 在 `0x7F094` 讀入選項；`sub_7F14C @ 0x7F14C` 在 `0x7F17C` 回寫。
- `sub_83669 @ 0x83669` 在 `0x836D3` 檢查該 byte：非零時於 `0x83704` 呼叫 `sub_83EFD`，為零時走立即重繪鏈。
- `sub_83EFD @ 0x83EFD..0x84356` 是展開呈現器。`0x84140..0x8428C` 以 0..9 共十步，將來源位置與目的矩形的差除以 10 後逐步插值；每步呼叫 `sub_136728` 繪製，最後在 `0x84308..0x8432F` 畫完整目的矩形。
- 另一路 `sub_C702E @ 0xC702E` 在 `0xC71EF` 檢查同一 byte，條件成立才於 `0xC7262` 呼叫相同展開呈現器。這證實它是多個說明／資訊面板共用的視覺選項，不是單一 SETTINGS 畫面資料。

## 強推論與停止線

- `sub_83EFD` 的完整參數型別與每個 caller 的畫面語意尚未全部命名；但「設定只選擇十步展開或立即顯示」已由同一設定 byte、分支與共同 renderer 閉合，足以實作玩家可見契約。
- remake 不逐行翻譯舊繪圖 driver；以固定 10 tick 的矩形插值呈現，標為介面轉接近似。說明內容仍由 JSON 語系檔提供，不從函式名臆造原版文字。
