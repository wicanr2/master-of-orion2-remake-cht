package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

// relocation_test.go:集結點與遷移連線的護欄。

func relocSession() *GameSession {
	s := &GameSession{Stars: make([]Star, 6)}
	for i := range s.Stars {
		s.Stars[i].Name = "S"
	}
	s.PlayerColonies = []engine.ColonyState{{}, {}}
	s.PlayerColonyStars = []int{0, 4}
	s.ShowRelocationLines = true
	s.Fleets = []Fleet{NewFleet(0)}
	return s
}

// TestRelocationDefaultsToNoneNotStarZero 釘住這個系統**唯一**的零值陷阱。
//
// Go 的 int 零值是 0,而 0 是**母星的索引**。如果平行陣列補齊時填零值,
// 每個新殖民地一建好就會把新造的艦全部往母星送——而那看起來完全像是遊戲規則。
func TestRelocationDefaultsToNoneNotStarZero(t *testing.T) {
	s := relocSession()
	for i := range s.PlayerColonies {
		if got := s.ColonyRelocation(i); got != ColonyRelocationNone {
			t.Errorf("殖民地 %d 預設應為「沒設定」(%d),實得 %d", i, ColonyRelocationNone, got)
		}
	}
	// 走一次會補齊平行陣列的路徑,再確認一次。
	s.SetColonyRelocation(0, 3)
	if got := s.ColonyRelocation(1); got != ColonyRelocationNone {
		t.Errorf("補齊平行陣列時新格應填 −1,實得 %d", got)
	}
}

// TestSetRelocationToOwnStarClears:送往自己 = 取消。
func TestSetRelocationToOwnStarClears(t *testing.T) {
	s := relocSession()
	s.SetColonyRelocation(0, 3)
	if s.ColonyRelocation(0) != 3 {
		t.Fatal("前置條件:應先設成 3")
	}
	s.SetColonyRelocation(0, 0) // 殖民地 0 就在星 0
	if got := s.ColonyRelocation(0); got != ColonyRelocationNone {
		t.Errorf("設成自己所在的星應等於取消,實得 %d", got)
	}
}

// TestRelocationRejectsOutOfRange:越界的星索引不能寫進去。
func TestRelocationRejectsOutOfRange(t *testing.T) {
	s := relocSession()
	if s.SetColonyRelocation(0, 99) {
		t.Error("越界的目標星不該被接受")
	}
	if s.SetColonyRelocation(9, 1) {
		t.Error("越界的殖民地索引不該被接受")
	}
}

// TestRelocationLinksRespectDisplayToggle:顯示開關關掉是**整層不畫**。
//
// 原版那支函式開頭就檢查 `byte_199BE4`,關掉直接 return——不是畫得淡一點。
func TestRelocationLinksRespectDisplayToggle(t *testing.T) {
	s := relocSession()
	s.SetColonyRelocation(0, 3)
	if len(s.RelocationLinks()) != 1 {
		t.Fatalf("開著應有 1 條連線,實得 %d", len(s.RelocationLinks()))
	}
	s.ShowRelocationLines = false
	if got := s.RelocationLinks(); got != nil {
		t.Errorf("關掉應完全不畫,實得 %v", got)
	}
}

// TestNewShipTravelsToRallyPoint:設了集結點,新造的艦要**自己飛過去**。
//
// 手冊說的是 "ships being automatically relocated" —— 那是一段航程,
// 星圖上那條線畫的就是它。所以不能瞬間移動到目的地。
func TestNewShipTravelsToRallyPoint(t *testing.T) {
	s := relocSession()
	s.SetColonyRelocation(1, 2) // 殖民地 1 在星 4,集結點設在星 2

	s.deliverNewShip(1, Ship{Name: "新艦"})

	var found *Fleet
	for i := range s.Fleets {
		if len(s.Fleets[i].Ships) == 1 && s.Fleets[i].Ships[0].Name == "新艦" {
			found = &s.Fleets[i]
		}
	}
	if found == nil {
		t.Fatal("新艦應該出現在某一支艦隊裡")
	}
	if found.AtStar != 4 {
		t.Errorf("應從生產它的殖民地(星 4)出發,實得 %d", found.AtStar)
	}
	if found.DestStar != 2 {
		t.Errorf("目的地應是集結點星 2,實得 %d", found.DestStar)
	}
	if found.ETA < 1 {
		t.Errorf("ETA 至少 1 回合(0 會被 advanceFleet 當成已抵達而永遠不推進),實得 %d", found.ETA)
	}
}

// TestNewShipsToSameRallyShareOneFleet:同起訖的新艦要併進同一支,不能每艘生一支。
func TestNewShipsToSameRallyShareOneFleet(t *testing.T) {
	s := relocSession()
	s.SetColonyRelocation(1, 2)
	before := len(s.Fleets)
	s.deliverNewShip(1, Ship{Name: "甲"})
	s.deliverNewShip(1, Ship{Name: "乙"})
	if len(s.Fleets) != before+1 {
		t.Errorf("兩艘同起訖的新艦應共用一支艦隊,艦隊數 %d → %d", before, len(s.Fleets))
	}
}

// TestNoRallyKeepsShipAtColony:沒設集結點,新艦留在原地。
func TestNoRallyKeepsShipAtColony(t *testing.T) {
	s := relocSession()
	s.deliverNewShip(0, Ship{Name: "留守"}) // 殖民地 0 在星 0,艦隊 0 也在星 0
	if len(s.Fleets) != 1 {
		t.Fatalf("沒設集結點不該生出新艦隊,實得 %d 支", len(s.Fleets))
	}
	if len(s.Fleets[0].Ships) != 1 || s.Fleets[0].DestStar != -1 {
		t.Errorf("應併進原地那支且不啟程,實得 %d 艘、目的地 %d",
			len(s.Fleets[0].Ships), s.Fleets[0].DestStar)
	}
}

// TestFleetsArriveAndMerge:多艦隊各自航行,抵達同一顆星就併起來。
//
// 不併的話,自動遷移每回合可能生出一支,艦隊清單會無限長大——
// 玩家看到的是「同一顆星上一堆各有一兩艘船的艦隊」,那不是原版的樣子。
func TestFleetsArriveAndMerge(t *testing.T) {
	s := relocSession()
	s.Fleets = []Fleet{
		{AtStar: 0, DestStar: -1, Ships: []Ship{{Name: "駐守"}}},
		{AtStar: 4, DestStar: 0, ETA: 1, Ships: []Ship{{Name: "來會合"}}},
	}
	s.advanceFleet()
	if len(s.Fleets) != 1 {
		t.Fatalf("抵達後應併成 1 支,實得 %d 支", len(s.Fleets))
	}
	if len(s.Fleets[0].Ships) != 2 {
		t.Errorf("併起來應有 2 艘,實得 %d", len(s.Fleets[0].Ships))
	}
	if s.Fleets[0].AtStar != 0 {
		t.Errorf("會合點應是星 0,實得 %d", s.Fleets[0].AtStar)
	}
}

// TestTravellingFleetsAreNotMerged:還在航行的不能併——併了會弄丟各自的目的地。
func TestTravellingFleetsAreNotMerged(t *testing.T) {
	s := relocSession()
	s.Fleets = []Fleet{
		{AtStar: 0, DestStar: 2, ETA: 3, Ships: []Ship{{Name: "甲"}}},
		{AtStar: 0, DestStar: 5, ETA: 3, Ships: []Ship{{Name: "乙"}}},
	}
	s.advanceFleet()
	if len(s.Fleets) != 2 {
		t.Fatalf("兩支都還在航行(目的地不同),不該被併,實得 %d 支", len(s.Fleets))
	}
	if s.Fleets[0].DestStar != 2 || s.Fleets[1].DestStar != 5 {
		t.Errorf("各自的目的地應保留,實得 %d / %d", s.Fleets[0].DestStar, s.Fleets[1].DestStar)
	}
}

// ---- 原版的兩段點選規則(Okay_To_Set_Relocate_Star_ @ 0x75035)----

// TestRelocateFromMustBeOwnColony:起點必須是自己有殖民地的星。
func TestRelocateFromMustBeOwnColony(t *testing.T) {
	s := relocSession()
	for i := range s.Stars {
		s.Stars[i].Explored = true
	}
	if r := s.CanRelocateFrom(0); r != "" {
		t.Errorf("星 0 是自己的殖民地,應可當起點,實得拒絕:%s", r)
	}
	if r := s.CanRelocateFrom(3); r == "" {
		t.Error("星 3 沒有自己的殖民地,不該能當起點")
	}
}

// TestRelocateRejectsBlackHoleBothEnds:黑洞起訖都不行(原版兩種訊息,規則同一條)。
func TestRelocateRejectsBlackHoleBothEnds(t *testing.T) {
	s := relocSession()
	for i := range s.Stars {
		s.Stars[i].Explored = true
	}
	s.Stars[0].Spectral = 6 // 母星改成黑洞(只為驗規則)
	s.Stars[2].Spectral = 6
	if r := s.CanRelocateFrom(0); r == "" {
		t.Error("黑洞不該能當起點")
	}
	if r := s.CanRelocateTo(2); r == "" {
		t.Error("黑洞不該能當終點")
	}
}

// TestRelocateRejectsUnexplored:沒探索過的星不行。
//
// 原版查的是逐玩家的探索位元遮罩 `star[+0x33] & (1<<玩家)`;
// remake 的 Star.Explored 是單玩家版本,語意對應。
func TestRelocateRejectsUnexplored(t *testing.T) {
	s := relocSession()
	s.Stars[0].Explored = true
	s.Stars[2].Explored = false
	if r := s.CanRelocateTo(2); r == "" {
		t.Error("沒探索過的星不該能當終點")
	}
}

// TestSetStarRelocationTwoStage:兩段點選的完整流程,含「點回自己 = 取消」。
func TestSetStarRelocationTwoStage(t *testing.T) {
	s := relocSession()
	for i := range s.Stars {
		s.Stars[i].Explored = true
	}
	if r := s.SetStarRelocation(0, 3); r != "" {
		t.Fatalf("星 0 → 星 3 應成功,實得:%s", r)
	}
	if got := s.ColonyRelocation(0); got != 3 {
		t.Errorf("集結點應是 3,實得 %d", got)
	}
	// 起訖同一顆 = 取消(原版 Cancel_Star_Relocation_)。
	if r := s.SetStarRelocation(0, 0); r != "" {
		t.Fatalf("起訖同一顆應成功(語意是取消),實得:%s", r)
	}
	if got := s.ColonyRelocation(0); got != ColonyRelocationNone {
		t.Errorf("點回自己應取消,實得 %d", got)
	}
}
