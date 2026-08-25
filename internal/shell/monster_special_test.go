package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestPlasmaFluxRangeAndSquaredDistanceFalloff(t *testing.T) {
	if PlasmaFluxRadiusPixels != 96 || !PlasmaFluxInRange(4, 2) || PlasmaFluxInRange(5, 0) {
		t.Fatalf("Plasma Flux 應使用 96px 歐氏半徑：r=%d in(4,2)=%v in(5,0)=%v",
			PlasmaFluxRadiusPixels, PlasmaFluxInRange(4, 2), PlasmaFluxInRange(5, 0))
	}
	if got := PlasmaFluxAttenuatedDamage(100, 3, 0); got != 61 {
		t.Fatalf("三格距離平方衰減=%d，預期 61", got)
	}
	if got := PlasmaFluxAttenuatedDamage(100, 5, 0); got != 0 {
		t.Fatalf("圈外傷害=%d，預期 0", got)
	}
}

func TestPlasmaFluxSizeDamageUsesDoubleRollPerSegment(t *testing.T) {
	rolls := []int{1, 2, 7, 3, 5, 1, 4, 6}
	i := 0
	got := PlasmaFluxSizeDamage(10, gamedata.SHIP_TITAN, func(limit int) int {
		value := rolls[i]
		i++
		if value > limit {
			t.Fatalf("測試擲值 %d 超過 %d", value, limit)
		}
		return value
	})
	// 五段：1；首擲2後取7；首擲3後取5；1；首擲4後取6。
	if got != 20 || i != 8 {
		t.Fatalf("尺寸分段傷害=%d draws=%d，預期 20/8", got, i)
	}
}

func TestPlasmaFluxFighterCasualtiesAvoidAndRollEachCraft(t *testing.T) {
	draws := 0
	if got := PlasmaFluxFighterCasualties(4, 2, 4, 1, func(limit int) int {
		draws++
		return 1
	}); got != 0 || draws != 0 {
		t.Fatalf("整隊迴避應為 0 傷亡且不擲逐架骰：killed=%d draws=%d", got, draws)
	}
	rolls := []int{10, 60, 50, 51} // chance=25*4/2=50
	i := 0
	got := PlasmaFluxFighterCasualties(4, 2, 4, 2, func(limit int) int {
		value := rolls[i]
		i++
		return value
	})
	if got != 2 || i != 4 {
		t.Fatalf("逐架傷亡=%d draws=%d，預期 2/4", got, i)
	}
}

func TestCausticSlimeStacksTicksFourFacingsAndDecays(t *testing.T) {
	target := CombatShip{HP: 200, MaxHP: 200, ArmorHP: 100, SizeClass: gamedata.SHIP_FRIGATE}
	AddCausticSlimeStrength(&target, 25)
	AddCausticSlimeStrength(&target, 10)
	if target.CausticSlimeStrength != 35 {
		t.Fatalf("黏液命中應累加，得到 %d", target.CausticSlimeStrength)
	}
	if got := TickCausticSlime(&target); got != 40 {
		t.Fatalf("四面各 35 點應先耗 100 裝甲再傷 40 結構，得到 %d", got)
	}
	if target.ArmorHP != 0 || target.HP != 160 || target.CausticSlimeStrength != 30 {
		t.Fatalf("回合狀態錯誤：armor=%d hp=%d slime=%d", target.ArmorHP, target.HP, target.CausticSlimeStrength)
	}
}

func TestCausticSlimeUsesEveryShieldFacingAndClampsDecay(t *testing.T) {
	target := CombatShip{HP: 100, MaxHP: 100, ShieldReduction: 5, SizeClass: gamedata.SHIP_FRIGATE,
		CausticSlimeStrength: 4}
	target.EnsureShieldFacings()
	before := target.ShieldFacingHP
	if got := TickCausticSlime(&target); got != 0 {
		t.Fatalf("四面護盾足以吸收時不應傷結構，得到 %d", got)
	}
	for i := range target.ShieldFacingHP {
		if target.ShieldFacingHP[i] >= before[i] {
			t.Fatalf("第 %d 面護盾未被黏液消耗：before=%d after=%d", i, before[i], target.ShieldFacingHP[i])
		}
	}
	if target.CausticSlimeStrength != 0 {
		t.Fatalf("小於 5 的殘量應夾到零，得到 %d", target.CausticSlimeStrength)
	}
}
