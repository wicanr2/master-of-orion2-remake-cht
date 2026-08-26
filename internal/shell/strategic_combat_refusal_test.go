package shell

import "testing"

func TestStrategicCombatEntrypointsReturnTypedRefusals(t *testing.T) {
	s := NewDemoSession()
	if got := s.BombardColony(-1).Reason; got != BombardInvalidStar {
		t.Fatalf("轟炸非法星系原因碼=%q", got)
	}
	if got := s.InvadeColony(-1).Reason; got != GroundInvalidStar {
		t.Fatalf("入侵非法星系原因碼=%q", got)
	}
	if got := s.MindControlColony(-1).Reason; got != GroundRequiresTelepathy {
		t.Fatalf("非心靈感應種族應先被 trait gate 擋下，原因碼=%q", got)
	}
	_, _, monsterCode := s.StartMonsterCombat(-1)
	if monsterCode != MonsterCombatNoMonster {
		t.Fatalf("無怪獸星系原因碼=%q", monsterCode)
	}
	if got := s.AttackMonster(-1).Reason; got != MonsterCombatNoMonster {
		t.Fatalf("快速怪獸入口原因碼=%q", got)
	}
}
