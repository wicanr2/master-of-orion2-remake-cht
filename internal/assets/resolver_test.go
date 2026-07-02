package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolverOverrideOrder(t *testing.T) {
	base := t.TempDir()
	patch := t.TempDir()

	// 基礎有 GAME.LBX 與 STARDB.LBX;patch 只覆蓋 GAME.LBX。
	write(t, base, "GAME.LBX", "base-game")
	write(t, base, "STARDB.LBX", "base-stardb")
	write(t, patch, "GAME.LBX", "patch-game")

	// patch 優先(在前)。
	r, err := NewResolver(patch, base)
	if err != nil {
		t.Fatal(err)
	}

	// GAME.LBX 應取 patch 版;STARDB.LBX 只有基礎有。
	if b, _ := r.Read("GAME.LBX"); string(b) != "patch-game" {
		t.Errorf("GAME.LBX = %q,預期 patch-game(patch 覆蓋)", b)
	}
	if b, _ := r.Read("STARDB.LBX"); string(b) != "base-stardb" {
		t.Errorf("STARDB.LBX = %q,預期 base-stardb", b)
	}
}

func TestResolverCaseInsensitive(t *testing.T) {
	base := t.TempDir()
	write(t, base, "Game.Lbx", "content")
	r, err := NewResolver(base)
	if err != nil {
		t.Fatal(err)
	}
	// 以任意大小寫查詢都應命中。
	for _, q := range []string{"GAME.LBX", "game.lbx", "Game.Lbx"} {
		if _, ok := r.Path(q); !ok {
			t.Errorf("查詢 %q 未命中(應大小寫不敏感)", q)
		}
	}
}

func TestResolverMissing(t *testing.T) {
	base := t.TempDir()
	r, err := NewResolver(base)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Path("NOPE.LBX"); ok {
		t.Error("不存在的檔應回傳 found=false")
	}
	if _, err := r.Read("NOPE.LBX"); err == nil {
		t.Error("Read 不存在的檔應回傳 error")
	}
}
