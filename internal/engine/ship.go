package engine

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

// ship.go 是 save.ShipDesign → 戰鬥引擎屬性(BeamAttacker/BeamTarget/ShipCombatState)的橋接,
// 對照 adapter.go 的 ColonyStateFromSave/PlayerStateFromSave 模式。只用 gamedata 現成公式
// (ComputerBonus/BeamOffense/BeamDefense/CombatSpeed/ShipCrewOffenseBonus/ShipCrewDefenseBonus/
// MaxComputerHP),不重寫規則。

// Specials 位元定義,對照 openorion2/src/gamestate.h:685-724 enum SpecialDevices。
// bit 索引 = enum 值本身(enum 從 1 起算,bit 0 未使用)。位元檢測邏輯對照
// openorion2/src/utils.cpp:402-404 checkBitfield:bitfield[bit/8] & (1<<(bit%8))(LSB-first)。
const (
	SpecAchillesTargetingUnit = 1
	SpecAugmentedEngines      = 2
	SpecAutomatedRepairUnit   = 3
	SpecBattlePods            = 4
	SpecBattleScanner         = 5
	SpecCloakingDevice        = 6
	SpecDamperField           = 7
	SpecDisplacementDevice    = 8
	SpecEcmJammer             = 9
	SpecEnergyAbsorber        = 10
	SpecExtendedFuelTanks     = 11
	SpecFastMissileRacks      = 12
	SpecHardShields           = 13
	SpecHeavyArmor            = 14
	SpecHighEnergyFocus       = 15
	SpecHyperXCapacitors      = 16
	SpecInertialNullifier     = 17
	SpecInertialStabilizer    = 18
	SpecLightningField        = 19
	SpecMultiphasedShields    = 20
	SpecMultiwaveEcmJammer    = 21
	SpecPhaseShifter          = 22
	SpecPhasingCloak          = 23
	SpecQuantumDetonator      = 24
	SpecRangemasterUnit       = 25
	SpecReflectionField       = 26
	SpecReinforcedHull        = 27
	SpecScoutLab              = 28
	SpecSecurityStations      = 29
	SpecShieldCapacitor       = 30
	SpecStealthField          = 31
	SpecStructuralAnalyzer    = 32
	SpecSubSpaceTeleporter    = 33
	SpecTimeWarpFacilitator   = 34
	SpecTransporters          = 35
	SpecTroopPods             = 36
	SpecWarpDissipator        = 37
	SpecWideAreaJammer        = 38
	SpecRegeneration          = 39
)

// checkSpecBit 對照 openorion2 checkBitfield(openorion2/src/utils.cpp:402-404)。
// bitfield 傳 nil 時一律回 false(對照原碼 `bitfield && (...)`:空指標視為未設)。
func checkSpecBit(bitfield []uint8, bit int) bool {
	if bitfield == nil {
		return false
	}
	idx := bit / 8
	if idx < 0 || idx >= len(bitfield) {
		return false
	}
	return bitfield[idx]&(1<<uint(bit%8)) != 0
}

// ShipHasSpecial 對照 ShipDesign::hasSpecial(openorion2/src/gamestate.cpp:827-833)。
func ShipHasSpecial(d *save.ShipDesign, id int) bool {
	return checkSpecBit(d.Specials[:], id)
}

// ShipHasWorkingSpecial 對照 ShipDesign::hasWorkingSpecial(openorion2/src/gamestate.cpp:835-839):
// 裝了該特殊系統,且未被打壞。specDamage 對照 save.Ship.DamagedSpecials;傳 nil 代表不追蹤損傷
// (例如尚未建造、只是設計藍圖時),視為全部正常運作。
func ShipHasWorkingSpecial(d *save.ShipDesign, id int, specDamage []uint8) bool {
	return ShipHasSpecial(d, id) && !checkSpecBit(specDamage, id)
}

// ShipBeamAttackFromDesign 推導光束命中加成(BA),對照
// ShipDesign::beamOffense(openorion2/src/gamestate.cpp:903-916)+ Ship::beamOffense
// (openorion2/src/gamestate.cpp:1688-1698)+ getBeamAttack(openorion2/src/gamestate.cpp:2360-2368,
// 疊加 traits[TRAIT_SHIP_ATTACK])。
//
//   - compDamage: 電腦 HP 損傷值(來自 save.Ship.ComputerDamage)。ShipDesign 本身不含此欄位
//     (那是艦體的可變戰損狀態,不是設計藍圖的一部分),故由呼叫端傳入;傳 0 表示電腦未受損。
//   - specDamage: 已損毀特殊裝置點陣圖(來自 save.Ship.DamagedSpecials);設計藍圖階段傳 nil。
//   - crewLevel: 艦員等級(0 新兵..4 超級精銳),見 gamedata.ShipCrewOffenseBonus。
//   - raceShipAttack: 種族/科技加成(openorion2 為 Player.traits[TRAIT_SHIP_ATTACK],屬玩家層級
//     資料,ShipDesign 不含,由呼叫端傳入;無此加成傳 0)。
func ShipBeamAttackFromDesign(d *save.ShipDesign, crewLevel int, compDamage int, specDamage []uint8, raceShipAttack int) int {
	computerWorking := compDamage < gamedata.MaxComputerHP(int(d.Size))
	battleScanner := ShipHasWorkingSpecial(d, SpecBattleScanner, specDamage)
	ret := gamedata.BeamOffense(int(d.Computer), computerWorking, battleScanner)
	ret += gamedata.ShipCrewOffenseBonus(crewLevel)
	ret += raceShipAttack
	return ret
}

// ShipBeamDefenseFromDesign 推導光束閃避加成(BD),對照
// ShipDesign::beamDefense(openorion2/src/gamestate.cpp:918-932)+ Ship::beamDefense
// (openorion2/src/gamestate.cpp:1700-1710)+ getBeamDefense(openorion2/src/gamestate.cpp:2388-2398,
// transDimensional 來自 owner->traits[TRAIT_TRANS_DIMENSIONAL])。
//
//   - driveDamage: 引擎損傷百分比(0-100,來自 save.Ship.DriveDamage);設計藍圖階段傳 0。
//   - specDamage: 同 ShipBeamAttackFromDesign。
//   - crewLevel: 艦員等級,見 gamedata.ShipCrewDefenseBonus。
//   - transDimensional: 種族科技(Player.traits[TRAIT_TRANS_DIMENSIONAL]),屬玩家層級資料,
//     ShipDesign 不含,由呼叫端傳入。
func ShipBeamDefenseFromDesign(d *save.ShipDesign, crewLevel int, driveDamage int, specDamage []uint8, transDimensional bool) int {
	reinforcedHull := ShipHasWorkingSpecial(d, SpecReinforcedHull, specDamage)
	augmentedEngines := ShipHasWorkingSpecial(d, SpecAugmentedEngines, specDamage)
	speed := gamedata.CombatSpeed(int(d.BaseCombatSpeed), int(d.Size), driveDamage, augmentedEngines, reinforcedHull, transDimensional)

	inertialNullifier := ShipHasWorkingSpecial(d, SpecInertialNullifier, specDamage)
	inertialStabilizer := ShipHasWorkingSpecial(d, SpecInertialStabilizer, specDamage)
	ret := gamedata.BeamDefense(speed, inertialNullifier, inertialStabilizer)
	ret += gamedata.ShipCrewDefenseBonus(crewLevel)
	return ret
}

// ShipBeamAttackWithOfficer 在 ShipBeamAttackFromDesign 之上疊加艦艇軍官的 Weaponry 技能加成,
// 對照 GameState::shipBeamOffense(openorion2/src/gamestate.cpp:2365-2377):
//
//	ret := sptr->beamOffense(ignoreDamage); ... if (sptr->officer >= 0) ret += officer.skillBonus(SKILL_WEAPONRY);
//
// officerWeaponryBonus 由呼叫端以
// gamedata.LeaderSkillBonus(int(gamedata.SKILL_WEAPONRY), tier, expLevel) 算好傳入
// (該艦未指派軍官,或軍官沒有 Weaponry 技能,傳 0 即可,行為與原碼「sptr->officer < 0」等價)。
// 只加一個新參數而不改既有函式簽章:ShipBeamAttackFromDesign 已有測試鎖住既有行為,不動它。
func ShipBeamAttackWithOfficer(d *save.ShipDesign, crewLevel int, compDamage int, specDamage []uint8, raceShipAttack int, officerWeaponryBonus int) int {
	return ShipBeamAttackFromDesign(d, crewLevel, compDamage, specDamage, raceShipAttack) + officerWeaponryBonus
}

// ShipBeamDefenseWithOfficer 在 ShipBeamDefenseFromDesign 之上疊加艦艇軍官的 Helmsman 技能加成,
// 對照 GameState::shipBeamDefense(openorion2/src/gamestate.cpp:2387-2405):
//
//	if (sptr->officer >= 0) ret += officer.skillBonus(SKILL_HELMSMAN);
//
// officerHelmsmanBonus 由呼叫端以
// gamedata.LeaderSkillBonus(int(gamedata.SKILL_HELMSMAN), tier, expLevel) 算好傳入。
func ShipBeamDefenseWithOfficer(d *save.ShipDesign, crewLevel int, driveDamage int, specDamage []uint8, transDimensional bool, officerHelmsmanBonus int) int {
	return ShipBeamDefenseFromDesign(d, crewLevel, driveDamage, specDamage, transDimensional) + officerHelmsmanBonus
}

// ShipCombatStateFromDesign 組出 ShipCombatState。BeamDefense 由 ShipBeamDefenseFromDesign 推導,
// 其餘欄位(ArmorHP/StructureHP/ShieldReduction/HardShield)需要武器/裝甲/護盾數值表——那些查表
// gamedata 尚未移植(openorion2 armor/shield 型別對應的 HP/減傷數值),為避免臆造數字,
// 一律由呼叫端(戰鬥流程或測試)以參數顯式傳入。
//
// TODO 需 armor HP/shield 數值表(未移植):待 gamedata 補上 d.Armor/d.Shield → HP/ShieldReduction/
// HardShield 的查表公式後,可以改成從 d 直接推導,現在保留參數化介面。
func ShipCombatStateFromDesign(
	d *save.ShipDesign,
	crewLevel int,
	driveDamage int,
	specDamage []uint8,
	transDimensional bool,
	armorHP int,
	structureHP int,
	shieldReduction int,
	hardShield bool,
) ShipCombatState {
	return ShipCombatState{
		StructureHP:     structureHP,
		ArmorHP:         armorHP,
		ShieldReduction: shieldReduction,
		HardShield:      hardShield,
		BeamDefense:     ShipBeamDefenseFromDesign(d, crewLevel, driveDamage, specDamage, transDimensional),
	}
}
