# AI 決策模式與架構

本文件說明 `internal/ai`(AI 決策層)與 `internal/diplomacy`(外交關係層)的架構與決策邏輯,以原始碼為準(`internal/ai/*.go`、`internal/diplomacy/*.go`,不含 `_test.go`)。

## 1. 兩種 AI 模式

`internal/ai/decider.go` 定義了 `AIMode`,讓 AI 決策邏輯可像主選單的 1.3/1.5 版本選擇一樣,由玩家在 remake 與 original 之間切換:

| 模式 | `AIMode` 常數 | 現況 | 說明 |
|---|---|---|---|
| 設計性重建 | `ModeRemake` | 已實作 | `internal/ai` 這套啟發式,權重/門檻皆為本專案設計值,非原版數值。 |
| 逆向原版 | `ModeOriginal` | 待 RE,fallback | 目標是還原原版 MOO2 AI 行為;官方手冊與英語社群公認未破解該邏輯(見 `community-mechanics-findings.md`),尚無可移植的權威資料。 |

統一介面是 `Decider`(`decider.go:29`),六個方法涵蓋殖民地工作分配、稅率、研究選題、生產優先、外交姿態與模式回報:

```go
type Decider interface {
    ColonyJobs(population, foodPerFarmer int) (farmers, workers, scientists int)
    TaxRate(treasuryBC int) int
    ResearchTopic(candidates []ResearchCandidate) int
    BuildPriority(threatenedByEnemy, hasColonizableTarget, infrastructureComplete bool) BuildPriority
    Stance(level diplomacy.RelationLevel) Stance
    Mode() AIMode
}
```

`RemakeDecider` 是目前唯一的實作,內部把六個方法委派回 `economy.go` / `research.go` / `military.go` / `diplomacy.go` 的既有函式,並多帶一組國庫門檻(`TreasuryLow=50`、`TreasuryHi=300`,設計值)。

工廠函式 `NewDecider(mode AIMode, p Profile) (d Decider, ok bool)`:

- `mode == ModeRemake`:回傳 `RemakeDecider`,`ok=true`。
- `mode == ModeOriginal`:**目前仍回傳 `RemakeDecider`(以 remake 啟發式代打)**,`ok=false`。呼叫端應以 `ok` 判斷是否要提示玩家「original 模式尚未提供,暫以 remake 邏輯執行」。

尚待完成的部分是新增 `OriginalDecider`(見第 3 節架構圖)並在 `NewDecider` 的 `ModeOriginal` 分支回傳它;在此之前,`ModeOriginal` 只是型別上存在的「保留位」,行為與 `ModeRemake` 完全相同。

### 目前的接線狀態

`internal/engine/ai.go` 是把 `internal/ai` 接進回合引擎的地方,目前**只接了經濟決策**:`ApplyAIEconomy`/`RunAIEmpireTurn` 透過 `ai.Decider` 介面呼叫 `ColonyJobs`/`TaxRate`(已改為經介面注入,支援 remake/original 模式選擇)。研究選題(`ResearchTopic`)、生產優先(`BuildPriority`)、外交姿態(`Stance`)三個介面方法目前**尚無呼叫端**接入回合引擎;主選單(`cmd/moo2/menu.go`)也還沒有 AI 模式的選項。換言之,`Decider`/`AIMode` 抽象層已就緒且經濟決策已走介面,其餘決策面向串接到回合流程與 UI 仍是後續工作。

## 2. remake AI 的決策邏輯

以下逐項列出程式碼中的判斷次序與設計值,**全部標【設計性重建,非原版】**——MOO2 官方手冊未給、社群逆向也未破解 AI 決策的實際規則(見 `docs/tech/community-mechanics-findings.md`)。

### 2.1 經濟(`economy.go`)

#### 工作分配 `DecideColonyJobs(population, foodPerFarmer, profile)`

1. 先分配足夠農夫餵飽全體人口:`farmers = ceil(population / foodPerFarmer)`(`foodPerFarmer <= 0` 時,如純機器人殖民地,不分配農夫)。
2. 若 `farmers` 超過人口數,夾限為人口數。
3. 剩餘人口 `remaining = population - farmers` 依 `Profile` 的工業/研究權重比例分配:
   `workers = remaining * IndustryWeight / (IndustryWeight + ResearchWeight)`,餘數全給 `scientists`。

| 輸入範例 | 結果 (農/工/科) |
|---|---|
| 10 人口、5 食物/農夫、balanced(1:1) | 2 / 4 / 4 |
| 10 人口、5 食物/農夫、aggressive(3:1) | 2 / 6 / 2 |
| 10 人口、5 食物/農夫、scientific(1:3) | 2 / 2 / 6 |
| 10 人口、0 食物/農夫(無法務農) | 0 / 5 / 5(依 balanced 權重) |

#### 稅率 `DecideTaxRate(treasuryBC, lowThreshold, highThreshold)`

| 國庫狀態 | 稅率 |
|---|---|
| `< lowThreshold`(警戒) | 50% |
| `[lowThreshold, highThreshold)`(中等) | 30% |
| `>= highThreshold`(充裕) | 10% |

三段稅率對齊 gamedata 稅率規則的 10% 級距。`engine/ai.go` 目前固定用 `lowThreshold=50`、`highThreshold=300`。

### 2.2 研究選題(`research.go`)

`DecideResearchTopic(candidates, profile)` 依 `Profile.IndustryWeight` 對 `ResearchWeight` 的相對大小分三支:

| 性格傾向 | 判斷條件 | 選題策略 |
|---|---|---|
| 好戰型 | `IndustryWeight > ResearchWeight` | 在「軍事領域」候選中選**成本最低**者(快速取得戰力);若軍事領域無候選,退回全體候選選最低成本 |
| 科學型 | `ResearchWeight > IndustryWeight` | 選**成本最高**者(長線投資高階科技) |
| 平衡型 | 權重相等 | 選**成本最低**者(穩定推進科技樹,避免研究隊列停滯) |

「軍事領域」是本套件的設計分類(`militaryAreas`,對照 gamedata 的 `ResearchArea` 常數值,`ai` 套件不 import gamedata、僅鏡射常數值以維持解耦):

| AreaIndex | 領域 | 視為軍事? |
|---|---|---|
| 0 | Biology | 否 |
| 1 | Power | 是 |
| 2 | Physics | 是 |
| 3 | Construction | 是 |
| 4 | Fields | 是 |
| 5 | Chemistry | 是 |
| 6 | Computers | 否 |
| 7 | Sociology | 否 |

同成本並列時保留候選清單中先出現者(結果可預期、可測試)。無候選時回傳 `-1`。

### 2.3 生產優先(`military.go`)

`DecideBuildPriority(profile, threatenedByEnemy, hasColonizableTarget, infrastructureComplete)` 依固定次序判斷,前者成立即回傳:

| 順位 | 條件 | 決策 |
|---|---|---|
| 1 | 受敵對勢力威脅 | 好戰性格(`IndustryWeight > ResearchWeight`)→ `BuildWarships`;其餘 → `BuildDefenses` |
| 2 | 基礎建設未完成 | `BuildColonyInfrastructure`(不論性格,優先度高於擴張與預設軍備) |
| 3 | 有可殖民目標 且 性格傾向擴張(`Profile.Name` 為 `expansionist` 或 `balanced`) | `BuildColonyShip` |
| 4(預設) | 以上皆不成立 | 好戰性格 → `BuildWarships`;其餘 → `BuildColonyInfrastructure` |

`BuildPriority` 四個列舉值:`BuildColonyInfrastructure`(殖民建設)、`BuildWarships`(戰艦)、`BuildColonyShip`(殖民船)、`BuildDefenses`(防禦設施)。

好戰判斷準則(`IndustryWeight > ResearchWeight`)與外交姿態(2.4 節)共用同一設計準則。「傾向擴張」的判斷以 `Profile.Name` 字串比對 `economy.go` 定義的 `ProfileExpansionist` / `ProfileBalanced`。

### 2.4 外交姿態(`diplomacy.go`)

`DecideStance(level, profile)` 依關係等級(見第 4 節)與性格決定:

| 關係區間 | 好戰性格 | 非好戰性格 |
|---|---|---|
| 敵對(`< RelationWary`,`IsHostile()`) | `StanceWar` | `StanceHostile` |
| 友好(`> RelationAffable`,`IsFriendly()`),且 `>= RelationUnity` | `StanceProposeTrade` | `StanceProposeAlliance` |
| 友好但 `< RelationUnity` | `StanceProposeTrade` | `StanceProposeTrade` |
| 中立區間 | `StanceNeutral`(伺機而動) | `StanceProposeTrade` |

`Stance` 五個列舉值:`StanceWar`、`StanceHostile`、`StanceNeutral`、`StanceProposeTrade`、`StanceProposeAlliance`。

### 2.5 四種性格 Profile(`economy.go`)

`Profile{Name, IndustryWeight, ResearchWeight}` 是驅動上述四個決策面向的唯一參數。目前預設四種:

| Profile | IndustryWeight | ResearchWeight | 好戰?(`Industry>Research`) | 傾向 |
|---|---|---|---|---|
| `ProfileAggressive`("aggressive") | 3 | 1 | 是 | 重工業(造艦),研究選最便宜軍事科技,受威脅時主動開戰/造艦 |
| `ProfileScientific`("scientific") | 1 | 3 | 否 | 重研究,選最貴科技做長線投資,受威脅時固守防禦,外交傾向結盟 |
| `ProfileBalanced`("balanced") | 1 | 1 | 否(相等) | 工研均分,選最便宜科技穩定推進,有機會即擴張 |
| `ProfileExpansionist`("expansionist") | 2 | 1 | 是 | 偏工業(造殖民船),有機會即擴張,好戰但比 aggressive 溫和 |

注意:`ProfileExpansionist` 的 `IndustryWeight(2) > ResearchWeight(1)`,因此在「好戰判斷」(2.3、2.4 節)與 aggressive 同屬好戰陣營,但在「傾向擴張」判斷（`isExpansionOriented`,按 `Name` 比對)又與 balanced 同屬擴張陣營——兩套判斷準則（數值權重 vs. 具名 Profile)並存,是刻意的設計分工,不是矛盾。

## 3. 架構圖

```
                     ┌────────────────────────┐
                     │   Decider (interface)   │
                     │ decider.go              │
                     │  ColonyJobs / TaxRate    │
                     │  ResearchTopic           │
                     │  BuildPriority / Stance   │
                     │  Mode()                  │
                     └───────────┬──────────────┘
                                 │ 實作
              ┌──────────────────┴───────────────────┐
              │                                       │
   ┌──────────▼───────────┐                ┌──────────▼────────────┐
   │   RemakeDecider        │                │  OriginalDecider(未實作) │
   │   decider.go            │                │  待逆向原版 AI 後新增      │
   │   Profile + 國庫門檻      │                │  docs/tech/            │
   │                         │                │  (original AI RE 研究,  │
   │  委派至同套件既有函式:     │                │   目前尚無此文件)         │
   │  - economy.go            │                └────────────────────────┘
   │    DecideColonyJobs
   │    DecideTaxRate
   │  - research.go
   │    DecideResearchTopic
   │  - military.go
   │    DecideBuildPriority
   │  - diplomacy.go
   │    DecideStance
   └──────────┬──────────────┘
              │ NewDecider(mode, profile) 工廠
              │  mode=ModeRemake   → RemakeDecider, ok=true
              │  mode=ModeOriginal → RemakeDecider(fallback), ok=false
              ▼
   ┌────────────────────────────┐        ┌───────────────────────────┐
   │  internal/engine/ai.go        │        │  internal/diplomacy/         │
   │  ApplyAIEconomy               │        │  (RelationLevel/Score/Event) │
   │  → decider.ColonyJobs         │        │  供 diplomacy.go 的 Stance    │
   │    / decider.TaxRate          │◄───────│  決策讀取關係等級              │
   │  （經 Decider 介面注入）        │        └───────────────────────────┘
   │  ⚠ ResearchTopic/BuildPriority │
   │    /Stance 尚無回合引擎呼叫端    │
   └────────────────────────────┘
```

## 4. 外交關係層(`internal/diplomacy`)

`diplomacy.Stance` 的輸入 `RelationLevel` 來自獨立的關係系統,同屬設計性重建(`relations.go` 檔頭聲明),與 AI 性格決策共同構成外交姿態邏輯:

- **17 級關係量表**(`FEUD` 到 `HARMONY`,`relations.go:18`):量表**名稱**對齊原版資料——來自遊戲資料 BILLTEXT,已譯於 `assets/i18n/misc.tsv`,顯示時可經 i18n 轉中文。這部分是**原版權威**。
- **數值分數 `RelationScore`**(範圍 `[-100, +100]`)、**分數→等級對映**(`RelationLevelForScore`,線性對稱映射,`NEUTRAL` 含 0 分)、**14 種事件的調整值**(`events.go` 的 `relationEventDelta`,如宣戰 `-40`、結盟 `+30`、貿易往來每回合 `+1`)、**每回合自然漂移速率**(`naturalDriftPerTurn = 1`,無事件時分數往中立回歸)——這些數字**全部是本專案設計值**,原版實際數字未知。
- `IsHostile()`(`< RelationWary`)、`IsFriendly()`(`> RelationAffable`)是 `diplomacy.go` 外交姿態判斷所依賴的門檻,同屬設計值。

## 5. 與原版的差異聲明

**`internal/ai` 目前提供的 remake AI 不是、也不試圖精確重現原版 MOO2 的 AI 行為。**

- MOO2 的 AI 決策邏輯與難度加成,官方手冊未公開規則,英語社群多年逆向也公認未破解(詳見 `docs/tech/community-mechanics-findings.md`)。因此本層**沒有任何權威來源可移植**,不同於 `internal/gamedata` 的公式層(可逐條對照手冊與 `openorion2` 原始碼驗證,見 `docs/tech/moo2-formulas-reference.md`)。
- `docs/tech/design-reconstruction.md` 是這一層的界線總覽文件:明確列出「原版權威資料」與「本專案設計」的分野,並要求每個設計性重建的檔頭與函式註解都標註【設計性重建,非原版】——本文件第 2、4 節的表格即依此原始碼註解整理。
- **original 模式待 RE**:`ModeOriginal` 目前只是型別上的保留位,`NewDecider` 對它 fallback 回 remake 邏輯並回傳 `ok=false`。要讓 original 模式名副其實,需要先完成原版 AI 的逆向工程研究(目前 `docs/tech/` 尚無對應的 `original-ai-re.md`,屬於待建立的後續工作),再新增 `OriginalDecider` 實作並接上 `NewDecider` 的 `ModeOriginal` 分支。
- 設計值可調:未來若實機逆向出原版真值(如 AI 性格分布、關係事件的真實調整量),可直接替換 `internal/ai` / `internal/diplomacy` 內的設計值,不影響已驗證的公式層與其餘架構。
