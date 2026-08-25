package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func prepareSurrenderFixture(t *testing.T) *GameSession {
	t.Helper()
	s := NewDemoSession()
	if len(s.AIPlayers) < 2 {
		t.Fatal("投降測試需要至少兩個 AI 帝國")
	}
	s.ensureStatusEmpireBaseline()
	loser, receiver := &s.AIPlayers[0], &s.AIPlayers[1]
	loser.Player.BC, loser.Player.ActiveFreighters = 123, 7
	loser.Ships = []Ship{{Name: "應拆除艦", Class: "戰艦"}}
	loser.ShipBuildProgress = 44
	loser.Leaders = []Leader{{ID: 77, Name: "投降領袖", RawLocation: 3, RawETA: 5, RawPlayerIndex: 1}}
	if len(loser.ColonyLeaderNames) < len(loser.Colonies) {
		loser.ColonyLeaderNames = make([]string, len(loser.Colonies))
	}
	loser.ColonyLeaderNames[0] = "投降領袖"
	grantTechnologyApplication(&loser.Player, gamedata.TOPIC_PHYSICS, gamedata.TECH_LASER_CANNON)
	receiver.Player.BC, receiver.Player.ActiveFreighters = 10, 2
	s.ensureAIAIState()
	s.ensureAIRelations()
	s.AIWars[0][1], s.AIWars[1][0] = true, true
	s.AIPolicies[0][1], s.AIPolicies[1][0] = gamedata.DIPLO_WAR, gamedata.DIPLO_WAR
	s.AITrade[0][1], s.AITrade[1][0] = true, true
	s.AIResearch[0][1], s.AIResearch[1][0] = true, true
	s.PlayerSpies = make([]int, len(s.AIPlayers))
	s.PlayerSpies[0], s.PlayerSpies[1] = 9, 60
	return s
}

func TestEmpireSurrenderQueuesEventBeforeDeferredTransfer(t *testing.T) {
	s := prepareSurrenderFixture(t)
	beforeColonies := len(s.AIPlayers[0].Colonies)
	if !s.queueEmpireSurrender(0, eventEmpireTarget{kind: eventEmpireAI, index: 1, alive: true}) {
		t.Fatal("合法投降應建立 pending record")
	}
	if len(s.PendingSurrenders) != 1 || len(s.StatusBroadcast.Queue) != 1 {
		t.Fatalf("pending／新聞未同時建立：pending=%+v queue=%+v", s.PendingSurrenders, s.StatusBroadcast.Queue)
	}
	report := s.StatusBroadcast.Queue[0]
	if report.EventID != 34 || report.TargetIndex != 0 || report.SecondaryTargetKind != "ai" || report.SecondaryTargetIndex != 1 {
		t.Fatalf("事件 34 雙帝國 record 錯誤：%+v", report)
	}
	if len(s.AIPlayers[0].Colonies) != beforeColonies {
		t.Fatal("setter 階段不得提早搬資產")
	}
	r := s.snapshot().restore()
	if len(r.PendingSurrenders) != 1 || r.PendingSurrenders[0].SurrenderAI != 0 || r.PendingSurrenders[0].ReceiverIndex != 1 {
		t.Fatalf("pending surrender 未隨快照往返：%+v", r.PendingSurrenders)
	}
}

func TestEmpireSurrenderAIToAITransfersOriginalAssetSet(t *testing.T) {
	s := prepareSurrenderFixture(t)
	loserColonies, receiverColonies := len(s.AIPlayers[0].Colonies), len(s.AIPlayers[1].Colonies)
	if !s.queueEmpireSurrender(0, eventEmpireTarget{kind: eventEmpireAI, index: 1, alive: true}) {
		t.Fatal("建立投降失敗")
	}
	s.resolvePendingSurrenders()
	loser, receiver := &s.AIPlayers[0], &s.AIPlayers[1]
	if len(receiver.Colonies) != receiverColonies+loserColonies || len(receiver.ColonyStars) != len(receiver.Colonies) ||
		len(receiver.ColonyBuildings) != len(receiver.Colonies) || len(receiver.ColonyMarines) != len(receiver.Colonies) {
		t.Fatalf("殖民地平行陣列未完整移交：colonies=%d stars=%d buildings=%d marines=%d",
			len(receiver.Colonies), len(receiver.ColonyStars), len(receiver.ColonyBuildings), len(receiver.ColonyMarines))
	}
	if receiver.Player.BC != 133 || receiver.Player.ActiveFreighters != 9 {
		t.Fatalf("國庫／貨運艦移交錯誤：BC=%d freighters=%d", receiver.Player.BC, receiver.Player.ActiveFreighters)
	}
	if !playerStateKnowsTech(receiver.Player, gamedata.TOPIC_PHYSICS, gamedata.TECH_LASER_CANNON) {
		t.Fatal("接收者未取得投降者已知科技")
	}
	if len(receiver.Leaders) == 0 || receiver.Leaders[len(receiver.Leaders)-1].RawLocation != -1 ||
		receiver.Leaders[len(receiver.Leaders)-1].RawETA != 0 {
		t.Fatalf("領袖未移交或仍有任命：%+v", receiver.Leaders)
	}
	if len(loser.Colonies) != 0 || len(loser.Ships) != 0 || loser.ShipBuildProgress != 0 || loser.Player.BC != 0 || loser.Player.ActiveFreighters != 0 {
		t.Fatalf("投降者未完整清理：%+v", loser)
	}
	if s.PlayerSpies[0] != 0 || s.PlayerSpies[1] != 63 {
		t.Fatalf("玩家派駐間諜未依原版上限改派：%v", s.PlayerSpies)
	}
	if s.AIWars[0][1] || s.AIPolicies[0][1] != gamedata.DIPLO_NONE || s.AITrade[0][1] || s.AIResearch[0][1] {
		t.Fatal("投降後外交／協議未清空")
	}
	s.detectEmpireEliminationBroadcasts()
	for _, report := range s.StatusBroadcast.Queue {
		if report.EventID == 29 {
			t.Fatal("投降不得再重複建立帝國滅亡事件 29")
		}
	}
}

func TestEmpireSurrenderAIToPlayerTransfersColoniesButDestroysShips(t *testing.T) {
	s := prepareSurrenderFixture(t)
	beforeColonies, beforeBC, beforeFreighters := len(s.PlayerColonies), s.Player.BC, s.Player.ActiveFreighters
	loserColonies := len(s.AIPlayers[0].Colonies)
	if !s.queueEmpireSurrender(0, eventEmpireTarget{kind: eventEmpirePlayer, index: 0, alive: true}) {
		t.Fatal("建立 AI→玩家投降失敗")
	}
	s.resolvePendingSurrenders()
	if len(s.PlayerColonies) != beforeColonies+loserColonies || len(s.PlayerColonyStars) != len(s.PlayerColonies) ||
		len(s.ColonyBuildings) != len(s.PlayerColonies) || len(s.PlayerColonyTanks) != len(s.PlayerColonies) {
		t.Fatalf("玩家殖民地平行陣列未完整接收：colonies=%d stars=%d buildings=%d tanks=%d",
			len(s.PlayerColonies), len(s.PlayerColonyStars), len(s.ColonyBuildings), len(s.PlayerColonyTanks))
	}
	if s.Player.BC != beforeBC+123 || s.Player.ActiveFreighters != beforeFreighters+7 {
		t.Fatalf("玩家未取得國庫／貨運艦：BC=%d freighters=%d", s.Player.BC, s.Player.ActiveFreighters)
	}
	if len(s.AIPlayers[0].Ships) != 0 {
		t.Fatal("投降者艦艇應刪除而非移交")
	}
	for _, star := range s.PlayerColonyStars[beforeColonies:] {
		if star >= 0 && star < len(s.Stars) && s.Stars[star].Owner != 1 {
			t.Fatalf("接收殖民星 %d 未標成玩家所有", star)
		}
	}
}

func TestEmpireSurrenderRejectsInvalidReceiverWithoutPartialTransfer(t *testing.T) {
	s := prepareSurrenderFixture(t)
	s.PendingSurrenders = []EmpireSurrender{{SurrenderAI: 0, ReceiverKind: "ai", ReceiverIndex: 99}}
	before := len(s.AIPlayers[0].Colonies)
	s.resolvePendingSurrenders()
	if len(s.AIPlayers[0].Colonies) != before || len(s.PendingSurrenders) != 0 {
		t.Fatal("非法接收者應失敗即關閉且不做半套轉移")
	}
}

func TestEmpireSurrenderCanTransferToInactiveHotseatSeat(t *testing.T) {
	s := prepareSurrenderFixture(t)
	if s.SetupHotseat(2) != 2 {
		t.Fatal("需要兩席熱座")
	}
	s.ensureStatusEmpireBaseline()
	beforeTarget := len(s.Seats[1].PlayerColonies)
	beforeActive := len(s.PlayerColonies)
	loserColonies := len(s.AIPlayers[0].Colonies)
	if !s.queueEmpireSurrender(0, eventEmpireTarget{kind: eventEmpireSeat, index: 1, alive: true}) {
		t.Fatal("建立 AI→非目前熱座席位投降失敗")
	}
	s.resolvePendingSurrenders()
	if len(s.Seats[1].PlayerColonies) != beforeTarget+loserColonies {
		t.Fatalf("非目前席位未取得殖民地：got=%d want=%d", len(s.Seats[1].PlayerColonies), beforeTarget+loserColonies)
	}
	if len(s.PlayerColonies) != beforeActive || s.ActiveSeat != 0 {
		t.Fatal("轉移非目前席位後必須恢復目前席位狀態")
	}
}

func TestAdvanceAISurrendersRequiresWarAndOverwhelmingPower(t *testing.T) {
	s := prepareSurrenderFixture(t)
	s.EnableAIVsAI = true
	s.AIPlayers[0].Player.BC, s.AIPlayers[0].FleetStrength = 0, 0
	s.AIPlayers[0].Ships = nil
	for i := range s.AIPlayers[0].Colonies {
		s.AIPlayers[0].Colonies[i].Population = 1
	}
	s.AIPlayers[1].FleetStrength = 500
	s.advanceAISurrenders()
	if len(s.PendingSurrenders) != 1 || s.PendingSurrenders[0].SurrenderAI != 0 || s.PendingSurrenders[0].ReceiverIndex != 1 {
		t.Fatalf("戰爭中壓倒性弱勢 AI 應建立投降：%+v", s.PendingSurrenders)
	}

	peace := prepareSurrenderFixture(t)
	peace.EnableAIVsAI = true
	peace.AIWars[0][1], peace.AIWars[1][0] = false, false
	peace.AIPolicies[0][1], peace.AIPolicies[1][0] = gamedata.DIPLO_PEACE, gamedata.DIPLO_PEACE
	peace.AIPlayers[0].FleetStrength, peace.AIPlayers[1].FleetStrength = 0, 500
	peace.advanceAISurrenders()
	if len(peace.PendingSurrenders) != 0 {
		t.Fatal("沒有戰爭不得只因國力差距投降")
	}
}
