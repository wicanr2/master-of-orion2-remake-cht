package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"

// ai.go:把 AI 決策層(internal/ai)接進回合引擎——讓 AI 玩家每回合自行分配殖民地工作與稅率。
// 透過 ai.Decider 介面注入,支援 remake / original 兩種 AI 模式(玩家可選)。
// engine → ai 單向依賴(ai 不 import engine)。
//
// ⚠ remake AI 為【設計性重建,非原版】;original 模式待 RE(見 internal/ai/decider.go)。

// ApplyAIEconomy 讓 AI(以 decider 決策)重新分配各殖民地的農夫/工人/科學家,並依國庫調整稅率。
// 回傳套用決策後的玩家與殖民地狀態(未跑回合;交給 RunEmpireTurn)。
func ApplyAIEconomy(ps PlayerState, colonies []ColonyState, decider ai.Decider) (PlayerState, []ColonyState) {
	out := make([]ColonyState, len(colonies))
	for i, cs := range colonies {
		// Maintenance(帝國固定支出)傳入供財政保底(見 ai.Decider.ColonyJobs 註解):AI 不應該
		// 把殖民地職務分配壓到連稅率上限都打平不了固定支出的程度,結構性赤字非原版行為,見
		// docs/tech/ai-fiscal-solvency.md。
		f, w, s := decider.ColonyJobs(cs.Population, cs.FoodPerFarmer, cs.IndustryPerWorker, cs.MoralePercent, ps.Maintenance)
		cs.Farmers, cs.Workers, cs.Scientists = f, w, s
		out[i] = cs
	}
	ps.TaxRate = decider.TaxRate(ps.BC)
	return ps, out
}

// RunAIEmpireTurn 是 AI 玩家的完整一回合:先套 AI 經濟決策,再跑帝國回合結算。
func RunAIEmpireTurn(ps PlayerState, colonies []ColonyState, decider ai.Decider) EmpireOutput {
	ps, colonies = ApplyAIEconomy(ps, colonies, decider)
	return RunEmpireTurn(ps, colonies)
}
