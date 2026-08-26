package shell

import "testing"

func TestResolveBattleStoresTypedRoundLog(t *testing.T) {
	s := NewDemoSession()
	result := s.ResolveBattle(s.PrimaryEnemyName())
	for i, round := range result.Log {
		if round.Round != i+1 || round.EnemyDestroyed < 0 || round.PlayerDestroyed < 0 {
			t.Fatalf("round %d invalid typed result: %+v", i, round)
		}
	}
}

func TestAntaranBattleEnemyUsesTypedKind(t *testing.T) {
	result := BattleResult{EnemyKind: BattleEnemyAntaran}
	if result.Enemy != "" || result.EnemyKind != BattleEnemyAntaran {
		t.Fatalf("special enemy must not require embedded display text: %+v", result)
	}
}
