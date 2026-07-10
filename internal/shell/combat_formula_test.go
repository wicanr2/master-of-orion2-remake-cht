package shell

import "testing"

// TestResolveShot 驗證真戰鬥公式接線:近距高攻必中且傷害穿甲、超遠距低攻低 roll 不中、護盾減傷。
func TestResolveShot(t *testing.T) {
	// 近距(1 格)、netAttack 高、roll 高 → 命中,傷害先耗裝甲再傷結構。
	r := ResolveShot(80 /*net*/, 10, 20 /*dmg*/, 1 /*range*/, 0 /*shield*/, 5 /*armorHP*/, 90 /*roll*/, false, false)
	if !r.Hit {
		t.Fatalf("近距高攻高 roll 應命中")
	}
	if r.RemainingArmorHP > 5 {
		t.Fatalf("命中後裝甲不應增加")
	}

	// 遠距(距離大)+ 低 netAttack + 低 roll → 命中門檻高,應未命中。
	miss := ResolveShot(0 /*net*/, 10, 20, 30 /*遠*/, 0, 10, 5 /*低 roll*/, false, false)
	if miss.Hit {
		t.Fatalf("遠距低攻低 roll 不應命中(門檻高)")
	}
	if miss.RemainingArmorHP != 10 {
		t.Fatalf("未命中裝甲應不變")
	}

	// 護盾減傷:同一發,護盾高者打到結構的傷害應較少或相等。
	hi := ResolveShot(90, 20, 20, 1, 0 /*無盾*/, 0 /*無甲*/, 90, false, false)
	lo := ResolveShot(90, 20, 20, 1, 15 /*高盾*/, 0, 90, false, false)
	if lo.DamageToStructure > hi.DamageToStructure {
		t.Fatalf("護盾應減傷:無盾 %d 應 >= 有盾 %d", hi.DamageToStructure, lo.DamageToStructure)
	}
}
