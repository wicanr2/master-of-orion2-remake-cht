package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// shell.Races 的每一格數值都必須等於 gamedata 那張一手表——這張表不准手改。
//
// 這是第 66 項的主要防線:先前那七個自編數字錯了八族,而**沒有任何測試會紅**,
// 因為每個數字都只有它自己那條斷言在守。改成逐族對照一手表之後,任何一格漂掉都會紅。
func TestEveryRaceMatchesTheOriginalTraitTable(t *testing.T) {
	if len(Races) != gamedata.OrigRaceCount {
		t.Fatalf("種族數應為 %d,得到 %d", gamedata.OrigRaceCount, len(Races))
	}
	seen := map[int]string{}
	for _, r := range Races {
		if prev, dup := seen[r.OrigIdx]; dup {
			t.Fatalf("原版編號 %d 被 %s 與 %s 重複佔用", r.OrigIdx, prev, r.EnName)
		}
		seen[r.OrigIdx] = r.EnName

		want := map[string]struct {
			got   int
			trait gamedata.RaceTrait
		}{
			"工業":   {r.IndBonus, gamedata.TRAIT_INDUSTRY},
			"科研":   {r.ResBonus, gamedata.TRAIT_SCIENCE},
			"農業":   {r.FoodBonus, gamedata.TRAIT_FARMING},
			"成長":   {r.GrowthPct, gamedata.TRAIT_POPULATION},
			"每人BC": {r.IncomePerPop, gamedata.TRAIT_MONEY},
			"艦攻":   {r.CombatPct, gamedata.TRAIT_SHIP_ATTACK},
			"艦防":   {r.ShipDefPct, gamedata.TRAIT_SHIP_DEFENSE},
			"地面戰":  {r.GroundCombatBonus, gamedata.TRAIT_GROUND_COMBAT},
			"諜報":   {r.SpyBonus, gamedata.TRAIT_SPYING},
		}
		for label, c := range want {
			if exp := gamedata.OrigRaceTrait(r.OrigIdx, c.trait); c.got != exp {
				t.Errorf("%s 的%s應為 %d(一手表),得到 %d", r.Name, label, exp, c.got)
			}
		}
	}
	if len(seen) != gamedata.OrigRaceCount {
		t.Errorf("13 個原版編號應各出現一次,實得 %d 個", len(seen))
	}
}

// 原版種族編號要對得上種族**選擇畫面**的肖像順序,不能只是自洽。
//
// 兩邊各自獨立:`cmd/moo2/raceselect.go` 的順序是量原版 RACESEL 資產排出來的,
// 這裡的 OrigIdx 是從 RACESTUF.LBX 讀出來的。名字對得上才表示兩邊指的是同一族。
func TestOrigIdxMatchesTheEnglishNameInGamedata(t *testing.T) {
	// shell 的英文名是複數形(Humans / Klackons),gamedata 那張表是原版單數形。
	plural := map[string]string{
		"Humans": "Human", "Psilons": "Psilon", "Klackons": "Klackon",
		"Meklars": "Meklar", "Darloks": "Darlok", "Trilarians": "Trilarian",
		"Elerians": "Elerian", "Gnolams": "Gnolam", "Silicoids": "Silicoid",
	}
	for _, r := range Races {
		want := gamedata.OrigRaceEnglishNames[r.OrigIdx]
		got := r.EnName
		if s, ok := plural[got]; ok {
			got = s
		}
		if got != want {
			t.Errorf("%s 的 OrigIdx=%d 指到 %q,對不上", r.EnName, r.OrigIdx, want)
		}
		if zh := gamedata.OrigRaceChineseNames[r.OrigIdx]; zh != r.Name {
			t.Errorf("%s 的中文名對不上一手表:%q vs %q", r.EnName, r.Name, zh)
		}
	}
}

// 艦艇攻擊與防禦是**兩個獨立特性**,壓成一個就表達不出阿爾卡里/姆瑞森的差別。
func TestShipAttackAndDefenseAreSeparate(t *testing.T) {
	idx := func(en string) int {
		for i, r := range Races {
			if r.EnName == en {
				return i
			}
		}
		t.Fatalf("找不到 %s", en)
		return -1
	}
	// 阿爾卡里:只有防禦。
	s := NewDemoSession()
	s.ApplyRace(idx("Alkari"))
	if s.RaceShipDefPct != 50 || s.RaceCombatPct != 0 {
		t.Errorf("阿爾卡里應是 防+50 / 攻 0,得到 防%d / 攻%d", s.RaceShipDefPct, s.RaceCombatPct)
	}
	// 埃雷里安:兩者都有,而且**值不同**(防 25、攻 20)。
	s = NewDemoSession()
	s.ApplyRace(idx("Elerians"))
	if s.RaceShipDefPct != 25 || s.RaceCombatPct != 20 {
		t.Errorf("埃雷里安應是 防+25 / 攻+20,得到 防%d / 攻%d", s.RaceShipDefPct, s.RaceCombatPct)
	}
}

// 種族艦艇防禦真的接進戰列(不是只存在 GameSession 欄位裡)。
func TestShipDefenseBonusReachesTheBattleLine(t *testing.T) {
	idx := func(en string) int {
		for i, r := range Races {
			if r.EnName == en {
				return i
			}
		}
		t.Fatalf("找不到 %s", en)
		return -1
	}
	// ⚠ 用**戰艦**不用 demo 艦隊的第一艘:那是偵察艦,艦體值 1,+50% 整數截斷後是 0。
	// 那不是接線壞了,是百分比模型在極小值上本來就看不出來——但拿它當測試前提就測不到東西。
	last := func(s *GameSession) combatant {
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{Name: "測試戰艦", Class: "戰艦"})
		cs, _ := s.mkPlayerCombatantsIndexed()
		if len(cs) == 0 {
			t.Fatal("測試前提不成立:應有可戰艦艇")
		}
		return cs[len(cs)-1]
	}
	plain := last(NewDemoSession())
	alkari := NewDemoSession()
	alkari.ApplyRace(idx("Alkari"))
	boosted := last(alkari)

	if want := plain.def + shipStrength("戰艦")*50/100; boosted.def != want {
		t.Errorf("阿爾卡里的防禦應為 %d(基準 %d + 50%%),得到 %d", want, plain.def, boosted.def)
	}
	// 反面:它不該連攻擊也一起變高——那正是壓成單一 CombatPct 的舊行為。
	if boosted.atk != plain.atk {
		t.Errorf("阿爾卡里不該有攻擊加成:%d vs %d", boosted.atk, plain.atk)
	}
}

// 諾蘭姆的低重力懲罰**只扣一次**。
//
// 先前 `GroundRaceCombatBonus(Gnolam)` 回 −10、`GroundApplyLowGPenalty` 又扣 −10,
// 合計 −20;而反組譯只寫了一次(`mov byte ptr [ecx+0Dh], 0F6h`)。
func TestGnolamLowGPenaltyIsAppliedOnce(t *testing.T) {
	idx := func(en string) int {
		for i, r := range Races {
			if r.EnName == en {
				return i
			}
		}
		t.Fatalf("找不到 %s", en)
		return -1
	}
	base := NewDemoSession()
	base.ApplyRace(idx("Humans")) // 人類沒有任何重力/地面特性
	gnolam := NewDemoSession()
	gnolam.ApplyRace(idx("Gnolams"))

	diff := base.playerMarineForce() - gnolam.playerMarineForce()
	if want := -gamedata.GroundLowGCombatPenalty; diff != want {
		t.Errorf("諾蘭姆應只比人類低 %d 點(低重力扣一次),實低 %d 點", want, diff)
	}
	if gnolam.RaceGroundBonus != 0 {
		t.Errorf("諾蘭姆的 TRAIT_GROUND_COMBAT 是 0,不該另有地面加成,得到 %d", gnolam.RaceGroundBonus)
	}
}

// 布拉西的地面戰 +10 有接上,而且它是**定值**不是百分比。
func TestBulrathiGroundBonusIsFlat(t *testing.T) {
	idx := func(en string) int {
		for i, r := range Races {
			if r.EnName == en {
				return i
			}
		}
		t.Fatalf("找不到 %s", en)
		return -1
	}
	base := NewDemoSession()
	base.ApplyRace(idx("Humans"))
	bul := NewDemoSession()
	bul.ApplyRace(idx("Bulrathi"))
	if diff := bul.playerMarineForce() - base.playerMarineForce(); diff != 10 {
		t.Errorf("布拉西地面戰應 +10 定值,實得 %+d", diff)
	}
}

// 薩克拉的地底加成**只在守自家殖民地時**生效。
//
// 這一條同時是「擋門理由過期」的收尾:先前檔頭寫「13 個標準種族也沒有一個具備
// Subterranean/High-G」——薩克拉有地底、布拉西有高重力,兩句都不成立。
func TestSakkraSubterraneanOnlyWhenDefending(t *testing.T) {
	idx := func(en string) int {
		for i, r := range Races {
			if r.EnName == en {
				return i
			}
		}
		t.Fatalf("找不到 %s", en)
		return -1
	}
	s := NewDemoSession()
	s.ApplyRace(idx("Sakkra"))
	atk, def := s.playerMarineForce(), s.playerDefendingMarineForce()
	if want := gamedata.GroundSubterraneanDefenseBonus; def-atk != want {
		t.Errorf("薩克拉守方應多 %d,實多 %d", want, def-atk)
	}
	// 反面:沒有地底特性的種族,攻守兩版應相同。
	h := NewDemoSession()
	h.ApplyRace(idx("Humans"))
	if h.playerMarineForce() != h.playerDefendingMarineForce() {
		t.Error("人類沒有地底特性,攻守兩版應相同")
	}
}

// 布拉西的高重力讓陸戰隊多挨一下才死。
func TestBulrathiHighGTakesAnExtraHit(t *testing.T) {
	idx := func(en string) int {
		for i, r := range Races {
			if r.EnName == en {
				return i
			}
		}
		t.Fatalf("找不到 %s", en)
		return -1
	}
	base := NewDemoSession()
	base.ApplyRace(idx("Humans"))
	bul := NewDemoSession()
	bul.ApplyRace(idx("Bulrathi"))
	if !bul.raceHasTrait(gamedata.TRAIT_HIGH_G) {
		t.Fatal("布拉西應有高重力特性")
	}
	if base.raceHasTrait(gamedata.TRAIT_HIGH_G) {
		t.Fatal("人類不該有高重力特性")
	}
	got := gamedata.GroundMarineHitsToKill(bul.raceHasTrait(gamedata.TRAIT_HIGH_G), false)
	want := gamedata.GroundMarineHitsToKill(false, false) + gamedata.GroundHighGRaceExtraHit
	if got != want {
		t.Errorf("高重力陸戰隊應多挨 %d 下:%d vs %d", gamedata.GroundHighGRaceExtraHit, got, want)
	}
}

// 達洛克的諜報 +20 與薩克拉的 −10 都接進攻守兩側。
func TestSpyRaceBonusReachesBothSides(t *testing.T) {
	idx := func(en string) int {
		for i, r := range Races {
			if r.EnName == en {
				return i
			}
		}
		t.Fatalf("找不到 %s", en)
		return -1
	}
	dar := NewDemoSession()
	dar.ApplyRace(idx("Darloks"))
	if dar.RaceSpyBonus != 20 {
		t.Errorf("達洛克諜報應 +20,得到 %d", dar.RaceSpyBonus)
	}
	sak := NewDemoSession()
	sak.ApplyRace(idx("Sakkra"))
	if sak.RaceSpyBonus != -10 {
		t.Errorf("薩克拉諜報應 −10,得到 %d", sak.RaceSpyBonus)
	}
	// 攻守兩側都吃得到(手冊那張表兩欄同值,見 spyTechBonusFor)。
	none := clonePlayerState(NewDemoSession().Player)
	if a := spyAttackerBonus(none, 0, dar.RaceSpyBonus) - spyAttackerBonus(none, 0, 0); a != 20 {
		t.Errorf("攻擊側種族加成應 +20,得到 %+d", a)
	}
	if d := spyDefenderBonus(none, 0, dar.RaceSpyBonus) - spyDefenderBonus(none, 0, 0); d != 20 {
		t.Errorf("防守側種族加成應 +20,得到 %+d", d)
	}
}

// 自訂種族不會誤查到某一族的布林特性。
//
// `raceOrigIdx()` 對 RaceIndex 越界回 −1,而 `OrigRaceHasTrait` 對越界回 false
// ——寧可少給也不要亂給(自訂種族的特殊能力目前只記錄不生效)。
func TestCustomRaceHasNoBooleanTraits(t *testing.T) {
	s := NewDemoSession()
	s.ApplyCustomRaceBonuses(Race{Name: "測試自訂", EnName: "Custom", OrigIdx: -1, IndBonus: 3})
	if s.raceOrigIdx() != -1 {
		t.Fatalf("自訂種族的 OrigIdx 應為 −1,得到 %d", s.raceOrigIdx())
	}
	for _, tr := range []gamedata.RaceTrait{
		gamedata.TRAIT_LOW_G, gamedata.TRAIT_HIGH_G, gamedata.TRAIT_SUBTERRANEAN,
		gamedata.TRAIT_AQUATIC, gamedata.TRAIT_TOLERANT, gamedata.TRAIT_WARLORD,
	} {
		if s.raceHasTrait(tr) {
			t.Errorf("自訂種族不該查到特性 %d", tr)
		}
	}
}

// raceIndexByEnName 查 shell.Races 的索引(測試共用)。
func raceIndexByEnName(t *testing.T, en string) int {
	t.Helper()
	for i, r := range Races {
		if r.EnName == en {
			return i
		}
	}
	t.Fatalf("找不到種族 %s", en)
	return -1
}
