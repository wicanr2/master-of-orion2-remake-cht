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
	if v := CouncilVotes(42); v != 42 {
		t.Errorf("CouncilVotes(42) = %d, want 42", v)
	}
	if v := CouncilVotes(0); v != 0 {
		t.Errorf("CouncilVotes(0) = %d, want 0", v)
	}
	if v := CouncilVotes(-5); v != 0 {
		t.Errorf("CouncilVotes(-5) = %d, want 0(已滅亡帝國無票)", v)
	}
}

// 2/3 超級多數門檻判定沿用 internal/engine.CheckHighCouncil,已有測試涵蓋
// (internal/engine/victory_test.go:TestCheckHighCouncilExactlyTwoThirds/OneVoteShort/
// InvalidTotal),不在本檔重複測試。
