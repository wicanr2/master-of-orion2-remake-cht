# 隨機事件 25「時空異象」靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`（1.31），SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`；分析既有 `.i64` 的 `/tmp` 副本。
- 位址基準：IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_random_event_stasis.py`；本輪 JSON：
  `/tmp/moo2-event-monsters-work/random-event-stasis.json`（不加入 Git）。
- 匯出保留原始函式名、位址、bytes、運算元與 callers；語意名稱只供導覽。

## 已證實

1. `Determine_Event_ @ 0x2230A` 先走一般 `sub_22D57` 受害帝國抽選；事件 25 建立分支
   `0x22CE8..0x22D1E` 再呼叫 `sub_23BEC @ 0x23BEC`，於該帝國有效殖民地中以 reservoir
   sampling 選星系。目標不是固定目前玩家。
2. 建立端把星系寫入 record `+0x07`、age `+0x06` 清零，再呼叫 `sub_242FC`；同星已有
   彗星、瘟疫、人口暴增、海盜活動、超新星或時空異象時候選失敗。
3. `sub_206A2 @ 0x2119F..0x2123F` 只在 record state 2 消費。每個 active turn 逐一掃目標星
   五個殖民槽，對存在的 colony 呼叫 `sub_E2A70 @ 0xE2A70` 重算殖民地與帝國聚合。
4. age `>4` 時才呼叫 `Random(20)`；結果等於 1 就把 state 設為 5。無論亂數結果，age
   `>20` 時也把 state 設為 5，最後才將 age 加一。因此 age 0..4 不擲骰、age 5 首次擲骰、
   age 21 強制結束。
5. `sub_23DFE @ 0x23DFE` 是事件殖民地 filter。其 callers 包含玩家聚合
   `sub_E2710 @ 0xE2710`、人口成長 `Apply_Colony_Pop_Growth_ @ 0xE2DCA` 及三個 AI
   殖民地計算函式，證明凍結不是玩家專用 UI 效果。
6. `sub_E2710 @ 0xE27EE..0xE2802` 在 filter 命中時不把 colony `+0xEB` RP 加入一般帝國
   研究；`Apply_Colony_Pop_Growth_ @ 0xE3355..0xE335F` 亦將事件殖民地送入專用狀態。

## 手冊已證實、靜態資料流尚未逐分項閉合

- 官方手冊 p.181 明定該星所有殖民地停止生產、成長與人口移動，也不消耗食物或支付維護。
- 本輪靜態資料流直接錨定研究聚合、人口成長與玩家／AI caller；食物、工業、建造、維護與
  人口移動的每個 raw 欄位尚未逐一列出，故標為「手冊已證實＋共用重算鏈強推論」，不冒稱
  每一分項都已由單獨指令閉合。

## 推翻的舊斷言

- 「用 `Random(100)<=5` 等價即可」：機率相同但原版消耗 `Random(20)`，會改變後續 PRNG。
- 「沒有強制停止，只要 500 回合內大概會結束」：錯；age 21 強制結束。
- 「只凍結目前玩家殖民地」：錯；目標可為 AI／熱座，filter 也有 AI 計算 caller。

## 未知

- `colony+0x13F` 的玩家可見語意。
- 事件 record state 2／5／6 與 GNN 畫面切換的精確同回合時序。
- 1.50 binary 未取得，無法排除版本差異。

