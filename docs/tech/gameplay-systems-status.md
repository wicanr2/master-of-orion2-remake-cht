# Gameplay 系統忠實化盤點(對齊原版進度)

> 日期:2026-07-10。目的:逐系統標「已用 gamedata 真公式(忠實)/自編近似(待接)」,給下輪乾淨接手。
> 原則:gamedata/ 的公式多已逐字轉寫自 openorion2 + 手冊(有測試),是**真值來源**;忠實化=把 shell/engine 的自編邏輯換成呼叫這些真公式。

## 已忠實(用 gamedata 真公式,可信)

| 系統 | 位置 | 用到的真公式 |
|---|---|---|
| 殖民地食物/產能/研究產出 | `engine/colony.go` `RunColonyTurn` | `MoraleProductionOutput` |
| 污染清理 | `engine/colony.go` `colonyPollution` | `PollutionTolerance/Eighths/PollutingProduction/CleanupCost` |
| 人口成長 | `engine/colony.go` `colonyGrowth` | `ColonyBaseGrowth/HousingBonus/ColonyGrowth` |
| 稅收 | `engine/empire.go` `RunEmpireTurn` | `IncomeTaxRevenue` |
| **研究/科技(2026-07-10 完成)** | `engine/research.go` + `shell/research.go` + `cmd/moo2/researchchoice.go` | `researchChoices`(真成本+選項)+ 抉擇 UI + 元件解鎖對映(見 `research-system-status.md`)|

→ **殖民地經濟核心已大致忠實**(比 `HONEST-STATUS.md` 舊描述好)。剩收入細項未接(見下)。

## 自編近似 / 未接真公式(待忠實化,依影響排序)

### 1. 戰鬥系統(最大缺口)
- 現況:`shell/session.go` `ResolveBattle` 是**抽象「戰力相減」**(加總雙方 power,依差額 `applyDamage`),格子戰術戰鬥 `tacticalScreen` 亦為簡化 attack−defense。
- gamedata **已備妥完整真公式**(未接):
  - 命中:`CombatHitThreshold`、`CombatClassicToHit`、`CombatAlternativeToHit`、射程 `CombatRangeLevel*`/`CombatRangeLevelPenalty`。
  - 傷害:`DamageForHit`(依命中結果算傷)、`DamageApplyDissipation`、`DamageMountAdjustedValue`。
  - 過盾/過甲:`DamageShieldCapacity`、`DamageAfterShield`(硬盾/穿盾)、`DamageApplyArmor`(穿甲)。
  - 球狀傷害:`DamageSphericalRoll`/`ShipRollCount`/`FlyerDestroyed`。
- **接線計畫(下輪)**:
  1. 擴 `CombatShip` 模型:加 `Defense`、`WeaponMinDmg/MaxDmg`、`ShieldReduction`、`ArmorHP`、`SizeClass`(從艦艇設計元件 Value 推)。
  2. 逐發解算改真流程:射程→命中門檻→`CombatClassicToHit`(擲骰;RNG 用既有 eventRand 或戰鬥專用種子,保持可重現)→`DamageForHit`→`DamageAfterShield`→`DamageApplyArmor`→扣структуру HP。
  3. 保留回合上限/UI;每步對 gamedata 測試值核。加 `combat_realformula_test.go`。
  - 驗收:同配置對原版戰鬥結果趨勢一致(命中率隨射程下降、盾/甲吸收、穿甲穿盾生效)。

### 2. 收入細項
- 未接:`TradeGoodsIncome`(需「工業配置到貿易品」模型)、`IncomeFoodSurplusRevenue`(需帝國食物池/運輸)、`IncomeCommandOverflowCost`、`IncomeFreighterMaintenanceCost`、`IncomeApplyGovernmentMoneyBonus`。
- 前置:先建「帝國食物池 + 運輸艦 + 工業分配(建造/貿易品/研究)」模型,才不會 piecemeal 出錯。

### 3. 艦艇設計(空間格)
- 現況:四下拉(武器/裝甲/護盾/特殊)。原版是艦體空間格 + 每元件佔格 + 改造 mod。gamedata 有元件資料,但「空間格」模型未建。

### 4. 其他自編
- `advancePopulation` 的 `popGrowthThreshold=300` 是 remake 調校值(存檔 pop_growth 未能乾淨反推,已在 session.go 標註 provenance)。
- 隨機事件、安塔蘭、外交、間諜、議會:多為簡化,gamedata 有 `spy.go`/`morale.go`/`ground.go` 等可漸進接。

## 建議下輪順序

戰鬥系統(1)影響最大且**真公式已全部備妥**(只差接線 + 擴 CombatShip 模型),是投報最高的下一個忠實化目標,比照研究系統的做法逐步落地 + 測試。
