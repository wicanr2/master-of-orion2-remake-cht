# 間諜回合、訓練、維護、AI 配置與 RACES 任務控制稽核

## 證據邊界

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址皆為 IDA linear、DOS/4GW LE object #1。
- 可重生匯出：[`spy-turn-policy-ida-20260828.json`](evidence/spy-turn-policy-ida-20260828.json)，
  腳本：[`audit_spy_turn_policy.py`](../../tools/ida/audit_spy_turn_policy.py)。匯出保留原始函式名、
  bytes、指令、交叉參照與反編譯導覽文字；外部符號名不作事實層。

## 已證實

### 一套 packed pool，同時表示 Agent 與外派 Spy

`player + 0xE57 + target` 每個有序帝國配對占一 byte：低 6 bit 是 0..63 人數，高 2 bit
是 0..3 任務碼。`Get／Set_Their_Spy_Number_ @ 0x1026CF／0x10278D` 保留任務位元，
`Get／Set_Their_Spy_Mission_ @ 0x1026F1／0x1027B5` 保留人數位元。自己指向自己時是全帝國共用的
防守 Agent；自己指向他國時是對該國的外派 Spy。這不是兩套獨立資源。

`Add_Their_Spy_ @ 0x10294B` 將指定 pair 加一並在 63 截止；raw `0x102982..0x102993`
是目前玩家外派包裝器。`Apply_Production_ @ 0xE36DF` 的 `0xE3B57..0xE3BD4` 證實產品碼
`-7` 完成時以 owner→owner 呼叫它，所以新訓練單位先進防守 Agent pool。滿 63 時函式回傳 0，
殖民地退回 100 工業並重新排入產品 `-7`；成功時另把帝國統計加一。由退回值可確認原版訓練成本
是 100 工業，不是直接支付 BC。

### 維護、破產與召回

`Compute_Player_Maintenance_ @ 0xE2000` 逐 target 呼叫 `Get_Their_Spy_Number_`，把包含 self Agent
在內的所有人數直接累加到間諜維護欄；因此每名每回合 1 BC。`Kill_Spy_ @ 0xEDCBF` 在帝國資金
不足時依 target 順序削減該玩家的 packed 人數，直到彌補赤字或無人可裁。
`Bring_Home_Spies_Vs_ @ 0x1019F0` 在失去接觸或帝國淘汰時，把 attacker→target 全數移回
attacker→attacker，再清空 target pair。

### 回合解算與三個任務碼

`Resolve_Spies_ @ 0x10192B` 固定先處理 Assassins，再為活動帝國呼叫
`Compute_Spy_Bonuses_`，亂序後替非真人帝國執行 `Allocate_AI_Spies_`，再亂序逐守方執行
`Resolve_Player_Spies_ @ 0x1014A4`。

任務碼的 gameplay consumer 已直接閉合：

- `0`：Espionage；淨分數至少 80 才呼叫 `Steal_App_`，至少 90 可嫁禍。
- `1`：Sabotage；淨分數至少 70 才呼叫 `Destroy_Random_Building_`，至少 90 可嫁禍。
- `2`：Hide；不執行偷竊或破壞，第二次 Spy-vs-Spy 對抗為攻方加 20。

第二次對抗淨分數至少 80 會累計一名防守 Agent 損失；至多 -80 會刪除一名外派 Spy。
兩邊 d100 的原版極值分支仍按 raw 指令保留，不能改寫成單一浮點成功率後宣稱逐值一致。

### AI 配置與 RACES 控制

`Allocate_AI_Spies_ @ 0x100D19` 每回合先收回自己所有外派人數，再決定 Agent 保留量：負攻擊評分
時全部留守；難度 0..2 參考真人與 AI 對手的最高防守數，難度 3..4 只參考真人；上限 63。
剩餘人數只分配給已接觸、仍存續且有可偷科技的帝國，權重會讀可偷科技價值、目標 Agent、
其他帝國已派人數及原版攻防分數。目標 personality raw 3 時有 `Random(8)==1` 的 1/8 機會指派
任務 `1`（Sabotage），其餘指派 `0`（Espionage）；原版 AI 不會選 `2`（Hide）。

`Init_Race_Display_Data_ @ 0x10BA3D` 把原版 mission 加一存入 RACES 畫面列的 local `+0x2C`，
所以 UI 值 1／2／3 對應 Espionage／Sabotage／Hide。`Adjust_Spy_Mission_Data_` raw
`0x10CC23..0x10CC4C` 把仍為 0 的列補成 3，即預設 Hide。拖曳函式
`0x10CCC5／0x10CD65` 在 Agent 與各國列之間搬移同一 pool、每列上限 63；
`Update_Spy_Stuff_ @ 0x10C88D` 再一次寫回 packed 人數與任務。

## 強推論與未知

- `player+0x27 == 3` 的正式 personality 名稱尚未由本切片的資料表／介面文案證實；數值、1/8
  分支與任務碼已證實，名稱不得由外部符號猜測。
- RACES 三顆按鈕的正式英文／繁中文字串與圖像順序仍應由 JIMTEXT／RACES 資產或正常畫面驗證；
  gameplay 值 0／1／2 已閉合，不阻塞玩法 RE。
- 原版 PRNG 的逐位元序列及訊息文案不在這份靜態切片內。

## 對 remake 的判定

目前 remake 的每名 1 BC 維護、63 上限與三任務基本效果有原版依據；「30 BC 直接訓練」、
「AI 每六回合免費增加一名」、依 remake personality 自選三任務，以及把新單位直接放入指定敵國，
都不是原版模型。依 RE-first gate，本輪只登記差異，不修改 Go／Ebitengine 行為；待整份玩家玩法
RE 知識庫閉合後再建立 READY spec。
