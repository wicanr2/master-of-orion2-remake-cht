package diplomacy

import "testing"

func TestRelationLevelForScore(t *testing.T) {
	cases := []struct {
		score int
		want  RelationLevel
	}{
		{-100, RelationFeud}, {-150, RelationFeud}, // 夾限
		{0, RelationNeutral},
		{100, RelationHarmony}, {200, RelationHarmony}, // 夾限
		{50, RelationAffable}, {-50, RelationTense},
	}
	for _, c := range cases {
		if got := RelationLevelForScore(c.score); got != c.want {
			t.Errorf("RelationLevelForScore(%d) = %d,預期 %d", c.score, got, c.want)
		}
	}
}

func TestRelationLevelName(t *testing.T) {
	if RelationFeud.Name() != "FEUD" || RelationNeutral.Name() != "NEUTRAL" || RelationHarmony.Name() != "HARMONY" {
		t.Errorf("等級名稱錯誤:%s/%s/%s", RelationFeud.Name(), RelationNeutral.Name(), RelationHarmony.Name())
	}
}

func TestRelationApplyEvents(t *testing.T) {
	r := &RelationState{Score: 0}
	if r.Apply(EventDeclareWar) != -40 { // 0-40
		t.Errorf("宣戰後分數 = %d,預期 -40", r.Score)
	}
	if !r.Level().IsHostile() {
		t.Error("宣戰後應為敵對(Tense < Wary)")
	}
	r2 := &RelationState{Score: 0}
	r2.Apply(EventSignAlliance) // +30
	if r2.Score != 30 || r2.Level() != RelationAmiable {
		t.Errorf("結盟後 score=%d level=%d,預期 30/Amiable", r2.Score, r2.Level())
	}
}

func TestRelationClampAndDrift(t *testing.T) {
	// 夾限:多次宣戰不低於 -Max
	r := &RelationState{Score: -90}
	r.Apply(EventDeclareWar) // -130 → 夾到 -100
	if r.Score != -RelationScoreMax {
		t.Errorf("夾限後 = %d,預期 %d", r.Score, -RelationScoreMax)
	}
	// 漂移:正往下、負往上、0 不動
	pos := &RelationState{Score: 40}
	pos.AdvanceTurn()
	neg := &RelationState{Score: -40}
	neg.AdvanceTurn()
	zero := &RelationState{Score: 0}
	zero.AdvanceTurn()
	if pos.Score != 39 || neg.Score != -39 || zero.Score != 0 {
		t.Errorf("漂移錯誤:pos=%d neg=%d zero=%d", pos.Score, neg.Score, zero.Score)
	}
}

func TestRelationFriendly(t *testing.T) {
	if !RelationHarmony.IsFriendly() || RelationNeutral.IsFriendly() {
		t.Error("友好判定錯誤")
	}
}
