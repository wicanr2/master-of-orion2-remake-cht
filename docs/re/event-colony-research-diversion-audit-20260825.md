# 事件殖民地研究轉用稽核（2026-08-25）

## 問題

超新星事件期間，受影響星系產生的研究點是否仍可同時投入帝國一般研究；舊探針將
`sub_23DFE` 命名為 `Pick_Random_Colony_No_Capitol` 是否成立。

## 證據基線

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、既有 `.i64`、`tools/ida/audit_capital_morale.py`
- 位址基準：IDA linear；DOS/4GW LE object #1
- 匯出方式：唯讀保留原始函式名、位址、bytes、運算元與 callers；未修改 IDB。

## 已證實

1. `sub_E2710 @ 0xE2710` 掃描玩家所有非前哨殖民地。每座殖民地先累計人口與產出摘要；在
   `0xE27F4` 呼叫 `sub_23DFE(colonyIndex)`。
2. `sub_23DFE` 回傳 0 時，`0xE2802..0xE280A` 才把 Colony `+0xEB` 與 `+0xED` 加入帝國聚合值。
   回傳非 0 時，殖民地自身產出欄位沒有被清除，但兩項不進帝國總量。
3. `sub_23DFE @ 0x23DFE` 只讀兩組事件全域：事件型別 byte、事件目標 player／planet，以及
   colony→planet→owner 對映；沒有讀 Capitol 科技、建築或首都旗標。
4. `sub_23DFE` 的 callers 包含 `sub_E2710`、`Apply_Colony_Pop_Growth_ @ 0xE2DCA` 與三個 AI
   殖民地計算函式，證明它是事件殖民地 filter，而非「隨機挑選非首都殖民地」。

## 強推論

- 兩種受影響事件中至少包含會凍結／轉用殖民地產出的持續事件。官方手冊對超新星明定「該
  星系產生的全部 Research Points 都投入尋找解法」；這與 `sub_E2710` 保留 Colony RP、但不納入
  帝國 RP 的資料流完全吻合。事件 raw type 2／6 與公開事件 ID 24／25 的精確對照尚未閉合，
  因此只把「超新星 RP 不得同時進一般研究」標為手冊＋consumer 強推論，不宣稱 raw type enum 已命名。

## 勘誤

- `tools/ida/late_oracle.idc` 與 `internal/shell/events.go` 曾把 `sub_23DFE` 稱為
  `Pick_Random_Colony_No_Capitol`。原始指令直接否定此名稱；後續文件與匯出只保留 raw 位址及
  「事件殖民地 filter」語意。事件摧毀建築排除 Capitol 仍可依 Capitol 不屬一般建築槽的資料模型
  保留，但不得再引用 `sub_23DFE` 作證。

## Remake 對映

- Colony output 仍計算並保留 Research，供超新星搶救進度讀取。
- 帝國研究聚合需略過被超新星轉用的 Colony Research；艦隊研究與條約研究不受影響。
- 時空異象既有零人口副本已同時停止產出、成長、食物與維護，不由本切片重複處理。

## 剩餘不確定性

- 事件 raw type 2／6 的完整名稱、兩個事件 record 的建立／清除函式與 AI 受害者 UI 尚待事件系統
  整體 RE；不影響本輪修正玩家超新星 RP 雙重使用。
