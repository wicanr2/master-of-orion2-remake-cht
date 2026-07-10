# 間諜(Spying)系統:最小可玩迴圈

> 2026-07-11。範圍:把 `internal/gamedata/spy.go`(手冊機率公式,先前零呼叫端死碼)接成一個
> 「訓練間諜 → 每回合結算 → 偷科技」的最小可玩迴圈。**不是完整間諜畫面**——逐對手分配、
> Espionage/Sabotage/Hide 任務選單、破壞(SABOTAGE)效果均延後,詳見下方「涵蓋範圍」。

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

## 硬門檻查核結論(建置前必答的問題)

任務要求「效果規則找得到才建」。查核結果:**偷科技(STEAL)的效果規則有明確依據——偷一項
「對方已知、我方未知」的科技**,理由:手冊原文「tries to steal technologies you have yet to
gain」限定偷來的必須是攻擊方尚未擁有的;而間諜是潛入對方殖民地行動,邏輯上只能偷到對方已經
擁有的科技(不可能偷到雙方都沒有、還沒被任何人研究出來的東西)。這條規則不是臆測,是「手冊
限定詞(尚未擁有)+ 任務描述(潛入對方殖民地)」兩者交集出的必然結論。

**破壞(SABOTAGE)沒有明確效果規則**:手冊只給「destroy some valuable piece of enemy
property」這樣的定性描述,沒有講具體破壞對象(建築?艦艇?生產儲值?)、沒有數值規則、
`openorion2` 也無任何對應程式碼可查——故本輪**只做 STEAL,SABOTAGE 標 TODO 不建**。

擲骰機率(成功/失敗判定)手冊有完整公式(`SpyEffectiveThreshold`/`SpyRollChance`),已在
`gamedata/spy.go` 移植且有單元測試覆蓋,直接複用。

## 最小迴圈涵蓋範圍

### 有做

1. **訓練間諜**:`GameSession.TrainSpy(targetIdx int) bool`——玩家花 `spyTrainCostBC`(30
   BC,remake 拍板值,見下方「remake 拍板值」節)訓練一名間諜派駐到 `AIPlayers[targetIdx]`。
   `PlayerSpies []int`(平行 `AIPlayers`)記錄各對手的間諜數,opt-in、新對局預設全 0。
2. **AI 自動訓練**:`advanceAI` 每 6 回合讓 AI 免費 +1 間諜(`AIOpponent.Spies`,上限 63),
   簡單週期政策,無 BC 成本模型(AI 經濟模型現行無法推導訓練成本,誠實簡化)。
3. **每回合結算(`advanceEspionage`,由 `EndTurn` 呼叫)**:
   - 間諜維護費:每個已訓練間諜每回合扣 `spyMaintenancePerSpyBC`(1 BC),opt-in(0 間諜
     時扣款為 0,不影響任何既有測試/經濟平衡)。
   - **STEAL 判定**:`AB = gamedata.SpySlotBonus(間諜數)`、`DB = 0`(見下方「防禦方 Agent
     未追蹤」)、`E = gamedata.SpyEffectiveThreshold(SpyThresholdSteal, DB, AB)`、
     `p = gamedata.SpyRollChance(E)`,擲 `rand.Float64() < p` 判定成功。
   - **成功後偷科技**:`spyStealOptions(attacker, defender)` 列出 defender 已知
     (`CompletedTopics`)、attacker 未知的科技候選,隨機挑一項,`applyTechTheft` 套用到
     attacker 的 `CompletedTopics`/`ChosenTech`/`ExplicitChoice`(語意比照研究「明確抉擇」:
     只解鎖偷到的那一項,不會連帶解鎖同主題的其餘選項)。無候選則記錄「得手但無可偷」訊息,
     不誤改任何狀態。
   - **SpyVsSpy 判定**:`resolveSpyVsSpy(AB, DB, hide=false)` 用手冊給的 ±80 淨值門檻
     (`gamedata.SpyVsSpyDefenderKillThreshold`/`AttackerKillThreshold`)判定攻守雙方是否有
     一方被擊殺,擊殺攻方會讓對應的 `PlayerSpies[i]`/`AIOpponent.Spies` 遞減。
   - 雙向對稱:玩家 → AI、AI → 玩家各跑一次上述流程。
   - 結果訊息記進 `GameSession.LastEspionage []string`(供回合摘要顯示,比照
     `LastEvent`/`LastAntares`/`LastBattle` 的既有慣例,下回合重算不存檔)。
4. **測試**(`internal/shell/spy_test.go`):`spyStealOptions` 找到/找不到可偷科技、
   `applyTechTheft` 只解鎖偷到的那一項不連帶解鎖同主題其餘選項、`resolveSpyVsSpy` 四種門檻
   情境(含 HIDE 加成邊界)、防禦方 bonus 升高會降低成功率(公式層級驗證)、`TrainSpy` 扣款
   /增加間諜數/BC 不足/越界索引、`spyStealAttempt` 用固定 rng 種子搜尋出的成功案例驗證偷竊
   確實套用、`advanceEspionage` 的維護費扣款精確等於間諜數 × 費率、0 間諜完全 no-op、多回合
   結算後間諜數不會變負。

### 延後(TODO,誠實標記,非遺漏)

1. **破壞(SABOTAGE)**:手冊無明確效果規則,見上方硬門檻查核結論,不臆測不做。
2. **逐對手分配 + 任務選單(Espionage/Sabotage/Hide)**:`PlayerSpies []int` 資料結構本身已
   支援逐對手分配(平行 `AIPlayers`),只是目前唯一一個 AI 對手時看不出差異;而「同一批間諜
   可以選擇 Espionage(蒐集情報)/Sabotage/Hide 三種任務」這個任務選單完全沒做——最小迴圈
   預設所有已訓練的間諜恆執行 STEAL。
3. **防禦方 Agent 未獨立追蹤**:手冊區分 Spy(攻擊,逐對手指派)與 Agent(防守,不分對手、
   全體共用)兩種 slot,各自累計 `SpySlotBonus`。本 remake 完全沒有 Agent 訓練/數量追蹤,
   `spyDefenderBonus()` 固定回 0——這正好對應手冊原文「defenses against enemy spies are
   active...even with zero defending agents」描述的「零 Agent」情境,不是遺漏,是誠實反映
   目前的簡化狀態。
4. **種族/科技/政府對間諜的加成**:手冊 Spy Bonuses 表列了種族特性(`SpyRaceTraitBonus`)、
   5 項科技(`SpyTechnologyBonus`:Neural Scanner/Telepathic Training/Cyber Security
   Link/Stealth Suit/Psionics)、政府型態(`SpyGovernmentDefenseBonus`,僅 Defense 欄)三項
   加成,`gamedata/spy.go` 都已備妥函式——但 `AIOpponent` 沒有種族/政府型態欄位、remake 也
   沒有「逐科技檢查是否擁有」的簡便介面(需要走 `psKnowsTech` 逐一核對 5 項科技,可行但本輪
   未接,累加到 `spyAttackerBonus`/`spyDefenderBonus` 即可),故 `spyAttackerBonus`/
   `spyDefenderBonus` 目前只接 `SpySlotBonus`,其餘三項一律 0。**接上後 SpyVsSpy 的 ±80
   門檻才有機會在正常遊戲流程中被觸發**(見下一點)。
5. **軍官(Telepath/Spy Master 技能)+ 暗殺(Assassins)**:手冊只給範圍(2~18、+2%~+18%),
   沒給技能等級 → 加成的精確映射公式,`gamedata/spy.go` 檔頭已標 TODO 保留範圍常數,未提供
   對應函式,故 remake 也未接。
6. **AI 科技模型的既有限制(非本輪引入)**:`advanceResearch()`(把已完成主題推進到下一個
   未完成主題)只接了玩家,AI 的 `ResearchTopic` 完成一次後不會自動往下推進(既有限制,見
   `internal/shell/session.go` `advanceResearch` 註解)。這代表 AI 長期而言只會完成 1~2 個
   研究主題,可偷科技的池子很小——是誠實反映 AI 科技模型目前的抽象程度,不是本輪新增的缺口,
   也不在本輪修正範圍內(修正 AI 研究推進屬於研究系統的既有 TODO,非間諜系統的責任)。
7. **UI**:只做了引擎層 `TrainSpy` 函式,`interactive.go` 未接對應畫面/按鈕(比照本專案其他
   「引擎層先行、UI 延後」的既有慣例,如軌道轟炸)。

## SpyVsSpy 目前為何幾乎不會觸發(誠實說明,非 bug)

`resolveSpyVsSpy` 用 `SpyVsSpyAttackerBonus(AB, hide) - SpyVsSpyDefenderBonus(DB)` 的淨值
比較手冊給的 ±80 門檻。現行 `AB` 只含 `SpySlotBonus`(間諜數上限 63 人 → 加成上限 41),
`DB` 固定 0(見上方「防禦方 Agent 未追蹤」),`SpyVsSpyDefenderBonus(0) = 20` 是基準值。即使
`AB` 拉滿 41,`net = 41 - 20 = 21`,遠不到 ±80 門檻——**透過目前正常遊戲流程幾乎不可能觸發
擊殺**,這是「輔助加成(種族/科技/政府)未接上」的誠實結果,不是判定邏輯有誤(單元測試用構造
出的 `ab`/`db` 數值直接驗證 `resolveSpyVsSpy` 本身邏輯正確,見 `spy_test.go`)。待接上上述第 4
點的三項加成後,門檻才有機會在真實對局中被跨過。

## remake 拍板值(非手冊精確數字,誠實標註)

- `spyTrainCostBC = 30`:手冊(GAME_MANUAL.pdf p.70「Ships & Spies」)只說間諜是透過殖民地
  建造佇列訓練出來的("Training a spy is unlike constructing a building or a ship, but it
  takes quite a lot of work..."),沒給具體成本數字,本 remake 也還沒有殖民地佇列的「間諜」
  建造選項。直接用 BC 簡化訓練流程,成本量級比照最低艦體(巡防艦 18 BC)抓一個 remake 拍板
  值。
- `spyMaintenancePerSpyBC = 1`:`engine.PlayerState.Maintenance` 欄位註解已載明「間諜維護費
  本專案尚無可推導模型」,這裡給的是 remake 佔位值,刻意不併入既有 `Maintenance` 欄位(避免
  牽動既有經濟測試的既定假設),改在 `advanceEspionage` 直接從 BC 扣。
- AI 每 6 回合 +1 間諜:純週期政策,無手冊依據,比照 `aiExpand` 每 5 回合擴張一顆星的既有
  節奏抓一個相近數字。

## 涉及檔案

- `internal/gamedata/spy.go`(既有,未改動):8 個機率公式,本輪只是首次被呼叫。
- `internal/shell/spy.go`(新增):`TrainSpy`、`ensurePlayerSpies`、`psKnowsTech`、
  `spyStealOptions`、`applyTechTheft`、`spyAttackerBonus`/`spyDefenderBonus`、
  `resolveSpyVsSpy`、`spyStealAttempt`、`advanceEspionage`。
- `internal/shell/spy_test.go`(新增):單元測試,見上方「有做」第 4 點。
- `internal/shell/session.go`:`AIOpponent.Spies`、`GameSession.PlayerSpies`/
  `LastEspionage`/`spyRand` 欄位;`EndTurn` 呼叫 `advanceEspionage()`;`advanceAI` 加 AI
  自動訓練週期政策。
- `internal/shell/persist.go`:`PlayerSpies`/`AIOpponent.Spies` 納入存讀檔(`LastEspionage`
  比照其餘回合暫態不存檔)。
