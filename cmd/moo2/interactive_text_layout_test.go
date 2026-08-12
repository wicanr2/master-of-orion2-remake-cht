package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestStarPanelTextSafeRectsStayInsidePanelAndButtons(t *testing.T) {
	panel := [4]int{starPanelX, starPanelY, starPanelW, starPanelH}
	inside := func(name string, r textSafeRect, outer [4]int) {
		t.Helper()
		if r.x < outer[0] || r.y < outer[1] || r.x+r.w > outer[0]+outer[2] || r.y+r.h > outer[1]+outer[3] {
			t.Fatalf("%s = (%d,%d,%d,%d) 超出外框 (%d,%d,%d,%d)", name, r.x, r.y, r.w, r.h,
				outer[0], outer[1], outer[2], outer[3])
		}
		if r.insetX < 0 || r.insetY < 0 || 2*r.insetX > r.w || 2*r.insetY > r.h {
			t.Fatalf("%s 內縮無效：%+v", name, r)
		}
	}

	inside("name", starPanelNameTextRect(), panel)
	inside("special", starPanelSpecialTextRect(), panel)
	inside("close", starPanelCloseTextRect(), panel)
	inside("environment", starPanelEnvironmentTextRect(353), panel)
	inside("gravity", starPanelEnvironmentTextRect(369), panel)
	inside("marine", starPanelMarineTextRect(), panel)
	if got := starPanelMarineTextRect().y + starPanelMarineTextRect().h; got > 402 {
		t.Fatalf("marine 行底端 y=%d，侵入第一顆按鈕", got)
	}
	nameRect, specialRect := starPanelNameTextRect(), starPanelSpecialTextRect()
	if nameRect.y != specialRect.y || nameRect.y+nameRect.h != specialRect.y+specialRect.h || nameRect.x+nameRect.w > specialRect.x {
		t.Fatalf("標題左右分欄未對齊或互相重疊：name=%+v special=%+v", nameRect, specialRect)
	}
	if got := starPanelSpecialTextRect().y + starPanelSpecialTextRect().h; got > starPanelEnvironmentTextRect(353).y {
		t.Fatalf("標題列底端 y=%d，侵入環境列", got)
	}
	for _, y := range []int{402, 424, 446} {
		inside("button", starPanelButtonTextRect(y), [4]int{starPanelButtonX, y, starPanelButtonW, 20})
	}
}

func TestStarPanelTextSafeRectsBoundLongBilingualData(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	tests := []struct {
		name string
		r    textSafeRect
		size float64
		text string
	}{
		{"長繁中星名", starPanelNameTextRect(), 14, "這是一個非常非常長的殖民星系名稱不應穿過右上關閉按鈕"},
		{"英文怪獸名稱", starPanelSpecialTextRect(), 10, "☠Ancient Guardian Monster Fleet"},
		{"長陸戰隊狀態", starPanelMarineTextRect(), 11, "艦隊陸戰隊 999／殖民地駐軍 999／艦員 Veteran(再 999 經驗升級)"},
		{"玩家殖民地名稱", starPanelButtonTextRect(424), 12, "▶ Rally: A Very Long Player Colony Name With Many Letters"},
		{"玩家前哨站名稱", starPanelButtonTextRect(402), 12, "▶ 建立前哨站：這是玩家輸入的超長前哨站名稱"},
	}
	for _, tc := range tests {
		lines := tc.r.lines(fnt, tc.text, tc.size)
		if len(lines) > tc.r.maxLines() {
			t.Fatalf("%s 行數 %d 超過上限 %d", tc.name, len(lines), tc.r.maxLines())
		}
		for _, line := range lines {
			w, h := fnt.Measure(line, tc.size)
			if w > tc.r.contentWidth() {
				t.Fatalf("%s 寬 %.0f 超出安全寬 %.0f：%q", tc.name, w, tc.r.contentWidth(), line)
			}
			if h > float64(tc.r.h-2*tc.r.insetY) {
				t.Fatalf("%s 字高 %.0f 超出安全高 %d：%q", tc.name, h, tc.r.h-2*tc.r.insetY, line)
			}
		}
	}
}

func TestInteractiveStarPanelCannotBypassTextSafeRects(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("無法定位測試原始檔")
	}
	sourcePath := strings.TrimSuffix(testFile, "_text_layout_test.go") + ".go"
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("讀取 interactive.go：%v", err)
	}
	start := strings.Index(string(source), "// starPanelTextLayoutStart")
	end := strings.Index(string(source), "// starPanelTextLayoutEnd")
	if start < 0 || end <= start {
		t.Fatal("找不到星圖選星文字版面守門標記")
	}
	startLine := 1 + strings.Count(string(source[:start]), "\n")
	endLine := 1 + strings.Count(string(source[:end]), "\n")

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, sourcePath, source, 0)
	if err != nil {
		t.Fatalf("解析 interactive.go：%v", err)
	}
	var bypasses []token.Position
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		pos := fset.Position(call.Pos())
		if pos.Line < startLine || pos.Line >= endLine {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && (sel.Sel.Name == "Draw" || sel.Sel.Name == "DrawCentered") {
			bypasses = append(bypasses, pos)
		}
		return true
	})
	if len(bypasses) != 0 {
		t.Fatalf("星圖選星面板不可直接繪字，未受 textSafeRect 管理的位置：%v", bypasses)
	}
}
