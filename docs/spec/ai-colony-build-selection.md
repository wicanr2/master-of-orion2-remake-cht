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
- 只完成多選主題但選了其他 application 時，不得觸發相應 Automated Factory／Barracks gate。
- 精確分支與類別式 fallback 不可混稱同一證據等級。
- 既有擴張測試以「建築＋造艦總投入」驗證新殖民地確實參與經濟，不再假設所有產出都是軍艦。
