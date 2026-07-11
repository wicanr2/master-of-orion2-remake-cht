# 誠實現況評估:還原度約 20%

> 日期:2026-07-04。依據:使用者實際遊玩後的驗收回饋(非測試結果)。
> 這份文件刻意直白,用來校正先前過度樂觀的進度敘述。

## 一句話結論

**2026-07-04 使用者實測時:「能載入原版美術、逐畫面中文化的畫面導覽器 + 幾個自製的玩具系統」,還原度約 20%。** ⚠ **此頭條為 2026-07-04 快照,2026-07-11 已大幅改變(見下方更新),但新還原度%不由我自評,待使用者重新實測定案。**

> **(2026-07-11 重大更新——頭條「不是能玩的遊戲」已過時,依 rule 63 更正)**:經本日 19 輪「接死碼 + 修 regression + 拓殖/多 AI」後,**遊戲現在是可玩的多帝國 4X 迴圈**(以下皆有 headless 模擬探針驗證,非自評):①經濟忠實+平衡(重力/士氣/礦產/建築效果/地形改造/污染/成長/饑荒/維護/收入;開局曾因指揮超支死亡螺旋,已用 SAVE10.GAM 真存檔 oracle 反推補基礎供給=5 修好,開局 BC 健康爬升)②玩家可拓殖(殖民船建新殖民地)③3 個 AI 對手各按性格獨立擴張成長(好戰型戰力遠高於科學型)④可贏一局(征服 + 真銀河議會選舉,玩家+3AI=4 帝國達手冊 ≥3 門檻)⑤戰鬥(beam/飛彈/球狀分流 + 地面戰陸戰隊/戰車/軌道轟炸)⑥指揮點數/領袖技能/間諜偷科技/勝利條件。**這是可驗證的「什麼現在能運作」,不是自評 fidelity%——vs 原版的還原度仍待使用者實測。** 仍缺(增強/打磨非可玩性阻塞):戰機/航母、武器改造 mod、安塔蘭母星+次元傳送門(第三勝利路徑)、完整 spy/leader/diplomacy/多AI目標選擇 UI、逐畫面像素對齊、地面戰 d100 結構與飛彈速度的實機校驗。詳見 `docs/tech/remaining-work-roadmap.md`。(註:MOO2 殖民地是人口職務模型非格子地形,先前「殖民地格子」寫法已於 §B 更正。)

## 為什麼「單元測試全綠」卻只有 20%

先前多輪我用「新增系統 + 單元測試通過 + headless 截圖有畫面」當作完成訊號,逐輪回報「研究/艦隊/人口/建築/事件/安塔蘭/種族/存讀檔/AI 都做完了」。這是**用內部訊號謊報完成**(正是專案規則 `rulebook/65` 警告的反例):

- 測試綠只證明「我寫的那套自製邏輯自洽」,不證明「它跟原版一樣」。
- 那些係數大多是我**自己編的 remake 近似值**(程式碼裡就標了「remake 調校值」),不是 MOO2 手冊的真實數字。
- 真正的驗收標準是「拿原版當 oracle 實測對齊」,而使用者一玩就發現差異巨大。

**唯一有效的還原度量尺 = 對原版實測比對,不是我的測試套件。** 這份評估以使用者實測為準。

## 真正已完成、扎實的部分(這 ~20% 的內容)

這些是有真實逆向/驗證基礎、值得保留的資產層成果:

1. **LBX 資產解碼**:容器解析、scan-line RLE 影像、多幀 delta 動畫累積(修好 DIPLOMAT)、調色盤鏈(無內嵌調色盤畫面的上色機制)。對照 openorion2 逐位元組驗證過。
2. **存檔格式唯讀解析**:SAVE10.GAM 全區段解出(殖民地/行星/星/領袖/玩家/艦艇),有回歸護欄。
3. **16 個原版畫面**:載入真原版 LBX 美術,以「擦底疊字」疊上中文。畫面看起來像原版。
4. **中文譯表**:科技/元件/選單/外交等數百條 UI 字串翻譯。
5. **可跑的技術骨架**:go/ebiten 視窗、docker 編譯/截圖流程、AppImage 打包、視窗縮放。

## 真正缺、或根本不忠實的部分(這 ~80%)

### A. 使用者本次直接點出的兩點
1. ~~**完全沒有音樂/音效**~~ **(2026-07-10 更新:基礎已打通)**。第一性原理翻案:MOO2 **無 XMI/MIDI**,音樂/音效是 LBX 內 PCM WAV,已整合 ebiten 音訊 + 主選單 BGM + 按鈕音效(`internal/audio`、`cmd/moo2/audiohook.go`,測試綠)。**仍待**:曲目↔場景對應對原版聆聽定案、星系/戰鬥各場景 BGM、戰鬥音效庫(CMBTSFX/SPHERSFX)。見 `docs/tech/audio-format.md`、`audio-track-map.md`。
2. **一進新遊戲就跟原版差異太大 + 按鍵沒完整對齊**:
   - 原版新遊戲流程:主選單 →(選好設定)→ **獨立的種族選擇畫面(13 族肖像 + 自訂種族點數)** → 真正的星系生成與母星配置 → 進到有真實殖民地資料的遊戲畫面。
   - 我的版本:把種族擠進設定畫面一格、ACCEPT 後跳到**程序生成的假星系 + 示範殖民地**。體驗與原版完全不同。
   - 各畫面的點擊熱區多為**估計座標**,不少還是「整畫面當返回鍵」;擦底疊字蓋掉了英文,但位置未逐畫面像素級對齊。
   - **(2026-07-10 更新)獨立種族選擇畫面已成**(13 族原版肖像 + 自訂點數 + 命名/旗色,見 `docs/tech/newgame-flow.md`)。**外交畫面已破解 DIPLOMAT.LBX 並重建**(逐族使節房+使節疊合,13 族對應核實,見 `diplomat-lbx-layout.md`)。**戰術戰鬥畫面已換原版美術**(STARBG 星空+COMBAT 控制列+可見 CMBTSHP 艦艇,見 `tactical-combat-assets.md`)。已補 `-gamegallery` headless 導覽,實測 8 畫面端到端在互動 app 內全繁中渲染。**仍待**:控制列烘進的英文按鈕中文化、艦型 sprite 完整對照、真星系/母星、逐畫面像素對齊。

### B. 更根本的:遊戲玩法是「自製」的,不是 MOO2
我加的那些回合系統是**很薄的 remake 草稿**,與原版規則差距極大:
- 殖民地:**更正先前假前提**——經 openorion2 `gamestate.h` + 存檔格式核實,MOO2 殖民地**本來就是人口職務模型**(每殖民者一個 `Job`:農/工/科,產出 = `food_per_farmer`/`industry_per_worker`/`research_per_scientist` × 行星屬性),**沒有逐格地形分配**。先前本文寫「格子地形 / 每格產出 / 尚無格子地形」是把 MOO2 誤當 Master of Magic 的錯誤描述,已刪——remake 的三職務數字模型正是 MOO2 的忠實結構。真正的表現層小缺口是「個別殖民者圖示 + 殖民者血統類型(android/alien/native)」,不是地形格。原版是 **30+ 種建築 + 污染/食物/貿易真公式**。**(2026-07-11 更新:40 棟建築已入建造清單/前置科技 gating(`colony-buildings.md`),其中 18 棟已依手冊頁碼忠實建模數值效果——per-worker 產出(自動工廠/機器人採礦廠/深層核心礦場/行星超級電腦/銀河網路中心/研究實驗室/氣候控制器/太空大學)與「殖民地整體固定加成」(自動工廠/機器人採礦廠/深層核心礦場/研究實驗室/行星超級電腦/銀河網路中心/水耕農場/地底農場)已分開累加,不再像先前那樣把手冊的固定值揉進 per-worker 湊數(小殖民地過度受益、大殖民地受益不足);另接了收入百分比(太空港+50%/證券交易所+100%,逐殖民地精確套用)、人口上限加成(生態圈+2)、固定成長點(複製中心)、機器人工廠礦產豐度分級固定加成。仍缺:軍事/防禦類(戰機/軌道防禦等 ~13 棟)仍只記錄已建、不影響數值——這些需要對應系統(艦隊駐防/軌道防禦)先建好。詳見 `docs/tech/colony-buildings.md` 建模狀態一節。**(2026-07-11 追加:重力懲罰系統已接進生產管線——`ColonyState` 新增 `PlanetGravity` 欄位,`RunColonyTurn` 對食物/工業/研究套用 Low-G -25%/Heavy-G -50% 懲罰,行星重力產生器`NormalizeGravity`旗標也已生效(能真正消除懲罰,不再是 no-op)。種族 Low-G/High-G 重力天賦仍未建模,固定以一般種族為基準;現實中唯一會產生非 Normal-G 殖民地的路徑是存檔載入(`ColonyStateFromSave` 讀 `save.Planet.Gravity`,兩者同源 openorion2 enum ordinal,直接轉型)——`NewDemoSession` 的玩家母星與 AI 母星目前固定 Normal-G(且是本專案唯一的殖民地建構點,地面入侵佔領敵方殖民地也只是複製既有 Normal-G 母星,尚無「開拓新殖民地」流程會派生 Low-G/Heavy-G 殖民地),故這次接線在 demo session 裡暫時看不出懲罰效果,真正生效要等存檔載入模式或殖民擴張系統落地。)****(2026-07-11 再追加:士氣系統已接線——`GameSession` 新增 `Government` 欄位,`internal/shell/session.go` 的 `colonyMoralePercent` 依政府基礎值(`gamedata.MoraleGovernmentBase`)+ 已建士氣建築(全息模擬艙 +20%、歡樂穹頂 +30%)算出 `ColonyState.MoralePercent`,取代先前無手冊依據的硬編 +10。⚠**這讓新遊戲母星起始士氣從硬編 +10 變成忠實值 0**(獨裁政府「無 Barracks -20%」因母星已建 Marine Barracks 被解除、淨額歸零,且未建任何士氣類建築,無正面加成)——第一回合食物/工業/研究產出因此比先前 demo 少一成,是手冊算出來的忠實值、不是退步。仍缺:多種族懲罰(remake 不追蹤殖民地人口是否含未同化外族血統,異族管理中心暫無可見效果)、首都淪陷懲罰(remake 無「首都被攻陷」狀態)、Virtual Reality Network(手冊定性為「成就」非建築,remake 無成就追蹤系統)、進階政府(Imperium/Confederation/Federation/Galactic Unification 一律近似成對應基礎型,不區分)。詳見 `docs/tech/colony-buildings.md` §6.1 士氣列與 `internal/gamedata/morale.go`。)****(2026-07-11 再追加:機器人工廠的礦產豐度分級已接線——比照重力接線手法,`ColonyState` 新增 `MineralRichness` 欄位,獨立保留建立殖民地當下的原始礦產豐度分類(不再從已烘進 `IndustryPerWorker` 的靜態費率事後反推)。`applyBuildingEffect` 依 `gamedata.ProdRoboticFactoryBonus`(既有查表函式)查出手冊固定值(Ultra Poor+5/Poor+8/Abundant+10/Rich+15/Ultra Rich+20)加進 `FlatIndustry`,不動 `IndustryPerWorker`,避免與已烘進的礦產費率重複計算。存檔行星由 `ColonyStateFromSave` 讀 `save.Planet.Minerals`(與 `gamedata.PlanetMinerals` 同源 openorion2 enum ordinal,直接轉型),母星固定 Abundant。詳見 `docs/tech/colony-buildings.md` §6.1 機器人工廠列。)****(2026-07-11 再追加:指揮評等(Command Rating)供需系統已接線——手冊 p.169「Every ship you build...uses points from this rating as a maintenance cost. The number of points a ship requires is the same as that ship's size class」+「For each rating point required by a ship that is not covered, 10 BCs come out of your income every turn」,先前 `gamedata.IncomeCommandOverflowCost` 是零呼叫端的死碼。現接上供給(`gamedata.CommandPointsFromBuildings`:星基+1/戰鬥站+2/星辰要塞+3,三者取代不疊加)與需求(`gamedata.ShipCommandCost`:依艦體 size class,Frigate=1..Doom Star=6,支援艦視同 Frigate,貨運艦隊手冊明文排除不計),`shell.GameSession.EndTurn` 每回合重算兩者寫入 `PlayerState.CommandPointsSupply`/`UsedCommandPoints`,`engine.RunEmpireTurn` 算超支懲罰併入 `NetBC`(曝露於 `EmpireOutput.CommandOverflowCost`)。⚠**(2026-07-11 當時的已知問題,已於同日稍後修復,見下段)新遊戲開局母星只有 1 座星基(+1 供給)卻有 3 艘開局艦艇(殖民船+2 偵察艦,各 1 點=3 點需求),缺口 2 點,每回合被動扣 20 BC**——這段文字當時誤判成「手冊機制被誠實呈現的結果,不是 bug」,實際上是 **regression**:接指揮點數系統時漏算了帝國基礎供給,見下段修復。仍缺:通訊科技(Tachyon/Subspace/Hyperspace Communications,手冊有「每軌道衛星 +1/+3」定性數字但需要「逐軌道衛星計數」的資料串接,尚未做)、Operations 軍官技能(手冊只有定性敘述無精確數字)、政府型態加成(Imperium +50% 有手冊數字,但本專案政府型態全域固定 Dictatorship、無 Imperium 這個狀態可觸發)、AI 對手(用抽象 `FleetStrength` 而非逐艦清單,無法推導需求,供需維持零值預設無懲罰)。詳見 `docs/tech/moo2-formulas-reference.md`「指揮評等供需」節、`docs/tech/colony-buildings.md` §6.1 星基/戰鬥站/星辰要塞列。)**(2026-07-11 追加,同日 regression 修復:上段「開局缺口 2 點/每回合 -20 BC」誤判為忠實機制,實為漏算帝國基礎指揮評等供給的 bug——用真實存檔 `SAVE10.GAM`(`/home/anr2/moo2-private-build/gamedata/mastori2/SAVE10.GAM`)oracle 反推:5 個活躍玩家(不同種族)各持 1 殖民地,`CommandPoints` 讀到 6(其中 1 名玩家=8)。比對已建成軌道衛星:讀到 6 的玩家只建星基(+1),6-1=5;讀到 8 的玩家建了星辰要塞(+3),8-3=5——5 個不同種族玩家一致反推出基礎值 5,與種族/政府無關,訂為 `gamedata.CommandPointsBase=5`,`shell.GameSession.totalCommandPointsSupply` 在逐殖民地建築供給之外,每帝國(非逐殖民地)加這一次。修復後開局供給=5(基礎)+1(星基)=6≥3(需求),不再超支,300 回合被動不建造測試的 BC 下限由 -4000 收回 -400(實測最低點約 -51)。**已知限制(TODO)**:SAVE10.GAM 5 名玩家都只有 1 殖民地,無法從單一存檔分辨這 5 點是「每帝國一次性」還是「每殖民地各自 +5」(兩種假設在單殖民地情境下算出的總供給相同),暫採 per-empire flat,待多殖民地存檔驗證。詳見 `internal/gamedata/income.go` 的 `CommandPointsBase` 註解。)**(2026-07-11 再追加:地形改造(Terraforming)/蓋亞轉化(Gaia Transformation)/土壤改良(Soil Enrichment)三個「Special」一次性行動已從零呼叫端的死碼接進殖民地建造佇列——`internal/gamedata/terraform.go` 移植的氣候階梯/人口係數公式先前只有單元測試、沒有任何遊戲流程會呼叫。現在 `ColonyState` 新增 `Climate` 欄位(比照 `PlanetGravity`/`MineralRichness` 的零值陷阱處理手法),新增 `gamedata/special_actions.go` 把這三項排進建造選單(前置科技:地形改造 `TOPIC_GENETIC_MUTATIONS`、蓋亞轉化 `TOPIC_TRANS_GENETICS`、土壤改良 `TOPIC_ADVANCED_BIOLOGY`,依 `openorion2/src/tech.cpp` 的 `research_choices[]` 陣列索引=`ResearchTopic` 列舉值交叉驗證,與既有 34 項建築的前置科技逐一核對 100% 相符),完工時把氣候沿手冊階梯推進一級、FoodPerFarmer 依手冊絕對值差值疊加(保留既有建築加成不被覆蓋)、PopMax 依 `pop_climate` 百分比係數等比例縮放。⚠**兩處誠實近似,非官方精確值**:①PopMax 縮放是「目前整體 PopMax × 新舊係數比例」,因為 remake 沒有獨立的「行星尺寸→基礎人口容量」對映表,無法精確重算「基礎值 × 新係數」;②建造成本(PP)手冊完全沒給,比照其餘 34 項建築的既有估計慣例、依同一 RP 量級外推(地形改造 260/蓋亞轉化 900/土壤改良 150),手冊講的「地形改造每次套用成本遞增」未模擬(固定成本,標 TODO)。地形改造在 Barren 的下一級手冊給兩個候選(Desert/Tundra)未給選擇條件,remake 固定選第一個候選。詳見 `docs/tech/colony-buildings.md` §6.1 地形改造列與 `internal/gamedata/terraform.go` 檔頭。)**
- 科技樹:~~原版每主題要在數個科技間**抉擇**、有真實 RP 成本表;我的是線性自動推進。~~ **(2026-07-10 已完整忠實化:真 RP 成本 + 每主題數科技間抉擇 UI(真中文名)+ 抉擇決定艦艇元件解鎖,三步全成、有測試;見 `docs/tech/research-system-status.md`。此為首個完整對齊原版的 gameplay 系統。剩抽象元件資料模型重設計小尾巴。)**
- 戰鬥:原版有真實武器機制(命中/傷害/射程/防禦/飛彈躲避/球狀傷害/地面戰);我的原本是抽象戰力相減。**(2026-07-10 部分忠實化:太空戰命中/傷害/過盾/過甲已接 gamedata 真公式(`ResolveShot`,對 openorion2/手冊核實),戰場已換原版 STARBG+COMBAT 美術。)** **(2026-07-11 修正一個誤植斷言,rule 63:這裡先前寫「仍待飛彈躲避、球狀傷害」是錯的——飛彈防禦/AMR/彈頭 Beam Defense(`gamedata/missile.go`)與球狀傷害(`gamedata/damage.go` 的 `DamageSpherical*`)公式早就已經移植自手冊、有測試,只是先前的戰鬥解算(`fireRound`/`battleVolley`)全部武器都硬套 beam 邏輯,飛彈被當 beam 打。這輪已修正:新增 `internal/shell/weapon_kind.go` 依武器名分類 beam/missile/spherical(核對手冊「Notes on Spherical Damage」後確認「死光」不是球狀武器——那是一般光束武器,且是 `DamageForHit` 手冊 worked example 的出處;現行武器表 `WeaponOptions` 目前也沒有任何真正對應到球狀武器的元件,只有核飛彈/麥克萊特飛彈兩個飛彈武器),新增 `shell.ResolveMissileShot`(AMR 攔截 + Jam Chance 躲避判定)/`shell.ResolveSphericalShot`(已測試,暫無武器掛載),`fireRound`/`battleVolley` 依武器類型分流,beam 行為不回歸(測試護欄)。**真正仍待實機/DOSBox oracle 定案的仍是兩點**:①`missile.go` 檔頭的「飛彈速度」手冊公式與附表自相矛盾,待動態驗證;②**地面戰核心傷亡解算結構**——`ResolveGroundBattle` 的「每回合雙方擲 d100+force、低者敗、平手歸守方」沿用一代(1oom)`game_ground_kill` 借用結構(見 `ground_battle.go` 檔頭),force 值雖用 MOO2 手冊表,但這個解算結構本身尚未對 MOO2 實機核實。(⚠ 本輪 subagent 一度誤刪此項稱「解算式已定案」——那是把「新接的戰車生成/軌道轟炸換算公式是手冊錨定」錯當成「核心傷亡解算結構已驗證」;Opus 驗證時改回。)艦型 sprite 完整對照(task 12)仍待,與戰鬥公式無關。詳見 `docs/tech/tactical-combat-weapon-kinds.md`。)** **(2026-07-11 地面戰從「僅陸戰隊」補完為「陸戰隊+戰車+軌道轟炸」:`gamedata/ground.go` 原本零呼叫端的裝甲營房戰車生成公式(`GroundArmorBarracksUnits`/`Cap`)與軌道轟炸換算公式(`GroundBombHitsFromDamage`/`GroundPlanetTotalHits`)接進活對局。**戰車**:比照陸戰隊完全對稱(`advanceArmor`/`LoadTanks`,新增 `FleetTanks`/`PlayerColonyTanks`/`ArmorBarracksAge`),與 `LoadMarines` 共用同一個運力池(無獨立戰車運輸艙位資料,標簡化);`InvadeColony` 攻方 `GroundForce` 混編陸戰隊+戰車(合併順序「陸戰隊在前、戰車在後」,靠 `ResolveGroundBattle`「前排先陣亡」規則從單一存活總數精確拆回兩兵種各自存活數);已研究 Battleoids(`TOPIC_ASTRO_CONSTRUCTION`)則戰車固定 3 hits+額外 force 加成。100 場對照組測試證實加戰車確實提升攻方勝率(無戰車 0.34 → 12 輛戰車 1.00),證明不是擺著沒用的死碼。**軌道轟炸**:新增 `internal/shell/orbital_bombardment.go`,`BombardColony(starIdx)` 模擬 10 輪齊射(重用既有 `ResolveShot`/`ResolveMissileShot`)→ 換算 hits → 直接扣減殖民地人口。⚠**仍缺**:守方戰車(AI 無 `ColonyBuildings` 追蹤,無資料可推導,誠實留白)、轟炸扣建築/儲存生產/駐軍(同理,AI 無對應持久資料)、轟炸套用光束/魚雷減半與電腦命中加成(戰術戰鬥層本身還沒有這兩項的獨立函式,非本輪引入的缺口)、人口歸零後是否摧毀殖民地(手冊未講,不臆測)、轟炸的 UI 操作介面(僅引擎層函式,`interactive.go` 未接按鈕)。詳見 `docs/tech/ground-combat-algorithm.md`「2026-07-11 裝甲營房戰車 + 軌道轟炸接線」節。)**
- 艦艇設計:原版有艦體空間格、每元件佔空間、改造(mod);我的是四個下拉選單。**(2026-07-11 部分忠實化:艦體空間格 + 武器佔格已接手冊確認值(`gamedata/shipspace.go` + `session.go` `ShipDesignSpaceUsed`/`ShipDesignFits`,詳見 `docs/tech/ship-design-space.md`),超格設計可被驗證函式擋下;仍待:改造 mod 佔格、特殊系統精確佔格(手冊無數字,現為 5% 估計值)、Design Dock UI 本身。)**
- 拓殖(Colonization):**從「無開拓殖民地流程」變成「最小拓殖已接」(2026-07-11)**。先前玩家
  只有母星、完全無法擴張——`shell.GameSession.ColonizeStar` 現已接線:玩家艦隊帶殖民船抵達無主
  適居星,可建立新殖民地(起始人口 1、全農起步,`gamedata.PlanetBasePopMax` 依行星大小/氣候算
  PopMax,兩者皆為手冊直接引文/openorion2 公式核實,非臆造),`cmd/moo2/interactive.go` 加了對應
  「建立殖民地」按鈕。**(2026-07-11 追加:AI 對手擴張已從「只標旗標無經濟」接上「建真殖民地、
  經濟隨擴張成長」)**——先前 `aiExpand`(每 5 回合佔一顆無主星)只設 `Star.Owner=2` +
  `OwnedStars++`,從不建立 `engine.ColonyState`,導致 AI 殖民地數恆為開局母星 1 筆、
  `RunEmpireTurn` 算出的 `TotalNetIndustry` 永遠停在初始母星產出、AI `FleetStrength` 只能線性
  成長。現在抽出共用函式 `internal/shell/colonization.go` 的 `newColonyFromStar(starIdx, gov,
  foodBonus, indBonus, resBonus)`,玩家 `ColonizeStar` 與 AI `aiExpand` 共用同一套建殖民地邏輯
  (氣候/重力/礦產/大小解析、PopMax 查表、全農起始、士氣算法一致),`aiExpand` 佔星時同步
  append 進 `AIOpponent.Colonies`/`ColonyStars`。40 回合探針對照:修前 AI 殖民地數恆 1、
  FleetStrength 線性(3→60,約每回合 +1.5);修後 AI 殖民地數隨回合增至 9、FleetStrength 加速
  成長(3→101,同期間近 2 倍),玩家開局 BC 軌跡不受影響(兩版本一致:102→…→96)。**AI 政府
  型態未建模**(`AIOpponent` 沒有 `Government` 欄位),`aiExpand` 建殖民地時士氣一律用
  `gamedata.MoraleGovDictatorship` 保守預設(與母星 `playerHomeworldColony` 基準一致);AI 也
  沒有種族加成模型,`foodBonus`/`indBonus`/`resBonus` 一律傳 0,不臆造。**仍缺**:AI 選星策略
  (目前是「星圖索引順序第一顆無主星」,非距離/資源導向)、行星選擇子畫面(每星固定一顆行星,暫不
  需要)、氣態巨星/小行星帶類型(本 remake 星系生成從未產生這兩類,科技 gate 掛勾點已留但無案例可
  測)。詳見 `docs/tech/colonization.md`。**(2026-07-11 更新:多 AI 對手數量已由 1 個擴為 3 個**
  ——`NewDemoSession` 建 3 個性格互異的 AI 對手(席隆人/科學傾向、姆瑞森人/好戰、布拉西人/擴張),
  各自不同母星星、獨立造艦/擴張/外交態勢,經 ~40 回合探針驗證三者都會各自成長,不是只有第一個在
  動。AI 選星策略(星圖索引順序,非距離/資源導向)、AI 對 AI 互動(3 個 AI 只各自獨立對玩家,彼此
  不打仗不外交)仍是既有簡化,見 `docs/tech/victory-conditions.md` 第 6 節 TODO。)**
- 外交/間諜/議會、隨機事件、安塔蘭母星與歐瑞恩守護者、勝利條件——**大多缺席或極度簡化**。
  **(2026-07-11 更新:勝利條件從「完全沒有」變成「兩條路徑已接引擎層」。** 銀河議會選舉
  (手冊 GAME_MANUAL.pdf p.183:半數銀河殖民+≥3存續種族才成立、票數依人口、2/3超級多數當選、
  AI當選時玩家可accept/reject)與殲滅所有對手,兩者都已接進 `EndTurn`/`InvadeColony`,沿用
  `internal/engine/victory.go`(2026-07-03 就存在但從未被呼叫過的死碼)+ 新增
  `internal/gamedata/council.go`(人口→票數、成立門檻)、`internal/shell/council.go`(狀態機/
  存讀檔),取代先前議會畫面用的無門檻/無2/3多數簡化版 `CouncilVote`。**UI 仍未做**:議會畫面只印
  文字狀態(尚未成立/待開/已分出勝負/待回應),沒有 accept/reject 互動熱區、沒有勝利/落敗結束畫面。
  Antares 母星次元傳送門勝利(手冊第三條路徑)完全沒有對應流程(無 Dimensional Portal/艦隊遠征/
  母星戰鬥),列 TODO 不硬做。詳見 `docs/tech/victory-conditions.md`。)**
  **(2026-07-11 再更新:議會由「玩家 vs 單一 AI 二元計票」擴為「N 帝國真議會」。** `NewDemoSession`
  已建 3 個 AI 對手(見上「拓殖」段落),場上存續帝國上限變成玩家+3 AI=4,手冊字面門檻
  `gamedata.CouncilMinExtantRaces`(≥3 存續種族)現在真的可達——先前的 shell 層資料模型限制近似
  覆寫值 `councilMinExtantRacesOverride`(=2)**已移除**。`advanceCouncil` 也從「雙方各算一次票」
  generalize 為逐帝國(玩家 + 每個 AI 各自獨立)算票、2/3 門檻用全體總票數判定,`PendingCouncilElection.
  EnemyName` 正確指向實際當選的那個 AI(不是寫死 `AIPlayers[0]`)。仍未實作的簡化:「候選人限定
  票數最高兩位 + 第三方外交搖擺票」這條手冊規則(需要 AI 對 AI 的外交關係模型,目前
  `AIOpponent.Relation` 只有對玩家一個方向),見 `docs/tech/victory-conditions.md` 第 5-6 節。)**
  **(2026-07-11 再更新:間諜從「公式移植未接(零呼叫端死碼)」變成「最小可玩迴圈已接」。**
  `internal/gamedata/spy.go` 的 8 個機率函式(手冊 `Notes on Spying`,`SpySlotBonus`/
  `SpyEffectiveThreshold`/`SpyRollChance`/`SpyVsSpy*` 等)先前雖已移植但無任何呼叫端;新增
  `internal/shell/spy.go` 接上「訓練間諜(`TrainSpy`,花 BC)→ 每回合結算 STEAL(偷科技)→
  SpyVsSpy 判定」的最小迴圈,玩家 ↔ 單一 AI 對手雙向生效。**只涵蓋 STEAL 一項任務**:破壞
  (SABOTAGE)手冊只定性描述效果、無數值規則,標 TODO 不做;逐對手分配/任務選單(Espionage/
  Sabotage/Hide 分開指派)延後,最小迴圈預設所有間諜對唯一 AI 對手做 STEAL;防禦方 Agent 不
  獨立追蹤(固定 DB=0,對應手冊「零 Agent 防禦仍生效」的描述,非遺漏);種族/科技/政府對間諜
  的加成現行無資料可推導,一律 0(TODO)。詳見 `docs/tech/spy-system.md`。)**

- 領袖/軍官:原本純裝飾——`demoLeaders()`(如「馮·諾伊曼 科學家 Level 5」)的技能字串不 bonus
  任何東西,`gamedata/officer.go`(`LeaderExpLevel`/`LeaderSkillBonus`)、`formulas.go`
  (`LeaderHireCost`)是移植好卻零呼叫端的死碼。**(2026-07-11 更新:從裝飾變成技能真正生效,但
  只有 2 個技能實際接進遊戲。**openorion2 全專案 grep 後發現:MOO2 25 個領袖技能裡,openorion2
  自己也只給 4 個(`SKILL_WEAPONRY`/`SKILL_HELMSMAN`/`SKILL_FAMOUS`/`SKILL_MEGAWEALTH`)寫了真
  正的效果消費端,其餘技能(科學/財政/工業/外交/間諜等)在原專案裡也只是「畫面顯示 + 可算數字
  但沒人用」——這不是 remake 特有的缺口,是原專案本身的完成度天花板。本輪接了「科學家」→固定
  研究點數(`ColonyState.FlatResearch`)、「貿易家」→收入百分比(`ColonyState.IncomeBonusPercent`)
  兩個殖民地領袖技能,母星研究產出因此從 30 → 55(`TestGameSessionEndTurn` 已同步更新)。**艦艇
  軍官(指揮官/工程師)目前仍是裝飾**:引擎層已備妥 `ShipBeamAttackWithOfficer`/
  `ShipBeamDefenseWithOfficer`(Weaponry/Helmsman 命中閃避加成)與 `HireLeader`(雇用金流),但
  remake 的 `shell.Ship` 沒有軍官指派欄位、也沒有戰鬥解算迴圈會呼叫這兩個新函式,demoLeaders
  的「指揮官」技能標籤本身也無法唯一對應到任何 openorion2 技能(待使用者定案)、「工程師」對應
  的艦艇維修率加成 remake 無承接系統。詳見 `docs/tech/leader-officer-skills.md`。)**

- 收入(Income)三個零呼叫端死碼(**2026-07-11 接線**,解鎖自本輪稍早修好的開局經濟平衡):
  ①**政府 money 加成**(Democracy +50%/Federation +75%,`gamedata.IncomeApplyGovernmentMoneyBonus`)
  已接進 `engine.RunEmpireTurn`(新增 `IncomeGovtMoneyBonusPercent` 查表 + `PlayerState.GovtBonusMoneyPercent`
  欄位,`shell.GameSession.EndTurn` 依 `s.Government` 算好傳入)。demo 預設 Dictatorship→0,no-op;
  AI 對手無 `Government` 欄位建模,不受影響。②**運輸艦維護費**(每艘使用中 -0.5 BC,
  `gamedata.IncomeFreighterMaintenanceCost`)已接進 `RunEmpireTurn`(新增 `PlayerState.ActiveFreighters`
  欄位),但本專案的艦種塑模(`COMBAT_SHIP`/`COLONY_SHIP`/`TRANSPORT_SHIP`/`OUTPOST_SHIP`)沒有
  「Freighter」這個艦種,呼叫端恆傳 0,目前是 no-op,接線先備妥。③**士氣對收入的調整**
  (`gamedata.IncomeMoraleAdjustedProduction`,手冊 p.170)**判定為不接**——查證後發現
  `RunColonyTurn` 早就把士氣套進食物/工業/研究產出,`RunEmpireTurn` 的稅收/餘糧收入/貿易品收入
  全部是從這個已調整過的產出直接換算,若再對收入套一次士氣就是雙重計算。刻意不接,理由完整記錄
  在 `engine/empire.go` 註解與 `docs/tech/moo2-formulas-reference.md`「士氣對收入的影響」節。
  這三項在 demo 對局皆是 no-op(政府=Dictatorship、無貨運艦種、母星 morale=0),已用 20 回合 BC
  軌跡探針確認接線前後一致(101→130 健康爬升,無退化),同重力/礦產這類「demo 看不出效果但對其他
  狀態忠實」的完成型態。

一句話(**2026-07-11 更新**):以上多為 2026-07-04~10 的描述;經 2026-07-11 的 19 輪接死碼+修 regression+拓殖+多 AI,**遊戲已是可玩的多帝國 4X 迴圈**(玩家與 3 AI 各自擴張、忠實+平衡經濟、可經征服/議會贏一局,皆有模擬探針驗證)。先前「無法真的玩一局、只能點畫面」的描述已過時(見頭條 2026-07-11 更新)。**仍待**:戰機/武器 mod/安塔蘭等增強、完整 UI 打磨、以及 vs 原版還原度的使用者實測定案——這些是「讓它更完整/更像原版」,不是「能不能玩」的阻塞。

## 要達到「高還原度」真正該做的(誠實 worklist)

依對玩家體驗的影響排序,不是依好做的程度:

1. **音樂/音效**(Phase 6):逆向 XMI + 音效格式,ebiten 音訊整合,SoundFont 音色(沿用 moo1 經驗)。
2. **忠實的新遊戲流程**:獨立種族選擇畫面(真肖像 + 自訂點數)、真實星系/母星生成、進遊戲後載入**真實殖民地初始狀態**。
3. **按鍵/熱區逐畫面像素級對齊**:用 IMGLOG/截圖逐屏校對每個原版按鈕的真實座標,取代估計與整畫面返回。
4. **忠實 gameplay 規則**(這才是主體工作量,對應 PLAN 的「從零重建引擎」軌):殖民地建築全表數值/科技抉擇樹/真實戰鬥機制/艦艇空間設計/污染食物貿易公式,逐系統以手冊為權威實作並**對原版實測比對**。
5. **移除或明確標注自製近似**:凡是我編的係數(武器傷害、建築效果、成長門檻、種族加成、事件、AI)都要換成手冊/逆向的真值,或在 UI/文件標明是 remake 近似。

## 給後續工作的鐵律(記取本次教訓)

- **驗收看原版實測,不看我的測試綠。** 測試只防自己的回歸,不代表還原度。
- **不再把「新增一個自製系統」當還原進度。** 還原 = 對齊原版,不是長出新東西。
- **誠實標注每個數字的來源**;自編的就說是自編的。
- 還原度用「對原版實測」估,目前:**約 20%**。
