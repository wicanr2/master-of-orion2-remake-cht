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

func TestResolveVolleyDestroysShip(t *testing.T) {
	atk := BeamAttacker{BeamAttack: 30, DamageMin: 10, DamageMax: 20, RangeLevel: 0}
	ship := &ShipCombatState{StructureHP: 20, ArmorHP: 10, BeamDefense: 20}
	// 每發傷害 13:第1發裝甲10+結構3、第2發結構13(→4)、第3發摧毀。給 5 發應在第3發停。
	rolls := [][2]int{{50, 50}, {50, 50}, {50, 50}, {50, 50}, {50, 50}}
	hits, destroyed := ResolveVolley(atk, ship, rolls)
	if !destroyed {
		t.Fatal("應被摧毀")
	}
	if hits != 3 {
		t.Errorf("命中發數 = %d,預期 3(第3發摧毀後停止)", hits)
	}
	if ship.ArmorHP != 0 || ship.StructureHP != 0 || !ship.Destroyed {
		t.Errorf("摧毀後狀態錯誤:%+v", ship)
	}
}

func TestApplyBeamShotArmorDepletes(t *testing.T) {
	atk := BeamAttacker{BeamAttack: 30, DamageMin: 10, DamageMax: 20, RangeLevel: 0}
	ship := &ShipCombatState{StructureHP: 100, ArmorHP: 10, BeamDefense: 20}
	// 第1發傷害13:裝甲吸10、結構受3、裝甲歸0。
	res := ApplyBeamShot(atk, ship, 50, 50)
	if res.DamageToArmor != 10 || res.DamageToStructure != 3 {
		t.Errorf("裝甲/結構分配錯誤:%+v", res)
	}
	if ship.ArmorHP != 0 || ship.StructureHP != 97 {
		t.Errorf("削減後狀態錯誤:armor=%d structure=%d", ship.ArmorHP, ship.StructureHP)
	}
}

func TestResolveVolleyMissesDoNotDamage(t *testing.T) {
	atk := BeamAttacker{BeamAttack: 30, DamageMin: 10, DamageMax: 20, RangeLevel: 0}
	ship := &ShipCombatState{StructureHP: 100, ArmorHP: 0, BeamDefense: 20}
	// toHitRoll=10 → 10+10=20 < 40 全部未命中。
	rolls := [][2]int{{10, 50}, {10, 50}, {10, 50}}
	hits, destroyed := ResolveVolley(atk, ship, rolls)
	if hits != 0 || destroyed || ship.StructureHP != 100 {
		t.Errorf("未命中不應造成傷害:hits=%d structure=%d", hits, ship.StructureHP)
	}
}
