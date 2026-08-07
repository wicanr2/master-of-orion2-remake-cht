package gamedata

import "testing"

// initial_buildings_test.go:開局建築優先清單的護欄。

// 上限表 byte_13A3A = 3, 5, 9,與手冊逐字相同。
func TestInitialBuildingCapMatchesTheOriginalTable(t *testing.T) {
	for _, c := range []struct{ level, want int }{{0, 3}, {1, 5}, {2, 9}} {
		if got := InitialBuildingCap(c.level); got != c.want {
			t.Errorf("等級 %d 的上限應為 %d,實得 %d", c.level, c.want, got)
		}
	}
	for _, lv := range []int{-1, 3, 99} {
		if got := InitialBuildingCap(lv); got != 5 {
			t.Errorf("超出範圍的等級 %d 應退回 5,實得 %d", lv, got)
		}
	}
}

// 清單順序照原版:開頭是防禦升級鏈**最強的排最前面**(41 → 8 → 40)。
func TestInitialBuildingOrderStartsWithTheStrongestDefence(t *testing.T) {
	want := []int{41, 8, 40, 21, 22}
	for i, w := range want {
		if InitialBuildingOrder[i] != w {
			t.Errorf("第 %d 個應為 %d,實得 %d", i, w, InitialBuildingOrder[i])
		}
	}
	if len(InitialBuildingOrder) != 32 {
		t.Errorf("清單長度應為 32,實得 %d", len(InitialBuildingOrder))
	}
	// 不能有重複——重複代表抄的時候把某個 dword 拆錯了。
	seen := map[int]bool{}
	for i, id := range InitialBuildingOrder {
		if seen[id] {
			t.Errorf("第 %d 個編號 %d 重複了", i, id)
		}
		seen[id] = true
	}
}

// ★ 手冊那句話的機器版驗證:一般等級的開局科技 × 這份清單 = 正好兩棟。
//
// 手冊:「Pre-warp and Average Tech games only start with Marine Barracks and a Star Base
// because no other techs are Known that are also in the default initial buildings list.」
//
// 這一條同時驗了三張表:優先清單、開局主題表(StartingTopicOrder)、建築前置表。
// 任何一張抄錯,這裡就會紅。
func TestAverageStartGivesExactlyMarineBarracksAndStarBase(t *testing.T) {
	known := map[ResearchTopic]bool{TOPIC_STARTING_TECH: true}
	for _, topic := range StartingTopics(1) {
		known[topic] = true
	}
	got := InitialBuildings(known, InitialBuildingCap(1))
	if len(got) != 2 {
		t.Fatalf("一般等級應正好兩棟(手冊),實得 %d 棟:%v", len(got), got)
	}
	// 順序照清單:Star Base(40)在 Marine Barracks(22)前面。
	if got[0] != "星基" || got[1] != "海軍陸戰隊營" {
		t.Errorf("應為 [星基 海軍陸戰隊營](清單順序 40 → 22),實得 %v", got)
	}
}

// 曲速前只知道 ENGINEERING,結果與一般相同——那兩棟的前置都是 ENGINEERING。
// 手冊把兩級並列講(「Pre-warp and Average Tech games only start with…」)正是這個意思。
func TestPrewarpStartMatchesAverage(t *testing.T) {
	pre := map[ResearchTopic]bool{TOPIC_STARTING_TECH: true}
	for _, topic := range StartingTopics(0) {
		pre[topic] = true
	}
	got := InitialBuildings(pre, InitialBuildingCap(0))
	if len(got) != 2 {
		t.Fatalf("曲速前也該是兩棟,實得 %d 棟:%v", len(got), got)
	}
}

// limit 要真的截斷,而且是照順序取前面的。
func TestInitialBuildingsRespectsTheLimit(t *testing.T) {
	known := map[ResearchTopic]bool{TOPIC_STARTING_TECH: true}
	for _, topic := range StartingTopicOrder {
		known[topic] = true
	}
	if got := InitialBuildings(known, 1); len(got) != 1 || got[0] != "星基" {
		t.Errorf("limit=1 應只回清單最前面那棟(星基),實得 %v", got)
	}
	if got := InitialBuildings(known, 0); got != nil {
		t.Errorf("limit=0 應回空,實得 %v", got)
	}
}

// known 為 nil / 空要回空,不是「全部都知道」。
//
// 這條是最容易寫反的一個預設值:誤判成全解會讓母星憑空多出十幾棟建築,
// 而畫面上看起來只是「開局比較強」,不會有錯誤訊息。
func TestInitialBuildingsWithNoKnownTechReturnsNothing(t *testing.T) {
	if got := InitialBuildings(nil, 9); got != nil {
		t.Errorf("沒有已知科技時應回空,實得 %v", got)
	}
	if got := InitialBuildings(map[ResearchTopic]bool{}, 9); got != nil {
		t.Errorf("空的已知科技表應回空,實得 %v", got)
	}
}
