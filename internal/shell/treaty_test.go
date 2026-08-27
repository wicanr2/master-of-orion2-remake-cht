package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestTributeModeUsesExecutableFiveAndTenPercentFormula(t *testing.T) {
	if TributeFivePercent.TributePercent() != 5 || TributeTenPercent.TributePercent() != 10 {
		t.Fatalf("固定納貢模式百分比錯誤: 5%%=%d,10%%=%d",
			TributeFivePercent.TributePercent(), TributeTenPercent.TributePercent())
	}
	if got := tributeCost(100, 0, TributeFivePercent); got != 5 {
		t.Fatalf("5%% 納貢成本 = %d,預期 5", got)
	}
	// 原版以「收入−既有納貢成本」作為下一筆的基準。
	if got := tributeCost(100, 5, TributeTenPercent); got != 9 {
		t.Fatalf("既有成本後 10%% 納貢成本 = %d,預期 9", got)
	}
	if got := tributeCost(-1, 0, TributeFivePercent); got != 0 {
		t.Fatalf("負收入不應產生納貢成本: %d", got)
	}
}

func TestTreatyTargetUsesExecutableGovernmentAndTraderModifiers(t *testing.T) {
	const base = 100
	tradeCases := []struct {
		name   string
		gov    gamedata.MoraleGovernmentType
		trader bool
		want   int
	}{
		{"獨裁", gamedata.MoraleGovDictatorship, false, 100},
		{"民主", gamedata.MoraleGovDemocracy, false, 150},
		{"聯邦", gamedata.MoraleGovFederation, false, 175},
		{"神級商人", gamedata.MoraleGovDictatorship, true, 150},
		{"民主神級商人", gamedata.MoraleGovDemocracy, true, 200},
	}
	for _, tc := range tradeCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := treatyTarget(base, tc.gov, tc.trader, true); got != tc.want {
				t.Fatalf("貿易目標 = %d,預期 %d", got, tc.want)
			}
		})
	}

	researchCases := []struct {
		gov  gamedata.MoraleGovernmentType
		want int
	}{
		{gamedata.MoraleGovFeudalism, 50},
		{gamedata.MoraleGovConfederation, 75},
		{gamedata.MoraleGovDictatorship, 100},
		{gamedata.MoraleGovImperium, 100},
		{gamedata.MoraleGovDemocracy, 150},
		{gamedata.MoraleGovFederation, 175},
		{gamedata.MoraleGovUnification, 100},
		{gamedata.MoraleGovGalacticUnification, 100},
	}
	for _, tc := range researchCases {
		if got := treatyTarget(base, tc.gov, false, false); got != tc.want {
			t.Errorf("政府 %d 研究目標 = %d,預期 %d", tc.gov, got, tc.want)
		}
	}
}

func TestAgreementValueUsesExecutableGoalFifthAndRemainderRoll(t *testing.T) {
	current := -10
	for turn := 1; turn <= 10; turn++ {
		current = advanceAgreementValue(current, 10, func() int {
			t.Fatal("可整除 5 的目標不應抽亂數")
			return 1
		})
		if turn == 5 && current != 0 {
			t.Fatalf("第 5 回合應損益兩平，got %d", current)
		}
	}
	if current != 10 {
		t.Fatalf("第 10 回合應到目標 10，got %d", current)
	}

	rolls := 0
	if got := advanceAgreementValue(0, 12, func() int { rolls++; return 2 }); got != 3 {
		t.Fatalf("goal=12 且 roll=2 應增加 2+1，got %d", got)
	}
	if got := advanceAgreementValue(0, 12, func() int { rolls++; return 3 }); got != 2 {
		t.Fatalf("goal=12 且 roll=3 應只增加 2，got %d", got)
	}
	if rolls != 2 {
		t.Fatalf("非整除目標每次應恰抽一次，got %d", rolls)
	}
	if got := advanceAgreementValue(20, 7, func() int {
		t.Fatal("目前值高於目標時不應抽亂數")
		return 1
	}); got != 7 {
		t.Fatalf("目標下降時應立即降到 7，got %d", got)
	}
}

func TestAgreementDirectionsFollowExecutableDescendingSlotOrder(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers = s.AIPlayers[:2]
	s.PlayerColonies[0].Population = 14 // base=7，讓每次都需要 remainder roll。
	for i := range s.AIPlayers {
		s.AIPlayers[i].Colonies[0].Population = 14
		t := &s.AIPlayers[i].Treaty
		t.TradeActive, t.ResearchActive = true, true
		t.TradeTurns, t.ResearchTurns = 1, 1
		t.PlayerTradeValue, t.PlayerResearchValue = -100, -100
		t.AITradeValue, t.AIResearchValue = -100, -100
	}
	// 順序應為 AI1 T/R、AI0 T/R、Player→AI1 T/R、Player→AI0 T/R。
	rolls := []int{1, 1, 5, 5, 1, 5, 5, 1}
	pos := 0
	s.advanceTreatiesWithRoller(func() int {
		roll := rolls[pos]
		pos++
		return roll
	})
	if pos != len(rolls) {
		t.Fatalf("擲骰次數=%d，預期 %d", pos, len(rolls))
	}
	if got := s.AIPlayers[1].Treaty; got.AITradeValue != -98 || got.AIResearchValue != -98 ||
		got.PlayerTradeValue != -98 || got.PlayerResearchValue != -99 {
		t.Fatalf("AI1／玩家方向順序不符：%+v", got)
	}
	if got := s.AIPlayers[0].Treaty; got.AITradeValue != -99 || got.AIResearchValue != -99 ||
		got.PlayerTradeValue != -99 || got.PlayerResearchValue != -98 {
		t.Fatalf("AI0／玩家方向順序不符：%+v", got)
	}
}

func TestTradeAgreementUsesBestActiveOriginalTraderLeaderBonus(t *testing.T) {
	leaders := []Leader{
		{Name: "已解除", Level: 5, RawStatus: originalLeaderLimboStatus, Skills: []LeaderSkill{{ID: int(gamedata.SKILL_TRADER), Tier: 2}}},
		{Name: "活動一般", Level: 3, RawExperience: 150, RawExperienceKnown: true, RawStatus: originalLeaderActiveStatus, Skills: []LeaderSkill{{ID: int(gamedata.SKILL_TRADER), Tier: 1}}},
		{Name: "活動進階", Level: 4, RawExperience: 500, RawExperienceKnown: true, RawStatus: 0, Skills: []LeaderSkill{{ID: int(gamedata.SKILL_TRADER), Tier: 2}}},
	}
	if got := tradeAgreementLeaderBonus(leaders, false); got != 75 {
		t.Fatalf("活動 Trader 應取最佳 75,got %d", got)
	}
	if got := treatyTargetWithLeader(100, gamedata.MoraleGovDictatorship, true, true, gotLeaderBonusForTest(leaders)); got != 225 {
		t.Fatalf("政府／神級商人／活動 Trader 目標=%d,want 225", got)
	}
}

func gotLeaderBonusForTest(leaders []Leader) int {
	return tradeAgreementLeaderBonus(leaders, false)
}

func TestTreatyStateKeepsFormalAndEconomicAgreementsSeparate(t *testing.T) {
	s := NewDemoSession()
	target := s.AIPlayers[0].Name

	if got := s.DiplomacyResponse("trade", target); got.Code == "" {
		t.Fatal("貿易提議應建立協議")
	}
	if got := s.DiplomacyResponse("research", target); got.Code == "" {
		t.Fatal("研究提議應建立協議")
	}
	if got := s.DiplomacyResponse("nonaggression", target); got.Code == "" {
		t.Fatal("互不侵犯提議應建立正式條約")
	}

	state := s.TreatyFor(target)
	if !state.TradeActive || !state.ResearchActive {
		t.Fatalf("貿易與研究協議應可並存: %+v", state)
	}
	if state.FormalPolicy != gamedata.DIPLO_NON_AGGRESSION {
		t.Fatalf("正式條約 = %d,預期互不侵犯 = %d", state.FormalPolicy, gamedata.DIPLO_NON_AGGRESSION)
	}

	s.DiplomacyResponse("break_trade", target)
	state = s.TreatyFor(target)
	if state.TradeActive || !state.ResearchActive {
		t.Fatalf("終止貿易不應清掉研究協議: %+v", state)
	}
	s.DiplomacyResponse("break_formal", target)
	state = s.TreatyFor(target)
	if state.FormalPolicy != gamedata.DIPLO_NONE {
		t.Fatalf("正式條約未終止: %+v", state)
	}
	if !s.AIPlayers[0].OriginalHumanBetrayalRaw {
		t.Fatal("玩家破壞正式條約後，AI→玩家 +0x727 應永久設為 1")
	}
	wantGrievance := -10
	if int(s.AIPlayers[0].Personality) == 4 {
		wantGrievance = -20
	}
	if got := s.AIPlayers[0].OriginalHumanTreatyGrievanceRaw; got != wantGrievance {
		t.Fatalf("正式違約 +0x7EE=%d，預期 %d", got, wantGrievance)
	}
	if s.AIPlayers[0].PopulationRaceSlotKnown &&
		(!s.AIPlayers[0].OriginalHumanTreatyVictimKnown ||
			s.AIPlayers[0].OriginalHumanTreatyVictimRaw != s.AIPlayers[0].PopulationRaceSlot) {
		t.Fatalf("+0x7F6 應記錄受害 AI slot：%+v", s.AIPlayers[0])
	}
}

func TestEconomicAgreementBreakDoesNotSetFormalBetrayal(t *testing.T) {
	s := NewDemoSession()
	target := s.AIPlayers[0].Name
	s.DiplomacyResponse("trade", target)
	s.DiplomacyResponse("break_trade", target)
	if s.AIPlayers[0].OriginalHumanBetrayalRaw {
		t.Fatal("只終止貿易不得寫正式條約 +0x727")
	}
	if s.AIPlayers[0].OriginalHumanTreatyGrievanceRaw != 0 || s.AIPlayers[0].OriginalHumanTreatyVictimKnown {
		t.Fatal("只終止貿易不得寫 +0x7EE／+0x7F6")
	}
}

func TestTreatySummaryPartsAreTypedAndOrdered(t *testing.T) {
	state := TreatyState{
		FormalPolicy: gamedata.DIPLO_ALLIANCE, PlayerTribute: TributeFivePercent,
		AITribute: TributeTenPercent, TradeActive: true, TradeTurns: 3, PlayerTradeValue: -2,
		ResearchActive: true, ResearchTurns: 4, PlayerResearchValue: 7,
		SpecialTrade: SpecialTradeState{Active: true, Kind: SpecialTradeFoodForCredits, Turns: 2, PlayerValue: 5},
	}
	got := TreatySummaryParts(state)
	wantKinds := []TreatySummaryKind{
		TreatySummaryAlliance, TreatySummaryPayingTribute, TreatySummaryReceivingTribute,
		TreatySummaryTrade, TreatySummaryResearch, TreatySummarySpecialFood,
	}
	if len(got) != len(wantKinds) {
		t.Fatalf("條約摘要項目數=%d，預期 %d：%+v", len(got), len(wantKinds), got)
	}
	for i, kind := range wantKinds {
		if got[i].Kind != kind {
			t.Fatalf("條約摘要第 %d 項=%s，預期 %s", i, got[i].Kind, kind)
		}
	}
	if got[1].Percent != 5 || got[2].Percent != 10 || got[3].Turns != 3 || got[3].Value != -2 || got[4].Value != 7 {
		t.Fatalf("條約摘要 typed 參數錯誤：%+v", got)
	}
	empty := TreatySummaryParts(TreatyState{})
	if len(empty) != 1 || empty[0].Kind != TreatySummaryFormalNone {
		t.Fatalf("空條約摘要=%+v，預期 formal_none", empty)
	}
	unknown := TreatySummaryParts(TreatyState{SpecialTrade: SpecialTradeState{Active: true, Kind: 99, Turns: 8}})
	if len(unknown) != 1 || unknown[0].Kind != TreatySummarySpecialUnknown || unknown[0].Turns != 8 {
		t.Fatalf("未知特殊貿易摘要=%+v", unknown)
	}
}

func TestTributeTreatyTransfersTreasuryAndCanEnd(t *testing.T) {
	s := NewDemoSession()
	target := s.AIPlayers[0].Name
	if got := s.DiplomacyResponse("tribute_5", target); got.Code == "" {
		t.Fatal("5%% 納貢提議失敗")
	}
	state := s.TreatyFor(target)
	if state.PlayerTribute != TributeFivePercent || state.AITribute != TributeNone {
		t.Fatalf("納貢方向或模式錯誤: %+v", state)
	}
	before := s.Player.BC
	s.DisableEvents = true
	s.EndTurn()
	gross := empireGrossBC(s.LastPlayerOutput)
	wantCost := tributeCost(gross, 0, TributeFivePercent)
	if s.LastPlayerOutput.TributeCost != wantCost {
		t.Fatalf("納貢成本 = %d,預期 %d(毛收入 %d)", s.LastPlayerOutput.TributeCost, wantCost, gross)
	}
	if s.Player.BC != before+s.LastPlayerOutput.NetBC {
		t.Fatalf("納貢後玩家國庫與回合摘要不一致: BC=%d before=%d net=%d",
			s.Player.BC, before, s.LastPlayerOutput.NetBC)
	}
	if got := s.DiplomacyResponse("break_tribute", target); got.Code == "" {
		t.Fatal("終止納貢失敗")
	}
	if state := s.TreatyFor(target); state.PlayerTribute != TributeNone || state.AITribute != TributeNone {
		t.Fatalf("終止納貢後狀態未清空: %+v", state)
	}
}

func TestTreatyIncomeAdvancesWithEmpireTurn(t *testing.T) {
	s := NewDemoSession()
	target := s.AIPlayers[0].Name
	// 固定雙方人口，讓測試只觀察協議進度，不依賴人口成長調校值。
	s.PlayerColonies[0].Population = 8
	s.AIPlayers[0].Colonies[0].Population = 8
	if got := s.DiplomacyResponse("trade", target); got.Code == "" {
		t.Fatal("貿易提議失敗")
	}
	if got := s.DiplomacyResponse("research", target); got.Code == "" {
		t.Fatal("研究提議失敗")
	}
	before := s.TreatyFor(target)
	if before.PlayerTradeValue >= 0 || before.PlayerResearchValue >= 0 {
		t.Fatalf("協議初始值應為負的投入期: %+v", before)
	}

	s.EndTurn()
	after := s.TreatyFor(target)
	if after.TradeTurns != 1 || after.ResearchTurns != 1 {
		t.Fatalf("第一次回合未推進協議回合數: %+v", after)
	}
	if s.LastPlayerOutput.TreatyIncomeBC != after.PlayerTradeValue {
		t.Fatalf("回合摘要貿易收益 = %d,狀態 = %d", s.LastPlayerOutput.TreatyIncomeBC, after.PlayerTradeValue)
	}
	if s.LastPlayerOutput.TreatyResearch != after.PlayerResearchValue {
		t.Fatalf("回合摘要研究收益 = %d,狀態 = %d", s.LastPlayerOutput.TreatyResearch, after.PlayerResearchValue)
	}

	first := after.PlayerTradeValue
	for i := 0; i < 4; i++ {
		s.EndTurn()
	}
	if s.TreatyFor(target).PlayerTradeValue <= first {
		t.Fatalf("貿易協議值應從投入期逐步改善: first=%d now=%d", first, s.TreatyFor(target).PlayerTradeValue)
	}
}

func TestSpecialTradeAdvancesBothKindsAndCanEnd(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	target := s.AIPlayers[0].Name

	if got := s.DiplomacyResponse("special_food", target); got.Code == "" {
		t.Fatal("食物—現金特殊貿易提議失敗")
	}
	before := s.TreatyFor(target).SpecialTrade
	if !before.Active || before.Kind != SpecialTradeFoodForCredits {
		t.Fatalf("食物特殊貿易狀態錯誤: %+v", before)
	}
	s.EndTurn()
	after := s.TreatyFor(target).SpecialTrade
	if after.Turns != 1 || after.PlayerValue <= before.PlayerValue {
		t.Fatalf("食物特殊貿易未推進: before=%+v after=%+v", before, after)
	}
	if s.LastPlayerOutput.TreatyIncomeBC <= 0 {
		t.Fatalf("食物特殊貿易應產生 BC 收益: %+v", s.LastPlayerOutput)
	}

	if got := s.DiplomacyResponse("break_special", target); got.Code == "" {
		t.Fatal("終止食物特殊貿易失敗")
	}
	if state := s.TreatyFor(target).SpecialTrade; state.Active {
		t.Fatalf("終止後特殊貿易仍 active: %+v", state)
	}
	if got := s.DiplomacyResponse("special_research", target); got.Code == "" {
		t.Fatal("研究交換特殊貿易提議失敗")
	}
	s.EndTurn()
	state := s.TreatyFor(target).SpecialTrade
	if state.Kind != SpecialTradeResearchExchange || state.Turns != 1 || state.PlayerValue >= 0 {
		t.Fatalf("研究交換特殊貿易狀態錯誤: %+v", state)
	}
	if s.LastPlayerOutput.TreatyResearch <= 0 {
		t.Fatalf("研究交換特殊貿易應產生研究收益: %+v", s.LastPlayerOutput)
	}
}

func TestTreatyStateSurvivesSaveRestore(t *testing.T) {
	s := NewDemoSession()
	target := s.AIPlayers[0].Name
	s.DiplomacyResponse("trade", target)
	s.DiplomacyResponse("alliance", target)
	s.EndTurn()
	want := s.TreatyFor(target)

	restored := s.snapshot().restore()
	got := restored.TreatyFor(target)
	if got != want {
		t.Fatalf("存檔還原後協議不同:\n got=%+v\nwant=%+v", got, want)
	}
}

func TestFormalTreatyBlocksAIOffensivePolicy(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Treaty.FormalPolicy = gamedata.DIPLO_ALLIANCE
	if !a.Treaty.BlocksOffensive() {
		t.Fatal("同盟應禁止攻勢")
	}
	a.StanceName = stanceNames[ai.StanceWar]
	if s.aiRaidWilling(0) {
		t.Fatal("同盟狀態下 AI 不應通過突襲守門")
	}
	if got := aiForeignPolicyFor(a); got != gamedata.DiploAlliance {
		t.Fatalf("正式同盟未映射為原版 ForeignPolicy Alliance: %d", got)
	}
}

func TestFormalWarPolicyOverridesRemakeStanceProjection(t *testing.T) {
	for _, tc := range []struct {
		policy gamedata.ForeignPolicy
		want   gamedata.AIForeignPolicy
	}{
		{gamedata.DIPLO_LIMITED_WAR, gamedata.DiploLimitedWar},
		{gamedata.DIPLO_WAR, gamedata.DiploWar},
		{gamedata.DIPLO_TOTAL_WAR, gamedata.DiploTotalWar},
	} {
		a := &AIOpponent{
			Treaty:     TreatyState{FormalPolicy: tc.policy},
			StanceName: stanceNames[ai.StanceNeutral],
			Relation:   40,
		}
		if got := aiForeignPolicyFor(a); got != tc.want {
			t.Errorf("formal policy %d 映射 = %d，預期 %d", tc.policy, got, tc.want)
		}
	}
}
