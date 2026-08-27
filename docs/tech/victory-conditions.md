# 勝利條件(Victory Conditions)

> 日期:2026-07-11(當日兩輪更新)。目的:記錄「銀河霸主2 怎麼贏一局」的手冊權威規則、remake
> 現況、以及尚未解決的資料模型限制,供後續接手不必重查手冊。**本專案第一次接上任何可達成的勝利
> 路徑**——先前 `docs/HONEST-STATUS.md` 點名的核心痛點「遊戲目前沒有任何勝利條件、無法贏一局」,
> 從這輪起有解。**第二輪更新(同日稍後)**:手冊三條勝利路徑已**全部接線(3/3)**——第一輪只接了
> 征服 + 議會兩條,安塔蘭母星反攻(第三條)當時仍是 TODO(見舊版第 6 節);本輪補上次元傳送門
> 解鎖反攻 + `AssaultAntares` 戰鬥 + 勝利偵測,詳見第 4 節。

## 1. 手冊權威規則(逐字引用 + 頁碼)

來源:`moo2_patch1.5/GAME_MANUAL.pdf`,用 `pdftotext -layout` 擷取(容器內無 pdftotext,host 有,
見記憶 `moo2-patch15-manual-text-extractable`)。第 12 章「The End of the Game」→「Winning」,manual
p.183(頁腳頁碼,`pdftotext` 輸出可見)。

> "The last and possibly most complicated method is to win an election of the Galactic
> Council. When half of the galaxy has been settled, the threat of war over competition for
> the habitable planets becomes too great. If there are 3 or more extant races, they gather
> and form the Galactic Council to prevent future war. The Council's only order of business is
> to select a leader to rule the entire galaxy. Based on the size of the population of each
> empire, the leader of every race is assigned a number of votes. Two contenders are chosen
> — those whose empires wield the most votes. How each race votes is determined on the
> basis of current diplomatic relations. If one of the nominees receives a full two-thirds
> majority of the votes, that leader becomes ruler of the galaxy and the game is over. Clearly,
> your intention is to prevent others from being elected until you can yourself be elected to
> hold sway over all of known space. Of course, there's no way the council can force you to
> accept a decision you don't agree with."

同一頁(p.182-183)另外兩條路徑:

> "Obviously, if yours is the only surviving race, as its emperor you rule the galaxy. Thus, you
> could win by conquering or destroying every colony of every other race — perhaps
> accepting an abject surrender or two."
>
> "An alternate method is to seek out and defeat the Antaran home fleet. This involves
> travelling to the Antaran homeworld, which is not possible until you have the right
> technology and build a Dimensional Gate. Once you defeat the awe-inspiring Antarans, all
> the other races in the galaxy recognise your overwhelming superiority and quickly
> capitulate. (This strategy is not available if you disabled Antaran Attacks when setting up
> your game.)"

`moo2_patch1.5/MANUAL_150.html`(1.50 patch notes,同一份手冊內容的另一份文件,獨立含「Win
Conditions」摘要,**與上述無版本差異**)額外補了計分公式一節,含「Council Win」的計分獎勵:

> "Council Win — Brings in a meager 100 points. The value can be changed with hi_score
> council."

### 手冊給出的硬數字 vs 只有定性描述

| 規則 | 手冊怎麼說 | 是否給精確數字 |
|---|---|---|
| 議會成立門檻:銀河殖民率 | 「When half of the galaxy has been settled」 | ✅ 明確(1/2) |
| 議會成立門檻:存續種族數 | 「If there are 3 or more extant races」 | ✅ 明確(≥3) |
| 勝出門檻 | 「a full two-thirds majority of the votes」 | ✅ 明確(2/3) |
| 人口 → 票數換算 | 「Based on the size of the population... assigned a number of votes」 | ❌ 只有定性描述,無換算係數 |
| 候選人怎麼選、第三方怎麼投票 | 「Two contenders are chosen... How each race votes is determined on the basis of current diplomatic relations」 | ❌ 只有定性描述,無公式 |
| 重開間隔(第幾屆之後多久再開) | 手冊未提;外交台詞(`assets/i18n/diplo.json`)證實會反覆召開 | ❌ 完全沒有數字 |
| Council Win 計分獎勵 | 「100 points」(MANUAL_150.html) | ✅ 明確,但本 remake 無計分系統可接 |

## 2. openorion2 沒有可抄的邏輯(這是從零設計)

`docs/tech/rules-implementation-audit.md` 第 10 項(2026-07-03 盤點)記載「openorion2 對
`victory|winner|win_condition|gameOver` 全 repo(C++ 參考專案本身)零命中」——這條記錄分析的是
**openorion2**,不是本 remake,至今仍然成立、沒有過期。也就是說,勝利條件這整個機制在
openorion2 裡確實連影子都沒有,只能依手冊從零設計,沒有既有 C++ 邏輯可對照。

## 3. 本 remake 現況(2026-07-11 這輪之前):已有純函式,但是死碼

`internal/engine/victory.go`(commit `2cccf18`,**2026-07-03 14:19**,比上面提到的盤點文件晚幾小時)
其實已經存在:

- `VictoryCondition` 列舉(`VictoryNone`/`VictoryExtermination`/`VictoryHighCouncil`/`VictoryAntaran`)
- `CheckExtermination(alive []bool) (bool, int)`——通用 N 人「只剩一位存活」判定
- `CheckHighCouncil(votesFor, totalVotes int) bool`——`votesFor*3 >= totalVotes*2`,整數運算避免浮點誤差
- `CheckAntaranVictory(antaranHomeworldConquered bool) bool`——接收布林旗標,不含母星戰鬥流程
- `CheckVictory(...)`——依滅絕 → 安塔蘭 → 議會的優先序整合三者

這組函式本身正確、有測試(`internal/engine/victory_test.go`),但**在這輪之前從未被
`internal/shell` 或 `cmd/moo2` 任何地方呼叫**——是一組沒接進實際回合流程的死碼,玩家永遠不會遇到它。
`CheckHighCouncil` 自己也誠實註記:「本函式的 votesFor/totalVotes 一律由呼叫端算好傳入」,把「人口
怎麼變成票數」「議會什麼時候該存在」這兩塊留白給呼叫端——這輪之前沒有呼叫端,所以這兩塊也一直是空的。

同時 `internal/shell/session.go` 另外還有一組**完全獨立、更粗糙**的 `CouncilVote`/`VoteResult`
(票數=人口、較高者當選,無成立門檻、無 2/3 多數、未接遊戲結束),只給 `cmd/moo2/interactive.go`
的議會畫面顯示用,是典型的「自編近似當真」——這輪已移除,見下。

## 4. 這輪(2026-07-11)接上的東西

### 4.1 gamedata 層(`internal/gamedata/council.go`)

補 `engine.CheckHighCouncil` 留白的兩塊,純函式、有測試(`council_test.go`):

- `CouncilEligible(settledStars, totalStars, extantRaces int) bool`——議會成立判定,字面對應手冊
  「半數殖民」+「≥3 存續種族」兩條件。`CouncilMinExtantRaces = 3`(手冊字面值,保留給未來多 AI 對手
  擴充時直接還原)。
- `CouncilVotes(population int) int`——人口→票數，採原版已證實的 **除以 10 後向上取整**；
  `sub_15B90 @ 0x15B90` 的原始指令與唯一議會 caller／consumer 見
  [`docs/re/parity-re-audit-20260812.md`](../re/parity-re-audit-20260812.md#議會票數)。
  `population<=0` 回傳 0 是 remake 的輸入安全政策。
- `CouncilWinScoreBonus = 100` 已接入最終分數；八項原版係數與剩餘 Picks 倍率的 IDA 證據見
  `docs/re/history-score-audit-20260825.md`。
- **2/3 超級多數門檻不重複實作**,直接沿用 `engine.CheckHighCouncil`;殲滅勝利同理沿用
  `engine.CheckExtermination`——避免兩套等價邏輯並存。

### 4.2 shell 層整合(`internal/shell/council.go`)

- `GameSession` 保存 `Victory VictoryState`、`PendingCouncilElection *CouncilElection`、
  `LastCouncilNotice *CouncilNotice`、`CouncilMeetings int`、`lastCouncilTurn int`。勝負、待決選舉、
  屆次與上次開會回合等持久狀態已同步存讀檔（見 `internal/shell/persist.go`）；
  `LastCouncilNotice` 與舊成品通知相同，只是當回合的顯示暫態，不寫入存檔。
- `advanceCouncil()`:`EndTurn` 每回合呼叫的狀態機。議會成立(`councilEligible`)且相對回合至少
  25 才能首次召開，後續每隔 25 回合再開；這組排程來自 `Check_For_Council_Meeting_ @ 0x168AF`
  raw 指令，而非手冊猜值。**逐帝國**(玩家 + 每個 AI 對手各自
  獨立,2026-07-11 由「玩家 vs 單一 AI 二元計票」generalize 為 N 帝國)算 `gamedata.CouncilVotes
  (該帝國殖民地人口加總)`,2/3 門檻用全體(玩家+所有 AI)總票數,依 `engine.CheckHighCouncil` 逐一
  判定:
  - 玩家達 2/3 → 立即勝利(`Victory.Reason = engine.VictoryHighCouncil`,不需要「接受」這一步——
    手冊那句「議會無法強迫你接受」只適用於「當選者不是你」的情境)。
  - 某個 AI 達 2/3 → 記錄 `PendingCouncilElection`(`EnemyName` = 該 AI 名稱,不是寫死的
    `AIPlayers[0]`),等玩家用 `RespondToCouncilElection(accept bool)` 回應:`accept=true` 結束遊戲
    判負,`accept=false` 不結束、下一屆再開(手冊原句直接翻譯成這個互動)。
  - 沒有任何一方達標 → 流會；`LastCouncilNotice` 以 typed 候選索引與票數記錄本屆結果，
    下一屆再開。固定雙語句型由 `assets/i18n/ui.json` 提供，不存於規則層。
- `advanceConquestVictory()`:殲滅所有對手,沿用 `engine.CheckExtermination`(對稱判定,理論上也涵蓋
  「玩家 0 殖民地、AI 存活」→ AI 勝利的方向,但本 remake 目前沒有任何機制會讓玩家殖民地清零,這個
  分支現況不可達,只是沿用同一個對稱函式的自然結果)。`InvadeColony` 攻陷 AI 唯一殖民地後立即呼叫一次
  (不用等下個 `EndTurn`),`EndTurn` 本身也每回合呼叫一次(防禦性)。
- `CouncilStatus()`:唯讀快照(是否成立/目前票數/待決/勝負),供 UI 讀取,不重算任何規則。
- `VictoryReasonLabel(engine.VictoryCondition) string`:中文化標籤(`engine` 是純規則層,不放
  UI 字串)。

### 4.3 UI(`cmd/moo2/interactive.go` 的 `council()` 場景)

**刻意不重建原版議會投票畫面**(座位圖/候選人肖像/動畫)。只把 `CouncilStatus()` 的結果誠實印成
文字:尚未成立 / 已成立待開 / 已分出勝負 / 待玩家回應。**沒有互動式 accept/reject 按鈕**——玩家
目前只能用 `GameSession.RespondToCouncilElection(bool)`(尚無 UI 熱區綁定這個呼叫)。這是本輪
「UI 最小化、延後原版重建」的刻意選擇,已在 `docs/HONEST-STATUS.md` 誠實標注。

### 4.4 安塔蘭母星反攻整合(2026-07-11 第二輪,`internal/shell/antaran_victory.go`)

手冊第二條勝利路徑(GAME_MANUAL.pdf p.183):

> "An alternate method is to seek out and defeat the Antaran home fleet. This involves
> travelling to the Antaran homeworld, which is not possible until you have the right
> technology and build a Dimensional Gate. Once you defeat the awe-inspiring Antarans, all
> the other races in the galaxy recognise your overwhelming superiority and quickly
> capitulate. (This strategy is not available if you disabled Antaran Attacks when setting up
> your game.)"

與次元傳送門本身效果(GAME_MANUAL.pdf p.106,「Multi-Dimensional Physics」小節):

> "A Dimensional Portal gives your fleets in the same system the ability to cross into the
> dimension from which the Antarans stage their attacks. To use this, select a fleet in the
> same system as the portal, then click the Attack Antarans button instead of selecting a
> destination."
>
> "A Dimensional Portal costs 2 BC in maintenance each turn."

**次元傳送門建築本身不是本輪新建**——`gamedata.Buildings`(`internal/gamedata/buildings.go`)
早在先前的建築全表萃取(`docs/tech/colony-buildings.md`)就已收錄(`BUILDING_DIMENSIONAL_PORTAL`,
`ColonyBuilding=14`;前置科技 `TOPIC_MULTIDIMENSIONAL_PHYSICS`,RP 成本 4500,維護 2 BC/回合,
與手冊逐字相符),只是`docs/tech/colony-buildings.md` §6.2 先前把它列在「記錄已建但不影響任何數值」
——建成後沒有任何後續流程,是**沒接線的死碼建築**。本輪要做的是「建成後解鎖反攻」這段流程,不是
重新定義建築資料。

**前置科技怎麼判定**:remake 的建造選單 gate(`availableBuildOptions`,見 `session.go`)是「主題
完成」層級(`s.Player.CompletedTopics[b.PrereqTopic]`),不看玩家在該主題底下具體選了哪個科技
(`TOPIC_MULTIDIMENSIONAL_PHYSICS` 底下有 `TECH_DIMENSIONAL_PORTAL`/`TECH_DISRUPTER_CANNON` 兩個
選項)——這是既有、其餘 39 項建築/衛星共用的既有慣例,本輪未新增或修改判定邏輯,只是確認
Dimensional Portal 沿用同一套既有機制即可正確 gate,不需要額外的「玩家是否選了 Dimensional Portal
這個科技」檢查。

**新增的流程**(`internal/shell/antaran_victory.go`):

- `hasDimensionalPortal()`:掃描 `s.ColonyBuildings`(各殖民地已完工建築去重 map)是否有任一筆記
  「次元傳送門」。remake 簡化:不要求「艦隊與傳送門同星系」(手冊原文有此限制,但 remake 的星際
  航行模型不追蹤「建築在哪個星系」與「艦隊目前在哪個星系」的交叉可達性),只要求帝國內任一殖民地
  已建成即視為前置滿足。
- `CanAssaultAntares()`(匯出):遊戲未結束 + `!DisableEvents`(手冊:關閉安塔蘭攻擊則本路徑不可用)
  + 已建傳送門 + 艦隊非空,四條件皆滿足才允許反攻。UI 用它決定是否顯示按鈕,`AssaultAntares` 內部
  也呼叫它做同一份判斷(避免兩處邏輯分岔)。
- `AssaultAntares()`:前置條件不滿足 → `ok=false`,不消耗艦隊、不觸發戰鬥。滿足 → 解算戰鬥:
  沿用 `ResolveBattle` 同款 `battleVolley` 逐回合齊射解算(最多 6 回合),防禦方使用
  `antaranHomeFleetDefense` 的反組譯艦級組成與已知戰機艙貢獻；艦級階梯與要塞仍是 remake 代理。
  **與 `ResolveBattle` 不同**:`ResolveBattle` 的
  `PlayerWon` 只要求「艦數比敵方多或敵方全滅」,`AssaultAntares` 要求防禦方**全滅**才算
  `PlayerWon`——手冊「Once you defeat the awe-inspiring Antarans」語意是徹底擊敗,不是打退,終局
  一戰用更嚴格的判定合理。戰勝 → `s.AntaranHomeworldConquered=true`;戰敗 → 套用艦隊損失
  (`removeWeakestShip`,比照 `ResolveBattle`),不設勝利旗標。
- `advanceAntaranVictory()`:`EndTurn` 每回合呼叫,偵測 `AntaranHomeworldConquered` 並沿用
  `engine.CheckAntaranVictory` 純函式判定(不重算邏輯),命中則設定
  `s.Victory={Over:true, Reason:engine.VictoryAntaran, Winner:"player"}`。呼叫順序排在
  `advanceConquestVictory`(殲滅)之後、`advanceCouncil`(議會)之前——與 `engine.CheckVictory`
  文件記載的「滅絕 → 安塔蘭 → 議會」優先序一致(見 `internal/engine/victory.go` 該函式註解)。

> **2026-08-09 勘誤（本段舊值保留作歷史快照）**：`docs/re/antaran-defense-fleet.md` 已從
> `Orion2.exe` 反組譯解出防禦上限 `{0,0,3,2,7,0,0,0,0}`，即 `Intruder ×3`、`Interdictor ×2`、
> `Harbinger ×7`，另加一座星際要塞；標準艦的非空 weapon ID／數量／raw flags 也已建立索引。
> 因此下面「6 艘末日之星」不再是當前實作或當前研究結論；目前仍未知的是 remake 艦級映射、
> 要塞完整設計、難度裝甲的完整對應與敵方即時武器傷害消費。

**⚠ 歷史快照（已由上方勘誤取代；母星防禦艦隊戰力,手冊/openorion2 當時未找到精確數字）**:`GAME_MANUAL.pdf`「Winning」
小節全文搜尋「Antaran」的 60 餘處出現,只有「the awe-inspiring Antarans」這句定性描述,沒有任何
具體的母星防禦艦隊組成或戰力數字(第 1 節手冊逐字引用已完整收錄相關段落,沒有遺漏)。
`openorion2`(`docs/tech/rules-implementation-audit.md` 第 10 項已記載)對 victory/winner 相關邏輯
全 repo 零命中,自然也沒有母星防禦艦隊的資料可抄——這是從零設計的部分,不是查漏。

歷史保守預設(`antaranHomeFleetDefense`,`internal/shell/antaran_victory.go`):**6 艘「末日之星」等級
戰力(`shipStrength("末日之星")==64`,MOO2 六級艦體中最高等級)**,合計戰力 384。理由:確保玩家
不能用隨手一支小艦隊反攻(呼應「awe-inspiring」的定性描述),但仍是「打得贏」的固定值,不是無限強
的裝飾性數字——玩家投入同等量級的末日之星艦隊(測試驗證 8 艘可穩定取勝)即可一戰。**這是 remake
保守預設,非手冊或 openorion2 給出的精確值**,若之後找到權威來源(如 openorion2 之外的其他反編/
社群逆向資料),應更新此常數並移除本段免責聲明。

**最小 UI**(`cmd/moo2/interactive.go` 的 `fleet()` 場景):在艦隊列表畫面左下空白區加一個
「攻打安塔蘭母星」文字提示 + 熱區,只在 `CanAssaultAntares()` 為真時顯示。
⚠ **2026-08-07 更新**:點擊改為**進入安塔蘭王座廳**(原版 `Main_Antaran_Room`,
`cmd/moo2/antaranroom.go`,用原版 `antaroom.LBX` 美術),由該畫面確認後才呼叫
`AssaultAntares()` 後導向既有的 `battleResult()` 戰鬥結果畫面(該畫面讀 `s.LastBattle`,
`AssaultAntares` 已寫入,直接複用,不需要新的結果畫面)。**刻意不做**:原版議會選舉那種
「灰階按鈕」美術(手冊沒有描述具體 UI 佈局可依循)、勝利/落敗專屬結束畫面(與議會選舉同款限制,
見 4.3 節同款「UI 最小化」決策)。

## 5. 資料模型限制(重要,誠實標注)

**2026-08-11 更新:`NewDemoSession` 維持 3 個 AI 對手，並已補上 AI↔AI 關係矩陣**(多帝國競爭骨架與
可選 AI-to-AI 戰爭／外交見 `docs/tech/ai-to-ai.md`),場上存續帝國數上限變成「玩家 + 3 AI」= 4。
成立門檻、候選人、第三方投票與玩家三選一現在都已接入；未映射外交分數仍是 oracle 差異：

1. **成立門檻「≥3 存續種族」現在真的可達成。** `councilMinExtantRacesOverride`(先前的 shell 層
   資料模型限制近似覆寫值,固定為 2)**已移除**,`councilEligible` 直接引用手冊字面值常數
   `gamedata.CouncilMinExtantRaces`(=3)——玩家 + 3 個 AI 對手共 4 個帝國,只要其中至少 3 個仍存續
   (各自至少 1 個殖民地)就滿足門檻,不再需要 remake 近似值。
2. **候選人／搖擺票流程已由 IDA 訂正。** `tallyCouncil` 取基礎票前兩名後，非候選 AI 對兩人
   各做一次獨立 1..200 檢定；只有單邊通過才投該候選人，雙邊同時通過或同時失敗均棄權。
   玩家由 `RespondToCouncilVote` 明確選候選人或棄權。固定 `councilSwingVoteMinRelation = 8`
   已移除；可映射分數與剩餘 raw 欄位見
   [`docs/re/council-voting-audit-20260824.md`](../re/council-voting-audit-20260824.md)。

議會排程的原始位址、bytes、日曆換算、旗標與奇數星邊界見
[`docs/re/council-schedule-audit-20260824.md`](../re/council-schedule-audit-20260824.md)。

## 6. TODO(誠實列出,不硬做)

> **Antares 母星次元傳送門勝利(手冊第二條路徑)已於 2026-07-11 第二輪接線**,原本列在這裡的
> TODO 已移除(舊敘述:「完全沒有對應流程——無 Dimensional Portal 科技/建造、無『派遣艦隊前往
> Antares』的星際航行目的地、無母星戰鬥」),避免留著過期斷言佔位(rule 63)。詳見第 4.4 節。
> 仍未做的子項目(不阻塞「能不能贏這條路徑」,是精修/深化):
> - **「艦隊與傳送門同星系」的精確前置**:remake 簡化為「帝國內任一殖民地已建成即滿足」,見 4.4
>   節誠實聲明,需要 remake 的星際航行模型先支援「建築所在星系 ↔ 艦隊所在星系」的可達性比對。
> - **母星防禦艦隊的完整戰鬥對齊**:艦隊上限、艦級名稱、標準艦非空武器槽與 ID 31 戰機艙已由
>   `docs/re/antaran-defense-fleet.md` 證實；仍缺 remake 艦級映射、要塞完整設計、難度裝甲完整表、
>   敵方即時武器／戰機命中傷害與敵方出擊政策。
> - **安塔蘭母星本身的星圖呈現**:remake 把反攻建模成「一場戰鬥」而非「星圖上一個可航行的目的地
>   + 母星星球」,手冊原文暗示是實際的星際航行(travelling to the Antaran homeworld)。
> - **歐瑞恩守護者(Orion Guardian)**:與安塔蘭母星是手冊裡兩個不同的終局戰(Score 小節分別提到
>   「defeats the Guardian and captures Orion」與「defeats the Antarans at Antares」兩種點數獎勵),
>   本輪只接安塔蘭母星這一條,歐瑞恩守護者仍完全沒有對應流程。
- **計分系統(Score Calculation)**:八項得分、議會／安塔蘭獎勵與種族倍率已有玩家可見 Hi-Score
  消費端；INFO 歷史圖目前仍待改成原版四項 350 格正規化 ring。`CouncilWinScoreBonus`
  只是預先記錄的權威值。
- **議會選舉結束畫面 + accept/reject 互動 UI**:目前只有文字狀態,沒有原版議會 3D 場景的投票動畫、
  沒有結束畫面(勝利/落敗的專屬畫面),`RespondToCouncilElection` 也還沒有 UI 熱區可以觸發。
- **候選人限定「票數最高兩位」+ 第三方外交搖擺票**:remake 已接入 `tallyCouncil` 與 AI↔AI
  `AIRelations`；剩下的是原版精確外交門檻與特殊表 oracle，不是 remake 漏接。
- **AI 對 AI 的戰爭／外交互動**:可選 `EnableAIVsAI` 已接入正式政策矩陣、貿易／研究資源交換與
  抽象艦隊攻擊／殖民地轉移；戰鬥仍沒有原版逐艦 blueprint，因此只稱 remake 抽象模型，詳見
  `docs/tech/ai-to-ai.md`。

## 7. 測試

- `internal/gamedata/council_test.go`:`CouncilEligible`(門檻邊界)、`CouncilVotes`(含負值/零值)。
- `internal/shell/council_test.go`:議會未成立不開會、成立後立即開第一屆、玩家達標直接勝利、AI 達標
  待回應(拒絕不結束+下一屆再開、接受才結束)、五五波流會、`DisableEvents` 關閉議會、殲滅勝利判定
  (含「對手仍存活不誤判」;2026-07-11 訂正為「全部 AI 對手殖民地清空」才觸發,見下方 multi_ai_test.go
  的「僅消滅部分 AI 不誤判」新案例)。
- `internal/shell/multi_ai_test.go`(2026-07-11 新增):`NewDemoSession` 建 3 個 AI 對手(母星不
  重疊、名稱/性格互異、`PlayerSpies` 平行陣列同步)、3 個 AI 各自獨立造艦/擴張、議會 N 帝國計票
  (非 `AIPlayers[0]` 當選也能正確判定、多 AI 人口分散時不誤判)、`gamedata.CouncilMinExtantRaces`
  真門檻(只剩 2 個存續帝國時議會不成立)、剛好達門檻(3 帝國)時 N 帝國計票邏輯仍正確、部分消滅
  AI 對手不誤判殲滅勝利。
- `internal/engine/victory_test.go`(既有,2026-07-03):`CheckExtermination`/`CheckHighCouncil`/
  `CheckAntaranVictory`/`CheckVictory` 純函式門檻邊界。
- `internal/shell/antaran_victory_test.go`(2026-07-11 第二輪新增；當時戰力說明為歷史快照):`CanAssaultAntares` 前置條件
  (無傳送門/艦隊為空/`DisableEvents` 皆擋下)、`AssaultAntares` 前置不滿足時不消耗艦隊不觸發戰鬥、
  弱艦隊戰敗不誤判勝利(且正確記錄 `LastBattle`)、強艦隊戰勝反組譯組成的防禦代理後
  `AntaranHomeworldConquered=true`,`advanceAntaranVictory` 隨後正確設定
  `Victory={Over:true, Reason:VictoryAntaran, Winner:"player"}`、殲滅與安塔蘭同時達成時
  `EndTurn` 呼叫順序確保殲滅優先(對齊 `engine.CheckVictory` 文件記載的優先序)。
