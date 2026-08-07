package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 手冊那個括號裡的例子把公式釘死了:「6 beams would immobilize a Doom Star」。
//
// 末日之星是第 5 級,6 = 5+1;而「Each beam can trap a **Frigate** class ship」→ 第 0 級 = 1 束。
// **兩端都對上,中間就是線性**。
func TestTractorBeamsToImmobilizeMatchesTheManualExample(t *testing.T) {
	want := map[gamedata.CombatShipClass]int{0: 1, 1: 2, 2: 3, 3: 4, 4: 5, 5: 6}
	for class, n := range want {
		if got := gamedata.TractorBeamsToImmobilize(class); got != n {
			t.Errorf("第 %d 級應需要 %d 束,得到 %d", class, n, got)
		}
	}
	// 手冊逐字的兩個錨點,分開再驗一次(上面那張表是內插出來的)。
	if gamedata.TractorBeamsToImmobilize(5) != 6 {
		t.Error("末日之星應需要 6 束(手冊原文的例子)")
	}
	if gamedata.TractorBeamsToImmobilize(0) != 1 {
		t.Error("巡防艦應一束就定住(手冊:Each beam can trap a Frigate class ship)")
	}
}

// 拖慢是「按比例」而且可累積(手冊:in proportion to its size / cumulative)。
func TestTractorSlowingIsProportionalAndCumulative(t *testing.T) {
	const speed = 24
	const class = gamedata.CombatShipClass(3) // 戰艦:需要 4 束
	prev := speed
	for beams := 1; beams <= 4; beams++ {
		got := gamedata.TractorSlowedSpeed(speed, class, beams)
		if got >= prev {
			t.Errorf("%d 束時應更慢:%d → %d", beams, prev, got)
		}
		prev = got
	}
	if prev != 0 {
		t.Errorf("束數達到需求時應完全定住,得到 %d", prev)
	}
	// 超過需求仍是 0,不會變負。
	if got := gamedata.TractorSlowedSpeed(speed, class, 99); got != 0 {
		t.Errorf("超量牽引仍應為 0,得到 %d", got)
	}
	// 沒有牽引就不動它。
	if got := gamedata.TractorSlowedSpeed(speed, class, 0); got != speed {
		t.Errorf("沒有牽引時速度不該變:%d vs %d", got, speed)
	}
}

// 只有**完全定住**才吃 −20 防禦(手冊那句的主詞是 immobilized)。
func TestOnlyImmobilizedShipsTakeTheDefensePenalty(t *testing.T) {
	sh := CombatShip{Name: "戰艦", Defense: 50, SizeClass: 3, CombatSpeed: 24}
	if got := TacticalEffectiveDefense(sh); got != 50 {
		t.Errorf("沒被牽引時防禦不變,得到 %d", got)
	}
	slowed := sh
	slowed.HeldByTractors = 2 // 拖慢但還能動
	if got := TacticalEffectiveDefense(slowed); got != 50 {
		t.Errorf("只是被拖慢不該扣防禦,得到 %d", got)
	}
	held := sh
	held.HeldByTractors = 4 // 完全定住
	if want := 50 + gamedata.TractorImmobilizedDefensePenalty; TacticalEffectiveDefense(held) != want {
		t.Errorf("完全定住應扣 %d,得到 %d", -gamedata.TractorImmobilizedDefensePenalty,
			TacticalEffectiveDefense(held))
	}
	// 不會扣成負的。
	weak := CombatShip{Defense: 5, SizeClass: 0, HeldByTractors: 1}
	if got := TacticalEffectiveDefense(weak); got < 0 {
		t.Errorf("防禦不該為負,得到 %d", got)
	}
}

// 停滯力場讓速度歸零(手冊:the ship cannot move)。
func TestStasisStopsMovementEntirely(t *testing.T) {
	sh := CombatShip{SizeClass: 0, CombatSpeed: 30}
	if TacticalEffectiveSpeed(sh) != 30 {
		t.Fatalf("測試前提不成立:未受影響時應為 30")
	}
	sh.InStasis = true
	if got := TacticalEffectiveSpeed(sh); got != 0 {
		t.Errorf("被停滯力場定住應完全不能動,得到 %d", got)
	}
}

// 狀態每回合**重算**:產生源被打掉、或目標飛出射程,效果就該消失。
func TestStatusEffectsAreRecomputedNotAccumulated(t *testing.T) {
	tractorRange := TacticalRangeSquares(gamedata.TractorBeamRangeSquares)
	mine := []CombatShip{{Name: "牽引艦", HP: 10, Col: 0, Row: 0, TractorBeams: 1}}
	theirs := []CombatShip{{Name: "目標", HP: 10, Col: tractorRange, Row: 0, SizeClass: 0, CombatSpeed: 20}}

	ApplyTacticalStatusEffects(mine, theirs)
	if theirs[0].HeldByTractors != 1 {
		t.Fatalf("射程內應被牽引,得到 %d", theirs[0].HeldByTractors)
	}

	// ① 目標飛出射程 → 鬆開
	theirs[0].Col = tractorRange + 5
	ApplyTacticalStatusEffects(mine, theirs)
	if theirs[0].HeldByTractors != 0 {
		t.Errorf("飛出射程應鬆開,得到 %d", theirs[0].HeldByTractors)
	}

	// ② 產生源被擊毀 → 鬆開(即使還在射程內)
	theirs[0].Col = 0
	mine[0].HP = 0
	ApplyTacticalStatusEffects(mine, theirs)
	if theirs[0].HeldByTractors != 0 {
		t.Errorf("產生源已被擊毀,效果應消失,得到 %d", theirs[0].HeldByTractors)
	}
}

// 多艘船的牽引光束可累加——remake 一艘船只有一個 Special 槽,所以只能靠數量。
func TestMultipleShipsStackTractorBeams(t *testing.T) {
	var mine []CombatShip
	for i := 0; i < 6; i++ {
		mine = append(mine, CombatShip{Name: "拖船", HP: 10, Col: 0, Row: 0, TractorBeams: 1})
	}
	doom := []CombatShip{{Name: "末日之星", HP: 100, Col: 0, Row: 0, SizeClass: 5, CombatSpeed: 13}}
	ApplyTacticalStatusEffects(mine, doom)
	if doom[0].HeldByTractors != 6 {
		t.Fatalf("六艘拖船應累加成 6 束,得到 %d", doom[0].HeldByTractors)
	}
	if TacticalEffectiveSpeed(doom[0]) != 0 {
		t.Error("手冊原文:6 beams would immobilize a Doom Star")
	}
	// 五束還動得了(這一條讓上面那個 0 有意義)。
	ApplyTacticalStatusEffects(mine[:5], doom)
	if TacticalEffectiveSpeed(doom[0]) == 0 {
		t.Error("五束不該定住末日之星")
	}
}

// 停滯力場的射程比牽引光束短——這是換算後仍必須保住的相對關係。
func TestStasisRangeIsShorterThanTractorRange(t *testing.T) {
	tr := TacticalRangeSquares(gamedata.TractorBeamRangeSquares)
	st := TacticalRangeSquares(gamedata.StasisFieldRangeSquares)
	if st < 1 {
		t.Errorf("短射程向下取整會變成 0(等於這個武器不存在),得到 %d", st)
	}
	if tr <= st {
		t.Errorf("牽引光束(原版 12 格)應比停滯力場(3 格)遠:%d vs %d", tr, st)
	}
	if got := TacticalRangeSquares(0); got != 0 {
		t.Errorf("0 應回 0,得到 %d", got)
	}
}

// 兩個新元件的主題對得上執行檔。
func TestStatusWeaponsHaveRealTopics(t *testing.T) {
	want := map[string]gamedata.Technology{
		tractorBeamName: gamedata.TECH_TRACTOR_BEAM,
		stasisFieldName: gamedata.TECH_STASIS_FIELD,
	}
	have := map[string]Component{}
	for _, c := range SpecialOptions {
		have[c.Name] = c
	}
	for name, tech := range want {
		c, ok := have[name]
		if !ok {
			t.Errorf("%s 應出現在 SpecialOptions", name)
			continue
		}
		if c.UnlockTech != tech {
			t.Errorf("%s 的解鎖科技對不上:%v", name, c.UnlockTech)
		}
		topic, ok := gamedata.OrigTechTopic(tech)
		if !ok {
			t.Errorf("執行檔應查得到 %s 的主題", name)
			continue
		}
		if c.Tech != topic {
			t.Errorf("%s 的主題應為 %v,得到 %v", name, topic, c.Tech)
		}
	}
}

// 裝了這兩個系統的船,在格子戰術裡真的帶著投射能力。
func TestStatusWeaponsReachTheTacticalShip(t *testing.T) {
	for _, c := range []struct {
		name    string
		tractor int
		stasis  bool
	}{{tractorBeamName, 1, false}, {stasisFieldName, 0, true}} {
		s := NewDemoSession()
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{
			Name: "測試艦", Class: "戰艦", Weapon: "雷射砲", Special: c.name,
		})
		player, _ := s.StartCombat("測試敵人")
		last := player[len(player)-1]
		if last.TractorBeams != c.tractor || last.HasStasisField != c.stasis {
			t.Errorf("%s:應為 牽引=%d 停滯=%v,得到 牽引=%d 停滯=%v",
				c.name, c.tractor, c.stasis, last.TractorBeams, last.HasStasisField)
		}
		if last.SizeClass != gamedata.CombatShipClass(3) {
			t.Errorf("%s:戰艦的級數應為 3,得到 %d", c.name, last.SizeClass)
		}
	}
}
