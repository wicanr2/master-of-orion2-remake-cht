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

// --- 第二層:鄰近價值(原版 `Proximity_Worth_To_Player_` @ 0xD2AEA)---
//
// 原版對每顆行星掃一遍射程內的恆星,依距離倒數累加:
//
//	自己的星:   worth += (該星我已探索 ? 120 : 80) / distance
//	別人的星:   worth += 100 / distance,並記錄最近距離
//
// 也就是「離我的地盤越近越值錢」。remake 先前的 AI 完全不看距離,會跑到星圖另一端殖民。

// AIProximityOwnWeight / AIProximityUnknownWeight / AIProximityOtherWeight 是原版的三個權重
// (見上方公式)。原版用「該星的旗標是否含我的 player bit」區分前兩者。
const (
	AIProximityOwnWeight     = 120
	AIProximityUnknownWeight = 80
	AIProximityOtherWeight   = 100
)

// AIProximityValue 依距離倒數累加鄰近價值。
// distances 是候選星到「每一顆我方已佔星」的距離(同一套單位即可,原版用的是預算好的距離表)。
// 距離 0 視為 1,避免除以零。
func AIProximityValue(distances []int) int {
	sum := 0
	for _, d := range distances {
		if d <= 0 {
			d = 1
		}
		sum += AIProximityOwnWeight / d
	}
	return sum
}

// --- 第三層:星系內協同(原版 `Compute_Contextual_Planet_Values_` @ 0xD3125)---
//
// 原版對每顆行星掃「同一恆星系內的其他行星」,依它們的歸屬調整這顆的分數:
//
//	contextual = base
//	           + (同系無主行星的 base 總和) / 8
//	           + (同系我方殖民地的 base 總和) / 4        // 只在這顆已有殖民地時
//	若這顆是空的且同系已有我方殖民地 → contextual = contextual * (size+1) / 10
//	若同系有敵方殖民地 n 個          → contextual /= (n + 2)
//
// 語意是「整系開發有加成、敵方已進駐的星系要避開」。
//
// ⚠ remake 是「一星一行星」(見 shell.genPlanets 註解),沒有同系兄弟行星。
// 這裡把「同一恆星系」映射成「鄰近的星」——結構與係數照原版,鄰近的定義是 remake 的。
type AIContextualInput struct {
	Base           int // 第一層 AIPlanetValue 的結果
	NeighborEmpty  int // 鄰近無主星的 base 值總和
	NeighborOwn    int // 鄰近我方殖民地的 base 值總和
	NeighborOwnN   int // 鄰近我方殖民地數
	NeighborEnemyN int // 鄰近敵方殖民地數
	Size           PlanetSize
	Colonized      bool // 這顆星本身是否已有殖民地
}

// AIContextualPlanetValue 套用星系內協同效應,回傳最終估值(夾在 0..65535)。
func AIContextualPlanetValue(in AIContextualInput) int {
	v := in.Base
	emptyBonus := in.NeighborEmpty / 8

	if in.Colonized {
		v += emptyBonus
		v += in.NeighborOwn / 4
	} else {
		if in.NeighborOwnN <= 0 {
			v += emptyBonus
		} else {
			// 已有我方鄰居:依行星大小縮放(大星才值得再開一個殖民地)。
			v = v * (int(in.Size) + 1) / 10
		}
	}
	if in.NeighborEnemyN > 0 {
		v /= in.NeighborEnemyN + 2
	}
	if v > 65535 {
		return 65535
	}
	if v < 0 {
		return 0
	}
	return v
}

// --- 第四層:已有殖民地的價值(原版 `Colony_Worth_To_Player_` @ 0xD2CAE)---
//
// 前三層評的是「無主星值不值得去佔」;這一層評的是「一個**已經有殖民地**的星值多少」——
// AI 用它挑攻擊目標、評估外交籌碼。兩者算法不同:未殖民星用礦產推估潛力,已殖民星直接用
// 該殖民地**目前的實際產出**。
//
// 原版公式(變數名依 openorion2 結構欄位還原):
//
//	if colony.is_outpost                       → 只算鄰近價值(前哨站沒有產出)
//	avgMaxPop = (該殖民地主人的人口上限 + 評估者的人口上限 + 1) / 2
//	prod      = 6*食物 + 3*(工業 + 研究)      // 兩個種族旗標會改權重,見下方註記
//	v         = (w0*目前人口 + w1*avgMaxPop) * prod * (avgMaxPop + 100 - 目前人口) / 100
//	v         = v * (100 - climate_maintenance[climate]/4) / 100
//	            海洋/沼澤再 × 3/4
//	v         = 依重力與種族重力天賦打折
//	v        += 特殊物產加分(金礦 1000、寶石礦 2000)
//	return (v >> 6) + 鄰近價值
//
// `(avgMaxPop + 100 - 目前人口)` 這一項是「還有多少成長空間」:人口越接近上限,這顆星對
// 侵略者的邊際價值越低。
//
// ⚠ remake 沒有建模、因而取中性值的部分(與 AIPlanetValue 同一批):
//   - 玩家結構偏移 2224 / 2225 兩個種族旗標會把產出權重改成 `4*(食+工+研)` 或 `6*(工+研)`
//     (後者等於「食物毫無價值」,推測是不吃食物的種族);remake 取預設的 `6*食 + 3*(工+研)`。
//   - 偏移 410 == 3 這個條件會讓重力懲罰輕一格(12→13/16、6→7/12);remake 取未緩和的值。
//   - 偏移 2207 的種族等級(值 0/1/2 → ×3/2、×4/3、×6/5)與殖民地偏移 319 的旗標搭配使用,
//     兩者語意都還沒釘死,remake 不套用——寧可少一個乘數,也不套用語意不明的數字。

// aiColonyObjectiveWeights[objective] = {目前人口權重, 平均人口上限權重}。
// 四組都是原版硬編值(`Colony_Worth_To_Player_` 依玩家結構偏移 2208 分派):
// 206→{95,5}、0→{90,10}、50→{85,15}、100→{80,20}。索引順序同 AIObjective。
//
// 注意這與 aiObjectiveWeights(未殖民星)是**不同的兩組數字**,不是同一張表被兩處共用——
// 原版就是各自硬編一套。
var aiColonyObjectiveWeights = [4][2]int{
	{95, 5},  // Mineral(原版值 206)
	{90, 10}, // BalancedLow(原版值 0)
	{85, 15}, // BalancedHigh(原版值 50)
	{80, 20}, // Population(原版值 100)
}

// aiColonySpecialBonus 是已殖民星的特殊物產加分(原版硬編:金礦 1000、寶石礦 2000)。
// 與未殖民星的 1280 / 2560 是不同數字,但**同樣是 1:2**,也同樣只判金礦/寶石礦。
func aiColonySpecialBonus(special int) int {
	switch special {
	case 4:
		return 1000
	case 5:
		return 2000
	}
	return 0
}

// AIColonyValueInput 是估一顆「已有殖民地」的行星所需的資料。
type AIColonyValueInput struct {
	IsOutpost  bool // colony.is_outpost:前哨站沒有產出,原版直接跳到只算鄰近價值
	Population int  // 該殖民地目前人口
	MaxPop     int  // (該殖民地主人的人口上限 + 評估者的人口上限 + 1) / 2
	Food       int  // 該殖民地的食物產出(原版取半單位再除 2,傳入時已是整單位)
	Industry   int  // 工業產出
	Research   int  // 研究產出
	Climate    PlanetClimate
	Gravity    PlanetGravity
	Special    int // planet.special(4/5 有加分)
	// RaceLowG / RaceHeavyG 是評估者種族的重力天賦(同 AIPlanetValueInput)。
	RaceLowG, RaceHeavyG bool
}

// AIColonyValue 依原版公式算一顆已殖民行星對某個 AI 的價值(0..65535)。
// 前哨站回 0(原版走的是「只算鄰近價值」那條路,鄰近價值由呼叫端另外加)。
func AIColonyValue(in AIColonyValueInput, obj AIObjective) int {
	if in.IsOutpost || in.Population <= 0 {
		return 0
	}
	o := int(obj)
	if o < 0 || o >= len(aiColonyObjectiveWeights) {
		o = int(AIObjectiveBalancedLow)
	}
	w := aiColonyObjectiveWeights[o]

	// 產出加權(預設分支,見檔頭註記)。
	prod := 6*in.Food + 3*(in.Industry+in.Research)
	if prod < 0 {
		prod = 0
	}

	v := (w[0]*in.Population + w[1]*in.MaxPop) * prod
	// 成長空間:人口越接近上限,對侵略者的邊際價值越低。夾在 0 以上,避免超額人口翻成負值。
	room := in.MaxPop + 100 - in.Population
	if room < 0 {
		room = 0
	}
	v = v * room / 100

	v = v * (100 - ClimateMaintenanceModifier(in.Climate)/4) / 100
	if in.Climate == OCEAN || in.Climate == SWAMP {
		v = v * 3 / 4
	}

	// 重力折扣。注意係數與 AIPlanetValue 不同(那邊 18/20 與 4/6,這邊 12/16 與 6/12)——
	// 原版對「已殖民」的重力懲罰更重,兩處分別硬編,不是同一組常數。
	switch in.Gravity {
	case LOW_G:
		if !in.RaceLowG {
			v = v * 12 / 16
		}
	case NORMAL_G:
		if in.RaceLowG {
			v = v * 12 / 16
		}
	case HEAVY_G:
		if !in.RaceHeavyG {
			v = v * 6 / 12
		}
	}

	v += aiColonySpecialBonus(in.Special)

	v >>= 6
	if v > 65535 {
		return 65535
	}
	if v < 0 {
		return 0
	}
	return v
}

// --- 第五層:敵方殖民地作為攻擊目標的價值(原版 `Enemy_Colony_Worth_To_Player_` @ 0xD8D11)---
//
// ⚠ 位址訂正:先前的缺口報告寫 0xD2B3E,那是錯的。符號表裡這個函式在 **0xD8D11**,
// module 95——與 `Assign_New_Colony_Ship_Destinations_`(0xD896F)、
// `Best_Uncolonized_Undestined_Planet_`(0xD8C79)同一個模組,也就是 AI 決定「艦隊往哪去」
// 的那一支,位置本身就佐證了它的用途。
//
// 這個函式短得出乎意料,而且設計很有意思:
//
//	w = 依「我對這個殖民地主人的外交狀態」決定的一組權重(兩個數字恆和為 6)
//	return (w.owner * 該行星對**主人**的基礎估值
//	      + w.self  * 該行星對**我**的基礎估值) / 6
//
// 兩個估值都取自 `_g_base_planet_values`(8 玩家 × 360 行星 × 2 bytes = 5760,由
// `Compute_Base_Planet_Values_` 每回合逐玩家填好),也就是第一層 AIPlanetValue 的結果。
// 組語裡那個 stride 0x2D0 = 720 bytes = 360 行星 × 2 bytes,正好對上這個維度。
//
// **權重偏向「主人的估值」而不是「我的估值」**——語意是「打他最痛的地方」,不是「搶我最想要
// 的地方」。在最極端的那一檔(狀態 6)甚至是 6:0,完全不看自己想不想要。
//
// 外交狀態編碼取自 openorion2 `enum ForeignPolicy`(0 無/1 互不侵犯/2 同盟/3 和平/
// 4 有限戰爭/5 戰爭,>5 也算戰爭)。原版的分派**不是單調的**:狀態 5(戰爭)與和平以下
// 共用同一組權重,只有 4(有限戰爭)與 6 各自一組。照抄,不「修正」成看起來比較合理的樣子。

// AIForeignPolicy 是外交狀態(openorion2 `enum ForeignPolicy`)。
type AIForeignPolicy int

const (
	DiploNone          AIForeignPolicy = 0
	DiploNonAggression AIForeignPolicy = 1
	DiploAlliance      AIForeignPolicy = 2
	DiploPeace         AIForeignPolicy = 3
	DiploLimitedWar    AIForeignPolicy = 4
	DiploWar           AIForeignPolicy = 5
	DiploTotalWar      AIForeignPolicy = 6 // openorion2 只說「>5 也算戰爭」;原版對 6 另有一組權重
)

// AIEnemyTargetWeightSum 是兩個權重的總和(原版最後除以 6)。
const AIEnemyTargetWeightSum = 6

// aiEnemyTargetWeights 回傳 (主人估值權重, 自己估值權重),兩者恆和為 6。
func aiEnemyTargetWeights(policy AIForeignPolicy) (owner, self int) {
	switch {
	case policy < DiploLimitedWar:
		return 5, 1
	case policy == DiploLimitedWar:
		return 4, 2
	case policy == DiploTotalWar:
		return 6, 0
	default: // 含 DiploWar(5) 與 >6:原版落在同一個 default
		return 5, 1
	}
}

// AIEnemyColonyValue 算一個敵方殖民地作為攻擊目標的價值。
//
// ownerValue / selfValue 分別是該行星對「殖民地主人」與「評估者」的基礎估值
// (= AIPlanetValue 的結果,原版存在 `_g_base_planet_values`)。
//
// shiftToSelf 對應原版那個「評估者有、而殖民地主人沒有」的種族旗標(玩家結構偏移 0x8B8):
// 成立時把權重往自己的估值挪一格(owner-1 / self+1)。該旗標語意尚未釘死,remake 一律傳
// false——寧可少一個修正,也不套用語意不明的條件。
func AIEnemyColonyValue(ownerValue, selfValue int, policy AIForeignPolicy, shiftToSelf bool) int {
	wOwner, wSelf := aiEnemyTargetWeights(policy)
	if shiftToSelf && wOwner > 0 {
		wOwner--
		wSelf++
	}
	v := (wOwner*ownerValue + wSelf*selfValue) / AIEnemyTargetWeightSum
	if v < 0 {
		return 0
	}
	if v > 65535 {
		return 65535
	}
	return v
}
