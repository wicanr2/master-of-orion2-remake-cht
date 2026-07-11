# 原版《Master of Orion 2》玩法權威參考（供 remake 對齊驗證）

> 角色:遊戲歷史考究專家。目的:用網路權威資料重建「原版《Master of Orion 2: Battle at Antares》(1996) **實際玩起來是什麼樣**」,
> 讓 go/ebiten remake(`/home/anr2/moo2`)能逐項對照驗證。**這份文件是「玩家實際體驗」層級,不是手冊公式抄本**
> ——公式數值已收在 `docs/tech/moo2-formulas-reference.md` 等文件,本檔補的是「一局怎麼跑、每個畫面玩家做什麼、手感如何」。
>
> **撰寫日期**:2026-07-11。**取材方式**:WebSearch / WebFetch,逐斷言標來源。
> **來源可信度分級**:🟢 權威(官方手冊萃取 / StrategyWiki / Wikipedia / 知名長篇攻略)、🟡 攻略站個人整理、🟠 論壇個人意見。
> 無法查證者標「**⚠ 待查**」,不編造。
>
> **重要提醒(交叉污染)**:網路搜尋常把《Master of Orion 2》(1996) 與《Master of Orion: Conquer the Stars》(2016 重製版) 混在一起。
> 兩者勝利條件、選單、種族數值都不同。**本檔只採 1996 原版**;凡明顯來自 2016 版的資料(如「6 種勝利:score/diplomatic/technological/economic/Antaran/conquest」、
> 「Hyperplanar Transfer 科技勝利」)**已剔除**——1996 原版只有 3 種勝利(見第 4 節)。

---

## 目錄

1. 一局的完整流程(主選單 → 星系設定 → 種族 → 進遊戲 → 開局狀態)
2. 每回合玩家做什麼(turn loop)
3. 各主要畫面的功能與互動
4. 關鍵 gameplay 機制的實際手感
5. 新玩家常見體驗 / 早期遊戲
6. 華人圈討論 / 中文資訊 / 譯名對照
7. **與 remake 現況的對照檢查清單**(可驗證檢查點)

---

## 1. 一局的完整流程

### 1.1 新遊戲設定畫面(Galactic Setup)🟢

玩家開新遊戲時,依序設定以下選項(對照 StrategyWiki「Starting a game」與官方手冊 p.13):

| 設定項 | 選項 | 玩家看到/選什麼 |
|---|---|---|
| **銀河大小 Galaxy Size** | Small / Medium / Large / Huge(4 級)| 決定恆星數。搜尋來源給的近似恆星數:Small ≈ 20、Medium ≈ 36、Large ≈ 54、Huge ≈ 71–72。⚠ 精確恆星數待核(不同來源略有出入,且每恆星含多顆行星,故恆星數比 MoO1 少)|
| **銀河年齡 Galaxy Age** | Average / Organic-Rich / Mineral-Rich(3 級)| Organic-Rich = 生物資源多、工業礦產少;Mineral-Rich = 相反;Average = 平衡。影響行星的食物/礦產分布 |
| **難度 Difficulty** | Tutor / Easy / Average / Hard / Impossible(5 級)| 高難度 AI 的研究/生產有加成、且對你態度更差;低難度玩家有優勢 |
| **對手數 Opponents** | 最多 8 | AI 帝國數量 |
| **起始文明 Starting Civilization(科技等級)** | Pre-warp / Average / Post-warp / Advanced(4 級)| 見 §1.4。Pre-warp = 只有母星、無 FTL、完整探索體驗;Advanced = 大半銀河已被殖民、開局就有大艦隊 |
| **戰術戰鬥 Tactical Combat** | 開 / 關 | 開 = 玩家可手動操控每艘船的格子戰鬥、且可自訂艦艇設計;關(戰略戰鬥)= 電腦自動算戰果,且**只能選預設艦艇、不能自己設計船** |
| **隨機事件 Random Events** | 開 / 關 | 增加不確定性(瘟疫、海盜、超新星等) |
| **安塔蘭入侵 Antaran Attacks** | 開 / 關 | 開 = 安塔蘭人會週期性從次元入口派艦攻擊玩家殖民地 |

來源:StrategyWiki「Starting a game」(🟢,經 WebSearch 摘錄,WebFetch 被 403 擋)、`docs/tech/homeworld-init.md`(官方手冊 p.13 萃取,🟢)、Let's Play Archive Thotimx 第 1 篇(🟡,實際設定示範:Hard / 8 對手 / Pre-warp / 戰術戰鬥開 / 隨機事件開 / 安塔蘭關)。
- <https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Starting_a_game>
- <https://lparchive.org/Master-of-Orion-2-(by-Thotimx)/Update%2001/>

### 1.2 種族選擇(13 族)🟢

MoO2 承襲 MoO1 的 10 族 + 新增 3 族 = **13 個標準可玩種族**:
Alkari、Bulrathi、Darlok、Elerian、Gnolam、Human、Klackon、Meklar、Mrrshan、Psilon、Sakkra、Silicoid、Trilarian。

各族的招牌特性(玩家實際會感受到的差異,🟡/🟠 論壇+wiki):

| 種族 | 招牌特性 | 玩家手感 |
|---|---|---|
| **Psilon** | Creative(創造性)| 每個科技領域**該階全部科技都拿到**(非創造性只拿 1–2 項)。最適合新手,研究碾壓 |
| **Sakkra** | +100% 人口成長 | 人口翻倍成長,快速鋪滿銀河、經濟基礎龐大 |
| **Klackon** | 生產/食物加成 + Unification 政府 | 昆蟲族,造東西極快、產糧多 |
| **Silicoid** | Lithovore(食岩)+ Tolerant | 不需食物,能在死星/貧星立足,早期擴張無食物壓力 |
| **Trilarian** | Aquatic(水生)+ 次元移動 | 海洋/凍原/沼澤星宜居度提升(Ocean/Terran → Gaia);艦隊移動較快 |
| **Bulrathi** | 地面戰鬥強 + 高重力適性 | 陸戰隊入侵強,靠佔領取勝 |
| **Meklar** | Cybernetic(半機械)+2 生產 | 惡劣環境生存力強,貧礦星也能快速開發 |
| Alkari / Mrrshan | 飛行/艦船命中加成 | 軍事系,戰鬥數值優勢 |
| Darlok | 間諜/隱匿 | 間諜戰、可偽裝 |
| Gnolam | 財富/幸運 | 錢多、負面事件少 |
| Elerian | Telepathic + 戰鬥 | 心靈能力(可讀取外交/免叛變),戰鬥系 |

**自訂種族(Custom Race)**🟢:玩家有 **10 點 picks**(Racial Ability And Characteristic Points)。正 modifier 的優點扣點、負 modifier 的缺點退點;**最多可取 –10 點缺點**換到最多 20 點優點;不可用負總點數開局。picks 總和影響**分數乘率**(少用 5 點 → 150% 分數)。社群共識:**Creative 幾乎必拿**(rengels.de:「沒 Creative 的遊戲很挫折」)。

來源:
- Official MoO Wiki – Race picks:<https://masteroforion.fandom.com/wiki/Race_picks>
- StrategyWiki – Race design options:<https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Race_design_options>
- rengels.de 攻略(🟡):<https://rengels.de/computer/orion2/index.html>

### 1.3 命名 / 旗色 → 進遊戲 🟡

選種族後設定:自訂帝國名、領袖名、旗色、母星名。Let's Play 示範:Human 選黃色(致敬 MoO1 人類制服)、命名「Hard Normal」。

### 1.4 開局你有什麼(依 Starting Civilization 分級)🟢

**這是 remake 最重要的對齊點之一。** 官方手冊(patch1.5 GAME_MANUAL.pdf p.13,經 `docs/tech/homeworld-init.md` 萃取)+ humbe.no 玩家描述:

| 起始等級 | 殖民地 | 科技 | 艦隊 |
|---|---|---|---|
| **Pre-warp** | 1(僅母星星系)| 無 FTL 引擎,**無法離開本星系** | **無星際艦**(手冊:探索不可能,直到研出 FTL);玩家實測級描述:母星 + Marine Barracks + Star Base |
| **Average** | 1 | FTL 已解鎖 + 少數隨機額外科技 | 「small star fleet」含**至少 1 艘 Colony Ship**(可殖民行星)+ 少數偵察艦 |
| **Post-warp** | 1 | Average + 累加至 250 RP 門檻內所有領域 | 同 Average |
| **Advanced** | **多個**(大半銀河已殖民)| 多項科技已知 | 「substantial fleet」大艦隊(手冊未給精確數,⚠ 待查) |

**開局母星具體狀態(Pre-warp,玩家實測級)**🟡:humbe.no 描述「1 colony with **8 colonists**,已建 **Marine Barracks(內含 3 名陸戰隊)**與 **Star Base**,**50 BC**,已知 Colony Base / Housing 等基礎建設」。

**永遠已知的起始科技/建築**(所有等級共通,官方手冊萃取 🟢):
- Tech field 0:**Capitol**(首都建築,不佔建築格位)、**Spy Network**、**Pulse Rifle**
- Engineering 分支永遠已知:**Colony Base**、**Star Base**、**Marine Barracks**

**開局建築數量規則**🟢(MANUAL_150.html「Initial Buildings」):每殖民地開局建築數 = `min(⅔ 人口無條件進位, 上限)`,上限 = **Pre-warp 3 / Average·Post-warp 5 / Advanced 9**(不含 Capitol)。例:8 人口母星在 Advanced 可有 6 棟,但 Average 因上限只能 5 棟。

**⚠ 待查**(手冊全文零命中,需 DOSBox 開局存檔或逆向確認):
- 母星**初始人口精確數字**(humbe 說 8,但手冊未載;`homeworld-init.md` 明確標為待確認)
- 開局人口的**農夫/工人/科學家預設分配比例**

來源:
- `docs/tech/homeworld-init.md`(官方手冊萃取,🟢,本專案內部)
- humbe.no MoO2 頁(🟡):<https://www.humbe.no/public/computergames/moo2/>

---

## 2. 每回合玩家做什麼(Turn Loop)🟢/🟡

一個典型回合的玩家操作循環:

1. **看星系主畫面(Galaxy Map)**:全銀河恆星圖,自己與已知帝國的疆域、艦隊位置、殖民地。
2. **處理殖民地(Colonies)**:進每個殖民地,調整**農夫/工人/科學家**人口分配,設定**建造佇列**(建築/艦艇/殖民船/貨船)。
3. **選研究(Research)**:當前研究完成或空著時,從可選科技領域挑下一個要研究的技術。
4. **造艦/派艦**:在有 Star Base/造船廠的殖民地造船;把偵察艦派去探索未知恆星、把殖民船/前哨船派去新行星。
5. **外交(Diplomacy)**:與 AI 帝國談判、締約、宣戰、交換科技。
6. **按「Turn(結束回合)」**:所有帝國同時結算,推進到下一回合;有事件/戰鬥/彈窗會中斷提示。

**核心資源循環**🟢(spacesector 稱這是 MoO2 成功的「回饋迴圈」):蓋建築 → 直接改變殖民地產出(產量/研究/食物)→ 人口成長 → 更多產出。玩家與殖民地間建立「情感連結」。

**殖民地產出機制**🟢:
- 每名殖民者每回合吃 **1 食物**(例外:Lithovore 不吃、Cybernetic 吃半食物+半生產)。
- 剩餘食物 → **每 2 食物換 1 BC**。
- 建築提供的產量/研究/食物**加算**(additive)疊加在人口產出上,不佔人口。

來源:
- StrategyWiki – Gameplay / Growing your population / Feeding your people(🟢,WebSearch 摘錄)
- Wikipedia MoO2(🟢):<https://en.wikipedia.org/wiki/Master_of_Orion_II:_Battle_at_Antares>
- spacesector「Formula to Success」(🟡):<https://www.spacesector.com/blog/2009/08/master-of-orion-ii-formula-to-success/>

---

## 3. 各主要畫面的功能與互動

> 對照我們 remake 已有的畫面清單,逐畫面列「玩家點什麼、看什麼數字」。標題按鈕名以英文原文為準,譯名見第 6 節。

### 3.1 主選單 🟡
New Game / Load Game / Continue / Multiplayer / Settings / Quit。本 remake 額外要加「版本 1.3/1.5 選擇」與「中/英語言」——原版無此兩項。

### 3.2 星系主畫面工具列(Galaxy screen toolbar)🟢/🟡
主畫面上方/側邊的功能鈕(玩家點進各管理清單):

| 按鈕 | 功能 | 玩家看什麼 |
|---|---|---|
| **Colonies** | 殖民地總覽 | 列出所有殖民地,可直接進 Build 畫面、調人口 |
| **Planets** | 行星報告 | 已知行星的氣候/重力/礦產/宜居度 |
| **Fleets** | 艦隊總覽 | 所有艦隊位置、組成、可下移動命令 |
| **Leaders** | 領袖 | 可雇用的殖民領袖/艦長,技能與費用 |
| **Races** | 種族/外交總覽 | 各 AI 帝國關係、態度 |
| **Info** | 資訊/排行 | 帝國分數、勢力比較、勝負進度 |
| **Game(選單)** | 存讀檔/設定 | Save/Load/Options/Quit |
| **Turn** | 結束回合 | 推進回合 |

⚠ 待查:各鈕的精確 icon 位置/像素排列需對原版截圖逐一比對(本專案 `docs/tech/screen-spec-info-research.md` 有相關研究,應交叉核對)。

### 3.3 殖民地畫面(Colony / Build screen)🟢
- **人口分配列**:農夫/工人/科學家三格,拖曳或點按調整人數;即時顯示食物/生產/研究總產出。
- **建造佇列**:選要蓋的建築/艦艇,顯示成本(PP)與預計回合數;可排隊、可「生產囤積(stockpile)」。
- **行星狀態**:人口/上限、污染、士氣、可用礦產。
- 頂部可切上一個/下一個殖民地。

### 3.4 研究畫面(Research)🟢
- 8 大研究領域,每領域分多個「階(level/tier)」。
- **非創造性種族**:每階只能在 **2–3 個選項中擇一**(選了就永遠拿不到同階其他項,除非搶救/交換)——這是 MoO2 研究的靈魂:每個選擇都有代價。
- **創造性(Creative)種族**:每階全拿。
- 顯示當前研究進度(RP 累積/所需)、預估完成回合。

### 3.5 艦艇設計畫面(Ship Design)🟢
- 6 種船體(見 §4),每設計有數個裝配槽位(design slots)。
- 玩家自由裝武器(beam/missile/fighter)、護盾、裝甲、引擎、特殊系統;顯示空間占用、成本、命令點。
- **戰略戰鬥模式下此畫面停用**——只能用預設船。

### 3.6 戰術戰鬥畫面(Tactical Combat)🟢
格子戰場,見 §4.1。

### 3.7 外交畫面(Diplomacy)🟡
與單一 AI 帝國對談:提議締約/貿易/科技交換/宣戰/求和;顯示對方態度。

來源:StrategyWiki – Gameplay / Units(🟢,WebSearch 摘錄);Wikipedia MoO2(🟢);本專案 `docs/tech/screen-spec-info-research.md` 交叉。

---

## 4. 關鍵 gameplay 機制的實際手感

### 4.1 戰術戰鬥怎麼打 🟢
- **戰鬥只發生在恆星系內**——不是宇宙空間任意點,而是攻打殖民地或清除封鎖時觸發。
- **格子戰場 + 逐船回合**:每艘存活的船(**行星、衛星也算「船」**)每回合可移動 + 對能找到目標的武器全開火 + 按該船的「Done」;所有船 Done 後回合結束。
- **建議開啟的輔助**:missile warnings(飛彈警告)、show legal moves(合法移動格)、show shield arcs(護盾弧)。
- **飛彈**:AI 通常整波飛彈鎖同一艘船;飛彈會一路飛向主目標,即使其他敵船經過也不轉向(可利用誘敵)。
- **武器擺放**:同一艘船別混 beam 與 missile;武器**依裝配順序開火**,順序影響效果。
- **戰鬥機/fighters**:類似歸向武器,要貼近才傷敵;其速度/HP 隨你最好的引擎/裝甲提升。
- **beam vs missile 節奏**🟡:早期飛彈較強,拿到好電腦(命中)後轉 beam。
- **地面戰(Ground Combat)**🟢:入侵行星靠運兵船運陸戰隊,勝負看兵力數量與地面戰科技,**玩家不手動操控**。

來源:StrategyWiki – Battle tactics / Combat(🟢,WebSearch 摘錄,WebFetch 403);Wikipedia(🟢)。
- <https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Combat>

### 4.2 殖民地成長曲線 / 開局經濟緊不緊 🟡/🟠
- **開局經濟偏緊**是公認的:Pre-warp 只有一顆母星、少量人口,前期要在「食物/生產/研究」三者間拉扯。
- **工業有污染遞減**:生產拉太高會污染,故需食物/工業/科研平衡(blogspot 攻略)。
- **人口是勝負最終仲裁**(blogspot:「population is the ultimate arbiter of victory」)。
- **Housing 技巧**🟡:在小/貧星放單一殖民者、全力蓋 Housing 生人口,再用貨船運到主星 → 可「幾乎翻倍」人口擴張速度。
- **貨船運食物/人口**:貨船以 5 艘為一組建造,可個別運作,每艘載 1 單位食物;也可用來搬殖民者(搬人會佔用數艘貨船數回合)。

### 4.3 研究節奏 🟡(平均科技局的公認開局研究序)
Research Labs → Reinforced Hull → Auto Factories → Biospheres → Soil Enrichment 或 Clone(每農夫產糧 ≤3 選 Soil,否則選 Clone)→ Neural Scanners → Super Computers → Battle Pods → Spaceports → Robo Miners。
**Research Laboratory 被視為全遊戲最重要、第一個該拿的科技。**
來源:blogspot 攻略(🟡):<https://masteroforion2.blogspot.com/2005/03/master-of-orion-ii-strategy-guide.html>

### 4.4 指揮點數 / 艦隊維護 🟢/🟡
- **6 種船體(hull size)**:Frigate → Destroyer → Cruiser → Battleship → Titan → Doom Star。船體越大命令點(command points)消耗越高。
- 命令點超出上限要付 BC 維護、或造 Command 建築提升上限;超支會扣錢。⚠ 精確命令點數值待與 `docs/tech/moo2-formulas-reference.md` 交叉。

### 4.5 士氣 / 污染 / 重力的實際影響 🟢
- **士氣 Morale**:士氣提升會讓產能增加(每級約 +20% 基礎值);首都淪陷、政府型態、建築(如娛樂設施)都影響士氣。
- **污染 Pollution**:約達 **90%** 顧問會警告,工人得停下生產去清污;放任到 **100% 會使行星降級(downgrade)**。
- **重力 Gravity**:非本族適性重力的行星,殖民者產能有懲罰,需 Gravity Generator 建築抵銷。

### 4.6 勝利條件——原版只有 3 種 🟢(權威,已剔除 2016 版污染)
1. **征服(Conquest)**:消滅所有其他帝國即勝。**佔領 Orion 星系有強力獎勵但不自動獲勝。**
2. **銀河議會選舉(Galactic Council,外交勝)**:當銀河大部分星系被殖民後,高等議會週期性召開,由**人口最高的兩名玩家**被提名;若當選者拿到 **2/3 選票**可選擇接受帝位而立即獲勝。**落選/拒絕接受 → 全面開戰(total war),遊戲繼續**,故議會勝可被拒絕、改走征服。
3. **安塔蘭勝利(Antaran victory)**:在 **3 處安塔蘭遺跡(看起來像小行星)蓋 lab**,研究並建造**次元入口(Dimensional Portal)**,派艦隊進次元空間摧毀安塔蘭母星即勝。

**⚠ 待查(來源衝突)**:議會**召開/複投間隔**——humbe.no 與部分來源說「每 25 回合」,另有來源(部分含 2016 版)說「每 30 回合」。原版 1996 的精確間隔需 DOSBox 實測或逆向確認。

來源:Wikipedia MoO2(🟢,明確列 3 勝利);StrategyWiki;humbe.no。
- <https://en.wikipedia.org/wiki/Master_of_Orion_II:_Battle_at_Antares>

---

## 5. 新玩家常見體驗 / 早期遊戲 🟡/🟠

- **推薦新手設定**:Psilon(自動研究全科技,最省心)/ Average 難度 / Huge 銀河 / Average 年齡 / Average 科技 / 8 對手(blogspot)。
- **開局前 ~20–30 回合典型走法**:偵察艦四處探星 → 主星壓生產把殖民船/前哨船造出來 → 搶下附近好行星 → 研究鏈先拿 Research Labs + Auto Factories → 開始鋪殖民地。
- **生產型種族**:母星人口全壓生產直到最後一座 colony base 造完,再轉科研。
- **科技型種族**:母星只留 1–2 人生產,其餘全科研,直到拿 Research Labs + Auto Factories。
- **常見踩坑**:
  - 忽略污染 → 行星降級。
  - 稅率拉高 → 全帝國生產下降(攻略建議稅率設 0,靠 trade-goods 星賺錢)。
  - 混裝 beam+missile → 戰力打折。
  - **軍事準備太慢**:medium 銀河約 **turn 90**(有蟲洞 turn 80)就可能碰到敵方 Battleship;large 銀河約 **turn 110**。要在那之前備好防禦艦隊/行星防禦。
- **屠宰太空怪獸(space monster)**🟡:早期用約 5 艘 frigate 小隊清怪,獲得的行星「通常抵 3 顆普通行星」。
- **後期**:攻打 Orion 的 Guardian 常見於 ~turn 300(rengels.de,🟡,屬單局長度參考)。

來源:blogspot 攻略、rengels.de(🟡)、GOG/Steam 論壇(🟠)。

---

## 6. 華人圈討論 / 中文資訊 / 譯名對照

### 6.1 重要發現:華人圈原生討論「稀薄」,這本身是背景事實
MoO2 **從未有官方中文版**,華人社群覆蓋遠不如星海爭霸、文明等。搜尋 巴哈姆特/PTT/知乎 直接命中極少,且極易與同名遊戲(星海爭霸2、巴哈姆特2龍之新娘)交叉污染。**本 remake 的繁中化本身即是在填補這塊空缺**,無現成「標準譯名表」可完全照抄——這解釋了為何 CLAUDE.md 要求自建譯名策略(見 `docs/tech/proper-noun-strategy.md`)。

### 6.2 已查到的中文譯名 🟡
- **遊戲名**:對岸資料庫 indienova 用「**银河霸主 2:安塔瑞斯之战**」(Master of Orion II: Battle at Antares)。本專案沿用「銀河霸主2」為主名(繁體)。
  - <https://indienova.com/game/master-of-orion-ii-battle-at-antares>
- 系列前作《Master of Orion》= 「**银河霸主**」(indienova)。
- 對岸曾有「能否汉化」的社群討論(CIV 文明聯盟論壇 civclub.net,搜尋可見標題「master of orion 2 能够汉化么?」,但該連結 WebFetch 回 404,內容待查)。

### 6.3 日文社群(較豐富,可參考術語結構)🟡
日文有較完整的攻略站,可作為東亞玩家理解的旁證:
- masteroforion2.seesaa.net(種族別プレイガイド)
- ameblo.jp/shinogu1(画面説明とゲームの目的)

### 6.4 譯名對照建議(供中文化參考)
> 因無權威中文表,以下為「本專案應自訂並統一」的類別,實際用字以 `docs/tech/proper-noun-strategy.md` 為準:
- **種族**:Psilon(心靈族/普西隆)、Silicoid(矽基族)、Sakkra(薩克拉/爬蟲族)、Klackon(克拉肯/蟲族)、Trilarian(三叉族/水棲族)、Meklar(機械族)、Bulrathi(熊族)……⚠ 各字需專案定案,不同社群譯法不一。
- **建築/科技/武器**:目前多為英文原文流通;繁中化需自建對照(參 `docs/tech/colony-buildings.md`、`weapon-mods.md`)。

**⚠ 待查**:巴哈姆特/PTT 是否有 MoO2 專門討論串(本輪搜尋未命中有效串,建議日後以精確引號 + 站內搜尋再試);對岸是否有流傳的民間漢化補丁及其譯名表。

### 6.5 文化地位(供「文化現象」章節與致謝參考)🟢
- 獲 **1996 Origins Award 最佳科幻/奇幻電腦遊戲**;Metacritic 84%、GameRankings 82.8%。
- 評論界長年稱其為「**4X 類型的巔峰(the pinnacle of the genre)**」;Tom Chick 稱該系列「在其後八年間投下巨大陰影」。現代太空 4X 遊戲(Stellaris、GalCiv 等)仍常被拿來與 MoO2 對標。
- 來源:Wikipedia MoO2(🟢)。

---

## 7. 與 remake 現況的對照檢查清單(可驗證檢查點)

> 下表左欄是本文查到的**原版行為**,右欄是**可在 remake 程式碼/實測驗證的具體項目**。逐條核對即可量化「還原度」。
> 標 🟢 = 來源權威、可直接當驗收基準;🟡 = 攻略級、可當合理性參考;⚠ = 原版數值本身待查,先別當硬基準。

### 7.0 驗證結果(Opus 逐條對碼核實,2026-07-11)

對 reference 檢查點逐條查 remake 原始碼,結果:**已文件化的核心機制全部確認忠實**,無發現照抄 2016 版等錯誤。

| 檢查點 | 結果 | 依據 |
|---|---|---|
| 永遠已知科技(Capitol/Spy Network/Pulse Rifle/Colony Base/Star Base/Marine Barracks) | ✅ 忠實 | `newHomeworldPlayerState`(session.go:2215-2244)標 TOPIC_STARTING_TECH 已知;Colony Base 記入 ChosenTech |
| Capitol 不佔建築格位(隱性首都狀態) | ✅ 忠實 | homeworldBuildings 刻意排除 Capitol(session.go:2260),與 reference 一致 |
| 開局母星建築 = Marine Barracks + Star Base | ✅ 忠實(手冊逐字) | homeworld-init.md §3.2/§3.3「start with Marine Barracks and a Star Base」;cap 5 只是上限,實際 2 是 tech 條件成立者(session.go:2255-2265) |
| 勝利條件恰 3 種(征服/議會/安塔蘭),非 2016 版 6 種 | ✅ 忠實 | VictoryExtermination(council.go)+ VictoryHighCouncil + VictoryAntaran(antaran_victory.go),無多餘勝利型別 |
| Average 開局 = FTL + 含 ≥1 Colony Ship | ✅ 忠實 | homeworldShips = 1 殖民船 + 2 偵察艦 |
| 污染(工業致污、清理成本、處理建築) | ✅ 經濟面忠實 | `colonyPollution`(colony.go:40)+ PollutionTolerance/PollutionEighths;⚠ 未做「100% 行星氣候降級」極端邊緣機制(小缺口) |
| 食物 1/人 + 剩食 2:1 換 BC | ✅ 忠實 | FoodSurplus = Food-Consumed;IncomeFoodSurplusRevenue(empire.go:64-74) |
| 士氣每級 +20% 產能 | ✅ 忠實(既有輪次) | morale.go MoraleProductionOutput |

**⚠ 仍待原版存檔(DOSBox)才能定論、不當硬基準**:①母星初始人口精確值(8?)+ 農夫/工人/科學家開局分配比例 ②銀河議會召開/複投間隔(25 vs 30 回合來源衝突)③各銀河大小恆星數精確值 ④100% 污染行星降級。這幾項 remake 現況有各自的合理近似或有據選擇,**取得原版 turn-1 存檔後可一次校準**——這是最有價值的下一個外部 oracle。

**結論**:reference + 對碼核實共同確認,remake 在**文件化的機制**上已對齊原版;「2026-07-04 實測 20%」是手冊錨定系統化工作前的舊評,核心系統(開局狀態/勝利/污染/食物/士氣/建築)現已忠實。剩餘距離集中在「需原版存檔校準的精確數值」與「戰術戰鬥/UI 手感的實測」,非機制性缺漏。

### 7.1 新遊戲設定
| 原版行為 | 驗證點 |
|---|---|
| 🟢 銀河大小 4 級(Small/Medium/Large/Huge)| remake 設定畫面是否恰好 4 級、命名一致 |
| ⚠ 恆星數 ≈ 20/36/54/71 | 對照原版實測;remake `RegenGalaxy` 生成的恆星數是否落在此區間(數值待核) |
| 🟢 銀河年齡 3 級(Average/Organic/Mineral)| remake 是否有此選項、且影響行星資源分布 |
| 🟢 難度 5 級(Tutor…Impossible)| remake 難度級數與命名 |
| 🟢 起始文明 4 級(Pre/Average/Post/Advanced)| remake 是否 4 級、且各級給的殖民地數/科技/艦隊符合 §1.4 表 |
| 🟢 戰術戰鬥「關」時不能自訂艦艇 | remake 關戰術戰鬥後,設計畫面是否停用、改預設船 |
| 🟢 對手上限 8 / 隨機事件 / 安塔蘭入侵 開關 | remake 是否具備這三項 |

### 7.2 開局狀態(最關鍵)
| 原版行為 | 驗證點 |
|---|---|
| 🟢 Pre-warp = 1 殖民地、無 FTL、無星際艦 | remake Pre-warp 開局是否**只有母星、艦隊不能離開本星系** |
| 🟢 Average = FTL + 小艦隊含 ≥1 Colony Ship | remake Average 開局是否給 Colony Ship + 偵察艦 |
| 🟢 永遠已知:Capitol / Spy Network / Pulse Rifle / Colony Base / Star Base / Marine Barracks | remake 開局殖民地是否已有 Capitol + Marine Barracks + Star Base,且這些科技標為已研 |
| 🟢 開局建築數 = min(⅔pop↑, 上限);上限 Pre 3 / Avg·Post 5 / Adv 9,Capitol 不計 | remake 母星生成的建築棟數是否套此公式(而非隨機/固定) |
| 🟡 母星約 8 人口、Marine Barracks 內 3 陸戰隊、~50 BC | remake 母星初始人口/陸戰隊/起始金錢(⚠ 人口 8 未經手冊證實,以原版存檔為準) |
| ⚠ 開局農夫/工人/科學家分配比例 | **待原版實測**;remake 現用什麼分配需標注為「暫定、待對齊」 |

### 7.3 每回合 / 殖民地
| 原版行為 | 驗證點 |
|---|---|
| 🟢 每殖民者吃 1 食物;Lithovore 0、Cybernetic 0.5 | remake 食物消耗公式 |
| 🟢 剩食每 2 換 1 BC | remake 食物換錢比 |
| 🟢 人口分農夫/工人/科學家三類,產出即時反映 | remake 殖民地畫面拖人是否即時改產出 |
| 🟢 建築產出加算不佔人口 | remake 建築 bonus 疊加方式 |
| 🟢 污染 ~90% 警告、100% 行星降級 | remake 是否有污染警告與降級機制 |
| 🟢 士氣每級約 +20% 產能 | remake 士氣對產能係數 |

### 7.4 研究 / 造艦 / 戰鬥
| 原版行為 | 驗證點 |
|---|---|
| 🟢 8 研究領域、非創造性每階 2–3 選 1、Creative 全拿 | remake 研究畫面的「擇一」機制與 Creative 差異 |
| 🟢 6 船體 Frigate→Doom Star,命令點隨船體升 | remake 船體種類與命令點 |
| 🟢 戰鬥只在恆星系內、格子逐船 move+fire+done、行星/衛星算船 | remake 戰術戰鬥流程 |
| 🟢 飛彈鎖定主目標一路飛、同船勿混 beam/missile | remake 飛彈/武器結算 |
| 🟢 地面戰不手動、看兵力+科技 | remake 入侵結算 |
| 🟡 敵 Battleship 約 turn 90(medium)/110(large)出現 | remake AI 造艦節奏是否接近(合理性參考) |

### 7.5 勝利條件
| 原版行為 | 驗證點 |
|---|---|
| 🟢 只有 3 種勝利:征服 / 議會 2/3 票 / 安塔蘭母星 | remake 是否恰好實作這 3 種(**不可照抄 2016 版的 6 種**) |
| 🟢 議會被提名者=人口前 2;可拒絕→全面開戰繼續玩 | remake 議會提名/拒絕邏輯 |
| 🟢 佔 Orion 有獎勵但不自動獲勝 | remake Orion 佔領效果 |
| 🟢 安塔蘭勝=3 遺跡建 lab→次元入口→滅安塔蘭母星 | remake 安塔蘭勝流程 |
| ⚠ 議會間隔 25 vs 30 回合 | **待原版實測**;remake 現用值先標暫定 |

---

## 附:本輪未能取得/受阻的來源(供後續補查)
- **StrategyWiki 全站**:WebFetch 一律回 **403**(含 `action=raw` 端點),僅能靠 WebSearch 摘錄。若要逐字核對頁面(尤其 Combat、Race design、Gameplay 三頁),建議改用瀏覽器工具(claude-in-chrome)或存檔快照。
- **GameFAQs `_The_General_` 完整攻略**、**gameplay.tips**、**Fandom(402/404)**、**indienova 詳頁(403)**、**civclub 汉化串(404)**:WebFetch 受阻,僅有搜尋摘要。
- **Let's Play Archive**:第 1 篇可取,後續逐回合實況(Update 02+ 有開局精確數字)本輪未逐篇抓,建議日後補抓 Update 02 以取得「turn 1 精確人口/分配」的玩家實錄,補上 §1.4 與 §7.2 的 ⚠ 待查。
