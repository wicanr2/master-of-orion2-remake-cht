package shell

import "testing"

// TestPopulationGrowthWriteback 驗證殖民地人口會隨回合成長並回寫 Population,且不超過 PopMax。
func TestPopulationGrowthWriteback(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true // 隔離瘟疫/隕石(會扣人口,干擾精確斷言)
	if len(s.PlayerColonies) == 0 {
		t.Fatal("需至少一個殖民地")
	}
	startPop := s.PlayerColonies[0].Population
	startWorkers := s.PlayerColonies[0].Workers
	popMax := s.PlayerColonies[0].PopMax

	// 跑足夠回合讓成長累加跨過門檻。
	for i := 0; i < 30; i++ {
		s.EndTurn()
	}

	endPop := s.PlayerColonies[0].Population
	if endPop <= startPop {
		t.Fatalf("30 回合後人口應成長:起始 %d → %d", startPop, endPop)
	}
	if endPop > popMax {
		t.Fatalf("人口 %d 超過上限 %d", endPop, popMax)
	}
	// 新人口應加到工人。
	if s.PlayerColonies[0].Workers <= startWorkers {
		t.Fatalf("新人口應分配為工人:起始 %d → %d", startWorkers, s.PlayerColonies[0].Workers)
	}
	t.Logf("殖民地0 人口 %d→%d(上限 %d),工人 %d→%d", startPop, endPop, popMax, startWorkers, s.PlayerColonies[0].Workers)
}

// TestPopulationCappedAtMax 驗證人口成長受 PopMax 硬上限。
func TestPopulationCappedAtMax(t *testing.T) {
	s := NewDemoSession()
	// 把第一殖民地逼近上限,跑很多回合,確認不越界。
	s.PlayerColonies[0].Population = s.PlayerColonies[0].PopMax - 1
	for i := 0; i < 200; i++ {
		s.EndTurn()
	}
	if s.PlayerColonies[0].Population > s.PlayerColonies[0].PopMax {
		t.Fatalf("人口 %d 越過上限 %d", s.PlayerColonies[0].Population, s.PlayerColonies[0].PopMax)
	}
}
