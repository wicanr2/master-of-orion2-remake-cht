package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// ensureAIShipDesigns 維持原版每位 AI 六筆設計庫的資料形狀。初始角色由 IDA 證實為
// hull 0..5 對 role 0..5；既有設計不會在一般正規化時被覆寫。
func (s *GameSession) ensureAIShipDesigns(i int) {
	if i < 0 || i >= len(s.AIPlayers) {
		return
	}
	a := &s.AIPlayers[i]
	if len(a.ShipDesigns) > PlayerShipDesignCount {
		a.ShipDesigns = a.ShipDesigns[:PlayerShipDesignCount]
	}
	for len(a.ShipDesigns) < PlayerShipDesignCount {
		hull := len(a.ShipDesigns)
		role := AutoDesignRole(hull)
		loadout, ok := s.autoDesignShipFor(a.Player, playerShipDesignClasses[hull], role)
		if !ok {
			loadout = AutoDesignLoadout{RawRole: role}
		}
		a.ShipDesigns = append(a.ShipDesigns, blueprintFromLoadout(playerShipDesignClasses[hull], loadout))
	}
}

// updateAIShipDesignsAfterTech 對映 Update_Player_Ship_Designs_ @ 0x57112 的 AI 分支：
// 只重建 hull 0..4，hull 5 與既有實艦保持不變。
func (s *GameSession) updateAIShipDesignsAfterTech(i int) {
	s.ensureAIShipDesigns(i)
	a := &s.AIPlayers[i]
	for hull := 0; hull <= 4; hull++ {
		role := AutoDesignRole(hull)
		if loadout, ok := s.autoDesignShipFor(a.Player, playerShipDesignClasses[hull], role); ok {
			a.ShipDesigns[hull] = blueprintFromLoadout(playerShipDesignClasses[hull], loadout)
		}
	}
}

func shipFromBlueprint(name string, design ShipBlueprint, profileOptions []Component, computerRaw, driveRaw int) Ship {
	w := pick(profileOptions, design.Weapon)
	armor := pick(ArmorOptions, design.Armor)
	shield := pick(ShieldOptions, design.Shield)
	special := pick(SpecialOptions, design.Special)
	ship := Ship{
		Name: name, Class: design.Class, Weapon: w.Name, Armor: armor.Name, Shield: shield.Name,
		Special: special.Name, WeaponAttack: w.Value, Mods: append([]string(nil), design.Mods...),
		Arc: design.Arc, WeaponAmmo: design.Ammo, WeaponMounts: cloneWeaponMounts(design.WeaponMounts),
		SpecialIDs: append([]int(nil), design.SpecialIDs...), Specials: cloneSpecialMounts(design.Specials),
	}
	if class, ok := shipClassFromName(design.Class); ok {
		ship.DesignSizeRaw, ship.DesignSizeRawKnown = uint8(class), true
		ship.BaseCombatSpeedRaw = uint8(gamedata.ShipCombatSpeed(driveRaw, class, false, false))
		ship.BaseCombatSpeedKnown = true
	}
	if computerRaw >= 0 && computerRaw <= 5 {
		ship.ComputerRaw, ship.ComputerRawKnown = uint8(computerRaw), true
	}
	if design.Armor >= 0 && design.Armor < len(ArmorOptions) {
		ship.ArmorRaw, ship.ArmorRawKnown = uint8(design.Armor), true
	}
	if design.Shield >= 0 && design.Shield < len(ShieldOptions) {
		ship.ShieldRaw, ship.ShieldRawKnown = uint8(design.Shield), true
	}
	// remake 新造艦從未受損且由新兵起步；這些合法零值必須標為已知，才能進入
	// 原版 sub_5EF4B 國力 producer，而不被誤判成舊 JSON 缺欄。
	ship.CrewLevelRawKnown = true
	ship.OriginalDamageKnown = true
	return ship
}

func (s *GameSession) syncAIShipStrength(i int) {
	if i < 0 || i >= len(s.AIPlayers) {
		return
	}
	total := 0
	for _, sh := range s.AIPlayers[i].Ships {
		if !isSupportShipClass(sh.Class) {
			total += shipStrength(sh.Class)
		}
	}
	s.AIPlayers[i].FleetStrength = total
}

func (s *GameSession) syncAICommandPoints(i int) {
	if i < 0 || i >= len(s.AIPlayers) {
		return
	}
	a := &s.AIPlayers[i]
	supply := gamedata.CommandPointsBase
	for _, buildings := range a.ColonyBuildings {
		supply += gamedata.CommandPointsFromBuildings(buildings)
	}
	used := 0
	for _, sh := range a.Ships {
		if class, ok := shipClassFromName(sh.Class); ok {
			used += gamedata.ShipCommandCost(class)
		}
	}
	a.Player.CommandPointsSupply = supply
	a.Player.UsedCommandPoints = used
}

// reduceAIShipStrength 將仍使用戰力比例的高階 raid／AI-vs-AI 近似結果落回實艦。
// 有實艦時以較弱艦優先移除，直到不超過目標；舊測試／舊狀態沒有實艦時保留純量相容。
func (s *GameSession) reduceAIShipStrength(i, target int) {
	if i < 0 || i >= len(s.AIPlayers) {
		return
	}
	a := &s.AIPlayers[i]
	if target < 0 {
		target = 0
	}
	if len(a.Ships) == 0 {
		a.FleetStrength = target
		return
	}
	for len(a.Ships) > 0 {
		s.syncAIShipStrength(i)
		if a.FleetStrength <= target {
			return
		}
		weakest := 0
		for j := 1; j < len(a.Ships); j++ {
			if shipStrength(a.Ships[j].Class) < shipStrength(a.Ships[weakest].Class) {
				weakest = j
			}
		}
		a.Ships = append(a.Ships[:weakest], a.Ships[weakest+1:]...)
	}
	s.syncAIShipStrength(i)
}

// advanceAIShipProduction 只保留給舊存檔相容與窄單元測試；正常 AI 回合已改走逐殖民地
// typed ship slot。這裡的「可負擔最高 hull 0..4」不是原版精確生產評分。
func (s *GameSession) advanceAIShipProduction(i, production int) {
	if i < 0 || i >= len(s.AIPlayers) || production <= 0 {
		return
	}
	s.ensureAIShipDesigns(i)
	a := &s.AIPlayers[i]
	a.ShipBuildProgress += production
	view := *s
	view.Player = a.Player
	for {
		chosen, cost := -1, 0
		for hull := 4; hull >= 0; hull-- {
			c, known := view.BlueprintDesignCost(a.ShipDesigns[hull])
			if known && view.BlueprintDesignFits(a.ShipDesigns[hull]) && c <= a.ShipBuildProgress {
				chosen, cost = hull, c
				break
			}
		}
		if chosen < 0 || cost <= 0 {
			break
		}
		a.ShipBuildProgress -= cost
		name := fmt.Sprintf("%s %s %d", a.Name, playerShipDesignClasses[chosen], len(a.Ships)+1)
		a.Ships = append(a.Ships, shipFromBlueprint(name, a.ShipDesigns[chosen], BuildWeaponOptions(s.RuleProfile),
			originalBestComputer(a.Player), aiDriveLevel(*a)))
	}
	if len(a.Ships) > 0 {
		s.syncAIShipStrength(i)
	}
}

func (s *GameSession) aiIndexByName(name string) int {
	for i := range s.AIPlayers {
		if s.AIPlayers[i].Name == name || stripAILabel(s.AIPlayers[i].Name) == name {
			return i
		}
	}
	return -1
}

// normalizeAIShipState 在讀取舊存檔時把只有 FleetStrength 的摘要決定性轉成 hull 0 實艦。
// 這是格式相容層，不是原版開局艦隊 oracle；新格式已有 Ships 時只重算衍生摘要。
func (s *GameSession) normalizeAIShipState(i int) {
	if i < 0 || i >= len(s.AIPlayers) {
		return
	}
	s.ensureAIShipDesigns(i)
	a := &s.AIPlayers[i]
	if len(a.Ships) == 0 && a.FleetStrength > 0 {
		unit := shipStrength(playerShipDesignClasses[0])
		for remaining := a.FleetStrength; remaining >= unit; remaining -= unit {
			name := fmt.Sprintf("%s %s %d", a.Name, playerShipDesignClasses[0], len(a.Ships)+1)
			a.Ships = append(a.Ships, shipFromBlueprint(name, a.ShipDesigns[0], BuildWeaponOptions(s.RuleProfile),
				originalBestComputer(a.Player), aiDriveLevel(*a)))
		}
	}
	s.syncAIShipStrength(i)
}

func aiRaceCombatBonuses(a AIOpponent) (attackPct, defensePct int) {
	if a.RaceIndex >= 0 && a.RaceIndex < len(Races) {
		return Races[a.RaceIndex].CombatPct, Races[a.RaceIndex].ShipDefPct
	}
	return 0, 0
}

func aiDriveLevel(a AIOpponent) int {
	tier := gamedata.DriveTierFromTechs(func(topic gamedata.ResearchTopic, tech gamedata.Technology) bool {
		return driveTechOwned(a.Player, topic, tech)
	})
	if tier < 1 {
		tier = 1
	}
	return tier
}

func (s *GameSession) aiShipCombatant(a AIOpponent, sh Ship, shipIdx int) combatant {
	body := shipStrength(sh.Class)
	attackPct, defensePct := aiRaceCombatBonuses(a)
	atk := body + sh.WeaponAttack
	atk += atk * attackPct / 100
	atk += shipBeamOffenseBonus(sh)
	hp := body * 3 * shipStructureMultiplier(sh) / 100
	if d := ShipDamage(sh); d > 0 {
		hp = maxInt(ShipDamageFloorHP, hp-d)
	}
	drive := aiDriveLevel(a)
	commando := fleetOfficerSkillMax(a.Leaders, a.Ships, gamedata.SKILL_COMMANDO)
	security := fleetOfficerSkillMax(a.Leaders, a.Ships, gamedata.SKILL_SECURITY)
	return combatant{
		hp: hp, maxHP: shipMaxHP(sh), atk: atk, def: body + body*defensePct/100 + shipBeamDefenseBonus(sh),
		wmin: atk / 2, wmax: atk, shield: s.nebulaShield(shieldReduceByName(sh.Shield), shipHasHardShield(sh)) * shipShieldMultiplier(sh) / 100,
		armor: effectiveArmorHP(sh), kind: weaponKindByName(sh.Weapon), weaponName: sh.Weapon,
		mods: append([]string(nil), sh.Mods...), weaponMounts: cloneWeaponMounts(sh.WeaponMounts),
		sizeClass: shipSizeClass(sh.Class), hasAMR: shipHasSpecial(sh, antiMissileRocketName),
		hasHEF: shipHasSpecial(sh, highEnergyFocusName), beamSystems: shipBeamAttackerSystems(sh),
		initiative: gamedata.CombatInitiative(atk, gamedata.ShipCombatSpeed(drive, shipSizeClass(sh.Class), false, false)),
		shots: gamedata.ShotsThisRound(shipShotsKind(sh), weaponKindByName(sh.Weapon) == WeaponKindBeam,
			weaponKindByName(sh.Weapon) == WeaponKindMissile, true),
		ammo: NormalizeWeaponAmmo(sh.Weapon, sh.WeaponAmmo), ammoSet: true,
		apNegated: shipNegatesArmorPiercing(sh), hasLightningField: shipHasLightningField(sh),
		hasDisplacement: shipHasDisplacementDevice(sh), hardShield: shipHasHardShield(sh),
		scannerJamReduction: bestPlayerScannerJamReduction(a.Player), missileEvasion: shipMissileEvasionBonus(sh),
		missileFTLLevel: drive, marines: ShipMarineComplement(sh), securityStations: shipHasSecurityStations(sh),
		marineStrength:   aiMarineForce(a),
		marineHitsToKill: gamedata.GroundMarineHitsToKill(aiRaceHasTrait(a, gamedata.TRAIT_HIGH_G), hasPoweredArmorFor(a.Player)),
		boardingBonus:    gamedata.ShipCrewBoardingBonus(gamedata.CrewLevelForXP(sh.CrewXP, aiRaceHasTrait(a, gamedata.TRAIT_WARLORD))),
		commandoBonus:    commando, securityBonus: security,
		assaultShuttles: shipHasAssaultShuttles(sh), cloakKind: shipCloakKind(sh),
		cloaked: shipCloakKind(sh) != CloakNone, energyAbsorber: shipHasSpecial(sh, energyAbsorberName),
		autoRepair: shipHasAutoRepair(sh), shipIdx: shipIdx,
	}
}

func (s *GameSession) aiCombatants(i int) ([]combatant, []int) {
	if i < 0 || i >= len(s.AIPlayers) {
		return nil, nil
	}
	a := s.AIPlayers[i]
	out, strengths := make([]combatant, 0, len(a.Ships)), make([]int, 0, len(a.Ships))
	for shipIdx, sh := range a.Ships {
		if isSupportShipClass(sh.Class) {
			continue
		}
		out = append(out, s.aiShipCombatant(a, sh, shipIdx))
		strengths = append(strengths, shipStrength(sh.Class))
	}
	return out, strengths
}

func (s *GameSession) applyAICombatantSurvivors(i int, survivors []combatant) {
	if i < 0 || i >= len(s.AIPlayers) {
		return
	}
	alive := make(map[int]combatant, len(survivors))
	for _, c := range survivors {
		if c.shipIdx >= 0 {
			alive[c.shipIdx] = c
		}
	}
	a := &s.AIPlayers[i]
	kept := make([]Ship, 0, len(alive))
	for idx, sh := range a.Ships {
		if c, ok := alive[idx]; ok {
			sh.Damage = maxInt(0, shipMaxHP(sh)-c.hp)
			kept = append(kept, sh)
		}
	}
	a.Ships = kept
	s.syncAIShipStrength(i)
}

func (s *GameSession) aiTacticalShips(i int) []CombatShip {
	if i < 0 || i >= len(s.AIPlayers) {
		return nil
	}
	a := s.AIPlayers[i]
	attackPct, defensePct := aiRaceCombatBonuses(a)
	drive := aiDriveLevel(a)
	commando := fleetOfficerSkillMax(a.Leaders, a.Ships, gamedata.SKILL_COMMANDO)
	security := fleetOfficerSkillMax(a.Leaders, a.Ships, gamedata.SKILL_SECURITY)
	marineStrength := aiMarineForce(a)
	marineHits := gamedata.GroundMarineHitsToKill(aiRaceHasTrait(a, gamedata.TRAIT_HIGH_G), hasPoweredArmorFor(a.Player))
	color := 1
	if a.ColorKnown {
		color = normalizeCMBTSHPColor(a.Color, 1)
	}
	out := make([]CombatShip, 0, len(a.Ships))
	for i, sh := range a.Ships {
		if isSupportShipClass(sh.Class) {
			continue
		}
		body := shipStrength(sh.Class)
		atk := body + sh.WeaponAttack
		atk += atk * attackPct / 100
		atk += shipBeamOffenseBonus(sh)
		hp := body * 3 * shipStructureMultiplier(sh) / 100
		if d := ShipDamage(sh); d > 0 {
			hp = maxInt(ShipDamageFloorHP, hp-d)
		}
		bays := fighterBaysForShip(sh)
		bayKind := FighterInterceptor
		if len(bays) > 0 {
			bayKind = bays[0]
		}
		cs := CombatShip{
			Name: sh.Name, HP: hp, MaxHP: shipMaxHP(sh), Attack: atk, Col: 6, Row: i % TacticalGridRows, Facing: 8,
			Defense: body + body*defensePct/100 + shipBeamDefenseBonus(sh), WeaponMin: atk / 2, WeaponMax: atk,
			ShieldReduction: s.nebulaShield(shieldReduceByName(sh.Shield), shipHasHardShield(sh)) * shipShieldMultiplier(sh) / 100,
			HardShield:      shipHasHardShield(sh), ArmorHP: effectiveArmorHP(sh), Kind: weaponKindByName(sh.Weapon),
			WeaponName: sh.Weapon, Mods: append([]string(nil), sh.Mods...), WeaponArc: NormalizeWeaponArc(sh.Weapon, sh.Arc),
			WeaponMounts: cloneWeaponMounts(sh.WeaponMounts), WeaponModes: NewTacticalWeaponModes(sh.WeaponMounts),
			HEF: shipHasSpecial(sh, highEnergyFocusName), APNegated: shipNegatesArmorPiercing(sh),
			MissileEvasion: shipMissileEvasionBonus(sh), HasAMR: shipHasSpecial(sh, antiMissileRocketName),
			HasLightningField: shipHasLightningField(sh), HasDisplacement: shipHasDisplacementDevice(sh),
			BeamSystems: shipBeamAttackerSystems(sh), DriveLevel: drive, ArmorLevelAboveTitanium: armorLevelAboveTitanium(sh.Armor),
			CombatSpeed: gamedata.ShipCombatSpeed(drive, shipSizeClass(sh.Class), false, false), SizeClass: shipSizeClass(sh.Class),
			FighterRacialDefenseBonus: defensePct, ScannerJamReduction: bestPlayerScannerJamReduction(a.Player),
			ShotsKind: shipShotsKind(sh), WeaponAmmo: NormalizeWeaponAmmo(sh.Weapon, sh.WeaponAmmo),
			CloakKind: shipCloakKind(sh), Cloaked: shipCloakKind(sh) != CloakNone,
			EnergyAbsorber: shipHasSpecial(sh, energyAbsorberName), Marines: ShipMarineComplement(sh),
			SecurityStations: shipHasSecurityStations(sh), MarineStrength: marineStrength, MarineHitsToKill: marineHits,
			BoardingBonus: gamedata.ShipCrewBoardingBonus(gamedata.CrewLevelForXP(sh.CrewXP, aiRaceHasTrait(a, gamedata.TRAIT_WARLORD))),
			CommandoBonus: commando, SecurityBonus: security, AssaultShuttles: shipHasAssaultShuttles(sh),
			Transporters: shipHasTransporters(sh), Charged: true,
			Initiative: gamedata.CombatInitiative(atk, gamedata.ShipCombatSpeed(drive, shipSizeClass(sh.Class), false, false)),
			SpriteIdx:  CombatSpriteForShip(sh, color), Bay: len(bays) > 0, BayKind: bayKind, Bays: append([]FighterKind(nil), bays...),
		}
		cs.ensureShieldFacings()
		out = append(out, cs)
	}
	return out
}
