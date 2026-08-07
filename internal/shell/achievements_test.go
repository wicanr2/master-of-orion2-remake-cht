package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 四項成就科技都查得出來——擋門理由「remake 無成就追蹤系統」不成立。
//
// 成就在 MOO2 就是科技,而「有沒有研究出來」一直查得到。
func TestAchievementsAreDetectable(t *testing.T) {
	none := engine.PlayerState{}
	for _, tech := range achievementTechs {
		if hasAchievement(none, tech) {
			t.Errorf("什麼都沒研究時不該有 %s", gamedata.TechnologyName(tech))
		}
		got := withTech(t, none, tech)
		if !hasAchievement(got, tech) {
			t.Errorf("研究出來之後應偵測得到 %s", gamedata.TechnologyName(tech))
		}
		// 反面:研究了這一項不該讓另一項也算數。
		for _, other := range achievementTechs {
			if other != tech && hasAchievement(got, other) {
				t.Errorf("研究 %s 不該連帶讓 %s 生效",
					gamedata.TechnologyName(tech), gamedata.TechnologyName(other))
			}
		}
	}
}

// 虛擬實境網路:全帝國士氣 +20%,與政體無關。
func TestVirtualRealityNetworkRaisesMoraleRegardlessOfGovernment(t *testing.T) {
	ps := withTech(t, engine.PlayerState{}, gamedata.TECH_VIRTUAL_REALITY_NETWORK)
	for _, gov := range []gamedata.MoraleGovernmentType{
		gamedata.MoraleGovFeudalism, gamedata.MoraleGovDemocracy, gamedata.MoraleGovUnification,
	} {
		if got := achievementMoralePercent(ps, gov); got != gamedata.MoraleVirtualRealityNetworkBonus {
			t.Errorf("政體 %d 下 VR 網路應 +%d,得到 %d", gov, gamedata.MoraleVirtualRealityNetworkBonus, got)
		}
	}
}

// 心靈學:**只有特定政體**吃得到(手冊列的是獨裁/帝國/封建/邦聯)。
//
// 這一條同時是上面那條的反面對照:如果實作寫成「有科技就一律加」,這裡會紅。
func TestPsionicsMoraleBonusIsGovernmentGated(t *testing.T) {
	ps := withTech(t, engine.PlayerState{}, gamedata.TECH_PSIONICS)
	for _, gov := range []gamedata.MoraleGovernmentType{
		gamedata.MoraleGovDictatorship, gamedata.MoraleGovImperium,
		gamedata.MoraleGovFeudalism, gamedata.MoraleGovConfederation,
	} {
		if got := achievementMoralePercent(ps, gov); got <= 0 {
			t.Errorf("政體 %d 應吃得到心靈學加成,得到 %d", gov, got)
		}
	}
	for _, gov := range []gamedata.MoraleGovernmentType{
		gamedata.MoraleGovDemocracy, gamedata.MoraleGovFederation,
		gamedata.MoraleGovUnification, gamedata.MoraleGovGalacticUnification,
	} {
		if got := achievementMoralePercent(ps, gov); got != 0 {
			t.Errorf("政體 %d 不在手冊清單上,不該加成,得到 %d", gov, got)
		}
	}
}

// 士氣要用**生效政體**:研究出帝國的獨裁玩家該拿帝國那一格。
//
// 帝國比獨裁多 +20% 全帝國士氣,只看 `s.Government` 會永遠拿不到。
func TestMoraleFollowsTheAdvancedGovernmentForm(t *testing.T) {
	s := NewDemoSession()
	s.Government = gamedata.MoraleGovDictatorship
	if got := s.effectiveGovernment(); got != gamedata.MoraleGovDictatorship {
		t.Fatalf("還沒研究進階政體時應為獨裁,得到 %d", got)
	}
	s.recalcAllColonyMorale()
	before := s.PlayerColonies[0].MoralePercent

	s.Player = withTech(t, s.Player, gamedata.TECH_IMPERIUM)
	if got := s.effectiveGovernment(); got != gamedata.MoraleGovImperium {
		t.Fatalf("研究出帝國後應為帝國,得到 %d", got)
	}
	s.recalcAllColonyMorale()
	if after := s.PlayerColonies[0].MoralePercent; after <= before {
		t.Errorf("帝國比獨裁多 +20%% 全帝國士氣,應提高:%d → %d", before, after)
	}
}

// 微晶構築:每個工業工人 +1 產能——是 per-worker 不是殖民地固定值。
func TestMicroliteConstructionIsPerWorkerNotFlat(t *testing.T) {
	base := engine.ColonyState{
		Population: 8, PopMax: 12, Farmers: 2, Workers: 6,
		FoodPerFarmer: 2, IndustryPerWorker: 3, ResearchPerScientist: 2,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT,
	}
	boosted := base
	boosted.IndustryPerWorkerBonus = gamedata.ProdMicroliteConstructionPerWorkerBonus

	got, want := engine.RunColonyTurn(boosted), engine.RunColonyTurn(base)
	if diff := got.GrossIndustry - want.GrossIndustry; diff != base.Workers {
		t.Errorf("6 個工人各 +1 應多 6 毛產能,實際多 %d(%d vs %d)",
			diff, got.GrossIndustry, want.GrossIndustry)
	}
}

// 奈米分解者:污染容忍值加倍 → 清理成本下降、淨工業上升。
func TestNanoDisassemblersDoublePollutionTolerance(t *testing.T) {
	base := engine.ColonyState{
		Population: 10, PopMax: 14, Farmers: 2, Workers: 8,
		FoodPerFarmer: 4, IndustryPerWorker: 5, ResearchPerScientist: 2,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT,
	}
	clean := base
	clean.NanoDisassemblers = true

	dirty, tidy := engine.RunColonyTurn(base), engine.RunColonyTurn(clean)
	if dirty.PollutionCleanupCost == 0 {
		t.Fatalf("測試前提不成立:沒有奈米分解者時應該有清理成本,得到 %d", dirty.PollutionCleanupCost)
	}
	if tidy.PollutionCleanupCost >= dirty.PollutionCleanupCost {
		t.Errorf("容忍值加倍後清理成本應下降:%d → %d", dirty.PollutionCleanupCost, tidy.PollutionCleanupCost)
	}
	if tidy.NetIndustry <= dirty.NetIndustry {
		t.Errorf("少清一點污染,淨工業應上升:%d → %d", dirty.NetIndustry, tidy.NetIndustry)
	}
}

// 同步是冪等的,而且回合結算真的有呼叫到。
//
// 成就每回合重算而不是「完工時設一次」——科技會被偷、被交換,建築不會。
func TestAchievementSyncIsIdempotentAndWiredIntoEndTurn(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Player = withTech(t, s.Player, gamedata.TECH_NANO_DISASSEMBLERS)

	s.EndTurn() // 應該在結算前就同步好
	if !s.PlayerColonies[0].NanoDisassemblers {
		t.Error("回合結算應把奈米分解者同步進殖民地")
	}
	// 再同步幾次不該改變結果(冪等)。
	for i := 0; i < 3; i++ {
		s.syncAchievementColonyFields()
	}
	if !s.PlayerColonies[0].NanoDisassemblers {
		t.Error("重複同步不該把效果弄丟")
	}

	// 科技沒了就該收回去——這正是「不能像建築那樣設一次」的理由。
	s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	s.syncAchievementColonyFields()
	if s.PlayerColonies[0].NanoDisassemblers {
		t.Error("科技不在了,效果也該收回")
	}
}

// 微晶構築與奈米分解者**互斥**——它們是同一個三選一主題(TOPIC 53,選項 108/113/203)。
//
// ⚠ 這是寫測試時撞到的:原本想一次給玩家兩項,結果後設的那項把前一項的
// `ChosenTech` 蓋掉了。**不是實作的 bug,是科技樹本來就不准。** 釘住它,
// 免得日後有人「順手」讓兩項同時生效。
func TestMicroliteAndNanoAreMutuallyExclusive(t *testing.T) {
	topicA, _ := gamedata.OrigTechTopic(gamedata.TECH_MICROLITE_CONSTRUCTION)
	topicB, _ := gamedata.OrigTechTopic(gamedata.TECH_NANO_DISASSEMBLERS)
	if topicA != topicB {
		t.Fatalf("測試前提不成立:兩項應在同一個主題,得到 %v / %v", topicA, topicB)
	}
	if len(gamedata.ResearchChoiceFor(topicA).Choices) < 2 {
		t.Fatal("測試前提不成立:該主題應是多選")
	}
	ps := withTech(t, engine.PlayerState{}, gamedata.TECH_MICROLITE_CONSTRUCTION)
	if !hasAchievement(ps, gamedata.TECH_MICROLITE_CONSTRUCTION) {
		t.Error("選了微晶構築就該有微晶構築")
	}
	if hasAchievement(ps, gamedata.TECH_NANO_DISASSEMBLERS) {
		t.Error("同一主題的另一個選項不該連帶生效")
	}
}

// 微晶構築真的接進回合結算(與奈米分解者互斥,所以另開一局測)。
func TestMicroliteConstructionSyncsIntoColonies(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Player = withTech(t, s.Player, gamedata.TECH_MICROLITE_CONSTRUCTION)
	s.EndTurn()
	if got := s.PlayerColonies[0].IndustryPerWorkerBonus; got != gamedata.ProdMicroliteConstructionPerWorkerBonus {
		t.Errorf("每工人加成應為 %d,得到 %d",
			gamedata.ProdMicroliteConstructionPerWorkerBonus, got)
	}
	if s.PlayerColonies[0].NanoDisassemblers {
		t.Error("選了微晶構築不該同時拿到奈米分解者")
	}
}
