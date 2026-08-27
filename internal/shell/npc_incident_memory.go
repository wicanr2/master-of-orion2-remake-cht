package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

func (s *GameSession) ensureOriginalAIIncidentState() {
	n := len(s.AIPlayers)
	s.AIIncidentReasonRaw = resizeIntMatrix(s.AIIncidentReasonRaw, n)
	s.AIIncidentMagnitudeRaw = resizeIntMatrix(s.AIIncidentMagnitudeRaw, n)
	s.AIIncidentMemoryRaw = resizeIntMatrix(s.AIIncidentMemoryRaw, n)
	s.AIIncidentBetrayalRaw = resizeBoolMatrix(s.AIIncidentBetrayalRaw, n)
}

// recordOriginalAIIncident 對應 Change_Relations_ 只保留絕對值較大的待處理變動。
func (s *GameSession) recordOriginalAIIncident(actor, target, reason, appliedDelta int) {
	if actor < 0 || target < 0 || actor >= len(s.AIPlayers) || target >= len(s.AIPlayers) || actor == target ||
		reason == 0 || appliedDelta == 0 {
		return
	}
	s.ensureOriginalAIIncidentState()
	if absInt(s.AIIncidentMagnitudeRaw[actor][target]) >= absInt(appliedDelta) {
		return
	}
	s.AIIncidentReasonRaw[actor][target] = reason
	s.AIIncidentMagnitudeRaw[actor][target] = appliedDelta
}

func (s *GameSession) clearOriginalAIIncidentMemory(a, b int) {
	s.ensureOriginalAIIncidentState()
	if a < 0 || b < 0 || a >= len(s.AIPlayers) || b >= len(s.AIPlayers) {
		return
	}
	s.AIIncidentMemoryRaw[a][b], s.AIIncidentMemoryRaw[b][a] = 0, 0
}

func originalSignedByteAdd(current, delta int) int {
	return int(int8(uint8(current + delta)))
}

func (s *GameSession) addOriginalTreatyCooldown(a, b int, policy gamedata.ForeignPolicy, tribute bool) {
	s.ensureOriginalAIIncidentState()
	if a < 0 || b < 0 || a >= len(s.AIPlayers) || b >= len(s.AIPlayers) {
		return
	}
	for _, pair := range [][2]int{{a, b}, {b, a}} {
		owner, target := pair[0], pair[1]
		delta, ok := gamedata.OriginalNPCTreatyCooldownDelta(
			originalAIRelationGovernment(s.AIPlayers[owner]),
			s.AIIncidentBetrayalRaw[owner][target], policy, tribute)
		if ok {
			s.AIDiplomacyCooldownRaw[owner][target] = originalSignedByteAdd(s.AIDiplomacyCooldownRaw[owner][target], delta)
		}
	}
}

// markOriginalAITreatyBetrayal 對應 sub_5138E：actor 破壞既有正式條約時，
// 在 target 看向 actor 的 +0x727 方向留下持久旗標。
func (s *GameSession) markOriginalAITreatyBetrayal(actor, target int) {
	s.ensureOriginalAIIncidentState()
	if actor < 0 || target < 0 || actor >= len(s.AIPlayers) || target >= len(s.AIPlayers) {
		return
	}
	if s.AIPolicies[actor][target] != gamedata.DIPLO_NONE &&
		s.AIPolicies[target][actor] < gamedata.DIPLO_LIMITED_WAR {
		s.AIIncidentBetrayalRaw[target][actor] = true
	}
}

// advanceOriginalAIIncidentMemory 對應 sub_252D5 的 ordered pair 掃描。pending reason／
// magnitude 保持方向性，+0x71F 每次處理後鏡射到反方向。
func (s *GameSession) advanceOriginalAIIncidentMemory(roll100 func() int) {
	s.ensureOriginalAIIncidentState()
	for actor := range s.AIPlayers {
		for target := range s.AIPlayers {
			if actor == target {
				continue
			}
			protected := (s.AIPolicies[actor][target] >= gamedata.DIPLO_NON_AGGRESSION &&
				s.AIPolicies[actor][target] <= gamedata.DIPLO_PEACE) ||
				s.AITrade[actor][target] || s.AIResearch[actor][target] || s.AITributeModes[actor][target] < 0
			out, ok := gamedata.OriginalNPCIncidentMemoryStep(gamedata.OriginalNPCIncidentMemoryInput{
				PendingReason:       s.AIIncidentReasonRaw[actor][target],
				PendingMagnitude:    s.AIIncidentMagnitudeRaw[actor][target],
				Memory:              s.AIIncidentMemoryRaw[actor][target],
				Government:          originalAIRelationGovernment(s.AIPlayers[actor]),
				ProtectedAgreement:  protected,
				DemocracyMemoryFlag: s.AIIncidentBetrayalRaw[actor][target],
			}, roll100)
			if !ok {
				continue
			}
			s.AIIncidentReasonRaw[actor][target] = out.PendingReason
			s.AIIncidentMagnitudeRaw[actor][target] = out.PendingMagnitude
			s.AIIncidentMemoryRaw[actor][target] = out.Memory
			s.AIIncidentMemoryRaw[target][actor] = out.Memory
		}
	}
}
