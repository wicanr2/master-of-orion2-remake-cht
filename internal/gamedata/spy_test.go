package gamedata

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestSpySlotBonus(t *testing.T) {
	// 手冊(Spy Bonuses):前 5 人各 +2、6-10 人各 +1、之後每 2 人 +1;
	// 「spy 11 the bonus is still +15 while spy 12 brings it up to +16」;
	// 上限「+41 for 62 or 63 spies」。
	cases := map[int]int{
		0:  0,
		1:  2,
		5:  10,
		6:  11,
		10: 15,
		11: 15,
		12: 16,
		62: 41,
		63: 41,
		64: 41, // 手冊上限 63,超出夾在範圍內
		-1: 0,  // 負數夾在範圍內
	}
	for n, want := range cases {
		if got := SpySlotBonus(n); got != want {
			t.Errorf("SpySlotBonus(%d) = %d,預期 %d", n, got, want)
		}
	}
}

func TestSpyGovernmentDefenseBonus(t *testing.T) {
	cases := map[SpyGovernmentType]int{
		SpyGovFeudalism:           0,
		SpyGovConfederation:       0,
		SpyGovDictatorship:        10,
		SpyGovImperium:            15,
		SpyGovDemocracy:           -10,
		SpyGovFederation:          -10,
		SpyGovUnification:         15,
		SpyGovGalacticUnification: 15,
	}
	for gov, want := range cases {
		if got := SpyGovernmentDefenseBonus(gov); got != want {
			t.Errorf("SpyGovernmentDefenseBonus(%d) = %d,預期 %d", gov, got, want)
		}
	}
}

func TestSpyRaceTraitBonus(t *testing.T) {
	if got := SpyRaceTraitBonus(SpyRaceTraitMinus3); got != -10 {
		t.Errorf("SpyRaceTraitBonus(-3 picks) = %d,預期 -10", got)
	}
	if got := SpyRaceTraitBonus(SpyRaceTraitPlus3); got != 10 {
		t.Errorf("SpyRaceTraitBonus(+3 picks) = %d,預期 10", got)
	}
	if got := SpyRaceTraitBonus(SpyRaceTraitPlus6); got != 20 {
		t.Errorf("SpyRaceTraitBonus(+6 picks) = %d,預期 20", got)
	}
	if got := SpyTelepathicRaceBonus; got != 10 {
		t.Errorf("SpyTelepathicRaceBonus = %d,預期 10", got)
	}
}

func TestSpyTechnologyBonus(t *testing.T) {
	cases := map[Technology]int{
		TECH_NEURAL_SCANNER:      10,
		TECH_TELEPATHIC_TRAINING: 5,
		TECH_CYBERSECURITY_LINK:  10,
		TECH_STEALTH_SUIT:        10,
		TECH_PSIONICS:            10,
		TECH_LASER_CANNON:        0, // 手冊未列,應回 0,不誤加成
	}
	for tech, want := range cases {
		if got := SpyTechnologyBonus(tech); got != want {
			t.Errorf("SpyTechnologyBonus(%d) = %d,預期 %d", tech, got, want)
		}
	}
}

func TestSpyEffectiveThreshold(t *testing.T) {
	// E = T + DB - AB
	if got := SpyEffectiveThreshold(SpyThresholdSteal, 20, 50); got != 50 {
		t.Errorf("SpyEffectiveThreshold(80,20,50) = %d,預期 50", got)
	}
	if got := SpyEffectiveThreshold(SpyThresholdSabotage, 0, 0); got != 70 {
		t.Errorf("SpyEffectiveThreshold(70,0,0) = %d,預期 70", got)
	}
}

func TestSpyRollChance(t *testing.T) {
	// 手算對照手冊分段公式:
	// E=-150(<=-100) → p=1
	if got := SpyRollChance(-150); !almostEqual(got, 1) {
		t.Errorf("SpyRollChance(-150) = %v,預期 1", got)
	}
	// E=-100(邊界,兩段公式應一致):1-(101-100)*(100-100)/2/10000 = 1-0 = 1
	if got := SpyRollChance(-100); !almostEqual(got, 1) {
		t.Errorf("SpyRollChance(-100) = %v,預期 1", got)
	}
	// E=-50: 1-(101-50)*(100-50)/2/10000 = 1-(51*50)/2/10000 = 1-1275/10000 = 0.8725
	if got := SpyRollChance(-50); !almostEqual(got, 0.8725) {
		t.Errorf("SpyRollChance(-50) = %v,預期 0.8725", got)
	}
	// E=0: (99-0)*(98-0)/2/9900+0.01 = (99*98)/2/9900+0.01 = 9702/2/9900+0.01 = 4851/9900+0.01 = 0.5
	if got := SpyRollChance(0); !almostEqual(got, 0.5) {
		t.Errorf("SpyRollChance(0) = %v,預期 0.5", got)
	}
	// E=50: (99-50)*(98-50)/2/9900+0.01 = (49*48)/2/9900+0.01 = 2352/2/9900+0.01 = 1176/9900+0.01
	want50 := 1176.0/9900.0 + 0.01
	if got := SpyRollChance(50); !almostEqual(got, want50) {
		t.Errorf("SpyRollChance(50) = %v,預期 %v", got, want50)
	}
	// E=99(邊界):(99-99)*(98-99)/2/9900+0.01 = 0+0.01 = 0.01
	if got := SpyRollChance(99); !almostEqual(got, 0.01) {
		t.Errorf("SpyRollChance(99) = %v,預期 0.01", got)
	}
	// E=200(>99) → p=0.01(下限)
	if got := SpyRollChance(200); !almostEqual(got, 0.01) {
		t.Errorf("SpyRollChance(200) = %v,預期 0.01", got)
	}
}

func TestSpyVsSpyBonus(t *testing.T) {
	// 手冊:「the defender gets an extra +20 bonus」
	if got := SpyVsSpyDefenderBonus(10); got != 30 {
		t.Errorf("SpyVsSpyDefenderBonus(10) = %d,預期 30", got)
	}
	// 手冊:「the attacker gets +20 if he has chosen HIDE」
	if got := SpyVsSpyAttackerBonus(15, true); got != 35 {
		t.Errorf("SpyVsSpyAttackerBonus(15,HIDE) = %d,預期 35", got)
	}
	if got := SpyVsSpyAttackerBonus(15, false); got != 15 {
		t.Errorf("SpyVsSpyAttackerBonus(15,非HIDE) = %d,預期 15(無加成)", got)
	}
	if SpyVsSpyDefenderKillThreshold != 80 || SpyVsSpyAttackerKillThreshold != -80 {
		t.Errorf("SpyVsSpy 擊殺門檻常數與手冊不符:defender=%d,attacker=%d",
			SpyVsSpyDefenderKillThreshold, SpyVsSpyAttackerKillThreshold)
	}
}
