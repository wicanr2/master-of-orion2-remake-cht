package main

import "testing"

func TestSpecialControlsStayInsideLogicalCanvasAndDoNotOverlap(t *testing.T) {
	actions := []string{"specialprev", "specialnext", "specialadd", "specialdel"}
	for i, action := range actions {
		r := designSpecialControlRect(action)
		if r[0] < 0 || r[1] < 0 || r[0]+r[2] > moo2ScreenW || r[1]+r[3] > moo2ScreenH {
			t.Fatalf("%s 超出 640×480：%v", action, r)
		}
		if i > 0 {
			prev := designSpecialControlRect(actions[i-1])
			if prev[0]+prev[2] > r[0] {
				t.Fatalf("%s 與 %s 熱區重疊：%v %v", actions[i-1], action, prev, r)
			}
		}
	}
}

func TestTacticalWeaponSlotRectsStayInLeftControlDeck(t *testing.T) {
	for i := 0; i < 8; i++ {
		r := tacticalWeaponSlotRect(i)
		if r[0] < 0 || r[1] < combatControlDeckY || r[0]+r[2] > 268 || r[1]+r[3] > moo2ScreenH {
			t.Fatalf("武器槽 %d 超出左側控制列安全框：%v", i, r)
		}
		for j := 0; j < i; j++ {
			p := tacticalWeaponSlotRect(j)
			if r[0] < p[0]+p[2] && r[0]+r[2] > p[0] && r[1] < p[1]+p[3] && r[1]+r[3] > p[1] {
				t.Fatalf("武器槽 %d/%d 重疊：%v %v", j, i, p, r)
			}
		}
	}
}
