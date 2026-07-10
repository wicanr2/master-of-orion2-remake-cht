package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

// setSpecBit 測試用小工具:在 d.Specials 設定第 id 個位元(對照 checkSpecBit 的位元順序)。
func setSpecBit(d *save.ShipDesign, id int) {
	d.Specials[id/8] |= 1 << uint(id%8)
}

func TestCheckSpecBit(t *testing.T) {
	var bf [5]uint8
	bf[0] = 1 << 5 // bit 5
	bf[3] = 1 << 3 // bit 27

	cases := []struct {
		bit  int
		want bool
	}{
		{5, true},
		{27, true},
		{4, false},
		{6, false},
		{39, false},
	}
	for _, c := range cases {
		if got := checkSpecBit(bf[:], c.bit); got != c.want {
			t.Errorf("checkSpecBit(bit=%d) = %v, want %v", c.bit, got, c.want)
		}
	}

	// nil bitfield 一律視為未設(對照 openorion2 checkBitfield 對空指標回 0)。
	if checkSpecBit(nil, 5) {
		t.Errorf("checkSpecBit(nil, 5) = true, want false")
	}
}

func TestShipHasSpecialAndWorking(t *testing.T) {
	var d save.ShipDesign
	setSpecBit(&d, SpecReinforcedHull)

	if !ShipHasSpecial(&d, SpecReinforcedHull) {
		t.Errorf("ShipHasSpecial(ReinforcedHull) = false, want true")
	}
	if ShipHasSpecial(&d, SpecAugmentedEngines) {
		t.Errorf("ShipHasSpecial(AugmentedEngines) = true, want false (未設)")
	}

	// 未追蹤損傷(specDamage=nil)時視為正常運作。
	if !ShipHasWorkingSpecial(&d, SpecReinforcedHull, nil) {
		t.Errorf("ShipHasWorkingSpecial(ReinforcedHull, nil) = false, want true")
	}

	// 裝了但被打壞:hasSpecial 仍 true,working 應為 false。
	var dmg [5]uint8
	dmg[SpecReinforcedHull/8] |= 1 << uint(SpecReinforcedHull%8)
	if !ShipHasSpecial(&d, SpecReinforcedHull) {
		t.Errorf("ShipHasSpecial(ReinforcedHull) after damage flag set on damage-map = false, want true (design 本身未變)")
	}
	if ShipHasWorkingSpecial(&d, SpecReinforcedHull, dmg[:]) {
		t.Errorf("ShipHasWorkingSpecial(ReinforcedHull, damaged) = true, want false")
	}
}

// TestShipBeamAttackFromDesign_FullBonus 手算:
// Size=2 → MaxComputerHP=5(computerHPTable[2]);compDamage=0<5 → 電腦正常。
// Computer=3 → computerBonusTable[3]=75。
// Specials 設 BattleScanner 且未損毀 → +50。
// BeamOffense = 75+50 = 125。
// crewLevel=2 → shipCrewOffenseBonuses[2]=30。raceShipAttack=10。
// 總計 125+30+10 = 165。
func TestShipBeamAttackFromDesign_FullBonus(t *testing.T) {
	var d save.ShipDesign
	d.Size = 2
	d.Computer = 3
	setSpecBit(&d, SpecBattleScanner)

	got := ShipBeamAttackFromDesign(&d, 2, 0, nil, 10)
	want := 165
	if got != want {
		t.Errorf("ShipBeamAttackFromDesign = %d, want %d", got, want)
	}
}

// TestShipBeamAttackFromDesign_ComputerAndScannerDamaged 手算:
// Size=2 → MaxComputerHP=5;compDamage=5(>=5) → 電腦已壞 → 命中加成不算computer bonus。
// BattleScanner 有裝但 specDamage 對應位元也設 → 視為損毀,不加 50。
// crewLevel=0 → +0,raceShipAttack=0 → 總計 0。
func TestShipBeamAttackFromDesign_ComputerAndScannerDamaged(t *testing.T) {
	var d save.ShipDesign
	d.Size = 2
	d.Computer = 3
	setSpecBit(&d, SpecBattleScanner)

	var dmg [5]uint8
	dmg[SpecBattleScanner/8] |= 1 << uint(SpecBattleScanner%8)

	got := ShipBeamAttackFromDesign(&d, 0, 5, dmg[:], 0)
	want := 0
	if got != want {
		t.Errorf("ShipBeamAttackFromDesign = %d, want %d", got, want)
	}
}

// TestShipBeamDefenseFromDesign_AugmentedAndNullifier 手算:
// BaseCombatSpeed=10,Size=2(driveHPTable[2]=10)。AugmentedEngines 未損毀 → ret=10+5=15。
// driveDamage=0,reinforcedHull=false → maxHP=10,hp=10,minHP=2*10/3=6。
// minHP(6)<hp(10):hp-=6→4,maxHP-=6→4,ret=15*4/4=15。transDimensional=false。combatSpeed=15。
// BeamDefense = 15*5=75;InertialNullifier 未損毀 → +100 → 175。crewLevel=3 → +50。總計225。
func TestShipBeamDefenseFromDesign_AugmentedAndNullifier(t *testing.T) {
	var d save.ShipDesign
	d.BaseCombatSpeed = 10
	d.Size = 2
	setSpecBit(&d, SpecAugmentedEngines)
	setSpecBit(&d, SpecInertialNullifier)

	got := ShipBeamDefenseFromDesign(&d, 3, 0, nil, false)
	want := 225
	if got != want {
		t.Errorf("ShipBeamDefenseFromDesign = %d, want %d", got, want)
	}
}

// TestShipBeamDefenseFromDesign_EngineDisabledByDamage 手算:
// BaseCombatSpeed=8,Size=1(driveHPTable[1]=5)。ReinforcedHull 未損毀 → maxDriveHP=5*3=15。
// driveDamage=40 → driveHP=15*(100-40)/100=9。minHP=2*15/3=10。
// minHP(10) < hp(9) 為 false → 引擎判定失能,ret=0。transDimensional=true → ret+=4 → combatSpeed=4。
// BeamDefense=4*5=20(無 nullifier/stabilizer)。crewLevel=0 → +0。總計20。
func TestShipBeamDefenseFromDesign_EngineDisabledByDamage(t *testing.T) {
	var d save.ShipDesign
	d.BaseCombatSpeed = 8
	d.Size = 1
	setSpecBit(&d, SpecReinforcedHull)

	got := ShipBeamDefenseFromDesign(&d, 0, 40, nil, true)
	want := 20
	if got != want {
		t.Errorf("ShipBeamDefenseFromDesign = %d, want %d", got, want)
	}
}

// TestShipCombatStateFromDesign 驗證 BeamDefense 與 TestShipBeamDefenseFromDesign_AugmentedAndNullifier
// 算出的值一致(225),其餘欄位(裝甲/結構/護盾)應為呼叫端傳入值的直接透傳——
// gamedata 尚未移植對應數值表,詳見 ship.go 的 TODO 註解。
func TestShipCombatStateFromDesign(t *testing.T) {
	var d save.ShipDesign
	d.BaseCombatSpeed = 10
	d.Size = 2
	setSpecBit(&d, SpecAugmentedEngines)
	setSpecBit(&d, SpecInertialNullifier)

	state := ShipCombatStateFromDesign(&d, 3, 0, nil, false, 50, 120, 3, true)

	if state.BeamDefense != 225 {
		t.Errorf("state.BeamDefense = %d, want 225", state.BeamDefense)
	}
	if state.ArmorHP != 50 {
		t.Errorf("state.ArmorHP = %d, want 50", state.ArmorHP)
	}
	if state.StructureHP != 120 {
		t.Errorf("state.StructureHP = %d, want 120", state.StructureHP)
	}
	if state.ShieldReduction != 3 {
		t.Errorf("state.ShieldReduction = %d, want 3", state.ShieldReduction)
	}
	if !state.HardShield {
		t.Errorf("state.HardShield = false, want true")
	}
	if state.Destroyed {
		t.Errorf("state.Destroyed = true, want false")
	}
}

// TestShipBeamAttackWithOfficer 沿用 TestShipBeamAttackFromDesign_FullBonus 的固定裝(基準值165),
// 疊加軍官 Weaponry 加成(例如 gamedata.LeaderSkillBonus(int(gamedata.SKILL_WEAPONRY),1,4)=25,
// captain code7 base5*(4+1)=25),對照 GameState::shipBeamOffense 的 officer.skillBonus 疊加。
func TestShipBeamAttackWithOfficer(t *testing.T) {
	var d save.ShipDesign
	d.Size = 2
	d.Computer = 3
	setSpecBit(&d, SpecBattleScanner)

	got := ShipBeamAttackWithOfficer(&d, 2, 0, nil, 10, 25)
	want := 165 + 25
	if got != want {
		t.Errorf("ShipBeamAttackWithOfficer = %d, want %d", got, want)
	}
}

// TestShipBeamAttackWithOfficer_NoOfficer 未指派軍官(或軍官無 Weaponry 技能)時傳 0,
// 行為應與 ShipBeamAttackFromDesign 完全一致(對照原碼 sptr->officer < 0 不加成)。
func TestShipBeamAttackWithOfficer_NoOfficer(t *testing.T) {
	var d save.ShipDesign
	d.Size = 2
	d.Computer = 3
	setSpecBit(&d, SpecBattleScanner)

	got := ShipBeamAttackWithOfficer(&d, 2, 0, nil, 10, 0)
	want := ShipBeamAttackFromDesign(&d, 2, 0, nil, 10)
	if got != want {
		t.Errorf("ShipBeamAttackWithOfficer(bonus=0) = %d, want %d(= ShipBeamAttackFromDesign)", got, want)
	}
}

// TestShipBeamDefenseWithOfficer 沿用 TestShipBeamDefenseFromDesign_AugmentedAndNullifier 的固定裝
// (基準值225),疊加軍官 Helmsman 加成(captain code3 base5*(4+1)=25,tier1 exp4)。
func TestShipBeamDefenseWithOfficer(t *testing.T) {
	var d save.ShipDesign
	d.BaseCombatSpeed = 10
	d.Size = 2
	setSpecBit(&d, SpecAugmentedEngines)
	setSpecBit(&d, SpecInertialNullifier)

	got := ShipBeamDefenseWithOfficer(&d, 3, 0, nil, false, 25)
	want := 225 + 25
	if got != want {
		t.Errorf("ShipBeamDefenseWithOfficer = %d, want %d", got, want)
	}
}
