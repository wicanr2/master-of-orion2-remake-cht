# 研究/科技系統:現況與「抉擇機制」還原計畫

> 日期:2026-07-10。目的:精確記錄研究系統**已忠實**與**尚缺**的部分,並給下輪執行「每主題數科技間抉擇」的乾淨計畫。避免半套實作(專案鐵律:對齊原版、不急就章、不自編)。

## 已忠實(可信,勿重做)

- **真科技樹資料**:`internal/gamedata/techtree.go` 的 `researchChoices[83]` 逐字轉寫自 openorion2 `tech.cpp:169–305`,含每個 `ResearchTopic` 的**真 RP 成本**、可選科技清單(`Choices`)、`ResearchAll` 旗標。公開 accessor `gamedata.ResearchChoiceFor(topic)`。
- **真成本已接入**:`shell.ResearchCost(t)` 直接取 `ResearchChoiceFor(t).Cost`;`engine/research.go` 的 `RunResearchPhase` 用真成本判定完成,並**保留溢出 RP 結轉**下一主題。→ 研究「要花多少點」已對齊原版。

## 尚缺的核心機制:每主題「抉擇一項科技」

原版 MOO2:完成一個研究主題後,若該主題 `ResearchAll=false`,玩家要在 `Choices` 的數個科技中**選一項**解鎖(其餘永久放棄,除非 Creative 特性);`ResearchAll=true` 才全解。這是 MOO2 招牌取捨,`HONEST-STATUS.md` 亦點名為缺口。

現況(不忠實):
- `engine/research.go:9–11` 註明「只回答主題是否完成,**不選擇** Choices 中哪一項」。
- `shell.ComponentUnlocked`(session.go:102)以 **`CompletedTopics[topic]`**(主題層級)判解鎖 → 等於**一次解鎖該主題全部選項**,無取捨。

## 執行計畫(下輪,乾淨落地)

分三步,每步可單元測試 + headless 驗證:

1. **模型層(engine)**:
   - `PlayerState` 加 `ChosenTech map[ResearchTopic]Technology` 與 `PendingChoice ResearchTopic`(0=無)。
   - `RunResearchPhase` 完成非 `ResearchAll` 主題時:設 `PendingChoice=topic`、**不**自動推進;`ResearchAll` 主題:把全部 `Choices` 記入 `ChosenTech` 並推進。
   - 加 `ChooseTech(topic, tech)`:驗證 tech ∈ `Choices`,寫入 `ChosenTech`,清 `PendingChoice`。
2. **解鎖 gating 改科技層級**:`ComponentUnlocked` 改為「該元件對應的 `Technology` 已在 `ChosenTech`(或其主題 `ResearchAll`)」。需確認元件↔Technology 對應(目前 Component.Tech 是 ResearchTopic,須擴充為 Technology 或加映射)。這是最需小心的一步,改完跑既有 `techtree_verify_test` 與艦艇設計解鎖測試護欄。
3. **抉擇 UI**:`PendingChoice` 非空時,研究畫面(SCIENCE.LBX)彈出該主題的 `Choices`(用真科技名 TECHNAME.LBX + i18n),玩家點選 → `ChooseTech`。AI 由 decider 自動選(可先取第一項或依性格)。
   - headless:腳本點選一項,截圖驗證。

## 驗收

- 完成一主題後,只有**被選的**科技對應元件解鎖(其餘不解),對照原版行為。
- 既有測試(engine/research、techtree_verify、艦艇設計)全綠。
- 對原版實測:同一主題選不同科技,艦艇設計可用元件不同。

## 注意

- `session.go:866` 事件「科學突破 +150 RP」是隨機事件加成(合理),非主線成本,保留。
- 種族 Creative/Uncreative 特性影響「解全部/只解一項」——待特性系統一併處理(見 `custom-race-picks.md`)。
