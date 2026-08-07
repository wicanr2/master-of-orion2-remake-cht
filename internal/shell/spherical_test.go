package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 球形武器分支終於掛得到武器了。
//
// `weapon_kind.go` 的檔頭原本寫著:「spherical 分支目前**沒有任何武器掛載**,
// ResolveSphericalShot 只提供已測試的解算函式待未來新增球形武器元件時串接」——
// 整條解算路徑是死碼。
func TestSphericalWeaponsAreMounted(t *testing.T) {
	names := []string{"脈衝星", "空間壓縮器"}
	inRoster := map[string]Component{}
	for _, c := range WeaponOptions {
		inRoster[c.Name] = c
	}
	for _, n := range names {
		c, ok := inRoster[n]
		if !ok {
			t.Errorf("武器清單裡缺 %s", n)
			continue
		}
		if weaponKindByName(n) != WeaponKindSpherical {
			t.Errorf("%s 應分類為球形武器", n)
		}
		topic, ok := gamedata.OrigTechTopic(c.UnlockTech)
		if !ok || topic != c.Tech {
			t.Errorf("%s 的研究主題應為執行檔給的 %v,元件表寫的是 %v", n, topic, c.Tech)
		}
	}
}

// 「per size class of target」:同一發打大船比打小船痛。
func TestSphericalDamageScalesWithTargetSize(t *testing.T) {
	dmg := func(size gamedata.CombatShipClass) int {
		total := 0
		for seed := int64(1); seed <= 200; seed++ {
			atk := []combatant{{hp: 100, atk: 24, wmin: 1, wmax: 24,
				kind: WeaponKindSpherical, weaponName: "脈衝星", shipIdx: -1}}
			def := []combatant{{hp: 1 << 20, atk: 1, def: 1, sizeClass: size, shipIdx: -1}}
			before := def[0].hp
			battleVolley(atk, &def, rand.New(rand.NewSource(seed)))
			total += before - def[0].hp
		}
		return total
	}
	small, big := dmg(gamedata.SHIP_FRIGATE), dmg(gamedata.SHIP_DOOMSTAR)
	if small <= 0 {
		t.Fatalf("測試前提不成立:球形武器應該打得動(小船總傷 %d)", small)
	}
	if big <= small {
		t.Errorf("打末日之星應比打護衛艦痛:%d vs %d", big, small)
	}
}

// 護衛艦那一級**不能**乘 0——取 index 會讓打護衛艦零傷害,那顯然不是規則。
func TestSmallestHullStillTakesSphericalDamage(t *testing.T) {
	atk := []combatant{{hp: 100, atk: 24, wmin: 24, wmax: 24,
		kind: WeaponKindSpherical, weaponName: "脈衝星", shipIdx: -1}}
	def := []combatant{{hp: 1 << 20, atk: 1, def: 1, sizeClass: gamedata.SHIP_FRIGATE, shipIdx: -1}}
	before := def[0].hp
	battleVolley(atk, &def, rand.New(rand.NewSource(1)))
	if def[0].hp >= before {
		t.Error("最小艦體也該吃到球形傷害(級數取 index+1 而不是 index)")
	}
}

// 只有空間壓縮器豁免護盾與裝甲——手冊只有那一項寫了那句話。
func TestOnlySpatialCompressorBypassesShieldAndArmor(t *testing.T) {
	if !weaponBypassesShieldAndArmor("空間壓縮器") {
		t.Error("空間壓縮器應豁免護盾與裝甲(手冊:structure only, ignoring shields and armor)")
	}
	for _, n := range []string{"脈衝星", "雷射", "核彈", "死光"} {
		if weaponBypassesShieldAndArmor(n) {
			t.Errorf("%s 手冊沒有那一句,不該豁免", n)
		}
	}

	// 端到端:同樣的攻防下,壓縮器打穿厚護盾的總傷應高於脈衝星。
	dmg := func(weapon string) int {
		total := 0
		for seed := int64(1); seed <= 200; seed++ {
			atk := []combatant{{hp: 100, atk: 32, wmin: 4, wmax: 32,
				kind: WeaponKindSpherical, weaponName: weapon, shipIdx: -1}}
			def := []combatant{{hp: 1 << 20, atk: 1, def: 1, shield: 10, armor: 50,
				sizeClass: gamedata.SHIP_CRUISER, shipIdx: -1}}
			before := def[0].hp
			battleVolley(atk, &def, rand.New(rand.NewSource(seed)))
			total += before - def[0].hp
		}
		return total
	}
	if comp, puls := dmg("空間壓縮器"), dmg("脈衝星"); comp <= puls {
		t.Errorf("空間壓縮器無視護盾/裝甲,對厚甲目標的結構傷應較高:%d vs %d", comp, puls)
	}
}

// 艦體等級真的從船帶進 combatant(不是永遠 0)。
func TestSizeClassReachesCombatant(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = []Ship{{Name: "大艦", Class: "末日之星"}, {Name: "小艦", Class: "巡防艦"}}
	cs := s.mkPlayerCombatants()
	if len(cs) != 2 {
		t.Fatalf("應有兩艘參戰艦,得到 %d", len(cs))
	}
	if cs[0].sizeClass <= cs[1].sizeClass {
		t.Errorf("末日之星的艦體等級應高於巡防艦:%d vs %d", cs[0].sizeClass, cs[1].sizeClass)
	}
}
