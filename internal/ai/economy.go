// Package ai 是 AI 決策層的【設計性重建】。
//
// ⚠ 重要:這不是原版 MOO2 AI。MOO2 的 AI 決策邏輯與難度加成官方手冊未給、社群也公認未破解
// （見 docs/tech/community-mechanics-findings.md)。本套件在使用者授權下,做一套「合理但非原版」
// 的 AI 決策啟發式,供 remake 有可運作的對手。所有權重與門檻都是設計選擇,非原版數值。
//
// 本套件操作乾淨的 int 輸入/輸出,不 import engine/save(維持可純測、無耦合)。
package ai

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
