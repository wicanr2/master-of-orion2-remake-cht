# 安塔蘭母星防禦艦隊——反組譯解讀

> 來源:`Orion2.exe.asm`(反組譯清單,位於 `/home/anr2/moo2-private-build/re/`,版權資產,不進本 repo)
> + `symbols_fixed.tsv`(原始符號表,提供函式/變數真名)。

## 輸入與位址基準

本檔目前採用的主輸入是 `Orion2.exe` SHA-256
`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`、
`Orion2.exe.i64` SHA-256
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；工具為 IDA Pro 9.4
`idat`／Hex-Rays `9.4.0.260610`，位址是 IDA 線性位址。`symbols_fixed.tsv` 與
`symbols_full.tsv` 在部分函式有衝突，故以 raw `sub_` 名稱、函式邊界與 writer 指令為主。
本檔不以較舊 `ORION95.EXE` 對應的 `orion2_all.c` 作目前輸入主證據。
> 本檔只放解讀結果與虛擬碼,組語行號全部對應反組譯清單,方便回頭核對。

## 結論摘要(先講重點)

安塔蘭母星防禦艦隊**不是固定艦隊**,是遊戲進行中逐步養大的:

- **保底**:任何時候開戰,**至少 1 艘 Harbinger(Titan 級)+ 1 座星際要塞**(這是寫死的下限,不看回合數)。
- **上限(養滿之後)**:**3 艘 Intruder(Large)+ 2 艘 Interdictor(Huge)+ 7 艘 Harbinger(Titan)+ 1 座星際要塞 = 12 艘戰艦 + 1 座要塞,共 13 艘**。
- **Raider(Small)、Marauder(Medium)這兩級永遠不會被這套養成機制建造**(上限表對它們兩個都是 0)。
- 上限的**艘數組成不隨難度變**;難度改變的是①**養到滿的速度**(資源累積速率 100%/150%/200%)、②**單艦的裝甲**(難度愈高單艦愈硬,愈大型的艦體級距愈明顯)、③一次性的建造成本 −10%(僅 Hard/Impossible 且已建出第一艘 Titan 後觸發)。

以下逐項給反組譯依據。remake 現在的保守預設「6 艘末日之星等級」應換成上面這組真值。

---

## 1. `Load_Antaran_Defense_Fleet_` 虛擬碼(sub_4D141 @ 0x4D141,asm 90095–90129)

```
proc Load_Antaran_Defense_Fleet_():
    // word_199182 就是 word_19917A[4](Titan 那一格),見下方「陣列版面」
    if word_19917A[4] < 1:
        word_19917A[4] = 1              ; 90099-90101:强制下限,任何時候至少 1 艘 Titan 級

    for size = 0 to 4:                   ; 90104-90123,5 種艦體尺寸(Small/Medium/Large/Huge/Titan)
        count = 0
        while count < word_19917A[size]: ; 90109-90117
            Load_Combat_Antaran_Ship_(ship_index = combat_total_ship_count)
            count += 1
            combat_total_ship_count += 1 ; word_1998C0,全域戰鬥艦艇陣列游標

    Load_Antaran_Star_Fortress_()        ; 90124,無條件外加 1 座星際要塞
```

`Load_Combat_Antaran_Ship_`(sub_4D10E @ 0x4D10E,asm 90054–90089)是純 switch,依 size 0–4 分派到:

| size | 艦名(反組譯字串) | Loader | 位址 |
|---|---|---|---|
| 0 Small | **Raider** | `Load_Small_Antaran_Combat_Ship_` | 0x55161 |
| 1 Medium | **Marauder** | `Load_Medium_Antaran_Combat_Ship_` | 0x5542C |
| 2 Large | **Intruder** | `Load_Large_Antaran_Combat_Ship_` | 0x55738 |
| 3 Huge | **Interdictor** | `Load_Huge_Antaran_Combat_Ship_` | 0x55B12 |
| 4 Titan | **Harbinger** | `Load_Titan_Antaran_Combat_Ship_` | 0x55F67 |

（`size≥5` 落 `def_4D11B` default case,switch 本身不處理——星際要塞不是這個 switch 的第 6 個 case,它是額外呼叫的獨立函式,見第 3 節。)

### `word_19917A` 陣列版面(BSS,遊戲開局全 0)

```
word_19917A   dw ?      ; [0] Small  current def ship count
              dd ?      ; [1][2] Medium, Large
              db 2 dup(?)
word_199182   dw ?      ; = word_19917A[4],即 size4=Titan
```

⚠ **訂正**:一開始以為 `word_199182` 是「第 5 格」對應 size4(Titan),这个从位址算術(`0x199182 − 0x19917A = 8 = 4×2`)是對的——`word_199182` 確實是 `word_19917A[4]`。但同一個陣列在**其他函式**(`Antaran_Defensive_Force_At_Maximum_`、`Build_Antaran_Ships_` 的除錯輸出)是**逐 byte 用滿 9 格**(size 0–8)在讀,不是只讀 5 格。也就是說 IDA 依個別函式的存取樣式各自貼了 `word_199182`/`word_19918A` 這些獨立名字,但**遊戲實際把它們當同一塊連續陣列使用**,只是 `Load_Antaran_Defense_Fleet_` 本身只走前 5 格(因為 5 種艦體 loader 就只有 5 個)。第 5–8 格(`word_199184`起、`word_19918A`)在防禦這邊的上限表恆為 0(見第 3 節),所以就算陣列邏輯上有 9 格,defensive 這頭**實際只有 size 2/3/4 會非零**。

---

## 2. `Build_Antaran_Defensive_Ships_` 虛擬碼(sub_63F9C @ 0x63F9C)

⚠ **交底前先訂正原本的假設**:原先以為 `movsd/movsd/movsw`(10 bytes)是把 `byte_181746` 範本表複製進某處。**追下去發現不是**——那三行 `movsd/movsd/movsw` 的來源暫存器是 `esi`(= `offset aDefensive`,字串 `"defensive"`,剛好 9 字元+`\0`=10 bytes),目的地是 `edi`(= `offset word_18FF78`,一塊除錯訊息暫存 buffer)。**這三行只是把字串 `"defensive\0"` 抄進除錯訊息 buffer 的開頭,跟 `byte_181746` 無關。** `byte_181746` 是透過 **暫存器傳遞**(`mov ebx, offset byte_181746` 之後直接 `call Build_Antaran_Ships_`,`ebx` 在被呼叫端當隱性參數用)交給共用引擎 `Build_Antaran_Ships_`(sub_63FF0)去讀。這是這次追蹤中唯一一處「猜錯要重講」的地方,記錄下來避免下次又踩。

```
proc Build_Antaran_Defensive_Ships_():        ; sub_63F9C,asm 124619-124646
    n_max_table   = &_n_max_antaran_def_ships  ; ebx = byte_181746,124625
    count_table   = &_antaran_current_def_ships; edx = word_19917A,124626
    log_buf       = &word_18FF78                ; edi,124627
    log_buf[0..9] = "defensive\0"                ; esi=aDefensive→movsd/movsd/movsw,124628+124633-124635
    resource_pool = &word_199178                 ; eax,defensive 專用資源池,124629
    is_offensive  = 0                            ; ecx=0,124630
    Build_Antaran_Ships_(resource_pool, count_table, n_max_table, is_offensive)  ; 124636
```

`Build_Antaran_Offensive_Ships_`(sub_63FCB @ 0x63FCB,asm 124652-124665)是鏡像版本,差異只在四個指標:
`n_max_table=&_n_max_antaran_off_ships(byte_18173D)`、`count_table=&word_19918C`、
`resource_pool=&word_199176`、`is_offensive=1`、`log_buf 開頭字串="offensive"`。兩者共用同一段
`movsd/movsd/movsw + call sub_63FF0` 程式碼(`sub_63FCB` 直接 `jmp` 進 `sub_63F9C` 內部的 `loc_63FBC`)。

### 共用引擎 `Build_Antaran_Ships_`(sub_63FF0 @ 0x63FF0,asm 124672-125000)

```
proc Build_Antaran_Ships_(resource_pool, count_table, n_max_table, is_offensive):
    elapsed = stardate − 35000 − Antaran_Delay_()      ; 124693-124699,0x88B8=35000(存檔 stardate 起始值,
                                                        ; 見 docs/tech/savegame-format.md);Antaran_Delay_
                                                        ; 依 byte_199CB5(0/1/≥2)回傳 200/100/0
    printed_header = false                              ; 124702
    if antaran_ship_cost[0] > *resource_pool:            ; 124701-124712,連最便宜的 size0 都買不起
        return                                            ; 直接放棄,不進入下面迴圈

    start = (byte_199CB5 == 2) ? 1 : 0                   ; 124713-124716,某未定名旗標可讓 size0 整個跳過
    for size = start to 8:                                ; 9 格(size 0–8),124718起,124938/124945 迴圈邊界
        // 依 size 決定「這一輪還准不准建這個尺寸」
        if size == 0:
            if No_Small_Ships_After_N_Turns_() < elapsed:  ; sub_647D7 @ 0x647D7,124721-124724
                continue                                    ; 過了門檻回合數,size0 整個停建
        elif size == 1:
            if No_Medium_Ships_After_N_Turns_() < elapsed:  ; sub_6481B @ 0x6481B,124735-124738
                continue
        // 兩個門檻函式本身都已完整解出(asm 125613-125685),數字如下:
        //   No_Small_Ships_After_N_Turns_ :  難度0-2 → 100 回合 ; 難度3(Hard) → ⌊12500/150⌋=83 ; 難度4(Impossible) → ⌊12500/200⌋=62
        //   No_Medium_Ships_After_N_Turns_:  難度0-2 → 199 回合 ; 難度3(Hard) → ⌊20000/150⌋=133 ; 難度4(Impossible) → ⌊20000/200⌋=100
        // 難度愈高,門檻回合數愈短——小型艦停建得愈早。
        // size 2..8 沒有回合門檻檢查

        if antaran_ship_cost[size] == 0: continue           ; 124747-124748,表項是 0 視為不存在
        if *resource_pool < antaran_ship_cost[size]: continue ; 124751-124754,買不起
        if n_max_table[size] <= count_table[size]: continue   ; 124755-124760,已經蓋到上限

        if not printed_header:
            printed_header = true
            log "\n(%4d.%d) Building Antaran %s ships:" % (stardate/10, stardate%10, "Defensive"/"Offensive")
            log ">> current def/off ships (by size) : " + count_table[0..8]     ; 124784/124790/124829
            if is_offensive:
                log ">> off ships deployed (by size): " + word_19919E[0..8]      ; 124864

        log "-------- building size %d, resources = %d, n_ships[%d] = %d" % (size, *resource_pool, size, count_table[size])
        log "-------- n_max_ships[%d] = %d, cost = %d" % (size, n_max_table[size], antaran_ship_cost[size])
        chosen = size
        break                                                 ; 找到第一個「合格+付得起+沒滿」的尺寸就停,esi=ebx

    if chosen == none: return                                 ; 124944,整輪都沒有可建的,放棄

    count_table[chosen] += 1                                   ; 124947,實際「建成」= 計數 +1
    // 一段只認 offensive 表的怪癖(不分呼叫端是防禦還是攻擊都寫 byte_18173D):
    if chosen <= 2 and _n_max_antaran_off_ships[chosen] == count_table[chosen]:
        _n_max_antaran_off_ships[chosen] = 0                    ; 124948-124954,size 0/1/2 一旦達到「攻擊上限值」就把
                                                                 ; 攻擊上限表歸零(即使這次是防禦端在建)——
                                                                 ; 兩個艦隊在 size 0/1/2 共用同一個「總量閥門」
    *resource_pool -= antaran_ship_cost[chosen]                  ; 124957-124960
    log "-------- after build: resources = %3d, n_ships[%d] = %2d"  ; 124968

    if chosen == 4 and not byte_19B89A and byte_199CB0 > 2:      ; 124972-124978,首次建出 Titan + 難度 Hard/Impossible
        byte_19B89A = true                                        ; 一次性旗標,只觸發一次
        for i = 0 to 8:
            antaran_ship_cost[i] = antaran_ship_cost[i] * 90 / 100  ; 124980-124992,全表成本永久 −10%
```

### 難度(`byte_199CB0`,0–4,本專案已確立;地面戰已用同一顆全域做「難度−2」偏移)如何介入

- **上限表本身不變**——`byte_199CB0` 在 `Build_Antaran_Ships_` 只出現在「第一次建出 Titan 後 −10% 成本」那個一次性觸發(需 `>2`,即 Hard/Impossible)。
- **資源累積速率**由 `Antaran_Defensive_Resource_Bonus_`(未命名函式 @ 0x63E06,asm 124412-124425)與
  `Antaran_Offensive_Resource_Bonus_`(@ 0x63E29,asm 124427-124440)決定,兩者邏輯相同:
  `難度≤2 → 100%`、`難度==3(Hard) → 150%`、`難度==4(Impossible) → 200%`。
  這兩個百分比在 `Increment_Antaran_Resource_Level_`(sub_645EC @ 0x645EC,asm 125361-125449)裡
  每 25 個 stardate 單位(= 25 回合,對照 `docs/re/calc-tech-value.md` 已確立的「`stardate−35000`=已過回合數」)結算一次:
  `防禦池(word_199178)` 與 `攻擊池(word_199176)` **各拿一份**同樣公式算出的份額,
  但只要 `Antaran_Defensive_Force_At_Maximum_()`(sub_646BD)回真(防禦艦隊已經全部蓋滿上限),
  第二份就整份轉去攻擊池——即「防禦沒蓋滿時兩邊同速养,蓋滿後資源全部倒去養攻擊艦隊」。
- **單艦裝甲**依難度分級加成,而且**艦體越大分級越細**(逐一核對反組譯,armor 呼叫是 `sub_127712`):

  | 艦體 | 難度 0-1(Simple/Easy) | 難度 2(Average) | 難度 3(Hard) | 難度 4(Impossible) |
  |---|---|---|---|---|
  | Raider(Small,0x55161) | +0 | +0 | +9 | +9(二段式,只看 `>2`) |
  | Marauder(Medium,0x5542C) | +0 | +0 | +9 | +21(三段) |
  | Intruder(Large,0x55738) | +0 | +9 | +21 | +38(四段,逐難度各自查表) |
  | Interdictor / Harbinger | 同 Intruder 的四段式程式碼形狀(`mov al, byte_199CB0` 後四路分支),**本次未逐位元組核對出確切加成值**,留白 |

  即:**艘數上限不隨難度變,但難度愈高、艦隊養滿愈快、每艘船愈硬。**

- **回合門檻(`No_Small/Medium_Ships_After_N_Turns_`)對「防禦」艦隊其實是空判斷**:因為
  `_n_max_antaran_def_ships` 對 size0/1 恆為 0,不管門檻回合數是多少、有沒有過,size0/1 都會在
  下一步的上限檢查被擋掉,結果不變。這兩個門檻函式**真正有作用的是攻擊艦隊**——
  `_n_max_antaran_off_ships` 的 size0/1 是 4/4(非零),所以「過了 100/83/62(小型)或
  199/133/100(中型)回合後不再考慮這個尺寸」對攻擊艦隊是真的會生效的規則。

---

## 3. 靜態範本表

### `_n_max_antaran_def_ships`(`byte_181746` @ 0x181746,asm 552955)——本次任務的核心表

原始位元組(小端序,10 bytes,`0x181746`–`0x18174F`):

```
db 0                    ; [181746] size0 Raider  cap=0
align 4                 ; [181747] IDA 判成對齊填補,但引擎讀 byte_181746[1] 一樣拿到這個位元組——
                        ;          由 IDA 顯示為 align(=0)這件事本身就是 IDA 只在偵測到 0 值時才會這樣標,
                        ;          所以可以放心當作「值=0」,不是遺失資料
dd 70203h               ; [181748..17B] = 03 02 07 00  → size2=0x03, size3=0x02, size4=0x07, size5=0x00
dd 0                    ; [18174C..17F] size6..9 = 0
```

Go 陣列字面值(索引 0–8 對應 Small/Medium/Large/Huge/Titan/預留×4):

```go
var nMaxAntaranDefShips = [9]byte{0, 0, 3, 2, 7, 0, 0, 0, 0}
// 索引:  Small Medium Large Huge Titan  -    -    -    -
```

`Antaran_Defensive_Force_At_Maximum_`(sub_646BD,asm 125455-125490)逐格核對這張表:
`n_max==0` 的格子(Small/Medium/預留格)永遠視為「不參與、跳過」,不會讓函式回傳「未滿」;
只有 size2/3/4 三格會真的拿目前計數跟這張表比。

### `_n_max_antaran_off_ships`(`byte_18173D` @ 0x18173D,asm 552950)——攻擊艦隊上限,對照用

```
db 4                    ; [18173D] Raider   cap=4
dw 304h                 ; [18173E..3F] = 04 03 → Marauder cap=4, Intruder cap=3
dd 202h                 ; [181740..43] = 02 02 00 00 → Huge cap=2, Titan cap=2, 預留=0
db 2 dup(0)             ; [181744..45] 預留=0
```

```go
var nMaxAntaranOffShips = [9]byte{4, 4, 3, 2, 2, 0, 0, 0, 0}
```

（攻擊艦隊偏好小型多艦:Raider×4 + Marauder×4 + Intruder×3 + Interdictor×2 + Harbinger×2 = 15 艘出擊上限,
和防禦艦隊「重型少艦」正好相反——這張表不是本次任務的目標,列出來只為交叉核對共用引擎的迴圈邊界。)

### `_antaran_ship_cost`(`byte_181734` @ 0x181734,asm 552946)——建造成本

```
db 2                    ; [181734] Raider  cost=2
db 5, 0Ch, 1Eh          ; [181735..37] Marauder=5, Intruder=12(0x0C), Interdictor=30(0x1E)
dd 4Bh                  ; [181738..3B] = 4B 00 00 00 → Titan=75(0x4B),預留×3=0
db 0                    ; [18173C] 預留=0
```

```go
var antaranShipCost = [9]byte{2, 5, 12, 30, 75, 0, 0, 0, 0}
```

---

## 4. 結論:安塔蘭母星防禦艦隊組成

### 下限(任何時候開戰都保證有)

`Load_Antaran_Defense_Fleet_` 每次執行都會把 `word_19917A[4]`(Titan 計數)拉到至少 1
(asm 90099-90101,`cmp word_199182,1 / jge / mov word_199182,1`),然後無條件呼叫
`Load_Antaran_Star_Fortress_`。**已知的寫入點裡,`word_19917A[]` 全陣列除了這一格,
其餘只有兩個地方會動:`Kill_Antaran_Defensive_Ship_` 的損耗遞減、`Build_Antaran_Ships_` 的
建造遞增**——沒有找到任何「開局灌一批初始艦」的寫入。`word_19917A` 是 BSS(`dw ?`),
新遊戲載入時預設歸零。因此:

> **新遊戲第一時間就打安塔蘭母星,理論上只會遇到 1 艘 Harbinger(Titan)+ 1 座星際要塞。**

### 上限(隨回合數養到滿之後)

依 `_n_max_antaran_def_ships = {0,0,3,2,7,0,0,0,0}`:

| 艦體 | 上限艘數 |
|---|---|
| Raider(Small) | 0(這套養成機制永遠不會建) |
| Marauder(Medium) | 0(同上) |
| Intruder(Large) | 3 |
| Interdictor(Huge) | 2 |
| Harbinger(Titan) | 7 |
| **星際要塞** | 1(每次都有,不計入上限表) |

> **養滿之後的安塔蘭母星防禦艦隊 = 3 艘 Intruder + 2 艘 Interdictor + 7 艘 Harbinger + 1 座星際要塞
> = 12 艘戰艦 + 1 座要塞,共 13 艘。**

### 是否隨難度變

**艘數上限不隨難度變**——`_n_max_antaran_def_ships` 是編譯期常數,`byte_199CB0` 沒有出現在
影響這張表的任何程式碼路徑裡。難度真正影響的三件事(見第 2 節表格):
① 資源累積速率(100/150/200%,決定養到滿的**速度**,不是**上限**)、
② 單艦裝甲加成(難度愈高單艦愈硬,愈大型艦體加成級距愈細)、
③ Hard/Impossible 建出第一艘 Titan 後全表成本永久 −10%(讓後續建造更快,一樣不影響上限艘數)。

### 損耗後怎麼補

`Kill_Antaran_Defensive_Ship_`(sub_646F9,asm 125496-125531)損失一艘時:
先把 `word_19917A[size]` 减 1(下限 0);**如果攻擊艦隊的同尺寸預備計數 `word_19918C[size] > 0`,
就立刻從那邊借調一艘轉正成防禦艦**(`word_19918C[size]--`、`word_19917A[size]++`)——
即防禦艦隊的損耗優先用「已經囤好但還沒出擊」的同型艦頂替,頂不了才等下一輪
`Build_Antaran_Defensive_Ships_` 用資源池慢慢重建。

## 4.5 2026-08-09 戰鬥 loader／compact design 位址勘誤

本節保留先前研究紀錄的更正脈絡，不把不同用途的 loader 名稱與位址靜默覆寫成新說法。
`Init_Ship_Designs`（raw `0x5514C`）的 jump table 與母星即時戰鬥 loader 是兩條平行路徑：

- 即時戰鬥 loader 的已證實對照是 `sub_55738` = Intruder、`sub_55B12` = Interdictor、
  `sub_55F67` = Harbinger；它們分別是 `Load_Combat_Antaran_Ship_` 的 Large、Huge、Titan 分支。
- `func_names.txt` 對這一段有「`Load_Medium_Antaran_Design_`」等歷史標籤，不能單獨當成
  語意證據；`Orion2.exe.asm` 的 `sub_4D10E` jump table、函式內 `aIntruder`／`aInterdictor`／
  `aHarbinger` 交叉參照與實際 `sub_55738`／`sub_55B12`／`sub_55F67` 寫入內容，才是本表採用的
  raw entry point 對照。原始標籤保留在 [`docs/re/weapon-mod-flags.md`](weapon-mod-flags.md)
  所述的反組譯追溯規則中，不用改名掩蓋這個衝突。
- compact strategic design 的 Large／Huge／Titan 分支是 `sub_559DC`／`sub_55E16`／`sub_562D6` 所在的
  `Init_Ship_Designs` 分派區；其中 `sub_559DC` 對應 Intruder 的 ID 31 ×3，`sub_562D6` 對應
  Harbinger 的 ID 31 ×6。`0x5565D` 與 `0x55E16` 不能再當成 Intruder／Harbinger 的即時戰鬥 loader 位址。
- 因此本研究紀錄先前把 `0x5565D`／`0x55E16` 寫成兩艘即時戰鬥設計 loader 的段落，現在以本節
  勘誤取代；原始行文仍留在版控歷史中，方便追查結論何時改變。
- `OrigWeaponTable` 的 ID `31` = `WeaponCatFighterBay`；remake 的
  `antaranDefenseUnit.CombatLoaderRaw`／`FighterBayWeaponID`／`FighterBayCount` 只保存這個已證實
  的槽位 provenance 與快速戰鬥已有的戰機艙貢獻，不代表其餘安塔蘭武器槽已解出。

### 已解出的標準戰鬥武器槽切片

在標準艦的戰鬥記錄中，槽起點是 `dword_192864 + 313*ship + 0x52 + 0x0B*slot`；
weapon ID 位於槽 `+0x00`、數量位於 `+0x02`、16 位元 raw flags 位於 `+0x04`。
以下只列已由 `sub_55738`／`sub_55B12`／`sub_55F67` 直接寫入的非空槽：

| 艦名 | raw loader | `(slot, weapon ID, quantity, raw flags)` | 消費狀態 |
|---|---|---|---|
| Intruder | `0x55738` | `(0,4,4,0)`、`(1,4,2,0x0002)`、`(2,24,5,0)`、`(3,13,2,0)`、`(4,31,3,0)` | ID／數量／flags 已保存；只有 ID 31 戰機艙接入既有快速戰鬥貢獻 |
| Interdictor | `0x55B12` | `(0,4,6,0)`、`(1,4,2,0x0002)`、`(2,24,15,0)`、`(3,13,2,0)`、`(4,4,8,0x0004)`、`(5,11,2,0)` | ID／數量／flags 已保存；其他槽的精確火力未知 |
| Harbinger | `0x55F67` | `(0,4,10,0)`、`(1,4,2,0x0002)`、`(2,24,20,0)`、`(3,13,3,0)`、`(4,4,15,0x0004)`、`(5,11,2,0x0002)`、`(6,37,1,0)`、`(7,31,6,0)` | ID／數量／flags 已保存；只有 ID 31 戰機艙接入既有快速戰鬥貢獻 |

這張表是「資料切片」，不是敵方火力解算器：`raw flags` 的 bit 語意、weapon ID 的完整
傷害／命中消費、敵方何時派出戰機與艦身方向仍未形成可重現規則，因此不把 ID 4／11／13／24／37
直接換成 remake 的攻擊值。

---

## 5. 誠實留白(讀不出來 / 沒追到底的部分)

- **`word_19917A[0..3]` 的真正起點**:上面推「新遊戲=0」是靠「窮舉了這個符號的全部
  讀寫 xref、沒看到第三個寫入點」這種**反向推論**,不是直接看到一行 `mov word_19917A,0` 的初始化碼。
  BSS 零填是標準做法,但不排除有一段用**動態計算位址**(例如 `rep movsd` 整段複製一個 struct)
  寫入這塊記憶體、而不會在反組譯裡以符號名 `word_19917A` 出現。這種寫法本次 grep 找不到,
  所以「新遊戲只有 1 艘 Titan+要塞」這個結論的信心來自「找不到反例」,不是「找到明確證據」。
- **`byte_199CB5` 是什麼**:餵給 `Antaran_Delay_`(回傳 200/100/0,決定資源累積從第幾回合開始算),
  明顯不是本專案已確立的難度旗標 `byte_199CB0`(兩者在 `Build_Antaran_Ships_` 同一個函式裡都用到,
  各自獨立比較,不會是同一顆)。本次沒有查它的寫入點,不知道對應哪個遊戲設定。
- **實際「養滿要幾回合」沒有算出數字**:公式(`elapsed = stardate−35000−Antaran_Delay_()`、
  每 25 回合結算一次、防禦池按 100/150/200% 累積、按成本表 2/5/12/30/75 花用)都齊了,
  但因為 `byte_199CB5` 未知、且防禦/攻擊資源分流有「防禦未滿前雙邊同速、滿了才轉單邊」這種
  依賴戰鬥損耗即時變化的分支,沒有一個乾淨的封閉公式可以直接套,沒有勉強湊一個估計值。
- **星際要塞(`Load_Antaran_Star_Fortress_` @ 0x4D18E,asm 90136-90546)只看到「比一般艦艇硬很多」,
  沒有精確戰力數字**:確認了①名字欄位是空字串(用共用的空字串常數 `dword_178A04`,不像
  Raider/Marauder 等五級戰艦各自有專屬艦名)、②裝甲初始化呼叫 `sub_127712` 8 次
  (數值 5,7,24,26,14,27,29,32,合計 164,對照 Raider 只有 2 次呼叫、合計 31,難度加成後最多 40)、
  ③武器初始化呼叫 `sub_127776` 8 次,掛載槽位數與一般艦艇相同(8 槽、每槽跨距 `0x0B`=11 bytes),
  但後面接一段依「總戰力值」按槽位容量比例分配傷害的演算法,**餵進這段分配算法的「總戰力值」
  (組語裡的 `di`)從哪裡算出來,沒有追回其源頭**——時間有限,這一段依任務指示屬於加分項,
  沒有勉強做完,留白。

### 2026-08-11 remake 消費端接線勘誤

上段保留的是「總戰力值來源未追回」的歷史結論；本輪沒有改寫它，而是把已直接讀到的
武器槽先接入 remake 的終局齊射。`Load_Antaran_Star_Fortress_` (`sub_4D18E`) 的四個
非空槽為：`(ID 11, seed 375, cap 99, raw flags 2)`、`(ID 4, seed 187, cap 198, 0)`、
`(ID 4, seed 187, cap 198, 4)`、`(ID 4, seed 375, cap 99, 2)`。ID／槽距／seed／
CapacityCap／raw flags 仍是**已證實**；`sub_6EE8E @ 0x6EE8E` 的 divisor 中間式已追回，
但 live tech 導出的當下數量與 flags bit 正式名稱仍是**未知**，故程式保留 raw 欄位並
在註解中標出推論邊界。

`antaranWeaponFirepower` 現在將光束、魚雷、具直接傷害的特殊武器按 CapacityCap 的
終局 full-cap policy 彙整到 `antaranDefenseCombatant.wmin/wmax`；獨立的
`antaranFortressRuntimeQuantity` 已保留 seed／divisor／live tech bucket 的正確計算，
待實際 runtime 狀態接入時才取代 full-cap policy。炸彈只對行星有效、戰機艙另走 Fighter Bays
中隊模型，不重複計算。這完成「要塞有完整已追回槽位火力」的 remake 消費端，但不宣稱
已追回原版總戰力分配演算法的所有中間欄位。
- **`Load_Antaran_Ship_Design_`(0x5514C,`Deploy_Antaran_Ships_` 專用的「設計」版本,
  跟本文的「戰鬥艦」版本是兩套平行系統)只看了 Small 那個 chunk**(確認名稱同樣是
  "Raider"),Medium/Large/Huge/Titan 四個 chunk 沒有展開比對,也沒有搞清楚
  `_Design_` 系列與 `_Combat_Ship_` 系列的實際差異用途(推測一個是給攻擊艦隊出擊部署用、
  一個是母星防禦即時載入戰鬥用,但沒有交叉驗證)。
- **Interdictor(Huge)、Harbinger(Titan)的難度裝甲加成表**:確認程式碼形狀
  (`mov al, byte_199CB0` 接四路難度分支,與 Intruder 相同寫法)存在,但沒有逐位元組
  抄出實際加成數字,只有 Small/Medium/Large 三級有完整數字。

## 2026-08-11 深度 IDA 勘誤：要塞 raw flag／容量 divisor／seed

上方歷史段落保留「99／198 是容量代理」的中途結論；本節是同一份
`Orion2.exe.i64`（SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`）
以 IDA Pro 9.4、IDA 線性位址重新追出的精確中間式。`sub_4D18E @ 0x4D18E` 在
`[record+0x33]=6` 後，`0x4D1D9`／`0x4D21A` 使用 `word_17F642 + class*0x0F`
的**直接 byte stride**；class 6 讀址是 `0x17F69C`，raw word `0x0384=900`。
先前探針把 0x0F 當 word stride 而報出的 `0x17F6F6=0x19` 已由
[`tools/ida/deep_fire_fortress.idc`](../../tools/ida/deep_fire_fortress.idc) 修正，不能再作
要塞常數。

本輸入的 raw 常數是 `word_180140=1200`、`word_180142=80`、`word_180144=120`、
`word_180146=120`。loader 的設計欄位為：

```text
AA = 3 * (floor(900 * 120 / 100) + 120) = 3600
C2 = 3 * (floor(900 * 120 / 100) + 120) = 3600
P  = signed_div(signed_div(5 * 1200, 4), 2) = 750
```

四個非空槽在 `+0x52 + slot*0x0B` 的 ID／quantity／raw flags 為：

| 槽 | ID | seed | raw `+0x04` | loader 上限／分割 |
|---:|---:|---:|---:|---|
| 0 | 11 | 375 | 2 | 99 |
| 1 | 4 | 187 | 0 | 198；超過 99 會拆同 ID／同 raw 的下一槽 |
| 2 | 4 | 187 | 4 | 198；超過 99 會拆同 ID／同 raw 的下一槽 |
| 3 | 4 | 375 | 2 | 99 |

`sub_6EE8E @ 0x6EE8E` 回傳的是 caller 用來除 seed 的 divisor，不是 live quantity。
要塞 caller 傳 `EAX=word_19988E`、`EDX=ID`、`EBX=1`、`CX=1`、stack raw fortress
flag `0` 與 slot raw `0/2/4`。對本四槽：`K=word_17F811[ID*0x1C]`，ID 4/11 分別
是 15/30；`sub_6A636 @ 0x6A636` 使 raw 0/2/4 的第一百分比成 100/200/50，
`sub_6A406 @ 0x6A406` 提供的第二個加成對這三種 raw 都是 0。`sub_6F11C @ 0x6F11C` 對 ID 4/11
靜態回傳 B=1，`sub_6E70A @ 0x6E70A` 的 required-tech raw field 分別是 `0x7B`／
`0x2F`，所以 T 仍取決於執行期科技資料。B=1 時 `sub_6E60E @ 0x6E60E` 的
base 是 T=2→700、T=3→600、其他→400，最後為
`floor((base*X+500)/1000)`，正值零結果夾成 1。

因此在三個 T bucket 下的靜態重算例為：ID 4 的 `X(raw0/raw2/raw4)=15/30/7`，
ID 11 raw2 的 X=60；slot 1 raw0 的 divisor `11/9/6`、數量 `17/20/31`，
slot 2 raw4 的 divisor `5/4/3`、數量 `37/46/62`，slot 0 raw2 的 divisor
`42/36/24`、數量 `8/10/15`。這些是公式範例，不是宣稱某個 runtime 開局必然值；
`T`、合法 `word_19988E` 角色與 `sub_6D048` 的 live BSS 狀態仍需原版 runtime。

**證據結論：** 四槽 ID、seed、raw flag、99／198 上限與容量 divisor 中間鏈為
**已證實**；raw bit 的正式玩法名稱、live tech 導出的當下數量與逐值 DOSBox oracle
為**未知／外部待驗**。remake 保留 raw 欄位並接入已追回直接火力，不把上限冒充 live
數量，也不把 raw `4` 命名成 ARM/FST。
