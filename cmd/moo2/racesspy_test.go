package main

import (
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// 種族關係畫面的間諜區塊座標來自**執行檔的 .data 靜態表**,不是量圖。
//
// 這張畫面在 openorion2 是 STUB,所以 remake 先前只能量圖(自編的「y 從 66 起、每列 +62」)。
// 現在有更上游的來源,而來源優先序是專案硬規則——這一支擋住「改回自編排版」。
func TestRacesSpyAnchorsComeFromTheExecutable(t *testing.T) {
	want := [racesMaxRows][2]int{
		{120, 126}, {120, 233}, {120, 338}, {120, 445},
		{332, 126}, {332, 233}, {332, 338},
	}
	if racesSpyAnchors != want {
		t.Errorf("間諜鈕錨點應是 %v,得到 %v", want, racesSpyAnchors)
	}
}

// 版面的形狀:左欄 4 列 + 右欄 3 列,兩欄 y 完全對齊。
//
// 這是那張表最顯眼的結構,抄錯任何一筆都會破——比逐筆比對更能抓到「順序弄反」這種錯。
func TestRacesLayoutIsFourLeftThreeRightWithSharedYs(t *testing.T) {
	const leftX, rightX = 120, 332
	for i := 0; i < 4; i++ {
		if racesSpyAnchors[i][0] != leftX {
			t.Errorf("第 %d 列應在左欄 x=%d,得到 %d", i, leftX, racesSpyAnchors[i][0])
		}
	}
	for i := 4; i < racesMaxRows; i++ {
		if racesSpyAnchors[i][0] != rightX {
			t.Errorf("第 %d 列應在右欄 x=%d,得到 %d", i, rightX, racesSpyAnchors[i][0])
		}
	}
	// 右欄三列的 y 與左欄前三列相同。
	for i := 0; i < 3; i++ {
		if racesSpyAnchors[i][1] != racesSpyAnchors[i+4][1] {
			t.Errorf("左欄第 %d 列 y=%d 應與右欄第 %d 列 y=%d 對齊",
				i, racesSpyAnchors[i][1], i+4, racesSpyAnchors[i+4][1])
		}
	}
}

// 關係滑桿也是同一種兩欄版面(x=105 / 528)。
func TestRacesRelationBarsAreTwoColumns(t *testing.T) {
	for i := 0; i < 4; i++ {
		if racesRelationBars[i][0] != 105 {
			t.Errorf("第 %d 條滑桿應在 x=105,得到 %d", i, racesRelationBars[i][0])
		}
	}
	for i := 4; i < racesMaxRows; i++ {
		if racesRelationBars[i][0] != 528 {
			t.Errorf("第 %d 條滑桿應在 x=528,得到 %d", i, racesRelationBars[i][0])
		}
	}
}

// 熱區數量跟著 AI 數走,並夾在版面上限(原版最多同時顯示 7 個已接觸種族)。
func TestRacesSpyHitRegionsClampToTheLayout(t *testing.T) {
	for _, n := range []int{0, 1, 3, 7} {
		if got := len(racesSpyHitRegions(n)); got != n {
			t.Errorf("%d 個 AI 應有 %d 個熱區,得到 %d", n, n, got)
		}
	}
	// 超過版面上限要夾住,不能越界索引那張 7 筆的表。
	if got := len(racesSpyHitRegions(99)); got != racesMaxRows {
		t.Errorf("超過上限應夾在 %d,得到 %d", racesMaxRows, got)
	}
}

// 動作字串 round-trip,而且不會把別的動作誤判成間諜。
func TestRacesSpyActionRoundTrips(t *testing.T) {
	for i := 0; i < racesMaxRows; i++ {
		got, ok := racesSpyActionIndex(racesSpyAction(i))
		if !ok || got != i {
			t.Errorf("第 %d 個動作 round-trip 失敗:得到 (%d,%v)", i, got, ok)
		}
	}
	// 這張畫面上其他動作不能被當成間諜——誤判會讓「宣戰」變成派間諜。
	for _, a := range []string{"audience", "declarewar", "report", "back", "spy", "spyX", "spy99"} {
		if _, ok := racesSpyActionIndex(a); ok {
			t.Errorf("%q 不該被判成間諜動作", a)
		}
	}
}

func TestRacesSpyMissionActionsAndRegions(t *testing.T) {
	for i := 0; i < racesMaxRows; i++ {
		got, ok := racesSpyMissionActionIndex(racesSpyMissionAction(i))
		if !ok || got != i {
			t.Fatalf("第 %d 個任務動作 round-trip 失敗:得到 (%d,%v)", i, got, ok)
		}
	}
	regions := racesSpyMissionHitRegions(racesMaxRows)
	if len(regions) != racesMaxRows {
		t.Fatalf("任務熱區數量 = %d,預期 %d", len(regions), racesMaxRows)
	}
	for i, region := range regions {
		wantX, wantY, wantW, wantH := racesSpySlotRect(i, 1)
		if region.x != wantX || region.y != wantY || region.w != wantW || region.h != wantH {
			t.Errorf("第 %d 個任務熱區 = %+v,預期 (%d,%d,%d,%d)",
				i, region, wantX, wantY, wantW, wantH)
		}
	}
	for _, action := range []string{"spy0", "spymission", "spymissionX", "spymission99", "report"} {
		if _, ok := racesSpyMissionActionIndex(action); ok {
			t.Errorf("%q 不應被判成合法任務動作", action)
		}
	}
}

func TestRacesSpySlotsStayInsideTheNativeButtonRow(t *testing.T) {
	for i := 0; i < racesMaxRows; i++ {
		prevRight := -1
		for slot := 0; slot < racesSpyButtonSlots; slot++ {
			x, y, w, h := racesSpySlotRect(i, slot)
			if y != racesSpyAnchors[i][1] || h != racesSpyButtonH {
				t.Fatalf("第 %d 列第 %d 槽不在原生按鈕列：(%d,%d,%d,%d)", i, slot, x, y, w, h)
			}
			if x < prevRight {
				t.Fatalf("第 %d 列第 %d 槽與前一槽重疊：x=%d, previous right=%d", i, slot, x, prevRight)
			}
			prevRight = x + w
		}
	}
}

func TestRacesSpyHideActionsAndRegions(t *testing.T) {
	for i := 0; i < racesMaxRows; i++ {
		got, ok := racesSpyHideActionIndex(racesSpyHideAction(i))
		if !ok || got != i {
			t.Fatalf("第 %d 個隱匿動作 round-trip 失敗:得到 (%d,%v)", i, got, ok)
		}
	}
	regions := racesSpyHideHitRegions(racesMaxRows)
	for i, region := range regions {
		wantX, wantY, wantW, wantH := racesSpySlotRect(i, 2)
		if region.x != wantX || region.y != wantY || region.w != wantW || region.h != wantH {
			t.Errorf("第 %d 個隱匿熱區 = %+v,預期 (%d,%d,%d,%d)",
				i, region, wantX, wantY, wantW, wantH)
		}
	}
}

func TestRacesDiplomacyActionsAndRegions(t *testing.T) {
	for i := 0; i < racesMaxRows; i++ {
		got, ok := racesDiplomacyActionIndex(racesDiplomacyAction(i))
		if !ok || got != i {
			t.Fatalf("第 %d 個外交動作 round-trip 失敗:得到 (%d,%v)", i, got, ok)
		}
	}
	regions := racesDiplomacyHitRegions(racesMaxRows)
	if len(regions) != racesMaxRows {
		t.Fatalf("外交熱區數量 = %d,預期 %d", len(regions), racesMaxRows)
	}
	for i, region := range regions {
		want := racesInfoRect(i)
		if region.x != want.x || region.y != want.y || region.w != want.w || region.h != want.h {
			t.Errorf("第 %d 個外交熱區 = %+v,座標或尺寸不符", i, region)
		}
	}
	for _, action := range []string{"report", "racediplomacy", "racediplomacyX", "racediplomacy99", "spy0"} {
		if _, ok := racesDiplomacyActionIndex(action); ok {
			t.Errorf("%q 不應被判成合法外交動作", action)
		}
	}
}

func TestRacesInfoTextStopsBeforeSliderAndSpyButtons(t *testing.T) {
	for i := 0; i < racesMaxRows; i++ {
		r := racesInfoRect(i)
		if r.x < 0 || r.y < 0 || r.x+r.w > racesRelationBars[i][0]-4 {
			t.Fatalf("第 %d 列資訊框越過關係滑桿：(%d,%d,%d,%d),滑桿 x=%d", i, r.x, r.y, r.w, r.h, racesRelationBars[i][0])
		}
		if r.y+r.h > racesSpyAnchors[i][1]-4 {
			t.Fatalf("第 %d 列資訊框越過間諜鈕：bottom=%d, spyY=%d", i, r.y+r.h, racesSpyAnchors[i][1])
		}
		for line := 0; line < 4; line++ {
			lr := racesInfoLineRect(i, line)
			if lr.x < r.x || lr.y < r.y || lr.x+lr.w > r.x+r.w || lr.y+lr.h > r.y+r.h {
				t.Fatalf("第 %d 列第 %d 行安全框 (%d,%d,%d,%d) 超出資料區 (%d,%d,%d,%d)",
					i, line, lr.x, lr.y, lr.w, lr.h, r.x, r.y, r.w, r.h)
			}
			if line > 0 {
				prev := racesInfoLineRect(i, line-1)
				if prev.y+prev.h > lr.y {
					t.Fatalf("第 %d 列資訊行 %d/%d 重疊：%+v %+v", i, line-1, line, prev, lr)
				}
			}
		}
	}
}

// 點陣字在 9–11px 請求下仍以完整 12px 字身繪製；每一行安全框必須容納
// 實際字身，而不是只驗證資料框座標沒有越界。
func TestRacesInfoTextSafeRectsContainBitmapGlyphs(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	samples := []struct {
		line int
		size float64
		text string
	}{
		{0, 11, "AI（自訂長名稱）"},
		{1, 10, "對你：非常友善"},
		{2, 10, "軍 999・星 99"},
		{3, 9, "他國：AI（自訂）非常友善"},
	}
	for _, sample := range samples {
		for i := 0; i < racesMaxRows; i++ {
			r := racesInfoLineRect(i, sample.line)
			clipped := r.clipped(fnt, sample.text, sample.size)
			w, h := fnt.Measure(clipped, sample.size)
			if w > r.contentWidth() || h > float64(r.h-2*r.insetY) {
				t.Fatalf("第 %d 列第 %d 行字身超出安全框：量測 %.1fx%.1f，內容 %.1fx%d，文字 %q",
					i, sample.line, w, h, r.contentWidth(), r.h-2*r.insetY, clipped)
			}
		}
	}
	// 最長的外交關係摘要必須經過同一個安全框截斷，不能把原始字串直接交給 overlay。
	r := racesInfoLineRect(0, 3)
	if got := r.clipped(fnt, strings.Repeat("他國：AI：非常友善、", 20), 9); got == "" {
		t.Fatal("外交關係摘要不應被整段吞掉")
	}
}

func TestRacesAgentControlsAreExplicitAndSeparated(t *testing.T) {
	regions := racesAgentHitRegions()
	if len(regions) != 2 {
		t.Fatalf("防守 Agent 應有訓練／解除兩個熱區,得到 %d", len(regions))
	}
	if regions[0].action != racesAgentTrainAction() || regions[1].action != racesAgentDismissAction() {
		t.Fatalf("防守 Agent 動作=%q/%q", regions[0].action, regions[1].action)
	}
	if regions[0].x+racesAgentW > regions[1].x {
		t.Fatal("防守 Agent 兩個熱區不應互相重疊")
	}
	checked := append(append([]hitRegion(nil), regions...), hitRegion{
		x: racesAgentStatusX, y: racesAgentStatusY, w: racesAgentStatusW, h: racesAgentStatusH,
	})
	for _, region := range checked {
		if region.y+region.h > 418 {
			t.Fatalf("Agent 控制侵入外交按鈕列：%+v", region)
		}
		for i := 0; i < racesMaxRows; i++ {
			info := racesInfoRect(i)
			if rectsOverlap(region.x, region.y, region.w, region.h, info.x, info.y, info.w, info.h) {
				t.Fatalf("Agent 控制侵入第 %d 個帝國資訊槽：%+v / %+v", i, region, info)
			}
			for slot := 0; slot < racesSpyButtonSlots; slot++ {
				x, y, w, h := racesSpySlotRect(i, slot)
				if rectsOverlap(region.x, region.y, region.w, region.h, x, y, w, h) {
					t.Fatalf("Agent 控制侵入第 %d 個帝國第 %d 間諜槽", i, slot)
				}
			}
		}
	}
}

func rectsOverlap(ax, ay, aw, ah, bx, by, bw, bh int) bool {
	return ax < bx+bw && bx < ax+aw && ay < by+bh && by < ay+ah
}

func TestRacesActionTextCentersInsideItsClickableRect(t *testing.T) {
	spyRect := racesSpySlotTextRect(0, 0)
	spy := centeredExtraTextInSafeRect(spyRect, 10, "增派 0", color.RGBA{})
	if spy.align != 1 || spy.x != float64(spyRect.x)+float64(spyRect.w)/2 || spy.y != float64(spyRect.y)+float64(spyRect.h)/2 {
		t.Fatalf("派間諜文字未置中於熱區：%+v", spy)
	}
	if spy.maxW != float64(racesSpyButtonW-6) {
		t.Fatalf("派間諜文字最大寬度 = %.0f, want %d", spy.maxW, racesSpyButtonW-6)
	}
	trainRect := racesAgentTrainTextRect()
	train := centeredExtraTextInSafeRect(trainRect, 10, "訓練（30 BC）", color.RGBA{})
	if train.align != 1 || train.x != float64(racesAgentTrainX+racesAgentW/2) || train.y != float64(racesAgentY+racesAgentH/2) {
		t.Fatalf("訓練 Agent 文字未置中於熱區：%+v", train)
	}
}

func TestRacesExternalTextCatalogAndRuntimeBounds(t *testing.T) {
	keys := []string{
		"races.transition.council", "races.transition.screen", "races.transition.galaxy",
		"races.info.you", "races.info.power_stars", "races.info.others", "races.info.none",
		"races.info.relation_entry", "races.info.relation_separator", "races.spy.add",
		"races.spy.mission.steal", "races.spy.mission.sabotage", "races.spy.mission.hide",
		"races.agent.status", "races.agent.train", "races.agent.dismiss",
	}
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, got)
			}
		}
		checkClippedTextFits(t, fnt, racesInfoLineRect(0, 1),
			racesText(lang, "races.info.you", infoStanceLabel(lang, "提議貿易")), 10)
		checkClippedTextFits(t, fnt, racesInfoLineRect(0, 2),
			racesText(lang, "races.info.power_stars", 999999, 99), 10)
		checkClippedTextFits(t, fnt, racesInfoLineRect(0, 3),
			racesText(lang, "races.info.others", strings.Repeat("Long Empire: Allied, ", 10)), 9)
		for slot, mission := range []shell.SpyMission{shell.SpyMissionSteal, shell.SpyMissionSabotage, shell.SpyMissionHide} {
			checkClippedTextFits(t, fnt, racesSpySlotTextRect(0, slot), uiText(lang, racesSpyMissionTextKey(mission)), 10)
		}
		checkClippedTextFits(t, fnt, racesAgentStatusTextRect(), racesText(lang, "races.agent.status", 63, 999999), 10)
		checkClippedTextFits(t, fnt, racesAgentTrainTextRect(), uiText(lang, "races.agent.train"), 10)
		checkClippedTextFits(t, fnt, racesAgentDismissTextRect(), uiText(lang, "races.agent.dismiss"), 10)
	}
}

func TestRacesSourceAndRulesHaveNoEmbeddedPlayerLabels(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func (b *sceneBuilder) races()")
	if start < 0 {
		t.Fatal("races source start not found")
	}
	end := strings.Index(source[start:], "// --- 外交對談畫面")
	if end < 0 {
		t.Fatal("races source end not found")
	}
	if strings.Contains(source[start:start+end], ".tr(") {
		t.Fatal("races() must not embed translated player text")
	}
	for _, path := range []string{"../../internal/shell/spy_mission.go", "../../internal/shell/session.go"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"SpyMissionLabel", "AIRelationName"} {
			if strings.Contains(string(raw), forbidden) {
				t.Fatalf("%s still exposes player label helper %s", path, forbidden)
			}
		}
	}
}
