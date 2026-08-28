package engine

import "testing"

func TestOriginalFoodTransportBalance(t *testing.T) {
	surplus := ColonyState{Population: 1, Farmers: 1, FoodPerFarmer: 5}
	deficit := ColonyState{Population: 3}
	short, ok := OriginalFoodTransport(PlayerState{ActiveFreighters: 1}, []ColonyState{surplus, deficit}, nil)
	if !ok || short.FoodFreighted != 1 || short.SurplusFreighters != -1 {
		t.Fatalf("short=%+v ok=%v", short, ok)
	}
	plenty, ok := OriginalFoodTransport(PlayerState{ActiveFreighters: 5}, []ColonyState{surplus, deficit}, nil)
	if !ok || plenty.FoodFreighted != 2 || plenty.SurplusFreighters != 3 {
		t.Fatalf("plenty=%+v ok=%v", plenty, ok)
	}
}

func TestOriginalFoodTransportSettlersConsumeFiveFreighters(t *testing.T) {
	got, ok := OriginalFoodTransport(PlayerState{ActiveFreighters: 5, SettlersFreighted: 1},
		[]ColonyState{{Population: 1, Farmers: 1, FoodPerFarmer: 5}, {Population: 3}}, nil)
	if !ok || got.FoodFreighted != 0 || got.SurplusFreighters >= 0 {
		t.Fatalf("settler transport=%+v ok=%v", got, ok)
	}
}
