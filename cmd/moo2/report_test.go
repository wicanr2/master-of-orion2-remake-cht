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
		BCGained: 100, ColonyIdx: -1,
	}
	b := &sceneBuilder{session: s, lang: i18n.Traditional}
	r := b.currentReport()
	if r == nil {
		t.Fatal("有星系發現時應該要播快報")
	}
	// 發現是自家勘查隊的回報,不是 GNN 新聞——台標列要不一樣。
	wantHeader := uiText(i18n.Traditional, "event.header.survey")
	if r.header != wantHeader {
		t.Errorf("台標列 = %q,want %q", r.header, wantHeader)
	}
	if !r.good {
		t.Error("原版這五種發現沒有負面的,應標為好消息(綠框)")
	}
	if !strings.Contains(r.body, "100 BC") {
		t.Errorf("內文應含入袋金額,實為 %q", r.body)
	}
}

func TestCurrentReportDiscoveryEnglishUsesBilingualPayload(t *testing.T) {
	s := shell.NewDemoSession()
	s.LastDiscovery = &shell.SystemDiscovery{
		StarName:   "測試星",
		StarNameEN: "Test",
		Special:    gamedata.SpaceDebris,
		BCGained:   50,
		ColonyIdx:  -1,
	}
	r := (&sceneBuilder{session: s, lang: i18n.English}).currentReport()
	if r == nil {
		t.Fatal("英文模式也應該要播星系發現快報")
	}
	if r.title != "Space Debris" || !strings.Contains(r.body, "Test system") {
		t.Errorf("英文快報沒有使用雙語 payload: title=%q body=%q", r.title, r.body)
	}
	if strings.Contains(r.body, "勘查小隊") {
		t.Errorf("英文快報仍外洩中文: %q", r.body)
	}
}

func TestDiscoveryBodyTextLocalizesTypedTechTopics(t *testing.T) {
	d := &shell.SystemDiscovery{StarName: "測試星", StarNameEN: "Test", Special: gamedata.AncientArtifacts,
		TechTopics: []gamedata.ResearchTopic{gamedata.TOPIC_ENGINEERING, gamedata.TOPIC_ADVANCED_GOVERNMENTS}}
	zh := discoveryBodyText(i18n.Traditional, d)
	en := discoveryBodyText(i18n.English, d)
	if !strings.Contains(zh, "、") || strings.Contains(en, "、") || !strings.Contains(en, ", ") {
		t.Fatalf("科技清單沒有依語系套用外部分隔符：zh=%q en=%q", zh, en)
	}
	if strings.Contains(en, "測試星") || !strings.Contains(en, "Test") {
		t.Errorf("英文科技發現沒有使用英文星名：%q", en)
	}
}

func TestCurrentReportKeepsLegacyDiscoveryPayload(t *testing.T) {
	s := shell.NewDemoSession()
	s.LastDiscovery = &shell.SystemDiscovery{StarName: "舊星", Special: gamedata.AncientArtifacts,
		Name: "遠古文物", Message: "舊版科技發現句子", TechGot: "舊科技", ColonyIdx: -1}
	r := (&sceneBuilder{session: s, lang: i18n.Traditional}).currentReport()
	if r == nil || r.body != "舊版科技發現句子" {
		t.Fatalf("舊存檔發現報告未安全回退：%+v", r)
	}
}

// 兩者同時發生時先播隨機事件(全銀河新聞優先於自家回報),發現的內容仍留在回合摘要文字裡。
func TestCurrentReportEventBeatsDiscovery(t *testing.T) {
	s := shell.NewDemoSession()
	s.LastEventReport = &shell.EventReport{Name: "瘟疫", Good: false, Message: "疫病蔓延"}
	s.LastDiscovery = &shell.SystemDiscovery{StarName: "測試星", Name: "海盜藏寶", Message: "找到藏寶", ColonyIdx: -1}
	r := (&sceneBuilder{session: s, lang: i18n.Traditional}).currentReport()
	if r == nil || r.header != uiText(i18n.Traditional, "event.header.gnn") {
		t.Fatalf("應優先播 GNN 事件快報,實得 %+v", r)
	}
	if r.good {
		t.Error("壞消息事件應標為紅框")
	}
}

func TestCurrentReportGNNOffKeepsDiscoveryAndDefersEventToSummary(t *testing.T) {
	s := shell.NewDemoSession()
	settings := s.EffectiveGameSettings()
	settings.ShowGNNReport = false
	s.ApplyGameSettings(settings)
	s.LastEventReport = &shell.EventReport{Name: "瘟疫", Good: false, Message: "疫病蔓延"}
	s.LastDiscovery = &shell.SystemDiscovery{Name: "太空殘骸", Message: "找到殘骸", ColonyIdx: -1}
	b := &sceneBuilder{session: s, lang: i18n.Traditional}
	r := b.currentReport()
	if r == nil || r.header != uiText(i18n.Traditional, "event.header.survey") {
		t.Fatalf("關閉 GNN 後應略過事件畫面但保留勘查回報：%+v", r)
	}
	if !b.shouldForceEventIntoSummary() {
		t.Fatal("關閉 GNN 後特殊事件必須強制留在一般回合摘要")
	}
}

func TestCurrentReportGNNOffWithoutDiscoveryHasNoReportScreen(t *testing.T) {
	s := shell.NewDemoSession()
	settings := s.EffectiveGameSettings()
	settings.ShowGNNReport = false
	settings.EndOfTurnSummary = false
	s.ApplyGameSettings(settings)
	s.LastEventReport = &shell.EventReport{Name: "瘟疫", Good: false, Message: "疫病蔓延"}
	b := &sceneBuilder{session: s, lang: i18n.Traditional}
	if b.shouldOpenReportScreen() {
		t.Fatal("關閉 GNN 後不應開啟 GNN 快報畫面")
	}
	if !b.shouldForceEventIntoSummary() {
		t.Fatal("即使一般摘要選項關閉，特殊事件仍必須進摘要")
	}
}

func TestShouldShowTurnSummarySeriousGate(t *testing.T) {
	s := shell.NewDemoSession()
	settings := s.EffectiveGameSettings()
	settings.EndOfTurnSummary = true
	settings.ShowOnlySeriousTurnSummary = true
	s.ApplyGameSettings(settings)
	b := &sceneBuilder{session: s, lang: i18n.Traditional}
	if b.shouldShowTurnSummary() {
		t.Fatal("只有一般經濟資料時，重要摘要模式應略過摘要")
	}
	s.LastRebellions = []shell.RebellionResult{{Triggered: true}}
	if !b.shouldShowTurnSummary() {
		t.Fatal("叛亂應觸發重要回合摘要")
	}
}

func TestShouldShowTurnSummaryGNNFallbackBeatsSeriousGate(t *testing.T) {
	s := shell.NewDemoSession()
	settings := s.EffectiveGameSettings()
	settings.EndOfTurnSummary = false
	settings.ShowOnlySeriousTurnSummary = true
	settings.ShowGNNReport = false
	s.ApplyGameSettings(settings)
	s.LastEventReport = &shell.EventReport{Name: "event", Message: "event"}
	b := &sceneBuilder{session: s, lang: i18n.Traditional}
	if !b.shouldShowTurnSummary() {
		t.Fatal("GNN 關閉後的特殊事件通知不可被摘要選項或重要摘要 gate 吞掉")
	}
}

func TestCurrentReportEventEnglishUsesBilingualPayload(t *testing.T) {
	s := shell.NewDemoSession()
	s.LastEventReport = &shell.EventReport{
		Name: "富商捐獻", NameEN: "Merchant Donation", Good: true,
		Message: "一名富商向帝國捐獻了 51 BC", MessageEN: "A wealthy merchant donated 51 BC to the empire.",
	}
	r := (&sceneBuilder{session: s, lang: i18n.English}).currentReport()
	if r == nil {
		t.Fatal("英文模式也應該要播事件快報")
	}
	if r.title != "Merchant Donation" || !strings.Contains(r.body, "donated 51 BC") {
		t.Errorf("英文事件快報沒有使用雙語 payload:title=%q body=%q", r.title, r.body)
	}
	if strings.Contains(r.body, "富商") {
		t.Errorf("英文事件快報仍外洩中文:%q", r.body)
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
	wantZH := uiText(i18n.Traditional, "event.header.gnn")
	wantEN := uiText(i18n.English, "event.header.gnn")
	if zh.header != wantZH {
		t.Errorf("中文台標 = %q,want %q", zh.header, wantZH)
	}
	if en.header != wantEN {
		t.Errorf("英文台標 = %q,want %q", en.header, wantEN)
	}
	if zh.tag == en.tag {
		t.Errorf("標記也該跟著語言換,兩者都是 %q", zh.tag)
	}
}
