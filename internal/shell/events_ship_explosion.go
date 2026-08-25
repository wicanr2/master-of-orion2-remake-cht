package shell

// shipExplosionResult 是事件 8 的 typed 結果。原版只移除一艘所選艦艇；
// OfficerName 非空表示該艦軍官一併死亡。
type shipExplosionResult struct {
	Lost        Ship
	OfficerName string
}

// killOfficerOnDestroyedShip 把原版軍官 status=0xFE 投影為從可用領袖 slice 移除。
// remake 的 OfficerID 零也是合法 ID，因此只有艦上名稱非空時才判定有指派。
func killOfficerOnDestroyedShip(leaders []Leader, sh Ship) ([]Leader, string) {
	if sh.OfficerName == "" {
		return leaders, ""
	}
	for i := range leaders {
		if !leaders[i].Ship || leaders[i].Name != sh.OfficerName {
			continue
		}
		if sh.OfficerID != 0 && leaders[i].ID != sh.OfficerID {
			continue
		}
		name := leaders[i].Name
		copy(leaders[i:], leaders[i+1:])
		return leaders[:len(leaders)-1], name
	}
	return leaders, ""
}

func eventLeaderUpkeepTotal(leaders []Leader) int {
	total := 0
	for _, leader := range leaders {
		total += leaderUpkeepCost(leader)
	}
	return total
}

// resolvePlayerShipExplosion 對映 sub_23CED 的單趟 reservoir sampling 與
// sub_206A2 case 8 的單艦移除。正常存在於 Fleets 的艦艇是 raw status 0／1 的
// typed projection；沒有「至少留一艘」或鄰艦傷害。
func (s *GameSession) resolvePlayerShipExplosion() (shipExplosionResult, bool) {
	var out shipExplosionResult
	if s == nil {
		return out, false
	}
	s.eventRandForTest()
	candidates := 0
	selectedFleet, selectedShip := -1, -1
	for fi := range s.Fleets {
		for si := range s.Fleets[fi].Ships {
			candidates++
			if s.eventRand.Intn(candidates) == 0 {
				selectedFleet, selectedShip = fi, si
			}
		}
	}
	if selectedFleet < 0 {
		return out, false
	}
	out.Lost = s.Fleets[selectedFleet].Ships[selectedShip]
	s.Leaders, out.OfficerName = killOfficerOnDestroyedShip(s.Leaders, out.Lost)
	s.Fleets[selectedFleet].Ships = append(s.Fleets[selectedFleet].Ships[:selectedShip],
		s.Fleets[selectedFleet].Ships[selectedShip+1:]...)
	s.Player.OfficerMaintenance = eventLeaderUpkeepTotal(s.Leaders)
	s.Player.UsedCommandPoints = s.usedCommandPoints()
	return out, true
}

func (s *GameSession) resolveAIShipExplosion(aiIndex int) (shipExplosionResult, bool) {
	var out shipExplosionResult
	if s == nil || aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return out, false
	}
	a := &s.AIPlayers[aiIndex]
	if len(a.Ships) == 0 {
		return out, false
	}
	s.eventRandForTest()
	selected := -1
	for i := range a.Ships {
		if s.eventRand.Intn(i+1) == 0 {
			selected = i
		}
	}
	out.Lost = a.Ships[selected]
	a.Leaders, out.OfficerName = killOfficerOnDestroyedShip(a.Leaders, out.Lost)
	a.Ships = append(a.Ships[:selected], a.Ships[selected+1:]...)
	a.Player.OfficerMaintenance = eventLeaderUpkeepTotal(a.Leaders)
	s.syncAIShipStrength(aiIndex)
	return out, true
}
