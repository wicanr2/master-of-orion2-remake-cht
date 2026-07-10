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

### 1. 戰鬥系統(進行中)
- **格子戰術戰鬥已接真公式(2026-07-10)**:`tacticalScreen.fireRound` 逐發用 `shell.ResolveShot`
  (射程等級→射程懲罰→命中門檻→`CombatClassicToHit`→`DamageForHit`→`DamageAfterShield`→`DamageApplyArmor`),
  RNG 依回合種子可重現;`CombatShip` 加 Defense/WeaponMin/Max/ShieldReduction/ArmorHP(remake 由艦艇設計推導,
  精確值待艦體空間格+元件佔格+軍官技能模型)。測試 `combat_formula_test.go`。
- **`ResolveBattle` 快速結算也已接真公式(2026-07-10)**:非互動自動戰鬥同樣逐發走 `ResolveShot`
  (每回合雙方齊射;種族加成入攻擊;RNG 依回合種子可重現);移除死碼 `applyDamage`。
  → **兩條戰鬥解算路徑(格子戰術 + 快速艦隊)現都用真 MOO2 戰鬥公式。**
- **護盾與裝甲已分離(2026-07-10)**:戰鬥時依元件名查表得裝甲 HP(`armorHPByName`)+ 護盾每發減傷
  (`shieldReduceByName`,依護盾階 0/2/4/6/8/10),兩路徑套用,`DamageAfterShield` 護盾機制真正生效。
- **仍待**:①球狀傷害/飛彈/戰機/地面戰未接;②護盾減傷精確 per-class 真值待逆向(現為階梯推導);
  ③per-ship 攻防/傷害為 remake 由艦艇設計推導(精確值需艦體空間格+元件佔格+軍官技能模型)。

### 1a. ★ 地面戰:已解算(2026-07-10 更新——推翻本節下方舊「故不做」結論)

> **本節下方原判定「地面戰需逆向、硬編=違反鐵律、故不做」已被使用者 directive 推翻並解決。**
> 使用者定案:手冊無 MOO2 解算式 → **沿用一代(1oom)`game_ground_kill` 公式**(d100+force 對決,明確無歧義)+ 二代手冊加成表/hits-to-kill。這**不是硬編臆造**,是有權威來源(1oom GPL 重製碼,逐位元組對齊原版)的忠實移植,符合鐵律。
> 已實作:`internal/gamedata/ground_battle.go` `ResolveGroundBattle` + 確定性測試(force 高方勝率 0.96、雙倍兵力 0.92、對稱 ~0.49、無死迴圈)。詳見 `ground-combat-algorithm.md`「解算式定案」。
> **仍待**:UI 入侵流程接線(運輸艦→抵敵殖民地→觸發→轉移殖民地);屬 task 16 後續。

### 1b. 飛彈/球狀傷害:仍需「演算法逆向」(2026-07-10 盤點)
- **地面戰**:gamedata `ground.go` 有完整**加成表**(裝甲/裝備/種族/Low-G/穴居防守 hits-to-kill,手冊 p.15-129 逐條驗證),但**手冊只描述加成、未給解算演算法**(戰力→命中機率的公式)——手冊 6916 段只說「advanced tech gives a better chance of winning」。故忠實 `ResolveGroundBattle` 需先**逆向遊戲內部解算迴圈**(或社群 wiki 反推),不能憑加成表自編機率公式。且需先建入侵流程(運輸艦載陸戰隊、抵敵殖民地觸發)。
- **飛彈**:gamedata `missile.go` 有 jam/AMR 命中/速度,但飛彈**飛行回合、點防攔截互動**的完整解算同樣超出手冊文字,需逆向。
- **結論**:這兩者與「球狀傷害/艦艇空間格」同屬**需逆向演算法的新子系統**,不是本輪「接 gamedata 真公式」那種可安全自驅的工作。硬編自製解算=違反不臆造鐵律,故不做;列為需 RE(動態 dump/反編/社群反推)的獨立任務。beam 戰鬥(命中/傷害/過盾/過甲)因手冊有 Classic Chance to Hit + Damage 公式且已轉寫進 gamedata,才能安全接線(已完成)。
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

## ★ task 16 核心 gameplay 執行順序(2026-07-10,主代理判斷,使用者授權自主排序)

依「對玩家體驗影響 × 有權威來源可自驅」排序:
1. **殖民地建築全表(進行中)**:5 棟 → 手冊 40 棟入 `gamedata/buildings.go`,綁前置科技 gating(subagent 實作中)。
2. **產出行星驅動**:`FoodPerFarmer`/`IndustryPerWorker` 現為固定值,改依 climate/gravity/mineral(手冊 yield 表)推導——讓不同行星經濟有別(MOO2 核心手感)。
3. **貿易財收入接線**:`income.go` 的 2:1 轉換公式已備,接進回合結算(需「工業配置到貿易品」模型 + 建造選單「Trade Goods」選項)。
4. **地面戰 UI 入侵流程**:解算已備(§1a),接運輸艦→入侵→轉移殖民地。
5. **艦艇設計(空間格)**:艦體空間格 + 元件佔格模型(§3);較大,最後。

每塊:手冊/openorion2/一代為權威 → 派 subagent 實作 → 主代理核實 diff/測試才 commit。飛彈/球狀傷害(§1b)仍需 RE,獨立處理。
