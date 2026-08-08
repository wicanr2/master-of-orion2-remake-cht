package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 三個系統各自只加速自己那一類武器(手冊逐字)。
func TestDoubleFireAppliesToTheRightWeaponKind(t *testing.T) {
	cases := []struct {
		special      string
		beam, missil int // 光束艦 / 飛彈艦 各能開幾次
	}{
		{"", 1, 1},
		{hyperXCapacitorsName, 2, 1},    // 手冊:allows a ship's **beam weapons** to fire twice
		{fastMissileRacksName, 1, 2},    // 手冊:two volleys of **missiles**
		{timeWarpFacilitatorName, 2, 2}, // 手冊:an additional round of **activity**(不分武器)
	}
	for _, c := range cases {
		beam := CombatShip{Kind: WeaponKindBeam, Charged: true, ShotsKind: shipShotsKind(Ship{Special: c.special})}
		mis := CombatShip{Kind: WeaponKindMissile, Charged: true, ShotsKind: beam.ShotsKind}
		if got := TacticalShotsThisRound(beam); got != c.beam {
			t.Errorf("%q 的光束艦應開 %d 次,得到 %d", c.special, c.beam, got)
		}
		if got := TacticalShotsThisRound(mis); got != c.missil {
			t.Errorf("%q 的飛彈艦應開 %d 次,得到 %d", c.special, c.missil, got)
		}
	}
}

// 冷卻的讀法:手冊的 unused 是「**完全沒開火**」,不是「沒有連射」。
//
// 這個差別在對戰裡很明顯:玩家不能靠「連射 → 單射 → 連射」無限循環。
func TestRechargeRequiresAFullyIdleRound(t *testing.T) {
	ships := []CombatShip{{
		Kind: WeaponKindBeam, Charged: true,
		ShotsKind: gamedata.ShotsDoubleBeam,
	}}
	// 第 1 回合:連射
	if got := TacticalShotsThisRound(ships[0]); got != 2 {
		t.Fatalf("滿電時應連射,得到 %d", got)
	}
	ships[0].Fired = true
	TacticalAdvanceCharge(ships)
	if ships[0].Charged {
		t.Fatal("連射之後應該沒電")
	}
	// 第 2 回合:只能單射;而且**開了火就充不到電**
	if got := TacticalShotsThisRound(ships[0]); got != 1 {
		t.Errorf("沒電時應只能單射,得到 %d", got)
	}
	ships[0].Fired = true
	TacticalAdvanceCharge(ships)
	if ships[0].Charged {
		t.Error("單射也算開火,不該充到電(手冊:remain **unused** for 1 turn)")
	}
	// 第 3 回合:完全不開火 → 充電
	ships[0].Fired = false
	TacticalAdvanceCharge(ships)
	if !ships[0].Charged {
		t.Error("整回合沒開火應該充飽")
	}
	if got := TacticalShotsThisRound(ships[0]); got != 2 {
		t.Errorf("充飽後應可再連射,得到 %d", got)
	}
}

// 時間扭曲加速器**沒有冷卻**——手冊沒有那句 unused 的限制。
func TestTimeWarpHasNoCooldown(t *testing.T) {
	ships := []CombatShip{{Kind: WeaponKindBeam, Charged: true, ShotsKind: gamedata.ShotsDoubleAny}}
	for round := 1; round <= 4; round++ {
		if got := TacticalShotsThisRound(ships[0]); got != 2 {
			t.Fatalf("第 %d 回合仍應開兩次(手冊:two combat rounds for every one),得到 %d", round, got)
		}
		ships[0].Fired = true
		TacticalAdvanceCharge(ships)
	}
	if gamedata.ShotsNeedsCooldown(gamedata.ShotsDoubleAny) {
		t.Error("時間扭曲加速器不該需要冷卻")
	}
	if !gamedata.ShotsNeedsCooldown(gamedata.ShotsDoubleBeam) ||
		!gamedata.ShotsNeedsCooldown(gamedata.ShotsDoubleMissile) {
		t.Error("超載電容與快速飛彈架都該需要冷卻")
	}
}

// 被停滯力場定住的船一次都不能開(第 70 項與這一項的交界)。
func TestStasisOverridesDoubleFire(t *testing.T) {
	sh := CombatShip{Kind: WeaponKindBeam, Charged: true, ShotsKind: gamedata.ShotsDoubleAny, InStasis: true}
	if got := TacticalShotsThisRound(sh); got != 0 {
		t.Errorf("被定住應完全不能開火,得到 %d", got)
	}
}

// 快速結算那一側:沒有這些系統的船 shots 恆為 1(RNG 消耗不變)。
func TestQuickResolutionShotsDefaultToOne(t *testing.T) {
	s := NewDemoSession()
	cs, _ := s.mkPlayerCombatantsIndexed()
	if len(cs) == 0 {
		t.Fatal("測試前提不成立:應有可戰艦艇")
	}
	for _, c := range cs {
		if c.shots != 1 {
			t.Errorf("一般艦艇應只開一次火,得到 %d", c.shots)
		}
	}
	// 裝了時間扭曲加速器的船在快速結算裡也打兩次。
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{
		Name: "加速艦", Class: "戰艦", Weapon: "雷射砲", Special: timeWarpFacilitatorName,
	})
	cs2, _ := s.mkPlayerCombatantsIndexed()
	if got := cs2[len(cs2)-1].shots; got != 2 {
		t.Errorf("時間扭曲加速器在快速結算也該開兩次,得到 %d", got)
	}
}

// 三個新元件的主題對得上執行檔。
func TestShotsSystemsHaveRealTopics(t *testing.T) {
	want := map[string]gamedata.Technology{
		fastMissileRacksName:    gamedata.TECH_FAST_MISSILE_RACKS,
		hyperXCapacitorsName:    gamedata.TECH_HYPERX_CAPACITORS,
		timeWarpFacilitatorName: gamedata.TECH_TIME_WARP_FACILITATOR,
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
