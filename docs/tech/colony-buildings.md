# 殖民地建築全表(手冊萃取)

> 目的:彙整 *Master of Orion II* 所有可建的殖民地建築(含軌道衛星),供 remake 建築系統(建造/維護/效果)
> 實作參考。日期:2026-07-10。
>
> **資料結構說明(第一性原理:先搞懂手冊怎麼分類,再抽表,不要用自己的分類套上去)**:`GAME_MANUAL.pdf`
> p.75-76「The Big List」把研究樹的每一項應用(application)標成五種型別之一:
> **Building**(殖民地建築,含建造成本+維護費)、**System**(艦載模組)、**Ship**(船體/引擎類科技)、
> **Special**(一次性/特殊行動,如 Colony Base、Terraforming)、**Satellite**(軌道衛星防禦設施)、
> **Achievement**(研究完成自動生效的帝國全局加成,**不需建造**,見下方「不列入本表」說明)、**Android**、
> **Equipment**(陸戰隊裝備,無建造/維護成本)。本表只收錄 **Building**(35 項)與 **Satellite**(5 項,
> 因為 Star Base/Battlestation/Star Fortress/Artemis System Net/Dimensional Portal 都是透過殖民地建造隊列
> 建造、有維護費的常駐設施,性質等同「軌道建築」),共 **40 項**,逐項附手冊頁碼與 `openorion2` 研究成本
> 交叉驗證。另有 1 項 Stellar Converter 標記為混合型別 `(Building/System)`,獨立列在二之一節,不計入
> 40 項總數(避免與型別統計混淆)。
>
> **不列入本表的型別**:
> - **Achievement**(如 Nano Disassemblers、Microlite Construction、Advanced City Planning):研究完成即
>   自動生效,不出現在建造隊列,不是「建築」。已在 `moo2-formulas-reference.md` §2(生產/污染)等處引用。
> - **Special**(如 Colony Base、Terraforming、Gaia Transformation、Soil Enrichment):是一次性行動或地形
>   轉換,不是常駐建築(建完就消耗掉,或者是套用在星球本身的狀態轉換,不是可維護的建物)。
> - **Capitol**:每個首都殖民地自動擁有、不計入建築格位上限,詳見 `homeworld-init.md` §3.2,性質特殊,
>   不放進本表。

## 一、資料來源與可信度分級

| 來源 | 給出的資料 | 可信度 |
|---|---|---|
| `moo2_patch1.5/GAME_MANUAL.pdf` p.75-112「The Big List」 | 每項建築的**效果敘述**(逐字)、**維護費**(BC/turn,幾乎每項都有)、互斥/取代關係 | 高(一手官方手冊,逐句核對) |
| `openorion2/src/tech.cpp:169-266`(`research_choices[]`,唯讀靜態表) | 每項建築所屬**研究主題的 RP 成本**(該建築與同主題其他科技共用同一 RP 值) | 高(GPL 原始碼靜態表,無隨機性) |
| 手冊「Buildings」章節本文 vs `game_manual.txt` 逐項比對(本文件實際解析方式) | **前置研究欄位名稱**(如 Robotics、Astro Engineering):由每個建築條目**前面最近一個獨立欄位標題行**取得,已與 `tech.cpp` 的 `techtree[8][14]` 陣列欄位命名逐一核對,**全數相符**(見下方驗證註記) | 高(結構化解析,非人工判讀) |
| `moo2_patch1.5/MANUAL_150.html`「Modding with Config → Buildings」 | **建造成本(PP)**:全手冊唯一一處给出具体 PP 数字,是作为 modding 语法范例(「Armor Barracks 的默认成本是 150 PP」),**不是完整表** | 中(單一範例,非系統性列表) |
| `community-mechanics-findings.md` §5 | 3-4 项零星 PP/BC 成本(Biospheres/Cloning Center/Spaceport)| 低(單一/間接網路來源,未核實原文)|

**⚠ 誠實聲明(建造成本 PP/BC 缺口)**:手冊敘述性內文**只給維護費(BC/turn),幾乎不給建造成本(PP)**——
這不是本次萃取遺漏,是通盤搜尋 `GAME_MANUAL.pdf` 全文「cost」「PP」關键字後確認的結論(見
`moo2-formulas-reference.md` §5 地形改造小節、`community-mechanics-findings.md` §5 已有相同結論)。
唯一的例外是 `MANUAL_150.html` 為了示範 modding 語法舉的「Armor Barracks = 150 PP + 2 BC/T 維護」這一個
具體數字(與本表的 GAME_MANUAL.pdf 維護費「2 BC」完全吻合,可信度高)。**其餘 34 項建築的 PP 建造成本
本文件不臆測**,一律標「待查證(需存檔/資料檔而非手冊文字)」,對應下方總表的「建造成本(PP)」欄。

## 二、殖民地建築(Building,35 項)

依 `GAME_MANUAL.pdf` 出現順序(= 研究樹分支順序:Construction → Power → Chemistry → Sociology →
Computers → Biology → Physics → Force Fields)排列。「研究成本」欄取自 `tech.cpp` 該建築所屬 RP 主題值,
同一研究主題內若有其他科技/建築同組,一併註明(代表玩家研究完那筆 RP 就同時解鎖該組全部項目)。

| # | 建築(英文) | 建議中譯 | 分類 | 前置研究欄位 | 研究成本(RP) | 維護費(BC/turn) | 建造成本(PP) | 效果 | 互斥/取代 | 頁碼 |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | Marine Barracks | 陸戰隊營房 | 防禦/陸戰 | Engineering(起始已知) | 50(與 Colony Base、Star Base 同組) | 1 | 待查證 | 建成立即產生最多 4 個陸戰隊單位;之後每 5 回合 +1,上限為「現有人口/2」與「星球人口上限/2」取較小值;特定政府下可消除士氣懲罰 | 無 | p.77 |
| 2 | Automated Factories | 自動化工廠 | 生產 | Advanced Construction | 150(與 Heavy Armor、Missile Base 同組) | 1 | 待查證 | 每個工業人口 +1 產能/回合,殖民地整體 +5 產能 | 無 | p.78 |
| 3 | Missile Base | 飛彈基地 | 防禦 | Advanced Construction | 150(與上同組) | 2 | 待查證 | 配備最佳飛彈(佔用 300 空間內盡量多),自動防禦來襲艦隊;只能被軌道轟炸摧毀 | 無 | p.78 |
| 4 | Armor Barracks | 裝甲營房 | 防禦/陸戰 | Astro Engineering | 400(與 Fighter Garrison、Spaceport 同組) | 2 | **150**(`MANUAL_150.html` 官方 modding 範例數字,唯一有 PP 來源的一項) | 建成立即產生最多 2 個裝甲營,之後每 5 回合 +1,上限為「現有人口/4」與「星球人口上限/4」取較小值;特定政府下可消除士氣懲罰 | 無 | p.79 |
| 5 | Spaceport | 太空港 | 貿易 | Astro Engineering | 400(與上同組) | 1 | 待查證 | 該殖民地所有來源的 BC 收入 +50% | 無 | p.79 |
| 6 | Fighter Garrison | 戰機基地 | 防禦 | Astro Engineering | 400(與上同組) | 2 | 待查證 | 依已解鎖的最高階戰機科技,可駐留 10 個攔截機中隊 / 6 個轟炸機中隊 / 4 個重型戰機中隊;每 10 回合全數整補;只能被軌道轟炸摧毀 | 無 | p.79 |
| 7 | Robo Mining Plant | 機器人採礦廠 | 生產 | Robotics | 650(與 Battlestation、Powered Armor 同組) | 2 | 待查證 | 每個工業人口 +2 產能,殖民地整體 +10 產能 | 無 | p.80 |
| 8 | Ground Batteries | 地面砲台 | 防禦 | Astro Construction | 1150(與 Battleoids、Titan Construction 同組) | 2 | 待查證 | 配備最佳光束武器的 Heavy Mount 與 Point Defense 版本(佔用 450 空間內盡量多);只能被軌道轟炸摧毀 | 無 | p.81 |
| 9 | Recyclotron | 再生反應爐 | 生產/環保 | Advanced Manufacturing | 1500(與 Automated Repair Unit、Planet Construction 同組) | 3 | 待查證 | 每單位人口(不論職業)額外產生 1 產能,且此產能不計入污染;1.50i 起可將 Toxic 星球轉為 Barren(`terraform.go` 已移植,見 `moo2-formulas-reference.md` §5) | 無 | p.81 |
| 10 | Robotic Factory | 機器人工廠 | 生產 | Advanced Robotics | 2000(與 Bomber Bays 同組) | 3 | 待查證 | 依礦產豐度加成:Ultra Poor +5、Poor +8、Abundant +10、Rich +15、Ultra Rich +20 | 無 | p.82 |
| 11 | Deep Core Mine | 深層核心礦場 | 生產 | Tectonic Engineering | 3500(與 Core Waste Dumps 同組) | 3 | 待查證 | 每個工人 +3 產能,殖民地整體 +15 產能 | 無 | p.82 |
| 12 | Core Waste Dumps | 核心廢料場 | 環保 | Tectonic Engineering | 3500(與 Deep Core Mine 同組) | 8 | 待查證 | 完全消除星球污染;建成時**取代** Pollution Processor 與 Atmospheric Renewer(以全額建造成本回收出售,非通常的半額) | **取代** #33 Pollution Processor + #34 Atmospheric Renewer | p.82 |
| 13 | Food Replicators | 食物複製機 | 食物 | Matter-Energy Conversion | 2750(與 Transporters 同組) | 10 | 待查證 | 可依需求將工業產能以 2:1 轉換成食物,每單位食物花費 1 BC | 無 | p.87 |
| 14 | Pollution Processor | 污染處理器 | 環保 | Advanced Chemistry | 650(與 Merculite Missile 同組) | 1 | 待查證 | 可處理殖民地一半產能對應的污染,按比例降低污染量(手冊:「process the waste from fully half of the colony's production」) | 與 Atmospheric Renewer 效果疊加(合計 1/8 產能致污染);被 Core Waste Dump 取代 | p.90 |
| 15 | Atmospheric Renewer | 大氣更新器 | 環保 | Molecular Compression | 1150(與 Iridium Fuel Cells、Pulson Missile 同組) | 3 | 待查證 | 消除四分之三工業產能造成的污染;與 Pollution Processor 同時存在時,合計只剩 1/8 產能致污染 | 與 Pollution Processor 疊加;被 Core Waste Dump 取代 | p.90 |
| 16 | Space Academy | 太空學院 | 軍事 | Military Tactics | 150(單一項目) | 2 | 待查證 | 此殖民地建造的艦艇船員起始等級 +1 級(Green→Regular 等);同星系內每有一座 Space Academy,所有駐留艦艇船員每回合額外 +1 經驗 | 無 | p.92 |
| 17 | Alien Management Center | 異族管理中心 | 社會 | Xeno Relations | 650(與 Xeno Psychology 同組) | 1 | 待查證 | 每 2 回合同化 1 單位被征服人口(不論政府);Charismatic/Repulsive 種族天賦會調整此基礎速率;消除多種族殖民地 -20% 士氣懲罰,並使未同化人口的叛亂機率減半 | 無 | p.92 |
| 18 | Planetary Stock Exchange | 行星證券交易所 | 貿易 | Macro Economics | 1150(單一項目) | 2 | 待查證 | 該殖民地收入 +100% | 無 | p.93 |
| 19 | Astro University | 太空大學 | 科研/生產 | Teaching Methods | 2000(單一項目) | 4 | 待查證 | 每單位受教育人口(農/工/科)額外 +1 對應產出(食物/產能/研究皆適用) | 無 | p.93 |
| 20 | Research Laboratory | 研究實驗室 | 科研 | Optronics | 150(與 Dauntless Guidance System、Optronic Computer 同組) | 1 | 待查證 | 每個科學家人口 +1 研究點;另自動產生 5 研究點 | 無 | p.94 |
| 21 | Planetary Supercomputer | 行星超級電腦 | 科研 | Positronics | 900(與 Holo Simulator、Positronic Computer 同組) | 2 | 待查證 | 每個科學家人口 +2 研究點,殖民地整體 +10 研究點 | 無 | p.95 |
| 22 | Holo Simulator | 全息模擬艙 | 士氣 | Positronics | 900(與上同組) | 1 | 待查證 | 殖民地士氣 +20% | 無 | p.96 |
| 23 | Autolab | 自動實驗室 | 科研 | Cybertronics | 2750(與 Cybertronic Computer、Structural Analyzer 同組) | 3 | 待查證 | 全自動產生 30 研究點/回合(不依賴人口) | 無 | p.96 |
| 24 | Galactic Cybernet | 銀河網路中心 | 科研 | Galactic Networking | 4500(與 Virtual Reality Network 同組) | 3 | 待查證 | 每個科學家人口 +3 研究點,殖民地整體 +15 研究點 | 無 | p.98 |
| 25 | Pleasure Dome | 歡樂穹頂 | 士氣 | Moleculartronics | 6000(與 Achilles Targeting Unit、Moleculartronic Computer 同組) | 3 | 待查證 | 殖民地士氣 +30% | 無 | p.98 |
| 26 | Hydroponic Farm | 水耕農場 | 食物 | Astro Biology | 80(與 Biospheres 同組) | 2 | 待查證 | 殖民地食物產出 +2 | 無 | p.99 |
| 27 | Biospheres | 生態圈 | 居住 | Astro Biology | 80(與上同組) | 1 | 待查證(社群來源:60 PP,低可信度) | 星球人口上限 +2 單位 | 無 | p.99 |
| 28 | Cloning Center | 複製中心 | 食物/人口 | Advanced Biology | 400(與 Death Spores、Soil Enrichment 同組) | 2 | 待查證(社群來源:100 PP,低可信度) | 人口成長 +100,000(0.1 單位)/回合,直到達星球人口上限為止 | 無 | p.99 |
| 29 | Subterranean Farms | 地底農場 | 食物 | Macro Genetics | 1500(與 Weather Controller 同組) | 4 | 待查證 | 星球食物產出 +4 | 無 | p.100 |
| 30 | Weather Controller | 氣候控制器 | 食物 | Macro Genetics | 1500(與上同組) | 3 | 待查證 | 每個農業人口食物產出 +2 | 無 | p.100 |
| 31 | Planetary Gravity Generator | 行星重力產生器 | 居住 | Artificial Gravity | 1150(與 Graviton Beam、Tractor Beam 同組) | 2 | 待查證 | 將星球重力正常化至 Normal-G,消除 Low-G/Heavy-G 的負面效果 | 無 | p.104 |
| 32 | Planetary Radiation Shield | 行星輻射護盾 | 防禦/地形 | Magneto Gravitics | 900(與 Class III Shield、Warp Dissipater 同組) | 1 | 待查證 | Radiated 氣候星球維持 Barren 狀態;軌道轟炸傷害 -5 | 被 #34 Planetary Flux Shield 取代 | p.108 |
| 33 | Warp Field Interdictor | 曲速力場干擾器 | 防禦 | Warp Fields | 2000(與 Lightning Field、Pulsar 同組) | 3 | 待查證 | 星系內半徑 3 秒差距範圍,使敵方艦艇移動速度降為 1 秒差距/回合 | 無 | p.109 |
| 34 | Planetary Flux Shield | 行星通量護盾 | 防禦/地形 | Quantum Fields | 4500(與 Class VII Shield、Wide Area Jammer 同組) | 3 | 待查證 | Radiated 氣候轉 Barren;軌道轟炸傷害 -10;**取代**已存在的 Planetary Radiation Shield | 取代 #32;被 #35 Planetary Barrier Shield 取代 | p.111 |
| 35 | Planetary Barrier Shield | 行星屏障護盾 | 防禦/地形 | Temporal Fields | 15000(與 Class X Shield、Phasing Cloak 同組) | 5 | 待查證 | Radiated 氣候轉 Barren;軌道轟炸傷害 -20;生物武器無法進入大氣層 | 取代 #34 Planetary Flux Shield | p.112 |

編號 32/34/35(Planetary Radiation/Flux/Barrier Shield)是同一組「行星護盾三階」,逐級取代前一級
(手冊原文各自明載「replaces any Planetary Radiation Shield」等),為求完整仍各自列一行並標明取代關係。

> **本節共 35 項,逐一對應手冊「(Building)」型別標記的全部項目**(已用結構化解析對全文 182 個型別
> 標記逐一計數驗證,`grep -c "(Building)"` = 35,與本表列數一致)。**Dimensional Portal**(標記
> `(Satellite)`)與**Stellar Converter**(標記 `(Building/System)` 混合型別)**不計入本節**,分別見下方
> 「三、軌道衛星」與緊接的「二之一、Building/System 混合型別」說明,避免與 35 項的精確計數混淆。

### 二之一、Building/System 混合型別(不計入 35 項,單獨列出)

| 建築(英文) | 建議中譯 | 前置研究欄位 | 研究成本(RP) | 維護費(BC/turn) | 效果 | 頁碼 |
|---|---|---|---|---|---|---|
| Stellar Converter(行星版) | 恆星轉換器(行星版) | Temporal Physics | 15000(與 Star Gate、Time Warp Facilitator 同組) | 6 | 對目標造成 400 點傷害 ×2(雙側共 1600),無視射程與防禦;手冊同時描述其**艦載系統版本**(裝在船上,從軌道對星球開火時)可直接摧毀整顆星球(化為小行星帶)——手冊原文用單一段落同時講兩種型態,故標記為 `(Building/System)` 混合型別,行星駐防版才計入「殖民地建築」語意 | p.106 |

## 三、軌道衛星(Satellite,5 項——同屬可建構造物,列入本表)

| 建築(英文) | 建議中譯 | 前置研究欄位 | 研究成本(RP) | 維護費(BC/turn) | 效果 | 取代關係 | 頁碼 |
|---|---|---|---|---|---|---|---|
| Star Base | 星堡 | Engineering(起始已知) | 50(與 Colony Base、Marine Barracks 同組) | 2 | 配備最新式武器的軌道平台;星球掃描範圍 +2 秒差距;沒有 Star Base 的星球無法建造超過驅逐艦(Destroyer)等級的船艦;+1 指揮評等 | 無 | p.77 |
| Battlestation | 戰鬥站 | Robotics | 650(與 Robo Mining Plant、Powered Armor 同組) | 3 | 比 Star Base 火力更強;掃描範圍 +4 秒差距;為己方艦隊 +10 光束攻擊;+2 指揮評等 | **取代**同軌道的 Star Base | p.80 |
| Star Fortress | 星辰要塞 | Superscalar Construction | 6000(與 Advanced City Planning、Heavy Fighter Bays 同組) | 4 | 比 Battlestation 更強;掃描範圍 +6 秒差距;為己方艦隊 +20 光束攻擊;+3 指揮評等 | **取代**同軌道的 Battlestation 或 Star Base | p.83 |
| Artemis System Net | 阿提米絲系統網 | Planetoid Construction | 7500(與 Doom Star Construction 同組) | 5 | 環繞整個星系的巨型水雷網;敵艦進入時依船體等級觸發機率(Frigate 20%~Doom Star 100%);每次觸發 8-28 枚水雷命中,每枚造成 20 點傷害減去目標護盾等級 | 無 | p.84 |
| Dimensional Portal | 次元傳送門 | Multi-Dimensional Physics | 4500(與 Disruptor Cannon 同組) | 2 | 同系統內的艦隊可跨越次元,對安塔蘭人發動攻擊(終局戰觸發點);手冊標記型別為 `(Satellite)`,故只在本節列出,不計入「二、殖民地建築」的 35 項 | 無 | p.106 |

## 四、跨表交叉驗證備註

- **`tech.cpp` 分支/欄位名稱與手冊逐字比對結果**:本表「前置研究欄位」欄取自
  `game_manual.txt` 中每個建築條目前最近一個獨立欄位標題行(結構化解析,非人工判讀),已與
  `openorion2/src/tech.cpp:103-166`(`techtree[8][14]` 陣列)的欄位常數名稱逐一核對,**全部相符**
  (如 Robo Mining Plant 前的欄位標題「Robotics」對應 `TOPIC_ROBOTICS`;Autolab 前的「Cybertronics」
  對應 `TOPIC_CYBERTRONICS`;Stellar Converter 前的「Temporal Physics」對應 `TOPIC_TEMPORAL_PHYSICS`),
  可放心作為 remake 建築系統的「前置科技」欄位依據。
- **與 `moo2-formulas-reference.md` 的關係**:該文件 §2(生產/污染)、§3(士氣)、§4(國庫收入)已對
  Automated Factories、Deep Core Mine、Recyclotron、Pollution Processor、Atmospheric Renewer、Core Waste
  Dump、Holo Simulator、Pleasure Dome、Virtual Reality Network、Psionics 等**個別建築的數值效果**做過
  Go 常數移植與手冊逐句核對;本文件是**橫向全表**(收錄全部 40 項建築/衛星,含研究成本與頁碼),兩者
  互補——實作建築系統時,數值效果請以 `moo2-formulas-reference.md` 已移植的 Go 常數為準,本表提供的是
  「還缺哪些建築沒實作」的總覽與「前置科技/研究成本」欄位。
- **與 `community-mechanics-findings.md` §5 的關係**:該文件先前結論「完整建築成本表(PP)找不到」,
  本文件確認同一結論(手冊敘述性內文沒有 PP 數字),但額外從 `MANUAL_150.html` modding 章節找到唯一
  官方數字(Armor Barracks = 150 PP),並從 `openorion2/src/tech.cpp` 補上了**全部 40 項的 RP 研究成本**
  ——RP 成本原本不在該文件的搜尋範圍內(該文件只搜社群 PP/BC 討論),本表是新增角度的補強,非重複。

## 五、待原版確認清單(誠實列表)

| 項目 | 狀態 |
|---|---|
| 34 項建築(除 Armor Barracks 外)的建造成本(PP) | 無手冊來源,需存檔/資料檔逆向;Biospheres/Cloning Center 兩項有低可信度社群數字,已標注來源等級 |
| `initial_buildings` 優先清單完整排序 | 手冊只透露 entry 1-5,見 `homeworld-init.md` §3.4 |
| Planetary Missile Base / Ground Batteries 具體傷害輸出數值(依裝備武器科技變動) | 手冊只給「配備最佳武器、佔用 N 空間」的規則,實際數值需搭配當時已解鎖的武器科技現算,非固定值 |
| 是否還有本表遺漏的殖民地建築(如更晚期或 1.31/1.50 新增項目) | 本表基於 `GAME_MANUAL.pdf`(patch 1.5 版手冊)「The Big List」完整解析 182 個型別標記項目,理論上已涵蓋全部;若原版 1.31 手冊有 1.5 才移除/新增的建築,需另外核對(本專案未取得 1.31 原始手冊全文,僅有 1.5 手冊) |

## 六、建模狀態(remake engine 對照,2026-07-11)

> 本節記錄 `internal/engine`(`ColonyState`/`RunColonyTurn`/`RunEmpireTurn`)與
> `internal/shell/session.go`(`applyBuildingEffect`)目前**實際套用**了哪些建築效果、哪些只是
> 「記錄已建但不影響數值」——避免本表的效果敘述欄被誤讀為「已全部實作」(rule 63:程式碼真相為準,
> 文件不能留過期斷言佔位)。

### 6.1 已忠實建模(數值真的會變動)

| 建築 | 建模方式 | engine 欄位 |
|---|---|---|
| 自動化工廠 | 每工人 +1(per-worker) + 殖民地固定 +5 | `IndustryPerWorker` + `FlatIndustry` |
| 機器人採礦廠 | 每工人 +2 + 固定 +10 | `IndustryPerWorker` + `FlatIndustry` |
| 深層核心礦場 | 每工人 +3 + 固定 +15 | `IndustryPerWorker` + `FlatIndustry` |
| 研究實驗室 | 每科學家 +1 + 固定 +5 | `ResearchPerScientist` + `FlatResearch` |
| 行星超級電腦 | 每科學家 +2 + 固定 +10 | `ResearchPerScientist` + `FlatResearch` |
| 銀河網路中心 | 每科學家 +3 + 固定 +15 | `ResearchPerScientist` + `FlatResearch` |
| 水耕農場 | 固定 +2(與農夫數無關) | `FlatFood` |
| 地底農場 | 固定 +4(與農夫數無關) | `FlatFood` |
| 氣候控制器 | 每農夫 +2(既有值本來就正確) | `FoodPerFarmer` |
| 太空大學 | 每受教育人口(農/工/科)各 +1 | `FoodPerFarmer`+`IndustryPerWorker`+`ResearchPerScientist` |
| 太空港 | 該殖民地 BC 收入 +50%(逐殖民地精確套用,見 `RunEmpireTurn`) | `IncomeBonusPercent` |
| 行星證券交易所 | 該殖民地 BC 收入 +100%(與太空港疊加) | `IncomeBonusPercent` |
| 生態圈 | 星球人口上限 +2(直接疊加,無獨立 Bonus 影子欄位) | `PopMax` |
| 複製中心 | 固定成長點/回合,直到達人口上限為止(remake 尺度換算,近似值非官方精確數字) | `FlatGrowth` |
| 污染處理器 / 大氣更新器 / 核心廢料場 | 既有 bool 旗標(本次未動) | `PollutionProcessor`/`AtmosphericRenewer`/`CoreWasteDump` |
| 行星重力產生器(p.104) | **2026-07-11 已接線**:`ColonyState` 新增 `PlanetGravity`(該殖民地行星重力)欄位,`colonyFood`/`RunColonyTurn` 對食物/工業/研究三種 per-worker 產出套用 `gamedata.GravityPenaltyPercent`/`GravityAdjustedProduction`(Low-G -25%、Heavy-G -50%,士氣與重力先加總成單一百分點再套一次公式,理由見 `internal/engine/colony.go` 註解)。`NormalizeGravity=true` 時懲罰強制歸零,旗標由 no-op 變成有效。種族 Low-G/High-G 重力天賦尚未建模,固定以 `gamedata.NORMAL_G` 當種族基準;固定加成(`FlatFood`/`FlatIndustry`/`FlatResearch`)不吃重力(remake 建模假設,見 engine 欄位/函式註解)。 | `PlanetGravity`+`NormalizeGravity` |
| 士氣類(全息模擬艙 +20%、歡樂穹頂 +30%、海軍陸戰隊營/裝甲營房解除政府 Barracks 懲罰) | **2026-07-11 已接線**:`GameSession` 新增 `Government`(政府型態)欄位 + `colonyMoralePercent`(`internal/shell/session.go`)——政府基礎值(`gamedata.MoraleGovernmentBase`,依 Barracks 有無決定是否套 -20%)+ 已建士氣建築加成(`gamedata.MoraleHoloSimulatorBonus`/`MoralePleasureDomeBonus`),算出 `ColonyState.MoralePercent`,由 `RunColonyTurn`/`MoraleProductionOutput` 消費(食物/工業/研究各 ±10%/格)。呼叫時機:政府變更(`ApplyGovernment`)、建築完工(`advanceBuilds`→`recalcColonyMorale`)。**母星起始士氣因此從硬編 +10 訂正為 0**(獨裁 + 已建 Marine Barracks 抵消懲罰,無士氣建築加成,見 `playerHomeworldColony` 註解)。 | `MoralePercent`(經 `GameSession.Government`+`ColonyBuildings` 算出) |
| 機器人工廠(p.82) | **2026-07-11 已接線**:比照 `PlanetGravity` 的接線手法(見上一列),`ColonyState` 新增獨立的 `MineralRichness` 欄位,保留建立殖民地當下的原始礦產豐度分類(不再從已烘進 `IndustryPerWorker` 的靜態費率事後反推)。`applyBuildingEffect` 依 `gamedata.ProdRoboticFactoryBonus(int(cs.MineralRichness))`(`internal/gamedata/production.go` 既有查表函式,索引與 `mineralProductionTable` 一致)查出手冊固定值——Ultra Poor+5/Poor+8/Abundant+10/Rich+15/Ultra Rich+20——加進 `FlatIndustry`;不動 `IndustryPerWorker`,避免與已烘進的礦產費率重複計算同一份豐度效果。存檔行星由 `ColonyStateFromSave` 讀 `save.Planet.Minerals`(與 `gamedata.PlanetMinerals` 同源 openorion2 enum ordinal,直接轉型),母星固定 Abundant(`playerHomeworldColony`)。 | `MineralRichness`+`FlatIndustry` |

### 6.2 記錄已建,但目前不影響任何數值(誠實 TODO)

| 建築 | 為何不建模 |
|---|---|
| 異族管理中心(p.92-93) | **部分接線,效果暫不可見**:`colonyMoralePercent` 的多種族懲罰(`gamedata.MoraleMultiRacialPenalty`)因 `ColonyState` 不追蹤「殖民地是否含未同化外族人口」而固定不套用,故此建築目前對士氣沒有可觀察差異;其「加速同化征服人口」效果同樣未建模(remake 無同化速率系統)。待多種族人口追蹤/同化系統落地後補上。 |
| 軍事/防禦類(太空學院、飛彈基地、地面砲台、戰鬥站、星辰要塞、阿提米絲系統網、次元傳送門、行星輻射/通量/屏障護盾、曲速力場干擾器、食物複製機、自動實驗室、再生反應爐) | 對應的艦隊駐防/軌道防禦/指揮評等/工業-食物轉換等系統本身尚未建;等對應系統就緒後再回頭補建模。海軍陸戰隊營/裝甲營房的駐軍生成另見 `ground_invasion.go`(海軍陸戰隊營已建模,裝甲營房尚未)。 |

## 參考來源

| 來源 | 用途 |
|---|---|
| `moo2_patch1.5/GAME_MANUAL.pdf` p.75-112「The Big List」 | 全部 40 項建築/衛星的效果敘述、維護費、互斥關係、頁碼 |
| `openorion2/src/tech.cpp:60-266`(`techtree`、`research_choices` 靜態表) | 每項建築的 RP 研究成本與同組科技,交叉驗證前置研究欄位名稱 |
| `moo2_patch1.5/MANUAL_150.html`「Modding with Config → Buildings」 | 建造成本(PP)機制說明與 Armor Barracks 唯一具體範例 |
| `docs/tech/moo2-formulas-reference.md` §2/§3/§4 | 個別建築數值效果的 Go 常數移植(已完成部分,互補不重複) |
| `docs/tech/community-mechanics-findings.md` §5 | 零星 PP 成本社群數字(低可信度,已標注) |
| `docs/tech/homeworld-init.md` §3 | 起始建築(Marine Barracks + Star Base + Capitol)如何從本表選出 |
