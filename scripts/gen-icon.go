//go:build ignore

// gen-icon 產生一個不依賴任何遊戲資產(版權素材)的 256x256 佔位圖示,
// 供 AppImage / .desktop 使用。純 Go stdlib 畫圖,無外部依賴。
// 用法: go run scripts/gen-icon.go <輸出.png>
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func main() {
	out := "icon.png"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	const size = 256
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	bg := color.RGBA{0x0a, 0x10, 0x24, 0xff}   // 深空藍黑
	ring := color.RGBA{0xd4, 0xaf, 0x37, 0xff} // 金色(orion 環)
	star := color.RGBA{0xf5, 0xe6, 0xa8, 0xff} // 淡金星芒

	cx, cy := float64(size)/2, float64(size)/2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bg)
			dx, dy := float64(x)-cx, float64(y)-cy
			dist := math.Hypot(dx, dy)
			// 外環(代表星系軌道)
			if dist > 90 && dist < 100 {
				img.Set(x, y, ring)
			}
			// 中心恆星
			if dist < 40 {
				t := 1 - dist/40
				c := color.RGBA{
					R: uint8(float64(star.R) * t),
					G: uint8(float64(star.G) * t),
					B: uint8(float64(star.B) * t),
					A: 0xff,
				}
				img.Set(x, y, blend(bg, c, t))
			}
		}
	}

	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func blend(bg, fg color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(bg.R)*(1-t) + float64(fg.R)*t),
		G: uint8(float64(bg.G)*(1-t) + float64(fg.G)*t),
		B: uint8(float64(bg.B)*(1-t) + float64(fg.B)*t),
		A: 0xff,
	}
}
