package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestEventPoolOnlyImplemented 驗證隨機事件只會抽到「remake 已實作」且非播報類的事件。
// 抽到未實作的事件會變成「跳出訊息但什麼都沒發生」的假事件,比事件少更糟。
func TestEventPoolOnlyImplemented(t *testing.T) {
	impl := map[int]bool{}
	for _, e := range gamedata.ImplementedRandomEvents() {
		impl[e.ID] = true
	}
	if len(impl) == 0 {
		t.Fatal("已實作事件池不應為空")
	}
	seen := map[int]bool{}
	for seed := int64(0); seed < 60; seed++ {
		s := NewDemoSession()
		s.EventSeed = seed
		for turn := 0; turn < 40; turn++ {
			s.EndTurn()
			if r := s.LastEventReport; r != nil {
				if !impl[r.EventID] {
					t.Fatalf("抽到未實作的事件 id=%d(%s)", r.EventID, r.Name)
				}
				seen[r.EventID] = true
			}
		}
	}
	if len(seen) < 5 {
		t.Errorf("60 局 × 40 回合只出現 %d 種事件,抽樣可能有問題", len(seen))
	}
	t.Logf("實際出現 %d 種事件(池中共 %d 種)", len(seen), len(impl))
}

// TestEventGoodFlagMatchesOriginalTable 驗證 remake 的好壞標記與原版
// _event_good_array @ 0x180E84 逐格相同(反組譯 dump 出來的 36 bytes)。
func TestEventGoodFlagMatchesOriginalTable(t *testing.T) {
	original := []int{
		1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1,
		1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1,
	}
	if len(gamedata.RandomEvents) != len(original) {
		t.Fatalf("事件數應為 %d(原版 _event_good_array 長度),got %d",
			len(original), len(gamedata.RandomEvents))
	}
	for i, e := range gamedata.RandomEvents {
		want := original[i] == 1
		if e.ID != i {
			t.Errorf("第 %d 項的 ID 應為 %d,got %d", i, i, e.ID)
		}
		if e.Good != want {
			t.Errorf("事件 %d(%s)好壞應為 %v(原版表),got %v", i, e.Name, want, e.Good)
		}
	}
}

// TestEventMessageBaseSpacing 驗證訊息索引佈局:從資產 8 起、每事件 4 條。
// 這是 EVENTMSG.LBX 的實際佈局((152-8)/4 = 36 與事件數吻合)。
func TestEventMessageBaseSpacing(t *testing.T) {
	for i, e := range gamedata.RandomEvents {
		if want := 8 + i*4; e.MsgBase != want {
			t.Errorf("事件 %d(%s)的訊息基底應為 %d,got %d", i, e.Name, want, e.MsgBase)
		}
	}
}

// TestEventsAreReproducible 驗證同一個 EventSeed 產生同一串事件(存讀檔後行為一致的前提)。
func TestEventsAreReproducible(t *testing.T) {
	run := func() []int {
		s := NewDemoSession()
		s.EventSeed = 12345
		var ids []int
		for turn := 0; turn < 60; turn++ {
			s.EndTurn()
			if r := s.LastEventReport; r != nil {
				ids = append(ids, r.EventID)
			}
		}
		return ids
	}
	a, b := run(), run()
	if len(a) != len(b) {
		t.Fatalf("同 seed 事件數不同:%d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("第 %d 個事件不可重現:%d vs %d", i, a[i], b[i])
		}
	}
	if len(a) == 0 {
		t.Error("60 回合應至少觸發一次事件")
	}
}

// TestMineralEventUpdatesBothPlanetAndColony 驗證礦產事件同時更新星圖行星資料與殖民地產能。
// 只改一邊會重演「面板說豐富、產出卻沒變」那種自打嘴巴(2026-08-06 母星氣候已踩過一次)。
func TestMineralEventUpdatesBothPlanetAndColony(t *testing.T) {
	s := NewDemoSession()
	s.eventRandForTest()
	idx, from, to, ok := s.shiftColonyMineral(+1)
	if !ok {
		t.Skip("這局沒有可提升礦產的殖民地")
	}
	star := s.PlayerColonyStarIndex(idx)
	if s.Planets[star].MineralID != to {
		t.Errorf("行星礦產應更新為 %v,got %v", to, s.Planets[star].MineralID)
	}
	if s.Planets[star].Mineral != mineralDisplayName(to) {
		t.Errorf("行星礦產顯示字串應同步,got %q", s.Planets[star].Mineral)
	}
	wantInd := gamedata.MineralIndustryPerWorker(to)
	if s.PlayerColonies[idx].IndustryPerWorker != wantInd {
		t.Errorf("殖民地每礦工工業應更新為 %d,got %d", wantInd, s.PlayerColonies[idx].IndustryPerWorker)
	}
	if from == to {
		t.Error("礦產等級應真的變動")
	}
}

// TestEventsHaveRealEffects 驗證事件不是只有文字:跑到出現事件為止,檢查狀態確實動過。
func TestEventsHaveRealEffects(t *testing.T) {
	for seed := int64(0); seed < 40; seed++ {
		s := NewDemoSession()
		s.EventSeed = seed
		for turn := 0; turn < 30; turn++ {
			before := struct {
				bc, research, ships, pop int
			}{s.Player.BC, s.Player.ResearchProgress, len(s.Ships), s.PlayerColonies[0].Population}
			s.EndTurn()
			r := s.LastEventReport
			if r == nil {
				continue
			}
			// 富商捐獻 / 海盜劫掠 / 研究類 / 艦船類都會動到上面某一項;
			// 殖民地類事件動的是殖民地欄位(這裡只抽樣檢查最容易觀察的幾個)。
			switch r.EventID {
			case 6: // 富商捐獻
				if s.Player.BC <= before.bc {
					t.Errorf("seed %d:富商捐獻後國庫應增加(%d → %d)", seed, before.bc, s.Player.BC)
				}
				return
			case 8, 13: // 艦船爆炸 / 叛變
				if len(s.Ships) >= before.ships {
					t.Errorf("seed %d:艦船事件後艦數應減少(%d → %d)", seed, before.ships, len(s.Ships))
				}
				return
			}
		}
	}
	t.Skip("40 局內沒抽到可直接觀察的事件類型")
}
