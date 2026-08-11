package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestResolveMissileShotWithModsECCM(t *testing.T) {
	plain := ResolveMissileShotWithMods(false, 0, 1, 67, 0, false, 34,
		10, 0, 0, false, MissileDefenses{}, "核飛彈", nil)
	if plain.Hit {
		t.Fatal("沒有 ECCM 時 34 應被 67% 干擾機率擋下")
	}
	withECCM := ResolveMissileShotWithMods(false, 0, 1, 67, 0, false, 34,
		10, 0, 0, false, MissileDefenses{}, "核飛彈", []gamedata.WeaponModCode{gamedata.ModMissileECCM})
	if !withECCM.Hit {
		t.Fatal("ECCM 應讓 67% 干擾機率減半，34 應命中")
	}
}

func TestResolveMissileShotWithModsEMGBypassesArmorAfterShield(t *testing.T) {
	plain := ResolveMissileShotWithMods(false, 0, 1, 0, 0, false, 1,
		20, 0, 10, false, MissileDefenses{}, "核飛彈", nil)
	emg := ResolveMissileShotWithMods(false, 0, 1, 0, 0, false, 1,
		20, 0, 10, false, MissileDefenses{}, "核飛彈", []gamedata.WeaponModCode{gamedata.ModEmissionsGuidance})
	if plain.DamageToStructure != 10 || plain.RemainingArmorHP != 0 {
		t.Fatalf("普通飛彈傷害=(%d, armor=%d), want structure 10 armor 0", plain.DamageToStructure, plain.RemainingArmorHP)
	}
	if emg.DamageToStructure != 20 || emg.RemainingArmorHP != 10 {
		t.Fatalf("EMG 傷害=(%d, armor=%d), want direct structure 20 armor 10", emg.DamageToStructure, emg.RemainingArmorHP)
	}
}

func TestResolveMissileShotWithModsTorpedoDamage(t *testing.T) {
	mods := []gamedata.WeaponModCode{
		gamedata.ModOverloadedTorpedo,
		gamedata.ModEnveloping,
	}
	shot := ResolveMissileShotWithMods(false, 0, 1, 0, 0, false, 1,
		40, 0, 0, false, MissileDefenses{}, "質子魚雷", mods)
	if !shot.Hit || shot.DamageToStructure != 240 {
		t.Fatalf("OVR+ENV 魚雷命中=%v 傷害=%d, want hit/240", shot.Hit, shot.DamageToStructure)
	}
}

func TestResolveMissileShotWithModsMIRVRollsEachWarhead(t *testing.T) {
	def := MissileDefenses{JamRolls: []int{1, 51, 1, 51}}
	shot := ResolveMissileShotWithMods(false, 0, 1, 50, 0, false, 1,
		10, 0, 0, false, def, "核飛彈", []gamedata.WeaponModCode{gamedata.ModMIRV})
	if !shot.Hit || shot.DamageToStructure != 20 {
		t.Fatalf("MIRV 逐彈頭結果命中=%v 傷害=%d, want hit/20", shot.Hit, shot.DamageToStructure)
	}
}

func TestResolvePointDefenseInterceptsStandardMissile(t *testing.T) {
	beamMods := []gamedata.WeaponModCode{gamedata.ModPointDefense}
	if !PointDefenseCanEngage("雷射", "核飛彈", beamMods) {
		t.Fatal("PD 光束應能對標準飛彈開火")
	}
	res := ResolvePointDefenseIntercept(PointDefenseShot{
		BeamWeaponName:    "雷射",
		BeamAttack:        100,
		BeamDamageMax:     20,
		BeamRangeSquares:  0,
		BeamRoll:          1,
		BeamMods:          beamMods,
		MissileWeaponName: "核飛彈",
		MissileFTLLevel:   gamedata.MissileFTLNuclear,
	})
	if !res.Fired || !res.Hit || res.DestroyedWarheads != 1 {
		t.Fatalf("PD 攔截結果=%+v, want fired/hit/1", res)
	}
	if res.MissileBeamDefense != 50 || res.InterceptionDurability != 4 {
		t.Fatalf("PD 讀到的飛彈數值=%+v, want BD=50,Dcv=4", res)
	}
}

func TestResolvePointDefenseUsesFastAndArmoredMissileFlags(t *testing.T) {
	beamMods := []gamedata.WeaponModCode{gamedata.ModPointDefense}
	plain := ResolvePointDefenseIntercept(PointDefenseShot{
		BeamWeaponName:    "雷射",
		BeamAttack:        83,
		BeamDamageMax:     20,
		BeamRangeSquares:  0,
		BeamRoll:          1,
		BeamMods:          beamMods,
		MissileWeaponName: "核飛彈",
		MissileFTLLevel:   gamedata.MissileFTLNuclear,
	})
	fastArm := ResolvePointDefenseIntercept(PointDefenseShot{
		BeamWeaponName:    "雷射",
		BeamAttack:        83,
		BeamDamageMax:     20,
		BeamRangeSquares:  0,
		BeamRoll:          1,
		BeamMods:          beamMods,
		MissileWeaponName: "核飛彈",
		MissileFTLLevel:   gamedata.MissileFTLNuclear,
		MissileMods:       []gamedata.WeaponModCode{gamedata.ModArmoredMissile, gamedata.ModFastMissile},
	})
	if !plain.Hit || plain.DestroyedWarheads != 1 {
		t.Fatalf("普通核飛彈應被 PD 擊落,結果=%+v", plain)
	}
	if fastArm.Hit {
		t.Fatalf("FST 應提高 Beam Defense 使這發未命中,結果=%+v", fastArm)
	}
	if fastArm.MissileBeamDefense != 70 || fastArm.InterceptionDurability != 8 {
		t.Fatalf("ARM/FST 數值=%+v, want BD=70,Dcv=8", fastArm)
	}
	if PointDefenseCanEngage("雷射", "質子魚雷", beamMods) {
		t.Fatal("魚雷不應進入 PD 飛彈攔截路徑")
	}
}

func TestResolvePointDefenseCarriesInterceptionRemainderAcrossShots(t *testing.T) {
	beamMods := []gamedata.WeaponModCode{gamedata.ModPointDefense}
	first := ResolvePointDefenseIntercept(PointDefenseShot{
		BeamWeaponName:            "雷射",
		BeamAttack:                100,
		BeamDamageMax:             2,
		BeamRangeSquares:          0,
		BeamRoll:                  1,
		BeamMods:                  beamMods,
		MissileWeaponName:         "核飛彈",
		MissileFTLLevel:           gamedata.MissileFTLNuclear,
		CarriedInterceptionDamage: 3,
	})
	if !first.Fired || !first.Hit || first.DestroyedWarheads != 1 {
		t.Fatalf("帶餘數的 ARM/FST 攔截應完成一枚彈頭,結果=%+v", first)
	}
	if first.RemainingInterceptionDamage != 0 {
		t.Fatalf("攔截餘數應消費完,得到 %d", first.RemainingInterceptionDamage)
	}

	miss := ResolvePointDefenseIntercept(PointDefenseShot{
		BeamWeaponName:            "雷射",
		BeamAttack:                -100,
		BeamDamageMax:             2,
		BeamRangeSquares:          0,
		BeamRoll:                  95,
		BeamMods:                  beamMods,
		MissileWeaponName:         "核飛彈",
		MissileFTLLevel:           gamedata.MissileFTLNuclear,
		CarriedInterceptionDamage: 2,
	})
	if !miss.Fired || miss.RemainingInterceptionDamage != 2 {
		t.Fatalf("攔截未命中時應保留餘數,結果=%+v", miss)
	}
}

func TestResolveMissileShotWithModsConsumesInterceptedWarheads(t *testing.T) {
	def := MissileDefenses{
		InterceptedWarheads: 1,
		JamRolls:            []int{1, 1, 1, 1},
	}
	shot := ResolveMissileShotWithMods(false, 0, 1, 0, 0, false, 1,
		10, 0, 0, false, def, "核飛彈", []gamedata.WeaponModCode{gamedata.ModMIRV})
	if !shot.Hit || shot.DamageToStructure != 30 {
		t.Fatalf("攔截一枚 MIRV 後命中=%v 傷害=%d, want hit/30", shot.Hit, shot.DamageToStructure)
	}
}

func TestWeaponModOptionsForWeapon(t *testing.T) {
	beam := WeaponModOptionsForWeapon("雷射")
	if len(beam) != len(WeaponModOptions) {
		t.Fatalf("光束改造數=%d, want %d", len(beam), len(WeaponModOptions))
	}
	missile := WeaponModOptionsForWeapon("核飛彈")
	if len(missile) != 5 || !WeaponModAppliesToWeapon("核飛彈", gamedata.ModMissileECCM) {
		t.Fatalf("核飛彈改造清單=%v, ECCM 應適用", missile)
	}
	if !WeaponModAppliesToWeapon("核飛彈", gamedata.ModArmoredMissile) ||
		!WeaponModAppliesToWeapon("核飛彈", gamedata.ModFastMissile) {
		t.Fatalf("核飛彈應提供 ARM/FST,清單=%v", missile)
	}
	if WeaponModAppliesToWeapon("核飛彈", gamedata.ModOverloadedTorpedo) {
		t.Fatal("OVR 不應適用於核飛彈")
	}
	torpedo := WeaponModOptionsForWeapon("質子魚雷")
	if len(torpedo) != 8 || !WeaponModAppliesToWeapon("質子魚雷", gamedata.ModOverloadedTorpedo) {
		t.Fatalf("質子魚雷改造清單=%v, OVR 應適用", torpedo)
	}
	if !WeaponModAppliesToWeapon("質子魚雷", gamedata.ModNoRangeDissipation) {
		t.Fatal("魚雷應提供 NR")
	}
	for _, name := range []string{"反物質魚雷", "電漿魚雷"} {
		if !WeaponIsTorpedo(name) || weaponKindByName(name) != WeaponKindMissile {
			t.Fatalf("%s 應走魚雷/飛彈解算分類", name)
		}
	}
}
