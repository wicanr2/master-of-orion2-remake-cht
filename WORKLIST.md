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
- [x] Bitmap(8-bit indexed):像素編碼與 Image 相同(image.go 已涵蓋);dirty-block 為 SDL 局部 blit 優化,ebiten 全繪不需,刻意不移植(見 docs/tech/lbx-format.md §2.7)
- [x] 存檔 schema(對照 gamestate.cpp,全部完成並驗證):
  - [x] reader + GameConfig(59B)+ Galaxy/Nebula(32B)
  - [x] Colony×250 / Planet×360 / Star×72 / Leader×67 / Player×8(內嵌 ShipDesign/Weapon/Settler)/ Ship×500
  - [x] 全區段解析驗證:SAVE10.GAM 解出種族 Trilarian/Alkari/Mrrshan/Sakkra/Klackon、首星 Orion、計數全合理、SeqEnd 收斂(203596,合成全零檔同值當回歸護欄)
- [x] 資料枚舉/常數字典(技術/建築/種族特性/氣候/礦產/特殊裝備)— internal/gamedata/enums.go(28 枚舉,自 gamestate.h 生成)+ docs/tech/enums.md + 抽查測試
- [x] 唯讀衍生公式移植(艦艇戰力/HP/戰速、行星產出、雇用費)— internal/gamedata/formulas.go + docs/tech/formulas.md + 測試(researchCost 待 LBX 資料表)
- [x] 檔案覆蓋順序載入(基礎 → 1.31)— internal/assets/resolver.go(有序搜尋路徑、大小寫不敏感、OpenLBX)+ 測試
- [x] 單元測試:lbx/save/gamedata/assets 皆有合成測試;lbx/save 另有 env-gated 真實檔測試(MOO2_LBX_TEST / MOO2_SAVE_TEST)

## Phase 2 — ebiten backend + 最小可跑 ⭐
- [x] ebiten 專案骨架(Update/Draw/Layout)— cmd/moo2 + ebiten v2.9.9
- [x] palette 上色 → `ebiten.Image`(Frame.ToRGBA → NewImageFromImage → DrawImage)
- [x] docker + xvfb 截圖流程打通 — docker/Dockerfile.ebiten(CGO+X11+GL+xvfb)+ scripts/screenshot.sh(ReadPixels 存 PNG,不依賴 WM)
- [x] ★ 顏色視覺驗證:MAINMENU 資產 21 於 ebiten 渲染出完整正確主選單(640×480)
- [x] 確認 MOO2 為 640×480(非 320×200);修正 kickoff 假設
- [ ] 實作 `Screen` 對應:registerTexture/drawTexture/fillRect/setClipRegion(抽介面,目前為直繪骨架)
- [ ] 滑鼠事件(cursor + 按鍵),補鍵盤
- [ ] 資產快取(避免每幀 NewImageFromImage)
- [x] ★ 里程碑 M2:載存檔 → 繪製星系圖(cmd/moo2 -save;SAVE10.GAM 的 36 星依座標/光譜/大小 + 星名 + 星雲,資料驅動)
- [ ] 星圖換真實星球 sprite(GALAXY.LBX asset 148,依 spectralClass×size)+ 星空背景(STARBG.LBX)

## Phase 3 — UI 框架 + 文字系統 + 主選單(做法見 `08` playbook)
- [ ] gui widget 樹翻譯(Toggle/Choice/ScrollBar/Label/Composite + ViewStack)
- [ ] callback → Go closure/interface
- [x] CJK 渲染:`internal/uifont`(ebiten text/v2,依尺寸快取 face;text/v2 原生向量 rasterize 取代手動 supersample)+ Measure
- [x] 顯示層覆蓋 i18n:`internal/i18n`(TSV 英文即 key + 查無 fallback + TranslateFormat)+ 測試
- [x] [HARD] 只翻顯示層,不動資料層(i18n 設計即如此)
- [x] 字型:NotoSansCJK-Regular.ttc 經 Go opentype.ParseCollection 驗證可解析+量測中文(★ [HARD] 相容檢查通過);galaxy 標題已渲染繁中
- [ ] 繪字描邊/陰影版 + 逐字斷行(目前基本 Draw/Measure;進階待用到時補)
- [ ] 字型子集 pyftsubset(docker)+ go:embed 內嵌(待譯文集齊;目前用完整 .ttc runtime 掛載)
- [x] 主選單中文化 + 截圖校對(cmd/moo2 -menu:擦底疊字六按鈕繼續/載入遊戲/…;before/after 見 docs/reference-screens.md)
- [ ] 主選單:語言 中/英 runtime 切換(mom 無此,我們要做)
- [ ] 主選單:版本 1.3/1.5 選擇框架
- [ ] 按鈕垂直置中微調(目前略偏上)+ hover 狀態中文

## Phase 4 — 畫面重建 + 完整中文化(做法見 `08` playbook)
- [x] 原版畫面對照組(`docs/reference-screens.md`:主選單/行星列表/建造,英文原貌 + 翻譯清單)
- [x] 通用畫面覆蓋渲染器(`cmd/moo2/overlay.go`:資料驅動擦底疊字,選單+行星列表共用)
- [x] 主選單中文化(6 按鈕)+ 行星列表中文化(18 標籤,before/after)
- [x] LBX 字串資源解析 + dumper(`internal/lbx/strings.go` + `cmd/lbxstrings`);TECHNAME 560 條科技名 dump 成功
- [x] **科技/元件名譯表完整(`assets/i18n/tech.tsv`:419 條唯一全翻)** — 研究主題/領域、武器/裝甲/護盾/引擎/電腦、建築、艦種、武器改造(含縮寫);覆蓋驗證 419/419 無遺漏
- [x] i18n TSV 守護測試(載入所有 assets/i18n/*.tsv + 佔位符一致性)
- [~] 擦底疊字改善(fill 加高;darkest 採樣反而過暗已還原)。「顯示篩選」寬粗英文仍微透,需整圖替換或更寬擦除(降級 todo)
- [ ] 其餘字串源逐一 dump + 翻:科技描述(TECHDESC)、種族類型名、事件、外交、艦名、星名、help
- [x] **★ 調色盤鏈解鎖(關鍵)**:對照 openorion2 `gfx.cpp Image::load` 破解「無內嵌調色盤畫面」上色機制(基底提供圖 + 本圖部分內嵌疊加);實作 `cmd/moo2/interactive.go` `resolvePalette`;研究選擇(TECHSEL,借 SCIENCE 調色盤)完整渲染驗證。見 `docs/tech/palette-chain.md`
- [ ] 依 `palette-chain.md` 對照表逐畫面上色:COLONY(COLONY2 50)/DESIGN/FLEET(FLEETLIST)/INFO/星系 GUI(GAMEMENU)… 提供圖 index 逐一對照 openorion2 建構子(勿憑記憶)

## Phase 4b — 串接互動(還原原版的骨幹,-game)⭐
> 各原版畫面不再各自獨立 flag,而是串成單一可導覽的互動程式(`cmd/moo2 -game`)。目標:開機進原版主選單,滑鼠點選在原版畫面間跳轉,全繁中。
- [x] 互動骨架:`origScreen`/`origTransition` 介面 + `overlayScreen`(真 LBX 背景 + 中文擦底疊字 + 點擊熱區)+ `sceneBuilder` + `interactiveApp`(ebiten.Game,支援 headless 腳本驗證)
- [x] 導覽:原版主選單(真美術)→「新遊戲/繼續」→ 真原版行星列表 →「返回」→ 主選單(headless 驗證通過)
- [x] 調色盤鏈畫面併入導覽 + 小於全螢幕視窗置中
- [x] 研究選擇畫面**完整中文化**(擦底疊字,PIL 量測校對,完整垂直切片)
- [x] 調色盤鏈擴充多段鏈(`paletteChain []assetRef`;艦隊三段鏈驗證)
- [x] **★ 星系主樞紐(galaxy GUI,BUFFER0.LBX 0)接成遊戲主畫面**:新遊戲→星系主畫面,
  底部工具列(座標取自 galaxy.cpp)導覽到 行星/艦隊(FLEET)/軍官(OFFICER)/科技總覽(INFO);
  各畫面返回樞紐。全部忠實原版美術,headless 驗證導覽鏈通過
- [x] 星系工具列中文化(殖民地/行星/艦隊/領袖/種族/情報/回合)
- [x] 艦隊列表中文化(艦隊作戰/全部/調動/拆解/軍官/支援/戰鬥/返回)
- [x] 軍官列表中文化(殖民地領袖/艦艇軍官/雇用/人才庫/解雇/返回)
- [x] 科技總覽中文化(星曆/歷史圖表/科技總覽/種族統計/回合摘要/參考資料/返回)
- [x] 擦底採樣穩健化:samplePlate 左緣帶+上下橫帶眾數;背景均勻畫面(info)改 overlayScreen.eraseColor 強制底色
- [ ] 殘留細修:galaxy 工具列 GAME 標題未翻;行星/艦隊個別按鈕邊緣極微殘(紋理按鈕固有)
- [ ] 各子畫面 RETURN 按鈕精確熱區(目前暫用全螢幕返回)
- [ ] 科技總覽左側選單各項可點進子畫面(歷史圖表/科技總覽/種族統計…)
- [x] 殖民地總覽畫面(COLSUM.LBX 0)接入 COLONIES 按鈕 + 完整中文化
- [x] 種族關係畫面(RACES.LBX 0)接入 RACES 按鈕 + 中文化(種族關係/會晤/報告/宣戰/忽略/加成/返回)
- [x] **★ 真新遊戲流程**:主選單→新遊戲→原版 NEW GAME 設定畫面(NEWGAME.LBX 28,調色盤鏈 RACEOPT#4→NEWGAME#1)→ACCEPT→星系主畫面;中文化(難度/星系大小/星系年齡/玩家數/科技等級/戰術戰鬥/隨機事件/安塔蘭攻擊/取消/接受)
- [x] 回合摘要畫面(TURNSUM.LBX#0)接入 TURN 流程(結束回合→摘要顯示本回合結算:星曆/淨工業/研究/食物/稅收/國庫變化/研究完成)→關閉回星系。中文化(回合摘要/關閉)
- [x] 已探測定位背景(remain-scan,待接入):議會 COUNCIL.LBX#1、艦艇設計 DESIGN.LBX#0、讀取存檔 LOADSAVE.LBX#11、外交 DIPLOMAT.LBX#29(有雜訊待查)
- [ ] 細修:NEW GAME 開關列/標題微殘、種族關係 ESPIONAGE/SABOTAGE/HIDE(24 標籤)未翻、各畫面按鈕精確熱區
- [x] **★ 核心遊戲迴圈第一步**:GameSession 接進 -game;TURN 按鈕呼叫 session.EndTurn()
  (結算帝國經濟 + AI 對手決策),星系畫面即時顯示星曆(3500 起,每回合+1年)+ 國庫 BC
  (overlayScreen.extras 動態文字機制)。驗證:TURN×2 → 星曆 3500→3502、國庫 100→106
- [ ] 待接入畫面:議會/艦艇設計/回合摘要/讀存檔(背景已備,需入口點+中文化);單一殖民地管理(原版走動態繪製,無預烘背景,需自建)
- [x] 殖民地總覽填即時資料:玩家各殖民地列出「殖民地 N / 農夫 / 工人 / 科學家」(來自 GameSession,對齊原版欄位,extras 動態文字)
- [ ] 其餘畫面填即時資料(行星列表、艦隊、軍官…);colony 名稱改用真星名;建造欄填建造佇列
- [ ] 星圖視窗渲染真實星系(galaxy.go -save 星球繪製移植)
- [ ] 真正的「新遊戲」流程(種族選擇/星系生成)取代目前的示範跳轉
- [ ] hover highlight 與原版一致(目前為細框提示)
- [ ] 淘汰自製簡約殼(`-play`):方向不符「與原版一模一樣」,改以原版 overlay 畫面 + 既有回合引擎(internal/engine)重建可玩迴圈
- [ ] 補齊需全域調色盤鏈的畫面(COLONY/DESIGN/COUNCIL/DIPLOMAT…)到對照組
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
- [x] 遊戲歷史與當年評價考究(`docs/history/moo2-history-and-reception.md`,角色:歷史考究專家,14 來源)
- [x] GitHub 致謝(README:openorion2/1oom/mom/字型/社群/Simtex)
- [x] 技術知識庫:LBX 資產格式 / 存檔格式 / 枚舉 / 公式 / ebiten 移植筆記(`docs/tech/`)
- [x] 華人圈中文討論資訊考究章節(`docs/history/moo2-chinese-community.md`,歷史考究專家,31 來源+誠實揭露侷限)
- [x] 華人圈文化現象(`docs/culture/moo2-chinese-cultural-phenomenon.md`,文案作家,事實有本、無 AI 味)
- [ ] sprite/tile 畫質優化可行性 markdown
- [ ] UI 界面調整可行性 markdown
- [ ] 技術知識庫:音樂整合 / 鍵盤滑鼠整合 / patch 處理 / 選單擴展(後續各 Phase 完成時補)
- [x] 三平台打包 CI(`docs/tech/packaging.md`):macOS(`.github/workflows/build-macos.yml`,`macos-14` runner 原生編 arm64+amd64 → `lipo` universal → `.app`/`.dmg`/`.tar.gz`)+ Linux/Windows(`.github/workflows/build-desktop.yml`);YAML 經 actionlint + yaml.safe_load 驗證,尚未在真 Mac 上實跑驗證(無 Mac 測試機)
- [x] 本機 docker 打包腳本(`docs/tech/packaging.md` §5):`scripts/package-appimage.sh`(Linux AppImage,linuxdeploy+appimagetool)、`scripts/package-windows.sh`(Windows zip)已實際跑過,`dist/MasterOfOrion2-cht-x86_64.AppImage`、`dist/MasterOfOrion2-cht-windows-amd64.zip` 皆產出並驗證內容(解壓/objdump 確認)。**推翻先前假設**:ebiten v2.9.9 Windows backend 已改純 Go(purego,無 cgo),`CGO_ENABLED=0` 即可跨編,不需 mingw-w64(`build-desktop.yml` 仍裝了 mingw,屬保守多餘,非錯誤,可留後續簡化)
- [ ] `cmd/moo2` 加可覆寫 assets/i18n 路徑(或 go:embed)取代相對路徑假設,讓 macOS `.app` 不需 launcher script 繞路(見 packaging.md §4 待辦)

## 工作方式(使用者定案)
- go/ebiten 參考路徑 = `~/master-of-maigc/repo`(魔法大帝繁中化,patch 疊 kazzmir/master-of-magic 引擎)
- **不用多代理 workflow**;翻譯一組一組慢慢做(單代理逐項,使用者可隨時審閱)
- 每輪更新 GitHub(遠端 `main`,已設 upstream)
