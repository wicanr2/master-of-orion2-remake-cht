# 隨機事件 Capitol 目標篩選規格

## 證據範圍

`sub_23DA0 @ 0x23DA0` 逐殖民地掃描並以 reservoir sampling 在合格候選中等機率選取；
原始條件包含 `colony+0x06==0` 與 `colony+0x13F==0`。後者已由
`docs/re/capitol-state-audit-20260826.md` 的建築槽、士氣 consumer、攻陷移除與重建鏈證實為
raw building 9 Capitol。相關事件 caller 與專用抽選器分別見地震、瘟疫、彗星、工業事故、
礦產及超新星 RE 文件。

## Typed 規則

1. 玩家、熱座目前席位與 AI 的 Capitol 狀態，一律由各自
   `ColonyBuildings[colony][CapitolBuildName]` 判定；不得以殖民地索引 0、母星或
   `CapitolPlanet` 單獨代替已完成建築槽。
2. 地震、瘟疫、彗星、工業事故與事件 11 礦產枯竭的候選必須排除已有 Capitol 的殖民地。
   Reservoir 計數只包含通過此條件的候選，不能先抽中 Capitol 後直接使整個事件失敗。
3. 事件 12 礦產發現使用 `sub_23D44`，不套 `sub_23DA0` 的 Capitol 條件。
4. 超新星候選星必須至少有一座 active 且無 Capitol 的殖民地；一旦事件成立，其 RP 消費與
   爆發效果仍依原版掃描該星全部 active 殖民地，不排除同星 Capitol。
5. 舊存檔或不完整測試夾具缺少建築 map 時視為「沒有 Capitol」；不得自行製造失都狀態。

## 驗收

- 只有 Capitol 與一般殖民地各一座時，上述單殖民地事件只能選一般殖民地。
- 全部候選都有 Capitol 時，事件不成立且不 panic。
- 礦產發現仍能選有 Capitol、礦產低於 Ultra Rich 的殖民地。
- 超新星不選「全數皆有 Capitol」的星，但同星另有一座無 Capitol 時可選。
- 玩家與 AI 使用相同 selector；熱座由換席後目前玩家資料或全銀河星系 helper 讀取正確席位。
