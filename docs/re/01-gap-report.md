# 總缺口報告:原版 Orion2.exe vs remake

> 日期:2026-08-06(當日稍晚修訂)。方法:解析原版執行檔內建的 8,589 個 Watcom 除錯符號
> (見 [`00-orion2-symbols.md`](00-orion2-symbols.md)),對照 remake 現行程式碼。
> **這是第一次能用原版二進位當基準做全面盤點**——先前只能靠手冊、攻略與 openorion2(純渲染殼)。

## ⚠ 本報告初版的一處錯誤已修正

初版依據的符號表 parser 把記錄格式讀反(位址其實在名字**之前**),導致
**name↔addr 全部錯開一格**。修正細節與方法論教訓見
[`00-orion2-symbols.md`](00-orion2-symbols.md)。

對本報告的影響:

- **Part A(畫面清單)與 Part B(模組分群)基本不受影響**——它們只用到名字與模組歸屬,
  而同一 `.c` 模組的符號連續,位移一格只動到模組邊界的一兩個名字。重算後
  module 122(歷史記錄)74 個函式、module 15(事件)58 個,與初版的 73/57 差一。
- **Part C(資料表)整段重寫**:初版對照到的是**相鄰那張表**的內容,所以才會出現
  「`_food_per_farmer_table` 的值是 64..101,名字一定掛錯」這種結論。修正後
  該表就是手冊 p.59 的十氣候食物值,名字沒錯。

## 可信度分級(每條結論都標)

| 級別 | 意義 |
|---|---|
| **A 硬證** | 符號名 + 位址,且已用「讀它的函式怎麼索引」交叉確認語意 |
| **B 推論** | 由符號名語意推斷用途,未反編驗證 |
| **C 待驗** | 反編結果含 `JUMPOUT`(IDA 函式邊界錯誤),不可採信 |

⚠ 本報告的「原版有 X」一律是 **A 級**(符號存在是事實);
「X 的具體行為/數值是 Y」則需個別確認,未確認的列在待深挖清單。

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
| 1 | ~~**`Colony`**(獨立殖民地畫面)~~ | 高 | ✅ 2026-08-06 已建(`cmd/moo2/colonyscreen.go`),從總覽點殖民地名進入;含 7 格建造佇列。版面未對齊原版(無 COLONY.LBX 版面資料),結構與流程已對 |
| 2 | ~~**`Event`**~~ | 高 | ✅ 2026-08-06 已建(`cmd/moo2/eventscreen.go`),GNN 新聞快報樣式;事件表同步換成原版 36 種 |
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
| **15** | **71** | `Init_Events_`、`Check_For_Event_` | `shell/events.go` | 事件清單已對齊原版 36 種(16 種已實作,其餘缺子系統,見 `gamedata/events.go`) |
| **27** | **56** | `Init_Diplomatic_Relations_`、`Diplomacy_Growth_`、`Change_Relations_` | `internal/diplomacy` | 原版關係演化更完整 |
| 20 | 54 | `Apply_Internal_Damage_`、`Repair_Combat_Ship_` | `gamedata/damage.go` | 原版有**內部艙損/維修** |
| 18 | 40 | `Safe_To_Fire_Sphere_Weapon_`、`Ai_Self_Destruct_Check_` | `shell/weapon_kind.go` | 原版戰鬥 AI 較深 |
| 19 | 42 | `Refresh_Combat_Screen_Full_` | `tacticalCombat` | |
| 65 | 37 | `Draw_Generic_Beam_`、`Draw_Ship_Burst_` | 戰鬥特效 | remake 特效較簡 |
| 138/314/316 | 141 | `Set_Music_File_`、AIL/Miles 驅動 | `internal/audio` | remake 直接播 PCM(已證等價) |
| 112/293 | 117 | Netmox / Hayes Modem | 無 | 多人連線 |

**最大的系統級缺口(A 級硬證)**:
1. **歷史記錄系統**(module 122,73 函式)——remake 完全沒有,History Graph 因此做不出來。
2. ~~事件系統~~ → **2026-08-06 已對齊**:36 種事件表 + GNN 快報畫面,16 種已可忠實結算。
   剩餘 20 種缺的是各自的子系統(太空怪獸實體、超新星、時空異象、曲速漏斗…),已逐項記在
   `gamedata.RandomEvents` 的 `Needs` 欄。
3. **前哨站(Outpost)**——原版到處都是 `..._And_Outposts_...`,remake 只有殖民地。
4. **艙損/維修**(module 20)——`Apply_Internal_Damage_`、`Repair_Combat_Ship_`。

---

## Part C — 資料表(修正後重寫)

原版把數值放在具名資料表,**這些是比手冊更權威的數值來源**(手冊會簡化、會有筆誤——
專案先前已抓到手冊自身的 AMR 命中率與飛彈速度矛盾)。

### C-1 已釘死並接進 remake(2026-08-06)

星系/行星生成整組骰表已 dump、確認語意、寫進 `internal/gamedata/galaxygen.go`,
星系生成改用原版模型。逐格數值見 [`00-orion2-symbols.md`](00-orion2-symbols.md);
語意由「哪個原版函式怎麼索引它」釘死,不是靠名字猜:

| 表 | 讀它的原版函式 | 索引 |
|---|---|---|
| `_star_class_table` | `Generate_Spectral_Class_` | `[spectral*3 + age]` |
| `_planet_size_table` | `Generate_Size_` | d10 累計骰表 |
| `_class_to_group` | `Get_Planet_Group_` | `[spectral*5 + orbit]` |
| `_normal_gal` / `_old_gal_climate_roll_table` | `Generate_Climate_` | `[climate*4 + group]` |
| `_class_to_mineral` | `Generate_Mineral_Class_` | `[(d10-1)*6 + spectral]` |
| `_gravity_table` | `Generate_Gravity_Class_` | `[mineral*5 + size]` |
| `_class_to_num_satellites` | `Generate_Number_Of_Satellites_` | `[(d10-1)*6 + spectral]` |
| `_planet_max_farms` | `Generate_Max_Farms_` | `[size]` |
| `_food_per_farmer_table` | `Generate_Food_Per_Farmer_` | `[climate]` |

### C-2 已交叉驗證,remake 數值本來就對(不必改)

| 表 | 結論 |
|---|---|
| `_food_per_farmer_table` = `0 0 0 1 1 2 2 1 2 3` | 與手冊 p.59 逐格相同 |
| `_minerals_per_mine` = `1 2 3 5 8` | 與手冊礦產豐度五級相同 |
| `_planet_max_population` = `5 10 15 20 25` | 等於 remake 既有的 `(size+1)*5` |

**這三項可以撤銷「手冊可能簡化過」的存疑**——原版硬編值與手冊一致。

### C-3 仍是缺口

| 原版資料表 | 用途 | remake 現況 |
|---|---|---|
| ~~`_personality_*` **14 張**~~ | **AI 性格行為** | ✅ 2026-08-06 已接(`ai/personality_tables.go`)。實際是 **14 張**不是 6 張,每張 7 欄對應 Personality 0-6 |
| ~~`_base_planet_values` / `_g_*`~~ | **AI 行星估值** | ✅ 2026-08-06 已移植 4 個公式(`gamedata/ai_planet_value.go`):`Uncolonized_Planet_Worth_To_Player_`(選星)、`Proximity_Worth_To_Player_`(距離)、`Compute_Contextual_Planet_Values_`(星系協同)、`Colony_Worth_To_Player_`(已殖民星,供 AI 挑攻擊目標)。剩 `Enemy_Colony_Worth_To_Player_` |
| ~~`_climate_maintenance_modifiers`~~ | 氣候維護成本 | ✅ 語意已確認:索引 = 氣候,由 `Uncolonized_Planet_Worth_To_Player_` 以 `[planet.climate]` 讀。值 = `[50,25,0,25,0,0,0,0,0,0]`,已用於 AI 估值;**尚未接進殖民地實際維護費** |
| `_planet_max_mines` = `2 4 6 9 12` | 各大小礦場上限 | 已建表,**尚未接進生產**(remake 無礦場上限概念) |
| ~~`_planet_special` / `_planet_special_weighted_chance`~~ | 行星特殊物產(12 種,權重和 100) | ✅ 2026-08-06 已整套接進(`gamedata/planet_special.go` + `shell/discovery.go`),見下方 C-4 |
| ~~`_ranged_to_hit_penalty` / `_ranged_damage_penalty`~~ | 射程命中/傷害懲罰(各 9 個 word) | ✅ 兩張表 remake 早已有且逐格相同(手冊值);傷害衰減已於 2026-08-06 接進 `ResolveShotWithMods` |
| ~~`_orbit_to_satellite_type`~~ | 行星類別(氣態巨星/小行星帶/一般行星) | ✅ 2026-08-06 維度已釘死並接進生成器,見下方 C-5 |
| `_spy_bonuses` | 間諜加成 | remake 一律 0(標 TODO) |
| `_ability_costs` | 自訂種族點數成本 | 用 patch1.5 config(已對) |
| `_tech_research_level_values` | 科技研究等級 | `gamedata/techtree.go` |
| `_high/_low/_moderate_*_values`(9 張) | 疑似 AI 難度曲線 | 未知 |

**最大的剩餘數值缺口**:①間諜加成 ②氣候維護費接進殖民地開銷 ③礦場上限。

### C-5 `_orbit_to_satellite_type`:維度是這樣釘死的(2026-08-06)

表是 50 bytes,擺法有 10×5 與 5×10 兩種可能,光看數字分不出來。決定性的證據是**表裡唯一
的那個 `4`**:它落在 (roll 1, orbit 0),而 `Generate_Satellite_Type_` 處理 4 的特例分支
寫死 `bl == 1 && orbit == 0`。兩邊完全咬合,擺法不可能是別的。

| roll | 軌道0 | 軌道1 | 軌道2 | 軌道3 | 軌道4 |
|---|---|---|---|---|---|
| 0 | 小行星帶 | 小行星帶 | 小行星帶 | 小行星帶 | 小行星帶 |
| 1 | ★特例 | 小行星帶 | 小行星帶 | 小行星帶 | 氣態巨星 |
| 2 | 一般 | 氣態巨星 | 小行星帶 | 氣態巨星 | 氣態巨星 |
| 3-4 | 一般 | 一般 | 氣態巨星 | 氣態巨星 | 氣態巨星 |
| 5 | 一般 | 一般 | 一般 | 一般 | 氣態巨星 |
| 6-9 | 一般 | 一般 | 一般 | 一般 | 一般 |

內圈岩石、外圈氣態,roll 越大整個系統越宜居——分布本身就有物理直覺,是「沒讀錯」的旁證。

**★特例**(10% 機率)在原版會依恆星光譜寫進一個 ≥4 的類別碼(`spectral==0 ? 5 : spectral+4`),
openorion2 的 `enum PlanetType` 只定義 1-3,那些碼的語意目前無從確認,remake 一律當小行星帶
處理並另用一個 bool 標出來,不臆造。

### ⚠ Random_ 語意訂正(連鎖影響三處)

`Random_` @ 0x1247A0 回傳的是 **1..n**,不是 C 慣例的 0..n-1(LCG 取樣、拒絕超界值,
最後 `div bucket` 再 **`inc eax`**)。本報告與程式碼先前把它當成 `rand()%n`,連帶錯了三處:

1. **遠古文物送幾項科技**:`Random_(4)/4+1` 不是「恆為 1」,而是 **1 項、25% 機率 2 項**。
2. **蓄水池抽樣** `Random_(k)==1`:原版是**正確的** 1/k;先前記的「第一個候選永遠不會被選中」
   是誤讀(k=1 時 `Random_(1)` 必回 1)。
3. **失散殖民地在不可耕行星上的職務** `Random_(2) & 3`:是**工人或科學家**,不是農夫或工人。

訂正的方法是回頭讀 `Random_` 本身,而不是繼續從呼叫端推語意——`_orbit_to_satellite_type`
的 `roll = Random_(10) - 1` 落在 0..9(剛好對上 10 列)也反過來佐證了這件事。

**手冊後來獨立證實了訂正後的結論**:p.60 System Specials 講遠古文物時寫
「the first empire to discover the system gets **one or two** free technology advancements」
——「一或兩項」,正是 `Random_(4)/4+1` 在 1..n 語意下的值域。

### C-4 行星特殊物產:手冊沒給的數字,反組譯給了(2026-08-06)

這一項值得單獨記,因為它是**「手冊不足以還原、非讀執行檔不可」的典型案例**:
手冊只說太空殘骸/海盜藏寶的所得「is added to your treasury」,金額一個字都沒提。

| 特殊物產 | 效果 | 來源 |
|---|---|---|
| 太空殘骸(2) | 抵達星系 → 國庫 **+50 BC** | `Do_System_Discoveries_At_Star_` @ 0xE9927:`add dword [player+32h], 32h` |
| 海盜藏寶(3) | 抵達星系 → 國庫 **+100 BC** | 同上,`add … 64h` |
| 金礦(4) | 殖民地 +5 BC/回合 | 手冊(AI 估值加分 1280 佐證) |
| 寶石礦(5) | 殖民地 +10 BC/回合 | 手冊(AI 估值加分 2560 = 金礦兩倍,比例一致) |
| 原住民(6) | 殖民時**額外 3 個人口單位**、全為農夫、每農夫 +2 食物,之後該 special 消失 | `Make_New_Colony_Or_Outpost_` @ 0xE5EB3:迴圈 colony+0x10→0x1C(stride 4),`[colony+0Ah]=4`,`[planet+0Fh]=0` |
| 失散殖民地(7) | 抵達星系 → 就地生出殖民地,人口 = **min(該行星人口上限, 3)** | `cmp al,3 / jbe / mov byte [colony+0Ah],3` |
| 受困英雄(8) | 抵達星系 → 免費得一名領袖 | 手冊 + 原版該分支只設訊息碼 |
| 遠古文物(10) | 抵達星系 → **白送 1 項可研究科技(25% 機率 2 項)**;殖民後每科學家 5 研究 | 掃 204 個研究主題挑 `RSTATE_READY` 者蓄水池抽樣;送幾項 = `Random_(4)/4+1`,而 `Random_` 回 1..n(見下方「Random_ 語意訂正」) |

欄位偏移怎麼確定不是猜的:同一個函式裡的行星指標 stride = 0x11 = 17 bytes = openorion2
`struct Planet` 的大小,而 `[planet+0x0F]` 對上該結構的 `special`;殖民地那邊
`[colony+0x0A]`(population)、`[colony+0x0C]`(colonists[])、`[colony+0xE2]`(climate)
三個偏移同時對上 `struct Colony`;人口單位的位元佈局(race bits0-3 / loyalty 4-6 / job 7-8)
與 `Colonist::load` 逐位元相同;原住民寫進去的 race id 是 **9**,而 openorion2 的
`MAX_RACES = MAX_PLAYERS+2` 註解寫明「player races + androids + natives」——8 機器人、9 原住民。
最後,符號表裡獨立存在一個 `Planet_Has_Splinter_Colony_`,內容正好是 `[planet+0x0F] == 7`。

**六個獨立證據互相咬合,不是靠單一位址猜語意。**

---

## 優先序(依「對還原度的槓桿 ÷ 成本」)

### 原版畫面流程(反編 call graph,2026-08-06)

從 `.asm` 的 call 關係過濾出畫面主迴圈函式,得到原版的畫面跳轉:

```
Main_Menu_Screen_   → Do_Mainmenu_Load_Screen_
Newgame_Screen_     → Race_Selection_Screen_
Race_Selection_Screen_ → Racial_Option_Screen_ | Flag_Screen_
Racial_Option_Screen_  → Flag_Screen_
Main_Screen_        → Do_Colony_Screen_ / Get_Ship_Stack_For_Officers_Screen_ / Get_Star_Id_For_Officers_Screen_
Race_Screen_        → Diplomacy_Screen_ / Race_Report_Screen_
Planet_Colonization_In_Main_Screen_ → Colony_Landing_Screen_
Hotseat_Screen_ / Start_Net_Screen_ → Race_Selection_Screen_
```

remake 的新遊戲流程順序與此一致;`Main_Screen_ → Do_Colony_Screen_` 是先前缺的那一層。

### 第一梯 — 已完成(2026-08-06)
1. ~~星系/行星生成表~~ → **已接進 remake**(`gamedata/galaxygen.go`,commit `f8bbcbd`)。
   光譜/大小/氣候/礦產/重力/行星數全部改用原版骰表,並加了分布回歸測試。
2. ~~殖民地經濟表~~ → **已交叉驗證**:`_food_per_farmer_table`、`_minerals_per_mine`、
   `_planet_max_population` 三項與 remake 現值一致,不必改(見 C-2)。
   `_climate_maintenance_modifiers` 仍待確認讀取者。
3. ~~INFO 5 個子畫面~~ → **已實作**(commit `fade3f7`),含 module 122 歷史記錄系統。

### 第二梯 — 下一批
4. ~~射程命中/傷害懲罰~~ → **已完成**。查證發現兩張表 remake 早就有且與原版逐格相同
   (`combatRangeLevelPenaltyTable`、`damageDissipationPenaltyTable`);真正的缺口是
   **傷害衰減從未被呼叫**——同一發雷射在 1 格與 23 格外傷害一樣。已接進
   `ResolveShotWithMods`,順帶讓 NR(No Range Dissipation)改造第一次有實際效果。
5. ~~`_personality_*` 表 + AI 行星估值~~ → **都已完成**(2026-08-06),含後續補上的
   `Proximity_Worth_To_Player_`、`Compute_Contextual_Planet_Values_`、`Colony_Worth_To_Player_`。
   最後一個 `Enemy_Colony_Worth_To_Player_` 是「攻擊目標的**額外**加權」,
   remake 目前用 `AIColonyValue ÷ 距離` 代打(見 `shell/ai_attack.go`)。

   **順帶補上的缺口:AI 宣戰之後真的會打了。** 先前 AI 只會擴張與宣戰,關係掉到 -40、
   態勢寫著「戰爭」,玩家卻毫髮無傷——整局唯一的軍事壓力來自安塔蘭人腳本。現在 AI 會依
   `Colony_Worth_To_Player_` 的估值挑玩家最有價值的殖民地突襲(`shell/ai_attack.go`),
   造成人口/國庫/建築損失;玩家的艦隊、駐軍、軌道防禦建築會擋,擋不住也會消耗攻方戰力。
   ⚠ **「何時打、打贏怎樣」是 remake 的模型**(原版決策函式尚未反編),只有目標估值是原版公式。
   300 回合探針:56 次突襲、最早第 26 回合、經濟未崩(結束 BC 641)。
6. **一星多行星** → 原版每顆恆星 1–5 顆行星各佔一條軌道;remake 的 `Stars`/`Planets`
   索引一一對應是 UI/拓殖/AI 共同的假設,拆開是跨層改造(見 `genPlanets` 註解)。

   **已做的一半**(2026-08-06):恆星系現在會生成完整的軌道天體組成(用 C-5 的原版表),
   代表行星挑「最適合殖民的那一顆」,其餘記在 `Planet.SystemBodies` 供顯示與日後的前哨站。
   `Stars`/`Planets` 的一一對應**沒有動**,所以 UI/拓殖/AI 完全不受影響。
   氣態巨星/小行星帶因此第一次真的出現在遊戲裡,且不能直接殖民(手冊 p.55「colonies can
   only survive on a solid planet」)——探針:960 顆星裡 49 顆不可殖民。
   **剩下的一半**是讓玩家能分別殖民同一系統的多顆行星,那才是跨層改造。
7. ~~獨立 Colony 畫面 + Event 畫面~~ → **兩個都已完成**(2026-08-06)。
8. **地面戰解算**(`Resolve_Ground_Combat_` / `Ground_Combat_Round_`)→ 取代目前沿用一代 1oom 的借用結構。

### 第三梯:補完整性
9. ~~前哨站~~ → **已完成**(2026-08-06,`internal/shell/outpost.go`):前哨船可建造、
   可在氣態巨星/小行星帶/一般行星建立軍事前哨站,前哨站是掃描站(併進 detection.go 的偵測源)
   且**沒有人口與產出**(手冊 p.85「produces nothing」,故不進 PlayerColonies);之後在同一顆星
   建殖民地時前哨站改建為海軍陸戰隊營(手冊逐字)。順帶補上**殖民船也終於可以建造**——
   先前開局送一艘、用掉就再也不能擴張。
   ⚠ 未兌現的一半:「延伸艦艇航程 / 加油站」(手冊 p.119/p.133)——remake 的 SendFleet 沒有
   航程上限這個概念,沒有可套用的機制,不臆造。
   ⚠ 手冊 p.50 的「前哨站升級成可住人殖民地」科技仍未做(需要對應科技旗標)。
10. ~~太空怪獸~~ → **已完成**(2026-08-06,`internal/shell/monster.go` + `gamedata/space_monster.go`)。
    這一項有個值得記的地方:**它一直被程式碼引用著卻不存在**——colonization.go 檔頭抄的手冊
    原文就寫著殖民船要「as long as all space monsters and enemy ships have been cleared from
    that planet's system」,但那個 gate 從來沒有東西可擋。

    - 五種怪獸的名字來自執行檔字串表(0x1F742C 起連續:Guardian / Amoeba / Dragon / Hydra /
      Crystal),對應五個 `Load_*_Ship_Design_` 函式與 `_monster_names` @ 0x199266
    - 傷害數字來自手冊 p.114「Monster Traits」逐字(水晶射線 40-80、電漿吐息必中上限 60、
      相位眼 5-10、龍焰必中上限 300 每格 -15、腐蝕黏液 25-50 每回合 -5)
    - 生成規則來自手冊 p.60 逐字:「a system with a monster will always have another special
      — that's usually what drew the monster there in the first place」,已落地成「擺怪獸時
      強制補一個特殊物產」
    - ⚠ 怪獸的**結構值**與挑選機率是 remake 估值(手冊只給武器傷害);原版的數量是新遊戲
      設定(`_user_wants_n_space_monsters` @ 0x19A006),remake 先用固定密度
    - 順帶依手冊 p.119「Support ships … **do not fight**」把殖民船/前哨船排除在戰鬥火力之外
      ——先前它們會以最低戰力混進戰列
11. ~~持續型隨機事件~~ → **已完成**(2026-08-06,`internal/shell/events_persistent.go`)。
    先前 9 個事件卡在「缺子系統」,真正缺的是同一個東西:remake 只有「單次結算」的事件模型,
    **沒有任何跨回合的事件狀態**。補上那個模型之後,手冊 p.180-181 就直接是規格書:

    | 事件 | 手冊給的數字 |
    |---|---|
    | 超新星(24) | ≥200 回合觸發、倒數 6-14 回合、系統研究點全投入搶救、失敗則全滅 + 行星變輻射 |
    | 時空異象(25) | 星系凍結:不生產不成長,也不吃食物不繳維護費;6 回合後每回合 5% 結束 |
    | 超空間獸(26) | 航行中的艦隊有機率損失一艘;6 回合後每回合 5% 離開 |
    | 蟲洞(28) | 航行中的艦隊「in a single turn」抵達 |
    | 怪獸入侵(19-23) | 變形蟲 ≥100、太空鰻 ≥150、水晶 ≥200、九頭蛇 ≥250、巨龍 ≥300 |

    超新星那條的張力也是手冊寫死的:「if the emperor doesn't accelerate the colony's research
    efforts, the colonies will discover the solution **one turn too late**」——remake 據此把
    需求量設成「系統自然產出 × (倒數+1)」,讓「什麼都不做剛好差一回合」成立。

    ⚠ 太空鰻是**近似**:手冊說牠「never attack colonies or outposts」、只封鎖,且 30 回合後
    會分裂(最多 4 隻)。remake 用盤據型怪獸代打「封鎖」那一半,另兩半未建模,已標在事件表的
    `Needs` 欄。

    **這一批順帶翻出一個過期斷言**:`advanceConquestVictory` 的註解寫著「remake 沒有任何機制
    會讓 PlayerColonies 完全清空,故玩家戰敗這個分支不可達」。超新星讓它可達了,而
    `CheckExtermination`(只剩一方存活)在「玩家死光但還有三個 AI」時回 false——400 回合探針
    實測到玩家 0 殖民地、遊戲卻繼續空轉。已補 `advancePlayerDefeat`(手冊 p.184 計分段明講
    「If an empire is eliminated by a random event」,帝國被隨機事件消滅是原版就有的概念)。
12. ~~Hall of Fame / Hi-Score~~ → **已完成**(2026-08-07,`gamedata/score.go` + `shell/score.go`
    + `cmd/moo2/hiscore.go`)。手冊 p.184 列了八條計分因素但**一個數字都沒給**;原版 module 60
    的一整組 `Get_*_Score_` 函式每個都短到能逐指令讀完,八條的係數全在裡面:

    | 項目 | 公式(反組譯) | 手冊對應的那句話 |
    |---|---|---|
    | 時間/星圖/種族數 | `nPlayers × (20×(星圖大小+1) + 80) − 已過回合數`;人口 0 則整項 0 | 「越快贏分越高」「星圖越大分越高」「種族越多分越高」三句合一 |
    | 人口 | 自己所有殖民地人口總和 | 「total number of population units … added to your score」 |
    | 俘虜人口 | `俘虜 × 2 ÷ (星圖大小 + 1)` | 「premium … **higher in smaller galaxies**」——除數就是那句話 |
    | 科技 | `3 × 已知主題 + 5 × Hyper-Advanced 等級` | 「First level Hyper-Advanced … worth **more** points than normal ones」——5 > 3 |
    | 殲滅種族 | 每族 50 | 「a boost」 |
    | 獵戶座 | 100 | 「a big chunk of points」 |
    | 議會勝利 | 100 | 「a substantial addition」 |
    | 安塔蘭勝利 | 250 | 「the **biggest** point bonus of all」 |

    **相對大小完全對上手冊的形容詞排序**:250 > 100 = 100 > 50。另有兩個順帶的交叉驗證:
    科技分掃的是 `player+0xC4` 起 0x53(83)長的研究主題陣列——0xC4 與
    `Do_System_Discoveries_At_Star_` 讀遠古文物時用的是同一個偏移,83 也與 remake 既有的
    研究主題數相同;時間分用 `word[0x192FD8] − 0x88B8` 算已過回合,0x88B8 = 35000 =
    星曆 3500.0 ×10,正是遊戲起始星曆。

    ⚠ remake 側的落差:①獵戶座系統還沒做,該項恆 0 ②「殲滅種族」原版有逐玩家的
    「這個玩家滅了誰」陣列(player+0x1F2),remake 沒追蹤是誰滅的,目前全算給玩家,
    AI 互滅時會高估——標明,不假裝精確。
13. ~~艙損/維修~~ → **已完成**(2026-08-07,`internal/shell/repair.go`)。這一項的起點不是
    「補一個修復系統」,而是先發現 **remake 根本沒有艦艇損傷這個概念**——一艘船不是完好就是
    被擊沉,打完慘勝的仗倖存艦跟出港時一模一樣。於是「自動修復」這個元件
    (`SpecialOptions` 裡的 `{"自動修復", …, TECH_AUTOMATED_REPAIR_UNIT}`)從加進來那天起
    就沒有任何效果:沒有損傷可修。

    | 規則 | 來源 |
    |---|---|
    | 停在自家據點(殖民地/前哨站)→ **完全修復** | 反組譯 `Repair_Ships_At_Colonies_` @ 0x580F5 直接呼叫 `Repair_Ship_Full_` @ 0x581F3——是完全修復,不是逐回合慢慢修 |
    | 自動修復元件:戰鬥中每回合修 **20%** 結構損傷 | 手冊 p.82 逐字 |
    | 自動修復元件 / 進階損害管制:**戰後完全修復** | 手冊 p.82 / p.80 逐字 |
    | 機械化種族:戰鬥中 10%/回合(常數已備,無呼叫端) | 手冊 p.25 逐字;remake 沒有種族特質欄位可掛 |

    **交叉驗證**:openorion2 讀存檔的 `struct Ship`(`gamestate.h:1268`)把「原版記了幾份損傷」
    逐欄位寫死了——

    ```c
    uint8_t  shieldDamage, driveDamage;   // percent
    uint8_t  computerDamage, crewLevel;
    uint8_t  damagedSpecials[(MAX_SHIP_SPECIALS+7)/8];
    uint16_t armorDamage, structureDamage;
    ```

    原版每艘船記**六份**(護盾/引擎/電腦/逐元件旗標/裝甲/結構),remake 只有
    `structureDamage` 這一份。`ships.cpp:1060` 的 `isSpecialDamaged(i)` 把壞掉的元件名字
    畫成損壞色,那是逐元件損傷唯一的 UI 出口。

    ⚠ 誠實留白:①**內部系統損傷**(引擎/武器/護盾/電腦/各元件的損壞度)——remake 的戰鬥是
    艦級抽象,沒有逐系統狀態,手冊那句「systems damage 10%/5% per round」無處可套,只做結構
    損傷這一半;原版對應的 `Apply_Internal_Damage_` @ 0x35251 是個依傷害類型分十幾條分支、
    操作艦艇結構 +0x29/+0xC2/+0x134 等欄位的大函式,要接它得先有逐系統模型 ②裝甲與結構在
    remake 合一 ③「進階損害管制」科技還不在科技樹裡,`playerHasAdvancedDamageControl` 恆
    false(標明是缺科技,不是規則沒接)。

    **UI 出口**:艦隊列表每艘船後面加「損傷 N%」(輕傷琥珀 / ≥50% 紅),完好的船不畫。
    沒有這一欄的話這個系統對玩家等於不存在——打完仗有傷卻沒地方看得到。
    截圖廊也補了艦隊列表這一張(`11_fleet.png`,共 15 張)。

    順帶記一個踩到的坑:截圖廊原本在 `galleryVictoryTick`(t28)注入損傷,而 t29 按了
    「結束回合」,截出來一艘傷都沒有——那不是顯示壞了,是 `EndTurn → advanceShipRepair`
    正常運作:艦隊開局就停在母星,照 `Repair_Ships_At_Colonies_` 的規則被完全修復。
    注入點改到最後一次結束回合之後(t40)才驗得到。

    順帶修掉一個對映錯誤:損傷寫回原本用外部平行陣列對映「第 k 個參戰艦 → 第 k 艘船」,
    但戰鬥中有人陣亡後這個對映就錯位,會把 A 船的傷記到 B 船上。改成把船索引放進
    `combatant` 結構裡,過濾陣亡者時整個 struct 複製、索引跟著倖存者走。
14. 安塔蘭房間、Smacker 過場。
   ~~行星特殊物產~~ → **已完成**(2026-08-06,見 C-4):12 種權重表 + 抵達發現(殘骸/藏寶/
   失散殖民地/受困英雄/遠古文物)+ 殖民效果(原住民人口、金礦寶石收入、文物研究),
   接進 `advanceFleet` 抵達點與快報畫面。
15. 多人連線(獨立子專案)。
