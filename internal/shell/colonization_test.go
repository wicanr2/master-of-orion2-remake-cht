package shell

import "testing"

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
	s.FleetAtStar = target
	s.FleetDestStar = -1
	s.FleetETA = 0
	return s, target
}

// TestColonizeStar_Success 驗證前置條件齊備(艦隊抵達無主星、載有殖民船)時,拓殖成功:
// PlayerColonies +1、Star.Owner 轉 1、殖民船從 s.Ships 移除、平行陣列(PlayerColonyStars 等)
// 長度與 PlayerColonies 同步。
func TestColonizeStar_Success(t *testing.T) {
	s, target := newFleetAtUnownedStarSession(t)
	beforeColonies := len(s.PlayerColonies)
	beforeShips := len(s.Ships)
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
	if len(s.Ships) != beforeShips-1 {
		t.Fatalf("殖民船應被消耗,Ships 應 -1(%d→%d),got %d", beforeShips, beforeShips-1, len(s.Ships))
	}
	if s.FleetHasColonyShip() {
		t.Fatal("拓殖後艦隊不應再有殖民船(唯一一艘已消耗)")
	}
	newColony := s.PlayerColonies[res.ColonyIndex]
	if newColony.Population != colonizeStartPopulation || newColony.Farmers != colonizeStartPopulation {
		t.Fatalf("新殖民地應全農起始(Population=Farmers=%d),got Population=%d Farmers=%d",
			colonizeStartPopulation, newColony.Population, newColony.Farmers)
	}
	if newColony.Workers != 0 || newColony.Scientists != 0 {
		t.Fatalf("新殖民地起始不應有工人/科學家,got Workers=%d Scientists=%d", newColony.Workers, newColony.Scientists)
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

// TestColonizeStar_PreconditionsChecked 驗證各前置條件缺一都會被擋下(Ok=false),
// 且不會誤動任何狀態(不消耗殖民船、不改 Star.Owner)。
func TestColonizeStar_PreconditionsChecked(t *testing.T) {
	// 條件 1:艦隊尚未抵達(仍在航行中)。
	s, target := newFleetAtUnownedStarSession(t)
	s.FleetETA = 3
	if res := s.ColonizeStar(target); res.Ok {
		t.Fatalf("艦隊未抵達不應允許拓殖,got Ok=true")
	}

	// 條件 2:目標星已有歸屬(玩家母星)。
	s2, _ := newFleetAtUnownedStarSession(t)
	s2.FleetAtStar = 0
	s2.FleetETA = 0
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
	s3.Ships = append(s3.Ships[:shipIdx], s3.Ships[shipIdx+1:]...)
	beforeShips := len(s3.Ships)
	if res := s3.ColonizeStar(target3); res.Ok {
		t.Fatalf("無殖民船不應允許拓殖,got Ok=true")
	}
	if len(s3.Ships) != beforeShips {
		t.Fatalf("拒絕拓殖不應改動 Ships,got len=%d want %d", len(s3.Ships), beforeShips)
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
