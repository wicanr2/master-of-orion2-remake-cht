# 網路多人聊天輸入框(Chat Box)反組譯筆記

> 分析對象:`/home/anr2/moo2-private-build/re/Orion2.exe.asm`(21MB 反組譯清單,純文字,行號用
> `grep -n` 取得)+ `/home/anr2/moo2-private-build/re/symbols_fixed.tsv`(位址→符號名對照)。
> IDA 目前壞掉(`Failed to initialize IDA as library, error code 4`),本輪**沒有**嘗試修它或跑它,
> 全程手動讀 `.asm` 文字解碼暫存器。本文只轉寫反組譯出的十進位/十六進位數字與函式呼叫關係,
> 不複製任何版權資產。
>
> 入口:`Chat_Box_Input_Loop_` = `sub_F55A4`(VA `0xF55A4`)。
> 座標系比照本目錄其他篇(`screen-coords-spy-leader.md`)已驗證的慣例:**640×480**,
> Watcom 暫存器呼叫慣例 `eax`/`edx`/`ebx`/`ecx` 傳前 4 個整數參數,其餘用 `push`(右到左)。

## 0. 先講最重要的一個修正:`Chat_Box_Input_Loop_` 本身不是迴圈

任務給的入口點名字帶「Loop」,但實際反組譯內容(`sub_F55A4`,行 354656–354706,共約 50 行)
**沒有任何跳回自身開頭的分支**——它是一個直線 + 少量條件分支的「每幀執行一次」函式,
執行到底就 `retn`。真正的「迴圈」在它的 15 個呼叫端各自的等待/戰鬥同步迴圈裡,
`Chat_Box_Input_Loop_` 只是被這些迴圈**每一輪呼叫一次**的「聊天送出檢查 + 換頁」小工具函式。

這點很重要:符號名 `Chat_Box_Input_Loop_` 是 IDA/前人取的**二手推論**,內容證實它既不是
modal 迴圈,函式本體也沒有處理 Enter/ESC/逾時——那些鍵盤事件的處理在別的地方(見 §7 誠實留白)。

## 1. `Chat_Box_Input_Loop_`(`sub_F55A4`)逐行控制流程

```
sub_F55A4:                                   ; 354656
    push ebx / push edx
    call sub_1171AB (Get_Input_)             ; 354660 → edx = 目前這一幀的「作用中」欄位/玩家序號
    eax = dword_192680 (_mp_screen)          ; 354662
    if ( _mp_screen[+0x10C] == 0 ) {         ; 354663  ← 「聊天欄位尚未初始化」旗標
        _input_field_active      = 1         ; 354665  (word_1844CA)
        _cursor_count(dword)     = 0         ; 354666-354667 (dword_1844C6,一次清 4 bytes)
        _continuous_string       = 0         ; 354668  (byte_1B071C)
        _active_input_field_first= 1         ; 354669  (byte_1B3E18)
        bx = _mp_screen[+0xBB]               ; 354670  ← 「目前輪到誰輸入」玩家序號
        _global_chat_string[0]   = 0         ; 354671  (byte_1AAC54,清空聊天輸入緩衝區)
        _active_input_field_number = bx      ; 354672  (dword_1844CE 的高位 word,即 +2)
        _mp_screen[+0x10C]        = 1         ; 354673  ← 標記「已初始化」
    }
loc_F5603:                                   ; 354675
    eax = _mp_screen
    if ( dx == _mp_screen[+0xBB] || dx == _mp_screen[+0xBD] ) {   ; 354677-354680
        // dx 是 Get_Input_ 回傳、目前這幀輸入所屬的「玩家序號」;
        // 只有等於 _mp_screen 記錄的兩個玩家序號欄位之一,才處理下面的送出邏輯
        if ( _global_chat_string[0] != 0 ) {                       ; 354683
            Send_Chat_Msg_( eax = &_global_chat_string )            ; 354685-354686
            _cursor_count(dword) = 0                                 ; 354687-354688
            _continuous_string    = 0                                ; 354689
            _global_chat_string[0]= 0                                ; 354690  (清空,準備下一則)
        }
        _active_input_field_first = 1        ; 354694 (byte_1B3E18,不論有沒有送出都執行)
        _input_field_active       = 1         ; 354696 (word_1844CA)
        _active_input_field_number= _mp_screen[+0xBB]  ; 354695-354697
    }
loc_F566F:                                   ; 354699  ← 上面兩層 if 沒過也會落到這裡
    call sub_124D41   (Set_Page_Off_)
    call sub_F1075    (Draw_Net_Next_Turn_Screen_)   ; 畫「等待畫面」背景 + 聊天紀錄(見 §6)
    call sub_1077D    (_TOGGLE_PAGES_)                ; 換頁(double buffer flip)
    pop edx / pop ebx
    retn
```

**關鍵發現**:`_input_field_active`、`_cursor_count`(dword_1844C6)、`_active_input_field_number`
等全域(symbols_fixed.tsv module 140,見 §2)是**整個 UI 通用的文字輸入欄位系統**的全域狀態
(`_cursor_position` / `_cursor_on` / `_field_box_mode` / `_help_box_mode` 都在同一組連續位址,
module 140 = 通用 UI 輸入欄位子系統)——聊天欄位只是借用這組全域的其中一個欄位槽,
**不是聊天專用的獨立資料結構**。這解釋了為什麼 `Chat_Box_Input_Loop_` 本身看起來「很短」:
真正的逐字元輸入處理(含 Enter 判斷)在這組通用欄位系統別處(見 §7)。

## 2. 呼叫端(交叉引用):誰在用這個函式

`grep -n "call    sub_F55A4$"` 共抓到 **15 個呼叫點,分布在 10 個不同函式**裡
(IDA 在 `sub_F55A4` 宣告處的 xref 註解只列了前兩個就用「...」省略,完整清單靠逐一回溯每個
`call` 所在的 `proc` 得到):

| 呼叫端(符號名) | 位址 | 呼叫點行號 | 上層呼叫者 | 所屬子系統 |
|---|---|---|---|---|
| `Do_Attacker_Beat_Colony_Stuff_` | 0xE87D2 | 335150 | `Do_1_Combat_` | 地面戰鬥(module 104) |
| `Do_1_Combat_` | 0xE938C | 335872 | `Search_For_Battles_` | 戰鬥解算(104) |
| `Client_Combat_Loop_` | 0xEA2F1 | 337521 | `Search_For_Battles_` | 戰鬥解算(104) |
| `Send_Game_Orders_` | 0xF7E95 | 358892 | `Client_Next_Turn_`(loc_FC325) | 網路同步(112/114) |
| `Send_Pick_Target_Info_` | 0xF88D8 | 359775 | `Do_Net_Combat_Update_` | 戰鬥網路同步(112) |
| `Send_You_Are_Target_Info_` | 0xF8F98 | 360450 | `Do_Nonhuman_Attack_Against_Human_` | 戰鬥網路同步(112) |
| `Send_Battle_Results_` | 0xF97CD | 361263 | `Client_Combat_Loop_`(loc_EA671) | 戰鬥網路同步(112) |
| `Host_Next_Turn_` | 0xFBFE2 | 364809 / 364843(2 處) | `Net_Next_Turn_` | 回合同步(114) |
| `Client_Next_Turn_` | 0xFC2D2 | 365038 / 365050 / 365079 / 365103(4 處) | `Net_Next_Turn_`(loc_FC623) | 回合同步(114) |
| `Net_Next_Turn_` | 0xFC470 | 365231 / 365257(2 處) | `Screen_Control_`(loc_10753) | 頂層畫面切換(1) |

**結論(對應任務問的「哪一張畫面」)**:

- Chat_Box_Input_Loop_ **只在多人連線遊戲開始之後**(等回合結算 / 等戰鬥結算)的等待期被呼叫,
  對應到 remake 已有的「等待畫面 `Net_Next_Turn`」——這條線是直接對得上的
  (`Net_Next_Turn_`/`Host_Next_Turn_`/`Client_Next_Turn_` 三支都在呼叫鏈上)。
- **它完全不會在「連線大廳」(`Choose_Net_Plyrs`)或「區網探索」(`Choose_Multi_Net_Game`)被呼叫**
  ——這兩個畫面的符號在整份 asm 裡沒有任何路徑呼叫到 `sub_F55A4`(用「誰呼叫了這兩張畫面自己的
  函式」反查也沒對到 `sub_F55A4` 的呼叫端清單)。換句話說:**聊天功能在原版裡不是「連線大廳就能聊」,
  是「進入遊戲、進入等待畫面才能聊」**。remake 若要在大廳畫面加聊天,那是**新設計**,不是照抄原版行為。
- 另外 7 個呼叫端(`Do_Attacker_Beat_Colony_Stuff_`、`Do_1_Combat_`、`Client_Combat_Loop_`、
  `Send_Game_Orders_`、`Send_Pick_Target_Info_`、`Send_You_Are_Target_Info_`、`Send_Battle_Results_`)
  全部屬於**多人戰鬥的網路同步流程**——這些函式本身在等對方的戰鬥指令/結果封包時,
  也順手呼叫 `Chat_Box_Input_Loop_` 檢查一次聊天,可以理解成「戰鬥結算等待畫面」也共用了同一個
  聊天小工具。remake 目前列出的畫面清單(連線大廳/等待畫面/連線狀態面板/區網探索/文字輸入彈窗)
  裡沒有明確對應「戰鬥網路同步等待」這張畫面——這是一個值得記錄的缺口,但**不在本次任務範圍**
  (任務只要求聊天框規格,不要求盤點戰鬥同步畫面)。

## 3. 輸入框的建立:`Add_Continuous_String_Input_Field_`(`sub_115BEA`)

`Net_Next_Turn_` 進畫面時由 `Add_Net_Next_Turn_Fields_`(`sub_EFCEA`,行 346922–347046)
呼叫 `Add_Continuous_String_Input_Field_`(`sub_115BEA`,行 408486–408681)建立聊天輸入欄位。
呼叫點在行 347031,參數(Watcom 暫存器慣例 + 6 個 stack 參數,`retn 18h` = 24 bytes = 6 個
dword,證實正好 6 個 stack 參數):

| 參數 | 傳遞方式 | 呼叫端行號(最後寫入值的那行) | 值 / 運算式 | callee 內用途(行號) | 存進欄位結構的 offset |
|---|---|---|---|---|---|
| `eax` | 暫存器 | 347027(`cwde`,承接 347014-347026 一串字串寬度運算) | `Get_String_Width_("(玩家名)  ")` + `var_10`(見 §7) | X 座標 | +0x00(行 408561-408562) |
| `edx` | 暫存器 | 347029 | `var_14`(見 §7,Y 相關) | Y 座標 | +0x02(行 408568-408569) |
| `ebx` | 暫存器 | 347026 | `0x23A(570) - 字串寬 - 3` | 加到 X 存成右邊界(X+寬) | +0x04(行 408589-408596) |
| `ecx` | 暫存器 | 347020 | **`0x11` = 17(十進位)** | 加到 Y 存成下邊界(Y+高) | +0x06(行 408597-408604) |
| stack arg_0(`[ebp+10h]`)| push(最後一個 push,離 retaddr 最近) | 347028 | `offset byte_1AAC54`(`_global_chat_string`) | 輸入緩衝區指標 | +0x18(行 408629-408637) |
| stack arg_4 | push | 347025 | **`0x50` = 80(十進位)** | 最大輸入字數,callee 內再鉗制到 `<=0x100(256)`(行 408605-408608) | +0x2C(行 408616-408617) |
| stack arg_8 | push | 347022 | `0` | 未能確認語意 | +0x20(word,行 408656-408657) |
| stack arg_C | push | 347019 | `0` | 未能確認語意(可能是次要 callback/緩衝區,值為 0=無) | +0x1C(dword,行 408649-408650) |
| stack arg_10 | push | 347017 | `0x29` = 41(`')'`的 ASCII 碼,但**不確定是不是被當字元用**,見下方註) | 未能確認語意 | +0x35(word,行 408629-408630) |
| stack arg_14 | push | 347016 | `0` | 未能確認語意 | +0x30(word,行 408663-408664) |
| (固定值,非參數) | — | — | `0x0B` = 11(十進位) | **欄位類型 tag**,硬編碼寫死 | +0x08(行 408670) |

**`+0x08` 欄位類型 tag 的交叉驗證**:同一個「通用欄位陣列」系統裡,`Add_Hidden_Field_`
(`sub_11438B`)在完全相同的 offset `+0x08` 寫死的是 `7`(行:`sub_11438B` 內
`mov word ptr [edx+8], 7`)。兩個不同的「Add_XXX_Field_」建構子在同一個 offset 寫入不同的常數,
證實 `+0x08` 是一個「欄位型別」列舉值,`Add_Continuous_String_Input_Field_` = 型別 11,
`Add_Hidden_Field_` = 型別 7。這是本文少數敢下「這個 offset 是什麼」結論的地方,
因為有兩個獨立呼叫端互相印證(比對照單一函式猜測硬一級,方法比照
`docs/re/screen-coords-spy-leader.md` 用的交叉驗證原則)。

**欄位陣列本身**:所有欄位(不分聊天/隱藏欄位/其他輸入框)存在同一個全域陣列
`_fields`(`off_184480`,module 140),每筆固定 **`0x37` = 55 bytes**(`imul _, 37h` 在
`sub_115BEA`/`sub_11438B` 內反覆出現),陣列上限 `_max_fields_count`(`word_18447E`)、
目前筆數 `_fields_count`(`word_1B3E0E`),超過上限會呼叫 `Exit_With_Message_` 印
`"Too many fields"`(行 408518,字串常數 `aTooManyFields`)。

### 3.1 幾何尺寸整理(可信度分級)

| 項目 | 數值 | 可信度 | 依據 |
|---|---:|---|---|
| 高度(Y2−Y1) | **17px**(`0x11`) | **高**——直接立即數,行 347020 | `ecx` 暫存器參數,無任何 LBX 資產查表介入 |
| 最大輸入字數 | **80 字**(`0x50`),callee 內鉗制到 `<=256`(`0x100`) | **高**——直接立即數,行 347025 + 408605-408608 | 兩個獨立聊天實作(本畫面 + 外交畫面,見 §8)都用同一個 80 |
| 寬度(X2−X1) | `570(0x23A) − 字串寬("(玩家名)  ") − 3` | **中**——570 是立即數(行 347018),但實際寬度**隨玩家名字長度變動**,不是固定寬 | 347014-347026 一串 `Get_String_Width_` 呼叫 |
| X 座標(左上角) | 無法化簡成單一立即數,見下方 §7 | **低**——牽涉 `Clear_Fields_` 回傳值 + `_mp_screen[+0xBF]` 執行期欄位 | 346942-346956 |
| Y 座標(左上角) | 同上 | **低** | 同上 |

⚠ 這正好對到任務事先提醒的坑:「寬高常常是 LBX 資產控制碼、不是字面尺寸」——這裡**反過來**,
高度/最大字數是**真的字面立即數**,但 X/Y 座標**才是真正需要執行期資料(玩家名字串寬度、
畫面基準點)才能算出來的值**,不能直接當成螢幕絕對座標抄。

## 4. 訊息緩衝區與歷史紀錄(`chat_info` @ `dword_1AA250`)

`dword_1AA250`(symbols_fixed.tsv 未命名,IDA 只標了 `dd ?`)是一個指向「聊天資訊」結構的指標,
在 `Net_Next_Turn_`(`sub_FC470`)進畫面時配置/寫入(行 365188:`mov dword_1AA250, eax`),
離開時清空(行 365304:`mov dword_1AA250, 0`)。結構內用到的 offset(從 `Receive_Chat_Msg_`
和 `Draw_Net_Next_Turn_Screen_` 反推):

| offset | 型別/用途 | 依據 |
|---|---|---|
| `+0x47C` | 目前已存訊息筆數(dword,類似 write index) | `Receive_Chat_Msg_` 行 315087/315093/315097/315112;`Draw_Net_Next_Turn_Screen_` 行 348974(繪圖迴圈的上限判斷) |
| `+0x52 * i`(起始) | 第 i 筆訊息的槽位起點(每槽 **82 bytes**) | `imul edx,[eax+47Ch],52h`,行 315097、315099 |
| 槽位 `+0` | 1 byte:**送話者玩家序號**(見 §5 的 GNN 保留值 8) | `Receive_Chat_Msg_` 行 315098(`mov [edx+eax], cl`);`Draw_Net_Next_Turn_Screen_` 行 348861(`movzx ax, byte ptr [eax]` 再與 8 比較) |
| 槽位 `+1` 起 | 訊息文字(null-terminated,最長 81 bytes) | `Receive_Chat_Msg_` 行 315101-315109(`lodsb`/`stosb` 迴圈到 0 為止) |

**容量上限 = 14 筆(`0x0E`)**:`Receive_Chat_Msg_` 行 315087
`cmp dword ptr [eax+47Ch], 0Eh` / `jnz` — 一旦筆數到達 14,就用 `memmove`
搬移 **`0x42A` = 1066 bytes**(行 315089-315091),`1066 = 13 × 82`,
**剛好是「丟掉最舊一筆、其餘 13 筆整批前移」的搬移量**,兩個數字互相驗證,不是巧合湊出來的。
搬移後把 write index 重設回 `0x0D`(13,行 315093),下一則訊息就寫在最後一槽——這是標準的
**FIFO 捲動式聊天記錄**,上限 14 行。

`Draw_Net_Next_Turn_Screen_` 的繪製迴圈(行 348972-348975)每幀都是「畫全部已存筆數」
(`for row < chat_info[+0x47C]`),沒有另外的「可視範圍/捲軸」——即畫面上能同時看到的行數
上限就是儲存上限 14 行,沒有更小的可視窗口。

⚠ **緩衝區宣告大小抓不到確切位元組數**:IDA 只在 `dword_1AA250` 本身標了名字(4 bytes),
之後到下一個具名符號(`dword_1AA254`)只有 4 bytes 間距——這代表 `chat_info` 指向的是**堆積
配置出來的一塊記憶體**(不是靜態陣列),真正的配置大小要追 `Net_Next_Turn_` 裡呼叫的配置函式
(本輪沒有展開),只能從「14 筆 × 82 bytes + 0x47C 之前的其他欄位」反推一個下限
`14×0x52 + 0x480 = 0x920`(約 2336 bytes)以上,**不是查到的準確值,是下限估計**,
不要當成準確結構大小使用。

## 5. 送出與接收

### 5.1 `Send_Chat_Msg_`(`sub_DD3B8`,行 315125–315181)

呼叫端只有一處:`Chat_Box_Input_Loop_` 行 354686(`eax = &_global_chat_string`)。

行為:對 `_player` 陣列(`dword_197F98`,每筆 stride `0xEA9`,筆數 `_NUM_PLAYERS`/`word_199998`)
逐一檢查,只對「狀態 byte `[+0x28] == 0x64`('d')」的玩家(連線中/在遊戲內的玩家)送出訊息——
對每個符合條件的玩家呼叫 `Mox_Send_Message_`(`sub_F6816`,行 315165)+
`Mox_Update_`(`sub_FE8BE`,行 315166)。`Mox_Send_Message_` 呼叫時 `edx = 0x27`(39,十進位),
這是**封包型別/tag**,與 `Main_Receive_Message_` 的 dispatch(見 5.3)case 39 對得上。

`Mox_` 前綴的兩支函式(`Mox_Send_Message_`/`Mox_Update_`)是**通用底層網路傳輸原語**,
不只給聊天用——本輪沒有展開它們的封包格式細節(超出「聊天框」範圍)。

### 5.2 `Send_GNN_Chat_Msg_`(`sub_DD42A`,行 315188–315248)

結構跟 `Send_Chat_Msg_`幾乎一樣,但多了「訊息截斷」:字串長度算出來後
`cmp ecx, 50h ; jle ...; mov byte ptr [buf], 0`(行 315226-315230)——
**超過 80 字元就在第 80 個字元後面截斷**,傳送時 `edx = 0x2D`(45,十進位)當封包 tag。

「GNN」在遊戲裡是 Galactic News Network(銀河新聞網),這支函式應該是系統/新聞式廣播訊息,
不是玩家手動輸入的聊天,但共用同一組「history 槽位」顯示機制(見 §6 的 GNN 分支)。

> **⚠ 主迴圈補一筆(2026-08-08,第 54 項(三個寫入端)):送出的 0x2D 在收訊端沒有對應的 case。**
> `Main_Receive_Message_` 那張 68 格跳表實際存在的 case 是 0..67 扣掉 4/45/49/51/53/54
> ——**沒有 case 45**。所以「送 0x2D」與「case 43 是 GNN 收訊路徑」這兩件事各自成立,
> 但**它們串不起來**。一般聊天沒有這個問題(送 0x27 = 39,收 case 39)。
>
> 這不足以斷言「原版的 GNN 廣播是壞的」——沒有核 `sub_F6816` 是不是唯一的傳送出口、
> 也沒有核有沒有第二個派送器。記在這裡是因為 `internal/netplay/chat.go` 先前把
> **收訊 case 號 0x2B 當成送訊型別號**記下來了,而那個錯誤就是從這個縫隙來的。

### 5.3 接收端 dispatch:`Main_Receive_Message_`(`sub_F5A9F`)

行 355895-355903(case 39,tag `0x27`):

```
cmp dword_1AA250, 0        ; chat_info 還沒配置就整個跳過(還沒進 Net_Next_Turn_ 畫面)
jz  跳過
eax = [ebp+var_134]        ; 封包帶的「送話者玩家序號」
call sub_DD351 (Receive_Chat_Msg_)   ; eax=送話者序號, edx=訊息文字指標(進入 dispatch 前已設好)
```

行 355977-355982(case 43,tag 對照到某種系統訊息):

```
cmp dword_1AA250, 0
jz  跳過
mov eax, 8                 ; ★ 強制把送話者序號設成 8,不是從封包讀
jmp loc_F649E              ; 共用同一段呼叫 Receive_Chat_Msg_ 的程式碼
```

**這證實了任務問的「封包裡有沒有帶送出者的玩家索引」**:一般聊天(case 39)**有**——
`eax` 直接來自封包解析出的欄位(`[ebp+var_134]`,dispatch 函式一開始解封包時填入,
本輪沒有往回追這個值在 `Main_Receive_Message_` 更早處是怎麼從封包位元組解出來的,
只確認它「是封包來的值,不是寫死的」)。而 GNN/系統訊息(case 43)**沒有**——
一律強制當成序號 8。

**序號 8 = 保留給系統/GNN 的「虛擬玩家」**:`Draw_Net_Next_Turn_Screen_` 繪製迴圈裡
(行 348861:`cmp ax, 8; jge loc_F149F`)——**送話者序號 >= 8 時走 GNN 顯示分支**
(格式化字串 `"( GNN )  %s"`,行 348930,不查玩家名字),**< 8 時走一般分支**
(格式化字串 `"(%s)  %s"`,行 348901,查 `_player` 陣列裡的玩家名字)。MOO2 最多 8 名玩家
(序號 0–7),序號 8 剛好是保留給系統訊息的「第 9 格」,這個推論有兩處(接收端強制值 + 繪製端
分支條件)互相印證,不是憑空猜的。

## 6. 聊天記錄的繪製(`Draw_Net_Next_Turn_Screen_`,`sub_F1075`,行 348558–349xxx)

這支函式先畫玩家名單(逐一走訪 `_NUM_PLAYERS`,行 348694-348841),再畫聊天記錄。

**聊天記錄區的座標原點**(行 348842-348848,緊接在玩家名單迴圈結束之後):

```
X_origin = var_38(= _mp_screen[+0xF3],對話框基準 X) + 0x18(24)
Y_origin = var_20(玩家名單 + 邊框圖塊畫完後累積的 Y 座標,執行期動態值) + 0x0E(14)
```

**X 方向有一個乾淨的立即數**:相對於對話框基準 X,聊天記錄一律往右偏移 **24px**(`0x18`,
行 348843)——這個是可信的固定相對位移。

**Y 方向沒有乾淨的立即數**:偏移的「基準」(`var_20`)是動態算出來的——它從
`_mp_screen[+0xF5]` 開始,一路累加「MULTIGM.LBX 邊框圖塊的實際高度」(靠
`sub_127C27` 載入 LBX 資源後讀取回傳結構 `[+6]` 欄位取得,行 348644-348672,
**不是寫死的數字**)以及玩家名單本身佔的高度(取決於 `_NUM_PLAYERS`)。
**這正好對應到任務事先提醒的坑**:這裡的「高度」欄位是從 LBX 資產動態查出來的,
只有 `+0x0E`(14px)這個「聊天記錄再往下讓一點」的相對位移是寫死立即數,
在此之上疊加的絕對 Y 沒有辦法從 exe 反組譯直接讀出一個字面常數。

**每行文字高度 = 12px(`0x0C`)**:繪製迴圈行 348967
`add [ebp+var_1C], 0Ch`——每畫完一行訊息,Y 座標往下推 12px 再畫下一行。

**GNN 分支**(送話者 >= 8,行 348909-348935):字色/字型固定用
`byte_199F34 = 0x10`(16,某個色盤索引,行 348910),格式化字串 `"( GNN )  %s"`。
**一般分支**(送話者 < 8,行 348852-348906):動態查 `Font_Colors2_`/
`Get_Current_Font_Style_` 取得該玩家的顯示色,格式化字串 `"(%s)  %s"`
(玩家名 + 訊息文字,玩家名來自 `_player[送話者序號]` 陣列)。

第一行(`var_28 == 0`,最新一筆)另外多畫一個矩形高亮框(行 348937-348952):
`sub_128AB6` 畫的矩形是 `(X_origin, Y_origin) 到 (X_origin+570, Y_origin+11)`——
**寬度又用到 `570(0x23A)` 這個常數**(與 §3 輸入框寬度公式用的是同一個數字,
互相印證「570px 是這個對話框內容區的固定內寬」),高度 `11px`(`0x0B`,跟輸入框高度
`17px` 不同,這裡只框住文字本身,不含行距)。這個高亮框只在最新一則訊息才畫,
應該是「有新訊息時的視覺提示」,但實際視覺效果本輪沒有實機驗證。

## 7. Get_Input_(`sub_1171AB`)與 X/Y 座標裡沒查完的部分

輸入框建立時的 `eax`/`edx` 暫存器參數(§3 表格裡的 X/Y),推回去源頭是
`Add_Net_Next_Turn_Fields_` 開頭(行 346942-346956):

```
eax = _mp_screen
dx  = eax[+0xF3]              ; 對話框基準 X(word)
ax  = eax[+0xF5]              ; 對話框基準 Y(word),與 dx 組成一個「點」放進 var_14
call sub_11C2F0 (Clear_Fields_)   ; 回傳值(edx)存進 var_10
edx = var_14 (基準點,重新載入)
dx += eax[+0xBF]               ; 再加上 _mp_screen[+0xBF] 這個欄位(語意未查)
...後面接一串暫存器互相搬動,最終餵給 sub_115BEA 的 eax/edx...
```

`Clear_Fields_`(`sub_11C2F0`)本輪沒有展開——它回傳值(`var_10`)被拿去加進 X 座標公式
(§3 表格 eax 那列的 `Get_String_Width_(...) + var_10`)。**卡在這裡的原因**:
`Clear_Fields_` 是通用欄位系統的另一支函式(跟 `_fields`/`_fields_count` 同一組),
展開它需要先弄懂整個「目前對話框在螢幕上的哪個位置」這個更上層的版面配置邏輯
(`_mp_screen` 結構本身,`+0xF3`/`+0xF5`/`+0xBF`/`+0xBB`/`+0xBD`/`+0x10C`/`+0x10E` 這些
offset 各自的完整語意),這已經超出「聊天框」這個題目,屬於「整個 Net_Next_Turn_ 畫面版面」
的範圍——誠實記錄:**沒有查,不是查不到而是判斷不在本次任務範圍內,方法上應該併入
`docs/re/01-gap-report.md` 或另開一篇專講 `Net_Next_Turn_` 版面的筆記**。

## 8. 相鄰/相關函式總表(`grep -i "chat\|msg\|message" symbols_fixed.tsv`)

聊天直接相關(已在本文用到):

| 符號 | 位址 | 本文涵蓋 |
|---|---|---|
| `Chat_Box_Input_Loop_` | 0xF55A4 | §1(主角) |
| `Send_Chat_Msg_` | 0xDD3B8 | §5.1 |
| `Send_GNN_Chat_Msg_` | 0xDD42A | §5.2 |
| `Receive_Chat_Msg_` | 0xDD351 | §5.3 |
| `Add_Continuous_String_Input_Field_` | 0x115BEA | §3 |
| `_global_chat_string`(byte_1AAC54) | 0x1AAC54 | §1、§3、§4 |
| `_continuous_string`(byte_1B071C) | 0x1B071C | §1 |
| `_allow_chat_mode`(byte_1AAE9F) | 0x1AAE9F | **查到符號名,但本輪沒找到任何指令引用它**——grep 全檔案只在 symbols_fixed.tsv 出現,`.asm` 裡沒有任何一行提到 `1AAE9F`。可能是死碼/未使用的符號,或 IDA 誤判的資料位置,誠實記錄為「查不到用途」。 |

外交畫面(`Diplomacy_Chat_`,0x1EF5B)有**自己獨立的一套聊天輸入**,用同一個
`Add_Continuous_String_Input_Field_` 建構子,但緩衝區是不同的全域
(`_chat_message` @ 0x19A4EC,對照 `Chat_Box_Input_Loop_` 用的 `_global_chat_string`
@ 0x1AAC54),兩者互不相干。有意思的印證:`Diplomacy_Chat_` 呼叫
`Add_Continuous_String_Input_Field_` 時,stack 參數 pattern 幾乎一模一樣
(`push 0; push 0x29; push 0; push 0; push 0x50; push &buffer`)——
**同樣用 `0x50`(80)當最大字數**,強化了「聊天訊息上限 80 字」是聊天子系統的共通設計,
不是巧合湊出的數字。外交聊天的 X 座標用的是寫死立即數 `0x2B6`(694,行內 `mov eax, 2B6h`),
但本輪沒有追完它前面是否還有進一步的字串寬度調整(跟 Net_Next_Turn 版一樣的 pattern),
**沒有把握直接拿 694 當螢幕絕對座標**,列在此處僅供之後比對用。

模組 14(外交)另一組聊天用全域:`_net_diplomacy_chat_message`(0x19A484)、
`_receive_chat_message`(0x19A550)——本輪沒有展開,推測是外交聊天的接收端緩衝,
與 `_chat_message`(送出端)配對。

`Msg`/`Message` 開頭的符號絕大多數(模組 46/55/58/65/102/104/110/125/131 等)是**遊戲內
事件訊息系統**(殖民地事件、戰報、外交提案等的文字提示框),跟「玩家對玩家聊天」是完全不同的
子系統,只是命名上都含 `msg`/`message`,本文不展開(超出「chat box」範圍,避免誤植)。

## 9. 誠實留白:查不到 / 沒有查完的東西

1. **輸入框的絕對螢幕座標(X, Y)**——查得到「相對位移」(X 基準+24px,Y 基準+14px)和
   「寬度公式」(570−字串寬−3),但兩個基準值(`_mp_screen[+0xF3]`/`[+0xF5]` 再疊加
   `Clear_Fields_` 回傳值與 `[+0xBF]`)本身需要展開 `Clear_Fields_`(`sub_11C2F0`)和
   `_mp_screen` 結構的完整語意才能算出最終數字,**本輪沒有展開,卡在需要先弄懂
   `Net_Next_Turn_` 整體版面配置這個更大的題目**。§7 已記錄卡在哪一行、哪個函式。
2. **逐字元鍵盤輸入(含 Enter 送出、Backspace、游標移動)的實作位置**——確認不在
   `Chat_Box_Input_Loop_` 本體內,推測在 `Get_Input_`(`sub_1171AB`)呼叫的
   `sub_11CEF5`(可能是取原始按鍵)與 `sub_112399`(可能是依欄位型別分派處理,
   吃欄位結構 `+0x35` 那個 word 當參數)裡面,**本輪沒有展開這兩支函式**——
   它們是整個通用文字輸入欄位系統的核心,不是聊天專用,判斷屬於更大的題目
   (「通用 UI 輸入欄位系統」)而非本次「聊天框」範圍,建議另開一篇筆記處理。
3. **ESC 取消 / 逾時**——`Chat_Box_Input_Loop_` 本體完全沒有處理這兩者的痕跡
   (沒有任何比較鍵值後跳到「清空不送出」的分支,唯一的清空路徑是送出成功之後)。
   不確定原版是否根本沒有「打到一半按 ESC 清空」這個功能,或者這個功能在 §7/§9-2
   提到的通用輸入欄位系統裡處理、影響不到 `_global_chat_string` 本身。**沒有查到反證,
   也沒有查到證據,誠實記錄為「不確定原版有沒有這個功能」**,不要在 remake 裡假設它一定有。
4. **`_allow_chat_mode`(0x1AAE9F)的用途**——符號名看起來像個開關,但全檔案 `grep`
   不到任何一條指令引用這個位址,可能是死碼或 IDA 誤標,見 §8。
5. **緩衝區 `chat_info`(`dword_1AA250`)指向的記憶體實際配置大小**——只反推出一個
   「至少要有多大」的下限估計(§4 結尾),沒有查到 `Net_Next_Turn_` 裡實際呼叫的配置
   函式與其傳入的大小參數。
6. **stack 參數 `arg_8`/`arg_C`/`arg_10`/`arg_14`(值分別是 0/0/0x29/0)存進欄位結構
   之後真正的用途**——只查到「存在哪個 offset」,沒有查到「這些 offset 之後被誰讀取、
   讀取後做什麼」,因此無法確認 `0x29`(41)是不是真的被當成字元使用,還是恰好是某種
   索引/ID。標成「查不到語意,只查到落點」。
7. **`Mox_Send_Message_`/`Mox_Update_` 的封包格式**——確認「聊天走 tag 0x27、
   GNN 走 tag 0x2D」,但封包標頭其餘欄位(長度、校驗等)沒有展開,因為那是通用網路傳輸層,
   不是聊天專屬。
8. **戰鬥網路同步流程(§2 表格裡的 7 個呼叫端)為什麼也需要聊天檢查**——只確認了「有呼叫」,
   沒有深入這 7 支函式各自的整體迴圈結構去確認呼叫 `Chat_Box_Input_Loop_` 的頻率
   (每幀一次?每個網路 tick 一次?),因為那些函式的主體是戰鬥同步邏輯,不是聊天,
   展開會嚴重超出本文範圍。

## 10. 給 remake 的重點摘要(不重複上面的表,只列可以直接拿去用的結論)

- 聊天只在「等待畫面」開放(`Net_Next_Turn_` 及其戰鬥網路同步的兄弟函式),連線大廳/
  區網探索沒有聊天——remake 若要在大廳加聊天是新設計,不是還原。
- 訊息長度上限 **80 字元**(兩個獨立實作互相印證,可信度高)。
- 歷史記錄 FIFO,**最多 14 行**,每行含 1 byte 送話者序號 + 最長 81 字元文字
  (可信度高,兩處數字互相印證)。
- 送話者序號 **8 保留給系統/GNN 訊息**,顯示成 `( GNN )  訊息內容`,不查玩家名字;
  一般玩家訊息顯示成 `(玩家名)  訊息內容`。
- 聊天記錄每行高度 **12px**,對話框內容寬度 **570px**(輸入框寬度公式與最新訊息高亮框寬度
  用的是同一個常數,互相印證)。
- 座標本身(輸入框絕對位置、聊天記錄絕對位置)**沒有乾淨的立即數**,原版是靠執行期組合
  「LBX 資產實際尺寸 + 玩家人數 + 通用欄位系統回傳值」算出來的——remake 不需要照抄這套
  動態運算,只要維持上面幾條有查到的「相對關係」(聊天記錄相對對話框基準右移 24px、
  再往下讓 14px、每行行高 12px、內容區寬 570px)即可,絕對位置可以按 remake 自己的版面
  重新設計。
