# 艦隊派遣、逐座標移動、截擊與抵達稽核

## 證據邊界

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址皆為 IDA linear、DOS/4GW LE object #1。
- 可重生匯出：[`fleet-interstellar-movement-ida-20260828.json`](evidence/fleet-interstellar-movement-ida-20260828.json)，
  腳本：[`audit_fleet_interstellar_movement.py`](../../tools/ida/audit_fleet_interstellar_movement.py)。
  外部符號只用於導航；結論以原始指令、交叉參照與資料流為準。

## 已證實

### 個別艦艇就是移動狀態的權威資料

原版沒有另外以一個「艦隊 ETA」取代途中狀態。129-byte ship record 的玩家可見移動欄位為：

| raw offset | 已證實用途 |
|---|---|
| `+0x63` | owner |
| `+0x64` | status；0 為停泊，1／2 為兩種移動編碼 |
| `+0x65` | 停泊時為 star；移動時保存加 500 或加 1000 的 destination 編碼 |
| `+0x67／+0x69` | 目前銀河 x／y |
| `+0x6B` | Navigator 路徑豁免旗標 |
| `+0x6C` | 本航程實際速度，單位為 parsec／turn |
| `+0x6D` | 顯示 ETA；每回合遞減，建築／殖民地所有權改變後可由 `Update_Ship_ETAs_` 重算 |
| `+0x7F` | AI 截擊／途中狀態標記；抵達或非移動狀態會清零 |

`Fltscrn_Move_Ships_ @ 0x75D48` 先把玩家勾選的 ship ID 寫入 `-1` 結尾清單，再交給
`Ships_Try_To_Move_To_ @ 0xFF799` 與 `Make_Ships_Move_To_ @ 0xFFD08`。後者逐 ship record
寫相同目的編碼、速度與 ETA。因此「拆分」只是選擇部分 ship ID 改寫；「合併」是多艘船具有
相同 owner、status、座標與目的後由畫面／戰鬥 consumer 視為同組，沒有另一份可獨立漂移的艦隊數量。

### 速度、座標與 ETA

銀河座標每 30 單位是一 parsec。`Move_Player_1_Turn_To_Star_ @ 0xEBB0C` 取得目標星座標，
以整數平方根求距離；對一回合內第 `i=1..speed` 個 parsec，按方向向目標投影 `30*i`，負方向
採對稱公式。若 `speed²×900 >= distance²`，該回合直接寫精確目的座標。

`Move_All_Ships_Toward_Stars_ @ 0xFFEEA` 是 `Next_Turn_Calc_` 唯一 caller：逐一掃描所有
status 1／2 ship，解碼目的，依 `+0x6C` 從目前 `+0x67／+0x69` 實際前進，`+0x6D` 減一；
只有座標等於目的星時才呼叫 `Make_Ship_Arrive_At_Star_ @ 0xFFDDA`。因此 ETA 是衍生顯示值，
不是移動真值。

一般帝國基礎速度來自 `player+0x5A0`；派遣時再取選中艦艇最低航程，並取其軍官最高 Navigator
加成。Navigator raw skill bit `0x02` 使用 `floor(level/3)+1`，bit `0x01` 使用
`floor(level/4)+1`，取最大值加到基礎速度。`Ship_Range_ @ 0xFF496` 則讀 `player+0x324`，
有 ship special bit 11 時變為 `(3×range+1)/2`。

### 星雲、黑洞與曲速場干擾器

- 星雲：`Point_Is_In_Nebula_N_ @ 0xEB9C8` 以 `(point-nebulaOrigin)/3` 索引實際遮罩，palette
  index `>5` 才算在內。逐回合移動每走一 parsec 檢查一次；第一次踏入就停止本回合後續步數，
  所以實際速度降為 1。Navigator 只會讓此檢查指標為 null，因而豁免星雲。
- 黑洞：`Initialize_Black_Hole_Blocks_ @ 0xEB87D` 先枚舉 spectral raw 6，為每一對非黑洞星
  建立 `star+0x1F` 起的 9-byte 阻擋 bitfield；從途中座標派遣則呼叫
  `Black_Hole_Blocks_Points_ @ 0xEB7FD` 即時計算。無 Navigator 且沒有 gate 例外時，命中即拒絕
  派遣，不是航行中扣速。
- 曲速場干擾器：`Point_Is_In_Warp_Field_Interdictor_Of_Star_ @ 0xEBAC3` 檢查
  `star+0x39` 是否有 owner 以外的 bit，並以 `dx²+dy² <= 8100 = (3×30)²` 判定半徑。
  逐步移動踏入後即停止該回合後續步數，故同樣降為 1 parsec／turn；Navigator 不會跳過此檢查。

### Gate、航程與補給

兩端 `star+0x3D` 的 owner bit 同時成立時，派遣結果直接是一回合；兩端 `star+0x3E` 同時成立
時，實際速度加 3。前者與 Star Gate、後者與 Jump Gate 的正式名稱由手冊及既有建築表交叉支持；
raw mask、`1 turn` 與 `+3` 的 consumer 是已證實，名稱映射為強推論。

一般航程以 `range²×900` 和星座標距離平方比較，並允許 treaty raw 2 的盟友殖民地作補給來源。
超出可達範圍會拒絕派遣；不是先放行、抵達時才裁決。超空間亂流另在非本回合玩家／非事件豁免
情況下直接擋下派遣。

### 中繼、截擊與抵達 consumer

`Move_Ships_With_Possible_Intermediate_ @ 0xD7923` 在直達不可行時掃描己方殖民地，暫時把選中
ship 的 star／座標設為候選點，分別計算兩段合法 ETA，選總回合最少者；這是 AI 殖民船、運輸艦
與 staging 路徑的玩家可見中繼規則。

`Interceptions_ @ 0xD9A7E` 會以 `Ships_Try_To_Move_To_` 評估移動中的目標群，選定後仍透過
`Make_Ships_Move_To_` 寫入個別 ship 的動態目的編碼，並以 `+0x7F=5` 標記途中截擊狀態。
其大段威脅／護航評分屬「原版 AI 決策器」矩陣列，本文件只關閉移動表示與下令 consumer，
不把整個 AI score 宣稱已完成。

抵達時 `Make_Ship_Arrive_At_Star_` 會設 status 0、寫目的 star 與精確 star x／y、清 ETA／截擊
標記、更新 star 的到訪／探索 mask、觸發首次探索事件，並標記真人報告刷新。`Search_For_Battles_`
在回合主鏈中緊接間諜、帝國與殖民地套用之後執行，使用已抵達的 ship 狀態搜尋戰鬥；途中 ship
不會因 ETA 歸零以外的 remake 代理自動加入星系戰。

## 強推論與未知

- status 1／2 與 `+500／+1000` 的位元級解碼、更新與事件 consumer 已證實，但兩種 status 的
  正式 UI 名稱尚未由字串資產閉合；不可僅依外部符號替它命名。
- `star+0x3D／+0x3E` 對 Star Gate／Jump Gate 的名稱是手冊與效果交叉得到的強推論；raw 行為已證實。
- 原版 PRNG 不參與直線座標步進；AI 截擊候選的完整 score 與 tie-break 留在 AI 決策器列。
- 畫面如何把同 key 個別 ship 聚成一個 icon 不影響移動規則，屬艦隊 UI 資料投影。

## 對 remake 的判定

目前 remake 用 `Fleet.ETA` 表示整段航程，玩家與 AI 又是不同資料模型；途中沒有每艘船座標、
動態目標、中繼與原版同 key 聚合，所以即使端點 ETA 看似合理，也不是原版垂直鏈。既有直線取樣
已接近星雲／黑洞／干擾器的 route gate，但原版是每回合逐 parsec 消費，而不是派遣時計算一次
固定 ETA。依 RE-first gate，本輪只登記差異，不修改 Go／Ebitengine 行為。
