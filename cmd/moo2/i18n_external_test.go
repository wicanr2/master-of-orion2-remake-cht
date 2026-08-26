package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestI18NFSOverride 驗證外部文案目錄可直接覆寫，不依賴目前工作目錄或內嵌副本。
func TestI18NFSOverride(t *testing.T) {
	dir := t.TempDir()
	want := []byte(`[{"key":"Continue","value":"繼續"}]`)
	if err := os.WriteFile(filepath.Join(dir, "menu.json"), want, 0o644); err != nil {
		t.Fatal(err)
	}
	fsys, root := i18nFS(dir)
	got, err := readFromFS(fsys, filepath.Join(root, "menu.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("外部文案內容 = %q，預期 %q", got, want)
	}
}
