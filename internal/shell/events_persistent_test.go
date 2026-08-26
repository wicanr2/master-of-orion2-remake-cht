package shell

import (
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// events_persistent_test.go:持續型隨機事件。
// 手冊 p.180-181 是逐事件的規格書,期望值全部標明出處;remake 估值的部分只測界限。

// IDA 1.31 sub_206A2：age 0..4 不抽骰，age 5 起每回合 1/20，age 21 強制解除。
func TestStasisOriginalLifetimeBoundaries(t *testing.T) {
	s := NewDemoSession()
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentStasis, StarIndex: 0, Turns: -1}}
	for turn := 1; turn <= 5; turn++ {
		s.advancePersistentEvents()
		if len(s.PersistentEvents) == 0 {
			t.Fatalf("第 %d 個 active turn 不得結束：%+v", turn, s.PersistentEvents)
		}
	}
	// advancePersistentEvents 先遞增再檢查；設為 19，下一次便檢查 age 20。
	s.PersistentEvents[0].Turns = 19
	for seed := int64(0); ; seed++ {
		probe := newRandStream(seed)
		if probe.Intn(20)+1 != 1 {
			s.eventRand = newRandStream(seed)
			break
		}
	}
	s.advancePersistentEvents()
	if len(s.PersistentEvents) == 0 || s.PersistentEvents[0].Turns != 20 {
		t.Fatalf("age 20 非命中應保留：%+v", s.PersistentEvents)
	}
	// age 21 不論骰值都走強制解除。
	s.advancePersistentEvents()
	if len(s.PersistentEvents) != 0 {
		t.Fatalf("age 21 應強制解除：%+v", s.PersistentEvents)
	}
}

func TestPersistentEventReportsEnglishProgress(t *testing.T) {
	s := NewDemoSession()
	homeStar := s.PlayerColonyStarIndex(0)
	s.LastPlayerOutput.Colonies = []engine.ColonyOutput{{Research: 1}}
	s.PersistentEvents = []PersistentEvent{{
		Kind: PersistentSupernova, StarIndex: homeStar, Countdown: 1, ResearchNeeded: 1,
	}}
	s.advancePersistentEvents()
	if !strings.Contains(s.LastPersistentEventEN, "stabilized") {
		t.Fatalf("持續事件應產生英文進度報告,got %q", s.LastPersistentEventEN)
	}
	if strings.Contains(s.LastPersistentEventEN, "星系") {
		t.Fatalf("持續事件英文報告不應沿用中文星系字串,got %q", s.LastPersistentEventEN)
	}
}

// 手冊 p.181:凍結的殖民地「unable to produce, grow … do not need food or cost maintenance」。
func TestStasisFreezesColonyProduction(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	homeStar := s.PlayerColonyStarIndex(0)

	// 先跑一回合當基準。
	s.EndTurn()
	baseIndustry := s.LastPlayerOutput.Colonies[0].NetIndustry
	if baseIndustry <= 0 {
		t.Fatalf("測試前提不成立:母星淨工業應 > 0,實為 %d", baseIndustry)
	}
	baseMaint := s.totalBuildingMaintenance()
	if baseMaint <= 0 {
		t.Fatalf("測試前提不成立:母星建築維護費應 > 0,實為 %d", baseMaint)
	}

	s.PersistentEvents = []PersistentEvent{{Kind: PersistentStasis, StarIndex: homeStar}}
	if !s.ColonyInStasis(0) {
		t.Fatal("母星應處於凍結狀態")
	}
	if got := s.totalBuildingMaintenance(); got != 0 {
		t.Errorf("凍結時建築維護費應為 0(手冊:cost no maintenance),實為 %d", got)
	}

	popBefore := s.PlayerColonies[0].Population
	s.EndTurn()
	if got := s.LastPlayerOutput.Colonies[0].NetIndustry; got != 0 {
		t.Errorf("凍結時淨工業應為 0(手冊:unable to produce),實為 %d", got)
	}
	if got := s.LastPlayerOutput.Colonies[0].FoodConsumed; got != 0 {
		t.Errorf("凍結時不吃食物(手冊:do not need food),實為 %d", got)
	}
	if s.PlayerColonies[0].Population != popBefore {
		t.Errorf("凍結時人口不該變動:%d → %d", popBefore, s.PlayerColonies[0].Population)
	}
	// 凍結是暫時的:殖民地本身的人口/職務資料要原封不動留著,解除後才能恢復。
	if s.PlayerColonies[0].Farmers+s.PlayerColonies[0].Workers+s.PlayerColonies[0].Scientists == 0 {
		t.Error("凍結不該把職務分配清空——那是不可逆的破壞,不是暫停")
	}
}

func TestStasisFreezesAIColonyThroughTurnConsumer(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 || len(s.AIPlayers[0].ColonyStars) == 0 {
		t.Fatal("demo 應有 AI 殖民地")
	}
	base := s.AIPlayers[0].Colonies[0]
	star := s.AIPlayers[0].ColonyStars[0]
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentStasis, StarIndex: star, Turns: -1}}
	turn := s.aiColoniesForTurn(0, s.AIPlayers[0].Colonies)
	out := engine.RunColonyTurn(turn[0])
	if out.NetIndustry != 0 || out.Research != 0 || out.FoodConsumed != 0 || out.PopGrowth != 0 {
		t.Fatalf("AI 凍結殖民地仍有產出或成長：%+v", out)
	}
	if !reflect.DeepEqual(s.AIPlayers[0].Colonies[0], base) {
		t.Fatal("AI 凍結只能改回合副本，不得破壞持久殖民地")
	}
}

func TestStasisFreezesAllOwnersAtSameStar(t *testing.T) {
	s := NewDemoSession()
	star := s.PlayerColonyStarIndex(0)
	s.AIPlayers[0].ColonyStars[0] = star
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentStasis, StarIndex: star, Turns: -1}}
	playerTurn := s.coloniesForTurn()
	aiTurn := s.aiColoniesForTurn(0, s.AIPlayers[0].Colonies)
	for owner, colony := range map[string]engine.ColonyState{"玩家": playerTurn[0], "AI": aiTurn[0]} {
		out := engine.RunColonyTurn(colony)
		if out.NetIndustry != 0 || out.Research != 0 || out.FoodConsumed != 0 || out.PopGrowth != 0 {
			t.Fatalf("同星%s殖民地未一起凍結：%+v", owner, out)
		}
	}
}

func TestStasisCanTargetAIAndInactiveHotseatEmpire(t *testing.T) {
	ev := *gamedata.RandomEventByID(25)
	s := NewDemoSession()
	if _, ok := s.applyRandomEventLocalizedToAI(ev, 0); !ok {
		t.Fatal("事件 25 應能套用到 AI 帝國")
	}
	if len(s.PersistentEvents) != 1 || s.PersistentEvents[0].Kind != PersistentStasis ||
		!colonyStarsContain(s.AIPlayers[0].ColonyStars, s.PersistentEvents[0].StarIndex) || s.PersistentEvents[0].Turns != -1 {
		t.Fatalf("AI 時空異象 record 錯誤：%+v", s.PersistentEvents)
	}

	h := NewDemoSession()
	if h.SetupHotseat(2) != 2 || len(h.Seats) < 2 {
		t.Fatal("需要兩席熱座")
	}
	wantStar := h.Seats[1].PlayerColonyStars[0]
	if _, ok := h.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireSeat, index: 1}); !ok {
		t.Fatal("事件 25 應能套用到非目前熱座帝國")
	}
	if h.ActiveSeat != 0 || len(h.PersistentEvents) != 1 || h.PersistentEvents[0].StarIndex != wantStar || h.PersistentEvents[0].Turns != -1 {
		t.Fatalf("熱座目標或目前席位失真：active=%d events=%+v", h.ActiveSeat, h.PersistentEvents)
	}
}

func TestStasisRejectsConflictingEventAtSelectedStar(t *testing.T) {
	s := NewDemoSession()
	star := s.PlayerColonyStarIndex(0)
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentSupernova, StarIndex: star}}
	if _, ok := s.startStasis(); ok {
		t.Fatal("sub_242FC：同星已有衝突事件時不得建立時空異象")
	}
	if len(s.PersistentEvents) != 1 || s.PersistentEvents[0].Kind != PersistentSupernova {
		t.Fatalf("衝突拒絕不得改寫既有事件：%+v", s.PersistentEvents)
	}
}

// 1.31 consumer：救不回來則該星所有 owner 的殖民地全滅，殖民行星變成輻射。
func TestSupernovaCountdownAndOutcome(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	homeStar := s.PlayerColonyStarIndex(0)
	s.EndTurn() // 讓 LastPlayerOutput 有值

	// 需求設成天文數字 → 一定救不回來,驗證失敗路徑。
	const countdown = 7
	s.PersistentEvents = []PersistentEvent{{
		Kind: PersistentSupernova, StarIndex: homeStar,
		Countdown: countdown, ResearchNeeded: 1 << 30,
	}}
	for i := 0; i < countdown; i++ {
		s.advancePersistentEvents()
	}
	if len(s.PersistentEvents) != 0 {
		t.Fatal("倒數結束後事件應移除")
	}
	if len(s.PlayerColonies) != 0 {
		t.Errorf("超新星應摧毀該系統所有殖民地,還剩 %d 座", len(s.PlayerColonies))
	}
	if s.Planets[homeStar].ClimateID != gamedata.RADIATED {
		t.Errorf("超新星後行星應變成輻射(手冊逐字),實為 %d", s.Planets[homeStar].ClimateID)
	}
	// 平行陣列必須同步縮短,否則之後任何依索引存取都會 panic。
	if len(s.Builds) != 0 || len(s.PlayerColonyStars) != 0 || len(s.ColonyBuildings) != 0 {
		t.Errorf("殖民地移除後平行陣列未同步:Builds=%d Stars=%d Buildings=%d",
			len(s.Builds), len(s.PlayerColonyStars), len(s.ColonyBuildings))
	}
}

func TestStartSupernovaUsesOriginalNeedFormulaAndGlobalTarget(t *testing.T) {
	s := NewDemoSession()
	delete(s.ColonyBuildings[0], CapitolBuildName)
	s.Turn = 201 // elapsed=200
	s.Difficulty = 4
	if _, ok := s.startSupernova(); !ok {
		t.Fatal("elapsed=200 且銀河有殖民地時應可建立超新星")
	}
	e := s.PersistentEvents[len(s.PersistentEvents)-1]
	if e.Countdown < 7 || e.Countdown > 11 {
		t.Fatalf("Impossible 難度倒數應為 7..11：%+v", e)
	}
	want := s.supernovaSystemResearch(e.StarIndex) * e.Countdown
	if e.ResearchNeeded != want {
		t.Fatalf("需求必須是 system RP×countdown，不得 +1：got %d want %d", e.ResearchNeeded, want)
	}
}

func TestSupernovaCanTargetAIOnlyStar(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 {
		t.Skip("需要 AI 殖民地")
	}
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].Population = 0
	}
	for ai := range s.AIPlayers {
		for ci := range s.AIPlayers[ai].Colonies {
			if ai != 0 || ci != 0 {
				s.AIPlayers[ai].Colonies[ci].Population = 0
			}
		}
	}
	s.Turn = 201
	want := s.AIPlayers[0].ColonyStars[0]
	delete(s.AIPlayers[0].ColonyBuildings[0], CapitolBuildName)
	if _, ok := s.startSupernova(); !ok {
		t.Fatal("專用 sub_23A5F 目標不得限定目前玩家")
	}
	if got := s.PersistentEvents[len(s.PersistentEvents)-1].StarIndex; got != want {
		t.Fatalf("唯一有效 AI 殖民星應被選中：got %d want %d", got, want)
	}
}

func TestSupernovaFailureDestroysAllOwnersAtStar(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 || len(s.AIPlayers[0].ColonyPlanets) == 0 {
		t.Skip("需要 AI 殖民地與行星")
	}
	star := s.PlayerColonyStarIndex(0)
	playerPlanet := s.ColonyPlanetIndex(0)
	aiPlanet := s.AIPlayers[0].ColonyPlanets[0]
	s.AIPlayers[0].ColonyStars[0] = star
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentSupernova, StarIndex: star,
		Countdown: 1, ResearchNeeded: 1 << 30}}
	s.advancePersistentEvents()
	if len(s.PlayerColonies) != 0 || len(s.AIPlayers[0].Colonies) != 0 {
		t.Fatalf("同星所有 owner 殖民地都應被摧毀：player=%d ai=%d",
			len(s.PlayerColonies), len(s.AIPlayers[0].Colonies))
	}
	for _, planet := range []int{playerPlanet, aiPlanet} {
		if planet >= 0 && planet < len(s.Planets) && s.Planets[planet].ClimateID != gamedata.RADIATED {
			t.Fatalf("被摧毀殖民地的行星 %d 應改為 Radiated", planet)
		}
	}
}

func TestSupernovaFailureDestroysInactiveHotseatColony(t *testing.T) {
	s := NewDemoSession()
	if s.SetupHotseat(2) != 2 || len(s.Seats) < 2 || len(s.Seats[1].PlayerColonies) == 0 {
		t.Skip("需要第二個熱座帝國")
	}
	star := s.PlayerColonyStarIndex(0)
	s.Seats[1].PlayerColonyStars[0] = star
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentSupernova, StarIndex: star,
		Countdown: 1, ResearchNeeded: 1 << 30}}
	s.advancePersistentEvents()
	if len(s.Seats[0].PlayerColonies) != 0 || len(s.Seats[1].PlayerColonies) != 0 || len(s.PlayerColonies) != 0 {
		t.Fatalf("同星目前與非目前熱座殖民地均應移除：seat0=%d seat1=%d loaded=%d",
			len(s.Seats[0].PlayerColonies), len(s.Seats[1].PlayerColonies), len(s.PlayerColonies))
	}
}

func TestAISupernovaResearchIsDiverted(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 {
		t.Skip("需要 AI 殖民地")
	}
	star := s.AIPlayers[0].ColonyStars[0]
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentSupernova, StarIndex: star}}
	got := s.aiColoniesForTurn(0, s.AIPlayers[0].Colonies)
	if len(got) == 0 || !got[0].ResearchDiverted {
		t.Fatal("AI 受影響殖民地 RP 也不得同時投入一般研究")
	}
}

// 搶救成功的路徑:研究量夠就解除,不會摧毀殖民地。
func TestSupernovaRescueSucceeds(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	homeStar := s.PlayerColonyStarIndex(0)
	s.EndTurn()

	s.PersistentEvents = []PersistentEvent{{
		Kind: PersistentSupernova, StarIndex: homeStar,
		Countdown: 10, ResearchNeeded: 1, // 一回合的研究就夠
	}}
	before := len(s.PlayerColonies)
	msgs := s.advancePersistentEvents()
	if len(s.PersistentEvents) != 0 {
		t.Fatal("研究量足夠時應立刻解除危機")
	}
	if len(s.PlayerColonies) != before {
		t.Error("搶救成功不該損失殖民地")
	}
	if len(msgs) == 0 || !strings.Contains(msgs[0], "解除") {
		t.Errorf("應有解除訊息,實為 %v", msgs)
	}
}

// IDA 1.31 sub_100519 → sub_FFDDA：事件建立回合立即抵達，不是 ETA=1。
func TestWormholeImmediatelyCompletesTravel(t *testing.T) {
	s := NewDemoSession()
	// 沒有航行中的艦隊 → 這個好事無處可用,應回 ok=false 讓事件重抽。
	s.Fleet().ETA, s.Fleet().DestStar = 0, -1
	if _, ok := s.applyWormhole(); ok {
		t.Error("沒有航行中的艦隊時不該套用蟲洞")
	}
	s.Fleet().DestStar, s.Fleet().ETA = 5, 9
	msg, ok := s.applyWormhole()
	if !ok {
		t.Fatal("有長途航行的艦隊時應可套用蟲洞")
	}
	if s.Fleet().ETA != 0 || s.Fleet().AtStar != 5 || s.Fleet().DestStar != -1 {
		t.Errorf("蟲洞應立即抵達：at=%d dest=%d ETA=%d", s.Fleet().AtStar, s.Fleet().DestStar, s.Fleet().ETA)
	}
	if !strings.Contains(msg, "立刻抵達") {
		t.Errorf("訊息應說明立即抵達,實為 %q", msg)
	}
}

func TestWormholeSelectionIsWeightedByShipCount(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = []Fleet{
		{AtStar: 0, DestStar: 3, ETA: 4, Ships: []Ship{{Name: "單艦"}}},
		{AtStar: 1, DestStar: 4, ETA: 5, Ships: []Ship{{Name: "甲"}, {Name: "乙"}}},
	}
	// 找一個在三艘船 reservoir 掃描後落到第二支艦隊的 seed；若實作退化成按艦隊均勻抽，
	// 這條固定亂數消費序列會失去約束力。
	for seed := int64(0); ; seed++ {
		probe, selected := newRandStream(seed), -1
		for candidate, fleet := range []int{0, 1, 1} {
			if probe.Intn(candidate+1) == 0 {
				selected = fleet
			}
		}
		if selected == 1 {
			s.eventRand = newRandStream(seed)
			break
		}
	}
	if _, ok := s.applyWormhole(); !ok {
		t.Fatal("兩支航行艦隊應有候選")
	}
	if s.Fleets[0].ETA != 4 || s.Fleets[1].ETA != 0 || s.Fleets[1].AtStar != 4 {
		t.Fatalf("應由逐船 reservoir 選中兩艦群組：%+v", s.Fleets)
	}
}

func TestWormholeSupportsAIAndInactiveHotseat(t *testing.T) {
	ev := *gamedata.RandomEventByID(28)
	ai := NewDemoSession()
	ai.AIPlayers[0].Ships = []Ship{{Name: "AI 艦", Class: "巡洋艦"}}
	ai.AIPlayers[0].FleetETA, ai.AIPlayers[0].FleetDestStar = 5, 4
	if _, ok := ai.applyRandomEventLocalizedToAI(ev, 0); !ok || ai.AIPlayers[0].FleetETA != 0 ||
		ai.AIPlayers[0].FleetStar != 4 || ai.AIPlayers[0].FleetDestStar != -1 {
		t.Fatalf("AI 蟲洞未立即抵達：%+v", ai.AIPlayers[0])
	}

	h := NewDemoSession()
	if h.SetupHotseat(2) != 2 || len(h.Seats[1].Fleets) == 0 {
		t.Fatal("需要第二個熱座帝國")
	}
	h.Seats[1].Fleets[0].Ships = []Ship{{Name: "熱座艦"}}
	h.Seats[1].Fleets[0].ETA, h.Seats[1].Fleets[0].DestStar = 6, 4
	if _, ok := h.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireSeat, index: 1}); !ok {
		t.Fatal("非目前熱座應可套用蟲洞")
	}
	f := h.Seats[1].Fleets[0]
	if h.ActiveSeat != 0 || f.ETA != 0 || f.AtStar != 4 || f.DestStar != -1 {
		t.Fatalf("熱座蟲洞或席位失真：active=%d fleet=%+v", h.ActiveSeat, f)
	}
}

func TestWormholeRejectsTransDimensionalEmpire(t *testing.T) {
	s := NewDemoSession()
	s.RaceIndex = -1
	s.CustomRaceTraits = uint32(1) << uint(gamedata.TRAIT_TRANS_DIMENSIONAL)
	s.Fleet().ETA, s.Fleet().DestStar = 4, 3
	if _, ok := s.applyWormhole(); ok || s.Fleet().ETA != 4 {
		t.Fatalf("ship+0x6D==1 的跨維度帝國不得成為事件 28 候選：fleet=%+v", *s.Fleet())
	}
}

// IDA 1.31 sub_100618：從指定帝國所有 status 1 航行船均勻抽一艘，沒有額外 20% 骰。
func TestWarpBeastOnlyStrikesTravellingFleets(t *testing.T) {
	s := twoFleetSession()
	s.Fleets[0].ETA = 0
	s.Fleets[1].ETA = 4
	stationary := len(s.Fleets[0].Ships)
	travelling := len(s.Fleets[1].Ships)
	if msg := s.warpBeastStrike(); msg == "" {
		t.Fatal("有航行艦時 active consumer 應立即拖走一艘，不得再擲 20%")
	}
	if len(s.Fleets[0].Ships) != stationary || len(s.Fleets[1].Ships) != travelling-1 {
		t.Fatalf("只能從航行艦隊移除：stationary=%d travelling=%d", len(s.Fleets[0].Ships), len(s.Fleets[1].Ships))
	}

	s.Fleets[1].ETA = 0
	n := s.ShipCount()
	if msg := s.warpBeastStrike(); msg != "" || s.ShipCount() != n {
		t.Fatalf("沒有航行艦不得損失：msg=%q ships=%d→%d", msg, n, s.ShipCount())
	}
}

func TestWarpBeastReservoirCanRemoveNonWeakestShip(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = []Fleet{{ETA: 3, Ships: []Ship{
		{Name: "弱艦", Class: "巡防艦"},
		{Name: "強艦", Class: "戰艦"},
	}}}
	for seed := int64(0); ; seed++ {
		probe := newRandStream(seed)
		_ = probe.Intn(1)
		if probe.Intn(2) == 0 {
			s.eventRand = newRandStream(seed)
			break
		}
	}
	if msg := s.warpBeastStrike(); msg == "" || len(s.Fleets[0].Ships) != 1 || s.Fleets[0].Ships[0].Name != "弱艦" {
		t.Fatalf("reservoir 應能抽走非最弱艦，不能沿用 removeWeakestShip：ships=%+v msg=%q", s.Fleets[0].Ships, msg)
	}
}

func TestWarpBeastTargetsAIAndInactiveHotseat(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].FleetETA = 3
	s.AIPlayers[0].Ships = []Ship{{Name: "AI 巡洋艦", Class: "巡洋艦"}, {Name: "AI 巡防艦", Class: "巡防艦"}}
	s.syncAIShipStrength(0)
	beforeAI := len(s.AIPlayers[0].Ships)
	e := &PersistentEvent{Kind: PersistentWarpBeast, TargetKind: eventEmpireAI, TargetIndex: 0}
	if msg, _ := s.warpBeastStrikeReport(e); msg == "" || len(s.AIPlayers[0].Ships) != beforeAI-1 {
		t.Fatalf("AI 航行艦應損失一艘：before=%d after=%d msg=%q", beforeAI, len(s.AIPlayers[0].Ships), msg)
	}

	h := NewDemoSession()
	if h.SetupHotseat(2) != 2 || len(h.Seats[1].Fleets) == 0 {
		t.Fatal("需要第二個熱座帝國與艦隊")
	}
	if len(h.Seats[1].Fleets[0].Ships) == 0 {
		h.Seats[1].Fleets[0].Ships = []Ship{{Name: "熱座巡洋艦", Class: "巡洋艦"}}
	}
	h.Seats[1].Fleets[0].ETA = 4
	beforeSeat := len(h.Seats[1].Fleets[0].Ships)
	e = &PersistentEvent{Kind: PersistentWarpBeast, TargetKind: eventEmpireSeat, TargetIndex: 1}
	if msg, _ := h.warpBeastStrikeReport(e); msg == "" || len(h.Seats[1].Fleets[0].Ships) != beforeSeat-1 {
		t.Fatalf("非目前熱座航行艦應損失一艘：before=%d after=%d msg=%q", beforeSeat, len(h.Seats[1].Fleets[0].Ships), msg)
	}
	if h.ActiveSeat != 0 {
		t.Fatalf("攻擊非目前席位不得切換 ActiveSeat：%d", h.ActiveSeat)
	}
}

func TestWarpBeastCreationPreservesEmpireTarget(t *testing.T) {
	ev := *gamedata.RandomEventByID(26)
	ai := NewDemoSession()
	if _, ok := ai.applyRandomEventLocalizedToAI(ev, 0); !ok {
		t.Fatal("事件 26 應能以 AI 帝國建立")
	}
	if len(ai.PersistentEvents) != 1 || ai.PersistentEvents[0].TargetKind != eventEmpireAI ||
		ai.PersistentEvents[0].TargetIndex != 0 || ai.PersistentEvents[0].Turns != -1 {
		t.Fatalf("AI 目標 record 錯誤：%+v", ai.PersistentEvents)
	}

	h := NewDemoSession()
	if h.SetupHotseat(2) != 2 {
		t.Fatal("需要兩席熱座")
	}
	if _, ok := h.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireSeat, index: 1}); !ok {
		t.Fatal("事件 26 應能以非目前熱座帝國建立")
	}
	if len(h.PersistentEvents) != 1 || h.PersistentEvents[0].TargetKind != eventEmpireSeat ||
		h.PersistentEvents[0].TargetIndex != 1 || h.PersistentEvents[0].Turns != -1 || h.ActiveSeat != 0 {
		t.Fatalf("熱座目標 record 或席位錯誤：active=%d events=%+v", h.ActiveSeat, h.PersistentEvents)
	}
}

func TestWarpBeastOriginalLifetimeAndNoStrikeOnEndTurn(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().ETA = 3
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentWarpBeast, StarIndex: -1, Turns: -1,
		TargetKind: eventEmpirePlayer, TargetIndex: 0}}
	for i := 0; i < 5; i++ {
		// 每回合重抽目標；將它固定回玩家，以驗證五個 active turn 都會執行 consumer。
		s.PersistentEvents[0].TargetKind, s.PersistentEvents[0].TargetIndex = eventEmpirePlayer, 0
		s.advancePersistentEvents()
		if len(s.PersistentEvents) == 0 {
			t.Fatalf("active turn %d 不得解除", i+1)
		}
	}
	// age 21 強制解除，而且解除判定先於刪艦。
	s.PersistentEvents[0].Turns = 20
	s.PersistentEvents[0].TargetKind, s.PersistentEvents[0].TargetIndex = eventEmpirePlayer, 0
	ships := s.ShipCount()
	s.advancePersistentEvents()
	if len(s.PersistentEvents) != 0 || s.ShipCount() != ships {
		t.Fatalf("age 21 應無條件解除且不得刪艦：events=%+v ships=%d→%d", s.PersistentEvents, ships, s.ShipCount())
	}
}

// 怪獸入侵事件的回合門檻是手冊逐條給的,不能提前發生。
func TestInvadingMonsterTurnGates(t *testing.T) {
	gates := []struct {
		kind    gamedata.SpaceMonster
		minTurn int
	}{
		{gamedata.MonsterAmoeba, 100},
		{gamedata.MonsterEel, 150},
		{gamedata.MonsterCrystal, 200},
		{gamedata.MonsterHydra, 250},
		{gamedata.MonsterDragon, 300},
	}
	for _, g := range gates {
		s := NewDemoSession()
		s.Turn = g.minTurn
		if _, ok := s.spawnInvadingMonster(g.kind, g.minTurn); ok {
			t.Errorf("怪獸 %d 在 elapsed=%d 前就入侵了(門檻 %d)", g.kind, s.Turn-1, g.minTurn)
		}
		s.Turn = g.minTurn + 1
		before := len(s.Monsters)
		if _, ok := s.spawnInvadingMonster(g.kind, g.minTurn); !ok {
			t.Errorf("怪獸 %d 到了第 %d 回合應該可以入侵", g.kind, g.minTurn)
			continue
		}
		if len(s.Monsters) != before+1 {
			t.Errorf("怪獸 %d 入侵後怪獸數應 +1", g.kind)
		}
	}
}

func TestInvadingMonsterTargetsVictimColonyStar(t *testing.T) {
	s := NewDemoSession()
	s.Turn = 101
	want := s.PlayerColonyStarIndex(0)
	if _, ok := s.spawnInvadingMonster(gamedata.MonsterAmoeba, 100); !ok {
		t.Fatal("有效玩家殖民地應可成為怪獸目標")
	}
	got := s.Monsters[len(s.Monsters)-1]
	if got.StarIndex != want {
		t.Fatalf("sub_23BEC 應選受害帝國殖民星：got %d want %d", got.StarIndex, want)
	}
}

func TestSpaceEelUsesIndependentEventMonsterType(t *testing.T) {
	s := NewDemoSession()
	s.Turn = 151
	if _, ok := s.spawnInvadingMonster(gamedata.MonsterEel, 150); !ok {
		t.Fatal("太空鰻事件應可建立")
	}
	m := s.Monsters[len(s.Monsters)-1]
	if m.Kind != gamedata.MonsterEel || m.Kind == gamedata.MonsterAmoeba {
		t.Fatalf("事件 22 不得再借用變形蟲：%+v", m)
	}
	st, ok := gamedata.MonsterStatsFor(m.Kind)
	if !ok || st.NameZH != "太空鰻" || st.DamageMin != 0 || st.DamageMax != 0 {
		t.Fatalf("太空鰻獨立定性資料錯誤：%+v ok=%v", st, ok)
	}
}

func TestInvadingMonsterCanTargetAIColony(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 || len(s.AIPlayers[0].ColonyStars) == 0 {
		t.Skip("需要 AI 殖民地")
	}
	want := s.AIPlayers[0].ColonyStars[0]
	result, ok := s.applyRandomEventLocalizedToAI(*gamedata.RandomEventByID(20), 0)
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("AI 怪獸事件應有雙語結果：ok=%v result=%+v", ok, result)
	}
	got := s.Monsters[len(s.Monsters)-1]
	if got.StarIndex != want || got.Kind != gamedata.MonsterCrystal {
		t.Fatalf("AI 事件應保存目標殖民星與正確種類：%+v wantStar=%d", got, want)
	}
}

func TestInvadingMonsterCanTargetInactiveHotseatSeat(t *testing.T) {
	s := NewDemoSession()
	if s.SetupHotseat(2) != 2 || len(s.Seats) < 2 || len(s.Seats[1].PlayerColonyStars) == 0 {
		t.Skip("需要第二個熱座帝國")
	}
	s.Turn = 251
	want := s.Seats[1].PlayerColonyStars[0]
	result, ok := s.applyRandomEventLocalizedToTarget(*gamedata.RandomEventByID(23),
		eventEmpireTarget{kind: eventEmpireSeat, index: 1, alive: true})
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("非目前熱座席位怪獸事件應成立：ok=%v result=%+v star=%d colonies=%+v guarded=%v stars=%d",
			ok, result, want, s.Seats[1].PlayerColonies, s.StarGuardedByMonster(want), len(s.Stars))
	}
	got := s.Monsters[len(s.Monsters)-1]
	if got.StarIndex != want || got.Kind != gamedata.MonsterHydra {
		t.Fatalf("熱座事件應保存第二席殖民星與九頭蛇：%+v wantStar=%d", got, want)
	}
}

func TestHyperspaceFluxRejectsMonsterEventCandidates(t *testing.T) {
	s := NewDemoSession()
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentHyperspaceFlux, StarIndex: -1}}
	ev := gamedata.RandomEventByID(19)
	if eventCandidateAllowed(s, ev, false, 100) {
		t.Fatal("sub_2230A 在超空間亂流期間必須拒絕事件 19–23")
	}
}

// 持續事件要能存檔往返(超新星倒數到一半讀檔不能歸零)。
func TestPersistentEventsSurviveSaveLoad(t *testing.T) {
	s := NewDemoSession()
	s.PersistentEvents = []PersistentEvent{
		{Kind: PersistentSupernova, StarIndex: 3, Turns: 2, Countdown: 7, ResearchNeeded: 500, ResearchDone: 120},
		{Kind: PersistentStasis, StarIndex: 4, Turns: 13},
		{Kind: PersistentWarpBeast, StarIndex: -1, Turns: 4, TargetKind: eventEmpireAI, TargetIndex: 2},
		{Kind: PersistentPopulationBoom, PlanetIndex: 17, Turns: 6},
		{Kind: PersistentPlague, PlanetIndex: 23, Turns: 4, ResearchNeeded: 90, ResearchDone: 31},
		{Kind: PersistentPirateActivity, StarIndex: 6, Turns: 2, Strength: 11, InitialStrength: 20},
		{Kind: PersistentHyperspaceFlux, StarIndex: -1, Turns: 7},
	}
	restored := s.snapshot().restore()
	if len(restored.PersistentEvents) != 7 {
		t.Fatalf("讀檔後持續事件剩 %d 筆,want 7", len(restored.PersistentEvents))
	}
	got := restored.PersistentEvents[0]
	if got.Countdown != 7 || got.ResearchDone != 120 || got.ResearchNeeded != 500 {
		t.Errorf("超新星進度沒存好:%+v", got)
	}
	stasis := restored.PersistentEvents[1]
	if stasis.Kind != PersistentStasis || stasis.StarIndex != 4 || stasis.Turns != 13 {
		t.Errorf("時空異象目標／年齡沒存好:%+v", stasis)
	}
	beast := restored.PersistentEvents[2]
	if beast.Kind != PersistentWarpBeast || beast.TargetKind != eventEmpireAI || beast.TargetIndex != 2 || beast.Turns != 4 {
		t.Errorf("超空間獸目標／年齡沒存好:%+v", beast)
	}
	boom := restored.PersistentEvents[3]
	if boom.Kind != PersistentPopulationBoom || boom.PlanetIndex != 17 || boom.Turns != 6 {
		t.Errorf("人口暴增目標／期間沒存好:%+v", boom)
	}
	plague := restored.PersistentEvents[4]
	if plague.Kind != PersistentPlague || plague.PlanetIndex != 23 || plague.ResearchNeeded != 90 || plague.ResearchDone != 31 {
		t.Errorf("瘟疫目標／治療進度沒存好:%+v", plague)
	}
	pirates := restored.PersistentEvents[5]
	if pirates.Kind != PersistentPirateActivity || pirates.StarIndex != 6 || pirates.Strength != 11 || pirates.InitialStrength != 20 {
		t.Errorf("海盜活動進度沒存好:%+v", pirates)
	}
	flux := restored.PersistentEvents[6]
	if flux.Kind != PersistentHyperspaceFlux || flux.StarIndex != -1 || flux.Turns != 7 {
		t.Errorf("超空間亂流進度沒存好:%+v", flux)
	}
}

// 事件表:先前標「缺子系統」的那批現在應該都標成已實作(gamedata 與 shell 兩邊要一致)。
func TestNewlyImplementedEventsAreDispatched(t *testing.T) {
	s := NewDemoSession()
	s.Turn = 400 // 過所有回合門檻
	s.Fleet().DestStar, s.Fleet().ETA = 5, 9
	for _, id := range []int{2, 9, 14, 19, 20, 21, 22, 23, 24, 25, 26, 28} {
		ev := gamedata.RandomEventByID(id)
		if ev == nil {
			t.Fatalf("事件 %d 不在表裡", id)
		}
		if !ev.Implemented {
			t.Errorf("事件 %d(%s)應已標記為已實作", id, ev.Name)
		}
		// 分派要真的有 case——沒有的話 applyRandomEvent 會落到最後的 return "", false。
		// 這裡用一個乾淨的 session 逐一試,只要求「不是因為沒有 case 而回 false」。
		fresh := NewDemoSession()
		delete(fresh.ColonyBuildings[0], CapitolBuildName)
		fresh.Turn = 400
		fresh.Player.ActiveFreighters = 100 // 事件 14 原版要求難度換算後至少有 5 點海盜強度。
		fresh.Fleet().DestStar, fresh.Fleet().ETA = 5, 9
		fresh.EndTurn() // 讓 LastPlayerOutput 有值(超新星需要)
		fresh.Turn = 400
		fresh.Fleet().DestStar, fresh.Fleet().ETA = 5, 9
		if msg, ok := fresh.applyRandomEvent(*ev); !ok || msg == "" {
			t.Errorf("事件 %d(%s)分派後沒有結果:ok=%v msg=%q", id, ev.Name, ok, msg)
		}
	}
}

// TestLongGameWithPersistentEvents 是 400 回合的端到端探針:新事件裡有幾個會**移除殖民地**
// (超新星)與**移除艦艇**(超空間獸),那正是最容易把平行陣列弄到長度不一致、之後某個
// 依索引存取的地方直接 panic 的操作。單元測試只驗一次,這裡讓它在真實回合迴圈裡跑滿。
func TestLongGameWithPersistentEvents(t *testing.T) {
	s := NewDemoSession()
	seen := map[PersistentEventKind]bool{}
	monstersSpawned := 0
	for turn := 0; turn < 400; turn++ {
		s.EndTurn()
		for _, e := range s.PersistentEvents {
			seen[e.Kind] = true
		}
		if n := len(s.Monsters); n > monstersSpawned {
			monstersSpawned = n
		}
		// 平行陣列長度不變量(殖民地被超新星摧毀後最容易破)。
		n := len(s.PlayerColonies)
		if len(s.Builds) != n || len(s.PlayerColonyStars) != n || len(s.ColonyBuildings) != n ||
			len(s.PlayerColonyMarines) != n || len(s.PlayerColonyTanks) != n || len(s.popAccum) != n {
			t.Fatalf("第 %d 回合平行陣列長度不一致:colonies=%d builds=%d stars=%d buildings=%d marines=%d tanks=%d pop=%d",
				s.Turn, n, len(s.Builds), len(s.PlayerColonyStars), len(s.ColonyBuildings),
				len(s.PlayerColonyMarines), len(s.PlayerColonyTanks), len(s.popAccum))
		}
		for j, c := range s.PlayerColonies {
			if c.Farmers+c.Workers+c.Scientists != c.Population {
				t.Fatalf("第 %d 回合殖民地 %d 職務分配 %d+%d+%d 與人口 %d 對不上",
					s.Turn, j, c.Farmers, c.Workers, c.Scientists, c.Population)
			}
		}
	}
	t.Logf("400 回合:出現過的持續事件種類 %d 種、怪獸最多同時 %d 隻、結束時殖民地 %d 座、BC=%d",
		len(seen), monstersSpawned, len(s.PlayerColonies), s.Player.BC)
	if s.Victory.Over {
		t.Logf("對局在第 %d 回合結束:%s 勝(理由碼 %d)", s.Victory.Turn, s.Victory.Winner, s.Victory.Reason)
	}
	// 玩家一旦沒有殖民地,遊戲就必須已經結束——不能無聲地繼續空轉。
	if len(s.PlayerColonies) == 0 && !s.Victory.Over {
		t.Error("玩家已無殖民地,對局卻沒有結束")
	}
}

// 玩家殖民地被清空(超新星等事件可致)要判定戰敗——先前這條路徑不可達,程式碼註解也這樣
// 寫著;超新星讓它變成可達,那個斷言就過期了。
func TestPlayerDefeatWhenAllColoniesLost(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	if s.Victory.Over {
		t.Fatal("開局不該已經結束")
	}
	s.PlayerColonies = nil
	s.Builds, s.ColonyBuildings, s.PlayerColonyStars = nil, nil, nil
	s.PlayerColonyMarines, s.PlayerColonyTanks, s.popAccum = nil, nil, nil
	s.MarineBarracksAge, s.ArmorBarracksAge = nil, nil

	s.EndTurn()
	if !s.Victory.Over {
		t.Fatal("玩家一座殖民地都不剩時應判定遊戲結束")
	}
	if s.Victory.Winner == "player" {
		t.Errorf("玩家全滅不該是玩家獲勝,Winner=%q", s.Victory.Winner)
	}
}

func TestSupernovaResearchIsDivertedFromNormalResearch(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	homeStar := s.PlayerColonyStarIndex(0)
	s.Player.ResearchProgress = 0
	s.PersistentEvents = []PersistentEvent{{
		Kind: PersistentSupernova, StarIndex: homeStar, Countdown: 10, ResearchNeeded: 1 << 30,
	}}

	s.EndTurn()
	if got := s.LastPlayerOutput.Colonies[0].Research; got <= 0 {
		t.Fatalf("超新星期間殖民地仍應產生可供搶救的 RP，got %d", got)
	}
	if got := s.LastPlayerOutput.TotalResearch; got != 0 {
		t.Fatalf("受影響星系 RP 不得投入一般研究，got %d", got)
	}
	if got := s.Player.ResearchProgress; got != 0 {
		t.Fatalf("一般研究進度不得使用同一批搶救 RP，got %d", got)
	}
	if len(s.PersistentEvents) != 1 || s.PersistentEvents[0].ResearchDone != s.LastPlayerOutput.Colonies[0].Research {
		t.Fatalf("搶救進度應收到完整殖民地 RP：event=%+v output=%+v", s.PersistentEvents, s.LastPlayerOutput.Colonies[0])
	}

	s.PersistentEvents = nil
	s.EndTurn()
	if s.LastPlayerOutput.TotalResearch <= 0 || s.Player.ResearchProgress <= 0 {
		t.Fatalf("事件解除後一般研究應恢復：total=%d progress=%d", s.LastPlayerOutput.TotalResearch, s.Player.ResearchProgress)
	}
}
