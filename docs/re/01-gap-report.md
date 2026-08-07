# 總缺口報告:原版 Orion2.exe vs remake

> 日期:2026-08-06(當日稍晚修訂)。方法:解析原版執行檔內建的 8,589 個 Watcom 除錯符號
> (見 [`00-orion2-symbols.md`](00-orion2-symbols.md)),對照 remake 現行程式碼。
> **這是第一次能用原版二進位當基準做全面盤點**——先前只能靠手冊、攻略與 openorion2(純渲染殼)。

## ⚠ 本報告初版的一處錯誤已修正

初版依據的符號表 parser 把記錄格式讀反(位址其實在名字**之前**),導致
**name↔addr 全部錯開一格**。修正細節與方法論教訓見
[`00-orion2-symbols.md`](00-orion2-symbols.md)。

對本報告的影響:

- **Part A(畫面清單)與 Part B(模組分群)基本不受影響**——它們只用到名字與模組歸屬,
  而同一 `.c` 模組的符號連續,位移一格只動到模組邊界的一兩個名字。重算後
  module 122(歷史記錄)74 個函式、module 15(事件)58 個,與初版的 73/57 差一。
- **Part C(資料表)整段重寫**:初版對照到的是**相鄰那張表**的內容,所以才會出現
  「`_food_per_farmer_table` 的值是 64..101,名字一定掛錯」這種結論。修正後
  該表就是手冊 p.59 的十氣候食物值,名字沒錯。

## 可信度分級(每條結論都標)

| 級別 | 意義 |
|---|---|
| **A 硬證** | 符號名 + 位址,且已用「讀它的函式怎麼索引」交叉確認語意 |
| **B 推論** | 由符號名語意推斷用途,未反編驗證 |
| **C 待驗** | 反編結果含 `JUMPOUT`(IDA 函式邊界錯誤),不可採信 |

⚠ 本報告的「原版有 X」一律是 **A 級**(符號存在是事實);
「X 的具體行為/數值是 Y」則需個別確認,未確認的列在待深挖清單。

---

## Part A — 畫面缺口(原版 53 個 vs remake 22 個)

### A-1 已對上(remake 有對應畫面)

| 原版 | remake | 備註 |
|---|---|---|
| `Main_Menu` | `menu` | ✅ 座標已對齊;2026-08-07 補上原版的「無存檔時 Continue / Load Game 灰階停用」|
| `Mainmenu_Load_Game_Popup_` @ 0x804B7 / `Do_Save_Game_Popup_` @ 0x7E154 | `loadGame` / `saveGameInPlay` | ✅ 2026-08-07 已建(`cmd/moo2/loadgame.go`),十格存檔選單,存讀共用同一視窗 |
| `GameMenuWindow`(遊戲中的「遊戲」選單)| `gameMenu` | ✅ 2026-08-07 已建(`cmd/moo2/gamemenu.go`),星系畫面頂端「遊戲」鈕先前是死的 |
| `Newgame` | `newGameSetup` | ⚠ 原版底排是 **PLAYERS**(對手數),remake 誤作 RACE 入口 |
| `Race_Selection` | `raceSelect` | ⚠ 版面左右相反(原版肖像左/2欄按鈕右) |
| `Racial_Option` | `customRace` | 自訂種族點數 |
| `Flag` | `nameFlag` | ⚠ 原版命名與旗幟是**兩個獨立畫面**,remake 合併且用色塊非旗幟圖 |
| `Main` / `Main_Main` / `Mini_Main` | `galaxy` | 原版有 mini 變體,remake 無 |
| `Colony_Summary` | `colonySummary` | |
| `Planet_Summary` / `Mini_Planet_Summary` | `planets` | |
| `Fleet` | `fleet` | |
| `Design` / `Darkened_Design` | `shipDesign` | |
| `Officer` | `officer` | |
| `Race` | `races` | |
| `Diplomacy` / `Diplomacy_Fade_In` | `diplomacy` | |
| `Main_Council` | `council` | |
| `Main_Combat` / `Super_Fast` | `tacticalCombat` | `Super_Fast` = 快速結算畫面 |
| `Turn_Summary` / `Dummy_Turn_Summary` | `turnSummary` | |

### A-2 ❌ 原版有、remake 完全沒有(依對玩家的影響排序)

| # | 原版畫面 | 影響 | 備註 |
|---|---|---|---|
| 1 | ~~**`Colony`**(獨立殖民地畫面)~~ | 高 | ✅ 2026-08-06 已建(`cmd/moo2/colonyscreen.go`),從總覽點殖民地名進入;含 7 格建造佇列。版面未對齊原版(無 COLONY.LBX 版面資料),結構與流程已對 |
| 2 | ~~**`Event`**~~ | 高 | ✅ 2026-08-06 已建(`cmd/moo2/eventscreen.go`),GNN 新聞快報樣式;事件表同步換成原版 36 種 |
| 3 | ~~**`Tech_Review`**~~ | 中高 | ✅ 已建(`cmd/moo2/infosubscreens.go` `infoTechReview`);當初「誤接成研究選擇跳板」的 issue #5-2 已修,研究選擇改由星系研究框進入 |
| 4 | ~~**`History`**~~ | 中高 | ✅ 已建(`infoHistory`),國力折線圖,可點擊切換人口/國庫/艦隊指標 |
| 5 | ~~**`Race_Stats`**~~ | 中 | ✅ 已建(`infoRaceStats`) |
| 6 | ~~**`Reference_Main` / `_Category` / `_How_To`**~~ | 中 | ✅ 已建(`infoReference` + `cmd/moo2/inforeview.go`),remake 合成單一「參考資料」分頁而非 3 個子畫面 |
| 7 | ~~**`Command_Points`**~~ | 中 | ✅ 2026-08-07 已建(`cmd/moo2/commandpoints.go`,原版 `Show_Command_Points_Screen_` @ 0x8BAB9),從星圖右欄第 2 格點進去 |
| 8 | ~~**`Colony_Landing` / `Colony_Combat` / `Colony_Bombing`**~~ | 中 | ✅ 2026-08-07 全部已建(`cmd/moo2/groundcombat.go`、`cmd/moo2/bombing.go`),版面座標取自反組譯(見下方第 15、18 項)|
| 9 | ~~**`Main_Antaran_Room`**~~ | 中 | ✅ 2026-08-07 已建(`cmd/moo2/antaranroom.go`),用原版 `antaroom.LBX` 資產 1(55 幀累積)當背景;留白:原版是推鏡動畫,remake 取最終定格 |
| 10 | ~~**`Hall_Of_Fame` / `Hi_Score`**~~ | 低 | ✅ 2026-08-07 已建(`cmd/moo2/hiscore.go` + `gamedata/score.go`),八項計分係數全來自反組譯 module 60 |
| 11 | ~~**`Smack`**~~ | 低 | ✅ 已建(`cmd/moo2/cutscene.go` + `internal/smk`,真的解 Smacker,不是靜態圖)|
| 12 | 多人連線 11 個畫面 | — | ✅ `MP_Setup`(`cmd/moo2/multiplayer.go`)與 `Hotseat`(`cmd/moo2/hotseat.go`)2026-08-07 已建,版面座標取自反組譯(見下方第 20 項)。`Net_Next_Turn`(第 75 項)與 `Choose_Net_Plyrs`(第 76 項)2026-08-07 已建。`Modem_Setup`/`NullModem_Setup`/`Comm Info` **不做**(硬體已不存在)。`Join_Net`/`Generic_Net_Info`/`SendGet_Net_Info` 是同一張畫面的不同狀態(第 77 項),`Choose_Multi_Net_Game` 見第 78 項。**11 張全部結案:8 做 / 3 不做** |

### A-3 remake 有、原版無獨立畫面

`research` / `researchChoice`(remake 自建的研究選擇)、`battleResult`、`info`(原版 INFO 是容器,子畫面才是實體)。
→ **不是錯**,但代表 remake 的研究流程與原版結構不同,值得反編確認原版怎麼進研究選擇。

---

## Part B — 系統缺口(原版 331 個原始碼模組)

`module` 欄位揭露原版的子系統分解。以下是**遊戲邏輯類**大模組與 remake 的對照:

| 原版模組 | 函式數 | 代表符號 | remake 對應 | 落差 |
|---|---|---|---|---|
| 74 | 106 | `N_Colonies_And_Outposts_At_Star_`、`N_Bldgs_` | `internal/shell`(含 `outpost.go`)| ✅ 前哨站已建(2026-08-06),可升級成殖民地 |
| 48 | 86 | `Absolute_Location_`、`Contact_With_One_Colony_` | `shell/colonization.go` | 星圖拓樸/接觸判定較簡化 |
| **102** | **84** | `_minerals_per_mine`、`_climate_maintenance_modifiers`、`Colony_Officer_` | `gamedata/colony.go` 等 | **殖民地經濟核心**,權威數值全在此 |
| 14 | 83 | `Diplomacy_Screen_`、`Get_Main_Repulsive_Diplomacy_Choices_` | `cmd/moo2` diplomacy | 原版有完整外交選項樹 |
| 47 | 80 | `Design_Name_`、`Build_Saved_Ship_Array_` | `shell/shipnames.go` | |
| **122** | **73** | `Record_History_`、`Bill_Init_`、`Is_Ignoring_` | `shell/history.go` | ✅ 已建,`infoHistory` 的國力折線圖由它供資料 |
| 141 | 71 | `Load_Font_File_`、`Get_String_Width_` | `internal/uifont` | |
| 28 | 67 | `Get_Ship_Combat_Bonuses_`、`Init_Ship_Designs_` | `gamedata/combat.go` | |
| 58 | 67 | `Load_Officer_Picture_`、`Assert_Marooned_Leaders_` | `shell` 領袖 | 原版有 marooned leader 機制 |
| **15** | **71** | `Init_Events_`、`Check_For_Event_` | `shell/events.go` | 事件清單已對齊原版 36 種(16 種已實作,其餘缺子系統,見 `gamedata/events.go`) |
| **27** | **56** | `Init_Diplomatic_Relations_`、`Diplomacy_Growth_`、`Change_Relations_` | `internal/diplomacy` | 原版關係演化更完整 |
| 20 | 54 | `Apply_Internal_Damage_`、`Repair_Combat_Ship_` | `gamedata/damage.go` + `shell/repair.go` | ✅ 艦艇損傷與修復已建(手冊 p.80/82 + `Repair_Ships_At_Colonies_` 錨定)|
| 18 | 40 | `Safe_To_Fire_Sphere_Weapon_`、`Ai_Self_Destruct_Check_` | `shell/weapon_kind.go` | 原版戰鬥 AI 較深 |
| 19 | 42 | `Refresh_Combat_Screen_Full_` | `tacticalCombat` | |
| 65 | 37 | `Draw_Generic_Beam_`、`Draw_Ship_Burst_` | 戰鬥特效 | remake 特效較簡 |
| 138/314/316 | 141 | `Set_Music_File_`、AIL/Miles 驅動 | `internal/audio` | remake 直接播 PCM(已證等價) |
| 112/293 | 117 | Netmox / Hayes Modem | 無 | 多人連線 |

### ⚠ 「最大的系統級缺口」那份清單**四條全部已經做掉了**(2026-08-07 逐條核實)

這一節原本列了四大缺口,並且被後續每一輪的摘要反覆引用。逐條 grep 之後:

| 原本的斷言 | 現況 | 證據 |
|---|---|---|
| ①**歷史記錄系統**——remake 完全沒有,History Graph 做不出來 | ✅ 已建 | `internal/shell/history.go`(6 個函式)+ `cmd/moo2` 的 `infoHistory` 折線圖 |
| ②事件系統 | ✅ 2026-08-06 已對齊 | 36 種事件表 + GNN 快報畫面;剩餘 20 種缺的是各自的子系統,逐項記在 `gamedata.RandomEvents` 的 `Needs` 欄 |
| ③**前哨站(Outpost)**——remake 只有殖民地 | ✅ 已建 | `internal/shell/outpost.go`(9 個函式)+ 進存檔 + 進熱座席位 + 可被 `consumeOutpostForColony` 升級成殖民地 |
| ④**艙損/維修**(module 20)| ✅ 已建 | `internal/shell/repair.go`(11 個函式),手冊 p.80/82/25 逐字 + `Repair_Ships_At_Colonies_` @ 0x580F5 雙重錨定 |

**教訓與第 38 項同一條**:斷言一旦寫進文件就會被當成現況引用,而程式碼會往前走、文件不會。
下結論前先 grep —— 這一整節的四條斷言,每一條都只要一次 `ls` 就能推翻。
(規則出處:`rulebook/63-truth-in-code-not-stale-markers.md`、CLAUDE.md「每一輪盤點、清除錯誤斷言」。)

### 核實過後,真正還缺的(2026-08-07)

> **這張表是 CLAUDE.md 指向的「唯一活來源」,所以它必須是準的。**
> 做完一項就在同一個 commit 裡把它從表上拿掉——留著會被當成現況引用,
> 讓人去做已經做完的事(2026-08-07 已經發生兩次)。

| 缺口 | 性質 | 備註 |
|---|---|---|
| **網路多人** | 整塊子系統 | ~~決定性化~~(72)、~~傳輸層 + 鎖步~~(73)、~~指令解譯器~~(74)、~~`Net_Next_Turn` 等待畫面~~(75)、~~連線大廳 + `Choose_Net_Plyrs` 名冊~~(76)、~~連線狀態面板 7 狀態~~(77)、~~區網探索 + `Choose_Multi_Net_Game`~~(78)、~~文字輸入彈窗~~(79)。**整塊完成**:傳輸層 + 鎖步 + 決定性 + 指令層 + 大廳 + 區網探索 + 6 張畫面 + 輸入框 + ~~聊天列~~(112:資料模型與畫面先前就做完了,缺的是中間**沒有任何 goroutine 在讀連線**——補上訊息幫浦 `netplay.Session` 之後,字才真的離開這台機器);3 張明確不做(硬體已不存在)。**剩下的是留白不是缺口**:幫浦建好後中途加入不支援(對局開打後不收人,碰不到);`Modem_Setup` / `NullModem_Setup` / `Comm Info` 三張**不做**(數據機與序列線硬體已不存在,remake 走 TCP——替不存在的硬體做設定畫面不是還原) |

| **手冊數值忠實度** | 橫跨全系統的**類別**,不是單一子系統 | 2026-08-08(第 111–127 項)才被辨識出來的一整類缺口:**規則寫好了但沒接、或接了但用的是自編數字**。這一天找到並修掉 17 項,其中不乏會直接影響玩法的(護盾減傷五級錯四級、武器傷害排序反了、AI 從不研究、征服人口免費全額生產)。**方法本身是可重複的**,見下方「怎麼繼續找」 |

> **⚠ 2026-08-08:這張表先前寫「只剩一列」,那是準的但也是誤導的。**
> 準,是因為表上列的都是**子系統級**的洞,而那些確實只剩網路多人一項(現已了結)。
> 誤導,是因為**還有一整類缺口這張表的形狀裝不下**——它們不是「某個子系統沒做」,
> 而是「做了但不忠實」。第 111–127 項那 17 項全部屬於後者,沒有任何一項會在
> 「哪個子系統沒做」的盤點裡出現。
>
> **怎麼繼續找(這一天用過、還有殘量的四種問法):**
>
> | 問法 | 抓到的例子 | 殘量(2026-08-08 收盤) |
> |---|---|---|
> | ① 哪些函式**從來沒被執行過**(一局 300 回合的覆蓋率) | 領袖維護費、征服人口 3/4 產出 | **見底** |
> | ② 哪些匯出常數**零消費** | 護盾減傷(五級錯四級) | **見底**(190 個,多數是原版列舉鏡像) |
> | ③ 哪些參數**被餵固定值** | 種族特性整叢(第 129 項)、`apNegated`、`hefBonus`、`ftlLevel`(第 136 項) | **gamedata 匯出函式那一側幾乎見底**(剩 5 小項);⚠ **cmd/ 與 shell 內部函式沒掃過**——第 136 項就是在 cmd/ 撞到一個硬編的 `ftlLevel=1` |
> | ④ 手冊有而 remake **沒抄的資料** | 武器傷害全表、裝甲倍率階梯、飛彈防禦家族 | **10 個元件**(第 133 項盤點 20 個 + 第 128 項的 p.127 特殊武器,第 134–139 項接掉 10 個)|
>
> ③ 的掃描器寫在第 129 項:配對括號抽出每個 `gamedata.X(...)` 的引數,逐參數位置看是不是
> 每個呼叫端都給同一個字面值。**跑一次就把整叢種族特性的洞挖出來**——那一叢先前
> 一個都不在任何清單上。剩下的 5 個小項:`ArtemisMineDamage(shieldClass=0)`、
> `DamageSphericalRoll(ordnancePercent=100)`(軍械技能)、`RebellionChancePermille`
> 的 `ownerIsHuman`/`exterminating`、`GroundPlanetTotalHits` 的戰車欄位(顯示用估計值)、
> `AIEnemyColonyValue(shiftToSelf)`。
>
> ④ 的盤點器寫在第 133 項:手冊標題的 `名稱 (System|Ship)` 對四張元件表自動比對。
> **下一輪的料就是那 20 個**(戰鬥艙、強化船體、保安站、戰鬥掃描器、匿蹤力場、
> 相位匿蹤、時間扭曲加速器、瞄準器兩型、超載電容、快速飛彈架、轟炸機庫、多相護盾、
> 能量吸收器、傳送器…)。
>
> ⚠ **這一輪最該記住的一件事**,不是任何一項的內容:
> 「擋門理由當時成立、之後沒有人回頭看」在第 111–133 項裡撞了**七次**
> (第 117/118/122/123/130/131/133)。每一次那句話都寫得很清楚、很有道理,
> 而且都是對的——**在寫下的那一天**。沒有任何機制會在它失效時提醒你。
> 唯一有效的做法是**定期重問**,而不是相信註解。
>
> ⚠ 表上原本的其餘七列在 2026-08-07 被逐一了結(第 65–71 項),移到下方「已了結」表。
> ⚠ 「只剩一列」講的是**這張表**,不是「remake 完成了」——
> 表外還有一整批「已實作但仍需對原版校準」的項目,那些看 `docs/HONEST-STATUS.md`。

**2026-08-07 逐條核實後從這張表刪掉的**(做完了,但表沒跟上):

| 原本的斷言 | 現況 |
|---|---|
| **同星系多殖民地** | ✅ 第 66 項:拓殖/前哨站的對象改成**行星**;人造行星改造完的天體也真的能殖民了 |
| **戰術格子的獨立戰機單位** | ✅ 第 69 項:中隊在格子上有自己的位置(出擊→飛向目標→貼身開火→返航補給→再出擊)。⚠ 同時訂正「中隊數 4/2」——那是把 p.127 的 **Shots** 欄讀成了中隊人數,正確答案是**一律 4 架** |
| AI 的遷移設定 | ✅ **不是缺口**(第 71 項):那個欄位的五個寫入者沒有一個在 AI 的程式碼裡——原版的 AI 也不設集結點 |
| AI 的同星系多殖民地 | ✅ 第 67 項:`aiExpand` 的候選集加進「自己已有殖民地的星系」;順帶修好入侵提早翻面的 bug |
| 遷移連線的顯示開關沒有 UI | ✅ 第 65 項(⚠ 那一列不是原版版面,原版有一整個設定畫面) |
| `Clear_All` 集結點沒有 UI 入口 | ✅ 第 70 項。⚠ 同時訂正——**ALL 鈕不是 Set_All**,手冊兩處明說它是「全選/全不選艦艇」 |
| 遷移確認框 | ✅ 第 68 項(版面取自 `Confirmation_Box_`)。⚠ 同時訂正——原版那個條件是**怪獸**不是艦隊 |
| `Command_Points` 專屬畫面沒做 | ✅ `cmd/moo2/commandpoints.go`(第 40 項) |
| 殖民地地表缺「擺放的最後一段抖動」 | ✅ 第 42 項 |
| 殖民地地表缺植被層 | ✅ 第 43 項(COLVEGGI,13 組 × 8 張 = 104 對上資產數) |
| 艦隊圖示 8 幀「動畫沒做」 | ✅ 第 54 項:**原版就不會動**(每次繪製前歸零幀號),假缺口 |
| 一般星球閃爍「三個常數沒追出來」 | ✅ 第 54 項:**出貨版是死碼**(沒有任何地方啟動它),假缺口 |
| 鍵盤快捷鍵 | ✅ 第 55 項全數接完(F1/F2、F5/F6、F9、F10、ALT+F9)。ALT+F1..F8 的設定開關仍不碰(PDF 邊欄標籤有 off-by-one 風險) |
| 星圖遷移連線 | ✅ 第 57 項。**星圖 4 層全部到齊**:蟲洞(35)、遷移連線(57)、星星、星門(47)+ 星雲(44)+ 外交燈號(50) |
| 多艦隊模型 | ✅ 第 56 項(資料層)+ 第 58 項(艦隊列表)+ 第 60 項(分艦隊) |
| RELOCATE 鈕的原版語意 | ✅ 第 59 項:兩段點選 + 四條合法性規則,已接 |
| 分艦隊 UI | ✅ 第 60 項(⚠ 入口是 remake 自己加的,原版在右側艦艇格) |

⚠ 這份清單同樣會過期。**引用前先 grep**,別把它當成永久事實。
——這一次就是:上面刪掉的三條,每一條都只要一次 `ls cmd/moo2/` 就能推翻,卻在表裡留了一整天。

---

## Part C — 資料表(修正後重寫)

原版把數值放在具名資料表,**這些是比手冊更權威的數值來源**(手冊會簡化、會有筆誤——
專案先前已抓到手冊自身的 AMR 命中率與飛彈速度矛盾)。

### C-1 已釘死並接進 remake(2026-08-06)

星系/行星生成整組骰表已 dump、確認語意、寫進 `internal/gamedata/galaxygen.go`,
星系生成改用原版模型。逐格數值見 [`00-orion2-symbols.md`](00-orion2-symbols.md);
語意由「哪個原版函式怎麼索引它」釘死,不是靠名字猜:

| 表 | 讀它的原版函式 | 索引 |
|---|---|---|
| `_star_class_table` | `Generate_Spectral_Class_` | `[spectral*3 + age]` |
| `_planet_size_table` | `Generate_Size_` | d10 累計骰表 |
| `_class_to_group` | `Get_Planet_Group_` | `[spectral*5 + orbit]` |
| `_normal_gal` / `_old_gal_climate_roll_table` | `Generate_Climate_` | `[climate*4 + group]` |
| `_class_to_mineral` | `Generate_Mineral_Class_` | `[(d10-1)*6 + spectral]` |
| `_gravity_table` | `Generate_Gravity_Class_` | `[mineral*5 + size]` |
| `_class_to_num_satellites` | `Generate_Number_Of_Satellites_` | `[(d10-1)*6 + spectral]` |
| `_planet_max_farms` | `Generate_Max_Farms_` | `[size]` |
| `_food_per_farmer_table` | `Generate_Food_Per_Farmer_` | `[climate]` |

### C-2 已交叉驗證,remake 數值本來就對(不必改)

| 表 | 結論 |
|---|---|
| `_food_per_farmer_table` = `0 0 0 1 1 2 2 1 2 3` | 與手冊 p.59 逐格相同 |
| `_minerals_per_mine` = `1 2 3 5 8` | 與手冊礦產豐度五級相同 |
| `_planet_max_population` = `5 10 15 20 25` | 等於 remake 既有的 `(size+1)*5` |

**這三項可以撤銷「手冊可能簡化過」的存疑**——原版硬編值與手冊一致。

### C-3 仍是缺口

| 原版資料表 | 用途 | remake 現況 |
|---|---|---|
| ~~`_personality_*` **14 張**~~ | **AI 性格行為** | ✅ 2026-08-06 已接(`ai/personality_tables.go`)。實際是 **14 張**不是 6 張,每張 7 欄對應 Personality 0-6 |
| ~~`_base_planet_values` / `_g_*`~~ | **AI 行星估值** | ✅ 2026-08-06 已移植 4 個公式(`gamedata/ai_planet_value.go`):`Uncolonized_Planet_Worth_To_Player_`(選星)、`Proximity_Worth_To_Player_`(距離)、`Compute_Contextual_Planet_Values_`(星系協同)、`Colony_Worth_To_Player_`(已殖民星,供 AI 挑攻擊目標)。剩 `Enemy_Colony_Worth_To_Player_` |
| ~~`_climate_maintenance_modifiers`~~ | 氣候維護成本 | ✅ 語意已確認:索引 = 氣候,由 `Uncolonized_Planet_Worth_To_Player_` 以 `[planet.climate]` 讀。值 = `[50,25,0,25,0,0,0,0,0,0]`,已用於 AI 估值;**尚未接進殖民地實際維護費** |
| `_planet_max_mines` = `2 4 6 9 12` | 各大小礦場上限 | 已建表,**尚未接進生產**(remake 無礦場上限概念) |
| ~~`_planet_special` / `_planet_special_weighted_chance`~~ | 行星特殊物產(12 種,權重和 100) | ✅ 2026-08-06 已整套接進(`gamedata/planet_special.go` + `shell/discovery.go`),見下方 C-4 |
| ~~`_ranged_to_hit_penalty` / `_ranged_damage_penalty`~~ | 射程命中/傷害懲罰(各 9 個 word) | ✅ 兩張表 remake 早已有且逐格相同(手冊值);傷害衰減已於 2026-08-06 接進 `ResolveShotWithMods` |
| ~~`_orbit_to_satellite_type`~~ | 行星類別(氣態巨星/小行星帶/一般行星) | ✅ 2026-08-06 維度已釘死並接進生成器,見下方 C-5 |
| `_spy_bonuses` | 間諜加成 | remake 一律 0(標 TODO) |
| `_ability_costs` | 自訂種族點數成本 | 用 patch1.5 config(已對) |
| `_tech_research_level_values` | 科技研究等級 | `gamedata/techtree.go` |
| `_high/_low/_moderate_*_values`(9 張) | 疑似 AI 難度曲線 | 未知 |

**最大的剩餘數值缺口**:①間諜加成 ②氣候維護費接進殖民地開銷 ③礦場上限。

### C-5 `_orbit_to_satellite_type`:維度是這樣釘死的(2026-08-06)

表是 50 bytes,擺法有 10×5 與 5×10 兩種可能,光看數字分不出來。決定性的證據是**表裡唯一
的那個 `4`**:它落在 (roll 1, orbit 0),而 `Generate_Satellite_Type_` 處理 4 的特例分支
寫死 `bl == 1 && orbit == 0`。兩邊完全咬合,擺法不可能是別的。

| roll | 軌道0 | 軌道1 | 軌道2 | 軌道3 | 軌道4 |
|---|---|---|---|---|---|
| 0 | 小行星帶 | 小行星帶 | 小行星帶 | 小行星帶 | 小行星帶 |
| 1 | ★特例 | 小行星帶 | 小行星帶 | 小行星帶 | 氣態巨星 |
| 2 | 一般 | 氣態巨星 | 小行星帶 | 氣態巨星 | 氣態巨星 |
| 3-4 | 一般 | 一般 | 氣態巨星 | 氣態巨星 | 氣態巨星 |
| 5 | 一般 | 一般 | 一般 | 一般 | 氣態巨星 |
| 6-9 | 一般 | 一般 | 一般 | 一般 | 一般 |

內圈岩石、外圈氣態,roll 越大整個系統越宜居——分布本身就有物理直覺,是「沒讀錯」的旁證。

**★特例**(10% 機率)在原版會依恆星光譜寫進一個 ≥4 的類別碼(`spectral==0 ? 5 : spectral+4`),
openorion2 的 `enum PlanetType` 只定義 1-3,那些碼的語意目前無從確認,remake 一律當小行星帶
處理並另用一個 bool 標出來,不臆造。

### ⚠ Random_ 語意訂正(連鎖影響三處)

`Random_` @ 0x1247A0 回傳的是 **1..n**,不是 C 慣例的 0..n-1(LCG 取樣、拒絕超界值,
最後 `div bucket` 再 **`inc eax`**)。本報告與程式碼先前把它當成 `rand()%n`,連帶錯了三處:

1. **遠古文物送幾項科技**:`Random_(4)/4+1` 不是「恆為 1」,而是 **1 項、25% 機率 2 項**。
2. **蓄水池抽樣** `Random_(k)==1`:原版是**正確的** 1/k;先前記的「第一個候選永遠不會被選中」
   是誤讀(k=1 時 `Random_(1)` 必回 1)。
3. **失散殖民地在不可耕行星上的職務** `Random_(2) & 3`:是**工人或科學家**,不是農夫或工人。

訂正的方法是回頭讀 `Random_` 本身,而不是繼續從呼叫端推語意——`_orbit_to_satellite_type`
的 `roll = Random_(10) - 1` 落在 0..9(剛好對上 10 列)也反過來佐證了這件事。

**手冊後來獨立證實了訂正後的結論**:p.60 System Specials 講遠古文物時寫
「the first empire to discover the system gets **one or two** free technology advancements」
——「一或兩項」,正是 `Random_(4)/4+1` 在 1..n 語意下的值域。

### C-4 行星特殊物產:手冊沒給的數字,反組譯給了(2026-08-06)

這一項值得單獨記,因為它是**「手冊不足以還原、非讀執行檔不可」的典型案例**:
手冊只說太空殘骸/海盜藏寶的所得「is added to your treasury」,金額一個字都沒提。

| 特殊物產 | 效果 | 來源 |
|---|---|---|
| 太空殘骸(2) | 抵達星系 → 國庫 **+50 BC** | `Do_System_Discoveries_At_Star_` @ 0xE9927:`add dword [player+32h], 32h` |
| 海盜藏寶(3) | 抵達星系 → 國庫 **+100 BC** | 同上,`add … 64h` |
| 金礦(4) | 殖民地 +5 BC/回合 | 手冊(AI 估值加分 1280 佐證) |
| 寶石礦(5) | 殖民地 +10 BC/回合 | 手冊(AI 估值加分 2560 = 金礦兩倍,比例一致) |
| 原住民(6) | 殖民時**額外 3 個人口單位**、全為農夫、每農夫 +2 食物,之後該 special 消失 | `Make_New_Colony_Or_Outpost_` @ 0xE5EB3:迴圈 colony+0x10→0x1C(stride 4),`[colony+0Ah]=4`,`[planet+0Fh]=0` |
| 失散殖民地(7) | 抵達星系 → 就地生出殖民地,人口 = **min(該行星人口上限, 3)** | `cmp al,3 / jbe / mov byte [colony+0Ah],3` |
| 受困英雄(8) | 抵達星系 → 免費得一名領袖 | 手冊 + 原版該分支只設訊息碼 |
| 遠古文物(10) | 抵達星系 → **白送 1 項可研究科技(25% 機率 2 項)**;殖民後每科學家 5 研究 | 掃 204 個研究主題挑 `RSTATE_READY` 者蓄水池抽樣;送幾項 = `Random_(4)/4+1`,而 `Random_` 回 1..n(見下方「Random_ 語意訂正」) |

欄位偏移怎麼確定不是猜的:同一個函式裡的行星指標 stride = 0x11 = 17 bytes = openorion2
`struct Planet` 的大小,而 `[planet+0x0F]` 對上該結構的 `special`;殖民地那邊
`[colony+0x0A]`(population)、`[colony+0x0C]`(colonists[])、`[colony+0xE2]`(climate)
三個偏移同時對上 `struct Colony`;人口單位的位元佈局(race bits0-3 / loyalty 4-6 / job 7-8)
與 `Colonist::load` 逐位元相同;原住民寫進去的 race id 是 **9**,而 openorion2 的
`MAX_RACES = MAX_PLAYERS+2` 註解寫明「player races + androids + natives」——8 機器人、9 原住民。
最後,符號表裡獨立存在一個 `Planet_Has_Splinter_Colony_`,內容正好是 `[planet+0x0F] == 7`。

**六個獨立證據互相咬合,不是靠單一位址猜語意。**

---

## 優先序(依「對還原度的槓桿 ÷ 成本」)

### 原版畫面流程(反編 call graph,2026-08-06)

從 `.asm` 的 call 關係過濾出畫面主迴圈函式,得到原版的畫面跳轉:

```
Main_Menu_Screen_   → Do_Mainmenu_Load_Screen_
Newgame_Screen_     → Race_Selection_Screen_
Race_Selection_Screen_ → Racial_Option_Screen_ | Flag_Screen_
Racial_Option_Screen_  → Flag_Screen_
Main_Screen_        → Do_Colony_Screen_ / Get_Ship_Stack_For_Officers_Screen_ / Get_Star_Id_For_Officers_Screen_
Race_Screen_        → Diplomacy_Screen_ / Race_Report_Screen_
Planet_Colonization_In_Main_Screen_ → Colony_Landing_Screen_
Hotseat_Screen_ / Start_Net_Screen_ → Race_Selection_Screen_
```

remake 的新遊戲流程順序與此一致;`Main_Screen_ → Do_Colony_Screen_` 是先前缺的那一層。

### 第一梯 — 已完成(2026-08-06)
1. ~~星系/行星生成表~~ → **已接進 remake**(`gamedata/galaxygen.go`,commit `f8bbcbd`)。
   光譜/大小/氣候/礦產/重力/行星數全部改用原版骰表,並加了分布回歸測試。
2. ~~殖民地經濟表~~ → **已交叉驗證**:`_food_per_farmer_table`、`_minerals_per_mine`、
   `_planet_max_population` 三項與 remake 現值一致,不必改(見 C-2)。
   `_climate_maintenance_modifiers` 仍待確認讀取者。
3. ~~INFO 5 個子畫面~~ → **已實作**(commit `fade3f7`),含 module 122 歷史記錄系統。

### 第二梯 — 下一批
4. ~~射程命中/傷害懲罰~~ → **已完成**。查證發現兩張表 remake 早就有且與原版逐格相同
   (`combatRangeLevelPenaltyTable`、`damageDissipationPenaltyTable`);真正的缺口是
   **傷害衰減從未被呼叫**——同一發雷射在 1 格與 23 格外傷害一樣。已接進
   `ResolveShotWithMods`,順帶讓 NR(No Range Dissipation)改造第一次有實際效果。
5. ~~`_personality_*` 表 + AI 行星估值~~ → **都已完成**(2026-08-06),含後續補上的
   `Proximity_Worth_To_Player_`、`Compute_Contextual_Planet_Values_`、`Colony_Worth_To_Player_`。
   最後一個 `Enemy_Colony_Worth_To_Player_` 是「攻擊目標的**額外**加權」,
   remake 目前用 `AIColonyValue ÷ 距離` 代打(見 `shell/ai_attack.go`)。

   **順帶補上的缺口:AI 宣戰之後真的會打了。** 先前 AI 只會擴張與宣戰,關係掉到 -40、
   態勢寫著「戰爭」,玩家卻毫髮無傷——整局唯一的軍事壓力來自安塔蘭人腳本。現在 AI 會依
   `Colony_Worth_To_Player_` 的估值挑玩家最有價值的殖民地突襲(`shell/ai_attack.go`),
   造成人口/國庫/建築損失;玩家的艦隊、駐軍、軌道防禦建築會擋,擋不住也會消耗攻方戰力。
   ⚠ **「何時打、打贏怎樣」是 remake 的模型**(原版決策函式尚未反編),只有目標估值是原版公式。
   300 回合探針:56 次突襲、最早第 26 回合、經濟未崩(結束 BC 641)。
6. **一星多行星** → 原版每顆恆星 1–5 顆行星各佔一條軌道;remake 的 `Stars`/`Planets`
   索引一一對應是 UI/拓殖/AI 共同的假設,拆開是跨層改造(見 `genPlanets` 註解)。

   **已做的一半**(2026-08-06):恆星系現在會生成完整的軌道天體組成(用 C-5 的原版表),
   代表行星挑「最適合殖民的那一顆」,其餘記在 `Planet.SystemBodies` 供顯示與日後的前哨站。
   `Stars`/`Planets` 的一一對應**沒有動**,所以 UI/拓殖/AI 完全不受影響。
   氣態巨星/小行星帶因此第一次真的出現在遊戲裡,且不能直接殖民(手冊 p.55「colonies can
   only survive on a solid planet」)——探針:960 顆星裡 49 顆不可殖民。
   **剩下的一半**是讓玩家能分別殖民同一系統的多顆行星,那才是跨層改造。
7. ~~獨立 Colony 畫面 + Event 畫面~~ → **兩個都已完成**(2026-08-06)。
8. **地面戰解算**(`Resolve_Ground_Combat_` / `Ground_Combat_Round_`)→ 取代目前沿用一代 1oom 的借用結構。

### 第三梯:補完整性
9. ~~前哨站~~ → **已完成**(2026-08-06,`internal/shell/outpost.go`):前哨船可建造、
   可在氣態巨星/小行星帶/一般行星建立軍事前哨站,前哨站是掃描站(併進 detection.go 的偵測源)
   且**沒有人口與產出**(手冊 p.85「produces nothing」,故不進 PlayerColonies);之後在同一顆星
   建殖民地時前哨站改建為海軍陸戰隊營(手冊逐字)。順帶補上**殖民船也終於可以建造**——
   先前開局送一艘、用掉就再也不能擴張。
   ⚠ 未兌現的一半:「延伸艦艇航程 / 加油站」(手冊 p.119/p.133)——remake 的 SendFleet 沒有
   航程上限這個概念,沒有可套用的機制,不臆造。
   ⚠ 手冊 p.50 的「前哨站升級成可住人殖民地」科技仍未做(需要對應科技旗標)。
10. ~~太空怪獸~~ → **已完成**(2026-08-06,`internal/shell/monster.go` + `gamedata/space_monster.go`)。
    這一項有個值得記的地方:**它一直被程式碼引用著卻不存在**——colonization.go 檔頭抄的手冊
    原文就寫著殖民船要「as long as all space monsters and enemy ships have been cleared from
    that planet's system」,但那個 gate 從來沒有東西可擋。

    - 五種怪獸的名字來自執行檔字串表(0x1F742C 起連續:Guardian / Amoeba / Dragon / Hydra /
      Crystal),對應五個 `Load_*_Ship_Design_` 函式與 `_monster_names` @ 0x199266
    - 傷害數字來自手冊 p.114「Monster Traits」逐字(水晶射線 40-80、電漿吐息必中上限 60、
      相位眼 5-10、龍焰必中上限 300 每格 -15、腐蝕黏液 25-50 每回合 -5)
    - 生成規則來自手冊 p.60 逐字:「a system with a monster will always have another special
      — that's usually what drew the monster there in the first place」,已落地成「擺怪獸時
      強制補一個特殊物產」
    - ⚠ 怪獸的**結構值**與挑選機率是 remake 估值(手冊只給武器傷害);原版的數量是新遊戲
      設定(`_user_wants_n_space_monsters` @ 0x19A006),remake 先用固定密度
    - 順帶依手冊 p.119「Support ships … **do not fight**」把殖民船/前哨船排除在戰鬥火力之外
      ——先前它們會以最低戰力混進戰列
11. ~~持續型隨機事件~~ → **已完成**(2026-08-06,`internal/shell/events_persistent.go`)。
    先前 9 個事件卡在「缺子系統」,真正缺的是同一個東西:remake 只有「單次結算」的事件模型,
    **沒有任何跨回合的事件狀態**。補上那個模型之後,手冊 p.180-181 就直接是規格書:

    | 事件 | 手冊給的數字 |
    |---|---|
    | 超新星(24) | ≥200 回合觸發、倒數 6-14 回合、系統研究點全投入搶救、失敗則全滅 + 行星變輻射 |
    | 時空異象(25) | 星系凍結:不生產不成長,也不吃食物不繳維護費;6 回合後每回合 5% 結束 |
    | 超空間獸(26) | 航行中的艦隊有機率損失一艘;6 回合後每回合 5% 離開 |
    | 蟲洞(28) | 航行中的艦隊「in a single turn」抵達 |
    | 怪獸入侵(19-23) | 變形蟲 ≥100、太空鰻 ≥150、水晶 ≥200、九頭蛇 ≥250、巨龍 ≥300 |

    超新星那條的張力也是手冊寫死的:「if the emperor doesn't accelerate the colony's research
    efforts, the colonies will discover the solution **one turn too late**」——remake 據此把
    需求量設成「系統自然產出 × (倒數+1)」,讓「什麼都不做剛好差一回合」成立。

    ⚠ 太空鰻是**近似**:手冊說牠「never attack colonies or outposts」、只封鎖,且 30 回合後
    會分裂(最多 4 隻)。remake 用盤據型怪獸代打「封鎖」那一半,另兩半未建模,已標在事件表的
    `Needs` 欄。

    **這一批順帶翻出一個過期斷言**:`advanceConquestVictory` 的註解寫著「remake 沒有任何機制
    會讓 PlayerColonies 完全清空,故玩家戰敗這個分支不可達」。超新星讓它可達了,而
    `CheckExtermination`(只剩一方存活)在「玩家死光但還有三個 AI」時回 false——400 回合探針
    實測到玩家 0 殖民地、遊戲卻繼續空轉。已補 `advancePlayerDefeat`(手冊 p.184 計分段明講
    「If an empire is eliminated by a random event」,帝國被隨機事件消滅是原版就有的概念)。
12. ~~Hall of Fame / Hi-Score~~ → **已完成**(2026-08-07,`gamedata/score.go` + `shell/score.go`
    + `cmd/moo2/hiscore.go`)。手冊 p.184 列了八條計分因素但**一個數字都沒給**;原版 module 60
    的一整組 `Get_*_Score_` 函式每個都短到能逐指令讀完,八條的係數全在裡面:

    | 項目 | 公式(反組譯) | 手冊對應的那句話 |
    |---|---|---|
    | 時間/星圖/種族數 | `nPlayers × (20×(星圖大小+1) + 80) − 已過回合數`;人口 0 則整項 0 | 「越快贏分越高」「星圖越大分越高」「種族越多分越高」三句合一 |
    | 人口 | 自己所有殖民地人口總和 | 「total number of population units … added to your score」 |
    | 俘虜人口 | `俘虜 × 2 ÷ (星圖大小 + 1)` | 「premium … **higher in smaller galaxies**」——除數就是那句話 |
    | 科技 | `3 × 已知主題 + 5 × Hyper-Advanced 等級` | 「First level Hyper-Advanced … worth **more** points than normal ones」——5 > 3 |
    | 殲滅種族 | 每族 50 | 「a boost」 |
    | 獵戶座 | 100 | 「a big chunk of points」 |
    | 議會勝利 | 100 | 「a substantial addition」 |
    | 安塔蘭勝利 | 250 | 「the **biggest** point bonus of all」 |

    **相對大小完全對上手冊的形容詞排序**:250 > 100 = 100 > 50。另有兩個順帶的交叉驗證:
    科技分掃的是 `player+0xC4` 起 0x53(83)長的研究主題陣列——0xC4 與
    `Do_System_Discoveries_At_Star_` 讀遠古文物時用的是同一個偏移,83 也與 remake 既有的
    研究主題數相同;時間分用 `word[0x192FD8] − 0x88B8` 算已過回合,0x88B8 = 35000 =
    星曆 3500.0 ×10,正是遊戲起始星曆。

    ⚠ remake 側的落差:①獵戶座系統還沒做,該項恆 0 ②「殲滅種族」原版有逐玩家的
    「這個玩家滅了誰」陣列(player+0x1F2),remake 沒追蹤是誰滅的,目前全算給玩家,
    AI 互滅時會高估——標明,不假裝精確。
13. ~~艙損/維修~~ → **已完成**(2026-08-07,`internal/shell/repair.go`)。這一項的起點不是
    「補一個修復系統」,而是先發現 **remake 根本沒有艦艇損傷這個概念**——一艘船不是完好就是
    被擊沉,打完慘勝的仗倖存艦跟出港時一模一樣。於是「自動修復」這個元件
    (`SpecialOptions` 裡的 `{"自動修復", …, TECH_AUTOMATED_REPAIR_UNIT}`)從加進來那天起
    就沒有任何效果:沒有損傷可修。

    | 規則 | 來源 |
    |---|---|
    | 停在自家據點(殖民地/前哨站)→ **完全修復** | 反組譯 `Repair_Ships_At_Colonies_` @ 0x580F5 直接呼叫 `Repair_Ship_Full_` @ 0x581F3——是完全修復,不是逐回合慢慢修 |
    | 自動修復元件:戰鬥中每回合修 **20%** 結構損傷 | 手冊 p.82 逐字 |
    | 自動修復元件 / 進階損害管制:**戰後完全修復** | 手冊 p.82 / p.80 逐字 |
    | 機械化種族:戰鬥中 10%/回合(常數已備,無呼叫端) | 手冊 p.25 逐字;remake 沒有種族特質欄位可掛 |

    **交叉驗證**:openorion2 讀存檔的 `struct Ship`(`gamestate.h:1268`)把「原版記了幾份損傷」
    逐欄位寫死了——

    ```c
    uint8_t  shieldDamage, driveDamage;   // percent
    uint8_t  computerDamage, crewLevel;
    uint8_t  damagedSpecials[(MAX_SHIP_SPECIALS+7)/8];
    uint16_t armorDamage, structureDamage;
    ```

    原版每艘船記**六份**(護盾/引擎/電腦/逐元件旗標/裝甲/結構),remake 只有
    `structureDamage` 這一份。`ships.cpp:1060` 的 `isSpecialDamaged(i)` 把壞掉的元件名字
    畫成損壞色,那是逐元件損傷唯一的 UI 出口。

    ⚠ 誠實留白:①**內部系統損傷**(引擎/武器/護盾/電腦/各元件的損壞度)——remake 的戰鬥是
    艦級抽象,沒有逐系統狀態,手冊那句「systems damage 10%/5% per round」無處可套,只做結構
    損傷這一半;原版對應的 `Apply_Internal_Damage_` @ 0x35251 是個依傷害類型分十幾條分支、
    操作艦艇結構 +0x29/+0xC2/+0x134 等欄位的大函式,要接它得先有逐系統模型 ②裝甲與結構在
    remake 合一 ③「進階損害管制」科技還不在科技樹裡,`playerHasAdvancedDamageControl` 恆
    false(標明是缺科技,不是規則沒接)。

    **UI 出口**:艦隊列表每艘船後面加「損傷 N%」(輕傷琥珀 / ≥50% 紅),完好的船不畫。
    沒有這一欄的話這個系統對玩家等於不存在——打完仗有傷卻沒地方看得到。
    截圖廊也補了艦隊列表這一張(`11_fleet.png`,共 15 張)。

    順帶記一個踩到的坑:截圖廊原本在 `galleryVictoryTick`(t28)注入損傷,而 t29 按了
    「結束回合」,截出來一艘傷都沒有——那不是顯示壞了,是 `EndTurn → advanceShipRepair`
    正常運作:艦隊開局就停在母星,照 `Repair_Ships_At_Colonies_` 的規則被完全修復。
    注入點改到最後一次結束回合之後(t40)才驗得到。

    順帶修掉一個對映錯誤:損傷寫回原本用外部平行陣列對映「第 k 個參戰艦 → 第 k 艘船」,
    但戰鬥中有人陣亡後這個對映就錯位,會把 A 船的傷記到 B 船上。改成把船索引放進
    `combatant` 結構裡,過濾陣亡者時整個 struct 複製、索引跟著倖存者走。
14. ~~安塔蘭房間~~ → **已完成**(2026-08-07,`cmd/moo2/antaranroom.go`)。原本這條勝利路徑的
    入口是艦隊列表左下角一行文字,點下去直接跳戰鬥結果——中間沒有確認、沒有戰力對比,
    前置條件不滿足時更是「點了完全沒反應」(`CanAssaultAntares` 回 false,而畫面上毫無跡象)。

    美術來源是反組譯 `sub_14C83`:`mov edx, 0 / mov eax, offset aAntaroomLbx / call sub_126B42`
    ——載 `antaroom.LBX` **資產 0**。實際查下去,資產 0 是個小圖但**帶內嵌調色盤**,資產 1 才是
    640×480、55 幀的 delta 動畫(鏡頭推進到安塔蘭統治者面前);用資產 0 的調色盤去解資產 1 才
    出得來,拿 `buffer0` 的色盤解會是一團彩色雜訊。累積成最終畫格的做法沿用外交議事廳
    (`DIPLOMAT#29`,38 幀)那條已驗證的路徑(`lbx.Image.AccumulatedRGBA`)。

    畫面內容:戰力對比(我方 `playerMilitary()` vs 安塔蘭防禦艦隊,同一套數字,不另算一份)、
    發動/撤退兩顆按鈕、以及**擋下時逐條講明卡在哪**(`AssaultAntaresBlockReason`:勝負已定 /
    本局關閉安塔蘭攻擊 / 沒有次元傳送門 / 沒有艦隊)。手冊 p.183 那句
    「not available if you disabled Antaran Attacks」現在玩家看得到了。

    ⚠ 留白:原版把 55 幀當推鏡動畫播,remake 只呈現最終定格(`overlayScreen` 沒有動畫層,
    加上去等於 55 張 640×480 貼圖常駐)。

15. ~~地面戰畫面~~ → **已完成**(2026-08-07,`cmd/moo2/groundcombat.go`)。這一項的價值不在
    「多一個畫面」,而在**版面座標一個都不是估的**——全部從反組譯挖出來:

    | 元素 | 真值 | 來源 |
    |---|---|---|
    | 攻方面板外框 / 暗化區 | 貼圖 (1,40);`Darken_Fill_(2,41,259,184)` | `sub_B8BC7` |
    | 守方面板外框 / 暗化區 | 貼圖 (378,40);`Darken_Fill_(379,41,638,184)` | `sub_B8C8B` |
    | 面板文字 x(置中)| 攻方 130 / 守方 508;列高 11,首列 y=50 | 同上 + `sub_1210FD` = `Print_Centered_` |
    | 兵種欄位 x | `261 / (兵種數+1) × (序號+1) + 基準X` | `Print_Troop_Totals_` @ 0xB896D |
    | 部隊落點 | `X = 基準X + Random_(50) − 20`;`Y = min(360 + Random_(85), 430)` | `sub_B88B2` |
    | 兩側基準 X | 攻方 **50** / 守方 **590** | 常數 `dword_B6CDE = 0x024E0032` |

    **一個關鍵訂正**:原本以為 COLGCBT 資產 21(261×149)是「戰場視窗外框」,照這個假設排版
    做出來的第一版整個是錯的。`Print_Troop_Totals_` 裡的 `mov eax, 105h`(=261)與資產 21 的
    寬度**恰好相同**,才看出它是**兵力統計面板的外框**——戰場是整個 640×480 畫面本身
    (部隊 y 落在 361–430,就是畫面底部那一帶)。兩個獨立來源對上同一個數字,這是內容自明的
    錨點,不是從檔名或尺寸猜的。

    順帶釘死、寫進註解免得重查的事實:每側最多 **40** 個單位、每單位 **25** 位元組記錄
    (`byte_19EB94 + side*0x3E8 + i*0x19`,欄位 +0 X / +2 Y / +5 狀態 / +6 兵種 / +7 側 /
    +0x0A 動畫偏移);**4 種兵種**;兵種小圖示來自 **RACEICON.LBX**(每族 13 個,
    索引 = 種族×13 + 7/8/9/10)而不是 COLGCBT;`Replace_Colgcbt_Color_With_Player_Colors_`
    @ 0xB8EFB 說明 dump 出來士兵腳下那塊洋紅色**不是影子,是帝國旗色的佔位色**。

    ⚠ 誠實留白:①原版是逐幀動畫的即時戰鬥,remake 的 `ResolveGroundBattle` 一次算完,
    沒有逐單位時間軸可驅動動畫 → 呈現戰後定格 ②原版底圖是該殖民地地表,remake 沒有那一層
    ③落點的 `Random_` 換成由單位序號推出的固定散布(範圍與原版一致,但可重現,截圖驗證需要)
    ④原版 4 種兵種,remake 只模型化陸戰隊與戰車營 ⑤**調色盤未定案**:COLGCBT 所有資產都沒有
    內嵌調色盤(原版沿用當時殖民地畫面的),remake 借 COLBLDG.LBX#0,渲染合理但未證實
    ⑥文字列高用 12 而非原版的 11——中文字比原版單位元組字型高,11 會上下相黏。**這是整個
    版面唯一沒照抄原版數字的地方。**

16. ~~載入遊戲視窗~~ → **已完成**(2026-08-07,`cmd/moo2/loadgame.go` + `internal/shell/saveslots.go`)。
    `LOADSAVE.LBX` 先前全 repo 零引用不是巧合——remake 根本只有**一個**存檔檔案,每回合覆寫,
    主選單的 Continue 與 Load Game 都是「讀那一個檔」,沒存檔時點下去靜默無反應。

    | 元素 | 真值 | 來源 |
    |---|---|---|
    | 背景 / LOAD 鈕 / CANCEL 鈕 | `game.lbx` 資產 20 / 21 / 22,調色盤取 `mainmenu.lbx` 資產 21 | openorion2 `LoadGameWindow` |
    | 視窗定位 | `x=(640−寬)/2`、`y=(480−高)/2` | 同上 |
    | LOAD 鈕 / CANCEL 鈕 | (37,337) / (171,338) 68×22 | `initWidgets` |
    | 存檔槽 | 10 格,第 i 格 (22, 22+31×i) 232×27;文字 (x+32, y+24+31×i) | `initWidgets` + `drawSlot` |
    | 槽位規格 | `SAVEGAME_SLOTS = 10`,**最後一格固定是自動存檔** | `mainmenu.h` + `drawSlot` |

    順帶了結兩件事:
    - **oracle issue #2 結案**:原版無存檔時 Continue / Load Game 是**灰階不可按**的
      (2026-07-12 archive.org 對照的結論)。remake 先前是「可按但靜默無反應」,玩家會以為壞了。
      現在無存檔就不給熱區 + 標籤畫成暗綠。
    - **主選單「名人堂」接錯的入口修正**:它先前被暫借給「研究選擇」畫面當調色盤鏈的示範入口,
      現在導向真正的最終得分畫面(`hiScore`,2026-08-07 已建)。

    途中發現的一個真 bug:**`PlayerName` 與 `FlagColor` 從來沒有進過存檔**——玩家在命名旗色
    畫面取的帝國名和選的旗色,一讀檔就變回預設值。做存檔槽列表要顯示帝國名時才撞到,已補進
    `sessionSnapshot`(舊存檔解出零值,退回預設,不會壞)。

    ⚠ 誠實留白:①**只有「載入」沒有「儲存」**——原版另有 `Do_Save_Game_Popup_` @ 0x7E154,
    remake 目前只在每回合結束時自動存檔,玩家無法主動存進指定格 ②~~原版視窗右側會依存檔類型
    畫單人/熱座/網路/數據機四種圖示~~ → **2026-08-07 已補**:`game.lbx` 資產 23-26 逐張確認是
    「1 人 / 3 人 / 2 人+螢幕 / M」,依存檔是不是熱座局選 23 或 24,位置 `_x+206, y+12`
    (openorion2 `drawSlot`)。同一輪順手把槽內文字改成原版的兩行版面(第一行存檔名、
    第二行星曆 `+32,+14` 與存檔時間 `+122,+14`)——先前是「名稱 + 星曆」擠在同一行。
    日期用兩位數年份,因為 x+122 到 x+206 之間只有 84px,原版的 C `%x` 在該 locale 同樣是
    八字元日期 ③存檔用 remake 自己的 JSON 格式,不是原版 `.GAM`
    (原版格式由 `internal/save` **唯讀**解析,寫回不在範圍內)。

17. ~~儲存遊戲視窗 + 遊戲中的「遊戲」選單~~ → **已完成**(2026-08-07,
    `cmd/moo2/gamemenu.go`;儲存視窗併進 `loadgame.go`)。

    **存檔與載入是同一個視窗**——這不是 remake 的簡化,是原版的結構:繪製函式只有一個,
    叫 `_Draw_Load_Save_Game_Popup_` @ 0x7F206,名字自己就說完了;差別只在說明列
    (`Set_Load_Game_Screen_Help_List_` @ 0x6F850 vs `Set_Save_Game_Screen_Help_List_` @ 0x6F865)
    與動作鈕的字。

    順帶把 openorion2 的資產索引拿反組譯核了一次:`Load_Mainmenu_Load_Game_Popup_` @ 0x803D9
    依序載 GAME.LBX 資產 **0x14–0x1A(20–26)**,正是 openorion2 的
    `ASSET_LOAD_BACKGROUND`…`ASSET_LOAD_MODEM` 那一串。兩個獨立來源互相印證。

    遊戲選單視窗(`GameMenuWindow`)的座標:視窗 **(144, 25)** ——硬編不是置中;
    SAVE (40,43) / LOAD (147,43) / NEW (40,88) / QUIT (147,88) / SETTINGS (40,307) /
    RETURN (151,307),精靈為 `game.lbx` 資產 1–6,背景資產 0,調色盤取 `buffer0.lbx` 資產 0。
    星系畫面上那顆「遊戲」鈕本身在 **(249, 5)** ——它先前**畫得出來但點了沒事**,
    等於玩家在遊戲中沒有任何辦法主動存檔。

    ⚠ 誠實留白:原版這個視窗中段有 Music / Sound Fx 兩條音量滑桿(背景圖上畫得很清楚),
    remake 的音訊層沒有音量控制介面,滑桿不畫也不接——畫一條拖不動的滑桿比沒有更糟;
    SETTINGS 同理,原版另有一整個設定畫面,remake 尚無對應內容,按鈕保留但按了只回訊息。

18. ~~`Colony_Bombing` 畫面~~ → **已完成**(2026-08-07,`cmd/moo2/bombing.go`)。

    反組譯挖到的:
    - `Draw_Colony_Bombing_Screen_` @ 0xB4800 標題 `Print_Centered_(319, 10)`——與地面戰
      同一個錨點,兩個畫面共用版面慣例。
    - `Do_Bomb_` @ 0xB4606:炸彈記錄**每筆 15 位元組**(+0 精靈 / +4 X / +6 Y / +8 幀 / +0x0E 啟用),
      每 tick 已啟用者幀 +1,**未啟用者 `Random_(5) == 1` 才啟用**——所以原版的炸彈是零零星星
      散著炸,不是整排同時落地。
    - `Add_Bomb_To_End_Of_Queue_` @ 0xB43A6:在 **49** 個目標槽上蓄水池抽樣挑下一個挨炸目標。
    - 背景是 `COLONY.LBX#8`(640×480,6 幀 delta)——殖民地的**建築格地面**,呼應那 49 個槽。

    **一個順帶確認的推論**:格線用 COLBLDG#0 上色後是一條紫色階(#6C306C/#8C488C/#B05CB0)。
    那不是美術本來的顏色,是**帝國旗色的佔位色**——原版有
    `Replace_Colgcbt_Color_With_Player_Colors_` @ 0xB8EFB 專門做這件事。remake 照做了
    (`recolorPlayerRamp`),格線會跟著玩家選的旗色變。這同時解釋了先前 dump 地面戰士兵時
    腳下那塊洋紅色是什麼。

    ⚠ 一個資料層的事實,記下來免得後人重查:`Load_Bombing_Anims_` @ 0xB435A 載
    **COLONY.LBX 資產 1**,但**這份遊戲資料裡該資產是 0 位元組**(資產 0–4 全部長度為 0,實測)。
    炸彈動畫的圖抽不出來,remake 用自繪爆點代替。

    ⚠ 其餘留白:①remake 的殖民地不是**空間格**模型(只有職務人數 + 建築集合),沒有「第 k 座
    建築在格子哪個位置」,所以彈著點是散在格面上而非打在真正的建築格 ②原版是逐幀動畫,
    remake 呈現戰後定格。

    另外訂正了一個版面推論:原版的 `Print_Centered_` 第二個引數是文字的**上緣**不是中心
    ——當中心的話 y=10 會被畫面上緣切掉,原版不可能這樣排。地面戰畫面的標題一併修正。

19. ~~Smacker 過場~~ → **已完成**(2026-08-07,`internal/smk/` + `cmd/moo2/cutscene.go`)。

    先釐清前提:MOO2 的片頭與各結局過場**不是 LBX**,是**裸的 Smacker 檔**,只是沿用了
    `.LBX` 副檔名——`INTRO.LBX` 開頭四個位元組就是 `SMK2`(480×160、1407 幀、≈13 fps),
    `WININFIN` / `LOSERFIN` / `ORIONFIN` / `ANATKFIN` / `AMEBAFIN` / `PLNTDFIN` / `DIMTVFIN` /
    `ANWINFIN` 同理。所以這一項要的是一個 SMK2/SMK4 解碼器,不是找檔案。

    **驗收**(INTRO / ORIONFIN / WININFIN 三個檔都成立):
    - 幀資料全部吃完後檔案**殘餘 0 位元組**——幀邊界錯一個位元組這數字就不會是 0。
    - 四棵 Huffman 樹的節點數都在標頭上界 `((size+3)>>2)+4` 之內。
    - 全片 1407 幀解得完,**沒有任何一次位元流走出樹外**。
    - **位元預算幾乎完全吻合**:1407 幀合計只多讀 721 bits(平均每幀 0.5 bit,就是最後一個
      code 跨過位元組邊界的填補)。樹的形狀固定,消耗的位元數只由「讀了哪些 code」決定
      ——預算對得上就代表 code 讀對了。
    - 第 1050 幀(經過一千多幀差分累積)畫面依然乾淨連貫。值解錯的解碼器會愈解愈爛。

    **真正的 bug**:編碼器在「剩餘區塊全部與上一幀相同」時就**停筆**,不會把 SKIP 一路寫到
    最後一格。原本照著「解滿 blocks 個區塊」跑,位元流用完後繼續從補零的位元讀,在畫面底部
    畫出垃圾。改成位元流用完就停之後,超讀從 **1,289,836 bits(162 幀)降到 721 bits(38 幀)**。

    ⚠ **一個誤判,記下來免得重犯**:中段幾幀(118–124,全片動作最劇烈的段落)方塊感很重,
    一開始被當成解碼錯誤查了很久。那是**原始素材本身**——1996 年 480×160、256 色、低位元率的
    Smacker,高動態段落本來就這樣。判別方法是量位元預算,不是看畫面。

    **已測過、確定不是原因的**(負面結果):escape 位元組序反轉(超讀 162→701 幀)、
    MRU 拿掉 `v != last[0]` 判斷(162→210)、對 SMK2 也讀 SMK4 的全彩模式位元(162→517),
    三個都更糟。

    接線:正常互動模式開場播片頭(整數倍放大置中、點任意處跳過、載不動就直接進主選單);
    headless 驗證與截圖廊直接進主選單——那些腳本是從主選單第一拍開始數 tick 的。

    **結局過場也接了**(2026-08-07,`internal/gamedata/cutscene.go`)。這一步的重點是
    **對映不能照檔名猜**——專案紀律明訂一手資料勝過檔名推論。三條獨立證據:

    ① **反組譯**(唯一的直接證據,但只涵蓋三個檔):執行檔字串表裡只有三個 `*FIN` 名字,
       而且呼叫端明白告訴我們**它們根本不是結局**——
       `AMEBAFIN` 與 `PLNTDFIN` ← `Bomb_Results_Popups_` @ 0xE85F7 +
       `Do_Attacker_Beat_Colony_Stuff_` @ 0xE87D2;`DIMTVFIN` ← `Tactical_Combat_` @ 0x47939。
       其餘六個名字執行檔**完全沒有字面引用**,存在 `ESTRINGS.LBX` 的字串池裡
       (按字母排序的資產名池,前後文認不出語意)。

    ② **尺寸分群**:九個檔剛好分成兩群——480×160(INTRO / WININFIN / GENWINFN / LOSERFIN /
       ANWINFIN)與 288×208(AMEBAFIN / PLNTDFIN / DIMTVFIN / ORIONFIN / ANATKFIN)。
       ①已證實的三個遊戲內事件動畫**全都是 288×208**,片頭則是 480×160。
       所以 **ORIONFIN 與 ANATKFIN 跟著事件那群走,不是結局**——光看檔名會全部猜錯。

    ③ **末幀內容**(解出來直接看):GENWINFN 收在**遊戲標題 logo**、LOSERFIN 收在**燃燒中的
       廢墟城市**、ORIONFIN 收在**行星表面的城市**(與 288×208 事件群一致)。

    接線:勝負分出 → 依 `Victory` 選過場(敗北 LOSERFIN / 安塔蘭勝 ANWINFIN / 其餘 GENWINFN)
    → 播完或點擊 → 最終得分。過場載不動就直接跳到最終得分,不擋結算。

    ⚠ **仍未定**:`WININFIN` 與 `GENWINFN` 都在結局群,但「哪一個對應哪一種勝利」沒有證據
    ——挑選它們的程式碼不在執行檔的字面引用裡。remake 一律用 `GENWINFN`(已由末幀確認是
    完整結局片),`WININFIN` 標為待定,不臆測。`ORIONFIN` / `ANATKFIN` 等事件動畫也還沒接
    (前者需要獵戶座星系,remake 還沒有)。清單見 `gamedata.UnmappedCutscenes`。

    ⚠ 誠實留白:**只有畫面沒有聲音**。Smacker 的音軌是壓縮的(片頭第 0 軌 23748 bytes、
    11025 Hz),解碼器跳過音訊區塊——那需要再實作一組 Smacker 音訊 Huffman,與畫面是兩套
    獨立的編碼。

20. ~~多人連線(獨立子專案)~~ → **熱座已完成**(2026-08-07,`internal/shell/hotseat.go` +
    `cmd/moo2/multiplayer.go` + `cmd/moo2/hotseat.go`);**網路 / 數據機 / 序列埠直連未做**。

    **原版的多人設定畫面整張版面都拿到了**(`Multi_Player_Screen_` @ 0xF4D99,初始化
    `sub_F42CA`、建 widget `sub_F009A`):背景 MULTIGM.LBX#0(640×480)、面板 #1(482×335,
    置中 `(0x280−w)/2` / `(0x1E0−h)/2` → (79,72));左欄四個連線方式 x +0x3B,
    y +0x5B/+0x7A/+0x9B/+0xBB;右欄四個動作 x +0x10D、同四列;CANCEL (+0xB0, +0x11E)。
    四顆按鈕的英文逐張 dump 確認是 NETWORK / MODEM / NULL MODEM / HOTSEAT,與
    `Set_Multi_Player_Game_Type_` @ 0xF5691 寫進 `byte_199F3A` 的四個模式碼
    (0=單人 1=熱座 2=網路 3=數據機)對得上。

    **連原版自己會隱藏的按鈕都照做**:`sub_F009A` 在選了 HOTSEAT 時把 JOIN GAME 的
    widget id 設成 `0FC18h`(無效值)——原版熱座模式下就沒有那顆鈕;COMM INFO 同理只在
    modem / null modem 建。

    **熱座的席位模型**:原版帝國資料本來就是 `player[i]` 陣列(stride 0xEA9)+ 當前索引
    `word_19999C`,所以 `Save_Hotseat_Map_Info_` @ 0x88F5D **每席只存七個 word**(星圖視野)。
    `Get_Multi_Player_N_Humans_` @ 0x121F0 則是去數 `player[i]` 裡控制碼為 100 的帝國
    ——「幾個真人」不是獨立設定,是「有幾個帝國被標成真人」。remake 的 `GameSession` 是單數
    欄位不是陣列,改成 `player[i]` 要動幾乎每個畫面,故走**席位交換**(`internal/shell/hotseat.go`):
    玩家側欄位整組進 `seat`,換人時存回目前席位、載入下一席。語意與 `player[current]` 等價。

    ⚠ 誠實留白(全部寫在 `hotseat.go` 檔頭,此處摘要):
    ①**沒有網路層**,NETWORK / MODEM / NULL MODEM 三顆在畫面上是灰的,點下去說明未實作;
    ②交接畫面用原版的**尺寸與文字錨點**(`Draw_Hotseat_Screen_` @ 0x626D6:視窗置中、
    文字 +0x0E/+0x46)但**底圖是自繪的**——原版底圖來自 `dword_19B874` 指向的已載入影像,
    那個全域在別處填,追不到具體 LBX 資產,故不宣稱像素對齊;
    ③非當前席位的帝國在 `EndTurn` 最後才結算(當前席位在 AI 決策之前、其餘在之後),
    差一個 AI 回合的資訊;④勝負判定與歷史快照只對當前席位跑,其餘席位打進安塔蘭母星
    或全滅不會結束對局——要補得先讓勝負判定吃「哪一位玩家」而不是隱含的 `s.Player`;
    ⑤真人席位是從 AI 對手**接管**過來的,而 `AIOpponent` 比玩家側薄(沒有建造佇列、領袖、
    間諜、前哨站),接手的真人是「有母星、有殖民地、有艦隊,但還沒開始蓋東西」的狀態。

21. ~~NEW GAME 設定畫面用估計座標、只有 3 個設定能選~~ → **2026-08-07 全部重挖**
    (`cmd/moo2/interactive.go` 的 `ngSettings` 檔頭有完整對照)。

    這是 HANDOFF 那條「openorion2 沒有的畫面就去反組譯挖立即數」的第二個實例
    (第一個是地面戰畫面)。三個獨立來源互相印證,沒有一個估的數:

    | 來源 | 給了什麼 |
    |---|---|
    | `sub_CCE2E`(建 widget) | 畫面原點 `word_1831D4`=X=15、`word_1831D6`=Y=5;五個設定框、五條數值列、三個開關、兩顆鈕的座標全是立即數 |
    | `sub_CCC3D`(畫值圖) | 3 欄 × 2 列迴圈:起點 (X+0x79, Y+0x77)、欄距 0x9B、列距 0x8C、右下格跳過 → 算出的欄 x = 121/276/431、列 y = 119/259,**與建 widget 的熱區 x1/y1 逐一相同** |
    | `NEWGAME.LBX` | 資產 1–22 剛好 22 張 65×65,而五個選擇器的選項數 3+5+4+7+3 = 22 |

    **修掉一個真的還原錯誤**:左下那個框在原版是 **PLAYERS**(帝國總數 2–8;變數
    `word_1A1366` 由 `byte_199CB1 − 2` 得來,配 `NEWGAME.LBX` 13–19 那七張數字圖
    「2 3 4 5 6 7 8」),remake 先前把它當 RACE 用,對手數則寫死 3。現在接上了,
    帝國總數同時決定熱座席位上限(席位是從 AI 帝國接管的)。

    **另外兩個先前選不到的設定也接上了**:
    - **GALAXY AGE**:效果(光譜加權 `StarClassWeights` + 氣候骰表)其實早就實作完,
      只是被一個常數寫死成 Average,UI 完全沒有這一欄。
    - **難度補到五級**:選擇器選項數是 5、`NEWGAME.LBX` 4–8 是五張手勢圖
      (伸手扶持 → … → 雙拳相抵),remake 先前只有四級,少的是最低的「教學」。
      ⚠ 補在索引 0,舊存檔的難度索引整體位移一格(會變簡單一級,不會壞)。

    **3 級 vs 4 級的矛盾也解掉了**:patch 1.5 手冊寫 TECH LEVEL 有四級
    (Pre-warp / Avg / Post-warp / Advanced),但反組譯的選擇器只有 3 個選項。
    1.5 的 CHANGELOG 第 1463 行寫著「Added pictures for cluster, random and post warp
    in newgame.lbx」——**Post-warp 是 1.5 才加的**,1.3 的 LBX 裡就是沒有第四張圖。
    兩邊沒有矛盾,是版本差異。

    ⚠ 誠實留白:TECH LEVEL **只存設定、不影響 gameplay**。手冊已經給出可用的硬證
    (初始建築數上限 Pre-warp 3 / Avg·Post-warp 5 / Advanced 9;開局已知科技領域數
    Pre-warp 2、其餘 6),但 remake 現在開局固定 2 個領域 = 等同 Pre-warp,
    要補得先查出 Average 該有的那 6 個領域是哪些——手冊只說預設第一個是 field #29,
    沒有一手表之前不臆造。細節記在 `shell.TechLevels` 註解。

22. ~~殖民地畫面「沒有原版版面資料」~~ → **2026-08-07 翻案 + 換上原版框架**
    (`cmd/moo2/colonyscreen.go` 檔頭有完整對照)。

    `colonyscreen.go` 原本寫著「remake 沒有原版 COLONY.LBX 的版面資料(repo 不含版權資產),
    故疊在 colsum.lbx 上自繪」。**兩句都不對**,而且第二句是會擋死後續工作的錯:

    - **版面資料在執行檔裡,不在 LBX 裡。**`Add_Job_Field_For_` @ 0xBCB4B(由
      `Add_Job_Fields_` @ 0xBCC3D 迴圈 i=0..2 呼叫)直接給出職業欄座標:
      `ecx = 0x136`(310)、`push 0x1FE`(510)→ x 310..510;
      `ebx = i*0x1E` 再 `+= 0x3E` → 第 i 列 y = 62 + 30i。
    - **框架美術也在,只是不在 COLONY.LBX。**是 **COLPUPS.LBX 資產 5**(640×480,
      中段透明)。COLONY.LBX#6(73 幀 delta)累積出來是**道路動畫**,不是底圖——
      這一條是實際 accumulate 出來看過才確定的,不是從檔名猜的。

    量框架圖的深色內凹面板得到:左 x 7..115 y 28..148 / 中 x 126..304 y 31..140 /
    **右 x 308..511 y 31..146**——右面板與反組譯的職業欄 310..510 **逐像素吻合**,
    三列 ×30 正好塞滿面板高度。第三個獨立來源。

    其餘量自同一張圖:上方資訊列 y 15..158、中段(行星表面)y 159..423、
    右下 LEADERS y 424..449 / RETURN y 456..479(x 551..637)、
    CHANGE (519,123,61,20) / BUY (588,123,40,20)。

    現況:畫面改用原版框架,職業三列在反組譯真值座標,LEADERS / RETURN 在原版位置。
    CHANGE(換目前建造項)與 BUY(花 BC 立即完工)remake 都還沒有對應功能,
    照主選單 Continue / Load Game 無存檔時的既有做法**畫成灰的 + 不給熱區**,
    不留英文也不給點了沒反應的中文鈕。

    ⚠ 誠實留白:中段(y 159..423)原版畫的是**行星表面 + 建築 sprite 依格點擺放**
    (`Make_Bldg_Array_For_Colony_` / `Bldg_Coords_To_Screen_Coord_` /
    `Sort_Bldg_Array_Columns_` / `Box_Bldg_Slot_` 那一整套,配 COLBLDG.LBX 的建築圖)。
    remake 還沒有那個子系統,那塊放的是自己的建造佇列與可建清單——**那是 remake 的版面,
    不是原版的**。原版的建造佇列本身是獨立彈出視窗(`Build_Queue_Popup_` @ 0xB4041,
    `Add_Build_Queue_Fields_` @ 0xB325A 給出 7 格:x 207..458、y 329+20i),座標已到手,
    等中段的行星表面子系統落地再一起做。

23. ~~種族選擇畫面「版面為合成近似」~~ → **2026-08-07 改用反組譯真值,順便修掉左右相反**
    (`cmd/moo2/raceselect.go` 檔頭有完整對照)。

    remake 原本的擺法是**左文字清單 + 右肖像**,而原版是**左肖像 + 右邊 2 欄 × 7 列的
    種族按鈕網格**——左右完全相反。三個獨立來源互證:

    | 來源 | 給了什麼 |
    |---|---|
    | `Race_Selection_Screen_` @ 0x5C510 建鈕迴圈(i=0..13) | y = `0x5A + 0x30×(i mod 7)` = 90 + 48i';x = `0x15F + 0x7E×(i div 7)` = 351 / 477 |
    | `Draw_Race_Selection_Screen_` @ 0x5BD97 | 肖像 = RACESEL 資產 `15 + 索引`(`lea edx,[eax+0Fh]`),畫在 `(0x36, 0x3F)` = (54, 63);標題橫幅資產 33 @ (366, 52) |
    | `RACESEL.LBX` 資產表 | 1–14 各 **123×45** 2 幀 → 與步距 126×48 差 3px 間隙,正是按鈕;15–28 各 **290×322** → 畫在 (54,63) 佔 54..344,與按鈕欄 x=351 剛好不重疊 |

    順帶把一個「靠讀圖得到」的結論升級成硬證:remake 註解原本寫「肖像↔種族對應**經讀圖
    確認**為字母序」——結論是對的,現在有 `lea edx, [eax+0Fh]` 這行直接證明。

    **原版這個畫面沒有 ACCEPT 鈕**:建的 widget 只有 14 顆種族鈕 + 一個 ESC 處理器
    (`sub_114C72("\x1B")`)。點種族即確認。remake 照做了(滑過看肖像、點下去確認),
    ⚠ 但這一改讓截圖廊的 t7「點 (540,451) 的接受鈕」什麼都不會命中,腳本會卡在種族畫面、
    後面每一張都截錯——已改成點「人類」那顆鈕的座標。**改版面時要一併檢查導覽腳本。**

    ⚠ 誠實留白:①「取消」鈕是 remake 自己加的(原版只綁 ESC,remake 沒有鍵盤路徑);
    ②肖像下方那條族名 + 能力描述也是 remake 的排版——原版這個畫面上沒有能力說明文字
    (`sub_103915(60,175,寬280)` 那個文字帶只在「該族已被別的玩家選走」的多人分支才畫)。

    像素對齊的待重挖名單只剩 **shipDesign**。

24. ~~艦艇設計畫面用估計座標~~ → **2026-08-07 六個艦體槽 + 底部三鈕改用反組譯真值**
    (`cmd/moo2/interactive.go` 的 `dsHullY` / `Design_Screen_` 檔頭區塊)。

    `sub_6C8F9`(由 `Add_Design_Buttons_` @ 0x69E62 對 i=0..5 呼叫)是一張乾淨的 switch 表:

    | | x | y1 | y2 | 高 |
    |---|---|---|---|---|
    | 巡防艦 | 118..227 | 54 | 69 | 15 |
    | 驅逐艦 | 同 | 70 | 84 | 14 |
    | 巡洋艦 | 同 | 85 | 102 | 17 |
    | 戰艦 | 同 | 103 | 117 | 14 |
    | 泰坦 | 同 | 118 | 132 | 14 |
    | 末日之星 | 同 | 133 | 149 | 16 |

    **原版這六格不等距**(x 一律 `0x76..0xE3`,寫在 switch 的 default 分支)。remake 先前
    照 17px 等距排,越往下偏得越多;右邊那欄艦體價格同樣照等距排,最後一格會壓到下面的
    總價那行。兩處都改成跟著 `dsHullY` 走。

    底部三顆鈕(`sub_1151B0`,引數是三個熱鍵字串 `aLb` / `+2` / `+4`):
    **(374, 443) / (461, 443) / (547, 443)**,先前估的差 6–11px。

    ⚠ 座標已到手但**尚未套用**(等確認那些格在原版顯示什麼):
    - 已裝元件清單列:x 55..68、y = 169 + 13i(`imul eax, esi, 0Dh` / `add eax, 0A9h`)
    - 右上兩個資訊面板:(437..627, 56..95) 與 (437..627, 97..123)。remake 目前把元件
      選擇列排在 x 300..600,與這兩格位置不同——要對齊得先追到它們的繪製端。

    這個畫面**先前從沒被截圖廊拍過**(要從艦隊列表點進去,而腳本沒走那一步),與 NEW GAME
    同一個盲點,已補上 `25_shipdesign.png`。

### 像素對齊重挖進度(用反組譯挖立即數這條路)

| 畫面 | 狀態 |
|---|---|
| 地面戰 | ✅ 2026-08-07(第一個實例,見第 15 項)|
| 軌道轟炸 | ✅ 2026-08-07(第 18 項)|
| NEW GAME 設定 | ✅ 2026-08-07(第 21 項,順便修掉 PLAYERS 被當 RACE 用)|
| 殖民地 | ✅ 2026-08-07(第 22 項,框架 = COLPUPS.LBX#5)|
| 種族選擇 | ✅ 2026-08-07(第 23 項,順便修掉左右擺反)|
| 艦艇設計 | ✅ 2026-08-07(第 24 項,六格不等距;右上兩個面板待追)|
| 多人設定 / 熱座交接 | ✅ 2026-08-07(第 20 項)|
| menu / planets / research / fleet / officer / info | ✅ 2026-07-12(openorion2 `initWidgets` 真值)|

**HANDOFF/HONEST-STATUS 上「colony/races/newgame/shipDesign 仍待重挖」那份名單已清空。**
剩下的不是「座標沒挖」,而是「子系統沒做」——殖民地的行星表面 + 建築 sprite 擺放、
艦艇設計右上那兩格的內容、戰機/航母戰鬥子模型。

> 行星表面那一項的**幾何已經到手**(第 27 項:7×7 角點表抽出來並對 COLONY.LBX#8 驗過),
> 卡的改成「建築型別編號對照」與「地表底圖來源」,不再是座標問題。

25. **TECH LEVEL 接上第一個 gameplay 效果:曲速前的 FTL 限制**(2026-08-07)。

    第 21 項留下的「TECH LEVEL 只存設定、不影響 gameplay」不再完全成立。手冊直引
    (已收在 `docs/tech/homeworld-init.md`):

    > "Every race has one colony — their home star system. Exploring outside that system is
    >  impossible until faster than light (FTL) technologies are discovered."

    接法:`FleetHasFTL()` — 只有**曲速前**(TechLevel 0)受限,且限制在研究完
    `FTLTopic` 後解除。`FTLTopic = TOPIC_NUCLEAR_FISSION`,因為 remake 科技樹裡
    `TECH_NUCLEAR_DRIVE`(MOO2 的入門 FTL 引擎)就在那一列(techtree.go 第 55 列,
    Cost 50、ResearchAll),**不在開局就給的 `TOPIC_STARTING_TECH` 裡**。
    一般 / 先進開局手冊描述本來就有星際艦,不受限。

    UI:派遣被擋下時在星圖顯示原因,不是「點了沒反應」。

    ⚠ **踩到一個真的零值陷阱,值得記下來**:`TechLevel` 的 Go 零值是 0 = 曲速前。
    接上限制的當下 `TestFleetInterstellarMovement` 立刻紅燈——`NewDemoSession` 從沒設過
    這一欄。如果沒有測試護著,**所有舊存檔讀回來都會變成曲速前、艦隊整個凍住**。
    修法與 `GalaxyAgeSet` 同款:加 `TechLevelSet` 標記,未設過一律退回「一般」,
    並補一條專門盯這件事的回歸測試(`TestTechLevelZeroValueDoesNotFreezeFleet`)。
    **加「零值有意義」的列舉欄位時,一律配一個 Set 標記。**

    ⚠ 仍未接的 TECH LEVEL 效果(手冊有硬證但缺表):
    - 初始建築數上限 Pre-warp 3 / Avg·Post-warp 5 / Advanced 9,且數量 = min(⌈⅔ 人口⌉, 上限)。
      remake 的母星建築是固定一組(2 棟),低於所有上限,只加上限不會有任何效果——
      要先有「依人口 + `initial_buildings` 優先序生成」的機制,而那張優先序表還沒到手。
    - 開局已知科技領域數:Pre-warp 2 個、其餘 6 個。remake 現在固定 2 個 = 等同 Pre-warp,
      所以選「一般」其實還少了 4 個領域。要補得先查出那 6 個是哪些(手冊只說預設第一個
      是 field #29)。

26. **行星表面 + 建築 sprite 擺放:先查清楚為什麼不能照上面那套方法做**(2026-08-07 調查)。

    這是殖民地畫面中段仍留白的那塊。追下去發現它**不是幾個立即數**,而是查表:

    - `CR_To_XY_` @ 0xBC5D8:`word_182C9C[56a + 8b + 4c]` — 一張四維 word 表
      (a 步距 56 位元組、b 步距 8、c 步距 4)。
    - `Bldg_Coords_To_Centered_Screen_Coord_` @ 0xBC866:`dword_182E24[...]` 與
      `dword_182E2C[...]` 兩張表相加再除二。

    也就是說格點→螢幕座標是**烘在資料段的幾何表**,沒有公式可抄。要做這一塊得先把那幾張表
    從執行檔的資料段抽出來、驗證維度與語意,再配 `Make_Bldg_Array_For_Colony_` /
    `Sort_Bldg_Array_Columns_` / `Box_Bldg_Slot_` 的排序與命中邏輯,以及 COLBLDG.LBX 的
    建築圖對應。**那是獨立工程,不是再挖幾個立即數就能收掉的**——這一輪不硬做,
    把結論記在這裡,免得下次又從頭找一遍才發現同一件事。

27. **行星表面格點:表抽出來了,第 26 項的「獨立工程」判斷只對了一半**(2026-08-07 續)。

    第 26 項的結論是「格點→螢幕座標是烘在資料段的幾何表,沒有公式可抄」——**這句是對的**。
    但它接著把整件事記成阻塞,那一步錯了:**表就在反組譯的資料段裡,讀出來就是了**。
    發現「這是查表不是公式」時,下一步是去讀那張表,不是把它列為待辦。

    ### 抽出來的東西

    | 來源 | 內容 |
    |---|---|
    | `word_182C9C`(392 位元組)| **7×7 角點**,每格 8 位元組 =(x, y)。索引 `a×56 + b×8 + c×4`,c=0 取 x、c=1 取 y |
    | `dword_182E24` | `word_182C9C` 的**完整副本**(執行檔裡放了兩份,內容逐值相同)|
    | `dword_182E2C` | = `dword_182E24 + 8`,差一個 b 槽——所以中心點公式是 (角點[a][b] + 角點[a+1][b+1]) / 2 |
    | `word_182FAC` | 同一組角點去掉第一列第一欄的 6×6 版(步距 48)。有 7×7 就推得出來,remake 用不到 |
    | `dword_BA784`(72 位元組)| `Add_Bldg_Fields_` 走訪 36 格的順序,36 組 (a,b),**a+b 由大到小 = 遠→近**(畫家演算法)|

    格點排出來是個透視菱形:(0,0) 在近端 y=492(超出畫布下緣)、(6,6) 在遠端 y=288,
    左右兩角 x=−21 / 666 也超出畫布——原版本來就是這樣裁的。

    ### 獨立驗證(這步別跳過)

    `COLONY.LBX` 資產 8 是**已經畫好位置**的單格高亮菱形(640×480 稀疏圖)。渲染出來量到的
    四角是上 (316,430) / 左 (225,461) / 右 (406,461) / 下(裁掉)——與表算出的格 (0,0)
    四角**逐點吻合**。表是從程式碼抽的、菱形是從美術檔渲染的,兩個獨立來源對上才算驗過。
    回歸測試釘在 `cmd/moo2/colonysurface_test.go`。

    ### 順帶推翻的版面斷言

    先前寫「中段(y 159..423)原版畫行星表面」——**行星表面不是關在那個框裡的**。
    `Draw_Colony_Screen_` 一開場就是 `C_Anims(1, 0, 639, 479)` + `Draw(0,0,…)`:地表是
    **整個 640×480 的底圖**,資訊面板疊在它上面(所以 COLPUPS#5 的面板是不透明的)。
    格點 y 288..492 / x −21..666 也印證這點。

    ### 建築 sprite:比想像的簡單,不需要座標

    `BLDG0..BLDG4.LBX` 每個資產都是 **640×480、已經畫好位置**的稀疏圖。實測 BLDG0 資產
    0..35 的墨點依序沿格點推進、資產 36 起回到近端 → 每 36 個一循環:

        資產編號 = 建築型別 × 36 + 格號

    畫的時候直接貼 (0,0) 就對位。`Draw_Building_With_Bottom_Centered_` 那套底邊置中是原版
    **產生**這些圖時用的,不是執行期需要的。資產數 360/360/360/360/324 → **49 種建築**。

    同理:`COLROADS.LBX`(156 個 640×480)是道路、`COLVEGGI.LBX`(104 個小圖)是植被,
    對應 `Draw_Road_List` 與 `Build_Veggie_List_Based_On_Bldg_List_`。

    ### 建築編號 → 圖檔:`Cache_Load_Bldg_` @ 0xAF6DC 把整條算式寫死了

        dec ebx                     ; 建築編號是 1-based(0 = 空格)
        idiv 10 → eax               ; 檔號 = (id−1) / 10          → BLDG{0..4}.LBX
        E_Strings_(0C9h) + sprintf_ ; 格式字串 "BLDG%d.LBX"
        idiv 10 → edx               ; 檔內型別 = (id−1) % 10
        imul ebx, edx, 24h          ; ×36
        call sub_BC8A6              ; (a,b) → 格號
        add edx, ebx                ; 資產 = 檔內型別×36 + 格號

    `sub_BC8A6` 也是算式不是表——**蛇行**:`slot = b×6 + a`(b 偶數)/ `b×6 + 5 − a`(b 奇數),
    與實測 BLDG0 資產 0..35 墨點的走法一致。

    ### 建築編號本身:兩個獨立來源對上

    | 來源 | 內容 |
    |---|---|
    | openorion2 `src/gamestate.h` | `BUILDING_NONE = 0` 起的 `BUILDING_*` 列舉,48 棟 |
    | 原版 `TECHNAME.LBX` 資產 0 | 第 295 條起 "No Building"、"Alien Control Center"、"Armor Barracks"…"Artificial Planet",**逐條與列舉同序**(openorion2 `src/lang.h` 亦寫 `TNAME_BUILDING_NONE 295`)|

    一個是別人重製專案的列舉、一個是原版資料檔的字串順序——對得起來才算數。
    remake 的 40 棟對照表寫在 `cmd/moo2/colonysurface.go` 的 `origBuildingID`,
    要手寫是因為手冊用字和遊戲內部字串不同("Automated Factories" vs "Automated Factory"、
    "Alien Management Center" vs "Alien Control Center"、"Planetary Stock Exchange" vs
    "Stock Exchange"),照字串比對會漏一半。

    ### 端到端驗證

    離線合成一張:取 9 棟建築、依算式算出各自的 (檔, 資產),疊上去,再把格點畫在最上層。
    **每一棟都正正落在自己的格子裡、站在格線平面上、遠近遮擋正確**。
    幾何來自執行檔資料段、資產編號來自執行檔算式、建築編號來自另外兩個來源——
    三條線各自獨立,最後在同一張圖上對上。

    ### 還缺的(不猜)

    - **哪一格放哪棟**的規則(`Make_Bldg_Array_For_Colony_` / `Sort_Bldg_Array_Columns_` /
      `Insert_Bldg_Into_Array_` / `Find_Replacement_Slot_For_Building_`)。
    - **陰影**:建築圖裡最大宗的顏色是 (108,48,108) 的洋紅,那是原版用來和地表混色的
      陰影索引,直接畫會變成一塊洋紅。要先弄清楚原版怎麼混。
    - **地表底圖**本身:`C_Anims` 的動畫清單是依行星氣候載入的,還沒追到來源 LBX。
      `COLBLDG.LBX#0` 不是地表,是**建造彈出視窗的框架**(Auto Build / REFIT / DESIGN /
      REPEAT BUILD / CANCEL / OK,中段透明)。
    - remake 目前把建造佇列放在中段——原版那裡是地表,佇列是獨立彈出視窗
      `Build_Queue_Popup_` @ 0xB4041(7 格 x 207..458、y 329+20i,座標已到手)。
      要上地表就得先把佇列搬進那個彈出視窗。

    落地的部分:`cmd/moo2/colonysurface.go`(角點表 + 走訪順序 + 格四角/中心/命中測試 +
    建築編號對照 + 資產算式)與 9 條回歸測試(含「每棟都對到編號」「對照表沒有多餘項」
    「算出的資產落在該檔真實資產數內」)。

    **畫面還沒改**,剩兩件事擋著:①上面那兩項(擺放規則、陰影混色);②remake 現在把建造佇列
    放在中段,而原版那裡是地表、佇列是獨立彈出視窗 `Build_Queue_Popup_` @ 0xB4041
    (7 格 x 207..458、y 329+20i,座標已到手)——要上地表得先把佇列搬進那個視窗。

28. **原版建築表挖出來:所有「估計的」建造成本換成真值**(2026-08-07)。

    第 27 項在追建築 sprite 時,`Real_Building_Name_` @ 0xBB40D 只有兩行——
    `imul eax, 13h` + `mov eax, off_17EB3D[eax]`——就指出資料段裡有一張**49 筆、每筆 19
    位元組**的建築表。抽出來的欄位:

    | 位移 | 型別 | 內容 |
    |---:|---|---|
    | +0 | dword | 名稱指標(全表同值,執行期才填)|
    | +4 | word | 建築編號(自身索引)|
    | +6 | word | 前置科技索引 |
    | +8 | dword | **建造成本 PP** |
    | +12 | word | **維護費 BC/turn** |
    | +14 | byte | 分類(**7 = 軌道衛星**)|

    ### 怎麼確定 +8 是 PP、+12 是維護費(不是用猜的)

    - Armor Barracks 的 +8 = **150**,正好是 remake 先前**唯一**有實據的那筆
      (`MANUAL_150.html` modding 範例)。
    - +12 與 remake 從手冊逐條抄的維護費 **40/40 全中**。那 40 個值是獨立來源,
      能全中代表欄位判讀**與**手寫的「remake 建築 ↔ 原版編號」對照表都對。
    - +14 = 7 的那五筆正好是 Artemis System Net / Battlestation / Dimensional Portal /
      Star Base / Star Fortress —— 也就是全部的軌道衛星,而
      `Make_Bldg_Array_For_Colony_` 就是用這個欄位把衛星排除在地表格點外。

    ### 舊估計值錯得不小

    | 建築 | 舊估計 | 原版真值 |
    |---|---:|---:|
    | 星辰要塞 | 800 | **2500** |
    | 行星屏障護盾 | 1200 | 500 |
    | 歡樂穹頂 | 800 | 250 |
    | 核心廢料場 | 550 | 200 |
    | 深層核心礦坑 | 550 | 250 |
    | 食物複製機 | 460 | 200 |
    | 銀河網路 | 650 | 250 |

    `internal/gamedata/buildings.go` 的 `EstimatedCost` 欄位整個拿掉了——40 項全是真值,
    留著那個旗標只會誤導。`special_actions.go` 裡能對到編號的三項(地形改造 250、
    蓋亞轉化 500、土壤改良 120)一併換成真值;運輸艦隊/殖民船/前哨船仍是估計值
    (原版的艦艇造價由艦體 + 元件逐項算,不在這張表裡),旗標保留。

    ### 順帶:地表擺放規則也讀完了(但還缺 PRNG)

    - `Make_Bldg_Array_For_Colony_` @ 0xBC30B:`Set_Random_Seed(colonyIdx, 0, 144)` →
      **同一個殖民地的擺法是固定的**;依編號順序逐棟 `Insert_Bldg_Into_Array`,
      分類 7 的丟去衛星清單;再依 `人口/3 + 1` 補「房屋」;最後 `Sort_Bldg_Array_Columns_`
      依 +14 分類做氣泡排序,再跑一段隨機微調。
    - `Insert_Bldg_Into_Array_` @ 0xBC05E:對所有空格做**蓄水池抽樣**(`Random(++n) == 1`),
      滿了才叫 `Find_Replacement_Slot_For_Building_` 挑最低優先的格子擠掉。
    - **房屋是借用衛星的編號畫的**:`-3` 依 `(colonyIdx + 房屋數 + 1) % 4` 存成
      3 / 14 / 40 / 41(Artemis / Dimensional Portal / Star Base / Star Fortress)。
      衛星本來就不畫在地表,編號空著,於是拿來當四種房屋外觀——
      montage 裡型別 2/13/39/40 確實是小房子而不是衛星,對得上。
    - **卡住的點**:擺法要完全重現得先實作原版的 `Random_` @ 0x1247A0 /
      `Set_Random_Seed_` @ 0x124820。那是下一步,不是阻塞。

29. **半透明標記索引:index >= 0xF0 從來不是顏色**(2026-08-07)。

    第 27 項留下的「陰影是洋紅」問題,追下去發現不只是陰影,而是 remake 一直誤解了
    一整類像素。原版的通用繪圖常式(module 168)對 **來源索引 >= 0xF0 的像素從不直接寫進
    畫面**,兩條路徑各有做法,由 `Draw_Bldg_CR_` @ 0xBB469 依 `byte_182ACA` 二選一:

    | 路徑 | 內圈 | 對 >= 0xF0 的處理 |
    |---|---|---|
    | `Draw_` @ 0x12A478 | `sub_12ACA4` | `dst = blendTable[(src << 8) + dst]` —— 混色查表 |
    | `Draw_No_Glass_` @ 0x129FF9 | `sub_12AAA1` | `cmp eax, 0F0h / jge` → **整個像素跳過** |

    也就是說 240..255 是**十六種半透明標記**(陰影、光束輝光、玻璃罩…),不是十六個顏色。

    ### 用量:不能一律當透明

    掃過玩家資料夾全部 LBX:

    | LBX | >= 0xF0 的像素占比 |
    |---|---:|
    | SPHERSFX | 100% |
    | BEAMS | 69% |
    | LOGO | 39% |
    | CMBTSFX | 35% |
    | COLONY | 27% |
    | MAINMENU | 13% / RACESEL | 13% |
    | BLDG0..4 | 7–11% |

    光束本來就是疊在星空上的半透明輝光——一律當透明會讓武器整個消失。**要不要丟由
    呼叫端依畫面決定**,不能在解碼層一刀切。

    ### 落地與還缺的

    `internal/lbx/image.go` 加了 `TranslucentIndexMin = 0xF0`、`Frame.HasTranslucent()`
    與 `Frame.ToRGBADropTranslucent()`(= `Draw_No_Glass_` 那條路徑,底下沒東西可混色時
    這是原版唯一適用的那條,不是近似)。既有的 `ToRGBA` 行為不變,不動到現有畫面。

    **混色表 `byte_1AB358` 在 BSS,執行期才建,不在執行檔裡**;產生它的程式碼還沒找到
    (寫入是透過指標,`.asm` 裡搜不到直接寫這個符號的指令)。openorion2 也沒解掉,
    `galaxy.cpp` 留著 `FIXME: analyze original game and calculate better shadow palette`。
    在找到之前不假造混色係數。

    ⚠ 順手學到的:預覽時我用「RGB 等於 (108,48,108) 就丟掉」做快速合成,結果**還有幾塊
    洋紅沒清掉**——因為 241..255 各自對到不同顏色。這正好反證了要**用索引判斷、不是用顏色**。

30. **中段還給行星表面,建造佇列搬回原版的彈出視窗**(2026-08-07,第 26/27/28/29 項的收尾)。

    第 27 項推翻的那句「中段(y 159..423)原版畫行星表面」其實還有更深一層:remake 把
    **建造佇列與可建清單**塞在中段,那是 remake 自己的版面;原版那裡是地表,佇列是另外
    一張彈出視窗。這一項把兩邊都搬回原位。

    ### 建造彈出視窗(`Build_Queue_Popup_` @ 0xB4041 → `cmd/moo2/buildqueue.go`)

    框架是 **COLBLDG.LBX#0**(先前誤以為是地表底圖,其實是這張;640×480、中段透明,
    烘著六顆鈕)。`Draw_Build_Queue_Popup_` @ 0xB3CF7 用 `sub_12A478(0, 0, img)` 貼在
    (0,0),所以裡面的座標全是**螢幕絕對座標**:

    | 元件 | 來源 | 座標 |
    |---|---|---|
    | 可建清單 | `Add_Buildings_Fields_` @ 0xB08CA | x 13..184、y = 20 + 19i(列高 19)|
    | 佇列 7 格 | `Add_Build_Queue_Fields_` @ 0xB325A | x 207..458、y = 329 + 20i(高 21)|
    | Auto Build | `sub_11523B`(toggle)| (490, 342) |
    | REFIT / DESIGN | `sub_1151B0` | (492, 379) / (561, 379) |
    | REPEAT BUILD | `sub_1151B0` | (503, 411) |
    | CANCEL / OK | `sub_1151B0` | (493, 447) / (560, 447) |

    六顆鈕**不另外貼資產**——框架圖上已經烘好了,再貼一次會疊出雙重邊框。中文模式擦底疊中文,
    英文模式讓路。REFIT / REPEAT BUILD / Auto Build remake 未接,畫成灰的不假裝可按。

    入口是框架上那顆 **CHANGE**(原版它就是「換要蓋什麼」),先前畫成灰的沒接功能。

    ### 殖民地畫面中段 → 行星表面

    `drawColonySurface`:格線(36 格,角點全是反組譯真值)+ 建築圖(已畫好位置的 640×480
    稀疏圖,遠→近依 `colonySlotOrder` 疊,半透明標記依 `Draw_No_Glass_` 丟掉)。

    ⚠ **哪一格放哪棟與原版不同,這點在程式碼裡明寫了。** 原版
    `Make_Bldg_Array_For_Colony_` 的擺法綁死在原版 PRNG(`Set_Random_Seed(colonyIdx, 0, 144)`
    → 蓄水池抽樣),那還沒實作。在 PRNG 到手之前用「依編號從近端往遠端填」——
    幾何、圖檔、遮擋順序、半透明處理全是原版真值,**只有落在哪一格是 remake 自己的**。
    房屋那段倒是照原版:數量 = 人口/3 + 1,外觀在 3/14/40/41 之間輪。

    地表底圖(`C_Anims(1, 0, 639, 479)` 那張,依氣候載入)還沒追到來源 LBX,目前只有格線。

    > **上面這三段已由第 31 項取代,別再當現況引用**:PRNG 已照抄、擺放已是原版演算法、
    > 地表底圖已接上、格線已移除。留著是為了看得到當時的判斷長什麼樣。

    ### 驗證

    截圖廊加了第 26 張(建造視窗)。中文模式全 17 張逐像素比對:**只有殖民地畫面變了**
    (預期中),其餘未變。
    ⚠ 主選單那張會因為「上一次跑的 demo 存檔還在不在」而讓 繼續 / 載入遊戲 的灰階不同——
    同一個容器裡連跑兩次截圖廊就會出現,不是程式改動。做截圖 diff 時要知道這件事。

31. **原版 PRNG 照抄 + 建築擺放換成原版演算法 + 地表底圖接上**(2026-08-07,第 30 項的續)。

    ### 原版 PRNG(`Random_` @ 0x1247A0 → `internal/gamedata/origrand.go`)

    32-bit LCG,乘數 `0x41C64E6D`、增量 `0x3039`(就是 ANSI C `rand()` 那組常數,
    但**保留完整 32-bit state**,沒有 `>>16` 也沒有 `& 0x7FFFFFFF`),外加**拒絕取樣**:
    `bucket = 0xFFFFFFFF / n`、`limit = bucket × n`,抽到 `>= limit` 就重抽,
    回傳 `state / bucket + 1` —— **1..n,不是 0..n-1**。`n == 0` 直接回 1 且不遞推。

    為什麼值得照抄:MOO2 好幾處靠特定種子重現特定結果,殖民地建築擺法就是
    `Set_Random_Seed(colonyIdx)` 起手。用 Go 的 `math/rand` 換不出同一組數。

    ### 建築擺放(`Make_Bldg_Array_For_Colony_` @ 0xBC30B)

    | 步驟 | 原版 | remake |
    |---|---|---|
    | 種子 | `Set_Random_Seed(colonyIdx)` | 同 |
    | 建築 | id 0..48 依序,分類 7(`byte_17EB4B[id×19]`)丟去衛星清單 | 同(id 1..48;0 是空格) |
    | 選格 | `Insert_Bldg_Into_Array_` @ 0xBC05E 的**蓄水池抽樣**:走訪 36 格,遇空格 `n++` 且 `Random(n) == 1` 就選它 | 同 |
    | 房屋 | 數量 = 人口/3 + 1,存進格陣的值是 3 / 0x0E / 0x28 / 0x29(跳表 `jpt_BC18C`) | 同 |
    | 排序 | `Sort_Bldg_Array_Columns_` @ 0xBBDC9:依分類氣泡排序,每格比三個鄰格,最多 6 輪 | 同 |
    | 微調 | 排序後 8 輪、對房屋鄰格約 1/3 機率互換 | ❌ 未做(見下) |

    房屋借的四個編號(3 / 14 / 40 / 41)**分類都是 7**,所以排序會把房屋推到索引最大那一角。
    這是原版本來的行為(排序的用意就是把同類聚在一起),不是 bug。格子稀疏時大多數鄰格對
    有一邊是空的、不參與交換,所以排序幾乎不動——看起來仍然很隨機,也是對的。

    ### ⚠ 抓到一個「不會有症狀」的軸向錯誤

    第一版把格陣索引的**外層當成角點表的第二維**,整張佈局沿對角線鏡射掉。
    徵狀只有「建築全擠在遠端那幾排」——太容易被當成隨機而放過。

    定案證據在 `Add_Bldg_Fields_` @ 0xBE44A,同一對 (v1, v2) 同時餵給兩邊:

    ```
    colony_bldgs[24×v1 + 4×v2]                        ; 格陣元素 → v1 是「列」
    Bldg_Coords_To_Centered_Screen_Coord(v1, v2, c)   ; 螢幕座標 → 索引 v1×56 + v2×8 + c×4
    ```

    → **格陣 `grid[a×6 + b]` ↔ 角點 `colonyGridCorners[a][b]`**,a 是步距 56 那一維。
    格號的算式**反過來**:`sub_BC8A6(a, b) = b×6 + a`(b 奇數時 `b×6 + 5 − a`)。
    格陣索引 a×6+b 與格號 b×6+a **不是同一個數**,別互相代用。
    護欄:`TestColonyGridKeyMatchesOriginalAddressing`(把上面兩行位址算式寫成可執行斷言)。

    ### 地表底圖(`Draw_Colony_Screen_` @ 0xBED21 開場那兩層)

    `C_Anims` @ 0xBBA8E 是一張 36 路跳表,解出來:

    | 層 | 來源 | 內容 |
    |---|---|---|
    | 底 | `C_Anims(1)` → **COLONY2.LBX#49** | 星空(整個檔只有這張是 640×480)|
    | 上 | `C_Anims(0)` → **PLANETS.LBX#N** | 地形(天空部分透明,露出底層)|

    ```
    資產 N = byte[colony + 0xE2] × 3 + byte[star_table + star×17 + 9]
    ```

    PLANETS.LBX 恰好 30 張 640×480 → **10 種氣候 × 3 個變體**,所以 +0xE2 欄是氣候(0..9)、
    星球表 +9 欄是變體(0..2)。渲染出來對得上:#27 = 9×3 整片蔥綠(Gaia)、#3 = 1×3
    赤紅熔岩裂縫(Radiated)、#0..2 是有毒星的三種面貌。remake 的 `gamedata.PlanetClimate`
    (TOXIC=0 … GAIA=9)與這個順序逐項相同。

    ⚠ **變體那一欄還沒還原**:它是銀河生成時寫進星球表 +9 的存檔欄位,產生規則沒追。
    remake 改用原版 PRNG 以星球索引起種取 1..3 —— 保證「同一顆星每次進去長一樣」這個玩家
    看得到的性質,但**不保證與原版同一局的那顆星相同**。追到規則前不要把它寫成真值。

    調色盤:PLANETS 每張只帶 80 色,缺的要基底補(`buffer0.lbx#0`),否則熔岩裂縫會變成
    一片洋紅——與第 29 項的洋紅是**不同原因**,別混為一談(那次是 index ≥ 0xF0)。

    ### ⚠ 繪製順序:地表在框架**之前**、建築在框架**之後**

    原版 `Draw_Colony_Screen_` 的次序是
    地表兩層 → `Draw_Colony_Info_Background`(框架)→ `Draw_Colony_Bldgs` → 資訊面板。
    兩個方向都踩過:

    - 框架先畫、地表後畫 → 底層那整片星空把上方三個資訊面板蓋成一片雜點。
    - 地表與建築一起畫在框架之前 → 近端那幾格的 y 壓到 423 以下(框架非透明區)會被切掉。

    原版沒有格線,佔位用的淡綠格線已移除。

    ### 軌道衛星(`Draw_Colony_Satellites_` @ 0xBE366,同日補上)

    分類 7 的建築不進地表格點,`Make_Bldg_Array_For_Colony_` 把它們丟進 `word_19F99C`
    那份 10 格的清單(依建築編號遞增附加),由這支另外畫。位置整段沒有查表:

    ```
    x = 295 + (i 偶數 ? +1 : −1) × i × 50     ; 295 = 0x127、50 = 0x32
    y = 162                                   ; 0xA2,固定
    ```

    → 295 / 245 / 395 / 145 / 495 / 45 / 595 …往兩側交錯散開。第 7 顆 x = −55 只剩右緣 2px,
    所以清單雖有 10 格,看得到的就是前 7 顆(原版本來的裁切)。

    圖檔來自 `sub_BE306` 的比較鏈,但**編號要經過 `sub_BBB9F` 的 `loc_BBBAF: add edx, 9`**:

    | 建築 | 編號 | `sub_BE306` | COLONY.LBX 資產 |
    |---|---|---|---|
    | 星際要塞 Star Fortress | 41 | 0 | 9 |
    | 戰鬥站 Battlestation | 8 | 1 | 10 |
    | 星基 Star Base | 40 | 2 | 11 |
    | 次元傳送門 Dimensional Portal | 14 | 3 | 12 |
    | 天網 Artemis System Net | 3 | 7 | 16 |

    ⚠ **那個 +9 不能漏。** 漏掉會去讀 COLONY.LBX 資產 0..4,而那五格在檔案裡是**零長度**的
    (offset 表開頭六個值都是 0x800),解出來是空圖 —— 畫面上什麼都不會出現,而且**不會報錯**。
    資產 9..16 全是 57×70,尺寸自洽;測試 `TestOrigSatelliteAssetsCoverEveryCategory7`
    直接把「資產必須落在 9..16」寫成斷言。

    抑制規則 `sub_BC21B`(回傳 1 = 這顆不畫)就是原版的**星基升級鏈**:
    星基(40)有戰鬥站(colony+0x13E)或星際要塞(colony+0x15F)就不畫;
    戰鬥站(8)有星際要塞就不畫。`0x136 + id` 是「這個殖民地有沒有這棟」的旗標陣列,
    0x136+8 = 0x13E、0x136+41 = 0x15F,兩個位移都對得起來。

    ### 仍是缺口

    - **擺放的最後一段微調**沒做(見上表)。影響房屋彼此的相對位置,**以及道路**——
      道路接在同一條亂數流上,抖動沒補則流位置錯開(見第 41 項)。
    - ~~**道路** `Draw_Road_List` 沒畫。~~ 第 41 項已補。
    - 衛星的**選取態閃爍**(`sub_BE271` 那條分支)沒做——remake 沒有點選衛星的互動。

    ### 驗證

    中文模式全 17 張逐像素比對:**只有殖民地畫面變了**,其餘未變。
    IDA 的 `-Ohexrays` 批次反編這一輪起不來(`Failed to initialize IDA as library, error code 4`,
    兩次都是),本項全部改用手讀 `.asm` 完成——軸向那個錯就是這樣抓到的。

32. **星圖:艦隊圖示換成原版的、旗色順序修正、拿掉浮在星圖上的研究文字**(2026-08-07)。

    ### 先把星圖的圖層順序抄下來(`Draw_Main_Main_Screen_` @ 0x8440E)

    | # | 圖層 | 位址 | remake |
    |---|---|---|---|
    | 1 | `Draw_Wormhole_Links_` | 0x85593 | ✅ 資料模型 + 連線 + 航行機制(第 35 項)|
    | 2 | `Draw_Relocation_Links_` | 0x85320 | ❌ |
    | 3 | `Draw_Stars_` → 逐星 `Draw_A_Star_` | 0x85550 / 0x83B02 | ✅ sprite 已接(第 34 項);閃爍動畫未做 |
    | 4 | `Draw_A_Gate_Icon_`(迴圈) | 0x83741 | ❌ |
    | 5 | `Print_Star_Names_` | 0x88CB7 | ⚠ 位置已對齊(第 34 項);字型樣式/描邊未做 |
    | 6 | `Draw_Black_Holes_` | 0x83BF9 | ❌ |
    | 7 | `Draw_Ship_Icons_` | 0xA070F | ✅ 本項 |
    | 8 | `Print_Main_Screen_Data_` | 0x87BAE | ⚠ 簡化(右欄五格數字)|
    | 9 | `Draw_Diplomacy_Request_Lights_` | 0x83D06 | ❌ |

    外層 `Draw_Main_Screen_` 另有 `Draw_Nebulae_` @ 0x84F8F 與 `Draw_Paralax_` @ 0x8500F。
    (視差背景已於第 33 項補上;星雲仍未做。)

    ### 艦隊圖示(`Get_Ship_Icon_Pict_Seg_` @ 0xA0D78)

    ```
    帝國艦隊(id 0..7):BUFFER0.LBX 資產 = 0xCD(205) + 旗色×4 + 縮放
    id 8              :              = 0xED(237) + 縮放
    id 9..14          :              = 0xF1(241) + (id−9)×4 + 縮放
    ```

    「×4」是因為原版星圖有**四段縮放**,每段一張圖(實測 205..208 = 11×11 / 12×11 /
    12×10 / 16×12,由小到大)。縮放由 `sub_79917` 給,原版再反過來映射(3→0、2→1、1→2、0→3)
    ——縮得最遠用最小的圖。remake 的星圖沒有縮放(固定全銀河檢視 = 原版縮得最遠那段),
    所以固定用縮放 0。每張圖有 8 幀(`Cycle_Ship_Icons_` @ 0x82DFF 在跑動畫),remake 只取第 0 幀。

    remake 先前在艦隊所在星畫一個 8×8 青色方塊,那是佔位。

    ⚠ 位置仍用 remake 自己算的星座標,不是 `Get_Ship_Icon_Coords_` @ 0xA0A5C。
    原版那支對 `word_1906C6` 做六路分支(艦隊在星系內的停泊槽 / 航行中的插值),
    要有對應的艦隊位置模型才有意義,remake 的艦隊移動是整段跳的。

    ### ⚠ 旗色順序錯位(又是一個看不出症狀的錯)

    艦隊圖示既然是 `205 + 旗色×4`,**旗色的順序本身就是資料**。
    remake 先前是 紅/黃/綠/**藍/白/紫/橙/棕**,後五個全錯位——選紅色沒事,選藍色會開出
    白色的艦隊。中文模式完全看不出來(顏色名還是對的)。

    兩個獨立來源對上才改:

    | 索引 | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
    |---|---|---|---|---|---|---|---|---|
    | BUFFER0 205/209/…/233 量到的代表色 | (192,101,96) | (193,173,28) | (81,155,61) | (190,190,198) | (143,184,216) | (198,154,111) | (196,141,193) | (223,139,45) |
    | openorion2 `gfx.h` `FONT_COLOR_PLAYER_*` | RED | YELLOW | GREEN | **SILVER** | BLUE | BROWN | PURPLE | ORANGE |

    一個量自原版美術、一個抄自別人的重製專案,逐項相同。原版第 4 色叫 **SILVER**(銀)
    不是 White——量到的 (190,190,198) 也確實偏灰。護欄:`TestFlagColorsMatchOriginalOrder`。

    ### 拿掉星圖左上那行「研究:<主題名>」

    那是 remake 自己加的一行,**原版星圖沒有這個東西**(研究主題只在研究畫面顯示),
    而且它就壓在左上角的星星與艦隊圖示上。原版右欄第 5 格(綠燒瓶)放的是數字,跟其他四格
    一樣——改成每回合研究點數,位置 (579, 405)(其他四格是 110 / 182 / 257 / 331,
    間距 ~74,第 5 格順推)。

    ### 驗證

    中文模式 17 張逐像素比對:變的是星圖、命名旗色、軌道轟炸三張(後兩張用 `FlagColors`),
    其餘未變。

33. **星圖視差星空背景**(2026-08-07,第 32 項的續)。

    remake 的星圖區先前是一片 `RGBA{6,6,16}` 純黑,那是佔位。原版底下鋪的是三層星空。

    ### `Draw_Paralax_` @ 0x8500F

    結構完全規則,三層一模一樣只差來源與位移:

    ```
    sub_8FD71(edx=0x16, ebx=0x20F, ecx=0x1A5)   ; 裁切區 x 22..527、y ..421
    for layer in 0, 1, 2:
        img = STARBG.LBX 資產 layer            ; 實測皆 640×480
        Draw(x,       y      )
        Draw(x − 640, y      )                  ; 0x280 = 640
        Draw(x − 640, y − 480)                  ; 0x1E0 = 480
        Draw(x,       y − 480)
    ```

    四次貼圖是**環繞平鋪**——捲動時從右邊/下面出去的要從左邊/上面接回來。位移是三對全域:

    | 層 | x | y |
    |---|---|---|
    | 0 | `word_199980` | `word_19997E` |
    | 1 | `word_19998A` | `word_19998E` |
    | 2 | `word_199984` | `word_199986` |

    三對各自更新且速率不同 —— 這就是視差:近的捲得快、遠的慢。

    ⚠ remake 的星圖不捲動(固定全銀河檢視),三個位移都是 0,環繞平鋪的另外三次全在畫布外,
    等於三層各貼一次 (0,0)。做出捲動時要把那四次補回來,不是改圖或改公式。

    ### ⚠ 底色不能省(踩過)

    第一版把純黑底改成「有原版資產就不填」,結果整片星圖變成**白底黑點**。
    原因不是調色盤錯——探針量出來層 0 最多的三個索引是 2/3/5,對到
    (16,16,24) / (24,24,32) / (44,44,52),**本來就是極暗的藍灰點疊在透明上**。
    底色一拿掉,透明處就露出底下 `buffer0.lbx#0` 的框架美術。
    原版也是先 `Fill` 再貼視差層(`Draw_Main_Screen_Filled_`)。

    ⚠ 這條規則**測不到**:ebiten 在 game loop 之外不准 `Image.At()`
    (`ui: ReadPixels cannot be called before the game starts`),所以「星圖框內是不是純黑」
    沒辦法寫成單元測試,只能靠截圖廊第 04 張驗收。規則寫在 `starBGFill` 的註解裡。

    ### 順帶修掉一個 nil 解參考

    `decodeAsset` 對 nil `*assets.Resolver` 會 nil 解參考。畫面層的降級路徑本來就該容許
    「資料夾不完整」,所以 `starBGImage` / `shipIconImage` / `colonyScreenImage` 三個都補了
    `b.res == nil` 守衛,並加了回歸測試。

    ### 仍是缺口

    - **星雲** `Draw_Nebulae_` @ 0x84F8F:每個星雲 `x = scale(nebula.x − 捲動) + 21`、
      `y` 同理,圖從 `dword_190298[i*16 + 縮放*4]` 這個快取取。**需要銀河生成先產出星雲表**
      (位置 + 型別),remake 的星系產生器目前沒有這個欄位——這是資料模型缺口,不是繪圖缺口。
    - 蟲洞連線、遷移連線、星門、黑洞、外交燈號:同樣都缺對應的資料模型或旗標。

34. **星圖:星球換成原版 sprite**(2026-08-07,第 33 項的續)。

    remake 先前用 `vector.DrawFilledCircle` 畫色圓,那是佔位。

    ### 資產編號:三個獨立來源互相印證

    | 來源 | 內容 |
    |---|---|
    | 反組譯 `Get_Star_Picture_Seg_` @ 0x81CD3 | `edx = al×6`;`al == 6` 時 `eax = 縮放`,否則 `eax = 縮放 + bl`;資產 = `0x94 + eax + edx` |
    | openorion2 `galaxy.cpp:1664` | `_starimg[s->spectralClass][_zoom + s->size]`,`ASSET_GALAXY_STAR_IMAGES 148`、`STAR_TYPE_COUNT 6`、`GALAXY_STAR_SIZES 6` |
    | 實測 BUFFER0.LBX | 148..183 正好 6 組 × 6 張,每組尺寸 33/29/25/23/21/17(遞減),各 5 幀 |

    → **資產 = 148 + 光譜×6 +(縮放 + 大小)**

    列舉 remake 本來就對上:`Star.Spectral` 0=藍…6=黑洞 = openorion2 `SpectralClass`;
    `Star.Size` 0=大..3=小 = `StarSize{Large, Medium, Small, Tiny}`。

    ### 公式自己證明了自己(黑洞那條分支)

    光譜 6 不加大小,算出來是 `148 + 6×6 + 縮放` = **184 + 縮放**。而 openorion2
    `galaxy.cpp:45` 另外命名了 `#define ASSET_GALAXY_BHOLE_IMAGES 184`,用法是
    `_bholeimg[縮放]`。**兩邊落在同一個數字上**——一邊是從組語推出來的算式、一邊是別人專案裡
    獨立命名的常數。這同時證實了基底 148、每組 6 張、光譜×6、以及「光譜 6 = 黑洞」四件事。

    ### ⚠ 縮放:原版那個值 remake 沒有對應物

    `Map_Scale_To_Zoom_Level_` @ 0x79917 把地圖比例(`word_199992` = 10/15/20/30,
    即 openorion2 的 `galaxySizeFactors`)換成縮放 0..3。openorion2 `galaxy.cpp:1385`
    揭露關鍵:**縮放是銀河尺寸的函式,不是玩家控制的**——銀河越大 → sizeFactor 越大 →
    星球畫得越小,因為要塞進同一個視窗。螢幕座標是 `21 + 10×(x − 捲動)/sizeFactor`
    (`transformX`),sizeFactor 越大距離越短,所以 **0 = 最放大、3 = 最縮小**。

    remake 這裡有**模型落差**:它把星球座標正規化成 0..1 攤滿視窗,銀河多大都一樣,也沒有
    捲動,所以那個對應接不過來,**不存在忠實值**。固定用 3(最縮小),理由是 remake 永遠
    一次顯示整個銀河。**這是 remake 的選擇不是原版真值**,做出捲動/縮放前別把它寫成真值。

    順帶:這個縮放語意也解釋了第 32 項艦隊圖示那個看起來反直覺的映射(3→0、2→1、1→2、0→3)
    ——縮小到底(3)配最小的圖(11×11),完全合理。

    ### 仍是缺口

    - 星球的 **5 幀閃爍動畫**沒做,只取第 0 幀(同艦隊圖示的 8 幀)。
    - 黑洞在原版是 `Draw_Black_Holes_` @ 0x83BF9 的**獨立迴圈**(黑洞是星球以外的地圖物件),
      remake 把它當光譜 6 的星球一起畫。圖是對的,但**阻擋航線**那套(`Star.blackHoleBlocks`)沒有。

    ### 星名位置(`Print_A_Star_Name_` @ 0x87768,併入第 34 項)

    原版是**置中在星球正下方**並夾在星圖框內,remake 先前畫在星球右側,長名字直接壓出框外:

    ```
    x = 星球中心 − 字寬/2          ; sub_12066F 量寬再減半
    y = 星球中心 + sprite 邊長/2 − 大小
    夾擠:x >= 0x16(22)、x + 字寬 <= 0x20F(527)   ← 與視差層的裁切區同一組數字
    ```

    ⚠ 顏色不是依擁有者:`sub_7A440` 的真名是 `Zoom_Level_Font_Style_`,選的是**縮放對應的
    字型樣式**(縮得越遠字越小),再加 `Set_Outline_Color(2)` 的描邊。
    (`FONT_COLOR_PLAYER_*` 那 5 階×8 色是別處在用,不是星名。)
    remake 的 CJK 字型沒有對應的樣式表,只搬了位置與夾擠。

35. **蟲洞:把星圖第 1 層卡住的資料模型補上**(2026-08-07)。

    第 33/34 項留下的結論是「剩下的層卡的是資料模型不是繪圖」。這一項就是去補模型——
    蟲洞是其中最有遊戲價值的一個(它是**機制**不是裝飾:把銀河兩頭接起來)。

    ### ⚠ 先分清楚兩個都叫「蟲洞」的東西

    MOO2 兩者都有,remake 也是,別合併:

    | | 是什麼 | remake |
    |---|---|---|
    | **隨機事件** | 一次性好事,把正在航行的艦隊直接送到(手冊 p.181「moves that fleet to their destination in a single turn」)| 早就有(`applyWormhole`)|
    | **星圖蟲洞** | **永久**連在星圖上的兩星捷徑,任何時候都能走 | 本項新增 |

    兩者共用「一回合就到」的語意,那句手冊原文正好當本項 ETA = 1 的錨點。

    ### 資料模型

    `Star.Wormhole int`(-1 = 無),對應原版星球結構 +0x29(int8,0xFF = 無)。
    **必須雙向**——openorion2 `gamestate.cpp:1946` 對單向蟲洞直接丟例外
    ("One-way wormholes not allowed"),原版 `Draw_Wormhole_Links_` 也是兩端各畫一次。

    ⚠ **舊存檔沒有這個欄位,JSON 解出來是零值 0**,那會讓每顆星都宣稱與星 0 有蟲洞
    (星圖畫滿放射狀連線、艦隊到處一回合直達)。讀檔一律走 `normalizeWormholes`,
    把越界/自己連自己/單向的一律清成 -1。這是本項最容易靜默壞掉的地方,有專門的測試。

    ### 產生規則(照抄得到的部分)

    `Generate_Wormhole_Links_` @ 0x8CC15 + `Set_Wormhole_Id_` @ 0x8D6D6:

    - 母星不可當端點(先把每個玩家的母星標進 `taken[]`)
    - 黑洞不可當端點;已有蟲洞的不再配
    - 挑第一端用**最多 200 次(0xC8)的拒絕取樣**
    - 第二端要過**最短距離門檻**(原版 `galaxySizeParam + 3`)——蟲洞不連鄰星
    - 候選收滿 19 個(0x13)就停

    ### ⚠ 沒照抄的:數量與距離門檻的單位

    `_n_wormholes`(`byte_182245`)**不是常數**——它在銀河產生過程中逐星累加
    (`sub_8C840`),上限 `galaxySizeParam × 4 + 4`。要忠實重現得連整個銀河產生器一起搬,
    而 remake 的星圖是「格點 + 抖動」的自有模型,兩邊接不起來。
    改用**與上限同構**的 `星數/8` 夾在 1..4(24 星給 3 個)。
    最短距離同理:原版的 `galaxySizeParam + 3` 是銀河座標單位,remake 是正規化 0..1,
    沒有換算依據,用 0.45 保住「不連鄰星」的語意。**兩者都是 remake 的選擇,不是原版真值。**

    ### 繪製與機制

    - 星圖第 1 層(在星球**之前**,所以線被星球蓋住而不是壓在上面),兩端都沒偵測到就不揭露
      (原版有同樣的探索狀態檢查)。remake 只畫 `i < j` 那一次,結果與原版兩端各畫一次相同。
    - `SendFleet`:兩端有蟲洞 → ETA 固定 1,不看距離。

36. **原版 48 棟建築,remake 到底缺哪幾棟?——把 8 個沒建模的編號查清楚**(2026-08-07)。

    `gamedata.Buildings` 是 40 項(手冊《The Big List》35 建築 + 5 衛星),而原版建築表是 48 棟。
    先前沒人把差集列出來過,所以「是不是缺了什麼」一直沒有答案。

    差集就是 `cmd/moo2/colonysurface.go` 的 `origBuildingID` 對不到的那 8 個編號。
    用 openorion2 `gamestate.h` 的 `BUILDING_*` 列舉認名字,再從原版建築表
    (`off_17EB3D`,19 位元組一列)讀真值:

    | id | 名稱 | 成本 PP | 維護 BC | 分類 | remake 現況 |
    |---:|---|---:|---:|---:|---|
    | 9 | Capitol | 200 | 0 | 1 | 建立殖民地時自動給予,不可建造——**正確地不在表裡** |
    | 11 | Colony Base | 200 | 0 | 0 | 同上(拓殖時自動) |
    | 17 | Gaia Transformation | 500 | 0 | 1 | ✅ `SpecialActions`(一次性) |
    | 18 | **Galactic Currency Exchange** | 250 | 3 | 5 | ❌ **完全沒有** |
    | 37 | Soil Enrichment | 120 | 0 | 0 | ✅ `SpecialActions` |
    | 42 | Stellar Converter(行星版) | 1000 | 6 | 0 | ✅ 2026-08-07 補上,見第 37 項 |
    | 44 | Terraforming | 250 | 0 | 0 | ✅ `SpecialActions` |
    | 48 | **Artificial Planet** | 800 | 0 | 0 | ❌ **完全沒有** |

    ### 這次抽表順帶驗了兩件事

    - 三個 `SpecialActions` 的成本(250 / 500 / 120)與這次重抽**逐項相同** —— 交叉驗證通過。
    - **維護費 0 = 一次性**:Terraforming / Gaia / Soil Enrichment / Artificial Planet 全是 0,
      而 Currency Exchange(3)與 Stellar Converter(6)有維護費 → 它們是**常駐建築**不是一次性。
      這條規則本身就能把 8 個編號分成「該進 `Buildings`」與「該進 `SpecialActions`」兩堆。

    ### ⚠ 順帶抓到一個死路科技

    `TOPIC_GALACTIC_ECONOMICS`(techtree.go,**6000 RP**)解鎖的是
    `TECH_GALACTIC_CURRENCY_EXCHANGE`,但 remake **沒有任何東西消費它**——
    玩家花 6000 研究點研究完,什麼也不會發生。這是目前科技樹裡最貴的一條死路。

    ### 為什麼這一輪沒有直接補上

    效果查不到,而**不編數字**是這個專案的紀律。四條路都走過:

    | 來源 | 結果 |
    |---|---|
    | 手冊《The Big List》(GAME_MANUAL.pdf)| **沒有這一條**——所以 remake 的 40 項表本來就抄不到它 |
    | patch 1.5 的 `MANUAL_150.html` | 零命中 |
    | 遊戲資料檔 | 掃過整個資料夾,只有 `HELP.LBX` / `TECHNAME.LBX` 有**名字**,沒有任何說明文字 |
    | 反組譯 | 建築表只有 19 位元組(名稱指標/編號/前置/成本/維護/分類),**沒有效果欄**——效果是寫死在程式碼裡的 |

    下一輪要補的話,路線是從 `colony + 0x136 + 18` 那個「已建」旗標的讀取點反追到收入計算,
    而不是再去翻手冊。成本/維護/前置/分類這四個真值已經在上表,不必重查。

37. **恆星轉換器(行星版)補上 + 一個乾淨的負面結果**(2026-08-07,第 36 項的續)。

    第 36 項列出三棟真的缺的建築,並寫下「下一輪的路線是從 `colony + 0x136 + id` 的旗標
    讀取點反追」。這一項就是去走那條路——結果有一棟做成了,兩棟得到了**為什麼做不成的證據**。

    ### 先做正對照,再下「不存在」的結論

    直接搜「有沒有讀 `colony + 0x148`(建築 18 的已建旗標)」會得到零命中,但零命中本身
    不能當結論(見 `~/diagnosis-notes/docs/02-query-returned-empty`)。所以改成**把 48 個
    位移全掃一遍**:

    - 43 個位移在反組譯裡出現過 → 技術本身有效(`sub_BC21B` 讀 `+0x13E` / `+0x15F` 就是實例)
    - **完全沒出現過的只有 5 個**:Artemis System Net(3)、Dimensional Portal(14)、
      Gaia Transformation(17)、Soil Enrichment(37)、Artificial Planet(48)

    這 5 個自己解釋了自己:兩個是**衛星**(由 `Draw_Colony_Satellites_` 另外處理),
    三個是**一次性**(做完效果就寫進行星/殖民地狀態,不需要再讀旗標)——
    而「維護費 0」那條規則挑出來的正是同一批。兩個獨立的判準指向同一個分組。

    ⚠ 至於建築 18:`+0x148]` 全檔只出現一次,而那一次的基底是 `dword_1AA21C`(科技已知陣列)
    不是殖民地。**同一個數字出現在不同結構裡**——這正是為什麼要做正對照而不是直接數命中。

    ### 做成的那一棟:恆星轉換器(行星版)

    兩個獨立來源對上才建模:

    | 來源 | 內容 |
    |---|---|
    | 手冊 p.106(`02-buildings.md` §十一)| 行星駐防版對目標造成 **400 傷 ×2**(共 1600),無視射程與防禦;維護 **6** |
    | 原版建築表第 42 列 | 成本 **1000 PP**、維護 **6**、分類 0(地表建築)、前置 Temporal Physics |

    維護費兩邊逐項相符,才敢動。接進 `colonyDefense`(+800)、`origBuildingID`(42,
    分類 0 → 地表 sprite 自動畫得出來)。

    ### ⚠ 順帶推翻一個記帳慣例被當成規則

    `TestBuildingsCount` 原本釘死 **40**,理由寫「手冊全表 35 建築 + 5 衛星」。
    但那是**手冊的記帳方式**——它把恆星轉換器放在單獨一節所以不計入,而原版建築表裡
    它就是第 42 棟,和其他 47 棟完全同構。**「40」從一開始就不是原版的數字**,
    只是抄手冊時連著記帳習慣一起抄了。現在是 41,測試註解裡寫明了為什麼會變。

    ### 另外兩棟為什麼還是沒做

    - **Galactic Currency Exchange**(18):效果四條路都查不到(第 36 項),而這次的位移掃描
      又證實它**不是**「一次性所以不用讀旗標」那一類。剩下的可能是效果走通用迴圈,
      要追 `Colony_BC_Production_` @ 0xE03F1 的快取欄位——那一支裡沒有任何建築旗標讀取,
      表示收入加成是預先算好存在殖民地結構的某個欄位。下一輪從那裡找。
    - **Artificial Planet**(48):旗標從沒被讀 + 維護費 0 → 確定是**一次性**,
      該進 `SpecialActions` 而不是 `Buildings`。但它的效果(把小行星帶變成行星)需要
      remake 的行星模型先有「小行星帶」這個型別,那是另一件事。成本 800 PP 已記錄。

    ### 仍未接進防禦解算的兩棟(有理由)

    飛彈基地與地面砲台手冊只給「佔 300 / 450 空間、裝當時最佳武器」的**規則**,
    傷害隨科技現算,沒有艦艇元件的空間模型算不出固定值。恆星轉換器是這批裡唯一給固定數字的,
    所以只有它接得進來。

38. **飛彈基地/地面砲台其實早就算得出來——是接線漏了,不是模型缺了**(2026-08-07,第 37 項的訂正)。

    第 37 項寫「飛彈基地與地面砲台手冊只給規則、傷害隨科技現算,**沒有艦艇元件的空間模型
    算不出固定值**」。**那句話是錯的。** 模型早就在:

    - `gamedata/satellite.go`:`MissileBaseSpace = 300`(手冊 p.78 確認值)、
      `GroundBatterySpace = 450`(p.81 確認值)、`SatelliteWeaponFitCount`、
      `SatelliteBeamSpaceWithArc`、`SatelliteStrengthScale`
    - `shell.retaliationAttackers` @ `orbital_bombardment.go`:已經用它算軌道轟炸的反擊,
      而且**連飛彈基地與地面砲台都已經支援**

    真正的缺口是 **`colonyDefense`(AI 突襲的防禦解算)沒接上這套**,它用的是
    `CommandPointsFromBuildings × 10` —— 一個自編係數,而且只認星基/戰鬥站/星辰要塞三級。

    ### ⚠ 順帶挖出一個自相矛盾:同一座星基在兩個地方值不同的分數

    | 路徑 | 星基值多少 |
    |---|---|
    | `colonyDefense`(舊)| `1 × 10` = **10** —— 比一艘巡洋艦(`shipStrength` 8)還強 |
    | `retaliationAttackers` | 依 space 預算推導 = **3–4** ≈ 驅逐艦 tier |

    而 `satellite.go` 的校準註解**明講**星基 ≈ 驅逐艦 tier(4)、戰鬥站 ≈ 巡洋艦 tier(8)。
    也就是說舊的 `×10` 與專案自己的校準文件互相打臉,**而且兩邊都測得綠**——
    因為沒有任何測試同時看這兩條路徑。現在有了(`TestColonyDefenceUsesSpaceBudgetModel`)。

    ### 改完之後

    `colonyDefense` 改用 `retaliationAttackers` 的 atk 加總,三件事一起對上:
    ① 反擊戰力隨已解鎖的武器科技成長,不再是寫死的 10/20/30;
    ② 飛彈基地與地面砲台真的有用了;
    ③ 1.3/1.5 的 beam arc-cost 差異(`RuleProfile`)自動吃到,不必再各寫一份。

    ### 一個測試前提被正確地推翻了

    `TestAIRaidRepelledByFleetAtStar` 有一段自我守衛(「測試前提不成立:願打門檻 X 已高於
    母星防禦 Y」),改完之後它**觸發了**:兩艘巡洋艦(16)配一座光禿禿的星基,防禦 19,
    而 AI 的願打門檻是 16×125% + 1 = 21 → AI 會贏。

    **那是正確的結果,不是 bug。** 這個測試守的是「把艦隊擺對地方有意義」,不是「星基很強」,
    所以把測試裡的母星升級成戰鬥站(真的有投資防禦的樣子),而不是把模型改回去遷就測試。
    那段自我守衛就是設計來在平衡變動時**要人做決定**的,它正常運作了。

39. **盤點:「最大的系統級缺口」那四條全部已經做掉了**(2026-08-07)。

    第 38 項的教訓是「**別編一個做不到的理由,先 grep**」。這一項把同一把尺對準文件自己:
    Part B 那份「最大的系統級缺口(A 級硬證)」清單,逐條核實之後**四條全中**——全部已完成。

    | 原本的斷言 | 一次 `ls` 就推翻 |
    |---|---|
    | 歷史記錄系統「remake 完全沒有」 | `internal/shell/history.go`(6 函式)+ `infoHistory` 折線圖 |
    | 前哨站「remake 只有殖民地」 | `internal/shell/outpost.go`(9 函式),進存檔、進熱座、可升級成殖民地 |
    | 艙損/維修 | `internal/shell/repair.go`(11 函式),手冊 p.80/82/25 + `Repair_Ships_At_Colonies_` 雙錨定 |
    | (事件系統早已標記完成) | — |

    Part A-2 的 `Smack`(Smacker 過場)同樣是過期的:`cmd/moo2/cutscene.go` + `internal/smk`
    **真的在解 Smacker**,不是靜態圖。

    ### 為什麼這件事重要,而不只是「更新一下文件」

    這四條被後續**每一輪的摘要反覆引用**當成現況,包括我自己寫的。文件裡的斷言一旦成形就會
    被當事實傳遞,而程式碼會往前走、文件不會 —— 於是「還缺什麼」的判斷整個偏掉,
    優先序也跟著偏。這正是 `rulebook/63-truth-in-code-not-stale-markers.md` 講的東西,
    也是 CLAUDE.md 要求「每一輪盤點、清除錯誤斷言」的理由。

    ### 這一輪也驗過、確認**不是**問題的

    - `cmd/moo2/diploview.go` 的 `diploRelationRows` 是寫死的 Klackon/Psilon/Silicoid 三筆。
      看起來像「外交畫面沒接真資料」,實際上那是 `-shot` 旗標的**獨立示範模式**
      (`main.go` 的 `runDiploView`),遊戲內的外交走 `b.diplomacy()`。**不是 bug。**

    ### 核實過後真正還缺的

    寫在 Part B 那一節的表裡(網路多人、`Command_Points` 專屬畫面、星圖 4 層、2 棟建築、
    地表道路與擺放微調)。⚠ 那份清單同樣會過期,**引用前先 grep**。

40. **指揮點數視窗建起來 + 一個「畫面自己打自己臉」的快取陳舊值**(2026-08-07)。

    第 39 項核實後的清單裡,`Command_Points` 專屬畫面是最小的一項。做掉它。

    ### 畫面結構(`Show_Command_Points_Screen_` @ 0x8BAB9,整支只有 30 行)

    ```
    sub_1191CA(&Draw_Mini_Main_Screen_, 1)      ; 背景重繪掛成「迷你星圖」
    sub_11438B(0, 0, 0x27F, 0x1DF, key=0x1B)    ; 整螢幕隱形欄位,ESC 關閉
    sub_128C32(0, 0, 0x27F, 0x1DF, 0)           ; Fill 清畫面
    Draw_Mini_Main_Screen_()                     ; 迷你星圖當底
    Show_Command_Points_(玩家索引)               ; → sub_E2000 組文字 → sub_DDF24 顯示
    ```

    → **迷你星圖當背景 + 一塊文字視窗,ESC / 點擊關閉**。

    ### 內容:符號名就是權威

    文字本身在執行期才載入的字串區塊裡(`sub_DD4FD` 用 `repne scasb` 逐條走),英文原句沒解出來。
    但執行檔**帶著符號表**,那幾條訊息的名字直接說明視窗顯示什麼:

    | 符號 | 欄位 |
    |---|---|
    | `_starting_command_points_msg` | 起始指揮點數 |
    | `_total_command_points_msg` | 指揮點數總計 |
    | `_total_command_points_used_msg` / `_total_command_point_used_msg` | 已使用(原版連單複數都分兩條)|
    | `_command_summary_msg` | 總結 |
    | `_command_points_window_field` | 這個視窗的欄位 |

    ⚠ 所以**結構與欄位組成是原版真值,中文用字是 remake 自己的**;視窗座標也是 remake 排的
    (原版走 `sub_DDF24` 那支泛用訊息視窗,座標是傳進去的,沒有「指揮點數專用」的立即數可抄)。
    ⚠ ESC 那一半沒接:`shell.InputState` 目前只帶滑鼠,加鍵盤要動共用結構。

    ### ⚠ 順帶抓到:畫面把自己的三個數字擺在一起才看得出來的矛盾

    第一版直接讀 `Player.CommandPointsSupply`,畫出來是:

    ```
    起始指揮點數    5
    軌道基地提供    0
    指揮點數總計    1     ← 5 + 0 = 1 ?
    ```

    原因:`Player.CommandPointsSupply` / `UsedCommandPoints` 是**只在 `EndTurn` 更新的快取欄位**,
    開局或剛蓋好星基還沒結算時是舊值。星圖右欄那個淨值數字吃的是同一組欄位,**同樣是舊的**,
    只是它單獨一個數字擺著,沒有旁邊兩行可以對照,所以一直沒被發現。

    修法:`shell.CommandPointsSupplyNow()` / `CommandPointsUsedNow()` 現算,視窗與星圖右欄都改用它。
    護欄 `commandpoints_live_test.go`:把快取欄位設成 −999 也要算得對、剛蓋好星基當下就要反映。

    **這是「把有關聯的數字放在同一個畫面上」本身就是一種驗證**的例子——單獨顯示的數字錯了
    沒人看得出來,三個擺在一起就自己露餡。

41. **殖民地地表的道路畫出來了,順手在原版資料裡撞到兩個位元組級的錯**(2026-08-07)。

    第 31 項留下的「道路沒畫」補完。`cmd/moo2/colonyroads.go`,護欄 `colonyroads_test.go`。

    ### 幾何:道路走格點,建築佔格子

    建築是 6×6 的**格子**,道路是 7×7 的**格點**,路段是格點之間的連線。四個方向的合法範圍
    由 `Load_Road_List_Anims_` @ 0xB5FBE 的跳過條件解出來:

    | dir | 合法範圍 | 幾何 | 段數 | 資產編號 |
    |---|---|---|---|---|
    | 0 | a 0..6、b 0..5 | (a,b)→(a,b+1) | 42 | `a×6 + b` |
    | 1 | a 0..5、b 0..6 | (a,b)→(a+1,b) | 42 | `42 + a×7 + b` |
    | 2 | a,b 0..5 | 格子的對角線 | 36 | `84 + a×6 + b` |
    | 3 | a,b 0..5 | 格子的另一條對角線 | 36 | `120 + a×6 + b` |

    42+42+36+36 = **156**,與 `COLROADS.LBX` 的資產數一模一樣 —— 這個等式就是幾何解對了的
    確認,不必再去量圖。編號是同一支函式裡一個從 0 開始的計數器依 dir → a → b 發下去的。

    ### 產生規則接在建築擺放的同一條亂數流上

    `Build_Road_List_Based_On_Bldg_List_` @ 0xB6099 由 `Make_Bldg_Array_For_Colony_` 在
    `loc_BC5B4` 呼叫,也就是**擺放 → 排序 → 抖動之後**。對每個有建築的格子抽三次 `Random(2)`:
    第一次決定畫左邊還是上邊(必畫一條),第二、三次各以一半機率補右邊、下邊。四條邊正好圍住
    那個格子,所以視覺上是「房子外圍畫框,框缺一到三邊」。

    ### dir 2 / dir 3 是出貨版沒用到的美術

    全執行檔對 `byte_19E57E` / `byte_19E57F`(dir 2、3 的旗標)**只有寫 0**,沒有任何一處寫 1
    (IDA 自己的 DATA XREF 也只記錄 `sub_B6099` 裡那兩個 `mov …, 0`)。所以 COLROADS.LBX 裡
    那 **72 張對角線圖從來不會出現在畫面上**。連帶讓產生器裡「空格子」那一整條分支變成死碼:
    它要求 dir2/dir3 非 0 才往下走,而那永遠不成立 —— 包括它裡面那支蓄水池抽樣
    `sub_B6860` 也永遠選不到候選,一次亂數都不會消耗。remake 因此只實作有建築那條分支,
    這**不是簡化,是把死碼認出來之後不抄**。

    ### 兩個原版資料錯誤(照抄不修)

    - **繪製順序表 `byte_B4D5B` 少一個格點。** 49 組 (b,a) 應該不重不漏、依 a+b 由遠到近遞減。
      實際上索引 7 是 (3,4)(a+b=7,夾在兩個 9 中間),而 **(5,4) 整張表沒出現**、(3,4) 出現兩次。
      把那一個位元組從 `0x03` 改成 `0x05`,49 格就剛好完美 —— 典型的手 key 表打錯一格。
      後果:格點 (5,4) 的路段永遠不畫,(3,4) 疊畫兩次(同一張圖,看不出來)。
    - **包圍判定表 `byte_B4DE9` / `byte_B4DF5` 的 Δa/Δb 對調了兩筆。** 要判斷格子 (a,b) 四條邊
      是否都有路,該查的是 d0@(a,b)、d1@(a,b)、d0@(a+1,b)、d1@(a,b+1);表裡卻寫成
      d0@(a,b+1)、d1@(a+1,b) —— 那是往外延伸的兩段,不是對邊。反正整條分支是死碼,不影響畫面。

    這兩個都**照抄不修**:修掉會讓 remake 比原版「正確」,而驗收標準是與原版一致。
    測試 `TestColonyRoadOrderMatchesOriginalTable` 把這件事釘死,並在失敗訊息裡說明為什麼不能改。

    ### 方法:表不要用手抄

    第一次是照著 IDA 的 `dd` 清單手抄成 Python 再解,結果 49 格只覆蓋 48、還有一處 a+b 不遞減 ——
    分不清是原版有錯還是自己抄錯。改成**直接從 `Orion2.exe` 讀位元組**才定案:

    - 先用「49 組不重不漏 + a+b 遞減」當指紋掃全檔 → **零命中**(所以不是自己解錯位址)。
    - 改用有把握的前 14 個位元組當錨點 → 全檔唯一命中,`0x13A3EF`。
    - 由此得 cseg01 的 VA→檔案位移 delta = `0x85694`,再用 `dword_B4DBD` 應為 `0x01000203`
      獨立交叉驗證通過;而且那個 dword 就緊接在 98 個位元組之後,起點與長度同時釘死。

    最後連 Go 的表字面量也是用腳本從執行檔產生的,不經手抄。

    ### ⚠ 仍是缺口

    - **道路位置不會與原版同一局逐格相同。** 道路吃的是抖動之後的亂數流位置,而 remake
      還沒實作那段抖動,兩邊的流早就錯開。**美術、格點、方向規則、密度都是原版真值,
      走法不是。** 要逐格對上,前置是把抖動補完。
    - **植被層整層沒做(新發現)。** 道路之後原版還跑 `sub_B6977`:每個**空**格子最多放 2 株
      `COLVEGGI.LBX` 的植物,密度看周圍道路數(路少的地方草多),外觀與像素抖動都現抽。
      這是原版有、remake 沒有的一整層地表細節,前置是 `sub_B6647`(依氣候選植物)與
      `sub_BC866`(格子 → 螢幕座標)。

42. **房屋抖動補完 + 母星的國會大廈一直沒畫**(2026-08-07,接第 41 項)。

    第 41 項說「道路位置對不上原版,因為抖動沒補」。這一項把抖動補上,順帶發現一件更基本的事:
    **remake 的地表格陣從第一步就少放了一棟建築。**

    ### 抖動:`Make_Bldg_Array_For_Colony_` 的 `loc_BC441`..`loc_BC560`

    排序把同類建築聚成整齊區塊,這一段再把房屋隨機挪一下。跑 8 輪,第 n 輪處理
    `houseStyles[n%4]`(表 `off_BA77C` = 3/14/40/41,與既有 `colonyHouseStyles` 逐項相同):
    找到那種房屋的格子,在 3×3 的偏移迴圈裡試著和另一格互換,**換成一次就結束該輪**。
    機率:目標格非空 → 抽兩次 `Random(3)`,任一次中 1 就換;空的 → 只抽一次。

    ### ⚠ 原版這段有兩個 bug(照抄不修)

    反組譯的關鍵三行(已用 `objdump` 對原始位元組獨立驗過,IDA 清單無誤):

    ```
    eax = var_14 - si ; if (eax < 0) eax = 0        ; 第一個座標:有夾到 0
    edx = var_18 - di ; test edx, edx               ; 第二個座標:算了、測了號誌…
    cmp ax, 5 ; jle → edx = ax ; else edx = 5       ; …然後 edx 被無條件蓋掉
    ```

    - **第二個座標整個被丟棄**:對應的 `jge / xor edx, edx` 沒有被編出來,而下一條指令兩條
      路徑都覆寫 `edx`。於是目標格的兩個座標都等於 `clamp(a - si, 0, 5)` ——
      **換位對象永遠落在 6×6 格陣的主對角線上**。
    - **內圈的 `di` 因此完全沒有作用**,三次迭代做一模一樣的事。remake 仍保留那個迴圈:
      它決定 `Random` 被抽幾次,而道路接在同一條流上。

    也就是說原版的「隨機微調」實際效果是**把房屋往對角線上搬**。測試
    `TestJitterColonyHousesTargetsDiagonalOnly` 專門防止有人把 `target := v*6 + v`
    當成打錯而「修好」——改成 (a-si, b-di) 之後畫面依然合理,但那就不是原版的畫面了。

    ### `Get_Bldg_CR_` @ 0xBBD37:找一棟建築會吃亂數

    掃描序 a 外 b 內,**命中就立刻回傳第一格**;但命中之前每碰到一個空格都會抽一次
    `Random(n)` 做蓄水池抽樣(原版用 `id == -1` 這條路徑找空位,順手共用了同一支函式)。
    所以「找建築」這個看起來唯讀的動作**會消耗亂數,而且消耗幾次取決於資料**。
    抖動之後、道路之前原版還無條件呼叫一次 `Get_Bldg_CR_(9)`,那次掃描同樣要照抄
    (後續那段交換由全域 `dword_182B19` gate 住,它初值 0 且唯一寫入點在 gate 內側,
    remake 沒有對應概念,**交換沒模擬**)。

    ### 順著編號 9 抓到的缺口:母星的國會大廈沒畫

    編號 9 = **Capitol**。第 36 項把它判成「建立殖民地時自動給予,不可建造 —— 正確地不在表裡」,
    對**建造選單**是對的,對**地表格陣**是錯的:它是一棟實體建築,佔一格、有美術
    (`bldg0.lbx` 資產 `8×36 + 格`)、會被畫出來。**「不在建造表裡」與「不在地表上」是兩件事**,
    先前混為一談,於是母星的地表一直少一棟。

    只有母星有,patch 1.5 手冊三處佐證:「最多建築數(**不計 Capitol**)是 ⅔ 人口無條件進位」、
    「**沒有** Capitol 的士氣懲罰可依政府設定」、「Colony Base 若加進 `initial_buildings`,
    會發給**每個玩家的母星**」。

    ### ⚠ 仍不能宣稱與原版逐格相同

    亂數流的**結構**接完整了(每一步消耗幾次都照抄,護欄逐項釘住),剩下的差異在**建築集合**:
    原版走殖民地的 49 個「有沒有這棟」旗標,裡面有 remake 沒建模的項目(Colony Base、
    已完成的一次性改造等)。集合差一棟,蓄水池抽樣從第一步就錯開。
    **「同一顆星會不會長一模一樣」還沒對原版實測過,不要當成已驗證。**

43. **殖民地地表的植被層(第 41 項發現的缺口,同日補上)**(2026-08-07)。

    `cmd/moo2/colonyveggie.go`,護欄 `colonyveggie_test.go`。

    ### 13 個群組 × 8 張 = 104

    `Pick_Random_Veggie_Anim_Entry_For_Colony_CR_` @ 0xB6647:

    ```
    資產 = 群組×8 + max(Random(8) − 1 − (a+b)/2, 0)
    ```

    後面那項就是**透視**:格點越遠(a+b 越大)越容易被壓到 0,也就是同組裡最小的那株。
    群組由氣候決定(`sub_B66D0`,一張 10 路跳表),最大群組 12 → 13 組。
    **13 × 8 = 104**,與 `COLVEGGI.LBX` 的資產數一模一樣(這次直接跑 `lbxinfo` 驗,不引用舊文件)。
    再交叉一項:群組 0 的前四張是 6×15、8×15、9×22、9×22 —— **組內編號越大株越大**,
    與那個透視項的方向一致。

    | 氣候 | 群組 |
    |---|---|
    | 0 有毒 | `Random(2)`:1→5、2→9 |
    | 1 輻射 / 2 荒蕪 | `Random(2)`:1→6、2→4 |
    | 3 沙漠 | 8 |
    | 4 苔原 | 4(預設) |
    | 5 海洋 | `Random(4)`:1→7、2→9、3→11、4→4 |
    | 6 沼澤 | 10 |
    | 7 乾燥 | 1 |
    | 8 類地 / 9 蓋亞 | `Random(4)`:1→12、2→3、3→0、4→2 |

    ### 密度:道路越多反而越容易長

    `sub_B6977` 只處理空格子,先數該格四條邊上有幾段路,然後:

    ```
    r = Random(7)
    if (r − 2) < 道路數  → 長
    else if 道路數 != 0  → 不長
    else                  → Random(建築數 + 2),回傳恆 ≥ 1 所以一定長
    ```

    0 條路 → 必長;k 條路(k>0)→ 機率 (k+1)/7。⚠ 最後那個 `Random(建築數+2)` 的結果
    **永遠不會是 0**(原版 `Random(n)` 回 1..n),判斷等於恆真——看起來想寫「建築越多越不長草」
    卻沒生效。照抄:判斷保留(要消耗那次亂數),結果一樣是「長」。

    每格最多 2 株,每株各抽一次 `Random(3)` 中 1 才長。

    ### 位置與繪製

    ```
    x = 格心x + Random(寬) − 寬/2
    y = 格心y + Random(高) − 高/4      ← 是高/4 不是高/2,植物根部落在格心稍下方
    ```

    格心用 `Bldg_Coords_To_Centered_Screen_Coord_` @ 0xBC866 = 角落 (a,b) 與 (a+1,b+1) 取平均,
    **remake 的 `colonyCellCenter` 已經是同一個算式**,直接用。

    繪製在 `Draw_Colony_Bldgs_` @ 0xBEBDC 裡:對 36 格的遠→近順序,**每一格先畫植被再畫建築**,
    不是「先畫完所有植被再畫所有建築」。差別在遮擋。

    ### 這裡也踩到那張對調的表

    密度用的 `sub_B67E4` 就是第 41 項提到「Δa/Δb 對調了兩筆」的那張表。差別是:
    **道路那條分支是死碼,這條是活的** —— 表錯不錯會真的影響畫面。
    `TestColonyRoadEdges4UsesOriginalTable` 用「格子的四條真邊只會數到 2」把它釘住。

    ### ⚠ 沒模擬的部分

    - `sub_B6B95` 的外圈次數由呼叫端的 `bl = (沒有格子被選取)` 決定。正常畫面 `bl = 1` →
      跑一圈走 `sub_C5D55`;有格子被選取時 `bl = 0` → **一株都不畫**。
      remake 沒有「選取格子」這個狀態,固定走正常路徑;另一條 `sub_C5D75` 的差別沒有追。
    - 每株的**顏色**沒有對原版逐張比對過(調色盤鏈與道路同源——原版兩者都從
      `dword_193174` 取,而道路的顏色是對的)。

    ### 順帶:一個效能坑

    尺寸是在**產生**階段就要用的(它進位置公式),而地表是每幀重算的,而 `decodeAsset`
    自己沒有快取 —— 不處理就變成每幀重解最多 72 張 LBX。加了 `colVegSizeCache`。

44. **星雲:星圖 4 層裡的第 1 層,而且是有規則的地形不是裝飾**(2026-08-07)。

    `internal/shell/nebula.go`(規則)+ `cmd/moo2/nebula.go`(圖與判定),
    護欄 `nebula_test.go`。第 32 項列的星圖 4 層,做掉第 1 層。

    ### 星雲是地形

    GAME_MANUAL.pdf 星圖那一節:

    > Ships traveling through a nebula are reduced in speed to 1 parsec per turn.
    > More importantly, the fierce ionization prevents deflector shields from
    > functioning without Hard Shields technology.

    戰鬥那一節(p.158)再講一次:「if combat takes place in a nebula, all shields become
    inoperative, except for those on ships equipped with Hard Shields.」

    ### 判定:遮罩像素 > 5(反組譯與手冊互相印證)

    `Point_Is_In_Nebula_N_` @ 0xEB9C8 拿 `(點 − 星雲原點) / 3` 去索引星雲圖的**調色盤索引值**,
    `> 5` 才算在裡面。patch 1.5 手冊逐字寫了同一件事:「a star is considered "in nebula" if the
    respective pixel value of the nebula picture is greater than 5」,還補充「deep in a Nebula a
    few dark pixels can be present causing a star at such a location to be considered "not in
    nebula"」—— **這個判定本身就有小破洞,那是原版行為**。

    每顆星的結果存在星球結構 +0x6F(`Initialize_Star_In_Nebula_Info_` @ 0xEBA96),
    也就是 `internal/save` 既有的 `Star.InNebula` 欄位。

    ### 數量與圖

    `Generate_Number_Of_Nebulas_` @ 0x8C4D3 是四路跳表:小 `Random(2)−1`、中 `Random(2)`、
    大 `Random(3)`、巨 `Random(3)+1`。上限 4 與 `internal/save` 從存檔格式反推的
    `maxNebulas = 4` 一致 —— 兩邊獨立對上。

    圖在 **STARBG.LBX**(和星空層同檔):0..5 是 640×480 的星空層,6 起每 4 張一組
    (一組 = 同一團的 4 個縮放等級),全檔 54 張 → (54−6)/4 = **12 種**。

    ### 順帶:激活了兩段構不到的碼

    `gamedata.DamageHardShieldBonus`(手冊「額外減傷 3」)先前**沒有元件載體**,等於死碼。
    這一輪把「硬化護盾」加進 `SpecialOptions`,掛在與隱形裝置同一個研究主題
    (`TOPIC_DISTORTION_FIELDS`,techtree.go 的三選一,`TECH_HARD_SHIELDS`)。
    同時戰術戰鬥的還擊路徑先前寫死 `hardShield = false`,一併接上。

    ### ⚠ 兩個踩過的錯,都寫進測試

    - **檔位換算**:第一版自己拿原版四檔的星數(20/36/54/71)取中點當界線。
      但 remake 的 `GalaxySizes`(12/24/36/48)**本身就是那四檔**,結果「中型」被判成檔位 0
      (星雲數有一半機率是 0)、「巨型」被判成檔位 2。徵狀是**開局星圖常常一團星雲都沒有**,
      而那看起來完全合理。改成直接查 `GalaxySizes` 取索引,
      `TestGalaxySizeClassMapsGameOptions` 釘住。
    - **調色盤鏈**:第一版沿用殖民地畫面的鏈(含殖民地框架盤),整團星雲畫成**鮮紅色**。
      STARBG 沒有內嵌調色盤、顏色全由鏈決定,星圖該用的是星空層那條 `buffer0.lbx#0`。

    找這兩個都不是靠讀碼,是**加一行 `println` 量出來的**:第一次量到
    `drawNebulae n= 0`(所以問題在產生不在繪製),修完再量才看到紅色那張。

    ### ⚠ 沒做的:移動懲罰

    「reduced in speed to 1 parsec per turn」需要一個「原本幾 parsec/turn」的基準,
    而 remake 的星圖移動是 `ETA = ceil(距離 × 8)` 這種**沒有單艦速度模型**的算法,
    沒有可換算的基準,硬套就是自己編一個倍率。**留著不做**。
    領袖技能 Navigator(「ignore the movement restrictions caused by nebulae and black holes」)
    與 Warp Field Interdictor 建築(「人造星雲場,半徑 3 秒差距」)都卡在同一個前置上。

    ### ⚠ 位置不是原版真值

    原版星雲座標在銀河座標系(遮罩是它的 1/3 解析度),remake 的星圖是自有模型、座標正規化,
    兩邊接不起來(與蟲洞同一個情況)。這裡在正規化空間隨機擺並避開母星。

45. **星圖移動換成秒差距模型:四條手冊規則從「無處可掛」變成可實作**(2026-08-07)。

    `internal/gamedata/starlane.go`(數值)+ `internal/shell/starlane.go`(接線),
    護欄 `starlane_test.go`。第 44 項留下的移動懲罰,前置補完。

    ### 為什麼卡住

    先前的星圖移動是 `ETA = ceil(正規化距離 × 8)` —— 一個**沒有速度概念**的固定換算。
    手冊有四條規則都以「秒差距/回合」表述,全都無處可掛:

    | 規則 | 手冊 |
    |---|---|
    | 星雲 | reduced in speed to 1 parsec per turn |
    | 黑洞 | No ship can safely pass within 2 parsecs of a black hole (unless … Navigator) |
    | Navigator | increases the speed of the fleet by 1 or 2 parsecs per turn |
    | Warp Field Interdictor | radius of 3 full parsecs … slows all enemy ships … to 1 parsec per turn |

    ### 三個真值把換算釘死

    **1 秒差距 = 30 個遊戲座標單位。** `Parsecs_Between_Points_` @ 0xEBE79 整支就是
    「回傳最小的 p 使得 `p² × 900 ≥ dx² + dy²`」,`900 = 30²` 是立即數。
    順帶得到:原版的秒差距距離是**整數,無條件進位**。

    **四檔銀河尺寸。** `sub_1693B6` 的 4 路跳表逐檔寫死:

    | 檔位 | 遊戲單位 | 秒差距 | SizeFactor |
    |---|---|---|---|
    | 0 小 | 506 × 400 | 16.9 × 13.3 | 10 |
    | 1 中 | 759 × 600 | 25.3 × 20.0 | 15 |
    | 2 大 | 1012 × 800 | 33.7 × 26.7 | 20 |
    | 3 巨 | 1518 × 1200 | 50.6 × 40.0 | 30 |

    三重交叉驗證同時成立:寬恆為 `SizeFactor × 50.6`、高恆為 `SizeFactor × 40`;
    而原版存檔 **SAVE10.GAM 讀出來就是 759 × 600 / SizeFactor 15**(檔位 1)。
    順帶量到那局的最近鄰距離:最小 2.25、平均 2.95、最大 4.24 秒差距 ——
    典型一跳約 3 秒差距,核融引擎(2 秒差距/回合)要 2 回合,合理。

    **星數 → 檔位。** `Galaxy_Size_From_N_Stars_` @ 0x798D2 的門檻是 **20 / 36 / 54 / 72**。

    ### 順帶修掉一個失真:遊戲提供的星系大小本來就不對

    remake 的 `GalaxySizes` 先前是 12/24/36/48(自訂),與原版四檔對不上。
    而**星雲數、銀河跨距這些表都是以檔位為索引的**,對不上就整串偏掉 ——
    第 44 項那個「開局常常一團星雲都沒有」就是這麼來的。改成 20/36/54/72。

    ### 引擎速度(手冊逐條)

    核融 2 / 融合 3 / 離子 4 / 反物質 5 / 超空間 6 / 相位 7 秒差距每回合。
    手冊每一條都補了同一句「This drive is added to all your ships … as soon as you complete
    your research」——**引擎是全帝國自動升級,不是單艦掛載的元件**,所以只看已研究的最高階。

    ### ⚠ 又一個「畫面上看不出來」的坑

    `FleetHasFTL` 對非曲速前開局**直接回 true、不看科技表**。於是那些開局的引擎階查出來是 0
    → 航速 0 → **ETA 全被夾成 1**。整個秒差距模型形同虛設,而畫面上只是「每一趟都 1 回合到」,
    看起來像船很快而已。修法是給一個原則性下界:**有 FTL 就至少是核融引擎**
    (手冊:「the slowest of the FTL propulsion systems」)。
    `TestFleetSpeedFallsBackToNuclearWhenFTL` 釘住。

    ### ⚠ 近似與未做

    - **「穿越星雲」近似成起點或終點在星雲內。** 原版的艦隊沿路徑逐段前進,remake 是兩點直接
      算 ETA,沒有路徑。差別出現在「兩端都在雲外、直線穿過一團星雲」——原版降速、remake 不會。
      要逐段判定得先有路徑模型。
    - Navigator 的「+1 或 +2」手冊沒寫判準,取 **+1**(保守下界,不是真值)。
    - **黑洞 2 秒差距禁行**與 **Warp Field Interdictor 3 秒差距干擾場**的常數已入表
      (`BlackHoleAvoidParsecs` / `InterdictorRadiusParsecs`),但**還沒接進派遣判定**——
      兩者都需要「路徑經過哪些星」這個同一個前置。

46. **星圖航線模型:三條規則的共同前置補完**(2026-08-07)。

    `internal/shell/route.go`,護欄 `route_test.go`。第 45 項留下的三項全部接上。

    ### 三條規則其實是同一個問題

    黑洞、干擾場、穿越星雲——三條的形狀都是「**這條航線離某個東西多近**」或
    「**這條航線有沒有穿過某塊區域**」。先前的星圖是「兩點直接算 ETA」,沒有「沿路」這個概念,
    所以三條一起卡住。一個線段模型全解。

    | 規則 | 判定 | 手冊 |
    |---|---|---|
    | 黑洞 | 線段到黑洞星 < 2 秒差距 → 拒絕派遣 | No ship can safely pass within 2 parsecs … unless … Navigator |
    | 干擾場 | 線段到敵方干擾器星 ≤ 3 秒差距 → 航速 1 | radius of 3 full parsecs … slows all enemy ships … to 1 parsec per turn |
    | 星雲 | 沿線取樣打到遮罩 → 航速 1 | Ships traveling **through** a nebula … |

    「through」那個字是重點:**兩端都在雲外、直線穿過去也算**。第 45 項那個「只看起訖點」
    的近似就是漏在這裡,現在是沿線取樣(每秒差距 4 點)的正解。

    ### 幾個刻意的判定細節

    - **線段不是直線**:目的星**之外**的延長線上有黑洞,不該擋住這趟航程。
    - **起訖點豁免**:目的地本身是黑洞是玩家自己選的,擋的是「路過」。
    - **干擾場不給 Navigator 豁免**:手冊那句豁免只寫「nebulae and black holes」——
      干擾場是人造的,不在那句涵蓋範圍。
    - **星雲的判定式改成探針**(`SetNebulaProbe`):沿線取樣需要「任意點在不在星雲內」,
      不只是「哪顆星在星雲內」。判定要讀遮罩,而規則層不碰資產,所以由 cmd/moo2 裝進來。
      ⚠ 未匯出欄位 = **不進存檔**,開新局與讀檔後都要重裝(兩處都補上了)。

    ### 實測可達性(不是斷言,是量的)

    每種銀河大小各跑 12 局,從母星出發:

    | 星數 | 黑洞/局 | 被擋的目的地 | ETA(核融引擎) |
    |---:|---:|---:|---|
    | 20 | 0.5 | 5.7% | 2..10 回合 |
    | 36 | 1.1 | 7.9% | 1..15 |
    | 54 | 1.7 | 13.7% | 1..19 |
    | 72 | 2.3 | 11.3% | 2..30 |

    86% 以上的目的地照樣可達,而且蟲洞不受黑洞限制。ETA 隨銀河大小拉開到 30 回合是
    **最慢引擎**的數字——換到相位引擎(7 秒差距/回合)同一趟約 9 回合,
    「研究更好的引擎」因此第一次有了實際意義。

    ### 順帶:第三次同一類效能坑

    沿線取樣會呼叫遮罩判定上百次,而 `decodeAsset` 沒有快取 —— 每次派遣重解上百次 LBX。
    加了 `nebMaskCache`。**這是第三次踩同一個形狀的坑**(前兩次是殖民地地表每幀重算、
    植被尺寸每幀重解),共同成因是 `decodeAsset` 本身無快取,而呼叫端各自為政。


47. **兩種星門:規則接進秒差距模型,標記上星圖**(2026-08-07)。

    星圖 4 層做掉第 3 層。`internal/gamedata/starlane.go` + `internal/shell/starlane.go`
    (規則)、`cmd/moo2/gate.go`(標記),護欄在 `starlane_test.go`。

    ### 兩條規則(手冊逐字)

    | 科技 | 效果 | 手冊 |
    |---|---|---|
    | 躍遷門 Jump Gate | 自己的殖民地之間 **+3 秒差距/回合** | increases the speed of your ships traveling between two of your colony systems by 3 parsecs a turn |
    | 星際之門 Star Gate | 自己的系統之間 **一回合到** | allows instantaneous (1 turn) travel between any two of your systems |

    兩者都是 **Achievement 科技**:研究到就在自己每個有殖民地的星系各生一個門
    (「forms a … wormhole terminus in **each** system in which you have … a colony」),
    不必逐星建造——所以判定只有「有沒有科技」+「這顆星是不是自己的殖民地」。

    ### 兩個順序上的決定

    - **躍遷門的加成放在懲罰之前**:星雲與干擾場都是「reduced **to** 1」——是覆寫不是相減,
      所以它們仍然贏。研究了躍遷門也不會讓你在星雲裡跑得比較快。
    - **星際之門在最前面**:它形成的是穩定的蟲洞終端、不走實空間,所以沿路的星雲/干擾場
      懲罰都不適用。與 remake 既有的蟲洞同一個語意。

    ### 這一項能做的原因

    第 45 項建了秒差距與航速、第 46 項建了航線——**先前沒有「秒差距/回合」這個量,
    這兩條規則根本寫不出來**。這是三項連著做下來的收成,不是獨立的一項。

    ### ⚠ 標記不是原版畫法

    原版 `Draw_A_Gate_Icon_` @ 0x83741 是一支 330 行的逐格動畫,資產來源不是字串常數、
    還沒追出來。這裡先用雙環把**資訊**呈現出來——看不出「這顆星有門」的話,
    那兩條速度規則等於隱形。追到資產再換成原版動畫。


48. **拓殖基地:國會大廈那個坑的另一半**(2026-08-07,task #46 的可實作部分)。

    第 42 項補上母星的國會大廈(編號 9)時,只治了一半。**編號 11 Colony Base 是完全對稱的
    另一半**:母星有國會大廈,其餘殖民地有拓殖基地,兩者都是拓殖時自動給予、不可建造的
    **實體建築**——佔一格、有美術、會被畫出來。

    第 36 項那張差集表對它的註記本來就寫著「拓殖時自動」,所以資料一直都在,
    只是同樣被「不在建造表裡」這個判斷連帶漏掉了。**「不在建造表」與「不在地表」是兩件事**
    —— 同一句話這是第二次寫,因為同一個判斷失誤造成了兩個缺口。

    護欄 `TestColonySurfacePlanNonHomeworldHasColonyBase` 除了驗有無,還驗
    **「每個殖民地恰好有國會大廈與拓殖基地其中一棟」** —— 兩者同時出現或同時缺席都會被抓到。

    ### task #46 剩下的部分

    - **一次性改造(17 Gaia / 37 Soil Enrichment / 44 Terraforming)** 是否在完成後仍佔一格,
      取決於原版的 `byte[colony + 0x136 + id]` 旗標完成後是否保留。**沒有查證,所以沒做**——
      remake 這三項只改氣候、不記錄「這個殖民地做過」,要補得先加狀態。
    - **對原版實測落點**仍未做(需要 archive.org 線上原版逐畫面對照)。


49. **銀河貨幣交易所:它根本不是建築**(2026-08-07)。

    第 36 項留下的兩個「完全沒有」的編號,解掉一個。

    ### 手冊直接寫了效果,而且寫了分類

    > Galactic Currency Exchange (**Achievement**) A galaxy-wide, central currency exchange …
    > **increases the income generated by all colonies (from all sources) by 50%**.

    「Achievement」與躍遷門/星際之門同一個標記——**研究完成即生效,不必建造**。
    所以 remake 把它接在科技擁有狀況上(`PlayerState.HasGalacticCurrencyExchange`),
    走的是與政府 money 加成同一層(帝國層級、逐殖民地迴圈之外),因為手冊的字是
    「all colonies (from all sources)」——整體收入的乘數,不是逐殖民地的建築加成。

    ### ⚠ 為什麼卡了這麼久:自訂的推論規則蓋過了一手來源

    第 36 項抽建築表時發現「維護費 0 = 一次性」這個規律,拿它把 8 個編號分成
    「該進 `Buildings`」與「該進 `SpecialActions`」兩堆。編號 18 有 250 PP 成本、3 BC 維護,
    於是被判成**常駐建築**,然後「效果是什麼」就查不到了——因為手冊的建築清單裡本來就沒有它。

    真正的位置在手冊的**科技說明**那一節,標記是 Achievement。
    **一手來源(手冊寫效果與分類)贏自訂的推論規則(維護費啟發式)。**
    那條啟發式對其餘 7 個編號仍然成立,錯的是把它當成充分條件。

    這也是為什麼原版建築表裡它有成本與維護費卻不可建造:那張表是 49 個編號的**通用結構**,
    不是「可建造建築清單」。同一張表裡的 Capitol(9)與 Colony Base(11)也一樣
    (見第 42、48 項)——三個編號、同一個誤判形狀。

    ### 剩下的:48 Artificial Planet(⚠ 本節的第一版寫錯,見第 51 項)


50. **AI 主動請求會談 + 星圖上緣的請求燈**(2026-08-07)。

    星圖 4 層做掉第 4 層。`internal/shell/audience.go`(狀態)+ `cmd/moo2/audience.go`(燈),
    護欄 `audience_test.go` 兩支。

    ### 卡的不是繪圖,是「誰在請求」根本不存在

    remake 的外交先前**只有玩家主動**:點進外交畫面提議,AI 只會回應。原版不是這樣——
    AI 會來敲門,星圖上方就亮起一排燈。這一層畫不出來,是因為狀態沒有。

    ### 表示法與版面是真值

    `Humans_Requesting_Diplomacy_` @ 0xFA795 整支只有 `mov al, byte_1AB054; retn` ——
    **一個位元遮罩,每位對手一個 bit**。版面在 `Draw_Diplomacy_Request_Lights_`:

    ```
    x = 0x1FA − 已畫個數 × 圖寬      ; 506,由右往左
    y = 5                            ; 貼星圖上緣
    ```

    兩個都是立即數。`TestAudienceLightLayoutMatchesOriginal` 直接釘住。

    ### ⚠ 觸發條件沒照抄(說明理由)

    原版設那個 bit 的地方在 `sub_F5A9F` —— 一支約 30 路跳表的 AI 行動分派函式,觸發散在
    各 case 裡。追出完整條件成本高、收穫有限,**所以沒抄**。

    remake 改接在既有的 AI 模型上:**態勢改變時來敲門**。`ai.DecideStance` 的五級裡有三級
    本身就是「要跟你講話」的語意(宣戰、提議貿易、提議結盟),態勢從 A 變成 B 正是原版 AI
    會主動聯絡玩家的時機。**這是接在既有模型上的推導,沒有引入任何新的門檻值。**

    中立/敵視不敲門:前者沒事,後者是態度不是提案。

    ### 順帶被測試逼出來的一個設計修正

    第一版把來意直接寫成中文(「宣戰」「提議貿易」),被 `TestEnglishModeGapDoesNotGrow`
    這支棘輪測試擋下來(英文模式缺口 26 條 > 上限 16)。那不只是翻譯問題——
    **規則層不該吐顯示字串**。改成代碼(`war`/`trade`/`alliance`),顯示文字留在 UI 層。
    既有的 `stanceNames` 是中文,那是先前留下的;新欄位不再擴散這個作法。

    ### ⚠ 圖不是原版的

    原版每盞燈是該種族的逐格動畫(幀號存在 `byte_19C148[種族]`),指標陣列由別處填,
    資產來源沒追。這裡用「來意色塊 + 一個字」呈現——玩家要看得出**誰在敲門、為什麼**,
    那才是這一層的作用。追到資產再換。


51. **訂正:「手冊全文搜尋零命中」是假陰性——PDF 的連字騙了我**(2026-08-07)。

    第 49 項結尾寫了:

    > 剩下的:48 Artificial Planet。手冊全文搜尋 **零命中**(不是漏查)。

    **那句是錯的。** 手冊裡有,而且寫得很完整。搜不到的原因是這本 PDF 用**連字**排版——
    `artificial` 實際上是 `arti` + `ﬁ`(U+FB01)+ `cial`,搜 `Artificial` 當然零命中。
    改搜小寫 `asteroid` 立刻命中,同一段就把規則講完了。

    ### 真正的規則(手冊逐字)

    > (Special) This technology allows a colony in the same system with an asteroid field or
    > gas giant to assemble this otherwise useless planetary material into a complete
    > artificial planet that can support a colony. This planet is **Barren, Normal G, and
    > mineral Abundant**. **Gas giants make Huge worlds, and asteroid belts make Large ones**.

    分類是 **(Special)** —— 與地形改造/Gaia 轉化同類的一次性殖民地行動,
    這也和它在原版建築表裡維護費 0 對得上(第 36 項那條啟發式在**這一筆**是對的)。

    手冊裡的全名是「Artiﬁcial Planet Construction (Special)」,對應 remake 科技樹的
    `TECH_PLANET_CONSTRUCTION`(TOPIC_ADVANCED_MANUFACTURING 三選一)——手冊裡它緊接在
    Automated Repair Unit 之後,而那正是同一個主題的另一個選項,兩邊的相鄰關係對得上。

    ### 而且 remake 自己的註解早就寫著

    `internal/shell/outpost.go` 的「尚未建模、誠實留白」那段裡有一條:

    > 手冊 p.50 提到有科技能把氣態巨星/小行星帶的前哨站升級成可住人殖民地
    > (行星固定 Barren/Normal-G/Abundant,氣態巨星化為 Huge、小行星帶化為 Large)

    **同一條規則、同樣的數值,早就在 repo 裡。** 兩個獨立來源(手冊 + 自己的舊註解)一致。

    ### 真正的阻塞不是「效果不明」,是資料模型

    remake 的 `Stars[i] ↔ Planets[i]` 是**一對一**(UI/拓殖/AI 全依賴這個對齊,見
    `Planet.SystemBodies` 的欄位註解)。而人造行星按定義是「在已經有殖民地的星系裡**再多**
    一顆世界」——轉換 `SystemBody` 之後沒有地方能放第二個殖民地,做出來是空的。

    所以這一項的阻塞從「效果不明」訂正為「**卡在一星一行星模型**」,與遷移連線
    (卡單一艦隊模型)同一類。前置是多行星殖民地。

    ### 教訓

    「查詢回空」不等於「不存在」。這次的假陰性來源是**排版連字**,是文字擷取類查詢的典型坑
    ——而我當時還特地加註「(不是漏查)」,反而把假陰性寫成了確信。
    下次對 PDF 做全文否定判斷前,至少要用**小寫、部分字根**再掃一次當正對照。


52. **查證:一次性改造完成後不佔地表格子**(2026-08-07,task #46 的最後一個未查證項)。

    第 48 項留下的問題:「一次性改造(17 Gaia / 37 Soil Enrichment / 44 Terraforming)
    完成後是否仍佔地表一格,取決於原版 `byte[colony + 0x136 + id]` 旗標完成後是否保留。
    **沒有查證,所以沒做**。」

    ### 查證方式:找旗標的寫入點,不是找建築的處理碼

    第一版的想法是「去找地形改造完成的那段碼」——但符號表裡沒有 terraform / gaia / soil
    任何一個名字。**改成從旗標本身下手**:`grep "136h], 1"` 與 `"136h], 0"`,
    全檔只有少數幾處,一眼就看到建築完工結算 `sub_13FD9` 那一處。

    ### 結論(定性的,不是機率的)

    `sub_13FD9` 裡「把旗標記進 `byte[colony + 0x136 + id]`」是**有條件**的:

    ```
    mov [ebp+var_8], 1          ; 函式開頭:預設「要記旗標」
    …                            ; 依建築編號分派
    cmp [ebp+var_8], 0
    jz  短路                     ; 被清成 0 就不記
    mov byte ptr [ebx+eax+136h], 1
    ```

    而 `var_8` 在整支函式裡**恰好被清成 0 四次**。那四個分支做的事一看就認得:

    | 分支 | 做的事 | 是誰 |
    |---|---|---|
    | `loc_141FE` | 氣候 8(Terran)→ 9(Gaia) | Gaia 轉化(17) |
    | `def_1423F` | 跳表改 `byte[planet+8]` 氣候 | 地形改造(44) |
    | `loc_142D7` | `inc byte[planet+0Bh]`、`or byte[planet+10h], 2` | 土壤改良(37) |
    | `loc_14426` | 寫入整組行星欄位(型別/大小/氣候/礦產) | 人造行星(48) |

    **四個分支、四個一次性編號,一一對應。** 旗標既然沒被設起來,
    `Make_Bldg_Array_For_Colony_` 那個讀旗標陣列的迴圈就不會擺它們。

    ### remake 天然就是對的,但把它釘住

    SpecialActions 不在 `gamedata.Buildings` 裡,`origBuildingID` 因此查不到它們——
    現況正確。加了 `TestColonySurfacePlanExcludesOneShotTransformations` 把這個「天然正確」
    釘住:哪天有人「順手」把一次性項目加進建築表,地表就會冒出四棟原版沒有的房子,
    而那看起來完全合理。

    ### task #46 結案

    - ✅ 母星國會大廈(第 42 項)、非母星拓殖基地(第 48 項)
    - ✅ 一次性改造不佔格(本項,**查證後確認不需要改**)
    - 剩下的「對原版實測落點」需要 archive.org 線上原版逐畫面對照,是**驗證工作**不是實作工作,
      移出這一項另計。

---

53. **黑洞的旋渦動畫:規則完整就做,不完整就不做**(2026-08-07)。

    星圖上的黑洞從第 34 項起就用對了圖(BUFFER0#187),但它是**靜止**的。
    原版不是——黑洞會轉。

    ### 推進規則整段可讀

    `Draw_Black_Holes_` @ 0x83BF9 對每個黑洞做兩件事:

    ```
    計數 = (_black_hole_anim_count[黑洞序號] + 1) % (幀數 × 2)
    幀號 = 計數 / 2
    ```

    也就是**每一幀停留 2 次重畫**,而且**每個黑洞各有獨立的計數器**(序號夾在 0..10)。

    那個「除以 2」不是單點讀出來的:一般星球的 `Draw_A_Star_` 裡也有同一個動作
    (`sar eax, 1`)。**兩支函式各自獨立出現同一個比例**,這是兩個來源。

    ### 資產面對得上

    `lbxinfo BUFFER0.LBX` 給的是硬事實:

    | 資產 | 尺寸 | 幀數 |
    |---|---|---|
    | 148..183(6 光譜 × 6 縮放的一般星球) | 33/29/25/23/21/17 | 各 **5** |
    | 184..187(黑洞,4 個縮放) | 40/34/34/25 | 各 **16** |

    黑洞 16 幀、一般星球 5 幀——兩種動畫本來就不是同一件事,和「黑洞在原版是星球以外的
    地圖物件、走自己的迴圈」這個結構一致。

    把 16 幀 dump 出來逐張比對,**16 張全不同**(md5 去重後仍 16 個),旋渦逐幀轉。
    不是 16 份同一張圖。

    ### ⚠ 一般星球**沒做**,而且這是刻意的

    `Draw_A_Star_` 的閃爍是**爆發式**的:計數器跑到 `star[+0x65]` 就停成 -1,
    另外還有一個全域併發預算 `word_19C164` 在管同時幾顆在閃。三個東西沒追出來——
    什麼時候開始閃、爆發長度、預算值。

    **不編那三個數**,所以一般星球維持第 0 幀。黑洞不一樣:它的動畫無條件連續,規則是完整的。
    這一項的分界線不是「難不難做」,是「規則有沒有解完」。

    ### ⚠ 絕對速度是 remake 的選擇,只有比例是真值

    remake 把「一次重畫」對應成「一個 ebiten 幀」。原版的主畫面重畫由它自己的迴圈驅動,
    頻率沒有解出來,所以**動畫跑多快是 remake 定的**;
    **「2 次重畫換 1 幀」這個比例才是原版真值**。

    ### 順帶修掉一個會靜悄悄壞掉的快取

    `starSpriteImage` 原本只解第 0 幀,快取 key 是 `lbx:資產`。一加上幀號,
    那個 key 就會讓 16 幀全部命中同一張——**畫面看起來完全正常,只是不會動**。
    key 改成帶幀號。

    這已經是第四次撞到同一個根因:`decodeAsset` 沒有快取,於是每個呼叫端各自長一層
    (`colBldgCache`、`colVegSizeCache`、`nebMaskCache`,現在星圖幀也借 `colBldgCache`)。
    **下次再遇到就該去修 `decodeAsset` 本身**,不要再加第五層。

    ### 落點

    - `cmd/moo2/starsprite.go`:`starSpriteFrame(資產, 幀)` + `starSpriteFrameCount`
      + `blackHoleFrameAt(tick, 幀數)`(純函式,抽出來才驗得動)
    - `cmd/moo2/interactive.go`:`sceneBuilder.animTick`,由 `Update()` 每幀推進
    - 測試釘住三件事:非黑洞恆為第 0 幀(把「刻意不做」寫成護欄)、
      比例是 2、幀數 0/1 時不越界

---

54. **訂正兩條:艦隊圖示不會動、一般星球也不會閃——都是原版行為,不是缺口**(2026-08-07,接第 53 項)。

    做完黑洞動畫,順手去追「艦隊圖示 8 幀為什麼沒動」。查完發現**要改的是文件不是程式**。

    ### 先把引擎層的規則挖出來

    原版有一支通用貼圖器 `sub_12A478`,整個 UI 都用它。動畫模型是這樣:

    | 位置 / 函式 | 意義 |
    |---|---|
    | `word[圖+4]` | 目前幀號 |
    | `word[圖+6]` | 幀數 |
    | `sub_12B726(圖)` | 幀號 ← 0 |
    | `sub_12B753(圖, n)` | 幀號 ← min(n, 幀數−1) |
    | `sub_12A478(x, y, 圖)` | 畫完**自動把幀號 +1**(`[圖+0Bh]` 的 0x20 位元控制循環) |

    **「畫一次進一幀」是貼圖器的預設**。所以呼叫端要靜止,就得每次先歸零;要自訂節奏,
    就每次寫死幀號。這一條解釋了後面兩個問題。

    ### ① `Cycle_Ship_Icons_` 不是動畫,是 F1/F2

    `shipicon.go` 的檔頭原本寫著「每張圖有 8 幀(原版 `Cycle_Ship_Icons_` @ 0x82DFF 在跑動畫),
    remake 只取第 0 幀,動畫沒做」。**兩句都錯**:

    - `sub_82DFF` 由**鍵盤跳表**叫進來(`sub_825A8` 的 case −1001 等),`bx` 是方向
      (0 → `inc edx`、非 0 → `dec edx`),挑到的艦隊丟給 `sub_831B1` 選取。
      它是「切換到上/下一支艦隊」。
    - 手冊逐字對上:「You can cycle through the known fleets using the keyboard shortcuts
      [F1] and [F2].」**兩個獨立來源指同一件事。**
    - `Draw_Ship_Icons_` @ 0xA070F 在**每次繪製前**都呼叫 `sub_12B726`(幀號歸零),
      所以艦隊圖示恆為第 0 幀。

    → remake 取第 0 幀**與原版一致**。原本記在檔頭的那個「缺口」不存在。

    ### ② 一般星球的閃爍在出貨版是死碼

    第 53 項寫的是「三個常數沒追出來,**不編那三個數**,所以先靜止」。查證後可以說得更強:
    **原版根本不會閃。**

    `Draw_A_Star_` 的閃爍分支要 `star[+0x64] >= 0` 才會走。全檔查證:

    - `star[+0x64]` / `star[+0x65]` 的**位元組**寫入只有四處 reset 迴圈(全寫 0xFF = −1),
      外加一處星系生成時的暫存搬運。**沒有「計數歸零 + 寫入爆發長度」那組設定。**
    - 全域預算 `word_19C164` **只有歸零與遞減,全檔沒有任何遞增**。
      一個只減不加的預算,不可能是還在運作的機制的閘門。

    **正對照**(這一步不能省——上一輪才因為 PDF 連字把假陰性寫成事實):用同樣的搜法去找
    星球結構 `+0x16`(光譜)的寫入端,找得到好幾處(`mov byte ptr [ebx+16h], 6` 等)。
    **方法本身會命中,所以這次的零命中是真的零。**

    順帶一個交叉驗證:reset 迴圈跑 `0x48 = 72` 次,正好等於 `GalaxyStarCounts[3] = 72`
    (最大銀河的星數)——兩邊獨立落在同一個數字上。

    ### 這一項的產出是刪掉兩個假缺口

    沒有新程式碼。`shipicon.go` 與 `starsprite.go` 的檔頭各改一段,測試註解跟著改——
    `TestStarSpriteFrameForOnlyAnimatesBlackHoles` 的意義從「釘住刻意的取捨」升級成
    **「擋住有人看到 5 幀資產就順手接上動畫」**,那會讓 remake 比原版多一個原版沒有的效果。

    ### 副產品:手冊的快捷鍵表(真值,尚未實作)

    追 F1/F2 時把手冊的快捷鍵段落一起掃出來了。**行文中直接寫死的**(不靠版面推斷):

    | 鍵 | 作用 |
    |---|---|
    | F1 / F2 | 循環切換已知艦隊(一個方向 / 反方向) |
    | F5 / F6 | 系統視窗開著時,切到下一個 / 上一個已殖民星系 |
    | F9 | 測距:點第一顆星,再把游標移到另一顆,顯示兩者的秒差距 |
    | F10 | 快速存檔(沿用上次的存檔名,**會直接覆蓋**) |
    | ALT + F9 | 從星圖載入遊戲 |

    ⚠ 手冊另有一組 `ALT + F1..F8` 對應遊戲設定開關,但那些鍵在 PDF 裡是**右側邊欄標籤**,
    抽出來的文字流會把它排到前一個選項的描述尾巴,**對應關係有 off-by-one 的風險**
    (中間還缺一個 F4)。**所以不寫進表**——要用得先對原版逐項確認。
    (這條謹慎直接來自第 51 項:同一份 PDF 已經騙過一次。)

    F9 特別值得做:remake 第 45 項已經把秒差距模型建好了(`ParsecsBetweenStars`),
    但玩家在畫面上**看不到任何秒差距數字**。接上 F9 等於把已經做完的東西露出來。

---

55. **手冊的星圖快捷鍵接上:F1/F2、F5/F6、F9**(2026-08-07,第 54 項副產品的實作)。

    第 54 項追 `Cycle_Ship_Icons_` 時把手冊的快捷鍵段落掃出來了。這一項把**行文中直接寫死**
    的那幾個接上(邊欄標籤的 ALT+Fn 那組仍然不碰,理由見第 54 項)。

    ### F9 測距:讓一套已經做好的模型露出來

    秒差距模型在第 45 項就建好了——1 秒差距 = 30 遊戲單位、距離取整、引擎速度/星雲減速/
    干擾器範圍全掛在上面。**但玩家在畫面上看不到任何秒差距數字**,整套模型是隱形的。

    手冊逐字:「use the keyboard shortcut [F9]. You'll need to click on the first star,
    then move the mouse cursor over any other star to see the distance (in parsecs)」

    注意它是**兩段式而且跟著游標即時更新**:按 F9 → 點第一顆 → 移到哪顆就顯示到哪顆。
    不是「點兩顆看結果」。remake 照這個行為做。

    截圖驗到的數字是 **15 秒差距**(中型銀河,`759÷30 × 600÷30 = 25.3 × 20` 秒差距),
    對角最長約 32——從左上角量到地圖中段拿到 15,落在該落的量級。

    ### F1/F2 目前只有一個元素,而那是資料模型的事

    原版走的是逐艦隊的表;remake 的玩家艦隊是**單一集合**(`FleetAtStar`),AI 對手只有
    抽象的 `FleetStrength`、在星圖上沒有位置。所以循環集合現在只有一個元素——
    按下去等於「把視角拉回自己的艦隊」。

    **這不是把規則做小,是同一個模型缺口**:它也卡著星圖的遷移連線層。
    `TestKnownFleetStarsIsSingleForNow` 把這個限制釘成測試,多艦隊做出來時它會紅,
    那時候該改的是測試。

    ### 兩個實作上的坑,都是「看起來完全正常」那一類

    1. **提示字釘在固定座標會壓到星星。** 第一版把「測距:移到另一顆星」畫在 (30,34),
       截圖一看正好蓋掉左上角那顆星的名字。改成**跟著游標走**——星圖每個角落都可能有星,
       沒有哪個固定位置是安全的。
    2. **截圖廊的示範終點寫死索引會踩到戰爭迷霧。** 第一版寫死「第 1 顆星」,而那顆還沒探索,
       `starAtScreen` 會跳過不可見的星 → 截圖停在提示上,什麼距離都沒畫。
       改成執行時挑第一顆**可見且不是起點**的星。

    另外把星球的點擊熱區與懸停判定收斂到同一個 `starHitHalf`(22×22 方框)。
    兩邊各寫一份判定,遲早會出現「點得到卻懸停不到」的邊緣像素。

    ### 落點

    - `internal/shell/starnav.go`:`KnownFleetStars` / `ColonizedStars` / `cycleStarList`
      + `CycleFleetStar` / `CycleColonizedStar`(純邏輯,環狀,清單空回 −1)
    - `internal/shell/input.go`:`InputState.Hotkey`(用字串不用 ebiten 的 key 型別,
      規則層才不必相依 ebiten;headless 腳本因此也能注入按鍵)
    - `cmd/moo2/hotkeys.go`:按鍵對照 + F9 測距的畫面
    - `cmd/moo2/interactive.go`:`overlayScreen.onHotkey`(快捷鍵先於滑鼠處理)、
      截圖廊第 28 張

    ### 補完 F10 / ALT+F9(同日)

    - **F10 快速存檔**:手冊說它「沿用上次的存檔名、**直接覆蓋且不可回復**」。
      remake 的「上次的存檔名」就是 `sceneBuilder.savePath`——開局是自動存檔那一格,
      從載入視窗讀過某一格之後會改成那一格。**語意天然對上,不必另建概念。**

      覆蓋是原版行為,所以不加確認框;但**一定要回報存到哪**——沒有回報的話,
      按下去成功與失敗看起來完全一樣。星圖既有的 `lastActionMsg` 畫在選中星面板裡
      (42,415),沒選星就看不到,所以另加一個會自己消失的短暫訊息(約 3 秒),
      畫在星圖底緣偏右、避開左下角面板。**會消失是刻意的**:一直掛著的「已存檔」
      會被誤讀成「還在存」。

    - **ALT+F9 載入**:`loadGameInPlay()` 早就有(遊戲選單走得到),這裡只是多一個入口。
      ⚠ ALT 組合要**先於單鍵表判定**,否則 ALT+F9 會被當成 F9 測距。

---

56. **多艦隊模型:把「全帝國只有一支艦隊」這個限制拆掉**(2026-08-07,第一階段)。

    remake 先前把玩家的兵力表示成一組欄位:`Ships` + `FleetAtStar` / `FleetDestStar` /
    `FleetETA` / `FleetMarines` / `FleetTanks`。**全帝國只有一支艦隊**,所有的船永遠在同一個
    地方、只能有一個航行任務。

    這個限制卡著三件事:星圖的**遷移連線層**(4 層裡的最後 1 層)、**F1/F2 循環艦隊**
    (第 55 項做出來但循環集合只有一個元素)、以及分艦隊多線任務——原版最基本的操作。

    ### ⚠ 這次重構真正的難點:`Ships` 有兩種語意

    舊程式碼裡的 `s.Ships` 混著兩件事,而**單一艦隊時兩者剛好相同**,所以分不出來:

    | 語意 | 用在哪 |
    |---|---|
    | 「**這支艦隊**的船」 | 戰鬥、載運陸戰隊、消耗殖民船 / 前哨船、航行中損失 |
    | 「**全帝國**的船」 | 指揮點數(手冊 p.169 明文)、國力、艦名編號、外交評估、艦隊列表 |

    盲目全改成 `s.Fleet().Ships` 會讓第二類在**真的有第二支艦隊時**默默算少,
    而那時候看起來完全正常——數字只是偏小。所以逐處分類,並用 `internal/shell/fleet_test.go`
    的**兩支艦隊**測試把分類釘住:分錯的那一處立刻算少,不必等 UI 做出來才發現。

    分類結果(非測試碼約 65 處)在各處都留了一行說明為什麼是這一邊。

    ### 順帶修正的三個行為

    1. **修復**:先前只看「玩家選中的那一支」艦隊有沒有停在據點。改成**逐艦隊各自判定**
       ——這才對得上原版(`Repair_Ships_At_Colonies_` 的迴圈也是逐艦隊走的)。
    2. **母星防禦**:先前同樣只看選中那一支,於是玩家把視角切到別支艦隊,母星就「沒有防禦」。
       **那是純粹的操作副作用,不該影響世界狀態。** 改成任何一支停在母星的艦隊都算。
    3. **隨機事件的「損失一艘艦」**:原版打的是整個帝國,不是玩家的操作焦點。
       改用 `removeShipGlobal`(跨艦隊索引)。

    ### 存檔:發現舊格式有個漏欄

    新格式序列化整個 `Fleets`;讀到舊檔(`len(Fleets) == 0`)就從舊欄位組成唯一的一支。
    **判準用「欄位在不在」不用版本號**——版本號會被別的改動一起往上帶。

    做遷移時發現:**舊格式從來沒有存戰車營**(有 `fleetMarines` 卻沒有對應的 `fleetTanks`)。
    也就是舊存檔讀回來戰車營一律歸零。那是舊格式本身的漏欄,不是這次弄丟的;
    新格式把整個 `Fleet` 序列化,這個洞順帶補上。

    ### 驗收

    重構**不改變行為**,所以驗收是「畫面要一模一樣」:重跑截圖廊 29 張,
    **28 張逐位元相同**;唯一不同的是載入視窗,差在存檔檔案的時間戳(每次跑都會變)。

    ### 還沒做的(第二階段)

    - 分艦隊 / 合併艦隊的 UI 與規則(選幾艘船拆出去)
    - 逐殖民地造艦 + 集結點 → 才畫得出**遷移連線層**
      (`AddShipToHomeFleet` 目前用「艦隊剛好停在那顆星就併進去」近似)
    - AI 的艦隊在星圖上還是沒有位置,所以 F1/F2 仍只走玩家自己的艦隊

---

57. **遷移連線:星圖 4 層裡的最後一層補上**(2026-08-07,第 56 項的續)。

    原版 `Draw_Relocation_Links_` @ 0x85320 是主畫面圖層順序的第 2 層。它是 remake
    最後一個沒做的星圖層——先前卡在「艦隊是單一集合」,多艦隊(第 56 項)做完之後,
    缺的只剩這一層自己的資料。

    ### 資料模型:兩支函式讀同一個欄位,互相印證

    ```
    sub_784F0(星, 玩家) → word[星×0x71 + 0x54 + 玩家×2]     ; 遷移目標星
    sub_78C94(星, 玩家) → 上面那個欄位 != -1                  ; 「有沒有設定」
    ```

    **每個(星, 玩家)一個目標星索引**,−1 = 沒設定。`Draw_Relocation_Links_` 的迴圈就是:
    走每一顆星,目前玩家在那裡有設定就畫一條線過去。

    手冊給了它的作用:「Relocation Lines controls the appearance of travel lines for those
    ships being **automatically relocated** between star systems.」——新造的艦會自動送過去,
    而且那是**一段航程**,不是瞬間移動。remake 因此建一支往目的地航行的艦隊,
    星圖上那條線畫的就是它。

    ### 顏色是真值,而且它很暗——那是原版的樣子

    原版把一個 8 位元組的**調色盤索引表**丟給畫線函式:

    ```
    dword_81C80  dd 2 dup(70706F6Eh)     ; = 6E 6F 70 70 6E 6F 70 70
    ```

    在 BUFFER0.LBX#0 的調色盤裡解出來是 **(0,20,0) / (4,56,4) / (0,76,0)** ——深綠。
    壓在黑色星空上很低調,截圖裡幾乎看不見。**那不是畫錯**:手冊自己就說
    「If you'd rather not clutter up the galaxy with them, turn this option off.」
    它的定位是**可以關掉的雜訊**,不是要搶眼的指示線。

    所以**不把顏色調亮**。驗證它有沒有畫出來的方法是截圖加亮
    (`-channel G -evaluate multiply 6`),不是為了「看得清楚」去改一個已經是真值的數字。

    `sub_A11C0` 另外收一個相位參數(`edi = 7 - 相位`,並依線的方向反轉表)——
    **那是讓漸層沿著線跑**,玩家因此看得出方向。⚠ 相位的來源沒追死(那個全域在星圖這一段
    只被讀、沒被寫),remake 用自己的 `animTick` 驅動,步進沿用已被證實**三次**的
    「2 次重畫換 1 步」比例(黑洞、一般星球、這裡)。

    ### 兩個實作上的坑

    1. **反鋸齒把線畫沒了。** 原版是逐像素寫調色盤索引、硬邊;remake 用向量畫線時開了
       反鋸齒,把本來就很暗的深綠又和黑底混一次——線幾乎消失。關掉反鋸齒才對得上。
    2. **截圖廊的示範目標和 F9 測距撞在一起。** 兩條線走同一對星,遷移連線整條藏在
       測距線底下,截圖看起來像沒做。改成挑**第二顆**可見的星。

    ### 唯一的零值陷阱,而且它很致命

    `ColonyRelocateTo` 的 Go 零值是 **0 = 母星的索引**。平行陣列補齊時如果填零值,
    **每個新殖民地一建好就會把新造的艦全部往母星送**——而那看起來完全像是遊戲規則。
    `growRelocation` 一律填 −1,`TestRelocationDefaultsToNoneNotStarZero` 把它釘住。

    ### 順手把一個「指名了不存在的護欄」的註解修掉

    `hotseat.go` 的 seat 型別上寫著「`TestSeatFieldsCoverPlayerSide` 用反射盯著它」,
    而**那支測試根本不存在**(加集結點欄位時發現)。指名了不存在的護欄比沒有註解更危險
    ——它讓人以為這裡有人在看,於是加欄位時不會特別小心。

    改成真的寫一支 `TestSeatRoundTripKeepsEveryField`:用反射把每個 seat 欄位塞成可辨識的
    非零值,`loadSeat` 再 `saveSeat` 抓回來比對,漏抄的欄位會停在零值。
    它立刻抓到 `SelectedFleet` ——不過那是測試填了越界值被不變量夾回,不是產品 bug,
    改成填合法值(兩支艦隊 + 選第 1 支)。

    ### 還沒做的

    - 原版是「(星, 玩家)」而 remake 是「第 i 個殖民地」——對玩家等價,
      但 **AI 的遷移設定沒有建模**(AI 沒有逐星的艦隊位置)。
    - 顯示開關 `ShowRelocationLines` 已建,但**還沒有 UI 可以切**
      (原版在設定畫面,對應手冊那組 ALT+Fn ——⚠ 哪一個鍵仍未確認,見第 54 項)。

---

58. **艦隊列表改成真的列「艦隊」+ 清掉 HONEST-STATUS 三條過期斷言**(2026-08-07)。

    ### 艦隊列表先前列的是「船」

    畫面標題是 **FLEET OPERATIONS**,而 remake 把名冊攤平成一長串船名——那是單艦隊時代的殘留:
    **全帝國只有一支艦隊時,「列船」與「列艦隊」看起來一樣。** 多艦隊之後就不一樣了,
    玩家需要看到哪幾艘在一起、停在哪、有沒有在航行,才能決定要操作哪一支。

    改成逐艦隊分組:標頭是「▶ 第 N 艦隊 — 星名(M 艘)」,航行中補「→ 目的地,K 回合」,
    標頭可點擊切換 `SelectedFleet`。`▶` 標出目前操作中的那一支——星圖上所有的按鈕
    (派遣/拓殖/轟炸/入侵)作用的都是它。

    ⚠ 原版這個畫面的美術上就烘著 **RELOCATE** 按鈕(remake 譯「調動」),而手冊說
    「You set up your Relocation orders on the **Fleet Operations console**」——
    **集結點的忠實入口就是這裡**,不是 remake 目前放的星圖面板。
    但那顆鈕按下去原版做什麼(選殖民地?選艦隊?)沒有反組譯確認,**所以先不接**,
    留星圖那條路能用。

    ### 順手清掉 HONEST-STATUS 的三條過期斷言

    | 原本寫的 | 查證結果 |
    |---|---|
    | 「中段的行星表面 + 建築 sprite 擺放子系統仍未做」 | **同一份文件裡自己打架**——同一天稍晚的段落就寫著地表、道路、抖動、國會大廈、植被全都做完了 |
    | 「建築集合仍與原版有差(Colony Base、一次性改造沒建模)」 | Colony Base 第 48 項補上;一次性改造第 52 項**查證後確認不該佔格**(原版記旗標那步對那四個編號是跳過的),remake 天然正確 |
    | 「戰機/航母(新戰鬥子模型)需先建基礎設施」 | 說得太重。戰機**已經接進快速艦隊戰鬥**(`FighterBayCombatContribution`,中隊數 4/2 是手冊 GM p.127 硬數字),`ResolveBattle` 與 `repair.go` 都在用。真正缺的是**戰術格子裡的獨立戰機單位**(出擊/攔截/回收) |

    **同一份文件前後矛盾**這件事本身值得記:那三條都寫在「誠實現況評估」裡,
    而它們正是外部(包括自動巡檢)判斷「還缺什麼」的依據。
    過期的缺口清單會讓人去做已經做完的事,或者把做完的事當成沒做。

---

59. **RELOCATE 鈕的原版語意追出來了:兩段點選 + 四條合法性規則**(2026-08-07,第 58 項留的問題)。

    第 58 項把 RELOCATE 鈕標成「按下去做什麼沒有反組譯確認,所以先不接」。查了符號表,
    整組符號都在,問題直接解掉:

    ```
    Okay_To_Set_Relocate_From_Star_ @ 0x74F8A    Star_Relocation_        @ 0x75180
    Okay_To_Set_Relocate_To_Star_   @ 0x74FAA    Cancel_Star_Relocation_ @ 0x7522B
    Okay_To_Set_Relocate_Star_      @ 0x75035    Set_All_Star_Relocations_ @ 0x785EC
    ```

    ### 流程:**兩段點選**,不是「選一個殖民地」

    `Star_Relocation_(&起點, &終點, 剛點到的星)`:

    ```
    if *起點 == −1:  驗證能不能當**起點** → 通過就記起來,結束
    else:            驗證能不能當**終點** → 通過就記起來
                     if *終點 == *起點: Cancel_Star_Relocation_    ; 點回自己 = 取消
    ```

    remake 先前的「星圖面板 → 點一顆星」是**第二段**;第一段被略過(用面板選中的那顆星
    當起點)。那是合理的捷徑但不是原版入口。現在兩條路都在:
    艦隊列表的 RELOCATE 走完整兩段(手冊逐字:「You set up your Relocation orders on the
    **Fleet Operations console**」),星圖面板的鈕走捷徑,規則面共用同一支 `SetStarRelocation`。

    ### 四條合法性規則(`Okay_To_Set_Relocate_Star_`,`dl` 區分起點/終點)

    | 規則 | 反組譯依據 |
    |---|---|
    | 黑洞不能當起點也不能當終點 | `cmp byte [星+0x16], 6` → 兩種訊息(0x83/0x84),同一條規則 |
    | 沒探索過的星不行 | `star[+0x33] & (1<<玩家)` 的位元測試(逐玩家的探索遮罩) |
    | 目的星上有艦隊 → **跳確認框**;當起點則直接不行 | `sub_7A47A` 走艦隊表 `word_192248[]` |
    | 起點必須是自己有殖民地的星 | `star[+0x38]` 的位元測試(`sub_79D1C`) |

    ⚠ 第三條的**確認框沒做**——remake 沒有 modal 對話框的基礎設施,目前直接允許。
    這是**已知的簡化,不是漏看**,寫在 `relocation.go` 檔頭。

    ### 順帶記兩個還沒接的原版功能

    `Set_All_Star_Relocations_` @ 0x785EC 與 `Clear_All_Star_Relocations_` @ 0x77BB1
    ——艦隊列表上的 **ALL**(remake 譯「全部」)鈕多半就是它們:一次把所有殖民地的集結點
    設成同一顆星 / 全部清掉。**還沒接**,也還沒確認 ALL 鈕是不是對應這一對。

---

60. **分艦隊:原版的「艦隊」是 ship stack,拆分是把船從串列摘下來**(2026-08-07,task #50 收尾)。

    ### 資料結構

    原版沒有「艦隊」這個型別,有的是 **ship stack**:

    ```
    word_192248[stack]      = 這一疊的頭一艘船 id
    word_1975D6[船 id × 5]  = 「下一艘」            ; 單向串列
    word_1975D4[船 id × 5]  = 那艘船在哪顆星
    word_199A02             = 目前的 stack 數
    ```

    `Split_Stack_` @ 0x5D689 收一組船 id,把它們從原本的串列摘下來、串成新的一個 stack,
    **接在 stack 表的尾端**。所以拆分的語意是「選一組船,抽出來組成一支新艦隊,位置不變」。
    remake 用切片而不是串列,語意逐項對得上(`SplitFleet`)。

    ### 三條擋下來的退化情形,其中一條是 remake 特有的

    | 情形 | 為什麼擋 |
    |---|---|
    | 沒選任何船 / 選了全部 | 「全選」不是拆分——那會留下一支空的舊艦隊 + 一支一樣的新艦隊 |
    | 艦隊索引或船索引越界 | 一般防禦 |
    | **艦隊正在航行** | remake 的航行是**整段跳的**,中途沒有位置,拆出來的那一半沒有可放的地方。**這是 remake 移動模型的後果,不是原版規則**——原版的 stack 隨時有座標 |

    ### 一個已知的簡化,寫成測試釘住

    陸戰隊 / 戰車營**全部留在原艦隊**:remake 把它們建模成艦隊層級的數字,不綁定到特定的船,
    所以拆分時沒有「哪些跟著走」的依據。要改成逐船攜行,得先讓運兵成為船的屬性。

    ### ⚠ 這個 UI 是 remake 自己加的

    原版艦隊列表的美術上**沒有 SPLIT 鈕**(烘著的是 ALL / RELOCATE / SCRAP / LEADERS /
    Support / Combat / RETURN)——原版是在**右側的艦艇格**選船再下令。
    remake 的右側格還沒接上選取,所以先用左側名冊勾選 + 一行文字當入口。
    追到原版怎麼下拆分令之後要換掉。

    ### 版面的坑:兩個東西搶同一塊底部

    第一版把拆分那一行放在船清單**之後**,而名冊是往下長的——結果和固定在 y=402 的
    「攻打安塔蘭母星」疊在一起。改放在**艦隊標頭底下**:名冊再長也不會撞到底部。

---

61. **一星多行星:軌道模型的資料層(第一階段)**(2026-08-07)。

    remake 的 `Stars[i]` ↔ `Planets[i]` 是**一對一**——一顆星一顆行星。MOO2 不是這樣。

    ### 「每個星系 5 個軌道」是真值,三個獨立來源

    | 來源 | 內容 |
    |---|---|
    | 偏移算術 | 軌道陣列在星球結構 +0x4A,下一個已知欄位(遷移目標)在 +0x54 → 中間 **10 位元組 = 5 個 word** |
    | `System_Planet_Scanned_To_Planet_Id_` @ 0x78CDB | `word[星×0x71 + 0x4A + 軌道×2]` = 軌道 → 行星 id(−1 = 空) |
    | 走訪迴圈 | 上界寫死:`cmp word ptr [var_4], 5; jge`(0x1CB31) |

    行星是**獨立的一張表**(`dword_1930D4`,每筆 **0x11 = 17 位元組**),
    `Planet_Orbit_` @ 0x783ED 讀 `byte[行星 id×0x11 + 3]` = 它在第幾號軌道。
    **雙向都有指標**:星 → 軌道 → 行星 id,以及行星 → 軌道號。

    ### 意外發現:骰表早就在骰整個星系了

    `genPlanets` 已經在跑 `RollNumSatellites` + 逐軌道 `RollSatelliteType`,
    然後**挑一顆代表行星**,其餘只存成 `Planet.SystemBodies` 的摘要(有軌道/類別/名字,
    沒有氣候礦產)。所以缺的不是骰表,是**「其他天體是二等公民」**這個表示法。

    ### 這一階段做的:形狀,不是內容

    - `Star.Orbits [5]int`(`OrbitEmpty = −1`),產生器把代表行星放進它骰到的那個軌道
    - 存取器:`PlanetAt`(第一個有行星的軌道 = 舊的 `Planets[星]`)、`PlanetsAt`、
      `PlanetStar` / `PlanetOrbit`(反查)、`FreeOrbit`(**人造行星要用它**)
    - 存檔遷移 `normalizeOrbits`

    **行為逐位元不變**——每顆星仍然只有一顆行星。`TestGeneratedStarsHaveExactlyOneOccupiedOrbit`
    把這個限制釘住,`TestPlanetAtMatchesLegacyParallelIndexing` 釘住相容性支點:
    一星一行星時 `PlanetAt(i)` 必須等於 `i`,否則舊呼叫端換過來會位移。

    ### ⚠ 又一次同款零值陷阱

    軌道表的 Go 零值是 5 個 **0**,而 0 是**行星 0 的索引**——不修的話每顆星都會宣稱
    軌道 0 上有行星 0,**而且不會報錯**,只會讓每顆星的行星資料看起來都一樣。
    這與 `Star.Wormhole`(零值 0 讓每顆星都連到星 0)、`ColonyRelocateTo`(零值 0 = 母星)
    是**同一個形狀的坑,第三次出現**。共同點:**索引型欄位的「沒有」必須是 −1,不能靠零值**。

    ### 下一階段(才解得開人造行星)

    把 `SystemBodies` 升格成真正的 `Planet` 條目、填滿軌道表,
    再讓殖民/前哨站/建築 48 以「星系裡的某一顆行星」為對象而不是「某一顆星」。
    那會動到約 33 個 `s.Planets[星]` 的呼叫端——與多艦隊那次同一套做法:
    先給存取器、逐處分類、用「兩顆行星」的測試把分類釘住。

62. **一星多行星 Step A:33 個呼叫端改走存取器**(2026-08-07,第 61 項的續)。

    第 61 項建好了軌道模型,但所有讀行星的地方仍然直接寫 `s.Planets[星]` ——
    那個式子**假設 Planets 與 Stars 平行**。一旦產生器開始填滿軌道(Step B),
    `len(Planets)` 就會大於 `len(Stars)`,那些式子會**默默讀到錯的行星**
    ——不會崩、不會報錯,只是資料錯位。

    所以先做一步**行為不變**的遷移:全部改走存取器。

    | 存取器 | 用途 |
    |---|---|
    | `PlanetAt(星) int` | 代表行星的索引 |
    | `PlanetOf(星) *Planet` | **可寫**指標(隨機事件改礦產/氣候、拓殖消耗特殊物產、抵達時的一次性發現) |
    | `PlanetDataAt(星) (Planet, bool)` | 唯讀複本 |

    ### 代表行星的挑法必須與產生器**逐字相同**

    `PlanetAt` 改成:依軌道順序找第一顆一般行星(可殖民);整組都不宜居時才退而取第一個天體。
    那正是 `genPlanets` 原本挑代表行星的規則。**不一致就會位移**——
    而位移的徵狀是「殖民地總覽說類地、殖民地畫面說凍原」那一類自打嘴巴,
    不是崩潰(2026-08-06 踩過同款)。

    順手拿掉幾個 `star < len(sess.Planets)` 的邊界檢查:那是「Planets 與 Stars 平行」
    這個假設的殘留,Step B 之後會**變成錯的**(而且是放行太多而不是擋太多)。

    ### 驗收

    行為不變,所以驗收是畫面要一模一樣:重跑截圖廊 **28/29 張逐位元相同**,
    唯一不同的載入視窗差在存檔時間戳。

    ### Step B 還沒做

    把 `SystemBodies` 升格成完整的 `Planet` 條目、填滿軌道表。⚠ 那會**改變同一 seed 的星系內容**
    (多骰了那些天體的氣候/礦產/重力),所以要給它們**獨立的亂數流**——
    這個專案已經用過這招(「行星生成用獨立的亂數流 seed+1,不讓抽取次數影響佈局」),
    否則代表行星之後的每一顆星都會漂掉。

63. **一星多行星 Step B:同系天體升格成真正的行星**(2026-08-07,task #51 收尾)。

    第 61 項建了軌道模型、第 62 項把呼叫端改走存取器,這一步把資料真的填進去:
    **同一顆恆星底下的每一個天體都是完整的 `Planet` 條目,各佔一條軌道。**
    `Planets` 因此**不再與 `Stars` 平行**(24 顆星 → 94 顆行星)。

    ### 獨立的亂數流,理由與既有的三條一樣

    非代表天體用 `bodyRand`(seed+5)而不是共用 `r`。這個專案已經為
    `genPlanets` / `genMonsters` / `genWormholes` 各開一條流,理由寫在原註解裡:
    **多骰幾次不能讓後面的東西跟著漂**。這裡同理——共用一條的話,
    第 0 顆星多骰的那幾次會把第 1 顆星以後**每一顆**的代表行星換掉。

    ### 測試抓到三個「還在假設平行」的地方,兩個在產品碼

    | 位置 | 症狀 |
    |---|---|
    | `monster.go` 兩處 | `planets[starIdx]` —— 怪獸的特殊物產會補到**別的星系**的行星上 |
    | `galaxy_gen_test.go` / `monster_test.go` | 測試自己也在平行索引 |

    這正是第 62 項先做「改走存取器」那一步的用意:**那些式子不會崩,只會讀到錯的行星**。
    抓到它們的不是編譯器(索引式在型別上是合法的),是**跑起來的不變量測試**
    ——「母星一定宜居」「有怪獸的星系一定有特殊物產」。

    順手把「代表行星怎麼挑」收斂成唯一一份實作(`representativePlanet`),
    生成階段與 `GameSession.PlanetAt` 共用。兩份實作一旦漂開,徵狀又是資料錯位而不是崩潰。

    ### `SystemBodies` 淘汰:它自己註解裡擔心的事解掉了

    那個欄位的原註解寫著「**這裡不重複放代表行星本身,避免兩份資料要同步的老問題**」
    ——它知道自己是折衷。現在同系天體是真正的行星,摘要文字改從軌道表算
    (`GameSession.SystemCompositionText` / `SystemBodyCountText`),**只有一份資料**。
    欄位留著只為了讀得回舊存檔的顯示。

    ### 三條測試換掉,因為它們釘的是階段性限制

    `TestGeneratedStarsHaveExactlyOneOccupiedOrbit` 的註解當初就寫著
    「升格之後這條會紅,**那時候該改的是測試**」。換成三條更強的:
    每個軌道條目都指到有效且**不重複**的行星、每顆行星都掛在某個軌道上、
    銀河裡**真的有**多天體星系(沒有這一條的話「升格」可能只是搬了位置而實際仍一星一顆),
    以及代表行星的挑法(有一般行星就一定挑它)。

    ### 驗收

    行星列表出現多天體星系(奧格卡 I / II、塔拉皮斯 I / II 各標「另有 N 天體」)。
    截圖廊只有 `12_planets.png` 變(內容變多是預期的)與載入視窗的時間戳。

    ### 這一步解鎖了什麼

    `FreeOrbit` 現在真的有意義——**人造行星**(建築 48)可以往空軌道放。
    同星系多殖民地也有了資料基礎。兩者的規則接線仍未做。

64. **人造行星:手冊推翻了 remake 自己的假設**(2026-08-07,建築 48 接線)。

    ### 先訂正一句寫了兩輪的話

    gap report 第 61 項(以及更早的第 51 項)寫著:

    > 人造行星按定義是**在既有星系裡再多一顆世界**,remake 的 Stars↔Planets 是一對一,
    > 轉換完沒地方放第二個殖民地。

    於是 `FreeOrbit` 被寫成「人造行星要用它:沒有空軌道就蓋不了」。**手冊說的不是那樣。**

    > 「This technology allows a colony in the same system with an asteroid field or gas giant
    >  to **assemble this otherwise useless planetary material** into a complete artificial
    >  planet that can support a colony. This planet is **Barren, Normal G, and mineral
    >  Abundant**. **Gas giants make Huge worlds, and asteroid belts make Large ones.**」

    它是把**既有的**氣態巨星或小行星帶組裝成行星——那顆天體**本來就佔著一條軌道**。
    所以前置是「同星系有材料」,不是「有空軌道」。
    `TestArtificialPlanetNeedsMaterialNotFreeOrbit` 把這個訂正釘住:
    **五個軌道全滿但有氣態巨星 → 可以蓋;有空軌道但沒有材料 → 蓋不了。**

    這是「**先查一手資料再推論**」的又一個實例:那句斷言是從「人造行星」這個名字推的,
    推得很合理,而且它擋了兩輪工作。

    ### 反組譯逐項吻合

    `sub_13FD9` 的那一段走**兩趟**掃這顆星的 5 條軌道:

    | 趟次 | 找什麼 | 結果尺寸 |
    |---|---|---|
    | 第一趟 | `planet[+4] == 2`(氣態巨星) | `var_1C = 4` |
    | 第二趟(第一趟沒中才跑) | `planet[+4] == 1`(小行星帶) | `var_1C = 3` |

    **氣態巨星優先**,而 4 / 3 在尺寸列舉裡正是 **Huge / Large** —— 與手冊那句逐字對上。
    接著把型別欄寫成 3(一般行星)並改寫整組欄位。

    ⚠ **兩趟不能合成一趟**:合起來的話,軌道較內的小行星帶會搶在外側的氣態巨星前面被挑走。
    測試用「小行星帶在內、氣態巨星在外」的配置把這個順序釘住。

    ### 成本用真值不用估值

    第一版順手寫了 `ProductionCost: 900`。第 36 項抽出來的原版建築表寫著 **800**。
    改掉——**專案裡已經有真值就不該再估一個**。

    ### 落點

    - `internal/gamedata/special_actions.go`:第 5 個 Special 行動(前置 `TOPIC_ADVANCED_MANUFACTURING`)
    - `internal/shell/artificialplanet.go`:兩趟掃描 + 固定結果(Barren / Normal G / Abundant)
    - `session.go` 的 Special 分派:沒有材料時**誠實地什麼都不發生**
      (與土壤改良在錯誤氣候上的處理同一個立場:不在選單擋,但套用時不硬塞效果)

65. **`Set_All` / `Clear_All` 集結點 + 遷移連線的顯示開關**(2026-08-07,把第 57/59 項留的小項清掉)。

    ### `Set_All` 有一個猜不到的細節

    ```
    for star = 0 .. 星數-1:
        if word[星×0x71 + 0x54 + 玩家×2] != -1:      ; ← 只改**已經有設定**的
            word[...] = 目標星
    ```

    **它不是「把每個殖民地都設成這顆星」。** 沒設過的殖民地不會被順便設上。

    直覺會做成「全部設成這顆」——而那會讓玩家按一下 ALL 就把**所有**新殖民地的產出全部抽走,
    包括他本來就想留在原地生產的那些。這是一個從按鈕名字("ALL")推不出來的規則,
    `TestSetAllOnlyRetargetsExisting` 把它釘住。

    `Clear_All_Star_Relocations_` 同結構,清成 −1。

    艦隊列表的 **ALL**(remake 譯「全部」)鈕接上前者;`Clear_All` 規則已實作並測試,
    但**還沒有 UI 入口**——原版那顆鈕按下去是哪一支沒有確認,不猜。

    > ⚠ **2026-08-07 訂正(第 70 項)**:上面那句「ALL 鈕接上前者」是**推測,而且推錯了**。
    > 手冊兩處明說 ALL 是「全選/全不選這支艦隊的艦艇」;原版的 `Set_All` / `Clear_All`
    > 是**星圖上的鍵盤指令**,不是按鈕。詳見第 70 項。

    ### 遷移連線的顯示開關有地方放了

    原版的開關在設定畫面(手冊那組 ALT+Fn ——⚠ 哪一個鍵仍未確認,見第 54 項)。
    remake 沒有設定畫面,但遊戲選單上那顆 **SETTINGS** 鈕本來就是死的
    (檔頭寫著「按鈕保留但不接」)。現在它展開一列設定,目前只有一項:遷移連線的顯示。

    ⚠ **那一列不是原版版面**——原版有一整個設定畫面。建了那個畫面之後要搬過去。
    但「一顆點了沒反應的鈕」與「一顆展開唯一一個真開關的鈕」相比,後者誠實得多。

66. **同星系多殖民地:拓殖的對象是行星,不是星**(2026-08-07,把第 61/63 項留的最後一步做完)。

    ### 一句話擋掉了整個擴張手段

    `ColonizeStar` 的第二道閘寫著:

    ```go
    if star.Owner != 0 {
        return ColonizationResult{Reason: "該星已有歸屬,不可拓殖"}
    }
    ```

    一星一行星的時代它是對的:一顆星只有一顆行星,有歸屬就等於沒空位。
    軌道模型上線之後(第 63 項,一個星系 1..5 個天體)那句話就變成
    「**你自己的星系不准再殖民**」——而手冊 p.61 的條件從頭到尾是
    「any uncolonized **planet**」,講的是行星不是星。

    ### 這一輪換掉的東西

    | 之前 | 之後 |
    |---|---|
    | `ColonizeStar(star)` 是唯一入口 | `ColonizePlanet(planet)` 是入口;`ColonizeStar` 變成「該星系第一顆可殖民行星」的捷徑 |
    | `newColonyFromStar(star)` 讀 `Planets[star]` | `newColonyFromPlanet(planet)` |
    | 殖民地只記 `PlayerColonyStars[i]` | 加上 `PlayerColonyPlanets[i]`(AI 側是 `AIOpponent.ColonyPlanets`) |
    | 前哨站只記在哪顆星 | `Outpost.PlanetIndex`(手冊 p.119:「build a military outpost on a single planet」) |
    | 殖民地名 = 該星代表行星名 | `ColonyName(i)` = 該殖民地**座落行星**的名字 |
    | 地表變體種子 = 星索引 | = 行星索引(否則同星系兩個殖民地地表一模一樣) |

    ### 又一次同款零值陷阱

    `PlayerColonyPlanets` / `Outpost.PlanetIndex` 的「未知」都必須是 **−1**。
    這是第四次了(`Star.Wormhole` → 全部連到星 0、`ColonyRelocateTo` → 全部指向母星、
    `Star.Orbits` → 每顆星都宣稱有行星 0)。索引型欄位的 Go 零值是一個**合法索引**,
    不是「沒有」。舊存檔沒有這兩個欄位 → nil → 退回「該星的代表行星 / 同星就算」,
    行為與加欄位前逐位元一致,`TestLegacySaveWithoutColonyPlanetsFallsBackToStar` 釘住。

    ### 前哨站順帶修好的一個 bug

    `consumeOutpostForColony` 原本只比對**星**。多天體之後,在一顆行星建殖民地會把
    同星系另一顆氣態巨星上的前哨站一起吃掉,還白送一座海軍陸戰隊營。
    現在比對行星(舊存檔的 −1 退回舊語意),`TestOutpostOnAnotherBodySurvivesColonizingNeighbour` 釘住。

    順便放寬了 `BuildOutpost` 的 `Owner != 0` 閘——氣態巨星/小行星帶常常就在自己已經殖民的
    星系裡,原版當然讓你在那裡建前哨站。改成只擋敵方(`Owner == 2`)。

    ### 選行星的 UI 就是原版那個畫面

    不必新造:**行星列表**(`PLNTSUM.LBX`)右下角本來就烘著
    `SEND COLONY SHIP` / `SEND OUTPOST SHIP` / `RETURN` 三顆鈕——那是原版選行星的地方。
    先前那個畫面是唯讀的展示(而且列的是 `Planets[0..7]`,與星系無關)。現在:

    - 列出**目前看得見的星系**(`shell.VisibleStars`,與星圖同一套可見性)的所有天體
    - 點一列選中(換色),第二行顯示所屬星系 + 該系天體數 + 已殖民/前哨站狀態
    - 兩顆動作鈕對選中的那顆行星作用;艦隊不在那個星系就**先派艦隊過去**
      (原版那兩顆鈕的字面意思就是「派船過去」)

    星圖的星系面板保留「下一顆可用天體」的捷徑鈕,鈕上寫出目標天體名——
    那個面板只有 402 / 424 兩列,塞不下逐行星的選擇,所以完整的選擇走行星列表。

    ### 落點

    - `internal/shell/colonization.go`:`ColonizePlanet` / `FirstColonizablePlanet` /
      `ColonyIndexOnPlanet` / `ColonyPlanetIndex` / `ColonyPlanet` / `ColonyName`
    - `internal/shell/outpost.go`:`Outpost.PlanetIndex` / `BuildOutpostOnPlanet` /
      `OutpostTargetPlanet` / `HasOutpostOnPlanet`
    - `internal/shell/session.go`:`PlayerColonyPlanets`、`AIOpponent.ColonyPlanets`、
      `syncAIColonyPlanets`(行星索引要等 `Planets` 生完才補得起來)
    - `internal/shell/persist.go` / `hotseat.go`:兩個新陣列的存讀與席位交接
    - `cmd/moo2/interactive.go`:行星列表接上選擇 + 兩顆動作鈕;`cmd/moo2/relocation.go`
      的 `starPanelColonyRows`(hits 與繪製共用同一份佈局,免得「畫得出來卻點不到」)
    - `internal/shell/multicolony_test.go`:8 支測試(同系兩殖民地、同行星不可二殖、
      敵系仍不可拓殖、捷徑會挑下一顆、前哨站不被鄰居吃掉、存檔往返、舊存檔退路、
      人造行星改造完可殖民)

67. **AI 也會在自己的星系裡拓殖第二顆行星**(2026-08-07,把第 66 項的另一半補上)。

    ### 一個 remake 自己造出來的不對稱

    第 66 項只改了玩家側。`aiExpand` 的候選集寫死成

    ```go
    for idx := range s.Stars {
        if s.Stars[idx].Owner != 0 { continue }   // ← 只找無主星
        ...
    }
    ```

    於是玩家可以在自己的星系裡塞滿殖民地,AI 一個星系永遠只有一個。這不是原版的規則差異,
    是 remake 改了一半留下的不對稱。

    ### `Star.Owner` 分不出是哪一個 AI

    候選集要加進「**自己**已有殖民地的星系」,而 `Star.Owner` 只有 0/1/2 三個值
    (無主 / 玩家 / AI),分不出 `2` 是哪一家。所以判定要走各自的 `ColonyStars` 清單
    (`aiCanExpandInto`),不能只看 `Owner`。兩支測試分別釘住「不能進玩家的星系」與
    「不能進**另一個 AI** 的星系,但可以進自己的」。

    ### 兩個會被灌水的計數器

    - `OwnedStars` 只在「本來無主」時才 `++`。在自己的星系裡多殖民一顆行星不會讓版圖變大,
      跟著加的話征服勝利判定與外交評分都會偏。
    - `PlanetColonized` 取代原本只查玩家的 `ColonyIndexOnPlanet`。只查玩家的話,
      AI 擴張會把殖民地疊到玩家(或另一個 AI)已經佔著的行星上——那是一星多殖民地
      打開之後才會出現的競態。

    ### 順帶修好入侵的一個 bug

    `InvadeColony` 打贏就無條件 `star.Owner = 1`。同星系多殖民地之後,
    打下星系裡的**一個**殖民地會把整顆星判給玩家,剩下那個敵方殖民地就變成
    「站在玩家星系裡的敵軍」——星圖顏色、可入侵性、`ColonizePlanet` 的 `Owner == 2` 閘全部對不上。

    現在星的歸屬(以及 `StarCaptured` 回報與 `OwnedStars--`)只在**該 AI 在這顆星上
    再也沒有殖民地**時才發生。同時過戶的殖民地改用 `AIOpponent.ColonyPlanets[colonyIdx]`
    這個真值,不再退回「該星系的代表行星」——那顆代表行星很可能正是還沒被打下來的那一個。

    ### 落點

    - `internal/shell/session.go`:`aiCanExpandInto` / `aiExpansionCandidates` / `aiExpand`
    - `internal/shell/colonization.go`:`PlanetColonized`(全帝國視角)
    - `internal/shell/ground_invasion.go`:歸屬翻面的條件 + 過戶行星用真值
    - `internal/shell/multicolony_test.go`:再加 5 支(AI 自家星系拓殖、OwnedStars 不灌水、
      不能進玩家/別家 AI 的星系、PlanetColonized 看得見 AI、入侵不提早翻面)

68. **是/否確認框 + 一條寫錯的規則**(2026-08-07,把第 59 項留的「沒有 modal 基礎設施」清掉)。

    ### 先訂正:那個條件不是艦隊,是怪獸

    `relocation.go` 檔頭與本報告都寫著:

    > ③ 目的星上有艦隊 → **跳確認框**問玩家(當起點則直接不行)

    逐指令讀 `Okay_To_Set_Relocate_Star_` @ 0x75035 之後,那一段是:

    ```
    loc_750B5:
        call Star_Guarded_By_Monster_(ebx=星, edx=&怪獸id)   ; sub_7A47A,符號表裡就有名字
        jz   loc_7511F                                       ; 沒怪獸 → 走另一條
        call Race_Name_(怪獸id)                              ; sub_784A0
        test ch, ch                                          ; ch = 「是不是終點」
        jz   loc_7511B                                       ; 起點 → 結果清 0(而且不出訊息)
        sprintf(訊息 0x87, 星, 怪獸名)
        User_Box_(kind=1)                                    ; = Confirmation_Box_,是/否
    ```

    **條件是怪獸不是艦隊。** 這句錯誤斷言的來源大概是「集結點會把新艦送過去,所以危險的是
    目的地有敵艦」——一個合理但沒查證的推論。四條規則的完整對照(訊息編號也在)寫進了
    `relocation.go` 的檔頭。

    另外原版對**起點**的怪獸是**靜默拒絕**(`loc_7511B` 只把結果清 0)。remake 出一句話:
    在一個有滑鼠提示的介面裡,靜默失敗只會讓玩家以為按鈕壞了。

    ### 確認框的版面(全部是立即數)

    | 元件 | 來源 | 值 |
    |---|---|---|
    | 底框 | `sub_12B7E1(0A1h, 75h, CONFIRM#0)` | (161, 117),圖 313×227 |
    | Y 鈕 | `sub_12B7E1(0EBh, 12Eh, CONFIRM#1)` | (235, 302),圖 54×24 |
    | N 鈕 | `sub_12B7E1(159h, 12Eh, CONFIRM#2)` | (345, 302),圖 55×24 |
    | Y 熱區 | `sub_11438B(eax=0EBh, edx=12Eh, ebx=11Eh, ecx=143h)` | 235..286 × 302..323 |
    | N 熱區 | `sub_11438B(eax=159h, edx=12Eh, ebx=18Ch, ecx=143h)` | 345..396 × 302..323 |
    | 文字 | `sub_77A74(eax=0CCh, edx=0D0h, ebx=0E0h)` | 左緣 204、垂直置中 208、**折行寬 224** |

    兩個交叉驗證:①熱區(51×21)比圖(54×24)小一圈,兩者的左上角完全重合;
    ②文字塊中心 204 + 224/2 = 316,底框中心 161 + 313/2 = 317.5 —— 對得上。

    `Draw_Confirm_Box_` @ 0x778E4 每幀把兩顆鈕的幀號歸 0,再把**游標所在**那顆設成 1
    → 第 1 幀是 hover 高亮。

    ### 沒有還原的一條

    原版在文字放不下時會**縮字級**:`sub_103CAF` 量高度,`var_C` 從 4 遞減到 1,
    直到高度 ≤ 126。remake 的字型層沒有那組原版字級,改用固定字級 + 自行折行。
    寫在檔頭,不假裝有。

    ### 落點

    - `cmd/moo2/confirmbox.go`:widget(疊在下層畫面上,框外點擊無效——modal 的重點)
    - `internal/shell/relocation.go`:`RelocateToNeedsConfirm` + 起點的怪獸拒絕 + 檔頭訂正
    - `cmd/moo2/relocation.go` / `interactive.go`:`pendingConfirm` 接線
      (處理器手上沒有下層畫面,所以只記下來,由呼叫端換成畫面)
    - 截圖廊第 30 張 `29_confirm.png`(訊息用**真的規則產出的那句**,才驗得到折行)

69. **戰術格子的獨立戰機單位 + 一個讀錯的欄位**(2026-08-07)。

    ### 先訂正:那一欄是 Shots,不是「出擊數」

    `gamedata/combat.go` 寫著:

    ```go
    // FighterInterceptorSquadron 是一個攔截機戰機庫每次出擊的戰機數(手冊 GM p.127「出擊數」欄:攔截機 4)
    const FighterInterceptorSquadron = 4
    // FighterHeavySquadron 是一個重戰機庫每次出擊的重戰機數(手冊 GM p.127「出擊數」欄:重戰機 2)
    const FighterHeavySquadron = 2
    ```

    p.127 那張表的表頭是

    ```
    Weapon | Armament | Shots | Size | Cost | Speed | Hits | Strat Dmg
    ```

    第三欄是 **Shots**——**每架返航前開幾次火**。中隊規模在正文裡,而且寫了兩次:

    > p.157「All fighter craft are installed in ships and launched to a target in **squadrons of four**.」
    > p.83 「Heavy Fighters are installed and launched in **squadrons of 4**.」

    Shots 欄同樣有正文對照,逐項吻合:攔截機「**fire 4 times** at point-blank range」;
    重戰機「drop one bomb and fire a beam … then hover … to drop the other bomb and fire a
    beam again」= 2 次。

    也就是說舊值把「一架打幾次」當成了「一隊有幾架」——**重戰機庫因此少算了一半的戰機**。
    `docs/knowledge-base/manual-cht/03-combat.md` 的欄名也一起改掉了(那份 kb 是這個誤讀的源頭)。

    順帶確認一件沒錯的事:速度/血量欄與正文對得上。表上攔截機 Speed 8-20、重戰機 6-18,
    而正文說攔截機「speed 10」、重戰機「speed 8」——差 2。套 `CombatFighterSpeed` 就通了:
    範圍的下限是 FTL 0(base − 2)、上限是 FTL 6(base + 10)。血量欄的下限
    (攔截機 2、重戰機 5)也正是正文的「can take 2 / 5 damage」。**沒有第二個錯。**

    ### 戰機是一個兵種,不是一個加成

    remake 先前只有「戰機庫 → 母艦戰力 +N」。手冊給戰機的是一整套與艦艇不同的規則,
    而那些規則要能被看見,中隊就得在格子上有自己的位置:

    | 手冊 | 落地 |
    |---|---|
    | 「launched to a target in squadrons of four … cannot separate the fighters」 | 一隊是一個單位(一個 token),不是四個 |
    | 「fly to their target and use whatever weapons they have at point-blank range」 | `StepToward` 走到曼哈頓距離 ≤1 才開火,不像艦砲有射程 |
    | 「fighters will attempt to return to their carrier once they are out of shots」 | `ShotsLeft` 歸零 → `Returning` |
    | 「Once safely back, any surviving fighters get repairs, rearm, refuel, and can be launched again」 | `Recover()` 補血補彈——**但不補人**(手冊寫的是 any **surviving**) |
    | 「With the exception of Interceptors, fighter craft cannot engage one another」 | `CanTargetFighter()` 只有攔截機為真 |
    | 「Like missiles, fighter craft are vulnerable to beam weapons」 | 貼身的敵艦會把戰機打下來 |
    | 「Fighters have a 50% chance to avoid … spherical weapon」 | `FighterAvoidsSpherical(roll)` |
    | 「(Interceptors) can take 2 damage … (Heavy) can take 5 damage」 | 血量是**每架**的,傷害一架一架吃(不是整隊一條血條) |
    | 「base hit points are modified by 2 times armor level above Titanium」 | `FighterHitsWithArmor` |

    ### 誠實留白

    - **「always attack from the weakest shield facing」**:remake 的護盾是單一數值
      (`CombatShip.ShieldReduction`),沒有四面分別的護盾,這條無處可套。
    - **轟炸機 / 突擊梭**:前者要炸彈對行星的規則、後者要把陸戰隊送上敵艦,各自依賴另一套系統。
      這一輪只做兩種純對艦戰機(攔截機、重戰機)。
    - **敵方不會派戰機**:`genEnemyFleet` 產出的敵艦沒有設計資料,讀不到「帶不帶戰機庫」。
    - **FTL 階 / 裝甲級**:艦艇設計還沒把「目前最佳引擎/裝甲」餵進戰鬥層,出擊時先傳 1 / 0
      (手冊公式在「剛研究出 FTL、鈦裝甲」時的值)。接上來就換真值。
    - **出擊鈕不是原版版面**:原版的控制列是烘死的美術,那七顆鈕各有其原意,
      不能拿其中一顆假裝是出擊。先擺在標題列右側。

    ### 落點

    - `internal/gamedata/combat.go`:`FighterSquadronSize` / `FighterShots*` / `FighterHits*` /
      `FighterHitsWithArmor`,兩支貢獻函式依真值重算
    - `internal/shell/fighter.go`:中隊狀態機(11 支測試)
    - `internal/shell/session.go`:`CombatShip.Bay` / `BayKind`(與快速結算讀同一份設計資料)
    - `cmd/moo2/tacticalfighter.go`:每回合的目標選擇、推進、結算與繪製(6 支測試)
    - 截圖廊 `16_tactical.png` 換成「一隊攔截機正飛向敵艦」的畫面

70. **ALL 鈕根本不是集結點**(2026-08-07,把第 65 項的一個推測推翻)。

    ### 兩處手冊,同一句話

    第 65 項寫著「艦隊列表的 ALL 鈕接上 `Set_All_Star_Relocations_`」,並且自己標了「推測」。
    手冊在兩個地方各講了一次它到底是什麼:

    > p.32(Fleet Window)「To select or deselect all of the ships in the window, you can use
    > the **All** button.」
    > p.47(艦隊操作台的三顆鈕)「**All**: Selects all of the ships in the fleet to prepare to
    > receive orders. (If all the ships are already selected, this deselects them instead.)」

    括號那句是 **toggle** 語意——已全選就變成全不選,不是「按一次全選、再按一次還是全選」。
    p.47 同時給出那三顆鈕的完整清單:**All / Relocate / Scrap**,`Set_All` 不在其中。

    ### 那 Set_All / Clear_All 從哪裡進來?

    星圖的輸入處理器 `sub_73980`:

    ```
    cmp eax, 0FFFFFBAFh   ; −1105 → Clear_All_Star_Relocations_(玩家) + 訊息 0x76
    cmp eax, 0FFFFFC13h   ; −1005 → 切換 byte_19BED0(「下一次點星要 Set_All」模式)
                          ;         之後點星才呼叫 Set_All_Star_Relocations_ + 訊息 0x77
    ```

    那組負數 id 是**鍵盤**來的:同一支函式裡 −1002 / −1001 是被拿來與滑鼠 widget id
    **併列**判斷的替代鍵(`cmp ax, [var_40]` … `jz` … `cmp eax, −1002`),而 −1000 到處都是
    (「沒有事件」)。兩個相關的 id **差 100**,看起來是「某鍵」與「ALT+同一鍵」。

    **是哪一顆鍵沒有確認**(id → 鍵碼的對照表還沒追),所以不綁快捷鍵、不猜——
    與第 54 項對手冊 ALT+Fn 邊欄標籤的保留同一個立場。

    ### 落地

    - **ALL 鈕** → `toggleSelectAllShips`:全選/全不選。選取狀態本來就有(分艦隊用的就是它),
      所以接上去之後「全選 → 拆分」兩下就做得完。i18n 的譯名從「全部」改成「全選」。
    - **Set_All / Clear_All** → 名冊下方兩個**明確標示為 remake 自加**的入口
      (字前加「＋」,與原版烘在美術上的鈕區隔)。追出鍵碼之後改成星圖快捷鍵。

    ### 一個沒有照手冊改的地方

    p.47 說 Relocate 的終點要「click on another system **you've a colony in**」。
    但 `Okay_To_Set_Relocate_Star_` 對終點只檢查黑洞/已探索/怪獸確認,
    `Player_Has_Colony_In_System_` 那一條只在**起點**分支裡。
    **程式碼是實際行為,手冊那句更像是描述常見用法**——不改規則,記在這裡。

71. **「AI 的遷移設定」不是缺口**(2026-08-07)。

    ### 一個被假設出來的缺口

    這張表上寫著「AI 沒有逐星的艦隊位置,所以沒有遷移可設」,列為資料模型缺口。
    那句話的前半是對的(remake 的 `AIOpponent.FleetStrength` 是一個抽象數字),
    但**後半的結論要先確認原版的 AI 有沒有在用這個欄位**。

    ### 五個寫入者,沒有一個在 AI 那邊

    | 寫入者 | 呼叫端 |
    |---|---|
    | `Universe_Generation_` | 開局把 `[星 + 玩家×2 + 0x54]` 對 8 個玩家全部初始化成 −1 |
    | `Set_Relocation_` | 只有 `Star_Relocation_`(玩家的兩段點選) |
    | `Clear_Star_Relocation_` | 也只有 `Star_Relocation_`(點回同一顆 = 取消) |
    | `Set_All_Star_Relocations_` | 星圖輸入處理器 `sub_73980` + `Main_Screen_` |
    | `Clear_All_Star_Relocations_` | `sub_73980` |

    讀取端 `Redirect_Newly_Built_Ships_` **確實是逐玩家跑的**(收 player 參數、
    查 `Has_Relocation_(星, 玩家)`),所以 AI 的欄位有人讀——只是永遠是 −1。
    欄位之所以逐玩家,是因為星球結構本來就替 8 個玩家各留一格:**多人對戰時每個人類玩家用自己那格**。

    結論:**原版的 AI 也不設集結點**。remake 這邊什麼都不用做;要替 AI 加會是加一條原版沒有的規則。

    ### 方法上的一個坑(值得記下來)

    第一次找寫入者用的是 `grep '\*2+54h]'`,結果**漏掉了兩個**——
    `Set_All` / `Clear_All` 把 `星基 + 玩家×2` 先加好,再 `mov [eax+54h], bx`,
    定址式裡根本沒有 `*2`:

    ```
    movsx eax, si
    add   eax, eax          ; 玩家×2
    add   eax, ecx          ; + 星基
    mov   [eax+54h], bx     ; ← grep 的樣式對不上
    ```

    正確做法是先把 asm 切成一支支函式,再找「同時碰 `dword_19306C`(星表)、`71h`(stride)
    與 `+54h]`」的那些。那組條件把兩個漏網的都撈了回來——**這就是這個結論的正對照**
    (規則:`~/diagnosis-notes/docs/02-query-returned-empty`,下「不存在」的結論前先證明
    查詢本身找得到已知存在的東西)。

72. **決定性化——網路多人的地基,順手抓出兩個存檔 bug**(2026-08-07)。

    ### 為什麼先做這一塊

    網路對戰的三個部分:9 個畫面、傳輸層、**決定性化**。前兩個要等能連線才驗得到,
    但**決定性是規則層自己的性質**——現在就測得了,而且要先測:等傳輸層上線才發現
    規則層本身不決定性,除錯成本會高一個數量級(那時候每一次分岔都要先排除是不是網路問題)。

    ### 狀態指紋用存檔快照當正規形式

    `StateHash()` = SHA-256(`json.Marshal(snapshot())`)。三個理由:

    1. **不必另外維護欄位清單**——新欄位只要進得了存檔就自動進得了指紋;反過來說,
       進不了存檔的欄位本來就不該影響對局結果(那是 UI 狀態)。
    2. `encoding/json` 對 map 的鍵**保證排序**,`ColonyBuildings` 這類 map 不會因為 Go 的
       隨機迭代順序讓兩台機器算出不同指紋。
    3. 指紋不合時直接 diff 兩邊的存檔 JSON,看得出是哪個欄位分岔——這一輪就是這樣抓到 bug 的。

    ### 閘門測試抓到的第一個 bug:亂數流的位置沒進存檔

    三條長壽命亂數流(事件 / 星系發現 / 間諜)只記種子、不記「抽到第幾個數」,於是

    ```
    存檔 → 讀檔 → 繼續玩   ← 事件序列從頭開始
    ```

    兩個後果:①讀檔會重播同一批事件(**存檔洗事件變得毫無成本**);
    ②網路對戰時中途讀檔的那台會與其他人分岔。

    修法有個坑:直覺是「記下抽了幾次,讀檔時重抽幾次跳過去」,但 `math/rand` 的
    `Intn` 與 `Float64` **從底層 source 取走的數量不一樣**,所以「重抽 n 次」必須連
    **抽的種類**都一樣才會落在同一格。`randstream.go` 改成直接騎在 `rand.Source64` 上,
    **每次抽取恰好消耗一個 uint64**——於是「跳過 n 次」就只是丟掉 n 個原始值。

    ### 第二個 bug:主選單選的規則版本撐不過存檔

    修完亂數流,閘門測試**還是紅的**。diff 兩邊的存檔 JSON:

    ```
    < "HyperAdvancedResearchCost": 25000
    > "HyperAdvancedResearchCost": 0
    ```

    `RuleProfile` **完全沒進存檔**。讀檔後它是零值——那既不是 1.3 也不是 1.5,
    而是「Version=1.3 但所有數值欄位都是 0」的混種:Hyper-Advanced 研究成本、電漿砲傷害、
    轟炸輪數、守方 Commando 加成、感測器加成、貨運現金加成**全部歸零**。

    CLAUDE.md 把「允許在主選單選擇版本(1.3 / 1.5)」列為專案目標,而那個選擇撐不過一次存讀檔。
    修法:存**版本**不存整個 profile(profile 是衍生資料,存衍生資料會在新增欄位時悄悄留下舊值);
    舊存檔缺欄位 → 0 → 還原成**完整的** `Profile13()`。

    ### 這一輪的形狀值得記下來

    「先建一個能自動發現分岔的閘門,再讓它去找 bug」——兩個 bug 都不是靠讀程式碼找到的,
    是閘門紅了之後 diff 出來的。第二個尤其:修完第一個之後如果沒有那支測試,
    會直接以為做完了。

    ### 落點

    - `internal/shell/determinism.go`:`StateHash` / `StateFingerprint`
    - `internal/shell/randstream.go`:可快轉的亂數流
    - `internal/shell/persist.go`:三個抽取次數 + `RuleVersion`
    - `internal/shell/determinism_test.go`:7 支閘門(同種子逐回合、加玩家指令、存讀檔指紋、
      讀檔後續發展、快轉單元、指紋敏感度正對照、規則版本存讀檔 ×2)
    - ⚠ 亂數推導改了,所以截圖廊有 5 張的**數值**跟著變(版面沒動),一併更新

73. **傳輸層 + 鎖步協定**(2026-08-07,`internal/netplay`)。

    ### 原版的形狀(反組譯,不是猜的)

    `Net_Next_Turn_` @ 0xFC470 的骨架:

    ```
    byte_1AAF7E[本方玩家] = 1                  ; 標記「我這回合結束了」
    for 每個其他玩家:
        if 玩家還在線(`[player+0x28] == 0x64`):
            sub_F6816(...)                      ; 把自己的狀態送過去
    Wait_Until_Net_Opponent_Finished_ @ 0x3FBFE ; 等對手
    ```

    也就是**鎖步**:各自下完指令 → 廣播「我好了」→ 全部到齊才推進回合。
    `byte_1AAF76` / `byte_1AAF7E` 是逐玩家的旗標陣列(玩家結構 stride 0xEA9)。

    remake 照這個形狀做,但**傳輸換成 TCP + JSON**——原版走 IPX / 數據機 / 序列埠,
    那三種在現在的機器上都不存在。這是**移植決策不是還原**,標在 `frame.go` 檔頭。

    ### 三個設計決定,各有理由

    | 決定 | 理由 |
    |---|---|
    | `internal/netplay` **不相依** `internal/shell` | 傳輸層不該知道規則。兩層各自測得完,端到端那支放在**外部測試套件** `netplay_test`,同時 import 兩邊而不讓生產程式碼耦合 |
    | 4 位元組長度前綴 | TCP 是位元組流沒有訊息邊界:一次 `Write` 的東西可能被分兩次讀到,也可能與下一則黏在一起。上限 4 MB 不是為了省記憶體,是**不讓對面一個壞掉的長度欄位要求我們配置 4 GB** |
    | 指令**依玩家編號**排序 | 不是為了好看:鎖步要求每台機器以同樣順序套用同樣的指令。依到達順序套用會讓「誰的封包先到」影響結果——那正是 lock-step 最典型的分岔來源 |

    ### 逐回合比對狀態指紋(原版沒有的一層)

    鎖步的前提是「同樣的指令序列在每台機器上算出同樣的結果」。那個前提一旦破了,
    兩邊會**安靜地**漂開,幾十回合之後才以「你的畫面跟我不一樣」爆出來——
    那時候已經回推不了是哪一步歪的。每回合比一次(第 72 項的 `StateHash`),
    分岔的那一回合就是問題發生的那一回合。

    ### 端到端測試

    `TestTwoPeersStayInSyncOverAPipe`:兩個對等端各跑一份 `GameSession`,
    over `net.Pipe()` 交換 25 回合的回合封包,收齊之後依玩家編號套用同一批指令再結束回合,
    **每一回合的指紋都必須相同**。含一個正對照:整場下來至少要有一條指令真的產生效果,
    否則這支測試等於只驗了「兩個什麼都不做的 session 一樣」。

    另有一支走**真的 TCP socket**(`net.Listen("tcp", "127.0.0.1:0")`),驗框架讀寫換成真
    socket 也一樣。

    ### 還沒做的

    - **9 個畫面**:`Join_Net` / `Modem_Setup` / `NullModem_Setup` / `Choose_Net_Plyrs` /
      `Choose_Multi_Net_Game` / `Generic_Net_Info` / `SendGet_Net_Info` / `Net_Next_Turn` /
      `Wait_For_*`(版面座標都還沒抽)
    - **指令解譯器**:把 UI 的每一顆按鈕對到一條 `netplay.Command`。端到端測試裡只解了三條
      (派艦隊 / 排建造 / 設集結點),夠證明鏈是通的,但不是完整的指令集

74. **玩家指令層**(2026-08-07,`internal/shell/command.go`)。

    ### 三個用途,只有一個是網路

    | 用途 | 為什麼需要「指令」這一層 |
    |---|---|
    | 網路對戰 | 兩台機器要套用**同樣的指令序列**才會算出同樣的狀態 |
    | 回放 / 除錯 | 一局的完整指令序列 + 起始種子 = 可完整重現的 bug 報告 |
    | 熱座 | 其實已經在做同一件事,只是指令直接就地套用、沒有序列化 |

    ### 兩條刻意的界線

    **① 指令層不做前置檢查。** `ColonizePlanet` 之類的方法自己會回絕不合法的操作
    (艦隊沒到、行星已被佔…),指令層再檢查一次只會變成兩份會漂開的規則。
    指令層唯一的責任是「把名字與參數對到正確的方法」。

    **② 不認得的指令名一律報錯,絕不靜默忽略。** 靜默忽略在鎖步裡是最糟的處理:
    一邊套用了、另一邊沒有,而且沒有人會知道——幾十回合之後才以「你的畫面跟我不一樣」爆出來。
    參數不足則走預設值(那代表送出端有 bug,而規則層自己會擋掉無效操作);
    真正該停下來的是「指令名不認得」,那是**兩邊版本不一致**的訊號。

    ### 型別分開、形狀一樣

    `shell.PlayerCommand` 與 `netplay.Command` 的欄位形狀一模一樣,但是**兩個型別**——
    傳輸層不 import 規則層(第 73 項的界線)。轉換發生在同時認識兩層的地方:
    正式對局裡是 `cmd/moo2`(組裝端),測試裡是外部測試套件。

    ### 23 條指令 + 一致性測試

    `PlayerCommandNames()` 就是「網路對戰目前支援到哪」的唯一答案。三支測試守著它:

    - `TestEveryListedCommandIsHandled`:表上有的都認得,而且清單有排序、無重複
    - `TestCommandPathMatchesDirectCall`:走指令層與直接呼叫方法必須得到**一模一樣的狀態**
      (逐條比對 `StateHash`)——指令層只是轉接,不該有自己的規則
    - `TestEveryPlayerCommandCanTravel`(netplay 側):每一條指令都要能過線再被規則層認得,
      兩份清單漂掉的話會出現「單機做得到、連線不同步」的靜默缺口

    端到端測試因此換掉了原本那個只解三條的迷你解譯器,改用真的指令層。

    ### 還沒做的

    **9 個網路畫面**:`Join_Net` / `Modem_Setup` / `NullModem_Setup` / `Choose_Net_Plyrs` /
    `Choose_Multi_Net_Game` / `Generic_Net_Info` / `SendGet_Net_Info` / `Net_Next_Turn` /
    `Wait_For_*`。版面座標都還沒抽——`Draw_Net_Next_Turn_Screen_` @ 0xF1075 與
    `Add_Net_Next_Turn_Fields_` @ 0xEFCEA 是抽的起點。

75. **`Net_Next_Turn` 等待畫面——第一張版面是「算」出來的畫面**(2026-08-07)。

    ### 它與先前每一張都不一樣

    remake 移植過的畫面,座標都是反組譯裡的**立即數**(`sub_1151B0(eax=x, edx=y)` 那種)。
    這一張不是:`Load_Net_Next_Turn_Screen_` @ 0xF3E42 依**資產尺寸**現算。把那段翻成算式:

    ```
    x    = (0x280 − 資產42.寬) / 2              ; 水平置中於 640
    總高 = 資產42.高 + 資產43.高 + 資產40.高
    y    = max(0, (0x1E0 − 總高) / 2)           ; 垂直置中於 480
    [win+0xBF]  = 資產42.高 + 資產43.高          ; 第三塊的相對位移
    [win+0x10E] = 2                             ; 字型 id
    ```

    代進 lbxinfo 量到的尺寸(42 = 630×48 十幀、43 = 630×179、40 = 630×221):

    | 元件 | 位置 |
    |---|---|
    | 標題帶(10 幀動畫) | (5, 16) |
    | 中段面板 | (5, 64) |
    | 下段面板 | (5, 243) |

    `Add_Net_Next_Turn_Fields_` @ 0xEFCEA 再給兩個真值:輸入列
    `y = y + [win+0xBF] + 0xBB = 430`、高 `0x11 = 17`;玩家列間距 `0x19 = 25`。

    **測試釘的是算式不是數字**:`TestNetWaitLayoutMatchesTheOriginalFormula` 重算一次
    `(640−630)/2`、`max(0,(480−448)/2)`,並確認三塊是**上下相接**的(原版是一塊接一塊堆下去,
    不是各自定位)。資產換了或算錯了,這裡會紅。

    ### 誠實留白

    - **玩家列的起始 y 沒有算出來**:那一段把座標藏在 window 結構的欄位裡,沒有直接的立即數。
      間距用真值 25,起點取中段面板內縮一段,標成估計值。
    - **聊天輸入列**:原版 y=430 那一列是文字欄位,remake 沒有聊天——畫成提示帶並寫明。
    - **狀態指紋擺在畫面上**不是裝飾:分岔時兩邊念一下那八個字元就知道是不是同一個狀態,
      不必先架 log 收集。

    ### 三張畫面**不做**,而且不是因為做不動

    `Modem_Setup` / `NullModem_Setup` / `Comm Info` 是**數據機與序列線**的設定。
    那些硬體現在不存在,remake 的連線走 TCP——**替不存在的硬體做設定畫面不是還原,是裝飾**。
    多人設定畫面上那兩顆鈕現在會直說這件事,而不是含糊的「本版未實作」。

    剩下的 5 張(`Join_Net` / `Choose_Net_Plyrs` / `Choose_Multi_Net_Game` /
    `Generic_Net_Info` / `SendGet_Net_Info`)是**連線流程**的畫面,要等 UI 端的連線流程
    做出來才有東西可顯示——先做畫面會做出一堆沒有資料來源的空框。

76. **連線大廳 + `Choose_Net_Plyrs` 名冊——第一個「尺寸隨資料變」的版面**(2026-08-07)。

    ### 先做大廳,不先做畫面

    上一項結尾寫的就是這一項的前提:剩下 5 張是連線流程的畫面,沒有資料來源時做出來
    只是空框。所以這一輪先補 `internal/netplay/lobby.go`——**主機聽、客戶端連、主機廣播名冊**:

    | 角色 | 做的事 |
    |---|---|
    | 主機 `Host(addr, name, seed)` | 開 listener,自己是 0 號;`AcceptOne` 收一個人、指派 id、**廣播整份名冊** |
    | 客戶端 `Join(addr, name, timeout)` | 送 `hello` → 收 `roster` → 拿到自己的 id 與種子 |

    兩個設計決定,而且都不是隨手選的:

    - **玩家編號由主機指派。** 鎖步的指令排序鍵就是玩家編號(見第 73 項),
      各自取號會撞號,撞號就等於兩邊的指令順序不同 → 一定分岔。
    - **種子由主機決定並廣播。** 種子決定整張星圖與所有隨機事件;各自產生種子
      就不是同一局了。名冊訊息把種子一起帶過去,連上線就已經是同一個世界。

    ### 版面:第一個會長高的視窗

    `Choose_Network_Plyrs_Screen_` @ 0xF0E17 的定位段:

    ```
    x    = (0x280 − 資產27.寬) / 2
    總高 = 資產28.高 × [win+0x1E1] − 1 + 資產27.高 + 資產29.高
    y    = (0x1E0 − 總高) / 2
    ```

    `[win+0x1E1]` 是**列數**——中段面板(資產 28)每位玩家重複一次。先前移植的畫面
    版面都是固定的(不論立即數或第 75 項那種現算),這是第一個**尺寸隨資料變**的:

    | 人數 | 總高 | y |
    |---|---|---|
    | 1 | 36×1 − 1 + 81 + 38 = 154 | 163 |
    | 4 | 36×4 − 1 + 81 + 38 = 262 | 109 |
    | 8 | 36×8 − 1 + 81 + 38 = 406 | 37 |

    `Add_Choose_Net_Plyrs_Fields_` @ 0xEFB50 給每列的點擊區(逐項立即數):
    `x1 = winX + 0x6A`、`y1 = winY + i×0x24 + 0x40`、`x2 = winX + 0x1B3`、`y2 = y1 + 0x1D`
    ——每列 **329×29**,列距 36,**正好等於資產 28 的高**。那個相等是交叉驗證:
    列距若與面板高不同,兩邊必有一個抄錯。

    測試同樣釘算式:重算 x、逐人數重算 y、確認「人越多視窗越高」、確認列距等於面板高、
    確認 8 列(上限,`[win+0xBB]` 那個 widget id 陣列的長度)時整個視窗仍在 640×480 內。

    ### 截圖抓到的一個版面錯誤

    大廳狀態那兩行字(位址 / 種子)第一版畫在底框(資產 29)**裡面**——照 38 px 的高度算
    完全放得下。截出來才發現:那 38 px 的可見內容只有**頂端那圈金屬圓角**,底下是透明的。
    結果第一行壓在圓角上、第二行掉到視窗外面。

    這一條讀程式讀不出來,因為「資產高 38」在數字上完全合理。**版面的驗收是看圖,
    不是算數字**——同第 65 項那條紀律。修法是把兩行移到底框**下方**、加一條擦底,
    並補 `TestChooseNetPlayersInfoLinesSitBelowTheWindowAndStayOnScreen` 把它釘住
    (1~8 列都要在畫面內)。

    ### 誠實留白

    - **沒有文字輸入框,所以「加入」只連得上本機。** 要連別台得先做輸入框——
      這是下一步,不是這一輪偷懶;`netLobbyDialAddr` 的註解寫明了。
    - **不能點列指派種族。** 原版這張畫面可以(`sub_EFABA` 在每列旁再建一組欄位),
      remake 的大廳只做到「誰連進來了」,種族仍走既有的單機流程。
      要接的話得把種族選擇整段納入連線流程,不做半套。
    - **沒有重連、沒有心跳、沒有加密。** 這是區網對戰的最低限度,寫在 `lobby.go` 檔頭。

    剩下 4 張(`Join_Net` / `Choose_Multi_Net_Game` / `Generic_Net_Info` / `SendGet_Net_Info`)。

77. **連線狀態面板——反組譯把「還缺 4 張」改寫成「還缺 1 張」**(2026-08-07)。

    ### 抽版面時發現那 4 張其實是 2 張

    ```
    0xF19C7  Draw_Generic_Net_Info_Screen_
    0xF19C7  Draw_Join_Net_Screen_          ← 同一個位址
    ```

    兩個名字指向同一段程式碼。往上追,`Reload_Generic_Net_Info_` @ 0xF53D7 收一個
    **資產編號**當參數,而下面這一整排都只是帶不同編號呼叫它:

    | 呼叫者 | 資產 | 尺寸 | 意思 |
    |---|---|---|---|
    | `Reload_Waiting_For_Joiners_Screen_` @ 0xF552A | 15 | 479×150,10 幀 | 等其他玩家加入 |
    | `Reload_Join_Net_Screen_` @ 0xF54CF | 23 | 478×70,4 幀 | 加入對局中 |
    | `Reload_Wait_For_Race_Info_` @ 0xF551B | 24 | 480×116,4 幀 | 等種族資料 |
    | `Reload_Initializing_Net_Info_` @ 0xF54BE | 25 | 478×70,4 幀 | 初始化連線 |
    | `Reload_Sending_Data_Info_` @ 0xF54D9 | 26 | 411×105,4 幀 | 傳送資料 |
    | `Reload_Generating_Map_Info_` @ 0xF53CB | 30 | 478×70,5 幀 | 產生星圖 |
    | `Reload_Getting_Data_Info_` @ 0xF54A0 | 31 | 443×105,4 幀 | 接收資料 |

    也就是「`Join_Net`」「`Generic_Net_Info`」「`SendGet_Net_Info`」不是三張畫面,
    是**同一張畫面的三個狀態**。照著畫面名清單一張一張做,會做出七份幾乎一樣的程式碼;
    追到共用的那個 loader 才看得出真正的形狀是「一個面板 + 一個狀態列舉」。

    版面又是算的:`x = (0x280 − 資產.寬)/2`、`y = (0x1E0 − 資產.高)/2`,字型 id 4。
    `Draw_SendGet_Net_Info_Screen_` @ 0xF2C8B 另給進度數字的兩組位移
    ——`[win+0x10F]==0`(傳送)→ (+0x72,+0x42)、`==1`(接收)→ (+0x79,+0x41),
    兩者共用同一段繪製,只差那幾個像素。

    ### 這一輪修掉兩個自己犯的錯,兩個都是截圖抓到的

    **一、把 `Add_Waiting_For_Joiners_Field_` 讀成「已加入人數」欄位。**
    截圖上那串數字壓在 `START NET GAME` 上,才回去查它呼叫的 `sub_1151B0` 是什麼——
    符號表寫著 **`Add_Button_Field_`**。那個 (+0x9E,+0x6A) 是**按鈕**的左上角,
    把資產 15 攤開來看正好對上。名字裡的 "Waiting_For_Joiners" 指的是「這顆鈕加在哪張畫面」,
    不是它顯示什麼。**符號名是二手推論,被呼叫的函式是一手事實。**

    **二、LBX 多幀動畫是 delta 幀,逐幀獨立上色會讓整張面板消失。**
    資產 15 的第 0 幀是完整面板,第 1~9 幀只帶會閃的那幾顆燈。remake 的
    `multigmFrameWithKey` 一直是逐幀獨立解碼——這個 bug 先前沒被發現,是因為截圖廊
    每一張都恰好落在第 0 幀,直到這張刻意多留幾拍才露出來。

    修在 `internal/lbx`(`AccumulatedUpToRGBA`)而不是這一個畫面:同樣的坑對
    資產 27(名冊標題帶)、42(等待畫面標題帶)都成立,只是還沒播到。
    測試造一張 2×1 兩幀的合成圖直接比對「逐幀上色 vs 累積上色」——
    錯的那個在畫面上是「什麼都沒有」,不做正對照就只會看到一片黑而不知道為什麼。

    ### 誠實留白

    - remake 目前只有「等待加入」這個狀態有觸發點(主機開大廳 → 這一張 → 點過去進名冊)。
      其餘六個狀態的資產與版面都對好了,**不是死碼,是連線流程還沒走到那幾步**。
    - 「加入對局中」永遠不會停留:`netplay.Join` 是同步的,連上或逾時都在同一個呼叫裡結束。
      原版有那一段,是因為它的連線走 IPX / 數據機,協商要好幾秒。
    - 已加入人數的位置是**量的**,不是反組譯真值——原版沒有給那個欄位。

    **剩 1 張**:`Choose_Multi_Net_Game`(版面已抽出,見下輪)。

78. **區網對局探索 + `Choose_Multi_Net_Game`——最後一張網路畫面**(2026-08-07)。

    ### 這張畫面的資料從哪來,原版沒有回答

    原版走 IPX,而 IPX **自帶**廣播式的服務公告——「列出區網上有哪些對局」是**協定**給的能力,
    不是遊戲自己做的。remake 走 TCP,TCP 沒有這個能力。照抄畫面而不補那一層,
    會得到一張永遠空的清單:畫面是畫出來了,但那不是還原。

    所以先補 `internal/netplay/discovery.go`:主機每秒往 `255.255.255.255:24502` 廣播
    一份「我在這裡」(名字、TCP 位址、人數),客戶端聽同一個埠、依 TCP 位址去重。
    **這一層是移植決策,不是還原**,標在該檔檔頭。

    三個實作決定:

    - **來源 IP 覆蓋封包裡寫的**:主機常常不知道自己對外是哪個位址(多網卡、容器內、NAT)。
      廣播來源的 IP 比它自己寫的可信,只有埠沿用封包裡的。
    - **清單依名稱排序,不依到達順序**:到達順序每次都不一樣,而順序決定玩家點到哪一場。
    - **`Browser` 不阻塞**:UI 是單執行緒的,`Discover` 那種「收兩秒再回傳」會讓畫面凍住兩秒
      (同 `lobby.AcceptOne` 的處境)。背景 goroutine 收,畫面每幀讀快照。

    測試全部走 **127.0.0.1**,不走真的廣播位址——測試不該把封包送到辦公室網路上。
    含一支正對照:同一輪送三份壞封包 + 一份合法的,證明過濾器不是把全部都丟掉。

    ### 版面

    `Load_Choose_Multi_Net_Game_Screen_` @ 0xF40D3:

    ```
    x = (0x280 − 資產41.寬)/2                    = (640−479)/2 = 80
    y = ((0x1E0 − 資產41.高) − 0x51)/2 + 0x25    = ((480−384)−81)/2 + 37 = 44
    ```

    ⚠ 那個 `−0x51`(81)剛好等於標題帶(資產 27)的高,但這張畫面**沒有畫標題帶**
    ——`Draw_Choose_Multi_Net_Game_Screen_` 只 blit 資產 41。也就是它是版面上的讓位,
    不是「上面還有一塊」。**照抄數字,不照抄自己對數字的解讀。**

    `Add_Choose_Multi_Net_Game_Fields_` @ 0xEFF87 給每列的熱區
    (`sub_11438B` = `Add_Hidden_Field_`,隱形熱區——美術已經畫好格子了):
    `x1=+0x26 / x2=+0x190 / y1=+0x40+i×0x1B / y2=y1+0x16`,即每列 **362×22**、列距 27、
    最多 10 場。底下一顆 `Add_Button_Field_` 在 (+0xBF, +0x158)。

    `Draw_Choose_Multi_Net_Game_Screen_` @ 0xF1AF4 再給文字的擺法:
    `x = +0x26+9`、`y = winY + var_C + (0x16 − 字高)/2`(var_C 起始 0x43,每列 +0x1B)
    ——**字在 22 px 的列裡垂直置中**。選中的那一列另有脈動亮度(`[win+0x1E8]` 在 −3..+4
    之間來回),配色從 (0x95,0x97,0x91) 換成 (0x97,0x99,0x91),就是整組色往上推兩階。

    ### 又一個截圖抓到的錯:上緣不是基線

    第一版把原版的 `(0x16 − 字高)/2` 加了一個字高當基線傳給 `uifont.Draw`,結果**整欄字
    掉到下一列**——截圖上選取框在第一列、字在第二列。原因是 `uifont.Draw` 底層是
    ebiten **text/v2**,`GeoM.Translate(x,y)` 的 y 是**行框上緣**(v1 才是基線)。
    原版的算式本來就是上緣,多加那一下是自己加的。

    測試除了重算算式,另釘一條:**第 i 列的字不得落進第 i+1 列的熱區**——那正是這個 bug 的症狀。

    ### 誠實留白

    - UDP 廣播只跨得過同一個廣播網域(同一個區網)。跨網段要打對方位址,那要文字輸入框。
      **原版的 IPX 也是同一個限制**,所以這不是退步。
    - 沒有簽章、沒有加密:區網上任何人都能廣播一個假的對局。與大廳本身的立場一致。
    - 原版可以在這張畫面**改對局名稱**(`Change_MP_Game_Name_` @ 0xF5777:長度上限 8、
      且要與既有對局不同名)。remake 還沒有輸入框,名稱取玩家名的前 8 字元;
      上限與唯一性的規則已記在 `netplay.GameNameMax`,做輸入框時直接套。

    **9 張網路畫面到此結案**:6 張做了(`MP_Setup` / `Hotseat` / `Net_Next_Turn` /
    `Choose_Net_Plyrs` / 狀態面板 7 狀態 / `Choose_Multi_Net_Game`),
    3 張明確不做(`Modem_Setup` / `NullModem_Setup` / `Comm Info` —— 硬體已不存在)。
    網路多人剩下的是**文字輸入框**(跨網段直連 + 改對局名 + 聊天列都等它)。

79. **文字輸入彈窗——原版有一個,而且有自己的 LBX**(2026-08-07)。

    ### 先查被呼叫的函式叫什麼

    remake 先前一路寫著「原版的輸入是內嵌欄位、remake 沒有輸入框」,那個判斷是錯的。
    `Change_MP_Game_Name_` @ 0xF5777 呼叫 `sub_91BB4`,而符號表寫著
    **`Remapped_Input_Box_Popup_`** ——原版有一個獨立的 modal 彈窗,連自己的 LBX 都有
    (`INBOX.LBX`,只有兩個資產:0 = 288×151 底框、1 = 98×28 的 ACCEPT 鈕)。

    這是這幾輪第三次靠「**符號名是二手推論,被呼叫的函式是一手事實**」修正判斷
    (前兩次:第 77 項的 `Add_Waiting_For_Joiners_Field_` → `Add_Button_Field_`、
    第 78 項的清單列 → `Add_Hidden_Field_`)。

    ### 版面

    `sub_91F14` 把呼叫端給的 (x, y) 攤成一組欄位,全部是立即數:

    | 欄位 | 值 | 意思 |
    |---|---|---|
    | `word_19C387` | y + 3 | 標題帶 y |
    | `word_19C384` | 0x36 = 54 | 標題帶高(字在裡面垂直置中) |
    | `word_19C38D/8F` | x+0x22, y+0x36 | 輸入欄 (x+34, y+54) |
    | `word_19C39B` | 0x1A = 26 | 輸入欄高 |
    | `word_19C389/8B` | x+0x60, y+0x64 | ACCEPT 鈕 (x+96, y+100) |
    | `word_19C399` | min(呼叫端上限, 0xCD) | 長度上限,硬上限 205 |

    `sub_91BD4` 再給輸入欄寬 = 資產0.寬 − 0x36 = **234**,標題**水平置中於彈窗寬**。

    ⚠ 輸入欄左邊距 34、右邊距 288−34−234 = **20**——不對稱。0x22 與 0x36 是兩個獨立的
    立即數,沒有理由相等。**照抄,並且寫一條測試防止有人「順手改成對稱」**。

    位置真值:`Star_Name_Popup_Screen_Center_X_` @ 0x923BE = 0xB1 = 177、
    `_Y_` @ 0x923C4 = 0x7D = 125。x 幾乎正好是水平置中((640−288)/2 = 176),
    y=125 則明顯高於垂直置中(164)——那是原版選的位置,不是置中算式。

    ### 接上去的兩處(都是原版本來就有的)

    - **改對局名稱**:主機按「開始新遊戲」→ 先問名稱(上限 8,`Change_MP_Game_Name_` 的
      `edx`)→ 才開大廳。原版就是這個順序,而且名稱是別人在清單上看到的東西,
      開完局才改就晚了。
    - **加入指定位址**:清單畫面多一顆「直接輸入位址」。⚠ **原版沒有這顆鈕**——
      IPX 自己找得到,不需要打位址。remake 的 UDP 廣播只跨得過同一個區網,
      跨網段就得能打位址。所以它擺在清單**外面**的空白帶,不佔用任何一個原版座標。

    ### 誠實留白

    原版的輸入處理走 `sub_91B89` 逐鍵掃描碼(IME 之前的年代)。remake 用 ebiten 的
    `AppendInputChars`——**移植決策**:掃描碼那一套在現代平台拿不到,而且會擋掉輸入法。
    代價是原版的某些鍵行為(插入模式之類)沒有還原。游標閃爍週期是自己訂的,原版沒抽。

    測試釘版面立即數、釘那個不對稱邊距、釘長度上限(含 rune 不是 byte:上限 8 要能打
    8 個中文字)、釘空字串退格不 panic、釘控制字元不進緩衝區。

    **網路多人到此完整**:傳輸層 + 鎖步 + 決定性 + 指令層 + 大廳 + 區網探索 + 6 張畫面
    + 輸入框。剩下的是**聊天列**(`Chat_Box_Input_Loop_` @ 0xF55A4 已定位,
    `Send_Chat_Msg_` @ 0xDD3B8 是送出端),那是加分項不是缺口。

80. **TECH LEVEL 的第二個真效果——那張「拿不到所以不臆造」的表挖到了**(2026-08-07)。

    ### 缺的是表,而 `shell.TechLevels` 的註解自己說了

    > 開局已知科技領域數:Pre-warp 2 個、其餘 6 個。remake 現在開局是 2 個 = **等同 Pre-warp**,
    > 所以預設選 Average 時其實還沒拿到該有的 6 個領域。要補得先查出那 6 個是哪些
    > (手冊只說預設第一個是 field #29),**沒有一手表之前不臆造**。

    也就是選單上寫著「一般」,拿到的卻是曲速前的科技——而且沒有任何錯誤訊息。

    ### `Init_Player_Tech_` @ 0x5E55F 給了兩樣

    **送幾個**由 `byte_199CB5`(NEW GAME 的 TECH LEVEL,來自 `word_1A1360`)決定:

    ```
    al == 0 → var_18 = 1
    al == 1 → var_18 = 6
    al == 2 → var_18 = 0x19 = 25
    ```

    **送哪些**在 `word_18111C` @ 0x18111C:`dw 1Dh, 37h, 16h, 39h, 1Ch, 17h`
    → **29, 55, 22, 57, 28, 23**。主迴圈前 6 次取這張固定表,第 7 次起改由 `sub_FD335`
    隨機挑——所以 25 級是「六個固定 + 十九個隨機」,不是二十五個固定。

    對到 remake 的主題編號:

    | 原版 | id | 主題 |
    |---|---|---|
    | 1Dh | 29 | `TOPIC_ENGINEERING` |
    | 37h | 55 | `TOPIC_NUCLEAR_FISSION` |
    | 16h | 22 | `TOPIC_CHEMISTRY` |
    | 39h | 57 | `TOPIC_PHYSICS` |
    | 1Ch | 28 | `TOPIC_ELECTRONICS` |
    | 17h | 23 | `TOPIC_COLD_FUSION` |

    ### 三方互證(不是一條線的推論)

    - **手冊**獨立說「預設的第一個是 field #29」——remake 先前就把這句記在註解裡了。
    - **反組譯**的 `word_18111C[0] = 0x1D = 29`。
    - **第二個 55 = 核分裂**,而 remake 早就把 `FTLTopic` 定成 `TOPIC_NUCLEAR_FISSION`
      (核融合引擎在那一層)。手冊說 Average「已具備星際航行所需的全部科技」——
      兩條獨立的線指到同一個編號。

    ### 接線時抓到的一個陷阱

    `applyStartingTech` 一開始只「加」不「減」,結果 `NewDemoSession` 用預設等級發過的
    核分裂會留在曲速前的局裡——**正好是「曲速前不該有 FTL」這條規則的反例,而且不會有
    任何錯誤訊息**。改成「先把固定表裡這一級不該有的清掉,再發該有的」,並補一支正對照測試
    (`TestPrewarpDoesNotKeepTheDefaultLevelsFTLTopic`:先斷言 demo 局本來就有,再斷言改成
    曲速前之後沒有)。

    **AI 一起發**:原版的 `Init_Player_Tech_` 第一個參數就是玩家編號,是逐玩家跑的。
    只發給玩家等於把 AI 永遠留在曲速前。

    ### 驗收看的是畫面不是測試

    截圖廊有 9 張變了,而且變得對:科技總覽從「已完成主題 2 項」變成 **7 項**
    (固定表 6 個 + field 0),建造清單多出**運輸艦隊**——它的前置正是
    `TOPIC_NUCLEAR_FISSION`。這是端到端的證據,測試綠不是。

    ### 誠實留白

    - 先進級在固定表之外**還會隨機送 19 個**,remake 當時只發固定表的 6 個。
      缺口大小由 `gamedata.StartingTopicRandomExtras` 回報——**讓缺口是一個看得見的數字,
      不是註解裡的一句話**。
      ⚠ **2026-08-07 第 99 項補上了**(照抄 `sub_FD335` 的挑選結構),那個函式現在的角色
      從「缺口大小」變成「還要再發幾個」。
    - 初始建築數上限(3/5/9)仍未接:要先有「依人口生成母星建築」的機制。
    - 原版 `byte_199CB5 >= 3` 那條路徑**沒有設 `var_18`**(`enter` 不清堆疊)。
      NEW GAME 只給三級,那個值進不來——**照抄一個未初始化的堆疊值不叫還原**。

81. **開局建築的優先清單——手冊只給了結論,清單本身在執行檔裡**(2026-08-07)。

    ### 缺的又是一張表,而程式碼自己標了

    `shell.StartingBuildingCount` 的註解:

    > 此函式只回傳「上限」,實際會生成哪些建築仍取決於 **initial_buildings 優先清單**與已知科技

    那份清單在 `Init_Homeworld_Colony2_` @ 0x13A3D:

    ```
    loc_13CCB:
        movsx ebx, word_17D8AC[ecx]      ; ← 優先清單(word 陣列,0 收尾)
        test  ebx, ebx / jz 結束
        movzx eax, byte_199CB5            ; TECH LEVEL
        movzx ax, ds:byte_13A3A[eax]      ; ← 上限表
        cmp   ax, [已放幾棟] / jle 結束
    ```

    **上限表 `byte_13A3A` = `db 3, 5, 9`** —— 與手冊逐字相同(「capped to 3 for Pre-warp,
    5 for Average/Postwarp and 9 for Advanced」)。remake 的 `BuildingCapPreWarp/Average/Advanced`
    早就是這三個數,現在有了第二個來源。

    **清單 `word_17D8AC`**(32 個原版建築編號,順序照原版):

    ```
    41  8  40  21  22  15  7  20  37  4  31  33  34  39  2  12  13  32
    35  10  43  16  28  25  18  19  24  26  47  27  6  5
    ```

    開頭 41 → 8 → 40 是 **Star Fortress → Battlestation → Star Base**:同一條防禦升級鏈,
    **最強的排最前面**。原版照這個順序往下發到上限為止,所以科技越先進拿到的是越上層那一棟。

    ### 四條獨立的線指到同一個答案

    手冊說「Pre-warp and Average Tech games only start with **Marine Barracks and a Star Base**
    because no other techs are Known that are also in the default initial buildings list」。

    拿這份清單 × 第 80 項的六個開局主題 × remake 自己的建築前置表跑一遍——
    清單裡科技條件成立的**正好只有 Star Base(40)與 Marine Barracks(22)**。
    清單、主題表、前置表、手冊那句話,四個獨立來源互相對上。

    落成測試 `TestAverageStartGivesExactlyMarineBarracksAndStarBase`:三張表任何一張抄錯都會紅。

    ### 驗收:截圖廊**零差異**

    這一輪把寫死的兩棟換成「從清單算」,而一般等級算出來仍然正好是那兩棟
    ——截圖廊 34 張逐位元組相同。行為保持的證明不是測試綠,是那個零。
    另有一支正對照 `TestHomeworldBuildingsMatchTheHardCodedPair`:新舊兩條路要走到同一個答案,
    只有新的那條綠證明不了它對。

    ### 誠實留白:缺口被釘在上一層

    **先進級目前仍然只有兩棟**,而且原因不在這一層:上限 9、⌈⅔×8⌉ = 6,名額有 6 個,
    但清單裡科技條件成立的只有兩棟——因為第 80 項留下的 **19 個隨機主題還沒發**。

    `TestAdvancedStartIsBlockedByTheMissingRandomTopics` 把這件事釘住,並附正對照:
    科技全解時這套機制**確實**會發滿 6 個名額。**機制是對的,缺的是上游的科技。**

    ⚠ **2026-08-07 更新**:第 99 項把那 19 個接上、第 100 項讓這一層讀真正的科技集合之後,
    先進級**真的發滿 6 棟**了。那支測試已改名為 `TestAdvancedStartFillsAllBuildingSlots`
    ——**當年寫的正對照預測了現在的結果**,不是事後改斷言。

    那 19 個要接得先港 `Choose_Tech_Application_` @ 0xFD335 ——294 行的 AI 科技權重選擇器
    (權重吃成本表 `dword_17D916`、種族旗標 `byte_17E084`、性格 `[player+0x28]`、
    政府別、`sub_FC845` 的估值)。**一次讀就照抄風險太高**,留作獨立一輪。

    順帶把 `origBuildingID` 從 `cmd/moo2/colonysurface.go` 搬進 `internal/gamedata`:
    畫地表 sprite 與這份優先清單要靠**同一份**編號對照,各抄一份遲早會漂開。

82. **飛彈速度:那個「手冊自相矛盾」不是矛盾,是漏了條件**(2026-08-07)。

    ### 原本的斷言

    `internal/gamedata/missile.go` 的檔頭從移植那天起就寫著:

    > ⚠ 手冊此段自相矛盾:明列公式 Speed = BaseSpeed(12) + 2*(FTLlevel-1) + FastBonus(4)
    > 得 14/16/…/26,但同段附表 Speed 欄為 10/12/…/22。本檔以「明列公式」為準,
    > 表格 Speed 欄推測為驅動本身速度(另一量)。**此落差需日後對實機行為動態驗證。**

    而 `docs/HONEST-STATUS.md` 把「飛彈速度」列在「需原版 oracle 對照」那一條裡。

    ### 反組譯一次解掉,而且不需要實機

    `Missile_Speed_` @ 0x3CD21 的最後三行:

    ```
    loc_3CE40:
        test [ebp+var_3], 10h     ; ← 旗標 0x10
        jz   short loc_3CE49
        add  edx, 4               ; 只有旗標成立才 +4
    ```

    那個 `+4` 是**有條件的**——旗標 0x10 是飛彈的 **Fast 改造**。所以:

    | 手冊的哪一段 | 對應什麼 |
    |---|---|
    | 附表 10/12/…/22 | **沒有** Fast 改造的一般飛彈 |
    | 明列公式(含 +4) | **裝了** Fast 改造的飛彈 |

    **兩者都對,只是條件不同。** 手冊沒說那個 FastBonus 是有條件的,於是讀起來像矛盾。

    remake 先前**無條件 +4**,等於每一枚飛彈都當成裝了 Fast 改造:
    Beam Defense 憑空高 20(5×4),飛彈比原版難打下來。

    ### 同一支函式還推翻了「基礎速度是 12」

    `Missile_Speed_` 依**武器類型**分檔:

    | 類型 | 基礎 | 加 FTL 項? | 玩家旗標成立時 |
    |---|---|---|---|
    | 0x0E..0x11 | 12 | 是 | — |
    | 0x12 / 0x13 | 20 | **否** | — |
    | 0x1C | 6 | 是 | 10 |
    | 0x1D / 0x1E | 8 | 是 | 12 |
    | 0x1F | 10 | 是 | 14 |
    | 0x28 | 24 | **否** | — |

    12 只是其中一檔——正好是手冊拿來舉例的那一檔,所以手冊的表對「那一檔」是準的。
    另外兩檔在原版是 `xor ecx, ecx`,**速度與驅動等級無關**;那一步很容易漏抄,
    所以測試特地釘住「這兩檔在 FTL 0 與 FTL 6 要相等」。

    ### 誠實留白

    `[player+0x8BC]` 那個玩家旗標(成立時 6→10、8→12、10→14)**還沒追出是什麼**。
    `MissileBaseSpeed` 留一個 `boosted` 參數、呼叫端目前一律傳 false ——
    **留一個誠實的參數,比把它寫死成「永遠沒有」再假裝完整好**。
    0x12/0x13 與 0x28 是哪些武器也還沒對出名字,不編。

    ### 這一項的教訓與第 63 條同源

    「手冊自相矛盾 → 選一個 → 標註待實機驗證」在文件裡放了很久,而正確答案一直在執行檔裡。
    **`docs/HONEST-STATUS.md` 把它列進「需原版 oracle 對照」是分類錯誤**:
    它不需要 oracle,它需要的是把靜態來源窮盡完。已從那一列移出。

83. **地面戰:結構不是「未核實」,是抄了一代的**(2026-08-07)。

    ### 原本的斷言

    `internal/gamedata/ground_battle.go` 的檔頭:

    > 解算結構取自**一代(1oom)**game_ground.c 的 game_ground_kill

    `docs/HONEST-STATUS.md`:「force 值用 MOO2 手冊表但**結構本身未對 MOO2 實機核實**」。

    ### 原版的一方是 26 個位元組,排得剛剛好

    `Ground_Combat_Round_` @ 0xEC4FE + `Resolve_Ground_Combat_` @ 0xEC601:

    ```
    +0x00  word     ?
    +0x02  word[4]  各部隊類型的攻擊力
    +0x0A  word[4]  各部隊類型的剩餘單位數
    +0x12  byte[4]  各部隊類型的耐受命中數
    +0x16  byte     當前部隊類型(0..3;4 = 全滅)
    +0x17  byte     當前類型的累積命中數
    +0x18  byte     本回合被打中但沒死的類型(0xFF = 無)
    +0x19  byte     本回合死了一個單位的類型(0xFF = 無)
    ```

    欄位剛好排滿 0x00..0x19 沒有空隙——這本身就是「**四種**部隊類型」的佐證,
    而 `cmp byte ptr [ebx+16h], 4 / jnb` 與迴圈條件 `A.type < 4 && B.type < 4` 是另外兩條。

    一回合:

    ```
    strA = A.攻擊力[A.當前類型] + Random(100)
    strB = B.攻擊力[B.當前類型] + Random(100)
    if (strA <= strB) A 挨一下
    if (strA >= strB) B 挨一下      ; ← 兩個獨立的 if
    ```

    ### 與一代結構的三處實質差異

    | | remake(一代結構) | 原版 |
    |---|---|---|
    | 平手 | `if/else` → **只有攻方**挨打(「平手歸守方」) | 兩個獨立的 `if` → **雙方都**挨打 |
    | 攻擊力 | 整隊一個 `Force` | **逐部隊類型**各一個 |
    | 傷害 | 每次 −1、扣到 `<= 0` 陣亡 | 累積命中,**`==` 耐受值**才陣亡,然後歸零 |

    第一點是**可觀察的機率差異**:d100 對 d100 平手是 1%,那 1% 在原版是雙方各損失一次,
    守方原本白拿的優勢沒有了。

    第三點的 `==`(`cmp cl, [ebx+eax+12h] / jnz`)在耐受值中途被改小的邊界上會表現不同
    ——照抄,不「改良」成 `>=`。

    ### 順帶消掉一段技術債

    `ground_invasion.go` 先前有一整段在解釋「為什麼把戰車營排在合併陣列尾端」:
    因為 `ResolveGroundBattle` 只回傳一個總存活數,要靠 `min(總存活, 戰車原始數)` 推算分兵種。
    原版的結構逐類型記數量,**戰後直接讀 `Count[類型]` 就是真實存活數**——
    那一整段說明連同它的 TODO 一起刪掉了。

    ### 誠實留白

    - **四種部隊類型沒有對出名字。** MOO2 的地面單位是陸戰隊 / 裝甲 / 機械戰士三種,
      第四種是什麼還沒追。remake 只用到兩種(陸戰隊 = 0、戰車營 = 1),
      所以只給這兩個常數——**編一個「類型 2 = ???」的名字比留白更糟**。
    - **每種部隊各自的攻擊力表還沒追**(`[side + type*2 + 2]` 的來源)。兩種目前都填同一個
      `atkForce`:填同值 = 維持現行數字,並把差異留在一個看得見的地方。
    - `ground_battle.go` 的舊解算**保留**(加成表與 builder 還在用,而且它是那三處差異的對照組),
      但檔頭已標明「新的解算不要再從這裡分支」。

84. **四種部隊類型是什麼——手冊那一句把三種對上了**(2026-08-07)。

    第 83 項留下「四種部隊類型沒有對出名字、每種的攻擊力表還沒追」。同一天追完。

    ### `Compute_Ground_Combat_Info_` @ 0xEC3CE 的四個 case

    ```
    case 0:  攻擊力 += 10 + 加成塊[+1] ; 耐受 += 1 + 加成塊[+2]
    case 1:  攻擊力 +=      加成塊[+3] ; 耐受 +=     加成塊[+4]
    case 2:  攻擊力 -= 10
    case 3:  攻擊力 -= 20              ; 基礎值取自**另一方**的加成塊(為 0 時整格歸零)
    ```

    `Compute_Colony_Ground_Combat_Info_` @ 0xED713 給殖民地填**三格**數量
    (`[+0x0A]`/`[+0x0C]`/`[+0x0E]` = 類型 0/1/2),第 4 格留 0(它傳 `ebx = 0`)。

    ### 手冊補上名字

    > 「Along the bottom left of the view are icons representing all the **Marine and Armor**
    > units stationed in defense of this planet. … In addition to Marine and Armor units your
    > **militia** are also shown here.」

    殖民地正好三種防守單位,而調整量把強弱排出來了:

    | 類型 | 調整 | 是什麼 |
    |---|---|---|
    | 0 | +10 攻擊、+1 耐受 | **裝甲**(手冊:tank battalions) |
    | 1 | 基準 | **陸戰隊** |
    | 2 | −10 攻擊 | **民兵**(未受訓的平民,最弱) |
    | 3 | −20,基礎取自另一方 | ⚠ **仍未定名**,殖民地不填它 |

    ### 順帶抓到一個順序錯

    remake 先前把陸戰隊排在類型 0、戰車營排在 1 —— **反的**。先前兩種填同一個攻擊力,
    所以順序不影響結果;接上逐類型的差之後就會差 10 點。已訂正,並加測試釘住
    「原版的類型 0 是裝甲」。

    ### 誠實留白

    - **只實作立即數的部分**。`加成塊[+1]`/`[+3]`(科技加成)那兩欄還沒對出意義,
      不含在調整量裡——**回一個「差不多」的值會讓日後追出真值時看不出哪裡被污染過**。
    - **守方的民兵沒有接**:數量公式在 `sub_EC61E`,還沒追;AI 也沒有裝甲營房的追蹤機制。
      那兩格留 0 = **少算守方兵力**,方向上對玩家有利——是明說的偏差,不是隨手的預設。
    - 類型 3 不編名字。

85. **民兵接上了——`Colony_N_Militia_` 是「除以 5」**(2026-08-07)。

    第 84 項留下「守方的民兵沒有接:數量公式在 `sub_EC61E`,還沒追」。同一天追完。

    ### `Colony_N_Militia_` @ 0xEC61E

    ```
    ecx = [colony+0x0A]            ; 人口
    eax = colony + 0x0C + 人口×4    ; 人口單位陣列的尾端(每個單位 4 位元組)
    往前掃每一個單位:
        if (([eax] & 0x0F) >= 8)      跳過   ; 低 4 bits 是擁有者編號
        if ([eax+1] & 0x04)           跳過   ; 某個旗標
        計數++
    return 計數 / 5
    ```

    兩個跳過條件都是**人口單位上的資料旗標**,不是規則:

    - `擁有者 >= 8`:玩家編號只有 0..7。`Init_Homeworld_Colony2_` 寫入時就是
      `and ebx, 0Fh` 把玩家編號塞進低 4 bits ——**同一個結構在兩支函式裡對得上**,
      這是「4 位元組一個人口單位」這個判讀的佐證。
    - `[+1] & 0x04`:初始化時那個位元是被清掉的(`and [eax+0Dh], 0F9h / or 2` 只設 bit 1),
      正常人口不會有它。是什麼設的還沒追。

    remake 的殖民地人口沒有逐單位的擁有者/旗標模型——每一格都是自己的、都能打,
    所以兩個條件恆不成立,結果就是 **⌊人口 / 5⌋**。
    等哪天有了逐單位模型(異族人口、奴隸…),這裡才需要真的去掃。

    ### 接進防守方

    守方現在是**陸戰隊 + 民兵**兩格(民兵攻擊力比陸戰隊低 10,見第 84 項的 case 2)。
    `DefenderStart` 的回報數字也跟著含民兵——不改的話畫面上會少報守方兵力。

    ⚠ **裝甲那一格仍留 0**:AI 沒有 `ColonyBuildings` 追蹤機制,無法判斷「AI 是否已建成
    裝甲營房」。沒有資料可誠實推導守方戰車數,不臆測。留 0 = 少算守方,方向上對玩家有利。

    ### 這一項改變了平衡,而且是往忠實的方向

    守方憑空多出 ⌊人口/5⌋ 個單位(母星 8 人口 → +1)。既有的入侵測試
    (`TestInvadeColony_StrongAttackerWinsMost` / `..._StrongDefenderWinsMost`)**仍然綠**,
    表示偏移沒有把勝負翻過去,只是把守方的下限抬起來——原版本來就是這樣。

86. **地面戰的加成塊——其中一條手冊完全沒寫:難度加成不給玩家**(2026-08-07)。

    第 84 項留下「那幾個加成塊欄位還沒逐欄對出意義」。追
    `Compute_Player_Ground_Combat_Bonuses_` @ 0xEC15C(產一個 **19 位元組**的加成塊)之後,
    大多數欄位對應的是手冊已經列出的加成類別(remake 的 `GroundArmorTechBonus` 等已經涵蓋),
    但**有兩條是手冊沒有的**。

    ### 一、基礎耐受命中數是 1,某個科技讓它變 2

    ```
    [out+0x0C] = 1  if [player+0x8AA] != 0
    耐受命中數 = [加成塊+0x0C] + 1          ; Compute_Ground_Combat_Info_
    ```

    也就是**預設一下就死一個單位**;那個科技讓所有部隊都變成要挨兩下。

    ### 二、地面戰的難度加成**不給人類玩家**

    ```
    if ([player+0x28] == 100) [out+0x0F] = 0        ; 100 = 人類玩家的標記
    else                      [out+0x0F] = 難度 − 2
    ; 玩家編號 >= 8 的那一側(安塔蘭 / 怪獸)走另一條路徑:
                              [out+0x0F] = 難度×2 − 4
    ```

    `[out+0x0F]` 被加進**所有部隊類型**的攻擊力。所以:

    | 參戰者 | 加成 |
    |---|---|
    | 人類玩家 | **0** |
    | AI 帝國 | 難度 − 2(普通 = 0、不可能 = +2、教學 = −2) |
    | 安塔蘭那側 | 難度×2 − 4(**恰好是 AI 的兩倍**) |

    兩點值得記:

    - **加成是以「普通」為基準往兩邊偏**,不是「難度越高一律加成」。教學難度下 AI 是**負的**。
    - `[player+0x28] == 100` 這個人類玩家標記在 `Init_Player_Tech_` @ 0x5E55F 也出現過
      (`cmp byte ptr [eax+28h], 64h`)——**同一個標記在兩支不相干的函式裡對得上**,
      這是判讀正確的佐證。

    已接進入侵流程:守方(AI)加 `難度 − 2`,攻方(人類玩家)**不加**——
    那不是漏掉,是原版就沒有,註解寫明以免日後被人「補上」。

    ### 誠實留白

    `[+5]`/`[+7]`/`[+9]` 那三張查表(stride 15 / 3 / 3,索引由 `sub_DC323`/`sub_DC416`/`sub_DC449`
    算出)、`[+0x0B]`、`[+0x10]` 還沒逐欄對出意義。它們對應的是手冊已經列出的那幾類加成,
    而 remake 已用手冊的表算過——**不重複實作**,免得同一個加成被加兩次。

87. **重力種族特性:High-G 逐字對上,Low-G 的「10%」其實是定值 −10**(2026-08-07)。

    第 86 項留下的加成塊欄位又追出三個,而且三個都能與手冊互證。

    ### 那個 `else` 就是「互斥」的證據

    ```
    cmp byte ptr [player+8AAh], 0     ; High-G
    jz  short loc_EC227
    mov byte ptr [out+0Ch], 1         ; → 耐受命中數 +1
    jmp short loc_EC234
    loc_EC227:
    cmp byte ptr [player+8A9h], 0     ; Low-G(**只有在不是 High-G 時才看**)
    jz  short loc_EC234
    mov byte ptr [out+0Dh], 0F6h      ; → 攻擊力 −10
    ```

    手冊明寫「High-G World and Low-G World are **mutually exclusive**」——
    而原版把互斥直接寫成 `if / else if`。**兩邊互證**。

    ### High-G:手冊逐字

    > 「High-G ground troops can sustain substantially more physical damage than other troops;
    > **they take 1 hit more than normal troops before being slain in ground combat.**」

    對上 `mov byte ptr [out+0Ch], 1`,而耐受命中數 = `[out+0x0C] + 1`
    (第 86 項)→ 一般 1 下、High-G 2 下。**一字不差**。

    ### Low-G:手冊寫「10%」,原版是**定值 −10**

    `mov byte ptr [ecx+0Dh], 0F6h` —— `0xF6` 是有號位元組的 **−10**,是個定值。
    它與其他所有加成一起加進攻擊力(`var_4`),而那些加成本身也都是 +10/+15/+20 這種定值。

    手冊那個「%」多半是行文上的隨手寫法(其他加成手冊也是寫「adds 10 to…」)。

    remake 先前照字面做成乘法(戰力 × 90%),註解裡還寫著「手冊未列出 10% 套用在哪個
    基準值、如何捨入」——**那個不確定性現在有答案了**。差異在典型戰力(30..60)下是
    3..6 點 vs 固定 10 點,不是可以忽略的捨入差;戰力 7 時舊版甚至因整數除法完全不扣。

    ⚠ 測試也跟著改寫:舊測試裡 `100 → 90` 這一列**兩種算法的答案剛好相同**——
    只測那一個數的話,這個改動驗不出來。新測試加了 50/10/7/0 與「定值 = 差與基準無關」的性質。

    ### Subterranean:從單一來源升級為雙來源

    ```
    cmp byte ptr [player+8ACh], 0     ; Subterranean
    jz  short loc_EC247
    cmp [ebp+var_4], 0                ; ← 呼叫端傳的旗標
    jz  short loc_EC247
    mov byte ptr [out+0Eh], 0Ah       ; → +10
    ```

    「只有守方才給」在原版是那個呼叫端旗標,而 `Compute_Colony_Ground_Combat_Info_`
    (殖民地 = 守方)傳的正是 1。數字 10 與「僅守方」兩個條件都對上 remake 既有的
    `GroundSubterraneanDefenseBonus`——這一條從「手冊單一來源」升級為雙來源,沒有改動。

    ### 誠實留白

    `[player+0x8A7]`(有號,加進所有類型)看起來就是種族的地面戰加成
    (remake 的 `GroundRaceCombatBonus`:布拉西 +10 / 諾蘭姆 −10),但**沒有直接證據**
    指出那個位元組就是它——不寫進程式碼,只記在這裡。
    `[+5]`/`[+7]`/`[+9]` 三張查表與 `[+0x10]` 同樣仍未定義。

88. **三張查表讀出來了——十二個科技 id 全部對上,而且 remake 少了一整條通道**(2026-08-07)。

    ### 索引函式的符號名直接說了是什麼

    | 函式 | 表 | 步幅 | 階數 |
    |---|---|---|---|
    | `Player_Best_Armor_` @ 0xDC323 | `word_17F63E` | 15 | 6 |
    | `Player_Best_Rifle_` @ 0xDC416 | `word_14A88` | 3 | 5 |
    | `Player_Best_Personal_Shield_` @ 0xDC449 | `word_14A9A` | 3 | 1 |

    三支都是**從表尾往前找第一個已知的科技**(`[player + 科技 + 0x117] == 3`)——
    也就是「取最高階」,不是加總。

    ### 讀表的方法:先建 VA → 檔案位移的對照

    `aMultigmLbx` 後面緊接著 `byte_17A061`,所以那個字串的 VA = `0x17A061 − 12`;
    在 exe 裡搜 `"MULTIGM.LBX\0"` 得到檔案位移 `0x1F86E9` → **delta = 0x7E694**。
    用另一個同名字串(`aMultigmLbx_0`)反推得 VA `0x178004`,落在 `;org 178000h` 之後 4 位元組
    ——**對得上,delta 可信**。

    ### 讀出來的三張表

    ```
    裝甲(word_17F63E,+0 科技 id、+3 加成)
      187 TECH_TITANIUM_ARMOR    +5     ← 手冊沒列
      191 TECH_TRITANIUM_ARMOR   +10    ← 手冊:adds 10
      203 TECH_ZORTRIUM_ARMOR    +15    ← 手冊:adds 15
      117 TECH_NEUTRONIUM_ARMOR  +20    ← 手冊:adds 20
        2 TECH_ADAMANTIUM_ARMOR  +25    ← 手冊:adds 25
      201 TECH_XENTRONIUM_ARMOR  +30    ← 手冊:+30

    步槍(word_14A88,+0 科技 id、+2 加成)
      145 TECH_PULSE_RIFLE       +0
      101 TECH_LASER_RIFLE       +5
       73 TECH_FUSION_RIFLE      +10
      128 TECH_PHASOR_RIFLE      +20
      138 TECH_PLASMA_RIFLE      +30

    個人護盾(word_14A9A)
      124 TECH_PERSONAL_SHIELD   +20    ← 手冊:by 20
    ```

    **十二個科技 id 全部對上 remake 的 `Technology` 列舉**,而裝甲那六項的上五項與手冊
    逐字相同、個人護盾也與手冊相同——這不是巧合,是「這三張表就是它們」的證明。

    ### 於是抓到兩個實質缺口

    - **鈦裝甲 +5 少了**。手冊沒列基礎裝甲的地面加成,remake 就回 0。
      鈦裝甲是**開局就有的**,所以那 5 點是每個帝國、每一場地面戰都少的。
    - **整條步槍通道 remake 完全沒有**。上限差 **30 點**——後期科技全開的帝國,
      remake 的地面部隊比原版弱整整 30。已補 `GroundRifleTechBonus` + `groundRifleBonusFor`
      並接進玩家與 AI 的 force。

    ### 順帶把兩個「給誰」訂正了

    加成塊的另外三個科技旗標也解出來了(`[player + 科技 + 0x117]`,減去 0x117 就是科技 id):

    | 位址 | 科技 id | 科技 | 寫進哪一欄 | 被誰讀走 |
    |---|---|---|---|---|
    | `[player+0x120]` | 9 | `TECH_ANTIGRAV_HARNESS` | `[out+0]` | **所有類型**共用的基礎 |
    | `[player+0x12F]` | 24 | `TECH_BATTLEOIDS` | `[out+1]`/`[out+2]` | **只有類型 0(裝甲)** |
    | `[player+0x1A7]` | 144 | `TECH_POWERED_ARMOR` | `[out+3]`/`[out+4]` | **只有類型 1(陸戰隊)** |

    也就是 Battleoids 是**裝甲專屬**的 +10 攻擊 **+1 耐受**(手冊只提了 +10),
    而動力裝甲是**陸戰隊專屬**的。remake 先前把這兩項都加給整支部隊。
    常數與依據已記進 `gamedata`(`GroundBattleoidExtraHits`、`GroundPoweredArmorAppliesTo`),
    **分兵種的接線留給下一輪**——這一輪先把「少了 35 點」補回來。

    ### 誠實留白

    `[player+0x8A7]`(有號,加進所有類型)與 `[+0x10]` 仍未定名。

89. **分兵種接線 + 手冊的四個 hits 數字全部被重建出來**(2026-08-07)。

    第 88 項留下「Battleoids / 動力裝甲的分兵種接線留給下一輪」。接完了,而且過程中發現
    **手冊列的那四個 hits 值,可以完全由反組譯的加法結構重建**。

    ### 這是整個地面戰模型正確性的最強證據

    手冊只給了四個成品值,反組譯給的是「基礎 + 分類型 + 分科技」的加法。兩邊算出來一模一樣:

    | 手冊值 | 反組譯的組成 | 結果 |
    |---|---|---|
    | 陸戰隊 1 | 基礎 `[+0x0C]+1` = 1 + 類型 1 的 delta 0 | **1** |
    | 陸戰隊 + 動力裝甲 2 | 1 + `[out+4]` = 1 | **2** |
    | 戰車營 2 | 1 + 類型 0 的 delta 1 | **2** |
    | 機械戰士 3 | 1 + 1 + `[out+2]` = 1 | **3** |
    | High-G +1 | `[+0x0C]` = 1 → 基礎變 2,全類型各 +1 | ✓ |

    四個獨立的手冊數字,由三個獨立的反組譯欄位加出來——**這種吻合不會來自誤讀**。
    落成 `TestManualHitValuesReconstructFromTheOriginalStructure`。

    ### 於是也發現一個算兩次的坑

    remake 的 `tankHitsToKillFor` 回的是**手冊的成品值**(戰車 2、機械戰士 3),
    而第 84 項接上的 `GroundTypeHitsDelta` 是**組成的一部分**。兩個一起用就變成 3 / 4。
    已改成只用成品值,並在該處註解寫明為什麼不再加 delta。

    ### 分兵種接線

    | 加成 | 原版寫進 | 被誰讀走 | remake 先前 |
    |---|---|---|---|
    | Anti-Grav Harness +10 | `[out+0]` | 所有類型 | ✓ 對 |
    | Personal Shield +20 | `[out+7]/[out+8]` | 所有類型 | ✓ 對 |
    | **Powered Armor +10** | `[out+3]` | **只有陸戰隊** | ✗ 加給整支部隊 |
    | **Battleoids +10** | `[out+1]` | **只有裝甲** | ✗ 加給整支部隊 |

    已拆成 `groundMarineOnlyBonusFor` / `groundTankOnlyBonusFor`,共用的那份不再包含它們。

    ### 順帶消掉一個「為了繞過錯誤而存在」的守門

    舊的 `tankForceBonusFor(ps, tankCount)` 有一個「`tankCount > 0` 才給」的判斷,
    註解寫著「0 輛戰車的一方不該白拿這個加成」。那個守門存在的理由,正是**加成被加進整側的
    force**——而那本身就是錯的。現在加成落在戰車那一格上,沒有戰車時那格本來就是空的,
    守門自然不需要了。**修好根因之後,補丁會自己掉下來。**

90. **聊天列補完:14 則的環、82 byte 的格子、兩種前綴,四個數全是一手的**(2026-08-07)。

    「等待其他玩家」那張畫面的輸入列先前是一條寫著「remake 未實作」的提示帶。
    做得動的理由是第 79 項把文字輸入框做出來了,而這一輪把原版那條線整條讀完:

    | 函式 | 位址 | 讀出什麼 |
    |---|---|---|
    | `Chat_Box_Input_Loop_` | 0xF55A4 | 點輸入列 → 進聊天模式;非空才送;送完清空重新武裝 |
    | `Send_Chat_Msg_` | 0xDD3B8 | 逐一走玩家陣列(stride 0xEA9),`edx = 27h` 是封包型別 |
    | `Receive_Chat_Msg_` | 0xDD351 | 環的結構:滿 14 則 memmove 掉最舊、每則 82 byte |
    | `sub_F1075` 的繪製段 | — | 兩種前綴、行距 12、x +24、首行 y +14 |

    ### 四個數字,每個都指得出出處

    | 值 | 出處 | 意思 |
    |---|---|---|
    | 14 | `cmp dword ptr [eax+47Ch], 0Eh` | 保留幾則 |
    | 82 | `imul edx, [eax+47Ch], 52h` | 每則佔幾 byte |
    | 80 | 82 − 發話者 1 − NUL 1 | 一則最多幾個字元 |
    | 8 | `cmp ax, 8 / jge`(繪製段) | 到這個編號以上是 GNN 新聞不是玩家 |

    計數欄位落在 `[+0x47C]` 自己就是第二條線:**14 × 82 = 1148 = 0x47C**
    ——陣列剛好塞滿到那裡,計數接在後面。不是同一個數字換句話說。

    ### 版面也自己對上了

    繪製段給的是相對偏移(x +0x18、首行 y +0x0E、行距 0x0C、每行擦 0x23A × 0x0B),
    套進這張畫面的資產 40(y=243):

    ```
    首行 257 → 14 行 × 12 → 最後一行 413,底部 424
    Add_Net_Next_Turn_Fields_ 給的輸入列:430
    ```

    **中間剩 6 px。** 偏移量出自繪製端、輸入列出自另一支欄位註冊函式,兩邊不知道彼此,
    結果嚴絲合縫——這是第二個獨立來源。落成 `TestChatLayoutFitsAboveTheInputRow`。

    ### 兩種前綴照抄,包含空格數

    ```
    發話者 < 8  → "(%s)  %s"      ; 右括號後**兩個**空格
    發話者 ≥ 8  → "( GNN )  %s"   ; 括號內側各一格,固定色 byte_199F34 = 0x10
    ```

    `( GNN )` 是 7 個字元,多數玩家名短於此——兩個空格讓內文起點對得比較齊。
    順手改成一個空格會讓兩路的內文錯開,所以照抄。

    ### 一個 remake 必須偏離原版的地方

    上限 80 是**緩衝區的 byte 數**,不是設計選擇。原版是單 byte ASCII,切在哪裡都合法;
    UTF-8 切在半個中文字上會變亂碼。`ChatTruncate` 守住 80 byte 但截在 rune 邊界
    ——中文一則約 26 字。那不是 remake 加的限制,是原版那個 82 byte 的格子。

    ### 誠實留白

    - `Send_Chat_Msg_` 只發給 `[player+0x28] == 'd'` 的玩家。**沒查到那個欄位的寫入端**,
      不知道 0x28 是什麼、為什麼是字母 d,所以 remake 不照抄這個判斷,改成發給所有已連線的對手。
      不編一個名字給它。
    - 送出目前**只進本機記錄**:鎖步的 `netplay.Table` 一回合只收一則,聊天是隨時可送的,
      塞進同一條線會壞掉鎖步。真接上連線時多一個 `WriteFrame(conn, ChatMessage(...))` 即可,
      `ChatLog` 這一端不必動。
    - 玩家列的顏色仍是 remake 自訂的兩色,沒接 `Get_Net_Next_Turn_Player_Colors_` @ 0xF31BB
      ——那支依帝國旗色配色,要先有網路對局的旗色名冊。

    ### 順帶:截圖廊的驗收覆蓋率補齊

    比對時發現 `docs/screenshots/` 只有 27 張,而 gallery 產 35 張——
    **八張從來沒進過版控,byte-diff 驗收對它們等於沒跑。** 其中七張(事件、安塔蘭室、
    高分榜、遊戲選單、片頭、結局、熱座)兩次跑完全一致,已補進版控;
    `18_loadgame` 帶存檔時間戳,兩次跑不一樣,刻意不收——收進去只會製造假警報。

91. **整棵研究樹從二手轉寫升格成一手驗證過**(2026-08-07)。

    `techtree.go` 的檔頭一直寫著「逐字轉寫自 openorion2 的 `tech.cpp`」。也就是說
    remake 的 83 個主題成本、每個主題的可選科技,證據等級**全部是二手**
    ——openorion2 讀錯了 remake 就跟著錯,而且沒有任何測試會發現。

    這一輪把同一張表從原版執行檔挖出來對了一次。

    ### 表怎麼找到的

    起點是第 88 項刻意延後的 `Choose_Tech_Application_` @ 0xFD335。讀它的過程中
    `Init_Player_Tech_` @ 0x5E55F 的發科技迴圈把兩張表的位址與欄位全交代了:

    ```
    imul ecx, topic, 17h
    movsx eax, word_17D90C[ecx]        ; 主題表 stride 23,+0 = 下一個主題
    mov   byte [player+topic+0C4h], 3  ; 本主題「已完成」
    mov   byte [player+eax+0C4h],   2  ; 下一個主題變「可研究」
    ...
    imul eax, tech, 0Dh
    mov  ax, word_17E07F[eax]          ; 科技表 stride 13,+0 = 屬於哪個主題
    ```

    成本欄位由 `sub_FD335` 的 `mov eax, dword_17D916[eax]` 定出來(+10, dword)。

    ### 對照結果

    | 項目 | 結果 |
    |---|---|
    | 83 個主題的成本 | **74 個逐字相同**,9 個有差(全部有解釋) |
    | 主題 → 科技的歸屬 | **81 個主題完全相同**,逐科技 **199 條吻合** |
    | remake 有、原版沒有的科技 | **0 個** |
    | 領域表 vs next 鏈 | **73 條銜接關係全中** |

    最後一列是最強的一條:openorion2 把樹寫成「8 個領域各一串主題」,原版執行檔寫成
    「每個主題一個後繼」的鏈。**兩種編碼互不知情,73 條銜接關係逐條吻合。**

    ### 九個不同的,全部有解釋

    **八個 Hyper-Advanced 主題:15000 vs 25000,是真的版本差異。** 三份執行檔各查一次:

    ```
    1996 原版      → 15000
    patch 1.31     → 15000   (整張表與 1996 版 byte-identical)
    patch 1.50.26  → 25000   (只有這 8 筆變,主題 74 仍是 15000)
    ```

    remake 的 `RuleProfile` 早就寫著 1.3 = 15000、1.5 = 25000,出處是社群
    `CHANGELOG_150.TXT` 1.50.9「Hyper-Advanced Tech Cost Bug」。現在那兩個數字
    **在三份執行檔裡各自有一手來源**,不再只靠 changelog 的一句話。

    **主題 74 XENON:原版寫 15000 + 8 個科技,openorion2 寫 0 並註「always unavailable」。**
    那句註解是對的,而且原版的編碼方式很乾脆:

    ```
    next[74] == 74          ← 後繼是它自己
    沒有任何別的主題的 next 指向 74
    ```

    主題要變可研究,唯一的路是別人的 next 指到它;指向自己就等於要先有它才能有它。
    **自環就是「永遠解不開」的寫法。** 那 8 個科技(黑洞產生器、阻尼力場、死光、粒子束、
    量子引爆器、反射力場、空間壓縮器、贊特隆裝甲)是安塔蘭專屬,打來的不是研究來的。

    ⚠ 主題 0 的 next 也是 0,但 0 同時是「鏈到此為止」的哨符(8 個 Hyper 主題都是 0),
    **分不出是自環還是終點就不當證據用**——主題 0 不可研究的理由是別的(開局科技的容器)。

    ### 順帶修掉一個真的錯誤

    `shell.StarterResearchTopics` 是一份手挑的 9 個「新手可選的早期主題」。
    拿原版的樹一比,那份清單在開局那一刻是錯的:

    ```
    不該出現(領域裡前面還有沒研究完的):進階建築學、進階生物學、人工智慧、進階化學
    漏掉了(該領域的隊首):        太空生物學、核融合物理學、光電子學
    ```

    原因是原版的研究**每個領域是一條線**,同時只有隊首那一個能選。已改成
    `AvailableResearchTopics(session)` 由樹算出來(`gamedata.AvailableTopics`)。
    ⚠ `-game` 主路徑的 `currentAreaTopic` 本來就是對的(2026-07-11 修盲選 bug 那次改的),
    這次的 `next` 鏈剛好是它的一手佐證;錯的是 `play.go` 那張簡易殼的研究畫面。

    ### 又一次撞到同一組六個主題

    `sub_FD2F9` 檢查六個硬編位址是否都等於 3,全中才回 1。那六個位址減掉 `0xC4`:

    ```
    0xDA→22  0xDB→23  0xE0→28  0xE1→29  0xFB→55  0xFD→57
    ```

    **正好是第 80 項從 `word_18111C` 挖到的六個開局主題**,第三個獨立來源。
    而且這次連角色都說清楚了:AI 的科技權重要等這六個全部完成才會啟用一般模式,
    之前只給這六個評分。

    順帶對上第 89 項:`OrigStartingTechs` 裡有 `TECH_PULSE_RIFLE`,而 remake 的
    `GroundRifleTechBonus` 把脈衝步槍定為 +0 基準點——先前的理由只是「它最低階」,
    現在有硬理由:**它在開局科技裡,每個帝國第一回合就有**,所以必然是零點。

    ### 誠實留白

    `sub_FC845`(逐科技估值,`sub_FD335` 的權重來源)**是 985 行**。上一輪判斷「一次讀就
    照抄風險太高」,這次量到了數字,結論不變。先進級開局的 19 個隨機主題當時仍沒接
    ——但這一項讓「缺的到底是什麼」變清楚了:一個吃「成本 ÷ 每回合研究點 = 幾回合」、
    再乘上種族/性格權重的加權隨機挑選器,而且權重那一半在那 985 行裡。

    ⚠ **2026-08-07 第 99 項接上了**:那句拆解正是接線的依據——除了權重那一項,
    其餘(候選集合、成本反比、視野放寬、加權隨機)全部照抄。`sub_FC845` 仍不照抄。

92. **三面行星護盾 + 自動實驗室 + 再生反應爐接線**(2026-08-07)。

    HONEST-STATUS 寫著「部分軍事/防禦建築(~13 棟,需艦隊駐防/軌道防禦系統先落地)」。
    照 rulebook 63 對程式碼盤點(掃建築名有沒有在 `buildings.go` 以外被消費過),
    實際是 **11 棟**,而且其中**有三棟根本不需要新子系統**——它們接的軌道轟炸早就有了。

    ### 先前為什麼沒接

    `fleetBombardDamage` 的註解自己講了:

    > 沒有「行星護盾」資料(damage.go DamageAfterShield 明講「本函式只處理艦對艦,
    > 行星護盾情境不適用」),故護盾/裝甲一律視為 0(無防禦)。

    那不是建模選擇,是缺資料。手冊三段各給了一個數字:

    | 建築 | 手冊原文 | 減傷 | 手冊維護費 | 建築表(來自執行檔) |
    |---|---|---|---|---|
    | Planetary Radiation Shield | reducing bombardment damage by 5 points | 5 | 1 BC | **1** |
    | Planetary Flux Shield | reduces all damage … by 10 points **per attack** | 10 | 3 BC | **3** |
    | Planetary Barrier Shield | reducing all damage … by 20 points **per attack** | 20 | 5 BC | **5** |

    右邊兩欄是第二個來源:減傷值與維護費出自手冊的**同一段文字**,而維護費三棟
    全部對得上執行檔的建築表——**那段文字可信,減傷值不是孤證**。

    ### 「per attack」決定了接在哪一行

    接在**逐發**傷害(`shot.DamageToStructure`)而不是總傷害。在 10 輪齊射下這兩個接法
    差一個數量級,而且用 hits 驗測不出來——`GroundBombHitsFromDamage` 除以 100 會把差異吃掉。
    所以測試釘的是 `TotalDamage`:10 輪 × (101 − 減傷),手算得出來。

    ### 取代不是疊加

    手冊每一段都寫了取代關係(「A Planetary Flux Shield **replaces** any Planetary
    Radiation Shield already in existence」),所以 `PlanetaryShieldReduction` 取**最強的那一面**。
    資料上真的同時出現兩棟時(存檔亂了),取最大值才是還原;加總會讓資料異常變成強化。

    ### 再生反應爐:接對地方比接上去重要

    手冊 p.81 兩句話缺一不可:

    > each unit of population generates 1 industrial production, **regardless of its assigned job**
    >
    > This increased production does **not count toward the planetary pollution level**

    第二句決定了接的位置。`FlatIndustry` 是在污染縮減之**前**併進 gross 的
    ——接那裡會讓這份產能跟著產生污染,**正好是手冊否定的那句**。
    改成一個旗標,在 `RunColonyTurn` 的污染切分點之後才加。

    測試也照這兩句話拆成兩個獨立斷言,外加一條**正對照**:
    同樣多出來的產能若接成 `FlatIndustry`,污染一定會變——那條在,才證明「污染沒變」
    不是因為這個殖民地本來就不污染。

    ### 自動實驗室

    手冊 p.96「generating 30 research points per turn」,一個數字、沒有 per-scientist 敘述,
    所以只動 `FlatResearch`。

    ### 誠實留白

    - 三棟護盾都寫著「Radiated 氣候轉 Barren」,屏障護盾還多一句「生物武器無法進入大氣層」。
      前者要接殖民地氣候欄位、後者要有生物武器這個分類,**這一輪只接減傷,並在檔頭寫明**。
    - 剩下 8 棟未接:食物複製機(要產能↔食物轉換管線 + 每單位 1 BC)、
      阿提米絲系統網(要艦體等級觸發機率 + 護盾等級,是完整的水雷子系統)、
      太空學院(要艦員經驗值子系統)、異族管理中心(同化子系統)、
      戰機基地(戰術格子戰鬥的獨立戰機單位)、恆星轉換器,以及先前已知的兩棟。
      **這些是真的需要新子系統,不是沒接線。**
    - `30_netwait.png` 變了:狀態指紋是存檔快照的 SHA-256,`ColonyState` 多一個欄位就會變。
      那是 `determinism.go` 註解寫明的設計(「新增欄位只要進得了存檔就自動進得了指紋」),
      畫面上其餘每一個像素都相同。

93. **食物複製機接線:一句話裡的三個限定詞,漏一個就是印鈔機**(2026-08-07)。

    手冊 p.85 一整句就是完整規格:

    > convert industrial production into food on a **two-for-one basis**
    > at a cost of **1 BC per food**, **as needed**.

    | 限定詞 | 規則 |
    |---|---|
    | two-for-one | 2 產能 → 1 食物 |
    | 1 BC per food | 換出來的每單位食物再花 1 BC(從國庫,不是從產能) |
    | **as needed** | **只補足缺口**,不換出盈餘 |

    ### 最後那條是整棟建築的平衡

    漏掉「as needed」,一個有複製機的殖民地會把全部產能換成食物,再靠既有的
    **餘糧出售**(`IncomeFoodSurplusRevenue`,每單位 0.5 BC)換回 BC。
    2 產能 → 1 食物 → 0.5 BC,而稅收那條路 1 產能也只換得到稅率比例的 BC——
    在高產能低稅率的局面下,那會變成一台比稅收更好賺的印鈔機。**原版沒有這個東西。**

    所以測試裡有一條專門釘它:有食物盈餘時,產能與盈餘兩個數字都必須**完全不動**。

    ### 維護費 10 BC 是它的設計,不是隨手訂的

    | 來源 | 維護費 |
    |---|---|
    | 手冊 p.85 | 10 BC |
    | remake 建築表(原版執行檔 `off_17EB3D + 12`) | **10** |

    而且 10 在整張建築表裡是**最貴的一棟**(第二貴是 5)。一棟能無視農業產出把饑荒填平
    的建築,代價就是每回合固定燒錢。測試連「它是全表最貴」都釘住——被改小就失衡了。

    ### 接在哪一行

    `RunColonyTurn` 裡,**污染扣完之後、人口成長之前**:換算要用的是可用產能(所以在污染之後),
    而成長同時吃「食物盈餘」與「淨產能」兩個數字(所以要在成長之前把兩個都改好)。
    BC 成本走 `EmpireOutput.FoodReplicatorCost`,和其他維護費一起進 `NetBC`。

    ### 誠實留白

    手冊沒說「國庫不夠付 1 BC/食物 時會怎樣」。**不編規則**:換算照做、成本照報,
    國庫可以被壓成負數(那是既有行為)。硬加一條「付不起就不換」會是憑空發明,
    而且會讓饑荒在破產時突然惡化——那種二階效應正是杜撰規則最容易搞砸的地方。

    ### 順帶清掉一段過期斷言

    `session.go` 有一段寫著「其餘 20 項(…再生反應爐、食物複製機、自動實驗室、
    行星輻射/通量/屏障護盾…)手冊效果不對應既有欄位,暫不建模」。
    第 92/93 項把那份清單裡的**六棟**做掉了,那段話已經與程式碼衝突,已改寫成
    「仍未建模的是真的缺子系統」並逐項列出缺哪個系統。

    ### 這一輪之後,建築表的狀態

    41 棟裡名字從未被程式碼消費的剩 **7 棟**:阿提米絲系統網(水雷子系統)、
    太空學院(艦員經驗值)、異族管理中心(同化)、戰機基地(格子戰鬥的獨立戰機單位)、
    恆星轉換器,以及先前已知的兩棟。**都是真的缺子系統。**

94. **阿提米絲系統網:水雷子系統**(2026-08-07)。

    上一項把它列在「真的缺子系統」那一欄。手冊 p.86 其實把整個子系統寫完了:

    > Any enemy ship entering that system has a chance to set off mines based on its
    > **size class**: Frigate = 20%, Destroyer = 30%, Cruiser = 40%, Battleship = 50%,
    > Titan = 80%, and Doom Star = 100%. Any affected ship sustains damage from
    > **8-28 mines**. Each mine inflicts **20 damage minus the ship's shield class**.

    三件事各自獨立而且相乘:

    | | 規則 |
    |---|---|
    | ① 觸發 | 逐艦擲,依艦體等級 20 / 30 / 40 / 50 / 80 / 100 % |
    | ② 水雷數 | 中招的船各擲一次,8–28 枚 |
    | ③ 每枚傷害 | 20 − 護盾等級 |

    ### remake 剛好已經有那兩個輸入

    - **艦體等級**:`shipStrength` 的六個類別(巡防/驅逐/巡洋/戰艦/泰坦/末日之星)
      正好是手冊那六個 size class。
    - **護盾等級**:remake 的護盾元件叫「第一級護盾」「第三級護盾」…「第十級護盾」
      ——那個「級」就是手冊說的 shield class,不必另建對照表。

    所以這棟建築缺的其實不是子系統,是**沒人把手冊那段翻成程式碼**。

    ### 為什麼大船反而危險

    機率隨體積上升(20% → 100%),而傷害不隨體積下降。水雷網的效果是**專打主力艦**:
    一群巡防艦大多開得過去,一艘末日之星必中。這與原版把它放在 Planetoid Construction
    (很後期)是一致的——它是拿來擋大艦隊的。測試釘住「機率單調上升」這個性質。

    ### 護盾在這裡的價值被放大

    三件事相乘,所以第十級護盾把每枚從 20 壓到 10,對一艘挨滿 28 枚的船就是**少 280 點**。
    測試用兩艘同型船跑同一組亂數流,驗「盾船的傷害總量剛好是裸船的一半」。

    ### 接在「進入」那一刻

    手冊寫的是「any enemy ship **entering** that system」——不是停留每回合。
    所以掛在 `advanceFleet` 的抵達那一段,順序放在**探索標記之後、一次性發現之前**:
    先挨雷再看見星系裡有什麼(雷區是進門就炸,發現是進去之後才看到的)。

    亂數用回合 + 星系當種子,與軌道轟炸同款——重播得到同樣結果,有測試釘住
    (網路對戰的決定性要求)。

    ### 誠實留白

    - **只對玩家艦隊生效。** AI 沒有「艦隊移動到某顆星」的模型(它的攻擊是抽象解算的),
      所以玩家自己蓋的水雷網目前擋不到 AI。**這是 AI 模型的缺口,不是這個系統沒接**
      ——AI 有真的艦隊移動時,`applyArtemisMines` 是它唯一的掛勾點。
    - 手冊沒說水雷網會不會被消耗、會不會對同一支艦隊重複觸發。**照字面做**:
      每次進入該星系逐艦擲一次,不消耗。
    - remake 的「偵察艦」「殖民船」不是原版的艦體等級——原版的偵察艦/殖民船/前哨艦/
      運輸艦都是蓋在 **Frigate 艦體**上的設計,所以套 20%。這是照原版的艦體分類,
      不是替沒有數字的東西編一個。
    - 被水雷炸沉的船直接離開艦隊,不進戰鬥記錄:remake 沒有「戰場外損失」這個事件類別,
      硬塞進戰鬥記錄會讓歷史紀錄出現一場沒有敵人的仗。

    ### 建築表狀態

95. **艦員經驗系統:三張加成表早就在,缺的是「經驗怎麼來」**(2026-08-07)。

    上一項把太空學院列在「缺艦員經驗值子系統」。盤點之後發現 remake 的狀況很特別:
    **加成表已經有三張,而且都對得上手冊**——

    | 表 | 位置 | 手冊欄 |
    |---|---|---|
    | `shipCrewOffenseBonuses = {0,15,30,50,75}` | `formulas.go` | BA |
    | `shipCrewDefenseBonuses = {0,15,30,50,75}` | `formulas.go` | BD |
    | `MissileCrew* = 0,7,15,25,37` | `missile.go` | ME |

    ——但**沒有任何一艘船有等級**。`shell.Ship` 沒有那個欄位,也沒有東西會讓它上升。
    三張表唯一的呼叫端在「讀存檔的船」那條路徑上(`engine.ShipBeamAttackFromDesign`),
    remake 自己造的船永遠是新兵。**表抄了、機制沒接。**

    ### 手冊 p.121 那張表補完

    第四條軌(Bo,登艦戰)先前沒人抄——因為 remake 還沒有登艦戰。這次補進來
    (`{0,5,10,15,20}`),即使暫時沒有呼叫端:**缺一條軌會讓下次有人接登艦戰時
    以為手冊沒給數字。**

    ### 統帥種族不是「升級快」,是整條階梯平移

    手冊那個星號:「This level is only attainable by crews in the service of a
    **Warlord** race.」配上前面那句「All ship crews start out as green rookies
    **unless** they're in service with a race that has the Warlord characteristic」
    ——兩句話講的是同一件事:

    ```
    一般種族   Green(0) → Regular(50) → Veteran(150) → Elite(500) → ✗
    統帥種族   ✗        → Regular(0)  → Veteran(50)  → Elite(150) → Ultra-Elite(500)
    ```

    所以兩張門檻表各有一個 **−1**(這個種族到不了這一級),而不是「一個很大的數」
    ——「到不了」與「很難到」在規則上是兩件事。

    ### 等級不存,只存經驗

    `Ship` 只加了 `CrewXP` 一個欄位,等級一律由它現算。存兩個欄位遲早會不同步
    (升級時忘了更新其中一個),存一個不會。太空學院的「起始等級 +1」因此也用經驗表達:
    起始 `CrewXP` 設成那一級的門檻,而不是另開一個「起始等級」欄位。

    ### 戰鬥經驗的三個限定詞

    > equal to the **halved sum of size classes (1-6)** of **destroyed (not captured)**
    > enemy ships (rounded down with a **minimum of 1**)

    - **halved**:打小船升得慢(擊沉一艘巡防艦 1/2 = 0 → 保底 1)
    - **destroyed not captured**:俘虜不算
    - **minimum 1**:但**一艘都沒沉時是 0 而不是 1**——那個保底講的是「有擊沉」的情況,
      把「贏了但一艘都沒沉」也給 1 是把保底條款擴大解釋

    還原「被擊沉的是哪些」用的是多重集合相減:`battleVolley` 就地移除陣亡者,
    呼叫端拿不到「誰死了」,但敵艦的 `atk` 就是戰力值、戰鬥中不變,
    所以「開打前的清單 − 結束時的倖存者」就還原得出來——**不必為此改戰鬥迴圈的介面**。

    remake 的 `shipStrength` 是 2 的冪(巡防 2、驅逐 4、巡洋 8、戰艦 16、泰坦 32、末日 64),
    正好對上手冊的 size class 1–6。

    ### 順帶收掉五個各自硬寫的 `false`

    `gamedata.Ground*BarracksCap` 早就有 `warlord` 參數(統帥種族營房容量加倍),
    但 shell 有**五處**各自硬寫 `false`。這次加的 `GameSession.RaceWarlord` 把它們統一
    ——特質系統補上時只要改一個地方,而不是五個而且很容易漏掉一個。
    (⚠ 目前沒有任何內建種族會設它,行為與先前完全相同。)

    ### 誠實留白

    - **只有玩家的船有艦員經驗。** AI 的艦隊在 remake 裡是每回合現生的戰力值
      (`genEnemyFleet`),沒有持久的船,自然沒有可累積經驗的對象。
    - 登艦戰加成有表沒有呼叫端(remake 還沒有登艦戰)。
    - 手冊說經驗來自「turn **in space**」,而 remake 沒有「在港內 vs 在太空」的區別
      ——所有船都在艦隊裡。目前的實作在 remake 的模型下等價,
      但之後若加了船塢/駐港狀態,這裡要跟著改。
    - 艦艇設計畫面直接造的船吃不到太空學院加成:那條路徑沒有「在哪造」的概念。
      逐殖民地造艦那條路(`deliverNewShip`)是正確的。

    ### 建築表狀態

96. **征服人口的同化系統**(2026-08-07)。

    異族管理中心先前被列在「缺同化子系統」。手冊把整張表逐個政體寫死了:

    | 政體 | 同化一單位人口 |
    |---|---|
    | 封建 Feudal | 8 |
    | 邦聯 Confederation | 4 |
    | 獨裁 Dictatorship | 8 |
    | 帝國 Imperium | 4 |
    | 民主 Democracy | 4 |
    | 聯邦 Federation | **2** |
    | 統一 Unification | **20** |
    | 銀河統一 Galactic Unification | 15 |

    ### 這不是細節,是「征服流 vs 和平流」的規則分野

    民主 4 回合 vs 統一 20 回合——**差五倍**。一個統一政體的帝國打下一顆 8 人口的星,
    要 160 回合才全部同化完;民主只要 32。統一政體的產出加成很高,而這條就是它的代價。

    異族管理中心那句「assimilates conquered populations at the rate of **1 per 2 turns**,
    **regardless of government**」直接蓋掉政體那一格——**對統一政體等於十倍速**。
    一棟維護費 1 BC 的建築有這種效果,是因為統一政體本來就設計成「你不該靠征服玩」。
    測試把那個「十倍」釘住。

    ### 兩個修正項,一個有數字一個沒有

    - **排斥 Repulsive**:「assimilate … at only **half** the normal rate」→ 回合數 ×2。
      而且手冊說這個修正套在「**this base rate**」上,所以連異族管理中心的固定值也吃它。
    - **魅力 Charismatic**:手冊只說「assimilate conquered colonists **easily**」,
      **沒有數字**;patch 1.5 的手冊也沒補。所以它**現在沒有任何效果**——
      `AssimilationTurns` 收了那個參數但不用它。

      這裡刻意寫了一支測試 `TestCharismaticHasNoQuantifiedEffectYet`:
      **把「刻意不做」與「忘了做」分開**。哪天有人塞一個猜的倍率進去,那支測試會紅,
      並在錯誤訊息裡要求連同誠實留白一起更新。

    ### 進階政體是研究出來的,不是選單選的

    四個進階政體(邦聯/帝國/聯邦/銀河統一)在原版是基礎政體的升級版,對應
    `TOPIC_ADVANCED_GOVERNMENTS` 這個四選一主題底下的四個科技。remake 的政府選單只有
    四個基礎型,所以 `assimilationGovernment()` 是「目前政體 + 有沒有那個科技」,
    判定沿用 `groundEquipTechOwned` 那組既有規則,不另立一套。

    ### 累積進度不歸零

    同化是「累積滿 N 回合同化一單位,餘數留著」而不是「每 N 回合歸零重來」——
    後者會在政體改變或蓋起異族管理中心時吃掉玩家已經累積的進度。
    測試釘住這一點:統一政體累了 3 回合之後蓋起管理中心(門檻 20 → 2),
    那 3 回合立刻兌現。

    ### 順帶抓到一個假斷言

    `session.go` 寫著「異族管理中心:士氣計算路徑已預留(**colonyMoralePercent 讀取此建築名**)」。
    **`colonyMoralePercent` 根本沒有讀它**——那個建築名在整個 repo 裡只出現在資料表與註解裡。
    這是「註解宣稱的事程式碼沒做」的典型:寫的時候可能真的打算接,但沒接成,
    註解留下來就變成假的可信度。已改寫成實際狀況。

    ### 誠實留白

    - **未同化人口目前沒有負面效果。** 手冊說多種族殖民地有 20% 士氣懲罰(建築可消除)、
      未同化人口增加叛亂機率(建築減半)。remake 的士氣路徑還沒接前者,叛亂系統不存在。
      所以現在同化只是一個會走完的計時器——**機制在、後果還沒接**,寫明白不假裝完整。
    - AI 攻下玩家殖民地那條路徑沒有同化(AI 沒有殖民地狀態的完整模型)。

    ### 建築表狀態

97. **戰機基地 + 行星版恆星轉換器,以及一個盤點方法的錯誤**(2026-08-07)。

    ### 兩棟建築,同一個接點

    remake 早就有「殖民地被軌道轟炸時反擊」那條路徑(`retaliationAttackers`),
    而且已經接了星基 / 戰鬥站 / 星辰要塞 / 飛彈基地 / 地面砲台。這兩棟缺的不是子系統。

    **戰機基地**(手冊 p.79):`10 Interceptor squadrons, 6 Bomber, or 4 Heavy Fighter,
    depending on the most advanced fighter technology`。中隊數**隨科技遞減**(10 → 6 → 4)
    ——每一階戰機更強,所以照抄這三個數,不要「順手」讓它遞增。測試釘住這個方向。

    **行星版恆星轉換器**(手冊 p.111):`400 points of damage to **each side** of a target
    — **1,600 total** — regardless of range and defense`。

    ### 第 91 項的一手科技表當場抓到一個錯

    寫戰機基地的分檔判定時,我把兩個科技都寫成 `TOPIC_ADVANCED_ROBOTICS`。
    拿第 91 項挖出來的 `OrigTechTopic` 查一次:

    ```
    TECH_BOMBER_BAYS        (31) → 主題 11 = TOPIC_ADVANCED_ROBOTICS          ✓
    TECH_HEAVY_FIGHTER_BAYS (83) → 主題 42 = TOPIC_SUPERSCALAR_CONSTRUCTION   ✗ 我寫錯
    ```

    重戰機那一檔會**永遠進不去**。這個錯不會讓任何測試變紅——它只是讓一條路徑靜靜地死掉。
    **第 91 項那張表在這裡第一次派上用場,而且是抓錯而不是查資料。**

    ### 差一點就送出去的雙重計算

    加完恆星轉換器之後,既有測試 `TestStellarConverterAddsColonyDefense` 紅了:
    「防禦 +1200,want +800」。查下去發現**它早就接過了**——在 `colonyDefense`
    (ai_attack.go)裡,用的是常數 `gamedata.StellarConverterName` 而不是字面字串。

    順著這條線又發現兩件事:

    1. **同一棟建築在兩條路徑上行為不一致**:它擋得住 AI 來襲(`colonyDefense` 有算),
       卻對軌道轟炸完全不反擊(`retaliationAttackers` 沒有它)。已統一到
       `retaliationAttackers`,`colonyDefense` 那邊的獨立加總移除。
    2. **`StellarConverterDefense = 800` 的來歷自己就矛盾**:註解寫著
       「手冊 p.106『400 傷 ×2(**雙側共 1600**)』」——但 400 × 2 = 800,不是 1600。
       回查手冊原文是「每一**面** 400、四面合計 1600」,1600 / 400 = **4 面**,不是兩側。
       已改成 `StellarConverterDamagePerSide`(400):remake 的防禦解算是抽象的、
       一次反擊就是一發打在一個目標上,沒有「同時打四面」;用 1600 會讓這棟建築價值變四倍。

    ### 盤點方法本身有洞

    第 92–96 項一路報的「還剩 N 棟未被消費」,用的是「建築名的**字面字串**有沒有出現在
    `buildings.go` 以外的程式碼」。那個掃法有一個盲點:**在 `buildings.go` 內部宣告成常數、
    再由別處引用**的建築會被漏判——恆星轉換器正是唯一一棟。

    掃法補上「常數引用」與「排除註解行」之後重跑:

    ```
    建築表 41 棟 → 字面字串消費 40 棟 / 經常數消費 1 棟 / 完全未消費 0 棟
    ```

    所以第 92 項那份「11 棟」的清單裡,**恆星轉換器是誤判**,真正的缺口是 10 棟,
    而第 92–97 項把那 10 棟全部接完了。

    ⚠ **「有程式碼消費」不等於「完整還原」。** 好幾棟仍有寫明的部分實作
    (護盾的氣候轉換、戰機基地的 10 回合整補、異族管理中心的士氣與叛亂條款…),
    那些留白各自記在對應的檔頭與 HONEST-STATUS,不因為這個計數歸零而消失。

    ### 一個如實記錄、沒有硬調的數字

    用 remake 現行的單隊近似值算出來的戰機基地戰力是
    **攔截機 480 / 轟炸機 120 / 重戰機 256**——**研究出轟炸機艙反而讓基地變弱**。

    那不是手冊的意思(手冊只給中隊數,沒說哪一檔比較強),是 `combat.go` 裡兩個標明過的
    近似值造成的假象:`fighterBeamDamageApprox=3` 配攔截機每架 4 次射擊 = 12,
    `fighterBombDamageApprox=5` 配轟炸機每架 1 次投彈 = 5。

    **沒有在這裡硬調數字去湊一條好看的曲線**——要修的是那兩個近似值,而那需要手冊或
    反組譯給出戰機的真實傷害,目前兩邊都沒有。已在接線處與測試裡各留一段說明。

98. **把上一輪自己寫的兩條留白關掉**(2026-08-07)。

    第 97 項的結尾特別強調「**有程式碼消費不等於完整還原**」,並列出仍有寫明的部分實作。
    其中兩條在寫的當下就已經解得開了——擋住它們的東西是我自己前幾輪剛加上去的。

    ### ① 多種族殖民地的 20% 士氣懲罰

    `gamedata.MoraleMultiRacialPenalty(hasAlienManagementCenter)` **早就存在而且是死碼**。
    它一直沒有呼叫端,理由寫在 `session.go` 的註解裡:「因 remake 無多種族人口追蹤,
    目前一律不套用」。

    第 96 項加上 `ColonyState.UnassimilatedPop` 之後,那個理由就不成立了:
    **未同化人口 > 0 就是多種族殖民地**。接上去之後三件事同時成立:

    - 攻下來的殖民地真的有代價(第 96 項寫的「機制在、後果還沒接」關掉了)
    - 異族管理中心的**第二條**手冊效果生效(第一條是同化速率)
    - 同化完最後一單位的那一刻懲罰消失——`advanceAssimilation` 每輪重算士氣,
      不然玩家會一直被扣到下次蓋建築為止(有測試釘住)

    ### ② 三面護盾的「Radiated 轉 Barren」

    三段手冊敘述用詞略異但意思相同:

    ```
    Radiation Shield 「Radiated worlds become Barren as long as the shield remains in place」
    Flux Shield      「The existence of a flux shield converts Radiated climates to Barren」
    Barrier Shield   「This shield converts Radiated climates into Barren」
    ```

    `ColonyState.Climate` 早就在(地形改造那一輪加的),所以只要在建成時走既有的
    `applyClimateChange`——那一支會連帶調整食物與人口上限,直接改 Climate 欄位不會。

    **⚠ 一個刻意的偏離,寫明白**:輻射護盾那句是「**as long as the shield remains in
    place**」——維持中而不是一次性。remake 接成一次性的,所以護盾被軌道轟炸摧毀之後
    那顆星**不會**變回 Radiated。理由是 remake 的建築效果**沒有一個是可逆的**
    (自動工廠被炸掉產能也不會退回去),為這一棟另建「效果可撤銷」的機制,
    代價遠大於它修正的失真。**這是選擇,不是疏忽。**

    ### 還沒接的那一條

    屏障護盾的「biological weapons cannot enter the planet's atmosphere」——
    remake 沒有「生物武器」這個分類,**沒接**。

    ### 這一輪的形狀

    沒有挖新的一手資料,做的是**回頭把自己標的留白逐條檢查一次**:
    哪些是「真的缺前置」、哪些只是「當時缺、現在有了」。兩條屬於後者。
    留白清單如果只增不減,它就會從「誠實記錄」退化成「免責聲明」。

99. **先進級開局的 19 個隨機主題:照抄結構,不照抄權重**(2026-08-07)。

    這個缺口從第 80 項就開著,第 88、91 兩次判斷「不照抄」,理由都是同一個。
    這次換個問法:**擋住的到底是哪一部分?**

    ### 讀得出來的與讀不出來的

    `Init_Player_Tech_` 的主迴圈跑 1 / 6 / 25 次:前 6 次取固定表 `word_18111C`,
    **第 7 次起改由 `sub_FD335` 隨機挑**。那支函式的評分是

    ```
    score = weight × horizon ÷ turns
    ```

    | 部分 | 出處 | 狀態 |
    |---|---|---|
    | `turns` | 主題成本 ÷ 每回合研究點,0 → 1 | **讀得出來** |
    | `horizon` | 15 起跳,不夠就 ×3÷2 再來一輪 | **讀得出來** |
    | 候選集合 | 只取狀態 2(現在可研究)的主題 | **讀得出來** |
    | 加權隨機 | `sub_FE96F` 走前綴和 | **讀得出來** |
    | `weight` | `sub_FC845` —— **985 行** | **不照抄** |

    **擋住的只有 `weight` 一項。** 把它一律當 1 之後 `score = horizon ÷ turns`
    ——選擇仍然由成本主導,只是失去「這個科技對這個種族/性格有多有用」那一層。

    ### 為什麼這比「只發六個」好

    先進級在原版是**開局就有 25 個主題**。remake 先前發 6 個,少了 19 個——
    那不是精度問題,是**少了一個等級的內容**。現在發滿 25 個,而且:

    - **只從「現在可研究」的主題挑**,所以是**沿著樹往上走**而不是從同一池子抽 19 次
      (測試逐領域檢查「已完成的主題前面不能有沒完成的」)
    - **偏好便宜的**,與原版的評分方向一致
    - **決定性**:同一顆種子重開得到同樣的開局;玩家與每個 AI 各走一條獨立的流

    第 91 項挖出來的 `OrigTopicCost`(83 個主題的一手成本)在這裡是燃料——
    沒有那張表,「偏好便宜的」就只能用 openorion2 的二手值。

    ### 誠實留白

    - **`weight` 一律 1**,少的是那 985 行的估值。AI 因此不會偏好「對它有用」的科技。
    - `sub_FD335` 尾巴還有一段依 `[player+0x28]`(值 1/2/4/5)的二次過濾。
      那個欄位**沒查到寫入端、沒有名字**——第 90 項在 `Send_Chat_Msg_` 裡也遇到同一個。
      **不照抄。**
    - 「每回合研究點」在開局那一刻其實還沒意義(還沒有回合跑過),
      呼叫端傳母星的初始研究產出,那是最接近原版當下狀態的值。

    ### 寫測試時抓到的一件事

    第一版測試斷言「先進級 = 25 個主題」,實際拿到 26。差的正好是
    `TOPIC_STARTING_TECH`——第 91 項認出來的那個「自環、永遠不會被解鎖」的容器主題,
    母星一律有。**不是程式錯,是斷言少算了一個已知的東西**;測試改成扣掉它。

100. **上游補完之後,下游要跟著讀真的東西**(2026-08-07)。

    第 99 項把 19 個隨機主題發出去了,但先進級的母星建築**還是只有兩棟**。
    原因很直接:`homeworldBuildingsFor(techLevel, pop)` 從**固定表**現算科技集合,
    **看不到那 19 個**。

    這正是第 81 項寫的那條依賴鏈——「開局建築取決於開局知道哪些科技」——只是方向反過來:
    上游補齊之後,下游如果還在自己算一份,補齊就傳不下去。

    改法很小:多一支 `homeworldBuildingsForKnown(techLevel, pop, known)` 吃**真正的**
    `CompletedTopics`,`applyStartingBuildings` 改走它。

    ### 結果自己對上了

    | TECH LEVEL | 開局主題 | 母星建築 |
    |---|---|---|
    | 曲速前 | 1 | **2**(海軍陸戰隊營 + 星基) |
    | 一般 | 6 | **2**(同上) |
    | 先進 | 25 | **6**(戰鬥站、星基、水耕農場、海軍陸戰隊營、生態圈、自動工廠) |

    前兩列與手冊逐字相符:「Pre-warp and Average Tech games only start with **Marine
    Barracks and a Star Base** because no other techs are Known that are also in the
    default initial buildings list」。

    第三列的 6 是名額數(上限 9、⌈⅔×8⌉ = 6)——**而那正是第 81 項寫測試時留的正對照
    預測的數字**:那支測試當年寫著「科技全解時這套機制**確實**會發滿 6 個名額」。
    缺口補上之後兩邊自己對上,**不是把斷言改成事後諸葛**。

    ### 順帶:把過期斷言清掉

    第 99 項做完之後,四處文件與註解仍寫著「19 個隨機主題還沒接」:

    - `gamedata/starting_tech.go` 兩處(`StartingTopics` 的 ⚠、`StartingTopicRandomExtras` 的來歷)
    - gap report 第 80 項的誠實留白、第 81 項的「缺口被釘在上一層」、第 91 項的結尾
    - `HONEST-STATUS.md` 的「開局建築的先進級也卡在這裡」

    全部改成現況。`TestAdvancedStartIsBlockedByTheMissingRandomTopics` 也改名為
    `TestAdvancedStartFillsAllBuildingSlots` 並反轉斷言——**它自己當年就寫著
    「那 19 個隨機主題若接上了,這條測試要跟著改」**。

    這一輪的形狀與第 98 項同款:**做完一件事之後,回頭找它讓哪些話變成假的**。
    留白與缺口記錄的價值在於它反映現況;一旦落後,它就從導航變成誤導。

101. **領袖技能:修掉一條疊加規則,補上四個技能,並確認一個「不是缺口」**(2026-08-07)。

    HONEST-STATUS 一直寫著「MOO2 25 個領袖技能,remake 接了 2 個」。這一輪逐條查手冊。

    ### 手冊給了一條 remake 一直做錯的規則

    p.137「Applicability」:

    > The effects of the **Megawealth and Researcher** abilities are **cumulative**,
    > but **the rest are not**. … the fleet gets the effect for that particular ability
    > of **the leader with the best applicable bonus**.

    `applyLeaderColonyBonuses` 先前是無條件 `+=`——**兩個貿易家就加兩份**,而原版只算最強的
    那一位。已改成先依技能分組、再依「累加 vs 取最佳」合成(`gamedata.LeaderSkillCombine`)。

    測試同時釘住兩邊:兩個貿易家**不**疊、兩個科學家**要**疊。少了後面那條正對照,
    「一律取最佳」也會讓前面那條通過。

    負加成另有一個坑:環保官是 −10%,取「數值最大」會挑到**最弱**的那位,所以合成是取
    **絕對值最大**。

    ### 單位是查出來的,不是猜的

    加成值在 `baseSkillValues[2]`(gamestate.cpp),**單位**在 `skillFormatStrings[2]`
    (officer.cpp)——兩張表在 openorion2 是分開的,只看數值不知道 10 是「10 點」還是「10%」。

    | 技能 | base | 格式 | 單位 |
    |---|---|---|---|
    | 環保官 | −10 | `%+d%%` | 百分比 |
    | 農業官 | +10 | `%+d%%` | 百分比 |
    | 財務官 | +10 | `%+d%%` | 百分比 |
    | **教官** | **+1** | **`%+d`** | **固定點數** |
    | 勞工官 | +10 | `%+d%%` | 百分比 |
    | 醫官 | +10 | `%+d%%` | 百分比 |
    | 科學官 | +10 | `%+d%%` | 百分比 |
    | 心靈導師 | +5 | `%+d%%` | 百分比 |
    | 戰術官 | +2 | `%+d` | 固定點數 |

    教官那一格是固定點數而非百分比,**正好對上手冊那句「Boosts the number of experience
    points earned each turn」**——是每回合多幾點,不是多幾成。兩個獨立來源指到同一個語意。

    ### 接了四個(標準仍是「有現成的承接欄位」)

    | 技能 | 承接欄位 |
    |---|---|
    | 財務官 +10% | `ColonyState.IncomeBonusPercent` |
    | 心靈導師 +5% | `ColonyState.MoralePercent` |
    | 醫官 +10% | `ColonyState.GrowthBonusSum` |
    | **教官 +1** | **艦員每回合經驗**(第 95 項才有的東西) |

    教官這一個是**第 95 項的下游**:艦員經驗系統做出來之前,那個技能沒有地方可接。

    ### 沒接的與理由

    環保官(降低「會產生污染的產能」的百分比——remake 的污染模型是 eighths 查表,
    沒有百分比入口)、農業官/勞工官/科學官(食物/工業/研究的**分項百分比**——
    `ColonyState` 只有 per-worker 與固定值兩種欄位)。

    ### 一個「不是缺口」的發現

    手冊在 Tactics 那一條的**最後一句**寫著:

    > Improves the coordination of the military forces in the system, adding to the
    > Beam Attack and the strength of ground troops.  **This skill is not implemented.**

    **原版自己就沒做。** remake 不做它與原版一致——把它列進「還沒接的技能」是分類錯誤。
    這句話值得記下來,否則下一個盤點的人會花時間去找它該有什麼效果。

    ### 順帶更正一個我自己的誤判

    查這一輪時我一度說「`loadHerodataMercs` 沒有呼叫端、真英雄池從沒裝進遊戲」——**錯的**,
    它在 `interactive.go:4384` 有呼叫,是我的 grep 被 `head` 截掉了。
    真英雄池早就接上;真正的缺口是技能對照表只認兩個,而那正是這一項修的。

102. **分項百分比:三個 admin 技能缺的只是一個欄位**(2026-08-07)。

    上一項把農業官 / 勞工官 / 科學官擋在門外,理由是「食物/工業/研究的**分項百分比**——
    `ColonyState` 只有 per-worker 與固定值兩種欄位」。回頭看那個理由:**缺的只是三個欄位**,
    而引擎早就有百分比進得去的地方。

    ### 引擎本來就有一條百分比路徑

    ```go
    pct := cs.MoralePercent + colonyGravityPenaltyPercent(cs)
    food     = GravityAdjustedProduction(Farmers*FoodPerFarmer, pct) + FlatFood
    gross    = GravityAdjustedProduction(Workers*IndustryPerWorker, pct) + FlatIndustry
    research = GravityAdjustedProduction(Scientists*ResearchPerScientist, pct) + FlatResearch
    ```

    士氣與重力就是走這條。加三個**分項**百分比只要在各自那一行多加自己那一項。

    ### 士氣是三項一起動,這三個不是

    這是它們與士氣的唯一差別,也是測試的主軸:農業官只動食物、勞工官只動工業、
    科學官只動研究,另外兩項必須**一動也不動**。

    正對照是「士氣**仍然**三項一起動」——少了它,「分項百分比其實沒接上」也會讓前面那支通過。

    第三支釘住「固定加成不吃百分比」(農夫為 0 時食物全來自 `FlatFood`,
    百分比不該放大它),與士氣/重力的既有處理一致。

    ### 科學官 ≠ 科學家

    兩個中文名很像,但在原版是不同的技能、不同的單位:

    | 技能 | id | 格式 | 落在 |
    |---|---|---|---|
    | 科學家 Researcher | 6(common) | `%+d` | `FlatResearch`(固定點數) |
    | 科學官 Science Leader | 38(admin) | `%+d%%` | `ResearchBonusPercent`(百分比) |

    而且**累加規則也不同**:科學家是手冊明列的兩個累加型之一,科學官不是(取最強那位)。
    有一支測試專門釘這件事——名字像就混用會同時錯兩處。

    ### 領袖技能的現況

    remake 現在接了 **9 個**:科學家、貿易家、財務官、心靈導師、醫官、教官、
    農業官、勞工官、科學官,加上地面戰的指揮官。

    仍未接的與理由:
    - **環保官**:降低「會產生污染的產能」的百分比——remake 的污染模型是 `PollutionEighths`
      查表(八分之幾),沒有百分比入口。要接得先把污染模型改成連續量。
    - **工程師**:艦艇維修**速率**加成——remake 的 `advanceShipRepair` 是照原版
      `Repair_Ship_Full_` 做的**一次修好**,沒有「速率」這個量可以加成。
      **這不是漏做,是那個量在這個模型裡不存在。**
      > ⚠ 2026-08-08(第 103 項)修正:**這一條只講對了一半。** 手冊那條有兩句,
      > 上面只看了第一句(戰鬥中的修復)。第二句「repairs all structural and internal
      > systems damage **after the battle is won**」對得上 remake 既有的
      > `repairAfterBattle`,已於第 103 項接上。
    - **戰術官**:原版自己就沒實作(第 101 項)。
    - 其餘 captain/common 技能(刺客、外交官、間諜大師、心靈感應…)對應的子系統
      (刺客擲骰、外交修正、反間諜)remake 沒有。






103. **HERODATA 的技能欄位是每技能 2 bit,不是 1 bit**(2026-08-08)。

    第 101、102 項把 admin 技能一個個接上去,但**真英雄一個都拿不到**——上游把技能位元讀錯了。

    ### 一手事實

    openorion2 `Leader::hasSkill`(gamestate.cpp:631-662):

    ```cpp
    skillnum = id & SKILLCODE_MASK;
    if (skillnum >= max) return 0;
    return (skills >> (2 * skillnum)) & 0x3;
    ```

    **每個技能佔 2 bit**,值就是技能階(0 無 / 1 一般 / 2 進階)。而 remake 的
    `cmd/moo2/herodatamercs.go` 讀的是:

    ```go
    skillCommonResearcherBit = 1 << 6 // SKILL_RESEARCHER
    skillCommonTraderBit     = 1 << 9 // SKILL_TRADER
    ```

    一個技能一個 bit。SKILL_RESEARCHER 真正的位置是 bit 12-13;bit 6 是
    **skillnum 3(SKILL_FAMOUS)** 的低位,bit 9 是 **skillnum 4(SKILL_MEGAWEALTH)** 的高位。
    **兩個標籤都貼錯人**,而且畫面上完全看不出來:名字是真的、等級是真的,只有技能是錯的。

    解碼函式 `gamedata.LeaderSkillTier` 早就照 `hasSkill` 寫對了——**只是這裡沒用它**。

    ### 順著這條線挖出的另外四個

    | # | 症狀 | 後果 |
    |---|---|---|
    | 1 | `Tier` 寫死 1 | 進階技能的 +50% 一次都沒發生過 |
    | 2 | 一位英雄只給一項技能 | 原版一人可有多項(所以才要 2 bit × N 欄位) |
    | 3 | 艦艇軍官通稱「指揮官」 | 那是 SKILL_COMMANDO 的譯名,`commandoLeaderTier` 掃字串 → **每一位**雇來的艦艇軍官都吃到地面戰加成 |
    | 4 | 效果查表用**中文標籤** | 英文模式下 `Skill` 是 "Scientist",查不到 → **所有領袖加成同時消失**,畫面無異狀 |

    第 4 條是這一輪最值得記的:**翻譯過的字串不能當識別鍵**。remake 有三處這樣寫
    (`leaderSkillIDByName`、`commandoLeaderTier`、`FleetHasNavigator`),三處都會在切成英文的
    那一刻靜默失效——沒有錯誤訊息,沒有崩潰,只是加成不見了。

    ### 修法

    - `gamedata/leader_skill_names.go`:27 個技能的 id ↔ 中英文名(英文名逐字取自
      GAME_MANUAL.pdf p.135-137 三段技能表)+ 原版的列舉順序(專屬技能在前,
      對照 `LeaderSkillsWidget::update`)。
    - `shell.Leader` 加 `Skills []LeaderSkill{ID, Tier}`,**id 是識別鍵、標籤只負責顯示**。
      沒有 `Skills` 的舊資料(demo 領袖、既有測試、舊存檔)退回用中文標籤反查,相容不變。
    - 三處標籤比對全部改成比對 id。
    - 艦艇軍官的通稱從「指揮官」改成「艦長」,並用測試釘住
      **通稱不可以撞到任何真技能的譯名**。

    ### 順帶接上的:工程師

    手冊 p.136 Engineer 那條有兩句,第 102 項只看了第一句就結案。第二句:

    > an Engineer that has not retreated from combat, repairs all structural and internal
    > systems damage **after the battle is won**.

    「戰後完全修復」正是 `shell.repairAfterBattle` 在做的事——它本來就有兩個觸發
    (自動修復元件、進階損害管制),加第三個只是多一個條件。`won` 這個新參數**只**影響
    工程師那一條:前兩條是裝備/科技的被動效果,手冊那兩句沒有勝負條件,不加。
    (有一支測試專門釘這件事——正對照是「打輸時自動修復元件仍然照修」。)

    誠實留白兩個:手冊的「has not retreated」在 remake 沒有東西可以判(戰鬥解算沒有撤退機制);
    軍官也沒有指派到艦隊,沿用 `commandoLeaderTier` 的既有近似(帝國內有就算數)。

    ### 領袖技能的現況

    接了 **10 個**:科學家、貿易家、財務官、心靈導師、醫官、教官、農業官、勞工官、科學官、
    工程師,加上地面戰的指揮官與航行的領航員 —— 共 12 項技能有實際效果。

    仍未接的與理由(與第 102 項相同,扣掉已接的工程師):環保官(污染模型是八分之幾的查表,
    沒有百分比入口)、戰術官(原版自己就沒實作)、其餘 captain/common 技能對應的子系統
    remake 沒有。
    > ⚠ 2026-08-08(第 104 項)更正:**環保官那個理由也是錯的**,見下一項。

104. **環保官:第三次用同一種方式擋錯**(2026-08-08)。

    第 101 項擋掉農業官/勞工官/科學官,理由是「沒有分項百分比欄位」——第 102 項發現缺的只是三個欄位。
    第 102 項擋掉工程師,理由是「維修速率那個量不存在」——第 103 項發現手冊那條的第二句對得上既有函式。
    這一輪輪到環保官,擋門理由是「污染模型是八分之幾的查表,沒有百分比入口」。

    **同一個毛病第三次**:把「這個量是查表算出來的」誤讀成「沒有地方可以再乘一個百分比」。

    ### 手冊逐字就是那個量的名字

    > Environmentalist: Reduces **the amount of production that causes pollution**
    > on the colonies in the system.

    「the amount of production that causes pollution」——`PollutionPollutingProduction`
    這支函式的名字本身就是那句話。入口一直都在,而且只有一個呼叫點。

    接的位置是**查表之後、扣容忍值之前**:

    ```
    eighths 查表 → 環保官百分比 → 減容忍值 → 一半 = 清理成本
    ```

    ⚠ 這**不是減產能**。接到 `gross` 那一行會變成「請一個環保官等於少一個工人」,
    那是手冊那句話的反面。有一支測試釘住「淨工業的增加額**恰好**等於清理成本的減少額」
    ——差額就代表動到了別的東西。

    ### 為什麼是相乘

    這條規則手冊沒有直接寫,但它給了同一類效果怎麼疊的**算術**(p.90 大氣更新器):

    > This effect is cumulative with that of the Pollution Processor; if both are in place,
    > **only one-eighth** of the industry produces pollution.

    污染處理器留 1/2、大氣更新器留 1/4,並存留 **1/8 = 1/2 × 1/4**。
    (手冊用的字是 "cumulative",但它自己算出來的數字是相乘——**數字贏字面**:
    相加得不到 1/8。)所以環保官的百分比當成同一條鏈上的另一個乘數,
    是手冊給過的唯一一條「污染削減怎麼疊」的規則,不是另創的模型。

    有一支測試專門把這條算術釘在 `PollutionEighths` 上:哪天有人把它改成相加,那裡會先響。

    ### 符號

    環保官的 `skillBonus` 是**負值**(base −10),而 `ColonyState` 那一欄存的是**正的減幅**
    ——存進去時 `-=`(負負得正),消費端讀起來就是「(100 − 減幅)」,公式裡不必處理負號。
    搞反會變成「請環保官讓污染變嚴重」,而且不會有任何既有測試報錯,所以另補了一支釘符號的。

    ### 領袖技能的現況

    **13 項有實際效果**。仍未接的只剩戰術官(原版自己就沒實作)與「子系統還沒有」的那幾項
    (刺客擲骰、外交修正、反間諜、艦艇軍官指派 UI)——**沒有一項是卡在「找不到入口」了**。

105. **叛亂:同化計時器的另一半**(2026-08-08)。

    第 96 項接了同化,而那個檔的檔頭自己寫著「叛亂系統根本不存在 …
    現在同化只是一個會走完的計時器——**機制在、後果還沒接**」。這一項接後果。

    ### 手冊給了規則,沒給數字

    > There is a chance each turn that the conquered aliens will attempt to revolt.
    > **The more unassimilated aliens, the larger the chance.** This chance is **doubled**
    > if the captured population is being **exterminated** and **halved** if there is an
    > **Alien Management Center** … A loss for you is a gain for the world's old ruler —
    > **the colony reverts back**.(p.165)

    「越多越可能」是什麼函數?沒寫。數字在 `Check_Rebellion_` @ 0xED260:

    ```
    cmp  byte ptr [colony+12Fh], 4   ; 4 = 未被征服(初始化時設的)→ 根本不檢定
    jz   結束
    ...                              ; ecx = 未同化且「原主帝國還在」的人口單位數
    imul edx, ecx, 0Ah               ; ★ 機率 = 單位數 × 10
        add edx, 難度×4 − 8          ; ★ 只在「主人是人類、叛軍是 AI」時
    ...  edx = edx / 2               ; ★ 異族管理中心
    ...  add edx, edx                ; ★ 滅絕政策
    mov  eax, 3E8h / call sub_1247A0 ; rand(1..1000)
    cmp  eax, edx / jg 結束          ; roll <= chance 才叛亂
    ```

    **每一單位未同化人口 1%**,順序是 基準 → 難度 → 減半 → 加倍。

    ### `[colony+0x137]` 是異族管理中心——三個呼叫端對上手冊的三句

    | 呼叫端 | 條件 | 手冊 |
    |---|---|---|
    | `Check_Rebellion_` | `!= 0` → 機率減半 | 「halving the chance of revolt」 |
    | `Apply_Assimilation_` | `!= 0` → 速率固定 120 | 「1 per 2 turns, **regardless of government**」 |
    | `sub_DDAD4` | `== 0` 才算多種族 | 「removes the 20% morale penalty from multi-racial colonies」 |

    而 `OrigBuildingID["Alien Management Center"] = 1`。三句對三個呼叫端,不是猜的。

    ### 順帶定名:`GroundTypeFourth` 就是叛軍

    那個常數從第 27 項起一直掛著「⚠ 未定名(−20,且基礎值取自另一方),
    **而殖民地防守方根本不填它**」。

    `Get_Rebellion_Info_` @ 0xEC65A 把守方三種部隊填進 `[+0x0A]`(裝甲)、`[+0x0C]`(陸戰隊)、
    `[+0x0E]`(民兵),而**叛軍的數量填進 `[+0x10]`**——同一陣列的第四格。**類型 3 = 叛軍。**

    這也回頭解釋了「守方根本不填它」那句:**叛軍永遠是攻方**。攻擊力 −20(四種裡最弱)
    也合理——起事的是沒受過訓練的被征服人口。常數已改名 `GroundTypeRebels`。

    ### 手冊沒提的第二個骰

    原版決定叛亂之後又擲了一次:`mov eax, ecx / call sub_1247A0` → 結果直接當叛軍部隊數。
    所以起事的是 **rand(1..未同化人口數)**,不是全部。

    ### 誠實留白

    - **鎮壓後死多少人**沒有依據。原版那段在 `Resolve_Rebellion_Troops_`,沒有逐指令抄;
      remake 採「起事者全滅」,**這是建模選擇不是逐字依據**,寫在函式註解裡。
    - **沒有滅絕政策**:remake 沒這個選項,「×2」那一路目前不會發生
      (函式仍收這個參數,等 UI 有那個選項直接接)。
    - `[colony+0x12F]` 的完整列舉沒查(已知 4 = 未征服、0 = 滅絕,1/2/3 語意未定)。
    - **只檢定玩家的殖民地**:AI 打下玩家殖民地那條路徑本來就沒有同化模型。

    ### IDA 這一輪是壞的

    `tools/ida.sh idc` 三種呼叫形式都回 `Failed to initialize IDA as library (error code 4)`。
    沒有繼續敲——改讀 `Orion2.exe.asm` 原始清單 + `symbols_fixed.tsv` 對名字,
    整條規則就是這樣抽出來的。**IDA 只是加速器,不是必要條件。**

106. **AI 艦隊會在星圖上移動了**(2026-08-08)。

    先前 AI 的突襲是**瞬移**的:`aiRaid` 直接拿 `aiRaidWilling` + `aiRaidTarget` 結算,
    AI 沒有位置,艦隊憑空出現在玩家殖民地上空。三個後果:

    1. **玩家看不到它來**——無法預警、無法攔截,只能事後看回合摘要。
    2. **阿提米絲系統網打不到 AI**。那棟建築的手冊效果是「任何**進入**該星系的敵艦」,
       而 AI 從來沒有「進入」這個動作。這條缺口從第 94 項一直記到現在。
    3. 星雲/黑洞/干擾場對 AI 完全沒有意義,因為它不移動。

    ### 改了什麼

    `AIOpponent` 加上位置與航線(`FleetStar` / `FleetPosSet` / `FleetDestStar` / `FleetETA`),
    `advanceAIFleets` 每回合推進,突襲的前提從「**想打**」改成「**打得到**」
    ——艦隊靜止且停在某個玩家殖民地的星上。

    **決策規則一條都沒改**:願打門檻(戰爭態勢/軍力領先/性格積極度)與目標估值
    (原版三層 `Colony_Worth_To_Player_` / `Enemy_Colony_Worth_To_Player_` / 距離)原封不動,
    只是搬到了**出發**那一刻,中間隔著一段航程。

    ### 這一輪自己引入又自己抓到的兩個 bug

    **① 停在原地的艦隊會每回合都打。** 願打門檻搬到出發時之後,結算端若不保留間隔守門,
    一支抵達後停著不動的艦隊會每一回合結算一次突襲。間隔守門因此**留在結算端**
    ——它問的是「多久能打一次」,與「要不要出兵」不是同一個問題。

    **② 四個守門測試會假綠。** `TestAIRaidGracePeriod` / `NeedsWarStance` /
    `NeedsStrengthAdvantage` / `PacifistNeverRaids` 原本呼叫 `advanceAIRaids`,
    而守門搬家之後它們會因為「艦隊本來就不在目標上」而通過——**測不到任何東西卻一路綠**。
    四支都改成驗出發那一端(`advanceAIFleets` 後有沒有 `FleetETA > 0`)。

    ### 存檔:靠截圖廊反證的一個漏接

    `aiSnapshot` 是**逐欄位手抄**的,加新欄位不會自動進存檔。第一次跑截圖廊時
    `30_netwait.png` 的狀態指紋**沒有變**——那正是「新欄位沒被序列化」的證據
    (指紋 = 存檔快照 JSON 的 SHA-256)。補進 `aiSnapshot` 與 `restore` 之後指紋才變。

    沒有這一步的話,一支飛了八回合快到玩家家門口的艦隊,**存一次檔就瞬移回母星**
    ——而且只有存讀檔的玩家會遇到。另補了一支存讀檔回歸測試。

    ### 誠實留白

    - **一個 AI 只有一支艦隊**:它的軍力是單一整數 `FleetStrength`,不是艦艇清單。
    - **水雷對 AI 是戰力折損,不是逐艦模擬**:原版逐艦擲觸發率(依艦體等級 20–100%)、
      逐艦扣血,AI 沒有艦艇資料。折算除數 10 是**建模選擇不是考據值**,寫在常數註解裡。
    - **艦員經驗仍拿不到**(同一個原因:沒有逐艦資料)。
    - **航線不判星雲/黑洞/干擾場**:玩家那條路徑模型綁在 `s.Player` 上。
    - **AI 之間不互相出兵**:需要 AI-vs-AI 戰鬥解算,remake 沒有。

107. **三個「待辦」複查:一個早就做完了、一個是證據不足、一個追到源頭**(2026-08-08)。

    WORKLIST 剩餘清冊裡的 F / H / I 三項,這一輪逐一複查而不是逐一開工。

    ### I(turn-1 校準)——**早就做完了**

    `docs/HONEST-STATUS.md` 寫著「turn-1 校準開放項:科學家分配(科3 vs 原版可能科4)待 playtest 定案」。
    grep 程式碼:

    ```go
    // internal/shell/session.go, playerHomeworldColony
    Farmers: 4, Workers: 2, Scientists: 2,
    ```

    與 2026-07-12 釘死的 oracle 完全一致——「科3」那個狀態在 **2026-08-06 的校正就不存在了**。
    同一份文件的第 14 行還寫著「農4/工1/科3」,也是同一批過期敘述。

    **這一條的存在本身是教訓:待辦事項會比它描述的問題活得更久。** 斷言「某項還沒做」之前先 grep
    (rule 63)——這一輪光是複查就結掉一項,比開工便宜得多。

    ### F(戰機基地 10 回合整補)——**證據不足,不是工時不足**

    手冊 p.86:「All ground-based squadrons of fighter craft are **totally renewed every 10 turns**.」

    要「整補」得先有東西**會少**。remake 的戰機基地戰力是當場算出來的:

    ```
    internal/shell/orbital_bombardment.go:218
        atk, _ := gamedata.FighterGarrisonCombatContribution(fighterGarrisonTierFor(defender))
    ```

    `fighterGarrisonTierFor` 只看科技,回傳 10/6/4 個中隊——**沒有任何地方存著「現在剩幾隊」**。
    補一個 10 回合計時器,它會把一個永遠等於滿額的值重設成滿額。

    而**手冊沒有描述任何耗損機制**。憑「整補」反推「一定會耗損」再自訂損失規則,那是臆造。
    往前推的下一步是去反組譯找中隊狀態欄位,不是先寫計時器。**擱置,理由寫清楚。**

    ### H(安塔蘭母星防禦艦隊)——**追到源頭了**

    先前寫「手冊/openorion2 均無精確數字,用保守預設 6 艘末日之星等級」。
    這一輪從符號表找到了整條鏈:

    `Load_Antaran_Defense_Fleet_` @ 0x4D141(只有 77 bytes,全文可讀):

    ```
    if word_199182 < 1: word_199182 = 1     ; 第 5 筆至少 1
    for bx = 0..4:                           ; ★ 5 種艦體尺寸
        while cx < word_19917A[bx*2]:        ; ★ 每種尺寸的數量
            Load_Combat_Antaran_Ship_(...)
    Load_Antaran_Star_Fortress_()            ; ★ 外加一座星際要塞
    ```

    兩個結構性事實立刻推翻了「6 艘同級」這個預設:**艦隊是五種尺寸的混編**,而且**還有一座星際要塞**。

    `word_19917A` 是 BSS(`dw ?`),數量執行期算;寫入端是
    `Build_Antaran_Defensive_Ships_` @ 0x63F9C(由 `Antaran_Invasion_Check_` @ 0x63D92 呼叫),
    它從靜態範本 `byte_181746` 以 `movsd/movsd/movsw`(每筆 10 bytes)複製。
    另有 `sub_646F9` 在艦艇損失時遞減、並從 `word_19918C[]` 補充。

    **範本表的數值與縮放規則已派工解讀中**,結果進 `docs/re/antaran-defense-fleet.md`。
    在那之前 remake 的 6 艘預設**不動**——結構知道了不等於數字知道了。

    ⚠ 順帶記一件事:`word_199182` 就是 `word_19917A[4]`(位址差 8 = 4 個 word)。
    開頭那個「夾成至少 1」夾的是**最大艦那一格**——防禦艦隊永遠至少有一艘最大的。
    這種「兩個看起來無關的符號其實是同一張表」的事,只看符號名看不出來,要算位址。

108. **兩個「無精確數字」的斷言同時被推翻**(2026-08-08)。

    這一輪派了兩支 sonnet 解資料表,兩份都抽驗過才收。

    ### 一、安塔蘭母星防禦艦隊:不是「6 艘同級」

    `internal/shell/antaran_victory.go` 原本寫著:

    > 誠實聲明(手冊/openorion2 皆無精確數字)… 保守預設:6 艘「末日之星」等級戰力

    **兩件事都錯了:數字有,而且組成根本不是同級的。**

    `Load_Antaran_Defense_Fleet_` @ 0x4D141 只有 77 bytes:五種艦體尺寸各取一張表的數量,
    **外加一座星際要塞**。上限表 `_n_max_antaran_def_ships`(`byte_181746`)逐位元組解出:

    | 索引 | 尺寸 | 上限 |
    |---|---|---|
    | 0/1 | Small(Raider)/ Medium(Marauder) | **0——永遠不造** |
    | 2 | Large(Intruder) | 3 |
    | 3 | Huge(Interdictor) | 2 |
    | 4 | Titan(Harbinger) | 7 |

    合計 **12 艘 + 1 座要塞**。難度**不改變組成**,只改變累積速度與逐艦裝甲加成。

    我自己核過原始位元組:`byte_181746 db 0` + `align 4` + `dd 70203h` → `{0,0,3,2,7,0,…}`,
    而符號名 `_n_max_antaran_def_ships` 直接證實語意。

    ⚠ 艦體尺寸→remake 六級戰力階梯的對照**是推論**(原版五級與 remake 六級沒有現成對照),
    取「相對順序不變、最大的對到最頂端」。**數量與分層是真值,那一層映射是選擇。**

    ### 二、開局隨機科技:錯的不是權重,是**粒度**

    `starting_random_tech.go` 的留白一直寫著「`weight` 一律 1」,把缺口框成「權重不準」。
    這一輪讀 `Choose_Tech_Application_` @ 0xFD335 的迴圈才發現框錯了:

    ```
    mov  ebx, 350h                   ; 0x350 = 848 = 212 × 4 —— 逐 tech-item 的分數陣列
    cmp  byte ptr [eax+117h], 1      ; [player + 0x117 + techIdx] 逐 tech-item 狀態
    cmp  byte ptr [ebx+eax+0C4h], 2  ; [player + 0xC4 + topic]   逐主題狀態
    movsx edx, di                    ; ★ 傳給估值函式的是 tech-item 索引
    ```

    兩個索引空間並存,而函式名直說了——Choose Tech **Application**。
    remake 發的是 `CompletedTopics[t] = true`(**整個主題**),而完成主題卻不做抉擇時
    `componentUnlockedFor` 會把底下的抉擇**全部**解鎖。

    **先進級開局在 remake 拿到的是原版的兩到三倍。** 這不是權重誤差,權重再怎麼校準也修不了。
    已補 `StartingRandomApplicationPick`:挑完主題再挑一項應用,設 `ChosenTech`/`ExplicitChoice`;
    `ResearchAll` 的主題(手冊明說三項全拿)維持全解鎖。

    ### 三張資料表:自我驗證抓到一個解碼方法論錯誤

    表③(212 筆 tech-item 記錄)的 topic 欄與第 91 項抽的 `OrigTechTopicTable` 逐筆比對,
    **211/212 吻合**。過程中先撞出 41 筆錯誤,原因是 IDA 的顯示慣例:
    **4-byte 欄位的數值只要巧合等於某個已知位址,就會顯示成 `offset labelname[+N]`**,
    即使那 4 bytes 語意上根本不是指標。改成還原真實位址之後才跳到 211/212。

    **唯一不吻合的 techIdx=29 沒有被改成「正確答案」。** 兩個獨立來源都指向 70,解碼出來是 889,
    問題侷限在那一筆的頭 4 bytes。照實留著——「調到吻合」是把答案倒著湊,那會讓解碼方式的
    可信度變成循環論證。有一支測試釘住這件事:哪天有人把它改成 212/212,那支會提醒他在對答案。

    另有一個獨立的邊界巧合佐證:表②的 category 42..48 全是 (0,0) 填充,
    而表③實測到的 category 最大值恰好是 41。**兩張表獨立解碼卻在同一邊界吻合。**

    ### 這一輪 agent 兩次推翻我的簡報

    - 我在派工簡報裡說原版座標系是 320×200,agent 拿三張畫面的立即數(全落在 0..639/0..479)
      推翻了它——而 remake 本來就是 640×480。
    - 我說 `movsd/movsd/movsw` 那段是在複製範本表,agent 查出那三行其實是把字串
      `"defensive\0"` 複製到除錯緩衝區,範本表是用暫存器傳的。

    **兩次都是我錯。** 派工要留把關,也要留反駁空間——照著錯的簡報做出來的東西,把關也抓不到。

109. **軍官畫面座標:openorion2 → 執行檔立即數**(2026-08-08)。

    第 41 項把估計熱區換成 openorion2 的真值座標,那是當時能拿到最好的來源。
    現在有更上游的:`Add_Officer_Screen_Fields_` @ 0x9264E 的立即數。

    依專案的來源優先序(**反組譯立即數 > 手冊 > openorion2 > LBX 尺寸 > 量圖**),
    以執行檔為準。差異:

    | 元素 | 先前(openorion2 / 量圖)| 執行檔立即數 |
    |---|---|---|
    | Colony Leaders 分頁 | x=20 | **x=9** |
    | Ship Leaders 分頁 | x=166 | **x=156** |
    | HIRE | (313, 440) | (313, **441**) |
    | POOL | (388, 440) | (388, **441**) |
    | DISMISS | (462, 440) | (**463**, **441**) |
    | RETURN | (540, 440) | (**538**, **441**) |
    | 清單列中心 | 90/199/308/417 | **88/197/306/415** |
    | 上下捲鈕 | **沒有** | (613, 22) / (613, 170) |

    我自己核了 HIRE 那一組(asm 199093:`mov eax, 139h` = 313、`mov edx, 1B9h` = 441)。

    ### 三件順帶修掉的事

    **① RETURN 先前自相矛盾**:熱區在 538(對的)、疊字標籤在 540(錯的)。同一顆按鈕兩個位置。

    **② 上下捲鈕從來沒接過。** 那兩顆的座標一直都在執行檔裡(`_officer_up_button_seg` /
    `_officer_dn_button_seg`),remake 沒做——所以**軍官清單超過四列就看不到後面的人**。
    現在接上了,並在最後一列下方標「還有 N 位」。

    **③ 座標抽成套件層級的資料**(`cmd/moo2/officerscreen.go`)。原本內嵌在 `officer()` 裡,
    而那支函式要真 LBX 資產才跑得起來——測試環境沒有,所以那些數字**測不到**。
    抽出來之後可以直接對數值寫回歸測試,而且「這些是查來的真值」在程式碼結構上看得見。

    ### 寬高一格都沒動

    執行檔那邊的寬高欄位是 LBX 資產控制碼、不是字面尺寸,所以**沒查到**。
    沒查到的就不動——不拿「看起來差不多」的數字去覆蓋既有值。

    ### 誠實留白:這個畫面沒有視覺回歸覆蓋

    截圖廊的 shot 是用**絕對 tick** 索引的,中間插一段導覽會把後面 24 張的 tick 全部位移。
    為了加一張圖去冒這個風險不划算,所以**軍官畫面目前不在截圖廊裡**
    ——座標由單元測試釘住,但「畫出來長什麼樣」沒有 byte-diff 保護。記在這裡。

    (順帶:`01b_newgame.png` 那張的命名慣例就是為了避開重編號,同一個理由。)

110. **間諜 UI:那張畫面不存在,所以規劃本身要改**(2026-08-08)。

    WORKLIST 的 D 項寫的是「spy / leader UI」,而 spy 那一半的原始假設是**做一張間諜畫面**。

    反組譯搜遍 `Spy_Screen` / `Espionage_Screen` 等關鍵字**零命中**。間諜的任務指派內嵌在
    `Race_Screen_` @ 0x10ACBA 這張**種族關係**畫面裡:每個已接觸種族一列,列上有關係滑桿,
    列旁邊有派間諜的按鈕。

    **照舊假設去做,方向從一開始就錯。** 這是負面結論的價值——它省掉的是一整張不存在的畫面。

    ### remake 那張畫面早就有,但座標是量圖的

    `sceneBuilder.races()` 存在,而它的註解自己寫著:

    > races 在 openorion2 是 STUB 無硬編座標,故用 PIL 量測的 overlay 位置當熱區來源

    現在有更上游的來源。`.data` 段的兩張靜態陣列(逐位元組解出):

    | 表 | 位址 | 內容 |
    |---|---|---|
    | `_race_spy_btns` | 0x18406F | 每個種族那排任務鈕的錨點 |
    | `_race_bar` | 0x18400D | 關係滑桿軌道的貼圖位置 |

    版面是 **左欄 4 列(x=120)+ 右欄 3 列(x=332),兩欄 y 完全對齊**
    ——那正是 MOO2 最多同時顯示 7 個對手的版面。remake 先前是自編的「y 從 66 起、每列 +62」。

    ### 三顆鈕只做了一顆,而那是刻意的

    原版每個種族有**三顆**任務鈕(x 偏移 0 / +76 / +149)。反組譯只看得出「同一列建三顆」,
    **沒有查到哪一顆對應哪個任務**——`Adjust_Spy_Mission_Data_` 只看到把任務欄位從 0 設成
    預設值 3,三顆各自的目標值沒追到。

    手冊的書寫順序是 Espionage→Sabotage→Hide,但**書寫順序不保證等於 UI 由左到右的順序**。

    所以只做最左邊那一顆,掛 remake 唯一有模型的任務(派間諜偷科技;破壞與隱匿在
    `internal/shell/spy.go` 就標著「手冊只有定性描述,沒給規則」)。

    **畫一顆意思說不出來的按鈕比不畫更糟**:玩家點下去不知道會發生什麼,而我也沒辦法
    在任何地方誠實描述它。另外兩顆的座標留在 `docs/re/screen-coords-spy-leader.md` 等順序查明。

    ### 寬高一樣沒查到

    執行檔那邊的寬高是 LBX 資產控制碼。已知的唯一約束是三顆鈕相鄰間距 76,所以單顆不會更寬。
    取 68×20 是 **remake 的選擇,不是真值**,寫在常數註解裡。

111. **屏障護盾擋生物武器:擋門是「分類的語意」,不是規則本身**(2026-08-08)。

    `internal/gamedata/planetary_shield.go` 的檔尾掛了一條沒接的規則:

    > 屏障護盾多一句「biological weapons cannot enter the planet's atmosphere」。
    > remake 沒有「生物武器」這個分類,**這條沒接**。

    這條擱置的理由現在看是誤判的:缺的不是規則,規則手冊寫得很完整;缺的是
    **「哪些科技算生物武器」這份名單**。而那份名單其實一直在執行檔裡,只是被我
    當成「不可解的枚舉」放過去了。

    ### category enum 是可以從成員反推的

    第 108 項從 `Calc_Tech_Value_` 挖出兩張表:`TechItemCategory[212]`(每項科技屬於
    哪一類)和 `TechCategoryDefaultMultiplier[49]`(每類的權重)。當時只用了權重,
    **category 的編號代表什麼意思沒解**——0..41 就是一串數字。

    但枚舉的語意可以從**成員**反推:把 212 項科技按 category 分組,同組有什麼共通點,
    表自己會說。分完之後 41 組每一組都是乾淨的功能類別(0=農業、1=工業、2=研究、
    26=光束武器、32=裝甲、33=護盾、39=燃料槽、40=士氣建築、41=索引 204–211 的填充)。

    而其中一組直接把這一項解開了:

    | category | 成員 | 組數 |
    |---|---|---|
    | **20** | Bio-Terminator、Death Spores | 2 |

    恰好兩項,不多不少。**分類是執行檔給的,不是我自己劃的**——這是這一項與「憑印象
    列個生物武器清單」的差別。`TestBiologicalWeaponsAreExactlyCategory20` 把兩者釘在一起。

    ### 手冊把剩下三件事都給齊了

    GAME_MANUAL.pdf p.99「Death Spores (System)」:

    > invading ships must introduce them into the target planet's atmosphere
    > **by orbital bombardment**. Each spore pod launched has a **10% chance to
    > kill one unit of colonist population**.

    加上後面效果表的兩個數字,三件事都有依據:**投放方式**(軌道轟炸)、
    **效果**(每莢 10% / 生物滅絕者 20% 殺 1 單位人口)、
    **反制**(屏障護盾「無法進入大氣層」——完全擋掉,不是減傷)。

    ### 一個實作上的陷阱:護盾要在建築吸收之前問

    `BombardColony` 的流程是「算傷害 → 建築吸收(按建築名字母序拆)→ 扣人口」。
    生物武器接在扣人口之後,若在那一步才查 `buildings` 有沒有屏障護盾,
    **「護盾擋不擋得住」就變成取決於字母序有沒有先輪到它被拆**——那是假的精確度,
    不是規則。所以 `bioBlocked` 與 `shield` 一樣在吸收迴圈之前就取好。

    `TestBombard_BarrierShieldDestroyedThisTurnStillBlocks` 守這條,而且它是 PASS 不是
    SKIP——那一波轟炸真的把護盾拆掉了,擋阻仍然成立。

    ### 誠實留白

    - **一次轟炸投幾莢沒有真值。** 手冊說「每一個發射出去的孢子莢」,但沒說一次投幾莢,
      而 remake 沒有「哪幾艘船掛了生物武器、各帶幾莢」的模型。取「艦隊艦艇數」
      (一艘一莢)是 **remake 的建模選擇**,寫在 `bioweapon.go` 檔頭與呼叫端註解裡。
    - **只有屏障護盾擋。** 輻射護盾與通量護盾的手冊敘述都沒有那一句。
      `TestOnlyTheBarrierShieldBlocksBiologicalWeapons` 把這件事釘住,免得日後
      被當成漏寫「順手補齊」。
    - **G 項的第二條仍然刻意不做**:護盾被炸掉後氣候不變回 Radiated。
      remake 沒有「可逆的建築效果」這層模型,這是既有的刻意偏離,不是這一項的遺漏。

    ### 沒有畫面回歸

    截圖廊零位元組差異——因為 `BioWeaponKills` 是單次轟炸結果、不是存檔欄位,
    也因為轟炸報告那一行只在真的殺到人時才多印一段(報告框高 118px、行距 24,
    第五行會掉出框外)。

112. **聊天列:資料模型與畫面都做完了,但字從來沒離開過這台機器**(2026-08-08)。

    `internal/netplay/chat.go` 從 `Receive_Chat_Msg_` @ 0xDD351 逐位元組解出了整個資料
    結構(保留 14 則、每則 82 byte、內文上限 80、發話者 ≥8 是 GNN),
    `cmd/moo2/netnextturn.go` 也把輸入行、游標、兩種前綴、版面全畫出來了。

    但 `sendChat` 的註解自己寫著:

    > ⚠ 目前**只加進本機記錄**……真的接上連線時,這裡再多一個 `WriteFrame(conn, …)`

    ### 缺的不是那一行 WriteFrame,是中間整層

    大廳(`lobby.go`)把大家連起來、發名冊,然後就結束了;回合表(`lockstep.go`)是
    純資料結構,自己不碰任何 socket。**連線建立之後,沒有任何 goroutine 在讀它。**

    所以「多一個 WriteFrame」寫得出去也沒人收——收訊端不存在。缺的是一層
    **訊息幫浦**:`internal/netplay/session.go`。

    ### 為什麼是 Poll 而不是回呼

    ebiten 的 `Update` 是單一 goroutine 每幀跑一次。讀 socket 一定要另一條 goroutine
    (否則畫面卡在 `Read` 上),但**規則狀態不能被那條 goroutine 直接改**——那是資料競爭,
    而且會讓鎖步的前提(同樣的指令序列算出同樣結果)失效:改動時機變成取決於封包
    什麼時候到。

    中間隔一個佇列,`Update` 每幀 `Poll()` 清空一次,**所有狀態變更都在 Update 這一條線上**。
    順序由佇列保證,與封包抵達時間無關。`-race` 下全綠。

    ### 星狀拓樸的轉發(移植決策,不是還原)

    原版 `Send_Chat_Msg_` @ 0xDD3B8 是逐一走過玩家陣列**直接發給每一個人**——IPX 廣播式
    的對等網路。remake 走 TCP:主機 listen、客戶端 dial,**客戶端之間沒有連線**。
    所以 A 說的話要讓 B 聽到,主機必須轉發。

    轉發帶出一個容易寫錯的地方,兩條測試各守一半:

    | 情境 | 正確做法 | 寫錯會怎樣 |
    |---|---|---|
    | 主機收到(第一手)| 以**連線**為準,蓋掉封包自報的 Player | 客戶端可以冒名說話 |
    | 客戶端收到(可能是轉發)| **信任**主機標的 Player | 所有人的話都顯示成主機說的 |

    還有一條正對照:訊息不轉回發話者自己——否則 Alice 會看到自己的話出現兩次
    (本機一次 + 繞回來一次),而「轉發給所有連線」的實作照樣能通過前面那條測試。

    ### 誠實留白

    - **幫浦建好之後再加入的人不會被納進來。** 延遲建立是為了不漏掉大廳階段後來的人,
      但代價是「開打後中途加入」不支援。remake 的對局開打後不收人,現階段碰不到。
    - **主機可以造假。** 客戶端信任主機標的發話者編號——星狀拓樸下主機本來就轉發所有流量,
      蓋不蓋 Player 都擋不住。這不是疏漏,是拓樸的必然。
    - 截圖廊零位元組差異:`netSession()` 在沒有連線時回 nil,畫面那一端全部容忍 nil。

113. **三個「沒有查到寫入端」的欄位,寫入端一直都在**(2026-08-08)。

    `docs/re/calc-tech-value.md` 第 6 節把四件事列為擋門,其中三件的理由都是
    「**沒有查到寫入端**」:

    | 欄位 | 當初的判定 |
    |---|---|
    | `[player+0x89F]` | 「未定名,種族特性相關(**猜**)」 |
    | `[player+0x28]` | 「0..6 的意義不確定……**不要猜是哪 7 種性格**」 |
    | `[player+0x205]` / `[player+0x206]` | 「**沒有查到寫入端**,不排除是別的東西」 |

    這一項做的事只有一件:**去 grep 那個寫入端**。

    ```
    grep -nE "mov +(byte ptr )?\[e..\+89Fh\]," Orion2.exe.asm
    ```

    四筆立即數寫入,當場就把欄位的身分講完了。

    ### `[player+0x89F]` = 政體

    `sub_E4204`(「取得某項科技」的效果套用函式,第一行就是
    `mov byte ptr [ecx+eax+117h], 3` ——把該科技標成已取得)按 techIdx 分支:

    | techIdx | 科技 | 寫入值 | asm 行 |
    |---|---|---|---|
    | 42 | Confederation | **1** | 327016 |
    | 92 | Imperium | **3** | 327006 |
    | 65 | Federation | **5** | 327021 |
    | 77 | Galactic Unification | **7** | 327011 |

    偶數是四個基本政體(封建/獨裁/民主/統一——正好是自訂種族那一欄能選的四個),
    奇數是它們的科技升級版,`值/2` 就是同一族。`Calc_Tech_Value_` 階段 F 那句
    `[+0x89F]/2 == 2` 因此讀得出來了:**民主/聯邦那一族**。

    ### 順帶驗到的第二件事:remake 的政體編號跟原版一模一樣

    `gamedata.AssimilationGovernment` 的順序當初是照手冊表格排的,**是 remake 自己的選擇**。
    對到執行檔的四個立即數之後 —— 1/3/5/7 全中,一項不差。

    這不是巧合能解釋的:兩邊各自從不同來源(手冊表格順序 vs 執行檔位元組)得到同一組編號,
    代表手冊那張表的排列本來就反映內部編號。已用 `government_orig_test.go` 釘住,
    順便釘住 `MoraleGovernmentType` 與它同號(原版只有一個 `[player+0x89F]`,
    Go 這邊分兩個列舉是歷史,不是原版有兩套)。

    ### `[+0x28]` / `[+0x205]` / `[+0x206]`:同一支函式的三次加權抽選

    三個欄位都在 `Init_NPC_Personalities_Objectives_Themes_` @ 0x589D6 裡被寫,
    形狀完全一樣,連寫三次:

    1. 把一組候選權重加總(`cmp eax,18h`/`10h`/`1Ch` → **6 / 4 / 7** 個候選);
    2. 總和 > 1000 就把每一格減半再算一次(防溢位);
    3. `sub_1247A0(總和)` = 原版的 `Random_`,回 1..總和;
    4. 逐格扣到 ≤ 0,扣了幾格就是抽中的索引;
    5. 寫進 `[+0x28]` → `[+0x205]` → `[+0x206]`。

    三次抽選的權重都先被 `byte_199CB0`(難度)加減過(`cmp byte_199CB0, 3` / `4` 各一段)
    ——難度越高,某些候選越容易被抽中。這同時再次佐證 `byte_199CB0` 是難度。

    而 `[+0x28]` 那一次有一個明文豁免:

    ```
    cmp  byte ptr [eax+28h], 64h   ; 已經是 100(人類)?
    setnz dl
    jz   loc_58E39
    mov  [eax+28h], cl             ; 不是 → 寫抽選結果
    ...
    loc_58E39: mov byte ptr [eax+28h], 64h   ; 是 → 把 100 寫回去
    ```

    **「100 = 人類」不只是既有結論,這裡是原版自己在保護它。**

    ### 誠實留白:名字對得上,但哪個是哪個沒有證據

    函式名列了三件事(Personalities / Objectives / Themes),而裡面正好三次抽選。
    **但「6 個候選的那次是不是性格」沒有進一步證據** —— 符號名是二手推論。
    這一項確定的是「三個欄位各自從 6 / 4 / 7 個候選裡加權抽一個,難度會改權重」,
    候選各自代表什麼仍未解,`calc-tech-value.md` 階段 C/D/E 的常數表因此**仍然不能照抄**。

    擋門從「查不到欄位怎麼來的」降級成「查不到候選代表什麼」——**是進展,不是解決**。

    ### 第三件:GNN 聊天封包的號碼記錯了

    `internal/netplay/chat.go` 寫著「GNN 走的是 case 43(`Send_GNN_Chat_Msg_` @ 0xDD42A)」,
    常數是 `ChatGNNOpcode = 0x2B`。逐項核對之後,那句話**把送訊端和收訊 case 湊成了一對,
    而它們不是同一個號碼**:

    | | 號碼 | 出處 |
    |---|---|---|
    | `Send_GNN_Chat_Msg_` 送出的型別 | **0x2D(45)** | asm 315237 `mov edx, 2Dh` |
    | 收訊跳表裡「speaker 固定 8」那一格 | **0x2B(43)** | `mov eax, 8; jmp` → `Receive_Chat_Msg_` |

    而那張 68 格跳表**沒有 case 45**(實際有的是 0..67 扣掉 4/45/49/51/53/54)。
    一般聊天則兩端對得起來(送 0x27 = 39,收 case 39 呼叫 `Receive_Chat_Msg_`)。

    ⚠ **不要把這讀成「原版的 GNN 廣播是壞的」**:我核的是這一條路徑上收發號碼對不起來,
    沒有核 `sub_F6816` 是不是唯一的傳送出口、也沒有核有沒有第二個派送器。
    能確定的只有一件:先前那個 `0x2B` 是**收訊 case 號**,被當成送訊型別號記下來了。
    常數已拆成 `ChatGNNSendOpcode` / `ChatGNNRecvCase` 兩個,不再假裝它們是同一個數。

    ### 這一輪的形狀

    沒有新工具、沒有新方法,就是**在下「查不到」這個結論之前多跑一次 grep**。
    三條擋門裡有兩條半是這樣解開的。

114. **AI 對手的科技,先前完全是偷來的**(2026-08-08)。

    這一項不是從逆向來的,是從**一個探針**來的。第 113 項訂正 `internal/ai/research.go`
    的過期斷言時順手看了一眼呼叫端,發現 `ai.DecideResearchTopic` **零個呼叫端**。

    跑一個 200 回合的探針:

    | 回合 | AI 0 主題 | AI 0 已完成 | 玩家主題 |
    |---|---|---|---|
    | 1 | 1 | 7 | 3 |
    | 20 | **1** | 7 | 3 |
    | 50 | **1** | 9 | 31 |
    | 100 | **1** | 11 | 9 |
    | 200 | **1** | 13 | 15 |

    **AI 的研究主題 200 回合都沒有換過。**

    ### 兩個洞

    **① 沒有人替 AI 選下一個主題。** `engine.RunEmpireTurn` 對 AI 一樣跑研究階段,主題完成
    也會標進 `CompletedTopics`——但 `ResearchTopic` 是誰設的?玩家那邊是研究畫面設的,
    AI 那邊**沒有**。所以 AI 每一回合把研究點投進一個早就完成的主題,無限重複完成同一項。

    那上表的「已完成」為什麼還是從 7 長到 13?**它偷的。** `spy.go` 的間諜每隔一段時間從
    玩家那裡偷一個主題——AI 對手的整條科技線,先前**完全來自玩家**。玩家不研究,
    AI 就永遠停在開局。

    **② 多選主題的抉擇永遠掛著。** `engine.recordCompletion` 對多選主題預設記第一項並掛起
    `HasPendingChoice`,等玩家去選。AI 從來沒有人替它選,所以它的 `ExplicitChoice` 永遠是空的
    ——而 `groundEquipTechOwned` 的判定是「主題完成 + **沒有明確抉擇** → 視為擁有」。
    也就是說 AI 每完成一個三選一主題,實際上**三個都拿到了**。

    修這一條反而是**削弱** AI,但那才是規則。

    ### 為什麼這在畫面上看不出來

    沒有任何 UI 顯示 AI 在研究什麼。但它讓每一個「吃 AI 科技」的系統都凍在第 1 回合:

    - `retaliationAttackers` → `bestUnlockedWeaponValue`:軌道防禦永遠用開局武器
    - `aiFleetSpeedParsecs`:AI 艦隊永遠是核融引擎的速度
    - `groundEquipTechOwned`:AI 地面部隊永遠沒有動力裝甲

    **三個系統都已經寫好了會讀科技,只是讀到的永遠是同一份。**

    ### 挑哪一項用原版的值,挑哪個主題還是設計的

    主題選擇沿用 `ai.DecideResearchTopic`(吃性格的設計啟發式)。原版的 `Calc_Tech_Value_`
    估的是**科技應用項**不是主題,而且它唯一把主題等級併進結果的那幾段(階段 I/J/K)
    全部乘在語意未解的 `word_1AB1xx` 上——**仍然不可照抄**。

    但**應用項的抉擇**用得上一手值:`gamedata.TechCategoryWeight` 就是階段 B 給 `ecx` 的
    初始值(`byte_17D196[category*2]`),而那一段是 `calc-tech-value.md` 第 7 節明列的
    「風險遠低於其他階段、可以先照抄」。同分取先出現的那個,不擲骰——整條 AI 研究線
    因此是決定性的(`determinism_test.go` 那組閘門的前提)。

    ### 一個容易寫成「只修一半」的地方

    重挑的觸發條件不能只看「本回合有研究完成」:主題也可能是**被間諜偷來**的
    (`spy.go` 直接寫 `CompletedTopics`,不經過研究階段),那時候 `ResearchDone` 是 false
    但目前主題已經完成了。只看 `ResearchDone` 會讓 AI 卡在一個偷來的主題上繼續投點——
    正是這一項要修的病,換個入口又長回來。所以改成每回合檢查「目前主題完成了沒」。
    `TestAIRepicksAfterStealingItsCurrentTopic` 守這條。

    ### 修好之後

    | 回合 | AI 0 主題 | 已完成 | 明確抉擇 |
    |---|---|---|---|
    | 50 | 1 | 9 | 2 |
    | 200 | **66** | **17** | **9** |

    截圖廊零位元組差異(截圖是第 1 回合的狀態,AI 研究還沒動)。

115. **領袖永久免費——用覆蓋率找出來的**(2026-08-08)。

    第 114 項是靠「這支函式有沒有呼叫端」找到的。那個問法有個明顯的盲點:
    `ai.DecideResearchTopic` **有**呼叫端(`RemakeDecider.ResearchTopic`),
    只是那個呼叫端自己也沒人叫。所以先做了兩輪掃描:

    | 問法 | 結果 |
    |---|---|
    | 生產碼裡零次提及的函式 | 2 個(`bitBytes`、`homeworldBuildingsLegacy`)|
    | 從 `cmd/` + 頂層宣告出發、傳遞閉包到不了的 | 一樣是那 2 個 |

    **名字層級是乾淨的。** 但 114 那類洞不是「名字沒被提到」,是「**從來沒有執行過**」
    ——所以換一個問法:跑一局 300 回合,量覆蓋率。

    ```
    go test -run TestZZLongGameForCoverage \
      -coverpkg=./internal/shell,./internal/engine,./internal/gamedata -coverprofile=...
    go tool cover -func=... | awk '$3=="0.0%"'
    ```

    539 支函式在一局自動跑完的對局裡一次都沒執行。**大多數是應該的**——
    軌道轟炸、地面入侵、訓練間諜這些是玩家主動觸發的,自動迴圈不會碰。
    所以真正的訊號不是「零覆蓋」,是「**零覆蓋 × 這件事應該自動發生**」。

    ### 逐一看過之後,兩個判斷

    **`engine.CheckVictory` 零覆蓋——但這不是洞。** 它是把三條勝利條件包在一起的便利函式,
    而 shell 有自己的三條路徑(`advanceConquestVictory` 走 `CheckExtermination`、
    `advanceCouncilElection`、`advanceAntaranVictory`),三支都每回合跑。
    優先序一致性已有測試釘著(`antaran_victory_test.go`)。**留著它是個小陷阱**
    (看起來權威、實際上不是被用的那個),但不影響玩法,這一輪不動它。

    **`gamedata.LeaderMaintenanceCost` 零覆蓋——這個是洞。**

    ### 領袖雇一次之後永久免費

    這支從 openorion2 的 `GameState::leaderMaintenanceCost` 移植過來,有單元測試,
    註解連 `LEADER_ID_LOKNAR` 那個不移植的硬編特例都交代了——**零個生產端呼叫**。

    手冊在講「遇難領袖」事件時把這條規則講得很明白:

    > In gratitude for the rescue, this leader joins your empire for **no hiring cost**.
    > You are still expected to **pay maintenance**, however.

    那句話之所以要特別寫,正是因為「免雇用費」與「免維護費」是兩件事。而 remake 的
    `grantMaroonedLeader` 給的領袖,先前兩樣都免。

    每位 `ceil(hireCost/100)`(下限 1),有 Megawealth 技能者免費。`hireCost` 走與
    `MercHireCost` **同一條公式**——兩邊用不同基準的話,同一位領袖會出現
    「雇用時算貴、維護時算便宜」這種對不起來的狀況。

    ### 誠實留白

    - **付不出來只是扣到 0,領袖不會離職。** 原版錢不夠會怎樣沒查到規則(手冊只說要付,
      沒說不付的後果),不自己發明一個懲罰。
    - **AI 不付。** `AIOpponent` 沒有 Leaders 欄位——不是漏掉,是那一整層還不存在。
    - 判定 Megawealth 走 `leaderSkills` 那條既有路徑,不比對技能字串:標籤會被翻譯,
      拿它當識別鍵在英文模式下查不到(`Leader.Skills` 檔頭記過這個坑)。

    ### 一個差點寫錯的地方

    國庫已經是負的時候不能再扣。`session.go` 對戰損 `bcLoss` 有一段註解記著同一個坑:
    「BC 為負時若只判斷 `bcLoss > BC` 會把損失夾成負值,`BC -= bcLoss` 反而變成**加錢**」。
    這裡直接沿用那個既有慣例,並有一條測試專門守「不該變多」。

    ### 畫面

    帝國摘要多一列「領袖薪餉」,單獨列不併進「維護支出」——那一列是**建築**維護,
    來源不同,混在一起玩家就看不出「解雇一個領袖能省多少」。
    截圖廊零位元組差異(新開一局沒有領袖,第 43 項起改成原版的雇用制)。

116. **征服來的人口先前是「免費全額生產」**(2026-08-08)。

    第 115 項用覆蓋率找洞時只看了「應該自動發生」那一類。同一份清單裡還有 196 支
    `internal/gamedata` 的函式零覆蓋——那些是**純規則**,零覆蓋的意思是
    「這條規則在一局裡從來沒有適用過」。

    大多數是應該的(軌道轟炸、地面戰、武器改造都要玩家主動觸發)。但兩支不是:

    | 函式 | 呼叫端 |
    |---|---|
    | `ProdAlienWorkerOutput` | **0 個** |
    | `ProdWorkerOutput` | **0 個** |

    ### 手冊寫的是 produces,不是 produces industry

    > Aliens appear in an enemy's colony that you've conquered, representing the population
    > left there by the former owner. At first, all aliens are **uncooperative**. Until they
    > are integrated into your empire, **each alien unit produces only three quarters what
    > it normally would**.

    remake 有 `UnassimilatedPop`(第 105 項的叛亂系統與多種族士氣懲罰都在用),
    但這條產出懲罰**沒接**——打下一座殖民地,當回合就拿到它的全額產能。
    「征服打法要付的利息」先前只有叛亂風險那一半。

    動詞是 **produces**,沒有限定工業,所以食物/工業/研究三項都套。

    ### 第二支函式是被第一支帶出來的

    `ProdWorkerOutput` 是「每工人至少 1 產能」的下限。它零覆蓋不是因為沒接
    ——是因為**永遠不會生效**:礦產表最低就是 1(`mineralProductionTable = {1,2,3,5,8}`),
    沒有任何東西能把 per-worker 壓到 1 以下。

    3/4 之後 `1 × 3/4 = 0` —— 極貧星上全員未整合的殖民地會產出**零工業**。
    下限這才真的擋到東西。兩支函式是同一條規則的兩半,一起接才對。

    ⚠ 下限只套**工業**:`ProdWorkerMinimum` 的手冊依據講的是「每個**工人**單位」,
    沒有講農夫與科學家,不擅自套到那兩項。食物/研究因此真的會被壓到 0——
    測試把這個不對稱釘住,免得日後被當成 bug「順手補齊」。

    ### 誠實留白:誰在做哪個工作

    手冊講的是「每個 alien **單位**」,而 remake 沒有「哪些外星人在做哪個工作」的模型
    ——`ColonyState` 只有 `UnassimilatedPop` 與三個職業人數,沒有交叉表。
    所以按**人口比例**把外星人攤到各職業(向下取整)。**那是 remake 的建模選擇**,
    不是手冊給的分配規則;手冊也沒說玩家能不能指定外星人的工作。

    ### 一條既有測試在這裡變成錯的

    `TestConqueredMarkerDoesNotLeakIntoEconomy` 斷言「征服標記不該改變經濟結算」,
    而它測的是整支 `markColonyConquered`——**那支同時設 `UnassimilatedPop`**。

    這條規則接上去之後,那個斷言反而變成錯的:未整合人口**本來就該**改變經濟結算。

    處理方式不是把它刪掉,是**把兩件事拆開**:記帳用的 `ConqueredFrom` /
    `ConqueredFromKnown` 仍然不該有經濟效果(原意保留);`UnassimilatedPop` 的經濟效果
    另立一條正對照,並多驗一件事——**同化完要恢復正常產出**,懲罰是暫時的不是永久烙印。

    截圖廊零位元組差異(截圖裡沒有被征服的殖民地)。

117. **一個擋門理由過期了三個月都沒人回頭看**(2026-08-08)。

    覆蓋率清單裡有 36 支 `gamedata` 函式**零覆蓋而且零呼叫端**。其中三支是同一個系統的:

    ```
    SpyGovernmentDefenseBonus
    SpyRaceTraitBonus
    SpyTechnologyBonus
    ```

    間諜系統(第 29 項)每回合都在跑。三支手冊加成一支都沒接。

    ### `spy.go` 自己寫了為什麼

    > 種族特性 / 科技 / 政府三項現行 remake 無法從 AIOpponent/engine.PlayerState 推導出
    > 對應資料(無種族間諜特性強度資料、**無逐科技模型可查是否擁有 spy.go 列的 5 項科技**、
    > AIOpponent 無政府型態欄位),一律回 0——TODO,待補上這些欄位後在此函式接上。

    寫得很清楚,而且**當時大概是對的**。問題是它從此沒有再被檢查過。

    逐條核對現況:

    | 擋門理由 | 現在還成立嗎 |
    |---|---|
    | 無逐科技模型 | ❌ **不成立**。`groundEquipTechOwned` 已是三個系統共用的判定(生物武器、地面裝備、進階政體),那 5 項科技在 `enums.go` 都有常數 |
    | 無種族間諜特性強度 | ✅ 仍成立。`TRAIT_SPYING` 只標記「有沒有」,沒有 −3/+3/+6 的分級,`Races` 表也沒這一欄 |
    | AIOpponent 無政府型態 | ✅ 仍成立(對 AI)。但**玩家有** `s.Government`,而手冊那一欄本來就只給 Defense |

    三條裡有一條過期。接上的就是那一條 —— 外加玩家側的政府加成。

    ### 政府那一項要走「進階政體」

    直接用 `s.Government` 會讓研究出帝國的獨裁玩家永遠拿獨裁那一格,而手冊對基本型與
    進階型給的是**不同的值**。所以走 `assimilationGovernment()` —— 那支已經在處理
    「基本政體 + 對應科技 → 進階形式」。

    這一步靠的是第 113 項的結果:`SpyGovernmentType`、`AssimilationGovernment`、
    `MoraleGovernmentType` 三個列舉編號相同,不是巧合——**原版只有一個 `[player+0x89F]`**,
    Go 這邊分成三個是歷史。新加一條測試把三者釘在一起,免得其中一個被重排時
    政府防諜加成安靜地查錯格。

    ### 「加總」是讀法,不是手冊原文

    手冊 Technology 那一列給 5 項科技各自的值,**沒說它們是不是累加**。這裡採加總,
    理由是那 5 項互不相關(神經掃描器、隱形衣、心靈學…),不是同一件事的三個階
    ——取最佳的話研究第二項就完全沒有意義。

    ⚠ 這與領袖技能那邊**刻意不同**:那裡手冊明寫「Megawealth 與 Researcher 是累加的」,
    反面暗示其餘取最佳(見 `gamedata/leader_skill_apply.go`)。這裡沒有那句話,所以
    這是推論不是引用,標在 `spyTechBonusFor` 的註解裡。

    ### 這一輪的形狀

    第 114–116 項是「找到沒接的東西」。這一項不一樣:**東西寫好了、擋門理由也寫好了,
    只是那個理由已經不成立**。`rulebook/63`(真相在程式碼裡,不在過期標記裡)講的正是
    這件事——而它在自己的 TODO 註解上同樣適用。

    截圖廊零位元組差異。

118. **「成就」科技的全帝國效果,四條一條都沒接**(2026-08-08)。

    第 117 項那個形狀又出現一次:**規則寫好了、擋門理由也寫好了,而理由已經不成立**。

    | 手冊規則 | 出處 | 呼叫端 |
    |---|---|---|
    | 虛擬實境網路:全帝國士氣 +20% | p.97-98 | **0** |
    | 心靈學:特定政體下士氣 +10% | p.100-101 | **0** |
    | 微晶構築:每個工業工人 +1 產能 | 手冊 | **0** |
    | 奈米分解者:行星污染容忍值 ×2 | 手冊 | **0** |

    `colonyMoralePercent` 的檔頭寫了為什麼:

    > Virtual Reality Network(全帝國 +20%,p.97-98):手冊定性為「成就」而非一般建築,
    > 不在 gamedata.Buildings 清單、**remake 也無「成就」追蹤系統,無從得知是否擁有**,故不套用。

    **「成就」在 MOO2 就是科技。** 它們不進建造清單是因為研究出來就自動生效,不是因為它們
    是另一種東西。而「有沒有研究出來」remake 一直查得到——`groundEquipTechOwned` 已經是
    **四個**系統共用的判定(生物武器、地面裝備、進階政體、間諜科技加成)。

    ### 順帶修掉一個同款的士氣 bug

    `colonyMoralePercent(s.Government, ...)` 傳的是**基本政體**。而手冊的士氣表對基本型與
    進階型給的是不同的值——**帝國比獨裁多 +20% 全帝國士氣**。所以研究出帝國的獨裁玩家,
    先前永遠拿獨裁那一格。

    這與第 117 項在間諜那邊修的是同一個 bug,只是在另一個系統。抽了一支
    `effectiveGovernment()` 出來,凡是查政體表的地方都走它。

    ### 成就要每回合重算,不能像建築那樣「完工時設一次」

    建築蓋好就永遠在;科技**會被偷、會被交換**,而 `groundEquipTechOwned` 的判定還吃
    「有沒有做過明確抉擇」——那也會變。所以 `syncAchievementColonyFields` 每回合重算,
    冪等,舊存檔讀進來也會自動補齊。有一條測試專門驗「科技沒了效果要收回」。

    ### 寫測試時撞到的一件事:微晶構築與奈米分解者互斥

    原本想一次給玩家兩項來測,結果後設的那項把前一項的 `ChosenTech` 蓋掉了。
    查下去才發現**它們是同一個三選一主題**(TOPIC 53,選項 108/113/203)——
    科技樹本來就不准同時擁有。

    **不是實作的 bug,是測試的前提錯了。** 已經另立一條測試把這個互斥關係釘住,
    免得日後有人「順手」讓兩項同時生效。

    ### 截圖廊:一張圖變了,而且是預期內的

    `30_netwait.png` 差了 **134 個像素,全部落在 x 83–129 / y 229–237** ——
    那正是狀態指紋的位置。`ColonyState` 多了兩個欄位,存檔 JSON 就變了,指紋跟著變。
    逐像素核對過範圍才重錄,不是看到 diff 就覆蓋。其餘 33 張零位元組差異。

119. **註解說「打得準也閃得掉」,程式只做了前半**(2026-08-08)。

    覆蓋率清單裡剩下的兩支艦員加成:

    ```
    ShipCrewBoardingBonus         0 個呼叫端
    ShipCrewMissileEvasionBonus   0 個呼叫端
    ```

    追下去才發現同一段還藏著第三件事——**`ShipCrewDefenseBonus` 也沒接**,
    而且那一項比另外兩支更難發現:它**有**呼叫端,在 `engine.BeamDefense` 裡。
    只是 `engine.BeamDefense` 在一局裡從來沒被執行過(shell 的戰鬥自己算,沒走它)。

    ### 註解與程式對不起來

    `mkPlayerCombatantsIndexed` 寫著:

    ```go
    // 艦員經驗(手冊 p.121 的 BA/BD 兩欄):老手打得準也閃得掉。
    crew := s.shipCrewLevel(sh)
    atk += gamedata.ShipCrewOffenseBonus(crew)
    ```

    註解說**兩欄**,程式只加了 BA 那一欄;`def` 從頭到尾只有艦體值。
    手冊 p.121 的 BA/BD 是分開的兩個加成,openorion2 的 `Ship::beamDefense` 也是這樣算的。

    **這種 bug 特別難抓**:註解讀起來完全正確,測試也不會紅——沒有人寫過「老手的防禦
    應該比新兵高」這條斷言。它是被「哪些 gamedata 函式從來沒執行過」問出來的。

    ### 飛彈閃避:擋門理由對了一半

    `ResolveMissileShot` 的 `defenderEvasionBonus` 一直恆傳 0,理由是:

    > 現行 remake 的艦艇設計/軍官系統尚未提供這些元件,呼叫端目前一律傳 0

    手冊列的閃避來源有六項。逐項核對:

    | 來源 | 現況 |
    |---|---|
    | ECM 干擾器 / 慣性穩定器 | ✅ 仍缺(SpecialOptions 沒有這些元件)|
    | 種族 Ship Defense 特性 | ✅ 仍缺(手冊連檔位名稱都沒列)|
    | **艦員經驗** | ❌ **算得出來**——`shipCrewLevel` 每回合都在更新 |
    | **舵手(Helmsman)軍官** | ❌ **算得出來**——`SKILL_HELMSMAN` 在第 103 項就進來了 |

    接上後兩項。舵手那一項是技能值的**一半**(手冊:「Half bonus of the Helmsman value」),
    而且**只算艦艇軍官**——舵手是開船的,殖民地領袖不會坐在艦橋上,這與 `starlane.go`
    挑領航員是同一條規則。多位取最佳不加總(手冊只在 Megawealth/Researcher 明說累加)。

    ### 登艦戰那一支仍然不接,而且理由是好的

    `ShipCrewBoardingBonus` 零呼叫端,但 `crew.go` 自己寫了為什麼:

    > 登艦戰先前沒有建模所以沒人抄。抄進來,即使暫時沒有呼叫端:
    > 缺一條軌會讓下次有人接登艦戰時以為手冊沒給數字。

    **remake 確實沒有登艦戰機制**(grep 過整個 repo,零命中)。這不是「忘了接」,
    是「那個系統還不存在」——與前幾項的「理由過期了」不同,這個理由現在仍然成立。
    不為了讓覆蓋率好看而硬接一個沒有消費端的加成。

    截圖廊零位元組差異。

120. **兩個「為了給玩家看」而寫的函式,畫面從來沒用過**(2026-08-08)。

    覆蓋率清單剩下的 36 支零呼叫端函式,這一輪逐支看完並分類。**大多數是合理的**:

    | 類別 | 例子 | 為什麼不接是對的 |
    |---|---|---|
    | 子系統還不存在 | `ShipCrewBoardingBonus`、`ResolveGroundBattle` | 登艦戰沒建模;地面戰走的是另一條原版路徑 |
    | 手冊資料不全 | `DamageEngineExplosionPotential` | 手冊沒給每格衰減率,接了就得自己編一個 |
    | 元件沒建模 | `ComputerHP`、`DamageShieldCapacity` | 艦艇子系統 HP 不在模型裡 |
    | 純存取器 | `PlanetSpecialWeight`、`WeaponModSpaceCostPercent` | 表本身有人用,只是沒人用這支包裝 |
    | 刻意的偏離 | `PlanetaryShieldEffectiveClimate` | 建築效果不可逆(第 98 項記錄過)|

    **兩支不是。** 而它們的檔頭自己說了為什麼會存在:

    > `AssimilationProgressNeeded`:抽出來是因為 **UI 要顯示**「這個殖民地還要幾回合才完全
    > 同化」——一個只在背景默默跑的機制對玩家等於不存在。

    > `CrewXPToNextLevel`:這個函式存在是為了**讓 UI 能顯示進度**——一個只會默默上升的
    > 等級對玩家等於不存在。

    **兩支都是為了畫面而抽出來的,而畫面從來沒有呼叫過。** 那兩句話因此在描述自己:
    同化與艦員經驗這兩個機制,對玩家來說一直等於不存在。

    艦員經驗更明顯——第 119 項才剛讓它影響命中、防禦與飛彈閃避三件事,而玩家**沒有任何
    地方看得到自己的艦員是什麼等級**。

    ### 接法

    - **殖民地畫面**:未同化人口 > 0 時多一行「未同化 N ／還需 M 回合」。
      已累積的進度要扣掉,否則每回合看到的數字都一樣,像是完全沒在推進。
    - **星圖資訊面板**:接在既有的「艦隊陸戰隊 N」那一行後面,顯示艦隊的艦員等級與
      距離升級還差多少經驗。取艦隊裡**最低**的那一艘——戰力由最弱的那條線決定,
      報最高的會讓玩家高估自己。

    ### 誠實留白

    - **沒有逐艦資訊面板。** 艦員等級目前只有「整支艦隊的最低值」這個摘要。
      要做逐艦顯示得先有那張畫面,不在這一輪。
    - **星圖那一行沒有視覺回歸覆蓋。** 截圖廊零位元組差異,而那代表那個面板在截圖的
      狀態下沒有展開(它要選中一顆有行星的星才畫)。shell 側的查詢有單元測試釘住,
      但「畫出來長什麼樣」沒有 byte-diff 保護——與第 109 項的軍官畫面同一個狀況,記在這裡。

121. **護盾減傷五級裡有四級是錯的,而正確的常數就躺在旁邊**(2026-08-08)。

    第 120 項把「零呼叫端的**函式**」清光了。但那個掃法有個洞——第 118 項的
    `MoraleVirtualRealityNetworkBonus` 是個 **const**,不是函式,當初是碰巧發現的。

    所以換個掃法:**gamedata 裡零消費的匯出常數/變數**。190 個。

    大多數是**原版列舉的鏡像**(`BUILDING_*`、`SPEC_*`、`TRAIT_*`、`DIPLO_*`、`STAR_*`),
    那些本來就是對照用的參考表,remake 用名稱字串當鍵——不是缺口。

    但其中一組是**活的戰鬥數字**:

    ```
    DamageShieldReductionClassI   = 1
    DamageShieldReductionClassIII = 3
    DamageShieldReductionClassV   = 5
    DamageShieldReductionClassVII = 7
    DamageShieldReductionClassX   = 10
    ```

    抄進來之後零消費。而 shell 那邊自己算了一套:

    ```go
    func shieldReduceByName(name string) int {
        for i, c := range ShieldOptions {
            if c.Name == name {
                return i * 2      // ← 清單索引 × 2
            }
        }
    }
    ```

    | 護盾 | 索引 | 索引×2 | 手冊 |
    |---|---|---|---|
    | 第一級 | 1 | **2** | **1** |
    | 第三級 | 2 | **4** | **3** |
    | 第五級 | 3 | **6** | **5** |
    | 第七級 | 4 | **8** | **7** |
    | 第十級 | 5 | 10 | 10 |

    **五級裡有四級是錯的,而且一律偏高**——只有最高級剛好對上。

    ### 元件名稱本身就寫著答案

    「第一級護盾／第三級護盾／第五級護盾／第七級護盾／第十級護盾」——那些數字就是
    Class I/III/V/VII/X,而手冊說得很直白:**每次攻擊減傷 = 等級數字**。

    `i * 2` 這個式子之所以看起來合理,是因為它在最高級剛好對上,而且單調遞增。
    **一個錯得有規律的公式比一個明顯壞掉的公式難發現得多。**

    ### 改用科技當鍵,不是名稱也不是索引

    - **索引**會被清單順序決定——那正是這次出問題的地方,以後插一級進去就全錯位。
    - **名稱**會被翻譯(英文模式下 `"第一級護盾"` 查不到)。

    所以 `gamedata.ShieldReductionForTech(TECH_CLASS_*_SHIELD)`。有一條測試直接比對
    「查名稱」與「查科技」兩條路徑的結果,確保它們不會再分岔。

    ### 截圖廊零位元組差異

    戰鬥數字改了但截圖沒變:截圖廊沒有進行中的戰鬥畫面,而護盾減傷只在逐發解算時生效。
    這也說明**視覺回歸抓不到這一類 bug**——它要靠「這個手冊常數有沒有人在用」才問得出來。

122. **一支偵察艦隊與一支末日之星艦隊,登陸能力完全相同**(2026-08-08)。

    第 121 項修的是「shell 自己編了一個數字,而手冊值就在 gamedata 裡沒人用」。
    順著同一族往下查(`armorHPByName` / `shipStrength` / `MarineTransportCapacity`),
    第三支中了同一個形狀,而且**更明顯**:

    ```go
    func (s *GameSession) MarineTransportCapacity() int {
        return len(s.Fleet().Ships) * gamedata.GroundTransportShipMarineCapacity  // 艦數 × 4
    }
    ```

    每一艘船不分艦體一律載 4 個陸戰隊單位。

    ### 擋門理由問錯了問題

    那段註解寫著:

    > ⚠ 簡化待精修:本 remake 尚無獨立的「運輸艦」船體類別……故無法像手冊那樣
    > 「每艘 Transport Ship 恰配 4 個 Marine 單位」精算。

    它假設運力來自**一種專門的運輸艦**。但手冊 p.121 的艦艇設計表**每一種艦體等級都有
    一欄 Marines**:

    | 艦體 | 護衛艦 | 驅逐艦 | 巡洋艦 | 戰艦 | 泰坦 | 末日之星 |
    |---|---|---|---|---|---|---|
    | Marines | 5 | 8 | 12 | 20 | 30 | 50 |

    **運力是艦體本身的屬性,不是某個艦種的特權。** 而 remake 一直有艦體等級——
    `shipClassFromName` 那支函式本身就是為了查**同一張表**的另一欄(Space)寫的。

    所以先前的效果是:一支護衛艦隊與一支末日之星艦隊的登陸能力完全相同,而手冊上差 10 倍。

    ### 為什麼這張表可以直接用

    那張表的「Comp.」與「Drive」兩欄**逐格對上** openorion2 的 `computerHPTable` /
    `driveHPTable`(`docs/tech/ship-design-space.md` 早就做過這個交叉驗證)。
    同一張表的其他欄位可信度相同——**不是「看起來像真的」,是有第二個獨立來源對過。**

    ### 同族的另外兩支:查過了,不動

    - **`armorHPByName`**(裝甲 HP 依元件):手冊 p.121 有 Armor/Struct. 欄(4/10/30/50/80/150),
      但那是**依艦體**的基準值,而裝甲科技的倍率**手冊沒給、openorion2 也沒有**
      (它是渲染殼,沒有戰鬥引擎)。只接一半會做出一個半知半解的模型,**不動**。
    - **`shipStrength`**(戰力點 1/2/4/8/16/32/64):它的註解本來就寫明是「供最小戰鬥解算」
      的抽象,沒有宣稱手冊忠實度。**那是誠實的簡化,不是錯值**,不改。

    兩者的差別就是這一項的判準:**宣稱是真值卻不是** vs **明說是抽象**。前者要修,後者不用。

    截圖廊零位元組差異。

123. **武器傷害全是估計值,而手冊原表一直在可抽文字的 PDF 裡**(2026-08-08)。

    第 121/122 項的形狀是「shell 編了一個數字,手冊值就在旁邊沒人用」。這一項更遠一步:
    **手冊值連抄都沒抄過**,而理由是一句錯的話。

    `docs/tech/component-values.md` 記著:

    > 其餘武器 Value 為依科技階遞增的**單調估計**(雷射 4→死光 25),保持排序合理與遊戲可玩,
    > 但未經手冊逐條核對。

    待辦第一條:

    > - [ ] OCR 掃描版手冊武器/裝甲/護盾附錄(若附錄存在於該 PDF;**9 頁本可能不含,需找完整手冊**)

    **不需要 OCR,也不需要另找手冊。** `moo2_patch1.5/GAME_MANUAL.pdf` 是可直接抽文字的
    (44 萬字元),武器表就在 p.124-127。而且**同一頁的 Size 欄第 46 項就抽過了**
    (`WeaponSpaceByName`)——當時抽了 Size,沒順手抽旁邊的 Damage 欄。

    ### 估計值錯得有多遠

    | 武器 | 舊估計 | 手冊 |
    |---|---|---|
    | 雷射 | 4 | 1-4 ✓ |
    | 質量投射器 | 8 | **6**(固定) |
    | 中子爆破槍 | 12 | 3-12 ✓ |
    | 核融合光束 | 16 | **2-6** |
    | 核飛彈 | 6 | **8** |
    | 麥克萊特飛彈 | 17 | **14** |
    | 高斯砲 | 18 | 18 ✓ |
    | 相位砲 | 19 | **5-20** |
    | 死光 | 25 | **50-100** |

    ### 最嚴重的不是偏差,是排序被弄反了

    **核融合光束在 remake 比中子爆破槍強(16 > 12),手冊上它比中子爆破槍弱(6 vs 12)。**

    這是單調估計法**必然**會犯的錯:它假設「科技越後面越強」,而 MOO2 的武器線本來就不是
    單調的——核融合光束是便宜的中階光束(成本 6),中子爆破槍貴一截(成本 8)還附帶
    「kills marines」。手冊把它們排在一起是因為用途不同,不是因為一個比一個強。

    有一條測試專門釘住這個反轉:`TestWeaponLineIsNotMonotonic`。有人「順手」把表改回
    單調就會紅。

    ### 沒有動的東西

    - **Cost 欄不動。** 手冊的 Cost(雷射 5、死光 75)與 remake 的生產成本(20/300)
      不是同一個單位——remake 的成本尺度是整條建造系統在用的,換掉會牽動生產節奏,
      而手冊沒給兩者的換算。
    - **`atk = 艦體 + 武器` 的合成不動。** 大艦在原版是靠**裝更多把武器**變強,
      而 remake 一艦一武器。艦體那一項是「掛載數」的替身;拿掉它會讓末日之星與護衛艦
      掛同一把雷射時傷害相同——那是另一個方向的失真。**修武器那一項,不動替身。**
    - **手冊還列了 remake 沒有的武器**(離子脈衝砲、引力波束、干擾者、重錘裝置、粒子束),
      擴充元件表不在這一輪。

    ### 版本相依那一項仍然分版

    電漿砲 1.31 為 6-30、1.50 為 4-20(`MANUAL_150.html` 明載),仍由
    `RuleProfile.PlasmaCannonMaxDamage` 覆寫。新加一條測試驗「**只有**電漿砲隨版本變」
    ——版本覆寫的作用域不該擴散到其他武器。

    截圖廊零位元組差異(戰鬥數值改變,但廊裡沒有進行中的戰鬥)。

124. **手冊有 18 把武器,remake 只做了 10 把**(2026-08-08)。

    第 123 項把武器傷害換成手冊真值時,順手記了一句「手冊還列了 remake 沒有的武器
    (離子脈衝砲、引力波束、干擾者、重錘裝置、粒子束),擴充元件表不在這一輪」。
    這一輪就是那一輪。

    加上飛彈那一頁(p.125)的脈衝飛彈、氙素飛彈、質子魚雷,一共**八項**。

    | 武器 | 手冊傷害 | 佔格 | 研究主題(執行檔給的)|
    |---|---|---|---|
    | 離子脈衝砲 | 2-10 | 30 | 離子分裂 |
    | 引力波束 | 3-15 | 15 | 人造重力 |
    | 質子魚雷 | 25 | 20 | 超維分裂 |
    | 脈衝飛彈 | 20 | 10* | 分子壓縮 |
    | 氙素飛彈 | 30 | 10* | 分子操控 |
    | 干擾者 | 40 | 20 | 多維物理 |
    | 粒子束 | 10-30 | 15 | 氙素科技 |
    | 重錘裝置 | 100(必中)| 50 | 超維物理 |

    (* 飛彈佔格依彈架 x2/x5/x10/x15/x20 遞增,remake 未實作彈架選擇,取最小彈架——
    與既有的核飛彈/麥克萊特飛彈同一個既有慣例,不是這一輪新引入的估計。)

    ### 研究主題不是猜的

    每一項的研究主題都取自 **`gamedata.OrigTechTopic`**——第 108 項從執行檔挖出來、
    211/212 對得上的那張表。不是照科技名字去科技樹裡找一個看起來合理的位置。

    有一條測試逐項核對「元件表寫的主題 == 執行檔給的主題」,所以將來手動改錯會紅。

    ### 飛彈的分類有兩個獨立來源同意

    `weaponKindByName` 決定走飛彈解算還是光束解算。新增的三項判定不是靠名字裡有沒有
    「飛彈」兩個字——**執行檔的 category 表把它們全歸在 category 21(飛彈/魚雷)**
    (第 111 項解出來的 enum 語意),而手冊 p.125 的 MISSILE 表列的正好是同一批。
    兩個獨立來源同意。測試把這個一致性釘住:category 21 ⟺ `WeaponKindMissile`。

    ### 一個容易漏的地方

    新增武器要同時進**三張表**(元件表、傷害表、佔格表)。漏掉佔格表最危險——
    `WeaponSpaceByName` 查不到會回 0,而 0 的意思是「不佔空間」,設計驗證會**靜默放行**
    一艘塞了十把重錘裝置的護衛艦。`TestEveryWeaponHasManualDamageSpaceAndTopic`
    是為此而寫的完整性閘門。

    ### 誠實留白

    - **Cost 是 remake 的尺度。** 手冊的 Cost 欄與 remake 的生產成本不是同一個單位
      (見第 123 項),新項依手冊成本的**相對名次**插在既有鄰居之間——那是選擇不是真值。
    - **重錘裝置的「always hits」沒接。** 手冊 specials 欄寫了,但 remake 的命中判定
      沒有「必中」這個旗標;順帶一提,執行檔把它歸在 category 38(特殊武器裝置)而不是
      光束——**手冊與執行檔在這一項上分類不同**,如實記錄,兩邊都沒有蓋掉。
    - 手冊 p.126-127 的炸彈與特殊武器(恆星轉換器、黑洞產生器…)仍未進元件表。

    ### 截圖廊

    `25_shipdesign.png` 差 **23 個像素**,x 389-393 / y 193-201——「武器 3/11」變成
    「武器 3/19」的那一個數字。逐像素量過範圍、把新舊裁圖並排看過才重錄。
    其餘 33 張零位元組差異。

125. **反飛彈火箭:攔截公式抄完了,但沒有一艘船裝得上**(2026-08-08)。

    `ResolveMissileShot` 的第一個參數是 `hasAMR`。兩個呼叫端都寫死 `false`,理由是:

    > 現行 remake 的 SpecialOptions 尚未提供「反飛彈火箭」這個可造艦元件,呼叫端目前
    > 一律傳 hasAMR=false(TODO:待新增該元件後,依目標艦是否裝載決定,**不在此臆造裝載狀態**)。

    **那句話是對的,而且處理方式也是對的**——沒有元件就不該假裝有。這一項不是訂正它,
    是把它等的那個元件補上。

    ### 攔截那一整套早就在了

    `gamedata/missile.go` 有完整的 AMR 模型:最大射程 15 格、距離換算成 Range 索引、
    逐索引的命中率表。`ResolveMissileShot` 也已經接好分支(`hasAMR` 為真且在射程內 →
    擲一顆獨立的骰子判攔截)。**唯一缺的是「誰身上有這東西」。**

    這與第 119 項的飛彈閃避是同一條線的兩端:那一項補的是**閃避**(艦員/舵手),
    這一項補的是**攔截**(元件)。手冊 p.123/p.125 把它們寫成飛彈防禦的兩個獨立機制。

    ### 元件的三個欄位分別從哪來

    | 欄位 | 來源 |
    |---|---|
    | 研究主題 | **執行檔**(`OrigTechTopic` → 進階工程),不是照名字猜 |
    | 分類佐證 | **執行檔** category 28(反飛彈/干擾),與手冊 p.127 的分類一致 |
    | Value = 0 | AMR 不加攻防——它的效果是攔截,加一個攻擊值等於偷偷讓它變成武器 |
    | Cost = 70 | **remake 的尺度**(見第 123/124 項),不是手冊值 |

    ### 兩條測試,一條正一條反

    - **正**:同一組種子跑 300 次,裝了 AMR 的一方被飛彈打中的次數應下降。
    - **反**:AMR **不該**擋光束——手冊寫的是「destroys incoming missile」。
      少了這條,一個「hasAMR 就整體減傷」的實作也會讓正向那條通過。

    ### 截圖廊

    `25_shipdesign.png` 差 **3 個像素**(x 533-537 / y 197-199):
    「特殊 1/8」變成「特殊 1/9」。裁圖看過確認是那個計數器才重錄。

    ### 誠實留白

    手冊 p.126-127 還有炸彈(核彈/融合彈/反物質彈/中子彈)與其他特殊武器
    (陀螺去穩器、牽引光束、脈衝星、電漿網、停滯力場、恆星轉換器、空間壓縮器、
    黑洞產生器)未進元件表。炸彈那一批需要先有「**只能打行星、打不到船**」這個武器種類
    ——現在的 `weaponKindByName` 預設回光束,直接加會讓一艘掛核彈的船在艦隊戰裡當光束艦用。
    **那是新機制不是新資料**,不在這一輪。

126. **炸彈:一個「打不到船」的武器種類**(2026-08-08)。

    第 125 項把炸彈擋在門外,理由寫得很明確:

    > 炸彈那一批需要先有「**只能打行星、打不到船**」這個武器種類——現在的
    > `weaponKindByName` 預設回光束,直接加會讓一艘掛核彈的船在艦隊戰裡當光束艦用。
    > **那是新機制不是新資料**,不在這一輪。

    這一輪做那個機制。

    ### 手冊把話說死了

    > Bombs installed in a ship are **only useful against planetary targets**
    > (though the Bomber special weapon does use bombs against enemy ships).

    所以炸彈不是「傷害比較低的光束」,是**在艦隊戰裡完全沒有這一發**。
    這個差別很重要:核彈的傷害 3-12 比同期光束(雷射 1-4、核融合光束 2-6)還高,
    落到 beam 分支等於白送一個強化。

    ### 「完全沒有這一發」不等於「0 傷害的命中」

    `battleVolley` 那一支對炸彈直接 `continue`,**連骰子都不擲**。

    擲了會怎樣?骰子是一條共用的隨機序列——多消耗兩顆,後面每一發的結果都位移,
    決定性測試會無故變動而且看不出原因。`TestBombDoesNotConsumeRandomness` 用
    「齊射後亂數流的位置」直接驗這件事,不是靠傷害為 0 間接推。

    ### 分類同樣有兩個獨立來源

    執行檔的 category 表把四種炸彈全歸在 **category 19(炸彈)**(第 111 項解出的 enum),
    而手冊 p.126 的 BOMB 表列的正好是同一批。

    ### 兩個猜錯的主題被測試接住了

    第 124 項加的完整性閘門(元件表的主題 == 執行檔給的主題)這一輪真的抓到東西:
    融合彈我寫成「進階化學」(執行檔是**進階融合**)、中子彈我寫成「板塊工程」
    (執行檔是**交相分裂**)。兩個都是「看名字很合理」的錯——正是那條測試存在的理由。

    ### 生物武器**刻意**不進元件表

    手冊同一張 BOMB 表還有死亡孢子(10%)與生物滅絕者(20%),但它們給的是**殺人口的
    百分比**不是傷害,而且**第 111 項已經用另一條路徑接好了**(科技擁有 → 轟炸時擲骰殺人口)。
    加進元件表會讓同一條規則生效兩次。有一條測試專門守住「它們不在元件表裡」。

    ### 截圖廊

    `25_shipdesign.png` 差 49 個像素:「武器 3/19」→「武器 4/23」——總數 +4(四種炸彈),
    已解鎖 +1(示範玩家本來就有核分裂那個主題)。裁圖看過才重錄。

127. **球形武器:一整條戰鬥解算分支,零武器掛載**(2026-08-08)。

    `weapon_kind.go` 的檔頭寫著:

    > 因此 spherical 分支目前**沒有任何武器掛載**,`ResolveSphericalShot` 只提供已測試的
    > 解算函式待未來新增球形武器元件時串接。

    `ResolveSphericalShot` 的函式註解也重複了一次同樣的話。**整條解算路徑是死碼**
    ——連 `battleVolley` 的 `case WeaponKindSpherical:` 那一段都從來沒有執行過。

    手冊 p.126「Notes on Spherical Damage」明列的球形武器是四項:
    Pulsar、Plasma Flux、Spatial Compressor、Engine Explosion。其中

    - **電漿通量**是海鰻怪獸專屬,不是可造艦元件;
    - **引擎爆炸**是船被打爆時的事件,不是武器;

    所以掛得上的是**脈衝星**與**空間壓縮器**兩項——而它們的數值就在 p.127 那張表上。

    ### 「per size class of target」需要目標的艦體等級

    手冊給脈衝星的傷害是「1-24 **per size class of target**」、空間壓縮器是
    「4-32 structural hits」。前者要知道**被打的那艘船有多大**——而 `combatant` 先前
    沒有艦體等級這個欄位(它只有 `shipStrength` 換算出來的戰力點)。補上 `sizeClass`,
    直接複用既有的 `shipClassFromName`。

    ⚠ **「級數」取 index+1 是讀法,不是手冊字面。** 手冊沒有列出級數的數字,只給了艦體
    名稱的順序;取 index 會讓護衛艦那一級乘 0(打護衛艦零傷害),那顯然不是規則。
    有一條測試專門守住「最小艦體也吃得到傷害」。

    ### 只有一項豁免護盾與裝甲

    手冊只在**空間壓縮器**那一格寫了「does all damage to **structure only**, ignoring
    shields and armor」;脈衝星沒有那一句。所以豁免是**逐武器**的,不是整個球形類別的屬性
    ——`ResolveSphericalShot` 早就把它做成參數(`bypassShieldAndArmor`),先前沒有呼叫端
    去決定該傳什麼。

    測試一正一反:壓縮器對厚甲目標的總結構傷應高於脈衝星;而脈衝星/雷射/核彈/死光
    都不該豁免。

    ### 第 124 項的閘門第二次抓到東西

    加完元件跑測試,`TestEveryWeaponHasManualDamageSpaceAndTopic` 立刻紅:
    「脈衝星 沒有手冊傷害值 / 沒有佔格值」。兩張表都忘了加。

    這是那條測試連續第二輪抓到真的疏漏(上一輪是第 126 項的兩個錯主題)。
    **新增武器要同時進三張表**這件事,靠註解提醒無效,靠測試才擋得住。

    ### 截圖廊

    `25_shipdesign.png` 差 10 個像素:「武器 4/23」→「武器 4/25」——總數 +2,
    已解鎖不變(示範玩家還沒有翹曲力場/氙素科技)。

128. **p.127 剩下的特殊武器:全部卡在機制,不是卡在資料**(2026-08-08)。

    第 124–127 項把手冊的武器表逐頁補進來:光束(p.124)、飛彈(p.125)、
    炸彈(p.126)、球形(p.126 清單 + p.127 數值)。剩下 p.127 那張特殊武器表。

    逐項查過,**沒有一項是「資料查不到」**——數值全都在表上。卡的是別的:

    | 武器 | 手冊效果 | remake 缺什麼 |
    |---|---|---|
    | 牽引光束 | reduces target speed | **戰鬥速度模型**。`gamedata.CombatSpeed` 移植好了但零呼叫端,戰術戰鬥沒有逐艦速度 |
    | 停滯力場 | places target in suspended animation | **「這一輪不能動」的狀態**,同上 |
    | 黑洞產生器 | immobilizes and destroys target | 同上,外加「即毀」不是傷害 |
    | 電漿網 | 5-25 **to each side** of target | **護盾分面(facing)模型**。`DamageShieldCapacity` 也是為了它寫的,同樣零呼叫端 |
    | 恆星轉換器 | 400 **to each side** | 同上。⚠ **行星版**已經接了(`StellarConverterRetaliationAttack`,轟炸反擊),艦載版沒有 |
    | 陀螺去穩器 | 1-4 per size class | 資料齊、艦體級數也有了(第 127 項)——但它**不在手冊的球形清單裡**,而 remake 的光束路徑沒有「per size class」這個乘數 |

    ### 為什麼停在這裡

    前面四項要的是**新的戰鬥狀態**(速度、行動禁止、護盾分面)。硬把它們當一般傷害武器
    加進去,玩家會看到一把叫「牽引光束」的東西在扣血——**名字對、行為錯**,比沒有它更糟。

    最後一項(陀螺去穩器)最接近可做:只差一個「這把武器的傷害要乘目標級數」的旗標。
    但那是把第 127 項為球形武器寫的乘法搬成通用屬性,**動的是分類模型不是加一列資料**
    ——與前面幾項同一個判斷:先有機制,再放資料。

    ### 這一輪同時修的一件事:活來源表的形狀

    `CLAUDE.md` 指名「唯一活來源」是本文件開頭那張「真正還缺的」表,而它 2026-08-07 起
    寫著「只剩一列」。那句話**準,但也誤導**:表上列的是**子系統級**的洞,而第 111–127 項
    那 17 項沒有一項是子系統級的——它們是「做了但不忠實」,**在那張表的形狀裡根本不會出現**。

    已在表上補一列「手冊數值忠實度」,並把這一天用過的四種問法與各自的殘量寫進去。
    ①(零覆蓋函式)②(零消費常數)已挖到見底;③(被餵固定值的參數)④(手冊沒抄的資料)
    還有。**③ 特別值得做:它抓的是「接了但等於沒接」,那種洞在測試與截圖上都看不出來。**

129. **種族只有 7 個自編數字,原版有 31 格特性——一手表挖出來全部換掉**(2026-08-08)。

    第 128 項在活來源表上補了「手冊數值忠實度」那一列,並列出四種還有殘量的問法。
    這一項走的是 **③ 被餵固定值的參數**。掃描結果指向同一個根:

        GroundMarineHitsToKill(highGRace, …)  highGRace 恆 false(5 個呼叫端)
        GroundTankHitsToKill(highGRace)        恆 false
        TradeGoodsIncome(…, fantasticTrader)   恆 false
        AIPlanetValueInput.RaceLowG/RaceHeavyG 零呼叫端
        GroundSubterraneanBonus                零呼叫端
        crewXPThresholdsWarlord                零呼叫端

    全部是**種族特性**。而 `engine.PlayerState` 只有一個 `TolerantRace bool`
    ——`shell.Race` 把整個種族壓成七個自編數字(工業/科研/農業/成長/起始BC/每人BC/戰鬥%)。

    ### 原版的形狀

    `internal/save/entities.go` 早就有 `Traits [31]int8`,讀的是玩家結構 `+0x89F`。
    遊戲層從來沒用過它。反組譯 `sub_12983`(開局玩家設定)給了完整鏈路:

        push    1Fh                        ; 31 格
        mov     edx, 7                     ; RACESTUF.LBX 的 asset 7
        mov     edx, dword_19B7DC[ecx*4]   ; 第 ecx 族的陣列
        add     eax, 89Fh                  ; → player+0x89F
        call    sub_12779E                 ; memcpy

    然後一個轉換迴圈把**選項等級**換成真值:`byte_17D1F9[traitIdx*3 + level]`,
    而且 `cmp dx, 0Ah; jl` ——**只跑 1..9**。特性 0(政體)與 10..30(布林)存原值。
    這一格很重要:照 `[idx*3+level]` 一路換到 30,會把「水棲=1」讀成 20。

    ### 三個獨立來源交叉驗證

    | 來源 | 內容 |
    |---|---|
    | `RACESTUF.LBX` asset 7 | 4 位元組表頭(13 筆 / 每筆 31 格)+ 13×31 選項等級 |
    | 執行檔 `byte_17D1F9`(檔案位移 0x1FB88D) | 10 列 × 3 級的換算表 |
    | `SAVE10.GAM` | 五族(Alkari/Klackon/Mrrshan/Sakkra/Trilarian)展開後的 `Traits[31]` |

    前兩者推出來的結果與第三者**逐格相同**。手冊行文再獨立對第四次:布拉西
    「+20 to Ship Attack」「+10 bonus in ground combat」、埃雷里安「+25 defensive and
    +20 offensive」、達洛克「+20 more likely」——全部對上。

    ⚠ 中途踩到一次:`conv` 只切了 30 位元組,而特性 9 的 3 檔要讀到第 30 格,
    達洛克的間諜 +20 被吃掉。**是手冊那句「+20 more likely」把它抓出來的**
    ——四個來源不是為了好看,是真的會互相攔截。

    ### 換算表(執行檔真值)

    | 特性 | 1 檔 | 2 檔 | 3 檔 |
    |---|---|---|---|
    | 人口成長 | −50 | +50 | +100 |
    | 農業 | −1 | +2 | +4 |
    | 工業 / 科研 / 金錢 | −1 | +1 | +2 |
    | 艦艇防禦 | −20 | +25 | +50 |
    | 艦艇攻擊 | −20 | +20 | +50 |
    | 地面戰 / 間諜 | −10 | +10 | +20 |

    ### 換掉之後,八族的數字是錯的

        克拉肯   工業+2         → 農業+2、工業+1(拿錯欄位)
        阿爾卡里 通用戰鬥+15    → 艦艇**防禦**+50(它沒有攻擊加成)
        姆瑞森   通用戰鬥+25    → 艦艇**攻擊**+50
        薩克拉   成長+30、食物+1 → 成長+100、農業+2、間諜−10
        崔拉里安 食物+1、成長+10 → 兩者皆 0(它的特性是水棲與跨維度)
        埃雷里安 科研+1         → 0(它的加成在艦艇攻防:防+25 攻+20)
        矽基     工業+1、成長−20 → 工業 0、成長−50
        達洛克   全 0           → 間諜+20

    「通用戰鬥%」這個欄位本身就是問題:原版把攻擊與防禦分成**兩個獨立特性**,
    壓成一個之後,阿爾卡里會拿到它根本沒有的攻擊加成,而姆瑞森的防禦憑空冒出來。
    已拆成 `CombatPct`(艦攻)與 `ShipDefPct`,並在戰列上對稱套用。

    ### 順手修掉一個真 bug:諾蘭姆的低重力被扣兩次

    `GroundRaceCombatBonus(GroundRaceGnolam)` 回 −10,而下一行的 `GroundApplyLowGPenalty`
    又扣一次 −10。反組譯只寫一次(`mov byte ptr [ecx+0Dh], 0F6h`),而諾蘭姆的
    `TRAIT_GROUND_COMBAT` 本來就是 **0**——那 −10 完全來自 `TRAIT_LOW_G`。
    改由特性表驅動之後重複自然消失,`GroundRace` 列舉與 `GroundRaceCombatBonus` 一併刪除
    (它們的存在理由是「只有手冊寫了數字的那兩族查得到」,而一手表出現後這個理由消失了)。

    ### 一併解掉的擋門理由

    `playerMarineForce` 的檔頭寫著:

    > Subterranean 加成、High-G hits-to-kill 未套用:…**13 個標準種族也沒有一個具備
    > Subterranean/High-G**,故無從套用,誠實留白而非臆測。

    **那句話可證為假。** 薩克拉有地底、布拉西有高重力。當時查不到不是因為原版沒有,
    是因為 remake 沒有特性模型。兩者都已接上(地底只在守自家殖民地時生效,
    與手冊 + 反組譯的呼叫端旗標一致)。諜報加成(達洛克 +20、薩克拉 −10)也接進攻守兩側。

    ### 這一項最該記住的一件事

    `docs/knowledge-base/manual-cht/01-races.md` **一年前就把正確數字寫進表裡了**
    ——薩克拉 +100%、姆瑞森 +50、阿爾卡里防禦 +50 全都在,而且第 211 行還明寫著
    「remake 的 `Races` 表採**概略調校值**…非手冊精確數字」。**答案一直在自己的 docs 裡,
    只是沒有人回頭查。** 這正是 `rules/00-rules-index.md` 那條「要說解不出來之前,
    先 grep 自己的 docs」講的情況,只是這次的形式不是「解不出來」而是「沒想到要查」。

    (反過來這一輪也訂正了那份文件兩處:薩克拉與克拉肯的農業是 +2 不是 +1
    ——`Traits[2]=2`,原版存檔直接寫著。一手資料贏手冊行文。)

    ### 誠實留白

      - **自訂種族拿不到布林特性。** `RaceOrigIdx` 記 −1,`OrigRaceHasTrait` 越界回 false。
        點數畫面目前只記錄數值型加成,寧可少給也不亂給。
      - **AI 對手不吃種族特性。** `AIOpponent` 沒有種族欄位——那一整層還不存在,不是漏接。
      - **`TradeGoodsIncome` 的 fantasticTrader 仍是 false。** 諾蘭姆有神級商人特性,
        查得到了,但那條路徑的呼叫端是帝國層收入,接線要動 income 管線,留給下一項。
      - **特性 31(貧瘠母星)在列舉裡有,陣列只有 0..30。** 原版放不下它,不替它捏一格。

130. **布林特性接線:五條寫好的規則終於有呼叫端**(2026-08-08)。

    第 129 項挖出了 13 族 × 31 格的特性表,把**數值型**那 9 格接了進去;
    布林那 21 格當時只做到「查得到」。這一項接「查得到之後怎麼用」。

    五處的規則**全部早就寫好了**,缺的一直只是「這一族到底有沒有」:

    | 特性 | 種族 | 先前的狀態 |
    |---|---|---|
    | 統帥 | 姆瑞森 | `RaceWarlord` 欄位在、5 個呼叫端都讀它,**沒有人寫入** |
    | 惹人厭 | 矽基 | `AssimilationTurns` 的 repulsive 分支同上 |
    | 寬容 | 矽基 | `engine.ColonyState.TolerantRace` **從來沒有寫入端** |
    | 神級商人 | 諾蘭姆 | `TradeGoodsIncome` / `IncomeFoodSurplusRevenue` 硬傳 `false` |
    | 魅力非凡 | 人類 | `raceDiploBonusPct` 硬比 `RaceIndex == 0` |

    前四項的擋門理由是同一句話,`session.go` 寫得很清楚:

    > **目前沒有任何內建種族會設它**——十三經典種族的特質表(Races)還沒有特質欄位

    那句話在寫下的當天是對的。第 129 項之後就不對了,而**沒有任何東西會提醒它過期**
    ——這是第 117/118/122/123 項同一個形狀,這一輪的第五次。

    ### 順手修的第五項:硬比索引

    `raceDiploBonusPct` 原本是 `if s.RaceIndex == 0 { // 人類 }`。這在兩個方向上都會錯:
    `Races` 重排一次就指到別族;而自訂種族選了魅力非凡也永遠拿不到。改成查特性。

    同時發現 `ApplyCustomRaceBonuses` **沒有把 RaceIndex 標成 −1**,於是自訂種族會停在
    預設的 0(人類)——布林特性由 RaceIndex 查,而 0 是一個合法索引,自訂種族會憑空
    拿到「魅力非凡」。已補。

    ### 走錯一次:把衍生狀態存起來

    第一版把種族編號與五個布林旗標當成 `GameSession` 欄位存進存檔。三個測試同時紅:

        TestSaveLoadRoundTripKeepsHash      存讀往返的狀態指紋對不上
        TestSeatRoundTripKeepsEveryField    熱座席位換人少抄了兩個欄位
        (另一條是前者的連鎖)

    根因不是漏抄,是**設計錯了**:那些值全部可以從 `RaceIndex` 算出來,而 `RaceIndex`
    本來就存。存起來就多了一個會不同步的真相來源,而且舊存檔沒有那個欄位會解出 0
    (= 阿爾卡里),整個查錯族。改成方法(`raceOrigIdx()` / `RaceWarlord()` / …),
    三個問題一起消失,存檔格式也不必動。

    > **測試紅的時候先問「這個狀態該不該存在」,再問「哪裡漏抄了」。**
    > 這一次如果順著錯誤訊息去補抄欄位,會補出一個能過測試但舊存檔仍然壞掉的版本。

    只有**跨層**的兩個欄位需要同步(引擎不認識 shell 的種族表):
    `engine.PlayerState.FantasticTrader` 與 `engine.ColonyState.TolerantRace`,
    在 `EndTurn` 開頭與成就同步(第 118 項)並列,冪等。

    ### 誠實留白

      - **魅力非凡在同化那一側仍不生效。** 現在查得到人類有它了,但手冊只說
        「assimilate conquered colonists **easily**」,沒給數字。
        **查得到 ≠ 知道加多少**——這一項解掉的是前者。外交 +50% 那一側有數字,已生效。
      - **自訂種族的 pick 尚未寫進特性。** 點數畫面目前只記錄數值型加成,
        所以自訂種族一項布林特性都拿不到(`RaceIndex = −1`)。寧可少給也不亂給。
      - **其餘布林特性(水棲/食岩/半機械/創造力/幸運/全知/匿蹤艦/跨維度/母星品質)未接。**
        它們要的是 remake 還沒有的機制——星球適居度模型、科技樹分支、偵測模型、母星生成
        ——不是「忘了接」。這與第 128 項對 p.127 特殊武器的判斷同一條線:先有機制,再放資料。

131. **高能聚焦:規則寫好了,但那個東西**裝不上**(2026-08-08)。

    ③ 那條問法(被餵固定值的參數)清掉種族那一叢之後,剩下的最大一項是:

        DamageMountAdjustedValue(base, hvBonus, hefBonus, pdPenalty, rangePenalty)
                                             ↑ 兩個呼叫端都恆傳 0

    `gamedata.DamageMountBonusHEF = 50` 在、公式在、註解裡連手冊原文都抄了
    (`Hv+HEF: 50*(100+50+50-30)%` 這種算例都寫好了),就是沒有人傳非 0 值進去。

    ### 為什麼

    第 36 項做了武器改造(mod)系統:HV / PD / AF / CO / AP / ENV / NR / SP 全部接上了。
    HEF 沒有——因為 **HEF 在手冊裡不是武器改造,是艦載系統**:

    > High Energy Focus **(System)**

    改造走 `Ship.Mods`,系統走 `Ship.Special` 與 `SpecialOptions` 那張表,而
    `SpecialOptions` 裡沒有它。於是「玩家裝不上 → 呼叫端沒有東西可傳 → 恆傳 0」。
    這不是漏接一行,是**東西被分到錯的類別去了**,所以整條路都沒有。

    ### 手冊那三句話,三個不同的位置

    > increasing the damage each of these weapons inflicts by **50%**.
    > It **does not improve the chances of hitting** a target at a greater distance,
    > nor does it **prevent the normal drop-off of damage over range**.

    三句各自對應程式裡的一個地方:傷害走 `DamageMountAdjustedValue` 的 hefBonus、
    命中走 `CombatHitThreshold`、距離衰減走 `DamageDissipationPenalty`。
    **只接第一個,另外兩個一個都不能碰。** `hef_test.go` 逐句釘住,其中命中那一條是
    **逐骰比對** 1..100 兩邊的命中結果必須完全相同——比「找一顆會 miss 的骰」嚴,
    後者只證明某一顆沒被改變,前者證明整條門檻沒有移動。

    ### 兩條戰鬥路徑都要接

    快速結算(`combatant.hasHEF`)與格子戰術(`CombatShip.HEF`)是兩條獨立的路。
    只接一邊的話,同一艘船在兩種戰鬥裡會打出不同傷害——而那種不一致在遊玩中幾乎察覺不到,
    只會讓人覺得「這場怎麼特別難打」。測試同時檢查兩邊。

    ### 順手清掉的過期註解

    `ResolveMissileShot` 的檔頭寫著:

    > 現行 remake 的 SpecialOptions 尚未提供「反飛彈火箭」這個可造艦元件,
    > 呼叫端目前一律傳 hasAMR=false(TODO:待新增該元件後…)

    **第 125 項已經把該元件補上了**,呼叫端也早就改成依目標艦是否裝載決定。
    註解比程式碼晚了六項。這是這一輪第六次撞到同一個形狀。

    ### 誠實留白

      - **建造成本 90 是 remake 值。** 手冊行文只給效果不給成本,執行檔的元件表還沒挖到
        ——RACESTUF 那條路(第 129 項)只有種族資料。取值參考同屬中後期系統的
        硬化護盾(100)與反飛彈火箭(70),與 `SpecialOptions` 其餘元件同一種標記方式。
      - **敵方艦不裝。** `genEnemyFleet` 沒有個別元件設計資料,一律 false
        ——與 Mods / HardShield 同款既有簡化,不是這一項新增的缺口。

    截圖:25_shipdesign 的「特殊 1/9 → 1/10」(521 px)。

132. **裝甲科技倍率:一則撤回,外加重裝甲與穿甲抵銷**(2026-08-08)。

    ### 先講撤回

    這一輪稍早(第 123 項一帶)拒絕動 `armorHPByName`,理由寫得很篤定:

    > 手冊的 Armor 欄是**逐艦體**的,而裝甲科技的倍率**手冊與 openorion2 都沒有**
    > (openorion2 是個沒有戰鬥引擎的渲染殼)。

    **後半句是錯的。** 手冊 Ship 條目裡逐級寫著,只是當時沒讀到那幾頁:

    | 裝甲 | 手冊原文 | 倍率 |
    |---|---|---|
    | 鈦 | 「standard armor for FTL ships」(基準,未給倍率) | 100% |
    | 三鈦 | increases the structural integrity ... by **100%** | 200% |
    | 佐特 | increases the structural integrity ... by **300%** | 400% |
    | 中子素 | boost the structural hits ... by **500%** | 600% |
    | 精金 | increases the structural hits ... by **700%** | 800% |
    | 氙素 | Ships with this armor have **10 times** the base structure and armor points | 1000% |

    **「我找不到」與「它不存在」被寫成了同一句話**,而那句話又被後續幾項當成既有結論引用。
    這是 `rules/00-rules-index.md` 那條「要推翻既有斷言之前先找出當初支持它的證據」的反面
    ——當初根本沒有證據可找。留在 `gamedata/armor_tech.go` 檔頭,不默默改掉。

    改完之後三格差最多:佐特 35→40、中子素 55→60、**氙素 120→100**
    (它是「10 倍」不是「+1100%」,那一格最容易抄成加法)。

    ⚠ **階梯是一手的,基準單位 10 是 remake 值。** 手冊講的是 structural integrity
    ——原版沒有獨立的「裝甲池」,裝甲科技決定的是艦艇結構點數,兩池是 remake 的抽象。
    兩件事分開記,免得日後有人把整條當成原版真值引用。

    ### 順帶訂正:氙素裝甲掛錯主題

    `ArmorOptions` 裡它掛在 `TOPIC_ARTIFICIAL_LIFE`、`UnlockTech = 0`,註解標「里程碑,proxy」。
    執行檔的 tech→topic 表說它屬 **Xenon Technology**(74),而 `TECH_XENTRONIUM_ARMOR`
    一直都在列舉裡(201)。「proxy」是當時沒查,不是查了查不到。

    ### 重裝甲:與高能聚焦一模一樣的形狀

    `DamageApplyArmor(dmg, armorHP, armorPiercing, apNegated)` 的 **apNegated 從寫出來就恆傳
    false**,而它的註解連兩句手冊原文都抄好了:

    > Heavy Armor (System): ... also **negates the Armor Piercing abilities** of enemy weapons
    > Xentronium Armor: ... **Negates armor piercing effects** of enemy weapons

    氙素裝甲在 `ArmorOptions` 裡,查得到;**重裝甲不在任何清單上**——因為它是
    「Heavy Armor **(System)**」,系統要進 `SpecialOptions`,而它不在。與第 131 項的
    高能聚焦同一個形狀:**東西被分到錯的類別,於是整條路都不存在。**

    補上之後兩個效果都有落點:

      - 裝甲耐受量 **×3**(手冊「triples the amount of damage the ship's armor can sustain」)
      - **抵銷穿甲**,與氙素裝甲取聯集(`shipNegatesArmorPiercing` 把兩條路併起來)

    端到端測試釘住:AP 改造打在會抵銷的目標上,傷害必須**全部由裝甲吸收**、結構吃 0。
    兩條戰鬥路徑(快速結算 / 格子戰術)都帶旗標——只接一邊會讓同一艘船在兩種戰鬥裡
    表現不同,而那種不一致在遊玩中幾乎察覺不到。

    ### 誠實留白

      - **建造成本 110 是 remake 值**,理由同高能聚焦(執行檔的元件表還沒挖到)。
      - **「三倍」乘的是裝甲那一池不是結構。** 手冊明說 "the ship's **armor** can sustain
        before damage gets through to the structure",所以只動 `armorHPByName` 的結果。
      - **兩池抽象本身沒有改。** 把裝甲折回結構點數是另一個層級的重構,不在這一項裡。

    截圖:25_shipdesign 的「特殊 1/10 → 1/11」(23 px)。

133. **手冊元件完整性盤點 + 飛彈防禦系統家族**(2026-08-08)。

    第 131(高能聚焦)與第 132(重裝甲)是**同一個形狀撞到兩次**:手冊裡標
    `(System)` 的東西沒進 `SpecialOptions`,於是那條規則的參數只能恆傳 0。
    撞第二次之後改問法——**不要一個一個等它自己冒出來**,把手冊所有元件條目
    對 remake 的四張元件表做一次完整性盤點。

    ### 盤點方法

    手冊條目標題的格式是 `名稱 (分類)`,分類有 System / Ship / Achievement / Weapon。
    抽出 System + Ship 共 88 個名稱,用 `gamedata.TechnologyName` 反查成 `Technology`,
    再看它有沒有出現在 `WeaponOptions / ArmorOptions / ShieldOptions / SpecialOptions`
    任一張表的 `UnlockTech`。**全自動,不靠記憶。**

    結果:**47 個手冊元件在 remake 裡裝不上**,另有 10 個科技名對不上
    (`Phasors` vs `Phasor`、`Class V Shields` vs `Class V Shield` 這類單複數差異)。

    ### 47 個的分桶

    | 桶 | 數量 | 說明 |
    |---|---|---|
    | **不需要元件槽** | 15 | 艦體等級(泰坦/末日之星)、殖民船/前哨船/運輸艦/貨船(建造項目)、各級引擎(星圖速度模型)、燃料槽(航程模型)、各級電腦(戰鬥電腦 proxy)——remake 用別的模型承接,不是缺口 |
    | **這一項接上** | 8 | 見下 |
    | **卡機制** | 4 | 陀螺去穩器 / 電漿網 / 停滯力場 / 牽引光束——第 128 項已逐項判定,缺的是戰鬥速度與護盾分面模型 |
    | **仍缺、可做** | 20 | 戰鬥艙、強化船體、保安站、戰鬥掃描器、匿蹤力場、相位匿蹤、時間扭曲加速器、阿基里斯/測距瞄準器、超載電容、快速飛彈架、轟炸機庫、多相護盾、能量吸收器、傳送器…… |

    最後那一桶是下一輪的料。**寫在這裡是為了讓它可見**——先前它不在任何清單上,
    所以既不會被做也不會被承認沒做。

    ### 這一項接的 8 個:飛彈防禦系統家族

    `gamedata/missile.go` 早就把手冊 p.123 整段搬完了,每一個都附原文與精確數字:

        MissileJammerECM            70    電子干擾器
        MissileJammerMultiWave     100    多波電子干擾器
        MissileJammerWideAreaSelf  130    廣域干擾器
        MissileInertialStabilizer   25    慣性穩定器
        MissileInertialNullifier    50    慣性抵消器
        MissileLightningFieldDestroyChance  50   閃電場
        MissileDisplacementDeviceMissChance 30   位移裝置

    七個常數,**生產端全部是 0**。`battleVolley` 的註解把理由寫得很清楚:

    > 防守方的飛彈閃避先前恆傳 0(「艦艇設計/軍官系統尚未提供這些元件」)。
    > 那句話對 ECM 干擾器/慣性穩定器**仍成立**

    加上部隊艙(`GroundTroopPodsMultiplier = 2`,手冊 p.79「doubling the number of
    Marines on board a ship」,同樣零生產端)共 8 個。**主題全部取自執行檔的
    tech→topic 表**,不是猜的——其中慣性抵消器與位移裝置同屬 Transwarp Fields(71),
    是三選一裡的兩個選項,研究時只能挑一個。

    ### 兩個判定順序,手冊自己給了

    閃電場排在**最前面**、位移裝置排在**最後面**,不是隨手安排的:

    > 閃電場:在 MIRV 飛彈分裂彈頭「**之前**」判定 → 它擋的是整枚飛彈
    > 位移裝置(與匿蹤同組):在分裂「**之後**」判定 → 它躲的是彈頭

    順序寫錯的話,一枚被閃電場摧毀的飛彈還會先跑一次閃避判定,多消耗一顆骰。

    ### 裝了才擲骰

    兩個新防禦各需要一顆骰。**沒裝就完全不動 RNG**——這是第 126 項炸彈分支學到的:
    亂數流位移一格,既有存檔與探針的所有戰鬥結果就全部改變。
    測試用 RNG 流位置驗,不是用「傷害是 0」推論。

    ### 兩條戰鬥路徑一起接

    格子戰術那一側的註解也還寫著「hasAMR/evasion 加成現行皆無對應可造艦元件,
    保守傳 0/false」——第 125 項就補了反飛彈火箭。**這一項把兩條路徑一次接齊**,
    免得再出現「同一艘船在快速結算與格子戰術裡表現不同」這種很難察覺的不一致。

    ### 誠實留白

      - **廣域干擾器只給自己那一格(130)。** 手冊另有「對艦隊其餘船艦 +70,且不與其他
        jammer 疊加」——remake 的戰列是逐艦獨立的,沒有艦隊層加成這個概念。
      - **一艘船只有一個 Special 槽。** 原版可以同時裝干擾器與慣性穩定器,remake 不行
        ——既有設計限制(見 `Ship.Special`),不是這一項引入的。
      - **匿蹤裝置的 50% 未命中沒接。** 手冊寫明「僅在裝置**啟動**時」,而 remake 沒有
        啟動/未啟動狀態。**查得到 ≠ 用得上。**

    截圖:25_shipdesign 的「特殊 1/11 → 1/19」(23 px)。

134. **第 133 項那一桶的第一批,外加修掉第 133 項自己的一個缺失**(2026-08-08)。

    ### 先講缺失

    第 133 項接慣性穩定器時,只給了 **+25 飛彈閃避**。手冊那一條的全文是:

    > The result is a **+50 addition to the ship's beam defense**, **+25 to the ship's
    > missile evasion**, and a halving of the movement cost for turning the ship in place.

    **三個效果,接了一個。**

    原因值得記:第 133 項是從 `gamedata/missile.go` 那一側往回找元件的——那個檔案只收
    飛彈相關的常數,`MissileInertialStabilizer = 25` 在那個脈絡裡看起來就是它的全部效果。
    **從單一檔案回推元件效果會漏東西**,正解是回去讀手冊那一條的全文。

    (而 `gamedata.BeamDefense`(移植自 openorion2 `ShipDesign::beamDefense`)其實早就有
    `inertialStabilizer` 參數並且加 50——又一個「規則在、呼叫端不在」,只是這次它躲在
    另一個檔案裡,從飛彈那一側完全看不到。)

    ### 接上的四個(第 133 項「仍缺、可做」那 20 個裡的第一批)

    選擇標準是**同時**滿足兩個條件,不滿足的不接:

      ① 手冊給了確切數字(不是「considerably」「vastly」這種形容詞)
      ② remake 已經有承接它的位置(不必先造一個新機制)

    | 元件 | 手冊原文 | 落點 |
    |---|---|---|
    | 戰鬥掃描器 | increases the ship's chance to hit with beam weapons by **50** | `combatant.atk` |
    | 強化船體 | **triples** the amount of structural damage a ship can sustain | `combatant.hp` |
    | 多相護盾 | increasing the maximum amount of damage that they can absorb by **50%** | `combatant.shield` |
    | 偵察實驗室 | Frigate = **1**, Destroyer = **2**, Cruiser = **4**, Battleship = **8**, Titan = **16**, Doom Star = **32** | 研究階段 |

    偵察實驗室那一條罕見地把整張表都列出來了。艦隊研究**併進** `TotalResearch` 而不是
    另開一條:研究階段只有一個投入口,分開會讓「研究完成」的判定要看兩個地方。

    ### 沒接的十幾個,與各自的理由

    寫在 `internal/shell/ship_systems.go` 檔尾,逐項列。摘要:

      - **機制不存在**:保安站(登艦戰,第 119 項)、增強引擎與時間扭曲加速器(戰鬥速度/
        回合結構,第 128 項)、匿蹤力場與相位匿蹤(可見性狀態)、超載電容與快速飛彈架
        (回合內射擊次數)、能量吸收器(儲能狀態)、傳送器(護盾分面)
      - **加了不影響任何東西**:戰鬥艙(「add equipment space」——remake 沒有逐元件佔格)
      - **會讓兩條戰鬥路徑不一致**:測距瞄準器(把距離縮成 1/3,而快速結算固定 range=2,
        只有格子戰術有真距離)——那正是第 131–133 項一直在防的事
      - **要先重構**:結構分析儀與阿基里斯瞄準器都要動 `ResolveShotWithMods` 的傷害鏈,
        而該函式的參數已經排到第 11 個。再加下去該先把攻方/守方系統各收成一個結構
        ——**那是重構,不該夾在資料項裡做**

    ### 誠實留白

      - **強化船體只接了結構那一半。** 手冊同一句還說「tripling the amount of damage
        required to destroy the drive system」——remake 沒有逐系統損毀(船不是完好就是被擊沉
        加一個累積損傷值)。
      - **偵察實驗室的弱點分析沒接。** 手冊說它讓艦隊「analyze the opponent's biology or
        structure and seek out weaknesses」——remake 的怪物戰鬥沒有「弱點」這個概念。
      - **`PlayerState.FleetResearch` 是每回合重算的衍生值。** 它會進存檔(整個 PlayerState
        都會),但在 `EndTurn` 開頭就被覆寫,舊存檔的陳舊值影響不到任何判定
        ——與第 130 項那個錯誤不同:那裡存的是**沒有人會重算**的種族編號。

    截圖:25_shipdesign 的「特殊 1/19 → 1/23」(26 px)、30_netwait 狀態指紋(101 px)。

135. **傷害鏈收成具名結構,順手解掉卡在它後面的兩個系統**(2026-08-08)。

    第 134 項把結構分析儀與阿基里斯瞄準器列在「要先重構」那一格,理由是:

    > 該函式的參數已經排到第 11 個。再加下去該先把攻方/守方系統各收成一個結構
    > ——**那是重構,不該夾在資料項裡做**

    這一項就是那個重構。

    ### 為什麼值得做

    `ResolveShotWithMods` 的呼叫端長這樣:

        ResolveShotWithMods(net, wmin, wmax, 2, d.shield, d.armor, roll,
            false, weaponModCodes(...), hefBonusFor(...), d.apNegated)

    `false, ..., 0, false` 這種尾巴**沒有人看得出哪個是什麼**,而且第 131/132/133 項
    每加一個系統就要回頭改每一個呼叫端與每一個測試。收成結構之後:

      - 呼叫端讀得出意思(具名欄位)
      - 加一個系統只動一個地方
      - **攻方/守方分開**,不會再出現「把守方的旗標填進攻方那一格」這種安靜的錯

    舊入口 `ResolveShot` / `ResolveShotWithMods` 保留為薄包裝,既有呼叫端與測試一行都沒改。
    **第一條測試就是驗這件事**:七組輸入,舊入口與新入口的結果必須逐欄位相同
    ——如果包裝寫歪了,後面所有斷言都在測錯的東西。

    ### 接上的兩個

    | 元件 | 手冊原文 | 位置 |
    |---|---|---|
    | 結構分析儀 | the damage done by beam weapons that **penetrate an enemy ship's shields** is **doubled** | 扣完護盾**之後**加倍 |
    | 阿基里斯瞄準器 | all beam weapons **ignore the target's armor completely** | 與 AP 改造同一個開關 |

    結構分析儀的**順序有意義**:「penetrate the shields」既是條件也是時機,所以加倍發生在
    扣完護盾之後。寫反的話護盾也會跟著被加倍的傷害穿透,那是另一種規則
    ——而且**在無護盾的目標身上兩種寫法結果相同**,所以測試特意用有護盾的目標。

    ### 一個明說的讀法

    阿基里斯瞄準器會不會被重裝甲/氙素裝甲抵銷,**手冊沒有明說**。那兩條寫的是
    「negates the Armor **Piercing abilities** of enemy weapons」,字面涵蓋「無視裝甲」
    這個能力,所以採「會被抵銷」的讀法並用測試釘住。標出來是為了讓日後改讀法的人
    知道自己在改的是一個讀法,不是一個 bug。

    ### 誠實留白

      - **飛彈與球形武器沒有跟著收。** 它們各自有不同的判定機制(`ResolveMissileShot` 已經
        在第 133 項收了一個 `MissileDefenses`),硬要三條路共用一個結構會做出一個
        「什麼都裝得下、什麼都說不清楚」的型別。
      - 第 133 項那一桶還剩 **14 個**,各自的擋門理由寫在 `ship_systems.go` 檔尾。

    截圖:25_shipdesign 的「特殊 1/23 → 1/25」(10 px)。

136. **戰鬥速度與引擎階:一張自我驗證的一手表,外加一個掃描器看不到的硬編值**(2026-08-08)。

    第 133 項那一桶剩下的 14 個,擋門理由聚成幾個共同缺失的機制。最大的一個是
    **戰鬥速度**——它一個人擋住五個元件(增強引擎、牽引光束、停滯力場、時間扭曲加速器、
    黑洞產生器),而 `gamedata.CombatSpeed` 早就移植好且零呼叫端。

    ### 表從執行檔挖出來,而且自己驗證自己

    `Current_Design_Min_Combat_Speed_`(0x6B82A)讀的是一張二維表:

        mov dx, [eax+1Bh]              ; 引擎階
        mov bx, [eax+0E5h]             ; 艦體等級 0..5
        imul eax, 2Eh                  ; 每列 46 位元組
        mov al, byte_17FE90[edx+eax]   ; 最小速(最大速在 +6)

    攤開來六階 × 六艦體:

        階1 min=[10  8  6  5  4  3]   max=[20 18 16 15 14 13]
        階2 min=[12 10  8  7  6  5]   max=[22 20 18 17 16 15]
        …每階全部 +2…
        階6 min=[20 18 16 15 14 13]   max=[30 28 26 25 24 23]

    三個規律,每一個都不像巧合:

      - **每升一階引擎全部 +2** ——與手冊那張戰機表的註腳逐字相符:
        「* base speed is modified by **+2 per drive level**」
      - **最大速恆等於最小速 +10**,36 組全部成立
      - 艦體越大越慢,而且**遞減幅度自己也遞減**(10/8/6/5/4/3)

    所以 `gamedata/combat_speed.go` **只存第一列**,其餘用公式算。存整張表反而讓那三個
    規律變成「36 個各自獨立的數字」,改壞一格不會有任何東西發現。

    ### 一個掃描器看不到的硬編值

    `gamedata` 裡一整組公式吃 `ftlLevel`(飛彈速度、飛彈光束閃避、戰機速度),
    註解都寫著「速度 = 基礎 + 2×(FTL−1)」。而唯一的生產端呼叫是:

        shell.NewFighterSquadron(s.BayKind, false, idx, s.Col, s.Row, **1**, 0)

    **硬編成 1**,所有戰機不論科技多高都跑得一樣慢。它在 `cmd/moo2/tacticalfighter.go`,
    而第 129 項那個掃描器只掃 `gamedata.X(...)` 的引數——**看不到 cmd/ 這一側**。

    > **掃描器的盲區也要記帳。** 「③ 已經掃過了」如果被當成「這一類洞已經清乾淨」,
    > 它就變成另一句過期的擋門理由——而這一輪已經撞過七次那個形狀。
    > ③ 的殘量表現在註明:只涵蓋 `gamedata` 的匯出函式,cmd/ 與 shell 內部函式沒掃。

    ### 接上的

      - **引擎階模型**(`driveLevel`):取最高已研究的引擎。MOO2 的引擎是自動升級的
        (手冊 Fusion Drive:「added to all your ships … **as soon as you complete your
        research**」),不是逐艦選裝。
      - **主動權排序**:手冊給了確切公式「A ship's initiative is equal to its current
        **Beam Attack plus 10 times its current combat speed**」。先前齊射是**艦隊清單
        順序**——等於「先造的先打」,與速度完全無關。
        ⚠ 用**穩定**排序:同分必須維持原序,否則同一場戰鬥可能打出不同結果,
        存檔與探針的可重現性整個失效。
      - **增強引擎**(+5 戰鬥速度)——第 134 項把它列在「機制不存在」那一格,現在成立了。
      - **戰機的引擎階與裝甲級數**改用真值(那個硬編的 1 / 0)。
      - 崔拉里安的**跨維度 +4** 有了第二個消費端(第 129 項挖出來的種族特性)。

    ⚠ 增強引擎的主題我一開始猜「進階建設」,執行檔說是 **Advanced Fusion**。
    **每一個主題都去查,不要因為前幾個猜對了就開始猜。**

    ### 誠實留白

      - **min/max 之間怎麼取,remake 不模擬。** 反組譯的 `Current_Design_Combat_Speed_`
        是 `speed = max − (ratio × (max−min) / 100)`,ratio 由設計的 +0xE9/+0xED 兩個欄位
        算出。那兩個欄位是什麼**沒有查證**(看起來像佔格用量/總格數),所以不猜——
        remake 取 max 那一端並標明差距最多是 10。
      - **移動距離還沒限制。** 戰術棋盤只有 8×6,而速度值是 13..30,直接當「一回合走幾格」
        會讓任何船一步橫跨全場。要限制就得先決定一個棋盤比例尺,**那是 remake 的設計決定
        不是轉寫**,不夾在這一項裡做。牽引光束/停滯力場那幾個仍然卡著。

    截圖:25_shipdesign 的「特殊 1/25 → 1/26」(10 px)。

137. **戰術移動不再是瞬移:棋盤比例尺從一手尺寸推出來**(2026-08-08)。

    第 136 項把戰鬥速度接上了,但移動距離**還是沒有限制**,理由寫得很明白:

    > 戰術棋盤只有 8×6,而速度值是 13..30,直接當「一回合走幾格」會讓任何船一步橫跨全場。
    > 要限制就得先決定一個棋盤比例尺,**那是 remake 的設計決定不是轉寫**。

    那句話有一半是對的,而另一半可以查。**原版棋盤多大是一手事實**——查到了,
    這個「設計決定」就從「憑感覺挑一個數字」變成「從一手尺寸推出來的換算」。

    ### 原版棋盤是 81 × 68

    `Assign_Combat_Grids_`(0x46CC8)開頭把整張格點清成 0xFFFF:

        loc_46CD4:
          xor  eax, eax
        loc_46CD6:
          imul esi, ecx, 88h              ; 列距 0x88 = 136 位元組 = 68 格 × 2
          mov  word_18C9A8[esi+ecx*2], 0FFFFh
          cmp  ax, 44h                    ; 0x44 = 68
          jl   short loc_46CD6
          inc  ebx
          cmp  bx, 51h                    ; 0x51 = 81
          jl   short loc_46CD4

    **兩個界限與列距互相驗證**:列距 136 位元組正好裝得下 68 個 uint16,而內圈上界就是 68。
    同一支函式稍後用 `[ship+21h]`/`[ship+22h]` 當索引寫回艦艇編號,所以那兩個位元組
    就是艦艇的棋盤座標。

    這也解釋了手冊那些「以格為單位」的射程為什麼可以那麼大:質子魚雷 24 格、
    傳送器 12 格、投彈 3 格——放在 81×68 上都是合理的比例。

    ### 比例尺

        81 / 8 ≈ 10.1        68 / 6 ≈ 11.3

    速度 13..30 → remake 盤面 **1..3 格**。

    ⚠ **「remake 用 8×6」本身是 remake 的簡化**,所以這個換算仍然是一個**明說的決定**。
    但它是**從一手尺寸推出來的**決定,而不是「覺得走 2 格差不多」——差別在於:
    以後把棋盤放大,這個換算會自動跟著對。

    保留的是原版的**相對關係**:小船走得比大船遠、引擎升級走得更遠、增強引擎再多一點。
    下限 1 格是刻意的——再慢的船也要能動,否則末日之星在低引擎階會完全無法移動,
    那不是手冊講的東西(手冊只說大船比較慢)。

    ### 距離度量要與射程一致

    移動用**曼哈頓距離**,與同一畫面的射程判定(`abs(dc)+abs(dr)`)同一個度量。
    **兩處用不同度量會讓「走得到卻打不到」變成一種很難解釋的行為**——玩家看到的是
    「我明明移到旁邊了」,而程式看到的是兩個不同的圓。

    ### 一個差點寫錯的順序

    移動力重置一開始放在 `t.round++` 那裡,看起來最自然(「回合開始時重置」)。
    **那是錯的**:同一個函式稍後會把陣亡艦從 `t.player` 壓縮掉,長度變短、索引往前移,
    而 `moveLeft` 還停在舊長度——選中第 3 艘會讀到第 5 艘的移動力,而且陣列會越界。
    改到壓縮**之後**重置。兩個索引平行的陣列就是會有這種問題,註解已標明。

    ### 誠實留白

      - **敵方艦不受移動限制**:`genEnemyFleet` 造出來的敵艦沒有引擎/艦體設計資料,
        而 AI 在戰術畫面也沒有移動行為(它只還擊)。這是既有簡化,不是這一項引入的。
      - **牽引光束/停滯力場仍然卡著**:它們要的不是「能走幾格」,是「**讓對方**走不動」
        ——需要一個逐艦的狀態旗標與回合結構。移動預算是那個機制的一半,不是全部。

138. **牽引光束與停滯力場:第 128 項那個「卡在機制」的判定終於解掉**(2026-08-08)。

    第 128 項盤點手冊 p.127 的特殊武器,把這兩個歸進「卡在機制,不是卡在資料」:

    | 武器 | 手冊效果 | 當時缺什麼 |
    |---|---|---|
    | 牽引光束 | reduces target speed | **戰鬥速度模型** |
    | 停滯力場 | places target in suspended animation | **「這一輪不能動」的狀態** |

    第 136 項建了戰鬥速度、第 137 項建了移動預算,**缺的最後一塊是「讓對方動不了」**
    ——那是逐艦狀態,不是自己的能力。這一項補上。

    ### 手冊的規則比想像中精確

    > 牽引光束:Each beam can **trap a Frigate class ship or slow a larger one — in
    > proportion to its size** — up to the maximum range of **12 squares**. The effect of
    > multiple Tractor Beams on a single target is **cumulative**. (Thus, for example,
    > **6 beams would immobilize a Doom Star**.) … An immobilized ship receives an
    > additional **−20 Ship Defense** penalty…

    **那個括號裡的例子把公式釘死了。** 末日之星是第 5 級,而 6 = 5+1;另一端
    「Each beam can trap a **Frigate**」= 第 0 級 1 束。**兩端都對上,中間就是線性**,
    不是猜的。這與第 127 項球形武器的「級數 = index+1」是同一個讀法,兩處互相支持。

    > 停滯力場:While suspended, the ship cannot move, fire, recharge …
    > **or be affected by any weapon**. It is effectively **removed from battle entirely**.

    「or be affected by any weapon」那一句最容易漏。**只做「不能動」會讓它變成活靶**
    ——那是相反的效果。實作把兩邊都擋:被定住的不能打,也不能被打。

    ### 每回合重算,不累積

    手冊把兩者都描述成**持續的場**而不是打出去的一發:

        停滯力場:it remains in effect **as long as the ship generating the field remains
                 undestroyed and in combat**
        牽引光束:**up to the maximum range of** 12 squares away

    所以每回合從零重算才是對的:產生源被打掉、或目標飛出射程,效果就該消失。
    累加的話會出現「產生源早就沒了,目標還定在那裡」。

    ### 射程換算:向上取整

    第 137 項的比例尺是 1:10。手冊的 12 格 → 1.18、3 格 → 0.29,**向下取整會讓停滯力場
    的射程變成 0(等於這個武器不存在)**。改成向上取整:2 格與 1 格,
    **「牽引比停滯遠」這個相對關係保住了**——那才是手冊在講的事,絕對值在 8 欄的盤面上
    本來就表達不出來。

    ### 三個順序,錯一個就壞

    回合結束時的處理順序是:**戰損壓縮 → 重算狀態 → 重置移動力**。

      - 狀態要在壓縮之後:產生源被擊毀的效果才會消失
      - 移動力要在狀態之後:移動格數是依**實際**速度算的,而實際速度吃狀態

    ### 誠實留白

      - **一艘船只有一個 Special 槽**,所以手冊的「multiple Tractor Beams … cumulative」
        在 remake 只能靠**多艘船**達成——而那正好是原版最常見的用法(一群小船拖住一艘大船)。
        測試裡六艘拖船定住末日之星那一條就是這個形狀。
      - **AI 不會用**:敵方艦由 `genEnemyFleet` 生成,沒有元件設計資料。既有簡化。
      - **「可被登艦」那一半沒接**:手冊說被定住的船 can be boarded,而登艦戰機制不存在
        (第 119 項)。這一項只接查得到落點的那些。
      - **時間扭曲加速器仍然沒接**:它要的是「一回合行動兩次」,而 remake 的戰術回合是
        「開火即結束回合」的單一結構,插一個額外回合會動到整個流程。與這一項的
        狀態旗標不是同一件事。

    截圖:25_shipdesign 的「特殊 1/26 → 1/28」(3 px)。

139. **陀螺去穩器:擋門理由是對的,而正解不是照著它做**(2026-08-08)。

    第 128 項把陀螺去穩器擋在外面,理由寫得很具體:

    > 資料齊、艦體級數也有了(第 127 項)——但它**不在手冊的球形清單裡**,
    > 而 remake 的光束路徑沒有「per size class」這個乘數。

    **那句話每個字都對。** 但它把問題導向了「要不要替光束加一個乘數」,
    而正確的問題是「**它到底是不是光束**」。

    手冊那一條:

    > This uncontrolled twirl causes **1–4 points of structural damage multiplied by the
    > size class of the ship**. **Shields and armor offer no protection** and are not damaged.

    **兩個定義性特徵都是球形家族的**:傷害依目標級數相乘、完全豁免護盾與裝甲。
    它沒有列在手冊 p.126 的球形清單上,但清單是分類法,行為才是分類。
    歸進球形路徑之後一行乘數都不必加——第 127 項寫的那條路直接接得上。

    > **擋門理由講的是「照現在的分類做不到」,不是「這件事做不到」。**
    > 這一輪撞了七次的那個形狀(理由過期)還有一個變體:**理由沒過期,但它預設了
    > 一個不必接受的前提。**

    ### 佔格值怎麼確定的

    手冊的特殊武器表在 PDF 抽取後欄位是打散的:五個武器名/效果先出來,
    接著才是 Size 欄的五個數字 `20 75 30 50 40`。

    **對齊靠的是脈衝星 = 50 正好對上第 127 項已經從 p.127 記下來的值。**
    第四項對上了,整列的對應就確定了——不必猜,也不必假裝那是顯然的。
    陀螺去穩器 = **75**。

    同一列還順便讀到:反飛彈火箭 20、牽引光束 30、電漿網 40、停滯力場 75。
    那四個目前走 `SpecialOptions`(建造成本)而不是佔格表,先記在 `shipspace.go` 裡,
    等 remake 有逐元件佔格的造艦模型時直接用。

    ### 誠實留白

      - **只有空間壓縮器與陀螺去穩器豁免盾甲**,脈衝星沒有那一句。第 127 項就標過
        「不要推廣到整個球形類別」,這一項加了測試把反面也釘住。
      - **剩下的 10 個**(超載電容 / 快速飛彈架 / 時間扭曲加速器 / 轟炸機庫 / 能量吸收器 /
        傳送器 / 匿蹤力場 / 相位匿蹤 / 保安站 / 戰鬥艙)各自的理由仍寫在
        `ship_systems.go` 檔尾。其中前三個共用同一個缺失機制:**回合內的行動次數**
        ——那是下一個值得建的東西,一次解三個。

    截圖:25_shipdesign 的「武器 4/25 → 4/26」(10 px)。
