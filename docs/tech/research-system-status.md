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

## 執行計畫與進度

1. ✅ **模型層(engine)完成**(2026-07-10,非破壞、有測試):
   - `PlayerState` 加 `ChosenTech map[ResearchTopic]Technology`、`PendingChoice`、`HasPendingChoice`。
   - `RunResearchPhase` 完成主題時 `recordCompletion`:ResearchAll/單選直接記;**多選預設記第一項並開 PendingChoice**(不阻塞回合,玩家可改選)。
   - `engine.ApplyResearchChoice(ps, tech)` 驗證合法選項後改選、清待決。
   - shell:`PendingResearchChoice()` / `ChooseResearchTech(tech)` / `ChosenTechFor(topic)`。
   - 測試:`internal/engine/research_choice_test.go`、`internal/shell/research_choice_test.go`(多選預設+改選+非法拒絕+ResearchAll 不待決)。
2. ✅ **解鎖 gating 改科技層級完成**(2026-07-10,非破壞、有測試):
   - 元件↔真科技校正:依 `docs/tech/component-tech-mapping.md` 把各元件掛正確主題 + `UnlockTech`(真 Technology)。里程碑(死光/氙素裝甲)/抽象(戰鬥電腦/重生程序)元件 `UnlockTech=TECH_NONE`(proxy 主題,待重設計)。
   - `PlayerState.ExplicitChoice`:`ApplyResearchChoice` 標記玩家明確抉擇過的主題。
   - `ComponentUnlocked`:未映射/未明確抉擇→主題層級(非破壞,AI/預設不回歸);已明確抉擇→僅所選科技對應元件解鎖。
   - `researchQueue` 自元件 `.Tech` 蒐集主題,校正後深層主題自動納入研究、逐步解鎖(不永久鎖)。
   - 測試:`component_gating_test.go`(明確抉擇收斂/未抉擇主題層級);既有 `TestResearchUnlockLoopOverTurns` 續綠(非破壞)。

   **→ 研究系統忠實化三步全部完成**:真成本 + 真選項抉擇 UI + 抉擇反映到元件解鎖。剩「戰鬥電腦/重生程序/里程碑科技」等資料模型層級的元件重設計(需 Component.Tech 支援里程碑語意),屬小尾巴。

<details><summary>校正發現(存查)</summary>

深入盤點(2026-07-10)發現的**資料校正需求**:
   remake 元件目前掛的 `Tech`(ResearchTopic)**大多與真科技樹的選項對不上**,無法乾淨映射到具體 Technology:
   - ✅ 對得上:質量投射器→`TECH_MASS_DRIVER`(在 ADVANCED_MAGNETISM 選項內)、麥克萊特飛彈→`TECH_MERCULITE_MISSILE`(在 ADVANCED_CHEMISTRY 內)。
   - ❌ 對不上:中子爆破槍掛 ADVANCED_CHEMISTRY,但該主題真選項是 {麥克萊特飛彈, 污染處理器},**無中子爆破槍**;核融合光束掛 ADVANCED_FUSION,真選項是 {增壓引擎, 核融合彈, 核融合引擎},**無光束**;高斯砲/相位砲/電漿砲同類對不上。
   → **前置工作(下輪)**:先把每個 remake 武器/裝甲/護盾元件**重新對應到它真正的 Technology + 正確主題**(對照 tech.tsv/techtree.go 真資料;可派子代理逐一核),校正 `Component.Tech`。校正後再:
     - `Component` 加 `UnlockTech Technology`;`PlayerState` 加 `ExplicitChoice map[topic]bool`(`ApplyResearchChoice` 設 true)。
     - `ComponentUnlocked`:未映射→維持主題層級(非破壞);已映射且該主題有明確抉擇→僅 `ChosenTech[topic]==UnlockTech` 解鎖;完成但未明確抉擇→主題層級後備(非破壞,AI/預設不回歸)。
   這樣抉擇才會真正反映到艦艇設計可用元件。(以上為校正前的分析,校正已於當日完成。)
</details>

3. ✅ **抉擇 UI 完成**(2026-07-10,可玩、headless 渲染驗證):
   - `gamedata/technames.go`:`Technology → 英文名`(203 條,對 tech.tsv 驗證;8 個 HYPER 填充項無名)。
   - `cmd/moo2/researchchoice.go`:回合結束若 `PendingResearchChoice` 非空 → 抉擇畫面(RACEOPT 框 + 真科技選項,經 TECHNAME/tech.tsv 中文化);點選 → `ChooseResearchTech` → 回合摘要。
   - 接線:`galaxy` 結束回合後偵測待決抉擇導向此畫面。
   - 驗證:進階建築學 → 自動化工廠/重型裝甲/行星飛彈基地(真資料),end-to-end 流程跑通。
   - AI 目前用預設第一項(decider 依性格選為後續小改)。

**目前狀態總結(2026-07-10)**:研究「每主題數科技間抉擇」**三步全部完成**——真成本 + 真選項抉擇 UI + 抉擇反映到元件解鎖。玩家研究一個主題、選定一項科技後,只有該科技對應的艦艇元件解鎖(明確抉擇),AI/預設維持主題層級不回歸。剩餘小尾巴:戰鬥電腦/重生程序/里程碑科技(死光/氙素裝甲)等元件的資料模型重設計(`Component.Tech` 目前只支援單一 ResearchTopic,無法表達電腦研究鏈/里程碑/種族特性語意)。

## 驗收

- 完成一主題後,只有**被選的**科技對應元件解鎖(其餘不解),對照原版行為。
- 既有測試(engine/research、techtree_verify、艦艇設計)全綠。
- 對原版實測:同一主題選不同科技,艦艇設計可用元件不同。

## 注意

- `session.go:866` 事件「科學突破 +150 RP」是隨機事件加成(合理),非主線成本,保留。
- 種族 Creative/Uncreative 特性影響「解全部/只解一項」——待特性系統一併處理(見 `custom-race-picks.md`)。
