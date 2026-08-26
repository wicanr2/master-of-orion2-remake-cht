package shell

import (
	"math/rand"
	"testing"
)

func lethalInitiativeCombatant(id, initiative int) combatant {
	return combatant{
		battleID: id, initiative: initiative, hp: 10, atk: 1000, def: 0,
		wmin: 100, wmax: 100, shots: 1, kind: WeaponKindBeam,
	}
}

func TestQuickCombatRoundShipInitiativeSwitchesGlobalOrder(t *testing.T) {
	player := []combatant{lethalInitiativeCombatant(1, 10)}
	enemy := []combatant{lethalInitiativeCombatant(2, 100)}
	eLost, pLost := resolveQuickCombatRound(&player, &enemy, true, rand.New(rand.NewSource(1)))
	if eLost != 0 || pLost != 1 {
		t.Fatalf("開啟主動權時高分敵艦應先擊沉我艦：敵損=%d 我損=%d", eLost, pLost)
	}

	player = []combatant{lethalInitiativeCombatant(1, 10)}
	enemy = []combatant{lethalInitiativeCombatant(2, 100)}
	eLost, pLost = resolveQuickCombatRound(&player, &enemy, false, rand.New(rand.NewSource(1)))
	if eLost != 1 || pLost != 0 {
		t.Fatalf("關閉主動權時應維持我方整隊先行：敵損=%d 我損=%d", eLost, pLost)
	}
}

func TestExpireTacticalStoredEnergy(t *testing.T) {
	ships := []CombatShip{{StoredEnergy: 17}, {StoredEnergy: 3}}
	ExpireTacticalStoredEnergy(ships)
	for i := range ships {
		if ships[i].StoredEnergy != 0 {
			t.Fatalf("第 %d 艘未清空儲能：%d", i, ships[i].StoredEnergy)
		}
	}
}
