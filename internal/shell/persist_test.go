package shell

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestSaveLoadRoundTrip 驗證存檔→讀檔後對局狀態一致,且讀回的 AI 可續跑、事件/成長系統續行。
func TestSaveLoadRoundTrip(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	// 推進若干回合並選種族,製造非初始狀態。
	s.ApplyRace(3) // 克拉肯
	s.SetCustomRaceUnusedPicks(5)
	s.LuckyEventCounter = 37
	s.EventLastTurn = 29
	s.EventAttemptCounter = 5
	for i := 0; i < 12; i++ {
		s.EndTurn()
	}
	// 強制走一次 Uncreative 的多選研究完成，驗證研究亂數流位置也進存檔。
	s.ApplyCustomRaceBonuses(Race{}, gamedata.TRAIT_UNCREATIVE)
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY
	s.Player.ResearchTopic = topic
	s.Player.ResearchProgress = gamedata.ResearchChoiceFor(topic).Cost
	s.EndTurn()
	if s.researchRand == nil || s.researchRand.Draws() == 0 {
		t.Fatal("存檔前應已有研究亂數抽取")
	}
	// 負成長 reservoir sampling 使用獨立長壽命亂數流，也必須保存抽取位置。
	s.populationRandForTurn().Intn(3)
	// 協議餘數補點也使用獨立長壽命亂數流。
	s.agreementRandForTurn().Intn(5)
	// 領袖招募擲骰與候選抽取使用另一條獨立亂數流。
	s.officerRandForTurn().Intn(100)
	s.OfficerCooldowns = map[int]int{44: 12}
	if len(s.AIPlayers) > 0 {
		s.AIPlayers[0].LuckyEventCounter = 91
		s.AIPlayers[0].OriginalFoodDeficitTurns = -32768
		s.AIPlayers[0].OriginalWarFlag60ERaw = 1
		s.AIPlayers[0].OriginalBlockadeGrievanceRaw = -7
		s.AIPlayers[0].OriginalHumanBetrayalRaw = true
		s.AIPlayers[0].OriginalRaw28 = 2
		s.AIPlayers[0].OriginalRaw28Known = true
		s.AIPlayers[0].OriginalHumanTreatyGrievanceRaw = -20
		s.AIPlayers[0].OriginalHumanTreatyVictimRaw = 3
		s.AIPlayers[0].OriginalHumanTreatyVictimKnown = true
		s.AIPlayers[0].OriginalHumanIncidentMemoryRaw = 3
		s.AIPlayers[0].OriginalHumanIncidentReasonRaw = 7
		s.AIPlayers[0].OriginalHumanIncidentKnown = true
		offer := aiPreferredLeader(45, false)
		s.AIPlayers[0].LeaderOffer = &offer
		s.AIPlayers[0].LeaderLastOfferTurn = s.Turn
		s.AIPlayers[0].ColonyLeaderNames = []string{"AI 管理官"}
	}
	s.SendFleet(5)
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Progress: 20, Cost: 60}
	// 新增的殖民地種族旗標也必須跟著 remake 存檔保存,不能只在當回合衍生。
	s.PlayerColonies[0].Lithovore = true
	s.PlayerColonies[0].Cybernetic = true
	s.Builds[0].ProgressHalf = 1

	path := filepath.Join(t.TempDir(), "save.json")
	if err := s.Save(path); err != nil {
		t.Fatalf("存檔失敗: %v", err)
	}
	if !SaveExists(path) {
		t.Fatal("存檔應存在")
	}

	got, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀檔失敗: %v", err)
	}

	// 逐項核對關鍵狀態。
	if got.Turn != s.Turn {
		t.Errorf("Turn 不符:%d vs %d", got.Turn, s.Turn)
	}
	if got.Player.BC != s.Player.BC {
		t.Errorf("BC 不符:%d vs %d", got.Player.BC, s.Player.BC)
	}
	if got.RaceIndex != s.RaceIndex {
		t.Errorf("種族不符:%d vs %d", got.RaceIndex, s.RaceIndex)
	}
	if got.ScoreBaseMultiplierPercent != 150 || got.FinalScore().MultiplierPercent != 150 {
		t.Errorf("種族分數倍率未保留:base=%d final=%d", got.ScoreBaseMultiplierPercent, got.FinalScore().MultiplierPercent)
	}
	if got.LuckyEventCounter != 37 {
		t.Errorf("Lucky 事件計數器未保留：got=%d want=37", got.LuckyEventCounter)
	}
	if got.EventLastTurn != 29 || got.EventAttemptCounter != 5 {
		t.Errorf("一般事件排程未保留：last=%d attempts=%d", got.EventLastTurn, got.EventAttemptCounter)
	}
	if len(got.Stars) != len(s.Stars) {
		t.Errorf("星數不符:%d vs %d", len(got.Stars), len(s.Stars))
	}
	if len(got.PlayerColonies) != len(s.PlayerColonies) ||
		got.PlayerColonies[0].Population != s.PlayerColonies[0].Population {
		t.Errorf("殖民地人口不符:%d vs %d", got.PlayerColonies[0].Population, s.PlayerColonies[0].Population)
	}
	if !got.PlayerColonies[0].Lithovore {
		t.Error("殖民地 Lithovore 旗標應在存讀檔後保留")
	}
	if !got.PlayerColonies[0].Cybernetic {
		t.Error("殖民地 Cybernetic 旗標應在存讀檔後保留")
	}
	if len(got.PlayerColonies[0].PopulationGroups) == 0 ||
		len(got.PlayerColonies[0].PopulationGroups) != len(s.PlayerColonies[0].PopulationGroups) {
		t.Errorf("typed population groups 應在 JSON 往返：got=%+v want=%+v",
			got.PlayerColonies[0].PopulationGroups, s.PlayerColonies[0].PopulationGroups)
	}
	if got.researchRand == nil || got.researchRand.Draws() != s.researchRand.Draws() {
		t.Errorf("研究亂數流位置未保留:%v vs %v", got.researchRand, s.researchRand)
	}
	if got.populationRand == nil || got.populationRand.Draws() != s.populationRand.Draws() {
		t.Errorf("人口亂數流位置未保留:%v vs %v", got.populationRand, s.populationRand)
	}
	if got.agreementRand == nil || got.agreementRand.Draws() != s.agreementRand.Draws() {
		t.Errorf("協議亂數流位置未保留:%v vs %v", got.agreementRand, s.agreementRand)
	}
	if got.officerRand == nil || got.officerRand.Draws() != s.officerRand.Draws() {
		t.Errorf("領袖招募亂數流位置未保留:%v vs %v", got.officerRand, s.officerRand)
	}
	if got.OfficerCooldowns[44] != 12 {
		t.Errorf("AI 拒絕 cooldown 未保留：%v", got.OfficerCooldowns)
	}
	if len(got.AIPlayers) == 0 || got.AIPlayers[0].LeaderOffer == nil ||
		got.AIPlayers[0].LeaderOffer.ID != 45 || got.AIPlayers[0].LeaderLastOfferTurn != s.Turn ||
		len(got.AIPlayers[0].ColonyLeaderNames) != 1 {
		t.Errorf("AI 領袖 offer／任命狀態未保留：%+v", got.AIPlayers)
	}
	if got.AIPlayers[0].LuckyEventCounter != 91 {
		t.Errorf("AI Lucky 事件計數器未保留：got=%d want=91", got.AIPlayers[0].LuckyEventCounter)
	}
	if got.AIPlayers[0].OriginalFoodDeficitTurns != -32768 || got.AIPlayers[0].OriginalWarFlag60ERaw != 1 ||
		got.AIPlayers[0].OriginalBlockadeGrievanceRaw != -7 || !got.AIPlayers[0].OriginalHumanBetrayalRaw ||
		!got.AIPlayers[0].OriginalRaw28Known || got.AIPlayers[0].OriginalRaw28 != 2 ||
		got.AIPlayers[0].OriginalHumanTreatyGrievanceRaw != -20 ||
		!got.AIPlayers[0].OriginalHumanTreatyVictimKnown || got.AIPlayers[0].OriginalHumanTreatyVictimRaw != 3 ||
		!got.AIPlayers[0].OriginalHumanIncidentKnown || got.AIPlayers[0].OriginalHumanIncidentMemoryRaw != 3 ||
		got.AIPlayers[0].OriginalHumanIncidentReasonRaw != 7 {
		t.Errorf("AI 原版宣戰 raw 欄位未保留：food=%d raw60E=%d blockade6BF=%d betrayal727=%v",
			got.AIPlayers[0].OriginalFoodDeficitTurns, got.AIPlayers[0].OriginalWarFlag60ERaw,
			got.AIPlayers[0].OriginalBlockadeGrievanceRaw, got.AIPlayers[0].OriginalHumanBetrayalRaw)
	}
	if got.Player.ResearchApplication != s.Player.ResearchApplication ||
		got.Player.HasResearchApplication != s.Player.HasResearchApplication {
		t.Errorf("目前研究 application 未保留:got=(%v,%v) want=(%v,%v)",
			got.Player.ResearchApplication, got.Player.HasResearchApplication,
			s.Player.ResearchApplication, s.Player.HasResearchApplication)
	}
	if got.Fleet().DestStar != s.Fleet().DestStar || got.Fleet().ETA != s.Fleet().ETA {
		t.Errorf("艦隊航行狀態不符")
	}
	if got.Builds[0].Name != s.Builds[0].Name || got.Builds[0].Progress != s.Builds[0].Progress {
		t.Errorf("建造佇列不符")
	}
	if got.Builds[0].ProgressHalf != s.Builds[0].ProgressHalf {
		t.Errorf("半單位建造進度不符:%d vs %d", got.Builds[0].ProgressHalf, s.Builds[0].ProgressHalf)
	}
	if len(got.AIPlayers) != len(s.AIPlayers) {
		t.Fatalf("AI 數不符:%d vs %d", len(got.AIPlayers), len(s.AIPlayers))
	}
	if got.AIPlayers[0].Decider == nil {
		t.Fatal("讀回的 AI Decider 應重建,不為 nil")
	}

	// 讀回的對局應可續跑一回合而不 panic。
	got.EndTurn()
	if got.Turn != s.Turn+1 {
		t.Errorf("讀回對局續跑後 Turn 應 +1:%d", got.Turn)
	}
	raceName := "客製種族"
	if got.RaceIndex >= 0 && got.RaceIndex < len(Races) {
		raceName = Races[got.RaceIndex].Name
	}
	t.Logf("存讀檔往返一致(Turn %d、BC %d、種族 %s、%d 星)", got.Turn-1, got.Player.BC, raceName, len(got.Stars))
}

// TestLoadRejectsMissing 驗證讀取不存在的檔回傳錯誤。
func TestLoadRejectsMissing(t *testing.T) {
	if _, err := LoadSession(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("讀取不存在的存檔應回傳錯誤")
	}
}

func TestRestoreMigratesLegacyAssimilationTurnCounter(t *testing.T) {
	s := NewDemoSession()
	s.ApplyCustomRaceBonuses(Race{OrigIdx: -1})
	s.Government = gamedata.MoraleGovUnification
	snap := s.snapshot()
	snap.AssimilationProgressVersion = 0
	snap.PlayerColonies[0].UnassimilatedPop = 1
	snap.PlayerColonies[0].AssimilationProgress = 3
	got := snap.restore()
	if got.AssimilationProgressVersion != 1 || got.PlayerColonies[0].AssimilationProgress != 36 {
		t.Fatalf("舊 JSON 的 3 回合應遷移為統一政體 36 raw 點，version/progress=%d/%d",
			got.AssimilationProgressVersion, got.PlayerColonies[0].AssimilationProgress)
	}
	if roundTrip := got.snapshot(); roundTrip.AssimilationProgressVersion != 1 ||
		roundTrip.PlayerColonies[0].AssimilationProgress != 36 {
		t.Fatalf("新格式快照應保存 raw 點，got version/progress=%d/%d",
			roundTrip.AssimilationProgressVersion, roundTrip.PlayerColonies[0].AssimilationProgress)
	}
}
