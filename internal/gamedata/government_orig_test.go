package gamedata

import "testing"

// 政體編號釘在**原版執行檔寫進 `[player+0x89F]` 的立即數**上。
//
// 出處與完整說明見 `assimilation.go` 的 AssimilationGovernment 檔頭。這一條的作用是
// 讓「重排列舉」變成一次紅燈,而不是一個安靜的語意漂移——那組編號同時被
// `MoraleGovernmentType`(士氣表)與存檔的 `government` 欄位吃著。
func TestGovernmentNumbersMatchTheExecutable(t *testing.T) {
	// techIdx → 取得該科技後原版寫進 [player+0x89F] 的值(sub_E4204)。
	cases := []struct {
		tech Technology
		gov  AssimilationGovernment
		byte int
	}{
		{TECH_CONFEDERATION, AssimConfederation, 1},
		{TECH_IMPERIUM, AssimImperium, 3},
		{TECH_FEDERATION, AssimFederation, 5},
		{TECH_GALACTIC_UNIFICATION, AssimGalacticUnification, 7},
	}
	for _, c := range cases {
		if int(c.gov) != c.byte {
			t.Errorf("%s:remake 編號 %d,原版寫入 %d", TechnologyName(c.tech), c.gov, c.byte)
		}
	}

	// 偶數是四個基本政體(自訂種族那一欄能選的四個),奇數是它們的科技升級版,
	// 而「值/2」就是同一族。
	pairs := []struct{ base, adv AssimilationGovernment }{
		{AssimFeudal, AssimConfederation},
		{AssimDictatorship, AssimImperium},
		{AssimDemocracy, AssimFederation},
		{AssimUnification, AssimGalacticUnification},
	}
	for _, p := range pairs {
		if p.base%2 != 0 {
			t.Errorf("基本政體 %d 應為偶數", p.base)
		}
		if p.adv != p.base+1 {
			t.Errorf("進階政體應緊接在基本政體之後:%d vs %d", p.adv, p.base)
		}
		if p.base/2 != p.adv/2 {
			t.Errorf("值/2 應同族:%d vs %d", p.base/2, p.adv/2)
		}
	}
}

// 士氣表的政體編號與同化表**必須是同一組**——原版只有一個 `[player+0x89F]`。
//
// 兩個列舉分開宣告是 Go 這一側的歷史(見 morale.go 檔頭),不是原版有兩套編號。
// 少了這條,其中一個被重排時另一個不會有任何反應。
func TestMoraleAndAssimilationGovernmentsShareTheSameNumbering(t *testing.T) {
	cases := []struct {
		morale MoraleGovernmentType
		assim  AssimilationGovernment
	}{
		{MoraleGovFeudalism, AssimFeudal},
		{MoraleGovConfederation, AssimConfederation},
		{MoraleGovDictatorship, AssimDictatorship},
		{MoraleGovImperium, AssimImperium},
		{MoraleGovDemocracy, AssimDemocracy},
		{MoraleGovFederation, AssimFederation},
		{MoraleGovUnification, AssimUnification},
		{MoraleGovGalacticUnification, AssimGalacticUnification},
	}
	for _, c := range cases {
		if int(c.morale) != int(c.assim) {
			t.Errorf("政體編號兩邊對不上:morale %d vs assim %d", c.morale, c.assim)
		}
	}
}
