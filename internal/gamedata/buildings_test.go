package gamedata

import "testing"

// TestBuildingsCount 驗證手冊全表 35 建築 + 5 衛星 = 40 項,一項不多不少。
func TestBuildingsCount(t *testing.T) {
	if got := len(Buildings); got != 40 {
		t.Fatalf("Buildings 應有 40 筆(35 建築 + 5 衛星),got %d", got)
	}
}

// TestBuildingsNoDuplicateNameZH 驗證中文名無重複(建造選單/存檔以中文名為 key,重複會互相覆蓋)。
func TestBuildingsNoDuplicateNameZH(t *testing.T) {
	seen := make(map[string]bool, len(Buildings))
	for _, b := range Buildings {
		if seen[b.NameZH] {
			t.Fatalf("重複的中文建築名:%s", b.NameZH)
		}
		seen[b.NameZH] = true
	}
}

// TestBuildingsPrereqTopicLegal 驗證每筆 PrereqTopic 都落在 researchChoices 有定義的範圍內
// (可安全傳入 ResearchChoiceFor 取得 RP 花費,不會索引越界)。
func TestBuildingsPrereqTopicLegal(t *testing.T) {
	for _, b := range Buildings {
		if int(b.PrereqTopic) < 0 || int(b.PrereqTopic) >= len(researchChoices) {
			t.Fatalf("%s (%s) 的 PrereqTopic=%v 超出合法範圍(0..%d)", b.NameZH, b.NameEN, b.PrereqTopic, len(researchChoices)-1)
		}
	}
}

// TestBuildingByNameZH 驗證中文名查找。
func TestBuildingByNameZH(t *testing.T) {
	b, ok := BuildingByNameZH("研究實驗室")
	if !ok {
		t.Fatal("應找到研究實驗室")
	}
	if b.NameEN != "Research Laboratory" {
		t.Fatalf("研究實驗室英文名應為 Research Laboratory,got %s", b.NameEN)
	}
	if _, ok := BuildingByNameZH("不存在的建築"); ok {
		t.Fatal("不存在的建築名不應找到")
	}
}

// TestAvailableBuildingsEmpty 驗證空科技(nil/空 map)不會 panic,且只回傳前置已滿足的項目——
// 本表 40 項全部要求特定研究主題,空科技下沒有任何一項的前置能被視為滿足,故應回傳空清單。
func TestAvailableBuildingsEmpty(t *testing.T) {
	if got := AvailableBuildings(nil); len(got) != 0 {
		t.Fatalf("nil completedTopics 應回傳空清單,got %d 筆", len(got))
	}
	if got := AvailableBuildings(map[ResearchTopic]bool{}); len(got) != 0 {
		t.Fatalf("空 map 應回傳空清單,got %d 筆", len(got))
	}
}

// TestAvailableBuildingsGated 驗證給定已完成研究主題後,只回傳對應可建項目。
func TestAvailableBuildingsGated(t *testing.T) {
	// 起始文明已知 Engineering:應解鎖 Marine Barracks(海軍陸戰隊營)+ Star Base(星基),
	// 但不應解鎖任何其他建築(如自動工廠需 Advanced Construction)。
	completed := map[ResearchTopic]bool{TOPIC_ENGINEERING: true}
	got := AvailableBuildings(completed)
	names := make(map[string]bool, len(got))
	for _, b := range got {
		names[b.NameZH] = true
	}
	if len(got) != 2 {
		t.Fatalf("只完成 Engineering 應解鎖 2 項(海軍陸戰隊營+星基),got %d: %+v", len(got), names)
	}
	if !names["海軍陸戰隊營"] || !names["星基"] {
		t.Fatalf("應包含海軍陸戰隊營+星基,got %+v", names)
	}

	// 追加 Advanced Construction:應再解鎖自動工廠 + 飛彈基地。
	completed[TOPIC_ADVANCED_CONSTRUCTION] = true
	got = AvailableBuildings(completed)
	names = make(map[string]bool, len(got))
	for _, b := range got {
		names[b.NameZH] = true
	}
	if len(got) != 4 {
		t.Fatalf("追加 Advanced Construction 後應解鎖 4 項,got %d: %+v", len(got), names)
	}
	if !names["自動工廠"] || !names["飛彈基地"] {
		t.Fatalf("應包含自動工廠+飛彈基地,got %+v", names)
	}
}

// TestBuildingsMaintenanceSampleAgainstManual 抽樣核對維護費(BC/turn)與
// docs/tech/colony-buildings.md 手冊數值一致(高可信度資料,非估計)。
func TestBuildingsMaintenanceSampleAgainstManual(t *testing.T) {
	cases := []struct {
		nameZH string
		want   int
	}{
		{"海軍陸戰隊營", 1}, // Marine Barracks
		{"太空港", 1},    // Spaceport
		{"核心廢料場", 8},  // Core Waste Dumps(全表最高維護費)
		{"食物複製機", 10}, // Food Replicators
		{"行星屏障護盾", 5}, // Planetary Barrier Shield
		{"星基", 2},     // Star Base
		{"星辰要塞", 4},   // Star Fortress
	}
	for _, c := range cases {
		b, ok := BuildingByNameZH(c.nameZH)
		if !ok {
			t.Fatalf("找不到建築 %s", c.nameZH)
		}
		if b.MaintenanceBC != c.want {
			t.Errorf("%s 維護費應為 %d,got %d", c.nameZH, c.want, b.MaintenanceBC)
		}
	}
}

// TestArmorBarracksCostIsManualSourced 驗證唯一有手冊實據的建造成本(150 PP)標記正確。
func TestArmorBarracksCostIsManualSourced(t *testing.T) {
	b, ok := BuildingByNameZH("裝甲營房")
	if !ok {
		t.Fatal("找不到裝甲營房")
	}
	if b.ProductionCost != 150 {
		t.Fatalf("裝甲營房建造成本應為手冊實據 150 PP,got %d", b.ProductionCost)
	}
	if b.EstimatedCost {
		t.Fatal("裝甲營房是唯一有手冊實據的建造成本,EstimatedCost 應為 false")
	}
}

// TestOtherCostsAreMarkedEstimated 抽樣驗證其餘建築的 PP 成本誠實標記為估計值。
func TestOtherCostsAreMarkedEstimated(t *testing.T) {
	for _, zh := range []string{"研究實驗室", "自動工廠", "星基", "生態圈", "行星屏障護盾"} {
		b, ok := BuildingByNameZH(zh)
		if !ok {
			t.Fatalf("找不到建築 %s", zh)
		}
		if !b.EstimatedCost {
			t.Fatalf("%s 應標記 EstimatedCost=true(手冊未給實據)", zh)
		}
	}
}

// TestCommandPointsFromBuildings 驗證星基/戰鬥站/星辰要塞的指揮評等供給,以及三者「取代關係」
// 下不疊加(GAME_MANUAL.pdf p.79/82/83)。
func TestCommandPointsFromBuildings(t *testing.T) {
	cases := []struct {
		name  string
		built map[string]bool
		want  int
	}{
		{"無任何軌道衛星", nil, 0},
		{"只有星基", map[string]bool{"星基": true}, 1},
		{"只有戰鬥站", map[string]bool{"戰鬥站": true}, 2},
		{"只有星辰要塞", map[string]bool{"星辰要塞": true}, 3},
		{"星基+戰鬥站同時記錄(取代關係,不疊加,取最高階)", map[string]bool{"星基": true, "戰鬥站": true}, 2},
		{"三者同時記錄(取代關係,不疊加,取最高階)", map[string]bool{"星基": true, "戰鬥站": true, "星辰要塞": true}, 3},
		{"其他無關建築不影響", map[string]bool{"海軍陸戰隊營": true}, 0},
	}
	for _, c := range cases {
		if got := CommandPointsFromBuildings(c.built); got != c.want {
			t.Errorf("%s: CommandPointsFromBuildings=%d, want %d", c.name, got, c.want)
		}
	}
}
