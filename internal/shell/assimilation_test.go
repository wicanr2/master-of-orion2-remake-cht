package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 攻下的殖民地整批是未同化人口,之後每 N 回合同化一單位。
func TestAssimilationCountsDownAfterConquest(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Government = gamedata.MoraleGovDemocracy // 民主:4 回合一單位
	c := engine.ColonyState{Population: 3, PopMax: 8, PlanetGravity: gamedata.NORMAL_G}
	markColonyConquered(&c, -1)
	if c.UnassimilatedPop != 3 {
		t.Fatalf("剛攻下應整批是外族人口(3),得到 %d", c.UnassimilatedPop)
	}
	s.PlayerColonies = append(s.PlayerColonies, c)
	idx := len(s.PlayerColonies) - 1
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, nil)
	}
	if got := s.AssimilationTurnsFor(idx); got != 4 {
		t.Fatalf("民主政體應是 4 回合,得到 %d", got)
	}
	// 3 回合還不夠。
	for i := 0; i < 3; i++ {
		s.advanceAssimilation()
	}
	if s.PlayerColonies[idx].UnassimilatedPop != 3 {
		t.Errorf("3 回合還不該同化任何人,得到未同化 %d", s.PlayerColonies[idx].UnassimilatedPop)
	}
	// 第 4 回合同化一單位。
	s.advanceAssimilation()
	if s.PlayerColonies[idx].UnassimilatedPop != 2 {
		t.Errorf("第 4 回合應同化一單位(剩 2),得到 %d", s.PlayerColonies[idx].UnassimilatedPop)
	}
	// 再 8 回合把剩下兩單位吃掉。
	for i := 0; i < 8; i++ {
		s.advanceAssimilation()
	}
	if got := s.PlayerColonies[idx].UnassimilatedPop; got != 0 {
		t.Errorf("12 回合後應全部同化,得到 %d", got)
	}
}

// 餘數留著繼續累,不是「每 N 回合歸零重來」——政體改變時不該吃掉已累積的進度。
func TestAssimilationProgressCarriesOverAcrossRateChanges(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Government = gamedata.MoraleGovUnification // 統一:20 回合
	c := engine.ColonyState{Population: 2, PopMax: 8, PlanetGravity: gamedata.NORMAL_G}
	markColonyConquered(&c, -1)
	s.PlayerColonies = append(s.PlayerColonies, c)
	idx := len(s.PlayerColonies) - 1
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, nil)
	}
	for i := 0; i < 3; i++ {
		s.advanceAssimilation()
	}
	if s.PlayerColonies[idx].AssimilationProgress != 3 {
		t.Fatalf("應累積 3 回合進度,得到 %d", s.PlayerColonies[idx].AssimilationProgress)
	}
	// 蓋起異族管理中心 → 門檻變 2,累積的 3 回合立刻兌現一單位、餘 1。
	s.ColonyBuildings[idx] = map[string]bool{alienManagementCenterName: true}
	s.advanceAssimilation() // 進度 4,門檻 2 → 同化兩單位
	if got := s.PlayerColonies[idx].UnassimilatedPop; got != 0 {
		t.Errorf("累積的進度應被沿用而不是歸零:未同化剩 %d", got)
	}
}

// 異族管理中心對統一政體是十倍速。
func TestAlienManagementCenterOverridesGovernmentInSession(t *testing.T) {
	s := NewDemoSession()
	s.Government = gamedata.MoraleGovUnification
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, nil)
	}
	slow := s.AssimilationTurnsFor(0)
	s.ColonyBuildings[0] = map[string]bool{alienManagementCenterName: true}
	fast := s.AssimilationTurnsFor(0)
	if slow != 20 || fast != 2 {
		t.Errorf("統一政體應 20 → 2 回合,得到 %d → %d", slow, fast)
	}
}

// 沒有外族人口的殖民地不該累積進度殘值。
func TestAssimilationLeavesCleanColoniesAlone(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	before := s.PlayerColonies[0].Population
	for i := 0; i < 30; i++ {
		s.advanceAssimilation()
	}
	if s.PlayerColonies[0].UnassimilatedPop != 0 || s.PlayerColonies[0].AssimilationProgress != 0 {
		t.Errorf("自己拓殖的殖民地不該有同化狀態:未同化 %d、進度 %d",
			s.PlayerColonies[0].UnassimilatedPop, s.PlayerColonies[0].AssimilationProgress)
	}
	if s.PlayerColonies[0].Population != before {
		t.Error("同化不該改變人口總數——它只是換旗子,不是增減人口")
	}
}

// 排斥種族減半(整條路徑,不只 gamedata)。
func TestRepulsiveRaceSlowsAssimilationInSession(t *testing.T) {
	s := NewDemoSession()
	s.Government = gamedata.MoraleGovDemocracy
	base := s.AssimilationTurnsFor(0)
	s.ApplyRace(raceIndexByEnName(t, "Silicoids")) // 矽基是惹人厭種族
	s.Government = gamedata.MoraleGovDemocracy     // ApplyRace 不動政體,重設保持與 base 同條件
	if !s.RaceRepulsive() {
		t.Fatal("矽基應是惹人厭種族")
	}
	if got := s.AssimilationTurnsFor(0); got != base*2 {
		t.Errorf("排斥種族應是兩倍回合:%d → %d", base, got)
	}
}

// 未同化人口 = 多種族殖民地 → 20% 士氣懲罰(手冊 p.66-67),異族管理中心消除。
//
// 這條把第 96 項寫的那句「機制在、後果還沒接」關掉:同化現在真的有代價了。
func TestUnassimilatedPopulationCostsMorale(t *testing.T) {
	clean := colonyMoralePercent(gamedata.MoraleGovDictatorship, nil, false, 0)
	multi := colonyMoralePercent(gamedata.MoraleGovDictatorship, nil, true, 0)
	if multi != clean-20 {
		t.Errorf("多種族殖民地應 −20 士氣:%d → %d", clean, multi)
	}
	// 異族管理中心消除它。
	withCenter := colonyMoralePercent(gamedata.MoraleGovDictatorship,
		map[string]bool{alienManagementCenterName: true}, true, 0)
	if withCenter != clean {
		t.Errorf("有異族管理中心時懲罰應消失:一般 %d、多種族+建築 %d", clean, withCenter)
	}
}

// 同化完最後一單位的那一刻,懲罰要跟著消失——不重算的話玩家會一直被扣。
func TestMoralePenaltyLiftsWhenAssimilationFinishes(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Government = gamedata.MoraleGovDemocracy // 4 回合一單位
	c := engine.ColonyState{Population: 1, PopMax: 8, PlanetGravity: gamedata.NORMAL_G}
	markColonyConquered(&c, -1)
	s.PlayerColonies = append(s.PlayerColonies, c)
	idx := len(s.PlayerColonies) - 1
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, nil)
	}
	s.recalcColonyMorale(idx)
	penalised := s.PlayerColonies[idx].MoralePercent

	for i := 0; i < 4; i++ {
		s.advanceAssimilation()
	}
	if s.PlayerColonies[idx].UnassimilatedPop != 0 {
		t.Fatalf("4 回合後應同化完,未同化剩 %d", s.PlayerColonies[idx].UnassimilatedPop)
	}
	if got := s.PlayerColonies[idx].MoralePercent; got != penalised+20 {
		t.Errorf("同化完之後 20 點懲罰應消失:%d → %d(期望 %d)", penalised, got, penalised+20)
	}
}
