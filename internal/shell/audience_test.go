package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestAudienceOnlyOnStanceChange 釘住觸發規則:**態勢改變**才敲門。
//
// 特別驗「開局第一次算出態勢不算改變」——否則每個新對局一開始就有一整排燈亮著。
func TestAudienceOnlyOnStanceChange(t *testing.T) {
	a := &AIOpponent{StanceName: "宣戰"}
	a.noteStanceChange("") // 開局:prev 為空
	if a.WantsAudience {
		t.Error("開局第一次算出態勢不該觸發會談請求")
	}

	a = &AIOpponent{StanceName: "宣戰"}
	a.noteStanceChange("宣戰") // 沒變
	if a.WantsAudience {
		t.Error("態勢沒變不該觸發")
	}

	a = &AIOpponent{StanceName: "宣戰"}
	a.noteStanceChange("中立") // 中立 → 宣戰
	if !a.WantsAudience || a.AudienceReason != AudienceReasonWar {
		t.Errorf("中立→宣戰應觸發且來意為 %s,實得 %v/%q",
			AudienceReasonWar, a.WantsAudience, a.AudienceReason)
	}
}

// TestAudienceStanceMapping 逐級核對哪些態勢會敲門。
//
// 中立/敵視不敲門:前者沒事,後者是態度不是提案。
func TestAudienceStanceMapping(t *testing.T) {
	cases := map[string]string{
		"宣戰":   AudienceReasonWar,
		"提議貿易": AudienceReasonTrade,
		"提議結盟": AudienceReasonAlliance,
		"中立":   "",
		"敵視":   "",
	}
	for stance, want := range cases {
		got, ok := audienceReasonForStance(stance)
		if want == "" {
			if ok {
				t.Errorf("態勢 %q 不該敲門,卻回了 %q", stance, got)
			}
			continue
		}
		if !ok || got != want {
			t.Errorf("態勢 %q 的來意應為 %q,實得 %q(ok=%v)", stance, want, got, ok)
		}
	}
}

// TestAudienceReasonsAreCodesNotDisplayText:來意是代碼不是顯示字串。
//
// 規則層吐中文的話,英文模式就沒救了(而且棘輪測試會擋)。
func TestAudienceReasonsAreCodesNotDisplayText(t *testing.T) {
	for _, r := range []string{AudienceReasonWar, AudienceReasonTrade, AudienceReasonAlliance, AudienceReasonOriginal} {
		for _, c := range r {
			if c > 127 {
				t.Errorf("來意代碼 %q 含非 ASCII 字元——那是顯示字串不是代碼", r)
				break
			}
		}
	}
}

// TestAudienceRequestsAndClear:請求清單與清除。
func TestAudienceRequestsAndClear(t *testing.T) {
	s := &GameSession{AIPlayers: []AIOpponent{
		{Name: "甲"},
		{Name: "乙", WantsAudience: true, AudienceReason: AudienceReasonTrade},
		{Name: "丙", WantsAudience: true, AudienceReason: AudienceReasonWar},
	}}
	got := s.AudienceRequests()
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("請求清單應為 [1 2],實得 %v", got)
	}
	if r := s.AudienceReasonFor(1); r != AudienceReasonTrade {
		t.Errorf("對手 1 的來意應為 %s,實得 %q", AudienceReasonTrade, r)
	}
	if r := s.AudienceReasonFor(0); r != "" {
		t.Errorf("沒有請求的對手來意應為空,實得 %q", r)
	}

	s.ClearAudienceRequestByName("乙")
	if got := s.AudienceRequests(); len(got) != 1 || got[0] != 2 {
		t.Errorf("清掉「乙」之後請求清單應為 [2],實得 %v", got)
	}
	// 越界不 panic。
	s.ClearAudienceRequest(-1)
	s.ClearAudienceRequest(99)
	s.ClearAudienceRequestByName("不存在")
	if got := s.AudienceRequests(); len(got) != 1 {
		t.Errorf("越界清除不該影響既有請求,實得 %v", got)
	}
}

func TestOriginalHumanDiplomaticRequestSurvivesSnapshotAndClearsWithAudience(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].WantsAudience = true
	s.AIPlayers[0].AudienceReason = AudienceReasonOriginal
	s.AIPlayers[0].OriginalHumanDiplomaticRequest = &gamedata.OriginalHumanDiplomaticRequest{
		Outcome: 1, ReasonCode: 106,
		Action: gamedata.OriginalHumanDiplomaticAction{Kind: gamedata.OriginalHumanDiplomaticActionTechnology, Technology: 42},
	}
	got := s.snapshot().restore()
	r := got.AIPlayers[0].OriginalHumanDiplomaticRequest
	if r == nil || r.Outcome != 1 || r.ReasonCode != 106 ||
		r.Action.Kind != gamedata.OriginalHumanDiplomaticActionTechnology || r.Action.Technology != 42 {
		t.Fatalf("原版外交請求存檔往返失真：%+v", r)
	}
	got.ClearAudienceRequest(0)
	if got.AIPlayers[0].OriginalHumanDiplomaticRequest != nil {
		t.Fatal("清除會談請求時必須一併清掉原版 payload")
	}
}
