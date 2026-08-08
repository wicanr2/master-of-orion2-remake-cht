package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// withTech 回傳一份「已明確研究出 tech」的玩家狀態。
func withTech(t *testing.T, base engine.PlayerState, techs ...gamedata.Technology) engine.PlayerState {
	t.Helper()
	ps := clonePlayerState(base)
	if ps.CompletedTopics == nil {
		ps.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	if ps.ExplicitChoice == nil {
		ps.ExplicitChoice = map[gamedata.ResearchTopic]bool{}
	}
	if ps.ChosenTech == nil {
		ps.ChosenTech = map[gamedata.ResearchTopic]gamedata.Technology{}
	}
	for _, tech := range techs {
		topic, ok := gamedata.OrigTechTopic(tech)
		if !ok {
			t.Fatalf("%s 應查得到所屬研究主題", gamedata.TechnologyName(tech))
		}
		ps.CompletedTopics[topic] = true
		ps.ExplicitChoice[topic] = true
		ps.ChosenTech[topic] = tech
	}
	return ps
}

// 手冊 Spy Bonuses 表的 Technology 列真的有作用了。
//
// 這一項修的擋門理由是「**無逐科技模型可查是否擁有 spy.go 列的 5 項科技**」——
// 而 `groundEquipTechOwned` 已經是三個系統共用的判定,那 5 項在 enums.go 都有常數。
func TestSpyTechnologyBonusIsWired(t *testing.T) {
	none := engine.PlayerState{}
	if got := spyTechBonusFor(none); got != 0 {
		t.Errorf("沒有任何間諜科技時應為 0,得到 %d", got)
	}
	one := withTech(t, none, gamedata.TECH_NEURAL_SCANNER)
	if got := spyTechBonusFor(one); got != 10 {
		t.Errorf("神經掃描器應 +10,得到 %d", got)
	}
	// 5 項互不相關的科技,加總而不是取最佳——取最佳的話研究第二項就完全沒有意義。
	two := withTech(t, none, gamedata.TECH_NEURAL_SCANNER, gamedata.TECH_TELEPATHIC_TRAINING)
	if got := spyTechBonusFor(two); got != 15 {
		t.Errorf("神經掃描器(10)+ 心靈感應訓練(5)應為 15,得到 %d", got)
	}
	// 反面:不在手冊那張表上的科技不該加成。
	other := withTech(t, none, gamedata.TECH_DEATH_SPORES)
	if got := spyTechBonusFor(other); got != 0 {
		t.Errorf("手冊沒列的科技不該加成,得到 %d", got)
	}
}

// 攻守兩側用同一套科技加成(手冊那張表兩欄同值)。
func TestSpyTechnologyBonusAppliesToBothSides(t *testing.T) {
	ps := withTech(t, engine.PlayerState{}, gamedata.TECH_STEALTH_SUIT)
	atk := spyAttackerBonus(ps, 0, 0)
	def := spyDefenderBonus(ps, 0, 0)
	if atk != 10 || def != 10 {
		t.Errorf("攻守兩側都應 +10(表中兩欄同值),得到 攻=%d 守=%d", atk, def)
	}
	// 攻擊側還要加上人數加成——科技那一項不該把它蓋掉。
	if withSpies := spyAttackerBonus(ps, 6, 0); withSpies <= atk {
		t.Errorf("有間諜人數時攻擊加成應更高:%d vs %d", withSpies, atk)
	}
}

// 政府防諜加成走「進階政體」那條路:研究出帝國的獨裁玩家該拿帝國那一格。
//
// 只看 `s.Government` 會永遠停在基本型——而手冊對基本型與進階型給的是不同的值。
func TestPlayerSpyGovernmentBonusFollowsTheAdvancedForm(t *testing.T) {
	s := NewDemoSession()
	s.Government = gamedata.MoraleGovUnification
	base := s.playerSpyGovernmentDefenseBonus()
	if base != gamedata.SpyGovernmentDefenseBonus(gamedata.SpyGovUnification) {
		t.Errorf("統一政體應拿統一那一格,得到 %d", base)
	}

	// 研究出銀河統一之後改拿進階那一格。
	s.Player = withTech(t, s.Player, gamedata.TECH_GALACTIC_UNIFICATION)
	if got := s.playerSpyGovernmentDefenseBonus(); got != gamedata.SpyGovernmentDefenseBonus(gamedata.SpyGovGalacticUnification) {
		t.Errorf("研究出銀河統一後應拿進階那一格,得到 %d", got)
	}
}

// 三個政體列舉的編號必須一致——原版只有一個 `[player+0x89F]`(第 54 項(三個寫入端))。
//
// Go 這邊分成 Assimilation / Morale / Spy 三個列舉是歷史,不是原版有三套編號。
// 少了這條,其中一個被重排時另外兩個不會有任何反應,而政府防諜加成會安靜地查錯格。
func TestSpyGovernmentNumbersMatchTheOtherTwoEnums(t *testing.T) {
	cases := []struct {
		spy    gamedata.SpyGovernmentType
		assim  gamedata.AssimilationGovernment
		morale gamedata.MoraleGovernmentType
	}{
		{gamedata.SpyGovFeudalism, gamedata.AssimFeudal, gamedata.MoraleGovFeudalism},
		{gamedata.SpyGovConfederation, gamedata.AssimConfederation, gamedata.MoraleGovConfederation},
		{gamedata.SpyGovDictatorship, gamedata.AssimDictatorship, gamedata.MoraleGovDictatorship},
		{gamedata.SpyGovImperium, gamedata.AssimImperium, gamedata.MoraleGovImperium},
		{gamedata.SpyGovDemocracy, gamedata.AssimDemocracy, gamedata.MoraleGovDemocracy},
		{gamedata.SpyGovFederation, gamedata.AssimFederation, gamedata.MoraleGovFederation},
		{gamedata.SpyGovUnification, gamedata.AssimUnification, gamedata.MoraleGovUnification},
		{gamedata.SpyGovGalacticUnification, gamedata.AssimGalacticUnification, gamedata.MoraleGovGalacticUnification},
	}
	for _, c := range cases {
		if int(c.spy) != int(c.assim) || int(c.spy) != int(c.morale) {
			t.Errorf("政體編號三邊對不上:spy=%d assim=%d morale=%d", c.spy, c.assim, c.morale)
		}
	}
}

// 端到端:防守方的科技真的讓對方比較難得手。
//
// 用同一組種子跑很多局比計數,而不是單局比大小——單局會被隨機性帶著走。
func TestDefenderTechMakesTheftHarder(t *testing.T) {
	attackerBase := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{},
		ChosenTech:      map[gamedata.ResearchTopic]gamedata.Technology{},
		ExplicitChoice:  map[gamedata.ResearchTopic]bool{},
	}
	// 防守方手上要有東西可偷,否則兩組都偷不到,測不出差別。
	defenderPlain := withTech(t, engine.PlayerState{}, gamedata.TECH_DEATH_SPORES)
	defenderGuarded := withTech(t, defenderPlain,
		gamedata.TECH_NEURAL_SCANNER, gamedata.TECH_CYBERSECURITY_LINK, gamedata.TECH_STEALTH_SUIT)

	count := func(def engine.PlayerState) int {
		hits := 0
		for seed := int64(1); seed <= 400; seed++ {
			atk := clonePlayerState(attackerBase)
			msgs, _ := spyStealAttempt(newRandStream(seed), &atk, def, 4, "我方", "AI", 0, 0, 0)
			for _, m := range msgs {
				if len(m) > 0 && containsSteal(m) {
					hits++
					break
				}
			}
		}
		return hits
	}
	plain, guarded := count(defenderPlain), count(defenderGuarded)
	if plain == 0 {
		t.Fatalf("測試前提不成立:沒有防禦科技時應該偷得到(%d 次)", plain)
	}
	if guarded >= plain {
		t.Errorf("防守方有間諜科技時得手次數不該持平或更高:無防禦 %d vs 有防禦 %d", plain, guarded)
	}
}

func containsSteal(msg string) bool {
	const marker = "偷得科技"
	for i := 0; i+len(marker) <= len(msg); i++ {
		if msg[i:i+len(marker)] == marker {
			return true
		}
	}
	return false
}
