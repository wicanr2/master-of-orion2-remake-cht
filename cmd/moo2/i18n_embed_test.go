package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEmbeddedI18NMatchesAssets 烘進執行檔的譯表要與 `assets/i18n/` 逐位元相同。
//
// `go:embed` 讀不到套件目錄之外的路徑,所以 `cmd/moo2/embedded/i18n/` 是一份副本。
// **兩份會漂移**,而漂移之後只有打包出去的執行檔會錯——開發時跑 `-i18n assets/i18n`
// 或從 repo 根目錄跑都看不出來。這條測試就是為了讓它在 CI 就紅。
func TestEmbeddedI18NMatchesAssets(t *testing.T) {
	const srcDir = "../../assets/i18n"
	src, err := os.ReadDir(srcDir)
	if err != nil {
		t.Skipf("讀不到 %s(不在 repo 內跑?):%v", srcDir, err)
	}
	fsys, dir := i18nFS("")
	n := 0
	for _, e := range src {
		if e.IsDir() || filepath.Ext(e.Name()) != ".tsv" {
			continue
		}
		want, err := os.ReadFile(filepath.Join(srcDir, e.Name()))
		if err != nil {
			t.Errorf("讀 %s:%v", e.Name(), err)
			continue
		}
		got, err := readFromFS(fsys, dir+"/"+e.Name())
		if err != nil {
			t.Errorf("烘進去的那份缺 %s:%v", e.Name(), err)
			continue
		}
		if string(got) != string(want) {
			t.Errorf("%s 與 assets/i18n 不同步——重新複製一份到 cmd/moo2/embedded/i18n/", e.Name())
		}
		n++
	}
	if n == 0 {
		t.Fatal("一個 .tsv 都沒比到,這條測試失去意義")
	}
}
