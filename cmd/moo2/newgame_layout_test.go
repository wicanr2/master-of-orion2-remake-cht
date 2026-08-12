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
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		b := &sceneBuilder{lang: lang, newGameDiff: newGameDiffDefault, newGameSize: 1,
			newGameAge: newGameAgeDefault, newGameEmpires: shell.DefaultOpponents + 1, newGameTech: newGameTechDefault}
		for _, st := range ngSettings {
			r := ngStripTextRect(st)
			x, y, w, h := ngStripRect(st)
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
