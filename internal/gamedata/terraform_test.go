package gamedata

import "testing"

func TestTerraformConstants(t *testing.T) {
	// GAME_MANUAL.pdf 直接給出的常數,防止之後被誤改。
	if TerraformSoilEnrichmentFoodBonusPerFarmer != 1 {
		t.Errorf("TerraformSoilEnrichmentFoodBonusPerFarmer = %d,預期 1", TerraformSoilEnrichmentFoodBonusPerFarmer)
	}
	if TerraformToxicNextClimate != BARREN {
		t.Errorf("TerraformToxicNextClimate = %v,預期 BARREN", TerraformToxicNextClimate)
	}
	if GaiaTransformationSourceClimate != TERRAN {
		t.Errorf("GaiaTransformationSourceClimate = %v,預期 TERRAN", GaiaTransformationSourceClimate)
	}
	if GaiaTransformationResultClimate != GAIA {
		t.Errorf("GaiaTransformationResultClimate = %v,預期 GAIA", GaiaTransformationResultClimate)
	}
}

func TestTerraformSoilEnrichmentWorks(t *testing.T) {
	// 手冊:Barren/Radiated/Toxic 無效,其餘氣候有效。
	cases := []struct {
		climate PlanetClimate
		want    bool
	}{
		{BARREN, false},
		{RADIATED, false},
		{TOXIC, false},
		{DESERT, true},
		{TUNDRA, true},
		{OCEAN, true},
		{SWAMP, true},
		{ARID, true},
		{TERRAN, true},
		{GAIA, true},
	}
	for _, c := range cases {
		if got := TerraformSoilEnrichmentWorks(c.climate); got != c.want {
			t.Errorf("TerraformSoilEnrichmentWorks(%v) = %v,預期 %v", c.climate, got, c.want)
		}
	}
}

func TestTerraformNextClimateOptions(t *testing.T) {
	cases := []struct {
		climate PlanetClimate
		want    []PlanetClimate
	}{
		{BARREN, []PlanetClimate{DESERT, TUNDRA}}, // 手冊:兩個候選,未給選擇條件
		{DESERT, []PlanetClimate{ARID}},
		{TUNDRA, []PlanetClimate{SWAMP}},
		{OCEAN, []PlanetClimate{TERRAN}},
		{ARID, []PlanetClimate{TERRAN}},
		{SWAMP, []PlanetClimate{TERRAN}},
		{TERRAN, nil},   // 手冊未提及 Terraforming 可再推進 Terran(那是 Gaia Transformation)
		{GAIA, nil},     // 已是鏈中終點
		{TOXIC, nil},    // 一般 Terraforming 不能建在 Toxic(見 TerraformToxicNextClimate)
		{RADIATED, nil}, // 手冊未提供 Radiated 的地形改造規則
	}
	for _, c := range cases {
		got := TerraformNextClimateOptions(c.climate)
		if len(got) != len(c.want) {
			t.Errorf("TerraformNextClimateOptions(%v) = %v,預期 %v", c.climate, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("TerraformNextClimateOptions(%v) = %v,預期 %v", c.climate, got, c.want)
				break
			}
		}
	}
}

func TestGaiaTransformationCanApply(t *testing.T) {
	cases := []struct {
		climate PlanetClimate
		want    bool
	}{
		{TERRAN, true},
		{GAIA, false},
		{ARID, false},
		{SWAMP, false},
		{BARREN, false},
	}
	for _, c := range cases {
		if got := GaiaTransformationCanApply(c.climate); got != c.want {
			t.Errorf("GaiaTransformationCanApply(%v) = %v,預期 %v", c.climate, got, c.want)
		}
	}
}

func TestTerraformPopMaxAfterClimateChange(t *testing.T) {
	cases := []struct {
		name                   string
		popMax                 int
		oldClimate, newClimate PlanetClimate
		want                   int
	}{
		// Barren(25%) → Desert(25%):同一係數,PopMax 不動。
		{"Barren→Desert 同係數不動", 20, BARREN, DESERT, 20},
		// Swamp(40%) → Terran(80%):係數翻倍,PopMax 翻倍。
		{"Swamp→Terran 係數翻倍", 20, SWAMP, TERRAN, 40},
		// Terran(80%) → Gaia(100%):係數 80→100,PopMax 等比例放大。
		{"Terran→Gaia 係數 80→100", 40, TERRAN, GAIA, 50},
		// Arid(60%) → Terran(80%):20*80/60=26(向下取整)。
		{"Arid→Terran 向下取整", 20, ARID, TERRAN, 26},
		// oldClimate 超出合法範圍(係數 0):不換算,原樣回傳,避免除以零。
		{"oldClimate 超出範圍不換算", 20, PlanetClimate(-1), TERRAN, 20},
	}
	for _, c := range cases {
		if got := TerraformPopMaxAfterClimateChange(c.popMax, c.oldClimate, c.newClimate); got != c.want {
			t.Errorf("%s: TerraformPopMaxAfterClimateChange(%d, %v, %v) = %d,預期 %d",
				c.name, c.popMax, c.oldClimate, c.newClimate, got, c.want)
		}
	}
}

func TestTerraformClimatePopFactorPercent(t *testing.T) {
	// MANUAL_150.html: pop_climate = 25 25 25 25 25 25 40 60 80 100
	// (順序依 enums.go 的 PlanetClimate,並與 openorion2/src/gamestate.cpp 的
	// climatePopFactors 交叉驗證,見 terraform.go 檔頭說明)。
	cases := []struct {
		climate PlanetClimate
		want    int
	}{
		{TOXIC, 25},
		{RADIATED, 25},
		{BARREN, 25},
		{DESERT, 25},
		{TUNDRA, 25},
		{OCEAN, 25},
		{SWAMP, 40},
		{ARID, 60},
		{TERRAN, 80},
		{GAIA, 100},
		{-1, 0}, // 超出範圍
		{10, 0}, // 超出範圍
	}
	for _, c := range cases {
		if got := TerraformClimatePopFactorPercent(c.climate); got != c.want {
			t.Errorf("TerraformClimatePopFactorPercent(%v) = %d,預期 %d", c.climate, got, c.want)
		}
	}
}
