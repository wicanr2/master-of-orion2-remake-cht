package gamedata

import "testing"

func TestOriginalNPCPowerModifierUsesExecutableWords(t *testing.T) {
	got, ok := OriginalNPCPowerModifier(100, 0x0002|0x0004|0x0020|0x2000)
	if !ok || got != 475 { // 100 +100 -50 +25 +300 = 475%。
		t.Fatalf("改造倍率=%d,%v，預期 475,true", got, ok)
	}
	if got, ok := OriginalNPCPowerModifier(64000, 0x2000); !ok || got != 64000 {
		t.Fatalf("word 上限=%d,%v，預期 64000,true", got, ok)
	}
}

func TestOriginalNPCShipPowerBeamObserverAndDamage(t *testing.T) {
	in := OriginalNPCShipPowerInput{
		Mounts:     []OriginalNPCPowerMount{{WeaponID: 3, WorkingCount: 10}},
		BeamAttack: 50, ObserverDefense: 20, OwnerBestComputer: 3, DesignComputer: 3,
		RemainingDurability: 100,
	}
	// Laser max4-20 defense=0，十門；命中 80%，耐久 factor100。
	if got, ok := OriginalNPCShipPower(in); !ok || got != 0 {
		t.Fatalf("高防禦觀察者國力=%d,%v，預期 0,true", got, ok)
	}
	in.ObserverDefense = 1
	// (4-1)*10 * 99% = 29（整數截斷）。
	if got, ok := OriginalNPCShipPower(in); !ok || got != 29 {
		t.Fatalf("方向光束國力=%d,%v，預期 29,true", got, ok)
	}
	in.RemainingDurability = 50 // factor 55
	if got, ok := OriginalNPCShipPower(in); !ok || got != 15 {
		t.Fatalf("受損方向國力=%d,%v，預期 15,true", got, ok)
	}
}

func TestOriginalNPCShipPowerEightSlotsAndFighterMapping(t *testing.T) {
	mounts := make([]OriginalNPCPowerMount, 9)
	for i := range mounts {
		mounts[i] = OriginalNPCPowerMount{WeaponID: 14, WorkingCount: 1, Ammo: 5}
	}
	in := OriginalNPCShipPowerInput{Mounts: mounts, BeamAttack: 50, OwnerBestComputer: 3,
		DesignComputer: 3, RemainingDurability: 100}
	if got, ok := OriginalNPCShipPower(in); !ok || got != 64 {
		t.Fatalf("只應累加前八槽：%d,%v", got, ok)
	}
	in.Mounts = []OriginalNPCPowerMount{{WeaponID: 31, WorkingCount: 2}}
	in.BestBeamWeaponID = 3
	if got, ok := OriginalNPCShipPower(in); !ok || got != 24 {
		t.Fatalf("戰機艙應換成六門最佳光束：%d,%v", got, ok)
	}
}

func TestOriginalNPCDurabilityFactor(t *testing.T) {
	for _, tc := range []struct{ in, want int }{{50, 55}, {100, 100}, {200, 125}} {
		if got := OriginalNPCDurabilityFactor(tc.in); got != tc.want {
			t.Errorf("factor(%d)=%d，預期 %d", tc.in, got, tc.want)
		}
	}
}
