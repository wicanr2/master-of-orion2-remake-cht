# 歷史圖與最終分數稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、IDAPython；位址均為 IDA 線性位址（DOS/4GW LE object #1）。
- 匯出器：`tools/ida/audit_history_score.py`；只在暫存 `.i64` 副本運行，保留原始名稱、位址、bytes 與 operand。
- 2026-08-28 追加正式證據：
  [`evidence/history-score-ida-20260828.json`](evidence/history-score-ida-20260828.json)，含
  `player+0x224`、`player+0x1F2`、`player+0x204` 全庫 operand 掃描、83 筆 topic record
  與 SHA-256 可回查 bytes。
- 以下原始算式與欄位排列為**已證實**；Go 無一對一 raw 快取時的 typed 對映另標示。

## Record_History_ 的四項資料

`Record_History_ @ 0x10208A` 由 `Next_Turn_Calc_ @ 0x136B3` 在 `0x137FD` 每回合直接呼叫。
每位玩家記錄四個 350-byte 環形陣列，索引在 `word_17D636`，到 `350 (0x15E)` 後回零：

| player record | 指標 | 本回合 raw 來源 |
|---|---|---|
| `+0x8DF[index]` | Fleet | 每艘有效艦依 hull class 加 `10×2^(class+1)` |
| `+0xA3D[index]` | Technology | `player+0x224` 的累積科技值 |
| `+0xB9B[index]` | Population | `player+0xA6` 的帝國人口彙總 |
| `+0xCF9[index]` | Buildings | `sub_E2671 @ 0xE2671` 加總所有己方非被佔領殖民地之已建建築 raw cost |

四項各有至少為 1 的全局 divisor。若任何玩家的 `raw/divisor > 250`，該 divisor 逐一增加，
直到所有本回合值不超過 250；divisor 改變時，所有玩家過去 350 格都依
`oldValue×oldDivisor/newDivisor` 重縮放，最後才寫入目前 ring index。故現行 remake 的 400 筆
未正規化人口／BC／艦隊模型並非原版資料形狀；BC 與「當回合 Research」也不是原版圖表指標。

## Technology `player+0x224` producer 閉合（2026-08-28）

全庫直接 operand 掃描只找到三個函式：

| raw 函式 | 位址 | `+0x224` 用途 |
|---|---:|---|
| `sub_E4535` | `0xE4535..0xE45FD` | 唯一直接清零／累加 writer |
| `sub_10208A` | `0x10208A..0x1022CB` | `Record_History_` 讀取端 |
| `sub_21B6D` | `0x21B6D..0x2230A` | 事件訊息 raw case 2 的科技排名讀取端 |

`sub_E4535` 每次先把 `player+0x224` 清零，再計算：

1. topic `74..0`：讀 `player+0xC4+topic == 3` 作已完成旗標，讀 23-byte topic record
   的四個 `technology_value_slots` 作額外 known-state 分子／分母，再乘該 topic 的
   `baseCost` 並作帶號整數除法。
2. 本版 83 筆 record 的四個 `technology_value_slots` **全部為零**，所以一般 topic 的實際
   1.31 公式精確簡化為：完成 topic 就加完整 `baseCost`，未完成加零；不是 application
   比例，也不是本回合 RP。
3. topic `75..82`：若 Hyper raw level 非零，加入
   `baseCost + 10000×(level−1)`。本版八筆 `baseCost` 均為 15000，因此各領域的逐級值為
   15000、25000、35000……。

83×23 bytes 位於 `0x17D904`，合併 SHA-256 為
`5a24f959d09333c177cef5d2459111244c1e2cadb01b2da3d9885c535d3b2877`。匯出同列保存
topic index、raw bytes、兩個研究選項導覽值、四個實際 technology-value slot 與 base cost；
研究選項欄只供表格導覽，不與全零 value slots 混稱。

時序亦已證實：`Next_Turn_Calc_ @ 0x136B3` 在 `0x13742` 先呼叫 `sub_E4F49`；後者對每位
存活帝國依序做產出重算、突破檢查，並於 `0xE4FA8` 呼叫 `sub_E4535`。主鏈直到
`0x137FD` 才呼叫 `Record_History_`，所以當回合歷史使用研究處理後重算出的科技值。

這一結論把 remake 先前「以完成主題成本重建 `+0x224`」由強推論升格為本版已證實；現行
Hyper `15000+10000×levelFromZero` 也與 raw 公式一致。單元測試仍只證 remake 自洽，升格依據
是上述唯一 writer、兩個 consumer、topic bytes 與主鏈時序。

## 殲滅歸屬 `player+0x1F2[target]` producer 閉合（2026-08-28）

全庫直接 operand 掃描證實 `+0x1F2` 只有四個函式使用：`sub_9D816`、`sub_9DEF7` 與
`sub_9E84C` 讀取，`sub_E45FF @ 0xE45FF..0xE481F` 是唯一 writer。`sub_E45FF` 先由傳入的
3753-byte player record 算出敗方索引；若該敗方 `+0x204` 小於 8，便在 `0xE47D9..0xE4813`
把當前星曆 `dword_192FD8` 的低 word 寫入：

```text
player[player[defeated]+0x204] + 0x1F2 + 2*defeated = currentStardate
```

`+0x204` 的全庫直接掃描同樣只找到四個函式：

| raw 函式 | 位址 | 已證實資料流 |
|---|---:|---|
| `sub_E4EB3` | `0xE4EB3..0xE4F49` | 無殖民地帝國的殲滅檢查完成後，把每個存活 player 的 `+0x204` 重設為 `-1` |
| `sub_E9D62` | `0xE9D62..0xEA8C4` | 非互動回合戰鬥鏈；若 `sub_E7DCA/sub_E8029` 選出殖民地記錄，於 `0xEA0F8` 把本次 battle side 寫到該殖民地舊 owner 的 `+0x204` |
| `sub_EAAA1` | `0xEAAA1..0xEB192` | 互動戰鬥狀態鏈；同一資料流於 `0xEAFE3` 寫入舊 owner 的 `+0x204` |
| `sub_E45FF` | `0xE45FF..0xE481F` | 讀取上述暫態 battle side，將星曆寫入該 side 的逐敗方 `+0x1F2` |

兩個 producer 都先以殖民地索引乘 361，從殖民地 record 首 byte 讀舊 owner，再乘 player stride
3753 定位敗方；寫入值則是同次戰鬥選出的 side/player 索引。`Next_Turn_Calc_` 在 `0x13756`
呼叫 `sub_E9D62`，後者末尾呼叫 `sub_E4EB3`；互動版由 `sub_F69FE` 在 `0xF6E05` 呼叫
`sub_EAAA1`。因此欄位是只活到殲滅檢查的戰鬥歸屬暫態，不是累積擊殺數。

上述位址、writer 唯一性、星曆值與逐敗方陣列形狀為**已證實**；把 battle side 稱為玩家語意上的
「殲滅者」是由最終分數 consumer 與兩條戰鬥 producer 共同支持的**強推論**。原版 RE 輸入鏈已
閉合；remake 是否保存此 8×8 歸屬矩陣是後續實作議題，不得以目前單人 fallback 反證原版規則。

## 最終分數 orchestrator

`sub_9D977 @ 0x9D977` 依序寫八項分數並加總到 score record `+0xAA`：

- 殲滅：`sub_9E84C`，逐玩家讀 `player+0x1F2[target] > 0`，每族 `+50`。
- 科技：`sub_9E90B/sub_9E973`，83 個 known topic 每個 `3`，八個 Hyper level 每個 `5`。
- 安塔蘭：`sub_9DB21`／raw getter `0x9E711`，`player+0x1F0` 非零得 `250`。
- 獵戶座：`sub_9DB39`／raw getter `0x9E8A3`，`player+0x1EF` 非零得 `100`。
- 議會：`sub_9EC32`／raw getter `0x9EA17`，`player+0x1F1` 非零得 `100`。
- 俘虜人口：`player+0x202×2/(galaxySize+1)`。
- 人口：`sub_9E9DA` 掃所有殖民地 owner 與人口。
- 時間：`playerCount×(20×(galaxySize+1)+80)−(stardate−35000)`；人口為零時回零。

此前八項係數大致正確，但漏了 orchestrator 的最後倍率：`sub_58F4A @ 0x58F4A` 先轉換
`player+0x89F` 種族能力並取得未使用 Picks；Evolutionary Mutation known-state 另加 4 Picks。
倍率為 `100+10×positiveUnusedPicks`，最後總分為：

```text
final = (rawEightPartTotal * multiplierPercent + 50) / 100
```

`0x9DAF8..0x9DB14` 沒有負分夾零。舊 Go 的 clamp 是無證據自訂，已移除。

## Remake 對映與剩餘限制

- **已接**：八項總和、未使用初始 Picks、Evolutionary Mutation 尚未消費的 4 Picks、百分比與四捨五入；倍率進 JSON／熱座狀態。
- **已接且已由本輪 RE 升格**：INFO History 四項 350-byte ring、四個 divisor、舊 JSON 遷移，
  以及一般 topic／Hyper 的 `player+0x224` 等價重建。
- **RE 已閉合、remake 尚未接**：原版逐玩家 `+0x1F2[target]` 保存殲滅星曆；現行單人 fallback
  尚未保存 8×8 歸屬，仍可能把 AI 互滅算給玩家，不能宣稱該輸入已對齊。
- remake 尚無 Evolutionary Mutation 再選能力 UI，因此取得科技後四點視為全未使用；未來加入 mutation UI 時，必須改存實際剩餘點數。
