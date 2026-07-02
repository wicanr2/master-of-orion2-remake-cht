package engine

// EmpireOutput 是一個帝國(玩家)一回合的結算結果:各殖民地經濟 + 帝國層級聚合 + 研究推進。
type EmpireOutput struct {
	Colonies         []ColonyOutput // 對應輸入 colonies 順序
	TotalFood        int            // 各殖民地食物盈餘總和
	TotalNetIndustry int            // 各殖民地淨工業總和
	TotalResearch    int            // 各殖民地研究總和(投入研究進度)
	Player           PlayerState    // 研究推進後的玩家狀態
	ResearchDone     bool           // 本回合是否有研究主題完成
}

// RunEmpireTurn 編排一個帝國的一回合:
//  1. 逐殖民地跑經濟結算(RunColonyTurn)。
//  2. 聚合帝國層級的食物盈餘 / 淨工業 / 研究點。
//  3. 用研究總點數推進研究進度(RunResearchPhase)。
//
// 注意:人口成長(各 ColonyOutput.PopGrowth)本回合只輸出、不回寫 Population——MOO2 的
// 成長以分數累積到門檻才 +1 人口單位,該累積門檻/尺度尚未對實機驗證,故不擅自換算(避免臆造)。
// 國庫(BC)結算需 income 公式(手冊未給精確式),同樣待移植,本編排器暫不動 BC。
func RunEmpireTurn(ps PlayerState, colonies []ColonyState) EmpireOutput {
	out := EmpireOutput{Colonies: make([]ColonyOutput, len(colonies))}
	for i, cs := range colonies {
		co := RunColonyTurn(cs)
		out.Colonies[i] = co
		out.TotalFood += co.FoodSurplus
		out.TotalNetIndustry += co.NetIndustry
		out.TotalResearch += co.Research
	}
	out.Player, out.ResearchDone = RunResearchPhase(ps, out.TotalResearch)
	return out
}
