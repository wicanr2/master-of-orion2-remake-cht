package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestDiplomacyProposalGridFits640x480(t *testing.T) {
	d := &diplomacyScreen{}
	for i := 0; i < 9; i++ {
		x, y, w, h := d.optRect(i)
		if x < 0 || y < 0 || x+w > moo2ScreenW || y+h > 365 {
			t.Fatalf("第 %d 個外交提案按鈕超出畫面配置: (%d,%d,%d,%d)", i, x, y, w, h)
		}
		for j := 0; j < i; j++ {
			px, py, pw, ph := d.optRect(j)
			if x < px+pw && px < x+w && y < py+ph && py < y+h {
				t.Fatalf("外交提案按鈕 %d 與 %d 重疊", i, j)
			}
		}
	}
}

func TestOriginalDiplomacyRequestButtonsFitAndResolveWithoutPrematureClear(t *testing.T) {
	d := &diplomacyScreen{}
	for i := 0; i < 2; i++ {
		x, y, w, h := d.requestRect(i)
		if x < 0 || y < 0 || x+w > moo2ScreenW || y+h > 430 {
			t.Fatalf("request button %d out of bounds: %d,%d,%d,%d", i, x, y, w, h)
		}
	}
	s := shell.NewDemoSession()
	r := gamedata.OriginalHumanDiplomaticRequest{Outcome: 3, ReasonCode: 105,
		Action: gamedata.OriginalHumanDiplomaticAction{Kind: gamedata.OriginalHumanDiplomaticActionDirect, DirectTier: 1}}
	s.AIPlayers[0].WantsAudience = true
	s.AIPlayers[0].AudienceReason = shell.AudienceReasonOriginal
	s.AIPlayers[0].OriginalHumanDiplomaticRequest = &r
	b := &sceneBuilder{session: s}
	d = &diplomacyScreen{b: b, requestAI: 0, request: &r}
	if _, ok := s.PendingOriginalAIHumanDiplomaticRequest(0); !ok {
		t.Fatal("進畫面前不得清掉 typed request")
	}
	if !d.resolveOriginalRequest(false) || d.request != nil {
		t.Fatal("reason 105 拒絕按鈕應執行 callback 並結束 pending request")
	}
}

func TestDiplomacyBreakButtonsStayAboveEndAudience(t *testing.T) {
	d := &diplomacyScreen{backRect: [4]int{250, 430, 140, 34}}
	for i := 0; i < 4; i++ {
		x, y, w, h := d.breakRect(i)
		if x < 0 || x+w > moo2ScreenW || y < 0 || y+h >= d.backRect[1] {
			t.Fatalf("第 %d 個終止按鈕與結束對談按鈕衝突: (%d,%d,%d,%d)", i, x, y, w, h)
		}
	}
}
