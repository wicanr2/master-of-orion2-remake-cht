package gamedata

import "testing"

func TestMiniaturizationSpaceCategoryLadders(t *testing.T) {
	if got := WeaponSpaceAtMiniLevelForCategory(100, 2, MiniSpaceGeneral); got != 65 {
		t.Fatalf("general level2=%d want 65", got)
	}
	if got := WeaponSpaceAtMiniLevelForCategory(100, 2, MiniSpaceTorpedoOrSpecial); got != 70 {
		t.Fatalf("torpedo/special level2=%d want 70", got)
	}
	if got := WeaponSpaceAtMiniLevelForCategory(100, 9, MiniSpaceFixed); got != 100 {
		t.Fatalf("fixed level9=%d want 100", got)
	}
	if got := WeaponSpaceAtMiniLevelForCategory(-12, 1, MiniSpaceTorpedoOrSpecial); got != -9 {
		t.Fatalf("負佔格 category1 level1=%d want -9", got)
	}
}

func TestMiniaturizationSpaceCategoryForOriginalTech(t *testing.T) {
	cases := []struct {
		tech Technology
		want MiniaturizationSpaceCategory
	}{
		{TECH_LASER_CANNON, MiniSpaceGeneral},
		{TECH_NUCLEAR_MISSILE, MiniSpaceGeneral},
		{TECH_NUCLEAR_BOMB, MiniSpaceGeneral},
		{TECH_PROTON_TORPEDOES, MiniSpaceTorpedoOrSpecial},
		{TECH_SPATIAL_COMPRESSOR, MiniSpaceTorpedoOrSpecial},
		{TECH_FIGHTER_BAYS, MiniSpaceFixed},
		{TECH_AUGMENTED_ENGINES, MiniSpaceFixed},
		{TECH_TIME_WARP_FACILITATOR, MiniSpaceTorpedoOrSpecial},
	}
	for _, c := range cases {
		if got := MiniaturizationSpaceCategoryForTech(c.tech); got != c.want {
			t.Errorf("tech %d category=%d want %d", c.tech, got, c.want)
		}
	}
}

func TestWeaponMiniaturizationAddsHyperRepeatLevel(t *testing.T) {
	tech := TECH_LASER_CANNON
	topic, _ := OrigTechTopic(tech)
	completed := map[ResearchTopic]bool{}
	cur := int(topic)
	for {
		next := OrigTopicNext[cur]
		if next == 0 || next == cur {
			t.Fatal("測試科技鏈未抵達 Hyper topic")
		}
		nextTopic := ResearchTopic(next)
		if IsHyperAdvancedTopic(nextTopic) {
			completed[nextTopic] = true
			got := WeaponMiniaturizationLevelWithHyper(tech, completed, map[ResearchTopic]int{nextTopic: 3})
			if got < 3 {
				t.Fatalf("Hyper 3 級應完整加進微型化，得到 %d", got)
			}
			withoutRepeat := WeaponMiniaturizationLevelWithHyper(tech, completed, map[ResearchTopic]int{})
			if got-withoutRepeat != 3 {
				t.Fatalf("Hyper 重複等級差=%d want 3", got-withoutRepeat)
			}
			break
		}
		completed[nextTopic] = true
		cur = next
	}
}

func TestAvailableTopicsKeepsCompletedHyperResearchable(t *testing.T) {
	completed := map[ResearchTopic]bool{}
	for _, area := range TechTree() {
		for _, topic := range area {
			completed[topic] = true
		}
	}
	available := AvailableTopics(completed)
	if len(available) != 8 {
		t.Fatalf("八個領域完成後仍應各提供一個 Hyper topic，得到 %v", available)
	}
	for _, topic := range available {
		if !IsHyperAdvancedTopic(topic) {
			t.Fatalf("完成整條科技樹後不應回到一般主題，得到 %d", topic)
		}
	}
}
