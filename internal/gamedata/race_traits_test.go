package gamedata

import "testing"

// SAVE10.GAM 裡五個種族的 Traits[31]——**原版自己寫出來的位元組**。
//
// 這五族是這一整張表的驗收基準:選項等級(RACESTUF.LBX)+ 換算表(執行檔)推出來的
// 結果必須與它們**逐格相同**。任何一格不同,就是展開規則或資料抄錯了。
var save10Traits = map[int][RaceTraitCount]int8{
	0:  {2, 0, 0, 0, 0, 0, 50, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},    // Alkari
	6:  {6, 0, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},     // Klackon
	8:  {2, 0, 0, 0, 0, 0, 0, 50, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},    // Mrrshan
	10: {0, 100, 2, 0, 0, 0, 0, 0, 0, -10, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Sakkra
	12: {2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0},     // Trilarian
}

// 展開規則對得上原版存檔——這是整張表的地基。
func TestExpandedTraitsMatchOriginalSave(t *testing.T) {
	for race, want := range save10Traits {
		got, ok := OrigRaceTraits(race)
		if !ok {
			t.Fatalf("%s 應查得到", OrigRaceEnglishNames[race])
		}
		if got != want {
			t.Errorf("%s 展開結果與 SAVE10.GAM 不符\n  得到 %v\n  應為 %v",
				OrigRaceEnglishNames[race], got, want)
		}
	}
}

// 布林特性**不經換算表**——這是展開規則最容易寫錯的一格。
//
// 轉換迴圈是 `inc edx; cmp dx, 0Ah; jl`,只跑 1..9。照 `[idx*3+level]` 一路換到 30
// 會讀進表後面的別的資料:崔拉里安的「水棲=1」會變成 20,矽基的「食岩=1」會變成 30。
func TestBooleanTraitsAreNotRunThroughTheLevelTable(t *testing.T) {
	cases := []struct {
		race  int
		trait RaceTrait
	}{
		{12, TRAIT_AQUATIC}, {12, TRAIT_TRANS_DIMENSIONAL},
		{11, TRAIT_LITHOVORE}, {11, TRAIT_TOLERANT}, {11, TRAIT_REPULSIVE},
		{8, TRAIT_WARLORD}, {10, TRAIT_SUBTERRANEAN}, {2, TRAIT_STEALTHY_SHIPS},
		{4, TRAIT_FANTASTIC_TRADERS}, {4, TRAIT_LUCKY}, {5, TRAIT_CHARISMATIC},
	}
	for _, c := range cases {
		if got := OrigRaceTrait(c.race, c.trait); got != 1 {
			t.Errorf("%s 的布林特性 %d 應為 1,得到 %d",
				OrigRaceEnglishNames[c.race], c.trait, got)
		}
	}
}

// 政體(特性 0)也不換算,而且用的是與其他三個政體列舉同一套編號(第 54 項(三個寫入端))。
func TestGovernmentTraitIsRawAndSharesTheEnum(t *testing.T) {
	// 手冊逐字:「The Alkari government is a Dictatorship.」
	if got := OrigRaceTrait(0, TRAIT_GOVERNMENT); got != int(MoraleGovDictatorship) {
		t.Errorf("阿爾卡里應是獨裁(%d),得到 %d", MoraleGovDictatorship, got)
	}
	if got := OrigRaceTrait(6, TRAIT_GOVERNMENT); got != int(MoraleGovUnification) {
		t.Errorf("克拉肯應是統一(%d),得到 %d", MoraleGovUnification, got)
	}
	if got := OrigRaceTrait(5, TRAIT_GOVERNMENT); got != int(MoraleGovDemocracy) {
		t.Errorf("人類應是民主(%d),得到 %d", MoraleGovDemocracy, got)
	}
	if got := OrigRaceTrait(10, TRAIT_GOVERNMENT); got != int(MoraleGovFeudalism) {
		t.Errorf("薩克拉應是封建(%d),得到 %d", MoraleGovFeudalism, got)
	}
}

// 手冊行文是第四個獨立來源,逐句對數值型特性。
//
// 這幾句話與 SAVE10 沒有交集(布拉西/埃雷里安/達洛克都不在那份存檔裡),
// 所以它們驗的是**推導出來、沒有存檔背書的那一半**。
func TestNumericTraitsMatchTheManualProse(t *testing.T) {
	cases := []struct {
		race  int
		trait RaceTrait
		want  int
		quote string
	}{
		{0, TRAIT_SHIP_DEFENSE, 50, "Alkari ships gain a +50 defensive bonus in space combat"},
		{1, TRAIT_SHIP_ATTACK, 20, "a +20 to the Ship Attack of all their ships"},
		{1, TRAIT_GROUND_COMBAT, 10, "Bulrathi enjoy a +10 bonus in ground combat"},
		{2, TRAIT_SPYING, 20, "Darlok spies are +20 more likely to be successful"},
		{3, TRAIT_SHIP_DEFENSE, 25, "Elerian ships gain a +25 defensive"},
		{3, TRAIT_SHIP_ATTACK, 20, "and +20 offensive bonus in combat"},
	}
	for _, c := range cases {
		if got := OrigRaceTrait(c.race, c.trait); got != c.want {
			t.Errorf("%s 特性 %d 應為 %d(手冊:%s),得到 %d",
				OrigRaceEnglishNames[c.race], c.trait, c.want, c.quote, got)
		}
	}
}

// 換算表三檔的形狀:1 檔是扣分、2/3 檔是加分且遞增。
//
// 這一條擋的是「表抄歪一格」——歪一格之後個別數字還是合理的,但單調性會壞掉。
func TestPickLevelsAreMonotonic(t *testing.T) {
	numeric := []RaceTrait{
		TRAIT_POPULATION, TRAIT_FARMING, TRAIT_INDUSTRY, TRAIT_SCIENCE, TRAIT_MONEY,
		TRAIT_SHIP_DEFENSE, TRAIT_SHIP_ATTACK, TRAIT_GROUND_COMBAT, TRAIT_SPYING,
	}
	for _, tr := range numeric {
		lv1, lv2, lv3 := RaceTraitPickValue(tr, 1), RaceTraitPickValue(tr, 2), RaceTraitPickValue(tr, 3)
		if lv1 >= 0 {
			t.Errorf("特性 %d 的 1 檔應是扣分選項,得到 %d", tr, lv1)
		}
		if !(lv2 > 0 && lv3 > lv2) {
			t.Errorf("特性 %d 的 2/3 檔應遞增為正:%d → %d", tr, lv2, lv3)
		}
		if RaceTraitPickValue(tr, 0) != 0 {
			t.Errorf("特性 %d 的 0 檔(沒選)應為 0", tr)
		}
	}
}

// 每一族都得有政體、而且至少有一項特性——全 0 的一列表示資料沒讀進來。
func TestEveryRaceHasDataAndAGovernment(t *testing.T) {
	for r := 0; r < OrigRaceCount; r++ {
		tr, ok := OrigRaceTraits(r)
		if !ok {
			t.Fatalf("種族 %d 應查得到", r)
		}
		nonZero := 0
		for i := 1; i < RaceTraitCount; i++ {
			if tr[i] != 0 {
				nonZero++
			}
		}
		if nonZero == 0 {
			t.Errorf("%s 一項特性都沒有,資料可疑", OrigRaceEnglishNames[r])
		}
		if gov := OrigRaceTrait(r, TRAIT_GOVERNMENT); gov < 0 || gov > int(MoraleGovGalacticUnification) {
			t.Errorf("%s 的政體編號 %d 超出範圍", OrigRaceEnglishNames[r], gov)
		}
		if OrigRaceEnglishNames[r] == "" || OrigRaceChineseNames[r] == "" {
			t.Errorf("種族 %d 名稱缺漏", r)
		}
	}
}

// 回傳的是複本,呼叫端改不到表。
func TestOrigRaceTraitsReturnsACopy(t *testing.T) {
	got, _ := OrigRaceTraits(0)
	got[TRAIT_SHIP_DEFENSE] = 99
	if again, _ := OrigRaceTraits(0); again[TRAIT_SHIP_DEFENSE] != 50 {
		t.Errorf("改複本不該污染表,得到 %d", again[TRAIT_SHIP_DEFENSE])
	}
}

// 越界一律回 0 / false,不 panic。
func TestOutOfRangeIsSafe(t *testing.T) {
	if _, ok := OrigRaceTraits(-1); ok {
		t.Error("負索引應回 false")
	}
	if _, ok := OrigRaceTraits(OrigRaceCount); ok {
		t.Error("超界索引應回 false")
	}
	if OrigRaceTrait(-1, TRAIT_FARMING) != 0 || OrigRaceTrait(0, RaceTrait(999)) != 0 {
		t.Error("越界查詢應回 0")
	}
	if OrigRaceHasTrait(99, TRAIT_AQUATIC) {
		t.Error("越界查詢應回 false")
	}
	if RaceTraitPickValue(TRAIT_AQUATIC, 1) != 0 {
		t.Error("布林特性不在換算表上,應回 0")
	}
}
