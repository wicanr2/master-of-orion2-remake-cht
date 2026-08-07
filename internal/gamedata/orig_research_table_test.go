package gamedata

import "testing"

// 這一組測試是「二手轉寫 vs 一手執行檔」的對照。
// techtree.go 的 83 列來自 openorion2;OrigTopicCost 來自原版執行檔的 word_17D90C。
// 兩邊各自獨立,對得上才代表轉寫是對的。

// 74 個主題的成本必須逐字相同。剩下 9 個各有解釋,由下面兩支測試分別釘住。
func TestResearchCostsMatchTheOriginalExecutable(t *testing.T) {
	matched := 0
	for i := range researchChoices {
		topic := ResearchTopic(i)
		if IsHyperAdvancedTopic(topic) || topic == TOPIC_XENON_TECHNOLOGY {
			continue // 這 9 個由專屬測試處理
		}
		if researchChoices[i].Cost != OrigTopicCost[i] {
			t.Errorf("主題 %d:轉寫值 %d,原版執行檔 %d",
				i, researchChoices[i].Cost, OrigTopicCost[i])
			continue
		}
		matched++
	}
	// 正對照:少了這個,「全部 continue 掉」也會讓上面通過。
	if matched != 74 {
		t.Fatalf("應有 74 個主題逐字相同,實際比對到 %d 個", matched)
	}
}

// 每個主題的可選科技必須與原版的科技表一致。
//
// 原版沒有「主題 → 科技」表(那四個欄位在執行檔裡是 0,執行期才填);
// 一手的方向是反過來的:科技表的每一筆說自己屬於哪個主題。
// 這裡拿轉寫值去對那個方向,199 條全中才算通過。
func TestResearchChoicesMatchTheOriginalTechTable(t *testing.T) {
	total := 0
	for i := range researchChoices {
		topic := ResearchTopic(i)
		for _, tech := range researchChoices[i].Choices {
			got, ok := OrigTechTopic(tech)
			if !ok {
				t.Errorf("主題 %d 列了科技 %d,但原版科技表裡查不到它", i, tech)
				continue
			}
			if got != topic {
				t.Errorf("科技 %d:轉寫把它放在主題 %d,原版科技表說它屬於主題 %d",
					tech, i, got)
				continue
			}
			total++
		}
	}
	if total != 199 {
		t.Fatalf("應有 199 條科技歸屬對上,實際 %d 條", total)
	}
}

// Hyper-Advanced 的成本差異是**真的版本差異**,不是轉寫錯誤。
// 1.3 走執行檔的 15000、1.5 走 25000,而 techtree.go 的硬編值是 1.5 的那個。
func TestHyperAdvancedCostIsAVersionDifferenceNotATranscriptionError(t *testing.T) {
	n := 0
	for i := range researchChoices {
		topic := ResearchTopic(i)
		if !IsHyperAdvancedTopic(topic) {
			continue
		}
		n++
		if OrigTopicCost[i] != Profile13().HyperAdvancedLevel1Cost {
			t.Errorf("主題 %d:1.3 profile 是 %d,但 1.3 執行檔的表是 %d",
				i, Profile13().HyperAdvancedLevel1Cost, OrigTopicCost[i])
		}
		if researchChoices[i].Cost != Profile15().HyperAdvancedLevel1Cost {
			t.Errorf("主題 %d:techtree 硬編 %d,應等於 1.5 profile 的 %d",
				i, researchChoices[i].Cost, Profile15().HyperAdvancedLevel1Cost)
		}
	}
	if n != 8 {
		t.Fatalf("Hyper-Advanced 應有 8 個主題,找到 %d 個", n)
	}
	if Profile13().HyperAdvancedLevel1Cost == Profile15().HyperAdvancedLevel1Cost {
		t.Fatal("兩版的值相同,這支測試就什麼都沒驗到")
	}
}

// 主題 74(XENON)在 techtree.go 裡是空的,而原版的表寫著 15000 與 8 個科技。
// 空得對——原版用**自環**表示它永遠不會被解鎖。
//
// ⚠ 這條結構論證**只對主題 74 成立**。主題 0 的 next 也是 0,但 0 同時是「鏈到此為止」
// 的哨符(8 個 Hyper-Advanced 主題的 next 都是 0),所以「next == 0」在主題 0 身上
// 分不出是自環還是終點。**分不出就不當證據用**——主題 0 不可研究的理由是別的
// (它是開局科技的容器,`Init_Player_Tech_` 直接發,不經研究選單)。
func TestXenonTopicIsEncodedAsAnUnreachableSelfLoop(t *testing.T) {
	const xenon = int(TOPIC_XENON_TECHNOLOGY)
	if !OrigTopicIsSelfLoop(TOPIC_XENON_TECHNOLOGY) {
		t.Errorf("主題 %d 應是自環(next 指向自己),實際 next = %d", xenon, OrigTopicNext[xenon])
	}
	// 自環之外沒有第二條入口:沒有任何別的主題把 74 當 next。
	for i, next := range OrigTopicNext {
		if i != xenon && next == xenon {
			t.Errorf("主題 %d 的 next 指向 %d ——那主題 74 就不是永遠解不開的了", i, xenon)
		}
	}
	for _, topic := range []ResearchTopic{TOPIC_STARTING_TECH, TOPIC_XENON_TECHNOLOGY} {
		if len(researchChoices[int(topic)].Choices) != 0 {
			t.Errorf("主題 %d 永遠不會出現在研究選單上,不該列科技", topic)
		}
	}
	// 正對照:一般主題不該是自環,否則上面那條規則就沒有鑑別力。
	if OrigTopicIsSelfLoop(TOPIC_ENGINEERING) {
		t.Error("TOPIC_ENGINEERING 不該是自環")
	}
	// 正對照:next == 0 在樹頂是「終點」而不是「指回主題 0」,兩者不可混用。
	if OrigTopicNext[int(TOPIC_HYPER_BIOLOGY)] != 0 {
		t.Error("Hyper-Advanced 是樹頂,next 應為哨符 0")
	}
}

// 兩個永遠解不開的主題底下各有哪些科技,記成資料。
func TestUnreachableTopicTechListsMatchTheOriginal(t *testing.T) {
	for _, tc := range []struct {
		topic ResearchTopic
		want  []Technology
	}{
		{TOPIC_STARTING_TECH, OrigStartingTechs},
		{TOPIC_XENON_TECHNOLOGY, OrigXenonTechs},
	} {
		for _, tech := range tc.want {
			if tech == TECH_NONE {
				continue // 原版把 TECH_NONE 也列在開局那組裡,不必反查
			}
			got, ok := OrigTechTopic(tech)
			if !ok || got != tc.topic {
				t.Errorf("科技 %d 應屬於主題 %d,原版科技表說 %d(找到=%v)",
					tech, tc.topic, got, ok)
			}
		}
	}
	if len(OrigXenonTechs) != 8 {
		t.Errorf("安塔蘭專屬科技應有 8 個,得到 %d 個", len(OrigXenonTechs))
	}
}

// 脈衝步槍是地面戰步槍加成的 +0 基準點——理由是它在開局科技裡,人人第一回合就有。
// 這條把 orig_research_table 與 ground.go 綁在一起:哪天有人把基準點改掉,這裡會紅。
func TestPulseRifleIsTheGroundRifleBaselineBecauseEveryoneStartsWithIt(t *testing.T) {
	found := false
	for _, tech := range OrigStartingTechs {
		if tech == TECH_PULSE_RIFLE {
			found = true
		}
	}
	if !found {
		t.Fatal("脈衝步槍應在開局科技裡")
	}
	if GroundRifleTechBonus(TECH_PULSE_RIFLE) != 0 {
		t.Errorf("開局就有的步槍必須是 +0 基準點,得到 %d",
			GroundRifleTechBonus(TECH_PULSE_RIFLE))
	}
}

// ★ 兩種編碼指到同一棵樹。
//
// openorion2 的 `techtree[8][14]` 把研究樹寫成「8 個領域各一串主題」;
// 原版執行檔把同一棵樹寫成 `next` 鏈(每個主題只有一個後繼)。
// 兩邊互不知情,所以「領域串裡相鄰的兩個主題」必須正好等於「next 關係」。
//
// 這條過了,就等於同時驗證了領域表與 next 表——而且說明了原版的研究是**線性推進**:
// 每個領域同時只有一個主題可研究,完成它才輪到下一個。
func TestAreaSequencesAgreeWithTheOriginalNextChain(t *testing.T) {
	links := 0
	for a, area := range techtree {
		for i := 0; i+1 < len(area); i++ {
			cur, next := int(area[i]), int(area[i+1])
			if OrigTopicNext[cur] != next {
				t.Errorf("領域 %d:領域表說主題 %d 之後是 %d,原版 next 表說是 %d",
					a, cur, next, OrigTopicNext[cur])
				continue
			}
			links++
		}
	}
	if links == 0 {
		t.Fatal("一條都沒比到——領域表是空的?")
	}
	t.Logf("兩種編碼吻合 %d 條銜接關係", links)
}

// 每個領域的最後一個主題,其 next 應是哨符 0(鏈到頂)。
func TestAreaLastTopicEndsTheChain(t *testing.T) {
	for a, area := range techtree {
		if len(area) == 0 {
			continue
		}
		last := int(area[len(area)-1])
		if OrigTopicNext[last] != 0 {
			t.Errorf("領域 %d 的最後一個主題 %d 的 next 是 %d,應為哨符 0",
				a, last, OrigTopicNext[last])
		}
	}
}
