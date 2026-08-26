# AI 殖民地建造選擇規格

## 目的

把 AI 生產從「全帝國工業直接造艦」改成「每座殖民地有自己的產品與進度」。本輪先閉合
原版空產品的建築候選控制流；完整帝國配額與戰鬥艦選擇仍由後續 RE 取代窄轉接層。

## 原版錨點

- `All_Colony_AI_ → Colony_AI_ → sub_D6E1D → sub_D10EE／Assign_Colony_Building_`。
- 空產品建築選擇掃 raw ID 1..48，先做可建 gate，再計分、套
  `(6-difficulty)×score >= maxScore` 門檻，最後加權抽選。
- 證據與分級見 `docs/re/ai-colony-build-selection-audit-20260826.md`。

## Remake 契約

1. `AIOpponent.ColonyBuilds` 以殖民地星系索引保存 `ColonyBuild`；進度隨 JSON 保存。
2. 每座殖民地只消費自己的 `ColonyOutput.NetIndustry`。同一份產能不可同時蓋建築與造艦。
3. 候選只含該 AI 科技已解鎖、且該殖民地尚未完成的 typed 建築。
4. 候選套原版已證實的難度濾門與加權抽選形狀；typed 分數與亂數位置明標近似。
   raw ID 4、7、12、34、36 例外：使用 IDA 已閉合的完整精確公式；`player+0x28==4`
   對應 `PersonalityHonorable`，人口取該殖民地 `Population`。raw ID 6、19、30、35 使用
   完整精確公式：AI 已選中／完成任一 Hyper field，或目前有原版優先建築 gate 時為 0；
   否則分別為 `11+4×[Erratic]`、`11`、`8+3×[Erratic]`、`5+2×[Erratic]`。
   raw ID 15 Biospheres 亦使用完整精確公式：priority gate 時為 0，否則
   `18+[Pacifist]`；late-tech 不影響此式。
   raw ID 16 Food Replicators 在逐種族人口資料完整時使用完整精確公式：主要人口與 owner
   同為 Lithovore 時為 0；其餘在帝國食物盈餘為負時 `8+[Pacifist]`，
   否則 `4+[Pacifist]`。主要人口依原版 player-slot 計數與 owner fallback 規則決定；
   profile 不完整時仍走明示類別 fallback，不把缺資料的零值冒充原版零分。
   raw ID 10 Cloning Center 使用原版結算前國庫／淨 BC 因子：priority gate 時為 0；
   否則為 `floor(budgetFactor/2)`，且只有負人口成長種族的 Pacifist 再加 1。
   `budgetFactor` 在結算前國庫小於 1500 時為 0；否則依原版 signed word 淨 BC
   朝零除以 64、unsigned 整數平方根後夾到 10。這不是亂數，不得消費候選抽選 RNG。
   raw ID 20 Holo Simulator 與 raw ID 31 Pleasure Dome 共用士氣建築 gate：
   Unification／Galactic Unification 時為 0；其餘政府在人口至少 3 時分別固定為 10／16，
   人口恰為 2 時只有 `budgetFactor>0` 才取同一正值，人口 0／1 為 0。兩式不讀
   priority gate、late-tech 或 personality；budget factor 只作 gate，不加入固定分數。
   raw ID 17 Gaia Transformation 使用 `budgetFactor+[Pacifist]`；不讀 priority gate、
   late-tech、人口或政府。它是一次性 Special 產品，不寫入 `ColonyBuildings`；候選只在
   已取得前置科技且殖民地為 Terran 時出現，完成後同步把 AI 殖民地與對應全局行星改成 Gaia。
   raw ID 44 Terraforming 在 priority gate 時為 0；否則為
   `climateBase+3×[Pacifist]+budgetFactor`。非 Aquatic／Aquatic 的 Barren、Desert、Tundra、
   Ocean、Swamp、Arid 基礎分依序為 `2/2、1/1、0/1、4/0、6/0、1/1`。候選只在
   `TerraformNextClimateOptions` 有結果時出現；它同樣是一次性 Special，完成後同步 AI
   殖民地與全局行星，Barren 多候選沿用既有固定第一項近似。
   raw ID 37 Soil Enrichment 在 priority gate、主要人口與 owner 同為 Lithovore，或
   `FoodPerFarmer<=0` 時為 0；否則為
   `3+2×[帝國食物盈餘<0]+2×[Pacifist]`。它只在既有土壤適用氣候提供候選，完工使
   `FoodPerFarmer+1`，不寫入常駐建築集合。
   raw ID 21 Hydroponic Farm 不讀 priority gate；`cache+2==0` 時為 0，否則依原版
   `colonyFoodHalf` 的 `0／1／2／其餘` 取 `12／11／10／6`，再加完整帝國食物赤字幅度與
   `4×[Pacifist]`。raw ID 43 Subterranean Farms 在 priority gate 或 `cache+2==0` 時為 0，
   同一四段快取取 `13／12／10／7`，再加完整赤字幅度與 `3×[Pacifist]`。raw ID 46
   Weather Controller 在 priority gate、`cache+2==0` 或 `colonyFoodHalf<=0` 時為 0；其餘
   食物赤字時為 `10+2×[Pacifist]`，非赤字時為 `5+2×[Pacifist]`。
   `colonyFoodHalf` 不得由已含 owner Farming／Aquatic 的 `FoodPerFarmer` 直接倍增；候選建立端
   必須以 `2×ClimateFoodPerFarmer(Climate)` 加已建 Weather Controller 的 4 與 Astro
   University 的 2 重建；基值為 0 且已知 Biomorphic Fungi 時先改成 2。若缺少完整人口
   profile 或食物快取來源，回報非 exact。
   raw ID 5 Atmospheric Renewer 與 raw ID 32 Pollution Processor 共用污染分數；raw ID 13
   Core Waste Dumps 使用同式，但 priority gate 成立時先歸零。主要人口為 Tolerant 或
   `PollutionCleanupCost<=5` 時為 0；清污成本 6..10 時為 `[Pacifist]`；大於 10 時為
   `floor(sqrt(PollutionCleanupCost))+[Pacifist]`。清污成本必須取本回合
   `ColonyOutput.PollutionCleanupCost`，不可改以總工業或 polluting production 估算；主要人口
   Tolerant 依同一 player-slot 選擇規則判定，profile 不完整時回報非 exact。raw 5／32 不得
   誤吃 priority gate，raw 13 不得漏掉。
   raw ID 25 Planetary Gravity Generator 使用 owner 的重力 trait 與行星重力：High-G owner
   只有在 Low-G 星球取 `3+[Pacifist]`；Low-G owner 在 Normal-G／Heavy-G 分別取
   `3／6+[Pacifist]`，Low-G 星球為 0；一般 owner 在 Low-G／Heavy-G 分別取
   `3／6+[Pacifist]`，Normal-G 為 0。若髒資料同時有 High-G／Low-G，High-G 優先。此式不讀
   priority gate、budget factor、late-tech 或主要人口種族。完工必須寫入
   `NormalizeGravity`，並由既有逐人口產出路徑消費。
   raw ID 29 Planetary Stock Exchange 與 raw ID 39 Spaceport 在 priority gate 時為 0，且
   分別要求人口至少 5／3；通過後皆為
   `floor((population+primaryPopulationCapacity+[Honorable])/3)`。raw ID 33 Recyclotron
   不讀 priority gate，公式為
   `floor((2×population+primaryPopulationCapacity)/3)+2×([primary non-Tolerant]+[Pacifist]+[Honorable])`。
   `primaryPopulationCapacity` 必須由主要人口完整 profile、PlanetSize／Climate、Advanced City
   Planning application 與 Biospheres 建築重建，不能直接把 owner 口徑的 `PopMax` 當成同義欄；
   profile 不完整時回報非 exact。三棟完工都要由正常帝國結算消費：兩棟貿易建築改變該殖民地
   BC 收入，再生反應爐增加不產生污染的人口產能。
   raw ID 38 Space Academy 在 priority gate 時為 0；其餘若淨工業低於 17、人口低於 5 且
   `budgetFactor==0` 時為 0，否則取
   `min(1000,isqrt32(uint32(int16(NetIndustry)-15)))`。`NetIndustry` 必須取本回合
   `ColonyOutput.NetIndustry`，不可改用 GrossIndustry。負差值依原版 unsigned helper 產生
   65535 後夾成 1000，不得平滑成 0。完工必須寫入 AI 殖民地建築 map，並由同星系 AI 艦艇
   每回合經驗加成 consumer 驗證；AI 匯總造艦尚無來源殖民地，因此不宣稱 AI 新艦起始等級
   已精確接線。
   raw ID 2 Armor Barracks 與 raw ID 22 Marine Barracks 使用同一份星系壓力 context。
   四個 reach count 依序是近圈條約、近圈無正式政策、近圈戰爭、延伸圈帝國數；另有其他
   帝國艦隊 ETA=9 與殖民地內戰爭帝國外族人口旗標。Armor Barracks 在人口 `<3` 且
   `budgetFactor==0` 時為 0，否則為
   `2×ETA9+條約+無政策+3×戰爭+延伸+[base!=0]×[Ruthless]`，再於 Marine Barracks 未建且
   `government/2<=1` 加 6、hostile alien population 加 1。Marine Barracks 的人口門檻是 2，
   基礎式為 `5×ETA9+條約+3×無政策+6×戰爭+2×延伸+[base!=0]×[Ruthless]`，再於 Armor
   Barracks 未建且同一政府 gate 加 12、hostile alien population 加 3。燃料科技須先轉成
   原版 `player+0x324` 的 R 秒差距口徑；若 session-wide 外交、航程、艦隊或 population-slot
   對映不完整，這兩式必須回報非 exact 並走明示 fallback。
5. 優先建築 gate 僅由已知科技 application、已建建築、殖民地礦產及 AI 生效政府組成：
   - Ultra Poor／Poor／Abundant 殖民地已知 Automated Factories 但未建 Automated Factory；或
   - Feudal／Confederation／Dictatorship／Imperium 已知但未建 Marine Barracks／Armor Barracks。
   科技主題完成但 application 未被授予，不得誤算成已知科技。
6. 完工後寫入 `ColonyBuildings`，並把已建模的累積經濟效果寫回該 AI 的 `ColonyState`。
7. 若沒有任何可建建築，該殖民地產能才交給既有 AI 艦艇產品轉接層；此轉接層不冒稱
   `sub_D10EE` 的完整戰鬥艦選擇器。
8. 舊存檔缺少 `ColonyBuilds` 時視為尚無目前產品，下一回合依當時狀態決定性建立。

## 驗收

- 有唯一可建建築時，產能只增加該殖民地產品進度，不增加造艦池。
- 產品完工後建築 map 與 typed 產出效果同步，產品清空。
- 無可建建築時，殖民地產能完整交給造艦轉接層。
- 存檔／讀檔保存進行中的 AI 殖民地產品。
- 五個完整閉合 raw ID 在一般與 Honorable 性格下逐式符合原版立即數公式。
- raw 6／19／30／35 在一般與 Erratic 性格下符合完整正值公式；晚期科技與兩類優先建築
  gate 的邊界皆精確歸零。
- raw 15 在一般／Pacifist 性格下分別為 18／19，priority gate 時為 0；late-tech 不得誤歸零。
- raw 16 在一般／Pacifist、食物赤字／非赤字邊界符合 `8/9`、`4/5`；只有主要人口與 owner
  同為 Lithovore 時為 0，profile 不完整時必須回報非 exact。
- raw 10 在結算前國庫 1499／1500、淨 BC 的平方根邊界、負淨 BC 與負／非負人口成長下
  符合完整公式；計分不得改變加權抽選的亂數位置。
- raw 20／31 在八種政府、人口 1／2／3 與 budget factor 0／正值的邊界符合完整固定分數；
  priority gate 與 personality 變化不得影響結果。
- raw 17 在結算前國庫 1499／1500、一般／Pacifist 性格下符合
  `budgetFactor+[Pacifist]`；非 Terran 不得成為候選，完工後不得殘留為常駐建築，且
  AI 殖民地與對應全局行星都必須是 Gaia。
- raw 44 的六種有效氣候在 Aquatic／非 Aquatic、一般／Pacifist、priority gate 與
  budget factor 0／正值邊界逐式符合內層跳表；Toxic／Radiated／Terran／Gaia 不得成為候選，
  完工不得寫入常駐建築，且殖民地／行星氣候必須同步前進一級。
- raw 37 在兩種 Lithovore 組合、`FoodPerFarmer` 0／1、食物盈餘 -1／0、一般／Pacifist
  與 priority gate 邊界符合公式；不適用氣候不得成為候選，完工只增加每農夫食物且不留常駐旗標。
- raw 21／43 的四段 `colonyFoodHalf` 表、赤字幅度（不只正負）、一般／Pacifist、cache gate
  與 priority gate 差異必須逐式測試；raw 46 另測 `colonyFoodHalf` 0／正值與赤字正負。
  三棟都要以唯一正常候選走過選擇、逐殖民地產能、完工旗標及 typed 食物效果。
- raw 5／13／32 在主要人口 Tolerant／非 Tolerant、清污成本 5／6／10／11、一般／Pacifist
  與 priority gate 邊界符合公式；三棟均以唯一正常候選走過選擇、逐殖民地產能與完工後
  污染旗標效果。測試必須把 `PollutionCleanupCost` 與 gross／polluting production 分開。
- raw 25 的一般／Low-G／High-G owner × 三種行星重力、一般／Pacifist、雙 trait 優先序與
  priority gate 不影響性逐格符合表格；唯一正常候選須走過逐殖民地產能、完工旗標與重力懲罰
  消費端，不能只測 `NormalizeGravity` 欄位被設為 true。
- raw 29／39 要測人口門檻前後、priority gate、一般／Honorable 與主要外族人口容量；raw 33
  要測 Tolerant、Pacifist、Honorable 各自的 `+2`、混合人口與 profile 不完整邊界。三棟皆以
  唯一正常候選走過逐殖民地產能、完工旗標及帝國收入／無污染產能消費端。
- raw 38 要測 priority gate、人口 4／5、budget factor 0／正值及淨工業 14／15／16／17 的
  unsigned 不連續邊界；唯一正常候選完工後，停泊同星系的 AI 實艦每回合必須多得學院經驗。
- raw 2／22 要逐項測四個 reach count、其他帝國 ETA=9、Ruthless、兩個人口／budget gate、
  八種政府、交叉兵營旗標與 hostile alien population；正常唯一候選完工後，必須由既有
  駐軍回合鏈觀察 Marine／Tank 數量成長，不能只檢查建築 map。
- 只完成多選主題但選了其他 application 時，不得觸發相應 Automated Factory／Barracks gate。
- 精確分支與類別式 fallback 不可混稱同一證據等級。
- 既有擴張測試以「建築＋造艦總投入」驗證新殖民地確實參與經濟，不再假設所有產出都是軍艦。
