# PLAN — 銀河霸主2 go/ebiten 重製 + 繁中化

> 依 `docs/kickoff/00-feasibility.md` 的可行性結論制定。核心認知:**這是「移植檢視器 + 從零重建遊戲引擎」兩件量級不同的事**,計畫據此分軌。
> 完整性優先(rulebook 83):各階段工作不因投報預先砍;難只改方法與時間。
> 可勾選的細項見 `WORKLIST.md`。本檔給階段輪廓與依賴。

## 指導原則

1. **先交付可跑的檢視器,建立技術信心**;再逐系統長出遊戲引擎。
2. **每輪都更新 GitHub repo**(patch/原始碼 + 文件),worklist 允許擴充。
3. **正確性 > 可落地 > 時程 > 可維護 > 效能**;規則以手冊為權威,疑點用第一性原理查證,不臆測。
4. **repo 只放我們的碼 + 文件 + 譯文 + 字型**,不含版權遊戲資產(玩家自備正版)。
5. 編譯/測試一律 docker;每畫面改動用 xvfb 截圖逐屏校對。

## 階段

### Phase 0 — Kick-off 與可行性(本輪,✅ 完成)
- 盤點 openorion2 完成度、中文化策略、字型、按鈕、LBX/patch、ebiten 移植。
- 產出 `docs/kickoff/00~06`、PLAN、WORKLIST、README(致謝)。

### Phase 1 — 資料層移植(純 Go,無畫面,可單元測試)
- 依賴:Phase 0。風險最低,先做。
- LBX 解碼器(容器 + scan-line RLE + palette)Go 重寫,對 1.31 .lbx 驗證。
- 存檔 schema 用 `encoding/binary` 逐欄重寫,能載入原版存檔。
- 資料枚舉/常數字典移植。
- 檔案覆蓋順序載入(基礎 → 1.31)。

### Phase 2 — ebiten backend + 最小可跑里程碑 ⭐
- 依賴:Phase 1。**第一個可交付里程碑**:開視窗 → 讀玩家正版 .lbx → 載入存檔 → 顯示星系圖。
- 實作滿足 openorion2 `Screen` 抽象介面的 ebiten backend(繪圖 + 滑鼠 + 鍵盤)。
- docker + xvfb 截圖流程打通。

### Phase 3 — UI 框架 + 文字系統 + 主選單(版本/語言選擇)
> 中文化技術路線採 mom 已驗證 playbook,見 `08-mom-ebiten-cht-playbook.md`。
- 依賴:Phase 2。
- `gui.cpp` widget 樹翻譯成 Go(callback → closure)。
- 全新文字系統:supersample 4× CJK glyph + 顯示層覆蓋 i18n(英文原文即 key,中/英 runtime 切換)。
- 字型:先 Noto Sans TC 打通,像素字型待驗 Go 解析後 A/B(見 `04`)。
- 主選單加「版本 1.3/1.5」「語言 中/英」選擇框架。

### Phase 4 — 畫面重建 + 完整中文化(含按鈕)
- 依賴:Phase 3。
- **開工先窮舉所有文字源(LBX 各類 + Go hardcode)並各寫 dumper**(漏一源 = 整類英文,見 `08` §7)。
- 逐畫面重建(主選單 → 星系圖 → 殖民地 → 科技 → 艦隊 → 軍官 → 種族…),參考 openorion2 佈局。
- 按鈕/烘字中文化依 `03`(擦底疊字 or 整圖替換 + IMGLOG 探查)。
- LBX 字串譯文表(逐源分檔 TSV,組合字串走 `TranslateFormat`),分批翻譯。
- 每畫面 xvfb 截圖校對破版/溢出/缺字。

### Phase 5 — Gameplay 引擎重建(最大工作量)
- 依賴:Phase 1(資料模型)。可與 Phase 3/4 部分並行。
- 從手冊逐系統實作:回合結算 → 殖民地經濟(人口/食物/工業/污染)→ 科技研究 → 建造 → 艦隊移動 → 戰術戰鬥 → 外交 → 事件 → AI。
- 以 openorion2 既有唯讀公式常數交叉驗證數值正確性。
- 把 openorion2 的 STUB(End-of-Turn/New Game/研究下訂…)換成真實邏輯。

### Phase 6 — 音樂 / 音效
- 依賴:Phase 1。逆向 .lbx 內音樂(XMI)/音效格式並播放(openorion2 無實作,從零)。

### Phase 7 — 版本 1.3 / 1.5 規則差異
- 依賴:Phase 5。
- 研究並列出「1.3 → 1.5 規則差異清單」(兩本手冊 + CHANGELOG_150 + PARAMETERS.CFG)。
- 設計 rule profile,讓同一引擎跑兩版規則;主選單切換生效。

### Phase 8 — 文件 / 考究 / 文化 / 畫質 & UI 研究
- 可與各階段並行的文件工作:
  - GitHub 致謝(openorion2、1oom 社群、字型作者、社群考據)。
  - 中文討論資訊考究章節(角色:遊戲歷史考究專家)。
  - 華人圈文化現象(角色:文案作家)。
  - sprite/tile 畫質優化可行性 markdown。
  - UI 界面調整可行性 markdown。
  - 技術知識庫:ebiten porting 心得、音樂整合、鍵鼠整合、patch 處理、選單擴展。

## 里程碑摘要

| M | 內容 | 對應 Phase |
|---|---|---|
| M1 | 資料層可讀 .lbx + 存檔(go test 綠) | 1 |
| M2 ⭐ | ebiten 顯示星系圖(檢視器上線) | 2 |
| M3 | 中/英切換 + 版本選擇框架 + 主選單中文化 | 3 |
| M4 | 主要畫面全中文化(含按鈕) | 4 |
| M5 | 可實際玩一回合(經濟+科技+建造結算) | 5 |
| M6 | 完整可玩(含戰鬥/外交/AI) | 5 |
| M7 | 1.3/1.5 雙版本 + 音樂 | 6,7 |
| M8 | 文件/考究/文化/畫質&UI 研究齊備 | 8 |
