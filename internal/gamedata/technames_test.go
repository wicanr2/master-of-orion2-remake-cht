package gamedata

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loadTechTSVKeys 讀 assets/i18n/tech.tsv,回傳所有英文 key 的集合(第一欄,略過註解/空行)。
func loadTechTSVKeys(t *testing.T) map[string]bool {
	t.Helper()
	fp := filepath.Join("..", "..", "assets", "i18n", "tech.tsv")
	f, err := os.Open(fp)
	if err != nil {
		t.Fatalf("開啟 %s 失敗: %v", fp, err)
	}
	defer f.Close()

	keys := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) == 0 || cols[0] == "" {
			continue
		}
		keys[cols[0]] = true
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("讀取 %s 失敗: %v", fp, err)
	}
	return keys
}

// TestTechnologyNamesMatchTSV 確保 TechnologyNames 每一條 value 都能在 tech.tsv 查到中文,
// 否則 UI 顯示科技名時會查無中文、退回英文原字串(功能沒壞但翻譯缺漏),此測試用來擋這種漏網。
func TestTechnologyNamesMatchTSV(t *testing.T) {
	keys := loadTechTSVKeys(t)
	for tech, name := range TechnologyNames {
		if !keys[name] {
			t.Errorf("Technology(%d) 對應英文名 %q 不在 tech.tsv key 集合內", int(tech), name)
		}
	}
}

// TestTechnologyNameKnownSamples 抽樣核對幾個已知科技的英文名,防止轉寫規則跑掉。
func TestTechnologyNameKnownSamples(t *testing.T) {
	cases := []struct {
		tech Technology
		want string
	}{
		{TECH_AUTOMATED_FACTORIES, "Automated Factories"},
		{TECH_ANDROID_WORKERS, "Android Workers"},
		{TECH_DEATH_RAY, "Death Ray"},
		{TECH_TERRAFORMING, "Terraforming"},
		// 連字號拼寫是人工核對 tech.tsv 後的特例,非機械轉寫規則的預設輸出。
		{TECH_ANTIGRAV_HARNESS, "Anti-Grav Harness"},
		{TECH_BIOTERMINATOR, "Bio-Terminator"},
		{TECH_SUBSPACE_TELEPORTER, "Sub-Space Teleporter"},
	}
	for _, c := range cases {
		if got := TechnologyName(c.tech); got != c.want {
			t.Errorf("TechnologyName(%v) = %q,預期 %q", c.tech, got, c.want)
		}
	}
}

// TestTechnologyNameFallback 查無對應時應回退為 Tech#N,不 panic、不回空字串。
func TestTechnologyNameFallback(t *testing.T) {
	unknown := Technology(99999)
	if got := TechnologyName(unknown); got != "Tech#99999" {
		t.Errorf("TechnologyName(未知) = %q,預期 Tech#99999", got)
	}
}
