# 開局母星初始狀態(手冊萃取)

> 目的:回答「新遊戲開局時,母星/初始殖民地的人口、建築、艦隊、科技、資源長什麼樣」,供 remake
> 真實起始狀態(取代目前 `RegenGalaxy` 程序生成)實作參考。日期:2026-07-10。
>
> **來源優先序**:①`moo2_patch1.5/GAME_MANUAL.pdf`(188 頁完整手冊,敘述性)②`moo2_patch1.5/MANUAL_150.html`
> (1.50 patch 說明書,含「Modding with Config」章節——這章雖是給 modder 看的參數說明,但**逐條點名了
> 遊戲內建預設值與判定機制**,是本文件最重要的單一來源)③`openorion2/src/tech.cpp`(研究樹靜態表,唯讀,
> 用來把手冊點名的科技/建築換算成精確 RP 成本)④既有專案文件(`newgame-flow.md`、`moo2-formulas-reference.md`、
> `community-mechanics-findings.md`)交叉引用,不重複已有內容。
>
> **CD Manual 說明**:`original_game/Master of Orion 2 - CD Manual.pdf` 只有 9 頁,是**安裝手冊**
> (System Requirements / Installing the Game / Loading the Game),第 7 頁原文明說「The instruction manual
> for Master of Orion II has been placed on the game CD-ROM in Adobe Acrobat Reader format」——真正的遊戲手冊
> 是另一個檔案,原始光碟版本本專案未收錄,**本文件不使用此檔案作為玩法來源**。所有玩法數字改以
> `moo2_patch1.5/GAME_MANUAL.pdf`(patch 1.5 隨附的完整手冊,內容涵蓋原版規則)為準。

## 一、Starting Civilization(起始文明等級,p.13)

`GAME_MANUAL.pdf` p.13「Starting Civilization」原文四級(Galactic Setup 畫面的第 5 項設定):

| 等級 | 手冊原文摘要 | 殖民地數 | 科技 | 艦隊 |
|---|---|---|---|---|
| **Pre-warp** | "Every race has one colony — their home star system. Exploring outside that system is impossible until faster than light (FTL) technologies are discovered." | 1(僅母星所在星系) | 無 FTL,見下「科技」節 | 無法離開本星系(無 FTL 引擎科技) |
| **Average** | "starts each race with the same single colony, but with all the technologies necessary for interstellar flight already achieved (plus a few random extras). Every empire has a small star fleet, including one Colony Ship capable of settling a planet." | 1 | FTL 已解鎖 + 隨機額外科技 | 「small star fleet」+ 至少 1 艘 Colony Ship |
| **Post-warp** | "an Average Civilization plus all technology fields up to 250 research points" | 1(與 Average 相同) | Average 基礎 + 額外累加至 250 RP 門檻內的所有科技領域 | 同 Average |
| **Advanced** | "Much of the galaxy is already explored and settled. Each race begins with several technological advancements in hand and a substantial fleet of ships capable of interstellar travel." | 多個(「much of the galaxy is already explored and settled」,對照 `newgame-flow.md`:Advanced 才會啟用多殖民地/`ApplyRace` 差異路徑)| 多項科技已知(手冊未列清單) | 「substantial fleet」(手冊未給精確數量,見下「待確認」)|

**與現有 `newgame-flow.md` 一致**:`newgame-flow.md` 第一節已標注此四級對應 Galactic Setup 第 5 決策項,本節在此基礎上補上四級各自的**手冊原文逐字引用**與下方「Initial Buildings」機制的量化細節。

## 二、初始人口

### 2.1 母星人口起始「數值」——手冊未給精確數字

`GAME_MANUAL.pdf` 與 `MANUAL_150.html` 全文搜尋「starting population / initial population / begin with a population」**零命中**。手冊「Your Home World」章節(p.54)只教玩家怎麼讀星球報告(恆星顏色、氣候、礦產),不含開局人口數字。

**⚠ 待原版確認**:母星初始人口具體數字(如「5 單位」)手冊沒有給,需要 DOSBox 開局存檔解析或逆向 `initial_buildings`/存檔初始化邏輯才能得到精確值。

### 2.2 人口決定「建築數量」的機制——手冊有明確規則(高可信度,一手來源)

`MANUAL_150.html`「Modding with Config → Population → Initial Buildings」段(這段雖標題是 modding,但描述的是**遊戲內建預設機制**,非 mod 專屬行為)：

> "The number and types of buildings that each colony has built at game start depends on: Starting tech level (Pre-warp, Avg, Post-warp, Advanced), Population size of a colony, Known techs and Priority listing in the table `initial_buildings`."
>
> "The number of starting buildings on each colony is capped to **3 for Pre-warp, 5 for Average/Postwarp and 9 for Advanced** game starts. Apart from cap, colonies with more population start with more buildings, the maximum number of buildings (**not counting the Capitol**) is **⅔ pop rounded up**. For example a HW with 8 pop can have 6 buildings on Advanced Tech start, but only 5 on Average start due to the cap."

即:每殖民地開局建築數 = `min(⅔ pop 無條件進位, 該起始等級的上限值)`,**Capitol 不計入此公式**(Capitol 是每個首都殖民地必有、不佔用建築格位的特殊建築,見下節)。這條規則本身**沒有給出母星人口的具體數字**,但確立了「人口→建築數」的量化關係,可作為 remake 從「已知人口」反推「應生成幾棟建築」的權威依據。

### 2.3 人口分配(農業/工人/科學家)——手冊未給開局分配比例

`GAME_MANUAL.pdf` 有完整的「指派人口去做農業/工業/科研」介面說明(Colony 畫面),但**沒有描述新遊戲開局時這批初始人口預設怎麼分配**。`community-mechanics-findings.md` 也未提及此項(未列入其 8 個研究主題)。

**⚠ 待原版確認**:開局人口分配比例(如「1 農業 + 1 工人」或全部待分配)無手冊來源,需 DOSBox 或存檔格式逆向確認。

## 三、初始建築

### 3.1 所有起始文明等級共通:Capitol + Engineering 兩項「絕對起始」

`MANUAL_150.html`「Modding with Config → Notes on Winning the Game → Score Calculation → Technology」段(原文,一手來源):

> "Tech field 0 (**Capitol, Spy Network and Pulse Rifle**) and Tech field Engineering (**Colony Base, Star Base and Marine Barracks**) are always known from the start and are worth 3 points each."

這句話同時回答了「初始建築」與「初始科技」兩個問題,且與下方 `openorion2/src/tech.cpp` 的靜態表**完全交叉印證**:

- `tech.cpp:212`(`TOPIC_ENGINEERING` 註解「already researched at game start」):`{50, 1, {TECH_COLONY_BASE, TECH_MARINE_BARRACKS, TECH_STAR_BASE}}` —— cost 50、flag 1(`ResearchAll`,全部視為已知),三項技術正是 Colony Base(特殊建設)、Marine Barracks(建築)、Star Base(衛星)。
- `tech.cpp:170`(`TOPIC_STARTING_TECH` 註解「already researched at game start」):`{0}` —— cost 0、無子項清單(因為手冊點名的 Capitol/Spy Network/Pulse Rifle 是「Tech field 0」這個通用起始欄位本身的效果,不是某個可研究主題底下的子技術,故 `research_choices[0]` 陣列為空,這點與手冊敘述一致而非矛盾)。

因此**任何起始文明等級**(含 Pre-warp)都保證擁有:

| 來源 | 效果 |
|---|---|
| Tech field 0(通用起始科技) | **Capitol**(首都建築,見下)、**Spy Network**(間諜能力解鎖)、**Pulse Rifle**(基礎陸戰隊裝備科技) |
| Tech field Engineering(Construction 分支永遠已知) | **Colony Base**(可建的殖民擴張特殊建設)、**Star Base**(可建衛星)、**Marine Barracks**(可建建築) |

### 3.2 Capitol——首都建築,不佔用建築格位

`MANUAL_150.html`:「Initial Buildings」段提到「the maximum number of buildings (**not counting the Capitol**)」——Capitol 是每個首都殖民地自動擁有、且**不計入⅔人口建築格位上限公式**的特殊建築。`GAME_MANUAL.pdf` 多處提及 Capitol 被攻陷會造成士氣懲罰(對照 `moo2-formulas-reference.md` §3 士氣節的「首都淪陷懲罰」表,-20%~-50% 依政府別),但沒有把 Capitol 本身列進「Buildings」大表(`colony-buildings.md` 因此**不收錄 Capitol**,因為它性質特殊,非玩家可建/可失去的一般建築)。

### 3.3 Pre-warp / Average 的實際起始建築:Marine Barracks + Star Base(僅此二項)

`MANUAL_150.html` 原文(高可信度,一手來源,直接點名):

> "**Pre-warp and Average Tech games only start with Marine Barracks and a Star Base** because no other techs are Known that are also in the default initial buildings list. If, for example, Colony Base building is added to the `initial_buildings`, it will be given to each players' homeworld at game start."

即:雖然 Engineering 科技解鎖了 Colony Base/Star Base/Marine Barracks 三項「可建」項目,但 Colony Base 本質是「用完即耗」的殖民行動(不是常駐建築,見手冊 p.75「Colony Base (Special)」條目),不會出現在殖民地的建築清單裡;故 Pre-warp/Average 開局母星的**已建成建築**只有 **Marine Barracks(建築)+ Star Base(衛星)** 兩項,加上不計入格位的 Capitol。三者總數對上 Pre-warp 的「上限 3」剛好吻合(Capitol 不算格位 + 2 個算格位的建築,若⅔人口≥2 則不受人口公式限制,直接受限於等級上限 3)。

### 3.4 `initial_buildings` 優先清單機制與已知的排序陷阱

`MANUAL_150.html` 原文:

> "The `initial_buildings` is a strict list, in that the game starts at entry 1, checks if it's allowed to place it and works its way down. The default list order potentially causes 1-3 pop colonies on Advanced to start without Marine Barracks because the first two list items are a satellite (#1-3) and Hydroponic Farms (#4). Marine Barracks is entry #5. See the Excel manual for further details."

這段透露 `initial_buildings` 表的**部分排序**(entry 1-3 = 某衛星、entry 4 = Hydroponic Farm、entry 5 = Marine Barracks),但完整清單(全部條目)手冊正文沒有列出,只說「詳見 Excel manual」(本專案未取得該 Excel 附錄)。

**⚠ 待原版確認**:`initial_buildings` 表完整排序(entry 6 以後)、Advanced 等級實際會出現哪些建築組合,需要該 Excel 附錄或存檔逆向才能確認。

### 3.5 Advanced 等級建築數量範例(手冊有具體算式範例)

沿用 2.2 節公式,手冊舉例:「a HW with 8 pop can have 6 buildings on Advanced Tech start, but only 5 on Average start due to the cap」——驗證:`⅔ × 8 = 5.33 → 無條件進位 6`;Advanced 上限 9(不受限,取 6);Average 上限 5(受限,取 5)。此例本身**已是手冊給的精確驗證數字**,可直接寫進 remake 的單元測試。

## 四、初始艦隊

### 4.1 手冊定性描述(Average 等級)

`GAME_MANUAL.pdf` p.13:「Every empire has a small star fleet, **including one Colony Ship** capable of settling a planet.」——只保證「至少 1 艘 Colony Ship」+「一支小艦隊」,沒有列出完整清單(型號/數量)。

### 4.2 「兩艘偵察艦」的間接證據(patch 1.50 changelog,可信度中——changelog 描述變更前後行為,隱含 classic 也是 2 艘)

`MANUAL_150.html` changelog 段(1.50 修正自動設計艦與手動設計艦不一致的 bug)：

> "A number of bugs causing auto-designed ships to be different from manually designed ships have been fixed. This includes both AI and player starting ships. ... Empty space provides a speed bonus. Consequently **the two starting scouts** will have 12 combat speed instead of 10."

這句在描述「修正後两艘起始偵察艦的戰鬥速度從 10 變 12」,隱含**經典版(1.31/1.5 前)就已经是 2 艘起始 Scout**,1.50 只改了速度數值、沒改數量。**可信度中**(changelog 附帶提及,非正式列表,但明確使用「the two starting scouts」這種確定性語氣,而非「some scouts」)。

### 4.3 尚無來源的部分

**⚠ 待原版確認**:
- Average/Post-warp 起始艦隊除「1 Colony Ship + 2 Scout」外是否還有其他艦(如 Outpost Ship、護衛艦)——手冊未列完整清單。
- Advanced 等級「substantial fleet」的具體型號與數量——手冊只有定性描述,無數字。
- Pre-warp 起始艦隊(理論上沒有 FTL 引擎,連星系都出不去)是否連 Colony Ship/Scout 都沒有,或者仍配備「系統內」用途的艦艇——手冊沒有明確說明,`newgame-flow.md` 也未確認。

## 五、初始科技

### 5.1 通用起始(所有等級皆有,§3.1 已列)

Tech field 0(Capitol/Spy Network/Pulse Rifle)+ Tech field Engineering(Colony Base/Star Base/Marine Barracks)。

### 5.2 各分支的「Tech field #0」機制(手冊有明確規則)

`MANUAL_150.html`「Modding with Config → Tech Tree」段:

> "The first tech field in each branch is always tech field #0 - Starting Technology. The field in position 0 and the techs that are in it are always assumed Known. ... The eight branches (or categories) of technology each have a specific starting tech field. These starting fields are keyed to fixed positions in the techfield list and can be set using `starting_techfield`. In seven branches the starting tech field is the first Researchable field of the branch, i.e. in a Pre-warp game the first fields shown when viewing the Research window. **The Construction branch however is different. No matter what you do, by default the first field is always Known (default is field #29 - Engineering).**"

即:8 個研究分支各自有一個「起始技術欄位」,7 個分支的起始欄位就是「第一個可研究欄位」本身(Pre-warp 玩家在 Research 畫面看到的第一格),**只有 Construction 分支特殊**——它的「起始欄位」被寫死指向 #29(Engineering),而非分支的第一個欄位。

### 5.3 Pre-warp:2 個已知科技欄位

> "Thus on Pre-warp, the game starts with two fields Known: Tech field 0 and the first field in Construction."

即 Pre-warp 只保證 Tech field 0(通用)+ Construction 分支的 Engineering 已知,其餘 7 個分支(Biology/Power/Physics/Force Fields/Chemistry/Computers/Sociology)**一個欄位都不知道**——這與「Pre-warp 沒有 FTL、困在本星系」的敘事一致(FTL 引擎科技在 Power 分支,Pre-warp 尚未解鎖)。

### 5.4 Average/Post-warp/Advanced:額外 6 個已知欄位(`prewarp_techfield` 表)+ 具體點名 5 個

> "Table `prewarp_techfield` is an ordered list with **6 tech fields** that are Known at the start of an Average, Post-warp or Advanced game. For Pre-warp only the first value matters (by default field #29)."

同一手冊「Score Calculation」段落又具體點名:

> "Thus, the Average Tech level starts with a tech score of 21 points (2\*3 points + 5\*3 points for **the Pre-warp tech fields Nuclear Fission, Cold Fusion, Chemistry, Electronics and Physics**)."

**⚠ 手冊內部有一處未完全對齊、如實記錄不代為裁決**:「`prewarp_techfield` 是 6 格清單」vs「Score Calculation 段具名列出 5 個欄位(Nuclear Fission/Cold Fusion/Chemistry/Electronics/Physics)」,兩處數字差 1。可能原因(未驗證,列出供後續查證):
1. `prewarp_techfield` 6 格清單中有 1 格是空/未使用(如對應 Construction 分支,但 Construction 已用 `starting_techfield` 單獨處理,若仍佔一個表格位但值無效則可解釋差 1);
2. 兩段描述的統計口徑不同(21 分公式明確算出 7 個已知欄位 = 2 基礎 + 5 具名,若 `prewarp_techfield` 6 格全部有效,則 Average 應有 8 個已知欄位而非 7,與 21 分公式矛盾)。

本文件**不臆測裁決**,如實記錄兩段原文,供後續用 DOSBox 存檔或 openorion2 tech.cpp 交叉比對釐清。**與現有 `moo2-formulas-reference.md` §6 的交叉觀察**:該文件列出研究樹「Cost=50 的 5 個主題」為 `TOPIC_CHEMISTRY/TOPIC_ELECTRONICS/TOPIC_ENGINEERING/TOPIC_NUCLEAR_FISSION/TOPIC_PHYSICS`(含 Engineering),與本節手冊點名的「5 個 Pre-warp 欄位」(Nuclear Fission/**Cold Fusion**/Chemistry/Electronics/Physics,不含 Engineering、改含 Cold Fusion)**不是同一組**——Engineering 是 Construction 分支自己的固定起始欄位(§5.2),不算在這 5 個「預熱」欄位內;Cold Fusion 是 Power 分支的第一個可研究欄位。兩份文件的列表看似相似但實際是兩個不同集合,記錄於此避免未來誤合併成同一張表。

### 5.5 Post-warp 額外:`postwarp_tech` 表最多 29 項已知科技

> "Changing tech positions might break the Post-warp Techtree. The table `postwarp_tech` was added and defines **a set of up to 29 Known technologies** at the start of a Post-warp game. The first 7 techs in this list are excluded in strategic mode."

對照 `GAME_MANUAL.pdf` p.13「Post-warp is an Average Civilization plus all technology fields up to 250 research points」——兩處互相印證:Post-warp = Average 基礎(§5.4)+ 額外一批具體科技(最多 29 項,原文未逐條列出,**待確認**其明細),累計研究點數上限約 250 RP。

### 5.6 Advanced:手冊沒有給精確科技清單

`GAME_MANUAL.pdf` p.13 只說「Each race begins with several technological advancements in hand」,`MANUAL_150.html` 也沒有給 Advanced 專屬的已知科技表。**⚠ 待原版確認**。

## 六、初始資源(BC 國庫、每回合產出)

### 6.1 手冊沒有給開局 BC 國庫數字

全文搜尋「starting BC / initial treasury / begin with ... BC」**零命中**。`MANUAL_150.html` 唯一出現的具體 BC 數字是作弊碼「`moola` Adds 1000 BCs to your imperial treasury」——這是**作弊指令**的效果,不代表正常開局國庫值,不可採用為起始 BC 數字。

**⚠ 待原版確認**:開局國庫 BC 數值(是否為 0、還是有初始預算)。

### 6.2 每回合產出——由建築/人口/礦產豐度決定,無「開局固定值」

手冊沒有給「開局第一回合固定產出多少 BC/production/research」的數字,因為這些數值本身是由 §2-3 已知的「人口分配 + 已建建築 + 星球礦產豐度」推導出來的**函數結果**,不是獨立設定值。參照 `moo2-formulas-reference.md` 已移植的 `PlanetBaseProduction`(依礦產豐度查表,§8)、`ColonyBaseGrowth`(§1)等公式——只要補上 §2.1/§2.3 缺的「起始人口數字與分配」,即可用現有公式算出開局第一回合的實際產出,不需要額外的「起始產出」手冊數字。

## 七、種族/政府/難度差異

### 7.1 種族天賦影響初始狀態,但影響的是「係數」而非「起始建築/人口清單」

`GAME_MANUAL.pdf`(p.15-18 各種族條目、p.24 天賦定義)描述 Large/Rich/Poor Home World 等種族天賦會影響母星的初始規模與礦產(如「Rich Home World means accelerated production」),但這些是**長期係數**(影響後續產出計算),不改變「開局第一回合建築/艦隊清單」這件事本身的機制(§3-4 的規則對所有種族一體適用,只有代入的「人口數字」「礦產豐度數字」會因種族/星球特性而不同)。

### 7.2 政府型態不影響開局初始狀態

手冊沒有提到政府型態(Feudalism/Democracy/…)會改變開局建築或艦隊清單;政府是遊戲中**後續可切換**的設定,不是開局階段的變量。(政府對既有殖民地的士氣/收入係數影響已在 `moo2-formulas-reference.md` §3/§4 記錄,與本文件「起始狀態」主題無關。)

### 7.3 難度(5 檔)不影響玩家自己的初始狀態,只影響 AI 與後續節奏

`GAME_MANUAL.pdf` p.9-10 五檔難度原文(Tutor/Easy/Average/Hard/Impossible)描述的都是「**其他種族**的產研速度與友善度」差異(如 Impossible:「The other races operate with significantly accelerated research and production」),**沒有一處提到玩家自己的起始人口/建築/艦隊會因難度改變**。AI 自身在不同難度下的加成表(Growth/Food/Prod/Res/BC/Command/Spy/Troops 九欄,五檔難度)已由 `original-ai-re.md` §2.1 完整收錄(來源同一 `MANUAL_150.html`),與本文件主題(玩家母星初始狀態)是兩個不同問題,不重複搬運,詳見該文件。

**「Tutorial」用詞澄清**:MOO2 沒有獨立的教學模式,任務提到的「tutorial 差異」對應的是 5 檔難度中最低檔 **Tutor**,其效果同上,只影響 AI 表現,不改變玩家母星初始狀態。

## 八、與現有專案文件的交叉驗證/補強

- **`terraform.go` 的氣候人口容量係數表(可信度提升)**:`moo2-formulas-reference.md` §5 該表(`25 25 25 25 25 25 40 60 80 100`)原本標注「交叉驗證用 openorion2 `climatePopFactors` 陣列」。本輪在 `MANUAL_150.html`「Population Capacities」段找到官方 modding 文件直接列出同一陣列:`pop_climate = 25 25 25 25 25 25 40 60 80 100;`,与既有表**逐項相符**。這是「遊戲出廠內建預設值」的官方文字確認(非社群逆向),建議可信度由「中(單一社群來源)」提升為「高(官方 modding 文件 + openorion2 陣列雙重確認)」。
- **`pop_max = 5 10 15 20 25;`**(星球大小 Tiny→Huge 五級的人口容量基準值)與 `community-mechanics-findings.md` §7「星球最大人口 = 5 × 星球大小等級」完全吻合,同樣可將該條可信度由「中(單一 lparchive 來源)」提升為「高」。
- **`newgame-flow.md` 第六節「⚠ 待 oracle」的真實起始狀態**:該文件先前結論是「手冊無精確起始數值,需 DOSBox 開局參考存檔解析」。本輪重新從 `MANUAL_150.html`(而非只查 `GAME_MANUAL.pdf`)找到了「Initial Buildings」「Tech Tree」兩段機制性說明,**部分推翻該結論**——建築數量與科技已知與否**有**明確規則可循(§3-5),不是完全無來源;但**具體的人口數字、BC 國庫數字、Advanced 完整艦隊/科技清單**確實仍待 DOSBox 或 Excel 附錄驗證。建議 `newgame-flow.md` 第六節後續更新時引用本文件,不要再寫「手冊完全無來源」這種一刀切結論。

## 九、彙整:待原版確認清單(誠實列表)

| 項目 | 狀態 |
|---|---|
| 母星開局人口具體數字 | 無來源,待 DOSBox/存檔逆向 |
| 開局人口分配比例(農/工/科) | 無來源 |
| `initial_buildings` 表完整排序(entry 6+) | 無來源(手冊指向 Excel manual,本專案未取得) |
| Advanced 等級具體已知科技清單 | 無來源 |
| Advanced/Average 起始艦隊完整型號與數量(除 1 Colony Ship + 2 Scout 外) | 無來源 |
| `postwarp_tech` 29 項科技明細清單 | 無來源 |
| 開局 BC 國庫數值 | 無來源 |
| `prewarp_techfield` 6 格 vs Score Calculation 5 個具名欄位的落差 | 手冊內部兩段未完全對齊,已如實記錄,不裁決 |

## 參考來源

| 來源 | 用途 |
|---|---|
| `moo2_patch1.5/GAME_MANUAL.pdf` p.9-14(Galactic Setup/Starting Civilization)、p.54(Your Home World)、p.52(History Graph 建築成本計分) | 起始文明四級定性描述、人口/建築計分規則 |
| `moo2_patch1.5/MANUAL_150.html`「Modding with Config → Population/Buildings/Tech Tree/Notes on Winning the Game」 | Initial Buildings 機制、Tech field 0/Engineering 起始科技、Score Calculation 起始科技點名、Population Capacities 陣列 |
| `openorion2/src/tech.cpp:169-220` | `TOPIC_STARTING_TECH`/`TOPIC_ENGINEERING` 的 RP 成本與子技術清單,交叉驗證手冊敘述 |
| `docs/tech/newgame-flow.md` | 新遊戲畫面流程、Galactic Setup 8 項決策(本文件的姊妹文件,互相引用不重複) |
| `docs/tech/moo2-formulas-reference.md` §5/§8 | 人口容量係數表、礦產豐度查表(本輪找到官方文字交叉驗證,可信度提升) |
| `docs/tech/community-mechanics-findings.md` §7 | 人口上限公式的社群補充(本輪找到官方文字交叉驗證) |
| `docs/tech/original-ai-re.md` §2.1 | AI 難度加成表(與本文件「玩家起始狀態」主題互補,不重複) |
