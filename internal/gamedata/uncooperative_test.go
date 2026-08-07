package gamedata

import "testing"

// 手冊:「each alien unit produces only **three quarters** what it normally would」。
func TestUncooperativeAliensProduceThreeQuarters(t *testing.T) {
	// 全員都是未整合外星人:每單位 4 → 3。
	if got := UncooperativeJobOutput(10, 4, 8, 8, false); got != 30 {
		t.Errorf("全員未整合時 10 人 × (4×3/4) 應為 30,得到 %d", got)
	}
	// 一個外星人都沒有:完全不打折。
	if got := UncooperativeJobOutput(10, 4, 0, 8, false); got != 40 {
		t.Errorf("沒有未整合人口時應為 40,得到 %d", got)
	}
	// 一半一半:8 人口裡 4 個未整合 → 10 個工人裡 5 個算外星人。
	if got := UncooperativeJobOutput(10, 4, 4, 8, false); got != 5*4+5*3 {
		t.Errorf("半數未整合應為 %d,得到 %d", 5*4+5*3, got)
	}
}

// 外星人依人口比例攤到各職業——remake 沒有「誰在做哪個工作」的模型(建模選擇,非手冊規則)。
func TestUncooperativeAlienUnitsSplitByPopulationShare(t *testing.T) {
	cases := []struct{ job, unassim, pop, want int }{
		{10, 8, 8, 10},  // 全員未整合 → 整個職業都是外星人
		{10, 0, 8, 0},   // 沒有未整合人口
		{10, 4, 8, 5},   // 一半
		{3, 1, 8, 0},    // 3×1/8 = 0(向下取整)
		{10, 3, 8, 3},   // 10×3/8 = 3
		{10, 99, 8, 10}, // unassim > pop → 夾住,不會算出比職業人數還多
		{0, 8, 8, 0},    // 沒人做這個工作
		{10, 8, 0, 0},   // 人口 0(除零防護)
	}
	for _, c := range cases {
		if got := UncooperativeAlienUnits(c.job, c.unassim, c.pop); got != c.want {
			t.Errorf("UncooperativeAlienUnits(%d,%d,%d) = %d,預期 %d",
				c.job, c.unassim, c.pop, got, c.want)
		}
	}
}

// 「每工人至少 1 產能」的下限這才真的擋到東西。
//
// 先前礦產表最低就是 1(`mineralProductionTable = {1,2,3,5,8}`),下限永遠不會生效,
// `ProdWorkerOutput` 因此零覆蓋。3/4 之後 1×3/4=0,下限才有作用。
func TestWorkerFloorFinallyBites(t *testing.T) {
	// 極貧星(每工人 1)全員未整合:沒有下限會變成 0。
	if got := UncooperativeJobOutput(6, 1, 8, 8, true); got != 6 {
		t.Errorf("工業套下限後 6 個工人應仍產 6,得到 %d", got)
	}
	// 對照:食物/研究不套那個下限(手冊只講「工人」),所以真的會掉到 0。
	if got := UncooperativeJobOutput(6, 1, 8, 8, false); got != 0 {
		t.Errorf("不套下限時 6 × (1×3/4) 應為 0,得到 %d", got)
	}
}

// 防呆:人數或單位產出非正時回 0,不會算出負數。
func TestUncooperativeJobOutputGuards(t *testing.T) {
	for _, c := range [][4]int{{0, 4, 0, 8}, {-1, 4, 0, 8}, {10, 0, 0, 8}, {10, -3, 0, 8}} {
		if got := UncooperativeJobOutput(c[0], c[1], c[2], c[3], true); got != 0 {
			t.Errorf("UncooperativeJobOutput%v = %d,預期 0", c, got)
		}
	}
}
