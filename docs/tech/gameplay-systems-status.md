# Gameplay 系統忠實化盤點(對齊原版進度)

> 日期:2026-07-10。目的:逐系統標「已用 gamedata 真公式(忠實)/自編近似(待接)」,給下輪乾淨接手。
> 原則:gamedata/ 的公式多已逐字轉寫自 openorion2 + 手冊(有測試),是**真值來源**;忠實化=把 shell/engine 的自編邏輯換成呼叫這些真公式。

## 2026-08-11 收尾勘誤

本文件保留歷史盤點；本輪已完成的 remake 消費端以新狀態覆蓋舊 TODO：CMBTSHP 移動後固定
tick timer、事件／爆炸戰略損傷消費、SABOTAGE 的結構化 AB/DB 分數與 Agent 數量扣除、
以及領袖 ETA 歸零時的殖民地計算 callback 近似均已接入。原版逐幀 timer、raw event score／
爆炸 record 與完整 callback 仍是 oracle 差異，不再把它們寫成「remake 漏接」。
程式／測試／證據邊界見 [`docs/re/remake-consumer-closure-20260811.md`](../re/remake-consumer-closure-20260811.md)。

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
  (`shieldReduceByName`),兩路徑套用,`DamageAfterShield` 護盾機制真正生效。
  ⚠ **2026-08-08(第 62 項(護盾減傷))訂正**:這裡原本寫「依護盾階 0/2/4/6/8/10」——那是清單索引 × 2,
  **五級裡有四級偏高**。手冊值是 **0/1/3/5/7/10**(`gamedata.DamageShieldReductionClass*`,
  已改由 `ShieldReductionForTech` 依科技查)。
- **仍待**:①球狀／飛彈的部分精細路徑與敵方逐艦戰機設計；玩家戰機出擊／返航／最弱護盾面已接(地面戰已於 §1a/1c 接線完成);~~②護盾減傷精確 per-class 真值待逆向~~
  (⚠ **這一項作廢**:真值**不需要逆向**,手冊常數從抄進來就一直躺在 `gamedata/damage.go`,只是沒人用。
  剩下的提示仍有效——
  **2026-07-11 提示**:`ship-design-space.md` §1 在手冊 p.121 表格額外挖到 Armor/Struct./Shield 三欄可能就是缺的
  ArmorHP/StructureHP/shipSize 查表,尚未核實接線,留給本項);③per-ship 攻防/傷害為 remake 由艦艇設計推導
  (空間格模型已完成,見 §3;精確值仍需軍官技能模型)。

### 1a. ★ 地面戰:已解算(2026-07-10 更新——推翻本節下方舊「故不做」結論)

> **本節下方原判定「地面戰需逆向、硬編=違反鐵律、故不做」已被使用者 directive 推翻並解決。**
> 歷史相容方案:手冊無 MOO2 解算式時曾沿用一代(1oom)`game_ground_kill` 公式；2026-08-07 已由 IDA 靜態追回原版四類型／平手雙擊路徑，live path 改見下一行。這段歷史方案仍保留作 `ResolveGroundBattle` 對照，不代表目前入侵流程使用它。
> 已實作:`internal/gamedata/ground_battle_orig.go` `ResolveGroundCombatOrig`（IDA 靜態追回 `Ground_Combat_Round_ @ 0xEC4FE`）並接入 `InvadeColony`；`ResolveGroundBattle` 保留作一代相容對照。確定性測試涵蓋平手雙方受擊、逐類型切換與回合終止；DOSBox 實機傷亡／亂數序列仍是低優先校準。詳見 `ground-combat-algorithm.md`。
> **2026-08-11 更新**:shell 層「模型 + 流程」與 `COLGCBT.LBX` 畫面抽樣均已接線(見 §1c)；仍待的是 DOSBox 實機傷亡／亂數序列、AI 守方裝甲營與入侵後人口校準，屬低優先 runtime oracle。

### 1b. 飛彈/球狀傷害:仍需「演算法逆向」(2026-07-10 盤點;地面戰已移出,見 §1a/1c)
- **飛彈**:gamedata `missile.go` 有 jam/AMR 命中/速度,但飛彈**飛行回合、點防攔截互動**的完整解算同樣超出手冊文字,需逆向。
- **結論**:飛彈同屬**需逆向演算法的新子系統**,不是本輪「接 gamedata 真公式」那種可安全自驅的工作。硬編自製解算=違反不臆造鐵律,故不做;列為需 RE(動態 dump/反編/社群反推)的獨立任務。beam 戰鬥(命中/傷害/過盾/過甲)因手冊有 Classic Chance to Hit + Damage 公式且已轉寫進 gamedata,才能安全接線(已完成);地面戰因使用者 directive 定案沿用一代公式,同樣已安全接線(見 §1a),兩者都**不**屬於本節「仍需 RE」的範圍。
- **~~艦艇空間格~~ 已移出本節(2026-07-11)**:原本把「艦艇空間格」也歸類成「需逆向演算法」是誤判——真正原因是先前只查過 `original_game/…CD Manual.pdf`(掃描圖,抽字 0 字元)與 `MANUAL_150.html`(1.50 異動摘要,非完整手冊),沒注意到 `moo2_patch1.5/GAME_MANUAL.pdf` 是**可正常抽字的 188 頁完整文字版手冊**,Ship Design 章節(p.119-132)有完整的艦體空間表 + 武器佔格表,不需要任何逆向工程。詳見 `ship-design-space.md`。

### 1c. ★ 地面戰 shell 層接線:已完成(2026-07-11)
- `internal/shell/ground_invasion.go`:陸戰隊生成(`advanceMarines`,接 `EndTurn`)→ 載運(`LoadMarines`,運力=艦數×手冊每艘 4 個單位的近似,無獨立運輸艦船體類別,標簡化)→ 入侵解算(`GameSession.InvadeColony`,組雙方 `gamedata.GroundSide` 接 `ResolveGroundCombatOrig`,rng 依回合+星索引種子化可重現)→ 勝則星 Owner 轉移 + 殖民地過戶(AI 端移除)。
- Force 計算重用既有 `ComponentUnlocked`/`ArmorOptions` 元件解鎖判定推導裝甲科技加成,避免地面戰科技狀態與造艦科技狀態不同步;~~種族加成僅套用手冊有明確數字的 Bulrathi/Gnolam~~ → **2026-08-08(第 65 項(種族特性31格))改由 `gamedata.OrigRaceTraits` 全 13 族驅動**;同時修掉諾蘭姆低重力被扣兩次的 bug,並補上薩克拉的地底(守方 +10)與布拉西的高重力(多挨一下)。
- 簡化項(標記待精修,不臆造):運輸艦運力近似、AI 守方兵力用「已運作 s.Turn 回合」近似(AI 無 ColonyBuildings 追蹤)、AI 側不套種族加成(AIOpponent 無種族欄位——那一整層還不存在,不是漏接)、入侵後保留人口以「守方存活戰鬥單位數」近似(手冊無精確公式)、可入侵範圍僅限 AI 開局母星(`aiExpand` 佔領的星未建殖民地模型)。
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
- ★ 以下三項寫下當時仍是「仍未接」,已於後續輪次全部接線,核實以 code 為準(rulebook 63):
  `IncomeCommandOverflowCost`(指揮評等供需,見 `moo2-formulas-reference.md`)、
  `IncomeApplyGovernmentMoneyBonus`(政府型態已接,見上方本節「已接」條目)、
  `IncomeFreighterMaintenanceCost`(2026-07-11(#4)新增「運輸艦隊」建造選項後,玩家側
  `ActiveFreighters` 可變非 0,維護費隨之生效;AI 對手仍未接同一建造流程,對 AI 恆 0)。

### 3. 艦艇設計(空間格)
- **(2026-08-09)設計／佔格層已接**:`internal/gamedata/shipspace.go` 建了艦體總空間表(`ShipHullSpace`,手冊 p.121 確認值)+ 武器佔格表(`WeaponSpaceByName`,手冊 p.124 確認值);`internal/shell/session.go` 的 `ShipDesignSpaceUsed`/`ShipDesignFits` 接進四下拉模型驗證設計是否超格。武器改裝(mod)與火線角的設計佔格／成本、UI 及保存已接；特殊系統未有手冊精確空間數字的估計值仍見 `ship-design-space.md`。**仍待**:火線角的戰術扇形命中消費端與小型化等級門檻。

### 4. 其他自編
- `advancePopulation` 的 `popGrowthThreshold=300` 是 remake 調校值(存檔 pop_growth 未能乾淨反推,已在 session.go 標註 provenance)。
- 隨機事件、安塔蘭、外交:多為簡化,`morale.go`/`ground.go` 等可漸進接。
- **間諜／外交(2026-08-11 更新)**:`gamedata/spy.go` 的機率公式已接上 STEAL/SABOTAGE/HIDE 可玩迴圈(`TrainSpy`、`PlayerSpyMissions`、`advanceEspionage`)，HIDE 套用 SpyVsSpy +20 並跳過偷科技；SABOTAGE 依原版 70 門檻與 `0x10130A`／`0x145EA` 清除 AI 殖民地的一棟已建建築，候選已改用原版 49 槽成本表（slot 9 跳過、權重讀 production cost +8）。AI → 玩家也依 personality 接上 SABOTAGE，會實際讀玩家建築池；這是 remake policy，不是原版完整 AI 策略。外交畫面已接和平／互不侵犯／同盟、貿易／研究協議、固定 5%／10% 週期納貢、終止、回合收益與存檔，原版政府倍率、神級商人 +50 個百分點與活動 Trader 經驗 bucket 的 tier 1/2 最大加成也已接。remake 端的 SABOTAGE 結構化 AB／DB／E、Agent 訓練／擊殺消費已接；原版 raw 完整分數／特殊槽位、一次性餽贈、特殊貿易 byte table／創造力係數、AI 防守 Agent/政體資料仍未知。詳見 `docs/tech/spy-system.md`、`docs/re/special-trade-sabotage-leader-eta-20260811.md` 與 `docs/RESEARCH-LOG.md`。
- **議會/勝利條件:2026-07-11 已從「完全沒有」接上兩條路徑**(銀河議會選舉 2/3 超級多數、殲滅所有
  對手),沿用先前死碼的 `internal/engine/victory.go` + 新增 `gamedata/council.go`/`shell/council.go`。
  詳見獨立文件 `docs/tech/victory-conditions.md`(手冊逐字引用 + 資料模型限制 + TODO)。

## 建議下輪順序

戰鬥系統(1)影響最大且**真公式已全部備妥**(只差接線 + 擴 CombatShip 模型),是投報最高的下一個忠實化目標,比照研究系統的做法逐步落地 + 測試。

## ★ task 16 核心 gameplay 執行順序(2026-07-10,主代理判斷,使用者授權自主排序)

依「對玩家體驗影響 × 有權威來源可自驅」排序:
1. **殖民地建築全表:✅ 完成**(41 棟,見 `docs/HONEST-STATUS.md`)。
   > 查證方式(不要引用文件,直接查):
   > `grep -n "len(Buildings)" internal/gamedata/buildings_test.go` → `if got := len(Buildings); got != 41`。
2. **產出行星驅動**:`FoodPerFarmer`/`IndustryPerWorker` 現為固定值,改依 climate/gravity/mineral(手冊 yield 表)推導——讓不同行星經濟有別(MOO2 核心手感)。
3. **貿易財收入接線:已完成(2026-07-11)**——建造選單新增「貿易品」選項 + `engine.RunEmpireTurn`
   接上 `TradeGoodsIncome`,見 §2。
4. **地面戰 UI 入侵流程**:模型 + 流程 shell 層已接線完成(§1c,2026-07-11);剩 UI 繪製/操作介面。
5. **艦艇設計(空間格)**:shell/gamedata 層已完成(2026-07-11,見 §3);UI 繪製留後續。

每塊:手冊/openorion2/一代為權威 → 派 subagent 實作 → 主代理核實 diff/測試才 commit。飛彈/球狀傷害(§1b)仍需 RE,獨立處理。
