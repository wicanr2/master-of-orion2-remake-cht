# 條約摘要 typed 資料與外部文案規格

## 資料契約

- `TreatySummaryParts(TreatyState)` 依固定順序回傳：正式條約、我方進貢、對方進貢、普通貿易、
  研究協議、特殊貿易。
- 每筆只保存 `TreatySummaryKind`、`Turns`、`Value`、`Percent`；不得保存顯示名稱或句子。
- 沒有任何協議時只回傳 `TreatySummaryFormalNone`，讓 UI 明確顯示無正式條約。
- 特殊貿易種類分成 food／research／unknown；未知 raw kind 不可猜成其中一種。

## UI 契約

- 每種摘要使用 `diplomacy.summary.*` JSON 鍵；項目間分隔符使用 `list.separator`。
- BC／RP 單位留在模板，整數值由 typed 欄位提供。
- 未知摘要種類使用 `diplomacy.summary.unknown`，不可把 enum 數值或鍵畫到畫面。
- 完成組合後仍由既有 560×24 安全框單行截斷，不能因項目增加而侵入按鈕列。

## 證據分級

- 條約欄位與普通協議逐回合值：已證實，見
  `docs/re/trade-research-agreement-turn-audit-20260825.md`。
- 單行摘要組合、措辭與特殊貿易名稱：remake 顯示轉接，不宣稱原版逐字一致。

## 驗證

- shell 測試釘住順序、typed 參數與空狀態。
- UI 測試釘住雙語鍵、格式化、分隔符與未知 fallback。
- 靜態測試禁止 `internal/shell/treaty.go` 保存既有中英文摘要句子。

