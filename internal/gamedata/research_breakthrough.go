package gamedata

// ResearchBreakthroughChance 對映 sub_E1EC6 @ 0xE1EC6。
// 原版只有累積研究 strictly greater than cost 時才有突破率：
// floor((progress-cost)*100/cost)，夾到 1..100；未超過或成本非正回 0。
func ResearchBreakthroughChance(cost, progress int) int {
	if cost <= 0 || progress <= cost {
		return 0
	}
	chance := (progress - cost) * 100 / cost
	if chance > 100 {
		return 100
	}
	if chance < 1 {
		return 1
	}
	return chance
}

// ResearchBreakthroughSucceeded 對映 sub_E44E0 @ 0xE450C..0xE4518。
// sub_1247A0(100) 的值域是 1..100，原版以 roll <= chance 成功。
func ResearchBreakthroughSucceeded(chance, roll int) bool {
	return chance > 0 && roll >= 1 && roll <= 100 && roll <= chance
}
