package save

// 本檔逐欄位移植 openorion2 gamestate.cpp 的各實體 load()。欄位名沿用原碼(轉 Go 匯出命名)。
// 未知/保留區以 skip 略過(欄位不儲存,但位元組數精確保留以維持 offset)。

// 陣列大小常數(對照 gamestate.h)。
const (
	maxPopulation      = 42
	maxRaces           = maxPlayers + 2 // 玩家種族 + androids + natives = 10
	maxBuildQueue      = 7
	maxBuildings       = 49
	maxOrbits          = 5
	maxSettlers        = 25
	maxResearchTopics  = 83
	maxAppliedTechs    = 204
	maxResearchAreas   = 8
	maxTechnologies    = maxAppliedTechs + maxResearchAreas // 212
	maxPlayerBlueprint = 5
	traitsCount        = 31
	maxHistoryLength   = 350
	maxShipSpecials    = 40
	maxShipWeapons     = 8
	maxLeaderTechSkill = 3

	starsNameSize   = 15
	leaderNameSize  = 0x0f // 15
	leaderTitleSize = 0x14 // 20
	shipNameSize    = 16
	playerNameSize  = 0x14 // 20
	playerRaceSize  = 15
)

// bitfield helper:blackHole/blockade 等以位元陣列儲存。
func bitBytes(n int) int { return (n + 7) / 8 }

// ── Colonist(4 bytes,bit-packed u32)──────────────────────
type Colonist struct {
	Race    uint8
	Loyalty uint8
	Job     uint8
	Flags   uint32
}

func (c *Colonist) load(r *reader) {
	raw := r.u32()
	c.Race = uint8(raw & 0xf)
	c.Loyalty = uint8((raw >> 4) & 0x7)
	c.Job = uint8((raw >> 7) & 0x3)
	c.Flags = raw >> 9
}

// ── Colony(361 bytes)──────────────────────────────────────
type Colony struct {
	Owner      uint8
	Unknown1   int8
	Planet     int16
	Unknown2   int16
	IsOutpost  uint8
	Morale     int8
	Pollution  uint16
	Population uint8
	ColonyType uint8

	Colonists      [maxPopulation]Colonist
	RacePopulation [maxRaces]uint16
	PopGrowth      [maxRaces]int16

	Age                  uint8
	FoodPerFarmer        uint8
	IndustryPerWorker    uint8
	ResearchPerScientist uint8
	MaxFarms             int8
	MaxPopulation        uint8
	Climate              uint8
	GroundStrength       uint16
	SpaceStrength        uint16
	TotalFood            uint16
	NetIndustry          uint16
	TotalResearch        uint16
	TotalRevenue         uint16
	FoodConsumption      uint8
	IndustryConsumption  uint8
	ResearchConsumption  uint8
	Upkeep               uint8
	FoodImported         int16
	IndustryConsumed     uint16
	ResearchImported     int16
	BudgetDeficit        int16
	RecycledIndustry     uint8

	FoodConsumptionCitizens    uint8
	FoodConsumptionAliens      uint8
	FoodConsumptionPrisoners   uint8
	FoodConsumptionNatives     uint8
	IndustryConsumptionCitz    uint8
	IndustryConsumptionAndroid uint8
	IndustryConsumptionAliens  uint8
	IndustryConsumptionPrison  uint8

	FoodConsumptionRaces     [maxPlayers]uint8
	IndustryConsumptionRaces [maxPlayers]uint8
	ReplicatedFood           uint8

	BuildQueue         [maxBuildQueue]int16
	FinishedProduction int16
	BuildProgress      uint16
	TaxRevenue         uint16
	Autobuild          uint8
	Unknown3           uint16
	BoughtProgress     uint16
	AssimilationProg   uint8
	PrisonerPolicy     uint8
	Soldiers           uint16
	Tanks              uint16
	TankProgress       uint8
	SoldierProgress    uint8

	Buildings [maxBuildings]uint8
	Status    uint16
}

func (c *Colony) load(r *reader) {
	c.Owner = r.u8()
	c.Unknown1 = r.i8()
	c.Planet = r.i16()
	c.Unknown2 = r.i16()
	c.IsOutpost = r.u8()
	c.Morale = r.i8()
	c.Pollution = r.u16()
	c.Population = r.u8()
	c.ColonyType = r.u8()
	for i := range c.Colonists {
		c.Colonists[i].load(r)
	}
	for i := range c.RacePopulation {
		c.RacePopulation[i] = r.u16()
	}
	for i := range c.PopGrowth {
		c.PopGrowth[i] = r.i16()
	}
	c.Age = r.u8()
	c.FoodPerFarmer = r.u8()
	c.IndustryPerWorker = r.u8()
	c.ResearchPerScientist = r.u8()
	c.MaxFarms = r.i8()
	c.MaxPopulation = r.u8()
	c.Climate = r.u8()
	c.GroundStrength = r.u16()
	c.SpaceStrength = r.u16()
	c.TotalFood = r.u16()
	c.NetIndustry = r.u16()
	c.TotalResearch = r.u16()
	c.TotalRevenue = r.u16()
	c.FoodConsumption = r.u8()
	c.IndustryConsumption = r.u8()
	c.ResearchConsumption = r.u8()
	c.Upkeep = r.u8()
	c.FoodImported = r.i16()
	c.IndustryConsumed = r.u16()
	c.ResearchImported = r.i16()
	c.BudgetDeficit = r.i16()
	c.RecycledIndustry = r.u8()
	c.FoodConsumptionCitizens = r.u8()
	c.FoodConsumptionAliens = r.u8()
	c.FoodConsumptionPrisoners = r.u8()
	c.FoodConsumptionNatives = r.u8()
	c.IndustryConsumptionCitz = r.u8()
	c.IndustryConsumptionAndroid = r.u8()
	c.IndustryConsumptionAliens = r.u8()
	c.IndustryConsumptionPrison = r.u8()
	for i := range c.FoodConsumptionRaces {
		c.FoodConsumptionRaces[i] = r.u8()
	}
	for i := range c.IndustryConsumptionRaces {
		c.IndustryConsumptionRaces[i] = r.u8()
	}
	c.ReplicatedFood = r.u8()
	for i := range c.BuildQueue {
		c.BuildQueue[i] = r.i16()
	}
	c.FinishedProduction = r.i16()
	c.BuildProgress = r.u16()
	c.TaxRevenue = r.u16()
	c.Autobuild = r.u8()
	c.Unknown3 = r.u16()
	c.BoughtProgress = r.u16()
	c.AssimilationProg = r.u8()
	c.PrisonerPolicy = r.u8()
	c.Soldiers = r.u16()
	c.Tanks = r.u16()
	c.TankProgress = r.u8()
	c.SoldierProgress = r.u8()
	for i := range c.Buildings {
		c.Buildings[i] = r.u8()
	}
	c.Status = r.u16()
}

// ── Planet(16 bytes)──────────────────────────────────────
type Planet struct {
	Colony     int16
	Star       uint8
	Orbit      uint8
	Type       uint8
	Size       uint8
	Gravity    uint8
	Unknown1   uint8
	Climate    uint8
	Bg         uint8
	Minerals   uint8
	Foodbase   uint8
	Terraforms uint8
	Unknown2   uint8
	MaxPop     uint8
	Special    uint8
	Flags      uint8
}

func (p *Planet) load(r *reader) {
	p.Colony = r.i16()
	p.Star = r.u8()
	p.Orbit = r.u8()
	p.Type = r.u8()
	p.Size = r.u8()
	p.Gravity = r.u8()
	p.Unknown1 = r.u8()
	p.Climate = r.u8()
	p.Bg = r.u8()
	p.Minerals = r.u8()
	p.Foodbase = r.u8()
	p.Terraforms = r.u8()
	p.Unknown2 = r.u8()
	p.MaxPop = r.u8()
	p.Special = r.u8()
	p.Flags = r.u8()
}

// ── Star(113 bytes)───────────────────────────────────────
type Star struct {
	Name          string
	X, Y          uint16
	Size          uint8
	Owner         int8
	PictureType   uint8
	SpectralClass uint8

	LastPlanetSelected [maxPlayers]uint8
	BlackHoleBlocks    [bitBytes72]uint8

	Special     uint8
	Wormhole    int8
	Blockaded   uint8
	BlockadedBy [maxPlayers]uint8

	Visited                 uint8
	JustVisited             uint8
	IgnoreColonyShips       uint8
	IgnoreCombatShips       uint8
	ColonizePlayer          int8
	HasColony               uint8
	HasWarpFieldInterdictor uint8
	NextWFIInList           uint8
	HasTachyon              uint8
	HasSubspace             uint8
	HasStargate             uint8
	HasJumpgate             uint8
	HasArtemisNet           uint8
	HasDimensionalPortal    uint8
	IsStagepoint            uint8

	OfficerIndex     [maxPlayers]int8
	PlanetIndex      [maxOrbits]int16
	RelocateShipTo   [maxPlayers]uint16
	SurrenderTo      [maxPlayers]uint8
	InNebula         uint8
	ArtifactsGaveApp uint8
}

// bitBytes72 = (MAX_STARS+7)/8 = 9。
const bitBytes72 = (maxStars + 7) / 8

func (s *Star) load(r *reader) {
	s.Name = r.cstr(starsNameSize)
	s.X = r.u16()
	s.Y = r.u16()
	s.Size = r.u8()
	s.Owner = r.i8()
	s.PictureType = r.u8()
	s.SpectralClass = r.u8()
	for i := range s.LastPlanetSelected {
		s.LastPlanetSelected[i] = r.u8()
	}
	for i := range s.BlackHoleBlocks {
		s.BlackHoleBlocks[i] = r.u8()
	}
	s.Special = r.u8()
	s.Wormhole = r.i8()
	s.Blockaded = r.u8()
	for i := range s.BlockadedBy {
		s.BlockadedBy[i] = r.u8()
	}
	s.Visited = r.u8()
	s.JustVisited = r.u8()
	s.IgnoreColonyShips = r.u8()
	s.IgnoreCombatShips = r.u8()
	s.ColonizePlayer = r.i8()
	s.HasColony = r.u8()
	s.HasWarpFieldInterdictor = r.u8()
	s.NextWFIInList = r.u8()
	s.HasTachyon = r.u8()
	s.HasSubspace = r.u8()
	s.HasStargate = r.u8()
	s.HasJumpgate = r.u8()
	s.HasArtemisNet = r.u8()
	s.HasDimensionalPortal = r.u8()
	s.IsStagepoint = r.u8()
	for i := range s.OfficerIndex {
		s.OfficerIndex[i] = r.i8()
	}
	for i := range s.PlanetIndex {
		s.PlanetIndex[i] = r.i16()
	}
	for i := range s.RelocateShipTo {
		s.RelocateShipTo[i] = r.u16()
	}
	r.skip(3) // 3 個未知 u8
	for i := range s.SurrenderTo {
		s.SurrenderTo[i] = r.u8()
	}
	s.InNebula = r.u8()
	s.ArtifactsGaveApp = r.u8()
}

// ── Leader(59 bytes = LEADER_DATA_SIZE)────────────────────
type Leader struct {
	Name           string
	Title          string
	Type           uint8
	Experience     uint16
	CommonSkills   uint32
	SpecialSkills  uint32
	Techs          [maxLeaderTechSkill]uint8
	Picture        uint8
	SkillValue     uint16
	Level          uint8
	Location       int16
	Eta            uint8
	DisplayLevelUp uint8
	Status         int8
	PlayerIndex    int8
}

func (l *Leader) load(r *reader) {
	l.Name = r.cstr(leaderNameSize)
	l.Title = r.cstr(leaderTitleSize)
	l.Type = r.u8()
	l.Experience = r.u16()
	l.CommonSkills = r.u32()
	l.SpecialSkills = r.u32()
	for i := range l.Techs {
		l.Techs[i] = r.u8()
	}
	l.Picture = r.u8()
	l.SkillValue = r.u16()
	l.Level = r.u8()
	l.Location = r.i16()
	l.Eta = r.u8()
	l.DisplayLevelUp = r.u8()
	l.Status = r.i8()
	l.PlayerIndex = r.i8()
}

// ── ShipWeapon(8 bytes)───────────────────────────────────
type ShipWeapon struct {
	Type         int16
	MaxCount     uint8
	WorkingCount uint8
	Arc          uint8
	Mods         uint16
	Ammo         uint8
}

func (w *ShipWeapon) load(r *reader) {
	w.Type = r.i16()
	w.MaxCount = r.u8()
	w.WorkingCount = r.u8()
	w.Arc = r.u8()
	w.Mods = r.u16()
	w.Ammo = r.u8()
}

// ── ShipDesign(99 bytes)──────────────────────────────────
type ShipDesign struct {
	Name            string
	Size            uint8
	Type            uint8
	Shield          uint8
	Drive           uint8
	Speed           uint8
	Computer        uint8
	Armor           uint8
	Specials        [bitBytesSpecials]uint8
	Weapons         [maxShipWeapons]ShipWeapon
	Picture         uint8
	Builder         uint8
	Cost            uint16
	BaseCombatSpeed uint8
	BuildDate       uint16
}

const bitBytesSpecials = (maxShipSpecials + 7) / 8 // 5

func (d *ShipDesign) load(r *reader) {
	d.Name = r.cstr(shipNameSize)
	d.Size = r.u8()
	d.Type = r.u8()
	d.Shield = r.u8()
	d.Drive = r.u8()
	d.Speed = r.u8()
	d.Computer = r.u8()
	d.Armor = r.u8()
	for i := range d.Specials {
		d.Specials[i] = r.u8()
	}
	for i := range d.Weapons {
		d.Weapons[i].load(r)
	}
	d.Picture = r.u8()
	d.Builder = r.u8()
	d.Cost = r.u16()
	d.BaseCombatSpeed = r.u8()
	d.BuildDate = r.u16()
}

// ── Ship(129 bytes)───────────────────────────────────────
type Ship struct {
	Design            ShipDesign
	Owner             uint8
	Status            uint8
	Star              int16
	X, Y              uint16
	GroupHasNavigator uint8
	WarpSpeed         uint8
	Eta               uint8
	ShieldDamage      uint8
	DriveDamage       uint8
	ComputerDamage    uint8
	CrewLevel         uint8
	CrewExp           uint16
	Officer           int16
	DamagedSpecials   [bitBytesSpecials]uint8
	ArmorDamage       uint16
	StructureDamage   uint16
	Mission           uint8
	JustBuilt         uint8
}

func (s *Ship) load(r *reader) {
	s.Design.load(r)
	s.Owner = r.u8()
	s.Status = r.u8()
	s.Star = r.i16()
	s.X = r.u16()
	s.Y = r.u16()
	s.GroupHasNavigator = r.u8()
	s.WarpSpeed = r.u8()
	s.Eta = r.u8()
	s.ShieldDamage = r.u8()
	s.DriveDamage = r.u8()
	s.ComputerDamage = r.u8()
	s.CrewLevel = r.u8()
	s.CrewExp = r.u16()
	s.Officer = r.i16()
	for i := range s.DamagedSpecials {
		s.DamagedSpecials[i] = r.u8()
	}
	s.ArmorDamage = r.u16()
	s.StructureDamage = r.u16()
	s.Mission = r.u8()
	s.JustBuilt = r.u8()
}

// ── SettlerInfo(4 bytes,bit-packed LE)────────────────────
type SettlerInfo struct {
	SourceColony      uint8
	DestinationPlanet uint8
	Player            uint8
	Eta               uint8
	Job               uint8
}

func (s *SettlerInfo) load(r *reader) {
	// BitStream 逐 byte 依需求讀取,26 bits 落在 4 bytes 內(見 openorion2 BitStream)。
	b0 := r.u8()
	b1 := r.u8()
	b2 := r.u8()
	b3 := r.u8()
	s.SourceColony = b0
	s.DestinationPlanet = b1
	s.Player = b2 & 0xf
	s.Eta = (b2 >> 4) & 0xf
	s.Job = b3 & 0x3
}

// ── Player(3753 bytes)────────────────────────────────────
type Player struct {
	Name        string
	Race        string
	Eliminated  uint8
	Picture     uint8
	Color       uint8
	Personality uint8 // 100 = 人類玩家
	Objective   uint8

	HomePlayerId         uint16
	NetworkPlayerId      uint16
	PlayerDoneFlags      uint8
	ResearchBreakthrough uint8
	TaxRate              uint8
	BC                   int32
	TotalFreighters      uint16
	SurplusFreighters    int16
	CommandPoints        uint16
	UsedCommandPoints    int16
	FoodFreighted        uint16
	SettlersFreighted    uint16

	Settlers [maxSettlers]SettlerInfo

	TotalPop         uint16
	FoodProduced     uint16
	IndustryProduced uint16
	ResearchProduced uint16
	BcProduced       uint16
	SurplusFood      int16
	SurplusBC        int16

	TotalMaintenance     int32
	BuildingMaintenance  uint16
	FreighterMaintenance uint16
	ShipMaintenance      uint16
	SpyMaintenance       uint16
	TributeCost          uint16
	OfficerMaintenance   uint16

	ResearchTopics   [maxResearchTopics]uint8
	Techs            [maxTechnologies]uint8
	ResearchProgress uint32
	HyperTechLevels  [maxResearchAreas]uint8
	ResearchTopic    uint8
	ResearchItem     uint8

	Blueprints        [maxPlayerBlueprint]ShipDesign
	SelectedBlueprint ShipDesign

	PlayerContacts   [maxPlayers]uint8
	PlayerRelations  [maxPlayers]int8
	ForeignPolicies  [maxPlayers]uint8
	TradeTreaties    [maxPlayers]uint8
	ResearchTreaties [maxPlayers]uint8

	Traits [traitsCount]int8

	FleetHistory      [maxHistoryLength]uint8
	TechHistory       [maxHistoryLength]uint8
	PopulationHistory [maxHistoryLength]uint8
	BuildingHistory   [maxHistoryLength]uint8

	Spies         [maxPlayers]uint8
	InfoPanel     uint8
	GalaxyCharted uint8
}

func (p *Player) load(r *reader) {
	r.skip(1) // 未知
	p.Name = r.cstr(playerNameSize)
	p.Race = r.cstr(playerRaceSize)
	p.Eliminated = r.u8()
	p.Picture = r.u8()
	p.Color = r.u8()
	p.Personality = r.u8()
	p.Objective = r.u8()
	p.HomePlayerId = r.u16()
	p.NetworkPlayerId = r.u16()
	p.PlayerDoneFlags = r.u8()
	r.skip(2) // dead field
	p.ResearchBreakthrough = r.u8()
	p.TaxRate = r.u8()
	p.BC = r.i32()
	p.TotalFreighters = r.u16()
	p.SurplusFreighters = r.i16()
	p.CommandPoints = r.u16()
	p.UsedCommandPoints = r.i16()
	p.FoodFreighted = r.u16()
	p.SettlersFreighted = r.u16()
	for i := range p.Settlers {
		p.Settlers[i].load(r)
	}
	p.TotalPop = r.u16()
	p.FoodProduced = r.u16()
	p.IndustryProduced = r.u16()
	p.ResearchProduced = r.u16()
	p.BcProduced = r.u16()
	p.SurplusFood = r.i16()
	p.SurplusBC = r.i16()
	p.TotalMaintenance = r.i32()
	p.BuildingMaintenance = r.u16()
	p.FreighterMaintenance = r.u16()
	p.ShipMaintenance = r.u16()
	p.SpyMaintenance = r.u16()
	p.TributeCost = r.u16()
	p.OfficerMaintenance = r.u16()
	for i := range p.ResearchTopics {
		p.ResearchTopics[i] = r.u8()
	}
	for i := range p.Techs {
		p.Techs[i] = r.u8()
	}
	p.ResearchProgress = r.u32()
	r.skip(45)
	for i := range p.HyperTechLevels {
		p.HyperTechLevels[i] = r.u8()
	}
	r.skip(253)
	p.ResearchTopic = r.u8()
	p.ResearchItem = r.u8()
	r.skip(3) // 未知
	for i := range p.Blueprints {
		p.Blueprints[i].load(r)
	}
	p.SelectedBlueprint.load(r)
	r.skip(12)
	for i := range p.PlayerContacts {
		p.PlayerContacts[i] = r.u8()
	}
	r.skip(139)
	for i := range p.PlayerRelations {
		p.PlayerRelations[i] = r.i8()
	}
	r.skip(8)
	for i := range p.ForeignPolicies {
		p.ForeignPolicies[i] = r.u8()
	}
	for i := range p.TradeTreaties {
		p.TradeTreaties[i] = r.u8()
	}
	for i := range p.ResearchTreaties {
		p.ResearchTreaties[i] = r.u8()
	}
	r.skip(608)
	for i := range p.Traits {
		p.Traits[i] = r.i8()
	}
	r.skip(33)
	for i := range p.FleetHistory {
		p.FleetHistory[i] = r.u8()
	}
	for i := range p.TechHistory {
		p.TechHistory[i] = r.u8()
	}
	for i := range p.PopulationHistory {
		p.PopulationHistory[i] = r.u8()
	}
	for i := range p.BuildingHistory {
		p.BuildingHistory[i] = r.u8()
	}
	for i := range p.Spies {
		p.Spies[i] = r.u8()
	}
	p.InfoPanel = r.u8()
	r.skip(21)
	p.GalaxyCharted = r.u8()
	r.skip(51)
}
