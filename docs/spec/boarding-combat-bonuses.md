# 登艦戰加成垂直規格

證據來源：`docs/re/boarding-combat-bonuses-audit-20260825.md`。

## Typed 輸入

每個快速戰鬥 `combatant` 與戰術戰鬥 `CombatShip` 在進場時固定帶入：

- 所屬帝國的陸戰隊 Strength 與 HitsToKill。
- 該艦艦員等級對應的 Boarding bonus（Bo）。
- 同 owner、有效參戰艦之已指派軍官的最高 Commando bonus。
- 同範圍最高 Security bonus；只由守方消費。
- 保安站旗標；只由守方加 `20`。

不得在解算器內假設攻守雙方屬於玩家，也不得從帝國全域未指派領袖取值。

## 解算

```text
attackStrength = marineStrength(attacker owner)
               + attacker crew Bo
               + attacker fleet max Commando

defenseStrength = marineStrength(defender owner)
                + defender crew Bo
                + defender fleet max Commando
                + defender fleet max Security
                + securityStations * 20
```

兩方 HitsToKill 分別由自己的 High-G 與 Powered Armor 狀態推導。`ResolveBoarding` 仍只接收已整理好的純資料，維持決定性與可測試性。

## 路徑與驗收

- 快速結算：`mkPlayerCombatantsIndexed`、`aiShipCombatant` → `quickBoardingAttempt`。
- 格子戰術：`StartCombat`、`aiTacticalShips` → `ShipBoardingAttack`。
- 玩家與 AI 的種族／科技必須能產生不同 Strength/HitsToKill。
- 艦員 Bo 必須同時影響攻方與守方。
- 多位軍官只取參戰艦隊最高值，不相加；未指派或支援艦軍官不生效。
- 舊存檔缺少軍官指派時安全回退為 0，不新增持久欄位。

