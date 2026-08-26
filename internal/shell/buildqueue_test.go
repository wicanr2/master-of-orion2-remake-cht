package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 建造佇列的驗收基準是原版 Add_Build_Queue_Fields_:7 格、完工自動接下一項。

func TestBuildQueueEnqueueFillsCurrentFirst(t *testing.T) {
	s := NewDemoSession()
	s.Builds[0] = ColonyBuild{} // 清空當前建造
	if !s.EnqueueBuild(0, "星際港", 60) {
		t.Fatal("空佇列的第一次排入應成功")
	}
	if s.Builds[0].Name != "星際港" {
		t.Errorf("當前沒有建造時第一項應直接成為當前建造,got %q", s.Builds[0].Name)
	}
	if len(s.BuildQueue[0]) != 0 {
		t.Errorf("第一項不該進後續佇列,got %d 項", len(s.BuildQueue[0]))
	}
	if !s.EnqueueBuild(0, "海洋研究所", 80) {
		t.Fatal("第二次排入應成功")
	}
	if len(s.BuildQueue[0]) != 1 || s.BuildQueue[0][0].Name != "海洋研究所" {
		t.Errorf("第二項應排到後續佇列,got %+v", s.BuildQueue[0])
	}
}

func TestBuildQueueRespectsSevenSlotLimit(t *testing.T) {
	s := NewDemoSession()
	s.Builds[0] = ColonyBuild{}
	for i := 0; i < BuildQueueTotalSlots; i++ {
		if !s.EnqueueBuild(0, "研究實驗室", 60) {
			t.Fatalf("第 %d 項應排得進去(總格數 %d)", i+1, BuildQueueTotalSlots)
		}
	}
	if s.EnqueueBuild(0, "溢出項", 60) {
		t.Errorf("超過 %d 格應被拒絕(原版 BUILD QUEUE 就是 %d 格)", BuildQueueTotalSlots, BuildQueueTotalSlots)
	}
	if got := len(s.BuildQueueFor(0)); got != BuildQueueTotalSlots {
		t.Errorf("佇列顯示長度應為 %d,got %d", BuildQueueTotalSlots, got)
	}
}

func TestBuildQueueAdvancesToNextOnCompletion(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	// 兩項都用低成本,確保幾回合內會完工。
	s.Builds[0] = ColonyBuild{Name: "海洋研究所", Cost: 1}
	s.ensureBuildQueue()
	s.BuildQueue[0] = []ColonyBuild{{Name: "研究實驗室", Cost: 1}}

	for i := 0; i < 40 && s.Builds[0].Name == "海洋研究所"; i++ {
		s.EndTurn()
	}
	if s.Builds[0].Name == "海洋研究所" {
		t.Fatal("前提失敗:第一項在 40 回合內沒完工")
	}
	if s.Builds[0].Name != "研究實驗室" {
		t.Errorf("第一項完工後應自動接上佇列下一項,got %q", s.Builds[0].Name)
	}
	if len(s.BuildQueue[0]) != 0 {
		t.Errorf("遞補後佇列應清空,got %+v", s.BuildQueue[0])
	}
}

func TestBuildQueueSkipsAlreadyBuilt(t *testing.T) {
	s := NewDemoSession()
	if s.ColonyBuildings[0] == nil {
		s.ColonyBuildings[0] = map[string]bool{}
	}
	s.ColonyBuildings[0]["海洋研究所"] = true // 假設已由別的途徑蓋好
	s.Builds[0] = ColonyBuild{}
	s.ensureBuildQueue()
	s.BuildQueue[0] = []ColonyBuild{{Name: "海洋研究所", Cost: 60}, {Name: "研究實驗室", Cost: 60}}

	s.popNextBuild(0)
	if s.Builds[0].Name != "研究實驗室" {
		t.Errorf("已蓋好的項目應被跳過,直接接下一項,got %q", s.Builds[0].Name)
	}
}

func TestBuildQueueSpecialActionsAreRepeatable(t *testing.T) {
	s := NewDemoSession()
	// Special 一次性行動可重複套用(見 advanceBuilds 註解),不應被「已蓋過」邏輯跳過。
	if s.buildAlreadyDone(0, gamedata.TerraformActionName) {
		t.Error("地形改造應永遠可重複排入")
	}
	if s.buildAlreadyDone(0, TradeGoodsBuildName) {
		t.Error("貿易品是持續性選項,不應被視為已完成")
	}
}

func TestBuildQueueDequeue(t *testing.T) {
	s := NewDemoSession()
	s.Builds[0] = ColonyBuild{Name: "甲", Cost: 10}
	s.ensureBuildQueue()
	s.BuildQueue[0] = []ColonyBuild{{Name: "乙", Cost: 10}, {Name: "丙", Cost: 10}}

	if !s.DequeueBuild(0, 1) { // 移除「乙」
		t.Fatal("移除佇列第 1 格應成功")
	}
	if len(s.BuildQueue[0]) != 1 || s.BuildQueue[0][0].Name != "丙" {
		t.Errorf("移除後應只剩「丙」,got %+v", s.BuildQueue[0])
	}
	if !s.DequeueBuild(0, 0) { // 移除當前建造,「丙」應遞補
		t.Fatal("移除當前建造應成功")
	}
	if s.Builds[0].Name != "丙" {
		t.Errorf("移除當前建造後應由佇列遞補,got %q", s.Builds[0].Name)
	}
	if s.DequeueBuild(0, 5) {
		t.Error("越界的 pos 不應回報移除成功")
	}
}

func TestBuildQueueSurvivesSaveLoad(t *testing.T) {
	s := NewDemoSession()
	s.Builds[0] = ColonyBuild{Name: "甲", Cost: 10}
	s.ensureBuildQueue()
	s.BuildQueue[0] = []ColonyBuild{{Name: "乙", Cost: 20}}

	restored := s.snapshot().restore()
	if len(restored.BuildQueue) == 0 || len(restored.BuildQueue[0]) != 1 ||
		restored.BuildQueue[0][0].Name != "乙" {
		t.Errorf("存讀檔後佇列應保留,got %+v", restored.BuildQueue)
	}
}

func TestBlockingBuildModesMatchOriginalSevenSlotScan(t *testing.T) {
	for _, tc := range []struct {
		name    string
		current ColonyBuild
		queue   []ColonyBuild
		want    string
		deleted int
		remain  []string
	}{
		{name: "last mode is allowed", current: ColonyBuild{Name: TradeGoodsBuildName}, want: "", remain: []string{TradeGoodsBuildName}},
		{name: "trade goods before product", current: ColonyBuild{Name: TradeGoodsBuildName}, queue: []ColonyBuild{{Name: "研究實驗室"}}, want: TradeGoodsBuildName, deleted: 1, remain: []string{"研究實驗室"}},
		{name: "multiple modes preserve final mode", current: ColonyBuild{Name: TradeGoodsBuildName}, queue: []ColonyBuild{{Name: HousingBuildName}}, want: TradeGoodsBuildName, deleted: 1, remain: []string{HousingBuildName}},
		{name: "multiple modes before product all removed", current: ColonyBuild{Name: HousingBuildName}, queue: []ColonyBuild{{Name: TradeGoodsBuildName}, {Name: "自動工廠"}}, want: HousingBuildName, deleted: 2, remain: []string{"自動工廠"}},
		{name: "ordinary queue untouched", current: ColonyBuild{Name: "自動工廠"}, queue: []ColonyBuild{{Name: TradeGoodsBuildName}}, want: "", remain: []string{"自動工廠", TradeGoodsBuildName}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewDemoSession()
			s.Builds[0] = tc.current
			s.ensureBuildQueue()
			s.BuildQueue[0] = append([]ColonyBuild(nil), tc.queue...)
			if got := s.BlockingBuildMode(0); got != tc.want {
				t.Fatalf("BlockingBuildMode=%q，預期 %q", got, tc.want)
			}
			if got := s.DeleteBlockingBuildModes(0); got != tc.deleted {
				t.Fatalf("刪除數=%d，預期 %d", got, tc.deleted)
			}
			q := s.BuildQueueFor(0)
			got := make([]string, 0, len(q))
			for _, item := range q {
				if item.Name != "" {
					got = append(got, item.Name)
				}
			}
			if len(got) != len(tc.remain) {
				t.Fatalf("剩餘佇列=%v，預期 %v", got, tc.remain)
			}
			for i := range got {
				if got[i] != tc.remain[i] {
					t.Fatalf("剩餘佇列=%v，預期 %v", got, tc.remain)
				}
			}
		})
	}
}
