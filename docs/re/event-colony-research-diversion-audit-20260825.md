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

## 2026-08-30 IDA 勘誤與閉合

- `sub_23DFE @ 0x23DFE` 並未讀取「事件 raw type」。它固定讀兩筆九位元組 record：
  `byte_19AC35` 是事件 16 record 的 `status +1`，`word_19AC37` 是其目標殖民地；
  `byte_19AC7D` 是事件 24 record 的 `status +1`，`word_19AC7F` 是其目標星系。
- `2／6` 是兩筆持續事件的活動／展示狀態值，不是事件 ID 或種類列舉。第一條比較輸入 colony
  index；第二條先由 colony `+2` 取 planet，再由 17-byte planet record `+2` 取 star，才與事件
  24 目標星系比較。因此事件 16 是瘟疫，事件 24 是超新星；事件 25 時空異象不在本 helper。
- `sub_E2710`、`Apply_Colony_Pop_Growth_ @ 0xE2DCA` 與 AI 殖民地 callers 共用此 filter。
  超新星的 RP 轉用與瘟疫／超新星的殖民地回合 gate 均為**已證實**，不再保留 raw 2／6
  名稱未知的舊斷言。
- 可重生證據：`tools/ida/audit_event_diversion_types.py`；匯出：
  `evidence/event-diversion-types-ida-20260830.json`。輸入 EXE SHA-256 與位址基準同上，工具為
  IDA Pro 9.4，正式 `.i64` 以一次性副本唯讀分析。

## 勘誤

- `tools/ida/late_oracle.idc` 與 `internal/shell/events.go` 曾把 `sub_23DFE` 稱為
  `Pick_Random_Colony_No_Capitol`。原始指令直接否定此名稱；後續文件與匯出只保留 raw 位址及
  「事件殖民地 filter」語意。事件摧毀建築排除 Capitol 仍可依 Capitol 不屬一般建築槽的資料模型
  保留，但不得再引用 `sub_23DFE` 作證。

## Remake 對映

- Colony output 仍計算並保留 Research，供超新星搶救進度讀取。
- 帝國研究聚合需略過被超新星轉用的 Colony Research；艦隊研究與條約研究不受影響。
- 時空異象既有零人口副本已同時停止產出、成長、食物與維護，不由本切片重複處理。

## 停止線

- 1.50 的事件改動與 GNN 逐幀呈現不納入標準 1.31 玩家玩法 gate；若日後加入 1.50 profile，
  以獨立二進位與雜湊重開，不從 1.31 靜態證據外推。
