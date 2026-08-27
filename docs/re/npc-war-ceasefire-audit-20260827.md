# NPC 宣戰、戰爭計時與停戰靜態稽核（2026-08-27）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；正式檔唯讀掛載，
  IDA Pro 9.4／IDAPython 只分析 `/tmp` 副本。
- 位址空間：IDA linear EA，DOS/4GW LE object #1。
- 完整原始指令、函式邊界、原名、bytes SHA-256、相對欄位讀寫端及 Hex-Rays 導覽輸出：
  [`evidence/npc-treaty-negotiations-ida-20260827.json`](evidence/npc-treaty-negotiations-ida-20260827.json)。
  偽碼只供導航，以下結論以原始指令及交叉參照為準。

## 已證實：回合順序與一般宣戰候選

`sub_252A7 @ 0x252A7` 依序呼叫 `sub_252D5`、NPC 條約談判 `sub_2552D`、宣戰候選
`sub_25DF1` 及停戰判定 `sub_2670A`。`sub_25DF1 @ 0x25DF1..0x2670A` 建立每個目標的
候選類型與 reason；一般 reason 23 分支要求接觸、冷卻 `+0x72F < 1`、本回合輪轉目標，並以
`sub_500CF` 國力比例比較下列門檻：

```text
3 * governmentScore + Random(200) + 125 + 25 * difficulty
+ policy 1/2/3 的 50/100/200
+ trade 50 + research 50
+ tribute mode 1/2 的 50/100
```

難度至少 3 時，已與另一個 AI 交戰的目標會被否決。來源帝國已有任何戰爭時不從一般候選宣戰；
否則從所有 type 1 候選均勻選一個，再呼叫 `sub_51078`。政府表沿用
`word_180CCC @ 0x180CCC`；`sub_500CF` 的比例、800 上限與來源每場第三方戰爭折半均已證實。

## 已證實：宣戰 writer

`sub_51078 @ 0x51078..0x5138E`，函式 bytes SHA-256
`60d21cda29b344dc9b38bfb95baba0b86ffe6474a5e6e927c6cd5b7aa70051cf`：

- AI↔AI 雙向政策固定寫 4；政策 5／6 屬有人類帝國參戰的分支。
- current relation `+0x617` 雙向寫成 `-75-Random(25)` 的 `-75..-99`。
- `+0x68F／+0x69F` 等四個談判記憶寫 -200／-200／-200／-130；remake 目前只表示已被
  NPC 談判消費的前兩欄，後兩欄仍保留為未表示證據限制。
- 清除協議、納貢及戰爭計時 `+0x717`、冷卻 `+0x72F`。

## 已證實：計時與停戰

`sub_5090C @ 0x5090C..0x50AD1`，函式 bytes SHA-256
`292ad531e48d457c2ab7431e7f432facc74b51add8081005b55626be064400ab`：

- 每個無序帝國配對的方向冷卻 `+0x72F` 若大於 0，各自減一；歸零且政策為 3 時雙向清為 0。
- 任一方向政策至少 4 時，雙向戰爭計時 `+0x717` 各加一，上限 250。

`sub_2670A @ 0x2670A..0x2689D` 對政策 4／5 且已接觸的配對使用：

```text
warDuration > 90 - 15 * difficulty - 20 * humanThirdPartyWarCount
```

AI↔AI 且沒有需要開啟人類外交畫面的接收者時，直接呼叫
`sub_524FB @ 0x524FB..0x52602`。該 writer 雙向寫政策 3、current relation 加 50 並在 0 封頂，
把 `+0x72F` 設為 30，並清除協議／納貢相關狀態。

## Remake 邊界

- **CONFORMED**：一般 reason 23 候選、AI↔AI policy 4、宣戰關係範圍、協議清除、戰爭計時、
  停戰門檻、30 回合冷卻、JSON 往返及熱座索引壓縮。
- **強推論資料投影**：原版 `+0x5EC` 已證實為逐艦、逐觀察者武器效能矩陣，但 remake producer
  尚未實作，仍以 `FleetStrength` 代入已證實的 `sub_500CF` 公式；詳見
  [`npc-power-matrix-audit-20260828.md`](npc-power-matrix-audit-20260828.md)。
- **未接**：reason 20／22／68／113 的特殊 producer、宣戰 writer 的 `+0x6AF／+0x6BF` 未消費
  記憶、有人類參戰的政策 5／6 分支，以及完整方向接觸／淘汰矩陣。
