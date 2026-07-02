package gamedata

// 飛彈防禦(Missile Defenses)/反飛彈火箭(Anti-Missile Rockets, AMR)公式,
// 移植自 MOO2 patch 1.5 官方手冊 moo2_patch1.5/MANUAL_150.html:
//   - "Notes on Missile Defenses"(p123:Weapons / Special Defensive Systems / Missile Evasion)
//   - "Notes on Anti-Missile Rockets"(p125:Rockets & Missiles / Range & Chance to Hit)
// openorion2 未實作這段戰鬥判定(只有 tech 名稱字串),故本檔為手冊到程式碼的首次移植。
// 只搬手冊明確列出、附精確數字的公式/表;沒有精確數字的一律標 TODO,不得自行編造。
//
// 手冊 "Notes on Beam Weapon Mechanics > Beam Defense of Missiles"(p117-120)的
// 飛彈 Beam Defense(FTLlevel/Speed + MissileBonus)於本檔下方實作(MissileSpeed/
// MissileWarheadBonus/MissileBeamDefense)。
//
// ⚠ 手冊此段自相矛盾:明列公式 Speed = BaseSpeed(12) + 2*(FTLlevel-1) + FastBonus(4)
// 得 14/16/18/20/22/24/26(FTL 0-6),但同段附表 Speed 欄為 10/12/14/16/18/20/22。
// 本檔以「明列公式」為準(手冊寫 "calculated as follows"),表格 Speed 欄推測為驅動本身
// 速度(另一量)。此落差需日後對實機行為動態驗證;FTLlevel 對映與 MissileBonus 表無爭議。

// --- Weapons(手冊 p123):哪些武器能攻擊飛彈/魚雷 ---
//
// 飛彈(Missiles)可被以下武器擊落:一般光束武器與 PD 光束武器、攔截機(Interceptors)、
// 反飛彈火箭(Anti-Missile Rockets)、球形武器(如 Pulsar)。
// 魚雷(Torpedoes)不同於飛彈,不能被上述任何武器瞄準或傷害。

// --- Special Defensive Systems(手冊 p123):對飛彈與魚雷皆有效的三種特殊防禦裝置 ---

const (
	// MissileLightningFieldDestroyChance 閃電力場(Lightning Field):
	// 對每一枚試圖命中裝備艦的飛彈、魚雷或戰機,各有 50% 機率直接摧毀。
	// 在 MIRV 飛彈分裂彈頭「之前」判定(匿蹤/位移裝置則在分裂之後判定)。
	MissileLightningFieldDestroyChance = 50

	// MissileCloakingDeviceMissChance 匿蹤裝置(Cloaking Device):
	// 僅在裝置啟動(艦身處於隱形狀態)時,飛彈彈頭與魚雷有 50% 機率未命中。
	MissileCloakingDeviceMissChance = 50

	// MissileDisplacementDeviceMissChance 位移裝置(Displacement Device):
	// 不論其他裝備或情況,飛彈彈頭與魚雷一律有 30% 機率完全未命中。
	MissileDisplacementDeviceMissChance = 30
)

// --- Missile Evasion(手冊 p123):飛彈/魚雷突破防禦、抵達艦體後的命中判定 ---

// MissileDefaultHitChance 飛彈/魚雷突破前述防禦、抵達目標艦後,若目標無任何閃避(evasion)
// 能力,預設 100% 命中。
const MissileDefaultHitChance = 100

// 飛彈閃避加成(Missile Evasion Bonus)的各項來源數值(手冊 p123)。
const (
	// MissileJammerECM ECM Jammer 提供的閃避加成。
	MissileJammerECM = 70
	// MissileJammerMultiWave MultiWave Jammer 提供的閃避加成。
	MissileJammerMultiWave = 100
	// MissileJammerWideAreaSelf Wide Area Jammer 對裝備艦自身的閃避加成。
	MissileJammerWideAreaSelf = 130
	// MissileJammerWideAreaFleet Wide Area Jammer 對艦隊其餘船艦的閃避加成;
	// 手冊註明此艦隊加成不與其他 jammer 疊加(not cumulative with other jammers)。
	MissileJammerWideAreaFleet = 70
	// MissileInertialStabilizer Inertial Stabilizer 提供的閃避加成。
	MissileInertialStabilizer = 25
	// MissileInertialNullifier Inertial Nullifier 提供的閃避加成。
	MissileInertialNullifier = 50
)

// MissileShipDefenseRacialBonus 種族「Ship Defense」特性的三個檔位加成。
// 手冊原文僅列出三個數值(-20 / +25 / +50),未附種族/檔位名稱,故以陣列保留數值本身,
// 檔位名稱留白(TODO 手冊未明列檔位名稱,待查證)。
var MissileShipDefenseRacialBonus = [3]int{-20, 25, 50}

// 艦員經驗等級(Crew experience level)提供的飛彈閃避加成(手冊 p123)。
const (
	MissileCrewGreen      = 0
	MissileCrewRegular    = 7
	MissileCrewVeteran    = 15
	MissileCrewElite      = 25
	MissileCrewUltraElite = 37
)

// MissileHelmsmanEvasionBonus 艦隊統帥(Helmsman)加成貢獻的閃避值,為其 Helmsman 數值的一半
// (手冊:「Half bonus of the Helmsman value」;範例以整數除法 10/2=5 計算)。
func MissileHelmsmanEvasionBonus(helmsmanValue int) int {
	return helmsmanValue / 2
}

// MissileJamChance 計算單一飛彈彈頭被幹擾(jammed)的機率(%)(手冊 p123 Missile Evasion)。
//
// 手冊原文:「The probability of a single missile warhead getting jammed is equal to the
// missile evasion bonus of the defender minus the best known scanner bonus of the attacker
// and this probability is halved if the missile has ECCM (Electronic Counter Countermeasures)」。
// 幹擾機率逐彈頭獨立判定,故 MIRV 飛彈可能只有部分彈頭被擋下。
//
// 手冊範例(逐項核對用):Wide Area Jammer 艦隊加成(+70)+ Stabilizer(+25)+ 種族懲罰(-20)
// + 一般艦員(+7)+ 統帥加成一半(10/2=5)= 87;攻擊方 Tachyon Scanner 已知加成 20;
// 飛彈具 ECCM 故再減半:P = [(87)-20]/2 = 33%。
func MissileJamChance(defenderEvasionBonus, attackerScannerBonus int, hasECCM bool) int {
	chance := defenderEvasionBonus - attackerScannerBonus
	if hasECCM {
		// 手冊範例 67/2=33.5 → 33%,採無條件捨去(整數除法對非負值等同 floor;
		// chance 為負值時 Go 的截斷方向與 floor 不同,惟手冊未給負值案例,此處不特別處理)。
		chance /= 2
	}
	return chance
}

// --- Notes on Anti-Missile Rockets(手冊 p125) ---

// MissileAMRMaxRangeSquares 反飛彈火箭(AMR)最大射程(格):手冊「maximum range of 15 squares」。
const MissileAMRMaxRangeSquares = 15

// MissileAMRRangeIndex 將格距離(squares)換算為 AMR 命中率公式所需的 Range 索引(0-6)。
//
// 手冊:AMR 使用的格→Range 換算比標準換算「多位移一格」,把 sq=0 視同標準換算的 range1
// (原文:"effectively shifting the Squares-to-Range conversion by one, thus treating range0
// as range1")。手冊核對表(AMR 格→Range):0-2→1、3-5→2、6-8→3、9-11→4、12-14→5、15-17→6。
// 换算式等價於 ceil((sq+1)/3);Range=0 僅在飛彈座標與開火艦艇中心完全重合的理論情況下成立
// (手冊註腳,見 MissileAMRChanceToHit(0)=65%),一般格距離公式到不了 Range=0。
// squareDistance 超過 MissileAMRMaxRangeSquares(15)時已超出 AMR 射程,呼叫端應先行判斷。
func MissileAMRRangeIndex(squareDistance int) int {
	return (squareDistance + 3) / 3 // ceil((squareDistance+1)/3)
}

// MissileAMRChanceToHit AMR 對飛彈的命中率(%),只與 Range 索引有關,與目標飛彈種類/血量/mods 無關
// (手冊:「Type, hit points or mods of the target missile do not matter for Chance-to-Hit
// calculation, only distance is relevant」)。命中一次只摧毀彈頭堆疊中的一枚飛彈。
//
// 手冊逐字公式:「AMR Chance-to-Hit = 70 - rounddown((Range + 2) * 10 / 3) - 1」。
// 但逐字代入 Range=0..6 得到 63/59/56/53/49/46/43,與手冊自己列出的核對表
// 65/61/58/55/51/48/45(Range 0-6)不符,每一項都少 2。反推:手冊表格才是可信的最終結果
// (該表逐列列出到小數點,顯然是作者自行算好貼上),故公式中的「-1」應是在 rounddown() 內
// 一起捨去,而非捨去後才減 1,即實際應為:70 - (rounddown((Range+2)*10/3) - 1)
// = 71 - rounddown((Range+2)*10/3)。已用手冊核對表逐列驗證,見 missile_test.go。
func MissileAMRChanceToHit(rangeIndex int) int {
	return 71 - (rangeIndex+2)*10/3
}

// --- Beam Defense of Missiles(手冊 p117-120) ---
//
// Missile Beam Defense = 5*Speed + MissileBonus。Speed 由 FTLlevel 決定(見上方矛盾註解)。

// 驅動科技對應的 FTLlevel(手冊 Drive/FTLlevel 表,無爭議)。
const (
	MissileFTLNone        = 0 // None
	MissileFTLNuclear     = 1 // Nuclear Drive
	MissileFTLFusion      = 2 // Fusion Drive
	MissileFTLIon         = 3 // Ion Drive
	MissileFTLAntiMatter  = 4 // Anti-Matter Drive
	MissileFTLHyper       = 5 // Hyper Drive
	MissileFTLInterphased = 6 // Interphased Drive
)

// 各彈頭型別的 MissileBonus(手冊 Missile/MissileBonus 表,無爭議)。
const (
	MissileWarheadNuclear   = -10 // Nuclear Missile
	MissileWarheadMerculite = 15  // Merculite Missile
	MissileWarheadPulson    = 40  // Pulson Missile
	MissileWarheadZeon      = 70  // Zeon Missile
)

// MissileSpeed 依 FTLlevel 回傳飛彈速度,採手冊明列公式:
//
//	Speed = BaseSpeed(12) + 2*(FTLlevel-1) + FastBonus(4)
//
// (與手冊附表 Speed 欄有 +4 落差,見檔頭矛盾註解;此處以明列公式為準。)
func MissileSpeed(ftlLevel int) int {
	return 12 + 2*(ftlLevel-1) + 4
}

// MissileBeamDefense 回傳飛彈的 Beam Defense = 5*Speed + warheadBonus。
// warheadBonus 用上方 MissileWarhead* 常數。
func MissileBeamDefense(ftlLevel, warheadBonus int) int {
	return 5*MissileSpeed(ftlLevel) + warheadBonus
}
