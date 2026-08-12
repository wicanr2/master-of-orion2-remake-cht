// zipdir 將目錄以穩定排序寫成 ZIP；只使用 Go 標準庫。
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	root := flag.String("root", "", "要封裝的目錄")
	prefix := flag.String("prefix", "", "ZIP 內的根目錄前綴")
	output := flag.String("output", "", "輸出 ZIP 路徑")
	flag.Parse()
	if *root == "" || *output == "" {
		fatal("必須指定 -root 與 -output")
	}
	rootInfo, err := os.Stat(*root)
	if err != nil || !rootInfo.IsDir() {
		fatal("找不到目錄或不是目錄: %s", *root)
	}
	*prefix = strings.Trim(strings.ReplaceAll(*prefix, `\`, "/"), "/")

	tmp, err := os.CreateTemp(filepath.Dir(*output), ".zipdir-*.tmp")
	if err != nil {
		fatal("建立暫存 ZIP 失敗: %v", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	zw := zip.NewWriter(tmp)
	err = filepath.WalkDir(*root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == *root {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("不支援符號連結: %s", path)
		}
		rel, err := filepath.Rel(*root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if *prefix != "" {
			name = *prefix + "/" + name
		}
		if entry.IsDir() {
			_, err = zw.Create(name + "/")
			return err
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("不支援特殊檔案: %s", path)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		h, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		h.Name = name
		h.Method = zip.Deflate
		// 封裝內容與排序相同時維持可重現雜湊；不可把每次打包的 mtime
		// 混進 ZIP 中造成偽差異。
		h.Modified = time.Unix(0, 0).UTC()
		writer, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if err == nil {
		err = zw.Close()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		fatal("寫入 ZIP 失敗: %v", err)
	}
	if err := os.Rename(tmpName, *output); err != nil {
		fatal("安裝 ZIP 失敗: %v", err)
	}
	// CreateTemp 預設是 0600；打包產物需能被同組 CI／發行流程讀取，但不放寬成
	// 可執行檔或可寫世界。
	if err := os.Chmod(*output, 0o644); err != nil {
		fatal("設定 ZIP 權限失敗: %v", err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "zipdir: "+format+"\n", args...)
	os.Exit(1)
}
