package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func mixedGroupFixture() ColonyState {
	return ColonyState{
		Population: 3, Farmers: 1, Workers: 1, Scientists: 1,
		OwnerRaceProfileKnown: true, OwnerRaceSlot: 0, OwnerRaceSlotKnown: true,
		PopulationGroups: []PopulationGroup{
			{RaceSlot: 0, RaceSlotKnown: true, Farmers: 1, Gravity: gamedata.NORMAL_G, ProfileKnown: true},
			{RaceSlot: 1, RaceSlotKnown: true, Workers: 1, PrisonerWorkers: 1, Scientists: 1,
				Gravity: gamedata.NORMAL_G, ProfileKnown: true},
		},
	}
}

func TestPopulationGroupsMutationChainKeepsRaceAndJobs(t *testing.T) {
	c := mixedGroupFixture()
	if !ShiftPopulationGroupJob(&c, gamedata.WORKER, gamedata.FARMER) {
		t.Fatal("應能改派外族 prisoner 工人")
	}
	c.Workers--
	c.Farmers++
	if c.PopulationGroups[1].PrisonerWorkers != 0 || c.PopulationGroups[1].PrisonerFarmers != 1 {
		t.Fatalf("改派應保留 race 與 prisoner：%+v", c.PopulationGroups[1])
	}
	MarkPopulationGroupsPrisoners(&c)
	if !AssimilateOnePopulationGroup(&c) || c.PopulationGroups[0].PrisonerFarmers != 0 {
		t.Fatalf("同化應先清第一個農夫但不改 race：%+v", c.PopulationGroups)
	}
	if !RemovePopulationGroupUnit(&c, gamedata.FARMER) {
		t.Fatal("傷亡應能從農夫群組扣人")
	}
	c.Farmers--
	c.Population--
	if !PopulationGroupsComplete(c) {
		t.Fatalf("傷亡後群組應仍完整：%+v", c)
	}
}
