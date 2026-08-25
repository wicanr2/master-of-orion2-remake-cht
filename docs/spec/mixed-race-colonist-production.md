# 混合種族人口群組與產出規格

## 型別契約

`ColonyState.PopulationGroups` 以 player slot 聚合人口。每群保存：

- `RaceSlot` 與 `RaceSlotKnown`；零是合法 slot，不能兼作未知哨兵。
- 三種職務人數與各職務 prisoner 人數。
- 該 slot 的農業／工業／科研數值 trait、Low-G／High-G 與 Aquatic。
- 該 slot 的人口成長 trait、種族人口上限能力及 0..999 成長累積點。

群組三職務總和必須等於 `Population`，且各 prisoner 數不得超過同職務。契約不完整時，僅對
舊 JSON 回退既有總數模型；新建、`.GAM` 匯入與征服後的狀態都必須完整。

## 產出契約

1. 殖民地既有 `*Per*` 保存 owner 快取。另保存 owner 的三項數值 trait；群組有效時先扣掉
   owner trait，再加入群組 trait。
2. 食物還要把 owner Aquatic 氣候差扣掉，再加入群組 Aquatic 氣候差。
3. 每群依自己的重力適性套 100%／75%／50%，Planetary Gravity Generator 對所有群組歸零。
4. prisoner 只產正常值 3/4；工人仍保有每人至少 1 產能。
5. 群組產出加總後才加入殖民地固定建築產出。舊 JSON 沒有完整群組時沿用 owner fallback。
6. Android 固定使用 `+6/+3/+3`，Natives 使用 `+4/+0/+0`；兩者忽略重力。

## 逐 slot 成長契約

1. 每群以 `ColonyBaseGrowth(group population, total population, group population limit)`
   獨立求基礎值，再套該群 `GrowthBonusPercent`、殖民地共用 `GrowthBonusSum` 與住房百分點。
2. 群組人口上限以殖民地 owner 的 `PopMax` 為快取，扣掉 owner Aquatic／Tolerant／Subterranean
   差，再加入該群差；建築與科技造成的共同容量增量因此保留。
3. 每群把本回合成長加入自己的 `GrowthPoints`；達 1000 時新增該群人口並扣 1000。
   `.GAM RacePopulation[slot]` 匯入此累積值。
4. `FlatGrowth` 是殖民地整體固定成長，加入 owner 群；舊 JSON 沒有完整群組時繼續使用既有
   `popAccum` 與 owner race fallback。
5. 原版負成長的消耗分類、短缺分配、候選優先序與亂數流程已閉合並移至
   `population-consumption-and-negative-growth.md`；本文件只保留正成長與群組共同契約。

## 狀態變更契約

- 新殖民地與母星：建立 owner 群組；原住民若存在則建立 slot 9 群組。
- `.GAM`：由前 `Population` 筆 packed colonist 依 race slot、job、PRISONER 聚合，再由原始
  `Players[slot].Traits` 注入 profile。
- 改派：在來源職務選固定、可重播的第一個有人的群組移動一名；優先移 prisoner 與既有 UI
  的 prisoner 改派近似一致。
- 征服／同化：征服只改 prisoner／loyalty 語意，不改 race slot；同化只清 prisoner，不改種族。
- 成長：新人口加入達門檻的實際群組；既有職務選擇仍是「先工人、缺糧則農夫」的已註記近似。
- 傷亡：從實際被扣職務的群組同步扣一人；不得只改 ColonyState 三個總數。

## 驗收

1. 同一 Heavy-G 星球上 Low-G 與 High-G 工人的產出不同，且加總符合逐群計算。
2. 席隆科學家移居人類殖民地仍保留 +2 research；殖民地 owner 不因此取得 +2。
3. `.GAM` 匯入後 race/job/prisoner 三個維度都能 JSON 往返。
4. 改派、征服、同化、成長與傷亡後群組總和持續等於殖民地總數。
