package gamedata

import (
	"reflect"
	"testing"
)

func TestResolveStrategicColonyDamageBuildingOrderAndExclusions(t *testing.T) {
	state := StrategicColonyDamageState{RawBuildingIDs: []int{7, 47, 22, 8, 22, 49, 1}, BuildingHitCost: 1}
	got := ResolveStrategicColonyDamage(state, 3, func(int) int { return 0 })
	if want := []int{22, 7, 1}; !reflect.DeepEqual(got.DestroyedBuildingIDs, want) {
		t.Fatalf("建築候選應依 raw ID 反序並排除防禦建築，got %v want %v", got.DestroyedBuildingIDs, want)
	}
}

func TestResolveStrategicColonyDamageUnitCostInsufficientStops(t *testing.T) {
	state := StrategicColonyDamageState{Marines: 1, MarineHitCost: 2, TankHitCost: 2, BuildingHitCost: 1}
	got := ResolveStrategicColonyDamage(state, 1, func(int) int { return 0 })
	if got.State.Marines != 1 || got.DamageSpent != 0 || got.DamageRemaining != 1 {
		t.Fatalf("抽中成本不足的陸戰隊應立即停止，got %+v", got)
	}
}

func TestResolveStrategicColonyDamageMarinesAndTanks(t *testing.T) {
	picks := []int{0, 0}
	got := ResolveStrategicColonyDamage(StrategicColonyDamageState{
		Marines: 1, Tanks: 1, MarineHitCost: 2, TankHitCost: 3, BuildingHitCost: 1,
	}, 5, func(int) int { p := picks[0]; picks = picks[1:]; return p })
	if got.MarinesLost != 1 || got.TanksLost != 1 || got.State.Marines != 0 || got.State.Tanks != 0 || got.DamageRemaining != 0 {
		t.Fatalf("陸戰隊／戰車成本與寫回錯誤：%+v", got)
	}
}

func TestResolveStrategicColonyDamageBuildProgressTailQuirk(t *testing.T) {
	got := ResolveStrategicColonyDamage(StrategicColonyDamageState{
		Population: 3, BuildProgress: 40, MarineHitCost: 1, TankHitCost: 2, BuildingHitCost: 1,
	}, 1, func(int) int { return 2 })
	if got.BuildProgressLost != 40 || got.State.BuildProgress != 0 || got.PopulationLost != 0 {
		t.Fatalf("tail index 小於完整建造進度時應一次清空進度：%+v", got)
	}
}

func TestResolveStrategicColonyDamageLastPopulation(t *testing.T) {
	got := ResolveStrategicColonyDamage(StrategicColonyDamageState{
		Population: 1, LastPopulationPoints: 100, MarineHitCost: 1, TankHitCost: 2, BuildingHitCost: 1,
	}, 2, nil)
	if got.State.Population != 0 || got.PopulationLost != 1 || !got.ColonyDestroyed || got.DamageSpent != 1 || got.DamageRemaining != 1 {
		t.Fatalf("最後人口應以一點傷害扣 100 人口點並留下多餘傷害：%+v", got)
	}
}
