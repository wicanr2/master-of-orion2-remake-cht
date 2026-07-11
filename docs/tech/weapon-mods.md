# 武器改造(Weapon Modifications)

> 2026-07-11。對應 Go:`internal/gamedata/weapon_mods.go`(+ `weapon_mods_test.go`)、
> `internal/shell/weapon_mods.go`、`internal/shell/session.go` 的
> `ShipDesignSpaceUsedWithMods`/`ShipDesignFitsWithMods`/`DesignCostWithMods`/`BuildShipWithMods`、
> `internal/shell/combat_formula.go` 的 `ResolveShotWithMods`、`cmd/moo2/interactive.go` 的艦艇
> 設計畫面(`shipDesign`)與戰術格鬥畫面(`fireRound`)。

## 出處(硬門檻:手冊逐字數字)

**`moo2_patch1.5/GAME_MANUAL.pdf`(patch 1.5 隨附完整手冊,188 頁,`pdftotext -layout` 可正常抽字,
非掃描圖)p.115-118,「Modifications」章節**(標題原文:「Nearly every weapon could be enhanced
in some way ... What follows is an introduction to the potential modifications available for the
weapons you can install in your ships. ... each one adds to the size and cost of the weapon, and
some are mutually exclusive」)。這是**逐項給精確數字的原始出處**——`ship-design-space.md` 第 3 節
先前只查到 p.128(Design Dock 說明頁)對「會加大 size/cost」的重申與部分 mod(未收 CO/ENV/NR/SP),
本輪重新用 `pdftotext -layout moo2_patch1.5/GAME_MANUAL.pdf` 抽全文後找到完整章節。

openorion2 原始碼核對結果:`grep -rli "HeavyMount\|PointDefense\|AutoFire\|Enveloping\|ArmorPiercing\|ContinuousFire\|WeaponMod"` 在整個 `openorion2/` 樹**零命中**——確認 openorion2 只是渲染殼
(見 memory `openorion2-is-renderer-not-engine`),沒有任何武器 mod 邏輯可抄,本檔數字完全來自手冊
文字 + 現有 `internal/gamedata/damage.go`/`combat.go` 已移植的 Hv/PD/HEF 公式框架。

## 8 個已接線的 mod(手冊精確數字)

只收「本輪要接線」的光束/通用 mod。飛彈/魚雷專屬 mod(ARM/ECCM/EMG/FST/MV/OVR、以及 NR 的魚雷版
「Not Reduced By Range」)手冊同樣給了精確數字,但 remake 的飛彈解算(`missile.go`/
`ResolveMissileShot`)目前尚未有 mod 掛鉤機制,故不接線,留待「飛彈 mod」任務(見文末 TODO)。

| 代碼 | 名稱 | 手冊原文摘要(p.115-118) | 佔格/成本 | 效果 |
|---|---|---|---|---|
| HV | Heavy Mount | 大型平台版,150% 傷害;射程懲罰(命中+衰減)減半;與 PD 互斥 | **+100%** | `hvBonus=+50` 餵 `DamageMountAdjustedValue`;射程等級改用 `CombatRangeLevelHeavy`(halved) |
| PD | Point Defense | 小型精簡版,半傷害;命中 +25%;射程懲罰(命中+衰減)加倍;與 HV 互斥 | **-50%**(減半) | `pdPenalty=+50`;`pdBonus=+25`(命中門檻);射程等級改用 `CombatRangeLevelPointDefense`(doubled) |
| AF | Auto-Fire | 3 連發,每發 -20% 命中;需 2 級小型化 | **固定 +50**(非百分比,見下方說明) | netAttack(BA+CO-AF-BD)每次射擊 -20 |
| CO | Continuous Fire | 持續開火,+25 命中;需 1 級小型化 | +50% | netAttack +25 |
| AP | Armor Piercing | 穿透除 Xentronium/Heavy Armor 外所有裝甲;需 1 級小型化 | +50% | `DamageApplyArmor` 的 `armorPiercing=true`(全額打結構) |
| ENV | Enveloping | 同時打四面護盾,傷害四倍;需 2 級小型化 | +100% | 命中後傷害 `*4` |
| NR | No Range Dissipation | 消除射程衰減;需 1 級小型化 | +25% | **未接效果**(見下方誠實記錄) |
| SP | Shield Piercing | 完全忽略護盾(對行星護盾無效);需 1 級小型化 | +50% | `DamageAfterShield` 的 `shieldPiercing=true` |

### AF 的 +50 為什麼是固定值不是百分比

手冊其餘 mod 條目一律明寫「by X%」,唯獨 AF 那句是「This modification increases the size and
cost of the weapon by 50」——**沒有 % 符號**,與其他 7 個 mod 的措辭明顯不同。`ship-design-space.md`
先前分析(未接線階段)已經注意到這個措辭差異並標註「固定值,非百分比」,本輪沿用同一結論
(`gamedata.WeaponModAutoFireFlatSpaceCost = 50`),多重來源交叉一致。

### CO/AF/PD 的命中點數:手冊 + 社群逆向交叉核對

手冊原文直接給 CO 是「increasing the weapon's accuracy by 25」(無 % 符號,點數)、AF 是「20%
penalty to its accuracy」(有 % 符號,但整個 Beam Attack/Defense 系統本身是點數制——電腦每級
+25、船員 Elite +75、Battle Scanner +50——不是額外的百分比縮放,PD "25% greater accuracy" 同理)。
`docs/tech/community-mechanics-findings.md` 引用 Olesh 的社群拆解列出同一套點數表:
「PD +25、Continuous Fire +25、Auto Fire -20」,與手冊逐字對照後採直接點數解讀:CO=+25、
AF(每次射擊)=-20、PD=+25(套用在 `CombatHitThreshold` 的 `pdBonus` 參數,而非 netAttack)。

## 佔格/成本疊加公式(誠實標註:非手冊逐字數字的部分)

`gamedata.WeaponSpaceWithMods(baseSpace, mods)`:百分比 mod **加總後一次套用**(如 AP+CO 都掛
=+50%+50%=+100%,而非連續複利 1.5×1.5=1.25=+125%),對照 `damage.go` 的
`DamageMountAdjustedValue` 手冊明載「Hv/PD/HEF interaction is not multiplicative but additive」
慣例類推。**手冊本身沒有明講「同一武器同時掛多個 mod 時,佔格百分比是加總一次套用還是連續複利」**
——這點類推自傷害公式的既定慣例,不是手冊逐字數字,若日後找到反證(如逆向遊戲存檔資料格式或
DOSBox 黑箱測試)需回頭修正。AF 的固定 +50 在百分比套用「之後」再相加(手冊明寫是固定值,不應
被其他 mod 的百分比再放大)。

成本(Cost)手冊原文「adds to the size **and cost**」——同一套百分比同時套用在兩者,`WeaponCostWithMods`
直接重用 `WeaponSpaceWithMods` 的公式(對 baseCost 而非 baseSpace)。

## HV/PD 命中/傷害的完整解算路徑(`ResolveShotWithMods`)

`internal/shell/combat_formula.go` 的 `ResolveShotWithMods` 是實際接線點(`battleVolley` 快速結算
與 `cmd/moo2/interactive.go` 的 `fireRound` 格鬥畫面都呼叫它):

1. `netAttack = netAttackBase + WeaponModNetAttackBonus(mods)`(CO/AF)。
2. `level = CombatRangeLevelForBeamMods(rangeSquares, mods)`——HV 用
   `CombatRangeLevelHeavy`(手冊「actual range is halved」)、PD 用
   `CombatRangeLevelPointDefense`(手冊「range is as if doubled」),都沒掛用一般
   `CombatRangeLevel`。
3. `threshold = CombatHitThreshold(CombatRangeLevelPenalty(level), WeaponModPDBonus(mods))`——PD 的
   +25 命中加成餵進既有的 `pdBonus` 參數(該參數原本就是為此保留,`combat.go` 註解早已標記
   「手冊未在本節給出精確數字」,本輪補上)。
4. 命中判定(`CombatClassicToHit`)通過後,若有 mod 才呼叫
   `DamageMountAdjustedValue(weaponMin/Max, hvBonus, 0, pdPenalty, 0)` 調整傷害潛力(150%/50%);
   **無 mod 時完全跳過這一步**,直接沿用原始 `weaponMin`/`weaponMax`(見下方回歸保護說明)。
5. `DamageForHit` 算出命中後傷害。
6. `WeaponModEnvelopingMultiply`:ENV 則 `*4`。
7. `DamageAfterShield(..., shieldPiercing=WeaponModShieldPiercing(mods))`。
8. `DamageApplyArmor(..., armorPiercing=WeaponModArmorPiercing(mods), false)`。

### 回歸保護:為什麼無 mod 時不能無腦呼叫 `DamageMountAdjustedValue`

`DamageMountAdjustedValue` 依手冊「the minimum damage potential is always 1」把結果夾限最少為
1。若對「無武裝」(`weaponMin=weaponMax=0`)這種武器無條件呼叫該函式,0 傷害會被誤夾成 1——
變成「沒裝武器的船打出 1 點傷害」這種不該有的回歸。`ResolveShotWithMods` 因此在 `len(mods)==0`
時完全跳過該函式呼叫,直接用原始 `weaponMin`/`weaponMax`,與加入 mod 系統前的 `ResolveShot`
逐位元相同(`combat_formula_test.go` 的 `TestResolveShotWithMods_NoModsMatchesLegacy` 涵蓋這個
邊界情況)。

## NR(No Range Dissipation)為什麼目前沒有可觀察效果

`internal/shell/combat_formula.go` 的 `ResolveShot`/`ResolveShotWithMods` 是既有簡化模型:
射程(`rangeSquares`)只影響**命中門檻**(`CombatRangeLevelPenalty`),從不對 `weaponMin`/
`weaponMax` 套用任何**傷害衰減**(`gamedata.DamageDissipationPenalty`/`DamageApplyDissipation`
在 shell 層完全沒被呼叫)。NR 的手冊效果是「消除傷害隨距離衰減」——但這套簡化模型本來就沒有
在模擬「傷害隨距離衰減」這件事,所以 NR 目前接了佔格/成本(佔格 UI 上會反映 +25%),但戰鬥傷害
不會有任何變化。**這不是遺漏,是要消除的機制本身還沒被模擬**。待 shell 層導入射程傷害衰減後
(有專門任務再做),NR 才有東西可以「消除」。

## 資料模型與存檔

- `shell.Ship` 新增 `Mods []string` 欄位(存 `gamedata.WeaponModCode` 字串,如 `"HV"`),直接
  被既有 `sessionSnapshot.Ships []Ship` 序列化,**不需要另外改 persist.go**——舊存檔沒有這個欄位,
  JSON 解碼會是 `nil`,行為與「無改造」完全一致(回歸安全)。
- `shell.CombatShip` 同樣新增 `Mods []string`,供 `StartCombat`(建立戰術格鬥雙方)與 `fireRound`
  (格鬥開火)使用。
- `combatant`(`ResolveBattle`/`battleVolley` 快速結算用的內部型別)新增 `mods []string` 欄位,
  由 `mkPlayer()` 從 `sh.Mods` 帶入;敵方艦隊(`genEnemyFleet`)沒有個別武器設計,一律 `nil`
  (既有簡化,非本輪引入)。

## UI(艦艇設計畫面,`cmd/moo2/interactive.go` 的 `shipDesign`)

8 個 mod 顯示成兩排各 4 個的 chip(HV/PD/AF/CO、AP/ENV/NR/SP),點擊切換(`shell.ToggleWeaponMod`);
HV/PD 手冊明訂互斥,勾選其一自動移除另一個。已勾選 chip 轉金色高亮,未勾選灰色。非光束武器(核飛彈/
麥克萊特飛彈)選中時,標題列改顯示「僅光束武器適用,此武器不支援」提示——熱區仍在但點擊不生效
(`shell.WeaponIsBeam` 判斷),避免版面跳動同時不誤導玩家。空間/成本顯示即時反映
`ShipDesignSpaceUsedWithMods`/`DesignCostWithMods`,建造前用 `ShipDesignFitsWithMods` 擋下超格設計
(同一份判斷,顯示與驗證不會不一致)。

UI 做到「可勾選 + 即時顯示 + 建造擋超格」這個最小可行版本,未做的部分(標記 TODO,見
`docs/tech/remaining-work-roadmap.md`):
- Chip 沒有滑鼠 hover 顯示手冊效果說明(如「+50% 傷害、-50% 射程懲罰」的文字提示)。
- 沒有「小型化等級」門檻檢查(手冊部分 mod 要求「weapon has undergone N 級小型化」才能裝,
  現行 remake 沒有小型化系統本身,故沒有這個門檻,任何已解鎖武器都能直接掛任何 mod——比原版寬鬆)。
- 火線角(Firing Arc:Fwd Ext/Back Ext/360 Degree,p.127-128)是平行、獨立於 mod 的機制,本輪
  不包含。

## 未接線的飛彈/魚雷專屬 mod(TODO,標記待考證/待做,非本輪範圍)

手冊 p.115-116 也給了這些精確數字,但 remake 的飛彈解算(`ResolveMissileShot`)本身沒有 mod
掛鉤機制,故只記錄不接線,避免定義了卻無人使用的死碼:

| 代碼 | 名稱 | 佔格/成本 | 效果(手冊摘要) |
|---|---|---|---|
| ARM | Heavily Armored(飛彈) | +25% | 摧毀所需傷害 ×2 |
| ECCM | Electronic Counter-Counter-Measures | +25% | 干擾偏移機率減半 |
| EMG | Emissions Guidance | **×4** | 命中護盾後繞過裝甲,直接傷驅動 |
| FST | Fast Missile | +25% | 每回合多移動 4 格,對應提升防禦 |
| MV | MIRV | +100% | 4 枚全額彈頭,傷害 ×4(不可用於行星轟炸) |
| NR(魚雷版) | Not Reduced By Range | +25% | 魚雷版的「消除射程衰減」 |
| OVR | Overloaded(魚雷) | +50%(整套系統) | 彈頭強度 +50% |

## 測試

- `internal/gamedata/weapon_mods_test.go`:佔格/成本百分比與固定值疊加、CO/AF netAttack 點數、
  PD 命中門檻加成、HV/PD 傷害調整端到端(透過既有 `DamageMountAdjustedValue`)、ENV 四倍、
  AP/SP 旗標、HV/PD 射程等級表選擇。
- `internal/shell/shipspace_test.go`:`ShipDesignSpaceUsedWithMods`/`ShipDesignFitsWithMods`/
  `DesignCostWithMods` 無 mod 回歸一致、HV 佔格加倍、超格擋下、mods 對飛彈武器無效。
- `internal/shell/combat_formula_test.go`:`ResolveShotWithMods` 無 mod 回歸一致(含 0 傷害邊界)、
  HV/PD 傷害調整、ENV 四倍、AP/SP 繞過裝甲/護盾、CO/AF 改變命中結果、HV/PD 射程等級表確實不同。
- 探針(`zz_probe_test.go`,已跑完刪除,不提交):20 回合開局 BC 走勢無 regression
  (100→130,對照既有 memory 基準)、mod 端到端生效驗證。

全部測試綠(`go test ./internal/... ./cmd/...`,uifont 的 X11/GLFW panic 為既有環境限制,排除)。
