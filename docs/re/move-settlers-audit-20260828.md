# 殖民者遷移、抵達與人口回寫稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／Hex-Rays 9.4.0.260610、
  `tools/ida/audit_move_settlers.py`；位址均為 IDA linear EA。
- 導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱不單獨作證據。
- 可重生證據：
  [`evidence/move-settlers-ida-20260828.json`](evidence/move-settlers-ida-20260828.json)。
- `memmove_` 只負責壓縮四位元組記錄；它是 C runtime 邊界，不納入玩法分母。本文件追查的是
  呼叫前後可見的人口、貨運艦、ETA、封鎖與訊息狀態。

## 遷移 ETA 與容量

`Settler_ETA_ @ 0xFED3F..0xFED77` 呼叫 `Parsecs_Between_Stars_ @ 0xEBEB7`，讀玩家
`+0x5A0` 航速並以 2 為下限：

```text
eta = min(15, ceil(parsecs / max(player+0x5A0, 2)))
```

同星距離為零，所以 ETA 也是零。

`Room_For_Another_ @ 0xFED77..0xFEE4C` 以 `sub_E0C1D(destination, race)` 算抵達人口的
種族特定容量。它計數目的殖民地中「自身種族容量小於等於抵達者容量」的既有人口；第三參數為真
時，也計入同目的地、符合相同容量比較的在途人口。UI 預檢傳真，實際抵達回寫傳假，避免把正在
結算的記錄重複計數。函式以 `count < incomingCapacity` 判定是否有空位。

## 玩家送出前檢查：`Pop_Tries_To_Settle_ @ 0xFEE4C..0xFF015`

唯一 caller 位於遷移 UI `0xB9F3A`。輸出旗標的 raw 語意如下；未能由獨立 consumer 命名者保留
raw 條件，不把推論升格為畫面文案：

| raw 旗標 | 已證實條件 |
|---:|---|
| 0 | 成功時設為 1；flag 13 同時保存 ETA |
| 1 | 來源與目的殖民地 owner 不同 |
| 2 | 來源星系的 `star+0x2A` 含 owner 封鎖 bit |
| 3 | 來源命中 `sub_2341E` 的 anomaly gate |
| 4 | 目的殖民地 `+0x06 != 0` |
| 5 | 來源人口等於 1 |
| 6 | 跨星且 `player+0x36 < 5 × (player+0x40 + 1)` |
| 7 | 跨星在途人口已達 25 |
| 8 | 來源命中 `sub_234B8`；高層事件語意尚未證實 |
| 9 | 目的地含在途人口後已無容量 |
| 11 | 目的星系遭該 owner 封鎖 |
| 12 | 目的星系有 space anomaly |
| 14 | 選中的來源人口 raw prisoner bit 10 已設 |

`player+0x36` 是存檔欄位 `totalFreighters`，`player+0x40` 是在途殖民者數。由此可直接證實每名
跨星殖民者需要五艘貨運艦的 gate；同星遷移不走此 gate。

## 四位元組在途記錄：`Settle_Pop_ @ 0xFF116..0xFF212`

唯一 caller 位於 UI `0xBA16A`。同星遷移立即呼叫人口加入函式；跨星則把記錄寫入
`player+0x42 + 4 × player+0x40`，再增加 `player+0x40`：

| byte | raw 內容 |
|---:|---|
| 0 | 來源殖民地索引 |
| 1 | 目的行星索引 |
| 2 low nibble | 來源人口的 race／player slot |
| 2 high nibble | ETA |
| 3 bits 0..1 | 來源人口職務 |

來源人口會先保留上述 race 與 job，再將 `colony+0x0A` 減一，以最後一筆 packed colonist 填補
被移走的洞。這證實既有 `SettlerInfo.Player` 欄位名稱過度狹窄；其 bytes 是人口 race／player
slot，而不是必然等於目前帝國 owner。

## 回合移動與抵達：`Move_Settlers_ @ 0xFF212..0xFF477`

`Next_Turn_Calc_ @ 0x136B3` 在 `Compute_Blockades_` 後、第二次
`Do_Colony_Calculations_` 前於 `0x13779` 呼叫本函式。它逐玩家反向掃描四位元組記錄：

1. ETA 高 nibble 每回合減一，來源 byte 改寫為 `0xFA`；尚未歸零的記錄保留。
2. ETA 歸零後，以目的行星找星系與殖民地。只有目的星未遭該玩家封鎖、沒有 anomaly、殖民地
   active、owner 仍相同時，才呼叫 `Add_Settler_To_Colony_ @ 0xFF015`。
3. 不論抵達成功或失敗，過期記錄都從陣列移除，`player+0x40` 減一並以 `memmove_` 壓縮。
4. 函式按 360 個目的行星彙整過期抵達，呼叫 `Add_Msg_ @ 0xEF629` 建立 raw type 5 訊息。
   第一個計數是所有過期記錄，第二個計數只在加入人口失敗時增加；精確顯示文案仍未知。

成功加入後緊接的第二次殖民地重算會在同一回合吸收新人口。不存在「在途記錄失敗後永久重試」
的路徑。

## 抵達人口欄位：`Add_Settler_To_Colony_ @ 0xFF015..0xFF116`

函式以不含其他在途者的模式重查容量；失敗回傳 false 且不增加人口。成功時建立 packed colonist：

- race 為記錄 byte 2 low nibble；loyalty bits 4..6 設為 `race & 7`；
- job 為記錄 byte 3 low 2 bits；設 working raw bit 9，清 prisoner raw bit 10；
- 寫入尾端後增加 `colony+0x0A`。

若抵達職務是農夫，函式另讀 `colony+0xE0`。該欄由
`Colony_Environmental_Stuff_ @ 0xE1CED` 寫成可耕作 sentinel：可耕為 `0xFF`，不可耕為
`0x00`，不是數值型農夫容量。不可耕殖民地上的一般抵達農夫會改成 raw job 2；特殊 race slot
`>=8` 的分支可把既有一般農夫改成工人。指令與欄位回寫已證實；特殊 slot 分流的設計理由維持
強推論。

## 閉合與 remake 邊界

- **已證實**：ETA 公式、五艘貨運艦 gate、25 筆上限、容量比較、完整送出檢查、四位元組記錄、
  每回合遞減、封鎖／anomaly／owner 抵達 gate、成功與失敗的移除、人口 packed 回寫及訊息 payload。
- **強推論**：raw type 5 是遷移抵達摘要；payload 計數方向由 producer 已知，但尚未以文字資產
  consumer 證實顯示名稱。
- **未知**：`sub_234B8` 的高層事件名稱、raw flag 4 的正式 UI 名稱，以及 type 5 的精確文案。
  這些不影響人口與貨運艦公式閉合，但仍留在玩家文案／事件索引的 RE 邊界。
- **remake 差異**：目前只解析 `.GAM` 記錄並保存 `SettlersFreighted` 純量，尚無原版等價的
  選擇、在途記錄、回合遞減及抵達人口垂直鏈。依 RE-first gate 本輪不寫 spec、不改玩法程式。
