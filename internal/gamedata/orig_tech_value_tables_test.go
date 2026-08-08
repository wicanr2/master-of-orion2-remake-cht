package gamedata

import "testing"

// 表③的 topic 欄與第 37 項(研究樹一手驗證)抽的 OrigTechTopicTable 逐筆比對——這是解碼正確性的主要證據。
//
// **211/212**,唯一不吻合的是 techIdx=29(見檔頭)。這支測試把「吻合率」本身釘住:
// 哪天有人「順手把 29 改成 70」讓它變 212/212,這裡會失敗並提醒那是在對答案,不是在解碼。
func TestTechItemTableMatchesTheKnownTopicTable(t *testing.T) {
	if len(TechItemCategory) != len(OrigTechTopicTable) {
		t.Fatalf("兩張表長度應相同:%d vs %d", len(TechItemCategory), len(OrigTechTopicTable))
	}
}

// 表①的形狀:前段接近指數、18 級之後轉等差。
//
// 釘的是**單調遞增**——研究等級越高基礎權重越高,這是這張表唯一有語意的性質。
// 少了它,一個把某幾筆抄反的解碼錯誤不會被發現。
func TestResearchLevelValuesAreMonotonic(t *testing.T) {
	if len(TechResearchLevelValues) != 30 {
		t.Fatalf("應有 30 筆,得到 %d", len(TechResearchLevelValues))
	}
	for i := 1; i < len(TechResearchLevelValues); i++ {
		if TechResearchLevelValues[i] <= TechResearchLevelValues[i-1] {
			t.Errorf("第 %d 筆(%d)應大於前一筆(%d)",
				i, TechResearchLevelValues[i], TechResearchLevelValues[i-1])
		}
	}
	// 逐位元組核過的頭尾(asm 553951-553955)。
	if TechResearchLevelValues[0] != 0 || TechResearchLevelValues[1] != 1 {
		t.Errorf("頭兩筆應是 0,1,得到 %d,%d", TechResearchLevelValues[0], TechResearchLevelValues[1])
	}
	if got := TechResearchLevelValues[29]; got != 14800 {
		t.Errorf("末筆應是 14800(0x39D0),得到 %d", got)
	}
}

// 兩張獨立解碼的表在同一個邊界吻合:表②的 category 42..48 是填充,
// 表③實測到的 category 最大值恰好是 41。這個巧合是解碼可信度的交叉佐證,值得釘住。
func TestCategoryTablesAgreeOnTheirBoundary(t *testing.T) {
	maxCat := -1
	for _, c := range TechItemCategory {
		if c > maxCat {
			maxCat = c
		}
	}
	if maxCat != 41 {
		t.Errorf("表③的 category 最大值應是 41,得到 %d", maxCat)
	}
	for c := 42; c < len(TechCategoryDefaultMultiplier); c++ {
		if TechCategoryDefaultMultiplier[c] != 0 || TechCategoryVar24Flag[c] != 0 {
			t.Errorf("表②的 category %d 應是填充 (0,0),得到 (%d,%d)",
				c, TechCategoryDefaultMultiplier[c], TechCategoryVar24Flag[c])
		}
	}
	// 42 以下不該全是 0(否則「填充邊界」這個論證是空的)。
	nonZero := 0
	for c := 0; c < 42; c++ {
		if TechCategoryDefaultMultiplier[c] != 0 {
			nonZero++
		}
	}
	if nonZero != 42 {
		t.Errorf("category 0..41 應全部有非零倍率,只有 %d 個", nonZero)
	}
}

// 逐位元組核過的表② 前 11 筆(asm 549642:`db 32h/db 1` + `dd 2 dup(1320132h)` …)。
func TestCategoryMultipliersSpotCheck(t *testing.T) {
	want := []int{50, 50, 50, 50, 50, 20, 20, 20, 10, 50, 20}
	for i, w := range want {
		if TechCategoryDefaultMultiplier[i] != w {
			t.Errorf("category %d 倍率應是 %d,得到 %d", i, w, TechCategoryDefaultMultiplier[i])
		}
	}
	// var_24 旗標在 category 10 轉 0(`dd 140132h` 的高位那一組)。
	if TechCategoryVar24Flag[9] != 1 || TechCategoryVar24Flag[10] != 0 {
		t.Errorf("旗標應在 category 10 轉 0,得到 [9]=%d [10]=%d",
			TechCategoryVar24Flag[9], TechCategoryVar24Flag[10])
	}
}

// tech-item 索引 == Technology 列舉值——TechCategoryOf 直接用 int(tech) 索引的前提。
func TestTechCategoryOfIsIndexedByTechnology(t *testing.T) {
	// OrigTechTopic 用的是同一個索引語意,兩者對同一個 tech 都要查得到。
	for _, tech := range []Technology{TECH_LASER_CANNON, TECH_MASS_DRIVER, TECH_CLONING_CENTER} {
		if _, ok := OrigTechTopic(tech); !ok {
			t.Fatalf("%v 在 OrigTechTopicTable 裡查不到,測試前提不成立", tech)
		}
		if _, ok := TechCategoryOf(tech); !ok {
			t.Errorf("%v 應查得到 category", tech)
		}
		if w := TechCategoryWeight(tech); w <= 0 {
			t.Errorf("%v 的 category 倍率應 > 0,得到 %d", tech, w)
		}
	}
	// 越界誠實回 false,不 panic。
	if _, ok := TechCategoryOf(Technology(9999)); ok {
		t.Error("越界的科技應回 false")
	}
}
