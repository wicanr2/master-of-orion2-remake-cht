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
- [x] LBX 容器解析(magic 0xfead、offset 表)— internal/lbx,真實檔驗證
- [x] scan-line RLE 影像解碼 — internal/lbx/image.go
- [x] palette 解析(6-bit → 8-bit)— 解碼與上色解耦(Frame.ToRGBA)
- [x] 影像多幀(frame offset 表)+ 多 palette variant(ToRGBA 套不同 palette)
- [ ] Bitmap(8-bit indexed)+ dirty block(gfx.cpp:588-765;image.go 已含 NOCOMPRESS/RLE,Bitmap sparse block 待補)
- [x] 存檔 schema(對照 gamestate.cpp,全部完成並驗證):
  - [x] reader + GameConfig(59B)+ Galaxy/Nebula(32B)
  - [x] Colony×250 / Planet×360 / Star×72 / Leader×67 / Player×8(內嵌 ShipDesign/Weapon/Settler)/ Ship×500
  - [x] 全區段解析驗證:SAVE10.GAM 解出種族 Trilarian/Alkari/Mrrshan/Sakkra/Klackon、首星 Orion、計數全合理、SeqEnd 收斂(203596,合成全零檔同值當回歸護欄)
- [x] 資料枚舉/常數字典(技術/建築/種族特性/氣候/礦產/特殊裝備)— internal/gamedata/enums.go(28 枚舉,自 gamestate.h 生成)+ docs/tech/enums.md + 抽查測試
- [x] 唯讀衍生公式移植(艦艇戰力/HP/戰速、行星產出、雇用費)— internal/gamedata/formulas.go + docs/tech/formulas.md + 測試(researchCost 待 LBX 資料表)
- [x] 檔案覆蓋順序載入(基礎 → 1.31)— internal/assets/resolver.go(有序搜尋路徑、大小寫不敏感、OpenLBX)+ 測試
- [ ] 單元測試:以 1.31 .lbx + 存檔為測資(進行中:lbx/save 皆有合成 + 真實檔測試)

## Phase 2 — ebiten backend + 最小可跑 ⭐
- [ ] ebiten 專案骨架(Update/Draw/Layout)
- [ ] 實作 `Screen` 對應:registerTexture/drawTexture/fillRect/setClipRegion
- [ ] 滑鼠事件(cursor + 按鍵),補鍵盤
- [ ] palette 上色 → `ebiten.Image` + 資產快取
- [ ] docker + xvfb 截圖流程打通
- [ ] 里程碑:開視窗 → 讀 .lbx → 載存檔 → 顯示星系圖

## Phase 3 — UI 框架 + 文字系統 + 主選單(做法見 `08` playbook)
- [ ] gui widget 樹翻譯(Toggle/Choice/ScrollBar/Label/Composite + ViewStack)
- [ ] callback → Go closure/interface
- [ ] CJK 渲染:supersample 4× glyph + `(rune,字高)` 快取 + 對齊呼叫端字高(基線 0.82)
- [ ] 繪字三路徑都支援 CJK(印字/描邊陰影/量寬 MeasureTextWidth)
- [ ] 逐字斷行(CJK 無空白;無斷點至少切一 rune)
- [ ] 顯示層覆蓋 i18n:TSV(英文原文即 key)+ 查無 fallback 英文 + `TranslateFormat` 模板
- [ ] [HARD] 只翻顯示層,不動資料層(避免破壞把英文當 key 的邏輯)
- [ ] 字型:先用 Noto Sans TC 打通;驗證候選字型 Go opentype 解析可行性(CFF/.ttc 有風險)
- [ ] 字型子集 pyftsubset(docker)+ go:embed 內嵌;加字重生子集
- [ ] 主選單:語言 中/英 runtime 切換(mom 無此,我們要做)
- [ ] 主選單:版本 1.3/1.5 選擇框架
- [ ] 主選單中文化 + 截圖校對

## Phase 4 — 畫面重建 + 完整中文化(做法見 `08` playbook)
- [ ] **[HARD] 開工先做:窮舉所有文字源(LBX 各類 + Go hardcode),各寫 dumper,用引擎自己 reader dump 精確 key**
- [ ] 逐畫面重建:主選單/載存檔/星系圖/行星清單/殖民地/科技研究/艦隊/軍官/種族資訊/對話框
- [ ] IMGLOG 探查模式:記錄 `(lbx,index)` 對照畫面 UI(盤點烘字按鈕/標籤用)
- [ ] 烘進 gfx 的英文:擦底疊字(cht_label 模式)or 整圖替換(image_override 模式)
- [ ] LBX 字串譯文表:科技名/描述、種族、事件、外交、艦名、星名、help(逐源分檔 TSV)
- [ ] 組合字串走 `TranslateFormat` 翻模板字面(佔位符數/序中英一致)
- [ ] 專有名詞術語表 + 「中文(英文)」小字控制碼(統一譯名,對齊 moo1/mom 經驗)
- [ ] 每畫面 xvfb + xdotool 導航 + import 截圖校對(破版/溢出/缺字/置中)

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
- [ ] AI 對手(策略見 `docs/kickoff/07-ai-strategy.md`:先參考 1oom `game_ai_classic.c` + GameFAQs 文獻,有必要才逆向)
  - [ ] 精讀 1oom `game_ai_classic.c`,抽「AI 決策流程」語言無關筆記
  - [ ] 精讀 GameFAQs MOO2 AI FAQ + 策略指南,補 MOO2 特有行為
  - [ ] 設計可插拔 AI 介面 + 難度加成係數
  - [ ] 標示「必須逆向才能確定」的項目(若有)
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
- [x] 技術知識庫:LBX 資產格式(`docs/tech/lbx-format.md`)
- [x] 技術知識庫:存檔格式(`docs/tech/savegame-format.md`)
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

## 工作方式(使用者定案)
- go/ebiten 參考路徑 = `~/master-of-maigc/repo`(魔法大帝繁中化,patch 疊 kazzmir/master-of-magic 引擎)
- **不用多代理 workflow**;翻譯一組一組慢慢做(單代理逐項,使用者可隨時審閱)
- 每輪更新 GitHub(遠端 `main`,已設 upstream)
