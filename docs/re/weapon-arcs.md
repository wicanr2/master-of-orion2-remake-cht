# 戰術武器射界與艦首朝向證據

本筆記只記錄目前能直接回查的原版方向消費端，以及重製採用的資料映射；沒有把
未驗證的圖像幀名稱當成方向事實。

## 輸入與定位契約

| 輸入 | SHA-256 | 工具／位址基準 |
|---|---|---|
| `Orion2.exe.asm` | `76cac6231a60da0fdba705907a88a853a1d345ed7bb7c15788b280fdbb259a18` | IDA Pro 9.4 匯出；`cseg01` 線性位址（`sub_XXXX`） |
| `orion2_all.c` | `c2b5c30701019c0cc58763eb29c2abddb55eb551e1e7e52f68070d629e694505` | Hex-Rays／IDA 9.4 反編譯；函式標頭的原始 `0x` 位址 |
| `func_names.txt` | `7d37a88c59fd3f31d0436fd5386b41b90fac5144346cffd66a0231f87cc4f04e` | 原始符號表；同一 `cseg01` 線性位址 |

反組譯名稱、相對位址與欄位偏移都保留原樣；下文的「已證實」指的是程式碼分支或
呼叫端已直接看到，不代表重製的畫面座標與原版資產幀已完全相同。

## 已證實的方向鏈

1. `Relative_Bearing @ 0x32AD1` 將角度轉成四位元遮罩：
   - `angle <= 60 || angle >= 300` 設 bit `0x1`；
   - `angle <= 120 || angle >= 240` 設 bit `0x2`；
   - `60 <= angle <= 300` 設 bit `0x4`；
   - `120 <= angle <= 240` 設 bit `0x8`。

   邊界是重疊的，例如 0 度得到 `0x3`、180 度得到 `0xC`，不能改寫成四個
   互斥象限。

2. `Relative_Bearing_XY @ 0x32A20` 先以 `angle_to_sin @ 0x1384B9` 計算目標相對
   攻擊艦的角度，再減去
   `22 * combat_record[0x23] + combat_record[0x23] / 2`，最後呼叫上述
   `Relative_Bearing`。因此 `combat_record +0x23` 是 0..15 的 16 向 heading
   消費欄位；這是「強推論」（欄位用途由讀寫與旋轉流程共同支持），不是只從欄位
   名稱猜測。

3. `Move_Ship @ 0x3F5F1` 將 `angle_to_sin` 的結果反向後，以
   `(4 * angle + 45) / 90 % 16` 量化為 16 向朝向。重製採同樣的 0..15 編碼：
   `0=右、4=上、8=左、12=下`。

4. `Ship_Can_Deploy_At @ 0x49043` 的初始部署分支將一方的 `v28` 設為 `0`、
   另一方的 `v26` 設為 `8`，並在每艘戰鬥艦記錄的 `+0x23` 寫入 `result[35]`。
   這確認了左右兩側面向彼此的 0／8 初始映射；具體哪個網路／玩家身份走哪個
   分支由原版對戰身份決定，重製固定以玩家左側 `Facing=0`、敵方右側
   `Facing=8` 對應目前格子戰場。

5. 多個原版攻擊端（例如 `Range_To_Ship @ 0x289C4`、`Evaluate_Possible_Move
   @ 0x2A810`、`Defensive_Fire_Check @ 0x37173`）都使用
   `((weapon_arc & Relative_Bearing(...)) != 0)` 形狀的檢查。`Ship_Can_Deploy_At`
   之外，`Design_Weapon_Firing_Arc @ 0x77DFA` 直接讀取每個設計武器的 arc 欄位。

## 重製接線

- `internal/gamedata/weapon_arc.go` 實作原始方向量化與重疊遮罩，保留 raw
  `ARC_FWD=1`、`ARC_FWD_EXT=2`、`ARC_BACK_EXT=4`、`ARC_BACK=8`、
  `ARC_MONSTER_360=15`、`ARC_360=16`。
- `internal/shell/combat_weapon_arc.go` 將 `CombatShip.Col/Row/Facing` 與
  `CombatShip.WeaponArc` 接到格子戰術合法射擊判定。`ARC_360` 與 `15` 全向；
  無效且沒有武器名稱的舊測試資料保留相容行為，避免歷史 fixture 因缺少新欄位
  被誤判成不能開火。
- `StartCombat` 以玩家 `Facing=0`、敵方 `Facing=8` 建立兩側艦隊；戰術格點移動
  依移動向量更新 16 向 heading。玩家與敵方還擊都走同一個射界判定。
- 快速艦隊結算不套用射界：原版快速路徑本身沒有格位、朝向或移動狀態，也沒有
  在下列攻擊鏈讀取 arc／bearing。重製保留同樣的抽象戰鬥，不把固定 `range=2`
  的 `battleVolley` 假造為格子戰術等價。

## 快速結算的反向證據

目前檢查的原版快速戰鬥鏈如下：

- `QGet_Target @ 0x41F20` 只依候選艦的剩餘結構值選目標；它不讀位置或 heading。
- `QGet_Target_SC @ 0x41F80`、`Strat_Special_Attack @ 0x4221F`、
  `Missile_Attack @ 0x420C0`、`Strat_Bomb_Attack @ 0x4213B` 與
  `Draw_Qcombat @ 0x4215C` 的傷害落點都進 `Strategic_Combat @ 0x40C2A`，
  後者只累加抽象 damage，沒有 `Relative_Bearing` 或 `Get_Weapon_Arc`。
- 殖民地快速結算 `Resolve_Strat_Colony_Damage @ 0x4257E` 直接呼叫上述快速攻擊
  函式；已檢查的該段 call graph 沒有再繞入戰術方向函式。

這是「強推論」而非對所有間接跳轉的數學證明：證據來自固定輸入的 IDA
反編譯 call graph 與 `orion2_all.c` 中 `Relative_Bearing`／`Get_Weapon_Arc` 的
直接呼叫點；目前沒有看到快速路徑的間接 arc consumer。因此快速結算的射界項目
結案為「原版抽象路徑不消費」，不新增未經原版支持的空間模型；格子戰術仍由前述
已證實方向鏈消費射界。

## 證據分級

| 項目 | 等級 | 理由 |
|---|---|---|
| 四位元 bearing 遮罩的條件與重疊邊界 | 已證實 | `Relative_Bearing @ 0x32AD1` 函式本體 |
| arc 與 bearing 以 bitwise AND 判斷 | 已證實 | 多個攻擊／防禦呼叫端的條件 |
| `+0x23` 是 16 向 heading 的數值消費方式 | 強推論 | `Relative_Bearing_XY`、`Move_Ship`、`Get_Facing` 的讀寫鏈 |
| 初始左右朝向為 0／8 | 已證實 | `Ship_Can_Deploy_At @ 0x49043` 的兩個部署分支 |
| 原版快速結算不消費射界 | 強推論 | `QGet_Target @ 0x41F20`、`QGet_Target_SC @ 0x41F80`、`Strategic_Combat @ 0x40C2A`、`Missile_Attack @ 0x420C0`、`Strat_Special_Attack @ 0x4221F` 的固定 call graph；未見 arc／bearing 讀取 |
| 目前重製的 CMBTSHP sprite 幀與 0..15 的每一幀視覺方向一一對應 | 未知 | 尚未用原版畫面／資產 oracle 完成幀號對照 |
