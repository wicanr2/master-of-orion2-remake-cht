package ai

import "testing"

func TestDecideColonyJobs(t *testing.T) {
	cases := []struct {
		pop, fpf   int
		p          Profile
		wF, wW, wS int
	}{
		{10, 5, ProfileBalanced, 2, 4, 4},   // 農2餵飽,餘8均分
		{10, 5, ProfileAggressive, 2, 6, 2}, // 餘8 → 工6研2(3:1)
		{10, 5, ProfileScientific, 2, 2, 6}, // 餘8 → 工2研6(1:3)
		{6, 2, ProfileBalanced, 3, 1, 2},    // 農3,餘3 → 工1研2
		{10, 0, ProfileBalanced, 0, 5, 5},   // 無法務農 → 全分工研
		{0, 5, ProfileBalanced, 0, 0, 0},    // 無人口
	}
	for _, c := range cases {
		f, w, s := DecideColonyJobs(c.pop, c.fpf, c.p)
		if f != c.wF || w != c.wW || s != c.wS {
			t.Errorf("DecideColonyJobs(%d,%d,%s) = (%d,%d,%d),預期 (%d,%d,%d)",
				c.pop, c.fpf, c.p.Name, f, w, s, c.wF, c.wW, c.wS)
		}
		// 驗證餵得飽:農夫產食 >= 人口(foodPerFarmer>0 時)
		if c.fpf > 0 && f*c.fpf < c.pop {
			t.Errorf("分配未餵飽人口:農%d*%d < 人口%d", f, c.fpf, c.pop)
		}
		// 驗證總分配 = 人口
		if f+w+s != c.pop {
			t.Errorf("分配總和 %d != 人口 %d", f+w+s, c.pop)
		}
	}
}

func TestDecideTaxRate(t *testing.T) {
	if DecideTaxRate(5, 20, 100) != 50 {
		t.Error("國庫低應提高稅率至 50")
	}
	if DecideTaxRate(50, 20, 100) != 30 {
		t.Error("國庫中等應 30")
	}
	if DecideTaxRate(200, 20, 100) != 10 {
		t.Error("國庫充裕應降至 10")
	}
}

// TestMinWorkersForSolvency 驗證財政保底門檻計算(見 economy.go 註解的第一性原理推導)。
// 母星實際數值(playerHomeworldColony/homeworldBuildings):IndustryPerWorker=3、
// MoralePercent=10、Maintenance=3。工人數 w 產出 gross=MoraleProductionOutput(w*3,10),
// 換算稅收 gamedata.IncomeTaxRevenue(gross,50)——逐一驗算:
//
//	w=1: gross=3*110/100=3,  tax=3*50/100=1  <3(不足)
//	w=2: gross=6*110/100=6,  tax=6*50/100=3  >=3(打平,是最小值)
func TestMinWorkersForSolvency(t *testing.T) {
	if got := MinWorkersForSolvency(3, 10, 3, 4); got != 2 {
		t.Errorf("母星保底工人數 = %d,預期 2", got)
	}
	if got := MinWorkersForSolvency(3, 10, 0, 4); got != 0 {
		t.Errorf("maintenanceBC<=0 應回 0,實得 %d", got)
	}
	if got := MinWorkersForSolvency(0, 10, 3, 4); got != 0 {
		t.Errorf("industryPerWorker<=0 應回 0(無法產工業打平),實得 %d", got)
	}
	// 財政不可能打平(maintenanceBC 遠超過 maxWorkers 能生產的上限)時,回 maxWorkers(盡力而為,
	// 不無限迴圈、不 panic)。
	if got := MinWorkersForSolvency(3, 10, 1000, 4); got != 4 {
		t.Errorf("財政不可能打平時應回 maxWorkers=4,實得 %d", got)
	}
}

// TestDecideColonyJobsSolvent 驗證財政保底只在「純比例分配打平不了」時介入,且只挪動最少量。
func TestDecideColonyJobsSolvent(t *testing.T) {
	// 母星實際場景(docs/tech/ai-fiscal-solvency.md):Population=8、FoodPerFarmer=2、
	// IndustryPerWorker=3、MoralePercent=10、Maintenance=3。
	// Scientific(1:3)純比例會給 farmers=4,workers=1,scientists=3——workers=1 打平不了
	// maintenanceBC=3(見 TestMinWorkersForSolvency,需要至少 2),保底應把 1 個科學家挪回
	// 工人:farmers=4,workers=2,scientists=2。
	if f, w, s := DecideColonyJobsSolvent(8, 2, 3, 10, 3, ProfileScientific); f != 4 || w != 2 || s != 2 {
		t.Errorf("Scientific 財政保底 = %d/%d/%d,預期 4/2/2", f, w, s)
	}
	// Aggressive(3:1)純比例給 farmers=4,workers=3,scientists=1——workers=3 已經 >= 保底需求 2,
	// 不應被保底邏輯多動:維持 4/3/1(與 DecideColonyJobs 純比例結果一致)。
	if f, w, s := DecideColonyJobsSolvent(8, 2, 3, 10, 3, ProfileAggressive); f != 4 || w != 3 || s != 1 {
		t.Errorf("Aggressive 不應被保底邏輯多動 = %d/%d/%d,預期 4/3/1(與純比例相同)", f, w, s)
	}
	// maintenanceBC<=0(如尚無建築維護費)時,行為應與純比例 DecideColonyJobs 完全相同——
	// 保底邏輯不介入。
	wantF, wantW, wantS := DecideColonyJobs(8, 2, ProfileScientific)
	if f, w, s := DecideColonyJobsSolvent(8, 2, 3, 10, 0, ProfileScientific); f != wantF || w != wantW || s != wantS {
		t.Errorf("maintenanceBC=0 應與純比例分配相同:得 %d/%d/%d,預期 %d/%d/%d", f, w, s, wantF, wantW, wantS)
	}
}
