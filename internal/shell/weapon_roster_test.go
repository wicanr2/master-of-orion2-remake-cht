package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 每一項武器都要有:手冊傷害、手冊佔格、執行檔查得到的研究主題。
//
// 這條是第 64 項(武器傷害真表)補完武器表之後的完整性閘門——**新增一項卻漏了其中一張表**
// 是最容易發生的事,而漏掉佔格會讓設計驗證靜默放行(查不到回 0 = 不佔空間)。
func TestEveryWeaponHasManualDamageSpaceAndTopic(t *testing.T) {
	for _, c := range WeaponOptions {
		if c.Name == "無武裝" {
			continue
		}
		if _, ok := gamedata.WeaponDamageByName(c.Name); !ok {
			t.Errorf("%s 沒有手冊傷害值", c.Name)
		}
		if sp := gamedata.WeaponSpaceByName[c.Name]; sp <= 0 {
			t.Errorf("%s 沒有佔格值(查不到會回 0 = 不佔空間,設計驗證會靜默放行)", c.Name)
		}
		// 里程碑元件(UnlockTech=0)走主題層級解鎖,不必對得到執行檔的科技。
		if c.UnlockTech == 0 {
			continue
		}
		topic, ok := gamedata.OrigTechTopic(c.UnlockTech)
		if !ok {
			t.Errorf("%s 的科技在執行檔的主題表上查不到", c.Name)
			continue
		}
		if topic != c.Tech {
			t.Errorf("%s 的研究主題應為執行檔給的 %v,元件表寫的是 %v", c.Name, topic, c.Tech)
		}
	}
}

// 元件表的最大傷害 == 手冊值(涵蓋第 64 項(武器傷害真表)新增的八項)。
func TestAllWeaponValuesMatchTheManualIncludingNewOnes(t *testing.T) {
	for _, c := range WeaponOptions {
		want, ok := gamedata.WeaponDamageByName(c.Name)
		if !ok {
			continue
		}
		if c.Value != want.Max {
			t.Errorf("%s 的最大傷害應為 %d,得到 %d", c.Name, want.Max, c.Value)
		}
	}
}

// 飛彈/魚雷的分類由**執行檔的 category 表**背書,不是靠名字裡有沒有「飛彈」。
//
// 手冊 p.125 的 MISSILE 表與執行檔 category 21 兩個獨立來源要一致。
func TestMissileClassificationAgreesWithTheExecutable(t *testing.T) {
	for _, c := range WeaponOptions {
		if c.UnlockTech == 0 {
			continue
		}
		cat := gamedata.TechItemCategory[c.UnlockTech]
		kind := weaponKindByName(c.Name)
		isMissileCat := cat == 21
		if isMissileCat != (kind == WeaponKindMissile) {
			t.Errorf("%s:執行檔 category=%d(飛彈=21),remake 分類 kind=%v,兩者不一致",
				c.Name, cat, kind)
		}
	}
}

// 新武器真的進得了武器清單(不是只加在 gamedata 裡沒人看得到)。
func TestNewWeaponsAppearInTheRoster(t *testing.T) {
	want := []string{"離子脈衝砲", "引力波束", "干擾者", "重錘裝置", "粒子束", "脈衝飛彈", "氙素飛彈", "質子魚雷"}
	have := map[string]bool{}
	for _, c := range WeaponOptions {
		have[c.Name] = true
	}
	for _, n := range want {
		if !have[n] {
			t.Errorf("武器清單裡缺 %s", n)
		}
	}
	// 版本清單也要跟著長(BuildWeaponOptions 是逐項複製的,漏改會只在某一版缺)。
	for _, prof := range []gamedata.RuleProfile{gamedata.Profile13(), gamedata.Profile15()} {
		if got, base := len(BuildWeaponOptions(prof)), len(WeaponOptions); got != base {
			t.Errorf("版本武器清單長度應與套件級一致:%d vs %d", got, base)
		}
	}
}
