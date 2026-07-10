# 領袖/軍官技能系統(2026-07-11 接線)

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
描述(如 Spiritual Leader「Raises morale」、Commando「ground combat strength」),精確數字/
換算公式手冊沒給——與既有 `internal/gamedata/morale.go`、`ground.go` 檔尾 TODO 清單的判斷標準
一致(手冊有精確數字才移植,只有文字定性描述不臆造)。

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
  **⚠ remake 的 `shell.Ship` 目前沒有軍官指派欄位、也沒有任何戰鬥解算迴圈會呼叫這兩個新函式**——
  這只是引擎層可用的公式,真正接進「玩家指派軍官上艦 → 戰鬥時生效」需要先有戰鬥畫面/艦艇軍官
  指派 UI,這兩者 remake 都還沒有(見 `docs/HONEST-STATUS.md`)。屬於「公式已備妥,等系統」。
- `internal/engine/leader.go`:`HireLeader(currentBC, cost int) (newBC int, ok bool)`——最小雇用
  金流機制(BC 夠不夠、扣款),供未來招募畫面呼叫。**領袖狀態轉換(ForHire→Working 等)不在本輪
  範圍**,`demoLeaders` 既有領袖視為已受雇,不需要走這個函式。

### 4.3 shell 接線(`internal/shell/session.go`)

- `Leader` 新增 `Tier int` 欄位(demoLeaders 皆保守設 1=一般技能,非 HERODATA 真實資料,不臆造
  「進階」)。
- `leaderSkillIDByName`:demoLeaders 中文標籤 → `gamedata` 技能 id 的映射表,只收 3 個語意清楚
  的:「科學家」→`SKILL_RESEARCHER`、「貿易家」→`SKILL_TRADER`、「工程師」→`SKILL_ENGINEER`。
- `leaderDisplayLevelToExpLevel(level int) int`:demoLeaders 既有 `Level`(1..5 顯示等級)換算
  `LeaderSkillBonus` 用的 `expLevel`(0..4),採 `Level-1` 夾在 `[0,4]`。
- `applyLeaderColonyBonuses(leaders []Leader, colony *engine.ColonyState)`:殖民地領袖
  (`Ship=false`)套到指定殖民地——`SKILL_RESEARCHER`(固定值)→`ColonyState.FlatResearch`,
  `SKILL_TRADER`(%)→`ColonyState.IncomeBonusPercent`(與太空港/證交所同一欄位,可疊加)。
  `NewDemoSession` 建完 `PlayerColonies[0]` 後呼叫一次,套到母星(demo 唯一殖民地)。

## 五、映射待人工定案(不確定,列出讓使用者確認)

- **「指揮官」(漢尼拔,Ship=true)刻意未收錄進 `leaderSkillIDByName`**:這是 demo 資料自訂的
  中文頭銜字,不是從 openorion2 技能表或手冊衍生的標籤——技能表裡沒有字面叫「Commander」的
  技能,任何映射都是我方猜測。候選:
  - `SKILL_WEAPONRY`(語意最接近「指揮官帶頭衝鋒」,且 remake 已有 `ShipBeamAttackWithOfficer`
    可承接)。
  - `SKILL_COMMANDO`(地面戰,但基礎加成手冊未給精確數字,見 `gamedata/ground.go` 檔尾 TODO,
    即使映射對了也套不進任何現有欄位)。
  - `SKILL_SECURITY`(艦艇安防,openorion2 沒有效果呼叫端)。
  - 目前保守選擇:**都不套**,漢尼拔在遊戲裡沒有任何技能加成,直到使用者定案。
- **「工程師」(圖靈,Ship=true)映射到 `SKILL_ENGINEER` 語意清楚,但 remake 沒有艦艇維修系統
  接收這個 %加成**(真實效果是「每回合修復艦體損傷的百分比」,remake 目前的 `save.Ship` 有
  損傷欄位但沒有「每回合自動維修」的引擎邏輯)——技能 id 對應沒有疑義,純粹是承接系統未建,
  標 TODO,不是映射問題。

## 六、TODO(手冊/openorion2 皆無精確數字或 remake 尚無承接系統,誠實不做)

- `SKILL_SCIENCE_LEADER`/`SKILL_FINANCIAL_LEADER`/`SKILL_LABOR_LEADER`/`SKILL_FARMING_LEADER`/
  `SKILL_ENVIRONMENTALIST`/`SKILL_MEDICINE`/`SKILL_INSTRUCTOR`:openorion2 無呼叫端,demoLeaders
  目前也沒有領袖標成這些技能名稱,暫不建映射(未來若要加同類「XX 領導」demo 角色,可比照
  Researcher/Trader 的模式:查 `skillFormatStrings` 決定固定值/%,再決定接哪個 `ColonyState`
  欄位)。
- `SKILL_SPIRITUAL_LEADER`/`SKILL_TACTICS`:手冊已明講「無精確數字/未實作」,見
  `internal/gamedata/morale.go`、`ground.go` 檔尾既有 TODO,不重複。
- `SKILL_ASSASSIN`/`SKILL_COMMANDO`/`SKILL_DIPLOMAT`/`SKILL_OPERATIONS`/`SKILL_SPYMASTER`/
  `SKILL_TELEPATH`:間諜/外交/地面戰系統 remake 尚未建（或該技能對應的量沒有精確公式），
  比照任務邊界「間諜主管/心靈感應/外交等需要未建系統的技能一律標 TODO」處理。
- `SKILL_FIGHTER_PILOT`/`SKILL_GALACTIC_LORE`/`SKILL_ORDNANCE`/`SKILL_SECURITY`/
  `SKILL_NAVIGATOR`:openorion2 無效果呼叫端(Navigator 只在 `skillBonus` 特例判斷式出現,沒有
  任何函式讀取它的回傳值),不建模。
- 艦艇軍官指派 UI + 戰鬥畫面:`ShipBeamAttackWithOfficer`/`ShipBeamDefenseWithOfficer` 已就緒,
  等 remake 有「軍官上艦」與「戰鬥解算」兩個系統才能真正串起來。
- 招募 UI:`HireLeader` 已就緒,等軍官列表畫面做「雇用」按鈕才能呼叫。

## 七、測試

- `internal/gamedata/officer_test.go`:`TestLeaderSkillTier`/`TestLeaderMaintenanceCost`/
  `TestLeaderHireModifier`(新增,3 個函式各數例)。
- `internal/engine/ship_test.go`:`TestShipBeamAttackWithOfficer`(含 `_NoOfficer` 對照組)、
  `TestShipBeamDefenseWithOfficer`。
- `internal/engine/leader_test.go`:`TestHireLeader`(5 例:足夠/剛好/不足/cost0/cost負)。
- `internal/shell/leader_test.go`:`TestApplyLeaderColonyBonuses_Researcher`/`_Trader`/
  `_ShipOfficerSkipped`/`_UnmappedSkillSkipped`/`_NoLeadersNoop`、
  `TestLeaderDisplayLevelToExpLevel`、`TestLeaderSkillIDByNameMapping`。
- `internal/shell/session_test.go`:`TestGameSessionEndTurn` 期望值 30→55 更新(母星 30 基礎研究
  + 馮·諾伊曼科學家技能 +25),註解記錄換算過程,避免日後誤以為是 regression。

全數跑過 `go build ./...`/`go vet ./...`/`go test ./internal/... ./cmd/...`(docker
`moo2-ebiten:latest`),除既有已知的 `internal/uifont` X11 環境限制外全綠。
