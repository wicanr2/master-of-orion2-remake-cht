package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

var originalComputerApplications = [...]gamedata.Technology{
	gamedata.TECH_NONE,
	gamedata.TECH_ELECTRONIC_COMPUTER,
	gamedata.TECH_OPTRONIC_COMPUTER,
	gamedata.TECH_POSITRONIC_COMPUTER,
	gamedata.TECH_CYBERTRONIC_COMPUTER,
	gamedata.TECH_MOLECULARTRONIC_COMPUTER,
}

// originalBestComputer 對應 sub_5679E 的一般帝國 1..5 掃描。
func originalBestComputer(ps engine.PlayerState) int {
	known := knownTechnologyApplications(ps)
	best := 0
	for index := 1; index < len(originalComputerApplications); index++ {
		if known[originalComputerApplications[index]] {
			best = index
		}
	}
	return best
}

func originalWeaponRawType(component Component) int {
	if component.Name == "無武裝" {
		return 0
	}
	if weapon, ok := gamedata.OrigWeaponByTech(component.UnlockTech); ok {
		return weapon.ID
	}
	return -1
}

// originalWeaponRawMods 只輸出已有 raw consumer 證據的改造。遇到其餘 typed mod 時
// 回 ok=false，呼叫端必須讓該 mount 保持不可供原版國力 producer 消費。
func originalWeaponRawMods(mods []string) (raw uint16, ok bool) {
	ok = true
	for _, mod := range mods {
		switch gamedata.WeaponModCode(mod) {
		case gamedata.ModHeavyMount:
			raw |= 0x0002
		case gamedata.ModPointDefense:
			raw |= 0x0004
		case gamedata.ModContinuousFire:
			raw |= 0x0010
		case gamedata.ModNoRangeDissipation:
			raw |= 0x0020
		case gamedata.ModAutoFire:
			raw |= 0x0080
		case gamedata.ModArmoredMissile:
			raw |= gamedata.MissileRawFlagArmored
		case gamedata.ModFastMissile:
			raw |= gamedata.MissileRawFlagFast
		case gamedata.ModOverloadedTorpedo:
			raw |= 0x4000
		default:
			ok = false
		}
	}
	return raw, ok
}

func originalWeaponMount(component Component, mods []string, arc gamedata.WeaponArc, ammo, count int) ShipWeaponMount {
	rawType := originalWeaponRawType(component)
	rawMods, modsKnown := originalWeaponRawMods(mods)
	if !modsKnown {
		rawType = -1
	}
	return ShipWeaponMount{
		RawType: rawType, Name: component.Name, MaxCount: count, WorkingCount: count,
		Arc: arc, RawMods: rawMods, Mods: append([]string(nil), mods...), Ammo: ammo, Attack: component.Value,
	}
}
