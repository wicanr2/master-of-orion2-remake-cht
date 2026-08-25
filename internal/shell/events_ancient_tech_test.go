package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func techState(techs ...gamedata.Technology) engine.PlayerState {
	ps := engine.PlayerState{}
	for _, tech := range techs {
		topic, ok := gamedata.OrigTechTopic(tech)
		if ok {
			grantTechnologyApplication(&ps, topic, tech)
		}
	}
	return ps
}

func TestOriginalAncientTechChoosesBeamAndHighestShield(t *testing.T) {
	target := techState(gamedata.TECH_LASER_CANNON)
	benchmark := techState(gamedata.TECH_PLASMA_CANNON, gamedata.TECH_CLASS_V_SHIELD)
	got := originalAncientTechApplications(target, []engine.PlayerState{target, benchmark})
	if len(got) != 2 {
		t.Fatalf("應授予光束與護盾兩項，got %v", got)
	}
	// Plasma Cannon 最大傷害 30；下一個最小非負差候選由 raw 武器表／tie 規則決定。
	if weapon, ok := gamedata.OrigWeaponByTech(got[0]); !ok || weapon.Cat != gamedata.WeaponCatBeam ||
		weapon.DamageMax < 30 || weapon.DamageMax > 80 {
		t.Fatalf("第一項不是基準 +0..50 的光束：tech=%v weapon=%+v", got[0], weapon)
	}
	if got[1] != gamedata.TECH_CLASS_V_SHIELD {
		t.Fatalf("第二項=%v，want Class V Shield", got[1])
	}
}

func TestOriginalAncientTechSkipsKnownAndDamageDifferenceOverFifty(t *testing.T) {
	// 目標已知 raw 0..39 的全部可研究光束與最高護盾，故不應重複授予。
	known := []gamedata.Technology{gamedata.TECH_CLASS_X_SHIELD}
	for id := 0; id < 40 && id < len(gamedata.OrigWeaponTable); id++ {
		weapon := gamedata.OrigWeaponTable[id]
		if weapon.Cat == gamedata.WeaponCatBeam && weapon.Tech != gamedata.TECH_NONE {
			known = append(known, weapon.Tech)
		}
	}
	target := techState(known...)
	got := originalAncientTechApplications(target, []engine.PlayerState{target})
	if len(got) != 0 {
		t.Fatalf("沒有未知合法候選時應為空，got %v", got)
	}
}

func TestAncientTechEventGrantsApplicationWithoutResearchProgress(t *testing.T) {
	s := NewDemoSession()
	s.Player = techState(gamedata.TECH_LASER_CANNON)
	s.Player.ResearchProgress = 321
	s.AIPlayers[0].Player = techState(gamedata.TECH_PLASMA_CANNON, gamedata.TECH_CLASS_III_SHIELD)
	ev := *gamedata.RandomEventByID(0)
	result, ok := s.applyRandomEventLocalized(ev)
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("古代科技應授予 application：result=%+v ok=%v", result, ok)
	}
	if s.Player.ResearchProgress != 321 {
		t.Fatalf("事件不得再增加 RP：got %d", s.Player.ResearchProgress)
	}
	if len(s.Player.GrantedTechs) == 0 && len(s.Player.CompletedTopics) <= 1 {
		t.Fatalf("沒有寫入任何科技 application：player=%+v", s.Player)
	}
}

func TestAIAncientTechEventWritesOnlyAI(t *testing.T) {
	s := NewDemoSession()
	s.Player = techState(gamedata.TECH_PLASMA_CANNON, gamedata.TECH_CLASS_V_SHIELD)
	s.AIPlayers[0].Player = techState(gamedata.TECH_LASER_CANNON)
	s.AIPlayers[0].Player.ResearchProgress = 222
	beforePlayerTopics := len(s.Player.CompletedTopics)
	ev := *gamedata.RandomEventByID(0)
	result, ok := s.applyRandomEventLocalizedToAI(ev, 0)
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("AI 古代科技應可結算：result=%+v ok=%v", result, ok)
	}
	if s.AIPlayers[0].Player.ResearchProgress != 222 {
		t.Fatalf("AI 事件不得修改 RP：got %d", s.AIPlayers[0].Player.ResearchProgress)
	}
	if len(s.Player.CompletedTopics) != beforePlayerTopics {
		t.Fatal("AI 事件不得授予目前玩家科技")
	}
}
