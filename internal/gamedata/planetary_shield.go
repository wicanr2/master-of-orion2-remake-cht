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
// ============ 氣候轉換(2026-08-07 補上)============
//
// 三棟的手冊敘述都有這一句,而且用詞略有差異但意思相同:
//
//	Radiation Shield 「**Radiated worlds become Barren** as long as the shield remains in place」
//	Flux Shield      「The existence of a flux shield **converts Radiated climates to Barren**」
//	Barrier Shield   「This shield **converts Radiated climates into Barren** by reducing solar radiation」
//
// ⚠ 輻射護盾那句寫的是 **as long as the shield remains in place**——「維持中」而不是
// 「一次性改造」。**remake 接成一次性的**:建成時走既有的 `shell.applyClimateChange`
// (與地形改造同一支),把 Climate 推到 BARREN 並連帶調整食物與人口上限。
//
// 差別在護盾**被軌道轟炸摧毀之後**:照手冊那顆星該變回 Radiated,remake 不會。
// 這是刻意的選擇而不是疏忽——remake 的建築效果**沒有一個是可逆的**
// (自動工廠被炸掉產能也不會退回去),為了這一棟另建一套「效果可撤銷」的機制,
// 代價遠大於它修正的失真。**寫在這裡,不假裝已經照手冊做。**
//
// ⚠ 只有 **Radiated → Barren** 這一階。其他氣候不受影響——手冊三句都只講 Radiated。
//
// ============ 屏障護盾多的那一句(2026-08-08 已接)============
//
// 屏障護盾多一句「biological weapons cannot enter the planet's atmosphere」。
// 這條曾以「remake 沒有『生物武器』這個分類」為由擱置——**那個理由是誤判的**:
// 缺的不是規則(手冊 p.99 寫得很完整),是「哪些科技算生物武器」這份名單,
// 而名單一直在執行檔裡(`Calc_Tech_Value_` 的 category 表,category 20 恰好兩項)。
//
// 現在接上了,見 `bioweapon.go`(名單 + 機率 + 擲骰)與
// `internal/shell/orbital_bombardment.go`(投放點在軌道轟炸)。

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

// PlanetaryShieldEffectiveClimate 回傳套用行星護盾之後,這個殖民地實際生效的氣候。
//
// 三面護盾任一存在都把 **Radiated 變成 Barren**(手冊三句一致,見檔頭)。其他氣候原樣回傳。
//
// 這支是**查詢用**(UI 顯示「這顆星實際是什麼氣候」);實際的狀態變更走
// `shell.applyClimateChange`,見檔頭關於「一次性 vs 維持中」的說明。
func PlanetaryShieldEffectiveClimate(climate PlanetClimate, buildings map[string]bool) PlanetClimate {
	if climate != RADIATED {
		return climate
	}
	if PlanetaryShieldReduction(buildings) > 0 {
		return BARREN
	}
	return climate
}
