package shell

import "testing"

// TestFleetInterstellarMovement 驗證艦隊星間航行:派遣至遠星後,需經數回合 ETA 遞減才抵達,
// 抵達後艦隊位置更新且該星標記為已探索。
func TestFleetInterstellarMovement(t *testing.T) {
	s := NewDemoSession()
	if s.Fleet().AtStar != 0 {
		t.Fatalf("艦隊初始應在母星 0,實得 %d", s.Fleet().AtStar)
	}

	// 找一顆與母星距離足夠(ETA>1)的目的星。
	dest := -1
	for i := 1; i < len(s.Stars); i++ {
		if s.SendFleet(i) {
			if s.Fleet().ETA > 1 {
				dest = i
				break
			}
			// ETA==1 的星太近,直接跳過:先手動重置再試下一顆。
			s.Fleet().DestStar, s.Fleet().ETA = -1, 0
		}
	}
	if dest < 0 {
		t.Fatal("找不到 ETA>1 的目的星")
	}
	if s.Stars[dest].Explored {
		t.Fatalf("目的星 %d 出發前不應已探索", dest)
	}

	eta := s.Fleet().ETA
	// 航行中不可再下新令。
	if s.SendFleet((dest % (len(s.Stars) - 1)) + 1) {
		t.Fatal("艦隊航行中不應接受新派遣令")
	}

	for i := 0; i < eta; i++ {
		if s.Fleet().AtStar == dest {
			t.Fatalf("第 %d 回合就抵達,早於 ETA %d", i, eta)
		}
		s.EndTurn()
	}
	if s.Fleet().AtStar != dest {
		t.Fatalf("經 ETA %d 回合後應抵達 %d,實得 %d(ETA 剩 %d)", eta, dest, s.Fleet().AtStar, s.Fleet().ETA)
	}
	if !s.Stars[dest].Explored {
		t.Fatalf("抵達後目的星 %d 應標記為已探索", dest)
	}
	t.Logf("艦隊 0 →星%d,航行 %d 回合抵達並探索", dest, eta)
}
