// lbxdump 把 .lbx 內的影像資產解碼、上色,輸出成 PNG,供人工檢視與驗證解碼正確性。
//
// 用法:
//
//	lbxdump <file.lbx> <outdir> [--pal <palette.lbx>:<index>] [--max N]
//
// 有內嵌調色盤的影像直接上色;無內嵌者需以 --pal 指定外部調色盤來源
// (某 .lbx 內某 index 影像的內嵌調色盤)。
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
)

func main() {
	palFlag := flag.String("pal", "", "外部調色盤來源 <palette.lbx>:<index>")
	maxAssets := flag.Int("max", 0, "最多處理幾個資產(0 = 全部)")
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "用法: lbxdump <file.lbx> <outdir> [--pal <file.lbx>:<idx>] [--max N]")
		os.Exit(2)
	}
	lbxPath, outDir := flag.Arg(0), flag.Arg(1)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}

	var extPal *lbx.Palette
	if *palFlag != "" {
		p, err := loadExternalPalette(*palFlag)
		if err != nil {
			fatal(fmt.Errorf("載入外部調色盤: %w", err))
		}
		extPal = p
	}

	arch, err := openArchive(lbxPath)
	if err != nil {
		fatal(err)
	}

	count := arch.Count()
	if *maxAssets > 0 && *maxAssets < count {
		count = *maxAssets
	}

	images, frames := 0, 0
	for i := 0; i < count; i++ {
		raw, err := arch.Asset(i)
		if err != nil {
			continue
		}
		im, err := lbx.DecodeImage(raw)
		if err != nil {
			continue // 非影像資產
		}
		pal := im.Embedded
		if pal == nil {
			pal = extPal
		}
		if pal == nil {
			fmt.Printf("資產 %d:影像 %dx%d %d 幀,無調色盤(略過,用 --pal 指定)\n", i, im.Width, im.Height, len(im.Frames))
			continue
		}
		images++
		for f, fr := range im.Frames {
			img := fr.ToRGBA(pal, im.KeyColor())
			name := filepath.Join(outDir, fmt.Sprintf("a%03d_f%02d.png", i, f))
			if err := writePNG(name, img); err != nil {
				fatal(err)
			}
			frames++
		}
	}
	fmt.Printf("完成:%s 共 %d 資產,輸出 %d 影像 / %d 幀 PNG 到 %s\n", lbxPath, arch.Count(), images, frames, outDir)
}

func openArchive(path string) (*lbx.Archive, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return lbx.Open(bytes.NewReader(data), int64(len(data)))
}

func loadExternalPalette(spec string) (*lbx.Palette, error) {
	idx := strings.LastIndex(spec, ":")
	if idx < 0 {
		return nil, fmt.Errorf("格式應為 <file.lbx>:<index>")
	}
	n, err := strconv.Atoi(spec[idx+1:])
	if err != nil {
		return nil, fmt.Errorf("index 非數字: %w", err)
	}
	a, err := openArchive(spec[:idx])
	if err != nil {
		return nil, err
	}
	raw, err := a.Asset(n)
	if err != nil {
		return nil, err
	}
	im, err := lbx.DecodeImage(raw)
	if err != nil {
		return nil, err
	}
	if im.Embedded == nil {
		return nil, fmt.Errorf("該資產無內嵌調色盤")
	}
	return im.Embedded, nil
}

func writePNG(name string, img *image.RGBA) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "錯誤:", err)
	os.Exit(1)
}
