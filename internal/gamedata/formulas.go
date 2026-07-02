package gamedata

// 唯讀衍生公式與查表,逐一移植 openorion2 gamestate.cpp(GPL v2)。
// 這些「規則常數表 + 公式」是 gameplay 重建的參考基準;special 裝備以 bool 傳入,
// 與存檔 bitfield 解耦(save 層負責解位元)。查表數值見 docs/tech/formulas.md。

// 各戰鬥艦級(size 0-5)的查表。MAX_COMBAT_SHIP_CLASSES = 6。
var (
	computerHPTable    = [6]int{1, 2, 5, 7, 10, 20}
	driveHPTable       = [6]int{2, 5, 10, 15, 20, 40}
	computerBonusTable = [6]int{0, 25, 50, 75, 100, 125}
	// 礦產豐度(PlanetMinerals 0-4)對應的基礎工業產出。
	mineralProductionTable = [5]int{1, 2, 3, 5, 8}
)

// PlanetBaseProduction 回傳礦產豐度對應的基礎產出(Planet::baseProduction)。
func PlanetBaseProduction(minerals int) int {
	if minerals < 0 || minerals >= len(mineralProductionTable) {
		return 0
	}
	return mineralProductionTable[minerals]
}

// MaxComputerHP 依艦級回傳電腦 HP 上限(ShipDesign::maxComputerHP)。
func MaxComputerHP(size int) int {
	if size < 0 || size >= len(computerHPTable) {
		return 0
	}
	return computerHPTable[size]
}

// ComputerHP 扣除損傷後的電腦 HP(ShipDesign::computerHP)。
func ComputerHP(size, compDamage int) int {
	m := MaxComputerHP(size)
	if compDamage < m {
		return m - compDamage
	}
	return 0
}

// MaxDriveHP 引擎 HP 上限;強化船殼(SPEC_REINFORCED_HULL)×3(ShipDesign::maxDriveHP)。
func MaxDriveHP(size int, reinforcedHull bool) int {
	if size < 0 || size >= len(driveHPTable) {
		return 0
	}
	ret := driveHPTable[size]
	if reinforcedHull {
		ret *= 3
	}
	return ret
}

// DriveHP 依損傷百分比(0-100)計算引擎 HP(ShipDesign::driveHP)。
func DriveHP(size, driveDamage int, reinforcedHull bool) int {
	if driveDamage >= 100 {
		return 0
	}
	return MaxDriveHP(size, reinforcedHull) * (100 - driveDamage) / 100
}

// CombatSpeed 戰鬥移動力(ShipDesign::combatSpeed)。
// 引擎損傷 >33%(以 HP 換算,非直接百分比)會使戰鬥中失去動力。
func CombatSpeed(baseCombatSpeed, size, driveDamage int, augmentedEngines, reinforcedHull, transDimensional bool) int {
	ret := baseCombatSpeed
	if augmentedEngines {
		ret += 5
	}
	maxHP := MaxDriveHP(size, reinforcedHull)
	hp := DriveHP(size, driveDamage, reinforcedHull)
	minHP := 2 * maxHP / 3
	if minHP < hp {
		hp -= minHP
		maxHP -= minHP
		if maxHP > 0 {
			ret = ret * hp / maxHP
		}
	} else {
		ret = 0
	}
	if transDimensional {
		ret += 4
	}
	return ret
}

// ComputerBonus 電腦型別(0-5)提供的命中加成。
func ComputerBonus(computerType int) int {
	if computerType < 0 || computerType >= len(computerBonusTable) {
		return 0
	}
	return computerBonusTable[computerType]
}

// BeamOffense 光束武器命中加成(ShipDesign::beamOffense)。
// computerWorking = 電腦未被完全打壞(compDamage < maxComputerHP)。
func BeamOffense(computerType int, computerWorking, battleScanner bool) int {
	ret := 0
	if computerWorking {
		ret += ComputerBonus(computerType)
	}
	if battleScanner {
		ret += 50
	}
	return ret
}

// BeamDefense 光束閃避(ShipDesign::beamDefense)。combatSpeed 由 CombatSpeed 算得後傳入。
func BeamDefense(combatSpeed int, inertialNullifier, inertialStabilizer bool) int {
	ret := combatSpeed * 5
	if inertialNullifier {
		ret += 100
	}
	if inertialStabilizer {
		ret += 50
	}
	return ret
}

// LeaderHireCost 軍官雇用費(Leader::hireCost):10*skillValue*(expLevel+1)+modifier,下限 0。
func LeaderHireCost(skillValue, expLevel, modifier int) int {
	ret := 10*skillValue*(expLevel+1) + modifier
	if ret < 0 {
		return 0
	}
	return ret
}
