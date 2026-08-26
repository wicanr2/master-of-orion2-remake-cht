# 隨機事件 24「超新星」靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`（1.31），SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`；資料庫使用 `/tmp` 可寫副本。
- 位址基準：IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_random_event_supernova.py`；本輪 JSON：
  `/tmp/moo2-event-monsters-work/random-event-supernova.json`（不加入 Git）。
- 匯出保留原始函式名、位址、bytes、運算元與 callers；附加語意不取代原始定位。

## 已證實

1. `Determine_Event_ @ 0x2230A` 在 elapsed turn 200 前拒絕事件 24；事件 24 不走一般
   `sub_22D57` 帝國權重，而呼叫專用 `sub_23A5F @ 0x23A5F`。
2. `sub_23A5F` 先 `Random(starCount)`，最多重試 1,000 次。候選星需有殖民地，且五個殖民槽中
   至少一筆 active、`colony+0x06==0`、`colony+0x13F==0`；成功回星系索引。這是全銀河星系
   rejection sampling，不是目前玩家殖民地抽樣。
3. 建立端把目標星寫入事件 record `+0x07`，並呼叫 `sub_242FC` 排除同星互斥事件。
4. 倒數公式位於 `0x22C50..0x22C68`：`Random(5)+10-difficulty`，所以五級難度的範圍依序為
   `11..15／10..14／9..13／8..12／7..11`（`Random` 為 1-based）。舊文件只寫「6–14」並不
   符合 1.31 指令；若官方手冊使用 6–14，屬版本／文件差異，不能覆蓋二進位 oracle。
5. 建立端逐一掃目標星五個殖民槽，呼叫 `sub_23B64 @ 0x23B64` 讀每座殖民地 `+0xEB` RP，
   加總後乘倒數；結果同時寫入 record `+0x09` 與 `+0x0B`。因此初始需求是
   `initialSystemRP × countdown`，不是 remake 自創的 `×(countdown+1)`。
6. `sub_206A2 @ 0x2107F..0x21173` 每個 active turn 再掃五個殖民槽：從 record `+0x09`
   扣除各殖民地當回合 `+0xEB`，並呼叫 `sub_E2A70` 重算殖民地／帝國聚合。剩餘需求 `<=0`
   時立即把事件狀態設為 5（成功結束），不再遞減倒數。
7. 若仍有需求，consumer 將 record `+0x06` 倒數減一；倒數歸零時，逐一把受影響殖民地所屬
   planet record `+0x08` 寫為 1，再呼叫 `sub_DCDAC(colony,-1)`，最後把事件狀態設為 5。
   破壞鏈涵蓋該星所有 owner 的 active 殖民槽，不只目前玩家。
8. `sub_E2710`／`sub_23DFE` 的既有證據與上述 consumer 交叉吻合：受事件星殖民地仍保存 RP，
   但不把該 RP 加入一般帝國研究聚合。

## 強推論與未知

- **強推論：**planet `+0x08=1` 對應 Radiated，與資料表及手冊結果一致；本輪未重新命名 raw
  planet 欄位，程式仍保留原位址證據。
- **已接線：**`colony+0x13F` 是 Capitol 建築槽；remake 候選星現在至少需要一座 active 且
  無 Capitol 的殖民地。事件成立後的 RP 與爆發 consumer 仍掃同星全部 active 殖民地。
- 1.50 二進位未取得；倒數若與 1.50 手冊 6–14 不同，須由版本 profile 處理，不能把 1.31
  指令改寫成手冊共同規則。

## 推翻的舊斷言

- 「倒數均勻 6–14」：不符合 1.31；實際為 `Random(5)+10-difficulty`。
- 「需求為自然產出×(倒數+1)」：錯；原版乘倒數。
- 「事件只挑目前玩家殖民地」：錯；原版從全銀河星系 rejection sampling。
- 「爆發只摧毀玩家殖民地並改一顆代表行星」：不完整；原版掃該星五個殖民槽，逐殖民地改
  planet raw climate 並摧毀，不分 owner。
