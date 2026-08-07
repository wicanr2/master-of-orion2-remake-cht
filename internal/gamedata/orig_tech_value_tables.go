package gamedata

// orig_tech_value_tables.go:`Calc_Tech_Value_` @ 0xFC845 用到的**三張靜態表**。
//
// 這三張表是原版 AI「這個科技對我值多少」的估值基礎(見 `docs/re/calc-tech-value.md`)。
// 完整的估值函式有 985 行、依賴多個語意未定的玩家欄位,**沒有整段照抄**;
// 這裡只收「解碼無歧義、而且 remake 用得到」的部分。
//
// ============ 解碼怎麼驗的 ============
//
// 表③(tech-item 記錄)的 topic 欄與 repo 裡既有的 `OrigTechTopicTable`(第 91 項抽的)
// **逐筆比對:211/212 吻合**。這不是「調到吻合」——過程中先撞出 41 筆錯誤,查出原因是
// IDA 的顯示慣例(4-byte 欄位的數值只要巧合等於某個已知位址,就會顯示成
// `offset labelname[+N]` 而不是十六進位常數,即使那 4 bytes 語意上根本不是指標),
// 改成還原真實位址之後才跳到 211/212。
//
// **唯一不吻合的是 techIdx=29**:解出 topic=889,而兩個獨立來源(`OrigTechTopicTable[29]`
// 與 `techtree.go` 的 `researchChoices[70]` 含 `TECH_BIOMORPHIC_FUNGI`)都指向 70。
// 問題侷限在那一筆的頭 4 bytes(`dd offset jpt_10376+3`)。**沒有把它改成 70**
// ——要排除得直接讀 exe 原始位元組繞過 IDA 的符號化顯示。這裡照實留著。
//
// 另有兩個獨立的邊界巧合佐證解碼正確:表②的 category 42..48 全是 (0,0) 填充,
// 而表③實測到的 category 最大值恰好是 41。兩張表獨立解碼卻在同一邊界吻合。
//
// ⚠ **tech-item 索引 == `Technology` 列舉值**(`OrigTechTopic` 就是直接 `int(tech)` 索引),
// 所以 `TechItemCategory` 可以用 `Technology` 直接查。

var TechResearchLevelValues = [30]int{
	0, 1, 3, 6, 12, 22, 36, 62, 102, 172,
	292, 492, 812, 1372, 2372, 3972, 6372, 9972, 10400, 10800,
	11200, 11600, 12000, 12400, 12800, 13200, 13600, 14000, 14400, 14800,
}

var TechCategoryDefaultMultiplier = [49]int{
	50, 50, 50, 50, 50, 20, 20, 20, 10, 50,
	20, 5, 10, 50, 50, 10, 10, 50, 20, 20,
	5, 20, 50, 10, 10, 10, 20, 5, 20, 5,
	10, 10, 50, 50, 10, 5, 10, 5, 5, 50,
	50, 5, 0, 0, 0, 0, 0, 0, 0,
}

var TechCategoryVar24Flag = [49]int{
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	0, 1, 1, 0, 0, 1, 0, 1, 0, 0,
	0, 0, 1, 0, 0, 0, 0, 1, 0, 1,
	0, 1, 0, 0, 1, 1, 0, 0, 0, 0,
	1, 0, 0, 0, 0, 0, 0, 0, 0,
}

var TechItemCategory = [212]int{
	0, 27, 32, 7, 17, 5, 0, 2, 1, 15,
	19, 18, 21, 28, 8, 9, 11, 35, 5, 4,
	23, 2, 1, 31, 15, 17, 27, 9, 20, 0,
	38, 30, 5, 33, 33, 33, 33, 33, 37, 6,
	9, 22, 13, 12, 25, 33, 29, 26, 20, 1,
	4, 39, 5, 37, 26, 22, 31, 28, 25, 29,
	38, 7, 11, 36, 29, 13, 30, 9, 0, 22,
	26, 19, 18, 16, 7, 3, 2, 13, 26, 26,
	38, 34, 31, 30, 2, 27, 40, 0, 18, 17,
	27, 14, 13, 23, 23, 18, 18, 26, 39, 17,
	26, 16, 28, 8, 26, 38, 21, 6, 17, 9,
	25, 28, 34, 4, 12, 26, 24, 32, 19, 19,
	18, 21, 25, 26, 15, 38, 37, 26, 16, 10,
	10, 1, 9, 9, 10, 3, 2, 26, 16, 21,
	38, 40, 4, 25, 15, 16, 21, 12, 28, 21,
	17, 27, 1, 34, 1, 2, 1, 24, 36, 36,
	24, 34, 0, 5, 3, 28, 12, 39, 9, 9,
	17, 38, 37, 12, 9, 27, 14, 36, 0, 17,
	14, 24, 12, 0, 39, 17, 22, 32, 35, 22,
	35, 32, 35, 6, 39, 40, 36, 5, 0, 28,
	11, 32, 21, 32, 41, 41, 41, 41, 41, 41,
	41, 41,
}

// TechCategoryOf 回傳某項科技所屬的 category(原版 `byte_17E082[techIdx*13]`)。
// 索引越界回 (0, false)。
func TechCategoryOf(tech Technology) (int, bool) {
	i := int(tech)
	if i < 0 || i >= len(TechItemCategory) {
		return 0, false
	}
	return TechItemCategory[i], true
}

// TechCategoryWeight 回傳某項科技的 category 預設倍率(原版 `Calc_Tech_Value_` 階段 B
// 給 `ecx` 的初始值,見 docs/re/calc-tech-value.md)。
//
// ⚠ **這只是估值的起點,不是最終權重。** 原版之後還有十幾段依政體/性格/種族天賦/
// 邊際遞減/上限扣抵的加成(階段 C–N),那些依賴的玩家欄位語意還沒查出來。
// 查不到 category 或 category 越界時回 0。
func TechCategoryWeight(tech Technology) int {
	c, ok := TechCategoryOf(tech)
	if !ok || c < 0 || c >= len(TechCategoryDefaultMultiplier) {
		return 0
	}
	return TechCategoryDefaultMultiplier[c]
}
