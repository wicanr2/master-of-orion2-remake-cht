# 飛彈彈架容量 IDA 稽核（2026-08-24）

## 問題與舊狀態

原版武器表中的標準飛彈 `Size=0`、`Ammo=5`；remake 卻把四種標準飛彈一律估成 10 佔格，
且 `Ship` 沒有彈藥欄，格子戰術可無限跨回合發射。需要確認 2／5／10／15／20 是什麼、
如何影響成本／佔格，以及設計記錄是否保存它。

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA database：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4（SDK 940）；位址皆為 IDA linear。
- 非破壞性探針：`tools/ida/audit_missile_ammo_racks.py`；不改名、不套型別、不寫回 database。

## 已證實

1. `sub_6EFEB` 判定 weapon type raw `1` 時，`Weapon_Cost_ @ 0x6EC74` 與
   `Weapon_Space_ @ 0x6EE8E` 都對參數值 `2/5/10/15/20` 查表；其增加量依序為
   `10/20/30/35/40`。其他值增加 0。
2. 兩函式都先把增加量加到武器表 base，再套改造與微型化；`Weapon_Cost_` 最後再乘武器數量，
   `Weapon_Space_` 亦在最後乘數量並保底 1。
3. `One_Weapon_Space_ @ 0x6EDC6` 對標準飛彈在 weapon base size 之外再加一個微型化後的 10，
   即最小 2 發彈架的單武器佔格。
4. 設計 caller `sub_68E78 @ 0x68E78` 從同一設計 record 讀：武器 ID `+0x35+2i`、數量
   `+0x45+2i`、Ammo `+0x55+i`、Mods `+0x5D+2i`、Arc `+0xAD+i`，再傳給成本／佔格函式。
   這與 repo 的原版 `.GAM` parser `save.ShipWeapon{Type,MaxCount,WorkingCount,Arc,Mods,Ammo}`
   形成獨立交叉驗證。
5. 原版武器表標準飛彈 `Ammo=5`，故舊存檔缺欄時的相容預設應為 5；魚雷本身 `Ammo=2`，
   不進 raw type 1 的可變彈架分支。

## 強推論與邊界

- `Ammo` 是一場戰鬥可發射的輪數，由欄名、設計 record、原版表與戰鬥語意共同支持；本輪 remake
  以每次實際飛彈／魚雷開火扣 1。尚未逐指令追回原版 combat runtime 的 decrement 寫址，故
  「戰術每發扣一」列為強推論，不冒稱逐位元已證實。
- remake 一艘船只有一個聚合武器槽、數量隱含為 1；本輪不虛構原版八槽與逐槽 WorkingCount。

