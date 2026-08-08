package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 硬化護盾的 −3 是「**每一次敵方攻擊**」,不分武器類型。
//
// 手冊逐字:「This reduces the damage of **each enemy attack** — by 3 points — regardless of
// whether or not the shield in that quarter has collapsed.」
//
// ⚠ 第 71 項(探針③內部函式)之前只有光束路徑吃得到,飛彈與球形武器的呼叫端一律傳 false
// ——不是有意的簡化,是那兩個參數位置從加進來那天起就沒人回頭填。
func TestHardShieldReducesEveryKindOfAttack(t *testing.T) {
	const wmax, shield, armor = 40, 5, 0

	t.Run("飛彈", func(t *testing.T) {
		soft := ResolveMissileShot(false, 0, 1, 0, 0, false, 100, wmax, shield, armor, false, MissileDefenses{})
		hard := ResolveMissileShot(false, 0, 1, 0, 0, false, 100, wmax, shield, armor, true, MissileDefenses{})
		if !soft.Hit || !hard.Hit {
			t.Fatalf("測試前提不成立:兩發都應命中(%+v / %+v)", soft, hard)
		}
		if want := soft.DamageToStructure - gamedata.DamageHardShieldBonus; hard.DamageToStructure != want {
			t.Errorf("硬化護盾應再減 %d:%d → %d(期望 %d)",
				gamedata.DamageHardShieldBonus, soft.DamageToStructure, hard.DamageToStructure, want)
		}
	})

	t.Run("球形武器", func(t *testing.T) {
		const aggD = 40
		soft := ResolveSphericalShot(aggD, shield, armor, false, false)
		hard := ResolveSphericalShot(aggD, shield, armor, true, false)
		if want := soft.DamageToStructure - gamedata.DamageHardShieldBonus; hard.DamageToStructure != want {
			t.Errorf("硬化護盾應再減 %d:%d → %d(期望 %d)",
				gamedata.DamageHardShieldBonus, soft.DamageToStructure, hard.DamageToStructure, want)
		}
	})

	t.Run("光束", func(t *testing.T) {
		base := BeamShot{NetAttack: 90, WeaponMin: 40, WeaponMax: 60, RangeSquares: 1, Roll: 40,
			Target: BeamTargetSystems{ShieldReduction: shield}}
		soft := ResolveBeamShot(base)
		hardIn := base
		hardIn.Target.HardShield = true
		hard := ResolveBeamShot(hardIn)
		if want := soft.DamageToStructure - gamedata.DamageHardShieldBonus; hard.DamageToStructure != want {
			t.Errorf("硬化護盾應再減 %d:%d → %d(期望 %d)",
				gamedata.DamageHardShieldBonus, soft.DamageToStructure, hard.DamageToStructure, want)
		}
	})
}

// 手冊那句的後半也要成立:護盾**已經被打穿**(減傷為 0)時,−3 仍然生效。
//
// 「regardless of whether or not the shield in that quarter has collapsed」——
// 這是硬化護盾與一般護盾的關鍵差別,不是文字修飾。
func TestHardShieldStillWorksWhenShieldsAreDown(t *testing.T) {
	const wmax = 40
	soft := ResolveMissileShot(false, 0, 1, 0, 0, false, 100, wmax, 0, 0, false, MissileDefenses{})
	hard := ResolveMissileShot(false, 0, 1, 0, 0, false, 100, wmax, 0, 0, true, MissileDefenses{})
	if want := soft.DamageToStructure - gamedata.DamageHardShieldBonus; hard.DamageToStructure != want {
		t.Errorf("護盾歸零時硬化護盾仍應減 %d:%d → %d",
			gamedata.DamageHardShieldBonus, soft.DamageToStructure, hard.DamageToStructure)
	}
}

// 硬化護盾真的接到兩條戰鬥路徑的**防守方**欄位上。
func TestHardShieldReachesBothCombatPaths(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{
		Name: "測試艦", Class: "戰艦", Weapon: "雷射砲", Shield: "第一級護盾", Special: "硬化護盾"})

	cs, _ := s.mkPlayerCombatantsIndexed()
	if !cs[len(cs)-1].hardShield {
		t.Error("快速結算的 combatant 應帶著硬化護盾旗標")
	}
	player, _ := s.StartCombat("測試敵人")
	if !player[len(player)-1].HardShield {
		t.Error("格子戰術的 CombatShip 應帶著硬化護盾旗標")
	}
	// 反面:沒裝的船不該有。
	plain := NewDemoSession()
	plain.Fleet().Ships = append(plain.Fleet().Ships, Ship{Name: "普通艦", Class: "戰艦", Weapon: "雷射砲"})
	pcs, _ := plain.mkPlayerCombatantsIndexed()
	if pcs[len(pcs)-1].hardShield {
		t.Error("沒裝硬化護盾的船不該有這個旗標")
	}
}

// 手冊同一段還有第三、第四個效果——一個已經做了、一個擋在缺席的系統後面。
//
// 釘住現況免得日後盤點時重算一次:
//   - 「allow ships to use their shields inside a nebula」→ 已做(nebula.go 的 nebulaShield)
//   - 「provide immunity to shield-piercing weapons」→ 已做(DamageAfterShield 的 shieldPiercing 分支)
//   - 「prevent enemies from using Transporters to send over Marines」→ **傳送器不存在**,
//     擋在登艦戰系統後面(第 60 項(打得準也閃得掉)),不是漏抄
func TestHardShieldOtherManualEffects(t *testing.T) {
	// 護盾穿透武器對硬化護盾無效。
	const dmg, shield = 40, 10
	if got := gamedata.DamageAfterShield(dmg, shield, false, true); got != dmg {
		t.Errorf("一般護盾應被穿透武器完全繞過,得到 %d", got)
	}
	if got := gamedata.DamageAfterShield(dmg, shield, true, true); got >= dmg {
		t.Errorf("硬化護盾應對穿透武器免疫(仍然減傷),得到 %d", got)
	}
	// 星雲內:沒有硬化護盾的船護盾歸零,有的不受影響。
	if !shipHasHardShield(Ship{Special: "硬化護盾"}) {
		t.Error("元件名比對應認得硬化護盾")
	}
}

// TestHardShieldReachesBeamCallSites 是**呼叫端**的回歸,不是公式的。
//
// ⚠ 2026-08-08 抓到的漏:`ResolveBeamShot` 從第 68 項(元件盤點+飛彈防禦)起就有 `Target.HardShield`,
// 而兩條戰鬥路徑的光束分支**都沒有填它**——`BeamTargetSystems` 是結構,漏填一個欄位
// 拿到的是零值,不會編譯失敗、也不會有任何測試紅。上面那幾條測試都直接呼叫
// `ResolveBeamShot` 並自己填 `HardShield: true`,所以驗的是公式,驗不到呼叫端。
//
// combatant.hardShield 的註解當時甚至寫著「三條路徑都要吃到,之前只有光束路徑接了」
// ——事實相反,光束是唯一沒接的那一條。
//
// 這條測試走 battleShot(快速結算的真正入口),同一顆種子、同一批船,只差硬化護盾。
func TestHardShieldReachesBeamCallSites(t *testing.T) {
	mk := func(hard bool) ([]combatant, []combatant) {
		atk := []combatant{{hp: 100, atk: 200, def: 0, wmin: 40, wmax: 40,
			kind: WeaponKindBeam, shots: 1}}
		def := []combatant{{hp: 1000, atk: 0, def: 0, wmin: 0, wmax: 0,
			kind: WeaponKindBeam, shots: 1, shield: 5, hardShield: hard, shipIdx: -1}}
		return atk, def
	}
	run := func(hard bool) int {
		a, d := mk(hard)
		rng := rand.New(rand.NewSource(20260808))
		battleShot(&a[0], &d, rng)
		return 1000 - d[0].hp
	}
	soft, hardDmg := run(false), run(true)
	if soft == 0 {
		t.Fatal("測試前提不成立:這一發應該命中")
	}
	if want := soft - gamedata.DamageHardShieldBonus; hardDmg != want {
		t.Errorf("快速結算的光束路徑沒有吃到硬化護盾:%d → %d(期望 %d)", soft, hardDmg, want)
	}
}
