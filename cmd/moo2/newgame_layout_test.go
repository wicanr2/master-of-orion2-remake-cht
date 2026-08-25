package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestNewGameSelectorTextUsesCenteredSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	// 這些是由完整 NEWGAME 背景圖量出的可視數值列，不是由
	// ngStripRect 自己推導出的期望值；測試因此能抓到座標系重複加原點。
	wantStrips := map[string][4]int{
		"diff":    {105, 204, 100, 20},
		"size":    {261, 204, 100, 20},
		"age":     {417, 204, 100, 20},
		"players": {105, 349, 100, 20},
		"tech":    {261, 349, 100, 20},
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		b := &sceneBuilder{lang: lang, newGameDiff: newGameDiffDefault, newGameSize: 1,
			newGameAge: newGameAgeDefault, newGameEmpires: shell.DefaultOpponents + 1, newGameTech: newGameTechDefault}
		for _, st := range ngSettings {
			r := ngStripTextRect(st)
			x, y, w, h := ngStripRect(st)
			want, ok := wantStrips[st.act]
			if !ok || [4]int{x, y, w, h} != want {
				t.Fatalf("%s 可視數值列 = (%d,%d,%d,%d)，want 量測矩形 (%d,%d,%d,%d)",
					st.act, x, y, w, h, want[0], want[1], want[2], want[3])
			}
			if r.x < x || r.y < y || r.x+r.w > x+w || r.y+r.h > y+h {
				t.Fatalf("%s 安全框 (%d,%d,%d,%d) 超出選擇器 (%d,%d,%d,%d)", st.act, r.x, r.y, r.w, r.h, x, y, w, h)
			}
			if r.x+r.w/2 != x+w/2 || r.y+r.h/2 != y+h/2 {
				t.Fatalf("%s 文字中心沒有與選擇器對齊", st.act)
			}
			for i := 0; i < st.n(b); i++ {
				st.set(b, i)
				label := r.clipped(fnt, st.label(b), 12)
				if got, _ := fnt.Measure(label, 12); got > r.contentWidth() {
					t.Fatalf("%s 的 %q 寬 %.0f 超出 %.0f", st.act, label, got, r.contentWidth())
				}
				if _, got := fnt.Measure(label, 12); got > float64(r.h-2*r.insetY) {
					t.Fatalf("%s 的 %q 高度 %.0f 超出 %.0f", st.act, label, got, float64(r.h-2*r.insetY))
				}
			}
		}
	}
}

func TestNewGameSetupRoutesLabelsThroughBoundedDrawer(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "interactive.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var setup *ast.FuncDecl
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == "newGameSetup" {
			setup = fn
			break
		}
	}
	if setup == nil {
		t.Fatal("找不到 newGameSetup")
	}
	seenBounded := false
	ast.Inspect(setup.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fn := call.Fun.(type) {
		case *ast.Ident:
			if fn.Name == "drawNewGameSettingLabel" {
				seenBounded = true
			}
		case *ast.SelectorExpr:
			if fn.Sel.Name == "drawCentered" || fn.Sel.Name == "DrawCentered" {
				t.Errorf("newGameSetup 不得繞過雙軸安全繪製器，直接呼叫 %s", fn.Sel.Name)
			}
		}
		return true
	})
	if !seenBounded {
		t.Fatal("newGameSetup 沒有使用 drawNewGameSettingLabel")
	}
}
