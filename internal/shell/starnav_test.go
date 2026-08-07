package shell

import "testing"

// starnav_test.go:星圖鍵盤導覽的護欄。

func navSession(nStars int) *GameSession {
	s := &GameSession{}
	s.Stars = make([]Star, nStars)
	for i := range s.Stars {
		s.Stars[i].Name = "S"
	}
	s.Fleet().AtStar = -1
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

// TestKnownFleetStarsCoversEveryFleet:F1/F2 要走遍**每一支**艦隊。
//
// 這條測試先前叫 `...IsSingleForNow`,釘的是「循環集合只有一個元素」這個模型限制,
// 並寫明「多艦隊做出來時該改的是測試」。多艦隊做出來了(見 fleet.go),所以改了。
func TestKnownFleetStarsCoversEveryFleet(t *testing.T) {
	s := navSession(10)
	s.Fleets = []Fleet{{AtStar: 4, DestStar: -1}, {AtStar: 1, DestStar: -1}, {AtStar: 7, DestStar: -1}}
	got := s.KnownFleetStars()
	want := []int{1, 4, 7}
	if len(got) != len(want) {
		t.Fatalf("應涵蓋三支艦隊 %v,實得 %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("應為 %v(依星索引排序),實得 %v", want, got)
		}
	}
	// 連按 F1 走遍三支再繞回來。
	cur := -1
	for _, exp := range []int{1, 4, 7, 1} {
		cur = s.CycleFleetStar(cur, true)
		if cur != exp {
			t.Fatalf("F1 應到 %d,實得 %d", exp, cur)
		}
	}
}

// TestKnownFleetStarsDedupsSameStar:兩支艦隊停同一顆星只算一個落點。
//
// 循環的是**視角**;不去重的話按第二次 F1 看起來像沒反應。
func TestKnownFleetStarsDedupsSameStar(t *testing.T) {
	s := navSession(10)
	s.Fleets = []Fleet{{AtStar: 3, DestStar: -1}, {AtStar: 3, DestStar: -1}}
	if got := s.KnownFleetStars(); len(got) != 1 || got[0] != 3 {
		t.Errorf("同一顆星上的兩支艦隊應只算一個落點 [3],實得 %v", got)
	}
}

// TestFleetOutOfRangeIsNotAFleet:AtStar 越界(未初始化的存檔等)不能算成一個落點。
func TestFleetOutOfRangeIsNotAFleet(t *testing.T) {
	s := navSession(3)
	s.Fleet().AtStar = 99
	if got := s.KnownFleetStars(); got != nil {
		t.Errorf("越界的 AtStar 不該算成艦隊落點,實得 %v", got)
	}
}
