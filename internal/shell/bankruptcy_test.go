package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func engineOutputForBankruptcy(player engine.PlayerState, spy, officer int) engine.EmpireOutput {
	return engine.EmpireOutput{Player: player, SpyMaintenanceCost: spy, OfficerMaintenanceCost: officer}
}

func TestBankruptcySellsBuildingForHalfAndReversesEffect(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = nil
	s.DisableEvents = true
	s.ColonyBuildings[0] = map[string]bool{"自動工廠": true}
	s.applyBuildingEffect(0, "自動工廠")
	beforeWorker, beforeFlat := s.PlayerColonies[0].IndustryPerWorker, s.PlayerColonies[0].FlatIndustry
	s.Player.BC = -20
	s.LastPlayerOutput.Player = s.Player
	s.resolvePlayerBankruptcy()
	if s.Player.BC != 10 {
		t.Fatalf("自動工廠半價退款應使 BC -20→10，got %d", s.Player.BC)
	}
	if s.ColonyBuildings[0]["自動工廠"] {
		t.Fatal("出售後建築旗標仍存在")
	}
	if got := s.PlayerColonies[0].IndustryPerWorker; got != beforeWorker-1 {
		t.Fatalf("每工人產能未反向撤銷：got %d want %d", got, beforeWorker-1)
	}
	if got := s.PlayerColonies[0].FlatIndustry; got != beforeFlat-5 {
		t.Fatalf("固定產能未反向撤銷：got %d want %d", got, beforeFlat-5)
	}
	if len(s.LastBankruptcy) != 1 || s.LastBankruptcy[0].RecoveredBC != 30 {
		t.Fatalf("處分報告錯誤：%+v", s.LastBankruptcy)
	}
}

func TestBankruptcyUsesFirstBuildingFilterBeforeProtectedCategory(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = nil
	s.ColonyBuildings[0] = map[string]bool{"自動工廠": true, "研究實驗室": true}
	s.Player.BC = -20
	s.LastPlayerOutput.Player = s.Player
	s.resolvePlayerBankruptcy()
	if s.ColonyBuildings[0]["研究實驗室"] {
		t.Fatal("第一輪應先出售 raw category 0 的研究實驗室")
	}
	if !s.ColonyBuildings[0]["自動工廠"] {
		t.Fatal("raw category 3 的自動工廠應留到第二輪")
	}
}

func TestBankruptcyDismissesOnlyNeededSpies(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = nil
	s.ColonyBuildings = nil
	s.PlayerSpies = []int{3}
	s.Player.BC = -2
	s.Player.SpyMaintenance = 3
	s.LastPlayerOutput = engineOutputForBankruptcy(s.Player, 3, 0)
	s.resolvePlayerBankruptcy()
	if s.Player.BC != 0 || s.PlayerSpies[0] != 1 {
		t.Fatalf("應裁撤兩名間諜補足赤字：BC=%d spies=%v", s.Player.BC, s.PlayerSpies)
	}
	if s.LastPlayerOutput.SpyMaintenanceCost != 1 {
		t.Fatalf("摘要間諜維護費未同步，got %d", s.LastPlayerOutput.SpyMaintenanceCost)
	}
}

func TestBankruptcyDismissesLeaderAfterSpies(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = nil
	s.ColonyBuildings = nil
	s.PlayerSpies = nil
	s.Leaders = demoLeaders()
	for len(s.Leaders) > 0 && leaderUpkeepCost(s.Leaders[0]) == 0 {
		s.Leaders = s.Leaders[1:]
	}
	if len(s.Leaders) == 0 {
		t.Fatal("測試資料需要至少一位有維護費的領袖")
	}
	upkeep := leaderUpkeepCost(s.Leaders[0])
	s.Player.BC = -1
	s.Player.OfficerMaintenance = upkeep
	s.LastPlayerOutput = engineOutputForBankruptcy(s.Player, 0, upkeep)
	before := len(s.Leaders)
	s.resolvePlayerBankruptcy()
	if len(s.Leaders) != before-1 || s.Player.BC < 0 {
		t.Fatalf("領袖未依序解雇：leaders=%d→%d BC=%d", before, len(s.Leaders), s.Player.BC)
	}
}

func TestBankruptcyNoAssetsKeepsNegativeTreasury(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = nil
	s.ColonyBuildings, s.PlayerSpies, s.Leaders = nil, nil, nil
	s.Player.BC = -7
	s.LastPlayerOutput.Player = s.Player
	s.resolvePlayerBankruptcy()
	if s.Player.BC != -7 || len(s.LastBankruptcy) != 0 {
		t.Fatalf("無可處分資產時不得憑空補錢：BC=%d actions=%+v", s.Player.BC, s.LastBankruptcy)
	}
}

func TestBankruptcyNonNegativeNoOp(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = nil
	s.Player.BC = 0
	before := len(s.ColonyBuildings[0])
	s.resolvePlayerBankruptcy()
	if len(s.ColonyBuildings[0]) != before || len(s.LastBankruptcy) != 0 {
		t.Fatal("非負國庫不應處分資產")
	}
}

func TestEndTurnRunsBankruptcyOnNormalPlayerPath(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = nil
	s.DisableEvents = true
	s.ColonyBuildings[0] = map[string]bool{"研究實驗室": true}
	s.Player.BC = -100
	s.EndTurn()
	if s.ColonyBuildings[0]["研究實驗室"] {
		t.Fatal("正常 EndTurn 玩家路徑未執行負國庫處分")
	}
	if len(s.LastBankruptcy) == 0 || s.LastBankruptcy[0].Kind != BankruptcySellBuilding {
		t.Fatalf("正常玩家路徑沒有結構化處分報告：%+v", s.LastBankruptcy)
	}
}

func TestBankruptcyScrapsOutpostBeforeBuildingsAndRefundsQuarterAtFriendlyStar(t *testing.T) {
	s := NewDemoSession()
	s.ColonyBuildings[0] = map[string]bool{"研究實驗室": true}
	s.Fleets = []Fleet{{AtStar: s.PlayerColonyStars[0], DestStar: -1, Ships: []Ship{{
		Name: "前哨一號", Class: OutpostShipClass, RawType: gamedata.OUTPOST_SHIP,
		RawTypeKnown: true, RawMissionKnown: true, ProductionCost: 60,
	}}}}
	s.Player.BC = -10
	s.LastPlayerOutput.Player = s.Player
	s.resolvePlayerBankruptcy()
	if s.Player.BC != 5 || len(s.Fleets[0].Ships) != 0 {
		t.Fatalf("友軍星前哨船應先拆並退 1/4：BC=%d ships=%d", s.Player.BC, len(s.Fleets[0].Ships))
	}
	if !s.ColonyBuildings[0]["研究實驗室"] || len(s.LastBankruptcy) != 1 ||
		s.LastBankruptcy[0].Kind != BankruptcyScrapShip || s.LastBankruptcy[0].RecoveredBC != 15 {
		t.Fatalf("處分順序或報告錯誤：%+v buildings=%v", s.LastBankruptcy, s.ColonyBuildings[0])
	}
}

func TestBankruptcyMissionRestrictedPassDefersMovingOutpost(t *testing.T) {
	s := NewDemoSession()
	s.ColonyBuildings = nil
	s.PlayerColonyStars = []int{0}
	s.Fleets = []Fleet{{AtStar: 0, DestStar: -1, Ships: []Ship{{
		Name: "任務中前哨船", Class: OutpostShipClass, RawType: gamedata.OUTPOST_SHIP, RawTypeKnown: true,
		RawMission: 1, RawMissionKnown: true, ProductionCost: 60,
	}}}}
	s.Player.BC = -20
	s.LastPlayerOutput.Player = s.Player
	s.resolvePlayerBankruptcy()
	if len(s.Fleets[0].Ships) != 0 || len(s.LastBankruptcy) != 1 {
		t.Fatalf("任務 1 應略過受限 pass、再由 unrestricted pass 拆除：%+v", s.LastBankruptcy)
	}
}

func TestBankruptcyAwayShipScrapsWithoutRefund(t *testing.T) {
	s := NewDemoSession()
	s.ColonyBuildings, s.PlayerSpies, s.Leaders = nil, nil, nil
	s.Fleets = []Fleet{{AtStar: 3, DestStar: 4, ETA: 2, Ships: []Ship{{
		Name: "遠航殖民船", Class: ColonyShipClass, RawType: gamedata.COLONY_SHIP,
		RawTypeKnown: true, RawMissionKnown: true, ProductionCost: 120,
	}}}}
	s.Player.BC = -7
	s.LastPlayerOutput.Player = s.Player
	s.resolvePlayerBankruptcy()
	if s.Player.BC != -7 || len(s.Fleets[0].Ships) != 0 || len(s.LastBankruptcy) != 1 ||
		s.LastBankruptcy[0].RecoveredBC != 0 {
		t.Fatalf("非友軍支援點應拆船但不退款：BC=%d actions=%+v", s.Player.BC, s.LastBankruptcy)
	}
}
