package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalAIIncidentMemoryMirrorsAndFeedsThirdPartyBonus(t *testing.T) {
	s := NewDemoSession()
	s.ensureAIAIState()
	s.ensureOriginalAIAIRelations()
	s.recordOriginalAIIncident(1, 2, 4, -10)
	s.advanceOriginalAIIncidentMemory(func() int { return 1 })
	if s.AIIncidentMemoryRaw[1][2] != 1 || s.AIIncidentMemoryRaw[2][1] != 1 {
		t.Fatalf("事件記憶應雙向鏡射：%v", s.AIIncidentMemoryRaw)
	}
	// 讓談判除了 inner→third 的 +5 外全部為零；只驗證 consumer 讀到同一矩陣。
	s.AIReputationRaw[0][1] = 0
	s.AITreatyBiasRaw[0][1] = 0
	s.AIAgreementBiasRaw[0][1] = 0
	rolls := []int{1, 1, 1, 1}
	pos := 0
	s.advanceOriginalAIAINegotiation(0, 1, func(n int) int {
		value := rolls[pos]
		pos++
		return value
	})
	if pos == 0 {
		t.Fatal("談判應消費有事件記憶的 ordered pair")
	}
}

func TestOriginalAIIncidentRecordKeepsStrongestAndTreatyClearsMemory(t *testing.T) {
	s := NewDemoSession()
	s.ensureAIAIState()
	s.recordOriginalAIIncident(0, 1, 4, -12)
	s.recordOriginalAIIncident(0, 1, 5, 3)
	if s.AIIncidentReasonRaw[0][1] != 4 || s.AIIncidentMagnitudeRaw[0][1] != -12 {
		t.Fatalf("較弱事件不得覆蓋：reason=%d magnitude=%d", s.AIIncidentReasonRaw[0][1], s.AIIncidentMagnitudeRaw[0][1])
	}
	s.AIIncidentMemoryRaw[0][1], s.AIIncidentMemoryRaw[1][0] = 3, 3
	s.AIPolicies[0][1], s.AIPolicies[1][0] = gamedata.DIPLO_NONE, gamedata.DIPLO_NONE
	s.clearOriginalAIIncidentMemory(0, 1)
	if s.AIIncidentMemoryRaw[0][1] != 0 || s.AIIncidentMemoryRaw[1][0] != 0 {
		t.Fatalf("締約／宣戰／停戰 writer 應清事件記憶：%v", s.AIIncidentMemoryRaw)
	}
}

func TestOriginalAITreatyBetrayalIsDirectionalAndChangesCooldown(t *testing.T) {
	s := NewDemoSession()
	s.ensureAIAIState()
	s.ensureOriginalAIAIRelations()
	s.AIPlayers[1].RaceIndex = 0
	s.AIPolicies[0][1], s.AIPolicies[1][0] = gamedata.DIPLO_NON_AGGRESSION, gamedata.DIPLO_NON_AGGRESSION
	s.declareOriginalAIAIWar(0, 1, func(int) int { return 1 })
	if !s.AIIncidentBetrayalRaw[1][0] || s.AIIncidentBetrayalRaw[0][1] {
		t.Fatalf("宣戰違約記憶方向錯誤：%v", s.AIIncidentBetrayalRaw)
	}
	s.AIDiplomacyCooldownRaw[0][1] = 120
	s.AIPlayers[0].RaceIndex = -1
	delta, ok := gamedata.OriginalNPCTreatyCooldownDelta(originalAIRelationGovernment(s.AIPlayers[0]), false,
		gamedata.DIPLO_ALLIANCE, false)
	if !ok {
		t.Fatal("測試政府應可映射 cooldown")
	}
	s.addOriginalTreatyCooldown(0, 1, gamedata.DIPLO_ALLIANCE, false)
	if want := originalSignedByteAdd(120, delta); s.AIDiplomacyCooldownRaw[0][1] != want {
		t.Fatalf("+0x72F 應保留原版 signed-byte 寫回：%d", s.AIDiplomacyCooldownRaw[0][1])
	}
}
