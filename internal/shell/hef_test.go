package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 高能聚焦裝得上——先前它連在元件清單上都沒有,所以 hefBonus 只能恆傳 0。
func TestHighEnergyFocusIsAnEquippableSystem(t *testing.T) {
	var found *Component
	for i := range SpecialOptions {
		if SpecialOptions[i].Name == highEnergyFocusName {
			found = &SpecialOptions[i]
			break
		}
	}
	if found == nil {
		t.Fatal("高能聚焦應出現在 SpecialOptions")
	}
	if found.Tech != gamedata.TOPIC_HIGH_ENERGY_DISTRIBUTION {
		t.Errorf("研究主題應為高能分配,得到 %v", found.Tech)
	}
	if found.UnlockTech != gamedata.TECH_HIGH_ENERGY_FOCUS {
		t.Errorf("解鎖科技對不上,得到 %v", found.UnlockTech)
	}
	if found.Cost <= 0 {
		t.Errorf("建造成本應為正值,得到 %d", found.Cost)
	}
	if found.Value != 0 {
		t.Errorf("高能聚焦不加攻防,Value 應為 0,得到 %d", found.Value)
	}
	// 它必須真的在那個三選一主題裡,否則研究出來也裝不上。
	choice := gamedata.ResearchChoiceFor(gamedata.TOPIC_HIGH_ENERGY_DISTRIBUTION)
	ok := false
	for _, c := range choice.Choices {
		if c == gamedata.TECH_HIGH_ENERGY_FOCUS {
			ok = true
		}
	}
	if !ok {
		t.Error("高能聚焦應是高能分配主題的選項之一")
	}
}

// 手冊那三句話各自對應一個位置——三句都測,不只測傷害那一句。
//
//	increasing the damage each of these weapons inflicts by 50%.
//	It does not improve the chances of hitting a target at a greater distance,
//	nor does it prevent the normal drop-off of damage over range.
func TestHEFRaisesDamageButNotAccuracyOrRange(t *testing.T) {
	const net, wmin, wmax, roll = 60, 40, 100, 50

	// ① 傷害:+50%。
	plain := ResolveShotWithMods(net, wmin, wmax, 1, 0, 0, roll, false, nil, 0, false)
	hef := ResolveShotWithMods(net, wmin, wmax, 1, 0, 0, roll, false, nil, gamedata.DamageMountBonusHEF, false)
	if !plain.Hit || !hef.Hit {
		t.Fatalf("測試前提不成立:兩發都該命中(%v / %v)", plain.Hit, hef.Hit)
	}
	if hef.DamageToStructure <= plain.DamageToStructure {
		t.Errorf("高能聚焦應提高傷害:%d → %d", plain.DamageToStructure, hef.DamageToStructure)
	}

	// ② 命中:不變。逐 roll 比對命中結果——裝與不裝必須**每一顆骰子都一樣**。
	//
	// 比「找一顆會 miss 的骰」更嚴:後者只證明某一顆沒被改變,前者證明整條門檻沒移動。
	// net 用 0(遠距離,命中率低)才會同時出現命中與未命中,測得到兩側。
	misses := 0
	for r := 1; r <= 100; r++ {
		a := ResolveShotWithMods(0, wmin, wmax, 6, 0, 0, r, false, nil, 0, false)
		b := ResolveShotWithMods(0, wmin, wmax, 6, 0, 0, r, false, nil, gamedata.DamageMountBonusHEF, false)
		if a.Hit != b.Hit {
			t.Fatalf("roll=%d 的命中結果被高能聚焦改變了(%v → %v)"+
				"——手冊明說 does not improve the chances of hitting", r, a.Hit, b.Hit)
		}
		if !a.Hit {
			misses++
		}
	}
	if misses == 0 {
		t.Fatal("測試前提不成立:這組參數下每一發都命中,測不到命中率有沒有被動到")
	}

	// ③ 距離衰減:仍然存在。遠距離的 HEF 傷害應低於近距離的 HEF 傷害。
	near := ResolveShotWithMods(net, wmin, wmax, 1, 0, 0, roll, false, nil, gamedata.DamageMountBonusHEF, false)
	far := ResolveShotWithMods(net, wmin, wmax, 6, 0, 0, roll, false, nil, gamedata.DamageMountBonusHEF, false)
	if !far.Hit {
		t.Skip("這個 roll 在遠距離沒打中,換不到可比的樣本")
	}
	if far.DamageToStructure >= near.DamageToStructure {
		t.Errorf("裝了高能聚焦,距離衰減仍應存在:近 %d vs 遠 %d",
			near.DamageToStructure, far.DamageToStructure)
	}
}

// 沒裝就是 0,不會憑空加成。
func TestHEFBonusIsZeroWithoutTheSystem(t *testing.T) {
	if got := hefBonusFor(false); got != 0 {
		t.Errorf("沒裝高能聚焦應為 0,得到 %d", got)
	}
	if got := hefBonusFor(true); got != gamedata.DamageMountBonusHEF {
		t.Errorf("裝了應為 %d,得到 %d", gamedata.DamageMountBonusHEF, got)
	}
}

// 兩條戰鬥路徑(快速結算 / 格子戰術)都認得這個系統。
//
// 少了任一邊,同一艘船在兩種戰鬥裡的傷害會不一樣——而那種不一致很難在遊玩中察覺。
func TestHEFReachesBothCombatPaths(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{
		Name: "測試旗艦", Class: "戰艦", Weapon: "雷射砲", Special: highEnergyFocusName,
	})

	// 快速結算:combatant.hasHEF。
	cs, _ := s.mkPlayerCombatantsIndexed()
	if len(cs) == 0 {
		t.Fatal("測試前提不成立:應有可戰艦艇")
	}
	if !cs[len(cs)-1].hasHEF {
		t.Error("快速結算的戰列應標記高能聚焦")
	}

	// 格子戰術:CombatShip.HEF。
	player, _ := s.StartCombat("測試敵人")
	last := player[len(player)-1]
	if !last.HEF {
		t.Error("格子戰術的艦艇應標記高能聚焦")
	}
	if last.Name != "測試旗艦" {
		t.Fatalf("測試前提不成立:最後一艘應是測試旗艦,得到 %q", last.Name)
	}
}
