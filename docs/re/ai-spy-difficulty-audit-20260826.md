# AI 間諜難度加成稽核（2026-08-26）

## 證據

- **已證實（官方表）**：`MANUAL_150.html` Generic AI bonuses 的 Spy Bonus 五級為
  `-2/-1/0/+1/+2`。
- **已證實（帝國攻防表）**：`Compute_Spy_Bonuses_ @ 0x100A83` 建立每個帝國的 attack／defense
  基底，科技、諜報 trait 與心靈感應共同進入兩表；Spy Master 只進 attack、Telepath 與政府只進
  defense。詳見 `sabotage-score-upstream-audit-20260825.md`。
- **強推論（難度注入位置）**：官方欄名只有單一 Spy Bonus，未拆 Attack／Defense；它是 AI
  帝國級難度值，故 remake 在 AI 的共同種族／帝國基底加入一次，讓 AI 作攻方與守方都消費。
  現有 IDA 匯出尚未顯示 difficulty global 在 `sub_100A83` 內直接讀取，可能由 caller／runtime
  profile 先寫入欄位；因此「同時作用攻守」不升格為指令級已證實。

## Remake 映射與停止線

`aiSpyEmpireBonus` 回傳種族、心靈感應及難度的共同值；玩家路徑不讀難度表。AI 任務政策、訓練
成本與 Agent 配置仍是獨立未閉合項，不因本切片完成。

