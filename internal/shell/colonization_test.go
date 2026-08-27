package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// newFleetAtUnownedStarSession 建一個新對局,把玩家艦隊直接擺到某顆無主星上空(已抵達,ETA=0),
// 供拓殖相關測試省去先跑 SendFleet/EndTurn 航行流程。回傳對局與目標星索引。比照
// ground_invasion_test.go 的 newFleetAtAIHomeSession 慣例。
func newFleetAtUnownedStarSession(t *testing.T) (*GameSession, int) {
	t.Helper()
	s := NewDemoSession()
	s.DisableEvents = true
	target := -1
	for i, st := range s.Stars {
		if st.Owner == 0 {
			target = i
			break
		}
	}
	if target < 0 {
		t.Fatal("找不到可用的無主星做測試")
	}
	s.Fleet().AtStar = target
	s.Fleet().DestStar = -1
	s.Fleet().ETA = 0
	return s, target
}

// TestColonizeStar_Success 驗證前置條件齊備(艦隊抵達無主星、載有殖民船)時,拓殖成功:
// PlayerColonies +1、Star.Owner 轉 1、殖民船從 s.Fleet().Ships 移除、平行陣列(PlayerColonyStars 等)
// 長度與 PlayerColonies 同步。
func TestColonizeStar_Success(t *testing.T) {
	s, target := newFleetAtUnownedStarSession(t)
	beforeColonies := len(s.PlayerColonies)
	beforeShips := len(s.Fleet().Ships)
	if !s.FleetHasColonyShip() {
		t.Fatal("測試前提錯誤:開局艦隊(homeworldShips)應含一艘殖民船")
	}

	res := s.ColonizeStar(target)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,ColonizeStar 應成功,got Reason=%q", res.Reason)
	}
	if res.StartPopulation != colonizeStartPopulation {
		t.Fatalf("起始人口應為 %d,got %d", colonizeStartPopulation, res.StartPopulation)
	}
	if res.PopMax < colonizeStartPopulation {
		t.Fatalf("PopMax(%d) 不應低於起始人口(%d)", res.PopMax, colonizeStartPopulation)
	}
	if len(s.PlayerColonies) != beforeColonies+1 {
		t.Fatalf("PlayerColonies 應 +1(%d→%d),got %d", beforeColonies, beforeColonies+1, len(s.PlayerColonies))
	}
	if s.Stars[target].Owner != 1 {
		t.Fatalf("拓殖後 Star.Owner 應轉 1,got %d", s.Stars[target].Owner)
	}
	if len(s.Fleet().Ships) != beforeShips-1 {
		t.Fatalf("殖民船應被消耗,Ships 應 -1(%d→%d),got %d", beforeShips, beforeShips-1, len(s.Fleet().Ships))
	}
	if s.FleetHasColonyShip() {
		t.Fatal("拓殖後艦隊不應再有殖民船(唯一一艘已消耗)")
	}
	newColony := s.PlayerColonies[res.ColonyIndex]
	if newColony.Population != colonizeStartPopulation || newColony.Farmers+newColony.Workers != colonizeStartPopulation {
		t.Fatalf("新殖民地人口與職務應守恆,got Population=%d Farmers=%d Workers=%d",
			newColony.Population, newColony.Farmers, newColony.Workers)
	}
	if newColony.Scientists != 0 {
		t.Fatalf("新殖民地起始不應有科學家,got Scientists=%d", newColony.Scientists)
	}

	// 平行陣列長度不變量:見 GameSession.PlayerColonyStars 欄位註解。
	if len(s.PlayerColonyStars) != len(s.PlayerColonies) {
		t.Fatalf("PlayerColonyStars 長度應與 PlayerColonies 同步,got %d vs %d", len(s.PlayerColonyStars), len(s.PlayerColonies))
	}
	if s.PlayerColonyStars[res.ColonyIndex] != target {
		t.Fatalf("PlayerColonyStars[%d] 應記錄目標星索引 %d,got %d", res.ColonyIndex, target, s.PlayerColonyStars[res.ColonyIndex])
	}
	if len(s.Builds) != len(s.PlayerColonies) || len(s.ColonyBuildings) != len(s.PlayerColonies) ||
		len(s.PlayerColonyMarines) != len(s.PlayerColonies) || len(s.MarineBarracksAge) != len(s.PlayerColonies) ||
		len(s.PlayerColonyTanks) != len(s.PlayerColonies) || len(s.ArmorBarracksAge) != len(s.PlayerColonies) {
		t.Fatalf("所有平行殖民地陣列長度都應與 PlayerColonies(%d)同步", len(s.PlayerColonies))
	}
}

func TestNewColonyStartingJobMatchesOriginalBranch(t *testing.T) {
	s := NewDemoSession()
	planetIdx := s.PlanetAt(0)
	if planetIdx < 0 {
		t.Fatal("測試需要一顆行星")
	}
	p := &s.Planets[planetIdx]
	p.Gen = 2
	p.TypeID = gamedata.HABITABLE
	p.SizeID = gamedata.MEDIUM_PLANET
	p.GravityID = gamedata.NORMAL_G
	p.MineralID = gamedata.ABUNDANT

	p.ClimateID = gamedata.TERRAN
	p.SpecialID = gamedata.NoSpecial
	farmColony, ok, refusal := s.newColonyFromPlanet(planetIdx, s.Government, 0, 0, 0)
	if !ok {
		t.Fatalf("可耕行星應可建立殖民地：%+v", refusal)
	}
	if farmColony.Farmers != 1 || farmColony.Workers != 0 || farmColony.PopulationGroups[0].Farmers != 1 {
		t.Fatalf("可耕行星應從農夫開始：%+v", farmColony)
	}

	p.ClimateID = gamedata.TOXIC
	p.SpecialID = gamedata.Natives
	workerColony, ok, refusal := s.newColonyFromPlanet(planetIdx, s.Government, 0, 0, 0)
	if !ok {
		t.Fatalf("有毒行星仍是合法殖民目標：%+v", refusal)
	}
	if workerColony.Population != 4 || workerColony.Farmers != 3 || workerColony.Workers != 1 {
		t.Fatalf("owner 應為工人、三位 Native 應為農夫：pop=%d farmers=%d workers=%d",
			workerColony.Population, workerColony.Farmers, workerColony.Workers)
	}
	if len(workerColony.PopulationGroups) != 2 || workerColony.PopulationGroups[0].Workers != 1 ||
		workerColony.PopulationGroups[1].Farmers != 3 {
		t.Fatalf("逐族職務分流錯誤：%+v", workerColony.PopulationGroups)
	}
}

// TestColonizeStar_PreconditionsChecked 驗證各前置條件缺一都會被擋下(Ok=false),
// 且不會誤動任何狀態(不消耗殖民船、不改 Star.Owner)。
func TestColonizeStar_PreconditionsChecked(t *testing.T) {
	// 條件 1:艦隊尚未抵達(仍在航行中)。
	s, target := newFleetAtUnownedStarSession(t)
	s.Fleet().ETA = 3
	if res := s.ColonizeStar(target); res.Ok {
		t.Fatalf("艦隊未抵達不應允許拓殖,got Ok=true")
	}

	// 條件 2:目標星已有歸屬(玩家母星)。
	s2, _ := newFleetAtUnownedStarSession(t)
	s2.Fleet().AtStar = 0
	s2.Fleet().ETA = 0
	if res := s2.ColonizeStar(0); res.Ok {
		t.Fatalf("已有歸屬的星不應允許拓殖,got Ok=true")
	}
	if len(s2.PlayerColonies) != 1 {
		t.Fatalf("拒絕拓殖不應改動 PlayerColonies,got len=%d", len(s2.PlayerColonies))
	}

	// 條件 3:艦隊未載運殖民船(先手動移除)。
	s3, target3 := newFleetAtUnownedStarSession(t)
	shipIdx := s3.findColonyShipIndex()
	if shipIdx < 0 {
		t.Fatal("測試前提錯誤:開局艦隊應有殖民船")
	}
	s3.Fleet().Ships = append(s3.Fleet().Ships[:shipIdx], s3.Fleet().Ships[shipIdx+1:]...)
	beforeShips := len(s3.Fleet().Ships)
	if res := s3.ColonizeStar(target3); res.Ok {
		t.Fatalf("無殖民船不應允許拓殖,got Ok=true")
	}
	if len(s3.Fleet().Ships) != beforeShips {
		t.Fatalf("拒絕拓殖不應改動 Ships,got len=%d want %d", len(s3.Fleet().Ships), beforeShips)
	}
	if s3.Stars[target3].Owner != 0 {
		t.Fatalf("拒絕拓殖不應改動 Star.Owner,got %d", s3.Stars[target3].Owner)
	}

	// 條件 4:無效星索引。
	s4, _ := newFleetAtUnownedStarSession(t)
	if res := s4.ColonizeStar(len(s4.Stars) + 100); res.Ok {
		t.Fatalf("無效星索引不應允許拓殖,got Ok=true")
	}
}

// TestColonizeStar_EconomyRunsAfterEndTurn 驗證拓殖後的新殖民地能正常參與 EndTurn 經濟結算,
// 不會 panic,且會產生對應的 LastPlayerOutput.Colonies 條目(見 shell.GameSession.EndTurn →
// engine.RunEmpireTurn(s.Player, s.PlayerColonies) 逐殖民地結算)。
func TestColonizeStar_EconomyRunsAfterEndTurn(t *testing.T) {
	s, target := newFleetAtUnownedStarSession(t)
	res := s.ColonizeStar(target)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}

	s.EndTurn() // 不應 panic

	if len(s.LastPlayerOutput.Colonies) != len(s.PlayerColonies) {
		t.Fatalf("EndTurn 後 LastPlayerOutput.Colonies 應涵蓋所有殖民地(含新殖民地),got %d want %d",
			len(s.LastPlayerOutput.Colonies), len(s.PlayerColonies))
	}
	if s.PlayerColonies[res.ColonyIndex].Population < 0 {
		t.Fatalf("新殖民地人口不應變負數,got %d", s.PlayerColonies[res.ColonyIndex].Population)
	}
	// 再多跑幾回合,確認不會 panic、popAccum/建築等平行陣列的索引不會越界。
	for i := 0; i < 5; i++ {
		s.EndTurn()
	}
	if len(s.PlayerColonies) < 2 {
		t.Fatalf("多跑幾回合後,新殖民地不應憑空消失,got len=%d", len(s.PlayerColonies))
	}
}

// TestClimateFromDisplay_CoversAllGenPlanetsClimates 驗證 genPlanets 用到的 7 種氣候顯示字串
// 都能對映到 gamedata.PlanetClimate(不會有「不應發生」的 !ok 分支被實際觸發)。
func TestClimateFromDisplay_CoversAllGenPlanetsClimates(t *testing.T) {
	for _, disp := range []string{"放射", "貧瘠", "海洋", "沙漠", "凍原", "有毒", "地獄"} {
		if _, ok := climateFromDisplay(disp); !ok {
			t.Errorf("climateDisplayToGamedata 缺少 genPlanets 會用到的氣候顯示字串 %q", disp)
		}
	}
}

// TestGravityMineralSizeFromDisplay_CoverGenPlanetsValues 同上,驗證重力/礦產/大小三個對映表
// 涵蓋 genPlanets 實際會產生的顯示字串。
func TestGravityMineralSizeFromDisplay_CoverGenPlanetsValues(t *testing.T) {
	for _, disp := range []string{"低", "常態", "高"} {
		if _, ok := gravityFromDisplay(disp); !ok {
			t.Errorf("gravityDisplayToGamedata 缺少 %q", disp)
		}
	}
	for _, disp := range []string{"貧瘠", "一般", "豐富", "富饒"} {
		if _, ok := mineralFromDisplay(disp); !ok {
			t.Errorf("mineralDisplayToGamedata 缺少 %q", disp)
		}
	}
	for _, disp := range []string{"巨大", "大型", "中型", "小型"} {
		if _, ok := sizeFromDisplay(disp); !ok {
			t.Errorf("sizeDisplayToGamedata 缺少 %q", disp)
		}
	}
}
