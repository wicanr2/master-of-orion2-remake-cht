package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoHardcodedI18NPath 擋下 `assets/i18n/...` 這種**相對於當前工作目錄**的譯表路徑。
//
// ⚠ 這條測試存在的理由是一次真的踩到:第 83 項(打包路徑)把五處 `reg.LoadFS(os.DirFS("assets/i18n"))`
// 換成烘進執行檔的版本之後,**從別的目錄跑仍然失敗**——還有七處直接 `os.Open` 單一 .json,
// 而那七處不走 LoadFS,grep `LoadFS` 找不到它們。
//
// 從 repo 根目錄跑永遠看不出這個問題,所以靠人眼是攔不住的。
func TestNoHardcodedI18NPath(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("讀不到套件目錄:%v", err)
	}
	var bad []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".go" {
			continue
		}
		if e.Name() == "i18nassets.go" || strings.HasSuffix(e.Name(), "_test.go") {
			continue // 前者是唯一該提到這個路徑的地方(註解),後者是本檔自己
		}
		b, err := os.ReadFile(e.Name())
		if err != nil {
			t.Errorf("讀 %s:%v", e.Name(), err)
			continue
		}
		src := string(b)
		if strings.Contains(src, "assets/i18n") {
			bad = append(bad, e.Name()+"(寫死 assets/i18n 路徑)")
		}
		// ⚠ 第二道:光看路徑字面攔不住 `os.Open(tsvPath)` —— 那正是這一輪漏掉的第七處
		// (`interactive.go` 的第二個譯表載入點,字面已經改成 "menu.json" 卻還走 os.Open,
		// 所以第一道檢查放它過去,而從別的目錄跑仍然掛)。
		for _, l := range strings.Split(src, "\n") {
			if strings.Contains(l, "os.Open(") && strings.Contains(strings.ToLower(l), "tsv") {
				bad = append(bad, e.Name()+"(os.Open 讀 tsv)")
			}
		}
	}
	if len(bad) > 0 {
		t.Errorf("這些檔案寫死了 assets/i18n 路徑,改用 OpenI18NJSON / i18nFS:%v", bad)
	}
}
