# Hyper-Advanced 重複等級成本稽核

日期：2026-08-24

## 問題

remake 已能重複研究八個 Hyper-Advanced topic，但每一級都沿用第一級 profile 成本，文件將
後續公式標為未知。本次追查原版成本 consumer 與 level promotion 寫入端。

## 輸入與工具

- `Orion2.exe` SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- `.i64` SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- IDA Pro 9.4，`metapc`，IDA linear／DOS4GW LE image
- 非破壞性探針：`tools/ida/audit_hyper_research_cost.py`

## 已證實

### 成本公式

修正外部符號索引的 `Player_Research_Cost_` 對應 raw `sub_E1E96 @ 0xE1E96..0xE1EC6`：

1. `dx <= 0` 時回 0。
2. 取 `topic*0x17` 後讀 `dword_17D916[topic*0x17]` 作 topic 基礎成本。
3. `topic < 75` 時直接回基礎成本。
4. `topic >= 75` 時讀 `byte [player + topic + 0x1D1]`，乘 `0x2710 = 10000` 後加到基礎成本。

因 `75 + 0x1D1 = 0x21C`，八個 terminal topic 75..82 正好對應玩家結構
`+0x21C..+0x223` 的八個 Hyper level bytes。故公式為：

```text
next_cost(topic) = version_base_cost(topic) + completed_level_byte(topic) * 10000
```

patch profile 已證實第一級基礎成本為 1.31 的 15000、1.50 的 25000，因此：

- 1.31：15000、25000、35000……
- 1.50：25000、35000、45000……

### level 寫入與保存

- raw `sub_10F919 @ 0x10F919..0x10F93B`，修正索引名
  `Promote_A_Hyper_Value_By_Field_`：若 `topic >=75`，直接 `inc byte [player + topic + 0x1D1]`。
- raw `sub_10F884 @ 0x10F884`／`sub_10F8B7 @ 0x10F8B7` 分別把玩家
  `+0x21C..+0x223` 八 bytes 保存到暫存表並恢復。
- raw `sub_10F8ED @ 0x10F8ED` 逐一遞增全部八個 bytes；它是獨立 promotion helper，不改變
  `Player_Research_Cost_` 對單一目前 topic 的公式。
- `sub_6D048 @ 0x6D048` 已在前一輪證實以相同 bytes 增加微型化等級，成本與微型化因此共用
  同一份 completed level 狀態。

## 勘誤

「後續 Hyper 成本公式未知、暫沿用第一級」已被上述直接讀寫鏈推翻。remake 內每級固定成本
不是近似可接受的 oracle 留白，而是可修正的玩法錯誤。

## Remake 對應

- `HyperAdvancedLevels[topic]` 保存 completed level byte 的語意。
- 研究結算、研究選單與 AI candidate 都必須使用 `base + levels*10000`。
- 舊 JSON 若只有 `CompletedTopics[hyper]=true` 而沒有 map，遷移為 level 1；下一級成本因此是
  base+10000，不得再以第一級成本完成。
