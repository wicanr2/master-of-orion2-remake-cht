package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestQuickCombatConsumesEveryWorkingWeaponMount(t *testing.T) {
	defOne := []combatant{{hp: 1000, maxHP: 1000}}
	one := []combatant{{shots: 1, weaponMounts: []ShipWeaponMount{
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360},
	}}}
	battleVolley(one, &defOne, rand.New(rand.NewSource(7)))
	oneDamage := 1000 - defOne[0].hp

	defTwo := []combatant{{hp: 1000, maxHP: 1000}}
	two := []combatant{{shots: 1, weaponMounts: []ShipWeaponMount{
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360},
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 1, Attack: 40, Arc: gamedata.ARC_360},
		{Name: "脈衝星", MaxCount: 1, WorkingCount: 0, Attack: 400, Arc: gamedata.ARC_360},
	}}}
	battleVolley(two, &defTwo, rand.New(rand.NewSource(7)))
	twoDamage := 1000 - defTwo[0].hp
	if oneDamage <= 0 || twoDamage <= oneDamage {
		t.Fatalf("第二個有效槽應增加傷害、零工作數槽不得取代它：one=%d two=%d", oneDamage, twoDamage)
	}
}

func TestQuickCombatTracksMissileAmmoPerMount(t *testing.T) {
	attackers := []combatant{{shots: 1, weaponMounts: []ShipWeaponMount{
		{Name: "核飛彈", MaxCount: 1, WorkingCount: 1, Ammo: 2, Attack: 20, Arc: gamedata.ARC_360},
		{Name: "麥克萊特飛彈", MaxCount: 1, WorkingCount: 1, Ammo: 2, Attack: 30, Arc: gamedata.ARC_360},
	}}}
	defenders := []combatant{{hp: 1000, maxHP: 1000}}
	battleVolley(attackers, &defenders, rand.New(rand.NewSource(11)))
	for i, mount := range attackers[0].weaponMounts {
		if mount.Ammo != 1 {
			t.Fatalf("飛彈槽 %d 應各自消耗一發，got %d", i, mount.Ammo)
		}
	}
}

func TestStartCombatPreservesPlayerWeaponMounts(t *testing.T) {
	s := NewDemoSession()
	// Ships[0] 是不進戰術戰場的殖民船；逐槽護欄必須放在第一艘真正戰鬥艦。
	s.Fleet().Ships[1].WeaponMounts = []ShipWeaponMount{
		{Name: "雷射", MaxCount: 2, WorkingCount: 2, Attack: 5, Arc: gamedata.ARC_FWD, Ammo: 255},
		{Name: "核飛彈", MaxCount: 1, WorkingCount: 1, Attack: 8, Arc: gamedata.ARC_360, Ammo: 5},
	}
	player, _ := s.StartCombat(s.PrimaryEnemyName())
	if len(player) == 0 || len(player[0].WeaponMounts) != 2 || player[0].WeaponMounts[1].Name != "核飛彈" {
		t.Fatalf("StartCombat 應保存玩家逐槽武器：%+v", player)
	}
	s.Fleet().Ships[1].WeaponMounts[1].Name = "已修改"
	if player[0].WeaponMounts[1].Name != "核飛彈" {
		t.Fatal("CombatShip 的逐槽資料必須與持久 Ship 深複製隔離")
	}
}

func TestQuickCombatUsesPointDefenseOutsideFirstSlot(t *testing.T) {
	attackers := []combatant{{
		hp: 100, maxHP: 100, shots: 1, kind: WeaponKindMissile,
		weaponName: "核飛彈", wmin: 20, wmax: 20, ammo: 5, ammoSet: true,
		missileFTLLevel: 1,
	}}
	defenders := []combatant{{
		hp: 1000, maxHP: 1000, atk: 100, wmax: 20,
		weaponName: "雷射",
		weaponMounts: []ShipWeaponMount{
			{Name: "雷射", WorkingCount: 1, Attack: 20},
			{Name: "雷射", WorkingCount: 1, Attack: 20,
				Mods: []string{string(gamedata.ModPointDefense)}},
		},
	}}
	battleVolley(attackers, &defenders, rand.New(rand.NewSource(3)))
	if len(defenders[0].pointDefenseSpentSlots) < 2 || !defenders[0].pointDefenseSpentSlots[1] {
		t.Fatalf("快速結算應消費第二槽 PD：%v", defenders[0].pointDefenseSpentSlots)
	}
	if defenders[0].pointDefenseSpentSlots[0] {
		t.Fatalf("沒有 PD 改造的第一槽不得標成已使用：%v", defenders[0].pointDefenseSpentSlots)
	}
}
