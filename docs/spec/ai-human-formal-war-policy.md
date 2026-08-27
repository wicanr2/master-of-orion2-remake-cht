# AI 對真人正式戰爭 policy 消費規格

規格狀態：`CONFORMED`（正式 policy 下游）；宣戰 producer 仍為 `DRAFT`

證據：[`../re/ai-human-diplomacy-dispatch-audit-20260828.md`](../re/ai-human-diplomacy-dispatch-audit-20260828.md)

1. `Treaty.FormalPolicy` 為 1..6 時，目標估值必須直接使用同 raw 編碼。
2. policy 4／5／6 分別映射 `DiploLimitedWar`／`DiploWar`／`DiploTotalWar`；不得依關係分數
   將 5 降成 4，或將 6 降成 5。
3. 只有正式 policy 為 0 的舊存檔相容路徑可暫由 `StanceName` 投影。
4. AI↔真人正式宣戰 producer 完成後，應移除上述相容 fallback。
