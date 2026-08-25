# 安塔蘭週期入侵規格

## 目標

依 `docs/re/antaran-periodic-invasion-audit-20260825.md`，把自訂 20／15 直接傷害腳本
改為可存檔的全局「資源 → 建艦 → 出兵 → 航行 → 戰鬥」狀態機。

## 已證實規則

- 科技等級延遲：曲速前 200、一般 100、先進 0 回合。
- 延遲後每 25 回合 pulse；第 `n` 個 pulse 的單份資源為
  `ceil(n×difficultyScale/100)`，scale=`100/100/100/150/200`。
- 每個 pulse 一份 offensive；第二份在 defensive 未滿時給 defensive，滿後給 offensive。
- 成本 `{2,5,12,30,75}`；offensive 上限 `{4,4,3,2,2}`；defensive 上限
  `{0,0,3,2,7}`。
- readiness、`Random(200)`、全帝國目標權重、Lucky `/3`、目標殖民星均勻抽樣、
  每次最多五艘與最大可用艦級部署，均依 RE 文件實作。
- 世界每回合只推進一次，熱座席位不得重複累積或重複出兵。

## Remake 資料模型

- `AntaranInvasionState` 保存資源、五級艦數、已部署數、readiness、成本是否已降階與
  pending fleets。
- `AntaranRaidFleet` 保存目標種類／索引／星系、五級艦數與 ETA，可經 JSON 往返。
- 原版中途座標無法由現有 ETA 模型逐點表示；remake 以明示的 route approximation 換算 ETA，
  不宣稱座標、截擊時點或原版 PRNG 位元一致。

## 戰鬥

- 玩家／熱座目標：抵達星系後，先讓同星停泊艦隊使用既有快速戰鬥 `battleVolley` 參戰，
  不得直接扣國庫。殖民地固定防禦尚未接入這條戰鬥者清單，必須與 owner 8 的戰後殖民地
  消費端一併留在 `WORKLIST.md`；在接線前不得把「同星艦隊戰鬥已完成」擴張成整個星系
  防禦已對齊。
- AI 目標：以 AI 已保存的艦型／聚合艦力進入同一強度尺度；不得固定把所有攻擊轉嫁玩家。
- 安塔蘭勝利只表示本次攻擊艦隊獲勝，不等於安塔蘭母星勝利狀態；安塔蘭被擊敗後清除 pending。
- owner 8 的精確戰術 record 尚未全閉合時，快速戰鬥映射必須標為可表示近似並留在 WORKLIST，
  不得用綠測試宣稱完整戰術 parity。

## 驗收

1. 曲速前／一般／先進的第一個 resource pulse 邊界測試。
2. 五級難度資源無條件進位、defensive 滿額改流 offensive、成本／上限測試。
3. readiness 與 `Random(200)` 邊界、Lucky 權重、目標星 eligibility 測試。
4. 單人與熱座一個世界回合只推進一次。
5. pending fleet JSON 往返、抵達後不直接扣 BC、玩家／AI 各一條戰鬥抽樣。
6. 全專案 `go test ./... -count=1`、`git diff --check` 與 Docker 清理。
