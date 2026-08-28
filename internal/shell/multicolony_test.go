package shell

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// multicolony_test.go:同一個星系裡的多個殖民地。
//
// 這是「一星多行星」升格之後才問得出來的問題。先前的閘是「該星已有歸屬就不可拓殖」——
// 一星一行星時那句話等價於「沒空位了」,多天體之後它會擋掉原版最基本的擴張手段:
// 在自己的星系裡拓殖第二顆行星。
//
// 手冊 p.61:「A Colony Ship can establish a colonial foothold on any uncolonized planet
// in its range」——條件是**那顆行星**沒被殖民,不是那顆星沒被殖民。

// twoHabitableSystem 造一顆有兩顆可殖民行星的星,回傳 (session, 星, 行星A, 行星B)。
func twoHabitableSystem(t *testing.T) (*GameSession, int, int, int) {
	t.Helper()
	s := NewDemoSession()
	star := -1
	for i := range s.Stars {
		if s.Stars[i].Owner == 0 && len(s.PlanetsAt(i)) >= 2 {
			star = i
			break
		}
	}
	if star < 0 {
		// demo 星圖不保證有現成的雙行星無主星系,就地造一個:借用某顆無主星的兩格軌道。
		for i := range s.Stars {
			if s.Stars[i].Owner == 0 && len(s.PlanetsAt(i)) >= 1 {
				star = i
				break
			}
		}
		if star < 0 {
			t.Fatal("找不到無主星")
		}
		s.Planets = append(s.Planets, Planet{Name: "增建天體"})
		extra := len(s.Planets) - 1
		o := s.Stars[star].Orbits
		placed := false
		for i := range o {
			if o[i] == OrbitEmpty {
				o[i] = extra
				placed = true
				break
			}
		}
		if !placed {
			t.Fatal("該星系軌道已滿,無法造第二顆行星")
		}
		s.Stars[star].Orbits = o
	}
	ps := s.PlanetsAt(star)
	a, b := ps[0], ps[1]
	for _, p := range []int{a, b} {
		s.Planets[p] = Planet{
			Name: "測試行星", Gen: planetGenVersion, TypeID: gamedata.HABITABLE,
			ClimateID: gamedata.TERRAN, GravityID: gamedata.NORMAL_G,
			MineralID: gamedata.ABUNDANT, SizeID: gamedata.MEDIUM_PLANET,
		}
	}
	s.Planets[a].Name = "測試行星 I"
	s.Planets[b].Name = "測試行星 II"
	s.Fleet().AtStar, s.Fleet().ETA = star, 0
	return s, star, a, b
}

// 同一個星系可以有兩個殖民地(不同行星)。
func TestTwoColoniesInOneSystem(t *testing.T) {
	s, star, a, b := twoHabitableSystem(t)
	s.Fleet().Ships = []Ship{{Class: ColonyShipClass}, {Class: ColonyShipClass}}

	first := s.ColonizePlanet(a)
	if !first.Ok {
		t.Fatalf("第一顆行星拓殖失敗:%s", first.Reason)
	}
	if s.Stars[star].Owner != 1 {
		t.Fatalf("拓殖後該星應歸玩家,實得 %d", s.Stars[star].Owner)
	}

	second := s.ColonizePlanet(b)
	if !second.Ok {
		t.Fatalf("同星系第二顆行星拓殖失敗(手冊 p.61 允許):%s", second.Reason)
	}
	if first.ColonyIndex == second.ColonyIndex {
		t.Fatal("兩個殖民地不該是同一個索引")
	}
	if got := s.ColonyPlanetIndex(first.ColonyIndex); got != a {
		t.Errorf("殖民地 %d 應在行星 %d,實得 %d", first.ColonyIndex, a, got)
	}
	if got := s.ColonyPlanetIndex(second.ColonyIndex); got != b {
		t.Errorf("殖民地 %d 應在行星 %d,實得 %d", second.ColonyIndex, b, got)
	}
	// 兩個殖民地在同一顆星,名字必須取自行星——否則玩家看到兩個同名殖民地。
	if s.ColonyName(first.ColonyIndex) == s.ColonyName(second.ColonyIndex) {
		t.Errorf("同星系的兩個殖民地不該同名(都叫 %q)", s.ColonyName(first.ColonyIndex))
	}
}

// 同一顆行星不能殖民兩次。
func TestSamePlanetCannotBeColonizedTwice(t *testing.T) {
	s, _, a, _ := twoHabitableSystem(t)
	s.Fleet().Ships = []Ship{{Class: ColonyShipClass}, {Class: ColonyShipClass}}

	if res := s.ColonizePlanet(a); !res.Ok {
		t.Fatalf("第一次拓殖失敗:%s", res.Reason)
	}
	if res := s.ColonizePlanet(a); res.Ok {
		t.Error("同一顆行星不該能殖民兩次")
	}
}

// 敵方的星系仍然不能拓殖(那要打下來)。
func TestEnemySystemStillCannotBeColonized(t *testing.T) {
	s, star, a, _ := twoHabitableSystem(t)
	s.Fleet().Ships = []Ship{{Class: ColonyShipClass}}
	s.Stars[star].Owner = 2

	if res := s.ColonizePlanet(a); res.Ok {
		t.Error("敵方星系不該能拓殖")
	}
}

// ColonizeStar 是「挑該星系第一顆可殖民行星」的捷徑,兩次呼叫應落在不同行星上。
func TestColonizeStarPicksNextFreePlanet(t *testing.T) {
	s, star, a, b := twoHabitableSystem(t)
	s.Fleet().Ships = []Ship{{Class: ColonyShipClass}, {Class: ColonyShipClass}}

	first := s.ColonizeStar(star)
	if !first.Ok {
		t.Fatalf("第一次拓殖失敗:%s", first.Reason)
	}
	second := s.ColonizeStar(star)
	if !second.Ok {
		t.Fatalf("第二次拓殖失敗:%s", second.Reason)
	}
	p1, p2 := s.ColonyPlanetIndex(first.ColonyIndex), s.ColonyPlanetIndex(second.ColonyIndex)
	if p1 == p2 {
		t.Fatalf("兩次拓殖落在同一顆行星 %d", p1)
	}
	if (p1 != a && p1 != b) || (p2 != a && p2 != b) {
		t.Errorf("拓殖的行星應是 %d/%d,實得 %d/%d", a, b, p1, p2)
	}
}

// 前哨站與殖民地可以共存於同一個星系的不同天體上,而且**前哨站不會被別的行星的殖民地吃掉**
// (手冊 p.85 的改建規則是「在前哨站所在的那顆天體上建殖民地」)。
func TestOutpostOnAnotherBodySurvivesColonizingNeighbour(t *testing.T) {
	s, _, a, b := twoHabitableSystem(t)
	// 行星 b 改成氣態巨星:那是前哨站的地盤,不能殖民。
	s.Planets[b].TypeID = gamedata.GAS_GIANT
	s.Fleet().Ships = []Ship{{Class: OutpostShipClass}, {Class: ColonyShipClass}}

	if res := s.BuildOutpostOnPlanet(b); !res.Ok {
		t.Fatalf("在氣態巨星建前哨站失敗:%s", res.Reason)
	}
	res := s.ColonizePlanet(a)
	if !res.Ok {
		t.Fatalf("同星系另一顆行星拓殖失敗:%s", res.Reason)
	}
	if !s.HasOutpostOnPlanet(b) {
		t.Error("鄰居行星建了殖民地,不該把氣態巨星上的前哨站吃掉")
	}
	if s.ColonyHasBuilding(res.ColonyIndex, OutpostMarineBarracks) {
		t.Error("前哨站不在這顆行星上,不該憑空給新殖民地一座海軍陸戰隊營")
	}
}

// 存檔要記得住「殖民地在哪顆行星」——不記的話讀檔後同星系的兩個殖民地會併回同一顆行星。
func TestColonyPlanetSurvivesSaveRoundTrip(t *testing.T) {
	s, _, a, b := twoHabitableSystem(t)
	s.Fleet().Ships = []Ship{{Class: ColonyShipClass}, {Class: ColonyShipClass}}
	first := s.ColonizePlanet(a)
	second := s.ColonizePlanet(b)
	if !first.Ok || !second.Ok {
		t.Fatalf("前置拓殖失敗:%s / %s", first.Reason, second.Reason)
	}

	path := filepath.Join(t.TempDir(), "multicolony.json")
	if err := s.Save(path); err != nil {
		t.Fatalf("存檔失敗:%v", err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀檔失敗:%v", err)
	}
	if p := got.ColonyPlanetIndex(first.ColonyIndex); p != a {
		t.Errorf("讀檔後殖民地 %d 應在行星 %d,實得 %d", first.ColonyIndex, a, p)
	}
	if p := got.ColonyPlanetIndex(second.ColonyIndex); p != b {
		t.Errorf("讀檔後殖民地 %d 應在行星 %d,實得 %d", second.ColonyIndex, b, p)
	}
}

// 舊存檔(沒有 PlayerColonyPlanets)讀進來要退回「該星的代表行星」,不能變成行星 0。
//
// ⚠ 這正是 Star.Wormhole / ColonyRelocateTo 踩過兩次的零值陷阱:索引型欄位的 Go 零值 0
// 是一個**合法索引**,不是「沒有」。
func TestLegacySaveWithoutColonyPlanetsFallsBackToStar(t *testing.T) {
	s := NewDemoSession()
	s.PlayerColonyPlanets = nil // 模擬舊存檔

	want := s.PlanetAt(s.PlayerColonyStarIndex(0))
	if got := s.ColonyPlanetIndex(0); got != want {
		t.Errorf("舊存檔應退回該星的代表行星 %d,實得 %d", want, got)
	}
}

// 人造行星把氣態巨星改造完之後,**那顆天體要能真的建殖民地**——這正是先前被
// 「一星一行星」擋住的那一步(改造完沒有地方能放第二個殖民地)。
func TestArtificialPlanetBecomesColonizable(t *testing.T) {
	s, star, a, b := twoHabitableSystem(t)
	s.Planets[b].TypeID = gamedata.GAS_GIANT
	s.Fleet().Ships = []Ship{{Class: ColonyShipClass}, {Class: ColonyShipClass}}

	first := s.ColonizePlanet(a)
	if !first.Ok {
		t.Fatalf("前置:第一顆行星拓殖失敗:%s", first.Reason)
	}
	if res := s.ColonizePlanet(b); res.Ok {
		t.Fatal("測試前提不成立:氣態巨星在改造前就能殖民")
	}

	got, ok := s.BuildArtificialPlanet(first.ColonyIndex)
	if !ok {
		t.Fatal("同星系有氣態巨星,人造行星應可建造")
	}
	if got != b {
		t.Fatalf("改造的應是氣態巨星 %d,實得 %d", b, got)
	}
	res := s.ColonizePlanet(b)
	if !res.Ok {
		t.Fatalf("改造後的人造行星應可殖民:%s", res.Reason)
	}
	if len(s.PlanetsAt(star)) < 2 {
		t.Fatal("改造不該讓天體從軌道表上消失")
	}
}

// AI 擴張時 ColonyPlanets 要跟著 ColonyStars 一起長——兩者長度一旦分家,
// 之後任何「這個 AI 殖民地在哪顆行星」的查詢都會錯位。
func TestAIExpansionKeepsColonyPlanetsInSync(t *testing.T) {
	s := NewDemoSession()
	expanded := false
	for turn := 0; turn < 60; turn++ {
		s.EndTurn()
		for i := range s.AIPlayers {
			a := s.AIPlayers[i]
			if len(a.ColonyPlanets) != len(a.ColonyStars) {
				t.Fatalf("回合 %d:AI %d 的 ColonyPlanets(%d) != ColonyStars(%d)",
					turn, i, len(a.ColonyPlanets), len(a.ColonyStars))
			}
			for j, p := range a.ColonyPlanets {
				if p < 0 || p >= len(s.Planets) {
					t.Fatalf("回合 %d:AI %d 殖民地 %d 的行星索引 %d 越界", turn, i, j, p)
				}
				if star := s.PlanetStar(p); star != a.ColonyStars[j] {
					t.Fatalf("回合 %d:AI %d 殖民地 %d 的行星 %d 在星 %d,但登記在星 %d",
						turn, i, j, p, star, a.ColonyStars[j])
				}
			}
			if len(a.ColonyStars) > 1 {
				expanded = true
			}
		}
	}
	if !expanded {
		t.Fatal("測試前提不成立:60 回合內沒有任何 AI 擴張過,這支測試等於什麼都沒驗")
	}
}

// --- AI 側的同星系多殖民地 ---

// aiOnlySession 造一個「只剩一個 AI、而且銀河裡沒有任何無主星」的局面:
// 這樣 aiExpand 唯一可能的去處就是它自己的星系。
func aiOnlySession(t *testing.T) *GameSession {
	t.Helper()
	s := NewDemoSession()
	s.AIPlayers = s.AIPlayers[:1]
	home := s.AIPlayers[0].ColonyStars[0]
	for i := range s.Stars {
		if i == home {
			continue
		}
		s.Stars[i].Owner = 1 // 其餘全部劃給玩家,斷掉「往外擴」這條路
	}
	// AI 母星系至少要有兩顆可殖民的天體,否則沒得驗。
	ps := s.PlanetsAt(home)
	if len(ps) < 2 {
		s.Planets = append(s.Planets, Planet{})
		extra := len(s.Planets) - 1
		o := s.Stars[home].Orbits
		placed := false
		for i := range o {
			if o[i] == OrbitEmpty {
				o[i] = extra
				placed = true
				break
			}
		}
		if !placed {
			t.Fatal("AI 母星系軌道已滿,無法造第二顆行星")
		}
		s.Stars[home].Orbits = o
		ps = s.PlanetsAt(home)
	}
	for _, p := range ps {
		s.Planets[p] = Planet{
			Name: "AI 天體", Gen: planetGenVersion, TypeID: gamedata.HABITABLE,
			ClimateID: gamedata.TERRAN, GravityID: gamedata.NORMAL_G,
			MineralID: gamedata.ABUNDANT, SizeID: gamedata.MEDIUM_PLANET,
		}
	}
	// 母星的殖民地登記在第一顆天體上。
	s.AIPlayers[0].ColonyPlanets = []int{ps[0]}
	return s
}

// AI 也能在自己的星系裡拓殖第二顆行星(先前 aiExpand 只找 Owner==0 的星,永遠做不到)。
func TestAIColonizesSecondPlanetInOwnSystem(t *testing.T) {
	s := aiOnlySession(t)
	home := s.AIPlayers[0].ColonyStars[0]
	before := len(s.AIPlayers[0].Colonies)
	ownedBefore := s.AIPlayers[0].OwnedStars

	s.aiExpand(0)
	a := s.AIPlayers[0]
	if len(a.Colonies) != before+1 {
		t.Fatalf("AI 應在自己的星系裡多出一個殖民地,殖民地數 %d → %d", before, len(a.Colonies))
	}
	if a.ColonyStars[len(a.ColonyStars)-1] != home {
		t.Errorf("新殖民地應在母星系 %d,實得 %d", home, a.ColonyStars[len(a.ColonyStars)-1])
	}
	// 同星系再殖民不會讓版圖變大——OwnedStars 若跟著加,征服勝利判定會被灌水。
	if a.OwnedStars != ownedBefore {
		t.Errorf("在自己的星系裡殖民不該增加持有星數(%d → %d)", ownedBefore, a.OwnedStars)
	}
	// 兩個殖民地必須在不同的行星上。
	if a.ColonyPlanets[0] == a.ColonyPlanets[1] {
		t.Errorf("兩個 AI 殖民地疊在同一顆行星 %d 上", a.ColonyPlanets[0])
	}
}

func TestAIPicksHighestValuePlanetWithinCurrentSystem(t *testing.T) {
	s := aiOnlySession(t)
	home := s.AIPlayers[0].ColonyStars[0]
	// aiOnlySession 只保證一顆未殖民天體；再補一顆，才能驗證嚴格最高分而非第一顆。
	extra := len(s.Planets)
	s.Planets = append(s.Planets, Planet{Name: "AI 高價天體", Gen: planetGenVersion, TypeID: gamedata.HABITABLE,
		ClimateID: gamedata.TERRAN, GravityID: gamedata.NORMAL_G, MineralID: gamedata.ABUNDANT,
		SizeID: gamedata.MEDIUM_PLANET})
	orbits := s.Stars[home].Orbits
	placed := false
	for orbit := range orbits {
		if orbits[orbit] == OrbitEmpty {
			orbits[orbit], placed = extra, true
			break
		}
	}
	if !placed {
		t.Fatal("測試星系缺少空軌道")
	}
	s.Stars[home].Orbits = orbits
	available := make([]int, 0, 4)
	for _, planet := range s.PlanetsAt(home) {
		if !s.PlanetColonized(planet) && s.Planets[planet].TypeID == gamedata.HABITABLE {
			available = append(available, planet)
		}
	}
	if len(available) < 2 {
		t.Fatalf("fixture 應有兩顆可比較行星，得到 %v", available)
	}
	// 完全同分時 sub_E65F8 從軌道 4 反掃且只接受嚴格較高，應保留較高軌道。
	if tie, _ := s.bestAIColonizablePlanet(0, home); tie != available[len(available)-1] {
		t.Fatalf("同分應保留較高軌道行星 %d，得到 %d", available[len(available)-1], tie)
	}
	// 讓後一軌道明顯優於前一軌道，驗證不是 FirstColonizablePlanet。
	s.Planets[available[0]].SizeID, s.Planets[available[0]].MineralID = gamedata.TINY_PLANET, gamedata.ULTRA_POOR
	s.Planets[available[1]].SizeID, s.Planets[available[1]].MineralID = gamedata.HUGE_PLANET, gamedata.ULTRA_RICH
	want, _ := s.bestAIColonizablePlanet(0, home)
	s.aiExpand(0)
	a := &s.AIPlayers[0]
	if got := a.ColonyPlanets[len(a.ColonyPlanets)-1]; got != want {
		t.Fatalf("同星系應殖民嚴格最高分行星 %d，得到 %d", want, got)
	}
}

// AI 不能往玩家的星系拓殖(那要打下來)。
func TestAIWillNotExpandIntoPlayerSystem(t *testing.T) {
	s := NewDemoSession()
	playerStar := s.PlayerColonyStarIndex(0)
	if s.aiCanExpandInto(0, playerStar) {
		t.Errorf("AI 不該能往玩家的星系 %d 拓殖", playerStar)
	}
}

// AI 也不能往**另一個 AI** 的星系拓殖——Star.Owner 只分「無主/玩家/AI」,分不出是哪一個 AI,
// 所以這條要靠各自的 ColonyStars 清單判斷。
func TestAIWillNotExpandIntoAnotherAISystem(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 2 {
		t.Skip("demo 局的 AI 少於兩家")
	}
	other := s.AIPlayers[1].ColonyStars[0]
	if s.aiCanExpandInto(0, other) {
		t.Errorf("AI 0 不該能往 AI 1 的星系 %d 拓殖", other)
	}
	if !s.aiCanExpandInto(1, other) {
		t.Errorf("AI 1 應該能在自己的星系 %d 裡再拓殖", other)
	}
}

// 玩家與 AI 不能把殖民地疊在同一顆行星上——PlanetColonized 要看見**所有帝國**的殖民地。
func TestPlanetColonizedSeesAIColonies(t *testing.T) {
	s := NewDemoSession()
	aiPlanet := s.AIPlayers[0].ColonyPlanets[0]
	if !s.PlanetColonized(aiPlanet) {
		t.Fatalf("AI 母星所在的行星 %d 應被視為已殖民", aiPlanet)
	}
	star := s.PlanetStar(aiPlanet)
	if got := s.FirstColonizablePlanet(star); got == aiPlanet {
		t.Errorf("可殖民天體不該挑到 AI 已經佔著的行星 %d", aiPlanet)
	}
}

// 打下敵方星系裡的**一個**殖民地,不該把整顆星判給玩家——同星系可能還有那個 AI 的另一個
// 殖民地。星若提早翻面,剩下的敵方殖民地會變成「站在玩家星系裡的敵軍」,
// 星圖顏色與可入侵性都對不上。
func TestInvadingOneOfTwoColoniesLeavesStarEnemy(t *testing.T) {
	s := aiOnlySession(t)
	a := &s.AIPlayers[0]
	home := a.ColonyStars[0]
	ps := s.PlanetsAt(home)

	// 在同一顆星上再給這個 AI 一個殖民地(第二顆天體)。
	second, ok, _ := s.newColonyFromPlanet(ps[1], gamedata.MoraleGovDictatorship, 0, 0, 0)
	if !ok {
		t.Fatal("測試前置:第二顆天體應可建殖民地")
	}
	a.Colonies = append(a.Colonies, second)
	a.ColonyStars = append(a.ColonyStars, home)
	a.ColonyPlanets = append(a.ColonyPlanets, ps[1])
	a.ColonyBuildings = append(a.ColonyBuildings, map[string]bool{})
	s.Stars[home].Owner = 2

	// 壓倒性的登陸部隊,確保一定打贏(這裡驗的是歸屬邏輯不是戰鬥解算)。
	s.Fleet().AtStar, s.Fleet().ETA = home, 0
	s.Fleet().Marines = 500

	res := s.InvadeColony(home)
	if !res.Ok || !res.AttackerWon {
		t.Fatalf("500 名陸戰隊應打得下來:ok=%v won=%v %s", res.Ok, res.AttackerWon, res.Reason)
	}
	if res.StarCaptured {
		t.Error("星上還有敵方殖民地,不該回報「整顆星拿下」")
	}
	if s.Stars[home].Owner != 2 {
		t.Errorf("星上還有敵方殖民地,歸屬不該翻面(實得 Owner=%d)", s.Stars[home].Owner)
	}
	if len(a.Colonies) != 1 {
		t.Fatalf("AI 應只剩一個殖民地,實得 %d", len(a.Colonies))
	}
	// 過戶過來的殖民地要落在**被打下來的那顆行星**上,不是該星系的代表行星。
	got := s.ColonyPlanetIndex(len(s.PlayerColonies) - 1)
	if got != ps[0] {
		t.Errorf("過戶的殖民地應在行星 %d,實得 %d", ps[0], got)
	}

	// 再打一次:這次是最後一個,星才該翻面。
	s.Fleet().Marines = 500
	res2 := s.InvadeColony(home)
	if !res2.Ok || !res2.AttackerWon {
		t.Fatalf("第二次入侵應成功:%s", res2.Reason)
	}
	if !res2.StarCaptured || s.Stars[home].Owner != 1 {
		t.Errorf("打完最後一個殖民地,整顆星才該歸玩家(StarCaptured=%v Owner=%d)",
			res2.StarCaptured, s.Stars[home].Owner)
	}
}
