# 領袖回合、ETA callback 與 AI 任命稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 外部符號索引：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱只供導覽，
  不取代原始函式名、位址或指令。
- 工具：IDA Pro 9.4／Hex-Rays 9.4.0.260610；位址為 IDA linear、DOS/4GW LE object #1。
- 可重生匯出：[`tools/ida/audit_leader_turn_chain.py`](../../tools/ida/audit_leader_turn_chain.py)
  與 [`evidence/leader-turn-chain-ida-20260828.json`](evidence/leader-turn-chain-ida-20260828.json)。
  匯出保存原始 bytes、operand、函式邊界、直接 caller、欄位讀寫端與兩個 AI heap buffer
  pointer slot 的 producer／consumer。
- 範圍：靜態資料流。`memset_`、配置器、存檔 I/O、stack probe 與其他 compiler／runtime
  helper 只作邊界定位，不進玩法分母。

## 已證實：逐回合狀態機與 ETA callback

`Deassign_Officer_ @ 0x933F2` 與 `Decrement_Officer_ETA_ @ 0x934CF` 是相鄰但不同的函式。
後者反向掃描 67 筆、每筆 `0x3B` bytes 的領袖記錄：

- `status +0x39 == 4`：每回合遞增 signed `ETA +0x37`；到 `30` 時呼叫
  `Deassign_Officer_`。
- `status == 1` 且 signed ETA 大於 0：每回合減 1；只有剛好降到 0 且
  `type +0x23 == 1` 時，才以 assignment/star `+0x35` 與 owner `+0x3A`
  呼叫 `Star_Colony_Calculation_ @ 0xE2AB1`。

`Star_Colony_Calculation_` 的 callback 不是只刷新一位領袖，也不以領袖 ID 作輸入。它掃描
該星系 record 的六個 planet slot `+0x48,+0x4A,...,+0x52`；每個有效 planet 對到的 colony
必須 owner 相符且 `colony+0x06 == 0`，才依序呼叫：

1. `Pre_Import_Computing_ @ 0xE1D59` 重建該殖民地衍生值；
2. `Pass_Out_Imports_ @ 0xDF8F0` 重新分配該帝國 imports；
3. 六槽處理完後，`Update_Player_Stats_ @ 0xE2710` 重算帝國彙總。

因此原版 callback 的玩家可見契約是「管理領袖抵達後，重算同星系合格殖民地與帝國彙總」。
remake 現行 `RawLocation=1`、撤銷／重套單一領袖增量與全殖民地士氣刷新，是較窄的資料模型
近似，不能再寫成原版 exact callback。

`Deassign_Officer_` 先呼叫 raw `sub_979A0` 調整領袖效果，再清除實際任命：管理領袖會清
star record 的對應 officer slot，艦長會清 ship `+0x74`；最後把 ETA 設 0、status 設 `-1`、
assignment 與 owner 都設 `-1`。這條清理鏈和上面的 ETA 遞減函式不可互換名稱。

## 已證實：AI 任命 dispatch

`Assign_Unassigned_Leaders_ @ 0xD73D4` 由高索引向低索引掃描 67 位、owner 相符的領袖：

| raw status | raw type | 呼叫 | 玩家可見用途 |
|---:|---:|---|---|
| 0 | 0 | `Assign_Fleet_Officer_ @ 0xD6FDA` | 首次指派艦長 |
| 0 | 非 0 | `Assign_Star_Officer_ @ 0xD7171` | 首次指派管理領袖 |
| 1 | 0 | `Reassign_Fleet_Officer_ @ 0xD7078` | 同位置艦艇間重派艦長 |
| 1 | 非 0 | 無 | 已任命管理領袖不逐回合重派 |

這推翻舊文件的「type 1、status 1 交給 `D7078` 做殖民地重派」；`D7078` 只處理艦長。

## 已證實：艦長候選排序

`Assign_Fleet_Officer_` 只接受 owner 相符、raw state `ship+0x64 <= 2`、`ship+0x11 == 0`
且 `ship+0x74 == -1` 的船。新候選必須在 byte `+0x10/+0x12/+0x16/+0x15` 全部不小於
目前候選，且 signed word `+0x72` 嚴格較大，才會取代；找到後交給 `sub_98C23` 寫入任命。

`Reassign_Fleet_Officer_` 從目前指派船出發，只考慮 `Ships_In_Same_Place_` 為真的同位置船，
其餘資格與五欄支配比較相同；只有勝出船改變才呼叫 `sub_98C23`。五個 raw ship 欄位的排序
與 signedness 已證實；本切片沒有把它們猜命名成戰力、裝甲或艦價。

## 已證實：管理領袖選星完整分數

`Assign_Star_Officer_` 先把 72 個 signed dword 星系分數清零，再掃描 owner 相符、
`colony+0x06 == 0` 且該星系方向 officer slot 為 `-1` 的殖民地。領袖 `specialSkills +0x2A`
使用每項兩個 bit 的 tier 配對；任一 tier bit 成立即加下列值：

| raw bit pair | 技能索引對照 | 每座殖民地／星系加分 |
|---|---|---|
| `0x1/0x2` | Environmentalist | 若 AI colony cache `+4 != 0`，加 signed `colony+0x08` 污染清理成本 |
| `0x4/0x8` | Farming Leader | 若 `colony+0xDD != 0`，加 signed `colony+0xE7` 食物總產出 |
| `0x10/0x20` | Financial Leader | 加 unsigned `colony+0x0A` 人口 |
| `0x100/0x200` | Labor Leader | 加 signed `colony+0xE9` 工業總產出 |
| `0x400/0x800` | Medicine | 加 `max(Colony_Race_Pop_Limit_(colony)-population,0)` |
| `0x1000/0x2000` | Science Leader | 加 signed `colony+0xEB` 研究總產出 |
| `0x4000/0x8000` | Spiritual Leader | 政府 raw `player+0x89F / 2 != 3` 時加人口；Unification 不加 |
| `0x10000/0x20000` | Tactics | 加該星系／owner 快取 `+2` 與 `+3` 的 unsigned 人數 |

技能名稱的索引順序由原版每技能 2-bit layout 與既有 HERODATA／UI 技能列舉交叉驗證；
上表的 raw mask、分數來源、signedness 與分支則直接來自 `0xD7223..0xD7353`。

殖民地 AI cache 是 `Compute_AI_Data_ @ 0xD3D34` 配置的 `0x6D6`-byte heap buffer，
每座殖民地 7 bytes。`+1` 在 `0xD492C..0xD4931` 直接接收
`Colony_Race_Pop_Limit_ @ 0xE0C1D` 回傳值，不再是未知 max-pop 猜測；`+4` 的 raw
producer 位於 `0xD4967..0xD4970`，其污染 consumer 語意已由獨立殖民地工業切片閉合。

星系／帝國 AI cache 是另一個 heap buffer，每星系 49 bytes、每帝國 6 bytes。
`Find_Players_In_Range_ @ 0xD3A68` 以星系座標與 player `+0x324` 生成正常航程及 1.5 倍航程
兩組布林表；`Compute_Star_Player_Range_Info_ @ 0xD3BA0` 再依 star `+0x38` 的帝國存在遮罩
與方向正式狀態 `player+0x627+other` 分桶。Tactics 實際讀的兩欄是：

- `+2`：正常航程內，方向正式狀態 raw `4..6` 的其他帝國數；
- `+3`：不在正常航程、但在 1.5 倍航程內的其他帝國數。

其餘 `+0/+1/+4/+6` 的 producer 也保留於證據匯出，但不是本任命分數 consumer 的輸入，
不為了「補完整 struct」擴張本切片。

掃描完後，原版由最高星系索引開始，只有遇到嚴格更高分才換候選，因此同分保留較高索引；
勝出分數必須非 0，才呼叫 `sub_98C87` 寫入任命。

## 證據邊界與 remake 差異

- **已證實**：ETA／limbo 狀態轉移、精確 callback 參數與同星系重算 consumer、四路 AI dispatch、
  艦長五欄支配排序、管理領袖每個 raw skill pair 的完整分數、兩個 AI cache producer 與 tie-break。
- **強推論**：外部符號名稱只作導覽；正式狀態 raw `4..6` 的逐一 UI 名稱不由本切片重複命名。
- **未知但不阻塞本列**：原版 PRNG 的逐 bit 序列、五個 raw ship 排序欄位的正式設計名稱，
  以及沒有被任命評分讀取的 AI cache 其他欄位。
- **remake 尚未對齊**：目前 ETA callback 只重套 typed 領袖效果；艦艇與星系排序使用較少的
  typed 欄位，沒有上述 raw 支配比較、Tactics 航程／外交壓力分數與嚴格 tie-break。依 RE-first
  gate 本輪只完成知識庫，不修改 Go／Ebitengine 行為，也不建立 READY spec。
