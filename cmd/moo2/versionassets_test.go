package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestBuildVersionAssetDirsSharedFallback(t *testing.T) {
	base := t.TempDir()
	dirs, err := buildVersionAssetDirs(base, "", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []gamedata.GameVersion{gamedata.VersionClassic13, gamedata.VersionCommunity15} {
		got, ok := dirs.forVersion(v)
		if !ok || len(got) != 1 || got[0] != base {
			t.Fatalf("版本 %v 應回退共用資料 %q,got %v,%v", v, base, got, ok)
		}
	}
}

func TestBuildVersionAssetDirsDoesNotCrossVersionFallback(t *testing.T) {
	base13 := t.TempDir()
	dirs, err := buildVersionAssetDirs("", base13, "")
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := dirs.forVersion(gamedata.VersionClassic13); !ok || len(got) != 1 || got[0] != base13 {
		t.Fatalf("1.31 路徑錯誤:got %v,%v", got, ok)
	}
	if _, ok := dirs.forVersion(gamedata.VersionCommunity15); ok {
		t.Fatal("只有 -data13 時不應把 1.31 資產偽裝成 1.5")
	}
}

func TestInitialVersionDetectsOfficial131Readme(t *testing.T) {
	base := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "README.TXT"), []byte("Version 1.31\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirs, err := buildVersionAssetDirs(base, "", "")
	if err != nil {
		t.Fatal(err)
	}
	got, err := initialVersion("auto", dirs, base, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != gamedata.VersionClassic13 {
		t.Fatalf("README 的 1.31 應成為 auto 初始版本,got %v", got)
	}
}

func TestInitialVersionExplicit15Wins(t *testing.T) {
	base13 := t.TempDir()
	base15 := t.TempDir()
	dirs, err := buildVersionAssetDirs("", base13, base15)
	if err != nil {
		t.Fatal(err)
	}
	got, err := initialVersion("1.5", dirs, "", base13, base15)
	if err != nil {
		t.Fatal(err)
	}
	if got != gamedata.VersionCommunity15 {
		t.Fatalf("明指定 1.5 應選 1.5,got %v", got)
	}
}

func TestSelectGameVersionSwapsAssetResolver(t *testing.T) {
	base13 := t.TempDir()
	base15 := t.TempDir()
	if err := os.WriteFile(filepath.Join(base13, "VERSION.TXT"), []byte("classic"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base15, "VERSION.TXT"), []byte("community"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirs, err := buildVersionAssetDirs("", base13, base15)
	if err != nil {
		t.Fatal(err)
	}
	res, err := assets.NewResolver(base13)
	if err != nil {
		t.Fatal(err)
	}
	b := &sceneBuilder{res: res, versionAssets: dirs, gameVersion: gamedata.VersionClassic13}
	if err := b.selectGameVersion(gamedata.VersionCommunity15); err != nil {
		t.Fatal(err)
	}
	data, err := b.res.Read("version.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "community" || b.gameVersion != gamedata.VersionCommunity15 {
		t.Fatalf("版本切換未換資產解析器:version=%v,data=%q", b.gameVersion, data)
	}
}

func TestNewGameBackgroundAssetDiffersByGameVersion(t *testing.T) {
	if got := newGameBackgroundAsset(gamedata.VersionClassic13); got != 28 {
		t.Fatalf("1.31 NEWGAME 滿版背景應為 #28,got %d", got)
	}
	if got := newGameBackgroundAsset(gamedata.VersionCommunity15); got != 31 {
		t.Fatalf("1.5 NEWGAME 滿版背景應為 #31,got %d", got)
	}
}

func TestNewGameBackgroundAssetFallsBackToCommonAssetWhenArchiveIsOld(t *testing.T) {
	if got := newGameBackgroundAssetForCount(gamedata.VersionCommunity15, 30); got != 28 {
		t.Fatalf("只有 30 個資產的共用 NEWGAME.LBX 應由 1.5 的 #31 回退到 #28,got %d", got)
	}
	if got := newGameBackgroundAssetForCount(gamedata.VersionCommunity15, 33); got != 31 {
		t.Fatalf("完整 1.5 NEWGAME.LBX 應保留 #31,got %d", got)
	}
	if got := newGameBackgroundAssetForCount(gamedata.VersionClassic13, 28); got != 28 {
		t.Fatalf("資產數剛好到 #27 時不應誤選背景,got %d", got)
	}
	if got := newGameBackgroundAssetForCount(gamedata.VersionClassic13, 29); got != 28 {
		t.Fatalf("1.31 的 #28 應可用,got %d", got)
	}
}
