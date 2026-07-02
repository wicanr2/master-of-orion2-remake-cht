package gamedata

import "testing"

func TestColonyBaseGrowth(t *testing.T) {
	// a = trunc(sqrt(FACTOR1*POPRACE*(POPMAX-POPAGG)/POPMAX)),FACTOR1=2000
	cases := []struct {
		popRace, popAgg, popMax, want int
	}{
		{5, 5, 10, 70},   // 2000*5*5/10=5000,sqrt≈70.7→70
		{1, 0, 10, 44},   // 2000*1*10/10=2000,sqrt≈44.7→44
		{10, 0, 10, 141}, // 2000*10*10/10=20000,sqrt≈141.4→141
		{5, 10, 10, 0},   // popAgg>=popMax→滿→0
		{5, 12, 10, 0},   // 超載→0
		{0, 0, 10, 0},    // 無該種族人口→0
		{5, 0, 0, 0},     // popMax 非法→0
	}
	for _, c := range cases {
		if got := ColonyBaseGrowth(c.popRace, c.popAgg, c.popMax); got != c.want {
			t.Errorf("ColonyBaseGrowth(%d,%d,%d) = %d,預期 %d",
				c.popRace, c.popAgg, c.popMax, got, c.want)
		}
	}
}

func TestColonyHousingBonus(t *testing.T) {
	// h = FACTOR2*PROD/POPAGG,FACTOR2=40
	if got := ColonyHousingBonus(40, 10); got != 160 { // 40*40/10
		t.Errorf("ColonyHousingBonus(40,10) = %d,預期 160", got)
	}
	if got := ColonyHousingBonus(25, 5); got != 200 { // 40*25/5
		t.Errorf("ColonyHousingBonus(25,5) = %d,預期 200", got)
	}
	if got := ColonyHousingBonus(40, 0); got != 0 {
		t.Errorf("popAgg=0 應回 0,得 %d", got)
	}
}

func TestColonyGrowth(t *testing.T) {
	// growth = a*(100+bonusSum)/100
	cases := []struct{ a, bonusSum, want int }{
		{70, 0, 70},     // b=1
		{70, 50, 105},   // +50%
		{100, 100, 200}, // +100%
		{100, -50, 50},  // 種族獎金 −50%
	}
	for _, c := range cases {
		if got := ColonyGrowth(c.a, c.bonusSum); got != c.want {
			t.Errorf("ColonyGrowth(%d,%d) = %d,預期 %d", c.a, c.bonusSum, got, c.want)
		}
	}
}
