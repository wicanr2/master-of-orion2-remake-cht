# 多人主設定畫面玩家文案稽核（2026-08-27）

## 問題與範圍

`cmd/moo2/multiplayer.go` 已使用原版 `MULTIGM.LBX` 與已追回座標，但按鈕表、標題、
TCP／熱座說明、錯誤訊息及轉場名稱仍保存中英文句子。本輪只處理主設定畫面；區網選局
`choosemultinetgame.go` 與席位等候室 `choosenetplyrs.go` 各有不同生命週期，保留為後續切片。

## 輸入與工具

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫唯讀來源：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；實際稽核使用暫存副本。
- 工具：IDA Pro 9.4／IDAPython，處理器 `metapc`，位址皆為 DOS/4GW 映像的 IDA 線性位址。
- 可重生匯出：`tools/ida/audit_multiplayer_setup_ui.py`。匯出保留原始函式名、位址、指令 bytes、
  caller 與 `MULTIGM.LBX` 交叉參照；沒有覆寫 `.i64` 名稱。

## 已證實

1. `sub_F42CA @ 0xF42CA..0xF44BB` 是多人資產／面板初始化鏈；caller 位於
   `sub_F4D99` 的 `0xF4DD9`、`0xF5033`、`0xF5135`。`0xF433E` 載入 `0x280`、
   `0xF4359` 載入 `0x1E0`，與 640×480 置中計算一致。
2. `sub_F009A @ 0xF009A..0xF03F2` 建立按鈕欄位；左右欄立即數可直接回查：
   x 偏移 `0x3B`（`0xF00C7` 等）與 `0x10D`（`0xF02B3` 等），列偏移
   `0x5B／0x7A／0x9B／0xBB`，取消列 y 偏移 `0x11E @ 0xF0240`。
3. `sub_F4D99` 的主要函式範圍為 `0xF4D99..0xF4F74`，IDA 另把後段區塊歸入同一 raw owner；
   唯一直接 caller 是 `sub_1049B @ 0x1061B`。因此本文件不把主要 range 誤寫成整個生命週期的
   唯一連續位址範圍。
4. `sub_F5691 @ 0xF5691..0xF5777` 由 `sub_F009A @ 0xF03E8` 呼叫，並把模式寫入
   `byte_199F3A`：`3 @ 0xF56C2／0xF56F5`、`1 @ 0xF572E`、`2 @ 0xF5756`。
   `sub_F4D99 @ 0xF5110` 會比較該 byte 是否為 1。
5. `MULTIGM.LBX` 字串位於 `0x178004` 與 `0x17A055`；後者有多人模組內大量 loader
   交叉參照，包含 `sub_F42CA` 內的 `0xF42F9..0xF4494`。

## remake 對映與證據等級

- **已證實**：面板置中、按鈕欄位座標、原版模式碼與 `MULTIGM.LBX` 資產來源。
- **強推論**：現行按鈕安全框以實際資產寬高或已證實 fallback 熱區內縮 3px，與原版 widget
  幾何一致；精確字墨 clipping 並非原版程式證據。
- **remake 轉接設計**：NETWORK 以 TCP 取代 IPX，熱座按鈕循環真人席位數，面板下方顯示
  TCP／熱座說明，並以現代錯誤文字回報 host／join 失敗。這些文字沒有冒稱原版字串。
- **停止線**：MODEM、NULL MODEM、COMM INFO 與 TEN 依使用者規則維持停用；不逆向 Win95／
  數據機 API 內部。玩家只需要看見其停用狀態與現代替代路徑。

## 驗收邊界

完成條件是主設定畫面的 Go 程式只保存語意鍵，雙語文案由 `assets/i18n/ui.json` 提供；標題、
按鈕、說明及錯誤列均有雙軸文字安全框，最長雙語模板不裁切，並以正常畫廊抽查。這不會把
現代 TCP 轉接介面升格為原版 IPX 行為逐位元還原。
