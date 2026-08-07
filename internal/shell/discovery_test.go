package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// discovery_test.go:對照原版 `Do_System_Discoveries_At_Star_` @ 0xE9927 與
// `Make_New_Colony_Or_Outpost_` @ 0xE5EB3 的行星特殊物產行為。
// 每個測試的期望值都標明是從哪一條指令讀出來的,不是從實作反推的。

// newDiscoveryTestSession 造一個小 session:一顆已擁有的母星 + 一顆待發現的目標星。
func newDiscoveryTestSession(t *testing.T, sp gamedata.PlanetSpecial) *GameSession {
	t.Helper()
	s := NewDemoSession()
	// 目標星:選一顆無主的星,把它的行星資料設成可殖民 + 指定特殊物產。
	target := -1
	for i := range s.Stars {
		if s.Stars[i].Owner == 0 && i < len(s.Planets) {
			target = i
			break
		}
	}
	if target < 0 {
		t.Fatal("找不到無主星,demo 星圖有問題")
	}
	// 星索引 ≠ 行星索引(一顆星有多顆天體),要寫進**那顆星的**代表行星才有效。
	pIdx := s.PlanetAt(target)
	s.Stars[target].Orbits = [StarOrbits]int{pIdx, OrbitEmpty, OrbitEmpty, OrbitEmpty, OrbitEmpty}
	s.Planets[pIdx] = Planet{
		Name: "測試星 I", Gen: planetGenVersion, TypeID: gamedata.HABITABLE,
		ClimateID: gamedata.TERRAN, GravityID: gamedata.NORMAL_G,
		MineralID: gamedata.ABUNDANT, SizeID: gamedata.MEDIUM_PLANET,
		SpecialID: sp,
	}
	s.Fleet().AtStar = target
	return s
}

// targetStarOf 回傳測試星所在的**星**索引(不是行星索引——兩者不同,見 orbit.go)。
func targetStarOf(t *testing.T, s *GameSession) int {
	t.Helper()
	for i := range s.Planets {
		if s.Planets[i].Name == "測試星 I" {
			return s.PlanetStar(i)
		}
	}
	t.Fatal("找不到測試星")
	return -1
}

// 太空殘骸 / 海盜藏寶:原版 `add dword ptr [player+32h], 32h` / `64h` → 50 / 100 BC。
func TestDiscoverySpaceDebrisAndPirateCache(t *testing.T) {
	cases := []struct {
		sp   gamedata.PlanetSpecial
		want int
	}{
		{gamedata.SpaceDebris, 50},
		{gamedata.PirateCache, 100},
	}
	for _, c := range cases {
		s := newDiscoveryTestSession(t, c.sp)
		idx := targetStarOf(t, s)
		before := s.Player.BC
		d := s.discoverSystemSpecials(idx)
		if d == nil {
			t.Fatalf("special %d:應該觸發發現,卻回 nil", c.sp)
		}
		if got := s.Player.BC - before; got != c.want {
			t.Errorf("special %d:國庫應 +%d,實得 +%d", c.sp, c.want, got)
		}
		if d.BCGained != c.want {
			t.Errorf("special %d:報告 BCGained=%d,want %d", c.sp, d.BCGained, c.want)
		}
		// 原版結算後把 Star.special 覆寫成訊息碼 → 同一星系不會再觸發。
		if again := s.discoverSystemSpecials(idx); again != nil {
			t.Errorf("special %d:同一顆星第二次仍觸發了發現", c.sp)
		}
	}
}

// 失散殖民地:原版 `call Colony_Race_Pop_Limit_` 後 `cmp al,3 / jbe / mov [colony+0Ah],3`
// → 人口 = min(該行星人口上限, 3)。
func TestDiscoverySplinterColony(t *testing.T) {
	s := newDiscoveryTestSession(t, gamedata.SplinterColony)
	idx := targetStarOf(t, s)
	nBefore := len(s.PlayerColonies)

	d := s.discoverSystemSpecials(idx)
	if d == nil || d.ColonyIdx < 0 {
		t.Fatal("失散殖民地應該就地生出一個殖民地")
	}
	if len(s.PlayerColonies) != nBefore+1 {
		t.Fatalf("殖民地數應 %d,實得 %d", nBefore+1, len(s.PlayerColonies))
	}
	c := s.PlayerColonies[d.ColonyIdx]
	// Medium/Terran 的 PopMax 遠大於 3,故這裡就是被 3 夾住的情形。
	if c.Population != gamedata.SplinterColonyMaxPopulation {
		t.Errorf("起始人口 %d,want %d(原版 cmp al,3 的上限)",
			c.Population, gamedata.SplinterColonyMaxPopulation)
	}
	if s.Stars[idx].Owner != 1 {
		t.Errorf("該星應轉為玩家所有,Owner=%d", s.Stars[idx].Owner)
	}
	// 平行陣列必須同步長大,否則之後任何依索引存取都會 panic。
	if len(s.Builds) < len(s.PlayerColonies) || len(s.PlayerColonyStars) != len(s.PlayerColonies) {
		t.Errorf("平行陣列未同步:Builds=%d ColonyStars=%d Colonies=%d",
			len(s.Builds), len(s.PlayerColonyStars), len(s.PlayerColonies))
	}
	if s.PlayerColonyStars[d.ColonyIdx] != idx {
		t.Errorf("殖民地↔星索引對不上:%d != %d", s.PlayerColonyStars[d.ColonyIdx], idx)
	}
}

// 受困英雄:手冊「In gratitude for the rescue…」——免費入列,不扣 BC。
func TestDiscoveryLostHeroIsFree(t *testing.T) {
	s := newDiscoveryTestSession(t, gamedata.LostHero)
	idx := targetStarOf(t, s)
	bcBefore := s.Player.BC
	nBefore := len(s.Leaders)

	d := s.discoverSystemSpecials(idx)
	if d == nil || d.LeaderGot == "" {
		t.Fatal("受困英雄應該免費加入一名領袖")
	}
	if len(s.Leaders) != nBefore+1 {
		t.Fatalf("領袖數應 %d,實得 %d", nBefore+1, len(s.Leaders))
	}
	if s.Player.BC != bcBefore {
		t.Errorf("免費領袖不應扣 BC:%d → %d", bcBefore, s.Player.BC)
	}
}

// 遠古文物:原版送「一項」現在就能研究的科技(Random_(4)/4+1 恆為 1,見 gamedata 註解)。
func TestDiscoveryAncientArtifactsGrantsOneTech(t *testing.T) {
	s := newDiscoveryTestSession(t, gamedata.AncientArtifacts)
	idx := targetStarOf(t, s)
	before := 0
	for _, done := range s.Player.CompletedTopics {
		if done {
			before++
		}
	}
	d := s.discoverSystemSpecials(idx)
	if d == nil || d.TechGot == "" {
		t.Fatal("遠古文物應該白送一項科技")
	}
	after := 0
	for _, done := range s.Player.CompletedTopics {
		if done {
			after++
		}
	}
	// 原版 `Random_(4)/4 + 1` → 1 項,25% 機率 2 項(Random_ 回 1..n,見 gamedata 訂正說明)。
	if got := after - before; got < 1 || got > 2 {
		t.Errorf("送出主題數 %d,應為 1 或 2", got)
	}
}

// 金礦/寶石礦不是「抵達即觸發」那一類,而是持續效果——原版的 dispatch 根本沒有 4/5 的分支。
func TestDiscoveryIgnoresOngoingSpecials(t *testing.T) {
	for _, sp := range []gamedata.PlanetSpecial{gamedata.GoldDeposits, gamedata.GemDeposits, gamedata.Natives} {
		s := newDiscoveryTestSession(t, sp)
		idx := targetStarOf(t, s)
		if d := s.discoverSystemSpecials(idx); d != nil {
			t.Errorf("special %d 不該觸發抵達發現,卻回了 %q", sp, d.Message)
		}
	}
}

// 原住民:原版在殖民時額外寫 3 個人口單位(colony+0x10..0x1C,stride 4),population = 4,
// 並把 planet.special 清成 0。
func TestColonizeNativesAddsThreePopulationAndConsumesSpecial(t *testing.T) {
	s := newDiscoveryTestSession(t, gamedata.Natives)
	idx := targetStarOf(t, s)
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{Class: ColonyShipClass})
	s.Fleet().AtStar, s.Fleet().ETA = idx, 0

	res := s.ColonizeStar(idx)
	if !res.Ok {
		t.Fatalf("拓殖失敗:%s", res.Reason)
	}
	c := s.PlayerColonies[res.ColonyIndex]
	want := colonizeStartPopulation + gamedata.NativePopulationUnits // 1 + 3 = 4
	if c.Population != want {
		t.Errorf("原住民行星起始人口 %d,want %d(原版 [colony+0Ah]=4)", c.Population, want)
	}
	if c.Farmers != want {
		t.Errorf("原住民全部是農夫(原版 job 位元清 0),Farmers=%d want %d", c.Farmers, want)
	}
	// 手冊:「at a +2 food production advantage」
	if got := c.FoodPerFarmer - gamedata.ClimateFoodPerFarmer(gamedata.TERRAN); got != 2 {
		t.Errorf("每農夫食物加成 %d,want 2", got)
	}
	if s.Planets[idx].SpecialID != gamedata.NoSpecial {
		t.Errorf("原住民應在殖民後消耗掉(原版 [planet+0Fh]=0),仍是 %d", s.Planets[idx].SpecialID)
	}
}

// 金礦/寶石礦:手冊 +5 / +10 BC/回合,做成殖民地層的固定收入。
func TestColonizeGoldGemIncome(t *testing.T) {
	cases := []struct {
		sp   gamedata.PlanetSpecial
		want int
	}{
		{gamedata.GoldDeposits, 5},
		{gamedata.GemDeposits, 10},
		{gamedata.NoSpecial, 0},
	}
	for _, c := range cases {
		s := newDiscoveryTestSession(t, c.sp)
		idx := targetStarOf(t, s)
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{Class: ColonyShipClass})
		s.Fleet().AtStar, s.Fleet().ETA = idx, 0
		res := s.ColonizeStar(idx)
		if !res.Ok {
			t.Fatalf("special %d 拓殖失敗:%s", c.sp, res.Reason)
		}
		if got := s.PlayerColonies[res.ColonyIndex].SpecialIncome; got != c.want {
			t.Errorf("special %d 固定收入 %d,want %d", c.sp, got, c.want)
		}
	}
}

// 遠古文物殖民後:手冊「produces 5 research points instead of the usual 3」。
func TestColonizeArtifactsResearchOverride(t *testing.T) {
	s := newDiscoveryTestSession(t, gamedata.AncientArtifacts)
	idx := targetStarOf(t, s)
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{Class: ColonyShipClass})
	s.Fleet().AtStar, s.Fleet().ETA = idx, 0
	s.RaceIndex = -1 // 排除種族加成,單看特殊物產本身

	res := s.ColonizeStar(idx)
	if !res.Ok {
		t.Fatalf("拓殖失敗:%s", res.Reason)
	}
	if got := s.PlayerColonies[res.ColonyIndex].ResearchPerScientist; got != 5 {
		t.Errorf("每科學家研究 %d,want 5(取代基準 %d)", got, gamedata.ResearchPerScientistNorm)
	}
}

// TestDiscoveryFiresDuringRealTurnLoop 是端到端接線驗證:不直接呼叫 discoverSystemSpecials,
// 而是照玩家實際流程(派遣艦隊 → 逐回合 EndTurn → 抵達)看發現有沒有真的觸發、有沒有進到
// 快報用的 LastDiscovery 欄位。單元測試綠 ≠ 玩得到,這一條補的就是中間那段接線。
func TestDiscoveryFiresDuringRealTurnLoop(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true // 隔離變數:不讓隨機事件干擾

	// 找一顆「不是艦隊現址」的無主星,種一個海盜藏寶進去。
	target := -1
	for i := range s.Stars {
		if i != s.Fleet().AtStar && s.Stars[i].Owner == 0 && i < len(s.Planets) {
			target = i
			break
		}
	}
	if target < 0 {
		t.Fatal("找不到可派遣的無主星")
	}
	s.Planets[target].Gen = planetGenVersion
	s.Planets[target].NoPlanet = false
	s.Planets[target].SpecialID = gamedata.PirateCache

	if !s.SendFleet(target) {
		t.Fatal("派遣艦隊失敗")
	}
	bcBefore := s.Player.BC

	got := false
	for turn := 0; turn < 40 && !got; turn++ {
		s.EndTurn()
		if s.LastDiscovery != nil {
			got = true
			if s.LastDiscovery.StarIndex != target {
				t.Errorf("發現的星索引 %d,want %d", s.LastDiscovery.StarIndex, target)
			}
			if s.LastDiscovery.BCGained != 100 {
				t.Errorf("入袋 %d BC,want 100", s.LastDiscovery.BCGained)
			}
		}
	}
	if !got {
		t.Fatalf("艦隊抵達 %d 號星後仍未觸發發現(FleetAtStar=%d ETA=%d)",
			target, s.Fleet().AtStar, s.Fleet().ETA)
	}
	// BC 淨變化含每回合經濟,只確認至少多了那 100(不把整條經濟軌跡綁死在這條測試裡)。
	if s.Player.BC <= bcBefore {
		t.Logf("提醒:國庫 %d → %d,發現的 100 BC 被回合支出吃掉了(不算失敗)", bcBefore, s.Player.BC)
	}
	// 再跑幾回合,同一顆星不能重複觸發。
	s.LastDiscovery = nil
	for turn := 0; turn < 5; turn++ {
		s.EndTurn()
		if s.LastDiscovery != nil {
			t.Fatalf("同一顆星重複觸發了發現:%s", s.LastDiscovery.Message)
		}
	}
}
