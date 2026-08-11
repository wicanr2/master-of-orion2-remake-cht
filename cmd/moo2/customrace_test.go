package main

import "testing"

func TestCustomRaceSpecialsCarryTraitsAndFitLogicalCanvas(t *testing.T) {
	specials := defaultSpecials()
	if len(specials) != 22 {
		t.Fatalf("客製種族特殊能力應有官方 22 項,得到 %d", len(specials))
	}
	for i, sp := range specials {
		if sp.trait == 0 {
			t.Errorf("第 %d 項特殊能力沒有特性編號", i)
		}
	}
	for i := range specials {
		x, y, w, h := (&customRaceScreen{}).spcRect(i)
		if x < crSpcX || x+w > 640 || y < crSpcY || y+h > 440 {
			t.Fatalf("第 %d 項特殊能力矩形超出邏輯畫布或撞到底部按鈕: (%d,%d,%d,%d)", i, x, y, w, h)
		}
	}
}

func TestCustomRaceNumericCombatPickCategoriesReachSeparateRaceFields(t *testing.T) {
	cats := defaultPickCats()
	for i := range cats {
		switch cats[i].name {
		case "艦艇攻擊", "艦艇防禦", "地面戰", "諜報":
			cats[i].sel = 2 // 各自選官方表的中間正向檔
		}
	}
	r := customRaceValues(cats)
	if r.CombatPct != 20 || r.ShipDefPct != 25 || r.GroundCombatBonus != 10 || r.SpyBonus != 10 {
		t.Fatalf("四類戰鬥／諜報 picks 應分別寫入艦攻／艦防／地面／諜報: %+v", r)
	}
}
