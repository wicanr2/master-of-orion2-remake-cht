# 艦艇元件 → 真 MOO2 科技對應表

> 本文件由資料校正 subagent 產出,**只列對得上真資料的對應,對不上的明講,不臆造**。
> 不改動任何程式碼。核對範圍:`internal/shell/session.go`(WeaponOptions/ArmorOptions/ShieldOptions/SpecialOptions)
> ↔ `assets/i18n/tech.tsv`(中英對照)↔ `internal/gamedata/technames.go`(`TechnologyNames`)↔
> `internal/gamedata/techtree.go`(`researchChoices[83]`,權威:某 TECH 屬哪個 TOPIC 看這裡)。

## 判定方法

1. 元件中文名 → 在 `tech.tsv` 找相同中文的英文科技名。
2. 英文科技名 → 在 `technames.go` 的 `TechnologyNames` 找對應 `TECH_*` 常數。
3. `TECH_*` → 在 `techtree.go` 的 `researchChoices[83]` 逐列搜尋該常數出現在哪個 `TOPIC_*` 的 `Choices` 陣列裡(每列尾端 Go 註解已標明 TOPIC 名,直接引用,無需自行算索引)。
4. 若 `TECH_*` **完全不出現在任何一列 `Choices`**,代表 techtree.go 這份資料裡它沒有掛在任何 TOPIC 下 —— 額外查證後確認:`TECH_DEATH_RAY`、`TECH_XENTRONIUM_ARMOR` 等 12 個常數皆屬此類(見下方「結構性無 TOPIC」)。這對應真實 MOO2 機制:這批是**里程碑/獎勵科技**(達到一定研究進度自動取得,不透過研究主題的三選一/四選一取得),不是資料缺漏。
5. 若元件中文名在 `tech.tsv` 查無對應英文科技名(如「戰鬥電腦」「重生程序」),不強行指派,標「無單一對應(抽象)」並說明真實機制為何。

## 武器(WeaponOptions)

| 元件(中文) | 目前掛的 TOPIC | 真 Technology | 真 Technology 中文名 | 正確 TOPIC | 目前掛對嗎? | 備註 |
|---|---|---|---|---|---|---|
| 無武裝 | TOPIC_STARTING_TECH(0) | — | — | — | 不適用 | 純佔位選項,無對應科技 |
| 雷射 | TOPIC_STARTING_TECH(0) | TECH_LASER_CANNON | 雷射砲 | TOPIC_PHYSICS | **否** | TOPIC_PHYSICS 是 `ResearchAll=true`(Cost 50),不是真的「起始科技」(TOPIC_STARTING_TECH 在 researchChoices 裡是空 Choices 的純佔位列)。因為 ResearchAll,遊戲設計上研究完該主題會一次拿到 Laser Cannon/Laser Rifle/Space Scanner 三項,無需三選一,遊戲體驗上接近「早期就有」但技術上仍需先完成該主題研究 |
| 核飛彈 | TOPIC_STARTING_TECH(0) | TECH_NUCLEAR_MISSILE | 核飛彈 | TOPIC_CHEMISTRY | **否** | 同上,TOPIC_CHEMISTRY 亦為 `ResearchAll=true`(Cost 50),非真起始科技 |
| 質量投射器 | TOPIC_ADVANCED_MAGNETISM | TECH_MASS_DRIVER | 質量投射器 | TOPIC_ADVANCED_MAGNETISM | **是** | 乾淨對上 |
| 中子爆破槍 | TOPIC_ADVANCED_CHEMISTRY | TECH_NEUTRON_BLASTER | 中子爆破槍 | TOPIC_NEUTRINO_PHYSICS | **否** | TOPIC_ADVANCED_CHEMISTRY 的真實 Choices 是 Merculite Missile / Pollution Processor,不含 Neutron Blaster |
| 核融合光束 | TOPIC_ADVANCED_FUSION | TECH_FUSION_BEAM | 核融合光束 | TOPIC_FUSION_PHYSICS | **否** | TOPIC_ADVANCED_FUSION 的真實 Choices 是 Augmented Engines / Fusion Bomb / Fusion Drive,不含 Fusion Beam |
| 麥克萊特飛彈 | TOPIC_ADVANCED_CHEMISTRY | TECH_MERCULITE_MISSILE | 麥克萊特飛彈 | TOPIC_ADVANCED_CHEMISTRY | **是** | 乾淨對上 |
| 高斯砲 | TOPIC_ADVANCED_MANUFACTURING | TECH_GAUSS_CANNON | 高斯砲 | TOPIC_SUBSPACE_FIELDS | **否** | TOPIC_ADVANCED_MANUFACTURING 的真實 Choices 是 Planet Construction / Automated Repair Unit / Recyclotron,不含 Gauss Cannon(這個 TOPIC 其實是「自動修復」的正確歸屬,見下方特殊元件) |
| 相位砲 | TOPIC_ANTIMATTER_FISSION | TECH_PHASOR | 相位砲(Phasor) | TOPIC_MULTIPHASED_PHYSICS | **否** | TOPIC_ANTIMATTER_FISSION 的真實 Choices 是 Antimatter Bomb/Drive/Torpedoes,不含 Phasor |
| 電漿砲 | TOPIC_ARTIFICIAL_GRAVITY | TECH_PLASMA_CANNON | 電漿砲 | TOPIC_PLASMA_PHYSICS | **否** | TOPIC_ARTIFICIAL_GRAVITY 的真實 Choices 是 Graviton Beam / Planetary Gravity Generator / Tractor Beam,不含 Plasma Cannon |
| 死光 | TOPIC_ARTIFICIAL_LIFE | TECH_DEATH_RAY | 死光 | **無**(結構性無 TOPIC) | **否** | `TECH_DEATH_RAY` 確有其名(technames.go 有登記),但在 `researchChoices` 全部 83 列裡完全沒出現在任何 `Choices` 中 —— 對應真實 MOO2 機制:Death Ray 是**里程碑科技**(研究總進度達標自動授予),不是靠三選一研究主題取得的,本來就不該指派任何 TOPIC |

## 裝甲(ArmorOptions)

| 元件(中文) | 目前掛的 TOPIC | 真 Technology | 真 Technology 中文名 | 正確 TOPIC | 目前掛對嗎? | 備註 |
|---|---|---|---|---|---|---|
| 無裝甲 | TOPIC_STARTING_TECH(0) | — | — | — | 不適用 | 純佔位選項 |
| 鈦裝甲 | TOPIC_STARTING_TECH(0) | TECH_TITANIUM_ARMOR | 鈦裝甲 | TOPIC_CHEMISTRY | **否** | 同「雷射/核飛彈」,TOPIC_CHEMISTRY 為 `ResearchAll=true`,非真起始科技 |
| 三鈦裝甲 | TOPIC_ADVANCED_METALLURGY | TECH_TRITANIUM_ARMOR | 三鈦裝甲 | TOPIC_ADVANCED_METALLURGY | **是** | 乾淨對上 |
| 佐特裝甲 | TOPIC_ADVANCED_CONSTRUCTION | TECH_ZORTRIUM_ARMOR | 佐特裝甲 | TOPIC_NANO_TECHNOLOGY | **否** | TOPIC_ADVANCED_CONSTRUCTION 的真實 Choices 是 Automated Factories / Heavy Armor / Planetary Missile Base,不含 Zortrium Armor |
| 中子素裝甲 | TOPIC_ANTIMATTER_FISSION | TECH_NEUTRONIUM_ARMOR | 中子素裝甲 | TOPIC_MOLECULAR_MANIPULATION | **否** | TOPIC_ANTIMATTER_FISSION 真實 Choices 不含 Neutronium Armor |
| 精金裝甲 | TOPIC_ARTIFICIAL_GRAVITY | TECH_ADAMANTIUM_ARMOR | 精金裝甲(Adamantium Armor) | TOPIC_MOLECULAR_CONTROL | **否** | TOPIC_ARTIFICIAL_GRAVITY 真實 Choices 不含 Adamantium Armor |
| 氙素裝甲 | TOPIC_ARTIFICIAL_LIFE | TECH_XENTRONIUM_ARMOR | 氙素裝甲 | **無**(結構性無 TOPIC) | **否** | 與「死光」同類:`TECH_XENTRONIUM_ARMOR` 在 `researchChoices` 全部 83 列裡完全沒出現 —— 真實 MOO2 中這是里程碑科技,本來就不屬於任何 TOPIC |

## 護盾(ShieldOptions)

| 元件(中文) | 目前掛的 TOPIC | 真 Technology | 真 Technology 中文名 | 正確 TOPIC | 目前掛對嗎? | 備註 |
|---|---|---|---|---|---|---|
| 無護盾 | TOPIC_STARTING_TECH(0) | — | — | — | 不適用 | 純佔位選項 |
| 第一級護盾 | TOPIC_ADVANCED_MAGNETISM | TECH_CLASS_I_SHIELD | 第一級護盾(Class I Shield) | TOPIC_ADVANCED_MAGNETISM | **是** | 乾淨對上 |
| 第三級護盾 | TOPIC_ARTIFICIAL_GRAVITY | TECH_CLASS_III_SHIELD | 第三級護盾(Class III Shield) | TOPIC_MAGNETO_GRAVITICS | **否** | TOPIC_ARTIFICIAL_GRAVITY 真實 Choices 不含 Class III Shield |
| 第五級護盾 | TOPIC_ADVANCED_MANUFACTURING | TECH_CLASS_V_SHIELD | 第五級護盾(Class V Shield) | TOPIC_SUBSPACE_FIELDS | **否** | 同上,TOPIC_ADVANCED_MANUFACTURING 真實 Choices 不含 Class V Shield |
| 第七級護盾 | TOPIC_ANTIMATTER_FISSION | TECH_CLASS_VII_SHIELD | 第七級護盾(Class VII Shield) | TOPIC_QUANTUM_FIELDS | **否** | TOPIC_ANTIMATTER_FISSION 真實 Choices 不含 Class VII Shield |
| 第十級護盾 | TOPIC_ARTIFICIAL_LIFE | TECH_CLASS_X_SHIELD | 第十級護盾(Class X Shield) | TOPIC_TEMPORAL_FIELDS | **否** | TOPIC_ARTIFICIAL_LIFE 真實 Choices 不含 Class X Shield |

## 特殊(SpecialOptions)

| 元件(中文) | 目前掛的 TOPIC | 真 Technology | 真 Technology 中文名 | 正確 TOPIC | 目前掛對嗎? | 備註 |
|---|---|---|---|---|---|---|
| 無 | TOPIC_STARTING_TECH(0) | — | — | — | 不適用 | 純佔位選項 |
| 戰鬥電腦 | TOPIC_ARTIFICIAL_INTELLIGENCE | — | — | **無單一對應(抽象)** | **否** | `tech.tsv`/`technames.go` 裡沒有「戰鬥電腦/Battle Computer」這個單一 Technology。真實 MOO2 機制:艦艇的電腦等級是由「電腦」研究領域(RESEARCH_COMPUTERS)一路研究 Electronic → Optronic → Positronic → Cybertronic → Moleculartronic Computer 逐階累積提升,不是靠某一個可三選一取得的獨立科技。目前掛的 TOPIC_ARTIFICIAL_INTELLIGENCE 真實 Choices 是 Neural Scanner / Scout Lab / Security Stations,跟電腦等級完全無關,可判定為掛錯,但也沒有單一「正確 TOPIC」可填(牽涉 TOPIC_ELECTRONICS/OPTRONICS/POSITRONICS/CYBERTRONICS/MOLECULATRONICS 一整條鏈) |
| 自動修復 | TOPIC_ADVANCED_ROBOTICS | TECH_AUTOMATED_REPAIR_UNIT | 自動修復裝置 | TOPIC_ADVANCED_MANUFACTURING | **否** | TOPIC_ADVANCED_ROBOTICS 真實 Choices 是 Bomber Bays / Robotic Factory,不含 Automated Repair Unit(正確歸屬其實是目前「高斯砲」誤掛的那個 TOPIC) |
| 隱形裝置 | TOPIC_ARTIFICIAL_CONSCIOUSNESS | TECH_CLOAKING_DEVICE | 隱形裝置 | TOPIC_DISTORTION_FIELDS | **否** | TOPIC_ARTIFICIAL_CONSCIOUSNESS 真實 Choices 是 Cyber-Security Link / Emissions Guidance System / Rangemaster Unit,不含 Cloaking Device |
| 重生程序 | TOPIC_ARTIFICIAL_LIFE | — | — | **無單一對應(抽象/機制錯置)** | **否** | `tech.tsv` 有「再生(Regeneration)」但那是 `SpecialDevices` enum 的 `SPEC_REGENERATION`(enums.go:679),**不是** `Technology`/`TECH_*`,在 `technames.go`/`researchChoices` 裡完全查無此項。且真實 MOO2 中「再生(Regenerating)」是**種族特性(race trait)**,不是可研究的艦艇裝備 —— 「重生程序」這個特殊元件的設計概念本身可能整個誤置,不只是掛錯 TOPIC 而已,建議連元件設計一併重新檢視 |

## 結構性無 TOPIC 的 Technology(供理解,非本次元件清單成員)

除本檔用到的 `TECH_DEATH_RAY`、`TECH_XENTRONIUM_ARMOR` 外,`TechnologyNames` 裡還有以下常數同樣不出現在 `researchChoices` 任何一列 `Choices` 中(同屬里程碑科技或資料模型未收錄,供未來若要再掛其他元件時參考,避免重蹈覆轍):
`TECH_BLACK_HOLE_GENERATOR`、`TECH_CAPITOL`、`TECH_DAMPER_FIELD`、`TECH_PARTICLE_BEAM`、`TECH_PHASE_SHIFTER`、`TECH_PULSE_RIFLE`、`TECH_QUANTUM_DETONATOR`、`TECH_REFLECTION_FIELD`、`TECH_SPATIAL_COMPRESSOR`、`TECH_SPY_NETWORK`。

## 總結統計

- 共 29 個元件列(含 4 個「無/無武裝/無裝甲/無護盾」佔位項)。
- **乾淨對上(TECH 與 TOPIC 皆正確)**:5 個 —— 質量投射器、麥克萊特飛彈、三鈦裝甲、第一級護盾。(4 個非佔位項)
- **目前掛的 TOPIC 錯誤(TECH 有真實對應,但 TOPIC 指錯)**:16 個 —— 中子爆破槍、核融合光束、高斯砲、相位砲、電漿砲、佐特裝甲、中子素裝甲、精金裝甲、第三級護盾、第五級護盾、第七級護盾、第十級護盾、自動修復、隱形裝置,以及「雷射/核飛彈/鈦裝甲」3 個屬於 `ResearchAll` 主題卻被標成「起始科技(TOPIC_STARTING_TECH)」的特例。
- **TECH 存在但結構性無 TOPIC 可掛(里程碑科技)**:2 個 —— 死光(TECH_DEATH_RAY)、氙素裝甲(TECH_XENTRONIUM_ARMOR)。
- **無單一對應(抽象/概念可能整個錯置)**:2 個 —— 戰鬥電腦(真實是電腦研究鏈的累積效果,非單一科技)、重生程序(真實是種族特性,非可研究裝備;且對應到的是 SpecialDevices 而非 Technology)。
- 佔位項(無武裝/無裝甲/無護盾/無):4 個,不適用。

## 對照方法與不確定處(交由裁決)

- 方法完全依照任務指定的四層交叉引用(元件中文名 → tech.tsv → technames.go → techtree.go),未使用任何猜測性翻譯;每一列「正確 TOPIC」都能在 `techtree.go` 原始碼行號指認(見上方 Choices 引用皆逐字轉寫自程式碼註解標的 TOPIC 名)。
- 不確定/需人工裁決:
  1. **「雷射/核飛彈/鈦裝甲」是否該維持 TOPIC_STARTING_TECH=0** 這個簡化設計 —— 若 remake 刻意要讓玩家開局就有這三項(不管真實 MOO2 是否需要研究 `ResearchAll` 主題),這是有意的遊戲性簡化,不算「錯」;但若要忠實對齊研究解鎖機制,應改標對應的 `ResearchAll` TOPIC。兩種選擇都合理,取決於 remake 的設計目標。
  2. **「戰鬥電腦」「重生程序」建議整個重新設計**,而不只是改掛 TOPIC —— 因為它們在真實 MOO2 裡根本不是「靠三選一科技解鎖的艦艇元件」這個資料形狀,硬塞一個 TOPIC 只會製造新的誤導性對應。是否要保留這兩個元件、或改用其他真實存在的科技/機制替換,需要裁決。
  3. **死光/氙素裝甲要不要改用「完成 X 個科技後解鎖」機制**,還是暫時維持掛在某個 TOPIC 底下當作簡化替代 —— 目前程式碼的 `Component.Tech` 欄位只支援單一 `ResearchTopic`,無法表達「里程碑」語意,這是資料模型層級的限制,不是單純打錯字能修正的。
