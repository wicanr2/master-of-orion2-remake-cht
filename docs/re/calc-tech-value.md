# `Calc_Tech_Value_`(`sub_FC845` @ 0xFC845)規格文件

> **現況入口（2026-08-25）**：本檔是歷史逐段抄錄，含已被 IDA 資料庫推翻的舊判讀；
> 現行證據、推論分級與 remake 邊界一律以
> [`starting-tech-application-audit-20260825.md`](starting-tech-application-audit-20260825.md) 為準。
> 已閉合的人類／AI 開局應用估值與單次抽選規格見
> [`../spec/starting-tech-application-selection.md`](../spec/starting-tech-application-selection.md)。

> ⚠ **2026-08-08(第 54 項(三個寫入端))訂正:第 5 節與第 6.2/6.3 節的三個「沒有查到寫入端」是錯的。**
> 寫入端都在,只是當初沒找到。已確立的部分:
>
> | 欄位 | 當初記的 | 實際上 |
> |---|---|---|
> | `[player+0x89F]` | 「未定名,種族特性相關(猜)」 | **政體**(`sub_E4204` 對四項政府科技各寫一個立即數:邦聯 1 / 帝國 3 / 聯邦 5 / 銀河統一 7,與 remake 的 `AssimilationGovernment` 編號逐項相同)。而 `lea esi,[eax+89Fh]` / `inc byte ptr [edx+eax+89Fh]` 顯示 0x89F 是**陣列起點**(到 0x8BE),政體只是第 0 格 |
> | `[player+0x28]` | 「0..6 意義不確定」 | `Init_NPC_Personalities_Objectives_Themes_` @ 0x589D6 的**第一次加權抽選**(6 個候選)。人類玩家(值 100)有明文豁免:`cmp byte [eax+28h],64h` 不等於才寫抽選結果,等於就把 100 寫回去 |
> | `[player+0x205]` / `[player+0x206]` | 「未定名,沒有查到寫入端」 | 同一支函式的**第二、三次加權抽選**(4 個候選 / 7 個候選)。三次抽選的權重都被 `byte_199CB0`(難度)加減過 |
>
> 三個欄位在同一支函式裡依序被寫,而函式名列的正好是三件事(Personalities / Objectives / Themes)。
> **但「哪一個是性格、哪一個是目標、哪一個是主題」沒有進一步證據**——名字是二手的,
> 這裡只確定「三個各自從 6 / 4 / 7 個候選裡加權抽一個」。候選各自代表什麼仍未解。
>
> 下方舊段落只供回查當時定位，不得作現況斷言。

> 逐指令抄寫來源:`/home/anr2/moo2-private-build/re/Orion2.exe.asm` 第 365535–366520 行
> (`sub_FC845 proc near` ... `sub_FC845 endp`,共 986 行,任務書標稱「985 行」)。
> 符號對照表:`/home/anr2/moo2-private-build/re/symbols_fixed.tsv`。
> 本文件**只讀不改**,不涉及任何 `.go` 檔;版權資產(asm/exe)未搬進本 repo。

呼叫端 `Choose_Tech_Application_`(`sub_FD335` @ 0xFD335,asm 366684–366978 行)用本函式算出的
`weight` 乘上 `horizon ÷ turns` 當「加權隨機挑科技」的權重,詳見
現行 remake 已在 `starting_original_application.go` 接入人類／AI 開局估值、raw profile
與應用級單次抽選；詳見 `ai-starting-tech-profile-audit-20260825.md`。

---

## 1. 函式簽章

```
Calc_Tech_Value_(playerIndex: int16 in EAX, techIdx: int16 in EDX) -> weight: int32 in EAX
```

- **呼叫慣例**:Watcom register calling convention。兩個參數都是 **word**,呼叫端一律先
  `movsx eax, <player 相關 word>` / `movsx edx, <tech-item 相關 word>` 再 `call`。
- **第二個參數是「個別科技應用項目」的索引(0..~211),不是「研究主題/領域」索引**
  (原版的 tech-item 陣列,stride 13 bytes,約 212 筆;領域/主題陣列是另一張,stride 23
  bytes,約 83 筆)。函式內部第一件事就是用這個索引去查 tech-item 表拿到它所屬的主題。
  之所以要特別強調:呼叫端 `Choose_Tech_Application_` 的迴圈變數命名（`edi` 從 1 數到
  0xD3=211)本來就是這個 tech-item 索引,任務書描述「topicID」這個講法容易誤導,本文件
  一律稱為 **`techIdx`**。
- **回傳值在 EAX**,但函式本體用 `EDI` 累積,結尾 `mov eax, edi` 才搬進 EAX(asm 366518
  行),接著 `jmp locret_FC813` 跳進**前一個函式 `sub_FC7AC` 共用的 epilogue**
  (`leave; pop edi; pop esi; pop ecx; pop ebx; retn`,asm 365505–365510 行)——這是
  Watcom 編譯器的尾端摺疊(tail-merge)優化,不是本函式自己的 epilogue,純屬正常現象。

### 1.1 六個呼叫端(交叉驗證簽章用)

| 呼叫端函式 | 位址 | asm 行號 | EAX 來源 | EDX 來源 |
|---|---|---|---|---|
| `Get_Exchange_Tech_List_` | 0x26D19 | 33484 | `[ebp+var_18]`(player 索引) | `word ptr [ebp+var_8]` |
| `Get_Exchange_Tech_List_` | 0x26D19 | 33552 | `ecx`(player 索引) | `word ptr [ebp+var_8]` |
| `Get_Differential_Tech_List_` | 0x27094 | 33804 | `di` | `bx` |
| `NPC_Tech_Exchange_Check_` | 0x2720F | 33980 | `di` | `cx` |
| `Get_Demand_Response_` | 0x539D9 | 99583 | `edi`(= `word ptr [ebp+var_10]`) | `word ptr [ebp+var_14]` |
| `Choose_Tech_Application_` | 0xFD335 | 366751 | `word ptr [ebp+var_6C4]`(player 索引,傳入時的原始參數) | `di`(tech-item 迴圈變數,1..211) |

六個呼叫端全部一致:EAX = 玩家索引、EDX = tech-item 索引。其中四個(`Get_Exchange_Tech_List_`
`Get_Differential_Tech_List_`、`NPC_Tech_Exchange_Check_`、`Get_Demand_Response_`)都是**科技
外交/交換**相關函式(module 17/27),說明 `Calc_Tech_Value_` 是一個被多個 AI 子系統共用的
「這個科技對這個玩家值多少」估值 oracle,不是 `Choose_Tech_Application_` 專屬。

### 1.2 玩家指標推導

函式內反覆出現的樣板(整個函式出現超過 30 次):

```
movsx eax, si        ; si = playerIndex(函式一進來就 mov esi, eax 存起來)
imul  eax, 0EA9h      ; 0xEA9 = 3753,玩家結構 stride(專案既有事實,非本次新發現)
mov   edx, dword_197F98   ; dword_197F98 = _player(symbols_fixed.tsv 已命名)= 玩家陣列基底
add   eax, edx             ; eax = &player[playerIndex]
```

之後對這個指標做 `[eax+0x28]`、`[eax+0x205]`、`[eax+0x206]`、`[eax+0x89F..0x8BB]`、
`[eax+0x117+techIdx]` 等偏移存取(見第 5 節)。

---

## 2. 讀取的全域表

| 位址 | `symbols_fixed.tsv` 名稱 | 存取方式(本函式內) | 用途(從用法反推) |
|---|---|---|---|
| `0x197F98` | `_player` | `dword_197F98` | 玩家陣列基底,stride 0xEA9 |
| `0x18360C` | `_tech_research_level_values` | `dword_18360C[level*4]`,365611 行 | 依「等級」(0..22,見下)查出的**基礎權重**(dword 陣列) |
| `0x17D196` | `_tech_categories` | `byte_17D196[category*2]`,365597 行 | 每個 category 的**預設倍率**(byte,2-byte stride 的欄位 A) |
| `0x17D197` | 未命名(緊接 `_tech_categories` 後) | `byte_17D197[category*2]`,365592 行 | 存進 `var_24`,後面當「是否跳過同類已研究邊際遞減」的旗標用(見 6.4) |
| `0x17E07F` | 未命名 | `word_17E07F[techIdx*13]`,365594/366213/366395/366447/366462 行 | tech-item 記錄(stride 13 bytes)偏移 +0 的 word 欄位 = **這個 tech-item 屬於哪個研究主題**(專案既有結論,`01-gap-report.md` 已記錄) |
| `0x17E082` | 未命名 | `byte_17E082[techIdx*13]`,365591 行;`byte_17E082[edi]`(edi=另一個 tech-item 索引×13),366207 行 | tech-item 記錄偏移 +3 的 byte 欄位。本函式讀出後存進 `ebx`,**從此以後 `ebx` 在函式其餘部分代表「這個 tech-item 的 category」**,不是 techIdx 也不是主題 id。命中的值散布在 1..0x28(40),數量遠超過「6 大研究領域」,比較像原始碼裡一個細分的技術種類 enum(可能是「這是護盾/是引擎/是殖民建築/…」這種粒度) |
| `0x17D91A` | 未命名 | `byte_17D91A[topic*23]`,365596/366220/366351(經 `sub_FC7AC`)行 | 主題記錄(stride 23 bytes,`01-gap-report.md` 已確立此 stride)偏移 +14 的 byte 欄位 = **這個主題的「等級/tier」**,函式內 clamp 到上限 0x16(22)(365599–365601 行) |
| `0x199CB0` | 未命名 | `byte_199CB0`,366176 行 | 見 6.5,推測與難度或回合數有關(僅推測,未證實) |
| `0x1AB10C` | `_best_category` | `word_1AB10C`,366296 行 | 見 6.6——只有這一個位址在符號表裡有名字,其餘 7 個相鄰位址都沒有 |
| `0x1AB10E`/`0x1AB110`/`0x1AB112`/`0x1AB114`/`0x1AB118`/`0x1AB11A`/`0x1AB122` | 未命名 | 366286/306/316/326/336/354/371/402/414/428 行 | 與 `_best_category` 相鄰、存取模式完全相同(每個位址對應一個固定 category 值),但**無法確認它們是不是同一個陣列**——位址間距不規則(見 6.6) |

---

## 3. 控制流程骨架

以下是**分階段虛擬碼**,每個常數旁標行號。變數命名沿用 IDA 的 `var_XX`,並附上我推斷的
語意(標「(猜)」的是合理推測但未證實)。

### 階段 A —— 入口守衛(365553–365587 行)

```
esi = playerIndex; var_4 = techIdx
player = &_player[playerIndex]
if !Player_May_Get_Tech_App_(player, techIdx):      # sub_E412B @ 0xE412B,365564 行呼叫
    return 0                                          # 365567→loc_FC870→365571-572 行
if player[techIdx + 0x117] == 3:                      # 365578 行,tech-item 狀態==3
    return 0                                            # 365579 行 jz loc_FC870
isHuman = (player[0x28] == 0x64)                          # 365580 行,100=人類(專案既有結論)
if !isHuman and techIdx == 0x34:                            # 365585 行,techIdx==52 對 AI 特殊擋掉
    return 0                                                  # 365586 行
```

### 階段 B —— 建立基礎值(365588–365615 行)

```
category = byte_17E082[techIdx*13]                    # 365591 行,ebx = category,直到函式結束都是這個意義
var_24   = byte_17D197[category*2]                     # 365592-593 行(後面當旗標用)
topic    = word_17E07F[techIdx*13]                       # 365594 行
level    = byte_17D91A[topic*23]                          # 365596 行(主題 tier)
ecx      = byte_17D196[category*2]                          # 365597 行 ← ecx 的初始值(整段函式的「倍率」暫存器)
var_8    = min(level, 0x16)                                   # 365598-601 行,clamp 22
edi      = dword_18360C[var_8*4]                                 # 365611 行 ← edi 的初始值(整段函式的「基礎權重」)
govByte  = player[0x205]                                          # 365610 行,值域 0..3,>3 走 default
```

`govByte` 之後驅動 `jpt_FC90D` 這張 4 項跳表(365615 行 `jmp cs:jpt_FC90D[eax*4]`)。

### 階段 C —— 依 `player[0x205]` 的 category 加成(365617–365673 行,`jpt_FC90D`)

四個分支各自比對固定的 `category` 值,命中就把 `ecx` 改成硬編常數:

| `player[0x205]` | 命中的 category | 結果 `ecx` |
|---|---|---|
| 0(`loc_FC937`) | 0x1A、0x1B、0x19 | 100(0x64,365671 行經 `loc_FC961`) |
| 1(`loc_FC915`) | 0x1A | 50(0x32,365624 行) |
| 1 | 0x13、0x1E | 100(0x64,經 `loc_FC961`) |
| 1 | 0x19 | 20(0x14,365635 行) |
| 2(`loc_FC948`) | 0x15、0x1D、0x18 | 100(0x64,經 `loc_FC961`) |
| 3(`loc_FC957`) | 0x24 | 50(0x32,經 `loc_FC91A`→365624 行共用) |
| 3 | 0x26 | 100(0x64,經 `loc_FC961`) |
| 其他 / 不命中 | — | 不變(維持階段 B 的 `byte_17D196[category*2]`) |

沒有這個函式以外的證據指出 `player[0x205]` 的確切意義;數值只到 3(+default),形狀像
一個**很小的列舉**(政府別?種族?待驗證,見第 6 節)。

### 階段 D —— 依 `player[0x206]` 的 category 加成(365675–365785 行,`jpt_FC988`)

同樣的模式,`player[0x206]` 值域 0..6(+default),七個分支:

| `player[0x206]` | 命中的 category | 結果 `ecx` |
|---|---|---|
| 0(`loc_FC990`) | 0x12 | 50(0x32,365693 行) |
| 0 | 0x17 | 50(0x32,經 `loc_FC995` 共用) |
| 1(`loc_FC9AA`) | 0x0F、0x10 | 20(0x14,365710 行) |
| 1 | 0x20 | 50(0x32,經 `loc_FC995`) |
| 1 | 0x23 | 100(0x64,365727 行) |
| 2(`loc_FC9FB`) | 0x12 | 50(0x32,經 `loc_FC995`) |
| 2 | 0x14 | 100(0x64,經 `loc_FC9C5`→365727 共用) |
| 3(`loc_FCA05`) | 0x1C | 100(0x64) |
| 3 | 0x1A | 20(0x14,經 `loc_FC9AF`) |
| 4(`loc_FCA11`) | 0x25 | 100(0x64) |
| 5(`loc_FCA16`) | 0x20 | 100(0x64) |
| 5 | 0x1F | 100(0x64) |
| 6(`loc_FCA20`) | 0x21 | 100(0x64) |
| 6 | 0x22 | 100(0x64) |

### 階段 E —— 依 `player[0x28]`(人類旗標/AI 性格值)的 category 加成(365786–365900 行)

重新讀 `player[0x28]`(365735 行),值 0 分三段(`al < 1` / `al ∈{1,2}` / `al ≥3`,再細分到
6),對照固定 `category` 集合,命中設 `ecx = 100`(0x64,大多數分支)或 `ecx = 50`
(0x32,category==0x27 專屬,365816 行)。命中的 category 集合:

```
al==0:               category ∈ {0xC, 0x3, 0xB}          → ecx=100(365847)
al==1:                category ∈ {0x11, 0x21}              → ecx=100
al==2:                 category==0x27                        → ecx=50(365816)
al==2(續):              category ∈ {0x9, 0xA}                  → ecx=100
al==3(<3 段,al==2重疊需照抄原分支結構):
   實際分支結構是 al<3 再依 al==0/1/2 三選一(365785 cmp al,1 / 365828 cmp ebx,2 / …)
al∈[3,5)(al==3 或 4):  category ∈ {0x1,0xC,0x3,0xB} 或 {0x11,0x21} 視 al 而定 → ecx=100
al∈[5,7)(al==5 或 6):  category ∈ {0x0,0x4} → ecx=100
```

> ⚠ 這一段(365785–365900 行)的分支彼此共用跳轉目標(`loc_FCA47`/`loc_FCA83` 等),
> 邏輯上等價於「對 `al`(0..6)分三個大段,每段各自比對 2–3 個 category」,但共用跳轉讓
> 精確的『哪個 al 值 × 哪個 category』對應表在最後幾格(al==3 那段)不是 100% 能從單純
> 跳轉鏈反推回真值——**這裡留白**,建議之後用 IDA 的 graph view 或動態驗證覆蓋這幾格
> (見第 6 節)。表格前 5 段(al==0/1/2)的對應關係抄寫時有反覆核對,可信度較高。

### 階段 F —— 種族特性旗標 × category 巨型 switch(365900–366130 行)

`category`(ebx)本身再展開成一個大 switch,對每個特定 category 值檢查 `player` 結構
`0x89F`–`0x8BB` 範圍內的特定 byte(逐一列在第 5 節),多數是「這個 flag 非 0/正數就把
`ecx` 設成某個常數」的形狀,例如:

```
if category == 6:                          # 365893 行
    if player[0x8AC] > 0: ecx = 20             # 365909/365911 行(0x14)
    if player[0x8A0] <= 0: ecx = 5                # 365919/365922 行
if category == 0..2 這一支還會查 player[0x8A1] 兩次、player[0x8B0]/[0x8B1] …
```

（完整的 category→offset→常數對照見第 4 節常數表,這裡不重複 986 行裡每一格分支。）

此外這段還有兩個**跟 category switch 平行、直接比對原始 `techIdx`**(不是 category)的
特例(366133–366165 行):

```
if techIdx == 5:
    if player[0x8B8] != 0: ecx = 1                 # 366161/366165 行
if techIdx == 0x83(131):
    if player[0x8AA] != 0: ecx = 1                   # 366147/366165 行(共用)
    elif player[0x8A9] != 0: ecx = 50                  # 366149/366151 行
```

### 階段 G —— 人類玩家專屬加成(366170–366199 行)

```
if player[0x28] == 0x64(人類):
    base = byte_199CB0² × 25 + 50                     # 366176-179 行:eax=byte_199CB0; eax*=eax; eax*=0x19(25); ecx=eax+0x32(50)
    if category ∈ {0x1A,0x1B,0x1D,0x1E,0x21,0x15}:        # 366180-190 行
        ecx <<= 2                                           # 366195 行(×4)
```

### 階段 H —— 同類已研究「邊際遞減」(366199–366247 行)

`var_24`(階段 B 存的 `byte_17D197[category*2]`)當閘門:

```
if var_24 != 0:
    ecx *= edi                                    # 366280 行(loc_FCF5F,跳過整段邊際遞減)
else:
    # 掃過所有 tech-item(0..203,stride 13),找同 category 裡「已經拿到」(state==3)的
    # tech-item,記錄其主題等級最大值 var_14(366199-366227 行迴圈,步進 0x0D=13,
    # 上限 0xCC=204,366227-228 行)
    ecx *= edi                                      # 366230 行
    if var_14 + 3 > var_8:                             # 366232 行,+3 margin
        # 差距 <3 級
        if category ∈ {0xE,0x10,0x12,0x18,0x19,0x20,0x27}:   # 366248-266 行
            ecx = 0                                              # 366262 行:這些 category 太接近已知科技就直接歸零
        else:
            ecx = (2×ecx) ÷ (3 − (var_14+3−var_8))                # 366270-280 行(margin=var_14+3-var_8 的餘數)
    else:
        # 差距 ≥3 級,明顯領先同類已知科技
        ecx = ((var_8 − var_14) × ecx) ÷ 3                          # 366232-241 行,÷3(366238 行)
```

### 階段 I —— category 專屬「上限扣抵」(366284–366360 行)

一系列固定形狀:`若 category==X 且 edi < capTable: ecx *= (capTable − edi)`——即「這個
category 的基礎權重(`edi`)離某個上限還差多少,就再乘那個差距」:

| category | cap 來源 | 行號 |
|---|---|---|
| 0x19 | `word_1AB10E` | 366284-290 |
| 0x12 | `word_1AB10C`(=`_best_category`) | 366294-300 |
| 0x13 | `word_1AB110` | 366304-310 |
| 0x0A | `word_1AB112` | 366314-320 |
| 0x0C | `word_1AB114` | 366324-330 |
| 0x0F | `word_1AB118` | 366334-340 |
| 0x08 | 見下(特例) | 366344-360 |

`category == 8` 是特例,不查表而是呼叫 `sub_FC7AC`(找「這個玩家在指定 category 裡已完成
的最高主題等級」,見附錄)兩次,分別查 category 0xF 和 0x10,兩者相加後與
`word_1AB118 × 2` 比較,若小於門檻則 `ecx *= 2`(366344–366359 行)。

### 階段 J/K —— 依 `player[0x205]` 的第二輪加成(366361–366434 行)

```
if player[0x205] == 2 and category == 0x18:               # 366367-369 行
    if edi < word_1AB11A: ecx *= (word_1AB11A − edi)          # 366371-375 行

govByte2 = player[0x205]                                        # 366384 行,重新讀一次
if govByte2 == 0 and category == 0x1A: ecx *= (word_1AB122−edi)   # 366421-434(經 loc_FD099/loc_FD08A)
if govByte2 == 1 and category == 0x1A: ecx *= (word_1AB122−edi)     # 366400-406
if govByte2 == 1 and category == 0x13: ecx *= (word_1AB122−edi)       # 366410-414(經 loc_FD08A 共用)
if govByte2 == 2 and category == 0x15: ecx *= (word_1AB122−edi)         # 366426-434
```
(全部都有 `edi < cap` 的前提才生效,否則跳過。)

### 階段 L —— 遊戲初期加成(366435–366447 行)

```
if (_stardate − 0x88B8) < 0x96 and category == 0x12:        # 366436-442 行:0x88B8=35000, 0x96=150
    ecx += ecx                                                  # 366446 行(×2)
```

### 階段 M —— 主題完成度加成(366447–366459 行)

```
topic = word_17E07F[techIdx*13]                     # 366449-451 行(重新查一次)
if player[topic + 0xC4] == 3:                          # 366455 行:這個 tech-item 所屬的「主題」已整個完成
    ecx = ecx × 5 ÷ 4                                     # 366457-458 行:lea ecx,[ecx+ecx*4]; shr ecx,2
```

### 階段 N —— `ecx==0` 時的「補完主題」後備分支(366459–366510 行)

```
if ecx != 0:
    goto 結束                                          # 366461 行
# ecx == 0 才走這裡:
topic = word_17E07F[techIdx*13]
if player[topic + 0xC4] != 2:            # 主題目前不是「可研究」狀態
    goto 結束(ecx 維持 0)
found_pending = false
for i in 0..203:                            # 366478-497 行迴圈,步進 13,上限 0xCC=204
    if i == techIdx: continue
    if word_17E07F[i*13] != topic: continue      # 只看同主題的其他 tech-item
    if player[i + 0x117] == 3: continue             # 那個 tech-item 已經是「已取得(狀態3)」,跳過
    found_pending = true; break                       # 找到一個「同主題還沒解決的」就夠了
if not found_pending:
    ecx = edi × 10                                       # 366511 行:這是本函式唯一「拿 edi(基礎權重)
                                                             #   直接×10」的地方——若挑了這個 tech-item
                                                             #   就會讓整個主題再無未解決項目,給大加成
結束:
edi = ecx
return edi
```

### 階段 O —— 回傳(366515–366520 行)

```
eax = edi
jmp locret_FC813   # 共用 epilogue(屬於前一個函式 sub_FC7AC),leave/pop×4/retn
```

---

## 4. 常數表(完整,含行號)——remake 要照抄的東西

> 表格依函式內出現順序排列。「用途」欄盡量標明是乘進 `ecx`、當比較門檻、還是查表索引。
> `0xEA9`(玩家 struct stride)、`0x117`(tech-item 狀態偏移)、`0xC4`(主題狀態偏移)、
> `0x0D`(tech-item stride)、`0x17`(主題 stride)這幾個是**專案已確立的既有事實**
> (見 `01-gap-report.md`),表中列出是為了完整,不算本次新發現。

| 行號 | 立即數 | 用途 |
|---|---|---|
| 365560 | `0xEA9`(3753) | player index → player 指標的 stride |
| 365578 | `3` | tech-item 狀態(`+0x117`)== 3 → 直接回傳 0 |
| 365580 | `0x64`(100) | `player[0x28]==100` → 人類玩家 |
| 365585 | `0x34`(52) | 非人類玩家時,`techIdx==52` → 直接回傳 0 |
| 365590 | `0x0D`(13) | tech-item 記錄 stride |
| 365595 | `0x17`(23) | 主題記錄 stride |
| 365599/365601 | `0x16`(22) | 主題等級(`level`)clamp 上限 |
| 365612 | `3` | `player[0x205]` 跳表邊界(>3 走 default) |
| 365624 | `0x32`(50) | `ecx` 常數(govByte==1,category==0x1A;govByte==3,category==0x24) |
| 365635 | `0x14`(20) | `ecx` 常數(govByte==1,category==0x19) |
| 365641/365643 | `0x1A`/`0x1B` | category 比對(govByte==0) |
| 365645 | `0x19` | category 比對(govByte==0,經 loc_FC944) |
| 365654/365656 | `0x15`/`0x1D` | category 比對(govByte==2) |
| 365658 | `0x18` | category 比對(govByte==2,經 loc_FC944 共用) |
| 365664 | `0x24`(36) | category 比對(govByte==3) |
| 365666 | `0x26`(38) | category 比對(govByte==3) |
| 365671/365727/365816/365847等 | `0x64`(100) | `ecx` 常數(多處重複使用同一常數,見階段 C/D/E) |
| 365680 | `6` | `player[0x206]` 跳表邊界(>6 走 default) |
| 365688 | `0x12`(18) | category 比對(`player[0x206]`==0) |
| 365693 | `0x32`(50) | `ecx` 常數(`player[0x206]`==0,category 0x12/0x17) |
| 365698 | `0x17`(23) | category 比對 |
| 365705 | `0x0F`(15) | category 比對(`player[0x206]`==1) |
| 365710 | `0x14`(20) | `ecx` 常數 |
| 365715/365717 | `0x10`/`0x20` | category 比對 |
| 365719 | `0x23`(35) | category 比對 |
| 365736 | `3` | `player[0x28]` 分段門檻(al<3) |
| 365739 | `5` | `player[0x28]` 分段門檻(al<5) |
| 365747/365749 | `0x12`/`0x14` | category 比對(`player[0x206]`==2) |
| 365755/365757 | `0x1C`/`0x1A` | category 比對(`player[0x206]`==3) |
| 365764 | `0x25`(37) | category 比對(`player[0x206]`==4) |
| 365770/365772 | `0x20`/`0x1F` | category 比對(`player[0x206]`==5) |
| 365778/365780 | `0x21`/`0x22` | category 比對(`player[0x206]`==6) |
| 365785 | `1` | `player[0x28]` 分段(al<1 / al∈{1,2} / …) |
| 365794/365796/365798 | `0xC`/`3`/`0xB` | category 比對(al==0 分支) |
| 365807/365809 | `0x11`/`0x21` | category 比對 |
| 365814 | `0x27`(39) | category 比對(命中 → `ecx=50`,365816 行) |
| 365821/365823 | `9`/`0xA` | category 比對 |
| 365828 | `2` | category 比對 |
| 365833/365835/365842 | `1`/`4`/`4` | category 比對 |
| 365851/365854/365857/365860/365866 | `0xC`/`0x1B`/`0x25`/`0x28`/`0x1C` | category 比對(進入巨型 switch 的分派層) |
| 365872/365875 | `0x12`/`0x19` | category 比對 |
| 365881 | `0x10` | category 比對 |
| 365887/365890/365893 | `2`/`4`/`6` | category 比對(category 值本身很小的一支) |
| 365909 | `0x8AC` | `player` 欄位偏移(byte,與 0 比較) |
| 365911 | `0x14`(20) | `ecx` 常數 |
| 365919 | `0x8A0` | `player` 欄位偏移 |
| 365922 | `5` | `ecx` 常數 |
| 365931/365941 | `0x8A1` | `player` 欄位偏移(讀兩次,分兩段判斷) |
| 365943 | `0x0A`(10) | `ecx` 常數 |
| 365951/365953 | `0x8B1`/`0x8B0` | `player` 欄位偏移 |
| 365958 | `0x14`(20) | `ecx` 常數 |
| 365967/365979 | `0x8A2` | `player` 欄位偏移 |
| 365981 | `0x64`(100) | `ecx` 常數 |
| 365988 | `0x8B6` | `player` 欄位偏移 |
| 365998/366005 | `0x8A3` | `player` 欄位偏移 |
| 366013 | `0x64`(100) | `ecx` 常數(`loc_FCC75`/`loc_FCBCE` 共用) |
| 366022/366027 | `0x8A4` | `player` 欄位偏移 |
| 366037/366050 | `0x8A5` | `player` 欄位偏移 |
| 366041 | `0x32`(50) | `ecx` 常數 |
| 366059/366068 | `0x8A6` | `player` 欄位偏移 |
| 366078 | `0x8A7` | `player` 欄位偏移 |
| 366089 | `0x8A8` | `player` 欄位偏移 |
| 366101/366115 | `0x89F` | `player` 欄位偏移(signed byte,÷2 後與 2/3 比較) |
| 366105 | `2` | `(player[0x89F])/2 == 2` 門檻 |
| 366119 | `3` | `(player[0x89F])/2 == 3` 門檻 |
| 366129 | `0x8BB` | `player` 欄位偏移 |
| 366134/366165 | `1` | `ecx` 常數 |
| 366138 | `5` | `techIdx==5` 特例門檻 |
| 366144 | `0x83`(131) | `techIdx==131` 特例門檻 |
| 366147 | `0x8AA` | `player` 欄位偏移 |
| 366149 | `0x8A9` | `player` 欄位偏移 |
| 366151 | `0x32`(50) | `ecx` 常數 |
| 366161 | `0x8B8` | `player` 欄位偏移 |
| 366174 | `0x64`(100) | `player[0x28]==100`(人類,第二次檢查) |
| 366176 | `byte_199CB0` | 見 6.5 |
| 366178 | `0x19`(25) | 人類專屬公式乘數 |
| 366179 | `0x32`(50) | 人類專屬公式加數 |
| 366180/366182/366184/366186/366188/366190 | `0x1A`/`0x1B`/`0x1D`/`0x1E`/`0x21`/`0x15` | category 比對(命中→`ecx<<=2`) |
| 366195 | `2`(shift) | `ecx *= 4` |
| 366199 | — | `var_24 != 0` 分支閘門(見階段 H) |
| 366227 | `0x0D`(13) | 迴圈步進(tech-item stride) |
| 366228 | `0xCC`(204) | 迴圈上限(掃過的 tech-item 數) |
| 366232 | `3` | margin(`var_14+3` vs `var_8`) |
| 366238 | `3` | 除數(`(var_8-var_14)*ecx / 3`) |
| 366248/366250/366252/366254/366256/366258/366260 | `0xE`/`0x10`/`0x12`/`0x18`/`0x19`/`0x20`/`0x27` | category 比對(命中→`ecx=0`) |
| 366270(隱含) | `2`(`add ecx,ecx`) | `ecx *= 2` |
| 366284/366294/366304/366314/366324/366334/366344 | `0x19`/`0x12`/`0x0A`/`0x0C`/`0x0F`/`0x08` | category 比對(cap 扣抵表,見階段 I) |
| 366347 | `0x0F`(15) | `sub_FC7AC` 參數(category==8 特例) |
| 366350 | `0x10`(16) | `sub_FC7AC` 參數(category==8 特例) |
| 366369 | `0x18`(24) | category 比對(`player[0x205]==2`) |
| 366367 | `2` | `player[0x205]==2` 門檻 |
| 366385 | `1` | `player[0x205]` 分段門檻 |
| 366388 | `2` | `player[0x205]` 分段門檻 |
| 366400/366410/366421/366426 | `0x1A`/`0x13`/`0x1A`/`0x15` | category 比對(`word_1AB122` 系列,見階段 K) |
| 366436 | `0x88B8`(35000) | `_stardate` 基準值 |
| 366438 | `0x96`(150) | 早期遊戲門檻(150 個 stardate 單位內) |
| 366442 | `0x12`(18) | category 比對(命中→`ecx *= 2`) |
| 366455 | `3` | 主題狀態(`+0xC4`)==3(整個主題已完成) |
| 366457/366458 | `×5`(`lea ecx,[ecx+ecx*4]`)/`÷4`(`shr 2`) | `ecx = ecx*5/4` |
| 366472 | `2` | 主題狀態(`+0xC4`)==2(可研究) |
| 366493 | `3` | tech-item 狀態(`+0x117`)==3(已取得,同主題掃描用) |
| 366499 | `0x0D`(13) | 迴圈步進 |
| 366503 | `0xCC`(204) | 迴圈上限 |
| 366511 | `0x0A`(10) | 「補完主題」後備分支:`ecx = edi × 10` |

---

## 5. 玩家欄位依賴表

| 欄位偏移 | 型別 | 出現行號 | 判斷內容 | 意義 |
|---|---|---|---|---|
| `[player+0x28]` | byte | 365580, 365735, 366174 | `==0x64(100)`→人類;否則當 0..6 的小整數比較 | **已確立**:100=人類玩家標記(專案既有結論)。0..6 的用法本函式內**推測**是 AI 性格/立場的列舉值,但沒有查到寫入端或名字對照表,不確定 |
| `[player+0x205]` | byte | 365610, 366367, 366384 | 值 0..3(+default)驅動一張跳表;後面又獨立當 0/1/2/其他 用兩次 | **未定名**。同一欄位在函式裡被讀了至少 3 次、做 3 種不同用途的分支,值域小(≤4),形狀像「政府別」或某種玩家群組,但無正面證據,**不要編名字** |
| `[player+0x206]` | byte | 365679 | 值 0..6(+default)驅動另一張跳表 | 未定名,同上情況 |
| `[player+0x117+techIdx]` | byte,依 techIdx 索引 | 365578, 366216(以 `var_C` 為索引掃描), 366493 | `==1`(呼叫端已知)/`==3`(本函式:已取得/obsolete) | tech-item 逐項狀態陣列,`Choose_Tech_Application_` 已用過同一張表 |
| `[player+0xC4+topic]` | byte,依主題索引 | 366455(`==3`), 366472(`==2`) | 3=主題已全部完成;2=主題目前可研究 | 主題狀態陣列,`01-gap-report.md` 已確立(`==2`可研究、`==3`完成) |
| `[player+0x89F]` | signed byte | 366101, 366115 | `(值)/2 == 2` 或 `== 3`(有號右移 1 位,等同 ÷2 取整) | Government；欄位身分已由 31-byte producer／consumer 普查確認，本函式內各政府值的科技估值語意仍須逐分支審查 |
| `[player+0x8A0]` .. `[player+0x8AC]`、`[player+0x8B0]`、`[player+0x8B1]`、`[player+0x8B6]`、`[player+0x8B8]`、`[player+0x8A9]`、`[player+0x8AA]`、`[player+0x8BB]` | signed byte | 365909–366161 | 每個特定 category（或特定 techIdx=5/131）各自查一個固定偏移，非 0／正負決定是否覆蓋 `ecx` | 欄位身分已由 `custom-race-trait-consumer-census-20260828.md` 確認；多數 category 仍需逐分支升格。`+0x8BB` 已於 2026-08-29 另由 `omniscience-stealthy-ships-audit-20260828.md` 閉合為 category `0x25` multiplier `5 -> 1`，不可再列為未解 |

---

## 6. 誠實留白(讀不出來 / 卡住的部分)

1. **`category`(即函式重新定義後的 `ebx`,來源 `byte_17E082[techIdx*13]`)的每個數值代表
   什麼技術種類,完全不知道。** 數值範圍 1..0x28(40),數量遠多於 MOO2 手冊講的「6 大
   研究領域」,比較可能是原始碼裡一個更細的內部 enum(可能對應武器/防具/引擎/殖民建築/
   偵測…等細分類),但這只是形狀推測,**沒有字串表或符號能核對**。常數表裡逐一列出的
   `cmp ebx, 0x1A` 之類比對,remake 若要照抄,**必須先解出這個 enum 每個值對應哪個實際
   tech-item/topic**,否則就算數字抄對也不知道套用在哪些科技上。建議下一輪:對每個
   `category` 值,反查 `byte_17E082` 表裡有哪些 techIdx 落在該值,再對照 `word_17E07F`
   查出主題、對照已知的 83 個主題成本表(`OrigTopicCost`),應該能還原每個 category 大致
   對應哪一群科技。

2. **`[player+0x205]`、`[player+0x206]` 的意義不確定。** 只從用法(小整數列舉、驅動跳表)
   猜是政府別或類似的玩家分類,**沒有查到寫入端**,不排除是別的東西(例如難度、種族
   ID、外交立場)。標記為未定名是刻意的,不要在 remake 裡替它們編一個聽起來合理的名字。

3. **`[player+0x28]` 除了「100=人類」之外的數值(0..6)意義不確定。** 已知 `01-gap-report.md`
   記錄過 `sub_FD335` 尾端也用同一欄位比對 1/2/4/5 做二次過濾,兩處合起來的值域是
   `{0,1,2,3,4,5,6}` ∪ `{100}`,形狀很像「AI 性格/立場列舉 + 100 代表人類」,但沒有
   名字對照表,**不要猜是哪 7 種性格**。

4. **365786–365900 行(階段 E,`al`==3 那一段的分支)的精確對應關係沒有 100% 抄對。**
   這段用了大量共用跳轉目標(`loc_FCA47`、`loc_FCA83`),多條路徑匯流到同一組指令,
   純靠靜態讀取跳轉鏈在理論上可以還原,但反覆核對後仍有一兩處「這個 category 到底是
   在 al==3 還是 al==4 那個子分支命中」無法 100% 確定(表格已標註「精確對應關係抄寫時有
   反覆核對,可信度較高」是指前面 5 段;最後這段有疑慮但我沒有把它排除在表格外,只在此
   處誠實註記)。建議之後要真的用到這幾格常數時,用 IDA graph view 重新核對一次跳轉。

5. **`byte_199CB0` 的意義不確定。** 用法是「取值後平方,乘 25 加 50」,只在人類玩家分支
   使用,且與 `Choose_Tech_Application_`(呼叫端)尾端的另一個閘門
   (`cmp byte_199CB0,0; jbe ...`)共用同一個全域。可能是難度等級、回合數,或某種
   遊戲進度計數器——**沒有查到寫入端,不確定**。

6. **`word_1AB10C`(`_best_category`)與相鄰 7 個位址(`word_1AB10E/110/112/114/118/11A/122`)
   是不是同一個陣列不確定。** 位址間距不規則(部分是 +2,部分跳 +4 或 +8,見下),而且
   只有第一個位址在符號表裡有名字。可能的解讀:
   - 這 8 個是**各自獨立的全域變數**(對應到 8 個特定 category 各自的「上限」常數),
     `_best_category` 這個名字可能只是反映其中一個變數的原始用途(例如 AI 目前最想要的
     研究方向),不代表整組是陣列;
   - 或者這是同一張表的不同欄位,但因為某些 category 的欄位本函式沒有用到而觀察不到。
   位址間距:`10C→10E`(+2)、`10E→110`(+2)、`110→112`(+2)、`112→114`(+2)、
   `114→118`(+4,跳過 `116`)、`118→11A`(+2)、`11A→122`(+8,跳過 `11C/11E/120`)。
   **兩種解讀都沒有排除,留白。**

7. **`sub_FC7AC`(category==8 特例呼叫的輔助函式,asm 365465–365513 行)只做了摘要閱讀,
   沒有像主函式一樣逐行核對。** 已確認它的作用是「掃過玩家已取得(狀態==3)的 tech-item
   (0..0x4A=74 筆,注意這個掃描上限跟主函式的 0xCC=204 不一樣),找出符合指定 category
   的裡面最高的主題等級」,回傳這個等級。函式內部常數(`0x4B`=75 迴圈上限)已經抄在
   6.6 節之外沒有進常數表,若之後要照抄請回頭補。

8. **本文件沒有逐一驗證每個 `jz`/`jnz`/`jbe`/`jge` 的旗標語意是否跟我轉譯的虛擬碼完全
   等價**——尤其是有符號比較(`jl`/`jle`/`jg`/`jge`)出現在 `player+0x8Ax` 那段,我假設
   這些是有符號 byte(因為 `movzx`/`movsx` 混用不一致,有些是 `cmp byte ptr [...], 0`
   後直接用 `jl`/`jge`,這隱含編譯器把它當有號數處理)。這個假設合理但沒有另外用測資
   驗證。

---

## 7. 給 remake 的建議(依我的判斷)

若要把這個 985 行的估值器完整照抄,成本非常高且第 6 節列出的好幾個「未定名欄位」
和「category enum 對照」都得先解出來,否則抄的是數字、套用位置是錯的。

**最小可行的中間方案**(不必等第 6 節全部解完):

1. 先解決 6.1(category enum 對應哪些主題/科技)——這是所有其他常數能不能用得上的前提。
   方法:對 83 個主題、212 個 tech-item,把 `byte_17E082` 讀出來的 category 值列成
   「category → 主題清單」對照表,人工核對幾個容易辨認的主題(例如已知的旗艦科技、
   終極科技)屬於哪個 category,反推 category 的大致語意分組。
2. 在還沒解出 category 語意前,**可以先照抄「等級→基礎權重」這條線**(階段 B,
   `dword_18360C` 查表 + `byte_17D196` 預設倍率),完全不涉及 `player+0x205/0x206/0x28`
   這些不確定欄位,只需要 `_tech_categories`(`byte_17D196`/`byte_17D197`)和
   `_tech_research_level_values`(`dword_18360C`)兩張表——這兩張表已經有名字、讀取方式
   單純(單一陣列索引,無跳表分支),風險遠低於其他階段。這樣至少能讓 `weight` 隨主題
   等級與 category 預設倍率變化,而不是一律當 1,且不需要動用任何「猜出來的」玩家欄位
   語意。
3. 階段 M(主題完成度 ×5/4)與階段 N(補完主題 ×10)這兩條「加成」邏輯**完全不依賴
   不確定欄位**(只用 `+0xC4`/`+0x117` 這兩張已確立的狀態表),可以連同第 2 點一起照抄,
   風險同樣低。
4. 階段 C/D/E/F/G/H/I/J/K/L(`player+0x205`/`0x206`/`0x28`/`0x89F..0x8BB`/`byte_199CB0`/
   `word_1AB1xx` 系列)全部**建議先不抄**,等 6.1–6.6 的留白有進一步證據再回來補,
   誤套的風險(欄位語意猜錯)遠高於維持現狀(`weight=1` 或只用第 2 點的簡化版)。
