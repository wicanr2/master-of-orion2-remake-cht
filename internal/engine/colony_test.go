package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestRunColonyTurn(t *testing.T) {
	cs := ColonyState{
		Population: 10, PopMax: 20,
		Farmers: 4, Workers: 4, Scientists: 2,
		FoodPerFarmer: 3, IndustryPerWorker: 5, ResearchPerScientist: 4,
		PlanetSize: gamedata.MEDIUM_PLANET, // 容忍 = 2*(2+1) = 6
	}
	got := RunColonyTurn(cs)

	// 食物:4*3=12,消耗 10,盈餘 2
	if got.Food != 12 || got.FoodConsumed != 10 || got.FoodSurplus != 2 || got.Starving {
		t.Errorf("食物錯誤:%+v", got)
	}
	// 工業:4*5=20;污染:eighths=8,產污=20,清理=(20-6)/2=7,淨=13
	if got.GrossIndustry != 20 || got.PollutingProduction != 20 ||
		got.PollutionCleanupCost != 7 || got.NetIndustry != 13 {
		t.Errorf("工業/污染錯誤:%+v", got)
	}
	// 研究:2*4=8
	if got.Research != 8 {
		t.Errorf("研究 = %d,預期 8", got.Research)
	}
	// 成長:base=sqrt(2000*10*10/20)=sqrt(10000)=100,bonus=0 → 100
	if got.PopGrowth != 100 {
		t.Errorf("成長 = %d,預期 100", got.PopGrowth)
	}
}

func TestRunColonyTurnPollutionProcessor(t *testing.T) {
	// 污染處理器:eighths 8→4,產污減半 → 清理成本降。
	cs := ColonyState{
		Population: 10, PopMax: 20, Workers: 4, IndustryPerWorker: 5,
		PlanetSize: gamedata.MEDIUM_PLANET, PollutionProcessor: true,
	}
	got := RunColonyTurn(cs)
	// gross=20,eighths=4,產污=20*4/8=10,清理=(10-6)/2=2,淨=18
	if got.PollutingProduction != 10 || got.PollutionCleanupCost != 2 || got.NetIndustry != 18 {
		t.Errorf("處理器污染錯誤:%+v", got)
	}
}

func TestRunColonyTurnStarving(t *testing.T) {
	// 食物不足 → 饑荒,成長為 0 並標 Starving。
	cs := ColonyState{
		Population: 10, PopMax: 20, Farmers: 1, FoodPerFarmer: 3,
		PlanetSize: gamedata.MEDIUM_PLANET,
	}
	got := RunColonyTurn(cs)
	if got.FoodSurplus != -7 || !got.Starving || got.PopGrowth != 0 {
		t.Errorf("饑荒處理錯誤:%+v", got)
	}
}

func TestRunColonyTurnTolerantRace(t *testing.T) {
	// Tolerant 種族:清理成本 0,淨工業 = 毛工業。
	cs := ColonyState{
		Population: 5, PopMax: 20, Workers: 6, IndustryPerWorker: 5,
		PlanetSize: gamedata.TINY_PLANET, TolerantRace: true,
	}
	got := RunColonyTurn(cs)
	if got.PollutionCleanupCost != 0 || got.NetIndustry != 30 {
		t.Errorf("Tolerant 污染錯誤:%+v", got)
	}
}

func TestRunColonyTurnMorale(t *testing.T) {
	// 士氣 +30%(3 格笑臉)→ 食物/工業/研究皆 ×1.3。
	cs := ColonyState{
		Population: 10, PopMax: 20, Farmers: 4, Workers: 2, Scientists: 2,
		FoodPerFarmer: 5, IndustryPerWorker: 10, ResearchPerScientist: 5,
		PlanetSize: gamedata.TINY_PLANET, TolerantRace: true, MoralePercent: 30,
	}
	got := RunColonyTurn(cs)
	if got.Food != 26 { // 20*130/100
		t.Errorf("士氣調整食物 = %d,預期 26", got.Food)
	}
	if got.GrossIndustry != 26 { // 20*1.3
		t.Errorf("士氣調整工業 = %d,預期 26", got.GrossIndustry)
	}
	if got.Research != 13 { // 10*1.3
		t.Errorf("士氣調整研究 = %d,預期 13", got.Research)
	}
}
