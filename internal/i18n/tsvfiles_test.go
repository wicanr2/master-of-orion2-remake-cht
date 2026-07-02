package i18n

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommittedTSVsLoad 載入 repo 內所有 assets/i18n/*.tsv,確保格式正確可載入,
// 並檢查含 printf 佔位符的模板中英數量一致(TranslateFormat 用,不一致會 panic)。
func TestCommittedTSVsLoad(t *testing.T) {
	dir := filepath.Join("..", "..", "assets", "i18n")
	files, err := filepath.Glob(filepath.Join(dir, "*.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Skip("無 TSV 檔")
	}
	for _, fp := range files {
		data, err := os.ReadFile(fp)
		if err != nil {
			t.Fatal(err)
		}
		c := New(Traditional)
		n, err := c.LoadTSV(strings.NewReader(string(data)))
		if err != nil {
			t.Errorf("%s 載入失敗: %v", filepath.Base(fp), err)
			continue
		}
		t.Logf("%s: %d 條", filepath.Base(fp), n)

		// 佔位符一致性檢查。
		for _, line := range strings.Split(string(data), "\n") {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			cols := strings.Split(line, "\t")
			if len(cols) < 2 || strings.TrimSpace(cols[1]) == "" {
				continue
			}
			en, zh := cols[0], cols[1]
			if strings.Count(en, "%") != strings.Count(zh, "%") {
				t.Errorf("%s:模板佔位符數不一致 %q → %q", filepath.Base(fp), en, zh)
			}
		}
	}
}
