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
	startFarmers := s.PlayerColonies[0].Farmers
	startWorkers := s.PlayerColonies[0].Workers
	startScientists := s.PlayerColonies[0].Scientists
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
	end := s.PlayerColonies[0]
	// 新人口一定要被指派職務。新增者會先試工人；若那會造成食物赤字，則依
	// assignNewColonist 的保守原版近似改派農夫，不能把「工人必增」當作不變量。
	assigned := end.Farmers + end.Workers + end.Scientists
	if assigned != end.Population {
		t.Fatalf("人口與職務數必須同步:人口 %d，農／工／科 %d/%d/%d",
			end.Population, end.Farmers, end.Workers, end.Scientists)
	}
	if end.Farmers <= startFarmers && end.Workers <= startWorkers && end.Scientists <= startScientists {
		t.Fatalf("人口成長後至少一種職務必須增加:農 %d→%d、工 %d→%d、科 %d→%d",
			startFarmers, end.Farmers, startWorkers, end.Workers, startScientists, end.Scientists)
	}
	t.Logf("殖民地0 人口 %d→%d(上限 %d),農／工／科 %d/%d/%d→%d/%d/%d",
		startPop, endPop, popMax, startFarmers, startWorkers, startScientists,
		end.Farmers, end.Workers, end.Scientists)
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
