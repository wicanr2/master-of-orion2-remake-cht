package shell

import (
	"fmt"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// hotseat.go:熱座多人(多位真人同機輪流下令)。
//
// ============ 原版是怎麼做的(反組譯)============
//
//	`Save_Hotseat_Map_Info_` @ 0x88F5D 是這個系統的核心,而它**小得出乎意料**——
//	每個席位只存七個 word(星圖捲動 x/y、縮放、選中艦隊框…,stride 8),外加
//	`Save_Hotseat_Fleet_Box_Ship_` @ 0x7872E。
//
//	換句話說:原版**根本不搬帝國資料**。帝國資料本來就是 `player[i]` 陣列(stride 0xEA9),
//	「現在輪到誰」只是一個索引(`word_19999C`;同一個索引在 `Player_Troop_Anim_` @ 0xBB723
//	等處也用來選種族圖示)。換人時要存的只有**這個人看星圖時的視野狀態**。
//
//	其餘相關符號:`Hotseat_Screen_` @ 0x628E2(交接畫面)、`Draw_Hotseat_Screen_` @ 0x626D6
//	(視窗置中:`(0x280−寬)/2`、`(0x1E0−高)/2`;文字在 `x+0x0E, y+0x46`)、
//	`Get_Multi_Player_N_Humans_` @ 0x121F0、字串 `"%d Human, Hot Seat"`。
//
// ============ remake 的做法與差異(誠實標明)============
//
//	remake 的 `GameSession` 是繞著「一個玩家 + N 個 AI」長出來的:玩家側是一堆單數欄位
//	(`Player` / `PlayerColonies` / `Ships` / `FleetAtStar`…),不是陣列。要做成原版那種
//	`player[i]`,得動到幾乎每個畫面與每條回合邏輯。
//
//	所以這裡走**席位交換**:把玩家側欄位整組搬進 `seat`,換人時存回目前席位、載入下一席。
//	對遊戲規則與 UI 而言完全透明——它們永遠只看到「目前這個玩家」,與原版的
//	`player[current]` 在語意上等價,差別只在資料放哪。
//
//	⚠ 這個差異有一個真實後果:原版隨時能讀到所有玩家的狀態,remake 的非當前席位在交換
//	期間是「凍結的快照」。目前沒有任何邏輯需要跨席讀取,但要加「真人玩家之間的外交」時
//	得先正視這一點。
//
//	⚠ 未做:原版存的那七個 word 是星圖視野(捲動/縮放),remake 的星圖沒有這些狀態;
//	有對應意義的只有 `SelectedStar`,已納入席位。

// MaxHotseatSeats 是熱座席位上限。原版多人上限是 8 個帝國,remake 沿用;
// 實際可用席數另受星圖大小限制(每席要一顆母星)。
const MaxHotseatSeats = 8

// DefaultOpponents 是新遊戲的**預設** AI 對手數(帝國總數 4 = 玩家 + 3)。
// 玩家可在 NEW GAME 畫面的 PLAYERS 欄改成 2..8 個帝國(見 shell.MinEmpires/MaxEmpires),
// 這裡只是沒選過時的起始值。先前這個 3 硬編在 customrace.go / raceselect.go 兩處。
const DefaultOpponents = 3

// seat 是一位真人玩家的完整帝國狀態。
//
// 欄位 = `GameSession` 裡所有「屬於當前玩家」的欄位。**新增玩家側欄位時這裡也要加**,
// 否則換人後那個欄位會被下一位玩家繼承。這是本檔最容易出錯的地方,
// `TestSeatRoundTripKeepsEveryField` 用反射盯著它:每個 seat 欄位都要真的被
// saveSeat 存進去、被 loadSeat 裝回來。
//
// (2026-08-07:這裡原本寫著「`TestSeatFieldsCoverPlayerSide` 用反射盯著它」,
//
//	而**那支測試根本不存在**。一個指名了不存在的護欄的註解比沒有註解更危險——
//	它讓人以為這裡有人在看。改成真的寫一支。)
type seat struct {
	Player                       engine.PlayerState
	PlayerColonies               []engine.ColonyState
	PlayerCapitolPlanet          int
	PlayerCapitolPlanetKnown     bool
	PlayerCapitolRebuildRequired bool
	PlayerColonyStars            []int
	PlayerColonyPlanets          []int
	PlayerColonyMarines          []int
	PlayerColonyTanks            []int
	MarineBarracksAge            []int
	ArmorBarracksAge             []int
	Builds                       []ColonyBuild
	BuildQueue                   [][]ColonyBuild
	AutoBuild                    []bool
	RepeatBuild                  []ColonyBuild
	ColonyBuildings              []map[string]bool
	PopAccum                     []int
	// Fleets / SelectedFleet 是這位玩家的艦隊(多艦隊模型,見 fleet.go)。
	// 先前是 Ships + FleetAtStar/DestStar/ETA/Marines/Tanks 一組欄位。
	Fleets        []Fleet
	SelectedFleet int
	ShipDesigns   []ShipBlueprint
	// ColonyRelocateTo 是這位玩家各殖民地的集結點(見 relocation.go)。
	// ⚠ 它是**玩家側**狀態:漏了的話換人後下一位會繼承上一位的集結點設定。
	ColonyRelocateTo  []int
	Leaders           []Leader
	ColonyLeaderNames []string
	MercPool          []Leader
	MercOfferedIdx    int
	MercLastOfferTurn int
	PlayerSpies       []int
	PlayerSpyMissions []SpyMission
	DefensiveAgents   int
	Outposts          []Outpost

	SelectedStar int

	RaceIndex                  int
	CustomRaceTraits           uint32
	CustomRaceRuntimeTraits    [gamedata.RaceTraitCount]int8
	PlayerName                 string
	FlagColor                  int
	RaceCombatPct              int
	RaceShipDefPct             int
	RaceGroundBonus            int
	RaceSpyBonus               int
	RaceGrowthPct              int
	Government                 gamedata.MoraleGovernmentType
	CapturedPop                int
	ScoreBaseMultiplierPercent int
	LuckyEventCounter          int

	// 以下是「上一回合發生在我身上的事」。它們看起來像顯示暫態,但在熱座裡必須隨席位走:
	// 星系主畫面的產出數字、回合摘要的完工清單、事件快報、突襲/發現/戰鬥回報,都是
	// **這個帝國的**回合結果。不隨席位走的話,換人後會看到上一位玩家的戰報。
	//
	// ⚠ 刻意**不**隨席位走的:`LastCouncilNotice`(議會是全星系新聞,所有人看到同一則)、
	// `Monsters` 與 `PersistentEvents`(怪獸守著哪顆星、超新星倒數、人口暴增／瘟疫目標,是星圖的
	// 狀態不是某個玩家的——跟著席位走會讓同一顆超新星每回合被倒數 N 次)。
	LastPlayerOutput          engine.EmpireOutput
	LastBuilt                 []BuildNotice
	LastEvent                 string
	LastPersistentEventEN     string
	LastEventReport           *EventReport
	LastDiscovery             *SystemDiscovery
	LastAntaranNotice         *AntaranNotice
	LastRaidReport            *AIRaidReport
	LastEspionage             []string
	LastBankruptcy            []BankruptcyAction
	LastBattle                *BattleResult
	AntaresRaids              int
	AntaranHomeworldConquered bool
}

// saveSeat 把目前的玩家側狀態抓成一個席位快照。
func (s *GameSession) saveSeat() seat {
	return seat{
		Player: s.Player, PlayerColonies: s.PlayerColonies, PlayerColonyStars: s.PlayerColonyStars,
		PlayerCapitolPlanet: s.PlayerCapitolPlanet, PlayerCapitolPlanetKnown: s.PlayerCapitolPlanetKnown,
		PlayerCapitolRebuildRequired: s.PlayerCapitolRebuildRequired,
		PlayerColonyPlanets:          s.PlayerColonyPlanets,
		PlayerColonyMarines:          s.PlayerColonyMarines, PlayerColonyTanks: s.PlayerColonyTanks,
		MarineBarracksAge: s.MarineBarracksAge, ArmorBarracksAge: s.ArmorBarracksAge,
		Builds: s.Builds, BuildQueue: s.BuildQueue,
		AutoBuild: s.AutoBuild, RepeatBuild: s.RepeatBuild,
		ColonyBuildings: s.ColonyBuildings,
		PopAccum:        s.popAccum, Leaders: s.Leaders, ColonyLeaderNames: s.ColonyLeaderNames,
		MercPool: s.MercPool, MercOfferedIdx: s.MercOfferedIdx, MercLastOfferTurn: s.MercLastOfferTurn,
		PlayerSpies: s.PlayerSpies, PlayerSpyMissions: s.PlayerSpyMissions, DefensiveAgents: s.DefensiveAgents, Outposts: s.Outposts,
		Fleets: s.Fleets, SelectedFleet: s.SelectedFleet, ShipDesigns: s.ShipDesigns, SelectedStar: s.SelectedStar,
		ColonyRelocateTo: s.ColonyRelocateTo,
		RaceIndex:        s.RaceIndex, CustomRaceTraits: s.CustomRaceTraits,
		CustomRaceRuntimeTraits: s.CustomRaceRuntimeTraits, PlayerName: s.PlayerName, FlagColor: s.FlagColor,
		RaceCombatPct: s.RaceCombatPct, RaceGrowthPct: s.raceGrowthPct,
		RaceShipDefPct: s.RaceShipDefPct, RaceGroundBonus: s.RaceGroundBonus,
		RaceSpyBonus: s.RaceSpyBonus,
		Government:   s.Government, CapturedPop: s.CapturedPop,
		ScoreBaseMultiplierPercent: s.ScoreBaseMultiplierPercent,
		LuckyEventCounter:          s.LuckyEventCounter,

		LastPlayerOutput: s.LastPlayerOutput, LastBuilt: s.LastBuilt,
		LastEvent: s.LastEvent, LastPersistentEventEN: s.LastPersistentEventEN, LastEventReport: s.LastEventReport, LastDiscovery: s.LastDiscovery,
		LastAntaranNotice: s.LastAntaranNotice, LastRaidReport: s.LastRaidReport,
		LastEspionage: s.LastEspionage, LastBankruptcy: s.LastBankruptcy, LastBattle: s.LastBattle,
		AntaresRaids: s.AntaresRaids, AntaranHomeworldConquered: s.AntaranHomeworldConquered,
	}
}

// loadSeat 把一個席位快照裝回玩家側狀態。
func (s *GameSession) loadSeat(v seat) {
	s.Player, s.PlayerColonies, s.PlayerColonyStars = v.Player, v.PlayerColonies, v.PlayerColonyStars
	s.PlayerCapitolPlanet, s.PlayerCapitolPlanetKnown = v.PlayerCapitolPlanet, v.PlayerCapitolPlanetKnown
	s.PlayerCapitolRebuildRequired = v.PlayerCapitolRebuildRequired
	s.PlayerColonyPlanets = v.PlayerColonyPlanets
	s.PlayerColonyMarines, s.PlayerColonyTanks = v.PlayerColonyMarines, v.PlayerColonyTanks
	s.MarineBarracksAge, s.ArmorBarracksAge = v.MarineBarracksAge, v.ArmorBarracksAge
	s.Builds, s.BuildQueue, s.AutoBuild, s.RepeatBuild, s.ColonyBuildings =
		v.Builds, v.BuildQueue, v.AutoBuild, v.RepeatBuild, v.ColonyBuildings
	s.popAccum, s.Leaders, s.ColonyLeaderNames = v.PopAccum, v.Leaders, v.ColonyLeaderNames
	s.MercPool, s.MercOfferedIdx, s.MercLastOfferTurn = v.MercPool, v.MercOfferedIdx, v.MercLastOfferTurn
	s.PlayerSpies, s.PlayerSpyMissions, s.DefensiveAgents, s.Outposts = v.PlayerSpies, v.PlayerSpyMissions, v.DefensiveAgents, v.Outposts
	s.Fleets, s.SelectedFleet, s.ShipDesigns, s.SelectedStar = v.Fleets, v.SelectedFleet, v.ShipDesigns, v.SelectedStar
	s.ColonyRelocateTo = v.ColonyRelocateTo
	s.ensureFleet() // 席位可能是空的(舊存檔/新建席位),維持「至少一支艦隊」的不變量
	s.ensureBuildQueue()
	s.RaceIndex, s.CustomRaceTraits, s.CustomRaceRuntimeTraits, s.PlayerName, s.FlagColor =
		v.RaceIndex, v.CustomRaceTraits, v.CustomRaceRuntimeTraits, v.PlayerName, v.FlagColor
	s.RaceCombatPct, s.raceGrowthPct = v.RaceCombatPct, v.RaceGrowthPct
	s.RaceShipDefPct, s.RaceGroundBonus = v.RaceShipDefPct, v.RaceGroundBonus
	s.RaceSpyBonus = v.RaceSpyBonus
	s.Government, s.CapturedPop = v.Government, v.CapturedPop
	s.ScoreBaseMultiplierPercent = v.ScoreBaseMultiplierPercent
	s.LuckyEventCounter = v.LuckyEventCounter

	s.LastPlayerOutput, s.LastBuilt = v.LastPlayerOutput, v.LastBuilt
	s.LastEvent, s.LastPersistentEventEN, s.LastEventReport, s.LastDiscovery = v.LastEvent, v.LastPersistentEventEN, v.LastEventReport, v.LastDiscovery
	s.LastAntaranNotice, s.LastRaidReport = v.LastAntaranNotice, v.LastRaidReport
	s.LastEspionage, s.LastBankruptcy, s.LastBattle = v.LastEspionage, v.LastBankruptcy, v.LastBattle
	s.AntaresRaids, s.AntaranHomeworldConquered = v.AntaresRaids, v.AntaranHomeworldConquered
}

// HotseatEnabled 回傳這局是否為熱座多人(席位數 > 1)。
func (s *GameSession) HotseatEnabled() bool { return len(s.Seats) > 1 }

// SeatCount 回傳席位數(單人為 1)。
func (s *GameSession) SeatCount() int {
	if len(s.Seats) < 1 {
		return 1
	}
	return len(s.Seats)
}

// seatFallbackName 是沒取名的席位的預設稱呼。
func seatFallbackName(i int) string { return fmt.Sprintf("第 %d 位玩家", i+1) }

// seatTakeoverName 是真人接管某個 AI 帝國之後,那一席的顯示名。
func seatTakeoverName(i int, aiName string) string {
	return fmt.Sprintf("%s(%s)", seatFallbackName(i), stripAILabel(aiName))
}

// SeatName 回傳第 i 席的帝國名。
func (s *GameSession) SeatName(i int) string {
	if i == s.ActiveSeat {
		if s.PlayerName != "" {
			return s.PlayerName
		}
		return seatFallbackName(i)
	}
	if i >= 0 && i < len(s.Seats) && s.Seats[i].PlayerName != "" {
		return s.Seats[i].PlayerName
	}
	return seatFallbackName(i)
}

// SetupHotseat 把目前這局變成 n 席熱座。
//
// 第 0 席沿用目前的玩家狀態(新遊戲流程已經建好的那一份);其餘席位各自從對應的 AI 對手
// **接管**——那些帝國已經有母星、殖民地與艦隊了,直接讓真人接手,不需要另外生成。
// 這是 remake 的作法(原版是開局就依人數配置帝國);好處是不動星圖生成,壞處是可用席數
// 受 AI 對手數限制,已在 UI 明示。
//
// n <= 1 或超出可接管的對手數時什麼都不做,回傳實際席位數。
func (s *GameSession) SetupHotseat(n int) int {
	if n <= 1 {
		s.Seats, s.ActiveSeat = nil, 0
		return 1
	}
	if n > MaxHotseatSeats {
		n = MaxHotseatSeats
	}
	if avail := 1 + len(s.AIPlayers); n > avail {
		n = avail
	}
	if n <= 1 {
		s.Seats, s.ActiveSeat = nil, 0
		return 1
	}
	// 保留舊 API 的預設行為:從 AI 清單尾端接管,但實際搬運走統一的
	// 明確索引路徑,避免兩套席位資料轉換逐漸分叉。
	indices := make([]int, 0, n-1)
	for i := len(s.AIPlayers) - 1; i >= 0 && len(indices) < n-1; i-- {
		indices = append(indices, i)
	}
	return s.SetupHotseatWithAIIndices(indices)
}

// SetupHotseatWithAIIndices 讓指定的 AI 帝國依 indices 順序轉成真人席位。
// indices 是呼叫前 AIPlayers 的索引;未被選中的 AI 保留為 AI,順序不變。
// 第 0 席永遠是原本玩家,其後各席依 indices 建立。
func (s *GameSession) SetupHotseatWithAIIndices(indices []int) int {
	if len(indices) == 0 {
		s.Seats, s.ActiveSeat = nil, 0
		return 1
	}

	capSelected := len(indices)
	if capSelected > MaxHotseatSeats-1 {
		capSelected = MaxHotseatSeats - 1
	}
	selected := make([]int, 0, capSelected)
	seen := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(s.AIPlayers) || seen[idx] || len(selected) >= MaxHotseatSeats-1 {
			continue
		}
		seen[idx] = true
		selected = append(selected, idx)
	}
	if len(selected) == 0 {
		s.Seats, s.ActiveSeat = nil, 0
		return 1
	}

	oldAI := append([]AIOpponent(nil), s.AIPlayers...)
	oldSpies := append([]int(nil), s.PlayerSpies...)
	oldSpyMissions := append([]SpyMission(nil), s.PlayerSpyMissions...)
	oldRelations := s.AIRelations
	oldRelationsRaw := s.AIRelationsRaw
	oldRelationsRawKnown := s.AIRelationsRawKnown
	oldReputationRaw := s.AIReputationRaw
	oldTreatyBiasRaw := s.AITreatyBiasRaw
	oldAgreementBiasRaw := s.AIAgreementBiasRaw
	oldTributeModes := s.AITributeModes
	oldWars := s.AIWars
	oldPolicies := s.AIPolicies
	oldTrade := s.AITrade
	oldResearch := s.AIResearch

	remaining := make([]AIOpponent, 0, len(oldAI)-len(selected))
	remainingSpies := make([]int, 0, len(oldAI)-len(selected))
	remainingSpyMissions := make([]SpyMission, 0, len(oldAI)-len(selected))
	remainingOldIndices := make([]int, 0, len(oldAI)-len(selected))
	for i, a := range oldAI {
		if seen[i] {
			continue
		}
		remaining = append(remaining, a)
		remainingOldIndices = append(remainingOldIndices, i)
		if i < len(oldSpies) {
			remainingSpies = append(remainingSpies, oldSpies[i])
		} else {
			remainingSpies = append(remainingSpies, 0)
		}
		if i < len(oldSpyMissions) {
			remainingSpyMissions = append(remainingSpyMissions, normalizedSpyMission(oldSpyMissions[i]))
		} else {
			remainingSpyMissions = append(remainingSpyMissions, SpyMissionSteal)
		}
	}

	// PlayerSpies 是平行 AIPlayers 的欄位;選走幾個 AI 後,第 0 席也要同步
	// 壓縮索引,否則下一回合會把間諜送到錯的對手。
	first := s.saveSeat()
	first.PlayerSpies = remainingSpies
	first.PlayerSpyMissions = remainingSpyMissions

	s.AIPlayers = remaining
	s.PlayerSpies = remainingSpies
	s.PlayerSpyMissions = remainingSpyMissions
	s.AIRelations = filterAIRelations(oldRelations, remainingOldIndices)
	s.AIRelationsRaw = filterAIRelations(oldRelationsRaw, remainingOldIndices)
	s.AIRelationsRawKnown = filterAIBoolMatrix(oldRelationsRawKnown, remainingOldIndices)
	s.AIReputationRaw = filterAIRelations(oldReputationRaw, remainingOldIndices)
	s.AITreatyBiasRaw = filterAIRelations(oldTreatyBiasRaw, remainingOldIndices)
	s.AIAgreementBiasRaw = filterAIRelations(oldAgreementBiasRaw, remainingOldIndices)
	s.AITributeModes = filterAIRelations(oldTributeModes, remainingOldIndices)
	s.AIWars = filterAIBoolMatrix(oldWars, remainingOldIndices)
	s.AIPolicies = filterAIPolicyMatrix(oldPolicies, remainingOldIndices)
	s.AITrade = filterAIBoolMatrix(oldTrade, remainingOldIndices)
	s.AIResearch = filterAIBoolMatrix(oldResearch, remainingOldIndices)
	s.Seats = make([]seat, len(selected)+1)
	s.Seats[0] = first
	for i, oldIdx := range selected {
		s.Seats[i+1] = seatFromAI(oldAI[oldIdx], i+1, len(remaining))
	}
	s.ActiveSeat = 0
	s.loadSeat(s.Seats[0])
	return len(s.Seats)
}

func filterAIRelations(rel [][]int, kept []int) [][]int {
	if len(kept) == 0 || len(rel) == 0 {
		return nil
	}
	out := make([][]int, len(kept))
	for i, oldI := range kept {
		out[i] = make([]int, len(kept))
		for j, oldJ := range kept {
			if oldI >= 0 && oldI < len(rel) && oldJ >= 0 && oldJ < len(rel[oldI]) {
				out[i][j] = rel[oldI][oldJ]
			}
		}
	}
	return out
}

// seatFromAI 把一個 AI 對手的帝國轉成真人席位。
//
// ⚠ 誠實簡化:`AIOpponent` 仍比玩家側薄(沒有玩家建造佇列、前哨站、傭兵池與對應的
// 生產決策),因此這些欄位在接管席位中維持空值。已具備的領袖、母星建築、艦隊、殖民地
// 平行陣列與玩家間諜欄位則在轉換時保留,接手的真人不是單純的空白帝國。
func seatFromAI(ai AIOpponent, idx, remainingAICount int) seat {
	raceIdx := aiRaceIndex(ai)
	v := seat{
		Player:              ai.Player,
		PlayerColonies:      append([]engine.ColonyState(nil), ai.Colonies...),
		PlayerColonyStars:   append([]int(nil), ai.ColonyStars...),
		PlayerColonyPlanets: append([]int(nil), ai.ColonyPlanets...),
		// 名字要去掉「AI (…)」外殼:接手的是真人,交接畫面寫「下一位:AI(布拉西人)」很怪。
		// 保留種族名當帝國名,玩家自己知道接的是哪一族。
		PlayerName:        seatTakeoverName(idx, ai.Name),
		RaceIndex:         raceIdx,
		FlagColor:         idx % len(FlagColors),
		SelectedStar:      -1,
		Government:        gamedata.MoraleGovDictatorship,
		Leaders:           append([]Leader(nil), ai.Leaders...),
		PlayerSpies:       make([]int, remainingAICount),
		PlayerSpyMissions: make([]SpyMission, remainingAICount),
		DefensiveAgents:   ai.DefensiveAgents,
	}
	if len(ai.ColonyBuildings) > 0 {
		v.ColonyBuildings = make([]map[string]bool, len(ai.ColonyBuildings))
		for i, buildings := range ai.ColonyBuildings {
			v.ColonyBuildings[i] = cloneBuildings(buildings)
		}
	}
	if raceIdx >= 0 && raceIdx < len(Races) {
		r := Races[raceIdx]
		v.RaceCombatPct, v.RaceShipDefPct = r.CombatPct, r.ShipDefPct
		v.RaceGroundBonus, v.RaceSpyBonus = r.GroundCombatBonus, r.SpyBonus
		v.RaceGrowthPct = r.GrowthPct
		for i := range v.PlayerColonies {
			v.PlayerColonies[i].IndustryPerWorker += r.IndBonus
			v.PlayerColonies[i].ResearchPerScientist += r.ResBonus
			v.PlayerColonies[i].FoodPerFarmer += r.FoodBonus
			v.PlayerColonies[i].IncomePerPop += r.IncomePerPop
		}
	}
	home := -1
	if len(ai.ColonyStars) > 0 {
		home = ai.ColonyStars[0] // 艦隊擺在自己的母星
	}
	v.Fleets = []Fleet{NewFleet(home)}
	// 平行陣列補齊到殖民地數,免得後續索引越界。
	n := len(v.PlayerColonies)
	v.Builds = make([]ColonyBuild, n)
	v.BuildQueue = make([][]ColonyBuild, n)
	// 建造控制與殖民地平行。不能留 nil 等到某一台客戶端切到此席位才由
	// ensureBuildQueue 補成空 slice：主機若尚未切換，兩端的共同快照就會出現
	// nil / 空 slice 形狀差異而誤判 lockstep 分岔。
	v.AutoBuild = make([]bool, n)
	v.RepeatBuild = make([]ColonyBuild, n)
	if len(v.ColonyBuildings) != n {
		buildings := make([]map[string]bool, n)
		copy(buildings, v.ColonyBuildings[:minLen(len(v.ColonyBuildings), n)])
		v.ColonyBuildings = buildings
	}
	v.PopAccum = make([]int, n)
	v.PlayerColonyMarines = make([]int, n)
	v.PlayerColonyTanks = make([]int, n)
	v.MarineBarracksAge = make([]int, n)
	v.ArmorBarracksAge = make([]int, n)
	return v
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// aiRaceIndex 優先使用新存檔的 RaceIndex;舊存檔沒有該欄位時,從既有 AI 名稱
// 做一次相容回退。回退只辨識已知種族,未知名稱仍保持人類零值。
func aiRaceIndex(ai AIOpponent) int {
	if ai.RaceIndex != 0 {
		if ai.RaceIndex >= 0 && ai.RaceIndex < len(Races) {
			return ai.RaceIndex
		}
		return -1
	}
	for i, r := range Races {
		if strings.Contains(ai.Name, r.Name) || strings.Contains(ai.Name, r.EnName) {
			return i
		}
	}
	return 0
}

// advanceIdleSeats 讓「不是當前這一席」的真人帝國也各自過完這一回合。
//
// 這是熱座最容易被漏掉、漏掉又最致命的一段:席位交換讓非當前席位變成凍結的快照,
// 如果只結算當前席位,其他真人的殖民地永遠不長人口、建造永遠不完工、艦隊永遠停在原地——
// 表面上遊戲跑得很順,實際上只有第一位玩家在玩。
//
// 由 `EndTurn` 在**最後**呼叫(所有世界側結算都跑完之後),這樣每一席看到的 `Turn`、
// AI 陣營狀態都一致,不會因為誰先誰後而拿到不同的回合數。
//
// ⚠ 誠實列出各席位**不對稱**的地方(全部源自「世界只推進一次」這個結構):
//   - 當前席位的經濟結算在 AI 決策**之前**,其餘席位在**之後**。差一個 AI 回合的資訊。
//   - `advancePersistentEvents`(超新星倒數等)是星圖狀態,只跑一次,不逐席跑——
//     逐席跑會讓同一顆超新星一回合被倒數 N 次。
//   - 勝負判定(`advanceConquestVictory` / `advancePlayerDefeat` / `advanceAntaranVictory`)
//     與 `recordHistory` 只對當前席位跑。其餘席位打進安塔蘭母星或全滅時不會結束對局。
//     要補這一塊得先讓勝負判定吃「哪一位玩家」而不是隱含的 `s.Player`。
func (s *GameSession) advanceIdleSeats() {
	cur := s.ActiveSeat
	s.Seats[cur] = s.saveSeat()
	for i := range s.Seats {
		if i == cur {
			continue
		}
		s.loadSeat(s.Seats[i])
		s.advanceSeatEmpire()
		s.Seats[i] = s.saveSeat()
	}
	s.loadSeat(s.Seats[cur]) // 把控制權還給原本那一席
}

// advanceSeatEmpire 推進「目前載入的這個帝國」一回合的玩家側結算。
//
// 內容 = `EndTurn` 裡所有只動玩家自己的步驟,順序照抄(見 EndTurn 本體);
// 世界側(AI 決策、回合數、議會、外交漂移、歷史快照)不在此。
func (s *GameSession) advanceSeatEmpire() {
	s.preparePlayerResearchApplication()
	s.prepPlayerDerived()
	s.LastPlayerOutput = engine.RunEmpireTurnWithResearchRoller(s.Player, s.coloniesForTurn(), s.researchBreakthroughRoll)
	s.recordPlayerPlagueResearch(s.LastPlayerOutput)
	s.Player = s.LastPlayerOutput.Player
	s.resolvePlayerBankruptcy()
	s.applyPlayerResearchRaceTrait(s.LastPlayerOutput.ResearchDone)
	if s.LastPlayerOutput.ResearchDone {
		applyResearchTopicGrantCallbacks(&s.Player, s.Player.ResearchTopic)
		s.UpdatePlayerShipDesignsAfterTech()
	}
	s.recoverFromFamine()
	s.advanceEspionage()
	s.advanceBuilds()
	s.advanceResearch()
	s.LastDiscovery = nil
	s.advanceFleet()
	s.advanceMarines()
	s.advanceArmor()
	s.advancePopulation()
	s.advanceShipRepair()
	s.advanceAIRaids()
	// 其餘真人席位只產生自己的 offer；AI 決策、拒絕 cooldown 與 AI offer 是世界側，
	// 只能由主 EndTurn 推進一次，否則會依熱座人數重複消費亂數與 cooldown。
	s.advancePlayerMercOffer()
}

// AdvanceSeat 把控制權交給下一席。
//
// 回傳 (下一席索引, 是否繞回第 0 席)。繞回 true 代表「所有真人都下完令了」——
// 呼叫端這時才推進世界(AI 決策 + 回合結算),而不是每個人按一次就跑一回合。
func (s *GameSession) AdvanceSeat() (next int, wrapped bool) {
	if !s.HotseatEnabled() {
		return 0, true
	}
	s.Seats[s.ActiveSeat] = s.saveSeat()
	next = s.ActiveSeat + 1
	if next >= len(s.Seats) {
		next, wrapped = 0, true
	}
	s.ActiveSeat = next
	s.loadSeat(s.Seats[next])
	return next, wrapped
}
