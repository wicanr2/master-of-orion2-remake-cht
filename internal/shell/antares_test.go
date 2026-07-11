package shell

import "testing"

// bcCrashFloor80Turns 是 TestAntaresRaidsScheduleAndEscalate 80 回合內允許的 BC 下限。
//
// 忠實 yield 經濟(母星 Terran/Abundant,見 docs/tech/colony-economy-maintenance.md)下,
// 建築維護費固定 3 BC/回合,但人口只剩 1 時,不論把僅存的 1 人配置成農夫或工人,收入都不到
// 3 BC(食物盈餘出售 0.5 BC/單位、稅收 40%,單人口撐死賺 1~2 BC)——這不是本輪任何算式錯誤,
// 而是「建築維護費不隨人口規模縮小」這個手冊本身就有的機制,在忠實(零緩衝)經濟下被誠實呈現
// 出來:此測試刻意無艦隊防禦、吃滿安塔蘭入侵傷害,人口會被反覆打到剩 1(母星人口下限本身仍
// 受下方斷言保護),因此本測試不再要求「BC 絕不為負」(那個假設建立在舊 placeholder 經濟
// NetBC 穩定 +3/回合累積出的巨額緩衝上,忠實經濟沒有這個緩衝)。改驗證「BC 不會失控式無下限
// 崩潰」——以本測試固定 EventSeed=42 的確定性軌跡實測,80 回合最低點在回合 43 觸底後回升
// (2026-07-12 校正母星分配 農4/工1/科3 後為 -24,先前 農4/工3/科1 較高工業時為 -3;較忠實的
// 低工業母星緩衝更薄故觸底更深,但仍有界且會恢復,非螺旋崩潰),這裡抓一個有餘裕但仍能抓到
// 「異常擴大化」的下限。2026-07-12 再校正開局 BC 100→50(SAVE10 oracle)後,同軌跡最低點
// 由 -24 降到 -31(回合43 觸底後仍回升至 -3、人口守 1;因 BC 低時買不起的支出會跳過而自限,
// 非線性下移),故下限放寬到 -40 留餘裕,仍能抓真正的無下限螺旋。
const bcCrashFloor80Turns = -40

// TestAntaresRaidsScheduleAndEscalate 驗證安塔蘭入侵:前期寬限不觸發,達排程回合週期性觸發,
// 次數遞增(升級),母星人口不低於 1,且 BC 不會失控式無下限崩潰(見 bcCrashFloor80Turns 註解:
// 忠實經濟下人口被打到剩 1 時,單人口收入結構性不足以覆蓋建築維護費,短暫轉負是誠實的經濟後果,
// 不是 bug)。
func TestAntaresRaidsScheduleAndEscalate(t *testing.T) {
	s := NewDemoSession()
	s.Ships = nil // 無艦隊防禦,吃滿傷害(方便觀察)

	raidTurns := []int{}
	for i := 0; i < 80; i++ {
		s.EndTurn()
		if s.LastAntares != "" {
			raidTurns = append(raidTurns, s.Turn)
		}
		if s.Player.BC < bcCrashFloor80Turns {
			t.Fatalf("BC 崩潰超出合理下限:%d(< %d)", s.Player.BC, bcCrashFloor80Turns)
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
