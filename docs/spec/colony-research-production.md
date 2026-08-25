# 殖民地研究建築產出規格

證據來源：[`docs/re/colony-research-production-audit-20260825.md`](../re/colony-research-production-audit-20260825.md)。

## 規則

1. 先以科學家人口、種族基礎研究、士氣及其他人口層修正計算人口研究量。
2. Research Laboratory、Planetary Supercomputer、Galactic Cybernet、Autolab 分別在殖民地總量固定加上 5、10、15、30 RP。
3. 這四棟建築不修改 `ResearchPerScientist`，固定加成也不乘以科學家人數。
4. 建築被拆除或破產處分時，只撤回相應 `FlatResearch`；不得扣減種族、星球特殊物或 Astro University 所提供的 per-scientist 值。
5. 四棟建築可並存，固定值相加。

## 驗收條件

- 建成每棟建築後，`ResearchPerScientist` 保持不變。
- `FlatResearch` 分別增加 5／10／15／30。
- 拆除後只撤回該固定值並回到原值。
- 科學家人數不同時，同一棟建築造成的總研究差仍為相同固定值。
