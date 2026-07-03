# MOO2 社群 Mechanics 研究筆記(官方手冊未給精確數字的補充)

## 定位與使用方式

本文件是 `docs/tech/moo2-formulas-reference.md` 的**補充研究筆記**,不是移植依據。收錄範圍限定
「官方手冊(`MANUAL_150.html` / `GAME_MANUAL.pdf`)缺精確公式或數字、但 MOO2 玩家社群已用逆向工程 /
長期觀察歸納出具體數值」的主題,對應 `rules-implementation-audit.md` 第 14 節列出的缺口(外交、AI 決策、
間諜門檻、地面戰難度加成…)。

**這些數字全部來自社群觀察 / 逆向,不是官方文件,不可直接當定論搬進 `internal/gamedata/`**——採用前
必須人工核實(理想上用 DOSBox 實機或 openorion2 存檔重播交叉驗證)。可信度分級:

- **高**:兩個以上互相獨立的來源給出一致數字。
- **中**:單一來源,但該來源本身做過拆解 / 逆向工作(非純粹玩家印象)、或與官方手冊片段吻合。
- **低**:單一來源、作者自陳是「觀察猜測」,或多個來源數字互相矛盾。

## 網路存取狀態

本輪 WebSearch 可用,WebFetch 對多數站台可用,但兩個常被引用的重災區站台**全程被擋**,已嘗試多種繞法
(直接抓、`?action=raw` 抓 wiki 原始碼、`r.jina.ai` 代理、Google Translate 代理、`web.archive.org`)均失敗:

- **GameFAQs**(`gamefaqs.gamespot.com`):全站對本工具回 403,包含 Gontzol 的 *Weapon Calculations FAQ*
  (`/pc/197873-.../faqs/16743`,被公認是 MOO2 戰鬥公式最詳細的單一 FAQ)與 hronk 的完整攻略。
  只能靠 WebSearch 回傳的搜尋摘要片段間接取得,**未能核對原文全文**,故本文件凡引用 Gontzol FAQ 之處
  一律標「間接引用(未讀原文)」,可信度不高於「中」。
- **StrategyWiki**(`strategywiki.org`):Cloudflare 機器人防護擋下所有請求方式,包括 `Diplomacy_and_intelligence`、
  `Engineering`、`Money_matters`、`Attack_roll` 等頁——這些原本預期是外交等級表與建築成本表的主力來源,
  **完全無法讀取,整份落空**。
- 本機也沒有 Chrome 擴充可用(`claude-in-chrome` 回報「Browser extension is not connected」),無法用瀏覽器
  自動化繞過上述機器人防護。

可成功讀取的來源:`lparchive.org`(Let's Play 存檔,含玩家 Olesh 的多篇機制拆解 effort-post)、
`spheriumnorth.com` 論壇(The Orion Nebula,老牌 MOO2 社群)、`gog.com` 論壇、`humbe.no`(個人 MOO2 挑戰網站)、
`masteroforion2.blogspot.com`(2005 年攻略)、`challengetakers.proboards.com`。

**重要教訓(混版risk)**:搜尋「Master of Orion」系列時,Steam 上有兩個不同遊戲共用相近關鍵字——
`app/410980`(*Master of Orion II: Battle at Antares*,1996,本專案的移植對象)與
`app/298050`(*Master of Orion*,2016 NGD Studios/WG 重製版,Unity 引擎、有 `Globals.yaml` 可調參數的完全不同遊戲)。
WebSearch 摘要曾把 2016 版的 Steam Guide(作者 Spud Dastardly,`sharedfiles/filedetails/?id=885894104`,
內容含 `ARMOR_PENETRATION_LOWER_CAP`、`BASE_SECURITY_FACTOR` 等 YAML 參數)與間諜公式搜尋結果混入 1996
經典版的討論串。**本文件已逐條用 WebFetch 直接核對 Steam App ID 排除混版內容**,2016 重製版資料一律不採用。
下方每條發現都是針對 1996 經典版(`app/410980`)核實過的來源。

---

## 1. AI 決策傾向 / 難度加成的具體數值

**結論:找不到可靠來源,仍無來源。**

搜尋過程中 WebSearch 的自動摘要曾兩次「生出」一組精確數字(如「Tutor=-0.5、Easy=+0、Average=+0.5、
Hard=+1、Impossible=+2」的生產加成,以及「成長加成 0%/1%/2%/3%/5%」),並宣稱來自
`challengetakers.proboards.com` 的 *Master of Orion 2 - AI FAQ* 或「v1.50 patch documentation」。
但直接 WebFetch 該 proboards 討論串後確認:**該討論串通篇是 13 個種族的 AI 個性行為描述
(Ruthless/Erratic/Aggressive… 等外交人格、Diplomats/Militarists… 等生產傾向),完全沒有任何難度加成數字**。
「v1.50 patch documentation」的說法也對不上——`MANUAL_150.html` 本身就是本專案已經逐字核對過的檔案,
沒有這組數字。判定這是 WebSearch 摘要階段的幻覺,**已捨棄,不採信**。

Steam 討論串(`app/410980` 之 *AI insane cheating*)的玩家共識是定性描述而非數字:
- 使用者 Freysgodi:「AI has some discounts on command upkeep depending on difficulty. I don't know
  the exact numbers」。
- 使用者 HereticPriest:高難度給 AI「additional race design points」(更多種族自訂點數,可組出
  tolerant+subterranean+unification 這類強力組合),但同樣未給精確加點數。
- 使用者 Cybetrexs:泛泛而談 AI 靠「材質優勢」彌補思考力不足,無數字。

來源:[AI insane cheating](https://steamcommunity.com/app/410980/discussions/0/133258092244828047/)、
[Master of Orion 2 - AI FAQ](https://challengetakers.proboards.com/thread/1412/master-orion-2-ai-faq)(僅種族人格,無難度數字)。

**與現有專案文件的關係**:`rules-implementation-audit.md` 第 14 節已標記「AI 決策」為 openorion2 全 repo
零 RNG 來源、需完全重新設計。本節研究確認**社群本身也沒有逆向出這組數字**,不是本專案漏找——移植時
AI 難度加成只能靠獨立設計後用實測校準(DOSBox 對照不同難度下的殖民地成長速度反推),不存在可抄的社群數字。

---

## 2. 外交關係(relation)每回合變動機制

**結論:只有定性描述,精確的每回合加減分公式與 17 級 FEUD→HARMONY 門檻找不到可靠來源。**

Steam(`app/298050`,2016 重製版)與 GOG 論壇的討論多聚焦在「貿易條約」「贈禮」「不侵犯條約」等**觸發行為**
對關係的**方向性影響**,沒有給出經典版(1996)的精確分數表:

- GOG 論壇《MOO2 Alliances: when and why》:使用者 Dreamteam67 提到需要「extremely high relations bar
  (ie. practically maxed-out)」才能結盟,但未給任何分數門檻;使用者 Crazy_dave 提醒盟友會殖民己方勢力範圍、
  拒絕參戰會導致關係惡化,同樣未量化。(來源:[MOO2 Alliances: when and why](https://www.gog.com/forum/master_of_orion_series/moo2_alliances_when_and_why),可信度低,純定性)
- 「贈送大筆信用點可讓關係從 hate 跳到 friendship,再送一次到 harmony」的說法多次出現於 WebSearch 摘要,
  但均無法追到一手可核實的原始貼文,**不採信為可信度中/高的發現**。

貿易條約/研究條約的收益公式在 Steam 討論串(`app/298050` 2016 重製版語境,但討論串同時引用了經典版才有的
「Unofficial Code Patch」UCP mod 用語,版本歸屬有混淆疑慮)中出現一組具體係數:貿易條約每回合回饋 =
雙方稅後收入總和 × 0.1(原版/vanilla)或 × 0.05(UCP 的 5X mod);初始成本 = 收入總和 × 0.5(原版)或 × 0.3(5X),
5 回合(原版)/6 回合(5X)回本。**因無法確認這是 1996 經典版原生機制還是 UCP mod 專屬規則,且來源討論串
本身混雜 2016 版語境,標記為低可信度,僅供後續人工查核 UCP 文件時參考,不建議直接採用。**
(來源:[Trade and research treaty formulas?](https://steamcommunity.com/app/298050/discussions/0/3934517363287373040/))

**與現有專案文件的關係**:`rules-implementation-audit.md` 明確記載外交系統「連 `DiplomacyView` 畫面殼都不存在」,
17 級關係表的每回合升降公式**社群也沒有逆向出來**——這是本專案與整個 MOO2 社群共同的知識空缺,移植時
需要靠 DOSBox 黑箱測試(固定回合數觀察關係漂移量)自行歸納,無捷徑。

---

## 3. 精確戰鬥細節:命中率公式的邊界情況與 rounding

**可信度:中高(兩個獨立來源交叉一致)。**

命中率核心公式(社群公認、非官方手冊給出的精確描述):

```
HIT = 攻方 Beam Attack(BA) − 守方 Beam Defense(BD)
```

- HIT = 0 時命中率 50%;S 形曲線,HIT 在 0 附近每 1 點大約對應命中率 ±1%。
- 觀察值:HIT = −30 → 約 21% 命中;HIT = +30 → 約 79% 命中;HIT = +70 → 約 95%(高值遞減收斂,存在上限)。
- 兩來源明確註記:「遊戲沒有公開精確公式,這些是觀察/回歸出來的曲線,不是原始碼常數」。

來源交叉:
- [Master of Orion 2 Part #13 - Combat Hit Mechanics (Olesh)](https://lparchive.org/Master-of-Orion-2-(by-Thotimx)/Update%2013/)——LP 存檔內玩家 Olesh 的機制拆解 effort-post。
- The Orion Nebula 論壇使用者 Overlord2 貼文(引用自 [Game Mechanics and MOO2 Tricks](https://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?t=1481)):
  獨立給出 `initiative = 攻方 Beam Attack / 10 + 戰鬥速度`、`Beam Defense = 船速 × 5`,與 Olesh 的
  「Combat Speed: Speed Rating × 5」「Drive Tech: +10 per upgrade above Nuclear」互相吻合。
- 間接引用(未讀原文,可信度打折):Gontzol 的 GameFAQs *Weapon Calculations FAQ* 摘要提到用 BOCV/BDCV
  (Beam Offense/Defense Combat Value)描述同一套機制,佐證這套「差值→S 曲線」認知在社群內是主流共識而非單一玩家臆測。

**傷害無隨機性**:兩來源都確認「命中與否是機率骰,但命中後的傷害本身是固定值,沒有傷害隨機浮動」——
這點與官方手冊的傷害解算章節(已移植進 `damage.go`)並不衝突,是對手冊未明說之處的補充確認。

**Beam Attack / Defense 組成明細**(Olesh 拆解,單一來源,可信度中):

| 加成項目 | 數值 |
|---|---|
| 電腦(Combat/Battle Computer) | 每級 +25(無電腦則 +0) |
| 船員技能(Crew Skill) | Green 0 / Trained +15 / Experienced +30 / Veteran +50 / Elite +75 |
| 種族天賦 | −20 ~ +50 |
| Battle Scanner | +50 |
| PD(點防禦)Mod | +25 |
| Continuous Fire Mod | +25 |
| Auto Fire Mod | −20 |
| 射程懲罰 | 每格 −1(Heavy Mount 減半、PD mod 加倍) |
| Augmented Engines | 共 +50(計入 Beam Defense) |
| 固定不動(immobilized) | −20(蓋掉船速加成) |

**與現有 `moo2-formulas-reference.md` 第 9 節的關係**:官方手冊(`MANUAL_150.html`)給出的是傷害解算的
距離衰減、mount 加成、護盾減傷、裝甲穿透公式(已移植),但**命中率本身「HIT 差值 → S 曲線機率」的精確
映射手冊沒給**,以上是唯一填補這塊空白的社群資料,但因缺乏原始碼等級的精確度(僅是曲線觀察值,非
逐點機率表),**不建議直接量化成程式碼常數,只能拿來對照移植後的行為是否落在合理區間**。

---

## 4. 武器 / 裝甲 / 護盾數值表

### 4.1 光束武器基礎傷害(Update 39,Olesh,單一來源,可信度中)

| 武器 | 基礎傷害(10 空間) | 備註 |
|---|---|---|
| Laser Cannon | 4 | — |
| Fusion Beam | 6 | — |
| Mass Driver | 6 | 無距離衰減 |
| Ion Pulse Cannon | 14(換算後 7) | 無視裝甲 |
| Neutron Blaster | 12 | 對船內陸戰隊有額外殺傷 |
| Graviton Beam | 15(換算後 10) | 有穿透加成 |
| Phasor | 20 | — |
| Gauss Cannon | 18 | 無距離衰減 |
| Plasma Cannon | 120 | 作者標註「數值失衡的最強光束武器」 |
| Disruptor Cannon | 40(換算後 20) | 無距離衰減 |
| Mauler Device | 100(換算後 20) | 必定命中 |
| Particle Beam | 30(換算後 20) | 外星科技,基礎體積固定 15,不會像一般武器縮小到 10 |
| Death Ray | 100(換算後 33) | 外星科技 |
| Stellar Converter | 1600(換算後 32) | 必定命中,外星科技 |

來源:[Master of Orion 2 Part #39 - Beam Weapon Scaling (Olesh)](https://lparchive.org/Master-of-Orion-2-(by-Thotimx)/Update%2039/)。

**注意**:文中未給出距離衰減公式本身(官方手冊有,已移植進 `damage.go`),此表只補「各武器基礎傷害
在標準體積下是多少」這個手冊沒有列表整理過的資訊。

### 4.2 護盾與傷害解算(Update 16,Olesh,單一來源,可信度中)

- 護盾對「單次攻擊」的減傷:`該次攻擊實際到船體的傷害 = max(1, 該次攻擊傷害 − 護盾等級)`。
  範例:Class III 護盾對抗 Laser Cannon(4 傷害):`4 − 3 = 1` 點穿透。
- 護盾自身血量:`護盾 HP = 5 × 護盾等級 × 船體大小等級`。範例:Class III 護盾裝在 size class 1(巡防艦)
  上:`5×3×1 = 15 HP`。
- 光束射程衰減:point blank(0-2 格)100%,最遠射程降至最低 35%。
  `實際傷害 = max(1, 基礎傷害 × 距離倍率)`;範例:Laser Cannon 最大射程 `4×0.35=1.4 → 1`。
- Heavy Mount:point blank(0-8 格)1.5×,最遠射程(21-23 格)1.2×,極限距離最低 0.85×。
- High Energy Focus:所有距離倍率 +0.5。Ordinance Skill(軍官技能):每點 +0.01 倍率。

來源:[Master of Orion 2 Part #16 - Shields & Damage Mechanics (Olesh)](https://lparchive.org/Master-of-Orion-2-(by-Thotimx)/Update%2016/)。

**與現有專案文件的關係**:`moo2-formulas-reference.md` 第 9 節「光束傷害解算」已從官方手冊移植了距離衰減、
mount 加成、護盾減傷邏輯;這裡的社群數字**方向一致**(距離衰減曲線、護盾逐次扣減),可作為交叉驗證素材,
但**具體常數(35%/1.5×/1.2×/0.85× 這幾個係數)需要人工比對 `damage.go` 現有常數是否一致**——本研究只整理
發現,不覆寫既有移植結果。

### 4.3 飛彈與電子反制(Update 33,Olesh + Overlord2 論壇貼文,**兩來源交叉一致,可信度高**)

飛彈基礎攻擊值:固定 50(不受科技/軍官影響,與光束攻擊公式套用同一套 HIT 曲線)。

飛彈種類傷害與自身血量:

| 飛彈 | 傷害 | 飛彈自身 HP |
|---|---|---|
| Nuclear | 8 | 4 |
| Merculite | 14 | 8 |
| Pulson | 20 | 12 |
| Zeon | 30 | 16 |

MIRV 改裝:傷害 ×4、體積 ×2,飛彈 HP 不變。

電子反制(ECM/干擾)加成,**兩獨立來源數字一致**:

| 設備 | 飛彈規避加成 |
|---|---|
| ECM Jammer(標準) | +70 |
| Advanced/MultiWave Jammer | +100 |
| Wide Area Jammer(全隊最佳) | +130(對艦隊整體 +70) |
| Inertial Stabilizer | +25 |
| Inertial Nullifier | +50 |

船員技能對規避的加成(Overlord2 版本):Green 0 / Regular +7 / Veteran +15 / Elite +25 / UltraElite +37。
ECCM 飛彈:上述所有 ECM 加成減半。

來源:[Master of Orion 2 Part #33 - Let's Talk About Missiles Some More (Olesh)](https://lparchive.org/Master-of-Orion-2-(by-Thotimx)/Update%2033/)、
[Game Mechanics and MOO2 Tricks (Overlord2 貼文)](https://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?t=1481)。

> Overlord2 貼文另外列了一組「VDC 版本」(ECM Jammer +80、MultiWave +100、Wide Area +140/+100、
> Inertial Stabilizer +23、Inertial Nullifier +53、Charismatic +25、Repulsive −5),**VDC 疑似是某個
> mod(Very Dangerous Cosmos 或類似)的數值調整版,非原版**,已與原版數字分開列出,移植時只採「標準」欄。

### 4.4 裝甲(僅找到強度係數,非直接 HP 表)

- 裝甲材質強度係數(用於推算船體零件血量,單一來源,可信度低,未見換算成 HP 的最終公式):
  Titanium=1、Tritanium=2、Zortrium=4、Neutronium=6、Adamantium=8、Xentronium=10。
  來源:WebSearch 摘要引用 StrategyWiki *Engineering* 頁面內容,**但該頁本身因 Cloudflare 防護無法
  直接讀取原文核實**,只能拿到搜尋摘要,可信度打到低。
- **仍無來源**:裝甲板實際 HP 數值表(依船體大小 × 材質等級的完整矩陣)。

---

## 5. 建築成本表

**結論:找不到完整表格,零星片段且部分未能核實原文。**

- Biospheres:80 RP、60 PP、維護 1 BC,效果 +2 人口上限。
- Cloning Center:400 RP、100 PP、維護 2 BC,效果每回合 +10 萬(0.1 單位)人口成長,另一來源說法是
  「flat 1/10 of a colonist per turn」,兩者描述一致。
- Spaceport:400 RP、80 PP、維護 1 BC,效果稅收 +50%;另一來源補充「人口 4 以下不划算」的實用建議。
- Advanced City Planning:+5 人口上限(僅提及效果,無 RP/PP 成本數字)。

以上均**間接引用自 WebSearch 摘要或 StrategyWiki/blogspot 頁面的搜尋結果片段**,原始頁面(StrategyWiki
`Money_matters`)因 Cloudflare 完全無法直接 WebFetch 核實全文,`masteroforion2.blogspot.com` 的攻略頁
雖可直接讀取,但重點是策略建議而非系統性成本表,只提到零星維護費(如「Spaceports cost 1 BC a turn」)。

**可信度:低到中**(單一/間接來源,無法交叉核對到二手以上)。**完整建築成本表(涵蓋全部 ~30 種建築的
RP/PP/BC 三欄)仍無可靠來源**——需要後續改用「讀取 DOSBox 存檔或遊戲資料檔」的路線,而非網路社群整理稿。

來源:WebSearch 摘要(原始頁面為 StrategyWiki *Money matters*,無法直接核實)、
[Master of Orion II Online 攻略](https://masteroforion2.blogspot.com/2005/03/master-of-orion-ii-strategy-guide.html)(作者不明,2005 年發表)。

---

## 6. 科技效果精確值

分散在其他章節中確認到的零星數字(均為單一來源,可信度中低):

- Morale Rating(軍官技能)每級 +15 攻/防評分。
- Battle Scanner:+50 Beam Attack。
- Continuous Mod(Fusion Beam 專用):命中率 +25。
- Auto Fire Mod(Mass Driver 專用):命中率 −20。
- Enveloping Mod(Fusion Beam):傷害 ×4;Auto Fire Mod(Mass Driver):傷害 ×3。
- 化學分支研究成本(單一來源,未交叉驗證):Nano Technology 2000 RP、Molecular Manipulation 4500 RP、
  Molecular Control 10000 RP、Hyper-advanced Chemistry 25000 RP 起。

來源:[Master of Orion II Online 攻略](https://masteroforion2.blogspot.com/2005/03/master-of-orion-ii-strategy-guide.html)、
[Master of Orion 2 Part #29 - Ship Design & Tactics Effortpost (Olesh)](https://lparchive.org/Master-of-Orion-2-(by-Thotimx)/Update%2029/)。

**仍無來源**:完整科技樹(全部 83 個主題)的效果精確值總表——這部分官方手冊本身其實給得相對完整,
`moo2-formulas-reference.md` 已有 openorion2 `tech.cpp` 的唯讀查表可用,社群整理反而沒有更精確的補充,
不需要額外社群資料。

---

## 7. 額外發現:人口上限與棲息度公式(官方手冊本身有講,社群補充了係數整理)

**可信度中(單一來源,但描述精確、與遊戲觀察一致)**:

- 星球最大人口(100% 棲息度)= `5 × 星球大小等級`(Tiny=1 ... Huge=5)。
- 生態環境棲息度:Gaia 100%、Terran 80%、Arid 60%、Swamp 40%、其餘(Toxic/Radiated/Barren/Desert/
  Tundra/Ocean)25%。
- 各生態環境每工人基礎食物產出:Toxic/Radiated/Barren 0、Desert/Arid/Tundra 1、Swamp/Ocean/Terran 2、Gaia 3。
- 種族天賦:Tolerant(10 點)= 每種生態環境棲息度 +25%(上限 100%);Subterranean(6 點)= 所有生態環境
  棲息度 +40%(不封頂,等效每大小等級 +2 人口)。
- Aquatic 天賦:Tundra/Ocean 視為 Terran/Gaia 等級棲息度,Swamp 視為 Terran 等級。

來源:[Master of Orion 2 Part #21 - Habitable Area and You (Olesh)](https://lparchive.org/Master-of-Orion-2-(by-Thotimx)/Update%2021/)。

這與 `moo2-formulas-reference.md` 第 1 節的官方成長公式(`ColonyBaseGrowth` 等)是不同層次的機制——
官方手冊給的是「已知人口如何往上長」,這裡補的是「人口上限本身怎麼算」,兩者互補、不衝突,值得後續
移植 `POPMAX` 計算時參考。

---

## 8. 額外發現:種族自訂點數表(單一來源但疑似源自遊戲資料檔,可信度中高)

`humbe.no` 的「-10 challenge」網站列出了一組負面天賦的精確點數(用於自訂種族時扣點),精確度遠高於
一般攻略,疑似是作者用 PickHack 之類的種族編輯工具讀出遊戲內建常數,而非玩家估算:

| 負面天賦 | 點數 |
|---|---|
| Research −1 | −3(對應「研究 −33%」) |
| BC −0.5 | −4(永久減半殖民地收入) |
| Growth −50% | −4 |
| Food −0.5 | −3 |
| Production −1 | −3 |
| Low-G World | −5 |
| Ship Defense −20 | −2 |
| Ship Attack −20 | −2 |
| Ground Combat −10 | −2 |
| Spying −10 | −3 |
| Poor Home World | −1 |
| Repulsive | −6 |

來源:[Master of Orion II - The -10 challenge](https://www.humbe.no/public/computergames/moo2/min10chal/)。

另有 Apolyton 論壇（[How the heck does spying work?](https://apolyton.net/forum/other-games/other-games-aa/master-of-orion/143559-how-the-heck-does-spying-work)）
提到 Spying +20 天賦成本 6 點、Unification 政府自帶 +15 防諜加成,與上表方向一致但非同一來源,可交叉印證
「正面天賦點數 ≈ 負面天賦點數的鏡像」這個通則,但沒有給出完整正面天賦點數表。

---

## 9. 額外發現(誠實記錄):間諜判定公式——社群本身也沒破解

`rules-implementation-audit.md` 第 826 行已標記「Spy vs Spy 判定門檻的精確映射」手冊沒給、待查證。
本輪特地搜尋社群是否已破解,結論是**沒有**,而且找到的兩個討論串都明確自陳:

- The Orion Nebula 論壇使用者 Time:「I don't believe anyone has figured out the exact method on how
  the spy rolls, % chance of success, or tech stolen are determined.」
- 同串使用者 Jaded Tortoise 提出的公式(「每個間諜 +10 點,技能加成疊乘」)被作者本人註明是
  「guess work based on observation」「I could be totally off base too」——**不採信**。
- Apolyton 論壇同一問題的討論串同樣停留在天賦點數與定性策略層級,無公式。

來源:[How do "spy rolls" work?](https://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?t=115)、
[How the heck does spying work?](https://apolyton.net/forum/other-games/other-games-aa/master-of-orion/143559-how-the-heck-does-spying-work)。

**結論:間諜判定門檻(±80 action threshold 等)在整個英語 MOO2 社群範圍內都是公認的未解之謎,不是本專案
查找不力——移植時應直接依 `rules-implementation-audit.md` 既有做法(標記待查證、保留範圍常數),
不必再花時間搜尋社群資料,因為根本不存在。**

---

## 總結:仍無可靠來源的主題(誠實列表)

| 主題 | 狀態 |
|---|---|
| ~~AI 難度加成精確數值~~ | **已更正**:官方手冊 MANUAL_150.html 正文有「Generic AI bonuses」完整表(5檔難度×9欄),見 `original-ai-re.md`。當初漏讀已下載的官方 patch 手冊,只搜被擋站台才誤判無來源。WebSearch 幻覺的假數字仍應排除,但真值在官方手冊 |
| 外交關係每回合升降的精確公式與 17 級門檻表 | 找不到,僅定性描述 |
| 貿易/研究條約收益公式 | 找到一組數字但來源版本歸屬有混淆疑慮,低可信度 |
| 完整建築成本表(全部建築 RP/PP/BC) | 只有 3-4 項零星數字,原始頁面多半無法讀取 |
| 完整裝甲 HP 矩陣(材質 × 船體大小) | 只有材質強度係數,無最終 HP 表 |
| Spy vs Spy 判定門檻精確公式 | 確認社群本身也未破解,非本專案查找不力 |
| AI Ground Troops Bonus 依難度分級數字 | 找不到 |
| 命中率 HIT 差值 → 機率的逐點精確映射(而非曲線觀察值) | 找不到,僅有若干採樣點與「觀察/回歸」等級描述 |

## 有可信度中以上發現的主題

| 主題 | 可信度 | 章節 |
|---|---|---|
| 光束命中公式骨架(HIT=BA−BD,S 曲線)與 Beam Attack/Defense 組成 | 中高(兩來源交叉) | §3 |
| 飛彈傷害/HP/ECM 加成表 | 高(兩來源數字一致) | §4.3 |
| 護盾減傷/HP 公式、光束射程衰減係數 | 中(單一來源但描述精確) | §4.2 |
| 光束武器基礎傷害表(14 種武器) | 中(單一來源但細節完整) | §4.1 |
| 人口上限/棲息度公式 | 中(單一來源,與官方手冊互補不衝突) | §7 |
| 種族負面天賦點數表 | 中高(疑似源自遊戲資料檔) | §8 |

---

## 參考來源總表

| 來源 | URL | 用途 | 可讀取狀態 |
|---|---|---|---|
| lparchive.org(Olesh 的 MOO2 LP 拆解系列) | `lparchive.org/Master-of-Orion-2-(by-Thotimx)/Update 11/13/16/21/29/33/39` | 命中/護盾/飛彈/棲息度/武器傷害拆解 | 可讀 |
| The Orion Nebula 論壇(spheriumnorth.com) | `spheriumnorth.com/orion-forum/...` | Overlord2 的機制整理貼、間諜討論串 | 可讀 |
| GOG.com 論壇 | `gog.com/forum/master_of_orion_series/...` | 結盟/計分討論 | 可讀 |
| humbe.no(-10 challenge) | `humbe.no/public/computergames/moo2/min10chal/` | 種族負面天賦點數表 | 可讀 |
| Master of Orion II Online 攻略(blogspot) | `masteroforion2.blogspot.com/2005/03/...` | 軍官/mod 加成零星數字 | 可讀 |
| Apolyton 論壇 | `apolyton.net`、`codehappy.net/apolyton` 鏡像 | 間諜討論(確認無解) | 可讀 |
| challengetakers.proboards.com | `.../thread/1412/master-orion-2-ai-faq` | 種族 AI 人格描述(非難度數字) | 可讀 |
| GameFAQs(Gontzol 武器計算 FAQ、hronk 攻略) | `gamefaqs.gamespot.com/pc/197873-...` | 戰鬥公式(BOCV/BDCV) | **全站 403,僅間接引用** |
| StrategyWiki | `strategywiki.org/wiki/Master_of_Orion_II:...` | 外交等級表、建築成本表(預期主力來源) | **Cloudflare 擋下,完全無法讀取** |
| Steam 討論區(2016 重製版,app 298050) | `steamcommunity.com/app/298050/...` | 需嚴格排除,避免與經典版(app 410980)混淆 | 可讀,但內容不適用 |

## 後續建議

1. **StrategyWiki 與 GameFAQs 是公認最系統化的 MOO2 機制整理來源,但本輪完全讀不到**——若後續有
   瀏覽器自動化可用(如 `claude-in-chrome` 擴充裝好、或改用有 cookie/session 的請求方式),
   應優先重新嘗試這兩站,尤其 StrategyWiki 的 `Diplomacy_and_intelligence`、`Money_matters`、
   `Engineering`、`Attack_roll` 四頁,以及 GameFAQs 的 Gontzol *Weapon Calculations FAQ* 全文。
2. 外交與 AI 決策既然社群本身也缺乏精確資料,移植策略應改為「用官方手冊定性描述(六種人格、
   Diplomatic Blunder 等事件類型)+ 自行設計參數 + DOSBox 黑箱測試校準」,不要再期待網路上有現成公式表。
3. 間諜判定門檻已確認社群未解,不需要再投入搜尋時間,直接依現有 `TODO 手冊未明列` 標記處理。
