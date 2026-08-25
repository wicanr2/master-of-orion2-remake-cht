# 同化進度與種族修正規格

## 規則

1. 每人口同化門檻固定 240 點。
2. 八政體基礎進度依序為 `30/60/30/60/60/120/12/16`。
3. 有異族管理中心時，基礎進度改為 120。
4. Charismatic 非零時進度乘 2；否則 Repulsive 非零時進度除 2。兩項同存時依原版分支只套 Charismatic。
5. 每回合把 rate 加入殖民地 raw 進度；每滿 240 就同化一人口並減 240，允許同回合跨多個門檻。
6. 完成最後人口後進度清零，並立即重算未同化士氣與叛亂相關狀態。

## UI 與存檔

- 每人口顯示回合數為 `ceil(240/rate)`。
- 全部剩餘 ETA 為 `ceil((未同化人口*240-目前進度)/rate)`，下限 1。
- JSON 保存進度格式版本。舊版「回合計數」在 restore 時乘上當下原版 rate，再夾到 0..239；新存檔直接保存 raw 點數。

## 驗收邊界

- 八政體一般、Charismatic、Repulsive 的 rate 與回合數表。
- Charismatic＋Repulsive 同存時採 Charismatic。
- 異族管理中心的一般 2／Charismatic 1／Repulsive 4 回合。
- 銀河統一 Charismatic 為 8 回合，防止用整數 `15/2=7` 取代 rate 模型。
- 政體或建築在有餘數時變更，不重解或遺失既有 raw 進度。
- 舊 JSON 遷移、存讀檔與 UI ETA 往返。

逆向證據見 [`../re/assimilation-race-traits-audit-20260825.md`](../re/assimilation-race-traits-audit-20260825.md)。
