package i18n

import (
	"strings"
	"testing"
)

func TestEscapeDecode(t *testing.T) {
	// \x8f 帝國名標記 + \n
	c := New(Traditional)
	c.LoadTSV(strings.NewReader("A\\x8fB\t甲\\x8f乙\n"))
	key := "A\x8fB"
	if got := c.Translate(key); got != "甲\x8f乙" {
		t.Errorf("含 \\x8f 的 key 未正確匹配/翻譯,得 %q", got)
	}
}
