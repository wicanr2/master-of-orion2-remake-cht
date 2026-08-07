package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 反飛彈火箭進得了元件清單,而且研究主題對得上執行檔。
func TestAntiMissileRocketIsAComponent(t *testing.T) {
	var found *Component
	for i := range SpecialOptions {
		if SpecialOptions[i].Name == antiMissileRocketName {
			found = &SpecialOptions[i]
		}
	}
	if found == nil {
		t.Fatal("特殊元件清單裡應該有反飛彈火箭")
	}
	topic, ok := gamedata.OrigTechTopic(gamedata.TECH_ANTIMISSILE_ROCKETS)
	if !ok {
		t.Fatal("AMR 的科技應查得到主題")
	}
	if found.Tech != topic {
		t.Errorf("研究主題應為執行檔給的 %v,元件表寫的是 %v", topic, found.Tech)
	}
	// 執行檔的分類與手冊 p.127 一致(category 28 = 反飛彈/干擾)。
	if cat := gamedata.TechItemCategory[gamedata.TECH_ANTIMISSILE_ROCKETS]; cat != 28 {
		t.Errorf("AMR 的執行檔分類應為 28(反飛彈/干擾),得到 %d", cat)
	}
}

// 裝了 AMR 的船,combatant 上的旗標要跟著設起來。
func TestAMRFlagReachesCombatant(t *testing.T) {
	s := NewDemoSession()
	if len(s.Fleet().Ships) == 0 {
		t.Fatal("測試前提不成立:示範對局應該有船")
	}
	for i := range s.Fleet().Ships {
		s.Fleet().Ships[i].Special = "無"
	}
	if s.mkPlayerCombatants()[0].hasAMR {
		t.Error("沒裝 AMR 的船不該有旗標")
	}
	for i := range s.Fleet().Ships {
		s.Fleet().Ships[i].Special = antiMissileRocketName
	}
	if !s.mkPlayerCombatants()[0].hasAMR {
		t.Error("裝了 AMR 的船應該有旗標")
	}
}

// AMR 真的攔得到飛彈:同一組種子下,裝 AMR 的一方被飛彈打中的次數應該較少。
//
// 用計數比較而不是單次比較——單次會被隨機性帶著走。
func TestAMRInterceptsMissiles(t *testing.T) {
	hits := func(withAMR bool) int {
		total := 0
		for seed := int64(1); seed <= 300; seed++ {
			atk := []combatant{{hp: 100, atk: 20, wmin: 10, wmax: 20, kind: WeaponKindMissile, shipIdx: -1}}
			def := []combatant{{hp: 100000, atk: 1, def: 1, armor: 0, shipIdx: -1, hasAMR: withAMR}}
			before := def[0].hp
			battleVolley(atk, &def, rand.New(rand.NewSource(seed)))
			if len(def) > 0 && def[0].hp < before {
				total++
			}
		}
		return total
	}
	bare, guarded := hits(false), hits(true)
	if bare == 0 {
		t.Fatalf("測試前提不成立:沒有 AMR 時飛彈應該打得中(%d 次)", bare)
	}
	if guarded >= bare {
		t.Errorf("裝了 AMR 命中次數應下降:無 AMR %d 次 vs 有 AMR %d 次", bare, guarded)
	}
}

// 正對照:AMR **不該**擋光束武器(手冊:它攔的是來襲飛彈)。
func TestAMRDoesNotStopBeams(t *testing.T) {
	hits := func(withAMR bool) int {
		total := 0
		for seed := int64(1); seed <= 300; seed++ {
			atk := []combatant{{hp: 100, atk: 40, wmin: 20, wmax: 40, kind: WeaponKindBeam, shipIdx: -1}}
			def := []combatant{{hp: 100000, atk: 1, def: 1, armor: 0, shipIdx: -1, hasAMR: withAMR}}
			before := def[0].hp
			battleVolley(atk, &def, rand.New(rand.NewSource(seed)))
			if len(def) > 0 && def[0].hp < before {
				total++
			}
		}
		return total
	}
	if bare, guarded := hits(false), hits(true); bare != guarded {
		t.Errorf("AMR 不該影響光束命中:無 AMR %d 次 vs 有 AMR %d 次", bare, guarded)
	}
}
