package gamedata

import "testing"

// 三個減傷值出自手冊三段各一句,逐棟釘住。
func TestPlanetaryShieldReductionMatchesTheManual(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{
		{BuildingPlanetaryRadiationShield, 5},
		{BuildingPlanetaryFluxShield, 10},
		{BuildingPlanetaryBarrierShield, 20},
	} {
		got := PlanetaryShieldReduction(map[string]bool{tc.name: true})
		if got != tc.want {
			t.Errorf("%s 應減 %d,得到 %d", tc.name, tc.want, got)
		}
	}
	if got := PlanetaryShieldReduction(nil); got != 0 {
		t.Errorf("沒有護盾應回 0,得到 %d", got)
	}
	if got := PlanetaryShieldReduction(map[string]bool{"太空港": true}); got != 0 {
		t.Errorf("無關建築不該給減傷,得到 %d", got)
	}
}

// 手冊寫的是「取代」不是「疊加」:同時出現時取最強那一面,不加總。
func TestPlanetaryShieldsReplaceRatherThanStack(t *testing.T) {
	all := map[string]bool{
		BuildingPlanetaryRadiationShield: true,
		BuildingPlanetaryFluxShield:      true,
		BuildingPlanetaryBarrierShield:   true,
	}
	if got := PlanetaryShieldReduction(all); got != PlanetaryBarrierShieldReduction {
		t.Errorf("三面同時存在應取最強的 %d,得到 %d(加總會是 %d)",
			PlanetaryBarrierShieldReduction, got,
			PlanetaryRadiationShieldReduction+PlanetaryFluxShieldReduction+PlanetaryBarrierShieldReduction)
	}
	two := map[string]bool{
		BuildingPlanetaryRadiationShield: true,
		BuildingPlanetaryFluxShield:      true,
	}
	if got := PlanetaryShieldReduction(two); got != PlanetaryFluxShieldReduction {
		t.Errorf("輻射+通量應取通量的 %d,得到 %d", PlanetaryFluxShieldReduction, got)
	}
}

// 減傷不會打出負數——負傷害累加起來會變成幫對方補血。
func TestPlanetaryShieldedDamageNeverGoesNegative(t *testing.T) {
	if got := PlanetaryShieldedDamage(3, 20); got != 0 {
		t.Errorf("傷害 3 對上減傷 20 應為 0,得到 %d", got)
	}
	if got := PlanetaryShieldedDamage(25, 20); got != 5 {
		t.Errorf("傷害 25 對上減傷 20 應為 5,得到 %d", got)
	}
	// 正對照:沒有護盾時原封不動,不是「一律歸零」。
	if got := PlanetaryShieldedDamage(25, 0); got != 25 {
		t.Errorf("沒有護盾應原封不動,得到 %d", got)
	}
}

// 維護費的第二個來源:手冊那三段同時給了減傷與維護費,而維護費對得上建築表
// (那張表來自原版執行檔)。三棟全中,代表那段文字可信,減傷值不是孤證。
func TestPlanetaryShieldMaintenanceMatchesTheBuildingTable(t *testing.T) {
	want := map[string]int{
		BuildingPlanetaryRadiationShield: 1,
		BuildingPlanetaryFluxShield:      3,
		BuildingPlanetaryBarrierShield:   5,
	}
	found := 0
	for _, b := range Buildings {
		if w, ok := want[b.NameZH]; ok {
			found++
			if b.MaintenanceBC != w {
				t.Errorf("%s 手冊維護費 %d BC,建築表 %d", b.NameZH, w, b.MaintenanceBC)
			}
		}
	}
	if found != len(want) {
		t.Fatalf("建築表裡應找到 %d 棟護盾,找到 %d 棟", len(want), found)
	}
}
