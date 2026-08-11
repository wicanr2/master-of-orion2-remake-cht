package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// routeTestSession 建一個座標好算的最小星圖:大小檔位固定、星放在指定的正規化位置。
//
// 星數取 GalaxyStarCounts[1](= 檔位 1,跨距 25.3 × 20 秒差距),因為那是有原版存檔
// (SAVE10.GAM)佐證的那一檔。
func routeTestSession(pos [][2]float64) *GameSession {
	s := &GameSession{}
	n := gamedata.GalaxyStarCounts[1]
	s.Stars = make([]Star, n)
	for i := range s.Stars {
		s.Stars[i] = Star{X: 0.999, Y: 0.999, Wormhole: -1} // 預設堆在角落,不干擾判定
	}
	for i, p := range pos {
		if i < len(s.Stars) {
			s.Stars[i] = Star{X: p[0], Y: p[1], Wormhole: -1}
		}
	}
	return s
}

// TestPointToSegmentDistUsesSegmentNotLine 釘住「線段」而不是「無限直線」。
//
// 差別是實質的:目的星**之外**的延長線上有黑洞,不該擋住這趟航程。
func TestPointToSegmentDistUsesSegmentNotLine(t *testing.T) {
	// 線段 (0,0)-(10,0);點 (20,0) 在延長線上,距離應為 10 而不是 0。
	if got := pointToSegmentDist(0, 0, 10, 0, 20, 0); got != 10 {
		t.Errorf("延長線上的點應算到端點距離 10,實得 %.3f", got)
	}
	// 垂足落在線段內:距離就是垂距。
	if got := pointToSegmentDist(0, 0, 10, 0, 5, 3); got != 3 {
		t.Errorf("垂足在線段內時應為垂距 3,實得 %.3f", got)
	}
	// 退化線段(兩端同點)不能除以零。
	if got := pointToSegmentDist(4, 4, 4, 4, 4, 7); got != 3 {
		t.Errorf("退化線段應退回點距 3,實得 %.3f", got)
	}
}

// TestRouteBlockedByBlackHole 釘住手冊那條:黑洞 2 秒差距內不可通行。
func TestRouteBlockedByBlackHole(t *testing.T) {
	w, h := gamedata.GalaxyParsecSpan(1)
	// 起點 (0,0)、終點正東方 12 秒差距處。
	s := routeTestSession([][2]float64{{0, 0}, {12 / w, 0}})
	if got := s.ParsecsBetweenStars(0, 1); got != 12 {
		t.Fatalf("測試星圖設定錯誤:兩星應相距 12 秒差距,實得 %d", got)
	}

	// 黑洞放在航線正中央偏北 1 秒差距 → 在 2 秒差距內,應該擋住。
	s.Stars[2] = Star{X: 6 / w, Y: 1 / h, Spectral: blackHoleSpectral, Wormhole: -1}
	if !s.RouteBlockedByBlackHole(0, 1) {
		t.Error("黑洞離航線 1 秒差距,應該擋住")
	}
	// 移到 3 秒差距外 → 不該擋。
	s.Stars[2].Y = 3 / h
	if s.RouteBlockedByBlackHole(0, 1) {
		t.Error("黑洞離航線 3 秒差距,不該擋住")
	}
}

// TestRouteBlackHoleIgnoresEndpoints:起訖點本身是黑洞不算「路過」。
//
// 擋的是**經過**——玩家自己選了黑洞當目的地是另一回事。
func TestRouteBlackHoleIgnoresEndpoints(t *testing.T) {
	w, _ := gamedata.GalaxyParsecSpan(1)
	s := routeTestSession([][2]float64{{0, 0}, {12 / w, 0}})
	s.Stars[1].Spectral = blackHoleSpectral
	if s.RouteBlockedByBlackHole(0, 1) {
		t.Error("目的星本身是黑洞不該被判成『路過黑洞』")
	}
}

// TestRouteBlackHoleNavigatorExemption 釘住手冊的括號:有領航員就不受限。
func TestRouteBlackHoleNavigatorExemption(t *testing.T) {
	w, h := gamedata.GalaxyParsecSpan(1)
	s := routeTestSession([][2]float64{{0, 0}, {12 / w, 0}})
	s.Fleets = []Fleet{{Ships: []Ship{{Name: "測試艦"}}, AtStar: 0, DestStar: -1}}
	s.Stars[2] = Star{X: 6 / w, Y: 1 / h, Spectral: blackHoleSpectral, Wormhole: -1}
	if !s.RouteBlockedByBlackHole(0, 1) {
		t.Fatal("前提不成立:這條航線本來就該被擋")
	}
	s.Leaders = append(s.Leaders, Leader{Name: "領航員", Skill: navigatorSkillLabel, Ship: true, Tier: 1})
	if len(s.Fleet().Ships) == 0 || !s.AssignOfficerToShip(0, 0, len(s.Leaders)-1) {
		t.Fatal("黑洞豁免測試需要先把領航員指派到艦艇")
	}
	if s.RouteBlockedByBlackHole(0, 1) {
		t.Error("有艦艇領航員時不該被黑洞擋住(手冊括號那句)")
	}
}

// TestRouteCrossesNebulaDetectsPassThrough 是這一組最重要的一支:
// **兩端都在星雲外、但直線穿過去**也算,因為手冊的字是 traveling *through* a nebula。
//
// 這正是前一版「只看起訖點」的近似會漏掉的情況。
func TestRouteCrossesNebulaDetectsPassThrough(t *testing.T) {
	w, _ := gamedata.GalaxyParsecSpan(1)
	s := routeTestSession([][2]float64{{0, 0.5}, {20 / w, 0.5}})
	// 一團擋在中間的「星雲」:正規化 x 落在 0.3..0.5 就算在裡面。
	s.SetNebulaProbe(func(x, y float64) bool { return x > 0.3 && x < 0.5 })
	if s.Stars[0].InNebula || s.Stars[1].InNebula {
		t.Fatalf("前提不成立:兩端都應在星雲外(x=%.3f, %.3f)", s.Stars[0].X, s.Stars[1].X)
	}
	if !s.RouteCrossesNebula(0, 1) {
		t.Error("航線直線穿過星雲,應判定為穿越(手冊:traveling *through* a nebula)")
	}

	// 把星雲移到航線之外 → 不該算。
	s.SetNebulaProbe(func(x, y float64) bool { return y > 0.9 })
	if s.RouteCrossesNebula(0, 1) {
		t.Error("星雲不在航線上,不該判定為穿越")
	}
}

// TestRouteCrossesNebulaWithoutProbe:沒裝探針時不憑空判定(headless 模擬即此路徑)。
func TestRouteCrossesNebulaWithoutProbe(t *testing.T) {
	s := routeTestSession([][2]float64{{0, 0}, {0.5, 0.5}})
	if s.RouteCrossesNebula(0, 1) {
		t.Error("沒有星雲遮罩時不該判定為穿越星雲")
	}
}

// TestRouteInterdictedSlowsFleet 釘住 Warp Field Interdictor:敵方星系 3 秒差距內的航線降速。
func TestRouteInterdictedSlowsFleet(t *testing.T) {
	w, h := gamedata.GalaxyParsecSpan(1)
	s := routeTestSession([][2]float64{{0, 0}, {12 / w, 0}})
	// 敵星在航線中段偏北 2 秒差距,蓋了干擾器。
	s.Stars[3] = Star{X: 6 / w, Y: 2 / h, Owner: 2, Wormhole: -1}
	s.AIPlayers = []AIOpponent{{
		ColonyStars:     []int{3},
		ColonyBuildings: []map[string]bool{{interdictorBuildingName: true}},
	}}
	if !s.RouteInterdicted(0, 1) {
		t.Error("敵星離航線 2 秒差距且有干擾器,應判定為受干擾")
	}
	// 移到 4 秒差距外(> 3)→ 不該受影響。
	s.Stars[3].Y = 4 / h
	if s.RouteInterdicted(0, 1) {
		t.Error("敵星離航線 4 秒差距,超出 3 秒差距半徑,不該受干擾")
	}
	// 沒蓋干擾器就不算。
	s.Stars[3].Y = 2 / h
	s.AIPlayers[0].ColonyBuildings[0] = map[string]bool{}
	if s.RouteInterdicted(0, 1) {
		t.Error("敵星沒蓋干擾器,不該判定為受干擾")
	}
}

// TestInterdictorNotExemptedByNavigator:手冊那句豁免只講「nebulae and black holes」,
// 干擾場是人造的,不在豁免範圍。
func TestInterdictorNotExemptedByNavigator(t *testing.T) {
	w, h := gamedata.GalaxyParsecSpan(1)
	s := routeTestSession([][2]float64{{0, 0}, {12 / w, 0}})
	s.Stars[3] = Star{X: 6 / w, Y: 2 / h, Owner: 2, Wormhole: -1}
	s.AIPlayers = []AIOpponent{{
		ColonyStars:     []int{3},
		ColonyBuildings: []map[string]bool{{interdictorBuildingName: true}},
	}}
	s.Leaders = append(s.Leaders, Leader{Name: "領航員", Skill: navigatorSkillLabel, Ship: true, Tier: 1})
	s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_NUCLEAR_FISSION: true}
	s.TechLevelSet, s.TechLevel = true, TechLevelDefault

	if got := s.fleetSpeedForTrip(0, 1); got != gamedata.InterdictorSpeed {
		t.Errorf("干擾場中即使有領航員也該降到 %d 秒差距/回合,實得 %d",
			gamedata.InterdictorSpeed, got)
	}
}

// TestInterdictorBuildingNameMatchesTable 防止譯名漂移。
//
// `ColonyBuildings` 是以中文名當 key 的 map,名字對不上就是**靜靜失效**
// (查不到 key = 沒有干擾場),不會有任何錯誤訊息。
func TestInterdictorBuildingNameMatchesTable(t *testing.T) {
	found := ""
	for _, b := range gamedata.Buildings {
		if b.NameEN == "Warp Field Interdictor" {
			found = b.NameZH
		}
	}
	if found == "" {
		t.Fatal("gamedata.Buildings 裡找不到 Warp Field Interdictor")
	}
	if found != interdictorBuildingName {
		t.Errorf("建築表的中文名是 %q,route.go 用的是 %q —— 對不上會靜靜失效",
			found, interdictorBuildingName)
	}
}

// TestSendFleetRefusesBlackHoleRoute:黑洞擋路時派遣要被拒絕,而不是默默算個 ETA。
func TestSendFleetRefusesBlackHoleRoute(t *testing.T) {
	w, h := gamedata.GalaxyParsecSpan(1)
	s := routeTestSession([][2]float64{{0, 0}, {12 / w, 0}})
	s.TechLevelSet, s.TechLevel = true, TechLevelDefault
	s.Fleet().AtStar, s.Fleet().DestStar, s.Fleet().ETA = 0, -1, 0

	if !s.SendFleet(1) {
		t.Fatal("前提不成立:沒有黑洞時這趟應該派得出去")
	}
	s.Fleet().DestStar, s.Fleet().ETA = -1, 0
	s.Stars[2] = Star{X: 6 / w, Y: 1 / h, Spectral: blackHoleSpectral, Wormhole: -1}
	if s.SendFleet(1) {
		t.Error("航線被黑洞擋住時不該接受派遣")
	}
}
