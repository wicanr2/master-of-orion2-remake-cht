package shell

import (
	"fmt"

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

// councilMinExtantRacesOverride 是本 remake 目前的議會成立「存續種族數」門檻,取代
// gamedata.CouncilMinExtantRaces 手冊字面值 3。
//
// 原因(資料模型限制,非手冊規則變更):shell.GameSession.AIPlayers 目前固定只有 1 個 AI 對手
// (見 NewDemoSession),場上存續帝國數上限就是「玩家 + 1 AI」= 2,若照搬手冊字面值 3,
// 議會在本 remake 現況下永遠無法成立——議會選舉這條勝利路徑會變成死路徑,違背本輪任務
// 「幫遊戲補上可達成的勝利條件」的目的。因此改用 2 作為現況下的可達門檻,並在
// docs/HONEST-STATUS.md、docs/tech 議會章節誠實標注這是資料模型限制下的近似,不是手冊原意。
// 未來若擴充 AIPlayers 支援多個 AI 對手,應把這裡改回 gamedata.CouncilMinExtantRaces(3)。
const councilMinExtantRacesOverride = 2

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

// councilEligible 判定議會這回合是否應該存在(見 councilMinExtantRacesOverride 說明本
// remake 對手冊字面種族數門檻的覆寫)。
func (s *GameSession) councilEligible() bool {
	settled, total := s.settledStarFraction()
	if total <= 0 {
		return false
	}
	return settled*2 >= total && s.extantRaceCount() >= councilMinExtantRacesOverride
}

// playerPopulationTotal/aiPopulationTotal 回傳各自帝國殖民地人口加總,做為
// gamedata.CouncilVotes 的輸入(手冊:票數依人口規模決定,見 gamedata/council.go)。
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

// advanceCouncil 是 EndTurn 每回合呼叫的議會選舉狀態機:
//  1. 遊戲已分出勝負、或議會尚未成立(councilEligible=false)、或上一屆選舉玩家還沒回應
//     (PendingCouncilElection!=nil)→ 不開會。
//  2. 距離上次開會不足 councilInterval 回合 → 不開會(首次成立後立刻召開第一屆,不用等)。
//  3. 開會:雙方票數 = gamedata.CouncilVotes(人口),依 engine.CheckHighCouncil 判定是否有
//     一方達 2/3 多數。
//     - 玩家達標 → 立即勝利(手冊:當選者若是玩家,遊戲直接結束,不需要「接受」這個步驟,
//     那個步驟只為了「當選者不是你、議會無法強迫你接受」這個情境存在)。
//     - AI 達標 → 記錄 PendingCouncilElection,等待 RespondToCouncilElection。
//     - 兩者皆未達標 → 流會,下一屆再開(手冊描述議會確實會反覆召開,見 diplo.tsv 台詞)。
//
// 本 remake 資料模型固定只有玩家 + 1 個 AI 對手(見 councilMinExtantRacesOverride),因此
// 「兩位候選人由票數最高者出線」與「其餘種族依外交關係投票」這兩條規則沒有第三方可套用,
// 直接把僅有的兩個帝國當成僅有的兩位候選人、各自的票就是自己的人口票(見
// gamedata/council.go 檔尾 TODO 說明)。
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

	pv := gamedata.CouncilVotes(s.playerPopulationTotal())
	ev := gamedata.CouncilVotes(s.aiPopulationTotal())
	total := pv + ev
	s.CouncilMeetings++
	s.lastCouncilTurn = s.Turn

	enemyName := "對手"
	if len(s.AIPlayers) > 0 {
		enemyName = s.AIPlayers[0].Name
	}

	switch {
	case engine.CheckHighCouncil(pv, total):
		s.Victory = VictoryState{Over: true, Reason: engine.VictoryHighCouncil, Winner: "player", Turn: s.Turn}
		s.LastCouncil = fmt.Sprintf("銀河議會第 %d 屆選舉:你以 %d/%d 票(達2/3多數)當選銀河領袖!",
			s.CouncilMeetings, pv, total)
	case engine.CheckHighCouncil(ev, total):
		s.PendingCouncilElection = &CouncilElection{Turn: s.Turn, PlayerVotes: pv, EnemyVotes: ev, TotalVotes: total, EnemyName: enemyName}
		s.LastCouncil = fmt.Sprintf("銀河議會第 %d 屆選舉:%s 以 %d/%d 票(達2/3多數)當選銀河領袖,尚待你回應是否接受",
			s.CouncilMeetings, enemyName, ev, total)
	default:
		s.LastCouncil = fmt.Sprintf("銀河議會第 %d 屆選舉:無人達到2/3多數,流會(我方 %d 票／%s %d 票)",
			s.CouncilMeetings, pv, enemyName, ev)
	}
}

// CouncilStatus 是議會目前狀態的唯讀快照,供 UI 呈現用(cmd/moo2 是 package main,無法直接讀
// GameSession 的未匯出欄位/方法如 councilEligible,故提供這個匯出方法統一取值)。PlayerVotes/
// EnemyVotes/TotalVotes 是「若這回合真的開會,票數會是多少」的即時試算,不代表本回合一定會
// 開會——是否真的開會/流會/分出勝負以 advanceCouncil 每回合的結算為準,這裡只是唯讀快照。
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
	if len(s.AIPlayers) > 0 {
		enemyName = s.AIPlayers[0].Name
	}
	pv := gamedata.CouncilVotes(s.playerPopulationTotal())
	ev := gamedata.CouncilVotes(s.aiPopulationTotal())
	return CouncilStatus{
		Eligible: s.councilEligible(), PlayerVotes: pv, EnemyVotes: ev, TotalVotes: pv + ev,
		EnemyName: enemyName, Meetings: s.CouncilMeetings, Pending: s.PendingCouncilElection,
		Victory: s.Victory,
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
