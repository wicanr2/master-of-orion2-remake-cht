package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommittedJSONsLoad 載入 repo 內所有 assets/i18n/*.json,確保格式正確可載入,
// 並檢查含 printf 佔位符的模板中英數量一致(TranslateFormat 用,不一致會 panic)。
func TestCommittedJSONsLoad(t *testing.T) {
	dir := filepath.Join("..", "..", "assets", "i18n")
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Skip("無 JSON 檔")
	}
	for _, fp := range files {
		data, err := os.ReadFile(fp)
		if err != nil {
			t.Fatal(err)
		}
		c := New(Traditional)
		n, err := c.LoadJSON(strings.NewReader(string(data)))
		if err != nil {
			t.Errorf("%s 載入失敗: %v", filepath.Base(fp), err)
			continue
		}
		t.Logf("%s: %d 條", filepath.Base(fp), n)

		var entries []Entry
		if err := json.Unmarshal(data, &entries); err != nil {
			t.Errorf("%s JSON 解碼失敗: %v", filepath.Base(fp), err)
			continue
		}
		for _, entry := range entries {
			en, zh := entry.Key, entry.Value
			if strings.Count(en, "%") != strings.Count(zh, "%") {
				t.Errorf("%s:模板佔位符數不一致 %q → %q", filepath.Base(fp), en, zh)
			}
		}
	}
}
