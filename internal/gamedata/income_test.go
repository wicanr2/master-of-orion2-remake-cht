package gamedata

import "testing"

func TestIncomeConstants(t *testing.T) {
	// GAME_MANUAL.pdf / MANUAL_150.html 直接給出的常數,防止之後被誤改。
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"TaxRateMinPercent", TaxRateMinPercent, 0},
		{"TaxRateMaxPercent", TaxRateMaxPercent, 50},
		{"TaxRateStepPercent", TaxRateStepPercent, 10},
		{"TaxConversionNumerator", TaxConversionNumerator, 1},
		{"TaxConversionDenominator", TaxConversionDenominator, 1},
		{"TradeGoodsConversionNumerator", TradeGoodsConversionNumerator, 1},
		{"TradeGoodsConversionDenominator", TradeGoodsConversionDenominator, 2},
		{"TradeGoodsFantasticTraderConversionNumerator", TradeGoodsFantasticTraderConversionNumerator, 1},
		{"TradeGoodsFantasticTraderConversionDenominator", TradeGoodsFantasticTraderConversionDenominator, 1},
		{"IncomeFoodSurplusNumerator", IncomeFoodSurplusNumerator, 1},
		{"IncomeFoodSurplusDenominator", IncomeFoodSurplusDenominator, 2},
		{"IncomeFoodSurplusFantasticTraderPerUnitBC", IncomeFoodSurplusFantasticTraderPerUnitBC, 1},
		{"IncomeCommandOverflowCostPerPoint", IncomeCommandOverflowCostPerPoint, 10},
		{"IncomeFreighterMaintenanceNumerator", IncomeFreighterMaintenanceNumerator, 1},
		{"IncomeFreighterMaintenanceDenominator", IncomeFreighterMaintenanceDenominator, 2},
		{"IncomeMoraleProductionPercentPerIcon", IncomeMoraleProductionPercentPerIcon, 10},
		{"IncomeGovtBonusDemocracyMoneyPercent", IncomeGovtBonusDemocracyMoneyPercent, 50},
		{"IncomeGovtBonusFederationMoneyPercent", IncomeGovtBonusFederationMoneyPercent, 75},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d,預期 %d", c.name, c.got, c.want)
		}
	}
}

// TestIncomeGovtBonusFormula 驗證 MANUAL_150.html 給的換算公式本身:
// "value * 5 is the percent bonus for the item, for example democracy has a 10 * 5 = 50% bonus."
func TestIncomeGovtBonusFormula(t *testing.T) {
	if got := incomeGovtBonusDemocracyMoneyRaw * incomeGovtBonusRawToPercentMultiplier; got != IncomeGovtBonusDemocracyMoneyPercent {
		t.Errorf("democracy_money 原始值 %d * %d = %d,預期 %d",
			incomeGovtBonusDemocracyMoneyRaw, incomeGovtBonusRawToPercentMultiplier, got, IncomeGovtBonusDemocracyMoneyPercent)
	}
	if got := incomeGovtBonusFederationMoneyRaw * incomeGovtBonusRawToPercentMultiplier; got != IncomeGovtBonusFederationMoneyPercent {
		t.Errorf("federation_money 原始值 %d * %d = %d,預期 %d",
			incomeGovtBonusFederationMoneyRaw, incomeGovtBonusRawToPercentMultiplier, got, IncomeGovtBonusFederationMoneyPercent)
	}
}

func TestIncomeTaxRateIsValid(t *testing.T) {
	cases := []struct {
		rate int
		want bool
	}{
		{0, true},
		{10, true},
		{20, true},
		{30, true},
		{40, true},
		{50, true},
		{5, false},   // 不是 10% 的倍數
		{60, false},  // 超過上限
		{-10, false}, // 低於下限
		{55, false},  // 超過上限且非倍數
	}
	for _, c := range cases {
		if got := IncomeTaxRateIsValid(c.rate); got != c.want {
			t.Errorf("IncomeTaxRateIsValid(%d) = %v,預期 %v", c.rate, got, c.want)
		}
	}
}

func TestIncomeTaxRevenue(t *testing.T) {
	// GAME_MANUAL.pdf p.37:稅率扣掉的產能以 1:1 換成 BC。
	cases := []struct{ totalIndustry, taxRatePercent, want int }{
		{100, 50, 50}, // p.168 例子:50% 稅率 → 半數產能變稅收
		{100, 10, 10},
		{100, 0, 0},
		{33, 10, 3}, // 33*10/100=3.3 無條件捨去 → 3
	}
	for _, c := range cases {
		if got := IncomeTaxRevenue(c.totalIndustry, c.taxRatePercent); got != c.want {
			t.Errorf("IncomeTaxRevenue(%d, %d) = %d,預期 %d", c.totalIndustry, c.taxRatePercent, got, c.want)
		}
	}
}

func TestIncomeTaxRemainingIndustry(t *testing.T) {
	// GAME_MANUAL.pdf p.168:50% 稅率時,「Only the remaining half is available for building.」
	cases := []struct{ totalIndustry, taxRatePercent, want int }{
		{100, 50, 50},
		{33, 10, 30},
		{100, 0, 100},
	}
	for _, c := range cases {
		if got := IncomeTaxRemainingIndustry(c.totalIndustry, c.taxRatePercent); got != c.want {
			t.Errorf("IncomeTaxRemainingIndustry(%d, %d) = %d,預期 %d", c.totalIndustry, c.taxRatePercent, got, c.want)
		}
	}
}

func TestTradeGoodsIncome(t *testing.T) {
	// GAME_MANUAL.pdf p.70:一般種族 2 產能換 1 BC;Fantastic Trader 1 產能換 1 BC。
	cases := []struct {
		industryAllocated int
		fantasticTrader   bool
		want              int
	}{
		{10, false, 5},
		{11, false, 5}, // 11/2=5.5 無條件捨去 → 5
		{0, false, 0},
		{10, true, 10},
		{7, true, 7},
	}
	for _, c := range cases {
		if got := TradeGoodsIncome(c.industryAllocated, c.fantasticTrader); got != c.want {
			t.Errorf("TradeGoodsIncome(%d, %v) = %d,預期 %d", c.industryAllocated, c.fantasticTrader, got, c.want)
		}
	}
}

func TestIncomeFoodSurplusRevenue(t *testing.T) {
	// GAME_MANUAL.pdf p.25:一般種族每單位剩餘糧食 0.5 BC;Fantastic Trader 1 BC。
	cases := []struct {
		surplusFoodUnits int
		fantasticTrader  bool
		want             int
	}{
		{4, false, 2},
		{5, false, 2}, // 5*0.5=2.5 無條件捨去 → 2
		{0, false, 0},
		{5, true, 5},
	}
	for _, c := range cases {
		if got := IncomeFoodSurplusRevenue(c.surplusFoodUnits, c.fantasticTrader); got != c.want {
			t.Errorf("IncomeFoodSurplusRevenue(%d, %v) = %d,預期 %d", c.surplusFoodUnits, c.fantasticTrader, got, c.want)
		}
	}
}

func TestIncomeCommandOverflowCost(t *testing.T) {
	// GAME_MANUAL.pdf p.169:每一點未覆蓋的指揮評等需求,每回合扣 10 BC。
	cases := []struct{ uncoveredCommandPoints, want int }{
		{3, 30},
		{0, 0},
		{-2, 0}, // 沒有超支
	}
	for _, c := range cases {
		if got := IncomeCommandOverflowCost(c.uncoveredCommandPoints); got != c.want {
			t.Errorf("IncomeCommandOverflowCost(%d) = %d,預期 %d", c.uncoveredCommandPoints, got, c.want)
		}
	}
}

func TestIncomeFreighterMaintenanceCost(t *testing.T) {
	// GAME_MANUAL.pdf p.169:每艘使用中的運輸艦每回合維護費 0.5 BC。
	cases := []struct{ activeFreighters, want int }{
		{5, 2},  // 一組運輸艦隊(5 艘):5*0.5=2.5 無條件捨去 → 2
		{10, 5}, // 兩組運輸艦隊:10*0.5=5
		{1, 0},  // 1*0.5=0.5 無條件捨去 → 0
		{0, 0},
	}
	for _, c := range cases {
		if got := IncomeFreighterMaintenanceCost(c.activeFreighters); got != c.want {
			t.Errorf("IncomeFreighterMaintenanceCost(%d) = %d,預期 %d", c.activeFreighters, got, c.want)
		}
	}
}

func TestIncomeMoraleAdjustedProduction(t *testing.T) {
	// GAME_MANUAL.pdf p.170:每一格士氣圖示 = 總產出變化 10%。
	cases := []struct{ baseProduction, netMoraleIcons, want int }{
		{100, 2, 120},
		{100, -3, 70},
		{50, 0, 50},
	}
	for _, c := range cases {
		if got := IncomeMoraleAdjustedProduction(c.baseProduction, c.netMoraleIcons); got != c.want {
			t.Errorf("IncomeMoraleAdjustedProduction(%d, %d) = %d,預期 %d", c.baseProduction, c.netMoraleIcons, got, c.want)
		}
	}
}

func TestIncomeApplyGovernmentMoneyBonus(t *testing.T) {
	// MANUAL_150.html:Democracy +50%、Federation +75% BC 收入加成。
	cases := []struct{ baseBC, bonusPercent, want int }{
		{100, IncomeGovtBonusDemocracyMoneyPercent, 150},
		{100, IncomeGovtBonusFederationMoneyPercent, 175},
		{80, 0, 80},
	}
	for _, c := range cases {
		if got := IncomeApplyGovernmentMoneyBonus(c.baseBC, c.bonusPercent); got != c.want {
			t.Errorf("IncomeApplyGovernmentMoneyBonus(%d, %d) = %d,預期 %d", c.baseBC, c.bonusPercent, got, c.want)
		}
	}
}
