package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// EmpireOutput 是一個帝國(玩家)一回合的結算結果:各殖民地經濟 + 帝國層級聚合 + 研究推進。
type EmpireOutput struct {
	Colonies         []ColonyOutput // 對應輸入 colonies 順序
	TotalFood        int            // 各殖民地食物盈餘總和
	TotalNetIndustry int            // 各殖民地淨工業總和
	TotalResearch    int            // 各殖民地研究總和(投入研究進度)
	TaxRevenue       int            // 各殖民地稅收 BC 總和
	NetBC            int            // 本回合國庫淨變化(TaxRevenue - Maintenance)
	Player           PlayerState    // 研究推進 + BC 結算後的玩家狀態
	ResearchDone     bool           // 本回合是否有研究主題完成
}

// RunEmpireTurn 編排一個帝國的一回合:
//  1. 逐殖民地跑經濟結算(RunColonyTurn)。
//  2. 聚合帝國層級的食物盈餘 / 淨工業 / 研究點。
//  3. 用研究總點數推進研究進度(RunResearchPhase)。
//
// 注意:人口成長(各 ColonyOutput.PopGrowth)在本引擎層只輸出、不回寫 Population——MOO2 的
// 成長以分數累積到門檻才 +1 人口單位,該累積門檻/尺度手冊未給、存檔未能乾淨反推(避免臆造)。
// 「累積→回寫 Population」由上層 shell.GameSession.advancePopulation 以 remake 調校門檻處理
// (見該處 provenance 註記),保持本引擎層公式純淨。國庫 BC 結算已於下方以稅收-維護費處理。
func RunEmpireTurn(ps PlayerState, colonies []ColonyState) EmpireOutput {
	out := EmpireOutput{Colonies: make([]ColonyOutput, len(colonies))}
	for i, cs := range colonies {
		co := RunColonyTurn(cs)
		out.Colonies[i] = co
		out.TotalFood += co.FoodSurplus
		out.TotalNetIndustry += co.NetIndustry
		out.TotalResearch += co.Research
		// 稅收:對各殖民地淨工業依帝國稅率抽稅(gamedata.IncomeTaxRevenue,1:1 換 BC)。
		out.TaxRevenue += gamedata.IncomeTaxRevenue(co.NetIndustry, ps.TaxRate)
	}
	out.Player, out.ResearchDone = RunResearchPhase(ps, out.TotalResearch)
	// 國庫結算:稅收 - 維護費。
	out.NetBC = out.TaxRevenue - ps.Maintenance
	out.Player.BC += out.NetBC
	return out
}
