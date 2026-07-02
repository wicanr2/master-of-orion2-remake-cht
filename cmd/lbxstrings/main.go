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
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
)

func main() {
	asset := flag.Int("asset", -1, "只 dump 指定資產(預設全部)")
	cstr := flag.Bool("cstr", false, "用 C-string 格式解析")
	offset := flag.Int("offset", 0, "C-string 解析起始 offset(rstring=4, estrings/hstrings=6)")
	loadfile := flag.Bool("loadfile", false, "用 loadFile 格式(每資產一則訊息)")
	diplo := flag.Bool("diplo", false, "用 DIPLOMSE 格式(每資產 header + N 個固定寬字串)")
	groups := flag.Int("groups", 1, "loadFile 語言群組數(英文取前 count/groups 個資產)")
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

	// loadFile 模式:每資產一則訊息,英文取前 count/groups 個。
	if *loadfile {
		end := arch.Count() / *groups
		for i := 0; i < end; i++ {
			raw, err := arch.Asset(i)
			if err != nil {
				continue
			}
			s, err := lbx.ParseFileEntry(raw)
			if err != nil || s == "" {
				continue
			}
			if *tsv {
				fmt.Printf("%s\t\t\n", escapeForTSV(s))
			} else {
				fmt.Printf("=== 訊息 %d ===\n%s\n", i, s)
			}
		}
		return
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
		if *diplo {
			strs, err = lbx.ParseDiploStrings(raw)
			if err != nil {
				continue
			}
		} else if *cstr {
			strs = lbx.ParseCStrings(raw, *offset)
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
					fmt.Printf("%s\t\t\n", escapeForTSV(s))
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

// escapeForTSV 把控制碼/非可印位元組轉成 \xNN,\n→\n,\t→\t,方便寫進 TSV。
func escapeForTSV(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\n':
			b.WriteString("\\n")
		case c == '\t':
			b.WriteString("\\t")
		case c == '\\':
			b.WriteString("\\\\")
		case c < 32 || c > 126:
			fmt.Fprintf(&b, "\\x%02x", c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "錯誤:", err)
	os.Exit(1)
}
