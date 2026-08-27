# 原版 AI 稅率規格

規格狀態：`CONFORMED`

證據：[`../re/ai-tax-rate-audit-20260828.md`](../re/ai-tax-rate-audit-20260828.md)

## 行為契約

1. 正常建立的 AI 帝國以 0% 稅率起步。
2. AI 每回合不得依國庫套用 remake 的 10／30／50% 稅率門檻。
3. 從 `.GAM` 或 JSON 載入的 AI 稅率必須保持原值；正常 AI 回合不覆蓋它。
4. 原版職務資料不足而回退 `ApplyAIEconomy` 時，只允許代理殖民地職務；呼叫端必須還原回退前
   稅率，避免 fallback 偷渡設計型調稅。
5. `DecideTaxRate` 可繼續供明示的 remake AI／模擬工具使用，但不得標成原版玩法。

## 驗收

- 新建 AI 的 `TaxRate == 0`。
- 將 AI 稅率設為非零後跑一個正常世界回合，值保持不變。
- 純 Go 全套回歸通過。
