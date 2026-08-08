package gamedata

// bioweapon.go:**生物武器**——軌道轟炸時直接殺人口,而屏障護盾整個擋掉。
//
// ============ 這一條為什麼卡了那麼久 ============
//
// `planetary_shield.go` 的檔尾一直掛著:
//
//	屏障護盾多一句「biological weapons cannot enter the planet's atmosphere」。
//	remake 沒有「生物武器」這個分類,**這條沒接**。
//
// 缺的是**分類**,不是規則。而那個分類 2026-08-08(第 52 項(生物武器分類))從執行檔挖出來了:
// `Calc_Tech_Value_` 用的 category 表裡,**category 20 就是生物武器**,成員恰好兩個
// ——`Bio-Terminator` 與 `Death Spores`(見 orig_tech_value_tables.go 的 enum 語意表)。
//
// 分類是執行檔給的,不是自己劃的。
//
// ============ 規則手冊給得很完整 ============
//
// GAME_MANUAL.pdf p.99「Death Spores (System)」:
//
//	They are so contagious and deadly that invading ships must introduce them into the
//	target planet's atmosphere **by orbital bombardment**. Each spore pod launched has a
//	**10% chance to kill one unit of colonist population**.
//
// 而手冊後面那張表給了兩項的機率:**Death Spores 10%、Bio-Terminator 20%**。
//
// 三件事因此都有依據:**投放方式**(軌道轟炸)、**效果**(每莢 10%/20% 殺 1 單位人口)、
// **反制**(屏障護盾「無法進入大氣層」= 完全擋掉)。
//
// ============ 誠實留白:一次轟炸投幾莢 ============
//
// 手冊說的是「每一個發射出去的孢子莢」,但**沒說一次轟炸投幾莢**,而 remake 也沒有
// 「哪幾艘船掛了生物武器、各帶幾莢」的模型。呼叫端取「艦隊艦艇數」當莢數
// ——**那是 remake 的建模選擇,不是手冊數字**,寫在呼叫端的註解裡。

// 生物武器每一莢殺死一單位人口的機率(百分比,GAME_MANUAL.pdf 的效果表)。
const (
	BioWeaponDeathSporesKillPercent   = 10
	BioWeaponBioTerminatorKillPercent = 20
)

// BiologicalWeaponKillPercent 回傳某項生物武器每莢的殺傷機率;不是生物武器回 0。
//
// 「是不是生物武器」查的是執行檔的 category 表(`IsBiologicalWeapon`),不是寫死名單
// ——這樣哪天 category 20 多了一項,這裡會自動認得它是生物武器(機率仍要另外補,
// 回 0 代表「知道它是生物武器,但沒有它的數字」,不會靜默當成一般武器)。
func BiologicalWeaponKillPercent(tech Technology) int {
	if !IsBiologicalWeapon(tech) {
		return 0
	}
	switch tech {
	case TECH_DEATH_SPORES:
		return BioWeaponDeathSporesKillPercent
	case TECH_BIOTERMINATOR:
		return BioWeaponBioTerminatorKillPercent
	}
	return 0
}

// BestBiologicalWeaponKillPercent 回傳這一方**最強**生物武器的殺傷機率。
//
// has 由呼叫端提供(判某項科技是否已擁有)。取最強而不是加總——一次轟炸投的是同一種莢。
func BestBiologicalWeaponKillPercent(has func(Technology) bool) int {
	best := 0
	for _, tech := range []Technology{TECH_DEATH_SPORES, TECH_BIOTERMINATOR} {
		if !has(tech) {
			continue
		}
		if p := BiologicalWeaponKillPercent(tech); p > best {
			best = p
		}
	}
	return best
}

// BiologicalWeaponBlocked 回報這個殖民地的建築是否擋得住生物武器。
//
// 手冊給屏障護盾那句:「biological weapons **cannot enter the planet's atmosphere**」
// ——是**完全擋掉**,不是減傷。所以這裡回布林而不是百分比。
//
// ⚠ 只有**屏障**護盾有這一句。輻射護盾與通量護盾的手冊敘述都沒有,所以它們不擋
// ——那不是漏寫,是三段文字的實際差異(見 planetary_shield.go 的三段引文)。
func BiologicalWeaponBlocked(buildings map[string]bool) bool {
	return buildings[BuildingPlanetaryBarrierShield]
}

// BiologicalWeaponPopKills 擲骰算出這次轟炸的生物武器殺了幾單位人口。
//
// pods 是投放的孢子莢數,killPercent 是每莢的殺傷機率。
// roll 傳 `rng.Intn`(回 [0,n)),每莢擲一次。
func BiologicalWeaponPopKills(pods, killPercent int, roll func(n int) int) int {
	if pods <= 0 || killPercent <= 0 {
		return 0
	}
	kills := 0
	for i := 0; i < pods; i++ {
		if roll(100) < killPercent {
			kills++
		}
	}
	return kills
}
