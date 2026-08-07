# WORKLIST — 銀河霸主2 go/ebiten 重製 + 繁中化

> ⚠ **勾選狀態會過期,以程式碼為準**(rule 63)。這份清單橫跨數十輪,底下 Phase 0–4 有不少
> `[ ]` 其實早就做完了(2026-08-07 已修掉一批可核實的)。要判斷「某項做了沒」,
> **先 grep 程式碼,別信這裡的方框**。各輪的實際成果以 `docs/re/01-gap-report.md`
> 與 `git log` 為準。

> ⚠ **「還原度 20%」是 2026-07-04 接原版美術前的過期快照,勿再當現況錨點。** 現為可玩的多帝國 4X 迴圈;詳見 [`docs/HONEST-STATUS.md`](docs/HONEST-STATUS.md)。
> 下方大量 `[x]` 中,gameplay 子系統類仍須以「對原版實測」重評數值對齊度(rulebook 65),不因測試綠自評還原度。
> 優先進度(2026-07-12 使用者確認):**①音樂/音效 已完成**(PCM 播放 + 場景 BGM 切換 + 按鈕音效;曲目↔場景精確身分待人耳定案)、**②忠實新遊戲流程 已完成**(版本選擇→難度/星系→14 族肖像選擇→自訂點數→命名旗色→真母星;turn-1 少數數值待 playtest 校準)、**④gameplay 子系統全接**(殘留=原版 oracle 校準數字)。
>
> ⚠ **2026-08-07 翻案**:上面第 ③ 項原本寫「像素對齊已做到 openorion2 oracle 上限,無對應 view 的
> 畫面只能維持 PIL(估計座標)」,並據此推論「至此無純自驅剩項,剩全是 oracle-gated(DOSBox)」。
> **兩句都錯。** 真正的一手座標在**原版執行檔的反組譯**裡,與 openorion2 有沒有實作那個畫面無關。
> 改用「反組譯立即數 → openorion2 `initWidgets` → LBX 資產尺寸交叉驗證」之後,先前被判為
> 「只能估計」的畫面(新遊戲設定 / 殖民地 / 種族選擇 / 艦艇設計 / 地面戰 / 多人設定 / 熱座交接)
> 全部重挖完成,而且挖出三個**只看畫面看不出來**的還原錯誤(新遊戲左下框是 PLAYERS 不是 RACE、
> 種族選擇左右擺反、艦艇設計六個艦體槽不等距)。詳見 `docs/re/01-gap-report.md`。
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

## ★ 2026-07-11(續 session:字型打磨 + game test 收尾 + 多人考據)
> 本段索引本 session 的完成項;細節見各 commit 與 docs/tech/。
- [x] **點陣中文字型改造(Stage 1/2,使用者指定「一致化 + 點陣字」)**:中文 UI 由 Noto 平滑向量改
  `bitmapfont/v4 FaceTC`(Cubic 11 + Ark Pixel,OFL)。最終為**混合字型**:內文(<18px)點陣、標題
  (≥18px)Noto 向量(避免點陣放大鋸齒);**主選單**整個純 Noto(使用者偏好)。可行性經全語料窮舉
  驗證(2258 字缺字 0)+ 覆蓋回歸測試(墨點判準,守未來新增字)。殖民地擦底改採按鈕面色
  (`plateFace`,修「黑塊蓋浮雕框」)+「已建:…」建築清單依欄寬截斷(`truncateToWidth`,修溢出建造欄)。
  commits `ea8821b`/`5bdb78b`/`4d84146`;設計 review `3589b9e`;決策 `docs/tech/pixel-font-decision.md`、
  `docs/tech/ui-typography-button-review.md`。
- [x] **戰鬥畫面對手改用真實 AI 名**(修硬編「賽隆人」stale 標籤,`PrimaryEnemyName` 取真 AI 種族名),`4a58665`。
- [x] **game test 全面實機驗證 + 修 GUI bug**(task 37/38/39):xvfb GUI 玩家路徑 8 畫面截圖 + 深度回合探針;
  修 3 個 GUI bug(研究盲選/殖民地面板/戰術糊字)、補殖民地總覽 **Empire Summary 面板 + Planetary/
  Production Info 懸停面板**(`c3540ee`,即 /goal 的「game test 回報問題」)。邏輯層 70 回合無 panic、
  三勝利路徑可達。
- [x] **多人對戰通訊考據 + Phase 9 開項**:原版 CD 手冊 OCR 出通訊方式(序列/數據機/IPX 區網/TEN),
  架構=決定性 lockstep;方向定案「保留 lockstep、傳輸換 TCP、先做熱座」。`docs/tech/multiplayer-architecture.md`、
  `d31d182`(見下方 Phase 9)。

## ★ 2026-08-07(反組譯對齊版面 + 英文模式 UI 層)
> 細節見 `docs/re/01-gap-report.md`、`docs/HONEST-STATUS.md`。

- **版面來源改成反組譯**:確立優先序「exe 立即數 > openorion2 `initWidgets` > LBX 資產尺寸交叉
  驗證 > 量圖」。用它重挖了新遊戲設定、殖民地、種族選擇、艦艇設計、地面戰、多人設定、熱座交接。
- **修掉三個看畫面看不出來的還原錯誤**:新遊戲左下框是 PLAYERS(不是 RACE,對手數先前寫死 3)、
  種族選擇左右擺反(原版是左肖像 + 右 2×7 網格)、艦艇設計六個艦體槽**不等距**(15/14/17/14/14/16)。
- **接上先前選了沒作用的設定**:星系年齡(先前被常數寫死成「普通」)、帝國總數 2–8、
  難度五級(補回教學)、起始科技等級的第一個真效果(曲速前開局沒有 FTL,艦隊出不了本星系)。
- **多人**:熱座(同機輪流)可玩 + 主選單 MULTI PLAYER 接上原版設定畫面。
- **英文模式 UI 層**:`cmd/moo2` 自繪畫面全部雙語。優先「讓路」(底下是原版美術就露出烘在圖上的
  英文),其餘用 `b.tr(zh, en)`。護欄是 `lang_gap_test.go`(go/ast 棘輪)。引擎層字串未做。
- **行星表面格點解鎖**(gap report 第 27 項):先前記成「獨立工程、卡在幾何表沒抽出來」,
  這輪把 `word_182C9C` 的 7×7 角點表抽出來,並用 `COLONY.LBX#8` 那張**已畫好位置**的高亮
  菱形逐點驗過(表來自程式碼、菱形來自美術檔,兩個獨立來源)。走訪順序來自 `dword_BA784`
  (遠→近,畫家演算法)。另發現建築 sprite 是**已畫好位置**的 640×480 稀疏圖
  (資產 = 型別×36 + 格號,共 49 種),貼 (0,0) 就對位。
  建築編號對照同日補齊:算式來自 `Cache_Load_Bldg_` @ 0xAF6DC,編號由 openorion2 的
  `BUILDING_*` 列舉與原版 `TECHNAME.LBX` 第 295 條起的字串順序互相對上。
  離線合成 9 棟端到端驗證通過(每棟正落在自己的格子裡、遠近遮擋正確)。
  落在 `cmd/moo2/colonysurface.go` + 9 條回歸測試。
- **原版建築表挖出來,建造成本全部換真值**(gap report 第 28 項):`Real_Building_Name_`
  兩行組語就指出資料段有 49 筆 × 19 位元組的建築表。+8 = 建造成本 PP、+12 = 維護費、
  +14 = 分類(7 = 衛星)。驗證方式:Armor Barracks 的 +8 = 150 對上先前唯一有實據的那筆,
  +12 與手冊維護費 **40/40 全中**(獨立來源,能全中代表欄位判讀與建築編號對照都對)。
  `EstimatedCost` 欄位整個拿掉——40 項全是真值。舊估計錯得不小:星辰要塞 800→**2500**、
  行星屏障護盾 1200→500、歡樂穹頂 800→250、核心廢料場 550→200。
- **地表擺放規則已讀完**(同第 28 項):`Set_Random_Seed(colonyIdx, 0, 144)` → 每個殖民地的
  擺法固定;逐棟蓄水池抽樣選空格;房屋借用衛星編號 3/14/40/41 當四種外觀。
  (原版 PRNG 已於同日實作,見下方第 31 項。)
- **半透明標記索引**(gap report 第 29 項):原版對來源索引 >= 0xF0 的像素**從不直接寫進畫面**
  ——`Draw_` 走混色查表、`Draw_No_Glass_` 整個跳過。240..255 是十六種半透明標記不是顏色。
  remake 先前一律當顏色上色,建築陰影因此變成洋紅。`internal/lbx` 加了
  `TranslucentIndexMin` / `HasTranslucent` / `ToRGBADropTranslucent`(= 跳過那條路徑),
  既有 `ToRGBA` 行為不變。**不能在解碼層一刀切當透明**:BEAMS 有 69% 是這種索引,
  丟掉光束就消失了。混色表在 BSS(執行期才建),產生它的程式碼還沒找到,不假造係數。
- **殖民地畫面中段還給行星表面,建造佇列搬回原版彈出視窗**(gap report 第 30 項):
  `cmd/moo2/buildqueue.go` = `Build_Queue_Popup_` @ 0xB4041(框架 COLBLDG.LBX#0,可建清單
  x 13..184 / y 20+19i、佇列 7 格 x 207..458 / y 329+20i、六顆鈕座標全是反組譯真值);
  入口是框架上那顆 CHANGE(原版它就是「換要蓋什麼」,先前畫成灰的沒接)。
  中段改畫行星表面(當時只有格線 + 建築圖)。
  中文模式 17 張逐像素比對:只有殖民地畫面變了。
- **原版 PRNG + 原版擺放演算法 + 地表底圖**(gap report 第 31 項,同日):
  - `internal/gamedata/origrand.go` = `Random_` @ 0x1247A0:32-bit LCG(`0x41C64E6D` / `0x3039`)
    + 拒絕取樣,回傳 **1..n**。戰鬥、事件、星系生成日後都能靠它更接近原版。
  - 擺放換成 `Make_Bldg_Array_For_Colony_` @ 0xBC30B:`Set_Random_Seed(colonyIdx)` → 建築
    依編號順序蓄水池抽樣選空格(分類 7 的衛星排除)→ 房屋 人口/3+1 → 依分類氣泡排序。
    只有最後那段隨機微調沒做。
  - 地表底圖:`C_Anims` @ 0xBBA8E 的跳表解出 **COLONY2.LBX#49**(星空)+ **PLANETS.LBX**
    第 `氣候×3 + 變體` 張(地形)。PLANETS 恰好 30 張 640×480 = 10 氣候 × 3 變體,
    與 `gamedata.PlanetClimate` 逐項相同(#27 = Gaia 蔥綠、#3 = Radiated 熔岩)。
    佔位格線移除——原版地表上沒有格線。
  - ⚠ **抓到一個沒有症狀的軸向錯誤**:格陣索引與角點表對調 → 整張佈局鏡射,徵狀只有
    「建築擠在遠端」。定案證據是 `Add_Bldg_Fields_` 同時用 (v1,v2) 當位址與座標參數。
    護欄 `TestColonyGridKeyMatchesOriginalAddressing`。
  - **軌道衛星**(`Draw_Colony_Satellites_` @ 0xBE366):x = 295 ± i×50、y = 162,
    圖是 COLONY.LBX 9/10/11/12/16。⚠ 編號要經過 `sub_BBB9F` 的 `add edx, 9` —— 漏掉會去讀
    資產 0..4,那五格在檔案裡是**零長度**的,畫面上什麼都不會出現而且不報錯。
    抑制規則 `sub_BC21B` 就是原版的星基 → 戰鬥站 → 星際要塞升級鏈。
  - 仍缺:建築集合與原版有差(Colony Base、已完成的一次性改造沒建模)→ 落點尚未對原版實測、
    植被層沒做、地形變體那一欄用 PRNG 代替原版的存檔欄位(不是真值,已標明)。
    擺放微調與母星國會大廈已於第 42 項補上。
- **星圖:艦隊圖示 + 旗色順序**(gap report 第 32 項,同日):
  - 艦隊圖示換成原版的(`Get_Ship_Icon_Pict_Seg_`:BUFFER0.LBX `205 + 旗色×4 + 縮放`),
    取代先前那個 8×8 青色方塊。remake 星圖沒有縮放,固定用縮放 0(最小那張)。
  - ⚠ **旗色順序修正**:先前是 紅/黃/綠/**藍/白/紫/橙/棕**,後五個全錯位——而艦隊圖示是
    `205 + 旗色×4`,所以選藍色會開出白色的艦隊,中文模式完全看不出來。兩個獨立來源對上
    才改(BUFFER0 每組實測代表色 + openorion2 `FONT_COLOR_PLAYER_*`):
    正確順序是 **紅/黃/綠/銀/藍/棕/紫/橙**(第 4 色原版叫 SILVER 不是 White)。
  - 拿掉星圖左上那行「研究:<主題名>」——原版星圖沒有這東西,而且它壓在星星與艦隊圖示上。
    右欄第 5 格(綠燒瓶)改放每回合研究點數,跟其他四格一致。
  - 順帶把整張星圖的 9 層繪製順序抄進 gap report,標出 remake 各層的完成度。
- **星圖視差星空背景**(gap report 第 33 項,同日):星圖區先前是一片純黑佔位,
  現在鋪的是原版 `Draw_Paralax_` 的三層星空(STARBG.LBX 資產 0/1/2,各 640×480)。
  原版三層各自以不同速率捲動並環繞平鋪;remake 星圖不捲動,位移固定 0 → 三層各貼一次。
  - ⚠ **底色不能省**(踩過):三層全是稀疏的暗點疊在透明上,底色一拿掉透明處就露出底下的
    框架美術,整片星圖變成白底黑點。原版也是先 `Fill` 再貼視差層。
    這條規則**測不到**(ebiten 在 game loop 外不准 `Image.At()`),只能靠截圖驗收。
  - 順帶修掉 `decodeAsset` 對 nil resolver 的 nil 解參考(三個新 helper 都補了守衛)。
  - 仍缺:星雲需要銀河生成先產出星雲表(資料模型缺口,不是繪圖缺口);蟲洞連線、遷移連線、
    星門、外交燈號同理。
- **星圖:星球換成原版 sprite**(gap report 第 34 項,同日):先前是自畫色圓。
  資產 = `148 + 光譜×6 +(縮放 + 大小)`,**三個獨立來源互相印證**(反組譯
  `Get_Star_Picture_Seg_`、openorion2 `_starimg[class][zoom+size]`、實測 BUFFER0 那 36 張)。
  - 公式自己證明了自己:光譜 6(黑洞)不加大小 → `148+36+縮放` = **184+縮放**,
    正好等於 openorion2 另外命名的 `ASSET_GALAXY_BHOLE_IMAGES 184`。
  - ⚠ 縮放在原版是**銀河尺寸的函式不是玩家控制**(銀河越大畫越小)。remake 把座標正規化
    攤滿視窗、沒有捲動,那個對應接不過來,**不存在忠實值**;固定用 3(最縮小),
    這是 remake 的選擇不是原版真值。
  - 仍缺:5 幀閃爍動畫、黑洞的獨立繪製路徑與阻擋航線機制。
- **蟲洞**(gap report 第 35 項,同日):第 33/34 項的結論是「剩下的層卡的是資料模型」,
  這一項去補模型。蟲洞是其中最有遊戲價值的——它是**機制**不是裝飾。
  - `Star.Wormhole`(-1 = 無,對應原版 +0x29),**必須雙向**(openorion2 對單向直接丟例外)。
  - 產生照抄原版拿得到的部分:母星/黑洞不可當端點、最多 200 次拒絕取樣、最短距離門檻、
    候選收滿 19 個就停。⚠ **數量與距離門檻的單位沒照抄**——原版數量是銀河產生過程中累加的
    (上限 `galaxySizeParam×4+4`),接不過來;改用同構的 `星數/8` 夾 1..4。非原版真值。
  - 星圖第 1 層畫連線(在星球之前),兩端都沒偵測到就不揭露。
  - `SendFleet`:兩端有蟲洞 → ETA 1(手冊 p.181「a single turn」)。
  - ⚠ **別和隨機事件的蟲洞搞混**——MOO2 兩者都有,remake 也是(`applyWormhole` 早就有)。
  - ⚠ **舊存檔沒有這個欄位,零值是 0**,會讓每顆星都宣稱與星 0 有蟲洞。讀檔走
    `normalizeWormholes`,有專門的回歸測試。
- **把「原版 48 棟 vs remake 40 棟」的差集查清楚**(gap report 第 36 項,同日):
  8 個沒建模的編號逐一認出來並從原版建築表讀真值。結論:2 個是自動給予(Capitol /
  Colony Base,正確地不該在表裡)、3 個已是 `SpecialActions`、**3 個真的缺**
  (Galactic Currency Exchange 250PP/3BC、Stellar Converter 1000PP/6BC、
  Artificial Planet 800PP/0BC)。
  - 順帶驗證:三個 SpecialAction 的成本與重抽逐項相同;且**維護費 0 = 一次性**這條規則
    正好把 8 個編號分成「常駐建築」與「一次性」兩堆。
  - ⚠ **抓到一條死路科技**:`TOPIC_GALACTIC_ECONOMICS`(6000 RP)解鎖的
    `TECH_GALACTIC_CURRENCY_EXCHANGE` 沒有任何東西消費它——研究完什麼都不會發生。
  - 這輪**沒有直接補上**:效果查不到(手冊沒這條、patch 手冊零命中、遊戲資料檔只有名字沒說明、
    建築表沒有效果欄),而不編數字是紀律。下一輪的路線寫在 gap report 第 36 項。
- **恆星轉換器(行星版)補上 + 一個乾淨的負面結果**(gap report 第 37 項,同日):
  走完第 36 項留下的路線。三棟裡做成一棟,另兩棟拿到「為什麼做不成」的證據。
  - **先做正對照再下結論**:把 48 個「已建」旗標位移全掃一遍——43 個出現過(技術有效),
    完全沒出現的只有 5 個,而那 5 個正好是 2 衛星 + 3 一次性,與「維護費 0」那條規則
    挑出來的是同一批。**兩個獨立判準指向同一個分組。**
    ⚠ 建築 18 的 `+0x148]` 全檔只出現一次,基底卻是科技陣列不是殖民地——同一個數字出現在
    不同結構裡,這正是不能直接數命中的理由。
  - **恆星轉換器**:手冊 p.106(400 傷 ×2、維護 6)與原版建築表第 42 列(1000 PP、維護 6)
    **維護費逐項相符**才建模。接進 `colonyDefense`(+800)與 `origBuildingID`(42)。
  - ⚠ **推翻一個把記帳慣例當成規則的斷言**:`TestBuildingsCount` 原本釘死 40,理由是
    「手冊全表 35+5」——但那是手冊把恆星轉換器另立一節的記帳方式,原版表裡它就是第 42 棟。
    **40 從一開始就不是原版的數字。** 現在 41。
  - (⚠ 「飛彈基地/地面砲台要等空間模型」這句**當天就被自己推翻了**,見下一項。)
- **飛彈基地/地面砲台接進防禦解算**(gap report 第 38 項,同日,第 37 項的訂正):
  空間模型早就在(`satellite.go` 的 300/450 都是手冊確認值,`retaliationAttackers` 也早就
  支援這兩棟)——缺的只是 `colonyDefense` 沒接上,它用的是自編的
  `CommandPointsFromBuildings × 10`。
  - ⚠ **挖出一個自相矛盾**:同一座星基在 `colonyDefense` 值 10(比巡洋艦 8 還強),
    在 `retaliationAttackers` 值 3–4(≈ 驅逐艦 tier)——而 `satellite.go` 的校準註解明講
    是後者。**兩邊都測得綠**,因為沒有任何測試同時看這兩條路徑。現在有了。
  - 改用同一套推導後:反擊戰力隨武器科技成長、兩棟防禦建築真的有用、1.3/1.5 的 arc-cost
    差異自動吃到。
  - `TestAIRaidRepelledByFleetAtStar` 的自我守衛**正確地觸發了**(防禦 19 < 願打門檻 21),
    處理方式是把測試裡的母星升級成戰鬥站,不是把模型改回去遷就測試。
  - IDA 的 `-Ohexrays` 這一輪起不來(error code 4,兩次),整項改用手讀 `.asm` 完成。
- **兩個零值陷阱**(都加了回歸測試):`TechLevel` 零值 = 曲速前 → 舊存檔艦隊全凍住;
  `i18n.English` 是 `Lang` 的零值 → 忘了設 lang 的建構路徑會靜默變英文。

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
- [x] README(含致謝)——`README.md` §致謝
- [x] 本機 git commit(push 待使用者確認)

## Phase 1 — 資料層移植(純 Go)
- [x] Go module 初始化 + docker build 環境(`go.mod`、`docker/`、`scripts/build.sh`)
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
- [x] CJK 渲染:`internal/uifont`(ebiten text/v2,依尺寸快取 face)+ Measure。**2026-07-11 升級為混合字型**:zh 內文用 `bitmapfont/v4 FaceTC` 點陣(<18px)、標題用 Noto 向量(≥18px);en 純 Noto。見 `docs/tech/pixel-font-decision.md`
- [x] 顯示層覆蓋 i18n:`internal/i18n`(TSV 英文即 key + 查無 fallback + TranslateFormat)+ 測試
- [x] [HARD] 只翻顯示層,不動資料層(i18n 設計即如此)
- [x] 字型:NotoSansCJK-Regular.ttc 經 Go opentype.ParseCollection 驗證可解析+量測中文(★ [HARD] 相容檢查通過);galaxy 標題已渲染繁中
- [ ] 繪字描邊/陰影版 + 逐字斷行(目前基本 Draw/Measure;混合字型上線後標題走平滑 Noto、內文點陣暗底可讀,描邊需求降為次要,待「字疊亮星點」處再補)
- [ ] 字型子集 pyftsubset(docker)+ go:embed 內嵌(待譯文集齊;目前用完整 .ttc runtime 掛載)
- [x] 主選單中文化 + 截圖校對(cmd/moo2 -menu:擦底疊字六按鈕繼續/載入遊戲/…;before/after 見 docs/reference-screens.md)
- [x] 主選單:語言 中/英 runtime 切換(2026-08-07,`cmd/moo2/interactive.go` 的 `toggleLang`)
      ——先前只有啟動旗標 `-lang`,進了遊戲換不掉,不符 `CLAUDE.md` 那條需求
- [x] **英文模式覆蓋率(UI 層)**:2026-08-07。切換機制早就有,內容沒補齊——overlay 那條路徑
      天生雙語(英文模式跳過擦字疊字、露原版美術),remake **自繪**畫面則寫死中文。
      這輪把 `cmd/moo2` 整層收掉:優先「讓路」(遊戲選單/多人設定/種族選擇/載入儲存/
      戰術控制列 → 直接露原版烘在圖上的英文),其餘用新建的 `b.tr(zh, en)`。
      順帶修掉兩個中文模式看不出來的 bug(RACESEL#33 標題橫幅從沒被畫過、
      `diplomatRaceIndex` 與艦體名各有一份會漂移的重複對照表)。
      護欄:`lang_gap_test.go`(go/ast 棘輪,只能往下調)+ `lang_coverage_test.go`(漏填英文欄)。
- [ ] **英文模式覆蓋率(引擎層)**:`internal/` 仍有約 690 條引擎產生的中文顯示字串
      (`session.go` 228、`gamedata/buildings.go` 83、`colonization.go` 61、`events.go` 44…),
      外加星名池 829 + 艦名池 672(原版英文原文在原始資料裡,接得回來)。
      症狀:英文模式下星名、建築名、行星屬性值、熱座席位名仍是中文。
      正解是讓引擎回鍵值、UI 端繪字,不是把 `tr()` 灑進引擎。
      逐檔清單、量法與收尾原則見 `docs/HONEST-STATUS.md`「英文模式覆蓋率」
- [x] 主選單:版本 1.3/1.5 選擇框架(`toggleVersion`,左下角)
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
- [x] 依 `palette-chain.md` 對照表逐畫面上色——機制是 `resolvePalette` + `paletteChain`,各畫面都在用。剩 COLGCBT(地面戰 sprite)的來源未定案,見 `cmd/moo2/groundcombat.go` 檔頭

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
- [x] ~~補齊需全域調色盤鏈的畫面到對照組~~ 這些畫面在遊戲裡早就跑起來了;`docs/reference-screens.md` 的靜態對照組收錄落後於實作,已在該文標明
- [ ] **[HARD] 開工先做:窮舉所有文字源(LBX 各類 + Go hardcode),各寫 dumper,用引擎自己 reader dump 精確 key**
- [~] 逐畫面重建:主選單/星系圖/行星清單/殖民地/科技研究/艦隊/軍官/種族資訊/對話框皆已建;**只剩「載存檔」畫面**(LOADSAVE.LBX 全 repo 零引用,主選單 Continue/Load 目前無存檔選單——原版 oracle 對照的 issue #2 仍開著)
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
- [x] 戰鬥:格子戰術戰鬥(2026-07-10 換原版美術:STARBG 星空+COMBAT 控制列+可見 CMBTSHP 艦艇+控制列 7 按鈕中文化;逐發用真 ResolveShot 命中/傷害/過盾/過甲);宣戰→戰術戰鬥→戰鬥結果。**(2026-07-11 更新:武器依 beam/missile/spherical 分流,飛彈躲避/AMR/球狀傷害公式接進解算,見 `tactical-combat-weapon-kinds.md`)**。**艦型 sprite 對照已接(task 12,2026-07-11)**:網搜定 CMBTSHP 色塊結構(8 色×45)+ 視覺比對定尺寸,戰鬥依艦級顯示不同大小 sprite、玩家/敵艦不同色塊,取代單一 placeholder(近似對照非原版精確 picture 映射,見 `docs/tech/cmbtshp-ship-sprites.md`)
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
  併入 `NetBC`。當時(本條寫下時)本專案艦種塑模(`gamedata.ShipType`:`COMBAT_SHIP`/`COLONY_SHIP`/
  `TRANSPORT_SHIP`/`OUTPOST_SHIP`)沒有獨立的「Freighter」艦種,呼叫端恆傳 0,目前 no-op,接線先備妥。
  ★ **此缺口已於同日稍後補上(見上方 Phase 7 §「#4 運輸艦淨現金版本差異」條)**:新增「運輸艦隊」
  建造選項後,玩家側 `ActiveFreighters` 真的變非 0,維護費隨之生效,並補上 1.3/1.5 版本現金加成
  差異。③**士氣對收入的調整**
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
- [~] **武器改造(Weapon Modifications)mod 系統(2026-07-11 第三輪)**:艦艇武器先前只是名字字串,無改造。硬門檻查證:`moo2_patch1.5/GAME_MANUAL.pdf`(`pdftotext -layout`)p.115-118「Modifications」章節逐字給出 8 個光束/通用 mod 的精確佔格/成本/效果數字(HV +100%/150%傷害/射程懲罰減半、PD -50%/半傷害/+25%命中/射程懲罰加倍、AF 固定+50/3連發each-20%命中、CO +50%/+25命中、AP +50%/穿甲、ENV +100%/四倍傷害、NR +25%/消除射程衰減、SP +50%/穿盾),openorion2 全專案 grep 零命中(純渲染殼,無 mod 邏輯可抄,見 memory `openorion2-is-renderer-not-engine`)。新增 `internal/gamedata/weapon_mods.go`(`WeaponModCode` + 8 個 mod 常數 + 佔格/成本/命中/傷害公式,逐項附手冊頁碼);接線:`internal/shell/session.go` 的 `ShipDesignSpaceUsedWithMods`/`ShipDesignFitsWithMods`/`DesignCostWithMods`/`BuildShipWithMods`(佔格/超格擋下/成本,原函式保留、委派 nil mods 回歸)、`combat_formula.go` 新增 `ResolveShotWithMods`(命中/傷害,`battleVolley` 與 `cmd/moo2/interactive.go` `fireRound` 共用同一份解算)、`shell.Ship`/`CombatShip` 新增 `Mods []string` 欄位(直接被既有 `sessionSnapshot` 序列化,免改 persist.go)、艦艇設計畫面新增 8 個 mod 勾選 chip(HV/PD 互斥自動切換、非 beam 武器提示不支援)。**回歸保護**:`DamageMountAdjustedValue` 對「命中傷害恆最少 1」的夾限會把「無武裝」(0 傷害)誤夾成 1,`ResolveShotWithMods` 在無 mod 時完全跳過該函式呼叫,逐位元對齊加入 mod 系統前的行為。20 回合開局 BC 探針(100→130)與既有基準一致,無 regression。標 `[~]` 非 `[x]`:NR 目前沒有可觀察傷害效果(現行戰鬥模型本身沒有射程傷害衰減可消除)、飛彈/魚雷專屬 mod(ARM/ECCM/EMG/FST/MV/OVR,手冊有數字但飛彈解算未接 mod 掛鉤)、小型化等級門檻未建模、火線角(Firing Arc)未接線。測試:`internal/gamedata/weapon_mods_test.go`、`internal/shell/shipspace_test.go`/`combat_formula_test.go` 新增案例(佔格倍率/超格擋下/HV-PD-AP-ENV-CO-AF 逐一驗證/無 mod 回歸)。詳見 `docs/tech/weapon-mods.md`。

## Phase 6 — 音樂 / 音效
> 第一性原理翻案(2026-07-10):MOO2 **沒有 XMI/MIDI 音樂**,全部是 LBX 內的 22050Hz 8-bit PCM WAV。故無需 SoundFont/OPL 合成——原封播原版 PCM 即 bit-identical。研究定案見 `docs/tech/audio-format.md`。
- [x] ~~逆向 .lbx 音樂(XMI)格式~~ → 實為 PCM WAV,存 STREAM/STREAMHD.LBX(格式研究文件已定案,含 provenance)
- [x] 逆向音效格式 → SOUND.LBX 內 WAV;entry0 為 20-byte 名稱表(BUTTON1…),已解出 68 個具名音效
- [x] ebiten 音訊播放整合 — `internal/audio`(WAV 解碼→16-bit stereo、Mixer BGM 迴圈+SFX;headless 停用避免無音效卡崩潰)+ 單元/真檔測試綠
- [x] 接線:主選單 BGM(STREAMHD)+ 按鈕點擊音效(BUTTON1)— `cmd/moo2/audiohook.go`
- [x] 曲目/UI 事件對應(2026-07-10 定案到靜態溯源極限):外交樂反組譯硬證(track 13/14/15);menu/galaxy/combat 對應 Play 函式在 DOS build 為死碼,維持時長啟發式(誠實標,再定案需聆聽或 Windows build RE)。見 `audio-track-map.md` 第七節
- [x] ~~`CMBTSFX/SPHERSFX` 巢狀音庫格式逆向~~ **(2026-07-11 前提翻案,rulebook 62/63)**:CMBTSFX/SPHERSFX **不是音效庫,是戰鬥視覺特效動畫**(79 資產,爆炸/光束/護盾命中多幀 sprite,標準 LBX 影像,`lbxinfo` 直接解得);戰鬥**音效**全在 SOUND.LBX(68 具名音效已解碼含 NRGBLAST/PHOTON/TORPDO1/EXPL/SHIPHIT1/SHIELD…)。見 `docs/tech/audio-format.md`
- [x] 戰鬥音效接線:SOUND.LBX 的 NRGBLAST/MISLFIRE/SHIPHIT1… 已接進戰術戰鬥(`cmd/moo2/audiohook.go`)
- [ ] (選)CMBTSFX 爆炸/光束特效動畫接進戰術戰鬥畫面(視覺增強)
- [x] ~~SoundFont 處理~~ → 不需要(無 MIDI 音樂)
- [ ] 桌面實測驗收:使用者對原版聆聽比對(主選單 BGM + 點擊音是否為正確曲/音)

## Phase 7 — 版本 1.3 / 1.5(2026-07-11 大幅推進)
- [x] 研究「1.3 → 1.5 規則差異清單」:逐條核對 1730 行 CHANGELOG_150 + MANUAL_150 + PARAMETERS.CFG,`docs/tech/version-1.3-1.5-diff.md`。結論:落在已實作系統的真差異只 3 個(多數 CHANGELOG 是 bug fix 或「新增可調參數但預設=經典值」)
- [x] rule profile 資料結構:`gamedata.RuleProfile` + `GameVersion` + `Profile13/15`(`internal/gamedata/ruleprofile.go`)
- [x] 1.3/1.5 profile 實作 + 驗證(值 + 預設 Profile15=現行 三層回歸斷言)
- [x] 主選單版本切換生效(**2026-07-11 收尾完成**):UI + 開局注入 + **三個 live 消費端全部接線**——①軌道轟炸齊射(1.3=5/1.5=10)②電漿砲傷害(`BuildShipWithMods` 改讀 `BuildWeaponOptions(s.RuleProfile)`,造艦定案值隨 `Ship.WeaponAttack` 帶進 `ResolveBattle`/`StartCombat`)③超先進科技研究成本(`engine.PlayerState.HyperAdvancedResearchCost`,`RunResearchPhase` 對 Hyper 主題套版本覆寫,`EndTurn` 對玩家+AI 同步注入避免規則不對稱;顯示層 `ResearchCostForDisplay` 帝國概況/研究選擇畫面一致)。測試:`BuildWeaponOptions(Profile15)` 逐元件等於套件級(behavior-preserving)+ 電漿傷害/研究成本隨 profile 變動斷言
- [~] **diff 全量表未實作系統補上(使用者指定,近似公式)**:批次 A(地面戰/轟炸 #5/#6/#7/#8/#9/#11)已做——#6 攻方指揮官倍率 + #11 行星尺寸轟炸幾何**完整接線**;**#7 建築+1hit 已於 2026-07-11 隨「AI 建築模型+軌道防禦」子系統接線**(見下);#5 守方指揮官2.5x/#8 civilian_armor(HP模型)/#9 防禦裝甲**仍因資料模型缺口(AI 無 Leaders、remake 採 hits 計數非 HP 模型)只做公式+版本欄位+TODO 掛鉤**(`docs/tech/ground-combat-algorithm.md`、`gamedata/ground_version_diff.go`);**批次 C(#13 掃描/偵測距離)已於 2026-07-11 完成**(見下)。diff 全量表 15 項至此全數盤點完畢。
  - [x] **AI 殖民地建築資料模型 + 軌道防禦吸收轟炸(2026-07-11,使用者選定方向)**:`AIOpponent` 新增 `ColonyBuildings []map[string]bool`(第三平行陣列,`buildDemoAIOpponents` 母星初始化為 `homeworldBuildings()` 獨立拷貝、`aiExpand` 新殖民地 append 空 map、`InvadeColony` 攻陷同步移除,三處維持等長;`persist.go` `aiSnapshot` 同步序列化,舊存檔 nil 安全降級)。`BombardColony`:bomb hits 先依建築名字母序(決定性、不吃 rng)摧毀建築(每棟耗 `GroundPlanetHitsPerBuilding + RuleProfile.BombardmentBuildingBonusHits`,**#7 版本接線**:1.3 每棟多 +1 hit),餘數 hits 才扣人口——軌道防禦讓人口在防禦被轟掉前受保護。`GroundBombardResult` 新增 `BuildingsDestroyed`/`BuildingsRemaining`。nil/空建築 map 逐位元回歸舊行為。測試:更新 `TestBombardColony_ReducesPopulationDeterministically`(母星2建築吸2 hits→popLoss 8→6)+ 4 新測試(建築先於人口吸收/餘數轉人口/#7 版本差異摧毀棟數不同/nil 回歸);moo2sim 20 回合軌跡與改動前逐字元相同(不動 EndTurn 經濟)。**本輪不做**防禦方反擊摧毀玩家艦(下一輪)。誠實簡化:#7 語意近似(CHANGELOG 原句模糊)、摧毀順序字母序非原版、入侵過戶不轉移建築、#8 HP模型未調和。
  - [x] **#14 衛星/軌道防禦基地「space 預算武器平台」+ 版本相依 beam arc-cost(2026-07-11)**:`internal/gamedata/satellite.go` 新增獨立衛星/基地 space 預算(飛彈基地 300、地面砲台 450——手冊 p.78/p.81 確認值;星基/戰鬥站/星辰要塞 250/500/1200——借用 `ShipHullSpace` 同量級近似值)+ arc-cost 佔格公式(比照 `WeaponSpaceWithMods`)+ fit 公式;`RuleProfile` 新增 `SatelliteBeamArcCostPct`(1.3=25/1.5=33)、`GroundBatteryBeamArcCostPct`(1.3=0/1.5=50,CHANGELOG_150.TXT 1.50.7/1.50.10)。`internal/shell/orbital_bombardment.go` `retaliationAttackers` 改簽名讀 defender 科技(`bestUnlockedWeaponValue`,新 helper)+ profile,取代舊 shipStrength 4/8/16 固定 tier,推導出「隨科技變強」+「隨版本 arc-cost 不同而不同」的反擊戰力。校準除數 `SatelliteStrengthScale=20` 使雷射參考點下星基/戰鬥站重現舊 tier 4/8,星辰要塞算出 20(非近似 19,誠實標見常數註解)。平衡 sanity:開局艦隊轟炸開局 AI 母星(僅星基),Profile13/15 各掃 Turn 0..14,最大損艦數皆為 1(不破壞平衡)。測試:`internal/gamedata/satellite_test.go`(fit/arc 公式錨點)+ `internal/shell/satellite_defense_test.go`(版本差異/科技效果/飛彈基地不吃 arc/地面砲台/平衡 sanity)。誠實限制:AI 現行資料模型無研究進度推進機制,`bestUnlockedWeaponValue` 在 `NewDemoSession` 自然對局裡恆落到 fallback 分支(雷射/核飛彈),「科技變強」效果目前只能在單元測試手動建構已解鎖科技的 `PlayerState` 觀察到。
  - [x] **#4 運輸艦淨現金版本差異(2026-07-11 補實作)**:新增「運輸艦隊」(Freighter Fleet)殖民地建造選項(`gamedata.FreighterFleetActionName`,前置科技 `TOPIC_NUCLEAR_FISSION`,估計建造成本 PP60——沿用既有 Special 一次性行動框架,見 `gamedata/special_actions.go`)。完工時 `shell.GameSession.applySpecialAction`:`s.Player.ActiveFreighters += gamedata.FreighterFleetShipsPerBuild`(手冊 p.168:每次建造 +5 艘)+ `s.Player.BC += s.RuleProfile.FreightersCashBonus`(新 `RuleProfile` 欄位,1.3=5/1.5=0,出處 MANUAL_150.html「Buildings & Freighters Free Cash Bug」+ CHANGELOG_150.TXT 1.50.8)。維護費(每艘 0.5 BC/回合)不用另外接——`engine.PlayerState.ActiveFreighters` 先前已接進 `RunEmpireTurn`(恆 0 no-op),本輪讓它真的變非 0,維護費隨之自動生效。**批次 B 的 #10 也已確認非差異**(見 `version-1.3-1.5-diff.md` #10),批次 B 至此結案。**簡化(誠實標)**:只模擬手冊「固定回饋」那一側,不模擬 0-3 BC 建造當下維護費立即扣款那一側;不做完整貨運/補給物流(殖民地間運食物/殖民者)——運輸艦本輪只有「可建造+維護費+版本現金加成」三件事;**AI 未接同一建造流程**,`ActiveFreighters` 對 AI 恆為 0。測試:`TestSpecialActionByNameZH`/`TestAvailableSpecialActions`(gamedata,新增運輸艦隊斷言)、`TestProfile13Values`/`TestProfile15Values`(gamedata,新增 `FreightersCashBonus` 斷言)、`TestFreighterFleetBuild*`(shell,新增:完工增加 ActiveFreighters+國庫、1.3 vs 1.5 現金加成差異、維護費隨後續回合生效、開局不建造回歸不變)。詳見 `docs/tech/version-1.3-1.5-diff.md` #4、`docs/tech/moo2-formulas-reference.md`「運輸艦淨現金版本差異」節。
  - [x] **#13 掃描/偵測距離:輕量戰爭迷霧(2026-07-11)**:新增 `internal/gamedata/detection.go`(`ScannerRangeParsec` 基礎2/Space4/Neutron6/Tachyon8、`OrbitalScannerBonusParsec` 星基+2/戰鬥站+4/星辰要塞+6 擇一取代不疊加、`ParsecToNormalized`=1/10 換算常數、`DetectionRangeNormalized` 加總換算——**全部近似**,手冊無公開 parsec 數字)+ `RuleProfile.SensorRangeVersionBonusParsec`(1.3=0/1.5=1,對應 MANUAL_150.html「Scanners and Communications Discrepancy」修正的整體近似,非逐科技數字)+ `internal/shell/detection.go`(`GameSession.VisibleStars`/`starVisible`,啟用先前無人讀取的 `Star.Explored` 死旗標;可見條件:已探索 ∪ 玩家自己的星 ∪ 落在玩家殖民地/艦隊偵測範圍內)。`cmd/moo2/interactive.go` `drawStarmap` 接上 fog 繪製(未偵測星降噪成暗灰小點、不畫星名/擁有環;可見星維持全繪)。調參依據:量測 `NewDemoSession()` 實際程序化星系(24星,種子42)鄰近星距離,使開局 Profile13 可見 3 顆星、Profile15 可見 7 顆星(母星區可見一小圈、遠星入霧,版本差異可觀察)。**誠實邊界**:fog 純視覺,不 gate 選星/派艦/殖民/轟炸等任何操作;不做敵艦 map blip(AI 艦隊為抽象戰力,無地圖座標,零地基)。測試:`internal/gamedata/detection.go` 無獨立測試檔(純查表函式,經 `ruleprofile_test.go` 的 `SensorRangeVersionBonusParsec` 斷言覆蓋)+ `internal/shell/detection_test.go`(6 個測試:母星可見+範圍外不可見、已探索恆可見、版本差異合成盤面+真實星系、軌道基地加成星辰要塞>星基、艦隊偵測源、`VisibleStars`/`starVisible` 接線+越界安全)。`go build`/`go vet`/`go test` 全過;`moo2sim -turns 20` 經濟軌跡不變(fog 不碰回合邏輯)。
- [ ] 資產分版(1.31 vs 1.5 LBX/資料)一起換——目前只分規則值,資產未分版

## Phase 8 — 文件 / 考究 / 文化 / 研究
- [x] 遊戲歷史與當年評價考究(`docs/history/moo2-history-and-reception.md`,角色:歷史考究專家,14 來源)
- [x] GitHub 致謝(README:openorion2/1oom/mom/字型/社群/Simtex)
- [x] 技術知識庫:LBX 資產格式 / 存檔格式 / 枚舉 / 公式 / ebiten 移植筆記(`docs/tech/`)
- [x] 華人圈中文討論資訊考究章節(`docs/history/moo2-chinese-community.md`,歷史考究專家,31 來源+誠實揭露侷限)
- [x] 華人圈文化現象(`docs/culture/moo2-chinese-cultural-phenomenon.md`,文案作家,事實有本、無 AI 味)
- [x] sprite/tile 畫質優化可行性 markdown(`docs/tech/sprite-tile-quality.md`)
- [x] UI 界面調整可行性 markdown(`docs/tech/ui-adjustment.md`)
- [ ] 技術知識庫:音樂整合 / 鍵盤滑鼠整合 / patch 處理 / 選單擴展(後續各 Phase 完成時補)
- [x] 三平台打包 CI(`docs/tech/packaging.md`):macOS(`.github/workflows/build-macos.yml`,`macos-14` runner 原生編 arm64+amd64 → `lipo` universal → `.app`/`.dmg`/`.tar.gz`)+ Linux/Windows(`.github/workflows/build-desktop.yml`);YAML 經 actionlint + yaml.safe_load 驗證,尚未在真 Mac 上實跑驗證(無 Mac 測試機)
- [x] 本機 docker 打包腳本(`docs/tech/packaging.md` §5):`scripts/package-appimage.sh`(Linux AppImage,linuxdeploy+appimagetool)、`scripts/package-windows.sh`(Windows zip)已實際跑過,`dist/MasterOfOrion2-cht-x86_64.AppImage`、`dist/MasterOfOrion2-cht-windows-amd64.zip` 皆產出並驗證內容(解壓/objdump 確認)。**推翻先前假設**:ebiten v2.9.9 Windows backend 已改純 Go(purego,無 cgo),`CGO_ENABLED=0` 即可跨編,不需 mingw-w64(`build-desktop.yml` 仍裝了 mingw,屬保守多餘,非錯誤,可留後續簡化)
- [ ] `cmd/moo2` 加可覆寫 assets/i18n 路徑(或 go:embed)取代相對路徑假設,讓 macOS `.app` 不需 launcher script 繞路(見 packaging.md §4 待辦)

## Phase 9 — 多人對戰(hotseat / 網路 lockstep→TCP)
> 考據定案見 `docs/tech/multiplayer-architecture.md`(原版通訊 OCR 自 CD 手冊 + 架構佐證自 patch 1.5 手冊)。
> 方向(使用者定案 2026-07-11):**保留原版決定性 lockstep 架構,傳輸換成 TCP**;起步先做熱座。
- [x] 原版多人通訊考據(手冊):序列/數據機/IPX 區網(2–8人)+ TEN 網際網路服務;DirectX 6.1→DirectPlay;決定性 lockstep + host 廣播 config + 同時回合(`docs/tech/multiplayer-architecture.md`)
- [x] **熱座(hotseat)**(2026-08-07):多位真人同機輪流下令。席位交換模型
  `internal/shell/hotseat.go`(原版是 `player[i]` 陣列 + 當前索引,remake 是單數欄位 →
  換人時整組搬進 `seat`);交接畫面 `cmd/moo2/hotseat.go`(座標取自 `Draw_Hotseat_Screen_`
  @ 0x626D6);「結束回合」改成全員下完令才推進世界;席位進存檔(原版也存遊戲模式)。
  ⚠ 非當前席位的帝國在 `EndTurn` 最後才結算(差一個 AI 回合的資訊),勝負判定只對當前
  席位跑——兩點都寫在 `advanceIdleSeats` 檔頭與 `docs/re/01-gap-report.md` 第 20 項。
- [x] **主選單「多人對戰」接上實際流程**(2026-08-07):`cmd/moo2/multiplayer.go`,
  整張版面取自反組譯(`Multi_Player_Screen_` @ 0xF4D99 / `sub_F42CA` / `sub_F009A`),
  含原版自己會隱藏的按鈕(熱座模式下沒有 JOIN GAME)。NETWORK / MODEM / NULL MODEM
  畫成灰的並明示未實作,不假裝可選。
- [ ] **引擎決定性化**:統一 RNG 種子(不用全域 `math/rand`/wall-clock)+ 消除影響模擬的 `range map` 不定序;加「兩機同指令序列→狀態雜湊比對」desync 偵測回歸測試
- [ ] **區網/線上 lockstep over TCP**:host/client、config 廣播、逐回合指令收齊→同步→結算、斷線重連、狀態雜湊校驗(中大型獨立子專案,排音樂/新遊戲流程/像素對齊之後)
- [ ] 熱座的**逐帝國真人/AI 標記**:原版是在玩家設定階段把個別帝國標成真人
  (`Get_Multi_Player_N_Humans_` 就是去數控制碼 100 的帝國),remake 目前只選「幾位真人」,
  由後往前接管 AI 對手。要對等得先有逐帝國的設定畫面。
- [ ] 熱座席位補齊玩家側系統:接管過來的 `AIOpponent` 沒有建造佇列 / 領袖 / 間諜 / 前哨站,
  第 2 席之後起步時這些是空的(見 `seatFromAI` 註解)。

## ★ 2026-08-07 盤點:gap report 的「最大系統級缺口」四條全部已完成

逐條 grep 之後發現 `docs/re/01-gap-report.md` Part B 的四大缺口清單**整份過期**——
歷史記錄系統(`shell/history.go`)、前哨站(`shell/outpost.go`)、艙損/維修
(`shell/repair.go`)全都已經建好,事件系統早就標記完成。Part A-2 的 Smacker 過場
(`cmd/moo2/cutscene.go` + `internal/smk`)同樣是過期的。

**為什麼重要**:這四條被後續每一輪的摘要當成現況反覆引用,於是「還缺什麼」的判斷
整個偏掉。文件裡的斷言一旦成形就會被當事實傳遞,而程式碼會往前走、文件不會。
細節與訂正後的清單見 gap report 第 39 項。

核實過後真正還缺的:網路多人(整塊)、`Command_Points` 專屬畫面、星圖 4 層
(星雲/遷移連線/星門/外交燈號,卡資料模型)、2 棟建築(真值已抽出,缺效果來源)、
殖民地地表的道路與擺放微調。

> 這幾項之後陸續了結:`Command_Points` 畫面(第 40 項)、地表道路(第 41 項)。
> 地表**擺放微調**與**植被層**仍缺,見第 41 項。

## ★ 2026-08-07 指揮點數視窗(gap report 第 40 項)

第 39 項核實後的清單裡最小的一項,做掉了。原版 `Show_Command_Points_Screen_` @ 0x8BAB9
整支只有 30 行:迷你星圖當背景 + 一塊文字視窗、ESC/點擊關閉。欄位組成由執行檔符號表
給出(`_starting_command_points_msg` / `_total_command_points_msg` /
`_total_command_point(s)_used_msg` / `_command_summary_msg`)——**結構是原版真值,
中文用字與視窗座標是 remake 自己的**。入口接在星圖右欄第 2 格(先前只顯示數字、點不開)。

⚠ **順帶抓到一個快取陳舊值**:`Player.CommandPointsSupply` / `UsedCommandPoints` 只在
`EndTurn` 更新,開局時是舊的。視窗第一版畫出「起始 5 + 軌道 0 = 總計 1」自打嘴巴;
星圖右欄那個淨值吃同一組欄位、**同樣是舊的**,只是單獨一個數字沒得對照所以一直沒被發現。
改用 `CommandPointsSupplyNow()` / `CommandPointsUsedNow()` 現算,兩處都修。

**把有關聯的數字放在同一個畫面上,本身就是一種驗證。**

## ★ 2026-08-07 殖民地地表道路(gap report 第 41 項)

第 31 項留下的「道路沒畫」補完。建築佔 6×6 的**格子**,道路走 7×7 的**格點**;四個方向
(兩條邊 + 兩條對角線)的合法段數是 42+42+36+36 = **156**,與 `COLROADS.LBX` 的資產數
一模一樣——這個等式就是幾何解對了的確認,不必去量圖。產生規則接在建築擺放的同一條亂數流上,
每個有建築的格子抽三次 `Random(2)`,在自己的四條邊上畫框。

**dir 2 / dir 3 的 72 張對角線圖,出貨版從來不會出現。** 全執行檔對那兩個旗標只有寫 0、
沒有一處寫 1,連帶讓產生器裡「空格子」那一整條分支變成死碼。remake 只實作有建築那條——
不是簡化,是認出死碼之後不抄。

⚠ **原版資料裡有兩個位元組級的錯,照抄不修**:繪製順序表少了格點 (5,4)、(3,4) 重複兩次;
包圍判定表的 Δa/Δb 對調了兩筆。修掉會讓 remake 比原版「正確」,而驗收標準是與原版一致。

**方法上的教訓:表不要用手抄。** 第一次照著 IDA 的 `dd` 清單手抄,解出 48/49 又有一處順序
不對,分不清是原版有錯還是自己抄錯。改成直接從 `Orion2.exe` 讀位元組才定案——先用「不重不漏
+ 遞減」當指紋掃全檔(零命中,排除自己解錯位址),再用有把握的前 14 個位元組當錨點掃到唯一位置,
最後連 Go 的表字面量也用腳本產生。

**同輪補上的缺口**:道路之後原版還跑一層**植被**(`COLVEGGI.LBX`),見下方第 43 項。

## ★ 2026-08-07 房屋抖動 + 母星的國會大廈(gap report 第 42 項)

把第 41 項留下的「抖動沒補所以道路對不上」補完,順帶抓到一件更基本的事:
**remake 的地表格陣從第一步就少放了一棟建築。**

抖動是排序之後的 8 輪隨機微調。⚠ 原版這段有兩個 bug,照抄不修:第二個座標算完之後被
**無條件覆蓋**(夾到 0 的那半段沒被編出來),於是換位對象**永遠落在主對角線上**;
內圈的偏移變數因此完全沒有作用(迴圈仍要留著,它決定抽幾次亂數)。
這兩點已用 `objdump` 對原始位元組獨立驗過,不是讀錯。

`Get_Bldg_CR_` 還有一個容易漏的語意:**「找一棟建築」會消耗亂數**——命中前每碰到一個空格
就抽一次(它和「找空位」共用同一支函式)。漏掉之後道路整串偏掉,而畫面看起來照樣正常。

**編號 9 = Capitol。** 先前判成「不可建造 → 正確地不在表裡」,對建造選單是對的,
對地表是錯的:它是實體建築,佔格、有美術、會被畫出來。**「不在建造表裡」與「不在地表上」
是兩件事。** 只有母星有(patch 1.5 手冊三處佐證)。

⚠ **仍不能宣稱與原版逐格相同**:流程結構接完整了,但**建築集合**仍有差
(Colony Base、已完成的一次性改造沒建模),集合差一棟則落點全偏。尚未對原版實測。

## ★ 2026-08-07 殖民地地表植被層(gap report 第 43 項)

原版每個殖民地畫面的空地上都長著草木,remake 整層沒有,所以畫面比原版空曠。

植物圖分群組、**每組固定 8 張**,組內編號越大株越大。資產 = `群組×8 + max(Random(8)−1−(a+b)/2, 0)`
—— 後面那項就是**透視**,越遠的格點越容易被壓到最小那株。群組由氣候決定(10 路跳表),
最大群組 12 → **13 組 × 8 = 104**,正好是 `COLVEGGI.LBX` 的資產數(這次直接跑 `lbxinfo` 驗)。
再交叉一項:群組 0 的前四張是 6×15、8×15、9×22、9×22,組內越大越大株,與透視項方向一致。

密度規則反直覺:**0 條路必長草;k 條路(k>0)機率 (k+1)/7**。而且最後那個
`Random(建築數+2)` 的結果永遠不是 0,判斷等於恆真——像是想寫「建築越多越不長草」卻沒生效,
照抄(那次抽樣要消耗亂數)。

繪製是**每格先植被再建築**的交錯,不是獨立一層——差別在遮擋。

⚠ 沒模擬:原版在「有格子被選取」時**一株都不畫**(remake 沒有這個狀態);
每株顏色沒有對原版逐張比對過。

⚠ 順帶踩到一個效能坑:尺寸在**產生**階段就要用(它進位置公式),而地表每幀重算、
`decodeAsset` 又沒快取——不處理就是每幀重解最多 72 張 LBX。

## 工作方式(使用者定案)
- go/ebiten 參考路徑 = `~/master-of-maigc/repo`(魔法大帝繁中化,patch 疊 kazzmir/master-of-magic 引擎)
- **不用多代理 workflow**;翻譯一組一組慢慢做(單代理逐項,使用者可隨時審閱)
- 每輪更新 GitHub(遠端 `main`,已設 upstream)
