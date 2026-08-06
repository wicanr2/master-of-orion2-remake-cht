// smkdump:把 Smacker 影片的指定幀解成 PNG(逆向驗收用)。
//
// 用法:smkdump <file.smk> <outdir> [每 N 幀存一張] [最多存幾張]
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/smk"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "用法: smkdump <file.smk> <outdir> [每N幀] [最多幾張]")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fatal(err)
	}
	d, err := smk.Open(data)
	if err != nil {
		fatal(err)
	}
	every, maxShots := 1, 8
	if len(os.Args) > 3 {
		every, _ = strconv.Atoi(os.Args[3])
	}
	if len(os.Args) > 4 {
		maxShots, _ = strconv.Atoi(os.Args[4])
	}
	if every < 1 {
		every = 1
	}
	outDir := os.Args[2]
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}
	fmt.Printf("%s %dx%d %d 幀 每幀 %d ms(幀資料吃完後殘餘 %d bytes,健康值 0)\n",
		d.H.Signature, d.H.Width, d.H.Height, d.H.Frames, d.H.FrameRateMS, d.Trailing())

	for _, w := range d.TreeWarnings() {
		fmt.Printf("  ⚠ %s\n", w)
	}
	shots, overFrames, overBits := 0, 0, 0
	for i := 0; i < d.H.Frames && shots < maxShots; i++ {
		pix, pal, err := d.DecodeNext()
		if err != nil {
			fatal(fmt.Errorf("第 %d 幀: %w", i, err))
		}
		if os.Getenv("SMKDUMP_USAGE") != "" {
			nb, bits := d.LastVideoUsage()
			if over := bits - nb*8; over > 0 {
				overFrames++
				overBits += over
				if overFrames <= 6 {
					total, overAt, byType := d.LastBlockStats()
					pct := 0
					if total > 0 && overAt >= 0 {
						pct = overAt * 100 / total
					}
					fmt.Printf("  超讀 幀 %d: %d bytes,多讀 %d bits;超讀起點 區塊 %d/%d(%d%%);型別次數 單色%d 全彩%d 略過%d 填色%d\n",
						i, nb, over, overAt, total, pct, byType[0], byType[1], byType[2], byType[3])
					fmt.Printf("       區塊%%/位元%%:")
					for _, c := range d.LastBlockCurve() {
						fmt.Printf(" %d/%d", c[0], c[1])
					}
					fmt.Println()
				}
			}
		}
		if i%every != 0 {
			continue
		}
		img := image.NewRGBA(image.Rect(0, 0, d.H.Width, d.H.Height))
		for k, idx := range pix {
			img.Set(k%d.H.Width, k/d.H.Width, color.RGBA{pal[int(idx)*3], pal[int(idx)*3+1], pal[int(idx)*3+2], 0xFF})
		}
		// 診斷用:同時輸出「調色盤索引當灰階」與「調色盤色卡」。
		// 索引圖乾淨而彩色圖髒 → 是調色盤的問題;索引圖就髒 → 是影像解碼的問題。
		if os.Getenv("SMKDUMP_DIAG") != "" {
			gray := image.NewRGBA(image.Rect(0, 0, d.H.Width, d.H.Height))
			for k, idx := range pix {
				gray.Set(k%d.H.Width, k/d.H.Width, color.RGBA{idx, idx, idx, 0xFF})
			}
			writePNG(filepath.Join(outDir, fmt.Sprintf("f%04d_idx.png", i)), gray)
			sw := image.NewRGBA(image.Rect(0, 0, 16*8, 16*8))
			for e := 0; e < 256; e++ {
				c := color.RGBA{pal[e*3], pal[e*3+1], pal[e*3+2], 0xFF}
				for yy := 0; yy < 8; yy++ {
					for xx := 0; xx < 8; xx++ {
						sw.Set((e%16)*8+xx, (e/16)*8+yy, c)
					}
				}
			}
			writePNG(filepath.Join(outDir, fmt.Sprintf("f%04d_pal.png", i)), sw)
		}
		name := filepath.Join(outDir, fmt.Sprintf("f%04d.png", i))
		f, err := os.Create(name)
		if err != nil {
			fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			fatal(err)
		}
		f.Close()
		shots++
	}
	if os.Getenv("SMKDUMP_USAGE") != "" {
		fmt.Printf("超讀的幀數 %d,合計多讀 %d bits\n", overFrames, overBits)
	}
	fmt.Printf("輸出 %d 張到 %s\n", shots, outDir)
}

func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "錯誤:", err)
	os.Exit(1)
}
