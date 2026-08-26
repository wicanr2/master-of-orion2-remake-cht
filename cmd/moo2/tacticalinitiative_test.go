package main

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func initiativeTestScreen(player, enemy []shell.CombatShip) *tacticalScreen {
	session := shell.NewDemoSession()
	settings := session.EffectiveGameSettings()
	settings.ShipInitiative = true
	session.ApplyGameSettings(settings)
	t := &tacticalScreen{
		b:        &sceneBuilder{session: session},
		player:   player,
		enemy:    enemy,
		rng:      rand.New(rand.NewSource(1)),
		acted:    make([]bool, len(player)),
		waited:   make([]bool, len(player)),
		moveLeft: freshMoveBudgets(player),
	}
	t.resetInitiativeQueue()
	t.advanceInitiativeQueue()
	return t
}

func initiativePlayer(name string, initiative, hp int) shell.CombatShip {
	return shell.CombatShip{
		Name: name, Attack: initiative, HP: hp, MaxHP: hp,
		Col: 1, Row: 1, Defense: 0, CombatSpeed: 1,
	}
}

func initiativeEnemy(name string, initiative, damage int) shell.CombatShip {
	attack := initiative
	if attack < 1000 {
		attack = 1000
	}
	return shell.CombatShip{
		Name: name, Attack: attack, HP: 100, MaxHP: 100,
		Col: 2, Row: 1, CombatSpeed: 1, WeaponName: "雷射", WeaponArc: gamedata.ARC_360,
		WeaponMin: damage, WeaponMax: damage,
	}
}

func TestTacticalInitiativeEnemyActsBeforeSlowerPlayer(t *testing.T) {
	screen := initiativeTestScreen(
		[]shell.CombatShip{initiativePlayer("慢艦", 10, 100)},
		[]shell.CombatShip{initiativeEnemy("快敵", 100, 10)},
	)
	if screen.player[0].HP >= 100 {
		t.Fatal("高主動權敵艦應在玩家取得操作權前開火")
	}
	if screen.sel != 0 || screen.round != 0 {
		t.Fatalf("敵艦行動後應把同回合操作權交給玩家：sel=%d round=%d", screen.sel, screen.round)
	}
}

func TestTacticalInitiativePlayerActsBeforeSlowerEnemy(t *testing.T) {
	screen := initiativeTestScreen(
		[]shell.CombatShip{initiativePlayer("快艦", 2000, 100)},
		[]shell.CombatShip{initiativeEnemy("慢敵", 10, 10)},
	)
	if screen.player[0].HP != 100 || screen.sel != 0 {
		t.Fatalf("高主動權玩家艦應先取得操作權：hp=%d sel=%d", screen.player[0].HP, screen.sel)
	}
}

func TestTacticalInitiativeWaitMovesPlayerBehindEnemy(t *testing.T) {
	screen := initiativeTestScreen(
		[]shell.CombatShip{initiativePlayer("快艦", 2000, 100)},
		[]shell.CombatShip{initiativeEnemy("慢敵", 50, 10)},
	)
	screen.waitSelectedAction()
	if screen.player[0].HP >= 100 {
		t.Fatal("WAIT 應讓排在後方的敵艦先行動")
	}
	if screen.sel != 0 || screen.acted[0] || !screen.waited[0] || screen.round != 0 {
		t.Fatalf("WAIT 後原艦應留在本回合尾端：sel=%d acted=%v waited=%v round=%d",
			screen.sel, screen.acted, screen.waited, screen.round)
	}
}

func TestTacticalInitiativeKilledShipLosesPendingAction(t *testing.T) {
	screen := initiativeTestScreen(
		[]shell.CombatShip{initiativePlayer("脆弱艦", 10, 1)},
		[]shell.CombatShip{initiativeEnemy("致命敵艦", 100, 100)},
	)
	if len(screen.player) != 0 || !screen.over || screen.won {
		t.Fatalf("先被擊沉的玩家艦不可再取得行動：players=%d over=%v won=%v",
			len(screen.player), screen.over, screen.won)
	}
}

func TestTacticalInitiativeCasualtyCompactionKeepsShipIdentity(t *testing.T) {
	screen := initiativeTestScreen(
		[]shell.CombatShip{
			initiativePlayer("先陣亡", 5, 1),
			initiativePlayer("仍待行動", 10, 100),
		},
		[]shell.CombatShip{initiativeEnemy("先手敵艦", 100, 100)},
	)
	if len(screen.player) != 1 || screen.player[0].Name != "仍待行動" {
		t.Fatalf("戰損壓縮應保留正確倖存艦：%+v", screen.player)
	}
	if screen.sel != 0 || screen.currentInitiativePlayerIndex() != 0 {
		t.Fatalf("壓縮後待行動項應重新定位到倖存艦：sel=%d current=%d",
			screen.sel, screen.currentInitiativePlayerIndex())
	}
}
