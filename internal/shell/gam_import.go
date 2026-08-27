package shell

import (
	"fmt"
	"os"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

// GAMImportReport 是一次原版 .GAM 匯入的可追溯摘要。
//
// `.GAM` 不是 remake 的 JSON 快照；它保存的是固定大小的原版全局記錄與
// 順序陣列。報告把「有轉進工作階段」與「刻意保留／跳過」分開，避免讀檔
// 成功被誤讀成所有未知欄位都已還原。
type GAMImportReport struct {
	SourceVersion    uint32
	SaveGameName     string
	Stardate         uint32
	Turn             int
	HumanPlayer      int
	StarCount        int
	PlanetCount      int
	ColonyCount      int
	PlayerCount      int
	ShipCount        int
	ImportedColonies int
	ImportedOutposts int
	ImportedAI       int
	ImportedShips    int
	ImportedLeaders  int
	SkippedShips     int
	SkippedBuildings int
	Notes            []string
}

// LoadGAMSession 讀取原版 .GAM，轉成可直接交給 remake shell 的工作階段。
// 原始檔只經 internal/save 唯讀解析，不會被寫回或改名。
func LoadGAMSession(path string) (*GameSession, GAMImportReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, GAMImportReport{}, fmt.Errorf("讀取原版 GAM %s: %w", path, err)
	}
	session, report, err := ImportGAM(data)
	if err != nil {
		return nil, report, fmt.Errorf("匯入原版 GAM %s: %w", path, err)
	}
	return session, report, nil
}

// ImportGAM 解析原版存檔並建立最小但可玩的 remake 工作階段。
//
// 已接入：星系／行星／殖民地／前哨站、玩家與 AI 經濟、外交旗標、67 筆
// 全局領袖、玩家艦隊、研究中的主題、建築與建造佇列。研究完成陣列的
// byte 語意、未知特殊槽與原版星圖中途航行暫態仍以報告註記，不用猜測值
// 覆蓋工作階段。
func ImportGAM(data []byte) (*GameSession, GAMImportReport, error) {
	raw, err := save.Load(data)
	if err != nil {
		return nil, GAMImportReport{}, err
	}

	report := GAMImportReport{
		SourceVersion: raw.Config.Version,
		SaveGameName:  raw.Config.SaveGameName,
		Stardate:      raw.Config.Stardate,
		Turn:          gamTurnFromStardate(raw.Config.Stardate),
		StarCount:     boundedCount(raw.StarCount, len(raw.Stars)),
		PlanetCount:   boundedCount(raw.PlanetCount, len(raw.Planets)),
		ColonyCount:   boundedCount(raw.ColonyCount, len(raw.Colonies)),
		PlayerCount:   boundedCount(raw.PlayerCount, len(raw.Players)),
		ShipCount:     boundedCount(raw.ShipCount, len(raw.Ships)),
	}
	report.HumanPlayer = gamHumanPlayer(raw.Players, report.PlayerCount, &report)
	if report.PlayerCount == 0 {
		return nil, report, fmt.Errorf("原版 GAM 沒有可匯入的玩家記錄")
	}

	session := &GameSession{
		Turn:                report.Turn,
		Stars:               importGAMStars(raw, report.HumanPlayer, &report),
		Planets:             importGAMPlanets(raw, report.StarCount, &report),
		SelectedStar:        -1,
		Difficulty:          0,
		EventSeed:           int64(raw.Config.Stardate)*1009 + int64(report.StarCount),
		DisableEvents:       !raw.Config.RandomEvents,
		ShowRelocationLines: raw.Config.ShowRelocationLines,
		GameSettings: GameSettings{
			Version:          gameSettingsVersion,
			EndOfTurnSummary: raw.Config.EndOfTurnSummary, EndOfTurnWait: raw.Config.EndOfTurnWait,
			EnemyMoves: raw.Config.EnemyMoves, ExpandingHelp: raw.Config.ExpandingHelp,
			AutoSelectShips: raw.Config.AutoSelectShips, Animations: raw.Config.Animations,
			AutoSelectColony: raw.Config.AutoSelectColony, ShowRelocationLines: raw.Config.ShowRelocationLines,
			ShowGNNReport:              raw.Config.ShowGNNReport,
			AutoDeleteTradeGoodHousing: raw.Config.AutoDeleteTradeGoodHousing,
			AutoSaveGame:               true,
			ShowOnlySeriousTurnSummary: raw.Config.ShowOnlySeriousTurnSummary,
			ShipInitiative:             raw.Config.ShipInitiative,
		},
		RuleProfile:                 gamedata.Profile15(),
		PlayerName:                  raw.Players[report.HumanPlayer].Name,
		FlagColor:                   int(raw.Players[report.HumanPlayer].Color),
		ColonyLeaderNames:           []string{},
		PlayerColonies:              []engine.ColonyState{},
		PlayerColonyStars:           []int{},
		PlayerColonyPlanets:         []int{},
		PlayerColonyMarines:         []int{},
		PlayerColonyTanks:           []int{},
		MarineBarracksAge:           []int{},
		ArmorBarracksAge:            []int{},
		ColonyBuildings:             []map[string]bool{},
		Builds:                      []ColonyBuild{},
		BuildQueue:                  [][]ColonyBuild{},
		Outposts:                    []Outpost{},
		Leaders:                     []Leader{},
		AIPlayers:                   []AIOpponent{},
		PlayerSpies:                 []int{},
		PlayerSpyMissions:           []SpyMission{},
		CustomRaceTraits:            0,
		AssimilationProgressVersion: 1,
		GalaxyAgeSet:                false,
		TechLevelSet:                false,
		AntaresRaids:                0,
	}

	importGAMPlayer(session, raw.Players[report.HumanPlayer], &report)
	importGAMColonies(session, raw, report.HumanPlayer, &report)
	importGAMOpponents(session, raw, report.HumanPlayer, &report)
	importGAMLeaders(session, raw, report.HumanPlayer, &report)
	importGAMShips(session, raw, report.HumanPlayer, &report)

	// `localName` 會使用 cmd 層目前注入的中英文名稱翻譯器；原始英文另外
	// 留在 NameEN，確保英文 fallback 與存檔回查不依賴當前語系。
	for i := range session.Stars {
		session.Stars[i].Name = session.localName(session.Stars[i].NameEN)
	}
	for fi := range session.Fleets {
		for si := range session.Fleets[fi].Ships {
			if session.Fleets[fi].Ships[si].Name != "" {
				session.Fleets[fi].Ships[si].Name = session.localName(session.Fleets[fi].Ships[si].Name)
			}
		}
	}
	if len(session.Fleets) == 0 {
		session.Fleets = []Fleet{NewFleet(importedHomeStar(session))}
	}
	session.SelectedFleet = 0
	session.syncRaceEngineFields()
	for i := range session.AIPlayers {
		session.syncAIRaceEngineFields(&session.AIPlayers[i])
	}

	return session, report, nil
}

func boundedCount(n, length int) int {
	if n < 0 {
		return 0
	}
	if n > length {
		return length
	}
	return n
}

func gamTurnFromStardate(stardate uint32) int {
	turn := int(stardate/10) - StartStardate + 1
	if turn < 1 {
		return 1
	}
	return turn
}

func gamHumanPlayer(players []save.Player, count int, report *GAMImportReport) int {
	for i := 0; i < count; i++ {
		if players[i].Personality == 100 {
			return i
		}
	}
	for i := 0; i < count; i++ {
		if players[i].Eliminated == 0 {
			report.Notes = append(report.Notes,
				"原版存檔沒有 Personality=100 的真人席位；依第一個未淘汰帝國作為匯入玩家。這符合觀察者／全 AI autosave 的實際形狀。")
			return i
		}
	}
	report.Notes = append(report.Notes, "原版存檔沒有可辨識的真人或未淘汰帝國；保守使用玩家索引 0。")
	return 0
}

func importGAMStars(raw *save.GameState, human int, report *GAMImportReport) []Star {
	n := report.StarCount
	out := make([]Star, n)
	width, height := float64(raw.Galaxy.Width), float64(raw.Galaxy.Height)
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	for i := 0; i < n; i++ {
		r := raw.Stars[i]
		orbits := emptyOrbits()
		for orbit := range r.PlanetIndex {
			idx := int(r.PlanetIndex[orbit])
			if idx >= 0 && idx < report.PlanetCount {
				orbits[orbit] = idx
			}
		}
		owner := 0
		if int(r.Owner) == human {
			owner = 1
		} else if int(r.Owner) >= 0 && int(r.Owner) < report.PlayerCount && raw.Players[int(r.Owner)].Eliminated == 0 {
			owner = 2
		}
		nameEN := r.Name
		if nameEN == "" {
			nameEN = fmt.Sprintf("ORIGINAL STAR %d", i+1)
		}
		wormhole := int(r.Wormhole)
		if wormhole < 0 || wormhole >= n {
			wormhole = -1
		}
		out[i] = Star{
			X:        float64(r.X) / width,
			Y:        float64(r.Y) / height,
			Spectral: int(r.SpectralClass),
			Size:     int(r.Size),
			NameEN:   nameEN,
			Name:     nameEN,
			Owner:    owner,
			Explored: r.Visited != 0 || owner == 1,
			Orbits:   orbits,
			Wormhole: wormhole,
			InNebula: r.InNebula != 0,
		}
	}
	return out
}

func importGAMPlanets(raw *save.GameState, starCount int, report *GAMImportReport) []Planet {
	n := report.PlanetCount
	out := make([]Planet, n)
	roman := [...]string{"I", "II", "III", "IV", "V"}
	for i := 0; i < n; i++ {
		r := raw.Planets[i]
		star := int(r.Star)
		if star < 0 || star >= starCount {
			star = -1
		}
		name := fmt.Sprintf("ORIGINAL PLANET %d", i+1)
		if star >= 0 && star < len(raw.Stars) && raw.Stars[star].Name != "" {
			orbit := int(r.Orbit)
			suffix := ""
			if orbit >= 0 && orbit < len(roman) {
				suffix = " " + roman[orbit]
			}
			name = raw.Stars[star].Name + suffix
		}
		climate := gamPlanetClimate(r.Climate)
		gravity := gamPlanetGravity(r.Gravity)
		mineral := gamPlanetMineral(r.Minerals)
		size := gamPlanetSize(r.Size)
		planetType := gamPlanetType(r.Type)
		out[i] = Planet{
			Name:        name,
			Climate:     climateDisplayName(climate),
			Gravity:     gravityDisplayName(gravity),
			Mineral:     mineralDisplayName(mineral),
			Size:        sizeDisplayName(size),
			Gen:         planetGenVersion,
			ClimateID:   climate,
			GravityID:   gravity,
			MineralID:   mineral,
			SizeID:      size,
			Orbit:       int(r.Orbit),
			NoPlanet:    r.Type == 0,
			SpecialID:   gamPlanetSpecial(r.Special),
			SpecialSeen: star >= 0 && raw.Stars[star].Visited != 0,
			TypeID:      planetType,
		}
	}
	return out
}

func gamPlanetType(raw uint8) gamedata.PlanetType {
	switch gamedata.PlanetType(raw) {
	case gamedata.ASTEROIDS, gamedata.GAS_GIANT, gamedata.HABITABLE:
		return gamedata.PlanetType(raw)
	default:
		return gamedata.HABITABLE
	}
}

func gamPlanetClimate(raw uint8) gamedata.PlanetClimate {
	if raw <= uint8(gamedata.GAIA) {
		return gamedata.PlanetClimate(raw)
	}
	return gamedata.TOXIC
}

func gamPlanetGravity(raw uint8) gamedata.PlanetGravity {
	if raw <= uint8(gamedata.HEAVY_G) {
		return gamedata.PlanetGravity(raw)
	}
	return gamedata.NORMAL_G
}

func gamPlanetMineral(raw uint8) gamedata.PlanetMinerals {
	if raw <= uint8(gamedata.ULTRA_RICH) {
		return gamedata.PlanetMinerals(raw)
	}
	return gamedata.POOR
}

func gamPlanetSize(raw uint8) gamedata.PlanetSize {
	if raw <= uint8(gamedata.HUGE_PLANET) {
		return gamedata.PlanetSize(raw)
	}
	return gamedata.MEDIUM_PLANET
}

func gamPlanetSpecial(raw uint8) gamedata.PlanetSpecial {
	// PlanetSpecial 的原版表目前是 0..11；對超出表的髒值回到「無特殊」，
	// 不把未知 flag 當成可觸發的發現效果。
	if raw <= 11 {
		return gamedata.PlanetSpecial(raw)
	}
	return gamedata.PlanetSpecial(0)
}

func importGAMPlayer(session *GameSession, raw save.Player, report *GAMImportReport) {
	session.Player = importedPlayerState(&raw)
	session.PlayerName = raw.Name
	session.FlagColor = int(raw.Color)
	session.RaceIndex = raceIndexForEnglishName(raw.Race)
	if session.RaceIndex >= 0 && session.RaceIndex < len(Races) {
		r := Races[session.RaceIndex]
		session.raceGrowthPct = r.GrowthPct
		session.RaceCombatPct = r.CombatPct
		session.RaceShipDefPct = r.ShipDefPct
		session.RaceGroundBonus = r.GroundCombatBonus
		session.RaceSpyBonus = r.SpyBonus
		session.Player.FantasticTrader = r.IncomePerPop > 0 || raw.Traits[gamedata.TRAIT_FANTASTIC_TRADERS] != 0
	} else {
		session.RaceIndex = -1
		for trait := gamedata.TRAIT_LOW_G; trait <= gamedata.TRAIT_POOR_HOMEWORLD; trait++ {
			if int(trait) < len(raw.Traits) && raw.Traits[trait] != 0 {
				session.CustomRaceTraits |= uint32(1) << uint(trait)
			}
		}
		session.raceGrowthPct = int(raw.Traits[gamedata.TRAIT_POPULATION])
		session.RaceCombatPct = int(raw.Traits[gamedata.TRAIT_SHIP_ATTACK])
		session.RaceShipDefPct = int(raw.Traits[gamedata.TRAIT_SHIP_DEFENSE])
		session.RaceGroundBonus = int(raw.Traits[gamedata.TRAIT_GROUND_COMBAT])
		session.RaceSpyBonus = int(raw.Traits[gamedata.TRAIT_SPYING])
		session.Player.FantasticTrader = raw.Traits[gamedata.TRAIT_FANTASTIC_TRADERS] != 0
		report.Notes = append(report.Notes, "玩家種族名稱不在 13 族表；保留 GAM Traits 位元作客製種族遮罩，數值加成採存檔 Trait 原值。")
	}
	gov := int(raw.Traits[gamedata.TRAIT_GOVERNMENT])
	if gov < int(gamedata.MoraleGovFeudalism) || gov > int(gamedata.MoraleGovGalacticUnification) {
		gov = int(gamedata.MoraleGovDictatorship)
	}
	session.Government = gamedata.MoraleGovernmentType(gov)

	session.PlayerSpies = make([]int, 0)
	session.PlayerSpyMissions = make([]SpyMission, 0)
	// ResearchTopics 的完成狀態 byte 編碼尚未由原版初始化／消費端完全證實；
	// 目前只轉入目前研究主題與進度，不誤把任意非零 byte 升格成已完成科技。
	report.Notes = append(report.Notes, "GAM 研究完成陣列保留 raw 但未猜測 byte 語意；目前研究主題／進度已匯入。")
}

func importedPlayerState(raw *save.Player) engine.PlayerState {
	state := engine.PlayerStateFromSave(raw)
	state.CommandPointsSupply = int(raw.CommandPoints)
	if raw.UsedCommandPoints > 0 {
		state.UsedCommandPoints = int(raw.UsedCommandPoints)
	}
	state.FantasticTrader = raw.Traits[gamedata.TRAIT_FANTASTIC_TRADERS] != 0
	return state
}

func importGAMColonies(session *GameSession, raw *save.GameState, human int, report *GAMImportReport) {
	for i := 0; i < report.ColonyCount; i++ {
		colony := &raw.Colonies[i]
		planetIndex := int(colony.Planet)
		if planetIndex < 0 || planetIndex >= report.PlanetCount {
			continue
		}
		planet := &raw.Planets[planetIndex]
		owner := int(colony.Owner)
		if owner != human {
			continue
		}
		if colony.IsOutpost != 0 {
			star := int(planet.Star)
			if star >= 0 && star < report.StarCount {
				session.Outposts = append(session.Outposts, Outpost{StarIndex: star, PlanetIndex: planetIndex, Turn: session.Turn})
				report.ImportedOutposts++
			}
			continue
		}
		state := engine.ColonyStateFromSave(colony, planet)
		applyGAMPopulationProfiles(&state, raw, owner)
		session.PlayerColonies = append(session.PlayerColonies, state)
		session.PlayerColonyStars = append(session.PlayerColonyStars, int(planet.Star))
		session.PlayerColonyPlanets = append(session.PlayerColonyPlanets, planetIndex)
		session.PlayerColonyMarines = append(session.PlayerColonyMarines, int(colony.Soldiers))
		session.PlayerColonyTanks = append(session.PlayerColonyTanks, int(colony.Tanks))
		session.MarineBarracksAge = append(session.MarineBarracksAge, 0)
		session.ArmorBarracksAge = append(session.ArmorBarracksAge, 0)
		session.ColonyBuildings = append(session.ColonyBuildings, importedBuildingSet(colony, report))
		session.ColonyLeaderNames = append(session.ColonyLeaderNames, "")
		current, queue := importedBuildQueue(colony, report)
		session.Builds = append(session.Builds, current)
		session.BuildQueue = append(session.BuildQueue, queue)
		report.ImportedColonies++
	}
}

func importGAMOpponents(session *GameSession, raw *save.GameState, human int, report *GAMImportReport) {
	aiIndexByRaw := make(map[int]int)
	for i := 0; i < report.PlayerCount; i++ {
		if i == human || raw.Players[i].Eliminated != 0 {
			continue
		}
		personality := ai.Personality(raw.Players[i].Personality)
		opp := AIOpponent{
			Name:               raw.Players[i].Name,
			Color:              int(raw.Players[i].Color),
			ColorKnown:         true,
			RaceIndex:          raceIndexForEnglishName(raw.Players[i].Race),
			PopulationRaceSlot: i, PopulationRaceSlotKnown: true,
			Player:      importedPlayerState(&raw.Players[i]),
			Personality: personality,
			Decider:     ai.NewRemakeDecider(ai.ProfileForPersonality(personality)),
			Relation:    int(raw.Players[human].PlayerRelations[i]),
			Treaty: TreatyState{
				FormalPolicy:   gamedata.ForeignPolicy(raw.Players[human].ForeignPolicies[i]),
				TradeActive:    raw.Players[human].TradeTreaties[i] != 0,
				ResearchActive: raw.Players[human].ResearchTreaties[i] != 0,
			},
			Spies:                 int(raw.Players[i].Spies[human]),
			OriginalWarFlag60ERaw: int(raw.Players[i].Raw60E),
			OwnedStars:            0,
			ColonyStars:           []int{}, ColonyPlanets: []int{}, ColonyBuildings: []map[string]bool{},
			Leaders: []Leader{},
		}
		aiIndexByRaw[i] = len(session.AIPlayers)
		session.AIPlayers = append(session.AIPlayers, opp)
		report.ImportedAI++
	}

	for starIndex := 0; starIndex < report.StarCount; starIndex++ {
		owner := int(raw.Stars[starIndex].Owner)
		if owner == human {
			continue
		}
		if aiIndex, ok := aiIndexByRaw[owner]; ok {
			session.AIPlayers[aiIndex].OwnedStars++
		}
	}
	for colonyIndex := 0; colonyIndex < report.ColonyCount; colonyIndex++ {
		colony := &raw.Colonies[colonyIndex]
		owner := int(colony.Owner)
		aiIndex, ok := aiIndexByRaw[owner]
		if !ok || colony.IsOutpost != 0 {
			continue
		}
		planetIndex := int(colony.Planet)
		if planetIndex < 0 || planetIndex >= report.PlanetCount {
			continue
		}
		planet := &raw.Planets[planetIndex]
		state := engine.ColonyStateFromSave(colony, planet)
		applyGAMPopulationProfiles(&state, raw, owner)
		session.AIPlayers[aiIndex].Colonies = append(session.AIPlayers[aiIndex].Colonies, state)
		session.AIPlayers[aiIndex].ColonyStars = append(session.AIPlayers[aiIndex].ColonyStars, int(planet.Star))
		session.AIPlayers[aiIndex].ColonyPlanets = append(session.AIPlayers[aiIndex].ColonyPlanets, planetIndex)
		session.AIPlayers[aiIndex].ColonyBuildings = append(session.AIPlayers[aiIndex].ColonyBuildings,
			importedBuildingSet(colony, report))
		// GAM 的 Soldiers/Tanks 是原始殖民地駐軍欄位；匯入後直接保留，
		// 不以 age 公式覆蓋既有存檔狀態。後續回合才由兵營公式補充。
		session.AIPlayers[aiIndex].ColonyMarines = append(session.AIPlayers[aiIndex].ColonyMarines, int(colony.Soldiers))
		session.AIPlayers[aiIndex].ColonyTanks = append(session.AIPlayers[aiIndex].ColonyTanks, int(colony.Tanks))
		session.AIPlayers[aiIndex].MarineBarracksAge = append(session.AIPlayers[aiIndex].MarineBarracksAge, 0)
		session.AIPlayers[aiIndex].ArmorBarracksAge = append(session.AIPlayers[aiIndex].ArmorBarracksAge, 0)
	}
	for shipIndex := 0; shipIndex < report.ShipCount; shipIndex++ {
		owner := int(raw.Ships[shipIndex].Owner)
		aiIndex, ok := aiIndexByRaw[owner]
		if !ok {
			continue
		}
		class := importedShipClass(raw.Ships[shipIndex].Design)
		session.AIPlayers[aiIndex].FleetStrength += shipStrength(class)
	}
	for i := range session.AIPlayers {
		if session.AIPlayers[i].OwnedStars == 0 && len(session.AIPlayers[i].Colonies) > 0 {
			session.AIPlayers[i].OwnedStars = len(session.AIPlayers[i].Colonies)
		}
	}
	if len(session.AIPlayers) > 0 {
		session.PlayerSpies = make([]int, len(session.AIPlayers))
		session.PlayerSpyMissions = make([]SpyMission, len(session.AIPlayers))
		for rawIndex, aiIndex := range aiIndexByRaw {
			session.PlayerSpies[aiIndex] = int(raw.Players[human].Spies[rawIndex])
		}
	}
}

func applyGAMPopulationProfiles(c *engine.ColonyState, raw *save.GameState, owner int) {
	if c == nil || raw == nil || owner < 0 || owner >= len(raw.Players) {
		return
	}
	ownerTraits := raw.Players[owner].Traits
	c.OwnerFoodBonus = int(ownerTraits[gamedata.TRAIT_FARMING])
	c.OwnerIndustryBonus = int(ownerTraits[gamedata.TRAIT_INDUSTRY])
	c.OwnerResearchBonus = int(ownerTraits[gamedata.TRAIT_SCIENCE])
	c.OwnerRaceProfileKnown = true
	c.OwnerRaceSlot, c.OwnerRaceSlotKnown = owner, true
	for i := range c.PopulationGroups {
		g := &c.PopulationGroups[i]
		if g.RaceSlot < 0 || g.RaceSlot >= len(raw.Players) {
			if food, industry, research, immune, ok := gamedata.SpecialColonistProduction(g.RaceSlot); ok {
				g.FoodBonus, g.IndustryBonus, g.ResearchBonus = food, industry, research
				g.Gravity, g.GravityImmune, g.ProfileKnown = gamedata.NORMAL_G, immune, true
			}
			continue
		}
		traits := raw.Players[g.RaceSlot].Traits
		g.FoodBonus = int(traits[gamedata.TRAIT_FARMING])
		g.IndustryBonus = int(traits[gamedata.TRAIT_INDUSTRY])
		g.ResearchBonus = int(traits[gamedata.TRAIT_SCIENCE])
		g.Gravity = raceGravityForTraits(traits[gamedata.TRAIT_LOW_G] != 0, traits[gamedata.TRAIT_HIGH_G] != 0)
		g.Aquatic = traits[gamedata.TRAIT_AQUATIC] != 0
		g.Cybernetic = traits[gamedata.TRAIT_CYBERNETIC] != 0
		g.Lithovore = traits[gamedata.TRAIT_LITHOVORE] != 0
		g.Tolerant = traits[gamedata.TRAIT_TOLERANT] != 0
		g.Subterranean = traits[gamedata.TRAIT_SUBTERRANEAN] != 0
		g.GrowthBonusPercent = int(traits[gamedata.TRAIT_POPULATION])
		g.ProfileKnown = true
	}
}

func importGAMLeaders(session *GameSession, raw *save.GameState, human int, report *GAMImportReport) {
	aiIndexByRaw := make(map[int]int)
	for rawIndex := 0; rawIndex < report.PlayerCount; rawIndex++ {
		if rawIndex == human || raw.Players[rawIndex].Eliminated != 0 {
			continue
		}
		for aiIndex := range session.AIPlayers {
			if session.AIPlayers[aiIndex].Name == raw.Players[rawIndex].Name {
				aiIndexByRaw[rawIndex] = aiIndex
				break
			}
		}
	}
	for i := range raw.Leaders {
		if raw.Leaders[i].Name == "" {
			continue
		}
		leader := importedLeader(raw.Leaders[i], i)
		owner := int(raw.Leaders[i].PlayerIndex)
		report.ImportedLeaders++
		if owner == human || owner < 0 || owner >= report.PlayerCount {
			session.Leaders = append(session.Leaders, leader)
			continue
		}
		if aiIndex, ok := aiIndexByRaw[owner]; ok {
			session.AIPlayers[aiIndex].Leaders = append(session.AIPlayers[aiIndex].Leaders, leader)
		}
	}
}

func importedLeader(raw save.Leader, id int) Leader {
	leaderType := int(raw.Type)
	if leaderType != gamedata.LeaderTypeCaptain && leaderType != gamedata.LeaderTypeAdmin {
		leaderType = gamedata.LeaderTypeAdmin
	}
	skills := make([]LeaderSkill, 0)
	for _, skillID := range gamedata.LeaderSkillIDsFor(leaderType) {
		tier := gamedata.LeaderSkillTier(skillID, leaderType, raw.CommonSkills, raw.SpecialSkills)
		if tier <= 0 {
			continue
		}
		skills = append(skills, LeaderSkill{ID: skillID, Tier: tier})
	}
	skillName := "原版技能未知"
	tier := 0
	if len(skills) > 0 {
		if names, ok := gamedata.LeaderSkillName(skills[0].ID); ok {
			skillName = names.ZH
		}
		tier = skills[0].Tier
	}
	level := int(raw.Level)
	if level < 1 {
		level = 1
	}
	if level > 5 {
		level = 5
	}
	return Leader{
		ID: id, Name: raw.Name, Skill: skillName, Level: level,
		Ship: leaderType == gamedata.LeaderTypeCaptain, Tier: tier, Skills: skills,
		RawExperience: int(raw.Experience), RawExperienceKnown: true,
		RawETA: int(raw.Eta), RawStatus: int(raw.Status), RawLocation: int(raw.Location),
		RawPlayerIndex: int(raw.PlayerIndex),
	}
}

func importGAMShips(session *GameSession, raw *save.GameState, human int, report *GAMImportReport) {
	for i := 0; i < report.ShipCount; i++ {
		ship := &raw.Ships[i]
		if int(ship.Owner) != human || ship.Design.Name == "" {
			continue
		}
		converted := importedShip(*ship, i, report)
		if ship.Officer >= 0 && int(ship.Officer) < len(raw.Leaders) {
			converted.OfficerName = raw.Leaders[ship.Officer].Name
		}
		star := int(ship.Star)
		if star < 0 || star >= len(session.Stars) {
			star = importedHomeStar(session)
		}
		fleet := -1
		for fi := range session.Fleets {
			if session.Fleets[fi].AtStar == star {
				fleet = fi
				break
			}
		}
		if fleet < 0 {
			session.Fleets = append(session.Fleets, NewFleet(star))
			fleet = len(session.Fleets) - 1
		}
		session.Fleets[fleet].Ships = append(session.Fleets[fleet].Ships, converted)
		report.ImportedShips++
	}
}

func importedHomeStar(session *GameSession) int {
	for _, star := range session.PlayerColonyStars {
		if star >= 0 && star < len(session.Stars) {
			return star
		}
	}
	return 0
}

func importedShip(raw save.Ship, id int, report *GAMImportReport) Ship {
	design := raw.Design
	mounts := importedWeaponMounts(design, report)
	weaponName, weaponAttack, arc, weaponAmmo := "無武裝", 0, gamedata.ARC_MONSTER_360, 0
	if len(mounts) > 0 {
		weaponName, weaponAttack, arc, weaponAmmo = mounts[0].Name, mounts[0].Attack, mounts[0].Arc, mounts[0].Ammo
	}
	armor := "無裝甲"
	if int(design.Armor) >= 0 && int(design.Armor) < len(ArmorOptions) {
		armor = ArmorOptions[design.Armor].Name
	}
	shield := "無護盾"
	if int(design.Shield) >= 0 && int(design.Shield) < len(ShieldOptions) {
		shield = ShieldOptions[design.Shield].Name
	}
	specialIDs := importedSpecialIDs(design)
	specials := importedSpecialMounts(design, specialIDs, report)
	special := "無"
	if len(specials) > 0 {
		special = specials[0].Name
	}
	name := design.Name
	if name == "" {
		name = fmt.Sprintf("ORIGINAL SHIP %d", id+1)
	}
	return Ship{
		Name: name, Class: importedShipClass(design), Weapon: weaponName,
		RawType: gamedata.ShipType(design.Type), RawTypeKnown: true,
		RawMission: raw.Mission, RawMissionKnown: true, ProductionCost: int(design.Cost),
		ComputerRaw: design.Computer, ComputerRawKnown: true,
		DesignSizeRaw: design.Size, DesignSizeRawKnown: true,
		ArmorRaw: design.Armor, ArmorRawKnown: true,
		ShieldRaw: design.Shield, ShieldRawKnown: true,
		BaseCombatSpeedRaw: design.BaseCombatSpeed, BaseCombatSpeedKnown: true,
		ShieldDamageRaw: raw.ShieldDamage, DriveDamageRaw: raw.DriveDamage,
		ComputerDamageRaw: raw.ComputerDamage, CrewLevelRaw: raw.CrewLevel, CrewLevelRawKnown: true,
		ArmorDamageRaw: int(raw.ArmorDamage), StructureDamageRaw: int(raw.StructureDamage),
		OriginalDamageKnown: true,
		DamagedSpecialsRaw:  raw.DamagedSpecials,
		Armor:               armor, Shield: shield, Special: special, Arc: arc,
		WeaponAmmo: weaponAmmo, WeaponMounts: mounts, SpecialIDs: specialIDs, Specials: specials,
		CombatPicture: int(design.Picture), CombatPictureKnown: true,
		OfficerID: int(raw.Officer), WeaponAttack: weaponAttack,
		BonusHP: armorComponentValue(design.Armor) + shieldComponentValue(design.Shield),
		Damage:  int(raw.StructureDamage), CrewXP: int(raw.CrewExp),
	}
}

func armorComponentValue(raw uint8) int {
	if int(raw) >= 0 && int(raw) < len(ArmorOptions) {
		return ArmorOptions[raw].Value
	}
	return 0
}

func shieldComponentValue(raw uint8) int {
	if int(raw) >= 0 && int(raw) < len(ShieldOptions) {
		return ShieldOptions[raw].Value
	}
	return 0
}

func importedShipClass(design save.ShipDesign) string {
	switch gamedata.ShipType(design.Type) {
	case gamedata.COLONY_SHIP:
		return "殖民船"
	case gamedata.TRANSPORT_SHIP:
		return "運兵船"
	case gamedata.OUTPOST_SHIP:
		return "前哨船"
	}
	classes := [...]string{"巡防艦", "驅逐艦", "巡洋艦", "戰艦", "泰坦", "末日之星"}
	if int(design.Size) >= 0 && int(design.Size) < len(classes) {
		return classes[design.Size]
	}
	return "巡防艦"
}

func importedWeaponMounts(design save.ShipDesign, report *GAMImportReport) []ShipWeaponMount {
	mounts := make([]ShipWeaponMount, 0, len(design.Weapons))
	for _, raw := range design.Weapons {
		if raw.Type < 0 || (raw.MaxCount == 0 && raw.WorkingCount == 0) {
			continue
		}
		id := int(raw.Type)
		name := fmt.Sprintf("原版武器#%d", id)
		attack := 0
		if id >= 0 && id < len(gamedata.OrigWeaponTable) {
			name = importedWeaponName(id)
			attack = gamedata.OrigWeaponTable[id].DamageMax
		} else {
			report.Notes = append(report.Notes, fmt.Sprintf("艦艇武器槽含未知 raw ID %d；保留名稱但攻擊值設 0。", id))
		}
		arc := gamedata.WeaponArc(raw.Arc)
		mounts = append(mounts, ShipWeaponMount{
			RawType: id, Name: name, MaxCount: int(raw.MaxCount), WorkingCount: int(raw.WorkingCount),
			Arc: arc, RawMods: raw.Mods, Ammo: int(raw.Ammo), Attack: attack,
		})
	}
	return mounts
}

func importedWeaponName(id int) string {
	names := map[int]string{
		1: "質量投射器", 2: "高斯砲", 3: "雷射", 4: "粒子束", 5: "核融合光束",
		6: "離子脈衝砲", 7: "引力波束", 8: "中子爆破槍", 9: "相位砲", 10: "干擾者",
		11: "死光", 12: "電漿砲", 13: "空間壓縮器", 14: "核飛彈", 15: "麥克萊特飛彈",
		16: "脈衝飛彈", 17: "氙素飛彈", 18: "反物質魚雷", 19: "質子魚雷", 20: "電漿魚雷",
		21: "核彈", 22: "核融合炸彈", 23: "反物質炸彈", 24: "中子炸彈", 25: "死亡孢子",
		26: "生物終結者", 27: "重錘裝置", 28: "突擊艇", 29: "重戰機庫", 30: "轟炸機庫",
		31: "戰機庫", 32: "停滯力場", 33: "反飛彈火箭", 34: "陀螺去穩器", 35: "電漿網",
		36: "脈衝星", 37: "黑洞產生器", 38: "恆星轉換器", 39: "牽引光束",
	}
	if name, ok := names[id]; ok {
		return name
	}
	return fmt.Sprintf("原版武器#%d", id)
}

func importedSpecialIDs(design save.ShipDesign) []int {
	var ids []int
	for byteIndex, raw := range design.Specials {
		for bit := 0; bit < 8; bit++ {
			if raw&(1<<uint(bit)) == 0 {
				continue
			}
			id := byteIndex*8 + bit
			ids = append(ids, id)
		}
	}
	return ids
}

func importedSpecialMounts(design save.ShipDesign, ids []int, report *GAMImportReport) []ShipSpecialMount {
	mounts := make([]ShipSpecialMount, 0, len(ids))
	for _, id := range ids {
		name, known := specialNameForRawID(id)
		if !known {
			report.Notes = append(report.Notes, fmt.Sprintf("艦艇 %q 的特殊裝置 raw ID %d 未知；保留但不猜效果。", design.Name, id))
		}
		mounts = append(mounts, ShipSpecialMount{RawID: id, Name: name})
	}
	return mounts
}

func importedBuildingSet(colony *save.Colony, report *GAMImportReport) map[string]bool {
	set := map[string]bool{}
	for rawID, status := range colony.Buildings {
		if status == 0 {
			continue
		}
		name, _, ok := importedBuilding(rawID)
		if !ok {
			report.SkippedBuildings++
			continue
		}
		set[name] = true
	}
	return set
}

func importedBuilding(id int) (name string, cost int, ok bool) {
	for _, building := range gamedata.Buildings {
		if gamedata.OrigBuildingID[building.NameEN] == id {
			return building.NameZH, building.ProductionCost, true
		}
	}
	return "", 0, false
}

func importedBuildQueue(colony *save.Colony, report *GAMImportReport) (ColonyBuild, []ColonyBuild) {
	current := ColonyBuild{}
	if colony.BuildQueue[0] != 0 {
		current = importedBuild(colony.BuildQueue[0], int(colony.BuildProgress), report)
	}
	queue := make([]ColonyBuild, 0, len(colony.BuildQueue)-1)
	for _, rawID := range colony.BuildQueue[1:] {
		if rawID == 0 {
			continue
		}
		queue = append(queue, importedBuild(rawID, 0, report))
	}
	return current, queue
}

func importedBuild(rawID int16, progress int, report *GAMImportReport) ColonyBuild {
	name, cost, ok := importedBuilding(int(rawID))
	if !ok {
		name = fmt.Sprintf("原版建造#%d", rawID)
		report.Notes = append(report.Notes, fmt.Sprintf("建造佇列保留未知 raw ID %d；不猜測其生產類型。", rawID))
	}
	return ColonyBuild{Name: name, Progress: progress, Cost: cost}
}
