package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// hotkeys_test.go:星圖快捷鍵的護欄(規則層的循環邏輯在 internal/shell/starnav_test.go)。

func hotkeyBuilder(nStars int) *sceneBuilder {
	sess := &shell.GameSession{}
	sess.Stars = make([]shell.Star, nStars)
	for i := range sess.Stars {
		sess.Stars[i].Explored = true
	}
	sess.SelectedStar = -1
	sess.Fleet().AtStar = -1
	b := &sceneBuilder{session: sess}
	b.measure.from = -1
	return b
}

// TestMeasureModeTogglesWithF9:F9 進測距、再按一次離開。
//
// 不給出口會把玩家鎖在測距模式裡(原版手冊沒寫怎麼離開,「同一個鍵切換」是這裡的選擇)。
func TestMeasureModeTogglesWithF9(t *testing.T) {
	b := hotkeyBuilder(5)
	if !b.handleGalaxyHotkey(shell.HotkeyMeasureDist) || !b.measure.on {
		t.Fatal("按 F9 應進入測距模式")
	}
	if !b.handleGalaxyHotkey(shell.HotkeyMeasureDist) || b.measure.on {
		t.Fatal("再按一次 F9 應離開測距模式")
	}
	if b.measure.from != -1 {
		t.Errorf("離開測距時起點應清掉,實得 %d", b.measure.from)
	}
}

// TestMeasureEatsStarClick:測距模式下點星是「定起點」,不是「選這顆星」。
//
// 這條擋的是把兩件事混在一起——玩家按了 F9 之後點星,不該同時跳出星系資訊面板。
func TestMeasureEatsStarClick(t *testing.T) {
	b := hotkeyBuilder(5)
	if b.measureClickedStar(2) {
		t.Fatal("沒進測距模式時不該吃掉點星")
	}
	b.handleGalaxyHotkey(shell.HotkeyMeasureDist)
	if !b.measureClickedStar(2) {
		t.Fatal("測距模式下應吃掉點星")
	}
	if b.measure.from != 2 {
		t.Errorf("起點應記成 2,實得 %d", b.measure.from)
	}
	// 再點一顆 = 換起點(手冊只描述「點第一顆星」,沒有第二次點擊的語意)。
	b.measureClickedStar(4)
	if b.measure.from != 4 {
		t.Errorf("再點一顆應換起點成 4,實得 %d", b.measure.from)
	}
	if b.session.SelectedStar != -1 {
		t.Errorf("測距的點擊不該動到選取星,實得 %d", b.session.SelectedStar)
	}
}

// TestColonyHotkeyMovesSelection:F5/F6 只換選取星,不動遊戲狀態。
func TestColonyHotkeyMovesSelection(t *testing.T) {
	b := hotkeyBuilder(10)
	b.session.PlayerColonyStars = []int{6, 2}
	settings := b.session.EffectiveGameSettings()
	settings.AutoSelectColony = true
	b.session.ApplyGameSettings(settings)
	if !b.handleGalaxyHotkey(shell.HotkeyNextColony) {
		t.Fatal("有殖民地時 F5 應有作用")
	}
	if b.session.SelectedStar != 2 {
		t.Errorf("F5 應選到索引最小的殖民星 2,實得 %d", b.session.SelectedStar)
	}
	if b.session.EffectiveGameSettings().AutoSelectColony {
		t.Fatal("手動 F5 巡覽後應依原版關閉 Auto Select Colony")
	}
	b.handleGalaxyHotkey(shell.HotkeyNextColony)
	if b.session.SelectedStar != 6 {
		t.Errorf("再按 F5 應到 6,實得 %d", b.session.SelectedStar)
	}
	b.handleGalaxyHotkey(shell.HotkeyPrevColony)
	if b.session.SelectedStar != 2 {
		t.Errorf("F6 應退回 2,實得 %d", b.session.SelectedStar)
	}
}

func TestAutoSelectedColonyIndex(t *testing.T) {
	b := hotkeyBuilder(8)
	b.session.PlayerColonyStars = []int{6, 2}
	settings := b.session.EffectiveGameSettings()
	settings.AutoSelectColony = true
	b.session.ApplyGameSettings(settings)

	if got := autoSelectedColonyIndex(b.session, 2); got != 1 {
		t.Fatalf("玩家殖民星 2 應映射到殖民地索引 1，實得 %d", got)
	}
	if got := autoSelectedColonyIndex(b.session, 4); got != -1 {
		t.Fatalf("非玩家殖民星不應自動進入，實得 %d", got)
	}
	settings.AutoSelectColony = false
	b.session.ApplyGameSettings(settings)
	if got := autoSelectedColonyIndex(b.session, 2); got != -1 {
		t.Fatalf("設定關閉時不應自動進入，實得 %d", got)
	}
	if got := autoSelectedColonyIndex(nil, 2); got != -1 {
		t.Fatalf("nil session 應安全拒絕，實得 %d", got)
	}
}

func TestColonyHotkeyWithoutTargetKeepsAutoSelect(t *testing.T) {
	b := hotkeyBuilder(4)
	settings := b.session.EffectiveGameSettings()
	settings.AutoSelectColony = true
	b.session.ApplyGameSettings(settings)
	if b.handleGalaxyHotkey(shell.HotkeyNextColony) {
		t.Fatal("沒有殖民地時 F5 不應成功")
	}
	if !b.session.EffectiveGameSettings().AutoSelectColony {
		t.Fatal("沒有巡覽目標時不應清除 Auto Select Colony")
	}
}

// TestHotkeyWithoutTargetsDoesNothing:沒有殖民地/艦隊時按鍵不能亂改選取。
func TestHotkeyWithoutTargetsDoesNothing(t *testing.T) {
	b := hotkeyBuilder(10)
	b.session.SelectedStar = 3
	for _, k := range []string{shell.HotkeyNextColony, shell.HotkeyPrevColony,
		shell.HotkeyNextFleet, shell.HotkeyPrevFleet} {
		if b.handleGalaxyHotkey(k) {
			t.Errorf("沒有對象時 %s 不該回報「要重繪」", k)
		}
	}
	if b.session.SelectedStar != 3 {
		t.Errorf("選取星不該被動到,實得 %d", b.session.SelectedStar)
	}
}

// TestHotkeyNilSessionIsSafe:主選單等沒有對局的畫面按到快捷鍵不能炸。
func TestHotkeyNilSessionIsSafe(t *testing.T) {
	b := &sceneBuilder{}
	if b.handleGalaxyHotkey(shell.HotkeyNextColony) {
		t.Error("沒有對局時不該有作用")
	}
	if b.handleGalaxyHotkey("") {
		t.Error("空字串不該有作用")
	}
}

// TestStarAtScreenMatchesClickBox:懸停判定要和點選熱區用同一個方框。
//
// 星圖的點選熱區是以星球螢幕座標為中心的 22×22 方框(starHitHalf)。測距要「移到哪就
// 顯示到哪」,如果兩邊各寫一份判定,就會出現「點得到卻懸停不到」的邊緣像素。
func TestStarAtScreenMatchesClickBox(t *testing.T) {
	b := hotkeyBuilder(3)
	// 給三顆分得開的星(正規化座標 0..1)。
	b.session.Stars[0].X, b.session.Stars[0].Y = 0.2, 0.2
	b.session.Stars[1].X, b.session.Stars[1].Y = 0.5, 0.5
	b.session.Stars[2].X, b.session.Stars[2].Y = 0.8, 0.8
	for i := range b.session.Stars {
		x, y := starScreenPos(b.session.Stars[i])
		if got := b.starAtScreen(x, y); got != i {
			t.Errorf("星心 (%d,%d) 應命中星 %d,實得 %d", x, y, i, got)
		}
		// 方框內的最遠角落仍要命中;再往外一格就不該命中。
		if got := b.starAtScreen(x+starHitHalf-1, y+starHitHalf-1); got != i {
			t.Errorf("方框內緣應仍命中星 %d,實得 %d", i, got)
		}
		if got := b.starAtScreen(x+starHitHalf, y+starHitHalf); got == i {
			t.Errorf("方框外(+%d)不該命中星 %d", starHitHalf, i)
		}
	}
}

// TestQuickSaveWritesAndReports:F10 要真的寫檔,而且要回報。
//
// 「回報」不是裝飾:原版的 F10 是**直接覆蓋**、沒有對話框,玩家按下去唯一的回饋就是那句話。
// 沒有它,存成功與存失敗看起來完全一樣。
func TestQuickSaveWritesAndReports(t *testing.T) {
	b := hotkeyBuilder(3)
	path := filepath.Join(t.TempDir(), "quick.json")
	b.savePath = path

	if !b.handleGalaxyHotkey(shell.HotkeyQuickSave) {
		t.Fatal("F10 應被處理")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("存檔沒寫出來:%v", err)
	}
	if b.flashMsg == "" {
		t.Error("存完應該有回報訊息")
	}
	if b.flashUntil <= b.animTick {
		t.Errorf("訊息應該還在有效期內(animTick %d,flashUntil %d)", b.animTick, b.flashUntil)
	}
}

// TestQuickSaveWithoutSlotSaysSo:沒有存檔位置時要說出來,不能安靜地什麼都不做。
func TestQuickSaveWithoutSlotSaysSo(t *testing.T) {
	b := hotkeyBuilder(3)
	b.savePath = ""
	b.handleGalaxyHotkey(shell.HotkeyQuickSave)
	if b.flashMsg == "" {
		t.Error("沒有存檔位置時應該回報,不能靜默")
	}
}

// TestFlashExpires:訊息要自己消失——一直掛著的「已存檔」會被誤讀成「還在存」。
func TestFlashExpires(t *testing.T) {
	b := hotkeyBuilder(3)
	b.animTick = 100
	b.flash("測試")
	if b.flashUntil != 100+flashTicks {
		t.Errorf("到期時間應為 %d,實得 %d", 100+flashTicks, b.flashUntil)
	}
	b.animTick = b.flashUntil - 1
	if !b.flashVisible() {
		t.Error("到期前一拍應仍要畫")
	}
	b.animTick = b.flashUntil
	if b.flashVisible() {
		t.Error("到期後不該再畫")
	}
	// 沒有訊息內容時不該畫(避免畫一個空底框)。
	b.flashMsg, b.animTick = "", 0
	if b.flashVisible() {
		t.Error("沒有訊息時不該畫")
	}
}

func TestHotkeyTextCatalogAndFormatContract(t *testing.T) {
	tests := []struct {
		key         string
		formatCount int
		args        []any
	}{
		{"hotkey.quicksave.no_slot", 0, nil},
		{"hotkey.quicksave.failed", 1, []any{"permission denied"}},
		{"hotkey.quicksave.success", 0, nil},
		{"hotkey.measure.select_origin", 0, nil},
		{"hotkey.measure.hover_target", 0, nil},
		{"hotkey.measure.distance", 1, []any{999}},
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
			if got := hotkeyText(lang, tc.key, tc.args...); strings.Contains(got, "%!") {
				t.Fatalf("語系 %v %q 格式化失敗：%q", lang, tc.key, got)
			}
		}
	}
}

func TestHotkeyOverlayTextRectsStayInsideGalaxyViewport(t *testing.T) {
	inside := func(name string, r textSafeRect) {
		t.Helper()
		if r.x < starVX0 || r.y < starVY0 || r.x+r.w > starVX1+1 || r.y+r.h > starVY1+1 {
			t.Fatalf("%s 安全框 %+v 超出星圖 viewport (%d,%d)-(%d,%d)",
				name, r, starVX0, starVY0, starVX1, starVY1)
		}
	}
	for _, p := range [][2]int{{starVX0, starVY0}, {starVX1, starVY0}, {starVX0, starVY1}, {starVX1, starVY1}, {320, 240}} {
		inside("游標提示", measureHintTextRect(p[0], p[1]))
	}
	for _, line := range [][4]int{
		{starVX0, starVY0, starVX1, starVY1},
		{starVX1, starVY0, starVX1, starVY1},
		{starVX0, starVY1, starVX1, starVY1},
	} {
		inside("距離", measureDistanceTextRect(line[0], line[1], line[2], line[3]))
	}
	flash := quickSaveFlashTextRect()
	inside("快速存檔", flash)
	if flash.x <= 238 {
		t.Fatalf("快速存檔框 %+v 侵入左側選星資訊面板", flash)
	}
}

func TestHotkeyOverlayLongestTextIsWidthBounded(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		values := []struct {
			r    textSafeRect
			text string
		}{
			{measureHintTextRect(starVX1, starVY1), hotkeyText(lang, "hotkey.measure.select_origin")},
			{measureHintTextRect(starVX0, starVY0), hotkeyText(lang, "hotkey.measure.hover_target")},
			{measureDistanceTextRect(starVX0, starVY0, starVX1, starVY1), hotkeyText(lang, "hotkey.measure.distance", 999999)},
			{quickSaveFlashTextRect(), hotkeyText(lang, "hotkey.quicksave.failed", strings.Repeat("permission denied 銀河存檔路徑", 20))},
		}
		for _, value := range values {
			clipped := value.r.clipped(fnt, value.text, 11)
			if w, _ := fnt.Measure(clipped, 11); w > value.r.contentWidth() {
				t.Fatalf("語系 %v 截斷後仍超寬：%.0f > %.0f，%q", lang, w, value.r.contentWidth(), clipped)
			}
		}
	}
}

func TestHotkeyPlayerTextIsNotEmbeddedInSource(t *testing.T) {
	source, err := os.ReadFile("hotkeys.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, forbidden := range []string{
		".tr(", "沒有存檔位置可用", "快速存檔失敗", "測距:點選第一顆星", "%d 秒差距",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("hotkeys.go 仍內嵌玩家文案 %q", forbidden)
		}
	}
}
