package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// report_test.go:快報面板的內容選擇(隨機事件 / 星系發現共用同一個版面)。
// 這裡驗的是「哪一則會被播出來」的接線,像素版面本身由截圖廊驗證。

func TestCurrentReportNilWhenNothingHappened(t *testing.T) {
	b := &sceneBuilder{session: shell.NewDemoSession()}
	if r := b.currentReport(); r != nil {
		t.Errorf("沒有事件也沒有發現時應回 nil,卻回了 %+v", r)
	}
	// session 為 nil(還沒開始遊戲)也不能 panic。
	if r := (&sceneBuilder{}).currentReport(); r != nil {
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
	b := &sceneBuilder{session: s}
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
	r := (&sceneBuilder{session: s}).currentReport()
	if r == nil || r.header != evGNNHeader {
		t.Fatalf("應優先播 GNN 事件快報,實得 %+v", r)
	}
	if r.good {
		t.Error("壞消息事件應標為紅框")
	}
}
