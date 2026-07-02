package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRegistryLoadAssets 載入 repo 內 assets/i18n,驗證 per-source 查詢與 merged 備援。
func TestRegistryLoadAssets(t *testing.T) {
	dir := filepath.Join("..", "..", "assets", "i18n")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("無 assets/i18n")
	}
	reg := NewRegistry(Traditional)
	n, err := reg.LoadFS(os.DirFS(dir), ".")
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	if n == 0 {
		t.Fatal("未載入任何來源")
	}
	t.Logf("載入 %d 個來源: %v", n, reg.Sources())

	// per-source 查詢:tech 來源應能翻譯已知科技名。
	if got := reg.Source("tech").Translate("Armor"); got != "裝甲車" {
		t.Errorf("tech Armor = %q,預期 裝甲車(地面單位)", got)
	}
	// misc 來源的同形詞 Armor 應為艦艇裝備分組「裝甲」,與 tech 不互相覆蓋。
	if got := reg.Source("misc").Translate("Armor"); got != "裝甲" {
		t.Errorf("misc Armor = %q,預期 裝甲(艦艇裝備分組)", got)
	}

	// merged 備援可查(不指定來源時)。
	if got := reg.Translate("Armor"); got == "Armor" {
		t.Error("merged 查無 Armor,備援表未併入")
	}

	// 查無來源回空表、不 panic,查無 key 回原字串。
	if got := reg.Source("不存在").Translate("Foobar"); got != "Foobar" {
		t.Errorf("空來源應回原字串,得 %q", got)
	}
}

// TestRegistryEnglishPassthrough 英文模式下所有來源直通原字串。
func TestRegistryEnglishPassthrough(t *testing.T) {
	dir := filepath.Join("..", "..", "assets", "i18n")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("無 assets/i18n")
	}
	reg := NewRegistry(English)
	if _, err := reg.LoadFS(os.DirFS(dir), "."); err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	if got := reg.Source("tech").Translate("Armor"); got != "Armor" {
		t.Errorf("英文模式應直通,得 %q", got)
	}
	reg.SetLang(Traditional) // 切語言後應生效
	if got := reg.Source("tech").Translate("Armor"); got != "裝甲車" {
		t.Errorf("切繁中後 Armor = %q,預期 裝甲車", got)
	}
}

// TestRaceTraitConsistency 守護:種族選擇(races)與資訊面板(raceinfo)是同批 RACESTUF 特性,
// 玩家直接可對照,共享的英文 key 譯文必須一致(見架構文件的統一決策)。
func TestRaceTraitConsistency(t *testing.T) {
	dir := filepath.Join("..", "..", "assets", "i18n")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("無 assets/i18n")
	}
	reg := NewRegistry(Traditional)
	if _, err := reg.LoadFS(os.DirFS(dir), "."); err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	races := reg.Source("races").m
	raceinfo := reg.Source("raceinfo").m
	for k, v := range raceinfo {
		if rv, ok := races[k]; ok && rv != v {
			t.Errorf("種族特性不一致 %q:races=%q raceinfo=%q(兩畫面可對照,須統一)", k, rv, v)
		}
	}
}
