# Omniscience／Stealthy Ships 原版消費端稽核（2026-08-28）

## 範圍與勘誤

本切片審查 `player+0x8BA`（Omniscience）與 `player+0x8BB`（Stealthy Ships）。完整 raw
指令、函式、caller、外部符號與輸入雜湊位於
`evidence/custom-race-trait-consumers-ida-20260828.json`。

掃描期間發現 `Remove_Non_Detected_Ships_ @ 0x5D953` 的 IDA 函式邊界包含位於
`0xF4D5E` 的 distant tail chunk；該 chunk 直接跳回 `0x5DC77`，不是按線性位址歸屬
`Modem_Setup_Screen_ @ 0xF4AC8` 的一般控制流。這證明兩件事：

1. 不能只用「最近的前置符號」決定 Watcom 最佳化後的非連續函式歸屬。
2. `0xF4D5E` 的 `+0x8BB` 命中必須另追 tail chunk 的 predecessor 與暫存器來源，不能只靠
   位移名稱升格。後續專用匯出已閉合這條來源鏈，結果見下文。

因此 31-byte 普查的 137 個數字是 **candidate owner 位址數**，不是 137 個已人工閉合公式。
本文件同步修正先前過度簡化的說法。

## Omniscience：已證實的共同判定

### `Player_Is_Omniscient_ @ 0x79E06`

函式只做一個 OR：

```text
One_Leader_With_Galactic_Lore_(player) != 0
OR player[playerIndex].trait[Omniscience] != 0
```

15 個直接 caller 覆蓋：艦隊移動、蟲洞連線、星系名稱、軍官畫面、行星造訪、派最近艦艇、
摘要掃描、行星資訊、系統可視性、存活玩家清單及種族文字。因此「全知」與「具有 Galactic
Lore 的領袖」在這些 UI／查詢路徑共用相同可視契約，但不代表所有底層資料 producer 都共用。

### 星系 owner 與系統資訊

`Star_Owner_ @ 0x7A1A8` 以
`omniscient = GalacticLore || trait` 呼叫 `Colony_Owners_(star, ..., !omniscient)`：一般玩家要求
可見性過濾，全知／Galactic Lore 則停用該過濾。四個 caller 是熱鍵處理、主星圖、回合間艦艇
圖示及星系 owner 更新。

`Print_Fltscrn_Scanned_Star_Name_ @ 0x761A5` 與 `Print_Scanned_Ship_Data_ @ 0x7670E` 的顯示 gate
均接受下列任一條件：局部強制顯示、Omniscience、Galactic Lore、或
`Contact_With_One_Colony_`。兩者唯一 caller 都在 Fleet Screen，因此這是掃描 UI 的顯示契約，
不是直接改寫探索狀態。

### 艦隊 stack 與探索事件

`Find_Ship_Stacks_ @ 0x5D41E` 建立星圖艦隊 stack 後，只有目前玩家 `+0x8BA == 0` 才呼叫
`Remove_Non_Detected_Ships_`。這條路徑直接讀 trait，沒有透過 `Player_Is_Omniscient_`，所以
Galactic Lore 與 Omniscience 在「艦隊 stack 過濾」並非已證實完全等價。

`Explored_New_Star_ @ 0xFD95A` 只由 `Display_Report_Aux_ @ 0xFE63E` 呼叫；`eax` 是玩家，
`dl` 是「實際顯示／消費」旗標。專用證據
`evidence/omniscience-discovery-report-ida-20260829.json` 已把兩種候選分開：

1. `star+0x34` 的玩家 bit：`Make_Ship_Arrive_At_Star_ @ 0xFFDDA` 只在該玩家原本沒有
   `star+0x33` 造訪 bit 時，同時設定 `+0x33` 與 `+0x34`。因此 `+0x34` 是首次抵達後的
   **待顯示新星系報告遮罩**，不是殖民權或底層造訪狀態。
2. `star+0x28 >= 50` 且 `%10 == player`：`Do_System_Discoveries_At_Star_ @ 0xE9927`
   從有合格星系資料的玩家中抽出 recipient，將一次性 special 編成
   `50／60／70／80／90 + player`。`Draw_Special_ @ 0xFD875` 再解碼為 raw special
   `2／3／7／8／10`，分別對應 enum 表的 Space Debris、Pirate Cache、Splinter Colony、
   Lost Hero、Ancient Artifacts。這是「發現類型＋受領帝國」，不是未知候選 bit。
   `Draw_Special_` 對 raw 10 另有公開顯示例外：即使目前玩家不是 encoded recipient，只要該星
   另有普通 `+0x34` report bit，仍可顯示 Ancient Artifacts；這不改變獎勵 recipient。

`Explored_New_Star_` 從星系陣列尾端往前找第一筆符合 `+0x34` bit 或一次性發現 recipient 的
record。若玩家有 Omniscience 且該筆只有普通 `+0x34` bit，函式清掉 bit 後繼續找下一筆，
不開 popup；若是一次性發現 recipient，即使有 Omniscience 仍呼叫 `Draw_Special_` 與
`New_System_Discovery_Popup_`，清掉 `+0x34` bit，並把 `star+0x28` 清零。這避免全知能力吞掉
Space Debris 等實際獎勵。

`Reset_One_Shot_System_Special_ @ 0xE98F2` 本身又是三個 Watcom chunks：入口
`0xE98F2..0xE98FB` 跳到 `0xF4F74..0xF4F8D` 判定 encoded range，再回
`0xE98FC..0xE9927` 清除對應 planet special／star special。此處再次證實不可用線性最近符號
判斷 owner；完整 chunks 與 bytes 已收進同一 evidence JSON。

因此 Omniscience 在這條鏈的精確效果是「自動消費普通首次抵達報告」，不是預先改寫全部
`star+0x33` 探索狀態，也不抑制一次性星系發現與獎勵。

### 隱藏熱鍵不是正常 trait producer

`Check_Hot_Keys_ @ 0x82809` 的兩組 raw key code 會反轉目前玩家 `+0x8BA`，並同步寫
`player+0xE69`、`player+0xE77` 及顯示訊息。這是玩家可觸發但未在一般 UI 公開的 cheat／debug
路徑；它可用來做原版 oracle，但不應被 remake 的種族建立流程或正常規則當作 producer。

## Stealthy Ships：已證實的三條主鏈

### AI 可見性資料 `Compute_AI_Data_ @ 0xD3D34`

在逐艦／逐觀察者建立 AI 可見性欄位時，`0xD427E..0xD42A9` 的分支形狀為：

```text
if owner < 8:
    if observer < 10:
        auxiliary_visibility[...] = 1
    if owner.StealthyShips == 0 and prior_visibility[...] == 0:
        if Ship_Has_Stealth_Device_(ship) == 0:
            prior_visibility[...] = 1
```

因此種族 Stealthy Ships 與逐艦 `Ship_Has_Stealth_Device_` 會阻止「無匿蹤時直接標為可見」的
同一 fallback。這不等於永久不可偵測；函式內其他 scanner、接觸與位置 producer 仍可能先把
`prior_visibility` 設為非零。

### 自動設計 `Add_Specials_To_Design_ @ 0x5FE14`

自動設計逐特殊裝置考慮 raw special **31 = STEALTH_FIELD（匿蹤力場）** 時，若玩家
`+0x8BB != 0` 就跳過該裝置；其他特殊裝置不受此 gate。其餘合法性、科技狀態、微型化空間與
剩餘容量仍照常檢查。12 個 caller 覆蓋 `Auto_Design_Ship_` 的多個 role 與 theme helper。

這表示原版把種族匿蹤視為足以讓 AI 不再浪費空間裝 Stealth Field；它不證明所有戰鬥中的
Stealth Field 數值效果都等同種族 trait。

### 艦隊過濾 `Remove_Non_Detected_Ships_ @ 0x5D953`

`Find_Ship_Stacks_` 證實非 Omniscience 玩家會進此過濾器。專用 IDA 匯出
`evidence/stealthy-distant-tail-ida-20260828.json` 閉合四個非連續 chunk：主體
`0x5D953..0x5DD84`、`0xF4D5C..0xF4D7E`、`0xF4D83..0xF4D8A` 與
`0x169357..0x16937A`。關鍵 predecessor／來源鏈為：

```text
0x5DC5D  eax = sign_extend(ship[0x63])       ; ship owner
0x5DC61  eax *= 0xEA9                        ; player record stride
0x5DC68  edx = &_player
0x5DC6D  jump 0xF4D5C
0xF4D5C  ebx = [edx]                         ; _player base
0xF4D5E  compare byte [eax+ebx+0x8BB], 1
0xF4D66  equal -> 0x5DC77
```

因此 `+0x8BB` 是 ship owner 的 Stealthy Ships，證據等級由強推論升為 **已證實**；它不是
`Modem_Setup_Screen_` 線性範圍內的偶然命中。IDA owner、外部符號與 raw 位址仍全部保留，
沒有破壞性修正函式邊界。

#### 精確偵測比較

函式先建立目前觀察者的 raw scanner level：基礎 `2`；三個玩家科技狀態 byte 依序在
`+0x1CC／+0x18B／+0x1B7 == 3` 時覆寫為 `4／6／9`。這三個 raw 科技位移尚未在本切片以
producer 表確認正式科技名稱，因此只保存位移與值，不套用二手 enum 名稱。

對每艘候選敵艦，函式計算最大偵測餘裕 `detectorMargin`：

- 初值 `-300`。
- 每座目前玩家殖民地，以 `ceil(Range(star, ship)/30)` 算 parsec 距離；同點距離直接把
  margin 設為 `1000`。否則使用 scanner level，raw colony `+0x15E` 非零時另加 `2`，
  `+0x13E` 非零再加 `4`，`+0x15F` 非零再加 `6`，最後減距離並取最大值。三個 colony flag
  的正式名稱留待各自 producer 表閉合，這裡不靠相鄰欄位猜名。
- 每艘 owner 為目前玩家且 raw status `<3` 的艦艇，同樣以 scanner level 減 parsec 距離；
  `Ship_Has_Special_(ship, 5)` 為真時再加 `2`。raw special 5 已由 enum 表交叉確認為
  Battle Scanner。

候選敵艦的 `concealment` 為：raw status `8` 時固定 `-5`；status `<8` 時，一般基值為 `0`
（status `1` 另取 `-ship[0x10]`），若 owner 的 Stealthy Ships **恰等於 1** 則基值改為 `500`。
之後逐一測試設計 bitfield，raw special 31 Stealth Field、6 Cloaking Device、23 Phasing Cloak
每項再加 `500`，可互相疊加，也可與種族 trait 疊加。

最終條件由 `0x5DD1E cmp bx,di` 與 `0x5DD21 jle` 證實：

```text
concealment <= detectorMargin  -> 保留在可見 ship stack
concealment >  detectorMargin  -> Delete_Ship_Node_，從該觀察者的 stack 移除
```

這是一條星圖艦隊可見性鏈，不是格子戰術命中率或不可選取目標的證據。

## AI profile 與科技估值

專用證據 `evidence/stealthy-profile-tech-ida-20260829.json` 已閉合先前兩個 raw 下游。

### 開局 raw profile

`Init_NPC_Personalities_Objectives_Themes_ @ 0x589D6` 的第一組六候選權重依 stack 順序為
`var_18／var_14／var_10／var_C／var_8／var_4`；抽選後寫入 `player+0x28`。`0x58D53`
檢查 Stealthy Ships **恰等於 1**，命中就在 `0x58D5D` 對 `var_14` 加 `100`，因此精確語意是：

```text
Stealthy Ships -> raw6 profile candidate 1 weight += 100
```

六個候選的正式 Personality／Objective／Theme 名稱仍沒有一手對照；但 accumulator、候選索引、
六項抽選與 `player+0x28` 回寫已閉合，不再把最終欄位列為未知。

### `Calc_Tech_Value_ @ 0xFC845`

`0xFCA9E..0xFCAA3` 證實只有 tech-item category **`0x25`** 進入 `0xFCDCA`；Stealthy Ships
非零時，`0xFCDE4` 把 category multiplier `ecx` 覆寫為 `1`。原始 category 表
`0x17D196 + category*2` 對 `0x25` 的預設 multiplier 是 `5`、flag 是 `0`。tech-item 表
`0x17E07F`（212×13 bytes）的 category `0x25` 共有四筆：

| tech ID | 受版控 enum | topic raw |
|---:|---|---:|
| 38 | `TECH_CLOAKING_DEVICE` | 26 |
| 53 | `TECH_DISPLACEMENT_DEVICE` | 71 |
| 126 | `TECH_PHASING_CLOAK` | 68 |
| 172 | `TECH_STEALTH_FIELD` | 27 |

因此效果是把非人類 AI 對整個匿蹤／位移 category 的基礎倍率由 `5` 降成 `1`，不是把單一
Stealth Field 的最終科技價值固定為 1。後續人類分支 `player+0x28 == 100` 會以
`difficulty²×25+50` 重建 multiplier，故此 trait 覆寫不會存活到人類估值結果；共用的 tier、
邊際遞減、其他帝國上限與研究回合等後段仍照常套用。

## 閉合判定

### Omniscience 已閉合

- 與 Galactic Lore 共用的 15-caller 查詢契約。
- Fleet Screen 星名／艦艇資料顯示 gate。
- `Star_Owner_` 停用殖民 owner 可見性過濾。
- 星圖艦隊 stack 跳過 non-detected filter。
- `star+0x34` 首次抵達報告 producer、五類 `star+0x28` 一次性發現編碼，以及全知略過普通
  popup、保留一次性獎勵的完整 consumer。
- 隱藏熱鍵直接反轉 trait 的 oracle 路徑。

### Stealthy Ships 已閉合

- AI 可見性 fallback 把種族 trait 與逐艦 stealth device 視為同一阻擋條件。
- AI 自動設計不再加入 raw special 31 Stealth Field。
- 星圖艦隊過濾的 owner trait 來源、掃描距離、殖民地／艦艇 detector margin、四種可疊加
  concealment 來源及最終保留／刪除比較。
- raw6 profile 候選 1 權重 `+100`，以及 AI category `0x25` 四科技 multiplier `5 -> 1`。
- 快速結算與格子戰術的傳遞邊界：`Qload_Ships_ @ 0x416CF` 逐一掃描 39 個特殊裝置，
  但 raw 6／23／31 在 `0x17EF0C + id*0x2F` 的效果類型與緊鄰 signed value 都是 `0／0`，
  不向 33-byte 快速結算記錄提供通用數值加成。`Load_Combat_Ship_ @ 0x4954A` 則在
  `0x49634..0x4963E` 把設計記錄 `+0x17／+0x76` 的五個 bitfield bytes 複製到 313-byte
  格子戰術記錄 `+0x4C／+0xB2`。這些 raw bytes、完整函式與 table records 保存於
  `evidence/stealth-battle-boundary-ida-20260829.json`。

### 仍未知

- raw 6 Cloaking Device、raw 23 Phasing Cloak 與 raw 31 Stealth Field 在格子戰術中，經複製後
  各自被哪些間接 consumer 讀取，以及精確命中、目標合法性與回合時序公式。這是特殊裝置本身的
  後續窄切片，不是 Stealthy Ships trait 的直接 consumer。

全庫 direct-operand census 只有五個 `player+0x8BB` 站點：開局 profile、星圖過濾、AI 自動設計、
AI 可見性與科技估值；`Ship_Has_Stealth_Device_ @ 0x5D3DB` 的唯一 caller 也只是
`Compute_AI_Data_ @ 0xD3D34`。快速／戰略與格子戰術載入器沒有 trait direct site，也沒有把 trait
轉成上述 bitfield 的 producer。故「trait 與三種裝置在戰鬥中數值等價」已被現有資料流否定；
證據等級為 **強推論**，其限制是尚未排除未被 direct census 捕捉的任意指標間接讀取。不得把
AI 自動設計跳過 Stealth Field 或星圖 concealment 同值 `+500` 升格成戰鬥等價證據。
