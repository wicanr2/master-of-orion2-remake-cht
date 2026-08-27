# AI 殖民地職務分配規格

規格狀態：整體 `DRAFT`；「未封鎖殖民地＋帝國平衡」子切片 `CONFORMED`。

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

## CONFORMED 驗收

1. 一般 race 的農夫會經候選迭代分配為工人／科學家，Android／Natives 保持原職。
2. 改職後 `PopulationGroups`、三職務總數及 PRISONER 總數一致。
3. 後段殖民地輸入不完整時回傳 fallback，且不得部分修改較早的呼叫端狀態。
4. `internal/engine`、`internal/shell` 與純 Go 回歸套件以 Docker、`-count=1` 通過。

## 整體 READY 前必須閉合

1. 封鎖狀態從原版 `.GAM`／當回合艦隊狀態到 `ColonyState` 的 producer。
2. 封鎖路徑 `+0xDD/+0xE0/+0xE7/+0xFC..+0xFF` 的完整欄位契約。

上述任一項未閉合前，本規格不得升為 `READY`，生產程式也不得以
personality weights 或平均種族產出填補缺欄。
