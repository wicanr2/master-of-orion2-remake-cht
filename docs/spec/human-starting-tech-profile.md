# 真人種族完成後的開局科技規格

## 正常新局順序

1. 先套 NEW GAME 難度、文明等級、銀河大小、年齡與對手數。
2. 建立銀河與 AI，開啟「種族尚未完成」pending 狀態。
3. 內建種族套用數值／布林特性與特性 0 政體；客製種族建立完整 31 格 runtime 特性後再套政府。
4. 在 pending 狀態以最終特性重建開局科技與母星開局建築，然後永久關閉 pending。

## 研究狀態契約

- 重建前清除 provisional `CompletedTopics`、`ChosenTech`、`ExplicitChoice`、Hyper 等級、pending choice 與研究進度；無條件保留 `TOPIC_STARTING_TECH`。
- 文明等級仍精確發放 1／6／25 個主題，先進級互斥應用仍只保存一項。
- 真人以自己的 runtime 特性抽 raw4／raw7，並將 raw4 送入 `sub_FC845` 共用後段。
- 每位 AI 仍使用自己的內建種族特性與獨立亂數流。
- 若 provisional `ResearchTopic` 已被開局發放完成，改指向重建後第一個可研究主題，進度歸零。
- pending 已關閉時，`ApplyGovernment` 不得改動任何研究狀態。

## 存檔／熱座

- pending 是只有新局畫面會使用的暫態，不存檔。
- 客製 31 格 runtime 特性需通過 JSON 與熱座席位往返；舊存檔零值安全。

證據見 [`../re/human-starting-tech-profile-audit-20260825.md`](../re/human-starting-tech-profile-audit-20260825.md)。
