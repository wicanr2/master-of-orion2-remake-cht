# 調色盤鏈:解鎖無內嵌調色盤的原版畫面

> 目標「畫面與原版一模一樣」的關鍵技術瓶頸與解法。第一性原理拆解 + 逐位元組對照 openorion2 權威來源。

## 問題

MOO2 的畫面背景圖存在各 `.lbx` 內。**部分**背景圖自帶內嵌調色盤(LBX 影像旗標 `FLAG_PALETTE=0x1000`),可自包含正確上色——主選單(MAINMENU.LBX 21)、行星列表(PLNTSUM.LBX 0)屬此類,先前已能渲染。

但**多數遊戲內畫面的背景圖不帶(或只帶部分)內嵌調色盤**:研究選擇(TECHSEL)、殖民地(COLONY)、艦艇設計、艦隊、星系 GUI、軍官、外交…。單獨解碼會因缺色而大片黑,是「畫面一模一樣」無法覆蓋全遊戲的主因。

## 機制(對照 openorion2 `gfx.cpp Image::load`)

原版影像載入時可帶一個「基底調色盤」(`base_palette`)。合併規則(逐行對照 `gfx.cpp:359-395`):

1. 先配置 256 色的 `_palettes[0]`。
2. 若呼叫端有給 `base_palette` → `memcpy` 整份 256 色當基底;否則若連自己的內嵌調色盤都沒有 → 丟 `Palette missing`。
3. 若本圖有 `FLAG_PALETTE` → 讀出 `(palstart, palsize)`,把**自己的部分內嵌範圍疊寫**到 `_palettes[0]` 上。

一句話:**最終調色盤 = 依序疊加各提供圖的內嵌範圍當基底 + 本圖自己的部分內嵌範圍覆蓋上去。**

> 修正(實測):提供圖**不必填滿 256 色**,只需其內嵌範圍涵蓋目標圖實際用到的索引即可。例:`buffer0.lbx#0` 只內嵌 0–191、`science.lbx#0` 只內嵌 192–255;消費端圖像剛好只用到被覆蓋的索引。真正填滿 256 色的提供圖很少(如 `info.lbx#1`)。
> 部分畫面是**多段鏈**(艦隊列表:`buffer0.lbx#0` → 疊 `fleet.lbx#111` → 疊 `fleet.lbx#0` 本身),`resolvePalette` 以 `paletteChain []assetRef` 依序疊加支援。

那張「基底提供圖」由各畫面自己指定,通常是同畫面的一個小圖/游標。openorion2 各畫面範例(`getImage(archive, asset, providerPalette)`):

| 畫面 | 背景 | 調色盤提供圖 | 出處 |
|---|---|---|---|
| 研究選擇 | TECHSEL.LBX 0 | SCIENCE.LBX 0(游標) | `tech.cpp:1033` |
| 研究選擇(可取消) | TECHSEL.LBX 0(+offset) | GALAXY.LBX `ASSET_GALAXY_GUI` | `tech.cpp:928` |
| 殖民地 colonist 圖 | RACEICON.LBX | COLONY2.LBX 50 | `colony.cpp:50` |
| 科技總覽 | INFO.LBX `ASSET_INFO_BG` | (該函式先取的 pal) | `info.cpp:1009` |
| 艦隊 | FLEETLIST.LBX `ASSET_FLEET_GUI` | (同上) | `ships.cpp:571` |
| 星系 GUI | GAMEMENU.LBX `ASSET_GALAXY_GUI` | (同上) | `galaxy.cpp:2924` |

> 逐畫面的提供圖與 asset index 都在 openorion2 各 `*.cpp` 的 view 建構子,追 `_bg = gameAssets->getImage(ARCHIVE, ASSET_BG, pal)` 的 `pal` 從哪來即可。

## 本專案的落地

解碼早已解耦:`internal/lbx` 的 `Frame.ToRGBA(pal *Palette, keyColor bool)` 吃任意調色盤,`Image.Embedded`(僅填 `[PalStart, PalStart+PalCount)`)保留本圖部分內嵌範圍。因此只需在渲染端合併:

```
merged := *provider.Embedded          // 基底(完整 256)
for i in [im.PalStart, im.PalStart+im.PalCount):
    merged[i] = im.Embedded[i]         // 疊本圖自己的範圍
rgba := im.Frames[0].ToRGBA(&merged, im.KeyColor())
```

已實作於 `cmd/moo2/interactive.go`:`paletteProvider` + `resolvePalette()`;`loadOverlayScreen(..., prov)` 走此鏈。

## 驗證

- 研究選擇畫面(TECHSEL.LBX 0,原本無法自包含上色):以 SCIENCE.LBX 0 為基底合併後,完整渲染出原版金屬面板 + 8 個研究領域(SELECT NEW RESEARCH / CONSTRUCTION / POWER / CHEMISTRY / SOCIOLOGY / COMPUTERS / BIOLOGY / PHYSICS / FORCE FIELDS)。
- 已接進 `-game` 互動程式(主選單 →「名人堂」入口 → 置中忠實原版研究選擇畫面 → 點擊返回)。

重現:`moo2 -game -data <遊戲夾> -lang zh -font <字型> -shot out.png`(headless 腳本點名人堂)。

## 落差 / 下一步(逐畫面 Phase 4)

- 研究領域名擦底疊字(建設/動力/化學/社會學/電腦/生物學/物理/力場)+ 座標校對(目前先呈現忠實英文原版)。
- 依上表把 COLONY / DESIGN / FLEET / INFO / 星系 GUI 等逐一以調色盤鏈載入 → 接進 `-game` 導覽 → 擦底疊字中文化。
- 提供圖 asset index 需逐畫面對照 openorion2 建構子確認(勿憑記憶)。
