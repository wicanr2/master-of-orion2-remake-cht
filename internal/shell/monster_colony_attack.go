package shell

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// monsterColonyImpact 是 owner 8 怪物抵達後的玩家可見戰果。
type monsterColonyImpact struct {
	Hits, PopulationLost, BuildingsDestroyed, MarinesLost, TanksLost, BuildProgressLost int
	MonsterDestroyed, ColonyDestroyed                                                   bool
}

func monsterGroupCount(m *MonsterGuard) int {
	if m == nil {
		return 0
	}
	if m.Kind == gamedata.MonsterEel {
		return normalizeEelGroup(m)
	}
	if m.Count > 0 {
		return m.Count
	}
	return 1
}

// monsterDefenseBattle 以 remake 目前可表示的固定防禦 combatant 重建 Do_1_Combat_ 的
// 「殖民地防禦先與怪物交戰」外形。怪物完整 blueprint 未閉合，故數值是明示近似。
func (s *GameSession) monsterDefenseBattle(m *MonsterGuard, buildings map[string]bool,
	defender engine.PlayerState) bool {
	defenses := retaliationAttackers(buildings, defender, s.RuleProfile)
	if len(defenses) == 0 || m == nil || m.Structure <= 0 {
		return false
	}
	monster := []combatant{{hp: m.Structure, maxHP: m.Structure, armor: m.Armor, def: 50}}
	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + int64(m.StarIndex)*7919 + int64(m.Kind)*131))
	for round := 0; round < 6 && len(monster) > 0; round++ {
		battleVolley(defenses, &monster, rng)
	}
	if len(monster) == 0 || monster[0].hp <= 0 {
		m.Structure = 0
		return true
	}
	m.Structure = monster[0].hp
	m.Armor = monster[0].armor
	return false
}

// monsterStrategicBombardHits 對應 Strategic_Bombardment_ 的固定三外圈與 /40。
// 每種怪物的精確 weapon mounts 尚未閉合，逐體單發傷害取現有手冊資料。
func (s *GameSession) monsterStrategicBombardHits(m *MonsterGuard, shield int) int {
	stats, ok := gamedata.MonsterStatsFor(m.Kind)
	if !ok || stats.DamageMax <= 0 {
		return 0
	}
	s.eventRandForTest()
	total, count := 0, monsterGroupCount(m)
	for round := 0; round < 3; round++ {
		for ship := 0; ship < count; ship++ {
			damage := stats.DamageMin
			if stats.DamageMax > stats.DamageMin {
				damage += s.eventRand.Intn(stats.DamageMax - stats.DamageMin + 1)
			}
			total += gamedata.PlanetaryShieldedDamage(damage, shield)
		}
	}
	return gamedata.StrategicBombardmentHitsFromDamage(total)
}

func (s *GameSession) resolveMonsterDamage(colony *engine.ColonyState, buildings map[string]bool,
	marines, tanks int, defender engine.PlayerState, highG bool, hits int) gamedata.StrategicColonyDamageResult {
	s.eventRandForTest()
	result := gamedata.ResolveStrategicColonyDamage(gamedata.StrategicColonyDamageState{
		Population: colony.Population, LastPopulationPoints: colony.BombardmentLastPopulationPoints,
		Marines: marines, Tanks: tanks, BuildProgress: colony.BombardmentBuildProgress,
		RawBuildingIDs: originalColonyBuildingIDs(buildings),
		MarineHitCost:  gamedata.GroundMarineHitsToKill(highG, hasPoweredArmorFor(defender)),
		TankHitCost:    tankHitsToKillFor(defender, highG),
		BuildingHitCost: gamedata.GroundPlanetHitsPerBuilding +
			s.RuleProfile.BombardmentBuildingBonusHits,
	}, hits, s.eventRand.Intn)
	colony.Population = result.State.Population
	colony.BombardmentLastPopulationPoints = result.State.LastPopulationPoints
	colony.BombardmentBuildProgress = result.State.BuildProgress
	normalizeColonyJobsAfterPopulationLoss(colony)
	return result
}

func impactFromMonsterDamage(hits int, result gamedata.StrategicColonyDamageResult) monsterColonyImpact {
	return monsterColonyImpact{Hits: hits, PopulationLost: result.PopulationLost,
		BuildingsDestroyed: len(result.DestroyedBuildingIDs), MarinesLost: result.MarinesLost,
		TanksLost: result.TanksLost, BuildProgressLost: result.BuildProgressLost,
		ColonyDestroyed: result.ColonyDestroyed}
}

func (s *GameSession) attackPlayerColonyWithMonster(m *MonsterGuard) monsterColonyImpact {
	i := -1
	for candidate := range s.PlayerColonies {
		if s.PlayerColonyStarIndex(candidate) == m.StarIndex {
			i = candidate
			break
		}
	}
	if i < 0 {
		return monsterColonyImpact{}
	}
	var buildings map[string]bool
	if i < len(s.ColonyBuildings) {
		buildings = s.ColonyBuildings[i]
	}
	if s.monsterDefenseBattle(m, buildings, s.Player) {
		return monsterColonyImpact{MonsterDestroyed: true}
	}
	hits := s.monsterStrategicBombardHits(m, gamedata.PlanetaryShieldReduction(buildings))
	marines, tanks := 0, 0
	if i < len(s.PlayerColonyMarines) {
		marines = s.PlayerColonyMarines[i]
	}
	if i < len(s.PlayerColonyTanks) {
		tanks = s.PlayerColonyTanks[i]
	}
	result := s.resolveMonsterDamage(&s.PlayerColonies[i], buildings, marines, tanks, s.Player,
		s.raceHasTrait(gamedata.TRAIT_HIGH_G), hits)
	for _, removed := range removeDestroyedRawBuildings(buildings, result.DestroyedBuildingIDs) {
		s.removeBuildingEffect(i, removed)
	}
	if i < len(s.PlayerColonyMarines) {
		s.PlayerColonyMarines[i] = result.State.Marines
	}
	if i < len(s.PlayerColonyTanks) {
		s.PlayerColonyTanks[i] = result.State.Tanks
	}
	impact := impactFromMonsterDamage(hits, result)
	if result.ColonyDestroyed {
		s.removePlayerColony(i)
		if !s.galaxyHasActiveColonyAtStar(m.StarIndex) {
			s.Stars[m.StarIndex].Owner = 0
		}
	} else {
		s.recalcColonyMorale(i)
	}
	return impact
}

func (s *GameSession) attackAIColonyWithMonster(m *MonsterGuard, aiIndex int) monsterColonyImpact {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return monsterColonyImpact{}
	}
	a := &s.AIPlayers[aiIndex]
	i := -1
	for candidate, star := range a.ColonyStars {
		if star == m.StarIndex && candidate < len(a.Colonies) {
			i = candidate
			break
		}
	}
	if i < 0 {
		return monsterColonyImpact{}
	}
	ensureAIGroundForceSlots(a)
	var buildings map[string]bool
	if i < len(a.ColonyBuildings) {
		buildings = a.ColonyBuildings[i]
	}
	if s.monsterDefenseBattle(m, buildings, a.Player) {
		return monsterColonyImpact{MonsterDestroyed: true}
	}
	hits := s.monsterStrategicBombardHits(m, gamedata.PlanetaryShieldReduction(buildings))
	result := s.resolveMonsterDamage(&a.Colonies[i], buildings, a.ColonyMarines[i], a.ColonyTanks[i],
		a.Player, aiRaceHasTrait(*a, gamedata.TRAIT_HIGH_G), hits)
	for _, removed := range removeDestroyedRawBuildings(buildings, result.DestroyedBuildingIDs) {
		removeBuildingEffectFromColony(&a.Colonies[i], removed)
	}
	a.ColonyMarines[i], a.ColonyTanks[i] = result.State.Marines, result.State.Tanks
	impact := impactFromMonsterDamage(hits, result)
	if result.ColonyDestroyed {
		s.removeAIColonyAfterEvent(aiIndex, i)
		if !s.galaxyHasActiveColonyAtStar(m.StarIndex) {
			s.Stars[m.StarIndex].Owner = 0
		}
	}
	return impact
}

// resolveEventMonsterColonyAttack 依目的星目前 owner 寫回玩家、熱座或 AI 殖民地。
func (s *GameSession) resolveEventMonsterColonyAttack(m *MonsterGuard) (monsterColonyImpact, bool) {
	if m == nil || m.Kind == gamedata.MonsterEel {
		return monsterColonyImpact{}, false
	}
	target, ok := s.eventEmpireTargetAtStar(m.StarIndex)
	if !ok {
		return monsterColonyImpact{}, false
	}
	switch target.kind {
	case eventEmpireAI:
		return s.attackAIColonyWithMonster(m, target.index), true
	case eventEmpireSeat:
		if target.index == s.ActiveSeat {
			return s.attackPlayerColonyWithMonster(m), true
		}
		if target.index < 0 || target.index >= len(s.Seats) {
			return monsterColonyImpact{}, false
		}
		active := s.ActiveSeat
		if active >= 0 && active < len(s.Seats) {
			s.Seats[active] = s.saveSeat()
		}
		s.loadSeat(s.Seats[target.index])
		impact := s.attackPlayerColonyWithMonster(m)
		s.Seats[target.index] = s.saveSeat()
		if active >= 0 && active < len(s.Seats) {
			s.loadSeat(s.Seats[active])
		}
		return impact, true
	default:
		return s.attackPlayerColonyWithMonster(m), true
	}
}

func (s *GameSession) monsterColonyImpactMessage(m *MonsterGuard, impact monsterColonyImpact) string {
	name := gamedata.MonsterNameZH(m.Kind)
	star := s.starName(m.StarIndex)
	if impact.MonsterDestroyed {
		return fmt.Sprintf("%s抵達 %s 星系後遭殖民地防禦擊毀", name, star)
	}
	if impact.ColonyDestroyed {
		return fmt.Sprintf("%s攻陷並摧毀 %s 星系殖民地", name, star)
	}
	return fmt.Sprintf("%s轟炸 %s 星系：%d 點戰略傷害，損失人口 %d、建築 %d",
		name, star, impact.Hits, impact.PopulationLost, impact.BuildingsDestroyed)
}
