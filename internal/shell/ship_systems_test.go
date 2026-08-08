package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// mkLast 造一艘指定系統的戰艦,回傳它在快速結算戰列裡的那一格。
func mkLast(t *testing.T, special string) combatant {
	t.Helper()
	s := NewDemoSession()
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{
		Name: "測試艦", Class: "戰艦", Weapon: "雷射砲", Shield: "第一級護盾", Special: special,
	})
	cs, _ := s.mkPlayerCombatantsIndexed()
	if len(cs) == 0 {
		t.Fatal("測試前提不成立:應有可戰艦艇")
	}
	return cs[len(cs)-1]
}

// ⚠ 第 69 項的缺失:慣性穩定器手冊給了**兩個**加成,當時只接了飛彈閃避那一半。
//
// 這條測試同時釘住兩半,免得日後又只補一邊。
func TestInertialStabilizerGivesBothBonuses(t *testing.T) {
	plain, stab := mkLast(t, ""), mkLast(t, "慣性穩定器")
	if got := stab.missileEvasion - plain.missileEvasion; got != gamedata.MissileInertialStabilizer {
		t.Errorf("飛彈閃避應 +%d,得到 %+d", gamedata.MissileInertialStabilizer, got)
	}
	if got := stab.def - plain.def; got != gamedata.ShipInertialStabilizerBeamDefense {
		t.Errorf("光束閃避應 +%d,得到 %+d(這正是第 69 項漏掉的那一半)",
			gamedata.ShipInertialStabilizerBeamDefense, got)
	}
	// 慣性抵消器兩項都更高。
	null := mkLast(t, "慣性抵消器")
	if null.def <= stab.def || null.missileEvasion <= stab.missileEvasion {
		t.Errorf("慣性抵消器兩項都應高於穩定器:光束 %d vs %d、飛彈 %d vs %d",
			null.def, stab.def, null.missileEvasion, stab.missileEvasion)
	}
}

// 戰鬥掃描器在**戰鬥裡**:+50 命中,而且不動閃避與結構。
//
// ⚠ 這條測試原本叫 TestBattleScannerRaisesOnlyAccuracy——那個「Only」是錯的。
// 手冊那一段有兩句話,第二句是「掃描範圍 +2 parsec(戰鬥之外)」,第 72 項才補上。
// **測試名稱裡的「只」是一種斷言**,而這一條斷的是我當時的理解,不是手冊。
func TestBattleScannerRaisesAccuracyInCombat(t *testing.T) {
	plain, scan := mkLast(t, ""), mkLast(t, "戰鬥掃描器")
	if got := scan.atk - plain.atk; got != gamedata.ShipBattleScannerBeamOffense {
		t.Errorf("命中應 +%d,得到 %+d", gamedata.ShipBattleScannerBeamOffense, got)
	}
	if scan.def != plain.def {
		t.Errorf("戰鬥掃描器不該動閃避:%d vs %d", scan.def, plain.def)
	}
	if scan.hp != plain.hp {
		t.Errorf("戰鬥掃描器不該動結構:%d vs %d", scan.hp, plain.hp)
	}
}

// 強化船體:結構三倍(手冊「triples」)。
func TestReinforcedHullTriplesStructure(t *testing.T) {
	plain, hull := mkLast(t, ""), mkLast(t, "強化船體")
	if want := plain.hp * gamedata.ShipReinforcedHullStructurePercent / 100; hull.hp != want {
		t.Errorf("結構應為 %d(三倍),得到 %d", want, hull.hp)
	}
	if hull.atk != plain.atk || hull.def != plain.def {
		t.Error("強化船體不該動攻防")
	}
}

// 多相護盾:吸收量 +50%。
func TestMultiPhasedShieldsAbsorbFiftyPercentMore(t *testing.T) {
	plain, mp := mkLast(t, ""), mkLast(t, "多相護盾")
	if plain.shield <= 0 {
		t.Fatalf("測試前提不成立:基準艦應有護盾(得到 %d)——測試裝的是第一級護盾", plain.shield)
	}
	if want := plain.shield * gamedata.ShipMultiPhasedShieldPercent / 100; mp.shield != want {
		t.Errorf("護盾吸收應為 %d(+50%%),得到 %d", want, mp.shield)
	}
}

// 四個效果在格子戰術那一側也要有——兩條路只接一邊會安靜地不一致。
func TestShipSystemsReachTheTacticalPath(t *testing.T) {
	mk := func(special string) CombatShip {
		s := NewDemoSession()
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{
			Name: "測試艦", Class: "戰艦", Weapon: "雷射砲", Shield: "第一級護盾", Special: special,
		})
		player, _ := s.StartCombat("測試敵人")
		return player[len(player)-1]
	}
	plain := mk("")
	if got := mk("戰鬥掃描器"); got.Attack-plain.Attack != gamedata.ShipBattleScannerBeamOffense {
		t.Errorf("格子戰術的命中加成應為 %d,得到 %+d",
			gamedata.ShipBattleScannerBeamOffense, got.Attack-plain.Attack)
	}
	if got := mk("慣性穩定器"); got.Defense-plain.Defense != gamedata.ShipInertialStabilizerBeamDefense {
		t.Errorf("格子戰術的閃避加成應為 %d,得到 %+d",
			gamedata.ShipInertialStabilizerBeamDefense, got.Defense-plain.Defense)
	}
	if got := mk("強化船體"); got.HP != plain.HP*3 {
		t.Errorf("格子戰術的結構應為三倍:%d vs %d", got.HP, plain.HP*3)
	}
	if got := mk("多相護盾"); got.ShieldReduction != plain.ShieldReduction*150/100 {
		t.Errorf("格子戰術的護盾應 +50%%:%d vs %d", got.ShieldReduction, plain.ShieldReduction*150/100)
	}
}

// 偵察實驗室:研究點數依艦體等級,手冊整張表都列出來了。
func TestScoutLabResearchFollowsTheHullTable(t *testing.T) {
	want := map[string]int{"巡防艦": 1, "驅逐艦": 2, "巡洋艦": 4, "戰艦": 8, "泰坦": 16, "末日之星": 32}
	for class, n := range want {
		s := NewDemoSession()
		before := s.FleetResearchPoints()
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{Name: "研究艦", Class: class, Special: "偵察實驗室"})
		if got := s.FleetResearchPoints() - before; got != n {
			t.Errorf("%s 的偵察實驗室應產生 %d 研究點,得到 %d", class, n, got)
		}
	}
	// 沒裝就沒有。
	s := NewDemoSession()
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{Name: "普通艦", Class: "戰艦"})
	if got := s.FleetResearchPoints(); got != 0 {
		t.Errorf("沒裝偵察實驗室不該產生研究點,得到 %d", got)
	}
}

// 艦隊研究真的併進回合結算的研究投入。
func TestFleetResearchReachesTheResearchPhase(t *testing.T) {
	base := NewDemoSession()
	base.DisableEvents = true
	base.EndTurn()
	plain := base.LastPlayerOutput.TotalResearch

	lab := NewDemoSession()
	lab.DisableEvents = true
	lab.Fleet().Ships = append(lab.Fleet().Ships, Ship{Name: "研究艦", Class: "末日之星", Special: "偵察實驗室"})
	lab.EndTurn()

	class, _ := shipClassFromName("末日之星")
	if want := plain + gamedata.ShipScoutLabResearch(class); lab.LastPlayerOutput.TotalResearch != want {
		t.Errorf("回合結算的總研究應為 %d(基準 %d + 末日之星 %d),得到 %d",
			want, plain, gamedata.ShipScoutLabResearch(class), lab.LastPlayerOutput.TotalResearch)
	}
}

// 五個新元件的主題全部對得上執行檔。
func TestNewShipSystemsHaveRealTopics(t *testing.T) {
	want := map[string]gamedata.Technology{
		"偵察實驗室": gamedata.TECH_SCOUT_LAB,
		"強化船體":  gamedata.TECH_REINFORCED_HULL,
		"多相護盾":  gamedata.TECH_MULTIPHASED_SHIELDS,
		"戰鬥掃描器": gamedata.TECH_BATTLE_SCANNER,
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
