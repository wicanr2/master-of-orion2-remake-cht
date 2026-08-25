# 快速戰鬥殖民地耐久規格

## 來源與範圍

本規格只實作 [`Get_Colony_Hits_` 靜態稽核](../re/colony-hits-audit-20260824.md)
已證實的耐久公式。它不定義轟炸後人口／駐軍／建築的傷亡分配。

## 輸入

- `population`：殖民地人口。
- `soldiers`：殖民地士兵數。
- `tanks`：殖民地戰車數。
- `rawBuildingIDs`：目前存在的原版 0..48 建築 ID 集合。

負的人口或部隊值按 0 處理；重複 ID 只計一次。範圍外 ID 忽略，不將錯誤資料
轉成耐久。這是型別安全邊界，不宣稱原版對損壞存檔採相同行為。

## 規則

```text
total = max(population, 0) + max(soldiers, 0) + max(tanks, 0)
total += 40 × count(unique valid building IDs excluding 8, 40, 41)
```

`8/40/41` 是 Battlestation／Star Base／Star Fortress，快速戰鬥另建戰鬥者，
不得重複加進殖民地本體。

## 接線

- `internal/gamedata.OriginalColonyCombatHits` 保存純規則。
- `BombardColony` 以轟炸後人口、駐軍與建築重算 `PlanetHitsRequired`，取代手冊
  近似的 `GroundPlanetTotalHits` 顯示值。
- 建築名稱必須先由既有 `OrigBuildingID` 對回 raw ID；無法映射者不冒算。

## 驗收

- 人口 8、士兵 4、戰車 2、兩棟一般建築應為 `94`。
- 三種軌道設施不增加本體耐久。
- 重複、範圍外 ID 與負數輸入不產生額外耐久。
- `BombardColony` 的顯示值會隨駐軍與一般建築改變，且不把星基重複計入。
