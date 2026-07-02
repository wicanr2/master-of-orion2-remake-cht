# LBX 資產格式(逆向 + 移植紀錄)

> 記錄 Master of Orion II `.lbx` 資產檔的精確位元組佈局。來源:openorion2 `src/lbx.cpp` / `src/gfx.cpp`(GPL v2)逐位元組核對 + 本專案 `internal/lbx` Go 移植 + 真實檔驗證。
> 所有數值皆**小端(little-endian)**。已驗證檔:TECHNAME/RACES/GAME/PATCH13/BEAMS/COLREFIT(patch 1.31)。

## 1. 容器格式(LBXArchive)

`.lbx` 是一個資產封存檔。Header 後接一張 offset 表,靠相鄰 offset 相減得出每個資產大小。

| offset | 型別 | 內容 |
|---|---|---|
| 0 | u16 | `assetCount`(資產數,不得為 0) |
| 2 | u16 | `magic`,**必為 `0xfead`** |
| 4 | u32 | 版本/旗標(讀後丟棄) |
| 8 | u32 | 第 0 個資產的起始 offset |
| 12 + 4·i | u32 | 第 i+1 個 offset(i = 0 … assetCount−1) |

即從 byte 8 起共有 **`assetCount + 1`** 個 u32 offset。第 i 個資產 = `[offset[i], offset[i+1])`,`size = offset[i+1] − offset[i]`(須嚴格遞增)。

- 資產內容為原始位元組;型別(影像/調色盤/字串/字型)由**使用端**決定,容器不標記。
- Go 實作:`internal/lbx/lbx.go`(`Open` / `Asset`),以 `io.ReaderAt` 惰性讀取,不整包載入。

驗證數據:TECHNAME.LBX=6 資產、RACES.LBX=64、GAME.LBX=32、PATCH13.LBX=2。

## 2. 影像資產格式(Image)

### 2.1 Header(12 bytes)

| offset | 型別 | 欄位 |
|---|---|---|
| 0 | u16 | width |
| 2 | u16 | height |
| 4 | u16 | (未知,略過) |
| 6 | u16 | frameCount(幀數) |
| 8 | u16 | frameTime(幀間毫秒) |
| 10 | u16 | flags(見下) |

### 2.2 Flags 位元

| 值 | 名稱 | 意義 |
|---|---|---|
| `0x2000` | JUNCTION | (junction 影像) |
| `0x1000` | PALETTE | 內嵌調色盤 |
| `0x0800` | KEYCOLOR | index 0 視為透明 |
| `0x0400` | FILLBG | 填背景 |
| `0x0100` | NOCOMPRESS | 未壓縮(逐位元組即 index) |

### 2.3 Frame offset 表

Header 之後接 **`frameCount + 1`** 個 u32,為**資產內絕對 offset**;最後一個須等於資產長度。第 i 幀資料 = `[offset[i], offset[i+1])`。

### 2.4 內嵌調色盤(僅 FLAG_PALETTE)

緊接 frame offset 表:

| 型別 | 欄位 |
|---|---|
| u16 | palStart(起始色索引) |
| u16 | palCount(色數;palStart+palCount ≤ 256) |
| palCount × 4 bytes | 色資料 |

每色 4 bytes:`byte0` 未用、`byte1 = R`、`byte2 = G`、`byte3 = B`。**R/G/B 為 6-bit VGA(0–63),載入時 `<<2` 放大成 8-bit**(0,4,8,…,252)。Alpha 固定 0xFF。
> 註:`<<2` 會遺失低 2 bit(未用 `<<2 | >>4` 補);本移植沿用原版行為以求還原一致。

### 2.5 scan-line RLE 解碼(壓縮幀)

未壓縮(FLAG_NOCOMPRESS):直接 `width×height` 個 byte,每 byte 一個 palette index,全部視為已寫入。

壓縮幀為逐掃描線 RLE:

1. 讀 u16 `size`(**必為 1**,首列標記)、u16 `y`(起始掃描線)。
2. `while y < height`:對每列 `x` 從 0 起:
   - 讀 u16 `size`、u16 `skip`。
   - 若 `size == 0`:`y += skip`,結束本列(換行)。
   - 否則:檢查 `x + skip + size ≤ width`;`x += skip + size`;跳過 `skip` 個透明像素;讀 `size` 個 byte(palette index)寫入。
   - **`size` 為奇數 → 補讀 1 byte 對齊**。

透明來源有二:① RLE 跳過(skip)未寫入的像素;② KEYCOLOR 時 index 0 的像素。

### 2.6 移植設計:解碼與上色解耦

Go 實作(`internal/lbx/image.go`)把「解碼」與「上色」分離,**優於** openorion2 的耦合寫法:

- `DecodeImage` → 每幀輸出 `Frame{Index []uint8, Written []bool}`(palette 無關)。
- `Frame.ToRGBA(pal, keyColor)` → 套指定調色盤上色成 `*image.RGBA`。
- 好處:同一幀可套不同 palette variant(原版功能),且解碼結果可 byte-精確斷言,不需真實調色盤即可測。

驗證數據:BEAMS.LBX 153/153、COLREFIT.LBX 5/5、GAME.LBX 32/32 個資產皆無誤解碼、幀尺寸一致。

### 2.7 Bitmap 與 dirty-block:格式相同,優化刻意不移植

openorion2 另有 `Bitmap`(`gfx.cpp:671-765`)。經核對,其像素編碼**與 Image 完全相同**(同 NOCOMPRESS、同 scan-line RLE、首列標記 size=1、奇數補位)。唯一差別是額外維護 **dirty-block 矩形清單**(`blocks`/`oldBlocks`),計算相鄰幀間變動的矩形。

這是 **SDL 局部 blit 的效能優化**(只重繪變動區)。本專案用 ebiten **每幀全繪**,不需要增量矩形 → 這道「柵欄」的前提在我們的架構下不存在,**刻意不移植**(非因投報跳過)。像素本身:`Bitmap` 存 8-bit 索引到 buffer,正是 `Frame.Index` 已提供的,故現有 `DecodeImage` 已涵蓋 Bitmap 的解碼需求。

> 尚未做:像素**顏色**的視覺正確性(需 Phase 2 能 render + 截圖比對)。
