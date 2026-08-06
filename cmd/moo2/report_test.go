package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// report_test.go:快報面板的內容選擇(隨機事件 / 星系發現共用同一個版面)。
// 這裡驗的是「哪一則會被播出來」的接線,像素版面本身由截圖廊驗證。
//
// ⚠ 建 sceneBuilder 一定要明寫 lang。`i18n.English` 是 `Lang` 的 **Go 零值**
// (`English Lang = iota`),所以 `&sceneBuilder{session: s}` 預設是英文模式——
// 面板文字會是英文,拿中文常數去比就會失敗。這不是測試的怪癖,production 端
// 任何忘了設 lang 的建構路徑同樣會靜默變英文。

func TestCurrentReportNilWhenNothingHappened(t *testing.T) {
	b := &sceneBuilder{session: shell.NewDemoSession(), lang: i18n.Traditional}
	if r := b.currentReport(); r != nil {
		t.Errorf("沒有事件也沒有發現時應回 nil,卻回了 %+v", r)
	}
	// session 為 nil(還沒開始遊戲)也不能 panic。
	if r := (&sceneBuilder{lang: i18n.Traditional}).currentReport(); r != nil {
		t.Errorf("session 為 nil 時應回 nil,卻回了 %+v", r)
	}
}

func TestCurrentReportDiscoveryUsesScoutHeader(t *testing.T) {
	s := shell.NewDemoSession()
	s.LastDiscovery = &shell.SystemDiscovery{
		StarName: "測試星", Special: gamedata.PirateCache,
		Name:     gamedata.PlanetSpecialName(gamedata.PirateCache),
		Message:  "勘查小隊在測試星星系裡找到海盜藏寶,變賣所得 100 BC 已入國庫。",
		BCGained: 100, ColonyIdx: -1,
	}
	b := &sceneBuilder{session: s, lang: i18n.Traditional}
	r := b.currentReport()
	if r == nil {
		t.Fatal("有星系發現時應該要播快報")
	}
	// 發現是自家勘查隊的回報,不是 GNN 新聞——台標列要不一樣。
	if r.header != evScoutHeader {
		t.Errorf("台標列 = %q,want %q", r.header, evScoutHeader)
	}
	if !r.good {
		t.Error("原版這五種發現沒有負面的,應標為好消息(綠框)")
	}
	if !strings.Contains(r.body, "100 BC") {
		t.Errorf("內文應含入袋金額,實為 %q", r.body)
	}
}

// 兩者同時發生時先播隨機事件(全銀河新聞優先於自家回報),發現的內容仍留在回合摘要文字裡。
func TestCurrentReportEventBeatsDiscovery(t *testing.T) {
	s := shell.NewDemoSession()
	s.LastEventReport = &shell.EventReport{Name: "瘟疫", Good: false, Message: "疫病蔓延"}
	s.LastDiscovery = &shell.SystemDiscovery{StarName: "測試星", Name: "海盜藏寶", Message: "找到藏寶", ColonyIdx: -1}
	r := (&sceneBuilder{session: s, lang: i18n.Traditional}).currentReport()
	if r == nil || r.header != evGNNHeader {
		t.Fatalf("應優先播 GNN 事件快報,實得 %+v", r)
	}
	if r.good {
		t.Error("壞消息事件應標為紅框")
	}
}

// TestCurrentReportHeaderFollowsLanguage:同一則事件,英文模式要換成英文台標。
// 這一條同時是**零值陷阱**的回歸測試:`i18n.English` 是 Lang 的 Go 零值,
// 忘了設 lang 的建構路徑會靜默落到英文——真的落到英文時,這裡的中文斷言會先炸。
func TestCurrentReportHeaderFollowsLanguage(t *testing.T) {
	newSession := func() *shell.GameSession {
		s := shell.NewDemoSession()
		s.LastEventReport = &shell.EventReport{Name: "瘟疫", Good: false, Message: "疫病蔓延"}
		return s
	}
	zh := (&sceneBuilder{session: newSession(), lang: i18n.Traditional}).currentReport()
	en := (&sceneBuilder{session: newSession(), lang: i18n.English}).currentReport()
	if zh == nil || en == nil {
		t.Fatal("兩種語言都該產出快報")
	}
	if zh.header != evGNNHeader {
		t.Errorf("中文台標 = %q,want %q", zh.header, evGNNHeader)
	}
	if en.header != evGNNHeaderEn {
		t.Errorf("英文台標 = %q,want %q", en.header, evGNNHeaderEn)
	}
	if zh.tag == en.tag {
		t.Errorf("標記也該跟著語言換,兩者都是 %q", zh.tag)
	}
}
