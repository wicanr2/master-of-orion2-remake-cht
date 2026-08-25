package main

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestTacticalMissileAmmoSpentAndDepleted(t *testing.T) {
	ts := &tacticalScreen{
		b: &sceneBuilder{},
		player: []shell.CombatShip{{
			Name: "飛彈艦", HP: 40, MaxHP: 40, Col: 1, Row: 1, Facing: 0,
			Attack: 100, WeaponMin: 8, WeaponMax: 8, Kind: shell.WeaponKindMissile,
			WeaponName: "核飛彈", WeaponArc: gamedata.ARC_360, WeaponAmmo: 1,
		}},
		enemy: []shell.CombatShip{{Name: "敵艦", HP: 200, MaxHP: 200, Col: 2, Row: 1}},
		rng:   rand.New(rand.NewSource(1)),
	}
	beforeHP := ts.enemy[0].HP
	ts.fireSelectedShip(0)
	if ts.player[0].WeaponAmmo != 0 || ts.enemy[0].HP >= beforeHP {
		t.Fatalf("第一發應造成傷害並把 Ammo 扣到 0: hp=%d→%d ammo=%d",
			beforeHP, ts.enemy[0].HP, ts.player[0].WeaponAmmo)
	}
	ts.acted = nil
	ts.player[0].Fired = false
	ts.fireSelectedShip(0)
	if ts.player[0].Fired || (!strings.Contains(ts.log, "彈藥耗盡") && !strings.Contains(ts.log, "ammunition depleted")) {
		t.Fatalf("耗盡時不應標 Fired，且應提示玩家: fired=%v log=%q", ts.player[0].Fired, ts.log)
	}
}

func TestEnemyMissileTriggersOffPointDefenseInSecondSlot(t *testing.T) {
	ts := &tacticalScreen{
		b: &sceneBuilder{},
		player: []shell.CombatShip{{
			Name: "防空艦", HP: 400, MaxHP: 400, Col: 1, Row: 1,
			Attack: 100, WeaponName: "雷射", WeaponMax: 8,
			WeaponMounts: []shell.ShipWeaponMount{
				{Name: "雷射", WorkingCount: 1, Attack: 8},
				{Name: "雷射", WorkingCount: 1, Attack: 30,
					Mods: []string{string(gamedata.ModPointDefense)}},
			},
			WeaponModes: []shell.TacticalWeaponMode{
				shell.TacticalWeaponReady, shell.TacticalWeaponOff,
			},
		}},
		enemy: []shell.CombatShip{{
			Name: "飛彈敵艦", HP: 100, MaxHP: 100, Col: 2, Row: 1, Facing: 8,
			Attack: 100, WeaponMin: 20, WeaponMax: 20,
			Kind: shell.WeaponKindMissile, WeaponName: "核飛彈",
			WeaponArc: gamedata.ARC_360, WeaponAmmo: 1, DriveLevel: 1,
		}},
		rng: rand.New(rand.NewSource(7)),
	}
	ts.enemyRetaliationDamage(0, 0)
	if ts.enemy[0].WeaponAmmo != 0 {
		t.Fatalf("敵方飛彈還擊應消耗一發彈藥：%d", ts.enemy[0].WeaponAmmo)
	}
	if len(ts.player[0].PointDefenseSpentSlots) < 2 || !ts.player[0].PointDefenseSpentSlots[1] {
		t.Fatalf("第二槽紅色 PD 應在來襲飛彈命中前開火：%v", ts.player[0].PointDefenseSpentSlots)
	}
	if ts.player[0].WeaponModes[1] != shell.TacticalWeaponOff {
		t.Fatalf("自動攔截後紅色槽仍須保持關閉：%v", ts.player[0].WeaponModes)
	}
}
