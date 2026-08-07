package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// TestAudienceLightLayoutMatchesOriginal 釘住原版的版面:由 x=506 往左排、y=5。
//
// 這兩個是立即數(`mov edx, 1FAh` / `mov edx, 5`),抄錯會讓燈跑到星圖裡面去。
func TestAudienceLightLayoutMatchesOriginal(t *testing.T) {
	x0, y0, w0, _ := audienceLightRect(0)
	if y0 != 5 {
		t.Errorf("燈的 y 應為 5(原版立即數),實得 %d", y0)
	}
	if x0+w0 != audienceLightRightX {
		t.Errorf("第 0 盞燈的右緣應為 %d,實得 %d", audienceLightRightX, x0+w0)
	}
	// 往左排:n 越大 x 越小,而且不重疊。
	x1, _, w1, _ := audienceLightRect(1)
	if x1 >= x0 {
		t.Errorf("第 1 盞燈應在第 0 盞左邊:x0=%d x1=%d", x0, x1)
	}
	if x1+w1 != x0 {
		t.Errorf("相鄰兩盞燈應緊貼不重疊:x1+w1=%d, x0=%d", x1+w1, x0)
	}
}

// TestAudienceLightAtHitsTheRightOpponent:點到第 n 盞燈要回到對應的對手索引。
func TestAudienceLightAtHitsTheRightOpponent(t *testing.T) {
	b := &sceneBuilder{session: &shell.GameSession{AIPlayers: []shell.AIOpponent{
		{Name: "甲"},
		{Name: "乙", WantsAudience: true, AudienceReason: shell.AudienceReasonWar},
		{Name: "丙", WantsAudience: true, AudienceReason: shell.AudienceReasonTrade},
	}}}
	// 第 0 盞 = 請求清單裡的第一個(對手 1)。
	x, y, w, h := audienceLightRect(0)
	if got := b.audienceLightAt(x+w/2, y+h/2); got != 1 {
		t.Errorf("第 0 盞燈應對到對手 1,實得 %d", got)
	}
	x, y, w, h = audienceLightRect(1)
	if got := b.audienceLightAt(x+w/2, y+h/2); got != 2 {
		t.Errorf("第 1 盞燈應對到對手 2,實得 %d", got)
	}
	// 燈以外的位置回 -1。
	if got := b.audienceLightAt(10, 300); got != -1 {
		t.Errorf("星圖中央不該命中任何燈,實得 %d", got)
	}
}

// TestDrawAudienceLightsWithoutFontDoesNotPanic:沒有字型時安靜跳過。
func TestDrawAudienceLightsWithoutFontDoesNotPanic(t *testing.T) {
	b := &sceneBuilder{session: &shell.GameSession{AIPlayers: []shell.AIOpponent{
		{Name: "甲", WantsAudience: true, AudienceReason: shell.AudienceReasonWar},
	}}}
	b.drawAudienceLights(nil)
	if got := audienceEnemyName(b.session, 0); got != "甲" {
		t.Errorf("對手名稱應為「甲」,實得 %q", got)
	}
	if got := audienceEnemyName(b.session, 9); got != "" {
		t.Errorf("越界應回空字串,實得 %q", got)
	}
}
