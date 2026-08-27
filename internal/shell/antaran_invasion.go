package shell

import (
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

const antaranRaidApproxETA = 3

// AntaranNoticeKind 是玩家可見安塔蘭通知的穩定種類；句型由 UI 外部 catalog 決定。
type AntaranNoticeKind uint8

const (
	AntaranNoticeNone AntaranNoticeKind = iota
	AntaranNoticeLaunched
	AntaranNoticeAIEngaged
	AntaranNoticeUndefended
	AntaranNoticeBattle
)

// AntaranNotice 保存一筆安塔蘭戰略入侵的型別化玩家結果，不保存任何語言句子。
type AntaranNotice struct {
	Kind       AntaranNoticeKind
	StarName   string
	StarNameEN string
	ETA        int
	ShipsLost  int
	Repelled   bool
}

// AntaranRaidFleet 是原版 owner 8 出征艦隊在 remake ETA 模型中的可存檔投影。
type AntaranRaidFleet struct {
	TargetKind  string
	TargetIndex int
	StarIndex   int
	ETA         int
	Ships       [5]int
}

// AntaranInvasionState 保存原版全局資源、建艦與出征 record。
type AntaranInvasionState struct {
	Initialized       bool
	OffensiveResource int
	DefensiveResource int
	OffensiveShips    [5]int
	DefensiveShips    [5]int
	DeployedShips     [5]int
	OffensiveMax      [5]int
	DefensiveMax      [5]int
	Costs             [5]int
	Readiness         int
	Discounted        bool
	Pending           []AntaranRaidFleet
}

func (s *GameSession) initAntaranInvasionState() {
	st := &s.AntaranInvasion
	if st.Initialized {
		return
	}
	st.Initialized = true
	st.OffensiveMax = gamedata.AntaranOffensiveMax
	st.DefensiveMax = gamedata.AntaranDefensiveMax
	st.Costs = gamedata.AntaranShipCosts
}

// advanceAntares 對應 Antaran_Invasion_Check_ 的全局一次性回合入口。
func (s *GameSession) advanceAntares() {
	s.LastAntaranNotice = nil
	if s.DisableEvents || s.AntaranHomeworldConquered {
		s.AntaranInvasion.Pending = nil
		return
	}
	s.initAntaranInvasionState()
	s.advanceAntaranRaidFleets()

	elapsed := s.Turn - 1
	st := &s.AntaranInvasion
	if grant, ok := gamedata.OriginalAntaranResourcePulse(elapsed, s.techLevel(), s.Difficulty); ok {
		st.OffensiveResource += grant
		if gamedata.OriginalAntaranDefenseComplete(st.DefensiveShips) {
			st.OffensiveResource += grant
		} else {
			st.DefensiveResource += grant
		}
	}
	delta := elapsed - gamedata.OriginalAntaranTechDelay(s.techLevel())
	gamedata.OriginalAntaranBuildShips(&st.OffensiveResource, &st.OffensiveShips,
		&st.OffensiveMax, &st.Costs, true, delta, s.techLevel(), s.Difficulty, &st.Discounted)
	gamedata.OriginalAntaranBuildShips(&st.DefensiveResource, &st.DefensiveShips,
		&st.DefensiveMax, &st.Costs, false, delta, s.techLevel(), s.Difficulty, &st.Discounted)
	// sub_63FF0 無論正在建 offensive 或 defensive，都以靜態 offensive maxima
	// 判斷前三級是否達上限並將該格清零。
	for class := 0; class <= 2; class++ {
		if st.OffensiveMax[class] > 0 && st.DefensiveShips[class] == st.OffensiveMax[class] {
			st.OffensiveMax[class] = 0
		}
	}

	if !gamedata.OriginalAntaranInvasionReady(st.OffensiveShips, st.DefensiveShips,
		st.DeployedShips, st.Costs, st.OffensiveResource, st.DefensiveResource) {
		return
	}
	st.Readiness++
	if !gamedata.OriginalAntaranInvasionRollSucceeds(st.Readiness, s.eventRoll(200)) {
		return
	}
	target, star, ok := s.chooseAntaranTarget()
	if !ok {
		return
	}
	raid := AntaranRaidFleet{TargetKind: target.kind.String(), TargetIndex: target.index,
		StarIndex: star, ETA: antaranRaidApproxETA}
	for count := 0; count < 5; count++ {
		picked := -1
		for class := len(st.OffensiveShips) - 1; class >= 0; class-- {
			if st.OffensiveShips[class] > st.DeployedShips[class] {
				picked = class
				break
			}
		}
		if picked < 0 {
			break
		}
		raid.Ships[picked]++
		st.DeployedShips[picked]++
	}
	if gamedata.OriginalAntaranWeightedStrength(raid.Ships, st.Costs) == 0 {
		return
	}
	st.Pending = append(st.Pending, raid)
	st.Readiness = 0
	s.AntaresRaids++
	s.LastAntaranNotice = &AntaranNotice{Kind: AntaranNoticeLaunched,
		StarName: s.starName(star), StarNameEN: s.starNameEN(star), ETA: raid.ETA}
}

func (s *GameSession) chooseAntaranTarget() (eventEmpireTarget, int, bool) {
	targets := s.eventEmpireTargets()
	populations, eligible, lucky := make([]int, len(targets)), make([]bool, len(targets)), make([]bool, len(targets))
	for i := range targets {
		populations[i], eligible[i], lucky[i] = targets[i].population, targets[i].alive, targets[i].lucky
	}
	weights := gamedata.OriginalAntaranTargetWeights(populations, eligible, lucky, s.Difficulty)
	total := 0
	for _, weight := range weights {
		total += weight
	}
	if total <= 0 {
		return eventEmpireTarget{}, 0, false
	}
	roll := s.eventRand.Intn(total)
	pick := -1
	for i, weight := range weights {
		if roll < weight {
			pick = i
			break
		}
		roll -= weight
	}
	if pick < 0 {
		return eventEmpireTarget{}, 0, false
	}
	target := targets[pick]
	stars := s.antaranTargetStars(target)
	if len(stars) == 0 {
		return eventEmpireTarget{}, 0, false
	}
	return target, stars[s.eventRand.Intn(len(stars))], true
}

func uniqueValidStars(values []int, max int) []int {
	seen := map[int]bool{}
	out := make([]int, 0, len(values))
	for _, star := range values {
		if star < 0 || star >= max || seen[star] {
			continue
		}
		seen[star] = true
		out = append(out, star)
	}
	return out
}

func (s *GameSession) antaranTargetStars(target eventEmpireTarget) []int {
	switch target.kind {
	case eventEmpireSeat:
		if target.index >= 0 && target.index < len(s.Seats) {
			return uniqueValidStars(s.Seats[target.index].PlayerColonyStars, len(s.Stars))
		}
	case eventEmpireAI:
		if target.index >= 0 && target.index < len(s.AIPlayers) {
			return uniqueValidStars(s.AIPlayers[target.index].ColonyStars, len(s.Stars))
		}
	default:
		return uniqueValidStars(s.PlayerColonyStars, len(s.Stars))
	}
	return nil
}

func antaranRaidCombatants(ships [5]int) []combatant {
	classes := []gamedata.CombatShipClass{gamedata.SHIP_FRIGATE, gamedata.SHIP_CRUISER,
		gamedata.SHIP_BATTLESHIP, gamedata.SHIP_TITAN, gamedata.SHIP_DOOMSTAR}
	strength := []int{2, 5, 12, 30, 75}
	out := []combatant{}
	for class, n := range ships {
		for i := 0; i < n; i++ {
			v := strength[class]
			out = append(out, combatant{hp: v * 3, atk: v, def: v, wmin: v / 2,
				wmax: v, armor: v, sizeClass: classes[class], shipIdx: -1})
		}
	}
	return out
}

func (s *GameSession) advanceAntaranRaidFleets() {
	st := &s.AntaranInvasion
	kept := st.Pending[:0]
	for i := range st.Pending {
		raid := &st.Pending[i]
		raid.ETA--
		if raid.ETA > 0 {
			kept = append(kept, *raid)
			continue
		}
		survivors := s.resolveAntaranRaid(*raid)
		for class := range raid.Ships {
			destroyed := raid.Ships[class] - survivors[class]
			st.DeployedShips[class] -= destroyed
			st.OffensiveShips[class] -= destroyed
			if st.DeployedShips[class] < 0 {
				st.DeployedShips[class] = 0
			}
			if st.OffensiveShips[class] < 0 {
				st.OffensiveShips[class] = 0
			}
		}
	}
	st.Pending = kept
}

// resolveAntaranRaid 讓抵達艦隊進入既有快速戰鬥，不再直接扣 BC。owner 8 的完整 raw
// combat record 尚未閉合，安塔蘭艦屬性映射仍是明示近似。
func (s *GameSession) resolveAntaranRaid(raid AntaranRaidFleet) (survivors [5]int) {
	if raid.TargetKind == eventEmpireAI.String() {
		if raid.TargetIndex < 0 || raid.TargetIndex >= len(s.AIPlayers) {
			return raid.Ships
		}
		power := gamedata.OriginalAntaranWeightedStrength(raid.Ships, s.AntaranInvasion.Costs)
		a := &s.AIPlayers[raid.TargetIndex]
		lost := power
		if lost > a.FleetStrength {
			lost = a.FleetStrength
		}
		a.FleetStrength -= lost
		s.LastAntaranNotice = &AntaranNotice{Kind: AntaranNoticeAIEngaged,
			StarName: s.starName(raid.StarIndex), StarNameEN: s.starNameEN(raid.StarIndex)}
		return raid.Ships
	}

	active := s.ActiveSeat
	if raid.TargetKind == eventEmpireSeat.String() && raid.TargetIndex >= 0 && raid.TargetIndex < len(s.Seats) && raid.TargetIndex != active {
		if active >= 0 && active < len(s.Seats) {
			s.Seats[active] = s.saveSeat()
		}
		s.loadSeat(s.Seats[raid.TargetIndex])
		defer func() {
			notice := s.LastAntaranNotice
			s.Seats[raid.TargetIndex] = s.saveSeat()
			if active >= 0 && active < len(s.Seats) {
				s.loadSeat(s.Seats[active])
			}
			s.LastAntaranNotice = notice
		}()
	}
	fleetIdx := -1
	for i := range s.Fleets {
		if s.Fleets[i].ETA == 0 && s.Fleets[i].AtStar == raid.StarIndex {
			fleetIdx = i
			break
		}
	}
	attackers := antaranRaidCombatants(raid.Ships)
	if fleetIdx < 0 {
		s.LastAntaranNotice = &AntaranNotice{Kind: AntaranNoticeUndefended,
			StarName: s.starName(raid.StarIndex), StarNameEN: s.starNameEN(raid.StarIndex)}
		return raid.Ships
	}
	selected := s.SelectedFleet
	s.SelectedFleet = fleetIdx
	defenders := s.mkPlayerCombatants()
	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + int64(raid.StarIndex)*97 + 63192))
	for round := 0; round < 6 && len(attackers) > 0 && len(defenders) > 0; round++ {
		battleVolley(defenders, &attackers, rng)
		battleVolley(attackers, &defenders, rng)
	}
	lost := len(s.Fleet().Ships) - len(defenders)
	for i := 0; i < lost; i++ {
		s.removeWeakestShip()
	}
	s.SelectedFleet = selected
	repelled := len(attackers) == 0
	for _, ship := range attackers {
		switch ship.sizeClass {
		case gamedata.SHIP_FRIGATE:
			survivors[0]++
		case gamedata.SHIP_CRUISER:
			survivors[1]++
		case gamedata.SHIP_BATTLESHIP:
			survivors[2]++
		case gamedata.SHIP_TITAN:
			survivors[3]++
		case gamedata.SHIP_DOOMSTAR:
			survivors[4]++
		}
	}
	s.LastAntaranNotice = &AntaranNotice{Kind: AntaranNoticeBattle,
		StarName: s.starName(raid.StarIndex), StarNameEN: s.starNameEN(raid.StarIndex),
		ShipsLost: lost, Repelled: repelled}
	return survivors
}
