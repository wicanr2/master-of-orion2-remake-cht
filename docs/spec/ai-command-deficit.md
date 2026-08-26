# AI 指揮赤字難度成本規格

> 狀態：2026-08-26；數值與乘法消費端均有靜態／官方證據。

1. `PlayerState.CommandOverflowCostPerPoint <= 0` 使用玩家預設 10 BC，維持舊存檔與所有玩家行為。
2. AI 回合依 `AIDifficultyBonus.CommandDeficitBC` 寫入 12／11／10／9／8。
3. 成本為 `max(0, UsedCommandPoints-CommandPointsSupply) × perPoint`。
4. 暫態覆寫不序列化；每回合由 `GameSession.Difficulty` 重建。
5. 測試至少覆蓋玩家預設、Hard AI 覆寫、無缺口及 shell AI 回合接線。

