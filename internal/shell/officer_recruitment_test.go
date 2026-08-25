package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOfficerRecruitChanceMatchesIDAFormula(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers = nil // 本測試只隔離玩家 offer；AI 路徑另有專屬測試。
	s.ApplyRace(raceIndexByEnName(t, "Psilons"))
	s.Turn = 4
	if got := s.officerRecruitChance(); got != 0 {
		t.Fatalf("前五回合不得招募，got %d", got)
	}
	s.Turn = 10
	if got := s.officerRecruitChance(); got != 11 {
		t.Fatalf("從未 offer 的基礎機率應為 Turn+1，got %d", got)
	}
	s.MercLastOfferTurn = 7
	if got := s.officerRecruitChance(); got != 4 {
		t.Fatalf("距上次 offer 三回合時應為 4，got %d", got)
	}

	s.ApplyRace(raceIndexByEnName(t, "Humans"))
	if got := s.officerRecruitChance(); got != 9 {
		t.Fatalf("Charismatic 應再加 5，got %d", got)
	}

	s.Leaders = []Leader{{ID: 20, Name: "名人", Level: 4,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_FAMOUS), Tier: 2}}}}
	// (4 elapsed + 5 charismatic + floor(4*1.5)) / (1 leader + 1) = 7。
	if got := s.officerRecruitChance(); got != 7 {
		t.Fatalf("進階 Famous 與領袖數除數不符，got %d", got)
	}
}

func TestAdvanceMercOffersUsesRandomPrefixAndSinglePendingOffer(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers = nil
	s.ApplyRace(raceIndexByEnName(t, "Psilons"))
	s.EventSeed = 77
	s.Turn = 200 // 機率超過 100，穩定通過百分比 gate。
	s.MercCandidatePool = []Leader{
		{ID: 10, Name: "殖民地甲", Level: 1},
		{ID: 11, Name: "艦長甲", Level: 1, Ship: true},
		{ID: 12, Name: "殖民地乙", Level: 2},
	}
	s.advanceMercOffers()
	if len(s.MercPool) != 1 {
		t.Fatalf("原版只有一個目前 offer，got %+v", s.MercPool)
	}
	first := s.MercPool[0]
	if s.MercLastOfferTurn != 200 {
		t.Fatalf("offer 回合未記錄：%d", s.MercLastOfferTurn)
	}

	// 下一次成功應取代舊 offer，且不可把仍在畫面的同一位再次抽回。
	s.Turn = 400
	s.advanceMercOffers()
	if len(s.MercPool) != 1 {
		t.Fatalf("新 offer 應取代舊 offer，got %+v", s.MercPool)
	}
	if sameLeader(first, s.MercPool[0]) {
		t.Fatalf("候選不可在仍可見時重複：%+v", s.MercPool[0])
	}
}

func TestOfficerRecruitmentRespectsFourSlotsPerType(t *testing.T) {
	s := NewDemoSession()
	s.Turn = 200
	s.Leaders = []Leader{
		{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4},
	}
	s.MercCandidatePool = []Leader{{ID: 20, Name: "第五位殖民地領袖"}}
	s.advanceMercOffers()
	if len(s.MercPool) != 0 {
		t.Fatalf("殖民地領袖滿四席時不得 offer 第五位：%+v", s.MercPool)
	}
}
