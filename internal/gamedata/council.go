package gamedata

// 銀河議會(Galactic Council)選舉勝利條件可驗證公式/常數,移植自
// moo2_patch1.5/GAME_MANUAL.pdf(p.183,「12. The End of the Game」> 「Winning」第三條路徑)
// 原文(已用 pdftotext -layout 擷取,逐字抄錄):
//
//	"The last and possibly most complicated method is to win an election of the Galactic
//	Council. When half of the galaxy has been settled, the threat of war over competition for
//	the habitable planets becomes too great. If there are 3 or more extant races, they gather
//	and form the Galactic Council to prevent future war. The Council's only order of business is
//	to select a leader to rule the entire galaxy. Based on the size of the population of each
//	empire, the leader of every race is assigned a number of votes. Two contenders are chosen
//	— those whose empires wield the most votes. How each race votes is determined on the
//	basis of current diplomatic relations. If one of the nominees receives a full two-thirds
//	majority of the votes, that leader becomes ruler of the galaxy and the game is over. Clearly,
//	your intention is to prevent others from being elected until you can yourself be elected to
//	hold sway over all of known space. Of course, there's no way the council can force you to
//	accept a decision you don't agree with."
//
// moo2_patch1.5/MANUAL_150.html(1.50 patch notes本身的「Win Conditions」摘要,與上文一致,
// 無版本差異)重申三條勝利路徑與 2/3 supermajority,並在「Score Calculation」章節給出
// 「Council Win」的計分獎勵:「Brings in a meager 100 points. The value can be changed with
// hi_score council.」——本檔一併收錄該分數常數,但完整計分系統(Score Calculation 整章)
// 本身在本 remake 尚未實作(無回合計分/歷史圖表),CouncilWinScoreBonus 只是預先记下的
// 權威值,供未來計分系統落地時直接引用,不代表計分已接線。
//
// docs/tech/rules-implementation-audit.md 第 10 項記載「openorion2 對 victory/winner/
// win_condition/gameOver 全 repo零命中」——該文件通篇分析對象是 openorion2(C++ 參考專案)
// 本身,不是本 remake(Go)的程式碼庫,這條記錄至今仍然成立,沒有過期,不要誤讀成「本 remake
// 也沒有勝利條件」。
//
// 本 remake 這邊,internal/engine/victory.go(commit 2cccf18,2026-07-03 14:19)其實已經有
// VictoryCondition 列舉 + CheckExtermination/CheckHighCouncil/CheckAntaranVictory/
// CheckVictory 四個純函式(引用同一份手冊原文),但當時只停在「純函式待命」——未被
// internal/shell 或 cmd/moo2 任何地方呼叫,是一組沒接進實際回合流程的死碼。本檔
// (CouncilEligible/CouncilVotes)補的正是 engine.CheckHighCouncil 自陳「votesFor/totalVotes
// 一律由呼叫端算好傳入」缺的那塊(人口→票數換算 + 議會成立門檻),2/3 超級多數的門檻判定本身
// 不重複實作,直接沿用 engine.CheckHighCouncil;殲滅勝利同樣沿用 engine.CheckExtermination,
// 不在本檔或 shell 層另建等價邏輯。整合位置見 internal/shell/council.go。

// CouncilMinExtantRaces 議會成立門檻之一:存續種族數(含玩家)。手冊原文「If there are 3 or
// more extant races, they gather and form the Galactic Council」——字面值 3。
//
// ⚠ 本 remake 目前的資料模型(shell.GameSession.AIPlayers)固定只支援 1 個 AI 對手
// (見 shell.NewDemoSession),故場上存續帝國數上限就是 2(玩家 + 1 AI),永遠不可能達到
// 這個字面門檻 3。這代表「議會依手冊字面值成立」在本 remake 現況下是不可觸發的死路徑。
// shell 層的實際整合(internal/shell 議會觸發判定)因此改用一個資料模型限制下的
// remake 近似覆寫值(見 shell 對應常數的註解),不是本檔改手冊數字——本檔保留手冊原文字面值
// 供未來擴充多 AI 對手時直接還原使用。
const CouncilMinExtantRaces = 3

// CouncilWinScoreBonus 議會選舉獲勝的計分獎勵(MANUAL_150.html「Score Calculation」>
// 「Council Win」:「Brings in a meager 100 points.」)。本 remake 尚無完整計分系統可接線,
// 此常數僅預先記錄權威值供未來使用。
const CouncilWinScoreBonus = 100

// CouncilEligible 判定銀河議會是否應該成立(手冊 p.183 兩條件同時成立):
//   - settledStars/totalStars:銀河至少半數星系已被殖民(settled*2 >= total)。
//   - extantRaces:存續帝國數 >= CouncilMinExtantRaces(字面值 3;呼叫端若因資料模型限制需要
//     覆寫這個門檻,應在呼叫端另建 remake 近似常數,不要改動本函式或 CouncilMinExtantRaces)。
func CouncilEligible(settledStars, totalStars, extantRaces int) bool {
	if totalStars <= 0 {
		return false
	}
	return settledStars*2 >= totalStars && extantRaces >= CouncilMinExtantRaces
}

// CouncilVotes 依人口換算票數。手冊原文只說「Based on the size of the population of each
// empire, the leader of every race is assigned a number of votes」,沒有給出人口→票數的
// 換算比例或公式(不像 Score Calculation 章節明講「+1 point for every unit of population」
// 那樣精確)。remake 近似:採最直接的 1 人口單位 = 1 票,理由是手冊全篇沒有出現任何其他
// 「人口單位」換算除數,且遊戲內其他以人口為基礎的量(如上述計分公式)同樣是 1:1 未縮放
// ——這是目前找不到更精確依據時最保守、最不引入額外編造係數的讀法。若未來找到官方/社群
// 逆向出的精確換算式,應更新本函式並移除這則近似說明。population<=0 回傳 0(該帝國已滅亡,
// 無票)。
func CouncilVotes(population int) int {
	if population <= 0 {
		return 0
	}
	return population
}

// 2/3 超級多數門檻判定不在本檔重複實作,直接呼叫 internal/engine.CheckHighCouncil
// (candidateVotes*3 >= totalVotes*2,整數運算避免浮點誤差,已有測試
// internal/engine/victory_test.go)。殲滅勝利同理沿用 internal/engine.CheckExtermination。

// 手冊規則的實作狀態(呼叫端需要精確公式時應先補查證,不要編造係數):
//   - 「兩位候選人由票數最高者出線」與「其餘種族依外交關係決定投給哪位候選人」:2026-07-12 已在
//     shell 層忠實實作(見 shell/council.go tallyCouncil / councilSwingVoteMinRelation)。條件是
//     資料模型已支援多 AI 對手(NewDemoSession 現建 3 AI,共 4 帝國,有真正的「第三方」可搖擺)
//     且 AI-vs-AI 外交關係矩陣已建立(shell/session.go advanceAIDiplomacy)。此前本註解記載的
//     「固定只有玩家+1 AI、沒有第三方、不實作搖擺票」已過時,是那個資料模型下的權宜說明。
//   - 議會多久召開一次(首次成立後的重複開會間隔):手冊原文沒有給出turns/years數字,只從
//     外交台詞(assets/i18n/diplo.tsv 大量「next meeting of the Council」「your last Council
//     vote」)得知議會確實會反覆召開。shell 層的重開間隔常數因此是 remake 排程選擇,非手冊值,
//     見 shell 對應常數註解。
