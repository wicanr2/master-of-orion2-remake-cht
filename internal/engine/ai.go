package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"

// ai.go:把 AI 決策層(internal/ai,設計性重建)接進回合引擎——讓 AI 玩家每回合自行
// 分配殖民地工作與稅率,再跑經濟結算。engine → ai 單向依賴(ai 不 import engine)。
//
// ⚠ AI 決策為【設計性重建,非原版 MOO2 行為】(見 internal/ai 套件說明)。

// 稅率決策的國庫警戒線【設計值】。
const (
	aiTreasuryLow  = 50
	aiTreasuryHigh = 300
)

// ApplyAIEconomy 讓 AI(依 profile)重新分配各殖民地的農夫/工人/科學家,並依國庫調整稅率。
// 回傳套用決策後的玩家與殖民地狀態(未跑回合;交給 RunEmpireTurn)。
func ApplyAIEconomy(ps PlayerState, colonies []ColonyState, profile ai.Profile) (PlayerState, []ColonyState) {
	out := make([]ColonyState, len(colonies))
	for i, cs := range colonies {
		f, w, s := ai.DecideColonyJobs(cs.Population, cs.FoodPerFarmer, profile)
		cs.Farmers, cs.Workers, cs.Scientists = f, w, s
		out[i] = cs
	}
	ps.TaxRate = ai.DecideTaxRate(ps.BC, aiTreasuryLow, aiTreasuryHigh)
	return ps, out
}

// RunAIEmpireTurn 是 AI 玩家的完整一回合:先套 AI 經濟決策,再跑帝國回合結算。
func RunAIEmpireTurn(ps PlayerState, colonies []ColonyState, profile ai.Profile) EmpireOutput {
	ps, colonies = ApplyAIEconomy(ps, colonies, profile)
	return RunEmpireTurn(ps, colonies)
}
