package main

import "testing"

// netinfo_test.go:狀態面板的版面護欄。同 netnextturn / choosenetplyrs,釘的是**算式**。

// 七個狀態的資產編號就是原版 `Reload_*_Info_` 傳給 `Reload_Generic_Net_Info_` 的立即數。
// 抄錯一個就會畫出別的狀態的圖,而畫面上還是「看起來有東西」——所以要釘住。
func TestNetInfoStateAssetIDsMatchTheOriginalCallers(t *testing.T) {
	want := map[netInfoState]int{
		netInfoWaitingForJoiners: 0x0F, // Reload_Waiting_For_Joiners_Screen_ @ 0xF552A
		netInfoJoining:           0x17, // Reload_Join_Net_Screen_            @ 0xF54CF
		netInfoWaitRaceInfo:      0x18, // Reload_Wait_For_Race_Info_         @ 0xF551B
		netInfoInitializing:      0x19, // Reload_Initializing_Net_Info_      @ 0xF54BE
		netInfoSendingData:       0x1A, // Reload_Sending_Data_Info_          @ 0xF54D9
		netInfoGeneratingMap:     0x1E, // Reload_Generating_Map_Info_        @ 0xF53CB
		netInfoGettingData:       0x1F, // Reload_Getting_Data_Info_          @ 0xF54A0
	}
	if len(netInfoStates()) != len(want) {
		t.Fatalf("狀態數 %d 與對照表 %d 不符", len(netInfoStates()), len(want))
	}
	for _, st := range netInfoStates() {
		w, ok := want[st]
		if !ok {
			t.Errorf("狀態 %d 不在對照表裡", int(st))
			continue
		}
		if int(st) != w {
			t.Errorf("狀態的資產編號應為 0x%X,實得 0x%X", w, int(st))
		}
	}
}

// x = (0x280 − w)/2、y = (0x1E0 − h)/2,而且七個狀態的面板都要留在畫面內。
func TestNetInfoWindowCentresEveryPanelOnScreen(t *testing.T) {
	// lbxinfo 量到的尺寸(multigm.lbx)。
	size := map[netInfoState][2]int{
		netInfoWaitingForJoiners: {479, 150},
		netInfoJoining:           {478, 70},
		netInfoWaitRaceInfo:      {480, 116},
		netInfoInitializing:      {478, 70},
		netInfoSendingData:       {411, 105},
		netInfoGeneratingMap:     {478, 70},
		netInfoGettingData:       {443, 105},
	}
	for _, st := range netInfoStates() {
		wh := size[st]
		w, h := wh[0], wh[1]
		x, y := netInfoWindow(w, h)
		if x != (moo2ScreenW-w)/2 {
			t.Errorf("資產 %d:x 應為 (640−%d)/2 = %d,實得 %d", int(st), w, (moo2ScreenW-w)/2, x)
		}
		if y != (moo2ScreenH-h)/2 {
			t.Errorf("資產 %d:y 應為 (480−%d)/2 = %d,實得 %d", int(st), h, (moo2ScreenH-h)/2, y)
		}
		if x < 0 || y < 0 || x+w > moo2ScreenW || y+h > moo2ScreenH {
			t.Errorf("資產 %d 的面板 (%d,%d,%d,%d) 超出 640×480", int(st), x, y, w, h)
		}
	}
}

// 傳送與接收共用同一段繪製,只差那兩個像素的位移——原版就是這樣,所以要**不相等**。
func TestNetInfoProgressPositionDiffersBetweenSendAndReceive(t *testing.T) {
	winX, winY := netInfoWindow(411, 105)
	sx, sy := netInfoProgressPos(winX, winY, false)
	if sx != winX+0x72 || sy != winY+0x42 {
		t.Errorf("傳送的進度位置應為 (+114,+66),實得 (%d,%d)", sx-winX, sy-winY)
	}
	rwinX, rwinY := netInfoWindow(443, 105)
	rx, ry := netInfoProgressPos(rwinX, rwinY, true)
	if rx != rwinX+0x79 || ry != rwinY+0x41 {
		t.Errorf("接收的進度位置應為 (+121,+65),實得 (%d,%d)", rx-rwinX, ry-rwinY)
	}
	if sx-winX == rx-rwinX && sy-winY == ry-rwinY {
		t.Error("傳送與接收的位移相同——原版是不同的兩組立即數,這裡抄漏了一組")
	}
}

// START NET GAME 鈕的位置是 Add_Waiting_For_Joiners_Field_ 的立即數,而且要落在面板內。
//
// 這條同時釘住那個訂正:0x9E/0x6A 是**按鈕**的左上角(該函式呼叫的是 Add_Button_Field_),
// 不是人數欄位。按鈕若跑出面板,代表又把它當成別的東西了。
func TestNetInfoStartButtonSitsInsideTheWaitingPanel(t *testing.T) {
	const w, h = 479, 150 // 資產 15
	winX, winY := netInfoWindow(w, h)
	bx, by := winX+netInfoStartBtnX, winY+netInfoStartBtnY
	if netInfoStartBtnX != 0x9E || netInfoStartBtnY != 0x6A {
		t.Errorf("按鈕位移應為 (0x9E,0x6A),實得 (0x%X,0x%X)", netInfoStartBtnX, netInfoStartBtnY)
	}
	if bx < winX || by < winY {
		t.Errorf("按鈕 (%d,%d) 在面板左上角之外", bx, by)
	}
	if bx+netInfoStartBtnW > winX+w || by+netInfoStartBtnH > winY+h {
		t.Errorf("按鈕 (%d,%d,%d,%d) 超出面板 %d×%d", bx-winX, by-winY,
			netInfoStartBtnW, netInfoStartBtnH, w, h)
	}
}

// 只有傳送/接收有進度數字;只有接收的 [win+0x10F] 是 1。
func TestNetInfoFlagsMatchTheOriginalCallers(t *testing.T) {
	for _, st := range netInfoStates() {
		wantProg := st == netInfoSendingData || st == netInfoGettingData
		if netInfoHasProgress(st) != wantProg {
			t.Errorf("狀態 %d 的進度旗標應為 %v", int(st), wantProg)
		}
		wantRecv := st == netInfoGettingData
		if netInfoIsReceiving(st) != wantRecv {
			t.Errorf("狀態 %d 的接收旗標應為 %v", int(st), wantRecv)
		}
	}
}

// 每個狀態都要有中文說明——原版的字烘在圖上,少一個就會是一張沒字的面板。
func TestEveryNetInfoStateHasACaption(t *testing.T) {
	b := &sceneBuilder{}
	for _, st := range netInfoStates() {
		if b.netInfoCaption(st) == "" {
			t.Errorf("狀態 %d 沒有說明文字", int(st))
		}
		if b.netInfoTitle(st) == "" {
			t.Errorf("狀態 %d 沒有標題文字", int(st))
		}
	}
}

func TestNetInfoStatusLabelMatchesPanelAssets(t *testing.T) {
	for _, st := range netInfoStates() {
		want := st == netInfoWaitingForJoiners || st == netInfoWaitRaceInfo ||
			st == netInfoSendingData || st == netInfoGettingData
		if got := netInfoHasStatusLabel(st); got != want {
			t.Errorf("狀態 %d 的 STATUS 欄 = %v, want %v", int(st), got, want)
		}
	}
}
