package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 艦員經驗的 **BD(防禦)** 那一欄先前沒加過。
//
// `mkPlayerCombatants` 的註解寫著「老手打得準**也閃得掉**」,而程式只做了 BA(攻擊)那一欄
// ——`def` 只有艦體值。手冊 p.121 的兩欄是分開的兩個加成。
func TestCrewExperienceRaisesDefenceNotJustOffence(t *testing.T) {
	s := NewDemoSession()
	if len(s.Fleet().Ships) == 0 {
		t.Fatal("測試前提不成立:示範對局應該有船")
	}
	for i := range s.Fleet().Ships {
		s.Fleet().Ships[i].CrewXP = 0
	}
	green := s.mkPlayerCombatants()
	if len(green) == 0 {
		t.Fatal("測試前提不成立:應有參戰艦")
	}

	// 把艦員練到高階(用 gamedata 的門檻,不自己編數字)。
	for i := range s.Fleet().Ships {
		s.Fleet().Ships[i].CrewXP = gamedata.CrewXPForLevel(gamedata.CrewLevelCount-2, s.RaceWarlord)
	}
	veteran := s.mkPlayerCombatants()

	if veteran[0].atk <= green[0].atk {
		t.Errorf("老手的攻擊應更高:%d → %d", green[0].atk, veteran[0].atk)
	}
	if veteran[0].def <= green[0].def {
		t.Errorf("老手的**防禦**也該更高(手冊 BD 欄),得到 %d → %d", green[0].def, veteran[0].def)
	}
}

// 飛彈閃避:老手比新兵難打中。
//
// 先前 `defenderEvasionBonus` 恆傳 0——那句擋門理由對 ECM 干擾器仍成立,
// 但艦員經驗與舵手技能兩項現在算得出來。
func TestCrewExperienceRaisesMissileEvasion(t *testing.T) {
	s := NewDemoSession()
	for i := range s.Fleet().Ships {
		s.Fleet().Ships[i].CrewXP = 0
	}
	green := s.mkPlayerCombatants()
	for i := range s.Fleet().Ships {
		s.Fleet().Ships[i].CrewXP = gamedata.CrewXPForLevel(gamedata.CrewLevelCount-2, s.RaceWarlord)
	}
	veteran := s.mkPlayerCombatants()

	if green[0].missileEvasion != 0 {
		t.Errorf("新兵的飛彈閃避應為 0(手冊 ME 欄綠色 = 0),得到 %d", green[0].missileEvasion)
	}
	if veteran[0].missileEvasion <= 0 {
		t.Errorf("老手應有飛彈閃避加成,得到 %d", veteran[0].missileEvasion)
	}
}

// 舵手技能貢獻閃避,而且**只算艦艇軍官**——殖民地領袖不會坐在艦橋上。
func TestHelmsmanEvasionCountsShipOfficersOnly(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = nil
	if got := s.helmsmanEvasionBonus(); got != 0 {
		t.Errorf("沒有領袖時應為 0,得到 %d", got)
	}

	// 殖民地領袖有舵手技能也不算(資料上不該出現,但規則要擋得住)。
	s.Leaders = []Leader{{Name: "陸上舵手", Skill: "舵手", Level: 5, Ship: false, Tier: 2,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_HELMSMAN), Tier: 2}}}}
	if got := s.helmsmanEvasionBonus(); got != 0 {
		t.Errorf("殖民地領袖的舵手技能不該算,得到 %d", got)
	}

	// 艦艇軍官才算。
	s.Leaders = []Leader{{Name: "艦橋舵手", Skill: "舵手", Level: 5, Ship: true, Tier: 2,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_HELMSMAN), Tier: 2}}}}
	if got := s.helmsmanEvasionBonus(); got <= 0 {
		t.Errorf("艦艇軍官的舵手技能應貢獻閃避,得到 %d", got)
	}
}

// 舵手加成是技能值的**一半**(手冊:「Half bonus of the Helmsman value」)。
func TestHelmsmanEvasionIsHalfTheSkillValue(t *testing.T) {
	const tier, level = 2, 5
	want := gamedata.MissileHelmsmanEvasionBonus(
		gamedata.LeaderSkillBonus(int(gamedata.SKILL_HELMSMAN), tier, leaderDisplayLevelToExpLevel(level)))
	s := NewDemoSession()
	s.Leaders = []Leader{{Name: "舵手", Skill: "舵手", Level: level, Ship: true, Tier: tier,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_HELMSMAN), Tier: tier}}}}
	if got := s.helmsmanEvasionBonus(); got != want {
		t.Errorf("應為技能值的一半 %d,得到 %d", want, got)
	}
}

// 多位舵手取最佳,不加總——手冊只在 Megawealth/Researcher 明說累加。
func TestHelmsmanEvasionTakesTheBestNotTheSum(t *testing.T) {
	s := NewDemoSession()
	one := Leader{Name: "甲", Skill: "舵手", Level: 5, Ship: true, Tier: 2,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_HELMSMAN), Tier: 2}}}
	s.Leaders = []Leader{one}
	single := s.helmsmanEvasionBonus()

	weak := Leader{Name: "乙", Skill: "舵手", Level: 1, Ship: true, Tier: 1,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_HELMSMAN), Tier: 1}}}
	s.Leaders = []Leader{one, weak}
	if got := s.helmsmanEvasionBonus(); got != single {
		t.Errorf("多一位較弱的舵手不該改變結果(取最佳):%d → %d", single, got)
	}
}
