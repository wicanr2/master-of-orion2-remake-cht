# 勝利條件(Victory Conditions)

> 日期:2026-07-11。目的:記錄「銀河霸主2 怎麼贏一局」的手冊權威規則、remake 現況、以及尚未解決的
> 資料模型限制,供後續接手不必重查手冊。**這是本專案第一次接上任何可達成的勝利路徑**——先前
> `docs/HONEST-STATUS.md` 點名的核心痛點「遊戲目前沒有任何勝利條件、無法贏一局」,本輪起有解。

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
| 重開間隔(第幾屆之後多久再開) | 手冊未提;外交台詞(`assets/i18n/diplo.tsv`)證實會反覆召開 | ❌ 完全沒有數字 |
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
- `CouncilVotes(population int) int`——人口→票數,採 **1:1 直接對應**(remake 近似;理由:手冊全篇
  沒有出現任何其他「人口單位」換算除數,且遊戲內其他以人口為基礎的量——如計分公式「+1 point per
  population unit」——同樣是 1:1 未縮放,是目前找不到更精確依據時最保守的讀法)。population<=0
  回傳 0(帝國已滅亡,無票)。
- `CouncilWinScoreBonus = 100`(MANUAL_150.html 權威值,預先記錄供未來計分系統使用,尚未接線——本
  remake 完全沒有計分系統,Score Calculation 整章都不在本輪範圍)。
- **2/3 超級多數門檻不重複實作**,直接沿用 `engine.CheckHighCouncil`;殲滅勝利同理沿用
  `engine.CheckExtermination`——避免兩套等價邏輯並存。

### 4.2 shell 層整合(`internal/shell/council.go`)

- `GameSession` 新增欄位:`Victory VictoryState`、`PendingCouncilElection *CouncilElection`、
  `LastCouncil string`、`CouncilMeetings int`、`lastCouncilTurn int`(存讀檔已同步,見
  `internal/shell/persist.go`)。
- `advanceCouncil()`:`EndTurn` 每回合呼叫的狀態機。議會成立(`councilEligible`)、距上次開會滿
  `councilInterval` 回合(首次成立立即開會,不用等)才開會;雙方票數 = `gamedata.CouncilVotes(人口)`;
  依 `engine.CheckHighCouncil` 判定:
  - 玩家達 2/3 → 立即勝利(`Victory.Reason = engine.VictoryHighCouncil`,不需要「接受」這一步——
    手冊那句「議會無法強迫你接受」只適用於「當選者不是你」的情境)。
  - AI 達 2/3 → 記錄 `PendingCouncilElection`,等玩家用 `RespondToCouncilElection(accept bool)` 回應:
    `accept=true` 結束遊戲判負,`accept=false` 不結束、下一屆再開(手冊原句直接翻譯成這個互動)。
  - 兩者皆未達標 → 流會,`LastCouncil` 記錄本屆票數,下一屆再開。
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

## 5. 資料模型限制(重要,誠實標注)

`shell.GameSession.AIPlayers` 目前**固定只有 1 個 AI 對手**(`NewDemoSession` 寫死一筆),場上存續
帝國數上限就是「玩家 + 1 AI」= 2。這對議會機制造成兩個字面上無法照搬手冊的地方:

1. **成立門檻「≥3 存續種族」永遠不可能達成。** 本 remake 用
   `councilMinExtantRacesOverride = 2`(`internal/shell/council.go`)覆寫這個門檻,**只影響 shell 層
   整合、不改動 `gamedata.CouncilMinExtantRaces` 這個手冊字面值常數**——未來若擴充 `AIPlayers` 支援
   多個 AI 對手,應把 override 改回手冊值 3。
2. **「兩位候選人由票數最高者出線」與「其餘種族依外交關係決定投給哪位候選人」這兩條規則沒有意義。**
   這兩條規則只在 ≥3 個帝國同場才有意義(需要「第三方」帝國把票投給兩位候選人之一)。本 remake 只有
   2 個帝國,沒有第三方可搖擺,直接把僅有的兩個帝國當成僅有的兩位候選人、各自的票就是自己的人口票。
   等資料模型支援 >1 個 AI 對手時,才需要真正實作「候選人由票數最高兩者出線 + 第三方依外交關係投票」
   ——`AIOpponent.Relation` 目前也只記錄「對玩家」的關係分數,沒有 AI 對 AI 的關係,這也是屆時需要
   一併補的資料。

`councilInterval = 8`(議會重開間隔)是 remake 排程選擇,手冊完全沒有給這個數字,只從外交台詞證實
議會確實會反覆召開;與 `antaresInterval`(15 回合,安塔蘭突襲)同數量級但較短,理由是議會需要「半數
銀河已殖民」這個較晚才達成的前置條件,間隔太長會讓一局遊戲只夠開 1-2 屆。

## 6. TODO(誠實列出,不硬做)

- **Antares 母星次元傳送門勝利**(手冊第二條路徑):完全沒有對應流程——無 Dimensional Portal 科技/
  建造、無「派遣艦隊前往 Antares」的星際航行目的地、無母星戰鬥。`engine.VictoryAntaran` 列舉值已存在
  但本 remake 永遠不會產生這個結果。需要一整套新子系統,超出本輪任務範圍。
- **計分系統(Score Calculation)**:manual/MANUAL_150.html 給了完整公式(時間分/人口分/科技分/
  殲滅加分/Guardian/Antares/Council 各項獎勵),本 remake 完全沒有計分/歷史圖表,`CouncilWinScoreBonus`
  只是預先記錄的權威值。
- **議會選舉結束畫面 + accept/reject 互動 UI**:目前只有文字狀態,沒有原版議會 3D 場景的投票動畫、
  沒有結束畫面(勝利/落敗的專屬畫面),`RespondToCouncilElection` 也還沒有 UI 熱區可以觸發。
- **多 AI 對手支援**:見上「資料模型限制」——這是讓議會機制真正貼近手冊(≥3 種族、候選人+外交搖擺票)
  的前置工程,目前不在範圍內。

## 7. 測試

- `internal/gamedata/council_test.go`:`CouncilEligible`(門檻邊界)、`CouncilVotes`(含負值/零值)。
- `internal/shell/council_test.go`:議會未成立不開會、成立後立即開第一屆、玩家達標直接勝利、AI 達標
  待回應(拒絕不結束+下一屆再開、接受才結束)、五五波流會、`DisableEvents` 關閉議會、殲滅勝利判定
  (含「對手仍存活不誤判」)。
- `internal/engine/victory_test.go`(既有,2026-07-03):`CheckExtermination`/`CheckHighCouncil`/
  `CheckAntaranVictory`/`CheckVictory` 純函式門檻邊界。
