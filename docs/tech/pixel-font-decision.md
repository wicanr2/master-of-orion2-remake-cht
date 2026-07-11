# 點陣中文字型決策:採用 bitmapfont/v4 FaceTC(2026-07-11)

> 承接 `docs/tech/ui-typography-button-review.md`(設計師 review)與使用者決策「一致化 + 點陣中文字型」。
> 本文記錄可行性驗證證據與定案,取代 `docs/kickoff/04-font-choice.md` 對「像素字型待驗」的 open 狀態。

## 1. 決策

- **中文 UI 字型改用 `github.com/hajimehoshi/bitmapfont/v4` 的 `FaceTC`(繁體中文優先點陣字)。**
- **Noto Sans CJK TC 保留為缺字 fallback**(對齊 `04`/`08` 的 [HARD]:缺字靜默消失是 MOO1 踩過的雷)。
- 這是 ebiten 作者維護的點陣字套件,`FaceTC` 底層由 **Cubic 11**(方舟像素,OFL-1.1,繁中導向)+ **Ark Pixel Font**(OFL-1.1)拼成,`font.Face` 介面 → `text.NewGoXFace()` 直接接進現行 `internal/uifont` 的 `text/v2` 管線,免自寫 BDF parser。

## 2. 為什麼點陣字要用「真 bitmap」而非「像素風 TTF 直接 rasterize」

`~/.claude/knowledge-base/retro-cht/qb64pe-game-linux-port` 的 [HARD] 教訓:**「字型升級走 BDF 路徑…不要引入 freetype rasterize TTF(會有 hinting 變糊問題)」**。現行 `uifont` 用 `opentype.NewFace`+`text/v2`(向量 rasterize + 抗鋸齒),直接換「像素風 TTF」在非整數尺寸/抗鋸齒下仍會糊。`bitmapfont/v4` 是**真點陣**(glyph 是 1-bit 位圖),渲染逐像素銳利,並**完全繞過** `08` §2 的 supersample 議題(點陣不需要 supersample,放大用整數最近鄰即可)。

## 3. 可行性驗證證據(headless `font.Drawer` 探針,非假設)

探針渲染 8 行涵蓋全風險畫面的實際 UI 中文(主選單/工具列/戰鬥控制列按鈕、13 種族名、科技/殖民地/勝利術語),結果:

| 驗證項 | 結果 |
|---|---|
| 字元覆蓋 | 8 行**合計缺字 0**,無 tofu(含 隆/瑞/薩/崔/矽/嚙 等密筆劃繁體字) |
| metrics | Ascent=12 Descent=4 Height=16;全形 glyph 12×13、半形 6×13 |
| 點陣觀感(2×) | 銳利方塊感,契合 MOO2 320×200 upscale 美術 |
| 小字可讀(1× 12px) | 密繁體字清晰,達 `04` 判準 4(小字可讀) |
| 整合 | `bitmapfont.FaceTC`(`font.Face`)可被 `text.NewGoXFace` 包裝,與現有 `Font.Face` 對稱 |

> 探針碼與輸出圖:session scratchpad `fontprobe/`(throwaway,不進 repo)。

## 4. 設計後果:字級階層由「5 級向量」收斂為「2 級點陣」

點陣字只在原生 12px 銳利,無法任意縮放。設計師 review §2 的 5 級向量階層(T1 20-22 / T2 16-17 / T3 13-15 / T4 12-13 / T5 11)在點陣字下**收斂為 2 級**:

| Tier | 語意 | 點陣實作 | px |
|---|---|---|---|
| 標題 | 畫面標題(戰術戰鬥/銀河議會/選擇新研究) | FaceTC 2× 最近鄰 | 24 |
| 內文 | 按鈕/表格/資訊面板/小字 | FaceTC 1× 原生 | 12 |

這比 5 級更**忠實** 90 年代遊戲(當年多半只有 1–2 種字級)。中間態(現行 14/15/16/20)一律歸入這兩級之一(標題→24、其餘→12)。**視覺層級改由「大小 2 級 + 顏色 + 位置」共同建立,不再靠連續字級**。

## 5. 實作計畫(分階段,階段間驗證)

1. **`uifont` 加點陣模式**:新增以 `bitmapfont.FaceTC` 為源的 face + 缺字時回落 Noto 的複合繪字;`Draw/DrawCentered/Measure` 對外 API 不變。2× 標題用整數縮放 + `FilterNearest`(避免放大糊)。
2. **切換 zh 預設字型為點陣**;英文/`-lang en` 維持 Noto(或評估 bitmapfont 半形拉丁,截圖比較後定)。
3. **字級常數歸 2 級**:全 `Draw` 呼叫點的 size 參數收斂為 12/24;逐畫面截圖檢查 12px 是否撐破密表(殖民地/行星列表現行 10px),必要處調版位而非縮字。
4. **一致化包**:戰鬥控制列 `drawBarLabelsCHT` 改用 `samplePlate`(消兩套機制)+ `uifont` 新增描邊/陰影繪字(對齊 `08` §2「描邊版漏做→標題中文被丟棄」),套用所有疊字按鈕。
5. **授權**:go.mod 加 `bitmapfont/v4`;README/致謝標 Cubic 11 + Ark Pixel + bitmapfont 的 OFL/Apache 授權。

## 6. 待驗 / 風險

- 12px 內文在現行 10px 密表可能撐破 → 階段 3 逐畫面截圖檢查,調版位。
- 英文介面是否也改點陣半形(統一風格)vs 維持 Noto → 截圖 A/B 後定,不阻塞中文路徑。
- `bitmapfont` 半形數字/標點與全形中文混排的基線/間距 → 實機截圖確認對齊。
