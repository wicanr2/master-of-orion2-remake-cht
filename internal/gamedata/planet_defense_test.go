package gamedata

import "testing"

// 手冊 p.79 的三個中隊數,而且**隨科技遞減**——每階戰機更強,所以數量往下走。
// 這條把「順手改成遞增」擋掉。
func TestFighterGarrisonSquadronsDecreaseWithTechTier(t *testing.T) {
	i := FighterGarrisonSquadrons(FighterGarrisonInterceptor)
	b := FighterGarrisonSquadrons(FighterGarrisonBomber)
	h := FighterGarrisonSquadrons(FighterGarrisonHeavyFighter)
	if i != 10 || b != 6 || h != 4 {
		t.Errorf("手冊是 10 / 6 / 4,得到 %d / %d / %d", i, b, h)
	}
	if !(i > b && b > h) {
		t.Error("中隊數應隨科技遞減(10 → 6 → 4),不是遞增")
	}
	// 超出範圍退回攔截機那一檔(手冊:available immediately)。
	if got := FighterGarrisonSquadrons(FighterGarrisonTier(99)); got != i {
		t.Errorf("未知檔次應退回攔截機的 %d,得到 %d", i, got)
	}
}

// 中隊數雖然遞減,整座基地的戰力仍應往上走——那正是「每階更強」的意思。
// 這條是上一條的另一半:少了它,把三個數字寫成 10/6/4 但戰力遞減也會通過。
func TestFighterGarrisonStrategicStrengthMatchesIDAFormula(t *testing.T) {
	if got := FighterGarrisonStrategicStrength(FighterGarrisonInterceptor, 4, 12, 0); got != 80 {
		t.Errorf("攔截機應為 (4*40)/2=80，得到 %d", got)
	}
	if got := FighterGarrisonStrategicStrength(FighterGarrisonBomber, 30, 40, 3); got != 984 {
		t.Errorf("轟炸機應為 ((30-3)*40+(40-3)*24)/2=984，得到 %d", got)
	}
	if got := FighterGarrisonStrategicStrength(FighterGarrisonHeavyFighter, 30, 40, 5); got != 820 {
		t.Errorf("重戰機應為 ((30-5)*32+(40-5)*24)/2=820，得到 %d", got)
	}
	if got := FighterGarrisonStrategicStrength(FighterGarrisonBomber, 3, 4, 10); got != 0 {
		t.Errorf("裝甲扣減後不得為負，得到 %d", got)
	}
	if got := FighterGarrisonStrategicStrength(FighterGarrisonHeavyFighter, 10000, 10000, 0); got != 64000 {
		t.Errorf("上限應為 64000，得到 %d", got)
	}
}

// 恆星轉換器:400/面、1600 總,取單面當反擊攻擊值。
func TestStellarConverterNumbers(t *testing.T) {
	if StellarConverterDamagePerSide != 400 {
		t.Errorf("手冊是每面 400,得到 %d", StellarConverterDamagePerSide)
	}
	if StellarConverterTotalDamage != 1600 {
		t.Errorf("手冊是總計 1600,得到 %d", StellarConverterTotalDamage)
	}
	if StellarConverterTotalDamage != StellarConverterDamagePerSide*StellarConverterSides {
		t.Error("1600 應等於 400 × 面數——兩個數字是同一段手冊給的,不能各走各的")
	}
	// 反擊取單面:抽象解算沒有「面」的概念,一次反擊不是同時打四面。
	if got := StellarConverterRetaliationAttack(); got != StellarConverterDamagePerSide {
		t.Errorf("反擊攻擊值應是單面的 %d,得到 %d", StellarConverterDamagePerSide, got)
	}
}

// 維護費:手冊與建築表(來自原版執行檔)兩個來源。
func TestPlanetDefenseMaintenanceMatchesTheBuildingTable(t *testing.T) {
	want := map[string]int{"Fighter Garrison": 2, "Stellar Converter": 6}
	found := 0
	for _, b := range Buildings {
		if w, ok := want[b.NameEN]; ok {
			found++
			if b.MaintenanceBC != w {
				t.Errorf("%s 手冊 %d BC,建築表 %d", b.NameEN, w, b.MaintenanceBC)
			}
		}
	}
	if found != len(want) {
		t.Fatalf("建築表裡應找到 %d 棟,找到 %d 棟", len(want), found)
	}
}

// 轟炸機中隊的貢獻:每次只投彈一發(FighterShotsBomber = 1),炸彈必中。
func TestFighterGarrisonRawTables(t *testing.T) {
	for _, id := range []int{1, 3, 4, 5, 9} {
		if !FighterGarrisonBeamWeaponEligible(id) {
			t.Errorf("武器 %d 應有 fighter eligibility", id)
		}
	}
	if FighterGarrisonBeamWeaponEligible(41) {
		t.Error("怪物武器 41 不得進玩家選擇器")
	}
	want := map[Technology]int{TECH_TITANIUM_ARMOR: 0, TECH_TRITANIUM_ARMOR: 1, TECH_ZORTRIUM_ARMOR: 3, TECH_NEUTRONIUM_ARMOR: 5, TECH_ADAMANTIUM_ARMOR: 7, TECH_XENTRONIUM_ARMOR: 10}
	for tech, reduction := range want {
		if got := FighterGarrisonArmorReduction(tech); got != reduction {
			t.Errorf("裝甲 %d 減傷=%d，預期 %d", tech, got, reduction)
		}
	}
}
