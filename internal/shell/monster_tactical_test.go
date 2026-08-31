package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func monsterTacticalSession(t *testing.T, kind gamedata.SpaceMonster) (*GameSession, int) {
	t.Helper()
	s, star := newMonsterTestSession(t, kind)
	st, _ := gamedata.MonsterStatsFor(kind)
	s.Monsters = []MonsterGuard{{StarIndex: star, Kind: kind, Structure: st.Structure, Armor: st.Armor}}
	return s, star
}

func TestStartMonsterCombatUsesExactBlueprint(t *testing.T) {
	s, star := monsterTacticalSession(t, gamedata.MonsterDragon)
	player, monsters, reason := s.StartMonsterCombat(star)
	if reason != "" || len(player) == 0 || len(monsters) != 1 {
		t.Fatalf("建立怪物戰術失敗：reason=%q player=%d monsters=%d", reason, len(player), len(monsters))
	}
	m := monsters[0]
	if m.HP != 80 || m.ArmorHP != 2500 || m.CombatSpeed != 18 || m.SizeClass != gamedata.SHIP_TITAN || m.SpriteIdx != 10 {
		t.Errorf("巨龍戰術欄位錯誤：%+v", m)
	}
	if len(m.WeaponMounts) != 2 || m.WeaponMounts[0].WorkingCount != 20 ||
		m.WeaponMounts[0].RawMods != 4 || m.WeaponMounts[1].RawMods != 0x4000 ||
		m.WeaponMounts[1].Name != "巨龍吐息" {
		t.Errorf("巨龍逐槽武器錯誤：%+v", m.WeaponMounts)
	}
	if len(m.WeaponMounts) >= 2 && !containsString(m.WeaponMounts[1].Mods, string(gamedata.ModOverloadedTorpedo)) {
		t.Errorf("巨龍吐息 typed mods 未保留 OVR：%v", m.WeaponMounts[1].Mods)
	}
	if NormalizeWeaponArc(m.WeaponMounts[0].Name, m.WeaponMounts[0].Arc) != gamedata.ARC_MONSTER_360 {
		t.Error("怪物 raw arc 15 不得被一般玩家設計正規化成前向")
	}
}

func TestStartMonsterCombatExcludesSupportShips(t *testing.T) {
	s, star := monsterTacticalSession(t, gamedata.MonsterAmoeba)
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{Name: "非戰鬥殖民船", Class: "殖民船"})
	player, _, reason := s.StartMonsterCombat(star)
	if reason != "" {
		t.Fatal(reason)
	}
	for _, ship := range player {
		if ship.Name == "非戰鬥殖民船" {
			t.Fatal("支援艦不得被投影成格子戰術戰鬥者")
		}
	}
}

func TestStartMonsterCombatIncludesConfirmedColonyStarBase(t *testing.T) {
	s, star := monsterTacticalSession(t, gamedata.MonsterAmoeba)
	s.PlayerColonyStars = []int{star}
	s.ColonyBuildings = []map[string]bool{{"星基": true}}
	player, _, reason := s.StartMonsterCombat(star)
	if reason != "" {
		t.Fatal(reason)
	}
	base := player[len(player)-1]
	if !base.OrbitalBase || base.SpriteIdx%45 != 40 || base.HP != 80 || base.ArmorHP != 80 || len(base.WeaponMounts) != 4 {
		t.Fatalf("Star Base 首幀契約錯誤：%+v", base)
	}
	counts := []int{3, 3, 6, 3}
	for i, count := range counts {
		if base.WeaponMounts[i].WorkingCount != count {
			t.Fatalf("Star Base 槽 %d 數量=%d，want %d", i, base.WeaponMounts[i].WorkingCount, count)
		}
	}
	if base.WeaponMounts[3].Ammo != 20 {
		t.Fatalf("Star Base 核飛彈 ammo=%d，want 20", base.WeaponMounts[3].Ammo)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestMonsterTacticalOutcomeWritesBackAndReplays(t *testing.T) {
	makeSession := func() (*GameSession, int) {
		s, star := monsterTacticalSession(t, gamedata.MonsterAmoeba)
		return s, star
	}
	direct, star := makeSession()
	player, monsters, reason := direct.StartMonsterCombat(star)
	if reason != "" {
		t.Fatal(reason)
	}
	survivors := map[string]bool{}
	for _, sh := range player {
		survivors[sh.Name] = true
	}
	monsters[0].ArmorHP, monsters[0].HP = 321, 37
	var recorded []PlayerCommand
	direct.SetCommandRecorder(func(c PlayerCommand) { recorded = append(recorded, c) })
	direct.ApplyMonsterTacticalOutcome(star, len(player), len(monsters), survivors, monsters, false)
	m := direct.MonsterAtStar(star)
	if m == nil || m.Armor != 321 || m.Structure != 37 || direct.LastBattle == nil || direct.LastBattle.Enemy != "太空變形蟲" {
		t.Fatalf("怪物戰後未正確回寫：monster=%+v battle=%+v", m, direct.LastBattle)
	}
	if len(recorded) != 1 || recorded[0].Name != CmdMonsterCombatOutcome {
		t.Fatalf("怪物戰術結果應是一條可重播指令：%+v", recorded)
	}
	replayed, _ := makeSession()
	if err := replayed.ApplyPlayerCommands(recorded); err != nil {
		t.Fatal(err)
	}
	if got, want := replayed.StateHash(), direct.StateHash(); got != want {
		t.Fatalf("怪物戰術重播狀態不同：%s vs %s", got[:12], want[:12])
	}
}

func TestMonsterTacticalVictoryRemovesGuard(t *testing.T) {
	s, star := monsterTacticalSession(t, gamedata.MonsterHydra)
	player, monsters, reason := s.StartMonsterCombat(star)
	if reason != "" || len(monsters) != 1 {
		t.Fatal(reason)
	}
	survivors := map[string]bool{}
	for _, sh := range player {
		survivors[sh.Name] = true
	}
	s.ApplyMonsterTacticalOutcome(star, len(player), len(monsters), survivors, nil, true)
	if s.MonsterAtStar(star) != nil {
		t.Fatal("怪物全滅後應解除星系封鎖")
	}
}

func TestMonsterTacticalOutcomeKeepsLaterSameStarSideAndReplays(t *testing.T) {
	makeSession := func() (*GameSession, int) {
		s, star := monsterTacticalSession(t, gamedata.MonsterAmoeba)
		crystal, _ := gamedata.MonsterStatsFor(gamedata.MonsterCrystal)
		s.Monsters = append(s.Monsters, MonsterGuard{StarIndex: star, Kind: gamedata.MonsterCrystal,
			Structure: crystal.Structure, Armor: crystal.Armor})
		return s, star
	}
	direct, star := makeSession()
	player, monsters, reason := direct.StartMonsterCombat(star)
	if reason != "" || len(monsters) == 0 {
		t.Fatalf("建立第一個怪獸 side 失敗：%q", reason)
	}
	survivors := map[string]bool{}
	for _, ship := range player {
		survivors[ship.Name] = true
	}
	var recorded []PlayerCommand
	direct.SetCommandRecorder(func(c PlayerCommand) { recorded = append(recorded, c) })
	direct.ApplyMonsterTacticalOutcome(star, len(player), len(monsters), survivors, nil, true)
	groups := direct.MonsterGroupsAtStar(star)
	if len(groups) != 1 || groups[0].Kind != gamedata.MonsterCrystal {
		t.Fatalf("勝利只能移除本次 side，後續 side 必須保留：%+v", groups)
	}
	if len(recorded) != 1 || recorded[0].Name != CmdMonsterCombatOutcome {
		t.Fatalf("戰果應記成單一鎖步指令：%+v", recorded)
	}
	replayed, _ := makeSession()
	if err := replayed.ApplyPlayerCommands(recorded); err != nil {
		t.Fatal(err)
	}
	if got, want := replayed.StateHash(), direct.StateHash(); got != want {
		t.Fatalf("同星多 side 戰果重播分岔：%s vs %s", got[:12], want[:12])
	}
}
