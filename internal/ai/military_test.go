package ai

import "testing"

// TestDecideBuildPriority 手算對照表【設計啟發式驗證,非原版數值】。
func TestDecideBuildPriority(t *testing.T) {
	cases := []struct {
		name       string
		p          Profile
		threatened bool
		canColony  bool
		infraDone  bool
		want       BuildPriority
	}{
		// 1. 受威脅:好戰性格造戰艦。
		{"受威脅_好戰造艦", ProfileAggressive, true, false, true, BuildWarships},
		// 2. 受威脅:其餘性格(科學)造防禦。
		{"受威脅_科學造防禦", ProfileScientific, true, false, true, BuildDefenses},
		// 2b. 受威脅:平衡性格同樣造防禦(非好戰)。
		{"受威脅_平衡造防禦", ProfileBalanced, true, false, true, BuildDefenses},
		// 3. 基建未完成:優先殖民建設,即使是好戰性格。
		{"基建未完成_好戰仍先建設", ProfileAggressive, false, false, false, BuildColonyInfrastructure},
		{"基建未完成_擴張仍先建設", ProfileExpansionist, false, true, false, BuildColonyInfrastructure},
		// 3b. 受威脅優先度高於基建未完成:兩者同時成立時以威脅為準。
		{"受威脅優先於基建未完成", ProfileAggressive, true, false, false, BuildWarships},
		// 4. 有可殖民目標且擴張/平衡性格 → 造殖民船。
		{"擴張性格有目標造殖民船", ProfileExpansionist, false, true, true, BuildColonyShip},
		{"平衡性格有目標造殖民船", ProfileBalanced, false, true, true, BuildColonyShip},
		// 4b. 有可殖民目標但好戰/科學性格不視為擴張優先,落到預設分支。
		{"好戰性格有目標仍造艦", ProfileAggressive, false, true, true, BuildWarships},
		{"科學性格有目標仍造建設", ProfileScientific, false, true, true, BuildColonyInfrastructure},
		// 5. 預設分支:無威脅、基建已完成、無可殖民目標。
		{"預設_好戰造艦", ProfileAggressive, false, false, true, BuildWarships},
		{"預設_科學造建設", ProfileScientific, false, false, true, BuildColonyInfrastructure},
		{"預設_平衡造建設", ProfileBalanced, false, false, true, BuildColonyInfrastructure},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := DecideBuildPriority(c.p, c.threatened, c.canColony, c.infraDone)
			if got != c.want {
				t.Errorf("DecideBuildPriority(%s, threatened=%v, canColony=%v, infraDone=%v) = %v,預期 %v",
					c.p.Name, c.threatened, c.canColony, c.infraDone, got, c.want)
			}
		})
	}
}
