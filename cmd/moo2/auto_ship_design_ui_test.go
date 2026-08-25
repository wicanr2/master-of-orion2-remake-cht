package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestShipDesignAutoInitialLoadoutIsStable(t *testing.T) {
	b := &sceneBuilder{session: shell.NewDemoSession()}
	if !b.loadShipDesign(2) {
		t.Fatal("應能從六艦體設計庫載入巡洋艦")
	}
	// 模擬玩家手動換武器並寫回；畫面重建不得重新套 Auto Design。
	b.designWeapon = 0
	if !b.saveShipDesign() {
		t.Fatal("應能寫回巡洋艦設計")
	}
	if !b.loadShipDesign(2) || !b.designLoaded || b.designWeapon != 0 {
		t.Fatal("持久設計不得在畫面重建時被 Auto Design 覆寫")
	}
}

func TestShipDesignUIEditsSelectedMountOnly(t *testing.T) {
	b := &sceneBuilder{session: shell.NewDemoSession()}
	if !b.loadShipDesign(2) {
		t.Fatal("應能載入巡洋艦設計")
	}
	firstBefore, _ := b.session.ShipDesign(2)
	idx, ok := b.session.AddShipDesignMount(2, 0)
	if !ok {
		t.Fatal("應能新增第二槽")
	}
	b.designMount = idx
	if !b.loadShipDesign(2) {
		t.Fatal("應能載入第二槽")
	}
	b.designWeapon = 0
	b.designMods = nil
	if !b.saveShipDesign() {
		t.Fatal("應能寫回目前槽")
	}
	got, _ := b.session.ShipDesign(2)
	if len(got.WeaponMounts) != 2 || got.WeaponMounts[1].Name != shell.WeaponOptions[0].Name {
		t.Fatalf("第二槽未寫回：%+v", got.WeaponMounts)
	}
	if got.WeaponMounts[0].Name != firstBefore.WeaponMounts[0].Name {
		t.Fatal("編輯第二槽不得污染第一槽")
	}
}

func TestShipDesignMountGeometryUsesDisjointSafeRects(t *testing.T) {
	for i := 0; i < 8; i++ {
		r := designMountSlotRect(i)
		if r[0] < 305 || r[1] < 180 || r[0]+r[2] > 545 || r[1]+r[3] > 198 {
			t.Fatalf("槽 %d 超出指定安全帶：%v", i, r)
		}
		if i > 0 {
			prev := designMountSlotRect(i - 1)
			if prev[0]+prev[2] > r[0] {
				t.Fatalf("槽 %d 與前槽重疊：%v / %v", i, prev, r)
			}
		}
	}
	for _, action := range []string{"mountadd", "mountdel", "mountdec", "mountinc"} {
		r := designMountControlRect(action)
		if r[0] < 548 || r[0]+r[2] > 632 || r[1] < 180 || r[1]+r[3] > 219 {
			t.Fatalf("控制 %s 超出右側 text-safe 區：%v", action, r)
		}
	}
}
