package main

import (
	"embed"
	"io/fs"
	"os"
)

// embedassets.go:把譯表烘進執行檔(第 83 項(打包路徑))。
//
// ============ 為什麼要做 ============
//
// 先前每個模式都寫著 `os.DirFS("assets/i18n")` —— **相對於當前工作目錄**。
// 也就是說執行檔只有從 repo 根目錄跑才找得到譯表;macOS `.app` 與 AppImage 都得靠
// launcher script 先 `cd` 過去。那不是打包問題,是**執行檔對外部檔案的隱性依賴**。
//
// ============ 為什麼只烘譯表,不烘字型與遊戲資料 ============
//
//   - **字型**是使用者自備的 TTC(授權不明),不能進 repo 也不能進執行檔。
//   - **遊戲資料**(LBX)是玩家的正版資產,同理。CLAUDE.md 的 [HARD] 規則。
//
// 譯表是本專案自己寫的,烘進去沒有授權問題,而且它正是「換一台機器就跑不起來」的那一份。
//
// ⚠ `go:embed` 讀不到 `../../assets`(embed 不能跨出套件目錄),所以譯表在
// `cmd/moo2/embedded/i18n/` 有一份副本。**兩份要同步**——`i18n_embed_test.go` 會比對,
// 不同步就會紅。

//go:embed embedded/i18n/*.tsv
var embeddedI18N embed.FS

// i18nFS 回傳譯表的來源。
//
// 有 -i18n 指定目錄就讀那個目錄(開發時改譯表不必重編);否則用烘進去的那份。
// 回傳的 fs.FS 與 dir 直接餵給 i18n.Registry.LoadFS。
func i18nFS(override string) (fs.FS, string) {
	if override != "" {
		return os.DirFS(override), "."
	}
	return embeddedI18N, "embedded/i18n"
}

// readFromFS 是 fs.ReadFile 的薄包裝(測試用,讓 embed.FS 與 os.DirFS 走同一條路)。
func readFromFS(fsys fs.FS, name string) ([]byte, error) { return fs.ReadFile(fsys, name) }

// i18nOverrideDir 是 -i18n 旗標的值,由 main 設定。空 = 用烘進去的那份。
//
// 用套件級變數而不是把它一路傳下去:讀單一 .tsv 的地方散在七個檔案裡
// (overlay/researchchoice/topicname/eventscreen/colonyscreen/hiscore/main),
// 全部改簽名的成本遠高於它換來的東西,而這個值在整個行程裡只設定一次。
var i18nOverrideDir string

// OpenI18NTSV 開一份譯表(如 "tech.tsv")。
//
// ⚠ **不要再寫 `os.Open("assets/i18n/xxx.tsv")`。** 那是相對於當前工作目錄的路徑,
// 執行檔只有從 repo 根目錄跑才找得到——`i18n_path_test.go` 會擋下新增的這種寫法。
func OpenI18NTSV(name string) (fs.File, error) {
	if i18nOverrideDir != "" {
		return os.DirFS(i18nOverrideDir).Open(name)
	}
	return embeddedI18N.Open("embedded/i18n/" + name)
}
