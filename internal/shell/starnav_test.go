package shell

import "testing"

// starnav_test.go:星圖鍵盤導覽的護欄。

func navSession(nStars int) *GameSession {
	s := &GameSession{}
	s.Stars = make([]Star, nStars)
	for i := range s.Stars {
		s.Stars[i].Name = "S"
	}
	s.FleetAtStar = -1
	return s
}

// TestCycleStarListWraps:環狀循環,兩個方向都要接得回去。
//
// 原版那支的邊界檢查是「索引 < 艦隊數」與「索引 > −1」,撞到就從另一端重來。
func TestCycleStarListWraps(t *testing.T) {
	list := []int{2, 5, 9}
	for _, c := range []struct {
		cur     int
		forward bool
		want    int
	}{
		{2, true, 5}, {5, true, 9}, {9, true, 2}, // 往後,尾接頭
		{2, false, 9}, {5, false, 2}, {9, false, 5}, // 往前,頭接尾
	} {
		if got := cycleStarList(list, c.cur, c.forward); got != c.want {
			t.Errorf("從 %d 往 forward=%v 應到 %d,實得 %d", c.cur, c.forward, c.want, got)
		}
	}
}

// TestCycleStarListFromOutside:目前選的星不在清單裡時,往後從頭、往前從尾。
//
// 這是實際會遇到的情形——玩家點了一顆沒有殖民地的星,再按 F5。
func TestCycleStarListFromOutside(t *testing.T) {
	list := []int{2, 5, 9}
	if got := cycleStarList(list, 7, true); got != 2 {
		t.Errorf("清單外往後應到第一個 2,實得 %d", got)
	}
	if got := cycleStarList(list, 7, false); got != 9 {
		t.Errorf("清單外往前應到最後一個 9,實得 %d", got)
	}
	if got := cycleStarList(list, -1, true); got != 2 {
		t.Errorf("未選取(−1)往後應到第一個 2,實得 %d", got)
	}
}

// TestCycleStarListEmpty:清單空回 −1,呼叫端據此不動選取。
func TestCycleStarListEmpty(t *testing.T) {
	if got := cycleStarList(nil, 3, true); got != -1 {
		t.Errorf("空清單應回 −1,實得 %d", got)
	}
}

// TestColonizedStarsFiltersPaddingAndDedups:
// `PlayerColonyStars` 可能含 −1 padding(見該欄位註解),也可能同星多殖民地。
func TestColonizedStarsFiltersPaddingAndDedups(t *testing.T) {
	s := navSession(10)
	s.PlayerColonyStars = []int{7, -1, 3, 7, 99, 3}
	got := s.ColonizedStars()
	want := []int{3, 7}
	if len(got) != len(want) {
		t.Fatalf("應剩 %v,實得 %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("應為 %v(依星索引排序),實得 %v", want, got)
		}
	}
}

// TestCycleColonizedStarWalksAllColonies:F5 連按應該走遍所有殖民地再繞回來。
func TestCycleColonizedStarWalksAllColonies(t *testing.T) {
	s := navSession(10)
	s.PlayerColonyStars = []int{4, 1, 8}
	cur := -1
	seen := map[int]bool{}
	for i := 0; i < 3; i++ {
		cur = s.CycleColonizedStar(cur, true)
		if cur < 0 {
			t.Fatalf("第 %d 次 F5 回了 −1", i)
		}
		seen[cur] = true
	}
	if len(seen) != 3 {
		t.Errorf("連按 3 次應走過 3 個殖民地,實際只到過 %v", seen)
	}
	if got := s.CycleColonizedStar(cur, true); got != 1 {
		t.Errorf("第 4 次應繞回最小索引 1,實得 %d", got)
	}
}

// TestCycleWithNoTargetsReturnsMinusOne:沒有殖民地/沒有艦隊時不能亂選一顆星。
func TestCycleWithNoTargetsReturnsMinusOne(t *testing.T) {
	s := navSession(10)
	if got := s.CycleColonizedStar(2, true); got != -1 {
		t.Errorf("沒有殖民地時 F5 應回 −1,實得 %d", got)
	}
	if got := s.CycleFleetStar(2, true); got != -1 {
		t.Errorf("艦隊位置未知時 F1 應回 −1,實得 %d", got)
	}
}

// TestKnownFleetStarsIsSingleForNow 釘住一個**已知的模型限制**,不是期望行為。
//
// remake 的玩家艦隊是單一集合(`FleetAtStar`),AI 對手只有抽象的 `FleetStrength`、
// 在星圖上沒有位置。所以 F1/F2 的循環集合現在只有一個元素。
//
// 多艦隊做出來時這條測試會紅——**那時候該改的是測試**,而且同一個模型缺口也卡著
// 星圖的遷移連線層(見 gap report)。
func TestKnownFleetStarsIsSingleForNow(t *testing.T) {
	s := navSession(10)
	s.FleetAtStar = 4
	got := s.KnownFleetStars()
	if len(got) != 1 || got[0] != 4 {
		t.Fatalf("目前應只有玩家自己那一支艦隊 [4],實得 %v", got)
	}
	if idx := s.CycleFleetStar(-1, true); idx != 4 {
		t.Errorf("F1 應把選取帶到艦隊所在星 4,實得 %d", idx)
	}
	if idx := s.CycleFleetStar(4, true); idx != 4 {
		t.Errorf("只有一支艦隊時再按 F1 應停在原地,實得 %d", idx)
	}
}

// TestFleetOutOfRangeIsNotAFleet:FleetAtStar 越界(未初始化的存檔等)不能算成一支艦隊。
func TestFleetOutOfRangeIsNotAFleet(t *testing.T) {
	s := navSession(3)
	s.FleetAtStar = 99
	if got := s.KnownFleetStars(); got != nil {
		t.Errorf("越界的 FleetAtStar 不該算成艦隊,實得 %v", got)
	}
}
