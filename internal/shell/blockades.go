package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func ownerSlotFromColonies(colonies []engine.ColonyState) (int, bool) {
	for _, c := range colonies {
		if c.OwnerRaceSlotKnown && c.OwnerRaceSlot >= 0 && c.OwnerRaceSlot < 8 {
			return c.OwnerRaceSlot, true
		}
	}
	return 0, false
}

func (s *GameSession) aiIndexByPopulationSlot(slot int) (int, bool) {
	for i := range s.AIPlayers {
		if s.AIPlayers[i].PopulationRaceSlotKnown && s.AIPlayers[i].PopulationRaceSlot == slot {
			return i, true
		}
	}
	return 0, false
}

func (s *GameSession) originalPolicyBetweenSlots(owner, target int) (gamedata.ForeignPolicy, bool) {
	ownerAI, ownerIsAI := s.aiIndexByPopulationSlot(owner)
	targetAI, targetIsAI := s.aiIndexByPopulationSlot(target)
	if ownerIsAI && targetIsAI {
		if ownerAI < len(s.AIPolicies) && targetAI < len(s.AIPolicies[ownerAI]) {
			return s.AIPolicies[ownerAI][targetAI], true
		}
		return 0, false
	}
	if ownerIsAI != targetIsAI {
		aiIndex := ownerAI
		if targetIsAI {
			aiIndex = targetAI
		}
		if aiIndex >= 0 && aiIndex < len(s.AIPlayers) {
			return s.AIPlayers[aiIndex].Treaty.FormalPolicy, true
		}
	}
	return 0, false
}

type originalBlockadeFleet struct {
	owner, star int
}

// recomputeOriginalBlockades 對映 sub_E5097：先清表，再由已抵達艦隊重建。
func (s *GameSession) recomputeOriginalBlockades() {
	for i := range s.Stars {
		s.Stars[i].BlockadedMask = 0
		s.Stars[i].BlockadedBy = [8]uint8{}
	}
	occupied := make([]uint8, len(s.Stars))
	addColonies := func(slot int, known bool, stars []int) {
		if !known || slot < 0 || slot >= 8 {
			return
		}
		for _, star := range stars {
			if star >= 0 && star < len(occupied) {
				occupied[star] |= 1 << slot
			}
		}
	}
	playerSlot, playerKnown := ownerSlotFromColonies(s.PlayerColonies)
	addColonies(playerSlot, playerKnown, s.PlayerColonyStars)
	for i := range s.Seats {
		if i == s.ActiveSeat {
			continue
		}
		slot, known := ownerSlotFromColonies(s.Seats[i].PlayerColonies)
		addColonies(slot, known, s.Seats[i].PlayerColonyStars)
	}
	for i := range s.AIPlayers {
		addColonies(s.AIPlayers[i].PopulationRaceSlot, s.AIPlayers[i].PopulationRaceSlotKnown, s.AIPlayers[i].ColonyStars)
	}

	var fleets []originalBlockadeFleet
	addFleets := func(slot int, known bool, fs []Fleet) {
		if !known || slot < 0 || slot >= 8 {
			return
		}
		for _, fleet := range fs {
			if fleet.ETA == 0 && fleet.AtStar >= 0 && fleet.AtStar < len(s.Stars) && len(fleet.Ships) > 0 {
				fleets = append(fleets, originalBlockadeFleet{owner: slot, star: fleet.AtStar})
			}
		}
	}
	addFleets(playerSlot, playerKnown, s.Fleets)
	for i := range s.Seats {
		if i == s.ActiveSeat {
			continue
		}
		slot, known := ownerSlotFromColonies(s.Seats[i].PlayerColonies)
		addFleets(slot, known, s.Seats[i].Fleets)
	}
	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		if a.PopulationRaceSlotKnown && a.FleetPosSet && a.FleetETA == 0 && a.FleetStar >= 0 &&
			a.FleetStar < len(s.Stars) && len(a.Ships) > 0 {
			fleets = append(fleets, originalBlockadeFleet{owner: a.PopulationRaceSlot, star: a.FleetStar})
		}
	}
	for _, fleet := range fleets {
		for target := 0; target < 8; target++ {
			if target == fleet.owner || occupied[fleet.star]&(1<<target) == 0 {
				continue
			}
			policy, known := s.originalPolicyBetweenSlots(fleet.owner, target)
			if !known || policy < gamedata.DIPLO_LIMITED_WAR {
				continue
			}
			s.Stars[fleet.star].BlockadedMask |= 1 << target
			s.Stars[fleet.star].BlockadedBy[target] |= 1 << fleet.owner
			// Change_Relations_ reason raw 7：只有「AI 被目前真人封鎖」會進
			// target personality 100 的 +0x6BF 積怨特例；戰時一般 +0x617 分數早退。
			if fleet.owner == playerSlot && playerKnown {
				if aiIndex, ok := s.aiIndexByPopulationSlot(target); ok {
					// Compute_Blockades 無條件先呼叫 Random_(5)；舊存檔即使尚無
					// typed 關係分數，也要保留相同亂數消費順序。
					base := -(s.diplomacyGrowthRandForTurn().Intn(5) + 1)
					if !s.AIPlayers[aiIndex].OriginalRelationKnown {
						continue
					}
					if delta, valid := gamedata.OriginalWarBlockadeGrievance(gamedata.OriginalRelationChangeInput{
						CurrentRaw: s.AIPlayers[aiIndex].OriginalRelationRaw, BaseDelta: base,
						ActorGovernment:   originalAIRelationGovernment(s.AIPlayers[aiIndex]),
						TargetCharismatic: s.RaceCharismatic(), Policy: policy,
					}); valid {
						s.AIPlayers[aiIndex].OriginalBlockadeGrievanceRaw = int(int16(
							s.AIPlayers[aiIndex].OriginalBlockadeGrievanceRaw + delta))
					}
				}
			}
		}
	}
	// sub_E5097 對 owner>=8 不查外交，直接封鎖同星全部殖民者，也不填 BlockadedBy。
	for _, monster := range s.Monsters {
		if monster.TransitETA == 0 && monster.StarIndex >= 0 && monster.StarIndex < len(s.Stars) {
			s.Stars[monster.StarIndex].BlockadedMask = occupied[monster.StarIndex]
		}
	}
	for _, fleet := range s.AntaranInvasion.Pending {
		if fleet.ETA == 0 && fleet.StarIndex >= 0 && fleet.StarIndex < len(s.Stars) {
			s.Stars[fleet.StarIndex].BlockadedMask = occupied[fleet.StarIndex]
		}
	}
}
