# 殖民地工業消耗、產出與污染稽核（2026-08-28）

## 問題與證據契約

既有文件已局部證實 Cybernetic／Android 的半工業消耗與 `colony+0x08` 清污成本，但沒有把
工業維護、工人產出、固定建築、污染、食物複製機及建造消費端接成同一條鏈。本切片只補 RE；
不修改 Go。`Colony_Job_Production_` 的共用 modifier 後續已由同日增補證據閉合，仍須經獨立
READY spec 才能授權實作修正。

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`tools/ida/audit_colony_turn_chain.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱不作證據。
- 可重生證據：
  [`evidence/colony-turn-chain-ida-20260828.json`](evidence/colony-turn-chain-ida-20260828.json)。
- `memset_`、`sprintf_`、`Format_Quotient_` 與共用 `locret_E1F97` epilogue 是 runtime／報表
  實作；只保留 callsite，不納入玩法分母。Hex-Rays 對 optional breakdown 區有未初始化區域變數，
  本結論以 normal call 的原始指令、欄位與 consumer 為準。

## 工業消耗：`Colony_Industry_Maintenance_ @ 0xDF546..0xDF66F`

唯一 caller 是 `Pre_Import_Computing_ @ 0xE1DB7`。函式清除 `colony+0x10C[8]` 與
`+0x100..+0x103`，再逐筆掃 packed colonist：

- 一般 player slot `<8`：只有 `player[slot]+0x8B0`（Cybernetic）每人增加 1 個半工業單位；
- Android slot 8：每人增加 2 個半工業單位；
- Natives slot 9：不消耗工業。

分類欄位為 owner `+0x100`、Android `+0x101`、外族 `+0x102`、prisoner `+0x103`；一般
player slot 另累積到 `+0x10C[slot]`。總整數消耗寫入：

```text
colony+0xF0 = ceil((ownerHalf + androidHalf + alienHalf + prisonerHalf) / 2)
             = (sumHalf + 1) / 2
```

`Event_Check_Space_Anomaly_ @ 0x2341E` 命中時 `+0xF0=0`；pre-import 開頭已清 derived 區，
所以分類 bytes 不會沿用前一回合。`memset_` 只負責清暫存，不是玩法公式。

## 工業產出外層：`Colony_Industry_Production_ @ 0xDEE1B..0xDF518`

normal caller 是 `Pre_Import_Computing_ @ 0xE1DC0`；`0xDF531` 是另一個 wrapper caller。
人口為零或 anomaly 命中時 `colony+0xE9=0`。正常路徑先以
`Colony_Job_Production_ @ 0xDE280`、job 1 計算逐工人工業，再加入：

| 原始旗標／資料 | gross 加成 |
| --- | ---: |
| `+0x13D` raw 7 Automated Factory | +5 |
| `+0x15A` raw 36 Robo Mining Plant | +10 |
| `+0x142` raw 12 Deep Core Mine | +15 |
| `+0x158` raw 34 Robotic Factory | 礦產表 `[5,8,10,15,20]` |
| `+0x157` raw 33 Recyclotron | +1 × colony population |

Robotic Factory 表位於 `byte_DD4DC @ 0xDD4DC`，bytes 為 `05 08 0A 0F 14`。上述 building
名稱由 building base `+0x136` 與獨立 `OrigBuildingID` 對映交叉驗證，不從效果猜名。

`Colony_Job_Production_` 是食物／工業／研究共用 helper，會逐 packed colonist 套來源 slot、
重力、prisoner、政府、領袖、封鎖及 AI difficulty modifier。該共用鏈已於 2026-08-28 由
[`colony-government-output-audit-20260825.md`](colony-government-output-audit-20260825.md)
補齊五級難度表與封鎖 consumer，函式 `0xDE280..0xDE664`、三個 caller、直接 helper、原始
operands 與 breakdown 均保存於同一份可重生證據。

## 原版污染基數與清理公式

原版不是拿上述 gross 全額計污染。`0xDEFC6` 建立的 polluting base 只有：

```text
pollutingBase = workerJobProduction + roboticFactoryMineralBonus
```

因此 Automated Factory `+5`、Robo Mining Plant `+10`、Deep Core Mine `+15` 與
Recyclotron `population` 都加入 gross，卻不進 polluting base。這直接反證 remake 目前把所有
`FlatIndustry` 一起送入污染公式的做法。

若 raw 13 Core Waste Dump 存在，`colony+0x08=0`。其餘路徑依序為：

1. 以 planet size 索引 `byte_DD4E1 @ 0xDD4E1`；bytes `02 04 06 08 0A`，即
   Tiny..Huge 容忍值 `[2,4,6,8,10]`。
2. player technology 113 Nano Disassemblers status `3` 時容忍值乘 2。
3. raw 32 Pollution Processor 把 polluting base 除 2，採 signed 向零截斷。
4. raw 5 Atmospheric Renewer 再除 4；兩者並存剩原值 1/8。
5. 所在星系 officer 若有 Environmentalist：普通／進階分別減少
   `10×(level+1)%`／`15×(level+1)%`，使用整數除法。
6. `excess = reducedPollutingBase - tolerance`；正值時
   `cleanup = ceil(excess/2) = (excess+1)/2`，否則 0。
7. 若 colony 混合人口含 Tolerant 一般 slot 或 Android，依人口比例調整：
   `cleanup = round(cleanup × nonTolerantPopulation / totalPopulation)`，原始式為
   `((total-tolerant)*cleanup + total/2)/total`。Natives 不在 tolerant count。

最後：

```text
colony+0x08 = max(cleanup, 0)
colony+0xE9 = grossIndustry - colony+0x08
```

這裡的 `+0xE9` 是污染後、食物複製機前的工業 word，尚未扣 `+0xF0` 工業消耗。

## 玩家可見建造 consumer：`Apply_Production_ @ 0xE36DF..0xE3E9A`

唯一 caller 是一次性殖民地套用鏈 `sub_E3F6E @ 0xE3FC8`。`0xE37DD..0xE382A` 先把
`colony+0x125` 與一次性 carry `+0x12C` 加入產品進度並清 `+0x12C`，再計算：

```text
availableIndustry = max(int16(colony+0xE9) - uint8(colony+0xF0), 0)
productProgress += availableIndustry
```

所以 Cybernetic／Android 工業消耗確實會減少建造產能，但扣除點在一次性
`Apply_Production_`，不是 `Colony_Industry_Production_`。食物複製機先依分類消耗保留量限制
換算並直接減 `+0xE9`，之後建造再扣 `+0xF0`，順序不可合併成一個無分類的早期 subtraction。

## 結論、反例與剩餘邊界

- **已證實**：工業消耗的半單位分類／整數總量、五種 fixed gross 加成、兩張五項表、污染
  建築順序、Nano、officer、逐人口 Tolerant 比例、奇數 excess 向上取整、`+0x08/+0xE9`
  writer，以及 `Apply_Production_` 的 `max(E9-F0,0)` consumer。
- **推翻舊斷言**：`PollutionCleanupCost` 的奇數 excess 不是向下取整；污染基數也不是完整
  gross／全部 `FlatIndustry`。現行 remake 兩處都與原版指令矛盾，須等 RE gate 關閉後進
  READY spec 修正。
- **仍未閉合**：optional breakdown 的逐格 UI 文案索引、`Colony_Industry_To_Tax_` 與 AI
  多個 `+0xF0` consumer。breakdown 文案不改變已閉合的玩法總產出；其餘 consumer 不能因
  外層污染鏈閉合而升格。
- **RE-first gate**：本輪只補知識庫與差異清單，不修改 Go 或測試期望。
