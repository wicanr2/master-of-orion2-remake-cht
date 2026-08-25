# 隨機事件帝國目標逆向稽核（2026-08-25）

## 問題與舊行為

一般事件排程已依 `sub_2230A` 接線，但 remake 的已實作效果仍隱含作用於目前載入的真人帝國。
本次先回答原版如何在全銀河帝國間選目標，以及 `player+0xA6` 的真實來源；效果回寫是後續獨立
垂直切片，不能由純權重測試代替。

## 證據契約

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、IDAPython；位址皆為 IDA 線性位址（DOS/4GW LE object #1）
- 資料庫：`Orion2.exe.i64` 的一次性可寫副本；原始 EXE／資料庫唯讀
- 可重生腳本：`tools/ida/audit_event_schedule.py`
- 推論等級：下列控制流、欄位來源與公式均為已證實；事件效果如何映射到 remake 的 AI 薄模型仍未知

## 已證實

### 呼叫與輸入

- `sub_2230A @ 0x2230A` 在 `0x22472..0x2247E` 讀 `byte_180E84[eventID]` 的 good flag，
  以 `AL` 傳入 `sub_22D57 @ 0x22D57`，回傳帝國索引。
- 事件 9 與 24，以及 Lucky 已指定目標的好事件，會繞過這次一般目標抽選；這是事件分派例外，
  不是所有事件都固定打目前玩家。

### `player+0xA6` 是總人口

- `sub_E2710 @ 0xE2710` 掃描殖民地 record（stride `0x169`），對 owner 相符且狀態 `+6 == 0`
  的殖民地，把 byte `+0x0A` 累加到區域變數，最後在 `0xE29D4..0xE29D7` 寫入
  `player+0xA6`。
- `sub_10208A @ 0x10208A` 亦把 `player+0xA6` 當人口歷史序列來源。兩個獨立消費端共同排除
  「抽象國力分數」的舊解釋。

### 候選與權重

- 存活條件是 `player+0x24 == 0`。
- 壞事件另呼叫 `sub_230B6 @ 0x230B6`，掃描 36 筆事件 record；若該帝國已有活動中的壞事件，
  暫不進候選池。
- 壞事件移除候選中人口最低者，其餘權重為 `(population-minPopulation)^2`。
- 好事件移除候選中人口最高者，其餘權重為 `(maxPopulation-population)^2`。
- 權重交給 `sub_586D4 @ 0x586D4`。一般壞事件若最後選中 Lucky 帝國，`sub_22D57` 回傳 `-1`；
  它不在建立權重池時預先移除 Lucky。

### Lucky 的全帝國順序

- `sub_245C4 @ 0x245C4` 先由玩家槽 0 起掃到 `word_199998-1`，替每個存活且
  `player+0x8B9 == 1` 的帝國增加 `player+0xE73`。
- `sub_24511 @ 0x24511` 再按相同槽順序擲骰；`0x245B7..0x245BF` 證實一旦已有成功槽，
  就在進下一輪前離開。故一個銀河回合最多只有一個 Lucky 強制目標，不是每位熱座玩家各自
  跑一份一般事件排程。

## Remake 對映與保留邊界

- `gamedata.OriginalEventVictimWeights` 現明確接收 `populations`，並以測試鎖定兩種極值排除、
  差平方及 eligibility 行為；共用正規化抽樣仍由 `OriginalEventWeightedChoice` 表達。
- remake 執行期以真人席位順序後接剩餘 `AIPlayers` 作為可重播槽序；熱座轉換目前沒有保存原版
  raw player slot，因此這個順序是資料模型對映，不宣稱與任意匯入局的 raw slot 完全相同。
- AI 的殖民地、艦隊與建築模型比玩家側薄；沒有對應資料的事件必須消耗候選並回報不適用，不能
  靠只改播報文字冒充效果完成。
