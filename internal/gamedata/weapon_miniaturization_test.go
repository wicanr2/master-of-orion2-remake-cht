package gamedata

import "testing"

func TestWeaponMiniaturizationLevelStartsAtNextTopic(t *testing.T) {
	topic, ok := OrigTechTopic(TECH_LASER_CANNON)
	if !ok {
		t.Fatal("雷射科技缺少主題")
	}
	next := OrigTopicNext[int(topic)]
	if got := WeaponMiniaturizationLevel(TECH_LASER_CANNON, map[ResearchTopic]bool{topic: true}); got != 0 {
		t.Fatalf("剛解鎖武器應為 0 級，got %d", got)
	}
	if got := WeaponMiniaturizationLevel(TECH_LASER_CANNON, map[ResearchTopic]bool{
		topic: true, ResearchTopic(next): true,
	}); got != 1 {
		t.Fatalf("完成下一主題應為 1 級，got %d", got)
	}
}
