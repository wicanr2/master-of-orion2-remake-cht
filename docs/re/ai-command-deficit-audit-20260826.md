# AI 指揮赤字難度成本稽核（2026-08-26）

## 證據

- **已證實**：`MANUAL_150.html` 的 Generic AI bonuses 表將 Command Deficit BC 列為
  Tutor／Easy／Average／Hard／Impossible = `12/11/10/9/8`。typed 真值已在
  `internal/ai.AIDifficultyBonus`，既有測試抽查 Tutor、Average、Impossible。
- **已證實**：`docs/re/player-maintenance-audit-20260825.md` 已定位原版 `0xE2517..0xE2584`：
  比較 used／supply 後，將未覆蓋指揮點乘難度成本。remake 現行 `RunEmpireTurn` 同樣先算
  `UsedCommandPoints-CommandPointsSupply`，但固定呼叫玩家每點 10 BC 的函式。
- **結論**：玩家維持手冊 p.169 的 10 BC；AI 必須依五級表覆寫每點成本，不能把 Average 的 10
  套給所有難度。

## 邊界

本切片只閉合「已算出的指揮點缺口 × 每點成本」。原版 AI 如何決定造多少艦、建哪級軌道站、
是否容忍赤字仍屬 AI state machine，不因成本接線而完成。

