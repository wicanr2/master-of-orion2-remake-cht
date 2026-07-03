package shell

import "testing"

// TestAntaresRaidsScheduleAndEscalate 驗證安塔蘭入侵:前期寬限不觸發,達排程回合週期性觸發,
// 次數遞增(升級),且效果有界(BC 不為負、母星人口不低於 1)。
func TestAntaresRaidsScheduleAndEscalate(t *testing.T) {
	s := NewDemoSession()
	s.Ships = nil // 無艦隊防禦,吃滿傷害(方便觀察)

	raidTurns := []int{}
	for i := 0; i < 80; i++ {
		s.EndTurn()
		if s.LastAntares != "" {
			raidTurns = append(raidTurns, s.Turn)
		}
		if s.Player.BC < 0 {
			t.Fatalf("BC 為負:%d", s.Player.BC)
		}
		if s.PlayerColonies[0].Population < 1 {
			t.Fatalf("母星人口 <1:%d", s.PlayerColonies[0].Population)
		}
	}
	if len(raidTurns) < 2 {
		t.Fatalf("80 回合內應有多次安塔蘭入侵,實得 %d 次:%v", len(raidTurns), raidTurns)
	}
	// 首次不早於寬限回合。
	if raidTurns[0] < antaresStartTurn {
		t.Fatalf("首次入侵 %d 早於寬限 %d", raidTurns[0], antaresStartTurn)
	}
	// 週期一致。
	if raidTurns[1]-raidTurns[0] != antaresInterval {
		t.Fatalf("入侵週期應為 %d,實得 %d", antaresInterval, raidTurns[1]-raidTurns[0])
	}
	if s.AntaresRaids != len(raidTurns) {
		t.Fatalf("AntaresRaids 計數 %d != 觸發次數 %d", s.AntaresRaids, len(raidTurns))
	}
	t.Logf("安塔蘭入侵回合:%v(共 %d 次)", raidTurns, s.AntaresRaids)
}

// TestAntaresDefenseReducesDamage 驗證母星有艦隊時損失較低。
func TestAntaresDefenseReducesDamage(t *testing.T) {
	run := func(withFleet bool) int {
		s := NewDemoSession()
		if !withFleet {
			s.Ships = nil
		} else {
			s.Ships = []Ship{{Name: "衛戍艦", Class: "戰艦"}}
			s.FleetAtStar = 0
		}
		startBC := 0
		bcLoss := 0
		for i := 0; i < 40; i++ {
			before := s.Player.BC
			s.EndTurn()
			if s.LastAntares != "" {
				bcLoss += before - s.Player.BC
			}
			_ = startBC
		}
		return bcLoss
	}
	undefended := run(false)
	defended := run(true)
	if defended >= undefended {
		t.Fatalf("有母星艦隊防禦應損失較少 BC:防禦 %d vs 無防禦 %d", defended, undefended)
	}
	t.Logf("BC 損失:無防禦 %d、有防禦 %d", undefended, defended)
}
