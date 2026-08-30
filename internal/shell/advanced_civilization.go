package shell

import (
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// applyAdvancedCivilizationColonies 把 Advanced_Civilization_Colonies_ 的候選／輪替結果
// 接入正常新遊戲。呼叫時玩家種族已 finalize，故玩家與 AI 的種族產出可正確進殖民地。
func (s *GameSession) applyAdvancedCivilizationColonies() {
	if s.techLevel() != 2 || len(s.Stars) == 0 {
		return
	}
	players := len(s.AIPlayers) + 1
	quota := gamedata.AdvancedCivilizationExtraPlanetQuota(len(s.Stars), players)
	if quota <= 0 {
		return
	}
	homes := make([]int, players)
	homes[0] = 0
	for i := range s.AIPlayers {
		homes[i+1] = -1
		if len(s.AIPlayers[i].ColonyStars) > 0 {
			homes[i+1] = s.AIPlayers[i].ColonyStars[0]
		}
	}
	candidates := make([][]gamedata.AdvancedCivilizationCandidate, players)
	maxDistance := 0
	for player, home := range homes {
		if home < 0 || home >= len(s.Stars) {
			continue
		}
		for star := range s.Stars {
			distance := s.ParsecsBetweenStars(home, star)
			if distance > maxDistance {
				maxDistance = distance
			}
			if distance > 9 || s.Stars[star].Spectral == blackHoleSpectral {
				continue
			}
			for _, planet := range s.PlanetsAt(star) {
				if s.PlanetColonized(planet) || planet < 0 || planet >= len(s.Planets) {
					continue
				}
				p := s.Planets[planet]
				if p.NoPlanet || p.TypeID != gamedata.HABITABLE || !climateColonizable(p.ClimateID) {
					continue
				}
				baseWorth := s.advancedCivilizationPlanetWorth(player, planet)
				proximity := gamedata.AdvancedCivilizationProximityPercent(distance)
				worth := baseWorth * proximity / 100
				if star == home {
					worth += baseWorth * 67 / 100
				}
				candidates[player] = append(candidates[player], gamedata.AdvancedCivilizationCandidate{
					Planet: planet, Distance: distance, Worth: worth,
				})
			}
		}
	}
	order := rand.New(rand.NewSource(s.EventSeed + 0x62C70)).Perm(players)
	selected := gamedata.AdvancedCivilizationChoose(candidates, order, quota, s.Difficulty, maxDistance)
	balanceRand := rand.New(rand.NewSource(s.EventSeed + 0x638A9))
	s.balanceAdvancedCivilizationPlanets(selected, balanceRand)
	s.redistributeAdvancedCivilizationSpecials(selected, homes, balanceRand)
	for player, planets := range selected {
		for _, planet := range planets {
			if player == 0 {
				s.addAdvancedPlayerColony(planet)
			} else {
				s.addAdvancedAIColony(player-1, planet)
			}
		}
	}
}

func (s *GameSession) advancedPlayerHasArtifactsHomeworld(player int) bool {
	if player == 0 {
		traits, ok := s.playerStartingRuntimeTraits()
		return ok && traits[gamedata.TRAIT_ARTIFACTS_HOMEWORLD] != 0
	}
	aiIndex := player - 1
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return false
	}
	race := aiRaceIndex(s.AIPlayers[aiIndex])
	return race >= 0 && race < len(Races) &&
		gamedata.OrigRaceHasTrait(Races[race].OrigIdx, gamedata.TRAIT_ARTIFACTS_HOMEWORLD)
}

func (s *GameSession) starHasAdvancedSpecial(star int) bool {
	for _, planet := range s.PlanetsAt(star) {
		if planet >= 0 && planet < len(s.Planets) && s.Planets[planet].SpecialID != gamedata.NoSpecial {
			return true
		}
	}
	return false
}

// redistributeAdvancedCivilizationSpecials 對應 raw special 4／5／10 的平衡支線。
// remake 沒有獨立 star.special 欄，故以同星所有 planet special 聚合成相同排他 gate。
func (s *GameSession) redistributeAdvancedCivilizationSpecials(selected [][]int, homes []int, r *rand.Rand) {
	counts := make([]int, len(selected))
	bestPlayer, bestAverage := 0, -1
	for player, planets := range selected {
		for _, planet := range planets {
			if planet >= 0 && planet < len(s.Planets) {
				switch s.Planets[planet].SpecialID {
				case gamedata.GoldDeposits, gamedata.GemDeposits, gamedata.AncientArtifacts:
					counts[player]++
				}
			}
		}
		if avg := s.advancedSelectedAverage(player, planets); avg > bestAverage {
			bestPlayer, bestAverage = player, avg
		}
	}
	target := counts[bestPlayer]
	specials := [...]gamedata.PlanetSpecial{gamedata.GoldDeposits, gamedata.GemDeposits, gamedata.AncientArtifacts}
	for player, planets := range selected {
		if player == bestPlayer || len(planets) == 0 {
			continue
		}
		for attempts := 0; attempts < 100 && counts[player] < target; attempts++ {
			planet := planets[r.Intn(len(planets))]
			if planet < 0 || planet >= len(s.Planets) || s.Planets[planet].SpecialID != gamedata.NoSpecial {
				continue
			}
			star := s.PlanetStar(planet)
			if star < 0 || s.starHasAdvancedSpecial(star) {
				continue
			}
			if player < len(homes) && planet == s.PlanetAt(homes[player]) && s.advancedPlayerHasArtifactsHomeworld(player) {
				continue
			}
			s.Planets[planet].SpecialID = specials[r.Intn(len(specials))]
			counts[player]++
		}
	}
}

func (s *GameSession) advancedSelectedAverage(player int, planets []int) int {
	if len(planets) == 0 {
		return 0
	}
	total := 0
	for _, planet := range planets {
		total += s.advancedCivilizationPlanetWorth(player, planet)
	}
	return total * 10 / len(planets)
}

// balanceAdvancedCivilizationPlanets 對應 Twiddle_Selected_Adv_Civ_Planets_ 的 90% 門檻。
// 三個 raw 欄位依各自上限投影為 SizeID 4、ClimateID 9、MineralID 4。
func (s *GameSession) balanceAdvancedCivilizationPlanets(selected [][]int, r *rand.Rand) {
	best := 0
	for player := range selected {
		if avg := s.advancedSelectedAverage(player, selected[player]); avg > best {
			best = avg
		}
	}
	target := best * 90 / 100
	for player, planets := range selected {
		if s.advancedSelectedAverage(player, planets) >= target {
			continue
		}
		for _, planet := range planets {
			for attempt := 0; attempt < 6 && s.advancedSelectedAverage(player, planets) < target; attempt++ {
				if planet < 0 || planet >= len(s.Planets) {
					break
				}
				p := &s.Planets[planet]
				switch r.Intn(3) {
				case 0:
					if p.SizeID < gamedata.HUGE_PLANET {
						p.SizeID++
						p.Size = sizeDisplayName(p.SizeID)
					}
				case 1:
					if p.ClimateID < gamedata.GAIA {
						p.ClimateID++
						p.Climate = climateDisplayName(p.ClimateID)
					}
				case 2:
					if p.MineralID < gamedata.ULTRA_RICH {
						p.MineralID++
						p.Mineral = mineralDisplayName(p.MineralID)
					}
				}
			}
			if s.advancedSelectedAverage(player, planets) >= target {
				break
			}
		}
	}
}

func (s *GameSession) advancedCivilizationPlanetWorth(player, planet int) int {
	if player > 0 {
		return s.aiBasePlanetValueAt(player-1, planet)
	}
	p := s.Planets[planet]
	traits, _ := s.playerStartingRuntimeTraits()
	return gamedata.AIPlanetValue(gamedata.AIPlanetValueInput{
		Habitable: true, MaxPop: gamedata.PlanetBasePopMax(p.SizeID, p.ClimateID),
		Minerals: p.MineralID, Climate: p.ClimateID, Gravity: p.GravityID,
		FoodBase: gamedata.ClimateFoodPerFarmer(p.ClimateID), Special: int(p.SpecialID),
		RaceLowG: traits[gamedata.TRAIT_LOW_G] != 0, RaceHeavyG: traits[gamedata.TRAIT_HIGH_G] != 0,
	}, gamedata.AIObjectiveBalancedLow)
}

func (s *GameSession) addAdvancedPlayerColony(planet int) bool {
	star := s.PlanetStar(planet)
	if star < 0 || s.PlanetColonized(planet) {
		return false
	}
	food, industry, research := 0, 0, 0
	if s.RaceIndex >= 0 && s.RaceIndex < len(Races) {
		r := Races[s.RaceIndex]
		food, industry, research = r.FoodBonus, r.IndBonus, r.ResBonus
	}
	colony, ok, _ := s.newColonyFromPlanet(planet, s.Government, food, industry, research)
	if !ok {
		return false
	}
	s.appendPlayerColony(colony, star, planet)
	s.Stars[star].Owner = 1
	s.consumeSpecialOnColonize(planet)
	return true
}

func (s *GameSession) addAdvancedAIColony(aiIndex, planet int) bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || s.PlanetColonized(planet) {
		return false
	}
	star := s.PlanetStar(planet)
	if star < 0 {
		return false
	}
	a := &s.AIPlayers[aiIndex]
	race := aiRaceIndex(*a)
	food, industry, research := 0, 0, 0
	if race >= 0 && race < len(Races) {
		food, industry, research = Races[race].FoodBonus, Races[race].IndBonus, Races[race].ResBonus
	}
	colony, ok, _ := s.newColonyFromPlanet(planet, gamedata.MoraleGovDictatorship, food, industry, research)
	if !ok {
		return false
	}
	if race >= 0 && race < len(Races) {
		orig := Races[race].OrigIdx
		colony.TolerantRace = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_TOLERANT)
		colony.Lithovore = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_LITHOVORE)
		colony.Aquatic = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_AQUATIC)
		colony.Subterranean = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_SUBTERRANEAN)
		colony.PopMax = racePopulationMax(colony.PlanetSize, colony.Climate, colony.Aquatic,
			colony.TolerantRace, colony.Subterranean)
	}
	colony.PopulationGroups = nil
	colony.OwnerRaceSlotKnown = false
	a.Colonies = append(a.Colonies, colony)
	a.ColonyStars = append(a.ColonyStars, star)
	a.ColonyPlanets = append(a.ColonyPlanets, planet)
	a.ColonyBuildings = append(a.ColonyBuildings, map[string]bool{})
	a.ColonyMarines = append(a.ColonyMarines, 0)
	a.ColonyTanks = append(a.ColonyTanks, 0)
	a.MarineBarracksAge = append(a.MarineBarracksAge, 0)
	a.ArmorBarracksAge = append(a.ArmorBarracksAge, 0)
	if s.Stars[star].Owner == 0 {
		a.OwnedStars++
	}
	s.Stars[star].Owner = 2
	s.consumeSpecialOnColonize(planet)
	s.syncAIRaceEngineFields(a)
	return true
}
