// Package i18n 提供顯示層字串覆蓋(display-layer override)在地化。
//
// 架構取自魔法大帝繁中化 playbook(docs/kickoff/08):
//   - **英文原文即 key**:譯表以英文原字串為鍵,查無則回原字串(英文模式零影響)。
//   - **只在顯示層翻譯,不動資料層**:遊戲邏輯常以英文字串當識別鍵,改資料層會壞邏輯。
//   - **TranslateFormat**:給 fmt.Sprintf 模板用 —— 先翻模板字面再填值
//     (填值後整串比對不會命中)。
//
// 譯表格式:JSON 陣列；每筆含 key、value 與可省略的 note。空 value 略過。
package i18n

import (
	"encoding/json"
	"io"
	"strings"
)

// Lang 是目前語言。
type Lang int

const (
	English Lang = iota
	Traditional
)

// Catalog 同時支援原版英文 key→繁中覆蓋，以及 remake 語意鍵→外部中英文文案。
type Catalog struct {
	lang    Lang
	m       map[string]string
	english map[string]string
}

// New 建立指定語言的空 Catalog。
func New(lang Lang) *Catalog {
	return &Catalog{lang: lang, m: make(map[string]string), english: make(map[string]string)}
}

// Lang 回傳目前語言。
func (c *Catalog) Lang() Lang { return c.lang }

// SetLang 切換語言(runtime 可切,對應主選單中/英切換)。
func (c *Catalog) SetLang(l Lang) { c.lang = l }

// Entry 是外部 JSON 譯表的一筆玩家可見文案。
type Entry struct {
	Key     string `json:"key"`
	English string `json:"english,omitempty"`
	Value   string `json:"value"`
	Note    string `json:"note,omitempty"`
}

// LoadJSON 從 JSON 讀入譯文並併入 catalog。同一 key 以先載入者優先
// （後載入的重複 key 略過）。JSON 本身負責換行、Tab 與控制碼的跳脫解碼。
// 回傳新增的條目數。
func (c *Catalog) LoadJSON(r io.Reader) (int, error) {
	var entries []Entry
	dec := json.NewDecoder(r)
	if err := dec.Decode(&entries); err != nil {
		return 0, err
	}
	added := 0
	for _, entry := range entries {
		key := strings.TrimSpace(decodeByteEscapes(entry.Key))
		val := decodeByteEscapes(entry.Value)
		english := decodeByteEscapes(entry.English)
		if key == "" {
			continue
		}
		if val != "" {
			if _, exists := c.m[key]; !exists {
				c.m[key] = val
				added++
			}
		}
		if english != "" {
			if _, exists := c.english[key]; !exists {
				c.english[key] = english
			}
		}
	}
	return added, nil
}

// Text 以穩定語意鍵查詢玩家文案。自繪 remake 畫面使用此入口，讓中英文句子都留在
// 外部 JSON；Traditional 優先 value，English 使用 english，缺譯時退回另一語言再退回鍵值。
// 原版英文字串作 key 的資料表仍使用 Translate，不改變既有位置式來源契約。
func (c *Catalog) Text(key string) string {
	return c.TextFor(c.lang, key)
}

// TextFor 與 Text 相同，但不修改 catalog 的目前語言，供共用惰性 catalog 的顯示端使用。
func (c *Catalog) TextFor(lang Lang, key string) string {
	key = strings.TrimSpace(key)
	if lang == English {
		if v, ok := c.english[key]; ok {
			return v
		}
		if v, ok := c.m[key]; ok {
			return v
		}
		return key
	}
	if v, ok := c.m[key]; ok {
		return v
	}
	if v, ok := c.english[key]; ok {
		return v
	}
	return key
}

// decodeByteEscapes 還原 JSON 文案中的原版單位元控制標記（例如 \x8f 帝國名）。
// JSON 的 \u008f 會變成 UTF-8 雙位元，不能表示原版 LBX 字串協定，因此資料檔明確保存 \xNN。
func decodeByteEscapes(s string) string {
	if !strings.Contains(s, `\x`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) && s[i+1] == 'x' {
			if v, ok := hexByte(s[i+2], s[i+3]); ok {
				b.WriteByte(v)
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func hexByte(a, b byte) (byte, bool) {
	hex := func(c byte) (byte, bool) {
		switch {
		case c >= '0' && c <= '9':
			return c - '0', true
		case c >= 'a' && c <= 'f':
			return c - 'a' + 10, true
		case c >= 'A' && c <= 'F':
			return c - 'A' + 10, true
		default:
			return 0, false
		}
	}
	hi, okHi := hex(a)
	lo, okLo := hex(b)
	return hi<<4 | lo, okHi && okLo
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
