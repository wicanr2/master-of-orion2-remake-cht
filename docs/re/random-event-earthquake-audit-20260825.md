# 隨機事件 7「地震」靜態稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 資料庫：`Orion2.exe.i64` 的一次性可寫副本；原始 EXE 與正式資料庫均唯讀。
- 工具：IDA Pro 9.4／IDAPython；位址均為 IDA linear、DOS/4GW LE object #1。
- 探針：`tools/ida/audit_event_earthquake.py`。它只匯出原始函式名、位址、bytes、
  operand、caller 與 callee，不改名、不套型別、不寫回正式資料庫。

## 事件建立與目標

**已證實**：`sub_2230A @ 0x2230A` 的事件 7 建立分支位於
`0x22A33..0x22A61`。它把目標 player slot 傳給 `sub_23DA0 @ 0x23DA0`，只把回傳的
colony index 寫入事件 record `word +3`；選不到殖民地時使事件失效。建立時沒有決定固定
人口傷亡或建築數。

`sub_23DA0` 由 colony 0 起依序掃描，以 reservoir sampling 在符合下列條件的殖民地中
等機率選一筆：

- `colony +0x00 == target player slot`；
- `colony +0x13F == 0`；
- `colony +0x06 == 0`。

所有直接 callsite 都把 player slot 放在 `eax`，而 helper 每遇到第 `n` 個合格候選便呼叫
`Random(n)`，只有結果等於 1 才取代目前目標。擁有者與 reservoir 行為為**已證實**；
`+0x06` 已由既有 `.GAM` layout 對回 outpost／無殖民地狀態，`+0x13F` 已證實是
Capitol 建築槽；remake 已以 `ColonyBuildings` 的 typed raw 9 狀態接入相同排除條件。

## 地震強度公式

**已證實**：事件消費端 `sub_206A2 @ 0x206A2` 的 case 7 位於
`0x20A4C..0x20AD2`。它先把 record 狀態設成結束，再以目標 colony index、record `+5`
與 `+7` 的指標呼叫 `sub_238A8 @ 0x238A8`，最後呼叫 `sub_E2A70` 重算殖民地。

`sub_238A8` 的原始指令給出：

1. `0x238D3..0x23918` 計數 49 個 `colony +0x136 + rawBuildingID` 非零旗標，並加上
   `colony +0x0A` 的人口單位數，得到 `P`。
2. `0x2391A..0x23930` 依序呼叫 `Random(3)`、`Random(2)`，得到 1..3 與 1..2。
3. `0x23932..0x2394C` 計算
   `damage = max(1, floor(P × (Random(3)+Random(2)) / 10))`。
4. `0x2394C..0x2395D` 建立一次戰略殖民地傷亡 record，將上述 damage 傳給
   `sub_DD2F2 @ 0xDD2F2`。

因此舊 Go 的「固定隨機扣 1～2 人口，再隨機拆一棟」不是原版公式。

## 傷亡與回寫

**已證實**：`sub_238A8` 不是另寫一套地震專用人口／建築規則，而是呼叫既有
`sub_DD2F2`；此 record 的 `word +6` 為零，故落入 `sub_DCEBD @ 0xDCEBD` 的戰略殖民地
傷亡候選池。該 helper 已於
[`strategic-colony-casualties-audit-20260824.md`](strategic-colony-casualties-audit-20260824.md)
獨立閉合：一般建築、陸戰隊、戰車、建造進度與人口共同參與等機率候選；成本不足即停止，
最後一名殖民者另走 100 點尾端規則。

`sub_238A8 @ 0x23962..0x23981` 將傷亡結果中的人口損失寫到事件 record `word +5`，並
計數 49 個建築摧毀旗標寫到 `word +7`。故 GNN 可分別報導死亡人口與摧毀建築數；它不是
只保存單一建築名稱。

## Remake 對映與限制

- 重用 `gamedata.ResolveStrategicColonyDamage`，避免地震與戰略轟炸對同一原版 helper
  產生兩套互相漂移的規則。
- 玩家、熱座與 AI 均需使用同一強度公式、同一事件亂數流、相同建築／駐軍／建造進度回寫。
- 若人口歸零，必須移除殖民地及其平行陣列，不能保留人口 0 的幽靈殖民地。
- remake 的建築 map 只能計數可對回 raw ID 的項目；未知匯入名稱不假裝是原版旗標。
- remake 不保存原版 packed colonist 排列；人口被抽中後以既有 deterministic 職務／群組
  正規化維持資料一致。候選集合與總人口回寫對齊，死亡的是哪一筆 packed colonist 仍是
  **明示近似**。
- `colony +0x13F` 已解為 Capitol 建築槽；地震目標抽選器已只在無 Capitol 候選間作 reservoir sampling。
