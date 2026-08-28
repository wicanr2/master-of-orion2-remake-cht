# 食物複製機與殖民地 BC 維護稽核（2026-08-28）

## 問題與證據契約

既有 remake 依手冊「two-for-one、1 BC per food、as needed」實作 2 工業換 1 食物，
並為 Cybernetic 的半食物需求新增跨回合半 BC 餘數；但後者沒有執行檔證據。本切片追查
原版實際換算、人口群組優先序、BC consumer 與氣候倍率。

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`tools/ida/audit_colony_turn_chain.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱不作證據。
- 可重生證據：
  [`evidence/colony-turn-chain-ida-20260828.json`](evidence/colony-turn-chain-ida-20260828.json)。
- 函式尾端跳到 `locret_E1F97` 是共用 epilogue；不追其 compiler/runtime 內部，也不納入玩法分母。

## `Colony_Replicators_ @ 0xDF66F..0xDF8C1`

唯一 caller 是 `Pre_Import_Computing_ @ 0xE1DD7`。函式先把 `colony+0x114` 清零，只有
`colony+0x146 != 0` 才執行；building flag base `+0x136` 加 raw ID 16，與
`OrigBuildingID` 的 Food Replicators 獨立對映一致。

相關 derived 欄位沿用已閉合的殖民地產出／消耗順序：

| 欄位 | 已證實用途 |
| --- | --- |
| `+0xE7` | 食物產出 |
| `+0xE9` | 工業產出 |
| `+0xFC..+0xFF` | owner、外族、prisoner、Natives 食物消耗 |
| `+0x100..+0x103` | owner、Android、外族、prisoner 工業消耗 |
| `+0x114` | 本回合複製出的整數食物；也是 BC 維護 consumer 的輸入 |

函式依四段優先序補食物赤字：

1. owner 食物；只保留 owner 工業消耗後可用的工業。
2. owner＋外族食物；保留 owner、Android、外族工業消耗。
3. owner＋外族＋prisoner 食物；保留四類工業消耗。
4. 再納入 Natives 食物；仍保留四類工業消耗。

每段只處理正赤字與正可用工業，並把可轉換量壓成偶數工業。最後：

```text
replicatedFood = spentIndustry / 2
colony+0x114    = replicatedFood
colony+0xE9    -= spentIndustry
colony+0xE7    += replicatedFood
```

所以執行檔精確行為是 2 工業換 1 個**整數**食物，只補依優先序累積後的赤字，不會製造盈餘。
`spentIndustry` 被強制為偶數，Cybernetic 形成的單一半食物赤字不會被複製機單獨補滿；原版 record
也沒有在此保存半食物或跨回合餘數。

## 氣候 producer 與行星護盾

`Colony_Environmental_Stuff_ @ 0xE1CED..0xE1D59` 從行星 record `+0x08` 把 climate ordinal
寫入 `colony+0xE2`。若 climate 是 raw 1（RADIATED），且 raw building 23 Barrier Shield、
24 Flux Shield 或 28 Radiation Shield 任一存在，就把有效 climate 改成 raw 2（BARREN）。
`0xE1D1E` 的 `cmp al, al` 永遠相等，後續 `jnz` 不可達；它不新增第四個 gate。

`byte_DD4BA @ 0xDD4BA` 的十個 bytes 是：

```text
32 19 00 19 00 00 00 00 00 00
```

依既有 `PlanetClimate` ordinal（TOXIC..GAIA）即維護倍率加成
`[50,25,0,25,0,0,0,0,0,0]`。此表亦被原版 AI 行星估值讀取，形成獨立 consumer 交叉證據。

## `Colony_BC_Maintenance_ @ 0xE094F..0xE0A14`

唯一 caller 是 `Pre_Import_Computing_ @ 0xE1E16`，在複製機之後。人口 `colony+0x0A == 0`
或 `Event_Check_Space_Anomaly_ @ 0x2341E` 命中時，`colony+0xF2` 維護費為零；否則：

1. 掃描 `colony+0x136` 起的 49 個 building flags。
2. 每個存在的建築加上 19-byte building record 中的 signed 16-bit 維護費。
3. 加上 `colony+0x114`，即每個複製食物 1 BC。
4. 乘 `100 + climateMaintenanceModifier[colony+0xE2]`。
5. 加 5000 後除以 10000，對正值等價於四捨五入至整數 BC。

```text
baseBC = buildingMaintenanceSum + replicatedFood
colony+0xF2 = round(baseBC * (100 + climateModifier) / 100)
colony+0xF9 = colony+0xF2 - colony+0xED
```

`colony+0xED` 是同一 pre-import 鏈較早產生的殖民地 BC 產出；`+0xF9` 因此是此處使用的
signed 淨 BC 差。BC 費用會與全部建築維護一起受氣候倍率及最後一次整數四捨五入影響，並非先為
每個半食物獨立累積半 BC。

## 結論、反例與 remake 邊界

- **已證實**：Food Replicators raw flag、2:1 整數換算、四段人口／資源優先序、只補赤字、
  `+0x114` writer／BC consumer、十項氣候表、三種護盾把 Radiated 視為 Barren、建築＋複製費用
  的整體氣候倍率與四捨五入、空殖民地／事件 gate。
- **推翻舊斷言**：remake 的半食物複製、半 BC 跨回合餘數不是原版近似，而是與 `DF66F` 的
  偶數工業 gate 及 `E094F` 的整數維護公式矛盾。它必須在未來 READY spec 中移除或隔離成明示的
  現代選項，不能留在「原版忠實」模式。
- `Colony_Environmental_Stuff_` 的 `colony+0xE0` 已由後續食物切片閉合為可耕作 gate；
  `Colony_BC_Production_` 的完整來源後續已由
  [`colony-bc-production-tax-audit-20260828.md`](colony-bc-production-tax-audit-20260828.md) 閉合。
  anomaly event 的其他資源欄位效果不因本切片升格。食物環境證據見
  [`colony-food-production-environment-audit-20260828.md`](colony-food-production-environment-audit-20260828.md)。
- **RE-first gate**：本輪不修改 Go、不寫 READY spec。差異先登記到 parity matrix／WORKLIST，
  等玩家可見玩法 RE 分母閉合後再依正式 spec 修正與做同狀態測試。
