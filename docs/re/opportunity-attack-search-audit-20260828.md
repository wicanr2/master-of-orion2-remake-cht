# AI 機會攻擊多艦隊搜尋稽核（2026-08-28）

## 證據邊界

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址為 IDA linear、DOS/4GW LE object #1。
- 原始函式名、指令、bytes、caller／callee 與外部符號導覽名見
  [`evidence/opportunity-attack-search-ida-20260828.json`](evidence/opportunity-attack-search-ida-20260828.json)。
  外部符號名只協助導覽；以下結論均以原始指令資料流審查。

## 已證實

1. `Set_Opportunity_Attacks_ @ 0x1FD80..0x1FEDD` 位於世界回合外交決策之前，逐 AI source 與
   真人 target 更新三個方向欄。source 必須存活、非真人，target 必須存活且為真人；雙方不同、
   無 hyperspace flux，且 source→target 的 `+0x584` 與 `+0x595` 均為 1。gate 不成立時直接寫
   `+0x837=-1`、`+0x847=0`、`+0x857=0`。
2. `Find_Opportunity_Attack_ @ 0xDBC5C` 要求航路表存在；source 有 Trans-Dimensional
   `+0x8BC` 時可跨過 hyperspace flux gate。`Find_Opportunity_Attack_Aux_ @ 0xDBB9F` 另要求
   source 起點的 route head 不為 -1 且難度大於 1，再以 source 的 `+0x59E` 星系進入搜尋。
3. `Gather_Player_Ships_At_Star_ @ 0xD93F8` 從每帝國 146-byte route table head 走全域 ship
   linked list，收集 owner 與星系相符的 129-byte ship records，以 -1 結尾。非 source
   `+0x59E` 星系時排除 `ship+0x7F == 5`；raw 5 的正式狀態名稱仍未知。
4. `Find_Best_Attack_From_Star_ @ 0xD94B3` 只有正常 AI 攻擊發射與本機會攻擊評估兩類 caller。
   它逐殖民地掃描，無強制 target 時才呼叫 `AI_May_Attack_Player_ @ 0xD7669`；強制真人 target
   時只保留該 owner。候選必須是尚未掃過的星系、source 已知該星系、艦隊可移動，而且估計
   航程小於 10 回合。
5. `Compute_Attack_Worthiness_ @ 0xD8ED2` 逐艦累加對該 owner 的太空攻擊值、對行星護盾階的
   地面／轟炸值與艦體權重，並加入依 source、星系、owner、航程索引的預估敵艦強度，以及
   防守殖民地在抵達前最多七個建造槽可完成的艦艇強度。攻擊值全零，或
   `4×預估敵艦強度 > 3×太空攻擊值` 時，候選為零並把該星系從目前航程至第九回合標成剪枝。
6. 通過前置 gate 後，評分還消費 `Enemy_Colony_Worth_To_Player_ @ 0xD8D11`、
   `Colony_Space_Strength_Vs_Player_ @ 0x5F804`、`Colony_Ground_Strength_Vs_Player_ @ 0x5F747`
   與 `N_Enemy_Stars_Reachable_ @ 0xD8E52`。正常攻擊模式會把低於最大值三分之二的權重清零，
   再對最多 250 座殖民地做加權抽選、移動艦艇並回寫 raw 任務型別；機會攻擊模式不移動艦艇。
7. 機會攻擊模式傳入五個 word 的 context：target player、固定門檻 100、候選 planet ID、
   enemy-colony worth、攻防壓力。只有 raw type 1 且壓力嚴格大於 100 時，才保留新的 planet ID；
   壓力的原始運算順序為
   `groundAttack × (100 × spaceAttack / (predictedShips + colonySpaceDefense)) / colonyGroundDefense`，
   並沿用原版的 1／3 正規化與大數右移防溢位。`sub_DBB9F` 最後分別輸出 planet ID、worth、
   signed-expanded pressure。因此三個持久欄精確語意是 `+0x837` 候選 planet ID、`+0x847`
   colony worth、`+0x857` 攻防壓力；舊稱「星系、分數、壓力」不精確。
8. `Enemy_Colony_Worth_To_Player_ @ 0xD8D11` 依方向正式狀態選兩組係數，另以雙方 raw
   `+0x8B8` 調整，最後混合 owner 與 source 對同一 planet 的 720-byte value table 並除以 6。
   `Player_Is_Hostile_To_Player_ @ 0xD8DE1` 依 raw policy 4..6、source `+0x27` 與 `+0x28`
   回傳 hostility；函式名不能取代這些原始分支。

## 強推論與未知

- `ship+0x7F == 5`、attack type 1／2／6、star `+0x6F` 與 context 中第一個 target word 的正式
  UI 名稱尚未由文字資產閉合；其數值、分支與玩家可見 consumer 已證實。
- `N_Enemy_Stars_Reachable_` 的第二個 helper `sub_1276F0` 正式語意仍未知，但其可達星計數只作
  正常攻擊評分乘數；機會攻擊 context 直接使用攻防壓力，不依賴最後的平方根乘數。
- 本切片閉合原版搜尋與三欄 producer，不代表 remake 的單主力艦隊 adapter 已與原版同構；
  依 RE-first gate，本輪不修改 Go／Ebitengine 行為。
