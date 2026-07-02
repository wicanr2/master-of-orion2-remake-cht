# 唯讀衍生公式與查表(移植紀錄)

> 來源:openorion2 `src/gamestate.cpp`(GPL v2)。這些是「從存檔資料算出顯示值、不改變狀態」的規則公式,
> 是 gameplay 重建(Phase 5)的數值基準。Go:`internal/gamedata/formulas.go`(special 以 bool 傳入,與存檔 bitfield 解耦)。

## 查表

| 表 | 索引 | 值 |
|---|---|---|
| computerHPTable | 艦級 size 0–5 | `1, 2, 5, 7, 10, 20` |
| driveHPTable | 艦級 size 0–5 | `2, 5, 10, 15, 20, 40` |
| computerBonusTable | 電腦型別 0–5 | `0, 25, 50, 75, 100, 125` |
| mineralProductionTable | 礦產豐度 0–4(ULTRA_POOR…ULTRA_RICH) | `1, 2, 3, 5, 8` |

## 公式

| 函式 | 定義 |
|---|---|
| `PlanetBaseProduction(minerals)` | `mineralProductionTable[minerals]` |
| `MaxComputerHP(size)` | `computerHPTable[size]` |
| `ComputerHP(size, compDamage)` | `max(0, MaxComputerHP − compDamage)` |
| `MaxDriveHP(size, reinforcedHull)` | `driveHPTable[size]`,強化船殼(SPEC_REINFORCED_HULL)×3 |
| `DriveHP(size, driveDamage%, reinforcedHull)` | `driveDamage≥100 → 0`;否則 `MaxDriveHP × (100−driveDamage) / 100` |
| `CombatSpeed(base, size, driveDamage, augEngines, reinforcedHull, transDim)` | `base` (+5 若 SPEC_AUGMENTED_ENGINES);以引擎 HP 計算損傷懲罰:`minHP = 2·maxHP/3`;`hp>minHP → ret·(hp−minHP)/(maxHP−minHP)`,否則 0(引擎損傷 >33% 於戰鬥中失去動力);最後 transdim +4 |
| `ComputerBonus(computerType)` | `computerBonusTable[computerType]` |
| `BeamOffense(computerType, computerWorking, battleScanner)` | 電腦未全毀 → `+ComputerBonus`;SPEC_BATTLE_SCANNER → `+50` |
| `BeamDefense(combatSpeed, inertialNullifier, inertialStabilizer)` | `combatSpeed·5`;SPEC_INERTIAL_NULLIFIER +100;SPEC_INERTIAL_STABILIZER +50 |
| `LeaderHireCost(skillValue, expLevel, modifier)` | `max(0, 10·skillValue·(expLevel+1) + modifier)` |

## 驗證(單元測試已涵蓋)

`PlanetBaseProduction(ULTRA_RICH)=8`、`ComputerHP(5,3)=17`、`MaxDriveHP(3,強化)=45`、`DriveHP(3,50%)=7`、
`CombatSpeed(base4,size0,transdim)=8`、`BeamOffense(電腦4,正常)=100`、`BeamDefense(4,慣性裝備)=170`、`LeaderHireCost(10,2,0)=300`。

## 待補

- `Player::researchCost`:依 `research_choices` 資料表(來自 LBX),非純公式,待 LBX 資料表載入後補。
- special-device 的 hasWorkingSpecial 位元語意在 save 層解位元後以 bool 餵入本層(bitfield = specials 已裝 AND NOT 已損)。
