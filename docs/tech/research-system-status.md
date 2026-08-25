# 研究/科技系統:現況與「抉擇機制」還原計畫

> 日期:2026-07-10。目的:精確記錄研究系統**已忠實**與**尚缺**的部分,並給下輪執行「每主題數科技間抉擇」的乾淨計畫。避免半套實作(專案鐵律:對齊原版、不急就章、不自編)。

## 已完成的 remake 資料與操作（不代表原版回合資料流已閉合）

- **真科技樹資料**:`internal/gamedata/techtree.go` 的 `researchChoices[83]` 逐字轉寫自 openorion2 `tech.cpp:169–305`,含每個 `ResearchTopic` 的**真 RP 成本**、可選科技清單(`Choices`)、`ResearchAll` 旗標。公開 accessor `gamedata.ResearchChoiceFor(topic)`。
- **成本與突破已接入**：`shell.ResearchCost(t)` 取 `ResearchChoiceFor(t).Cost`；
  `engine/research.go` 依原版超額比例擲突破率。成功後進度清零，不再保留自編的溢出 RP。

## 核心機制:每主題「抉擇一項科技」

原版 MOO2 在開始研究 field 時便選定 application；突破時授予既有選項。若
`ResearchAll=false`，一般種族在數項科技中擇一；Creative 突破時全解，Uncreative 在可選集合
形成時隨機限縮成一項；`ResearchAll=true` 本來就全解。IDA 證據見
[`../re/research-application-selection-audit-20260825.md`](../re/research-application-selection-audit-20260825.md)。

目前狀態（共同基礎模型）:
- `PlayerState.ResearchApplication` 保存研究中的 application；`ChosenTech` 只保存突破後真正取得的科技。
- `shell.preparePlayerResearchApplication`／`prepareAIResearchApplication` 在投入 RP 前套用一般、Creative、Uncreative 分支。
- `shell.ComponentUnlocked` 以 `ExplicitChoice` 區分一般種族明確擇一與 Creative／預設主題層級全解。

## 執行計畫與進度

1. ✅ **模型層(engine)完成**(2026-07-10,非破壞、有測試):
   - `PlayerState` 以 `ResearchApplication`／`HasResearchApplication` 保存目前選項；`ChosenTech`、`ExplicitChoice` 僅表示已取得科技。
   - `RunResearchPhase` 完成主題時授予研究前已選定項；舊存檔沒有此欄位時才開一次相容 PendingChoice。
   - Creative 不需單選；Uncreative 由可存檔研究亂數流在研究開始時限縮為一項。
   - `engine.SelectResearchApplication` 驗證選項但不解鎖；`ApplyResearchChoice` 只留作舊存檔相容。
   - shell:`PendingResearchChoice()` / `ChooseResearchTech(tech)` / `ChosenTechFor(topic)`。
   - 測試涵蓋研究前選擇、未突破不解鎖、突破授予、非法拒絕、ResearchAll 與存讀檔。
2. ✅ **解鎖 gating 改科技層級完成**(2026-07-10,非破壞、有測試):
   - 元件↔真科技校正:依 `docs/tech/component-tech-mapping.md` 把各元件掛正確主題 + `UnlockTech`(真 Technology)。里程碑(死光/氙素裝甲)/抽象(戰鬥電腦/重生程序)元件 `UnlockTech=TECH_NONE`(proxy 主題,待重設計)。
   - `PlayerState.ExplicitChoice`:`ApplyResearchChoice` 標記玩家明確抉擇過的主題。
   - `ComponentUnlocked`:未映射/未明確抉擇→主題層級(非破壞,AI/預設不回歸);已明確抉擇→僅所選科技對應元件解鎖。
   - `researchQueue` 自元件 `.Tech` 蒐集主題,校正後深層主題自動納入研究、逐步解鎖(不永久鎖)。
   - 測試:`component_gating_test.go`(明確抉擇收斂/未抉擇主題層級);既有 `TestResearchUnlockLoopOverTurns` 續綠(非破壞)。

   **→ remake 的研究資料、抉擇 UI 與元件解鎖三步已接通。** 原版
   `Check_For_Research_Breakthrough_ @ 0xE44E0` 的累積、突破 RNG、成功清零與完成入口已形成
   IDA→規格→Go 垂直證據；殖民地完整研究產出與應用選擇時序仍是核心忠實度工作。

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
   - `cmd/moo2/researchchoice.go`:點選研究 field 後若為多選，先進 application 選擇畫面；點選只設定目前 application，再回星系。
   - 結束回合前也有待選閘門，避免尚未決定 application 就先推進世界。
   - 驗證:進階建築學 → 自動化工廠/重型裝甲/行星飛彈基地(真資料),end-to-end 流程跑通。
   - AI 目前用預設第一項(decider 依性格選為後續小改)。

**目前狀態總結（2026-08-25 重新稽核）**：玩家可完成主題、選科技並解鎖元件；研究突破已
依原版改為嚴格超過成本後按超額比例擲骰，成功清零。研究 application 的玩家／AI 選擇時序、
Creative／Uncreative 分支亦已由 IDA 閉合並接線；其餘殖民地人口修正與特殊科技 callback 列於
[`../re/parity-matrix.tsv`](../re/parity-matrix.tsv)，不得再稱研究系統已忠實完成。

## 2026-08-09 種族研究差異接線

玩家與 AI 在研究開始前套用同一組種族規則：Creative 不進擇一畫面而於突破時取得全部應用；Uncreative
使用獨立且可存檔的研究亂數流預選一項。一般種族保留研究選擇 UI。`ResearchDraws`
會隨 session snapshot 保存，避免存讀檔後重新抽到另一項。

## 驗收

- 完成一主題後,只有**被選的**科技對應元件解鎖(其餘不解),對照原版行為。
- 既有測試(engine/research、techtree_verify、艦艇設計)全綠。
- 對原版實測:同一主題選不同科技,艦艇設計可用元件不同。

## 注意

- `session.go:866` 事件「科學突破 +150 RP」是隨機事件加成(合理),非主線成本,保留。
- 種族 Creative/Uncreative 特性影響「解全部/只解一項」已接入；研究亂數流的位置也隨存檔保存。
