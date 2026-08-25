# 領袖／軍官技能系統(2026-07-11 接線,2026-08-10 remake 收尾)

> **本文第二、三節是一手資料(技能 id 編碼、加成公式、格式字串對照),仍然有效。**
> 第四節以後是接線現況,已由第 45/45/45 項大幅改寫——2026-07-11 版本裡「只接三個技能」
> 「指揮官待定案」「工程師沒有承接系統」那些敘述**都已過期**,見第 4.3 節與第五節。

## 一、問題背景

接手時 `internal/shell/session.go` 的 `Leader{Name, Skill string, Level int, Ship bool}` +
`demoLeaders()`(如「馮·諾伊曼 科學家 Level 5」)純粹是軍官列表畫面的顯示資料,技能字串不
bonus 任何東西。`internal/gamedata/officer.go`(`LeaderExpLevel`/`LeaderSkillBonus`)與
`formulas.go`(`LeaderHireCost`)是已移植但零呼叫端的死碼(公式對,沒人叫)。本輪把技能真正接
進殖民地產出與(有限的)戰鬥/雇用引擎函式。

## 二、技能定義來源(硬門檻查證結果)

**技能 id 列舉已在 `internal/gamedata/enums.go` 生成**(`type LeaderSkills int` +
`SKILL_ASSASSIN`..`SKILL_TACTICS`,對照 `openorion2/src/gamestate.h:602-631`,
`enums_test.go` `TestEnumSpotValues` 已回歸鎖 3 個抽查值)——**這件事之前沒人接上**,`officer.go`
自己重複定義了 `skillMegawealth`/`skillNavigator` 兩個私有常數,沒有引用完整枚舉。本輪清掉重複
定義,`officer.go` 直接 `int(SKILL_MEGAWEALTH)`/`int(SKILL_NAVIGATOR)`。

技能 id 編碼(`gamestate.h:36-45`):bit4-5 = 類型(`0x00` common / `0x10` captain / `0x20`
admin),bit0-3 = 該類型內的技能碼(依 enum 宣告順序,0 起算)。

技能加成公式(`Leader::skillBonus`,`gamestate.cpp:664-692`,已在 `officer.go` `LeaderSkillBonus`
移植且有單元測試):

- tier(技能階)0 → 恆 0(沒有這技能)。
- Navigator 用專屬值表 `navigatorSkillValues[tier>1][expLevel]`,不套下列通則。
- 一般技能:`base = baseSkillValues[type][code]`(`gamestate.cpp:75-79` 三行常數表,已移植);
  除 Megawealth 外乘以 `(expLevel+1)`;tier>1(進階技能)再 `+50%`。

技能單位(`officer.cpp:75-87` `skillFormatStrings`,決定該 base 值是「固定數字」還是「百分比」,
是判斷「該技能該接到 remake 哪個欄位」的關鍵一手資料,先前完全沒人查過):

| 技能 | 類型 | 格式 | base 值 |
|---|---|---|---|
| Assassin | common | `%d%%` | 2 |
| Commando | common | `%+d` | 2 |
| Diplomat | common | `%+d` | 10 |
| Famous | common | `%+dBC` | -60(恆負,雇用費修正) |
| Megawealth | common | `%+dBC` | 10(不隨等級倍增) |
| Operations | common | `%+d` | 2 |
| **Researcher** | common | `%+d`(固定研究點,非%) | **5** |
| Spymaster | common | `%+d%%` | 2 |
| Telepath | common | `%+d%%` | 2 |
| **Trader** | common | `%+d%%`(收入%) | **10** |
| Engineer | captain | `%+d%%` | 2 |
| Fighter Pilot | captain | `%+d` | 5 |
| Galactic Lore | captain | `%+d` | 5 |
| **Helmsman** | captain | `%+d`(直接加 BD) | **5** |
| Navigator | captain | 專屬值表 | — |
| Ordnance | captain | `%+d` | 5 |
| Security | captain | `%+d` | 2 |
| **Weaponry** | captain | `%+d`(直接加 BA) | **5** |
| Environmentalist | admin | `%+d%%` | -10 |
| Farming Leader | admin | `%+d%%` | 10 |
| **Financial Leader** | admin | `%+d%%`(收入%) | **10** |
| Instructor | admin | `%+d` | 1 |
| **Labor Leader** | admin | `%+d%%`(工業%) | **10** |
| Medicine | admin | `%+d%%` | 10 |
| **Science Leader** | admin | `%+d%%`(研究%) | **10** |
| Spiritual Leader | admin | `%+d`(士氣,無換算公式) | 5 |
| Tactics | admin | `%+d`(手冊自陳未實作) | 2 |

## 三、實際生效的技能(openorion2 gamestate.cpp 有真呼叫端)

grep 全 `openorion2/src/*.cpp` 後,只有 4 個技能在 openorion2 本身有真正的計算呼叫端(其餘
20+ 個技能只有畫面顯示/skillBonus 可算但沒有效果消費端——**openorion2 是 GUI 殼,不是完整引擎,
這點與既有 memory `openorion2-is-renderer-not-engine` 一致,連原專案自己都沒把大多數領袖技能的
「玩法效果」寫出來**):

1. `SKILL_WEAPONRY` → `GameState::shipBeamOffense`(`gamestate.cpp:2372-2374`):`ret +=
   officer.skillBonus(SKILL_WEAPONRY)`,直接加進艦艇命中值(BA)。
2. `SKILL_HELMSMAN` → `GameState::shipBeamDefense`(`gamestate.cpp:2400-2402`):同樣直接加進
   閃避值(BD)。
3. `SKILL_FAMOUS` → `GameState::leaderHireModifier`(`gamestate.cpp:2407-2426`):對該玩家所有
   「已受雇」領袖取 Famous `skillBonus` 的**最小值(MIN,非累加)**,當雇用費修正 modifier
   餵給 `Leader::hireCost`。
4. `SKILL_MEGAWEALTH` → `GameState::leaderMaintenanceCost`(`gamestate.cpp:2428-2441`):
   `hasSkill(SKILL_MEGAWEALTH)` 為真則維護費全免(不算 skillBonus 數值,是布林開關)。
   附帶硬編例外 `LEADER_ID_LOKNAR`(=65,特定英雄免費)——那是具名英雄 ID 的例外,不是技能規則,
   remake 沒有這個角色,不移植。

Research/Farming/Financial/Labor/Science/Instructor/Environmentalist/Spiritual/Tactics/
Assassin/Commando/Diplomat/Operations/Spymaster/Telepath/Engineer/Fighter Pilot/Galactic
Lore/Ordnance/Security/Navigator(移動力用途)在 openorion2 全專案 grep 零命中(不是搜尋落空,
是真的沒有效果消費端——已用 `rulebook/62` 反向溯源 SOP 驗證,`shipCombatSpeed`/`shipBeamOffense`/
`shipBeamDefense`/`leaderHireModifier`/`leaderMaintenanceCost` 已是 `gamestate.cpp` 全部呼叫
`Leader::skillBonus`/`hasSkill` 的地方,沒有第五個呼叫點)。這些技能的「效果」只存在於手冊文字
描述(如 Spiritual Leader「Raises morale」、Commando「ground combat strength」)。

> ⚠ **2026-08-08 更正這一段的最後一句。** 原文寫「精確數字/換算公式手冊沒給」,
> 那是把「openorion2 沒有消費端」誤讀成「沒有數字」。**數字一直都在**:加成值是
> `baseSkillValues`(第二節那三行常數表),單位是 `skillFormatStrings`(第二節那張對照表)
> ——兩者合起來就足以決定該技能接進 remake 的哪個欄位。缺的從來是**承接的子系統**,
> 不是數字。2026-08-10 收尾後,26 項技能至少有一個 remake 消費端；唯一刻意不接的是
> 手冊明說原版也未實作的 Tactics。Famous 的雇用費折扣與招募機率都已接；後者由
> 2026-08-25 的 IDA 靜態資料流補證，不再沿用固定週期模型（見第六節）。
> 真正「沒有數字」的只有戰術官,而那是因為**原版自己就沒實作**。

## 四、本輪建置範圍(只接對應到 remake 已存在系統的技能)

### 4.1 gamedata 新增(`internal/gamedata/officer.go`)

- `LeaderTypeCaptain`/`LeaderTypeAdmin`(對照 `gamestate.h:32-33`)。
- `LeaderSkillTier(skillID, leaderType int, commonSkills, specialSkills uint32) int`:對照
  `Leader::hasSkill`(`gamestate.cpp:631-662`),從 2-bit 位元組解出技能階。**這個函式讓未來讀
  真存檔的 `save.Leader`(`CommonSkills`/`SpecialSkills` 欄位,`internal/save/entities.go:340-341`,
  本身已完整解析但零呼叫端)可以算出真實技能階**,不必依賴 demo 資料手動指定 Tier。
- `LeaderMaintenanceCost(hireCost int, hasMegawealth bool) int`:port `leaderMaintenanceCost`。
- `LeaderHireModifier(famousBonuses []int) int`:port `leaderHireModifier`(MIN 語意)。

### 4.2 engine 新增

- `internal/engine/ship.go`:`ShipBeamAttackWithOfficer`/`ShipBeamDefenseWithOfficer`——在既有
  `ShipBeamAttackFromDesign`/`ShipBeamDefenseFromDesign`(已有測試鎖住既有行為,故不改簽章)之上
  疊加軍官 Weaponry/Helmsman 加成,對照 `shipBeamOffense`/`shipBeamDefense` 的疊加方式。
  `shell.Ship.OfficerName`／`OfficerID` 現已提供逐艦來源；快速結算與格子戰術會讀同一份指派資料，
  UI 路徑為艦隊畫面選船 → `LEADERS` → 點艦艇軍官列。
- `internal/engine/leader.go`:`HireLeader(currentBC, cost int) (newBC int, ok bool)`——最小雇用
  金流機制(BC 夠不夠、扣款),供未來招募畫面呼叫。**領袖狀態轉換(ForHire→Working 等)不在本輪
  範圍**,`demoLeaders` 既有領袖視為已受雇,不需要走這個函式。

### 4.3 shell 接線(現況,2026-08-10)

**識別鍵是技能 id,不是中文標籤。** `shell.Leader` 帶 `Skills []LeaderSkill{ID, Tier}`,
由 HERODATA 的技能位元解出(每技能 2 bit = 技能階,見第 45 項(領袖技能));`Skill` 那個字串只負責顯示,
會隨語言翻譯,**不可拿來比對**。沒有 `Skills` 的舊資料(demo 領袖、既有測試、舊存檔)
退回用 `gamedata.LeaderSkillIDByZH` 反查單一技能。

技能名字表在 `internal/gamedata/leader_skill_names.go`:27 個技能的 id ↔ 中英文名,
英文名逐字取自 GAME_MANUAL.pdf p.135-137;另含 `LeaderSkillIDsFor(leaderType)` 給出原版技能欄的
列舉順序(專屬技能在前,對照 `LeaderSkillsWidget::update`)。

`applyLeaderColonyBonuses` 逐**技能**跑(不是逐領袖),先依技能分組收集、再依
`gamedata.LeaderSkillCombine` 合成(手冊 p.137:只有 Megawealth 與 Researcher 累加,
其餘取最強那一位),最後由 switch 決定落在哪個 `ColonyState` 欄位。

艦艇軍官則由 `internal/shell/officer_assignment.go` 逐艦查詢：`AssignOfficerToShip` 會維持一位
軍官只服務一艘船，`OfficerName` 空值代表未指派；`OfficerID` 對應原版 `_leaders[]` 序號，
缺欄位時才以名稱回退。Weaponry／Helmsman、Navigator、Engineer
各自從該船查詢，不再把帝國內任一艦艇軍官套到所有艦艇。JSON／熱座沿用既有 `Fleet` 保存，
完整來源 ID 與限制見 `docs/re/officer-ids.md`；原版 `.GAM` 全局 `0x3B×0x43` 讀寫鏈已由
IDA 靜態證實，但重製直接 importer 仍未實作。

## 五、目前有實際效果的技能(26 項；另有 1 項依原版不實作)

| 技能 | 落在 | 消費端 |
|---|---|---|
| 刺客 Assassin | 每回合逐位刺客、逐對手各擲一次防守 Agent 行動 | `advanceLeaderAssassinActions` |
| 指揮官 Commando | 地面戰 force 加成 | `ground_invasion.go` `commandoLeaderTier` |
| 外交官 Diplomat | 提案關係增益的可觀察代理值 | `diplomacyRelationGain`；不是原版獨立接受分數宣稱 |
| 名人 Famous | 雇用費折扣取最強者；逐回合招募機率 | `MercHireCost`／`officerRecruitChance` |
| 巨富 Megawealth | 每回合固定 BC、領袖維護費免除 | `EndTurn`／`leader_upkeep.go` |
| 後勤官 Operations | 帝國指揮點數供給 | `totalCommandPointsSupply` |
| 科學家 Researcher | `FlatResearch`(固定點數,**累加型**) | `applyLeaderColonyBonuses` |
| 間諜大師 Spymaster | 進攻方間諜加成 | `advanceEspionage` |
| 心靈感應者 Telepath | 防守方 Agent 加成 | `advanceEspionage` |
| 貿易家 Trader | `IncomeBonusPercent` | 同上 |
| 工程師 Engineer | 被指派艦隊的船戰後完全修復(**打贏才有**) | `repair.go` `assignedEngineerTier` |
| 戰機飛行員 Fighter Pilot | 參戰戰機命中／防禦資料 | `StartCombat`／`tacticalfighter.go` |
| 銀河學者 Galactic Lore | 星圖立即揭露；太空怪獸／安塔蘭戰鬥加成 | `StarChartVisible`／`monster.go`／`antaran_victory.go` |
| 舵手 Helmsman | 艦艇閃避與飛彈閃避 | `StartCombat`／`shipOfficerMissileEvasionBonus` |
| 領航員 Navigator | 被指派艦隊航速 + 星雲/黑洞豁免 | `starlane.go` `FleetHasNavigator` |
| 軍械官 Ordnance | 艦艇武器傷害上限加成 | `StartCombat`／`mkPlayerCombatantsIndexed` |
| 保安官 Security | 登艦防守陸戰隊戰力 | `BoardingDefense`／`CombatShip.SecurityBonus` |
| 武器官 Weaponry | 被指派艦艇光束命中 | `StartCombat`／`mkPlayerCombatantsIndexed` |
| 財務官 Financial Leader | `IncomeBonusPercent` | 同上 |
| 心靈導師 Spiritual Leader | `MoralePercent` | 同上 |
| 醫官 Medicine | `GrowthBonusSum` | 同上 |
| 農業官 Farming Leader | `FoodBonusPercent` | 同上 |
| 勞工官 Labor Leader | `IndustryBonusPercent` | 同上 |
| 科學官 Science Leader | `ResearchBonusPercent` | 同上 |
| 教官 Instructor | 艦員每回合經驗(固定點數) | `leaderInstructorXPBonus` / `crew.go` |
| 環保官 Environmentalist | `PollutionReductionPercent`(**正的減幅**) | `engine/colony.go` `colonyPollution` |

> 2026-07-11 版本這一節寫的是「指揮官映射待人工定案,候選 SKILL_WEAPONRY /
> SKILL_COMMANDO / SKILL_SECURITY」——**已定案為 SKILL_COMMANDO**(手冊 p.135 Commando)。
> 同版本說工程師「沒有承接系統」,那是只看了手冊那條的第一句;第二句
> 「repairs all structural and internal systems damage after the battle is won」
> 對得上既有的 `repairAfterBattle`,已接。

## 六、仍未接的與理由

- **戰術官 Tactics**:**原版自己就沒實作**——手冊那條的最後一句明寫
  「This skill is not implemented」。不做它與原版一致,不是缺口。
- **Famous 的招募機率已補證**：`sub_9781D @ 0x9781D` 以 Famous 的兩個技能 bit
  `0x40/0x80` 分別加入 `level+1` 與 `15*(level+1)/10`，再除以目前領袖總數加一。
  候選 gate 與隨機前綴見 `docs/re/random-officer-recruitment-audit-20260825.md`。
- **Diplomat 的原版精確接受門檻**：remake 沒有獨立外交點數欄位，現在把技能值映射到既有
  關係增益，讓效果可玩且可測；這是有標註的可觀察代理值，不是原版 byte-level 門檻。

其餘 captain/common 技能的 remake 消費端已完成；未來若補到 `VESA.COM` runtime，才追求
Diplomat 接受門檻與其餘技能逐值 oracle，不因這些未知回頭虛構公式。

## 七、測試

一手資料層(`internal/gamedata`):

- `officer_test.go`:`TestLeaderSkillTier`(**每技能 2 bit** 的解碼)/`TestLeaderMaintenanceCost`/
  `TestLeaderHireModifier`。
- `leader_skill_apply_test.go`:累加 vs 取最強的合成規則(含環保官負值取絕對值最大)。
- `leader_skill_names_test.go`:27 個技能一個不多一個不少、中文名不重複、
  「指揮官」只屬於 SKILL_COMMANDO、列舉順序專屬在前。

接線層:

- `officer_assignment_test.go`:指派／改派／解除、逐艦 Weaponry／Helmsman 與 JSON round-trip；
  `starlane_test.go`／`route_test.go`／`repair_test.go` 以真實指派驗證 Navigator／Engineer 消費端。

- `internal/shell/leader_test.go`:標籤退回路徑、**id 勝過顯示標籤**(英文模式不會靜默失效)、
  一位領袖多項技能、進階階比一般階強 50%。
- `internal/shell/leader_skill_test.go`:兩個貿易家不疊 vs 兩個科學家會疊(正對照)、
  分項百分比只動自己那一項 vs 士氣三項一起動(正對照)、科學官 ≠ 科學家。
- `internal/shell/repair_test.go`:工程師打贏才修 vs 打輸不修(正對照)、
  工程師必須是艦艇軍官、**自動修復元件不看勝負**(確認新的 `won` 參數沒有波及既有觸發)。
- `internal/shell/ground_invasion_test.go`:Commando 取最高階、無 Commando 領袖回 0。
- `cmd/moo2/herodatamercs_test.go`:2 bit 解碼(含「舊遮罩 `1<<6` 指到的是 SKILL_FAMOUS」
  這條回歸鎖)、進階階讀得出來、specialSkills 依領袖類型解讀、一人多技能且專屬在前、
  標籤翻譯但 id 不變、**類別通稱不可撞到真技能譯名**。
- `internal/shell/leader_effects_test.go`:Operations／Famous／Diplomat／Megawealth、
  Assassin、Spymaster／Telepath、Galactic Lore 的星圖與戰鬥分離、Ordnance／Security
  同時到達快速與格子戰術兩條路徑。
- `internal/shell/officer_assignment_test.go`:Fighter Pilot 取參戰艦隊最高值並傳入戰機資料。
- `internal/shell/economy_20_turn_test.go`:固定開局 20 回合 BC／人口／士氣／食物／工業／研究
  探針，輸出供平衡體感與後續調校使用。
