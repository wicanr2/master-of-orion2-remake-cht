package gamedata

// planetary_shield.go:**行星護盾**——三棟建築各自把軌道轟炸的每次攻擊減掉一個定值。
//
// ============ 為什麼先前是 0 ============
//
// `shell.fleetBombardDamage` 的註解一直寫著:
//
//	沒有「行星護盾」資料(damage.go DamageAfterShield 明講「本函式只處理艦對艦,
//	行星護盾情境不適用」),故護盾/裝甲一律視為 0(無防禦)。
//
// 那不是建模選擇,是缺資料。手冊三段各給了一個數字,而且都寫死在同一句型裡:
//
//	Planetary Radiation Shield  「reducing bombardment damage by 5 points」
//	Planetary Flux Shield       「reduces all damage done to the colony from orbit
//	                             by 10 points per attack」
//	Planetary Barrier Shield    「reducing all damage against a planet by 20 points
//	                             per attack」
//
// **「per attack」是逐次攻擊,不是整場轟炸一次。** 所以接的位置是逐發傷害那一行,
// 不是總傷害那一行——這兩個接法在 10 輪齊射下差了一個數量級。
//
// ============ 三棟是互相取代的,不是疊加 ============
//
// 手冊每一段都寫了取代關係:
//
//	「A Planetary Flux Shield **replaces** any Planetary Radiation Shield already
//	 in existence on that world.」
//
// 所以一顆星上實際只會有一面。`PlanetaryShieldReduction` 取**最強的那一面**而不是加總:
// 資料上真的同時出現兩棟時(存檔亂了、或建造流程有洞),取最大值才是還原,加總會憑空變強。
//
// ============ 維護費是第二個來源 ============
//
// | 建築 | 手冊維護費 | remake 建築表(來自原版執行檔 `off_17EB3D + 12`) |
// |---|---|---|
// | Planetary Radiation Shield | 1 BC | 1 |
// | Planetary Flux Shield | 3 BC | 3 |
// | Planetary Barrier Shield | 5 BC | 5 |
//
// 三棟全中。維護費與減傷值出自手冊的同一段文字,而維護費又能對上執行檔的表
// ——**那一段文字是可信的**,減傷值不是孤證。
//
// ============ 沒接的那兩件事 ============
//
// 三棟都寫著「Radiated 氣候轉 Barren」,屏障護盾還多一句「生物武器無法進入大氣層」。
// 前者要接殖民地的氣候欄位、後者要有生物武器這個分類,兩者都不在這一輪的範圍
// ——**只接減傷,而且說明白只接了減傷**,不假裝這三棟已經完整。

// 行星護盾的每次攻擊減傷(GAME_MANUAL.pdf,見檔頭引文)。
const (
	PlanetaryRadiationShieldReduction = 5
	PlanetaryFluxShieldReduction      = 10
	PlanetaryBarrierShieldReduction   = 20
)

// 三棟護盾在建築表裡的中文名(`Buildings` 的 NameZH)。
const (
	BuildingPlanetaryRadiationShield = "行星輻射護盾"
	BuildingPlanetaryFluxShield      = "行星通量護盾"
	BuildingPlanetaryBarrierShield   = "行星屏障護盾"
)

// PlanetaryShieldReduction 回傳這個殖民地每次受到軌道攻擊要扣掉的傷害。
//
// 取**最強的一面**而不是加總:手冊寫明後者取代前者,同時存在是資料異常,
// 加總會讓那個異常變成強化(見檔頭)。
func PlanetaryShieldReduction(buildings map[string]bool) int {
	switch {
	case buildings[BuildingPlanetaryBarrierShield]:
		return PlanetaryBarrierShieldReduction
	case buildings[BuildingPlanetaryFluxShield]:
		return PlanetaryFluxShieldReduction
	case buildings[BuildingPlanetaryRadiationShield]:
		return PlanetaryRadiationShieldReduction
	}
	return 0
}

// PlanetaryShieldedDamage 把一次攻擊的傷害扣掉護盾,不會低於 0。
//
// 抽成函式而不是讓呼叫端自己減,是為了讓「不會打出負傷害」只有一個地方要顧
// ——負傷害會在累加時變成幫敵人補血。
func PlanetaryShieldedDamage(damage, reduction int) int {
	d := damage - reduction
	if d < 0 {
		return 0
	}
	return d
}
