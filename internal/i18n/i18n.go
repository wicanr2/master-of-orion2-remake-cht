// Package i18n 提供顯示層字串覆蓋(display-layer override)在地化。
//
// 架構取自魔法大帝繁中化 playbook(docs/kickoff/08):
//   - **英文原文即 key**:譯表以英文原字串為鍵,查無則回原字串(英文模式零影響)。
//   - **只在顯示層翻譯,不動資料層**:遊戲邏輯常以英文字串當識別鍵,改資料層會壞邏輯。
//   - **TranslateFormat**:給 fmt.Sprintf 模板用 —— 先翻模板字面再填值
//     (填值後整串比對不會命中)。
//
// 譯表格式:TSV 三欄 `英文原文<TAB>中文<TAB>備註`(備註選填);空中文欄略過(選擇性覆蓋)。
package i18n

import (
	"bufio"
	"io"
	"strings"
)

// Lang 是目前語言。
type Lang int

const (
	English Lang = iota
	Traditional
)

// Catalog 是一份英文→中文對照表 + 目前語言狀態。
type Catalog struct {
	lang Lang
	m    map[string]string
}

// New 建立指定語言的空 Catalog。
func New(lang Lang) *Catalog {
	return &Catalog{lang: lang, m: make(map[string]string)}
}

// Lang 回傳目前語言。
func (c *Catalog) Lang() Lang { return c.lang }

// SetLang 切換語言(runtime 可切,對應主選單中/英切換)。
func (c *Catalog) SetLang(l Lang) { c.lang = l }

// LoadTSV 從 TSV 讀入譯文並併入 catalog。同一 key 以**先載入者優先**
// (後載入的重複 key 略過),對應 mom 以檔名字母序控制優先權的做法。
// 回傳新增的條目數。
func (c *Catalog) LoadTSV(r io.Reader) (int, error) {
	added := 0
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 2 {
			continue
		}
		key := strings.TrimSpace(cols[0])
		val := strings.TrimSpace(cols[1])
		if key == "" || val == "" {
			continue // 空中文 → 選擇性覆蓋,略過
		}
		if _, exists := c.m[key]; exists {
			continue // 先載入者優先
		}
		c.m[key] = val
		added++
	}
	return added, sc.Err()
}

// Translate 回傳字串的當前語言版本。英文模式或查無 → 回原字串(TrimSpace 後查找,
// 對齊引擎讀 LBX 時的 trim)。
func (c *Catalog) Translate(s string) string {
	if c.lang == English {
		return s
	}
	if v, ok := c.m[strings.TrimSpace(s)]; ok {
		return v
	}
	return s
}

// TranslateFormat 翻譯 fmt.Sprintf 的模板字面(如 "Cost %v" → "花費 %v")。
// 語意同 Translate,獨立命名以標示「這是模板、要在 Sprintf 前翻」。
// [注意] 佔位符(%v/%d/%s…)的數量與順序,中英譯文必須一致,否則 Sprintf 會出錯。
func (c *Catalog) TranslateFormat(tmpl string) string {
	return c.Translate(tmpl)
}

// Has 回傳是否有該 key 的譯文(供缺譯稽核用)。
func (c *Catalog) Has(s string) bool {
	_, ok := c.m[strings.TrimSpace(s)]
	return ok
}

// Size 回傳譯表條目數。
func (c *Catalog) Size() int { return len(c.m) }
