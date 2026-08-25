package shell

import "fmt"

func originalPirateActivityStrength(freighters, difficulty int) int {
	if freighters < 0 {
		freighters = 0
	}
	var strength int
	switch difficulty {
	case 0:
		strength = freighters / 5
	case 1:
		strength = freighters * 2 / 5
	case 2:
		strength = freighters * 3 / 5
	case 4:
		strength = freighters * 4 / 5
	default: // 原版 case 3（Hard）直接保留 T；越界值亦走 raw default。
		strength = freighters
	}
	if strength < 5 {
		return 0
	}
	return strength
}

func colonyStarsContain(stars []int, star int) bool {
	for _, candidate := range stars {
		if candidate == star {
			return true
		}
	}
	return false
}

// pirateActivityConflictAtStar 對應 sub_242FC 的同星互斥表。
func (s *GameSession) pirateActivityConflictAtStar(star int) bool {
	for i := range s.PersistentEvents {
		e := &s.PersistentEvents[i]
		switch e.Kind {
		case PersistentComet, PersistentSupernova, PersistentStasis, PersistentPirateActivity:
			if e.StarIndex == star {
				return true
			}
		case PersistentPlague, PersistentPopulationBoom:
			if s.PlanetStar(e.PlanetIndex) == star {
				return true
			}
		}
	}
	return false
}

func (s *GameSession) pirateActivityFreighterTotal(star int) int {
	total := 0
	if colonyStarsContain(s.PlayerColonyStars, star) {
		total += s.Player.ActiveFreighters
	}
	if s.HotseatEnabled() {
		for i := range s.Seats {
			if i == s.ActiveSeat || !colonyStarsContain(s.Seats[i].PlayerColonyStars, star) {
				continue
			}
			total += s.Seats[i].Player.ActiveFreighters
		}
	}
	for i := range s.AIPlayers {
		if colonyStarsContain(s.AIPlayers[i].ColonyStars, star) {
			total += s.AIPlayers[i].Player.ActiveFreighters
		}
	}
	return total
}

func (s *GameSession) pirateActivityDestroyFreighters(star int) int {
	lost := 0
	if colonyStarsContain(s.PlayerColonyStars, star) && s.Player.ActiveFreighters > 0 {
		s.Player.ActiveFreighters--
		lost++
	}
	if s.HotseatEnabled() {
		for i := range s.Seats {
			if i == s.ActiveSeat || !colonyStarsContain(s.Seats[i].PlayerColonyStars, star) || s.Seats[i].Player.ActiveFreighters <= 0 {
				continue
			}
			s.Seats[i].Player.ActiveFreighters--
			lost++
		}
	}
	for i := range s.AIPlayers {
		if colonyStarsContain(s.AIPlayers[i].ColonyStars, star) && s.AIPlayers[i].Player.ActiveFreighters > 0 {
			s.AIPlayers[i].Player.ActiveFreighters--
			lost++
		}
	}
	return lost
}

func (s *GameSession) pickPirateActivityStar(stars []int) (int, bool) {
	chosen, candidates := -1, 0
	seen := make(map[int]bool)
	for _, star := range stars {
		if star < 0 || star >= len(s.Stars) || seen[star] || s.pirateActivityConflictAtStar(star) {
			continue
		}
		seen[star] = true
		candidates++
		if s.eventRand.Intn(candidates) == 0 {
			chosen = star
		}
	}
	return chosen, chosen >= 0
}

func (s *GameSession) appendPirateActivity(star int) (PersistentEvent, bool) {
	strength := originalPirateActivityStrength(s.pirateActivityFreighterTotal(star), s.Difficulty)
	if strength == 0 {
		return PersistentEvent{}, false
	}
	e := PersistentEvent{Kind: PersistentPirateActivity, StarIndex: star,
		Strength: strength, InitialStrength: strength}
	s.PersistentEvents = append(s.PersistentEvents, e)
	return e, true
}

func (s *GameSession) startPlayerPirateActivity() (string, bool) {
	star, ok := s.pickPirateActivityStar(s.PlayerColonyStars)
	if !ok {
		return "", false
	}
	e, ok := s.appendPirateActivity(star)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%s 星系爆發海盜活動，威脅強度 %d；停泊艦隊已開始清剿", s.starName(star), e.Strength), true
}

func (s *GameSession) startAIPirateActivity(aiIndex int) (string, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return "", false
	}
	star, ok := s.pickPirateActivityStar(s.AIPlayers[aiIndex].ColonyStars)
	if !ok {
		return "", false
	}
	e, ok := s.appendPirateActivity(star)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%s的 %s 星系爆發海盜活動，威脅強度 %d", stripAILabel(s.AIPlayers[aiIndex].Name), s.starName(star), e.Strength), true
}

func (s *GameSession) stepPirateActivity(e *PersistentEvent) (bool, string, string) {
	if e.InitialStrength <= 0 || e.Strength <= 0 {
		return true, "海盜活動因威脅強度耗盡而結束", "Pirate activity ended after its threat strength was exhausted."
	}
	lost := 0
	chance := e.Strength * 100 / e.InitialStrength
	if s.eventRoll(100) <= chance {
		lost = s.pirateActivityDestroyFreighters(e.StarIndex)
	}
	e.Strength -= s.cometInterceptionStrength(e.StarIndex)
	if e.Strength <= 0 {
		return true, fmt.Sprintf("%s 星系的艦隊已肅清海盜；本回合 %d 艘運輸船遭毀", s.starName(e.StarIndex), lost),
			fmt.Sprintf("Ships in the %s system eliminated the pirates; %d freighters were lost this turn.", s.starNameEN(e.StarIndex), lost)
	}
	if lost > 0 {
		return false, fmt.Sprintf("%s 星系海盜摧毀 %d 艘運輸船，剩餘威脅 %d/%d", s.starName(e.StarIndex), lost, e.Strength, e.InitialStrength),
			fmt.Sprintf("Pirates in the %s system destroyed %d freighters; threat remaining %d/%d.", s.starNameEN(e.StarIndex), lost, e.Strength, e.InitialStrength)
	}
	return false, "", ""
}
