package uifont

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	bitmapfont "github.com/hajimehoshi/bitmapfont/v4"
	"golang.org/x/image/math/fixed"
)

// TestBitmapTCCoverage 是 [HARD] 守門測試(取代 runtime fallback):
// 掃描 repo 全語料(assets/i18n/*.tsv 各欄 + internal/、cmd/ 下所有 .go 檔)出現的每一個
// CJK/全形字元,逐字檢查 bitmapfont.FaceTC 是否有實際墨點(ink)。任一字缺就 Fatal 並列出缺字清單,
// 讓「未來新增未涵蓋字」在 build/CI 就被抓到,而不是 runtime 靜默消失成方塊/空白
// (對齊 docs/tech/pixel-font-decision.md 的窮舉驗證與 mom playbook 的缺字 [HARD] 教訓)。
//
// 為何不能只查 GlyphBounds 的 ok:實測過 bitmapfont/v4 的 tcFace/lazyFace/bitmap.Face
// 實作(internal/bitmap/bitmap.go),GlyphBounds 對任何 rune<0x10000 一律回 ok=true——
// 它只依 Unicode 寬度分類(全形/半形)算一個矩形,完全不檢查底層點陣圖該格是否真的畫了字。
// 用 GlyphBounds.ok 當缺字判準會恆為「0 缺字」,是假陽性的守門。真正能分辨「缺字→空白格」
// 與「有字」的只有實際取 Glyph() 的 mask 再檢查是否有任一像素 alpha!=0(有墨點)。
// 已用私用區碼點(U+E000)驗證此法可正確判為無墨點,常見繁體字(銀河霸主/瑞/崔等)判為有墨點。
func TestBitmapTCCoverage(t *testing.T) {
	root := repoRoot(t)

	chars := map[rune]struct{}{}

	// assets/i18n/*.tsv:三欄 英文<TAB>中文<TAB>備註,逐欄掃描(備註也可能含中文範例)。
	tsvDir := filepath.Join(root, "assets", "i18n")
	tsvFiles, err := filepath.Glob(filepath.Join(tsvDir, "*.tsv"))
	if err != nil {
		t.Fatalf("glob %s: %v", tsvDir, err)
	}
	if len(tsvFiles) == 0 {
		t.Fatalf("在 %s 找不到任何 .tsv,repo 根目錄定位可能有誤", tsvDir)
	}
	for _, path := range tsvFiles {
		collectCJK(t, path, chars)
	}

	// internal/、cmd/ 下所有 .go 檔(含硬編中文字串與中文註解/docstring)。
	for _, sub := range []string{"internal", "cmd"} {
		dir := filepath.Join(root, sub)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			collectCJK(t, path, chars)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}

	if len(chars) == 0 {
		t.Fatal("語料掃描結果為 0 個 CJK 字元,掃描邏輯或路徑可能有誤")
	}

	var missing []rune
	for r := range chars {
		if !hasInk(r) {
			missing = append(missing, r)
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i] < missing[j] })

	t.Logf("語料共 %d 個唯一 CJK/全形字元,bitmapfont.FaceTC 缺字 %d 個", len(chars), len(missing))
	if len(missing) > 0 {
		var sb strings.Builder
		for _, r := range missing {
			sb.WriteRune(r)
			sb.WriteByte(' ')
		}
		t.Fatalf("bitmapfont.FaceTC 缺字 %d 個(需修字型或補 fallback): %s", len(missing), sb.String())
	}
}

// hasInk 檢查 bitmapfont.FaceTC 對 r 實際畫出的 glyph mask 是否有任一像素 alpha!=0。
// dot 位置不影響 mask 內容(bitmap.Face.Glyph 只用 dot 算目的地矩形 dr,mask 本身
// 依 rune 直接查點陣圖表格 mx/my,見 internal/bitmap/bitmap.go),故用零點即可。
func hasInk(r rune) bool {
	_, mask, _, _, ok := bitmapfont.FaceTC.Glyph(fixed.Point26_6{}, r)
	if !ok || mask == nil {
		return false
	}
	b := mask.Bounds()
	if b.Empty() {
		return false
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := mask.At(x, y).RGBA()
			if a != 0 {
				return true
			}
		}
	}
	return false
}

// collectCJK 讀取檔案內容,把每個 CJK/全形字元(isCJK)加進 set。
func collectCJK(t *testing.T, path string, set map[rune]struct{}) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("讀取 %s: %v", path, err)
	}
	for _, r := range string(data) {
		if isCJK(r) {
			set[r] = struct{}{}
		}
	}
}

// repoRoot 用 runtime.Caller 定位本檔位置,回推 repo 根目錄
// (internal/uifont/coverage_test.go 往上兩層)。
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 取得本檔路徑失敗")
	}
	// thisFile = <root>/internal/uifont/coverage_test.go
	root := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("推算的 repo 根目錄 %s 找不到 go.mod: %v", root, err)
	}
	return root
}
