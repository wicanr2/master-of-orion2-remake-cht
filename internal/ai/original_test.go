package ai

import "testing"

// TestAIDifficultyBonus 手算對照 MANUAL_150.html「Generic AI bonuses」原表
// (docs/tech/original-ai-re.md §2.1),逐難度抽查幾個代表欄位。
func TestAIDifficultyBonus(t *testing.T) {
	// Tutor:Growth=0、Food=-1/4(quarters=-1)、CommandDeficitBC=12、AntaranMarines=-4。
	tutor, ok := AIDifficultyBonus(DifficultyTutor)
	if !ok {
		t.Fatalf("DifficultyTutor: ok=false")
	}
	if tutor.GrowthPercent != 0 || tutor.FoodQuarters != -1 || tutor.CommandDeficitBC != 12 || tutor.AntaranMarines != -4 {
		t.Errorf("Tutor = %+v, want Growth=0 FoodQuarters=-1 CommandDeficitBC=12 AntaranMarines=-4", tutor)
	}
	if got := tutor.Food(); got != -0.25 {
		t.Errorf("Tutor.Food() = %v, want -0.25", got)
	}

	// Average:官方手冊基準難度,BC=1/4(quarters=1),Command=10,其餘加成欄全 0。
	avg, ok := AIDifficultyBonus(DifficultyAverage)
	if !ok {
		t.Fatalf("DifficultyAverage: ok=false")
	}
	if avg.GrowthPercent != 2 || avg.BCQuarters != 1 || avg.CommandDeficitBC != 10 || avg.SpyBonus != 0 {
		t.Errorf("Average = %+v, want Growth=2 BCQuarters=1 CommandDeficitBC=10 SpyBonus=0", avg)
	}
	if got := avg.BC(); got != 0.25 {
		t.Errorf("Average.BC() = %v, want 0.25", got)
	}

	// Impossible:題目點名核對 Command=8、Growth=+4;另抽查 Prod=2(quarters=8)、AntaranMarines=4。
	imp, ok := AIDifficultyBonus(DifficultyImpossible)
	if !ok {
		t.Fatalf("DifficultyImpossible: ok=false")
	}
	if imp.CommandDeficitBC != 8 {
		t.Errorf("Impossible.CommandDeficitBC = %d, want 8", imp.CommandDeficitBC)
	}
	if imp.GrowthPercent != 4 {
		t.Errorf("Impossible.GrowthPercent = %d, want 4", imp.GrowthPercent)
	}
	if imp.ProdQuarters != 8 || imp.AntaranMarines != 4 || imp.TroopsMarines != 2 || imp.SpyBonus != 2 {
		t.Errorf("Impossible = %+v, want ProdQuarters=8 AntaranMarines=4 TroopsMarines=2 SpyBonus=2", imp)
	}
	if got := imp.Prod(); got != 2.0 {
		t.Errorf("Impossible.Prod() = %v, want 2.0", got)
	}

	// Hard:抽查分數欄位換算(BC=2/4=0.5,Res=1)。
	hard, ok := AIDifficultyBonus(DifficultyHard)
	if !ok {
		t.Fatalf("DifficultyHard: ok=false")
	}
	if got := hard.BC(); got != 0.5 {
		t.Errorf("Hard.BC() = %v, want 0.5", got)
	}
	if got := hard.Res(); got != 1.0 {
		t.Errorf("Hard.Res() = %v, want 1.0", got)
	}

	// 超出範圍:ok 必須是 false,且回傳零值(不可誤當 Tutor)。
	if _, ok := AIDifficultyBonus(Difficulty(-1)); ok {
		t.Errorf("Difficulty(-1): ok = true, want false")
	}
	if _, ok := AIDifficultyBonus(Difficulty(5)); ok {
		t.Errorf("Difficulty(5): ok = true, want false")
	}
}

// TestClassicRacePersonality 手算對照 AIRACES.CFG classic 值(docs/tech/original-ai-re.md §1.3),
// 抽查幾族,含 mod≠classic 的 Humans/Trilarians(確認取的是 classic,不是 mod)。
func TestClassicRacePersonality(t *testing.T) {
	// Klackons classic:0 0 0 0 0 0 0 1 2 2 → Xenophobic 70% / Ruthless 10% / Aggressive 20%。
	klackons, ok := ClassicRacePersonality("Klackons")
	if !ok {
		t.Fatalf("Klackons: ok=false")
	}
	want := [10]Personality{
		PersonalityXenophobic, PersonalityXenophobic, PersonalityXenophobic, PersonalityXenophobic,
		PersonalityXenophobic, PersonalityXenophobic, PersonalityXenophobic,
		PersonalityRuthless,
		PersonalityAggressive, PersonalityAggressive,
	}
	if klackons != want {
		t.Errorf("Klackons = %v, want %v", klackons, want)
	}

	// Psilons classic:3 3 4 5 5 5 5 5 5 5 → Erratic 20% / Honorable 10% / Pacifist 70%。
	psilons, ok := ClassicRacePersonality("Psilons")
	if !ok {
		t.Fatalf("Psilons: ok=false")
	}
	wantPsilons := [10]Personality{
		PersonalityErratic, PersonalityErratic,
		PersonalityHonorable,
		PersonalityPacifist, PersonalityPacifist, PersonalityPacifist, PersonalityPacifist,
		PersonalityPacifist, PersonalityPacifist, PersonalityPacifist,
	}
	if psilons != wantPsilons {
		t.Errorf("Psilons = %v, want %v", psilons, wantPsilons)
	}

	// Humans:AIRACES.CFG 的 `=` mod 值是 3 3 3 4 4 4 4 4 5 5,`##` classic 值是
	// 3 4 4 4 4 4 4 4 5 5——必須取 classic(索引 1、2 都要是 Honorable,不是 mod 的 3/3)。
	humans, ok := ClassicRacePersonality("Humans")
	if !ok {
		t.Fatalf("Humans: ok=false")
	}
	if humans[0] != PersonalityErratic {
		t.Errorf("Humans[0] = %v, want Erratic(classic 首格=3),確認沒誤用 mod 值", humans[0])
	}
	if humans[1] != PersonalityHonorable || humans[2] != PersonalityHonorable {
		t.Errorf("Humans[1..2] = %v %v, want Honorable Honorable(classic),mod 值錯誤地會是 Erratic Erratic(3 3)", humans[1], humans[2])
	}

	// Trilarians:mod 值第 2 格是 3(Erratic),classic 第 2 格是 4(Honorable)——同樣驗證取 classic。
	trilarians, ok := ClassicRacePersonality("Trilarians")
	if !ok {
		t.Fatalf("Trilarians: ok=false")
	}
	if trilarians[1] != PersonalityHonorable {
		t.Errorf("Trilarians[1] = %v, want Honorable(classic),mod 值錯誤地會是 Erratic", trilarians[1])
	}

	// classic 分布表裡沒有任何種族被指派 Dishonored(6)——與 §1.1「Dishonored 是外交狀態,非開局
	// 性格」的結論一致,13 族全數掃描確認。
	all := []string{"Alkari", "Bulrathi", "Darloks", "Elerians", "Gnolams", "Humans", "Klackons",
		"Meklars", "Mrrshan", "Psilons", "Sakkra", "Silicoids", "Trilarians"}
	if len(all) != len(classicRacePersonality) {
		t.Fatalf("classicRacePersonality 族數 = %d, want %d", len(classicRacePersonality), len(all))
	}
	for _, race := range all {
		dist, ok := ClassicRacePersonality(race)
		if !ok {
			t.Errorf("ClassicRacePersonality(%q): ok=false", race)
			continue
		}
		for i, p := range dist {
			if p == PersonalityDishonored {
				t.Errorf("%s[%d] = Dishonored,classic 分布表不應出現此值", race, i)
			}
		}
	}

	// 查無種族:ok 必須是 false。
	if _, ok := ClassicRacePersonality("NotARace"); ok {
		t.Errorf("ClassicRacePersonality(NotARace): ok = true, want false")
	}
}
