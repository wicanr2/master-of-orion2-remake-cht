# AI 常態研究選擇規格

## 輸入

- AI 持久化的原版 raw6／raw4／raw7 與 31 格 runtime 種族特性。
- 難度 `0..4`、相對開局回合、目前可研究的八領域隊首、已知 application、其他帝國已知
  application，以及本回合實際研究產出。
- 可隨存檔往返的研究亂數流。

## 契約

1. 新建 AI 時依既有 `sub_589D6` 規則產生一次 raw profile，之後持久保存；不可每次選題重抽。
2. 目前 field 尚未完成且沒有研究狀態改變時，不重選、不清除進度。
3. 需要新研究時，將八領域各自第一個未完成 field 組成可用集合，交給
   `StartingOriginalApplicationPick` 進行一次 application 級抽選。
4. 抽中的 application 同時決定 field；非全授予 field 立即保存
   `ResearchApplication`，不得再由 category 最大值做第二次選擇。
5. 開局 150 回合內的 category `0x12` 倍增只在 `relativeTurn < 150` 生效；不得永久烘入共用估值。
6. profile 未知的舊存檔保留現行 `DecideResearchTopic` fallback，並維持明示的「設計性重建」標記。
7. Hyper-Advanced field 仍可重複成為候選，成本使用目前 level；完成後可再次選中。

## 驗收

- 固定 raw profile、候選、研究產出與研究亂數位置時，選中的 field/application 可重播。
- 測試應證明選擇是單次 application 抽樣：不得先選 field 後改挑另一個 application。
- 未完成 field 不改題；完成、偷得目前 field 或狀態變更時才重選。
- JSON 往返後 raw profile 與下一次研究選擇一致。
- profile 未知時不 panic，且走既有 fallback。

證據見 [`../re/ai-normal-research-selection-audit-20260826.md`](../re/ai-normal-research-selection-audit-20260826.md)。
