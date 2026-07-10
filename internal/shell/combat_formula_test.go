package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestResolveShot 驗證真戰鬥公式接線:近距高攻必中且傷害穿甲、超遠距低攻低 roll 不中、護盾減傷。
func TestResolveShot(t *testing.T) {
	// 近距(1 格)、netAttack 高、roll 高 → 命中,傷害先耗裝甲再傷結構。
	r := ResolveShot(80 /*net*/, 10, 20 /*dmg*/, 1 /*range*/, 0 /*shield*/, 5 /*armorHP*/, 90 /*roll*/, false, false)
	if !r.Hit {
		t.Fatalf("近距高攻高 roll 應命中")
	}
	if r.RemainingArmorHP > 5 {
		t.Fatalf("命中後裝甲不應增加")
	}

	// 遠距(距離大)+ 低 netAttack + 低 roll → 命中門檻高,應未命中。
	miss := ResolveShot(0 /*net*/, 10, 20, 30 /*遠*/, 0, 10, 5 /*低 roll*/, false, false)
	if miss.Hit {
		t.Fatalf("遠距低攻低 roll 不應命中(門檻高)")
	}
	if miss.RemainingArmorHP != 10 {
		t.Fatalf("未命中裝甲應不變")
	}

	// 護盾減傷:同一發,護盾高者打到結構的傷害應較少或相等。
	hi := ResolveShot(90, 20, 20, 1, 0 /*無盾*/, 0 /*無甲*/, 90, false, false)
	lo := ResolveShot(90, 20, 20, 1, 15 /*高盾*/, 0, 90, false, false)
	if lo.DamageToStructure > hi.DamageToStructure {
		t.Fatalf("護盾應減傷:無盾 %d 應 >= 有盾 %d", hi.DamageToStructure, lo.DamageToStructure)
	}
}

// TestResolveMissileShotAMRIntercept 驗證 AMR(反飛彈火箭)攔截分支:命中率查
// gamedata.MissileAMRChanceToHit(依 MissileAMRRangeIndex 換算的 Range 索引),
// amrRoll <= 命中率則整枚飛彈被擊落、不進入 Jam Chance 判定、不造成傷害。
func TestResolveMissileShotAMRIntercept(t *testing.T) {
	// sq=0..2 → Range1 → AMR 命中率 61%(見 gamedata.TestMissileAMREndToEnd)。
	hit := ResolveMissileShot(true, 1, 61 /*amrRoll<=61 命中*/, 0, 0, false, 1,
		10 /*weaponMax*/, 0, 20, false)
	if hit.Hit {
		t.Fatalf("amrRoll<=AMR命中率應被攔截,不應命中")
	}
	if hit.RemainingArmorHP != 20 {
		t.Fatalf("被 AMR 攔截不應扣裝甲,got %d", hit.RemainingArmorHP)
	}

	miss := ResolveMissileShot(true, 1, 62 /*amrRoll>61 攔截失敗*/, 0, 0, false, 1,
		10, 0, 20, false)
	if !miss.Hit {
		t.Fatalf("AMR 攔截失敗、且無閃避裝備時應命中")
	}

	// 超出 AMR 最大射程(15 格)不觸發攔截判定。
	beyond := ResolveMissileShot(true, gamedata.MissileAMRMaxRangeSquares+1, 1, 0, 0, false, 1,
		10, 0, 20, false)
	if !beyond.Hit {
		t.Fatalf("超出 AMR 射程應直接跳過攔截判定,直接進入命中流程")
	}
}

// TestResolveMissileShotJamChance 驗證 Jam Chance 判定:手冊範例(Wide Area Jammer 艦隊
// +70、Stabilizer +25、種族 -20、一般艦員 +7、統帥一半 5 = 87;掃描器 20;ECCM 減半)
// jamChance=33%,對照 gamedata.TestMissileJamChance。
func TestResolveMissileShotJamChance(t *testing.T) {
	const defenderEvasionBonus = 87
	const attackerScannerBonus = 20
	// jamChance=33 → hitChance=67。jamRoll=67 命中,68 被幹擾。
	hit := ResolveMissileShot(false, 0, 1, defenderEvasionBonus, attackerScannerBonus, true, 67,
		10, 0, 20, false)
	if !hit.Hit {
		t.Fatalf("jamRoll(67)<=hitChance(67) 應命中")
	}
	miss := ResolveMissileShot(false, 0, 1, defenderEvasionBonus, attackerScannerBonus, true, 68,
		10, 0, 20, false)
	if miss.Hit {
		t.Fatalf("jamRoll(68)>hitChance(67) 應被幹擾,不應命中")
	}

	// 無任何閃避裝備(現行 remake 現況):jamChance=0,任何 jamRoll(1-100)都必中,
	// 對應手冊「若目標無任何閃避能力,預設100%命中」(MissileDefaultHitChance)。
	for _, roll := range []int{1, 50, 100} {
		always := ResolveMissileShot(false, 0, 1, 0, 0, false, roll, 10, 0, 20, false)
		if !always.Hit {
			t.Fatalf("無閃避裝備時 jamRoll=%d 也應必中", roll)
		}
	}
}

// TestResolveMissileShotDamageThroughShieldArmor 驗證命中後傷害走與 beam 相同的
// 過盾→過甲管線(DamageAfterShield/DamageApplyArmor),只是命中判定機制不同。
func TestResolveMissileShotDamageThroughShieldArmor(t *testing.T) {
	r := ResolveMissileShot(false, 0, 1, 0, 0, false, 1, 20 /*weaponMax*/, 5 /*shield*/, 10 /*armor*/, false)
	if !r.Hit {
		t.Fatalf("應命中")
	}
	// 20 傷害 - 護盾 5 = 15;裝甲 10 只能扛 10,溢出 5 打結構。
	if r.DamageToStructure != 5 || r.RemainingArmorHP != 0 {
		t.Fatalf("DamageToStructure=%d RemainingArmorHP=%d,預期 5/0", r.DamageToStructure, r.RemainingArmorHP)
	}
}

// TestResolveMissileVsBeamDivergence 舉具體例證明 missile 分支確實不是 beam 邏輯的別名:
// 同樣「攻方 Beam Attack 極弱、守方 Beam Defense 極強」的情境,beam 命中門檻判定會 miss,
// missile 因為改用 Jam Chance(現行無閃避裝備 → 必中)結果不同。
func TestResolveMissileVsBeamDivergence(t *testing.T) {
	const netAttack = -900 // 攻方遠遜於守方,beam 幾乎必不中(只剩 roll>95 的 5% 例外)
	const roll = 50        // 中段 roll,不落在 roll>95 例外
	beam := ResolveShot(netAttack, 5, 10, 10 /*range*/, 0, 20, roll, false, false)
	if beam.Hit {
		t.Fatalf("beam 在極端劣勢 net attack + 中段 roll 下不應命中(前提有誤)")
	}

	missile := ResolveMissileShot(false, 0, 1, 0, 0, false, roll, 10, 0, 20, false)
	if !missile.Hit {
		t.Fatalf("missile 應忽略 net attack、只看 Jam Chance(現行無閃避裝備必中),結果卻仍未命中")
	}
}

// TestResolveSphericalShot 驗證球形武器傷害走 DamageAfterShield/DamageApplyArmor
// (bypassShieldAndArmor=false),以及 Spatial-Compressor 類「忽略護盾裝甲、全打結構」
// (bypassShieldAndArmor=true)與「最低傷害 1」的夾限。
func TestResolveSphericalShot(t *testing.T) {
	r := ResolveSphericalShot(20 /*aggD*/, 5 /*shield*/, 10 /*armor*/, false, false)
	if !r.Hit || r.DamageToStructure != 5 || r.RemainingArmorHP != 0 {
		t.Fatalf("got Hit=%v DamageToStructure=%d RemainingArmorHP=%d,預期 true/5/0",
			r.Hit, r.DamageToStructure, r.RemainingArmorHP)
	}

	bypass := ResolveSphericalShot(20, 5, 10, false, true)
	if bypass.DamageToStructure != 20 || bypass.RemainingArmorHP != 10 {
		t.Fatalf("bypass 模式應忽略護盾/裝甲、全部打結構,got DamageToStructure=%d RemainingArmorHP=%d",
			bypass.DamageToStructure, bypass.RemainingArmorHP)
	}

	// aggD=0 應夾為 1(手冊「minimum damage of 1 against ships」);armorHP=0 讓這 1 點
	// 直接打結構,才能從 DamageToStructure 觀察到夾限生效(armorHP>0 時 1 點會被裝甲
	// 全吸收,DamageToStructure 仍是 0,那是裝甲機制本身的正常行為,不是夾限沒生效)。
	clamped := ResolveSphericalShot(0, 0, 0, false, false)
	if clamped.DamageToStructure != 1 {
		t.Fatalf("手冊:對艦最低傷害 1,got %d", clamped.DamageToStructure)
	}
}

// TestWeaponKindByName 核對武器→戰鬥解算路徑分類(見 weapon_kind.go 的核對依據)。
// 特別驗證「死光」不能被誤歸類成 spherical(手冊 Notes on Spherical Damage 列的球形武器是
// Pulsar/Plasma Flux/Spatial Compressor,死光是 damage.go DamageForHit worked example 的
// 出處,屬一般光束武器)。
func TestWeaponKindByName(t *testing.T) {
	cases := []struct {
		name string
		want WeaponKind
	}{
		{"核飛彈", WeaponKindMissile},
		{"麥克萊特飛彈", WeaponKindMissile},
		{"雷射", WeaponKindBeam},
		{"質量投射器", WeaponKindBeam},
		{"中子爆破槍", WeaponKindBeam},
		{"核融合光束", WeaponKindBeam},
		{"高斯砲", WeaponKindBeam},
		{"相位砲", WeaponKindBeam},
		{"電漿砲", WeaponKindBeam},
		{"死光", WeaponKindBeam}, // 見上方註解:不是 spherical
		{"無武裝", WeaponKindBeam},
	}
	for _, c := range cases {
		if got := weaponKindByName(c.name); got != c.want {
			t.Errorf("weaponKindByName(%q) = %v,預期 %v", c.name, got, c.want)
		}
	}
}

// TestBattleVolleyDispatchByWeaponKind 驗證 battleVolley 真的依 attacker.kind 分流,而不是
// 全部仍走 beam(回歸/整合層級測試,呼應 combat_formula_test.go 純函式層級的
// TestResolveMissileVsBeamDivergence)。
func TestBattleVolleyDispatchByWeaponKind(t *testing.T) {
	// armor 刻意設得比武器 wmax(10)低,讓命中後的傷害能溢出打到結構(hp),
	// 才能用 hp 是否變動來判斷「有沒有真的命中」。
	mkDefender := func() combatant { return combatant{hp: 100, def: 999, armor: 5, shield: 0} }

	// missile 分支:不看 def/net attack,只受 Jam Chance;現行無任何飛彈閃避元件 →
	// jamChance=0 → 每個 seed 都必中(deterministic,不受 roll 影響)。
	for seed := int64(1); seed <= 10; seed++ {
		defenders := []combatant{mkDefender()}
		attackers := []combatant{{wmin: 5, wmax: 10, kind: WeaponKindMissile}}
		battleVolley(attackers, &defenders, rand.New(rand.NewSource(seed)))
		if defenders[0].hp == 100 {
			t.Fatalf("seed=%d:missile 應忽略極端劣勢 net attack 必中,hp 卻維持 100(疑仍走 beam 判定)", seed)
		}
	}

	// beam 分支:同樣極端劣勢 net attack(atk=1,def=999)在此距離門檻下只有 roll>95(5%)
	// 才會命中,掃過多個 seed 應能找到至少一個「未命中」的 seed——證明 beam 分支仍是
	// net-attack/range 門檻判定,與 missile 明顯不同,行為未被這次改動動到。
	missCount := 0
	for seed := int64(1); seed <= 30; seed++ {
		defenders := []combatant{mkDefender()}
		attackers := []combatant{{atk: 1, wmin: 5, wmax: 10, kind: WeaponKindBeam}}
		battleVolley(attackers, &defenders, rand.New(rand.NewSource(seed)))
		if defenders[0].hp == 100 {
			missCount++
		}
	}
	if missCount == 0 {
		t.Fatalf("beam 極端劣勢 net attack 應在 30 個 seed 中至少出現一次未命中(理論 95%% 機率),卻全部命中")
	}
}
