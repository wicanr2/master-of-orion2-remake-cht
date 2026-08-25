package shell

import "testing"

func TestOriginalPirateActivityStrength(t *testing.T) {
	wants := []int{20, 40, 60, 100, 80}
	for difficulty, want := range wants {
		if got := originalPirateActivityStrength(100, difficulty); got != want {
			t.Errorf("difficulty %d strength=%d, want %d", difficulty, got, want)
		}
	}
	if got := originalPirateActivityStrength(24, 0); got != 0 {
		t.Errorf("低於原版門檻應取消，got %d", got)
	}
}

func TestPirateActivityDestroysFreightersForEveryPresentEmpire(t *testing.T) {
	s := NewDemoSession()
	star := s.PlayerColonyStarIndex(0)
	s.Player.ActiveFreighters = 7
	s.Fleets = []Fleet{NewFleet(star)} // 沒有船，避免同回合直接清零。
	if len(s.AIPlayers) == 0 {
		t.Fatal("測試需要 AI")
	}
	s.AIPlayers[0].ColonyStars = []int{star}
	s.AIPlayers[0].Player.ActiveFreighters = 4
	s.eventRand = newRandStream(1)
	e := PersistentEvent{Kind: PersistentPirateActivity, StarIndex: star, Strength: 10, InitialStrength: 10}
	done, _, _ := s.stepPirateActivity(&e)
	if done {
		t.Fatal("沒有清剿艦艇時事件不應結束")
	}
	if s.Player.ActiveFreighters != 6 || s.AIPlayers[0].Player.ActiveFreighters != 3 {
		t.Fatalf("同星帝國應各損失一艘，player=%d ai=%d", s.Player.ActiveFreighters, s.AIPlayers[0].Player.ActiveFreighters)
	}
}

func TestPirateActivityCountsHotseatFreightersAndAllOwnersDefense(t *testing.T) {
	s := NewDemoSession()
	if s.SetupHotseat(2) != 2 {
		t.Fatal("測試需要兩席熱座")
	}
	star := s.PlayerColonyStarIndex(0)
	s.Player.ActiveFreighters = 25
	s.Seats[1].PlayerColonyStars = []int{star}
	s.Seats[1].Player.ActiveFreighters = 25
	s.Seats[1].Fleets = []Fleet{{AtStar: star, DestStar: -1, Ships: []Ship{{Class: "泰坦"}}}}
	if got := s.pirateActivityFreighterTotal(star); got != 50 {
		t.Fatalf("熱座運輸船總數=%d, want 50", got)
	}
	s.Fleets = []Fleet{NewFleet(star)}
	s.eventRand = newRandStream(2)
	e := PersistentEvent{Kind: PersistentPirateActivity, StarIndex: star, Strength: 5, InitialStrength: 5}
	done, _, _ := s.stepPirateActivity(&e)
	if !done {
		t.Fatal("非當前席位停泊的 Titan 也應參與共同清剿")
	}
}

func TestPirateActivityConflictsWithCometAndPersists(t *testing.T) {
	s := NewDemoSession()
	planet, star := s.ColonyPlanetIndex(0), s.PlayerColonyStarIndex(0)
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentComet, PlanetIndex: planet, StarIndex: star}}
	if !s.pirateActivityConflictAtStar(star) {
		t.Fatal("同星彗星應阻止海盜活動")
	}
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentPirateActivity, StarIndex: star, Turns: 3, Strength: 9, InitialStrength: 17}}
	restored := s.snapshot().restore()
	if len(restored.PersistentEvents) != 1 {
		t.Fatalf("讀檔後事件數=%d", len(restored.PersistentEvents))
	}
	e := restored.PersistentEvents[0]
	if e.Kind != PersistentPirateActivity || e.StarIndex != star || e.Strength != 9 || e.InitialStrength != 17 {
		t.Fatalf("海盜活動存檔往返失真：%+v", e)
	}
}
