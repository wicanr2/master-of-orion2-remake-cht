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
5. 完工後寫入 `ColonyBuildings`，並把已建模的累積經濟效果寫回該 AI 的 `ColonyState`。
6. 若沒有任何可建建築，該殖民地產能才交給既有 AI 艦艇產品轉接層；此轉接層不冒稱
   `sub_D10EE` 的完整戰鬥艦選擇器。
7. 舊存檔缺少 `ColonyBuilds` 時視為尚無目前產品，下一回合依當時狀態決定性建立。

## 驗收

- 有唯一可建建築時，產能只增加該殖民地產品進度，不增加造艦池。
- 產品完工後建築 map 與 typed 產出效果同步，產品清空。
- 無可建建築時，殖民地產能完整交給造艦轉接層。
- 存檔／讀檔保存進行中的 AI 殖民地產品。
- 既有擴張測試以「建築＋造艦總投入」驗證新殖民地確實參與經濟，不再假設所有產出都是軍艦。
