package gamedata

import "testing"

func TestMissileRackBaseValue(t *testing.T) {
	want := map[int]int{2: 10, 5: 20, 10: 30, 15: 35, 20: 40}
	for ammo, value := range want {
		got, ok := MissileRackBaseValue(ammo)
		if !ok || got != value {
			t.Fatalf("Ammo %d 的彈架值=%d,%v want %d,true", ammo, got, ok, value)
		}
	}
	if _, ok := MissileRackBaseValue(3); ok {
		t.Fatal("非法 Ammo 不應有彈架值")
	}
}

func TestMissileRackNormalizeAndCycle(t *testing.T) {
	if got := NormalizeMissileRackAmmo(0); got != 5 {
		t.Fatalf("舊存檔零值應回 5，got %d", got)
	}
	if got := CycleMissileRackAmmo(5); got != 10 {
		t.Fatalf("5 下一格=%d want 10", got)
	}
	if got := CycleMissileRackAmmo(20); got != 2 {
		t.Fatalf("20 下一格=%d want 2", got)
	}
}
