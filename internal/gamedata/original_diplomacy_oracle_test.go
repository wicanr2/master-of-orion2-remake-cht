package gamedata

import "testing"

func TestOriginalDiplomacyRawTablesKeepAddressedValues(t *testing.T) {
	if len(OriginalTradeAgreementValues) != OriginalDiplomacyOracleWordCount ||
		len(OriginalDiplomacyPersonalityValues) != OriginalDiplomacyOracleWordCount ||
		len(OriginalGiftResponseValues) != OriginalDiplomacyOracleWordCount ||
		len(OriginalTradeEventValues) != OriginalDiplomacyOracleWordCount {
		t.Fatal("外交 raw table 應各保留完整 16 個 signed word")
	}
	if got := OriginalTradeAgreementValues[4]; got != 50 {
		t.Fatalf("raw 0x18105C government 4=%d,want 50", got)
	}
	if got := OriginalDiplomacyPersonalityValues[10]; got != -520 {
		t.Fatalf("raw 0x180DB8[10]=%d,want -520", got)
	}
	if got := OriginalGiftResponseValues[6]; got != -70 {
		t.Fatalf("raw 0x180CCC[6]=%d,want -70", got)
	}
	if got := OriginalTradeEventValues[8]; got != -10 {
		t.Fatalf("raw 0x181070[8]=%d,want -10", got)
	}
}

func TestOriginalTradeAgreementKnownBranches(t *testing.T) {
	if got, ok := OriginalTradeStartRelationDelta(0, 1); !ok || got != 7 {
		t.Fatalf("mode1 relation delta=(%d,%v),want (7,true)", got, ok)
	}
	if got, ok := OriginalTradeStartRelationDelta(4, 2); !ok || got != 100 {
		t.Fatalf("mode2 relation delta=(%d,%v),want (100,true)", got, ok)
	}
	if got, ok := OriginalTradeAgreementGoalPercent(2, true); !ok || got != 150 {
		t.Fatalf("trait goal percent=(%d,%v),want (150,true)", got, ok)
	}
	if got, ok := OriginalTradeAgreementGoalPercent(5, true); !ok || got != 225 {
		t.Fatalf("federation trait goal percent=(%d,%v),want (225,true)", got, ok)
	}
	if _, ok := OriginalTradeStartRelationDelta(16, 1); ok {
		t.Fatal("越界政府索引不應取得外交 relation delta")
	}
}

func TestOriginalSpecialTradeLeaderBonusTable(t *testing.T) {
	for _, tc := range []struct {
		experience, tier, want int
	}{
		{0, 1, 10}, {59, 1, 10}, {60, 1, 20}, {500, 1, 50},
		{0, 2, 15}, {150, 2, 45}, {500, 2, 75},
	} {
		if got := OriginalSpecialTradeLeaderBonus(tc.experience, tc.tier, false, 0); got != tc.want {
			t.Errorf("experience=%d tier=%d bonus=%d,want %d", tc.experience, tc.tier, got, tc.want)
		}
	}
	if got := OriginalSpecialTradeLeaderBonus(1000, 2, true, 1); got != 90 {
		t.Fatalf("Warlord 高經驗 Trader tier2 bonus=%d,want 90", got)
	}
	if got := OriginalSpecialTradeLeaderBonus(1000, 2, true, 0x42); got != 75 {
		t.Fatalf("特殊 leader 0x42 不應吃 Warlord 最高桶,bonus=%d,want 75", got)
	}
	if got, ok := OriginalTradeAgreementGoalPercentWithLeader(4, true, 45); !ok || got != 245 {
		t.Fatalf("政府／神級商人／活動 Trader 目標=(%d,%v),want (245,true)", got, ok)
	}
}
