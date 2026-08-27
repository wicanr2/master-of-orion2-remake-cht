# 殖民地起始職務規格

狀態：CONFORMED（2026-08-27）

## 範圍

依 `docs/re/colonization-starting-job-audit-20260827.md`，修正一般玩家與共用殖民地 builder
的第一位殖民者職務。前哨站內部 raw cache、原版 PRNG 與畫面動畫不在本切片。

## 契約

- `naturalFood` 為未套種族／建築加成的行星自然食物產出。
- `naturalFood <= 0`、Lithovore 或 Cybernetic 任一成立時，殖民船帶來的 owner 人口為工人；
  否則為農夫。
- 原住民特殊物產的三位 Native 人口永遠為農夫，不跟隨 owner 的起始職務。
- `ColonyState.Farmers/Workers/Scientists` 與 `PopulationGroups` 必須同時反映同一分配，
  `Population` 總數不變。
- 玩家與 AI 都經 `newColonyFromPlanet`，不得在兩個呼叫端複製不同規則。

## 驗收

- 純規則測試覆蓋自然食物、零食物、Lithovore、Cybernetic。
- shell 測試覆蓋一般可耕行星、不可自然耕作行星，以及帶原住民時 owner 工人／Native 農夫分流。
- 既有殖民成功、殖民船消耗、平行陣列與回合經濟測試維持通過。

## 實作

- `internal/gamedata/colonization.go` 保存純規則。
- `internal/shell/colonization.go` 同步更新 aggregate 與逐族職務，玩家／AI 共用。
- Docker 聚焦測試 `go test ./internal/gamedata ./internal/shell -run 'Coloniz|SpecialExtraPopulation'`
  通過。
