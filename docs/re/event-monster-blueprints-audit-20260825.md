# 事件怪物精確藍圖稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4；位址為 IDA linear、DOS/4GW LE object #1。
- 可重生唯讀匯出器：`tools/ida/audit_event_monster_blueprints.py`；本輪輸出為
  `/tmp/event-monster-blueprints-v2.json`。腳本不改名、不套型別、不寫回原始資料庫。

## 已證實

五個 `Load_*_Design` 函式直接填入 99-byte `ShipDesign`。欄位版面由既有存檔解析器與
`Load_Combat_Ship_ @ 0x4954A` 的逐欄複製交叉確認。

| 怪物 | raw type | size | drive | computer | 武器槽（ID×數量；mods） | picture | combat speed | 結構 | 裝甲 |
|---|---:|---:|---:|---:|---|---:|---:|---:|---:|
| Amoeba | 10 | 3 | 2 | 2 | 45×2；0、23×5；0 | 8 | 10 | 50 | 750 |
| Crystal | 11 | 4 | 4 | 5 | 42×1；2、26×5；0 | 9 | 10 | 80 | 2500 |
| Dragon | 12 | 4 | 6 | 5 | 41×20；4、40×1；`0x4000` | 10 | 18 | 80 | 2500 |
| Eel | 13 | 3 | 6 | 4 | 44×2；0 | 11 | 23 | 50 | 1000 |
| Hydra | 14 | 4 | 2 | 2 | 43×5；2 | 12 | 6 | 80 | 1500 |

共同欄位為 type=0、shield=0、speed=1、armor type=0、arc=15。`sub_56726 @ 0x56726`
的 switch 證實 raw 10/14→drive 2、11→4、12/13→6。基礎戰速取
`byte_17FE90[drive*46+size]`；Dragon 再加 4、Eel 再加 8。

`sub_58425 @ 0x58425` 以艦體表 `0x180020 + size*36 + 4` 的 Struct. 值計算結構；五份
藍圖 armor type 都為 0、special 位元不含 Reinforced Hull，因此 size 3/4 分別為 50/80。
`sub_58387 @ 0x58387` 對 raw type 10..14 且 drive 非零直接回傳上表怪物專用裝甲。

## 強推論與停止線

- 武器 ID、數量、arc、mods、ammo 是已證實原始欄位；mods `2`、`4`、`0x4000` 的逐武器
  runtime 語意仍須由攻擊消費端閉合，不能只按名稱猜。
- 本輪把精確設計與雙血池接進 remake；快速戰鬥的命中選目標與特殊武器逐格效果仍是既有
  抽象，不宣稱原版戰術逐指令一致。
- `cx==0` 的展示／縮圖分支使用不同數量與 picture 20..24；正常事件戰鬥傳入非零分支，
  因此展示分支不混入玩法藍圖。
