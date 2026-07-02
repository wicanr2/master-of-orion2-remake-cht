package uifont

import (
	"strings"
	"testing"
)

// fakeMeasure 是注入式假量測:CJK rune 寬 2,其他 rune 寬 1,逐 rune 累加。
// 不需要真字型即可驗證 WrapText 的折行邏輯。
func fakeMeasure(s string) float64 {
	var w float64
	for _, r := range s {
		if isCJK(r) {
			w += 2
		} else {
			w += 1
		}
	}
	return w
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxWidth float64
		want     []string
	}{
		{
			name:     "純CJK依寬度斷行",
			s:        "中文測試字串",
			maxWidth: 6, // 每個 CJK 寬 2 → 每行 3 字
			want:     []string{"中文測", "試字串"},
		},
		{
			name:     "CJK剛好整除",
			s:        "銀河霸主",
			maxWidth: 4, // 每行 2 字
			want:     []string{"銀河", "霸主"},
		},
		{
			name:     "含換行符的多段_段數正確_空段保留",
			s:        "第一段\n\n第三段",
			maxWidth: 100,
			want:     []string{"第一段", "", "第三段"},
		},
		{
			name:     "純換行符無其他內容",
			s:        "\n\n",
			maxWidth: 10,
			want:     []string{"", "", ""},
		},
		{
			name:     "拉丁句子在空白斷_不切斷單字",
			s:        "hello world foo",
			maxWidth: 5, // "hello"=5, "world"=5, "foo"=3
			want:     []string{"hello", "world", "foo"},
		},
		{
			name:     "拉丁句子空白斷_多字塞同一行",
			s:        "go go go",
			maxWidth: 5, // "go go"=5 (g,o,space,g,o), 下一個 " go" 放不下
			want:     []string{"go go", "go"},
		},
		{
			name:     "超長單一token硬切",
			s:        "abcdefgh",
			maxWidth: 3,
			want:     []string{"abc", "def", "gh"},
		},
		{
			name:     "超長token夾在句子中間仍硬切",
			s:        "go abcdefgh go",
			maxWidth: 3,
			want:     []string{"go", "abc", "def", "gh", "go"},
		},
		{
			name:     "maxWidth小於等於0只依換行符分段",
			s:        "這是一段很長很長很長不會被折行的文字 with spaces too",
			maxWidth: 0,
			want:     []string{"這是一段很長很長很長不會被折行的文字 with spaces too"},
		},
		{
			name:     "maxWidth為負值只依換行符分段",
			s:        "第一行\n第二行",
			maxWidth: -1,
			want:     []string{"第一行", "第二行"},
		},
		{
			name:     "CJK與拉丁混合",
			s:        "Go語言真好用",
			maxWidth: 6, // G=1 o=1 語=2 言=2 => "Go語言"=6 ; 真=2 好=2 用=2 => "真好用"=6
			want:     []string{"Go語言", "真好用"},
		},
		{
			name:     "空字串",
			s:        "",
			maxWidth: 10,
			want:     []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapText(fakeMeasure, tt.s, tt.maxWidth)
			if len(got) != len(tt.want) {
				t.Fatalf("行數不符: got=%q want=%q", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("第 %d 行不符: got=%q want=%q (全部 got=%q)", i, got[i], tt.want[i], got)
				}
			}
		})
	}
}

// TestWrapText_LineWidthWithinLimit 驗證一般情況下每行實際量測寬度不超過 maxWidth。
// 超長單一 token 被硬切時,若單一 rune 本身寬度就超過 maxWidth(此測試未觸發,因
// fakeMeasure 最大單字寬為 2 <= 所用 maxWidth),硬切後仍應每片 <= maxWidth。
func TestWrapText_LineWidthWithinLimit(t *testing.T) {
	cases := []struct {
		s        string
		maxWidth float64
	}{
		{"中文測試字串連續文字用來驗證寬度上限", 8},
		{"the quick brown fox jumps over the lazy dog", 12},
		{"混合 mixed 文字 text 一起 together 測試", 10},
		{"supercalifragilisticexpialidocious", 4}, // 超長單一 token,強制硬切
	}
	for _, c := range cases {
		lines := WrapText(fakeMeasure, c.s, c.maxWidth)
		for i, ln := range lines {
			w := fakeMeasure(ln)
			if w > c.maxWidth {
				t.Errorf("s=%q maxWidth=%v 第 %d 行超寬: line=%q width=%v", c.s, c.maxWidth, i, ln, w)
			}
			if strings.HasPrefix(ln, " ") || strings.HasPrefix(ln, "\t") {
				t.Errorf("s=%q 第 %d 行行首殘留空白: line=%q", c.s, i, ln)
			}
		}
	}
}

// TestWrapText_NoWordSplit 驗證非 CJK 單字在有足夠空間換行的情況下不會被從中間切開。
func TestWrapText_NoWordSplit(t *testing.T) {
	s := "hello world foo bar"
	lines := WrapText(fakeMeasure, s, 5)
	words := map[string]bool{"hello": true, "world": true, "foo": true, "bar": true}
	for _, ln := range lines {
		for _, tok := range strings.Fields(ln) {
			if !words[tok] {
				t.Errorf("出現非預期(可能被切開)的 token: %q (line=%q)", tok, ln)
			}
		}
	}
}
