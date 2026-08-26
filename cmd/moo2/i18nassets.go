package main

import (
	"io/fs"
	"os"
	"path/filepath"
)

// i18nFS 回傳譯表的來源。
//
// 有 -i18n 指定目錄就讀該目錄；否則依序尋找執行檔旁與目前目錄下的 assets/i18n。
// 玩家文案刻意保持為外部 JSON，不使用 go:embed 烘入程式。
// 回傳的 fs.FS 與 dir 直接餵給 i18n.Registry.LoadFS。
func i18nFS(override string) (fs.FS, string) {
	if override != "" {
		return os.DirFS(override), "."
	}
	if exe, err := os.Executable(); err == nil {
		for _, dir := range []string{
			filepath.Join(filepath.Dir(exe), "assets", "i18n"),
			filepath.Join(filepath.Dir(exe), "i18n"),
		} {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return os.DirFS(dir), "."
			}
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		for i := 0; i < 4; i++ {
			dir := filepath.Join(cwd, "assets", "i18n")
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return os.DirFS(dir), "."
			}
			parent := filepath.Dir(cwd)
			if parent == cwd {
				break
			}
			cwd = parent
		}
	}
	return os.DirFS(filepath.Join("assets", "i18n")), "."
}

// readFromFS 是 fs.ReadFile 的薄包裝(測試用,讓 embed.FS 與 os.DirFS 走同一條路)。
func readFromFS(fsys fs.FS, name string) ([]byte, error) { return fs.ReadFile(fsys, name) }

// i18nOverrideDir 是 -i18n 旗標的值,由 main 設定。空 = 用烘進去的那份。
//
// 用套件級變數而不是把它一路傳下去:讀單一 .json 的地方散在七個檔案裡
// (overlay/researchchoice/topicname/eventscreen/colonyscreen/hiscore/main),
// 全部改簽名的成本遠高於它換來的東西,而這個值在整個行程裡只設定一次。
var i18nOverrideDir string

// OpenI18NJSON 開一份譯表（如 "tech.json"）。
//
// ⚠ **不要再寫 `os.Open("assets/i18n/xxx.json")`。** 那是相對於當前工作目錄的路徑,
// 執行檔只有從 repo 根目錄跑才找得到——`i18n_path_test.go` 會擋下新增的這種寫法。
func OpenI18NJSON(name string) (fs.File, error) {
	if i18nOverrideDir != "" {
		return os.DirFS(i18nOverrideDir).Open(name)
	}
	fsys, dir := i18nFS("")
	return fsys.Open(filepath.Join(dir, name))
}
