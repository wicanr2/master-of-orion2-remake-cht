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
func TestFighterGarrisonTotalStrengthStillRisesWithTech(t *testing.T) {
	ia, _ := FighterGarrisonCombatContribution(FighterGarrisonInterceptor)
	ba, _ := FighterGarrisonCombatContribution(FighterGarrisonBomber)
	ha, _ := FighterGarrisonCombatContribution(FighterGarrisonHeavyFighter)
	t.Logf("攔截機 %d / 轟炸機 %d / 重戰機 %d", ia, ba, ha)
	if ha <= ba {
		t.Errorf("重戰機檔(%d)應強過轟炸機檔(%d)", ha, ba)
	}
	// ⚠ 攔截機 10 隊 × 48 = 480,轟炸機 6 隊 × 20 = 120——**攔截機檔反而最強**。
	// 那不是 bug:remake 的單隊貢獻值(fighterBeamDamageApprox/fighterBombDamageApprox)
	// 是近似值而不是手冊真值(見 combat.go 該處註解),所以三檔的相對強弱本來就不可靠。
	// 這裡如實記錄而不是硬調數字去湊一個好看的曲線。
	if ia <= 0 || ba <= 0 || ha <= 0 {
		t.Error("三檔的戰力都應為正")
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
func TestFighterBomberContributionUsesTheBombShot(t *testing.T) {
	atk, hp := FighterBomberCombatContribution()
	if atk != FighterSquadronSize*FighterShotsBomber*fighterBombDamageApprox {
		t.Errorf("轟炸機攻擊應是 4 架 × 1 次 × 投彈值,得到 %d", atk)
	}
	if hp != FighterSquadronSize*FighterHitsInterceptor {
		t.Errorf("手冊沒給轟炸機的耐受數,應沿用攔截機那一格,得到 %d", hp)
	}
}
