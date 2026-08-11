package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestResolvePointDefenseFighterShotUsesFighterBeamDefense(t *testing.T) {
	mods := []gamedata.WeaponModCode{gamedata.ModPointDefense}

	miss := ResolvePointDefenseFighterShot(PointDefenseFighterShot{
		BeamWeaponName:     "雷射",
		BeamAttack:         100,
		BeamDamageMax:      8,
		BeamRoll:           1,
		BeamMods:           mods,
		FighterBeamDefense: 1000,
	})
	if !miss.Fired || miss.Hit {
		t.Fatalf("PD 應開火但應被戰機防禦擋下，結果 %+v", miss)
	}

	hit := ResolvePointDefenseFighterShot(PointDefenseFighterShot{
		BeamWeaponName:     "雷射",
		BeamAttack:         100,
		BeamDamageMax:      8,
		BeamRoll:           1,
		BeamMods:           mods,
		FighterBeamDefense: gamedata.CombatFighterBeamDefense(14, 0, 0, 0),
	})
	if !hit.Fired || !hit.Hit || hit.DamageToFighter <= 0 {
		t.Fatalf("速度 14 的戰機應被 PD 命中並受傷，結果 %+v", hit)
	}
	if hit.FighterBeamDefense != 70 {
		t.Fatalf("戰機防禦應為 5*14=70，實得 %d", hit.FighterBeamDefense)
	}

	if noPD := ResolvePointDefenseFighterShot(PointDefenseFighterShot{
		BeamWeaponName: "雷射",
		BeamAttack:     100,
		BeamDamageMax:  8,
		BeamRoll:       1,
	}); noPD.Fired {
		t.Fatalf("沒有 PD 改造時不應自動開火，結果 %+v", noPD)
	}
}
