# 間諜(Spying)系統:STEAL/SABOTAGE/HIDE 可玩迴圈

> 2026-08-11 更新。範圍:把 `internal/gamedata/spy.go`(手冊機率公式,先前零呼叫端死碼)接成
> 「訓練間諜 → 逐對手選 STEAL/SABOTAGE/HIDE → 每回合結算 → 回合摘要」的可玩迴圈。
> 原版三顆任務鈕左右語意仍尚未由反組譯確認，remake 使用明確標籤的三任務循環控制。

## 手冊出處

- `moo2_patch1.5/MANUAL_150.html`「Notes on Spying」段(p.113-115):`Spy Bonuses`/`Assassins`/
  `Roll Chance`/`Spy vs Spy` 四小節,已完整逐字移植進 `internal/gamedata/spy.go`(8 個函式,
  見該檔檔頭)。
- `moo2_patch1.5/GAME_MANUAL.pdf` p.174-175「Espionage」段(主手冊,定性描述任務效果):

  > On a mission, your Spy goes off into the colonies of another race (of your choice,
  > naturally). Undercover as one of them, the spy gathers information, **tries to steal
  > technologies you have yet to gain**, attempts to destroy some valuable piece of enemy
  > property or tries to remain hidden while still attempting to kill enemy agents.

  以及 p.49-50「Races」畫面說明(`Espionage`/`Sabotage`/`Hide` 三個任務選單按鈕的存在)。
- `openorion2/src/gamestate.h`:`spies[MAX_PLAYERS]`(`uint8_t`,每對手一個間諜位元狀態)、
  `SPY_MISSION_STEAL/SABOTAGE/HIDE`(`SPY_MISSION_MASK=0xc0`)、`spyMaintenance`——資料模型
  oracle,確認欄位存在但（依 `docs/tech/rules-implementation-audit.md` 第 7 節)**openorion2
  沒有任何任務結算邏輯**,`spies[]` 除讀檔外從未被賦值。
- 原版執行檔（輸入 SHA256
  `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`，LE 32-bit 位址空間）
  的 `N_Spies_Bonus` @ raw `0x1014A4` 已證實任務碼 `1` 使用破壞門檻 `70` 並呼叫
  `Steal_App` @ raw `0x10130A`；`Steal_App` 逐殖民地掃描建築旗標、以原版建築表
  `off_17EB3D` 的 `+8` 建造成本加權抽選，最後呼叫 `Add_Building` @ raw `0x145EA`。
  該函式對選定槽位寫入 `0`，因此「破壞建築」的直接效果已是已證實行為；`sub_101483`
  的三段式 slot helper、`sub_1014A4` 的 packed relationship byte／兩張 score table 讀取
  位置與門檻分支也已由 IDA Pro 9.4 追回。兩張 table 的上游填值、lucky roll 細節與
  特殊槽位政策仍分級保留。

## 硬門檻查核結論(建置前必答的問題)

任務要求「效果規則找得到才建」。查核結果:**偷科技(STEAL)與隱匿(HIDE)的可驗證部分有明確依據——偷一項
「對方已知、我方未知」的科技**,理由:手冊原文「tries to steal technologies you have yet to
gain」限定偷來的必須是攻擊方尚未擁有的;而間諜是潛入對方殖民地行動,邏輯上只能偷到對方已經
擁有的科技(不可能偷到雙方都沒有、還沒被任何人研究出來的東西)。這條規則不是臆測,是「手冊
限定詞(尚未擁有)+ 任務描述(潛入對方殖民地)」兩者交集出的必然結論。

HIDE 的手冊文字是「remain hidden while still attempting to kill enemy agents」，
`gamedata.SpyVsSpyAttackerBonus` 明確給選擇 HIDE 的攻擊方 +20。因此 remake 的 HIDE
不執行偷科技擲骰，只進 Spy vs Spy 判定；「不再同時偷科技」是依手冊三種任務以
`or` 分列所得的**強推論**。目前沒有獨立防守 Agent 數量欄位，若判定擊殺防守 Agent，
只記錄訊息，不把 AI 的進攻 Spy 誤扣成防守 Agent。

**破壞(SABOTAGE)的直接效果已由原版執行檔補證**:手冊只給「destroy some valuable piece
of enemy property」的定性描述，但原版 `0x10130A` 的讀寫端已具體顯示它在已建建築槽中選取
目標，並由 `0x145EA` 清除該槽旗標。remake 只把能由 `OrigBuildingID` 與建築表對回的正常
建築接入；按原版 `off_17EB3D` 49 筆建築表 `+8` 生產成本做穩定加權抽選。raw 槽
0..48、stride `0x13`、槽 9 skip 與完整成本快照見
[`internal/gamedata/original_building_table.go`](../../internal/gamedata/original_building_table.go)。
原版 `Random(total)` 的 1-based 累積選擇已映射為 remake 的 0-based 累積選擇；仍未知的是
特殊／未知槽的遊戲政策，不是正常成本表本身。

擲骰機率(成功/失敗判定)手冊有完整公式(`SpyEffectiveThreshold`/`SpyRollChance`),已在
`gamedata/spy.go` 移植且有單元測試覆蓋,直接複用。

## 2026-08-11 SABOTAGE 分數與 Agent runtime 收尾

`internal/shell/spy.go` 的 `spyMissionScore` 現在把一次任務的可用分數完整攤平：

- 攻方：`SpySlotBonus(Spies)`、攻方已知間諜科技、攻方種族／`Spy Master` 領袖 bonus。
- 守方：`SpySlotBonus(DefensiveAgents)`、守方已知間諜科技、政府、種族／`Telepath` 領袖 bonus。
- SABOTAGE 使用已證實的 `T=70`，再由 `E=T+DB-AB` 與 `SpyRollChance(E)` 得到成功率；建築
  命中後仍使用 49 槽／slot 9 skip／`+8` 建造成本加權清除。

這是原版兩張帝國攻防表加上逐對手 slot 的等價分層：`SpySlotBonus` 已呼叫與 raw
`sub_101483 @ 0x101483` 相同的 helper，且 `T=70`、Agent 消費與 49 槽建築權重已接；
不是宣稱原版兩張 score table 的上游填值與未命名 raw record 已逐欄還原。
`TrainDefensiveAgent`／`DismissDefensiveAgent` 使用 63 上限；Spy-vs-Spy 判定擊殺防守方時，
`advanceEspionage` 會消費一名 Agent，玩家與 AI 方向相同。訓練 30 BC 與 AI 週期免費補人
仍是 remake 拍板值；原版現已證實為殖民地產品 `-7`、100 工業、新人先進 self Agent pool，
AI 則每回合收回並重配既有人數。完整證據見 `docs/re/spy-turn-policy-audit-20260828.md`。

## 最小迴圈涵蓋範圍

### 有做

1. **訓練間諜**:`GameSession.TrainSpy(targetIdx int) bool`——玩家花 `spyTrainCostBC`(30
   BC,remake 拍板值,見下方「remake 拍板值」節)訓練一名間諜派駐到 `AIPlayers[targetIdx]`。
   `PlayerSpies []int`(平行 `AIPlayers`)記錄各對手的間諜數,opt-in、新對局預設全 0；
   `PlayerSpyMissions []SpyMission` 同步保存各對手的 STEAL/SABOTAGE/HIDE 任務,舊存檔缺欄位時退回
   STEAL。
2. **AI 自動訓練**:`advanceAI` 每 6 回合讓 AI 免費 +1 間諜(`AIOpponent.Spies`,上限 63),
   簡單週期政策,無 BC 成本模型(AI 經濟模型現行無法推導訓練成本,誠實簡化)。
3. **每回合結算(`advanceEspionage`,由 `EndTurn` 呼叫)**:
   - 間諜維護費:每個已訓練間諜每回合扣 `spyMaintenancePerSpyBC`(1 BC),opt-in(0 間諜
     時扣款為 0,不影響任何既有測試/經濟平衡)。
   - **STEAL／SABOTAGE 判定**:使用 `spyMissionScore` 明列 Spies／Agents slot、科技、
     政府、種族／領袖 bonus；`E = gamedata.SpyEffectiveThreshold(T, DB, AB)`、
     `p = gamedata.SpyRollChance(E)`,擲 `rand.Float64() < p` 判定成功。SABOTAGE 的 `T=70`。
   - **成功後偷科技**:`spyStealOptions(attacker, defender)` 列出 defender 已知
     (`CompletedTopics`)、attacker 未知的科技候選,隨機挑一項,`applyTechTheft` 套用到
     attacker 的 `CompletedTopics`/`ChosenTech`/`ExplicitChoice`(語意比照研究「明確抉擇」:
     只解鎖偷到的那一項,不會連帶解鎖同主題的其餘選項)。無候選則記錄「得手但無可偷」訊息,
     不誤改任何狀態。
   - **SABOTAGE 判定**:使用原版任務碼 `1` 已證實的門檻 `SpyThresholdSabotage=70`，成功後
     對防守方殖民地的已知建築候選依原版 `ProductionCost` 加權，移除一棟並寫入回合摘要；
     沒有可對回的建築時不改狀態。現行玩家 → AI 路徑已接入，AI → 玩家仍採原有 STEAL 預設。
   - **SpyVsSpy 判定**:`resolveSpyVsSpy(AB, DB, hide=false)` 用手冊給的 ±80 淨值門檻
     (`gamedata.SpyVsSpyDefenderKillThreshold`/`AttackerKillThreshold`)判定攻守雙方是否有
     一方被擊殺,擊殺攻方會讓對應的 `PlayerSpies[i]`/`AIOpponent.Spies` 遞減。
   - 雙向對稱:玩家 → AI、AI → 玩家各跑一次上述流程。
   - **HIDE 判定**:玩家將該 AI 的任務切為 HIDE 後，跳過 STEAL，改用
     `resolveSpyVsSpy(AB, DB, hide=true)`；介面以明確標籤循環三種任務。
   - 結果訊息記進 `GameSession.LastEspionage []string`(供回合摘要顯示,比照
     `LastEvent`／`LastAntaranNotice`／`LastBattle` 的既有慣例，下回合重算不存檔）。
4. **測試**(`internal/shell/spy_test.go`):SABOTAGE 候選的穩定排序／建造成本權重／無候選 no-op、
   成功移除建築與失敗保留 map；另有 `spyStealOptions` 找到/找不到可偷科技、
   `applyTechTheft` 只解鎖偷到的那一項不連帶解鎖同主題其餘選項、`resolveSpyVsSpy` 四種門檻
   情境(含 HIDE 加成邊界)、防禦方 bonus 升高會降低成功率(公式層級驗證)、`TrainSpy` 扣款
   /增加間諜數/BC 不足/越界索引、`spyStealAttempt` 用固定 rng 種子搜尋出的成功案例驗證偷竊
   確實套用、`advanceEspionage` 的維護費扣款精確等於間諜數 × 費率、0 間諜完全 no-op、多回合
   結算後間諜數不會變負。

### 已由 2026-08-28 RE 關閉、但 remake 尚待改造

1. **SABOTAGE 的原版 oracle 差異**:任務碼、70 門檻、建築旗標清除、49 筆 raw 建築成本表、
   槽 9 skip、建造成本權重、raw relationship byte 的低 6 位 count／高 2 位 mode、
   `sub_101483` slot helper 與 remake 可用資料模型的完整 AB/DB 分數已接入；原版兩張
   score table 的上游填值與主要亂數 consumer 均已閉合；特殊／未知建築槽的正式 UI 名稱不屬
   本段玩法 gate。
2. **原版三顆任務值已對位**:packed 值 0／1／2 為 Espionage／Sabotage／Hide，RACES local
   值為原值加一，未設定列預設 3（Hide）。資產上的實際左右圖像與正式文案仍是 UI 索引工作。
3. **防守方 Agent 已接；原版資料差異仍保留**:手冊區分 Spy(攻擊,逐對手指派)與 Agent(防守,
   不分對手、全體共用)兩種 slot,各自累計 `SpySlotBonus`。本 remake 已有訓練／解除、63 上限、
   AI 週期補充，且成功的 Spy-vs-Spy 擊殺會扣除 Agent；零 Agent 仍保留手冊所述的基本防禦。
4. **AI 防守側 remake 資料模型差異**:原版攻防 score、政體／領袖、Agent 留守與科技價值配置
   已追回；remake `RaceIndex`／Agent 結構仍未依該模型改造。
   手冊 Spy Bonuses 表列了種族特性(`SpyRaceTraitBonus`)、
   5 項科技(`SpyTechnologyBonus`:Neural Scanner/Telepathic Training/Cyber Security
   Link/Stealth Suit/Psionics)、政府型態(`SpyGovernmentDefenseBonus`,僅 Defense 欄)三項
   加成,`gamedata/spy.go` 都已備妥函式；玩家側已透過 `psKnowsTech` 逐一核對 5 項科技
   並接入 `spyMissionScore`；剩餘工作是 RE gate 後的規格與實作，不是原版資料留白。
5. **軍官(Telepath/Spy Master 技能)+ 暗殺(Assassins)**:手冊只給範圍(2~18、+2%~+18%),
   沒給技能等級 → 加成的精確映射公式,`gamedata/spy.go` 檔頭已標 TODO 保留範圍常數,未提供
   對應函式,故 remake 也未接。
6. **AI 科技模型的既有限制(非本輪引入)**:`advanceResearch()`(把已完成主題推進到下一個
   未完成主題)只接了玩家,AI 的 `ResearchTopic` 完成一次後不會自動往下推進(既有限制,見
   `internal/shell/session.go` `advanceResearch` 註解)。這代表 AI 長期而言只會完成 1~2 個
   研究主題,可偷科技的池子很小——是誠實反映 AI 科技模型目前的抽象程度,不是本輪新增的缺口,
   也不在本輪修正範圍內(修正 AI 研究推進屬於研究系統的既有 TODO,非間諜系統的責任)。
7. **已接**(gap-report 第 51 項(間諜UI))。而且規劃本身被一個負面發現改寫了——
   **原版根本沒有獨立的間諜畫面**(搜 `Spy_Screen`/`Espionage_Screen` 零命中),
   任務指派內嵌在「Races 種族關係」畫面裡,所以 remake 也接在那裡(`cmd/moo2/racesspy.go`)。
   gameplay 任務值已閉合；只剩原版資產上的左右圖像／正式字串要在 UI 索引中確認。

## SpyVsSpy 目前為何幾乎不會觸發(誠實說明,非 bug)

`resolveSpyVsSpy` 用 `SpyVsSpyAttackerBonus(AB, hide) - SpyVsSpyDefenderBonus(DB)` 的淨值
比較手冊給的 ±80 門檻。玩家攻擊側的 `AB` 已含 `SpySlotBonus`、玩家種族與科技加成；
`DB` 現在包含防守方 Agents slot、已接的科技／政府與種族／領袖 bonus；AI 仍缺原版
raw 政體／score record，未知部分保守傳 0。單元測試用構造出的 `spyMissionScore` 與
`ab`/`db` 數值驗證 SABOTAGE 門檻、Agent 消費、`resolveSpyVsSpy` 與 HIDE +20，
避免把 remake 近似分數誤當成原版完整 raw 模型。

## remake 現行近似（已知不同於原版）

- `spyTrainCostBC = 30`:手冊(GAME_MANUAL.pdf p.70「Ships & Spies」)只說間諜是透過殖民地
  建造佇列訓練出來的("Training a spy is unlike constructing a building or a ship, but it
  takes quite a lot of work..."),沒給具體成本數字,本 remake 也還沒有殖民地佇列的「間諜」
  建造選項。直接用 BC 簡化訓練流程,成本量級比照最低艦體(巡防艦 18 BC)抓一個 remake 拍板
  值。
- `spyMaintenancePerSpyBC = 1`：數值已由 `Compute_Player_Maintenance_ @ 0xE2000` 證實；
  remake 直接在 `advanceEspionage` 扣款的時序與報表歸屬仍不同於原版維護主鏈。
- AI 每 6 回合 +1 間諜：純 remake 週期政策；原版沒有此規則，而是把殖民地訓練所得的同一
  pool 每回合依 Agent 防守需求與可偷科技價值重新分配。

## 2026-08-11 AI SABOTAGE 接線勘誤

玩家對 AI 的 SABOTAGE 原本已能依 70 門檻與建造成本加權池清除建築，但 AI → 玩家
方向仍固定呼叫 STEAL，導致 `AIOpponent.ColonyBuildings` 只被動接受破壞、玩家建築
永遠不是目標。本輪已修正：

- `aiSpyMission` 依 remake 已有的 AI personality 選任務：`Xenophobic`／`Ruthless`／
  `Aggressive` 走 SABOTAGE，`Erratic` 偶數回合走 SABOTAGE，其餘走 STEAL。
- AI 的攻擊者仍使用自身 `Spies`，防守側仍使用玩家 `DefensiveAgents`、政府／種族／
  領袖 bonus；SABOTAGE 成功後把 `s.ColonyBuildings` 傳入同一個已測試的建築清除鏈。
- 這是**remake personality policy**，不是原版策略。原版已證實只有目標 personality raw 3
  時以 1/8 機率選 Sabotage，其餘選 Espionage，且不由 AI 選 Hide；raw 3 的正式名稱仍未知。

護欄：`TestAdvanceEspionageAISabotageUsesRuthlessPolicyAndPlayerBuildings` 以 200 個
可重播種子抽樣確認三種侵略性格確實能刪除玩家建築；既有玩家 → AI SABOTAGE 測試
仍保留，兩個方向不共用錯誤的建築 slice。

## 涉及檔案

- `internal/gamedata/spy.go`(既有,未改動):8 個機率公式,本輪沿用既有 `SpyThresholdSabotage`。
- `internal/shell/spy.go`:`TrainSpy`、`ensurePlayerSpies`、`psKnowsTech`、
  `spyStealOptions`、`applyTechTheft`、`spyAttackerBonus`/`spyDefenderBonus`、
  `resolveSpyVsSpy`、`spyStealAttempt`、`advanceEspionage`。
- `internal/shell/spy_test.go`(新增):單元測試,見上方「有做」第 4 點。
- `internal/shell/session.go`:`AIOpponent.Spies`、`GameSession.PlayerSpies`/
  `LastEspionage`/`spyRand` 欄位;`EndTurn` 呼叫 `advanceEspionage()`;`advanceAI` 加 AI
  自動訓練週期政策。
- `internal/shell/persist.go`:`PlayerSpies`/`AIOpponent.Spies` 納入存讀檔(`LastEspionage`
  比照其餘回合暫態不存檔)。
