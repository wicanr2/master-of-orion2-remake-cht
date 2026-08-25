# GNN 狀態播報 29–35 靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`（1.31），SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`；分析既有 `.i64` 的 `/tmp` 副本。
- 位址基準：IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_event_status_broadcasts.py`；本輪 JSON：
  `/tmp/moo2-event-monsters-work/event-status-broadcasts.json`（不加入 Git）。
- 29..35 record 固定為 `0x19ACA9..0x19ACE7`，每筆九位元組；匯出保留所有直接 refs。

## 已證實觸發與欄位

| ID | 建立端 | 已證實觸發 | record／新聞欄位 |
|---|---|---|---|
| 29 | `sub_233AB @ 0x233AB` | `sub_E4EB3 @ 0xE4EB3` 掃描原先 active、已無有效殖民地的帝國；先呼叫 `sub_E45FF` 清理，再建立播報 | `+0x00=帝國槽`、state 1；四種隨機文案 |
| 30 | `sub_23563 @ 0x23563` | 計算每個帝國至少有一座有效殖民地的星系數，找最大值；最多播三階段 | `+0x00=帝國槽`、`+0x03=最大殖民星數`、`+0x05=階段 1..3` |
| 31 | `Determine_Event_ @ 0x2288B..0x228DC` | 呼叫事件 30 檢查後，`Random(40)==1`、elapsed `>50`、議會旗標為 0 | state 6、`+0x05=Random(4)-1` 四種排行榜類別 |
| 32 | `sub_233B8 @ 0x233B8` | `sub_FFDDA @ 0xFFE21..0xFFE31`：一般帝國艦艇抵達 star raw type `0x0B` 且 `star+0x33==0` 時 | `+0x00=發現者帝國槽`、state 1；四種隨機文案 |
| 33 | `sub_233C5 @ 0x233C5` | 本 1.31 `.i64` 沒有直接 caller | `+0x00=帝國槽`、state 1；與 29 共用四種隨機文案分派 |
| 34 | `sub_233D2 @ 0x233D2` | `sub_2670A @ 0x2670A` 的 AI 外交投降鏈：`sub_E4D06` 寫 pending surrender 後立即建立播報；資產由後續 `sub_E4DC9/sub_E4B5F` 轉移 | `+0x00=投降帝國槽`、`+0x03=接收帝國槽`、state 1 |
| 35 | `sub_233E6 @ 0x233E6` | 本 1.31 `.i64` 沒有直接 caller | `+0x00=帝國槽`、`+0x03=殖民地索引`、state 1；四種隨機文案 |

## 事件 30 門檻

- `word_19ACB7` 是已播階段，值大於 2 後永久停止。
- `S` 為銀河星數、`H=trunc(S/2)`、`M` 為最大殖民星數；第 `stage=0..2` 次門檻為
  `trunc((stage+2)*H/4)`，達標即播並把 stage 加一。
- stage 2 未達一般門檻時另有議會前窄分支：`H-2 <= M < H` 且議會旗標為 0 也可播第三階段。
- 計數單位是「有至少一座有效殖民地的星系」，不是殖民地筆數或總人口。

## 證據限制與 remake 對映

- 事件 33／35 setter 有明確 record 契約，但 1.31 無直接 caller；可能是未使用入口、間接 callback
  或版本功能。remake 若在已存在的「擊敗安塔蘭」「殖民地叛亂」玩家路徑建立同形新聞，證據等級
  必須標為 contract-confirmed / trigger-approximation，不得稱原版 caller 已閉合。
- 事件 32 的 star raw type 0x0B 與 Orion 語意由事件表及守護者上下文強推論；原始欄位仍保留。
- 事件 34 的完整勘誤、延後 consumer 與資產清單見
  [`empire-surrender-audit-20260825.md`](empire-surrender-audit-20260825.md)。
- 1.50 binary 未取得，無法核對 setter caller 與門檻版本差異。

## 推翻的過期斷言

- 稽核前 `RandomEvents` 把 29、30、33 標 `Implemented=true`，但當時 remake 沒有任何
  `EventReport` 觸發端；該狀態已由本輪垂直鏈實作取代。事件 34 則在後續深查
  `sub_E4D06/sub_E4DC9/sub_E4B5F` 後才改為已實作。
- 31 不是固定每 N 回合排行榜，而是 `Determine_Event_` 路徑中的 1/40 檢查並受 elapsed／議會 gate。
- 30 的「帝國壯大」不是泛稱國力，原版此切片明確比較殖民星系數並保存三階段。
