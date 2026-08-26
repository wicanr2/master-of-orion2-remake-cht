# AI 難度經濟加成規格

> 狀態：2026-08-26 接線；數值為官方表已證實，quarter 捨入與操作順序為強推論。

## 輸入與作用域

- 輸入為 `GameSession.Difficulty` 的 0..4 與 `internal/ai.AIDifficultyBonus`。
- 只作用於 `AIPlayers`；玩家與熱座真人不得取得加成。
- 不修改持久 `AIOpponent.Colonies` 的基礎產出率，避免每回合累加。

## 規則

1. 每個 AI 殖民地暫態加入 `GrowthPercent`。
2. 食物、工業、研究分別加入 `farmers/workers/scientists × 對應 quarters ÷ 4`。
3. BC 加入 `population × BCQuarters ÷ 4`，再進該殖民地既有收入建築倍率。
4. quarter 除法採數學向下取整；例如 Tutor 一名農夫的 `-1/4` 必須成為 `-1`，不能被截成 0。
5. 非法難度安全回退為零加成。

## 驗收

- engine 純規則測試涵蓋正值、負 quarter、BC 與玩家零值不變。
- shell 測試證明 AI 暫態副本取得難度值、原殖民地不被污染，且玩家路徑不讀取該表。
- Docker 目標測試通過後，更新 `WORKLIST.md` 與 parity matrix；不得把此切片描述成完整原版 AI。

