package gamedata

// AdvancedCivilizationExtraPlanetQuota 對應 Num_Adv_Civ_Planets_ @ 0x62BB7。
// 全程使用整數除法；最後減一是排除每位玩家已存在的母星。
func AdvancedCivilizationExtraPlanetQuota(starCount, playerCount int) int {
	if starCount < 0 || playerCount <= 0 {
		return 0
	}
	quota := ((starCount / 2) * 10) / playerCount
	quota = (quota+9)/10 - 1
	if quota < 0 {
		return 0
	}
	return quota
}

// AdvancedCivilizationStartingBC 對應 Orion2.exe 1.31 的 sub_E5832：
// Advanced Civilization 開局依 TRAIT_MONEY raw 值設定國庫，而不是沿用一般開局 50 BC。
// 原版公式是 (raw+2)*100；合法標準值 -1／0／+1 分別得到 100／200／300 BC。
func AdvancedCivilizationStartingBC(moneyRaw int) int {
	return (moneyRaw + 2) * 100
}
