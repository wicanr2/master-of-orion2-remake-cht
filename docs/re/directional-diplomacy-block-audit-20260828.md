# 外交方向資料區 `player+0x60F..+0x88F` 稽核（2026-08-28）

## 問題與範圍

原版 AI 外交先前雖已分別追回關係、條約、事件記憶、宣戰、停戰、需求與對真人目標，仍缺一份
能防止欄位互相誤套的 ordered-pair 資料契約。本頁以玩家可見外交 producer／consumer 為範圍，
不把 compiler helper、C runtime、檔案函式或 Windows／DOS 平台內部列入完成分母。

`tools/ida/audit_directional_diplomacy_block.py` 掃描 `player+0x584..+0x88F` 的直接 displacement，
用來找候選站點；直接 operand 掃描本身不能看見先取址後間接存取，也可能抓到同函式其他 record
的相同位移。因此下表只把已回查原始指令、初始化器與至少一個 producer／consumer 的語意列為
「已證實」；只有初始化／清除端的欄位維持未知。

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4、Hex-Rays 導覽稿與 IDAPython；正式判定以原始指令／bytes 為準。
- 位址基準：IDA linear，DOS/4GW LE object #1。
- 機器可回查證據：
  [`directional-diplomacy-block-ida-20260828.json`](evidence/directional-diplomacy-block-ida-20260828.json)。
- 原始函式名與位址不作破壞性改名；下列中文語意只是附加索引。

## 資料形狀

玩家 record stride 是 `0xEA9`（3753）。表中 `byte[8]／word[8]／dword[8]` 都以
`observer record + target index` 表示方向資料；標成「每帝國純量」者不可誤作 target array。

### 關係、條約與回合記憶

| raw offset | 形狀／初值 | 附加語意與主要證據 | 等級 |
|---|---|---|---|
| `+0x60F` | `byte[8]=0` | 曾與 target 建立接觸；`Compute_Contacts_ @ 0xEB192` 的首次／重新接觸分支及 UI consumer | 已證實 |
| `+0x617` | signed `byte[8]`，種族配對初值 | 目前方向關係；`Diplomacy_Growth_ @ 0x4DD6B`、`Change_Relations_ @ 0x4E3B5` | 已證實 |
| `+0x61F` | signed `byte[8]`，同一種族配對初值 | 關係自然漂移目標；`0x4E1D7..0x4E25D` | 已證實 |
| `+0x627` | signed `byte[8]=0` | 正式外交 policy：0 無、1 互不侵犯、2 同盟、3 暫時停戰、4..6 戰爭類狀態 | 已證實 |
| `+0x62F` | `byte[8]=0` | 貿易協議 active／階段；建立 `0x101EE3`、逐回合 `0x101E77`、解約 `0x519AC` | 已證實 |
| `+0x637` | `byte[8]=0` | 研究協議 active／階段；建立 `0x101F82`、逐回合 `0x101E77`、解約 `0x5175B` | 已證實 |
| `+0x63F` | signed `word[8]=0` | 方向納貢 raw mode／值；`Start_Tribute_Treaty_ @ 0x52049` 與 `Break_Tribute_ @ 0x51C02` | 已證實 |
| `+0x68F` | signed `word[8]=0` | 條約提案持久 modifier；每回合向零回復，談判／回應後下降 | 已證實 |
| `+0x69F` | signed `word[8]=0` | 貿易／研究協議提案持久 modifier；每回合向零回復 | 已證實 |
| `+0x6AF` | signed `word[8]=0` | 四槽外交 modifier 的第三槽；科技交換、評分、宣戰及回復 consumer 已閉合，產品正式名稱未知 | 已證實行為；名稱未知 |
| `+0x6BF` | signed `word[8]=0` | 四槽外交 modifier 的第四槽；條約提案、封鎖、評分、宣戰及回復 consumer 已閉合，產品正式名稱未知 | 已證實行為；名稱未知 |
| `+0x6D7` | signed `byte[8]=0` | 方向 reputation；宣戰／解約／外交回應 writer，談判／議會／提案 consumer | 已證實 |
| `+0x6EF` | signed `word[8]=0` | 方向戰爭損失累積；戰鬥 writer `sub_4B184`、戰損比較 `0x51F0F／0x51FE5`，宣戰／停戰清零 | 已證實 |
| `+0x6FF` | signed `word[8]=0` | 殖民破壞／易手怨值；四條戰略／地面 writer、reason 22 與投降評分 consumer | 已證實 |
| `+0x70F` | `byte[8]=0` | 初始化、宣戰與停戰只見清零；尚無已證實非零 writer 或玩家 consumer | 未知，不阻塞目前玩法 RE |
| `+0x717` | unsigned `byte[8]=0` | 戰爭持續回合；戰時雙向加一、封頂 250，宣戰／停戰歸零 | 已證實 |
| `+0x71F` | unsigned `byte[8]=0` | 重複外交事件記憶；由 pending 事件演化、可鏡射，影響談判與需求 | 已證實 |
| `+0x727` | `byte[8]=0` | 單向永久違約旗標；actor 毀約寫在 victim→actor，初始化之外無 clearer | 已證實 |
| `+0x72F` | unsigned `byte[8]=0` | 方向外交／停戰冷卻；停戰設 30 並逐回合遞減，首次接觸與締約亦依政府／旗標調整 | 已證實 |
| `+0x737` | signed `word[8]=0` | 上一個剛清除的 pending magnitude 影子；阻止同回合自然漂移抵銷事件 | 已證實 |
| `+0x747` | `byte[8]` | 最近被中止的正式／經濟協議 raw 類型；正式 policy 原值、貿易 7、研究 8 | 已證實 |
| `+0x74F` | `byte[8]=0` | 上一回合正式 policy 影子；`End_Of_Turn_Diplomacy_Adjustments_` 每回合從 `+0x627` 複製 | 已證實 |
| `+0x7EC` | 每帝國 signed word，非 array | 食物赤字連續回合；`+0xB0<0` 增加、否則歸零，特殊宣戰與 AI→真人評分消費 | 已證實；不屬方向矩陣 |
| `+0x7EE` | signed `byte[8]=0` | 方向違約積怨；`Break_Treaties_` 寫入，議會、提案與 AI→真人目標評分消費 | 已證實 |
| `+0x7F6` | signed `byte[8]=-1` | `+0x7EE` 對應的受害帝國 slot；AI→真人 reason 177／178 分流 | 已證實 |
| `+0x7FE` | `byte[8]` | `Break_Treaties_` 保存第三方看向 actor 的當時 policy；尚未找到獨立直接 consumer | 已證實 writer；下游語意未知 |
| `+0x806` | unsigned `byte[8]` | 玩家需求 mode 5 成功後的方向禁止計時；每回合遞減，AI 間諜配置在非零時跳過 target | 已證實 |
| `+0x80E` | unsigned `byte[8]` | 玩家需求 mode 8 成功後的方向禁止計時；每回合遞減，貿易／研究與 AI 經濟 gate 消費 | 已證實 |
| `+0x817` | unsigned `word[8]=0` | 距上次提案累積權重：接觸 pair 每回合增加、封頂 250；提案後雙向歸零，餽贈依價值除小 | 已證實 |
| `+0x827` | signed `word[8]`，依 AI personality 初始化 | NPC proposal response memory；接受／拒絕依性格增減並雙向鏡射，議會投票直接消費 | 已證實 |
| `+0x88F` | unsigned `byte[8]=0` | AI→真人接觸回合，接觸後加一、封頂 250；建立外交攻擊／需求後雙向歸零 | 已證實 |

### 每回合 pending 訊息與 payload

| raw offset | 形狀／初值 | 附加語意與主要證據 | 等級 |
|---|---|---|---|
| `+0x64F` | `byte[8]=0` | pending reason；`Change_Relations_` 寫、訊息選擇與事件記憶消費、回合末清除 | 已證實 |
| `+0x657` | `byte[8]=0` | 本回合選定的外交訊息 raw ID；`Determine_Diplomacy_Messages_` 等 producer、外交 UI consumer | 已證實 |
| `+0x65F` | signed `word[8]=0` | pending 最強事件幅度；絕對值較大才覆寫，回合末搬到 `+0x737` | 已證實 |
| `+0x66F` | signed `word[8]=0` | `Change_Relations_` payload B；不同 reason 的產品語意不同，文字 codec 直接格式化 | 已證實 raw 契約 |
| `+0x67F` | signed `word[8]=0` | `Change_Relations_` payload A；可承載行星／星系等 reason-specific ID | 已證實 raw 契約 |
| `+0x6CF` | `byte[8]=0` | `Determine_Bad_Message_` 保存 reason 1..9，AI→真人事件評分消費 | 已證實 |
| `+0x6DF` | `byte[8]` | 解約通知類型：正式 policy 原值、貿易 7、研究 8；訊息 codec 與提案 dispatcher 消費 | 已證實 |
| `+0x6E7` | `byte[8]` | 餽贈回應類別 1..4；`Get_Gift_Response_ @ 0x53723` 寫入，訊息 codec／dispatcher 消費 | 已證實 |
| `+0x757` | `byte[8]=0` | 初始化與 `Clear_Diplomacy_Messages_` 清除；直接 operand 集未見非零 writer／consumer | 未知，不阻塞目前玩法 RE |
| `+0x75F` | `byte[8]=0` | 餽贈／需求 payload 類別或納貢 tier；`Set_Demands_`、訊息 codec、`Npc_Give_Gift_` | 已證實 raw 契約 |
| `+0x767` | `byte[8]=0` | 科技 application payload，0 表示無；需求／餽贈 staging 與實際授予 consumer | 已證實 |
| `+0x76F` | `byte[8]=0` | 初始化與回合末清除；直接 operand 集未見非零 writer／consumer | 未知，不阻塞目前玩法 RE |
| `+0x777` | signed `word[8]=0` | BC payload；需求／餽贈 staging、文字與資源移轉 consumer | 已證實 |
| `+0x787` | signed `byte[8]=0` | 第三帝國／關聯帝國 payload；同盟要求與外交文字 consumer | 已證實 raw 契約 |
| `+0x78F` | `dword[8]` | NPC 科技交換 proposal 的 raw 32-bit payload；`NPC_Tech_Exchange_Check_ @ 0x2720F` 寫、非玩家提案畫面消費 | 已證實 raw 契約；正式型別未知 |
| `+0x7AF` | `byte[8]` | 上述科技交換的 application ID；文字 codec 與科技授予 consumer | 已證實 |
| `+0x7B7` | signed `word[8]`，`-1` 表示無 | 殖民地／行星 payload；合法時由 `Npc_Give_Gift_` 移交 | 已證實 |
| `+0x7CC` | signed `word[8]=0` | 訊息 93／125 的 BC payload shadow；由全域 staging `word_19B580` 複製 | 已證實 raw 契約 |
| `+0x7DC` | `byte[8]=0` | 訊息 93／125 的科技 application payload shadow；由 `byte_19B587` 複製 | 已證實 raw 契約 |
| `+0x7E4` | signed `byte[8]=-1` | 訊息關聯第三帝國 slot；reward、需求與解約路徑消費／清除 | 已證實 raw 契約 |

### AI 對真人任務與機會攻擊

| raw offset | 形狀／初值 | 附加語意與主要證據 | 等級 |
|---|---|---|---|
| `+0x7C7` | 每帝國 signed word `-1` | 待執行的目標 planet ID；AI 外交決策寫入，艦隊目標／偷襲 consumer 清除 | 已證實；非方向 array |
| `+0x7C9` | 每帝國 byte `0` | `+0x7C7` 的任務／reason raw code；外交畫面與 AI 艦隊 consumer | 已證實；非方向 array |
| `+0x7CA` | 每帝國 signed byte | 待執行任務的目標 owner／玩家 slot；艦隊目標鏈消費後設 `-1` | 已證實行為；初始化來源仍需補一個直接站點 |
| `+0x7CB` | 每帝國 signed byte `-1` | pending reason 9／AI 對真人請求的關聯 player slot | 已證實 raw 契約；非方向 array |
| `+0x816` | 每帝國 unsigned byte `0` | AI→真人新目標決策冷卻；每回合遞減，選定任務時寫入隨機值 | 已證實；非方向 array |
| `+0x837` | signed `word[8]=-1` | 機會攻擊候選 planet ID | 已證實 |
| `+0x847` | signed `word[8]` | 候選敵方殖民地 worth | 已證實 |
| `+0x857` | signed `dword[8]` | 候選攻防 pressure | 已證實 |
| `+0x887` | signed `byte[8]` | AI→真人評分選出的 worst reason；拒絕 106 會搬入 `+0x7C9` | 已證實 |

## 初始化與清除順序

`Init_Diplomatic_Relations_ @ 0x4D78E` 先初始化所有 ordered pair，再初始化每帝國純量。
因此 `+0x7C7／+0x7C9／+0x7CB／+0x816` 不可因位址接近矩陣而誤判成 target array。
`Clear_Diplomacy_Messages_ @ 0x5090C` 先把 `+0x65F` 搬到 `+0x737`，再清 pending reason、
訊息、餽贈／需求 payload；之後才遞減停戰冷卻並累加戰爭 duration。這個順序是
`+0x737` 能阻止同回合自然漂移抵銷事件的必要條件。

## remake 對映與未來規格閘門

目前 remake 將多個方向欄壓入對稱 `Treaty`／`AIPolicies`，無法無損表示原版的 reputation、
四槽 modifier、proposal age／response memory、break payload、兩種需求禁止計時及 AI 任務純量。
這是已證實的資料模型差異；依 RE-first gate，本頁只建立證據，不撰寫 READY spec，也不修改
Go／Ebitengine 行為。

RE 知識庫仍保留以下誠實邊界：

- `+0x70F／+0x757／+0x76F` 只有初始化／清除端，未證實玩家可見 consumer；不為了填滿表格猜名。
- `+0x7FE` 已證實 writer，但獨立下游 consumer 尚未知。
- `+0x78F` 的 32-bit 儲存與科技交換路徑已證實，正式產品型別仍未知。
- direct displacement inventory 不證明沒有間接 writer；若後續玩家路徑需要上述欄位，必須從
  取址端與 caller 重新開窄切片，不能用「全庫直接 xref 為零」宣稱不存在。

其餘表列欄位已具初始化、producer／consumer 與玩家可見用途，可作後續 DRAFT spec 的 RE
輸入；在整份 parity matrix 的玩家玩法列閉合並經使用者確認前，仍不得直接開工。
