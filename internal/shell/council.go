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

// CouncilNoticeKind 是一屆議會在回合摘要留下的型別化結果；固定玩家文案由 UI catalog 提供。
type CouncilNoticeKind uint8

const (
	CouncilNoticeInsufficientCandidates CouncilNoticeKind = iota + 1
	CouncilNoticeVoteRequested
	CouncilNoticePlayerElected
	CouncilNoticeEnemyElectedPending
	CouncilNoticeNoMajority
)

// CouncilNotice 只保存重建議會回合通知所需的動態資料。CandidateIdx 沿用 -1=玩家、>=0=AI。
type CouncilNotice struct {
	Kind          CouncilNoticeKind
	Meeting       int
	CandidateIdx  [2]int
	CandidateName [2]string
	Votes         [2]int
	TotalVotes    int
	WinnerSlot    int // 0/1；只在 PlayerElected／EnemyElectedPending 有意義
}

// CouncilVotePending 保存原版 sub_1633C 的真人三選一畫面之前已完成的 AI 計票。
// 候選索引沿用 -1=玩家、>=0=AIPlayers；Choice 由 RespondToCouncilVote 傳入 0/1/2。
type CouncilVotePending struct {
	Turn            int
	CandidateIdx    [2]int
	CandidateName   [2]string
	CandidateVotes  [2]int
	TotalVotes      int
	PlayerBaseVotes int
	Rows            []CouncilVoteRow
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
	return gamedata.CouncilEligible(settled, total, s.extantRaceCount())
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

func (s *GameSession) ensureCouncilState() {
	if s.councilRand == nil {
		s.councilRand = newRandStream(s.EventSeed*2654435761 + 17)
	}
	n := 1 + len(s.AIPlayers)
	if len(s.CouncilLastVotes) != n {
		old := s.CouncilLastVotes
		s.CouncilLastVotes = make([]int, n)
		for i := range s.CouncilLastVotes {
			s.CouncilLastVotes[i] = councilVoteAbstain
		}
		copy(s.CouncilLastVotes, old)
	}
}

func (s *GameSession) councilPolicy(voter, target int) gamedata.ForeignPolicy {
	if voter == -1 || target == -1 {
		i := voter
		if i == -1 {
			i = target
		}
		if i >= 0 && i < len(s.AIPlayers) {
			return s.AIPlayers[i].Treaty.FormalPolicy
		}
		return gamedata.DIPLO_NONE
	}
	if voter >= 0 && voter < len(s.AIPolicies) && target >= 0 && target < len(s.AIPolicies[voter]) {
		return s.AIPolicies[voter][target]
	}
	return gamedata.DIPLO_NONE
}

func (s *GameSession) councilAgreement(voter, target int, research bool) bool {
	if voter == -1 || target == -1 {
		i := voter
		if i == -1 {
			i = target
		}
		if i < 0 || i >= len(s.AIPlayers) {
			return false
		}
		if research {
			return s.AIPlayers[i].Treaty.ResearchActive
		}
		return s.AIPlayers[i].Treaty.TradeActive
	}
	m := s.AITrade
	if research {
		m = s.AIResearch
	}
	return voter >= 0 && voter < len(m) && target >= 0 && target < len(m[voter]) && m[voter][target]
}

func (s *GameSession) councilTargetTraits(target int) (charismatic, repulsive, imperium bool) {
	if target == -1 {
		return s.RaceCharismatic(), s.RaceRepulsive(), s.effectiveGovernment() == gamedata.MoraleGovImperium
	}
	if target < 0 || target >= len(s.AIPlayers) {
		return
	}
	r := s.AIPlayers[target].RaceIndex
	return gamedata.OrigRaceTrait(r, gamedata.TRAIT_CHARISMATIC) != 0,
		gamedata.OrigRaceTrait(r, gamedata.TRAIT_REPULSIVE) != 0,
		gamedata.OrigRaceTrait(r, gamedata.TRAIT_GOVERNMENT) == int(gamedata.MoraleGovImperium)
}

// councilSupportsCandidate 對應 Vote_Check_ 已映射的消費端。四個尚未命名的 raw 情緒欄與
// sub_78398 不用 Relation 猜填；因此這是流程精確、已映射分數精確的保守子集。
func (s *GameSession) councilSupportsCandidate(voter, target, other int) bool {
	p := s.councilPolicy(voter, target)
	if p >= gamedata.DIPLO_LIMITED_WAR {
		return false
	}
	score := 0
	if p == gamedata.DIPLO_ALLIANCE {
		score = 200
	}
	if s.councilPolicy(voter, other) >= gamedata.DIPLO_LIMITED_WAR {
		score += 100
	}
	ch, rep, imp := s.councilTargetTraits(target)
	if ch {
		score += 40
	}
	if rep {
		score -= 100
	}
	if imp {
		score += 30
	}
	vi := voter + 1
	if vi >= 0 && vi < len(s.CouncilLastVotes) && s.CouncilLastVotes[vi] == target {
		score += 50
	}
	if p == gamedata.DIPLO_NON_AGGRESSION {
		score += 50
	}
	if s.councilAgreement(voter, target, false) {
		score += 25
	}
	if s.councilAgreement(voter, target, true) {
		score += 25
	}
	return s.councilRand.Intn(200)+1 <= score
}

// councilTally 是一屆議會計票結果(手冊 GAME_MANUAL.pdf p.183「兩位候選人由票數最高者出線,
// 其餘種族依外交關係決定投給哪位候選人」的忠實建模)。
type councilTally struct {
	candIdx   [2]int           // 兩位候選人的帝國索引(-1=玩家)
	candName  [2]string        // 候選人名(display)
	candVotes [2]int           // 候選人最終得票(自身基礎票 + 收到的搖擺票)
	total     int              // 全體帝國基礎票總和(2/3 門檻的分母)
	valid     bool             // 是否湊足兩位候選人(帝國數<2 時為 false)
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
//  3. 其餘 AI 對兩名候選人各做一次獨立 Vote_Check_；單邊通過才投票，雙真或雙假均棄權。
//     玩家不在這裡自動代投，而是建立 CouncilVotePending 等待三選一。
//  4. 候選人自身票數計入自己。
//
// 分母固定為全體基礎票總和（含棄權者）。亂數來自獨立 councilRand，抽取位置進存檔；
// 唯讀 CouncilBreakdown 只呈現已建立的待投票結果，不會提前擲骰。
func (s *GameSession) tallyCouncil() councilTally {
	s.ensureAIRelations()
	s.ensureCouncilState()
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
	// 其餘 AI 對兩位候選人分別擲 Vote_Check_；真人留給 sub_1633C 對應的待投票狀態。
	for k := 2; k < len(order); k++ {
		e := emps[order[k]]
		votedFor := councilVoteAbstain
		if e.idx == -1 {
			rows = append(rows, councilVoteRow{idx: e.idx, name: e.name, baseVotes: e.votes, votedForIdx: votedFor})
			continue
		}
		a := s.councilSupportsCandidate(e.idx, ca.idx, cb.idx)
		b := s.councilSupportsCandidate(e.idx, cb.idx, ca.idx)
		switch {
		case a && !b:
			votesA += e.votes
			votedFor = ca.idx
		case b && !a:
			votesB += e.votes
			votedFor = cb.idx
		}
		s.CouncilLastVotes[e.idx+1] = votedFor
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
//  2. 原版最早第 25 回合召開；後續距離上次開會不足 25 回合 → 不開會。
//  3. 開會:呼叫 tallyCouncil 完成候選人與 AI 選票，建立 PendingCouncilVote 等待真人三選一。
//  4. RespondToCouncilVote 加入真人選票後才檢查 2/3；AI 當選再進既有接受／拒絕流程。
//
// 2026-08-24 依 Council_Votes_／Vote_Check_ 訂正：固定關係門檻與自動玩家選票均已移除。
func (s *GameSession) advanceCouncil() {
	s.LastCouncilNotice = nil
	if s.DisableEvents || s.Victory.Over || s.PendingCouncilElection != nil || s.PendingCouncilVote != nil {
		return
	}
	if !s.councilEligible() {
		return
	}
	if s.Turn < gamedata.CouncilFirstMeetingTurn {
		return
	}
	if s.CouncilMeetings > 0 && s.Turn-s.lastCouncilTurn < gamedata.CouncilMeetingInterval {
		return
	}

	tally := s.tallyCouncil()
	s.CouncilMeetings++
	s.lastCouncilTurn = s.Turn
	if !tally.valid {
		s.LastCouncilNotice = &CouncilNotice{Kind: CouncilNoticeInsufficientCandidates, Meeting: s.CouncilMeetings}
		return
	}
	rows := make([]CouncilVoteRow, 0, len(tally.rows))
	for _, r := range tally.rows {
		rows = append(rows, CouncilVoteRow{Name: s.councilDisplayName(r.idx), IsPlayer: r.idx == -1,
			BaseVotes: r.baseVotes, IsCandidate: r.isCandidate,
			Abstained: r.votedForIdx == councilVoteAbstain, VotedFor: s.councilDisplayName(r.votedForIdx)})
	}
	s.PendingCouncilVote = &CouncilVotePending{Turn: s.Turn, CandidateIdx: tally.candIdx,
		CandidateName:  [2]string{s.councilDisplayName(tally.candIdx[0]), s.councilDisplayName(tally.candIdx[1])},
		CandidateVotes: tally.candVotes, TotalVotes: tally.total,
		PlayerBaseVotes: gamedata.CouncilVotes(s.playerPopulationTotal()), Rows: rows}
	s.LastCouncilNotice = &CouncilNotice{Kind: CouncilNoticeVoteRequested, Meeting: s.CouncilMeetings,
		CandidateIdx: tally.candIdx, CandidateName: s.PendingCouncilVote.CandidateName,
		Votes: tally.candVotes, TotalVotes: tally.total}
}

// RespondToCouncilVote 完成原版 sub_1633C 的三選一；choice 0/1 投候選人，2 表示棄權。
func (s *GameSession) RespondToCouncilVote(choice int) {
	s.recordPlayerCommand(PlayerCommand{Name: CmdVoteCouncil, Args: []int{choice}})
	p := s.PendingCouncilVote
	if p == nil || choice < 0 || choice > 2 {
		return
	}
	if choice < 2 {
		p.CandidateVotes[choice] += p.PlayerBaseVotes
		s.CouncilLastVotes[0] = p.CandidateIdx[choice]
	} else {
		s.CouncilLastVotes[0] = councilVoteAbstain
	}
	s.PendingCouncilVote = nil
	for c := 0; c < 2; c++ {
		if !engine.CheckHighCouncil(p.CandidateVotes[c], p.TotalVotes) {
			continue
		}
		if p.CandidateIdx[c] == -1 {
			s.Victory = VictoryState{Over: true, Reason: engine.VictoryHighCouncil, Winner: "player", Turn: p.Turn}
			s.LastCouncilNotice = &CouncilNotice{Kind: CouncilNoticePlayerElected, Meeting: s.CouncilMeetings,
				CandidateIdx: p.CandidateIdx, CandidateName: p.CandidateName,
				Votes: p.CandidateVotes, TotalVotes: p.TotalVotes, WinnerSlot: c}
		} else {
			s.PendingCouncilElection = &CouncilElection{Turn: p.Turn, PlayerVotes: p.PlayerBaseVotes, EnemyVotes: p.CandidateVotes[c], TotalVotes: p.TotalVotes, EnemyName: p.CandidateName[c]}
			s.LastCouncilNotice = &CouncilNotice{Kind: CouncilNoticeEnemyElectedPending, Meeting: s.CouncilMeetings,
				CandidateIdx: p.CandidateIdx, CandidateName: p.CandidateName,
				Votes: p.CandidateVotes, TotalVotes: p.TotalVotes, WinnerSlot: c}
		}
		return
	}
	s.LastCouncilNotice = &CouncilNotice{Kind: CouncilNoticeNoMajority, Meeting: s.CouncilMeetings,
		CandidateIdx: p.CandidateIdx, CandidateName: p.CandidateName,
		Votes: p.CandidateVotes, TotalVotes: p.TotalVotes}
}

// CouncilStatus 是議會目前狀態的唯讀快照,供 UI 呈現用(cmd/moo2 是 package main,無法直接讀
// GameSession 的未匯出欄位/方法如 councilEligible,故提供這個匯出方法統一取值)。PlayerVotes/
// EnemyVotes/TotalVotes 是「若這回合真的開會,票數會是多少」的即時試算,不代表本回合一定會
// 開會——是否真的開會/流會/分出勝負以 advanceCouncil 每回合的結算為準,這裡只是唯讀快照。
//
// ⚠ EnemyVotes/EnemyName(Pending==nil 時)是「全體 AI 陣營合計」的簡化顯示,不是 advanceCouncil
// 逐帝國分算 2/3 勝負的真實依據——合計數字只適合當粗略摘要,不能反推「哪個 AI 快贏了」。要看
// 逐帝國票數與搖擺票去向,改用 CouncilBreakdown()(2026-07-12 新增,議會畫面已改用它攤開明細,
// 見 interactive.go council())。CouncilStatus 保留只為向後相容/Pending 摘要。Pending!=nil 時
// EnemyName/EnemyVotes 改用 CouncilElection 記錄的真實當選 AI(advanceCouncil 寫入),準確。
type CouncilStatus struct {
	Eligible    bool // 議會目前是否已成立(councilEligible)
	PlayerVotes int
	EnemyVotes  int
	TotalVotes  int
	EnemyName   string
	Meetings    int              // 已召開過的屆數
	Pending     *CouncilElection // 非 nil = 有待玩家回應的選舉結果
	PendingVote *CouncilVotePending
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
		EnemyName: enemyName, Meetings: s.CouncilMeetings, Pending: s.PendingCouncilElection, PendingVote: s.PendingCouncilVote,
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
	if p := s.PendingCouncilVote; p != nil {
		return CouncilBreakdown{Valid: true, Rows: p.Rows, Candidates: p.CandidateName, CandVotes: p.CandidateVotes,
			Total: p.TotalVotes, Threshold: (p.TotalVotes*2 + 2) / 3}
	}
	// Vote_Check_ 有亂數；沒有正在進行的選舉時不得為了畫面預覽提前擲骰。
	return CouncilBreakdown{}
}

// VictoryReasonLabel 是 victoryReasonLabel 的匯出版本,供 cmd/moo2 顯示中文勝利路徑描述用。
func VictoryReasonLabel(r engine.VictoryCondition) string {
	return victoryReasonLabel(r)
}

// RespondToCouncilElection 是玩家對「非玩家當選」的回應(手冊:「there's no way the council
// can force you to accept a decision you don't agree with」)。
//
//	accept=true  → 接受落敗,遊戲結束(Victory.Winner=當選 AI 名稱)。
//	accept=false → 拒絕接受,不結束遊戲,清空待決狀態,下一屆 25 回合後再開會。
//
// PendingCouncilElection==nil 時呼叫視為無操作(沒有待決選舉可回應)。
func (s *GameSession) RespondToCouncilElection(accept bool) {
	s.recordPlayerCommand(PlayerCommand{Name: CmdRespondCouncil, Args: []int{boolInt(accept)}})
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
//   - 玩家 0 殖民地、**且只剩一個** AI 對手存活 → 該 AI 以 VictoryExtermination「勝利」。
//
// ⚠ 2026-08-06 訂正:這段註解原本寫著「remake 沒有任何機制會讓 PlayerColonies 完全清空,
// 故這個分支現況下不可達」——**那個斷言已經過期**。超新星事件(events_persistent.go,
// 手冊 p.181:「all colonies are destroyed」)會摧毀整個星系的殖民地,玩家確實可能歸零。
//
// 而且 CheckExtermination 是「只剩一方存活」的判定:玩家死光但還有三個 AI 活著時它回 false,
// 遊戲會就這樣繼續跑下去,玩家永遠沒有殖民地也不會結束。玩家戰敗因此由
// advancePlayerDefeat 單獨判定,見該函式。
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

// advancePlayerDefeat 判定玩家戰敗:一座殖民地都不剩。
//
// 手冊 p.184 講計分時明確提到「If an empire is eliminated by a random event or by Antarans」
// ——**帝國被隨機事件消滅是原版就有的概念**,不是 remake 自創的失敗條件。remake 這一側的
// 觸發來源目前是超新星(手冊 p.181:「all of the system's inhabitants are killed and all
// colonies are destroyed」)。
//
// 為什麼不靠 advanceConquestVictory:那條走的是 CheckExtermination(只剩一方存活),
// 玩家死光但還有多個 AI 活著時它不成立,遊戲會無聲地繼續跑——這是 2026-08-06 的 400 回合
// 探針實際跑出來的狀態(結束時玩家 0 殖民地、BC 0,但遊戲沒結束)。
//
// 只在「有 AI 對手」的對局判定,避免測試/工具建構的空 session 意外判定戰敗。
func (s *GameSession) advancePlayerDefeat() {
	if s.Victory.Over || len(s.AIPlayers) == 0 {
		return
	}
	if len(s.PlayerColonies) > 0 {
		return
	}
	// 勝者取目前殖民地最多的 AI(手冊沒規定多方混戰時誰算贏家,取實力最強者是自然選擇)。
	best, bestN := -1, -1
	for i, a := range s.AIPlayers {
		if n := len(a.Colonies); n > bestN {
			best, bestN = i, n
		}
	}
	winner := "對手帝國"
	if best >= 0 && bestN > 0 {
		winner = s.AIPlayers[best].Name
	}
	s.Victory = VictoryState{Over: true, Reason: engine.VictoryExtermination, Winner: winner, Turn: s.Turn}
}
