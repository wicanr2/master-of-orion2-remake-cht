package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// mkCloakShip 造一艘帶指定匿蹤系統、開場即隱形的戰鬥艦。
func mkCloakShip(k CloakKind) CombatShip {
	return CombatShip{Name: "測試艦", HP: 100, MaxHP: 100, Defense: 20,
		CloakKind: k, Cloaked: k != CloakNone}
}

// TestCloakBeamDefenseBonus 隱形裝置未開火時 +80 光束防禦(手冊逐字)。
func TestCloakBeamDefenseBonus(t *testing.T) {
	plain := CombatShip{Defense: 20}
	cloaked := mkCloakShip(CloakDevice)
	if got := TacticalEffectiveDefenseAtRound(plain, 1); got != 20 {
		t.Errorf("沒裝匿蹤的防禦=%d,期望 20", got)
	}
	if got, want := TacticalEffectiveDefenseAtRound(cloaked, 1), 20+gamedata.ShipCloakingDeviceBeamDefense; got != want {
		t.Errorf("隱形中的防禦=%d,期望 %d", got, want)
	}
}

// TestCloakLostOnFire 手冊:「When a cloaked ship does attack, it loses these bonuses.」
// ——是開火**當下**失效,不是下一回合。
func TestCloakLostOnFire(t *testing.T) {
	sh := mkCloakShip(CloakDevice)
	CloakOnFire(&sh)
	if sh.Cloaked {
		t.Fatal("開火之後不該還是隱形")
	}
	if got := TacticalEffectiveDefenseAtRound(sh, 1); got != 20 {
		t.Errorf("開火之後防禦=%d,期望回到 20", got)
	}
}

// TestCloakRecloakNeedsFullQuietRound 手冊:「it must remain uncloaked until it spends
// one full turn without firing; then it can recloak.」
//
// 「一整回合沒開火」與行動次數家族的 Fired 是同一個訊號,所以這裡直接走 TacticalAdvanceCharge
// ——那正是實際戰鬥迴圈會呼叫的東西,而不是另外開一條測試專用路徑。
func TestCloakRecloakNeedsFullQuietRound(t *testing.T) {
	ships := []CombatShip{mkCloakShip(CloakDevice)}
	// 第 1 回合開火 → 失去隱形,而且回合結束時不該恢復。
	CloakOnFire(&ships[0])
	ships[0].Fired = true
	TacticalAdvanceCharge(ships)
	if ships[0].Cloaked {
		t.Fatal("開火那一回合結束時不該重新隱形")
	}
	// 第 2 回合完全沒開火 → 回合結束時恢復隱形。
	TacticalAdvanceCharge(ships)
	if !ships[0].Cloaked {
		t.Error("停火一整回合之後應該重新隱形")
	}
}

// TestPhasingCloakUntargetableThenDegrades 相位匿蹤前 10 回合完全不能被選為目標,
// 第 11 回合起降級成隱形裝置(手冊:「functions just like a Cloaking Device until the
// end of that combat」)。
func TestPhasingCloakUntargetableThenDegrades(t *testing.T) {
	sh := mkCloakShip(CloakPhasing)
	for r := 1; r <= gamedata.ShipPhasingCloakCombatRounds; r++ {
		if !CloakUntargetable(sh, r) {
			t.Fatalf("第 %d 回合相位匿蹤應該完全鎖定不了", r)
		}
	}
	r := gamedata.ShipPhasingCloakCombatRounds + 1
	if CloakUntargetable(sh, r) {
		t.Errorf("第 %d 回合相位匿蹤應該已經降級,可以被鎖定", r)
	}
	if got, want := TacticalEffectiveDefenseAtRound(sh, r), 20+gamedata.ShipCloakingDeviceBeamDefense; got != want {
		t.Errorf("降級後的防禦=%d,期望 %d(等同隱形裝置)", got, want)
	}
	// 降級之後開火,一樣失去加成。
	CloakOnFire(&sh)
	if CloakMissileMissChance(sh, r) != 0 {
		t.Error("開火之後不該還有飛彈未命中加成")
	}
}

// TestCloakMissileMissChance 隱形時飛彈 50% 未命中(手冊逐字),而且**真的接到解算裡**
// ——擲骰 ≤50 應該未命中,>50 才照常判定。
func TestCloakMissileMissChance(t *testing.T) {
	const wmax = 40
	base := func(def MissileDefenses) ShotResult {
		return ResolveMissileShot(false, 0, 1, 0, 0, false, 1, wmax, 0, 0, false, def)
	}
	if r := base(MissileDefenses{}); !r.Hit {
		t.Fatal("測試前提不成立:無防禦時應命中")
	}
	miss := base(MissileDefenses{CloakMissChance: gamedata.ShipCloakingDeviceMissileMissChance, CloakRoll: 50})
	if miss.Hit {
		t.Error("擲出 50(≤50%)應該未命中")
	}
	hit := base(MissileDefenses{CloakMissChance: gamedata.ShipCloakingDeviceMissileMissChance, CloakRoll: 51})
	if !hit.Hit {
		t.Error("擲出 51(>50%)應該照常命中")
	}
}

// TestRangemasterShortensHitRangeOnly 測距瞄準器把命中用的距離縮成 1/3,
// **但傷害衰減照真實距離算**(手冊特地寫了第二句排除它)。
//
// 這條測試的價值在後半:只驗「遠距離比較打得中」的話,把衰減也一起改掉會照樣過。
func TestRangemasterShortensHitRangeOnly(t *testing.T) {
	// 9 格:命中門檻 60;裝了測距瞄準器等於 3 格,門檻 40。
	// netAttack=10 + roll=40 = 50 落在兩者之間,所以同一發在「有/沒有」之間會翻面。
	mk := func(rm bool) BeamShot {
		return BeamShot{NetAttack: 10, WeaponMin: 20, WeaponMax: 20, RangeSquares: 9, Roll: 40,
			Attacker: BeamAttackerSystems{Rangemaster: rm}}
	}
	if r := ResolveBeamShot(mk(false)); r.Hit {
		t.Fatalf("測試前提不成立:這一發在 9 格外應該打不中(%+v)", r)
	}
	if r := ResolveBeamShot(mk(true)); !r.Hit {
		t.Fatalf("裝了測距瞄準器應該打得中(%+v)", r)
	}

	// 後半才是這條測試真正的價值:**傷害衰減不受影響**。
	// 只驗「遠距離比較打得中」的話,把衰減也一起改掉會照樣過。
	sure := func(rm bool, sq int) int {
		r := ResolveBeamShot(BeamShot{NetAttack: 500, WeaponMin: 20, WeaponMax: 20,
			RangeSquares: sq, Roll: 1, Attacker: BeamAttackerSystems{Rangemaster: rm}})
		if !r.Hit {
			t.Fatalf("測試前提不成立:netAttack=500 必中(%+v)", r)
		}
		return r.DamageToStructure
	}
	near, far := sure(false, 3), sure(false, 9)
	if near == far {
		t.Fatalf("測試前提不成立:3 格與 9 格的衰減相同(都是 %d),這條比對沒有鑑別力", near)
	}
	if got := sure(true, 9); got != far {
		t.Errorf("測距瞄準器不該影響傷害衰減:9 格傷害 %d,期望與未裝的 %d 相同(而不是 3 格的 %d)",
			got, far, near)
	}
}

// TestEnergyAbsorberStoresAndFires 能量吸收器:被打時存 1/4,發射時自動命中。
func TestEnergyAbsorberStoresAndFires(t *testing.T) {
	def := CombatShip{Name: "吸收艦", HP: 100, MaxHP: 100, EnergyAbsorber: true}
	EnergyAbsorberAbsorb(&def, 40)
	if want := 40 / gamedata.EnergyAbsorberStoredFraction; def.StoredEnergy != want {
		t.Fatalf("轉存 %d,期望 %d", def.StoredEnergy, want)
	}
	// 沒裝的船不該存。
	plain := CombatShip{}
	EnergyAbsorberAbsorb(&plain, 40)
	if plain.StoredEnergy != 0 {
		t.Error("沒裝吸收器的船不該有儲能")
	}
	// 發射:距離 1 格(衰減 0),自動命中,而且儲能清空。
	tgt := CombatShip{HP: 100, MaxHP: 100, ArmorHP: 0}
	r := EnergyAbsorberRelease(&def, &tgt, 1, 0)
	if !r.Hit || r.DamageToStructure != 10 {
		t.Errorf("發射儲能:%+v,期望命中且 10 點傷害", r)
	}
	if def.StoredEnergy != 0 {
		t.Errorf("發射後儲能=%d,應該清空", def.StoredEnergy)
	}
	// 儲能為 0 時不算一發。
	if r2 := EnergyAbsorberRelease(&def, &tgt, 1, 0); r2.Hit {
		t.Error("沒有儲能時不該有這一發")
	}
}

// TestEnergyAbsorberBlockedByDisplacement 手冊唯一寫明的例外:目標有位移裝置時不保證命中。
func TestEnergyAbsorberBlockedByDisplacement(t *testing.T) {
	src := CombatShip{EnergyAbsorber: true, StoredEnergy: 40}
	tgt := CombatShip{HP: 100, MaxHP: 100, HasDisplacement: true}
	if r := EnergyAbsorberRelease(&src, &tgt, 1, gamedata.MissileDisplacementDeviceMissChance); r.Hit {
		t.Error("擲進位移裝置的機率內應該未命中")
	}
	src.StoredEnergy = 40
	if r := EnergyAbsorberRelease(&src, &tgt, 1, gamedata.MissileDisplacementDeviceMissChance+1); !r.Hit {
		t.Error("擲出位移裝置的機率外應該自動命中")
	}
}

// TestCloakComponentsReachCombatShip 四個新元件都要真的變成 CombatShip 上的旗標
// ——「元件表有」不等於「效果有接」(第 72 項(元件表有≠效果有接))。
func TestCloakComponentsReachCombatShip(t *testing.T) {
	cases := []struct {
		special string
		check   func(CombatShip) bool
	}{
		{"隱形裝置", func(c CombatShip) bool { return c.CloakKind == CloakDevice && c.Cloaked }},
		{"相位匿蹤", func(c CombatShip) bool { return c.CloakKind == CloakPhasing && c.Cloaked }},
		{"能量吸收器", func(c CombatShip) bool { return c.EnergyAbsorber }},
		{"測距瞄準器", func(c CombatShip) bool { return c.BeamSystems.Rangemaster }},
	}
	for _, c := range cases {
		s := NewDemoSession()
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{
			Name: "測試艦", Class: "戰艦", Weapon: "雷射砲", Special: c.special})
		player, _ := s.StartCombat("測試敵人")
		if !c.check(player[len(player)-1]) {
			t.Errorf("%s 沒有接到格子戰術的 CombatShip 上", c.special)
		}
		cs, _ := s.mkPlayerCombatantsIndexed()
		last := cs[len(cs)-1]
		switch c.special {
		case "隱形裝置":
			if last.cloakKind != CloakDevice || !last.cloaked {
				t.Error("隱形裝置沒有接到快速結算的 combatant 上")
			}
		case "相位匿蹤":
			if last.cloakKind != CloakPhasing || !last.cloaked {
				t.Error("相位匿蹤沒有接到快速結算的 combatant 上")
			}
		case "能量吸收器":
			if !last.energyAbsorber {
				t.Error("能量吸收器沒有接到快速結算的 combatant 上")
			}
		case "測距瞄準器":
			if !last.beamSystems.Rangemaster {
				t.Error("測距瞄準器沒有接到快速結算的 combatant 上")
			}
		}
	}
}
