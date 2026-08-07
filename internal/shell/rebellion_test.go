package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 自己拓殖的殖民地永遠不會叛亂——手冊 p.184「no colony (except captured ones) ever revolts」。
//
// 原版用獨立的政策旗標(`[colony+0x12F] == 4` 直接返回)當閘門;remake 靠
// 「UnassimilatedPop == 0」等價擋掉。這一條確認那個等價成立。
func TestSelfColonizedColoniesNeverRebel(t *testing.T) {
	s := NewDemoSession()
	for i := range s.PlayerColonies {
		if s.PlayerColonies[i].UnassimilatedPop != 0 {
			t.Fatalf("開局殖民地 %d 不該有未同化人口", i)
		}
	}
	before := len(s.PlayerColonies)
	for turn := 0; turn < 50; turn++ {
		s.Turn = turn
		if got := s.advanceRebellions(); len(got) != 0 {
			t.Fatalf("第 %d 回合不該有叛亂:%+v", turn, got)
		}
	}
	if len(s.PlayerColonies) != before {
		t.Errorf("殖民地數量不該變:%d → %d", before, len(s.PlayerColonies))
	}
}

// 機率 100% 時一定爆發——正對照,確認整條檢定路徑真的接上了。
//
// 100 單位未同化人口 = 1000‰,而擲骰上界就是 1000,`roll <= chance` 恆成立。
func TestRebellionActuallyFiresAtCertainty(t *testing.T) {
	s := newRebellionSession(t, 100 /* 未同化人口 */, 0 /* 守軍 */)
	got := s.advanceRebellions()
	if len(got) != 1 || !got[0].Triggered {
		t.Fatalf("機率 100%% 應必定爆發,得到 %+v", got)
	}
	if got[0].Rebels < 1 || got[0].Rebels > 100 {
		t.Errorf("叛軍數應落在 1..100,得到 %d", got[0].Rebels)
	}
}

// 守軍打贏 → 起事的人口沒了,殖民地還在自己手上。
func TestSuppressedRebellionCostsPopulationButKeepsTheColony(t *testing.T) {
	s := newRebellionSession(t, 100, 400 /* 壓倒性守軍 */)
	popBefore := s.PlayerColonies[0].Population
	colonies := len(s.PlayerColonies)

	got := s.advanceRebellions()
	if len(got) != 1 || !got[0].Triggered {
		t.Fatalf("應該爆發,得到 %+v", got)
	}
	if !got[0].DefenderWon {
		t.Fatalf("400 陸戰隊對上最多 100 個叛軍應該守得住,得到 %+v", got[0])
	}
	if got[0].ColonyLost {
		t.Error("守住了就不該丟殖民地")
	}
	if len(s.PlayerColonies) != colonies {
		t.Errorf("殖民地數量不該變:%d → %d", colonies, len(s.PlayerColonies))
	}
	if s.PlayerColonies[0].Population >= popBefore {
		t.Errorf("鎮壓要死人:人口 %d → %d", popBefore, s.PlayerColonies[0].Population)
	}
	if s.PlayerColonies[0].Population < 1 {
		t.Errorf("鎮壓不該把殖民地滅掉,人口 %d", s.PlayerColonies[0].Population)
	}
}

// 守軍打輸 → 殖民地**還給舊主**(手冊 p.165「the colony reverts back」)。
//
// 這一條同時是平行陣列的回歸鎖:removePlayerColony 漏掉任何一個陣列,
// 下面的長度不變量就會抓到。
func TestLostRebellionRevertsTheColonyToItsOldRuler(t *testing.T) {
	s := newRebellionSession(t, 100, 0 /* 完全沒有守軍 */)
	aiColoniesBefore := len(s.AIPlayers[0].Colonies)
	starIdx := s.PlayerColonyStarIndex(0)

	got := s.advanceRebellions()
	if len(got) != 1 || !got[0].Triggered {
		t.Fatalf("應該爆發,得到 %+v", got)
	}
	if got[0].DefenderWon {
		t.Fatalf("沒有守軍不該守得住,得到 %+v", got[0])
	}
	if !got[0].ColonyLost || got[0].RevertedToAI != 0 {
		t.Errorf("應還給 AI 0,得到 ColonyLost=%v RevertedToAI=%d", got[0].ColonyLost, got[0].RevertedToAI)
	}
	if len(s.AIPlayers[0].Colonies) != aiColoniesBefore+1 {
		t.Errorf("AI 應多一個殖民地:%d → %d", aiColoniesBefore, len(s.AIPlayers[0].Colonies))
	}
	// AI 的四個平行陣列要一起長。
	ai := s.AIPlayers[0]
	if len(ai.ColonyStars) != len(ai.Colonies) || len(ai.ColonyPlanets) != len(ai.Colonies) ||
		len(ai.ColonyBuildings) != len(ai.Colonies) {
		t.Errorf("AI 平行陣列長度不一致:Colonies=%d Stars=%d Planets=%d Buildings=%d",
			len(ai.Colonies), len(ai.ColonyStars), len(ai.ColonyPlanets), len(ai.ColonyBuildings))
	}
	if ai.ColonyStars[len(ai.ColonyStars)-1] != starIdx {
		t.Errorf("還回去的殖民地應在星 %d,得到 %d", starIdx, ai.ColonyStars[len(ai.ColonyStars)-1])
	}
	// 還回去之後那筆不再是「征服所得」——新主人不用同化自己的子民。
	back := ai.Colonies[len(ai.Colonies)-1]
	if back.UnassimilatedPop != 0 || back.ConqueredFromKnown {
		t.Errorf("還回去的殖民地不該留著征服標記:%+v", back)
	}
	assertPlayerColonyArraysConsistent(t, s)
}

// 異族管理中心讓機率減半——建築真的接上了,不是只在 gamedata 層有函式。
func TestAlienManagementCenterHalvesTheChanceInPlay(t *testing.T) {
	mk := func(withCenter bool) int {
		s := newRebellionSession(t, 10, 400)
		if withCenter {
			s.ColonyBuildings[0] = map[string]bool{alienManagementCenterName: true}
		}
		// 直接讀檢定算出來的機率(不管有沒有觸發)。
		r, _ := s.checkRebellionAt(0)
		return r.ChancePermille
	}
	plain, center := mk(false), mk(true)
	if plain <= 0 {
		t.Fatalf("沒有建築時應有機率,得到 %d‰", plain)
	}
	if center != plain/2 {
		t.Errorf("異族管理中心應減半:%d‰ → 預期 %d‰,實得 %d‰", plain, plain/2, center)
	}
}

// 同化完之後就不再擲骰——同化不只是計時器,它是叛亂風險的計時器。
func TestFullyAssimilatedColonyStopsRolling(t *testing.T) {
	s := newRebellionSession(t, 100, 400)
	if _, ok := s.checkRebellionAt(0); !ok {
		t.Fatal("還有未同化人口時應該有事情發生")
	}
	s.PlayerColonies[0].UnassimilatedPop = 0
	if r, ok := s.checkRebellionAt(0); ok {
		t.Errorf("同化完就不該再擲骰,得到 %+v", r)
	}
}

// --- 測試輔助 ---

// newRebellionSession 建一個「第 0 個殖民地是從 AI 0 手上打下來的」對局。
func newRebellionSession(t *testing.T, unassimilated, marines int) *GameSession {
	t.Helper()
	s := NewDemoSession()
	if len(s.PlayerColonies) == 0 || len(s.AIPlayers) == 0 {
		t.Fatal("demo 對局應該有玩家殖民地與 AI")
	}
	c := &s.PlayerColonies[0]
	c.Population = unassimilated
	c.PopMax = unassimilated + 10
	markColonyConquered(c, 0)
	for len(s.PlayerColonyMarines) < len(s.PlayerColonies) {
		s.PlayerColonyMarines = append(s.PlayerColonyMarines, 0)
	}
	for len(s.PlayerColonyTanks) < len(s.PlayerColonies) {
		s.PlayerColonyTanks = append(s.PlayerColonyTanks, 0)
	}
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, nil)
	}
	s.PlayerColonyMarines[0] = marines
	s.PlayerColonyTanks[0] = 0
	// 其餘殖民地不要一起叛亂,把測試搞混。
	for i := 1; i < len(s.PlayerColonies); i++ {
		s.PlayerColonies[i].UnassimilatedPop = 0
	}
	return s
}

// assertPlayerColonyArraysConsistent 檢查所有玩家殖民地平行陣列的長度不變量。
//
// 這些陣列是**延遲配置**的(可以比 PlayerColonies 短),所以檢查的是「不可以更長」
// ——更長代表移除時漏了某一個,而那種錯會靜默地讓索引錯位。
func assertPlayerColonyArraysConsistent(t *testing.T, s *GameSession) {
	t.Helper()
	n := len(s.PlayerColonies)
	for _, a := range []struct {
		name string
		got  int
	}{
		{"Builds", len(s.Builds)},
		{"ColonyBuildings", len(s.ColonyBuildings)},
		{"PlayerColonyMarines", len(s.PlayerColonyMarines)},
		{"MarineBarracksAge", len(s.MarineBarracksAge)},
		{"PlayerColonyTanks", len(s.PlayerColonyTanks)},
		{"ArmorBarracksAge", len(s.ArmorBarracksAge)},
		{"PlayerColonyStars", len(s.PlayerColonyStars)},
		{"PlayerColonyPlanets", len(s.PlayerColonyPlanets)},
	} {
		if a.got > n {
			t.Errorf("%s 長度 %d 超過 PlayerColonies 的 %d——移除時漏了這一個陣列", a.name, a.got, n)
		}
	}
}

// 叛軍那一側用的是第四種部隊,不是陸戰隊——弄錯的話叛軍會強得離譜。
func TestRebelsFightAsTheWeakestUnitType(t *testing.T) {
	if gamedata.GroundTypeStrengthDelta(gamedata.GroundTypeRebels) >=
		gamedata.GroundTypeStrengthDelta(gamedata.GroundTypeMilitia) {
		t.Error("叛軍應該比民兵還弱")
	}
	// 同樣人數下,叛軍打不贏「同人數的陸戰隊守軍」——這是相對強度的煙霧測試。
	lost := 0
	for turn := 1; turn <= 20; turn++ {
		s := newRebellionSession(t, 30, 30)
		s.Turn = turn
		if !s.resolveRebellionCombat(0, 30, s.playerMarineForce()) {
			lost++
		}
	}
	if lost > 10 {
		t.Errorf("同人數下守方(陸戰隊+民兵)輸了 %d/20 場——叛軍太強,檢查用的是哪個部隊類型", lost)
	}
}

// 沒有舊主(舊存檔或舊主已滅)時,殖民地脫離帝國但不會 panic。
func TestRebellionWithoutAKnownOldRulerDoesNotCrash(t *testing.T) {
	s := newRebellionSession(t, 100, 0)
	s.PlayerColonies[0].ConqueredFromKnown = false
	before := len(s.PlayerColonies)

	got := s.advanceRebellions()
	if len(got) != 1 || !got[0].ColonyLost {
		t.Fatalf("應該丟掉殖民地,得到 %+v", got)
	}
	if got[0].RevertedToAI != -1 {
		t.Errorf("沒有舊主時應回報 −1,得到 %d", got[0].RevertedToAI)
	}
	if len(s.PlayerColonies) != before-1 {
		t.Errorf("殖民地應少一個:%d → %d", before, len(s.PlayerColonies))
	}
	assertPlayerColonyArraysConsistent(t, s)
}

// 未同化人口不該被 engine 的經濟結算意外讀成別的東西——加了兩個新欄位之後的煙霧測試。
func TestConqueredMarkerDoesNotLeakIntoEconomy(t *testing.T) {
	base := engine.ColonyState{
		Population: 8, PopMax: 12, Farmers: 4, Workers: 4,
		FoodPerFarmer: 2, IndustryPerWorker: 3, ResearchPerScientist: 2,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT,
	}
	conquered := base
	markColonyConquered(&conquered, 1)
	if got, want := engine.RunColonyTurn(conquered), engine.RunColonyTurn(base); got != want {
		t.Errorf("征服標記不該改變經濟結算\n有標記: %+v\n無標記: %+v", got, want)
	}
}
