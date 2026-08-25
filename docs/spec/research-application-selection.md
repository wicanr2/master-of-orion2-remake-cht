# 研究應用選擇與授予規格

證據來源：[`docs/re/research-application-selection-audit-20260825.md`](../re/research-application-selection-audit-20260825.md)。

## 持久狀態

- `ResearchTopic`：目前研究 field／topic。
- `ResearchApplication`：目前 topic 已選定、突破後要取得的科技 application。
- `HasResearchApplication`：避免合法 application 0 與「尚未選定」混淆。
- `ChosenTech`／`ExplicitChoice`：只表示已完成後真正擁有的科技，不得用來暫存尚未突破的選項。
- 舊 `PendingChoice` 可保留作 UI 待選旗標及舊存檔相容，但 pending 本身不得解鎖科技。

## 正常玩家規則

1. 選擇新 topic 時清除上一個目前 application 與研究進度。
2. `ResearchAll` 或 Creative topic 不需單選 application。
3. 單一 application 自動選定。
4. 一般多選 topic 在開始投入 RP 前顯示選擇 UI；選定只寫目前 application，不寫 `ChosenTech`。
5. Uncreative 在 topic 開始時由可存檔研究亂數流選定唯一 application，不顯示完整選項 UI。
6. 按結束回合時若一般玩家仍有待選 application，先開選擇 UI，不得先推進世界回合。
7. 突破時只授予已選定 application；Creative／`ResearchAll` 授予整組語意。
8. 完成後自動選下一 topic 時，立刻套用同一套 application 準備規則。

## AI 規則

- AI 在開始 topic 前先選 application。
- Creative 不單選；Uncreative 由該局可存檔研究亂數流限縮為一項；一般 AI 使用既有應用估值選擇。
- 突破後不得再以 post-completion fallback 暫時取得整組科技。

## 相容與驗收

- 舊存檔若只有突破後 `PendingChoice`，仍允許以舊相容分支完成一次改選；新流程不得產生此狀態。
- 研究前選定 application 經 JSON 存讀檔保持一致。
- 未突破前 `psKnowsTech` 對所選 application 必須為 false。
- 突破後只選項為 true；Creative 全部為 true。
- UI 點選研究領域後，多選 topic 先進 application 選擇，再回星系。
- 全套正常玩家、熱座、AI、存檔與研究解鎖測試通過。
