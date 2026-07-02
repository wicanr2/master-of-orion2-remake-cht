package i18n

import (
	"io/fs"
	"path"
	"sort"
	"strings"
)

// Registry 持有多個具名 per-source catalog,忠於 MOO2 引擎「每字串表獨立、以 (表,id) 查詢」
// 的設計:同一英文字在不同來源(TECHNAME / ESTRINGS / RACESTUF…)可有不同譯文,不互相覆蓋。
// 詳見 docs/tech/i18n-catalog-architecture.md。
//
// 用法:
//
//	reg := i18n.NewRegistry(i18n.Traditional)
//	reg.LoadFS(os.DirFS("assets/i18n"), ".")   // 或 embed.FS
//	reg.Source("tech").Translate("Fusion Beam") // 用 tech 來源的譯文
//	reg.Translate("Colony")                      // merged 備援(不指定來源時)
type Registry struct {
	lang    Lang
	sources map[string]*Catalog
	merged  *Catalog // 各來源併入(檔名字母序、先載入者優先),供不指定來源時備援
}

// NewRegistry 建立指定語言的空 Registry。
func NewRegistry(lang Lang) *Registry {
	return &Registry{lang: lang, sources: map[string]*Catalog{}, merged: New(lang)}
}

// LoadFS 從 fsys 的 dir 目錄載入所有 *.tsv:每檔成為一個以「去 .tsv 副檔名的檔名」命名的
// 來源 catalog(如 tech.tsv → 來源 "tech"),並依檔名字母序併入 merged 備援表
// (先載入者優先,與單一 Catalog 的優先權規則一致)。回傳載入的來源數。
func (r *Registry) LoadFS(fsys fs.FS, dir string) (int, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return 0, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".tsv") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names) // 字母序決定 merged 的先載入優先權(可重現)
	for _, name := range names {
		f, err := fsys.Open(path.Join(dir, name))
		if err != nil {
			return len(r.sources), err
		}
		cat := New(r.lang)
		_, err = cat.LoadTSV(f)
		f.Close()
		if err != nil {
			return len(r.sources), err
		}
		src := strings.TrimSuffix(name, ".tsv")
		r.sources[src] = cat
		for k, v := range cat.m { // 併入 merged(先載入者優先)
			if _, ok := r.merged.m[k]; !ok {
				r.merged.m[k] = v
			}
		}
	}
	return len(names), nil
}

// Source 回傳指定來源的 catalog;不存在時回一個空 catalog(Translate 查無即回原字串,不 panic)。
func (r *Registry) Source(name string) *Catalog {
	if c, ok := r.sources[name]; ok {
		return c
	}
	return New(r.lang)
}

// Merged 回傳合併備援 catalog。
func (r *Registry) Merged() *Catalog { return r.merged }

// Translate 以 merged 備援表翻譯(不指定來源時用;可比較畫面請改用 Source(x).Translate)。
func (r *Registry) Translate(s string) string { return r.merged.Translate(s) }

// SetLang 切換語言並套用到所有來源與 merged 表(對應主選單中/英切換)。
func (r *Registry) SetLang(l Lang) {
	r.lang = l
	for _, c := range r.sources {
		c.SetLang(l)
	}
	r.merged.SetLang(l)
}

// Lang 回傳目前語言。
func (r *Registry) Lang() Lang { return r.lang }

// Sources 回傳已載入的來源名(字母序)。
func (r *Registry) Sources() []string {
	out := make([]string, 0, len(r.sources))
	for k := range r.sources {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Conflicts 回傳跨來源同 key 但譯文不同的鍵 → {來源: 譯文}。
// per-source 查詢下這不是 bug(見架構文件),但可供稽核監看新增的跨畫面不一致。
func (r *Registry) Conflicts() map[string]map[string]string {
	byKey := map[string]map[string]string{}
	for src, c := range r.sources {
		for k, v := range c.m {
			if byKey[k] == nil {
				byKey[k] = map[string]string{}
			}
			byKey[k][src] = v
		}
	}
	out := map[string]map[string]string{}
	for k, sv := range byKey {
		seen := map[string]bool{}
		for _, v := range sv {
			seen[v] = true
		}
		if len(seen) > 1 {
			out[k] = sv
		}
	}
	return out
}
