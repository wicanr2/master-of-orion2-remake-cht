package shell

import "testing"

func TestWeakestShieldFacingChoosesLowestCapacity(t *testing.T) {
	target := CombatShip{
		ShieldReduction:          3,
		ShieldFacingsInitialized: true,
		ShieldFacingHP:           [4]int{12, 4, 9, 4},
	}
	if got := target.WeakestShieldFacing(); got != 1 {
		t.Fatalf("最弱分面應為索引 1(平手取低索引),得到 %d", got)
	}
}

func TestFighterDamageUsesWeakestShieldFacing(t *testing.T) {
	target := CombatShip{
		ShieldReduction:          2,
		ShieldFacingsInitialized: true,
		ShieldFacingHP:           [4]int{10, 1, 10, 10},
	}
	structure, shield := FighterDamageAtWeakestShield(&target, 4, 3)
	// 每架原始 3，既有單發減傷 2；最弱面只有 1 容量，所以只吸收 1，
	// 剩下 11 進入後續裝甲／結構流程。
	if structure != 11 || shield != 1 {
		t.Fatalf("最弱分面分流 = structure %d/shield %d, want 11/1", structure, shield)
	}
	if target.ShieldFacingHP[1] != 0 {
		t.Fatalf("最弱分面應被扣到 0,得到 %d", target.ShieldFacingHP[1])
	}
}
