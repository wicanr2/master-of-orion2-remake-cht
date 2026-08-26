# 隨機事件 2：彗星來襲靜態稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4；位址為 DOS/4GW LE object #1 的 IDA 線性位址。
- 唯讀匯出器：`tools/ida/audit_event_comet.py`。
- 下列指令、欄位資料流與公式標為**已證實**；remake 艦隊缺少原版逐段航行 status
  所造成的表示差異另標為**近似**。

## 建立端

`sub_2230A @ 0x22908..0x229C7` 的事件 2 分支：

1. 以 `sub_23DA0` 從目標帝國抽殖民地，將 colony index 寫入事件 record `word +3`。
2. 以 `sub_23E60` 拒絕同殖民地既有事件 2、14、16、17、24、25；無目標或衝突即取消。
3. `word +5 = word +7 = 10 × (Random(5)+10+difficultyRaw)`。兩欄分別是目前耐久與
   初始耐久；由逐回合扣減與 GNN 百分比 consumer 交叉證實。
4. `byte +2 = Random(5)+10-difficultyRaw`，是剩餘倒數。
5. 一般事件排程另把事件 2 的最早日期設為 elapsed turn 200；既有
   `random-event-schedule-audit-20260825.md` 已閉合此門檻。

## 每回合攔截與狀態

`sub_206A2 @ 0x20848..0x20918`：

1. 事件 state 6 不再處理；其他狀態先由 colony → planet → star 取得目標星系。
2. `sub_23B28 @ 0x23B28..0x23B64` 掃描 500 筆 0x81-byte ship record。只有
   `ship+0x65 == targetStar` 且 `ship+0x64 == 0` 的艦艇加入攔截；每艘貢獻
   `ship+0x10 + 1`。`ShipDesign` 的 16-byte 名稱後即 `Size`，故 `+0x10` 是艦體級 0..5。
   helper 不檢查 owner，因此所有符合條件的停泊艦艇都協助攔截。
3. 每回合以攔截總和扣 `word +5`。若仍大於 0，倒數 `byte +2` 減 1；歸零時呼叫
   `sub_23780` 結算撞擊並把 state 設成 5。
4. 若耐久已不大於 0，事件不再減倒數或撞擊；state 設成 5，代表成功攔截。
5. `Random(20)==1` 且目前耐久不同於初始耐久時可把 state 設成 4，以播報進度。
   `sub_21371 @ 0x21410..0x21476` 以
   `(initial-current)×100/initial`、倒數及 state 4／5／6 組 GNN 變體；這是播報
   狀態，不改變上述玩法公式。

## 撞擊傷害

`sub_23780 @ 0x23780..0x23833`：

1. 計數殖民地 49 格 `colony+0x136+buildingID` 建築旗標，並加人口 `colony+0x0A`。
2. `damage = max(1, (population+rawBuildingCount) × (Random(3)+Random(3)) / 10)`。
3. 將 damage 送入 `sub_DD2F2`／`sub_DCEBD` 的一般戰略殖民地傷害鏈；因此彗星
   不是無條件刪除整座殖民地，而是傷及一般建築、駐軍、建造進度與人口，最後人口
   歸零時才可能摧毀殖民地。

## Remake 邊界

- **已證實**：建立公式、200 回合門檻、事件互斥、逐回合 `Size+1` 攔截、無 owner
  過濾、撞擊公式及共用戰略傷害鏈。
- **近似**：原版 `Status==0` 是逐艦 record 狀態；remake 以 `Fleet.ETA==0` 且
  `AtStar==targetStar` 表示停泊，AI 以 `FleetETA==0/FleetStar` 表示。這保留玩家可見
  的「已抵達艦艇才能攔截」，但不是 raw status 的一對一表示。
- **已接線**：`sub_23DA0` 的 raw `colony+0x13F==0` 是排除 Capitol 殖民地；remake 已以
  玩家／AI 各自的 `ColonyBuildings` typed 狀態限制候選。
- **相依缺口**：原版互斥表包含事件 14；remake 尚未建立海盜活動持續 record，目前事件 14
  不會進事件池，故現況不可能重疊。實作事件 14 時必須把其 planet/star record 接回
  `cometTargetConflicted`，不能據此文件宣稱該相依已永久閉合。
