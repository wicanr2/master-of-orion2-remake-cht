# AI 殖民地職務分配規格

規格狀態：可表示的封鎖／未封鎖殖民地與帝國平衡 `CONFORMED`。

證據與未知邊界見
[`../re/ai-colony-jobs-audit-20260828.md`](../re/ai-colony-jobs-audit-20260828.md)。

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

## CONFORMED 驗收

1. 一般 race 的農夫會經候選迭代分配為工人／科學家，Android／Natives 保持原職。
2. 改職後 `PopulationGroups`、三職務總數及 PRISONER 總數一致。
3. 後段殖民地輸入不完整時回傳 fallback，且不得部分修改較早的呼叫端狀態。
4. 新增的 engine 與 shell 定向回歸須以 Docker、`-count=1` 通過；全套既有
   shell 隨機長跑另由總體測試債追蹤，不作本公式的虛假完成證據。

封鎖造成的 AI 對真人積怨已接線；仍未完成的是 remake 尚不可表示的真人對真人 policy。
它屬封鎖世界狀態，不改變本文件已閉合的職務公式。資料不完整時仍不得以
personality weights 或平均種族產出冒稱原版路徑。
