# WORKLIST — 銀河霸主2 go/ebiten 重製 + 繁中化

> 可勾選工作清單,對應 `PLAN.md` 階段。允許擴充(CLAUDE.md)。完整性優先:不預先砍項;卡關記錄方法換路,不寫「暫緩/低投報」。
> 圖例:`[ ]` 待辦 `[~]` 進行中 `[x]` 完成。

## Phase 0 — Kick-off / 可行性(本輪)
- [x] 盤點 openorion2 完成度(`docs/kickoff/01`)
- [x] 中文化策略(`02`)
- [x] 按鈕中文化策略,參考 moo1 避免重蹈覆轍(`03`)
- [x] 字型選擇研究(`04`)
- [x] LBX 資產 + patch 1.3/1.5 處理與版本架構(`05`)
- [x] ebiten 移植策略(`06`)
- [x] 可行性總論(`00`)
- [x] PLAN.md / WORKLIST.md
- [x] .gitignore(擋版權素材)
- [ ] README(含致謝)
- [ ] 本機 git commit(push 待使用者確認)

## Phase 1 — 資料層移植(純 Go)
- [ ] Go module 初始化 + docker build 環境
- [ ] LBX 容器解析(magic 0xfead、offset 表)
- [ ] scan-line RLE 影像解碼
- [ ] palette 解析(6-bit → 8-bit)
- [ ] 影像多幀 + 多 palette variant
- [ ] Bitmap(8-bit indexed)+ dirty block
- [ ] 存檔 schema:GameConfig/Galaxy/Colony×250/Planet×360/Star×72/Leader×67/Player×8/Ship×500
- [ ] 資料枚舉/常數字典(技術/建築/種族特性/氣候/礦產/特殊裝備)
- [ ] 唯讀衍生公式移植(艦艇戰力/研究成本/行星產出/雇用費)
- [ ] 檔案覆蓋順序載入(基礎 → 1.31)
- [ ] 單元測試:以 1.31 .lbx + 存檔為測資

## Phase 2 — ebiten backend + 最小可跑 ⭐
- [ ] ebiten 專案骨架(Update/Draw/Layout)
- [ ] 實作 `Screen` 對應:registerTexture/drawTexture/fillRect/setClipRegion
- [ ] 滑鼠事件(cursor + 按鍵),補鍵盤
- [ ] palette 上色 → `ebiten.Image` + 資產快取
- [ ] docker + xvfb 截圖流程打通
- [ ] 里程碑:開視窗 → 讀 .lbx → 載存檔 → 顯示星系圖

## Phase 3 — UI 框架 + 文字系統 + 主選單
- [ ] gui widget 樹翻譯(Toggle/Choice/ScrollBar/Label/Composite + ViewStack)
- [ ] callback → Go closure/interface
- [ ] 文字系統:ebiten text/v2 + TTF
- [ ] i18n 模組 `lang.Get(key)`,en/zh
- [ ] 字型定案(Cubic 11/Fusion Pixel + Noto 保底)+ 缺字掃描
- [ ] 主選單:語言 中/英 切換
- [ ] 主選單:版本 1.3/1.5 選擇框架
- [ ] 主選單中文化 + 截圖校對

## Phase 4 — 畫面重建 + 完整中文化
- [ ] 逐畫面重建:主選單/載存檔/星系圖/行星清單/殖民地/科技研究/艦隊/軍官/種族資訊/對話框
- [ ] 盤點各畫面「烘進 gfx 的按鈕/標籤」清單
- [ ] 按鈕中文化(拆背景+標籤兩層,依 `03`)
- [ ] LBX 字串譯文表:科技名/描述、種族、事件、外交、艦名、星名、help
- [ ] 專有名詞術語表(統一譯名,對齊 moo1 經驗)
- [ ] 每畫面 xvfb 截圖校對(破版/溢出/缺字/置中)

## Phase 5 — Gameplay 引擎重建
- [ ] 回合結算主迴圈
- [ ] 殖民地經濟:人口成長/食物/工業/研究/污染
- [ ] 建造佇列與建築效果
- [ ] 科技研究樹推進
- [ ] 艦隊移動 + 星圖導航
- [ ] 艦艇設計
- [ ] 戰術戰鬥
- [ ] 外交
- [ ] 隨機事件 / 安塔蘭人
- [ ] AI 對手
- [ ] 開新遊戲流程(取代 STUB)
- [ ] 以手冊逐系統對照驗證規則正確性

## Phase 6 — 音樂 / 音效
- [ ] 逆向 .lbx 音樂(XMI)格式
- [ ] 逆向音效格式
- [ ] ebiten 音訊播放整合
- [ ] SoundFont 處理(承襲 moo1 音樂音色經驗)

## Phase 7 — 版本 1.3 / 1.5
- [ ] 研究「1.3 → 1.5 規則差異清單」(手冊×2 + CHANGELOG_150 + PARAMETERS.CFG 逐條)
- [ ] rule profile 資料結構設計
- [ ] 1.3 profile 實作 + 驗證
- [ ] 1.5 profile 實作 + 驗證
- [ ] 主選單版本切換生效(規則 + 資產一起換)

## Phase 8 — 文件 / 考究 / 文化 / 研究
- [ ] GitHub 致謝(openorion2/1oom 社群/字型作者/社群考據)
- [ ] 中文討論資訊考究章節(角色:遊戲歷史考究專家)
- [ ] 華人圈文化現象(角色:文案作家)
- [ ] sprite/tile 畫質優化可行性 markdown
- [ ] UI 界面調整可行性 markdown
- [ ] 技術知識庫:ebiten porting 心得
- [ ] 技術知識庫:音樂整合
- [ ] 技術知識庫:鍵盤/滑鼠整合
- [ ] 技術知識庫:patch 處理
- [ ] 技術知識庫:選單擴展
- [ ] 取得 `~/master-of-magic` 參考後回填 ebiten 實戰心得

## 待澄清(等使用者)
- [ ] `~/master-of-magic`(go/ebiten 參考)正確路徑
- [ ] 是否啟用多代理 workflow(ultracode)加速後續輪次
