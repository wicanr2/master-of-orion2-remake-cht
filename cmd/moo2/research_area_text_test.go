package main

import (
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestResearchAreaTextCatalogAndFormatContract(t *testing.T) {
	tests := []struct {
		key         string
		formatCount int
		args        []any
	}{
		{"research.area.topic_cost", 2, []any{"Trans Dimensional Physics", 999999}},
		{"research.area.hyper_level_cost", 3, []any{"Hyper-Advanced Construction", 99, 999999}},
		{"research.area.complete", 0, nil},
		{"research.area.transition.galaxy", 0, nil},
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, tc := range tests {
			template := uiText(lang, tc.key)
			if template == "" || template == tc.key {
				t.Fatalf("語系 %v 缺少 %q", lang, tc.key)
			}
			if got := strings.Count(template, "%"); got != tc.formatCount {
				t.Fatalf("語系 %v %q 格式欄位=%d，want %d：%q", lang, tc.key, got, tc.formatCount, template)
			}
			if got := researchAreaText(lang, tc.key, tc.args...); strings.Contains(got, "%!") {
				t.Fatalf("語系 %v %q 格式化失敗：%q", lang, tc.key, got)
			}
		}
	}
}

func TestResearchAreaDynamicTextIsNotEmbeddedInResearchFunction(t *testing.T) {
	source, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	all := string(source)
	start := strings.Index(all, "func (b *sceneBuilder) research()")
	end := strings.Index(all[start:], "// planetListRows")
	if start < 0 || end < 0 {
		t.Fatal("找不到 research() 來源切片")
	}
	section := all[start : start+end]
	for _, forbidden := range []string{".tr(", "星系主畫面", "%s ・ %d RP", "已完成本領域全部科技"} {
		if strings.Contains(section, forbidden) {
			t.Fatalf("research() 仍內嵌玩家文案 %q", forbidden)
		}
	}
}

func TestResearchAreaTopicTextRectsStayInsideAreaPanels(t *testing.T) {
	if len(researchAreaHits) != 8 {
		t.Fatalf("研究領域熱區=%d，want 8", len(researchAreaHits))
	}
	for _, hit := range researchAreaHits {
		r := researchAreaTopicTextRect(hit)
		if r.x < hit.x || r.y < hit.y || r.x+r.w > hit.x+hit.w || r.y+r.h > hit.y+hit.h {
			t.Fatalf("%s 文字安全框 %+v 超出面板 %+v", hit.action, r, hit)
		}
		if r.y < hit.y+26 {
			t.Fatalf("%s 文字安全框侵入標題帶：%+v", hit.action, r)
		}
	}
}

func TestResearchAreaLongDynamicLabelsAreWidthBounded(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		values := []string{
			researchAreaText(lang, "research.area.topic_cost", strings.Repeat("Trans Dimensional 銀河科技", 8), 999999),
			researchAreaText(lang, "research.area.hyper_level_cost", strings.Repeat("Hyper-Advanced 超進階", 8), 99, 999999),
			researchAreaText(lang, "research.area.complete"),
		}
		for _, hit := range researchAreaHits {
			r := researchAreaTopicTextRect(hit)
			for _, value := range values {
				e := centeredExtraTextInSafeRect(r, 12, value, color.RGBA{})
				if e.maxW != r.contentWidth() {
					t.Fatalf("%s maxW=%.0f，want %.0f", hit.action, e.maxW, r.contentWidth())
				}
				clipped := r.clipped(fnt, value, 12)
				if w, _ := fnt.Measure(clipped, 12); w > r.contentWidth() {
					t.Fatalf("%s 截斷後仍超寬：%.0f > %.0f，%q", hit.action, w, r.contentWidth(), clipped)
				}
			}
		}
	}
}
