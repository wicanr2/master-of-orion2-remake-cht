package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestStartingBuildingCountManualExample 驗證手冊 worked example(docs/tech/homeworld-init.md
// §3.5):HW pop 8 → Advanced 上限內可有 6 棟建築,但 Average 因上限 5 只能有 5 棟。
func TestStartingBuildingCountManualExample(t *testing.T) {
	if got := StartingBuildingCount(8, BuildingCapAdvanced); got != 6 {
		t.Fatalf("pop8/Advanced 應為 6:got %d", got)
	}
	if got := StartingBuildingCount(8, BuildingCapAverage); got != 5 {
		t.Fatalf("pop8/Average 應因上限降為 5:got %d", got)
	}
}

// TestStartingBuildingCountCapsAndRounding 驗證 ⅔ pop 無條件進位 + 上限截斷的邊界情形。
func TestStartingBuildingCountCapsAndRounding(t *testing.T) {
	cases := []struct{ pop, cap, want int }{
		{0, BuildingCapAverage, 0},
		{1, BuildingCapAverage, 1},                    // ceil(2/3)=1
		{2, BuildingCapAverage, 2},                    // ceil(4/3)=2
		{3, BuildingCapAverage, 2},                    // ceil(6/3)=2
		{4, BuildingCapAverage, 3},                    // ceil(8/3)=3
		{100, BuildingCapPreWarp, BuildingCapPreWarp}, // 上限截斷
	}
	for _, c := range cases {
		if got := StartingBuildingCount(c.pop, c.cap); got != c.want {
			t.Fatalf("StartingBuildingCount(%d,%d)=%d,want %d", c.pop, c.cap, got, c.want)
		}
	}
}

// TestNewDemoSessionHomeworldState 驗證 NewDemoSession 產生忠實 Average 起始文明等級母星
// (docs/tech/homeworld-init.md):單一母星、Marine Barracks+Star Base 已建、起始科技已知、
// 1 殖民船+2 偵察艦的起始艦隊。
func TestNewDemoSessionHomeworldState(t *testing.T) {
	s := NewDemoSession()

	if len(s.PlayerColonies) != 1 {
		t.Fatalf("玩家應只有 1 座母星,got %d", len(s.PlayerColonies))
	}
	hw := s.PlayerColonies[0]
	if hw.PlanetSize != gamedata.LARGE_PLANET {
		t.Fatalf("母星應為 Large,got %v", hw.PlanetSize)
	}
	if hw.Population != 8 {
		t.Fatalf("母星起始人口應為 8(手冊未給精確值,採 worked example 值),got %d", hw.Population)
	}

	if len(s.ColonyBuildings) != 1 {
		t.Fatalf("應有 1 份殖民地建築紀錄,got %d", len(s.ColonyBuildings))
	}
	b := s.ColonyBuildings[0]
	if !b["海軍陸戰隊營"] || !b["星基"] {
		t.Fatalf("母星應已建成海軍陸戰隊營+星基,got %+v", b)
	}
	if len(b) != 2 {
		t.Fatalf("Average 起始母星應僅 2 項常駐建築(不含 Capitol/Colony Base),got %+v", b)
	}

	if !s.Player.CompletedTopics[gamedata.TOPIC_STARTING_TECH] {
		t.Fatal("Tech field 0(Capitol/Spy Network/Pulse Rifle)應標記已知")
	}
	if !s.Player.CompletedTopics[gamedata.TOPIC_ENGINEERING] {
		t.Fatal("Tech field Engineering(Colony Base/Star Base/Marine Barracks)應標記已知")
	}

	if len(s.Ships) != 3 {
		t.Fatalf("起始艦隊應為 3 艘(1 殖民船+2 偵察艦),got %d", len(s.Ships))
	}
	classCount := map[string]int{}
	for _, sh := range s.Ships {
		classCount[sh.Class]++
	}
	if classCount["殖民船"] != 1 {
		t.Fatalf("應有 1 艘殖民船,got %d", classCount["殖民船"])
	}
	if classCount["偵察艦"] != 2 {
		t.Fatalf("應有 2 艘偵察艦,got %d", classCount["偵察艦"])
	}

	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) != 1 {
		t.Fatal("AI 對手應同樣只有 1 座母星(對稱)")
	}
}
