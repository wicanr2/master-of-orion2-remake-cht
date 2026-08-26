# 建造佇列 UI 與文字證據稽核（2026-08-26）

## 輸入與工具

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython 3.12，`metapc`；位址皆為 DOS/4GW 映像的
  IDA linear address。
- 非破壞性匯出腳本：`tools/ida/audit_build_queue_ui.py`。腳本只讀正式資料庫副本，
  保留 raw `sub_*` 名稱、函式邊界、指令 bytes 與 caller，不寫回 `.i64`。

## 已證實

1. `sub_B325A @ 0xB325A..0xB32B4` 由 `sub_B3E7A @ 0xB3E7A` 呼叫；
   `0xB3263` 的 `bf49010000`、`0xB3268` 的 `c745fc5e010000`、
   `0xB327C` 的 `bbca010000` 與 `0xB328E` 的 `b8cf000000` 分別載入
   `329`、`350`、`458`、`207`，支持七格佇列的 x/y 錨點。
2. `sub_B08CA @ 0xB08CA..0xB094C` 由 `sub_B3E7A` 的 `0xB4033` 呼叫；
   `0xB08D9`／`0xB0925` 載入 x=13，`0xB0912` 載入 x=184，支持左側
   可建清單的水平範圍。
3. `sub_B3CF7 @ 0xB3CF7..0xB3E75` 有兩個直接 caller，皆位於
   `sub_B4041 @ 0xB4041..0xB432D`；後者又由主畫面跳表 case 25 呼叫。
   這條 caller 鏈支持 `Build_Queue_Popup_`／`Draw_Build_Queue_Popup_` 的畫面用途。
4. 官方手冊第 70–71 頁支持七格 BUILD QUEUE、AUTO BUILD、REFIT、
   REPEAT BUILD 與後者排除 Housing／Trade Goods；規則邊界詳見
   `docs/tech/colony-production-controls.md`。

## 符號衝突勘誤

`func_names.txt` 把 `Add_Buildings_Fields_` 放在 `0xB094C`，但 IDA 顯示它是
另一個完整相鄰函式 `sub_B094C @ 0xB094C..0xB09CE`，且 `sub_B3E7A` 先後呼叫
`sub_B08CA` 與 `sub_B094C`。目前只能證實建造清單使用的 x=13..184 立即數位於
`sub_B08CA`；`sub_B094C` 的精確玩家語意維持未知，不以外部名稱覆蓋 raw 位址。

`func_names.txt` 另把 `Draw_Build_Queue_Popup_` 放在 `0xB3E75`。IDA 顯示該處只是
`sub_B3E75 @ 0xB3E75..0xB3E7A` 的五位元組呼叫 thunk，且有 21 個 caller；真正被
`sub_B4041` 呼叫兩次、具有完整繪製函式邊界的是 `sub_B3CF7`。本輪文件與程式註解
保留 `0xB3CF7`，並把兩份外部符號表的衝突明列為勘誤。

## remake 映射與未知

- `cmd/moo2/buildqueue.go` 保留 x=13..184 清單、x=207..458 七格佇列及既有
  COLBLDG 六按鈕座標。
- **remake 近似**：AUTO BUILD 的固定優先序與目前可重複 Special 集合不是本輪
  IDA 證實的原版「最佳」判斷；顯示文案必須明示其近似性。
- **未知**：`sub_B094C` 的精確欄位用途、原版所有狀態訊息組句，以及缺少
  `COLBLDG.LBX` 時的原版行為。本輪只提供現代安全退路，不宣稱原版 parity。
