package gamedata

import "testing"

// 從 openorion2 tech.cpp:169-305 直接抽取的基準(cost 與每列 choices 數),
// 由主代理(Opus)用獨立解析腳本產生,交叉驗證 techtree.go 的逐字轉寫無誤。
var cResearchCosts = [83]int{0, 400, 650, 150, 80, 250, 4500, 250, 1500, 250, 150, 2000, 2000, 2000, 1500, 400, 1150, 4500, 80, 1150, 400, 250, 50, 80, 3500, 2750, 3500, 1500, 50, 50, 2750, 150, 6000, 4500, 900, 1150, 650, 3500, 4500, 6000, 10000, 900, 6000, 1150, 1500, 900, 2750, 1150, 10000, 6000, 4500, 4500, 2000, 2000, 900, 50, 150, 50, 7500, 3500, 900, 4500, 650, 900, 2750, 1500, 250, 3500, 15000, 15000, 7500, 7500, 2000, 650, 0, 25000, 25000, 25000, 25000, 25000, 25000, 25000, 25000}
var cResearchNumChoices = [83]int{0, 3, 2, 3, 3, 3, 4, 3, 3, 2, 1, 2, 1, 3, 3, 3, 3, 2, 2, 3, 3, 3, 4, 3, 3, 3, 3, 3, 1, 3, 2, 2, 1, 2, 2, 1, 3, 3, 3, 3, 3, 3, 3, 1, 2, 3, 2, 3, 2, 3, 3, 2, 3, 3, 2, 3, 3, 3, 2, 3, 3, 3, 3, 3, 3, 2, 3, 2, 3, 3, 3, 3, 3, 2, 0, 1, 1, 1, 1, 1, 1, 1, 1}

func TestResearchChoicesMatchC(t *testing.T) {
	for i := 0; i < 83; i++ {
		rc := ResearchChoiceFor(ResearchTopic(i))
		if rc.Cost != cResearchCosts[i] {
			t.Errorf("topic %d: Cost = %d,C 基準 %d", i, rc.Cost, cResearchCosts[i])
		}
		if len(rc.Choices) != cResearchNumChoices[i] {
			t.Errorf("topic %d: len(Choices) = %d,C 基準 %d", i, len(rc.Choices), cResearchNumChoices[i])
		}
	}
}
