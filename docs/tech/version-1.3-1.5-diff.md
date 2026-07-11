# patch 1.3(1.31)→ patch 1.5(1.50)規則/數值差異 + 版本 profile 設計建議

> 2026-07-11。目的:為 CLAUDE.md 核心需求「主選單選擇 1.3 或 1.5」打底——先確認**真正需要分版的值有多少**,
> 再提版本 profile 資料結構。純研究,不改任何程式/資料檔。
>
> 方法(rulebook 62/65):以 `moo2_patch1.5/CHANGELOG_150.TXT`(1730 行,1.50.0→1.50.26 全部版本)
> 為主軸逐條過,聚焦「玩家會感受到的規則/數值變化」,跳過純 bug/crash/UI 技術修正;數字交叉核對
> `MANUAL_150.html`(1.50 patch notes 版,含明確 "1.31" 對照段落)與 `PARAMETERS.CFG`(3574 行,many
> 參數註解 `(default, classic)`,可直接判定「1.5 出廠預設是否等於 1.3」)。凡引用數字一律標行號/段落
> 出處;查無出處者標「待查」,不編造。

## 0. 結論先講:能落地程式碼的真正數值差異其實很少

逐條過完整份 CHANGELOG 後,**落在本專案已實作系統上、且是真正「數值改變」(非只是新增可調參數、
非純 timing/bookkeeping)的項目最初只找到 3 條**(§2),後續輪次陸續補上前置子系統、確證更多項目
真的落在已實作系統上(見下方 §0.5 進度追蹤,含本文件最新項目數的唯一真相)。CHANGELOG 裡另有
大量新增 `xxx_config_parameter`——這些**多半是把既有經典行為暴露成可調參數,
預設值本來就等於 1.3 經典值**(`PARAMETERS.CFG` 逐一標註 `(default, classic)`),不代表 1.5 出廠預設
真的改了規則。這對版本 profile 是好消息:第一版分版範圍可以很小(§6)。

## 0.5 實作進度追蹤(2026-07-11 起,每輪更新——「還差哪些」看這節)

> 對照 §1 全量表 15 項的當前實作狀態。狀態語意:**✅ 完成**(已接進遊戲迴圈或確認無需做)、
> **⚠ 半做**(公式/資料層在,但消費端未接)、**❌ 未排**(真缺口,需前置子系統)。commit 為
> 本 repo 提交雜湊。動手前先看這表,避免重做已完成項或誤把「確認非差異」當缺口。

| # | 項目 | 狀態 | 依據 / commit |
|---|---|---|---|
| 1 | 研究成本 Hyper-Advanced Lv1(15k/25k) | ✅ 完成 | `3eef521` 消費端接進 `engine.RunResearchPhase` + `EndTurn` 玩家/AI 注入 |
| 2 | 電漿砲傷害(30/20) | ✅ 完成 | `3eef521` `BuildShipWithMods` 改讀 `BuildWeaponOptions(RuleProfile)`,隨 `Ship.WeaponAttack` 進戰鬥 |
| 3 | 軌道轟炸齊射(5/10) | ✅ 完成 | `RuleProfile.BombardmentVolleys` 接 `fleetBombardDamage`(先前輪次) |
| **4** | **運輸艦淨現金(freighters_cash_bonus)** | ✅ **完成**(2026-07-11) | 新增「運輸艦隊」建造選項(`gamedata.FreighterFleetActionName`,前置科技 `TOPIC_NUCLEAR_FISSION`)補上先前缺的「貨運艦建造事件追蹤」子系統。完工時 `shell.GameSession.applySpecialAction`:`s.Player.ActiveFreighters += gamedata.FreighterFleetShipsPerBuild`(手冊 p.168 每次 +5 艘)+ `s.Player.BC += s.RuleProfile.FreightersCashBonus`(新欄位,1.3=5/1.5=0)。**簡化**:只模擬手冊表格「固定回饋」那一側,不模擬 0-3 BC 建造當下維護費立即扣款那一側(金額小,1.40+ 已改下回合扣);AI 對手未接同一建造流程,`ActiveFreighters` 對 AI 仍恆 0 |
| **5** | **防禦方指揮官 2.5x(defender commando)** | ✅ **完成** | `gamedata.GroundCommandoDefenderForceBonus` + `RuleProfile.DefenderCommandoBonus` 已接進 `InvadeColony` 守方(`internal/shell/ground_invasion.go`);前置的 AI 領袖資料模型(`AIOpponent.Leaders`)已補上,`buildDemoAIOpponents` 依種族性格開局固定指派(布拉西人 Tier2/姆瑞森人 Tier1/席隆人無)——**誠實近似**:非手冊逐字的隨機雇用機制,見 `session.go` `AIOpponent.Leaders`/`demoAIOpponentSetup.commandoTier` 欄位註解 |
| 6 | commando 倍率門檻 | ✅ 非差異 | `PARAMETERS.CFG:2745-2753`「(default, classic)」,1.5 預設=1.3,無需做 |
| **7** | **轟炸建築 +1hit(BombardmentBuildingBonusHits)** | ✅ **完成** | `b239f94` 隨 AI 建築模型接進 `BombardColony` 建築吸收(1.3 每棟 +1 hit);語意近似已標註 |
| 8 | civilian_armor(100hp) | ✅ 非差異 | `PARAMETERS.CFG:1778-1786`,值兩版相同;remake 採 hits 計數模型非 HP 模型,兩者未調和(誠實標註) |
| 9 | 防禦建築結構倍率(100) | ✅ 非差異 | `PARAMETERS.CFG:1772-1775` |
| 10 | 研究突破隨機性 | ✅ 非差異 | `PARAMETERS.CFG:542-545`;remake 研究本就無「突破機率」建模 |
| 11 | 行星尺寸轟炸分級(3-4-6-7-8) | ✅ 非差異 + 幾何已接線 | 1.5 系列內部自我修正回 classic;`GroundBombardPopulationLoss` 已用行星尺寸係數 |
| 12 | 起始偵察艦速度(10/12) | ✅ 等同 | remake `CombatSpeed()` 是統一公式=1.5 修正後行為;重現 1.3 的「自動 vs 手動設計不一致」bug 無意義 |
| **13** | **掃描/偵測距離** | ✅ **完成**(2026-07-11) | 新增 `internal/gamedata/detection.go`(掃描科技/軌道基地 parsec 查表 + 換算)+ `internal/shell/detection.go`(`GameSession.VisibleStars`/`starVisible`,啟用先前無人讀取的 `Star.Explored` 死旗標)+ `RuleProfile.SensorRangeVersionBonusParsec`(1.3=0/1.5=1);`cmd/moo2/interactive.go` `drawStarmap` 接上輕量戰爭迷霧(未偵測星降噪點、不畫名/擁有環)。**全面近似**:parsec 數值、換算常數皆無原版來源,詳見下方 §1 第 13 列與 `detection.go` 檔頭。fog 純視覺,不 gate 任何操作;不做敵艦 map blip(AI 無地圖座標) |
| **14** | **衛星/砲台佔格(beam arc cost)** | ✅ **完成**(2026-07-11) | `internal/gamedata/satellite.go` 把軌道基地/飛彈基地/地面砲台建模成「space 預算→塞入依科技解鎖的最佳武器」,beam 佔格套 `RuleProfile.SatelliteBeamArcCostPct`(1.3=25/1.5=33)、`GroundBatteryBeamArcCostPct`(1.3=0/1.5=50,CHANGELOG_150.TXT 1.50.7/1.50.10);`internal/shell/orbital_bombardment.go` `retaliationAttackers` 改用此模型取代舊 shipStrength 4/8/16 固定 tier。飛彈基地(300 space)/地面砲台(450 space)為手冊 p.78/p.81 確認值,星基/戰鬥站/星辰要塞(250/500/1200)是借用 ShipHullSpace 同量級的近似值;校準除數 `SatelliteStrengthScale=20` 使雷射參考點下星基/戰鬥站重現舊 tier 4/8,星辰要塞算出 20(非近似 19,見常數註解的誠實落差說明)。平衡 sanity(開局艦隊轟炸開局 AI 母星,Turn 0..14 掃描)兩版本最大損艦數皆為 1,見 `internal/shell/satellite_defense_test.go`。 |
| 15 | 維護費入帳時機 | ✅ 非差異 | 淨額 0,只差「哪一回合帳上出現」,remake 逐回合重算不模擬「完工瞬間」 |

**真正「還沒安排、且是真缺口」的**(需前置子系統,非公式缺口):目前**無**。#4(貨運建造事件
追蹤)、#13(掃描子系統)先前列在此處,已分別於 2026-07-11 補上前置子系統並接線完成,見上表
——舊版本文件曾把兩者列在此處,已更正(rulebook 63:錯誤斷言直接更正,不留 stale marker)。

15 項全量表現況:§2 三條真差異(#1/#2/#3)+ #4/#5/#7/#13/#14(2026-07-11 消費端接線完成)已完成;
#6/#8/#9/#10/#11/#15 為確認的非差異;#12 已等同 1.5。全量表 15 項至此**全數盤點完畢**。

## 1. 差異清單(全量表)

依「是否落在本專案已實作系統」排序;類別欄對應 WORKLIST.md 的系統分類。

| # | 類別 | 1.3(1.31)行為 | 1.5(1.50)行為 | 落在已實作系統? | 來源 |
|---|---|---|---|---|---|
| 1 | 研究成本(Hyper-Advanced Lv1) | **實際成本 15,000 RP**(顯示卻是 25,000,官方承認的 bug) | 顯示與實際皆 **25,000 RP**(1.50.9 修正) | ✅ 是,`internal/gamedata/techtree.go` 8 個 `TOPIC_HYPER_*` | CHANGELOG_150.TXT 1.50.9「Fixed actual tech cost of first level Hyper-Advanced from 15k to 25k.」+ MANUAL_150.html「Hyper-Advanced Tech Cost Bug: Fixed the cost of level 1 hyper-advanced tech fields that were shown as 25k research points but had a real cost of 15k. Now both actual and displayed cost is 25k RP.」 |
| 2 | 戰鬥傷害(電漿砲) | Damage **6–30** | Damage **4–20** | ✅ 是,`internal/shell/session.go` `WeaponOptions`「電漿砲」 | MANUAL_150.html「Plasma Cannon min/max damage from 6/30 to 4/20」(已收錄於 `docs/tech/component-values.md`) |
| 3 | 軌道轟炸命中換算 | Bomb 武器 = **5 次攻擊**當量;戰機 = **0**(無當量) | Bomb 武器 = **10 次攻擊**當量;戰機 = **1** 次 | ✅ 是,`internal/shell/orbital_bombardment.go` `fleetBombardDamage`(現行 `for round:=0;round<10` 即 1.5 值) | CHANGELOG_150.TXT 1.50.9「Fixed bomb hits calculation for orbital bombardment: Bomb weapons now get bomb hits equivalent to 10 instead of 5 attacks. Fighters get the equivalence of 1 strike instead of 0.」 |
| 4 | 經濟(新造運輸艦淨現金) | 完工時**淨得 +2~5 BC**(0-3 BC 立即成本 + 固定 5 BC 補償,官方手冊原文承認此為 1.3 就有的「一律有淨利」quirk,非等到 1.40 才算 bug) | `freighters_cash_bonus` 出廠預設 **0**(1.50.8 起),淨得 0 BC | ✅ 是(2026-07-11 完成)——`gamedata.IncomeFreighterMaintenanceCost`(每回合 0.5 BC/艘持續維護費)先前已寫好;本輪補上「完工當下的一次性現金效果」:新增「運輸艦隊」建造選項(`gamedata.FreighterFleetActionName`),完工時 `s.Player.ActiveFreighters += gamedata.FreighterFleetShipsPerBuild`(+5 艘)+ `s.Player.BC += s.RuleProfile.FreightersCashBonus`(新欄位,1.3=5/1.5=0)。**簡化**:只模擬手冊表格的「固定回饋」那一側,不模擬 0-3 BC 建造當下維護費立即扣款那一側 | MANUAL_150.html「Buildings & Freighters Free Cash Bug」全段(1.31/1.40/1.50 三欄對照表)+ CHANGELOG 1.50.8「Changed freighters_cash_bonus default from 5 to 0 BC」 |
| 5 | 地面戰:防禦方指揮官加成 | 防禦方 Commando 技能領袖**無**額外加成(僅攻方有 2.5x) | 新增:防禦方 Commando 領袖也給 **2.5x** 加成(攻方不變) | ✅ 是(2026-07-11 完成)——`gamedata.GroundCommandoDefenderForceBonus` + `RuleProfile.DefenderCommandoBonus` 已接進 `InvadeColony` 守方,前置的 `AIOpponent.Leaders` 資料模型(種族性格近似指派)已補齊 | MANUAL_150.html「Commando Leader: A defending commando gives 2.5x the regular commando bonus to ground troops, just like an attacking commando already gives in classic.」 |
| 6 | 地面戰:commando 倍率門檻 | 攻方 5x/7.5x、守方 2x/3x(依技能等級) | **出廠預設不變**,只是新增 `ground_commando_attacker_x2`/`ground_commando_defender_x5` 讓玩家可調換 | ❌ 否(且屬確認非差異) | `PARAMETERS.CFG:2745-2753`「(default, classic)」逐條標註 |
| 7 | 轟炸:建築 hits 加成 | 未記錄文件的 **+1 hit** bonus(bug) | 移除該 +1 bug | ❌ 否——本專案軌道轟炸只扣人口,不扣建築(見 `docs/tech/ground-combat-algorithm.md`「範圍限制」),此差異對本專案模型無作用 | CHANGELOG_150.TXT 1.50.10「Undocumented +1 hit bonus for civilian buildings during bombardment removed.」 |
| 8 | 轟炸:建築/人口裝甲值 | `civilian_armor`(非防禦建築/人口單位)= **100 hp**(所有裝甲等級皆同) | 出廠預設不變,只是暴露成可調參數 | ❌ 否(且屬確認非差異) | `PARAMETERS.CFG:1778-1786`「Default is 100 hp regardless of armor (classic).」 |
| 9 | 地面戰:防禦建築結構倍率 | `ground_defense_armor_multiplier` = **100**(對應鈦裝甲等級 100 結構點) | 出廠預設不變 | ❌ 否(確認非差異) | `PARAMETERS.CFG:1772-1775`「Default is 100 ... (classic).」 |
| 10 | 研究:突破隨機性 | `fixed_research_cost=0`(有隨機突破機率) | 出廠預設不變 | ❌ 否(確認非差異;本專案研究系統本來就沒模擬「突破機率」這件事,只有固定 RP 成本) | `PARAMETERS.CFG:542-545`「(default, classic)」 |
| 11 | 轟炸:炸彈換算的行星大小分級 | 3-4-6-7-8(classic) | 中途版本(1.50.4 起)曾錯改,**1.50.11 已修回同一組 3-4-6-7-8** | ❌ 否——本專案不模擬行星尺寸對轟炸區域的幾何影響;且此差異在 1.5 系列內部自我修正,對 1.3 vs 最終 1.5 而言**不構成差異** | CHANGELOG_150.TXT 1.50.11「Restored planet sizes for bombardment to classic 3-4-6-7-8.」 |
| 12 | 開局:起始偵察艦戰鬥速度 | 手動設計與自動設計艦速度不一致 bug,起始 2 艘 Scout 戰鬥速度為 **10** | 修正後為 **12**(空間格利用速度加成一致套用) | ⚠ 邊緣——本專案 `CombatSpeed()` 是通用公式(見 `formulas.go`),未特別為「自動設計 vs 手動設計」分流,無法乾淨對應此 bug;可視為低優先 | CHANGELOG_150.TXT 引文見 `docs/tech/homeworld-init.md:111`(已收錄) |
| 13 | 掃描/通訊距離 | Tachyon/Neutron/Sensors 等偵測科技的顯示值與實際值有 1 格落差(如 Tachyon 顯示 3、實際 4) | 修正「顯示值=實際值」,多數偵測距離**預設值 +1** | ✅ 是(2026-07-11 完成)——新增輕量戰爭迷霧子系統:`gamedata.ScannerRangeParsec`(基礎 2/Space 4/Neutron 6/Tachyon 8,**近似**遞增序,手冊無公開數字)+ `gamedata.OrbitalScannerBonusParsec`(星基+2/戰鬥站+4/星辰要塞+6,**近似**,擇一取代不疊加)+ `gamedata.ParsecToNormalized`(1/10,**近似**換算常數,依 `NewDemoSession()` 實際程序化星系鄰近星距離調參)+ `RuleProfile.SensorRangeVersionBonusParsec`(1.3=0/1.5=1,本欄位即對應本行「預設值 +1」的整體近似,非逐科技數字);`shell.GameSession.VisibleStars`/`starVisible` 用「已探索(啟用 `Star.Explored` 死旗標)∪ 玩家自己的星 ∪ 落在任一玩家殖民地/艦隊偵測範圍內」判定可見,`drawStarmap` 接上 fog 繪製(未偵測星降噪成暗點、不畫名/擁有環)。fog **純視覺**,不 gate 任何操作;**不做敵艦 map blip**(AI 艦隊無地圖座標,零地基) | MANUAL_150.html「Scanners and Communications Discrepancy」表(見下方 §4 附註,表格經 HTML 去標籤後欄位對不齊,**精確數字待用原始 HTML `<table>` 結構重新萃取,本表僅供方向性參考**——故本專案採統一 +1 parsec 近似,非逐科技逐字重現) |
| 14 | 衛星/地面砲台佔格 | 光束武器在衛星 arc cost +25%、地面砲台 +0% | 修正為統一 arc cost,衛星/地面砲台空間分別 +40%/+50%(1.50.7),之後衛星再由 +40%→+33.3%(1.50.10) | ✅ 是(2026-07-11 完成)——`internal/gamedata/satellite.go` 新增獨立的軌道基地/飛彈基地/地面砲台 space 預算模型(不是 6 級標準艦體表,是另一組專用常數),`RuleProfile.SatelliteBeamArcCostPct`(25/33)、`GroundBatteryBeamArcCostPct`(0/50)接進 `internal/shell/orbital_bombardment.go` `retaliationAttackers` | CHANGELOG_150.TXT 1.50.7、1.50.10 |
| 15 | 新建築/新間諜維護費入帳時機 | 完工當回合立即扣費 + 補償(淨額 0) | 改為下回合扣費、取消補償(淨額同樣 0) | ❌ 否(且屬淨額非差異——只差「哪一回合帳上出現」,本專案 `BuiltMaintenanceBC` 是逐回合依「目前已建成清單」重算,不模擬「完工瞬間」這個時間點,模型顆粒度不到這一層) | MANUAL_150.html「Buildings & Freighters Free Cash Bug」表 |

## 2. 落在已實作系統上的差異(核心,共 3 條——見上表 #1–#3)

這 3 條全部滿足:(a) 官方文件白紙黑字給出具體數字,(b) 對應到本專案**目前正在跑的程式碼路徑**
(非死碼、非未來計畫),(c) 我們現有實作值 = 1.5 的值(因為 `techtree.go`/`session.go`/
`orbital_bombardment.go` 的資料來源本來就是 patch 1.5 隨附的 `GAME_MANUAL.pdf`/`MANUAL_150.html`)。
換言之:**現況 = 事實上的「1.5 profile」**,要支援 1.3 選項,只需要一組「1.3 覆寫值」:

| 值 | 現有程式碼(= 1.5) | 1.3 覆寫值 |
|---|---|---|
| Hyper-Advanced Lv1 研究成本(8 個 `TOPIC_HYPER_*` 主題共用) | 25000 | 15000 |
| 電漿砲 Damage(`WeaponOptions` 電漿砲 `Value`) | 20 | 30 |
| 軌道轟炸模擬齊射輪數(`fleetBombardDamage` for 迴圈上限) | 10 | 5 |
| 軌道轟炸戰機命中當量(目前本專案未區分戰機類型,無對應變數——見下方待辦) | — | — |

電漿砲的手冊數字其實是「min/max 各改一次」(6→4、30→20),但本專案 `Component` 結構只有單一
`Value`(代表最大傷害)欄位,沒有 `MinValue`——目前的 20/30 對照只能覆寫「最大傷害」這一項;
若要精確重現最小傷害(4 vs 6),需要先幫 `Component` 加一個欄位,這是比「分版」更大的既有資料模型
擴充,本檔誠實標記為「1.3 profile 只能做到最大傷害對齊,最小傷害維持現行單值模型的既有限制」。

軌道轟炸「戰機命中當量」目前本專案的 `fleetBombardDamage` 是逐艦逐輪算傷害,沒有「戰機 vs 非戰機
武器」的分流(不像 `weaponKindByName` 那樣分 beam/missile/spherical),故 1.3 的「戰機 0 當量」/
1.5 的「戰機 1 當量」這條差異目前**無法對應到任何變數**——這是「發現了差異,但我們連 1.5 值本身
都还没有为它建模」的情況,誠實標「待未來戰機武器分類任務接線後再談分版」。

## 3. [★ 歷史快照,已於 2026-07-11 補上,核實以上方 §0.5/§1 #4 為準] 已實作但尚未接線的系統

> 本節寫下當時(#4 經濟)確實是「公式有、消費端無、機制本身還不存在」;**2026-07-11 同日稍後
> 已補上完整機制**(見 §0.5 #4 現況),下面段落保留當時的分析過程當歷史記錄,結論(「現在談分版
> 為時過早」)已不成立,不可再引用。

`gamedata.IncomeFreighterMaintenanceCost`(每回合持續維護費 0.5 BC/艘,手冊 p.169 有據)當時已寫好
但零呼叫端。手冊裡真正 1.3→1.5 有差的是**另一條機制**——「完工當下一次性補償」
(`freighters_cash_bonus`,1.3 淨賺 2-5 BC、1.5 淨賺 0 BC)——當時這條機制本專案完全沒有程式碼。
現況:新增「運輸艦隊」建造選項(`gamedata.FreighterFleetActionName`)補上建造事件追蹤,
`RuleProfile.FreightersCashBonus`(1.3=5/1.5=0)接進 `applySpecialAction`,詳見 §0.5 #4、
`docs/tech/moo2-formulas-reference.md`「運輸艦淨現金版本差異」節。

## 4. shipped-game 預設值來源(PARAMETERS.CFG `(default, classic)` 交叉核對)

`PARAMETERS.CFG`(patch 1.5 內附,3574 行)對每個可調參數多附**「Default is X (classic)」**這類註解,
代表「1.5 出廠預設值就是經典 1.3 的值,這個參數只是把它暴露出來讓玩家/mod 調」。逐一核對後確認
以下項目 **1.5 出廠預設 = 1.3 經典行為,非真差異**(對應上表 #6/#8/#9/#10):

- `ground_commando_attacker_x2`/`ground_commando_defender_x5`(地面戰指揮官加成倍率)——`PARAMETERS.CFG:2745-2753`
- `civilian_armor`(轟炸建築/人口裝甲值 100hp)——`PARAMETERS.CFG:1778-1786`
- `ground_defense_armor_multiplier`(地面防禦建築結構倍率 100)——`PARAMETERS.CFG:1772-1775`
- `fixed_research_cost`(研究突破隨機性,預設仍隨機)——`PARAMETERS.CFG:542-545`

這對「盡量對照 1.5 預設 vs 1.3 原值」的任務要求(§3)而言,誠實的答案是:**這幾項刻意查過,結果
是「找到明確數字,但數字本身兩版相同」**——不是找不到,是差異不存在。詳細版本消歧陷阱(為何不能
直接盲抄 CFG,`##` 標記的是「improved」mod 值 vs classic,不是「1.3 vs 1.5」这条轴)已由既有文件
`docs/tech/patch15-cfg-data-source.md` 完整記錄,本檔沿用同一套紀律,不重複展開。

## 5. 版本 profile 資料結構設計建議

### 5.1 最小可行設計(對應§2 三個確證差異)

> ★ 歷史快照:下面是 2026-07-11 當輪最初的設計稿(僅 3 欄位)。**已實作且持續擴充**——現行
> `internal/gamedata/ruleprofile.go` 欄位數已超過此處,以該檔為準,不要照抄下面的欄位清單當
> 現況(rulebook 63)。保留本段只為了展示「最小可行設計」的思路本身。

```go
// internal/gamedata/ruleprofile.go(建議新檔,尚未實作,設計稿——已實作,見上方 ★ 說明)

package gamedata

// GameVersion 對應 CLAUDE.md「主選單選擇 1.3 或 1.5」的兩個選項。
type GameVersion int

const (
	VersionClassic13  GameVersion = iota // 官方最後正式 patch 1.31
	VersionCommunity15                    // 社群非官方 patch 1.50(本專案現行資料的預設來源)
)

// RuleProfile 收斂「已確證、且落在已實作系統上」的版本相依數值。
// ⚠ 刻意保持精簡——只收 docs/tech/version-1.3-1.5-diff.md §2 列出的 3 條,不要為了「看起來完整」
// 塞進未確證或未接線的欄位。新差異確證後才加欄位,見 §5.3 擴充路徑。
type RuleProfile struct {
	Version GameVersion

	// 研究:Hyper-Advanced 第一級科技(8 個 TOPIC_HYPER_* 主題共用同一個成本)。
	// 來源:MANUAL_150.html「Hyper-Advanced Tech Cost Bug」+ CHANGELOG_150.TXT 1.50.9。
	HyperAdvancedLevel1Cost int

	// 戰鬥:電漿砲最大傷害(Component.Value)。來源見 component-values.md。
	// 注意:手冊同時記載最小傷害 6→4,但 Component 結構目前只有單一 Value(最大傷害)欄位,
	// 無法表示最小值差異——這是既有資料模型限制,非本 profile 遺漏。
	PlasmaCannonMaxDamage int

	// 軌道轟炸:fleetBombardDamage 模擬齊射的輪數。來源:CHANGELOG_150.TXT 1.50.9。
	BombardmentVolleys int
}

func Profile13() RuleProfile {
	return RuleProfile{
		Version:                 VersionClassic13,
		HyperAdvancedLevel1Cost: 15000,
		PlasmaCannonMaxDamage:   30,
		BombardmentVolleys:      5,
	}
}

func Profile15() RuleProfile {
	return RuleProfile{
		Version:                 VersionCommunity15,
		HyperAdvancedLevel1Cost: 25000, // = 現行 techtree.go 硬編值
		PlasmaCannonMaxDamage:   20,    // = 現行 session.go 硬編值
		BombardmentVolleys:      10,    // = 現行 orbital_bombardment.go 硬編值
	}
}
```

### 5.2 接線點(現有程式碼要改的三處,設計層面,尚未實作)

| 現有程式碼 | 目前寫法 | 建議改法 |
|---|---|---|
| `internal/gamedata/techtree.go` 8 個 `TOPIC_HYPER_*` 條目 | `Cost: 25000` 硬編 | `researchChoices` 這 8 條的 `Cost` 改成讀 `activeProfile.HyperAdvancedLevel1Cost`(需把 `researchChoices` 從套件級 `var` 改成依 profile 產生的函式,或在 `ResearchChoiceFor` 內對這 8 個 topic 特判覆寫) |
| `internal/shell/session.go` `WeaponOptions` 電漿砲 | `{"電漿砲", 200, 20, ...}` 硬編 | 建構 `WeaponOptions` 時對電漿砲那一列的 `Value` 改讀 `activeProfile.PlasmaCannonMaxDamage`(`Component` 陣列從套件級 `var` 改成 `func BuildWeaponOptions(p RuleProfile) []Component`,呼叫端一次性建構,不必每次查表) |
| `internal/shell/orbital_bombardment.go` `fleetBombardDamage` | `for round := 0; round < 10; round++` 硬編 | 迴圈上限改讀 `s.RuleProfile.BombardmentVolleys`(`GameSession` 需新增 `RuleProfile` 欄位,由 `NewDemoSession`/新遊戲流程注入) |

`GameSession` 需要新增一個 `RuleProfile` 欄位(建構時由主選單選擇結果注入),這是唯一貫穿全專案的
新增狀態;`RuleProfile` 本身應視為**唯讀設定**,遊戲開始後不可變(避免中途切版本造成存檔/平衡
不一致——原版本身也是「一開局就決定規則集」,無 mid-game 切換)。

### 5.3 擴充路徑(未來新差異確證後怎麼加)

1. 新差異先進本檔 §1 表格(來源標行號/段落),分類「已實作系統」/「未實作」。
2. 只有「已實作系統」且「兩版數字真的不同」(§4 教訓:很多新增參數的 1.5 預設=1.3 經典值,不算差異)
   才加進 `RuleProfile` 欄位 + `Profile13()`/`Profile15()`。
3. 若差異牽涉到現有資料結構容不下的維度(如電漿砲最小傷害,`Component` 沒有 `MinValue` 欄位)——
   分兩步走:先評估「值不值得為此擴充資料結構」,再擴充,不要為了塞版本差異而扭曲既有結構。
4. `RuleProfile` 預期會隨著本專案更多系統忠實化(戰機分類、衛星船體、運輸艦建造事件等)持續變大,
   但**不要提前為未確證/未接線的系統占位欄位**——空欄位比缺欄位更容易誤導後續開發誤以為「已考慮
   版本差異」。

## 6. 第一版最小分版建議(可直接排入 WORKLIST)

**只做 §2 三個值**:`HyperAdvancedLevel1Cost`(15000/25000)、`PlasmaCannonMaxDamage`(30/20)、
`BombardmentVolleys`(5/10)。理由:

- 三者皆有官方文件逐字數字佐證(非社群逆向、非推測)。
- 三者皆落在本專案「已在跑」的程式碼路徑上,接線成本低(見 §5.2,三處都是把硬編常數換成讀
  profile 欄位,無需改資料流向或新增系統)。
- 三者對玩家可感知(研究成本、武器傷害、轟炸強度),符合任務「玩家會感受到的規則變化」門檻。
- **誠實揭露**:本節是 2026-07-11 當輪最初的「第一版最小分版建議」,寫下當時多數 CHANGELOG
  條目要嘛是純 bug fix(如衛星生成 bug)、要嘛是新增的可調參數但預設值等於經典值(§4)、要嘛
  落在本專案當時尚未實作的系統(戰機分類/領袖加成/運輸艦建造事件/衛星 space 模型等)。**這不是
  研究不夠深入,是逐條核實後的真實結論**——之後每完成一個新系統(#5 領袖加成、#7 轟炸建築 hits、
  #14 衛星 space 預算模型、#4 運輸艦建造事件追蹤皆已在同一輪陸續補上,見 §0.5 追蹤表當前狀態),
  就回頭比對 CHANGELOG 是否有該系統的版本差異,再擴充 `RuleProfile`——#13(掃描子系統)已於
  2026-07-11 同批補上(輕量戰爭迷霧,見 §0.5/§1 第 13 列),全量表 15 項至此全數盤點完畢,
  無仍未排項目。

## 7. 主選單「選版本」與 profile 的關係(範圍澄清)

CLAUDE.md 要求「主選單選擇版本 1.3 or 1.5」——本檔的 `RuleProfile` 解決的是**選版本後,遊戲規則
數值要跟著變**這一半;「主選單 UI 本身要有這個選項」是另一半(UI/流程層,不在本檔研究範圍,但
握手位置很單純:新遊戲流程在建立 `GameSession` 前先決定 `GameVersion`,傳入
`Profile13()`/`Profile15()` 其中之一)。兩者可分開排入 WORKLIST 的不同任務。

## 8. 來源清單

- `moo2_patch1.5/CHANGELOG_150.TXT`(1730 行,1.50.0–1.50.26 全部版本逐條核對)
- `moo2_patch1.5/MANUAL_150.html`(1.50 patch notes 版手冊,經 `python3 -re` 去 HTML 標籤後全文
  關鍵字比對;「Scanners and Communications Discrepancy」表格因去標籤後欄位錯位,標記待用原始
  `<table>` 結構重新萃取)
- `moo2_patch1.5/MOO2-1.50.26.zip` 內 `patch/150/docs/PARAMETERS.CFG`(3574 行,`(default, classic)`
  註解逐條核對 §4 四項)
- 既有專案文件(交叉引用,未重複研究):`docs/tech/component-values.md`(電漿砲差異原始發現)、
  `docs/tech/patch15-cfg-data-source.md`(CFG 版本消歧陷阱,本檔沿用其紀律)、
  `docs/tech/ship-design-space.md`(艦體/武器空間表,確認衛星非本專案範圍)、
  `docs/tech/ground-combat-algorithm.md`(地面戰現行公式,確認領袖加成未實作)、
  `docs/tech/research-system-status.md`(研究成本現行接線點)、
  `docs/tech/colony-economy-maintenance.md`(運輸艦維護費現行接線狀態)、
  `docs/tech/homeworld-init.md`(起始偵察艦速度差異原始發現)、
  `docs/tech/rules-implementation-audit.md`(已預先點名本任務為 Phase 7 待辦)
