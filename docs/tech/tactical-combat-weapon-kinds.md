# 戰鬥解算依武器類型分流:beam / missile / spherical

> 日期:2026-07-11。狀態:beam/missile 已接線並有回歸測試;spherical 已提供解算函式,現行武器表無武器掛載(見下)。

## 一、問題:公式早就移植了,但戰鬥解算沒用

`internal/gamedata/missile.go`(飛彈防禦/AMR/彈頭 Beam Defense,逐字移植手冊 patch1.5
`MANUAL_150.html` p117-125)與 `internal/gamedata/damage.go` 的 `DamageSpherical*`(球形武器
傷害,移植自「Notes on Spherical Damage」p126)都**已經有實作、有測試**(`missile_test.go`/
`damage_test.go`)。但實際戰鬥解算——`cmd/moo2/interactive.go` 的 `tacticalScreen.fireRound`
與 `internal/shell/session.go` 的 `battleVolley`——對**所有武器**都呼叫同一個 beam 邏輯
`shell.ResolveShot`(`internal/shell/combat_formula.go`)。結果:飛彈武器(如「麥克萊特飛彈」,
`session.go` `WeaponOptions`)被當成普通光束打,飛彈躲避/AMR 攔截從未生效;`docs/HONEST-STATUS.md`
先前也把這寫成「仍待:飛彈躲避、球狀傷害(需 DOSBox oracle)」——這是**誤植**,已於本輪修正
(rule 63:不留錯誤斷言佔位)。

## 二、武器類型分類(`internal/shell/weapon_kind.go`)

`WeaponKind`(`WeaponKindBeam`/`WeaponKindMissile`/`WeaponKindSpherical`)依 `Component.Name`
(`WeaponOptions`,`session.go`)分類,`weaponKindByName` 只有兩條規則:

| 武器名 | 分類 | 依據 |
|---|---|---|
| 核飛彈 | missile | 對應手冊 Missile/MissileBonus 表(p120)的 `MissileWarheadNuclear`(-10) |
| 麥克萊特飛彈 | missile | 對應同表的 `MissileWarheadMerculite`(+15) |
| 其餘(雷射/質量投射器/中子爆破槍/核融合光束/高斯砲/相位砲/電漿砲/死光/無武裝) | beam | 手冊未列為飛彈或球形武器 |

`WeaponOptions` 目前**沒有任何武器分類到 spherical**。這點特別核對過手冊
「Notes on Spherical Damage > Spherical Weapons」(p126)明列的球形武器清單:

- Pulsar(2-24/size class,6 格半徑)
- Plasma Flux(Eel 專屬,10-40)
- Spatial Compressor(4-32,無視護盾裝甲、只傷結構,4 格射程)
- Engine Explosion(引擎爆炸,5×引擎 HP,Quantum Detonator 三倍)

**沒有「死光」(Death Ray)。** 這點很重要:原本交辦這個任務時的分類提示以「死光/恆星轉換器」
舉例球狀武器,但核對手冊原文後,死光是**一般光束武器**——而且正是 `damage.go`
`DamageForHit`(「Different Min-Max Damage」)兩個 worked example 的出處(`damage_test.go`
`TestDamageForHit` 逐字複現手冊算式),把死光改分類成 spherical 會直接與已核對的手冊數字矛盾。
`WeaponOptions` 裡也沒有恆星轉換器(Stellar Converter)這個元件。**故 spherical 分支目前無任何
實際武器掛載**,`shell.ResolveSphericalShot` 是備妥、已測試、等未來新增球形武器元件時串接的函式,
不是死碼副本。

## 三、beam(不動)

`shell.ResolveShot` 簽章與行為完全不變:射程等級→射程懲罰→命中門檻→
`CombatClassicToHit`→`DamageForHit`→`DamageAfterShield`→`DamageApplyArmor`。回歸測試見
`internal/shell/combat_formula_test.go` 的 `TestResolveShot`(既有)+
`TestBattleVolleyDispatchByWeaponKind` 的 beam 分支(新增,證明極端劣勢 net attack 下 30 個
seed 至少出現一次未命中,行為與改動前一致)。

## 四、missile(新接線:`shell.ResolveMissileShot`)

流程對應手冊「Notes on Missile Defenses > Missile Evasion」(p123)+
「Notes on Anti-Missile Rockets」(p125),兩個階段是**獨立事件**,呼叫端要各擲一顆獨立
1-100(`amrRoll`/`jamRoll`),不能像 beam 那樣共用同一個 roll:

1. **AMR 攔截**(`hasAMR`/`amrRangeSquares`/`amrRoll`):若目標裝有反飛彈火箭,依
   `gamedata.MissileAMRChanceToHit(gamedata.MissileAMRRangeIndex(距離))` 判定整枚飛彈被
   擊落(不進入下一階段、不造成傷害)。**現行 remake 的 `SpecialOptions` 沒有「反飛彈火箭」
   這個可造艦元件**,呼叫端(`fireRound`/`battleVolley`)目前一律傳 `hasAMR=false`——
   這不是漏做,是誠實反映現行元件表沒有 AMR,待補上該元件後改用其解鎖狀態決定。
2. **Jam Chance 躲避**(`defenderEvasionBonus`/`attackerScannerBonus`/`hasECCM`/`jamRoll`):
   `gamedata.MissileJamChance` 算出幹擾機率,`gamedata.MissileDefaultHitChance`(100)減去它
   得命中率。四個組成閃避加成的來源——ECM Jammer/Stabilizer 系列、種族 Ship Defense、艦員
   經驗、統帥(Helmsman)——**現行 remake 的艦艇設計與軍官系統都沒有建模**,呼叫端一律傳 0。
   這退化成手冊本身講的基準情境「若目標無任何閃避能力,預設100%命中」,不是臆造值,只是恰好
   對應現況(無裝備)。
3. **命中後傷害**:手冊只給「listed」固定傷害值(如「Nuclear Missile Damage lowered from
   8 to 6」),沒有給像 beam 命中裕度那樣的內插公式,故命中後傷害直接用武器的
   `WeaponMax`(不套用 beam 專用、需要 net-attack/hit-threshold 的
   `gamedata.DamageForHit`),再走與 beam 相同的 `DamageAfterShield`/`DamageApplyArmor`
   (手冊只有 Shield Piercing/Armor Piercing mod 才豁免護盾/裝甲,飛彈本身未掛任何 mod)。

## 五、spherical(新接線但暫無武器掛載:`shell.ResolveSphericalShot`)

對「艦艇」目標:aggD(`gamedata.DamageSphericalRoll`,呼叫端已對同一 slot 全部武器加總)
穿過 `DamageAfterShield`/`DamageApplyArmor`(一般球形武器未講明豁免護盾/裝甲);
`bypassShieldAndArmor=true` 供 Spatial-Compressor 類武器啟用「全打結構、忽略護盾裝甲」。
夾限「對艦最低傷害 1」。

**已知未搬的一段**(damage.go 原本就標注,這輪沒有新增臆測):手冊「Damage Calculation >
Ships」還有一步「the number of rolls is determined by size class + 1」次
`random(aggD)`、「each re-rolled if the outcome is not 1」的加總才是最終傷害,這個重骰
終止條件手冊描述不足以還原成確定性演算法,`gamedata.DamageSphericalShipRollCount` 的函式
註解已明載不移植。`ResolveSphericalShot` 保守地直接用 aggD 當對艦傷害,不臆造重骰後的加總值。

## 六、仍待實機/DOSBox oracle 定案的兩點(其餘都已有手冊依據)

1. `missile.go` 檔頭記載:手冊「飛彈速度」明列公式(`Speed = BaseSpeed(12) + 2*(FTLlevel-1)
   + FastBonus(4)`)與同一段附表的 Speed 欄數字對不上(逐項少 4),目前以「明列公式」為準,
   附表欄位推測是另一個量(驅動本身速度),需要動態驗證確認。
2. 地面戰傷亡係數校準(與本輪戰鬥類型分流無關,`ground-combat-algorithm.md` 既有 TODO)。

## 七、殘留限制(誠實記錄,不是本輪任務範圍)

- 敵方艦隊(`genEnemyFleet`)只有戰力純量,沒有個別武器設計,`CombatShip.Kind`/
  `combatant.kind` 一律零值 `WeaponKindBeam`——敵艦目前不會用飛彈/球形武器攻擊玩家(既有
  簡化,非本輪引入)。
- 「Beam Defense of Missiles」(手冊 p117-120,`gamedata.MissileBeamDefense`)描述的是
  **其他艦艇用光束/PD 武器射擊「飛行中的飛彈本身」把它打下來**這個獨立子機制,需要「選擇
  攻擊目標是飛彈而非艦艇」這個目前 UI/引擎都沒有的動作(`fireRound` 只支援「選敵艦→射程內我
  艦全部開火」)。這輪沒有新增這個子動作(範圍是「依武器類型分流既有的射→打邏輯」,不是新增
  戰鬥子系統/UI),`MissileBeamDefense` 函式本身已測試、可用,留待之後真的要做「攔截飛彈」
  這個獨立玩法時再串接。
