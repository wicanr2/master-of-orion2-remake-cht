# 匿蹤裝置格子戰術 consumer 稽核（2026-08-30）

## 範圍與證據

本切片只追 raw special 6（Cloaking Device）、23（Phasing Cloak）、31（Stealth Field）從
艦艇設計 bitfield 進入 313-byte 格子戰術記錄後的玩家可見 consumer。IDA Pro 9.4 線性位址、
原始運算元、完整函式與局部呼叫端保存在：

- `evidence/cloak-tactical-consumers-ida-20260830.json`
- `evidence/cloak-state-offsets-ida-20260830.json`

輸入 `Orion2.exe` SHA-256 為
`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`；IDA 資料庫 SHA-256
為 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。外部符號只供導覽，
所有結論仍以 raw 位址、bytes、0x139 stride 與資料流為準。

## 已證實：裝置載入與狀態機

- `Load_Combat_Ship_ @ 0x4954A` 把設計記錄 `+0x17／+0x76` 的五個特殊裝置 bitfield bytes
  複製到戰鬥記錄 `+0x4C／+0xB2`；後者是停用／損壞遮罩。
- `Does_Combat_Ship_Have_Special_ @ 0x4B0D3` 只有在 present bit 為 1 且 disabled bit 為 0
  才回傳 true。
- `Init_Special_Devices_ @ 0x4C9F6` 將 raw 6 初始化為 `state +0x40 = 1`；若沒有 raw 6
  而有 raw 23，初始化為 `state = 4`、`counter +0x41 = 10`；兩者皆無則 `state = 0`。
  raw 31 不進此狀態機。
- `Fire_Ship_ @ 0x38B5E`、`Do_Combat_Turn_ @ 0x42F7F` 與
  `Initiate_Boarding_Action_ @ 0x459E7` 都在開火／執行動作前把 `1→2`、`4→6`，也就是解除
  Cloaking／Phasing 的防護狀態。`Init_Ship_For_Start_Of_Turn_ @ 0x42B70` 再把 `2→3`、
  `6→7`；回合尾 `Do_Combat_Turn_ @ 0x44D85..0x44E04` 把 `3→1`、`7→4`。
- Phasing 只有在 `state == 4` 時於每回合尾遞減 `+0x41`；由 1 減至 0 時依序播放狀態 5、1
  的 cloak 畫面，再把狀態降為 1。因此「10 回合後降級為 Cloaking Device」已由 raw
  狀態轉換證實，不是手冊近似。

## 已證實：目標、飛彈與防禦

- `Select_Ship_To_Target_ @ 0x2A46A`、`Get_Current_Ship_Target_ @ 0x2ACF9`、
  `Retarget_Missile_ @ 0x3DDD8` 與 `Defensive_Fire_Check_ @ 0x36ED1` 都排除 `state == 4`；
  `Move_Ship_ @ 0x3EE0F` 也不對該狀態觸發 defensive fire。這閉合 Phasing 的不可鎖定與
  現存飛彈改鎖鏈。
- `Ship_Specials_Defensive_Bonus_ @ 0x36A63` 在 `state == 1` 且模式參數為 0 時加入
  `0x50 = 80`。`Defensive_Combat_Bonus_ @ 0x35D0D` 的呼叫端把模式 0 結果存入戰鬥記錄
  `+0x36`，模式 1 存入 `+0x38`；故 Cloaking Device 的 +80 只進第一種防禦值，不會錯加到
  第二種。欄位的玩家名稱沿用既有 OCV／DCV 規格，不以本切片另行猜名。
- `Resolve_Missile_ @ 0x3D2DF` 在 `state == 1` 時執行 `Random(100) < 50`，命中該分支即把
  局部命中比較值設為 `-1000`；`Expected_Missile_Damage_ @ 0x28EAE` 對相同狀態把預期傷害
  做向零除二。這同時閉合實際 50% miss 與戰術 AI 的期望值 consumer。

## raw 31 的停止條件

全庫 `+0x40／+0x41` 直接運算元普查與三個特殊裝置的初始化鏈中，raw 31 Stealth Field 沒有
進入 cloak 戰術狀態；它仍只屬於先前已證實的戰略星圖 concealment 與 AI 設計 consumer。
這是本版 executable 的**已證實資料流邊界**，不是宣稱任何間接指標都不存在。若未來找到能把
raw 31 寫入 `+0x40` 的新 producer，才重開此切片；目前不再把它列為格子戰術公式缺口。

## Remake 對照

`internal/shell/cloak.go` 已具備 +80、飛彈 50% miss、動作後解除、完整停火回合重隱、Phasing
10 回合不可鎖定後降級等玩家規則。本輪只補證據與規格，不改 Go；後續應另以雙戰鬥路徑測試
確認快速結算的手冊近似與格子戰術 raw 狀態機沒有被誤合併。
