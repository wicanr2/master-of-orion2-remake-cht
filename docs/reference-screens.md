# 原版畫面對照組(中文化 before/after 基準)

蒐集原版《Master of Orion II》各畫面的英文原貌,作為中文化成果的對照基準:每個畫面日後配上我們的繁中版並排,即成 before → after。同時,各畫面可見的英文 UI 字串就是該畫面的**翻譯清單**。

畫面由本專案的 `cmd/lbxdump` 從玩家正版 `.lbx` 解碼渲染(非外部截圖)。

> **本文只是 before/after 展示用的靜態對照組,收錄進度落後於實作很多。**
> 遊戲本身的畫面現況請看 `-gamegallery` 產出的截圖廊(`cmd/moo2/interactive.go` 的
> `buildGalleryScript`),或直接看 `cmd/moo2/*.go`。

## 已收錄

### 主選單(MAINMENU.LBX 資產 21)✅ 已中文化

| 原版英文 | 繁中化(擦底疊字) |
|---|---|
| ![](images/reference/en-main-menu.png) | ![](images/reference/cht-main-menu.png) |

英文 UI(對應 `assets/i18n/menu.json`,已翻):Continue、Load Game、New Game、Multi Player、Hall of Fame、Quit Game。
手法:六按鈕英文烘在背景圖(New Game 只有 hover sprite、idle 靠背景證實),故用擦底疊字 —— 採樣按鈕底色蓋掉英文 + 疊置中中文(座標取自 openorion2 mainmenu.cpp:415,y=172/195/217/240/262/285)。
重現:`moo2 -menu -data <遊戲夾> -font <CJK字型> -shot out.png`。

### 行星列表(PLNTSUM.LBX 資產 0)✅ 已中文化

| 原版英文 | 繁中化 |
|---|---|
| ![](images/reference/en-planets-list.png) | ![](images/reference/cht-planets-list.png) |

18 個標籤覆蓋(`assets/i18n/planets.json`,座標多取自 openorion2 `PlanetsListView::initWidgets`):
欄位(行星/氣候/重力/礦產/大小)、排序優先 + 排序鈕、顯示篩選 + 5 個篩選(無敵蹤/正常重力/適居環境/礦產豐富/射程內)、派殖民船/派前哨船/返回。
重現:`moo2 -planets -data <遊戲夾> -font <字型> -shot out.png`。
> 已知小瑕疵:「顯示篩選」下方原本又寬又粗的 DISPLAY RESTRICTIONS 以單一採樣色擦不乾淨,邊緣微透 —— 待改用「從空白處採樣底色」精修。

### 殖民地建造(COLBLDG.LBX 資產 0)
![](images/reference/en-colony-build.png)

英文 UI(待翻,建議 `assets/i18n/colony.json`):Auto Build、Refit、Design、Repeat Build、Cancel、OK。

## ~~待補(需全域調色盤鏈,Phase 4)~~ → 調色盤鏈早已完成

⚠ **過期斷言已刪**(2026-08-07)。原文說「殖民地主畫面(COLONY)、艦艇設計(DESIGN)、
殖民地系統顯示(COLSYSDI)、議會(COUNCIL)、外交(DIPLOMAT)、艦隊(FLEET)、科技選擇(TECHSEL)
等畫面需 Phase 4 的全域調色盤鏈才能上色」——**這些畫面在遊戲裡早就跑起來了**。
機制是 `cmd/moo2/interactive.go` 的 `paletteChain` + `resolvePalette`(逐個提供圖疊上去,
目標圖自己的內嵌調色盤最後蓋),`lbxdump` 也有對應的 `--pal <file.lbx>:<idx>`。

唯一仍未定案的是 **COLGCBT(地面戰 sprite)**:它所有資產都沒有內嵌調色盤,原版是沿用當時
殖民地畫面的調色盤,remake 借 `COLBLDG.LBX#0`,渲染合理但未證實(見 `cmd/moo2/groundcombat.go` 檔頭)。

## 用途

1. **中文化 before/after 展示**:同畫面英文原版 vs 繁中版並排。
2. **翻譯清單來源**:各畫面英文 UI 字串 → 對應 `assets/i18n/*.json`(英文原文即 key)。
3. **烘字位置參考**:按鈕/標籤在圖上的座標。⚠ 但這只是最後手段——座標的一手來源是
   **原版執行檔的反組譯**(繪製呼叫的立即數),其次是 openorion2 的 `initWidgets`,量圖排最後。
