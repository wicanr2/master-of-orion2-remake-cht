# 人口成長點數尺度稽核（2026-08-24）

## 結論

- **已證實（官方 patch 1.50 規則）**：人口成長率與 `PopGrowth` 累加值以「千人」為點數單位；
  1 個遊戲人口單位是 1,000 點。複製中心的 `+100k` 因此是每回合固定 `+100` 點。
- **已推翻**：`internal/shell.popGrowthThreshold = 300` 是 remake 調校值，會讓同一份官方成長公式
  約快 3.33 倍兌換人口；它不是原版尺度。
- **已證實（remake 資料流）**：`gamedata.ColonyGrowth` 產生的點數進入
  `LastPlayerOutput.Colonies[i].PopGrowth`，種族百分比只在 `advancePopulation` 套一次，再存入
  `popAccum`。既有 JSON 存檔的 `popAccum` 已是同一種成長點數，不需比例遷移。
- **未知／本輪不冒稱完成**：原版 1.31 執行檔是否逐指令使用相同常數、多種族殖民地逐族分配、
  難度 AI 加成 `i`、事件與職務重排的完整分支仍需 `Apply_Colony_Pop_Growth_` 資料流證據。

## 輸入與定位

| 輸入 | 雜湊／位址空間 | 用途 |
|---|---|---|
| `moo2_patch1.5/MANUAL_150.html` | SHA-256 `ac81d580921d0078496127cb7fb7e064e74bf8cb26f2db958b26b5dff09acba9` | 官方 patch 1.50 人口公式與設定契約 |
| `ORION2.EXE` 既有 IDA 資料庫索引 | IDA linear `Apply_Colony_Pop_Growth_ @ 0xE2DCA` | 原版回合消費端定位；本輪未新增反組譯斷言 |
| `internal/save/entities.go` | 原版 `.GAM` `Colony.PopGrowth [maxRaces]int16` | 確認原版保存逐族累加值，而非 remake 顯示人口 |

本輪 IDA Pro 容器授權仍無法啟動，因此沒有把新手冊結論假裝成 IDA 證據；函式位址只沿用既有
符號索引，證據等級不升級。

## 尺度推導

官方 1.50 手冊的「Notes on Population Growth」同時給出：

1. 基礎值 `a = trunc(sqrt(2000 * POPRACE * (POPMAX - POPAGG) / POPMAX))`；
2. 種族比例使用 `c = 100 * (POPRACE / POPAGG)`；
3. 複製中心標準加成是 `+100k`，而設定預設值是 `100`；
4. 複製中心加成在乘法成長值截斷後才加入。

因此設定／存檔成長點的 `100` 就是 100k 人口，完整的一個遊戲人口單位（1,000k）必須累積
1,000 點。這也解釋典型基礎成長值約 90：它代表每回合約 90k，而不是每三回合增加一整格人口。

## Remake 垂直鏈與停止線

`ColonyBaseGrowth` → `ColonyGrowth` → `engine.ColonyOutput.PopGrowth` →
`shell.advancePopulation` → `GameSession.popAccum` → `Population`／職務指派 → JSON `PopAccum`。

本輪只修正已由官方契約閉合的尺度。逐族成長與 AI 難度項會改變資料模型，必須先追回原版欄位
與回寫順序；在那之前保留為待辦，不用新的平均公式補洞。
