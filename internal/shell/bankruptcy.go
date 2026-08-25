package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

type BankruptcyActionKind string

const (
	BankruptcySellBuilding  BankruptcyActionKind = "sell_building"
	BankruptcyScrapShip     BankruptcyActionKind = "scrap_ship"
	BankruptcyDismissSpy    BankruptcyActionKind = "dismiss_spy"
	BankruptcyDismissLeader BankruptcyActionKind = "dismiss_leader"
)

type BankruptcyAction struct {
	Kind        BankruptcyActionKind
	ColonyIndex int
	SlotIndex   int
	Name        string
	RecoveredBC int
}

type bankruptcyShipPass struct {
	typ                    gamedata.ShipType
	requireFriendlySupport bool
	requireIdleMission     bool
}

// bankruptcyShipPasses 保留 sub_EE0B0 @ 0xEE2C2..0xEE47A 的原始呼叫順序；三個
// 分段之間分別插入建築篩選 0、1，第三段後才裁撤間諜／戰鬥艦／領袖／建築篩選 2。
var bankruptcyShipPasses = [3][]bankruptcyShipPass{
	{
		{gamedata.OUTPOST_SHIP, true, true}, {gamedata.OUTPOST_SHIP, true, false},
		{gamedata.TRANSPORT_SHIP, true, true}, {gamedata.COMBAT_SHIP, true, true},
		{gamedata.OUTPOST_SHIP, false, true}, {gamedata.OUTPOST_SHIP, false, false},
		{gamedata.TRANSPORT_SHIP, false, true}, {gamedata.COMBAT_SHIP, false, true},
	},
	{
		{gamedata.OUTPOST_SHIP, true, true}, {gamedata.OUTPOST_SHIP, true, false},
		{gamedata.OUTPOST_SHIP, false, true}, {gamedata.OUTPOST_SHIP, false, false},
		{gamedata.TRANSPORT_SHIP, true, true}, {gamedata.COMBAT_SHIP, true, true},
		{gamedata.TRANSPORT_SHIP, false, true}, {gamedata.COMBAT_SHIP, false, true},
	},
	{
		{gamedata.TRANSPORT_SHIP, true, false}, {gamedata.TRANSPORT_SHIP, false, false},
		{gamedata.COLONY_SHIP, true, true}, {gamedata.COLONY_SHIP, false, true},
		{gamedata.COLONY_SHIP, true, false}, {gamedata.COLONY_SHIP, false, false},
	},
}

type bankruptcyBuildingCandidate struct {
	colony, id, score int
	name              string
	building          gamedata.Building
}

func (s *GameSession) bankruptcyBuildingCandidate(stage int) (bankruptcyBuildingCandidate, bool) {
	best := bankruptcyBuildingCandidate{}
	found := false
	for colony, built := range s.ColonyBuildings {
		ids := make(map[int]bool)
		for name, present := range built {
			if !present {
				continue
			}
			if id, ok := gamedata.OriginalBuildingIDForName(name); ok {
				ids[id] = true
			}
		}
		for name, present := range built {
			if !present {
				continue
			}
			id, ok := gamedata.OriginalBuildingIDForName(name)
			if !ok {
				continue
			}
			building, ok := gamedata.BuildingByNameZH(name)
			if !ok {
				continue
			}
			if !gamedata.BankruptcyBuildingEligible(stage, id, building.ProductionCost) {
				continue
			}
			score := gamedata.BankruptcyBuildingScore(id, building.ProductionCost, int(s.Government), func(other int) bool { return ids[other] })
			candidate := bankruptcyBuildingCandidate{colony: colony, id: id, score: score, name: name, building: building}
			if !found || candidate.score < best.score ||
				(candidate.score == best.score && (candidate.colony < best.colony ||
					(candidate.colony == best.colony && candidate.id < best.id))) {
				best, found = candidate, true
			}
		}
	}
	return best, found
}

func (s *GameSession) addBankruptcyRecovery(amount int) {
	if amount <= 0 {
		return
	}
	s.Player.BC += amount
	s.LastPlayerOutput.NetBC += amount
	s.LastPlayerOutput.Player.BC = s.Player.BC
}

func (s *GameSession) sellBankruptcyBuildings(stage int) {
	for s.Player.BC < 0 {
		candidate, ok := s.bankruptcyBuildingCandidate(stage)
		if !ok {
			return
		}
		delete(s.ColonyBuildings[candidate.colony], candidate.name)
		s.removeBuildingEffect(candidate.colony, candidate.name)
		refund := candidate.building.ProductionCost / 2
		s.addBankruptcyRecovery(refund)
		s.LastBankruptcy = append(s.LastBankruptcy, BankruptcyAction{
			Kind: BankruptcySellBuilding, ColonyIndex: candidate.colony, SlotIndex: -1,
			Name: candidate.name, RecoveredBC: refund,
		})
	}
}

func bankruptcyShipType(sh Ship) gamedata.ShipType {
	if sh.RawTypeKnown {
		return sh.RawType
	}
	switch sh.Class {
	case ColonyShipClass:
		return gamedata.COLONY_SHIP
	case OutpostShipClass:
		return gamedata.OUTPOST_SHIP
	case "運兵船", "運輸艦":
		return gamedata.TRANSPORT_SHIP
	default:
		return gamedata.COMBAT_SHIP
	}
}

func bankruptcyShipMissionEligible(sh Ship) bool {
	mission := uint8(0)
	if sh.RawMissionKnown {
		mission = sh.RawMission
	}
	return mission == 0 || mission == 3 || mission == 7
}

func (s *GameSession) bankruptcyShipAtFriendlySupport(fleet int) bool {
	if fleet < 0 || fleet >= len(s.Fleets) || s.Fleets[fleet].ETA > 0 {
		return false
	}
	star := s.Fleets[fleet].AtStar
	for _, own := range s.PlayerColonyStars {
		if own == star {
			return true
		}
	}
	for _, outpost := range s.Outposts {
		if outpost.StarIndex == star {
			return true
		}
	}
	return false
}

// 原版 dword_1AA3F0 是 sub_5EF17 對每個存活對手累加的戰鬥效能，並選最低值。
// remake 沒有該逐對手評估器，以下只近似其「先拆最弱艦」排序；type／位置／任務／退款
// 與呼叫順序仍直接依原版指令。這個近似不影響支援艦，因原版矩陣只填 type 0。
func bankruptcyShipScore(sh Ship) int {
	score := shipStrength(sh.Class) * 100
	if len(sh.WeaponMounts) == 0 {
		return score + sh.WeaponAttack
	}
	for _, mount := range sh.WeaponMounts {
		count := mount.WorkingCount
		if count <= 0 {
			count = mount.MaxCount
		}
		score += mount.Attack * count
	}
	return score
}

func (s *GameSession) scrapBankruptcyShips(pass bankruptcyShipPass) {
	for s.Player.BC < 0 {
		bestFleet, bestShip, bestScore := -1, -1, int(^uint(0)>>1)
		// 原版由最高 ship index 往低處掃描，分數相同時保留先遇到的最高索引。
		for fi := len(s.Fleets) - 1; fi >= 0; fi-- {
			friendly := s.bankruptcyShipAtFriendlySupport(fi)
			if pass.requireFriendlySupport && !friendly {
				continue
			}
			for si := len(s.Fleets[fi].Ships) - 1; si >= 0; si-- {
				ship := s.Fleets[fi].Ships[si]
				if bankruptcyShipType(ship) != pass.typ ||
					(pass.requireIdleMission && !bankruptcyShipMissionEligible(ship)) {
					continue
				}
				score := bankruptcyShipScore(ship)
				if score < bestScore {
					bestFleet, bestShip, bestScore = fi, si, score
				}
			}
		}
		if bestFleet < 0 {
			return
		}
		ship := s.Fleets[bestFleet].Ships[bestShip]
		refund := 0
		if s.bankruptcyShipAtFriendlySupport(bestFleet) {
			cost := ship.ProductionCost
			if cost <= 0 {
				cost = ShipCost(ship.Class)
			}
			refund = cost / 4
		}
		s.clearOfficerAssignment(ship.Name)
		s.Fleets[bestFleet].Ships = append(s.Fleets[bestFleet].Ships[:bestShip], s.Fleets[bestFleet].Ships[bestShip+1:]...)
		s.addBankruptcyRecovery(refund)
		s.LastBankruptcy = append(s.LastBankruptcy, BankruptcyAction{
			Kind: BankruptcyScrapShip, ColonyIndex: bestFleet, SlotIndex: bestShip,
			Name: ship.Name, RecoveredBC: refund,
		})
	}
}

func (s *GameSession) dismissBankruptcySpies() {
	for slot := range s.PlayerSpies {
		for s.Player.BC < 0 && s.PlayerSpies[slot] > 0 {
			s.PlayerSpies[slot]--
			if s.Player.SpyMaintenance > 0 {
				s.Player.SpyMaintenance--
			}
			if s.LastPlayerOutput.SpyMaintenanceCost > 0 {
				s.LastPlayerOutput.SpyMaintenanceCost--
			}
			s.addBankruptcyRecovery(spyMaintenancePerSpyBC)
			s.LastBankruptcy = append(s.LastBankruptcy, BankruptcyAction{
				Kind: BankruptcyDismissSpy, ColonyIndex: -1, SlotIndex: slot,
				RecoveredBC: spyMaintenancePerSpyBC,
			})
		}
	}
}

func (s *GameSession) dismissBankruptcyLeaders() {
	for i := 0; i < len(s.Leaders) && s.Player.BC < 0; {
		leader := s.Leaders[i]
		upkeep := leaderUpkeepCost(leader)
		if upkeep <= 0 {
			i++
			continue
		}
		if leader.Ship {
			s.clearOfficerAssignment(leader.Name)
		} else {
			s.unassignColonyLeaderByName(leader.Name)
		}
		copy(s.Leaders[i:], s.Leaders[i+1:])
		s.Leaders = s.Leaders[:len(s.Leaders)-1]
		if s.Player.OfficerMaintenance >= upkeep {
			s.Player.OfficerMaintenance -= upkeep
		}
		if s.LastPlayerOutput.OfficerMaintenanceCost >= upkeep {
			s.LastPlayerOutput.OfficerMaintenanceCost -= upkeep
		}
		s.addBankruptcyRecovery(upkeep)
		s.LastBankruptcy = append(s.LastBankruptcy, BankruptcyAction{
			Kind: BankruptcyDismissLeader, ColonyIndex: -1, SlotIndex: i,
			Name: leader.Name, RecoveredBC: upkeep,
		})
	}
}

// resolvePlayerBankruptcy 重現 sub_EE0B0 已證實的拆船、建築、間諜與領袖交錯順序。
func (s *GameSession) resolvePlayerBankruptcy() {
	s.LastBankruptcy = nil
	if s.Player.BC >= 0 {
		return
	}
	for _, pass := range bankruptcyShipPasses[0] {
		s.scrapBankruptcyShips(pass)
	}
	s.sellBankruptcyBuildings(0)
	for _, pass := range bankruptcyShipPasses[1] {
		s.scrapBankruptcyShips(pass)
	}
	s.sellBankruptcyBuildings(1)
	for _, pass := range bankruptcyShipPasses[2] {
		s.scrapBankruptcyShips(pass)
	}
	s.dismissBankruptcySpies()
	s.scrapBankruptcyShips(bankruptcyShipPass{typ: gamedata.COMBAT_SHIP, requireFriendlySupport: true})
	s.scrapBankruptcyShips(bankruptcyShipPass{typ: gamedata.COMBAT_SHIP})
	s.dismissBankruptcyLeaders()
	s.sellBankruptcyBuildings(2)
}

// removeBuildingEffect 是 applyBuildingEffect 的可逆對偶。氣候改良等已發生的世界狀態不回退。
func (s *GameSession) removeBuildingEffect(i int, name string) {
	if i < 0 || i >= len(s.PlayerColonies) {
		return
	}
	removeBuildingEffectFromColony(&s.PlayerColonies[i], name)
	s.recalcColonyMorale(i)
}

// removeBuildingEffectFromColony 是不依賴目前玩家席位的 typed 反向效果，供 AI
// 殖民地遭地震等世界事件摧毀建築時共用。氣候改良等已落地的世界狀態不回退。
func removeBuildingEffectFromColony(c *engine.ColonyState, name string) {
	if c == nil {
		return
	}
	subFloor := func(v *int, delta int) {
		*v -= delta
		if *v < 0 {
			*v = 0
		}
	}
	switch name {
	case "自動工廠":
		subFloor(&c.IndustryPerWorker, 1)
		subFloor(&c.FlatIndustry, 5)
	case "研究實驗室":
		subFloor(&c.FlatResearch, 5)
	case "太空港":
		subFloor(&c.IncomeBonusPercent, 50)
	case "機器人採礦廠":
		subFloor(&c.IndustryPerWorker, 2)
		subFloor(&c.FlatIndustry, 10)
	case "深層核心礦場":
		subFloor(&c.IndustryPerWorker, 3)
		subFloor(&c.FlatIndustry, 15)
	case "污染處理器":
		c.PollutionProcessor = false
	case "大氣更新器":
		c.AtmosphericRenewer = false
	case "核心廢料場":
		c.CoreWasteDump = false
	case "行星超級電腦":
		subFloor(&c.FlatResearch, 10)
	case "銀河網路中心":
		subFloor(&c.FlatResearch, 15)
	case "水耕農場":
		subFloor(&c.FlatFood, 2)
	case "地底農場":
		subFloor(&c.FlatFood, 4)
	case "氣候控制器":
		subFloor(&c.FoodPerFarmer, 2)
	case "行星證券交易所":
		subFloor(&c.IncomeBonusPercent, 100)
	case "太空大學":
		subFloor(&c.FoodPerFarmer, 1)
		subFloor(&c.IndustryPerWorker, 1)
		subFloor(&c.ResearchPerScientist, 1)
	case "生態圈":
		subFloor(&c.PopMax, 2)
	case "複製中心":
		subFloor(&c.FlatGrowth, gamedata.CloningCenterGrowthPoints)
	case "自動實驗室":
		subFloor(&c.FlatResearch, 30)
	case "食物複製機":
		c.FoodReplicators = false
	case "再生反應爐":
		c.Recyclotron = false
	case "行星重力產生器":
		c.NormalizeGravity = false
	case "機器人工廠":
		subFloor(&c.FlatIndustry, gamedata.ProdRoboticFactoryBonus(int(c.MineralRichness)))
	}
}
