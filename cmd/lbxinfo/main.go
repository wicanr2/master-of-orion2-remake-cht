// lbxinfo:dump 一個 LBX 內每個資產的影像結構(尺寸/旗標/幀數/是否內嵌調色盤)。
// 用途:看清 DIPLOMAT 這類檔的佈局(哪些是螢幕大圖=使節房、哪些是小圖/調色盤源),
// 不靠猜。headless、只讀,不進 repo 版控素材。
package main

import (
	"fmt"
	"os"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法: lbxinfo <file.lbx>")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	st, _ := f.Stat()
	a, err := lbx.Open(f, st.Size())
	if err != nil {
		panic(err)
	}
	fmt.Printf("資產數: %d\n", a.Count())
	fmt.Printf("%-4s %-6s %-6s %-6s %-8s %-7s %s\n", "idx", "w", "h", "frames", "flags", "pal?", "palN")
	for i := 0; i < a.Count(); i++ {
		raw, err := a.Asset(i)
		if err != nil {
			fmt.Printf("%-4d <read err: %v>\n", i, err)
			continue
		}
		im, err := lbx.DecodeImage(raw)
		if err != nil {
			fmt.Printf("%-4d <decode err, %d bytes>\n", i, len(raw))
			continue
		}
		pal := "no"
		if im.Embedded != nil {
			pal = "YES"
		}
		fmt.Printf("%-4d %-6d %-6d %-6d 0x%04x   %-7s %d\n",
			i, im.Width, im.Height, len(im.Frames), im.Flags, pal, im.PalCount)
	}
}
