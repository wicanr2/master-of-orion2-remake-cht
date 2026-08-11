package shell

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestHireMercAtSelectsCandidateAndKeepsQueueOrder(t *testing.T) {
	s := NewDemoSession()
	s.Player.BC = 1000
	s.MercPool = []Leader{
		{Name: "殖民地甲", Skill: "科學家", Level: 1, Ship: false, Tier: 1},
		{Name: "艦艇乙", Skill: "工程師", Level: 1, Ship: true, Tier: 1},
	}

	if !s.HireMercAt(1) {
		t.Fatal("應能雇用指定的第二位傭兵")
	}
	if len(s.Leaders) == 0 || s.Leaders[len(s.Leaders)-1].Name != "艦艇乙" {
		t.Fatalf("指定傭兵沒有加入 Leader Pool: %+v", s.Leaders)
	}
	if len(s.MercPool) != 1 || s.MercPool[0].Name != "殖民地甲" {
		t.Fatalf("移除中間候選後佇列順序錯誤: %+v", s.MercPool)
	}
}

func TestReturnShipOfficerToPoolClearsAssignment(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = []Fleet{{Ships: []Ship{{Name: "甲"}}, AtStar: 0, DestStar: -1}}
	s.Leaders = []Leader{{Name: "艦長", Skill: "工程師", Level: 3, Ship: true, Tier: 1}}
	if !s.AssignOfficerToShip(0, 0, 0) {
		t.Fatal("前置指派失敗")
	}

	if !s.ReturnShipOfficerToPool("艦長") {
		t.Fatal("應能將艦艇軍官送回人才庫")
	}
	if got := s.Fleets[0].Ships[0].OfficerName; got != "" {
		t.Fatalf("POOL 後仍有艦艇指派: %q", got)
	}
	if _, ok := s.OfficerForShip(s.Fleets[0].Ships[0]); ok {
		t.Fatal("POOL 後不應再有生效中的艦艇軍官")
	}
}

func TestDismissShipOfficerRemovesAssignmentAndLeader(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = []Fleet{{Ships: []Ship{{Name: "甲"}}, AtStar: 0, DestStar: -1}}
	s.Leaders = []Leader{
		{Name: "保留者", Skill: "工程師", Level: 1, Ship: true, Tier: 1},
		{Name: "解雇者", Skill: "武器官", Level: 2, Ship: true, Tier: 1},
	}
	if !s.AssignOfficerToShip(0, 0, 1) {
		t.Fatal("前置指派失敗")
	}

	if !s.DismissShipOfficer("解雇者") {
		t.Fatal("應能解雇艦艇軍官")
	}
	if len(s.Leaders) != 1 || s.Leaders[0].Name != "保留者" {
		t.Fatalf("解雇後領袖池錯誤: %+v", s.Leaders)
	}
	if got := s.Fleets[0].Ships[0].OfficerName; got != "" {
		t.Fatalf("解雇後艦艇仍留有指派: %q", got)
	}
}

func TestDismissShipOfficerDoesNotRemoveColonyLeader(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = []Leader{{Name: "殖民地領袖", Skill: "科學家", Level: 2, Ship: false, Tier: 1}}
	if s.DismissShipOfficer("殖民地領袖") {
		t.Fatal("艦艇解雇 API 不應刪除殖民地領袖")
	}
	if len(s.Leaders) != 1 {
		t.Fatal("拒絕解雇後殖民地領袖不應消失")
	}
}

func TestColonyLeaderAssignmentIsReversibleAndPersists(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = []Leader{{
		ID: 77, Name: "殖民地科學家", Skill: "科學家", Level: 4, Tier: 2,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_RESEARCHER), Tier: 2}},
	}}
	if len(s.PlayerColonies) == 0 {
		t.Fatal("測試局應有母星殖民地")
	}
	base := s.PlayerColonies[0].FlatResearch
	if !s.AssignLeaderToColony(0, 0) {
		t.Fatal("殖民地領袖指派應成功")
	}
	withLeader := s.PlayerColonies[0].FlatResearch
	if withLeader <= base {
		t.Fatalf("研究領袖應增加殖民地研究，%d→%d", base, withLeader)
	}
	if !s.AssignLeaderToColony(0, 0) || s.PlayerColonies[0].FlatResearch != withLeader {
		t.Fatal("重複指派不應重複累加")
	}
	path := filepath.Join(t.TempDir(), "colony-leader.json")
	if err := s.Save(path); err != nil {
		t.Fatalf("保存殖民地領袖: %v", err)
	}
	loaded, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀回殖民地領袖: %v", err)
	}
	if got, ok := loaded.ColonyLeaderFor(0); !ok || got.Name != "殖民地科學家" {
		t.Fatalf("讀回後殖民地領袖錯誤: %+v/%v", got, ok)
	}
	if !loaded.UnassignLeaderFromColony(0) || loaded.PlayerColonies[0].FlatResearch != base {
		t.Fatalf("解除指派應撤回研究加成，得到 %d，基準 %d", loaded.PlayerColonies[0].FlatResearch, base)
	}
}

func TestShipAssignmentRemovesColonyLeaderRole(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = []Fleet{{Ships: []Ship{{Name: "甲艦"}}, AtStar: 0, DestStar: -1}}
	s.Leaders = []Leader{{ID: 78, Name: "雙職測試", Skill: "工程師", Level: 1, Tier: 1, Ship: true,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_ENGINEER), Tier: 1}}}}
	// 測試資料先以殖民地領袖身分指派，再轉換成艦艇軍官，確認不會殘留雙重職位。
	s.Leaders[0].Ship = false
	if !s.AssignLeaderToColony(0, 0) {
		t.Fatal("前置殖民地指派失敗")
	}
	s.Leaders[0].Ship = true
	if !s.AssignOfficerToShip(0, 0, 0) {
		t.Fatal("轉任艦艇軍官應成功")
	}
	if _, ok := s.AssignedColonyForLeader("雙職測試"); ok {
		t.Fatal("轉任艦艇軍官後不應仍保留殖民地職位")
	}
}
