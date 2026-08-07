package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// starsprite_test.go:星球 sprite 資產編號的護欄。
//
// 這條算式有三個獨立來源互相印證(見 starsprite.go 檔頭):反組譯的
// `Get_Star_Picture_Seg_`、openorion2 的 `_starimg[class][zoom+size]` + `ASSET_GALAXY_STAR_IMAGES 148`、
// 以及「光譜 6 算出來剛好等於 openorion2 另外命名的 `ASSET_GALAXY_BHOLE_IMAGES 184`」。
// 所以這裡敢把絕對數字寫死。

// TestStarSpriteAssetMatchesOriginal:資產 = 148 + 光譜×6 + (縮放 + 大小)。
func TestStarSpriteAssetMatchesOriginal(t *testing.T) {
	for spectral := 0; spectral <= 5; spectral++ {
		for size := 0; size <= 3; size++ {
			idx := starSpriteZoom + size
			if idx > 5 {
				idx = 5
			}
			want := 148 + spectral*6 + idx
			if got := starSpriteAsset(spectral, size); got != want {
				t.Errorf("光譜 %d 大小 %d:資產 = %d,want %d", spectral, size, got, want)
			}
			// 光譜 0..5 的圖全在 148..183(實測 BUFFER0 那 36 張)。
			if got := starSpriteAsset(spectral, size); got < 148 || got > 183 {
				t.Errorf("光譜 %d 大小 %d:資產 %d 超出星球那 36 張(148..183)", spectral, size, got)
			}
		}
	}
}

// TestStarSpriteBlackHoleLandsOnBholeAssets:光譜 6(黑洞)**不加大小**,
// 算出來要落在 openorion2 另外命名的 `ASSET_GALAXY_BHOLE_IMAGES 184` 那組(184..187)。
//
// 這條同時是整條公式的交叉驗證:如果基底、每組張數、或「6 = 黑洞」有任何一項抄錯,
// 這裡就對不上 184。
func TestStarSpriteBlackHoleLandsOnBholeAssets(t *testing.T) {
	const bholeBase = 184 // openorion2 galaxy.cpp:45
	want := bholeBase + starSpriteZoom
	for size := 0; size <= 3; size++ {
		if got := starSpriteAsset(6, size); got != want {
			t.Errorf("黑洞(大小 %d)資產 = %d,want %d ——黑洞不該隨大小變", size, got, want)
		}
	}
}

// TestStarSpriteAssetClampsOutOfRange:光譜/大小越界要夾回去,不能溢位到別組的圖。
func TestStarSpriteAssetClampsOutOfRange(t *testing.T) {
	if got := starSpriteAsset(-1, 0); got != starSpriteAsset(0, 0) {
		t.Errorf("光譜 -1 應夾成 0,實得資產 %d", got)
	}
	if got := starSpriteAsset(99, 0); got != starSpriteAsset(6, 0) {
		t.Errorf("光譜 99 應夾成 6,實得資產 %d", got)
	}
	if got := starSpriteAsset(0, -5); got != starSpriteAsset(0, 0) {
		t.Errorf("大小 -5 應夾成 0,實得資產 %d", got)
	}
	// 大小超出 3(原版 StarSize 只有 Large/Medium/Small/Tiny)也不能跨組。
	if got := starSpriteAsset(0, 99); got > 153 {
		t.Errorf("大小 99 的資產 %d 溢位到下一個光譜(應夾在 148..153)", got)
	}
}

// TestStarSpriteFrameForOnlyAnimatesBlackHoles 釘住「只有黑洞會動」。
//
// 這不是取捨,是查證結果:`Draw_A_Star_` 裡的閃爍碼在出貨版**是死碼**——
// 啟動它要把 `star[+0x64]` 設成 ≥ 0,而全檔對那個欄位的位元組寫入只有 reset(0xFF);
// 全域預算 `word_19C164` 更是只減不加。詳見 starsprite.go 檔頭。
//
// 所以這條測試擋的是「有人看到 5 幀資產就順手把一般星球也接上動畫」——
// 那會讓 remake 比原版**多**一個原版沒有的效果。
func TestStarSpriteFrameForOnlyAnimatesBlackHoles(t *testing.T) {
	b := &sceneBuilder{} // 無資產解析器:幀數查不到,一律回 0
	for _, spectral := range []int{0, 1, 2, 3, 4, 5} {
		for tick := 0; tick < 20; tick++ {
			b.animTick = tick
			if got := b.starSpriteFrameFor(shell.Star{Spectral: spectral, Size: 1}); got != 0 {
				t.Fatalf("光譜 %d 的星球應恆為第 0 幀,tick %d 得到 %d", spectral, tick, got)
			}
		}
	}
}

// TestBlackHoleFrameAdvancesEveryTwoRedraws 釘住原版的推進比例:
// 幀號 = 計數 / 2,也就是**每一幀停留 2 次重畫**。
//
// 這個 2 在兩個地方獨立出現(`Draw_Black_Holes_` 的 `幀數 × 2` 再除以 2、
// 一般星球 `Draw_A_Star_` 的 `sar eax, 1`),不是單點讀出來的。
func TestBlackHoleFrameAdvancesEveryTwoRedraws(t *testing.T) {
	if blackHoleHoldRedraws != 2 {
		t.Fatalf("每幀停留次數應為 2(原版真值),實得 %d", blackHoleHoldRedraws)
	}
	// 黑洞那張實測 16 幀,所以一圈是 32 次重畫。
	const frames = 16
	for _, c := range []struct{ tick, want int }{
		{0, 0}, {1, 0}, {2, 1}, {3, 1}, {30, 15}, {31, 15}, {32, 0}, {33, 0},
	} {
		if got := blackHoleFrameAt(c.tick, frames); got != c.want {
			t.Errorf("tick %d 應為第 %d 幀,實得 %d", c.tick, c.want, got)
		}
	}
}

// TestBlackHoleFrameNeverOutOfRange:任何 tick 都不能算出越界的幀號。
//
// 這條擋的是「資產只有 1 幀」或「解不出資產(幀數 0)」時的除零/越界——
// 畫面缺資產要降級,不是 panic。
func TestBlackHoleFrameNeverOutOfRange(t *testing.T) {
	for _, frames := range []int{0, 1, 2, 5, 16} {
		for tick := -3; tick < 100; tick++ {
			got := blackHoleFrameAt(tick, frames)
			if got < 0 || (frames > 0 && got >= frames) || (frames <= 1 && got != 0) {
				t.Fatalf("幀數 %d、tick %d 算出越界幀號 %d", frames, tick, got)
			}
		}
	}
}

// TestBlackHoleAssetIgnoresStarSize:黑洞那條分支**不加大小偏移**。
//
// 原版 `Get_Star_Picture_Seg_` 的 `al == 6` 分支是 `eax = zoom`(其餘光譜才是 `zoom + bl`),
// 對應 openorion2 的 `_bholeimg[i]`,i 只跑縮放 0..3 → 資產 184..187。
// 所以四種星球大小都要落在同一張。
//
// 落點是 `184 + 縮放`,而 remake 的縮放固定 3(見 starsprite.go 檔頭 ⚠:那是 remake 的選擇),
// 於是實際取的是 **BUFFER0#187**。實測該張 25×25、16 幀,和 184..187 一整組一致。
func TestBlackHoleAssetIgnoresStarSize(t *testing.T) {
	want := 184 + starSpriteZoom
	for size := 0; size < 4; size++ {
		if got := starSpriteAsset(blackHoleSpectralClass, size); got != want {
			t.Errorf("黑洞(大小 %d)應為資產 %d,實得 %d", size, want, got)
		}
	}
}
