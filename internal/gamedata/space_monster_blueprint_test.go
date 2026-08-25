package gamedata

import "testing"

func TestEventMonsterBlueprintsMatchOriginalLoaders(t *testing.T) {
	tests := []struct {
		kind                                SpaceMonster
		raw, drive, speed, structure, armor int
		weapons                             []MonsterWeaponMount
	}{
		{MonsterAmoeba, 10, 2, 10, 50, 750, []MonsterWeaponMount{{45, 2, 15, 0, 0}, {23, 5, 15, 0, 10}}},
		{MonsterCrystal, 11, 4, 10, 80, 2500, []MonsterWeaponMount{{42, 1, 15, 2, 0}, {26, 5, 15, 0, 10}}},
		{MonsterDragon, 12, 6, 18, 80, 2500, []MonsterWeaponMount{{41, 20, 15, 4, 0}, {40, 1, 15, 0x4000, 0}}},
		{MonsterEel, 13, 6, 23, 50, 1000, []MonsterWeaponMount{{44, 2, 15, 0, 0}}},
		{MonsterHydra, 14, 2, 6, 80, 1500, []MonsterWeaponMount{{43, 5, 15, 2, 0}}},
	}
	for _, tc := range tests {
		b, ok := MonsterBlueprintFor(tc.kind)
		if !ok || b.RawType != tc.raw || b.Drive != tc.drive || b.BaseCombatSpeed != tc.speed ||
			b.Structure != tc.structure || b.Armor != tc.armor {
			t.Fatalf("怪物 %d 藍圖=%+v", tc.kind, b)
		}
		if len(b.Weapons) != len(tc.weapons) {
			t.Fatalf("怪物 %d 武器槽數=%d，預期 %d", tc.kind, len(b.Weapons), len(tc.weapons))
		}
		for i := range tc.weapons {
			if b.Weapons[i] != tc.weapons[i] {
				t.Errorf("怪物 %d 槽 %d=%+v，預期 %+v", tc.kind, i, b.Weapons[i], tc.weapons[i])
			}
			if _, ok := OrigWeaponByID(b.Weapons[i].WeaponID); !ok {
				t.Errorf("怪物 %d 槽 %d 武器 ID %d 不在原版表", tc.kind, i, b.Weapons[i].WeaponID)
			}
		}
		st, _ := MonsterStatsFor(tc.kind)
		if st.Structure != b.Structure || st.Armor != b.Armor || st.Estimated {
			t.Errorf("怪物 %d stats 未鏡射精確雙血池：%+v", tc.kind, st)
		}
	}
}

func TestMonsterWeaponRawMountDamage(t *testing.T) {
	tests := []struct {
		mount  MonsterWeaponMount
		lo, hi int
	}{
		{MonsterWeaponMount{WeaponID: 42, Count: 1, Mods: MonsterWeaponModHeavyMount}, 60, 120},
		{MonsterWeaponMount{WeaponID: 41, Count: 20, Mods: MonsterWeaponModPointDefense}, 2, 5},
		{MonsterWeaponMount{WeaponID: 40, Count: 1, Mods: 0x4000}, 300, 300},
		{MonsterWeaponMount{WeaponID: 43, Count: 5, Mods: MonsterWeaponModHeavyMount}, 90, 90},
	}
	for _, tc := range tests {
		lo, hi, ok := MonsterWeaponDamageRange(tc.mount)
		if !ok || lo != tc.lo || hi != tc.hi {
			t.Errorf("mount=%+v damage=%d..%d ok=%v，預期 %d..%d", tc.mount, lo, hi, ok, tc.lo, tc.hi)
		}
	}
	if !MonsterWeaponAlwaysHits(40) || !MonsterWeaponAlwaysHits(43) || MonsterWeaponAlwaysHits(42) {
		t.Error("必中怪物武器集合錯誤")
	}
	if w, _ := OrigWeaponByID(23); w.Cat != WeaponCatBomb {
		t.Fatal("Amoeba 第二槽應是炸彈，快速對艦反擊必須排除")
	}
	dragon := MonsterWeaponMount{WeaponID: 40, Count: 1, Mods: MonsterWeaponModOverloadedTorpedo}
	if lo, hi, ok := MonsterWeaponQuickDamageRange(dragon); !ok || lo != 450 || hi != 450 {
		t.Fatalf("Dragon 快速結算 OVR 傷害=%d..%d ok=%v，預期 450..450", lo, hi, ok)
	}
}
