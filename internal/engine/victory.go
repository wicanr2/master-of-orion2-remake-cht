package engine

// VictoryCondition 列舉 MOO2 的勝利方式。
//
// 手冊原文(moo2_patch1.5/MANUAL_150.html,"Win Conditions" 一節):
//
//	"There are three ways to win the game - exterminate all opponents; get elected
//	as the supreme leader of the galaxy by a two-thirds supermajority vote; or lead
//	a successful assault against the Antaran homeworld by sending your fleet to
//	their realm using a Dimensional Portal."
//
// 詳細規則見 CD 手冊(moo2_patch1.5/GAME_MANUAL.pdf,第 12 章
// "The End of the Game" → "Winning",第 182-183 頁)。
type VictoryCondition int

const (
	// VictoryNone 表示本回合尚未達成任何勝利條件。
	VictoryNone VictoryCondition = iota
	// VictoryExtermination:滅絕勝利——只剩一個玩家存活。
	// 手冊原文(GAME_MANUAL.pdf 第 183 頁):"if yours is the only surviving race,
	// as its emperor you rule the galaxy."
	VictoryExtermination
	// VictoryHighCouncil:銀河議會(Galactic Council)選舉勝利——候選人得票達
	// 三分之二超級多數。見 CheckHighCouncil 的手冊引文。
	VictoryHighCouncil
	// VictoryAntaran:擊敗安塔蘭母星勝利。
	// 手冊原文(GAME_MANUAL.pdf 第 183 頁):"seek out and defeat the Antaran home
	// fleet ... Once you defeat the awe-inspiring Antarans, all the other races in
	// the galaxy recognise your overwhelming superiority and quickly capitulate."
	VictoryAntaran
)

// CheckExtermination 判斷「只剩一個未被消滅的玩家」的滅絕勝利。
// alive[i] = true 表示玩家 i 仍存活(尚有殖民地/未被消滅、未投降出局)。
//
// 回傳 (是否達成滅絕勝利, 勝者 index)；未達成(存活數 != 1,含全滅或多方並存)時
// winner = -1。
func CheckExtermination(alive []bool) (bool, int) {
	winner := -1
	count := 0
	for i, a := range alive {
		if a {
			count++
			winner = i
		}
	}
	if count == 1 {
		return true, winner
	}
	return false, -1
}

// CheckHighCouncil 判斷銀河議會(Galactic Council)選舉勝利:候選人得票是否達到
// 「full two-thirds majority」的超級多數門檻。
//
// 手冊原文(GAME_MANUAL.pdf 第 183 頁,"Winning" 小節):
//
//	"When half of the galaxy has been settled ... If there are 3 or more extant
//	races, they gather and form the Galactic Council ... Based on the size of the
//	population of each empire, the leader of every race is assigned a number of
//	votes. Two contenders are chosen — those whose empires wield the most votes.
//	... If one of the nominees receives a full two-thirds majority of the votes,
//	that leader becomes ruler of the galaxy and the game is over."
//
// 手冊只描述「票數依人口規模分配」,未給「人口 → 票數」的精確換算公式或係數,
// GAME_MANUAL.pdf 與 MANUAL_150.html(1.5 版新增內容)均未補上此公式。
// TODO 票數計算公式待查證(目前無出處可移植,不臆造係數);故本函式的
// votesFor/totalVotes 一律由呼叫端算好傳入,本函式只負責門檻判定。
//
// 用整數運算避免浮點誤差:votesFor*3 >= totalVotes*2 等價於
// votesFor/totalVotes >= 2/3。totalVotes <= 0 視為無效輸入,一律不通過。
func CheckHighCouncil(votesFor, totalVotes int) bool {
	if totalVotes <= 0 {
		return false
	}
	return votesFor*3 >= totalVotes*2
}

// CheckAntaranVictory 判斷是否已攻陷安塔蘭母星(Antaran homeworld)。
// 手冊原文見 VictoryAntaran 的引文;細節(需先建 Dimensional Gate/Portal 才能
// 出征安塔蘭領域)屬艦隊/科技前置條件,不在本函式判定範圍——本函式只接收
// 「母星是否已被攻陷」的結果旗標。
func CheckAntaranVictory(antaranHomeworldConquered bool) bool {
	return antaranHomeworldConquered
}

// CheckVictory 綜合判定本回合是否達成任一勝利條件,回傳 (勝利條件, 勝者 index)。
// 三條件互不排斥(手冊:「If you time it right you can achieve two win conditions」),
// 但一局遊戲只需其一即結束;手冊未規定同回合多條件命中時的優先序,本函式依
// 滅絕 → 安塔蘭 → 議會選舉的順序回傳第一個命中者。
//
// 參數:
//   - alive: 各玩家存活旗標,供 CheckExtermination 使用。
//   - antaranWinner, antaranHomeworldConquered: 供 CheckAntaranVictory 使用;
//     antaranWinner 是攻陷母星的玩家 index(由呼叫端決定,本函式不追蹤艦隊歸屬)。
//   - councilWinner, votesFor, totalVotes: 供 CheckHighCouncil 使用;
//     councilWinner 是得票候選人的玩家 index。
func CheckVictory(alive []bool, antaranWinner int, antaranHomeworldConquered bool, councilWinner, votesFor, totalVotes int) (VictoryCondition, int) {
	if ok, winner := CheckExtermination(alive); ok {
		return VictoryExtermination, winner
	}
	if CheckAntaranVictory(antaranHomeworldConquered) {
		return VictoryAntaran, antaranWinner
	}
	if CheckHighCouncil(votesFor, totalVotes) {
		return VictoryHighCouncil, councilWinner
	}
	return VictoryNone, -1
}
