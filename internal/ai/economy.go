// Package ai 是 AI 決策層的【設計性重建】。
//
// ⚠ 重要:這不是原版 MOO2 AI。MOO2 的 AI 決策邏輯與難度加成官方手冊未給、社群也公認未破解
// （見 docs/tech/community-mechanics-findings.md)。本套件在使用者授權下,做一套「合理但非原版」
// 的 AI 決策啟發式,供 remake 有可運作的對手。所有權重與門檻都是設計選擇,非原版數值。
//
// 本套件操作乾淨的 int 輸入/輸出,不 import engine/save(維持可純測、無耦合);財政保底計算
// (見 MinWorkersForSolvency)改用既有 gamedata 換算式(MoraleProductionOutput/
// IncomeTaxRevenue/TaxRateMaxPercent),不重新發明公式,gamedata 為無依賴的純資料/公式包,
// import 它不會造成循環依賴。
package ai

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// Profile 是 AI 的施政傾向【設計】。權重決定「餵飽人口後,餘力偏工業或研究」。
type Profile struct {
	Name           string
	IndustryWeight int // 工業偏好權重
	ResearchWeight int // 研究偏好權重
}

// 預設幾種性格傾向【設計值】。
var (
	ProfileAggressive   = Profile{"aggressive", 3, 1}   // 好戰:重工業(造艦)
	ProfileScientific   = Profile{"scientific", 1, 3}   // 科學:重研究
	ProfileBalanced     = Profile{"balanced", 1, 1}     // 平衡
	ProfileExpansionist = Profile{"expansionist", 2, 1} // 擴張:偏工業(造殖民船)
)

// ceilDiv 回傳 ceil(a/b)(b>0)。
func ceilDiv(a, b int) int {
	if b <= 0 {
		return 0
	}
	return (a + b - 1) / b
}

// DecideColonyJobs 決定一個殖民地的工作分配(農夫/工人/科學家)【設計啟發式】:
//  1. 先分配足夠農夫餵飽全體人口(每人口吃 1 食物):farmers = ceil(population / foodPerFarmer)。
//  2. 餘力依 profile 的工業/研究權重分配給工人/科學家。
//
// foodPerFarmer<=0(無法務農,如純機器人殖民地)時全體不務農,全部依權重分工業/研究。
func DecideColonyJobs(population, foodPerFarmer int, p Profile) (farmers, workers, scientists int) {
	if population <= 0 {
		return 0, 0, 0
	}
	if foodPerFarmer > 0 {
		farmers = ceilDiv(population, foodPerFarmer)
		if farmers > population {
			farmers = population
		}
	}
	remaining := population - farmers
	wsum := p.IndustryWeight + p.ResearchWeight
	if remaining <= 0 || wsum <= 0 {
		return farmers, remaining, 0
	}
	workers = remaining * p.IndustryWeight / wsum
	scientists = remaining - workers
	return farmers, workers, scientists
}

// MinWorkersForSolvency 回傳「即使把稅率開到手冊上限(gamedata.TaxRateMaxPercent=50%),稅收
// 仍至少打平 maintenanceBC」所需的最少工人數,上限為 maxWorkers(呼叫端通常傳「餵飽人口後的
// 剩餘人力」,不會去動農夫)。
//
// 第一性原理:不論 AI 施政多偏好研究,殖民地的稅收長期低於固定支出(建築維護費)就是結構性
// 赤字,國庫必然無下限發散——這不是「原版數值」(原版 AI 邏輯未公開,見
// docs/tech/original-ai-re.md),而是任何理性政體都會遵守的財政常識:先確保有能力在稅率上限
// 打平帳,餘力才依偏好投入研究/更多工業。用「稅率上限」而非「當下稅率」當基準,是刻意留的
// 安全邊際下限——AI 的 DecideTaxRate 平時多半用較低稅率(國庫充裕時只收 10~30%),只有國庫見底
// 才會拉滿 50%,若用當下稅率當門檻,國庫充裕時反而會算出「不需要保底」,錯失提前佈局的機會。
//
// 直接呼叫既有 gamedata 公式(MoraleProductionOutput 套用士氣、IncomeTaxRevenue 換算稅收),
// 不重新發明轉換率,亦不假造污染清理項——本專案 AI 殖民地規模下,達到打平所需的工人數產出遠低於
// 星球污染容忍值(見 docs/tech/ai-fiscal-solvency.md 逐步驗算),忽略污染清理不影響本函式在目前
// 場景下的正確性;若未來 AI 出現高工業大型殖民地,建議連同容忍值/清理成本一併傳入重新驗算。
//
// maintenanceBC<=0(無固定支出需要打平)或 industryPerWorker<=0(無法產出工業,如純研究/機器人
// 殖民地)時直接回 0(不需要、也無法用工人打平)。
func MinWorkersForSolvency(industryPerWorker, moralePercent, maintenanceBC, maxWorkers int) int {
	if maintenanceBC <= 0 || industryPerWorker <= 0 {
		return 0
	}
	for w := 0; w <= maxWorkers; w++ {
		gross := gamedata.MoraleProductionOutput(w*industryPerWorker, moralePercent)
		if gamedata.IncomeTaxRevenue(gross, gamedata.TaxRateMaxPercent) >= maintenanceBC {
			return w
		}
	}
	return maxWorkers
}

// DecideColonyJobsSolvent 是 DecideColonyJobs 加上「財政保底」的版本【設計啟發式】:
//  1. 先照 profile 權重比例分配(同 DecideColonyJobs)。
//  2. 若分配到的工人數不足以在稅率上限打平 maintenanceBC(MinWorkersForSolvency),從科學家
//     挪一部分回工人,直到打平或科學家歸零為止——只挪動「維持財政自立所需的最少量」,不會把
//     偏研究的性格整個打平成平衡型(見 docs/tech/ai-fiscal-solvency.md 各 profile 實測數字)。
//
// 這是本專案唯一「AI 職務分配」的財政保底邏輯:原本的 DecideColonyJobs 純比例分配在
// Scientific(研究權重遠高於工業)且殖民地人口小(母星規模)時,會把工人壓到個位數,稅收
// 上限打平不了固定維護費,國庫必然單調變負無下限(見 docs/tech/ai-fiscal-solvency.md §1
// 實測記錄)。這裡不改 DecideColonyJobs 本體(維持其他呼叫端/測試對「純比例分配」的既有預期),
// 另開這個保底版本給 Decider.ColonyJobs 使用。
func DecideColonyJobsSolvent(population, foodPerFarmer, industryPerWorker, moralePercent, maintenanceBC int, p Profile) (farmers, workers, scientists int) {
	farmers, workers, scientists = DecideColonyJobs(population, foodPerFarmer, p)
	need := MinWorkersForSolvency(industryPerWorker, moralePercent, maintenanceBC, workers+scientists)
	for workers < need && scientists > 0 {
		scientists--
		workers++
	}
	return farmers, workers, scientists
}

// DecideTaxRate 決定稅率(%)【設計啟發式】:國庫低於警戒線就提高稅率,充裕就降低鼓勵生產。
// 稅率夾在 0-50、以 10 為級距(對齊 gamedata 的稅率規則)。
func DecideTaxRate(treasuryBC, lowThreshold, highThreshold int) int {
	switch {
	case treasuryBC < lowThreshold:
		return 50
	case treasuryBC < highThreshold:
		return 30
	default:
		return 10
	}
}
