# Master of Orion II 遊戲公式參考

## 前言

本專案(`master-of-orion2-remake-cht`)是目前已知**世界首個**以 Go/ebiten 重寫的 *Master of Orion II* remake,目標是在保留原版規則的前提下完整重建遊戲邏輯,並提供完整繁體中文化。這份文件把目前已移植進 `internal/gamedata/` 的規則公式集中彙整、附上來源出處與驗證方式,作為後續開發與外部查證的知識庫,也是本專案在規則考據上的產出之一。目前涵蓋 12 個系統。

### 公式的兩個來源

`internal/gamedata/rules-implementation-audit.md`(`docs/tech/rules-implementation-audit.md`)逐系統盤點過 openorion2 (GPL v2,本專案參考的既有 C++ 實作)後確認:**openorion2 是存檔載入與檢視殼,不是回合制遊戲引擎**。全 repo 對 `endTurn/nextTurn/processTurn` 與 `rand()/mt19937` 等隨機數來源零命中——它能正確顯示艦艇戰力、研究進度、軍官技能這類「已知輸入 → 顯示輸出」的**唯讀衍生公式**,但不含任何「回合推進 → 狀態改變」的規則引擎(戰鬥解算、殖民地成長、外交、間諜結算等)。

因此本專案的公式移植分兩路:

1. **openorion2 唯讀表**:艦艇電腦/引擎 HP、戰速、光束攻防、軍官經驗/技能/雇用費、研究樹拓撲(`research_choices[83]`)——這些是 openorion2 `gamestate.cpp`/`tech.cpp` 裡真正在跑、且已被 UI 使用的公式,直接逐字轉寫進 Go。
2. **官方手冊權威公式**:openorion2 沒有實作的系統(殖民地成長、生產/污染、士氣、國庫收入、光束命中、光束傷害解算、飛彈防禦、地面戰、間諜),改以 `moo2_patch1.5/MANUAL_150.html`(1.50 patch 說明書,含 "Notes on Population Growth"、"Notes on Spying"、"Notes on Missile Defenses"、"Notes on Anti-Missile Rockets"、"Notes on Beam Weapon Mechanics"、"Notes on Orbital Assault" 等附錄段落,俗稱「The Algorithm」)與 `moo2_patch1.5/GAME_MANUAL.pdf`(隨 1.50 patch 附的完整遊戲手冊)逐條抽取常數與公式移植。

### 驗證方法

- **手算對照測試**:每條公式在 `internal/gamedata/*_test.go` 都有對應測試,用手冊原文附的 worked example(如光束命中 range 23 sq → hit_threshold 95、間諜 jam chance 87% 對照範例算出 33%)或手算數值逐條斷言,不只測「函式能跑」。
- **交叉驗證腳本**:研究樹 83 個主題的成本與可選科技數,由主代理獨立寫解析腳本抽取 openorion2 `tech.cpp:169-305` 的 C 陣列基準值,與 Go 移植結果逐列比對(`techtree_verify_test.go`)。
- **Opus 逐條核實**:間諜(`spy.go`)、光束命中(`combat.go`)等以手冊為唯一來源、無 openorion2 程式碼可對照的系統,由獨立審查回合逐句核對手冊原文與 Go 常數/公式是否一致。
- **不臆測原則**:手冊未給精確數字或公式的項目(如軍官技能等級對應暗殺機率、Spy vs Spy 判定門檻的精確映射),一律標記待查證並保留範圍常數,不自行外插或臆造——各檔案內以 `TODO 手冊未明列` 註記。

本文件所有常數與公式均以 `internal/gamedata/` 原始碼為準,並附 openorion2 原始碼行號與手冊出處,供後續查核與 1.31/1.50 差異研究使用。

---

## 目錄

1. [殖民地人口成長](#1-殖民地人口成長-colonygo)
2. [生產與污染](#2-生產與污染-productiongo)
3. [士氣](#3-士氣-moralego)
4. [國庫收入](#4-國庫收入-incomego)
5. [研究樹](#5-研究樹-techtreego)
6. [軍官](#6-軍官-officergo)
7. [艦艇衍生值](#7-艦艇衍生值-formulasgo)
8. [光束武器命中](#8-光束武器命中-combatgo)
9. [光束傷害解算](#9-光束傷害解算-damagego)
10. [飛彈防禦與反飛彈火箭](#10-飛彈防禦與反飛彈火箭-missilego)
11. [地面戰](#11-地面戰-groundgo)
12. [間諜](#12-間諜-spygo)
13. [手冊自相矛盾記錄](#13-手冊自相矛盾記錄)
14. [尚未移植/待查證](#14-尚未移植待查證)

---

## 1. 殖民地人口成長(`colony.go`)

**來源**:MOO2 patch 1.5 官方手冊「Notes on Population Growth」(`moo2_patch1.5/MANUAL_150.html`)。openorion2 未實作此邏輯,只從存檔讀 `pop_growth[]` 顯示,無公式可抄,故以手冊為唯一權威來源。

### 手冊變數定義

| 變數 | 意義 |
|---|---|
| `POPRACE` | 該種族在此殖民地的人口 |
| `POPAGG` | 星球總人口(含被併吞人口 annexed pop、原住民 natives、機器人 droids) |
| `POPMAX` | 該星球的人口上限 |
| `PROD` | 殖民地總工業產出 |
| `FACTOR1` | 2000(基礎成長公式係數) |
| `FACTOR2` | 40(住房公式係數) |

人口上限硬性夾在 42(1.50 patch note "Population Capacity Clipped To 42")。

### 公式

**基礎成長率 a**(`ColonyBaseGrowth`):

```
a = trunc[ sqrt( FACTOR1 * POPRACE * (POPMAX - POPAGG) / POPMAX ) ]
```

`popAgg >= popMax`(已滿)或參數非法時回 0;`trunc` 為無條件捨去(Go 用 `int(math.Sqrt(v))` 實現)。

**住房獎金 h**(`ColonyHousingBonus`,僅「住房中 housing」狀態適用):

```
h = FACTOR2 * PROD / POPAGG
```

**最終成長**(`ColonyGrowth`):

```
growth = a * b
b = (100 + g + r + i + t + l + e + h) / 100
```

七項獎金(百分點):`g` 一般獎金(預設 0)、`r` 種族獎金(−50%~+100%,可為負)、`i` AI 難度獎金(0-4)、`t` 科技獎金(microbiotics +25、universal antidote +50)、`l` 軍官獎金(僅取最佳醫療加成)、`e` 事件獎金(繁榮 boom +100、瘟疫 plague −200)、`h` 住房獎金。

### Go 函式對照

| 函式 | 簽名 | 對應公式 |
|---|---|---|
| `ColonyBaseGrowth(popRace, popAgg, popMax int) int` | `colony.go:21` | 基礎成長率 a |
| `ColonyHousingBonus(prod, popAgg int) int` | `colony.go:34` | 住房獎金 h |
| `ColonyGrowth(baseGrowth, bonusSum int) int` | `colony.go:44` | 最終成長 a\*b |

### 驗證範例(`colony_test.go`)

- `ColonyBaseGrowth(5,5,10)`:`2000*5*5/10=5000`,`sqrt(5000)≈70.7` → `70`。
- `ColonyBaseGrowth(10,0,10)`:`2000*10*10/10=20000`,`sqrt≈141.4` → `141`。
- `ColonyHousingBonus(40,10)`:`40*40/10=160`。
- `ColonyGrowth(100,-50)`:種族 −50% 獎金 → `100*(100-50)/100=50`。

---

## 2. 生產與污染(`production.go`)

**來源**:`GAME_MANUAL.pdf`(patch 1.5 隨附完整手冊)「System Overview / Yield / Population」章節(約 p.64-67)與各建築說明章節(約 p.78-90)。`MANUAL_150.html`(1.50 patch 說明書)本身只在 UI bugfix 段落提過一次 pollution,無數值公式;openorion2 只有 `Planet::baseProduction()` 單一查表(見第 7 節),未實作生產分配與污染公式,故本節數值全部來自 `GAME_MANUAL.pdf`。

### 生產常數

| 常數 | 值 | 手冊出處(約頁) | 原文 |
|---|---|---|---|
| `ProdWorkerMinimum` | 1 | p.66 | "Each unit produces at least 1 production, no matter what the situation." |
| `ProdAutomatedFactoryPerWorkerBonus` | +1/工人 | p.78 | "increasing the output of each industrial unit of population by +1 production each turn" |
| `ProdAutomatedFactoryFlatBonus` | +5/殖民地 | p.78 | "giving the colony +5 production" |
| `ProdDeepCoreMinePerWorkerBonus` | +3/工人 | p.82 | "increases the productivity of each worker unit by 3 production" |
| `ProdDeepCoreMineFlatBonus` | +15/殖民地 | p.82 | "and the colony by 15" |
| `ProdRecyclotronPerPopulation` | +1/人口(不計污染) | p.81 | "each unit of population generates 1 industrial production, regardless of its assigned job. This increased production does not count toward the planetary pollution level" |
| `ProdMicroliteConstructionPerWorkerBonus` | +1/工業工人(全帝國) | — | "increases the output of all your empire's industrial workers by 1 production per turn each" |
| `ProdAlienUncooperativeNumerator/Denominator` | 3/4 | p.67 | 未整合外星人口每單位只產出正常值 3/4 |

**機器人工廠(Robotic Factory)** 依礦產豐度給殖民地產能加成(索引與 `mineralProductionTable` 一致):

| 礦產豐度 | ULTRA_POOR | POOR | ABUNDANT | RICH | ULTRA_RICH |
|---|---|---|---|---|---|
| 加成 | +5 | +8 | +10 | +15 | +20 |

`ProdWorkerOutput(base int) int` 套用最低產出下限(`max(base, 1)`);`ProdAlienWorkerOutput(base int) int` 回傳 `base * 3 / 4`(向下取整)。

### 污染公式

**污染容忍值**(`PollutionTolerance`,p.66):

```
tolerance = 2 * size_class   (size_class 為 1-based:Tiny=1 ... Huge=5)
```

手冊範例:"a medium planet (size class 3) has a pollution tolerance of 6 production"(medium=size class 3 → 6,已用 pdftotext 核對 `GAME_MANUAL.pdf` 原文)。Go 的 `PlanetSize` 是 0-based(`TINY_PLANET=0...HUGE_PLANET=4`),故實作為 `2*(int(size)+1)`。奈米分解者(Nano Disassemblers)使容忍值加倍(`PollutionToleranceWithNanoDisassemblers`)。

**清理成本**(`PollutionCleanupCost`,p.66):

```
超出容忍值部分的一半用於清理污染(向下取整);Tolerant 特性種族(含矽晶生物)不受污染影響、免清理。
excess = production - tolerance
cleanup = excess > 0 ? excess / 2 : 0   (tolerantRace 一律 0)
```

**仍會產生污染的產能比例**(`PollutionEighths`,以 8 分之幾表示):

| 建築組合 | 比例 | 手冊出處(p.90) |
|---|---|---|
| 無建築 | 8/8 | — |
| 只有污染處理器(Pollution Processor) | 4/8 | "process the waste from fully half of the colony's production" |
| 只有大氣更新器(Atmospheric Renewer) | 2/8 | "cuts out the pollution produced by three-quarters of the industry"(剩 1/4) |
| 兩者皆有 | 1/8 | "only one-eighth of the industry produces pollution"(手冊直接給組合值,非額外假設的 1/2×1/4) |
| 核心廢料場(Core Waste Dump,p.83) | 0/8 | "eliminates all pollution on the planet"(取代前兩者) |

`PollutionPollutingProduction(production, eighths int) int` = `production * eighths / 8`(向下取整)。

> 手冊只描述「處理器/更新器如何折算致污染產能比例」,未給出這個比例與 `PollutionTolerance`/`PollutionCleanupCost` 兩段式規則合併運算的逐字範例;合併順序(先縮減產能再算容忍值,或反之)留給呼叫端決定,`production.go` 註解已標註此假設。

---

## 3. 士氣(`morale.go`)

**來源**:GAME_MANUAL.pdf p.65-66「Morale is in the top box」與 p.169-170「Morale」章節(兩處原文一致);政府基礎值、首都淪陷懲罰見 p.165-167 Imperial Policy > Government 段與各政府段落 p.21-22。openorion2 未實作 morale 邏輯(`gamestate.cpp` 無對應計算),故以手冊原文數字為唯一權威來源。手冊沒有給精確數字的項目(Spiritual Leader 的 morale 加成、Tactics 技能)不移植,列於本節末 TODO。

### 核心公式

手冊原文:「Each smile represents a 10% bonus to all production (food, industry, research, and income); each frown denotes a 10% penalty to all production.」(p.65-66、p.169-170 一致)

```
最終產出 = 基礎產出 *(100 + moralePercent)/ 100
```

`moralePercent` 為下列各來源加總後的百分點(政府基礎值、首都淪陷懲罰、建築/成就加成、多種族懲罰…);手冊沒有列出互斥反例,故各來源之間視為直接加總。

`MoraleProductionOutput(base, moralePercent int) int`(`morale.go:50`):`base*(100+moralePercent)/100`。

### 政府基礎士氣

| 政府 | 無 Barracks | 有 Barracks | 首都淪陷懲罰 |
|---|---|---|---|
| Feudalism / Confederation | -20% | 0% | -50% |
| Dictatorship | -20% | 0% | -35% |
| Imperium | 0%(-20%+20%帝國加成) | +20%(20%帝國加成) | -35% |
| Democracy / Federation | 0% | 0% | -20% |
| Unification / Galactic Unification | 不適用本公式(見下) | 不適用 | 0%(手冊:「Capture of the capital has no effect」) |

`MoraleGovernmentBase(gov MoraleGovernmentType, hasBarracks bool) int`(`morale.go:85`):Imperium 固定 +20%(手冊:「All colonies have a 20% Morale bonus.」),無 Barracks 時再疊加各政府表列的懲罰;Imperium 的「-20% 無 Barracks 懲罰」與「+20% 帝國加成」是兩個獨立疊加項,故 Imperium 有 Barracks 時淨 +20%、無 Barracks 時淨 0%。

`MoraleCapitalCapturedPenalty(gov MoraleGovernmentType) int`(`morale.go:119`):查 `moraleCapitalCapturedTable`,對照上表右欄。

### 建築/成就加成(固定值)

| 常數 | 值 | 手冊出處 |
|---|---|---|
| `MoraleHoloSimulatorBonus` | +20% | p.95-96「increases a planet's morale by 20%」 |
| `MoralePleasureDomeBonus` | +30% | p.97-98「increases colony morale by 30%」 |
| `MoraleVirtualRealityNetworkBonus` | +20%(全帝國) | p.97-98「increases morale by 20% in every colony throughout the entire empire」 |
| `MoralePsionicsBonus` | +10%(限定政府) | p.100-101,見下 |

`MoralePsionicsGovernmentBonus(gov MoraleGovernmentType) int`(`morale.go:137`):Psionics 成就僅在 Dictatorship/Imperium/Feudalism/Confederation 生效 +10%,Democracy/Federation/Unification 系不適用回 0(手冊:「morale is raised by 10% throughout the empire if your government is a Dictatorship, Imperium, Feudalism, or Confederation.」)。

`MoraleMultiRacialPenalty(hasAlienManagementCenter bool) int`(`morale.go:127`):多種族殖民地無 Alien Management Center 時 -20%,有則 0%(手冊 p.66-67、p.92-93 兩處一致)。

### Unification 特例

手冊:「Things that boost or lower Morale have no effect. The Morale of the race's populations cannot be modified in any way.」Unification 系政府不套用一般士氣公式,改為固定的 **food+industry**(不含 research/income)產出加成:

```
Unification:          +50%
Galactic Unification: +100%
```

`MoraleUnificationProductionBonus(gov MoraleGovernmentType) int`(`morale.go:155`)。**呼叫端不可混用**:一般政府用 `MoraleProductionOutput` + `MoraleGovernmentBase`(四項產出皆適用),Unification 系改用本函式(僅 food/industry)。

### 驗證範例(`morale_test.go`)

- `MoraleProductionOutput(100,20)=120`(2 個笑臉)、`MoraleProductionOutput(100,-35)=65`(Dictatorship 首都淪陷)。
- `MoraleGovernmentBase(Imperium,false)=0`、`MoraleGovernmentBase(Imperium,true)=20`。
- `MoraleUnificationProductionBonus(Unification)=50`、`(GalacticUnification)=100`。

---

## 4. 國庫收入(`income.go`)

**來源**:GAME_MANUAL.pdf p.37(Treasury indicator,稅率範圍/級距/1:1 轉換)、p.168(Taxes 章節,50% 例子與 Trade Goods 2:1 對照)、p.70(Trade Goods 轉換與 Fantastic Trader)、p.25(Fantastic Trader 剩餘糧食換算)、p.169(指揮評等超支費用、運輸艦維護費)、p.170(士氣對產出的影響,與 morale.go 同一手冊段落但套用在收入上)。MANUAL_150.html「Modding with Config → Additional Settings」補充 Democracy/Federation 政府對「money」的加成(換算公式「value*5=百分比」)。

### 稅率

`TaxRateMinPercent=0`、`TaxRateMaxPercent=50`、`TaxRateStepPercent=10`(p.37:「values range from 0% to 50% with increments of 10%」)。`IncomeTaxRateIsValid(taxRatePercent int) bool`(`income.go:103`)檢查是否為合法級距值。

稅收轉換 1:1(p.37:「converted into money 1:1」;p.168 再確認「Taxes also have a better conversion rate (1:1) than Trade Goods (2:1)」):

```
IncomeTaxRevenue(totalIndustry, taxRatePercent)     = totalIndustry * taxRatePercent / 100   (`income.go:114`,無條件捨去)
IncomeTaxRemainingIndustry(totalIndustry, taxRate)  = totalIndustry - IncomeTaxRevenue(...)   (`income.go:121`)
```

手冊 p.168 例子:50% 稅率 →「fully half your production potential goes toward taxes. Only the remaining half is available for building.」

### 貿易財(Trade Goods)

一般種族 2 產能換 1 BC,Fantastic Trader 1 產能換 1 BC(p.70:「Every 2 industry converts to 1 BC... unless you are a Fantastic Trader in which case every 1 industry converts to 1 BC」)。`TradeGoodsIncome(industryAllocated int, fantasticTrader bool) int`(`income.go:128`),無條件捨去。

### 糧食剩餘

一般種族每單位剩餘糧食換 0.5 BC,Fantastic Trader 每單位換 1 BC(p.25:「1 BC (instead of the usual half) for every surplus unit of food generated」)。`IncomeFoodSurplusRevenue(surplusFoodUnits int, fantasticTrader bool) int`(`income.go:139`),無條件捨去。

### 指揮評等超支與運輸艦維護

| 項目 | 費用 | 手冊出處 |
|---|---|---|
| 指揮評等(Command Rating)每點未覆蓋需求 | 每回合 -10 BC | p.169「For each rating point required by a ship that is not covered, 10 BCs come out of your income every turn」 |
| 每艘使用中運輸艦(Freighter) | 每回合 -0.5 BC(無條件捨去) | p.169「each freighter that is in use costs 1/2 BC per turn for maintenance」;未使用中的運輸艦不計費 |

`IncomeCommandOverflowCost(uncoveredCommandPoints int) int`(`income.go:148`)、`IncomeFreighterMaintenanceCost(activeFreighters int) int`(`income.go:157`)。

### 士氣對收入的影響

與 morale.go 同一手冊段落(p.170):「Every morale icon on the Colony screen represents a change of 10% in the total production output of the colony... adding to the food, industry, science, and income of a world.」`IncomeMoraleAdjustedProduction(baseProduction, netMoraleIcons int) int`(`income.go:167`):`baseProduction*(100+netMoraleIcons*10)/100`。與 `morale.go` 的 `MoraleProductionOutput` 是同一條手冊規則的兩個獨立實作(收入與殖民地產出分別呼叫,避免跨檔案耦合),數值邏輯必須保持一致,修改其一時應同步檢查另一邊。

### 政府對 BC 收入的加成

MANUAL_150.html:「value * 5 is the percent bonus for the item, for example democracy has a 10 * 5 = 50% bonus to research.」該節列出 `democracy_money=10`、`federation_money=15`,故:

```
Democracy:  10*5 = 50%
Federation: 15*5 = 75%
```

`IncomeGovtBonusDemocracyMoneyPercent=50`、`IncomeGovtBonusFederationMoneyPercent=75`;`IncomeApplyGovernmentMoneyBonus(baseBC, bonusPercent int) int`(`income.go:175`)。手冊只列出這兩種政府對「money」的加成,其餘政府的 `govt_bonus` 只列 science/food/production,不套用本函式。

### 驗證範例(`income_test.go`)

- `IncomeTaxRevenue(100,50)=50`(p.168 例子)、`IncomeTaxRevenue(33,10)=3`(33*10/100=3.3 捨去)。
- `TradeGoodsIncome(11,false)=5`(11/2=5.5 捨去)、`TradeGoodsIncome(7,true)=7`。
- `IncomeFreighterMaintenanceCost(5)=2`(5*0.5=2.5 捨去)、`IncomeFreighterMaintenanceCost(1)=0`。
- `IncomeApplyGovernmentMoneyBonus(100,50)=150`、`(100,75)=175`。

---

## 5. 研究樹(`techtree.go`)

**來源**:逐字轉寫自 openorion2 `tech.cpp:69-167`(`techtree[8][14]`,8 個研究領域各含哪些主題)與 `tech.cpp:169-305`(`research_choices[83]`,每個主題的花費與可選科技)。常數定義:`gamestate.h:62-65`(`MAX_RESEARCH_AREAS=8`)、`tech.h:27`(`MAX_RESEARCH_CHOICES=4`)、`tech.cpp:40`(`MAX_AREA_TOPICS=14`)。

這是 openorion2 盤點中價值最高的可複用資產:研究樹拓撲資料完整、正確,且不含隨機性或狀態變化,純粹是「靜態資料表」,故直接逐字轉寫,不重新設計。

### 結構

- **8 個研究領域**(`ResearchArea`):Biology / Power / Physics / Construction / Force Fields / Chemistry / Computers / Sociology。
- **83 個研究主題**(`ResearchTopic`),每個領域含最多 14 個主題,依原始 C 陣列順序排列。
- 每個主題有 `Cost`(研究點數)、`ResearchAll`(是否需研究完該主題所有科技才算完成,如 `TOPIC_CHEMISTRY`/`TOPIC_ELECTRONICS`/`TOPIC_ENGINEERING`/`TOPIC_NUCLEAR_FISSION`/`TOPIC_PHYSICS` 五個起始領域為 `true`)、`Choices`(可選的 212 個 `Technology` 之子集)。

### 成本分布(83 主題,節錄代表性檔位)

| Cost | 主題數 | 範例 |
|---|---|---|
| 0 | 2 | `TOPIC_STARTING_TECH`、`TOPIC_XENON_TECHNOLOGY` |
| 50 | 4 | `TOPIC_CHEMISTRY`、`TOPIC_ELECTRONICS`、`TOPIC_ENGINEERING`、`TOPIC_NUCLEAR_FISSION`、`TOPIC_PHYSICS`(起始科技,`ResearchAll=true`) |
| 80–650 | 多數低階科技 | `TOPIC_ADVANCED_ENGINEERING`(80)、`TOPIC_MILITARY_TACTICS`(150)、`TOPIC_ADVANCED_CHEMISTRY`(650) |
| 900–4500 | 中階科技 | `TOPIC_ROBOTICS`(650)、`TOPIC_CYBERTRONICS`(2750)、`TOPIC_ADVANCED_GOVERNMENTS`(4500) |
| 6000–15000 | 高階科技 | `TOPIC_GALACTIC_ECONOMICS`(6000)、`TOPIC_TEMPORAL_FIELDS`/`TOPIC_TEMPORAL_PHYSICS`(15000) |
| 25000 | 8 | 全部 `TOPIC_HYPER_*` 超科技領域(Hyper Biology/Power/Physics/Construction/Fields/Chemistry/Computers/Sociology) |

### Go 函式對照

| 函式 | 用途 |
|---|---|
| `ResearchChoiceFor(topic ResearchTopic) ResearchChoice` | 回傳指定主題的花費與可選科技清單 |
| `TechTree() [][]ResearchTopic` | 回傳 8 個研究領域各自的主題清單(深拷貝) |

### 驗證方式

`techtree_verify_test.go` 內建兩個獨立抽取的 C 基準陣列(`cResearchCosts[83]`、`cResearchNumChoices[83]`,由主代理獨立寫解析腳本從 `tech.cpp:169-305` 產生),逐列比對 Go 移植結果的 `Cost` 與 `len(Choices)`,全 83 列數值相符。

---

## 6. 軍官(`officer.go`)

**來源**:openorion2 `gamestate.cpp:607-701`(`Leader::expLevel/hasSkill/skillBonus/hireCost`),SA1 盤點標記為「唯讀公式,品質最好、可直接複用」的系統。技能 id 編碼定義於 `gamestate.h`:bit4-5 = 技能類型(0 common / 1 captain / 2 admin),bit0-3 = 技能碼。

### 經驗等級

門檻表(`gamestate.cpp:68` `leaderExpThresholds[] = {60,150,300,500,0}`):

| 累積經驗 | < 60 | 60–149 | 150–299 | 300–499 | ≥ 500 |
|---|---|---|---|---|---|
| 等級 | 0 | 1 | 2 | 3 | 4 |

`LeaderExpLevel(experience int) int`(`officer.go:12`,對照 `gamestate.cpp:607-617`)逐一比對門檻回傳等級。

### 技能加成

`baseSkillValues[type][code]`(`gamestate.cpp:75-79`,三種類型各自的基礎值陣列):

| 類型 | 陣列(依 code 0..n) |
|---|---|
| common(type 0) | `2, 2, 10, -60, 10, 2, 5, 2, 2, 10` |
| captain(type 1) | `2, 5, 5, 5, 1, 5, 2, 5` |
| admin(type 2) | `-10, 10, 10, 1, 10, 10, 10, 5, 2` |

一般規則(`LeaderSkillBonus`,`officer.go:46`,對照 `gamestate.cpp:664-692`):

```
tier <= 0            → 0
skillID == Navigator  → navigatorSkillValues[tier>1 ? 1 : 0][expLevel]
其餘:
  base = baseSkillValues[type][code]
  非 Megawealth(0x04) → base *= (expLevel + 1)
  tier > 1(進階技能)  → base += base / 2   (+50%)
```

領航員(Navigator,id `0x14`)不套用通則,改查專屬表 `navigatorSkillValues`:

| tier | exp0 | exp1 | exp2 | exp3 | exp4 |
|---|---|---|---|---|---|
| 一般(tier 1) | 1 | 1 | 2 | 2 | 3 |
| 進階(tier>1) | 1 | 1 | 3 | 3 | 4 |

鉅富(Megawealth,id `0x04`)是唯一不隨經驗等級倍增的技能(`gamestate.cpp:680-683` 特判)。

### 雇用費

`LeaderHireCost(skillValue, expLevel, modifier int) int`(`formulas.go:140`,對照 `gamestate.cpp:700-701`):

```
hireCost = max(0, 10 * skillValue * (expLevel + 1) + modifier)
```

### 驗證範例(`officer_test.go`)

- 暗殺(common code0,base 2)tier1 exp4:`2*(4+1)=10`。
- 暗殺 tier2 exp4(進階 +50%):`10 + 10/2 = 15`。
- 鉅富 tier2 exp3:`10 + 10/2 = 15`(不因 exp3 而先乘 4)。
- 領航 tier2 exp4:查表 `navigatorSkillValues[1][4] = 4`。
- `LeaderHireCost(10,2,0) = 10*10*3 = 300`。

---

## 7. 艦艇衍生值(`formulas.go`)

**來源**:openorion2 `gamestate.cpp`,`ShipDesign` 類別的六個唯讀屬性公式(`computerHP`/`driveHP`/`combatSpeed`/`beamOffense`/`beamDefense`,`gamestate.cpp:841-930`)與艦員加成表(`gamestate.cpp:162-167`)。這些公式驅動艦隊清單畫面的戰力顯示,不是戰鬥解算本身,但戰鬥解算需要這些衍生值當輸入。

### 查表(艦級 size 0-5,`MAX_COMBAT_SHIP_CLASSES=6`)

| 表 | openorion2 位置 | 索引 0 | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|---|---|
| `computerHPTable` | `gamestate.cpp:150` | 1 | 2 | 5 | 7 | 10 | 20 |
| `driveHPTable` | `gamestate.cpp:154` | 2 | 5 | 10 | 15 | 20 | 40 |
| `computerBonusTable`(電腦型別 0-5) | `gamestate.cpp:158` | 0 | 25 | 50 | 75 | 100 | 125 |
| `mineralProductionTable`(礦產豐度 0-4) | `gamestate.cpp:31` | 1 | 2 | 3 | 5 | 8 | — |
| `shipCrewOffenseBonuses`(艦員等級 0-4) | `gamestate.cpp:162` | 0 | 15 | 30 | 50 | 75 | — |
| `shipCrewDefenseBonuses`(艦員等級 0-4) | `gamestate.cpp:166` | 0 | 15 | 30 | 50 | 75 | — |

### 公式

| Go 函式 | openorion2 對照 | 定義 |
|---|---|---|
| `PlanetBaseProduction(minerals int) int` | `gamestate.cpp:518-520` | `mineralProductionTable[minerals]` |
| `MaxComputerHP(size int) int` | `gamestate.cpp:841-843` | `computerHPTable[size]` |
| `ComputerHP(size, compDamage int) int` | `gamestate.cpp:845-849` | `max(0, MaxComputerHP - compDamage)` |
| `MaxDriveHP(size int, reinforcedHull bool) int` | `gamestate.cpp:851-859` | `driveHPTable[size]`,強化船殼(`SPEC_REINFORCED_HULL`)×3 |
| `DriveHP(size, driveDamage int, reinforcedHull bool) int` | `gamestate.cpp:861-869` | `driveDamage≥100 → 0`;否則 `MaxDriveHP × (100−driveDamage) / 100` |
| `CombatSpeed(...)` | `gamestate.cpp:871-901` | 見下 |
| `ComputerBonus(computerType int) int` | `gamestate.cpp:158-161`(表) | `computerBonusTable[computerType]` |
| `BeamOffense(computerType int, computerWorking, battleScanner bool) int` | `gamestate.cpp:903-915` | 電腦未全毀 → `+ComputerBonus`;`SPEC_BATTLE_SCANNER` → `+50` |
| `BeamDefense(combatSpeed int, inertialNullifier, inertialStabilizer bool) int` | `gamestate.cpp:918-930` | `combatSpeed·5`;`SPEC_INERTIAL_NULLIFIER` +100;`SPEC_INERTIAL_STABILIZER` +50 |
| `ShipCrewOffenseBonus(crewLevel int) int` | `gamestate.cpp:1697` | `shipCrewOffenseBonuses[crewLevel]` |
| `ShipCrewDefenseBonus(crewLevel int) int` | `gamestate.cpp:1710` | `shipCrewDefenseBonuses[crewLevel]` |
| `LeaderHireCost(...)` | `gamestate.cpp:700-701` | 見第 6 節 |

**戰鬥移動力**(`CombatSpeed`,`gamestate.cpp:871-901`)是本組公式中最複雜的一條——引擎損傷 >33%(以 HP 換算,非直接百分比)會使艦艇在戰鬥中完全失去動力:

```go
ret := baseCombatSpeed
if augmentedEngines { ret += 5 }              // SPEC_AUGMENTED_ENGINES
maxHP := MaxDriveHP(size, reinforcedHull)
hp    := DriveHP(size, driveDamage, reinforcedHull)
minHP := 2 * maxHP / 3
if minHP < hp {
    hp -= minHP; maxHP -= minHP
    ret = ret * hp / maxHP
} else {
    ret = 0                                    // 引擎 HP 未過 2/3 門檻 → 戰鬥中失去動力
}
if transDimensional { ret += 4 }
```

### 驗證範例(`formulas_test.go`)

`PlanetBaseProduction(ULTRA_RICH)=8`、`ComputerHP(5,3)=17`、`MaxDriveHP(3,強化)=45`、`DriveHP(3,50%)=7`、`CombatSpeed(base4,size0,transdim)=8`、`BeamOffense(電腦4,正常)=100`、`BeamDefense(4,慣性裝備)=170`、`LeaderHireCost(10,2,0)=300`。

---

## 8. 光束武器命中(`combat.go`)

**來源**:MOO2 patch 1.5 官方手冊「Notes on Beam Weapon Mechanics」章節(`MANUAL_150.html`)。openorion2 全 repo 對 `combat`/`Combat` 字串的命中全部是艦隊列表 UI 的分類篩選(`ships.cpp:375,853`),且全 repo 零 RNG 來源,沒有任何命中率或傷害解算函式可對照,故本節數值全部來自手冊,已用手冊原文逐句核對。

### Range Level 換算(格→level)

「1 range unit = 3 squares」是三種掛載類型(Regular/Point Defense/Heavy)共用的基礎換算,`combatRegularRangeLevelRaw(sq) = ceil(sq/3)`(`sq<=0` 回 0)。三種掛載各自的夾限/縮放規則:

| 掛載類型 | Go 函式 | 換算規則 | 手冊依據 |
|---|---|---|---|
| Regular | `CombatRangeLevel(sq)` | `raw(sq)`,夾限於 8 | Range Penalty 表「Regular (sq)」列:`0→0,1-3→1,4-6→2,…,22-24→8` |
| Point Defense | `CombatRangeLevelPointDefense(sq)` | `raw(sq)*2`,夾限於 8 | "Point Defense weapons get a penalty as if range is doubled";PD 列只在偶數 level 有值,證實是 **range level 加倍**而非距離格數加倍再查表 |
| Heavy | `CombatRangeLevelHeavy(sq)` | `raw(sq)/2`,夾限於 8 | "for Heavy mount weapons the actual range is halved (and rounded down)";對照 Hv (sq) 列 `0-3→0,4-9→1,…,46-51→8` 逐一驗證是先算未夾限的 Regular raw level 再整除 2 |

### Range Penalty 表

| Range Level | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|---|
| Penalty | 0 | 0 | 10 | 20 | 30 | 40 | 55 | 70 | 85 |

`CombatRangeLevelPenalty(level int) int`(越界夾限到 0-8 端點)。Classic Fusion Beam、Plasma Cannon、Mauler Device 有內建「2x range to-hit penalty」(手冊:武器 flag #1),`CombatRangeLevelPenaltyDoubled(level, doubled)` 套用此加倍規則(先算完 Hv/Reg/PD 距離換算,再整體乘 2)。

### Hit Threshold 與命中判定

**共用門檻**(`CombatHitThreshold`,手冊 `[3b]`):

```
hit_threshold = min(40 + range_penalty − PD_bonus, 95)
```

**Classic Chance to Hit Formula**(`CombatClassicToHit`,手冊原文三分支):

```
[1] random(100) > 95                          → 必中(5% 幸運一擊,無視其他數值)
[2] BA+CO-AF-BD >= 99                         → 必中
[3a] random(100) + BA+CO-AF-BD >= hit_threshold → 命中
```

`BA+CO-AF-BD` = Beam Attack + 電腦命中加成 − Point Defense 目標 AF − 目標 Beam Defense,由呼叫端算好後以 `netAttack` 傳入。

**1.50 Alternative Chance to Hit Formula(Optional)**(`CombatAlternativeToHit`,`simplified_beam_formula=1` 時啟用):

```
[1] random(100) > 95                                        → 必中
[2] BA+CO-AF-BD − range_penalty + PD_bonus >= 99             → 必中
[3] BA+CO-AF-BD − range_penalty + PD_bonus + random(100) >= 40 → 命中
```

手冊指出 Classic 公式的缺陷是距離懲罰可能被 `[2]` 的高 BA/低 BD 情境完全蓋掉;Alternative 公式讓距離懲罰「一致地」影響命中率。

### 驗證範例(`combat_test.go`,對照手冊 Damage Potential Examples)

- Range 23 sq(level 8,penalty 85,PD 0):`hit_threshold = min(40+85-0, 95) = 95`。
- Range 11 sq(level 4,penalty 30,PD 0):`hit_threshold = min(40+30-0, 95) = 70`。

### 戰機速度與 Beam Defense

**戰機速度**(`CombatFighterSpeed`,手冊 "TransDimensionalBonus is 4 for all fighter types"):

```
Speed = BaseSpeed + 2*(FTLlevel-1) + 4
```

| 戰機類型 | BaseSpeed |
|---|---|
| Interceptor | 10 |
| Assault Shuttle | 6 |
| Bomber | 8 |
| Heavy Fighter | 8 |

**戰機 Beam Defense**(`CombatFighterBeamDefense`):

```
BeamDefense = 5*Speed + RacialShipDefenseBonus + FighterPilotBonus + HelmsmanBonus(僅 1.50)
```

---

## 9. 光束傷害解算(`damage.go`)

**來源**:MANUAL_150.html「Notes on Beam Weapon Mechanics > Damage Potential」(距離衰減表、Hv/PD/HEF 加成公式)與同章節「Different Min-Max Damage」(命中後傷害內插公式);GAME_MANUAL.pdf 的 Shield/Armor/Weapon Mods 附錄(護盾容量、Hard Shields、Armor Piercing)。openorion2 全 repo 無傷害解算邏輯,本節數值全部來自手冊。本節處理「命中後的傷害量」,命中率(to-hit)已在第 8 節(`combat.go`)移植,兩者輸入輸出銜接但不重複定義。

### 距離衰減表(與 to-hit 的 Range Penalty 表是兩張不同的表)

`damage.go` 的距離衰減表對應手冊 Damage Potential 表的「Penalty」列,依 range level 0-8:

| Range Level | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|---|
| Penalty | 0 | 0 | 10 | 20 | 30 | 40 | 50 | 60 | 65 |
| 傷害% | 100% | 100% | 90% | 80% | 70% | 60% | 50% | 40% | 35% |

**注意(容易混淆,務必區分)**:這張表與第 8 節 `combat.go` 的 to-hit「Range Penalty 表」(`0,0,10,20,30,40,55,70,85`)是**兩張不同的表**。level 0-5 數值恰好相同,level 6-8 明顯不同(50/60/65 vs 55/70/85),純屬巧合、不可互相取代。前者(本節)決定「命中後傷害打幾折」,後者(第 8 節)決定「命不命中」。呼叫端須各自用對應的查表函式(`DamageDissipationPenalty` vs `CombatRangeLevelPenalty`);`DamageForHit` 需要同時吃這兩張表的輸出(見下)。

`DamageDissipationPenalty(level int) int`(`damage.go:35`):`damageDissipationPenaltyTable[level]`,超出 0-8 夾限至端點。Range level 沿用第 8 節 `CombatRangeLevel`/`CombatRangeLevelPointDefense`/`CombatRangeLevelHeavy` 算出的同一個 level,兩節共用距離換算,各自查自己的懲罰表。

### Hv / PD / HEF mount 加成(百分點相加,非乘法)

手冊原文:「These bonuses are percentages of damage of a regular mount beam. ... Their interaction is not multiplicative but additive and they work the same way as dissipation range penalty」。

| 常數 | 值 | 手冊出處 |
|---|---|---|
| `DamageMountBonusHeavy` | +50 | Heavy Mount「cause 150% of the normal amount of damage」 |
| `DamageMountPenaltyPointDefense` | -50 | Point Defense「inflict only half the damage of a full-size beam」 |
| `DamageMountBonusHEF` | +50 | High Energy Focus(System)「increasing the damage ... by 50%」 |

```
DamageMountAdjustedValue(base, hvBonus, hefBonus, pdPenalty, rangePenaltyPoints) =
    round( base * (100 + hvBonus + hefBonus - pdPenalty - rangePenaltyPoints) / 100 )
    夾限最小值為 1(手冊:「the minimum damage potential is always 1」)
```

`damage.go:93`,`round` 用 `damageRoundDiv100`(手冊:「Round to nearest applies」)。已用手冊「a 50-100 beam with dissipation Range_penalty of 30」的 5 組共 10 個數字逐一核對(`damage_test.go`)。

`DamageApplyDissipation(minDmg, maxDmg, level int) (int, int)`(`damage.go:114`):只套用距離衰減(呼叫 `DamageMountAdjustedValue` 時 Hv/HEF/PD 全填 0),對應 Phasor/Mauler/Death Ray 三組手冊 worked example,已逐 level(0-8)全數核對。

`NR`(Not Reduced by Range,如 Mass Driver/Gauss/Disrupter)武器手冊明文「no dissipation penalty applies」,呼叫端不應呼叫此函式,直接用基礎傷害。

### 命中後傷害量(`DamageForHit`)

手冊「Different Min-Max Damage」公式:

```
UNCAPPED DAMAGE = min_dmg + (max_dmg-min_dmg+1) * A / B
A = roll_plus_attack - hit_threshold
B = 100 - hit_threshold
roll_plus_attack = min(random(100) + BA+CO-AF-BD, 100)
CAPPED DAMAGE = 上式結果夾限於 max_dmg
```

`DamageForHit(minDmg, maxDmg, roll, netAttack, hitThreshold int) int`(`damage.go:143`)。與第 8 節 `CombatClassicToHit` 的必中分支([1] `random(100)>95`、[2] `netAttack>=99`)共用,兩者皆直接回傳 `maxDmg`。`hitThreshold` 用第 8 節 `CombatHitThreshold` 算好傳入(已夾限 ≤95,故 `B` 恆 >5,不會除以 0)。已用手冊 Death Ray 兩組 worked example(range 23 sq / 11 sq,roll=85、netAttack=10)逐步驗算核對。

### 護盾(Shield)

Class I/III/V/VII/X 護盾(遊戲只有這 5 級,無 II/IV/VI/VIII/IX),每次攻擊減傷 = 等級數字,單一 facing 總容量 = `5 * 等級數字 * 船體 size`:

| 護盾等級 | 每次攻擊減傷 | 單一 facing 總容量(size 倍率) |
|---|---|---|
| Class I | 1 | 5x size |
| Class III | 3 | 15x size |
| Class V | 5 | 25x size |
| Class VII | 7 | 35x size |
| Class X | 10 | 50x size |

`DamageShieldCapacity(shieldReduction, shipSize int) int`(`damage.go:193`)、`DamageAfterShield(dmg, shieldReduction int, hardShield, shieldPiercing bool) int`(`damage.go:210`)。Hard Shields 額外 -3(`DamageHardShieldBonus`)且使 Shield Piercing 失效(手冊:「provide immunity to shield-piercing weapons」);Shield Piercing 對行星護盾不適用(僅限艦對艦,呼叫端另行判斷)。

### 裝甲穿透(Armor Piercing)

手冊:「AP: Armor Piercing beam weapons penetrate any type of armor except Xentronium and Heavy Armor. All of the damage done passes through as if there were no armor at all.」目標裝有 Heavy Armor 或 Xentronium Armor 時 AP 失效,退回一般裝甲/結構分配規則。`DamageApplyArmor(dmg, armorHP int, armorPiercing, apNegated bool) (dmgToArmor, dmgToStructure, remainingArmorHP int)`(`damage.go:238`)。

### 球形武器(僅移植手冊講清楚數字的片段)

- `DamageSphericalRoll(damageMin, roll, ordnancePercent int) int`(`damage.go:260`):`D=(min+roll)*ordnance%`。
- `DamageSphericalShipRollCount(sizeClass CombatShipClass) int`(`damage.go:271`):骰數 = size class + 1(手冊:「a frigate gets one roll, a destroyer two rolls」)。
- `DamageSphericalFlyerDestroyed(aggD, hitPoints, roll int) bool`(`damage.go:277`):`aggD*25/hitPoints >= roll` 則摧毀。
- `DamageEngineExplosionPotential(maxEngineHP int, quantumDetonator bool) int`(`damage.go:289`):`5*maxEngineHP`,Quantum Detonator 使其 x3。

重骰機制(「re-rolled if the outcome is not 1」)與引擎爆炸的逐格衰減率手冊描述不足,不移植,見第 14 節。

### 驗證範例(`damage_test.go`)

- `DamageMountAdjustedValue(50,50,0,0,30)=60`(Hv,base 50,range penalty 30)。
- `DamageApplyDissipation(50,100,4)=(35,70)`(Death Ray,level4)。
- `DamageForHit(...)` 兩組 Death Ray worked example:range 23 sq → 18(=min_dmg);range 11 sq → 65。
- `DamageShieldCapacity(DamageShieldReductionClassVII,4)=140`(35x4)。
- `DamageApplyArmor(30,20,true,false)=(0,30,20)`(AP 全部繞過裝甲)。

---

## 10. 飛彈防禦與反飛彈火箭(`missile.go`)

**來源**:手冊「Notes on Missile Defenses」(Weapons / Special Defensive Systems / Missile Evasion)與「Notes on Anti-Missile Rockets」(Range & Chance to Hit)。openorion2 未實作此段戰鬥判定(只有 tech 名稱字串),本檔是手冊到程式碼的首次移植。

### 特殊防禦裝置(對飛彈與魚雷皆有效)

| 裝置 | 效果 | 值 |
|---|---|---|
| 閃電力場(Lightning Field) | 每枚飛彈/魚雷/戰機各有機率直接摧毀(在 MIRV 分裂前判定) | 50% |
| 匿蹤裝置(Cloaking Device) | 裝置啟動時飛彈彈頭/魚雷有機率未命中 | 50% |
| 位移裝置(Displacement Device) | 不論其他條件,一律有機率完全未命中 | 30% |

魚雷(Torpedoes)不同於飛彈(Missiles):不能被一般光束/PD 光束/攔截機/反飛彈火箭瞄準或傷害。

### 飛彈閃避(Missile Evasion)

無任何閃避能力時,飛彈突破防禦後預設 100% 命中(`MissileDefaultHitChance`)。閃避加成來源(手冊 p123):

| 來源 | 加成 |
|---|---|
| ECM Jammer | 70 |
| MultiWave Jammer | 100 |
| Wide Area Jammer(對自身) | 130 |
| Wide Area Jammer(對艦隊其餘船艦) | 70(不與其他 jammer 疊加) |
| Inertial Stabilizer | 25 |
| Inertial Nullifier | 50 |
| 種族 Ship Defense 特性(三檔) | −20 / +25 / +50 |
| 艦員經驗(Green/Regular/Veteran/Elite/Ultra Elite) | 0 / 7 / 15 / 25 / 37 |
| 統帥(Helmsman)加成 | `helmsmanValue / 2` |

**Jam Chance**(`MissileJamChance`,手冊原文):被幹擾機率 = 防禦方閃避加成 − 攻擊方已知掃描器加成,若飛彈具 ECCM 則此機率減半;逐彈頭獨立判定(故 MIRV 可能只有部分彈頭被擋)。

```
chance = defenderEvasionBonus - attackerScannerBonus
if hasECCM { chance /= 2 }
```

手冊範例驗證:Wide Area Jammer 艦隊加成(70)+ Stabilizer(25)+ 種族懲罰(−20)+ 一般艦員(7)+ 統帥一半(10/2=5)= 87;攻擊方 Tachyon Scanner 已知加成 20;具 ECCM:`(87−20)/2 = 33%`。

### 反飛彈火箭(AMR)

最大射程 15 格(`MissileAMRMaxRangeSquares`)。格→Range 索引換算比標準換算「多位移一格」(手冊:"treating range0 as range1"):

```
MissileAMRRangeIndex(sq) = ceil((sq+1)/3) = (sq+3)/3   (整數除法)
```

| 格距離 | 0-2 | 3-5 | 6-8 | 9-11 | 12-14 | 15-17 |
|---|---|---|---|---|---|---|
| Range | 1 | 2 | 3 | 4 | 5 | 6 |

**AMR 命中率**(`MissileAMRChanceToHit`,見第 13 節手冊矛盾記錄取得的裁決公式):

```
MissileAMRChanceToHit(range) = 71 - (range+2)*10/3   (整數除法)
```

| Range | 0 | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|---|
| Chance-to-Hit | 65% | 61% | 58% | 55% | 51% | 48% | 45% |

命中一次只摧毀彈頭堆疊中的一枚飛彈,且與目標飛彈種類/血量/mods 無關,只與距離有關。

### 飛彈 Beam Defense

`MissileSpeed`/`MissileBeamDefense`(見第 13 節手冊矛盾記錄):

```
Speed = 12 + 2*(FTLlevel-1) + 4
BeamDefense = 5*Speed + MissileBonus(彈頭型別)
```

| 彈頭型別 | MissileBonus |
|---|---|
| Nuclear | −10 |
| Merculite | 15 |
| Pulson | 40 |
| Zeon | 70 |

---

## 11. 地面戰(`ground.go`)

**來源**:GAME_MANUAL.pdf p.15-16(種族 Ground Combat 加成:Bulrathi/Gnolam)、p.21(Combat Modifiers 定義)、p.24(Low-G/High-G/Subterranean)、p.27(Warlord,barracks 容量加倍)、p.77(Marine Barracks)、p.79(Troop Pods/Armor Barracks)、p.80-81(Powered Armor/Battleoids)、p.85(Transport Ship)、p.90-92(Tritanium/Zortrium/Neutronium/Adamantium Armor 對地面戰力加成)、p.108-109(Anti-Grav Harness/Personal Shield)、p.114(Xentronium Armor)、p.162-164(Invading a Colony 流程敘述,無額外數字公式);MANUAL_150.html p.129「Notes on Orbital Assault > Orbital Bombardment」(Estimated Bomb Hits / Planet Hits 表)。openorion2 未實作地面戰鬥判定邏輯(只有存檔欄位與科技/建築名稱),本節為手冊到程式碼的首次移植。沒有精確數字的項目(Commando Leader 基準加成、AI Ground Troops Bonus、Stored Production 命中曲線)一律不臆測,列於本節末 TODO。

### Barracks 建造與人口上限

| 建築 | 初始單位 | 之後速度 | 人口上限 | Warlord 加倍 |
|---|---|---|---|---|
| Marine Barracks | 4(p.77) | 每 5 回合 +1 | `min(現有人口/2, 星球人口上限/2)`(p.77) | x2(p.27) |
| Armor Barracks | 2(p.79) | 每 5 回合 +1 | `min(現有人口/4, 星球人口上限/4)`(p.79) | x2(p.27) |

```
GroundMarineBarracksCap(currentPop, planetMaxPop, warlord) = min(currentPop/2, planetMaxPop/2) [*2 若 warlord]   (`ground.go:64`)
GroundArmorBarracksCap(currentPop, planetMaxPop, warlord)  = min(currentPop/4, planetMaxPop/4) [*2 若 warlord]   (`ground.go:81`)
GroundMarineBarracksUnits(turnsSinceBuilt, ...) = min(4 + turnsSinceBuilt/5, cap)   (`ground.go:97`)
GroundArmorBarracksUnits(turnsSinceBuilt, ...)  = min(2 + turnsSinceBuilt/5, cap)   (`ground.go:111`)
```

Transport Ship 建成時配 4 個 Marine 單位(`GroundTransportShipMarineCapacity=4`,p.85);Troop Pods 使艦上 Marine 數翻倍(`GroundTroopPodsMultiplier=2`,p.79)。

### 單位血量(Hits to Kill)

手冊 p.129 Planet Hits 表基礎值,加上 p.24/p.80/p.81 的修飾條件:

| 單位 | 基礎 hits | High-G(+1) | Powered Armor(+1,僅 Marine) |
|---|---|---|---|
| Marine | 1 | 適用 | 適用 |
| Tank | 2 | 適用 | 不適用(手冊 Tank 列未列 Powered Armor) |
| Battleoid(取代 Tank) | 3(固定,p.81) | 不適用 | 不適用 |

`GroundMarineHitsToKill(highGRace, poweredArmor bool) int`(`ground.go:148`)、`GroundTankHitsToKill(highGRace bool) int`(`ground.go:163`,無 poweredArmor 參數)、`GroundBattleoidHitsToKill=3`(固定值,取代 Tank 後不再套用 `GroundTankHitsToKill`)。

### 地面部隊戰力加成表

裝甲科技(擇一套用「目前已知最佳」一項,不與較低階疊加,手冊逐條各自獨立描述最佳裝甲的加成):

| 科技 | 加成 | 手冊出處 |
|---|---|---|
| Tritanium Armor | +10 | p.90 |
| Zortrium Armor | +15 | p.91 |
| Neutronium Armor | +20 | p.91 |
| Adamantium Armor | +25 | p.92 |
| Xentronium Armor | +30 | p.114 |

`GroundArmorTechBonus(tech Technology) int`(`ground.go:189`)。基礎 Titanium Armor 手冊未提供地面戰力加成,回 0。

裝備科技(可與裝甲加成疊加):

| 科技 | 加成 | 手冊出處 |
|---|---|---|
| Powered Armor | +10(戰鬥評等) | p.80 |
| Anti-Grav Harness | +10(地面戰鬥評等) | p.108 |
| Personal Shield | +20(Marine/Armor 戰鬥評等) | p.109 |

`GroundEquipmentTechBonus(tech Technology) int`(`ground.go:215`)。Battleoids 的 +10(相對 Tank,`GroundBattleoidCombatBonus`)是獨立常數,不走此表。

種族與地形加成:

| 來源 | 加成 | 手冊出處 |
|---|---|---|
| Bulrathi(種族) | +10 | p.15 |
| Gnolam(種族) | -10 | p.16 |
| Low-G(地形,任何 Low-G 種族通用,獨立於種族固定值) | -10%(乘算) | p.24 |
| Subterranean(僅防守己方殖民地) | +10 | p.24 |

`GroundRaceCombatBonus(race GroundRace) int`(`ground.go:244`)、`GroundApplyLowGPenalty(strength int) int`(`ground.go:263`,`strength - strength*10/100`)、`GroundSubterraneanBonus(defending bool) int`(`ground.go:273`)。

### 轟炸(Orbital Bombardment,MANUAL_150.html p.129)

```
GroundBombHitsFromDamage(totalDamage) = min(totalDamage / 100, 320)   (`ground.go:311`)
```

手冊:「All remaining ships fire all weapons 10 times... total damage is calculated from it. This damage is divided by 100 to get the displayed number... The maximum number of bomb hits for the fleet in orbit is 320.」`totalDamage`(含光束/魚雷減半、電腦加成等)由呼叫端算好傳入,本函式只做除以 100 與夾限。行星飛彈規避率固定 7%(`GroundPlanetMissileEvasionPercent`,手冊:「The planet has 7% missile evasion」)。

Planet Hits 表(每項對地面設施/人口需要的「hit 數」):

| 項目 | hits |
|---|---|
| 每棟建築 | 1 |
| 有儲存生產(>0) | 1(手冊未給「越多儲存生產、機率越高」的精確曲線,只保留觸發條件) |
| 每整數人口 | 1 |
| 每人口零頭(100k) | 1 |
| Marine / Tank | 見上方 hits-to-kill |

`GroundPlanetTotalHits(buildings int, storedProductionPositive bool, fullPop, popFraction, marines, marineHitsEach, tanks, tankHitsEach int) int`(`ground.go:327`)。

### 驗證範例(`ground_test.go`)

- `GroundMarineBarracksCap(10,20,false)=5`、`(10,20,true)=10`(Warlord x2)。
- `GroundMarineBarracksUnits(30,20,20,false)=10`(4+30/5=10,已達上限 min(10,10))。
- `GroundMarineHitsToKill(true,true)=3`(基礎1+HighG1+PoweredArmor1)。
- `GroundBombHitsFromDamage(32000)=320`(剛好達上限)、`(40000)=320`(超過夾住)。
- `GroundPlanetTotalHits(5,true,3,1,4,1,2,2)=18`(5建築+1儲存+3人口+1零頭+4 Marine+4 Tank hits)。

---

## 12. 間諜(`spy.go`)

**來源**:手冊「Notes on Spying」(p113,含 Spy Bonuses / Assassins / Roll Chance / Spy vs Spy 四小節)。openorion2 未實作間諜邏輯(`spies[]` 除讀檔外從未被賦值,無 `SpyView` 類別),本檔無原始碼可對照,一律以手冊原文數字為準。

### 派駐加成曲線

`SpySlotBonus(spyCount int) int`,依派駐人數(夾限於 0-63)遞增:

```
count <= 5   → 2*count            (前 5 人各 +2)
count <= 10  → 10 + (count-5)     (第 6-10 人各 +1)
count > 10   → 15 + (count-10)/2  (之後每 2 人 +1)
```

上限 +41(62 或 63 人)。手冊原文特別舉例:"spy 11 the bonus is still +15 while spy 12 brings it up to +16"。

### 靜態加成表

| 來源 | 加成表 |
|---|---|
| 政府型態(僅 Defense) | Feudalism 0、Confederation 0、Dictatorship 10、Imperium 15、Democracy −10、Federation −10、Unification 15、Galactic Unification 15 |
| 種族間諜特性(3 檔) | −3 picks: −10、+3 picks: +10、+6 picks: +20 |
| 心靈感應種族(Telepathic) | +10(Defense/Offense 同值) |
| 科技加成 | Neural Scanner 10、Telepathic Training 5、Cybersecurity Link 10、Stealth Suit 10、Psionics 10 |
| 軍官(範圍,精確映射待查) | Telepath Defense 2-18、Spy Master Offense 2-18 |

### Roll Chance(成功機率封閉解)

行動門檻(`SpyThreshold*`):偷竊成功 80、偷竊成功且嫁禍 90、破壞成功 70、破壞成功且嫁禍 90。

有效門檻 `E = T + DB − AB`(`SpyEffectiveThreshold`)。雙方各擲 `random(100)`,`AR=100`(幸運骰)必定成功,或 `AR-DR > E` 成功。此機制的封閉解機率(`SpyRollChance`,手冊 Roll Chance → Formula):

```
E <= -100  : p = 1
-100<E<-1  : p = 1 - (101+E)*(100+E) / 2 / 10000
0<=E<=99   : p =     (99-E)*(98-E)  / 2 / 9900  + 0.01
E > 99     : p = 0.01
```

手算對照(`spy_test.go`):`E=-50 → p=0.8725`、`E=0 → p=0.5`、`E=99 → p=0.01`。

### Spy vs Spy

防禦方 +20 固定加成(`SpyVsSpyDefenderBonus`);攻擊方選擇 HIDE 指令時 +20(`SpyVsSpyAttackerBonus`)。判定門檻:防禦方在 +80 被殺,攻擊方在 −80 被殺(`SpyVsSpyDefenderKillThreshold`/`SpyVsSpyAttackerKillThreshold`);雙方都可能因幸運骰同時折損間諜。**待查證**:手冊未給這節判定用的 action threshold(T)基準值,也未明列 ±80 門檻與 `SpyEffectiveThreshold`/`SpyRollChance` 的精確對應公式。

---

## 13. 手冊自相矛盾記錄

MOO2 patch 1.5 官方手冊在兩處數值與其自身附表不一致,已在程式碼註解記錄推導過程與裁決依據,標記待實機動態驗證。

### AMR 命中率:逐字公式 vs 附表差 2

手冊逐字公式:

```
AMR Chance-to-Hit = 70 - rounddown((Range + 2) * 10 / 3) - 1
```

代入 Range=0..6 得 `63/59/56/53/49/46/43`,但手冊自己列出的核對表是 `65/61/58/55/51/48/45`——**每一項都少 2**。

**裁決**:附表才是可信結果(逐列列到個位數,顯然是作者算好貼上),反推「−1」應在 `rounddown()` 內一起捨去,而非捨去後才減 1。實際公式應為:

```
70 - (rounddown((Range+2)*10/3) - 1) = 71 - rounddown((Range+2)*10/3)
```

已用 Range 0-6 逐項代入驗證,71 減法版本與附表完全一致(見第 10 節命中率表)。`internal/gamedata/missile.go` 採用此裁決版本,`missile_test.go` 對兩張表都有測試斷言。

### 飛彈速度:公式比表格多 4

手冊明列公式:`Speed = BaseSpeed(12) + 2*(FTLlevel-1) + FastBonus(4)`,代入 FTL 0-6 得 `14/16/18/20/22/24/26`。

但同段附表(Drive/FTLlevel/Speed/Missile/MissileBonus)的 Speed 欄是 `10/12/14/16/18/20/22`——**與公式恆差 4**。

**裁決**:`internal/gamedata/missile.go` 以「明列公式」為準(手冊寫 "calculated as follows",語意上是主要規則),推測附表 Speed 欄記錄的是「驅動本身速度」這個不同的量(用於星圖移動,非戰鬥 Beam Defense 計算)。此落差尚未有第二個獨立來源可交叉驗證,標記待日後對實機行為做動態驗證確認。FTLlevel 對映表(None=0 … Interphased=6)與 MissileBonus 表(依彈頭型別)本身無爭議。

### Fantastic Trader 貿易財加成:「1:1 轉換」vs「+50%」兩種敘述

GAME_MANUAL.pdf 對 Fantastic Trader 種族特質的貿易財加成,在兩處給了數字上一致但敘述方式互相矛盾的說法:

- p.70(Trade Goods 一般說明,與其他種族並列):「Every 2 industry converts to 1 BC... unless you are a Fantastic Trader in which case every 1 industry converts to 1 BC」——直接給轉換比,換算下來是一般種族的 **2 倍**(2:1 → 1:1)。
- p.25(種族特質「Fantastic Traders」專屬說明):「on top of that, traders get a **50% bonus** to all income derived from producing trade goods」——字面讀法是一般種族的 **1.5 倍**,而非 2 倍。

兩處的「倍率」讀法不一致(2 倍 vs 1.5 倍),但無法判斷是否為同一件事的兩種模糊敘述,或是疊加關係(1:1 轉換之上再疊加 50%,變 3 倍)。**裁決**:`internal/gamedata/income.go` 的 `TradeGoodsIncome` 採用 p.70 這句「與一般種族轉換比並列、直接給數字」的敘述(1:1),因為它與貿易財換算規則本身描述方式一致、格式對稱,判讀歧異度較低;p.25 的「+50%」說法**未在此另外實作**,標記待查證原版程式行為(TODO,見 `income.go` 常數區塊註解)。與第 13 節前兩則(AMR 命中率、飛彈速度)的差異:那兩則手冊的公式與附表可互相代入驗證出「哪個是筆誤」,這一則的兩種讀法都是完整敘述句、缺乏第三個數字可佐證,不確定性高於前兩則。

### Laser Cannon 距離衰減範例表:單一儲存格與公式算出值不同

MANUAL_150.html「Reduced by Range」表用 Phasor(base 5-20)、Mauler(base 100-100)、Death Ray(base 50-100)三組共 30+ 個數字逐格核對 `DamageApplyDissipation` 公式(`base * dmg% ,四捨五入`),全數吻合。但同一張表的 Laser Cannon(base 1-4)範例列,在 range level 7(19-21 sq,傷害% 40%)欄印的是「**1-1**」(max=1),而套用相同公式算出 `round(4*40/100)=round(1.6)=2`(max=2),對不上。

**裁決**:判定 Laser Cannon 那一格是手冊排版/校對誤差,不因單一儲存格回頭修改已被其餘三組武器、Hv/PD/HEF 加成公式(`DamageMountAdjustedValue`)共同驗證過的通用公式。`internal/gamedata/damage.go` 的 `DamageApplyDissipation` 保留公式版本(max=2),`damage_test.go` 的 `TestDamageApplyDissipationLaserKnownMismatch` 明文記錄這個已知落差,避免日後誤以為沒注意到、或誤把公式改成遷就這一格。

---

## 14. 尚未移植/待查證

依 `docs/tech/rules-implementation-audit.md` 與 `docs/tech/game-logic-port.md` 的盤點,以下系統 openorion2 完全沒有可複用的邏輯(連 UI 殼常常都沒有),需要完全依手冊從零設計:

| 系統 | 狀態 | 備註 |
|---|---|---|
| 傷害解算細節 | 部分完成(第 9 節 `damage.go`) | 距離衰減、Hv/PD/HEF mount 加成、命中後傷害內插、護盾減傷、裝甲穿透已移植;球形武器(Pulsar/Plasma Flux/Spatial Compressor)的重骰終止條件、引擎爆炸逐格衰減率手冊描述不足,仍待查證 |
| 地面戰(登陸/轟炸) | 部分完成(第 11 節 `ground.go`) | Barracks 建造/人口上限、單位血量、裝甲/裝備/種族/地形戰力加成、轟炸命中換算已移植;Commando Leader 基準加成、AI Ground Troops Bonus 依難度分級數字、Stored Production 命中曲線手冊未給精確數字,仍待查證 |
| 外交 | 未見 | 連 `DiplomacyView` 畫面殼都不存在;AI 六種目標性格判定、Diplomatic Blunder/Marriage 事件全無邏輯可抄 |
| AI 決策 | 未見 | 全 repo 零 RNG 來源,任何 AI 判斷邏輯都要重新設計 |
| 回合編排(把上述公式串成回合) | 待補 | `researchProgress`/`experience` 等欄位全 repo 除建構子外從未被賦值,無回合結算函式存在 |
| RNG(命中/間諜/閃避擲骰) | 待補 | 各公式已給出「決定性機率/門檻」,但實際擲骰與可重現的 RNG(含 seed 管理、存檔是否存 RNG 狀態)尚未設計 |
| 星系/星圖生成 | 未見 | 星系形狀/星星分布/行星屬性/特殊天體的隨機生成演算法要整個重寫 |
| 種族特性效果套用 | 僅列舉常量 | `RaceTrait` 32 項有列舉與唯讀顯示,但沒有任何函式把特性數值套進生產/戰鬥/成長公式(因為這些公式所在系統本身當時就沒有實作) |
| 勝利條件 | 未見 | 三種勝利路徑(殲滅/票選/次元傳送門攻陷 Antares)與計分公式全無 |
| Antaran/Orion 事件 | 僅美術資源 | 隨機襲擊、守護者遭遇戰、次元傳送門終局戰觸發鏈全無邏輯 |

> 完整逐系統盤點(含每個結論的原始碼行號依據)見 `docs/tech/rules-implementation-audit.md`;移植進度追蹤見 `docs/tech/game-logic-port.md`。本文件與兩者互為參照:本文件是「公式本身」的查閱介面,`rules-implementation-audit.md` 是「openorion2 有沒有可抄的邏輯」的原始碼盤點,`game-logic-port.md` 是移植的進度總表。

---

## 參考來源總表

| 來源 | 路徑 | 用途 |
|---|---|---|
| openorion2 原始碼(GPL v2) | `openorion2/src/gamestate.cpp`、`tech.cpp` | 唯讀衍生公式、研究樹拓撲、艦艇/軍官查表 |
| MOO2 patch 1.5 說明書(HTML) | `moo2_patch1.5/MANUAL_150.html` | 殖民地成長、光束命中/傷害解算、飛彈防禦、地面戰(轟炸)、間諜、政府 BC 收入加成的官方公式附錄 |
| MOO2 patch 1.5 完整手冊(PDF) | `moo2_patch1.5/GAME_MANUAL.pdf` | 生產/污染、士氣、國庫收入、地面戰(Barracks/裝甲/裝備)、護盾/裝甲穿透的系統說明章節 |
| SA1 盤點 | `docs/tech/rules-implementation-audit.md` | openorion2 逐系統實作程度盤點 |
| 移植進度 | `docs/tech/game-logic-port.md` | 各公式移植狀態與優先序 |
| 既有唯讀公式速查 | `docs/tech/formulas.md` | 艦艇/殖民地/軍官公式的精簡版(本文件的完整可查證版本) |
