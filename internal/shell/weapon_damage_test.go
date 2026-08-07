package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 元件表的最大傷害必須等於手冊值——不能再有「單調估計」偷偷長回來。
func TestWeaponValuesMatchTheManual(t *testing.T) {
	checked := 0
	for _, c := range WeaponOptions {
		want, ok := gamedata.WeaponDamageByName(c.Name)
		if !ok {
			continue // 手冊表上沒有的(無武裝)——不在這條的範圍
		}
		checked++
		if c.Value != want.Max {
			t.Errorf("%s 的最大傷害應為手冊值 %d,得到 %d", c.Name, want.Max, c.Value)
		}
	}
	if checked < 8 {
		t.Fatalf("測試前提不成立:應該核到至少 8 項武器,只核到 %d 項", checked)
	}
}

// 武器線**不是單調遞增**——這正是舊估計法必然出錯的地方。
//
// 核融合光束在舊表裡比中子爆破槍強(16 > 12),手冊上它比中子爆破槍弱(6 vs 12)。
// 這條把那個反轉釘住:如果有人「順手」把表改回單調,這裡會紅。
func TestWeaponLineIsNotMonotonic(t *testing.T) {
	valueOf := func(name string) int {
		for _, c := range WeaponOptions {
			if c.Name == name {
				return c.Value
			}
		}
		t.Fatalf("找不到武器 %s", name)
		return 0
	}
	neutron, fusion := valueOf("中子爆破槍"), valueOf("核融合光束")
	if fusion >= neutron {
		t.Errorf("核融合光束(%d)在手冊上比中子爆破槍(%d)弱,不該反過來", fusion, neutron)
	}
	// 質量投射器同樣比雷射之後的估計低。
	if valueOf("質量投射器") != 6 {
		t.Errorf("質量投射器應為手冊的固定 6,得到 %d", valueOf("質量投射器"))
	}
}

// 固定傷害武器的 Min == Max(手冊沒有給範圍的那幾項)。
func TestFixedDamageWeaponsHaveNoRange(t *testing.T) {
	for _, name := range []string{"質量投射器", "高斯砲", "核飛彈", "麥克萊特飛彈"} {
		d, ok := gamedata.WeaponDamageByName(name)
		if !ok {
			t.Fatalf("%s 應在手冊表上", name)
		}
		if d.Min != d.Max {
			t.Errorf("%s 手冊給的是固定值,Min 應等於 Max,得到 %d-%d", name, d.Min, d.Max)
		}
	}
	// 反面:有範圍的那幾項 Min < Max。
	for _, name := range []string{"雷射", "中子爆破槍", "核融合光束", "相位砲", "死光"} {
		d, _ := gamedata.WeaponDamageByName(name)
		if d.Min >= d.Max {
			t.Errorf("%s 手冊給的是範圍,Min 應小於 Max,得到 %d-%d", name, d.Min, d.Max)
		}
	}
}

// 版本相依那一項仍然由 RuleProfile 覆寫(1.31 電漿砲 30 / 1.50 為 20)。
func TestPlasmaCannonStaysVersionDependent(t *testing.T) {
	v13 := BuildWeaponOptions(gamedata.Profile13())
	v15 := BuildWeaponOptions(gamedata.Profile15())
	find := func(opts []Component, name string) int {
		for _, c := range opts {
			if c.Name == name {
				return c.Value
			}
		}
		t.Fatalf("找不到 %s", name)
		return 0
	}
	if find(v13, "電漿砲") == find(v15, "電漿砲") {
		t.Error("電漿砲的傷害應隨版本不同(1.31 為 30、1.50 為 20)")
	}
	// 其餘武器兩版相同——版本覆寫只該動電漿砲那一項。
	for _, c := range v15 {
		if c.Name == "電漿砲" {
			continue
		}
		if find(v13, c.Name) != c.Value {
			t.Errorf("%s 不該隨版本改變:%d vs %d", c.Name, find(v13, c.Name), c.Value)
		}
	}
}
