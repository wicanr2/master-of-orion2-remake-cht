# Hyper-Advanced 重複研究成本規格

## 規則

對八個 `IsHyperAdvancedTopic`：

```text
base = HyperAdvancedCost(ruleProfile)
level = HyperAdvancedLevels[topic] // 已完成級數，尚未完成第一級為 0
cost = base + level * 10000
```

非 Hyper topic 維持 `ResearchChoiceFor(topic).Cost`。

## 消費端

- `engine.RunResearchPhase`：實際完成門檻與溢出扣除。
- `GameSession.ResearchCostForDisplay`：研究選單顯示。
- `aiResearchCandidates`：AI 選題成本，且已完成 terminal Hyper 仍須作為該領域下一個候選。
- JSON／熱座／TCP 快照：沿用 `HyperAdvancedLevels`，不新增重複狀態。

## 遷移

舊存檔 `HyperAdvancedLevels == nil` 時，每個已完成 Hyper topic 補為 level 1。這次遷移完成後，
下一級成本立即成為 `base+10000`。

## 邊界測試

- 1.50 level 0／1／2 成本為 25000／35000／45000。
- 1.31 level 0／1／2 成本為 15000／25000／35000。
- 第二級在 `cost-1` 不完成、在 `cost` 完成並遞增 level。
- 舊存檔的一級狀態不能用 25000 完成 1.50 第二級，必須累積到 35000。
- AI 完成全部一般科技後仍會取得八個 terminal Hyper candidate，成本含各自 level。
