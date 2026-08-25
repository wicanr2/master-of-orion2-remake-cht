package gamedata

import "testing"

func TestCouncilEligible(t *testing.T) {
	cases := []struct {
		name                            string
		settledStars, totalStars, races int
		want                            bool
	}{
		{"未達半數已殖民", 5, 24, 3, false},
		{"剛好半數已殖民+3種族", 12, 24, 3, true},
		{"奇數星數採原版向下取整", 2, 5, 3, true},
		{"奇數星數仍未達向下取整", 1, 5, 3, false},
		{"超過半數但只有2種族(手冊字面值不成立)", 20, 24, 2, false},
		{"總星數0視為不成立", 0, 0, 3, false},
		{"半數以上+4種族", 18, 24, 4, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CouncilEligible(c.settledStars, c.totalStars, c.races)
			if got != c.want {
				t.Errorf("CouncilEligible(%d,%d,%d) = %v, want %v", c.settledStars, c.totalStars, c.races, got, c.want)
			}
		})
	}
}

func TestCouncilVotes(t *testing.T) {
	tests := []struct {
		population int
		want       int
	}{
		{-5, 0},
		{0, 0},
		{1, 1},
		{10, 1},
		{11, 2},
		{42, 5},
	}
	for _, test := range tests {
		if got := CouncilVotes(test.population); got != test.want {
			t.Errorf("CouncilVotes(%d) = %d, want %d", test.population, got, test.want)
		}
	}
}

// 2/3 超級多數門檻判定沿用 internal/engine.CheckHighCouncil,已有測試涵蓋
// (internal/engine/victory_test.go:TestCheckHighCouncilExactlyTwoThirds/OneVoteShort/
// InvalidTotal),不在本檔重複測試。
