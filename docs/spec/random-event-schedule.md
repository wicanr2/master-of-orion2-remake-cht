# 一般隨機事件排程規格

證據來源：[`docs/re/random-event-schedule-audit-20260825.md`](../re/random-event-schedule-audit-20260825.md)。

## 狀態

- `EventLastTurn`：相對起始星曆的最後成功事件回合；新局與舊 JSON 零值皆為 0。
- `EventAttemptCounter`：最後事件後的保護檢查次數，範圍 0..5；新局為 0。
- 兩欄屬於全局事件排程，必須進 JSON 快照；不得跟著熱座席位交換。

## 排程

1. `DisableEvents` 為 remake 的硬隔離旗標；開啟時不消耗事件亂數、不改計數器。
2. 先跑 Lucky 累積／擲骰。最早日期使用 `elapsed = Turn-1`，必須 `elapsed >= 50`。
3. 沒有 Lucky 強制事件時才跑一般排程：
   - `elapsed < 50` 或 `elapsed < EventLastTurn`：不擲一般排程骰。
   - `EventAttemptCounter < 5`：加一，當次門檻為 0。
   - 否則按五級難度計算 delta 門檻，擲 `Intn(512)+1`，`roll <= threshold` 才進候選。
4. 最多抽五次 `Intn(29)`。Tutor 或 Lucky 強制路徑拒絕壞事件；事件未實作、未達最早日期、
   無適用目標或效果 helper 回報不適用，均消耗一次候選。
5. 一項事件成功結算後，寫 `EventLastTurn=elapsed` 並清零 `EventAttemptCounter`；五次都失敗則保留。

## 驗收

- 純規則測試覆蓋五級整數公式、前五次保護、1／512 擲骰邊界及非法難度安全回退。
- `Turn=50` 不得觸發；`Turn=51` 才進第一次保護檢查。
- 固定亂數源可證明最多五次候選、Tutor 無壞事件、未實作事件不是從池中預先消失。
- JSON 往返保留最後事件回合、檢查計數器與事件亂數抽取位置。

