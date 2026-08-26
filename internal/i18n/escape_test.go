package i18n

import (
	"strings"
	"testing"
)

func TestEscapeDecode(t *testing.T) {
	// 0x8f 帝國名標記由 JSON 的 Unicode 跳脫表示。
	c := New(Traditional)
	c.LoadJSON(strings.NewReader(`[{"key":"A\\x8fB","value":"甲\\x8f乙"}]`))
	key := "A\x8fB"
	if got := c.Translate(key); got != "甲\x8f乙" {
		t.Errorf("含 \\x8f 的 key 未正確匹配/翻譯,得 %q", got)
	}
}
