package gamedata

import "testing"

func TestOriginalHumanTechValueKnownSliceAdvancedFactories(t *testing.T) {
	state := OriginalStartingValueState{Human: true, Difficulty: 0, InitialSixKnown: true}
	if got, want := OriginalHumanTechValueKnownSlice(TECH_AUTOMATED_FACTORIES, state), 300; got != want {
		t.Fatalf("自動化工廠開局估值=%d，原版公式期望 %d", got, want)
	}
	state.Difficulty = 2
	if got, want := OriginalHumanTechValueKnownSlice(TECH_AUTOMATED_FACTORIES, state), 900; got != want {
		t.Fatalf("難度 2 的自動化工廠估值=%d，原版公式期望 %d", got, want)
	}
}

func TestStartingOriginalApplicationPickConsumesOneRoll(t *testing.T) {
	calls := 0
	topic, tech, ok := StartingOriginalApplicationPick(
		[]ResearchTopic{TOPIC_ADVANCED_CHEMISTRY}, 50,
		OriginalStartingValueState{Human: true, InitialSixKnown: true},
		func(total int) int {
			calls++
			if total <= 0 {
				t.Fatalf("加權總和必須為正，得到 %d", total)
			}
			return 0
		},
	)
	if !ok || topic != TOPIC_ADVANCED_CHEMISTRY || tech != TECH_MERCULITE_MISSILE {
		t.Fatalf("應依科技索引挑中 Merculite Missile，得到 topic=%d tech=%d ok=%v", topic, tech, ok)
	}
	if calls != 1 {
		t.Fatalf("原版應用級抽選只能消耗一次 RNG，得到 %d 次", calls)
	}
}

func TestRollOriginalAITechProfileRawWeights(t *testing.T) {
	var totals []int
	p := RollOriginalAITechProfile([RaceTraitCount]int8{}, 0, 1, func(total int) int {
		totals = append(totals, total)
		return 0
	})
	if p.Raw6 != 0 || p.Raw4 != 0 || p.Raw7 != 0 {
		t.Fatalf("全零種族且 roll=0 應選三組首項，得到 %+v", p)
	}
	want := []int{9, 6, 13}
	if len(totals) != len(want) {
		t.Fatalf("應正好抽三組 raw profile，得到 totals=%v", totals)
	}
	for i := range want {
		if totals[i] != want[i] {
			t.Fatalf("第 %d 組初始權重總和=%d，IDA 期望 %d", i, totals[i], want[i])
		}
	}
}

func TestRollOriginalAITechProfileRaw27ZeroBoostsThirdRaw7Weight(t *testing.T) {
	var totals []int
	RollOriginalAITechProfile([RaceTraitCount]int8{}, 0, 0, func(total int) int {
		totals = append(totals, total)
		return 0
	})
	if got, want := totals[2], 16; got != want {
		t.Fatalf("raw27=0 時七項表總和=%d，IDA 期望 %d", got, want)
	}
}

func TestRollOriginalAITechProfileUsesConvertedTrait6Literally(t *testing.T) {
	collect := func(value int8) []int {
		traits := [RaceTraitCount]int8{}
		traits[TRAIT_SHIP_DEFENSE] = value
		var totals []int
		RollOriginalAITechProfile(traits, 0, 1, func(total int) int {
			totals = append(totals, total)
			return 0
		})
		return totals
	}
	baseline := collect(0)
	for _, runtimeValue := range []int8{25, 50} {
		got := collect(runtimeValue)
		for i := range baseline {
			if got[i] != baseline[i] {
				t.Fatalf("runtime 艦防 %d 不得誤觸原版 20/40 分支：got=%v baseline=%v", runtimeValue, got, baseline)
			}
		}
	}
	if got, want := collect(20), []int{19, 16, 113}; len(got) != len(want) {
		t.Fatalf("特性 6=20 應維持三組抽選，得到 %v", got)
	} else {
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("特性 6=20 的第 %d 組總和=%d，IDA 期望 %d", i, got[i], want[i])
			}
		}
	}
}

func TestRollOriginalAIRaw27UsesRaceTableAndDifficultyClamp(t *testing.T) {
	if got, want := RollOriginalAIRaw27(0, 0, func(int) int { return 0 }), 4; got != want {
		t.Fatalf("Alkari 難度0、roll0 raw27=%d，期望 %d", got, want)
	}
	if got, want := RollOriginalAIRaw27(0, 4, func(int) int { return 0 }), 3; got != want {
		t.Fatalf("Alkari 難度4應夾到第0格，得到 %d，期望 %d", got, want)
	}
}

func TestAIProfileCategoryOverridesFollowRawSwitches(t *testing.T) {
	p := OriginalAITechProfile{Raw4: 1, Raw7: 1, Raw6: 2}
	if got, want := aiProfileCategoryValue(0x19, p), 20; got != want {
		t.Fatalf("raw4=1 的 category 0x19=%d，期望 %d", got, want)
	}
	if got, want := aiProfileCategoryValue(0x23, p), 100; got != want {
		t.Fatalf("raw7=1 的 category 0x23=%d，期望 %d", got, want)
	}
	if got, want := aiProfileCategoryValue(0x27, p), 50; got != want {
		t.Fatalf("raw6=2 的 category 0x27=%d，期望 %d", got, want)
	}
}

func TestOriginalTechValueEarlyCategoryBonusStopsAtTurn150(t *testing.T) {
	tech := Technology(0)
	for i := 1; i < len(TechItemCategory); i++ {
		if TechItemCategory[i] == 0x12 {
			tech = Technology(i)
			break
		}
	}
	if tech == 0 {
		t.Fatal("測試資料應至少有一項 category 0x12 application")
	}
	early := OriginalHumanTechValueKnownSlice(tech, OriginalStartingValueState{Human: true, RelativeTurn: 149})
	late := OriginalHumanTechValueKnownSlice(tech, OriginalStartingValueState{Human: true, RelativeTurn: 150})
	if early != late*2 {
		t.Fatalf("category 0x12 前150回合估值應恰為後期兩倍：early=%d late=%d tech=%d", early, late, tech)
	}
}
