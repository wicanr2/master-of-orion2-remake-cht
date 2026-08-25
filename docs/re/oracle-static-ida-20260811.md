# 原版 oracle：IDA Pro 靜態批次（2026-08-11）

## 範圍與限制

本批次只做靜態反組譯，不啟動原版，不執行 `VESA.COM`，也不把 DOSBox 缺件補成假設。
目的為確認音訊分派鏈、回查外交音樂賦值鏈，並把相鄰的 raw
`sub_3AC20 @ 0x3AC20`／`sub_3AD57 @ 0x3AD57`、敵方戰機、外交門檻與 `.GAM`
匯入的證據等級固定下來；外部符號名稱衝突不阻塞 remake。

執行環境：

- 工具：IDA Pro 9.4 `idat`，Hex-Rays `9.4.0.260610`；容器映像
  `ida-pro-9.4-ver2:latest`。
- 探針：[`tools/ida/oracle_probe.idc`](../../tools/ida/oracle_probe.idc)，非破壞性地讀取
  函式邊界、交叉參照、原始 `sub_`／`word_` 名稱與有限指令窗口；輸入 `.i64` 複製到
  容器暫存區，沒有保存或修改私有資料庫。
- 位址基準：IDA 線性位址；DOS LE object #1 code base `0x10000`。因此 object-offset ledger
  的 `0x3CD21` 若要對照這次 IDA 輸出，必須先轉成線性位址 `0x13CD21`，不可把兩種數字
  當成同一個位址。
- 結果：`runtime_not_executed`。原始 report 留在本輪 Docker 暫存輸出；本文件是其受版控
  的摘要與證據分級。

輸入檔案雜湊（SHA-256）：

| 輸入 | 大小 | SHA-256 |
|---|---:|---|
| `Orion2.exe` | 2,644,842 | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` |
| `Orion2.exe.i64`（本批次主要資料庫） | 23,691,583 | `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e` |
| `Orion2.exe.asm`（交叉參考，不取代資料庫） | 21,182,592 | `76cac6231a60da0fdba705907a88a853a1d345ed7bb7c15788b280fdbb259a18` |
| `symbols_fixed.tsv`（外部符號索引） | — | `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28` |
| `symbols_full.tsv`（外部符號索引） | — | `bd8c57fa008797d62123eb22b0f03f9b1b644eaafc9b6e1538cc0496aa1a504e` |

## 2026-08-11 同一輸入深度批次（覆蓋指定三項舊未知）

本節是同日第二輪、直接以目前 `Orion2.exe.i64` 的 IDA 交叉參照與函式邊界追出的結果；
下方首輪 bounded dump 的「未知」敘述保留作歷史，不再代表這三條資料流的目前結論。探針
仍是 [`tools/ida/oracle_probe.idc`](../../tools/ida/oracle_probe.idc)，執行方式為複製 `.i64`
到容器暫存區後以 IDA Pro 9.4 `idat -A` 讀取，未保存原始資料庫；輸出只記錄原始
`sub_`／`dword_` 名稱、IDA 線性位址與指令運算元。

### 先處理位址與來源衝突

兩份外部符號索引在同一個位址區段有位移／別名衝突。本節以 IDA 的原始名稱、函式邊界、
直接 xref 與資料流為主，外部名稱只作附加索引：

| IDA 線性位址 | `symbols_fixed.tsv` | `symbols_full.tsv` | 本節採用的保守寫法 |
|---:|---|---|---|
| `0x4D10E` | `Load_Combat_Antaran_Ship_` | `Repair_All_Combat_Ships_` | `sub_4D10E` 五分支 dispatcher |
| `0x53146` | `Diplomacy_Test_` | `Check_Treaty_Proposal_` | `sub_53146` 評分／回傳桶 |
| `0x533F4` | `Get_Break_Treaty_Message_` | `Diplomacy_Test_` | `sub_533F4` 隨機外交測試 |
| `0x539D9` | `Get_Demand_Response_` | `Get_Gift_Response_` | `sub_539D9` 外交回應分支 |
| `0x10E2F` | `Load_Game_` | `_special_names_seg` | `sub_10E2F` 直接開啟 `.GAM` 讀入 |
| `0x1160B` | `Save_Game_` | `Load_Game_` | `sub_1160B` 直接開啟 `.GAM` 寫出 |
| `0x1930DC` | `_leaders` | `_save_officers` | 原始全局指標 `dword_1930DC` |

`/private/re/orion2_all.c` 的 SHA-256 是
`c2b5c30701019c0cc58763eb29c2abddb55eb551e1e7e52f68070d629e694505`，但它與目前輸入不是同一
版本：同目錄的 `ORION95.EXE` SHA-256 是
`6e19afdc98f1aedcb8d2f974d5b658b0c855f54529bdabdde193f5266e275185`，目前主輸入
`Orion2.exe` 則是上表的 `7ae2ac...`。因此本節不以 `orion2_all.c` 的變數名、反編譯函式名或
行號作主證據；它只能保留為歷史搜尋索引。

### 敵方五級 blueprint 與防禦艦資料

**已證實（raw blueprint writer／槽位資料）：**

- `sub_4D10E @ 0x4D10E` 在 `0x4D10E` 比較 `dx,4`，由 `0x4D123`、`0x4D129`、
  `0x4D12F`、`0x4D135`、`0x4D13B` 分別跳到 `sub_55161`、`sub_5542C`、
  `sub_55738`、`sub_55B12`、`sub_55F67`；這是五種艦體設計 writer 的直接 dispatcher。
- `sub_4D141 @ 0x4D141` 以 `word_19917A[size*2]` 驅動五種尺寸的數量迴圈，呼叫上述
  dispatcher；`sub_4D18E @ 0x4D18E` 另行建立星際要塞設計。這證實「艦隊載入」確實把
  blueprint writer 的結果放入戰鬥記錄，但要塞的完整戰力換算仍未完全追回。
- 五個 writer 都以 `dword_192864 + 0x139*ship_index` 為記錄基址，並在 `+0x25` 寫入
  尺寸／類別值 `0..4`；字串 xref 分別是 `Raider`、`Marauder`、`Intruder`、
  `Interdictor`、`Harbinger`。這些是 raw 寫入與字串交叉參照的已證實結果；`+0x25` 的
  高階遊戲語意仍保留原始欄位索引。

對每個非空武器槽，`slot_base = record + 0x52 + 0x0B*slot`；本輪直接讀到的
`(weapon ID, quantity, raw flags)` 如下。raw flags 只表示記錄中的 `slot+0x04` 16 位元值，
不把 bit 4／2／0 直接命名成 ARM、FST 或其他改造：

| raw writer／原始名稱 | 非空槽 `(ID,數量,raw flags)` |
|---|---|
| `sub_55161 @ 0x55161`／Raider | `(4,2,0)`、`(24,1,0)` |
| `sub_5542C @ 0x5542C`／Marauder | `(4,3,0)`、`(4,1,2)`、`(24,1,0)` |
| `sub_55738 @ 0x55738`／Intruder | `(4,4,0)`、`(4,2,2)`、`(24,5,0)`、`(13,2,0)`、`(31,3,0)` |
| `sub_55B12 @ 0x55B12`／Interdictor | `(4,6,0)`、`(4,2,2)`、`(24,15,0)`、`(13,2,0)`、`(4,8,4)`、`(11,2,0)` |
| `sub_55F67 @ 0x55F67`／Harbinger | `(4,10,0)`、`(4,2,2)`、`(24,20,0)`、`(13,3,0)`、`(4,15,4)`、`(11,2,2)`、`(37,1,0)`、`(31,6,0)` |

因此「敵方精確 blueprint」在**靜態原始記錄層**已結案：目前 remake 實際使用的
`Intruder`／`Interdictor`／`Harbinger` 三組槽位已與同一輸入對齊，並修正 Interdictor／
Harbinger 第 4 槽的 raw flags `0x0004`。尚未宣稱的部分是 raw 欄位的完整語意、要塞總戰力
分配、敵方戰機何時出擊及逐彈命中／傷害；那些屬於消費端 oracle，不是 blueprint writer
本身未找到。

**證據等級：** writer 的位址、函式邊界、名稱、record stride、尺寸值與上述槽位為**已證實**；
把尺寸映射成 remake 的 Battleship／Titan／Doom Star 是**強推論／模型映射**；要塞強度與
raw flags 的 bit 名稱是**未知**。

### 外交精確門檻

#### `sub_53146 @ 0x53146`：評分回傳桶

`sub_53146` 從 `dword_197F98 + 0xEA9*empire` 的欄位、政府／性格表、關係陣列與
`sub_E5E09` 組成最後的 `bx` 評分；`0x531E7` 另呼叫 `sub_1247A0` 並以 `0x64` 作為
亂數上界。函式尾端的立即數分支是完整且無歧義的：

| 最終 `score` | 回傳值 | 原始分支 |
|---:|---:|---|
| `< -75` | `0` | `0x533C4` 比較 `0xFFFFFFB5`，`0x533C9` 清零 |
| `-75 ≤ score < -50` | `2` | `0x533CF` 比較 `0xFFFFFFCE`，`0x533D4` 寫入 `2` |
| `-50 ≤ score < 0` | `1` | `0x533DD` 測試 `bx`，`0x533E2` 寫入 `1` |
| `score ≥ 0` | `3` | `0x533EB` 寫入 `3` |

這是**已證實的外交評分桶門檻**，但不是把 `score` 直接等同於 remake 的 `Relation`；
前面的欄位加總、政府／性格修正與亂數仍會影響輸入。

#### `sub_533F4 @ 0x533F4` 與 `sub_539D9 @ 0x539D9`：回應分支

`sub_533F4` 不是另一組固定四桶：`0x53410` 呼叫 `sub_1247A0` 上界 `0xC8`，
`0x5345B` 以 `word_180DB8[government]` 作比較，並在 `dx > 0x64` 時進一步呼叫
`sub_51078`；這是一個含亂數、政體表與雙方狀態的外交測試，不能誠實縮寫成單一常數。

`sub_539D9` 的 `var_4` 先在 `0x53A92` 分成九種回應模式；在不同模式／外交狀態下，
目前 IDA 原始指令直接證實以下回應邊界（`word ptr [esi]` 是原版訊息碼，`[ecx]` 是
該分支的成功／失敗旗標）：

| 路徑 | 分支門檻 | 原始位址／結果 |
|---|---|---|
| 外交狀態 `+0x627` 為 1／2 的路徑 | `score < -100` 或 `sub_53E96(...) == 6` | `0x53C7E`／`0x53C9D`，訊息 `0xA5`，旗標清零 |
| 同一路徑 | `-100 ≤ score < -50` | `0x53CC4`／`0x53CE9`，訊息 `0xA6` |
| 同一路徑 | `-50 ≤ score < -25` | `0x53D0F`／`0x53D7E`，訊息 `0xA7` |
| 同一路徑 | `-25 ≤ score < 0` | `0x53D85`／`0x53D8E`，訊息 `0xA8` |
| 同一路徑 | `score ≥ 0` | `0x53E38`／`0x53E3D`，訊息 `0xA9`，旗標寫 1 |
| 另一個外交狀態路徑 | `score < -150` 或 `sub_53E96(...) == 6` | `0x53DA3`／`0x53C9D`，訊息 `0xA5` |
| 另一個外交狀態路徑 | `-150 ≤ score < -75` | `0x53DC2` 後走 `0x53D7E`，訊息 `0xA7` |
| 另一個外交狀態路徑 | `-75 ≤ score < 0` | `0x53E2D`／`0x53D8E`，訊息 `0xA8` |
| 另一個外交狀態路徑 | `score ≥ 0` | `0x53E2D`／`0x53E38`，訊息 `0xA9`，旗標寫 1 |

這些立即數與比較關係是**已證實**；「哪一個 `var_4` 模式對應現金、科技、條約或需求」
以及完整 `score` 輸入欄位的遊戲語意仍是**強推論／部分未知**。因此 remake 不直接把
`-75`、`-50` 或 `-25` 硬塞成單一 AI 接受率；報告現在可以聲稱「原版門檻常數已追回」，
不能聲稱「所有外交提案的 end-to-end 接受公式都已追回」。

### `.GAM` 全局匯入與寫出對稱

**已證實（同一輸入、同一全局指標、同一 record shape）：**

- `sub_10E2F @ 0x10E2F` 在 `0x10E80` 組合 `.GAM`、在 `0x10E88` 使用 `"rb"`，
  `0x10EE9` 以 `fread_` 讀 4 bytes，`0x10EEE` 比較版本 `0xE0`；通過後由
  `0x10FB8..0x114B0` 連續讀取多個全局陣列／欄位。
- 領袖／軍官全局區塊在 `0x110DB..0x110F1`：`edx=0x3B`、`eax=dword_1930DC`、
  每輪 `edi += 0x3B`，直到 `edi == 0xF71`；`0xF71 = 0x3B * 0x43 = 59 * 67`。
  這是直接 `fread_(dword_1930DC + offset, 0x3B, 0x43, file)` 的等價資料流，並非
  只讀目前選中的一名軍官。
- `sub_1160B @ 0x1160B` 是同一輸入的寫出端：`0x11673` 組合 `.GAM`、`0x1167B`
  使用 `"wb"`、`0x116E9` 寫版本 `0xE0`；`0x11821..0x11832` 以 `ebx=0x43`、
  `edx=0x3B`、`eax=dword_1930DC` 呼叫 `fwrite_`。讀／寫的指標、元素大小與筆數一致。
- `sub_1307F @ 0x1307F` 以 `0x13090`／`0x13096` 傳入 `0x3B`／`0x43`，
  `0x130A6` 設 `ecx=0xF71` 並把資料複製到 `dword_1930DC`，接著在 `0x130B9..0x13169`
  以 `si < 0x43` 逐筆使用 `+0x26`、`+0x2A`、`+0x2E`、`+0x2F`、`+0x32`。這是
  67 筆全局記錄的初始化／消費端交叉證據。
- IDA 資料 xref 也保留 `dword_1930DC` 的原始讀寫端：`0x10E23` 設指標、
  `0x110E5` 讀入、`0x1182B` 寫出、`0x130AB` 及後續多個遊戲函式消費；沒有把
  `_leaders` 或 `_save_officers` 的外部別名當成唯一語意來源。

因此「原版 `.GAM` 是否把全局領袖／軍官區塊匯入」已達**已證實**；「remake 是否已能
直接讀原版 `.GAM`」仍是另一件事，目前仍使用 JSON 保存，原生 `.GAM` importer 不在本輪
實作範圍。原版任命／任期門檻也不是檔案匯入本身，仍列為未追回的下游規則。

### 深度批次的結論邊界

| 項目 | 本輪結論 | remake 狀態 |
|---|---|---|
| 敵方 blueprint | 五級 writer、名稱、record stride、尺寸值與非空武器槽已由同一 `.i64` 靜態證實；要塞完整戰力與 raw bit 語意未完 | 三種實際防禦艦已接，並修正兩個 `0x0004` raw flags |
| 外交精確門檻 | 評分桶 `-75/-50/0` 與回應路徑 `-100/-50/-25/0`、`-150/-75/0` 已由比較指令證實；完整輸入欄位／模式語意仍部分未知 | 不冒充原版 AI end-to-end 接受表 |
| `.GAM` 全局匯入 | `sub_10E2F` 讀入、`sub_1160B` 寫出、`sub_1307F` 消費的 `0x3B×0x43` 全局區塊已證實 | 原生 `.GAM` importer 仍是可選擴充，不影響 JSON remake |

## 首輪靜態鏈（保留作歷史結果）

| IDA 線性位址／原始名稱 | 觀察 | 等級 |
|---|---|---|
| `0x24677`／`sub_24677` | `cmp word_180EB8, 0x64`；`<= 100` 走 `STREAM.LBX`，`> 100` 走 `STREAMHD.LBX`，後者在 `0x246FF` 減 `0x64`。有多個直接 xref，包含 `0x14AE9`、`0x14C7A`、`0x1526D`。 | **已證實**：單一編號空間與兩個 LBX 的分流 |
| `0x2484F`／`sub_2484F` | `j___clock_` 後以 `3` 除，`0x248C8` `inc edx`，`0x248CE` 保存 `1..3`，再開 `STREAM.LBX`。 | **已證實**：背景曲池為 STREAM 1–3 |
| `0x2496C`／`sub_2496C` | `j___clock_` 後以 `3` 除，`0x249BB` 加 `4`，`0x249C3` 保存 `4..6`，再開 `STREAM.LBX`。 | **已證實**：戰鬥曲池為 STREAM 4–6 |
| `0x1B92E`／`sub_1B92E` | `0x1B95E` 讀取 `byte [dword_197F98 + 0xEA9*族索引 + 0x25]`；在 `0x1BD2B` 再次讀取同一欄位，`0x1BD31` 加一，`0x1BD32` 寫入 `word_19AA3C`。 | **已證實**：好關係曲目計算是逐族記錄欄位 + 1；變數別名仍以 raw 名稱為準 |
| `0x1BD38..0x1BD50` | `eax=3`；`0x1BD46 call sub_1247A0`；`0x1BD4B add eax, 0xD`；`0x1BD50 mov word_19AA46, ax`。 | **已證實**：壞關係曲目為 helper(3)+13 |
| `0x1247A0`／`sub_1247A0` | 有 `0xFFFFFFFF / N` 的拒絕取樣門檻，並使用 `0x41C64E6D`、`0x3039` 更新 LCG，再回傳範圍內整數。 | **已證實**：該 helper 是有界均勻亂數路徑；原始名稱仍是 `sub_1247A0` |

外部 `symbols_fixed.tsv` 將 `word_19AA3C`／`word_19AA46` 分別標為
`_diplomacy_good_music`／`_diplomacy_bad_music`；`symbols_full.tsv` 對同一批全域有不同舊名稱。
因此本文件採「dataflow 已證實、語意別名強推論」的寫法，不以外部名稱取代 IDA 原始定位。

## 首輪 ledger 目標位址的回查結果（保留；不覆蓋同輸入 raw 位址）

下表保留原始 ledger 標籤，並明確指出這次 IDA 是否足以支持該標籤。所有位址都是 IDA 線性位址。

| ledger 標籤 | IDA 結果 | 等級／用途 |
|---|---|---|
| `object1+0xD0F0` → `0x1D0F0`，舊標為 `Start_Diplomacy_Music_` | 位址落在短小 `sub_1D0F0`／未建立函式邊界，沒有外交音樂 body 證據。 | **未知／勘誤**：保留舊索引，不再用此位址支持語意。真正音樂賦值鏈是 `0x1BD2B..0x1BD50`。 |
| `object1+0x533F4` → `0x1533F4`，舊標為 `Diplomacy_Test_` | 落在 `sub_153388` 內，看到格式化／狀態呼叫與 `+4` 參數，沒有可安全升格的接受門檻。 | **未知**：原版外交精確門檻仍需 runtime oracle。 |
| `object1+0x539D9` → `0x1539D9`，舊標為 `Get_Gift_Response_` | 落在 `sub_15396C` 的分支，看到多個資料表與關係候選路徑，但未能從這個 bounded dump 還原完整分數／接受公式。 | **強推論／未知**：餽贈方向可追，精確門檻與完整表仍不確定。 |
| `object1+0x3AD57` → `0x13AD57`，舊標為 `Fire_Fighter_Bomb_` | 落在 `sub_13AD33`；函式開啟 `"r+b"`、讀取 8 bytes、檢查 magic `0xFEAD`，再從 `+0x804` 讀 `0x1FE` bytes。 | **已證實僅為檔案／資料讀取鏈**；不足以證實逐彈炸彈傷害或命中參數。 |
| `object1+0x3DFE0` → `0x13DFE0`，舊標為 `Fighter_Ocv_` | 落在 `sub_13DFBF`；可見 runtime 狀態複製／計算，但沒有敵方逐艦 blueprint。 | **強推論／未知**：只能作戰機下游線索。 |
| `object1+0x3CD21` → `0x13CD21`，外部索引曾標 `Missile_Speed_` | IDA raw body 是間接呼叫、`putch_` 與回傳；`symbols_fixed.tsv` 與 `symbols_full.tsv` 對此 object offset 的語意也互相衝突。 | **位址／命名衝突**：本批次不把它當新速度證據；保留舊 `docs/re/weapon-mod-flags.md` 歷史索引，待以同一來源資料庫重做地址對照。 |
| `object1+0x3C892` → `0x13C892`，外部索引曾標 `Fire_Missile_`／`Draw_Missiles_` | 位於 `sub_13C22C` 尾端；目前 dump 可見距離／速度樣式計算與 runtime 狀態，未能單獨證實完整戰機 blueprint。 | **未知**：不阻塞 remake 的抽象戰機模型。 |
| `object1+0x2B7CC` → `0x12B7CC`，外部索引曾標 `Average_Missile_Kills_`／`Get_Effective_Missile_Strength_` | 位於 `sub_12B79D`，可見讀取 record 的 raw bit `0x10`，但沒有完整 caller 語意。 | **強推論／未知**：raw bit 與精確攔截器命名仍需更可靠的同源資料流。 |

## 首輪批次當時未證實項目（保留歷史；深度批次已更新者不再適用）

- `VESA.COM` 的載入／VESA runtime 路徑：本批次沒有執行檔或 DOSBox runtime 入口，保持未知。
- 首輪 bounded dump 當時把敵方逐艦設計、戰機完整 blueprint、兩個 raw 函式的逐彈參數
  一併列為局部資料流；其中敵方五級 blueprint 的 raw writer／槽位與戰機下游消費端已由
  上方同一輸入深度批次覆蓋，仍未完成的是兩個 raw 函式的逐彈 runtime 參數。
- 外交特殊貿易 byte table、完整 `SABOTAGE` 分數／防守策略，以及從原始欄位一路到每種提案
  的 end-to-end AI 接受表，仍列 oracle 差異；上方已追回的是外交比較常數與回應分支門檻。
- `.GAM` 全局領袖匯入的檔案讀寫鏈已由上方深度批次證實；仍未知的是原版任命／任期下游規則，
  而 remake 的 JSON 領袖 ID 保存與原生 `.GAM` importer 仍是兩件事。

首輪結果本身仍保留供追溯；目前執行時差異仍不阻塞 remake。只有在取得可啟動 `VESA.COM` 的
DOSBox 環境後，才值得做動態逐值驗收。後續若推翻本表結論，必須保留本文件、輸入雜湊、位址
基準與新舊證據差異。

## 2026-08-11 remake 接線收尾（保留上方 oracle 邊界）

同一份輸入與位址基準下，這一輪把已足夠支撐 remake 的下游接入，但沒有把「可靜態證實」
誇大成「所有 runtime 逐值等價」：

| 項目 | 接入內容 | 證據等級／邊界 |
|---|---|---|
| 敵方戰機下游命中／傷害 | `internal/gamedata/fighter_damage.go` 使用 ID 31 第二組 `1..4 / 4..16 / 2..7`；`internal/shell/fighter_attack.go` 分開轉寫 `sub_3AD57 @ 0x3AD57` 的 roll `<=95`／40 命中式與 `sub_3AC20 @ 0x3AC20` 的直接插值式；Bomber profile 走後者，逐架送入最弱盾／甲／結構消費切片 | 兩條公式的靜態算術為**已證實**；`sub_3DF8D @ 0x3DF8D` 的完整攻方加成欄位、raw flag `&4` 正式名稱與 raw record 輸出指標仍是**強推論／未知**，故實作保留 `RawFlags`，不改名 |
| 星際要塞完整火力 | `sub_4D18E @ 0x4D18E` 的四個非空槽以 seed／raw flags／CapacityCap 保存：`(11,375,2,99)`、`(4,187,0,198)`、`(4,187,4,198)`、`(4,375,2,99)`；快速戰鬥採 full-cap policy，光束／魚雷／特殊武器彙整進終局齊射，炸彈與戰機艙分流 | 槽位 ID／seed／raw flags／容量上限為**已證實**；`sub_6EE8E @ 0x6EE8E` 的 divisor 中間式已追回，live tech 導出的當下數量與 raw flag 正式名稱仍是**未知**，不把 cap 當固定 runtime quantity |
| `.GAM` 支援 | `internal/shell/gam_import.go`／`LoadGAMSession`／`ImportGAM`；`LoadSession` 以 little-endian `0xE0` magic 分流，轉入星系、行星、殖民地、AI、67 筆領袖、艦隊、建築、外交旗標與研究中主題 | `.GAM` 讀入全局 `0x3B×0x43` 的原版檔案形狀為**已證實**；研究完成 byte 語意、special bit 與原版任命／任期下游仍**未知**，匯入報告會明示而不猜測 |

抽樣驗證使用 `SAVE10.GAM`：`stardate=35000`、36 星、82 行星、1 個 fallback 玩家殖民地、4 個 AI、3 艘玩家艦、67 筆非空全局領袖；`internal/shell` 與 `internal/gamedata` 測試通過。這不取代未取得 `VESA.COM` runtime 時的逐彈動態 oracle。

## 2026-08-11 Fire／要塞深度追查勘誤（IDA Pro 9.4）

本節覆蓋本文件較早把 `Fire_Fighter_Bomb` 參數與要塞容量中間值統列為未知的快照；舊
結論保留，以下以同一份 `Orion2.exe.i64` 的原始函式邊界、指令運算元、資料表 bytes
與直接 caller context 更正。工具仍是 IDA Pro 9.4 `idat`／Hex-Rays 9.4.0.260610，
容器是 `ida-pro-9.4-ver2:latest`，位址基準是 IDA 線性位址、DOS LE object #1 code
base `0x10000`；只讀資料庫，`runtime_not_executed`。輸入雜湊沿用本檔表頭：
`Orion2.exe.i64=4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`、
`Orion2.exe=7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`；本輪
非破壞性探針是 [`deep_fire_fortress.idc`](../../tools/ida/deep_fire_fortress.idc)。

### 先保留函式名稱衝突

兩份外部符號索引對相鄰函式互相錯位：

| IDA 線性位址 | `func_names.txt` | `symbols_fixed.tsv` | 保守引用 |
|---:|---|---|---|
| `0x3AC20` | `Fire_Special_Weapon_` | `Fire_Fighter_Bomb_` | `sub_3AC20` |
| `0x3AD57` | `Fire_Fighter_Bomb_` | `Fire_Fighter_Beam_` | `sub_3AD57` |

`func_names.txt` SHA-256 是 `7d37a88c59fd3f31d0436fd5386b41b90fac5144346cffd66a0231f87cc4f04e`，
`symbols_fixed.tsv` SHA-256 是 `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`。
因此下文不把外部名稱當成已證實語意，而是把兩個 raw function 都列出；這也避免把
`sub_3AD57` 的戰機對艦路徑與 `sub_3AC20` 的相鄰另一條公式混成一條公式。

### `sub_3AC20 @ 0x3AC20`：外部索引標作 `Fire_Fighter_Bomb_` 的候選路徑

**已證實的輸入／輸出契約（raw）：**

- `EAX` 低 16 位先 sign-extend，作 `dword_192B14 + int16(EAX)*0x1A` 的 runtime
  record；record 的 `+0x06` 進 `DI`，`+0x04` 經 `dword_192864 + signed*0x139` 找設計。
- 設計的 `byte +0x20` 經 `word_199242[design_byte*2]` 找武器表索引 `W`；原始
  `raw/min/max` 分別是 `word_17F815[W*0x1C]`、`word_17F817[W*0x1C]`、
  `word_17F819[W*0x1C]`。
- `EBX`／`ECX` 是兩個 word 輸出指標，入口立即清零；`EDX` 及 stack 參數保留給
  `sub_39985 @ 0x39985` 的下游契約。record 的 `+0x0F` 是本次迴圈數量，這個
  函式把每次結果累加到兩個輸出。

**已證實的隨機／傷害式：**

令 `R = sub_1247A0(100)`，`R` 為 **1..100**。令 `S=max-min`：

```text
if S > 0:
    D = min + floor((floor(100 / (2*S)) + R) * S / 100)
else:
    D = min
```

這條相鄰路徑沒有 `sub_3AD57` 的 OCV／40 命中門檻分支；不能只因外部索引把它叫作
`Fire_Fighter_Bomb_`，就把另一條公式套過來。

重製程式以 `internal/shell.ResolveFighterBomb` 保留這條純數值式，並在
`cmd/moo2/tacticalfighter.go` 的 Bomber profile 分流使用；這是依 caller 的
weapon ID `0x1E`（Bomber Bays）接入已證實公式，不是替兩份外部索引決定函式名稱，
也不宣稱已把 raw record 的兩個 word 輸出指標完整模型化。

### `sub_3AD57 @ 0x3AD57`：戰機對艦命中／傷害路徑

**已證實的 runtime 參數：**

- `EAX` 低 16 位同樣索引 `dword_192B14 + int16(EAX)*0x1A`；`DI=[record+0x06]`。
  目標設計由 `[record+0x04]` 乘 `0x139` 找到，武器表索引為
  `W=word_199254[design_byte+0x20]`。
- `raw=word_17F815[W*0x1C]`、`min=word_17F817[W*0x1C]`、
  `max=word_17F819[W*0x1C]`；`ECX` 是主要 word 輸出指標，`[arg_0]` 是次要
  word 輸出指標，入口都清零。入口的原始 `EDX` 與 `EBX` 先保存；保存的 `EBX`
  決定艦艇／戰機 runtime record 下游分支。
- `var_18=0`；只有 `test byte ptr [raw+1],1`（raw bit `0x0100`）時才設成
  `0x40`。所以在這個函式內 `var_18` 的可達值只有 `0`／`0x40`。

令 `OCV` 為攻防修正：保存的 `EBX==0` 時，
`OCV=sub_3DF8D(W)-target_design[+0x36]`；否則是
`OCV=sub_3DF8D(W)-sub_3DFE0(runtime[DI].word0, target_design.byte+0x20, ECX)`。
`sub_3DF8D @ 0x3DF8D` 的下游組合已證實為：`sub_36266(W)`、
`sub_5680D(W)*0x16` 對 `word_17FE00` 的查表、`W<8` 時玩家欄位
`dword_197F98 + W*0xEA9 + 0x8A6`，再加 `sub_35EAE(W)` 與 `0x32`；其中
`sub_36266`／`sub_5680D` 的每個欄位語意仍不命名為既定事實。

正常隨機分支令 `R=sub_1247A0(100)`、`M` 為修正後擲骰：

```text
M = R
if R <= 95:
    M = min(R + OCV, 100)       # 只有上界夾限，沒有下界夾限
if M < 40:
    miss
else if max - min > 0:
    D = min + floor((M - 40) * (max - min + 1) / 60)
else:
    D = min
if DI == 0:
    D = signed_trunc_toward_zero(D / 2)
```

`max-min+1` 與沒有最終 `max` 夾限是原始指令的實際端點行為；`M=100` 可能得到
`max+1`，不可自行改成一般化的 `[min,max]` 插值。`var_18` 的 `test var_18,4`
位於 `0x3AE4E`，但 `0`／`0x40` 都不含 bit `0x04`，故其後的 `+0x19`／半 min/max
分支是**不可達**。這推翻 remake 曾經把 `RawFlags & 4` 當作命中／半傷害效果的接法；
目前應保留 raw 欄位，但不得賦予該 bit 未證實的玩法語意。

命中後，保存的 `EBX==0` 走 `sub_39985 @ 0x39985` 的盾／裝甲／結構消費；否則走
`sub_3A0B9 @ 0x3A0B9`，對 `dword_192B14 + DI*0x1A` 的 `+0x0F` runtime 數量扣除
並累加輸出。最後仍經 `sub_3D299` active check；戰機分支另呼叫 `sub_3DDD8`／
`sub_3BC80`。這些下游控制流是**已證實**，但 `sub_3DF8D` 的完整設計／科技語意與
兩份外部名稱的真正命名，仍是**強推論／未知**。

### `sub_1247A0 @ 0x1247A0`：隨機 helper

**已證實：** `N=0` 回傳 1；`N>0` 先算 `floor(0xFFFFFFFF/N)`，以
`dword_1B9E34 = dword_1B9E34*0x41C64E6D + 0x3039` 的 LCG 做 rejection，最後
回傳 `floor(PRNG / quotient)+1`。因此本節兩處 `sub_1247A0(100)` 都是均勻候選的
1..100 介面；沒有 runtime seed／每局序列值，不能只靠靜態資料聲稱逐次擲骰相同。

### `sub_4D18E @ 0x4D18E`：要塞 raw flag、seed 與容量上限

**先更正探針的地址算式：** `0x4D1D9`／`0x4D21A` 的原始指令先做
`class*0x0F`，再以 `word_17F642[eax]` 直接取址；表格記錄是 **0x0F bytes**，
不是 0x0F words。class 6 的實際讀址是 `0x17F642 + 6*0x0F = 0x17F69C`，
raw bytes `84 03`，故值為 **`0x0384=900`**。探針舊版把 stride 乘 2 後輸出的
`0x17F6F6=0x19` 是探針報告錯誤，不是原版值；`deep_fire_fortress.idc` 已改成
直接 byte stride 並保留這項註記。

loader 先寫 `[record+0x32]=5`、`[record+0x33]=6`。因此：

```text
classPower = word_17F642[0x0F-byte-stride, class=6] = 900
AA = 3 * (floor(classPower * word_180146 / 100) + word_180146)
C2 = 3 * (floor(classPower * word_180144 / 100) + word_180144)
```

本輸入的 `word_180140=1200`、`word_180142=80`、`word_180144=120`、
`word_180146=120`，所以 `[+0xAA]=[+0xC2]=3*(1080+120)=3600`。這些是 loader
建立的 raw 設計欄位，不等於四槽當下已裝彈數量。

基礎 seed 的原始整數算式在 `0x4D23A..0x4D251`：

```text
P = signed_div(signed_div(5 * word_180140, 4), 2)
  = signed_div(signed_div(6000, 4), 2)
  = 750
```

四個非空槽的 raw 寫入與 caller（`0x4D42D`、`0x4D4BA`、`0x4D5A3`、`0x4D687`）為：

| 槽 | weapon ID | seed | raw `+0x04` | caller 上限／分割行為 |
|---:|---:|---:|---:|---|
| 0 | 11 | `signed_div(P,2)=375` | `2` | 先夾 `99` |
| 1 | 4 | `signed_div(P,4)=187` | `0` | 先夾 `198`；超過 `99` 時拆出同 ID／同 raw 的下一槽 |
| 2 | 4 | `signed_div(P,4)=187` | `4` | 先夾 `198`；超過 `99` 時拆出同 ID／同 raw 的下一槽 |
| 3 | 4 | `signed_div(P,2)=375` | `2` | 先夾 `99` |

每槽的 `+0x52` 是 ID、`+0x54`／`+0x5B` 是數量及鏡像、`+0x56` 是 raw flags、
`+0x58=-1`、`+0x5A=1`、`+0x55=0x0F`。這證實 remake 保存的四個 ID／raw flag；
`99/198` 是上限／拆槽規則，不應直接稱為任何 runtime 開局必然裝彈量。

### `sub_6EE8E @ 0x6EE8E`：容量 divisor 的完整中間鏈

此函式**回傳 divisor／容量因子**，caller 再用 `quantity=floor(seed/divisor)`；
不能把它直接當作數量。raw 呼叫契約是：`EAX=word_19988E`（防守帝國索引）、
`EDX=weapon ID`、`EBX=乘數`、`CX=mode flags`，stack `[arg_0]` 是第一個參數（要塞
caller 傳 `0`），`[arg_4]` 是第二個參數（raw flags `0/2/4`）。其鏈如下：

1. `K=word_17F811[ID*0x1C]`、`Cat=byte_17F80F[ID*0x1C]`。`Cat==1` 時第一個
   stack 參數的 `2/5/10/15/20` 分別加 `10/20/30/35/40`，要塞傳 0 故此項為 0。
2. `sub_6A636(raw)` 的第一個百分比輸出從 100 開始：raw bit `0x02` 查
   `word_17FD15` row 1 得 `200`，raw bit `0x04` 查 row 2 得 `50`，因此本輪
   `raw 0/2/4` 的第一百分比分別是 `100/200/50`。`sub_6EE8E` 的第二個加成輸入
   由 `sub_6A406(raw)` 提供；對這三個 raw 值都是 `0`，不能把 `sub_6A636` 的
   兩個輸出誤寫成 X 式中的兩個相同百分比。
3. `Cat==0` 時，`CX` 的 bit `0x02` 或 `0x04` 使 `modeExtra=25`，bit `0x10`
   使 `modeExtra=50`，否則 0；要塞 caller 的 `CX=1`，所以為 0。
4. `sub_6A406(raw)` 對 raw `0/2/4` 都回傳 0；它其餘檢查的 mask 是
   `0x0040..0x0080` 與 `0x0100..0x2000`，透過 `sub_8E94D` 找最高 set-bit 的
   row。已讀到的 row 是 `0, +100, -50, +50, +50, +25, +50, +50, +100,
   +100, +25, +25, +25, +300, +50`，但 row 的玩法名稱仍未知。
5. 令 `K1=signed_div(K*percent1,100)`、
   `X=signed_div((100+modeExtra+percent2)*K1,100)`。`B=sub_6F11C(word_17F80D[ID*0x1C])`。
   對 ID 4／11，原始分類 bytes 與特殊 ID 比對使靜態可證 `B=1`；並非把 `B` 命名為
   某個未證實的科技能力。
6. `T=sub_6E70A(word_19988E, word_17F80D[ID*0x1C])`。ID 4 的 required-tech
   raw field 是 `0x7B`，ID 11 是 `0x2F`；若玩家資料的
   `dword_197F98 + empire*0xEA9 + 0x117 + T` byte 等於 1，回傳 0，否則落到
   `sub_6D048`。所以 `T` 仍由執行期科技狀態決定，靜態檔不能指定唯一分支。
7. `sub_6E60E(X,T,B)` 先按 `B` 選 base，再算
   `signed_div(base*X+500,1000)`，正 `X` 的零結果夾成 1；`EBX=1` 的要塞 caller
   最後不再改倍率。`B=1` 時 `T=2→base=700`、`T=3→600`、其他 `T→400`。

因此對本輸入的 ID 4／11、要塞 `CX=1`，可直接重算：

```text
divisor(ID, raw, T) = floor((base(B=1,T) * X(raw) + 500) / 1000)
quantity = floor(seed / divisor)
```

其中 ID 4 的 `K=15`、ID 11 的 `K=30`，所以 `X`（raw 0／2／4）分別為
ID 4：`15/30/7`，ID 11：`30/60/15`（signed integer division）；例如 raw 0 的
ID 4 divisor 在 `T=2/3/other` 為 `11/9/6`，slot 1 的 seed 187 因而是
`17/20/31`（尚未套 198 上限）。raw 4 的相應 divisor 為 `5/4/3`，slot 2 的
seed 187 因而是 `37/46/62`。ID 11 raw 2 的 divisor 為 `42/36/24`，slot 0
seed 375 因而是 `8/10/15`。這些數字是**靜態公式在三個 T bucket 的重算例**，
不是宣稱某個 SAVE 開局的 runtime 量；`T`、`word_19988E` 的合法玩家角色與
`sub_6D048` 的 live BSS 狀態仍需原版執行期才能逐值驗收。

### 本輪結論與 remake 邊界

- Fire 的隨機範圍、命中門檻、插值端點、下游消費呼叫與 raw bit `0x04` 不可達性已是
  **已證實**；完整 `sub_3DF8D` 欄位命名、兩份外部符號哪個才是原版函式名，以及
  runtime seed／每次 target 狀態仍是**強推論／未知**。
- 要塞四槽 ID、seed、raw flags、上限／拆槽、class table 的實際地址與
  `sub_6EE8E` 的 divisor 中間算式已是**已證實**；raw bit 的正式玩法名稱、
  `T=sub_6E70A(...)` 的 live 值與當下數量仍是**未知 runtime 輸入**。
- 因此 remake 可以保留 raw flags、以已證實的戰機下游與要塞直接火力消費運作；不應
  把 `99/198` 當必然 live quantity，也不應把 raw `4` 命名為 ARM/FST 或再加一個
  靜態未證實效果。只有取得可啟動 `VESA.COM`／DOSBox runtime 後，才補逐彈／逐槽
  逐值對照；不再為與玩家行為無關的內部函式開新的挖掘迴圈。

## 2026-08-11 地面戰／事件／爆炸／CMBTSHP／外交尾項勘誤

本節仍使用同一份 `Orion2.exe.i64`（SHA-256
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`）、IDA Pro 9.4／
Hex-Rays 9.4.0.260610、IDA 線性位址與非破壞性 IDC。輸出為 static-only；沒有執行
原版，也沒有把主機缺少 `VESA.COM` 的 DOSBox 嘗試當成 runtime 證據。

本節新增的 CMBTSHP／SABOTAGE／領袖 callback 深度窗口由
[`tools/ida/consumer_closure.idc`](../../tools/ida/consumer_closure.idc) 產生；它只輸出
原始函式名、函式邊界、直接 xref、原始 operand 與 raw bytes，不改寫 `.i64`。前一輪
`oracle_probe.idc` 的歷史結果仍保留，兩者都以 IDA 線性位址為準。

### 地面戰：傷亡、亂數、AI 裝甲與人口欄位

- `raw_0xEC4FE` 對每一個當前地面戰類型各做兩次
  `sub_1247A0(100) + currentType*2 + 2`。`Aroll <= Broll` 時攻方命中，
  `Aroll >= Broll` 時守方命中；相等時兩側都命中。`sub_1247A0` 的 32-bit LCG
  與 rejection sampling 由 `gamedata.OrigRand` 保存 1-based 契約。
- `raw_0xED59D` 讀守方兵力 byte table；殖民地／玩家 raw flag `+0x8BD` 未設時走
  半值，設時走原值，最後至少為 1。`raw_0xED674` 對裝甲營房以人口 `/4` 或 `/2`
  取上限，最後至少為 1；`raw_0xED713`／`0xED260`／`0xECECA` 串起卸載、佔領與
  地面部隊欄位。
- `raw_0xECECA` 直接寫殖民地 `+0x130` 的入侵／守方部隊欄；該靜態鏈沒有證實
  它直接寫人口欄。因此 remake 保留 AI surviving marines/tanks，captured colony
  保留原人口，不用「剩餘部隊數」假造人口回寫。這是**已證實欄位邊界**；原版另一路
  戰後人口 consumer 與原始 global random seed 仍是**未知**。
- `InvadeColony` 已接這個解算器與原版 LCG 的 seed adapter；adapter seed 是 remake
  的可重播映射，不宣稱等同原版 SAVE global seed。抽樣測試鎖定平手雙命中、AI 裝甲營、
  存活兵力回寫與 captured population 保留。

### 事件漂移與候選權重

- `raw_0x201F9` 對 36 個事件記錄槽逐槽檢查；2026-08-25 後續控制流確認只有 36 槽全為
  active 且星曆可被 20 整除時才呼叫 `sub_201A4`，舊「沒有 active record」解讀錯誤。
- `sub_22D57 @ 0x22D57` 以候選帝國總人口 `word [empire+0xA6]` 建池；該欄由
  `sub_E2710 @ 0xE2710` 累加有效殖民地人口 `+0x0A` 後寫入：
  好事件排除目前最高人口，壞事件排除目前最低人口，其餘權重為與該極值差值的平方；
  `sub_230B6` 負責其他 eligibility。
- `sub_586D4` 先加總正權重；只要總和 `>=0x200`，每輪將所有正權重整除 2，
  再用 `Random(total)` 做累積抽樣。`OriginalEventWeightedChoice` 與
  `OriginalEventVictimWeights` 已測試這兩段純公式。
- `sub_21371` 是 36-case event dispatcher，使用 `byte_19ABA5[eax*9]` 與多個 raw
  event record 欄位；remake 仍沒有完整 record lifecycle 與 AI／熱座目標效果回寫。
- 2026-08-25 由完整 `sub_2230A` 控制流訂正：`1/2, 2/3, 3/4, 4/5, 5/6` 是一般事件
  排程依難度套用於「距上次事件日期」的門檻，再與 `Random(0x200)` 比較，不是僅限艦隊
  強度的下游。Go 已接此排程；證據見 `random-event-schedule-audit-20260825.md`。

### 爆炸連鎖與下游消費

- `raw_0x3868F` 逐 `si=1..word_1998C0-1` 掃描；先以 `Random(100) <=
  word_180EC8[type]` gating。艦隊路徑取 `Random(0xC9)+0x4A`，清暫存 type 後呼叫
  `sub_40C2A`；殖民地路徑取 `Random(0x191)+0x95`，每步減 `0x14`，呼叫
  `sub_39985`，再把 `sub_494A8` 的 engine potential 寫入四個方向欄，並可能寫
  colony status 5。
- `sub_40C2A @ 0x40C2A` 對 target type 7 取 `damage/4`；其他 target 扣除
  `word_17F6C1[type]` 並夾至 0，再累加 damage `+0x1F`、上限 `30000`，達 target
  `+0x1D` 時寫 status 1。
- `sub_494A8 @ 0x494A8` 的 base 是
  `word_17F6C1[size] * 5 * (maxEngineLevel+1)`；size>=5 改用 `+2`。raw
  `0x14` 分支再做 `3*damage/2`。`OriginalEngineExplosionBasePotential` 與
  `OriginalEngineExplosionRaw14Branch` 已保存此純公式。
- `sub_39985 @ 0x39985` 的 `sub_4B0D3` 多個 flag／tech gates、`sub_3A3C3` 行星
  damage table、fleet／colony record 與 `sub_36810` 目標選擇尚未全部命名。已追回的
  `Random(201)+74`、`Random(401)+149`、每步 `-20`、type 7 四分之一與 resistance
  consumer 已進 `original_explosion_oracle.go`；沒有把 strategic `Ship.Damage`
  當成原版同一欄位，避免錯誤連鎖接線。

### CMBTSHP 圖片與動畫

- `sub_30062 @ 0x30062` 讀標準艦 raw picture，精確索引為
  `45*player_color + raw_ship_picture`；色塊內 raw picture `0..43` 合法，`44`
  (`0x2C`) 是 MONSTER／palette-holder sentinel，不是標準艦 sprite。
- LBX 靜態解碼確認每個標準艦資產有 20 幀。`sub_3F5F1 @ 0x3F5F1` 回傳 0..15
  heading；`sub_3F628 @ 0x3F628` 讀 record `+0x23`，依最短方向每次 `+1/-1`
  並以 16 取餘後重新載圖。
- remake 的圖片映射已精確接到 `CMBTSHPSpriteIndex` 與 `.GAM` raw picture；
  `CMBTSHPFrameForHeading` 是顯示 adapter。`sub_31F25 @ 0x31F25` 的繪圖迴圈會
  呼叫輸入／重繪路徑與 `sub_30631`，但 `sub_30062`／`sub_30631`／`sub_31F25`／
  `sub_3F628` 的深度窗口沒有找到 frame counter 寫入、clock／timer 呼叫或中間幀停留
  常數。因此動畫時序仍是**未知**，不能把最近角度換算稱為原版 timer；這部分依使用者
  決策採可重播固定 tick 近似。

### 特殊貿易、SABOTAGE 與活動領袖 ETA

- `sub_101BA4 @ 0x101BA4` 的政府／`+0x8B7` 神級商人段為 `100/150/175` 加
  `+50`；隨後掃描 `0x3B×0x43` 領袖，僅納入同帝國且 raw status `<3`，raw
  `+0x28` 的 Trader tier 1/2 對應 `(bucket+1)*10/*15`。`sub_94951 @ 0x94951`
  讀 `+0x24` 經驗並使用 `sub_93D4B @ 0x93D4B` 的 60/150/300/500/1000 分桶，
  Warlord `+0x8BD` 在最高桶可給 5；ID `0x42` 忽略 Warlord。這段已接入
  `OriginalSpecialTradeLeaderBonus`、GAM raw experience 與 Treaty target。
- `Steal_App @ 0x10130A` 掃殖民地 49 槽（stride `0x13`），已建旗標在 `+0x136`，
  skip slot 9，權重為 `off_17EB3D + 8` 生產成本，累積 total 後 `Random(total)`，
  再由 `Add_Building @ 0x145EA` 清除所選槽。49 筆 raw table、任務 70 門檻與
  `sub_101483 @ 0x101483` slot helper 已接入／對齊；`sub_1014A4 @ 0x1014A4` 另已
  證實低 6 位 count、高 2 位 mode、兩張 score table 的讀取位置、亂數位置與
  60／70／80／90／±80 分支。兩張 table 的 raw bytes 是 `0xFF` 初始化值，上游填值、
  完整 score 欄位、toggle 邊界、特殊／未知槽政策與 AI 防守 Agent 的原版 raw policy
  仍未知；remake 以可保存 AB／DB／E 近似完成，不升格為 raw parity。
- `Deassign_Officer @ 0x934CF` 對 status 4 的 `+0x37` 遞增至 30，對 status 1
  的 `+0x37` 遞減，歸零且 location 1 時呼叫 `sub_E2AB1`。`sub_E2AB1` 掃描六個
  `+0x48..+0x52` raw 槽，符合 empire／active 條件時依序呼叫 `sub_E1D59`、
  `sub_DF8F0`，最後呼叫 `sub_E2710` 重算帝國彙總欄位；這條完整 callback 鏈已由 IDA
  證實，但 raw 設計／艦隊欄位沒有安全的一對一 remake 模型。`Get_Ship_Leader_ETA
  @ 0x98F42` 的一般 fallback 是 5，暫時池另讀 raw ETA；remake 以既有可撤銷領袖增量、
  全殖民地士氣刷新近似完成玩家可感知 callback。

本節對應程式／測試與完整表格索引見
[`docs/re/special-trade-sabotage-leader-eta-20260811.md`](special-trade-sabotage-leader-eta-20260811.md)。

### 2026-08-11 remake 消費端收尾（允許近似）

依使用者本輪決策「remake 完成為主」，上述未知不再阻塞可玩流程，已接入以下明示為
近似的 consumer：

- CMBTSHP：戰術畫面在移動輸入後使用固定 tick 的 `CMBTSHPFrameAtTick` 播放短掃掠，
  停止後固定在最近 heading 幀；不把這組 tick 宣稱成原版 timer。
- 事件／爆炸已拆分：`sub_3868F`、`sub_39985`、`sub_40C2A` 是戰鬥／殖民地爆炸鏈；
  事件 8 的 `sub_206A2 @ 0x20AD7..0x20B61` 不呼叫它們，而是由 `sub_23CED`
  reservoir 選一艘、`sub_941C6` 處理死亡軍官、`sub_A163A` 移除單艦。remake 已依後者接線，
  證據見 `random-event-ship-explosion-audit-20260825.md`。
- SABOTAGE：`spyMissionScore` 明列 remake 可用的 AB／DB 每一項來源，使用 `T=70`、
  `E=T+DB-AB`、`SpyRollChance(E)`；低 6 位 raw count 的三段式 helper 已與
  `gamedata.OriginalSpyScoreHelper` 對齊，防守 Agent 的訓練、上限與被 Spy-vs-Spy
  擊殺後的真實數量扣除已接。原版兩張 score table 的上游填值／未命名欄位仍是 oracle
  差異，不能把近似模型誤寫成 raw parity。
- 領袖 ETA：raw `1→0` 且 location=1 時呼叫 `applyLeaderETACallback`，刷新已指派
  殖民地的領袖增量與所有殖民地士氣；不解雇領袖。這是 `sub_E2AB1`／
  `sub_E1D59`／`sub_DF8F0`／`sub_E2710` 的玩家可感知 remake 近似，不宣稱 raw
  設計／帝國欄位逐值 callback parity。

對應程式、測試、Docker 抽樣與尺度說明見
[`remake-consumer-closure-20260811.md`](remake-consumer-closure-20260811.md)。
