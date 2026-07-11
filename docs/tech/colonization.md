# 最小拓殖(Colonization)接線紀錄

> 2026-07-11。目的:讓玩家能用殖民船在無主適居星建立新殖民地——先前玩家只有母星、完全無法擴張,
> 這是「能玩完整一局」目前最大的解鎖(見 `remaining-work-roadmap.md` B 項)。本檔記錄硬門檻查證
> 結果、實作範圍、已知簡化與 TODO。

## 1. 硬門檻查證(手冊原文引用,`moo2_patch1.5/GAME_MANUAL.pdf`,`pdftotext -layout` 直接萃取
文字,非 OCR)

### 1.1 適居性

p.55「Planets」小節:

> "Planets come into two different categories: gas giants and habitable worlds. You can build
> a military outpost in a close orbit around a gas giant (and in an asteroid belt), but
> colonies can only survive on a solid planet."

p.61「Creation」小節:

> "A Colony Ship can establish a colonial foothold on any uncolonized planet in its range, as
> long as all space monsters and enemy ships have been cleared from that planet's system."

p.50 節錄(Advanced Colonization Techniques 一類科技的效果描述):

> "This technology allows a colony in the same system with an asteroid field or gas giant to
> support a colony. This planet is Barren, Normal G, and mineral Abundant. Gas giants make
> Huge worlds, and asteroid belts make Large ones."

**結論**:MOO2 的行星分兩大類——氣態巨星/小行星帶(只能建軍事前哨,需要額外科技才能升級成
真正殖民地)、一般固態行星(「habitable worlds」,殖民船可直接殖民,不需額外科技,只要求系統內
無敵艦/怪物)。

**本 remake 現況**:星系生成(`internal/shell/session.go` 的 `genGalaxy`/`genPlanets`)每顆星
固定生成一顆「一般行星」,**從未生成氣態巨星或小行星帶**——`gamedata.PlanetType`(`ASTEROIDS`/
`GAS_GIANT`/`HABITABLE`)這個 enum 雖然存在(由 openorion2 `gamestate.h` 生成),但 `genPlanets`
完全沒有使用它標記行星類別。因此「哪些行星需要額外科技」這個問題在目前的資料模型裡**沒有實際
案例可套用**——所有可拓殖目標必然是「一般行星」,直接可殖民。`internal/shell/colonization.go`
的 `climateColonizable` 函式保留為未來擴充掛勾點(若日後補上氣態巨星/小行星帶星系生成),現階段
恆真,不是刻意放寬規則。

### 1.2 新殖民地起始狀態

p.61-62「Creation」小節,直接引文:

> "A Colony Base establishes a new site in the same system as the colony that built the base
> in the first place... When the Colony Base is complete, the new colony is established with
> one unit of population. (The building colony supplies this unit, but does not lose one of
> its own.)"

> "The Colony Ship is a long-range mobile version of the Colony Base... When the ship is built,
> a new unit of population is gathered to board the ship; the planet that builds the ship does
> not suffer any loss of population."

**結論**:Colony Base 與 Colony Ship 兩種建立殖民地的方式,起始人口一致 = **1 單位**。這是直接
手冊引文,非猜測,高信心。

手冊全文未提及新殖民地會自動附帶任何建築——對照母星起始建築(`homeworldBuildings()`:海軍陸戰隊營
+ 星基)是「Pre-warp/Average Tech games only」的明確特例文字,新殖民地的沉默代表**起始無建築**。
`ColonizeStar` 因此把新殖民地的 `ColonyBuildings` 條目設為 `nil`(空)。

**初始工作分配**(Farmers/Workers/Scientists):手冊完全沒有規則。population=1 時本檔選擇
「全農」,而非比照 `session.go` `advancePopulation` 既有的「新增人口預設分配為工人」慣例——
後者是「已有 farmer、經濟穩定的殖民地,人口成長 +1」的慣例;若套用在 population=1、Farmers=0 的
全新殖民地,會導致 Food=0(0 個農夫)但 FoodConsumed=1,首回合立即觸發饑荒(`recoverFromFamine`
下一回合才會修正回 farmer)。「全農」是任務指示中列出的簡單保守預設之一,選它是為了避免這個
不必要的首回合饑荒瞬間,而非有手冊依據——如需更貼近原版體驗(如新殖民地也走一次玩家可調整的
「Colonial Statistics」對話框),留待後續 UI 工作。

### 1.3 PopMax(人口上限)公式

p.55-56「Size」小節給出各尺寸的人口容量範圍(依環境浮動),但未給精確公式;精確公式來自
`openorion2/src/gamestate.cpp:2288` `GameState::planetMaxPop`:

```c
ret = ((ptr->size + 1) * 5 * climateFactor + 50) / 100;
```

(`size` 為 0-based `TINY=0..HUGE=4`;`climateFactor` = 氣候的人口容量係數,0-100,即本 remake
既有的 `gamedata.TerraformClimatePopFactorPercent`,`MANUAL_150.html` modding 附錄 `pop_climate`
參數的同一份數字。)

交叉驗證(climateFactor 代入 25 與 100 兩端,對照手冊逐段人口容量範圍):

| Size | climateFactor=25 | climateFactor=100 | 手冊原文範圍 |
|---|---|---|---|
| Tiny(0) | 1 | 5 | "1–5" |
| Small(1) | 3 | 10 | "3–10" |
| Medium(2) | 4 | 15 | "4–15" |
| Large(3) | 5 | 20 | "5–20" |
| Huge(4) | 6 | 25 | "6–25" |

逐項完全相符,高信心。已移植為 `internal/gamedata/planet_yield.go` 的 `PlanetBasePopMax(size,
climate)`,並有對應單元測試(`TestPlanetBasePopMax ManualRanges`)。

⚠ 與既有 `playerHomeworldColony()`(母星)的 PopMax=20(Large/Terran)不完全相符:代入
size=LARGE(3)、climate=TERRAN(climateFactor=80)得 `(4*5*80+50)/100=16`,非 20。母星的 20 是
`docs/tech/homeworld-init.md` 既有慣例值(可能含未拆解的起始文明加成),本次**不回頭套用**去改動
母星既有數字(避免既有經濟平衡 regression),`PlanetBasePopMax` 只用於 `ColonizeStar` 建立的
新殖民地。

## 2. 實作範圍

- `internal/shell/colonization.go`(新檔):
  - `GameSession.ColonizeStar(starIdx int) ColonizationResult`:核心引擎函式,前置條件、建立
    邏輯、平行陣列同步、消耗殖民船,見檔內逐段註解。
  - `GameSession.FleetHasColonyShip()` / `findColonyShipIndex()`:艦隊殖民船查詢。
  - `climateFromDisplay`/`gravityFromDisplay`/`mineralFromDisplay`/`sizeFromDisplay`:把
    `session.go` `genPlanets` 產生的「中文顯示字串」(該函式本是純展示用途,從未 import
    `gamedata`)轉成 `gamedata` 型別化 enum,供建構 `engine.ColonyState` 使用。這一層轉換是本次
    接線的關鍵發現——若不做這層對映,玩家在行星列表看到的氣候/礦產/重力跟殖民地實際套用的內部
    規則會是兩套不相干的資料,是需要修正的架構落差,不是可略過的枝節。
  - `climateColonizable`:見 §1.1,現階段恆真的 gate 掛勾點。
- `internal/gamedata/planet_yield.go`:新增 `PlanetBasePopMax`(見 §1.3)。
- `internal/shell/session.go`:`GameSession` 新增 `PlayerColonyStars []int` 欄位(比照既有
  `AIOpponent.ColonyStars`,平行 `PlayerColonies` 記錄各殖民地所在星索引),`NewDemoSession`
  初始化為 `[]int{0}`(母星=星0)。
- `internal/shell/ground_invasion.go`:`InvadeColony` 過戶敵方殖民地時同步 append
  `PlayerColonyStars`(先前完全沒有這個欄位,過戶的殖民地無法知道自己在哪顆星——這是拓殖前置
  的資料完整性缺口,順手補上)。
- `internal/shell/persist.go`:`PlayerColonyStars` 納入存讀檔快照。
- `cmd/moo2/interactive.go`:星系主畫面選中一顆無主星、艦隊已抵達、載有殖民船時,顯示「建立
  殖民地」按鈕(綠色,與既有派遣/入侵/載運按鈕同一互斥切換邏輯),點擊呼叫 `ColonizeStar` 並顯示
  結果訊息。**未做**:行星選擇子畫面(手冊原文的 System 視窗)——目前每星固定一顆行星,不需要
  選擇,故省略。

## 3. 已知簡化 / TODO

- 種族的 Food/Industry/Research 加成(`shell.Races[idx]`)在建立新殖民地當下手動疊加一次——因為
  `ApplyRace`/`ApplyCustomRaceBonuses` 只在新遊戲開局套一次,不會回頭套用到之後才建立的殖民地。
  這與既有 `InvadeColony` 直接複製 AI 殖民地數值(不重算玩家種族加成)是不同的簡化選擇,兩者都是
  誠實近似,標注於此供比較。
- Aquatic/Tolerant/Subterranean 等種族特性對 PopMax 的加成(`gamestate.cpp:2288` 原始函式裡的
  修飾項)未套用——本 remake 沒有種族特性追蹤系統(見 `custom-race-picks.md`),留白。
- 氣態巨星/小行星帶科技 gate(§1.1)無實際案例可測,是未來若補上這兩類行星才會啟用的掛勾點。
- 不支援「同系統多顆行星、選擇殖民哪一顆」(手冊原文的 System 視窗選擇畫面)——本 remake 每星
  固定一顆行星,本輪任務也明確排除行星選擇子畫面。
- `PlayerColonyStars` 是本次新增欄位,若讀取舊版(本次修改前)存檔,JSON 反序列化會得到 nil
  slice——下一次 `ColonizeStar`/`InvadeColony` 呼叫時會自動 padding 補齊(見兩處程式碼的
  padding loop),不會 panic,但舊存檔本身直到那之前不會有這個欄位的資料。

## 4. 驗證

- 單元測試:`internal/shell/colonization_test.go`(成功拓殖、四種前置條件擋下、拓殖後 EndTurn
  經濟正常運作、氣候/重力/礦產/大小的顯示字串對映表覆蓋率)。
- `internal/gamedata/planet_yield_test.go`:`TestPlanetBasePopMaxManualRanges`(§1.3 交叉驗證)。
- Regression 探針(跑完即刪,未提交):20 回合開局 BC 軌跡不變(101→130,與既有基準一致)、
  拓殖後新殖民地 10 回合內經濟穩定運作(人口不崩潰、不 panic)。
