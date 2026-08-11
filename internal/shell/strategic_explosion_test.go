package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestStrategicShipExplosionConsumesOracleAndWritesCollateralDamage(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = []Fleet{{Ships: []Ship{
		{Name: "主爆炸艦", Class: "巡防艦"},
		{Name: "受波及艦", Class: "巡洋艦"},
		{Name: "第三艦", Class: "戰艦"},
	}}}
	s.SelectedFleet = 0
	s.EventSeed = 19
	s.eventRand = newRandStream(19)

	result, ok := s.resolveStrategicShipExplosion()
	if !ok {
		t.Fatal("至少兩艘艦時應能結算戰略艦船爆炸")
	}
	if result.Potential < gamedata.OriginalShipExplosionRollOffset ||
		result.Potential >= gamedata.OriginalShipExplosionRollOffset+gamedata.OriginalShipExplosionRollRange {
		t.Fatalf("主爆炸勢能未落在 oracle 範圍:%d", result.Potential)
	}
	if s.ShipCount() != 2 || result.Lost.Name == "" {
		t.Fatalf("主艦應只移除一艘,ships=%d lost=%+v", s.ShipCount(), result.Lost)
	}
	if len(result.Collateral) == 0 {
		t.Fatal("爆炸連鎖應至少把一次下游 damage consumer 寫入倖存艦")
	}
	for _, collateral := range result.Collateral {
		if collateral.RawDamage <= 0 || collateral.AppliedDamage <= 0 {
			t.Fatalf("連鎖消費值不合理:%+v", collateral)
		}
	}
}

func TestStrategicShipExplosionKeepsLastShipSafe(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = []Fleet{{Ships: []Ship{{Name: "唯一艦", Class: "巡防艦"}}}}
	s.eventRand = newRandStream(1)
	if _, ok := s.resolveStrategicShipExplosion(); ok {
		t.Fatal("只剩一艘艦時不應觸發爆炸事件")
	}
	if s.ShipCount() != 1 {
		t.Fatalf("安全護欄不應改變艦數,got %d", s.ShipCount())
	}
}
