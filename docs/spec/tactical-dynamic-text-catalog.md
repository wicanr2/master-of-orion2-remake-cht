# 格子戰術動態文案外部化規格

## 證據界線

原版三態武器槽與待命狀態恢復已有 `docs/re/tactical-weapon-status-audit-20260824.md`
的 IDA／手冊證據；動態戰報的證據邊界見
`docs/re/tactical-dynamic-battle-log-audit-20260827.md`。戰報逐字內容、右鍵資訊彈窗外觀與
循環方向仍為未知；remake 的單行戰報屬操作轉接，不宣稱原版逐字一致。

## 契約

- Go 僅保存 `tactical.*` 語意鍵、狀態列舉與動態參數；玩家可見固定句子及格式模板放在
  `assets/i18n/ui.json`。
- `tacticalText` 是格子戰術固定文案入口；格式化模板的中英文參數數量與順序必須相同。
- 艦名、武器名稱、傷害、彈藥與射界是具型別執行期資料，不複製進文案目錄。
- 所有訊息仍經既有 `drawTacticalMessage` 安全框量測及截斷，不新增繞過安全框的直接繪字。
- 戰報鍵依用途分為行動佇列、移動、目標拒絕、開火結果、回合摘要與勝負；相同語意必須
  共用同一鍵，不可在不同分支複製近似句子。
- 戰後轉場名稱亦屬玩家可見固定文案，使用 `tactical.transition.result`，不得留在 `goTo` 參數。

## 分段收斂

本切片涵蓋畫面標題、開場提示、武器槽狀態／明細，以及移動、開火、回合與勝敗戰報。
驗收時必須確認 `newTacticalScreenForShips` 到 `drawTacticalMessage` 的執行切片不再含 `.tr(`；
不得為了外部化而改動戰鬥規則。
