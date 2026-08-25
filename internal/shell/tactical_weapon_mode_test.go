package shell

import "testing"

func TestNewTacticalWeaponModesMatchesMountShapeAndLegacyFallback(t *testing.T) {
	if got := NewTacticalWeaponModes(nil); len(got) != 1 || got[0] != TacticalWeaponReady {
		t.Fatalf("舊單槽艦應取得一個 ready 模式：%v", got)
	}
	mounts := []ShipWeaponMount{{Name: "雷射"}, {Name: "核飛彈"}, {Name: "脈衝星"}}
	got := NewTacticalWeaponModes(mounts)
	if len(got) != len(mounts) {
		t.Fatalf("模式數應與 weapon mounts 相同：%d/%d", len(got), len(mounts))
	}
	for i, mode := range got {
		if mode != TacticalWeaponReady {
			t.Fatalf("新戰鬥第 %d 槽應為 ready，得到 %d", i, mode)
		}
	}
}
