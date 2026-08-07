package main

import "testing"

// 軍官畫面的座標在第 109 項從 openorion2 換成**原版執行檔的立即數**。
//
// 來源優先序是這個專案的硬規則:**反組譯立即數 > 手冊 > openorion2 > LBX 尺寸 > 量圖**。
// 哪天有人「照 openorion2 改回去」,這裡會擋下來。
func TestOfficerScreenCoordsComeFromTheExecutable(t *testing.T) {
	want := map[string][2]int{
		"HIRE":    {313, 441},
		"POOL":    {388, 441},
		"DISMISS": {463, 441},
		"RETURN":  {538, 441},
		// 兩個分頁鈕先前是 20/166(量圖),執行檔說 9/156。
		"Colony Leaders": {9, 11},
		"Ship Officers":  {156, 11},
	}
	got := map[string][2]int{}
	for _, o := range officerOverlays() {
		got[o.enKey] = [2]int{o.x, o.y}
	}
	for name, w := range want {
		g, ok := got[name]
		if !ok {
			t.Errorf("找不到 %s 疊字", name)
			continue
		}
		if g != w {
			t.Errorf("%s 座標應是 %v(執行檔立即數),得到 %v", name, w, g)
		}
	}
}

// 四鈕 x 間距恆為 75——那是 asm 裡最顯眼的規律,抄錯一個就破了。
func TestOfficerButtonsAreEvenlySpaced(t *testing.T) {
	order := []string{"HIRE", "POOL", "DISMISS", "RETURN"}
	xs := map[string]int{}
	for _, o := range officerOverlays() {
		xs[o.enKey] = o.x
	}
	for i := 1; i < len(order); i++ {
		if d := xs[order[i]] - xs[order[i-1]]; d != officerButtonSpacing {
			t.Errorf("%s→%s 的 x 間距應是 %d,得到 %d",
				order[i-1], order[i], officerButtonSpacing, d)
		}
	}
}

// 上下捲鈕先前**完全不存在**——清單超過四列就看不到後面的人。
func TestOfficerScreenHasScrollButtons(t *testing.T) {
	want := map[string][2]int{
		"scrollUp":   {613, 22},
		"scrollDown": {613, 170},
	}
	for _, h := range officerHitRegions() {
		if w, ok := want[h.action]; ok {
			if h.x != w[0] || h.y != w[1] {
				t.Errorf("%s 座標應是 %v,得到 (%d,%d)", h.action, w, h.x, h.y)
			}
			delete(want, h.action)
		}
	}
	for name := range want {
		t.Errorf("找不到 %s 熱區", name)
	}
}

// 清單四列:起點 88、列距 109(執行檔那個迴圈的熱區範圍 34–142 / 143–251 / …)。
func TestOfficerRowCentersMatchTheExecutable(t *testing.T) {
	got := officerRowCenters()
	want := []float64{88, 197, 306, 415}
	if len(got) != len(want) {
		t.Fatalf("應有 4 列,得到 %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 列中心應是 %v,得到 %v", i, want[i], got[i])
		}
	}
	// 列距一致——執行檔那四列高度精確相同,是解碼正確的內部佐證。
	for i := 1; i < len(got); i++ {
		if d := got[i] - got[i-1]; d != 109 {
			t.Errorf("第 %d 列與前一列的間距應是 109,得到 %v", i, d)
		}
	}
}

// 熱區與疊字要對齊:RETURN 先前熱區在 538(對的)、疊字標籤在 540(錯的),自相矛盾。
func TestOfficerHitsAndOverlaysAgree(t *testing.T) {
	overlays := map[string][2]int{}
	for _, o := range officerOverlays() {
		overlays[o.enKey] = [2]int{o.x, o.y}
	}
	for _, pair := range [][2]string{{"Return", "RETURN"}, {"hire", "HIRE"}} {
		var h hitRegion
		for _, r := range officerHitRegions() {
			if r.action == pair[0] {
				h = r
			}
		}
		o := overlays[pair[1]]
		if h.x != o[0] || h.y != o[1] {
			t.Errorf("%s 的熱區 (%d,%d) 與疊字 %v 不一致", pair[0], h.x, h.y, o)
		}
	}
}
