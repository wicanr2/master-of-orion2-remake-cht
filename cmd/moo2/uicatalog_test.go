package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

// 固定 UI 鍵的雙語欄位必須真的存在；Catalog.TextFor 的 fallback 會把缺鍵
// 顯示成識別字，單看一般畫面測試不一定會發現。動態組合鍵由各畫面測試驗證。
func TestLiteralUITextKeysHaveBothLanguages(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "assets", "i18n", "ui.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entries []i18n.Entry
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatal(err)
	}
	byKey := make(map[string]i18n.Entry, len(entries))
	for _, entry := range entries {
		if _, exists := byKey[entry.Key]; exists {
			t.Errorf("ui.json 重複鍵 %q", entry.Key)
		}
		byKey[entry.Key] = entry
	}
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	checked := 0
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) < 2 {
				return true
			}
			fn, ok := call.Fun.(*ast.Ident)
			if !ok || fn.Name != "uiText" {
				return true
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true // 動態鍵由其使用畫面的測試驗證
			}
			key, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Errorf("%s: 無法解析 UI 鍵：%v", fset.Position(literal.Pos()), err)
				return true
			}
			checked++
			entry, exists := byKey[key]
			// 空白分隔符有組句語意，只有真正的空字串算缺譯。
			if !exists || entry.English == "" || entry.Value == "" {
				t.Errorf("%s: UI 鍵 %q 缺少中英雙語欄位", fset.Position(literal.Pos()), key)
			}
			return true
		})
	}
	if checked == 0 {
		t.Fatal("未檢查到任何固定 uiText 鍵")
	}
}

func TestNetInfoPlayerTextComesFromExternalJSON(t *testing.T) {
	cases := []struct {
		key, zh, en string
	}{
		{"netinfo.caption.waiting_for_joiners", "等待其他玩家加入", "Waiting for players to join"},
		{"netinfo.caption.joining", "加入對局中", "Joining game"},
		{"netinfo.caption.wait_race_info", "等待種族資料", "Waiting for race info"},
		{"netinfo.caption.initializing", "初始化連線", "Initializing network"},
		{"netinfo.caption.sending_data", "傳送資料", "Sending data"},
		{"netinfo.caption.generating_map", "產生星圖", "Generating map"},
		{"netinfo.caption.getting_data", "接收資料", "Getting data"},
		{"netinfo.title.waiting_for_joiners", "加入網路遊戲設定", "JOIN NETWORK GAME SETUP"},
		{"netinfo.title.joining", "等待加入遊戲", "WAITING TO JOIN GAME"},
		{"netinfo.title.wait_race_info", "接收種族設定", "RECEIVING RACE SETUPS"},
		{"netinfo.title.initializing", "初始化網路", "INITIALIZING NETWORK"},
		{"netinfo.title.sending_data", "傳送遊戲資料", "SENDING GAME DATA"},
		{"netinfo.title.generating_map", "產生星圖", "GENERATING MAP"},
		{"netinfo.title.getting_data", "接收遊戲資料", "RECEIVING GAME DATA"},
		{"netinfo.label.status", "狀態", "STATUS"},
		{"netinfo.button.start", "開始連線對局", "START NET GAME"},
	}
	for _, tc := range cases {
		if got := uiText(i18n.Traditional, tc.key); got != tc.zh {
			t.Errorf("uiText(zh,%q)=%q，預期 %q", tc.key, got, tc.zh)
		}
		if got := uiText(i18n.English, tc.key); got != tc.en {
			t.Errorf("uiText(en,%q)=%q，預期 %q", tc.key, got, tc.en)
		}
	}
}

func TestNetInfoSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("netinfo.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if strings.Contains(src, ".tr(") {
		t.Fatal("netinfo.go 不得再用 tr 內嵌中英文玩家文案")
	}
	for _, text := range []string{
		"等待其他玩家加入", "Waiting for players to join",
		"開始連線對局", "START NET GAME",
	} {
		if strings.Contains(src, `"`+text+`"`) {
			t.Errorf("netinfo.go 仍內嵌玩家文案 %q", text)
		}
	}
}
