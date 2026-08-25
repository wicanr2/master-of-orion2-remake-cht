# 隨機事件 14「海盜活動」規格

證據來源：[`../re/random-event-pirate-activity-audit-20260825.md`](../re/random-event-pirate-activity-audit-20260825.md)。

## 建立

1. 在被選中的帝國殖民星系中，以均勻 reservoir sampling 選一個候選星系。
2. 若同星系已有彗星、瘟疫、人口暴增、時空異象、海盜活動或超新星，建立失敗。
3. 找出該星系內至少有一座殖民地的所有帝國，累加其 `ActiveFreighters` 為 T。
4. 初始強度依難度為 `T/5, 2T/5, 3T/5, T, 4T/5`；整數除法向零截斷。
   小於 5 時事件不成立。成功時目前與初始強度相同。

## 每回合

1. 計算 `lossChance=floor(currentStrength*100/initialStrength)`。
2. `Random(100) <= lossChance` 時，星系內每個 `ActiveFreighters>0` 的帝國各減一艘。
3. 所有停在該星系且 `ETA==0` 的玩家、熱座與 AI 艦艇，不分 owner，每艘以
   `shipSizeClass+1` 扣減海盜強度。
4. 強度降至 0 或以下時事件結束；否則保存剩餘強度供下一回合繼續。

## 垂直鏈與驗收

- 事件表 ID 14 標示已實作，玩家、非目前熱座席與 AI 都能被排程建立。
- `PersistentEvent` 保存星系、目前強度與初始強度，JSON 往返不得遺失。
- 測試至少覆蓋五級難度公式、低於 5 取消、跨帝國運輸船損失、不分 owner 的同星清剿、
  事件互斥，以及存檔往返。
- 英文模式需有安全播報，不以中文內容充當英文 fallback。

