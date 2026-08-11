package gamedata

import "testing"

// 手冊 p.85 那句話的三個限定詞各驗一條。
func TestFoodReplicatorConvertFollowsTheManual(t *testing.T) {
	// two-for-one:補 3 單位食物要花 6 點產能。
	food, spent := FoodReplicatorConvert(3, 100)
	if food != 3 || spent != 6 {
		t.Errorf("缺口 3、產能充足時應換 3 食物花 6 產能,得到 %d / %d", food, spent)
	}
	// as needed:產能再多也只補到缺口為止,不會換出盈餘。
	food, spent = FoodReplicatorConvert(2, 1000)
	if food != 2 || spent != 4 {
		t.Errorf("「as needed」應只補缺口 2,得到 %d 食物 / %d 產能", food, spent)
	}
	// 產能不夠時補多少算多少(整數除法,不會超支)。
	food, spent = FoodReplicatorConvert(10, 7)
	if food != 3 || spent != 6 {
		t.Errorf("產能 7 只換得起 3 食物(花 6),得到 %d / %d", food, spent)
	}
	if spent > 7 {
		t.Errorf("花掉的產能 %d 超過可用的 7", spent)
	}
}

// 沒有缺口就什麼都不做——這條是「不變成印鈔機」的關鍵。
func TestFoodReplicatorDoesNothingWithoutADeficit(t *testing.T) {
	for _, deficit := range []int{0, -1, -50} {
		if food, spent := FoodReplicatorConvert(deficit, 1000); food != 0 || spent != 0 {
			t.Errorf("缺口 %d 應不換算,得到 %d 食物 / %d 產能", deficit, food, spent)
		}
	}
	if food, spent := FoodReplicatorConvert(5, 0); food != 0 || spent != 0 {
		t.Errorf("沒有產能應換不出東西,得到 %d / %d", food, spent)
	}
	if food, spent := FoodReplicatorConvert(5, -3); food != 0 || spent != 0 {
		t.Errorf("負產能應換不出東西,得到 %d / %d", food, spent)
	}
}

func TestFoodReplicatorConvertHalfPreservesCyberneticHalfFood(t *testing.T) {
	foodHalf, spentHalf := FoodReplicatorConvertHalf(1, 2)
	if foodHalf != 1 || spentHalf != 2 {
		t.Fatalf("半食物缺口應換出 1 半食物並花 2 半產能,got %d / %d", foodHalf, spentHalf)
	}
	foodHalf, spentHalf = FoodReplicatorConvertHalf(5, 5)
	if foodHalf != 2 || spentHalf != 4 {
		t.Fatalf("半產能不足時應按 2 半產能/半食物夾住,got %d / %d", foodHalf, spentHalf)
	}
}

// 維護費 10 BC:手冊與建築表(來自原版執行檔)兩個來源。
// 順帶釘住「它是全表最貴」——那是這棟建築的平衡設計,被改小就失衡了。
func TestFoodReplicatorMaintenanceIsTheHighestInTheTable(t *testing.T) {
	var mine, max int
	maxName := ""
	found := false
	for _, b := range Buildings {
		if b.MaintenanceBC > max {
			max, maxName = b.MaintenanceBC, b.NameZH
		}
		if b.NameEN == "Food Replicators" {
			mine, found = b.MaintenanceBC, true
		}
	}
	if !found {
		t.Fatal("建築表裡找不到食物複製機")
	}
	if mine != 10 {
		t.Errorf("手冊 p.85 說 10 BC,建築表是 %d", mine)
	}
	if mine != max {
		t.Errorf("食物複製機應是全表最貴(%d),但最貴的是 %s 的 %d", mine, maxName, max)
	}
}
