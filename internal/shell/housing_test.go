package shell

import "testing"

// TestHousingBuildAcceleratesGrowth 驗證「住宅」建造選項真的接上了成長加成,
// 不是只在選單裡多一項。
//
// 背景:gamedata.ColonyHousingBonus 與 engine.ColonyState.Housing 早就存在,
// 但 2026-08-06 之前沒有任何地方設過那個旗標(建造選單裡根本沒有「住宅」),
// 整條住房成長路徑是死碼。
func TestHousingBuildAcceleratesGrowth(t *testing.T) {
	grow := func(build string) int {
		s := NewDemoSession()
		s.DisableEvents = true
		s.Builds[0] = ColonyBuild{Name: build}
		start := s.PlayerColonies[0].Population
		for i := 0; i < 12; i++ {
			s.EndTurn()
		}
		return s.PlayerColonies[0].Population - start
	}
	withHousing := grow(HousingBuildName)
	withTrade := grow(TradeGoodsBuildName)
	if withHousing <= withTrade {
		t.Errorf("選住宅的人口成長應快於貿易品:住宅 +%d vs 貿易品 +%d", withHousing, withTrade)
	}
}

// TestHousingFlagSyncsToColony 驗證旗標確實同步到 engine 層(而不是只有名字對)。
func TestHousingFlagSyncsToColony(t *testing.T) {
	s := NewDemoSession()
	s.Builds[0] = ColonyBuild{Name: HousingBuildName}
	s.syncTradeGoodsFlag()
	if !s.PlayerColonies[0].Housing {
		t.Error("選住宅後 ColonyState.Housing 應為 true")
	}
	if s.PlayerColonies[0].TradeGoods {
		t.Error("選住宅時不應同時標成貿易品")
	}
	s.Builds[0] = ColonyBuild{Name: TradeGoodsBuildName}
	s.syncTradeGoodsFlag()
	if s.PlayerColonies[0].Housing {
		t.Error("改選貿易品後 Housing 應關掉")
	}
}

// TestHousingIsAlwaysAvailable 驗證住宅與貿易品一樣不受科技 gate,永遠在可建清單裡。
// 這正是開局選單「只剩貿易品跟不建造」的解法。
func TestHousingIsAlwaysAvailable(t *testing.T) {
	s := NewDemoSession()
	found := false
	for _, o := range s.AvailableBuildOptions() {
		if o.Name == HousingBuildName {
			found = true
		}
	}
	if !found {
		t.Error("住宅應恆在可建清單中(無前置科技)")
	}
}
