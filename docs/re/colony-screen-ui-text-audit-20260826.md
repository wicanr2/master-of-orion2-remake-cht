# 殖民地主畫面 UI 與文字證據稽核（2026-08-26）

## 輸入與工具

- `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- `Orion2.exe.i64` SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- IDA Pro 9.4／IDAPython 3.12，處理器 `metapc`；位址為 DOS/4GW 映像的
  IDA linear address。
- 非破壞性匯出：`tools/ida/audit_colony_screen_ui.py`。正式輸入唯讀掛載，腳本
  保留 raw `sub_*` 名稱、函式邊界、caller、bytes 與原始運算元，不寫回資料庫。

## 已證實

1. `sub_BCB4B @ 0xBCB4B..0xBCBA0` 由 `sub_BCC3D` 的 `0xBCC4C`
   直接呼叫。其指令：
   - `0xBCB50 6bdb1e`：列索引乘 `0x1E`（30）；
   - `0xBCB7A 83c33e`：加 `0x3E`（62）；
   - `0xBCB7D b936010000`：載入 x=310；
   - `0xBCB82 68fe010000`：壓入 x=510。
   因此三個職業欄的 x=`310..510`、y=`62+30i` 有直接指令證據。
2. `sub_BCC3D @ 0xBCC3D..0xBCC6D` 由 `sub_BF456` 呼叫；其迴圈 caller
   與既有三列畫面結構一致。
3. `sub_BED21 @ 0xBED21..0xBEF4D` 有兩個 caller，均位於
   `sub_C058A @ 0xC058A..0xC0965`；`0xC07F1`、`0xC083B` 直接呼叫
   `sub_BED21`。`sub_C058A` 又由主流程 `sub_1049B` 與 `sub_FE02C` 呼叫，
   支持完整 Colony Screen → Draw Colony Screen 鏈。
4. `COLPUPS.LBX#5` 的 640×480 框架、三個上方面板與既有座標量測，和上述
   職業欄指令互相獨立吻合；它是版面資產證據，不單獨證明動態文字內容。

## 符號衝突勘誤

`func_names.txt` 把 `Colony_Screen_` 放在 `0xC0965`。IDA 顯示該處是另一個
`sub_C0965 @ 0xC0965..0xC097A`，沒有直接 caller；完整且實際呼叫繪製函式的是
`sub_C058A @ 0xC058A..0xC0965`。本輪以 raw 邊界與 caller 為證據，保留兩個
原始位址，不用外部語意名覆蓋資料庫。

同類偏移也出現在 `Load_Colony_Screen_Seg_`（`func_names.txt` 為 `0xBB9FD`，
`symbols_fixed.tsv` 為 `0xBB954`）。本輪沒有使用該 loader 推導玩家行為，因此只記錄
衝突，不替它判定精確語意。

## remake 映射與證據邊界

- `cmd/moo2/colonyscreen.go` 的職業熱區與三列文字框沿用 x=`310..510`、
  y=`62+30i`；CHANGE／BUY／LEADERS／RETURN 仍依 `COLPUPS.LBX#5` 框架量測。
- 產出、同化、建築清單及「點此 +1」是 remake 動態資訊層；本輪只外部化文案並
  約束安全框，不把這些組句宣稱為原版逐字重現。
- 中段地表與建築繪製另有既有 RE；本輪不更動地形、建築格點或生產規則。
