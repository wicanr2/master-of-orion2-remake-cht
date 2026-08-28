# 人口成長完整 runtime 與回寫稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 外部符號索引：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱只供導覽。
- 工具：IDA Pro 9.4／Hex-Rays 9.4.0.260610；位址為 IDA linear、DOS/4GW LE object #1。
- 可重生匯出：[`tools/ida/audit_population_growth.py`](../../tools/ida/audit_population_growth.py)
  與 [`evidence/population-growth-ida-20260828.json`](evidence/population-growth-ida-20260828.json)。
  匯出保留原始函式名、位址、bytes、operand、caller、偽碼導覽與 raw option bytes。
- 範圍：1.31 靜態資料流。`memset_`、`isqrt_`、隨機、洗牌及 weighted-choice helper 的
  內部實作不重製；只保存會影響玩家結果的輸入／輸出契約。

## 已證實：逐族 rate producer

`Colony_Pop_Grows_ @ 0xE1839..0xE1CED` 對每個有一般人口的 player slot 重建 signed
`colony+0xC8[slot]`。人口為 0 時同時清空十筆 rate `+0xC8` 與十筆累積池 `+0xB4`；
時空異象 `Event_Check_Space_Anomaly_ @ 0x2341E` 命中時只把本回合 rate 清零，既有累積池保留。

食物／工業短缺的負 rate 先寫入 `+0xC8`，其 producer、半單位與分配優先序已由混合人口專題
閉合。本節補齊同函式後半的正 rate。對每個人口數 `P_r > 0` 且 race population limit
`M_r > P_total` 的一般 player slot：

```text
base_r = isqrt(2000 * P_r * (M_r - P_total) / M_r)

factor_r = 100
         + signed player[r]+0x8A0
         + technology_bonus(owner)
         + best_medicine_leader_bonus(colony)
         + event_bonus(colony)
         + housing_bonus(colony)
         + ai_difficulty_bonus(colony owner, race slot)

rate_r = trunc(factor_r * base_r / 100)
       + trunc(100 * P_r / P_total) if Cloning Center exists
```

最後把 `rate_r` 加到既有 signed shortage rate，而不是用正成長覆蓋短缺。
`isqrt_ @ 0x134C92` 的內部算法屬 runtime helper 停止線；輸入整數與向下平方根回傳契約則是
玩家公式的一部分。

### 各百分點 producer

| 項目 | 原始指令／欄位 | 已證實值 |
|---|---|---:|
| race growth | `0xE1B78..0xE1B83`, signed `player[r]+0x8A0` | 直接加在基準 100 |
| Microbiotics | `player[owner]+0x182 == 3` | `+25` |
| Universal Antidote | `player[owner]+0x1D8 == 3` | `+50`，命中時不再加 Microbiotics |
| Medicine tier 1 | officer `specialSkills+0x2B & 0x04` | `10 × (experience bucket+1)` |
| Medicine tier 2 | officer `specialSkills+0x2B & 0x08` | `15 × (experience bucket+1)` |
| population boom | `Event_Check_Population_Boom_ @ 0x23509` | `+100` |
| plague | `Event_Check_Plague_ @ 0x234B8` | `-200` |
| housing raw mode | `colony+0x115 == -3` | `40 × (colony+0xE9-colony+0xF0) / P_total` |
| AI difficulty | owner 非真人時的 `byte_199CB0` | raw difficulty `0..4` 直接加百分點 |
| Cloning Center | building slot 10，即 `colony+0x140 != 0` | 最後依 `100 × P_r / P_total` 逐族加入 |

科技位移由 player tech array 起點 `+0x117` 交叉驗證：`0x182-0x117=107`
是 Microbiotics，`0x1D8-0x117=193` 是 Universal Antidote。Cloning Center 並非整座殖民地
先加 100 再任意給一族；它逐族向零截斷，所以混合人口時各族固定項總和可能小於 100。

非 owner slot 若含 packed prisoner，且 `colony+0x12F == 0`、factor 為正，會先乘
`(該 slot 全人口-prisoner 人口)/該 slot 全人口`；這使正在滅絕的 captured population
不取得自然正成長。owner slot或沒有 prisoner 的 slot 不走此衰減。

## 1.31 的 reserved growth flag

`sub_16946E @ 0x16946E` 另測試 star record #0 的保留 byte `star[0]+0x0E` bit 3；命中會在
factor 增加 `150`。這不是 star 名稱字元：`Generate_Home_Worlds_ → sub_169245` 明確把
四個 raw option bit 合成到 `star[0]+0x0E`，再把相鄰 `+0x0D` 寫 0；
`Do_Change_Star_Name_ @ 0x922C2` 改名時會暫存並恢復 `+0x0E`，同樣把 `+0x0D` 固定為 0。

bit 3 的來源是 executable data `0x178755` bit 3。正式 1.31 輸入的該 byte 為 `0x00`，
全程式只有 `sub_16945B` 讀取、沒有直接 writer，因此標準 1.31 新局此 `+150` 不生效。
它的形狀符合由 patch／profile 改寫 binary data、再封裝進 save 的一般設定旗標，但本切片沒有
1.50 binary 可證明其正式設定名；故「1.31 預設為 0」是**已證實**，「patch-configured general
growth bonus」是**強推論**。它不應被猜成科技、種族或難度加成。

## 已證實：累積、滅絕與人口回寫

`Apply_Colony_Pop_Growth_ @ 0xE2DCA..0xE3456` 在正常人口殖民地依固定順序處理：

1. 時空異象命中或人口為 0 時整段不套用。
2. `colony+0x12F == 0`（既有叛亂 consumer 已交叉證實為滅絕政策）時，從 Natives slot 9，
   或非 owner 且 prisoner bit `0x400` 已設的人口中 reservoir 抽一筆刪除。只剩最後一筆或已無
   候選時把 `+0x12F` 改成 4；刪除最後一筆時送出 type 2 訊息。
3. 對所有一般 player slot、Android 8、Natives 9 先套 signed 負 rate：
   `pool[slot] += rate[slot]`。每低於 0 一個千點，只要該 slot 超過「必留一人」門檻便刪一人，
   再把 pool 加回 1000。必留 slot 依 owner、人口最多一般 slot、高索引 tie、Android、Natives
   的既有優先序選定。
4. 所有負池完成後，才處理一般 player slot 的正 rate。owner 先固定在順序第 0，其餘 slot
   走 `Shuffle_Sint_ @ 0xFE9F5`；Android／Natives 不在自然正成長名單。
5. 每當 pool 達 1000，先以 `Colony_Race_Pop_Limit_` 檢查容量；滿載時把 pool 夾為 999，
   否則新增一筆 packed colonist、扣 1000，race low nibble 寫該 slot、loyalty 寫 colony owner。

負成長刪除同 slot 候選時，非農夫 job bits `0x180` 優先；同優先群內用
`Random(++count)==1` reservoir sampling，最後以 packed array 尾筆回填洞。

## 已證實：新人口初始職務

新增人口的 job bits `7..8` 依下列優先序決定：

1. 若殖民地已啟用食物生產且供需／帝國缺糧／同 owner 無其他合格殖民地的 raw gate 要求補糧，
   選 job 0（farmer）。
2. 否則若 `colony+0xF0` 與 `+0xE9` 的工業需求 gate 要求補工，選 job 1（worker）。
3. 否則若 `Event_Check_Colony_Researching_ @ 0x23DFE` 命中，選 job 2（scientist）。
4. 否則統計現有人口的 job 0..2；只要 worker 或 scientist 非零，就以三欄現有數量交給
   `Get_Weighted_Choice_Int_ @ 0xFE96F`，否則預設 worker。

這些 job 分支發生在「哪一族先跨 1000」確定後，不會重排既有人口。完整 UI 名稱由職務
列舉交叉驗證；helper 內部亂數算法不在 remake 範圍，輸入三個權重與回傳索引契約保留。

## 證據分級與 remake 差異

- **已證實**：1.31 的 2000／40／1000 常數、逐族完整 factor、科技互斥優先序、Medicine、
  事件、AI 難度、Cloning Center 分攤、prisoner 衰減、滅絕、負池→正池順序、容量與職務回寫。
- **強推論**：`0x178755` bit 3 是 patch/profile 可改的 general growth option；1.31 預設關閉及
  它的 raw `+150` consumer 則已證實。
- **刻意不宣稱**：原版 PRNG 逐 bit、packed array 與 remake typed groups 的排列一致，或
  1.50 binary 使用相同 reserved flag。
- **remake 尚需重審**：混合人口的 Cloning Center 逐族截斷、科技只取最強、AI difficulty
  百分點、滅絕政策、prisoner 正成長衰減與新增人口 job priority 尚未全數依這條 1.31 鏈接線。
  依 RE-first gate 本輪不修改 Go，也不建立 READY spec。
