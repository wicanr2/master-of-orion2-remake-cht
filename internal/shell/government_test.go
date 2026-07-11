package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

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

// TestEndTurnGovtBonusMoneyWiring 驗證 EndTurn 依 s.Government 算出的
// gamedata.IncomeGovtMoneyBonusPercent 有實際接進 engine.RunEmpireTurn(不是死碼)。
//
// 判定依據(income 死碼接線 2026-07-11 輪次):民主(moraleGovByIndex[3])在 MANUAL_150.html
// 有 +50% money 加成,demo 預設獨裁(0 加成)——切換到民主後,同一組殖民地稅收應嚴格更高;
// 維持獨裁則應與切換前一致(no-op)。用「稅收隨政府切換而變 / 不變」這種相對比較,而非硬編某
// 一版本的絕對稅收數字,避免耦合到 demo 母星初始 yield 之後若調整而改壞這條測試。
func TestEndTurnGovtBonusMoneyWiring(t *testing.T) {
	// ⚠ 2026-07-12:母星分配校正為忠實的 農4/工1/科3(工1)後,母星淨工業僅 3,稅收(工業×40%)
	// floor 到 1,民主 +50%(1×1.5=1.5→1)被整數捨去,兩組稅收都變 1、相對比較失效——這是低產出
	// 整數捨入假象,非 wiring 壞掉。為隔離「加成確實流進 TaxRevenue」,兩組都把母星工人臨時拉高到
	// 工業足以讓 +50% 越過捨入門檻(測試專屬設定,不影響正式母星忠實分配)。
	bumpWorkers := func(s *GameSession) {
		s.PlayerColonies[0].Population = 20
		s.PlayerColonies[0].Farmers = 10
		s.PlayerColonies[0].Workers = 10
		s.PlayerColonies[0].Scientists = 0
	}

	// 獨裁(demo 預設):s.Player.GovtBonusMoneyPercent 應為 0,EndTurn 前後稅收不受本次接線影響。
	sDict := NewDemoSession()
	bumpWorkers(sDict)
	sDict.EndTurn()
	if sDict.Player.GovtBonusMoneyPercent != 0 {
		t.Fatalf("獨裁 GovtBonusMoneyPercent = %d,預期 0", sDict.Player.GovtBonusMoneyPercent)
	}
	dictTax := sDict.LastPlayerOutput.TaxRevenue

	// 民主:切換政府後 EndTurn,GovtBonusMoneyPercent 應反映 +50%,且稅收嚴格高於獨裁那組
	// (同一份殖民地/稅率,唯一差異是政府 money 加成)。
	sDemo := NewDemoSession()
	bumpWorkers(sDemo)
	sDemo.ApplyGovernment(3) // 民主
	sDemo.EndTurn()
	if sDemo.Player.GovtBonusMoneyPercent != gamedata.IncomeGovtBonusDemocracyMoneyPercent {
		t.Fatalf("民主 GovtBonusMoneyPercent = %d,預期 %d", sDemo.Player.GovtBonusMoneyPercent, gamedata.IncomeGovtBonusDemocracyMoneyPercent)
	}
	demoTax := sDemo.LastPlayerOutput.TaxRevenue
	if dictTax <= 0 {
		t.Fatalf("獨裁稅收應 > 0 才能比較加成效果,實得 %d(demo 初始稅率可能為 0,測試前提失效)", dictTax)
	}
	if demoTax <= dictTax {
		t.Fatalf("民主稅收 = %d,應嚴格高於獨裁稅收 %d(+50%% money 加成應生效)", demoTax, dictTax)
	}
}
