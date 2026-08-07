package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 裝甲 HP 要照手冊的倍率階梯,不是自編數字。
//
// ⚠ 這一條同時是一則撤回的驗收:先前的判斷是「裝甲科技的倍率手冊與 openorion2 都沒有,
// 所以不動 armorHPByName」——手冊 Ship 條目裡逐級寫著(+100%/+300%/+500%/+700%/10 倍),
// 只是當時沒讀到那幾頁。「我找不到」與「它不存在」是兩件事。
func TestArmorHPFollowsTheManualLadder(t *testing.T) {
	base := armorHPByName("鈦裝甲")
	if base <= 0 {
		t.Fatalf("測試前提不成立:鈦裝甲應有正的 HP,得到 %d", base)
	}
	cases := []struct {
		name string
		tech gamedata.Technology
	}{
		{"鈦裝甲", gamedata.TECH_TITANIUM_ARMOR},
		{"三鈦裝甲", gamedata.TECH_TRITANIUM_ARMOR},
		{"佐特裝甲", gamedata.TECH_ZORTRIUM_ARMOR},
		{"中子素裝甲", gamedata.TECH_NEUTRONIUM_ARMOR},
		{"精金裝甲", gamedata.TECH_ADAMANTIUM_ARMOR},
		{"氙素裝甲", gamedata.TECH_XENTRONIUM_ARMOR},
	}
	for _, c := range cases {
		want := base * gamedata.ArmorStructurePercent(c.tech) / gamedata.ArmorStructurePercentTitanium
		if got := armorHPByName(c.name); got != want {
			t.Errorf("%s 應為 %d(基準 %d × %d%%),得到 %d",
				c.name, want, base, gamedata.ArmorStructurePercent(c.tech), got)
		}
	}
	// 階梯必須嚴格遞增——抄歪一格之後個別數字還是合理的,單調性會壞掉。
	prev := 0
	for _, c := range cases {
		got := armorHPByName(c.name)
		if got <= prev {
			t.Errorf("%s 的裝甲 HP 應高於前一級:%d → %d", c.name, prev, got)
		}
		prev = got
	}
}

// 氙素裝甲是「10 倍」不是「+1100%」——這一格最容易抄成加法。
func TestXentroniumIsTenTimesNotElevenTimes(t *testing.T) {
	if got := gamedata.ArmorStructurePercent(gamedata.TECH_XENTRONIUM_ARMOR); got != 1000 {
		t.Errorf("氙素裝甲應為 1000%%(手冊:10 times the base),得到 %d%%", got)
	}
	if base, xen := armorHPByName("鈦裝甲"), armorHPByName("氙素裝甲"); xen != base*10 {
		t.Errorf("氙素裝甲應是鈦裝甲的 10 倍:%d vs %d", xen, base*10)
	}
}

// 氙素裝甲的研究主題與解鎖科技訂正(先前掛錯主題且沒有解鎖科技)。
func TestXentroniumArmorHasItsRealTopicAndTech(t *testing.T) {
	var found *Component
	for i := range ArmorOptions {
		if ArmorOptions[i].Name == "氙素裝甲" {
			found = &ArmorOptions[i]
		}
	}
	if found == nil {
		t.Fatal("氙素裝甲應在 ArmorOptions")
	}
	topic, ok := gamedata.OrigTechTopic(gamedata.TECH_XENTRONIUM_ARMOR)
	if !ok {
		t.Fatal("執行檔的 tech→topic 表應查得到氙素裝甲")
	}
	if found.Tech != topic {
		t.Errorf("主題應為執行檔的 %v,得到 %v", topic, found.Tech)
	}
	if found.UnlockTech != gamedata.TECH_XENTRONIUM_ARMOR {
		t.Errorf("解鎖科技應為氙素裝甲,得到 %v", found.UnlockTech)
	}
}

// 重裝甲系統:裝甲耐受量三倍(手冊「triples」)。
func TestHeavyArmorTriplesArmorHP(t *testing.T) {
	plain := Ship{Name: "無系統", Class: "戰艦", Armor: "佐特裝甲"}
	heavy := Ship{Name: "重裝甲", Class: "戰艦", Armor: "佐特裝甲", Special: heavyArmorName}
	if a, b := effectiveArmorHP(plain), effectiveArmorHP(heavy); b != a*gamedata.ArmorHeavyArmorMultiplier {
		t.Errorf("重裝甲應是 %d 倍:%d vs %d", gamedata.ArmorHeavyArmorMultiplier, b, a)
	}
	// 沒有裝甲就沒得乘(0 × 3 仍是 0,不會憑空長出裝甲)。
	none := Ship{Name: "裸船", Class: "戰艦", Armor: "無裝甲", Special: heavyArmorName}
	if got := effectiveArmorHP(none); got != 0 {
		t.Errorf("沒有裝甲時重裝甲不該憑空產生 HP,得到 %d", got)
	}
}

// 穿甲抵銷有**兩條路**,兩條都要算(手冊各給了一句)。
func TestArmorPiercingIsNegatedByBothXentroniumAndHeavyArmor(t *testing.T) {
	cases := []struct {
		name string
		ship Ship
		want bool
	}{
		{"氙素裝甲", Ship{Class: "戰艦", Armor: "氙素裝甲"}, true},
		{"重裝甲系統", Ship{Class: "戰艦", Armor: "佐特裝甲", Special: heavyArmorName}, true},
		{"兩者皆有", Ship{Class: "戰艦", Armor: "氙素裝甲", Special: heavyArmorName}, true},
		{"一般裝甲", Ship{Class: "戰艦", Armor: "佐特裝甲"}, false},
		{"無裝甲", Ship{Class: "戰艦", Armor: "無裝甲"}, false},
		// 反面:別的系統不該有這個效果。
		{"高能聚焦", Ship{Class: "戰艦", Armor: "佐特裝甲", Special: highEnergyFocusName}, false},
	}
	for _, c := range cases {
		if got := shipNegatesArmorPiercing(c.ship); got != c.want {
			t.Errorf("%s 的穿甲抵銷應為 %v,得到 %v", c.name, c.want, got)
		}
	}
}

// 端到端:AP 改造打在會抵銷的目標上就不再穿透。
//
// 這是 apNegated 那個「從寫出來就恆傳 false」的參數第一次有生產端。
func TestArmorPiercingActuallyStopsAtNegatingTargets(t *testing.T) {
	const net, wmin, wmax, roll, armor = 80, 40, 60, 60, 200
	ap := []gamedata.WeaponModCode{gamedata.ModArmorPiercing}

	through := ResolveShotWithMods(net, wmin, wmax, 1, 0, armor, roll, false, ap, 0, false)
	stopped := ResolveShotWithMods(net, wmin, wmax, 1, 0, armor, roll, false, ap, 0, true)
	if !through.Hit || !stopped.Hit {
		t.Fatalf("測試前提不成立:兩發都該命中(%v / %v)", through.Hit, stopped.Hit)
	}
	if through.DamageToStructure == 0 {
		t.Fatal("測試前提不成立:AP 應直接穿到結構")
	}
	if stopped.DamageToStructure != 0 {
		t.Errorf("穿甲被抵銷時傷害應全部由裝甲吸收,結構卻吃了 %d", stopped.DamageToStructure)
	}
	if stopped.RemainingArmorHP >= armor {
		t.Errorf("被抵銷的那一發應改為消耗裝甲:%d → %d", armor, stopped.RemainingArmorHP)
	}
}

// 兩條戰鬥路徑都帶著這個旗標(同 HEF 的理由:只接一邊會安靜地不一致)。
func TestAPNegationReachesBothCombatPaths(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{
		Name: "測試重甲艦", Class: "戰艦", Weapon: "雷射砲", Armor: "佐特裝甲", Special: heavyArmorName,
	})
	cs, _ := s.mkPlayerCombatantsIndexed()
	if len(cs) == 0 {
		t.Fatal("測試前提不成立:應有可戰艦艇")
	}
	if !cs[len(cs)-1].apNegated {
		t.Error("快速結算的戰列應標記穿甲抵銷")
	}
	if want := armorHPByName("佐特裝甲") * gamedata.ArmorHeavyArmorMultiplier; cs[len(cs)-1].armor != want {
		t.Errorf("快速結算的裝甲 HP 應為 %d,得到 %d", want, cs[len(cs)-1].armor)
	}

	player, _ := s.StartCombat("測試敵人")
	last := player[len(player)-1]
	if !last.APNegated {
		t.Error("格子戰術的艦艇應標記穿甲抵銷")
	}
	if want := armorHPByName("佐特裝甲") * gamedata.ArmorHeavyArmorMultiplier; last.ArmorHP != want {
		t.Errorf("格子戰術的裝甲 HP 應為 %d,得到 %d", want, last.ArmorHP)
	}
}
