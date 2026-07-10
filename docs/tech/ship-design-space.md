# 艦艇設計空間格模型(Ship Design Space)

> 2026-07-11。對應 Go:`internal/gamedata/shipspace.go`(+ `shipspace_test.go`)、
> `internal/shell/session.go` 的 `ShipDesignSpaceUsed`/`ShipDesignFits`/`shipClassFromName`
> (+ `internal/shell/shipspace_test.go`)。**本輪只做 shell/gamedata 層 + 驗證函式,UI 繪製
> (Design Dock 畫面本身、`interactive.go`)不在範圍內,留給後續任務。**

## 為什麼這次能查到精確數字(關鍵發現)

`docs/tech/component-values.md` 先前的結論是「線上無機讀的完整權威武器表」,因為當時檢查的
`original_game/…CD Manual.pdf` 是**掃描圖**(pdftotext 抽字 0 字元)。但 `moo2_patch1.5/GAME_MANUAL.pdf`
是 **patch 1.5 隨附的完整手冊(188 頁)**,由 Google Docs 匯出(`Producer: Skia/PDF … Google Docs
Renderer`),是**可正常抽字的文字 PDF**,先前的盤點沒注意到這份檔案其實含完整手冊本文(只查過
`MANUAL_150.html`,那份是 1.50 patch 的「異動摘要」,不是完整手冊)。用 `pdftotext -layout` 抽出全文後,
「Ship Design」章節(約 p.119-132)完整保留,含本檔所有數字表。

## 1. 艦體總空間(GAME_MANUAL.pdf p.121,已確認)

手冊原文(p.120-121):「When you select a class, the default icon for that ship size appears in
the Icon box, along with a notation of the **total amount of space available in that hull size**.」

| Class(艦體等級) | Cost | **Space(總空間)** | Marines | Armor | Struct. | Comp. | Drive | Shield |
|---|---|---|---|---|---|---|---|---|
| Frigate(護衛艦) | 20 | **25** | 5 | 4 | 4 | 1 | 2 | 1 |
| Destroyer(驅逐艦) | 70 | **60** | 8 | 10 | 10 | 2 | 5 | 2 |
| Cruiser(巡洋艦) | 250 | **120** | 12 | 30 | 30 | 5 | 10 | 6 |
| Battleship(戰艦) | 600 | **250** | 20 | 50 | 50 | 7 | 15 | 7 |
| Titan(泰坦) | 1500 | **500** | 30 | 80 | 80 | 10 | 20 | 10 |
| Doom Star(末日之星) | 4000 | **1200** | 50 | 150 | 150 | 20 | 40 | 20 |

**交叉驗證**:「Comp.」欄(1,2,5,7,10,20)與「Drive」欄(2,5,10,15,20,40)分別**完全對應**
openorion2 `gamestate.cpp:150`(`computerHPTable`)與 `gamestate.cpp:154`(`driveHPTable`),已寫進
`internal/gamedata/formulas.go`(`docs/tech/moo2-formulas-reference.md` 第 8 節)。這證實手冊這張表
與 openorion2 的唯讀查表是同一份原始資料的兩種呈現,「Space」欄可信度同等。

**附帶發現(超出本輪範圍,留給未來任務)**:「Armor」/「Struct.」欄(4,10,30,50,80,150,兩欄同值)
與「Shield」欄(1,2,6,7,10,20)目前 codebase 尚未使用——`docs/tech/moo2-formulas-reference.md` 第 9 節
明確記錄「艦艇武裝/裝甲/護盾數值表 未見」、`internal/gamedata/damage.go` 的 `DamageShieldCapacity`
註解也寫著「shipSize 為船體 size class 對應的數值…呼叫端自行從…換算,本函式不假設固定對照」。
這兩欄很可能就是缺的那張表(Armor/Struct 供 ArmorHP/StructureHP 換算、Shield 欄可能就是
`DamageShieldCapacity` 缺的 `shipSize` 參數表),但這屬於「戰鬥傷害模型」而非本輪「設計空間格」
的任務範圍,**故本輪不採用、不實作**,留待下一輪戰鬥系統忠實化任務查證後接線。

`internal/gamedata/shipspace.go` 的 `ShipHullSpace(class CombatShipClass) int` 對照上表 Space 欄,
索引沿用既有 `CombatShipClass`(0=Frigate…5=DoomStar,與 formulas.go 同一組索引)。

## 2. 武器佔格(GAME_MANUAL.pdf p.124-127,已確認)

手冊原文(p.123-124):「Space: The amount of space inside the hull that each emplacement takes up.」
「Weapons that will fit into the hull (under the current configuration) are highlighted. Those
that won't fit are dimmed.」——證實武器佔格是**固定值**,與艦體大小無關(裝不裝得下純看剩餘空間)。

### 束射武器(Beam,p.124)

| 武器 | Damage | **Size** | Cost | Specials |
|---|---|---|---|---|
| Laser Cannon(雷射) | 1-4 | **10** | 5 | - |
| Fusion Beam(核融合光束) | 2-6 | **10** | 6 | dr |
| Mass Driver(質量投射器) | 6 | **10** | 7 | nr |
| Ion Pulse Cannon | 2-10 | **30** | 15 | emp |
| Neutron Blaster(中子爆破槍) | 3-12 | **10** | 8 | kills marines |
| Graviton Beam | 3-15 | **15** | 12 | esd |
| Gauss Cannon(高斯砲) | 18 | **10** | 10 | nr |
| Phasor(相位砲) | 5-20 | **10** | 10 | - |
| Plasma Cannon(電漿砲) | 6-30 | **25** | 15 | dr, env |
| Disrupter | 40 | **20** | 25 | nr |
| Mauler Device | 100 | **50** | 75 | always hits |
| Particle Beam | 10-30 | **15** | 35 | sp |
| Death Ray(死光) | 50-100 | **30** | 75 | kills marines, co |

> Plasma Cannon 傷害值 1.31(6-30)與 1.50(4-20,`MANUAL_150.html` 記載)不同,但 **Size 不受版本
> 影響**(見 `component-values.md` 的版本相依記錄);本表 Size 欄取自 `GAME_MANUAL.pdf`(標題頁本身
> 印「version 1.50」),兩版通用。

### 飛彈/魚雷(p.125-126)

| 武器 | Dmg | **Size** | Cost | 備註 |
|---|---|---|---|---|
| Nuclear Missile(核飛彈) | 8 | **10/20/30/35/40**(依彈架 x2/x5/x10/x15/x20) | 10-22 | 彈架容量越大,Size/Cost 同比放大 |
| Merculite Missile(麥克萊特飛彈) | 14 | 同上 | 10-22 | 同上 |
| Pulson Missile | 20 | 同上 | 同上 | 同上 |
| Zeon Missile | 30 | 同上 | 同上 | 同上 |
| A-M Torpedo | 25 | **20** | 15 | 固定值(魚雷無彈架選擇) |
| Proton Torpedo | 40 | **30** | 20 | 固定值 |
| Plasma Torpedo | 120 | **40** | 75 | 固定值 |

飛彈(Missile,非魚雷)的 Size 是「依彈架容量遞增的一組值」而非單一值——本專案簡化模型
(`session.go` 尚未實作彈架容量選擇)取**最小彈架(x2 = 10)**當估計值。

### 炸彈(Bomb,p.126)

| 武器 | Damage | Size | Cost |
|---|---|---|---|
| Nuclear Bomb | 3-12 | 5 | 3 |
| Fusion Bomb | 4-24 | 7 | 5 |
| Anti-Matter Bomb | 5-40 | 7 | 6 |
| Neutronium Bomb | 10-60 | 10 | 9 |
| Death Spores | 10%(殺人口) | 5 | 5 |
| Bio-Terminator | 20%(殺人口) | 7 | 8 |

### 戰機/特殊武器(p.126-127)

| 武器 | Size | Cost | 備註 |
|---|---|---|---|
| Interceptor | 30 | 10 | 戰機 |
| Bomber | 60 | 30 | 戰機 |
| Heavy Fighter | 80 | 50 | 戰機 |
| Assault Shuttle | 25 | 10 | 戰機(載陸戰隊) |
| Anti Missile Rocket | 20 | 5 | 特殊武器 |
| Gyro Destabilizer | 75 | 50 | 特殊武器 |
| Tractor Beam | 30 | 20 | 特殊武器 |
| Pulsar | 50 | 30 | 特殊武器 |
| Plasma Web | 40 | 40 | 特殊武器 |
| Stasis Field | 75 | 75 | 特殊武器 |
| Stellar Converter | 500 | 500 | 特殊武器 |
| Spatial Compressor | 50 | 40 | 特殊武器 |
| Black Hole Generator | 150 | 150 | 特殊武器 |

`internal/gamedata/shipspace.go` 的 `WeaponSpaceByName` 對照本專案既有
`internal/shell/session.go` `WeaponOptions` 的 10 個元件名,逐項標「確認值」或「估計」
(飛彈類取最小彈架)。

## 3. 武器改裝(Mods)對佔格的影響(p.128,已確認,未接線)

| 改裝 | 對 Size/Cost 的影響 |
|---|---|
| Fwd Ext / Back Ext(延伸射界 240°) | +25% |
| 360 Degree(全周射界) | +50% |
| AF(Auto-Fire) | Size/Cost 各 **+50**(固定值,非百分比) |
| AP(Armor Piercing) | +50% |
| ARM(飛彈裝甲) | +25% |
| HV(Heavy Mount) | +100% |
| PD(Point Defense) | **-50%**(減半) |
| Shield Piercing | +50% |

本輪未接線(session.go 目前的武器選擇沒有 mod 系統),留待後續「武器改裝」任務,先誠實記錄
公式來源避免日後重找。

## 4. 特殊系統(Specials Area)佔格:手冊只有定性描述,無數字(誠實標「待查」)

手冊原文(p.128-129,Design Dock 說明):「This console is a list of all the special systems you
have available. For each system, the listing includes its name, **the space it requires**, what
it adds to the production cost, and a brief description of its effects. Note that, **contrary to
weapons, special systems cost more to install in larger ships and take up more space in a larger
hull**.」

確認的機制:特殊系統(如 Battle Scanner、Cloaking Device、Automated Repair Unit、Extended Fuel
Tanks、Subspace Teleporter 等)的佔格**依艦體大小縮放**,與武器(固定值)相反。但手冊全文
(`GAME_MANUAL.pdf` 188 頁 + `MANUAL_150.html`)搜遍都**沒有給出任何一個特殊系統的精確空間數字
或百分比公式**——不像武器有逐項 Size 表可抄。

`internal/gamedata/shipspace.go` 的 `SpecialSpaceEstimatePercent = 5`(艦體空間的 5%)是**誠實標註
的估計值**,不是手冊數字,只用來讓 `ShipDesignSpaceUsed`/`ShipDesignFits` 在目前簡化模型下有個
非零、合理量級的佔格,避免「特殊系統完全不佔空間」這個更失真的預設。

**待辦**:精確的特殊系統空間係數需要:
- 逆向遊戲資料檔(非手冊文字)本身的 Special System 佔格表,或
- 找到社群已逆向的精確版本(可信度需再核實,`community-mechanics-findings.md` 目前未收錄此項)。

## 5. 裝甲/護盾不佔空間(重要澄清,可能與任務假設不同)

手冊原文(p.121-122,Automatics 段落):「To the right of the Hull Size box are the Automatics.
These two boxes list all of the ship essentials — **the engines, armor, shields, and computer**.
For each of these, the best that you have available is automatically installed in every ship you
build or refit. … Don't worry, **the engines, armor, and fuel cells do not take up space that
might be used for optional systems**.」

也就是說,MOO2 原版的空間預算(Weapons Area + Specials Area 共用的「Space Available」計數器)
**只包含武器與特殊系統兩類**;裝甲、護盾、電腦、引擎是「自動裝上目前科技最好的一套」的
Automatics,**不佔用**這個空間預算(玩家只能在 Automatics 區塊把護盾/電腦降級以省成本,不涉及
空間)。

這與本次任務指示的「四下拉(武器/裝甲/護盾/特殊各選一)…每個元件佔用空間」假設**不完全一致**
——原版的「裝甲/護盾佔格」這個前提在手冊裡找不到依據,反而手冊明確說它們不佔格。本專案既有的
`internal/shell/session.go` 四下拉模型(把裝甲/護盾當成跟武器同層級的可選 Component)本身已經是
對手冊的簡化(手冊沒有「選裝甲/護盾佔格」這件事);為了不讓錯誤假設繼續往下傳,
`ShipDesignSpaceUsed(class, weapon, armor, shield, special)` 保留完整四參數簽名(維持呼叫相容),
但**裝甲/護盾一律回報 0 空間**,對齊手冊行為,不是遺漏或臆造。

## 6. Go 函式對照

| 函式 | 位置 | 說明 |
|---|---|---|
| `ShipHullSpace(class CombatShipClass) int` | `gamedata/shipspace.go` | 艦體總空間(p.121 確認值) |
| `WeaponSpaceByName map[string]int` | `gamedata/shipspace.go` | 各武器佔格(p.124 確認值 + 飛彈估計) |
| `SpecialSpaceEstimatePercent`、`SpecialSpace(hullSpace int, hasSpecial bool) int` | `gamedata/shipspace.go` | 特殊系統佔格估計模型(**非手冊數字**) |
| `shipClassFromName(class string) (gamedata.CombatShipClass, bool)` | `shell/session.go` | 中文艦體名 → enum;偵察艦等非 Design 六級艦體近似 Frigate |
| `ShipDesignSpaceUsed(class string, weapon, armor, shield, special int) int` | `shell/session.go` | 已用空間 = 武器佔格 + 特殊系統估計佔格(裝甲/護盾回報 0,見 §5) |
| `ShipDesignFits(class string, weapon, armor, shield, special int) bool` | `shell/session.go` | 已用空間 <= 艦體總空間 |

## 7. 測試

- `internal/gamedata/shipspace_test.go`:艦體空間逐列核對(6 級全對)、邊界(非法 class 回 0)、
  武器佔格逐項核對、`SpecialSpace` 估計模型行為(含捨去/最少 1 格)。
- `internal/shell/shipspace_test.go`:裝甲/護盾不佔空間驗證、特殊系統估計佔格驗證、
  小艦體(護衛艦 25)塞死光(30)超格、大艦體(末日之星 1200)容納、恰好用滿的邊界
  (護衛艦 25 裝電漿砲 25)、未知艦體(偵察艦)近似判定。

## 8. 與既有文件的落差(本輪已同步更新)

- `docs/tech/gameplay-systems-status.md` §1b 原本把「艦艇空間格」與「飛彈/球狀傷害」並列為
  「需逆向演算法的新子系統」——這個歸類不成立:本輪證實艦艇空間格**不需要 RE**,只是先前沒找到
  完整手冊文字版(§0 已說明原因),已於該文件更正,飛彈/球狀傷害仍歸「需 RE」不變。
- `docs/tech/gameplay-systems-status.md` §3、`WORKLIST.md`、`docs/HANDOFF.md`、
  `docs/HONEST-STATUS.md` 的「艦艇設計(空間格)」相關條目已同步標記 shell/gamedata 層完成度,
  UI 繪製仍待後續任務。
