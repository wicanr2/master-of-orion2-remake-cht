package main

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestEnemyPlasmaFluxDamagesBothSidesInRadiusOnce(t *testing.T) {
	ts := &tacticalScreen{
		rng: rand.New(rand.NewSource(17)),
		player: []shell.CombatShip{
			{Name: "近距目標", HP: 100, MaxHP: 100, Col: 1, Row: 0, SizeClass: gamedata.SHIP_FRIGATE},
			{Name: "圈外目標", HP: 100, MaxHP: 100, Col: 5, Row: 0, SizeClass: gamedata.SHIP_TITAN},
		},
		enemy: []shell.CombatShip{
			{Name: "太空鰻", HP: 100, MaxHP: 100, Col: 0, Row: 0,
				WeaponMin: 10, WeaponMax: 10, WeaponMounts: []shell.ShipWeaponMount{{
					Name: "電漿通量", Attack: 10, WorkingCount: 2, MaxCount: 2, Arc: gamedata.ARC_MONSTER_360,
				}}},
			{Name: "同陣營鄰艦", HP: 100, MaxHP: 100, Col: 0, Row: 1, SizeClass: gamedata.SHIP_DESTROYER},
		},
	}
	dealt := ts.enemyRetaliationDamage(0, 0)
	if dealt <= 0 || ts.player[0].HP >= 100 {
		t.Fatalf("近距敵艦應受傷：dealt=%d hp=%d", dealt, ts.player[0].HP)
	}
	if ts.enemy[1].HP >= 100 {
		t.Fatalf("原版沒有 owner 過濾，同陣營鄰艦應受傷：hp=%d", ts.enemy[1].HP)
	}
	if ts.player[1].HP != 100 || ts.enemy[0].HP != 100 {
		t.Fatalf("圈外艦與射手不得受傷：far=%d shooter=%d", ts.player[1].HP, ts.enemy[0].HP)
	}
}

func TestEnemyPlasmaFluxCanDamageBothSidesFighterSquadrons(t *testing.T) {
	observedBothSides := false
	for seed := int64(1); seed <= 64; seed++ {
		ts := &tacticalScreen{
			rng: rand.New(rand.NewSource(seed)),
			player: []shell.CombatShip{
				{Name: "啟動目標", HP: 100, MaxHP: 100, Col: 1, Row: 0},
			},
			enemy: []shell.CombatShip{
				{Name: "太空鰻", HP: 100, MaxHP: 100, Col: 0, Row: 0},
			},
			squads: []shell.FighterSquadron{
				{Kind: shell.FighterInterceptor, Alive: 4, HPEach: 1, MaxHPEach: 1, Col: 1, Row: 0},
				{Kind: shell.FighterHeavy, Enemy: true, Alive: 4, HPEach: 1, MaxHPEach: 1, Col: 0, Row: 1},
				{Kind: shell.FighterBomber, Alive: 4, HPEach: 1, MaxHPEach: 1, Col: 5, Row: 0},
			},
		}
		_, fired := ts.enemyPlasmaFluxShot(0, 100, 100, 2)
		if !fired {
			t.Fatal("Plasma Flux 應成功啟動")
		}
		if ts.squads[2].Alive != 4 {
			t.Fatalf("圈外中隊不應受損：seed=%d alive=%d", seed, ts.squads[2].Alive)
		}
		if ts.squads[0].Alive < 4 && ts.squads[1].Alive < 4 {
			observedBothSides = true
			break
		}
	}
	if !observedBothSides {
		t.Fatal("64 條可重播亂數序列中未觀察到雙方圈內中隊同時未迴避")
	}
}
