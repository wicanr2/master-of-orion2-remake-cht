# 星系封鎖規格

規格狀態：`.GAM` typed 保存、可表示艦隊 producer、AI 職務 consumer 與 AI 對真人
封鎖積怨 `CONFORMED`；真人對真人 policy `DRAFT`。

證據見 [`../re/blockades-audit-20260828.md`](../re/blockades-audit-20260828.md)。

## Typed 契約

- `Star.BlockadedMask` 的 bit `p` 表示 player slot `p` 在該星遭封鎖。
- `Star.BlockadedBy[p]` 的 bit `q` 表示 fleet owner `q` 正在封鎖 player `p`。
- `.GAM` 匯入必須逐 byte 保存兩者，不得把非零值壓成 bool。
- 正常回合重算必須先整表清除，再只由有效、已抵達的艦隊重建。

## 已接線

- 玩家、非目前熱座席位與 AI 的逐殖民地星系資料共同形成 occupied owner mask。
- 只採 `ETA==0`、合法所在星且有艦艇的玩家／熱座／AI 艦隊；policy raw `>=4`
  才封鎖其他帝國，不封鎖自己的同星殖民地。
- 停泊事件怪物與已抵達安塔蘭出征艦隊走 owner `>=8` 分支，封鎖同星所有殖民者，
  不填一般 `BlockadedBy`。
- `GameSession.EndTurn` 在艦隊移動與 AI 突襲後重算，下一回合 AI 職務分配先消費該 mask。
- 真人艦隊封鎖 AI 時消耗 `-Random_(5)`，依 `Change_Relations_` 的關係、政體與
  Charismatic 修飾後向下除四，寫入 AI 的 `OriginalBlockadeGrievanceRaw`；戰時一般
  relation score 不變。舊存檔若沒有 typed 關係仍消耗亂數，但不猜寫積怨。

## 仍未 READY

1. 真人對真人的正式外交矩陣；目前熱座／現代多人沒有可表示的雙向 raw policy。
2. raw policy 6 的正式名稱；runtime 比較保留 `>=4`，不把名稱未知變成行為缺口。
3. 真人側反向 `+0x6BF` 沒有現行玩家可見 consumer，暫不建立無消費端影子欄位。
