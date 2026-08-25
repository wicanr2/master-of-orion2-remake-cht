package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 每回合每艘船 +1(手冊 p.121「Each turn in space counts for 1 experience point」)。
func TestCrewGainsOneExperiencePerTurn(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	before := s.Fleet().Ships[0].CrewXP
	s.advanceCrewExperience()
	if got := s.Fleet().Ships[0].CrewXP - before; got != 1 {
		t.Errorf("每回合應 +1 經驗,得到 +%d", got)
	}
	// 50 回合之後應該從新兵升成正規兵(門檻 50)。
	for i := 0; i < 49; i++ {
		s.advanceCrewExperience()
	}
	if lv := s.shipCrewLevel(s.Fleet().Ships[0]); lv != gamedata.CrewRegular {
		t.Errorf("累積 50 經驗應升正規兵,得到等級 %d(經驗 %d)", lv, s.Fleet().Ships[0].CrewXP)
	}
}

func TestCrewTurnExperienceCapsAtOriginalFiveHundred(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = []Ship{
		{Name: "戰鬥艦", Class: "戰艦", CrewXP: 499},
		{Name: "殖民船", Class: ColonyShipClass, CrewXP: 500},
		{Name: "前哨船", Class: OutpostShipClass, CrewXP: 499},
	}
	s.Fleet().ETA = 3 // 排除學院，只驗基本 +1 與支援艦也進 XP consumer。
	s.advanceCrewExperience()
	for _, sh := range s.Fleet().Ships {
		if sh.CrewXP != gamedata.CrewXPTurnMaximum {
			t.Errorf("%s 回合 XP 應夾在 %d，got %d", sh.Name, gamedata.CrewXPTurnMaximum, sh.CrewXP)
		}
	}
	s.advanceCrewExperience()
	for _, sh := range s.Fleet().Ships {
		if sh.CrewXP != gamedata.CrewXPTurnMaximum {
			t.Errorf("%s 在上限後不應繼續成長，got %d", sh.Name, sh.CrewXP)
		}
	}
}

// 太空學院在同一星系時每回合多 +1(手冊 p.97)。
func TestSpaceAcademyAddsOneExperiencePerTurnInTheSameSystem(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	star := s.colonyStar(0)
	if star < 0 {
		t.Fatal("測試前提:第一個殖民地應有對應的星")
	}
	s.Fleet().AtStar = star
	s.Fleet().ETA = 0

	base := s.Fleet().Ships[0].CrewXP
	s.advanceCrewExperience()
	plain := s.Fleet().Ships[0].CrewXP - base

	s.ColonyBuildings[0][spaceAcademyName] = true
	base2 := s.Fleet().Ships[0].CrewXP
	s.advanceCrewExperience()
	withAcademy := s.Fleet().Ships[0].CrewXP - base2

	if withAcademy != plain+1 {
		t.Errorf("同星系一座太空學院應每回合多 +1:沒學院 +%d、有學院 +%d", plain, withAcademy)
	}
}

// 航行中的艦隊拿不到學院加成——AtStar 只在停泊時有意義。
func TestSpaceAcademyDoesNotReachFleetsInTransit(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	star := s.colonyStar(0)
	s.ColonyBuildings[0][spaceAcademyName] = true
	s.Fleet().AtStar = star
	s.Fleet().ETA = 3 // 航行中
	base := s.Fleet().Ships[0].CrewXP
	s.advanceCrewExperience()
	if got := s.Fleet().Ships[0].CrewXP - base; got != 1 {
		t.Errorf("航行中應只拿到基本的 +1,得到 +%d", got)
	}
}

// 太空學院讓該殖民地造的船起始等級 +1(手冊 p.97),用起始經驗表達。
func TestSpaceAcademyRaisesStartingCrewLevel(t *testing.T) {
	s := NewDemoSession()
	if got := s.newShipCrewXP(0); got != 0 {
		t.Errorf("沒有學院時新船應是新兵(0 經驗),得到 %d", got)
	}
	s.ColonyBuildings[0][spaceAcademyName] = true
	xp := s.newShipCrewXP(0)
	if lv := gamedata.CrewLevelForXP(xp, false); lv != gamedata.CrewRegular {
		t.Errorf("有學院時新船應是正規兵,得到等級 %d(經驗 %d)", lv, xp)
	}
	// 別的殖民地的學院不算。
	if got := s.newShipCrewXP(-1); got != 0 {
		t.Errorf("沒指定殖民地時應走一般起始等級,得到 %d", got)
	}
}

// 艦員等級真的加進戰鬥攻擊力(手冊 BA 欄)。
func TestCrewLevelFeedsIntoCombatAttack(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.RaceCombatPct = 0
	s.Fleet().Ships = []Ship{{Name: "測試艦", Class: "巡洋艦", Weapon: "雷射砲", Armor: "無裝甲", Shield: "無護盾", Special: "無", WeaponAttack: 10}}
	green, _ := s.mkPlayerCombatantsIndexed()
	if len(green) != 1 {
		t.Fatalf("應有 1 艘參戰艦,得到 %d", len(green))
	}
	s.Fleet().Ships[0].CrewXP = gamedata.CrewXPForLevel(gamedata.CrewElite, false)
	elite, _ := s.mkPlayerCombatantsIndexed()
	if diff := elite[0].atk - green[0].atk; diff != gamedata.ShipCrewOffenseBonus(gamedata.CrewElite) {
		t.Errorf("精銳艦員應加 +%d 攻擊,實際多 %d",
			gamedata.ShipCrewOffenseBonus(gamedata.CrewElite), diff)
	}
}

// 被擊沉的敵艦艦體等級由「開打前 − 倖存」還原。
func TestDestroyedEnemySizeClassesReconstruction(t *testing.T) {
	start := []int{2, 8, 64, 8}   // 巡防、巡洋、末日之星、巡洋
	surv := []combatant{{atk: 8}} // 只剩一艘巡洋艦
	got := destroyedEnemySizeClasses(start, surv)
	// 倖存的那艘巡洋艦消掉的是清單裡**第一艘** 8,所以被擊沉的依原順序是 2、64、8。
	want := []int{1, 6, 3}
	if len(got) != len(want) {
		t.Fatalf("應還原出 %d 艘被擊沉,得到 %d 艘 %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 艘的艦體等級應是 %d,得到 %d", i, want[i], got[i])
		}
	}
	// 正對照:全部倖存時應是空的。
	all := []combatant{{atk: 2}, {atk: 8}, {atk: 64}, {atk: 8}}
	if got := destroyedEnemySizeClasses(start, all); len(got) != 0 {
		t.Errorf("全部倖存時不該有被擊沉的,得到 %v", got)
	}
}

// 戰力值 → 艦體等級的對照(remake 的 shipStrength 是 2 的冪)。
func TestShipSizeClassFromStrength(t *testing.T) {
	for _, tc := range []struct{ st, want int }{
		{1, 1}, {2, 1}, {4, 2}, {8, 3}, {16, 4}, {32, 5}, {64, 6}, {128, 6},
	} {
		if got := shipSizeClassFromStrength(tc.st); got != tc.want {
			t.Errorf("戰力 %d 應對到艦體等級 %d,得到 %d", tc.st, tc.want, got)
		}
	}
}

// 打贏才給經驗(手冊「Each battle **won**」)。
func TestBattleCrewXPOnlyOnVictory(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	before := s.Fleet().Ships[0].CrewXP
	got := s.awardBattleCrewXP(12, map[int]bool{0: true}) // 兩艘末日之星 = 12,折半 6
	if got != 6 {
		t.Errorf("擊沉兩艘末日之星應給 6 經驗,得到 %d", got)
	}
	if diff := s.Fleet().Ships[0].CrewXP - before; diff != 6 {
		t.Errorf("倖存艦應加到 6 經驗,實際 +%d", diff)
	}
	// 原版零擊沉仍套 min-1。
	before2 := s.Fleet().Ships[0].CrewXP
	if got := s.awardBattleCrewXP(0, map[int]bool{0: true}); got != 1 {
		t.Errorf("沒擊沉任何船的勝方 recipient 應給 1,得到 %d", got)
	}
	if s.Fleet().Ships[0].CrewXP != before2+1 {
		t.Error("沒擊沉任何船時仍應依原版加 1")
	}
	// 戰鬥寫入不套每回合 500 cap。
	s.Fleet().Ships[0].CrewXP = 500
	s.awardBattleCrewXP(2, map[int]bool{0: true})
	if got := s.Fleet().Ships[0].CrewXP; got != 501 {
		t.Errorf("戰鬥 XP 不應套每回合 500 cap,got %d", got)
	}
}

func TestTacticalCombatOutcomeAwardsOnlySurvivingParticipantsAndReplays(t *testing.T) {
	makeSession := func() *GameSession {
		s := NewDemoSession()
		s.Fleet().Ships = []Ship{
			{Name: "陣亡巡防艦", Class: "巡防艦", CrewXP: 10},
			{Name: "倖存戰艦", Class: "戰艦", CrewXP: 500},
			{Name: "殖民船", Class: ColonyShipClass, CrewXP: 20},
		}
		return s
	}

	direct := makeSession()
	var recorded []PlayerCommand
	direct.SetCommandRecorder(func(c PlayerCommand) { recorded = append(recorded, c) })
	direct.ApplyCombatOutcome("測試敵軍", 2, 2, map[string]bool{"倖存戰艦": true}, true, 4)
	if len(direct.Fleet().Ships) != 1 || direct.Fleet().Ships[0].Name != "倖存戰艦" {
		t.Fatalf("戰後應移除陣亡參戰艦及遭交戰的支援艦，got %+v", direct.Fleet().Ships)
	}
	if got := direct.Fleet().Ships[0].CrewXP; got != 502 {
		t.Errorf("倖存戰艦應取得 4/2=2 且突破 500，got %d", got)
	}
	if direct.LastBattle == nil || direct.LastBattle.CrewXPGained != 2 || direct.LastBattle.PlayerLosses != 1 {
		t.Fatalf("戰報應記 XP=2、損失=1，got %+v", direct.LastBattle)
	}
	if len(recorded) != 1 || len(recorded[0].Args) != 4 || recorded[0].Args[3] != 4 {
		t.Fatalf("戰鬥 command 應記 destroyed hull-class sum=4，got %+v", recorded)
	}

	replayed := makeSession()
	if err := replayed.ApplyPlayerCommands(recorded); err != nil {
		t.Fatalf("重播戰鬥結果失敗：%v", err)
	}
	if got, want := replayed.StateHash(), direct.StateHash(); got != want {
		t.Fatalf("戰鬥 XP command 重播後狀態不同：%s vs %s", got[:12], want[:12])
	}
}

// 統帥種族的船一出廠就是正規兵(手冊 p.121 那句 unless)。
func TestWarlordShipsStartAsRegulars(t *testing.T) {
	s := NewDemoSession()
	if lv := s.shipCrewLevel(Ship{}); lv != gamedata.CrewGreen {
		t.Errorf("一般種族的新船應是新兵,得到等級 %d", lv)
	}
	// ⚠ 2026-08-08:改成真的選姆瑞森,而不是手動把旗標設成 true。
	// 統帥不再是可以外部指派的欄位,而是由種族算出來的(見 raceOrigIdx)——
	// 這樣測到的才是「姆瑞森玩家的船出廠就是正規兵」,而不只是「這個 bool 有作用」。
	s.ApplyRace(raceIndexByEnName(t, "Mrrshan"))
	if !s.RaceWarlord() {
		t.Fatal("姆瑞森應是統帥種族")
	}
	if lv := s.shipCrewLevel(Ship{}); lv != gamedata.CrewRegular {
		t.Errorf("統帥種族的新船應是正規兵,得到等級 %d", lv)
	}
}
