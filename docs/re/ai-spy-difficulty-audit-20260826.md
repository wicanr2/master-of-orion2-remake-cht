# AI 間諜難度加成稽核（2026-08-26）

## 證據

- **已證實（官方表）**：`MANUAL_150.html` Generic AI bonuses 的 Spy Bonus 五級為
  `-2/-1/0/+1/+2`。
- **已證實（帝國攻防表）**：`Compute_Spy_Bonuses_ @ 0x100A83` 建立每個帝國的 attack／defense
  基底，科技、諜報 trait 與心靈感應共同進入兩表；Spy Master 只進 attack、Telepath 與政府只進
  defense。詳見 `sabotage-score-upstream-audit-20260825.md`。
- **2026-08-28 已證實並勘誤**：`Resolve_Player_Spies_ @ 0x1014A4` 是直接 difficulty
  consumer。只有攻方不是 human、守方是 human 時，最終攻守差值加入 `difficulty-2`；
  `Compute_Spy_Bonuses_ @ 0x100A83` 不讀難度。因此舊「AI attack／defense 共同注入」強推論
  已被原始指令推翻。完整直接 consumer 索引見
  [`difficulty-consumers-audit-20260828.md`](difficulty-consumers-audit-20260828.md)。

## Remake 映射與停止線

remake 現行 `aiSpyEmpireBonus` 把難度混入 AI 共同帝國基底，會同時影響攻守，與原版不符；
依 RE-first gate 本輪只登記差異，不修改 Go。AI 任務政策、訓練成本與 Agent 配置已由後續
`spy-turn-policy-audit-20260828.md` 閉合。
