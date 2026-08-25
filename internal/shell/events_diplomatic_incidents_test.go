package shell

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func makeAIAliveForDiplomaticIncident(a *AIOpponent) {
	if colonyPopulationTotal(a.Colonies) == 0 {
		a.Colonies = append(a.Colonies, engine.ColonyState{Population: 1})
	}
}

func TestDiplomaticIncidentRequiresWarAndUsesOriginalMarriageCap(t *testing.T) {
	s := NewDemoSession()
	for i := range s.AIPlayers {
		s.AIPlayers[i].Colonies = nil
	}
	makeAIAliveForDiplomaticIncident(&s.AIPlayers[0])
	s.AIPlayers[0].Relation = -40
	ev := *gamedata.RandomEventByID(5)
	if _, ok := s.applyRandomEventLocalized(ev); ok {
		t.Fatal("和平配對不得觸發原版事件 5")
	}
	s.AIPlayers[0].Treaty.FormalPolicy = gamedata.DIPLO_WAR
	result, ok := s.applyRandomEventLocalized(ev)
	if !ok || !strings.Contains(result.Message, stripAILabel(s.AIPlayers[0].Name)) {
		t.Fatalf("事件 5 應使用同一戰爭對象播報：ok=%v result=%+v", ok, result)
	}
	if s.AIPlayers[0].Relation != -10 {
		t.Fatalf("raw -25 戰爭上限應投影為 -10，got %d", s.AIPlayers[0].Relation)
	}
	if s.AIPlayers[0].Treaty.FormalPolicy != gamedata.DIPLO_WAR {
		t.Fatal("外交聯姻不得自行終止戰爭")
	}
}

func TestDiplomaticAssassinationUpdatesAIAISymmetrically(t *testing.T) {
	s := NewDemoSession()
	s.PlayerColonies = nil
	for i := range s.AIPlayers {
		s.AIPlayers[i].Colonies = nil
	}
	for i := 0; i < 2; i++ {
		makeAIAliveForDiplomaticIncident(&s.AIPlayers[i])
	}
	s.ensureAIRelations()
	s.AIPolicies = resizePolicyMatrix(s.AIPolicies, len(s.AIPlayers))
	s.AIPolicies[0][1], s.AIPolicies[1][0] = gamedata.DIPLO_WAR, gamedata.DIPLO_WAR
	s.AIRelations[0][1], s.AIRelations[1][0] = -10, -10
	result, ok := s.applyRandomEventLocalizedToTarget(*gamedata.RandomEventByID(4),
		eventEmpireTarget{kind: eventEmpireAI, index: 0})
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("AI／AI 事件 4 應成立：ok=%v result=%+v", ok, result)
	}
	if s.AIRelations[0][1] != s.AIRelations[1][0] || s.AIRelations[0][1] >= -10 {
		t.Fatalf("關係應對稱惡化：%d/%d", s.AIRelations[0][1], s.AIRelations[1][0])
	}
}

func TestDiplomaticIncidentPartnerReservoirIsReproducible(t *testing.T) {
	pick := func(seed int64) int {
		s := NewDemoSession()
		s.eventRand = newRandStream(seed)
		for i := range s.AIPlayers {
			makeAIAliveForDiplomaticIncident(&s.AIPlayers[i])
			s.AIPlayers[i].Treaty.FormalPolicy = gamedata.DIPLO_LIMITED_WAR
		}
		partner, ok := s.pickDiplomaticIncidentPartner(eventEmpireTarget{kind: eventEmpirePlayer})
		if !ok {
			t.Fatal("至少應有一個戰爭配對")
		}
		return partner.index
	}
	if a, b := pick(77), pick(77); a != b {
		t.Fatalf("同 seed 的 reservoir 選擇必須可重播：%d vs %d", a, b)
	}
}

func TestDiplomaticIncidentDoesNotRedrawAfterPickingPeace(t *testing.T) {
	s := NewDemoSession()
	for i := range s.AIPlayers {
		s.AIPlayers[i].Colonies = nil
	}
	makeAIAliveForDiplomaticIncident(&s.AIPlayers[0])
	makeAIAliveForDiplomaticIncident(&s.AIPlayers[1])
	s.AIPlayers[0].Treaty.FormalPolicy = gamedata.DIPLO_WAR
	s.AIPlayers[1].Treaty.FormalPolicy = gamedata.DIPLO_PEACE
	var seed int64 = -1
	for candidate := int64(0); candidate < 100; candidate++ {
		s.eventRand = newRandStream(candidate)
		partner, ok := s.pickDiplomaticIncidentPartner(eventEmpireTarget{kind: eventEmpirePlayer})
		if ok && partner.index == 1 {
			seed = candidate
			break
		}
	}
	if seed < 0 {
		t.Fatal("測試 seed 範圍內應能抽到和平對象")
	}
	s.eventRand = newRandStream(seed)
	if _, ok := s.applyDiplomaticIncident(4, eventEmpireTarget{kind: eventEmpirePlayer}); ok {
		t.Fatal("原版先抽後驗：抽到和平對象後候選應失敗，不得重抽戰爭對象")
	}
}

func TestDiplomaticIncidentWritesInactiveHotseatTarget(t *testing.T) {
	s := NewDemoSession()
	if s.SetupHotseatWithAIIndices([]int{0}) != 2 {
		t.Fatal("測試需要兩個熱座席位")
	}
	for i := range s.AIPlayers {
		s.AIPlayers[i].Colonies = nil
	}
	makeAIAliveForDiplomaticIncident(&s.AIPlayers[0])
	s.AIPlayers[0].Treaty.FormalPolicy = gamedata.DIPLO_WAR
	s.AIPlayers[0].Relation = -40
	result, ok := s.applyRandomEventLocalizedToTarget(*gamedata.RandomEventByID(5),
		eventEmpireTarget{kind: eventEmpireSeat, index: 1})
	if !ok || result.Message == "" || s.ActiveSeat != 0 {
		t.Fatalf("非目前席位應結算後恢復目前席位：ok=%v active=%d result=%+v", ok, s.ActiveSeat, result)
	}
	if s.AIPlayers[0].Relation != -10 {
		t.Fatalf("非目前席位事件應回寫共同外交邊：got %d", s.AIPlayers[0].Relation)
	}
}
