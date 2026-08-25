# 戰術逐槽點防自動開火規格

## 證據邊界

- 已證實：原版戰鬥 runtime 保留八個 11-byte 武器槽；點防尚未開火時會在飛彈／戰機接觸前
  自動開火，紅色關閉不會禁用這個例外。
- 強推論：同槽多門武器依 WorkingCount 各自開火一次。
- 未知：原版多槽的完整亂數消費與早停順序；remake 固定依槽序、門數順序處理。

## 資料契約

1. CombatShip.PointDefenseSpentSlots 與 WeaponMounts 等長，只存在本場戰鬥。
2. WeaponMounts 為空時，回退單槽 WeaponName／Mods／PointDefenseSpent。
3. typed 槽的 Name 必須是光束、Mods 必須含 PD、WorkingCount 大於零才能自動開火。
4. WeaponModes 只控制玩家主動齊射；自動 PD 不讀取它，因此紅色 Off 仍會迎擊。

## 解算契約

1. 依槽序找本回合未開火的 PD 槽，每槽依 WorkingCount 開火。
2. 飛彈攔截依次累加 DestroyedWarheads，並將 RemainingInterceptionDamage 傳給下一門 PD。
3. 戰機攔截每門獲得獨立命中擲骰；中隊已全滅時可早停。
4. 一個槽觸發後標記本回合已使用；回合交界只清使用標記，不清攔截傷害餘數。
5. 快速結算套用同款逐槽規則，但沒有 WeaponModes。

## 驗收

- PD 只在第二槽仍可攔截飛彈與戰機。
- 第二槽為 TacticalWeaponOff 仍自動開火，且模式不被改寫。
- 已觸發的 PD 槽在同回合不重複開火；下回合可再開火。
- 多門 PD 合併攔截結果，攔截餘數不丟失。
- 舊單槽測試亂數序與結果不變。

