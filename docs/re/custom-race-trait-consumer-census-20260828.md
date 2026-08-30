# 客製種族 31-byte 特性消費端普查（2026-08-28）

## 結論與限制

本輪用 IDA Pro 9.4 對整個函式空間搜尋可能以玩家陣列基址讀寫
`player+0x89F..player+0x8BD` 的指令，不再只檢查預先挑選的少數函式。結果確認
runtime 是 **31 個 signed byte**；`POOR_HOMEWORLD` 不是第 32 byte，而是 index 15
（`player+0x8AE`）的 `-1` 編碼。

這是一份 **direct-operand census**，不是「31 項行為已全部閉合」的聲明：

- 已證實：31-byte mapping 與各已人工審查切片的 raw operand；JSON 保存函式邊界、指令 bytes、
  原始運算元、前後視窗與精確位址。
- 候選：全庫共 137 個 IDA owner 位址。Watcom distant tail chunk 與錯誤函式邊界會讓 owner／
  線性符號範圍不一致，故未人工追完 register provenance 的列不能只因命中 offset 就升格。
- 強推論：`symbols_fixed.tsv` 的 exact-address 名稱可用來判斷子系統，但名稱本身不證明公式。
- 未知：經指標傳遞、複製後再讀、packed colonist race 或衍生欄位的間接消費端，不會被本掃描
  自動捕捉，仍須由 producer／consumer 資料流補齊。
- 本輪沒有修改 remake，也沒有把 compiler helper、C runtime 或 Windows API 算入玩法分母。

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA database SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 外部符號表 SHA-256：`f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`
- 工具：IDA Pro 9.4，IDA linear address，DOS/4GW LE object #1
- 可重生腳本：`tools/ida/audit_custom_race_trait_consumers.py`
- 完整機器證據：`evidence/custom-race-trait-consumers-ida-20260828.json`

匯出契約刻意同時保留 IDA 原始 `sub_xxx` 與受雜湊約束的外部符號名稱；後者不能覆蓋前者。

## 31 格覆蓋矩陣

下表的「直接消費端」只列代表性原始符號；完整清單與每一條指令在 JSON。
「閉合狀態」指原版玩家可見資料流，不指 remake 是否已有相似效果。

| idx / offset | 導覽名稱 | 已定位的代表性直接消費端 | 閉合狀態 |
|---|---|---|---|
| 0 / `+0x89F` | Government | `Init_Homeworld_Colony2_`、`Cost_Reduction_For_Govt_Type_`、`Colony_Morale_`、`Apply_Assimilation_`、`Compute_Player_Maintenance_`、`Compute_Spy_Bonuses_` | 主要玩家鏈已閉合；AI 評分、產品合法性 raw 表與 occupation policy 仍分切片 |
| 1 / `+0x8A0` | Population | `Init_Homeworld_Colony2_`、`Colony_Pop_Grows_`、`Uncolonized_Planet_Worth_To_Player_`、`Init_NPC_Personalities_Objectives_Themes_`、`Calc_Tech_Value_` | 玩家成長 runtime、raw profile 權重與 trait 科技倍率已閉合；其餘開局 colony 差異另列 |
| 2 / `+0x8A1` | Farming | `Colony_Empire_Base_Food2_Produced_` | 玩家產出主鏈已閉合；開局／AI 權重仍待窄切片 |
| 3 / `+0x8A2` | Industry | `Colony_Empire_Base_Industry_Produced_` | 玩家產出、污染與建造主鏈已閉合；開局／AI 權重仍待窄切片 |
| 4 / `+0x8A3` | Science | `Colony_Empire_Base_Research_Produced_` | 玩家逐人口與固定研究主鏈已閉合；開局／AI 權重仍待窄切片 |
| 5 / `+0x8A4` | Money | `Colony_BC_Production_`、`Twiddle_Initial_Homeworlds_` | 玩家 BC 主鏈與 Advanced Civilization 初始國庫 `(raw+2)*100` 已閉合；後者不是母星 planet 修改 |
| 6 / `+0x8A5` | Ship Defense | `Defensive_Combat_Bonus_`、`Missile_Dcv_`、`Qload_Ships_`、`Base_Generic_Dcv_`、`Get_Ship_Combat_Bonuses_` | 主要 consumer 閉合；格子 DCV／飛彈與戰機下游、快速 record 與估值均直接加 signed raw 值，見 `ship-racial-bonuses-transdimensional-audit-20260828.md` |
| 7 / `+0x8A6` | Ship Attack | `Do_Auto_Ship_Turn_`、`Offensive_Combat_Bonus_`、`Fighter_Ocv_`、`Base_Generic_Ocv_`、`Get_Ship_Combat_Bonuses_` | 主要 consumer 閉合；格子 OCV／戰機、戰術 AI、快速 record 與估值均直接加 signed raw 值，見同上 |
| 8 / `+0x8A7` | Ground Combat | `Compute_Player_Ground_Combat_Bonuses_` | 主公式已有地面戰稽核；AI 與人口回寫鏈仍由地面戰列判定 |
| 9 / `+0x8A8` | Spying | `Compute_Needs_`、`Compute_Spy_Bonuses_`、`Allocate_AI_Spies_` | signed 攻守加項、AI 生產／bootstrap、配置與任務主鏈已閉合；personality 正式名稱仍待 |
| 10 / `+0x8A9` | Low-G | `Enforce_Gravity_`、`Gravity_Player_Production_Factor_`、`Resolve_Bomb_Hit_`、`Compute_Player_Ground_Combat_Bonuses_` | 母星、生產、地戰 -10 與 AI 選址已閉合；完整轟炸強度上游另列 |
| 11 / `+0x8AA` | High-G | `Enforce_Gravity_`、`Gravity_Player_Production_Factor_`、`Resolve_Bomb_Hit_`、`Compute_Player_Ground_Combat_Bonuses_` | 母星、生產、耐受、轟炸門檻與 AI 選址已閉合 |
| 12 / `+0x8AB` | Aquatic | `Modify_Home_Worlds_`、`Colony_Empire_Base_Food2_Produced_`、`Player_Effective_Climate_`、`Size_And_Climate_Race_Pop_Limit_` | 食物／氣候／人口上限主鏈已閉合；母星生成仍待窄切片 |
| 13 / `+0x8AC` | Subterranean | `Uncolonized_Planet_Worth_To_Player_`、`Size_And_Climate_Race_Pop_Limit_`、`Compute_Player_Ground_Combat_Bonuses_`、`Init_NPC_Personalities_Objectives_Themes_`、`Calc_Tech_Value_` | 人口上限、地面防守 +10、raw profile 權重與 category 6 科技倍率已閉合 |
| 14 / `+0x8AD` | Large Homeworld | `Modify_Home_Worlds_` | 標準母星 planet size raw 3 與 Advanced Civ 平衡均已閉合 |
| 15 / `+0x8AE` | Rich/Poor Homeworld | `Modify_Home_Worlds_` | signed `-1/0/+1` 與礦產 raw 1／保持／3 已閉合 |
| 16 / `+0x8AF` | Artifacts Homeworld | `Twiddle_Selected_Adv_Civ_Planets_`、`Modify_Home_Worlds_` | 標準母星 special raw 10 與 Advanced Civ special 平衡均已閉合 |
| 17 / `+0x8B0` | Cybernetic | `Strategic_Combat_`、`Repair_All_Combat_Ships_`、`Colony_Industry_Maintenance_`、`Apply_Colony_Pop_Growth_`、`Uncolonized_Planet_Worth_To_Player_`、`Colony_Worth_To_Player_`、`Init_NPC_Personalities_Objectives_Themes_`、`Calc_Tech_Value_` | 食物／工業維護、格子每回合 +10%、戰後／戰略全修、AI 行星／殖民地 worth、raw profile 權重與 category 0 科技倍率 20 已閉合；profile 候選正式名稱未知 |
| 18 / `+0x8B1` | Lithovore | `Tech_Is_Legal_For_Player_`、`Ensure_At_Least_1_Food_Planet_`、`Colony_Food_Maintenance_`、`Apply_Colony_Pop_Growth_`、`Uncolonized_Planet_Worth_To_Player_`、`Colony_Worth_To_Player_`、`Init_NPC_Personalities_Objectives_Themes_`、`Calc_Tech_Value_` | 零食物、初始職務、六科技 gate、AI 食物保障 bypass、AI 行星／殖民地 worth、raw profile 權重與 category 0 科技倍率 1／優先序已閉合；profile 候選正式名稱未知 |
| 19 / `+0x8B2` | Repulsive | `Vote_Check_`、`Diplomacy_Screen_`、`NPC_To_NPC_Treaty_Negotiations_`、`Determine_Diplomacy_Messages_`、`Chance_To_Hire_Hero_`、`Apply_Assimilation_` | 主要公式／gate 閉合；議會 -100、proposal -50、領袖 -10／÷2、同化 ÷2、AI talker 1 與外交路徑限制已證實；choice／message／leader flag 表仍待，見 `repulsive-charismatic-trait-audit-20260828.md` |
| 20 / `+0x8B3` | Charismatic | `Vote_Check_`、`Change_Relations_`、`Get_Tech_Exchange_Reaction_`、`Chance_To_Hire_Hero_`、`Apply_Assimilation_` | 主要公式閉合；議會 +40、proposal／科技交換 +50、關係 delta 正×2負÷2、領袖 +5／+10、同化 ×2、AI talker 3 已證實，見同上 |
| 21 / `+0x8B4` | Uncreative | `Init_Player_Tech_`、`Player_Gets_Tech_App_` | 可選集合形成時的單項亂數限縮已閉合 |
| 22 / `+0x8B5` | Creative | `Init_Player_Tech_`、`Player_Gets_Tech_App_`、`Give_Player_Field_` | 突破時整 field 授予與固定全授予 fields 已閉合 |
| 23 / `+0x8B6` | Tolerant | `Determine_Event_`、`Compute_AI_Data_`、`Colony_Industry_Production_`、`Colony_BC_Production_`、`Init_NPC_Personalities_Objectives_Themes_`、`Calc_Tech_Value_` | 容量、混合人口污染、事件 gate、raw profile ID4 +100 與 category 4 科技倍率 1 已閉合 |
| 24 / `+0x8B7` | Fantastic Traders | `Colony_BC_Production_`、`Update_Player_Stats_`、`Trade_Agreement_Goal_` | 殖民地 Trade Goods、帝國食物盈餘與協議目標三鏈已閉合 |
| 25 / `+0x8B8` | Telepathic | `Ai_Self_Destruct_Check_`、`Boarding_Action_Type_`、`Capture_Ship_`、`Player_Can_Mind_Control_Colony_`、`Enemy_Colony_Worth_To_Player_`、`Compute_Spy_Bonuses_` | 部分閉合；間諜 +10、外交 +25、心控合法性、無運輸艦接管、prisoner 回寫及戰術／AI raw 分支已證實；五個 raw 下游仍待，見 `telepathic-trait-audit-20260828.md` |
| 26 / `+0x8B9` | Lucky | `Get_Event_Victim_`、`Get_Antaran_Victim_`、`Determine_Lucky_Players_Events_`、`Update_Lucky_Players_Events_` | 事件目標與累積鏈已有獨立稽核；仍須併入事件總列 |
| 27 / `+0x8BA` | Omniscience | `Find_Ship_Stacks_`、`Print_Fltscrn_Scanned_Star_Name_`、`Print_Scanned_Ship_Data_`、`Player_Is_Omniscient_`、`Star_Owner_`、`Explored_New_Star_` | 主要玩家鏈已閉合；Galactic Lore 共用查詢、Fleet UI、owner／stack 過濾、首次抵達報告遮罩、五類一次性發現保留與隱藏熱鍵已分鏈，見 `omniscience-stealthy-ships-audit-20260828.md` |
| 28 / `+0x8BB` | Stealthy Ships | `Remove_Non_Detected_Ships_`、`Add_Specials_To_Design_`、`Compute_AI_Data_`、`Init_NPC_Personalities_Objectives_Themes_`、`Calc_Tech_Value_` | 五個 direct site 的玩家鏈已閉合；快速結算 raw 6／23／31 table effect 均為 0／0，格子載入器只複製裝置 bitfield，沒有 trait→bitfield producer。trait 與裝置戰鬥等價已由強推論否定；裝置各自的格子間接 consumer 另追，見同上 |
| 29 / `+0x8BC` | Trans-Dimensional | `Missile_Speed_`、`Calc_Current_Speed_`、`Get_FTL_Speed_`、`Get_Ship_Combat_Bonuses_`、`Ships_Try_To_Move_To_` | 主要 consumer 閉合；格子／飛彈速度 +4、估值 DCV +20、FTL +2 與 Hyperspace Flux 免疫已分鏈；撤退完整 gate 等仍待，見同上 |
| 30 / `+0x8BD` | Warlord | `Calc_Ship_Level_`、`Owned_Officer_Level_`、`Compute_Player_Maintenance_`、四個 ground-unit limit 函式、`Init_NPC_Personalities_Objectives_Themes_` | 艦員／領袖 +1、每殖民地 command +2、地面容量兩倍與 raw profile ID2 +10 已閉合 |

## 本輪推翻的舊假設

1. 「22 項 UI 選項有消費端」不能推出「31-byte runtime 全部已閉合」。UI 選項數與 runtime
   欄位數是不同分母。
2. `TRAIT_POOR_HOMEWORLD=31` 只是二手程式碼的 convenience enum；原版 runtime 沒有 index 31。
3. `+0x8B8` 可由 `Convert_Custom_Race_Flags_` 的 31-byte mapping 與多個原始符號交叉確認為
   Telepathic；舊文件若仍稱名稱未知，屬過期斷言。
4. `+0x8BA` 與 `+0x8BB` 不是只影響 remake 自建的可見度數字；原版各自有多個地圖、掃描、
   艦隊過濾、設計與 AI consumer，必須分支追查。

## 下一批最小 RE 切片

依玩家體驗與交叉依賴排序：

1. NPC profile／科技估值的全部 trait direct site 已閉合；後續只追 raw profile／category 的
   正式玩家名稱及非 trait 共同表，不再逐 trait 重開相同函式。
2. Ship Attack／Ship Defense／Trans-Dimensional 剩餘 raw 下游：strategic record 欄名、四個
   bonus 輸出型別、撤退完整 gate、科技 raw ID 與全域 RNG 序列；共同 NPC raw profile 已閉合。
3. Repulsive／Charismatic 剩餘表格：完整 choice／message ID、advanced officer 候選、AI leader
   flag、sneak-attack 最終權重、talker raw gate 與 RNG；共同 NPC raw profile 已閉合。
4. Government：主要玩家鏈已整併於 `government-trait-audit-20260828.md`；後續只追 AI 評分、
   產品合法性 raw 表與 occupation policy，不重開已閉合公式。
5. 其餘經濟欄的初始 colony 窄切片；Advanced Civilization 全圖分配、Money 國庫與共同 NPC
   raw profile 已閉合，不再以母星 trait 名義重開。

只有每一切片具備 raw 位址、輸入狀態、公式／表、玩家可見 consumer 與未知邊界後，才可把
parity matrix 的「客製種族效果」列升格；目前仍是進行中。
