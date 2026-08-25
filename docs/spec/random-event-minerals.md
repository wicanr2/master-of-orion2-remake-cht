# 隨機事件 11／12 礦產規格

證據來源：[`random-event-minerals-audit-20260825.md`](../re/random-event-minerals-audit-20260825.md)。

## 目標與效果

| 事件 | 目標條件 | 效果 |
|---|---|---|
| 11 礦產枯竭 | 目標行星必須為 Ultra Rich | `4 → 3` |
| 12 礦產發現 | 目標行星礦產必須小於 4 | `min(4, old+2)` |

每次先從帝國殖民地均勻抽一座，再檢查條件，最多 200 次；耗盡即回報事件不適用。

## 狀態同步

- 同步更新 `Planet.MineralID`、顯示字串、`ColonyState.MineralRichness` 與
  `IndustryPerWorker`。
- 不改 `Planet.GravityID`／重力顯示。
- 玩家、熱座席位與 AI 使用同一套選擇及效果規則；事件不得寫到其他帝國。

## 驗收

- 枯竭不選 raw 0..3；發現不選 raw 4。
- 驗證 `4→3`、`0→2`、`3→4`。
- 驗證行星／殖民地產能同步且重力不變。
- 抽樣驗證玩家與 AI 的狀態隔離。
