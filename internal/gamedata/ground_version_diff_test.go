package gamedata

import "testing"

// TestGroundCommandoAttackerForceBonus 驗證 #6(非差異項)攻方 Commando 加成表:tier 0 無加成,
// tier 1(一般)= 5,tier ≥2(進階)= 7(2.5×3=7.5 捨去,見 ground_version_diff.go 註解)。
func TestGroundCommandoAttackerForceBonus(t *testing.T) {
	cases := []struct {
		tier int
		want int
	}{
		{0, 0},
		{1, 5},
		{2, 7},
		{3, 7}, // tier 沒有 3 級,但函式對 >=2 一視同仁,不 panic
	}
	for _, c := range cases {
		if got := GroundCommandoAttackerForceBonus(c.tier); got != c.want {
			t.Errorf("GroundCommandoAttackerForceBonus(%d) = %d, want %d", c.tier, got, c.want)
		}
	}
}

// TestDefenderCommandoBonusVersionDiff 驗證 #5 真正的版本差異:同一 tier,Profile13 與
// Profile15 的 DefenderCommandoBonus 套進 GroundCommandoDefenderForceBonus 後,算出的守方
// 加成必須不同(1.3 維持基準值 2/3 不變,1.5 追平攻方變成 5/7)。
func TestDefenderCommandoBonusVersionDiff(t *testing.T) {
	p13 := Profile13()
	p15 := Profile15()

	if p13.DefenderCommandoBonus == p15.DefenderCommandoBonus {
		t.Fatalf("DefenderCommandoBonus 兩版不應相同:13=%v 15=%v", p13.DefenderCommandoBonus, p15.DefenderCommandoBonus)
	}

	for _, tier := range []int{1, 2} {
		b13 := GroundCommandoDefenderForceBonus(tier, p13.DefenderCommandoBonus)
		b15 := GroundCommandoDefenderForceBonus(tier, p15.DefenderCommandoBonus)
		if b13 == b15 {
			t.Errorf("tier=%d:1.3 守方加成(%d)與 1.5 守方加成(%d)不應相同", tier, b13, b15)
		}
		if b15 <= b13 {
			t.Errorf("tier=%d:1.5 守方加成(%d)應高於 1.3(%d)(1.5 追平攻方 2.5x 加乘)", tier, b15, b13)
		}
	}

	// 具體數字釘死(避免未來改動悄悄變更近似值卻無測試察覺):
	if got := GroundCommandoDefenderForceBonus(1, p13.DefenderCommandoBonus); got != 2 {
		t.Errorf("tier1 1.3 守方加成 = %d, want 2", got)
	}
	if got := GroundCommandoDefenderForceBonus(1, p15.DefenderCommandoBonus); got != 5 {
		t.Errorf("tier1 1.5 守方加成 = %d, want 5", got)
	}
	if got := GroundCommandoDefenderForceBonus(2, p13.DefenderCommandoBonus); got != 3 {
		t.Errorf("tier2 1.3 守方加成 = %d, want 3", got)
	}
	if got := GroundCommandoDefenderForceBonus(2, p15.DefenderCommandoBonus); got != 7 {
		t.Errorf("tier2 1.5 守方加成 = %d, want 7", got)
	}
}

// TestGroundCommandoDefenderForceBonus_NoTierNoBonus tier=0(無 Commando 領袖)不論版本一律 0。
func TestGroundCommandoDefenderForceBonus_NoTierNoBonus(t *testing.T) {
	if got := GroundCommandoDefenderForceBonus(0, Profile13().DefenderCommandoBonus); got != 0 {
		t.Errorf("tier0 1.3 守方加成 = %d, want 0", got)
	}
	if got := GroundCommandoDefenderForceBonus(0, Profile15().DefenderCommandoBonus); got != 0 {
		t.Errorf("tier0 1.5 守方加成 = %d, want 0", got)
	}
}

// TestBombardmentPlanetSizeScaling 驗證 #11(非差異項)行星尺寸幾何近似:同樣的 hits,行星
// 越大(Huge)扣的人口越少(較耐轟),越小(Tiny)扣的越多,Medium 維持「loss==hits」的基準行為
// (與換係數前的既有邏輯一致,見 GroundBombardPopulationLoss 註解)。
func TestBombardmentPlanetSizeScaling(t *testing.T) {
	const hits = 12

	tiny := GroundBombardPopulationLoss(hits, TINY_PLANET)
	small := GroundBombardPopulationLoss(hits, SMALL_PLANET)
	medium := GroundBombardPopulationLoss(hits, MEDIUM_PLANET)
	large := GroundBombardPopulationLoss(hits, LARGE_PLANET)
	huge := GroundBombardPopulationLoss(hits, HUGE_PLANET)

	if medium != hits {
		t.Errorf("Medium 應維持 loss==hits(基準),got %d want %d", medium, hits)
	}
	if !(tiny > small && small > medium && medium > large && large > huge) {
		t.Errorf("行星越大應扣越少人口,got tiny=%d small=%d medium=%d large=%d huge=%d", tiny, small, medium, large, huge)
	}
	if huge <= 0 {
		t.Errorf("Huge 行星扣減量不應為 0 或負(hits=%d),got %d", hits, huge)
	}

	// 具體數字釘死(3-4-6-7-8 係數,baseline=6,見 ground_version_diff.go):
	wantTiny, wantSmall, wantMedium, wantLarge, wantHuge := 24, 18, 12, 10, 9
	if tiny != wantTiny {
		t.Errorf("tiny loss = %d, want %d", tiny, wantTiny)
	}
	if small != wantSmall {
		t.Errorf("small loss = %d, want %d", small, wantSmall)
	}
	if medium != wantMedium {
		t.Errorf("medium loss = %d, want %d", medium, wantMedium)
	}
	if large != wantLarge {
		t.Errorf("large loss = %d, want %d", large, wantLarge)
	}
	if huge != wantHuge {
		t.Errorf("huge loss = %d, want %d", huge, wantHuge)
	}
}

// TestBombardmentPlanetSizeScaling_ZeroOrNegativeHits hits<=0 一律回 0,不產生負數人口損傷。
func TestBombardmentPlanetSizeScaling_ZeroOrNegativeHits(t *testing.T) {
	if got := GroundBombardPopulationLoss(0, MEDIUM_PLANET); got != 0 {
		t.Errorf("hits=0 應回 0,got %d", got)
	}
	if got := GroundBombardPopulationLoss(-5, HUGE_PLANET); got != 0 {
		t.Errorf("hits=-5 應回 0,got %d", got)
	}
}

// TestGroundDefenseArmorMultiplier_LockedValue #9(非差異項)PARAMETERS.CFG:1772-1775 逐字
// 數字鎖定,避免未來改動悄悄變更卻無測試察覺。目前無消費端(掛鉤備妥,見常數註解)。
func TestGroundDefenseArmorMultiplier_LockedValue(t *testing.T) {
	if GroundDefenseArmorMultiplier != 100 {
		t.Errorf("GroundDefenseArmorMultiplier = %d, want 100", GroundDefenseArmorMultiplier)
	}
}

// TestGroundCivilianArmorHP_LockedValue #8(非差異項)PARAMETERS.CFG:1778-1786 逐字數字鎖定。
// 目前無消費端(掛鉤備妥,見常數註解)。
func TestGroundCivilianArmorHP_LockedValue(t *testing.T) {
	if GroundCivilianArmorHP != 100 {
		t.Errorf("GroundCivilianArmorHP = %d, want 100", GroundCivilianArmorHP)
	}
}
