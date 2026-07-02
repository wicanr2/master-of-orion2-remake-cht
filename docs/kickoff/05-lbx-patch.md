# LBX 資產與 patch 1.3 / 1.5 處理(kick-off)

> 回答 CLAUDE.md 兩件事:①「patch 如何處理」②「主選單可選版本 1.3 / 1.5」的架構。
> LBX 解析細節見 `01-openorion2-assessment.md` §1。本文聚焦資料來源與版本策略。

## 1. 資料來源分層(玩家自備正版)

```
玩家的正版 MOO2 安裝
  ├─ 基礎 *.lbx(原始 1996 資產:圖、palette、字串、音樂)
  ├─ [1.31] 覆蓋用 *.lbx(見 §2)
  └─ [1.5]  patched EXE + PARAMETERS.CFG + config schema(見 §3)
        ↓ 玩家指到這個目錄
我們的 go/ebiten 程式  ←── 執行期讀取,不散布任何版權檔
```

- repo **不含**任何 .lbx / EXE / 手冊(見 `.gitignore`)。玩家指定遊戲資料夾。
- 我們複用 openorion2 的 LBX 解碼演算法(純資料轉換),用 Go 重寫。

## 2. Patch 1.31 的本質:替換 .lbx 資料檔

`moo2_patch1.31/MOO2V131.ZIP` 內容是一組**替換用 .lbx** + `ORION2.EXE`:
`GAME.LBX`、`TECHNAME.LBX`、`TECHDESC.LBX`、`RACES.LBX`、`MSGENG.LBX`、`HELP.LBX`、`COLREFIT.LBX`… 等。
這是經典 1996/97 DOS 官方 patch —— **直接覆蓋基礎遊戲的同名 .lbx**,修正資料與平衡。

→ 對我們:1.3 版 = 「基礎 .lbx 疊上 1.31 的覆蓋 .lbx」。資料讀取層要支援**檔案覆蓋順序**(後載覆蓋先載),等同原版 patch 的行為。

## 3. Patch 1.5 的本質:活躍維護中的社群 patch(與 1.3 完全不同)

`moo2_patch1.5/MOO2-1.50.26.zip` 是**現代社群 patch,仍在更新**(CHANGELOG 最新 `1.50.26, May 30 2026`)。結構:

| 檔案 | 作用 |
|---|---|
| `ORION150.EXE` | patched DOS 執行檔 —— **1.5 的遊戲規則改動主要烘在這**(DOSBox 執行) |
| `PARAMETERS.CFG` | mod 設定檔(mod_name/mod_id、LBX 與介面設定、大量可調參數) |
| `docs/config.json`(1.3MB) | **參數 schema**:定義各類(如 `AI`)可調參數的欄位、型別、min/max —— 不是遊戲內容,是「有哪些旋鈕」的定義 |
| `150/scripts/main/MAIN0-9.LUA` | **玩家 context 工具腳本**(在主畫面按 0-9 叫出,顯示存檔資訊/統計)—— ⚠️ **不是核心規則引擎**,是選用的資訊 overlay |
| `docs/MANUAL_150.PDF` / `.html` | 1.5 額外手冊(規則差異的權威來源) |
| `mods/maps/*.CFG` | 地圖預設 |

**重要澄清(避免臆測)**:1.5 的**規則差異烘在 patched EXE + PARAMETERS.CFG**,不是在 Lua。Lua(MAIN*.LUA)是選用的玩家資訊工具(檔頭自述 "player context script... press 0 in main screen")。所以「支援 1.5」≠「跑 Lua」,而是「複製 1.5 相對 1.3 的規則差異」。

## 4. 「主選單選 1.3 / 1.5」的架構結論

因為 gameplay 本來就要從零重建(見 `01` §5),我們**不是** patch 原版 EXE,而是自己實作規則。所以版本支援的正解是:

1. **規則引擎參數化**:把可能因版本而異的數值/開關做成一組「規則參數集」(rule profile)。
2. **兩個 profile**:`v1.3`(以原版手冊 + 1.31 .lbx/公式為準)、`v1.5`(以 1.5 手冊 + CHANGELOG + PARAMETERS.CFG 的差異為準)。
3. **資產也隨版本**:1.3 載基礎+1.31 .lbx;1.5 載 1.5 對應的 .lbx/文字。
4. 主選單切換 profile + 資產集。

→ 這需要一份**「1.3 → 1.5 規則差異清單」研究**(讀兩本手冊 + CHANGELOG_150 + PARAMETERS.CFG 逐條比對),排進 WORKLIST。完整性優先:1.5 的每條 changelog 差異都該收,不預先砍。

## 5. 待辦

- [ ] Go 重寫 LBX 解碼器(容器 + RLE + palette),對 1.31 覆蓋 .lbx 驗證可讀。
- [ ] 建立「檔案覆蓋順序」載入機制(基礎 → 1.31)。
- [ ] 研究並列出「1.3 → 1.5 規則差異清單」(手冊 + CHANGELOG + PARAMETERS.CFG)。
- [ ] 設計 rule profile 資料結構,讓 1.3/1.5 都是同一引擎不同參數。
