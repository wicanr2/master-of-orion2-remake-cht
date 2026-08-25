# 先進級開局科技應用選擇規格

## 範圍

先進級開局在固定六個主題後，再完成十九次選擇。每次必須從當下研究樹前沿的科技應用直接評分與抽選，選中的應用同時決定完成的主題。

## 規則

1. 每次重算各領域第一個未完成主題，不能跳過前置主題。
2. 人類玩家使用 `sub_589D6` 初始化 raw4，再套 `sub_FC845` 已閉合的人類與共用估值鏈；
   raw4 的原始欄位位址已證實，但不以英文符號推定其語意名稱。
3. AI 使用 `sub_589D6` raw profile、`sub_FC845` AI／共用估值與 `sub_FD335`
   難度門檻；完整契約見 [`ai-starting-tech-valuation.md`](ai-starting-tech-valuation.md)。
4. 分數採 `value × horizon / max(1, topicCost/researchPerTurn)`；視野從 15 起，必要時乘 `3/2`。
5. 一次選擇只可消耗一次加權抽選亂數。玩家與各 AI 使用彼此獨立、可重播的種子流。
6. `ResearchAll` 主題取得全部應用；其他主題只把被抽中的應用寫入 `ChosenTech`／`ExplicitChoice`。
7. 存檔、熱座與 TCP 快照沿用既有 `PlayerState` 欄位，不新增不可往返的暫態狀態。
8. 正式 UI 必須在最終種族與政府完成後才定案開局抽選；規格見
   [`human-starting-tech-profile.md`](human-starting-tech-profile.md)。

## 驗收

- 同種子得到相同主題與應用；不同種子至少在主題或應用層不同。
- 十九次後研究樹沒有越過未完成前置。
- 單一前沿抽選只有一次 RNG 呼叫。
- 固定公式樣本（Automated Factories）在難度 0／2 的估值分別為 300／900。

逆向依據見 [`../re/starting-tech-application-audit-20260825.md`](../re/starting-tech-application-audit-20260825.md)。
