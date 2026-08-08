package gamedata

import "testing"

// TestBuildingsCount 驗證表的筆數,一項不多不少。
//
// ⚠ 2026-08-07 從 40 改成 41,原因值得寫下來:先前的 40 是**手冊《The Big List》的記帳
// 慣例**(35 建築 + 5 衛星),不是遊戲規則。手冊把恆星轉換器(行星版)放在單獨一節
// (§十一,p.106)所以不計入 40;但原版建築表(`off_17EB3D`)裡它就是第 42 棟,
// 和其他 47 棟完全同構——有成本(1000 PP)、有維護費(6 BC)、有分類(0 = 地表建築)。
//
// 也就是說「40」從一開始就不是原版的數字,只是抄手冊時連著記帳方式一起抄了。
// 真正的上界是原版建築表的 48 棟;其餘 7 個編號不在這裡的理由各自不同,
// 逐一列在 `docs/re/01-gap-report.md` 第 12 項(2 個自動給予、3 個是 SpecialActions、
// 2 個仍缺)。改這個數字前先去看那一項。
func TestBuildingsCount(t *testing.T) {
	if got := len(Buildings); got != 41 {
		t.Fatalf("Buildings 應有 41 筆(手冊 40 項 + 恆星轉換器),got %d", got)
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

// TestBuildingCostsMatchOriginalTable 抽驗建造成本已換成原版執行檔建築表的真值。
//
// 先前這兩支測試驗的是相反的事:「裝甲營房是唯一有實據的 150 PP」「其餘一律標
// EstimatedCost=true」。2026-08-07 把 `off_17EB3D` 那張表抽出來之後,全部 40 項都有真值,
// `EstimatedCost` 欄位也拿掉了——測試跟著改成驗真值,不是把舊斷言留著當歷史。
//
// 這裡挑的五筆都是**舊估計值錯很多**的,錯回去會立刻被抓到。
func TestBuildingCostsMatchOriginalTable(t *testing.T) {
	want := map[string]int{
		"裝甲營房":   150,  // 舊值也是 150(唯一有 modding 範例佐證的那筆)
		"核心廢料場":  200,  // 舊估計 550
		"食物複製機":  200,  // 舊估計 460
		"歡樂穹頂":   250,  // 舊估計 800
		"星辰要塞":   2500, // 舊估計 800——差最多的一筆
		"行星屏障護盾": 500,  // 舊估計 1200
		"研究實驗室":  60,
	}
	for zh, pp := range want {
		b, ok := BuildingByNameZH(zh)
		if !ok {
			t.Fatalf("找不到建築 %s", zh)
		}
		if b.ProductionCost != pp {
			t.Errorf("%s 建造成本 = %d,原版表為 %d", zh, b.ProductionCost, pp)
		}
	}
}

// TestNoBuildingHasZeroCost:全 40 項都要有成本。抽表時漏一筆會變成 0 PP(瞬間蓋好)。
func TestNoBuildingHasZeroCost(t *testing.T) {
	for _, b := range Buildings {
		if b.ProductionCost <= 0 {
			t.Errorf("%s(%s)建造成本為 %d", b.NameZH, b.NameEN, b.ProductionCost)
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
