# 研究領域動態文案與版面稽核（2026-08-27）

## 問題與證據來源

本輪確認 `research()` 八個領域框內的目前主題、RP 成本與超進階等級文字，並區分原版
`TECHSEL.LBX` 證據與 remake 為避免盲選新增的操作摘要。

- 原版輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具與位址基準：IDA Pro 9.4；DOS/4GW LE image 的 IDA linear address。
- 既有證據：`research-application-selection-audit-20260825.md`、
  `research-choice-ui-text-audit-20260826.md` 與 `screen-spec-info-research.md`。

## 已證實

- `sub_10DC12 @ 0x10DC12` 是玩家研究選擇函式；`0x10E389..0x10E3AA` 依序寫回 field 與
  application。原版在研究開始前便選定 application，不是突破後才選。
- 原版使用 `TECHSEL.LBX#0`，調色盤由 `SCIENCE.LBX#0` 提供；八個領域按鈕文字烘在資產中。
- `ResearchSelectWindow` 每個領域顯示下一個 field／application 選項，研究成本格式為 `%u RP`；
  超進階 field 由 `sub_E4410` 的 `field >= 75` 分支增加等級。

## 強推論、remake 適配與未知

- **強推論：**領域框顯示目前主題與成本能避免玩家盲選，且不改變 field/application 寫回結果。
- **remake 適配：**現行畫面每個領域只顯示一條「主題＋成本」，application 多選另由獨立面板處理；
  這與原版整合式清單不同，不可宣稱像素或控制流 parity。
- **未知：**原版動態項目的精確字級、完整文字安全框及 widget ID；現行框由已知 208×98／214×98
  熱區與烘字標題帶推導。

## 找到的實作缺陷與對映

- 一般主題模板、超進階等級、全領域完成及星系轉場仍在 Go 內組雙語句子，須移至 `ui.json`。
- 舊 `extraText` 只有 `maxW`，沒有明確高度；改用每個領域熱區推導的
  `textSafeRect(x+6,y+26,w-12,28)`，文字與熱區共用中心軸，採單行實際字型省略。
- 主題名稱繼續由 `tech.json` 提供，不複製到 UI catalog；Go 只帶 typed topic、等級與 RP。

## 驗證

- catalog／格式參數測試、八框雙軸 containment、來源切片無 `.tr(`、目前主題規則測試、
  全套測試及雙語畫廊共同驗收。畫廊只證明現行 adapter 可讀，不升格原版逐像素證據。
