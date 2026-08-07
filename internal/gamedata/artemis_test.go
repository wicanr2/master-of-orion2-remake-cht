package gamedata

import "testing"

// 手冊 p.86 的六個觸發機率逐項釘住。
func TestArtemisTriggerPercentMatchesTheManual(t *testing.T) {
	for _, tc := range []struct {
		hull ArtemisHullClass
		want int
		name string
	}{
		{ArtemisFrigate, 20, "Frigate"},
		{ArtemisDestroyer, 30, "Destroyer"},
		{ArtemisCruiser, 40, "Cruiser"},
		{ArtemisBattleship, 50, "Battleship"},
		{ArtemisTitan, 80, "Titan"},
		{ArtemisDoomStar, 100, "Doom Star"},
	} {
		if got := ArtemisTriggerPercent(tc.hull); got != tc.want {
			t.Errorf("%s 應為 %d%%,得到 %d%%", tc.name, tc.want, got)
		}
	}
	// 機率隨體積單調上升——這是「水雷網專打主力艦」的性質,被改亂就不是原版了。
	prev := -1
	for h := ArtemisFrigate; h <= ArtemisDoomStar; h++ {
		p := ArtemisTriggerPercent(h)
		if p <= prev {
			t.Errorf("艦體等級 %d 的機率 %d 沒有比前一級 %d 高", h, p, prev)
		}
		prev = p
	}
	// 超出範圍退回最小艦體的機率,不是 0——回 0 會讓分類失誤變成「穿得過雷區」。
	if got := ArtemisTriggerPercent(ArtemisHullClass(99)); got != 20 {
		t.Errorf("未知艦體應退回 Frigate 的 20%%,得到 %d", got)
	}
}

// 每枚 20 − 護盾等級,而且不會是負的。
func TestArtemisMineDamageSubtractsShieldClass(t *testing.T) {
	for _, tc := range []struct{ shield, want int }{
		{0, 20}, {1, 19}, {3, 17}, {5, 15}, {7, 13}, {10, 10},
	} {
		if got := ArtemisMineDamage(tc.shield); got != tc.want {
			t.Errorf("護盾等級 %d 每枚應 %d 傷,得到 %d", tc.shield, tc.want, got)
		}
	}
	if got := ArtemisMineDamage(50); got != 0 {
		t.Errorf("護盾等級超過 20 時應為 0 而非負數,得到 %d", got)
	}
}

// 水雷數 8–28,骰值超出範圍要夾住。
func TestArtemisMineCountStaysWithinTheManualRange(t *testing.T) {
	if got := ArtemisMineCount(0); got != ArtemisMinesMin {
		t.Errorf("骰 0 應是最少的 %d 枚,得到 %d", ArtemisMinesMin, got)
	}
	if got := ArtemisMineCount(ArtemisMineRollSpan); got != ArtemisMinesMax {
		t.Errorf("骰滿應是最多的 %d 枚,得到 %d", ArtemisMinesMax, got)
	}
	for _, roll := range []int{-5, 0, 7, 20, 999} {
		n := ArtemisMineCount(roll)
		if n < ArtemisMinesMin || n > ArtemisMinesMax {
			t.Errorf("骰值 %d 算出 %d 枚,超出 %d–%d", roll, n, ArtemisMinesMin, ArtemisMinesMax)
		}
	}
}

// 三件事相乘:護盾在這裡的價值被放大。
func TestArtemisShipDamageMultipliesMinesByPerMineDamage(t *testing.T) {
	if got := ArtemisShipDamage(28, 0); got != 28*20 {
		t.Errorf("無護盾挨滿 28 枚應 %d 傷,得到 %d", 28*20, got)
	}
	if got := ArtemisShipDamage(28, 10); got != 28*10 {
		t.Errorf("第十級護盾挨滿 28 枚應 %d 傷,得到 %d", 28*10, got)
	}
	// 第十級護盾在最壞情況下省下的量。寫出來是因為這正是這棟建築的權衡點。
	if saved := ArtemisShipDamage(28, 0) - ArtemisShipDamage(28, 10); saved != 280 {
		t.Errorf("第十級護盾在滿命中時應省下 280 點,得到 %d", saved)
	}
	if got := ArtemisShipDamage(0, 0); got != 0 {
		t.Errorf("沒中招應 0 傷,得到 %d", got)
	}
}

// 維護費 5 BC:手冊與建築表(來自原版執行檔)兩個來源。
func TestArtemisMaintenanceMatchesTheBuildingTable(t *testing.T) {
	for _, b := range Buildings {
		if b.NameEN == "Artemis System Net" {
			if b.MaintenanceBC != 5 {
				t.Errorf("手冊 p.86 說 5 BC,建築表是 %d", b.MaintenanceBC)
			}
			return
		}
	}
	t.Fatal("建築表裡找不到阿提米絲系統網")
}
