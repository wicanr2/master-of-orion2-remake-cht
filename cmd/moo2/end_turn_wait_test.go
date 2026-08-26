package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestContinuousTurnsFollowEndOfTurnWaitSetting(t *testing.T) {
	s := shell.NewDemoSession()
	b := &sceneBuilder{session: s, animTick: 20}
	b.startContinuousTurns()
	if b.continuousTurns {
		t.Fatal("原版預設 End Of Turn Wait 開啟時不得連續推進")
	}
	settings := s.EffectiveGameSettings()
	settings.EndOfTurnWait = false
	s.ApplyGameSettings(settings)
	b.startContinuousTurns()
	if !b.continuousTurns || b.continuousTurnAt != 20+continuousTurnInterval {
		t.Fatalf("關閉等待後應排定下一回合：running=%v at=%d", b.continuousTurns, b.continuousTurnAt)
	}
}

func TestContinuousTurnsRejectHotseatAndNetwork(t *testing.T) {
	hotseat := shell.NewDemoSession()
	hotseat.SetupHotseat(2)
	settings := hotseat.EffectiveGameSettings()
	settings.EndOfTurnWait = false
	hotseat.ApplyGameSettings(settings)
	b := &sceneBuilder{session: hotseat}
	b.startContinuousTurns()
	if b.continuousTurns {
		t.Fatal("熱座不得自動跳過其他玩家交令")
	}

	s := shell.NewDemoSession()
	settings = s.EffectiveGameSettings()
	settings.EndOfTurnWait = false
	s.ApplyGameSettings(settings)
	b = &sceneBuilder{session: s, networkPending: true}
	b.startContinuousTurns()
	if b.continuousTurns {
		t.Fatal("網路對局不得繞過鎖步協議")
	}
}

func TestStopContinuousTurnsClearsSchedule(t *testing.T) {
	b := &sceneBuilder{continuousTurns: true, continuousTurnAt: 99}
	b.stopContinuousTurns()
	if b.continuousTurns || b.continuousTurnAt != 0 {
		t.Fatalf("停止後仍留有排程：running=%v at=%d", b.continuousTurns, b.continuousTurnAt)
	}
}
