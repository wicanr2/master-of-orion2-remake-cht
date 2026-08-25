# 所有者種族重力產出規格

## 狀態

`ColonyState.RaceGravityKnown` 為真時，`RaceGravity` 保存目前所有者種族的
Low-G／Normal-G／High-G。未知或舊 JSON 一律回退 Normal-G，避免 Go 零值 Low-G
把舊存檔無端減產。

## 規則

- Planetary Gravity Generator 生效時懲罰固定為 0。
- 否則以 `GravityPenaltyPercent(planetGravity, raceGravity)` 套用食物、工業與研究。
- trait 同時含 High-G／Low-G 時採 High-G，對齊 `sub_DDF2C` 分支順序。
- 玩家、客製種族、AI 與 `.GAM` 匯入都必須在正常玩家路徑同步欄位。

## 明確未涵蓋

混合種族殖民地仍需要逐 colonist race 的重力值；本切片只修正 owner／單一種族人口，
不得據此把逐種族產出標成完成。
