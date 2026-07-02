# 魔法大帝(mom)ebiten 繁中化 playbook — 可搬用做法

> 來源:`~/master-of-maigc/repo`(魔法大帝繁中化,疊在 kazzmir/master-of-magic 的 Go1.25+Ebiten 引擎,patch-only,已三平台出包)。
> 其自有的 `docs/localization-methodology.md` 是完整 playbook,本文是「對銀河霸主2 可直接搬用」的萃取版。
> **重要前提差異**:mom 是 **patch 既有 Go 引擎**;我們是 **從零用 Go 重寫 openorion2(C++)**。
> → 與**引擎無關**的做法(i18n 覆蓋層、字型子集、驗證紀律、漏譯成因)**直接照搬**;與「注入既有繪字迴圈」相關的做法,我們改成「自己的繪字管線一開始就內建」。

## 1. i18n:顯示層覆蓋 + 英文原文即 key(★ 直接照搬 ★)

- **單一機制**:顯示前才把英文換中文(display-layer override),**不動資料層**。
- **[HARD] 為什麼是顯示層**:引擎常把**英文字串當邏輯 key**(mom 實例:`if string(name)=="+6 Defense"`)。改資料層中文 → 遊戲邏輯壞掉。openorion2 同樣大量從 LBX 讀字串當識別 → **這條一定會遇到,照搬**。
- **表格式**:TSV 三欄 `英文原文<TAB>中文<TAB>備註`。**英文原文即 key**(好處:同表可相容不同遊戲版本,對我們正好對應 1.3/1.5)。
- **查找**:`TrimSpace` 後精確 map 查找(對齊引擎讀 LBX 時的 trim);查無 → 回原字串(英文版零影響);空中文欄 → 略過(選擇性覆蓋)。
- **`TranslateFormat`**:處理 `fmt.Sprintf` 模板 —— 先翻**模板字面**(`"Cost %v" → "花費 %v"`)再填值(填值後整串比對不會命中)。[HARD] 佔位符數量/順序中英必一致(否則 panic)。內嵌列舉要各自再翻。只包顯示用 Sprintf,跳過 log/存檔/`==` 比對。
- **載入優先序**:env 目錄(dev 覆寫)先載、內嵌表補齊,先到先得;用檔名字母序控制同 key 優先(專名檔早於通用檔)。
- **專名「中文(英文)」**:專名檔存「中文(半形英文)」,用控制碼(`\x0e`…`\x0f`)標記尾端英文,繪製時降字級 0.65、對齊基線、不繪標記本身。老玩家仍認得原文。

## 2. CJK 文字渲染(取其原理,實作因我們從零而異)

mom 是 patch,所以**注入**引擎既有的 rune 迭代繪字迴圈(`glyphIndex=int(c)-32` 只認 ASCII,非 ASCII 被靜默丟棄 → 在丟棄前插 CJK 分支)。**我們從零寫,直接把 CJK 當一等公民**(可用 ebiten `text/v2` 或自製 glyph 快取)。但這些原理不分實作,**照搬**:

- **[HARD] supersample,不可「小字放大」**:CJK glyph 以「logical 字高 × 4」(`cjkSupersample=4`)高解析 rasterize,再縮回 logical footprint。**絕不**拿小點陣放大(會糊)。在 320×200 邏輯座標系繪製即可,不需要獨立高解析合成層。
- **glyph 快取**:key = `(rune, 字高)`;face 也依字高快取。
- **[HARD] 字級對齊呼叫端字高**:CJK 依「該處字型的 GlyphHeight」決定尺寸,否則中文比英文高 → 行距重疊破版(mom 真跑畫面才抓到)。基線比例約 0.82。
- **所有繪字路徑都要支援 CJK**:一般印字、**描邊/陰影版**、**量寬(MeasureTextWidth)**。mom 描邊版一開始漏做 → 標題中文被丟棄。量寬漏做 → 置中/置右/換行全算錯。
- **[HARD] CJK 無空白 → 自寫逐字斷行**:原版 `splitText` 靠空白斷行,對中文長串回傳空 → 整段被丟。要逐 rune 斷行,且「無可斷點時至少切一 rune」。

## 3. 烘進圖片的英文(按鈕/標題)

font 畫的文字自動走覆蓋層即可;**烘進點陣圖**的英文兩條路:

1. **擦底疊字 helper**(`cht_label.go`):先用深色 `vector.DrawFilledRect` 擦掉英文(保外框),再置中 + drop-shadow 疊銳利中文。字級由傳入字型決定(太小就換大字型)。**必須在標準 UI 繪製之後疊**。
2. **整張圖替換**(`image_override.go`):在**單一載圖入口**攔截,若外部目錄有 `<lbx>_<index>.png` 就回傳中文版圖,版權 LBX 不動。附**探查模式**(IMGLOG):記錄每個被請求的 `(lbx,index)`,幫你對照哪張圖是哪個 UI。
- Draw closure 內用後宣告的圖要**重抓**(GetImages 有快取很便宜),避免 Go 閉包抓錯變數。

## 4. 字型(★ 修正先前建議,見 `04`)

- **實際用 Noto Sans CJK TC**(OFL,可散布)。即時 rasterize 與內嵌子集都是它。
- **[HARD] 教訓**:原想用 AR PL UMing(點陣風),但**其 `.ttc` 是 CFF/舊式,`golang.org/x/image/sfnt` 解析失敗**。→ **任何字型定案前,先確認 Go 的 opentype/sfnt 解析得動**(glyf 輪廓 TTF 較穩,CFF/.ttc 有風險)。
- **子集**:`pyftsubset`(docker + uv venv),字集 = 「所有 .go 非 ASCII 字 + 全 TSV 中文欄」`sort -u`;`--no-hinting --desubroutinize --drop-tables+=GSUB,GPOS`;用 Regular 保字重一致。約 392KB(~1,700 字)。
- **go:embed 內嵌** → standalone 零外部字型依賴。**[HARD] 每次加新譯文要重生子集**,否則新字顯示空白方塊。

## 5. Patch / repo 紀律(部分適用)

- **相同(照搬)**:repo **只放**我們的碼 + 文件 + TSV 譯表 + 字型子集 + 烘字腳本;**不放**版權 LBX/exe/手冊/上游引擎。字型子集 + 譯表 go:embed 內嵌。
- **不同**:mom 用 `fetch-engine(釘 commit) + 單一權威 patch 0099` 疊上游 Go 引擎;**我們是自己的 Go 原始碼,直接進 repo**,沒有上游 Go 引擎要 patch(openorion2 只是 C++ 參考,不編進我們的 build)。
- **[HARD] 別 gofmt 上游既有檔**(對 mom):我們無此問題,但自訂 gofmt 紀律一致。

## 6. 開發 / 驗證(★ 直接照搬 ★)

- **全程 docker**,無本機 toolchain。build image 內含 X11/ALSA/xvfb/imagemagick。
- **[HARD] headless 逐畫面截圖驗證**:`xvfb-run` 跑 → `xdotool` 導航點選單 → `import -window root` 抓圖 → Read 看。**「CI 編譯全綠 ≠ 畫面對」**,每批改動都截圖驗。
- headless 音效雷:null PCM(`pcm.!default {type null}`)或關音樂,否則 oto ALSA panic。`LIBGL_ALWAYS_SOFTWARE=1`。
- **dump 用引擎自己的 reader**(別自己 parse LBX 二進位)dump 精確英文 key;headless dump 要帶 `--frames N`/`timeout` 否則空轉燒 CPU。
- 三平台打包:Linux AppImage(CGO)、Windows(`CGO_ENABLED=0`→ 免外部 DLL)、macOS arm64(需 CGo/Metal → GitHub Actions macos runner,不能從 Linux 交叉編)。
- go:embed 新檔名增量不重嵌 → 清 GOCACHE 或 touch 引用檔。

## 7. 漏譯的系統性成因(★ 最值錢,開工先防 ★)

- **字串不只在資料檔,也 hardcode 在 Go source** → 別假設「文字全在 LBX」。
- **多個獨立文字源,漏 dump 一個 = 整類畫面英文**(mom 漏過整類法術描述、建造短描述)。openorion2 的 `TextManager` 就有主文字/事件/種族/艦名/星名/科技/外交/help 多源。
  → **[HARD] 開工第一步:窮舉所有文字源(含 hardcode),各寫一支 dumper**(對齊 rulebook 83「同類檔先列舉全」)。
- **組合字串**填值後比不中 → 翻模板字面(`TranslateFormat`)。
- **大小寫/標點不一致**配不上 key → 正規化。
- **同英文不同語境**(collision)→ 該處改用獨立 key。
- **並行翻譯 agent 切塊**:大量字串切多塊各寫獨立輸出檔,翻譯 agent 不碰 git/不 build,主 agent 統一整合 + build;逐塊核對行數抓漏。

## 可直接搬清單(TL;DR)

1. 顯示層覆蓋 + TSV(英文即 key)+ `TranslateFormat` 模板 —— i18n 地基。
2. supersample(4×)glyph + `(rune,字高)` 快取 + 對齊呼叫端字高 + 逐字斷行。
3. **顯示層翻譯,不動資料層**(避免破壞把英文當 key 的邏輯)[HARD]。
4. 擦底疊字 + 整圖替換 + IMGLOG 探查 —— 處理烘進圖的英文。
5. go:embed 字型子集(Noto TC OFL,pyftsubset docker)+ 譯表,零外部依賴;加字重生子集。
6. docker + xvfb + xdotool + import 逐畫面截圖;「編譯綠 ≠ 畫面對」。
7. 開工先窮舉所有文字源(含 hardcode)並各寫 dumper,用引擎自己 reader dump 精確 key。
