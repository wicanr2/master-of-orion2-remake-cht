# 總缺口報告:原版 Orion2.exe vs remake

> 日期:2026-08-06(當日稍晚修訂)。方法:解析原版執行檔內建的 8,589 個 Watcom 除錯符號
> (見 [`00-orion2-symbols.md`](00-orion2-symbols.md)),對照 remake 現行程式碼。
> **這是第一次能用原版二進位當基準做全面盤點**——先前只能靠手冊、攻略與 openorion2(純渲染殼)。


---

## 這 85 項煉出來的規則(先讀這一段,底下是參考資料)

> 編號項裡的教訓高度重複——「擋門理由會過期」出現十次、「兩條戰鬥路徑」
> 出現在第 68–71 項每一項。合併成下面 12 條;**要用的是這些,不是逐項讀完**。
> 每條後面是產生它的項次,要看證據再去翻。

### 一、關於「還缺什麼」的判斷

1. **擋門理由會過期,而且可能當初就寫錯。** 「為什麼這項還沒做」那句話在寫下的那天
   通常是對的,但沒有任何機制會在它失效時提醒你。更糟的一種是**從寫下的那天就不成立**
   ——「手冊沒公開這個數字」往往只表示沒查過。**標「近似」的前提是真的查過。**
   〔58 59 63 64 65 66 68 70 71 72〕
2. **「元件表裡有」不等於「效果有接」。** 一項可以在表裡、花得了錢,而程式碼沒有任何
   地方讀它。盤點時要分開問這兩件事;偵測法是掃名字在表以外出現幾次,零次就是裝飾品。〔72〕
3. **任何清單都會過期,包括警告別人清單會過期的那份清單。** 做完一項就回去刪那一行。〔72〕
4. **「已經沒有剩下的了」是最貴的斷言**,因為它讓下一輪不去找。〔全部 75 項〕
5. **掃描器只負責縮小範圍,判定仍要逐條看程式碼。** 照著可疑清單「順手都改掉」,
   會把正確的斷言改錯。〔72〕

### 二、關於改動的完整性

6. **戰鬥有兩條路徑(快速結算 / 格子戰術),任何效果改動都要問「兩條都接了嗎」。**
   只接一條會安靜地不一致。〔68–71〕
7. **一個元件常有兩個效果,只接一半不會有任何徵兆。** 從單一檔案回推元件效果會漏東西;
   要從**手冊那一段的每一句**回推。〔68 71〕
8. **資料層跟著實作層缺,是很難發現的洞。** 單看資料層會以為「這一組本來就只有這些」。〔70〕
9. **測試名稱裡的「只」是一種斷言**,而它斷的往往是你當時的理解,不是手冊。〔71〕

### 三、關於逆向

10. **一手來源永遠贏二手推論。** 一手 = 執行檔立即數、LBX 位元組、手冊行文;
    二手 = 檔名、美術字樣、命名習慣。**openorion2 沒有 ≠ 沒有真值**——反組譯全都有。〔14 65 67〕
11. **不要 grep `.asm`,要查 IDA 資料庫的 xref 圖。** `.asm` 是攤平的文字,沒有交叉參考。
    而且除錯符號表的 `obj1+偏移` **不是**線性位址(差一個 object base)。〔73〕
12. **查詢回空時先做正對照**——拿一個已知一定成立的目標跑同一套查法。零命中與
    「查法壞了」長得一模一樣。〔73〕

### 四、驗收

- **測試綠 ≠ 對齊原版。** 單元測試只證自製邏輯自洽。
- **逐位元比對過場截圖廊**(34 張)是本專案實際在用的驗收。
- **狀態指紋變了就代表持久化狀態變了**,必須查出原因,不可因為「改動有理」就放過。〔71〕
- **手冊沒給的數字就不填。** 不臆造,寧可記錄擋在哪。
- **壞掉的驗證腳本比沒有驗證更糟。** 它給的是「我檢查過了」的錯覺。腳本回報一堆不符時,
  先懷疑腳本本身抓錯範圍,再去改被檢查的東西。〔76〕
- **「引用有沒有失效」不等於「引用對不對」。** 被改成一個**存在但錯誤**的項次,失效檢查
  完全抓不到——要比對的是引用指向的內容,不是編號存不存在。〔76〕

---

## 85 項速查表(引用裡的括號就是這裡的短標)

| # | 短標 | # | 短標 | # | 短標 |
|---|---|---|---|---|---|
| 1 | 星系生成表 | 27 | 獨立戰機單位 | 53 | 聊天列字沒離開機器 |
| 2 | 一星多行星缺口 | 28 | AI遷移非缺口 | 54 | 三個寫入端 |
| 3 | Colony+Event 畫面 | 29 | 決定性化 | 55 | AI 科技先前靠偷 |
| 4 | 地面戰解算 | 30 | TECH LEVEL 第二效果 | 56 | 領袖永久免費 |
| 5 | NEW GAME 設定畫面 | 31 | 開局建築清單 | 57 | 征服人口產出 |
| 6 | 殖民地畫面框架 | 32 | 飛彈速度 | 58 | 擋門理由過期三個月 |
| 7 | 曲速前FTL限制 | 33 | 地面戰結構 | 59 | 成就科技效果 |
| 8 | 行星表面調查 | 34 | 重力種族特性 | 60 | 打得準也閃得掉 |
| 9 | 星圖艦隊圖示 | 35 | 三張查表 | 61 | 兩個函式畫面沒用過 |
| 10 | 蟲洞 | 36 | 聊天列補完 | 62 | 護盾減傷 |
| 11 | 48棟建築盤點 | 37 | 研究樹一手驗證 | 63 | 登陸能力完全相同 |
| 12 | 四大系統缺口盤點 | 38 | 行星護盾等三棟 | 64 | 武器傷害真表 |
| 13 | 指揮點數視窗 | 39 | 艦員經驗 | 65 | 種族特性31格 |
| 14 | 地表道路 | 40 | 同化系統 | 66 | 高能聚焦 |
| 15 | 星雲 | 41 | 戰機基地 | 67 | 裝甲科技倍率 |
| 16 | 秒差距模型 | 42 | 關掉兩條留白 | 68 | 元件盤點+飛彈防禦 |
| 17 | 拓殖基地 | 43 | 先進級開局主題 | 69 | 戰鬥速度與引擎階 |
| 18 | 銀河貨幣交易所 | 44 | 下游讀真值 | 70 | 陀螺去穩器 |
| 19 | AI請求會談 | 45 | 領袖技能 | 71 | 探針③內部函式 |
| 20 | 手冊搜尋假陰性 | 46 | 叛亂 | 72 | 元件表有≠效果有接 |
| 21 | 改造不佔格 | 47 | AI艦隊移動 | 73 | 音樂場景表 |
| 22 | 黑洞動畫 | 48 | 三個待辦複查 | 74 | 文件政策 |
| 23 | 多艦隊模型 | 49 | 安塔蘭防禦艦隊 | 75 | 編號壓縮 |
| 24 | 軌道資料層 | 50 | 軍官畫面座標 | 76 | 交叉引用不必開文件 |
| 25 | 人造行星 | 51 | 間諜UI | 77 | 元件表真值 |
| 79 | 武器表與艦體表 | 80 | 登艦戰 | 81 | 淘汰簡約殼 |
| 82 | 音樂兩個缺 | 83 | 打包路徑 | 84 | 名稱池雙語化 |
| 85 | 元件名英文 |  |  |  |  |
| 26 | 確認框 | 52 | 生物武器分類 | 78 | 音樂接線 |

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
| `Do_Main_Game_Popup_` @ 0x7DD41 / `_Draw_Main_Game_Popup_` @ 0x7F701(遊戲中的「遊戲」選單)| `gameMenu` | ✅ 2026-08-07 已建(`cmd/moo2/gamemenu.go`),星系畫面頂端「遊戲」鈕先前是死的 |
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
| 12 | 多人連線 11 個畫面 | — | ✅ `MP_Setup`(`cmd/moo2/multiplayer.go`)與 `Hotseat`(`cmd/moo2/hotseat.go`)2026-08-07 已建,版面座標取自反組譯(見下方第 3 項(Colony+Event 畫面))。`Net_Next_Turn`(第 29 項(決定性化))與 `Choose_Net_Plyrs`(第 29 項(決定性化))2026-08-07 已建。`Modem_Setup`/`NullModem_Setup`/`Comm Info` **不做**(硬體已不存在)。`Join_Net`/`Generic_Net_Info`/`SendGet_Net_Info` 是同一張畫面的不同狀態(第 29 項(決定性化)),`Choose_Multi_Net_Game` 見第 29 項(決定性化)。**11 張全部結案:8 做 / 3 不做** |

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

### 「最大的系統級缺口」那份清單:四條全部已經做掉了(2026-08-07 逐條核實)

| 項目 | 現況 | 證據 |
|---|---|---|

| ①**歷史記錄系統** | ✅ 已建 | `internal/shell/history.go`(6 個函式)+ `cmd/moo2` 的 `infoHistory` 折線圖 |
| ②事件系統 | ✅ 2026-08-06 已對齊 | 36 種事件表 + GNN 快報畫面;剩餘 20 種缺的是各自的子系統,逐項記在 `gamedata.RandomEvents` 的 `Needs` 欄 |
| ③**前哨站(Outpost)** | ✅ 已建 | `internal/shell/outpost.go`(9 個函式)+ 進存檔 + 進熱座席位 + 可被 `consumeOutpostForColony` 升級成殖民地 |
| ④**艙損/維修**(module 20) | ✅ 已建 | `internal/shell/repair.go`(11 個函式),手冊 p.80/82/25 逐字 + `Repair_Ships_At_Colonies_` @ 0x580F5 雙重錨定 |

**教訓與第 11 項(48棟建築盤點)同一條**:斷言一旦寫進文件就會被當成現況引用,而程式碼會往前走、文件不會。
下結論說「X 做了 / 沒做」之前先 `grep` —— 這四條每一條都只要一次 `ls` 就能推翻。
(規則出處:`rulebook/63-truth-in-code-not-stale-markers.md`。)

### 剩餘工作在哪裡

> **剩餘工作表已於 2026-08-08 搬到 [`WORKLIST.md`](../../WORKLIST.md) 頂端,那裡是唯一活來源。**
>
> 這裡原本也有一份,而兩份活表並存正是本專案反覆踩過的坑
> ——複製出來的每一份都會過期,而過期的斷言會被當成現況引用。
>
> **本文的職責改為兩件事,都不是「現況」**:
>
> 1. **RE 硬資料**:196 個十六進位位址、131 列表格(執行檔立即數、資料表、公式)。
>    這是全 repo 最貴的內容,每一項重挖都是數小時起跳。
> 2. **工程日誌**:每一項是怎麼被找到的、當初怎麼推錯的。
>    第 74 項(文件政策)訂下的分界——**其餘所有文件只留現況,錯誤推導的全文只留在這裡**。
>
> 要知道「還剩什麼」,看 WORKLIST;要知道「某個數字從哪來 / 某項當初怎麼錯的」,看這裡。

**2026-08-07 逐條核實後從這張表刪掉的**(做完了,但表沒跟上):

| 原本的斷言 | 現況 |
|---|---|
| **同星系多殖民地** | ✅ 第 24 項:拓殖/前哨站的對象改成**行星**;人造行星改造完的天體也真的能殖民了 |
| **戰術格子的獨立戰機單位** | ✅ 第 27 項:中隊在格子上有自己的位置(出擊→飛向目標→貼身開火→返航補給→再出擊)。⚠ 同時訂正「中隊數 4/2」——那是把 p.127 的 **Shots** 欄讀成了中隊人數,正確答案是**一律 4 架** |
| AI 的遷移設定 | ✅ **不是缺口**(第 28 項(AI遷移非缺口)):那個欄位的五個寫入者沒有一個在 AI 的程式碼裡——原版的 AI 也不設集結點 |
| AI 的同星系多殖民地 | ✅ 第 24 項:`aiExpand` 的候選集加進「自己已有殖民地的星系」;順帶修好入侵提早翻面的 bug |
| 遷移連線的顯示開關沒有 UI | ✅ 第 23 項(多艦隊模型)⚠ 那一列不是原版版面,原版有一整個設定畫面 |
| `Clear_All` 集結點沒有 UI 入口 | ✅ 第 23 項(多艦隊模型)。⚠ 同時訂正——**ALL 鈕不是 Set_All**,手冊兩處明說它是「全選/全不選艦艇」 |
| 遷移確認框 | ✅ 第 26 項(確認框,版面取自 `Confirmation_Box_`)。⚠ 同時訂正——原版那個條件是**怪獸**不是艦隊 |
| `Command_Points` 專屬畫面沒做 | ✅ `cmd/moo2/commandpoints.go`(第 13 項(指揮點數視窗)) |
| 殖民地地表缺「擺放的最後一段抖動」 | ✅ 第 14 項(地表道路) |
| 殖民地地表缺植被層 | ✅ 第 14 項(COLVEGGI,13 組 × 8 張 = 104 對上資產數) |
| 艦隊圖示 8 幀「動畫沒做」 | ✅ 第 22 項:**原版就不會動**(每次繪製前歸零幀號),假缺口 |
| 一般星球閃爍「三個常數沒追出來」 | ✅ 第 22 項:**出貨版是死碼**(沒有任何地方啟動它),假缺口 |
| 鍵盤快捷鍵 | ✅ 第 22 項(黑洞動畫)全數接完(F1/F2、F5/F6、F9、F10、ALT+F9)。ALT+F1..F8 的設定開關仍不碰(PDF 邊欄標籤有 off-by-one 風險) |
| 星圖遷移連線 | ✅ 第 23 項(多艦隊模型)。**星圖 4 層全部到齊**:蟲洞(35)、遷移連線(57)、星星、星門(47)+ 星雲(44)+ 外交燈號(50) |
| 多艦隊模型 | ✅ 第 23 項(多艦隊模型:資料層 / 艦隊列表 / 分艦隊三個小節) |
| RELOCATE 鈕的原版語意 | ✅ 第 23 項:兩段點選 + 四條合法性規則,已接 |
| 分艦隊 UI | ✅ 第 23 項(⚠ 入口是 remake 自己加的,原版在右側艦艇格) |

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


## 一、符號表、畫面清單、資料表(2026-08-06)

1. ~~星系/行星生成表~~ → **已接進 remake**(`gamedata/galaxygen.go`,commit `f8bbcbd`)。

   光譜/大小/氣候/礦產/重力/行星數全部改用原版骰表,並加了分布回歸測試。


    ### ~~殖民地經濟表~~ → **已交叉驗證**:`_food_per_farmer_table`、`_minerals_per_mine`、


       `_planet_max_population` 三項與 remake 現值一致,不必改(見 C-2)。
       `_climate_maintenance_modifiers` 仍待確認讀取者。


    ### ~~INFO 5 個子畫面~~ → **已實作**(commit `fade3f7`),含 module 122 歷史記錄系統。


    ### 第二梯 — 下一批


    ### ~~射程命中/傷害懲罰~~ → **已完成**。查證發現兩張表 remake 早就有且與原版逐格相同


       (`combatRangeLevelPenaltyTable`、`damageDissipationPenaltyTable`);真正的缺口是
       **傷害衰減從未被呼叫**——同一發雷射在 1 格與 23 格外傷害一樣。已接進
       `ResolveShotWithMods`,順帶讓 NR(No Range Dissipation)改造第一次有實際效果。


    ### ~~`_personality_*` 表 + AI 行星估值~~ → **都已完成**(2026-08-06),含後續補上的


       `Proximity_Worth_To_Player_`、`Compute_Contextual_Planet_Values_`、`Colony_Worth_To_Player_`。
       最後一個 `Enemy_Colony_Worth_To_Player_` 是「攻擊目標的**額外**加權」,
       remake 目前用 `AIColonyValue ÷ 距離` 代打(見 `shell/ai_attack.go`)。

       `Colony_Worth_To_Player_` 的估值挑玩家最有價值的殖民地突襲(`shell/ai_attack.go`),

2. **一星多行星** → 原版每顆恆星 1–5 顆行星各佔一條軌道;remake 的 `Stars`/`Planets`

   索引一一對應是 UI/拓殖/AI 共同的假設,拆開是跨層改造(見 `genPlanets` 註解)。

   代表行星挑「最適合殖民的那一顆」,其餘記在 `Planet.SystemBodies` 供顯示與日後的前哨站。
   `Stars`/`Planets` 的一一對應**沒有動**,所以 UI/拓殖/AI 完全不受影響。

3. ~~獨立 Colony 畫面 + Event 畫面~~ → **兩個都已完成**(2026-08-06)。


    ### ~~前哨站~~ → **已完成**(2026-08-06,`internal/shell/outpost.go`):前哨船可建造、


       可在氣態巨星/小行星帶/一般行星建立軍事前哨站,前哨站是掃描站(併進 detection.go 的偵測源)
       且**沒有人口與產出**(手冊 p.85「produces nothing」,故不進 PlayerColonies);之後在同一顆星
       建殖民地時前哨站改建為海軍陸戰隊營(手冊逐字)。順帶補上**殖民船也終於可以建造**——
       先前開局送一艘、用掉就再也不能擴張。
       ⚠ 未兌現的一半:「延伸艦艇航程 / 加油站」(手冊 p.119/p.133)——remake 的 SendFleet 沒有
       航程上限這個概念,沒有可套用的機制,不臆造。
       ⚠ 手冊 p.50 的「前哨站升級成可住人殖民地」科技仍未做(需要對應科技旗標)。


    ### ~~太空怪獸~~ → **已完成**(2026-08-06,`internal/shell/monster.go` + `gamedata/space_monster.go`)。


        這一項有個值得記的地方:**它一直被程式碼引用著卻不存在**——colonization.go 檔頭抄的手冊
        原文就寫著殖民船要「as long as all space monsters and enemy ships have been cleared from
        that planet's system」,但那個 gate 從來沒有東西可擋。

        - 五種怪獸的名字來自執行檔字串表(0x1F742C 起連續:Guardian / Amoeba / Dragon / Hydra /
          Crystal),對應五個 `Load_*_Ship_Design_` 函式與 `_monster_names` @ 0x199266
          設定(`_user_wants_n_space_monsters` @ 0x19A006),remake 先用固定密度


    ### ~~持續型隨機事件~~ → **已完成**(2026-08-06,`internal/shell/events_persistent.go`)。


        先前 9 個事件卡在「缺子系統」,真正缺的是同一個東西:remake 只有「單次結算」的事件模型,
        **沒有任何跨回合的事件狀態**。補上那個模型之後,手冊 p.180-181 就直接是規格書:

        | 事件 | 手冊給的數字 |
        |---|---|
        | 超新星(24) | ≥200 回合觸發、倒數 6-14 回合、系統研究點全投入搶救、失敗則全滅 + 行星變輻射 |
        | 時空異象(25) | 星系凍結:不生產不成長,也不吃食物不繳維護費;6 回合後每回合 5% 結束 |
        | 超空間獸(26) | 航行中的艦隊有機率損失一艘;6 回合後每回合 5% 離開 |
        | 蟲洞(28) | 航行中的艦隊「in a single turn」抵達 |
        | 怪獸入侵(19-23) | 變形蟲 ≥100、太空鰻 ≥150、水晶 ≥200、九頭蛇 ≥250、巨龍 ≥300 |
        `Needs` 欄。
        **這一批順帶翻出一個過期斷言**:`advanceConquestVictory` 的註解寫著「remake 沒有任何機制
        `CheckExtermination`(只剩一方存活)在「玩家死光但還有三個 AI」時回 false——400 回合探針
        實測到玩家 0 殖民地、遊戲卻繼續空轉。已補 `advancePlayerDefeat`(手冊 p.184 計分段明講


    ### ~~Hall of Fame / Hi-Score~~ → **已完成**(2026-08-07,`gamedata/score.go` + `shell/score.go`


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
        `Do_System_Discoveries_At_Star_` 讀遠古文物時用的是同一個偏移,83 也與 remake 既有的
        研究主題數相同;時間分用 `word[0x192FD8] − 0x88B8` 算已過回合,0x88B8 = 35000 =
        「這個玩家滅了誰」陣列(player+0x1F2),remake 沒追蹤是誰滅的,目前全算給玩家,


    ### ~~艙損/維修~~ → **已完成**(2026-08-07,`internal/shell/repair.go`)。這一項的起點不是


        「補一個修復系統」,而是先發現 **remake 根本沒有艦艇損傷這個概念**——一艘船不是完好就是
        被擊沉,打完慘勝的仗倖存艦跟出港時一模一樣。於是「自動修復」這個元件
        (`SpecialOptions` 裡的 `{"自動修復", …, TECH_AUTOMATED_REPAIR_UNIT}`)從加進來那天起
        就沒有任何效果:沒有損傷可修。

        | 規則 | 來源 |
        |---|---|
        | 戰鬥艦停在自家據點→ **完全修復** | `Repair_Ships_At_Colonies_` @ 0x580F5 直接呼叫 `Repair_Ship_Full_` @ 0x581F3；2026-08-24 IDAPython 補證 `Design.Type==COMBAT_SHIP`、Status／Star／owner bit 門檻。殖民地／前哨站對該 bit 的寫入端仍為強推論，見 `ship-repair-audit-20260824.md` |
        | 自動修復元件:戰鬥中每回合修 **20%** 結構損傷 | 手冊 p.82 逐字 |
        | 自動修復元件 / 進階損害管制:**戰後完全修復** | 手冊 p.82 / p.80 逐字 |
        | 機械化種族:戰鬥中 10%/回合(常數已備,無呼叫端) | 手冊 p.25 逐字;remake 沒有種族特質欄位可掛 |
        ```c
        uint8_t  shieldDamage, driveDamage;   // percent
        uint8_t  computerDamage, crewLevel;
        uint8_t  damagedSpecials[(MAX_SHIP_SPECIALS+7)/8];
        uint16_t armorDamage, structureDamage;
        ```
        `structureDamage` 這一份。`ships.cpp:1060` 的 `isSpecialDamaged(i)` 把壞掉的元件名字
        損傷這一半;原版對應的 `Apply_Internal_Damage_` @ 0x35251 是個依傷害類型分十幾條分支、
        操作艦艇結構 +0x29/+0xC2/+0x134 等欄位的大函式,要接它得先有逐系統模型 ②裝甲與結構在
        remake 合一 ③「進階損害管制」科技還不在科技樹裡,`playerHasAdvancedDamageControl` 恆
        順帶記一個踩到的坑:截圖廊原本在 `galleryVictoryTick`(t28)注入損傷,而 t29 按了
        正常運作:艦隊開局就停在母星,照 `Repair_Ships_At_Colonies_` 的規則被完全修復。
        `combatant` 結構裡,過濾陣亡者時整個 struct 複製、索引跟著倖存者走。


    ### ~~安塔蘭房間~~ → **已完成**(2026-08-07,`cmd/moo2/antaranroom.go`)。原本這條勝利路徑的


        入口是艦隊列表左下角一行文字,點下去直接跳戰鬥結果——中間沒有確認、沒有戰力對比,
        前置條件不滿足時更是「點了完全沒反應」(`CanAssaultAntares` 回 false,而畫面上毫無跡象)。

        美術來源是反組譯 `sub_14C83`:`mov edx, 0 / mov eax, offset aAntaroomLbx / call sub_126B42`
        ——載 `antaroom.LBX` **資產 0**。實際查下去,資產 0 是個小圖但**帶內嵌調色盤**,資產 1 才是
        出得來,拿 `buffer0` 的色盤解會是一團彩色雜訊。累積成最終畫格的做法沿用外交議事廳
        (`DIPLOMAT#29`,38 幀)那條已驗證的路徑(`lbx.Image.AccumulatedRGBA`)。
        發動/撤退兩顆按鈕、以及**擋下時逐條講明卡在哪**(`AssaultAntaresBlockReason`:勝負已定 /
        ⚠ 留白:原版把 55 幀當推鏡動畫播,remake 只呈現最終定格(`overlayScreen` 沒有動畫層,


    ### ~~地面戰畫面~~ → **已完成**(2026-08-07,`cmd/moo2/groundcombat.go`)。這一項的價值不在


        「多一個畫面」,而在**版面座標一個都不是估的**——全部從反組譯挖出來:

        | 元素 | 真值 | 來源 |
        |---|---|---|
        | 攻方面板外框 / 暗化區 | 貼圖 (1,40);`Darken_Fill_(2,41,259,184)` | `sub_B8BC7` |
        | 守方面板外框 / 暗化區 | 貼圖 (378,40);`Darken_Fill_(379,41,638,184)` | `sub_B8C8B` |
        | 面板文字 x(置中)| 攻方 130 / 守方 508;列高 11,首列 y=50 | 同上 + `sub_1210FD` = `Print_Centered_` |
        | 兵種欄位 x | `261 / (兵種數+1) × (序號+1) + 基準X` | `Print_Troop_Totals_` @ 0xB896D |
        | 部隊落點 | `X = 基準X + Random_(50) − 20`;`Y = min(360 + Random_(85), 430)` | `sub_B88B2` |
        | 兩側基準 X | 攻方 **50** / 守方 **590** | 常數 `dword_B6CDE = 0x024E0032` |
        做出來的第一版整個是錯的。`Print_Troop_Totals_` 裡的 `mov eax, 105h`(=261)與資產 21 的
        (`byte_19EB94 + side*0x3E8 + i*0x19`,欄位 +0 X / +2 Y / +5 狀態 / +6 兵種 / +7 側 /
        索引 = 種族×13 + 7/8/9/10)而不是 COLGCBT;`Replace_Colgcbt_Color_With_Player_Colors_`
        @ 0xB8EFB 說明 dump 出來士兵腳下那塊洋紅色**不是影子,是帝國旗色的佔位色**。
        ⚠ 誠實留白:①原版是逐幀動畫的即時戰鬥,remake 的 `ResolveGroundBattle` 一次算完,
        ③落點的 `Random_` 換成由單位序號推出的固定散布(範圍與原版一致,但可重現,截圖驗證需要)


    ### ~~載入遊戲視窗~~ → **已完成**(2026-08-07,`cmd/moo2/loadgame.go` + `internal/shell/saveslots.go`)。


        `LOADSAVE.LBX` 先前全 repo 零引用不是巧合——remake 根本只有**一個**存檔檔案,每回合覆寫,
        主選單的 Continue 與 Load Game 都是「讀那一個檔」,沒存檔時點下去靜默無反應。

        | 元素 | 真值 | 來源 |
        |---|---|---|
        | 背景 / LOAD 鈕 / CANCEL 鈕 | `game.lbx` 資產 20 / 21 / 22,調色盤取 `mainmenu.lbx` 資產 21 | openorion2 `LoadGameWindow` |
        | 視窗定位 | `x=(640−寬)/2`、`y=(480−高)/2` | 同上 |
        | LOAD 鈕 / CANCEL 鈕 | (37,337) / (171,338) 68×22 | `initWidgets` |
        | 存檔槽 | 10 格,第 i 格 (22, 22+31×i) 232×27;文字 (x+32, y+24+31×i) | `initWidgets` + `drawSlot` |
        | 槽位規格 | `SAVEGAME_SLOTS = 10`,**最後一格固定是自動存檔** | `mainmenu.h` + `drawSlot` |
        - **oracle issue #2 結案**:原版無存檔時 Continue / Load Game 是**灰階不可按**的
        - **主選單「名人堂」接錯的入口修正**:它先前被暫借給「研究選擇」畫面當調色盤鏈的示範入口,
          現在導向真正的最終得分畫面(`hiScore`,2026-08-07 已建)。
        途中發現的一個真 bug:**`PlayerName` 與 `FlagColor` 從來沒有進過存檔**——玩家在命名旗色
        `sessionSnapshot`(舊存檔解出零值,退回預設,不會壞)。
        儲存流程亦由同一個 `loadGameScreen` 依 mode 分流；單人／熱座圖示使用 `game.lbx`
        資產 23／24，槽內依原版維持名稱、星曆／時間兩行。原版 `.GAM` 由 `internal/save`
        唯讀匯入，後續寫入 remake JSON，不覆寫原始存檔。


    ### ~~儲存遊戲視窗 + 遊戲中的「遊戲」選單~~ → **已完成**(2026-08-07,


        `cmd/moo2/gamemenu.go`;儲存視窗併進 `loadgame.go`)。

        叫 `_Draw_Load_Save_Game_Popup_` @ 0x7F206,名字自己就說完了;差別只在說明列
        (`Set_Load_Game_Screen_Help_List_` @ 0x6F850 vs `Set_Save_Game_Screen_Help_List_` @ 0x6F865)
        順帶把 openorion2 的資產索引拿反組譯核了一次:`Load_Mainmenu_Load_Game_Popup_` @ 0x803D9
        `ASSET_LOAD_BACKGROUND`…`ASSET_LOAD_MODEM` 那一串。兩個獨立來源互相印證。
        遊戲選單視窗(原版 `Do_Main_Game_Popup_` / openorion2 `MainMenuWindow`)的座標:
        視窗 **(144, 25)** ——硬編不是置中;
        RETURN (151,307),精靈為 `game.lbx` 資產 1–6,背景資產 0,調色盤取 `buffer0.lbx` 資產 0。


    ### ~~`Colony_Bombing` 畫面~~ → **已完成**(2026-08-07,`cmd/moo2/bombing.go`)。


        反組譯挖到的:
        - `Draw_Colony_Bombing_Screen_` @ 0xB4800 標題 `Print_Centered_(319, 10)`——與地面戰
          同一個錨點,兩個畫面共用版面慣例。
        - `Do_Bomb_` @ 0xB4606:炸彈記錄**每筆 15 位元組**(+0 精靈 / +4 X / +6 Y / +8 幀 / +0x0E 啟用),
          每 tick 已啟用者幀 +1,**未啟用者 `Random_(5) == 1` 才啟用**——所以原版的炸彈是零零星星
          散著炸,不是整排同時落地。
        - `Add_Bomb_To_End_Of_Queue_` @ 0xB43A6:在 **49** 個目標槽上蓄水池抽樣挑下一個挨炸目標。
        - 背景是 `COLONY.LBX#8`(640×480,6 幀 delta)——殖民地的**建築格地面**,呼應那 49 個槽。

        `Replace_Colgcbt_Color_With_Player_Colors_` @ 0xB8EFB 專門做這件事。remake 照做了
        (`recolorPlayerRamp`),格線會跟著玩家選的旗色變。這同時解釋了先前 dump 地面戰士兵時
        ⚠ 一個資料層的事實,記下來免得後人重查:`Load_Bombing_Anims_` @ 0xB435A 載
        另外訂正了一個版面推論:原版的 `Print_Centered_` 第二個引數是文字的**上緣**不是中心


    ### ~~Smacker 過場~~ → **已完成**(2026-08-07,`internal/smk/` + `cmd/moo2/cutscene.go`)。


        先釐清前提:MOO2 的片頭與各結局過場**不是 LBX**,是**裸的 Smacker 檔**,只是沿用了
        `.LBX` 副檔名——`INTRO.LBX` 開頭四個位元組就是 `SMK2`(480×160、1407 幀、≈13 fps),
        `WININFIN` / `LOSERFIN` / `ORIONFIN` / `ANATKFIN` / `AMEBAFIN` / `PLNTDFIN` / `DIMTVFIN` /
        `ANWINFIN` 同理。所以這一項要的是一個 SMK2/SMK4 解碼器,不是找檔案。

        - **位元預算幾乎完全吻合**:1407 幀合計只多讀 721 bits(平均每幀 0.5 bit,就是最後一個
        **結局過場也接了**(2026-08-07,`internal/gamedata/cutscene.go`)。這一步的重點是
           `AMEBAFIN` 與 `PLNTDFIN` ← `Bomb_Results_Popups_` @ 0xE85F7 +
           `Do_Attacker_Beat_Colony_Stuff_` @ 0xE87D2;`DIMTVFIN` ← `Tactical_Combat_` @ 0x47939。
           其餘六個名字執行檔**完全沒有字面引用**,存在 `ESTRINGS.LBX` 的字串池裡
        接線:勝負分出 → 依 `Victory` 選過場(敗北 LOSERFIN / 安塔蘭勝 ANWINFIN / 其餘 GENWINFN)
        ⚠ **仍未定**:`WININFIN` 與 `GENWINFN` 都在結局群,但「哪一個對應哪一種勝利」沒有證據
        ——挑選它們的程式碼不在執行檔的字面引用裡。remake 一律用 `GENWINFN`(已由末幀確認是
        完整結局片),`WININFIN` 標為待定,不臆測。`ORIONFIN` / `ANATKFIN` 等事件動畫也還沒接
        (前者需要獵戶座星系,remake 還沒有)。清單見 `gamedata.UnmappedCutscenes`。


    ### ~~多人連線(獨立子專案)~~ → **熱座已完成**(2026-08-07,`internal/shell/hotseat.go` +


        `cmd/moo2/multiplayer.go` + `cmd/moo2/hotseat.go`);**網路 / 數據機 / 序列埠直連未做**。

        **原版的多人設定畫面整張版面都拿到了**(`Multi_Player_Screen_` @ 0xF4D99,初始化
        `sub_F42CA`、建 widget `sub_F009A`):背景 MULTIGM.LBX#0(640×480)、面板 #1(482×335,
        置中 `(0x280−w)/2` / `(0x1E0−h)/2` → (79,72));左欄四個連線方式 x +0x3B,
        y +0x5B/+0x7A/+0x9B/+0xBB;右欄四個動作 x +0x10D、同四列;CANCEL (+0xB0, +0x11E)。
        `Set_Multi_Player_Game_Type_` @ 0xF5691 寫進 `byte_199F3A` 的四個模式碼
        **連原版自己會隱藏的按鈕都照做**:`sub_F009A` 在選了 HOTSEAT 時把 JOIN GAME 的
        **熱座的席位模型**:原版帝國資料本來就是 `player[i]` 陣列(stride 0xEA9)+ 當前索引
        `word_19999C`,所以 `Save_Hotseat_Map_Info_` @ 0x88F5D **每席只存七個 word**(星圖視野)。
        `Get_Multi_Player_N_Humans_` @ 0x121F0 則是去數 `player[i]` 裡控制碼為 100 的帝國
        ——「幾個真人」不是獨立設定,是「有幾個帝國被標成真人」。remake 的 `GameSession` 是單數
        欄位不是陣列,改成 `player[i]` 要動幾乎每個畫面,故走**席位交換**(`internal/shell/hotseat.go`):
        玩家側欄位整組進 `seat`,換人時存回目前席位、載入下一席。語意與 `player[current]` 等價。
        ⚠ 誠實留白(全部寫在 `hotseat.go` 檔頭,此處摘要):
        ②交接畫面用原版的**尺寸與文字錨點**(`Draw_Hotseat_Screen_` @ 0x626D6:視窗置中、
        文字 +0x0E/+0x46)但**底圖是自繪的**——原版底圖來自 `dword_19B874` 指向的已載入影像,
        ③非當前席位的帝國在 `EndTurn` 最後才結算(當前席位在 AI 決策之前、其餘在之後),
        或全滅不會結束對局——要補得先讓勝負判定吃「哪一位玩家」而不是隱含的 `s.Player`;
        ⑤真人席位是從 AI 對手**接管**過來的。這一輪已補上 `RaceIndex` 與明確
        `SetupHotseatWithAIIndices`:熱座畫面會逐一勾選要接管的帝國,未選中的 AI 與 AI 關係矩陣
        保留,玩家間諜欄位依剩餘 AI 重排;席位轉換也保留種族產出/戰鬥加成、領袖、母星建築、
        艦隊與殖民地平行陣列。`AIOpponent` 仍比玩家側薄(沒有建造佇列、前哨站、傭兵池,
        也沒有可直接轉成玩家建造佇列的 AI 生產決策),這些欄位接管後維持空值,是目前明列的模型差異。

4. **地面戰解算**(`Resolve_Ground_Combat_` / `Ground_Combat_Round_`)→ 取代目前沿用一代 1oom 的借用結構。

### 第三梯:補完整性

5. ~~NEW GAME 設定畫面用估計座標、只有 3 個設定能選~~ → **2026-08-07 全部重挖**

    (`cmd/moo2/interactive.go` 的 `ngSettings` 檔頭有完整對照)。

    | 來源 | 給了什麼 |
    |---|---|
    | `sub_CCE2E`(建 widget) | 畫面原點 `word_1831D4`=X=15、`word_1831D6`=Y=5;五個設定框、五條數值列、三個開關、兩顆鈕的座標全是立即數 |
    | `sub_CCC3D`(畫值圖) | 3 欄 × 2 列迴圈:起點 (X+0x79, Y+0x77)、欄距 0x9B、列距 0x8C、右下格跳過 → 算出的欄 x = 121/276/431、列 y = 119/259,**與建 widget 的熱區 x1/y1 逐一相同** |
    | `NEWGAME.LBX` | 資產 1–22 剛好 22 張 65×65,而五個選擇器的選項數 3+5+4+7+3 = 22 |
    `word_1A1366` 由 `byte_199CB1 − 2` 得來,配 `NEWGAME.LBX` 13–19 那七張數字圖
    - **GALAXY AGE**:效果(光譜加權 `StarClassWeights` + 氣候骰表)其實早就實作完,
    - **難度補到五級**:選擇器選項數是 5、`NEWGAME.LBX` 4–8 是五張手勢圖
    沒有一手表之前不臆造。細節記在 `shell.TechLevels` 註解。


## 二、逐畫面重建與子系統補齊(2026-08-07)


    ### ~~種族選擇畫面「版面為合成近似」~~ → **2026-08-07 改用反組譯真值,順便修掉左右相反**


        (`cmd/moo2/raceselect.go` 檔頭有完整對照)。

        | 來源 | 給了什麼 |
        |---|---|
        | `Race_Selection_Screen_` @ 0x5C510 建鈕迴圈(i=0..13) | y = `0x5A + 0x30×(i mod 7)` = 90 + 48i';x = `0x15F + 0x7E×(i div 7)` = 351 / 477 |
        | `Draw_Race_Selection_Screen_` @ 0x5BD97 | 肖像 = RACESEL 資產 `15 + 索引`(`lea edx,[eax+0Fh]`),畫在 `(0x36, 0x3F)` = (54, 63);標題橫幅資產 33 @ (366, 52) |
        | `RACESEL.LBX` 資產表 | 1–14 各 **123×45** 2 幀 → 與步距 126×48 差 3px 間隙,正是按鈕;15–28 各 **290×322** → 畫在 (54,63) 佔 54..344,與按鈕欄 x=351 剛好不重疊 |


    ### ~~艦艇設計畫面用估計座標~~ → **2026-08-07 六個艦體槽 + 底部三鈕改用反組譯真值**


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
        總價那行。兩處都改成跟著 `dsHullY` 走。
        底部三顆鈕(`sub_1151B0`,引數是三個熱鍵字串 `aLb` / `+2` / `+4`):
    | 畫面 | 狀態 |
    |---|---|
    | 地面戰 | ✅ 2026-08-07(第一個實例,見第 3 項(Colony+Event 畫面))|
    | 軌道轟炸 | ✅ 2026-08-07(第 3 項(Colony+Event 畫面))|
    | NEW GAME 設定 | ✅ 2026-08-07(第 5 項(NEW GAME 設定畫面),順便修掉 PLAYERS 被當 RACE 用)|
    | 殖民地 | ✅ 2026-08-07(第 6 項(殖民地畫面框架),框架 = COLPUPS.LBX#5)|
    | 種族選擇 | ✅ 2026-08-07(第 5 項(NEW GAME 設定畫面),順便修掉左右擺反)|
    | 艦艇設計 | ✅ 2026-08-07(第 5 項(NEW GAME 設定畫面),六格不等距;右上兩個面板待追)|
    | 多人設定 / 熱座交接 | ✅ 2026-08-07(第 3 項(Colony+Event 畫面))|
    | menu / planets / research / fleet / officer / info | ✅ 2026-07-12(openorion2 `initWidgets` 真值)|
        建築編號 → 圖檔算式在 `Cache_Load_Bldg_` @ 0xAF6DC,見 `cmd/moo2/colonysurface.go`。

6. **殖民地畫面換上原版框架**(2026-08-07)

    (`cmd/moo2/colonyscreen.go` 檔頭有完整對照)。

    - **版面資料在執行檔裡,不在 LBX 裡。**`Add_Job_Field_For_` @ 0xBCB4B(由
      `Add_Job_Fields_` @ 0xBCC3D 迴圈 i=0..2 呼叫)直接給出職業欄座標:
      `ecx = 0x136`(310)、`push 0x1FE`(510)→ x 310..510;
    - **框架美術也在,只是不在 COLONY.LBX。**是 **COLPUPS.LBX 資產 5**(640×480,
    (`Make_Bldg_Array_For_Colony_` / `Bldg_Coords_To_Screen_Coord_` /
    `Sort_Bldg_Array_Columns_` / `Box_Bldg_Slot_` 那一整套,配 COLBLDG.LBX 的建築圖)。
    不是原版的**。原版的建造佇列本身是獨立彈出視窗(`Build_Queue_Popup_` @ 0xB4041,
    `Add_Build_Queue_Fields_` @ 0xB325A 給出 7 格:x 207..458、y 329+20i),座標已到手,

7. **TECH LEVEL 接上第一個 gameplay 效果:曲速前的 FTL 限制**(2026-08-07)。

    第 5 項(NEW GAME 設定畫面)留下的「TECH LEVEL 只存設定、不影響 gameplay」不再完全成立。手冊直引
    (已收在 `docs/tech/homeworld-init.md`):

    `FTLTopic` 後解除。`FTLTopic = TOPIC_NUCLEAR_FISSION`,因為 remake 科技樹裡
    `TECH_NUCLEAR_DRIVE`(MOO2 的入門 FTL 引擎)就在那一列(techtree.go 第 55 列,
    Cost 50、ResearchAll),**不在開局就給的 `TOPIC_STARTING_TECH` 裡**。
    ⚠ **踩到一個真的零值陷阱,值得記下來**:`TechLevel` 的 Go 零值是 0 = 曲速前。
    接上限制的當下 `TestFleetInterstellarMovement` 立刻紅燈——`NewDemoSession` 從沒設過
    修法與 `GalaxyAgeSet` 同款:加 `TechLevelSet` 標記,未設過一律退回「一般」,
    並補一條專門盯這件事的回歸測試(`TestTechLevelZeroValueDoesNotFreezeFleet`)。
      要先有「依人口 + `initial_buildings` 優先序生成」的機制,而那張優先序表還沒到手。

8. **行星表面 + 建築 sprite 擺放:先查清楚為什麼不能照上面那套方法做**(2026-08-07 調查)。

    這是殖民地畫面中段仍留白的那塊。追下去發現它**不是幾個立即數**,而是查表:

    - `CR_To_XY_` @ 0xBC5D8:`word_182C9C[56a + 8b + 4c]` — 一張四維 word 表
    - `Bldg_Coords_To_Centered_Screen_Coord_` @ 0xBC866:`dword_182E24[...]` 與
    從執行檔的資料段抽出來、驗證維度與語意,再配 `Make_Bldg_Array_For_Colony_` /
    `Sort_Bldg_Array_Columns_` / `Box_Bldg_Slot_` 的排序與命中邏輯,以及 COLBLDG.LBX 的


    ### **行星表面格點:表抽出來了,第 8 項(行星表面調查)的「獨立工程」判斷只對了一半**(2026-08-07 續)。


        第 8 項(行星表面調查)的結論是「格點→螢幕座標是烘在資料段的幾何表,沒有公式可抄」——**這句是對的**。
        但它接著把整件事記成阻塞,那一步錯了:**表就在反組譯的資料段裡,讀出來就是了**。
        發現「這是查表不是公式」時,下一步是去讀那張表,不是把它列為待辦。

        | 來源 | 內容 |
        |---|---|
        | `word_182C9C`(392 位元組)| **7×7 角點**,每格 8 位元組 =(x, y)。索引 `a×56 + b×8 + c×4`,c=0 取 x、c=1 取 y |
        | `dword_182E24` | `word_182C9C` 的**完整副本**(執行檔裡放了兩份,內容逐值相同)|
        | `dword_182E2C` | = `dword_182E24 + 8`,差一個 b 槽——所以中心點公式是 (角點[a][b] + 角點[a+1][b+1]) / 2 |
        | `word_182FAC` | 同一組角點去掉第一列第一欄的 6×6 版(步距 48)。有 7×7 就推得出來,remake 用不到 |
        | `dword_BA784`(72 位元組)| `Add_Bldg_Fields_` 走訪 36 格的順序,36 組 (a,b),**a+b 由大到小 = 遠→近**(畫家演算法)|
        `COLONY.LBX` 資產 8 是**已經畫好位置**的單格高亮菱形(640×480 稀疏圖)。渲染出來量到的
        回歸測試釘在 `cmd/moo2/colonysurface_test.go`。
        `Draw_Colony_Screen_` 一開場就是 `C_Anims(1, 0, 639, 479)` + `Draw(0,0,…)`:地表是
        `BLDG0..BLDG4.LBX` 每個資產都是 **640×480、已經畫好位置**的稀疏圖。實測 BLDG0 資產
        畫的時候直接貼 (0,0) 就對位。`Draw_Building_With_Bottom_Centered_` 那套底邊置中是原版
        同理:`COLROADS.LBX`(156 個 640×480)是道路、`COLVEGGI.LBX`(104 個小圖)是植被,
        對應 `Draw_Road_List` 與 `Build_Veggie_List_Based_On_Bldg_List_`。
        ### 建築編號 → 圖檔:`Cache_Load_Bldg_` @ 0xAF6DC 把整條算式寫死了
        `sub_BC8A6` 也是算式不是表——**蛇行**:`slot = b×6 + a`(b 偶數)/ `b×6 + 5 − a`(b 奇數),
        | 來源 | 內容 |
        |---|---|
        | openorion2 `src/gamestate.h` | `BUILDING_NONE = 0` 起的 `BUILDING_*` 列舉,48 棟 |
        | 原版 `TECHNAME.LBX` 資產 0 | 第 295 條起 "No Building"、"Alien Control Center"、"Armor Barracks"…"Artificial Planet",**逐條與列舉同序**(openorion2 `src/lang.h` 亦寫 `TNAME_BUILDING_NONE 295`)|
        remake 的 40 棟對照表寫在 `cmd/moo2/colonysurface.go` 的 `origBuildingID`,
        - **哪一格放哪棟**的規則(`Make_Bldg_Array_For_Colony_` / `Sort_Bldg_Array_Columns_` /
          `Insert_Bldg_Into_Array_` / `Find_Replacement_Slot_For_Building_`)。
        - **陰影**:建築圖裡最大宗的顏色是 (108,48,108) 的洋紅,那是原版用來和地表混色的
        - **地表底圖**本身:`C_Anims` 的動畫清單是依行星氣候載入的,還沒追到來源 LBX。
          `Build_Queue_Popup_` @ 0xB4041(7 格 x 207..458、y 329+20i,座標已到手)。
        落地的部分:`cmd/moo2/colonysurface.go`(角點表 + 走訪順序 + 格四角/中心/命中測試 +
        放在中段,而原版那裡是地表、佇列是獨立彈出視窗 `Build_Queue_Popup_` @ 0xB4041


    ### **原版建築表挖出來:所有「估計的」建造成本換成真值**(2026-08-07)。


        第 8 項(行星表面調查)在追建築 sprite 時,`Real_Building_Name_` @ 0xBB40D 只有兩行——
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
          (`MANUAL_150.html` modding 範例)。
          `Make_Bldg_Array_For_Colony_` 就是用這個欄位把衛星排除在地表格點外。
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
        - `Make_Bldg_Array_For_Colony_` @ 0xBC30B:`Set_Random_Seed(colonyIdx, 0, 144)` →
          **同一個殖民地的擺法是固定的**;依編號順序逐棟 `Insert_Bldg_Into_Array`,
          分類 7 的丟去衛星清單;再依 `人口/3 + 1` 補「房屋」;最後 `Sort_Bldg_Array_Columns_`
        - `Insert_Bldg_Into_Array_` @ 0xBC05E:對所有空格做**蓄水池抽樣**(`Random(++n) == 1`),
          滿了才叫 `Find_Replacement_Slot_For_Building_` 挑最低優先的格子擠掉。
        - **房屋是借用衛星的編號畫的**:`-3` 依 `(colonyIdx + 房屋數 + 1) % 4` 存成
        - **卡住的點**:擺法要完全重現得先實作原版的 `Random_` @ 0x1247A0 /
          `Set_Random_Seed_` @ 0x124820。那是下一步,不是阻塞。


    ### **半透明標記索引:index >= 0xF0 從來不是顏色**(2026-08-07)。


        第 8 項(行星表面調查)留下的「陰影是洋紅」問題,追下去發現不只是陰影,而是 remake 一直誤解了
        一整類像素。原版的通用繪圖常式(module 168)對 **來源索引 >= 0xF0 的像素從不直接寫進
        畫面**,兩條路徑各有做法,由 `Draw_Bldg_CR_` @ 0xBB469 依 `byte_182ACA` 二選一:

        | 路徑 | 內圈 | 對 >= 0xF0 的處理 |
        |---|---|---|
        | `Draw_` @ 0x12A478 | `sub_12ACA4` | `dst = blendTable[(src << 8) + dst]` —— 混色查表 |
        | `Draw_No_Glass_` @ 0x129FF9 | `sub_12AAA1` | `cmp eax, 0F0h / jge` → **整個像素跳過** |
        | LBX | >= 0xF0 的像素占比 |
        |---|---:|
        | SPHERSFX | 100% |
        | BEAMS | 69% |
        | LOGO | 39% |
        | CMBTSFX | 35% |
        | COLONY | 27% |
        | MAINMENU | 13% / RACESEL | 13% |
        | BLDG0..4 | 7–11% |
        `internal/lbx/image.go` 加了 `TranslucentIndexMin = 0xF0`、`Frame.HasTranslucent()`
        與 `Frame.ToRGBADropTranslucent()`(= `Draw_No_Glass_` 那條路徑,底下沒東西可混色時
        這是原版唯一適用的那條,不是近似)。既有的 `ToRGBA` 行為不變,不動到現有畫面。
        **混色表 `byte_1AB358` 在 BSS,執行期才建,不在執行檔裡**;產生它的程式碼還沒找到
        `galaxy.cpp` 留著 `FIXME: analyze original game and calculate better shadow palette`。


    ### **中段還給行星表面,建造佇列搬回原版的彈出視窗**(2026-08-07,第 8/8/8/8 項的收尾)。


        第 8 項(行星表面調查)推翻的那句「中段(y 159..423)原版畫行星表面」其實還有更深一層:remake 把
        **建造佇列與可建清單**塞在中段,那是 remake 自己的版面;原版那裡是地表,佇列是另外
        一張彈出視窗。這一項把兩邊都搬回原位。

        ### 建造彈出視窗(`Build_Queue_Popup_` @ 0xB4041 → `cmd/moo2/buildqueue.go`)
        烘著六顆鈕)。`Draw_Build_Queue_Popup_` @ 0xB3CF7 用 `sub_12A478(0, 0, img)` 貼在
        | 元件 | 來源 | 座標 |
        |---|---|---|
        | 可建清單 | `Add_Buildings_Fields_` @ 0xB08CA | x 13..184、y = 20 + 19i(列高 19)|
        | 佇列 7 格 | `Add_Build_Queue_Fields_` @ 0xB325A | x 207..458、y = 329 + 20i(高 21)|
        | Auto Build | `sub_11523B`(toggle)| (490, 342) |
        | REFIT / DESIGN | `sub_1151B0` | (492, 379) / (561, 379) |
        | REPEAT BUILD | `sub_1151B0` | (503, 411) |
        | CANCEL / OK | `sub_1151B0` | (493, 447) / (560, 447) |


    ### **原版 PRNG 照抄 + 建築擺放換成原版演算法 + 地表底圖接上**(2026-08-07,第 8 項(行星表面調查)的續)。


        ### 原版 PRNG(`Random_` @ 0x1247A0 → `internal/gamedata/origrand.go`)

        32-bit LCG,乘數 `0x41C64E6D`、增量 `0x3039`(就是 ANSI C `rand()` 那組常數,
        但**保留完整 32-bit state**,沒有 `>>16` 也沒有 `& 0x7FFFFFFF`),外加**拒絕取樣**:
        `bucket = 0xFFFFFFFF / n`、`limit = bucket × n`,抽到 `>= limit` 就重抽,
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
        定案證據在 `Add_Bldg_Fields_` @ 0xBE44A,同一對 (v1, v2) 同時餵給兩邊:
        ```
        colony_bldgs[24×v1 + 4×v2]                        ; 格陣元素 → v1 是「列」
        Bldg_Coords_To_Centered_Screen_Coord(v1, v2, c)   ; 螢幕座標 → 索引 v1×56 + v2×8 + c×4
        ```
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
        赤紅熔岩裂縫(Radiated)、#0..2 是有毒星的三種面貌。remake 的 `gamedata.PlanetClimate`
        原版 `Draw_Colony_Screen_` 的次序是
        地表兩層 → `Draw_Colony_Info_Background`(框架)→ `Draw_Colony_Bldgs` → 資訊面板。
        ### 軌道衛星(`Draw_Colony_Satellites_` @ 0xBE366,同日補上)
        分類 7 的建築不進地表格點,`Make_Bldg_Array_For_Colony_` 把它們丟進 `word_19F99C`
        ```
        x = 295 + (i 偶數 ? +1 : −1) × i × 50     ; 295 = 0x127、50 = 0x32
        y = 162                                   ; 0xA2,固定
        ```
        圖檔來自 `sub_BE306` 的比較鏈,但**編號要經過 `sub_BBB9F` 的 `loc_BBBAF: add edx, 9`**:
        | 建築 | 編號 | `sub_BE306` | COLONY.LBX 資產 |
        |---|---|---|---|
        | 星際要塞 Star Fortress | 41 | 0 | 9 |
        | 戰鬥站 Battlestation | 8 | 1 | 10 |
        | 星基 Star Base | 40 | 2 | 11 |
        | 次元傳送門 Dimensional Portal | 14 | 3 | 12 |
        | 天網 Artemis System Net | 3 | 7 | 16 |
        (offset 表開頭六個值都是 0x800),解出來是空圖 —— 畫面上什麼都不會出現,而且**不會報錯**。
        資產 9..16 全是 57×70,尺寸自洽;測試 `TestOrigSatelliteAssetsCoverEveryCategory7`
        抑制規則 `sub_BC21B`(回傳 1 = 這顆不畫)就是原版的**星基升級鏈**:
        星基(40)有戰鬥站(colony+0x13E)或星際要塞(colony+0x15F)就不畫;
        戰鬥站(8)有星際要塞就不畫。`0x136 + id` 是「這個殖民地有沒有這棟」的旗標陣列,
        0x136+8 = 0x13E、0x136+41 = 0x15F,兩個位移都對得起來。
        - **擺放的最後一段微調**沒做(見上表)。影響房屋彼此的相對位置,**以及道路**——
        - ~~**道路** `Draw_Road_List` 沒畫。~~ 第 14 項(地表道路)已補。
        - 衛星的**選取態閃爍**(`sub_BE271` 那條分支)沒做——remake 沒有點選衛星的互動。

9. **星圖:艦隊圖示換成原版的、旗色順序修正、拿掉浮在星圖上的研究文字**(2026-08-07)。

    ### 先把星圖的圖層順序抄下來(`Draw_Main_Main_Screen_` @ 0x8440E)

    | # | 圖層 | 位址 | remake |
    |---|---|---|---|
    | 1 | `Draw_Wormhole_Links_` | 0x85593 | ✅ 資料模型 + 連線 + 航行機制(第 10 項(蟲洞))|
    | 2 | `Draw_Relocation_Links_` | 0x85320 | ❌ |
    | 3 | `Draw_Stars_` → 逐星 `Draw_A_Star_` | 0x85550 / 0x83B02 | ✅ sprite 已接(第 9 項(星圖艦隊圖示));閃爍動畫未做 |
    | 4 | `Draw_A_Gate_Icon_`(迴圈) | 0x83741 | ❌ |
    | 5 | `Print_Star_Names_` | 0x88CB7 | ⚠ 位置已對齊(第 9 項(星圖艦隊圖示));字型樣式/描邊未做 |
    | 6 | `Draw_Black_Holes_` | 0x83BF9 | ❌ |
    | 7 | `Draw_Ship_Icons_` | 0xA070F | ✅ 本項 |
    | 8 | `Print_Main_Screen_Data_` | 0x87BAE | ⚠ 簡化(右欄五格數字)|
    | 9 | `Draw_Diplomacy_Request_Lights_` | 0x83D06 | ❌ |
    外層 `Draw_Main_Screen_` 另有 `Draw_Nebulae_` @ 0x84F8F 與 `Draw_Paralax_` @ 0x8500F。
    ### 艦隊圖示(`Get_Ship_Icon_Pict_Seg_` @ 0xA0D78)
    ```
    帝國艦隊(id 0..7):BUFFER0.LBX 資產 = 0xCD(205) + 旗色×4 + 縮放
    id 8              :              = 0xED(237) + 縮放
    id 9..14          :              = 0xF1(241) + (id−9)×4 + 縮放
    ```
    12×10 / 16×12,由小到大)。縮放由 `sub_79917` 給,原版再反過來映射(3→0、2→1、1→2、0→3)
    所以固定用縮放 0。每張圖有 8 幀(`Cycle_Ship_Icons_` @ 0x82DFF 在跑動畫),remake 只取第 0 幀。
    ⚠ 位置仍用 remake 自己算的星座標,不是 `Get_Ship_Icon_Coords_` @ 0xA0A5C。
    原版那支對 `word_1906C6` 做六路分支(艦隊在星系內的停泊槽 / 航行中的插值),
    | 索引 | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
    |---|---|---|---|---|---|---|---|---|
    | BUFFER0 205/209/…/233 量到的代表色 | (192,101,96) | (193,173,28) | (81,155,61) | (190,190,198) | (143,184,216) | (198,154,111) | (196,141,193) | (223,139,45) |
    | openorion2 `gfx.h` `FONT_COLOR_PLAYER_*` | RED | YELLOW | GREEN | **SILVER** | BLUE | BROWN | PURPLE | ORANGE |
    不是 White——量到的 (190,190,198) 也確實偏灰。護欄:`TestFlagColorsMatchOriginalOrder`。
    中文模式 17 張逐像素比對:變的是星圖、命名旗色、軌道轟炸三張(後兩張用 `FlagColors`),


    ### **星圖視差星空背景**(2026-08-07,第 9 項(星圖艦隊圖示)的續)。


        remake 的星圖區先前是一片 `RGBA{6,6,16}` 純黑,那是佔位。原版底下鋪的是三層星空。

        ### `Draw_Paralax_` @ 0x8500F
        ```
        sub_8FD71(edx=0x16, ebx=0x20F, ecx=0x1A5)   ; 裁切區 x 22..527、y ..421
        for layer in 0, 1, 2:
            img = STARBG.LBX 資產 layer            ; 實測皆 640×480
            Draw(x,       y      )
            Draw(x − 640, y      )                  ; 0x280 = 640
            Draw(x − 640, y − 480)                  ; 0x1E0 = 480
            Draw(x,       y − 480)
        ```
        | 層 | x | y |
        |---|---|---|
        | 0 | `word_199980` | `word_19997E` |
        | 1 | `word_19998A` | `word_19998E` |
        | 2 | `word_199984` | `word_199986` |
        原版也是先 `Fill` 再貼視差層(`Draw_Main_Screen_Filled_`)。
        沒辦法寫成單元測試,只能靠截圖廊第 04 張驗收。規則寫在 `starBGFill` 的註解裡。
        `decodeAsset` 對 nil `*assets.Resolver` 會 nil 解參考。畫面層的降級路徑本來就該容許
        「資料夾不完整」,所以 `starBGImage` / `shipIconImage` / `colonyScreenImage` 三個都補了
        - **星雲** `Draw_Nebulae_` @ 0x84F8F:每個星雲 `x = scale(nebula.x − 捲動) + 21`、


    ### **星圖:星球換成原版 sprite**(2026-08-07,第 9 項(星圖艦隊圖示)的續)。


        remake 先前用 `vector.DrawFilledCircle` 畫色圓,那是佔位。

        | 來源 | 內容 |
        |---|---|
        | 反組譯 `Get_Star_Picture_Seg_` @ 0x81CD3 | `edx = al×6`;`al == 6` 時 `eax = 縮放`,否則 `eax = 縮放 + bl`;資產 = `0x94 + eax + edx` |
        | openorion2 `galaxy.cpp:1664` | `_starimg[s->spectralClass][_zoom + s->size]`,`ASSET_GALAXY_STAR_IMAGES 148`、`STAR_TYPE_COUNT 6`、`GALAXY_STAR_SIZES 6` |
        | 實測 BUFFER0.LBX | 148..183 正好 6 組 × 6 張,每組尺寸 33/29/25/23/21/17(遞減),各 5 幀 |
        列舉 remake 本來就對上:`Star.Spectral` 0=藍…6=黑洞 = openorion2 `SpectralClass`;
        `Star.Size` 0=大..3=小 = `StarSize{Large, Medium, Small, Tiny}`。
        `Map_Scale_To_Zoom_Level_` @ 0x79917 把地圖比例(`word_199992` = 10/15/20/30,
        即 openorion2 的 `galaxySizeFactors`)換成縮放 0..3。openorion2 `galaxy.cpp:1385`
        (`transformX`),sizeFactor 越大距離越短,所以 **0 = 最放大、3 = 最縮小**。
        - 黑洞在原版是 `Draw_Black_Holes_` @ 0x83BF9 的**獨立迴圈**(黑洞是星球以外的地圖物件),
          remake 把它當光譜 6 的星球一起畫。圖是對的,但**阻擋航線**那套(`Star.blackHoleBlocks`)沒有。
        ### 星名位置(`Print_A_Star_Name_` @ 0x87768,併入第 9 項(星圖艦隊圖示))
        ```
        x = 星球中心 − 字寬/2          ; sub_12066F 量寬再減半
        y = 星球中心 + sprite 邊長/2 − 大小
        夾擠:x >= 0x16(22)、x + 字寬 <= 0x20F(527)   ← 與視差層的裁切區同一組數字
        ```
        ⚠ 顏色不是依擁有者:`sub_7A440` 的真名是 `Zoom_Level_Font_Style_`,選的是**縮放對應的

10. **蟲洞:把星圖第 1 層卡住的資料模型補上**(2026-08-07)。

    第 9/9 項留下的結論是「剩下的層卡的是資料模型不是繪圖」。這一項就是去補模型——
    蟲洞是其中最有遊戲價值的一個(它是**機制**不是裝飾:把銀河兩頭接起來)。

    | | 是什麼 | remake |
    |---|---|---|
    | **隨機事件** | 一次性好事,把正在航行的艦隊直接送到(手冊 p.181「moves that fleet to their destination in a single turn」)| 早就有(`applyWormhole`)|
    | **星圖蟲洞** | **永久**連在星圖上的兩星捷徑,任何時候都能走 | 本項新增 |
    ("One-way wormholes not allowed"),原版 `Draw_Wormhole_Links_` 也是兩端各畫一次。
    (星圖畫滿放射狀連線、艦隊到處一回合直達)。讀檔一律走 `normalizeWormholes`,
    `Generate_Wormhole_Links_` @ 0x8CC15 + `Set_Wormhole_Id_` @ 0x8D6D6:
    `_n_wormholes`(`byte_182245`)**不是常數**——它在銀河產生過程中逐星累加
    (`sub_8C840`),上限 `galaxySizeParam × 4 + 4`。要忠實重現得連整個銀河產生器一起搬,
    - `SendFleet`:兩端有蟲洞 → ETA 固定 1,不看距離。

11. **原版 48 棟建築,remake 到底缺哪幾棟?——把 8 個沒建模的編號查清楚**(2026-08-07)。

    `gamedata.Buildings` 是 40 項(手冊《The Big List》35 建築 + 5 衛星),而原版建築表是 48 棟。
    先前沒人把差集列出來過,所以「是不是缺了什麼」一直沒有答案。

    差集就是 `cmd/moo2/colonysurface.go` 的 `origBuildingID` 對不到的那 8 個編號。
    用 openorion2 `gamestate.h` 的 `BUILDING_*` 列舉認名字,再從原版建築表
    (`off_17EB3D`,19 位元組一列)讀真值:
    | id | 名稱 | 成本 PP | 維護 BC | 分類 | remake 現況 |
    |---:|---|---:|---:|---:|---|
    | 9 | Capitol | 200 | 0 | 1 | 非統一政體開局首都自帶；失守後可在指定首都以 200 PP 重建 |
    | 11 | Colony Base | 200 | 0 | 0 | 同上(拓殖時自動) |
    | 17 | Gaia Transformation | 500 | 0 | 1 | ✅ `SpecialActions`(一次性) |
    | 18 | **Galactic Currency Exchange** | 250 | 3 | 5 | ❌ **完全沒有** |
    | 37 | Soil Enrichment | 120 | 0 | 0 | ✅ `SpecialActions` |
    | 42 | Stellar Converter(行星版) | 1000 | 6 | 0 | ✅ 2026-08-07 補上,見第 11 項(48棟建築盤點) |
    | 44 | Terraforming | 250 | 0 | 0 | ✅ `SpecialActions` |
    | 48 | **Artificial Planet** | 800 | 0 | 0 | ❌ **完全沒有** |
    - 三個 `SpecialActions` 的成本(250 / 500 / 120)與這次重抽**逐項相同** —— 交叉驗證通過。
    - **維護費 0 = 一次性**:Terraforming / Gaia / Soil Enrichment / Artificial Planet 全是 0,
      這條規則本身就能把 8 個編號分成「該進 `Buildings`」與「該進 `SpecialActions`」兩堆。
    `TOPIC_GALACTIC_ECONOMICS`(techtree.go,**6000 RP**)解鎖的是
    `TECH_GALACTIC_CURRENCY_EXCHANGE`,但 remake **沒有任何東西消費它**——
    | 來源 | 結果 |
    |---|---|
    | 手冊《The Big List》(GAME_MANUAL.pdf)| **沒有這一條**——所以 remake 的 40 項表本來就抄不到它 |
    | patch 1.5 的 `MANUAL_150.html` | 零命中 |
    | 遊戲資料檔 | 掃過整個資料夾,只有 `HELP.LBX` / `TECHNAME.LBX` 有**名字**,沒有任何說明文字 |
    | 反組譯 | 建築表只有 19 位元組(名稱指標/編號/前置/成本/維護/分類),**沒有效果欄**——效果是寫死在程式碼裡的 |
    下一輪要補的話,路線是從 `colony + 0x136 + 18` 那個「已建」旗標的讀取點反追到收入計算,


    ### **恆星轉換器(行星版)補上 + 一個乾淨的負面結果**(2026-08-07,第 11 項(48棟建築盤點)的續)。


        第 11 項(48棟建築盤點)列出三棟真的缺的建築,並寫下「下一輪的路線是從 `colony + 0x136 + id` 的旗標
        讀取點反追」。這一項就是去走那條路——結果有一棟做成了,兩棟得到了**為什麼做不成的證據**。

        直接搜「有沒有讀 `colony + 0x148`(建築 18 的已建旗標)」會得到零命中,但零命中本身
        - 43 個位移在反組譯裡出現過 → 技術本身有效(`sub_BC21B` 讀 `+0x13E` / `+0x15F` 就是實例)
        - **完全沒出現過的只有 5 個**:Artemis System Net(3)、Dimensional Portal(14)、
        這 5 個自己解釋了自己:兩個是**衛星**(由 `Draw_Colony_Satellites_` 另外處理),
        ⚠ 至於建築 18:`+0x148]` 全檔只出現一次,而那一次的基底是 `dword_1AA21C`(科技已知陣列)
        | 來源 | 內容 |
        |---|---|
        | 手冊 p.106(`02-buildings.md` §十一)| 行星駐防版對目標造成 **400 傷 ×2**(共 1600),無視射程與防禦;維護 **6** |
        | 原版建築表第 42 列 | 成本 **1000 PP**、維護 **6**、分類 0(地表建築)、前置 Temporal Physics |
        維護費兩邊逐項相符,才敢動。接進 `colonyDefense`(+800)、`origBuildingID`(42,
        `TestBuildingsCount` 原本釘死 **40**,理由寫「手冊全表 35 建築 + 5 衛星」。
        - **Galactic Currency Exchange**(18):效果四條路都查不到(第 11 項(48棟建築盤點)),而這次的位移掃描
          要追 `Colony_BC_Production_` @ 0xE03F1 的快取欄位——那一支裡沒有任何建築旗標讀取,
        - **Artificial Planet**(48):旗標從沒被讀 + 維護費 0 → 確定是**一次性**,
          該進 `SpecialActions` 而不是 `Buildings`。但它的效果(把小行星帶變成行星)需要


    ### **飛彈基地/地面砲台其實早就算得出來——是接線漏了,不是模型缺了**(2026-08-07,第 11 項(48棟建築盤點)的訂正)。


        飛彈基地與地面砲台的空間預算模型**早就在**,第 11 項(48棟建築盤點)把它當成缺口是漏查:

        - `gamedata/satellite.go`:`MissileBaseSpace = 300`(手冊 p.78 確認值)、
          `GroundBatterySpace = 450`(p.81 確認值)、`SatelliteWeaponFitCount`、
          `SatelliteBeamSpaceWithArc`、`SatelliteStrengthScale`
        - `shell.retaliationAttackers` @ `orbital_bombardment.go`:已經用它算軌道轟炸的反擊,
        真正的缺口是 **`colonyDefense`(AI 突襲的防禦解算)沒接上這套**,它用的是
        | 路徑 | 星基值多少 |
        |---|---|
        | `colonyDefense`(舊)| `1 × 10` = **10** —— 比一艘巡洋艦(`shipStrength` 8)還強 |
        | `retaliationAttackers` | 依 space 預算推導 = **3–4** ≈ 驅逐艦 tier |
        而 `satellite.go` 的校準註解**明講**星基 ≈ 驅逐艦 tier(4)、戰鬥站 ≈ 巡洋艦 tier(8)。
        因為沒有任何測試同時看這兩條路徑。現在有了(`TestColonyDefenceUsesSpaceBudgetModel`)。
        `colonyDefense` 改用 `retaliationAttackers` 的 atk 加總,三件事一起對上:
        ③ 1.3/1.5 的 beam arc-cost 差異(`RuleProfile`)自動吃到,不必再各寫一份。
        `TestAIRaidRepelledByFleetAtStar` 有一段自我守衛(「測試前提不成立:願打門檻 X 已高於

12. **盤點:「最大的系統級缺口」那四條全部已經做掉了**(2026-08-07)。

    第 11 項(48棟建築盤點)的教訓是「**別編一個做不到的理由,先 grep**」。這一項把同一把尺對準文件自己:
    Part B 那份「最大的系統級缺口(A 級硬證)」清單,逐條核實之後**四條全中**——全部已完成。

    | 原本的斷言 | 一次 `ls` 就推翻 |
    |---|---|
    | 歷史記錄系統「remake 完全沒有」 | `internal/shell/history.go`(6 函式)+ `infoHistory` 折線圖 |
    | 前哨站「remake 只有殖民地」 | `internal/shell/outpost.go`(9 函式),進存檔、進熱座、可升級成殖民地 |
    | 艙損/維修 | `internal/shell/repair.go`(11 函式),手冊 p.80/82/25 + `Repair_Ships_At_Colonies_` 雙錨定 |
    | (事件系統早已標記完成) | — |
    Part A-2 的 `Smack`(Smacker 過場)同樣是過期的:`cmd/moo2/cutscene.go` + `internal/smk`
    - `cmd/moo2/diploview.go` 的 `diploRelationRows` 是寫死的 Klackon/Psilon/Silicoid 三筆。
      (`main.go` 的 `runDiploView`),遊戲內的外交走 `b.diplomacy()`。**不是 bug。**
    寫在 Part B 那一節的表裡(網路多人、`Command_Points` 專屬畫面、星圖 4 層、2 棟建築、

13. **指揮點數視窗建起來 + 一個「畫面自己打自己臉」的快取陳舊值**(2026-08-07)。

    第 12 項(四大系統缺口盤點)核實後的清單裡,`Command_Points` 專屬畫面是最小的一項。做掉它。

    ### 畫面結構(`Show_Command_Points_Screen_` @ 0x8BAB9,整支只有 30 行)
    ```
    sub_1191CA(&Draw_Mini_Main_Screen_, 1)      ; 背景重繪掛成「迷你星圖」
    sub_11438B(0, 0, 0x27F, 0x1DF, key=0x1B)    ; 整螢幕隱形欄位,ESC 關閉
    sub_128C32(0, 0, 0x27F, 0x1DF, 0)           ; Fill 清畫面
    Draw_Mini_Main_Screen_()                     ; 迷你星圖當底
    Show_Command_Points_(玩家索引)               ; → sub_E2644 包裝 sub_E2000 組文字
                                                  ; → loc_DDF24 尾端顯示
    ```
    文字本身在執行期才載入的字串區塊裡(`sub_DD4FD` 用 `repne scasb` 逐條走),英文原句沒解出來。
    | 符號 | 欄位 |
    |---|---|
    | `_starting_command_points_msg` | 起始指揮點數 |
    | `_total_command_points_msg` | 指揮點數總計 |
    | `_total_command_points_used_msg` / `_total_command_point_used_msg` | 已使用(原版連單複數都分兩條)|
    | `_command_summary_msg` | 總結 |
    | `_command_points_window_field` | 這個視窗的欄位 |
    2026-08-26 IDA 勘誤:`loc_DDF24` 其實是 `sub_DDEFB @ 0xDDEFB..0xDDF2C`
    內的尾端 call site,不是獨立泛用視窗函式;仍沒有「指揮點數專用」的座標證據。
    ⚠ ESC 那一半沒接:`shell.InputState` 目前只帶滑鼠,加鍵盤要動共用結構。
    第一版直接讀 `Player.CommandPointsSupply`,畫出來是:
    ```
    起始指揮點數    5
    軌道基地提供    0
    指揮點數總計    1     ← 5 + 0 = 1 ?
    ```
    原因:`Player.CommandPointsSupply` / `UsedCommandPoints` 是**只在 `EndTurn` 更新的快取欄位**,
    護欄 `commandpoints_live_test.go`:把快取欄位設成 −999 也要算得對、剛蓋好星基當下就要反映。

14. **殖民地地表的道路畫出來了,順手在原版資料裡撞到兩個位元組級的錯**(2026-08-07)。

    第 8 項(行星表面調查)留下的「道路沒畫」補完。`cmd/moo2/colonyroads.go`,護欄 `colonyroads_test.go`。

    由 `Load_Road_List_Anims_` @ 0xB5FBE 的跳過條件解出來:
    | dir | 合法範圍 | 幾何 | 段數 | 資產編號 |
    |---|---|---|---|---|
    | 0 | a 0..6、b 0..5 | (a,b)→(a,b+1) | 42 | `a×6 + b` |
    | 1 | a 0..5、b 0..6 | (a,b)→(a+1,b) | 42 | `42 + a×7 + b` |
    | 2 | a,b 0..5 | 格子的對角線 | 36 | `84 + a×6 + b` |
    | 3 | a,b 0..5 | 格子的另一條對角線 | 36 | `120 + a×6 + b` |
    42+42+36+36 = **156**,與 `COLROADS.LBX` 的資產數一模一樣 —— 這個等式就是幾何解對了的
    `Build_Road_List_Based_On_Bldg_List_` @ 0xB6099 由 `Make_Bldg_Array_For_Colony_` 在
    `loc_BC5B4` 呼叫,也就是**擺放 → 排序 → 抖動之後**。對每個有建築的格子抽三次 `Random(2)`:
    全執行檔對 `byte_19E57E` / `byte_19E57F`(dir 2、3 的旗標)**只有寫 0**,沒有任何一處寫 1
    (IDA 自己的 DATA XREF 也只記錄 `sub_B6099` 裡那兩個 `mov …, 0`)。所以 COLROADS.LBX 裡
    `sub_B6860` 也永遠選不到候選,一次亂數都不會消耗。remake 因此只實作有建築那條分支,
    - **繪製順序表 `byte_B4D5B` 少一個格點。** 49 組 (b,a) 應該不重不漏、依 a+b 由遠到近遞減。
    - **包圍判定表 `byte_B4DE9` / `byte_B4DF5` 的 Δa/Δb 對調了兩筆。** 要判斷格子 (a,b) 四條邊
    測試 `TestColonyRoadOrderMatchesOriginalTable` 把這件事釘死,並在失敗訊息裡說明為什麼不能改。
    分不清是原版有錯還是自己抄錯。改成**直接從 `Orion2.exe` 讀位元組**才定案:
    - 改用有把握的前 14 個位元組當錨點 → 全檔唯一命中,`0x13A3EF`。
    - 由此得 cseg01 的 VA→檔案位移 delta = `0x85694`,再用 `dword_B4DBD` 應為 `0x01000203`
    - **道路位置不會與原版同一局逐格相同。** 道路吃的是抖動之後的亂數流位置,而 remake
    - **植被層整層沒做(新發現)。** 道路之後原版還跑 `sub_B6977`:每個**空**格子最多放 2 株
      `COLVEGGI.LBX` 的植物,密度看周圍道路數(路少的地方草多),外觀與像素抖動都現抽。
      這是原版有、remake 沒有的一整層地表細節,前置是 `sub_B6647`(依氣候選植物)與
      `sub_BC866`(格子 → 螢幕座標)。


    ### **房屋抖動補完 + 母星的國會大廈一直沒畫**(2026-08-07,接第 14 項(地表道路))。


        第 14 項(地表道路)說「道路位置對不上原版,因為抖動沒補」。這一項把抖動補上,順帶發現一件更基本的事:
        **remake 的地表格陣從第一步就少放了一棟建築。**

        ### 抖動:`Make_Bldg_Array_For_Colony_` 的 `loc_BC441`..`loc_BC560`
        `houseStyles[n%4]`(表 `off_BA77C` = 3/14/40/41,與既有 `colonyHouseStyles` 逐項相同):
        反組譯的關鍵三行(已用 `objdump` 對原始位元組獨立驗過,IDA 清單無誤):
        ```
        eax = var_14 - si ; if (eax < 0) eax = 0        ; 第一個座標:有夾到 0
        edx = var_18 - di ; test edx, edx               ; 第二個座標:算了、測了號誌…
        cmp ax, 5 ; jle → edx = ax ; else edx = 5       ; …然後 edx 被無條件蓋掉
        ```
        - **第二個座標整個被丟棄**:對應的 `jge / xor edx, edx` 沒有被編出來,而下一條指令兩條
          路徑都覆寫 `edx`。於是目標格的兩個座標都等於 `clamp(a - si, 0, 5)` ——
        - **內圈的 `di` 因此完全沒有作用**,三次迭代做一模一樣的事。remake 仍保留那個迴圈:
          它決定 `Random` 被抽幾次,而道路接在同一條流上。
        `TestJitterColonyHousesTargetsDiagonalOnly` 專門防止有人把 `target := v*6 + v`
        ### `Get_Bldg_CR_` @ 0xBBD37:找一棟建築會吃亂數
        (後續那段交換由全域 `dword_182B19` gate 住,它初值 0 且唯一寫入點在 gate 內側,
        (`bldg0.lbx` 資產 `8×36 + 格`)、會被畫出來。**「不在建造表裡」與「不在地表上」是兩件事**,
        「**沒有** Capitol 的士氣懲罰可依政府設定」、「Colony Base 若加進 `initial_buildings`,


    ### **殖民地地表的植被層(第 14 項(地表道路)發現的缺口,同日補上)**(2026-08-07)。


        `cmd/moo2/colonyveggie.go`,護欄 `colonyveggie_test.go`。

        `Pick_Random_Veggie_Anim_Entry_For_Colony_CR_` @ 0xB6647:
        ```
        資產 = 群組×8 + max(Random(8) − 1 − (a+b)/2, 0)
        ```
        群組由氣候決定(`sub_B66D0`,一張 10 路跳表),最大群組 12 → 13 組。
        **13 × 8 = 104**,與 `COLVEGGI.LBX` 的資產數一模一樣(這次直接跑 `lbxinfo` 驗,不引用舊文件)。
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
        `sub_B6977` 只處理空格子,先數該格四條邊上有幾段路,然後:
        ```
        r = Random(7)
        if (r − 2) < 道路數  → 長
        else if 道路數 != 0  → 不長
        else                  → Random(建築數 + 2),回傳恆 ≥ 1 所以一定長
        ```
        ```
        x = 格心x + Random(寬) − 寬/2
        y = 格心y + Random(高) − 高/4      ← 是高/4 不是高/2,植物根部落在格心稍下方
        ```
        格心用 `Bldg_Coords_To_Centered_Screen_Coord_` @ 0xBC866 = 角落 (a,b) 與 (a+1,b+1) 取平均,
        **remake 的 `colonyCellCenter` 已經是同一個算式**,直接用。
        繪製在 `Draw_Colony_Bldgs_` @ 0xBEBDC 裡:對 36 格的遠→近順序,**每一格先畫植被再畫建築**,
        密度用的 `sub_B67E4` 就是第 14 項(地表道路)提到「Δa/Δb 對調了兩筆」的那張表。差別是:
        `TestColonyRoadEdges4UsesOriginalTable` 用「格子的四條真邊只會數到 2」把它釘住。
        - `sub_B6B95` 的外圈次數由呼叫端的 `bl = (沒有格子被選取)` 決定。正常畫面 `bl = 1` →
          跑一圈走 `sub_C5D55`;有格子被選取時 `bl = 0` → **一株都不畫**。
          remake 沒有「選取格子」這個狀態,固定走正常路徑;另一條 `sub_C5D75` 的差別沒有追。
          `dword_193174` 取,而道路的顏色是對的)。
        尺寸是在**產生**階段就要用的(它進位置公式),而地表是每幀重算的,而 `decodeAsset`
        自己沒有快取 —— 不處理就變成每幀重解最多 72 張 LBX。加了 `colVegSizeCache`。

15. **星雲:星圖 4 層裡的第 1 層,而且是有規則的地形不是裝飾**(2026-08-07)。

    `internal/shell/nebula.go`(規則)+ `cmd/moo2/nebula.go`(圖與判定),
    護欄 `nebula_test.go`。第 9 項(星圖艦隊圖示)列的星圖 4 層,做掉第 1 層。

    `Point_Is_In_Nebula_N_` @ 0xEB9C8 拿 `(點 − 星雲原點) / 3` 去索引星雲圖的**調色盤索引值**,
    每顆星的結果存在星球結構 +0x6F(`Initialize_Star_In_Nebula_Info_` @ 0xEBA96),
    也就是 `internal/save` 既有的 `Star.InNebula` 欄位。
    `Generate_Number_Of_Nebulas_` @ 0x8C4D3 是四路跳表:小 `Random(2)−1`、中 `Random(2)`、
    大 `Random(3)`、巨 `Random(3)+1`。上限 4 與 `internal/save` 從存檔格式反推的
    `gamedata.DamageHardShieldBonus`(手冊「額外減傷 3」)先前**沒有元件載體**,等於死碼。
    這一輪把「硬化護盾」加進 `SpecialOptions`,掛在與隱形裝置同一個研究主題
    (`TOPIC_DISTORTION_FIELDS`,techtree.go 的三選一,`TECH_HARD_SHIELDS`)。
    - **檔位換算**:第一版自己拿原版四檔的星數(20/36/54/71)取中點當界線。
      但 remake 的 `GalaxySizes`(12/24/36/48)**本身就是那四檔**,結果「中型」被判成檔位 0
      而那看起來完全合理。改成直接查 `GalaxySizes` 取索引,
      `TestGalaxySizeClassMapsGameOptions` 釘住。
    - **調色盤鏈**:第一版沿用殖民地畫面的鏈(含殖民地框架盤),整團星雲畫成**鮮紅色**。
    找這兩個都不是靠讀碼,是**加一行 `println` 量出來的**:第一次量到

16. **星圖移動換成秒差距模型:四條手冊規則從「無處可掛」變成可實作**(2026-08-07)。

    `internal/gamedata/starlane.go`(數值)+ `internal/shell/starlane.go`(接線),
    護欄 `starlane_test.go`。第 15 項(星雲)留下的移動懲罰,前置補完。

    | 規則 | 手冊 |
    |---|---|
    | 星雲 | reduced in speed to 1 parsec per turn |
    | 黑洞 | No ship can safely pass within 2 parsecs of a black hole (unless … Navigator) |
    | Navigator | increases the speed of the fleet by 1 or 2 parsecs per turn |
    | Warp Field Interdictor | radius of 3 full parsecs … slows all enemy ships … to 1 parsec per turn |
    **1 秒差距 = 30 個遊戲座標單位。** `Parsecs_Between_Points_` @ 0xEBE79 整支就是
    **四檔銀河尺寸。** `sub_1693B6` 的 4 路跳表逐檔寫死:
    | 檔位 | 遊戲單位 | 秒差距 | SizeFactor |
    |---|---|---|---|
    | 0 小 | 506 × 400 | 16.9 × 13.3 | 10 |
    | 1 中 | 759 × 600 | 25.3 × 20.0 | 15 |
    | 2 大 | 1012 × 800 | 33.7 × 26.7 | 20 |
    | 3 巨 | 1518 × 1200 | 50.6 × 40.0 | 30 |
    **星數 → 檔位。** `Galaxy_Size_From_N_Stars_` @ 0x798D2 的門檻是 **20 / 36 / 54 / 72**。
    remake 的 `GalaxySizes` 先前是 12/24/36/48(自訂),與原版四檔對不上。
    `FleetHasFTL` 對非曲速前開局**直接回 true、不看科技表**。於是那些開局的引擎階查出來是 0
    `TestFleetSpeedFallsBackToNuclearWhenFTL` 釘住。
    - **「穿越星雲」近似成起點或終點在星雲內。** 原版的艦隊沿路徑逐段前進,remake 是兩點直接
    - **黑洞 2 秒差距禁行**與 **Warp Field Interdictor 3 秒差距干擾場**的常數已入表
      (`BlackHoleAvoidParsecs` / `InterdictorRadiusParsecs`),但**還沒接進派遣判定**——


    ### **星圖航線模型:三條規則的共同前置補完**(2026-08-07)。


        `internal/shell/route.go`,護欄 `route_test.go`。第 16 項(秒差距模型)留下的三項全部接上。

        | 規則 | 判定 | 手冊 |
        |---|---|---|
        | 黑洞 | 線段到黑洞星 < 2 秒差距 → 拒絕派遣 | No ship can safely pass within 2 parsecs … unless … Navigator |
        | 干擾場 | 線段到敵方干擾器星 ≤ 3 秒差距 → 航速 1 | radius of 3 full parsecs … slows all enemy ships … to 1 parsec per turn |
        | 星雲 | 沿線取樣打到遮罩 → 航速 1 | Ships traveling **through** a nebula … |
        - **線段不是直線**:目的星**之外**的延長線上有黑洞,不該擋住這趟航程。
        - **起訖點豁免**:目的地本身是黑洞是玩家自己選的,擋的是「路過」。
        - **干擾場不給 Navigator 豁免**:手冊那句豁免只寫「nebulae and black holes」——
        - **星雲的判定式改成探針**(`SetNebulaProbe`):沿線取樣需要「任意點在不在星雲內」,
        | 星數 | 黑洞/局 | 被擋的目的地 | ETA(核融引擎) |
        |---:|---:|---:|---|
        | 20 | 0.5 | 5.7% | 2..10 回合 |
        | 36 | 1.1 | 7.9% | 1..15 |
        | 54 | 1.7 | 13.7% | 1..19 |
        | 72 | 2.3 | 11.3% | 2..30 |
        沿線取樣會呼叫遮罩判定上百次,而 `decodeAsset` 沒有快取 —— 每次派遣重解上百次 LBX。
        加了 `nebMaskCache`。**這是第三次踩同一個形狀的坑**(前兩次是殖民地地表每幀重算、
        植被尺寸每幀重解),共同成因是 `decodeAsset` 本身無快取,而呼叫端各自為政。


    ### **兩種星門:規則接進秒差距模型,標記上星圖**(2026-08-07)。


        星圖 4 層做掉第 3 層。`internal/gamedata/starlane.go` + `internal/shell/starlane.go`
        (規則)、`cmd/moo2/gate.go`(標記),護欄在 `starlane_test.go`。

        | 科技 | 效果 | 手冊 |
        |---|---|---|
        | 躍遷門 Jump Gate | 自己的殖民地之間 **+3 秒差距/回合** | increases the speed of your ships traveling between two of your colony systems by 3 parsecs a turn |
        | 星際之門 Star Gate | 自己的系統之間 **一回合到** | allows instantaneous (1 turn) travel between any two of your systems |
        - **躍遷門的加成放在懲罰之前**:星雲與干擾場都是「reduced **to** 1」——是覆寫不是相減,
        - **星際之門在最前面**:它形成的是穩定的蟲洞終端、不走實空間,所以沿路的星雲/干擾場
        原版 `Draw_A_Gate_Icon_` @ 0x83741 是一支 330 行的逐格動畫,資產來源不是字串常數、

17. **拓殖基地:國會大廈那個坑的另一半**(2026-08-07,task #46 的可實作部分)。

    第 14 項(地表道路)補上母星的國會大廈(編號 9)時,只治了一半。**編號 11 Colony Base 是完全對稱的
    另一半**:指定首都有國會大廈,其餘殖民地有拓殖基地。兩者都是佔一格、有美術、會被畫出的
    **實體建築**；國會大廈在失守後可於重新指定的首都重建。

    護欄 `TestColonySurfacePlanNonHomeworldHasColonyBase` 除了驗有無,還驗
    - **一次性改造(17 Gaia / 37 Soil Enrichment / 44 Terraforming)** 是否在完成後仍佔一格,
      取決於原版的 `byte[colony + 0x136 + id]` 旗標完成後是否保留。**沒有查證,所以沒做**——
    - **對原版實測落點**仍未做(需要 archive.org 線上原版逐畫面對照)。

18. **銀河貨幣交易所:它根本不是建築**(2026-08-07)。

    第 11 項(48棟建築盤點)留下的兩個「完全沒有」的編號,解掉一個。

    所以 remake 把它接在科技擁有狀況上(`PlayerState.HasGalacticCurrencyExchange`),
    「該進 `Buildings`」與「該進 `SpecialActions`」兩堆。編號 18 有 250 PP 成本、3 BC 維護,

19. **AI 主動請求會談 + 星圖上緣的請求燈**(2026-08-07)。

    星圖 4 層做掉第 4 層。`internal/shell/audience.go`(狀態)+ `cmd/moo2/audience.go`(燈),
    護欄 `audience_test.go` 兩支。

    `Humans_Requesting_Diplomacy_` @ 0xFA795 整支只有 `mov al, byte_1AB054; retn` ——
    **一個位元遮罩,每位對手一個 bit**。版面在 `Draw_Diplomacy_Request_Lights_`:
    ```
    x = 0x1FA − 已畫個數 × 圖寬      ; 506,由右往左
    y = 5                            ; 貼星圖上緣
    ```
    兩個都是立即數。`TestAudienceLightLayoutMatchesOriginal` 直接釘住。
    原版設那個 bit 的地方在 `sub_F5A9F` —— 一支約 30 路跳表的 AI 行動分派函式,觸發散在
    remake 改接在既有的 AI 模型上:**態勢改變時來敲門**。`ai.DecideStance` 的五級裡有三級
    第一版把來意直接寫成中文(「宣戰」「提議貿易」),被 `TestEnglishModeGapDoesNotGrow`
    **規則層不該吐顯示字串**。改成代碼(`war`/`trade`/`alliance`),顯示文字留在 UI 層。
    既有的 `stanceNames` 是中文,那是先前留下的;新欄位不再擴散這個作法。

20. **訂正:「手冊全文搜尋零命中」是假陰性——PDF 的連字騙了我**(2026-08-07)。

    第 18 項(銀河貨幣交易所)結尾寫了:

    `artificial` 實際上是 `arti` + `ﬁ`(U+FB01)+ `cial`,搜 `Artificial` 當然零命中。
    改搜小寫 `asteroid` 立刻命中,同一段就把規則講完了。
    `TECH_PLANET_CONSTRUCTION`(TOPIC_ADVANCED_MANUFACTURING 三選一)——手冊裡它緊接在
    `internal/shell/outpost.go` 的「尚未建模、誠實留白」那段裡有一條:
    `Planet.SystemBodies` 的欄位註解)。而人造行星按定義是「在已經有殖民地的星系裡**再多**
    一顆世界」——轉換 `SystemBody` 之後沒有地方能放第二個殖民地,做出來是空的。

21. **查證:一次性改造完成後不佔地表格子**(2026-08-07,task #46 的最後一個未查證項)。

    第 17 項(拓殖基地)留下的問題:「一次性改造(17 Gaia / 37 Soil Enrichment / 44 Terraforming)
    完成後是否仍佔地表一格,取決於原版 `byte[colony + 0x136 + id]` 旗標完成後是否保留。
    **沒有查證,所以沒做**。」

    全檔只有少數幾處,一眼就看到建築完工結算 `sub_13FD9` 那一處。
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
    `Make_Bldg_Array_For_Colony_` 那個讀旗標陣列的迴圈就不會擺它們。
    SpecialActions 不在 `gamedata.Buildings` 裡,`origBuildingID` 因此查不到它們——
    現況正確。加了 `TestColonySurfacePlanExcludesOneShotTransformations` 把這個「天然正確」

22. **黑洞的旋渦動畫:規則完整就做,不完整就不做**(2026-08-07)。

    星圖上的黑洞從第 9 項(星圖艦隊圖示)起就用對了圖(BUFFER0#187),但它是**靜止**的。
    原版不是——黑洞會轉。

    `Draw_Black_Holes_` @ 0x83BF9 對每個黑洞做兩件事:
    ```
    計數 = (_black_hole_anim_count[黑洞序號] + 1) % (幀數 × 2)
    幀號 = 計數 / 2
    ```
    那個「除以 2」不是單點讀出來的:一般星球的 `Draw_A_Star_` 裡也有同一個動作
    | 資產 | 尺寸 | 幀數 |
    |---|---|---|
    | 148..183(6 光譜 × 6 縮放的一般星球) | 33/29/25/23/21/17 | 各 **5** |
    | 184..187(黑洞,4 個縮放) | 40/34/34/25 | 各 **16** |
    `Draw_A_Star_` 的閃爍是**爆發式**的:計數器跑到 `star[+0x65]` 就停成 -1,
    另外還有一個全域併發預算 `word_19C164` 在管同時幾顆在閃。三個東西沒追出來——
    `starSpriteImage` 原本只解第 0 幀,快取 key 是 `lbx:資產`。一加上幀號,
    這已經是第四次撞到同一個根因:`decodeAsset` 沒有快取,於是每個呼叫端各自長一層
    (`colBldgCache`、`colVegSizeCache`、`nebMaskCache`,現在星圖幀也借 `colBldgCache`)。
    **下次再遇到就該去修 `decodeAsset` 本身**,不要再加第五層。
    - `cmd/moo2/starsprite.go`:`starSpriteFrame(資產, 幀)` + `starSpriteFrameCount`
    - `cmd/moo2/interactive.go`:`sceneBuilder.animTick`,由 `Update()` 每幀推進


    ### **訂正兩條:艦隊圖示不會動、一般星球也不會閃——都是原版行為,不是缺口**(2026-08-07,接第 22 項(黑洞動畫))。


        做完黑洞動畫,順手去追「艦隊圖示 8 幀為什麼沒動」。查完發現**要改的是文件不是程式**。

        原版有一支通用貼圖器 `sub_12A478`,整個 UI 都用它。動畫模型是這樣:
        | 位置 / 函式 | 意義 |
        |---|---|
        | `word[圖+4]` | 目前幀號 |
        | `word[圖+6]` | 幀數 |
        | `sub_12B726(圖)` | 幀號 ← 0 |
        | `sub_12B753(圖, n)` | 幀號 ← min(n, 幀數−1) |
        | `sub_12A478(x, y, 圖)` | 畫完**自動把幀號 +1**(`[圖+0Bh]` 的 0x20 位元控制循環) |
        ### ① `Cycle_Ship_Icons_` 不是動畫,是 F1/F2
        `Cycle_Ship_Icons_` 不是動畫,而艦隊圖示在原版也不會動:
        - `sub_82DFF` 由**鍵盤跳表**叫進來(`sub_825A8` 的 case −1001 等),`bx` 是方向
          (0 → `inc edx`、非 0 → `dec edx`),挑到的艦隊丟給 `sub_831B1` 選取。
        - `Draw_Ship_Icons_` @ 0xA070F 在**每次繪製前**都呼叫 `sub_12B726`(幀號歸零),
        `Draw_A_Star_` 的閃爍分支要 `star[+0x64] >= 0` 才會走。全檔查證:
        - 全域預算 `word_19C164` **只有歸零與遞減,全檔沒有任何遞增**。
        沒有新程式碼。`shipicon.go` 與 `starsprite.go` 的檔頭各改一段,測試註解跟著改——
        `TestStarSpriteFrameForOnlyAnimatesBlackHoles` 的意義從「釘住刻意的取捨」升級成
        | 鍵 | 作用 |
        |---|---|
        | F1 / F2 | 循環切換已知艦隊(一個方向 / 反方向) |
        | F5 / F6 | 系統視窗開著時,切到下一個 / 上一個已殖民星系 |
        | F9 | 測距:點第一顆星,再把游標移到另一顆,顯示兩者的秒差距 |
        | F10 | 快速存檔(沿用上次的存檔名,**會直接覆蓋**) |
        | ALT + F9 | 從星圖載入遊戲 |
        F9 特別值得做:remake 第 16 項(秒差距模型)已經把秒差距模型建好了(`ParsecsBetweenStars`),


    ### **手冊的星圖快捷鍵接上:F1/F2、F5/F6、F9**(2026-08-07,第 22 項(黑洞動畫)副產品的實作)。


        第 22 項(黑洞動畫)追 `Cycle_Ship_Icons_` 時把手冊的快捷鍵段落掃出來了。這一項把**行文中直接寫死**
        的那幾個接上(邊欄標籤的 ALT+Fn 那組仍然不碰,理由見第 22 項(黑洞動畫))。

        原版走的是逐艦隊的表;remake 的玩家艦隊是**單一集合**(`FleetAtStar`),AI 對手只有
        抽象的 `FleetStrength`、在星圖上沒有位置。所以循環集合現在只有一個元素——
        `TestKnownFleetStarsIsSingleForNow` 把這個限制釘成測試,多艦隊做出來時它會紅,
        1. **提示字釘在固定座標會壓到星星。** 第一版把「測距:移到另一顆星」畫在 (30,34),
        2. **截圖廊的示範終點寫死索引會踩到戰爭迷霧。** 第一版寫死「第 1 顆星」,而那顆還沒探索,
           `starAtScreen` 會跳過不可見的星 → 截圖停在提示上,什麼距離都沒畫。
        另外把星球的點擊熱區與懸停判定收斂到同一個 `starHitHalf`(22×22 方框)。
        - `internal/shell/starnav.go`:`KnownFleetStars` / `ColonizedStars` / `cycleStarList`
          + `CycleFleetStar` / `CycleColonizedStar`(純邏輯,環狀,清單空回 −1)
        - `internal/shell/input.go`:`InputState.Hotkey`(用字串不用 ebiten 的 key 型別,
        - `cmd/moo2/hotkeys.go`:按鍵對照 + F9 測距的畫面
        - `cmd/moo2/interactive.go`:`overlayScreen.onHotkey`(快捷鍵先於滑鼠處理)、
        - **F10 快速存檔**:手冊說它「沿用上次的存檔名、**直接覆蓋且不可回復**」。
          remake 的「上次的存檔名」就是 `sceneBuilder.savePath`——開局是自動存檔那一格,
          按下去成功與失敗看起來完全一樣。星圖既有的 `lastActionMsg` 畫在選中星面板裡
        - **ALT+F9 載入**:`loadGameInPlay()` 早就有(遊戲選單走得到),這裡只是多一個入口。

23. **多艦隊模型:把「全帝國只有一支艦隊」這個限制拆掉**(2026-08-07,第一階段)。

    remake 先前把玩家的兵力表示成一組欄位:`Ships` + `FleetAtStar` / `FleetDestStar` /
    `FleetETA` / `FleetMarines` / `FleetTanks`。**全帝國只有一支艦隊**,所有的船永遠在同一個
    地方、只能有一個航行任務。

    ### ⚠ 這次重構真正的難點:`Ships` 有兩種語意
    舊程式碼裡的 `s.Ships` 混著兩件事,而**單一艦隊時兩者剛好相同**,所以分不出來:
    | 語意 | 用在哪 |
    |---|---|
    | 「**這支艦隊**的船」 | 戰鬥、載運陸戰隊、消耗殖民船 / 前哨船、航行中損失 |
    | 「**全帝國**的船」 | 指揮點數(手冊 p.169 明文)、國力、艦名編號、外交評估、艦隊列表 |
    而那時候看起來完全正常——數字只是偏小。所以逐處分類,並用 `internal/shell/fleet_test.go`
    1. **修復**:先前只看「玩家選中的那一支」艦隊有沒有停在據點。改成**逐艦隊各自判定**
       ——這才對得上原版(`Repair_Ships_At_Colonies_` 的迴圈也是逐艦隊走的)。
    2. **母星防禦**:先前同樣只看選中那一支,於是玩家把視角切到別支艦隊,母星就「沒有防禦」。
    3. **隨機事件的「損失一艘艦」**:原版打的是整個帝國,不是玩家的操作焦點。
       改用 `removeShipGlobal`(跨艦隊索引)。
    新格式序列化整個 `Fleets`;讀到舊檔(`len(Fleets) == 0`)就從舊欄位組成唯一的一支。
    做遷移時發現:**舊格式從來沒有存戰車營**(有 `fleetMarines` 卻沒有對應的 `fleetTanks`)。
    新格式把整個 `Fleet` 序列化,這個洞順帶補上。
      (`AddShipToHomeFleet` 目前用「艦隊剛好停在那顆星就併進去」近似)


    ### **遷移連線:星圖 4 層裡的最後一層補上**(2026-08-07,第 23 項(多艦隊模型)的續)。


        原版 `Draw_Relocation_Links_` @ 0x85320 是主畫面圖層順序的第 2 層。它是 remake
        最後一個沒做的星圖層——先前卡在「艦隊是單一集合」,多艦隊(第 23 項(多艦隊模型))做完之後,
        缺的只剩這一層自己的資料。

        ```
        sub_784F0(星, 玩家) → word[星×0x71 + 0x54 + 玩家×2]     ; 遷移目標星
        sub_78C94(星, 玩家) → 上面那個欄位 != -1                  ; 「有沒有設定」
        ```
        **每個(星, 玩家)一個目標星索引**,−1 = 沒設定。`Draw_Relocation_Links_` 的迴圈就是:
        ```
        dword_81C80  dd 2 dup(70706F6Eh)     ; = 6E 6F 70 70 6E 6F 70 70
        ```
        `sub_A11C0` 另外收一個相位參數(`edi = 7 - 相位`,並依線的方向反轉表)——
        只被讀、沒被寫),remake 用自己的 `animTick` 驅動,步進沿用已被證實**三次**的
        1. **反鋸齒把線畫沒了。** 原版是逐像素寫調色盤索引、硬邊;remake 用向量畫線時開了
        2. **截圖廊的示範目標和 F9 測距撞在一起。** 兩條線走同一對星,遷移連線整條藏在
        `ColonyRelocateTo` 的 Go 零值是 **0 = 母星的索引**。平行陣列補齊時如果填零值,
        `growRelocation` 一律填 −1,`TestRelocationDefaultsToNoneNotStarZero` 把它釘住。
        `hotseat.go` 的 seat 型別上寫著「`TestSeatFieldsCoverPlayerSide` 用反射盯著它」,
        改成真的寫一支 `TestSeatRoundTripKeepsEveryField`:用反射把每個 seat 欄位塞成可辨識的
        非零值,`loadSeat` 再 `saveSeat` 抓回來比對,漏抄的欄位會停在零值。
        它立刻抓到 `SelectedFleet` ——不過那是測試填了越界值被不變量夾回,不是產品 bug,
        - ~~顯示開關 `ShowRelocationLines` 已建,但**還沒有 UI 可以切**~~ ⚠ **2026-08-08 追認:已過期**——第 23 項(多艦隊模型)接進遊戲選單(`cmd/moo2/gamemenu.go:175`)


    ### **艦隊列表改成真的列「艦隊」+ 清掉 HONEST-STATUS 三條過期斷言**(2026-08-07)。


        ### 艦隊列表先前列的是「船」

        標頭可點擊切換 `SelectedFleet`。`▶` 標出目前操作中的那一支——星圖上所有的按鈕
        | 原本寫的 | 查證結果 |
        |---|---|
        | 「中段的行星表面 + 建築 sprite 擺放子系統仍未做」 | **同一份文件裡自己打架**——同一天稍晚的段落就寫著地表、道路、抖動、國會大廈、植被全都做完了 |
        | 「建築集合仍與原版有差(Colony Base、一次性改造沒建模)」 | Colony Base 第 17 項(拓殖基地)補上;一次性改造第 21 項(改造不佔格)**查證後確認不該佔格**(原版記旗標那步對那四個編號是跳過的),remake 天然正確 |
        | 「戰機/航母(新戰鬥子模型)需先建基礎設施」 | 說得太重。戰機**已經接進快速艦隊戰鬥**(`FighterBayCombatContribution`,中隊數 4/2 是手冊 GM p.127 硬數字),`ResolveBattle` 與 `repair.go` 都在用。真正缺的是**戰術格子裡的獨立戰機單位**(出擊/攔截/回收) |


    ### **RELOCATE 鈕的原版語意追出來了:兩段點選 + 四條合法性規則**(2026-08-07,第 23 項(多艦隊模型)留的問題)。


        第 23 項(多艦隊模型)把 RELOCATE 鈕標成「按下去做什麼沒有反組譯確認,所以先不接」。查了符號表,
        整組符號都在,問題直接解掉:

        ```
        Okay_To_Set_Relocate_From_Star_ @ 0x74F8A    Star_Relocation_        @ 0x75180
        Okay_To_Set_Relocate_To_Star_   @ 0x74FAA    Cancel_Star_Relocation_ @ 0x7522B
        Okay_To_Set_Relocate_Star_      @ 0x75035    Set_All_Star_Relocations_ @ 0x785EC
        ```
        ```
        if *起點 == −1:  驗證能不能當**起點** → 通過就記起來,結束
        else:            驗證能不能當**終點** → 通過就記起來
                         if *終點 == *起點: Cancel_Star_Relocation_    ; 點回自己 = 取消
        ```
        **Fleet Operations console**」),星圖面板的鈕走捷徑,規則面共用同一支 `SetStarRelocation`。
        ### 四條合法性規則(`Okay_To_Set_Relocate_Star_`,`dl` 區分起點/終點)
        | 規則 | 反組譯依據 |
        |---|---|
        | 黑洞不能當起點也不能當終點 | `cmp byte [星+0x16], 6` → 兩種訊息(0x83/0x84),同一條規則 |
        | 沒探索過的星不行 | `star[+0x33] & (1<<玩家)` 的位元測試(逐玩家的探索遮罩) |
        | 目的星上有艦隊 → **跳確認框**;當起點則直接不行 | `sub_7A47A` 走艦隊表 `word_192248[]` |
        | 起點必須是自己有殖民地的星 | `star[+0x38]` 的位元測試(`sub_79D1C`) |
        這是**已知的簡化,不是漏看**,寫在 `relocation.go` 檔頭。
        `Set_All_Star_Relocations_` @ 0x785EC 與 `Clear_All_Star_Relocations_` @ 0x77BB1


    ### **分艦隊:原版的「艦隊」是 ship stack,拆分是把船從串列摘下來**(2026-08-07,task #50 收尾)。


        ### 資料結構

        ```
        word_192248[stack]      = 這一疊的頭一艘船 id
        word_1975D6[船 id × 5]  = 「下一艘」            ; 單向串列
        word_1975D4[船 id × 5]  = 那艘船在哪顆星
        word_199A02             = 目前的 stack 數
        ```
        `Split_Stack_` @ 0x5D689 收一組船 id,把它們從原本的串列摘下來、串成新的一個 stack,
        remake 用切片而不是串列,語意逐項對得上(`SplitFleet`)。
        | 情形 | 為什麼擋 |
        |---|---|
        | 沒選任何船 / 選了全部 | 「全選」不是拆分——那會留下一支空的舊艦隊 + 一支一樣的新艦隊 |
        | 艦隊索引或船索引越界 | 一般防禦 |
        | **艦隊正在航行** | remake 的航行是**整段跳的**,中途沒有位置,拆出來的那一半沒有可放的地方。**這是 remake 移動模型的後果,不是原版規則**——原版的 stack 隨時有座標 |


    ## 三、一星多行星 / 多艦隊 / 拓殖 / AI(2026-08-07)


    ### **`Set_All` / `Clear_All` 集結點 + 遷移連線的顯示開關**(2026-08-07,把第 23/23 項留的小項清掉)。


        ### `Set_All` 有一個猜不到的細節

        ```
        for star = 0 .. 星數-1:
            if word[星×0x71 + 0x54 + 玩家×2] != -1:      ; ← 只改**已經有設定**的
                word[...] = 目標星
        ```
        `TestSetAllOnlyRetargetsExisting` 把它釘住。
        `Clear_All_Star_Relocations_` 同結構,清成 −1。
        艦隊列表的 **ALL**(remake 譯「全部」)鈕接上前者;`Clear_All` 規則已實作並測試,
        手冊兩處明說 ALL 是「全選/全不選這支艦隊的艦艇」;原版的 `Set_All` / `Clear_All` 是別的東西。


    ### **ALL 鈕根本不是集結點**(2026-08-07,把第 23 項(多艦隊模型)的一個推測推翻)。


        ### 兩處手冊,同一句話

        第 23 項(多艦隊模型)寫著「艦隊列表的 ALL 鈕接上 `Set_All_Star_Relocations_`」,並且自己標了「推測」。
        p.47 同時給出那三顆鈕的完整清單:**All / Relocate / Scrap**,`Set_All` 不在其中。
        星圖的輸入處理器 `sub_73980`:
        ```
        cmp eax, 0FFFFFBAFh   ; −1105 → Clear_All_Star_Relocations_(玩家) + 訊息 0x76
        cmp eax, 0FFFFFC13h   ; −1005 → 切換 byte_19BED0(「下一次點星要 Set_All」模式)
                              ;         之後點星才呼叫 Set_All_Star_Relocations_ + 訊息 0x77
        ```
        - **ALL 鈕** → `toggleSelectAllShips`:全選/全不選。選取狀態本來就有(分艦隊用的就是它),
        - **Set_All / Clear_All** → 名冊下方兩個**明確標示為 remake 自加**的入口
        但 `Okay_To_Set_Relocate_Star_` 對終點只檢查黑洞/已探索/怪獸確認,
        `Player_Has_Colony_In_System_` 那一條只在**起點**分支裡。

24. **一星多行星:軌道模型的資料層(第一階段)**(2026-08-07)。

    remake 的 `Stars[i]` ↔ `Planets[i]` 是**一對一**——一顆星一顆行星。MOO2 不是這樣。

    | 來源 | 內容 |
    |---|---|
    | 偏移算術 | 軌道陣列在星球結構 +0x4A,下一個已知欄位(遷移目標)在 +0x54 → 中間 **10 位元組 = 5 個 word** |
    | `System_Planet_Scanned_To_Planet_Id_` @ 0x78CDB | `word[星×0x71 + 0x4A + 軌道×2]` = 軌道 → 行星 id(−1 = 空) |
    | 走訪迴圈 | 上界寫死:`cmp word ptr [var_4], 5; jge`(0x1CB31) |
    行星是**獨立的一張表**(`dword_1930D4`,每筆 **0x11 = 17 位元組**),
    `Planet_Orbit_` @ 0x783ED 讀 `byte[行星 id×0x11 + 3]` = 它在第幾號軌道。
    `genPlanets` 已經在跑 `RollNumSatellites` + 逐軌道 `RollSatelliteType`,
    然後**挑一顆代表行星**,其餘只存成 `Planet.SystemBodies` 的摘要(有軌道/類別/名字,
    - 存取器:`PlanetAt`(第一個有行星的軌道 = 舊的 `Planets[星]`)、`PlanetsAt`、
      `PlanetStar` / `PlanetOrbit`(反查)、`FreeOrbit`(**人造行星要用它**)
    - 存檔遷移 `normalizeOrbits`
    **行為逐位元不變**——每顆星仍然只有一顆行星。`TestGeneratedStarsHaveExactlyOneOccupiedOrbit`
    把這個限制釘住,`TestPlanetAtMatchesLegacyParallelIndexing` 釘住相容性支點:
    這與 `Star.Wormhole`(零值 0 讓每顆星都連到星 0)、`ColonyRelocateTo`(零值 0 = 母星)
    把 `SystemBodies` 升格成真正的 `Planet` 條目、填滿軌道表,


    ### **一星多行星 Step A:33 個呼叫端改走存取器**(2026-08-07,第 24 項(軌道資料層)的續)。


        第 24 項(軌道資料層)建好了軌道模型,但所有讀行星的地方仍然直接寫 `s.Planets[星]` ——
        那個式子**假設 Planets 與 Stars 平行**。一旦產生器開始填滿軌道(Step B),
        `len(Planets)` 就會大於 `len(Stars)`,那些式子會**默默讀到錯的行星**
        ——不會崩、不會報錯,只是資料錯位。

        | 存取器 | 用途 |
        |---|---|
        | `PlanetAt(星) int` | 代表行星的索引 |
        | `PlanetOf(星) *Planet` | **可寫**指標(隨機事件改礦產/氣候、拓殖消耗特殊物產、抵達時的一次性發現) |
        | `PlanetDataAt(星) (Planet, bool)` | 唯讀複本 |
        `PlanetAt` 改成:依軌道順序找第一顆一般行星(可殖民);整組都不宜居時才退而取第一個天體。
        那正是 `genPlanets` 原本挑代表行星的規則。**不一致就會位移**——
        把 `SystemBodies` 升格成完整的 `Planet` 條目、填滿軌道表。⚠ 那會**改變同一 seed 的星系內容**


    ### **一星多行星 Step B:同系天體升格成真正的行星**(2026-08-07,task #51 收尾)。


        第 24 項(軌道資料層)建了軌道模型、第 24 項(軌道資料層)把呼叫端改走存取器,這一步把資料真的填進去:
        **同一顆恆星底下的每一個天體都是完整的 `Planet` 條目,各佔一條軌道。**
        `Planets` 因此**不再與 `Stars` 平行**(24 顆星 → 94 顆行星)。

        非代表天體用 `bodyRand`(seed+5)而不是共用 `r`。這個專案已經為
        `genPlanets` / `genMonsters` / `genWormholes` 各開一條流,理由寫在原註解裡:
        | 位置 | 症狀 |
        |---|---|
        | `monster.go` 兩處 | `planets[starIdx]` —— 怪獸的特殊物產會補到**別的星系**的行星上 |
        | `galaxy_gen_test.go` / `monster_test.go` | 測試自己也在平行索引 |
        順手把「代表行星怎麼挑」收斂成唯一一份實作(`representativePlanet`),
        生成階段與 `GameSession.PlanetAt` 共用。兩份實作一旦漂開,徵狀又是資料錯位而不是崩潰。
        ### `SystemBodies` 淘汰:它自己註解裡擔心的事解掉了
        (`GameSession.SystemCompositionText` / `SystemBodyCountText`),**只有一份資料**。
        `TestGeneratedStarsHaveExactlyOneOccupiedOrbit` 的註解當初就寫著
        `FreeOrbit` 現在真的有意義——**人造行星**(建築 48)可以往空軌道放。


    ### **同星系多殖民地:拓殖的對象是行星,不是星**(2026-08-07,把第 24/24 項留的最後一步做完)。


        ### 一句話擋掉了整個擴張手段

        `ColonizeStar` 的第二道閘寫著:
        ```go
        if star.Owner != 0 {
            return ColonizationResult{Reason: "該星已有歸屬,不可拓殖"}
        }
        ```
        | 之前 | 之後 |
        |---|---|
        | `ColonizeStar(star)` 是唯一入口 | `ColonizePlanet(planet)` 是入口;`ColonizeStar` 變成「該星系第一顆可殖民行星」的捷徑 |
        | `newColonyFromStar(star)` 讀 `Planets[star]` | `newColonyFromPlanet(planet)` |
        | 殖民地只記 `PlayerColonyStars[i]` | 加上 `PlayerColonyPlanets[i]`(AI 側是 `AIOpponent.ColonyPlanets`) |
        | 前哨站只記在哪顆星 | `Outpost.PlanetIndex`(手冊 p.119:「build a military outpost on a single planet」) |
        | 殖民地名 = 該星代表行星名 | `ColonyName(i)` = 該殖民地**座落行星**的名字 |
        | 地表變體種子 = 星索引 | = 行星索引(否則同星系兩個殖民地地表一模一樣) |
        `PlayerColonyPlanets` / `Outpost.PlanetIndex` 的「未知」都必須是 **−1**。
        這是第四次了(`Star.Wormhole` → 全部連到星 0、`ColonyRelocateTo` → 全部指向母星、
        `Star.Orbits` → 每顆星都宣稱有行星 0)。索引型欄位的 Go 零值是一個**合法索引**,
        行為與加欄位前逐位元一致,`TestLegacySaveWithoutColonyPlanetsFallsBackToStar` 釘住。
        `consumeOutpostForColony` 原本只比對**星**。多天體之後,在一顆行星建殖民地會把
        現在比對行星(舊存檔的 −1 退回舊語意),`TestOutpostOnAnotherBodySurvivesColonizingNeighbour` 釘住。
        順便放寬了 `BuildOutpost` 的 `Owner != 0` 閘——氣態巨星/小行星帶常常就在自己已經殖民的
        不必新造:**行星列表**(`PLNTSUM.LBX`)右下角本來就烘著
        `SEND COLONY SHIP` / `SEND OUTPOST SHIP` / `RETURN` 三顆鈕——那是原版選行星的地方。
        - 列出**目前看得見的星系**(`shell.VisibleStars`,與星圖同一套可見性)的所有天體
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


    ### **AI 也會在自己的星系裡拓殖第二顆行星**(2026-08-07,把第 24 項(軌道資料層)的另一半補上)。


        ### 一個 remake 自己造出來的不對稱

        第 24 項(軌道資料層)只改了玩家側。`aiExpand` 的候選集寫死成
        ```go
        for idx := range s.Stars {
            if s.Stars[idx].Owner != 0 { continue }   // ← 只找無主星
            ...
        }
        ```
        ### `Star.Owner` 分不出是哪一個 AI
        候選集要加進「**自己**已有殖民地的星系」,而 `Star.Owner` 只有 0/1/2 三個值
        (無主 / 玩家 / AI),分不出 `2` 是哪一家。所以判定要走各自的 `ColonyStars` 清單
        (`aiCanExpandInto`),不能只看 `Owner`。兩支測試分別釘住「不能進玩家的星系」與
        - `OwnedStars` 只在「本來無主」時才 `++`。在自己的星系裡多殖民一顆行星不會讓版圖變大,
        - `PlanetColonized` 取代原本只查玩家的 `ColonyIndexOnPlanet`。只查玩家的話,
        `InvadeColony` 打贏就無條件 `star.Owner = 1`。同星系多殖民地之後,
        「站在玩家星系裡的敵軍」——星圖顏色、可入侵性、`ColonizePlanet` 的 `Owner == 2` 閘全部對不上。
        現在星的歸屬(以及 `StarCaptured` 回報與 `OwnedStars--`)只在**該 AI 在這顆星上
        - `internal/shell/session.go`:`aiCanExpandInto` / `aiExpansionCandidates` / `aiExpand`
        - `internal/shell/colonization.go`:`PlanetColonized`(全帝國視角)
        - `internal/shell/ground_invasion.go`:歸屬翻面的條件 + 過戶行星用真值
        - `internal/shell/multicolony_test.go`:再加 5 支(AI 自家星系拓殖、OwnedStars 不灌水、

25. **人造行星:手冊推翻了 remake 自己的假設**(2026-08-07,建築 48 接線)。

    ### 先訂正一句寫了兩輪的話

    於是 `FreeOrbit` 被寫成「人造行星要用它:沒有空軌道就蓋不了」。**手冊說的不是那樣。**
    `TestArtificialPlanetNeedsMaterialNotFreeOrbit` 把這個訂正釘住:
    `sub_13FD9` 的那一段走**兩趟**掃這顆星的 5 條軌道:
    | 趟次 | 找什麼 | 結果尺寸 |
    |---|---|---|
    | 第一趟 | `planet[+4] == 2`(氣態巨星) | `var_1C = 4` |
    | 第二趟(第一趟沒中才跑) | `planet[+4] == 1`(小行星帶) | `var_1C = 3` |
    - `internal/gamedata/special_actions.go`:第 5 個 Special 行動(前置 `TOPIC_ADVANCED_MANUFACTURING`)
    - `internal/shell/artificialplanet.go`:兩趟掃描 + 固定結果(Barren / Normal G / Abundant)
    - `session.go` 的 Special 分派:沒有材料時**誠實地什麼都不發生**

26. **是/否確認框 + 一條寫錯的規則**(2026-08-07,把第 23 項(多艦隊模型)留的「沒有 modal 基礎設施」清掉)。

    ### 先訂正:那個條件不是艦隊,是怪獸

    `relocation.go` 檔頭與本報告都寫著:
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
    `relocation.go` 的檔頭。
    另外原版對**起點**的怪獸是**靜默拒絕**(`loc_7511B` 只把結果清 0)。remake 出一句話:
    | 元件 | 來源 | 值 |
    |---|---|---|
    | 底框 | `sub_12B7E1(0A1h, 75h, CONFIRM#0)` | (161, 117),圖 313×227 |
    | Y 鈕 | `sub_12B7E1(0EBh, 12Eh, CONFIRM#1)` | (235, 302),圖 54×24 |
    | N 鈕 | `sub_12B7E1(159h, 12Eh, CONFIRM#2)` | (345, 302),圖 55×24 |
    | Y 熱區 | `sub_11438B(eax=0EBh, edx=12Eh, ebx=11Eh, ecx=143h)` | 235..286 × 302..323 |
    | N 熱區 | `sub_11438B(eax=159h, edx=12Eh, ebx=18Ch, ecx=143h)` | 345..396 × 302..323 |
    | 文字 | `sub_77A74(eax=0CCh, edx=0D0h, ebx=0E0h)` | 左緣 204、垂直置中 208、**折行寬 224** |
    `Draw_Confirm_Box_` @ 0x778E4 每幀把兩顆鈕的幀號歸 0,再把**游標所在**那顆設成 1
    原版在文字放不下時會**縮字級**:`sub_103CAF` 量高度,`var_C` 從 4 遞減到 1,
    - `cmd/moo2/confirmbox.go`:widget(疊在下層畫面上,框外點擊無效——modal 的重點)
    - `internal/shell/relocation.go`:`RelocateToNeedsConfirm` + 起點的怪獸拒絕 + 檔頭訂正
    - `cmd/moo2/relocation.go` / `interactive.go`:`pendingConfirm` 接線

27. **戰術格子的獨立戰機單位 + 一個讀錯的欄位**(2026-08-07)。

    ### 先訂正:那一欄是 Shots,不是「出擊數」

    `gamedata/combat.go` 寫著:
    ```go
    // FighterInterceptorSquadron 是一個攔截機戰機庫每次出擊的戰機數(手冊 GM p.127「出擊數」欄:攔截機 4)
    const FighterInterceptorSquadron = 4
    // FighterHeavySquadron 是一個重戰機庫每次出擊的重戰機數(手冊 GM p.127「出擊數」欄:重戰機 2)
    const FighterHeavySquadron = 2
    ```
    ```
    Weapon | Armament | Shots | Size | Cost | Speed | Hits | Strat Dmg
    ```
    而正文說攔截機「speed 10」、重戰機「speed 8」——差 2。套 `CombatFighterSpeed` 就通了:
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
    - **「always attack from the weakest shield facing」**:`CombatShip` 已保存四面護盾容量,
      但戰機自動尋找最弱面與艦身旋轉仍未接入;一般艦艇命中鏈已能按來向選面。
    - **轟炸機 / 突擊梭**:前者要炸彈對行星的規則、後者要把陸戰隊送上敵艦,各自依賴另一套系統。
    - **敵方不會派戰機**:`genEnemyFleet` 產出的敵艦沒有設計資料,讀不到「帶不帶戰機庫」。
    - **FTL 階 / 裝甲級**:艦艇設計還沒把「目前最佳引擎/裝甲」餵進戰鬥層,出擊時先傳 1 / 0
    - **出擊鈕不是原版版面**:原版的控制列是烘死的美術,那七顆鈕各有其原意,
    - `internal/gamedata/combat.go`:`FighterSquadronSize` / `FighterShots*` / `FighterHits*` /
      `FighterHitsWithArmor`,兩支貢獻函式依真值重算
    - `internal/shell/fighter.go`:中隊狀態機(11 支測試)
    - `internal/shell/session.go`:`CombatShip.Bay` / `BayKind`(與快速結算讀同一份設計資料)
    - `cmd/moo2/tacticalfighter.go`:每回合的目標選擇、推進、結算與繪製(6 支測試)

28. **「AI 的遷移設定」不是缺口**(2026-08-07)。

    ### 一個被假設出來的缺口

    那句話的前半是對的(remake 的 `AIOpponent.FleetStrength` 是一個抽象數字),
    | 寫入者 | 呼叫端 |
    |---|---|
    | `Universe_Generation_` | 開局把 `[星 + 玩家×2 + 0x54]` 對 8 個玩家全部初始化成 −1 |
    | `Set_Relocation_` | 只有 `Star_Relocation_`(玩家的兩段點選) |
    | `Clear_Star_Relocation_` | 也只有 `Star_Relocation_`(點回同一顆 = 取消) |
    | `Set_All_Star_Relocations_` | 星圖輸入處理器 `sub_73980` + `Main_Screen_` |
    | `Clear_All_Star_Relocations_` | `sub_73980` |
    讀取端 `Redirect_Newly_Built_Ships_` **確實是逐玩家跑的**(收 player 參數、
    `Set_All` / `Clear_All` 把 `星基 + 玩家×2` 先加好,再 `mov [eax+54h], bx`,
    ```
    movsx eax, si
    add   eax, eax          ; 玩家×2
    add   eax, ecx          ; + 星基
    mov   [eax+54h], bx     ; ← grep 的樣式對不上
    ```
    正確做法是先把 asm 切成一支支函式,再找「同時碰 `dword_19306C`(星表)、`71h`(stride)


## 四、網路多人整塊(2026-08-07)

> **2026-08-10 現況勘誤**：本節是 2026-08-07 的施工日誌；其中「網路多人到此結案」只
> 描述傳輸／畫面／測試骨架完成，不代表正式 `cmd/moo2` 對局可玩。生產呼叫圖目前仍缺
> 名冊→共同開局、席位映射、全玩家操作指令收集、真正回合結算與分岔停機。原版 IPX／
> 數據機／序列／TEN 不恢復，依 `docs/tech/multiplayer-architecture.md` 決策改走 TCP
> lockstep；目前待辦以 `WORKLIST.md` 2026-08-10 盤點結論為準。

29. **決定性化——網路多人的地基,順手抓出兩個存檔 bug**(2026-08-07)。

    ### 為什麼先做這一塊

    1. **不必另外維護欄位清單**——新欄位只要進得了存檔就自動進得了指紋;反過來說,
    2. `encoding/json` 對 map 的鍵**保證排序**,`ColonyBuildings` 這類 map 不會因為 Go 的
    ```
    存檔 → 讀檔 → 繼續玩   ← 事件序列從頭開始
    ```
    修法有個坑:直覺是「記下抽了幾次,讀檔時重抽幾次跳過去」,但 `math/rand` 的
    `Intn` 與 `Float64` **從底層 source 取走的數量不一樣**,所以「重抽 n 次」必須連
    **抽的種類**都一樣才會落在同一格。`randstream.go` 改成直接騎在 `rand.Source64` 上,
    ```
    < "HyperAdvancedResearchCost": 25000
        ```
    `RuleProfile` **完全沒進存檔**。讀檔後它是零值——那既不是 1.3 也不是 1.5,
    - `internal/shell/determinism.go`:`StateHash` / `StateFingerprint`
    - `internal/shell/randstream.go`:可快轉的亂數流
    - `internal/shell/persist.go`:三個抽取次數 + `RuleVersion`
    - `internal/shell/determinism_test.go`:7 支閘門(同種子逐回合、加玩家指令、存讀檔指紋、


    ### **傳輸層 + 鎖步協定**(2026-08-07,`internal/netplay`)。


        ### 原版的形狀(反組譯,不是猜的)

        `Net_Next_Turn_` @ 0xFC470 的骨架:
        ```
        byte_1AAF7E[本方玩家] = 1                  ; 標記「我這回合結束了」
        for 每個其他玩家:
            if 玩家還在線(`[player+0x28] == 0x64`):
                sub_F6816(...)                      ; 把自己的狀態送過去
        Wait_Until_Net_Opponent_Finished_ @ 0x3FBFE ; 等對手
        ```
        `byte_1AAF76` / `byte_1AAF7E` 是逐玩家的旗標陣列(玩家結構 stride 0xEA9)。
        那三種在現在的機器上都不存在。這是**移植決策不是還原**,標在 `frame.go` 檔頭。
        | 決定 | 理由 |
        |---|---|
        | `internal/netplay` **不相依** `internal/shell` | 傳輸層不該知道規則。兩層各自測得完,端到端那支放在**外部測試套件** `netplay_test`,同時 import 兩邊而不讓生產程式碼耦合 |
        | 4 位元組長度前綴 | TCP 是位元組流沒有訊息邊界:一次 `Write` 的東西可能被分兩次讀到,也可能與下一則黏在一起。上限 4 MB 不是為了省記憶體,是**不讓對面一個壞掉的長度欄位要求我們配置 4 GB** |
        | 指令**依玩家編號**排序 | 不是為了好看:鎖步要求每台機器以同樣順序套用同樣的指令。依到達順序套用會讓「誰的封包先到」影響結果——那正是 lock-step 最典型的分岔來源 |
        那時候已經回推不了是哪一步歪的。每回合比一次(第 29 項(決定性化)的 `StateHash`),
        `TestTwoPeersStayInSyncOverAPipe`:兩個對等端各跑一份 `GameSession`,
        - **9 個畫面**:`Join_Net` / `Modem_Setup` / `NullModem_Setup` / `Choose_Net_Plyrs` /
          `Choose_Multi_Net_Game` / `Generic_Net_Info` / `SendGet_Net_Info` / `Net_Next_Turn` /
        - **指令解譯器**:把 UI 的每一顆按鈕對到一條 `netplay.Command`。端到端測試裡只解了三條


    ### **玩家指令層**(2026-08-07,`internal/shell/command.go`)。


        ### 三個用途,只有一個是網路

        | 用途 | 為什麼需要「指令」這一層 |
        |---|---|
        | 網路對戰 | 兩台機器要套用**同樣的指令序列**才會算出同樣的狀態 |
        | 回放 / 除錯 | 一局的完整指令序列 + 起始種子 = 可完整重現的 bug 報告 |
        | 熱座 | 其實已經在做同一件事,只是指令直接就地套用、沒有序列化 |
        **① 指令層不做前置檢查。** `ColonizePlanet` 之類的方法自己會回絕不合法的操作
        `shell.PlayerCommand` 與 `netplay.Command` 的欄位形狀一模一樣,但是**兩個型別**——
        正式對局裡是 `cmd/moo2`(組裝端),測試裡是外部測試套件。
        - `TestEveryListedCommandIsHandled`:表上有的都認得,而且清單有排序、無重複
        - `TestCommandPathMatchesDirectCall`:走指令層與直接呼叫方法必須得到**一模一樣的狀態**
          (逐條比對 `StateHash`)——指令層只是轉接,不該有自己的規則
        - `TestEveryPlayerCommandCanTravel`(netplay 側):每一條指令都要能過線再被規則層認得,
        **9 個網路畫面**:`Join_Net` / `Modem_Setup` / `NullModem_Setup` / `Choose_Net_Plyrs` /
        `Choose_Multi_Net_Game` / `Generic_Net_Info` / `SendGet_Net_Info` / `Net_Next_Turn` /
        `Wait_For_*`。版面座標都還沒抽——`Draw_Net_Next_Turn_Screen_` @ 0xF1075 與
        `Add_Net_Next_Turn_Fields_` @ 0xEFCEA 是抽的起點。


    ### **`Net_Next_Turn` 等待畫面——第一張版面是「算」出來的畫面**(2026-08-07)。


        ### 它與先前每一張都不一樣

        這一張不是:`Load_Net_Next_Turn_Screen_` @ 0xF3E42 依**資產尺寸**現算。把那段翻成算式:
        ```
        x    = (0x280 − 資產42.寬) / 2              ; 水平置中於 640
        總高 = 資產42.高 + 資產43.高 + 資產40.高
        y    = max(0, (0x1E0 − 總高) / 2)           ; 垂直置中於 480
        [win+0xBF]  = 資產42.高 + 資產43.高          ; 第三塊的相對位移
        [win+0x10E] = 2                             ; 字型 id
        ```
        | 元件 | 位置 |
        |---|---|
        | 標題帶(10 幀動畫) | (5, 16) |
        | 中段面板 | (5, 64) |
        | 下段面板 | (5, 243) |
        `Add_Net_Next_Turn_Fields_` @ 0xEFCEA 再給兩個真值:輸入列
        **測試釘的是算式不是數字**:`TestNetWaitLayoutMatchesTheOriginalFormula` 重算一次
        - **玩家列的起始 y 沒有算出來**:那一段把座標藏在 window 結構的欄位裡,沒有直接的立即數。
        - **聊天輸入列**:原版 y=430 那一列是文字欄位,remake 沒有聊天——畫成提示帶並寫明。
        - **狀態指紋擺在畫面上**不是裝飾:分岔時兩邊念一下那八個字元就知道是不是同一個狀態,
        `Modem_Setup` / `NullModem_Setup` / `Comm Info` 是**數據機與序列線**的設定。
        剩下的 5 張(`Join_Net` / `Choose_Net_Plyrs` / `Choose_Multi_Net_Game` /
        `Generic_Net_Info` / `SendGet_Net_Info`)是**連線流程**的畫面,要等 UI 端的連線流程


    ### **連線大廳 + `Choose_Net_Plyrs` 名冊——第一個「尺寸隨資料變」的版面**(2026-08-07)。


        ### 先做大廳,不先做畫面

        只是空框。所以這一輪先補 `internal/netplay/lobby.go`——**主機聽、客戶端連、主機廣播名冊**:
        | 角色 | 做的事 |
        |---|---|
        | 主機 `Host(addr, name, seed)` | 開 listener,自己是 0 號;`AcceptOne` 收一個人、指派 id、**廣播整份名冊** |
        | 客戶端 `Join(addr, name, timeout)` | 送 `hello` → 收 `roster` → 拿到自己的 id 與種子 |
        - **玩家編號由主機指派。** 鎖步的指令排序鍵就是玩家編號(見第 29 項(決定性化)),
        - **種子由主機決定並廣播。** 種子決定整張星圖與所有隨機事件;各自產生種子
        `Choose_Network_Plyrs_Screen_` @ 0xF0E17 的定位段:
        ```
        x    = (0x280 − 資產27.寬) / 2
        總高 = 資產28.高 × [win+0x1E1] − 1 + 資產27.高 + 資產29.高
        y    = (0x1E0 − 總高) / 2
        ```
        `[win+0x1E1]` 是**列數**——中段面板(資產 28)每位玩家重複一次。先前移植的畫面
        | 人數 | 總高 | y |
        |---|---|---|
        | 1 | 36×1 − 1 + 81 + 38 = 154 | 163 |
        | 4 | 36×4 − 1 + 81 + 38 = 262 | 109 |
        | 8 | 36×8 − 1 + 81 + 38 = 406 | 37 |
        `Add_Choose_Net_Plyrs_Fields_` @ 0xEFB50 給每列的點擊區(逐項立即數):
        `x1 = winX + 0x6A`、`y1 = winY + i×0x24 + 0x40`、`x2 = winX + 0x1B3`、`y2 = y1 + 0x1D`
        並補 `TestChooseNetPlayersInfoLinesSitBelowTheWindowAndStayOnScreen` 把它釘住
        - **沒有文字輸入框,所以「加入」只連得上本機。** 要連別台得先做輸入框——
          這是下一步,不是這一輪偷懶;`netLobbyDialAddr` 的註解寫明了。
        - **不能點列指派種族。** 原版這張畫面可以(`sub_EFABA` 在每列旁再建一組欄位),
        - **沒有重連、沒有心跳、沒有加密。** 這是區網對戰的最低限度,寫在 `lobby.go` 檔頭。
        剩下 4 張(`Join_Net` / `Choose_Multi_Net_Game` / `Generic_Net_Info` / `SendGet_Net_Info`)。


    ### **連線狀態面板——反組譯把「還缺 4 張」改寫成「還缺 1 張」**(2026-08-07)。


        ### 抽版面時發現那 4 張其實是 2 張

        ```
        0xF19C7  Draw_Generic_Net_Info_Screen_
        0xF19C7  Draw_Join_Net_Screen_          ← 同一個位址
        ```
        兩個名字指向同一段程式碼。往上追,`Reload_Generic_Net_Info_` @ 0xF53D7 收一個
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
        版面又是算的:`x = (0x280 − 資產.寬)/2`、`y = (0x1E0 − 資產.高)/2`,字型 id 4。
        `Draw_SendGet_Net_Info_Screen_` @ 0xF2C8B 另給進度數字的兩組位移
        ——`[win+0x10F]==0`(傳送)→ (+0x72,+0x42)、`==1`(接收)→ (+0x79,+0x41),
        **一、把 `Add_Waiting_For_Joiners_Field_` 讀成「已加入人數」欄位。**
        截圖上那串數字壓在 `START NET GAME` 上,才回去查它呼叫的 `sub_1151B0` 是什麼——
        符號表寫著 **`Add_Button_Field_`**。那個 (+0x9E,+0x6A) 是**按鈕**的左上角,
        `multigmFrameWithKey` 一直是逐幀獨立解碼——這個 bug 先前沒被發現,是因為截圖廊
        修在 `internal/lbx`(`AccumulatedUpToRGBA`)而不是這一個畫面:同樣的坑對
        - 「加入對局中」永遠不會停留:`netplay.Join` 是同步的,連上或逾時都在同一個呼叫裡結束。
        **剩 1 張**:`Choose_Multi_Net_Game`(版面已抽出,見下輪)。


    ### **區網對局探索 + `Choose_Multi_Net_Game`——最後一張網路畫面**(2026-08-07)。


        ### 這張畫面的資料從哪來,原版沒有回答

        所以先補 `internal/netplay/discovery.go`:主機每秒往 `255.255.255.255:24502` 廣播
        - **來源 IP 覆蓋封包裡寫的**:主機常常不知道自己對外是哪個位址(多網卡、容器內、NAT)。
        - **清單依名稱排序,不依到達順序**:到達順序每次都不一樣,而順序決定玩家點到哪一場。
        - **`Browser` 不阻塞**:UI 是單執行緒的,`Discover` 那種「收兩秒再回傳」會讓畫面凍住兩秒
          (同 `lobby.AcceptOne` 的處境)。背景 goroutine 收,畫面每幀讀快照。
        `Load_Choose_Multi_Net_Game_Screen_` @ 0xF40D3:
        ```
        x = (0x280 − 資產41.寬)/2                    = (640−479)/2 = 80
        y = ((0x1E0 − 資產41.高) − 0x51)/2 + 0x25    = ((480−384)−81)/2 + 37 = 44
        ```
        ——`Draw_Choose_Multi_Net_Game_Screen_` 只 blit 資產 41。也就是它是版面上的讓位,
        `Add_Choose_Multi_Net_Game_Fields_` @ 0xEFF87 給每列的熱區
        (`sub_11438B` = `Add_Hidden_Field_`,隱形熱區——美術已經畫好格子了):
        `x1=+0x26 / x2=+0x190 / y1=+0x40+i×0x1B / y2=y1+0x16`,即每列 **362×22**、列距 27、
        最多 10 場。底下一顆 `Add_Button_Field_` 在 (+0xBF, +0x158)。
        `Draw_Choose_Multi_Net_Game_Screen_` @ 0xF1AF4 再給文字的擺法:
        ——**字在 22 px 的列裡垂直置中**。選中的那一列另有脈動亮度(`[win+0x1E8]` 在 −3..+4
        第一版把原版的 `(0x16 − 字高)/2` 加了一個字高當基線傳給 `uifont.Draw`,結果**整欄字
        掉到下一列**——截圖上選取框在第一列、字在第二列。原因是 `uifont.Draw` 底層是
        - 原版可以在這張畫面**改對局名稱**(`Change_MP_Game_Name_` @ 0xF5777:長度上限 8、
          上限與唯一性的規則已記在 `netplay.GameNameMax`,做輸入框時直接套。
        **9 張網路畫面到此結案**:6 張做了(`MP_Setup` / `Hotseat` / `Net_Next_Turn` /
        `Choose_Net_Plyrs` / 狀態面板 7 狀態 / `Choose_Multi_Net_Game`),
        3 張明確不做(`Modem_Setup` / `NullModem_Setup` / `Comm Info` —— 硬體已不存在)。


    ### **文字輸入彈窗——原版有一個,而且有自己的 LBX**(2026-08-07)。


        ### 先查被呼叫的函式叫什麼

        `Change_MP_Game_Name_` @ 0xF5777 呼叫 `sub_91BB4`,而符號表寫著
        **`Remapped_Input_Box_Popup_`** ——原版有一個獨立的 modal 彈窗,連自己的 LBX 都有
        (`INBOX.LBX`,只有兩個資產:0 = 288×151 底框、1 = 98×28 的 ACCEPT 鈕)。
        (前兩次:第 29 項(決定性化)的 `Add_Waiting_For_Joiners_Field_` → `Add_Button_Field_`、
        第 29 項(決定性化)的清單列 → `Add_Hidden_Field_`)。
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
        位置真值:`Star_Name_Popup_Screen_Center_X_` @ 0x923BE = 0xB1 = 177、
        `_Y_` @ 0x923C4 = 0x7D = 125。x 幾乎正好是水平置中((640−288)/2 = 176),
        - **改對局名稱**:主機按「開始新遊戲」→ 先問名稱(上限 8,`Change_MP_Game_Name_` 的
          `edx`)→ 才開大廳。原版就是這個順序,而且名稱是別人在清單上看到的東西,
        - **加入指定位址**:清單畫面多一顆「直接輸入位址」。⚠ **原版沒有這顆鈕**——
        原版的輸入處理走 `sub_91B89` 逐鍵掃描碼(IME 之前的年代)。remake 用 ebiten 的
        `AppendInputChars`——**移植決策**:掃描碼那一套在現代平台拿不到,而且會擋掉輸入法。
        + 輸入框。剩下的是**聊天列**(`Chat_Box_Input_Loop_` @ 0xF55A4 已定位,
        `Send_Chat_Msg_` @ 0xDD3B8 是送出端),那是加分項不是缺口。


    ## 五、手冊與反組譯校準(2026-08-07~08)

30. **TECH LEVEL 的第二個真效果——那張「拿不到所以不臆造」的表挖到了**(2026-08-07)。

    ### 缺的是表,而 `shell.TechLevels` 的註解自己說了

    ### `Init_Player_Tech_` @ 0x5E55F 給了兩樣
    **送幾個**由 `byte_199CB5`(NEW GAME 的 TECH LEVEL,來自 `word_1A1360`)決定:
    ```
    al == 0 → var_18 = 1
    al == 1 → var_18 = 6
    al == 2 → var_18 = 0x19 = 25
    ```
    **送哪些**在 `word_18111C` @ 0x18111C:`dw 1Dh, 37h, 16h, 39h, 1Ch, 17h`
    → **29, 55, 22, 57, 28, 23**。主迴圈前 6 次取這張固定表,第 7 次起改由 `sub_FD335`
    | 原版 | id | 主題 |
    |---|---|---|
    | 1Dh | 29 | `TOPIC_ENGINEERING` |
    | 37h | 55 | `TOPIC_NUCLEAR_FISSION` |
    | 16h | 22 | `TOPIC_CHEMISTRY` |
    | 39h | 57 | `TOPIC_PHYSICS` |
    | 1Ch | 28 | `TOPIC_ELECTRONICS` |
    | 17h | 23 | `TOPIC_COLD_FUSION` |
    - **手冊**獨立說「預設的第一個是 field #29」——remake 先前就把這句記在註解裡了。
    - **反組譯**的 `word_18111C[0] = 0x1D = 29`。
    - **第二個 55 = 核分裂**,而 remake 早就把 `FTLTopic` 定成 `TOPIC_NUCLEAR_FISSION`
    `applyStartingTech` 一開始只「加」不「減」,結果 `NewDemoSession` 用預設等級發過的
    (`TestPrewarpDoesNotKeepTheDefaultLevelsFTLTopic`:先斷言 demo 局本來就有,再斷言改成
    **AI 一起發**:原版的 `Init_Player_Tech_` 第一個參數就是玩家編號,是逐玩家跑的。
    `TOPIC_NUCLEAR_FISSION`。這是端到端的證據,測試綠不是。
      缺口大小由 `gamedata.StartingTopicRandomExtras` 回報——**讓缺口是一個看得見的數字,
      ⚠ **2026-08-07 第 43 項(先進級開局主題)補上了**(照抄 `sub_FD335` 的挑選結構),那個函式現在的角色
    - 原版 `byte_199CB5 >= 3` 那條路徑**沒有設 `var_18`**(`enter` 不清堆疊)。

31. **開局建築的優先清單——手冊只給了結論,清單本身在執行檔裡**(2026-08-07)。

    ### 缺的又是一張表,而程式碼自己標了

    `shell.StartingBuildingCount` 的註解:
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
    **清單 `word_17D8AC`**(32 個原版建築編號,順序照原版):
    ```
    41  8  40  21  22  15  7  20  37  4  31  33  34  39  2  12  13  32
    35  10  43  16  28  25  18  19  24  26  47  27  6  5
    ```
    落成測試 `TestAverageStartGivesExactlyMarineBarracksAndStarBase`:三張表任何一張抄錯都會紅。
    另有一支正對照 `TestHomeworldBuildingsMatchTheHardCodedPair`:新舊兩條路要走到同一個答案,
    `TestAdvancedStartIsBlockedByTheMissingRandomTopics` 把這件事釘住,並附正對照:
    先進級**真的發滿 6 棟**了。那支測試已改名為 `TestAdvancedStartFillsAllBuildingSlots`
    那 19 個要接得先港 `Choose_Tech_Application_` @ 0xFD335 ——294 行的 AI 科技權重選擇器
    (權重吃成本表 `dword_17D916`、種族旗標 `byte_17E084`、性格 `[player+0x28]`、
    政府別、`sub_FC845` 的估值)。**一次讀就照抄風險太高**,留作獨立一輪。
    順帶把 `origBuildingID` 從 `cmd/moo2/colonysurface.go` 搬進 `internal/gamedata`:

32. **飛彈速度:那個「手冊自相矛盾」不是矛盾,是漏了條件**(2026-08-07)。

    ### 原本的斷言

    `internal/gamedata/missile.go` 的檔頭從移植那天起就寫著:
    `Missile_Speed_` @ 0x3CD21 的最後三行:
    ```
    loc_3CE40:
        test [ebp+var_3], 10h     ; ← 旗標 0x10
        jz   short loc_3CE49
        add  edx, 4               ; 只有旗標成立才 +4
    ```
    | 手冊的哪一段 | 對應什麼 |
    |---|---|
    | 附表 10/12/…/22 | **沒有** Fast 改造的一般飛彈 |
    | 明列公式(含 +4) | **裝了** Fast 改造的飛彈 |
    `Missile_Speed_` 依**武器類型**分檔:
    | 類型 | 基礎 | 加 FTL 項? | 玩家旗標成立時 |
    |---|---|---|---|
    | 0x0E..0x11 | 12 | 是 | — |
    | 0x12 / 0x13 | 20 | **否** | — |
    | 0x1C | 6 | 是 | 10 |
    | 0x1D / 0x1E | 8 | 是 | 12 |
    | 0x1F | 10 | 是 | 14 |
    | 0x28 | 24 | **否** | — |
    `[player+0x8BC]` 那個玩家旗標(成立時 6→10、8→12、10→14)**還沒追出是什麼**。
    `MissileBaseSpeed` 留一個 `boosted` 參數、呼叫端目前一律傳 false ——

33. **地面戰:結構不是「未核實」,是抄了一代的**(2026-08-07)。

    ### 原本的斷言

    `internal/gamedata/ground_battle.go` 的檔頭:
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
    ```
    strA = A.攻擊力[A.當前類型] + Random(100)
    strB = B.攻擊力[B.當前類型] + Random(100)
    if (strA <= strB) A 挨一下
    if (strA >= strB) B 挨一下      ; ← 兩個獨立的 if
    ```
    | | remake(一代結構) | 原版 |
    |---|---|---|
    | 平手 | `if/else` → **只有攻方**挨打(「平手歸守方」) | 兩個獨立的 `if` → **雙方都**挨打 |
    | 攻擊力 | 整隊一個 `Force` | **逐部隊類型**各一個 |
    | 傷害 | 每次 −1、扣到 `<= 0` 陣亡 | 累積命中,**`==` 耐受值**才陣亡,然後歸零 |
    `ground_invasion.go` 先前有一整段在解釋「為什麼把戰車營排在合併陣列尾端」:
    因為 `ResolveGroundBattle` 只回傳一個總存活數,要靠 `min(總存活, 戰車原始數)` 推算分兵種。
    - **四種部隊類型沒有對出名字。** MOO2 的地面單位是陸戰隊 / 裝甲 / 機械戰士三種,
    - **每種部隊各自的攻擊力表還沒追**(`[side + type*2 + 2]` 的來源)。兩種目前都填同一個
      `atkForce`:填同值 = 維持現行數字,並把差異留在一個看得見的地方。
    - `ground_battle.go` 的舊解算**保留**(加成表與 builder 還在用,而且它是那三處差異的對照組),


    ### **四種部隊類型是什麼——手冊那一句把三種對上了**(2026-08-07)。


        第 33 項(地面戰結構)留下「四種部隊類型沒有對出名字、每種的攻擊力表還沒追」。同一天追完。

        ### `Compute_Ground_Combat_Info_` @ 0xEC3CE 的四個 case
        ```
        case 0:  攻擊力 += 10 + 加成塊[+1] ; 耐受 += 1 + 加成塊[+2]
        case 1:  攻擊力 +=      加成塊[+3] ; 耐受 +=     加成塊[+4]
        case 2:  攻擊力 -= 10
        case 3:  攻擊力 -= 20              ; 基礎值取自**另一方**的加成塊(為 0 時整格歸零)
        ```
        `Compute_Colony_Ground_Combat_Info_` @ 0xED713 給殖民地填**三格**數量
        | 類型 | 調整 | 是什麼 |
        |---|---|---|
        | 0 | +10 攻擊、+1 耐受 | **裝甲**(手冊:tank battalions) |
        | 1 | 基準 | **陸戰隊** |
        | 2 | −10 攻擊 | **民兵**(未受訓的平民,最弱) |
        | 3 | −20,基礎取自另一方 | ⚠ **仍未定名**,殖民地不填它 |
        - **只實作立即數的部分**。`加成塊[+1]`/`[+3]`(科技加成)那兩欄還沒對出意義,
        - **守方的民兵沒有接**:數量公式在 `sub_EC61E`,還沒追;AI 也沒有裝甲營房的追蹤機制。


    ### **民兵接上了——`Colony_N_Militia_` 是「除以 5」**(2026-08-07)。


        第 33 項(地面戰結構)留下「守方的民兵沒有接:數量公式在 `sub_EC61E`,還沒追」。同一天追完。

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
        - `擁有者 >= 8`:玩家編號只有 0..7。`Init_Homeworld_Colony2_` 寫入時就是
        `DefenderStart` 的回報數字也跟著含民兵——不改的話畫面上會少報守方兵力。
        ⚠ **裝甲那一格仍留 0**:AI 沒有 `ColonyBuildings` 追蹤機制,無法判斷「AI 是否已建成
        (`TestInvadeColony_StrongAttackerWinsMost` / `..._StrongDefenderWinsMost`)**仍然綠**,


    ### **地面戰的加成塊——其中一條手冊完全沒寫:難度加成不給玩家**(2026-08-07)。


        第 33 項(地面戰結構)留下「那幾個加成塊欄位還沒逐欄對出意義」。追
        `Compute_Player_Ground_Combat_Bonuses_` @ 0xEC15C(產一個 **19 位元組**的加成塊)之後,
        大多數欄位對應的是手冊已經列出的加成類別(remake 的 `GroundArmorTechBonus` 等已經涵蓋),
        但**有兩條是手冊沒有的**。

        ```
        [out+0x0C] = 1  if [player+0x8AA] != 0
        耐受命中數 = [加成塊+0x0C] + 1          ; Compute_Ground_Combat_Info_
        ```
        ```
        if ([player+0x28] == 100) [out+0x0F] = 0        ; 100 = 人類玩家的標記
        else                      [out+0x0F] = 難度 − 2
        ; 玩家編號 >= 8 的那一側(安塔蘭 / 怪獸)走另一條路徑:
                                  [out+0x0F] = 難度×2 − 4
        ```
        | 參戰者 | 加成 |
        |---|---|
        | 人類玩家 | **0** |
        | AI 帝國 | 難度 − 2(普通 = 0、不可能 = +2、教學 = −2) |
        | 安塔蘭那側 | 難度×2 − 4(**恰好是 AI 的兩倍**) |
        - **加成是以「普通」為基準往兩邊偏**,不是「難度越高一律加成」。教學難度下 AI 是**負的**。
        - `[player+0x28] == 100` 這個人類玩家標記在 `Init_Player_Tech_` @ 0x5E55F 也出現過
        `[+5]`/`[+7]`/`[+9]` 那三張查表(stride 15 / 3 / 3,索引由 `sub_DC323`/`sub_DC416`/`sub_DC449`


    ### **分兵種接線 + 手冊的四個 hits 數字全部被重建出來**(2026-08-07)。


        第 35 項(三張查表)留下「Battleoids / 動力裝甲的分兵種接線留給下一輪」。接完了,而且過程中發現
        **手冊列的那四個 hits 值,可以完全由反組譯的加法結構重建**。

        | 手冊值 | 反組譯的組成 | 結果 |
        |---|---|---|
        | 陸戰隊 1 | 基礎 `[+0x0C]+1` = 1 + 類型 1 的 delta 0 | **1** |
        | 陸戰隊 + 動力裝甲 2 | 1 + `[out+4]` = 1 | **2** |
        | 戰車營 2 | 1 + 類型 0 的 delta 1 | **2** |
        | 機械戰士 3 | 1 + 1 + `[out+2]` = 1 | **3** |
        | High-G +1 | `[+0x0C]` = 1 → 基礎變 2,全類型各 +1 | ✓ |
        落成 `TestManualHitValuesReconstructFromTheOriginalStructure`。
        remake 的 `tankHitsToKillFor` 回的是**手冊的成品值**(戰車 2、機械戰士 3),
        而第 33 項(地面戰結構)接上的 `GroundTypeHitsDelta` 是**組成的一部分**。兩個一起用就變成 3 / 4。
        | 加成 | 原版寫進 | 被誰讀走 | remake 先前 |
        |---|---|---|---|
        | Anti-Grav Harness +10 | `[out+0]` | 所有類型 | ✓ 對 |
        | Personal Shield +20 | `[out+7]/[out+8]` | 所有類型 | ✓ 對 |
        | **Powered Armor +10** | `[out+3]` | **只有陸戰隊** | ✗ 加給整支部隊 |
        | **Battleoids +10** | `[out+1]` | **只有裝甲** | ✗ 加給整支部隊 |
        已拆成 `groundMarineOnlyBonusFor` / `groundTankOnlyBonusFor`,共用的那份不再包含它們。

34. **重力種族特性:High-G 逐字對上,Low-G 的「10%」其實是定值 −10**(2026-08-07)。

    第 33 項(地面戰結構)留下的加成塊欄位又追出三個,而且三個都能與手冊互證。

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
    它與其他所有加成一起加進攻擊力(`var_4`),而那些加成本身也都是 +10/+15/+20 這種定值。
    ```
    cmp byte ptr [player+8ACh], 0     ; Subterranean
    jz  short loc_EC247
    cmp [ebp+var_4], 0                ; ← 呼叫端傳的旗標
    jz  short loc_EC247
    mov byte ptr [out+0Eh], 0Ah       ; → +10
    ```
    「只有守方才給」在原版是那個呼叫端旗標,而 `Compute_Colony_Ground_Combat_Info_`
    `GroundSubterraneanDefenseBonus`——這一條從「手冊單一來源」升級為雙來源,沒有改動。
    `[player+0x8A7]`(有號,加進所有類型)看起來就是種族的地面戰加成
    (remake 的 `GroundRaceCombatBonus`:布拉西 +10 / 諾蘭姆 −10),但**沒有直接證據**

35. **三張查表讀出來了——十二個科技 id 全部對上,而且 remake 少了一整條通道**(2026-08-07)。

    ### 索引函式的符號名直接說了是什麼

    | 函式 | 表 | 步幅 | 階數 |
    |---|---|---|---|
    | `Player_Best_Armor_` @ 0xDC323 | `word_17F63E` | 15 | 6 |
    | `Player_Best_Rifle_` @ 0xDC416 | `word_14A88` | 3 | 5 |
    | `Player_Best_Personal_Shield_` @ 0xDC449 | `word_14A9A` | 3 | 1 |
    三支都是**從表尾往前找第一個已知的科技**(`[player + 科技 + 0x117] == 3`)——
    `aMultigmLbx` 後面緊接著 `byte_17A061`,所以那個字串的 VA = `0x17A061 − 12`;
    在 exe 裡搜 `"MULTIGM.LBX\0"` 得到檔案位移 `0x1F86E9` → **delta = 0x7E694**。
    用另一個同名字串(`aMultigmLbx_0`)反推得 VA `0x178004`,落在 `;org 178000h` 之後 4 位元組
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
    - **鈦裝甲 +5 少了**。手冊沒列基礎裝甲的地面加成,remake 就回 0。
    - **整條步槍通道 remake 完全沒有**。上限差 **30 點**——後期科技全開的帝國,
      remake 的地面部隊比原版弱整整 30。已補 `GroundRifleTechBonus` + `groundRifleBonusFor`
    加成塊的另外三個科技旗標也解出來了(`[player + 科技 + 0x117]`,減去 0x117 就是科技 id):
    | 位址 | 科技 id | 科技 | 寫進哪一欄 | 被誰讀走 |
    |---|---|---|---|---|
    | `[player+0x120]` | 9 | `TECH_ANTIGRAV_HARNESS` | `[out+0]` | **所有類型**共用的基礎 |
    | `[player+0x12F]` | 24 | `TECH_BATTLEOIDS` | `[out+1]`/`[out+2]` | **只有類型 0(裝甲)** |
    | `[player+0x1A7]` | 144 | `TECH_POWERED_ARMOR` | `[out+3]`/`[out+4]` | **只有類型 1(陸戰隊)** |
    常數與依據已記進 `gamedata`(`GroundBattleoidExtraHits`、`GroundPoweredArmorAppliesTo`),
    `[player+0x8A7]`(有號,加進所有類型)與 `[+0x10]` 仍未定名。

36. **聊天列補完:14 則的環、82 byte 的格子、兩種前綴,四個數全是一手的**(2026-08-07)。

    「等待其他玩家」那張畫面的輸入列先前是一條寫著「remake 未實作」的提示帶。
    做得動的理由是第 29 項(決定性化)把文字輸入框做出來了,而這一輪把原版那條線整條讀完:

    | 函式 | 位址 | 讀出什麼 |
    |---|---|---|
    | `Chat_Box_Input_Loop_` | 0xF55A4 | 點輸入列 → 進聊天模式;非空才送;送完清空重新武裝 |
    | `Send_Chat_Msg_` | 0xDD3B8 | 逐一走玩家陣列(stride 0xEA9),`edx = 27h` 是封包型別 |
    | `Receive_Chat_Msg_` | 0xDD351 | 環的結構:滿 14 則 memmove 掉最舊、每則 82 byte |
    | `sub_F1075` 的繪製段 | — | 兩種前綴、行距 12、x +24、首行 y +14 |
    | 值 | 出處 | 意思 |
    |---|---|---|
    | 14 | `cmp dword ptr [eax+47Ch], 0Eh` | 保留幾則 |
    | 82 | `imul edx, [eax+47Ch], 52h` | 每則佔幾 byte |
    | 80 | 82 − 發話者 1 − NUL 1 | 一則最多幾個字元 |
    | 8 | `cmp ax, 8 / jge`(繪製段) | 到這個編號以上是 GNN 新聞不是玩家 |
    計數欄位落在 `[+0x47C]` 自己就是第二條線:**14 × 82 = 1148 = 0x47C**
    繪製段給的是相對偏移(x +0x18、首行 y +0x0E、行距 0x0C、每行擦 0x23A × 0x0B),
    ```
    首行 257 → 14 行 × 12 → 最後一行 413,底部 424
    Add_Net_Next_Turn_Fields_ 給的輸入列:430
    ```
    結果嚴絲合縫——這是第二個獨立來源。落成 `TestChatLayoutFitsAboveTheInputRow`。
    ```
    發話者 < 8  → "(%s)  %s"      ; 右括號後**兩個**空格
    發話者 ≥ 8  → "( GNN )  %s"   ; 括號內側各一格,固定色 byte_199F34 = 0x10
    ```
    UTF-8 切在半個中文字上會變亂碼。`ChatTruncate` 守住 80 byte 但截在 rune 邊界
    - `Send_Chat_Msg_` 只發給 `[player+0x28] == 'd'` 的玩家。**沒查到那個欄位的寫入端**,
    - 送出目前**只進本機記錄**:鎖步的 `netplay.Table` 一回合只收一則,聊天是隨時可送的,
      `ChatLog` 這一端不必動。
    - 玩家列的顏色仍是 remake 自訂的兩色,沒接 `Get_Net_Next_Turn_Player_Colors_` @ 0xF31BB
    比對時發現 `docs/screenshots/` 只有 27 張,而 gallery 產 35 張——

37. **整棵研究樹從二手轉寫升格成一手驗證過**(2026-08-07)。

    `techtree.go` 的檔頭一直寫著「逐字轉寫自 openorion2 的 `tech.cpp`」。也就是說
    remake 的 83 個主題成本、每個主題的可選科技,證據等級**全部是二手**
    ——openorion2 讀錯了 remake 就跟著錯,而且沒有任何測試會發現。

    起點是第 35 項(三張查表)刻意延後的 `Choose_Tech_Application_` @ 0xFD335。讀它的過程中
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
    | 項目 | 結果 |
    |---|---|
    | 83 個主題的成本 | **74 個逐字相同**,9 個有差(全部有解釋) |
    | 主題 → 科技的歸屬 | **81 個主題完全相同**,逐科技 **199 條吻合** |
    | remake 有、原版沒有的科技 | **0 個** |
    | 領域表 vs next 鏈 | **73 條銜接關係全中** |
    ```
    1996 原版      → 15000
    patch 1.31     → 15000   (整張表與 1996 版 byte-identical)
    patch 1.50.26  → 25000   (只有這 8 筆變,主題 74 仍是 15000)
    ```
    remake 的 `RuleProfile` 早就寫著 1.3 = 15000、1.5 = 25000,出處是社群
    `CHANGELOG_150.TXT` 1.50.9「Hyper-Advanced Tech Cost Bug」。現在那兩個數字
    ```
    next[74] == 74          ← 後繼是它自己
    沒有任何別的主題的 next 指向 74
    ```
    `shell.StarterResearchTopics` 是一份手挑的 9 個「新手可選的早期主題」。
    ```
    不該出現(領域裡前面還有沒研究完的):進階建築學、進階生物學、人工智慧、進階化學
    漏掉了(該領域的隊首):        太空生物學、核融合物理學、光電子學
    ```
    `AvailableResearchTopics(session)` 由樹算出來(`gamedata.AvailableTopics`)。
    ⚠ `-game` 主路徑的 `currentAreaTopic` 本來就是對的(2026-07-11 修盲選 bug 那次改的),
    這次的 `next` 鏈剛好是它的一手佐證;錯的是 `play.go` 那張簡易殼的研究畫面。
    `sub_FD2F9` 檢查六個硬編位址是否都等於 3,全中才回 1。那六個位址減掉 `0xC4`:
    ```
    0xDA→22  0xDB→23  0xE0→28  0xE1→29  0xFB→55  0xFD→57
    ```
    **正好是第 30 項(TECH LEVEL 第二效果)從 `word_18111C` 挖到的六個開局主題**,第三個獨立來源。
    順帶對上第 33 項:`OrigStartingTechs` 裡有 `TECH_PULSE_RIFLE`,而 remake 的
    `GroundRifleTechBonus` 把脈衝步槍定為 +0 基準點——先前的理由只是「它最低階」,
    `sub_FC845`(逐科技估值,`sub_FD335` 的權重來源)**是 985 行**。上一輪判斷「一次讀就
    其餘(候選集合、成本反比、視野放寬、加權隨機)全部照抄。`sub_FC845` 仍不照抄。

38. **三面行星護盾 + 自動實驗室 + 再生反應爐接線**(2026-08-07)。

    HONEST-STATUS 寫著「部分軍事/防禦建築(~13 棟,需艦隊駐防/軌道防禦系統先落地)」。
    照 rulebook 63 對程式碼盤點(掃建築名有沒有在 `buildings.go` 以外被消費過),
    實際是 **11 棟**,而且其中**有三棟根本不需要新子系統**——它們接的軌道轟炸早就有了。

    `fleetBombardDamage` 的註解自己講了:
    | 建築 | 手冊原文 | 減傷 | 手冊維護費 | 建築表(來自執行檔) |
    |---|---|---|---|---|
    | Planetary Radiation Shield | reducing bombardment damage by 5 points | 5 | 1 BC | **1** |
    | Planetary Flux Shield | reduces all damage … by 10 points **per attack** | 10 | 3 BC | **3** |
    | Planetary Barrier Shield | reducing all damage … by 20 points **per attack** | 20 | 5 BC | **5** |
    接在**逐發**傷害(`shot.DamageToStructure`)而不是總傷害。多次攻擊下這兩個接法
    差一個數量級；2026-08-24 已另由 IDA 證實 runtime 結果採除以 40，手冊除以 100 僅屬顯示估算。
    所以測試釘的是 `TotalDamage`：原版固定三外圈中的每次攻擊都先套減傷，手算得出來。
    Radiation Shield already in existence」),所以 `PlanetaryShieldReduction` 取**最強的那一面**。
    第二句決定了接的位置。`FlatIndustry` 是在污染縮減之**前**併進 gross 的
    改成一個旗標,在 `RunColonyTurn` 的污染切分點之後才加。
    同樣多出來的產能若接成 `FlatIndustry`,污染一定會變——那條在,才證明「污染沒變」
    所以只動 `FlatResearch`。
    - `30_netwait.png` 變了:狀態指紋是存檔快照的 SHA-256,`ColonyState` 多一個欄位就會變。
      那是 `determinism.go` 註解寫明的設計(「新增欄位只要進得了存檔就自動進得了指紋」),


    ### **食物複製機接線:一句話裡的三個限定詞,漏一個就是印鈔機**(2026-08-07)。


        手冊 p.85 一整句就是完整規格:

        | 限定詞 | 規則 |
        |---|---|
        | two-for-one | 2 產能 → 1 食物 |
        | 1 BC per food | 換出來的每單位食物再花 1 BC(從國庫,不是從產能) |
        | **as needed** | **只補足缺口**,不換出盈餘 |
        **餘糧出售**(`IncomeFoodSurplusRevenue`,每單位 0.5 BC)換回 BC。
        | 來源 | 維護費 |
        |---|---|
        | 手冊 p.85 | 10 BC |
        | remake 建築表(原版執行檔 `off_17EB3D + 12`) | **10** |
        `RunColonyTurn` 裡,**污染扣完之後、人口成長之前**:換算要用的是可用產能(所以在污染之後),
        BC 成本走 `EmpireOutput.FoodReplicatorCost`,和其他維護費一起進 `NetBC`。
        `session.go` 有一段寫著「其餘 20 項(…再生反應爐、食物複製機、自動實驗室、


    ### **阿提米絲系統網:水雷子系統**(2026-08-07)。


        上一項把它列在「真的缺子系統」那一欄。手冊 p.86 其實把整個子系統寫完了:

        | | 規則 |
        |---|---|
        | ① 觸發 | 逐艦擲,依艦體等級 20 / 30 / 40 / 50 / 80 / 100 % |
        | ② 水雷數 | 中招的船各擲一次,8–28 枚 |
        | ③ 每枚傷害 | 20 − 護盾等級 |
        - **艦體等級**:`shipStrength` 的六個類別(巡防/驅逐/巡洋/戰艦/泰坦/末日之星)
        - **護盾等級**:remake 的護盾元件叫「第一級護盾」「第三級護盾」…「第十級護盾」
        所以掛在 `advanceFleet` 的抵達那一段,順序放在**探索標記之後、一次性發現之前**:
        - **只對玩家艦隊生效。** AI 沒有「艦隊移動到某顆星」的模型(它的攻擊是抽象解算的),
          ——AI 有真的艦隊移動時,`applyArtemisMines` 是它唯一的掛勾點。

39. **艦員經驗系統：2026-08-07 建模，2026-08-24 補齊每回合原版鏈**。

    上一項把太空學院列在「缺艦員經驗值子系統」。盤點之後發現 remake 的狀況很特別:
    **加成表已經有三張,而且都對得上手冊**——

    | 表 | 位置 | 手冊欄 |
    |---|---|---|
    | `shipCrewOffenseBonuses = {0,15,30,50,75}` | `formulas.go` | BA |
    | `shipCrewDefenseBonuses = {0,15,30,50,75}` | `formulas.go` | BD |
    | `MissileCrew* = 0,7,15,25,37` | `missile.go` | ME |
    2026-08-07 盤點當時尚無艦艇等級；其後已加入 `Ship.CrewXP`，三張表也已接到
    `engine.ShipBeamAttackFromDesign` 等攻防消費端。2026-08-24 再以 IDA Pro 9.4 確認
    `Do_All_Ships_XP_Check_ @ 0x14A27` 的每回合來源：固定掃描 500 筆、接受 `Status < 5`、
    `Owner < 8`，每艦 +1、同星系每座己方太空學院 +1、取活動領袖最強 Instructor，結果
    封頂 500；詳見 `ship-crew-xp-audit-20260824.md`。同日再閉合不同入口
    `sub_4B184 @ 0x4B184`：戰後每艘 winner-side linked Ship 取得
    `max(1, floor(被摧毀敵艦艦體級總和/2))`，直接寫 `Ship+0x72`，不套每回合 500 cap；
    詳見 `ship-battle-crew-xp-audit-20260824.md`。
    ```
    一般種族   Green(0) → Regular(50) → Veteran(150) → Elite(500) → ✗
    統帥種族   ✗        → Regular(0)  → Veteran(50)  → Elite(150) → Ultra-Elite(500)
    ```
    `Ship` 只加了 `CrewXP` 一個欄位,等級一律由它現算。存兩個欄位遲早會不同步
    起始 `CrewXP` 設成那一級的門檻,而不是另開一個「起始等級」欄位。
    - **halved**:打小船升得慢(擊沉一艘巡防艦 1/2 = 0 → 保底 1)
    - **destroyed not captured**:俘虜不算
    - **minimum 1**：IDA `0x4B913..0x4B919` 證實總和為 0 時仍強制為 1；舊的
      「零擊沉為 0」是手冊延伸解讀，已由 executable consumer 推翻。
    還原「被擊沉的是哪些」用的是多重集合相減:`battleVolley` 就地移除陣亡者,
    呼叫端拿不到「誰死了」,但敵艦的 `atk` 就是戰力值、戰鬥中不變,
    remake 的 `shipStrength` 是 2 的冪(巡防 2、驅逐 4、巡洋 8、戰艦 16、泰坦 32、末日 64),
    ### 順帶收掉五個各自硬寫的 `false`
    `gamedata.Ground*BarracksCap` 早就有 `warlord` 參數(統帥種族營房容量加倍),
    但 shell 有**五處**各自硬寫 `false`。這次加的 `GameSession.RaceWarlord` 把它們統一
    - **只有玩家的船有艦員經驗。** AI 的艦隊在 remake 裡是每回合現生的戰力值
      (`genEnemyFleet`),沒有持久的船,自然沒有可累積經驗的對象。
      逐殖民地造艦那條路(`deliverNewShip`)是正確的。

40. **征服人口的同化系統**(2026-08-07)。

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
    - **排斥 Repulsive**:「assimilate … at only **half** the normal rate」→ 回合數 ×2。
    - **魅力 Charismatic**:手冊只說「assimilate conquered colonists **easily**」,
      `AssimilationTurns` 收了那個參數但不用它。
      這裡刻意寫了一支測試 `TestCharismaticHasNoQuantifiedEffectYet`:
    `TOPIC_ADVANCED_GOVERNMENTS` 這個四選一主題底下的四個科技。remake 的政府選單只有
    判定沿用 `groundEquipTechOwned` 那組既有規則,不另立一套。
    `session.go` 寫著「異族管理中心:士氣計算路徑已預留(**colonyMoralePercent 讀取此建築名**)」。
    **`colonyMoralePercent` 根本沒有讀它**——那個建築名在整個 repo 裡只出現在資料表與註解裡。
    - **未同化人口目前沒有負面效果。** 手冊說多種族殖民地有 20% 士氣懲罰(建築可消除)、

41. **戰機基地 + 行星版恆星轉換器,以及一個盤點方法的錯誤**(2026-08-07)。

    ### 兩棟建築,同一個接點

    remake 早就有「殖民地被軌道轟炸時反擊」那條路徑(`retaliationAttackers`),
    寫戰機基地的分檔判定時,我把兩個科技都寫成 `TOPIC_ADVANCED_ROBOTICS`。
    拿第 37 項(研究樹一手驗證)挖出來的 `OrigTechTopic` 查一次:
    ```
    TECH_BOMBER_BAYS        (31) → 主題 11 = TOPIC_ADVANCED_ROBOTICS          ✓
    TECH_HEAVY_FIGHTER_BAYS (83) → 主題 42 = TOPIC_SUPERSCALAR_CONSTRUCTION   ✗ 我寫錯
    ```
    加完恆星轉換器之後,既有測試 `TestStellarConverterAddsColonyDefense` 紅了:
    「防禦 +1200,want +800」。查下去發現**它早就接過了**——在 `colonyDefense`
    (ai_attack.go)裡,用的是常數 `gamedata.StellarConverterName` 而不是字面字串。
    1. **同一棟建築在兩條路徑上行為不一致**:它擋得住 AI 來襲(`colonyDefense` 有算),
       卻對軌道轟炸完全不反擊(`retaliationAttackers` 沒有它)。已統一到
       `retaliationAttackers`,`colonyDefense` 那邊的獨立加總移除。
    2. **`StellarConverterDefense = 800` 的來歷自己就矛盾**:註解寫著
       已改成 `StellarConverterDamagePerSide`(400):remake 的防禦解算是抽象的、
    `buildings.go` 以外的程式碼」。那個掃法有一個盲點:**在 `buildings.go` 內部宣告成常數、
    ```
    建築表 41 棟 → 字面字串消費 40 棟 / 經常數消費 1 棟 / 完全未消費 0 棟
    ```
    那不是手冊的意思(手冊只給中隊數,沒說哪一檔比較強),是 `combat.go` 裡兩個標明過的

42. **把上一輪自己寫的兩條留白關掉**(2026-08-07)。

    第 41 項(戰機基地)的結尾特別強調「**有程式碼消費不等於完整還原**」,並列出仍有寫明的部分實作。
    其中兩條在寫的當下就已經解得開了——擋住它們的東西是我自己前幾輪剛加上去的。

    它一直沒有呼叫端,理由寫在 `session.go` 的註解裡:「因 remake 無多種族人口追蹤,
    第 40 項(同化系統)加上 `ColonyState.UnassimilatedPop` 之後,那個理由就不成立了:
    - 同化完最後一單位的那一刻懲罰消失——`advanceAssimilation` 每輪重算士氣,
    ```
    Radiation Shield 「Radiated worlds become Barren as long as the shield remains in place」
    Flux Shield      「The existence of a flux shield converts Radiated climates to Barren」
    Barrier Shield   「This shield converts Radiated climates into Barren」
    ```
    `ColonyState.Climate` 早就在(地形改造那一輪加的),所以只要在建成時走既有的
    `applyClimateChange`——那一支會連帶調整食物與人口上限,直接改 Climate 欄位不會。

43. **先進級開局的 19 個隨機主題:照抄結構,不照抄權重**(2026-08-07)。

    這個缺口從第 30 項(TECH LEVEL 第二效果)就開著,第 88、91 兩次判斷「不照抄」,理由都是同一個。
    這次換個問法:**擋住的到底是哪一部分?**

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
    - **只從「現在可研究」的主題挑**,所以是**沿著樹往上走**而不是從同一池子抽 19 次
    - **偏好便宜的**,與原版的評分方向一致
    - **決定性**:同一顆種子重開得到同樣的開局;玩家與每個 AI 各走一條獨立的流
    第 37 項(研究樹一手驗證)挖出來的 `OrigTopicCost`(83 個主題的一手成本)在這裡是燃料——
    - **`weight` 一律 1**,少的是那 985 行的估值。AI 因此不會偏好「對它有用」的科技。
    - `sub_FD335` 尾巴還有一段依 `[player+0x28]`(值 1/2/4/5)的二次過濾。
      那個欄位**沒查到寫入端、沒有名字**——第 36 項(聊天列補完)在 `Send_Chat_Msg_` 裡也遇到同一個。
    `TOPIC_STARTING_TECH`——第 37 項(研究樹一手驗證)認出來的那個「自環、永遠不會被解鎖」的容器主題,

44. **上游補完之後,下游要跟著讀真的東西**(2026-08-07)。

    第 43 項(先進級開局主題)把 19 個隨機主題發出去了,但先進級的母星建築**還是只有兩棟**。
    原因很直接:`homeworldBuildingsFor(techLevel, pop)` 從**固定表**現算科技集合,
    **看不到那 19 個**。

    `CompletedTopics`,`applyStartingBuildings` 改走它。
    | TECH LEVEL | 開局主題 | 母星建築 |
    |---|---|---|
    | 曲速前 | 1 | **2**(海軍陸戰隊營 + 星基) |
    | 一般 | 6 | **2**(同上) |
    | 先進 | 25 | **6**(戰鬥站、星基、水耕農場、海軍陸戰隊營、生態圈、自動工廠) |
    - `gamedata/starting_tech.go` 兩處(`StartingTopics` 的 ⚠、`StartingTopicRandomExtras` 的來歷)
    全部改成現況。`TestAdvancedStartIsBlockedByTheMissingRandomTopics` 也改名為
    `TestAdvancedStartFillsAllBuildingSlots` 並反轉斷言——**它自己當年就寫著

45. **領袖技能:修掉一條疊加規則,補上四個技能,並確認一個「不是缺口」**(2026-08-07)。

    HONEST-STATUS 一直寫著「MOO2 25 個領袖技能,remake 接了 2 個」。這一輪逐條查手冊。

    `applyLeaderColonyBonuses` 先前是無條件 `+=`——**兩個貿易家就加兩份**,而原版只算最強的
    那一位。已改成先依技能分組、再依「累加 vs 取最佳」合成(`gamedata.LeaderSkillCombine`)。
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
    | 技能 | 承接欄位 |
    |---|---|
    | 財務官 +10% | `ColonyState.IncomeBonusPercent` |
    | 心靈導師 +5% | `ColonyState.MoralePercent` |
    | 醫官 +10% | `ColonyState.GrowthBonusSum` |
    | **教官 +1** | **艦員每回合經驗**(第 39 項(艦員經驗)才有的東西) |
    `ColonyState` 只有 per-worker 與固定值兩種欄位)。
    查這一輪時我一度說「`loadHerodataMercs` 沒有呼叫端、真英雄池從沒裝進遊戲」——**錯的**,
    它在 `interactive.go:4384` 有呼叫,是我的 grep 被 `head` 截掉了。


    ### **分項百分比:三個 admin 技能缺的只是一個欄位**(2026-08-07)。


        上一項把農業官 / 勞工官 / 科學官擋在門外,理由是「食物/工業/研究的**分項百分比**——
        `ColonyState` 只有 per-worker 與固定值兩種欄位」。回頭看那個理由:**缺的只是三個欄位**,
        而引擎早就有百分比進得去的地方。

        ```go
        pct := cs.MoralePercent + colonyGravityPenaltyPercent(cs)
        food     = GravityAdjustedProduction(Farmers*FoodPerFarmer, pct) + FlatFood
        gross    = GravityAdjustedProduction(Workers*IndustryPerWorker, pct) + FlatIndustry
        research = GravityAdjustedProduction(Scientists*ResearchPerScientist, pct) + FlatResearch
        ```
        第三支釘住「固定加成不吃百分比」(農夫為 0 時食物全來自 `FlatFood`,
        | 技能 | id | 格式 | 落在 |
        |---|---|---|---|
        | 科學家 Researcher | 6(common) | `%+d` | `FlatResearch`(固定點數) |
        | 科學官 Science Leader | 38(admin) | `%+d%%` | `ResearchBonusPercent`(百分比) |
        - **環保官**:降低「會產生污染的產能」的百分比——remake 的污染模型是 `PollutionEighths`
        - **工程師**:艦艇維修**速率**加成——remake 的 `advanceShipRepair` 是照原版
          `Repair_Ship_Full_` 做的**一次修好**,沒有「速率」這個量可以加成。
        `repairAfterBattle` 已於第 45 項(領袖技能)接上。
        - **戰術官**:原版自己就沒實作(第 45 項(領袖技能))。


    ### **HERODATA 的技能欄位是每技能 2 bit,不是 1 bit**(2026-08-08)。


        第 101、102 項把 admin 技能一個個接上去,但**真英雄一個都拿不到**——上游把技能位元讀錯了。

        ```cpp
        skillnum = id & SKILLCODE_MASK;
        if (skillnum >= max) return 0;
        return (skills >> (2 * skillnum)) & 0x3;
        ```
        `cmd/moo2/herodatamercs.go` 讀的是:
        ```go
        skillCommonResearcherBit = 1 << 6 // SKILL_RESEARCHER
        skillCommonTraderBit     = 1 << 9 // SKILL_TRADER
        ```
        解碼函式 `gamedata.LeaderSkillTier` 早就照 `hasSkill` 寫對了——**只是這裡沒用它**。
        | # | 症狀 | 後果 |
        |---|---|---|
        | 1 | `Tier` 寫死 1 | 進階技能的 +50% 一次都沒發生過 |
        | 2 | 一位英雄只給一項技能 | 原版一人可有多項(所以才要 2 bit × N 欄位) |
        | 3 | 艦艇軍官通稱「指揮官」 | 那是 SKILL_COMMANDO 的譯名,`commandoLeaderTier` 掃字串 → **每一位**雇來的艦艇軍官都吃到地面戰加成 |
        | 4 | 效果查表用**中文標籤** | 英文模式下 `Skill` 是 "Scientist",查不到 → **所有領袖加成同時消失**,畫面無異狀 |
        (`leaderSkillIDByName`、`commandoLeaderTier`、`FleetHasNavigator`),三處都會在切成英文的
        - `gamedata/leader_skill_names.go`:27 個技能的 id ↔ 中英文名(英文名逐字取自
        - `shell.Leader` 加 `Skills []LeaderSkill{ID, Tier}`,**id 是識別鍵、標籤只負責顯示**。
          沒有 `Skills` 的舊資料(demo 領袖、既有測試、舊存檔)退回用中文標籤反查,相容不變。
        「戰後完全修復」正是 `shell.repairAfterBattle` 在做的事——它本來就有兩個觸發
        (自動修復元件、進階損害管制),加第三個只是多一個條件。`won` 這個新參數**只**影響
        軍官也沒有指派到艦隊,沿用 `commandoLeaderTier` 的既有近似(帝國內有就算數)。


    ### **環保官:第三次用同一種方式擋錯**(2026-08-08)。


        第 45 項(領袖技能)擋掉農業官/勞工官/科學官,理由是「沒有分項百分比欄位」——第 45 項(領袖技能)發現缺的只是三個欄位。
        第 45 項(領袖技能)擋掉工程師,理由是「維修速率那個量不存在」——第 45 項(領袖技能)發現手冊那條的第二句對得上既有函式。
        這一輪輪到環保官,擋門理由是「污染模型是八分之幾的查表,沒有百分比入口」。

        「the amount of production that causes pollution」——`PollutionPollutingProduction`
        ```
        eighths 查表 → 環保官百分比 → 減容忍值 → 一半 = 清理成本
        ```
        ⚠ 這**不是減產能**。接到 `gross` 那一行會變成「請一個環保官等於少一個工人」,
        有一支測試專門把這條算術釘在 `PollutionEighths` 上:哪天有人把它改成相加,那裡會先響。
        環保官的 `skillBonus` 是**負值**(base −10),而 `ColonyState` 那一欄存的是**正的減幅**

46. **叛亂:同化計時器的另一半**(2026-08-08)。

    第 40 項(同化系統)接了同化,而那個檔的檔頭自己寫著「叛亂系統根本不存在 …
    現在同化只是一個會走完的計時器——**機制在、後果還沒接**」。這一項接後果。

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
    ### `[colony+0x137]` 是異族管理中心——三個呼叫端對上手冊的三句
    | 呼叫端 | 條件 | 手冊 |
    |---|---|---|
    | `Check_Rebellion_` | `!= 0` → 機率減半 | 「halving the chance of revolt」 |
    | `Apply_Assimilation_` | `!= 0` → 速率固定 120 | 「1 per 2 turns, **regardless of government**」 |
    | `sub_DDAD4` | `== 0` 才算多種族 | 「removes the 20% morale penalty from multi-racial colonies」 |
    ### 順帶定名:`GroundTypeFourth` 就是叛軍
    `Get_Rebellion_Info_` @ 0xEC65A 把守方三種部隊填進 `[+0x0A]`(裝甲)、`[+0x0C]`(陸戰隊)、
    也合理——起事的是沒受過訓練的被征服人口。常數已改名 `GroundTypeRebels`。
    - **鎮壓後死多少人**沒有依據。原版那段在 `Resolve_Rebellion_Troops_`,沒有逐指令抄;
    - **沒有滅絕政策**:remake 沒這個選項,「×2」那一路目前不會發生
    - `[colony+0x12F]` 的完整列舉沒查(已知 4 = 未征服、0 = 滅絕,1/2/3 語意未定)。
    - **只檢定玩家的殖民地**:AI 打下玩家殖民地那條路徑本來就沒有同化模型。
    沒有繼續敲——改讀 `Orion2.exe.asm` 原始清單 + `symbols_fixed.tsv` 對名字,

47. **AI 艦隊會在星圖上移動了**(2026-08-08)。

    先前 AI 的突襲是**瞬移**的:`aiRaid` 直接拿 `aiRaidWilling` + `aiRaidTarget` 結算,
    AI 沒有位置,艦隊憑空出現在玩家殖民地上空。三個後果:

    1. **玩家看不到它來**——無法預警、無法攔截,只能事後看回合摘要。
    2. **阿提米絲系統網打不到 AI**。那棟建築的手冊效果是「任何**進入**該星系的敵艦」,
    `AIOpponent` 加上位置與航線(`FleetStar` / `FleetPosSet` / `FleetDestStar` / `FleetETA`),
    `advanceAIFleets` 每回合推進,突襲的前提從「**想打**」改成「**打得到**」
    (原版三層 `Colony_Worth_To_Player_` / `Enemy_Colony_Worth_To_Player_` / 距離)原封不動,
    **② 四個守門測試會假綠。** `TestAIRaidGracePeriod` / `NeedsWarStance` /
    `NeedsStrengthAdvantage` / `PacifistNeverRaids` 原本呼叫 `advanceAIRaids`,
    四支都改成驗出發那一端(`advanceAIFleets` 後有沒有 `FleetETA > 0`)。
    `aiSnapshot` 是**逐欄位手抄**的,加新欄位不會自動進存檔。第一次跑截圖廊時
    (指紋 = 存檔快照 JSON 的 SHA-256)。補進 `aiSnapshot` 與 `restore` 之後指紋才變。
    - **一個 AI 只有一支艦隊**:它的軍力是單一整數 `FleetStrength`,不是艦艇清單。
    - **水雷對 AI 是戰力折損,不是逐艦模擬**:原版逐艦擲觸發率(依艦體等級 20–100%)、
    - **艦員經驗仍拿不到**(同一個原因:沒有逐艦資料)。
    - **航線不判星雲/黑洞/干擾場**:玩家那條路徑模型綁在 `s.Player` 上。
    - **AI 之間不互相出兵**:需要 AI-vs-AI 戰鬥解算,remake 沒有。

48. **三個「待辦」複查:一個早就做完了、一個是證據不足、一個追到源頭**(2026-08-08)。

    WORKLIST 剩餘清冊裡的 F / H / I 三項,這一輪逐一複查而不是逐一開工。

    ```go
    // internal/shell/session.go, playerHomeworldColony
    Farmers: 4, Workers: 2, Scientists: 2,
    ```
    ```
    internal/shell/orbital_bombardment.go:218
        atk, _ := gamedata.FighterGarrisonCombatContribution(fighterGarrisonTierFor(defender))
    ```
    `fighterGarrisonTierFor` 只看科技,回傳 10/6/4 個中隊——**沒有任何地方存著「現在剩幾隊」**。
    `Load_Antaran_Defense_Fleet_` @ 0x4D141(只有 77 bytes,全文可讀):
    ```
    if word_199182 < 1: word_199182 = 1     ; 第 5 筆至少 1
    for bx = 0..4:                           ; ★ 5 種艦體尺寸
        while cx < word_19917A[bx*2]:        ; ★ 每種尺寸的數量
            Load_Combat_Antaran_Ship_(...)
    Load_Antaran_Star_Fortress_()            ; ★ 外加一座星際要塞
    ```
    `word_19917A` 是 BSS(`dw ?`),數量執行期算;寫入端是
    `Build_Antaran_Defensive_Ships_` @ 0x63F9C(由 `Antaran_Invasion_Check_` @ 0x63D92 呼叫),
    它從靜態範本 `byte_181746` 以 `movsd/movsd/movsw`(每筆 10 bytes)複製。
    另有 `sub_646F9` 在艦艇損失時遞減、並從 `word_19918C[]` 補充。
    ⚠ 順帶記一件事:`word_199182` 就是 `word_19917A[4]`(位址差 8 = 4 個 word)。

49. **安塔蘭防禦艦隊與 `Calc_Tech_Value_` 兩張資料表解出來了**(2026-08-08)。

    這一輪派了兩支 sonnet 解資料表,兩份都抽驗過才收。

    `Load_Antaran_Defense_Fleet_` @ 0x4D141 只有 77 bytes:五種艦體尺寸各取一張表的數量,
    **外加一座星際要塞**。上限表 `_n_max_antaran_def_ships`(`byte_181746`)逐位元組解出:
    | 索引 | 尺寸 | 上限 |
    |---|---|---|
    | 0/1 | Small(Raider)/ Medium(Marauder) | **0——永遠不造** |
    | 2 | Large(Intruder) | 3 |
    | 3 | Huge(Interdictor) | 2 |
    | 4 | Titan(Harbinger) | 7 |
    而符號名 `_n_max_antaran_def_ships` 直接證實語意。
    `starting_random_tech.go` 的留白一直寫著「`weight` 一律 1」,把缺口框成「權重不準」。
    這一輪讀 `Choose_Tech_Application_` @ 0xFD335 的迴圈才發現框錯了:
    ```
    mov  ebx, 350h                   ; 0x350 = 848 = 212 × 4 —— 逐 tech-item 的分數陣列
    cmp  byte ptr [eax+117h], 1      ; [player + 0x117 + techIdx] 逐 tech-item 狀態
    cmp  byte ptr [ebx+eax+0C4h], 2  ; [player + 0xC4 + topic]   逐主題狀態
    movsx edx, di                    ; ★ 傳給估值函式的是 tech-item 索引
    ```
    `componentUnlockedFor` 會把底下的抉擇**全部**解鎖。
    已補 `StartingRandomApplicationPick`:挑完主題再挑一項應用,設 `ChosenTech`/`ExplicitChoice`;
    `ResearchAll` 的主題(手冊明說三項全拿)維持全解鎖。
    表③(212 筆 tech-item 記錄)的 topic 欄與第 37 項(研究樹一手驗證)抽的 `OrigTechTopicTable` 逐筆比對,
    - 我說 `movsd/movsd/movsw` 那段是在複製範本表,agent 查出那三行其實是把字串

50. **軍官畫面座標:openorion2 → 執行檔立即數**(2026-08-08)。

    第 14 項(地表道路)把估計熱區換成 openorion2 的真值座標,那是當時能拿到最好的來源。
    現在有更上游的:`Add_Officer_Screen_Fields_` @ 0x9264E 的立即數。

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
    **② 上下捲鈕從來沒接過。** 那兩顆的座標一直都在執行檔裡(`_officer_up_button_seg` /
    `_officer_dn_button_seg`),remake 沒做——所以**軍官清單超過四列就看不到後面的人**。
    **③ 座標抽成套件層級的資料**(`cmd/moo2/officerscreen.go`)。原本內嵌在 `officer()` 裡,

51. **間諜 UI:那張畫面不存在,所以規劃本身要改**(2026-08-08)。

    WORKLIST 的 D 項寫的是「spy / leader UI」,而 spy 那一半的原始假設是**做一張間諜畫面**。

    反組譯搜遍 `Spy_Screen` / `Espionage_Screen` 等關鍵字**零命中**。間諜的任務指派內嵌在
    `Race_Screen_` @ 0x10ACBA 這張**種族關係**畫面裡:每個已接觸種族一列,列上有關係滑桿,
    | 表 | 位址 | 內容 |
    |---|---|---|
    | `_race_spy_btns` | 0x18406F | 每個種族那排任務鈕的錨點 |
    | `_race_bar` | 0x18400D | 關係滑桿軌道的貼圖位置 |
    **沒有查到哪一顆對應哪個任務**——`Adjust_Spy_Mission_Data_` 只看到把任務欄位從 0 設成
    `internal/shell/spy.go` 就標著「手冊只有定性描述,沒給規則」)。


## 六、手冊數值忠實化(2026-08-08)

52. **屏障護盾擋生物武器:擋門是「分類的語意」,不是規則本身**(2026-08-08)。

    `internal/gamedata/planetary_shield.go` 的檔尾掛了一條沒接的規則:

    第 49 項(安塔蘭防禦艦隊)從 `Calc_Tech_Value_` 挖出兩張表:`TechItemCategory[212]`(每項科技屬於
    | category | 成員 | 組數 |
    |---|---|---|
    | **20** | Bio-Terminator、Death Spores | 2 |
    列個生物武器清單」的差別。`TestBiologicalWeaponsAreExactlyCategory20` 把兩者釘在一起。
    `BombardColony` 現依 `sub_DCEBD` 在一般建築、駐軍、建造進度與人口間隨機分配傷亡。
    生物武器接在一般傷亡之後；若到那一步才查 `buildings` 有沒有屏障護盾，
    不是規則。所以 `bioBlocked` 與 `shield` 一樣在吸收迴圈之前就取好。
    `TestBombard_BarrierShieldDestroyedThisTurnStillBlocks` 守這條,而且它是 PASS 不是
    - **一次轟炸投幾莢沒有真值。** 手冊說「每一個發射出去的孢子莢」,但沒說一次投幾莢,
      (一艘一莢)是 **remake 的建模選擇**,寫在 `bioweapon.go` 檔頭與呼叫端註解裡。
    - **只有屏障護盾擋。** 輻射護盾與通量護盾的手冊敘述都沒有那一句。
      `TestOnlyTheBarrierShieldBlocksBiologicalWeapons` 把這件事釘住,免得日後
    - **G 項的第二條仍然刻意不做**:護盾被炸掉後氣候不變回 Radiated。
    截圖廊零位元組差異——因為 `BioWeaponKills` 是單次轟炸結果、不是存檔欄位,

53. **聊天列:資料模型與畫面都做完了,但字從來沒離開過這台機器**(2026-08-08)。

    `internal/netplay/chat.go` 從 `Receive_Chat_Msg_` @ 0xDD351 逐位元組解出了整個資料
    結構(保留 14 則、每則 82 byte、內文上限 80、發話者 ≥8 是 GNN),
    `cmd/moo2/netnextturn.go` 也把輸入行、游標、兩種前綴、版面全畫出來了。

    但 `sendChat` 的註解自己寫著:
    大廳(`lobby.go`)把大家連起來、發名冊,然後就結束了;回合表(`lockstep.go`)是
    **訊息幫浦**:`internal/netplay/session.go`。
    ebiten 的 `Update` 是單一 goroutine 每幀跑一次。讀 socket 一定要另一條 goroutine
    (否則畫面卡在 `Read` 上),但**規則狀態不能被那條 goroutine 直接改**——那是資料競爭,
    中間隔一個佇列,`Update` 每幀 `Poll()` 清空一次,**所有狀態變更都在 Update 這一條線上**。
    原版 `Send_Chat_Msg_` @ 0xDD3B8 是逐一走過玩家陣列**直接發給每一個人**——IPX 廣播式
    | 情境 | 正確做法 | 寫錯會怎樣 |
    |---|---|---|
    | 主機收到(第一手)| 以**連線**為準,蓋掉封包自報的 Player | 客戶端可以冒名說話 |
    | 客戶端收到(可能是轉發)| **信任**主機標的 Player | 所有人的話都顯示成主機說的 |
    - **幫浦建好之後再加入的人不會被納進來。** 延遲建立是為了不漏掉大廳階段後來的人,
    - **主機可以造假。** 客戶端信任主機標的發話者編號——星狀拓樸下主機本來就轉發所有流量,

54. **三個「沒有查到寫入端」的欄位,寫入端一直都在**(2026-08-08)。

    `docs/re/calc-tech-value.md` 第 6 節把四件事列為擋門,其中三件的理由都是
    「**沒有查到寫入端**」:

    | 欄位 | 當初的判定 |
    |---|---|
    | `[player+0x89F]` | 「未定名,種族特性相關(**猜**)」 |
    | `[player+0x28]` | 「0..6 的意義不確定……**不要猜是哪 7 種性格**」 |
    | `[player+0x205]` / `[player+0x206]` | 「**沒有查到寫入端**,不排除是別的東西」 |
    ```
    grep -nE "mov +(byte ptr )?\[e..\+89Fh\]," Orion2.exe.asm
    ```
    ### `[player+0x89F]` = 政體
    `sub_E4204`(「取得某項科技」的效果套用函式,第一行就是
    | techIdx | 科技 | 寫入值 | asm 行 |
    |---|---|---|---|
    | 42 | Confederation | **1** | 327016 |
    | 92 | Imperium | **3** | 327006 |
    | 65 | Federation | **5** | 327021 |
    | 77 | Galactic Unification | **7** | 327011 |
    奇數是它們的科技升級版,`值/2` 就是同一族。`Calc_Tech_Value_` 階段 F 那句
    `[+0x89F]/2 == 2` 因此讀得出來了:**民主/聯邦那一族**。
    `gamedata.AssimilationGovernment` 的順序當初是照手冊表格排的,**是 remake 自己的選擇**。
    代表手冊那張表的排列本來就反映內部編號。已用 `government_orig_test.go` 釘住,
    順便釘住 `MoraleGovernmentType` 與它同號(原版只有一個 `[player+0x89F]`,
    ### `[+0x28]` / `[+0x205]` / `[+0x206]`:同一支函式的三次加權抽選
    三個欄位都在 `Init_NPC_Personalities_Objectives_Themes_` @ 0x589D6 裡被寫,
    3. `sub_1247A0(總和)` = 原版的 `Random_`,回 1..總和;
    5. 寫進 `[+0x28]` → `[+0x205]` → `[+0x206]`。
    三次抽選的權重都先被 `byte_199CB0`(難度)加減過(`cmp byte_199CB0, 3` / `4` 各一段)
    ——難度越高,某些候選越容易被抽中。這同時再次佐證 `byte_199CB0` 是難度。
    ```
    cmp  byte ptr [eax+28h], 64h   ; 已經是 100(人類)?
    setnz dl
    jz   loc_58E39
    mov  [eax+28h], cl             ; 不是 → 寫抽選結果
    ...
    loc_58E39: mov byte ptr [eax+28h], 64h   ; 是 → 把 100 寫回去
    ```
    `internal/netplay/chat.go` 寫著「GNN 走的是 case 43(`Send_GNN_Chat_Msg_` @ 0xDD42A)」,
    | | 號碼 | 出處 |
    |---|---|---|
    | `Send_GNN_Chat_Msg_` 送出的型別 | **0x2D(45)** | asm 315237 `mov edx, 2Dh` |
    | 收訊跳表裡「speaker 固定 8」那一格 | **0x2B(43)** | `mov eax, 8; jmp` → `Receive_Chat_Msg_` |
    一般聊天則兩端對得起來(送 0x27 = 39,收 case 39 呼叫 `Receive_Chat_Msg_`)。
    沒有核 `sub_F6816` 是不是唯一的傳送出口、也沒有核有沒有第二個派送器。
    常數已拆成 `ChatGNNSendOpcode` / `ChatGNNRecvCase` 兩個,不再假裝它們是同一個數。

55. **AI 對手的科技,先前完全是偷來的**(2026-08-08)。

    這一項不是從逆向來的,是從**一個探針**來的。第 54 項(三個寫入端)訂正 `internal/ai/research.go`
    的過期斷言時順手看了一眼呼叫端,發現 `ai.DecideResearchTopic` **零個呼叫端**。

    | 回合 | AI 0 主題 | AI 0 已完成 | 玩家主題 |
    |---|---|---|---|
    | 1 | 1 | 7 | 3 |
    | 20 | **1** | 7 | 3 |
    | 50 | **1** | 9 | 31 |
    | 100 | **1** | 11 | 9 |
    | 200 | **1** | 13 | 15 |
    **① 沒有人替 AI 選下一個主題。** `engine.RunEmpireTurn` 對 AI 一樣跑研究階段,主題完成
    也會標進 `CompletedTopics`——但 `ResearchTopic` 是誰設的?玩家那邊是研究畫面設的,
    那上表的「已完成」為什麼還是從 7 長到 13?**它偷的。** `spy.go` 的間諜每隔一段時間從
    **② 多選主題的抉擇永遠掛著。** `engine.recordCompletion` 對多選主題預設記第一項並掛起
    `HasPendingChoice`,等玩家去選。AI 從來沒有人替它選,所以它的 `ExplicitChoice` 永遠是空的
    ——而 `groundEquipTechOwned` 的判定是「主題完成 + **沒有明確抉擇** → 視為擁有」。
    - `retaliationAttackers` → `bestUnlockedWeaponValue`:軌道防禦永遠用開局武器
    - `aiFleetSpeedParsecs`:AI 艦隊永遠是核融引擎的速度
    - `groundEquipTechOwned`:AI 地面部隊永遠沒有動力裝甲
    主題選擇沿用 `ai.DecideResearchTopic`(吃性格的設計啟發式)。原版的 `Calc_Tech_Value_`
    全部乘在語意未解的 `word_1AB1xx` 上——**仍然不可照抄**。
    但**應用項的抉擇**用得上一手值:`gamedata.TechCategoryWeight` 就是階段 B 給 `ecx` 的
    因此是決定性的(`determinism_test.go` 那組閘門的前提)。
    (`spy.go` 直接寫 `CompletedTopics`,不經過研究階段),那時候 `ResearchDone` 是 false
    但目前主題已經完成了。只看 `ResearchDone` 會讓 AI 卡在一個偷來的主題上繼續投點——
    `TestAIRepicksAfterStealingItsCurrentTopic` 守這條。
    | 回合 | AI 0 主題 | 已完成 | 明確抉擇 |
    |---|---|---|---|
    | 50 | 1 | 9 | 2 |
    | 200 | **66** | **17** | **9** |

56. **領袖永久免費——用覆蓋率找出來的**(2026-08-08)。

    第 55 項(AI 科技先前靠偷)是靠「這支函式有沒有呼叫端」找到的。那個問法有個明顯的盲點:
    `ai.DecideResearchTopic` **有**呼叫端(`RemakeDecider.ResearchTopic`),
    只是那個呼叫端自己也沒人叫。所以先做了兩輪掃描:

    | 問法 | 結果 |
    |---|---|
    | 生產碼裡零次提及的函式 | 2 個(`bitBytes`、`homeworldBuildingsLegacy`)|
    | 從 `cmd/` + 頂層宣告出發、傳遞閉包到不了的 | 一樣是那 2 個 |
    ```
    go test -run TestZZLongGameForCoverage \
      -coverpkg=./internal/shell,./internal/engine,./internal/gamedata -coverprofile=...
    go tool cover -func=... | awk '$3=="0.0%"'
    ```
    **`engine.CheckVictory` 零覆蓋——但這不是洞。** 它是把三條勝利條件包在一起的便利函式,
    而 shell 有自己的三條路徑(`advanceConquestVictory` 走 `CheckExtermination`、
    `advanceCouncilElection`、`advanceAntaranVictory`),三支都每回合跑。
    優先序一致性已有測試釘著(`antaran_victory_test.go`)。**留著它是個小陷阱**
    **歷史發現：`gamedata.LeaderMaintenanceCost` 曾是零覆蓋缺口。**
    **2026-08-25 勘誤：已由 IDA Pro 閉合 `sub_94A9D @ 0x94A9D` 並接進單次帝國經濟結算；
    下列文字保留用來說明缺口形成原因，不代表目前狀態。**
    註解連 `LEADER_ID_LOKNAR` 那個不移植的硬編特例都交代了——**零個生產端呼叫**。
    `grantMaroonedLeader` 給的領袖,先前兩樣都免。
    每位 `ceil(hireCost/100)`(下限 1),有 Megawealth 技能者免費。`hireCost` 走與
    `MercHireCost` **同一條公式**——兩邊用不同基準的話,同一位領袖會出現
    - **付不出來只是扣到 0,領袖不會離職。** 原版錢不夠會怎樣沒查到規則(手冊只說要付,
    - ~~**AI 不付。** `AIOpponent` 沒有 Leaders 欄位——不是漏掉,是那一整層還不存在。~~
      ⚠ **2026-08-08 追認:`AIOpponent` 現在有 `Leaders` 欄位了**(版本 diff #5「守方指揮官 2.5x」
    - 判定 Megawealth 走 `leaderSkills` 那條既有路徑,不比對技能字串:標籤會被翻譯,
      拿它當識別鍵在英文模式下查不到(`Leader.Skills` 檔頭記過這個坑)。
    國庫已經是負的時候不能再扣。`session.go` 對戰損 `bcLoss` 有一段註解記著同一個坑:

57. **征服來的人口先前是「免費全額生產」**(2026-08-08)。

    第 56 項(領袖永久免費)用覆蓋率找洞時只看了「應該自動發生」那一類。同一份清單裡還有 196 支
    `internal/gamedata` 的函式零覆蓋——那些是**純規則**,零覆蓋的意思是
    「這條規則在一局裡從來沒有適用過」。

    | 函式 | 呼叫端 |
    |---|---|
    | `ProdAlienWorkerOutput` | **0 個** |
    | `ProdWorkerOutput` | **0 個** |
    remake 有 `UnassimilatedPop`(第 46 項(叛亂)的叛亂系統與多種族士氣懲罰都在用),
    `ProdWorkerOutput` 是「每工人至少 1 產能」的下限。它零覆蓋不是因為沒接
    ⚠ 下限只套**工業**:`ProdWorkerMinimum` 的手冊依據講的是「每個**工人**單位」,
    ——`ColonyState` 只有 `UnassimilatedPop` 與三個職業人數,沒有交叉表。
    `TestConqueredMarkerDoesNotLeakIntoEconomy` 斷言「征服標記不該改變經濟結算」,
    而它測的是整支 `markColonyConquered`——**那支同時設 `UnassimilatedPop`**。
    處理方式不是把它刪掉,是**把兩件事拆開**:記帳用的 `ConqueredFrom` /
    `ConqueredFromKnown` 仍然不該有經濟效果(原意保留);`UnassimilatedPop` 的經濟效果

58. **一個擋門理由過期了三個月都沒人回頭看**(2026-08-08)。

    覆蓋率清單裡有 36 支 `gamedata` 函式**零覆蓋而且零呼叫端**。其中三支是同一個系統的:

    ```
    SpyGovernmentDefenseBonus
    SpyRaceTraitBonus
    SpyTechnologyBonus
    ```
    ### `spy.go` 自己寫了為什麼
    | 擋門理由 | 現在還成立嗎 |
    |---|---|
    | 無逐科技模型 | ❌ **不成立**。`groundEquipTechOwned` 已是三個系統共用的判定(生物武器、地面裝備、進階政體),那 5 項科技在 `enums.go` 都有常數 |
    | 無種族間諜特性強度 | ✅ 仍成立。`TRAIT_SPYING` 只標記「有沒有」,沒有 −3/+3/+6 的分級,`Races` 表也沒這一欄 |
    | AIOpponent 無政府型態 | ✅ 仍成立(對 AI)。但**玩家有** `s.Government`,而手冊那一欄本來就只給 Defense |
    直接用 `s.Government` 會讓研究出帝國的獨裁玩家永遠拿獨裁那一格,而手冊對基本型與
    這一步靠的是第 54 項(三個寫入端)的結果:`SpyGovernmentType`、`AssimilationGovernment`、
    `MoraleGovernmentType` 三個列舉編號相同,不是巧合——**原版只有一個 `[player+0x89F]`**,
    反面暗示其餘取最佳(見 `gamedata/leader_skill_apply.go`)。這裡沒有那句話,所以
    這是推論不是引用,標在 `spyTechBonusFor` 的註解裡。
    只是那個理由已經不成立**。`rulebook/63`(真相在程式碼裡,不在過期標記裡)講的正是

59. **「成就」科技的全帝國效果,四條一條都沒接**(2026-08-08)。

    第 58 項(擋門理由過期三個月)那個形狀又出現一次:**規則寫好了、擋門理由也寫好了,而理由已經不成立**。

    | 手冊規則 | 出處 | 呼叫端 |
    |---|---|---|
    | 虛擬實境網路:全帝國士氣 +20% | p.97-98 | **0** |
    | 心靈學:特定政體下士氣 +10% | p.100-101 | **0** |
    | 微晶構築:每個工業工人 +1 產能 | 手冊 | **0** |
    | 奈米分解者:行星污染容忍值 ×2 | 手冊 | **0** |
    `colonyMoralePercent` 的檔頭寫了為什麼:
    是另一種東西。而「有沒有研究出來」remake 一直查得到——`groundEquipTechOwned` 已經是
    建築蓋好就永遠在;科技**會被偷、會被交換**,而 `groundEquipTechOwned` 的判定還吃
    「有沒有做過明確抉擇」——那也會變。所以 `syncAchievementColonyFields` 每回合重算,
    原本想一次給玩家兩項來測,結果後設的那項把前一項的 `ChosenTech` 蓋掉了。
    那正是狀態指紋的位置。`ColonyState` 多了兩個欄位,存檔 JSON 就變了,指紋跟著變。

60. **註解說「打得準也閃得掉」,程式只做了前半**(2026-08-08)。

    覆蓋率清單裡剩下的兩支艦員加成:

    ```
    ShipCrewBoardingBonus         0 個呼叫端
    ShipCrewMissileEvasionBonus   0 個呼叫端
    ```
    追下去才發現同一段還藏著第三件事——**`ShipCrewDefenseBonus` 也沒接**,
    而且那一項比另外兩支更難發現:它**有**呼叫端,在 `engine.BeamDefense` 裡。
    只是 `engine.BeamDefense` 在一局裡從來沒被執行過(shell 的戰鬥自己算,沒走它)。
    `mkPlayerCombatantsIndexed` 寫著:
    ```go
    // 艦員經驗(手冊 p.121 的 BA/BD 兩欄):老手打得準也閃得掉。
    crew := s.shipCrewLevel(sh)
    atk += gamedata.ShipCrewOffenseBonus(crew)
    ```
    註解說**兩欄**,程式只加了 BA 那一欄;`def` 從頭到尾只有艦體值。
    `ResolveMissileShot` 的 `defenderEvasionBonus` 一直恆傳 0,理由是:
    | 來源 | 現況 |
    |---|---|
    | ECM 干擾器 / 慣性穩定器 | ✅ 仍缺(SpecialOptions 沒有這些元件)|
    | 種族 Ship Defense 特性 | ✅ 仍缺(手冊連檔位名稱都沒列)|
    | **艦員經驗** | ❌ **算得出來**——`shipCrewLevel` 每回合都在更新 |
    | **舵手(Helmsman)軍官** | ❌ **算得出來**——`SKILL_HELMSMAN` 在第 45 項(領袖技能)就進來了 |
    而且**只算艦艇軍官**——舵手是開船的,殖民地領袖不會坐在艦橋上,這與 `starlane.go`
    `ShipCrewBoardingBonus` 零呼叫端,但 `crew.go` 自己寫了為什麼:

61. **兩個「為了給玩家看」而寫的函式,畫面從來沒用過**(2026-08-08)。

    覆蓋率清單剩下的 36 支零呼叫端函式,這一輪逐支看完並分類。**大多數是合理的**:

    | 類別 | 例子 | 為什麼不接是對的 |
    |---|---|---|
    | 子系統還不存在 | `ShipCrewBoardingBonus`、`ResolveGroundBattle` | 登艦戰沒建模;地面戰走的是另一條原版路徑 |
    | 手冊資料不全 | `DamageEngineExplosionPotential` | 手冊沒給每格衰減率,接了就得自己編一個 |
    | 元件沒建模 | `ComputerHP`、`DamageShieldCapacity` | 艦艇子系統 HP 不在模型裡 |
    | 純存取器 | `PlanetSpecialWeight`、`WeaponModSpaceCostPercent` | 表本身有人用,只是沒人用這支包裝 |
    | 刻意的偏離 | `PlanetaryShieldEffectiveClimate` | 建築效果不可逆(第 42 項(關掉兩條留白)記錄過)|
    > `AssimilationProgressNeeded`:抽出來是因為 **UI 要顯示**「這個殖民地還要幾回合才完全
    > `CrewXPToNextLevel`:這個函式存在是為了**讓 UI 能顯示進度**——一個只會默默上升的
    - **殖民地畫面**:未同化人口 > 0 時多一行「未同化 N ／還需 M 回合」。
    - **星圖資訊面板**:接在既有的「艦隊陸戰隊 N」那一行後面,顯示艦隊的艦員等級與
    - **沒有逐艦資訊面板。** 艦員等級目前只有「整支艦隊的最低值」這個摘要。
    - **星圖那一行沒有視覺回歸覆蓋。** 截圖廊零位元組差異,而那代表那個面板在截圖的

62. **護盾減傷五級裡有四級是錯的,而正確的常數就躺在旁邊**(2026-08-08)。

    第 61 項(兩個函式畫面沒用過)把「零呼叫端的**函式**」清光了。但那個掃法有個洞——第 59 項(成就科技效果)的
    `MoraleVirtualRealityNetworkBonus` 是個 **const**,不是函式,當初是碰巧發現的。

    ```
    DamageShieldReductionClassI   = 1
    DamageShieldReductionClassIII = 3
    DamageShieldReductionClassV   = 5
    DamageShieldReductionClassVII = 7
    DamageShieldReductionClassX   = 10
    ```
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
    - **索引**會被清單順序決定——那正是這次出問題的地方,以後插一級進去就全錯位。
    - **名稱**會被翻譯(英文模式下 `"第一級護盾"` 查不到)。

63. **一支偵察艦隊與一支末日之星艦隊,登陸能力完全相同**(2026-08-08)。

    第 62 項(護盾減傷)修的是「shell 自己編了一個數字,而手冊值就在 gamedata 裡沒人用」。
    順著同一族往下查(`armorHPByName` / `shipStrength` / `MarineTransportCapacity`),
    第三支中了同一個形狀,而且**更明顯**:

    ```go
    func (s *GameSession) MarineTransportCapacity() int {
        return len(s.Fleet().Ships) * gamedata.GroundTransportShipMarineCapacity  // 艦數 × 4
    }
    ```
    | 艦體 | 護衛艦 | 驅逐艦 | 巡洋艦 | 戰艦 | 泰坦 | 末日之星 |
    |---|---|---|---|---|---|---|
    | Marines | 5 | 8 | 12 | 20 | 30 | 50 |
    `shipClassFromName` 那支函式本身就是為了查**同一張表**的另一欄(Space)寫的。
    那張表的「Comp.」與「Drive」兩欄**逐格對上** openorion2 的 `computerHPTable` /
    `driveHPTable`(`docs/tech/ship-design-space.md` 早就做過這個交叉驗證)。
    - **`armorHPByName`**(裝甲 HP 依元件):手冊 p.121 有 Armor/Struct. 欄(4/10/30/50/80/150),
    - **`shipStrength`**(戰力點 1/2/4/8/16/32/64):它的註解本來就寫明是「供最小戰鬥解算」

64. **武器傷害全是估計值,而手冊原表一直在可抽文字的 PDF 裡**(2026-08-08)。

    第 62/63 項的形狀是「shell 編了一個數字,手冊值就在旁邊沒人用」。這一項更遠一步:
    **手冊值連抄都沒抄過**,而理由是一句錯的話。

    **不需要 OCR,也不需要另找手冊。** `moo2_patch1.5/GAME_MANUAL.pdf` 是可直接抽文字的
    (`WeaponSpaceByName`)——當時抽了 Size,沒順手抽旁邊的 Damage 欄。
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
    有一條測試專門釘住這個反轉:`TestWeaponLineIsNotMonotonic`。有人「順手」把表改回
    - **Cost 欄不動。** 手冊的 Cost(雷射 5、死光 75)與 remake 的生產成本(20/300)
    - **`atk = 艦體 + 武器` 的合成不動。** 大艦在原版是靠**裝更多把武器**變強,
    - **手冊還列了 remake 沒有的武器**(離子脈衝砲、引力波束、干擾者、重錘裝置、粒子束),
    電漿砲 1.31 為 6-30、1.50 為 4-20(`MANUAL_150.html` 明載),仍由
    `RuleProfile.PlasmaCannonMaxDamage` 覆寫。新加一條測試驗「**只有**電漿砲隨版本變」


    ### **手冊有 18 把武器,remake 只做了 10 把**(2026-08-08)。


        第 64 項(武器傷害真表)把武器傷害換成手冊真值時,順手記了一句「手冊還列了 remake 沒有的武器
        (離子脈衝砲、引力波束、干擾者、重錘裝置、粒子束),擴充元件表不在這一輪」。
        這一輪就是那一輪。

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
        每一項的研究主題都取自 **`gamedata.OrigTechTopic`**——第 49 項(安塔蘭防禦艦隊)從執行檔挖出來、
        `weaponKindByName` 決定走飛彈解算還是光束解算。新增的三項判定不是靠名字裡有沒有
        兩個獨立來源同意。測試把這個一致性釘住:category 21 ⟺ `WeaponKindMissile`。
        `WeaponSpaceByName` 查不到會回 0,而 0 的意思是「不佔空間」,設計驗證會**靜默放行**
        一艘塞了十把重錘裝置的護衛艦。`TestEveryWeaponHasManualDamageSpaceAndTopic`
        - **Cost 是 remake 的尺度。** 手冊的 Cost 欄與 remake 的生產成本不是同一個單位
        - **重錘裝置的「always hits」沒接。** 手冊 specials 欄寫了,但 remake 的命中判定


    ### **反飛彈火箭:攔截公式抄完了,但沒有一艘船裝得上**(2026-08-08)。


        `ResolveMissileShot` 的第一個參數是 `hasAMR`。兩個呼叫端都寫死 `false`,理由是:

        `gamedata/missile.go` 有完整的 AMR 模型:最大射程 15 格、距離換算成 Range 索引、
        逐索引的命中率表。`ResolveMissileShot` 也已經接好分支(`hasAMR` 為真且在射程內 →
        | 欄位 | 來源 |
        |---|---|
        | 研究主題 | **執行檔**(`OrigTechTopic` → 進階工程),不是照名字猜 |
        | 分類佐證 | **執行檔** category 28(反飛彈/干擾),與手冊 p.127 的分類一致 |
        | Value = 0 | AMR 不加攻防——它的效果是攔截,加一個攻擊值等於偷偷讓它變成武器 |
        | Cost = 70 | **remake 的尺度**(見第 64/64 項),不是手冊值 |
        - **正**:同一組種子跑 300 次,裝了 AMR 的一方被飛彈打中的次數應下降。
        - **反**:AMR **不該**擋光束——手冊寫的是「destroys incoming missile」。
        ——現在的 `weaponKindByName` 預設回光束,直接加會讓一艘掛核彈的船在艦隊戰裡當光束艦用。


    ### **炸彈:一個「打不到船」的武器種類**(2026-08-08)。


        第 64 項(武器傷害真表)把炸彈擋在門外,理由寫得很明確:

        ⚠ `weaponKindByName` 預設回光束,直接加會讓一艘掛核彈的船在艦隊戰裡當光束艦用。
        `battleVolley` 那一支對炸彈直接 `continue`,**連骰子都不擲**。
        決定性測試會無故變動而且看不出原因。`TestBombDoesNotConsumeRandomness` 用


    ### **球形武器:一整條戰鬥解算分支,零武器掛載**(2026-08-08)。


        `weapon_kind.go` 的檔頭寫著:

            `ResolveSphericalShot` 的函式註解也重複了一次同樣的話。**整條解算路徑是死碼**
        ——連 `battleVolley` 的 `case WeaponKindSpherical:` 那一段都從來沒有執行過。
        - **電漿通量**是海鰻怪獸專屬,不是可造艦元件;
        - **引擎爆炸**是船被打爆時的事件,不是武器;
        「4-32 structural hits」。前者要知道**被打的那艘船有多大**——而 `combatant` 先前
        沒有艦體等級這個欄位(它只有 `shipStrength` 換算出來的戰力點)。補上 `sizeClass`,
        直接複用既有的 `shipClassFromName`。
        ——`ResolveSphericalShot` 早就把它做成參數(`bypassShieldAndArmor`),先前沒有呼叫端
        加完元件跑測試,`TestEveryWeaponHasManualDamageSpaceAndTopic` 立刻紅:


    ## 七、艦載元件忠實化(2026-08-08)


    ### **p.127 剩下的特殊武器:全部卡在機制,不是卡在資料**(2026-08-08)。


        第 64–64 項把手冊的武器表逐頁補進來:光束(p.124)、飛彈(p.125)、
        炸彈(p.126)、球形(p.126 清單 + p.127 數值)。剩下 p.127 那張特殊武器表。

        | 武器 | 手冊效果 | remake 缺什麼 |
        |---|---|---|
        | 牽引光束 | reduces target speed | **戰鬥速度模型**。`gamedata.CombatSpeed` 移植好了但零呼叫端,戰術戰鬥沒有逐艦速度 |
        | 停滯力場 | places target in suspended animation | **「這一輪不能動」的狀態**,同上 |
        | 黑洞產生器 | immobilizes and destroys target | 同上,外加「即毀」不是傷害 |
        | 電漿網 | 5-25 **to each side** of target | **護盾分面(facing)模型**。`DamageShieldCapacity` 也是為了它寫的,同樣零呼叫端 |
        | 恆星轉換器 | 400 **to each side** | 同上。⚠ **行星版**已經接了(`StellarConverterRetaliationAttack`,轟炸反擊),艦載版沒有 |
        | 陀螺去穩器 | 1-4 per size class | 資料齊、艦體級數也有了(第 64 項(武器傷害真表))——但它**不在手冊的球形清單裡**,而 remake 的光束路徑沒有「per size class」這個乘數 |
        `CLAUDE.md` 指名「唯一活來源」是本文件開頭那張「真正還缺的」表,而它 2026-08-07 起

65. **種族只有 7 個自編數字,原版有 31 格特性——一手表挖出來全部換掉**(2026-08-08)。

    第 64 項(武器傷害真表)在活來源表上補了「手冊數值忠實度」那一列,並列出四種還有殘量的問法。
    這一項走的是 **③ 被餵固定值的參數**。掃描結果指向同一個根:

    全部是**種族特性**。而 `engine.PlayerState` 只有一個 `TolerantRace bool`
    ——`shell.Race` 把整個種族壓成七個自編數字(工業/科研/農業/成長/起始BC/每人BC/戰鬥%)。
    `internal/save/entities.go` 早就有 `Traits [31]int8`,讀的是玩家結構 `+0x89F`。
    遊戲層從來沒用過它。反組譯 `sub_12983`(開局玩家設定)給了完整鏈路:
        add     eax, 89Fh                  ; → player+0x89F
    | 來源 | 內容 |
    |---|---|
    | `RACESTUF.LBX` asset 7 | 4 位元組表頭(13 筆 / 每筆 31 格)+ 13×31 選項等級 |
    | 執行檔 `byte_17D1F9`(檔案位移 0x1FB88D) | 10 列 × 3 級的換算表 |
    | `SAVE10.GAM` | 五族(Alkari/Klackon/Mrrshan/Sakkra/Trilarian)展開後的 `Traits[31]` |
    ⚠ 中途踩到一次:`conv` 只切了 30 位元組,而特性 9 的 3 檔要讀到第 30 格,
    | 特性 | 1 檔 | 2 檔 | 3 檔 |
    |---|---|---|---|
    | 人口成長 | −50 | +50 | +100 |
    | 農業 | −1 | +2 | +4 |
    | 工業 / 科研 / 金錢 | −1 | +1 | +2 |
    | 艦艇防禦 | −20 | +25 | +50 |
    | 艦艇攻擊 | −20 | +20 | +50 |
    | 地面戰 / 間諜 | −10 | +10 | +20 |
    已拆成 `CombatPct`(艦攻)與 `ShipDefPct`,並在戰列上對稱套用。
    `GroundRaceCombatBonus(GroundRaceGnolam)` 回 −10,而下一行的 `GroundApplyLowGPenalty`
    `TRAIT_GROUND_COMBAT` 本來就是 **0**——那 −10 完全來自 `TRAIT_LOW_G`。
    改由特性表驅動之後重複自然消失,`GroundRace` 列舉與 `GroundRaceCombatBonus` 一併刪除
    `playerMarineForce` 的檔頭寫著:
    「remake 的 `Races` 表採**概略調校值**…非手冊精確數字」。**答案一直在自己的 docs 裡,
      - **自訂種族拿不到布林特性。** `RaceOrigIdx` 記 −1,`OrigRaceHasTrait` 越界回 false。
      - **AI 對手不吃種族特性。** `AIOpponent` 沒有種族欄位——那一整層還不存在,不是漏接。
      - **`TradeGoodsIncome` 的 fantasticTrader 仍是 false。** 諾蘭姆有神級商人特性,
      - **特性 31(貧瘠母星)在列舉裡有,陣列只有 0..30。** 原版放不下它,不替它捏一格。


    ### **布林特性接線:五條寫好的規則終於有呼叫端**(2026-08-08)。


        第 65 項(種族特性31格)挖出了 13 族 × 31 格的特性表,把**數值型**那 9 格接了進去;
        布林那 21 格當時只做到「查得到」。這一項接「查得到之後怎麼用」。

        | 特性 | 種族 | 先前的狀態 |
        |---|---|---|
        | 統帥 | 姆瑞森 | `RaceWarlord` 欄位在、5 個呼叫端都讀它,**沒有人寫入** |
        | 惹人厭 | 矽基 | `AssimilationTurns` 的 repulsive 分支同上 |
        | 寬容 | 矽基 | `engine.ColonyState.TolerantRace` **從來沒有寫入端** |
        | 神級商人 | 諾蘭姆 | `TradeGoodsIncome` / `IncomeFoodSurplusRevenue` 硬傳 `false` |
        | 魅力非凡 | 人類 | `raceDiploBonusPct` 硬比 `RaceIndex == 0` |
        前四項的擋門理由是同一句話,`session.go` 寫得很清楚:
        `raceDiploBonusPct` 原本是 `if s.RaceIndex == 0 { // 人類 }`。這在兩個方向上都會錯:
        `Races` 重排一次就指到別族;而自訂種族選了魅力非凡也永遠拿不到。改成查特性。
        同時發現 `ApplyCustomRaceBonuses` **沒有把 RaceIndex 標成 −1**,於是自訂種族會停在
        第一版把種族編號與五個布林旗標當成 `GameSession` 欄位存進存檔。三個測試同時紅:
        根因不是漏抄,是**設計錯了**:那些值全部可以從 `RaceIndex` 算出來,而 `RaceIndex`
        `engine.PlayerState.FantasticTrader` 與 `engine.ColonyState.TolerantRace`,
        在 `EndTurn` 開頭與成就同步(第 59 項(成就科技效果))並列,冪等。
          - **魅力非凡在同化那一側仍不生效。** 現在查得到人類有它了,但手冊只說
          - **自訂種族的 pick 尚未寫進特性。**（歷史狀態；已由第 95 項修正，現在以
            `CustomRaceTraits` 保存客製選項。）
          - **其餘布林特性(幸運/全知/匿蹤艦/跨維度/母星品質)未接。** 水棲、食岩與創造力／缺乏創造力已在後續項目接入可證實的規則。
            這句仍成立的部分是深層消費端未接，不是選項沒有保存。

66. **高能聚焦:規則寫好了,但那個東西**裝不上**(2026-08-08)。

    ③ 那條問法(被餵固定值的參數)清掉種族那一叢之後,剩下的最大一項是:

    改造走 `Ship.Mods`,系統走 `Ship.Special` 與 `SpecialOptions` 那張表,而
    `SpecialOptions` 裡沒有它。於是「玩家裝不上 → 呼叫端沒有東西可傳 → 恆傳 0」。
    三句各自對應程式裡的一個地方:傷害走 `DamageMountAdjustedValue` 的 hefBonus、
    命中走 `CombatHitThreshold`、距離衰減走 `DamageDissipationPenalty`。
    **只接第一個,另外兩個一個都不能碰。** `hef_test.go` 逐句釘住,其中命中那一條是
    快速結算(`combatant.hasHEF`)與格子戰術(`CombatShip.HEF`)是兩條獨立的路。
    `ResolveMissileShot` 的檔頭寫著:
      - **建造成本 90 是 remake 值。** 手冊行文只給效果不給成本,執行檔的元件表還沒挖到
        硬化護盾(100)與反飛彈火箭(70),與 `SpecialOptions` 其餘元件同一種標記方式。
      - **敵方艦不裝。** `genEnemyFleet` 沒有個別元件設計資料,一律 false

67. **裝甲科技倍率:一則撤回,外加重裝甲與穿甲抵銷**(2026-08-08)。

    ### 先講撤回

    這一輪稍早(第 64 項(武器傷害真表)一帶)拒絕動 `armorHPByName`,理由寫得很篤定:
    | 裝甲 | 手冊原文 | 倍率 |
    |---|---|---|
    | 鈦 | 「standard armor for FTL ships」(基準,未給倍率) | 100% |
    | 三鈦 | increases the structural integrity ... by **100%** | 200% |
    | 佐特 | increases the structural integrity ... by **300%** | 400% |
    | 中子素 | boost the structural hits ... by **500%** | 600% |
    | 精金 | increases the structural hits ... by **700%** | 800% |
    | 氙素 | Ships with this armor have **10 times** the base structure and armor points | 1000% |
    ——當初根本沒有證據可找。留在 `gamedata/armor_tech.go` 檔頭,不默默改掉。
    `ArmorOptions` 裡它掛在 `TOPIC_ARTIFICIAL_LIFE`、`UnlockTech = 0`,註解標「里程碑,proxy」。
    執行檔的 tech→topic 表說它屬 **Xenon Technology**(74),而 `TECH_XENTRONIUM_ARMOR`
    氙素裝甲在 `ArmorOptions` 裡,查得到;**重裝甲不在任何清單上**——因為它是
    「Heavy Armor **(System)**」,系統要進 `SpecialOptions`,而它不在。與第 66 項(高能聚焦)的
      - **抵銷穿甲**,與氙素裝甲取聯集(`shipNegatesArmorPiercing` 把兩條路併起來)
      - **建造成本 110 是 remake 值**,理由同高能聚焦(執行檔的元件表還沒挖到)。
      - **「三倍」乘的是裝甲那一池不是結構。** 手冊明說 "the ship's **armor** can sustain
        before damage gets through to the structure",所以只動 `armorHPByName` 的結果。
      - **兩池抽象本身沒有改。** 把裝甲折回結構點數是另一個層級的重構,不在這一項裡。

68. **手冊元件完整性盤點 + 飛彈防禦系統家族**(2026-08-08)。

    第 131(高能聚焦)與第 132(重裝甲)是**同一個形狀撞到兩次**:手冊裡標
    `(System)` 的東西沒進 `SpecialOptions`,於是那條規則的參數只能恆傳 0。
    撞第二次之後改問法——**不要一個一個等它自己冒出來**,把手冊所有元件條目
    對 remake 的四張元件表做一次完整性盤點。

    抽出 System + Ship 共 88 個名稱,用 `gamedata.TechnologyName` 反查成 `Technology`,
    任一張表的 `UnlockTech`。**全自動,不靠記憶。**
    (`Phasors` vs `Phasor`、`Class V Shields` vs `Class V Shield` 這類單複數差異)。
    | 桶 | 數量 | 說明 |
    |---|---|---|
    | **不需要元件槽** | 15 | 艦體等級(泰坦/末日之星)、殖民船/前哨船/運輸艦/貨船(建造項目)、各級引擎(星圖速度模型)、燃料槽(航程模型)、各級電腦(戰鬥電腦 proxy)——remake 用別的模型承接,不是缺口 |
    | **這一項接上** | 8 | 見下 |
    | **卡機制** | 4 | 陀螺去穩器 / 電漿網 / 停滯力場 / 牽引光束——第 64 項(武器傷害真表)已逐項判定,缺的是戰鬥速度與護盾分面模型 |
    | **仍缺、可做** | 20 | 戰鬥艙、強化船體、保安站、戰鬥掃描器、匿蹤力場、相位匿蹤、時間扭曲加速器、阿基里斯/測距瞄準器、超載電容、快速飛彈架、轟炸機庫、多相護盾、能量吸收器、傳送器…… |
    `gamedata/missile.go` 早就把手冊 p.123 整段搬完了,每一個都附原文與精確數字:
    七個常數,**生產端全部是 0**。`battleVolley` 的註解把理由寫得很清楚:
      - **廣域干擾器只給自己那一格(130)。** 手冊另有「對艦隊其餘船艦 +70,且不與其他
      - **一艘船只有一個 Special 槽。** 原版可以同時裝干擾器與慣性穩定器,remake 不行
        ——既有設計限制(見 `Ship.Special`),不是這一項引入的。
      - **匿蹤裝置的 50% 未命中沒接。** 手冊寫明「僅在裝置**啟動**時」,而 remake 沒有


    ### **第 68 項(元件盤點+飛彈防禦)那一桶的第一批,外加修掉第 68 項(元件盤點+飛彈防禦)自己的一個缺失**(2026-08-08)。


        ### 先講缺失

        原因值得記:第 68 項(元件盤點+飛彈防禦)是從 `gamedata/missile.go` 那一側往回找元件的——那個檔案只收
        (而 `gamedata.BeamDefense`(移植自 openorion2 `ShipDesign::beamDefense`)其實早就有
        `inertialStabilizer` 參數並且加 50——又一個「規則在、呼叫端不在」,只是這次它躲在
        | 元件 | 手冊原文 | 落點 |
        |---|---|---|
        | 戰鬥掃描器 | increases the ship's chance to hit with beam weapons by **50** | `combatant.atk` |
        | 強化船體 | **triples** the amount of structural damage a ship can sustain | `combatant.hp` |
        | 多相護盾 | increasing the maximum amount of damage that they can absorb by **50%** | `combatant.shield` |
        | 偵察實驗室 | Frigate = **1**, Destroyer = **2**, Cruiser = **4**, Battleship = **8**, Titan = **16**, Doom Star = **32** | 研究階段 |
        偵察實驗室那一條罕見地把整張表都列出來了。艦隊研究**併進** `TotalResearch` 而不是
        寫在 `internal/shell/ship_systems.go` 檔尾,逐項列。摘要:
          - **機制不存在**:保安站(登艦戰,第 60 項(打得準也閃得掉))、增強引擎與時間扭曲加速器(戰鬥速度/
          - **加了不影響任何東西**:戰鬥艙(「add equipment space」——remake 沒有逐元件佔格)
          - **會讓兩條戰鬥路徑不一致**:測距瞄準器(把距離縮成 1/3,而快速結算固定 range=2,
          - **要先重構**:結構分析儀與阿基里斯瞄準器都要動 `ResolveShotWithMods` 的傷害鏈,
          - **強化船體只接了結構那一半。** 手冊同一句還說「tripling the amount of damage
          - **偵察實驗室的弱點分析沒接。** 手冊說它讓艦隊「analyze the opponent's biology or
          - **`PlayerState.FleetResearch` 是每回合重算的衍生值。** 它會進存檔(整個 PlayerState
            都會),但在 `EndTurn` 開頭就被覆寫,舊存檔的陳舊值影響不到任何判定


    ### **傷害鏈收成具名結構,順手解掉卡在它後面的兩個系統**(2026-08-08)。


        第 68 項(元件盤點+飛彈防禦)把結構分析儀與阿基里斯瞄準器列在「要先重構」那一格,理由是:

        `ResolveShotWithMods` 的呼叫端長這樣:
          - **攻方/守方分開**,不會再出現「把守方的旗標填進攻方那一格」這種安靜的錯
        舊入口 `ResolveShot` / `ResolveShotWithMods` 保留為薄包裝,既有呼叫端與測試一行都沒改。
        | 元件 | 手冊原文 | 位置 |
        |---|---|---|
        | 結構分析儀 | the damage done by beam weapons that **penetrate an enemy ship's shields** is **doubled** | 扣完護盾**之後**加倍 |
        | 阿基里斯瞄準器 | all beam weapons **ignore the target's armor completely** | 與 AP 改造同一個開關 |
          - **飛彈與球形武器沒有跟著收。** 它們各自有不同的判定機制(`ResolveMissileShot` 已經
            在第 68 項(元件盤點+飛彈防禦)收了一個 `MissileDefenses`),硬要三條路共用一個結構會做出一個
          - 第 68 項(元件盤點+飛彈防禦)那一桶還剩 **14 個**,各自的擋門理由寫在 `ship_systems.go` 檔尾。

69. **戰鬥速度與引擎階:一張自我驗證的一手表,外加一個掃描器看不到的硬編值**(2026-08-08)。

    第 68 項(元件盤點+飛彈防禦)那一桶剩下的 14 個,擋門理由聚成幾個共同缺失的機制。最大的一個是
    **戰鬥速度**——它一個人擋住五個元件(增強引擎、牽引光束、停滯力場、時間扭曲加速器、
    黑洞產生器),而 `gamedata.CombatSpeed` 早就移植好且零呼叫端。

    `Current_Design_Min_Combat_Speed_`(0x6B82A)讀的是一張二維表:
      - **每升一階引擎全部 +2** ——與手冊那張戰機表的註腳逐字相符:
      - **最大速恆等於最小速 +10**,36 組全部成立
    所以 `gamedata/combat_speed.go` **只存第一列**,其餘用公式算。存整張表反而讓那三個
    `gamedata` 裡一整組公式吃 `ftlLevel`(飛彈速度、飛彈光束閃避、戰機速度),
    **硬編成 1**,所有戰機不論科技多高都跑得一樣慢。它在 `cmd/moo2/tacticalfighter.go`,
          - **引擎階模型**(`driveLevel`):取最高已研究的引擎。MOO2 的引擎是自動升級的
      - **主動權排序**:手冊給了確切公式「A ship's initiative is equal to its current
      - **增強引擎**(+5 戰鬥速度)——第 68 項(元件盤點+飛彈防禦)把它列在「機制不存在」那一格,現在成立了。
      - **戰機的引擎階與裝甲級數**改用真值(那個硬編的 1 / 0)。
      - **min/max 之間怎麼取,remake 不模擬。** 反組譯的 `Current_Design_Combat_Speed_`
      - **移動距離還沒限制。** 戰術棋盤只有 8×6,而速度值是 13..30,直接當「一回合走幾格」


    ### **戰術移動不再是瞬移:棋盤比例尺從一手尺寸推出來**(2026-08-08)。


        第 69 項(戰鬥速度與引擎階)把戰鬥速度接上了,但移動距離**還是沒有限制**,理由寫得很明白:

        `Assign_Combat_Grids_`(0x46CC8)開頭把整張格點清成 0xFFFF:
        **那是錯的**:同一個函式稍後會把陣亡艦從 `t.player` 壓縮掉,長度變短、索引往前移,
        而 `moveLeft` 還停在舊長度——選中第 3 艘會讀到第 5 艘的移動力,而且陣列會越界。
          - **敵方艦不受移動限制**:`genEnemyFleet` 造出來的敵艦沒有引擎/艦體設計資料,
          - **牽引光束/停滯力場仍然卡著**:它們要的不是「能走幾格」,是「**讓對方**走不動」


    ### **牽引光束與停滯力場:第 64 項(武器傷害真表)那個「卡在機制」的判定終於解掉**(2026-08-08)。


        第 64 項(武器傷害真表)盤點手冊 p.127 的特殊武器,把這兩個歸進「卡在機制,不是卡在資料」:

        | 武器 | 手冊效果 | 當時缺什麼 |
        |---|---|---|
        | 牽引光束 | reduces target speed | **戰鬥速度模型** |
        | 停滯力場 | places target in suspended animation | **「這一輪不能動」的狀態** |
          - **一艘船只有一個 Special 槽**,所以手冊的「multiple Tractor Beams … cumulative」
          - **AI 不會用**:敵方艦由 `genEnemyFleet` 生成,沒有元件設計資料。既有簡化。
          - **「可被登艦」那一半沒接**:手冊說被定住的船 can be boarded,而登艦戰機制不存在
          - **時間扭曲加速器仍然沒接**:它要的是「一回合行動兩次」,而 remake 的戰術回合是

70. **陀螺去穩器:擋門理由是對的,而正解不是照著它做**(2026-08-08)。

    第 64 項(武器傷害真表)把陀螺去穩器擋在外面,理由寫得很具體:

    那四個目前走 `SpecialOptions`(建造成本)而不是佔格表,先記在 `shipspace.go` 裡,
      - **只有空間壓縮器與陀螺去穩器豁免盾甲**,脈衝星沒有那一句。第 64 項(武器傷害真表)就標過
      - **剩下的 10 個**(超載電容 / 快速飛彈架 / 時間扭曲加速器 / 轟炸機庫 / 能量吸收器 /
        `ship_systems.go` 檔尾。其中前三個共用同一個缺失機制:**回合內的行動次數**


    ### **一回合開幾次火:建一次機制,一次解三個元件**(2026-08-08)。


        第 70 項(陀螺去穩器)末尾記著:剩下的 10 個裡有三個卡在**同一個**缺失機制。這一項建那個機制。

        | 元件 | 手冊原文 | 適用範圍 |
        |---|---|---|
        | 超載電容 | allows a ship's **beam weapons** to fire twice in a single turn | 只有光束 |
        | 快速飛彈架 | allow a ship to fire **two volleys of missiles** in a single turn | 只有飛彈 |
        | 時間扭曲加速器 | gain an **additional round of activity** at the end of each combat round | 不分武器 |
          - **格子戰術**有真的回合交界,所以完整實作充能:`TacticalAdvanceCharge` 在回合結束時
            依 `Fired` 更新 `Charged`。
          - **快速結算**沒有跨回合狀態(它就是一次齊射),所以充能一律視為滿。
        `battleVolley` 的迴圈體有 60 行(四種武器類型分流),要連射就得重複它。
        只是把 `attackers[i]` 換成參數、把 `break`/`continue` 換成 `return`。
        ⚠ `shots` 的零值是 0,那會讓船完全不開火——所以呼叫端夾在至少 1。
          - **AI 不會用**:敵方艦由 `genEnemyFleet` 生成,沒有元件設計資料。既有簡化。
          - **剩下 7 個**(轟炸機庫 / 能量吸收器 / 傳送器 / 匿蹤力場 / 相位匿蹤 / 保安站 /
            戰鬥艙)各自的理由仍在 `ship_systems.go` 檔尾。它們**不共用機制**,


    ### **轟炸機庫:註解說四種,程式碼只有兩種**(2026-08-08)。


        第 70 項(陀螺去穩器)末尾把預期調低了:剩下 6 個不共用機制,要一個一個評估。
        第一個評下來是**最便宜的那個**——因為模型早就在。

        `shell.FighterKind` 的檔頭:
        **註解說四種,底下兩個。** 而 `gamedata` 那一側:
        **加一個不會做任何事的型別只是把洞藏起來**——現在 `FighterKind` 有三個,


    ## 八、盤點方法與文件政策(2026-08-08)

71. **探針 ③ 的另外半邊:shell 與 cmd 內部函式**(2026-08-08)。

    活表上 ③ 那一格留著一句話:「⚠ **cmd/ 與 shell 內部函式沒掃過**」。
    第 65 項(種族特性31格)的掃描器只看 `gamedata.X(...)`——**跨套件呼叫**,所以整個 shell/cmd 內部
    從來沒被掃過。第 69 項(戰鬥速度與引擎階)那個硬編的 `ftlLevel=1` 就是在 cmd/ 撞到的,那是撞運氣。

    內部函式還多了方法呼叫、多行引數、同名歧義,所以這次直接 `go/ast`:
    第一版噪音很大:`t.pStart` 這種**欄位存取**被當成固定值了。我原本想放行的是
    `gamedata.TECH_X` 那種跨套件具名常數,而 `ast.SelectorExpr` 兩者都是。
    收緊成「`X.Sel` 的 `X` 必須是已知套件名」之後,60 條噪音變成 **17 條線索**。
    | 缺口 | 手冊怎麼寫 | remake 先前 |
    |---|---|---|
    | 硬化護盾只在光束路徑生效 | 「reduces the damage of **each enemy attack** — by 3 points」 | 飛彈與球形武器的呼叫端恆傳 `false` |
    | 攻方掃描器對飛彈干擾的抵銷 | 迅子「lowering the target's Missile Evasion by **20 points**」、中子 **40** | 恆傳 `0` |
    | 戰鬥掃描器的掃描範圍 | 「ships equipped with Battle Scanners have a scanning range **2 parsecs greater**」 | 沒接 |
    | 三個掃描科技的偵測距離 | 空間 **2** / 迅子 **4** / 中子 **6** parsec | 自編 4/8/6,**而且順序反了** |
    第四列最值得記。`gamedata/detection.go` 的檔頭曾把三個掃描科技的 parsec 值標成「近似」,
    第 68 項(元件盤點+飛彈防禦)加戰鬥掃描器時,我寫了一條 `TestBattleScannerRaisesOnlyAccuracy`。
    它就是基準本身。`ParsecToNormalized` 的調參註解寫的目標是「基礎 2 + 母星星基 2 = 4 parsec」,

72. **盤點:「元件表有」不等於「效果有接」**(2026-08-08)。

    使用者要求盤點剩餘工作。照規矩先查證再回答——結果查證本身變成這一項。

    **隱形裝置在 `SpecialOptions` 裡,而整份程式碼沒有任何地方讀它。**
    最後那一句的語意,就是第 70 項(陀螺去穩器)為了超載電容建的 `Fired`。
    | 元件 | 為什麼現在做得了 |
    |---|---|
    | 能量吸收器 | 第 70 項(陀螺去穩器)建了跨回合的 `Charged`/`Fired`;手冊指名的例外(位移裝置)第 68 項(元件盤點+飛彈防禦)也在 |
    | 戰鬥艙 | `ShipHullSpace` 與 `ShipDesignSpaceUsed` 都在 |
    | 相位匿蹤 | 只需要「不可被選為目標」一個旗標 + 回合數(`t.round` 已有) |
    `internal/shell/ship_systems.go` 檔尾那份「這一輪沒有接的,與各自的理由」,


    ### **三份文件對齊現況**(2026-08-08)。


        使用者要求更新 `CLAUDE.md`、`CONTEXT.md`,並重新盤點 worklist。

        - **快速結算 / 格子戰術** — 戰鬥的兩條路徑。第 68–71 項有五項的缺陷都是
        - **元件表有 ≠ 效果有接** — 第 72 項(元件表有≠效果有接)的教訓,直接收成一個詞條。
        - **擋門理由** — 附上「會過期、而且更糟的是可能當初就寫錯」。
        - **狀態指紋 / 畫廊 / 棘輪** — 三個驗收工具,各自寫明**它不能證明什麼**
        順帶查了 CLAUDE.md 自己列的交付項有沒有被追丟——`PLAN.md`、致謝、文化現象、


    ### **第二次 review:錯誤斷言的副本比原件多**(2026-08-08)。


        使用者要求「再 review 一次確保沒有遺漏」,並補了一句「刪除錯誤的記憶」。

        第 72 項(元件表有≠效果有接)只動了 `WORKLIST.md`。這一輪掃過所有 markdown,發現同一批假斷言在
        | 檔案 | 假斷言 | 事實 |
        |---|---|---|
        | `README.md` | 「行星表面 + 建築 sprite 擺放卡在幾何表尚未抽出」 | 7×7 角點表與 `Cache_Load_Bldg_` 算式都抽出來了 |
        | `README.md` | 「戰機/航母未做」「多 AI 目標選擇為索引順序」 | 兩者都完成(第 27/70 項、`AIEnemyColonyValue`) |
        | `REMAKE-COMPLETION-ASSESSMENT.md` | 「需先建基礎設施 — 戰機/航母」 | 同上 |
        | `oracle-comparison-20260712.md` | 表格列掛著 ❌「點 Continue 靜默無反應」 | **同一份文件末尾早就寫了「兩項都已完成」** |
        | `ebiten-notes.md` §6 | 三條「待續」 | 全部已了結或方向已改 |
        | `PLAN.md` | Screen 抽象層、widget 樹移植 | 方向已改,沒有執行 |
        | `component-values.md` | 「裝甲缺科技倍率」 | 第 67 項(裝甲科技倍率)已解 |
        | `newgame-flow.md` | 「版面為合成近似,尚未像素對齊」 | 2026-08-07 已用執行檔立即數座標 |
        | 本文 | 「戰機/航母戰鬥子模型」仍缺 | 同上 |
        `CLAUDE.md` 的「已知會被誤引的過期錨點」清單新增了一條:
        這幾輪我自己新寫的可證偽數字全部回頭對過:49 處 `DrawImage`、畫廊 34 張、譯表 5,046 條、
        24 份 TSV、`SpecialOptions` 32 項、`docs/tech` 54 篇、棘輪 16——**逐項相符**。


    ### **第三次 review:改用掃描器,不再靠「我記得要查什麼」**(2026-08-08)。


        第 144、145 項都是我先想到「哪裡可能有問題」再去查。那有個明顯的上限:
        **想不到的就查不到**。這一輪換方法。

        | 檔案 | 假斷言 | 事實 |
        |---|---|---|
        | `tactical-combat-weapon-kinds.md` | 「`WeaponOptions` 沒有任何武器分類到 spherical」 | 第 64/70 項補了三把 |
        | `01-gap-report.md` | 「`ShowRelocationLines` 已建但還沒有 UI 可以切」 | 第 23 項(多艦隊模型)接進遊戲選單 |
        | `01-gap-report.md` | 「`AIOpponent` 沒有 Leaders 欄位——那一整層還不存在」 | 有了;維護費仍未對 AI 收,但理由變成「沒接」 |
        | `WORKLIST.md` | 「AI 一個星系只會有一個殖民地」 | `aiExpansionCandidates` 含自己已有殖民地的星系 |
        | `WORKLIST.md` | 「沒有任何一艘船有艦員等級,`shell.Ship` 沒那個欄位」 | `Ship.CrewXP` 在,第 60 項(打得準也閃得掉)接上攻防兩側 |
        | `moo2-formulas-reference.md` | 「`RaceTrait` 沒有任何函式把數值套進公式」 | 第 65/65 項全接 |
        | `ai-fiscal-solvency.md` | 「`Maintenance` 尚未接線」 | `engine/ai.go:19` 傳進 `decider.ColonyJobs` |
        | `spy-system.md` | 「`interactive.go` 未接間諜畫面/按鈕」 | 第 51 項(間諜UI)接進種族關係畫面 |
        | `colony-economy-maintenance.md` §2.2 | 「BC 保證單調遞減至負值」 | 2026-07-11 修掉,300 回合最低點 −3710 → −51 |
        掃描器把 `gameplay-systems-status.md` 的「AI 無 `ColonyBuildings` 追蹤」「`AIOpponent`
        無種族欄位」也列成可疑,查下去**兩條都仍然成立**(`AIOpponent` 的欄位清單裡確實沒有)。
        `ai-decision-modes.md` 的「`ModeOriginal` 仍回傳 `RemakeDecider`」同樣仍然成立(還掛著 TODO)。

73. **音樂曲目↔場景:不需要人耳,執行檔全寫著**(2026-08-08)。

    使用者:「音樂曲目可以從反組譯得知,目前有 ida_pro 了,這一點從記憶內改掉避免又混亂。」

    實際位址是 **`0x2484F`** 與 **`0x2496C`**。差 **`0x10000`**——**object base 沒加**。
    (`symbols_fixed.tsv` 第一筆 `___begtext = 0x10003` 就寫在那裡。)
    `Play_Background_Music_` 有 **15 個呼叫端**,一個死碼都不是。
    | 函式 | 行為 |
    |---|---|
    | `Play_Streaming_Music_` @ `0x24677` | 指定曲目。**編號 ≤ 100 → `STREAM.LBX` 索引 = 編號;> 100 → `STREAMHD.LBX` 索引 = 編號 − 100**(單一編號空間) |
    | `Play_Background_Music_` @ `0x2484F` | `clock() % 3 + 1` → **STREAM 1 / 2 / 3 隨機** |
    | `Play_Combat_Music_` @ `0x2496C` | `clock() % 3 + 4` → **STREAM 4 / 5 / 6 隨機** |
    `Play_Streaming_Music_` 的 `edx`(下一首)有兩個哨兵值:`-1` = 沒有下一首;
    `-2` = 播完接**隨機 STREAM 1..3**(與 `Play_Background_Music_` 同一組)。
    | 場景(除錯符號真名) | 曲目 |
    |---|---|
    | `Main_Antaran_Room_Screen_` / `Draw_…` | STREAMHD **#20** |
    | `Main_Council_Screen_` / `Draw_…` / `Draw_Council_Player_Voted_Winner_` | STREAMHD **#19** |
    | `Start_Main_Event_` / `Draw_Event_Screen_` | STREAMHD **#18** |
    | `Science_Room_` / `_Tech_Select_` | STREAMHD **#17**,播完接隨機 STREAM 1..3 |
    | `Design_Screen_` / `Draw_Design_Screen_` / `Draw_Generic_Replacement_Box_` | STREAM **#8** |
    | `Colony_Combat_Screen_` | STREAM **#10** |
    | `Draw_Diplomacy_Synch_Mode_` | STREAMHD **#`word_19AA44`**(逐族/依關係動態,見上一輪 §7.4 解出的 good/bad 規則) |
    | `Tactical_Combat_` | `Play_Combat_Music_` → STREAM 4/5/6 隨機 |
    | `main__0`、外交、議會、安塔蘭廳、艦艇設計、殖民地戰鬥、事件檢查… | `Play_Background_Music_` → STREAM 1/2/3 隨機 |
    **「主選單放哪一首」這個問題本身問錯了**——原版主選單走 `Play_Background_Music_`,
    per-colony 中隊計數欄位」。查了:`Fighter_Garrison_Strength_` @ `0x5F64C`
    **從帝國記錄當場算**(`imul edx, 0EA9h` + `dword_197F98` 取該帝國記錄,
    再讀 `[eax+0x16A]` / `[eax+0x136]` 兩個科技旗標),**沒有逐殖民地的中隊存量**。
    所以 remake 現行做法(`fighterGarrisonTierFor` 只看科技)**與原版一致**。
    `Calc_Tech_Value_` 階段 C–K 這一輪**沒有動**。不假裝兩項都查完了。

74. **文件政策:過期的斷言直接刪除**(2026-08-08)。

    **全 repo 一條規則,沒有例外:任何文件都不留錯誤斷言的內容。**
    不劃刪節線、不加「⚠ 原本寫著…」的訂正註記——留著就會被讀到,而且錯的那句還排在前面。
    要知道某項當初怎麼錯的,看 git log。

    - **正確結論與 RE 硬資料**——位址、資料表、公式。
    - **「這件事做完了」的完成標記**(`~~題目~~ ✅ 已建)——那不是錯誤斷言,
    - **教訓,一律正面敘述**。例如 §7.5 只留:「零命中時先做正對照;
    `audio-track-map.md` 的待辦裡還有 `obj1+0x9082` 這種**除錯表偏移**,
    正是第 73 項(音樂場景表)踩到的(少加 `0x10000` 的 object base)。


    ### **本文的職責:RE 硬資料 + 工程日誌,不是現況**(2026-08-08)。


        **剩餘工作表在 [`WORKLIST.md`](../../WORKLIST.md) 頂端,那裡是唯一活來源。**
        `CLAUDE.md` / `CONTEXT.md` / `HANDOFF.md` / 五份 kickoff 的指標全部指向那裡。

        | 內容 | 量 |
        |---|---|
        | RE 硬資料(執行檔立即數、資料表、公式) | **196 個十六進位位址、131 列表格** |
        | 工程日誌(每一項是怎麼被找到的) | 150 個編號項 |
        那些位址是一次性成本的知識——`Play_Background_Music_ @0x2484F`、行星表面 7×7 角點表、
        種族特性換算表、`stride 0xEA9`、VA→檔案位移 `0x7E694`、戰鬥速度表、
        `Assign_Combat_Grids_` 的 81×68。每一項重挖都是數小時起跳。


    ### **本文的錯誤紀錄清除**(2026-08-08)。


        依第 74 項(文件政策)的規則,把本文裡 16 處**複述錯誤斷言**的段落刪掉,只留正確結論:

        | 改成 |
        |---|
        | 殖民地畫面的版面與框架美術**都有真值來源**,而且兩者都不在 COLONY.LBX 裡 |
        | 飛彈基地/地面砲台的空間預算模型**早就在**,第 11 項(48棟建築盤點)把它當成缺口是漏查 |
        | `Cycle_Ship_Icons_` 不是動畫,而艦隊圖示在原版也不會動 |
        | 安塔蘭母星防禦艦隊的數字**執行檔裡有**,而且組成不是同級的 |
        | 手冊三個掃描科技的條目**各自明寫著自己的 parsec 值** |
        | 元件表的「舊理由」欄 → **「為什麼現在做得了」** |


    ### **150 項煉成 12 條規則**(2026-08-08)。


        使用者:「合併規則,150 條太多了。」

        | 組 | 條數 | 內容 |
        |---|---|---|
        | 判斷「還缺什麼」 | 5 | 擋門理由會過期也可能當初就錯、元件表有≠效果有接、清單會過期、「已經沒有剩下的了」最貴、掃描器只縮小範圍 |
        | 改動的完整性 | 4 | 兩條戰鬥路徑、一個元件兩個效果、資料層跟著實作層缺、測試名稱裡的「只」是斷言 |
        | 逆向 | 3 | 一手贏二手(openorion2 沒有 ≠ 沒有真值)、查 xref 圖不要 grep `.asm`、零命中先做正對照 |
        | 驗收 | 4 | 測試綠≠對齊原版、畫廊逐位元、狀態指紋、不臆造 |
        **150 項就此降級成參考資料**;`CLAUDE.md` 與 `CONTEXT.md` 的指標同步改成


    ### **151 項壓縮:8,163 行 → 3,582 行**(2026-08-08)。


        使用者:「150 項太多了,我希望合併與刪除重複的。」

        第 74 項(文件政策)只是在最前面加了一張規則總表,檔案本身沒有變小。這一項動 151 項本身。

        ### 保留什麼、刪什麼

        每一項只留三樣:**標題**、**結論段**、以及**帶一手資料的行**
        (十六進位位址、表格、程式碼區塊、反引號識別字、粗體條目)。
        其餘敘述全部刪除——那些正是重複的部分,教訓已經收進最前面那 12 條。

        | | 壓縮前 | 壓縮後 |
        |---|---|---|
        | 行數 | 8,163 | **3,582**(−56%)|
        | 十六進位位址 | 196 | **197** |
        | 表格列 | 131 | **778**(先前的計數只數了頂層縮排的)|

        **一手資料一筆都沒少**,少的是同一個教訓的第 N 次複述。

        ### 分成八組

        編號**沒有動**(全 repo 有 41 處交叉引用),改成加八個分組標題:
        符號表與資料表 / 逐畫面重建 / 一星多行星與 AI / 網路多人 /
        手冊與反組譯校準 / 手冊數值忠實化 / 艦載元件忠實化 / 盤點方法與文件政策。

        ### 壓縮的代價,如實記錄

        機械式壓縮會切斷跨段落的引述——這一輪清掉 8 處被截成單行的孤兒引述。
        要看某項的完整敘述,`git log -p docs/re/01-gap-report.md`。

        無程式碼改動,截圖無差異。

75. **編號壓縮:152 → 75**(2026-08-08)。

    使用者:「1. 編號壓縮 2. 調整交叉引用的編號。」

    ### 合併

    22 組主題合併——續篇併回父項、同一條線的整組併成一項:

    | 併成 | 原本 |
    |---|---|
    | 資產與畫面完成清單 | 舊 1–5 / 7 / 9–20 |
    | 行星表面整條線 | 舊 26–31 |
    | 多艦隊整條線 | 舊 56–60 / 65 / 70 |
    | 一星多行星整條線 | 舊 61–63 / 66 / 67 |
    | 網路多人整塊 | 舊 72–79 |
    | 地面戰整條線 | 舊 83–86 / 89 |
    | 武器忠實化整條線 | 舊 123–128 |
    | 文件政策整組 | 舊 148–152 |

    子項內容全部保留,收成合併項底下的 `###` 小節。

    ### 交叉引用

    **802 處、106 個檔案**改寫完成——268 處在 Go 註解裡(83 個檔案),其餘在 markdown。
    做法是先建舊→新對照表(**含被併掉的項次 → 父項的新編號**),再機械替換
    `第 N 項` / `第 N/M 項` / `第 N–M 項` 三種寫法,最後掃一次驗證:
    **沒有任何引用指向不存在的項次**。

    ⚠ 唯一的「第 0 項」是 `buildqueue.go` 講建造佇列的第 0 格,不是本文項次,沒有被動到
    ——**機械替換要能分辨同形不同義的字串**,這次靠的是 0 不在對照表裡。

    無程式碼邏輯改動(只動註解),測試全綠、`gofmt` 乾淨。

76. **交叉引用不必再開文件才看得懂**(2026-08-08)。

    使用者:「調整交叉引用 802 處,避免多層查詢。」

    ### 問題

    `// 見第 68 項` 這種引用要**開一份 3,800 行的文件、找到第 68 項、讀完**才知道是什麼。
    程式碼註解裡有 268 處這種引用,而讀註解的人多半只是想確認一件事。

    ### 做法:短標寫進引用本身

    **717 處**裸引用加上短標——`第 68 項` → `第 68 項(元件盤點+飛彈防禦)`。
    一層就懂,不必跳轉。

    - 只動**單一項次且後面沒有緊接說明**的引用(後面已經有 `(`/`:` 的不重複加)。
    - 多重引用(`第 26/27 項`)與範圍引用維持原樣——加了會太長。
    - 另外在本文最前面加一張 **75 項速查表**(三欄),括號裡的短標就是那張表的內容。

    ### 這一輪的複查抓到兩個真的改壞

    **① 跨文件的項次被套用了 gap-report 的對照表。**
    `rules-implementation-audit.md 第 10 項` 被改成 `第 3 項`、
    `doc-audit-20260808.md 第 3/4 項` 被改成 `第 2 項`——**那些是別份文件的編號**。
    五處已還原。機械替換要先問「這個數字屬於哪個命名空間」,
    而我上一輪只驗了「有沒有指向不存在的項次」——**被改成一個存在的錯項次,那個檢查抓不到**。

    **② 短標貼到同編號的另一個小節上。**
    合併之後,舊的「分艦隊」那一項合併之後只是第 23 項(多艦隊模型)底下的一個小節,
    而先前手寫的括號還停在小節名上,於是同一個編號在不同地方掛著不同的名字。WORKLIST 的 49 列日誌也有同樣問題(列標題自己已經說明內容,
    短標反而指向別的小節)。已統一:同一項次在全 repo 只有一個短標。

    ### 順帶修掉合併時的一個 bug

    上一項的合併腳本有個缺陷:舊編號 1/2/3 在檔案裡出現**兩次**(Part C 的子清單也是
    頂層編號),`bounds` 被後者覆蓋,於是輸出了一份**逐行完全相同的重複區塊**,
    而且編號序列變成 1,3..76(缺 2)。已刪重複、整體下移一格,
    現在是**連續的 1..75**,失效引用 0。

    **教訓**:機械式重寫要驗證的不只是「有沒有壞掉的引用」,還有**輸出本身的完整性**
    ——這次是靠「項數 76 但唯一值 75」這個計數對不上才發現的。

    無程式碼邏輯改動(只動註解),測試全綠、`gofmt` 乾淨。

---

77. **原版元件表挖到了,五個「擋著」的元件同一輪全部接完**(2026-08-08)。

    ### 挖到什麼

    `Orion2.exe` 資料段 `0x17EEE0` 起、每筆 47 位元組的**特殊裝置表**,39 筆完整:
    裝置編號 / 解鎖科技 / **六個艦級各自的空間** / **六個艦級各自的成本**。
    格式與定位方式寫在 `docs/tech/special-device-table.md`,程式碼在
    `internal/gamedata/special_devices.go`。

    這張表終結了一句在 `SpecialOptions` 每一列重複了很多次的話:
    「成本是 remake 值,執行檔的元件表還沒挖到」。

    ### 欄位偏移怎麼定的(不是照長度猜)

    | 訊號 | 位址 | 推出什麼 |
    |---|---|---|
    | `Special_Devices_Available_` | `0x5F9EA` | 拿 +6 欄去查 `[帝國紀錄+0x117+科技編號] == 3` → **+6 是科技編號**,3 = 已研究 |
    | `Init_Internal_Space_` | `0x36470` | 讀 +8 欄 → **+8 起是空間** |
    | 戰鬥艙那一列 | — | +8 欄是**負數**,而手冊說它「add equipment space」→ 負佔格就是這句話的實作 |

    ### 兩道交叉驗證(抽完才發現對得上)

    - 39 筆的科技編號**逐筆**對上 remake 既有的 `TECH_*` 列舉,零不符。那份列舉先前是從
      別的來源建的。測試釘的是**執行檔讀出來的整數**,不是 `TECH_*` 常數,列舉一動就紅。
    - 戰鬥艙的負佔格 = 艦體空間的一半,連截斷都一樣:艦體 25/60/120/250/500/1200
      (**抄自手冊 p.121**)→ 戰鬥艙 −12/−30/−60/−125/−250/−600,12 = floor(25/2)。

    ### 順帶挖到巨型通量器

    `Total_Design_Space_`(`0x6E81F`)是 `imul 125` / `idiv 100`。手冊給了「+25%」這個
    比例,而**截斷方式只有執行檔給得出來**——巡防艦 25 → 31,不是四捨五入。
    同一趟確認 `[帝國紀錄+0x170]` 就是 `TECH_MEGAFLUXERS`(`0x170 − 0x117 = 89`)。

    ⚠ 巨型通量器 (+25%) 與戰鬥艙 (+50%) 看起來同類,**實作位置不同**:前者乘在「可用空間」
    上,後者是「已用空間」那一側的負值。混在一起做會算錯。

    ### 五個元件

    第 72 項(元件表有≠效果有接)的「B. 擋門理由已經過期或本來就錯」那一格清空了:

    | 元件 | 舊擋門理由 | 實際狀況 |
    |---|---|---|
    | 戰鬥艙 | 「remake 沒有逐元件佔格的造艦模型」 | **當初就寫錯**,那套一直都在;缺的是手冊沒給的數字,而那在執行檔裡 |
    | 隱形裝置 | ——(在元件表裡躺著,沒有任何程式碼讀它) | 手冊給了完整規則:未開火 +80 光束防禦、飛彈 50% 未命中、開火即失效、停火一整回合重隱 |
    | 相位匿蹤 | 「需戰鬥可見性」 | 只需要**不可被選為目標**一個旗標 + 回合數,兩者都有 |
    | 測距瞄準器 | 「快速結算固定 range=2」 | 只對一半——格子戰術傳的是真距離 |
    | 能量吸收器 | 「需儲能狀態」 | 第 70 項(陀螺去穩器)已經建了跨回合狀態 |

    測距瞄準器那一條手冊寫得特別小心:射程縮成 1/3 **只影響命中**,「the dissipation of
    damage potential is **not** affected」。所以射程等級要算兩次——先前 remake 只算一次,
    任何接法都會連傷害一起改,那正是手冊特地寫第二句要排除的事。

    ### 這一輪抓到的漏

    `ResolveBeamShot` 從第 68 項(元件盤點+飛彈防禦)起就有 `Target.HardShield`,而**兩條戰鬥路徑的光束分支
    都沒有填它**。`combatant.hardShield` 的註解甚至寫著「三條路徑都要吃到,之前只有光束
    路徑接了」——事實相反,光束是唯一沒接的那一條。

    **「Resolve\* 有這個參數」不等於「呼叫端有填」。** 結構欄位漏填拿到的是零值,不會編譯
    失敗;而既有的硬化護盾測試全都直接呼叫 `ResolveBeamShot` 並自己填 `HardShield: true`
    ——驗的是公式,驗不到呼叫端。新測試走 `battleShot`(真正的入口),並用**正對照**確認過:
    把修補拿掉,它會紅。

    ### 玩家可見契約

    - **能量吸收器的儲能是「下一次開火時自動射出」**,不是玩家的另一個按鈕。原版讓玩家挑
      時機;remake 的格子戰術一次點擊就是「這艘船對那艘船開火」,沒有第二個動作可以掛。
    - Ship Initiative 開啟時，快速結算與格子戰術都以敵我合併的主動權序列行動，儲能可跨回合
      累積；關閉時雙方分批行動，未使用儲能於完整戰鬥回合末清空。格子戰術以單場暫態 ID
      保護戰損壓縮後的行動定位，WAIT 會把目前艦移到合併序列尾端。

    驗收:畫廊 34 張逐位元,只有 `25_shipdesign.png` 變(特殊 1/32 → 1/36),已更新。
    **`30_netwait.png` 的狀態指紋不變**。全 package 測試綠,`gofmt` 乾淨。

---

78. **音樂場景表從文件搬進程式碼**(2026-08-08)。

    第 73 項(音樂場景表)把對照表解出來了,而**程式碼沒有跟著改**——`audiohook.go` 還跑著上一輪的時長啟發式
    猜測值,註解裡寫著「低信心、待人耳聆聽定案」。**解出來 ≠ 接上去**,這與第 72 項(元件表有≠效果有接)
    的「元件表有 ≠ 效果有接」是同一個形狀,只是換了個子系統。

    ### 改了什麼

    | | 先前 | 現在 |
    |---|---|---|
    | 編號空間 | `musicClips` 的 0-based 序號,只載 STREAMHD | 原版的**單一編號空間**:≤100 = STREAM 的 entry、>100 = STREAMHD 的 entry(−100),兩個 LBX 都載 |
    | 主選單 | 固定一首 | `Play_Background_Music_` → STREAM 1/2/3,**每次進去重擲** |
    | 星圖 | 固定一首 | 同上 |
    | 戰術戰鬥 | 固定一首(時長猜的) | `Play_Combat_Music_` → STREAM 4/5/6 隨機 |
    | 外交 | STREAMHD 序號 13/14/15(**當成 0-based,實際差一格**) | 編號 113/114/115 = STREAMHD entry 13/14/15 |
    | 科學室 / 事件 / 議會 / 安塔蘭廳 / 艦艇設計 / 殖民地戰鬥 | 沒有音樂 | STREAMHD #17/#18/#19/#20、STREAM #8/#10 |

    ### 一道事後才發現的交叉驗證

    實測玩家資料夾:`stream.lbx` 的 entryIDs 是 **[1 2 3 4 5 6 8 10]**。
    反組譯點名的 STREAM 曲目是 1/2/3(背景)、4/5/6(戰鬥)、8(艦艇設計)、10(殖民地戰鬥)
    ——**正好就是這八個存在的槽**,而它從來沒點過的 7 與 9,正是那兩個非 WAV 的空槽。

    這同時排除了另一種讀法:若編號指的是「第幾條 WAV」而不是 entry id,這個吻合不會成立
    ——而先前的 `bgmDiploBadPool = {13,14,15}` 正是被當成 0-based 序號用,等於**整組差一格**。

    ### 兩個講明的缺

    - **科學室播完不接下一首。** 原版 `Play_Streaming_Music_` 的 `-2` 哨兵會接隨機 STREAM 1..3;
      remake 的 Mixer 沒有「播完接下一首」這個機制。
    - **外交只有「關係差」那一組。** 「關係好」是 `該族 empire 記錄 offset 0x25 + 1`,
      那張逐族靜態表還沒追出來。

    `docs/tech/audio-track-map.md` 同步改寫:第三節換成反組譯的表,第 5.5 節那份逐場景
    猜測值**刪除**(被第三節推翻),5.1–5.4 的結構性結論保留(它們與反組譯一致)。

    驗收:畫廊 34 張逐位元**全部相同**(音樂不進截圖,這一輪本來就不該動畫面)。測試全綠。

---

79. **原版武器表與艦體表——順手抓到質子魚雷抄錯一格**(2026-08-08)。

    第 77 項(元件表真值)只抽了特殊裝置表,而同一區還有兩張:武器表(`0x17F807` 起,46 筆 × 28 位元組)
    與艦體表(`0x180020` 起,stride 36)。定位方式一樣——**找讀它的函式**,
    `Special_Devices_Available_` 在「這一格是武器」那條分支用 `imul eax, 1Ch` + `word_17F80D`,
    而 `0x17F80D − 0x17F807 = 6`,與特殊裝置表同一個版面家族。

    ### 武器表抓到的抄錯

    remake 的**質子魚雷傷害 25、佔格 20**,而那是 **A-M 魚雷**那一格的數字。
    手冊 p.125 那張表在 PDF 抽取後欄位打散成:

    ```
    A-M Torpedo / 20 / 30 / 25 / Proton Torpedo / 40 / Plasma Torpedo / 120
    ```

    當初讀成「Proton/A-M Torpedo 25」——**兩個名字併在一起寫**,那個註解自己就洩漏了不確定。
    執行檔逐格給出:A-M 20 格 25 傷、質子 **30 格 40 傷**、電漿 40 格 120 傷。
    手冊那幾個數字一個不差,只是欄位歸屬相反。已訂正三處。

    **教訓**:註解裡出現「A/B」這種併寫,通常代表當初分不出來。那是可以主動去找的訊號。

    ### 順帶證實一個版本差異

    電漿砲:元件清單存 20(1.50),執行檔給 30——因為反組譯的是 **1.31 的 exe**。
    `RuleProfile.PlasmaCannonMaxDamage` 早就寫著 1.31 = 30。兩份獨立來源在這一格對上了,
    **測試也照這個讀法寫**,不是把它當成不符去改。

    ### 兩件手冊沒講而這張表講了的事

    - **飛彈本體的 Size 是 0**——佔格在彈架上(手冊只給了彈架的 10/20/30/35/40)。
      remake 當時對四種飛彈估 10，那是該輪尚未閉合彈架前的建模取捨。
      **2026-08-24 勘誤**：後續 IDA 已證實 2/5/10/15/20 發映射 10/20/30/35/40，
      正常艦艇設計與兩條戰鬥路徑均已接線；見 `missile-ammo-rack-audit-20260824.md`。
      ⚠ **不要「照執行檔訂正成 0」**,那會讓飛彈完全不佔空間,比現在更失真。
    - 編號 40..45 沒有解鎖科技,是安塔蘭/太空怪物專用武器。

    順帶解出兩個欄位的語意:`+8` 是**類別**(光束/飛彈/魚雷/炸彈/戰機艙/特殊)、
    `+9` 是**彈藥數**(光束 255 無限、飛彈 5、魚雷 2、炸彈 10、戰機艙 1、反飛彈火箭 20)。
    戰機艙那三筆的第二組傷害欄逐格對上手冊 p.127 的攔截機 1-4 / 重戰機 4-16 / 轟炸機 2-7。

    ### 艦體表:兩欄確定,一欄不下結論

    `+0` 空間(25/60/120/250/500/1200)與 `+2` 陸戰隊(5/8/12/20/30/50)**與手冊 p.121 逐格相同**,
    所以記錄起點與 stride 是確定的。

    `+34` 那一欄前五格是 70/250/600/1500/4000,看起來像成本,而**末日之星是 400**
    ——比泰坦低一個數量級。可能欄位認錯、可能有特例。**在解釋得通之前不寫進程式碼。**
    這一格如果硬填,就會變成下一個「當初就寫錯」的擋門理由。

    ### 成本欄為什麼兩張表都沒接線

    執行檔的武器成本是 remake 尺度的約 1/4,而**艦體成本的方向相反**(remake 18、執行檔 70)。
    只換一邊會把相對平衡弄壞,而艦體成本那一欄還沒解釋得通。兩邊都先不動,數字留在表裡。

    驗收:畫廊 34 張逐位元**全部相同**。全 package 測試綠。

---

80. **登艦戰:手冊自己說了要復用哪套解算器**(2026-08-08)。

    第 80 項(登艦戰)把保安站 / 傳送器 / 突擊艇三項接入共用規則;手冊把解算方式
    **直接指回地面戰**:

    > The Marines boarding the ship and those defending the ship fight it out
    > **in the same way as ground troops do when a colony is invaded**.

    地面戰的原版解算器(`ResolveGroundCombatOrig`,抄自 `Resolve_Ground_Combat_` @ `0xEC601`)
    早就在了;陸戰隊人數(`ShipHullMarines`,手冊 p.121)也在;部隊艙翻倍也接好了。
    **缺的從來不是公式,是沒有人把那兩句話接起來。**

    ### 接了什麼

    | 元件 | 手冊規則 |
    |---|---|
    | 突擊艇 | 戰機家族(「fighters (like the Interceptors)」),4 架一隊、每架載 1 個陸戰隊單位、速度 6、血量 3;抵達目標放下陸戰隊,**一律嘗試奪船**;放完人艇就漂著 |
    | 保安站 | 守方陸戰隊 **+20** |
    | 部隊艙第二半 | 「The additional marines both defend the ship **and can board enemy ships**」——先前只接了運力那一半 |

    奪船與突襲兩種目的都做了:奪船是殺光守軍換旗;突襲**不傷結構/裝甲/護盾**,只拆內部系統
    (每一下 1~2 個)。手冊那句「smaller systems are more likely to be destroyed」在 remake
    沒有落點——一艘船只有一個特殊系統槽,沒有大小不同的一堆系統可以挑。只做了一半,標明。

    ### 傳送器:分面護盾與硬化護盾例外已接

    手冊的前置是「面向攻擊方的護盾**已經被打穿**」,射程是 **12 格**。目前
    `CombatShip` 保存四面護盾容量,武器命中依來向扣除對應分面,`ShipBoardingReachAgainst`
    在 12 格內檢查該面是否失效；硬化護盾仍會阻擋傳送器。這條鏈有單元測試覆蓋完整容量、
    最後一點傷害、超出 12 格與硬化護盾例外。

    分面索引採固定世界座標四向近似；原版艦身旋轉、方向命名與戰機自動尋找最弱面仍保持
    未解，不把這層近似宣稱成原版座標對應。

    ### 匿蹤力場:擋門理由過期了,但結論不變

    舊理由是「AI 艦隊是抽象戰力、沒有地圖座標」——第 47 項(AI艦隊移動)之後 AI 艦隊有 `FleetStar` 了,
    **理由確實過期**。重查之後結論仍然不變,但換了個理由:AI 的出兵決策讀的是玩家**殖民地**
    (`aiLaunchRaidFleet`),從來不讀玩家艦隊位置。所以「在星圖上對敵方隱形」在 remake 仍然
    沒有消費端——缺的是「AI 會不會攔截玩家艦隊」這件事。

    這一項已從 `ship_systems.go` 的擋門清單移出:**留在那裡會讓人以為理由還是舊的那個**。

    ### 順手把佔格的退路從估計值換成武器表

    牽引光束 / 停滯力場 / 反飛彈火箭 / 三種戰機艙 / 突擊艇在原版是**武器**,不在特殊裝置表裡,
    先前一律落回「艦體空間 5%」的估計。第 79 項(武器表與艦體表)抽出武器表之後它們有真值了。
    現在只剩**戰鬥電腦**一項吃估計(原版的電腦是獨立槽,兩張表都沒有它)。

    ### 兩個講明的建模取捨

    - **奪船只表現成「那艘船退出戰鬥」**,不是真的換手。手冊說奪船後還要贏下整場戰鬥才留得住
      (心靈感應種族除外),真的把船搬過來要動戰後結算與艦隊清單兩處。
    - 快速結算的登艦**一場一次**——那條路徑沒有回合的概念,而突擊艇本來就是一次性的。

    驗收:畫廊 34 張逐位元,只有 `25_shipdesign.png` 變(特殊 1/36 → 1/38),已更新。
    狀態指紋不變。全 package 測試綠。保安站 +20 那條測試用**正對照**確認過會紅。

---

81. **`-play` 簡約殼整個刪掉**(2026-08-08)。

    活表上這一項的描述是:「旗標仍在 `main.go:189`,連帶 `colonyview.go` 等一批畫面是死碼。
    ⚠ `HONEST-STATUS` 稱它『已淘汰』指的是**不再是主路徑**,不是程式碼已移除。」

    ### 刪之前先查了兩件事(柵欄原則)

    1. **`colonyview.go` 不是它的東西。** 那份描述把兩件事綁在一起了——`colonyview.go`
       走的是 `-colony` 旗標,是與 `-lbx` / `-race` 同一類的**開發用單畫面檢視模式**,
       不是簡約殼的一部分。留著。
    2. **`transition` 與 `screen` 兩個型別住在 `play.go`,但十來個檔案在用。**
       它們不是簡約殼的東西,只是當初先寫在那裡。拆到 `screen.go`,`play.go` 才刪得乾淨。

    刪掉 448 行 + 兩個旗標(`-play` / `-play-record`)+ `main.go` 的分支。

    ### 順帶清掉兩處「已淘汰」的斷言

    - `lang_gap_test.go` 的整檔豁免清單裡有一筆 `"play.go": "已淘汰的 -play 簡約殼"`
      ——檔案沒了,豁免也該沒。
    - `HONEST-STATUS.md` 寫著「程式碼裡唯一寫死、會顯示到畫面上的英文在 `cmd/moo2/play.go`」
      ——那句現在是**「沒有了」**,不是「在某個已淘汰的地方」。

    **「已淘汰」與「已移除」是兩個狀態**,而文件把前者寫成後者的次數比想像中多。

    全 package 測試綠,棘輪 16 不變(那個檔案本來就在豁免清單上,刪掉不影響計數)。

---

82. **音樂那兩個講明的缺補完——其中一個要找的東西根本不存在**(2026-08-08)。

    第 78 項(音樂接線)接線時留了兩個缺,兩個都補上了。

    ### 缺一:科學室播完不接下一首

    原版 `Play_Streaming_Music_` 的 `edx = −2` 哨兵是「這首播完接隨機 STREAM 1..3」,
    而 remake 的 Mixer 只有無限迴圈播放。加了 `PlayBGMOnce` + `BGMFinished`,
    互動主迴圈每幀 `tickBGM()` 檢查。

    ⚠ `BGMFinished` 必須先看「這首是不是單次播放」再看 `IsPlaying()`——
    迴圈播放的曲子永遠不會結束,而「還沒開始播」與「播完了」在 `IsPlaying()` 上是同一個值。

    ### 缺二:外交「關係好」那一組——**那張表不存在**

    上一輪記著:「`_diplomacy_good_music` 證實為逐族資料表驅動,尚未取得資料表本身數值。」
    於是待辦上掛著「追出 `+0x25` 欄位的逐族預設值靜態表」。

    **沒有那張表。** `sub_12983`(帝國建立)顯示 `[帝國紀錄+0x25]` 就是**種族索引**本身:

    ```asm
    mov  eax, 0Dh              ; 13 族
    call Random_               ; 隨機挑一個
    ...                        ; 與既有帝國比對,重複就重挑
    mov  [eax+25h], dl         ; 寫進新帝國
    mov  esi, dword_192630[ecx*4]  ; 用同一個索引查種族名字串
    ```

    所以 `_diplomacy_good_music = [+0x25] + 1` 就是字面上的「**種族索引 + 1**」
    (STREAMHD 1..13),中間沒有任何表。remake 這一側用 `diplomatRaceIndex`
    (0..12,原版字母序,已對 RACESEL/DIPLOMAT 逐族核實)——與 `[+0x25]` 同一套編號。

    **教訓**:「這是資料表驅動的」與「有一張資料表」是兩回事。前者只表示值來自某個欄位,
    而那個欄位可能就是索引本身。上一輪把前者寫成後者,於是待辦上掛了一個找不到的東西
    三個星期——**找不到的原因是它不存在,不是沒找到**。
    這與第 73 項(音樂場景表)的「零命中先做正對照」同源:查詢回空要先問「我在找的東西存在嗎」。

    ### 仍然沒解的一小塊(講明)

    原版依**什麼條件**判定「關係好/壞」還沒追出來——那在 `Start_Diplomacy_Music_` 的
    呼叫端,不是這兩個變數的賦值處。remake 用關係分數 >= 0 當分界,**那是 remake 的讀法**。

    驗收:畫廊 34 張逐位元全部相同(音樂不進截圖)。全 package 測試綠。

83. **譯表烘進執行檔——而「改完就以為好了」被實測打回來兩次**(2026-08-08)。

    活表:「`cmd/moo2` 需要可覆寫 assets/i18n 路徑(或 `go:embed`),讓 macOS `.app` 不必
    launcher script 繞路。」問題不是打包,是**執行檔對當前工作目錄的隱性依賴**:
    每個模式都寫著 `os.DirFS("assets/i18n")`,只有從 repo 根目錄跑才找得到譯表。

    ### 實測打回來兩次

    第一次:改掉 5 處 `reg.LoadFS(os.DirFS("assets/i18n"))` 之後,
    **從 `/tmp/elsewhere` 跑仍然掛**——還有七處直接 `os.Open("assets/i18n/xxx.json")`
    讀單一檔案,那七處不走 `LoadFS`,grep `LoadFS` 找不到它們。

    第二次:把所有 `"assets/i18n/xxx.json"` 字面改成 `"xxx.tsv"` 之後,**還是掛**——
    `interactive.go` 有第二個譯表載入點,字面已經改對了卻**還走 `os.Open`**。
    路徑字面的檢查放它過去了。

    **兩次都是「從別的目錄實際跑一次」抓到的,不是靠讀程式碼。**
    從 repo 根目錄跑永遠看不出這個問題——這就是為什麼防呆測試有兩道:
    ①沒有 `assets/i18n` 字面 ②沒有 `os.Open(...tsv...)`。第二道就是第二次踩到的那個形狀。

    ### 為什麼只烘譯表

    字型是使用者自備的 TTC(授權不明)、遊戲資料是玩家的正版 LBX,兩者都不能進執行檔
    (CLAUDE.md 的 `[HARD]`)。譯表是本專案自己寫的,而且它正是「換一台機器就跑不起來」的那一份。

    當時 `go:embed` 讀不到 `../../assets`（不能跨出套件目錄），所以曾建立 `cmd/moo2/embedded/i18n/`；2026-08-26 已依外部文案決策移除
    有一份副本,`TestEmbeddedI18NMatchesAssets` 逐位元比對兩份——**不同步就會紅**。
    `-i18n <dir>` 保留為開發覆寫(改譯表不必重編)。

    驗收:`cd /tmp/elsewhere && moo2 -game …` 真的產出截圖。畫廊 34 張逐位元全部相同。

84. **名稱池改存英文原文——英文模式最大的一塊缺口**(2026-08-08)。

    活表:「`internal/` 有 2,188 條寫死中文,1,501 條是星名/艦名池(原版英文池的譯本,
    **接得回來**)。」這一項就是把那 1,501 條接回來。

    ### 它為什麼不是「漏翻」

    池子存中文的話,英文模式的星圖上會出現中文星名——而那**沒有任何 `tr()` 補得回來**,
    因為資料本身就是中文。這與 UI 字串的缺口是不同的問題:UI 缺口補一個 `tr()` 就好,
    資料缺口要換資料。

    ### 英文池是還原出來的,不是重抄的

    先前的中文池是**依原始索引順序**的,而 `starname-random.tsv` / `shipname.tsv` 是
    英文→中文。反查(中文→英文)的結果:**星名 0 歧義、艦名 0 歧義 0 缺漏**——
    每一條中文都只對應一個英文,所以還原是唯一的。

    ### 翻譯發生在「取名當下」,不是「顯示當下」

    這個選擇有具體後果:**中文模式的輸出與這一輪之前逐位元相同**(畫廊 34 張 0 張不同,
    含狀態指紋)。改成顯示時翻的話,存檔會存英文,指紋會變,而那個變動沒有遊戲意義。

    另一個理由是下游:星名/艦名會被玩家改名、會進戰報字串、會進網路封包,
    那些地方沒有一個知道「這個字串是不是可翻的專有名詞」。存最終字串最單純。

    ### 注入點

    `internal/shell` 不 import i18n:`SetNameTranslator` 由 `cmd/moo2` 在載入譯表之後
    注入一次。**英文模式不裝**(nil = 恆等)——不是「裝一個空的」,那樣譯表裡剛好有的 key
    會把英文換成中文,而且會隨譯表內容漂移。

    ⚠ 翻譯器只吃**專有名詞那兩張表**,不吃一般 UI 譯表。混在一起的話像 `Wolf`(艦名)
    這種與 UI 詞彙撞字的 key 會被別張表搶走。

    驗收:中文畫廊 34 張逐位元**全部相同**;英文畫廊的星圖確認顯示 Ching / Sabazius /
    Ogka / Pollux / Mizar 等英文星名。

    ### 還沒接完的(下一輪)

    英文畫廊的艦隊畫面顯示初始艦隊仍是中文(`拓荒號` / `先驅一號` / `殖民船` / `偵察艦`)
    ——那些是**開局艦隊的硬編名字**與**艦級名**,不在名稱池裡,屬剩下那 ~970 條。

85. **英文版面撞牆:每修一層就冒出下一層,而每一層只有截圖看得到**(2026-08-08)。

    使用者:「文字排版也要注意,必要時放大畫面,添入合適的大小的文字。」

    ### 先說沒做的:2× 畫布

    `rulebook/81` 的處方是「拉高內部畫布、別縮字」。但那條規則的前提是**低解析**
    (320×200),而 remake 的畫布已經是 **640×480 = 原版美術的原生解析**。
    真要再放大是 **1280×960 的 2× 邏輯畫布**(美術 nearest 放大、文字 2× 字級)。
    這一輪沒做,**沒做,不假裝做了**;下一輪做完了,見第 86 項(hi-res 畫布)。

    ### 做的:把「英文比中文長」造成的破版一層層挖出來

    艦艇設計畫面,每修好一層就露出下一層——**每一層都只有跑截圖才看得到**:

    | 層 | 症狀 | 修法 |
    |---|---|---|
    | 1 | 「已解鎖…」被畫布右緣切掉 | 依實際量測折行(`fnt.Wrap`),不是把字改短 |
    | 2 | 六列艦體空間壓到面板分隔線 | 改 3 列 × 2 欄 |
    | 3 | 英文模式分隔符 `▸` 是豆腐框 | 中文走點陣字有這個字、**英文走純向量字沒有**;改 ASCII 冒號 |
    | 4 | 英文模式元件名還是中文 | `Component.Name` 同時是查表 key,不能換成英文 → 從 `UnlockTech` 推導原版科技名 |
    | 5 | 英文模式武器改造標籤還是中文 | `weapon_mods.go` 只有中文那一份,補英文 |
    | 6 | 補完英文之後改造標籤**疊在一起** | 4 欄 × 76px 放不下 `No Range Dissipation (NR)`;改 2 欄 × 4 列,座標與熱區共用一個函式 |

    第 6 層還順手修掉一個**兩份寫死座標**的漂移風險:改造晶片的熱區與繪製先前各有一份
    座標表,現在共用 `designModChipRect`。

    ### 兩個測試抓到的真問題

    - **死光的 `UnlockTech` 是 0、主題掛在人造生命**——兩個都不對(執行檔武器表給
      `TECH_DEATH_RAY`=47,手冊把它放在 Xenon Technologies)。是英文名推導的測試抓到的:
      推導不出英文名 = `UnlockTech` 有問題。
      ⚠ 仍未解:手冊說 Xenon Technologies **不能靠研究拿到**,而 remake 把它當一般主題
      ——那是既有偏差,列進待辦。
    - **棘輪偵測器把 map key 算成缺口**。修的是偵測器不是棘輪(「只能往下調」):
      map 字面值的 key 是查表用的,不會被畫出來。陣列/結構的元素仍然要數。

    ### 這一輪最該記住的

    **英文模式的破版,中文截圖一張都看不到。** 六層裡有四層是英文專屬,而且是「補完翻譯
    之後才出現」——第 5 層修好直接造成第 6 層。所以驗收一定要**雙語各跑一次畫廊**,
    不能只看中文那一份逐位元相同就收工。

    驗收:中文畫廊 34 張逐位元(基準已更新那一張)、英文畫廊逐張看過。測試全綠。

86. **hi-res 畫布(2×):420 個呼叫點一個都沒改,靠「錄→放」換來的**(2026-08-08)。

    使用者:「go」(承接第 85 項提的 2× 畫布提案)。

    ### 為什麼不是改 420 個呼叫點

    第 85 項估過:直接把畫布拉到 1280×960,要動 **420 個繪製呼叫點**加命中區映射。
    那個估計沒有錯,錯的是**把它當成唯一的做法**。

    實際做法是把「文字」與「美術」拆成兩層:

    ```
    ① 畫面照舊畫進 640×480 離屏 —— 一行呼叫端都不用改
    ② 過程中的文字繪製被記錄下來而不畫(internal/uifont/record.go)
    ③ 離屏 nearest 放大 2× 貼到真畫布 —— 美術是銳利的整數倍放大
    ④ 記錄的文字用 2× 座標、2× 字級重畫 —— 在最終解析度重新柵格化
    ```

    改動只有三處:`uifont.Font.Draw` 分岔成「錄」或「畫」、`interactiveApp.drawScene`
    做合成、`pollInput` 把游標座標除回 640 空間。**所有畫面的座標仍然是 640×480 邏輯座標。**

    效果:CJK 從 10–13px 變成 20–26px,而且因為混合字型的門檻是 18,**2× 之後全部走向量**
    ——不是把點陣字放大,是真的用大字級重新排版。美術一個像素都沒糊。

    ### 三個只有量過才會知道的細節

    | 量測 | 數字 | 處置 |
    |---|---|---|
    | 墨水高度 1× → 2× | 11px → 22px(剛好 2 倍) | 不用補 |
    | 墨水**上緣** | 從 y+2 漂到 y+4(邏輯) | `Replay` 對齊**行框中心**而非左上角 |
    | 墨水寬度 | 2× 反而**窄 8%**(向量比點陣緊) | 不補;補了會把依 1× 折過行的段落再擠回去 |

    第二項不補的話,原版面板內高只有 96px 要塞四列,漂 2px 就足以讓末列被下邊框切掉。

    ### 代價:z 序 —— 而且第一次跑就撞上了

    錄下來的字一律最後重播,所以「先畫字、再用不透明面板蓋住字」會失效。
    第一次跑 2× 畫廊,指揮點數視窗上就浮著「藤斯塔」「巴哈姆」兩個星名,
    確認框裡浮著三個。**remake 到處都是這種寫法,不是少數例外。**

    修法不是逐處插屏障(要審 70 個呼叫點,漏一個就是看不見的回歸,而且以後每加一個面板
    都可能再漏),而是**把「畫面板」這個動作本身變成屏障**:cmd/moo2 裡所有
    `vector.DrawFilledRect` 與 `dst.DrawImage` 統一改走 `fillPanel` / `drawPanelImage`
    (`cmd/moo2/zorder.go`),它們會先把已錄的文字沖出去。

    沖的時候**必須把離屏清空**,不能重貼整張:離屏背景是不透明的,重貼會把上一輪剛畫好的字
    整片洗掉。清空之後每一輪只帶「這一輪新畫的美術」,source-over 的結合律保證疊出來與
    直接畫在同一張圖上一致。

    屏障還帶矩形相交判定(`uifont.BarrierRect`):不落在任何已錄文字上的美術直接略過,
    一幀幾十次填色不會變成幾十次全畫面重貼。`uiScale==1` 時 hook 是 nil,整條零成本
    ——1× 逐位元回歸驗證過(除了本輪刻意改的 6 張版面,其餘完全相同)。

    ### 順手挖出來的四個既有版面缺陷(1× 就有,2× 才看得見)

    | 畫面 | 缺陷 | 真值來源 |
    |---|---|---|
    | 星圖底部工具列 | 擦底板只蓋 445..455,烘進的 `COLONIES`/`PLANETS`/… 上緣在 440 → **每格都露出上面 5 列** | 英文模式跑同一張畫廊圖再掃亮字 |
    | 殖民地總覽排序列 | 整排偏右,每個中文標籤左邊掛一小截英文;`RETURN` 的板在 585..631,字其實在 552..594 → **完全沒蓋到** | 同上;七格改成互相重疊,連續無縫 |
    | 戰術控制列 | 右欄整欄偏右 9px、列偏下 3~5px | 英文模式量浮雕亮邊:左欄 x 274..327、右欄 338..391 |
    | 艦艇設計 | 元件四列 `69+24i` 末列畫到 153、面板內高只到 149;`國庫` 畫在框外 | 掃背景圖邊框 |

    **共同點:全部要用「英文模式那張圖」才量得到。** 用蓋著英文的中文截圖去量英文的位置,
    本來就量不準——這是第 85 項「雙語各跑一次畫廊」的延伸:英文那份不只用來看破版,
    **它就是版面座標的一手來源**。

    ### 截圖存檔的 alpha:一個看了很久的「白噪點 bug」其實不存在

    殖民地總覽下排第三格,在**全部 34 張基準圖裡**都是一塊白底黑點的噪點方塊,
    看起來像 LBX 解碼壞掉。查下去:那一區 `alpha=0`,而 `saveScreenshot` 把帶 alpha 的
    RGBA 直接存成 PNG,每個檢視器都會把它疊到白底。玩家螢幕上看到的是**黑**。

    改成存檔前壓到不透明黑(預乘 alpha 保留 RGB 即可)之後,那格顯示的是原版底圖的**星空**。
    外交議事廳的背景也一起變乾淨了。

    **截圖是這個專案唯一的驗收管道,截圖說謊比畫面有洞更貴。**

    ### 順手補的:中文折行的避頭尾

    種族說明「外交手腕高明,雇用領袖較廉(民主政體)」折成「…較廉(」+「民主政體)」
    ——開括號孤零零留在行尾。`uifont.applyKinsoku` 補上兩條規則:開括號不留行尾(往下推,
    一定安全)、收尾標點不在行首(把上一行末字推下來墊著,**且只在推完仍放得下時才做**)。
    兩條規則都配了正對照:拿掉 `applyKinsoku` 兩個測試都轉紅。

87. **戰術控制列:七顆鈕翻譯做完了,熱區一個都沒有**(2026-08-08)。

    第 86 項跑 2× 畫廊時發現的。`barButtonsCHT` 那張表把七顆鈕的中文標籤與座標寫得很完整,
    `drawBarLabelsCHT` 也把中文疊上去了——**但 `tacticalScreen.update` 裡從來沒有對應的判定**。
    畫面上看起來能點,點下去什麼都不會發生。

    ### 為什麼撐了這麼多輪沒被發現

    「按鈕長得對」與「按鈕能按」是兩個不同的問題,而**截圖只能證明前者**。
    先前每一輪都在看 `16_tactical.png`,每一輪看到的都是七顆中文按鈕,結論都是「控制列已中文化」
    ——那個結論本身沒錯,只是回答的不是「這個畫面能不能操作」。

    這與第 72 項(元件表有≠效果有接)是同一個形狀:**盤點時問的問題決定了會不會發現缺口**。
    差別是那次的證據是元件表,這次是截圖。

    ### 七顆各自接到什麼

    | 鈕 | 處置 |
    |---|---|
    | 自動 | 用**同一套格子規則**重複 `fireRound` 打完。⚠ 刻意**不走快速結算**:兩條路的公式細節不同,自動的結果若與手動不一致,玩家會學到「按自動比較划算」——那是規則漏洞不是便利功能 |
    | 掃描 | 手冊「Scan gives you information about an enemy ship」→ 切模式,點敵艦看資料 |
    | 登船 | 接第 80 項(登艦戰)。解算放 `shell.GameSession.ShipBoardingAttack`,畫面層只做選取/距離/戰報 |
    | 撤退 | 倖存艦離場、判定未勝(走既有的 over/won → `ApplyCombatOutcome`) |
    | ~~等待 / 完成~~ → **已完成**(2026-08-09) | `tacticalScreen` 建立逐艦行動佇列；WAIT 將未行動艦移到佇列尾端，DONE 結束目前艦，最後一艦才結算戰機與敵方回擊；手動開火只作用於選中艦 |
    | 選項 | ❌ 原版開的是設定畫面,remake 還沒有那個畫面 |

    原本的「後三顆點下去會說明自己為什麼沒有反應」只剩選項鈕成立；等待/完成已改為真正改變逐艦行動狀態。

2026-08-09 接線：`tacticalScreen` 新增 `acted` / `waited` 行動表。手動開火改走
`fireSelectedShip`，非最後一艦只結算該艦攻擊；最後一艦或 DONE 才進入回合交界，執行戰機、敵方回擊、充能、狀態與移動力重置。
`cmd/moo2/tacticalturn_test.go` 以正對照釘住 WAIT 佇列、DONE 回合交界與「未選艦不得跟著開火」。

    ### 守這個缺口只能靠測試

    截圖證明不了「能按」,所以補了兩條:七顆鈕的正中央都要命中自己、熱區互不重疊。
    正對照:把 `barButtonHit` 改成恆回 -1,測試立刻轉紅。

    ### 登艦戰力的近似,標清楚

    攻方吃玩家的陸戰隊戰力(種族特性 + 領袖 + 動力裝甲);**守方用同一個基礎值**
    ——敵艦(`genEnemyFleet`)沒有種族/領袖/科技資料,與 `Mods`/`HardShield` 一律留零值
    是同款既有簡化。守方唯一的差異來自保安站的 +20(手冊逐字)。

88. **英文模式的引擎層殘量:開局艦隊名與支援艦艦級**(2026-08-08)。

    第 84 項(名稱池雙語化)把 1,501 條星名/艦名改成「存英文原文、取名當下翻」,但
    留了一句「英文畫廊的艦隊畫面顯示初始艦隊仍是中文」。這一輪收掉。

    | 缺口 | 修法 |
    |---|---|
    | 開局三艘船的名字硬編中文(`拓荒號`/`先驅一號`/`先驅二號`) | `homeworldShips(tr)` 改存英文原文,譯表補在 `shipname.tsv` 尾端**獨立標記**——那三個是 remake 自訂的(原版從艦名池抽,而這裡沒有 rng),混進上面 535 條會讓「哪些是原版的」查不出來 |
    | 支援艦艦級沒有英文(`殖民船`/`偵察艦`) | 六個戰鬥艦體在 `dsHullOrder` 裡,支援艦不在。補 `supportClassEN`,經 `shipClassLabel` 顯示。⚠ 左邊那一欄仍是**查表 key**(`ShipCost`/`isSupportShipClass` 拿它比對),不能換成英文 |

    ### 驗收

    中文畫廊 **34 張逐位元不變**(翻譯器把英文原文翻回同樣的中文,含狀態指紋);
    英文畫廊的艦隊清單真的變成 `Pathfinder / Vanguard I / Vanguard II` + `Colony Ship / Scout`。

    測試守的是**接線**而不是字串:`homeworldShips(nil)` 要拿到英文原文,接上翻譯器要變中文。
    只驗前者的話,「存了英文但沒人翻」也會過。

89. **英文引擎層：威脅與持續事件播報**(2026-08-09)。

    第 88 項只收掉艦隊名與支援艦艦級,但回合摘要仍有幾條由 `internal/shell` 直接組出的
    中文敘述。這一輪沒有把查表 key 全域替換,而是沿報告資料結構補雙語欄位，避免
    `special_device_map`、`weapon_damage`、`shipspace` 這類內部 key 被翻壞。

    - `AIRaidReport.MessageEN`：AI 突襲的擊退／突破結果保留 AI 種族英文名與星名，
      並由回合摘要與 INFO 畫面在英文模式選用。
    - `GameSession.LastAntaresEN`：安塔蘭突襲警報補英文模板，熱座席位快照同步保存顯示暫態。
    - `LastPersistentEventEN`：超新星倒數／解除／爆發、時空異象消散、超空間獸航道訊息
      補英文進度；事件 19–28 的怪獸／持續事件初始報告不再只顯示 `A ... event has been reported.`。

    驗證：`TestAIRaid*`、`TestAntares*`、`TestPersistentEventReportsEnglishProgress`，以及
    全套 Docker + Xvfb `go test ./...`；中文模板仍保留原欄位，不改動中文畫廊路徑。

90. **1.31／1.5 資產路徑與主選單版本切換接線**(2026-08-09)。

    原本 `GameVersion` 只影響 `RuleProfile`；`assets.Resolver` 雖然支援多層覆蓋，卻沒有
    依主選單版本選不同的資料根。這會讓畫面標籤顯示 1.5、規則套 1.5，實際仍讀 1.31
    的 LBX，屬於靜默混版。

    `cmd/moo2/versionassets.go` 新增兩版路徑描述與 `auto` 判定：`-data13`、`-data15`
    各自可用逗號串 `patch,base`，未指定的欄位才回退共用 `-data`；沒有另一版專用路徑時
    不會把已指定的 1.31 目錄偽裝成 1.5。`auto` 會讀資料目錄的 `README.TXT` 版本標記，
    找不到標記才沿用 1.5 預設。`sceneBuilder.selectGameVersion` 在主選單切換時重建
    `assets.Resolver`，並清除會攜帶舊版 LBX 的快取；缺少資產時留在原版本。

    私有驗證輸入：`/home/anr2/moo2-private-build/gamedata/mastori2/README.TXT` 明載
    `Version 1.31`，`ORION95.EXE` 內也有同一版本字串；1.5 輸入為工作區外掛入的
    `moo2_patch1.5/MOO2-1.50.26.zip`，只讀掛載並在一次性 Docker `/tmp` 解出，未進 repo。
    `versionassets_test.go` 驗證共用回退、禁止跨版本偽裝、README auto 判定，以及切換後
    解析器真的讀到另一版本目錄。雙版本各跑 35 張 Docker + Xvfb 畫廊；第一次回歸發現
    1.5 `NEWGAME.LBX` 資產數由 30 增為 33，滿版背景由 #28 順延到 #31，已由
    `newGameBackgroundAsset` 依 `GameVersion` 選取並以測試釘住。修正後兩版 `NEW GAME`
    與後續種族選擇畫面皆完整可見。

91. **英文模式：殖民地行星／建築與歷史圖表資料顯示**(2026-08-09)。

    英文畫廊的殖民地管理畫面仍把 `Planet.Climate`、`Mineral`、`Size`、`Gravity` 與
    `ColonyBuildingNames` 的中文 key 直接畫出；INFO 歷史圖表也直接畫
    `HistoryMetricName` 與 `HistoryEmpireNames`。這些不是規則缺口，而是 UI 顯示層漏接雙語資料。

    `cmd/moo2/englishlabels.go` 新增集中轉換：新生成行星(`Gen>0`)讀既有 enum ID，舊存檔則對
    保留的中文字串反查 `climateNames`／`mineralNames`／`planetSizeNames`／`gravityNames`；
    建築與 Special action 分別查 `gamedata.BuildingByNameZH`、`SpecialActionByNameZH` 的
    `NameEN`。繁中路徑仍沿用原字串與 `、` 排版，不改中文畫面；未知舊值不猜翻譯。
    歷史圖表英文指標使用 `Population`／`Treasury`／`Fleet Strength`，玩家圖例為 `You`，
    AI 優先讀已保存的 `RaceIndex`，舊存檔才以名稱反查種族英文名。

    `englishlabels_test.go` 釘住新生成行星 ID、特殊物產、建築／Special action、歷史指標、
    新舊 AI 圖例，以及繁中不被改寫。完整 Docker + Xvfb `go test ./...` 與建置通過；目前
    最新程式以真正 1.31 資料跑英文畫廊 35 張、以 1.5 ZIP 的 8 個覆蓋 LBX 跑英文畫廊 35 張，
    關鍵殖民地／歷史畫面已目視檢查。

92. **英文模式：NEW GAME 設定值列**(2026-08-09)。

    1.5 實體資產英文畫廊檢查時，`NEW GAME` 的五個自繪值列仍直接取
    `shell.Difficulties`、`GalaxySizes`、`GalaxyAges`、`TechLevels` 的中文 `Name`，
    所以會出現「普通／中型／一般」；帝國數量本來已由 `b.tr` 正確顯示。

    `englishlabels.go` 新增難度(Tutor/Easy/Average/Hard/Impossible)、星系大小、年齡、
    起始科技等級的英文對照；設定索引與 shell 中文資料不變，繁中仍取原 `Name`。測試釘住
    兩種語言的預設值與繁中不變。完整 Docker + Xvfb 測試與建置通過；1.31／1.5 英文畫廊各
    35 張，另跑 1.31 繁中畫廊 35 張，值列已分別目視確認。這個修正與 91 項同樣只改顯示層，
    不改遊戲規則。

93. **英文模式：殖民地總覽與行星列表窄欄摘要**(2026-08-09)。

    1.5 英文畫廊的 `09_colonysummary.png` 仍顯示中文 `貿易品`／`Built` 清單；
    `12_planets.png` 的氣候、重力、礦產、大小與特殊物產列也仍直接露出中文。前一輪
    已處理共用環境列，本輪把殖民地總覽的目前建造／已建項目與行星列表接到同一組顯示層
    轉換；舊存檔未知值仍保留原文，不猜翻譯。星名下方的「同系還有 N 個天體」另由
    `PlanetsAt` 的數量產生英文窄欄 `N more`，不解析 shell 的中文短字串。

    `englishlabels_test.go` 覆蓋建造項目與多天體摘要的中英對照及繁中不變。最後一輪
    Docker + Xvfb `go test -buildvcs=false ./...`、建置通過；修正後以 1.5 patch 的 8 個
    LBX 覆蓋檔跑完整 35 張英文畫廊，`12_planets.png` 已目視確認 `Ogka 1 more`／
    `Joseki 3 more` 不再被截成半句，也沒有中文殘留。

94. **英文模式：外交、戰術與熱座直接顯示名稱**(2026-08-09)。

    1.31／1.5 英文畫廊檢查又抓到三處規則層中文 key 被自繪畫面直接畫出：外交使節與
    三顆提議鈕、戰術敵艦／戰機型別與熱座交接的 AI 接管席位。規則與存檔資料不改；
    `aiEmpireEnglishName`、`enemyDisplayName`、`combatShipLabel`、`fighterKindLabel`、
    `hotseatNameLabel` 集中負責最後一刻轉換，並順手讓種族關係資訊面板的 AI 圖例使用
    同一條顯示路徑。

    `EnglishDisplayLabels` 測試鎖定 Psilons／Sakkra Ship、Interceptor、Bulrathi 席位與
    繁中原樣回傳。1.31、1.5 英文畫廊及 1.31 繁中畫廊各 35 張已成功產生；英文
    `15_diplomacy.png`、`16_tactical.png`、`24_hotseat.png` 與繁中外交畫面已目視確認。
這只收掉實際畫廊抓出的顯示層缺口，不宣稱 `internal/` 所有動態敘述已完成英文化。

## 95. 客製種族特殊能力資料鏈（2026-08-09）

`cmd/moo2/customrace.go` 原本只把生產／成長／艦攻等數值聚合進 `shell.Race`；特殊能力
雖然能在畫面上勾選，開局後沒有任何持久來源，因此連已經存在於引擎的低／高重力、穴居、
戰爭領主、跨維度、魅力、惹人厭、寬容、貿易奇才公式都不會看到客製選項。

新增 `GameSession.CustomRaceTraits` 位元遮罩，`ApplyCustomRaceBonuses` 接收客製畫面選到的
`gamedata.RaceTrait`，`raceHasTrait` 在 `RaceIndex=-1` 時讀取該遮罩；同一欄位已加入
`sessionSnapshot` 與熱座 `seat` 的存取。舊存檔缺欄位時零值代表「沒有客製特殊能力」，不會
把客製種族誤認成阿爾卡里或人類。

這輪只把**已有消費端與公式**的能力標為生效；大型／富礦／貧礦母星、水生、幸運等選項當時只保存語意，
沒有假造尚不存在的星球、科技或事件模型。創造力／缺乏創造力已由第 100 項補上研究完成規則。對照測試
`TestCustomRaceTraitsReachExistingRulesAndSaveLoad` 與 `TestSeatRoundTripKeepsEveryField`。

## 96. 客製種族艦防／地面戰／諜報 picks（2026-08-09）

官方 `custom-race-picks.md` 主表本來就有 Ship Defense（−20／+25／+50）、Ground Combat
（−10／+10／+20）與 Spying（−10／+10／+20）三組數值 picks，但 `defaultPickCats` 只提供
艦艇攻擊；因此不是公式缺證，而是畫面與 `Race` 欄位漏接。

補上三組互斥循環選項，並讓四組 combat 類別依名稱分別寫入 `RaceCombatPct`、`RaceShipDefPct`、
`RaceGroundBonus`、`RaceSpyBonus`。`TestCustomRaceNumericCombatPickCategoriesReachSeparateRaceFields`
以官方中間正向檔鎖定四欄不串線；畫面總類別數與特殊能力清單也由
`TestCustomRaceSpecialsCarryTraitsAndFitLogicalCanvas` 保護。

## 97. 客製種族官方特殊能力清單與半機械化修復（2026-08-09）

`custom-race-picks.md` 的官方主表列出 22 項特殊能力；客製畫面原先只列 16 項，遺漏遺物母星、
半機械化、食岩、心靈感應、全知與匿蹤艦。畫面改成兩欄完整提供 22 項，並新增半機械化／食岩
互斥組；所有選項仍透過 `CustomRaceTraits` 保存，不把尚未建模的深層效果誤標為完成。

手冊 p.25 明確寫出半機械化種族「after any combat, they repair their ships completely」。
現有艦艇模型已有戰後修復入口，因此 `repairAfterBattle` 會讓半機械化種族在勝敗兩種結果後
完全修復；逐系統 10%／5% 回合修復仍需要獨立的內部損傷模型。`TestCyberneticRaceFullyRepairsAfterAnyBattle`
鎖定戰敗後也修復的正對照。

## 98. 食岩族殖民地食物消耗（2026-08-09）

手冊與既有遊戲考據明確區分三種食物消耗：一般人口每回合 1、食岩族 0、半機械族 0.5。
本輪把可在現有整數食物帳本中無歧義表達的食岩族規則接入 `engine.ColonyState.Lithovore`：
`RunColonyTurn` 不再扣人口食物、不標饑荒；AI 會把誤分配的農夫轉回工人；玩家與 AI 的新殖民地、
回合種族同步與 remake JSON 存檔均保留這個旗標。`TestRunColonyTurnLithovoreNeedsNoFood`、
`TestEndTurnSyncsLithovoreIntoPlayerAndAIColonies` 與存讀檔測試覆蓋此垂直切片。

半機械族的 0.5 食物／0.5 生產消耗已在後續垂直切片接入：`ColonyOutput` 新增半單位精確帳本，
帝國稅收／餘糧／貿易品收入、建造進度與 ETA 讀取精確值，舊整數欄位與 JSON 存檔保持相容。食物
複製機仍只換完整食物單位；半 BC 付款規則沒有原版證據，刻意保留為開放問題。

## 101. 半機械族半單位食物／生產帳本（2026-08-09）

手冊 help.tsv line 459 明確寫出半機械族每人口消耗半食物、以半生產單位補足；openorion2 的
`food_consumption_*` 與 `industry_consumption_*` 欄位也明確標示為 half-units。`engine.ColonyState`
新增 `Cybernetic`，`RunColonyTurn` 產生 `FoodHalf`／`FoodConsumedHalf`／`FoodSurplusHalf`／
`IndustryConsumedHalf`／`NetIndustryHalf`；整數欄位保留顯示與舊資料相容用途。

帝國收入改讀半單位換算，玩家與 AI 的種族同步會寫入旗標；玩家建造進度與 ETA 使用 `ProgressHalf`
累積奇數餘數，session snapshot 也保存新欄位。測試覆蓋奇數人口、收入、建造累積、玩家／AI 同步與
存讀檔。半機械生產消耗在污染清理與再生反應爐之後扣除，是依「獨立消費帳本」的強推論，尚非手冊
逐字證實；複製機的半 BC 付款規則仍未知。

## 99. 客製種族母星與環境人口規則（2026-08-09）

手冊 `help.tsv` 的 Cost[1]/[2]/[-1]/[3] 明確給出大型母星、富礦母星、貧礦母星與遺物母星
的效果；Cost[5]/[6]/[10] 也明確給出水生、穴居與環境耐受的食物／人口／污染規則。本輪新增
`race_homeworld.go` 作為單一對映層：

- 玩家套用內建或客製種族時，母星的大小、礦產、每科學家 +2 遺物研究與殖民地欄位同步更新。
- 新殖民地把水生 Tundra→Terran、Swamp→Ocean、Terran→Gaia 對映到食物與人口上限；環境耐受
  以 Terran 計人口上限且沿用免污染；穴居加 `+2×(size+1)` 最大人口。
- `ColonyState` 保存 `Aquatic`／`Subterranean` 旗標，玩家／AI 新殖民地與地形改造沿用同一組查表。

AI 母星仍保留既有「無完整種族環境初始化」的模型差異；本輪只把其新擴張殖民地與回合經濟接上，
避免改寫既有 AI 母星資料而破壞目前有證據的轟炸／經濟基準。

## 100. 創造力／缺乏創造力研究完成規則（2026-08-09）

手冊 `help.tsv` 明確寫出 Creative 會取得研究領域內全部應用，Uncreative 則在每個領域
隨機取得一項，且兩者互斥。既有研究引擎已經保留 `ResearchAll`、多選主題與
`ExplicitChoice` 的資料語意，本輪只在主題完成邊界接入種族規則：

- Creative 清除 `PendingChoice`，保留未明確抉擇的主題層級解鎖，因此該領域所有科技應用都能通過既有門檻。
- Uncreative 以獨立、可存檔的 `researchRand` 自動選一項並寫入 `ChosenTech`／`ExplicitChoice`，不再把玩家送進一般擇一畫面。
- 玩家與 AI 都走同一個純規則 helper；存檔保存亂數抽取次數，避免讀檔後重複或分岔研究選項。

覆蓋測試為 `TestCreativeResearchUnlocksEveryApplicationWithoutPendingChoice`、
`TestUncreativeResearchRandomlyChoosesOneApplication` 與完整 `internal/shell` 測試；
這不包含尚未建模的心靈感應、幸運、全知、匿蹤艦等其他種族深層能力。

## 102. 飛彈／魚雷改造垂直切片（2026-08-09）

手冊 `GAME_MANUAL.pdf` p.115-116 已給出 ECCM、EMG、MIRV、魚雷 ENV 與 OVR 的改造效果；
本輪不把只有資料、沒有消費端的 ARM/FST 列為完成。新增 `ResolveMissileShotWithMods`，
保留 `ResolveMissileShot` 舊入口作為無改造相容包裝：

- ECCM 透過既有 `MissileJamChance` 將干擾機率減半；EMG 先扣護盾，再以穿甲旗標直接傷害結構。
- MIRV 以四枚彈頭逐一套用干擾、匿蹤與位移判定；AMR 命中只減少一枚彈頭，不把結果粗暴乘四。
- 魚雷 ENV 使用既有四倍命中傷害管線，OVR 將彈頭強度乘 150%；`Ship.Mods` 已接入設計成本／
  佔格、快速結算、格子戰術與設計 UI，且按武器類型過濾歷史殘留代碼。

證據分級：ECCM/MV/OVR 為**已證實（手冊）**；EMG 的「直接結構」與現有「先扣盾後過甲」
資料管線相接為**強推論**；魚雷四分面視覺／容量仍是**模型近似**。ARM「摧毀所需傷害×2」
與 FST「提高飛彈 Beam Defense」分別缺少現行攔截器／飛彈防禦消費端，魚雷 NR 缺少射程衰減
模型，三者維持**未知／待接線**。測試為 `internal/gamedata/weapon_mods_test.go`、
`internal/shell/missile_mods_test.go`；Docker + Xvfb 完整測試及 `cmd/moo2` 建置通過。

## 103. 艦艇軍官由全域近似改為逐艦指派（2026-08-09）

原版存檔的 `Ship` 結構在 `internal/save/entities.go:436-473` 對應一個 `int16 Officer` 欄位；
`openorion2/src/gamestate.cpp:2365-2405` 的 `shipBeamOffense`／`shipBeamDefense` 也只有在
`sptr->officer >= 0` 時加上 Weaponry／Helmsman。先前 remake 只有 `GameSession.Leaders` 的
全帝國清單，因而出現「雇了艦艇軍官就每艘船都吃加成」的模型缺口；`ShipBeamAttackWithOfficer`
與 `ShipBeamDefenseWithOfficer` 雖有公式，沒有玩家可操作的資料來源。

本輪新增 `Ship.OfficerName`（空字串即未指派）與 `AssignOfficerToShip`／解除／查詢 helper：

- 艦隊畫面先選船，按 `LEADERS` 進軍官畫面；點艦艇軍官列可指派／改派，再點同一列解除。
- 改派先清除同一軍官在全帝國其他船的欄位；殖民地領袖與待雇傭兵不能提前上船。
- 快速結算與格子戰術的 Weaponry／Helmsman、星際航行 Navigator、戰後 Engineer 都讀逐艦
  指派；`Ship` 被拆到新艦隊時欄位隨船移動，熱座與 JSON snapshot 走既有保存路徑。
- 重製 JSON 使用名稱而非原版數字英雄 ID，這是明示的重製模型設計，不宣稱與 `.GAM` 位元格式
  相同；舊存檔缺欄位時視為未指派。原版 UI `Check_Officer_Fields_` 尚未完整反組譯，故畫面
  的列點選行為是可玩重製 UI，不把其座標語意升格成原版證據。

驗證：`internal/shell/officer_assignment_test.go` 固定檢查單艘生效、改派移除舊船效果、
殖民地領袖拒絕、JSON round-trip；並以 Navigator／Engineer／黑洞路徑回歸測試驗證既有消費端。

**勘誤（2026-08-09）**：上段的「使用名稱而非原版數字英雄 ID」是當時切片完成時的
快照，現在已由 `gamestate.cpp:1724-1725,1870-1871,2372-2401` 追回固定
`_leaders[0..66]` 索引鏈；`HERODATA`、`shell.Leader.ID` 與 `Ship.OfficerID` 已保存
該來源序號。名稱仍保留作舊 JSON 回退；這輪仍未擴張成原版 `.GAM` 全局匯入。
完整追溯見 `docs/re/officer-ids.md`。

## 104. 食物複製機半食物／半 BC（2026-08-11）

前文「複製機仍只換完整食物、半 BC 未知」是 2026-08-09 的歷史狀態，現已由程式
接線取代，但不把它誤寫成原版手冊逐字證實：

- `gamedata.FoodReplicatorConvertHalf` 以 half-food／half-industry 計算：每半食物
  花 2 半產能，完整食物仍等於手冊的 2 產能；因此 Cybernetic 奇數人口的半食物
  缺口不會先被 `/2` 捨掉。
- `ColonyOutput.FoodReplicatedHalf` 與 `FoodReplicatorCostHalfBC` 暴露精確值，
  `PlayerState.FoodReplicatorBCHalfRemainder` 跨回合保存；每兩個半 BC 才從帝國
  國庫扣 1 BC。這是依原版 half-unit 食物帳本與手冊 1 BC／完整食物的**強推論**，
  原版沒有直接給出碎單位付款時機。
- 回歸護欄為 `TestFoodReplicatorConvertHalfPreservesCyberneticHalfFood`、
  `TestFoodReplicatorsCoverCyberneticHalfFood` 與
  `TestRunEmpireTurnFoodReplicatorHalfBCCarries`。

## 105. 低優先視覺／runtime oracle 收斂（2026-08-11）

`COUNCIL.LBX#1` 的 640×480／10 幀與 `ANTAROOM.LBX#1` 的 640×480／55 幀已由
隔離 `lbxdump` 結構探針確認，議會與安塔蘭畫面現在都逐幀累積播放；逐像素／逐幀
對照契約與網路參考連結見 `docs/re/visual-oracle-20260811.md`。`CMBTSHP` 艦型
索引仍維持視覺近似，地面戰實機傷亡、事件漂移與爆炸連鎖不因靜態畫面通過而升格
為完成。
