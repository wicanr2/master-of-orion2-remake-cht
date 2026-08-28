# AI 殖民地職務分配規格

規格狀態：封鎖／未封鎖職務、帝國工業／研究平衡、食物運輸容量與追加農夫公式
為 `CONFORMED`；Watcom `qsort` 等價候選的逐人口排列為 `APPROXIMATED`。

證據與未知邊界見
[`../re/ai-colony-jobs-audit-20260828.md`](../re/ai-colony-jobs-audit-20260828.md) 與
[`../re/ai-food-assignment-audit-20260828.md`](../re/ai-food-assignment-audit-20260828.md)。

## 預定範圍

- 取代 `ApplyAIEconomy` 內逐殖民地一次性 `ColonyJobs` 的原版模式路徑。
- 保留封鎖與未封鎖分流、逐人口排序、邊際輸出、帝國級平衡與
  每次改職後重算。
- 新局、原版 `.GAM` 匯入與 remake JSON 讀檔都要經過同一條 typed 路徑。
- 玩家帝國不吃 AI 指派；難度加成只在已證實的產出階段套用。

## 已閉合，可供實作的子切片

- raw job `0/1/2` 對應農夫／工人／科學家。
- 四個 comparator 的主鍵與全部已見 tie-break。
- `player+0xAA/+0xAC` 的 producer、尺度與帝國停止比較方向。
- `PopulationGroups` 足以表示已見排序資料；同 race／profile／prisoner／job 的
  人口在 comparator 上等價，實作可用「群組＋人數」而不改變排序結果。
- `sub_D5FE1` 只把 Android／Natives 放進不可重配前段；Tolerant／PRISONER
  讀取只形成另一個策略旗標，不會把一般人口排除於職務候選。

`engine.ApplyOriginalAIUnblockadedJobs` 已依此建立逐 population group 的等價
colonist 候選，先套最低工人／半工業消耗，再以帝國研究與淨工業逐人平衡；
每次改職同步總數、race group 與 PRISONER。呼叫端只有在逐種族 profile 與
`colony+0xDD` 等價值完整時使用，否則明示回退既有 remake AI。這不代表整份
AI 職務分配完成。

`engine.ApplyOriginalAIJobs` 另已接 `sub_D61E7`：有農業的封鎖殖民地依
`food-industry` 由大到小選農夫，糧食足夠後由末端把其餘人口改為工人；無農業時
依事件 filter 全改工人或科學家。Android／Natives 的前置區保持不變。shell 於
每回合 AI 職務階段讀取進入回合時的 `Star.BlockadedMask`，符合原版「先消費舊 mask、
艦隊移動／戰鬥後才重算下一輪 mask」的主鏈順序。

IDA 後續證實 `sub_D6E1D` 不會在 `sub_D66B3` 停止；建造 pass 後仍呼叫
`sub_D6AD4 → sub_D6A00`，依 `player+0xB0/+0x38` 反覆增加農夫。remake 現以
`TotalFoodHalf` 對映 `+0xB0`，並以 `ActiveFreighters`、`SettlersFreighted`、
`FoodFreighted`、`SurplusFreighters` 對映 `player+0x36/+0x40/+0x3E/+0x38`。
每名在途殖民者占 5 艘；未封鎖殖民地依本地盈虧與可用艦數運糧，運力不足時只從本地仍缺糧的
殖民地選追加農夫。運輸壓力成立時，AI 依 `Random(10)<=difficulty` 直接增加 5 艘貨運艦。

`sub_D652C` 在逐殖民地前段職務完成後保存 `PollutionCleanupCost&0xff`；帝國工業／研究
平衡可能因增加工人而提高清污成本。追加農夫候選中，工人分數為
`food-worker+[cleanupBefore<cleanupNow]*1000`，其他可改派職務為 `food-research`。
殖民地內依分數降冪，再依 food、raw job、race slot 升冪；跨殖民地只比較首候選分數，
同分保留較高 colony index。

`sub_D66B3` 的停止比較必須使用和帝國結算相同的本回合難度／事件暫態加成；只把新職務與
`PopulationGroups` 寫回持久殖民地，不得把暫態成長或難度欄位重複累加。原版比較器無法在
同種族、同產出時決定次序，而 remake 已折疊逐人口陣列；等價類別採最小改職的確定性重建，
並在折疊結果令帝國研究為 0 時，於人口至少 2 且未封鎖的殖民地保留一名可研究人口；單人口
殖民地與封鎖分支仍完全遵守各自原版規則。此條只保證可重現與不中斷研究，不列為
原版精確 tie-break。

## CONFORMED 驗收

1. 一般 race 的農夫會經候選迭代分配為工人／科學家，Android／Natives 保持原職。
2. 改職後 `PopulationGroups`、三職務總數及 PRISONER 總數一致。
3. 後段殖民地輸入不完整時回傳 fallback，且不得部分修改較早的呼叫端狀態。
4. 新增的 engine 與 shell 定向回歸須以 Docker、`-count=1` 通過；全套既有
   shell 隨機長跑另由總體測試債追蹤，不作本公式的虛假完成證據。
5. 正常 200 回合下，至少一個非 Creative AI 必須完成新主題並保存 application 擇一；
   不得再因全工人／零農夫造成母星人口縮至 1 的死亡螺旋。
6. 充足運力允許盈餘殖民地供應缺糧殖民地；零運力才要求缺糧殖民地本地補農夫。
7. 貨運艦維護費只計使用量：`SurplusFreighters>0` 時為
   `floor((TotalFreighters-SurplusFreighters)/2)`，否則為 `floor(TotalFreighters/2)`。
8. 清污成本未上升時依 `food-currentJobOutput` 選人；清污成本上升時工人候選取得
   `+1000`，且保存值必須維持原版低 byte 截斷。
9. 職務重算使用本回合暫態難度／事件輸入，但持久狀態不得累加這些暫態欄位。
10. 等價候選重建不得讓仍有可研究人口的 AI 永久維持 `TotalResearch=0`；此驗收屬
    `APPROXIMATED`，不把綠燈升格為原版 `qsort` parity。

封鎖造成的 AI 對真人積怨已接線；仍未完成的是 remake 尚不可表示的真人對真人 policy。
它屬封鎖世界狀態，不改變本文件已閉合的職務公式。資料不完整時仍不得以
personality weights 或平均種族產出冒稱原版路徑。

本規格只宣稱 AI 殖民地職務與食物貨運鏈閉合；不代表艦隊目標、外交、產品配額等其他
AI state machine 已完成。
