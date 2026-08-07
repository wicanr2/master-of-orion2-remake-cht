package main

import "testing"

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
