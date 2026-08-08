package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 炸彈在**艦隊戰裡完全沒有作用**(手冊 p.126:only useful against planetary targets)。
//
// 這是這一項存在的理由:先前 `weaponKindByName` 查不到就回 beam,一艘掛核彈的船會
// 當光束艦用——而核彈的傷害(3-12)比同期光束高,等於免費的強化。
func TestBombsCannotHitShips(t *testing.T) {
	def := []combatant{{hp: 1000, atk: 1, def: 1, shipIdx: -1}}
	atk := []combatant{{hp: 100, atk: 60, wmin: 30, wmax: 60, kind: WeaponKindBomb, shipIdx: -1}}
	before := def[0].hp
	for seed := int64(1); seed <= 50; seed++ {
		battleVolley(atk, &def, rand.New(rand.NewSource(seed)))
	}
	if def[0].hp != before {
		t.Errorf("炸彈不該對艦艇造成任何傷害:%d → %d", before, def[0].hp)
	}
}

// 正對照:同樣的攻防值換成光束就打得動——證明上面那個「沒傷害」不是因為打不中。
func TestSameStatsAsBeamDoesHitShips(t *testing.T) {
	def := []combatant{{hp: 100000, atk: 1, def: 1, shipIdx: -1}}
	atk := []combatant{{hp: 100, atk: 60, wmin: 30, wmax: 60, kind: WeaponKindBeam, shipIdx: -1}}
	before := def[0].hp
	for seed := int64(1); seed <= 50; seed++ {
		battleVolley(atk, &def, rand.New(rand.NewSource(seed)))
	}
	if def[0].hp >= before {
		t.Errorf("同樣攻防的光束應該打得動:%d → %d", before, def[0].hp)
	}
}

// 炸彈不擲骰——擲了會位移後面每一發的隨機序列,讓決定性測試無故變動。
func TestBombDoesNotConsumeRandomness(t *testing.T) {
	roll := func(kind WeaponKind) int {
		rng := rand.New(rand.NewSource(42))
		def := []combatant{{hp: 100000, atk: 1, def: 1, shipIdx: -1}}
		atk := []combatant{{hp: 100, atk: 10, wmin: 5, wmax: 10, kind: kind, shipIdx: -1}}
		battleVolley(atk, &def, rng)
		return rng.Intn(1 << 30) // 齊射之後亂數流的位置
	}
	fresh := func() int {
		rng := rand.New(rand.NewSource(42))
		return rng.Intn(1 << 30)
	}
	if roll(WeaponKindBomb) != fresh() {
		t.Error("炸彈那一發不該消耗亂數")
	}
	if roll(WeaponKindBeam) == fresh() {
		t.Error("測試前提不成立:光束那一發應該有消耗亂數")
	}
}

// 四種炸彈的分類、傷害、佔格、主題都齊,而且分類與執行檔一致(category 19)。
func TestBombRosterIsComplete(t *testing.T) {
	names := []string{"核彈", "融合彈", "反物質彈", "中子彈"}
	inRoster := map[string]Component{}
	for _, c := range WeaponOptions {
		inRoster[c.Name] = c
	}
	for _, n := range names {
		c, ok := inRoster[n]
		if !ok {
			t.Errorf("武器清單裡缺 %s", n)
			continue
		}
		if weaponKindByName(n) != WeaponKindBomb {
			t.Errorf("%s 應分類為炸彈", n)
		}
		if cat := gamedata.TechItemCategory[c.UnlockTech]; cat != 19 {
			t.Errorf("%s 的執行檔分類應為 19(炸彈),得到 %d", n, cat)
		}
		d, ok := gamedata.WeaponDamageByName(n)
		if !ok || d.Max <= 0 {
			t.Errorf("%s 缺手冊傷害值", n)
		}
		if c.Value != d.Max {
			t.Errorf("%s 的 Value 應為手冊最大傷害 %d,得到 %d", n, d.Max, c.Value)
		}
		if gamedata.WeaponSpaceByName[n] <= 0 {
			t.Errorf("%s 缺佔格值", n)
		}
	}
}

// 生物武器**不該**進元件表——第 53 項已用另一條路徑(科技擁有 → 轟炸擲骰)接好了,
// 加進來會讓同一條規則生效兩次。
func TestBioWeaponsStayOutOfTheComponentRoster(t *testing.T) {
	for _, c := range WeaponOptions {
		if c.Name == "死亡孢子" || c.Name == "生物滅絕者" {
			t.Errorf("%s 不該進武器元件表(第 53 項已走科技擁有那條路徑)", c.Name)
		}
	}
}
