# 貿易／研究協議逐回合規格

## 目的

以 `docs/re/trade-research-agreement-turn-audit-20260825.md` 的靜態原版證據，取代現行
「剩餘差距除以五」近似，同時保持玩家、AI、回合經濟與存讀檔的垂直鏈可重播。

## 狀態

每份玩家↔AI `TreatyState` 保留：

- 貿易／研究 active flag；
- 玩家方向與 AI 方向的 current value；
- remake UI 使用的回合計數；
- 特殊貿易狀態（不受本規格改寫）。

goal 是每回合衍生值，不新增存檔欄位。舊存檔若 active、turn=0 且雙方 current 都是零，
仍以 `-minPopulation/2` 補成原版負起點。

## 純規則

提供一個帶 roller 的純函式：

```text
advanceAgreementValue(current, goal, roll5) -> next
```

規則：

1. `current >= goal` 時直接回傳 goal，不抽亂數。
2. 否則計算 Go 整數除法的 `q=goal/5`、`r=goal%5`；目前 goal 為非負。
3. `r!=0` 才呼叫一次 `roll5()`，它必須回傳 1..5；`roll<=r` 時 bonus=1。
4. 回傳 `min(goal, current+q+bonus)`。
5. 不加入「最少移動一點」；所有補點只能來自原版餘數亂數。

## 世界回合順序

`GameSession.advanceTreaties` 對玩家↔AI 的可表示子集採：

1. 從最後一個 AI 到第一個 AI，依序處理 AI trade、AI research。
2. 再從最後一個 AI 到第一個 AI，依序處理 player trade、player research。
3. 每份 active 協議的 TradeTurns／ResearchTurns 只加一次。
4. 最後推進 remake-only special trade，一份每回合一次。
5. 產出的雙方 BC／RP 仍只各餵一次 `RunEmpireTurn`。

## 亂數與存檔

- 新增 agreement 專用 `randStream`，惰性種子為 `EventSeed*2654435761+29`。
- snapshot 保存 `AgreementDraws`，restore 依相同種子快轉。
- 舊存檔缺欄位時 draw=0，保持確定性安全降級。
- 這只承諾 remake 存讀檔與多人重播一致；不宣稱原版 PRNG bytes 相同。

## 驗證

- goal=10、current=-10：五回合到 0、十回合到 10。
- goal=12：每回合固定 +2，餘數 roll 1/2 成功、3/4/5 失敗；goal 可整除時零抽取。
- current 高於下降後 goal：單回合立即降到 goal。
- 多 AI 時確認 AI 方向與玩家方向皆為逆索引，且 trade 先於 research 的抽取順序。
- snapshot／restore 後 agreement RNG 抽取位置與後續狀態指紋一致。
- 正常 `EndTurn` 的 `TreatyIncomeBC/TreatyResearch` 與狀態 current 相同。

