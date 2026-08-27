package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// StatusBroadcastState 保存 GNN 29..35 的長壽命去重與待播佇列。LastEventReport 是顯示暫態，
// 不能拿它判斷是否播過；否則存讀檔或同回合多事件都會重播／遺失。
type StatusBroadcastState struct {
	Queue           []EventReport `json:"queue,omitempty"`
	EmpireAlive     []bool        `json:"empireAlive,omitempty"`
	GrowthStage     int           `json:"growthStage,omitempty"`
	OrionStars      []int         `json:"orionStars,omitempty"`
	AntaranReported bool          `json:"antaranReported,omitempty"`
}

func (s *GameSession) queueStatusBroadcast(report EventReport) {
	s.StatusBroadcast.Queue = append(s.StatusBroadcast.Queue, report)
}

func (s *GameSession) publishNextStatusBroadcast() {
	if s.LastEventReport != nil || len(s.StatusBroadcast.Queue) == 0 {
		return
	}
	report := s.StatusBroadcast.Queue[0]
	s.StatusBroadcast.Queue = append([]EventReport(nil), s.StatusBroadcast.Queue[1:]...)
	s.LastEvent, s.LastEventReport = report.Message, &report
	s.broadcastEventReport(&report)
}

func (s *GameSession) statusTargetReport(id int, target eventEmpireTarget, message, messageEN string) EventReport {
	ev := gamedata.RandomEventByID(id)
	name := fmt.Sprintf("狀態播報 %d", id)
	if ev != nil {
		name = ev.Name
	}
	return EventReport{EventID: id, Name: name, NameEN: eventNameEN(id), Good: true,
		Message: message, MessageEN: messageEN, TargetKind: target.kind.String(), TargetIndex: target.index,
		TargetName: s.eventEmpireTargetName(target)}
}

func (s *GameSession) ensureStatusEmpireBaseline() {
	targets := s.eventEmpireTargets()
	if len(s.StatusBroadcast.EmpireAlive) == len(targets) {
		return
	}
	s.StatusBroadcast.EmpireAlive = make([]bool, len(targets))
	for i := range targets {
		s.StatusBroadcast.EmpireAlive[i] = targets[i].alive
	}
}

func (s *GameSession) detectEmpireEliminationBroadcasts() {
	targets := s.eventEmpireTargets()
	if len(s.StatusBroadcast.EmpireAlive) != len(targets) {
		s.ensureStatusEmpireBaseline()
		return
	}
	for i := range targets {
		if s.StatusBroadcast.EmpireAlive[i] && !targets[i].alive {
			s.queueStatusBroadcast(s.statusTargetReport(29, targets[i], "", ""))
		}
		s.StatusBroadcast.EmpireAlive[i] = targets[i].alive
	}
}

func uniqueActiveStars(stars []int, coloniesActive func(int) bool) int {
	seen := make(map[int]struct{})
	for i, star := range stars {
		if star >= 0 && coloniesActive(i) {
			seen[star] = struct{}{}
		}
	}
	return len(seen)
}

func (s *GameSession) statusColonizedStarCounts() ([]eventEmpireTarget, []int) {
	targets := s.eventEmpireTargets()
	counts := make([]int, len(targets))
	for i, target := range targets {
		switch target.kind {
		case eventEmpireSeat:
			v := s.Seats[target.index]
			counts[i] = uniqueActiveStars(v.PlayerColonyStars, func(j int) bool {
				return j < len(v.PlayerColonies) && v.PlayerColonies[j].Population > 0
			})
		case eventEmpireAI:
			a := s.AIPlayers[target.index]
			counts[i] = uniqueActiveStars(a.ColonyStars, func(j int) bool {
				return j < len(a.Colonies) && a.Colonies[j].Population > 0
			})
		default:
			counts[i] = uniqueActiveStars(s.PlayerColonyStars, func(j int) bool {
				return j < len(s.PlayerColonies) && s.PlayerColonies[j].Population > 0
			})
		}
	}
	return targets, counts
}

func originalExpansionBroadcastDue(stars, stage, maxColonized int, councilFormed bool) bool {
	if stage < 0 || stage > 2 || stars < 1 {
		return false
	}
	half := stars / 2
	threshold := (stage + 2) * half / 4
	if maxColonized >= threshold {
		return true
	}
	return stage == 2 && !councilFormed && maxColonized >= half-2 && maxColonized < half
}

// advanceStatusGrowthAndRanking 對應 sub_23563 與其後事件 31 分支；只在 Determine_Event
// 確實被呼叫的時點使用，不是每回合固定新聞。
func (s *GameSession) advanceStatusGrowthAndRanking() {
	targets, counts := s.statusColonizedStarCounts()
	maxIndex, maxCount := -1, 0
	for i, count := range counts {
		if count > maxCount {
			maxIndex, maxCount = i, count
		}
	}
	stage := s.StatusBroadcast.GrowthStage
	if maxIndex >= 0 && originalExpansionBroadcastDue(len(s.Stars), stage, maxCount,
		s.CouncilMeetings > 0 || s.PendingCouncilElection != nil) {
		target := targets[maxIndex]
		name := s.eventEmpireTargetName(target)
		s.StatusBroadcast.GrowthStage++
		report := s.statusTargetReport(30, target,
			fmt.Sprintf("%s 的版圖已擴張至 %d 個殖民星系，成為銀河第 %d 階段的強權", name, maxCount, stage+1),
			fmt.Sprintf("%s now controls %d colonized systems and has reached expansion stage %d.", name, maxCount, stage+1))
		s.queueStatusBroadcast(report)
	}
	if s.Turn-1 <= 50 || s.CouncilMeetings > 0 || s.PendingCouncilElection != nil || s.eventRoll(40) != 1 {
		return
	}
	category := s.eventRand.Intn(4)
	labelsZH := [...]string{"艦隊實力", "科技水準", "人口", "建築"}
	labelsEN := [...]string{"fleet strength", "technology", "population", "buildings"}
	s.queueStatusBroadcast(EventReport{EventID: 31, Name: "排行榜播報", NameEN: eventNameEN(31), Good: true,
		Message:    fmt.Sprintf("銀河新聞網公布最新帝國排行榜：本期評比項目為%s", labelsZH[category]),
		MessageEN:  fmt.Sprintf("Galactic News Network published new empire rankings for %s.", labelsEN[category]),
		TargetKind: "galaxy", TargetIndex: -1, TargetName: labelsZH[category]})
}

func intSliceContains(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (s *GameSession) queueOrionDiscoveryBroadcast(target eventEmpireTarget, star int) {
	if star < 0 || intSliceContains(s.StatusBroadcast.OrionStars, star) {
		return
	}
	guarded := false
	for _, monster := range s.Monsters {
		if monster.StarIndex == star && monster.Kind == gamedata.MonsterGuardian {
			guarded = true
			break
		}
	}
	if !guarded {
		return
	}
	s.StatusBroadcast.OrionStars = append(s.StatusBroadcast.OrionStars, star)
	name := s.eventEmpireTargetName(target)
	s.queueStatusBroadcast(s.statusTargetReport(32, target,
		fmt.Sprintf("%s 的艦隊發現了由守護者盤據的獵戶座 %s 星系", name, s.starName(star)),
		fmt.Sprintf("A fleet of %s discovered Orion in the %s system, guarded by the Guardian.", name, s.starNameEN(star))))
}

func (s *GameSession) queueAntaranDefeatBroadcast() {
	if s.StatusBroadcast.AntaranReported {
		return
	}
	s.StatusBroadcast.AntaranReported = true
	target := eventEmpireTarget{kind: eventEmpirePlayer, index: 0, alive: true}
	if s.HotseatEnabled() {
		target = eventEmpireTarget{kind: eventEmpireSeat, index: s.ActiveSeat, alive: true}
	}
	name := s.eventEmpireTargetName(target)
	s.queueStatusBroadcast(s.statusTargetReport(33, target,
		fmt.Sprintf("%s 已攻陷安塔蘭母星，終結來自異次元的威脅", name),
		fmt.Sprintf("%s conquered the Antaran homeworld and ended the threat from another dimension.", name)))
}

func (s *GameSession) queueRebellionBroadcasts(results []RebellionResult) {
	for _, result := range results {
		if !result.ColonyLost {
			continue
		}
		target := eventEmpireTarget{kind: eventEmpireAI, index: result.RevertedToAI, alive: result.RevertedToAI >= 0}
		name := "獨立叛軍"
		if result.RevertedToAI >= 0 && result.RevertedToAI < len(s.AIPlayers) {
			name = s.eventEmpireTargetName(target)
		}
		report := s.statusTargetReport(35, target,
			fmt.Sprintf("%s 的叛軍奪取 %s，殖民地現由%s控制", result.ColonyName, result.ColonyName, name),
			fmt.Sprintf("Rebels seized %s; the colony is now controlled by %s.", result.ColonyName, name))
		report.TargetName = name
		s.queueStatusBroadcast(report)
	}
}
