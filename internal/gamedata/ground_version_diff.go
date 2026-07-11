package gamedata

// ground_version_diff.go:2026-07-11 補實作 docs/tech/version-1.3-1.5-diff.md §1 表 #5/#6/#7/
// #8/#9/#11(地面戰/軌道轟炸的 patch 1.3 vs 1.5 差異項)。逐項近似 + 誠實標記,詳見
// docs/tech/ground-combat-algorithm.md「2026-07-11 版本差異補實作」節。
//
// 分類原則(依 version-1.3-1.5-diff.md 的既有結論):
//   - #6(指揮官基礎倍率)、#8(civilian_armor 100hp)、#9(地面防禦建築結構倍率)、#11(轟炸行星
//     尺寸幾何,1.5 系列內部自我修正、對 1.3 vs 最終 1.5 不構成差異)皆為「非差異項」——兩版預設
//     值相同,只實作一次,不進 RuleProfile(避免空版本欄位誤導)。
//   - #5(防禦方 Commando 加成)、#7(轟炸建築 +1 hit)是真正的版本差異項,欄位已加進
//     internal/gamedata/ruleprofile.go 的 RuleProfile。

// --- #6 + #5:Commando 領袖地面戰倍率 ---
//
// 來源交叉核對(手冊 + PARAMETERS.CFG,見 docs/tech/version-1.3-1.5-diff.md #5/#6):
//   - PARAMETERS.CFG:2745-2753——
//     「ground_commando_defender_x5:If 1, defender ground troops get commando bonus valued
//     5x & 7.5x skill level, If 0 (default, classic), it's 2x & 3x skill level.」
//     「ground_commando_attacker_x2:If 1, attacker ground troops get commando bonus valued
//     2x & 3x skill level, If 0 (default, classic), it's 5x & 7.5x skill level.」
//     ⇒ 出廠預設(1.3/1.5 皆同,非差異項):攻方 5x/7.5x、守方 2x/3x(依技能階:tier 1 一般/
//     tier ≥2 進階)。
//   - MANUAL_150.html(1.50 patch notes):「A defending commando gives 2.5x the regular
//     commando bonus to ground troops, just like an attacking commando already gives in
//     classic.」⇒ 「2x/3x」是「regular commando bonus」基準值本身,攻方在兩版都已經是
//     「2.5x 該基準值」(2.5×2=5、2.5×3=7.5,與 CFG 攻方數字完全吻合);1.3 守方沒有這個
//     2.5x 加乘、維持基準值 2x/3x 不變,1.5 起守方追平攻方,也套用 2.5x 加乘變成 5x/7.5x。
//
// ⚠ 近似(誠實標記,非手冊逐字精確值):
//  1. 「regular commando bonus」基準值本身(2/3)手冊/CFG 只以「skill level」的相對倍率描述,
//     沒有給出獨立的「這是對應哪個絕對戰力刻度」說明;本檔直接把 2/3(tier1/tier2)當成最終
//     force 加成點數本身(與 GroundArmorTechBonus 等其他加成同單位:combat rating 加成點),
//     不再另外乘一個未知的獨立基準值——這是求可實作/可測試的簡化,非手冊給出的獨立驗證值。
//  2. 「領袖指派到某次入侵」remake 沒有對應模型(Leaders 是帝國全域清單,無「指派到哪支艦隊/
//     哪個殖民地」欄位),故呼叫端(internal/shell/ground_invasion.go commandoLeaderTier)用
//     「帝國是否擁有 Commando 技能領袖」當代理條件,不論該領袖實際所在艦隊/殖民地。
//  3. 7.5 捨去到整數 7(Go int 之間運算,見下方函式的 int() 截斷),既有加成表(GroundArmorTechBonus
//     等)本來就都是整數點數,沒有 0.5 的精度,故沿用整數慣例。

// GroundCommandoAttackerForceBonus 回傳攻方 Commando 領袖(依技能階 tier:0 無/1 一般/2 進階)
// 對地面戰 force 的加成。兩版預設相同(非差異項,見上方來源交叉核對)。
func GroundCommandoAttackerForceBonus(tier int) int {
	switch {
	case tier >= 2:
		return 7 // 2.5 × 3 = 7.5,捨去至 7
	case tier == 1:
		return 5 // 2.5 × 2 = 5
	default:
		return 0
	}
}

// GroundCommandoDefenderForceBonus 回傳守方 Commando 領袖對地面戰 force 的加成,版本相依:
// defenderCommandoBonus 由呼叫端傳入 RuleProfile.DefenderCommandoBonus(1.3=1.0,即維持
// 「regular commando bonus」基準值 2/3 不變;1.5=2.5,追平攻方的 2.5x 加乘變成 5/7.5→7)。
func GroundCommandoDefenderForceBonus(tier int, defenderCommandoBonus float64) int {
	var base float64
	switch {
	case tier >= 2:
		base = 3
	case tier == 1:
		base = 2
	default:
		return 0
	}
	return int(base * defenderCommandoBonus)
}

// --- #9:地面防禦建築結構倍率(ground_defense_armor_multiplier) ---
//
// PARAMETERS.CFG:1772-1775「Default is 100 ... (classic).」出廠預設兩版相同,非差異項。
// 對應手冊地面防禦建築(地面砲台/飛彈基地,以鈦裝甲 Tritanium Armor 為基準結構點)。
//
// ⚠ 掛鉤備妥、待防禦建築系統(誠實標記,非藉口不做):本 remake 完全沒有「地面防禦建築」這個
// 資料實體——ColonyBuildings 只是 map[string]bool 的有/無旗標,無 HP/結構值欄位;AI 側連
// ColonyBuildings 追蹤都沒有(見 internal/shell/ground_invasion.go InvadeColony 註解「守方
// 戰車 TODO 未接」的同款理由)。要真正套用這個倍率,需要先幫地面防禦建築建一個「結構點」資料
// 模型(哪些建築算「地面防禦」、初始 HP、如何被轟炸/入侵削減),這超出本輪「補版本差異數值」
// 的範圍,不臆測發明資料結構。本常數先鎖定手冊/CFG 數字本身,供未來該資料模型完成後直接引用。
const GroundDefenseArmorMultiplier = 100

// --- #8:civilian_armor(轟炸建築/人口裝甲值) ---
//
// PARAMETERS.CFG:1778-1786「Default is 100 hp regardless of armor (classic).」出廠預設兩版
// 相同,非差異項。與 #7(BombardmentBuildingBonusHits,見 ruleprofile.go)同屬「轟炸建築模型」
// 的一部分。
//
// ⚠ 掛鉤備妥、待建築損傷模型(誠實標記):本 remake 軌道轟炸目前只扣人口(見
// internal/shell/orbital_bombardment.go BombardColony「範圍限制」),不扣建築——AI 沒有
// ColonyBuildings 持久資料可扣,扣了會是憑空生資料。本常數先鎖定數字本身,供未來建築損傷模型
// 完成後直接引用,不臆測提前套用。
const GroundCivilianArmorHP = 100

// --- #11:轟炸行星尺寸幾何(3-4-6-7-8) ---
//
// CHANGELOG_150.TXT 1.50.11「Restored planet sizes for bombardment to classic 3-4-6-7-8.」
// 1.5 系列中途(1.50.4 起)曾錯改,1.50.11 已修回同一組數字,對 1.3 vs 最終 1.5 而言不構成
// 版本差異(非差異項),只實作一次。

// GroundPlanetSizeBombardCoefficient 回傳手冊/CHANGELOG「classic 3-4-6-7-8」對應各行星尺寸
// 的幾何係數:Tiny=3/Small=4/Medium=6/Large=7/Huge=8。未知尺寸(理論上不會發生,PlanetSize
// 只有 5 個列舉值)退回 Medium 基準,不臆測。
func GroundPlanetSizeBombardCoefficient(size PlanetSize) int {
	switch size {
	case TINY_PLANET:
		return 3
	case SMALL_PLANET:
		return 4
	case MEDIUM_PLANET:
		return 6
	case LARGE_PLANET:
		return 7
	case HUGE_PLANET:
		return 8
	default:
		return 6
	}
}

// GroundBombardPopulationLoss 依行星尺寸係數把「轟炸命中數(hits)」換算成實際扣減的人口單位數
// (大行星較耐轟)。
//
// ⚠ 近似(誠實標記,非手冊精確公式):手冊/CHANGELOG 只給出這組尺寸係數本身(3-4-6-7-8),沒有
// 給「係數如何代入人口損傷公式」的精確算式——原版這組係數很可能牽涉行星地圖網格/區域數(每個
// tile 各自累積 hits),但 remake 沒有這層行星地圖模型,無法逐 tile 模擬。本函式採「係數與損傷
// 成反比、以 Medium(6)為基準 1:1」的最簡單合理近似:
//
//	loss = hits × 6 / coefficient(size)
//
// 即 Medium 維持現行行為(loss==hits,與換係數前的既有邏輯一致),Tiny/Small 係數較小 → loss
// higher(較脆弱),Large/Huge 係數較大 → loss 較低(較耐轟)。結果捨去至整數(向下取整),至少 0。
// hits<=0 時直接回 0。
func GroundBombardPopulationLoss(hits int, size PlanetSize) int {
	if hits <= 0 {
		return 0
	}
	const baseline = 6 // Medium 基準,對應既有(換係數前)loss==hits 的行為
	coef := GroundPlanetSizeBombardCoefficient(size)
	if coef <= 0 {
		coef = baseline
	}
	loss := hits * baseline / coef
	if loss < 0 {
		loss = 0
	}
	return loss
}
