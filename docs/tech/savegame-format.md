# 存檔格式 save?.gam(逆向 + 移植紀錄)

> 記錄 Master of Orion II 存檔的精確佈局。來源:openorion2 `src/gamestate.cpp`(GPL v2)逐欄位核對 + 本專案 `internal/save` Go 移植 + 真實檔 `SAVE10.GAM`(208000 bytes)驗證。
> 全部**小端**。字串欄為固定長度,截到第一個 NUL。

## 0. 關鍵常數與 offset

| 名稱 | 值 | 說明 |
|---|---|---|
| 版本 | `0xe0` | `GameConfig.version`,不符即拒絕 |
| `GameConfig` | offset 0 | 存檔開頭 |
| `COLONY_COUNT_OFFSET` | `0x25b`(603) | colonyCount 位置,順序區起點 |
| galaxy offset | `0x31be4`(203748) | galaxy 存在檔尾,獨立 seek |
| 順序區結尾(SeqEnd) | **203596** | ships 讀完的位置(< 203748,無碰撞) |
| 檔案大小(SAVE10) | 208000 | 固定大小存檔 |

**存檔非純序列**:先讀 `GameConfig`(offset 0),seek 到 `0x31be4` 讀 galaxy,再 seek 回 `0x25b` 讀順序區。

## 1. 讀取序列(GameState::load)

```
GameConfig            @0
galaxy                @0x31be4(seek)
── 順序區 @0x25b(seek)起,以下連續 ──
u16 colonyCount   → Colony  × 250 (MAX_COLONIES)
u16 planetCount   → Planet  × 360 (MAX_STARS×MAX_ORBITS = 72×5)
u16 starCount     → Star    × 72  (MAX_STARS)
                    Leader  × 67  (LEADER_COUNT,無獨立計數)
u16 playerCount   → Player  × 8   (MAX_PLAYERS)
u16 shipCount     → Ship    × 500 (MAX_SHIPS)
```

各實體陣列**一律含滿 MAX_ 個**(未使用者為空資料);`*Count` 只是有效數。

## 2. 結構大小一覽(wire size,bytes)

| 結構 | 大小 | 備註 |
|---|---|---|
| GameConfig | 59 | version4 + name37 + stardate4 + 14 旗標 |
| Nebula | 5 | x2 y2 type1 |
| Galaxy | 32 | sizeFactor1 + skip4 + w2 + h2 + skip2 + Nebula×4(20) + nebulaCount1 |
| Colonist | 4 | bit-packed u32 |
| Colony | 361 | 見 §4 |
| Planet | 16 | |
| Star | 113 | |
| Leader | 59 | = LEADER_DATA_SIZE(交叉驗證通過) |
| ShipWeapon | 8 | |
| ShipDesign | 99 | name16 + 7 + specials5 + Weapon×8(64) + 6 |
| Ship | 129 | ShipDesign99 + 30 |
| SettlerInfo | 4 | bit-packed |
| Player | 3755 | 見 §5 |

陣列容量常數:MAX_POPULATION=42、MAX_RACES=10(玩家+androids+natives)、MAX_BUILD_QUEUE=7、MAX_BUILDINGS=49、MAX_ORBITS=5、MAX_SETTLERS=25、MAX_RESEARCH_TOPICS=83、MAX_TECHNOLOGIES=212(204 applied + 8 areas)、MAX_PLAYER_BLUEPRINTS=5、TRAITS_COUNT=31、MAX_HISTORY_LENGTH=350、MAX_SHIP_SPECIALS=40、MAX_SHIP_WEAPONS=8。

## 3. GameConfig(59 bytes)

version u32 → name[37] → stardate u32 → 14 個 u8 旗標:multiplayer、endOfTurnSummary、endOfTurnWait、randomEvents、enemyMoves、expandingHelp、autoSelectShips、animations、autoSelectColony、showRelocationLines、showGNNReport、autoDeleteTradeGoodHousing、showOnlySeriousTurnSummary、shipInitiative。

## 4. Colony(361 bytes)欄位順序

owner u8, unknown1 i8, planet i16, unknown2 i16, is_outpost u8, morale i8, pollution u16, population u8, colony_type u8;
Colonist[42];race_population[10] u16;pop_growth[10] i16;
age/foodPerFarmer/industryPerWorker/researchPerScientist u8, maxFarms i8, maxPopulation u8, climate u8;
groundStrength/spaceStrength/totalFood/netIndustry/totalResearch/totalRevenue u16;
foodConsumption/industryConsumption/researchConsumption/upkeep u8;
foodImported i16, industryConsumed u16, researchImported i16, budgetDeficit i16, recycledIndustry u8;
food 消耗 citizens/aliens/prisoners/natives u8;industry 消耗 citizens/androids/aliens/prisoners u8;
foodConsumptionRaces[8] u8;industryConsumptionRaces[8] u8;replicatedFood u8;
buildQueue[7] i16;finishedProduction i16, buildProgress u16, taxRevenue u16, autobuild u8, unknown3 u16, boughtProgress u16, assimilationProgress u8, prisonerPolicy u8, soldiers u16, tanks u16, tankProgress u8, soldierProgress u8;
buildings[49] u8;status u16。

**Colonist(4 bytes,u32 bit-packed)**:`race = bits[0:4]`、`loyalty = bits[4:7]`、`job = bits[7:9]`、`flags = bits[9:]`。

## 5. Player(3755 bytes)欄位順序(含相對 skip)

skip1;name[20];race[15];eliminated/picture/color/personality/objective u8(personality 100 = 人類);
homePlayerId u16, networkPlayerId u16, playerDoneFlags u8, skip2(dead), researchBreakthrough u8, taxRate u8, BC i32, totalFreighters u16, surplusFreighters i16, commandPoints u16, usedCommandPoints i16, foodFreighted u16, settlersFreighted u16;
SettlerInfo[25];
totalPop/foodProduced/industryProduced/researchProduced/bcProduced u16, surplusFood i16, surplusBC i16;
totalMaintenance i32, building/freighter/ship/spy/tribute/officer Maintenance u16;
researchTopics[83] u8;techs[212] u8;researchProgress u32;**skip45**;hyperTechLevels[8] u8;**skip253**;researchTopic u8, researchItem u8;skip3;
blueprints[5] ShipDesign;selectedBlueprint ShipDesign;**skip12**;playerContacts[8] u8;**skip139**;playerRelations[8] i8;skip8;foreignPolicies[8]/tradeTreaties[8]/researchTreaties[8] u8;**skip608**;traits[31] i8;skip33;
fleetHistory[350]/techHistory[350]/populationHistory[350]/buildingHistory[350] u8;spies[8] u8, infoPanel u8;skip21;galaxyCharted u8;skip51。

**SettlerInfo(4 bytes,LE bit-packed)**:byte0=sourceColony、byte1=destinationPlanet、byte2 低 4 bit=player 高 4 bit=eta、byte3 低 2 bit=job(其餘保留)。原版用 `BitStream` 逐 byte 依需求讀,26 bit 落在 4 byte 內。

## 6. Star(113)/ Leader(59)/ Ship(129)重點

- **Star**:name[15], x/y u16, size u8, owner i8, pictureType/spectralClass u8, lastPlanetSelected[8] u8, blackHoleBlocks[9](=(72+7)/8) u8, special u8, wormhole i8, blockaded u8, blockadedBy[8] u8, 15 個 u8 旗標(visited…isStagepoint), officerIndex[8] i8, planetIndex[5] i16, relocateShipTo[8] u16, skip3, surrenderTo[8] u8, inNebula u8, artifactsGaveApp u8。
- **Leader**:name[15], title[20], type u8, experience u16, commonSkills u32, specialSkills u32, techs[3] u8, picture u8, skillValue u16, level u8, location i16, eta u8, displayLevelUp u8, status i8, playerIndex i8。**合計 59 = LEADER_DATA_SIZE**(獨立交叉驗證)。
- **Ship**:ShipDesign(99) + owner/status u8, star i16, x/y u16, groupHasNavigator/warpSpeed/eta/shieldDamage/driveDamage/computerDamage/crewLevel u8, crewExp u16, officer i16, damagedSpecials[5], armorDamage u16, structureDamage u16, mission u8, justBuilt u8。

## 7. 真實檔驗證(SAVE10.GAM)

| 欄位 | 值 |
|---|---|
| 存檔名 | `(Auto Save)` |
| stardate | 35000 |
| galaxy | 759 × 600,sizeFactor 15 |
| 計數 | colony 5 / planet 82 / star 36 / player 5 / ship 21(皆合理,readCount 上限把關) |
| 首星 | `Orion` |
| 有名星系數 | 36 == starCount(強對位證據) |
| 玩家 | Klirr(Trilarian)、Karaaw Hrik(Alkari)、Parasha Vrrn(Mrrshan)、Sarezaear(Sakkra)、Qurtirqul(Klackon) |
| SeqEnd | 203596(順序區收斂,合成全零檔同值當回歸護欄) |

> personality 全非 100 → 此 autosave 無標記人類玩家(觀戰/全 AI),非 schema 錯誤 —— 名稱/種族/計數對位已證明結構正確。
> 尚未逆向:`0x25b` 之前(除 GameConfig 59 bytes)的區段內容、順序區與 galaxy 間的 152 byte 空隙、Colony `unknown1/2/3` 等 FIXME 欄位語意。
