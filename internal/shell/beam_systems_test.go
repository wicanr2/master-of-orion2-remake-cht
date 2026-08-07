package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 重構的驗收:舊入口與新入口對同一組輸入必須給出**完全相同**的結果。
//
// 這一條先跑,其餘測試才有意義——如果包裝寫歪了,後面所有斷言都在測錯的東西。
func TestOldEntryPointsDelegateIdentically(t *testing.T) {
	ap := []gamedata.WeaponModCode{gamedata.ModArmorPiercing}
	cases := []struct {
		name                                            string
		net, wmin, wmax, rng, shield, armor, roll, hefB int
		hard, apNeg                                     bool
		mods                                            []gamedata.WeaponModCode
	}{
		{"無改造", 60, 40, 100, 1, 0, 0, 50, 0, false, false, nil},
		{"遠距離", 60, 40, 100, 6, 0, 0, 50, 0, false, false, nil},
		{"有盾有甲", 80, 30, 70, 2, 5, 60, 30, 0, false, false, nil},
		{"高能聚焦", 60, 40, 100, 1, 0, 0, 50, gamedata.DamageMountBonusHEF, false, false, nil},
		{"穿甲", 80, 40, 60, 1, 0, 200, 60, 0, false, false, ap},
		{"穿甲被抵銷", 80, 40, 60, 1, 0, 200, 60, 0, false, true, ap},
		{"硬化護盾", 80, 30, 70, 2, 5, 60, 30, 0, true, false, nil},
	}
	for _, c := range cases {
		old := ResolveShotWithMods(c.net, c.wmin, c.wmax, c.rng, c.shield, c.armor, c.roll,
			c.hard, c.mods, c.hefB, c.apNeg)
		neu := ResolveBeamShot(BeamShot{
			NetAttack: c.net, WeaponMin: c.wmin, WeaponMax: c.wmax,
			RangeSquares: c.rng, Roll: c.roll, Mods: c.mods,
			Attacker: BeamAttackerSystems{HEFBonus: c.hefB},
			Target: BeamTargetSystems{
				ShieldReduction: c.shield, ArmorHP: c.armor,
				HardShield: c.hard, APNegated: c.apNeg,
			},
		})
		if old != neu {
			t.Errorf("%s:舊入口與新入口結果不同\n  舊 %+v\n  新 %+v", c.name, old, neu)
		}
	}
}

// 結構分析儀:過盾之後的傷害加倍,而且**護盾照扣**。
//
// 順序寫反(先加倍再扣盾)在無護盾時看不出差別,有護盾時才會露餡——所以測有護盾的情況。
func TestStructuralAnalyzerDoublesDamageAfterShields(t *testing.T) {
	const net, wmin, wmax, roll, shield = 90, 40, 60, 40, 8
	base := BeamShot{
		NetAttack: net, WeaponMin: wmin, WeaponMax: wmax, RangeSquares: 1, Roll: roll,
		Target: BeamTargetSystems{ShieldReduction: shield},
	}
	plain := ResolveBeamShot(base)
	if !plain.Hit || plain.DamageToStructure == 0 {
		t.Fatalf("測試前提不成立:基準應命中且有傷害(%+v)", plain)
	}
	sa := base
	sa.Attacker.StructuralAnalyzer = true
	got := ResolveBeamShot(sa)
	if want := plain.DamageToStructure * gamedata.ShipStructuralAnalyzerMultiplier; got.DamageToStructure != want {
		t.Errorf("過盾傷害應加倍為 %d,得到 %d", want, got.DamageToStructure)
	}
	// 若順序寫反(先加倍再扣盾),結果會是 (dmg*2 - shield) 而不是 (dmg - shield)*2。
	if wrong := (plain.DamageToStructure+shield)*2 - shield; got.DamageToStructure == wrong {
		t.Errorf("加倍發生在扣盾之前了(得到 %d = 先加倍的結果)", got.DamageToStructure)
	}
}

// 阿基里斯瞄準器:光束一律無視裝甲,不必掛 AP 改造。
func TestAchillesUnitIgnoresArmorWithoutTheAPMod(t *testing.T) {
	const net, wmin, wmax, roll, armor = 90, 40, 60, 40, 200
	base := BeamShot{
		NetAttack: net, WeaponMin: wmin, WeaponMax: wmax, RangeSquares: 1, Roll: roll,
		Target: BeamTargetSystems{ArmorHP: armor},
	}
	plain := ResolveBeamShot(base)
	if !plain.Hit {
		t.Fatal("測試前提不成立:基準應命中")
	}
	if plain.DamageToStructure != 0 {
		t.Fatalf("測試前提不成立:沒有穿甲時傷害應全被裝甲吃掉,得到 %d", plain.DamageToStructure)
	}
	ach := base
	ach.Attacker.AchillesUnit = true
	got := ResolveBeamShot(ach)
	if got.DamageToStructure == 0 {
		t.Error("阿基里斯瞄準器應讓傷害直接進結構")
	}
	if got.RemainingArmorHP != armor {
		t.Errorf("無視裝甲時裝甲不該被消耗:%d → %d", armor, got.RemainingArmorHP)
	}
	// 與 AP 改造的效果應一致。
	apShot := base
	apShot.Mods = []gamedata.WeaponModCode{gamedata.ModArmorPiercing}
	if ap := ResolveBeamShot(apShot); ap != got {
		t.Errorf("應與 AP 改造效果相同:%+v vs %+v", ap, got)
	}
}

// 阿基里斯瞄準器**會被**重裝甲/氙素裝甲抵銷。
//
// ⚠ 這是讀法不是手冊字面:那兩條寫「negates the Armor Piercing abilities of enemy weapons」,
// 字面涵蓋「無視裝甲」這個能力。釘住它是為了讓日後改讀法的人知道自己在改什麼。
func TestAchillesIsNegatedByArmorPiercingNegation(t *testing.T) {
	const net, wmin, wmax, roll, armor = 90, 40, 60, 40, 200
	shot := BeamShot{
		NetAttack: net, WeaponMin: wmin, WeaponMax: wmax, RangeSquares: 1, Roll: roll,
		Attacker: BeamAttackerSystems{AchillesUnit: true},
		Target:   BeamTargetSystems{ArmorHP: armor, APNegated: true},
	}
	got := ResolveBeamShot(shot)
	if got.DamageToStructure != 0 {
		t.Errorf("被抵銷時傷害應全部由裝甲吸收,結構卻吃了 %d", got.DamageToStructure)
	}
	if got.RemainingArmorHP >= armor {
		t.Errorf("被抵銷的那一發應改為消耗裝甲:%d → %d", armor, got.RemainingArmorHP)
	}
}

// 兩個新系統在兩條戰鬥路徑上都掛得上。
func TestBeamSystemsReachBothPaths(t *testing.T) {
	for _, name := range []string{"結構分析儀", "阿基里斯瞄準器", highEnergyFocusName} {
		s := NewDemoSession()
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{
			Name: "測試艦", Class: "戰艦", Weapon: "雷射砲", Special: name,
		})
		cs, _ := s.mkPlayerCombatantsIndexed()
		got := cs[len(cs)-1].beamSystems
		if got == (BeamAttackerSystems{}) {
			t.Errorf("%s:快速結算的攻方系統不該全空", name)
		}
		player, _ := s.StartCombat("測試敵人")
		if last := player[len(player)-1]; last.BeamSystems != got {
			t.Errorf("%s:兩條路徑的攻方系統應相同:%+v vs %+v", name, last.BeamSystems, got)
		}
	}
}

// 兩個新元件的主題對得上執行檔。
func TestNewBeamSystemsHaveRealTopics(t *testing.T) {
	want := map[string]gamedata.Technology{
		"結構分析儀":   gamedata.TECH_STRUCTURAL_ANALYZER,
		"阿基里斯瞄準器": gamedata.TECH_ACHILLES_TARGETING_UNIT,
	}
	have := map[string]Component{}
	for _, c := range SpecialOptions {
		have[c.Name] = c
	}
	for name, tech := range want {
		c, ok := have[name]
		if !ok {
			t.Errorf("%s 應出現在 SpecialOptions", name)
			continue
		}
		if c.UnlockTech != tech {
			t.Errorf("%s 的解鎖科技對不上:%v", name, c.UnlockTech)
		}
		topic, ok := gamedata.OrigTechTopic(tech)
		if !ok {
			t.Errorf("執行檔應查得到 %s 的主題", name)
			continue
		}
		if c.Tech != topic {
			t.Errorf("%s 的主題應為 %v,得到 %v", name, topic, c.Tech)
		}
	}
}
