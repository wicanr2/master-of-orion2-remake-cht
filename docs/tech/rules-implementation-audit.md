# openorion2 遊戲規則實作程度盤點

> 目的:給 Phase 5(Gameplay 引擎重建)團隊一份「openorion2 到底有沒有規則邏輯可抄」的逐系統盤點。
> 方法:對每個系統實際 `Read` 相關 `.cpp/.h`,追每個候選函式「有沒有被呼叫」「是不是真的算邏輯」,
> 不以 struct 欄位存在或函式名稱當證據。對照手冊來源見文末「參考資料」。

## 總結

openorion2 是**存檔載入與檢視殼,不是回合制遊戲引擎**。全 repo(`src/*.cpp` 約 16,600 行核心邏輯檔)
對 `endTurn|nextTurn|processTurn|advanceTurn` 零命中,`main.cpp` 的兩條執行路徑(載入存檔看星系圖 /
開主選單)完全不含任何回合處理呼叫。更關鍵的是**全 repo 對 `rand()/srand()/mt19937/random_device` 零命中**——
一個沒有任何隨機數來源的程式碼庫不可能算得出戰鬥命中率、殖民地成長、AI 決策這類手冊明訂「有隨機成分」的規則,
這比「找不到 combat 函式」更能說明問題的本質:**不是漏抄某幾個函式,而是規則引擎整層不存在**。
本次逐一驗證 11 個系統後,僅發現「唯讀衍生顯示公式」(從存檔欄位算一個顯示數字,不改變任何狀態、
不含隨機性)——如艦艇戰力查表、研究剩餘 RP、軍官雇用費、軍官技能加成——集中在 `gamestate.cpp`。
這些公式本身正確且可直接複用(已收錄於 `formulas.md`),但它們是「已知輸入→顯示輸出」的算式,
不是「回合推進→狀態改變」的引擎。全 repo 22 處 `STUB(view)`(`guimisc.h:27` 定義)涵蓋了幾乎所有
「會改變遊戲狀態」的按鈕:結束回合、開新遊戲、雇用/解雇軍官、艦隊搬遷/報廢、殖民地列表、種族列表、
縮放。外交與間諜甚至連 STUB 畫面殼都沒有——沒有 `DiplomacyView`/`SpyView` 這類類別存在。
給 remake 團隊的定位:**資產解碼器與存檔 schema 是可放心複用的地基,但 11 個系統的「規則」都要從手冊
+ 其他重製專案(1oom、FreeOrion)重新設計與實作,不能指望在 openorion2 找到可抄的核心邏輯。**

## 總表

| 系統 | 實作程度 | 關鍵原始碼位置 | 缺口/待補 |
|---|---|---|---|
| 1. 星系/星圖生成 | ❌未見 | `gamestate.cpp:238-280`(`Galaxy::load/validate`,純解序列化)、`galaxy.cpp:2994`(`MainMenuWindow::newGame` = `STUB`) | 星系形狀/星星分布/行星屬性/特殊天體隨機生成算法整個要重寫;連「新遊戲設定」畫面殼(星系大小/玩家數/種族選擇)都不存在(`mainmenu.h` 只有 `MainMenuView`/`LoadGameWindow` 兩個類別) |
| 2. 殖民地成長與生產 | ⬜僅資料結構無邏輯 | `gamestate.cpp:298-462`(`Colony::Colony/load/validate`)、`colony.cpp` 全檔(僅 `ColonistPickerWidget` UI 排序/繪製)、`gamestate.cpp:518-520`(`Planet::baseProduction` 單一查表) | 人口成長公式(手冊 `Notes on Population Growth` 的 `POPRACE/POPAGG/POPMAX` 曲線)、工人分配→食物/工業/研究產出換算、污染累積、建築加成、稅率→BC 全部要從零實作;`surplusFood/surplusBC/pollution/pop_growth[]` 等欄位只被讀不被算(`galaxy.cpp:1942-1979` 純顯示) |
| 3. 研究樹 | 🟡部分 | `tech.cpp:169`(`research_choices[]` 資料表,含 cost/prerequisite 分組)、`gamestate.cpp:1250-1269`(`Player::researchCost`,唯讀公式)、`tech.cpp:1008`(`ResearchSelectWindow::techSelected` = 硬編 `"Not implemented yet"` 訊息框) | 科技樹拓撲資料可用,但「選定研究主題並逐回合累加 RP、達標解鎖」的引擎不存在:`researchProgress` 全 repo 除建構子/讀檔外從未被賦值(`grep researchProgress\s*=` 僅 `gamestate.cpp:1047` 初始化);`canResearchTopic`(`gamestate.cpp:1242`)只讀存檔裡的 `researchTopics[]` 狀態,不是靠前置科技圖推導 |
| 4. 艦艇設計與戰鬥 | ❌未見(戰鬥)/ ⬜(設計資料) | `gamestate.cpp:841-934`(`ShipDesign::computerHP/driveHP/combatSpeed/beamOffense/beamDefense`,唯讀屬性公式)、`ships.cpp` 全檔(`FleetListView`/`ShipGridWidget` 純清單 UI) | 戰鬥解算(命中率、傷害套用、多回合戰術戰鬥、戰略戰鬥模擬)全無,且無 RNG 可用;`ships.cpp` 找不到任何 `ShipDesignWindow`/設計介面類別,自訂艦艇設計 UI 連殼都不存在;`clickRelocate`/`clickScrap`(`ships.cpp:1205-1207`)是 `STUB` |
| 5. 地面戰(登陸/轟炸) | ❌未見 | 全 repo 對 `invade/invasion/bombard/marine/assault` 命中僅止於 `gamestate.h`/`tech.cpp` 的科技與建築**列舉常量**(`TECH_MARINE_BARRACKS`、`TECH_TROOP_PODS`、`TECH_ASSAULT_SHUTTLES`),無任何函式 | 手冊 `Notes on Orbital Assault`(炸彈命中模擬、建築/人口/儲存產能命中判定)、地面部隊登陸戰完全要從零設計,openorion2 無可抄 |
| 6. 外交 | ⬜僅資料結構無邏輯(無畫面殼) | `gamestate.h:594-599`(`DIPLO_NONE`..`DIPLO_WAR` 列舉)、`gamestate.cpp:2010-2029`(`GameState::validate` 只檢查 `playerRelations[i][j]==playerRelations[j][i]` 等雙向對稱性,屬存檔完整性檢查非外交邏輯) | 全 repo 找不到任何 `DiplomacyView`/`DiplomacyWindow` 類別——連 STUB 殼都沒有;`playerRelations/foreignPolicies/tradeTreaties/researchTreaties` 除讀檔外從未被賦值,AI 決策、關係隨事件變化(手冊 `AI & Diplomacy` 提到的 Diplomatic Blunder/Marriage 事件)全無 |
| 7. 間諜 | ⬜僅資料結構無邏輯(無畫面殼) | `gamestate.h:90-93`(`SPY_MISSION_STEAL/SABOTAGE/HIDE` 位元遮罩)、`gamestate.cpp:1990`(僅驗證 `spies[j]` 欄位合法範圍) | 找不到任何 `SpyView`/`SpyWindow` 類別;`spies[]` 除讀檔外從未賦值;手冊 `Notes on Spying` 的技能加成表(種族/政府/領袖/科技/派駐間諜數疊加)、任務成功率、被抓後果全無實作 |
| 8. 軍官/英雄 | 🟡部分 | `gamestate.cpp:607-701`(`Leader::expLevel/hasSkill/skillBonus/hireCost`,真實唯讀公式)、`officer.cpp:713-792`(`askAssignOfficer` 有 ETA 判斷邏輯)、`officer.cpp:715,1205`(`assignOfficer`/`HireLeaderWindow::clickHire` 硬編 `"Not implemented yet"`) | 技能加成/雇用費/階級公式可直接複用;但「指派軍官到艦隊/星球」與「雇用軍官」在使用者按下確認後**都不落地**(不是用 `STUB` 巨集,而是另外硬寫同款訊息框,搜尋 `STUB(` 會漏掉這兩處,需額外搜 `"Not implemented yet"`);`experience` 欄位全 repo 除建構子/讀檔外從未增加,無戰功/任職累積經驗機制 |
| 9. 種族特性效果 | ⬜僅資料結構無邏輯(有唯讀顯示畫面) | `gamestate.h:727-760`(`RaceTrait` 列舉 32 項)、`info.cpp:860-900`(種族資訊畫面,真實可運作,唯讀顯示 `pptr->traits[j]`)、`gamestate.cpp:2554`(`dump()` 除錯輸出,非遊戲邏輯) | `traits[]` 只被讀來顯示數值/名稱,找不到任何函式把種族特性數值套用到生產/戰鬥/成長公式(因為這些公式本身也不存在);Phase 5 需把手冊種族特性表(生產加成、戰鬥加成、外交加成等)接進新引擎的對應公式 |
| 10. 勝利條件 | ❌未見 | 全 repo 對 `victory/winner/win_condition/gameOver` 零命中 | 手冊 `Notes on Winning the Game` 三種勝利路徑(殲滅全部對手、三分之二超級多數票選、經次元傳送門攻陷 Antares 母星)與計分公式(時間分/人口分/殲滅加分)全無,需從零設計含分數結算與遊戲結束畫面 |
| 11. Antaran/Orion 事件 | ❌未見(僅美術/常量) | `gamestate.h:219`(`ORION_SPECIAL = 11` 星球特殊類型列舉)、`ships.cpp:54-131`(Antaran 艦船/Orion 守護者僅美術 sprite 載入,`PALSPRITE_ANTARAN`/`PALSPRITE_GUARDIAN`) | Antaran 隨機襲擊觸發、Orion 守護者遭遇戰、擊敗守護者後的科技獎勵(手冊:預設死亡射線科技)、次元傳送門攻打 Antares 母星的觸發鏈全無邏輯,只有可重用的美術資源索引 |

## 分系統補充

### 1. 星系/星圖生成
`Galaxy::load`(`gamestate.cpp:242-254`)只是把 `sizeFactor/width/height/nebulas[]` 從存檔流依序讀出;
`Galaxy::validate`(`gamestate.cpp:256-280`)只檢查 `sizeFactor` 落在四個合法檔位(10/15/20/30,對應
`galaxySizeFactors[]`,`gamestate.cpp:29`)與星雲座標落在galaxy範圍內——這是存檔完整性檢查,不是生成演算法。
`Star::load`(`gamestate.cpp:1420-1486`)同樣是逐欄位讀取。`galaxySizeFactors[]` 唯一的另一個用途是
`galaxy.cpp:1385,1599,1603` 的畫面縮放係數換算,與生成無關。`MainMenuWindow::newGame`
(`galaxy.cpp:2994`)是 `STUB(_parent)`,且 `mainmenu.h` 裡連「新遊戲設定」畫面類別都不存在——這代表
openorion2 連 UI 殼都沒替這個系統搭好,不是「邏輯缺、殼在」的情況。手冊 `Galaxy Generation` 一節(1.50
patch notes)描述的是 mapgen 既有 bug 修正(星雲內黑洞替換、最近星系保底可殖民行星、衛星數量上限
250 的裁剪機制),這些都是 remake 要重新設計的生成規則參考點。

### 2. 殖民地成長與生產
`colony.cpp` 全檔只有 `ColonistPickerWidget` 一個類別,功能是把 `_colony->colonists[]` 依種族/職業/身分
(俘虜)分組排序後畫成頭像列(`update()` `colony.cpp:81-130`、`redraw()` `colony.cpp:183-217`)——純顯示
排版,不含任何產出換算。`Colony` 類別(`gamestate.h`)本身只有建構子/`load`/`validate` 三個成員函式
(`gamestate.cpp:298,372,462`),欄位如 `pollution`、`pop_growth[MAX_RACES]`、`food_per_farmer` 全部
`stream.readUint8/readUint16LE` 直接載入,從未被任何公式重算(`grep food_per_farmer` 全 repo只有讀檔賦值
與 `colony.cpp:207` 的顯示判斷)。`Planet::baseProduction()`(`gamestate.cpp:518-520`)是唯一真正的公式,
但只是「礦產豐度→基礎產出倍率」單一查表(5 檔:1/2/3/5/8),不是完整的人口分配→產出計算鏈。
手冊 `Notes on Population Growth` 給出的成長公式(`POPRACE/POPAGG/POPMAX` 搭配 general/race/AI/tech/
leader/event/housing 七項加成)在 openorion2 沒有任何對應實作,`surplusFood`/`surplusBC` 等玩家層級
彙總欄位在 `galaxy.cpp:1942-1979` 只被 `buf.printf` 顯示,不被計算。

### 3. 研究樹
`tech.cpp` 絕大部分是 `TechListWidget`/`ResearchSelectWidget`/`ResearchSelectWindow`/
`ResearchListWindow` 四個 UI 類別(科技清單捲動、分頁、高亮、點擊選取的畫面邏輯)。真正的資料價值在
`research_choices[MAX_RESEARCH_TOPICS]`(`tech.cpp:169`起,例如 `{650, 0, {TECH_MERCULITE_MISSILE,
TECH_POLLUTION_PROCESSOR}}`)——這是完整的科技樹拓撲(每個研究槽的花費與可選科技分組),對 remake
有直接參考價值。`Player::researchCost`(`gamestate.cpp:1250-1269`)用這個表算出「還差多少 RP」,是
可靠的唯讀公式,已收錄 `formulas.md`。但真正決定「規則有沒有跑」的是
`ResearchSelectWindow::techSelected`(`tech.cpp:1008-1011`):
```cpp
void ResearchSelectWindow::techSelected(int x, int y, int arg) {
	new MessageBoxWindow(_parent, "Not implemented yet");
	close();
}
```
玩家點選研究主題後,選擇不會寫回 `Player::researchTopic`,`researchProgress` 全 repo 除
`gamestate.cpp:1047`(建構子歸零)與讀檔外沒有第二處賦值——沒有回合結算函式替它累加 RP。手冊
`Research & Technology` 一節提到「turn 0 產生的 RP 若當回合未選科技不會遺失」這類回合結算細節,在
openorion2 完全無對應邏輯可比對,因為根本沒有回合結算。

### 4. 艦艇設計與戰鬥
`ShipDesign` 的六個公式函式(`gamestate.cpp:841-934`:`maxComputerHP`/`computerHP`/`maxDriveHP`/
`driveHP`/`combatSpeed`/`beamOffense`/`beamDefense`)與 `formulas.md` 記載的完全一致——這些是「已知
装備損傷狀態→顯示戰力數字」的唯讀公式,用於艦隊清單畫面(`ships.cpp` 的 `generateShipInfo`,
`ships.cpp:933-1106`)顯示裝備血條與戰力估算,**不是戰鬥解算**。全 repo 對戰鬥相關字串
(`combat`/`Combat`)的命中在 `ships.cpp` 全部是 `ShipGridWidget` 的「作戰艦 vs 支援艦」分類篩選函式
(`selectedCombatCount`/`filterCombat` 等,`ships.cpp:375,853`),與 `galaxy.cpp`/`galaxy.h` 對
combat/battle 零命中相互印證:openorion2 沒有任何戰術或戰略戰鬥的傷害/命中判定函式,且全 repo 零
RNG 來源,無法產生手冊 `Tactical Combat`/`Strategic Combat` 描述的命中率、主動權排序等隨機或半隨機
結果。艦艇「自訂設計」介面(選裝備、排插槽)在 `ships.h`/`galaxy.h`/`info.h` 中找不到任何對應類別,
是完全空白,不是 stub 殼。`FleetListView::clickRelocate`/`clickScrap`(`ships.cpp:1205-1207`)是
`STUB`,搬遷與報廢艦隊在畫面上按得下去但無行為。

### 5. 地面戰
地面戰(登陸部隊、轟炸、佔領)在 openorion2 沒有任何函式層級的存在。搜尋
`invade/invasion/troop/bombard/marine/assault` 全部命中都落在 `gamestate.h` 的科技/建築/裝備**列舉常量**
(`TECH_MARINE_BARRACKS`、`TECH_TROOP_PODS`、`SPEC_TROOP_PODS`、`TECH_ASSAULT_SHUTTLES`)與
`tech.cpp` 的研究樹資料表項目——這些只是科技樹裡「有一個叫這個名字的研究/建築選項」,沒有任何函式
讀取這些數值做地面戰鬥計算。手冊 `Ground Defenses & Troops`、`Notes on Orbital Assault` 描述的完整
規則(戰鬥機隊出擊/歸位、地面部隊上限與重建、軌道轟炸模擬 10 回合射擊、建築/人口/儲存產能各自的
命中判定)在 remake 必須從手冊逐條重建。

### 6. 外交
`DIPLO_NONE`..`DIPLO_WAR`(`gamestate.h:594-599`)是外交狀態的列舉值,`Player` 類別確實有
`playerRelations`/`foreignPolicies`/`tradeTreaties`/`researchTreaties`/`playerContacts` 這些欄位,但
唯一讀寫它們的地方是 `GameState::validate`(`gamestate.cpp:2010-2029`),而且做的事是「檢查玩家 i 對 j
的關係值是否等於玩家 j 對 i 的關係值」這種**雙向對稱性檢查**(存檔完整性驗證),不是外交邏輯——沒有
任何函式依據事件、AI 個性、種族特性去*改變*這些數值。更關鍵的是,全 repo 沒有 `DiplomacyView`/
`DiplomacyWindow` 這類畫面類別,`galaxy.h`/`info.h` 都沒有,說明這個系統連「按鈕會跳出 not implemented」
的殼都不存在,遠比外掛 STUB 的系統更空白。手冊 `AI & Diplomacy` 一節提到的 AI 六種目標性格
(Diplomat/Ecologist/Expansionist/Industrialist/Militarist/Technologist)判定與外交事件(Diplomatic
Blunder、Marriage)對關係值的影響,都要在 remake 從零設計。

### 7. 間諜
`SPY_MISSION_STEAL/SABOTAGE/HIDE`(`gamestate.h:91-93`,以 `SPY_MISSION_MASK = 0xc0` 取高兩位元)
是任務類型的位元遮罩定義,`Player::spies[]`(對每個對手一組間諜位元狀態)與 `spyMaintenance`
維護費欄位都存在,但 `GameState::validate`(`gamestate.cpp:1990`)同樣只做「`spies[j]` 值落在合法遮罩
範圍內」的存檔完整性檢查。全 repo 找不到 `SpyView`/`SpyWindow` 或任何任務結算函式,`spies[]` 除讀檔
外沒有第二處賦值。手冊 `Notes on Spying`/`Spying` 兩節給出完整的技能加成表(政府型態
Feudalism/Dictatorship/Imperium/Democracy 等各自的防禦/攻擊加成、每個間諜插槽依派駐人數遞減的加成
規則、63 名防禦/攻擊間諜上限),這些量化規則在 openorion2 沒有任何對應實作痕跡,需要從手冊全部
重建。

### 8. 軍官/英雄
這是本次盤點裡「唯讀公式」品質最好的系統:`Leader::expLevel()`(`gamestate.cpp:607-617`)用
`leaderExpThresholds[] = {60, 150, 300, 500, 0}`(`gamestate.cpp:68`)把累積經驗值分級;
`Leader::hasSkill()`/`skillBonus()`(`gamestate.cpp:631-692`)正確處理了 2-bit 技能等級編碼、
領航員技能的特殊 tier/等級對照表(`navigatorSkillValues`)、進階技能 +50% 加成;`hireCost()`
(`gamestate.cpp:700-701`)是手冊定義的雇用費公式 `10 × skillValue × (expLevel+1) + modifier`。這些都
是可直接搬進 remake 的正確算式。但「軍官系統會不會動」要看兩個提交點:`officer.cpp` 的
`askAssignOfficer()`(`officer.cpp:718-792`)其實做了一段真實的**判斷邏輯**——計算重新指派軍官是否
需要 `LEADER_MOVE_TIME` 移動時間(同艦隊/同星系內免費、跨星系要等)——但按下確認後呼叫的
`assignOfficer()`(`officer.cpp:713-716`)是:
```cpp
void LeaderListView::assignOfficer(int x, int y, int arg) {
	cancelSelect(0, 0, 0);
	new MessageBoxWindow(this, "Not implemented yet");
}
```
指派永遠不落地。雇用同樣:`HireLeaderWindow::clickHire()`(`officer.cpp:1204-1207`)也是同款硬編
訊息框,不是用 `STUB()` 巨集寫的,單純 `grep "STUB("` 會漏掉這兩處,必須另外搜尋
`"Not implemented yet"` 字串。`experience` 欄位(`gamestate.cpp:563` 初始化為 0)除讀檔外全 repo
沒有第二處賦值,代表沒有戰功或任職時間累積經驗的機制。

### 9. 種族特性效果
`RaceTrait` 列舉(`gamestate.h:727-760`)定義了 32 項特性(政府、人口、農業、工業、科學、金錢、
艦艇防禦/攻擊、地面戰、間諜、低重力/高重力/水生/地底、大型/富饒/文物母星、電子化、食礦、
排斥/魅力、無創造力/有創造力、寬容、天生商人、心靈感應、幸運、全知、隱形艦艇等)。`info.cpp:860-900`
的種族資訊畫面是**真的可運作**的畫面(對應 `01-openorion2-assessment.md` 已列的「可運作」清單),會
把 `pptr->traits[j]` 逐項讀出並轉成文字顯示(`gameLang->estrings(ESTR_GOVERNMENT_NAMES +
pptr->traits[TRAIT_GOVERNMENT])` 等)。但這是唯讀顯示,不是效果套用:因為第 2、4 項(殖民地生產、
戰鬥)本身沒有公式,`traits[]` 自然也沒有機會被拿去乘進任何加成計算。`gamestate.cpp:2554` 附近的
`traits[j]` 讀取屬於 `GameState::dump()` 除錯輸出函式(`fprintf` 系列),同樣不是遊戲邏輯。Phase 5
把種族特性接進新引擎時,要對照手冊種族特性表逐項定義加成套用點(生產倍率、戰鬥加成、外交修正等),
openorion2 只能提供「有哪 32 項特性」的清單,不能提供「怎麼套用」的邏輯。

### 10. 勝利條件
全 repo 對 `victory|winner|win_condition|gameOver` 大小寫不敏感搜尋零命中,`galaxy.cpp`/`gamestate.cpp`
都沒有任何判斷「遊戲是否結束」的函式,也沒有計分函式。手冊 `Notes on Winning the Game` 定義三條勝利
路徑(殲滅全部對手、三分之二超級多數票選為銀河領袖、經由次元傳送門攻陷 Antares 母星)與詳細計分
公式(依星系大小/玩家數的起始時間分、每回合遞減、每單位人口 +1 分、每殲滅一位玩家 +50 分等),這些
在 remake 裡屬於「純規則設計」,openorion2 沒有任何可比對或可抄的程式碼。

### 11. Antaran/Orion 事件
`ORION_SPECIAL = 11`(`gamestate.h:219`)是星球特殊類型列舉裡的一項(其餘特殊類型見
`savegame-format.md`/`enums.md`),`Star::validate`(`gamestate.cpp:1539-1541`)與 `Planet::validate`
(`gamestate.cpp:552`)只檢查這個值落在合法範圍。`ships.cpp:54-131` 載入 Antaran 艦船
(`PALSPRITE_ANTARAN`、`MAX_SHIPTYPES_ANTARAN = 5`)與 Orion 守護者(`PALSPRITE_GUARDIAN`)的美術
sprite,供艦隊清單畫面顯示用——這些 ID/資源索引對 remake 的美術管線有參考價值,但沒有任何函式
處理「Antaran 隨機來襲」「進入 Orion 系統觸發守護者戰鬥」「擊敗守護者後給死亡射線科技」「用次元
傳送門攻打 Antares 母星觸發終局戰」這類事件鏈邏輯,manual 對這幾個事件的敘述在 openorion2 找不到
任何對應程式碼。

## 後續建議

1. **不要在 openorion2 裡找戰鬥/外交/間諜/勝利條件的邏輯——它們不存在,連畫面殼都常常沒有。**
   直接依手冊 + 其他重製專案(1oom 的 C 引擎、FreeOrion 的規則實作)設計新引擎,openorion2 只提供
   資料 schema(存檔欄位)與極少數唯讀顯示公式,兩者角色不同不要混淆。
2. **`STUB()` 巨集只是全部缺口的一部分**,`officer.cpp` 已發現至少兩處(`assignOfficer`、
   `HireLeaderWindow::clickHire`)用硬編 `MessageBoxWindow(..., "Not implemented yet")` 達到相同效果但
   繞過巨集——後續若要完整列出「所有非功能按鈕」清單,除了 `grep "STUB("` 還要再搜
   `"Not implemented yet"` 字串,避免漏記。
3. **`research_choices[]`(`tech.cpp:169`)與 `Leader` 的技能/經驗/雇用費公式群
   (`gamestate.cpp:607-701`)是本次盤點中價值最高的可複用資產**,可直接對照手冊驗證後搬進 Go
   `internal/gamedata`,比殖民地/戰鬥系統的「僅列舉常量」價值高一個量級。
4. **殖民地成長公式(手冊 `Notes on Population Growth`)與軌道轟炸模擬(`Notes on Orbital Assault`)
   是規則細節最複雜、最需要提前排入 Phase 5 設計文件的兩塊**——兩者都涉及多項加成疊加與是否四捨五入
   的邊界規則(1.50 patch notes 特別記載了多個因取整方式錯誤導致的經典版 bug),移植時要先決定「抄
   1.31 經典公式」還是「抄 1.50 修正後公式」,並在 `docs/tech/` 記下版本差異,避免後續 Phase 7
   (1.3/1.5 差異研究)重工。
5. **全 repo 零 RNG 來源**這件事本身要寫進 Phase 5 的架構決策:新引擎的隨機數生成器（含 seed 管理、
   存檔內是否要存 RNG 狀態以支援可重現的多人同步)要當一等公民設計進去,不能延用 openorion2 的任何
   殘留假設。

## 參考資料

- 手冊來源:`/home/anr2/moo2/moo2_patch1.5/MANUAL_150.html`(1.50 patch notes,含 Galaxy Generation /
  Notes on Population Growth / Research & Technology / Tactical Combat / Strategic Combat /
  Ground Defenses & Troops / Spying / Notes on Spying / AI & Diplomacy / Notes on Winning the
  Game / Notes on Orbital Assault / Notes on Special Systems 等章節)。
  `original_game/Master of Orion 2 - CD Manual.pdf` 經 `pdfinfo` 確認僅 9 頁,內容是安裝指南
  (System Requirements / Installing the Game),**不是完整規則手冊**,本次盤點未使用其作規則來源,
  如需 1.31 基礎版規則全文,需另外取得原版遊戲內建說明書或社群 wiki。
- 既有文件:`docs/kickoff/01-openorion2-assessment.md`(整體完成度盤點,結論一致並在此擴充逐系統細節)、
  `docs/tech/formulas.md`(本文件多處引用的唯讀公式已由該檔收錄)、`docs/tech/enums.md`/
  `savegame-format.md`(資料結構背景)。
