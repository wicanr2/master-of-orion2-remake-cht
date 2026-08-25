# 武器微型化分類與 Hyper 重複等級 IDA 稽核（2026-08-24）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA database：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4（SDK 940），位址為 IDA linear。
- 非破壞性探針：`tools/ida/audit_weapon_mini_categories_hyper.py`；保留 raw name、位址、bytes、
  caller 與欄位 xref，不改名、不套型別、不寫回正式 database。

## 已證實

1. `sub_5E1E3 @ 0x5E1E3` 會依原版資料表把科技 runtime record `+6` 寫成類型：武器為 3、
   船體特殊裝置為 9；因此該 byte 的靜態零值不是 runtime 值。
2. `sub_6F11C @ 0x6F11C` 依科技類型及武器 record `+0x0F` 選出 `sub_6E60E` 的三種 raw category：
   - category 0：beam、標準飛彈與炸彈；
   - category 1：魚雷及 raw special weapon；
   - category 2：Assault Shuttles、Bomber Bays、Fighter Bays、Heavy Fighter Bays。
   船體特殊裝置通常為 category 1；Augmented Engines（tech 20）、Reinforced Hull（56）、
   Extended Fuel Tanks（63）、Heavy Armor（82）、Troop Pods（192）固定為 category 2。
3. `sub_6E60E @ 0x6E60E` 的佔格千分比階梯：
   - category 0：`1000/800/650/500/350/250`；
   - category 1：`1000/800/700/600/500/400`；
   - category 2：任何等級均為 `1000`。
   前五格由 jump table `0x6E5E6`／`0x6E5FA` 的 raw targets 證實，超過第四級走 default；
   公式為 `(base*perMille+500)/1000`，正值結果保底 1。
4. `sub_6B519 @ 0x6B519` 的成本階梯不收 category，仍統一使用
   `100/75/55/40/30/25%`。
5. `sub_6D048 @ 0x6D048` 對一般主題沿 `word_17D90C` 後繼鏈計數；遇到 topic `75..82`
   不再讀一般完成狀態，而是把玩家 record `player*0xEA9 + 0x21C + (topic-75)` 的 byte
   直接加進等級。這是八個研究領域各自的 Hyper-Advanced 重複等級，不是單一布林值。

## 強推論與邊界

- Go 的 `OrigWeaponTable.Cat` 與原版 weapon record `+0x0F` 已逐表對齊，可作為 raw weapon
  kind 的 typed 表示；依它選 category 不需再以中文名稱猜測。
- `Choose_Hyper_Advanced_Tech_ @ 0xFC734` 的 1..20 隨機表只證實 AI 選八種 Hyper tech 的
  權重，不是玩家重複研究的成本公式。後續成本已另由 `Player_Research_Cost_ @ 0xE1E96`
  閉合為 `profile base + completed level×10000`，見
  [`hyper-repeated-cost-audit-20260824.md`](hyper-repeated-cost-audit-20260824.md)。
