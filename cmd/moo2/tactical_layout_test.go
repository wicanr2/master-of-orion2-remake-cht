package main

import "testing"

func TestTacticalOriginalSkeletonRegionsStayInsideCanvas(t *testing.T) {
	if combatControlDeckY != 351 || combatControlDeckY+129 != moo2ScreenH {
		t.Fatalf("COMBAT 控制甲板必須固定在 (0,351,640,129)：y=%d screen=%d", combatControlDeckY, moo2ScreenH)
	}
	if gcX0 < 0 || gcY0 < 0 || gcX0+gcCols*gcCW > moo2ScreenW || gcY0+gcRows*gcCH > combatControlDeckY {
		t.Fatalf("隱形戰術格位不得跨入控制甲板：(%d,%d)-(%d,%d)",
			gcX0, gcY0, gcX0+gcCols*gcCW, gcY0+gcRows*gcCH)
	}
	if tacticalShipInfoY < combatControlDeckY || tacticalSystemsY < combatControlDeckY ||
		tacticalShipInfoX+tacticalShipInfoW > moo2ScreenW || tacticalSystemsX+tacticalSystemsW > moo2ScreenW {
		t.Fatal("選艦／Systems 資訊必須完全位於原版控制甲板")
	}
}

func TestTacticalWeaponRowsRemainInOriginalDeck(t *testing.T) {
	for i := 0; i < 8; i++ {
		r := tacticalWeaponSlotRect(i)
		if r[0] < 108 || r[0]+r[2] > 266 || r[1] < combatControlDeckY || r[1]+r[3] > moo2ScreenH {
			t.Fatalf("武器列 %d 越出原版 WEAPONS／SPECIALS 區：%v", i, r)
		}
		if barButtonHit(r[0]+2, r[1]+r[3]/2) >= 0 {
			t.Fatalf("武器列 %d 與控制鈕熱區重疊：%v", i, r)
		}
	}
}

func TestTacticalFighterLaunchAdapterStaysInSpecialsDeck(t *testing.T) {
	x, y, w, h := launchRect()
	if x < 108 || x+w > 266 || y < combatControlDeckY || y+h > moo2ScreenH {
		t.Fatalf("戰機出擊 adapter 必須留在 SPECIALS 控制甲板：(%d,%d,%d,%d)", x, y, w, h)
	}
	for i := 0; i < 8; i++ {
		r := tacticalWeaponSlotRect(i)
		if x < r[0]+r[2] && x+w > r[0] && y < r[1]+r[3] && y+h > r[1] {
			t.Fatalf("戰機出擊 adapter 與武器列 %d 重疊", i)
		}
	}
}
