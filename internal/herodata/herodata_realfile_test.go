package herodata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
)

// realfile 測試:設 MOO2_HERODATA_TEST(或 MOO2_AUDIO_TEST)指向真實 gamedata 目錄才跑,
// 驗證 parser 能吃下原版 HERODATA.LBX。測資為版權物,不入 repo,故只斷言「結構」
// (筆數/欄位合理性),不硬編任何英雄名。
func TestParseRealHerodata(t *testing.T) {
	dir := os.Getenv("MOO2_HERODATA_TEST")
	if dir == "" {
		dir = os.Getenv("MOO2_AUDIO_TEST")
	}
	if dir == "" {
		t.Skip("未設 MOO2_HERODATA_TEST / MOO2_AUDIO_TEST,跳過真實 HERODATA 測試")
	}
	f, err := os.Open(filepath.Join(dir, "HERODATA.LBX"))
	if err != nil {
		t.Skipf("開 HERODATA.LBX 失敗(跳過):%v", err)
	}
	defer f.Close()
	fi, _ := f.Stat()
	arch, err := lbx.Open(f, fi.Size())
	if err != nil {
		t.Fatalf("lbx.Open: %v", err)
	}
	raw, err := arch.Asset(0)
	if err != nil {
		t.Fatalf("讀 asset 0: %v", err)
	}
	leaders, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// 結構斷言(不硬編英雄名):真檔應為 67 筆。
	if len(leaders) != 67 {
		t.Errorf("英雄數 = %d,預期 67", len(leaders))
	}
	named, ship, colony := 0, 0, 0
	for _, l := range leaders {
		if l.Name != "" {
			named++
		}
		if l.Ship() {
			ship++
		} else {
			colony++
		}
	}
	if named < 60 {
		t.Errorf("有名字的英雄僅 %d,預期近 67(name 欄位解析可能有誤)", named)
	}
	if ship == 0 || colony == 0 {
		t.Errorf("艦艇軍官/殖民地領袖分類異常:ship=%d colony=%d(type 欄位解析可能有誤)", ship, colony)
	}
	t.Logf("HERODATA 解析成功:%d 英雄(有名 %d、艦艇軍官 %d、殖民地領袖 %d)", len(leaders), named, ship, colony)
}
