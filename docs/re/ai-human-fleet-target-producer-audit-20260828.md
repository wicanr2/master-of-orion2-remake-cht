# AI 對真人艦隊目標 producer 稽核

日期：2026-08-28

## 證據契約

- `Orion2.exe` SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- `.i64` SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- IDA Pro 9.4；DOS/4GW LE image、IDA linear EA；只分析正式資料庫的可寫副本。
- raw instructions／bytes／xrefs：
  [`evidence/ai-human-fleet-target-ida-20260828.json`](evidence/ai-human-fleet-target-ida-20260828.json)。
- Hex-Rays 輸出只供導覽；變數名與型別不是證據。

## 已證實

1. `sub_53EDB @ 0x53EDB..0x544A1` 是正常每回合 AI 對真人決策 producer；由主回合鏈
   `sub_136B3 @ 0x136B3..0x13822` 的 `0x13706` 呼叫 `sub_252A7`，再於 `0x252A7`
   呼叫 `sub_53EDB`。它不是 `sub_DB257` 的軍事接戰階段。
2. producer 排除 `player+0x28==100` 的 source，只處理未淘汰 AI；僅在 `player+0x7C7==-1`
   且沒有任何 human target 的 formal policy `>=4` 時掃描真人候選。
3. 真人候選必須仍有效、未進 formal war、具備接觸／可互動狀態，且不能處於三種已占用的
   `player+0x74F` 狀態。精確欄位語意仍有部分未知，故不以推測名稱取代 raw offset。
4. `sub_544A1 @ 0x544A1..0x54CC0` 以 directional relation `player+0x617`、personality class、
   外交事件記憶、接觸狀態、雙方排名／科技趨勢、難度及多次 `Random_` 決定四類結果。
   舊 `aiRaidGraceTurns=12`、`aiRaidStrengthMargin=125` 與
   `PersonalityLosingGroundChance` 擬亂數不在此原版鏈中。
5. 類型 2 成功時，`sub_53EDB` 只有在 personality class 不等於 4、候選 target 存在、
   `player+0x857` 估值高於門檻且 `player+0x847>0` 時，才把 target 寫入
   `player+0x7C7`、任務碼寫入 `player+0x7C9`，並把 `player+0x816` 寫成
   `Random_(20)+20`，即 20–39 回合。
6. `sub_DB47E @ 0xDB47E..0xDB659` 是之後的 AI 軍事階段；它依序準備目標表、呼叫
   `sub_D7764`、再呼叫 `sub_DB257`。`sub_DB257` 讀 `+0x7C7/+0x7CA`，在目標矩陣 entry
   非零時呼叫 `sub_51078`，屬接戰／宣戰 consumer，不是目標 producer。

## Remake 對映與限制

- 已接：新增可持久化 `OriginalHumanTargetDecisionCooldown` 對映 `player+0x816`；大於 0
  時每世界回合遞減且不啟動新真人目標，成功派艦後依原版範圍寫 20–39。
- 已移除：producer 的固定 12 回合寬限、1.25 倍軍力門檻與 losing-ground personality
  擬亂數。10 回合 `LastRaidTurn` 只留作 remake 單一主力艦隊停在同星時避免每回合重複
  結算，不能稱作原版 target cooldown。
- 尚未閉合：`sub_544A1` 所需的完整 directional incident memory、排名／科技趨勢欄位與四類
  結果。現行願戰來源仍是明示的 `DecideStance` 相容 fallback；本輪不把它升格為原版決策。

