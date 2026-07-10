package engine

import "testing"

func TestHireLeader(t *testing.T) {
	cases := []struct {
		name      string
		currentBC int
		cost      int
		wantBC    int
		wantOK    bool
	}{
		{"BC 足夠,扣款成功", 100, 40, 60, true},
		{"BC 剛好等於花費,成功", 40, 40, 0, true},
		{"BC 不足,拒絕且不扣款", 30, 40, 30, false},
		{"cost 為0(如 Megawealth 全免),必成功且不扣款", 10, 0, 10, true},
		{"cost 為負數視為0", 10, -5, 10, true},
	}
	for _, c := range cases {
		gotBC, gotOK := HireLeader(c.currentBC, c.cost)
		if gotBC != c.wantBC || gotOK != c.wantOK {
			t.Errorf("%s: HireLeader(%d,%d) = (%d,%v),預期 (%d,%v)",
				c.name, c.currentBC, c.cost, gotBC, gotOK, c.wantBC, c.wantOK)
		}
	}
}
