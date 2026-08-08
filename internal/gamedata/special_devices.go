package gamedata

// special_devices.go:**原版執行檔的特殊裝置表**——39 項 × (空間 / 成本) × 6 個艦體等級。
//
// ============ 這張表從哪裡來 ============
//
// `Orion2.exe` 的資料段 `0x17EEE0` 起,每筆 47 (0x2F) 位元組:
//
//	+0  dd  名稱字串指標
//	+4  dw  裝置編號(= SpecialDevices 列舉值,0..39 連號)
//	+6  dw  解鎖科技編號(= Technology 列舉值)
//	+8  6×dw  **空間**,依艦體等級 巡防/驅逐/巡洋/戰艦/泰坦/末日之星
//	+20 6×dw  **成本**,同上六級
//
// 欄位定位不是猜的,三個獨立訊號指同一組偏移:
//
//  1. `Special_Devices_Available_`(0x5F9EA)讀 `word_17EEE6[i*0x2F]`(= +6)當成
//     `[帝國紀錄 + 0x117 + 科技編號] == 3` 的索引 → **+6 是科技編號**。
//  2. `Init_Internal_Space_`(0x36470)讀 +8 那一欄 → **+8 起是空間**。
//  3. 戰鬥艙那一列的 +8 欄是**負數**(−12/−30/−60/−125/−250/−600),而手冊寫
//     「add equipment space without increasing the hull size」——負佔格正是「加空間」的實作。
//
// ============ 兩道交叉驗證(都是事後才發現對得上,不是湊出來的)============
//
//   - **39 項的科技編號逐項對上 remake 既有的 `TECH_*` 列舉,零不符。**
//     那份列舉是先前從別的來源建立的,兩邊獨立。
//   - 戰鬥艙的負佔格剛好是 `ShipHullSpace` 的**一半**(25/60/120/250/500/1200 → 12/30/60/125/250/600,
//     12 = floor(12.5))。而 `ShipHullSpace` 是從**手冊 p.121** 抄的。兩份來源在小數點的截斷上
//     都一致,不可能是巧合。
//
// ============ 這張表取代了什麼 ============
//
// `shell.SpecialOptions` 每一列的成本先前都標著「remake 值:手冊行文沒給系統的建造成本,
// 執行檔的元件表還沒挖到」。**那句「還沒挖到」現在過期了。**
// 佔格則先前走 `SpecialSpace()` 的 `SpecialSpaceEstimatePercent` 估計值。
//
// ⚠ 原版的空間與成本**都隨艦體等級變動**,不是單一數字——這是與 remake 舊模型最大的差異:
// 同一個系統裝在末日之星上比裝在巡防艦上貴 25 倍、佔 25 倍空間。

// SpecialDeviceStats 是一項特殊裝置在原版表裡的三個欄位。
type SpecialDeviceStats struct {
	// Tech 是解鎖它的科技(執行檔 +6 欄)。
	Tech Technology
	// Space 是佔用的艦體空間,依 CombatShipClass 索引(0=巡防艦 … 5=末日之星)。
	// **負值代表這個系統反而增加可用空間**(目前只有戰鬥艙)。
	Space [6]int
	// Cost 是建造成本(BC),同樣依艦體等級索引。
	Cost [6]int
}

// specialDeviceTable 是原版表的逐筆搬運。**不要手改任何數字**——要改先回頭讀
// `docs/tech/special-device-table.md` 記的抽取方式,重跑一次。
var specialDeviceTable = map[SpecialDevices]SpecialDeviceStats{
	SPEC_ACHILLES_TARGETING_UNIT: {TECH_ACHILLES_TARGETING_UNIT, [6]int{10, 15, 25, 50, 100, 250}, [6]int{15, 22, 37, 75, 150, 375}},
	SPEC_AUGMENTED_ENGINES:       {TECH_AUGMENTED_ENGINES, [6]int{10, 15, 25, 50, 100, 250}, [6]int{5, 7, 12, 25, 50, 125}},
	SPEC_AUTOMATED_REPAIR_UNIT:   {TECH_AUTOMATED_REPAIR_UNIT, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_BATTLE_PODS:             {TECH_BATTLE_PODS, [6]int{-12, -30, -60, -125, -250, -600}, [6]int{20, 60, 175, 400, 800, 2500}},
	SPEC_BATTLE_SCANNER:          {TECH_BATTLE_SCANNER, [6]int{12, 18, 30, 60, 125, 325}, [6]int{5, 7, 12, 25, 50, 125}},
	SPEC_CLOAKING_DEVICE:         {TECH_CLOAKING_DEVICE, [6]int{12, 18, 30, 60, 125, 325}, [6]int{12, 18, 30, 60, 125, 325}},
	SPEC_DAMPER_FIELD:            {TECH_DAMPER_FIELD, [6]int{5, 7, 12, 25, 50, 125}, [6]int{7, 10, 18, 40, 75, 190}},
	SPEC_DISPLACEMENT_DEVICE:     {TECH_DISPLACEMENT_DEVICE, [6]int{10, 15, 25, 50, 100, 250}, [6]int{15, 22, 40, 75, 150, 375}},
	SPEC_ECM_JAMMER:              {TECH_ECM_JAMMER, [6]int{10, 15, 25, 50, 100, 250}, [6]int{5, 7, 12, 25, 50, 125}},
	SPEC_ENERGY_ABSORBER:         {TECH_ENERGY_ABSORBER, [6]int{12, 18, 30, 60, 125, 325}, [6]int{12, 18, 30, 60, 125, 325}},
	SPEC_EXTENDED_FUEL_TANKS:     {TECH_EXTENDED_FUEL_TANKS, [6]int{12, 18, 30, 60, 125, 325}, [6]int{5, 10, 15, 30, 60, 100}},
	SPEC_FAST_MISSILE_RACKS:      {TECH_FAST_MISSILE_RACKS, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_HARD_SHIELDS:            {TECH_HARD_SHIELDS, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_HEAVY_ARMOR:             {TECH_HEAVY_ARMOR, [6]int{10, 15, 25, 50, 100, 250}, [6]int{5, 7, 12, 25, 50, 125}},
	SPEC_HIGH_ENERGY_FOCUS:       {TECH_HIGH_ENERGY_FOCUS, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_HYPERX_CAPACITORS:       {TECH_HYPERX_CAPACITORS, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_INERTIAL_NULLIFIER:      {TECH_INERTIAL_NULLIFIER, [6]int{15, 25, 40, 75, 150, 375}, [6]int{22, 40, 60, 100, 225, 500}},
	SPEC_INERTIAL_STABILIZER:     {TECH_INERTIAL_STABILIZER, [6]int{15, 25, 40, 75, 150, 375}, [6]int{5, 7, 12, 25, 50, 125}},
	SPEC_LIGHTNING_FIELD:         {TECH_LIGHTNING_FIELD, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_MULTIPHASED_SHIELDS:     {TECH_MULTIPHASED_SHIELDS, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_MULTIWAVE_ECM_JAMMER:    {TECH_MULTIWAVE_ECM_JAMMER, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_PHASE_SHIFTER:           {TECH_PHASE_SHIFTER, [6]int{10, 15, 25, 50, 100, 250}, [6]int{15, 22, 37, 75, 150, 375}},
	SPEC_PHASING_CLOAK:           {TECH_PHASING_CLOAK, [6]int{15, 25, 40, 75, 150, 375}, [6]int{22, 40, 60, 100, 225, 500}},
	SPEC_QUANTUM_DETONATOR:       {TECH_QUANTUM_DETONATOR, [6]int{2, 3, 7, 15, 30, 75}, [6]int{3, 5, 10, 22, 45, 100}},
	SPEC_RANGEMASTER_UNIT:        {TECH_RANGEMASTER_UNIT, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_REFLECTION_FIELD:        {TECH_REFLECTION_FIELD, [6]int{10, 15, 25, 50, 100, 250}, [6]int{15, 22, 37, 75, 150, 375}},
	SPEC_REINFORCED_HULL:         {TECH_REINFORCED_HULL, [6]int{10, 15, 25, 50, 100, 250}, [6]int{5, 7, 12, 25, 50, 125}},
	SPEC_SCOUT_LAB:               {TECH_SCOUT_LAB, [6]int{12, 20, 40, 75, 150, 375}, [6]int{7, 12, 20, 37, 75, 180}},
	SPEC_SECURITY_STATIONS:       {TECH_SECURITY_STATIONS, [6]int{5, 7, 12, 25, 50, 125}, [6]int{2, 3, 6, 12, 25, 70}},
	SPEC_SHIELD_CAPACITOR:        {TECH_SHIELD_CAPACITORS, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_STEALTH_FIELD:           {TECH_STEALTH_FIELD, [6]int{10, 15, 25, 50, 100, 250}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_STRUCTURAL_ANALYZER:     {TECH_STRUCTURAL_ANALYZER, [6]int{7, 10, 18, 40, 75, 190}, [6]int{10, 15, 25, 50, 100, 250}},
	SPEC_SUB_SPACE_TELEPORTER:    {TECH_SUBSPACE_TELEPORTER, [6]int{10, 15, 25, 50, 100, 250}, [6]int{15, 22, 37, 75, 150, 375}},
	SPEC_TIME_WARP_FACILITATOR:   {TECH_TIME_WARP_FACILITATOR, [6]int{15, 25, 40, 75, 150, 375}, [6]int{22, 40, 60, 100, 225, 500}},
	SPEC_TRANSPORTERS:            {TECH_TRANSPORTERS, [6]int{10, 15, 25, 50, 100, 250}, [6]int{5, 7, 12, 25, 50, 125}},
	SPEC_TROOP_PODS:              {TECH_TROOP_PODS, [6]int{12, 18, 30, 60, 125, 325}, [6]int{5, 7, 12, 25, 50, 125}},
	SPEC_WARP_DISSIPATOR:         {TECH_WARP_DISSIPATER, [6]int{15, 25, 40, 75, 150, 375}, [6]int{15, 25, 40, 75, 150, 375}},
	SPEC_WIDE_AREA_JAMMER:        {TECH_WIDE_AREA_JAMMER, [6]int{20, 30, 50, 100, 200, 500}, [6]int{20, 30, 50, 100, 200, 500}},
	SPEC_REGENERATION:            {TECH_AUTOMATED_REPAIR_UNIT, [6]int{4, 10, 20, 40, 80, 200}, [6]int{5, 10, 20, 50, 125, 500}},
}

// SpecialDeviceStatsFor 回傳一項特殊裝置的原版數值;查無回 (零值, false)。
func SpecialDeviceStatsFor(dev SpecialDevices) (SpecialDeviceStats, bool) {
	st, ok := specialDeviceTable[dev]
	return st, ok
}

// SpecialDeviceSpace 回傳一項特殊裝置裝在指定艦體等級上佔用的空間。
//
// **負值是有意義的**(戰鬥艙):呼叫端要把它加進「已用空間」的總和,加上去之後總和會變小,
// 那正是原版的算法——不要在這裡取絕對值或截成 0。
func SpecialDeviceSpace(dev SpecialDevices, class CombatShipClass) int {
	st, ok := specialDeviceTable[dev]
	if !ok || class < 0 || int(class) >= len(st.Space) {
		return 0
	}
	return st.Space[class]
}

// SpecialDeviceCost 回傳一項特殊裝置裝在指定艦體等級上的建造成本(BC)。
func SpecialDeviceCost(dev SpecialDevices, class CombatShipClass) int {
	st, ok := specialDeviceTable[dev]
	if !ok || class < 0 || int(class) >= len(st.Cost) {
		return 0
	}
	return st.Cost[class]
}

// DesignSpaceMegafluxersPercent 是巨型通量器對整艘船可用空間的加成(百分點)。
//
// 手冊逐字:「Megafluxers increase the amount of space on each ship by **25%**.」
// 執行檔 `Total_Design_Space_`(0x6E81F)的實作是 `imul eax, 7Dh` 之後 `idiv 100`
// ——**×125/100,整數除法截斷**。手冊給比例、執行檔給截斷方式,兩者合起來才是完整規則。
const DesignSpaceMegafluxersPercent = 125

// DesignHullSpace 回傳一艘船的**可用**空間:艦體基礎空間,巨型通量器再 ×125/100。
//
// 對應 `Total_Design_Space_`(0x6E81F):讀艦體表的空間欄,若帝國已研究巨型通量器
// (`[帝國紀錄+0x170] == 3`,而 0x170 − 0x117 = 89 = TECH_MEGAFLUXERS)就乘 125 再整除 100。
//
// ⚠ 戰鬥艙**不在這個函式裡**——原版把它做成負佔格(見 SpecialDeviceSpace),
// 走的是「已用空間」那一側。這兩個 +25% / +50% 看起來像同一種東西,實作位置卻不同。
func DesignHullSpace(class CombatShipClass, hasMegafluxers bool) int {
	space := ShipHullSpace(class)
	if hasMegafluxers {
		space = space * DesignSpaceMegafluxersPercent / 100
	}
	return space
}
