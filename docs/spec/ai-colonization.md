# AI 殖民規格

規格狀態：實際殖民來源與母星原生重力 `CONFORMED`；多艦隊航線與 Colony Base 來源 `DRAFT`。

證據見 [`../re/ai-colonization-audit-20260828.md`](../re/ai-colonization-audit-20260828.md)。

## 已接契約

1. 新局 AI 的 Average 開局資料包含一艘 typed `COLONY_SHIP`。
2. AI 沒有 Colony Ship 時不得建立遠距殖民地；成功殖民後必須消耗該艦並重算軍力。
3. 殖民地建立仍走玩家／AI 共用 typed 行星建構器，並同步所有 AI 殖民地平行陣列。
4. AI 母星的 `PlanetGravity` 必須等於該種族 `RaceGravity`；Low-G／High-G 不得在自己母星吃
   非原生重力懲罰。
5. 精確 AI 職務器接線後，30／40 回合正常路徑仍須同時保有工業、軍力成長及至少一次合法擴張。

## 尚未閉合

- 目前 `aiExpand` 仍以單一主力艦隊代理 Colony Ship 位置，尚未按原版逐星記錄只在船已抵達的
  星系殖民。
- Colony Base 的同星系建造、消耗與來源殖民地人口／產品鏈尚未接 AI。
- 多艘 Colony Ship、領袖解除任命、五行星原版 `sub_D27A7` 精確排序與多艦隊並行仍為 DRAFT。
