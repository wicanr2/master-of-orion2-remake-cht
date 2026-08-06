# 總缺口報告:原版 Orion2.exe vs remake

> 日期:2026-08-06。方法:解析原版執行檔內建的 8,590 個 Watcom 除錯符號
> (見 [`00-orion2-symbols.md`](00-orion2-symbols.md)),對照 remake 現行程式碼。
> **這是第一次能用原版二進位當基準做全面盤點**——先前只能靠手冊、攻略與 openorion2(純渲染殼)。

## 可信度分級(每條結論都標)

| 級別 | 意義 |
|---|---|
| **A 硬證** | 符號名 + 已驗證的位址映射;「原版有這個函式/資料表」是事實 |
| **B 推論** | 由符號名語意推斷用途,未反編驗證函式體 |
| **C 待驗** | 反編結果含 `JUMPOUT`(IDA 函式邊界錯誤),**不可採信**,需先修邊界 |

⚠ 本報告的「原版有 X」一律是 **A 級**(符號存在是事實);
「X 的具體行為/數值是 Y」則**尚未驗證**,列在下方待深挖清單。

---

## Part A — 畫面缺口(原版 53 個 vs remake 22 個)

### A-1 已對上(remake 有對應畫面)

| 原版 | remake | 備註 |
|---|---|---|
| `Main_Menu` / `Mainmenu_Load` | `menu` | ✅ 座標已對齊 |
| `Newgame` | `newGameSetup` | ⚠ 原版底排是 **PLAYERS**(對手數),remake 誤作 RACE 入口 |
| `Race_Selection` | `raceSelect` | ⚠ 版面左右相反(原版肖像左/2欄按鈕右) |
| `Racial_Option` | `customRace` | 自訂種族點數 |
| `Flag` | `nameFlag` | ⚠ 原版命名與旗幟是**兩個獨立畫面**,remake 合併且用色塊非旗幟圖 |
| `Main` / `Main_Main` / `Mini_Main` | `galaxy` | 原版有 mini 變體,remake 無 |
| `Colony_Summary` | `colonySummary` | |
| `Planet_Summary` / `Mini_Planet_Summary` | `planets` | |
| `Fleet` | `fleet` | |
| `Design` / `Darkened_Design` | `shipDesign` | |
| `Officer` | `officer` | |
| `Race` | `races` | |
| `Diplomacy` / `Diplomacy_Fade_In` | `diplomacy` | |
| `Main_Council` | `council` | |
| `Main_Combat` / `Super_Fast` | `tacticalCombat` | `Super_Fast` = 快速結算畫面 |
| `Turn_Summary` / `Dummy_Turn_Summary` | `turnSummary` | |

### A-2 ❌ 原版有、remake 完全沒有(依對玩家的影響排序)

| # | 原版畫面 | 影響 | 備註 |
|---|---|---|---|
| 1 | **`Colony`**(獨立殖民地畫面) | 高 | 原版 `Colony` 與 `Colony_Summary` 是**兩個不同畫面**;remake 只有總覽,缺單一殖民地的詳細管理畫面 |
| 2 | **`Event`** | 高 | 原版事件有**專屬畫面**;remake 把事件塞進回合摘要文字(與 issue #3「事件黃框」直接相關) |
| 3 | **`Tech_Review`** | 中高 | INFO 子畫面;remake 誤接成研究選擇跳板(issue #5-2 根因,**二進位證實**) |
| 4 | **`History`** | 中高 | INFO 子畫面,國力折線圖(issue #5-1);資料源是原版 module 122 的 `Record_History_` |
| 5 | **`Race_Stats`** | 中 | INFO 子畫面 |
| 6 | **`Reference_Main` / `_Category` / `_How_To`** | 中 | INFO 的內建說明書(3 個子畫面) |
| 7 | **`Command_Points`** | 中 | 原版有**專屬指揮點數畫面**;remake 只在星系右欄顯示一個數字 |
| 8 | **`Colony_Landing` / `Colony_Combat` / `Colony_Bombing`** | 中 | 地面戰/轟炸的**畫面**;remake 有引擎層解算但無畫面(只有文字結果) |
| 9 | **`Main_Antaran_Room`** | 中 | 安塔蘭房間;remake 的安塔蘭勝利是「艦隊列表一個文字按鈕」 |
| 10 | **`Hall_Of_Fame` / `Hi_Score`** | 低 | 名人堂/高分榜(主選單有入口) |
| 11 | **`Smack`** | 低 | Smacker 過場影片播放 |
| 12 | 多人連線 11 個畫面 | — | `Join_Net`/`MP_Setup`/`Hotseat`/`Modem_Setup`/`NullModem_Setup`/`Choose_Net_Plyrs`/`Choose_Multi_Net_Game`/`Generic_Net_Info`/`SendGet_Net_Info`/`Net_Next_Turn`/`Wait_For_*` — WORKLIST 已列獨立子專案 |

### A-3 remake 有、原版無獨立畫面

`research` / `researchChoice`(remake 自建的研究選擇)、`battleResult`、`info`(原版 INFO 是容器,子畫面才是實體)。
→ **不是錯**,但代表 remake 的研究流程與原版結構不同,值得反編確認原版怎麼進研究選擇。

---

## Part B — 系統缺口(原版 331 個原始碼模組)

`module` 欄位揭露原版的子系統分解。以下是**遊戲邏輯類**大模組與 remake 的對照:

| 原版模組 | 函式數 | 代表符號 | remake 對應 | 落差 |
|---|---|---|---|---|
| 74 | 106 | `N_Colonies_And_Outposts_At_Star_`、`N_Bldgs_` | `internal/shell` 散落 | 原版有**前哨站(Outpost)**概念,remake 無 |
| 48 | 86 | `Absolute_Location_`、`Contact_With_One_Colony_` | `shell/colonization.go` | 星圖拓樸/接觸判定較簡化 |
| **102** | **84** | `_minerals_per_mine`、`_climate_maintenance_modifiers`、`Colony_Officer_` | `gamedata/colony.go` 等 | **殖民地經濟核心**,權威數值全在此 |
| 14 | 83 | `Diplomacy_Screen_`、`Get_Main_Repulsive_Diplomacy_Choices_` | `cmd/moo2` diplomacy | 原版有完整外交選項樹 |
| 47 | 80 | `Design_Name_`、`Build_Saved_Ship_Array_` | `shell/shipnames.go` | |
| **122** | **73** | `Record_History_`、`Bill_Init_`、`Is_Ignoring_` | **無** | **歷史記錄系統**(History Graph 資料源) |
| 141 | 71 | `Load_Font_File_`、`Get_String_Width_` | `internal/uifont` | |
| 28 | 67 | `Get_Ship_Combat_Bonuses_`、`Init_Ship_Designs_` | `gamedata/combat.go` | |
| 58 | 67 | `Load_Officer_Picture_`、`Assert_Marooned_Leaders_` | `shell` 領袖 | 原版有 marooned leader 機制 |
| **15** | **57** | `Init_Events_`、`Check_For_Event_` | `shell` 事件(簡化) | **事件系統**,原版是完整子系統 |
| **27** | **56** | `Init_Diplomatic_Relations_`、`Diplomacy_Growth_`、`Change_Relations_` | `internal/diplomacy` | 原版關係演化更完整 |
| 20 | 54 | `Apply_Internal_Damage_`、`Repair_Combat_Ship_` | `gamedata/damage.go` | 原版有**內部艙損/維修** |
| 18 | 40 | `Safe_To_Fire_Sphere_Weapon_`、`Ai_Self_Destruct_Check_` | `shell/weapon_kind.go` | 原版戰鬥 AI 較深 |
| 19 | 42 | `Refresh_Combat_Screen_Full_` | `tacticalCombat` | |
| 65 | 37 | `Draw_Generic_Beam_`、`Draw_Ship_Burst_` | 戰鬥特效 | remake 特效較簡 |
| 138/314/316 | 141 | `Set_Music_File_`、AIL/Miles 驅動 | `internal/audio` | remake 直接播 PCM(已證等價) |
| 112/293 | 117 | Netmox / Hayes Modem | 無 | 多人連線 |

**最大的系統級缺口(A 級硬證)**:
1. **歷史記錄系統**(module 122,73 函式)——remake 完全沒有,History Graph 因此做不出來。
2. **事件系統**(module 15,57 函式)——remake 只有簡化隨機事件。
3. **前哨站(Outpost)**——原版到處都是 `..._And_Outposts_...`,remake 只有殖民地。
4. **艙損/維修**(module 20)——`Apply_Internal_Damage_`、`Repair_Combat_Ship_`。

---

## Part C — 資料表缺口(37 張權威數值表)

原版把數值放在具名資料表,**這些是比手冊更權威的數值來源**(手冊會簡化、會有筆誤——
專案先前已抓到手冊自身的 AMR 命中率與飛彈速度矛盾)。

| 原版資料表 | 用途 | remake 現況 |
|---|---|---|
| `_food_per_farmer_table` | 各氣候每農夫food | `gamedata` 用手冊值 |
| `_climate_modifier_table` / `_climate_roll_table` / `_normal_gal_climate_roll_table` / `_old_gal_climate_roll_table` | **星系生成**氣候骰表 | remake 自訂生成(oracle 對照發現星圖明顯比原版稀) |
| `_planet_size_table` / `_gravity_table` / `_spectral_class_table` / `_star_class_table` | 行星/恆星生成 | 同上 |
| `_minerals_per_mine` / `_minerals_extracted_table` | 礦產產出 | 手冊值 |
| `_climate_maintenance_modifiers` | 氣候維護成本 | **remake 無此概念** |
| `_ability_costs` | 自訂種族點數成本 | 用 patch1.5 config(已對) |
| `_personality_relation_modifiers` 等 **6 張** | **AI 性格行為** | remake 是手寫 `ai.Profile` |
| `_base_planet_values` / `_contextual_planet_values` / `_g_*` | **AI 行星估值** | remake AI 選星是「星圖索引順序」 |
| `_spy_bonuses` | 間諜加成 | remake 一律 0(標 TODO) |
| `_tech_research_level_values` | 科技研究等級 | `gamedata/techtree.go` |
| `_high/_low/_moderate_*_values`(9 張) | 疑似 AI 難度曲線 | 未知 |

**最大的數值缺口**:①星系生成整組骰表 ②AI 性格與行星估值 ③氣候維護成本。

---

## 優先序建議(依「對還原度的槓桿 ÷ 成本」)

### 第一梯:高槓桿、成本低(建議先做)
1. **反編 5 張星系/行星生成表 + 對應函式** → 修正「星圖密度比原版稀」(oracle 已觀察到)。表是靜態資料,dump 即得。
2. **反編殖民地經濟表**(`_food_per_farmer_table`、`_minerals_per_mine`、`_climate_maintenance_modifiers`)→ 用權威值取代手冊值,順帶解決母星開局態校準。
3. **INFO 5 個子畫面**(Tech_Review / History / Race_Stats / Turn_Summary / Reference)→ issue #5 的完整解;其中 History 需先建歷史記錄系統(module 122)。

### 第二梯:中槓桿
4. **`_personality_*` 6 張表 + AI 行星估值** → 把 remake 手寫 AI 換成原版行為模型。
5. **獨立 Colony 畫面 + Event 畫面** → 兩個原版有而 remake 缺的核心畫面。
6. **地面戰解算**(`Resolve_Ground_Combat_` / `Ground_Combat_Round_`)→ 取代目前沿用一代 1oom 的借用結構。

### 第三梯:補完整性
7. 前哨站、艙損/維修、Hall of Fame / Hi-Score、安塔蘭房間、Smacker 過場。
8. 多人連線(獨立子專案)。

---

## 待解的技術前提

**IDA 函式邊界修正**:Watcom 共用尾碼導致部分函式反編出 `JUMPOUT`(如 `Colony_Research_Per_Scientist_`、
`Missile_Speed_`)。深挖**函式**前要先解這個;但**資料表**不受影響,可以直接 dump ——
這也是把「第一梯」排在資料表的原因。
