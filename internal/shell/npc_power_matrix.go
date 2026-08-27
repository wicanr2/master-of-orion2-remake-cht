package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

var originalFighterBeamEligible = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 9: true}

// originalBestFighterBeam 對應 sub_5699C 的一般帝國分支：只比較已知、帶靜態
// fighter flag 的 beam，DamageMax 必須嚴格較大才取代預設 Laser(raw 3)。
func originalBestFighterBeam(ps engine.PlayerState) int {
	known := knownTechnologyApplications(ps)
	bestID, bestDamage := 3, -1
	if weapon, ok := gamedata.OrigWeaponByID(bestID); ok {
		bestDamage = weapon.DamageMax
	}
	for id := 0; id < 40; id++ {
		weapon, ok := gamedata.OrigWeaponByID(id)
		if !ok || !originalFighterBeamEligible[id] || weapon.Cat != gamedata.WeaponCatBeam ||
			!known[weapon.Tech] || weapon.DamageMax <= bestDamage {
			continue
		}
		bestID, bestDamage = id, weapon.DamageMax
	}
	return bestID
}

// originalBestBomb 對應 sub_56CA2 的一般帝國分支。
func originalBestBomb(ps engine.PlayerState) int {
	known := knownTechnologyApplications(ps)
	bestID, bestDamage := 21, -1
	if weapon, ok := gamedata.OrigWeaponByID(bestID); ok {
		bestDamage = weapon.DamageMax
	}
	for id := 0; id < 40; id++ {
		weapon, ok := gamedata.OrigWeaponByID(id)
		if !ok || weapon.Cat != gamedata.WeaponCatBomb || !known[weapon.Tech] || weapon.DamageMax <= bestDamage {
			continue
		}
		bestID, bestDamage = id, weapon.DamageMax
	}
	return bestID
}

func originalShipSpecialWorking(sh Ship, rawID int) bool {
	ids := sh.SpecialIDs
	if len(ids) == 0 && len(sh.Specials) > 0 {
		ids = specialIDsFromMounts(sh.Specials)
	}
	for slot, id := range ids {
		if id != rawID {
			continue
		}
		return slot >= len(sh.DamagedSpecialsRaw) || sh.DamagedSpecialsRaw[slot] == 0
	}
	return false
}

func originalAICrewLevel(a AIOpponent, sh Ship) (int, bool) {
	if sh.CrewLevelRawKnown {
		level := int(sh.CrewLevelRaw)
		return level, level >= gamedata.CrewGreen && level <= gamedata.CrewUltraElite
	}
	return gamedata.CrewLevelForXP(sh.CrewXP, aiRaceHasTrait(a, gamedata.TRAIT_WARLORD)), true
}

func originalNPCPowerMounts(sh Ship) ([]gamedata.OriginalNPCPowerMount, bool) {
	if len(sh.WeaponMounts) == 0 {
		return nil, false
	}
	capacity := len(sh.WeaponMounts)
	if capacity > 8 {
		capacity = 8
	}
	mounts := make([]gamedata.OriginalNPCPowerMount, 0, capacity)
	for slot, mount := range sh.WeaponMounts {
		if slot >= 8 {
			break
		}
		if mount.RawType < 0 {
			return nil, false
		}
		mounts = append(mounts, gamedata.OriginalNPCPowerMount{
			WeaponID: mount.RawType, WorkingCount: mount.WorkingCount, RawMods: mount.RawMods, Ammo: mount.Ammo,
		})
	}
	return mounts, true
}

// originalNPCDirectionalFleetPower 產生 sub_5EF4B 對 owner 艦隊、observer 科技與
// 種族防禦的單向總值。任一必要 raw 欄缺失便失敗即關閉，讓呼叫端明確採舊存檔回退。
func (s *GameSession) originalNPCDirectionalFleetPower(owner, observer int) (int, bool) {
	if owner < 0 || observer < 0 || owner >= len(s.AIPlayers) || observer >= len(s.AIPlayers) {
		return 0, false
	}
	a, target := s.AIPlayers[owner], s.AIPlayers[observer]
	// 舊存檔與精簡測試可能只有抽象 FleetStrength。非零純量卻沒有實艦時不能把
	// 「沒有可逐艦計算的資料」誤判成原版精確零國力。
	if len(a.Ships) == 0 && a.FleetStrength > 0 {
		return 0, false
	}
	ownerBestComputer := originalBestComputer(a.Player)
	reduction, ok := gamedata.OriginalNPCComputerWeaponReduction(originalBestComputer(target.Player))
	if !ok {
		return 0, false
	}
	_, targetDefense := aiRaceCombatBonuses(target)
	observerDefense, ok := gamedata.OriginalNPCObserverDefense(aiDriveLevel(target), targetDefense,
		aiRaceHasTrait(target, gamedata.TRAIT_TRANS_DIMENSIONAL))
	if !ok {
		return 0, false
	}
	bestBeam, bestBomb := originalBestFighterBeam(a.Player), originalBestBomb(a.Player)
	ownerAttack, _ := aiRaceCombatBonuses(a)
	total := 0
	for _, sh := range a.Ships {
		if isSupportShipClass(sh.Class) {
			continue
		}
		if !sh.ComputerRawKnown || !sh.DesignSizeRawKnown || !sh.ArmorRawKnown || !sh.OriginalDamageKnown {
			return 0, false
		}
		mounts, valid := originalNPCPowerMounts(sh)
		if !valid {
			return 0, false
		}
		crew, valid := originalAICrewLevel(a, sh)
		if !valid {
			return 0, false
		}
		beamAttack, valid := gamedata.OriginalNPCShipBeamAttack(int(sh.ComputerRaw), crew,
			officerSkillBonusForShip(a.Leaders, sh, gamedata.SKILL_WEAPONRY), ownerAttack,
			originalShipSpecialWorking(sh, int(gamedata.SPEC_BATTLE_SCANNER)))
		if !valid {
			return 0, false
		}
		durability, valid := gamedata.OriginalNPCShipDurability(int(sh.DesignSizeRaw), int(sh.ArmorRaw),
			originalShipSpecialWorking(sh, int(gamedata.SPEC_REINFORCED_HULL)),
			originalShipSpecialWorking(sh, int(gamedata.SPEC_HEAVY_ARMOR)),
			sh.ArmorDamageRaw, sh.StructureDamageRaw)
		if !valid {
			return 0, false
		}
		power, valid := gamedata.OriginalNPCShipPower(gamedata.OriginalNPCShipPowerInput{
			Mounts: mounts, BeamAttack: beamAttack, ObserverDefense: observerDefense,
			ObserverWeaponReduction: reduction, OwnerBestComputer: ownerBestComputer,
			DesignComputer: int(sh.ComputerRaw), RemainingDurability: durability,
			BestBeamWeaponID: bestBeam, BestBombWeaponID: bestBomb,
		})
		if !valid {
			return 0, false
		}
		total += power
	}
	return total, true
}

// originalAIPowerMatrix 對每個 ordered pair 產生方向國力。exact=false 只會出現在
// 舊存檔缺 raw 欄位時；此時保留既有 FleetStrength 相容值，但不得宣稱原版精確。
func (s *GameSession) originalAIPowerMatrix() (power [][]int, exact [][]bool) {
	n := len(s.AIPlayers)
	power, exact = resizeIntMatrix(nil, n), resizeBoolMatrix(nil, n)
	for owner := 0; owner < n; owner++ {
		for observer := 0; observer < n; observer++ {
			if value, ok := s.originalNPCDirectionalFleetPower(owner, observer); ok {
				power[owner][observer], exact[owner][observer] = value, true
			} else {
				power[owner][observer] = s.AIPlayers[owner].FleetStrength
			}
		}
	}
	return power, exact
}
