# AI 殖民地職務分配靜態稽核

日期：2026-08-28

## 證據契約

- 輸入 `Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 正式 `Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4，映像 `ida-pro-9.4-idapython:py312-v1`；位址為
  DOS/4GW LE image 的 IDA linear EA。
- 非破壞性匯出：
  [`evidence/ai-colony-jobs-ida-20260828.json`](evidence/ai-colony-jobs-ida-20260828.json)；
  保留原始 `sub_` 名、函式邊界、bytes、運算元與 caller。

`ida-pro-9.4-idapython:locked-v1` 本輪因 `Python3TargetDLL` 與 license
環境未載入而退出 1，沒有輸出證據；依正對照改用已驗證的
`py312-v1` 後，同一份 `.i64` 成功產生 545,982-byte JSON。前一次空輸出
不作為「函式不存在」或玩法結論。

## 已證實的控制流

1. `sub_D6E1D @ 0xD6E1D` 先分流封鎖殖民地
   `sub_D61E7 @ 0xD61E7` 與未封鎖殖民地 `sub_D652C @ 0xD652C`，
   處理完才呼叫帝國級 `sub_D66B3 @ 0xD66B3`。
2. 兩條殖民地路徑都先呼叫 `sub_D5FE1 @ 0xD5FE1`；它掃描
   `colony+0x0C` 起的 4-byte colonist records，計數當前職務，並以
   `qsort_` 排序可重新指派的人口。
3. `sub_D652C` 以 `(population+4)/5` 建立一個最低需求，呼叫
   `sub_E1D59` 重算殖民地輸出，並逐個修改 colonist record
   的 2-bit job 欄位；不是只寫轉換後的 farmers/workers/scientists 總數。
4. `sub_DDFD3 @ 0xDDFD3` 以 raw job 索引呼叫
   `funcs_DDFF0 @ 0x18355C`；三個 table target 依序是已獨立閉合的
   食物 `sub_DE0C6`、工業 `sub_DED47`、研究 `sub_DFE77`，因此 raw
   `0/1/2` 分別是農夫／工人／科學家。
5. `sub_D648A @ 0xD648A` 對候選範圍兩端分別以科學家 raw `2`
   與工人 raw `1` 呼叫 `sub_DDFD3`，將研究−工業邊際差寫入每個 13-byte
   殖民地 AI 記錄 `+8` 與 `+0x0A`。沒有可指派人口時分別寫
   `0x8001` 與 `0x7FFF`作為 sentinel。
6. `sub_D66B3` 先以 `sub_E2D72` 重算帝國，然後從所有
   未封鎖殖民地選最大 `+8` 候選，把一名人口改為科學家 raw `2`；
   每次都呼叫 `sub_D648A` 與 `sub_E2A70` 重算。到達帝國停止條件後，
   再以最小 `+0x0A` 候選逐名改為工人 raw `1`。舊稿把兩個 raw job
   方向寫反；本結論以 `sub_DDFD3` 的間接表與實際 bit write 共同訂正。
7. 帝國停止條件直接讀 `player+0xAA` 與 `player+0xAC`。權重基值
   為 `10` 與 `18`；`player+0xB2 < 0` 時後者加
   `isqrt(-signed(player+0xB2))`，late-tech `player+0x59D` 使前者歸零，
   raw personality 3 使前者 `+2`，raw personality 4 使後者 `+2`。
8. `sub_E2710 @ 0xE2710` 是上述帝國欄位的 producer：它把各殖民地
   `+0xE9` 累加後寫入 signed word `player+0xAA`（工業），並把殖民地
   `+0xEB` 及其他研究來源累加、上限夾至 `32767` 後寫入
   `player+0xAC`（研究）。`sub_D66B3` 比較
   `research*researchWeight` 與 `industry*industryWeight`；前者較小時
   繼續加入科學家，否則轉入加入工人的循環。

## 已證實的排序鍵

- `sub_D5FA9 @ 0xD5FA9`：race slot `8`／`9` 排在一般種族之前；同類回傳
  `0`，沒有第二鍵。
- `sub_D614D @ 0xD614D`：封鎖分支以
  `food(raw 0)-industry(raw 1)` 由大到小排序；同分時 food 由小到大。
- `sub_D6315 @ 0xD6315`：未封鎖的一般排序以
  `research(raw 2)-industry(raw 1)` 由大到小；同分時 industry 由小到大，
  再以 race slot 由小到大。
- `sub_D63A6 @ 0xD63A6`：只有 `colony+0xDD>0` 才使用；以
  `food+industry-2*research` 由大到小，接著 industry 由大到小、research
  由小到大、race slot 由小到大。`colony+0xDD<=0` 時委派 `sub_D6315`。

以上產出鍵都由 `sub_DDFD3` 對同一名 colonist 分別套 raw job 後取得，
不是殖民地平均值。`PopulationGroups` 以 race slot、職務、prisoner 與逐種族
產出 profile 聚合；同群人口在這些排序鍵上等價，因此足以無損表示目前已見的
排序輸入，不必為排序目的展開成逐人陣列。

## 封鎖分支的已知邊界

`sub_D61E7` 先套 `sub_D5FE1` 與 `sub_D614D`。`colony+0xDD>0` 時，
每次改職後以 `sub_E1D59` 重算，並比較 `2*signed(colony+0xE7)` 與
`byte colony+0xFC..+0xFF` 的需求總和；不足時在目前農夫數尚未達
`colony+0xE0` 前，依排序把下一名可改派人口變成農夫，否則把末端候選變成
工人並重跑。`colony+0xDD<=0` 時，`sub_23DFE` 的回傳值決定把一般可改派人口
全設為 raw `2` 或 raw `1`；但 `sub_23DFE` 已由獨立稽核證實是事件殖民地
filter，這條呼叫不可改名成一般「無農業判斷」。

上述第 1–8 點是原始指令、caller 與資料流可回查的已證實結論。

## 對 remake 的直接反證

`internal/engine.ApplyAIEconomy` 目前逐殖民地呼叫
`Decider.ColonyJobs`，每個殖民地一次得出三種總數。`RemakeDecider` 則先餵飽全體，
再依設計性 Industry／Research weights 切分餘數。原版是單人口排序、
邊際產出、全帝國迭代與每步重算；因此不能藉由讓
`NewDecider(ModeOriginal)` 回傳同一個逐殖民地 helper 就宣稱原版 AI。

## 待閉合

- 封鎖狀態的 producer 與 `ColonyState` 垂直表示；目前 `.GAM` 的 `Star.Blockaded`
  已可解析，但 engine／shell 尚未把它接到 AI 殖民地職務分流。
- `colony+0xDD/+0xE0/+0xE7/+0xFC..+0xFF` 的完整欄位契約，以及
  `sub_23DFE` 在 `+0xDD<=0` 這條事件耦合分支的玩家可見理由。
- `sub_D5FE1` 已證實只把 Android／Natives 計入前置區間；一般 race slot
  全部留在後續排序範圍。`player+0x8B6`（Tolerant）、PRISONER bit 10 與
  `colony+0x12F` raw `2/3` 只影響 `AI record+0x0C` 的策略計數，不是把該人口
  從 qsort／改職範圍移除。`+0x0C` 的非職務消費端仍可另案追查，但不再阻塞
  未封鎖職務切片。
