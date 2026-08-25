# 逐艦艦員經驗回合鏈靜態稽核（2026-08-24）

## 證據契約

- 原版輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，分析前 SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；只分析暫存副本。
- 工具：IDA Pro 9.4／IDAPython，SDK 940；位址均為 IDA linear address。
- 探針：`tools/ida/audit_ship_repair.py`，以 `MOO2_IDA_ROOTS` 指定 XP 函式；不改名、
  不套型別、不寫回資料庫。

## 主回合與逐艦門檻

**已證實**：`sub_14A27 @ 0x14A27` 的唯一直接 caller 是
`Next_Turn_Calc_ / sub_136B3 @ 0x136B3` 內的 `0x13760`。它清空八個玩家加成槽，
呼叫 `sub_14915 @ 0x14915` 彙總領袖加成，再固定掃描 500 筆 0x81-byte Ship record。

`0x14A58..0x14A76` 只處理 `Ship.Status < 5` 且 `Ship.Owner < 8` 的 record；Owner 由
`sub_77FF5 @ 0x77FF5` 做範圍驗證並作為八格加成索引。原版這條鏈沒有
`Design.Type == COMBAT_SHIP` 門檻，因此殖民船／運輸船／前哨船等有效玩家 ship record
同樣累積艦員經驗。這與靠港修復只接受戰鬥艦是兩套不同篩選，不可共用 predicate。

## 每回合經驗公式

**已證實**：`sub_149D5 @ 0x149D5` 先呼叫 `sub_1487A @ 0x1487A`，再執行：

```text
CrewExp = min(CrewExp + 1 + strongestInstructorBonus, 500)
```

欄位 `Ship +0x72` 由 `.GAM` layout 對回 `CrewExp`。上限來自
`word_17D186 @ 0x17D186 = 0x01F4 = 500`；`0x149F1` 讀舊值、`0x149F9` 加 1、
`0x149FA` 加玩家加成、`0x14A03..0x14A07` 夾上限、`0x14A1A` 寫回。

`sub_14915` 掃描 67 筆 0x3B-byte Leader record，只接受有效、已雇用且玩家索引 0..7
的領袖；檢查 raw skill 22／23，呼叫 `sub_94BB2` 取得值，並對每位玩家只保留最大值。
raw 22／23 與一般／進階 Instructor 的對映由既有 HERODATA／技能列舉交叉支撐為
**強推論**；「不累加、只取最強」則由 `0x149B8..0x149C0` 的 max 寫回為**已證實**。

## 同星系太空學院

**已證實**：`sub_1487A` 只在 `Ship.Status == 0` 時讀 `Ship.Star`，走訪該星五條行星軌道。
每找到一座屬於該船 Owner、且 Colony record `+0x15C` 非零的殖民地，就把 CrewExp 加一。
`+0x15C` 位於 0x169-byte Colony record 的 building flag 區，與 raw building ID 38
（Space Academy）相符。因此：

- 停在星系內的船，每座己方太空學院每回合額外 +1。
- 航行／其他非零 Status 的船仍取得基本 `1 + Instructor`，但不取得學院加成。
- 同星系多顆行星可各有學院，效果逐座累加。

上述控制流、五軌走訪、owner 比較與逐座 `inc word [Ship+0x72]` 均為**已證實**。

## 等級門檻與 Warlord

`sub_147E7 @ 0x147E7` 依序比較四個 stride-8 raw words：

| 位址 | 值 | 消費語意 |
|---|---:|---|
| `0x17D176` | 50 | Regular 門檻 |
| `0x17D17E` | 150 | Veteran 門檻 |
| `0x17D186` | 500 | Elite 門檻，同時是每回合 XP 上限 |
| `0x17D18E` | 1000 | Ultra Elite 原始門檻 |

每通過一個門檻等級加一；Player record `+0x8BD` 非零時再加一，最後夾到 raw level 5。
該 player flag 與 Warlord 的對映由既有 custom-race layout 支撐為**強推論**。因每回合 XP
在 500 封頂，一般種族最高 Elite；Warlord 在相同 XP 階梯上再平移一級，因此可達
Ultra Elite。raw level 的數值包含原版內部額外基底，remake 公開 enum 使用 0..4；
目前 `CrewLevelForXP` 的玩家可見五級對映與此資料流一致。

## remake 勘誤與邊界

- 現行每回合基本 +1、最強 Instructor、停泊同星系每座學院 +1、50／150／500 門檻與
  Warlord 平移皆有原始資料流支撐。
- 舊 Go `advanceCrewExperience` 沒有 500 上限，長局 CrewXP 會無限成長；本輪依
  `0x149FC..0x14A1A` 修正。
- 戰勝後擊沉敵艦的 XP 由另一條戰鬥寫回鏈產生；本輪證據只閉合每回合
  `Do_All_Ships_XP_Check_`，不把同一個 500 cap 未經追查套到戰後獎勵。
- AI 在原版同樣有逐艦 record；remake 自 2026-08-25 起已有持久實艦與 `CrewXP` 欄位，
  單一主力艦隊亦接上每回合、同星系太空學院與帝國 Instructor 累積。AI 多艦隊位置仍未建模。
