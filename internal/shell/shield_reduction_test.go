package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 護盾減傷要對得上手冊,不是「清單索引 × 2」。
//
// 元件名稱本身就寫著答案(第一級/第三級/第五級/第七級/第十級),而手冊說得很直白:
// 每次攻擊減傷 = 等級數字。先前五級裡有四級偏高。
func TestShieldReductionMatchesTheManual(t *testing.T) {
	cases := []struct {
		name string
		want int
	}{
		{"無護盾", 0},
		{"第一級護盾", 1},
		{"第三級護盾", 3},
		{"第五級護盾", 5},
		{"第七級護盾", 7},
		{"第十級護盾", 10},
	}
	for _, c := range cases {
		if got := shieldReduceByName(c.name); got != c.want {
			t.Errorf("%s 的每次攻擊減傷應為 %d,得到 %d", c.name, c.want, got)
		}
	}
	// 不存在的名稱回 0,不 panic。
	if got := shieldReduceByName("不存在的護盾"); got != 0 {
		t.Errorf("未知護盾應回 0,得到 %d", got)
	}
}

// 減傷值不該取決於**清單順序**——那正是舊算法出問題的原因。
//
// 這一條的作法是直接繞過清單:同一級護盾的科技,不管它排在第幾個,答案都一樣。
func TestShieldReductionDoesNotDependOnListOrder(t *testing.T) {
	for _, c := range ShieldOptions {
		byName := shieldReduceByName(c.Name)
		byTech := gamedata.ShieldReductionForTech(c.UnlockTech)
		if byName != byTech {
			t.Errorf("%s:查名稱得 %d、查科技得 %d,兩者應一致", c.Name, byName, byTech)
		}
	}
	// 護盾清單是遞增的——順序本身仍該有意義(貴的擋得多)。
	prev := -1
	for _, c := range ShieldOptions {
		got := shieldReduceByName(c.Name)
		if got < prev {
			t.Errorf("%s 的減傷 %d 低於前一級的 %d,清單順序與強度不一致", c.Name, got, prev)
		}
		prev = got
	}
}

// 減傷真的會傳進戰鬥解算:掛了護盾的船,對方每一發都少扣一點。
func TestShieldReductionReachesCombat(t *testing.T) {
	s := NewDemoSession()
	if len(s.Fleet().Ships) == 0 {
		t.Fatal("測試前提不成立:示範對局應該有船")
	}
	for i := range s.Fleet().Ships {
		s.Fleet().Ships[i].Shield = "無護盾"
	}
	bare := s.mkPlayerCombatants()

	for i := range s.Fleet().Ships {
		s.Fleet().Ships[i].Shield = "第十級護盾"
	}
	shielded := s.mkPlayerCombatants()

	if shielded[0].shield <= bare[0].shield {
		t.Errorf("掛第十級護盾應提高減傷:%d → %d", bare[0].shield, shielded[0].shield)
	}
	if shielded[0].shield != gamedata.DamageShieldReductionClassX {
		t.Errorf("第十級護盾的減傷應為手冊值 %d,得到 %d(星雲會另外扣,demo 母星不在星雲)",
			gamedata.DamageShieldReductionClassX, shielded[0].shield)
	}
}
