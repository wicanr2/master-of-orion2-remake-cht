package gamedata

import "testing"

func TestMoraleProductionOutput(t *testing.T) {
	// base*(100+moralePercent)/100,手冊 p.65-66、p.169-170:每點士氣 = 10% 產出變化。
	cases := []struct{ base, morale, want int }{
		{100, 0, 100},
		{100, 20, 120}, // 2 個笑臉
		{100, -20, 80}, // 2 個哭臉
		{100, -35, 65}, // Dictatorship 首都淪陷
		{200, 50, 300},
	}
	for _, c := range cases {
		if got := MoraleProductionOutput(c.base, c.morale); got != c.want {
			t.Errorf("MoraleProductionOutput(%d,%d) = %d,預期 %d", c.base, c.morale, got, c.want)
		}
	}
}

func TestMoraleGovernmentBase(t *testing.T) {
	// 手冊 p.165-167、p.21-22:Feudal/Confederation/Dictatorship 無 Barracks -20%,有則 0%。
	cases := []struct {
		gov         MoraleGovernmentType
		hasBarracks bool
		want        int
	}{
		{MoraleGovFeudalism, false, -20},
		{MoraleGovFeudalism, true, 0},
		{MoraleGovConfederation, false, -20},
		{MoraleGovConfederation, true, 0},
		{MoraleGovDictatorship, false, -20},
		{MoraleGovDictatorship, true, 0},
		// Imperium:固定 +20%,無 Barracks 再疊加 -20%(淨 0%);有 Barracks 淨 +20%。
		{MoraleGovImperium, false, 0},
		{MoraleGovImperium, true, 20},
		{MoraleGovDemocracy, false, 0},
		{MoraleGovDemocracy, true, 0},
		{MoraleGovFederation, false, 0},
		{MoraleGovFederation, true, 0},
		{MoraleGovUnification, false, 0},
		{MoraleGovGalacticUnification, false, 0},
	}
	for _, c := range cases {
		if got := MoraleGovernmentBase(c.gov, c.hasBarracks); got != c.want {
			t.Errorf("MoraleGovernmentBase(%v,%v) = %d,預期 %d", c.gov, c.hasBarracks, got, c.want)
		}
	}
}

func TestMoraleCapitalCapturedPenalty(t *testing.T) {
	// 手冊 p.165-167:各政府首都淪陷懲罰。
	cases := []struct {
		gov  MoraleGovernmentType
		want int
	}{
		{MoraleGovFeudalism, -50},
		{MoraleGovConfederation, -50},
		{MoraleGovDictatorship, -35},
		{MoraleGovImperium, -35},
		{MoraleGovDemocracy, -20},
		{MoraleGovFederation, -20},
		{MoraleGovUnification, 0},
		{MoraleGovGalacticUnification, 0},
	}
	for _, c := range cases {
		if got := MoraleCapitalCapturedPenalty(c.gov); got != c.want {
			t.Errorf("MoraleCapitalCapturedPenalty(%v) = %d,預期 %d", c.gov, got, c.want)
		}
	}
}

func TestMoraleMultiRacialPenalty(t *testing.T) {
	// 手冊 p.66-67、p.92-93:無 Alien Management Center 時 -20%,有則 0%。
	if got := MoraleMultiRacialPenalty(false); got != -20 {
		t.Errorf("MoraleMultiRacialPenalty(false) = %d,預期 -20", got)
	}
	if got := MoraleMultiRacialPenalty(true); got != 0 {
		t.Errorf("MoraleMultiRacialPenalty(true) = %d,預期 0", got)
	}
}

func TestMoralePsionicsGovernmentBonus(t *testing.T) {
	// 手冊 p.100-101:僅 Dictatorship/Imperium/Feudalism/Confederation 適用 +10%。
	cases := []struct {
		gov  MoraleGovernmentType
		want int
	}{
		{MoraleGovFeudalism, 10},
		{MoraleGovConfederation, 10},
		{MoraleGovDictatorship, 10},
		{MoraleGovImperium, 10},
		{MoraleGovDemocracy, 0},
		{MoraleGovFederation, 0},
		{MoraleGovUnification, 0},
		{MoraleGovGalacticUnification, 0},
	}
	for _, c := range cases {
		if got := MoralePsionicsGovernmentBonus(c.gov); got != c.want {
			t.Errorf("MoralePsionicsGovernmentBonus(%v) = %d,預期 %d", c.gov, got, c.want)
		}
	}
}

func TestMoraleUnificationProductionBonus(t *testing.T) {
	// 手冊 p.166-167:Unification +50%、Galactic Unification +100%,僅作用於 food/industry。
	cases := []struct {
		gov  MoraleGovernmentType
		want int
	}{
		{MoraleGovUnification, 50},
		{MoraleGovGalacticUnification, 100},
		{MoraleGovFeudalism, 0},
		{MoraleGovDemocracy, 0},
	}
	for _, c := range cases {
		if got := MoraleUnificationProductionBonus(c.gov); got != c.want {
			t.Errorf("MoraleUnificationProductionBonus(%v) = %d,預期 %d", c.gov, got, c.want)
		}
	}
}

func TestMoraleBuildingBonusConstants(t *testing.T) {
	// 手冊固定加成常數,直接對照原文數字。
	if MoraleHoloSimulatorBonus != 20 {
		t.Errorf("MoraleHoloSimulatorBonus = %d,預期 20", MoraleHoloSimulatorBonus)
	}
	if MoralePleasureDomeBonus != 30 {
		t.Errorf("MoralePleasureDomeBonus = %d,預期 30", MoralePleasureDomeBonus)
	}
	if MoraleVirtualRealityNetworkBonus != 20 {
		t.Errorf("MoraleVirtualRealityNetworkBonus = %d,預期 20", MoraleVirtualRealityNetworkBonus)
	}
}
