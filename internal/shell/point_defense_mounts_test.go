package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestTacticalPointDefenseUsesSecondSlotAndIgnoresOffMode(t *testing.T) {
	ship := CombatShip{
		WeaponName: "雷射", WeaponMax: 20,
		WeaponMounts: []ShipWeaponMount{
			{Name: "雷射", WorkingCount: 1, Attack: 20},
			{Name: "雷射", WorkingCount: 2, Attack: 30, Mods: []string{string(gamedata.ModPointDefense)}},
		},
		WeaponModes: []TacticalWeaponMode{TacticalWeaponReady, TacticalWeaponOff},
	}
	got := AvailableTacticalPointDefenseMounts(&ship)
	if len(got) != 1 || got[0].Slot != 1 || got[0].Count != 2 || got[0].BeamDamageMax != 30 {
		t.Fatalf("第二槽紅色 PD 仍應是自動迎擊候選：%+v", got)
	}
	if ship.WeaponModes[1] != TacticalWeaponOff {
		t.Fatalf("尋找自動 PD 不得改寫主動開火模式：%v", ship.WeaponModes)
	}
	MarkTacticalPointDefenseMountSpent(&ship, got[0].Slot)
	if again := AvailableTacticalPointDefenseMounts(&ship); len(again) != 0 {
		t.Fatalf("同回合已使用槽不得重複迎擊：%+v", again)
	}
	ResetTacticalPointDefenseSpent(&ship)
	if again := AvailableTacticalPointDefenseMounts(&ship); len(again) != 1 {
		t.Fatalf("回合交界後 PD 槽應可再次迎擊：%+v", again)
	}
	if ship.WeaponModes[1] != TacticalWeaponOff {
		t.Fatalf("重置自動 PD 不得把紅色槽切回可用：%v", ship.WeaponModes)
	}
}

func TestTacticalPointDefenseLegacyFallback(t *testing.T) {
	ship := CombatShip{
		WeaponName: "雷射", WeaponMax: 20,
		Mods: []string{string(gamedata.ModPointDefense)},
	}
	got := AvailableTacticalPointDefenseMounts(&ship)
	if len(got) != 1 || got[0].Slot != -1 || got[0].Count != 1 {
		t.Fatalf("舊單槽艦應保留相容 PD：%+v", got)
	}
	MarkTacticalPointDefenseMountSpent(&ship, -1)
	if len(AvailableTacticalPointDefenseMounts(&ship)) != 0 {
		t.Fatal("舊單槽 PD 同回合不得重複迎擊")
	}
}
