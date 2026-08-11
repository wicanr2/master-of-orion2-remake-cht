package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func testLeader(skill gamedata.LeaderSkills, ship bool) Leader {
	return Leader{
		Name: "測試領袖", Level: 1, Ship: ship,
		Skills: []LeaderSkill{{ID: int(skill), Tier: 1}},
	}
}

func TestCommonLeaderEmpireEffects(t *testing.T) {
	plain := NewDemoSession()
	withOperations := NewDemoSession()
	withOperations.Leaders = []Leader{testLeader(gamedata.SKILL_OPERATIONS, false)}
	if got, want := withOperations.CommandPointsSupplyNow(), plain.CommandPointsSupplyNow()+
		gamedata.LeaderSkillBonus(int(gamedata.SKILL_OPERATIONS), 1, 0); got != want {
		t.Fatalf("Operations 供給 = %d, want %d", got, want)
	}

	ld := Leader{Name: "候選", Level: 1, Ship: false}
	baseCost := plain.MercHireCost(ld)
	withFamous := NewDemoSession()
	withFamous.Leaders = []Leader{testLeader(gamedata.SKILL_FAMOUS, false)}
	if got := withFamous.MercHireCost(ld); got >= baseCost {
		t.Fatalf("Famous 應降低後續雇用費: plain=%d famous=%d", baseCost, got)
	}

	withDiplomat := NewDemoSession()
	ai := &withDiplomat.AIPlayers[0]
	before := ai.Relation
	withDiplomat.Leaders = []Leader{testLeader(gamedata.SKILL_DIPLOMAT, false)}
	withDiplomat.DiplomacyResponse("trade", ai.Name)
	want := clampRelation(before + withDiplomat.diplomacyRelationGain(10))
	if ai.Relation != want {
		t.Fatalf("Diplomat 外交關係 = %d, want %d", ai.Relation, want)
	}
}

func TestMegawealthAddsEmpireIncome(t *testing.T) {
	plain := NewDemoSession()
	wealthy := NewDemoSession()
	plain.DisableEvents, wealthy.DisableEvents = true, true
	wealthy.Leaders = []Leader{testLeader(gamedata.SKILL_MEGAWEALTH, false)}

	plain.EndTurn()
	wealthy.EndTurn()
	want := gamedata.LeaderSkillBonus(int(gamedata.SKILL_MEGAWEALTH), 1, 0)
	if got := wealthy.Player.BC - plain.Player.BC; got != want {
		t.Fatalf("Megawealth 回合收入差 = %d, want %d", got, want)
	}
	if wealthy.LastPlayerOutput.Player.BC != wealthy.Player.BC {
		t.Fatalf("回合摘要國庫未同步: output=%d state=%d",
			wealthy.LastPlayerOutput.Player.BC, wealthy.Player.BC)
	}
}

func TestSpyLeaderBonusesReachBothSides(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = []Leader{
		testLeader(gamedata.SKILL_SPYMASTER, false),
		testLeader(gamedata.SKILL_TELEPATH, false),
	}
	spymaster := leaderEmpireSkillBonus(s.Leaders, gamedata.SKILL_SPYMASTER)
	telepath := leaderEmpireSkillBonus(s.Leaders, gamedata.SKILL_TELEPATH)
	baseAttack := spyAttackerBonus(s.Player, 0, s.raceSpyBonusForActions())
	withAttack := spyAttackerBonus(s.Player, 0, s.raceSpyBonusForActions()+spymaster)
	if withAttack-baseAttack != spymaster {
		t.Fatalf("Spymaster AB 差 = %d, want %d", withAttack-baseAttack, spymaster)
	}
	baseDefense := spyDefenderBonusWithAgents(s.Player, s.playerSpyGovernmentDefenseBonus(),
		s.raceSpyBonusForActions(), 0)
	withDefense := spyDefenderBonusWithAgents(s.Player, s.playerSpyGovernmentDefenseBonus(),
		s.raceSpyBonusForActions()+telepath, 0)
	if withDefense-baseDefense != telepath {
		t.Fatalf("Telepath DB 差 = %d, want %d", withDefense-baseDefense, telepath)
	}
}

func TestAssassinGetsAnIndependentTurnAction(t *testing.T) {
	found := false
	for seed := int64(0); seed < 500 && !found; seed++ {
		s := NewDemoSession()
		s.Leaders = []Leader{{Name: "刺客", Level: 5, Skills: []LeaderSkill{{
			ID: int(gamedata.SKILL_ASSASSIN), Tier: 2,
		}}}}
		s.AIPlayers[0].DefensiveAgents = 1
		s.spyRand = newRandStream(seed)
		s.advanceLeaderAssassinActions()
		found = s.AIPlayers[0].DefensiveAgents == 0
	}
	if !found {
		t.Fatal("500 個可重播種子都沒有刺客行動，表示 Assassin 沒有接到回合消費端")
	}
}

func TestGalacticLoreSeparatesChartKnowledgeFromFleetDetection(t *testing.T) {
	plain := &GameSession{
		Stars:             []Star{{X: 0, Y: 0, Owner: 1}, {X: 1, Y: 1}},
		PlayerColonyStars: []int{0},
		ColonyBuildings:   []map[string]bool{{}},
		Fleets:            []Fleet{{AtStar: 0}},
	}
	plainChart := plain.StarChartVisible()
	if plainChart[1] {
		t.Fatal("沒有 Galactic Lore 時，遠星不應直接出現在星圖")
	}

	lore := *plain
	lore.Leaders = []Leader{testLeader(gamedata.SKILL_GALACTIC_LORE, false)}
	chart := lore.StarChartVisible()
	if !chart[1] {
		t.Fatal("Galactic Lore 應立即揭露星圖天體")
	}
	if lore.VisibleStars()[1] {
		t.Fatal("Galactic Lore 不應把抽象敵方艦隊偵測誤變成全知")
	}
}

func TestCaptainOrdnanceAndSecurityReachBothCombatPaths(t *testing.T) {
	s := NewDemoSession()
	s.Fleets[0].Ships = []Ship{{Name: "測試戰艦", Class: "戰艦", Weapon: "雷射砲", Armor: "無裝甲", Shield: "無護盾"}}
	s.Leaders = []Leader{{Name: "艦長", Level: 1, Ship: true, Skills: []LeaderSkill{
		{ID: int(gamedata.SKILL_ORDNANCE), Tier: 1},
		{ID: int(gamedata.SKILL_SECURITY), Tier: 1},
	}}}
	if !s.AssignOfficerToShip(0, 0, 0) {
		t.Fatal("測試艦艇軍官指派失敗")
	}
	ordnance := gamedata.LeaderSkillBonus(int(gamedata.SKILL_ORDNANCE), 1, 0)
	player, _ := s.StartCombat("測試敵人")
	if len(player) != 1 {
		t.Fatalf("StartCombat 玩家艦數 = %d, want 1", len(player))
	}
	if player[0].WeaponMax-player[0].Attack != ordnance {
		t.Fatalf("StartCombat Ordnance 最大傷害差 = %d, want %d",
			player[0].WeaponMax-player[0].Attack, ordnance)
	}
	if player[0].SecurityBonus != gamedata.LeaderSkillBonus(int(gamedata.SKILL_SECURITY), 1, 0) {
		t.Fatalf("StartCombat SecurityBonus = %d", player[0].SecurityBonus)
	}
	quick := s.mkPlayerCombatants()
	if len(quick) != 1 || quick[0].wmax-quick[0].atk != ordnance || quick[0].securityBonus <= 0 {
		t.Fatalf("快速戰鬥未接 Ordnance/Security: %+v", quick)
	}
}
