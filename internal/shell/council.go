package shell

import (
	"fmt"
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 銀河議會(Galactic Council)選舉勝利條件的引擎層整合。權威規則來源見
// internal/gamedata/council.go 檔頭的手冊逐字引用(GAME_MANUAL.pdf p.183)。本檔處理
// 「什麼時候開會、誰的票、誰贏、贏了之後遊戲怎麼繼續/結束」這些 shell 層狀態機邏輯;
// 純數值判定(門檻/票數/2/3多數/滅絕)一律呼叫 gamedata 或 engine 既有函式,不在這裡重算——
// 2/3 超級多數與滅絕勝利的判定沿用 internal/engine/victory.go 既有的 CheckHighCouncil/
// CheckExtermination(2026-07-03 就已存在,但先前從未被 shell/cmd 呼叫過,是一組沒接進實際
// 回合流程的死碼,見 gamedata/council.go 檔頭說明)。

// 2026-07-11 訂正:NewDemoSession 已由 1 個 AI 對手擴為 3 個(見該函式),場上存續帝國數
// 上限變成「玩家 + 3 AI」= 4,手冊字面門檻 gamedata.CouncilMinExtantRaces(3)現在真的可達,
// 不再需要 councilMinExtantRacesOverride 這個「資料模型限制下的近似覆寫值」——已移除該常數,
// councilEligible 直接引用 gamedata.CouncilMinExtantRaces。舊有「固定只有 1 AI 故永遠死路徑」
// 的說明不再成立,見下方 councilEligible/advanceCouncil。

// councilInterval 是議會成立後、每屆選舉的間隔回合數。手冊沒有給出這個數字(見
// gamedata/council.go 檔尾說明),只從外交台詞(assets/i18n/diplo.tsv「next meeting of the
// Council」「your last Council vote」)得知議會確實會反覆召開。8 回合是 remake 排程選擇,
// 與 antaresInterval(15回合,安塔蘭突襲)同數量級但較短,理由:議會選舉需要「半數銀河已殖民」
// 這個較晚才達成的前置條件,若再疊加太長的重開間隔,一局遊戲可能只夠開 1-2 屆,體驗上很難
// 感受到「反覆投票、外交態勢影響選情」這個手冊描述的機制。若之後找到手冊/社群逆向出的精確
// 間隔值,應更新此常數。
const councilInterval = 8

// victoryReasonLabel 回傳 engine.VictoryCondition 的中文化描述,供回合摘要/畫面顯示
// (engine 是純規則層,不放 UI 字串,故中文標籤放在 shell)。
func victoryReasonLabel(r engine.VictoryCondition) string {
	switch r {
	case engine.VictoryExtermination:
		return "殲滅所有對手"
	case engine.VictoryHighCouncil:
		return "銀河議會選舉當選銀河領袖"
	case engine.VictoryAntaran:
		return "攻陷安塔蘭母星"
	default:
		return ""
	}
}

// VictoryState 記錄本局遊戲是否已分出勝負。Over=false 時其餘欄位無意義。
//
// ⚠ 目前 Over=true 後 EndTurn 仍會繼續正常推進(不會強制擋下後續操作)——這是刻意的最小整合:
// 「遊戲結束後應該鎖死操作、顯示結束畫面」屬於 UI/流程層決定,本輪任務只接引擎層勝負判定,
// 見 docs/HONEST-STATUS.md 誠實標注此限制。
//
// Reason 沿用 engine.VictoryCondition(engine.VictoryExtermination/VictoryHighCouncil/
// VictoryAntaran),不在 shell 另建重複列舉。engine.VictoryAntaran 本 remake 完全沒有對應流程
// (無 Dimensional Portal、無「派遣艦隊前往 Antares 母星」的航行目的地、無母星戰鬥),需要一整套
// 新子系統,超出本輪任務範圍,故 shell 層目前不會產生 Reason==VictoryAntaran 的 VictoryState,
// 留待後續 worklist(見 docs/HONEST-STATUS.md 勝利條件章節 TODO)。
type VictoryState struct {
	Over   bool
	Reason engine.VictoryCondition
	Winner string // "player" 或 AI 名稱(如 s.AIPlayers[i].Name)
	Turn   int    // 達成勝利的回合數
}

// CouncilElection 是一屆已召開、且某方達到 2/3 多數但當選者不是玩家的議會選舉——依手冊
// 「there's no way the council can force you to accept a decision you don't agree with」,
// 曝露給玩家一個 accept/reject 選擇(見 GameSession.RespondToCouncilElection)。
type CouncilElection struct {
	Turn        int
	PlayerVotes int
	EnemyVotes  int
	TotalVotes  int
	EnemyName   string
}

// settledStarFraction 回傳目前已殖民星數(Owner!=0)與銀河總星數,供
// gamedata.CouncilEligible 的「半數銀河已殖民」條件使用。
func (s *GameSession) settledStarFraction() (settled, total int) {
	for _, st := range s.Stars {
		if st.Owner != 0 {
			settled++
		}
	}
	return settled, len(s.Stars)
}

// extantRaceCount 回傳目前存續帝國數(玩家至少還有 1 個殖民地算 1、每個至少還有 1 個殖民地的
// AI 對手各算 1)。用於議會成立門檻,見 councilMinExtantRacesOverride 註解(資料模型限制,
// 本 remake 上限恆為 2)。
func (s *GameSession) extantRaceCount() int {
	n := 0
	if len(s.PlayerColonies) > 0 {
		n++
	}
	for _, a := range s.AIPlayers {
		if len(a.Colonies) > 0 {
			n++
		}
	}
	return n
}

// councilEligible 判定議會這回合是否應該存在:半數銀河已殖民 + 存續帝國數達
// gamedata.CouncilMinExtantRaces(手冊字面值 3;見該常數註解,NewDemoSession 現有玩家+3 AI=4
// 個帝國上限,字面門檻可達,不再需要 remake 覆寫值)。
func (s *GameSession) councilEligible() bool {
	settled, total := s.settledStarFraction()
	if total <= 0 {
		return false
	}
	return settled*2 >= total && s.extantRaceCount() >= gamedata.CouncilMinExtantRaces
}

// playerPopulationTotal/aiPopulationTotal 回傳各自帝國殖民地人口加總,做為
// gamedata.CouncilVotes 的輸入(手冊:票數依人口規模決定,見 gamedata/council.go)。
//
// aiPopulationTotal 是「所有 AI 對手合計」——只供 CouncilStatus 顯示用的既有相容欄位
// (UI 呈現「我方 vs 全體 AI 陣營合計」的粗略對照,見該型別註解),不是 advanceCouncil 實際
// 判定 2/3 多數勝負的依據。真正判定勝負時每個帝國(玩家與各 AI)各自獨立算票,見 advanceCouncil
// 的 empireVotes,不把多個 AI 的人口灌成同一票——手冊原文是「每個種族的領袖各自被分配一定票數」
// (leader of every race is assigned a number of votes),多 AI 情境下把它們合計成一票會讓
// 「某一個 AI 單獨達 2/3」與「AI 們合計達 2/3 但個別都沒過半」這兩種手冊語意不同的情況混淆。
func (s *GameSession) playerPopulationTotal() int {
	n := 0
	for _, c := range s.PlayerColonies {
		n += c.Population
	}
	return n
}

func (s *GameSession) aiPopulationTotal() int {
	n := 0
	for _, a := range s.AIPlayers {
		for _, c := range a.Colonies {
			n += c.Population
		}
	}
	return n
}

// empireVote 是 advanceCouncil 用來逐帝國(玩家或某個 AI)算票的中介結構。
// idx==-1 代表玩家,>=0 代表 s.AIPlayers[idx]。
type empireVote struct {
	idx   int    // -1=玩家,否則 AIPlayers 索引
	name  string // "player" 或 AIOpponent.Name
	votes int    // 該帝國自身基礎票數(gamedata.CouncilVotes(人口))
}

// councilRelation 回傳投票者(voter)對某帝國(target)的外交關係分數,供議會搖擺票偏好判定。
// idx==-1 代表玩家。玩家↔AI 以 AIOpponent.Relation 作對稱代理(remake 未單獨建模玩家對 AI 的
// 關係);AI↔AI 用 AIRelations 矩陣。分數越高越傾向把票投給對方。
func (s *GameSession) councilRelation(voter, target int) int {
	switch {
	case voter == target:
		return 40 // 對自己最友好(候選人投自己)
	case voter == -1: // 玩家 → AI target
		if target >= 0 && target < len(s.AIPlayers) {
			return s.AIPlayers[target].Relation
		}
	case target == -1: // AI voter → 玩家
		if voter >= 0 && voter < len(s.AIPlayers) {
			return s.AIPlayers[voter].Relation
		}
	default: // AI voter → AI target
		if voter >= 0 && voter < len(s.AIRelations) && target >= 0 && target < len(s.AIRelations[voter]) {
			return s.AIRelations[voter][target]
		}
	}
	return 0
}

// councilSwingVoteMinRelation 是搖擺票的棄權門檻:非候選帝國唯有對某候選人的外交關係達此值
// (「友好」以上,對齊 AIRelationName 的 >=8 界線)才會把票投給它,否則棄權。手冊:議會中對兩位
// 候選人都不夠友好的種族會棄權,棄權票計入 2/3 的分母卻不歸任何候選人——這正是選情膠著、反覆
// 流會的來源(若強迫每票都投給某候選人,中立票會全湧向領先者,失真為「幾乎每屆都有人當選」)。
const councilSwingVoteMinRelation = 8

// councilTally 是一屆議會計票結果(手冊 GAME_MANUAL.pdf p.183「兩位候選人由票數最高者出線,
// 其餘種族依外交關係決定投給哪位候選人」的忠實建模)。
type councilTally struct {
	candIdx   [2]int    // 兩位候選人的帝國索引(-1=玩家)
	candName  [2]string // 候選人名(display)
	candVotes [2]int    // 候選人最終得票(自身基礎票 + 收到的搖擺票)
	total     int       // 全體帝國基礎票總和(2/3 門檻的分母)
	valid     bool      // 是否湊足兩位候選人(帝國數<2 時為 false)
	rows      []councilVoteRow // 逐帝國投票明細(依基礎票降冪),供 UI 呈現;判定邏輯不讀此欄
}

// councilVoteRow 是一屆選舉中單一帝國的投票明細(供議會畫面逐列呈現)。
type councilVoteRow struct {
	idx         int    // -1=玩家,否則 AIPlayers 索引
	name        string // "player" 或 AIOpponent.Name
	baseVotes   int    // 該帝國自身票數
	isCandidate bool   // 是否為兩位候選人之一
	votedForIdx int    // 搖擺票投給的候選人 idx;-1/>=0 皆可能,candidateAbstain 表棄權;候選人=自身 idx
}

// councilVoteAbstain 是 councilVoteRow.votedForIdx 的哨兵值,表示該帝國棄權(未投給任一候選人)。
// 用 -2(玩家是 -1、AI 是 >=0,都不會撞到)。
const councilVoteAbstain = -2

// tallyCouncil 忠實模擬一屆選舉:
//  1. 每個帝國(玩家 + 各 AI)依人口算基礎票(gamedata.CouncilVotes)。
//  2. 票數最高的兩位帝國出線為候選人(穩定排序,平手時保留原順序=玩家優先)。
//  3. 其餘帝國(含玩家若非候選人)各自把「全部票數」投給外交關係較好、且達友好門檻
//     (councilSwingVoteMinRelation)的那位候選人;對兩位都不夠友好則棄權(不投票,但票數仍計入
//     分母 total)。無亂數,可決定性。
//  4. 候選人自身票數計入自己。
//
// 棄權建模見 councilSwingVoteMinRelation:分母固定為全體基礎票總和(含棄權者),故中立局勢下
// 領先者難以單靠自身票達 2/3,選情會反覆流會,與手冊描述一致。玩家若非候選人,這裡自動依關係
// 代投/棄權(advanceCouncil 在 EndTurn 內非互動呼叫);「玩家親自選票」屬 UI 互動功能,列為 TODO
// (見 docs/HONEST-STATUS.md 議會章節)。
func (s *GameSession) tallyCouncil() councilTally {
	s.ensureAIRelations()
	emps := make([]empireVote, 0, 1+len(s.AIPlayers))
	emps = append(emps, empireVote{idx: -1, name: "player", votes: gamedata.CouncilVotes(s.playerPopulationTotal())})
	for i, a := range s.AIPlayers {
		pop := 0
		for _, c := range a.Colonies {
			pop += c.Population
		}
		emps = append(emps, empireVote{idx: i, name: a.Name, votes: gamedata.CouncilVotes(pop)})
	}
	total := 0
	for _, e := range emps {
		total += e.votes
	}
	if len(emps) < 2 {
		return councilTally{total: total}
	}
	// 票數前二為候選人(穩定排序,平手保留原順序,玩家在最前)。
	order := make([]int, len(emps))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return emps[order[a]].votes > emps[order[b]].votes })
	ca, cb := emps[order[0]], emps[order[1]]
	votesA, votesB := ca.votes, cb.votes
	rows := make([]councilVoteRow, 0, len(order))
	rows = append(rows,
		councilVoteRow{idx: ca.idx, name: ca.name, baseVotes: ca.votes, isCandidate: true, votedForIdx: ca.idx},
		councilVoteRow{idx: cb.idx, name: cb.name, baseVotes: cb.votes, isCandidate: true, votedForIdx: cb.idx})
	// 其餘帝國依外交關係把全部票投給偏好且達友好門檻的候選人;對兩位都不夠友好則棄權。
	for k := 2; k < len(order); k++ {
		e := emps[order[k]]
		relA := s.councilRelation(e.idx, ca.idx)
		relB := s.councilRelation(e.idx, cb.idx)
		votedFor := councilVoteAbstain
		switch {
		case relA >= relB && relA >= councilSwingVoteMinRelation:
			votesA += e.votes
			votedFor = ca.idx
		case relB > relA && relB >= councilSwingVoteMinRelation:
			votesB += e.votes
			votedFor = cb.idx
			// 否則:對兩位候選人都不夠友好,棄權(不投票,票數仍在 total 分母內)。
		}
		rows = append(rows, councilVoteRow{idx: e.idx, name: e.name, baseVotes: e.votes, votedForIdx: votedFor})
	}
	return councilTally{
		candIdx: [2]int{ca.idx, cb.idx}, candName: [2]string{ca.name, cb.name},
		candVotes: [2]int{votesA, votesB}, total: total, valid: true, rows: rows,
	}
}

// advanceCouncil 是 EndTurn 每回合呼叫的議會選舉狀態機:
//  1. 遊戲已分出勝負、或議會尚未成立(councilEligible=false)、或上一屆選舉玩家還沒回應
//     (PendingCouncilElection!=nil)→ 不開會。
//  2. 距離上次開會不足 councilInterval 回合 → 不開會(首次成立後立刻召開第一屆,不用等)。
//  3. 開會:呼叫 tallyCouncil 忠實計票(票數最高兩位帝國出線為候選人,其餘帝國依外交關係把票
//     投給友好的候選人、對兩位都不夠友好則棄權),再依 engine.CheckHighCouncil 用「候選人得票 vs
//     全體基礎票總和」判定是否達 2/3 多數(玩家候選人優先判定;只可能有一方達 2/3,門檻排他)。
//     - 玩家(候選人)達標 → 立即勝利(手冊:當選者若是玩家,遊戲直接結束,不需要「接受」步驟)。
//     - 某 AI(候選人)達標 → 記錄 PendingCouncilElection(EnemyName=該 AI),等 RespondToCouncilElection。
//     - 無人達標 → 流會,下一屆再開(手冊描述議會反覆召開,見 diplo.tsv 台詞)。
//
// 2026-07-12 由「每帝國各自票數獨立比 2/3」的簡化 generalize,升級為 tallyCouncil 的忠實搖擺票
// 模型(兩位候選人 + 第三方依外交關係投票/棄權),見 tallyCouncil 與 councilSwingVoteMinRelation
// 註解。此前的簡化(每帝國獨立比、不分候選人、無搖擺票)是「還沒有 AI 對 AI 關係矩陣」時的權宜
// 讀法;AIRelations 建立後(見 session.go advanceAIDiplomacy),才有資料支撐手冊原文的搖擺票規則。
// 舊測試斷言的「票數分散即流會」在忠實模型下仍成立——中立局勢下第三方棄權,分母含棄權票,領先者
// 難單靠自身票達 2/3(見 multi_ai_test.go 的 NoneReachesMajority / NeutralAISAbstain 對照)。
func (s *GameSession) advanceCouncil() {
	s.LastCouncil = ""
	if s.DisableEvents || s.Victory.Over || s.PendingCouncilElection != nil {
		return
	}
	if !s.councilEligible() {
		return
	}
	if s.CouncilMeetings > 0 && s.Turn-s.lastCouncilTurn < councilInterval {
		return
	}

	tally := s.tallyCouncil()
	s.CouncilMeetings++
	s.lastCouncilTurn = s.Turn

	// 玩家最終得票(候選人則含搖擺票,非候選人則為自身基礎票)供顯示/CouncilElection 記錄用。
	playerVotes := gamedata.CouncilVotes(s.playerPopulationTotal())
	for c := 0; c < 2; c++ {
		if tally.candIdx[c] == -1 {
			playerVotes = tally.candVotes[c]
		}
	}

	// 逐候選人判定 2/3;玩家候選人優先判定(玩家自己當選直接勝利,不需 accept 步驟)。
	// 只可能有一位達 2/3(門檻排他),故判定順序不影響結果,只影響玩家/AI 分支。
	for c := 0; c < 2; c++ {
		if !tally.valid || tally.candIdx[c] != -1 {
			continue
		}
		if engine.CheckHighCouncil(tally.candVotes[c], tally.total) {
			s.Victory = VictoryState{Over: true, Reason: engine.VictoryHighCouncil, Winner: "player", Turn: s.Turn}
			s.LastCouncil = fmt.Sprintf("銀河議會第 %d 屆選舉:你以 %d/%d 票(達2/3多數)當選銀河領袖!",
				s.CouncilMeetings, tally.candVotes[c], tally.total)
			return
		}
	}
	for c := 0; c < 2; c++ {
		if !tally.valid || tally.candIdx[c] == -1 {
			continue
		}
		if engine.CheckHighCouncil(tally.candVotes[c], tally.total) {
			s.PendingCouncilElection = &CouncilElection{Turn: s.Turn, PlayerVotes: playerVotes,
				EnemyVotes: tally.candVotes[c], TotalVotes: tally.total, EnemyName: tally.candName[c]}
			s.LastCouncil = fmt.Sprintf("銀河議會第 %d 屆選舉:%s 以 %d/%d 票(達2/3多數)當選銀河領袖,尚待你回應是否接受",
				s.CouncilMeetings, tally.candName[c], tally.candVotes[c], tally.total)
			return
		}
	}
	if tally.valid {
		s.LastCouncil = fmt.Sprintf("銀河議會第 %d 屆選舉:候選人 %s（%d 票)、%s（%d 票)皆未達2/3多數,流會(全體 %d 票)",
			s.CouncilMeetings, tally.candName[0], tally.candVotes[0], tally.candName[1], tally.candVotes[1], tally.total)
	} else {
		s.LastCouncil = fmt.Sprintf("銀河議會第 %d 屆選舉:候選人不足,流會", s.CouncilMeetings)
	}
}

// CouncilStatus 是議會目前狀態的唯讀快照,供 UI 呈現用(cmd/moo2 是 package main,無法直接讀
// GameSession 的未匯出欄位/方法如 councilEligible,故提供這個匯出方法統一取值)。PlayerVotes/
// EnemyVotes/TotalVotes 是「若這回合真的開會,票數會是多少」的即時試算,不代表本回合一定會
// 開會——是否真的開會/流會/分出勝負以 advanceCouncil 每回合的結算為準,這裡只是唯讀快照。
//
// ⚠ EnemyVotes/EnemyName(Pending==nil 時)是「全體 AI 陣營合計」的簡化顯示,不是
// advanceCouncil 實際判定 2/3 多數勝負時逐帝國分算的真實依據(見 advanceCouncil 的
// empireVote/votes)——3 個 AI 對手性格互異,個別人口可能差很多,合計數字只適合當 UI 摘要,
// 不能拿來反推「哪個 AI 快贏了」。若要做真正的 N-way 議會 UI(逐帝國票數列表、指認哪個 AI
// 領先),屬 UI 大改範圍,本輪任務標 TODO 不做(cmd/moo2 目前的議會畫面本來就只顯示单行文字
// 摘要,見 interactive.go council())。Pending!=nil 時 EnemyName/EnemyVotes 改用
// CouncilElection 裡記錄的真實當選 AI(advanceCouncil 寫入),準確,不受本限制影響。
type CouncilStatus struct {
	Eligible    bool // 議會目前是否已成立(councilEligible)
	PlayerVotes int
	EnemyVotes  int
	TotalVotes  int
	EnemyName   string
	Meetings    int              // 已召開過的屆數
	Pending     *CouncilElection // 非 nil = 有待玩家回應的選舉結果
	Victory     VictoryState
}

// CouncilStatus 回傳議會目前狀態快照(見型別註解)。
func (s *GameSession) CouncilStatus() CouncilStatus {
	enemyName := "對手"
	switch len(s.AIPlayers) {
	case 0:
		// 保留預設值 "對手"
	case 1:
		enemyName = s.AIPlayers[0].Name
	default:
		enemyName = fmt.Sprintf("全體 AI 陣營(%d 方合計)", len(s.AIPlayers))
	}
	pv := gamedata.CouncilVotes(s.playerPopulationTotal())
	ev := gamedata.CouncilVotes(s.aiPopulationTotal())
	return CouncilStatus{
		Eligible: s.councilEligible(), PlayerVotes: pv, EnemyVotes: ev, TotalVotes: pv + ev,
		EnemyName: enemyName, Meetings: s.CouncilMeetings, Pending: s.PendingCouncilElection,
		Victory: s.Victory,
	}
}

// CouncilVoteRow 是議會逐帝國投票明細的匯出版本(供 cmd/moo2 議會畫面呈現)。VotedFor 是該帝國
// 這一票投給的候選人 display 名;候選人自身列 IsCandidate=true、VotedFor 為自己;棄權列
// Abstained=true、VotedFor 空字串。
type CouncilVoteRow struct {
	Name        string // display:玩家為「你」,AI 為其名稱
	IsPlayer    bool
	BaseVotes   int
	IsCandidate bool
	Abstained   bool
	VotedFor    string
}

// CouncilBreakdown 是一屆選舉的完整逐帝國明細(供議會畫面逐列呈現搖擺票/棄權)。Valid=false 表示
// 帝國數不足兩位(無選舉)。Threshold 是達 2/3 所需票數(向上取整,顯示用)。
type CouncilBreakdown struct {
	Valid      bool
	Rows       []CouncilVoteRow // 依基礎票降冪
	Candidates [2]string
	CandVotes  [2]int
	Total      int
	Threshold  int
}

// councilDisplayName 把帝國 idx(-1=玩家)轉成畫面用中文名。
func (s *GameSession) councilDisplayName(idx int) string {
	if idx == -1 {
		return "你"
	}
	if idx >= 0 && idx < len(s.AIPlayers) {
		return s.AIPlayers[idx].Name
	}
	return "對手"
}

// CouncilBreakdown 回傳「若這回合開會」的逐帝國票數與投票去向明細(即時試算,不代表本回合一定
// 開會;是否真的開會以 advanceCouncil 為準,同 CouncilStatus 的即時試算語意)。供議會畫面把搖擺票
// 攤開呈現,取代舊的單行合計摘要。
func (s *GameSession) CouncilBreakdown() CouncilBreakdown {
	t := s.tallyCouncil()
	if !t.valid {
		return CouncilBreakdown{Total: t.total}
	}
	rows := make([]CouncilVoteRow, 0, len(t.rows))
	for _, r := range t.rows {
		row := CouncilVoteRow{
			Name: s.councilDisplayName(r.idx), IsPlayer: r.idx == -1,
			BaseVotes: r.baseVotes, IsCandidate: r.isCandidate,
		}
		switch {
		case r.isCandidate:
			row.VotedFor = s.councilDisplayName(r.idx) // 候選人投自己
		case r.votedForIdx == councilVoteAbstain:
			row.Abstained = true
		default:
			row.VotedFor = s.councilDisplayName(r.votedForIdx)
		}
		rows = append(rows, row)
	}
	return CouncilBreakdown{
		Valid: true, Rows: rows,
		Candidates: [2]string{s.councilDisplayName(t.candIdx[0]), s.councilDisplayName(t.candIdx[1])},
		CandVotes:  t.candVotes, Total: t.total,
		Threshold: (t.total*2 + 2) / 3, // ceil(total*2/3):達 2/3 所需票數
	}
}

// VictoryReasonLabel 是 victoryReasonLabel 的匯出版本,供 cmd/moo2 顯示中文勝利路徑描述用。
func VictoryReasonLabel(r engine.VictoryCondition) string {
	return victoryReasonLabel(r)
}

// RespondToCouncilElection 是玩家對「非玩家當選」的回應(手冊:「there's no way the council
// can force you to accept a decision you don't agree with」)。
//
//	accept=true  → 接受落敗,遊戲結束(Victory.Winner=當選 AI 名稱)。
//	accept=false → 拒絕接受,不結束遊戲,清空待決狀態,下一屆(councilInterval 回合後)再開會。
//
// PendingCouncilElection==nil 時呼叫視為無操作(沒有待決選舉可回應)。
func (s *GameSession) RespondToCouncilElection(accept bool) {
	if s.PendingCouncilElection == nil {
		return
	}
	pending := s.PendingCouncilElection
	if accept {
		s.Victory = VictoryState{Over: true, Reason: engine.VictoryHighCouncil, Winner: pending.EnemyName, Turn: pending.Turn}
	}
	s.PendingCouncilElection = nil
}

// advanceConquestVictory 偵測手冊第一條勝利路徑:殲滅所有對手,沿用 engine.CheckExtermination
// (alive[0]=玩家,alive[1:]=各 AI 對手,依 AIPlayers 順序;「存活」= 該帝國目前殖民地數 > 0)。
// CheckExtermination 本身是雙向對稱的「只剩一方存活」判定,故本函式同時涵蓋兩種結果:
//   - 玩家存活、所有 AI 對手皆 0 殖民地 → 玩家以 VictoryExtermination 勝利(InvadeColony
//     攻陷 AI 唯一殖民地後會把該筆從 AIOpponent.Colonies 移除,見 ground_invasion.go)。
//   - 玩家 0 殖民地、某 AI 對手存活 → 該 AI 以 VictoryExtermination「勝利」(玩家戰敗)。本
//     remake 目前沒有任何機制會讓 PlayerColonies 完全清空(安塔蘭突襲只扣人口不摧毀殖民地、
//     AI 無地面入侵玩家的邏輯),故這個分支現況下不可達,只是沿用同一個對稱判定函式的
//     自然結果,不是額外實作的機制。
//
// len(s.AIPlayers)==0 視為未設置對手,不觸發(避免測試/工具建構的 GameSession 意外判定勝利)。
func (s *GameSession) advanceConquestVictory() {
	if s.Victory.Over {
		return
	}
	if len(s.AIPlayers) == 0 {
		return
	}
	alive := make([]bool, 1+len(s.AIPlayers))
	alive[0] = len(s.PlayerColonies) > 0
	for i, a := range s.AIPlayers {
		alive[i+1] = len(a.Colonies) > 0
	}
	ok, winner := engine.CheckExtermination(alive)
	if !ok {
		return
	}
	winnerName := "player"
	if winner > 0 {
		winnerName = s.AIPlayers[winner-1].Name
	}
	s.Victory = VictoryState{Over: true, Reason: engine.VictoryExtermination, Winner: winnerName, Turn: s.Turn}
}
