# Telepathic 原版消費端稽核（2026-08-28）

## 範圍與證據

本切片承接 `custom-race-trait-consumer-census-20260828.md`，只審查已確認為 Telepathic 的
`player+0x8B8`。IDA Pro 9.4 證據位址、bytes、完整函式、直接 caller 與外部符號名稱均保存在
`evidence/custom-race-trait-consumers-ida-20260828.json`；輸入雜湊與位址基準同該檔。

以下「已證實」只表示原始指令與 caller 資料流已閉合。外部符號名稱用於導覽，不取代
`sub_xxx`、raw 位址或運算元。

## 一、間諜與外交

### `Compute_Spy_Bonuses_ @ 0x100A83`

已證實攻守雙方共用的基礎項含：

```text
base = 10 × signed(player+0x8B8)
     + signed(player+0x8A8)
     + Tech_Spy_Bonus_(player)
```

之後再各自加入領袖技能與政府查表項。直接 caller 是 `Resolve_Spies_ @ 0x10192B` 與
`Draw_Race_Text_ @ 0x10BFBD`，因此同一個值同時進玩法解算與種族資訊 UI。Telepathic 的 raw
布林值為 1 時提供 **+10**，不是任意百分比。

### `Diplomacy_Test_ @ 0x53146`

`0x5338E..0x53398` 已證實受測玩家 `+0x8B8 != 0` 時，外交檢定 accumulator 加 **25**。
同函式相鄰分支對 Repulsive 減 50；五個直接 caller 全位於
`Check_Treaty_Proposal_ @ 0x5272D`。本切片只閉合這個 +25 項，不把完整 treaty proposal
公式重複搬入此文件。

### `Calc_Tech_Value_ @ 0xFC845`

`edx` 入口參數原樣保存為 `var_4`，`0xFCDE9..0xFCE46` 證實這是 raw tech ID 特例，不是
category switch：只有 **raw tech 5** 時檢查 Telepathic，命中便把局部估值設為 1。受版控 enum
把 raw 5 對回 `TECH_ALIEN_MANAGEMENT_CENTER`；raw 131 是相鄰但獨立的 Aquatic／Subterranean
特例。六個 caller 服務科技交換、需求回應與 AI 科技應用選擇，所以這是 AI／外交估值權重，
不是額外玩家能力。

## 二、戰術登艦與俘獲艦

### `Boarding_Action_Type_ @ 0x2C129`

當攻方算出的登艦強度大於守方門檻 `defense-20` 時：

- 攻方 owner `+0x8B8 != 0`：在 `0x2C3EB..0x2C427` 直接回傳 raw action type **2**。
- 非 Telepathic：還要比較雙方 combat ship `+0xAF`，並滿足較嚴格的 `defense-10` 條件，才回傳 2。

兩個直接 caller 均在戰術 AI `Do_Auto_Ship_Turn_ @ 0x29837`。其下游
`Resolve_Ai_Boarding_ @ 0x2BF73` 對 type 1 只轉移陸戰隊／crew，對 **type 2** 則直接呼叫
`Resolve_Capture_ @ 0x37DA8`，故 type 2 已閉合為「執行俘獲解算」。它仍不是無條件成功：
`Resolve_Capture_` 會跑雙方地面戰式傷亡與剩餘兵力，再決定是否呼叫 `Capture_Ship_`。

### `Capture_Ship_ @ 0x38312`

艦艇換 owner、重導飛彈並移除持續效果後，`0x38411..0x38446` 檢查新 owner：

- 新 owner 是特殊 owner `>=8`，或不是 Telepathic：將 combat ship `+0xB0` 寫為 1。
- 新 owner `<8` 且是 Telepathic：跳過該寫入。

直接 caller 是 `Resolve_Capture_ @ 0x37DA8` 與 `Crystal_Control_ @ 0x29790`。2026-08-30 的
全庫 direct-operand census 找到 38 個具 0x139 stride 的戰鬥艦 `+0xB0` 讀寫站點：非零時被
`Combat_Ship_Is_Functional_`、戰術 AI、目標估值、飛彈／光束強度、撤退、艦隊技能、改鎖、
回合行動與勝負判定排除；`Init_Ship_For_Start_Of_Turn_` 對值 1 清移動量與武器狀態。
因此其正式資料契約可寫成「俘獲後戰鬥失能／排除旗標」；Telepathic 新 owner 跳過旗標，代表
擄獲艦不受這層當場失能限制。這不擴張成戰鬥外永久狀態。

### `Ai_Self_Destruct_Check_ @ 0x28168`

若局部攻守差 `delta > 0`，`0x28572..0x28585` 對 Telepathic owner 使用 `10×delta`，否則使用
`5×delta`，再累加入 AI 自毀檢定分數。唯一 caller 是 `Do_Auto_Ship_Turn_`。這與前述登艦鏈
相連，但不代表所有 Telepathic 艦艇都會自毀；最終仍受總分、`Random(100)` 與 combat ship
另一 raw gate 限制。

## 三、殖民地心靈控制

### 共同合法性 `Player_Can_Mind_Control_Colony_ @ 0xC622A`

回傳 true 必須同時滿足：

1. 進攻玩家 `+0x8B8 != 0`。
2. `Player_Has_Ship_Size_Or_Larger_At_Star_(attacker, 2, star)` 成立；raw size 門檻是 2。
3. 對殖民地 owner 呼叫 `Player_Has_Leader_With_General_Skill_At_Star_(owner, 0, star)` 的兩次
   原版檢查皆為 false。兩次呼叫的暫存器實參相同；完整 helper `0xC6052..0xC60E8` 只掃 67 筆
   領袖記錄並呼叫唯讀 bit-test `Officer_Has_General_Skill_ @ 0x9467D`，沒有寫端或 RNG。
   因此重複呼叫沒有玩家可見副作用；remake 可合併為一次，而原版 raw 重複仍保留於證據。
4. 殖民地 owner `+0x8B8 == 0`；Telepathic 帝國不能被此能力心控。

直接 caller 覆蓋：

- `Add_Colony_Combat_Fields_`：殖民地戰鬥 popup 欄位。
- `Evaluate_Colony_Combat_Input_`：玩家操作判定與失敗訊息。
- `Do_Attacker_Beat_Colony_Stuff_`：戰略戰鬥／回合解算。

`Player_Can_Mind_Control_Colony_With_Help_ @ 0xC613B` 使用相同條件，但按第一個失敗原因顯示
raw string 0x233／0x24C／0x200／0x201，再回傳 false；成功回傳 true。這證明它不是另一套規則，
而是帶玩家提示的合法性 wrapper。

### 戰鬥與 AI 接線

- `Add_Colony_Combat_Fields_ @ 0xCADC8`：僅當進攻方 Telepathic、模式參數不是 2、殖民地 raw
  `+0x15 > 0` 時，才計算並顯示心控欄位；否則把欄位座標寫成 `-1000` 隱藏。
- `Do_Attacker_Beat_Colony_Stuff_ @ 0xE87D2`：正常 `Player_Owns_Transports_` 成立會開啟殖民地
  接管；沒有 transports 時，攻方 Telepathic 且守方非 Telepathic 也會設同一接管旗標。函式內
  另有兩個 `Player_Can_Mind_Control_Colony_` caller，故 UI 合法性與戰略解算共享同一判定。
- `Get_Best_Colony_Target_ @ 0xE78A7`：AI 候選中「攻方 Telepathic、目標 owner 非 Telepathic」
  會設局部布林值，讓沒有一般殖民條件的候選仍可通過後續 gate。
- `Enemy_Colony_Worth_To_Player_ @ 0xD8D11..0xD8DE0`：依 raw 關係類別先建立兩個權重
  `(attackerWeight,targetWeight)`：類別 4 為 `(2,4)`、類別 6 為 `(0,6)`、其餘為 `(1,5)`；
  Telepathic 攻方對非 Telepathic 目標再變成 `(attackerWeight+1,targetWeight-1)`。完整回傳為
  `(targetOwnerRawValue×targetWeight + attackerRawValue×attackerWeight) / 6`，採 signed 整數除法。
  兩個 raw value 皆來自 `off_183554 + player*0x2D0 + planetIndex*2`；表的玩家可見欄名仍未知，
  但最終權重與除數已閉合。

## 四、征服後人口 ownership

`Change_Pop_Ownership_ @ 0xECBF7` 逐 packed population 回寫 owner／prisoner：異族人口通常設
`0x0400` prisoner bit；新 owner `+0x8B8 != 0` 時在 `0xECD89..0xECD96` 跳過設 prisoner。
唯一 caller 是 `Change_Colony_Ownership_ @ 0xECF41`。這也修正了
`rebellions-audit-20260828.md` 仍把 `+0x8B8` 名稱列為未知的過期敘述。

## 閉合判定

### 已證實

- 間諜基礎 +10、外交檢定 +25。
- 心控需要 Telepathic、同星系 raw size >=2 艦艇、無對應將領阻擋、目標非 Telepathic。
- 沒有 transports 時，合法 Telepathic 攻方仍可走殖民地接管路徑。
- 征服人口不設 prisoner bit。
- 戰術 AI 的登艦 action type、自毀分數與俘獲後 raw flag 有 Telepathic 分支。
- AI 選殖民地與敵殖民地估值直接消費 Telepathic 攻守組合。

### 2026-08-30 下游閉合

- raw action type 2 已證實 dispatch 到俘獲解算，而非無條件成功。
- combat ship `+0xB0` 的戰鬥失能／排除 consumer 已完成全庫 direct-operand census。
- 敵殖民地最終 worth 的雙權重、Telepathic 位移與 `/6` 已閉合。
- 科技特例是 raw 5 `TECH_ALIEN_MANAGEMENT_CENTER`，估值覆寫為 1。
- 重複 general-skill helper 鏈為唯讀，沒有 RNG 或寫入副作用。

本輪證據另存 `evidence/telepathic-downstreams-ida-20260830.json`。Telepathic 的上述 raw 下游
不再阻塞客製種族 parity；剩餘客製種族工作須以其他 trait 各自的活表缺口判定。
