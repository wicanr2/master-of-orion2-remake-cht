package shell

import (
	"fmt"

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
// `TestSeatFieldsCoverPlayerSide` 用反射盯著它。
type seat struct {
	Player              engine.PlayerState
	PlayerColonies      []engine.ColonyState
	PlayerColonyStars   []int
	PlayerColonyMarines []int
	PlayerColonyTanks   []int
	MarineBarracksAge   []int
	ArmorBarracksAge    []int
	Builds              []ColonyBuild
	BuildQueue          [][]ColonyBuild
	ColonyBuildings     []map[string]bool
	PopAccum            []int
	Ships               []Ship
	Leaders             []Leader
	MercPool            []Leader
	MercOfferedIdx      int
	PlayerSpies         []int
	Outposts            []Outpost

	FleetAtStar   int
	FleetDestStar int
	FleetETA      int
	FleetMarines  int
	FleetTanks    int
	SelectedStar  int

	RaceIndex     int
	PlayerName    string
	FlagColor     int
	RaceCombatPct int
	RaceGrowthPct int
	Government    gamedata.MoraleGovernmentType
	CapturedPop   int

	// 以下是「上一回合發生在我身上的事」。它們看起來像顯示暫態,但在熱座裡必須隨席位走:
	// 星系主畫面的產出數字、回合摘要的完工清單、事件快報、突襲/發現/戰鬥回報,都是
	// **這個帝國的**回合結果。不隨席位走的話,換人後會看到上一位玩家的戰報。
	//
	// ⚠ 刻意**不**隨席位走的:`LastCouncil`(議會是全星系新聞,所有人看到同一則)、
	// `Monsters` 與 `PersistentEvents`(怪獸守著哪顆星、超新星在倒數第幾回合,是星圖的
	// 狀態不是某個玩家的——跟著席位走會讓同一顆超新星每回合被倒數 N 次)。
	LastPlayerOutput          engine.EmpireOutput
	LastBuilt                 []string
	LastEvent                 string
	LastEventReport           *EventReport
	LastDiscovery             *SystemDiscovery
	LastAntares               string
	LastRaid                  string
	LastRaidReport            *AIRaidReport
	LastEspionage             []string
	LastBattle                *BattleResult
	AntaresRaids              int
	AntaranHomeworldConquered bool
}

// saveSeat 把目前的玩家側狀態抓成一個席位快照。
func (s *GameSession) saveSeat() seat {
	return seat{
		Player: s.Player, PlayerColonies: s.PlayerColonies, PlayerColonyStars: s.PlayerColonyStars,
		PlayerColonyMarines: s.PlayerColonyMarines, PlayerColonyTanks: s.PlayerColonyTanks,
		MarineBarracksAge: s.MarineBarracksAge, ArmorBarracksAge: s.ArmorBarracksAge,
		Builds: s.Builds, BuildQueue: s.BuildQueue, ColonyBuildings: s.ColonyBuildings,
		PopAccum: s.popAccum, Ships: s.Ships, Leaders: s.Leaders,
		MercPool: s.MercPool, MercOfferedIdx: s.MercOfferedIdx,
		PlayerSpies: s.PlayerSpies, Outposts: s.Outposts,
		FleetAtStar: s.FleetAtStar, FleetDestStar: s.FleetDestStar, FleetETA: s.FleetETA,
		FleetMarines: s.FleetMarines, FleetTanks: s.FleetTanks, SelectedStar: s.SelectedStar,
		RaceIndex: s.RaceIndex, PlayerName: s.PlayerName, FlagColor: s.FlagColor,
		RaceCombatPct: s.RaceCombatPct, RaceGrowthPct: s.raceGrowthPct,
		Government: s.Government, CapturedPop: s.CapturedPop,

		LastPlayerOutput: s.LastPlayerOutput, LastBuilt: s.LastBuilt,
		LastEvent: s.LastEvent, LastEventReport: s.LastEventReport, LastDiscovery: s.LastDiscovery,
		LastAntares: s.LastAntares, LastRaid: s.LastRaid, LastRaidReport: s.LastRaidReport,
		LastEspionage: s.LastEspionage, LastBattle: s.LastBattle,
		AntaresRaids: s.AntaresRaids, AntaranHomeworldConquered: s.AntaranHomeworldConquered,
	}
}

// loadSeat 把一個席位快照裝回玩家側狀態。
func (s *GameSession) loadSeat(v seat) {
	s.Player, s.PlayerColonies, s.PlayerColonyStars = v.Player, v.PlayerColonies, v.PlayerColonyStars
	s.PlayerColonyMarines, s.PlayerColonyTanks = v.PlayerColonyMarines, v.PlayerColonyTanks
	s.MarineBarracksAge, s.ArmorBarracksAge = v.MarineBarracksAge, v.ArmorBarracksAge
	s.Builds, s.BuildQueue, s.ColonyBuildings = v.Builds, v.BuildQueue, v.ColonyBuildings
	s.popAccum, s.Ships, s.Leaders = v.PopAccum, v.Ships, v.Leaders
	s.MercPool, s.MercOfferedIdx = v.MercPool, v.MercOfferedIdx
	s.PlayerSpies, s.Outposts = v.PlayerSpies, v.Outposts
	s.FleetAtStar, s.FleetDestStar, s.FleetETA = v.FleetAtStar, v.FleetDestStar, v.FleetETA
	s.FleetMarines, s.FleetTanks, s.SelectedStar = v.FleetMarines, v.FleetTanks, v.SelectedStar
	s.RaceIndex, s.PlayerName, s.FlagColor = v.RaceIndex, v.PlayerName, v.FlagColor
	s.RaceCombatPct, s.raceGrowthPct = v.RaceCombatPct, v.RaceGrowthPct
	s.Government, s.CapturedPop = v.Government, v.CapturedPop

	s.LastPlayerOutput, s.LastBuilt = v.LastPlayerOutput, v.LastBuilt
	s.LastEvent, s.LastEventReport, s.LastDiscovery = v.LastEvent, v.LastEventReport, v.LastDiscovery
	s.LastAntares, s.LastRaid, s.LastRaidReport = v.LastAntares, v.LastRaid, v.LastRaidReport
	s.LastEspionage, s.LastBattle = v.LastEspionage, v.LastBattle
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
	s.Seats = make([]seat, n)
	s.Seats[0] = s.saveSeat()
	// 其餘席位由後面的 AI 對手轉成真人:把該 AI 的帝國搬進席位,再從 AI 清單移除。
	for i := 1; i < n; i++ {
		ai := s.AIPlayers[len(s.AIPlayers)-1]
		s.AIPlayers = s.AIPlayers[:len(s.AIPlayers)-1]
		s.Seats[i] = seatFromAI(ai, i)
	}
	s.ActiveSeat = 0
	s.loadSeat(s.Seats[0])
	return n
}

// seatFromAI 把一個 AI 對手的帝國轉成真人席位。
//
// ⚠ 誠實簡化:`AIOpponent` 是比玩家側薄很多的模型(沒有建造佇列、領袖、間諜、前哨站…),
// 轉過來的席位那些欄位是空的——接手的真人從「有母星、有殖民地、有艦隊,但還沒開始
// 蓋東西」的狀態起步。要完全對等得先把 AIOpponent 補成完整帝國,那是另一條線。
func seatFromAI(ai AIOpponent, idx int) seat {
	v := seat{
		Player:            ai.Player,
		PlayerColonies:    ai.Colonies,
		PlayerColonyStars: append([]int(nil), ai.ColonyStars...),
		// 名字要去掉「AI (…)」外殼:接手的是真人,交接畫面寫「下一位:AI(布拉西人)」很怪。
		// 保留種族名當帝國名,玩家自己知道接的是哪一族。
		PlayerName: seatTakeoverName(idx, ai.Name),
		// ⚠ AIOpponent 沒有 RaceIndex 欄位(它的種族只以名字與性格呈現),接手的席位
		// 一律當人類(索引 0)。要對等得先讓 AIOpponent 記下自己是哪一族。
		RaceIndex:     0,
		FlagColor:     idx % len(FlagColors),
		FleetAtStar:   -1,
		FleetDestStar: -1,
		SelectedStar:  -1,
		Government:    gamedata.MoraleGovDictatorship,
	}
	if len(ai.ColonyStars) > 0 {
		v.FleetAtStar = ai.ColonyStars[0] // 艦隊擺在自己的母星
	}
	// 平行陣列補齊到殖民地數,免得後續索引越界。
	n := len(v.PlayerColonies)
	v.Builds = make([]ColonyBuild, n)
	v.BuildQueue = make([][]ColonyBuild, n)
	v.ColonyBuildings = make([]map[string]bool, n)
	v.PopAccum = make([]int, n)
	v.PlayerColonyMarines = make([]int, n)
	v.PlayerColonyTanks = make([]int, n)
	v.MarineBarracksAge = make([]int, n)
	v.ArmorBarracksAge = make([]int, n)
	return v
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
	s.prepPlayerDerived()
	s.LastPlayerOutput = engine.RunEmpireTurn(s.Player, s.coloniesForTurn())
	s.Player = s.LastPlayerOutput.Player
	s.recoverFromFamine()
	s.advanceEspionage()
	s.advanceBuilds()
	s.advanceResearch()
	s.LastDiscovery = nil
	s.advanceFleet()
	s.advanceMarines()
	s.advanceArmor()
	s.advancePopulation()
	s.advanceEvents()
	s.advanceShipRepair()
	s.advanceAntares()
	s.advanceAIRaids()
	s.advanceMercOffers()
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
