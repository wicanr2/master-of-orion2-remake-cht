# 原版帝國接觸、斷聯與首次接觸稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；外部名稱不單獨作證據。
- 工具：IDA Pro 9.4／Hex-Rays 9.4.0.260610、
  [`tools/ida/audit_contacts.py`](../../tools/ida/audit_contacts.py)；位址均為 IDA linear EA，
  DOS/4GW LE object #1。
- 可重生證據：[`evidence/contacts-ida-20260828.json`](evidence/contacts-ida-20260828.json)。
- `memset_` 只清接觸矩陣，`Bring_Home_Spies_Vs_` 只取其玩家可見召回契約；兩者的 C runtime／
  內部搬移不納入玩法分母。

## 回合位置與初始化 caller

`Compute_Contacts_ / sub_EB192 @ 0xEB192..0xEB4C1` 有兩個 caller：

- `Next_Turn_Calc_ @ 0x137A1`：在殖民地槽整理後、`Determine_First_Contacts_ @ 0x137A6`、
  外交訊息與回合尾外交調整之前重算。
- `Twiddle_Initial_Homeworlds_ @ 0xE583B`：建立開局母星時先形成初始接觸狀態。

正常回合沒有艦隊參數，也不掃艦艇 record；接觸 producer 的可達資料是殖民地、行星、星系、
帝國航程與聯盟航程 helper。艦隊抵達可改變探索或殖民地，卻不是本函式的直接接觸條件。

## `Compute_Contacts_`：每回合由殖民地航程重建

### 1. 清表並保留上一回合

函式逐帝國把八個 `player+0x584[target]` byte 複製到 64-byte local snapshot，然後：

- 清零八個 `+0x584` 接觸 byte；
- 清零八個 `+0x595` 正常航程 byte；
- 把自己的 `+0x584[self]` 與 `+0x595[self]` 設為 1；
- 清 `player+0x58C` 本回合新接觸 bitset。

因此 `+0x584` 不是只增不減的發現史，而是逐回合重建的雙向目前接觸矩陣；`+0x58D[target]`
才保留曾建立新接觸的累積 byte，另由 `sub_10248B` 在「目前已斷聯但曾接觸」清單模式消費。

### 2. 雙殖民地範圍判定

函式對每一對有效殖民地（`colony+0x02` planet index 不為 `-1`）取得 owner 及其星系，分方向呼叫：

- `Star_In_Extended_Range_Of_Player_ @ 0xFF68A`；
- 若延伸航程成立，再呼叫 `Star_In_Normal_Range_Of_Player_ @ 0xFF666`。

`FF666` 讀 source `player+0x324` 航程，並透過 `sub_FF5F8` 納入正式聯盟 raw 2；更底層星圖距離、
蟲洞及逐帝國 `star+0x33` 造訪遮罩已有獨立航程證據。本切片只記它們在接觸 producer 的用途。

對殖民地 A（owner A）與 B（owner B）：

```text
extended(A -> starB) || extended(B -> starA)
    => contact[A][B] = contact[B][A] = 1

normal(A -> starB) => normalRange[A][B] = 1
normal(B -> starA) => normalRange[B][A] = 1
```

所以目前接觸是雙向的，但 `+0x595` 正常航程能力是方向性的。只要任一方向達到延伸航程，兩帝國
便維持接觸；並不要求同星、互相都在正常航程，也沒有逐艦接觸分支。

### 3. 新接觸與變更旗標

每個 source 由 target slot 高至低比較上一回合與新 `+0x584`。第一個差異會令外部原始符號
`_something_interesting_happened @ 0x1AB124[source]` 設為 1；名稱只作導航，高層 UI 刷新語意
維持強推論。若差異是 `0 -> 1`，函式另：

- 令 `player+0x58C |= 1 << target`；
- 增加 `player+0x58D[target]` 一次。

迴圈找到第一個差異便結束該 source 的比較，因此同一回合多個新接觸仍全部寫入 `+0x584`，但
`+0x58C／+0x58D` 只記 target slot 由高至低的第一個變更。這是原始控制流結果，不可把
`+0x58C` 解讀成完整的本回合 contact delta bitset。

### 4. 斷聯 consumer

對每個 ordered pair，若 `+0x584[target]==0`，函式立即：

- 清 `player+0x62F[target]` 貿易協議；
- 清 `player+0x637[target]` 研究協議；
- 以 `(source,target)` 與 `(target,source)` 各呼叫一次
  `Bring_Home_Spies_Vs_ @ 0x1019F0`。

`Diplomacy_Broken_Contacts_ @ 0x52602` 在回合尾再次對斷聯 pair 清 `+0x65F` pending 幅度及
`+0x62F／+0x637`。因此失去殖民地航程可以終止協議、清外交 pending 並召回間諜，不只是隱藏
RACES 畫面的一列。

## `Determine_First_Contacts_ @ 0x50B57..0x50DAC`

唯一 caller 是 `Next_Turn_Calc_ @ 0x137A6`。外層只處理 `player+0x28==100` 的真人席位，再掃
其目前 `+0x584[target]==1` 的 target。指令沒有排除 `target==source`；`sub_4D78E` 又把整張
`+0x60F` 初始化為零，所以第一次重算也會走 diagonal self entry。原版 UI 是否忽略該方向訊息
尚未由 consumer 證實；跨帝國規則不依賴替這個 diagonal 行為猜高層名稱。

### 首次建立：方向 `+0x60F[target]==0`

1. 將真人 source 的 `+0x60F[target]` 設為 1。
2. 將雙方向 `+0x627` 正式 policy 設為 0。
3. 雙方向各把 `2 × word_18105C[personality]` 加到 `+0x72F[target]`；若 personality raw 4
   且該方向永久違約旗標 `+0x727==1`，改用表索引 6。`word_18105C` 的已證實有效值是
   `5／10／20／5／50／40／5`。
4. 外層真人 source 遇到 AI target 時，以 `Change_Relations_` 的特殊 sentinel
   `delta=-10000`、reason 17、其餘 payload 0 建立 pending 外交訊息；該 sentinel 不套一般
   關係算術。
5. 真人對真人且 `word_19A0E2 != 3` 時，直接把 message raw `target personality + 1` 寫入
   `targetRecord+0x657[source]`。模式 3 不寫此訊息。

### 已曾接觸後重新建立：`+0x60F[target]!=0`

只有本回合 `+0x58C` 對 target 的 bit 被選為第一個變更時才處理。真人對 AI 使用同一
`delta=-10000` sentinel，但 reason 改為 18。真人對真人且模式不為 3 時：

- 若方向關係 `+0x617 < 0` 且 policy `<4`，寫 message raw `target personality + 7`；
- 否則寫 message raw `target personality + 13`。

這些 raw message 數值與 producer 已證實；其每個資產文案索引尚未逐項命名，不以猜測文字取代。

## 玩家可見查詢 helper

- `Contact_With_Player_ @ 0x78FB8`：target 合法時直接回傳
  `player[source]+0x584[target]`，否則 0。
- `Contact_With_One_Colony_ @ 0x78F4B`：掃一個星系最多五顆行星，找到有效殖民地後，以目前
  玩家 `word_19999C` 呼叫上述 helper；任一殖民地 owner 已接觸便回 true。
- `Get_Player_Contact_Count_ @ 0x5008B`：回傳 1 加上 source 對其他帝國的目前接觸數。
- `+0x595` 另被 `Set_Opportunity_Attacks_`、外交提案與 AI 目標 producer 消費，證實它不是
  第二份相同接觸表，而是接觸成立後的方向性正常航程守門。

## 閉合與 remake 邊界

- **已證實**：兩個 caller、逐回合清表、殖民地 pair、延伸／正常航程分流、雙向目前接觸、
  方向性正常航程、第一個變更、新／舊接觸記錄、斷聯協議與間諜 consumer、首次／重新接觸的
  policy、cooldown、reason 17／18 與真人訊息 raw 分支。
- **強推論**：`_something_interesting_happened` 是需要刷新玩家報告的外部狀態；原始符號與 writer
  已知，但本文件未窮盡它的 UI consumer。
- **未知**：`+0x657` 各 raw message 的精確資產文案名稱、模式 3 的正式產品名稱，以及
  diagonal self entry 的訊息 consumer、`+0x58D` byte 溢位後是否有實際長局影響。這些不推翻
  跨帝國接觸與斷聯規則。
- **明確反證**：本 producer 不因敵方艦隊同星直接建立接觸，也不寫 `star+0x33` 探索遮罩。
  探索是航程 helper 的輸入之一，不是接觸建立後的回寫。
- **remake 差異**：`.GAM` parser 只保留 `+0x584`，跳過 `+0x58C..+0x60D`；shell 又沒有獨立
  方向接觸／正常航程矩陣，目前把可外交 AI 或同星主力艦隊視為接觸。這無法表示斷聯、協議清除、
  間諜召回、首次／重新接觸或真人對真人方向狀態。依 RE-first gate 本輪不寫 spec、不改玩法。
