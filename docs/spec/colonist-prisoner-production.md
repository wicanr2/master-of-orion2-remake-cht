# 逐職務未同化人口產出規格

## 狀態契約

`ColonyState` 保存 `UnassimilatedFarmers`、`UnassimilatedWorkers`、
`UnassimilatedScientists`。三者總和等於 `UnassimilatedPop` 時，視為完整逐職務狀態；
否則視為舊 JSON，按既有總人口比例安全回退。

## 產出

- 一般人口每人產出 `perUnit`。
- 該職務 prisoner 每人產出 `floor(perUnit × 3 / 4)`。
- 工人仍套原版「每工人至少 1 產能」下限；農夫與科學家不套。
- 固定建築產出、士氣、領袖與重力的既有套用位置不因本切片改變。

## 狀態變更

- `.GAM`：由前 `Population` 筆 packed colonist 的 job 與 `PRISONER` 精確建立三欄，
  並匯入 `AssimilationProg`。
- 征服：目前全部人口成為 prisoner，三欄直接複製當下職務總數。
- 玩家改派職務：若逐職務狀態完整且來源職務有 prisoner，連同一名 prisoner 移動。
  這對應 remake 未提供「點選 citizen 或 prisoner 圖示」的 UI 限制，屬可重播近似。
- 同化：每完成一單位，固定依農夫、工人、科學家順序清除一名 prisoner；原版精確 packed
  選人順序仍未知，不宣稱 parity。

## 驗收

1. 同樣三名 prisoner 放在不同職務時，只降低對應產出。
2. `.GAM` adapter 能保留 prisoner 的職務與同化進度。
3. 征服、改派、同化後三欄總和持續等於 `UnassimilatedPop`。
4. 舊 JSON 只有 `UnassimilatedPop` 時仍走比例 fallback，不改變既有存檔結果。
