package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestEventEmpireTargetUsesPopulationExtremesAcrossPlayerAndAI(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(17)
	s.PlayerColonies[0].Population = 10
	if len(s.AIPlayers) < 2 {
		t.Fatal("示範對局至少需要兩個 AI")
	}
	s.AIPlayers[0].Colonies[0].Population = 20
	s.AIPlayers[1].Colonies[0].Population = 40
	s.AIPlayers = s.AIPlayers[:2]

	bad := *gamedata.RandomEventByID(6)
	bad.Good = false
	target, ok := s.chooseEventEmpireTarget(bad, eventEmpireTarget{}, false)
	if !ok {
		t.Fatal("三個存活帝國應能選出壞事件目標")
	}
	if target.kind == eventEmpirePlayer {
		t.Fatal("壞事件必須先排除最低人口的目前玩家")
	}

	good := bad
	good.Good = true
	target, ok = s.chooseEventEmpireTarget(good, eventEmpireTarget{}, false)
	if !ok {
		t.Fatal("三個存活帝國應能選出好事件目標")
	}
	if target.kind == eventEmpireAI && target.index == 1 {
		t.Fatal("好事件必須先排除最高人口的 AI[1]")
	}
}

func TestLuckyGlobalScanIncrementsAllAndStopsAtFirstSuccess(t *testing.T) {
	s := NewDemoSession()
	if got := s.SetupHotseatWithAIIndices([]int{0}); got != 2 {
		t.Fatalf("需要兩個熱座席位，got %d", got)
	}
	mask := uint32(1) << uint(gamedata.TRAIT_LUCKY)
	for i := range s.Seats {
		s.Seats[i].RaceIndex = -1
		s.Seats[i].CustomRaceTraits = mask
		s.Seats[i].LuckyEventCounter = 8000
	}
	s.RaceIndex = -1
	s.CustomRaceTraits = mask
	s.LuckyEventCounter = 8000
	s.eventRand = newRandStream(1)

	target, ok := s.advanceAllLuckyEventCounters(50)
	if !ok || target.kind != eventEmpireSeat || target.index != 0 {
		t.Fatalf("第一個成功槽應是 seat[0]：target=%+v ok=%v", target, ok)
	}
	if s.Seats[0].LuckyEventCounter != 0 {
		t.Fatalf("成功槽應清零，got %d", s.Seats[0].LuckyEventCounter)
	}
	if s.Seats[1].LuckyEventCounter != 8001 {
		t.Fatalf("所有 Lucky 槽必須先累加，後續槽不再擲骰／清零，got %d", s.Seats[1].LuckyEventCounter)
	}
}

func TestHotseatIdleEmpireDoesNotRepeatGlobalEventSchedule(t *testing.T) {
	s := NewDemoSession()
	if got := s.SetupHotseatWithAIIndices([]int{0}); got != 2 {
		t.Fatalf("需要兩個熱座席位，got %d", got)
	}
	s.Turn = 51
	s.Difficulty = 2
	s.EventAttemptCounter = 0
	s.advanceEvents()
	if s.EventAttemptCounter != 1 {
		t.Fatalf("主席位應完成一次保護檢查，got %d", s.EventAttemptCounter)
	}
	s.advanceSeatEmpire()
	if s.EventAttemptCounter != 1 {
		t.Fatalf("閒置席位不得重跑全局事件排程，got %d", s.EventAttemptCounter)
	}
}

func TestAIEventWritebackUsesAIState(t *testing.T) {
	s := NewDemoSession()
	s.Turn = 100
	beforeAI := s.AIPlayers[0].Player.BC
	beforePlayer := s.Player.BC
	ev := *gamedata.RandomEventByID(6)
	result, ok := s.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireAI, index: 0})
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("AI 富商捐獻應有雙語結果：result=%+v ok=%v", result, ok)
	}
	if s.AIPlayers[0].Player.BC != beforeAI+500 {
		t.Fatalf("AI 國庫未回寫：got %d want %d", s.AIPlayers[0].Player.BC, beforeAI+500)
	}
	if s.Player.BC != beforePlayer {
		t.Fatalf("AI 事件不得誤寫目前玩家：%d → %d", beforePlayer, s.Player.BC)
	}
}

func TestAIPirateRaidUsesOriginalTreasuryPercentage(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(23)
	s.AIPlayers[0].Player.BC = 1000
	ev := *gamedata.RandomEventByID(15)
	result, ok := s.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireAI, index: 0})
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("AI 海盜劫掠應有雙語結果：result=%+v ok=%v", result, ok)
	}
	loss := 1000 - s.AIPlayers[0].Player.BC
	if loss < 300 || loss > 500 {
		t.Fatalf("原版 30..50%% 損失越界：loss=%d", loss)
	}

	s.AIPlayers[0].Player.BC = 99
	if _, ok := s.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireAI, index: 0}); ok {
		t.Fatal("AI 國庫不足 100 BC 時不得套用海盜劫掠")
	}
}

func TestAIComputerVirusUsesOriginalLossRange(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(29)
	s.AIPlayers[0].Player.ResearchProgress = 200
	ev := *gamedata.RandomEventByID(3)
	result, ok := s.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireAI, index: 0})
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("AI 電腦病毒應有雙語結果：result=%+v ok=%v", result, ok)
	}
	loss := 200 - s.AIPlayers[0].Player.ResearchProgress
	if loss < 51 || loss > 100 {
		t.Fatalf("原版 51..100 RP 損失越界：loss=%d", loss)
	}

	s.AIPlayers[0].Player.ResearchProgress = 9
	if _, ok := s.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireAI, index: 0}); ok {
		t.Fatal("AI 研究進度不足 10 RP 時不得套用電腦病毒")
	}
}

func TestHotseatEventWritebackTargetsInactiveSeat(t *testing.T) {
	s := NewDemoSession()
	if got := s.SetupHotseatWithAIIndices([]int{0}); got != 2 {
		t.Fatalf("需要兩個熱座席位，got %d", got)
	}
	s.Turn = 100
	beforeActive := s.Player.BC
	beforeInactive := s.Seats[1].Player.BC
	ev := *gamedata.RandomEventByID(6)
	result, ok := s.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireSeat, index: 1})
	if !ok || result.Message == "" {
		t.Fatalf("非目前席位事件應可結算：result=%+v ok=%v", result, ok)
	}
	if s.Seats[1].Player.BC != beforeInactive+500 {
		t.Fatalf("非目前席位國庫未回寫：got %d want %d", s.Seats[1].Player.BC, beforeInactive+500)
	}
	if s.Player.BC != beforeActive {
		t.Fatalf("結算後必須恢復目前席位：%d → %d", beforeActive, s.Player.BC)
	}
}
