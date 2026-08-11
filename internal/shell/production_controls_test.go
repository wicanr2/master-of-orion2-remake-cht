package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestBuyBuildCostBoundariesAndCompletion(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Cost: 60}
	if got := s.BuildBuyCostBC(0); got != 240 {
		t.Fatalf("未開工 BUY 應為 4 BC/PP=240，得到 %d", got)
	}
	s.Builds[0].Progress = 30
	if got := s.BuildBuyCostBC(0); got != 60 {
		t.Fatalf("完成一半 BUY 應為 2 BC/剩餘 PP=60，得到 %d", got)
	}
	s.Builds[0].Progress = 40
	if got := s.BuildBuyCostBC(0); got != 40 {
		t.Fatalf("完成過半後 BUY 應維持 2 BC/剩餘 PP=40，得到 %d", got)
	}
	s.Builds[0].Progress = 0
	s.Player.BC = 300
	paid, ok := s.BuyCurrentBuild(0)
	if !ok || paid != 240 || s.Player.BC != 60 {
		t.Fatalf("BUY 扣款不正確：paid=%d ok=%v treasury=%d", paid, ok, s.Player.BC)
	}
	if !colonyBuildComplete(s.Builds[0]) {
		t.Fatal("BUY 後當前項目應標記完成，等待 EndTurn 套用效果")
	}
	s.EndTurn()
	if !s.ColonyBuildings[0]["自動工廠"] {
		t.Fatal("BUY 後 EndTurn 應完成建築並套用效果")
	}
}

func TestAutoBuildPriorityAndPersistence(t *testing.T) {
	s := NewDemoSession()
	s.Builds[0] = ColonyBuild{}
	if !s.SetAutoBuild(0, true) || s.Builds[0].Name != HousingBuildName {
		t.Fatalf("未滿人口啟用自動建造應先選住宅，得到 %+v", s.Builds[0])
	}
	s.PlayerColonies[0].Population = s.PlayerColonies[0].PopMax
	s.Player.CompletedTopics[gamedata.TOPIC_ADVANCED_CONSTRUCTION] = true
	s.advanceBuilds()
	if s.Builds[0].Name != "自動工廠" {
		t.Fatalf("滿人口且自動工廠已解鎖時，固定優先項應選自動工廠，得到 %+v", s.Builds[0])
	}
	restored := s.snapshot().restore()
	if !restored.ColonyAutoBuild(0) || restored.Builds[0].Name != "自動工廠" {
		t.Fatalf("存讀檔後 AUTO BUILD 與目前項目應保留，得到 auto=%v build=%+v",
			restored.ColonyAutoBuild(0), restored.Builds[0])
	}
}

func TestRepeatBuildRecreatesRepeatableSpecial(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Builds[0] = ColonyBuild{}
	if s.SetRepeatBuild(0, "自動工廠", 1) {
		t.Fatal("一般建築不能當 REPEAT BUILD 目標")
	}
	if s.SetRepeatBuild(0, HousingBuildName, 1) {
		t.Fatal("住宅不能當 REPEAT BUILD 目標")
	}
	if !s.SetRepeatBuild(0, gamedata.ColonyShipActionName, 1) {
		t.Fatal("殖民船是可重複 Special，應能指定 REPEAT BUILD")
	}
	startShips := s.ShipCount()
	s.EndTurn()
	if s.ShipCount() != startShips+1 {
		t.Fatalf("重複殖民船完工應交付一艘船，原 %d 現 %d", startShips, s.ShipCount())
	}
	if got := s.Builds[0]; got.Name != gamedata.ColonyShipActionName || got.Progress != 0 {
		t.Fatalf("完工後應立即重建同一 Special，得到 %+v", got)
	}
	if !s.SetRepeatBuild(0, "", 0) {
		t.Fatal("清除 REPEAT BUILD 應成功")
	}
	if got := s.RepeatBuildFor(0); got.Name != "" {
		t.Fatalf("清除後不應殘留重複目標，得到 %+v", got)
	}
}

func TestQueueRefitPersistsCompletesAndCancellationDestroysSource(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Builds[0] = ColonyBuild{}
	s.Player.CompletedTopics[gamedata.TOPIC_CHEMISTRY] = true
	source := Ship{
		Name: "試作巡防艦", Class: "巡防艦",
		Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無",
	}
	startShips := s.ShipCount()
	s.Fleet().Ships = append(s.Fleet().Ships, source)
	shipIndex := len(s.Fleet().Ships) - 1
	job, err := s.QueueRefit(0, s.SelectedFleet, shipIndex)
	if err != nil {
		t.Fatalf("應能把停泊的巡防艦排入改裝：%v", err)
	}
	if job.Target.Class != source.Class || job.Target.Armor != "鈦裝甲" {
		t.Fatalf("改裝必須保留艦體並套用已解鎖模板，得到 %+v", job.Target)
	}
	if got := s.ShipCount(); got != startShips {
		t.Fatalf("改裝中來源艦應離開艦隊，原 %d 現 %d", startShips, got)
	}
	if s.Builds[0].Refit == nil || s.Builds[0].Cost != RefitCostPP(source, job.Target) {
		t.Fatalf("改裝工作未正確寫入佇列：%+v", s.Builds[0])
	}
	restored := s.snapshot().restore()
	if restored.Builds[0].Refit == nil || restored.Builds[0].Refit.Target.Armor != "鈦裝甲" {
		t.Fatalf("改裝佇列必須跨存讀檔保存，得到 %+v", restored.Builds[0])
	}
	restored.DisableEvents = true
	restored.Builds[0].Cost = 1
	restored.EndTurn()
	found := false
	for _, f := range restored.Fleets {
		if f.AtStar != restored.PlayerColonyStarIndex(0) {
			continue
		}
		for _, sh := range f.Ships {
			if sh.Name == source.Name && sh.Armor == "鈦裝甲" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("改裝完成後艦艇應回到原殖民地星系")
	}

	cancel := NewDemoSession()
	cancel.Builds[0] = ColonyBuild{}
	cancel.Player.CompletedTopics[gamedata.TOPIC_CHEMISTRY] = true
	cancel.Fleet().Ships = append(cancel.Fleet().Ships, source)
	if _, err := cancel.QueueRefit(0, cancel.SelectedFleet, len(cancel.Fleet().Ships)-1); err != nil {
		t.Fatalf("取消案例無法排入改裝：%v", err)
	}
	if !cancel.DequeueBuild(0, 0) {
		t.Fatal("移除改裝工作應成功")
	}
	for _, sh := range cancel.AllShips() {
		if sh.Name == source.Name {
			t.Fatal("移除改裝工作後來源艦必須依原版規則報廢")
		}
	}
}

func TestRefitCostUsesManualMinimumFormula(t *testing.T) {
	source := Ship{Class: "巡防艦", Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"}
	target := source
	target.Armor = "鈦裝甲"
	if got := RefitCostPP(source, target); got != 60 {
		t.Fatalf("改裝成本應為 2*(48-18)=60，得到 %d", got)
	}
	if got := RefitCostPP(target, source); got != ShipCost(source.Class)/4 {
		t.Fatalf("降級或同成本時應套艦體四分之一最低價，得到 %d", got)
	}
}

func TestRefitCruiserRequiresOrbitalBase(t *testing.T) {
	s := NewDemoSession()
	s.Builds[0] = ColonyBuild{}
	s.Player.CompletedTopics[gamedata.TOPIC_CHEMISTRY] = true
	delete(s.ColonyBuildings[0], "星基")
	cruiser := Ship{
		Name: "無基地巡洋艦", Class: "巡洋艦",
		Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無",
	}
	s.Fleet().Ships = append(s.Fleet().Ships, cruiser)
	idx := len(s.Fleet().Ships) - 1
	if _, err := s.QueueRefit(0, s.SelectedFleet, idx); err == nil {
		t.Fatal("沒有星基時，巡洋艦改裝必須被拒絕")
	}
	s.ColonyBuildings[0]["星基"] = true
	if _, err := s.PreviewRefit(0, s.SelectedFleet, idx); err != nil {
		t.Fatalf("有星基後巡洋艦應可預覽改裝：%v", err)
	}
}

func TestProductionControlCommandsReplayToSameState(t *testing.T) {
	base := NewDemoSession()
	base.Builds[0] = ColonyBuild{Name: "自動工廠", Cost: 60}
	base.Player.BC = 999
	viaUI := cloneProductionSession(t, base)
	var recorded []PlayerCommand
	viaUI.SetCommandRecorder(func(c PlayerCommand) { recorded = append(recorded, c) })
	viaUI.SetAutoBuild(0, true)
	if _, ok := viaUI.BuyCurrentBuild(0); !ok {
		t.Fatal("測試前提：BUY 應成功")
	}
	if !viaUI.SetRepeatBuild(0, gamedata.ColonyShipActionName, 120) {
		t.Fatal("測試前提：應能設定殖民船重複建造")
	}
	replayed := cloneProductionSession(t, base)
	if err := replayed.ApplyPlayerCommands(recorded); err != nil {
		t.Fatalf("重播生產控制指令失敗：%v", err)
	}
	if got, want := replayed.StateHash(), viaUI.StateHash(); got != want {
		t.Fatalf("BUY/AUTO/REPEAT 指令重播後狀態不同：%s vs %s", got[:12], want[:12])
	}

	refitBase := NewDemoSession()
	refitBase.Builds[0] = ColonyBuild{}
	refitBase.Player.CompletedTopics[gamedata.TOPIC_CHEMISTRY] = true
	refitBase.Fleet().Ships = append(refitBase.Fleet().Ships, Ship{
		Name: "重播巡防艦", Class: "巡防艦",
		Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無",
	})
	refitUI := cloneProductionSession(t, refitBase)
	refitRecorded := []PlayerCommand{}
	refitUI.SetCommandRecorder(func(c PlayerCommand) { refitRecorded = append(refitRecorded, c) })
	if _, err := refitUI.QueueRefit(0, refitUI.SelectedFleet, len(refitUI.Fleet().Ships)-1); err != nil {
		t.Fatalf("測試前提：改裝應能排入：%v", err)
	}
	refitReplay := cloneProductionSession(t, refitBase)
	if err := refitReplay.ApplyPlayerCommands(refitRecorded); err != nil {
		t.Fatalf("重播改裝指令失敗：%v", err)
	}
	if got, want := refitReplay.StateHash(), refitUI.StateHash(); got != want {
		t.Fatalf("REFIT 指令重播後狀態不同：%s vs %s", got[:12], want[:12])
	}
}

// cloneProductionSession 走真正的網路快照編碼路徑。snapshot().restore() 會共用 slice
// 的底層陣列，適合測試存檔欄位重建，卻不適合拿來模擬兩台獨立的多人機器。
func cloneProductionSession(t *testing.T, s *GameSession) *GameSession {
	t.Helper()
	data, err := s.MarshalSnapshot()
	if err != nil {
		t.Fatalf("序列化生產控制起始狀態失敗：%v", err)
	}
	clone, err := RestoreSnapshot(data)
	if err != nil {
		t.Fatalf("還原生產控制起始狀態失敗：%v", err)
	}
	return clone
}
