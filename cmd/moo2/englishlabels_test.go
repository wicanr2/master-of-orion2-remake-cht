package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestColonyPlanetRowsUseEnglishIDs(t *testing.T) {
	p := shell.Planet{
		Climate: "類地", Gravity: "常態", Mineral: "普通", Size: "中型",
		Gen: 1, ClimateID: gamedata.TERRAN, GravityID: gamedata.NORMAL_G,
		MineralID: gamedata.ABUNDANT, SizeID: gamedata.MEDIUM_PLANET,
		SpecialID: gamedata.AncientArtifacts,
	}
	got := colonyPlanetRows(i18n.English, p)
	joined := strings.Join(got, "|")
	for _, want := range []string{"Terran", "Medium", "Minerals Abundant", "Gravity Normal-G", "Ancient Artifacts"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("英文行星資訊缺少 %q: %q", want, joined)
		}
	}
	if strings.Contains(joined, "類地") || strings.Contains(joined, "常態") {
		t.Fatalf("英文行星資訊仍含中文: %q", joined)
	}
}

func TestRuntimeLabelCatalogCoverage(t *testing.T) {
	groups := [][]string{
		{"info.history.metric.population", "info.history.metric.technology", "info.history.metric.fleet", "info.history.metric.buildings"},
		{"newgame.value.difficulty.0", "newgame.value.difficulty.1", "newgame.value.difficulty.2", "newgame.value.difficulty.3", "newgame.value.difficulty.4"},
		{"newgame.value.size.0", "newgame.value.size.1", "newgame.value.size.2", "newgame.value.size.3"},
		{"newgame.value.age.0", "newgame.value.age.1", "newgame.value.age.2"},
		{"newgame.value.tech.0", "newgame.value.tech.1", "newgame.value.tech.2"},
		{"tactical.fighter.kind.interceptor", "tactical.fighter.kind.heavy", "tactical.fighter.kind.bomber", "tactical.fighter.kind.assault_shuttle"},
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, group := range groups {
			for _, key := range group {
				if got := uiText(lang, key); got == "" || got == key {
					t.Fatalf("語系 %v 缺少 %q", lang, key)
				}
			}
		}
	}
	formats := []struct {
		key   string
		count int
	}{
		{"galaxy.system.more_bodies", 1},
		{"combat.ship.named", 2},
		{"hotseat.player.numbered", 1},
		{"hotseat.player.numbered_race", 2},
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, format := range formats {
			if got := strings.Count(uiText(lang, format.key), "%"); got != format.count {
				t.Fatalf("語系 %v %q 格式欄位=%d，want %d", lang, format.key, got, format.count)
			}
		}
	}
}

func TestRuntimeLabelInvalidIndexesUseCurrentLanguageFallback(t *testing.T) {
	for _, label := range []func(i18n.Lang, int) string{
		newGameDifficultyLabel, newGameGalaxySizeLabel, newGameGalaxyAgeLabel, newGameTechLevelLabel,
	} {
		if got := label(i18n.Traditional, -1); got != "未知" {
			t.Fatalf("繁中無效索引 fallback=%q，want 未知", got)
		}
		if got := label(i18n.English, 999); got != "Unknown" {
			t.Fatalf("英文無效索引 fallback=%q，want Unknown", got)
		}
	}
}

func TestRuntimePlayerLabelsAreNotEmbeddedInEnglishLabelsSource(t *testing.T) {
	source, err := os.ReadFile("englishlabels.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, forbidden := range []string{
		`return "Technology"`, `return "Unknown"`, `"Tutor", "Easy"`, `"Small", "Medium"`,
		`"Young", "Average"`, `"Pre-warp", "Average"`, `"Unknown Empire"`, `"Unknown Ship"`,
		`"Heavy Fighter"`, `"Assault Shuttle"`, `fmt.Sprintf("Player %s`, `fmt.Sprintf("%s Ship %s`,
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("englishlabels.go 仍內嵌玩家文案 %q", forbidden)
		}
	}
}

func TestColonyBuildingLabel(t *testing.T) {
	for _, tc := range []struct {
		zh, en string
	}{
		{zh: "星基", en: "Star Base"},
		{zh: gamedata.TerraformActionName, en: "Terraforming"},
	} {
		if got := colonyBuildingLabel(i18n.English, tc.zh); got != tc.en {
			t.Errorf("%q 英文名 = %q, want %q", tc.zh, got, tc.en)
		}
		if got := colonyBuildingLabel(i18n.Traditional, tc.zh); got != tc.zh {
			t.Errorf("繁中名稱被改寫: %q -> %q", tc.zh, got)
		}
	}
}

func TestEnglishUnknownLabelsUseSafeFallback(t *testing.T) {
	if got := englishEnvironmentName("climate", 10, "舊存檔未知氣候"); got != "Unknown" {
		t.Fatalf("未知行星 enum 的英文 fallback = %q, want Unknown", got)
	}
	if got := colonyBuildingLabel(i18n.English, "舊存檔未知建築"); got != "Unknown Build" {
		t.Fatalf("未知建造項目的英文 fallback = %q, want Unknown Build", got)
	}
	if got := enemyDisplayName(i18n.English, &shell.GameSession{}, "舊存檔未知帝國"); got != "Unknown Empire" {
		t.Fatalf("未知帝國的英文 fallback = %q, want Unknown Empire", got)
	}
}

func TestBuildItemLabel(t *testing.T) {
	for _, tc := range []struct {
		name, want string
	}{
		{name: shell.TradeGoodsBuildName, want: "Trade Goods"},
		{name: shell.HousingBuildName, want: "Housing"},
		{name: "星基", want: "Star Base"},
	} {
		if got := buildItemLabel(i18n.English, tc.name); got != tc.want {
			t.Errorf("建造項目 %q 英文名 = %q, want %q", tc.name, got, tc.want)
		}
		if got := buildItemLabel(i18n.Traditional, tc.name); got != tc.name {
			t.Errorf("繁中建造項目被改寫: %q -> %q", tc.name, got)
		}
	}
}

func TestEnglishDisplayLabels(t *testing.T) {
	s := &shell.GameSession{AIPlayers: []shell.AIOpponent{{Name: "AI (席隆人)", RaceIndex: 1}}}
	if got := enemyDisplayName(i18n.English, s, "席隆人"); got != "Psilons" {
		t.Errorf("外交對手英文名 = %q, want Psilons", got)
	}
	if got := combatShipLabel(i18n.English, s, "席隆人艦2"); got != "Psilons Ship 2" {
		t.Errorf("敵艦英文名 = %q, want Psilons Ship 2", got)
	}
	if got := combatShipLabel(i18n.Traditional, s, "席隆人艦2"); got != "席隆人艦2" {
		t.Errorf("繁中敵艦名稱被改寫: %q", got)
	}
	if got := fighterKindLabel(i18n.English, shell.FighterInterceptor); got != "Interceptor" {
		t.Errorf("戰機英文名 = %q", got)
	}
	if got := hotseatNameLabel(i18n.English, "第 2 位玩家(布拉西人)"); got != "Player 2 (Bulrathi)" {
		t.Errorf("熱座英文名 = %q", got)
	}
}

func TestSystemBodyCountLabel(t *testing.T) {
	s := &shell.GameSession{Stars: []shell.Star{{Orbits: [5]int{0, 1, 2, shell.OrbitEmpty, shell.OrbitEmpty}}}}
	if got := systemBodyCountLabel(i18n.English, s, 0); got != "2 more" {
		t.Errorf("英文天體數 = %q, want 2 more", got)
	}
	if got := systemBodyCountLabel(i18n.Traditional, s, 0); got != "另有 2 天體" {
		t.Errorf("繁中天體數 = %q, want 另有 2 天體", got)
	}
}

func TestHistoryLabelsEnglish(t *testing.T) {
	if got := historyMetricLabel(i18n.English, shell.HistoryTechnology); got != "Technology" {
		t.Fatalf("科技指標 = %q", got)
	}
	s := &shell.GameSession{AIPlayers: []shell.AIOpponent{
		{Name: "AI (Psilons)", RaceIndex: 1},
		{Name: "AI (薩克拉)", RaceIndex: 0},
	}}
	got := historyEmpireLabels(i18n.English, s)
	if strings.Join(got, "|") != "You|Psilons|Sakkra" {
		t.Fatalf("英文歷史圖例 = %v", got)
	}
	zh := historyEmpireLabels(i18n.Traditional, s)
	if strings.Join(zh, "|") != "你|AI (Psilons)|AI (薩克拉)" {
		t.Fatalf("繁中歷史圖例被改寫 = %v", zh)
	}
}

func TestNewGameLabelsEnglish(t *testing.T) {
	checks := []struct {
		got, want string
	}{
		{newGameDifficultyLabel(i18n.English, 2), "Average"},
		{newGameGalaxySizeLabel(i18n.English, 1), "Medium"},
		{newGameGalaxyAgeLabel(i18n.English, 1), "Average"},
		{newGameTechLevelLabel(i18n.English, 1), "Average"},
	}
	for _, tc := range checks {
		if tc.got != tc.want {
			t.Errorf("英文 NEW GAME 值 = %q, want %q", tc.got, tc.want)
		}
	}
	if got := newGameDifficultyLabel(i18n.Traditional, 2); got != "普通" {
		t.Fatalf("繁中難度被改寫: %q", got)
	}
	if got := newGameGalaxySizeLabel(i18n.Traditional, 1); got != "中型" {
		t.Fatalf("繁中星系大小被改寫: %q", got)
	}
}
