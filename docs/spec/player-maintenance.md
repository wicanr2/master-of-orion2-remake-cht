# 玩家維護費規格

## 目標

每回合在一次帝國經濟結算中產生可追查的建築、運輸艦、指揮點超支、間諜、納貢與軍官分項；
不得在間諜或領袖後處理階段再次扣除同一筆費用。

## remake 契約

1. `prepPlayerDerived` 在收入結算前重算建築、間諜、軍官與指揮點供需。
2. `RunEmpireTurn` 計算運輸艦與指揮點超支，連同輸入分項一次寫入 `NetBC`。
3. `advanceEspionage` 只處理任務與傷亡；`advanceLeaderLimbo` 只處理領袖狀態。
4. 間諜分項為所有 `PlayerSpies` 正值加總，每名 1 BC；零值與負值不產生費用。
5. 軍官分項沿用 `LeaderUpkeepTotal`；Megawealth 免維護。
6. 納貢維持跨帝國轉帳階段，摘要以 `TributeCost` 顯示，不併入 engine 的本地輸入。
7. 國庫允許因一次結算變負；不得把維護費夾到現有 BC，否則會漏收。原版後續資產處分另列工作。

## 驗收

- 純 engine 測試驗證各分項只扣一次。
- shell 測試驗證間諜任務不直接扣款，`EndTurn` 輸出正確間諜分項。
- 領袖既有維護費與 Megawealth 測試保持通過。

