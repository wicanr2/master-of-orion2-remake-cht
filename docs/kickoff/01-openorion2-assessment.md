# openorion2 完成度盤點(ground truth)

> 目的:決定「基於 openorion2 改寫成 go/ebiten」時,哪些層可複用、哪些須從頭建。
> 全部 ground 在實際程式碼(`~/moo2/openorion2/src`,約 23,776 行 C++/SDL2)。
> 取得方式:openorion2 為 GPL 開源專案,本機當參考基底,不重新散布(見 `.gitignore`)。

## 一句話結論

openorion2 的自述「**partial savegame viewer, no gameplay**」經查證屬實。它是「原版資產解析 + 唯讀存檔檢視器」,
**沒有任何遊戲規則模擬**。給 port 專案的最大禮物是 **LBX 資產解碼器 + 完整存檔資料 schema**(近乎可直接複用);
UI 框架邏輯值得翻譯;SDL 與 bitmap 字型層要換掉;**真正的「遊戲」——所有規則模擬——必須從原版手冊 + 逆向從零打造**。

## 可複用 vs 需重建(總表)

| 模組 | 檔案:行 | 判定 | 理由 |
|---|---|---|---|
| LBX 容器解析 | `lbx.cpp:169-223` | ✅ 直接照抄 | 純資料,無平台耦合(magic `0xfead`,offset 表相減得 size) |
| RLE/palette/影像解碼 | `gfx.cpp:290-531,588-765` | ✅ 直接照抄 | 演算法級;僅結尾 texture 上傳換 ebiten |
| 存檔 schema + 解析 | `gamestate.cpp/h` 全 | ✅✅ 高價值 | 完整 MOO2 資料模型逆向(見下) |
| 資料枚舉/常數字典 | `gamestate.h:118-760` | ✅ 直接照抄 | 技術≈200 項/建築/種族特性字典 |
| 唯讀衍生公式 | `gamestate.cpp:850-990` 等 | ✅ 參考 | 部分規則數值/公式(艦艇戰力、研究成本、行星產出) |
| UI framework 邏輯 | `gui.cpp/h`(≈2,000 行) | ⚠️ 翻譯移植 | retained-mode widget 樹;callback 改 Go closure |
| 各檢視畫面佈局 | `galaxy/info/tech/officer/ships.cpp` | ⚠️ 參考佈局 | 座標/資產 id 可抄,繪製呼叫要改 |
| 字型渲染 | `gfx.cpp:863-1224` | ❌ 建議重建 | **bitmap 單 byte glyph,結構上無法顯示 CJK** |
| SDL 繪圖後端 | `sdl_screen.cpp`(480 行) | ❌ 換 ebiten | 已被 `screen.h` 抽象介面隔離乾淨 |
| 事件迴圈 | `sdl_events.cpp` | ❌ 換 ebiten | 改 Update/Draw 模型;**原版只處理滑鼠,無鍵盤** |
| 音樂/音效 | (無) | ❌ 從零 | README 宣稱依賴 SDL2_mixer,但全 codebase 無任何 `Mix_` 呼叫,完全未實作 |
| **Gameplay(回合/戰鬥/經濟/科技/AI)** | (無) | ❌❌ **從手冊全建** | **專案最大工作量所在** |

## 1. LBX 資產格式(最該複用)✅

- 容器:`uint16 assetCount` + `uint16 magic(0xfead)`,跳 4 bytes,一串 `uint32` offset 表,相鄰相減得各 asset size(`lbx.cpp:182-200`)。`loadAsset(id)` seek 讀出 `MemoryReadStream`。
- 資產類型:影像(多幀 + 多 palette,`gfx.cpp:326`)、Bitmap(8-bit indexed,`gfx.cpp:588`)、Palette(6-bit VGA `<<2` 成 8-bit,`gfx.cpp:290`)、bitmap 字型(`fonts.lbx`,`gfx.cpp:1053`)、字串(見 §6)。
- **RLE 解碼** `Image::decodeFrame`(`gfx.cpp:476-531`):scan-line RLE,每行讀 size/skip,`size==0` 跳行,否則 skip 透明 + size 實色,奇數補 1 對齊。字型 glyph 另有 run-length(`gfx.cpp:1010-1051`)。
- 與後端唯一耦合:`gameScreen->registerTexture(...)`(`gfx.cpp:442`)→ 換成 `ebiten.NewImageFromImage`。

## 2. 存檔解析(完整資料 schema)✅✅

`GameState::load`(`gamestate.cpp:1844`)依序讀:GameConfig → Galaxy+Nebula → **Colony×250** → **Planet×360** → **Star×72** → **Leader×67** → **Player×8** → **Ship×500(內嵌 ShipDesign)** → 動態組 Fleet。
每欄位順序、magic(如 `LEADER_DATA_SIZE=59`)確定,Go `encoding/binary` LittleEndian 可逐欄重寫得相容讀取器。
含少量 `// FIXME: analyze` 未知欄位(逆向未 100%,但主體完整)。**這等於把 MOO2 完整資料模型逆向好了。**

## 3. GUI/繪圖框架(邏輯可複用,SDL 要換)⚠️

- **平台無關 UI 邏輯**:`gui.cpp/h` —— Widget 樹(Toggle/Choice/ScrollBar/Label/Composite)、callback 物件、視窗/視圖堆疊 ViewStack、TextLayout。可照結構移植,callback 改 Go func。
- **SDL 綁定(隔離乾淨)**:`screen.h` 定義**抽象 `Screen` 介面**;`sdl_screen.cpp` 是唯一 SDL 實作,`sdl_events.cpp` 是事件迴圈。**port 基本只需重寫這兩檔(約 600 行)成 ebiten backend**,`gui.cpp` 那 2,000 行 UI 邏輯可直接翻譯。
- 邏輯座標固定 640×480(`screen.h:26`),`SDL_RenderSetLogicalSize` 縮放 → ebiten `SetWindowSize`/layout 對應。

## 4. 已實作畫面(約半數空殼)⚠️

- **可運作**:主選單、載入存檔(掃 `save?.gam`)、星系圖(`galaxy.cpp` 3,034 行最完整)、行星清單、選玩家、艦隊清單、科技/研究檢視、軍官清單、種族資訊、通用對話框。
- **空殼 STUB**:New Game(彈「Not implemented」)、Multiplayer、Scoreboard、End-of-Turn、Save、Colonies、Races、Zoom、艦隊操作(Relocate/Scrap)、軍官 Hire/Dismiss、研究下訂單。
- 一句話:**能「看」存檔各面向,任何「改變狀態」的操作都是空的。**

## 5. Gameplay 規則(幾乎完全沒有)❌

grep 全 codebase:**沒有**回合處理、戰鬥、經濟、科技推進、AI。唯一存在的是「唯讀衍生屬性計算」(從存檔算顯示值,不改狀態):艦艇戰力(查表 `computerHPTable`)、研究成本、行星產出、軍官雇用費。這些「規則常數表 + 公式」對重建有參考價值,但**距可玩極遠**。整個回合制模擬引擎必須從手冊 + 逆向從零重建。

## 6. 文字/在地化(字串從 LBX,字型是瓶頸)⚠️

- 字串**全從 .lbx 讀,幾乎無硬編碼**。`TextManager`(`lbx.cpp:385-762`)管理主文字/事件/種族/艦名/星名/科技/外交/help;UI 走 `gameLang->hstrings(...)`。
- **現成 i18n 骨架**:內建 5 語言(英德法西義),各對應不同 .lbx 檔名陣列,`selectLanguage()` 重載 TextManager+FontManager。
- **中文化核心障礙**:字型是 `fonts.lbx` 的 bitmap glyph,`Font` 以 **byte 值(0-255)當 glyph 索引**(`gfx.cpp:906-916`)→ 單 byte、最多 256 字,**結構上無法顯示 CJK**。
- → 這反而是**機會**:port 時本來就要丟棄整套 bitmap font,改 ebiten `text/v2` + TTF CJK(見 `02/04`);`TextManager` 結構可留,資料來源改外部 UTF-8 檔放中文。
