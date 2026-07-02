package engine

import "testing"

func TestResolveBeamShotHit(t *testing.T) {
	// 近距離(level0,門檻40),netAttack=10,toHitRoll=50 → 命中。
	// 傷害:衰減 level0 不變(10-20),DamageForHit(10,20,50,10,40)=13。無護盾/裝甲 → 全入結構。
	atk := BeamAttacker{BeamAttack: 30, DamageMin: 10, DamageMax: 20, RangeLevel: 0}
	tgt := BeamTarget{BeamDefense: 20} // netAttack=30-20=10
	r := ResolveBeamShot(atk, tgt, 50, 50)
	if !r.Hit {
		t.Fatal("應命中(60>=40)")
	}
	if r.DamageToStructure != 13 || r.DamageToShield != 0 || r.DamageToArmor != 0 {
		t.Errorf("傷害分配錯誤:%+v", r)
	}
}

func TestResolveBeamShotMiss(t *testing.T) {
	atk := BeamAttacker{BeamAttack: 30, DamageMin: 10, DamageMax: 20, RangeLevel: 0}
	tgt := BeamTarget{BeamDefense: 20}     // netAttack=10
	r := ResolveBeamShot(atk, tgt, 10, 50) // 10+10=20 < 40 → miss
	if r.Hit || r.DamageToStructure != 0 {
		t.Errorf("應未命中:%+v", r)
	}
}

func TestResolveBeamShotShield(t *testing.T) {
	// 護盾每擊減 5:傷害 13 → 護盾吸收 5、結構 8。
	atk := BeamAttacker{BeamAttack: 30, DamageMin: 10, DamageMax: 20, RangeLevel: 0}
	tgt := BeamTarget{BeamDefense: 20, ShieldReduction: 5}
	r := ResolveBeamShot(atk, tgt, 50, 50)
	if !r.Hit || r.DamageToShield != 5 || r.DamageToStructure != 8 {
		t.Errorf("護盾解算錯誤:%+v", r)
	}
}
