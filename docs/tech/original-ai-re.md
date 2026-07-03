# 原版 MOO2 AI 逆向考據筆記(original AI 模式研究基礎)

## 0. 定位與鐵律

本文件目的是為 `internal/ai` 未來的 `OriginalDecider`(見 `docs/tech/ai-decision-modes.md` 第 3 節架構圖的
「待逆向原版 AI」節點)累積考據材料,供玩家在主選單於「remake AI」與「original AI」間切換時,
original 模式有真實依據可循,而不是 remake 啟發式的重貼標籤。

**與 `internal/ai` 現況的關係**:目前 `internal/ai/{diplomacy,economy,military,research}.go` 是**設計性重建**
(`ai-decision-modes.md` 明文標註「全部標【設計性重建,非原版】」),`ModeOriginal` 只是型別保留位、
`NewDecider` 對它 fallback 回 remake 邏輯。本文件**不修改任何程式碼**,單純整理逆向/考據結果,
讓未來實作 `OriginalDecider` 時有材料可用。

**鐵律(比照 `patch15-cfg-data-source.md` 的教訓)**:

1. **不臆造**。找不到來源就寫「無來源」,不用似是而非的社群傳言充數。
2. **classic 原值 vs 1.50 improved mod 值一律分清楚標記**。這是本專案已知的移植地雷——`patch1.5` 的
   `.CFG` 檔是「improved mod」配置,不是 classic 1.3/1.5 原版行為,兩者混用會讓 original AI 模式變成
   「四不像」。
3. **社群摘要不可信,只信 WebFetch 讀到的原文**;讀不到原文就標「間接引用」並壓低可信度。
4. **與既有文件的已知衝突,見第 6 節**——本輪研究推翻了 `community-mechanics-findings.md` 第 1 節與
   `ai-decision-modes.md` 第 5 節「AI 難度加成官方手冊未公開」的斷言,兩份文件尚未回頭修正(本輪任務
   範圍不含改動其他檔案),下一輪需要處理。

---

## 1. AI 性格系統(personality)

### 1.1 官方遊戲字串(高可信度,一手資料)

專案既有的字串抽取結果 `assets/i18n/estrings.tsv`(來自遊戲 `ESTRINGS.LBX`,非社群轉述)第 400-410 行
列出以下詞條,並各自標了原始分類:

| 英文 | 中文(專案既有翻譯) | 字串分類(estrings.tsv 原標) |
|---|---|---|
| Xenophobic | 排外 | 種族特質 |
| Ruthless | 冷酷無情 | 種族特質 |
| Aggressive | 好戰 | 種族特質 |
| Erratic | 反覆無常 | 種族特質 |
| Honorable | 重信譽 | 種族特質 |
| Pacifistic | 和平主義 | 種族特質 |
| Dishonored | 失信 | **外交狀態**(注意:不是「種族特質」) |
| Diplomat | 外交官 | 領袖職業 |
| Militarist | 軍國主義者 | 種族傾向 |
| Expansionist | 擴張主義者 | 種族傾向 |
| Technologist | 科技主義者 | 種族傾向 |
| Industrialist | 工業主義者 | 種族傾向 |
| Ecologist | 生態主義者 | 種族傾向 |

【來源】`/home/anr2/moo2/assets/i18n/estrings.tsv`(本專案既有的遊戲字串抽取檔)。【可信度:高】——
這是遊戲原始字串表,不是社群轉述。

**重要澄清(修正 AIRACES.CFG 註解的字面印象)**:`AIRACES.CFG` 的 `race_personality` 註解把 0-6 七個值
標成「0: xenophobic … 6: dishonored」,容易誤讀成「7 種同等級的性格」。但官方字串表把 6 種
(Xenophobic/Ruthless/Aggressive/Erratic/Honorable/Pacifistic)歸類為「種族特質」,唯獨 **Dishonored 被歸類為
「外交狀態」**——即 Dishonored 很可能不是遊戲開局隨機決定的固定性格,而是**外交行為觸發的狀態**
(例如背棄條約後被貼上「failed to honor a treaty」的標籤),`race_personality` 表把它借用成分布表第 7 個
可能值,只是這個表的欄位設計把「初始性格」與「外交後果狀態」共用同一組編碼(0-6)。【可信度:中】——
這是基於官方字串分類 + AIRACES.CFG 編碼交叉推論,尚未有原始碼等級證據證實兩者是否真的走同一個
`personality` 位元組(openorion2 只有存檔欄位,無邏輯,見 1.4 節)。

### 1.2 openorion2 的存檔欄位證據(高可信度,直接程式碼)

`openorion2/src/gamestate.h:1085`、`gamestate.cpp:1011,1080-1081` 顯示 `Player` 結構有獨立的
`uint8_t personality` 欄位,直接從存檔串流讀出(`personality = stream.readUint8();`),與另一個獨立欄位
`uint8_t objective`(`gamestate.h:1086`,`OBJECTIVE_HUMAN = 100` 作為「這是人類玩家、無 AI objective」的
哨兵值,見 `gamestate.h:56`)分開存放。

【來源】`openorion2/src/gamestate.h:1085-1086`、`gamestate.cpp:1011-1012,1080-1081,2539-2542`、
`galaxy.cpp:1432,1687,2103,2184`(僅用 `objective != OBJECTIVE_HUMAN` 判斷是否為 AI 玩家,無其他 AI 決策
邏輯)。【可信度:高,一手程式碼】——但這只證實「性格與 objective 各佔一個存檔位元組」,**不含任何
決策演算法**,與 `rules-implementation-audit.md` 「AI 決策邏輯零 RNG 來源」的既有結論一致,未推翻。

### 1.3 AIRACES.CFG:classic 種族性格分布表(中高可信度,官方 mod 工具內附 classic 對照值)

【來源】`patch1.5/MOO2-1.50.26.zip` 內 `AIRACES.CFG`(已解壓於本輪 scratchpad)。【可信度:中高】——
這是官方 1.50 改版工具內建的「mod 值 vs classic 值」逐行對照,不是社群逆向猜測;但仍是 **1.5 patch
附帶的資料,不是 1.3 原版獨立驗證過的數字**,兩者理論上應相同(性格分布屬於「種族設計常數」,1.3→1.5
通常不改這類數值),但本專案未能對 1.3 原始碼或存檔做交叉驗證,故列「中高」而非「高」。

**格式**(取自 `PARAMETERS.CFG` 第 826-844 行的官方說明):

```
race_personality <種族> <random0..random9> = 0(Xenophobic) | 1(Ruthless) | 2(Aggressive)
                                            | 3(Erratic) | 4(Honorable) | 5(Pacifist) | 6(Dishonored)
```

`random0`~`random9` 是 10 個等機率欄位(每格 10%),每格填入一個性格代碼(0-6)——這是用「10 格離散
抽樣」編碼機率分布的常見手法,不是 7 個固定檔位。**性格選擇公式**(`PARAMETERS.CFG:842-844`,原文照抄):

> To determine the column that personality is picked from, a random roll is subtracted by difficulty
> setting. Formula: `roll(10) + 1 - difficulty_byte (0-4)`

即:`column = Random(10) + 1 - difficulty_byte`,`difficulty_byte` 依「tutor easy average hard impossible」
文件慣用順序推論為 0-4(**此欄位對應關係是推論,CFG 未明講數字對到哪個難度名稱**,可信度中)。若
`column` 超出 1-10 範圍,合理推測會被夾限(clamp)到邊界,但 CFG 未明說夾限規則(**未驗證,標「待驗證」**)。

**意涵**:難度越高(`difficulty_byte` 越大),`column` 越小,越常落在陣列前段的性格值;難度越低,越常
落在陣列後段。由於陣列是「非遞減排序」,這代表**難度會系統性偏移 AI 的性格傾向**,但偏向哪個性格
（好戰或溫和)是**逐種族而定**,不是全域「難度越高越好戰」這種簡化說法(見下方各族分布,像 Klackons
陣列前段是 Xenophobic 排外、Psilons 陣列後段才是 Pacifist 和平主義,兩族的「難度效應方向」剛好相反)。
此選擇公式**沒有 `##`/`#` 標記**(它是說明性註解,不是可調數值),推論這描述的是遊戲既有機制本身,
非 1.50 mod 新增可調項,但仍**未在原版 EXE 或 1.3 存檔上驗證過**,可信度中。

**逐種族 classic 性格分布**(取 `##`/`#` 後的 classic 對照值;13 族中僅 **Humans、Trilarians** 的 mod 值與
classic 值不同,其餘 11 族 mod=classic):

| 種族 | classic random0-9 原始值 | 分布(性格:機率) |
|---|---|---|
| Alkari | 3 4 4 4 4 4 4 4 5 5 | Erratic 10% / Honorable 70% / Pacifist 20% |
| Bulrathi | 1 2 2 2 2 2 2 2 3 3 | Ruthless 10% / Aggressive 70% / Erratic 20% |
| Darloks | 1 1 0 2 2 2 2 2 2 2 | Xenophobic 10% / Ruthless 20% / Aggressive 70% |
| Elerians | 0 0 1 1 1 1 1 2 2 2 | Xenophobic 20% / Ruthless 50% / Aggressive 30% |
| Gnolams | 1 1 2 2 3 3 3 5 5 5 | Ruthless 20% / Aggressive 20% / Erratic 30% / Pacifist 30% |
| **Humans** | 3 4 4 4 4 4 4 4 5 5(**mod 改為** 3 3 3 4 4 4 4 4 5 5) | classic: Erratic 10% / Honorable 70% / Pacifist 20%(mod: Erratic 30% / Honorable 50% / Pacifist 20%) |
| Klackons | 0 0 0 0 0 0 0 1 2 2 | Xenophobic 70% / Ruthless 10% / Aggressive 20% |
| Meklars | 0 2 2 3 3 3 3 3 3 3 | Xenophobic 10% / Aggressive 20% / Erratic 70% |
| Mrrshan | 0 1 1 1 1 1 1 1 2 2 | Xenophobic 10% / Ruthless 70% / Aggressive 20% |
| Psilons | 3 3 4 5 5 5 5 5 5 5 | Erratic 20% / Honorable 10% / Pacifist 70% |
| Sakkra | 1 2 2 2 2 2 2 2 3 3 | Ruthless 10% / Aggressive 70% / Erratic 20% |
| Silicoids | 0 0 0 0 0 0 0 2 2 3 | Xenophobic 70% / Aggressive 20% / Erratic 10% |
| **Trilarians** | 3 4 4 4 4 4 5 5 5 5(**mod 改為** 3 3 4 4 4 5 5 5 5 5) | classic: Erratic 10% / Honorable 50% / Pacifist 40%(mod: Erratic 20% / Honorable 30% / Pacifist 50%) |

無種族的 classic 分布含 Dishonored(6)——與 1.1 節「Dishonored 是外交狀態、非開局性格」的推論一致
(至少在 classic 表中,沒有種族「天生」被指派成 Dishonored)。

【來源】`AIRACES.CFG` 第 11-23 行(逐行 `##`/`#` 消歧)。【可信度:中高,見上方限定】。

### 1.4 官方性格行為描述(社群一手引用,中可信度)

`MANUAL_150.html`(patch 1.50 說明書)與 `GAME_MANUAL.pdf`(1996 原版遊戲手冊,188 頁,已用 `pdftotext`
轉出純文字核對)**都沒有**用「Xenophobic/Ruthless/Aggressive/…」這組詞彙寫一段性格行為說明——`GAME_MANUAL.pdf`
只在描述「Report(情報回報)」功能時提到「這些回報包含對方種族性格的描述」(`agents assigned to the
selected race. These include a description of the race's personality`,manual 第 50 頁),但沒有把 7 種
性格的行為傾向列成文件裡的表格。換言之,**官方手冊本身不是這組性格行為描述的來源**,以下引用的是
社群 FAQ:

【來源】`https://challengetakers.proboards.com/thread/1412/master-orion-2-ai-faq`,作者 Onishiba,本輪已用
WebFetch 直接讀取原文(非摘要)。【可信度:中】——單一社群來源,但作者行文像是拆解過遊戲字串/實測
歸納,用詞與 §1.1 官方字串表吻合(未出現 Dishonored,與官方「Dishonored 非種族特質」的分類一致,構成
一個弱交叉驗證)。

7 種性格的行為傾向(原文譯,英文原句附註):

| 性格 | 行為傾向(原文照譯) |
|---|---|
| Ruthless(冷酷無情) | 「幾乎不需要挑釁就會開戰,且願意犧牲星艦」("attack with little or no provocation and will sacrifice starships") |
| Erratic(反覆無常) | 「難以預測,今年可能和平,明年就宣戰」("unpredictable. One year they may be peaceful, the next year they will go to war") |
| Aggressive(好戰) | 「一旦取得優勢局面就會隨時進攻」("will attack any time they reach a favorable position") |
| Pacifistic(和平主義) | 「極力維持和平關係,很少主動開戰」("eager to maintain peaceful relations. They rarely attack") |
| Honorable(重信譽) | 「不會攻擊關係良好的對象」("will not attack those they are on good terms with") |
| Xenophobic(排外) | 「不信任任何人,正面外交效果打對折」("distrust everyone, halving the effects of positive diplomacy") |
| Dishonored(失信) | 該篇未提及(與 §1.1 的「非種族特質」分類一致) |

**逐種族性格/objective 觀察**(同一來源,質性描述,無法逐一與 §1.3 的 classic 分布數字對齊,只能當
「玩家實測經驗」的佐證,**不可信度高於中**):Alkari(Pacifistic 或 Honorable)、Bulrathi(Aggressive
Expansionist,也可能 Erratic/Ruthless)、Darloks(Aggressive Militarists)、Elerians(Aggressive 且
Ruthless)、Gnolams(Aggressive Diplomats,也可能 Erratic/Industrialist/Technologist)、Humans(Pacifistic
或 Honorable;固定是 Diplomats)、Klackons(Ruthless,也可能 Xenophobic/Aggressive)、Meklars(Erratic,
常是 Technologists)、Mrrshan(Ruthless Militarists,也可能 Aggressive/Industrialists)、Psilons
(Pacifistic Technologists)、Sakkra(Aggressive,有時 Erratic)、Silicoids(Xenophobic Industrialists,
也可能 Erratic/Expansionist)、Trilarians(Pacifistic,有時 Honorable)。

**與 §1.3 classic 分布表的交叉檢驗**:多數吻合(如 Klackons 社群觀察「Ruthless/Xenophobic/Aggressive」對應
classic 表 Xenophobic 70%/Ruthless 10%/Aggressive 20%;Psilons 社群觀察「Pacifistic」對應 classic 表
Pacifist 70%),但非精確一致(社群觀察是玩家多局遊戲的印象,classic 表是機率分布,兩者本質不同,
不應期待逐字對上)。Humans 一項有落差:社群稱「固定是 Diplomats」(objective,非 personality),
與 §2 的 6 種 objective 系統相關,見下節。

---

## 2. AI 難度加成(本輪最重要發現:推翻既有文件「無來源」結論)

### 2.1 官方手冊「AI Opponents / Generic AI bonuses」表(高可信度,一手官方資料)

`MANUAL_150.html`(1.50 patch manual)「Modding → AI Opponents → Generic AI bonuses」一節有**逐難度的精確
加成表**,本輪用 Python 直接解析原始 `<table>` HTML(逐 `<td>` 取值,避開純文字 strip 造成的數字黏連
誤判)萃取,取得完整 9 欄 × 5 難度:

| 難度 | Growth% | Food | Prod | Res | BC | Command Deficit BC | Spy Bonus | Troops & Marines | Antaran Marines |
|---|---|---|---|---|---|---|---|---|---|
| Tutor | 0 | -1/4 | -1/2 | -1/2 | 0 | 12 | -2 | -2 | -4 |
| Easy | +1 | 0 | 0 | 0 | 0 | 11 | -1 | -1 | -2 |
| Avg | +2 | 1/4 | 1/2 | 1/2 | 1/4 | 10 | 0 | 0 | 0 |
| Hard | +3 | 1/2 | 1 | 1 | 2/4 | 9 | 1 | 1 | 2 |
| Imp | +4 | 1 | 2 | 2 | 3/4 | 8 | 2 | 2 | 4 |

【來源】`MANUAL_150.html`(`AI Opponents` 章節,原始 `<table class="c23">`)。【可信度:高】——官方 1.50
patch manual 直接給出的表格,不是社群逆向、不是社群摘要。手冊行文「Below table summarizes standard AI
bonuses per difficulty level」(標準 AI 加成)是對既有機制的陳述,不是 1.50 新增的加成本身——**這張表
應視為 classic 基準值**,1.50 新增的是「可以用 config 調整這些值」的能力,不是數字本身變了(手冊沒說
數字有變動)。可信度定為「高」但仍標註「未在 1.3 原版獨立驗證,理論上 classic 沿用」。

**與 `internal/ai/decider.go` 既有的 remake 設計值互相獨立**,不衝突(remake 用的是 `TreasuryLow/Hi=50/300`
這類自訂國庫門檻,和這裡的「每人口單位加成」是完全不同的機制)。

### 2.2 PARAMETERS.CFG:`ai_productivity_bonus` / `ai_income_bonus` 完整 5 檔位數值

手冊原文緊接著給出兩個 config key 的**全部 5 個難度值**(不是 `PARAMETERS.CFG` 檔案本身列出的——
該檔僅示範 `hard` 一行當語法範例,完整表格要看手冊):

```
ai_productivity_bonus tutor      = -10;
ai_productivity_bonus easy       =   0;
ai_productivity_bonus average    =  10;
ai_productivity_bonus hard       =  20;
ai_productivity_bonus impossible =  40;

ai_income_bonus tutor      = 0;
ai_income_bonus easy       = 0;
ai_income_bonus average    = 1;
ai_income_bonus hard       = 2;
ai_income_bonus impossible = 3;
```

`ai_productivity_bonus` 定義為「數值/20 = 每人口單位 Prod/Res 加成」;交叉驗算後可推出 **Food 加成 =
數值/40**(手冊沒有明講這個 /40,是本輪用 2.1 節的表逐列驗算出來的,5 個難度全部對得上,見下表):

| 難度 | productivity_bonus | /20(Prod=Res) | /40(Food) | 對照 2.1 節表格 |
|---|---|---|---|---|
| Tutor | -10 | -1/2 | -1/4 | Prod=-1/2 Res=-1/2 Food=-1/4 ✓ |
| Easy | 0 | 0 | 0 | 全 0 ✓ |
| Avg | 10 | 1/2 | 1/4 | Prod=1/2 Res=1/2 Food=1/4 ✓ |
| Hard | 20 | 1 | 1/2 | Prod=1 Res=1 Food=1/2 ✓ |
| Imp | 40 | 2 | 1 | Prod=2 Res=2 Food=1 ✓ |

`ai_income_bonus` 定義「數值/4 = 每人口 BC 加成」,同樣 5 檔全對上 2.1 節 BC 欄(0, 0, 1/4, 2/4, 3/4)。
【來源】`MANUAL_150.html` AI Opponents 章節原文 + 本輪推算驗證。【可信度:高】(表格數字為手冊一手
資料;/40 除數為本輪推算,但 5/5 全對,判定可信)。

**與 `PARAMETERS.CFG` 實際內容的關係**:`PARAMETERS.CFG`(3574 行)本身**只示範 `hard` 一行**
(`ai_productivity_bonus hard = 20;`、`ai_income_bonus hard = 2;`)當語法範例,不是完整表——這與
`patch15-cfg-data-source.md` 已知的 `.CFG` 檔案模式一致(多數段落只給一個代表值示範語法,而非窮舉
表格,`ai_race_variant_table Alkari hard v1 = 1 7 2;` 也是同樣的單行示範)。**完整 5 檔數值只在手冊正文
出現,CFG 檔案本身沒有**,這點糾正了原本以為「PARAMETERS.CFG grep 找不到就是沒有」的預期。

### 2.3 難度也影響「性格選擇」與「Antaran 攻擊機率」(手冊原文摘錄,高可信度)

手冊在 Generic AI bonuses 表後緊接的一句話(原文):「Note that difficulty level also influences other
aspects of the game, like the way AI conducts diplomacy and spying, AI's eagerness to attack, random
events, Antaran likelihood of picking a human player as victim (/10, /5, 0, *1.5, *3), etc.」

即 Tutor→Antaran 選中人類玩家當受害者的機率 ÷10,Easy→÷5,Avg→基準(×1),Hard→×1.5,Imp→×3。
這組倍率對應「Tutor Easy Avg Hard Imp」五檔,與 §1.3 的 `difficulty_byte(0-4)` 推論順序一致(再一次
交叉驗證了難度欄位的排序方向)。【來源】同 2.1。【可信度:高】。

### 2.4 與既有文件的差異(必須記錄,見第 6 節)

`community-mechanics-findings.md` 第 1 節結論「找不到可靠來源,仍無來源」——**本節找到的來源不是
社群數字,而是官方 1.50 patch manual 正文**,該文件當時搜尋的是 GameFAQs/StrategyWiki 等社群站台
(被擋),沒有去讀本專案已有的 `MANUAL_150.html`。`ai-decision-modes.md` 第 5 節「MOO2 的 AI 決策邏輯與
難度加成,官方手冊未公開規則」这句判斷,**至少對「難度加成」這半句已不成立**——AI 決策邏輯(性格如何
逐回合轉化成行動)仍然未公開,但難度加成的精確數字現在有官方一手來源。

---

## 3. AI 決策傾向(研究/造艦/外交/擴張)

### 3.1 AI Objective(6 種,獨立於 personality)

`MANUAL_150.html`「AI & Diplomacy」changelog 章節記載一個 1.50 修正的 bug(原文摘錄,高可信度一手資料):

> AI Objective Bug: At the start of a game, each AI will get one of 6 objectives assigned: Diplomat,
> Ecologist, Expansionist, Industrialist, Militarist or Technologist. The AI's race picks are an input
> for this determination. The function that is used for that purpose had incorrect values for Ship
> Defense and Ship Attack (SD+20, SD+40, SA+25) and as a result those traits were not taken into
> account. These values are now variable and depend on the actual settings for Custom Race (where
> classic is: SD+25, SD+50 and SA+20).

**考據重點**:

- objective 由「種族 picks」(自訂種族點數,含 Ship Defense/Ship Attack 等)當輸入算出——這是**唯一
  一處官方文件明講「AI 決策輸入來源」的地方**,證實 objective 選擇不是純隨機,是種族設計點數的函式。
- **classic(1.3/未修正 1.50 前)有一個 bug**:計算 objective 時,Ship Defense 應為 `SD+25`(第一級)、
  `SD+50`(第二級),Ship Attack 應為 `SA+20`,但 classic 程式碼誤用了 `SD+20`、`SD+40`、`SA+25` 三個
  錯位的數字,導致這三個特性**沒有真正被納入計算**(用錯常數去比對種族點數表,永遠比不中,等同該
  特性被忽略)。**這代表:若要重現 classic(含 1.3)的 objective 行為,必須刻意複製這個 bug**,不能用
  「修正後的正確版本」,否則行為會跟原版 1.3/1.5(未修正)不一致。1.50 patch 修正後的版本才是
  `SD+25/SD+50/SA+20` 三個正確值。
- 6 種 objective 与 §1.1 官方字串表的「種族傾向」分類(Militarist/Expansionist/Technologist/
  Industrialist/Ecologist,5 個)基本對應,唯獨 **Diplomat** 在 `estrings.tsv` 裡被歸類成「領袖職業」而非
  「種族傾向」——推論是同一個字串(`Diplomat`)在遊戲裡身兼兩種用途(AI 開局目標 + 領袖技能名稱),
  本專案既有的字串抽取工具按前後文分類時只標記了其中一種用途,**這是本專案既有 i18n 分類的已知盲點,
  不是新發現的遊戲機制**,列此供未來校正 `estrings.tsv` 分類參考(不在本文件範圍內修正)。
- `openorion2` 沒有實作這個 objective 計算函式(見 §1.2),只有存檔欄位,無法從程式碼交叉驗證上述 bug
  的具體判斷邏輯(如「種族點數表怎麼比對成 6 選 1」),**這部分演算法仍是未解的黑盒**。

【可信度:官方 bug 描述本身「高」(一手手冊原文);由此反推「classic 需要複製 bug 才能重現原版行為」
是本輪推論,「中高」。】

### 3.2 AISHIPS.CFG 造艦偏好(mod 值,非 classic——已用手冊比對證實)

`AISHIPS.CFG`(165 行)**沒有** `##`/`#` classic 對照標記(不同於 `AIRACES.CFG`),原本無法判斷是否為
classic 值。本輪用 `MANUAL_150.html`「AI Ships → Auto Design Ships → Table I」找到官方手冊親自列出的
**classic** `standard_beam_missile_x2` 完整表(6 個艦體大小),與目前 `AISHIPS.CFG` 內容逐行比對:

| 艦體 | classic(手冊原文) | 1.50 improved mod(AISHIPS.CFG) | 是否相同 |
|---|---|---|---|
| frigate | -1 -1 -1 15 -1 -1 -1 -1 -1 85 -1 0 | -1 20 -1 20 -1 -1 -1 -1 40 20 -1 0 | 否 |
| destroyer | 10 0 -1 5 -1 -1 -1 -1 45 30 -1 10 | -1 20 -1 20 -1 -1 -1 -1 40 20 -1 0 | 否 |
| cruiser | 15 0 -1 0 -1 -1 -1 10 30 30 -1 15 | -1 25 -1 25 -1 -1 -1 0 40 10 -1 0 | 否 |
| battleship | 15 0 -1 5 -1 -1 -1 15 25 25 -1 15 | -1 25 -1 25 -1 -1 -1 0 45 5 -1 0 | 否 |
| titan | 15 0 -1 10 -1 -1 -1 15 25 20 -1 15 | -1 25 -1 25 -1 -1 -1 -1 45 5 -1 0 | 否 |
| doomstar | 15 0 -1 15 -1 -1 -1 15 20 20 -1 15 | -1 25 -1 25 -1 -1 -1 -1 50 0 -1 0 | 否 |

**結論:`AISHIPS.CFG` 全部是 1.50 improved mod 調整過的值,不是 classic。** 這是本輪對
`patch15-cfg-data-source.md` 既有警告(「CFG 是 improved mod,不適合盲抄」)的具體實證——之前只是理論
上的警告,現在有逐行比對證據。**只有手冊親自列出的 Table I(`standard_beam_missile_x2`)這一組 6 行
數字**可信是 classic;`AISHIPS.CFG` 其餘 7 個表(hybrid_beam_carrier、hybrid_beam_missile_x5、
standard_missile_x10、standard_carrier、special_device、special_weapon1)以及 `ai_ship_design_*_theme`
系列(bio_weapons/capture/cloaking/beam_defense/missile_defense/armor/shield/beam_offense_specials/…)
**手冊沒有列出對應的 classic 版本**,無法比對,只能標「未知是否為 classic」。

【來源】`MANUAL_150.html` AI Ships 章節 + `AISHIPS.CFG`。【可信度:高(比對本身),中(其餘 7 表 classic
與否未知)】。

**造艦邏輯結構**(手冊原文說明,高可信度,不含具體 classic 數字,純規則說明):

- 每個難度/種族開局會分配 12 欄的「設計基準表」之一(依已知科技如 BHG/Stellar Converter 有機率切到
  `special_weapon1` 表,見 `ai_ship_design_bhg_chance`/`ai_ship_design_stellar_chance`)。
- 12 欄代表:Theme Special / Beam Special / Missile Special / Defense Special / Special Weapon /
  Special Device / Fighters / Missile(x2/x5/x10)or Torpedo / Heavy Beams / Beams / Point Defense
  Beams / Bombs,數值是「該艦體空間的百分比預算」,Auto-Design **由左到右**依序填裝,-1 代表忽略該欄。
  Classic 表每列總和 = 100(若 mod 表總和 > 100 會產生「作弊船」,超編裝備)。
- **Theme Special 是開局時依種族被指定的「7 選 N」風格**(Bio Weapons / Capture / Cloaking / Beam
  Defense / Missile Defense / Armor / Shield),每個風格底下有 5 個候選(武器或特殊裝置),用單一數值
  加權、由上到下依序嘗試裝上艦(受船體空間與已研發科技限制)。

### 3.3 1oom 架構模板(同開發商同世代 4X 的架構借鏡,非 MOO2 本身驗證資料)

`docs/kickoff/07-ai-strategy.md` 已記錄 1oom(MOO1 開源重製)`game_ai_classic.c` 的架構:可插拔 vtable
(`game_ai_classic`)+ 回合分兩階段(p1 戰略/艦隊調度、p2 艦艇自動設計)。**這是 MOO1 的驗證過重製碼,
不是 MOO2 本身的資料**,只能當「同開發商(Simtex)、同世代 4X 遊戲」的**架構借鏡**——MOO2 的 AI 介面
設計(如 `Decider` interface)可以參考這個分階段結構,但具體數值/演算法細節必須以本文件 §1-2 的
MOO2 一手資料為準,1oom 的常數不能直接套用到 MOO2。

### 3.4 社群觀察補充(僅列有來源者)

除 §1.4 已引用的 Onishiba FAQ 帖(涵蓋種族性格 + objective 的質性描述)外,本輪嘗試以新角度搜尋
「AI 研究順序偏好」「AI 造艦偏好」「AI 擴張策略」等主題(`site:lparchive.org`、`site:spheriumnorth.com`、
`site:gog.com` 限定搜尋),搜尋結果只回傳 StrategyWiki(仍被 Cloudflare 擋,WebFetch 回 403,與
`community-mechanics-findings.md` 已知結論一致,非新發現)、Steam 討論串(2016 重製版語境居多,混版
風險,未採用)、`masteroforion2.blogspot.com`(2005 攻略,標題泛用未見 AI 專項內容,未深入抓取)。
**沒有找到超出 `community-mechanics-findings.md` 既有範圍的新社群數據**——AI 研究選題順序、造艦決策的
量化細節,社群層級同樣沒有逆向出來,維持「無來源」。

---

## 4. 對 original AI 模式實作的建議

### 4.1 能忠實還原的部分

| 項目 | 依據 | 實作建議 |
|---|---|---|
| 13 族 classic 性格分布(§1.3) | AIRACES.CFG classic 值 | 開局用 `Random(10)+1-difficulty_byte`(§1.3 公式)幫每個 AI 玩家抽一個 personality(0-6),分布表直接照抄 |
| 5 檔難度的 Growth/Food/Prod/Res/BC/Command/Spy/Troops/Antaran 加成(§2.1-2.2) | 官方手冊一手表格 | 直接做成查表常數,`ai_productivity_bonus`/`ai_income_bonus` 的 /20、/40、/4 除數公式已驗證,可直接實作 |
| Antaran 選人類玩家當目標的難度倍率(§2.3) | 官方手冊原文 | ÷10/÷5/×1/×1.5/×3 五檔倍率可直接套用(若專案有 Antaran 事件系統) |
| classic `standard_beam_missile_x2` 造艦基準表(§3.2) | 官方手冊 Table I | 6 個艦體大小的 12 欄預算可直接照抄,當「AI 標準武裝艦」設計依據 |
| 7 種性格的定性行為傾向(§1.4) | 社群 FAQ(中可信度) | 可當 `Stance`/`BuildPriority` 決策的**方向性**參考(如 Ruthless→更早觸發 `StanceWar`),但門檻/權重仍需自行設計 |

### 4.2 只能近似或需自行設計的部分

- **逐回合決策演算法**(性格分數如何轉成「這回合要不要宣戰/求和/造船/擴張」的具體判斷式)——
  官方手冊、社群、`openorion2` 三方都沒有給出,只能仿照 1oom 的階段式架構(§3.3)自行設計,標【設計性
  重建】。
- **AI Objective 的種族點數判斷函式**(§3.1,如何用 SD/SA/GC 等點數算出 6 選 1)——手冊只給了
  「classic bug 用錯 3 個常數」這條線索,沒有給出完整判斷式,需自行設計或進一步逆向 EXE。
- **`AISHIPS.CFG` 其餘 7 個造艦基準表 + theme 加權表**(§3.2)——無 classic 對照,若要用只能標
  「疑似 1.50 mod 值,非 classic 驗證值」,或改為自行設計仿照 Table I 的風格外插。
- **性格選擇公式的 clamp 規則**(§1.3)——`column` 超出 1-10 範圍時如何處理,未驗證,需自行決定
  (夾限到邊界是最合理的猜測,但非官方證實)。
- **AI 研究選題、擴張路徑的量化細節**(§3.4)——維持 `community-mechanics-findings.md` 已有結論,
  無來源,只能靠 DOSBox 黑箱測試或沿用 remake 的設計性邏輯。

### 4.3 風險/待驗證清單

1. `difficulty_byte(0-4)` 是否真的照 tutor/easy/average/hard/impossible 順序編碼——目前是合理推論,
   未見官方明文數字對照,建議之後找到原版存檔或 debug 版 EXE 時驗證。
2. §1.3 性格選擇公式的隨機數是否會 clamp、如何 clamp——未驗證。
3. `AIRACES.CFG` 的 `ai_race_variant_table`(hard/impossible 難度種族變體)本身系統**確認是 classic 機制**
   (手冊「Unlock Extra AI Race Variants」明講「classic 遊戲內建 3 個 variant,程式碼裡其實有 5 個」),
   但**具體數值**——本輪只能靠逐行 `##`/`#` 判斷「這行有沒有被 1.50 mod 改過」,**CFG 檔沒有像
   `race_personality` 一樣附上 classic 對照值**,所以「有改過」的行完全無法還原 classic 原始數字,
   只有明確標註「no changes」的 Mrrshan 區塊(§AIRACES.CFG 第 112-120 行)可信是 classic。這點容易被
   誤讀成「AIRACES.CFG 全部都有 classic 對照」,特此強調澄清。
4. Objective bug 的「SD+20/SD+40/SA+25 vs 正確 SD+25/SD+50/SA+20」——只知道錯在哪三個常數,不知道
   這三個常數在完整判斷式裡怎麼跟 6 個 objective 的權重比較,無法重建完整演算法。
5. `AISHIPS.CFG` 除 Table I 外的其餘表格是否為 classic——未知,標「疑似 mod 值」。

---

## 5. 一手資料清單(供覆核)

- `patch1.5/MOO2-1.50.26.zip` 內 `AIRACES.CFG`(122 行)、`AISHIPS.CFG`(165 行)、`PARAMETERS.CFG`(3574 行,
  本輪 grep 關鍵字:`ai_`、`difficulty`、`bonus`、`personality`)。
- `moo2_patch1.5/MANUAL_150.html`(1.50 patch manual,已用 Python 去除 `<style>/<script>` 後轉純文字全文
  檢索,並對關鍵表格改用逐 `<td>` 解析避免文字黏連誤判)。
- `moo2_patch1.5/GAME_MANUAL.pdf`(1996 原版遊戲手冊,188 頁,`pdftotext` 全文檢索;
  `original_game/Master of Orion 2 - CD Manual.pdf` 只有 9 頁且無文字層,無法檢索,推測是掃描版附件
  非完整手冊)。
- `openorion2/src/gamestate.h`、`gamestate.cpp`、`galaxy.cpp`(grep `ai_|personality|difficulty|objective`)。
- `assets/i18n/estrings.tsv`(本專案既有字串抽取,含官方 `ESTRINGS.LBX` 字串分類)。
- `https://challengetakers.proboards.com/thread/1412/master-orion-2-ai-faq`(WebFetch 原文,作者 Onishiba)。
- `docs/kickoff/07-ai-strategy.md`、`docs/tech/community-mechanics-findings.md`、
  `docs/tech/patch15-cfg-data-source.md`、`docs/tech/ai-decision-modes.md`、
  `docs/tech/rules-implementation-audit.md`(既有專案文件,已讀過避免重工)。

---

## 6. 與既有文件的衝突(留供下一輪處理,本輪不修改其他檔案)

- `docs/tech/community-mechanics-findings.md` 第 1 節「AI 決策傾向 / 難度加成的具體數值」結論
  「找不到可靠來源,仍無來源」——**難度加成部分已被本文件 §2 推翻**(官方手冊有精確表格),
  該文件當時只搜了 GameFAQs/StrategyWiki/Steam 討論串等社群站台,沒有讀本專案已有的 `MANUAL_150.html`
  正文(可能是因為該文件的搜尋重點放在「社群逆向數字」,而非重新翻一次已下載的官方文件)。「AI 決策
  邏輯」(逐回合演算法)部分仍然無來源,這半句沒有被推翻。
- `docs/tech/ai-decision-modes.md` 第 5 節「MOO2 的 AI 決策邏輯與難度加成,官方手冊未公開規則」——
  同上,「難度加成」半句已不成立,「AI 決策邏輯」半句仍然成立。
- 建議下一輪:把上述兩處改成「難度加成有官方數值(見 `original-ai-re.md` §2),AI 決策邏輯本身仍未
  公開」,並在 `ai-decision-modes.md` 第 3 節架構圖把「原版 AI RE 研究,目前尚無此文件」改成指向本檔。
