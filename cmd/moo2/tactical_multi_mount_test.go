package main

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func multiMountTacticalScreen(mounts []shell.ShipWeaponMount) *tacticalScreen {
	return &tacticalScreen{
		b: &sceneBuilder{},
		player: []shell.CombatShip{{
			Name: "多槽艦", HP: 80, MaxHP: 80, Col: 1, Row: 1, Facing: 0,
			Attack: 100, WeaponMin: 20, WeaponMax: 40, Kind: shell.WeaponKindSpherical,
			WeaponName: "脈衝星", WeaponArc: gamedata.ARC_360, WeaponAmmo: 255,
			WeaponMounts: mounts, Charged: true,
		}},
		enemy: []shell.CombatShip{{Name: "目標", HP: 1000, MaxHP: 1000, Col: 2, Row: 1}},
		rng:   rand.New(rand.NewSource(19)),
	}
}

func TestTacticalMultiMountConsumesEveryWorkingSlot(t *testing.T) {
	one := multiMountTacticalScreen([]shell.ShipWeaponMount{
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360, Ammo: 255},
	})
	one.fireSelectedShip(0)
	oneDamage := 1000 - one.enemy[0].HP

	two := multiMountTacticalScreen([]shell.ShipWeaponMount{
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360, Ammo: 255},
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360, Ammo: 255},
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 0, Attack: 400, Arc: gamedata.ARC_360, Ammo: 255},
	})
	two.fireSelectedShip(0)
	twoDamage := 1000 - two.enemy[0].HP
	if oneDamage <= 0 || twoDamage <= oneDamage {
		t.Fatalf("格子戰術第二個有效槽應增加傷害，零工作數槽不得開火：one=%d two=%d", oneDamage, twoDamage)
	}
}

func TestTacticalMultiMountKeepsAmmoAndArcPerSlot(t *testing.T) {
	ts := &tacticalScreen{
		b: &sceneBuilder{},
		player: []shell.CombatShip{{
			Name: "混裝艦", HP: 80, MaxHP: 80, Col: 1, Row: 1, Facing: 0,
			Attack: 100, WeaponMin: 8, WeaponMax: 8, Kind: shell.WeaponKindMissile,
			WeaponName: "核飛彈", WeaponArc: gamedata.ARC_FWD, WeaponAmmo: 2,
			WeaponMounts: []shell.ShipWeaponMount{
				{Name: "雷射", MaxCount: 1, WorkingCount: 1, Attack: 8, Arc: gamedata.ARC_FWD, Ammo: 255},
				{Name: "麥克萊特飛彈", MaxCount: 1, WorkingCount: 1, Attack: 14, Arc: gamedata.ARC_360, Ammo: 2},
			},
		}},
		// 目標在艦艏反方向：第一槽 FWD 被擋，第二槽 360 仍應射擊。
		enemy: []shell.CombatShip{{Name: "目標", HP: 500, MaxHP: 500, Col: 0, Row: 1}},
		rng:   rand.New(rand.NewSource(23)),
	}
	ts.fireRoundForActors(0, []int{0}, false)
	if ts.player[0].WeaponMounts[0].Ammo != 255 {
		t.Fatalf("被射界擋住的光束槽應維持無限彈藥 sentinel，got %d", ts.player[0].WeaponMounts[0].Ammo)
	}
	if ts.player[0].WeaponMounts[1].Ammo != 1 {
		t.Fatalf("360 度第二槽應獨立扣一發，got %d", ts.player[0].WeaponMounts[1].Ammo)
	}
	if !ts.player[0].Fired || ts.enemy[0].HP >= 500 {
		t.Fatalf("第二槽應正常開火並造成傷害：fired=%v hp=%d", ts.player[0].Fired, ts.enemy[0].HP)
	}
}

func TestTacticalMultiMountDoesNotResetShipLevelState(t *testing.T) {
	ts := multiMountTacticalScreen([]shell.ShipWeaponMount{
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360, Ammo: 255},
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360, Ammo: 255},
	})
	ts.player[0].Cloaked = true
	ts.player[0].CloakKind = shell.CloakDevice
	ts.player[0].StoredEnergy = 20
	ts.fireRoundForActors(0, []int{0}, false)
	if ts.player[0].Cloaked || ts.player[0].StoredEnergy != 0 || !ts.player[0].Fired {
		t.Fatalf("逐槽派送不得還原匿蹤／儲能／Fired：cloaked=%v stored=%d fired=%v",
			ts.player[0].Cloaked, ts.player[0].StoredEnergy, ts.player[0].Fired)
	}
}

func TestTacticalWeaponStandbySkipsOnceThenReturnsReady(t *testing.T) {
	ts := multiMountTacticalScreen([]shell.ShipWeaponMount{
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 80, Arc: gamedata.ARC_360, Ammo: 255},
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 20, Arc: gamedata.ARC_360, Ammo: 255},
	})
	ts.player[0].WeaponModes = []shell.TacticalWeaponMode{shell.TacticalWeaponStandby, shell.TacticalWeaponReady}
	if !ts.fireRoundForActors(0, []int{0}, false) {
		t.Fatal("仍有 ready 槽時應可完成齊射")
	}
	if ts.player[0].WeaponModes[0] != shell.TacticalWeaponReady {
		t.Fatalf("standby 槽應在同艦成功齊射後恢復 ready：%v", ts.player[0].WeaponModes)
	}
	firstDamage := 1000 - ts.enemy[0].HP
	if firstDamage <= 0 || firstDamage > 20 {
		t.Fatalf("第一次只能由第二個低傷害槽射擊，damage=%d", firstDamage)
	}
}

func TestTacticalWeaponOffPersistsAndAllDisabledDoesNotAct(t *testing.T) {
	ts := multiMountTacticalScreen([]shell.ShipWeaponMount{
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360, Ammo: 255},
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360, Ammo: 255},
	})
	ts.player[0].WeaponModes = []shell.TacticalWeaponMode{shell.TacticalWeaponOff, shell.TacticalWeaponStandby}
	before := ts.enemy[0].HP
	if ts.fireRoundForActors(0, []int{0}, false) {
		t.Fatal("全關閉／待命時不得假裝完成射擊")
	}
	if ts.enemy[0].HP != before || ts.player[0].WeaponModes[0] != shell.TacticalWeaponOff ||
		ts.player[0].WeaponModes[1] != shell.TacticalWeaponStandby {
		t.Fatalf("失敗齊射不得傷敵或改模式：hp=%d modes=%v", ts.enemy[0].HP, ts.player[0].WeaponModes)
	}
}

func TestTacticalWeaponRowCyclesWithoutEndingShipAction(t *testing.T) {
	ts := multiMountTacticalScreen([]shell.ShipWeaponMount{
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360, Ammo: 255},
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360, Ammo: 255},
	})
	ts.b.lang = i18n.Traditional
	ts.sel = 0
	r := tacticalWeaponSlotRect(1)
	ts.update(shell.InputState{MouseX: r[0] + r[2]/2, MouseY: r[1] + r[3]/2, ClickReleased: true})
	if ts.player[0].WeaponModes[1] != shell.TacticalWeaponStandby || (len(ts.acted) > 0 && ts.acted[0]) {
		t.Fatalf("點槽應只切模式、不結束行動：modes=%v acted=%v", ts.player[0].WeaponModes, ts.acted)
	}
}

func TestTacticalWeaponRightClickDescribesSlotWithoutChangingMode(t *testing.T) {
	ts := multiMountTacticalScreen([]shell.ShipWeaponMount{
		{Name: "脈衝星", MaxCount: 3, WorkingCount: 2, Attack: 40, Arc: gamedata.ARC_360, Ammo: 4,
			Mods: []string{"連射"}},
	})
	ts.b.lang = i18n.Traditional
	ts.sel = 0
	r := tacticalWeaponSlotRect(0)
	ts.update(shell.InputState{
		MouseX:             r[0] + r[2]/2,
		MouseY:             r[1] + r[3]/2,
		RightClickReleased: true,
	})
	if ts.player[0].WeaponModes[0] != shell.TacticalWeaponReady {
		t.Fatalf("右鍵資訊不得改變開火模式：%v", ts.player[0].WeaponModes)
	}
	for _, want := range []string{"傷害上限40", "\u5f48\u85e54", "連射"} {
		if !strings.Contains(ts.log, want) {
			t.Fatalf("右鍵應顯示有界的武器明細，缺 %q：log=%q", want, ts.log)
		}
	}
}
