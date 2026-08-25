# 隨機事件 11／12 礦產效果稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4，既有 `.i64`，`tools/ida/audit_event_minerals.py`
- 位址：IDA linear，DOS/4GW LE object #1
- 範圍：靜態反組譯；未做原版動態 oracle

## 已證實

1. `sub_2325E @ 0x2325E` 是事件 11 的目標抽選器。它最多重試 200 次，每次呼叫
   `sub_23DA0 @ 0x23DA0`，只接受 `planet+0x0A > 3`，亦即 raw 礦產 4（Ultra Rich）。
2. `sub_23DA0 @ 0x23DA0` 以 reservoir sampling 均勻抽取目標玩家殖民地，並要求
   `colony+6 == 0` 及 `colony+0x13F == 0`。`+0x13F` 的玩家語意目前未知；不可把它命名成
   「非首都」或其他未證實名稱。
3. `sub_232BB @ 0x232BB` 是事件 12 的目標抽選器。它最多重試 200 次，每次呼叫
   `sub_23D44 @ 0x23D44`，只接受 `planet+0x0A < 4`。
4. `sub_206A2` 的事件 11 消費端 `0x20C09..0x20C77` 寫入
   `max(0, oldMineral-1)`；事件 12 消費端 `0x20C87..0x20CCB` 寫入
   `min(4, oldMineral+2)`。兩者有殖民地時都呼叫 `sub_E2A70` 重算殖民地。
5. 這段消費端只寫 `planet+0x0A`，沒有寫 `planet+9`；因此礦產事件不應按星球生成表重算重力。

## Remake 邊界

- 已依上述證據實作礦產條件、200 次拒絕取樣、精確增減量、上限、殖民地產能回寫，以及
  玩家／熱座／AI 路徑。
- remake 的殖民地模型沒有 `colony+0x13F` 的已證實對等狀態；事件 11 暫時不套這個 raw filter。
  這是明示近似，不影響已證實的 Ultra Rich 限制與 4→3 效果。
- 抽選器重試耗盡後共用尾端的完整回傳資料流未做動態驗證；remake 採「事件不適用」而非自訂
  低索引 fallback。
