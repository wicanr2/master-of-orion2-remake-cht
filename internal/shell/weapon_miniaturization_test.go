package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestSessionWeaponMiniaturizationAffectsDesignAndModGate(t *testing.T) {
	s := &GameSession{}
	laser := 1
	tech := WeaponOptions[laser].UnlockTech
	topic, ok := gamedata.OrigTechTopic(tech)
	if !ok {
		t.Fatal("雷射科技缺少主題")
	}
	s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{topic: true}
	base := s.DesignCostWithTech("巡洋艦", laser, 0, 0, 0, nil, gamedata.ARC_FWD)
	if containsWeaponMod(s.WeaponModOptionsForPlayer(laser), gamedata.ModAutoFire) {
		t.Fatal("剛解鎖雷射不應提供 AF")
	}
	next := gamedata.OrigTopicNext[int(topic)]
	s.Player.CompletedTopics[gamedata.ResearchTopic(next)] = true
	mini := s.DesignCostWithTech("巡洋艦", laser, 0, 0, 0, nil, gamedata.ARC_FWD)
	if mini >= base {
		t.Fatalf("完成後續科技後成本未下降: %d >= %d", mini, base)
	}
	if !containsWeaponMod(s.WeaponModOptionsForPlayer(laser), gamedata.ModNoRangeDissipation) {
		t.Fatal("一級微型化應提供 NR")
	}
}
