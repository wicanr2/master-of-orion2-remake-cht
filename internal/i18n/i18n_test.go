package i18n

import (
	"fmt"
	"strings"
	"testing"
)

const sampleJSON = `[
  {"key":"Continue","value":"繼續","note":"menu"},
  {"key":"Load Game","value":"載入遊戲"},
  {"key":"New Game","value":"新遊戲"},
  {"key":"Quit Game","value":"結束遊戲"},
  {"key":"Cost %v","value":"花費 %v","note":"模板"},
  {"key":"Empty","value":""}
]`

func newZH(t *testing.T) *Catalog {
	t.Helper()
	c := New(Traditional)
	n, err := c.LoadJSON(strings.NewReader(sampleJSON))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 { // 6 行有效但 Empty 的中文空 → 5
		t.Fatalf("載入條目 = %d,預期 5", n)
	}
	return c
}

func TestTranslate(t *testing.T) {
	c := newZH(t)
	if got := c.Translate("Continue"); got != "繼續" {
		t.Errorf("Continue → %q,預期 繼續", got)
	}
	if got := c.Translate("  New Game  "); got != "新遊戲" { // TrimSpace 查找
		t.Errorf("含空白 New Game → %q", got)
	}
	if got := c.Translate("Unknown"); got != "Unknown" { // 查無回原字串
		t.Errorf("未知字串應回原文,得 %q", got)
	}
	if got := c.Translate("Empty"); got != "Empty" { // 空中文欄未覆蓋
		t.Errorf("空中文欄應回原文,得 %q", got)
	}
}

func TestEnglishMode(t *testing.T) {
	c := newZH(t)
	c.SetLang(English)
	if got := c.Translate("Continue"); got != "Continue" {
		t.Errorf("英文模式應回原文,得 %q", got)
	}
}

func TestTranslateFormat(t *testing.T) {
	c := newZH(t)
	tmpl := c.TranslateFormat("Cost %v")
	if tmpl != "花費 %v" {
		t.Fatalf("模板翻譯 = %q,預期 花費 %%v", tmpl)
	}
	if got := fmt.Sprintf(tmpl, 42); got != "花費 42" {
		t.Errorf("填值後 = %q,預期 花費 42", got)
	}
}

func TestFirstWins(t *testing.T) {
	c := New(Traditional)
	c.LoadJSON(strings.NewReader(`[{"key":"Power","value":"能量"}]`))
	c.LoadJSON(strings.NewReader(`[{"key":"Power","value":"電力"}]`)) // 後載入應被略過
	if got := c.Translate("Power"); got != "能量" {
		t.Errorf("先載入者優先,得 %q,預期 能量", got)
	}
}

// TestLoadJSONTrimsKey 回歸：含尾端／開頭換行的 key 須能被查詢命中。
func TestLoadJSONTrimsDecodedKey(t *testing.T) {
	data := `[
  {"key":"Reduced to reduced intensity.\n","value":"zh尾端換行"},
  {"key":"\n\nThis colony does not allow farming.","value":"zh開頭換行"}
]`
	c := New(Traditional)
	if _, err := c.LoadJSON(strings.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ query, want string }{
		{"Reduced to reduced intensity.\n", "zh尾端換行"}, // 遊戲原字串帶尾端 \n
		{"Reduced to reduced intensity.", "zh尾端換行"},   // trim 後
		{"\n\nThis colony does not allow farming.", "zh開頭換行"},
		{"This colony does not allow farming.", "zh開頭換行"},
	}
	for _, tc := range cases {
		if got := c.Translate(tc.query); got != tc.want {
			t.Errorf("Translate(%q) = %q,預期 %q", tc.query, got, tc.want)
		}
	}
}
