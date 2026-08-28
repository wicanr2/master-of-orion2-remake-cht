# AI 殖民地職務分配規格

規格狀態：封鎖／未封鎖職務、帝國工業／研究平衡、食物運輸容量與追加農夫切片
`CONFORMED`；`sub_D682A` packed 額外 `+1000` 分支仍為 `DRAFT`。

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

封鎖造成的 AI 對真人積怨已接線；仍未完成的是 remake 尚不可表示的真人對真人 policy。
它屬封鎖世界狀態，不改變本文件已閉合的職務公式。資料不完整時仍不得以
personality weights 或平均種族產出冒稱原版路徑。

`sub_D682A` 的 packed 額外 `+1000` 分支仍是 `DRAFT`；它會影響候選排序的少數條件，
但不推翻已閉合的容量、運糧、維護費與 AI 增艦鏈。
