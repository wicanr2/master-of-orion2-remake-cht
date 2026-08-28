# Ship Attack／Defense／Trans-Dimensional 原版稽核（2026-08-28）

## 範圍

本切片審查三個 runtime signed byte：

- `player+0x8A5`：Ship Defense，客製值 `-20／+25／+50`。
- `player+0x8A6`：Ship Attack，客製值 `-20／+20／+50`。
- `player+0x8BC`：Trans-Dimensional，布林值 0／1。

完整 IDA Pro 9.4 指令、bytes、caller 與外部符號位於
`evidence/custom-race-trait-consumers-ida-20260828.json`。本文件分開「格子戰術實際命中」、
「戰略／快速資料載入與估值」及「星圖移動」，不以單一 bonus helper 代替所有路徑。

## Ship Defense

### 格子戰術實際解算

- `Defensive_Combat_Bonus_ @ 0x35D0D`：一般 owner `<8` 時，將 signed `+0x8A5` 原值加入
  defensive combat bonus。caller 包含 `Combat_Bonus_`、`Set_Ocv_Dcv_For_All_`、
  `Tactical_Bombardment_` 與 `Tactical_Combat_`，因此是格子戰術主 DCV 鏈。
- `Missile_Dcv_ @ 0x3DFE0`：一般 owner 且有效 player index 時，把 `+0x8A5` 原值加入 missile
  defense 值。caller 包含 `Average_Missile_Kills_`、`Fire_Beam_Weapon_` 與
  `Fire_Fighter_Beam_`；同一防禦特性會進飛彈／戰機光束命中下游。
- `Determine_Beam_Weapon_Modifications_ @ 0x60B59`：自動設計／武器改造評分以
  `5×armor-derived base + ShipDefense` 建立防禦值，再從攻方電腦／命中基礎扣除。唯一 caller
  是 `Add_Beam_Weapon_To_Design_`，這是設計 AI consumer，不是戰鬥時重算。

### 戰略／快速資料與估值

- `Qload_Ships_ @ 0x416CF`：快速／戰略 combat record 初始化時，局部 defensive 欄先取
  `10×crew/level-like value`，再加 Ship Defense 原值。caller 是 `Strategic_Combat_` 與
  `Strategic_Bombardment_`。
- `Base_Generic_Dcv_ @ 0x5EAE9`：`5×armor defense table + ShipDefense`；若 Trans-Dimensional
  再加 20。它被殖民地衛星強度與 `Ship_Strength_Vs_Shield_` 使用。
- `Get_Ship_Combat_Bonuses_ @ 0x54E5B`：Ship Defense 全值加入一個輸出，另以
  `truncTowardZero(ShipDefense/2)` 加入第二輸出；caller 是詳細艦艇資訊、艦艇強度估值與 Fleet
  掃描資料。這是顯示／戰略估值鏈，不能取代格子戰術 `Defensive_Combat_Bonus_`。

## Ship Attack

### 格子戰術實際解算

- `Offensive_Combat_Bonus_ @ 0x366DD`：一般 owner `<8` 時把 signed `+0x8A6` 原值加入 OCV。
  caller 包含 `Combat_Bonus_`、`Set_Ocv_Dcv_For_All_`、`Tactical_Bombardment_` 與
  `Tactical_Combat_`。
- `Fighter_Ocv_ @ 0x3DF8D`：戰機 OCV 為武器／速度基礎加 Ship Attack 原值，再加
  `Get_Fleet_Pilot_Bonus_` 與固定 50。兩個 caller 都在 `Fire_Fighter_Beam_`，所以敵我戰機
  精確命中鏈確實消費 owner 的 Ship Attack，不是只在艦艇面板顯示。
- `Do_Auto_Ship_Turn_ @ 0x29837`：戰術 AI 的攻擊評分亦直接加 Ship Attack；唯一上游是
  `Do_Combat_Turn_` 的兩個呼叫點。

### 戰略／快速資料與估值

- `Base_Generic_Ocv_ @ 0x5EB39`：電腦／武器表基礎加 Ship Attack 原值。caller 包含
  `Qload_Ships_`，因此快速／戰略 combat record 的 OCV 會消費此特性。
- `Satelite_Strength_Vs_Ships_ @ 0x5EB72` 與 `Ship_Strength_Vs_Shield_ @ 0x5EF4B` 亦直接把
  Ship Attack 加入殖民地太空防禦及艦艇強度評估。
- `Get_Ship_Combat_Bonuses_` 把 Ship Attack 全值加入 offensive 輸出，服務詳細資訊與戰略估值。

### 開局 AI profile

`Init_NPC_Personalities_Objectives_Themes_ @ 0x589D6` 對 Ship Defense 20／40 與 Ship Attack
25／50 使用不同 accumulator 權重。這證明內建種族 raw 值與客製 UI 值不完全同一組檔位；
profile accumulator 的正式欄位仍由 AI profile 切片判定，不在此文件猜名。

## Trans-Dimensional：戰鬥速度

### 格子戰術艦艇速度

- `Calc_Current_Speed_ @ 0x4528F`：一般 owner 具 Trans-Dimensional 時，current speed 加 **4**。
  caller 覆蓋回合初始化、戰術主迴圈、登艦、牽引光束、防禦射程與艙損。
- `Load_Display_Combat_Ship_ @ 0x4CB1E`：詳細艦艇顯示的速度欄同步加 4。
- `Get_Ship_Combat_Bonuses_ @ 0x54E5B` 與 `Base_Generic_Dcv_ @ 0x5EAE9`：防禦／強度估值加
  **20**。這與速度 `+4` 是兩個不同尺度，不能互換。

### 飛彈速度

`Missile_Speed_ @ 0x3CD21` 依飛彈類型得到不同 base；Trans-Dimensional 一律加 **4**：

| 原 base | Trans-Dimensional |
|---:|---:|
| 8 | 12 |
| 10 | 14 |
| 6 | 10 |

caller 是 `Fire_Missile_`、`Move_Missile_`、`Resolve_Missile_` 與 `Missile_Dcv_`，因此同一速度
同時進部署、逐 tick 移動、命中解算與防禦估值。

## Trans-Dimensional：FTL 與超空間亂流

### FTL 速度

- `Calc_Player_FTL_Speed_ @ 0x57597`：最佳 warp drive 表值加 **2**。
- `Get_FTL_Speed_ @ 0x575D6`：指定玩家／drive 的表值加 **2**。

caller 覆蓋玩家科技初始化、全艦航程更新、建造中的艦艇、自動／戰略設計及玩家設計更新，故
這是設計與 runtime 共用的 FTL producer，不是 UI-only 值。

### Hyperspace Flux 免疫

`Event_Check_Hyperspace_Flux_ @ 0x233FA` 成立時，普通帝國會被下列路徑擋住；
Trans-Dimensional 可直接繞過：

- `NPC_Declarations_Of_War_ @ 0x25DF1` 的特殊戰爭分支。
- `Move_All_AI_ @ 0xDBB29` 的 AI 移動。
- `Find_Opportunity_Attack_ @ 0xDBC5C` 的機會攻擊搜尋。
- `Player_Threatens_Player_ @ 0xDBCC8` 的外交威脅資料。
- `Ships_Try_To_Move_To_ @ 0xFF799` 的通用下令／航路函式；它有 38 個 caller，涵蓋玩家、AI、
  運輸艦、殖民艦、攔截、安塔蘭與事件艦艇。

此結論與 `random-event-hyperspace-flux-audit-20260825.md` 的事件主鏈一致。

### 撤退

`Determine_Retreat_Ships_ @ 0xE6CAA` 對一般帝國檢查 `+0x8BC`；具 Trans-Dimensional 時跳過
一個普通帝國限制分支。兩個 caller 均在 `Do_1_Combat_`。該分支後續仍包含艦艇狀態、星系與
目的地判定，因此只證實「放寬撤退 gate」，不宣稱可無條件撤退。

## 兩條戰鬥路徑閉合判定

| 效果 | 格子戰術 | 快速／戰略 | 狀態 |
|---|---|---|---|
| Ship Attack | `Offensive_Combat_Bonus_`、`Fighter_Ocv_`、戰術 AI | `Base_Generic_Ocv_`→`Qload_Ships_`、艦艇／衛星估值 | 主要數值 consumer 已證實 |
| Ship Defense | `Defensive_Combat_Bonus_`、`Missile_Dcv_` | `Qload_Ships_`、`Base_Generic_Dcv_`、艦艇估值 | 主要數值 consumer 已證實 |
| Trans-Dimensional | current speed +4、missile speed +4 | generic DCV／估值 +20；FTL +2 屬星圖鏈 | 主要數值 consumer 已證實 |

### 尚未閉合

- `Qload_Ships_` 各局部欄位對 33-byte strategic combat record 的正式 offset 名稱。
- `Get_Ship_Combat_Bonuses_` 四個輸出參數的完整型別／正式名稱；目前只保留相對算式。
- Trans-Dimensional 的撤退限制分支完整條件。
- `Calc_Tech_Value_` 對三項特性的 raw tech ID／category。
- AI profile accumulator 的正式欄位。
- 全域 RNG 與命中擲骰逐位元序列；本切片只閉合 bonus producer／consumer。

因此三項能力的主要玩家可見數值與兩條戰鬥路徑已由原始指令證實，但上述 record 命名與周邊
分支未完成前，仍只可標「主要 consumer 閉合」，不可宣稱整個戰鬥 parity 完整。
