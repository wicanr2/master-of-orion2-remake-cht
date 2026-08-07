package gamedata

import "testing"

// 「哪些科技算生物武器」查的是**執行檔的 category 表**,不是這裡寫死的名單。
// 這一條把兩者釘在一起:category 20 的成員恰好是死亡孢子與生物滅絕者,不多不少。
//
// 少了這條,哪天有人「順手」把某項科技的 category 改掉,生物武器規則會靜默套到別的科技上。
func TestBiologicalWeaponsAreExactlyCategory20(t *testing.T) {
	var got []Technology
	for idx, cat := range TechItemCategory {
		if cat == TechCategoryBiologicalWeapon {
			got = append(got, Technology(idx))
		}
	}
	want := []Technology{TECH_BIOTERMINATOR, TECH_DEATH_SPORES}
	if len(got) != len(want) {
		t.Fatalf("category %d 的成員數應為 %d,得到 %d(%v)", TechCategoryBiologicalWeapon, len(want), len(got), got)
	}
	for _, w := range want {
		if !IsBiologicalWeapon(w) {
			t.Errorf("%v(%s)應被判定為生物武器", w, TechnologyName(w))
		}
	}
	// 反面:隨手挑幾項不同類別的科技都不該中。
	for _, notBio := range []Technology{TECH_SOIL_ENRICHMENT, TECH_CLONING_CENTER, TECH_UNIVERSAL_ANTIDOTE, TECH_POWERED_ARMOR} {
		if IsBiologicalWeapon(notBio) {
			t.Errorf("%v(%s)不是生物武器卻被判定為是", notBio, TechnologyName(notBio))
		}
	}
}

// 手冊 p.99 的兩個效果數字:死亡孢子 10%、生物滅絕者 20%。
func TestBiologicalWeaponKillPercentMatchesManual(t *testing.T) {
	cases := []struct {
		tech Technology
		want int
	}{
		{TECH_DEATH_SPORES, 10},
		{TECH_BIOTERMINATOR, 20},
		{TECH_SOIL_ENRICHMENT, 0}, // 不是生物武器 → 0
	}
	for _, c := range cases {
		if got := BiologicalWeaponKillPercent(c.tech); got != c.want {
			t.Errorf("%s 每莢殺傷機率應為 %d%%,得到 %d%%", TechnologyName(c.tech), c.want, got)
		}
	}
}

// 兩項都有時取**最強**那一個,不是加總——一次轟炸投的是同一種莢。
func TestBestBiologicalWeaponTakesTheStrongerNotTheSum(t *testing.T) {
	both := func(tech Technology) bool {
		return tech == TECH_DEATH_SPORES || tech == TECH_BIOTERMINATOR
	}
	if got := BestBiologicalWeaponKillPercent(both); got != BioWeaponBioTerminatorKillPercent {
		t.Errorf("兩項都有時應取 20%%(不是相加的 30%%),得到 %d%%", got)
	}
	only := func(tech Technology) bool { return tech == TECH_DEATH_SPORES }
	if got := BestBiologicalWeaponKillPercent(only); got != BioWeaponDeathSporesKillPercent {
		t.Errorf("只有死亡孢子時應為 10%%,得到 %d%%", got)
	}
	none := func(Technology) bool { return false }
	if got := BestBiologicalWeaponKillPercent(none); got != 0 {
		t.Errorf("兩項都沒有時應為 0%%,得到 %d%%", got)
	}
}

// 只有**屏障**護盾擋得住生物武器。輻射護盾與通量護盾的手冊敘述都沒有那一句
// ——這條測試存在的意義是:那不是漏寫,是三段文字的實際差異,不准「順手補齊」。
func TestOnlyTheBarrierShieldBlocksBiologicalWeapons(t *testing.T) {
	cases := []struct {
		building string
		want     bool
	}{
		{BuildingPlanetaryBarrierShield, true},
		{BuildingPlanetaryRadiationShield, false},
		{BuildingPlanetaryFluxShield, false},
	}
	for _, c := range cases {
		got := BiologicalWeaponBlocked(map[string]bool{c.building: true})
		if got != c.want {
			t.Errorf("%s:擋生物武器應為 %v,得到 %v", c.building, c.want, got)
		}
	}
	if BiologicalWeaponBlocked(nil) {
		t.Error("沒有任何建築時不該擋住生物武器")
	}
}

// 擲骰:每莢獨立擲一次 rand(100) < killPercent。
//
// roll 由測試控制,所以這裡驗的是**規則**(幾莢擲幾次、邊界怎麼算),不是隨機性。
func TestBiologicalWeaponPopKillsRollsOncePerPod(t *testing.T) {
	// 每一擲都回 0(< 任何正機率)→ 每莢都殺。
	rolls := 0
	always := func(n int) int { rolls++; return 0 }
	if got := BiologicalWeaponPopKills(7, 10, always); got != 7 {
		t.Errorf("每擲必中時 7 莢應殺 7 單位,得到 %d", got)
	}
	if rolls != 7 {
		t.Errorf("應每莢擲一次(7 次),實際擲了 %d 次", rolls)
	}
	// 邊界:roll 回傳恰好等於 killPercent 時**不算中**(< 而非 <=),
	// 否則 10% 實際會變成 11%。
	exact := func(n int) int { return 10 }
	if got := BiologicalWeaponPopKills(7, 10, exact); got != 0 {
		t.Errorf("roll 恰等於機率門檻時不該算中,得到 %d 個命中", got)
	}
	// 沒有生物武器(0%)或沒有船(0 莢)時完全不擲骰。
	rolls = 0
	if got := BiologicalWeaponPopKills(7, 0, always); got != 0 || rolls != 0 {
		t.Errorf("0%% 時應回 0 且不擲骰,得到 kills=%d rolls=%d", got, rolls)
	}
	if got := BiologicalWeaponPopKills(0, 20, always); got != 0 || rolls != 0 {
		t.Errorf("0 莢時應回 0 且不擲骰,得到 kills=%d rolls=%d", got, rolls)
	}
}
