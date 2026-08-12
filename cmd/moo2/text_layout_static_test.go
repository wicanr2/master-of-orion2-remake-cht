package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"
)

func parseMoo2Source(t *testing.T, name string) (*token.FileSet, *ast.File) {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("無法定位測試原始檔")
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join(filepath.Dir(testFile), name), nil, 0)
	if err != nil {
		t.Fatalf("解析 %s：%v", name, err)
	}
	return fset, file
}

func TestInfoSubscreensCannotBypassTextSafeRects(t *testing.T) {
	fset, file := parseMoo2Source(t, "infosubscreens.go")
	var bypasses []token.Position
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		ident, ok := lit.Type.(*ast.Ident)
		if ok && ident.Name == "extraText" {
			bypasses = append(bypasses, fset.Position(lit.Pos()))
		}
		return true
	})
	if len(bypasses) != 0 {
		t.Fatalf("INFO 子頁不可直接建立 extraText，未受安全框管理的位置：%v", bypasses)
	}
}

func TestNetNextTurnCannotBypassTextSafeRects(t *testing.T) {
	fset, file := parseMoo2Source(t, "netnextturn.go")
	var bypasses []token.Position
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && (sel.Sel.Name == "Draw" || sel.Sel.Name == "DrawCentered") {
			bypasses = append(bypasses, fset.Position(call.Pos()))
		}
		return true
	})
	if len(bypasses) != 0 {
		t.Fatalf("多人等待畫面不可直接繪字，未受安全框管理的位置：%v", bypasses)
	}
}
