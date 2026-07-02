package gamedata

import "testing"

// TestTechResearchChoicesCount 驗證 researchChoices 列數與 C 版
// research_choices[MAX_RESEARCH_TOPICS] 一致 (tech.h:27, gamestate.h:63)。
func TestTechResearchChoicesCount(t *testing.T) {
	if len(researchChoices) != 83 {
		t.Fatalf("len(researchChoices) = %d, want 83", len(researchChoices))
	}
}

// TestTechResearchChoicesSpotCheck 抽查 researchChoices 是否逐字對齊
// tech.cpp:169-305 的 research_choices[]。
func TestTechResearchChoicesSpotCheck(t *testing.T) {
	cases := []struct {
		name        string
		topic       ResearchTopic
		wantCost    int
		wantAll     bool
		wantChoices []Technology
	}{
		{
			// TOPIC_STARTING_TECH: 遊戲開局即已研究,cost 0、無可選項。
			name:        "TOPIC_STARTING_TECH",
			topic:       TOPIC_STARTING_TECH,
			wantCost:    0,
			wantAll:     false,
			wantChoices: nil,
		},
		{
			// index 6:政體研究,cost 4500,四選一。
			name:        "TOPIC_ADVANCED_GOVERNMENTS",
			topic:       TOPIC_ADVANCED_GOVERNMENTS,
			wantCost:    4500,
			wantAll:     false,
			wantChoices: []Technology{TECH_CONFEDERATION, TECH_FEDERATION, TECH_GALACTIC_UNIFICATION, TECH_IMPERIUM},
		},
		{
			// research_all=1 (true) 的例子:cost 50,四項全解鎖,無需選擇。
			name:        "TOPIC_CHEMISTRY(research_all)",
			topic:       TOPIC_CHEMISTRY,
			wantCost:    50,
			wantAll:     true,
			wantChoices: []Technology{TECH_EXTENDED_FUEL_TANKS, TECH_NUCLEAR_MISSILE, TECH_STANDARD_FUEL_CELLS, TECH_TITANIUM_ARMOR},
		},
		{
			// TOPIC_XENON_TECHNOLOGY: 永遠無法研究,{0} 填充。
			name:        "TOPIC_XENON_TECHNOLOGY",
			topic:       TOPIC_XENON_TECHNOLOGY,
			wantCost:    0,
			wantAll:     false,
			wantChoices: nil,
		},
		{
			// 末項 TOPIC_HYPER_SOCIOLOGY: cost 25000,單一科技。
			name:        "TOPIC_HYPER_SOCIOLOGY",
			topic:       TOPIC_HYPER_SOCIOLOGY,
			wantCost:    25000,
			wantAll:     false,
			wantChoices: []Technology{TECH_HYPER_SOCIOLOGY},
		},
		{
			// TOPIC_ENGINEERING: 開局已研究的工程學,cost 50、research_all=1。
			name:        "TOPIC_ENGINEERING",
			topic:       TOPIC_ENGINEERING,
			wantCost:    50,
			wantAll:     true,
			wantChoices: []Technology{TECH_COLONY_BASE, TECH_MARINE_BARRACKS, TECH_STAR_BASE},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ResearchChoiceFor(c.topic)
			if got.Cost != c.wantCost {
				t.Errorf("Cost = %d, want %d", got.Cost, c.wantCost)
			}
			if got.ResearchAll != c.wantAll {
				t.Errorf("ResearchAll = %v, want %v", got.ResearchAll, c.wantAll)
			}
			if len(got.Choices) != len(c.wantChoices) {
				t.Fatalf("len(Choices) = %d, want %d (%v vs %v)", len(got.Choices), len(c.wantChoices), got.Choices, c.wantChoices)
			}
			for i := range c.wantChoices {
				if got.Choices[i] != c.wantChoices[i] {
					t.Errorf("Choices[%d] = %v, want %v", i, got.Choices[i], c.wantChoices[i])
				}
			}
		})
	}
}

// TestTechEnumSpotValues 抽查 Technology 常數整數值,確認與 C 版
// enum Technology (gamestate.h:326 起) 一致 (由 enums.go 生成,這裡是回歸保護)。
func TestTechEnumSpotValues(t *testing.T) {
	cases := []struct {
		name string
		got  Technology
		want int
	}{
		{"TECH_NONE", TECH_NONE, 0},
		{"TECH_CONFEDERATION", TECH_CONFEDERATION, 42},
		{"TECH_HYPER_SOCIOLOGY", TECH_HYPER_SOCIOLOGY, 211},
	}
	for _, c := range cases {
		if int(c.got) != c.want {
			t.Errorf("%s = %d, want %d", c.name, int(c.got), c.want)
		}
	}
}

// TestTechTreeAreas 驗證 techtree 的領域數與各領域 topic 數,對齊
// tech.cpp:69-167 的 techtree[MAX_RESEARCH_AREAS][MAX_AREA_TOPICS]
// (陣列中用來補滿 MAX_AREA_TOPICS 的 TOPIC_STARTING_TECH 填充值不計入)。
func TestTechTreeAreas(t *testing.T) {
	areas := TechTree()
	if len(areas) != 8 {
		t.Fatalf("len(TechTree()) = %d, want 8", len(areas))
	}

	// index 0 = Biology (RESEARCH_BIOLOGY),9 個 topic,末項為 TOPIC_HYPER_BIOLOGY。
	biology := areas[0]
	if len(biology) != 9 {
		t.Fatalf("len(areas[0]) (Biology) = %d, want 9", len(biology))
	}
	if biology[0] != TOPIC_ASTRO_BIOLOGY {
		t.Errorf("areas[0][0] = %v, want TOPIC_ASTRO_BIOLOGY", biology[0])
	}
	if biology[len(biology)-1] != TOPIC_HYPER_BIOLOGY {
		t.Errorf("areas[0][last] = %v, want TOPIC_HYPER_BIOLOGY", biology[len(biology)-1])
	}

	// index 7 = Sociology (RESEARCH_SOCIOLOGY),7 個 topic。
	sociology := areas[7]
	if len(sociology) != 7 {
		t.Fatalf("len(areas[7]) (Sociology) = %d, want 7", len(sociology))
	}
	if sociology[0] != TOPIC_MILITARY_TACTICS {
		t.Errorf("areas[7][0] = %v, want TOPIC_MILITARY_TACTICS", sociology[0])
	}
	if sociology[len(sociology)-1] != TOPIC_HYPER_SOCIOLOGY {
		t.Errorf("areas[7][last] = %v, want TOPIC_HYPER_SOCIOLOGY", sociology[len(sociology)-1])
	}
}
