# CJK 畫面渲染樣板(百科檢視器為首例)

未來每個「自繪中文畫面」都走這條鏈路。以百科檢視器(`cmd/moo2 -help-viewer`)為首個實例。

## 全鏈路

```
LBX 資料檔 ──ParseXxx──▶ 英文原字串 ──Registry.Source(來源).Translate──▶ 中文
                                                                        │
                                          Font.Wrap(size, maxWidth) ◀───┘
                                                   │  CJK-aware 換行
                                                   ▼
                                     自繪 ebiten 面板(Draw 逐行)──▶ headless 截圖驗證
```

## 三塊可重用基建

1. **`i18n.Registry`(per-source catalog)** — `internal/i18n/registry.go`
   - `reg.LoadFS(os.DirFS("assets/i18n"), ".")` 載入全部 TSV,每檔一個具名來源。
   - 畫面字串來自哪個 LBX,就用對應來源查:百科本文用 `reg.Source("help")`;
     標題可能是科技名,用 `reg.Translate`(merged 備援)。
   - 忠於 MOO2 `(表,id)` 查詢,同形詞各表獨立(見 `i18n-catalog-architecture.md`)。

2. **`uifont.WrapText` / `Font.Wrap`** — `internal/uifont/wrap.go`
   - ebiten `text/v2` **沒有**自動換行,須自備。
   - CJK 逐字可斷、拉丁字詞在空白斷、尊重 `\n`、超長 token 才硬切。
   - `WrapText(measure func(string) float64, ...)` 用注入式量測 → 純邏輯可單測(fake measure);
     `Font.Wrap` 包真實 `Measure`。

3. **自繪面板 + 截圖 harness** — `cmd/moo2/help.go`
   - 深色面板自繪(不依賴原版背景圖 → 截圖只需該 LBX + 字型,無需整套遊戲美術)。
   - `-shot out.png -frames N`:跑 N 幀存 PNG 後結束(headless 驗證,承 mom「編譯綠 ≠ 畫面對」)。

## 控制碼淨化

HELP 本文含 `\x07X<數字>.` 欄位定位碼(表格排版)。**先用 raw key 查譯文,再淨化供顯示**
(`sanitizeHelpText` 把 `\x07…` 換空白)。順序不可反:譯表 key 是含控制碼的原文。

## 重現截圖(dev)

```bash
SP=<scratchpad>            # 內含玩家自備 HELP.LBX 的資料夾 lbxtest/
docker run --rm \
  -v "$PWD:/src" -v "$PWD/.docker-cache/go:/go" \
  -v "$SP/lbxtest:/data:ro" -v "/usr/share/fonts/opentype/noto:/fonts:ro" -v "$SP:/out" \
  -w /src moo2-ebiten bash -c '
    go build -buildvcs=false -o /tmp/moo2 ./cmd/moo2
    Xvfb :99 -screen 0 800x600x24 >/dev/null 2>&1 & sleep 2; export DISPLAY=:99
    /tmp/moo2 -data /data -help-viewer -help-index 2 -lang zh \
      -font /fonts/NotoSansCJK-DemiLight.ttc -shot /out/help.png -frames 3'
```

> 註:uifont 套件 import ebiten,`go test ./internal/uifont` 需 Xvfb(否則 ebiten init panic)。
> 純邏輯的 wrap 測試也因同套件連帶需要 display,故一律在 `moo2-ebiten` image + Xvfb 下跑。

## 尚待(百科檢視器之後)

- 條目清單/索引導覽、捲動(目前 MVP 單條、超出面板下緣截斷)。
- 插圖:`HelpEntry.Archive/AssetID/Frame` 已解析,尚未載圖顯示。
- 表格類條目(`\x07` 定位)目前淨化為空白,未還原欄位對齊。
