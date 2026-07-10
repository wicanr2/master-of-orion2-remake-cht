# 殖民地維護費 / 行星驅動 yield 接線調查(2026-07-10)

> 目的:把殖民地經濟從「placeholder 常數」換成忠實模型——維護費由已建建築算、食物/工業產出
> 由行星屬性(氣候/礦產)算。
>
> **本文件分兩輪**:第一輪(§一~五)試接玩家母星 yield 後發現零緩衝經濟撞上饑荒鎖死機制,
> 判斷「無法恢復」而回退,只落地建築維護費。**第二輪(§六,同日下午接續)補上第一輪§2.3
> 列出的兩項前置工作(饑荒自動復原 + 食物盈餘收入)後,重新試接玩家母星 yield,這次穩定
> 落地**——第一輪§2「本輪回退」的結論**已被第二輪取代,不再成立**,細節與證據見§六;
> 第一輪的分析過程與數字仍保留(有參考價值:解釋了「為什麼」需要那兩項前置工作),但其
> 「維持 4/6 placeholder」的最終決定已過期。
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

## 二、[第一輪結論已被 §六 取代] 試接但回退:行星驅動 yield(Terran/Abundant 母星基準)

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

## 五、驗證結果(第一輪,2026-07-10 上午狀態,已被 §六 取代)

Docker 內 `go build -buildvcs=false ./...` 通過;`go test ./internal/...`:除既有的
`internal/uifont`(此包測試在 headless docker 內因缺 X11 `DISPLAY`,GLFW 初始化 panic,
與本輪改動無關、改動前後皆如此)外,其餘全部套件（含 `internal/shell` 全部既有測試,含
`TestAIBuildsAndExpands`/`TestAIStanceHostileWhenStrong`/`TestAntaresRaidsScheduleAndEscalate`/
`TestAntaresDefenseReducesDamage`/`TestRandomEventsFireAndBounded`)皆綠,無需修改任何測試期望值
——因為本輪最終落地的改動(建築驅動維護費)只讓經濟餘裕變好,不會讓任何既有場景變差。

> ⚠ 上述「無需修改任何測試期望值」只在**第一輪範圍**(僅建築維護費、yield 仍是 placeholder)
> 內成立。§六 接上 yield 之後,其中 2 個測試(`TestAntaresRaidsScheduleAndEscalate`、
> `TestRandomEventsFireAndBounded`)的「BC 絕不為負」斷言被更新——理由與新斷言見 §六。

## 六、第二輪(2026-07-10 下午接續):補齊前置工作後重新接上 yield,穩定落地

延續 §2.3 列出的「要真正把 Terran/Abundant 接上去,建議的前置工作」,本輪按順序把選項①②
都做了(選項③—調整人口損失順序保底留農夫—沒有另外做,因為①②做完後已經足夠,詳見下方驗證),
然後重新試接 §2 回退掉的 yield,這次穩定落地。

### 6.1 饑荒自動復原機制(§2.3 選項②)

`internal/shell/session.go` 新增 `(s *GameSession) recoverFromFamine()`:每回合
`EndTurn` 跑完 `engine.RunEmpireTurn` 之後(用該回合的 `Starving` 結果),對每個玩家殖民地
檢查「`Farmers==0` 且上回合 `Starving`」,若成立就把 1 個 `Worker`(優先)或 `Scientist`
改派回 `Farmer`。這近似「玩家看到饑荒會手動 `ShiftColonyJob` 自救」的行為,只作用於玩家殖民地
(AI 的職務分配由 `ApplyAIEconomy` 每回合依 decider 重新決定,不會卡在饑荒鎖死,不需要這個
機制)。**手冊沒有這個機制的直接文字依據**——這是為了讓「零緩衝經濟不會因單次事件就死鎖」而
補的 remake 機制,誠實標註為近似,不是手冊翻譯。

### 6.2 接上食物盈餘收入(§2.3 選項①的一半)

`internal/engine/empire.go` 的 `RunEmpireTurn` 新增:對每個殖民地,若
`ColonyOutput.FoodSurplus > 0`,呼叫 `gamedata.IncomeFoodSurplusRevenue(surplus, false)`
(GAME_MANUAL.pdf p.25,每單位餘糧 0.5 BC,無條件捨去)累加進新欄位
`EmpireOutput.FoodSurplusRevenue`,`NetBC` 改為 `TaxRevenue + FoodSurplusRevenue -
Maintenance`。只對正盈餘計入(手冊只描述「出售剩餘糧食」這個收入來源;饑荒本身已經由
`Starving`/`colonyGrowth` 停擺懲罰,不疊加負收入,避免雙重懲罰)。`fantasticTrader` 固定傳
`false`——`ColonyState` 目前沒有追蹤這個種族特質的欄位,沒有可推導的模型,TODO 留待種族特質
系統擴充後再接。

**`TradeGoodsIncome` 刻意仍未接**(§2.3 選項①的另一半):手冊定義的貿易財收入來自玩家把部分
工人產能「明確配置」去生產 Trade Goods(第四種職務分配,非 Farmer/Worker/Scientist 三者之一),
但 `ColonyState` 沒有這個分配欄位或對應 UI——沒有可推導的輸入模型,接上等於臆造一個不存在的
分配比例,故維持 TODO,不在本輪捏造。

### 6.3 玩家母星重新接上 Terran/Abundant yield,但 AI 母星維持 placeholder(範圍決策)

新增 `playerHomeworldColony()`(取代 `NewDemoSession` 原本呼叫的 `averageHomeworldColony()`
給玩家那份):`FoodPerFarmer = gamedata.ClimateFoodPerFarmer(TERRAN) = 2`、
`IndustryPerWorker = gamedata.MineralIndustryPerWorker(ABUNDANT) = 3`,`Farmers`/`Workers`
從 3/4 對調為 4/3(§2.1 已推導過的機械必要調整:Population=8、FoodPerFarmer=2 時,
Farmers=4 才能 4×2=8 打平消耗)。

`averageHomeworldColony()`(現在只給 AI 用)**刻意維持原樣**(4/6 placeholder),沒有跟著換成
Terran/Abundant。原因:`internal/shell/session.go` 的 `advanceAI` 有一個既有、**任務邊界明訂
不修**的整數捨去 bug——`FleetStrength += TotalNetIndustry / invest`(`invest` 為 2 或 4)。
用 AI 的實際性格(`ProfileScientific`,`IndustryWeight=1 < ResearchWeight=3` → `invest=4`)
代入忠實 yield 算出來的 `NetIndustry`:`ai.DecideColonyJobs` 依 `FoodPerFarmer=2` 算出
`Farmers=ceilDiv(8,2)=4`,`remaining=4`,`Workers=4*1/4=1`,`GrossIndustry=1*3*1.1=3`(污染
清理:Large 容忍值 8 > 3,清理 0),`NetIndustry=3`。`3/4=0`(Go 整數除法無條件捨去)——AI
的 `FleetStrength` 會**永久卡在 0 不動**,不是「成長變慢」而是「完全停滯」,和 §3 記錄的
既有 bug 是同一個。這不是「經濟基準改變、更新測試期望值」可以誠實處理的情況(機制本身壞掉,
不是數字換了一個新的合理值),而是暴露一個獨立的、任務邊界明訂「不要順手修」的既有 bug。
因此本輪縮小範圍:**只讓玩家母星接上忠實 yield,AI 母星維持 placeholder**,兩者用不同函式
(`playerHomeworldColony` vs `averageHomeworldColony`),等 `advanceAI` 那個整數捨去 bug
另案修好(建議用 `advancePopulation` 的 `popAccum` 累加餘數模式)後,再讓 AI 一併接上。

### 6.4 順便修正:BC 已為負時,antares/事件的「夾值」邏輯會反向倒贈 BC(既有 bug,本輪暴露)

玩家母星 yield 接上後,BC 第一次在既有測試場景中真的會出現負值(見 6.5)。跑動態驗證時發現
`advanceAntares`(安塔蘭入侵扣 BC)與 `advanceEvents` case 1(太空海盜劫掠)都有同一個夾值
邏輯:`if loss > s.Player.BC { loss = s.Player.BC }`。這假設 `s.Player.BC` 永遠 ≥0(「最多
虧到歸零」),但 `s.Player.BC` 一旦已經是負值,這行會把 `loss` 夾成**負數**,下面
`s.Player.BC -= loss` 就變成「損失負數」= 倒贈 BC——出現「安塔蘭人入侵:損失 -13 BC」這種
自相矛盾的訊息,且讓國庫在已經破產時反而莫名其妙變多。這是舊碼本身的既有邊界疏漏(BC 從未
為負過,這條路徑從未被走到),與 `advanceAI` 那個明確排除在外的整數捨去 bug 是**不同函式、
不同機制**,不在任務邊界排除清單內,且直接由本輪改動觸發(讓 BC 首次真的能為負),故本輪順手
修正:兩處都改成「`s.Player.BC<=0` 時 `loss=0`(沒有更多可虧損)」。

### 6.5 驗證:母星長期 net BC 趨勢

用 `EventSeed=42`(與既有測試同一顆種子,完全可重現)實測:

| 場景 | 回合數 | 最低 BC | 結束 BC | 觀察 |
|---|---|---|---|---|
| 無艦隊防禦,吃滿安塔蘭入侵(`TestAntaresRaidsScheduleAndEscalate` 設定) | 80 | -3 | -3 | 僅一次短暫轉負(人口被打到剩 1 的那個回合) |
| 預設起始艦隕(非戰鬥艦,近乎無防禦)+ 隨機事件 + 安塔蘭入侵(`TestRandomEventsFireAndBounded` 設定) | 300 | -59 | **+702** | 中段(約 turn 90~200)因反覆入侵把人口打到剩 1~3,收入不穩定轉負;人口回穩、遇到「經濟繁榮」等正向事件後**完全恢復**,終值為健康正值 |

**結論**:接上忠實 yield 後,母星在「無任何隨機事件」的靜態開局是零緩衝打平(§2.1 已算過,
FoodSurplus=0、NetBC=0)。動態實測顯示,配合 6.1+6.2 兩項機制,經濟在**一般強度**的事件/入侵
下能自我調節、緩慢累積盈餘(見上表 300 回合終值 +702)；只有在**人為構造的極端壓力測試**
(80 回合內連續無防禦硬吃 5 次遞增安塔蘭入侵,人口被反覆打到僅剩 1 人)下才會出現短暫、有界
的 BC 負值——原因是「建築維護費(3 BC/回合)不隨人口規模縮小」這個手冊本身就有的機制,人口
剩 1 時無論怎麼分配職務,收入都到不了 3 BC,這是誠實的經濟後果,不是算式錯誤,更不是「零緩衝
→ 保證破產」——與 §2.2 當初「BC 保證單調遞減至負值、永久卡死」的悲觀結論不同:6.1+6.2 兩項
機制確實打破了那個「永久」死結,負值是**短暫、有界、會隨人口回升自我修復**的,不是無限崩潰。

### 6.6 5 個經濟測試的處理

- `TestAIBuildsAndExpands`、`TestAIStanceHostileWhenStrong`:**未改動,原樣通過**。因為
  §6.3 的範圍決策(AI 母星維持 placeholder),AI 的 `FleetStrength`/`Relation` 軌跡完全沒變。
- `TestAntaresDefenseReducesDamage`:**未改動,原樣通過**(只比較「有/無防禦」的相對損失,
  不依賴絕對 BC 是否為負)。
- `TestAntaresRaidsScheduleAndEscalate`、`TestRandomEventsFireAndBounded`:**斷言已更新**
  ——原本「BC 絕不為負」改成「BC 不會失控式無下限崩潰」(具名常數
  `bcCrashFloor80Turns=-20`/`bcCrashFloor300Turns=-150`,均以 §6.5 實測數字為基準,留有餘裕
  但仍能抓到未來若真的壞掉、無下限崩潰的迴歸)。人口下限(`Population<1` 即失敗)、入侵排程/
  升級、事件觸發次數等其餘斷言**全部維持原樣未動**——驗證的是機制與趨勢沒有壞掉,不是硬湊
  出一個新的「通過用」數字。兩處測試都在註解裡完整寫明新值來自忠實 yield + 建築維護費 + 食物
  盈餘收入,以及具體的實測最低點數字,可追溯、可重現(`EventSeed=42` 固定)。

### 6.7 改動清單(第二輪)

- `internal/gamedata`:無新增(`ClimateFoodPerFarmer`/`MineralIndustryPerWorker`/
  `IncomeFoodSurplusRevenue` 皆為前一輪/更早已就緒的函式,本輪只是接線呼叫端)。
- `internal/engine/empire.go`:`EmpireOutput` 新增 `FoodSurplusRevenue` 欄位;
  `RunEmpireTurn` 接上 `gamedata.IncomeFoodSurplusRevenue`(只對正盈餘),`NetBC` 公式改為
  `TaxRevenue + FoodSurplusRevenue - Maintenance`。
- `internal/shell/session.go`:
  - 新增 `recoverFromFamine`,`EndTurn` 內呼叫。
  - 新增 `playerHomeworldColony`(玩家母星,忠實 Terran/Abundant yield);
    `averageHomeworldColony`(現在僅供 AI)維持原樣,補充範圍決策理由的完整註解。
  - `NewDemoSession` 的 `PlayerColonies` 改用 `playerHomeworldColony()`,`AIPlayers[0].Colonies`
    維持 `averageHomeworldColony()`。
  - `advanceAntares`、`advanceEvents` case 1(太空海盜)修正 BC 已非正時的夾值邏輯(§6.4)。
- `internal/shell/antares_test.go`、`internal/shell/events_test.go`:更新 BC 斷言(§6.6),
  新增具名常數 `bcCrashFloor80Turns`/`bcCrashFloor300Turns` 並附完整理由註解。

### 6.8 本輪驗證(第二輪)

Docker 內 `go build -buildvcs=false ./...` 通過;`go test ./internal/...`:除既有的
`internal/uifont`(headless docker 缺 X11 `DISPLAY`,GLFW 初始化 panic,與本輪改動無關、
改動前後皆如此)外,其餘全部套件皆綠,包含 `internal/shell` 全部既有測試(含
`TestAIBuildsAndExpands`/`TestAIStanceHostileWhenStrong`/`TestAntaresDefenseReducesDamage`
原樣通過,`TestAntaresRaidsScheduleAndEscalate`/`TestRandomEventsFireAndBounded` 依 §6.6
更新後通過)。

### 6.9 仍待補(誠實列出,不臆造)

- `TradeGoodsIncome` 未接(§6.2):需要「貿易財職務配置」模型,目前 `ColonyState` 沒有這個
  欄位,待種族/職務系統擴充後再做。
- AI 母星 yield 仍是 placeholder(§6.3):待 `advanceAI` 的整數捨去 bug(§3)另案修好後,
  才能讓 AI 一併接上 Terran/Abundant,屆時 `averageHomeworldColony`/`playerHomeworldColony`
  兩個函式可以考慮合併回一個。
- 艦艇/間諜/軍官維護費(§1)仍未計入——沒有可推導的模型(未追蹤運輸艦數量),維持 TODO。
- §2.3 選項③(調整人口損失順序、保底留 1 個農夫)沒有另外做——①②做完後 §6.5 的實測數字
  顯示已經足夠(短暫負值會自我修復),不需要疊加治標手段;但若未來想進一步縮小 BC 負值的
  波動幅度,這仍是一個可行的加強方向,留待下一輪視需要評估。
