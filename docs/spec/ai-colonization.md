# AI 殖民規格

規格狀態：實際殖民來源、抵達閘門、同星系五軌道選點與母星原生重力 `CONFORMED`；
單主力艦隊跨星航線 `APPROXIMATED`；多艦隊與 Colony Base 來源 `DRAFT`。

證據見 [`../re/ai-colonization-audit-20260828.md`](../re/ai-colonization-audit-20260828.md)。

## 已接契約

1. 新局 AI 的 Average 開局資料包含一艘 typed `COLONY_SHIP`。
2. AI 沒有 Colony Ship 時不得建立遠距殖民地；成功殖民後必須消耗該艦並重算軍力。
3. 殖民地建立仍走玩家／AI 共用 typed 行星建構器，並同步所有 AI 殖民地平行陣列。
4. AI 母星的 `PlanetGravity` 必須等於該種族 `RaceGravity`；Low-G／High-G 不得在自己母星吃
   非原生重力懲罰。
5. 精確 AI 職務器接線後，30／40 回合正常路徑仍須同時保有工業、軍力成長及至少一次合法擴張。
6. `All_AI_Colonize_` 每回合只消費目前艦隊所在星系的 Colony Ship；`FleetETA>0` 時不得建立
   殖民地。跨星目標必須先保存 `FleetDestStar/FleetETA`，抵達後的下一次殖民掃描才能消耗船。
7. 同星系只掃 `PlanetsAt(star)` 的五個軌道，以 `sub_D27A7` 對每顆未殖民行星計基礎價值，
   嚴格較高才替換；不得以第一顆可殖民行星或全圖 contextual 分數代替。

## 尚未閉合

- 跨星選路仍以單一主力艦隊代理：含 Colony Ship 的整支艦隊一同航行，尚未表示每艘船與多支
  艦隊。此 adapter 已有 ETA／抵達閘門，但不是原版多艦隊 route planner。
- Colony Base 的同星系建造、消耗與來源殖民地人口／產品鏈尚未接 AI。
- 多艘 Colony Ship 的獨立位置、領袖解除任命與多艦隊並行仍為 DRAFT。
