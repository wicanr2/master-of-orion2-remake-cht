package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestCurrentAreaTopicKeepsHyperAdvancedSelectable(t *testing.T) {
	s := shell.NewDemoSession()
	s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	physics := gamedata.TechTree()[2]
	for _, topic := range physics {
		s.Player.CompletedTopics[topic] = true
	}
	last := physics[len(physics)-1]
	s.Player.HyperAdvancedLevels = map[gamedata.ResearchTopic]int{last: 2}
	got, cost, done := currentAreaTopic(s, 2)
	if got != last || done || cost <= 0 {
		t.Fatalf("完成物理領域後應可選下一級 Hyper: topic=%d cost=%d done=%v", got, cost, done)
	}
}
