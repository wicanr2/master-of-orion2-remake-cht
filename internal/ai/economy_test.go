package ai

import "testing"

func TestDecideColonyJobs(t *testing.T) {
	cases := []struct {
		pop, fpf   int
		p          Profile
		wF, wW, wS int
	}{
		{10, 5, ProfileBalanced, 2, 4, 4},   // 農2餵飽,餘8均分
		{10, 5, ProfileAggressive, 2, 6, 2}, // 餘8 → 工6研2(3:1)
		{10, 5, ProfileScientific, 2, 2, 6}, // 餘8 → 工2研6(1:3)
		{6, 2, ProfileBalanced, 3, 1, 2},    // 農3,餘3 → 工1研2
		{10, 0, ProfileBalanced, 0, 5, 5},   // 無法務農 → 全分工研
		{0, 5, ProfileBalanced, 0, 0, 0},    // 無人口
	}
	for _, c := range cases {
		f, w, s := DecideColonyJobs(c.pop, c.fpf, c.p)
		if f != c.wF || w != c.wW || s != c.wS {
			t.Errorf("DecideColonyJobs(%d,%d,%s) = (%d,%d,%d),預期 (%d,%d,%d)",
				c.pop, c.fpf, c.p.Name, f, w, s, c.wF, c.wW, c.wS)
		}
		// 驗證餵得飽:農夫產食 >= 人口(foodPerFarmer>0 時)
		if c.fpf > 0 && f*c.fpf < c.pop {
			t.Errorf("分配未餵飽人口:農%d*%d < 人口%d", f, c.fpf, c.pop)
		}
		// 驗證總分配 = 人口
		if f+w+s != c.pop {
			t.Errorf("分配總和 %d != 人口 %d", f+w+s, c.pop)
		}
	}
}

func TestDecideTaxRate(t *testing.T) {
	if DecideTaxRate(5, 20, 100) != 50 {
		t.Error("國庫低應提高稅率至 50")
	}
	if DecideTaxRate(50, 20, 100) != 30 {
		t.Error("國庫中等應 30")
	}
	if DecideTaxRate(200, 20, 100) != 10 {
		t.Error("國庫充裕應降至 10")
	}
}
