# WORKLIST — 銀河霸主2 go/ebiten 重製 + 繁中化

> ⚠ **誠實現況(2026-07-04,使用者實測):還原度約 20%,尚不能真的玩。詳見 [`docs/HONEST-STATUS.md`](docs/HONEST-STATUS.md)。**
> 下方大量 `[x]` 是「自製系統 + 單元測試通過」,**不等於對齊原版**。真實驗收要對原版實測(rulebook 65)。
> 最高優先未做:①音樂/音效 ②忠實新遊戲流程(獨立種族選擇+真殖民地初始) ③按鍵逐畫面像素對齊 ④忠實 gameplay 規則。
>
> 可勾選工作清單,對應 `PLAN.md` 階段。允許擴充(CLAUDE.md)。完整性優先:不預先砍項;卡關記錄方法換路,不寫「暫緩/低投報」。
> 圖例:`[ ]` 待辦 `[~]` 進行中 `[x]` 完成(⚠ 多為自製系統的完成,非原版對齊)。

## ★ 2026-07-10 session 進展摘要(接手後,對原版/手冊驗證)
> 本段為快速索引,細節散見各 Phase 與 docs/tech/。

**已完成並截圖/資料驗證:**
- [x] **音訊基礎**:PCM WAV 原封播放 + 主選單 BGM + 按鈕音效(`internal/audio`、`audiohook.go`)
- [x] **研究系統完整忠實化**:真 RP 成本 + 每主題抉擇 UI(真中文名)+ 抉擇解鎖元件(對 openorion2 核實,`research-system-status.md`)
- [x] **獨立種族選擇流程**:13 族肖像 + 自訂點數 + 命名/旗色(`newgame-flow.md`)
- [x] **外交畫面破解重建**:DIPLOMAT.LBX 全破解(13 palette+13 房+13 使節,配對律 r)+ 逐族使節疊合 + 13 族對應核實(`diplomat-lbx-layout.md`)
- [x] **戰術戰鬥換原版美術**:STARBG 星空 + COMBAT 控制列 + CMBTSHP 可見艦艇;太空戰接真命中/傷害/過盾/過甲公式(`ResolveShot`);**控制列 7 按鈕中文化**(自動/掃描/登船/撤退/等待/完成/選項)(`tactical-combat-assets.md`)
- [x] **中文化稽核補漏**:galaxy 工具列 ZOOM→縮放、頂部 GAME→遊戲(擦底疊字)
- [x] **`-gamegallery` 端到端截圖廊**:8 畫面互動 app 內全繁中渲染驗證(修無限迴圈 CPU bug:硬性終止+timeout)

**四 directive 收官狀態(對手冊/攻略/一代/EXE 驗證,不再等使用者 oracle):**
- [x] **音樂曲目↔場景定案**(task 13):三輪對原版確認到靜態溯源極限——外交樂**反組譯硬證**(Orion2.exe `_diplomacy_bad_music=Get_Random(3)+13` → track 13/14/15);menu/galaxy/combat 因對應 Play 函式在 DOS build 為死碼,維持時長啟發式(誠實標)。
- [x] **地面戰係數**(task 14):RE 定案用一代 1oom `game_ground_kill`(d100+force)+ 二代加成表/hits-to-kill;`ResolveGroundBattle` 實作+確定性測試綠。剩 UI 入侵接線(歸 task 16)。
- [x] **真母星初始狀態**(task 15):Average 忠實開局實作(單一母星、Marine Barracks+Star Base、起始科技對 tech.cpp 驗證、建築數公式、1 Colony+2 Scout)+ 測試綠。
- [~] **核心 gameplay 還原**(task 16):逐塊自驅中(見下)。

**task 16 分塊進度(2026-07-10,使用者授權自主排序):**
- [x] 殖民地建築 5→40 棟入 `gamedata/buildings.go` + 前置科技 gating(`colony-buildings.md`)
- [x] 行星→產出 yield 表(`planet_yield.go`,climate 食物/mineral 工業/gravity,手冊頁碼有據)
- [x] 維護費由建築算(`BuiltMaintenanceBC`,母星 3 BC,取代無據平坦 5)
- [x] 經濟可持續化(玩家+AI 對稱):饑荒復原 + 食物盈餘收入(手冊 p.25)+ 玩家/AI 母星行星驅動 yields;300 回合自我修復、測試更新到忠實基準
- [x] 修 AI 艦隊投資整數捨去 bug(餘數池,FleetStrength 正確成長)+ AI 接忠實 yield
- [x] 地面戰「模型 + 流程」shell 層接線(task 16 續):陸戰隊生成(Marine Barracks 依手冊公式補充,`advanceMarines`)、載運(`LoadMarines`,運力=艦數×手冊每艘 4 的近似,無獨立運輸艦船體類別,標簡化)、入侵解算(`GameSession.InvadeColony`,組 `gamedata.GroundForce` 接 `ResolveGroundBattle`,rng 依回合+星索引種子化可重現)、勝負後續(星 Owner 轉移 + 殖民地過戶/AI 端移除,`internal/shell/ground_invasion.go` + `ground_invasion_test.go`)。剩 UI 繪製/操作介面未做(不碰 interactive.go,歸後續 task)。
- [x] 地面戰補完:裝甲營房戰車 + 軌道轟炸接線(task 16 續,2026-07-11):`gamedata/ground.go` 原本零呼叫端的 `GroundArmorBarracksUnits`/`GroundArmorBarracksCap`/`GroundTankHitsToKill`/`GroundBombHitsFromDamage`/`GroundPlanetTotalHits` 全部接進活對局。**戰車生成**:`advanceArmor`(比照 `advanceMarines`,新增 `PlayerColonyTanks`/`ArmorBarracksAge` 平行陣列,已接進 `EndTurn`)+ `LoadTanks`(與 `LoadMarines` 共用同一個 `MarineTransportCapacity()` 運力池,標簡化)+ `InvadeColony` 攻方 `GroundForce` 混編陸戰隊+戰車(合併順序「陸戰隊在前、戰車在後」,技術原因是靠 `ResolveGroundBattle`「前排先陣亡」規則從單一 `AttackerSurvived` 總數精確拆回兩個兵種各自存活數,不需改動 `gamedata.GroundUnit` 結構);已研究 Battleoids(`TOPIC_ASTRO_CONSTRUCTION`)則戰車固定 3 hits(`GroundBattleoidHitsToKill`)+ 額外 +10 force(`GroundBattleoidCombatBonus`,僅戰車數>0 時套用)。守方戰車 TODO 未接(AI 無 `ColonyBuildings` 追蹤,無從得知是否已建裝甲營房,不臆測)。**軌道轟炸**:新增 `internal/shell/orbital_bombardment.go`,`GameSession.BombardColony(starIdx)` 引擎函式(手冊 p.129,10 輪齊射模擬 `fleetBombardDamage` 重用既有 `ResolveShot`/`ResolveMissileShot` → `GroundBombHitsFromDamage` 換算 hits → 直接扣減殖民地人口,夾在 0)。範圍限制(誠實標註,非杜撰):只扣人口,不扣建築/儲存生產/駐軍(AI 無對應持久資料可扣);轟炸不佔領殖民地(手冊:入侵才佔領);光束/魚雷減半、電腦命中加成、行星護盾在轟炸/戰術戰鬥層本身都還沒有獨立函式,沿用既有 `ResolveShot` 未套用這兩項(TODO);UI 觸發僅引擎層函式,`interactive.go` 未接對應按鈕(誠實延後)。測試:`ground_invasion_test.go`(戰車生成/上限/共用運力/Battleoid/入侵混編拆解/勝率提升對照組)+ `orbital_bombardment_test.go`(前置條件/確定性/人口扣減鏈用保證命中滿傷艦隊手算驗證/不佔領)全綠。詳見 `docs/tech/ground-combat-algorithm.md`「2026-07-11 裝甲營房戰車 + 軌道轟炸接線」節。
- [~] 艦艇設計(空間格):shell/gamedata 層已完成(2026-07-11,`gamedata/shipspace.go` + `session.go` `ShipDesignSpaceUsed`/`ShipDesignFits`,手冊 p.121/124-127 確認值,見 `docs/tech/ship-design-space.md`);仍待武器改裝(mod)佔格接線 + Design Dock UI 繪製。（飛彈/球狀傷害已於下方戰鬥公式分流任務接線,與本項無關)
- [x] 戰鬥公式依武器類型分流(**2026-07-11**):飛彈躲避/AMR 攔截/球狀傷害的公式其實先前就已移植自手冊(`gamedata/missile.go`/`gamedata/damage.go`,有測試),只是戰鬥解算(`cmd/moo2/interactive.go` `fireRound`、`internal/shell/session.go` `battleVolley`)全部武器都走 beam 邏輯(`shell.ResolveShot`),飛彈(核飛彈/麥克萊特飛彈)被當 beam 打。這輪修正:新增 `internal/shell/weapon_kind.go` 依武器名分類 beam/missile/spherical(核對手冊「Notes on Spherical Damage」確認死光不是球狀武器,是一般光束武器且是 `DamageForHit` 手冊 worked example 出處,現行武器表也沒有任何真正的球狀武器);新增 `shell.ResolveMissileShot`(AMR 攔截 + Jam Chance 躲避)、`shell.ResolveSphericalShot`(已測試但暫無武器掛載,備妥待未來新增);`fireRound`/`battleVolley` 依 `CombatShip.Kind`/`combatant.kind` 分流,beam 行為不變(回歸測試)。詳見 `docs/tech/tactical-combat-weapon-kinds.md`。
- [x] AI 財政赤字修正:職務保底(MinWorkersForSolvency/DecideColonyJobsSolvent,只 Scientific 挪 1 人)+ 順修 AI 職務回寫 bug;AI BC 從發散(-217)改收斂有界(48),測試綠(見 ai-fiscal-solvency.md)
- [x] TradeGoodsIncome 接線(2026-07-11):貿易品是建造佇列選項(非第四種職務配置,原判斷是誤判)——建造選單新增「貿易品」、`engine.ColonyState.TradeGoods` + `syncTradeGoodsFlag`、`RunEmpireTurn` 接上 2:1 換算(`EmpireOutput.TradeGoodsRevenue`);Fantastic Trader 仍 TODO。見 `docs/tech/gameplay-systems-status.md` §2
- [x] 原版 672 艦名池翻譯並接入(取代硬編 10 名)(2026-07-11:190 組基底詞意譯+羅馬數字流水號保留,`assets/i18n/shipname.tsv` + `internal/shell/shipnames.go`,見 `docs/tech/proper-noun-strategy.md` 艦名節)
- [x] 原版 829 隨機星名池翻譯並接入(取代二十八宿占位池)(2026-07-11:829 條英文名彼此互不重複——真名/圍棋術語彩蛋/克蘇魯神話等專有名詞優先意譯,虛構短音節規則化音譯,`assets/i18n/starname-random.tsv` + `internal/shell/starnames.go`,`genGalaxy` 改用 `randomStarNamePool`,二十八宿 `starNamePool` 已移除;見 `docs/tech/proper-noun-strategy.md` 隨機星名節)
- [x] **勝利條件(2026-07-11)**:銀河議會選舉(手冊 GAME_MANUAL.pdf p.183,`gamedata/council.go`
  +`shell/council.go`)——議會成立門檻(半數銀河已殖民 + 存續帝國數)、票數=人口(手冊無精確換算
  公式,近似1:1)、2/3超級多數勝出(沿用 `internal/engine/victory.go` 既有但先前從未接線的
  `CheckHighCouncil`)、AI當選時玩家可 accept/reject(手冊:議會無法強迫接受)、玩家達標立即
  勝利。另接殲滅所有對手勝利(沿用同檔 `CheckExtermination`,`InvadeColony` 攻陷AI唯一殖民地後
  立即偵測)。UI 僅議會畫面文字狀態,無獨立結束畫面/accept-reject 互動介面(見 HONEST-STATUS)。
  Antares母星次元傳送門勝利當時仍全無(**已於 2026-07-11 第二輪接線,見下方新任務**)。飛彈躲避/AMR/球狀傷害已接進戰鬥解算(見 task 16 分塊「戰鬥公式依
  武器類型分流」)。**(舊斷言訂正,2026-07-11 見下一項)**:議會成立門檻最初因本 remake 資料模型
  固定只有 1 個 AI 對手,曾用 `councilMinExtantRacesOverride`(=2)覆寫手冊字面值 3——這個覆寫值
  與相關斷言已隨下一項的多 AI 升級移除/訂正,不再成立。
- [x] **多 AI 對手(N=3)+ 真議會(2026-07-11)**:`NewDemoSession` 由建 1 個 AI 對手擴為 3 個
  (`internal/shell/session.go`)——3 個不同母星星(`genGalaxy` 新增 `aiHomes` 參數,均勻攤開
  母星索引,`aiHomes=1` 時與舊版逐位元相同,`RegenGalaxy` 呼叫端行為不變)、3 種不同種族名+
  `ai.Profile` 性格(席隆人/科學、姆瑞森人/好戰、布拉西人/擴張)、`PlayerSpies` 平行陣列同步。
  議會 generalize:移除 `councilMinExtantRacesOverride`,`councilEligible` 直接用手冊字面值
  `gamedata.CouncilMinExtantRaces`(=3,玩家+3AI=4個帝國,門檻真的可達);`advanceCouncil` 由
  「玩家 vs 單一 AI 二元計票」改為逐帝國(玩家+每個AI各自獨立)算票、2/3門檻用全體總票數判定,
  `PendingCouncilElection.EnemyName` 正確指向實際當選的 AI(非寫死 `AIPlayers[0]`)。~40 回合
  regression 探針驗證:3 個 AI 各自獨立成長(殖民地/軍力隨性格分化)、玩家開局經濟不 regression、
  議會用真門檻正常召開、全程無 panic、spy 對每個 AI 都結算。仍未做:AI 選星策略(索引順序非
  距離導向)、AI 對 AI 互動(彼此不打仗不外交)、「候選人限定票數最高兩位+第三方外交搖擺票」
  (需要 AI 對 AI 關係模型)。詳見 `docs/tech/victory-conditions.md`、`internal/shell/multi_ai_test.go`。
- [x] **安塔蘭勝利路徑(第三條,2026-07-11 第二輪)**:次元傳送門(手冊 p.106,`gamedata.Buildings`
  早已存在,`BUILDING_DIMENSIONAL_PORTAL`,前置 `TOPIC_MULTIDIMENSIONAL_PHYSICS`,先前建成後無
  任何後續流程)建成後解鎖 `internal/shell/antaran_victory.go` 的 `GameSession.AssaultAntares()`——
  沿用 `ResolveBattle` 同款 `battleVolley` 解算(比一般戰鬥更嚴格:要求防禦方全滅才算勝,呼應手冊
  「defeat the awe-inspiring Antarans」語意)。**母星防禦艦隊戰力手冊/openorion2 均無精確數字**
  (手冊只用「awe-inspiring」定性描述),保守預設 6 艘末日之星等級戰力(合計戰力384),已誠實標注
  待考證。戰勝設 `AntaranHomeworldConquered=true`,`advanceAntaranVictory`(`EndTurn` 呼叫,順序排
  在殲滅之後、議會之前,對齊 `engine.CheckVictory` 文件記載的優先序)偵測並結束遊戲
  (`Reason=engine.VictoryAntaran`)。`CanAssaultAntares()` 前置:遊戲未結束+`!DisableEvents`
  (手冊:關閉安塔蘭攻擊則本路徑不可用)+已建傳送門+艦隊非空。最小 UI:艦隊列表畫面(`fleet()`)
  加一個文字提示熱區,只在前置滿足時顯示,點擊後導向既有戰鬥結果畫面(複用 `LastBattle`)。
  單測:`internal/shell/antaran_victory_test.go`(前置條件各分支擋下、弱艦隊戰敗不誤判、強艦隊
  戰勝後正確偵測勝利、殲滅與安塔蘭同時成立時優先序不亂)。詳見 `docs/tech/victory-conditions.md`
  §4.4。**手冊三條勝利路徑至此全數接線可達成。**
- [x] **間諜最小可玩迴圈(2026-07-11)**:`gamedata/spy.go`(手冊 `Notes on Spying` 8 個機率
  函式,先前零呼叫端死碼)接上 `internal/shell/spy.go`——訓練間諜(`TrainSpy`,花 30 BC
  remake 拍板值)→ 每回合結算(`advanceEspionage`,由 `EndTurn` 呼叫)偷科技(STEAL,偷一項
  「對方已知、我方未知」的科技,依 GAME_MANUAL.pdf p.174-175「tries to steal technologies
  you have yet to gain」推出)→ SpyVsSpy 判定(±80 淨值門檻)。玩家 ↔ 每個 AI 對手雙向生效
  (`PlayerSpies`/`AIOpponent.Spies` 皆為平行陣列/逐一結算,`NewDemoSession` 現有 3 個 AI 對手
  時同樣各自獨立算,見上一項多 AI 升級),維護費 opt-in(0 間諜時零影響)。**只做 STEAL**:破壞
  (SABOTAGE)手冊無數值規則,標 TODO 不做;逐對手分配/任務選單(Espionage/Sabotage/Hide)延後;
  防禦方 Agent 不獨立追蹤(DB 固定 0,對應手冊「零 Agent 防禦仍生效」);種族/科技/政府對間諜的
  加成現行無資料可推導,一律 0(TODO)。詳見 `docs/tech/spy-system.md`。

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
- [x] 其餘字串源逐一 dump + 翻(2026-07-11 盤點:多數已完成,見 assets/i18n/):科技描述 techdesc.tsv(83)、種族 races/raceinfo.tsv、事件 event.tsv(98)、外交 diplo.tsv(780)、help.tsv(704)、母星名 starname.tsv、技能 skilldesc.tsv、estrings(585)/rstring(178)/antaran、艦名 shipname.tsv(535,同日稍後完成,見下方獨立項)、隨機星名 starname-random.tsv(829,同日稍後完成,見下方獨立項)
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
- [x] galaxy 工具列 GAME 標題已翻(→遊戲)+ ZOOM 已翻(→縮放)(2026-07-10);行星/艦隊個別按鈕邊緣極微殘(紋理按鈕固有)為長尾
- [ ] 各子畫面 RETURN 按鈕精確熱區(目前暫用全螢幕返回)
- [x] 科技總覽「科技總覽」列可點進研究選擇畫面(其餘選單項待接)
- [x] 殖民地總覽畫面(COLSUM.LBX 0)接入 COLONIES 按鈕 + 完整中文化
- [x] 種族關係畫面(RACES.LBX 0)接入 RACES 按鈕 + 中文化(種族關係/會晤/報告/宣戰/忽略/加成/返回)
- [x] **★ 真新遊戲流程**:主選單→新遊戲→原版 NEW GAME 設定畫面(NEWGAME.LBX 28,調色盤鏈 RACEOPT#4→NEWGAME#1)→ACCEPT→星系主畫面;中文化(難度/星系大小/星系年齡/玩家數/科技等級/戰術戰鬥/隨機事件/安塔蘭攻擊/取消/接受)
- [x] **★ 獨立種族選擇畫面(2026-07-10,對原版流程還原)**:依 GAME_MANUAL 流程,設定畫面 Accept 改導向獨立種族選擇畫面(`cmd/moo2/raceselect.go`,RACEOPT#0 螢幕框 + 14 族中文名 + 真肖像 RACESEL 15–28 字母序 + 描述 + 取消/接受)。取代原「設定畫面擠一格循環種族」。研究見 `docs/tech/newgame-flow.md`。
  - [~] 版面像素對齊原版 + 用 RACESEL 名稱按鈕圖/描述板;Custom 點數畫面;命名+旗色;依 Starting Civilization 真實母星初始(WORKLIST 續,task 8)
- [x] 回合摘要畫面(TURNSUM.LBX#0)接入 TURN 流程(結束回合→摘要顯示本回合結算:星曆/淨工業/研究/食物/稅收/國庫變化/研究完成)→關閉回星系。中文化(回合摘要/關閉)
- [x] 艦艇設計畫面(DESIGN.LBX#0)接入(艦隊→點艦艇格→艦艇設計)+ 中文化(艦艇設計/巡防艦…末日之星/清除/取消/建造);艦隊 RETURN 改精確熱區
- [x] 議會畫面(COUNCIL.LBX#1)接入 + 投票系統(2026-07-11 大改,見下方「勝利條件」任務):舊版
  `CouncilVote`(無成立門檻、無2/3多數、票數=人口較高者當選)已移除,畫面改讀
  `GameSession.CouncilStatus()` 誠實呈現議會是否已成立/目前票數/是否已分出勝負或待玩家回應
- [x] 已探測定位背景(remain-scan,待接入):讀取存檔 LOADSAVE.LBX#11(空存檔格)、外交 DIPLOMAT.LBX#29(有雜訊待查)
- [x] **存檔/讀檔(remake 自身格式)**:GameSession JSON 序列化(shell/persist.go),AI Decider 以性格重建、含未匯出遊戲狀態;每回合自動存檔(UserConfigDir),主選單「載入遊戲/繼續」讀回續玩。測試 TestSaveLoadRoundTrip(Turn/BC/種族/星系/艦隊/建造/AI 一致且可續跑)
- [ ] 細修:NEW GAME 開關列/標題微殘、種族關係 ESPIONAGE/SABOTAGE/HIDE(24 標籤)未翻、各畫面按鈕精確熱區
- [x] **★ 核心遊戲迴圈第一步**:GameSession 接進 -game;TURN 按鈕呼叫 session.EndTurn()
  (結算帝國經濟 + AI 對手決策),星系畫面即時顯示星曆(3500 起,每回合+1年)+ 國庫 BC
  (overlayScreen.extras 動態文字機制)。驗證:TURN×2 → 星曆 3500→3502、國庫 100→106
- [x] 待接入畫面:議會/艦艇設計/回合摘要 已接入並中文化(見上)+ 單一殖民地管理已自建(見下);讀存檔背景已備、入口待接
- [x] 殖民地總覽填即時資料:玩家各殖民地列出「殖民地 N / 農夫 / 工人 / 科學家」(來自 GameSession,對齊原版欄位,extras 動態文字)
- [x] 行星列表填即時資料(每星生成行星:名/氣候/重力/礦產/大小,PIL 量測列對齊)
- [x] 軍官列表填即時資料(領袖名單:名/專長/等級,4 槽位對齊)**(2026-07-11 追加:技能字串已從純裝飾接上真加成,見 Phase 5「領袖/軍官技能」條目)**
- [x] 艦隊畫面填即時資料(艦隊名冊:艦名/艦體等級)
- [x] 造船系統:艦艇設計點艦體等級→BuildShip 加艦到 session→艦隊顯示新艦(第二個互動系統)
- [x] 單一殖民地管理:殖民地總覽點職務欄→ShiftColonyJob 重分配人口(影響下回合經濟);造船改扣國庫 BC(戰力×20)
- [x] 建造佇列:殖民地總覽建造欄可點選建築(住宅/工廠/研究實驗室/星港),結束回合以淨工業累積建造,完成回合摘要通知
- [ ] colony 名稱改用真星名;4 列表畫面接真存檔/生成資料
- [x] 星圖互動:點星→黃色高亮環+左下角行星資訊面板(名/氣候/大小/重力/礦產)+ 派遣艦隊鈕(星間航行)
- [x] 程序化星系生成:genGalaxy(種子亂數,抖動網格佈 24 星,隨機光譜/大小/洗牌星名,玩家/AI 母星)取代固定佈局
- [x] 星系大小接 NEW GAME 設定:GALAXY SIZE 框可點選(小型12/中型24/大型36/巨型48星),ACCEPT 依選定大小 RegenGalaxy
- [x] 難度設定生效:DIFFICULTY 框可選(簡單/普通/困難/不可能),敵方戰力倍率套用到戰鬥
- [x] **★ DIPLOMAT 解碼修復**:對照 openorion2 發現多幀動畫需「累積各幀(delta)+ 未寫入填 palette[0]」(先前當透明→白噪)。lbx.Image.AccumulatedRGBA;外交議事廳(DIPLOMAT#29,38幀)以真原版圖 + diplomat#0 調色盤渲染,疊外交對談(和平/貿易/威脅)→ 16/16 原版畫面皆真圖
- [x] 種子隨新遊戲變化(newGameSeed 遞增);戰術戰鬥艦艇移動+射程限制開火(格位/選取/移動)
- [x] 數值對齊 MOO2 規格:艦體成本(空殼生產18/60/180/540/1620/4860,設計畫面顯示)、建築成本(自動工廠/海軍營/研究實驗室60、太空港100、星基300)、研究成本(gamedata 權威 cResearchCosts 表)
- [x] 艦艇設計武器元件:選主武器(無武裝/雷射/質量投射器/核飛彈/離子砲,各成本+攻擊加成),建造成本=艦體+武器、戰鬥攻擊=艦體+武器
- [x] 完整艦艇元件系統:武器/裝甲/護盾/特殊四類元件(各含成本+效果係數),設計畫面循環選擇+顯示總價;建造套用(裝甲/護盾→HP、武器/戰鬥電腦→攻擊)
- [x] 元件解鎖綁研究科技:各進階元件標記所需 gamedata 研究主題,未完成研究則鎖住(循環跳過),設計畫面顯示已解鎖數;研究→解鎖元件→造艦系統打通
- [x] 元件品項擴充:29 個 MOO2 真實元件(武器11:雷射→死光/裝甲7:鈦→氙素/護盾6:第一~第十級/特殊5),真譯名(tech.tsv)+ 遞增係數 + 各綁研究科技門檻
- [~] 元件係數對齊:武器 Value 改真「最大傷害」,錨定 patch 1.5 官方確認值(中子爆破槍12/高斯砲18/電漿砲20);其餘標注單調估計。provenance + 阻塞點(完整表僅存於掃描版手冊,需 OCR;且係數版本相依)記於 docs/tech/component-values.md
- [ ] 精確全表:OCR 掃描版手冊附錄 或 逆向私有 gamedata 武器表;建版本專屬 profile(1.3/1.5 數值分版)
- [x] **研究自動推進 → 動態解鎖迴圈**:目前主題完成後自動推進到下一個未完成元件主題(researchQueue 依成本遞增),玩數回合便逐步解鎖進階元件。測試 TestResearchUnlockLoopOverTurns 驗證 40 回合解鎖 7→15、完成 6 主題
- [x] 新遊戲種族選擇:NEW GAME 設定畫面加種族選擇框(13 經典種族循環選,顯示名+特性),ACCEPT 套 ApplyRace 起始加成(工業/研究/食物/成長/國庫/戰鬥百分點,對齊各族招牌特性)。測試 TestApplyRaceBonuses/SakkraGrowthFaster/MrrshanCombatBonus
- [ ] hover highlight 與原版一致(目前為細框提示)
- [ ] 淘汰自製簡約殼(`-play`):方向不符「與原版一模一樣」,改以原版 overlay 畫面 + 既有回合引擎(internal/engine)重建可玩迴圈
- [ ] 補齊需全域調色盤鏈的畫面(COLONY/DESIGN/COUNCIL/DIPLOMAT…)到對照組
- [ ] **[HARD] 開工先做:窮舉所有文字源(LBX 各類 + Go hardcode),各寫 dumper,用引擎自己 reader dump 精確 key**
- [ ] 逐畫面重建:主選單/載存檔/星系圖/行星清單/殖民地/科技研究/艦隊/軍官/種族資訊/對話框
- [ ] IMGLOG 探查模式:記錄 `(lbx,index)` 對照畫面 UI(盤點烘字按鈕/標籤用)
- [ ] 烘進 gfx 的英文:擦底疊字(cht_label 模式)or 整圖替換(image_override 模式)
- [x] LBX 字串譯文表:科技名/描述、種族、事件、外交、星名、help、技能、殖民地、議會、選單等 22 個逐源分檔 TSV 已完成(assets/i18n/*.tsv);艦名池(2026-07-11 補完,shipname.tsv)、隨機星名池(2026-07-11 補完,starname-random.tsv)均已落地,四個專有名詞池全數定案
- [ ] 組合字串走 `TranslateFormat` 翻模板字面(佔位符數/序中英一致)
- [ ] 專有名詞術語表 + 「中文(英文)」小字控制碼(統一譯名,對齊 moo1/mom 經驗)
- [ ] 每畫面 xvfb + xdotool 導航 + import 截圖校對(破版/溢出/缺字/置中)

## Phase 5 — Gameplay 引擎重建
- [x] 回合結算主迴圈(engine.RunEmpireTurn:殖民地經濟聚合+稅收+國庫+研究推進)
- [x] 殖民地經濟:食物/工業/研究/稅收/國庫已實作(engine);人口成長回寫 Population 已補(shell.advancePopulation 累加 PopGrowth 達門檻 +1 人口、新單位為工人、受 PopMax 上限;門檻為 remake 調校值,provenance 見 session.go 註記)。測試 TestPopulationGrowthWriteback/CappedAtMax
- [x] 建造佇列 + 建築長期效果:advanceBuilds 完工後套用永久產出加成,每殖民地每種只套一次(ColonyBuildings 去重);殖民地總覽顯示已建建築。**(2026-07-11 忠實化訂正)**先前把手冊「殖民地整體固定加成」揉進 per-worker 欄位湊數(自動工廠工業/工人+2、研究實驗室研究/科學家+5 等,小殖民地過度受益、大殖民地不足),現分開建模:per-worker 訂正回手冊值 + 新增 `FlatFood`/`FlatIndustry`/`FlatResearch`(固定加成)、`IncomeBonusPercent`(太空港+50%/證券交易所+100%,逐殖民地精確套用於 `RunEmpireTurn`)、`PopMax` 直接加成(生態圈+2)、`FlatGrowth`(複製中心)。機器人工廠(2026-07-11 已接線,見下)。共 18 棟已忠實建模數值,詳見 `docs/tech/colony-buildings.md` §6。測試 TestBuildingLongTermEffect/TestResearchLabEffect/TestSpaceportIncomeBonusPercent/TestBiospheresRaisesPopMax 等(engine+shell)
- [x] 機器人工廠礦產豐度分級接線(p.82)**(2026-07-11)**:比照重力懲罰的接線手法(`4c2a26a`),`engine.ColonyState` 新增 `MineralRichness gamedata.PlanetMinerals` 欄位,獨立保留建立殖民地當下的原始礦產豐度分類(先前只烘進 `IndustryPerWorker` 靜態費率,事後拿不回原始分類)。零值陷阱處理:`gamedata.ULTRA_POOR` ordinal=0,故全部既有 `ColonyState{...}` 建構點(engine/shell 測試、`cmd/moo2sim`)皆已明確補上本欄位。`applyBuildingEffect` 的機器人工廠 case 依 `gamedata.ProdRoboticFactoryBonus(int(cs.MineralRichness))`(`internal/gamedata/production.go` 既有查表函式,索引與 `mineralProductionTable` 一致)查出手冊固定值(Ultra Poor+5/Poor+8/Abundant+10/Rich+15/Ultra Rich+20)加進 `FlatIndustry`,不動 `IndustryPerWorker`。存檔行星由 `ColonyStateFromSave` 讀 `save.Planet.Minerals`(與 `gamedata.PlanetMinerals` 同源 openorion2 enum ordinal,可直接轉型,同重力)。母星固定 Abundant。測試:`TestRoboticFactoryEffect`(母星 Abundant+10)、`TestRoboticFactoryEffectByMineralRichness`(五級分級逐一驗證,含 UltraPoor+5/Rich+15)(shell)。
- [x] 重力懲罰接進生產管線(**2026-07-11**):`ColonyState` 新增 `PlanetGravity` 欄位,`colonyFood`/`RunColonyTurn` 對食物/工業/研究三種 per-worker 產出套用 `gamedata.GravityPenaltyPercent`(Low-G -25%、Heavy-G -50%;士氣+重力先加總成單一百分點再套一次 `GravityAdjustedProduction`,避免兩次連續整數除法的複合誤差,理由見 `internal/engine/colony.go` 註解)。行星重力產生器 `NormalizeGravity` 旗標由 no-op 變成真的會歸零懲罰。`ColonyStateFromSave`(存檔↔engine 橋接)同步接上 `save.Planet.Gravity`(與 `gamedata.PlanetGravity` 同源 openorion2 enum ordinal,直接轉型)。種族 Low-G/High-G 重力天賦未建模,固定以一般種族為基準;固定加成(Flat*)不吃重力。**已知現實限制**:本專案唯一的殖民地建構點(`NewDemoSession`/`playerHomeworldColony`)固定 Normal-G,尚無「開拓新殖民地」流程會產生 Low-G/Heavy-G 殖民地,故此接線在 demo session 暫不可見,主要對存檔載入模式(`RunGameTurn`)生效。測試 TestRunColonyTurnGravityHeavyPenalty/TestRunColonyTurnGravityNormalizeGravityCancelsPenalty/TestRunColonyTurnGravityNormalGNoPenalty/TestRunColonyTurnGravityAndMoraleCombinedPercent/TestColonyStateFromSaveGravityMapping(engine)
- [x] 士氣(Morale)接進 MoralePercent(**2026-07-11**):`GameSession` 新增 `Government`(`gamedata.MoraleGovernmentType`)欄位,`ApplyGovernment` 記錄政府型態(`Governments` 索引→`moraleGovByIndex`,四選一映射到對應基礎政府,進階政府 Imperium/Confederation/Federation/Galactic Unification 不區分)。新函式 `colonyMoralePercent`(`internal/shell/session.go`)= `gamedata.MoraleGovernmentBase(gov, hasBarracks)`(手冊 -20%/無 Barracks)+ 全息模擬艙(`MoraleHoloSimulatorBonus`+20%)+ 歡樂穹頂(`MoralePleasureDomeBonus`+30%),依 `ColonyBuildings` 讀取已建建築;政府變更(`ApplyGovernment`)與建築完工(`advanceBuilds`→`recalcColonyMorale`)皆會重算。**母星起始 `MoralePercent` 從無據硬編 +10 訂正為忠實值 0**(獨裁 + 已建 Marine Barracks 抵消 -20% 懲罰,無士氣建築加成;見 `playerHomeworldColony` 註解,`TestGameSessionEndTurn` 已同步訂正預期值 33→30)。誠實未套用(手冊有但不假裝精確):多種族懲罰(`MoraleMultiRacialPenalty`,remake 不追蹤殖民地是否多種族,異族管理中心暫無可見效果)、首都淪陷懲罰(remake 無首都被攻陷狀態)、Virtual Reality Network(手冊定性為「成就」非建築,不在 `gamedata.Buildings`,remake 無成就系統)。測試 TestColonyMoralePercentDictatorshipNoBarracks/TestColonyMoralePercentBarracksCancelsPenalty/TestColonyMoralePercentHoloSimulatorAndPleasureDomeStack/TestColonyMoralePercentGovernmentDiffers/TestApplyGovernmentRecalculatesMorale/TestMoralePercentAffectsColonyProduction(shell)。詳見 `docs/tech/colony-buildings.md` §6.1 士氣列、`docs/HONEST-STATUS.md` 2026-07-11 追加段。
- [x] 指揮評等(Command Rating)供需接線(**2026-07-11**):手冊 p.169「size class」公式(Frigate=1..Doom Star=6,`gamedata.ShipCommandCost`,以 Titan=5/Doom Star=6 兩處具體數字交叉驗證)+「每未覆蓋點 -10 BC」超支懲罰,先前 `gamedata.IncomeCommandOverflowCost` 是零呼叫端死碼。供給端:星基+1/戰鬥站+2/星辰要塞+3(三者取代不疊加,`gamedata.CommandPointsFromBuildings`)。`engine.PlayerState` 新增 `CommandPointsSupply`/`UsedCommandPoints` 欄位,`shell.GameSession.EndTurn` 每回合依實際已建成軌道衛星(`totalCommandPointsSupply`)與艦隊(`usedCommandPoints`)重算,`engine.RunEmpireTurn` 算超支併入 `NetBC`(新增 `EmpireOutput.CommandOverflowCost` 曝露懲罰金額)。當時誤判「開局母星 1 座星基(+1)vs 3 艘開局艦艇(需求3),缺口2點恆定-20BC/回合」為手冊忠實結果,實為**regression**(見下方同日修復項)。誠實未做(手冊有數字但架構未跟上,詳見 `docs/tech/moo2-formulas-reference.md`「指揮評等供需」節):通訊科技(Tachyon+1/Hyperspace+3,每軌道衛星)、Imperium 政府 +50%(本專案政府型態全域固定 Dictatorship,無 Imperium 狀態)、Operations 軍官技能(手冊無精確數字)、AI 對手(抽象 FleetStrength 無逐艦清單,供需維持零值無懲罰)。測試 TestShipCommandCost/TestShipCommandCostOutOfRange/TestCommandPointsFromBuildings(gamedata)、TestRunEmpireTurnCommandOverflow/TestRunEmpireTurnCommandSupplyCoversDemand(engine)、TestTotalCommandPointsSupply/TestUsedCommandPoints/TestUsedCommandPointsEmptyFleet/TestEndTurnCommandOverflowPenalty/TestUsedCommandPointsUsesGamedataTable(shell)。
- [x] 指揮評等開局死亡螺旋 regression 修復(**2026-07-11**,同日接線後發現):上一項漏算了帝國基礎指揮評等供給,誤判「開局-20BC/回合」為忠實機制。用真實存檔 `SAVE10.GAM`(`/home/anr2/moo2-private-build/gamedata/mastori2/SAVE10.GAM`)oracle 反推(rulebook 62/64):5 個活躍玩家(不同種族)各持 1 殖民地,`CommandPoints` 讀到 6(其中 1 名玩家=8);比對已建成軌道衛星,讀到 6 的玩家只建星基(6-1=5),讀到 8 的玩家建星辰要塞(8-3=5)——5 個不同種族玩家一致反推基礎值 5,與種族/政府無關。新增 `gamedata.CommandPointsBase=5`(`income.go`,含完整 oracle 推導註解),`shell.GameSession.totalCommandPointsSupply()` 在逐殖民地建築供給之外每帝國加這一次(非逐殖民地)。修復後開局供給=5+1(星基)=6≥3(需求),不再超支;20 回合探針軌跡:BC 從第 2 回合 101 穩定爬升至第 21 回合 136,人口穩定在 8→9,無死亡螺旋(修復前:BC 第 7 回合轉負、第 21 回合 -255,人口第 20 回合起餓死)。300 回合被動不建造測試(`events_test.go` `bcCrashFloor300Turns`)實測最低點從 -3710(第 273 回合)改善到約 -51(第 133 回合),門檻由 -4000 收回 -400。**已知限制(TODO)**:單一存檔皆 1 殖民地,無法分辨此 5 點是 per-empire flat 還是 per-colony,暫採 per-empire flat,待多殖民地存檔驗證。測試更新:`TestTotalCommandPointsSupply`(6→11)、`TestEndTurnCommandOverflowPenalty`(改用外加艦隊建構真實超支情境,原始 20/10/0 三情境已不成立)。詳見 `docs/HONEST-STATUS.md`/`docs/tech/moo2-formulas-reference.md`「指揮評等供需」節/`docs/tech/remaining-work-roadmap.md` A項。
- [x] 科技研究樹推進(engine.RunResearchPhase 累積+完成判定+溢出保留;session.advanceResearch 自動推進主題)
- [x] 艦隊移動 + 星圖導航:SendFleet 依星距換算 ETA,EndTurn 跨回合推進,抵達標記探索;星圖點星→面板「派遣艦隊至此星」鈕 + 青色艦隊標記 + 航行連線 + ETA 顯示。測試 TestFleetInterstellarMovement
- [ ] 艦艇設計
- [x] 戰鬥:格子戰術戰鬥(2026-07-10 換原版美術:STARBG 星空+COMBAT 控制列+可見 CMBTSHP 艦艇+控制列 7 按鈕中文化;逐發用真 ResolveShot 命中/傷害/過盾/過甲);宣戰→戰術戰鬥→戰鬥結果。**(2026-07-11 更新:武器依 beam/missile/spherical 分流,飛彈躲避/AMR/球狀傷害公式接進解算,見 `tactical-combat-weapon-kinds.md`)**。艦型 sprite 完整對照(task 12)仍待
- [x] 外交對談(2026-07-10 破解 DIPLOMAT.LBX 換原版美術:逐族使節房+使節疊合,13 族對應對 RACESEL 核實);銀河議會選舉勝利條件(2026-07-11,見下方勝利條件任務,取代原本無門檻/無2/3多數的簡化投票)
- [x] 隨機事件系統:每回合 30% 觸發 6 種 MOO2 風格事件(經濟繁榮/太空海盜/富礦脈/瘟疫/科學突破/隕石),效果有界(BC 不為負、人口不低於1)、種子化可重現,顯示於回合摘要。測試 TestRandomEventsFireAndBounded/Reproducible
- [x] 安塔蘭人入侵:週期性終局威脅(前20回合寬限,之後每15回合一次),強度隨次數升級,攻母星(人口+BC損失,有界),母星艦隊可部分防禦減損;顯示於回合摘要(紅色警報)。測試 TestAntaresRaidsScheduleAndEscalate/DefenseReducesDamage
- [~] AI 對手主動行為:造艦(淨工業投資軍力,好戰性格更多)/ 擴張(每5回合佔無主星,**2026-07-11 更新:改用共用函式 `newColonyFromStar` 建真 `engine.ColonyState`,不再只標旗標——見下方「AI 拓殖建真殖民地」條**)/ 外交態勢(依 AI-玩家軍力差+難度漂移關係→ai.DecideStance 宣戰/敵視/中立/提議貿易/結盟);種族關係畫面顯示各 AI 名/態勢/軍力/佔星。測試 TestAIBuildsAndExpands/StanceHostileWhenStrong/AIExpand_CreatesRealColony/AIExpand_EconomyGrowsWithColonyCount/AIExpand_NoOpWhenNoUnownedStars。深層策略見 `docs/kickoff/07-ai-strategy.md`:先參考 1oom `game_ai_classic.c` + GameFAQs 文獻,有必要才逆向)
  - [ ] 精讀 1oom `game_ai_classic.c`,抽「AI 決策流程」語言無關筆記
  - [ ] 精讀 GameFAQs MOO2 AI FAQ + 策略指南,補 MOO2 特有行為
  - [x] 設計可插拔 AI 介面(ai.Decider)+ 難度加成係數(已用於經濟+態勢)
  - [ ] 標示「必須逆向才能確定」的項目(若有)
- [x] 開新遊戲流程:種族選擇 + 星系大小/難度 → ApplyRace/RegenGalaxy(見 Phase 4b)
- [x] 地形改造(Terraforming)/蓋亞轉化(Gaia Transformation)/土壤改良(Soil Enrichment)接線(**2026-07-11**):`internal/gamedata/terraform.go` 移植好的氣候階梯/人口係數公式先前零呼叫端(死碼),現接進殖民地建造佇列。新增 `engine.ColonyState.Climate` 欄位(比照 `PlanetGravity`/`MineralRichness` 的零值陷阱處理:`gamedata.TOXIC` ordinal=0,`playerHomeworldColony`/`ColonyStateFromSave` 皆已明確補上;此欄位不像 Gravity/MineralRichness 被每回合核心公式讀取,只在地形改造/蓋亞轉化套用瞬間讀寫,故其餘既有測試字面值不受影響、無需逐一補值)。新增 `internal/gamedata/special_actions.go`:`SpecialAction`/`SpecialActions`/`SpecialActionByNameZH`/`AvailableSpecialActions`,把這三項「Special」型別一次性行動(區別於常駐 Building,不計入 `colony-buildings.md` 40 項建築表)排進 `availableBuildOptions`/`allBuildOptions`。前置科技(地形改造 `TOPIC_GENETIC_MUTATIONS`、蓋亞轉化 `TOPIC_TRANS_GENETICS`、土壤改良 `TOPIC_ADVANCED_BIOLOGY`)取自 `openorion2/src/tech.cpp` 的 `research_choices[]`(陣列索引=`ResearchTopic` 列舉值,已與既有 34 項建築前置科技逐一交叉核對 100% 相符,地形改造的 `TOPIC_GENETIC_MUTATIONS` 亦與 `terraform.go` 檔頭「移植自...『Genetic Mutations』章節」的手冊出處吻合)。`shell.advanceBuilds` 新增分流:這三項完工時呼叫 `applySpecialAction`(不記入 `ColonyBuildings` dedup map,因手冊明講地形改造可重複套用,若記入 dedup 會被既有「已建過不再套用」邏輯擋下第二次),推進氣候(`TerraformNextClimateOptions`/`GaiaTransformationCanApply`)並用新增的 `gamedata.TerraformPopMaxAfterClimateChange` 等比例縮放 PopMax、`ClimateFoodPerFarmer` 差值疊加 FoodPerFarmer(保留既有建築加成)。**誠實近似/TODO**:PopMax 縮放非精確重算(remake 無「行星尺寸→基礎人口容量」對映表,詳見該函式註解);建造成本(PP)手冊無數據,比照其餘估計建築的 RP 量級外推(260/900/150),手冊「地形改造每次套用成本遞增」未模擬(固定成本);Barren 地形改造下一級的兩個候選(Desert/Tundra)手冊未給選擇條件,固定選第一個。測試:`TestTerraformPopMaxAfterClimateChange`/`TestSpecialActionByNameZH`/`TestAvailableSpecialActions`(gamedata)、`TestTerraformAdvancesClimateFoodAndPopMax`/`TestTerraformNoOpWhenNoNextClimate`/`TestGaiaTransformationRequiresTerran`/`TestSoilEnrichmentBlockedOnHostileClimate`/`TestSoilEnrichmentWorksOnHospitableClimate`(shell)。詳見 `docs/tech/colony-buildings.md` §6.1 地形改造列、`docs/HONEST-STATUS.md` 2026-07-11 追加段。
- [x] income.go 三個零呼叫端死碼接線(**2026-07-11**,解鎖自本輪稍早的開局經濟平衡修復):
  ①**政府 money 加成**(MANUAL_150.html govt_bonus democracy_money=10→50%/federation_money=15→75%,
  `gamedata.IncomeApplyGovernmentMoneyBonus`)。新增 `gamedata.IncomeGovtMoneyBonusPercent(gov)` 查表
  (Democracy→50、Federation→75、其餘→0)+ `engine.PlayerState.GovtBonusMoneyPercent` 欄位(呼叫端
  算好傳入,同 `Maintenance`/`CommandPointsSupply` 輸入模式)。`shell.GameSession.EndTurn` 依
  `s.Government` 算好傳入,`RunEmpireTurn` 在逐殖民地迴圈**結束後**(帝國層級,非逐殖民地——政府
  是帝國屬性不是殖民地建築)對 `TaxRevenue+FoodSurplusRevenue+TradeGoodsRevenue` 套一次,差額併入
  `TaxRevenue`。demo 預設 Dictatorship→0,no-op;AI 對手無 `Government` 欄位建模,不受影響。
  ②**運輸艦(Freighter)維護費**(每艘使用中 -0.5 BC,`gamedata.IncomeFreighterMaintenanceCost`)。
  新增 `engine.PlayerState.ActiveFreighters` 欄位,`RunEmpireTurn` 算出 `EmpireOutput.FreighterMaintenanceCost`
  併入 `NetBC`。本專案艦種塑模(`gamedata.ShipType`:`COMBAT_SHIP`/`COLONY_SHIP`/`TRANSPORT_SHIP`/
  `OUTPOST_SHIP`)沒有獨立的「Freighter」艦種(`TRANSPORT_SHIP` 是地面入侵運兵船,非手冊講的貨運
  艦隊),呼叫端恆傳 0,目前 no-op,接線先備妥。③**士氣對收入的調整**
  (`gamedata.IncomeMoraleAdjustedProduction`,手冊 p.170)**判定為刻意不接**:查證
  `internal/engine/colony.go` `RunColonyTurn` 發現士氣(`MoralePercent`)早就套進食物/工業/研究的
  per-worker 產出(`pct := cs.MoralePercent + colonyGravityPenaltyPercent(cs)` 套 `GravityAdjustedProduction`),
  `RunEmpireTurn` 的 `TaxRevenue`(讀 `co.NetIndustry`)/`FoodSurplusRevenue`(讀 `co.FoodSurplus`)/
  `TradeGoodsRevenue`(讀 `co.NetIndustry`)全部是從這個已調整過的產出直接換算,若再套一次士氣就是
  雙重計算(同一筆錢士氣生效兩次)。故不呼叫該函式,判定依據完整記錄在 `engine/empire.go` 註解與
  `docs/tech/moo2-formulas-reference.md`「士氣對收入的影響」節;函式本身與其單元測試保留(驗證公式
  正確,非死碼)。三項在 demo 對局皆 no-op(政府=Dictatorship、無貨運艦種、母星 morale=0),20 回合
  BC 軌跡探針確認接線前後一致(101→130 健康爬升,無 regression)。測試:
  `TestIncomeGovtBonusFormula`/`TestIncomeFreighterMaintenanceCost`/`TestIncomeMoraleAdjustedProduction`/
  `TestIncomeApplyGovernmentMoneyBonus`(gamedata,原有公式測試)、
  `TestRunEmpireTurnGovtBonusMoneyPercent`/`TestRunEmpireTurnGovtBonusMoneyPercentZeroNoOp`/
  `TestRunEmpireTurnFreighterMaintenance`/`TestRunEmpireTurnFreighterMaintenanceZeroNoOp`(engine,新增)、
  `TestEndTurnGovtBonusMoneyWiring`(shell,新增)。詳見 `docs/HONEST-STATUS.md` 2026-07-11 收入死碼段落、
  `docs/tech/moo2-formulas-reference.md`「政府對 BC 收入的加成」/「士氣對收入的影響」節。
- [~] 以手冊逐系統對照驗證規則正確性(task 16 進行中:地面戰解算/真母星/建築全表/行星 yield/建築維護費 已逐項對手冊或一代驗證並實作;經濟可持續化+yield 接線進行中)
- [x] 最小拓殖(Colonization)接線(**2026-07-11**):先前玩家只有母星、完全無法擴張——「能玩完整一局」的最大缺口(見 `remaining-work-roadmap.md` B 項)。硬門檻查證(`GAME_MANUAL.pdf` 直接引文):適居性(p.55/p.61,一般行星=habitable worlds 可由殖民船直接殖民,不需額外科技;氣態巨星/小行星帶才需另建軍事前哨+科技,本 remake 星系生成從未產生這兩類行星,gate 現階段恆真、留掛勾點)、起始人口=1(p.61-62,Colony Base/Colony Ship 手冊原文一致)、起始無建築(對照母星起始建築是手冊明講的特例)。PopMax 公式移植自 `openorion2/src/gamestate.cpp:2288` `GameState::planetMaxPop`,已與手冊 p.55-56 各尺寸人口容量範圍逐項交叉驗證(新增 `gamedata.PlanetBasePopMax` + `TestPlanetBasePopMaxManualRanges`)。新增 `internal/shell/colonization.go`:`GameSession.ColonizeStar(starIdx)` 引擎函式(前置條件:艦隊已抵達無主星+載有殖民船;成功則建新 `engine.ColonyState`——起始人口 1、全農(避免population=1、Farmers=0 的首回合饑荒,任務保守預設非手冊規則)、種族加成手動疊加(`ApplyRace` 只在開局套一次,不會回頭套用到後建殖民地)、消耗一艘殖民船、平行陣列同步),`session.go` 新增 `GameSession.PlayerColonyStars`(比照 `AIOpponent.ColonyStars`,`InvadeColony` 過戶殖民地時同步補上,先前完全沒有這個對映),`cmd/moo2/interactive.go` 加「建立殖民地」按鈕(星系主畫面,選中無主星+艦隊已抵達+載有殖民船時顯示)。**發現的架構落差**:`genPlanets` 的行星顯示字串(氣候/重力/礦產/大小)先前完全獨立於 `gamedata` 型別 enum(純展示用途),新增四個對映函式(`climateFromDisplay` 等)把玩家看到的顯示值轉成建構殖民地要用的型別值,避免兩者各算各的。**仍缺(當時)**:AI 側主動拓殖(`aiExpand` 維持先前「只標旗標、無殖民地模型」簡化,**已於下一輪補上,見下方「AI 拓殖建真殖民地」條**)、行星選擇子畫面(每星固定一顆行星,暫不需要)。測試:`internal/shell/colonization_test.go`(成功拓殖/四種前置條件擋下/拓殖後 EndTurn 經濟正常/顯示字串對映覆蓋率)。Regression 探針確認:20 回合開局 BC 軌跡不變(101→130)、拓殖後新殖民地 10 回合經濟穩定不崩潰。詳見 `docs/tech/colonization.md`、`docs/HONEST-STATUS.md`、`docs/tech/remaining-work-roadmap.md` B 項。
- [x] AI 拓殖建真殖民地(**2026-07-11 追加**):上一條的「仍缺」補上——`aiExpand` 先前只設
  `Star.Owner=2`+`OwnedStars++`,從不建立 `engine.ColonyState`,AI 殖民地數恆為開局母星 1 筆、
  `RunEmpireTurn` 的 `TotalNetIndustry` 永遠停在初始母星產出,AI 版圖擴張與經濟成長脫鉤。抽出
  `internal/shell/colonization.go` 的共用函式 `newColonyFromStar(starIdx, gov, foodBonus,
  indBonus, resBonus) (engine.ColonyState, ok, reason)`,把 `ColonizeStar`(玩家)原本內嵌的
  「氣候/重力/礦產/大小解析 → PopMax 查表 → 全農起始 → 士氣算法」搬進去,兩處呼叫端(玩家
  `ColonizeStar`、AI `aiExpand`)共用同一套建法,不再各算各的。`aiExpand` 佔星時 append 進
  `AIOpponent.Colonies` + `ColonyStars`(AIOpponent 唯二的殖民地平行陣列——不像玩家有
  Builds/ColonyBuildings/PlayerColonyMarines 等逐殖民地建造/駐軍追蹤,因為 EndTurn 對 AI 只呼叫
  `RunEmpireTurn` 結算經濟,從不呼叫那些玩家專屬的 advance* 流程,故無需同步)。**AI 政府型態
  未建模**(`AIOpponent` 無 `Government` 欄位),士氣一律用 `gamedata.MoraleGovDictatorship`
  保守預設;AI 無種族加成模型,`foodBonus`/`indBonus`/`resBonus` 一律傳 0,誠實簡化不臆造。
  維持既有「每 5 回合擴張一次」節奏不變(未改成每回合)。40 回合探針對照:修前 AI 殖民地數恆
  1、FleetStrength 線性成長(3→60);修後 AI 殖民地數隨回合增至 9、FleetStrength 加速成長
  (3→101),玩家開局 BC 軌跡兩版本一致(102→…→96),無 regression。測試:
  `internal/shell/ai_behavior_test.go` 新增 `TestAIExpand_CreatesRealColony`(佔星後建真殖民地、
  平行陣列同步)、`TestAIExpand_EconomyGrowsWithColonyCount`(殖民地數增加後軍力成長加速)、
  `TestAIExpand_NoOpWhenNoUnownedStars`(無星可擴張時安全 no-op)。詳見
  `docs/HONEST-STATUS.md`、`docs/tech/remaining-work-roadmap.md` B 項。
- [~] 領袖/軍官技能接線(**2026-07-11**):`internal/gamedata/officer.go`(`LeaderExpLevel`/`LeaderSkillBonus`,先前零呼叫端死碼)+ `formulas.go`(`LeaderHireCost`)首次真正接進遊戲。硬門檻查證:技能 id 列舉已在 `enums.go` 生成(`SKILL_ASSASSIN`..`SKILL_TACTICS`,對照 `openorion2/src/gamestate.h:602-631`),`officer.go` 原本重複定義兩個私有常數未引用完整枚舉,已清掉重複、改直接引用。openorion2 全專案 grep 確認只有 4 個技能有真呼叫端:`SKILL_WEAPONRY`/`SKILL_HELMSMAN`(艦艇命中/閃避加成)、`SKILL_FAMOUS`(雇用費修正,MIN 非累加)、`SKILL_MEGAWEALTH`(維護費全免開關);其餘 20+ 技能 openorion2 本身也沒有效果消費端(只有畫面/skillBonus 可算)。本輪只接對應到 remake 已存在系統的技能:殖民地領袖(`Ship=false`)—— demoLeaders「科學家」(`SKILL_RESEARCHER`,固定研究點)套 `ColonyState.FlatResearch`、「貿易家」(`SKILL_TRADER`,收入%)套 `ColonyState.IncomeBonusPercent`,`NewDemoSession` 建完母星後呼叫 `applyLeaderColonyBonuses` 生效(`TestGameSessionEndTurn` 研究預期值 30→55 同步更新,+25 來自科學家技能)。艦艇軍官:新增 `engine.ShipBeamAttackWithOfficer`/`ShipBeamDefenseWithOfficer`(疊加 Weaponry/Helmsman,不改既有已鎖定行為的函式簽章)+ `engine.HireLeader`(最小雇用金流)+ `gamedata.LeaderSkillTier`(讀真存檔 `save.Leader` 位元技能階,供未來取代 demo 手動 Tier)+ `LeaderMaintenanceCost`/`LeaderHireModifier`,皆為公式已就緒、等系統(remake 尚無艦艇軍官指派欄位/戰鬥解算迴圈/招募 UI)。**待人工定案**:demoLeaders「指揮官」(漢尼拔)技能標籤無唯一對應(不是 openorion2 技能表的字面詞),刻意不映射,目前無加成;「工程師」(圖靈)映射到 `SKILL_ENGINEER` 清楚,但真實效果(艦艇維修率)remake 無承接系統,標 TODO。詳見 `docs/tech/leader-officer-skills.md`。測試:`TestLeaderSkillTier`/`TestLeaderMaintenanceCost`/`TestLeaderHireModifier`(gamedata)、`TestShipBeamAttackWithOfficer`/`TestShipBeamDefenseWithOfficer`/`TestHireLeader`(engine)、`TestApplyLeaderColonyBonuses_*`/`TestLeaderDisplayLevelToExpLevel`/`TestLeaderSkillIDByNameMapping`(shell)。標 `[~]` 非 `[x]`:只接了 2/25+ 技能的真實效果,多數技能因 openorion2/手冊本身無精確效果定義而 TODO,不算完整。

## Phase 6 — 音樂 / 音效
> 第一性原理翻案(2026-07-10):MOO2 **沒有 XMI/MIDI 音樂**,全部是 LBX 內的 22050Hz 8-bit PCM WAV。故無需 SoundFont/OPL 合成——原封播原版 PCM 即 bit-identical。研究定案見 `docs/tech/audio-format.md`。
- [x] ~~逆向 .lbx 音樂(XMI)格式~~ → 實為 PCM WAV,存 STREAM/STREAMHD.LBX(格式研究文件已定案,含 provenance)
- [x] 逆向音效格式 → SOUND.LBX 內 WAV;entry0 為 20-byte 名稱表(BUTTON1…),已解出 68 個具名音效
- [x] ebiten 音訊播放整合 — `internal/audio`(WAV 解碼→16-bit stereo、Mixer BGM 迴圈+SFX;headless 停用避免無音效卡崩潰)+ 單元/真檔測試綠
- [x] 接線:主選單 BGM(STREAMHD)+ 按鈕點擊音效(BUTTON1)— `cmd/moo2/audiohook.go`
- [x] 曲目/UI 事件對應(2026-07-10 定案到靜態溯源極限):外交樂反組譯硬證(track 13/14/15);menu/galaxy/combat 對應 Play 函式在 DOS build 為死碼,維持時長啟發式(誠實標,再定案需聆聽或 Windows build RE)。見 `audio-track-map.md` 第七節
- [ ] `CMBTSFX/SPHERSFX` 巢狀音庫格式逆向(戰鬥期音效)
- [x] ~~SoundFont 處理~~ → 不需要(無 MIDI 音樂)
- [ ] 桌面實測驗收:使用者對原版聆聽比對(主選單 BGM + 點擊音是否為正確曲/音)

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
