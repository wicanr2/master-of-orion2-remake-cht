# 殖民地維護費 / 行星驅動 yield 接線調查(2026-07-10)

> 目的:把殖民地經濟從「placeholder 常數」換成忠實模型——維護費由已建建築算、食物/工業產出
> 由行星屬性(氣候/礦產)算。本文件記錄本輪**已落地**的部分,以及**已試接但發現會讓經濟進入
> 無法恢復的破產迴圈、因此刻意回退**的部分與其證據,供下一輪接手判斷優先順序。
>
> 鐵律:誠實 > 測試綠。本輪嚴禁反推常數硬湊測試,寧可少做、留 TODO,也不臆造/掩蓋。

## 一、已落地:建築驅動維護費

**改動前**:`internal/shell/session.go` 的 `newHomeworldPlayerState`、`cmd/moo2sim` 各處把
`Player.Maintenance` 寫死為 `5`——查無手冊或存檔依據,純 remake placeholder。

**改動後**:

- `internal/gamedata/buildings.go` 新增 `BuiltMaintenanceBC(built map[string]bool) int`:
  對一組「已建成」建築(中文名為 key)加總 `Buildings` 表的 `MaintenanceBC` 欄位。40 項建築的
  `MaintenanceBC` 全數為手冊實據(見該檔檔頭聲明,唯一「估計值」欄位是 `ProductionCost`,與
  `MaintenanceBC` 無關)。
- `internal/shell/session.go`:
  - `newHomeworldPlayerState` 的起始 `Maintenance` 改為
    `gamedata.BuiltMaintenanceBC(homeworldBuildings())` —— 母星起始只有「海軍陸戰隊營」(1 BC)
    +「星基」(2 BC)= **3 BC/回合**(取代無據的 5)。
  - 新增 `(s *GameSession) totalBuildingMaintenance() int`:加總 `s.ColonyBuildings` 全部殖民地
    已建成建築的維護費。
  - `EndTurn()` 開頭改為 `s.Player.Maintenance = s.totalBuildingMaintenance()`,每回合依「玩家當下
    實際已建成建築清單」重算,取代常數——玩家日後蓋更多建築,維護費會如實增加。

**未涵蓋(誠實列出,不臆造)**:艦艇/間諜/軍官維護費。手冊有「每艘使用中運輸艦 0.5 BC/回合」
（`gamedata.IncomeFreighterMaintenanceCost`,已實作待接線)與指揮評等超支費用
（`gamedata.IncomeCommandOverflowCost`,同樣已實作待接線),但本專案目前完全不追蹤運輸艦數量,
`Ships` 只有作戰艦艇,沒有可推導的模型,故本輪不計入,標 TODO。AI 對手（`AIOpponent`)沒有
`ColonyBuildings` 追蹤機制,`Maintenance` 由 `newHomeworldPlayerState` 設一次後不再重算——這對
AI 是誠實的(AI 的建築集合在本專案裡本就從未變動過),不是遺漏。

**影響**:3 BC/回合 < 舊的 5 BC/回合,對經濟餘裕只有正面影響,不會讓任何既有測試變差
(已跑過完整 `go test ./internal/...`,`internal/shell` 全綠,見下方「五、驗證結果」)。

## 二、試接但回退:行星驅動 yield(Terran/Abundant 母星基準)

`internal/gamedata/planet_yield.go`(前一輪已完成)提供手冊有據的
`ClimateFoodPerFarmer(TERRAN)=2`、`MineralIndustryPerWorker(ABUNDANT)=3`。本輪指示是把
`averageHomeworldColony()` 的 `FoodPerFarmer:4, IndustryPerWorker:6`(無據 placeholder)換成
這兩個查表值。**已實際試接、實測、且回退**,原因如下。

### 2.1 靜態開局數字:剛好打平,不是舒服的正值

Population=8。若沿用舊的 Farmers=3/Workers=4,新 FoodPerFarmer=2 只夠 3×2=6 食物,餵不飽
8 人口(結構性饑荒)。調整為 Farmers=4/Workers=3(人口分配這是機械必要調整,理由見程式碼
`averageHomeworldColony` 註解,不是為了湊測試反推)後:

| 項目 | 數值 | 備註 |
|---|---|---|
| Food | 4×2=8,士氣+10%→8(整數捨去) | 剛好等於消耗 |
| FoodConsumed | 8×1=8 | Population×1 |
| **FoodSurplus** | **0** | 打平,零緩衝 |
| GrossIndustry | 3×3=9,士氣+10%→9 | |
| PollutionCleanupCost | (9-8)/2=0(Large 星球容忍值8) | |
| **NetIndustry** | **9** | |
| TaxRevenue(稅率40%) | 9×40/100=3 | |
| **Maintenance(建築,見上節)** | **3** | |
| **NetBC(第1回合)** | **0** | 打平,不是負的,但也沒有任何緩衝 |

單看這張表,`0 ≥ 0`,字面上符合「可持續」的門檻。但這只是**沒有任何隨機事件發生**時的靜態快照。

### 2.2 動態實測:零緩衝經濟撞上既有饑荒-鎖死機制,保證破產

用固定 `EventSeed=42`(`TestRandomEventsFireAndBounded`/`TestAntaresRaidsScheduleAndEscalate`
既有測試用的同一顆種子)實際跑 100~300 回合(用暫時性 debug test,未進版控),觀察到:

1. 隨機事件(瘟疫/隕石)與安塔蘭入侵都會扣殖民地人口,扣除順序（`losePop`/`advanceAntares`現有邏輯)
   是「扣人數最多的職務」,**不保證留下至少 1 個農夫**。
2. 一旦連續幾次事件把 Population 打到只剩 1 人,而那 1 人剛好是科學家(不是農夫)——
   Food=0,FoodConsumed=1,FoodSurplus=-1,`Starving=true`。
3. `internal/engine/colony.go` 的 `colonyGrowth`:「饑荒時（`foodSurplus<0`)不套用成長公式,回 0」
   ——這個回合起,人口成長**永久停擺**,因為食物赤字不會自己恢復(沒有農夫,永遠 0 食物)。
4. 本專案**沒有任何機制**把工人/科學家重新指派回農夫(`ShiftColonyJob` 是玩家手動 UI 操作,
   AI/自動流程不會做這件事)。於是殖民地卡死在 Workers=0 → NetIndustry=0 → TaxRevenue=0,
   而建築維護費(3 BC)仍然每回合照扣——**BC 保證單調遞減至負值**,以固定種子重現、非機率僥倖
   （已用 debug trace 逐回合列印驗證,見下方數字節錄)。

實測節錄(玩家母星,`s.Ships=nil`,`EventSeed=42`,忠實 yield 已接上時的軌跡):

```
turn=50  F=2 W=0 S=1 pop=3   (瘟疫+安塔蘭同回合,人口被打到剩3)
turn=53  F=1 W=0 S=1 pop=2   (隕石又扣1)
turn=54  F=0 W=0 S=1 pop=1   (瘟疫再扣1,農夫歸零——饑荒鎖死起點)
turn=55~101: NetInd=0, TaxRev=0, Maint=3, NetBC=-3 每回合,BC 持續走負
turn=77: BC=-3(跌破0)
```

### 2.3 結論與判斷

這不是「稅率/維護費算式差一點」的量級問題,而是「Terran/Abundant 基準本身零食物緩衝」+
「饑荒鎖死機制(成長公式停擺、無自動改派農夫)」疊加後的**結構性死結**。任務指示的判斷準則
（`若 net BC ≥ 0 → 更新測試;若 < 0 → 誠實回報、縮小改動範圍`)是以「開局第一回合靜態值」界定,
但實測顯示這個界定不足以代表真正的可玩性——單看靜態值會誤判為「可持續」,實際跑完整套既有
測試(含隨機事件、安塔蘭入侵)有 4 個測試因此炸裂:`TestAntaresRaidsScheduleAndEscalate`、
`TestRandomEventsFireAndBounded` 直接因 BC 轉負失敗;`TestAIBuildsAndExpands`、
`TestAIStanceHostileWhenStrong` 則是另一個由「數值變小」牽連出的獨立問題(見下節)。

依「誠實 > 測試綠」與「改動縮到安全子集」的指示,**本輪回退**
`averageHomeworldColony()` 的 Farmers/Workers/FoodPerFarmer/IndustryPerWorker,維持原本的
`4/6`(仍是無據 placeholder,誠實標註,不假裝已修好),只保留建築維護費那部分的忠實改動。

**要真正把 Terran/Abundant 接上去,建議的前置工作(任一項先做都能解掉零緩衝問題)**:

1. 接上 `gamedata.TradeGoodsIncome`/`IncomeFoodSurplusRevenue`(手冊有據,已寫好在
   `internal/gamedata/income.go`,目前**全專案零呼叫處**,純被動代碼)——多一條收入來源,
   NetBC 才有正緩衝可以吸收事件衝擊。
2. 或補上「饑荒自動復原」機制:Population>0 但 Farmers=0 且 Starving 時,自動把 1 個
   Worker/Scientist 改派回 Farmer(邏輯上等同玩家會做的事),避免永久鎖死。
3. 兩者都不做的話,至少把隨機事件/安塔蘭的人口扣除順序改成「保底留 1 個農夫」,降低死鎖機率
   （治標,不解決零緩衝本身)。

## 三、順便發現(非本輪任務範圍,列出供評估):AI 艦隊投資的整數捨去 bug

追查 `TestAIBuildsAndExpands` 失敗原因時發現:`session.go` 的 `advanceAI`——

```go
invest := 4 // Scientific 性格 IndustryWeight(1) 不 > ResearchWeight(3),維持 4
if out.TotalNetIndustry > 0 {
    a.FleetStrength += out.TotalNetIndustry / invest
}
```

當 `TotalNetIndustry` 小(例如 AI 用 Scientific 性格分配後 Workers 只剩 1,NetIndustry=3)時,
`3/4=0`(Go 整數除法無條件捨去),`FleetStrength` **永遠 +0**,不是「累積變慢」而是「完全停滯」
——這是既有的整數捨去 bug,舊的高 IndustryPerWorker(6)placeholder 數值恰好大到能整除出
非零結果(掩蓋了這個 bug),換成寫實的低數值後才會暴露。**這與維護費/yield 接線無關**,是
`advanceAI` 既有邏輯本身的缺陷,建議之後另開任務用「累加餘數」的方式修(類似
`advancePopulation` 的 `popAccum` 模式),不屬於本輪任務授權範圍(未動它),記錄於此避免下一輪
遺忘。

## 四、改動清單(本輪實際落地部分)

- `internal/gamedata/buildings.go`:新增 `BuiltMaintenanceBC`。
- `internal/shell/session.go`:
  - `newHomeworldPlayerState` 起始 `Maintenance` 改為建築加總(3,取代無據 5)。
  - 新增 `totalBuildingMaintenance`,`EndTurn` 每回合重算 `Player.Maintenance`。
  - `averageHomeworldColony` **維持原樣**(4/6 placeholder),補充註解記錄「已試接 Terran/
    Abundant 但回退」的完整理由與證據指標(指向本文件)。
- `internal/engine/engine.go`:`PlayerState.Maintenance` 欄位註解更新,說明目前只計入建築,
  艦艇/間諜/軍官維護費待補。

## 五、驗證結果

Docker 內 `go build -buildvcs=false ./...` 通過;`go test ./internal/...`:除既有的
`internal/uifont`(此包測試在 headless docker 內因缺 X11 `DISPLAY`,GLFW 初始化 panic,
與本輪改動無關、改動前後皆如此)外,其餘全部套件（含 `internal/shell` 全部既有測試,含
`TestAIBuildsAndExpands`/`TestAIStanceHostileWhenStrong`/`TestAntaresRaidsScheduleAndEscalate`/
`TestAntaresDefenseReducesDamage`/`TestRandomEventsFireAndBounded`)皆綠,無需修改任何測試期望值
——因為本輪最終落地的改動(建築驅動維護費)只讓經濟餘裕變好,不會讓任何既有場景變差。
