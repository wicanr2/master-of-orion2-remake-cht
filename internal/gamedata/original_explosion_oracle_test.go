package gamedata

import "testing"

func TestOriginalExplosionRollBoundsAndOffsets(t *testing.T) {
	if got := OriginalShipExplosionDamageRoll(-1); got != 74 {
		t.Fatalf("ship roll 下界=%d,want 74", got)
	}
	if got := OriginalShipExplosionDamageRoll(200); got != 274 {
		t.Fatalf("ship roll 上界=%d,want 274", got)
	}
	if got := OriginalColonyExplosionPotentialRoll(400); got != 549 {
		t.Fatalf("colony roll 上界=%d,want 549", got)
	}
}

func TestOriginalExplosionChainAndEngineFormula(t *testing.T) {
	if got := OriginalExplosionChainNextPotential(20); got != 0 {
		t.Fatalf("20 potential 走一步應歸零,got %d", got)
	}
	if got := OriginalExplosionChainNextPotential(61); got != 41 {
		t.Fatalf("61 potential 走一步應剩 41,got %d", got)
	}
	table := []int{10, 20, 30, 40, 50, 60}
	if got := OriginalEngineExplosionBasePotential(2, 3, table); got != 600 {
		t.Fatalf("size<5 engine factor 應為 +1,got %d", got)
	}
	if got := OriginalEngineExplosionBasePotential(5, 3, table); got != 1500 {
		t.Fatalf("size>=5 engine factor 應為 +2,got %d", got)
	}
}

func TestOriginalExplosionDamageConsumer(t *testing.T) {
	if got := OriginalExplosionDamageConsumer(100, 7, 999); got != 25 {
		t.Fatalf("target type 7 應取四分之一,got %d", got)
	}
	if got := OriginalExplosionDamageConsumer(100, 2, 30); got != 70 {
		t.Fatalf("一般船型應扣 resistance,got %d", got)
	}
	if got := OriginalExplosionDamageConsumer(30, 2, 30); got != 0 {
		t.Fatalf("扣除 resistance 後不可為負,got %d", got)
	}
}
