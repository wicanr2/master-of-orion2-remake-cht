# 歷史圖與最終分數稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、IDAPython；位址均為 IDA 線性位址（DOS/4GW LE object #1）。
- 匯出器：`tools/ida/audit_history_score.py`；只在暫存 `.i64` 副本運行，保留原始名稱、位址、bytes 與 operand。
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
- **仍待本項下一切片**：把 INFO History 從三項未正規化 400 筆改為上述四項 350-byte ring、保存四個 divisor，並為舊 JSON 做明確遷移。
- **資料模型限制**：原版逐玩家 `+0x1F2[target]` 的殲滅歸屬尚未保存；現行單人 fallback 仍會把 AI 互滅算給玩家，不能宣稱該輸入已對齊。
- remake 尚無 Evolutionary Mutation 再選能力 UI，因此取得科技後四點視為全未使用；未來加入 mutation UI 時，必須改存實際剩餘點數。
