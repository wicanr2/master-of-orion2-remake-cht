package shell

import "testing"

// TestApplyGovernment 驗證政府型態對已建模資源的乘數效果(手冊 p.20–23 明列百分比)。
func TestApplyGovernment(t *testing.T) {
	// 民主:研究 +50%。
	s := NewDemoSession()
	base := s.PlayerColonies[0].ResearchPerScientist
	s.ApplyGovernment(3) // 民主
	if got, want := s.PlayerColonies[0].ResearchPerScientist, (base*3+1)/2; got != want {
		t.Fatalf("民主研究應 ×1.5:%d → 期望 %d,得 %d", base, want, got)
	}

	// 封建:研究減半。
	s = NewDemoSession()
	base = s.PlayerColonies[0].ResearchPerScientist
	s.ApplyGovernment(1) // 封建
	if got, want := s.PlayerColonies[0].ResearchPerScientist, base/2; got != want {
		t.Fatalf("封建研究應減半:%d → 期望 %d,得 %d", base, want, got)
	}

	// 統一:食物 +50%、產能 +50%。
	s = NewDemoSession()
	bf := s.PlayerColonies[0].FoodPerFarmer
	bi := s.PlayerColonies[0].IndustryPerWorker
	s.ApplyGovernment(2) // 統一
	if got, want := s.PlayerColonies[0].FoodPerFarmer, (bf*3+1)/2; got != want {
		t.Fatalf("統一食物應 ×1.5:%d → 期望 %d,得 %d", bf, want, got)
	}
	if got, want := s.PlayerColonies[0].IndustryPerWorker, (bi*3+1)/2; got != want {
		t.Fatalf("統一產能應 ×1.5:%d → 期望 %d,得 %d", bi, want, got)
	}

	// 獨裁:基準,研究不變。
	s = NewDemoSession()
	base = s.PlayerColonies[0].ResearchPerScientist
	s.ApplyGovernment(0)
	if got := s.PlayerColonies[0].ResearchPerScientist; got != base {
		t.Fatalf("獨裁研究應不變:%d → %d", base, got)
	}
}
