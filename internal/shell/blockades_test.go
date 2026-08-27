package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func setBlockadeTestPlayerSlot(s *GameSession, slot int) {
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].OwnerRaceSlot = slot
		s.PlayerColonies[i].OwnerRaceSlotKnown = true
		s.PlayerColonies[i].PopulationGroups = []engine.PopulationGroup{{
			RaceSlot: slot, RaceSlotKnown: true, Workers: s.PlayerColonies[i].Population,
		}}
	}
}

func TestRecomputeOriginalBlockadesPlayerAndAI(t *testing.T) {
	s := NewDemoSession()
	setBlockadeTestPlayerSlot(s, 0)
	if len(s.AIPlayers) == 0 || len(s.PlayerColonyStars) == 0 {
		t.Fatal("demo 必須有玩家殖民地與 AI")
	}
	star := s.PlayerColonyStars[0]
	ai := &s.AIPlayers[0]
	ai.Treaty.FormalPolicy = gamedata.DIPLO_WAR
	ai.FleetPosSet, ai.FleetStar, ai.FleetDestStar, ai.FleetETA = true, star, -1, 0
	ai.Ships = []Ship{{Name: "BLOCKADER"}}
	s.Stars[star].BlockadedMask = 0xFF
	s.Stars[star].BlockadedBy = [8]uint8{0xFF, 0xFF}
	s.recomputeOriginalBlockades()
	playerSlot, known := ownerSlotFromColonies(s.PlayerColonies)
	if !known || s.Stars[star].BlockadedMask&(1<<playerSlot) == 0 {
		t.Fatalf("戰爭中已抵達 AI 艦隊應封鎖玩家：star=%+v slot=%d", s.Stars[star], playerSlot)
	}
	if s.Stars[star].BlockadedBy[playerSlot]&(1<<ai.PopulationRaceSlot) == 0 {
		t.Fatalf("逐受害者 blockader mask 未寫入：%08b", s.Stars[star].BlockadedBy[playerSlot])
	}
	ai.Treaty.FormalPolicy = gamedata.DIPLO_PEACE
	s.recomputeOriginalBlockades()
	if s.Stars[star].BlockadedMask != 0 || s.Stars[star].BlockadedBy != [8]uint8{} {
		t.Fatalf("每回合應先清表，和平艦隊不得留下舊封鎖：%+v", s.Stars[star])
	}
}

func TestRecomputeOriginalBlockadesIgnoresTransitAndOwnColony(t *testing.T) {
	s := NewDemoSession()
	setBlockadeTestPlayerSlot(s, 0)
	star := s.AIPlayers[0].ColonyStars[0]
	ai := &s.AIPlayers[0]
	ai.Treaty.FormalPolicy = gamedata.DIPLO_WAR
	ai.FleetPosSet, ai.FleetStar, ai.FleetDestStar, ai.FleetETA = true, star, star, 1
	ai.Ships = []Ship{{Name: "TRANSIT"}}
	s.recomputeOriginalBlockades()
	if s.Stars[star].BlockadedMask != 0 {
		t.Fatalf("航行中艦隊不得封鎖：%08b", s.Stars[star].BlockadedMask)
	}
	ai.FleetETA = 0
	s.recomputeOriginalBlockades()
	if s.Stars[star].BlockadedMask&(1<<ai.PopulationRaceSlot) != 0 {
		t.Fatalf("艦隊不得封鎖自己在同星的殖民地：%08b", s.Stars[star].BlockadedMask)
	}
}

func TestRecomputeOriginalBlockadesOwnerEightBlocksAllColonists(t *testing.T) {
	s := NewDemoSession()
	setBlockadeTestPlayerSlot(s, 0)
	star := s.PlayerColonyStars[0]
	s.Monsters = []MonsterGuard{{Kind: gamedata.MonsterDragon, StarIndex: star, TransitETA: 0}}
	s.recomputeOriginalBlockades()
	playerSlot, _ := ownerSlotFromColonies(s.PlayerColonies)
	if s.Stars[star].BlockadedMask&(1<<playerSlot) == 0 {
		t.Fatalf("owner 8 停泊怪物應封鎖同星全部殖民者：%08b", s.Stars[star].BlockadedMask)
	}
	if s.Stars[star].BlockadedBy != [8]uint8{} {
		t.Fatalf("owner 8 分支不應填一般 player blockader mask：%+v", s.Stars[star].BlockadedBy)
	}
}

func TestRecomputeOriginalBlockadesRecordsAIGrievanceWithoutChangingRelation(t *testing.T) {
	s := NewDemoSession()
	setBlockadeTestPlayerSlot(s, 0)
	ai := &s.AIPlayers[0]
	star := ai.ColonyStars[0]
	playerSlot, known := ownerSlotFromColonies(s.PlayerColonies)
	if !known {
		t.Fatal("玩家 slot 必須可知")
	}
	ai.Treaty.FormalPolicy = gamedata.DIPLO_WAR
	// 正關係分支會先把負向 delta 倍增，避免 Charismatic /2 將 Random_(5)=1
	// 截成零；此測試要穩定覆蓋 +0x6BF 的實際寫入。
	ai.OriginalRelationKnown, ai.OriginalRelationRaw = true, 50
	s.Fleets = []Fleet{{AtStar: star, DestStar: -1, ETA: 0, Ships: []Ship{{Name: "HUMAN BLOCKADER"}}}}
	beforeRelation := ai.OriginalRelationRaw
	s.recomputeOriginalBlockades()
	if s.Stars[star].BlockadedBy[ai.PopulationRaceSlot]&(1<<playerSlot) == 0 {
		t.Fatalf("玩家艦隊應出現在 AI 的 blockader mask：%08b", s.Stars[star].BlockadedBy[ai.PopulationRaceSlot])
	}
	if ai.OriginalBlockadeGrievanceRaw >= 0 {
		t.Fatalf("戰時封鎖應累積負向 +0x6BF 積怨：%d", ai.OriginalBlockadeGrievanceRaw)
	}
	if ai.OriginalRelationRaw != beforeRelation {
		t.Fatalf("policy>=4 在 0x4E75C 早退，不得改一般 +0x617：%d -> %d", beforeRelation, ai.OriginalRelationRaw)
	}
}

func TestRecomputeOriginalBlockadesConsumesGrievanceRollWithoutTypedRelation(t *testing.T) {
	s := NewDemoSession()
	setBlockadeTestPlayerSlot(s, 0)
	ai := &s.AIPlayers[0]
	star := ai.ColonyStars[0]
	ai.Treaty.FormalPolicy = gamedata.DIPLO_WAR
	ai.OriginalRelationKnown = false
	s.Fleets = []Fleet{{AtStar: star, DestStar: -1, ETA: 0, Ships: []Ship{{Name: "HUMAN BLOCKADER"}}}}
	s.recomputeOriginalBlockades()
	if s.diplomacyGrowthRand == nil || s.diplomacyGrowthRand.Draws() != 1 {
		t.Fatalf("原版形成封鎖後必須消耗一次 Random_(5)，即使舊存檔沒有 typed 關係：%+v", s.diplomacyGrowthRand)
	}
	if ai.OriginalBlockadeGrievanceRaw != 0 {
		t.Fatalf("未知關係不可猜寫積怨：%d", ai.OriginalBlockadeGrievanceRaw)
	}
}
