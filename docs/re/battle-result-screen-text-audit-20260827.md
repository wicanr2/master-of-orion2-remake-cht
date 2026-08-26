# 艦隊戰果摘要文案稽核（2026-08-27）

## 現況與證據

- **已證實（remake source）**：`battleResult()` 使用 `TURNSUM.LBX#0` 作通用面板，顯示 `BattleResult` 的敵方、勝敗、開戰艦數、逐回合損失與總損失；它不是已由 IDA 證實的原版專用戰果畫面。
- **已證實（本輪修改前的 source 狀態）**：快速結算 `ResolveBattle` 與安塔蘭終局 `AssaultAntares` 都曾在每回合解算後寫入 `BattleResult.Log []string`，字串由規則層直接組成中文；本輪已依下方結論改為 typed 記錄。
- **已證實（格子戰術）**：`ApplyTacticalCombatOutcome` 寫入同一 `BattleResult`，但不建立逐回合 log；戰果畫面仍可由總數欄呈現。
- **強推論／介面 adapter**：現行面板、標題與摘要句是 remake 顯示設計，不是原版逐字／逐像素 oracle。

## 本輪邊界

將逐回合 log 改成 typed `BattleRoundResult`，固定標題、結果及格式模板移至 JSON。敵方名稱仍是對局動態資料；安塔蘭固定敵方顯示改由 typed enemy kind 讓 UI 查 catalog。本輪不改快速結算、格子戰術、艦員經驗或終局勝敗公式。
