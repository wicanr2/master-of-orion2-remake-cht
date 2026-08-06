package shell

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// events_persistent_test.go:持續型隨機事件。
// 手冊 p.180-181 是逐事件的規格書,期望值全部標明出處;remake 估值的部分只測界限。

// 手冊 p.181:「After it has been in effect for six turns, there is a five percent chance
// each turn that the anomaly will end.」——六回合之內絕不結束。
func TestPersistentEventMinimumSixTurns(t *testing.T) {
	s := NewDemoSession()
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentStasis, StarIndex: 0}}
	for turn := 1; turn <= persistentEventMinTurns-1; turn++ {
		s.advancePersistentEvents()
		if len(s.PersistentEvents) == 0 {
			t.Fatalf("第 %d 回合就結束了,手冊說前 %d 回合不會結束", turn, persistentEventMinTurns)
		}
	}
	// 第六回合起才開始擲 5%;跑很多回合最後總會結束(這裡只確認它「會」結束,不卡死)。
	for turn := 0; turn < 500 && len(s.PersistentEvents) > 0; turn++ {
		s.advancePersistentEvents()
	}
	if len(s.PersistentEvents) != 0 {
		t.Error("500 回合後時空異象仍未結束,5% 機率沒有接上")
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

// 手冊 p.181 超新星:倒數 6-14 回合;救不回來則該系統全滅、行星變成輻射。
func TestSupernovaCountdownAndOutcome(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	homeStar := s.PlayerColonyStarIndex(0)
	s.EndTurn() // 讓 LastPlayerOutput 有值

	// 需求設成天文數字 → 一定救不回來,驗證失敗路徑。
	s.PersistentEvents = []PersistentEvent{{
		Kind: PersistentSupernova, StarIndex: homeStar,
		Countdown: supernovaMinCountdown, ResearchNeeded: 1 << 30,
	}}
	for i := 0; i < supernovaMinCountdown; i++ {
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

// 蟲洞(手冊 p.181:「moves that fleet to their destination in a single turn」)。
func TestWormholeShortensTravelToOneTurn(t *testing.T) {
	s := NewDemoSession()
	// 沒有航行中的艦隊 → 這個好事無處可用,應回 ok=false 讓事件重抽。
	s.FleetETA, s.FleetDestStar = 0, -1
	if _, ok := s.applyWormhole(); ok {
		t.Error("沒有航行中的艦隊時不該套用蟲洞")
	}
	s.FleetDestStar, s.FleetETA = 5, 9
	msg, ok := s.applyWormhole()
	if !ok {
		t.Fatal("有長途航行的艦隊時應可套用蟲洞")
	}
	if s.FleetETA != 1 {
		t.Errorf("蟲洞後 ETA 應為 1(手冊:a single turn),實為 %d", s.FleetETA)
	}
	if !strings.Contains(msg, "1 回合") {
		t.Errorf("訊息應說明縮短為 1 回合,實為 %q", msg)
	}
}

// 超空間獸(手冊 p.181:航行中的艦隊有機率損失一艘船)。
func TestWarpBeastOnlyStrikesTravellingFleets(t *testing.T) {
	s := NewDemoSession()
	s.FleetETA = 0 // 沒在航行
	n := len(s.Ships)
	for i := 0; i < 50; i++ {
		s.warpBeastStrike()
	}
	if len(s.Ships) != n {
		t.Errorf("沒有航行中的艦隊不該損失艦艇:%d → %d", n, len(s.Ships))
	}

	// 航行中:多擲幾次總會中(20% remake 值),艦艇會減少。
	s.FleetETA = 5
	hit := false
	for i := 0; i < 200 && len(s.Ships) > 0; i++ {
		if s.warpBeastStrike() != "" {
			hit = true
			break
		}
	}
	if !hit {
		t.Error("航行中 200 次都沒被抓到,機率沒接上")
	}
}

// 怪獸入侵事件的回合門檻是手冊逐條給的,不能提前發生。
func TestInvadingMonsterTurnGates(t *testing.T) {
	gates := []struct {
		kind    gamedata.SpaceMonster
		minTurn int
	}{
		{gamedata.MonsterAmoeba, 100},
		{gamedata.MonsterCrystal, 200},
		{gamedata.MonsterHydra, 250},
		{gamedata.MonsterDragon, 300},
	}
	for _, g := range gates {
		s := NewDemoSession()
		s.Turn = g.minTurn - 1
		if _, ok := s.spawnInvadingMonster(g.kind, g.minTurn); ok {
			t.Errorf("怪獸 %d 在第 %d 回合就入侵了(手冊門檻 %d)", g.kind, s.Turn, g.minTurn)
		}
		s.Turn = g.minTurn
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

// 持續事件要能存檔往返(超新星倒數到一半讀檔不能歸零)。
func TestPersistentEventsSurviveSaveLoad(t *testing.T) {
	s := NewDemoSession()
	s.PersistentEvents = []PersistentEvent{
		{Kind: PersistentSupernova, StarIndex: 3, Turns: 2, Countdown: 7, ResearchNeeded: 500, ResearchDone: 120},
		{Kind: PersistentWarpBeast, StarIndex: -1, Turns: 4},
	}
	restored := s.snapshot().restore()
	if len(restored.PersistentEvents) != 2 {
		t.Fatalf("讀檔後持續事件剩 %d 筆,want 2", len(restored.PersistentEvents))
	}
	got := restored.PersistentEvents[0]
	if got.Countdown != 7 || got.ResearchDone != 120 || got.ResearchNeeded != 500 {
		t.Errorf("超新星進度沒存好:%+v", got)
	}
}

// 事件表:先前標「缺子系統」的那批現在應該都標成已實作(gamedata 與 shell 兩邊要一致)。
func TestNewlyImplementedEventsAreDispatched(t *testing.T) {
	s := NewDemoSession()
	s.Turn = 400 // 過所有回合門檻
	s.FleetDestStar, s.FleetETA = 5, 9
	for _, id := range []int{19, 20, 21, 22, 23, 24, 25, 26, 28} {
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
		fresh.Turn = 400
		fresh.FleetDestStar, fresh.FleetETA = 5, 9
		fresh.EndTurn() // 讓 LastPlayerOutput 有值(超新星需要)
		fresh.Turn = 400
		fresh.FleetDestStar, fresh.FleetETA = 5, 9
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
