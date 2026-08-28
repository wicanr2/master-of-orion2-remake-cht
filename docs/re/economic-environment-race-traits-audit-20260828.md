# 經濟與環境種族特性整併稽核（2026-08-28）

## 範圍與證據契約

本文件整併 `player+0x8A0..+0x8A4`、`+0x8AB..+0x8AC`、`+0x8B6..+0x8B7`
的玩家可見垂直鏈。它不以「remake 有相似欄位」當作原版證據，而是引用已由 IDA 指令、
writer、consumer 與 caller 閉合的窄切片。

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4 與受版控 IDAPython 匯出器；位址均為 IDA linear、DOS/4GW LE
  object #1。外部符號只供導覽。
- C runtime、Watcom helper、平方根內部算法及報表格式化不納入玩法分母；玩家公式只保留其
  已證實的輸入／輸出契約。

## 五項 signed 數值特性

| index／offset | 原版消費尺度 | 已證實玩家效果 |
| --- | --- | --- |
| 1／`+0x8A0` Population | 百分點 | `Colony_Pop_Grows_ @ 0xE1839` 將 signed raw 直接加到成長 factor 100；最終為 `trunc(factor×isqrt(2000×P×(M-total)/M)/100)`，再接短缺、Cloning Center 與 1000 點人口池。 |
| 2／`+0x8A1` Farming | half-food | `sub_DE0C6 @ 0xDE0C6` 對每名一般種族農夫直接加 signed raw；之後才套 Aquatic、重力、士氣、政體、領袖、封鎖與 AI 難度。 |
| 3／`+0x8A2` Industry | 每工人 PP | `sub_DED47` 對每名一般種族工人直接加 signed raw；結果進 `DE280(job=1)`，再進 gross、污染、工業維護與建造消費鏈。 |
| 4／`+0x8A3` Science | 每科學家 RP | `sub_DFE77 @ 0xDFE77` 對每名一般種族科學家直接加 signed raw；結果進 `DE280(job=2)`，再加殖民地固定研究建築。 |
| 5／`+0x8A4` Money | 每人口 0.5 BC | `Colony_BC_Production_ @ 0xE03F1` 使用 `quarterFactor=2×signed(raw)+4`，乘有機人口後以 `(value+2)/4` 四捨五入；一般 raw 0 即每人口 1 BC。 |

這五項不是百分比共用欄位：Population 是百分點、Farming 是半食物、Industry／Science 是
每人口整數基礎值、Money 是每級 0.5 BC。任何共用「經濟百分比」adapter 都會破壞原版尺度。

## Aquatic、Subterranean 與 Tolerant

### Aquatic `+0x8AB`

- `sub_DE0C6` 只在有效氣候 raw 4／5／6／8 重映氣候食物值，差額以 half-food 加入；不是
  無條件農業加成。
- `Player_Effective_Climate_` 與 `Size_And_Climate_Race_Pop_Limit_` 使用同一 trait 形成
  Ocean／Swamp／Terran／Gaia 的有效氣候，再進人口上限與 AI 殖民地容量快取。
- 行星護盾先把 Radiated 改成 Barren，之後才計有效氣候與食物；順序已由 colony writer 鏈閉合。

### Subterranean `+0x8AC`

- 人口上限在氣候／尺寸基值後加入 `2×(planetSize+1)`；Advanced City Planning `+5` 與
  Biospheres `+2` 是其外層另外加項。
- `Compute_Player_Ground_Combat_Bonuses_` 對防守方另加地面戰 bonus `+10`；它不是人口上限
  加成的衍生效果。

### Tolerant `+0x8B6`

- 人口上限的 climate factor 額外 `+25`，再夾到 100；不是把最終人口直接增加 25%。
- 污染依 packed population 中 Tolerant 一般人口與 Android 的比例縮減：
  `round(cleanup×nonTolerantPopulation/totalPopulation)`。混合殖民地不能只看 owner trait。
- `Determine_Event_` 的工業事故及 AI 建築評分另有直接 gate；其事件排程與評分已分別保存在
  專題文件，不應反推成另一個經濟倍率。

## Fantastic Traders `+0x8B7`

- 殖民地生產 Trade Goods 時，一般種族把可用工業以 `(industry+1)/2` 換成 BC；Fantastic
  Traders 為 1:1。
- `Update_Player_Stats_` 聚合非封鎖殖民地食物盈餘時，Fantastic Traders 全額換 BC；一般
  玩家採半額換算。這是帝國彙總 consumer，不可與單一殖民地 Trade Goods 重複計算。
- `Trade_Agreement_Goal_ @ 0x101BA4` 的政府／領袖倍率基值另加 50；它改變協議目標，不是
  每回合額外直接發 50 BC。

## 閉合狀態與剩餘邊界

- **已證實**：五項 signed raw 的精確尺度及主要產出 consumer；Aquatic 食物／氣候／容量；
  Subterranean 容量／地戰；Tolerant 容量／混合人口污染；Fantastic Traders 的殖民地、帝國
  與協議三個互不重複的 consumer。
- **仍須另切片**：新局母星 `Twiddle_Initial_Homeworlds_` 對 Money 的配置影響、各 trait 的
  AI personality 初始權重及 `Calc_Tech_Value_` 估值常數。這些不改變上列玩家公式，但會影響
  AI 行為或開局生成，不能以本整併文件冒稱閉合。
- 直接依據詳見 `population-growth-runtime-audit-20260828.md`、
  `colony-food-production-environment-audit-20260828.md`、
  `colony-industry-production-pollution-audit-20260828.md`、
  `colony-research-production-audit-20260825.md`、`colony-bc-production-tax-audit-20260828.md`、
  `ground-combat-audit-20260828.md` 與 `ai-colony-build-selection-audit-20260826.md`。
