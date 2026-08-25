package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestEventPoolOnlyImplemented 驗證實際播報只會來自「remake 已實作」且非播報類的事件。
// 原版候選仍從 29 個 ID 抽；未實作項會消耗候選但不產生假播報。
func TestEventPoolOnlyImplemented(t *testing.T) {
	impl := map[int]bool{}
	for _, e := range gamedata.ImplementedRandomEvents() {
		impl[e.ID] = true
	}
	allowed := map[int]bool{}
	for _, e := range gamedata.RandomEvents {
		if e.Implemented {
			allowed[e.ID] = true
		}
	}
	if len(impl) == 0 {
		t.Fatal("已實作事件池不應為空")
	}
	seen := map[int]bool{}
	for seed := int64(0); seed < 60; seed++ {
		s := NewDemoSession()
		s.EventSeed = seed
		for turn := 0; turn < 220; turn++ {
			s.EndTurn()
			if r := s.LastEventReport; r != nil {
				if !allowed[r.EventID] {
					t.Fatalf("抽到未實作的事件 id=%d(%s)", r.EventID, r.Name)
				}
				if r.EventID <= 28 {
					seen[r.EventID] = true
				}
			}
		}
	}
	if len(seen) < 5 {
		t.Errorf("60 局 × 220 回合只出現 %d 種事件,抽樣可能有問題", len(seen))
	}
	t.Logf("實際出現 %d 種事件(池中共 %d 種)", len(seen), len(impl))
}

// TestEventGoodFlagMatchesOriginalTable 驗證隨機事件 0..28 的好壞標記與原版
// byte_180E84 @ 0x180E84 逐格相同；29..35 是狀態播報，不索引此表。
func TestEventGoodFlagMatchesOriginalTable(t *testing.T) {
	original := []int{
		1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1,
		1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}
	for i, wantRaw := range original {
		e := gamedata.RandomEvents[i]
		want := wantRaw == 1
		if e.ID != i {
			t.Errorf("第 %d 項的 ID 應為 %d,got %d", i, i, e.ID)
		}
		if e.Good != want {
			t.Errorf("事件 %d(%s)好壞應為 %v(原版表),got %v", i, e.Name, want, e.Good)
		}
	}
}

func TestLuckyEventCandidateAndCounter(t *testing.T) {
	bad := gamedata.RandomEventByID(3)
	good := gamedata.RandomEventByID(0)
	normal := NewDemoSession()
	normal.Difficulty = 2
	if !eventCandidateAllowed(normal, bad, false, 300) || !eventCandidateAllowed(normal, good, false, 300) {
		t.Fatal("一般難度必須讓好壞事件都進共同候選鏈")
	}
	tutor := NewDemoSession()
	tutor.Difficulty = 0
	if eventCandidateAllowed(tutor, bad, false, 300) {
		t.Fatal("Tutor 不得接受壞事件候選")
	}
	s := NewDemoSession()
	s.Difficulty = 2
	s.ApplyCustomRaceBonuses(Race{}, gamedata.TRAIT_LUCKY)
	if !eventCandidateAllowed(s, bad, false, 300) || eventCandidateAllowed(s, bad, true, 300) {
		t.Fatal("Lucky 不得預先移除一般壞事件；強制事件鏈則只能接受好事件")
	}
	if !eventCandidateAllowed(s, good, true, 300) {
		t.Fatal("Lucky 強制事件必須接受已實作好事件")
	}
	s.Turn = 51
	s.LuckyEventCounter = 79
	if s.advanceLuckyEventCounter(11) || s.LuckyEventCounter != 80 {
		t.Fatalf("threshold=10 的 roll 11 應失敗並保留 80：counter=%d", s.LuckyEventCounter)
	}
	if !s.advanceLuckyEventCounter(10) || s.LuckyEventCounter != 0 {
		t.Fatalf("threshold=10 的 roll 10 應成功並清零：counter=%d", s.LuckyEventCounter)
	}
	s.Turn = 50
	s.LuckyEventCounter = 79
	if s.advanceLuckyEventCounter(10) || s.LuckyEventCounter != 0 {
		t.Fatal("Turn-1 未滿 50 時成功應清零但不得觸發事件")
	}
}

func TestEventScheduleStartsAtRelativeTurn50(t *testing.T) {
	s := NewDemoSession()
	s.Difficulty = 2
	s.Turn = 50
	s.advanceEvents()
	if s.EventAttemptCounter != 0 {
		t.Fatalf("Turn=50 的 elapsed=49 不得開始排程，got attempts=%d", s.EventAttemptCounter)
	}
	s.Turn = 51
	s.advanceEvents()
	if s.EventAttemptCounter != 1 {
		t.Fatalf("Turn=51 的 elapsed=50 應完成第一次保護檢查，got attempts=%d", s.EventAttemptCounter)
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
		for turn := 0; turn < 220; turn++ {
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
		t.Error("220 回合應至少觸發一次事件")
	}
}

// TestEventsHaveRealEffects 驗證事件不是只有文字:跑到出現事件為止,檢查狀態確實動過。
func TestEventsHaveRealEffects(t *testing.T) {
	for seed := int64(0); seed < 40; seed++ {
		s := NewDemoSession()
		s.EventSeed = seed
		for turn := 0; turn < 320; turn++ {
			before := struct {
				bc, research, ships, pop int
			}{s.Player.BC, s.Player.ResearchProgress, len(s.Fleet().Ships), s.PlayerColonies[0].Population}
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
			case 8: // 艦船爆炸
				if len(s.Fleet().Ships) >= before.ships {
					t.Errorf("seed %d:艦船事件後艦數應減少(%d → %d)", seed, before.ships, len(s.Fleet().Ships))
				}
				return
			}
		}
	}
	t.Skip("40 局內沒抽到可直接觀察的事件類型")
}
