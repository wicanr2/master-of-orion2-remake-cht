package main

import "testing"

// netnextturn_test.go:等待畫面的版面護欄。
//
// 這張畫面的座標不是立即數,是 `Load_Net_Next_Turn_Screen_` 依**資產尺寸**現算的
// (見 netnextturn.go 檔頭)。所以測的不是「常數等於某個數」,而是**算式仍然成立**——
// 哪天資產換了或算錯了,這裡會紅。

// 資產尺寸(lbxinfo 量的)。放成常數而不是去解 LBX:測試不該依賴玩家自備的遊戲資料。
const (
	nntBannerW, nntBannerH = 630, 48
	nntMidH                = 179
	nntBottomH             = 221
)

// x = (0x280 − 資產42.寬) / 2、y = max(0, (0x1E0 − 總高) / 2)。
func TestNetWaitLayoutMatchesTheOriginalFormula(t *testing.T) {
	wantX := (moo2ScreenW - nntBannerW) / 2
	if nntX != wantX {
		t.Errorf("x 應為 (640−%d)/2 = %d,實得 %d", nntBannerW, wantX, nntX)
	}
	total := nntBannerH + nntMidH + nntBottomH
	wantY := (moo2ScreenH - total) / 2
	if wantY < 0 {
		wantY = 0
	}
	if nntBannerY != wantY {
		t.Errorf("y 應為 max(0,(480−%d)/2) = %d,實得 %d", total, wantY, nntBannerY)
	}
	// 三塊上下相接(原版是一塊接一塊堆下去的,不是各自定位)。
	if nntMidY != nntBannerY+nntBannerH {
		t.Errorf("中段應緊接標題帶下方(%d),實得 %d", nntBannerY+nntBannerH, nntMidY)
	}
	if nntBotY != nntMidY+nntMidH {
		t.Errorf("下段應緊接中段下方(%d),實得 %d", nntMidY+nntMidH, nntBotY)
	}
}

// 輸入列 y = y + [win+0xBF] + 0xBB,其中 [win+0xBF] = 資產42.高 + 資產43.高。
func TestNetWaitInputRowMatchesTheOriginalFormula(t *testing.T) {
	winBF := nntBannerH + nntMidH // [win+0xBF]
	want := nntBannerY + winBF + 0xBB
	if nntInputY != want {
		t.Errorf("輸入列 y 應為 %d + %d + 187 = %d,實得 %d", nntBannerY, winBF, want, nntInputY)
	}
	if nntInputH != 0x11 {
		t.Errorf("輸入列高應為 0x11 = 17,實得 %d", nntInputH)
	}
	if nntRowStep != 0x19 {
		t.Errorf("玩家列間距應為 0x19 = 25,實得 %d", nntRowStep)
	}
}

// 畫面上的東西不能超出 640×480,也不能互相壓到。
func TestNetWaitElementsFitOnScreen(t *testing.T) {
	if nntBotY+nntBottomH > moo2ScreenH {
		t.Errorf("下段面板底緣 %d 超出畫面高 %d", nntBotY+nntBottomH, moo2ScreenH)
	}
	if nntInputY+nntInputH > moo2ScreenH {
		t.Errorf("輸入列底緣 %d 超出畫面高 %d", nntInputY+nntInputH, moo2ScreenH)
	}
	// 玩家列要留在中段面板裡(繪製迴圈也有同一條界線,這裡把它釘住)。
	if nntRowFirst < nntMidY || nntRowFirst >= nntBotY {
		t.Errorf("玩家列起點 %d 應落在中段面板 [%d, %d) 內", nntRowFirst, nntMidY, nntBotY)
	}
}
