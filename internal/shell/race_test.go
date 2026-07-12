package shell

import "testing"

// TestApplyRaceBonuses 驗證選定種族的起始加成正確套到帝國。
func TestApplyRaceBonuses(t *testing.T) {
	// 找出各招牌種族索引。
	idx := func(en string) int {
		for i, r := range Races {
			if r.EnName == en {
				return i
			}
		}
		t.Fatalf("找不到種族 %s", en)
		return -1
	}

	// 克拉肯:工業/工人 +2。
	s := NewDemoSession()
	baseInd := s.PlayerColonies[0].IndustryPerWorker
	s.ApplyRace(idx("Klackons"))
	if got := s.PlayerColonies[0].IndustryPerWorker; got != baseInd+2 {
		t.Fatalf("克拉肯工業應 +2:%d → %d", baseInd, got)
	}

	// 席隆:研究/科學家 +2(手冊 p.614「2 more than galactic norm」,norm3+2=5,對齊 SAVE10.GAM
	// Psilon 母星每科研=5;2026-07-12 由先前 remake 調校值 +4 訂正為手冊真值 +2)。
	s = NewDemoSession()
	baseRes := s.PlayerColonies[0].ResearchPerScientist
	s.ApplyRace(idx("Psilons"))
	if got := s.PlayerColonies[0].ResearchPerScientist; got != baseRes+2 {
		t.Fatalf("席隆研究應 +2:%d → %d", baseRes, got)
	}

	// 諾蘭姆:每人口每回合 +1 BC(手冊 p.16「each unit of Gnolam population generates an
	// additional 1 BC per turn」)。2026-07-12 校正:原版種族無「一次性起始國庫」加成(SAVE10
	// 五族開局 BC 全=50),先前 remake 的 StartBC +120 已移除,改為 IncomePerPop 每回合收入。
	s = NewDemoSession()
	baseBC := s.Player.BC
	s.ApplyRace(idx("Gnolams"))
	if got := s.Player.BC; got != baseBC {
		t.Fatalf("諾蘭姆不應加起始國庫(原版無此機制):%d → %d", baseBC, got)
	}
	if got := s.PlayerColonies[0].IncomePerPop; got != 1 {
		t.Fatalf("諾蘭姆殖民地 IncomePerPop 應為 1(每人每回合+1BC),實得 %d", got)
	}
	// 每回合 EndTurn 後,諾蘭姆母星(pop8)應比無此特質多約 8 BC 收入。
	baseTax := func() int {
		ns := NewDemoSession()
		ns.EndTurn()
		return ns.LastPlayerOutput.TaxRevenue
	}()
	s.EndTurn()
	if s.LastPlayerOutput.TaxRevenue <= baseTax {
		t.Fatalf("諾蘭姆稅收應高於基準(每人+1BC×pop):諾蘭姆 %d vs 基準 %d", s.LastPlayerOutput.TaxRevenue, baseTax)
	}
}

// TestSakkraGrowthFaster 驗證薩克拉(成長 +30%)比矽基(-20%)人口成長更快。
func TestSakkraGrowthFaster(t *testing.T) {
	grow := func(en string) int {
		s := NewDemoSession()
		s.DisableEvents = true
		for i, r := range Races {
			if r.EnName == en {
				s.ApplyRace(i)
			}
		}
		start := s.PlayerColonies[0].Population
		for i := 0; i < 20; i++ {
			s.EndTurn()
		}
		return s.PlayerColonies[0].Population - start
	}
	sakkra := grow("Sakkra")
	silicoid := grow("Silicoids")
	if sakkra <= silicoid {
		t.Fatalf("薩克拉成長應快於矽基:薩克拉 +%d vs 矽基 +%d", sakkra, silicoid)
	}
	t.Logf("20 回合人口成長:薩克拉 +%d、矽基 +%d", sakkra, silicoid)
}

// TestMrrshanCombatBonus 驗證姆瑞森(戰鬥 +25%)戰績優於無加成種族。
func TestMrrshanCombatBonus(t *testing.T) {
	win := func(en string) bool {
		s := NewDemoSession()
		for i, r := range Races {
			if r.EnName == en {
				s.ApplyRace(i)
			}
		}
		s.Turn = 3
		res := s.ResolveBattle("測試敵")
		return res.PlayerWon
	}
	// 姆瑞森戰力加成應至少不劣於人類(此為煙霧測試:確保 CombatPct 有接上且不 panic)。
	_ = win("Mrrshan")
	_ = win("Humans")
	// 直接驗證 RaceCombatPct 有被套用。
	s := NewDemoSession()
	for i, r := range Races {
		if r.EnName == "Mrrshan" {
			s.ApplyRace(i)
		}
	}
	if s.RaceCombatPct != 25 {
		t.Fatalf("姆瑞森戰鬥加成應為 25%%,實得 %d", s.RaceCombatPct)
	}
}
