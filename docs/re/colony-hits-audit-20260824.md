# `Get_Colony_Hits_` 靜態稽核（2026-08-24）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫副本：`moo2-Orion2-consumer-work.i64`，SHA-256
  `6562313be340a6bb80d43f25478446ba0bae24285ac86f0419b4f7de02a14fd0`。
- 工具：IDA Pro 9.4／IDAPython，SDK 940；本文位址均為 IDA linear address。
- 探針：`tools/ida/audit_colony_hits.py`。探針只匯出原始函式名、位址、bytes、
  operand 與交叉參照，不改名、不套型別、不寫回資料庫。

## 已證實：回傳公式

`sub_42371 @ 0x42371` 的唯一直接 caller 是 `sub_416CF @ 0x416CF` 的
`0x41A54`；回傳 `AX` 立即寫入快速戰鬥 record `+0x1D`。函式以
`word_199878` 選取殖民地，record stride 為 `0x169`：

```text
hits = u8[colony+0x0A]
     + u16[colony+0x130]
     + u16[colony+0x132]

for rawBuildingID = 0..48:
    if colony[0x136+rawBuildingID] != 0
       and rawBuildingID not in {8, 40, 41}:
        hits += 40
```

欄位可由已驗證的 `.GAM` `Colony` layout 對回：`+0x0A` 是人口、`+0x130`
是士兵、`+0x132` 是戰車，`+0x136..+0x166` 是 49 個建築旗標。原版建築列舉的
`8/40/41` 分別是 Battlestation、Star Base、Star Fortress；它們在快速戰鬥中另有
自己的戰鬥者，因此不重複計入殖民地本體耐久。

證據等級：上述控制流、位址、常數、欄位與三個排除 ID 均為**已證實**。

## 已證實：與戰略轟炸的關係

`sub_416CF` 於建立殖民地戰鬥者時把本函式回傳值寫入 record `+0x1D`；同一條
快速戰鬥鏈以 record `+0x1F` 累積結構傷害。這證明它是殖民地本體的快速戰鬥耐久，
不是「一次轟炸後要隨機拆哪些建築」的分配表。

## 相關證據邊界

`sub_E87D2` 的戰略傷亡 caller／consumer、候選順序與回寫已另由
`docs/re/strategic-colony-casualties-audit-20260824.md` 追明；兩份證據分工如下：

- 本文只定義快速戰鬥殖民地本體的總耐久。
- 傷亡稽核只定義戰略轟炸結果池如何隨機回寫各殖民地欄位。
- 16-bit 極端溢位仍未驗證；合法遊戲資料不接近該邊界，remake 不模擬溢位。
