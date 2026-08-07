package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// starbg_test.go:星圖底的護欄。
//
// ⚠ **底色那條規則測不到,只能靠截圖驗收**:ebiten 在 game loop 之外不准 `Image.At()`
// (`ui: ReadPixels cannot be called before the game starts`),所以「星圖框內是不是純黑底」
// 沒辦法寫成單元測試。規則本身寫在 `starBGFill` 的註解裡,驗收看截圖廊第 04 張。
//
// 為什麼那條規則重要:三層星空**全都是稀疏的點疊在透明上**(層 0 最多的索引 2/3/5 對到
// (16,16,24) 這種極暗藍灰),底下沒有黑底的話,透明處會露出 `buffer0.lbx#0` 的框架美術
// ——實際踩過一次,整片星圖變成白底黑點。

// TestStarmapBackgroundSurvivesMissingAssets:沒有資產解析器時不能 panic。
//
// 這條擋的是實際踩過的坑:`decodeAsset` 對 nil `*assets.Resolver` 會 nil 解參考,
// 而畫面層的降級路徑本來就該容許「資料夾不完整」。
func TestStarmapBackgroundSurvivesMissingAssets(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("沒有資產時不該 panic,實得 %v", r)
		}
	}()
	b := &sceneBuilder{}
	b.drawStarmapBackground(ebiten.NewImage(moo2ScreenW, moo2ScreenH))

	if im := b.starBGImage(0); im != nil {
		t.Error("沒有資產解析器時 starBGImage 應回 nil")
	}
	if b.drawShipIconAt(ebiten.NewImage(64, 64), 0, 32, 32) {
		t.Error("沒有資產解析器時 drawShipIconAt 應回 false,讓呼叫端退回方塊")
	}
}

// TestStarBGLayerCount:原版 `Draw_Paralax_` 硬寫三段(STARBG.LBX 資產 0/1/2),
// 改這個數字等於改原版行為,要有反組譯依據。
func TestStarBGLayerCount(t *testing.T) {
	if starBGLayers != 3 {
		t.Errorf("視差層數 = %d,原版是 3(Draw_Paralax_ @ 0x8500F 三段)", starBGLayers)
	}
	if starBGLBX != "starbg.lbx" {
		t.Errorf("來源 = %q,原版是 STARBG.LBX", starBGLBX)
	}
}
