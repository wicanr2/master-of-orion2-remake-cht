package uifont

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseRealFont 是 mom [HARD] 檢查:確認候選 CJK 字型能被 Go opentype 解析,
// 並能量測中文字寬(代表 glyph 可用)。以 MOO2_FONT_TEST 指定字型檔;.ttc 走集合解析。
func TestParseRealFont(t *testing.T) {
	path := os.Getenv("MOO2_FONT_TEST")
	if path == "" {
		t.Skip("未設 MOO2_FONT_TEST")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var f *Font
	if strings.EqualFold(filepath.Ext(path), ".ttc") {
		f, err = LoadCollection(data, 0)
	} else {
		f, err = Load(data)
	}
	if err != nil {
		t.Fatalf("Go opentype 解析失敗(字型不相容,需換檔): %v", err)
	}

	// 量測中文字:寬度須 > 0,代表 glyph 存在且可 rasterize。
	for _, s := range []string{"銀河霸主", "繼續", "新遊戲"} {
		w, h := f.Measure(s, 16)
		if w <= 0 || h <= 0 {
			t.Errorf("量測 %q = %.1fx%.1f,應 > 0(glyph 可能缺失)", s, w, h)
		} else {
			t.Logf("%q @16px = %.1fx%.1f", s, w, h)
		}
	}
}
