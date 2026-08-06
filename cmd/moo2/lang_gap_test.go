package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// lang_gap_test.go:英文模式缺口的**棘輪**。
//
// 為什麼要有這支:拿 grep 數「含中文的字串字面值」量不出覆蓋率——`tr("繼續", "CONTINUE")` 本身
// (順帶一提,那個 grep 的字元類上界 U+9FFF 本身就是個沒有字形的碼點,寫進 .go 註解會被
//  internal/uifont 的缺字守門測試抓出來,所以這裡用文字描述而不貼 pattern)
// 就含中文字面值,補得越多 grep 的數字反而不動。這裡改用 go/ast 只數「**沒有**被 tr()
// 包住、又真的會畫到畫面上」的中文字面值,那才是英文模式下會露中文的量。
//
// 這支是棘輪不是驗收:它擋住往回長,不宣稱缺口已清零。真正的驗收是跑 `-lang en` 的
// 過場截圖廊逐張看(見 docs/HONEST-STATUS.md「英文模式覆蓋率」)。

// langGapCeiling 是目前允許的未包覆數量。**只能往下調,不能往上調。**
// 補完一批就把數字改小;要往上調代表英文模式退步了,先問為什麼。
//
// 2026-08-07 降到 16。**剩下這 16 條全是偵測器分不出來的東西,不是真缺口**:
//   - `shipClassZH` / `designHull` / `customrace` 的「政府型態」:那是**查表 key**,
//     不是顯示字(換成英文 shell 那邊直接查不到)。
//   - `colonyview` / `diploview` / `raceinfo` / `menu` / `planets` / `galaxy` 的
//     視窗標題與示範資料:那幾支是 dev-only 的單畫面檢視模式,不在主遊戲流程裡。
//   - `multiplayer` 的「熱座 %d 人」:英文模式在它之前就 `continue` 讓路了,跑不到。
//
// 換句話說,真正的 UI 缺口目前是 0——但別把這句當「英文模式做完了」:
// 引擎層(internal/)仍會回中文字串,那是另一輪,見 docs/HONEST-STATUS.md。
const langGapCeiling = 16

// 不算缺口的呼叫:場景名(只進 log)、印到 stderr 的訊息、error 值。
// 這些不會畫到畫面上,維持中文是對的。
var langGapExemptCalls = map[string]bool{
	"goTo": true, "Fprintf": true, "Fprintln": true, "Fprint": true,
	"Errorf": true, "New": true, "Printf": true, "Println": true, "Print": true,
	// 這兩個建構式的最後一個引數也是場景名(取消時回哪裡),只進 log。
	"newCutsceneScreen": true, "newLoadGameScreen": true, "backHit": true,
	// 視窗標題是給視窗管理員看的,不是遊戲畫面上的字。
	"SetWindowTitle": true,
}

// 整檔豁免。**每一筆都要寫理由**——這裡是唯一可以「不補而不算缺口」的出口,
// 不寫理由就等於把缺口藏起來。
var langGapExemptFiles = map[string]string{
	// -play 是已淘汰的簡約殼,不是主遊戲畫面;主遊戲走 -game。
	"play.go": "已淘汰的 -play 簡約殼",
	// 全是 CLI 旗標說明(`-h` 印在終端機)與視窗標題,不是遊戲畫面上的字。
	"main.go": "CLI 旗標說明與視窗標題",
}

func hasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// TestEnglishModeGapDoesNotGrow 掃描本套件所有非測試檔,數出沒被 tr() 包住的中文字面值。
func TestEnglishModeGapDoesNotGrow(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("列檔失敗:%v", err)
	}
	fset := token.NewFileSet()
	perFile := map[string]int{}
	spots := []string{} // file:line + 內容,給接手的人直接跳過去補
	total := 0

	// 先掃一遍全套件,收集有 `xxxEn` 姊妹宣告的變數名——`var cats` 配 `var catsEn`
	// 就代表那份清單的英文版已經備好(顯示端依語言換整份 slice),不算缺口。
	pairedVars := pairedEnglishVars(fset, files)

	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		if _, ok := langGapExemptFiles[path]; ok {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("讀 %s 失敗:%v", path, err)
		}
		// 用 ParseFile 而不是 grep:註解不會被算進來,字串邊界也不會誤判。
		f, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			t.Fatalf("parse %s 失敗:%v", path, err)
		}

		// 先收集「已被豁免」的字面值節點,再數剩下的。
		exempt := map[ast.Node]bool{}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := calleeName(call.Fun)
			if name == "tr" || langGapExemptCalls[name] {
				// 整棵引數子樹都算已處理——`tr("主題:"+x+"(…)", …)` 的字面值
				// 藏在 BinaryExpr 底下,只標最外層節點會漏掉。
				for _, a := range call.Args {
					ast.Inspect(a, func(m ast.Node) bool {
						exempt[m] = true
						return true
					})
				}
			}
			return true
		})

		// 資料表的配對慣例:同一個 composite literal 裡,中文字串旁邊擺著純 ASCII 字串
		// (`{"阿爾卡里", "Alkari", …}`、`{"人口成長", "Population Growth", …}`),
		// 代表英文欄已經填好、顯示端會挑。這種不算缺口。
		ast.Inspect(f, func(n ast.Node) bool {
			cl, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			var han []ast.Node
			hasASCII := false
			for _, el := range cl.Elts {
				lit, ok := el.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				v, err := strconv.Unquote(lit.Value)
				if err != nil {
					continue
				}
				if hasHan(v) {
					han = append(han, lit)
				} else if strings.TrimSpace(v) != "" {
					hasASCII = true
				}
			}
			if hasASCII {
				for _, h := range han {
					exempt[h] = true
				}
			}
			return true
		})

		// overlay 路徑的資料:`[]labelRect{…}` 是「擦掉原版烘的英文、疊中文」用的,
		// 英文模式整段被 overlayScreen 跳過(露原版英文),所以那些中文不是缺口。
		ast.Inspect(f, func(n ast.Node) bool {
			cl, ok := n.(*ast.CompositeLit)
			if !ok || compositeTypeName(cl.Type) != "labelRect" {
				return true
			}
			ast.Inspect(cl, func(m ast.Node) bool {
				exempt[m] = true
				return true
			})
			return true
		})

		// 語言分支:條件提到 `i18n.Traditional` 的 if,兩邊分支裡的字面值都算已處理——
		// 那是「讓路」那條路徑的寫法(英文模式跳過擦底疊字,露原版美術上烘的英文)。
		ast.Inspect(f, func(n ast.Node) bool {
			ifs, ok := n.(*ast.IfStmt)
			if !ok || ifs.Cond == nil {
				return true
			}
			var cond strings.Builder
			ast.Inspect(ifs.Cond, func(m ast.Node) bool {
				if id, ok := m.(*ast.Ident); ok {
					cond.WriteString(id.Name + ".")
				}
				return true
			})
			if !strings.Contains(cond.String(), "Traditional") {
				return true
			}
			for _, br := range []ast.Node{ifs.Body, ifs.Else} {
				if br == nil {
					continue
				}
				ast.Inspect(br, func(m ast.Node) bool {
					exempt[m] = true
					return true
				})
			}
			return true
		})

		// 有 `xxxEn` 姊妹宣告的變數:整份清單視為已備好英文版。
		// `var x = …` 與函式內的 `x := …` 都要涵蓋——infosubscreens 的 cats/hows 就是後者。
		markPaired := func(names []ast.Expr, values []ast.Expr) {
			for i, nm := range names {
				id, ok := nm.(*ast.Ident)
				if !ok || !pairedVars[id.Name] || i >= len(values) {
					continue
				}
				ast.Inspect(values[i], func(m ast.Node) bool {
					if lit, ok := m.(*ast.BasicLit); ok {
						exempt[lit] = true
					}
					return true
				})
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.ValueSpec:
				names := make([]ast.Expr, len(d.Names))
				for i, nm := range d.Names {
					names[i] = nm
				}
				markPaired(names, d.Values)
			case *ast.AssignStmt:
				markPaired(d.Lhs, d.Rhs)
			}
			return true
		})

		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || exempt[ast.Node(lit)] {
				return true
			}
			v, err := strconv.Unquote(lit.Value)
			if err != nil || !hasHan(v) {
				return true
			}
			perFile[path]++
			total++
			pos := fset.Position(lit.Pos())
			spots = append(spots, path+":"+strconv.Itoa(pos.Line)+"  "+v)
			return true
		})
	}

	if total > langGapCeiling {
		names := make([]string, 0, len(perFile))
		for k := range perFile {
			names = append(names, k)
		}
		sort.Slice(names, func(i, j int) bool { return perFile[names[i]] > perFile[names[j]] })
		var b strings.Builder
		for _, n := range names {
			b.WriteString("\n    " + n + ": " + strconv.Itoa(perFile[n]))
		}
		t.Errorf("英文模式缺口 %d 條,超過上限 %d(棘輪只能往下調)%s", total, langGapCeiling, b.String())
	}
	t.Logf("英文模式缺口:%d 條(上限 %d)", total, langGapCeiling)
	// -v 時列出實際位置(限量),省得接手的人自己再寫一支掃描器。
	if testing.Verbose() {
		sort.Strings(spots)
		for i, sp := range spots {
			if i >= 400 {
				t.Logf("  …另有 %d 處", len(spots)-i)
				break
			}
			t.Log("  " + sp)
		}
	}
}

// pairedEnglishVars 找出所有「有 `xxxEn` 姊妹宣告」的變數名(含區域變數)。
// 慣例:`cats` / `catsEn`、`infoTabTitles` / `infoTabTitlesEn`。
func pairedEnglishVars(fset *token.FileSet, files []string) map[string]bool {
	declared := map[string]bool{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.ValueSpec:
				for _, nm := range d.Names {
					declared[nm.Name] = true
				}
			case *ast.AssignStmt:
				for _, lhs := range d.Lhs {
					if id, ok := lhs.(*ast.Ident); ok {
						declared[id.Name] = true
					}
				}
			}
			return true
		})
	}
	paired := map[string]bool{}
	for name := range declared {
		if declared[name+"En"] {
			paired[name] = true
		}
	}
	return paired
}

// compositeTypeName 取 composite literal 的型別名:`[]labelRect{…}` → "labelRect"。
func compositeTypeName(t ast.Expr) string {
	switch e := t.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.ArrayType:
		return compositeTypeName(e.Elt)
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.StarExpr:
		return compositeTypeName(e.X)
	}
	return ""
}

// calleeName 取呼叫端最後一段名字:`b.tr(...)` → "tr"、`fmt.Fprintf(...)` → "Fprintf"。
func calleeName(fn ast.Expr) string {
	switch e := fn.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	}
	return ""
}
