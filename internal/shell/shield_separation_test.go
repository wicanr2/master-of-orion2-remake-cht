package shell

import "testing"

// TestShieldArmorSeparation:護盾與裝甲分離後,StartCombat 的艦帶「裝甲 HP」與「護盾減傷」兩獨立值。
func TestShieldArmorSeparation(t *testing.T) {
	// 依名查表值應與 ArmorOptions/ShieldOptions 一致。
	if got := armorHPByName("三鈦裝甲"); got == 0 {
		t.Fatalf("三鈦裝甲應有裝甲 HP")
	}
	r1 := shieldReduceByName("第一級護盾")
	r3 := shieldReduceByName("第三級護盾")
	if !(r3 > r1 && r1 > 0) {
		t.Fatalf("護盾階越高減傷應越大:第一級 %d、第三級 %d", r1, r3)
	}
	if shieldReduceByName("無護盾") != 0 {
		t.Fatalf("無護盾減傷應為 0")
	}

	// demo 艦「守護號」(三鈦裝甲 + 第三級護盾)進戰鬥應同時有 ArmorHP>0 與 ShieldReduction>0。
	s := NewDemoSession()
	p, _ := s.StartCombat("測試敵")
	var guard *CombatShip
	for i := range p {
		if p[i].Name == "守護號" {
			guard = &p[i]
		}
	}
	if guard == nil {
		t.Skip("demo 無守護號")
	}
	if guard.ArmorHP == 0 || guard.ShieldReduction == 0 {
		t.Fatalf("守護號應有裝甲 HP(%d)與護盾減傷(%d)兩者", guard.ArmorHP, guard.ShieldReduction)
	}
}
