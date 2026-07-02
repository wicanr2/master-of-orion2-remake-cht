// Package assets 依優先序在多個資料夾中解析遊戲資產檔名。
//
// 原版 patch 1.31 是把替換用的 .lbx 檔覆蓋進遊戲目錄。本層以「搜尋路徑」模型重現:
// 給定有序目錄清單(高優先在前,如 [patch1.31, 基礎安裝]),解析檔名時回傳第一個
// 命中的目錄中的檔案 → 後載覆蓋先載。之後 1.5 版同理疊更多層。
//
// DOS 檔名慣為大寫(GAME.LBX),玩家目錄可能大小寫不一,故比對大小寫不敏感。
package assets

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
)

// Resolver 以有序目錄清單解析資產檔名(dirs[0] 優先權最高)。
type Resolver struct {
	dirs []string
	// 每個目錄的「小寫檔名 → 實際檔名」快取,支援大小寫不敏感比對。
	listings []map[string]string
}

// NewResolver 建立解析器,dirs 依優先序排列(高優先在前)。
func NewResolver(dirs ...string) (*Resolver, error) {
	r := &Resolver{dirs: append([]string(nil), dirs...)}
	r.listings = make([]map[string]string, len(dirs))
	for i, d := range dirs {
		entries, err := os.ReadDir(d)
		if err != nil {
			return nil, fmt.Errorf("assets: 讀取目錄 %q 失敗: %w", d, err)
		}
		m := make(map[string]string, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			m[strings.ToLower(e.Name())] = e.Name()
		}
		r.listings[i] = m
	}
	return r, nil
}

// Path 依優先序找出檔名對應的實際路徑(大小寫不敏感)。found 為 false 表示各層皆無。
func (r *Resolver) Path(name string) (path string, found bool) {
	key := strings.ToLower(name)
	for i, dir := range r.dirs {
		if actual, ok := r.listings[i][key]; ok {
			return filepath.Join(dir, actual), true
		}
	}
	return "", false
}

// Read 讀出解析到的檔案內容。
func (r *Resolver) Read(name string) ([]byte, error) {
	p, ok := r.Path(name)
	if !ok {
		return nil, fmt.Errorf("assets: 找不到 %q(搜尋 %d 個目錄)", name, len(r.dirs))
	}
	return os.ReadFile(p)
}

// OpenLBX 解析並開啟一個 .lbx 資產封存檔。整檔讀進記憶體(.lbx 最大約數 MB),
// 以 bytes.Reader 當 ReaderAt,不留檔案 handle。
func (r *Resolver) OpenLBX(name string) (*lbx.Archive, error) {
	data, err := r.Read(name)
	if err != nil {
		return nil, err
	}
	arch, err := lbx.Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("assets: 開啟 %q: %w", name, err)
	}
	return arch, nil
}
