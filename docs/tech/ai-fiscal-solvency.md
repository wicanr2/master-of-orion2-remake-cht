# AI 財政保底(職務分配財政保底,2026-07-11)

> 對應 `WORKLIST.md`「AI 財政赤字調整」項、commit `bff55ce`「誠實待補」段落記錄的結構性赤字。
> 延續 `docs/tech/colony-economy-maintenance.md`(玩家/AI 母星忠實 yield 接線)之後的收尾工作。
> 鐵律:誠實 > 測試綠,本輪嚴禁反推常數硬湊測試。

## 一、問題:AI(Scientific 性格)國庫無下限線性變負

`bff55ce` 把 AI 母星接上與玩家相同的忠實 Terran/Abundant yield(`playerHomeworldColony`)後,
留下一段誠實記錄的待補:AI 用 `ProfileScientific`(IndustryWeight=1,ResearchWeight=3)分配
職務時,Population=8、FoodPerFarmer=2 先分 4 個農夫餵飽人口,剩下 4 人依 1:3 比例分工/研究,
`workers = 4*1/4 = 1`。工人=1 時:

```
GrossIndustry = MoraleProductionOutput(1*3, 10) = 3*110/100 = 3
PollutionCleanupCost = 0(3 < Large 容忍值 8)
NetIndustry = 3
TaxRevenue(稅率上限 50%) = 3*50/100 = 1
Maintenance(母星起始建築維護費,newHomeworldPlayerState 設一次不再重算) = 3
NetBC = 1 - 3 = -2 / 回合(稅率未到 50% 時更低,見下方實測)
```

`Maintenance=3` 是「稅率開到手冊上限也打平不了」的固定支出——這不是稅率/係數差一點的量級
問題,是結構性的:AI 只要維持 1:3 這個比例,NetIndustry 就永遠鎖死在 3,BC 保證單調變負、
無下限。已用 `EventSeed=42` 實測(暫時性 debug test,未進版控)驗證:150 回合從 100 線性
掉到 -217,無回穩跡象。

## 二、根因定位:確認 AI 走的路徑

任務指示先確認「AI 是否也走 `RunEmpireTurn`(玩家的食物盈餘收入是否 AI 也拿得到)」。追查
`internal/engine/ai.go` `RunAIEmpireTurn`:

```go
func RunAIEmpireTurn(ps PlayerState, colonies []ColonyState, decider ai.Decider) EmpireOutput {
    ps, colonies = ApplyAIEconomy(ps, colonies, decider)
    return RunEmpireTurn(ps, colonies)
}
```

**結論:AI 走的是同一條 `RunEmpireTurn`**,`FoodSurplusRevenue`(食物盈餘收入)AI 本來就拿
得到,對稱,不需要另外接線。問題純粹出在 `ApplyAIEconomy` 決定的職務分配本身(工人太少
→ NetIndustry 太低 → 稅收上限打平不了固定維護費),不是收入來源缺漏。

### 順便發現的獨立 bug:AI 職務重分配從未寫回存檔

追查過程中發現 `internal/shell/session.go` 的 `EndTurn` 原本直接呼叫
`engine.RunAIEmpireTurn(s.AIPlayers[i].Player, s.AIPlayers[i].Colonies, ...)`,只把回傳的
`out.Player` 寫回 `s.AIPlayers[i].Player`,`ApplyAIEconomy` 重新分配過的 `colonies`(職務
分配結果)只在函式內部傳給 `RunEmpireTurn` 算完當回合經濟就丟棄,從未寫回
`s.AIPlayers[i].Colonies`(該欄位會被 `persist.go` 存檔)。

這個 bug **目前對經濟結算本身無害**:因為 AI 殖民地的 `Population`/`FoodPerFarmer` 從未被
其他機制改變(AI 沒有 `advancePopulation`),`ApplyAIEconomy` 每回合都是從同一組靜態輸入
重新算出同一個結果,「丟棄」與「保留上回合結果」在數值上沒有差異。但欄位本身是錯的——如果
未來 UI 或存檔讀取 `AIPlayers[i].Colonies.Farmers/Workers/Scientists`,看到的會是「從未更新」
的初始值(母星建構時的 4/3/1),不是決策器實際決定的分配。發現後在本輪一併修正(見四之 3)。

## 三、決策:職務分配保底(任務指示選項 1),不調食物盈餘收入(選項 2 已確認對稱、不需再接)

### 3.1 為什麼不是「調稅率上限」

`gamedata.TaxRateMaxPercent = 50` 是手冊實據(GAME_MANUAL.pdf p.37:"values range from 0%
to 50%"),不可迴避地改成更高——會失去忠實性,且不改變「NetIndustry 太低」這個根因。

### 3.2 為什麼是「職務分配保底」而非硬改 profile 權重數字

`ai.Profile` 的 `IndustryWeight`/`ResearchWeight` 是既有【設計值】(非原版精確數字,見
`internal/ai/economy.go` 檔頭聲明),直接把 Scientific 的 1:3 改成比如 1:1 可以讓測試綠燈,
但那是「反推常數湊測試」——沒有第一性原理依據,純粹是為了讓某個特定人口/yield 組合打平而
硬調。改成之後,換一個人口規模(例如人口成長後 remaining 變多),原本的 1:3 又會產生同樣的
問題(不同人口下工人數不同,固定維護費也可能不同),數字換了但問題的形狀沒變。

### 3.3 落地方案:財政保底層(不改變 profile 權重本身)

新增 `internal/ai/economy.go` 兩個函式:

- `MinWorkersForSolvency(industryPerWorker, moralePercent, maintenanceBC, maxWorkers int) int`
  ——回傳「即使把稅率開到手冊上限,稅收仍至少打平 maintenanceBC」所需的最少工人數,直接呼叫
  既有 `gamedata.MoraleProductionOutput`/`gamedata.IncomeTaxRevenue`/`gamedata.TaxRateMaxPercent`
  換算,不重新發明公式、不假造污染清理項(本專案 AI 殖民地規模下,打平所需工人的產出遠低於
  星球污染容忍值,逐步驗算見下方,忽略污染清理不影響本函式在目前場景下的正確性)。
- `DecideColonyJobsSolvent(population, foodPerFarmer, industryPerWorker, moralePercent,
  maintenanceBC int, p Profile) (farmers, workers, scientists int)` ——先照 `DecideColonyJobs`
  純比例分配,若工人數不足以打平(`MinWorkersForSolvency`),從科學家挪最少量回工人直到打平
  或科學家歸零。**只挪動打平所需的最少量**,不會把偏研究的性格整個打平成平衡型。

第一性原理:不論施政多偏好研究,殖民地稅收長期低於固定支出就是結構性赤字,國庫必然無下限
發散——這是任何理性政體都會遵守的財政常識(先確保有能力打平帳,餘力才依偏好投入研究),
不是原版數值(原版 AI 邏輯未公開,見 `docs/tech/original-ai-re.md`)。

用「稅率上限」而非「當下稅率」當保底基準,是刻意留的安全邊際:AI 的 `DecideTaxRate` 平時
多半用較低稅率(國庫充裕時 10~30%),只有國庫見底才拉滿 50%。若用當下稅率當門檻,國庫充裕
時反而會算出「不需要保底」,錯過提前佈局的機會,回到「BC 觸底才臨時反應」的被動狀態。

`ai.Decider` 介面的 `ColonyJobs` 簽章因此擴充,新增 `industryPerWorker, moralePercent,
maintenanceBC` 三個參數(`RemakeDecider.ColonyJobs` 委派給 `DecideColonyJobsSolvent`);
`engine.ApplyAIEconomy` 呼叫端補上 `cs.IndustryPerWorker, cs.MoralePercent, ps.Maintenance`。
`maintenanceBC<=0` 時保底邏輯完全不介入(等同呼叫舊版 `DecideColonyJobs`),不影響任何
`Maintenance` 尚未接線的既有測試案例(`TestApplyAIEconomy`/`TestRunAIEmpireTurn` 的
`PlayerState` 都沒設 `Maintenance`,預設零值 0,行為與改動前逐位元組相同)。

### 3.4 母星實際數字驗算

`playerHomeworldColony()`:IndustryPerWorker=3、MoralePercent=10;母星起始建築維護費
`homeworldBuildings()` 加總 = 3(海軍陸戰隊營 1 + 星基 2,`BuiltMaintenanceBC` 手冊實據)。

```
w=1: gross=MoraleProductionOutput(1*3,10)=3*110/100=3   tax=IncomeTaxRevenue(3,50)=1   <3
w=2: gross=MoraleProductionOutput(2*3,10)=6*110/100=6   tax=IncomeTaxRevenue(6,50)=3   >=3 ✓
```

`MinWorkersForSolvency(3,10,3,4) = 2`。四種既有 profile 在 Population=8/FoodPerFarmer=2
(farmers=4,remaining=4)下的保底介入結果:

| Profile | 純比例 W/S | 保底後 W/S | 是否介入 |
|---|---|---|---|
| Aggressive(3:1) | 3/1 | 3/1 | 否(已 ≥2) |
| Expansionist(2:1) | 2/2 | 2/2 | 否(已 ≥2,`4*2/3=2` 整數捨去) |
| Balanced(1:1) | 2/2 | 2/2 | 否(已 ≥2) |
| **Scientific(1:3)** | **1/3** | **2/2** | **是,挪 1 人** |

只有 Scientific 需要介入,且只挪 1 個科學家回工人——不是把研究型 AI 打平成平衡型,是「打平
帳本」所需的最小調整。

## 四、驗證結果

### 4.1 AI BC 150 回合趨勢(`EventSeed=42`,`NewDemoSession`,實測數字)

修前(Scientific 純比例分配,W=1 恆定):

```
turn=1   BC=97  TaxRate=30
turn=20  BC=43  TaxRate=50
turn=40  BC=3   TaxRate=50
turn=100 BC=-117
turn=150 BC=-217   ← 無下限線性發散,150 回合仍在加速惡化,無回穩跡象
```

修後(保底介入,W=2/S=2):

```
turn=1   BC=98  TaxRate=30  Fleet=1
turn=20  BC=60  TaxRate=30  Fleet=30
turn=30  BC=48  TaxRate=50  Fleet=45   ← BC<50 觸發稅率上限,此後打平
turn=40  BC=48  TaxRate=50  Fleet=60
turn=100 BC=48  TaxRate=50  Fleet=150
turn=150 BC=48  TaxRate=50  Fleet=225  ← BC 穩定在 48,不再變動;FleetStrength 持續成長
```

**結論:BC 從「無下限線性發散」變成「有界,穩定收斂在 48」**——國庫充裕時(BC≥50)維持
較低稅率(30%,略虧錢但可接受,國庫還有餘裕),一旦 BC 跌破 50 觸發稅率上限 50%,
`NetIndustry=6` 換算稅收剛好打平 `Maintenance=3`,此後每回合 `NetBC=0`,BC 恆定不再繼續
惡化。這是公式算出來的自然結果,不是為了打平而反推湊出的數字(48 這個具體收斂值也不是刻意
湊的——它是「BC 第一次跌破 50 觸發稅率跳到 50% 那個回合的殘值」,换一組初始 BC/稅率門檻會
收斂到不同數字,但「有界、不再發散」這個性質不會變)。

副作用:`advanceAI` 的 `FleetInvestPool` 造艦投資池,`NetIndustry` 從修前的 3 提升到修後
的 6,`FleetStrength` 成長速度也跟著變快(150 回合:修前資料未及測到相同尺度,直接對照
`bff55ce` 記錄的「turn10=7→turn120=90」與本輪「turn10=15→turn150=225」,方向一致、
幅度更快,不是回歸)。

### 4.2 既有測試處理

- `TestAIBuildsAndExpands`、`TestAIStanceHostileWhenStrong`:**未改動測試碼,原樣通過**——
  兩者都用不等式斷言(軍力有無成長、態勢是否轉敵對),保底介入只讓 AI 更強,方向不變。
  實測:`AI 軍力 0→45、佔領 7 星、態勢「提議貿易」`;`AI 關係 -40、態勢「敵視」`。
- `TestApplyAIEconomy`、`TestRunAIEmpireTurn`(`internal/engine/ai_test.go`):**未改動測試碼,
  原樣通過**——兩者的 `PlayerState` 都沒設 `Maintenance`(零值 0),保底邏輯
  `maintenanceBC<=0` 時直接不介入,行為與改動前相同。
- `TestAntaresRaidsScheduleAndEscalate`、`TestRandomEventsFireAndBounded`、
  `TestAntaresDefenseReducesDamage`:**未改動,原樣通過**——這三個測的是玩家經濟(見
  `docs/tech/colony-economy-maintenance.md` §6.6),與本輪只動 AI 職務分配無關。
- `internal/ai/decider_test.go`:`Decider.ColonyJobs` 簽章擴充(新增 3 個參數),既有兩個
  呼叫點改傳 `maintenanceBC=0`(等同舊行為),斷言數字不變。
- 新增 `internal/ai/economy_test.go` 的 `TestMinWorkersForSolvency`、
  `TestDecideColonyJobsSolvent`:直接驗證保底函式本身(逐步驗算見 §3.4),以及「Aggressive
  不該被多動」「maintenanceBC=0 時應與純比例分配相同」兩個防呆案例。

### 4.3 build/test

Docker(`moo2-ebiten:latest`)內 `go build -buildvcs=false ./...` 通過;
`go test ./internal/...` 除既有的 `internal/uifont`(headless 缺 X11 `DISPLAY`,GLFW 初始化
panic,與本輪改動無關、改動前後皆如此)外,其餘全部套件皆綠。

## 五、改動清單

- `internal/ai/economy.go`:新增 `MinWorkersForSolvency`、`DecideColonyJobsSolvent`(import
  `internal/gamedata`,純資料/公式包,無循環依賴風險)。
- `internal/ai/decider.go`:`Decider.ColonyJobs` 簽章擴充;`RemakeDecider.ColonyJobs` 改委派
  `DecideColonyJobsSolvent`。
- `internal/engine/ai.go`:`ApplyAIEconomy` 呼叫端補上 `cs.IndustryPerWorker`、
  `cs.MoralePercent`、`ps.Maintenance`。
- `internal/shell/session.go`:`EndTurn` 的 AI 迴圈改成先呼叫 `engine.ApplyAIEconomy` 並把
  結果寫回 `s.AIPlayers[i].Colonies`,再呼叫 `engine.RunEmpireTurn`(修正「職務重分配從未
  寫回存檔」的獨立 bug,見二之「順便發現」)。
- `internal/ai/decider_test.go`:兩處 `ColonyJobs` 呼叫補上新參數(`maintenanceBC=0`,行為
  不變)。
- `internal/ai/economy_test.go`:新增 `TestMinWorkersForSolvency`、
  `TestDecideColonyJobsSolvent`。

## 六、仍待補(誠實列出,不臆造)

- AI 國庫穩定在「打平」而非「持續累積盈餘」——與玩家經濟(有 `recoverFromFamine` +
  `FoodSurplusRevenue` 兩項機制疊加、300 回合終值 +702)相比,AI 目前只有保底打平,沒有更
  積極的盈餘管理(例如國庫充裕時把稅率門檻拉得更保守)。這不影響「不再無下限發散」的驗收
  目標,但若之後想讓 AI 經濟更「健康」而非「打平邊緣」,可以評估调整 `RemakeDecider` 的
  `TreasuryLow`/`TreasuryHi` 門檻,或讓保底邏輯用更高的目標稅率(如 30%)而非稅率上限
  (50%)——本輪刻意不做這個加碼,保持改動範圍在「解決結構性赤字」本身,避免另外引入
  未經驗證的新設計判斷。
- `MinWorkersForSolvency` 忽略污染清理成本:本專案 AI 殖民地規模下不影響正確性(§3.3 已
  說明),但若未來 AI 出現高工業大型殖民地(污染清理成本不再是 0),這個保底計算會低估
  實際所需稅收(因為忽略清理成本會讓估計的 NetIndustry 偏高),需要之後連同容忍值/清理
  成本一併重新驗算。
- AI 母星 `Population` 恆定不變(AI 沒有 `advancePopulation`)——本輪沒有變動這件事,只是
  在調查「職務重分配為何從未寫回」時確認了這個既有限制仍然成立,列此避免下一輪誤以為
  已經修好。
