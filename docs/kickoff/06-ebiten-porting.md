# ebiten 移植策略(kick-off)

> 把 openorion2(C++/SDL2)的架構映射到 go/ebiten。依據見 `01-openorion2-assessment.md`。
> 註:CLAUDE.md 列的 go/ebiten 參考 `~/master-of-magic` 本機不存在,本文以 openorion2 架構 + ebiten 官方模型推導,待補該參考後回填實戰心得。

## 1. 兩個框架的核心模型對映

| openorion2 (SDL2) | ebiten | 移植動作 |
|---|---|---|
| `main_loop`:`SDL_PollEvent` + `SDL_Delay(10)`(`sdl_events.cpp`) | `ebiten.RunGame`,每 tick 呼叫 `Update()` / `Draw()` | 把事件輪詢改成在 `Update()` 讀 `inpututil` 狀態 |
| 抽象 `Screen` 介面(`screen.h`) | 實作一個 `ebitenScreen` 滿足同介面 | 只換後端,不動上層 `gui.cpp` 邏輯 |
| `registerTexture`/`drawTexture`/`fillRect`/`setClipRegion` | `ebiten.Image` + `DrawImage`/`SubImage`(clip)/`vector` | 逐方法對應 |
| 邏輯座標 640×480 + `RenderSetLogicalSize` | `Game.Layout(w,h)` 回傳邏輯尺寸,ebiten 自動縮放 | 直接對應 |
| 只有滑鼠事件 | `ebiten.CursorPosition` + `inpututil.IsMouseButtonJustPressed` | 補上鍵盤(原版沒有,我們加)|
| bitmap font(單 byte glyph) | `text/v2` + `opentype`(TTF) | **整組換掉**,支援 UTF-8/CJK(見 `02/04`)|

## 2. 分層移植順序(風險由低到高)

1. **純資料層(無平台耦合)先移**:LBX 解碼器、存檔 schema、資料枚舉/常數。可寫純 Go + 單元測試,不需畫面就能驗(用 1.31 .lbx / 存檔當測資)。
2. **ebiten backend**:實作 `Screen` 對應(繪圖 + 事件),先能開視窗、清屏、畫一張 texture。
3. **UI framework 翻譯**:`gui.cpp` widget 樹 → Go。callback 物件改成 Go closure/interface。
4. **文字系統(全新)**:`text/v2` + CJK 字型 + i18n `lang.Get`(見 `02`)。
5. **畫面重建**:主選單 → 星系圖 →…,參考 openorion2 佈局座標。
6. **gameplay 引擎(最大工作量,從手冊建)**:回合、經濟、科技、戰鬥、AI。

> 前 5 步是「把檢視器搬到 ebiten」,第 6 步才是「做成遊戲」。兩者工作量不同量級,PLAN 要分清。

## 3. 慣用法轉換注意

- C++ 手動 texture 生命週期 → Go GC;`ebiten.Image` 不必手動釋放,但大量小圖建議做 atlas / 快取。
- C++ callback template(`GuiMethodCallback`)→ Go 可用 `func()` 或小 interface,程式更短。
- palette 上色:openorion2 解出 `uint32 xRGB` buffer → `image.RGBA` → `ebiten.NewImageFromImage`。
- 效能:每幀重建 draw-list 沒問題,但避免每幀 `NewImageFromImage`(貴)——資產解一次快取成 `*ebiten.Image`。

## 4. 開發/驗證環境(對齊本機授權)

- **編譯一律 docker**(CLAUDE.md 授權;ebiten 需 GL/X 依賴,容器要備妥)。
- **測試在背景 docker + xvfb**;截圖用 `import -window root` 後 Read 圖逐屏校對(承襲 moo1 §5 驗證紀律)。
- 純資料層用 `go test` 直接驗,不需畫面。

## 5. 待辦 / 待驗證

- [ ] 建最小 ebiten 專案:docker build + xvfb 跑起來、開視窗、畫一張從 .lbx 解出的圖(打通資料層→畫面)。
- [ ] 確認 ebiten 在 headless docker + xvfb 的截圖流程(GL 後端 or `EBITENGINE_GRAPHICS_LIBRARY`)。
- [ ] 取得 `~/master-of-magic` go/ebiten 參考後,回填其在 CJK/字型/打包上的實戰心得。
