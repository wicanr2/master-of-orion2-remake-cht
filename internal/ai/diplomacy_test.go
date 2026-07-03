package ai

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/diplomacy"
)

func TestDecideStance(t *testing.T) {
	cases := []struct {
		level diplomacy.RelationLevel
		p     Profile
		want  Stance
	}{
		{diplomacy.RelationFeud, ProfileAggressive, StanceWar},                // 敵對+好戰→開戰
		{diplomacy.RelationFeud, ProfileScientific, StanceHostile},            // 敵對+非好戰→敵視
		{diplomacy.RelationHarmony, ProfileScientific, StanceProposeAlliance}, // 極友好+非好戰→結盟
		{diplomacy.RelationPeaceful, ProfileScientific, StanceProposeTrade},   // 友好但未達UNITY→貿易
		{diplomacy.RelationHarmony, ProfileAggressive, StanceProposeTrade},    // 友好+好戰→貿易(不結盟)
		{diplomacy.RelationNeutral, ProfileScientific, StanceProposeTrade},    // 中立+非好戰→貿易
		{diplomacy.RelationNeutral, ProfileAggressive, StanceNeutral},         // 中立+好戰→中立伺機
	}
	for _, c := range cases {
		if got := DecideStance(c.level, c.p); got != c.want {
			t.Errorf("DecideStance(%s, %s) = %d,預期 %d", c.level.Name(), c.p.Name, got, c.want)
		}
	}
}
