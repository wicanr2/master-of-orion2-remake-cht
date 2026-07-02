package uifont

import (
	"strings"
	"unicode"
)

// isCJK 判斷 rune 是否屬於「可任意字間斷行」的 CJK 範圍:
// CJK 部首補充～統一表意文字、CJK 標點、全形、諺文。
func isCJK(r rune) bool {
	switch {
	case r >= 0x2E80 && r <= 0x9FFF:
		return true
	case r >= 0x3000 && r <= 0x303F:
		return true
	case r >= 0xFF00 && r <= 0xFFEF:
		return true
	case r >= 0xAC00 && r <= 0xD7A3:
		return true
	}
	return false
}

type wrapTokenKind int

const (
	tokenWord wrapTokenKind = iota
	tokenSpace
	tokenCJK
)

type wrapToken struct {
	kind wrapTokenKind
	text string
}

// tokenizeForWrap 把一段(不含 '\n')文字切成折行用的最小單位:
//   - 每個 CJK 字元自成一個 token(可任意斷)
//   - 連續空白自成一個 token(斷行點,行首不保留)
//   - 連續非空白非 CJK 字元自成一個 token(視為單詞,盡量不從中間切)
func tokenizeForWrap(s string) []wrapToken {
	runes := []rune(s)
	var toks []wrapToken
	i := 0
	for i < len(runes) {
		r := runes[i]
		switch {
		case isCJK(r):
			toks = append(toks, wrapToken{tokenCJK, string(r)})
			i++
		case unicode.IsSpace(r):
			j := i + 1
			for j < len(runes) && unicode.IsSpace(runes[j]) {
				j++
			}
			toks = append(toks, wrapToken{tokenSpace, string(runes[i:j])})
			i = j
		default:
			j := i + 1
			for j < len(runes) && !isCJK(runes[j]) && !unicode.IsSpace(runes[j]) {
				j++
			}
			toks = append(toks, wrapToken{tokenWord, string(runes[i:j])})
			i = j
		}
	}
	return toks
}

// hardSplitByWidth 把一段沒有天然斷點的文字(超長 word token)依 maxWidth 逐字硬切,
// 每片盡量塞滿但不超過 maxWidth;單一 rune 本身就超寬時仍單獨成片(無法再切)。
func hardSplitByWidth(measure func(string) float64, s string, maxWidth float64) []string {
	runes := []rune(s)
	var pieces []string
	i := 0
	for i < len(runes) {
		j := i + 1
		for j < len(runes) && measure(string(runes[i:j+1])) <= maxWidth {
			j++
		}
		pieces = append(pieces, string(runes[i:j]))
		i = j
	}
	if len(pieces) == 0 {
		pieces = append(pieces, "")
	}
	return pieces
}

// wrapSegment 把一段不含 '\n' 的文字依 maxWidth(> 0)折成多行。
func wrapSegment(measure func(string) float64, s string, maxWidth float64) []string {
	toks := tokenizeForWrap(s)
	var lines []string
	cur := ""

	// placeOversized 把在「空行」上仍放不下的 token 拆成 (完整行..., 剩餘 cur)。
	placeOversized := func(tok wrapToken) {
		switch tok.kind {
		case tokenWord:
			pieces := hardSplitByWidth(measure, tok.text, maxWidth)
			for i := 0; i < len(pieces)-1; i++ {
				lines = append(lines, pieces[i])
			}
			cur = pieces[len(pieces)-1]
		default: // tokenCJK:單一字元已是最小單位,放不下也只能單獨成行
			cur = tok.text
		}
	}

	for _, tok := range toks {
		if cur == "" && tok.kind == tokenSpace {
			// 行首不保留造成換行的空白
			continue
		}
		candidate := cur + tok.text
		if measure(candidate) <= maxWidth {
			cur = candidate
			continue
		}
		if cur == "" {
			placeOversized(tok)
			continue
		}
		// 目前這行放不下新 token:先把現有內容收成一行(去掉尾端空白)
		lines = append(lines, strings.TrimRight(cur, " \t"))
		if tok.kind == tokenSpace {
			cur = ""
			continue
		}
		if measure(tok.text) <= maxWidth {
			cur = tok.text
		} else {
			cur = ""
			placeOversized(tok)
		}
	}
	if cur != "" {
		lines = append(lines, strings.TrimRight(cur, " \t"))
	}
	if len(lines) == 0 {
		lines = append(lines, "")
	}
	return lines
}

// WrapText 依像素寬度把 s 折成多行。measure(str) 回傳 str 在目標字級下的像素寬。
// 規則:
//  1. 先依明確換行符 '\n' 分段,每段各自折行(空段保留為一個空字串,維持段距)。
//  2. CJK 字元(中日韓,見 isCJK)可在任意「字與字之間」斷行。
//  3. 非 CJK(拉丁字母/數字)盡量在空白處斷,避免把一個英文單字從中間切開;
//     但若單一 token 本身就超過 maxWidth,才允許硬切。
//  4. 每行盡量塞滿但不超過 maxWidth;行首不保留造成換行的空白。
//
// maxWidth <= 0 時不折行(僅依 '\n' 分段)。
func WrapText(measure func(string) float64, s string, maxWidth float64) []string {
	segs := strings.Split(s, "\n")
	out := make([]string, 0, len(segs))
	for _, seg := range segs {
		if maxWidth <= 0 {
			out = append(out, seg)
			continue
		}
		out = append(out, wrapSegment(measure, seg, maxWidth)...)
	}
	return out
}

// Wrap 是便利方法:用本字型在 size 字級下的實際量測折行。
func (f *Font) Wrap(s string, size, maxWidth float64) []string {
	return WrapText(func(str string) float64 {
		w, _ := f.Measure(str, size)
		return w
	}, s, maxWidth)
}
