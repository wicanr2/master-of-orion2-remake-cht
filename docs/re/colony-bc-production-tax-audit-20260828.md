# 殖民地工業轉稅、BC 產出與帝國彙總稽核（2026-08-28）

## 問題與證據契約

本切片閉合 `Pre_Import_Computing_` 尾段尚未確認的工業轉稅與殖民地 BC producer，並追到
`Update_Player_Stats_` 的帝國收入 consumer。它只補 RE 知識庫；依 RE-first gate，不修改
Go／Ebitengine 行為或既有測試期望。

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4／Hex-Rays 9.4.0.260610、
  `tools/ida/audit_colony_turn_chain.py`；位址均為 DOS/4GW LE object #1 的 IDA linear EA。
- 外部導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`。名稱只供導航，
  函式邊界、指令、欄位 writer／consumer 才是證據。
- 可重生匯出：
  [`evidence/colony-turn-chain-ida-20260828.json`](evidence/colony-turn-chain-ida-20260828.json)。
- `strcpy_`、`sprintf_` 與共用 epilogue 只負責報表字串或 C runtime 控制流，不納入玩法分母。

## 工業轉稅：`sub_E08F6 @ 0xE08F6..0xE094F`

唯一 caller 是 `Pre_Import_Computing_ @ 0xE1E06`。函式先讀已算出的工業與工業維護：

```text
availableIndustry = signed colony+0xE9 - unsigned colony+0xF0
```

若結果小於等於零，`colony+0x127` 寫 0。否則讀 owner 的 unsigned `player+0x31` 百分比，
用 signed `idiv 100` 向零截斷：

```text
taxBC          = availableIndustry * taxRatePercent / 100
colony+0x127   = taxBC
colony+0xE9   -= taxBC
```

因此稅收是每座殖民地在扣除工業人口維護後逐座取整，轉換比 1:1；被抽走的工業會在 BC producer
與後續 `Apply_Production_` 前直接從 `+0xE9` 扣除。它不是帝國總工業一次取整，也不是只在 UI
顯示的估算值。稅率選單／writer 已由
[`ai-tax-rate-audit-20260828.md`](ai-tax-rate-audit-20260828.md) 閉合：人類 UI 寫入，AI 新局為 0%。

## 殖民地 BC：`sub_E03F1 @ 0xE03F1..0xE08C8`

callers 是報表路徑 `0xE08E1` 與 pre-import `0xE1E0F`。人口為零或
`Event_Check_Space_Anomaly_ @ 0x2341E` 命中時，`colony+0xED` 寫 0；其餘路徑以下列彼此
獨立的加項組成總額。

### 有機人口基礎收入與 AI 難度

人口記錄從 `colony+0x0C` 起、每筆 4 bytes；函式只計 low nibble race slot `< 8` 的人口。
Android 與 Natives 不進此基礎收入。依各人口 race slot 的 `player+0x8B6` 分成兩個暫存計數，
但玩法公式只使用兩者總和，故該分類對總 BC 無影響。

```text
quarterFactor = 2 * signed owner player+0x8A4 + 4
if owner player+0x28 != 100:
    quarterFactor += difficultyQuarterBonus[difficulty]

difficultyQuarterBonus = [0, 0, 1, 2, 3]
organicPopulation      = count(raceSlot < 8)
basePopBC              = round_positive(quarterFactor * organicPopulation / 4)
```

`byte_DD4E6 @ 0xDD4E6` 的原始 bytes 是 `00 00 01 02 03`。正數取整由加 2 後除 4 完成，
即最接近整數、半數進位。一般種族是每人口 1 BC；Money trait 每級改變每人口 0.5 BC；
非人類玩家另依五級難度增加每人口 0／0／0.25／0.5／0.75 BC，最後才對整座殖民地取整。

### 士氣

若有有機人口且政府 family 不是 raw `government/2 == 3` 的 Unification，函式以
`colony+0x07` 的 raw 士氣套用：

```text
moraleBC = round_signed(basePopBC * colonyMoraleRaw / 20)
```

正值加 10、負值採對稱路徑後除 20；raw 士氣每 1 點代表 5%。士氣只調整 `basePopBC`，不調整
工業稅、貿易品、Gold／Gems 或後述建築／科技／政體加項。

### 行星特產與共同基數

星系內行星 record `planet+0x0F` 的 raw special：4（Gold）加 5 BC、5（Gems）加 10 BC，
其餘加 0。令：

```text
specialBC = 5 if rawSpecial == 4 else 10 if rawSpecial == 5 else 0
B = basePopBC + specialBC
```

下列每個 bonus 都獨立以同一個 `B` 計算，不會彼此複利，也不作用於稅收、貿易品或士氣加項：

| 來源 | 原始 gate | 加項 |
| --- | --- | --- |
| Galactic Currency Exchange | `player+0x162 == 3` | `floor(B/2)` |
| Planetary Stock Exchange | `colony+0x153 != 0`，raw building 29 | `B` |
| Spaceport | `colony+0x15D != 0`，raw building 39 | `floor(B/2)` |
| Democracy | government raw 4 | `floor(B/2)` |
| Federation | government raw 5 | `floor(3B/4)` |
| Financial Leader，normal | `leader+0x2A & 0x10` | `round(B*(L+1)/10)` |
| Financial Leader，advanced | `leader+0x2A & 0x20` | `round(3B*(L+1)/20)` |

領袖 advanced bit 優先；`L` 由 `sub_94951` 的有效等級取得。normal 分子加 5 後除 10，
advanced 分子加 10 後除 20，皆為正數最接近整數、半數進位。

### 貿易品

目前產品 `colony+0x115` 為 raw `-37`、`-2` 或 `-1` 時，都進貿易品分支。
`Product_Name_ @ 0xCF398` 已交叉證實 `-37` 回傳 Trade Goods 名稱；`-2`／`-1` 是共用選項／
fallback sentinel，但 BC consumer 明確把三者視為同一類。使用的是已抽稅後的可用工業：

```text
tradeIndustry = signed colony+0xE9 - unsigned colony+0xF0
if tradeIndustry <= 0: tradeGoodsBC = 0
else if player+0x8B7 != 0: tradeGoodsBC = tradeIndustry
else: tradeGoodsBC = (tradeIndustry + 1) / 2
```

所以 Fantastic Trader 為 1:1；一般種族 2:1，但奇數工業**向上取整**，不是向下取整。

### 精確總式與 writer

```text
colony+0xED =
    basePopBC
  + moraleBC
  + specialBC
  + financialLeaderBC
  + galacticCurrencyExchangeBC
  + planetaryStockExchangeBC
  + spaceportBC
  + governmentBC
  + tradeGoodsBC
  + colony+0x127
```

報表字串逐項使用同一批暫存加項，形成獨立 consumer；總額不是由字串反推。

## 帝國收入 consumer：`sub_E2710 @ 0xE2710`

`Update_Player_Stats_` 掃描 owner 相符的 active 殖民地，把 `colony+0xED` 加入 BC 累計，並寫：

| player 欄位 | 已證實聚合值 |
| --- | --- |
| `+0xA6` | 人口 |
| `+0xA8` | 可外運食物產出 |
| `+0xAA` | 工業 |
| `+0xAC` | 研究 |
| `+0xAE` | gross BC income |
| `+0xB4` | `sub_E2000` 算出的維護費 |
| `+0xB2` | `+0xAE - +0xB4` 的本回合淨收入 |

函式另聚合非封鎖殖民地的食物產出與消耗；正食物盈餘對 Fantastic Trader 全額換 BC，其他玩家
先以 signed `/2` 向零截斷（正值即向下取整），再加入 `+0xAE`。這個帝國級食物盈餘轉換與
`E03F1` 的逐殖民地 BC 公式分開。

活動領袖的 Megawealth normal／advanced 分別直接加 10／15 BC；Researcher 另加研究。
條約／納貢陣列也在此彙總，已閉合的維護費、納貢與國庫 consumer 見
[`player-maintenance-audit-20260825.md`](player-maintenance-audit-20260825.md)：`sub_E2000` 寫
`+0xB4`，`sub_E4F49 @ 0xE4F8E..0xE4F95` 再把 `+0xB2` 加入 `player+0x32` 國庫。

## 對現行 remake 的已知差異

以下是新證據直接反證的現行行為；本輪只登記，不提前修程式：

1. `RunEmpireTurn` 對每人口半 BC 與 AI quarter bonus 分段向下取整；原版先聚合整座有機人口，
   再最接近整數取整。
2. remake 使用總 `Population`，無法保證排除 Android／Natives；原版明確只計 race slot `<8`。
3. remake 沒有把士氣套到有機人口基礎 BC；原版只調整 `basePopBC`。
4. remake 的 `IncomeBonusPercent` 把 Spaceport／Stock Exchange／Financial 套到稅、食物、
   貿易品與特產等大 subtotal；原版各自只以 `B = basePopBC + specialBC` 算獨立加項。
5. remake 在帝國層把政府與 Galactic Currency Exchange 乘到整筆收入；原版是逐殖民地、
   `B`-only 的加項，彼此不複利。
6. `TradeGoodsIncomeHalf` 對一般貿易品向下取整；原版對正奇數可用工業使用 `ceil(n/2)`。

## 完成邊界

- **已證實**：工業轉稅順序、欄位、逐殖民地取整與 1:1 回寫；有機人口收入、AI 難度表、
  士氣、Gold／Gems、五種科技／建築／政體加項、Financial Leader、貿易品 sentinel／換算、
  `colony+0xED` 總式，以及 `Update_Player_Stats_` 的收入／維護／國庫主 consumer。
- **強推論**：人口在 BC 報表中分成兩組的精確顯示名稱；兩組只以總和進玩法公式，不阻塞 remake。
- **未知且另案**：`Event_Check_Space_Anomaly_` 對其他資源的完整事件語意、條約／納貢每一個方向欄位
  的高層名稱，以及 `Update_Player_Stats_` 其餘非收入 cache。它們不改變本切片已閉合的算式。
- **排除**：`sprintf_`／`strcpy_` 內部、共用 epilogue、Watcom 除法與 stack helper；只保留其
  整數語意，不納入 RE 知識庫分母或 remake 範圍。
