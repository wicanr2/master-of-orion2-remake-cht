# Cybernetic／Lithovore 玩家可見消費鏈稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4、`tools/ida/audit_custom_race_trait_consumers.py`；位址均為
  IDA linear、DOS/4GW LE object #1。
- 可重生證據：
  [`evidence/custom-race-trait-consumers-ida-20260828.json`](evidence/custom-race-trait-consumers-ida-20260828.json)。
  匯出保留原始函式名、位址、bytes、operand、caller 與外部導航符號。

## Cybernetic `player+0x8B0`

### 食物與工業維護

`Colony_Food_Maintenance_ @ 0xDEB4B` 逐 packed colonist 計 half-food：一般人口為 2，
Cybernetic 為 1，Lithovore 為 0；最後以 `(sumHalf+1)/2` 寫整數維護。`Colony_Industry_Maintenance_
@ 0xDF546` 則令 Cybernetic 每人口消耗 1 half-industry，Android 為 2、其餘一般人口為 0。
分類與食物複製機／建造 consumer 已在既有殖民地鏈閉合，不能把兩個半單位提早合併取整。

### 格子戰術回合內修復

`Repair_All_Combat_Ships_ @ 0x4CFE7` 的唯一 caller 是
`End_Of_Turn_Bookeeping_ @ 0x4A575`。它分別取得雙方 `Get_Fleet_Engineer_Bonus_ @ 0x35E0C`
的最高值，再對 Cybernetic 一側加 **10**；艦上符合 raw special gate 時另加 20。總和作為
百分率傳給 `Repair_Combat_Ship_ @ 0x35821`。

`Repair_Combat_Ship_` 直接以 `current-or-maximum structure × rate / 100` 與
`system capacity × rate / 200` 計算修復量，最低 1；依序修復結構／裝甲、兩個主要系統、
八個武器槽、特殊裝置容量與另一組八槽。因此 Cybernetic 的 `+10` 是每個格子戰術回合的
**十個百分點維修率**，不是固定 10 HP，也不是這個函式內立即完全修復。

### 戰後與戰略戰鬥

`End_Of_Combat_ @ 0x4B184` 在倖存艦持久化前，把 Cybernetic 與 Advanced Damage Control
科技放在和工程師修復相同的清損 gate；它重建 combat ship 狀態並清除系統／特殊裝置損傷，
所以「戰後完全修復」仍成立，但與上述回合內 `+10%` 是兩條不同路徑。

`Strategic_Combat_ @ 0x40148` 在相同 Cybernetic／科技 gate 下把 33-byte combatant
`+0x1F` 歸零；該欄已由戰後 consumer 證實會回寫 `ship+0x7D` 結構損傷。故快速／AI 戰略
解算也不留下 Cybernetic 倖存艦的持久結構損傷。

### AI 與殖民地評估

`Uncolonized_Planet_Worth_To_Player_`、`Colony_Worth_To_Player_` 與 NPC profile 初始化均有
Cybernetic 分支，但完整權重還混合氣候、產出與 personality accumulator；本文件只定位其
存在，不由鄰近立即數猜完整 AI 公式。

## Lithovore `player+0x8B1`

### 零食物維護與初始職務

同一 `Colony_Food_Maintenance_` 令 Lithovore 一般人口 half-food 為 0。新殖民地與新增人口
職務分支把 Lithovore／Cybernetic 視為不需優先補一般農業的種族；開局第一人口在對應 gate
下指派為 worker，而不是 farmer。AI 的 `Ensure_At_Least_1_Food_Planet_` 亦對 Lithovore
直接略過一般糧食行星保障流程。

### 科技合法性

`Tech_Is_Legal_For_Player_ @ 0x5E481` 在 Lithovore 成立時令下列六項 application 不合法：

| raw ID | 受版控 enum／名稱 |
| ---: | --- |
| 6 | `TECH_ANDROID_FARMERS`／Android Farmers |
| 29 | `TECH_BIOMORPHIC_FUNGI`／Biomorphic Fungi |
| 68 | `TECH_FOOD_REPLICATORS`／Food Replicators |
| 87 | `TECH_HYDROPONIC_FARM`／Hydroponic Farm |
| 162 | `TECH_SOIL_ENRICHMENT`／Soil Enrichment |
| 178 | `TECH_SUBTERRANEAN_FARMS`／Subterranean Farms |

這是研究 application 的合法性 gate，不表示整個 topic 不可研究；同 topic 的其他選項仍須按
Creative／Uncreative 與正常選擇規則處理。

### AI 殖民地資料

`Compute_AI_Data_` 分開檢查主要人口與 owner 是否皆為 Lithovore；只有兩者同時成立時才清除
一般食物建築需求旗標。混合殖民地不能只看 owner trait。行星／殖民地 worth 另有 Lithovore
分支，但完整分數仍保留為 AI 窄切片。

## 勘誤、remake 差異與停止線

- 舊文件只強調 Cybernetic「戰後完全修復」，漏掉格子戰術每回合 `+10%`。兩者均為已證實，
  不能互相替代。現行 remake 的 `repairAfterBattle` 只覆蓋戰後結果，尚缺回合內修復規格與
  可表示逐系統損傷的資料模型。
- Lithovore 六項科技禁用表、零食物維護與 AI 食物保障 bypass 已證實；完整 AI worth 權重仍待。
- `isqrt_`、除法輔助碼、C runtime、Watcom stack probe 與平台 API 不納入玩法分母。
