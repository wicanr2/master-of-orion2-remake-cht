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

→ **殖民地經濟核心已大致忠實**(比 `HONEST-STATUS.md` 舊描述好)。收入細項已接稅收/餘糧/貿易品
三項(見下 §2),剩指揮/運輸艦/政府加成三項未接。

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
- **仍待**:①球狀傷害/飛彈/戰機未接(地面戰已於 §1a/1c 接線完成);②護盾減傷精確 per-class 真值待逆向(現為階梯推導,
  **2026-07-11 提示**:`ship-design-space.md` §1 在手冊 p.121 表格額外挖到 Armor/Struct./Shield 三欄可能就是缺的
  ArmorHP/StructureHP/shipSize 查表,尚未核實接線,留給本項);③per-ship 攻防/傷害為 remake 由艦艇設計推導
  (空間格模型已完成,見 §3;精確值仍需軍官技能模型)。

### 1a. ★ 地面戰:已解算(2026-07-10 更新——推翻本節下方舊「故不做」結論)

> **本節下方原判定「地面戰需逆向、硬編=違反鐵律、故不做」已被使用者 directive 推翻並解決。**
> 使用者定案:手冊無 MOO2 解算式 → **沿用一代(1oom)`game_ground_kill` 公式**(d100+force 對決,明確無歧義)+ 二代手冊加成表/hits-to-kill。這**不是硬編臆造**,是有權威來源(1oom GPL 重製碼,逐位元組對齊原版)的忠實移植,符合鐵律。
> 已實作:`internal/gamedata/ground_battle.go` `ResolveGroundBattle` + 確定性測試(force 高方勝率 0.96、雙倍兵力 0.92、對稱 ~0.49、無死迴圈)。詳見 `ground-combat-algorithm.md`「解算式定案」。
> **2026-07-11 更新**:shell 層「模型 + 流程」接線已完成(見 §1c),**仍待**只剩 UI 繪製/操作介面(不碰 interactive.go,屬後續 task)。

### 1b. 飛彈/球狀傷害:仍需「演算法逆向」(2026-07-10 盤點;地面戰已移出,見 §1a/1c)
- **飛彈**:gamedata `missile.go` 有 jam/AMR 命中/速度,但飛彈**飛行回合、點防攔截互動**的完整解算同樣超出手冊文字,需逆向。
- **結論**:飛彈同屬**需逆向演算法的新子系統**,不是本輪「接 gamedata 真公式」那種可安全自驅的工作。硬編自製解算=違反不臆造鐵律,故不做;列為需 RE(動態 dump/反編/社群反推)的獨立任務。beam 戰鬥(命中/傷害/過盾/過甲)因手冊有 Classic Chance to Hit + Damage 公式且已轉寫進 gamedata,才能安全接線(已完成);地面戰因使用者 directive 定案沿用一代公式,同樣已安全接線(見 §1a),兩者都**不**屬於本節「仍需 RE」的範圍。
- **~~艦艇空間格~~ 已移出本節(2026-07-11)**:原本把「艦艇空間格」也歸類成「需逆向演算法」是誤判——真正原因是先前只查過 `original_game/…CD Manual.pdf`(掃描圖,抽字 0 字元)與 `MANUAL_150.html`(1.50 異動摘要,非完整手冊),沒注意到 `moo2_patch1.5/GAME_MANUAL.pdf` 是**可正常抽字的 188 頁完整文字版手冊**,Ship Design 章節(p.119-132)有完整的艦體空間表 + 武器佔格表,不需要任何逆向工程。詳見 `ship-design-space.md`。

### 1c. ★ 地面戰 shell 層接線:已完成(2026-07-11)
- `internal/shell/ground_invasion.go`:陸戰隊生成(`advanceMarines`,接 `EndTurn`)→ 載運(`LoadMarines`,運力=艦數×手冊每艘 4 個單位的近似,無獨立運輸艦船體類別,標簡化)→ 入侵解算(`GameSession.InvadeColony`,組雙方 `gamedata.GroundForce` 接 `ResolveGroundBattle`,rng 依回合+星索引種子化可重現)→ 勝則星 Owner 轉移 + 殖民地過戶(AI 端移除)。
- Force 計算重用既有 `ComponentUnlocked`/`ArmorOptions` 元件解鎖判定推導裝甲科技加成,避免地面戰科技狀態與造艦科技狀態不同步;種族加成僅套用手冊有明確數字的 Bulrathi/Gnolam。
- 簡化項(標記待精修,不臆造):運輸艦運力近似、AI 守方兵力用「已運作 s.Turn 回合」近似(AI 無 ColonyBuildings 追蹤)、AI 側不套種族加成(AIOpponent 無 RaceIndex)、入侵後保留人口以「守方存活戰鬥單位數」近似(手冊無精確公式)、可入侵範圍僅限 AI 開局母星(`aiExpand` 佔領的星未建殖民地模型)。
- 測試:`ground_invasion_test.go`(強攻方/強守方勝率、前置條件檢查、可重現性、Marine Barracks 成長上限、載運上限)。
- 詳細設計/簡化清單見 `ground-combat-algorithm.md`「2026-07-11 shell 層接線」一節。**仍待**:UI 繪製/操作介面(不碰 interactive.go)。
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
- **已接(2026-07-11)**:`TradeGoodsIncome`——貿易品是「建造佇列選項」(與 Housing 同類),不需要
  「第四種職務配置」這個原判斷是誤判。`internal/shell/session.go` 建造選單新增 `TradeGoodsBuildName`
  (「貿易品」,恆可選、Cost=0,同「不建造」不累積建造進度);`engine.ColonyState` 新增
  `TradeGoods bool`,`syncTradeGoodsFlag` 依建造選單同步;`engine.RunEmpireTurn` 對該旗標為真的殖民地,
  以其 `NetIndustry` 呼叫 `gamedata.TradeGoodsIncome`(一般種族 2:1)累加進新欄位
  `EmpireOutput.TradeGoodsRevenue`,計入 `NetBC`。`fantasticTrader` 固定傳 `false`(同
  `IncomeFoodSurplusRevenue` 既有 TODO,待種族特質系統補上種族欄位)。`IncomeFoodSurplusRevenue`
  同樣已接(見 `colony-economy-maintenance.md` §6.2,不需要「帝國食物池/運輸艦」模型,只需正
  `FoodSurplus` 即可,原判斷同樣過度前置)。
- 仍未接:`IncomeCommandOverflowCost`、`IncomeFreighterMaintenanceCost`(需追蹤運輸艦數量)、
  `IncomeApplyGovernmentMoneyBonus`(需政府形式系統)。

### 3. 艦艇設計(空間格)
- **(2026-07-11)shell/gamedata 層已完成**:`internal/gamedata/shipspace.go` 建了艦體總空間表(`ShipHullSpace`,手冊 p.121 確認值)+ 武器佔格表(`WeaponSpaceByName`,手冊 p.124 確認值);`internal/shell/session.go` 的 `ShipDesignSpaceUsed`/`ShipDesignFits` 接進四下拉模型驗證設計是否超格。細節、估計值標註(特殊系統佔格手冊無數字,5% 估計)、與「裝甲/護盾不佔空間」的手冊澄清見 `ship-design-space.md`。**仍待**:武器改裝(mod)對佔格的影響(手冊已有公式,未接線)、Design Dock 畫面 UI 繪製(不碰 `interactive.go`,歸後續 task)。

### 4. 其他自編
- `advancePopulation` 的 `popGrowthThreshold=300` 是 remake 調校值(存檔 pop_growth 未能乾淨反推,已在 session.go 標註 provenance)。
- 隨機事件、安塔蘭、外交、間諜、議會:多為簡化,gamedata 有 `spy.go`/`morale.go`/`ground.go` 等可漸進接。

## 建議下輪順序

戰鬥系統(1)影響最大且**真公式已全部備妥**(只差接線 + 擴 CombatShip 模型),是投報最高的下一個忠實化目標,比照研究系統的做法逐步落地 + 測試。

## ★ task 16 核心 gameplay 執行順序(2026-07-10,主代理判斷,使用者授權自主排序)

依「對玩家體驗影響 × 有權威來源可自驅」排序:
1. **殖民地建築全表(進行中)**:5 棟 → 手冊 40 棟入 `gamedata/buildings.go`,綁前置科技 gating(subagent 實作中)。
2. **產出行星驅動**:`FoodPerFarmer`/`IndustryPerWorker` 現為固定值,改依 climate/gravity/mineral(手冊 yield 表)推導——讓不同行星經濟有別(MOO2 核心手感)。
3. **貿易財收入接線:已完成(2026-07-11)**——建造選單新增「貿易品」選項 + `engine.RunEmpireTurn`
   接上 `TradeGoodsIncome`,見 §2。
4. **地面戰 UI 入侵流程**:模型 + 流程 shell 層已接線完成(§1c,2026-07-11);剩 UI 繪製/操作介面。
5. **艦艇設計(空間格)**:shell/gamedata 層已完成(2026-07-11,見 §3);UI 繪製留後續。

每塊:手冊/openorion2/一代為權威 → 派 subagent 實作 → 主代理核實 diff/測試才 commit。飛彈/球狀傷害(§1b)仍需 RE,獨立處理。
