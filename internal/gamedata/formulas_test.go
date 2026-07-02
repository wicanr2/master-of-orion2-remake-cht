package gamedata

import "testing"

func TestFormulas(t *testing.T) {
	// PlanetBaseProduction:mineralProductionTable = {1,2,3,5,8}
	if got := PlanetBaseProduction(int(ULTRA_RICH)); got != 8 {
		t.Errorf("PlanetBaseProduction(ULTRA_RICH) = %d,預期 8", got)
	}
	if got := PlanetBaseProduction(int(POOR)); got != 2 {
		t.Errorf("PlanetBaseProduction(POOR) = %d,預期 2", got)
	}

	// 電腦/引擎 HP
	if got := MaxComputerHP(2); got != 5 {
		t.Errorf("MaxComputerHP(2) = %d,預期 5", got)
	}
	if got := ComputerHP(5, 3); got != 17 { // 20 - 3
		t.Errorf("ComputerHP(5,3) = %d,預期 17", got)
	}
	if got := MaxDriveHP(3, false); got != 15 {
		t.Errorf("MaxDriveHP(3,false) = %d,預期 15", got)
	}
	if got := MaxDriveHP(3, true); got != 45 { // ×3 強化船殼
		t.Errorf("MaxDriveHP(3,true) = %d,預期 45", got)
	}
	if got := DriveHP(3, 50, false); got != 7 { // 15*50/100
		t.Errorf("DriveHP(3,50,false) = %d,預期 7", got)
	}
	if got := DriveHP(3, 100, false); got != 0 {
		t.Errorf("DriveHP(3,100) = %d,預期 0", got)
	}

	// CombatSpeed:base=4 size=0(maxHP=2)無損傷 → 4;transdim → 8
	if got := CombatSpeed(4, 0, 0, false, false, false); got != 4 {
		t.Errorf("CombatSpeed 無損傷 = %d,預期 4", got)
	}
	if got := CombatSpeed(4, 0, 0, false, false, true); got != 8 {
		t.Errorf("CombatSpeed transdim = %d,預期 8", got)
	}
	// driveDamage=50 → hp(1) 不大於 minHP(1) → 0
	if got := CombatSpeed(4, 0, 50, false, false, false); got != 0 {
		t.Errorf("CombatSpeed 重損 = %d,預期 0", got)
	}

	// 命中/閃避
	if got := ComputerBonus(4); got != 100 {
		t.Errorf("ComputerBonus(4) = %d,預期 100", got)
	}
	if got := BeamOffense(4, true, false); got != 100 {
		t.Errorf("BeamOffense 電腦正常 = %d,預期 100", got)
	}
	if got := BeamOffense(4, false, true); got != 50 { // 電腦壞,只剩掃描器
		t.Errorf("BeamOffense 電腦壞+掃描器 = %d,預期 50", got)
	}
	if got := BeamDefense(4, false, false); got != 20 { // 4*5
		t.Errorf("BeamDefense = %d,預期 20", got)
	}
	if got := BeamDefense(4, true, true); got != 170 { // 20+100+50
		t.Errorf("BeamDefense 慣性裝備 = %d,預期 170", got)
	}

	// 雇用費:10*skillValue*(expLevel+1)+modifier,下限 0
	if got := LeaderHireCost(10, 2, 0); got != 300 {
		t.Errorf("LeaderHireCost(10,2,0) = %d,預期 300", got)
	}
	if got := LeaderHireCost(10, 2, -9999); got != 0 {
		t.Errorf("LeaderHireCost 下限 = %d,預期 0", got)
	}
}

func TestShipCrewBonuses(t *testing.T) {
	// gamestate.cpp:162-167:{0,15,30,50,75}(新兵→超級精銳)
	want := []int{0, 15, 30, 50, 75}
	for lvl, w := range want {
		if got := ShipCrewOffenseBonus(lvl); got != w {
			t.Errorf("ShipCrewOffenseBonus(%d) = %d,預期 %d", lvl, got, w)
		}
		if got := ShipCrewDefenseBonus(lvl); got != w {
			t.Errorf("ShipCrewDefenseBonus(%d) = %d,預期 %d", lvl, got, w)
		}
	}
	// 越界回 0
	if ShipCrewOffenseBonus(-1) != 0 || ShipCrewOffenseBonus(5) != 0 {
		t.Error("越界 crewLevel 應回 0")
	}
}
