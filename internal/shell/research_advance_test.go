package shell

import "testing"

// TestResearchUnlockLoopOverTurns 驗證「玩數回合 → 研究逐步完成 → 艦艇元件逐步解鎖」的
// 動態迴圈:一開始只有起始元件解鎖,持續結束回合後,已解鎖的元件數應增加。
func TestResearchUnlockLoopOverTurns(t *testing.T) {
	s := NewDemoSession()

	countUnlocked := func() int {
		n := 0
		for _, opts := range [][]Component{WeaponOptions, ArmorOptions, ShieldOptions, SpecialOptions} {
			for _, c := range opts {
				if s.ComponentUnlocked(c) {
					n++
				}
			}
		}
		return n
	}

	start := countUnlocked()

	// 跑足夠多回合讓最便宜的主題(cost 150)累積完成並推進。
	for i := 0; i < 40; i++ {
		s.EndTurn()
	}

	end := countUnlocked()
	if end <= start {
		t.Fatalf("研究解鎖迴圈未生效:起始解鎖 %d,40 回合後 %d(應增加)", start, end)
	}
	if len(s.Player.CompletedTopics) == 0 {
		t.Fatalf("40 回合後應至少完成一個研究主題,CompletedTopics 為空")
	}
	t.Logf("解鎖元件:%d → %d;完成主題數 %d", start, end, len(s.Player.CompletedTopics))
}
