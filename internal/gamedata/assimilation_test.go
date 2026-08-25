package gamedata

import "testing"

// 手冊 p.21–24 的八個政體逐個釘住。
func TestAssimilationTurnsMatchTheManual(t *testing.T) {
	for _, tc := range []struct {
		gov  AssimilationGovernment
		want int
		name string
	}{
		{AssimFeudal, 8, "封建"},
		{AssimConfederation, 4, "邦聯"},
		{AssimDictatorship, 8, "獨裁"},
		{AssimImperium, 4, "帝國"},
		{AssimDemocracy, 4, "民主"},
		{AssimFederation, 2, "聯邦"},
		{AssimUnification, 20, "統一"},
		{AssimGalacticUnification, 15, "銀河統一"},
	} {
		if got := AssimilationTurns(tc.gov, false, false, false); got != tc.want {
			t.Errorf("%s 應是 %d 回合,得到 %d", tc.name, tc.want, got)
		}
	}
	// 統一與民主差五倍——那是原版把「和平流」與「征服流」分開的規則手段,
	// 被改掉就不是這個遊戲了。
	if AssimilationTurns(AssimUnification, false, false, false) !=
		5*AssimilationTurns(AssimDemocracy, false, false, false) {
		t.Error("統一應是民主的五倍回合數")
	}
}

// 四個進階政體是基礎政體的升級版,而且都同化得更快。
func TestAdvancedGovernmentsAssimilateFaster(t *testing.T) {
	for _, base := range []AssimilationGovernment{
		AssimFeudal, AssimDictatorship, AssimDemocracy, AssimUnification,
	} {
		adv := AssimilationAdvancedForm(base)
		if adv == base {
			t.Errorf("政體 %d 應有進階形式", base)
			continue
		}
		b := AssimilationTurns(base, false, false, false)
		a := AssimilationTurns(adv, false, false, false)
		if a >= b {
			t.Errorf("進階政體 %d 的 %d 回合應少於基礎 %d 的 %d 回合", adv, a, base, b)
		}
	}
	// 已經是進階形式的回傳自己(沒有第三層)。
	if got := AssimilationAdvancedForm(AssimFederation); got != AssimFederation {
		t.Errorf("聯邦沒有更進階的形式,應回自己,得到 %d", got)
	}
}

// 異族管理中心「regardless of government」——直接蓋掉政體那一格。
func TestAlienManagementCenterOverridesTheGovernment(t *testing.T) {
	for _, gov := range []AssimilationGovernment{
		AssimFeudal, AssimDictatorship, AssimDemocracy, AssimUnification, AssimGalacticUnification,
	} {
		if got := AssimilationTurns(gov, true, false, false); got != AssimilationCenterTurns {
			t.Errorf("政體 %d 有異族管理中心時應是固定 %d 回合,得到 %d",
				gov, AssimilationCenterTurns, got)
		}
	}
	// 對統一政體來說那是十倍速——這棟建築的價值就在這裡。
	slow := AssimilationTurns(AssimUnification, false, false, false)
	fast := AssimilationTurns(AssimUnification, true, false, false)
	if slow/fast != 10 {
		t.Errorf("統一政體蓋了異族管理中心應是十倍速:%d → %d", slow, fast)
	}
}

// 排斥種族速度減半(手冊 only half the normal rate → 回合數 ×2)。
func TestRepulsiveRacesAssimilateAtHalfRate(t *testing.T) {
	base := AssimilationTurns(AssimDemocracy, false, false, false)
	rep := AssimilationTurns(AssimDemocracy, false, true, false)
	if rep != base*2 {
		t.Errorf("排斥種族應是兩倍回合數:%d → %d", base, rep)
	}
	// 建築的固定值也要吃這個修正(手冊:「The adjustment for a Charismatic or
	// Repulsive race is applied to **this base rate**」)。
	if got := AssimilationTurns(AssimUnification, true, true, false); got != AssimilationCenterTurns*2 {
		t.Errorf("有建築的排斥種族應是 %d 回合,得到 %d", AssimilationCenterTurns*2, got)
	}
}

func TestCharismaticDoublesOriginalAssimilationRate(t *testing.T) {
	if got := AssimilationRate(AssimDemocracy, false, false, true); got != 120 {
		t.Fatalf("魅力民主 rate = %d, want 120", got)
	}
	if got := AssimilationTurns(AssimGalacticUnification, false, false, true); got != 8 {
		t.Fatalf("魅力銀河統一應 ceil(240/32)=8，got %d", got)
	}
	if got := AssimilationTurns(AssimUnification, true, false, true); got != 1 {
		t.Fatalf("魅力＋異族管理中心應 1 回合，got %d", got)
	}
}

func TestCharismaticPrecedesRepulsiveWhenBothFlagsExist(t *testing.T) {
	got := AssimilationRate(AssimDemocracy, false, true, true)
	if got != 120 {
		t.Fatalf("兩旗標同存應只走魅力分支，rate=%d want 120", got)
	}
}

// 未知政體退回獨裁(remake 的預設),不是回 0 ——0 會變成「瞬間同化」。
func TestUnknownGovernmentFallsBackInsteadOfZero(t *testing.T) {
	got := AssimilationTurns(AssimilationGovernment(99), false, false, false)
	if got != AssimilationTurns(AssimDictatorship, false, false, false) {
		t.Errorf("未知政體應退回獨裁的回合數,得到 %d", got)
	}
	if got <= 0 {
		t.Error("回合數不可為 0——那等於瞬間同化")
	}
}

func TestAssimilationRemainingTurnsUsesRawProgress(t *testing.T) {
	if got := AssimilationRemainingTurns(2, 36, 120); got != 4 {
		t.Fatalf("2 人口、36/240、rate120 ETA=%d want 4", got)
	}
}
