# AI 領袖招募與任命規格

依據：`docs/re/ai-leader-assignment-audit-20260825.md`。

## 回合順序

1. 先處理各 AI 上回合留下的單一待聘 offer，再重做既有領袖任命。
2. 之後依玩家槽 0→N-1 順序執行隨機領袖 offer；玩家使用 `MercPool`，AI 使用各自的 `LeaderOffer`。同一候選不得同時出現在任何帝國、offer 或拒絕 cooldown。
3. AI 拒絕的候選進入 30 回合 cooldown；每回合遞減，歸零後重新可選。這是原版 status 4／ETA 的 typed remake 表示。
4. 熱座的其他真人席位只推進自己的 offer；AI 決策與 cooldown 屬世界側，每個世界回合只能推進一次。

## AI 接受規則

- 至少有 Assassin、Famous、Megawealth、Operations、Spymaster、Telepath、Helmsman、Weaponry 之一才有接受意願。
- Researcher 只在 AI 有有效研究主題時算偏好；Trader 只在 AI 為 Repulsive 時算偏好。
- 對應類型已有四名領袖時拒絕。
- 嚴格要求 `treasury > hireCost + 50`；接受時扣除 `hireCost`、加入 AI `Leaders`，拒絕則進 cooldown。offer 在兩種結果後都清除。

## 任命近似

- 艦長：每位未指派艦長依序配到尚無軍官、戰力最高的 AI 艦艇；寫入 `OfficerName/OfficerID`，並把領袖標成 active assignment。
- 殖民地領袖：每位未指派管理領袖配到尚無領袖、人口與產出最高的殖民地；套用同一套 typed 領袖加成並保存 `ColonyLeaderNames`。
- raw 艦艇五欄與殖民地 callback 沒有一對一資料時明確標為近似，不以固定種族 Commando 取代動態鏈。

## 存檔與測試

- `LeaderOffer`、`LeaderLastOfferTurn`、`ColonyLeaderNames`、全局 cooldown 與既有 `Leaders` 必須 JSON 往返。
- 測試涵蓋技能 gate、嚴格 50 BC reserve、拒絕 cooldown、offer 延遲、四席限制、艦艇／殖民地任命與存檔往返。
