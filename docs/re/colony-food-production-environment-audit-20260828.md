# 殖民地食物基礎、環境與總產出稽核（2026-08-28）

## 問題與證據契約

既有文件分別知道 `colony+0xDD` 是半食物快取、`+0xE0` 是 AI 職務分配 gate，以及
`+0xE7` 是食物總產出，但沒有把三個 writer、建築、科技、氣候與 `DE280` 串成同一條
玩家可見鏈。本切片只補 RE，不修改 Go，也不提前產生 READY spec。

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`tools/ida/audit_colony_turn_chain.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱不作證據。
- 可重生證據：
  [`evidence/colony-turn-chain-ida-20260828.json`](evidence/colony-turn-chain-ida-20260828.json)。
- `memset_`、`sprintf_`、`Format_Quotient_` 與字串複製只服務 local breakdown／報表；依
  runtime 停止線保留 callsite，不納入玩法分母。

## 每農夫環境基礎：`Colony_Food2_Per_Farmer_ @ 0xDE03E..0xDE0C6`

函式由 colony 的 planet index 讀 `planet+0x0B`，先建立 half-unit 基礎：

```text
foodHalfBase = 2 × uint8(planet+0x0B)
```

接著依固定順序套用：

1. 基礎為 0 且 owner `player+0x134 == 3` 時改為 2。`0x134-0x117=29`，與受版控科技
   application ordinal 29 及獨立研究狀態表共同閉合為 Biomorphic Fungi。
2. `colony+0x164`（building raw 46，Weather Controller）存在時加 4 half-unit。
3. `colony+0x13A`（building raw 4，Astro University）存在時加 2 half-unit。

後兩項只在目前值大於 0 時套用；因此 Weather Controller／Astro University 不會單獨讓完全
不可耕作的環境開始產糧，Biomorphic Fungi 則會先把零值抬成 2，再允許建築加成。
building raw ID 與名稱由 building base `+0x136`、AI 建造跳表及 `OrigBuildingID` 交叉驗證。

`Pre_Import_Computing_ @ 0xE1D6F..0xE1D7B` 以 owner 呼叫此函式並把 low byte 寫入
`colony+0xDD`。所以 `+0xDD` 是**尚未加入 Farming、Aquatic、重力、士氣、政體或領袖修正**的
每農夫 half-unit 環境／建築快取，不等同最終每名 colonist 產出。

## 混合人口的食物基礎：`sub_DE0C6 @ 0xDE0C6..0xDE22C`

`DE280(job 0)` 透過 dispatch 對每名農夫呼叫此函式：

- owner 且不要求 breakdown 時可直接使用 `colony+0xDD`；外族 slot 會重跑 `DE03E`。
- 一般 slot `<8` 加 `player[slot]+0x8A1` Farming raw 值。
- Android slot 8 固定加 6，Natives slot 9 固定加 4。
- Aquatic `player[slot]+0x8AB` 只在有效氣候 raw 4／5／6／8 的分支重映氣候食物值；
  修正為 `2×(aquaticClimateFood-normalClimateFood)`，因此仍使用 half-unit 尺度。
- 函式以 colony、owner slot 與 race slot 維護十格 cache；`memset_` 只清該 cache，不是玩法。

此結果再由已閉合的 `DE280` 同時套重力、prisoner、士氣、政體、殖民地領袖、封鎖與 AI
難度，不能把 `colony+0xDD` 直接乘農夫數當作混合人口最終食物。

## 有效氣候與可耕作 gate：`Colony_Environmental_Stuff_ @ 0xE1CED..0xE1D59`

函式先把 `planet+0x08` climate ordinal 寫入 `colony+0xE2`。raw climate 1（Radiated）且
raw building 23／24／28 三種行星護盾任一存在時，有效氣候改寫為 raw 2（Barren）。
`0xE1D1E` 的 `cmp al,al` 必然相等，後續 `jnz` 不可達，不能虛構第四個 shield gate。

之後以 owner 呼叫 `DE03E`：

```text
colony+0xE0 = (foodHalfBase != 0) ? 0xFF : 0x00
```

因此 `+0xE0` 是可耕作布林 gate，不是最多農夫數、人口容量或食物數值。AI 封鎖職務分配將它
作為「是否還能指派農夫」的 gate，是獨立玩家可見 consumer。

## 食物總產出：`Colony_Food_Production_ @ 0xDE664..0xDEB1D`

normal caller 是 `Pre_Import_Computing_ @ 0xE1DC9`。人口 `colony+0x0A == 0` 或
`Event_Check_Space_Anomaly_ @ 0x2341E` 命中時，`colony+0xE7=0`。其餘路徑為：

```text
farmerFood = DE280(colony, job=0)
hydroponic = (colony+0x14B raw 21 Hydroponic Farm) ? 2 : 0
subterranean = (colony+0x161 raw 43 Subterranean Farms) ? 4 : 0
colony+0xE7 = farmerFood + hydroponic + subterranean
```

兩座建築是殖民地固定整數食物，不乘農夫數。正常 pre-import 呼叫不傳報表 buffer；後續
`Colony_Replicators_ @ 0xDF66F` 才把 `colony+0x114` 複製食物加入 `+0xE7`。帶報表 buffer 的
顯示路徑會在格式化尾端把既有 `+0x114` 加入顯示總數，這是重建同一玩家可見總和，並非正常
回合再執行一次複製或重複加值。

## 閉合結論與 remake 邊界

- **已證實**：`+0xDD` half-unit writer、Biomorphic Fungi、Weather Controller、Astro
  University、逐 race Farming／Aquatic、Android／Natives、有效氣候、三種護盾、`+0xE0`
  可耕作 gate、空殖民地／anomaly gate、Hydroponic／Subterranean 固定食物及 `+0xE7` writer。
- **強推論**：`planet+0x0B` 的高層欄名採氣候食物基礎；writer、值域與多個 consumer 已一致，
  但本文件仍保留 raw operand，不以欄名取代原位址。
- optional 598-byte breakdown 的逐格 UI 文案索引未全數命名，但不改變總產出，不列為玩法公式
  缺口。
- remake 的 owner-only `FoodPerFarmer` 快取若直接用於混合人口，會漏掉逐 race Farming、
  Aquatic 與特殊 slot；這項差異留待 RE gate 關閉後由 READY spec 處理。
