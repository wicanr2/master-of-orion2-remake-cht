# Fighter Garrison 格子戰術建立、部署與回寫鏈（2026-08-28）

## 結論

原版 Fighter Garrison 不是在戰鬥開始時直接放置十／六／四個獨立 token。它先把行星本體建立為
313-byte combatant，再依最高戰機科技把兩個 fighter weapon slot 裝到該 combatant；每個 slot
只可發射一次，發射時每個「中隊」展開為四架 26-byte fighter runtime。因此總量為：

| raw kind | 技術分支 | 每槽中隊 | 槽數 | 總中隊 | 總架數 |
|---:|---|---:|---:|---:|---:|
| 31 | 預設攔截機 | 5 | 2 | 10 | 40 |
| 30 | 轟炸機 | 3 | 2 | 6 | 24 |
| 29 | 重戰機，優先於轟炸機 | 2 | 2 | 4 | 16 |

raw 數值、優先序、兩槽、每槽數量與 `×4` 均為**已證實**；三個中文機型名稱依科技欄、武器表及
既有 fighter runtime 分流交叉驗證，列為**強推論**。戰機由行星 combatant 的實際部署座標發射，
不是另外猜一個駐防出生點。

戰鬥中的個別損失只改 runtime 與兩個 weapon slot；存活架數除以四後才回收。沒有任何路徑把
剩餘架數寫回 colony record，所以 Fighter Garrison 只要建築仍在，下一場會重新建立完整
10／6／4 中隊。若行星防禦受創時隨機抽中 raw building ID 47，原版立即移除 Fighter Garrison，
並清空行星 combatant 所有 fighter slot；已在外面的 runtime 不由該 helper 強制刪除。

## 證據契約

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 外部符號索引 SHA-256：
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`
- 工具：IDA Pro 9.4／IDAPython；位址基準為 IDA linear、DOS/4GW LE object #1
- 探針：`tools/ida/audit_fighter_garrison_tactical.py`
- 非破壞性證據：
  `docs/re/evidence/fighter-garrison-tactical-ida-20260828.json`

探針匯出 19 個 gameplay root、完整原始指令、bytes、caller／callee、Hex-Rays 導覽文字及
colony 防禦欄位 operand site。正式 `.i64` 唯讀掛載後複製到 `/tmp`；沒有改名、套型別或寫回。
Hex-Rays 的函式參數與區域變數只供導覽，以下結論以 raw instructions 與資料流為準。

## 快速結算與格子戰術建立器分流

- `Qload_Ships_ @ 0x416CF` 建立 33-byte 快速結算記錄，供 `Strategic_Combat_` 與戰略轟炸使用。
- `Tactical_Combat_ @ 0x47939` 在 `0x48546` 呼叫
  `Load_Colony_Defense_ @ 0x4A9E9`；後者再呼叫
  `Load_Tactical_Colony_ @ 0x4AA36`，建立 313-byte 格子戰術行星 combatant。

兩份資料結構、武器欄與消費端不同。本文只用後者證明格子戰術，不拿快速結算 record 代替。

## Fighter Garrison 兩槽 producer（已證實）

`Load_Tactical_Colony_` 以 `colony + 0x165` 判定建築是否存在；由已驗證的 49-building layout，
`0x136 + 47 = 0x165`，即 raw building ID 47。原始指令 `0x4ADCE..0x4AE56`：

1. 預設 `BX=31`、`EDX=5`。
2. owner `player+0x16A == 3` 時改為 `BX=29`、`EDX=2`。
3. 否則 owner `player+0x136 == 3` 時改為 `BX=30`、`EDX=3`。
4. 迴圈固定跑兩次；每次建立一個 11-byte weapon slot：
   - `+0x52`：raw kind 29／30／31；
   - `+0x54` 與 `+0x5B`：每槽中隊數 2／3／5；
   - `+0x58 = 1`、`+0x5A = 1`：可用／單次發射狀態；
   - raw flags `+0x56 = 0`。

重戰機欄先判斷並直接跳到建槽點，所以兩項科技同時存在時重戰機優先。

## 從中隊到逐架 runtime（已證實）

`Fire_Missile_ @ 0x3C892` 從母體 313-byte record 與選定的 11-byte slot 建立一筆 26-byte
runtime。武器表類別 `byte_17F80F[28*kind] == 4` 時，`0x3CA86..0x3CAAB` 將 slot 數量乘四寫入
runtime `+0x0F`。因此兩槽實際產生 40／24／16 架。

runtime 保存母體 combatant index、目標、座標、存活數與攻擊次數。kind 29／30／31 的光束、
炸彈、目標失效、返航與母體八槽回收已由
`fighter-runtime-audit-20260828.md` 閉合；本文件不重複逆向同一 consumer。

## 部署與發射位置（已證實）

`Deploy_Ships_ @ 0x49043` 把行星 combatant 保持為部署基準 record，依戰場方向設定：

- 一側為 `x=65, y=34`；
- 另一側為 `x=10, y=34`；
- 行星 combatant 的部署 class／size byte 在排序後設為 3。

`Fire_Missile_ @ 0x3CAFC..0x3CB22` 讀母體 record `+0x21／+0x22`，各乘 20 再加
`Ship_Center_Offsets_`，形成 runtime 的初始像素座標。故駐防戰機從行星 combatant 的中心發射；
戰場方向會鏡射 x，但沒有獨立的 Fighter Garrison 出生格。

## 建築毀損與戰後狀態（已證實）

`Does_Combat_Planet_Have_Defenses_ @ 0x3A142` 把四棟防禦建築映成 bit 1／2／4／8；
Fighter Garrison 的 `colony+0x165` 是 bit 8，會參與戰鬥勝負與目標判定。

`Apply_Damage_To_Planet_ @ 0x3A3C3` 累積行星防禦傷害；每越過一個裝甲修正後的 100 點單位，
便呼叫 `Destroy_Colony_Defense_ @ 0x3A19E`。後者只從目前存在的 raw building
ID 26／27／42／47 建立候選，以 `Random(candidateCount)-1` 抽取並呼叫
`Remove_Building_(colony, rawID)`。抽中 47 時，`0x3A36F..0x3A39D` 掃八個 weapon slot，
對武器表類別 4 的 slot 將 raw kind 與數量清零。

返航 consumer 只把 `runtime survivors / 4` 回填同 kind 的母體 weapon slot，沒有寫
`colony+0x165` 或其他持久人口／建築欄。由此可得：

- 個別戰機損失只持續到本場戰鬥；
- 建築被摧毀是唯一的 Fighter Garrison 持久化損失；
- 清槽 helper 不主動掃除已發射 runtime，後者仍依既有母體存活／失效規則完成或銷毀。

## Remake 對應與邊界

- remake 已有一般艦載戰機 runtime，但格子戰術沒有從 colony Fighter Garrison 建立上述兩槽，
  也沒有原版 40／24／16 架與防禦建築被抽中後的清槽鏈。
- remake 的 `FighterGarrisonStrategicStrength` 只處理快速／抽象強度；它不能證明格子戰術已接。
- 圖片、音效與 fighter 抖動動畫不影響本鏈的數量、攻擊或回寫；只保存播放邊界，不納入 RE 分母。
- compiler helper、`qsort`、`memset` 及繪圖／音訊平台內部不判讀；只保留排序結果與初始化契約。

原版 1.31 的 Fighter Garrison 從建築欄到兩槽、逐架 runtime、部署、攻擊、返航、建築毀損與
持久化邊界至此閉合。正式 UI 字句與逐像素 fighter sprite 不影響玩法規則，分列在視覺資料盤點。
