# AI 對真人正式戰爭 policy 消費規格

規格狀態：`CONFORMED`（AI 艦隊接戰 writer 與正式 policy 下游）

證據：[`../re/ai-human-diplomacy-dispatch-audit-20260828.md`](../re/ai-human-diplomacy-dispatch-audit-20260828.md)

1. `Treaty.FormalPolicy` 為 1..6 時，目標估值必須直接使用同 raw 編碼。
2. policy 4／5／6 分別映射 `DiploLimitedWar`／`DiploWar`／`DiploTotalWar`；不得依關係分數
   將 5 降成 4，或將 6 降成 5。
3. 只有正式 policy 為 0 的舊存檔相容路徑可暫由 `StanceName` 投影。
4. 原版完整派艦／選擇真人目標 producer 完成後，應移除上述相容 fallback。

## AI 艦隊接戰 writer

1. AI 艦隊抵達玩家殖民星、且正式 policy 小於 4 時，在 raid 解算前宣戰。
2. 難度 0–2 寫 raw policy 5；難度 3–4 寫 raw policy 6。
3. 清除正式和平類條約、貿易、研究、雙向納貢與特殊交換。
4. 關係 raw 寫 `-75-roll`，其中 `roll` 為原版 `Random(25)` 的 0..24，並同步 normalized 顯示值。
5. `StanceName` 只同步為戰爭顯示；正式 policy 才是規則真值。
6. 建立 `WantsAudience=true`、`AudienceReason=war`，表示原版訊息 dispatcher 的請求 bit。
7. writer 必須冪等：policy 已至少 4 時不得再次清協議、重擲關係或重複改狀態。
8. raw policy 6 必須通過 `Diplomacy_Growth_` 的條約 pass，並在關係漂移 pass 套用
   `policy>=4` 的 -90 上限；不得因舊 enum 上限 5 而 fail-closed 跳過整回合。

## 驗收結果

- 難度 2／3 分別寫 raw 5／6，清協議並建立宣戰會談。
- 正常 AI 艦隊 ETA 歸零、抵達玩家殖民星時建立正式戰爭。
- writer 冪等測試確認既有戰爭不消耗亂數。
- raw 6 關係成長、-90 drift、JSON snapshot 往返及完整純 Go 回歸通過。

原版完整派艦／選擇真人目標策略不在此 CONFORMED 範圍；現行出兵意願仍是明示 fallback。
