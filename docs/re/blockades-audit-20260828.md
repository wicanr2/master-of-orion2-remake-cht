# 原版星系封鎖靜態稽核

日期：2026-08-28

## 證據契約

- 輸入 `Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 正式 `Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`；位址是 DOS/4GW
  LE image 的 IDA linear EA。
- 非破壞性證據：[`evidence/blockades-ida-20260828.json`](evidence/blockades-ida-20260828.json)；
  匯出腳本 [`../../tools/ida/audit_blockades.py`](../../tools/ida/audit_blockades.py)。

## 已證實

1. `Compute_Blockades_` 的 raw 函式是 `sub_E5097 @ 0xE5097..0xE5275`。
   星系 stride 是 `0x71`；`star+0x2A` 不是 bool，而是「哪些 player slot 被封鎖」
   的 8-bit mask。`star+0x2B+blockedPlayer` 是「哪些 fleet owner 正在封鎖該
   player」的 8-bit mask。
2. `0xE50A0..0xE50D5` 每次呼叫先清除所有星系的 `+0x2A` 與八個 `+0x2B` byte；
   封鎖是每回合重建，不是只增不減的持久事件旗標。
3. 艦隊 stride 是 `0x81`。`0xE50F1..0xE511F` 只接受 raw `fleet+0x64==0`、
   `fleet+0x65<72`、`fleet+0x11==0` 的記錄，再以 `fleet+0x65` 找所在星、
   `fleet+0x63` 讀 owner。欄位的玩家可見契約是「有效、已抵達、星系合法」；
   更細的內部 status enum 尚未命名。
4. owner `<8` 時，函式逐一掃該星 `star+0x38` 的殖民者 mask；排除艦隊自己，
   並讀艦隊 owner 的 `player+0x627+colonyOwner`。raw policy `4..6` 才寫入
   `star+0x2A |= 1<<colonyOwner` 與
   `star+0x2B+colonyOwner |= 1<<fleetOwner`。既有 enum 已閉合 raw 4／5 為
   Limited War／War；raw 6 必須保留 raw 邊界，不自行改名。
5. owner `>=8` 時，`0xE51A0..0xE51A6` 直接令 `star+0x2A=star+0x38`，即封鎖
   該星所有殖民帝國；這條分支沒有填一般八帝國的 `BlockadedBy` mask。
6. 第二階段逐 blocked player／blockader pair 掃描，只在 blockader 自己沒有該星殖民地時，
   呼叫 `Random_(7)`，取負值後交給 `sub_4E3B5` 關係變更鏈；呼叫 reason raw `5`，
   星系索引也隨 stack 傳入。精確 UI 訊息與政府修飾由共用關係函式負責，不在此重複解釋。
7. 唯一直接 caller 位於 `sub_136B3 @ 0x136B3` 的 `0x13774`。同一主鏈先在
   `0x13751` 執行殖民地 AI、`0x13756` 搜尋戰鬥，再於 `0x13774` 重算封鎖；因此
   本回合職務分配消費的是進入回合時既有 mask，艦隊移動／戰鬥後建立的 mask 供下一輪使用。

## Remake 反證與邊界

- `internal/save.Star` 雖已解析 `Blockaded`／`BlockadedBy`，但 shell 過去匯入時丟棄；
  且把 raw 欄命名成單數容易誤讀為 bool。本輪先以 `Star.BlockadedMask` 與
  `Star.BlockadedBy` 無損保存 `.GAM` 狀態。
- 正常回合仍缺依玩家／熱座／AI 艦隊、星系多殖民 owner mask 與 raw policy 重算的 producer。
- `sub_D61E7` 封鎖殖民地職務 consumer 已定位，但 `colony+0xE0` producer 尚未閉合；
  在此之前不能把未封鎖職務公式套到封鎖殖民地。
