package shell

import (
	"encoding/json"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestBattleoidsCallbackGrantsArmorBarracksWithoutReplacingChoice(t *testing.T) {
	ps := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{
			gamedata.TOPIC_ASTRO_ENGINEERING:  true,
			gamedata.TOPIC_ASTRO_CONSTRUCTION: true,
		},
		ChosenTech: map[gamedata.ResearchTopic]gamedata.Technology{
			gamedata.TOPIC_ASTRO_ENGINEERING:  gamedata.TECH_FIGHTER_GARRISON,
			gamedata.TOPIC_ASTRO_CONSTRUCTION: gamedata.TECH_BATTLEOIDS,
		},
		ExplicitChoice: map[gamedata.ResearchTopic]bool{
			gamedata.TOPIC_ASTRO_ENGINEERING:  true,
			gamedata.TOPIC_ASTRO_CONSTRUCTION: true,
		},
	}

	applyResearchTopicGrantCallbacks(&ps, gamedata.TOPIC_ASTRO_CONSTRUCTION)
	applyResearchTopicGrantCallbacks(&ps, gamedata.TOPIC_ASTRO_CONSTRUCTION) // 冪等

	if !playerStateKnowsTech(ps, gamedata.TOPIC_ASTRO_ENGINEERING, gamedata.TECH_ARMOR_BARRACKS) {
		t.Fatal("取得 Battleoids 後應連帶知道 Armor Barracks")
	}
	if !playerStateKnowsTech(ps, gamedata.TOPIC_ASTRO_ENGINEERING, gamedata.TECH_FIGHTER_GARRISON) {
		t.Fatal("連帶授予不得清除 Astro Engineering 原選擇")
	}
	if got := ps.ChosenTech[gamedata.TOPIC_ASTRO_ENGINEERING]; got != gamedata.TECH_FIGHTER_GARRISON {
		t.Fatalf("主要選擇遭覆蓋：got %v", got)
	}
	if len(ps.GrantedTechs) != 1 || !ps.GrantedTechs[gamedata.TECH_ARMOR_BARRACKS] {
		t.Fatalf("額外科技集合錯誤：%+v", ps.GrantedTechs)
	}
}

func TestTechTheftAddsSecondApplicationAndRunsCallback(t *testing.T) {
	ps := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_ASTRO_ENGINEERING: true},
		ChosenTech:      map[gamedata.ResearchTopic]gamedata.Technology{gamedata.TOPIC_ASTRO_ENGINEERING: gamedata.TECH_FIGHTER_GARRISON},
		ExplicitChoice:  map[gamedata.ResearchTopic]bool{gamedata.TOPIC_ASTRO_ENGINEERING: true},
	}
	applyTechTheft(&ps, spyStealOption{Topic: gamedata.TOPIC_ASTRO_CONSTRUCTION, Tech: gamedata.TECH_BATTLEOIDS})

	if !playerStateKnowsTech(ps, gamedata.TOPIC_ASTRO_CONSTRUCTION, gamedata.TECH_BATTLEOIDS) ||
		!playerStateKnowsTech(ps, gamedata.TOPIC_ASTRO_ENGINEERING, gamedata.TECH_ARMOR_BARRACKS) {
		t.Fatalf("偷得 Battleoids 後的授予鏈不完整：%+v", ps)
	}
	if ps.ChosenTech[gamedata.TOPIC_ASTRO_ENGINEERING] != gamedata.TECH_FIGHTER_GARRISON {
		t.Fatal("callback 不得覆蓋另一主題既有選擇")
	}
}

func TestGrantedTechsJSONRoundTrip(t *testing.T) {
	want := engine.PlayerState{GrantedTechs: map[gamedata.Technology]bool{gamedata.TECH_ARMOR_BARRACKS: true}}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got engine.PlayerState
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if !got.GrantedTechs[gamedata.TECH_ARMOR_BARRACKS] {
		t.Fatalf("JSON 往返遺失額外科技：%s", b)
	}
}
