# 事件怪物 owner 8 航行鏈靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`（MOO2 1.31），SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`。
- 位址基準：IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_event_monster_route.py`；本輪 JSON 為
  `/tmp/moo2-monster-route-20260825/event-monster-route-v3.json`，不加入 Git。
- 註記均為外加導覽；JSON 保留原始函式名、位址、bytes、運算元及直接交叉參照。

## 已證實

1. `sub_A16BF @ 0xA16BF` 先依 raw type 10..14 載入五種事件怪物設計，再以旗標 1 呼叫
   `sub_A1762 @ 0xA1762`。建立成功後才以目標星呼叫 `sub_A1A23 @ 0xA1A23`；事件怪物不是在
   目標殖民星直接建立。
2. `sub_A1762` 的旗標 1 分支會以 `Random(30)`、`Random(5)+1`、目標星 raw `+0x0F/+0x11`
   與畫面／銀河邊界產生外圍座標，最後交給 `sub_100010 @ 0x100010`。後者把設計前
   `0x63` bytes 複製到新 ship，設 raw owner/type 10..14、status 0、目前座標及所在星。
3. `sub_A1A23` 以 `sub_FF799 @ 0xFF799` 規劃路徑、`sub_FFD08 @ 0xFFD08` 寫入 ship
   `+0x64..+0x6D` 航行欄位，最後以目的星 raw 座標呼叫 `sub_EBE79 @ 0xEBE79`。
4. `sub_EBE79` 回傳最小整數 `p` 使 `p²×900 >= dx²+dy²`；亦即
   `ceil(直線距離/30)`。這與既有 `Parsecs_Between_Points_` 證據相同，一 parsec 為 30 raw units。
5. `sub_FF799 @ 0xFF807..0xFF811` 對 raw owner/type `>=8` 固定路徑速度為 1；一般帝國才讀
   `player+0x5A0`。因此事件怪物航速精確為每回合 1 parsec，初始距離值也就是抵達所需回合數。
6. `Move_All_Ships_ @ 0xFFEEA` 只推進 status 1／2 艦艇；`sub_EBB0C @ 0xEBB0C` 每回合按
   `speed×30` raw units 移動，抵達後由 `sub_FFDDA @ 0xFFDDA` 把 status 設回 0、所在星及
   座標改成目的星。太空鰻抵達時另把 ship `+0x61` age 清零。
7. `sub_DB8D8 @ 0xDB8D8` 只對停泊怪物執行後續行為。非太空鰻每回合可經
   `sub_DB6D2 @ 0xDB6D2` 選新目標；選不到時建立 raw 19 訊息並刪船，選到時重新走
   `sub_FF799/sub_FFD08`。太空鰻每 30 回合先分裂，再走同一目標選擇器。

## 強推論、近似與未知

- **強推論：**設計 loader 的 `+0x15` 值不是事件怪物的戰略航速；owner `>=8` 在共同路徑規劃器
  被固定覆寫為 1，故不能以五個 loader 的 `2/5/5/4/2` 當 ETA 除數。
- **近似：**remake 使用正規化銀河座標，沒有原版視窗捲動 globals
  `word_19996E/word_199954/word_199A0A/word_199A0C`。可保留原版 1 parsec/turn 與
  `ceil(distance)`，但外圍出生點先以銀河矩形邊界及 `1..5` parsec 外移重建；不得宣稱逐 RNG
  或逐 raw 座標相同。
- **未知：**`sub_DB6D2` 的完整 12-byte 星系評分 record、排序 callback `sub_DB659` 與殖民地
  攻擊戰後 consumer 尚未形成 typed 垂直鏈。本切片只閉合首次事件怪物的航行與抵達。

## 推翻的舊結論

- 「事件建立後怪物立即盤據目標星」：錯；原版先建立 owner 8 航行艦艇。
- 「五個怪物依 loader 設計速度飛行」：錯；戰略路徑對 owner `>=8` 固定速度 1。
- 「`sub_EBE79` 只是無關距離級別」：不完整；配合固定 1 parsec/turn，它正是首次航程 ETA。
