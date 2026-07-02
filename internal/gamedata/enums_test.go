package gamedata

import "testing"

// 抽查生成的枚舉值是否對齊 openorion2 gamestate.h(生成正確性回歸)。
func TestEnumSpotValues(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"TECH_NONE", int(TECH_NONE), 0},
		{"TECH_GAIA_TRANSFORMATION", int(TECH_GAIA_TRANSFORMATION), 74},
		{"TECH_HYPER_SOCIOLOGY", int(TECH_HYPER_SOCIOLOGY), 211}, // 末項 → 共 212
		{"RESEARCH_SOCIOLOGY", int(RESEARCH_SOCIOLOGY), 7},
		{"GAIA", int(GAIA), 9},
		{"ULTRA_RICH", int(ULTRA_RICH), 4},
		{"BlackHole", int(BlackHole), 6},
		{"BUILDING_GAIA_TRANSFORMATION", int(BUILDING_GAIA_TRANSFORMATION), 17},
		{"SKILL_ASSASSIN", int(SKILL_ASSASSIN), 0},
		{"SKILL_ENGINEER", int(SKILL_ENGINEER), 16},   // CAPTAIN_SKILLS_TYPE
		{"SKILL_ENVIRONMENTALIST", int(SKILL_ENVIRONMENTALIST), 32}, // ADMIN_SKILLS_TYPE
		{"SKILL_TACTICS", int(SKILL_TACTICS), 40},
		{"ANDROID", int(ANDROID), 8},
		{"ORION_SPECIAL", int(ORION_SPECIAL), 11},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d,預期 %d", c.name, c.got, c.want)
		}
	}
}
