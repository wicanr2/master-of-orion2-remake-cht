package gamedata

// AI 行星估值(原版 `Uncolonized_Planet_Worth_To_Player_` @ 0xD27A7,module 92)。
//
// remake 的 AI 選星先前是「掃過星圖,遇到第一顆能殖民的就佔」——完全不看行星好壞。
// 原版有一整套估值:`Compute_Base_Planet_Values_` 每回合替每個玩家把全星圖 360 顆行星
// 各算一個 16-bit 分數存進 `_g_base_planet_values`(8 玩家 × 720 bytes = 5760),
// AI 據此挑目標。
//
// 欄位語意怎麼確定的:反編出來的 `a1[4]`、`a1[6]`、`a1[8]`… 逐一對上 openorion2 的
// `struct Planet`(type/size/gravity/climate/minerals/foodbase/special),而該結構的
// **17 bytes 大小**正好等於反編迴圈裡的 stride(`17 * planetIndex`),互相印證。
// 氣候編碼(TOXIC=0…GAIA=9)也與 remake 既有的 PlanetClimate 相同。
//
// 原版公式(重建;變數名依 openorion2 結構欄位還原):
//
//	if planet.type != HABITABLE            → 0
//	maxPop   = 玩家可用的人口上限(含地下城市加成、種族 +5)
//	minePer  = minerals_per_mine[planet.minerals]        // 1/2/3/5/8
//	combined = maxPop * (minePer + 1) / 3
//	v        = 依 AI 目標傾向加權(combined 與 minePer 的四種配比)
//	v       += 食物基礎 × (150 或 75)
//	v       += 特殊物產加分(special 4/5/10/11 → 1280/2560/640/1600)
//	v        = maxPop * v * (minePer + 7) / 10
//	v        = v * (100 - climate_maintenance[climate]/4) / 100
//	           海洋/沼澤(需要額外科技才好用)再 × 3/4
//	v        = 依重力與種族重力天賦打折
//	return min(v >> 6, 65535)
//
// ⚠ remake 沒有建模的部分,已在下方逐處標明並取中性值:種族的地下城市科技、
// Low-G / Heavy-G 天賦、以及讓食物加分減半的那個旗標。這些會讓分數的絕對值與原版有落差,
// 但**行星之間的相對排序**(AI 該挑哪顆)才是這個函式的用途,核心項都在。

// AIObjective 是 AI 的目標傾向,決定行星估值裡「人口」與「礦產」的權重配比。
// 原版存在玩家結構偏移 2208,實測有 4 個取值(0 / 50 / 100 / 206),各對應一組權重。
type AIObjective int

const (
	// AIObjectiveMineral 最重礦產、最不看人口(原版值 206:combined×1 + minePer×80)。
	AIObjectiveMineral AIObjective = iota
	// AIObjectiveBalancedLow 原版值 0:combined×2 + minePer×75。
	AIObjectiveBalancedLow
	// AIObjectiveBalancedHigh 原版值 50:combined×3 + minePer×70。
	AIObjectiveBalancedHigh
	// AIObjectivePopulation 最重人口(原版值 100:combined×4 + minePer×65)。
	AIObjectivePopulation
)

// aiObjectiveWeights[objective] = {人口綜合量權重, 每礦工工業權重}。
// 四組都是原版硬編值(見 AIObjective 各常數註解)。
var aiObjectiveWeights = [4][2]int{
	{1, 80}, // Mineral
	{2, 75}, // BalancedLow
	{3, 70}, // BalancedHigh
	{4, 65}, // Population
}

// climateMaintenanceModifiers 是各氣候的維護成本修正(原版 `_climate_maintenance_modifiers`
// @ cseg01 0xDD4BA,10 bytes,索引 = PlanetClimate)。
//
// 值本身反直覺(Barren 是 0、Desert 卻是 25),但兩次獨立 dump 一致,且反編的
// `climate_maintenance_modifiers[planet.climate]` 索引方式明確,照抄不臆改。
var climateMaintenanceModifiers = [10]int{50, 25, 0, 25, 0, 0, 0, 0, 0, 0}

// ClimateMaintenanceModifier 回傳該氣候的維護成本修正。
func ClimateMaintenanceModifier(c PlanetClimate) int {
	if c < 0 || int(c) >= len(climateMaintenanceModifiers) {
		return 0
	}
	return climateMaintenanceModifiers[c]
}

// AIPlanetValueInput 是估一顆未殖民行星所需的資料。
type AIPlanetValueInput struct {
	Habitable bool // planet.type == HABITABLE;false 一律 0 分(氣態巨星/小行星帶)
	MaxPop    int  // 該玩家在這顆行星的人口上限
	Minerals  PlanetMinerals
	Climate   PlanetClimate
	Gravity   PlanetGravity
	FoodBase  int // planet.foodbase:該行星的食物基礎
	Special   int // planet.special:特殊物產代碼(4/5/10/11 有加分)
	// RaceLowG / RaceHeavyG 是該玩家種族的重力天賦(原版玩家結構偏移 2217/2218)。
	// 兩者都 false = 一般 Normal-G 種族。
	RaceLowG, RaceHeavyG bool
}

// aiSpecialBonus 是行星特殊物產的固定加分(原版硬編的四個 case)。
// 其餘 special 代碼沒有加分——不是漏列,原版就只判這四個。
func aiSpecialBonus(special int) int {
	switch special {
	case 4:
		return 1280
	case 5:
		return 2560
	case 10:
		return 640
	case 11:
		return 1600
	}
	return 0
}

// AIPlanetValue 依原版公式算一顆未殖民行星對某個 AI 的價值(0..65535)。
// 分數只用來排序候選目標,絕對值不具獨立意義。
func AIPlanetValue(in AIPlanetValueInput, obj AIObjective) int {
	if !in.Habitable || in.MaxPop <= 0 {
		return 0
	}
	minePer := MineralIndustryPerWorker(in.Minerals)
	combined := in.MaxPop * (minePer + 1) / 3

	o := int(obj)
	if o < 0 || o >= len(aiObjectiveWeights) {
		o = int(AIObjectiveBalancedLow)
	}
	w := aiObjectiveWeights[o]
	v := w[0]*combined + w[1]*minePer

	// 食物加分。原版還有一個旗標會把 150 減半成 75(玩家結構偏移 2224),
	// remake 沒有對應欄位,取未減半的 150。
	v += 150 * in.FoodBase

	v += aiSpecialBonus(in.Special)

	v = in.MaxPop * v * (minePer + 7) / 10

	// 氣候維護成本折扣。海洋/沼澤在原版還會再 ×3/4——那是「需要額外科技才好用」的懲罰,
	// 有該科技時整段跳過;remake 沒有那個科技旗標,一律套用(對排序的影響是一致的)。
	v = v * (100 - ClimateMaintenanceModifier(in.Climate)/4) / 100
	if in.Climate == OCEAN || in.Climate == SWAMP {
		v = v * 3 / 4
	}

	// 重力:與種族重力天賦的搭配決定折扣(原版四個分支,見檔頭)。
	switch in.Gravity {
	case LOW_G:
		if !in.RaceLowG {
			v = v * 18 / 20
		}
	case NORMAL_G:
		if in.RaceLowG {
			v = v * 18 / 20
		}
	case HEAVY_G:
		if !in.RaceHeavyG {
			v = v * 4 / 6
		}
	}

	v >>= 6
	if v > 65535 {
		return 65535
	}
	if v < 0 {
		return 0
	}
	return v
}
