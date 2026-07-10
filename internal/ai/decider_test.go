package ai

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/diplomacy"
)

func TestRemakeDecider(t *testing.T) {
	var d Decider = NewRemakeDecider(ProfileScientific)
	if d.Mode() != ModeRemake || d.Mode().Name() != "remake" {
		t.Errorf("模式錯誤:%d", d.Mode())
	}
	// 委派既有函式:科學傾向 10 人口/5 食物 → 2/2/6(maintenanceBC=0:不觸發財政保底,見
	// TestColonyJobsSolvent 測保底邏輯本身)
	f, w, s := d.ColonyJobs(10, 5, 10, 0, 0)
	if f != 2 || w != 2 || s != 6 {
		t.Errorf("ColonyJobs 委派錯誤:%d/%d/%d", f, w, s)
	}
	if d.TaxRate(500) != 10 { // 國庫充裕
		t.Error("TaxRate 委派錯誤")
	}
	if d.Stance(diplomacy.RelationFeud) != StanceHostile { // 科學非好戰,敵對→敵視
		t.Error("Stance 委派錯誤")
	}
}

func TestNewDeciderModes(t *testing.T) {
	// remake 模式:ok=true
	if _, ok := NewDecider(ModeRemake, ProfileBalanced); !ok {
		t.Error("remake 模式應可用")
	}
	// original 模式:目前 fallback,ok=false,但仍回可用決策器
	d, ok := NewDecider(ModeOriginal, ProfileBalanced)
	if ok {
		t.Error("original 模式尚未實作,ok 應為 false")
	}
	if d == nil {
		t.Fatal("fallback 決策器不應為 nil")
	}
	// fallback 決策器仍能運作(remake 行為)
	if f, _, _ := d.ColonyJobs(4, 2, 10, 0, 0); f != 2 {
		t.Errorf("fallback 決策器 ColonyJobs 錯誤:農%d", f)
	}
}
