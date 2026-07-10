package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestComponentGatingExplicitChoice:明確抉擇某主題後,僅所選科技對應元件解鎖(其餘同主題元件不解)。
// 未明確抉擇則主題層級解鎖(非破壞)。
func TestComponentGatingExplicitChoice(t *testing.T) {
	// ADVANCED_MAGNETISM 選項 {Class I Shield, ECM Jammer, Mass Driver};
	// remake 元件:質量投射器(UnlockTech=MASS_DRIVER)、第一級護盾(UnlockTech=CLASS_I_SHIELD)同掛此主題。
	topic := gamedata.TOPIC_ADVANCED_MAGNETISM
	find := func(opts []Component, name string) Component {
		for _, c := range opts {
			if c.Name == name {
				return c
			}
		}
		t.Fatalf("找不到元件 %s", name)
		return Component{}
	}
	massDriver := find(WeaponOptions, "質量投射器")
	shield1 := find(ShieldOptions, "第一級護盾")
	if massDriver.Tech != topic || shield1.Tech != topic {
		t.Fatalf("前提:兩元件應同掛 %v", topic)
	}

	// (A) 完成主題但未明確抉擇 → 兩者皆解鎖(主題層級,非破壞)。
	s := NewDemoSession()
	s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{topic: true}
	if !s.ComponentUnlocked(massDriver) || !s.ComponentUnlocked(shield1) {
		t.Fatalf("未明確抉擇應主題層級解鎖兩者")
	}

	// (B) 明確抉擇 MASS_DRIVER → 質量投射器解鎖、第一級護盾不解。
	s2 := NewDemoSession()
	s2.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{topic: true}
	s2.Player.ChosenTech = map[gamedata.ResearchTopic]gamedata.Technology{topic: gamedata.TECH_MASS_DRIVER}
	s2.Player.ExplicitChoice = map[gamedata.ResearchTopic]bool{topic: true}
	if !s2.ComponentUnlocked(massDriver) {
		t.Fatalf("明確選 MASS_DRIVER 應解鎖質量投射器")
	}
	if s2.ComponentUnlocked(shield1) {
		t.Fatalf("明確選 MASS_DRIVER 不應解鎖第一級護盾")
	}
}
