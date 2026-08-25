# 事件殖民地研究轉用規格

## 狀態與資料流

- `engine.ColonyState.ResearchDiverted` 是單回合輸入，不是持久事件狀態。
- `RunColonyTurn` 仍照常產生完整 `ColonyOutput.Research`。
- `RunEmpireTurn` 聚合一般研究點時，略過 `ResearchDiverted=true` 的殖民地；其他殖民地、
  `FleetResearch` 與 `TreatyResearch` 正常加入。
- `GameSession.coloniesForTurn` 依 `PersistentSupernova` 的 `StarIndex`，在副本上設定
  `ResearchDiverted=true`；不得污染持久的 `PlayerColonies`。
- `stepSupernova` 從 `LastPlayerOutput.Colonies` 讀取受影響星系 RP，且只加入
  `PersistentEvent.ResearchDone`。

## 邊界

- 同星系多殖民地全部轉用。
- 其他星系研究仍投入目前一般研究主題。
- 超新星解除或爆發並移除事件後，下一回合不再轉用。
- 時空異象維持既有完全凍結，不需要另設 `ResearchDiverted`。
- 熱座席位共用 `coloniesForTurn`，不得走不同規則。

## 驗證

1. engine：兩座等產出殖民地，只有一座轉用時，一般研究只收到另一座；輸出陣列仍保留兩座 RP。
2. shell：超新星目標星系 RP 只增加搶救進度，不增加一般研究；其他星系 RP 仍增加一般研究。
3. shell：事件結束後研究恢復。
4. JSON／多人快照只保存 `PersistentEvents`，不保存暫態 `ResearchDiverted`；載入後由星系關係重建。
