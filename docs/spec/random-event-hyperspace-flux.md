# 隨機事件 9「超空間亂流」規格

證據來源：[`../re/random-event-hyperspace-flux-audit-20260825.md`](../re/random-event-hyperspace-flux-audit-20260825.md)。

## 建立與生命週期

1. 事件是全銀河唯一持續 record，不選帝國、殖民地或星系；已 active 時不得重複追加。
2. 建立當回合為第 1 次 consumer。前五次 consumer 不會解除。
3. 第六次起每回合擲 `Random(20)==1`；成功解除。進入 consumer 時年齡大於 20，
   即使骰值失敗也強制解除。
4. record 的年齡、active 狀態必須經 JSON、熱座及多人快照往返。

## 航行效果

1. active 時，所有非跨維度帝國的在途艦隊保持原 `ETA`、`AtStar` 與 `DestStar`；不抵達、
   不觸發探索、水雷或發現。
2. active 時，非跨維度玩家／熱座席位的 `SendFleet` 回 false，且不得改寫原任務。
3. 非跨維度 AI 的在途 `FleetETA` 不遞減，也不得發起玩家突襲或 AI 對 AI 航程。
4. 玩家或 AI 具有 `TRAIT_TRANS_DIMENSIONAL` 時完全免疫：既有航程正常推進，也可建立新航程。
5. 事件解除發生在該回合航行計算之後，因此解除當回合仍凍結；下一回合才恢復航行。

## 顯示與驗收

- 建立與解除各提供繁中及英文訊息；持續凍結的每回合不製造重複洗版訊息。
- 純規則測試覆蓋第 5／6 次 consumer、5% 成功、age>20 強制解除、玩家新命令、玩家在途、
  AI 在途／新命令、跨維度例外與存檔往返。
- 1.31／1.50 的彗星與怪獸候選衝突不在本規格假裝閉合，保留為版本 profile 待辦。

