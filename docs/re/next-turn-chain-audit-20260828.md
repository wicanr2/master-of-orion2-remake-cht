# 回合主鏈固定順序稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`tools/ida/audit_next_turn_chain.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部符號導航：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`。
  名稱只供導航；函式邊界、call operand、bytes 與順序以 IDA 資料庫為證據。
- 可重生證據：
  [`evidence/next-turn-chain-ida-20260828.json`](evidence/next-turn-chain-ida-20260828.json)。
- 本輪只封閉主鏈拓樸與邊界，不把子函式名稱本身升格成其內部玩法已證實。

## 已證實的外層控制流

`Next_Turn_Calc_ / sub_136B3 @ 0x136B3..0x13822` 有 **52 個直接 call 指令**。主體除三個
條件分支與一個逐玩家迴圈外，依下表固定順序執行；沒有 switch 或間接 call。

| # | callsite | raw target | 外部符號導航名稱 | 主鏈角色／證據狀態 |
|---:|---:|---:|---|---|
| 1 | `0x136C7` | `0x1191CA` | `Assign_Auto_Function_` | 低模式 callback 邊界；平台／UI 內部排除 |
| 2 | `0x136CC` | `0xEEB3A` | `Allocate_Msg_Slots_` | 回合訊息槽初始化 |
| 3 | `0x136E3` | `0x63D92` | `Antaran_Invasion_Check_` | 安塔蘭週期；具獨立 RE 文件 |
| 4 | `0x136E8` | `0xDCBB0` | `Apply_Evolutionary_Upgrades_` | 進化突變升級 |
| 5 | `0x136ED` | `0xD3D34` | `Compute_AI_Data_` | AI／方向國力與殖民資料 |
| 6 | `0x136F2` | `0x1FFED` | `Resolve_Delayed_Diplomacy_Orders_` | 延後外交命令 |
| 7 | `0x136F7` | `0x1FD80` | `Set_Opportunity_Attacks_` | 外交／艦隊攻擊機會 producer |
| 8 | `0x136FC` | `0x5090C` | `Clear_Diplomacy_Messages_` | 清回合外交訊息 |
| 9 | `0x13701` | `0x252A7` | `NPC_Diplomacy_` | NPC 談判、宣戰與停戰 |
| 10 | `0x13706` | `0x4DD6B` | `Diplomacy_Growth_` | 關係與條約逐回合變化 |
| 11 | `0x1370B` | `0xD7439` | `Do_AI_Leaders_` | AI 領袖管理 |
| 12 | `0x13710` | `0xCFCB6` | `Compute_Empire_Building_Needs_` | AI 帝國建造需求摘要 |
| 13 | `0x13715` | `0xD6F67` | `All_Colony_AI_` | AI 逐殖民地決策 |
| 14 | `0x1371A` | `0xDBB29` | `Move_All_AI_` | AI 移動 |
| 15 | `0x1371F` | `0xE67F6` | `All_AI_Colonize_` | AI 殖民 |
| 16 | `0x13724` | `0xDCA69` | `All_AI_Tech_Select_` | AI 研究選擇 |
| 17 | `0x13729` | `0xEDF92` | `Make_Scrap_Ships_Dead_` | 報廢艦狀態套用 |
| 18 | `0x1372E` | `0xFD81C` | `Initialize_Reports_` | 本回合報告初始化 |
| 19 | `0x13733` | `0x101E77` | `Process_Trade_And_Research_Agreements_` | 貿易／研究協議收益 |
| 20 | `0x13738` | `0xFFEEA` | `Move_All_Ships_Toward_Stars_` | 全艦星際移動 |
| 21 | `0x1373D` | `0x10192B` | `Resolve_Spies_` | 間諜回合解算 |
| 22 | `0x13742` | `0xE4F49` | `Apply_All_Player_Changes_` | 帝國經濟、研究與衍生值套用 |
| 23 | `0x13747` | `0xE3FDC` | `Apply_All_Colony_Changes_` | 殖民地變更與同化套用 |
| 24 | `0x1374C` | `0xE4DC9` | `Do_Surrenders_` | 延後投降資產移交 |
| 25 | `0x13751` | `0xD574D` | `Deallocate_AI_Data_` | AI 暫存資料生命週期；無獨立玩法語意 |
| 26 | `0x13756` | `0xE9D62` | `Search_For_Battles_` | 搜尋及處理戰鬥 |
| 27 | `0x1375B` | `0x2230A` | `Determine_Event_` | 新事件抽選／建立 |
| 28 | `0x13760` | `0x14A27` | `Do_All_Ships_XP_Check_` | 全艦艦員經驗 |
| 29 | `0x13765` | `0x206A2` | `Event_Twiddle_` | 持續事件效果 consumer |
| 30 | `0x1376A` | `0xED8BF` | `Lose_Out_Of_Range_Ships_` | 超出補給範圍艦艇損失 |
| 31 | `0x1376F` | `0xE2B31` | `Do_Colony_Calculations_` | 殖民地計算第一遍 |
| 32 | `0x13774` | `0xE5097` | `Compute_Blockades_` | 封鎖重算 |
| 33 | `0x13779` | `0xFF212` | `Move_Settlers_` | 殖民人口遷移 |
| 34 | `0x1377E` | `0xE2B31` | `Do_Colony_Calculations_` | 同函式第二遍；不是誤重複 |
| 35 | `0x13783` | `0x10011B` | `Allocate_New_Ship_Slots_` | 艦艇記錄槽整理 |
| 36 | `0x13788` | `0x92FDA` | `Check_For_Officer_Level_` | 領袖升級 |
| 37 | `0x1378D` | `0x934CF` | `Decrement_Officer_ETA_` | 領袖 ETA 遞減 |
| 38 | `0x13792` | `0xED44A` | `Check_All_Rebellions_` | 叛亂檢查 |
| 39 | `0x13797` | `0x580F5` | `Repair_Ships_At_Colonies_` | 殖民地艦艇維修 |
| 40 | `0x1379C` | `0xE64F4` | `Allocate_New_Colony_Slots_` | 殖民地記錄槽整理 |
| 41 | `0x137A1` | `0xEB192` | `Compute_Contacts_` | 接觸矩陣重算 |
| 42 | `0x137A6` | `0x50B57` | `Determine_First_Contacts_` | 首次接觸事件 |
| 43 | `0x137AB` | `0x10087D` | `Reset_Needless_Ignore_Combat_Ships_Flags_` | 艦隊戰鬥忽略旗標整理 |
| 44 | `0x137B0` | `0x4EB06` | `Determine_Diplomacy_Messages_` | 外交訊息建立 |
| 45 | `0x137B5` | `0x4DAB2` | `End_Of_Turn_Diplomacy_Adjustments_` | 外交回合尾 producer |
| 46 | `0x137CC` | `0x168AF` | `Check_For_Council_Meeting_` | 議會排程；條件呼叫 |
| 47 | `0x137D8` | `0x97A66` | `Random_Officer_Check_` | 對每個 player slot 執行一次 |
| 48 | `0x137E9` | `0xEF827` | `Add_Ongoing_Event_Msgs_` | 持續事件訊息 |
| 49 | `0x137EE` | `0x110B5C` | `Check_Release_Version_` | build/runtime 邊界；只保留回傳 gate |
| 50 | `0x137F8` | `0x92C87` | `Assert_Marooned_Leaders_` | 上項回傳 0 才呼叫 |
| 51 | `0x137FD` | `0x10208A` | `Record_History_` | 歷史記錄；星曆增加前 |
| 52 | `0x1381B` | `0x123E6C` | `Set_Mouse_List_` | UI 輸入邊界；平台內部排除 |

## 三個條件與一個迴圈

1. `0x136B4..0x136C7`：`byte_199F3A <= 1` 才設定低模式 auto callback。只記錄 callback
   選擇；callback／輸入框架內部不是 remake 玩法 RE。
2. `0x136D1..0x136E3`：只有 `byte_199CAF != 0 && byte_1991B0 == 0` 才做安塔蘭入侵檢查。
   兩欄的玩家語意已有安塔蘭專題文件；此處證實它位於所有 AI／外交處理之前。
3. `0x137BA..0x137CC`：只有 `byte_199F3A == 0 && byte_1ACE74 == 0` 才檢查議會；這是模式／
   暫態 gate，不改變已閉合的議會 25 回合排程公式。
4. `0x137D1..0x137E7`：以 `word_199998` 為上限，由 player slot 0 遞增，每槽呼叫一次
   `Random_Officer_Check_`。因此招募在外交、事件、殖民地、叛亂與修理之後執行。

`Record_History_` 返回後才在 `0x13802` 增加 `dword_192FD8`，所以所有本回合玩法變更與歷史
取樣都使用增加前的星曆。最後的 `Set_Mouse_List_` 只在 `byte_19C196 == 0` 呼叫，屬 UI 邊界。

## 完成邊界

- **已證實**：52 個直接 call、每個 callsite／target bytes、固定順序、兩次殖民地計算、三個
  條件 gate、逐玩家領袖招募迴圈，以及歷史記錄早於星曆遞增。
- **強推論／導航**：外部符號提供的函式名稱；只有另有專題文件閉合者才能當作已證實語意。
- **排除範圍**：`Assign_Auto_Function_`、`Check_Release_Version_` 與 `Set_Mouse_List_` 的
  compiler／runtime／平台內部不納入 remake；只保留會改變玩家可見流程的 callback、回傳與
  呼叫條件。`Deallocate_AI_Data_` 只作生命週期邊界，不因配置內部增加玩法分母。
- **仍待 RE**：不是再追主鏈本體，而是 parity matrix 各子系統列尚未閉合的 producer、規則、
  consumer 與資料模型。尤其 `Do_Colony_Calculations_` 兩遍之間的封鎖／遷移影響、完整間諜
  外圈、外交回應、地面戰及戰機傷害仍各自保留為獨立缺口。
