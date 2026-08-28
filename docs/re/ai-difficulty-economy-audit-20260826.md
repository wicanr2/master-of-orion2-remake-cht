# AI 難度經濟加成證據稽核（2026-08-26）

## 問題

`internal/ai.AIDifficultyBonus` 已保存五級難度表，但沒有玩家路徑消費端；AI 每回合仍和玩家使用
相同食物、工業、研究、成長與人頭收入，故「資料表存在」不能算完成。

## 證據與等級

- **已證實（官方規格）**：`MANUAL_150.html` 的 `AI Opponents / Generic AI bonuses` 表明列
  Tutor..Impossible 的 Growth、Food、Prod、Res、BC、Command Deficit、Spy、Troops／Marines
  與 Antaran Marines 五列數值。既有逐格抄錄在 `docs/tech/original-ai-re.md` §2.1，typed 表在
  `internal/ai/original.go`；五列測試已存在。
- **已證實（remake 消費鏈）**：`GameSession.EndTurn` 對每個 AI 先呼叫 `ApplyAIEconomy`，再把
  殖民地交給 `RunEmpireTurnWithResearchRoller`。此前兩者之間沒有讀取 `AIDifficultyBonus`。
- **2026-08-28 已證實並縮小未知**：`Colony_Job_Production_ @ 0xDE280` 直接把
  `byte_DD4D7[difficulty]` 加入共用職務 subtotal；`Colony_Industry_Production_ @ 0xDEE1B`
  另在 derived 工業鏈加入 `(base*difficulty+4)/8`；`Colony_BC_Production_ @ 0xE03F1` 先把
  `byte_DD4E6[difficulty]` 加入 quarter factor，再以整座有機人口聚合後 `/4`。BC 的套用與捨入
  邊界已由指令證實；Food／Prod／Res 的 raw table 值與共用職務順序見完整 consumer 索引。

## Remake 映射

- 難度加成只寫入 `aiColoniesForTurn` 的暫態副本及 AI 的暫態 `PlayerState`，不得污染持久殖民地，
  也不得套到玩家或熱座真人。
- `GrowthPercent` 加入既有 `GrowthBonusSum`；Food／Prod／Res 以 quarter 欄位進 `ColonyState`；
  BC 以 quarter-per-pop 欄位進 `PlayerState`。
- 負 quarter 使用數學向下取整，避免 Go 對負數朝零截斷把 Tutor 的小額懲罰吃掉。

## 仍未知

- Food／Prod／Res 三職務如何由同一 raw quarter table 對應官方欄名，仍須和各職務 consumer
  的完整 subtotal 對照；BC 的整座聚合捨入已不再未知。
- Command Deficit 與 Spy 已由 `ai-command-deficit-audit-20260826.md`、
  `ai-spy-difficulty-audit-20260826.md` 的獨立切片追查；2026-08-28 已證實原版 Spy 只在
  AI 攻擊真人時對 resolution 差值加入 `difficulty-2`，remake 的共同攻防注入須待 RE gate 後修正。
  Troops／Marines 經現況稽核確認已由
  `GroundDifficultyBonus(difficulty, GroundAIEmpire)` 接到 AI 殖民地防守與叛軍，五級公式
  `difficulty-2` 正是 `-2..+2`；Antaran Marines 雖有 `GroundAntaranSide` 的 `2*difficulty-4`
  typed helper，仍缺非測試 runtime 呼叫端。本切片不以經濟接線冒稱整張難度表全部閉合。
