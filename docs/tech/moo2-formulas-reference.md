# Master of Orion II 遊戲公式參考

## 前言

本專案(`master-of-orion2-remake-cht`)是目前已知**世界首個**以 Go/ebiten 重寫的 *Master of Orion II* remake,目標是在保留原版規則的前提下完整重建遊戲邏輯,並提供完整繁體中文化。這份文件把目前已移植進 `internal/gamedata/` 的規則公式集中彙整、附上來源出處與驗證方式,作為後續開發與外部查證的知識庫,也是本專案在規則考據上的產出之一。

### 公式的兩個來源

`internal/gamedata/rules-implementation-audit.md`(`docs/tech/rules-implementation-audit.md`)逐系統盤點過 openorion2 (GPL v2,本專案參考的既有 C++ 實作)後確認:**openorion2 是存檔載入與檢視殼,不是回合制遊戲引擎**。全 repo 對 `endTurn/nextTurn/processTurn` 與 `rand()/mt19937` 等隨機數來源零命中——它能正確顯示艦艇戰力、研究進度、軍官技能這類「已知輸入 → 顯示輸出」的**唯讀衍生公式**,但不含任何「回合推進 → 狀態改變」的規則引擎(戰鬥解算、殖民地成長、外交、間諜結算等)。

因此本專案的公式移植分兩路:

1. **openorion2 唯讀表**:艦艇電腦/引擎 HP、戰速、光束攻防、軍官經驗/技能/雇用費、研究樹拓撲(`research_choices[83]`)——這些是 openorion2 `gamestate.cpp`/`tech.cpp` 裡真正在跑、且已被 UI 使用的公式,直接逐字轉寫進 Go。
2. **官方手冊權威公式**:openorion2 沒有實作的系統(殖民地成長、生產/污染、光束命中、飛彈防禦、間諜),改以 `moo2_patch1.5/MANUAL_150.html`(1.50 patch 說明書,含 "Notes on Population Growth"、"Notes on Spying"、"Notes on Missile Defenses"、"Notes on Anti-Missile Rockets" 等附錄段落,俗稱「The Algorithm」)與 `moo2_patch1.5/GAME_MANUAL.pdf`(隨 1.50 patch 附的完整遊戲手冊)逐條抽取常數與公式移植。

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
3. [研究樹](#3-研究樹-techtreego)
4. [軍官](#4-軍官-officergo)
5. [艦艇衍生值](#5-艦艇衍生值-formulasgo)
6. [光束武器命中](#6-光束武器命中-combatgo)
7. [飛彈防禦與反飛彈火箭](#7-飛彈防禦與反飛彈火箭-missilego)
8. [間諜](#8-間諜-spygo)
9. [手冊自相矛盾記錄](#9-手冊自相矛盾記錄)
10. [尚未移植/待查證](#10-尚未移植待查證)

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

**來源**:`GAME_MANUAL.pdf`(patch 1.5 隨附完整手冊)「System Overview / Yield / Population」章節(約 p.64-67)與各建築說明章節(約 p.78-90)。`MANUAL_150.html`(1.50 patch 說明書)本身只在 UI bugfix 段落提過一次 pollution,無數值公式;openorion2 只有 `Planet::baseProduction()` 單一查表(見第 5 節),未實作生產分配與污染公式,故本節數值全部來自 `GAME_MANUAL.pdf`。

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

## 3. 研究樹(`techtree.go`)

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

## 4. 軍官(`officer.go`)

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

## 5. 艦艇衍生值(`formulas.go`)

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
| `LeaderHireCost(...)` | `gamestate.cpp:700-701` | 見第 4 節 |

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

## 6. 光束武器命中(`combat.go`)

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

## 7. 飛彈防禦與反飛彈火箭(`missile.go`)

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

**AMR 命中率**(`MissileAMRChanceToHit`,見第 9 節手冊矛盾記錄取得的裁決公式):

```
MissileAMRChanceToHit(range) = 71 - (range+2)*10/3   (整數除法)
```

| Range | 0 | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|---|
| Chance-to-Hit | 65% | 61% | 58% | 55% | 51% | 48% | 45% |

命中一次只摧毀彈頭堆疊中的一枚飛彈,且與目標飛彈種類/血量/mods 無關,只與距離有關。

### 飛彈 Beam Defense

`MissileSpeed`/`MissileBeamDefense`(見第 9 節手冊矛盾記錄):

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

## 8. 間諜(`spy.go`)

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

## 9. 手冊自相矛盾記錄

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

已用 Range 0-6 逐項代入驗證,71 減法版本與附表完全一致(見第 7 節命中率表)。`internal/gamedata/missile.go` 採用此裁決版本,`missile_test.go` 對兩張表都有測試斷言。

### 飛彈速度:公式比表格多 4

手冊明列公式:`Speed = BaseSpeed(12) + 2*(FTLlevel-1) + FastBonus(4)`,代入 FTL 0-6 得 `14/16/18/20/22/24/26`。

但同段附表(Drive/FTLlevel/Speed/Missile/MissileBonus)的 Speed 欄是 `10/12/14/16/18/20/22`——**與公式恆差 4**。

**裁決**:`internal/gamedata/missile.go` 以「明列公式」為準(手冊寫 "calculated as follows",語意上是主要規則),推測附表 Speed 欄記錄的是「驅動本身速度」這個不同的量(用於星圖移動,非戰鬥 Beam Defense 計算)。此落差尚未有第二個獨立來源可交叉驗證,標記待日後對實機行為做動態驗證確認。FTLlevel 對映表(None=0 … Interphased=6)與 MissileBonus 表(依彈頭型別)本身無爭議。

---

## 10. 尚未移植/待查證

依 `docs/tech/rules-implementation-audit.md` 與 `docs/tech/game-logic-port.md` 的盤點,以下系統 openorion2 完全沒有可複用的邏輯(連 UI 殼常常都沒有),需要完全依手冊從零設計:

| 系統 | 狀態 | 備註 |
|---|---|---|
| 傷害解算細節 | 待補 | 本文件涵蓋「命中判定」,尚未涵蓋傷害量計算、球形武器(Pulsar/Plasma Flux/Spatial Compressor)範圍傷害、护盾削減等 |
| 地面戰(登陸/轟炸) | 未見 | 手冊「Notes on Orbital Assault」(轟炸模擬 10 回合射擊、建築/人口/儲存產能各自命中判定)與「Ground Defenses & Troops」全無 openorion2 對應程式碼 |
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
| MOO2 patch 1.5 說明書(HTML) | `moo2_patch1.5/MANUAL_150.html` | 殖民地成長、光束命中、飛彈防禦、間諜的官方公式附錄 |
| MOO2 patch 1.5 完整手冊(PDF) | `moo2_patch1.5/GAME_MANUAL.pdf` | 生產/污染的系統說明章節 |
| SA1 盤點 | `docs/tech/rules-implementation-audit.md` | openorion2 逐系統實作程度盤點 |
| 移植進度 | `docs/tech/game-logic-port.md` | 各公式移植狀態與優先序 |
| 既有唯讀公式速查 | `docs/tech/formulas.md` | 艦艇/殖民地/軍官公式的精簡版(本文件的完整可查證版本) |
