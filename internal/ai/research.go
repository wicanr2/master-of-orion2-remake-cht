// Package ai 是 AI 決策層的【設計性重建】,本檔（research.go)同屬此範疇。
//
// ⚠ 重要:AI 研究選題策略是【設計性重建,非原版 MOO2 行為】。MOO2 官方手冊與社群逆向都未公開
// AI 選研究主題的實際演算法（見 docs/tech/community-mechanics-findings.md),本檔是在此空白下
// 設計的一套「合理但非原版」啟發式,單純讓 remake 的 AI 對手能自動推進科技樹。所有分類、門檻與
// 排序規則都是設計選擇,不代表原版數值或邏輯。
package ai

// ResearchCandidate 是一個可選研究主題【設計】。
//
// 由呼叫端(engine/gamedata 層)組出候選清單再傳入 ai 套件——本套件刻意不 import gamedata,
// 維持純可測、與遊戲資料解耦。
type ResearchCandidate struct {
	TopicID   int // 主題 ID,呼叫端自訂,ai 套件不解讀內容,只原樣回傳
	Cost      int // 研究成本(RP),數值越大代表主題越「高階」
	AreaIndex int // 研究領域索引(0-7),對應 gamedata ResearchArea 常數值:
	//    0=Biology 1=Power 2=Physics 3=Construction
	//    4=Fields  5=Chemistry 6=Computers 7=Sociology
	// (此處鏡射 gamedata 的常數值以維持解耦,不 import gamedata。)
}

// militaryAreas 是【設計】認定的「軍事相關」研究領域集合:Power(反應爐/引擎)、Physics(光束武器/
// 護盾)、Construction(裝甲/船體)、Fields(力場護盾)、Chemistry(飛彈/推進劑)。
// Biology(0)、Computers(6)、Sociology(7) 視為非軍事(生態/資訊/內政取向),不列入。
// 這是設計者的分類選擇,非原版資料。
var militaryAreas = map[int]bool{
	1: true, // Power
	2: true, // Physics
	3: true, // Construction
	4: true, // Fields
	5: true, // Chemistry
}

// DecideResearchTopic 依 AI 性格(Profile)從候選研究主題中選一個【設計啟發式】,回傳選中的
// TopicID;candidates 為空時回傳 -1。
//
// 策略(依 IndustryWeight 對 ResearchWeight 的相對大小決定,三分支皆為設計選擇):
//   - 好戰型(IndustryWeight > ResearchWeight,如 ProfileAggressive/ProfileExpansionist):
//     優先在「軍事領域」(militaryAreas)候選中選成本最低者,快速取得戰力提升;
//     若候選中沒有軍事領域主題,退回全體候選中選成本最低者。
//   - 科學型(ResearchWeight > IndustryWeight,如 ProfileScientific):
//     選成本最高者,視為長線投資高階科技,不計較短期進度。
//   - 平衡型(權重相等,如 ProfileBalanced):
//     選成本最低者,穩定推進科技樹、避免研究隊列停滯過久。
//
// 同成本並列時,保留候選清單中先出現者(穩定、可預期的決策順序)。
func DecideResearchTopic(candidates []ResearchCandidate, p Profile) int {
	if len(candidates) == 0 {
		return -1
	}

	switch {
	case p.IndustryWeight > p.ResearchWeight:
		if id, ok := cheapestInAreas(candidates, militaryAreas); ok {
			return id
		}
		return cheapestTopic(candidates)
	case p.ResearchWeight > p.IndustryWeight:
		return costliestTopic(candidates)
	default:
		return cheapestTopic(candidates)
	}
}

// cheapestTopic 回傳候選清單中成本最低者的 TopicID(候選非空)。
func cheapestTopic(candidates []ResearchCandidate) int {
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.Cost < best.Cost {
			best = c
		}
	}
	return best.TopicID
}

// costliestTopic 回傳候選清單中成本最高者的 TopicID(候選非空)。
func costliestTopic(candidates []ResearchCandidate) int {
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.Cost > best.Cost {
			best = c
		}
	}
	return best.TopicID
}

// cheapestInAreas 在 candidates 中只考慮 AreaIndex 落在 areas 集合內者,回傳成本最低的
// TopicID;若 candidates 中沒有任何主題落在 areas 內,ok 回傳 false。
func cheapestInAreas(candidates []ResearchCandidate, areas map[int]bool) (topicID int, ok bool) {
	found := false
	var best ResearchCandidate
	for _, c := range candidates {
		if !areas[c.AreaIndex] {
			continue
		}
		if !found || c.Cost < best.Cost {
			best = c
			found = true
		}
	}
	if !found {
		return -1, false
	}
	return best.TopicID, true
}
