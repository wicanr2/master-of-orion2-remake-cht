// lbxstrings dump 一個 .lbx 內各資產的字串(先試固定寬格式,可選 C-string 格式),
// 供找出名稱/描述表並建立翻譯 TSV。
//
// 用法:
//
//	lbxstrings <file.lbx> [--asset N] [--cstr] [--tsv]
//	  --asset N  只 dump 第 N 個資產
//	  --cstr     改用 C-string 格式解析
//	  --tsv      輸出 TSV 三欄骨架(英文<TAB><TAB>),供翻譯
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
)

func main() {
	asset := flag.Int("asset", -1, "只 dump 指定資產(預設全部)")
	cstr := flag.Bool("cstr", false, "用 C-string 格式解析")
	tsv := flag.Bool("tsv", false, "輸出 TSV 骨架供翻譯")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "用法: lbxstrings <file.lbx> [--asset N] [--cstr] [--tsv]")
		os.Exit(2)
	}
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	arch, err := lbx.Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		fatal(err)
	}

	for i := 0; i < arch.Count(); i++ {
		if *asset >= 0 && i != *asset {
			continue
		}
		raw, err := arch.Asset(i)
		if err != nil {
			continue
		}
		var strs []string
		if *cstr {
			strs = lbx.ParseCStrings(raw, 0)
		} else {
			strs, err = lbx.ParseFixedStrings(raw)
			if err != nil {
				if !*tsv {
					fmt.Printf("# 資產 %d:非固定寬字串(%v)\n", i, err)
				}
				continue
			}
		}
		if *tsv {
			for _, s := range strs {
				if s != "" {
					fmt.Printf("%s\t\t\n", s)
				}
			}
		} else {
			fmt.Printf("=== 資產 %d:%d 條字串 ===\n", i, len(strs))
			for j, s := range strs {
				fmt.Printf("  [%d] %q\n", j, s)
			}
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "錯誤:", err)
	os.Exit(1)
}
