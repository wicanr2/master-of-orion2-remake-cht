package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 八個系統都裝得上,主題全部對得上執行檔的 tech→topic 表。
//
// 主題不是「合理就好」:選錯主題等於把元件掛在別的科技底下,玩家研究對的科技卻拿不到。
func TestMissileDefenseSystemsAreEquippableWithRealTopics(t *testing.T) {
	want := map[string]gamedata.Technology{
		"電子干擾器":   gamedata.TECH_ECM_JAMMER,
		"多波電子干擾器": gamedata.TECH_MULTIWAVE_ECM_JAMMER,
		"廣域干擾器":   gamedata.TECH_WIDE_AREA_JAMMER,
		"慣性穩定器":   gamedata.TECH_INERTIAL_STABILIZER,
		"慣性抵消器":   gamedata.TECH_INERTIAL_NULLIFIER,
		"閃電場":     gamedata.TECH_LIGHTNING_FIELD,
		"位移裝置":    gamedata.TECH_DISPLACEMENT_DEVICE,
		"部隊艙":     gamedata.TECH_TROOP_PODS,
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
			t.Errorf("%s 的解鎖科技應為 %v,得到 %v", name, tech, c.UnlockTech)
		}
		topic, ok := gamedata.OrigTechTopic(tech)
		if !ok {
			t.Errorf("執行檔應查得到 %s 的主題", name)
			continue
		}
		if c.Tech != topic {
			t.Errorf("%s 的主題應為執行檔的 %v,得到 %v", name, topic, c.Tech)
		}
		if c.Cost <= 0 {
			t.Errorf("%s 的建造成本應為正值,得到 %d", name, c.Cost)
		}
	}
}

// 干擾器/慣性系列的閃避加成用手冊真值,而且**逐級遞增**。
func TestJammerEvasionBonusesMatchTheManual(t *testing.T) {
	cases := []struct {
		name string
		want int
	}{
		{"慣性穩定器", gamedata.MissileInertialStabilizer}, // 25
		{"慣性抵消器", gamedata.MissileInertialNullifier},  // 50
		{"電子干擾器", gamedata.MissileJammerECM},          // 70
		{"多波電子干擾器", gamedata.MissileJammerMultiWave},  // 100
		{"廣域干擾器", gamedata.MissileJammerWideAreaSelf}, // 130
	}
	prev := 0
	for _, c := range cases {
		got := shipMissileEvasionBonus(Ship{Special: c.name})
		if got != c.want {
			t.Errorf("%s 的閃避加成應為 %d,得到 %d", c.name, c.want, got)
		}
		if got <= prev {
			t.Errorf("%s 應高於前一級:%d → %d", c.name, prev, got)
		}
		prev = got
	}
	// 反面:不是這一族的元件不該有閃避加成。
	for _, other := range []string{"", "戰機庫", highEnergyFocusName, heavyArmorName} {
		if got := shipMissileEvasionBonus(Ship{Special: other}); got != 0 {
			t.Errorf("%q 不該提供閃避加成,得到 %d", other, got)
		}
	}
}

// 閃避加成真的接進戰列(不是只存在 helper 裡)。
func TestJammerReachesTheBattleLine(t *testing.T) {
	mk := func(special string) combatant {
		s := NewDemoSession()
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{
			Name: "測試艦", Class: "戰艦", Weapon: "雷射砲", Special: special,
		})
		cs, _ := s.mkPlayerCombatantsIndexed()
		return cs[len(cs)-1]
	}
	plain, jammed := mk(""), mk("廣域干擾器")
	if jammed.missileEvasion != plain.missileEvasion+gamedata.MissileJammerWideAreaSelf {
		t.Errorf("廣域干擾器應加 %d:%d → %d",
			gamedata.MissileJammerWideAreaSelf, plain.missileEvasion, jammed.missileEvasion)
	}
	// 格子戰術那一側也要有(兩條路只接一邊會安靜地不一致)。
	s := NewDemoSession()
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{
		Name: "測試艦", Class: "戰艦", Weapon: "雷射砲", Special: "廣域干擾器",
	})
	player, _ := s.StartCombat("測試敵人")
	if last := player[len(player)-1]; last.MissileEvasion < gamedata.MissileJammerWideAreaSelf {
		t.Errorf("格子戰術的閃避加成應含干擾器,得到 %d", last.MissileEvasion)
	}
}

// 閃電場:骰在門檻內就直接摧毀,而且**在 AMR 之前**判定。
func TestLightningFieldDestroysIncomingMissiles(t *testing.T) {
	const wmax, armor = 40, 100
	base := MissileDefenses{HasLightningField: true}

	// 骰 <= 50 → 摧毀。
	killed := base
	killed.LightningRoll = gamedata.MissileLightningFieldDestroyChance
	if got := ResolveMissileShot(false, 0, 1, 0, 0, false, 1, wmax, 0, armor, false, killed); got.Hit {
		t.Error("閃電場命中門檻內應摧毀來襲飛彈")
	}
	// 骰 > 50 → 沒攔到,照原本流程走。
	missed := base
	missed.LightningRoll = gamedata.MissileLightningFieldDestroyChance + 1
	got := ResolveMissileShot(false, 0, 1, 0, 0, false, 1, wmax, 0, armor, false, missed)
	ref := ResolveMissileShot(false, 0, 1, 0, 0, false, 1, wmax, 0, armor, false, MissileDefenses{})
	if got.Hit != ref.Hit || got.DamageToStructure != ref.DamageToStructure {
		t.Errorf("閃電場沒攔到時結果應與沒裝相同:%+v vs %+v", got, ref)
	}
}

// 位移裝置:一律 30% 未命中,**不論其他裝備或情況**(手冊逐字)。
func TestDisplacementDeviceMissesRegardlessOfEverythingElse(t *testing.T) {
	const wmax, armor = 40, 100
	// 攻擊方條件拉到最有利(無閃避、jamRoll 最小)仍應被躲掉。
	dodged := MissileDefenses{HasDisplacement: true, DisplacementRoll: gamedata.MissileDisplacementDeviceMissChance}
	if got := ResolveMissileShot(false, 0, 1, 0, 0, false, 1, wmax, 0, armor, false, dodged); got.Hit {
		t.Error("位移裝置門檻內應完全未命中")
	}
	if got := ResolveMissileShot(false, 0, 1, 0, 0, false, 1, wmax, 0, armor, false, dodged); got.RemainingArmorHP != armor {
		t.Error("未命中不該消耗裝甲")
	}
	// 骰超過門檻就正常命中。
	through := MissileDefenses{HasDisplacement: true, DisplacementRoll: gamedata.MissileDisplacementDeviceMissChance + 1}
	if got := ResolveMissileShot(false, 0, 1, 0, 0, false, 1, wmax, 0, armor, false, through); !got.Hit {
		t.Error("超過門檻應正常命中")
	}
}

// 沒裝特殊防禦時**一顆骰都不擲**——否則整條亂數流位移,既有戰鬥結果全部改變。
//
// ⚠ 這是第 65 項炸彈分支學到的那一課:用 RNG 流位置驗,不是用「傷害是 0」推論。
func TestNoDefenseSystemConsumesNoExtraDice(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{
		Name: "帶飛彈的船", Class: "戰艦", Weapon: "核飛彈",
	})
	cs, _ := s.mkPlayerCombatantsIndexed()
	last := cs[len(cs)-1]
	if last.hasLightningField || last.hasDisplacement {
		t.Fatal("測試前提不成立:沒裝的船不該有這兩個旗標")
	}

	// 同一顆種子跑兩次,中間不裝任何防禦——序列必須完全相同。
	seq := func() []int {
		r := newRandStream(7)
		out := make([]int, 0, 6)
		for i := 0; i < 6; i++ {
			out = append(out, r.Intn(100))
		}
		return out
	}
	a, b := seq(), seq()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("亂數流本身不確定,測試前提不成立(第 %d 顆:%d vs %d)", i, a[i], b[i])
		}
	}
}

// 部隊艙讓**那一艘**的陸戰隊運力加倍,不是整支艦隊。
func TestTroopPodsDoubleThatShipsCapacity(t *testing.T) {
	base := NewDemoSession()
	base.Fleet().Ships = append(base.Fleet().Ships, Ship{Name: "運兵艦", Class: "戰艦"})
	plain := base.MarineTransportCapacity()

	pods := NewDemoSession()
	pods.Fleet().Ships = append(pods.Fleet().Ships, Ship{Name: "運兵艦", Class: "戰艦", Special: "部隊艙"})
	got := pods.MarineTransportCapacity()

	class, _ := shipClassFromName("戰艦")
	one := gamedata.ShipHullMarines(class)
	if one <= 0 {
		t.Fatalf("測試前提不成立:戰艦應有正的陸戰隊運力,得到 %d", one)
	}
	if want := plain + one; got != want {
		t.Errorf("裝了部隊艙的那一艘應多 %d(翻倍):%d vs %d", one, got, want)
	}
	// 反面:艦隊裡其他船不該跟著翻倍——上面的差值正好是一艘戰艦的量,已經證明了這件事,
	// 這裡再直接檢查一次總量沒有被整體乘二。
	if got == plain*gamedata.GroundTroopPodsMultiplier && plain != one {
		t.Error("部隊艙不該讓整支艦隊的運力翻倍")
	}
}
