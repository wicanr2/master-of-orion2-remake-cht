package gamedata

import "math"

// 殖民地人口成長公式,移植自 MOO2 patch 1.5 官方手冊附錄「Notes on Population Growth」(p111,
// moo2_patch1.5/MANUAL_150.html)。openorion2 未實作成長邏輯(只從存檔讀 pop_growth),故以手冊為權威來源。
//
// 手冊變數:POPRACE=該種族人口、POPAGG=星球總人口(含被併吞人口/原住民/機器人)、POPMAX=人口上限、
// PROD=殖民地總工業產出。人口皆以「人口單位」計;人口上限硬上限 42(patch 1.5「Population Capacity Clipped To 42」)。
const (
	growthFactor1 = 2000 // FACTOR1(growth_formula_factor 預設)
	housingFactor = 40   // FACTOR2(housing_formula_factor 預設)
	MaxPopulation = 42   // 人口硬上限(patch 1.5)
)

// ColonyBaseGrowth 回傳基礎成長率 a(未計獎金),手冊公式:
//
//	a = trunc[ (FACTOR1 * POPRACE * (POPMAX - POPAGG) / POPMAX) ^ 0.5 ]
//
// popAgg>=popMax(已滿)或參數非法時回 0。
func ColonyBaseGrowth(popRace, popAgg, popMax int) int {
	if popMax <= 0 || popRace <= 0 || popAgg >= popMax {
		return 0
	}
	v := float64(growthFactor1) * float64(popRace) * float64(popMax-popAgg) / float64(popMax)
	if v < 0 {
		return 0
	}
	return int(math.Sqrt(v)) // trunc(向下取整)
}

// ColonyHousingBonus 回傳住房獎金 h(百分點),手冊 h = FACTOR2 * PROD / POPAGG(僅「住房中」適用)。
// popAgg<=0 回 0。
func ColonyHousingBonus(prod, popAgg int) int {
	if popAgg <= 0 {
		return 0
	}
	return housingFactor * prod / popAgg
}

// ColonyGrowth 回傳最終成長 a*b,手冊:b = (100 + g+r+i+t+l+e+h)/100,a 先 trunc 再乘 b。
// bonusSum = 七項獎金(百分點)之和(g 一般/r 種族/i AI/t 科技/l/e/h 住房);全 0 時 b=1。
// r(種族獎金)可為負(−50% 記 −50),故 bonusSum 可為負。
func ColonyGrowth(baseGrowth, bonusSum int) int {
	return baseGrowth * (100 + bonusSum) / 100
}
