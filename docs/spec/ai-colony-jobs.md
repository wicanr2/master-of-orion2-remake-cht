# AI 殖民地職務分配規格

規格狀態：`DRAFT`。

證據與未知邊界見
[`../re/ai-colony-jobs-audit-20260828.md`](../re/ai-colony-jobs-audit-20260828.md)。

## 預定範圍

- 取代 `ApplyAIEconomy` 內逐殖民地一次性 `ColonyJobs` 的原版模式路徑。
- 保留封鎖與未封鎖分流、逐人口排序、邊際輸出、帝國級平衡與
  每次改職後重算。
- 新局、原版 `.GAM` 匯入與 remake JSON 讀檔都要經過同一條 typed 路徑。
- 玩家帝國不吃 AI 指派；難度加成只在已證實的產出階段套用。

## READY 前必須閉合

1. 四個 `qsort_` comparator 的完整比較鍵與 tie-break。
2. `player+0xAA/+0xAC` 的 producer、signedness 與尺度。
3. 封鎖路徑的食物與最低產能停止邊界。
4. 現有 `PopulationGroups` 資料能否無損表示原版排序輸入。

上述任一項未閉合前，本規格不得升為 `READY`，生產程式也不得以
personality weights 或平均種族產出填補缺欄。
