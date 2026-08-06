# Orion2.exe 內建除錯符號表:解析、映射與 IDA 符號化

> 日期:2026-08-06。輸入檔:`Orion2.exe`(patch 1.31 安裝目錄),
> SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`,2,644,842 bytes。
> 工作區在 **repo 外** `/home/anr2/moo2-private-build/re/`(版權資產不入 repo,[HARD]);
> `.exe` / `.i64` / `.asm` / 全量符號 TSV 全部只留在該處,本文件只記方法與結論。

## 為什麼這件事重要

先前多個 gameplay 缺口被標成「需 DOSBox 實機 oracle」(飛彈速度、地面戰解算結構、母星開局態、
安塔蘭防禦戰力)。實際上 **原版執行檔自帶完整除錯符號**,這些機制在二進位裡都有具名函式。
依 `rulebook/62` 的 oracle 優先序:**反編出來的機器碼 > 社群攻略/手冊反推**——
這是本專案目前能取得的最強 oracle,勝過 openorion2(純渲染殼)與手冊。

## 執行檔型態(先前未記載)

| 檔案 | 型態 | 可反編 |
|---|---|---|
| `Orion2.exe` | **DOS/4GW LE**,32-bit protected mode,Watcom C++ | ✅ 有符號 + Hex-Rays 可用 |
| `ORION95.EXE` | PE32(Win95 版) | 可反編但**無符號** |
| `SETSOUND.EXE` | DOS/4GW LE | 音效設定,非遊戲邏輯 |

## 符號表格式(逐位元組驗證)

IDA 9.4 載入 LE 主體,但**不解析**這個符號表(資料庫裡全是 `sub_`/`unk_`)。自行解析:

```
記錄 = [type:1][namelen:1][name:namelen][offset:4 LE][seg:2 LE][module:2 LE]
```

- `type`:2 = 資料,4 = 函式(另有 3、5)
- `offset`:**映像相對位址**(非模組相對——逐模組檢查證實 offset 單調遞增:
  module 1 = 0x49B–0xCA6 → module 361 = 0x1506F8–0x1522BF)
- `module`:來源模組(.c 檔)索引,可用來把同一子系統的函式分群

### 位址映射(統計正對照驗證)

IDA 載入的段基址:`cseg01` org **0x10000**(程式碼+初始化資料)、`dseg02` org 0x178000。

```
EA = 0x10000 + offset        (seg == 1)
```

**驗證方法**(不是憑印象宣稱):抽出 IDA 全部 5,135 個函式起點,計算 4,013 個 type-4 符號
在各假說下的命中率——

| 假說 | 命中函式起點 |
|---|---|
| `+0x10000` | **3,430 / 4,013 = 85.5%** |
| `+0x178000` | 0 / 4,013 = 0.0% |
| `+0`(不加) | 20 / 4,013 = 0.5% |

85.5% vs 0% 是決定性差距(隨機命中率趨近 0)。未命中的 14.5% 是「IDA 沒建成函式」
(只透過跳表/間接呼叫抵達),已用 `add_func()` 補建 332 個。

### 名稱↔位址配對的獨立驗證

映射對不代表**名字掛對位址**。用語意不容誤認的函式反證:
`Draw_Diplomacy_Screen_` 反編後呼叫 `Draw_Multi_File_Animation_Stencil` / `Draw_Multi_File_Animation`
——與專案已知「DIPLOMAT.LBX 是多幀 delta 使節動畫」完全吻合,配對成立。

## 產出

| 項目 | 數量 |
|---|---|
| 解析出的符號 | 8,590 |
| 其中函式(type 4, seg 1) | 4,013 |
| 成功寫回 IDA 資料庫 | 6,375 |
| 補建函式 | 332 |

## Hex-Rays:32-bit 可用(修正舊斷言)

`~/.claude/knowledge-base/retro/ida-pro-9.4.md` 記載「Hex-Rays 不支援」——那是 **16-bit real mode**
的結論。本專案實測:

- **IDAPython 仍不可用**(`-S*.py` exit 1 無輸出,32-bit 也一樣)→ 腳本一律寫 IDC。
- **Hex-Rays 可用,走批次 `-O` 參數**(不需 Python):
  ```
  idat -A "-Ohexrays:/work/out.c:<函式名>" Orion2.exe.i64
  ```
  輸出開頭會標 `Detected compiler: Watcom C++`。`:ALL` 在此版產出空檔,只能逐一指名。

## ⚠ 已知限制:函式邊界錯誤 → 部分反編結果不可信

Watcom 產生的碼有共用尾碼/跨函式跳轉,IDA(尤其我強制補建的那 332 個)常把邊界劃錯。
症狀是反編輸出裡出現 **`JUMPOUT(0x…)`** 或 `control flows out of bounds`。

實例:`Colony_Research_Per_Scientist_`(0xDFE77)反編出的函式體帶 `JUMPOUT(0xDEE14)`,
且內容看起來是快取/memset 而非研究計算——**該結果不可採信**,需要先修正函式邊界。

**[HARD] 紀律:反編結果若含 `JUMPOUT` / 邊界警告,不得直接當作該機制的權威答案。**
處理方式:①改讀原始組語 ②手動修邊界(`set_func_end`)後重反編 ③找真正的呼叫端確認語意。
反例警示:`Missile_Speed_` 反編出來是「座標 /20 後檢查是否落在 32×18 視窗內」的可見性判斷,
與「速度」無關——在修正邊界前不得據此下任何關於飛彈速度的結論。

## ⚠⚠ 資料符號:名稱↔位址配對**不可直接採信**(2026-08-06 實測)

程式碼符號的配對已用兩種方式驗過(統計 85.5% + `Draw_Diplomacy_Screen_` 語意反證),但
**資料符號不同**——實測抓到明確錯配:

| 符號名 | 位址 | 內容 | 誰讀它(xref) | 判定 |
|---|---|---|---|---|
| `_planet_size_table` | 0x17D7FC | `2 4 5 7 10` | `Generate_Gravity_Class_`、`Enforce_Gravity_`、`Init_Non_Homeworld_Colony_` | ✅ 語意成立(MOO2 重力由行星大小決定) |
| `_gravity_table` | 0x17D743 | `0 0 1 0 0 1 …`(全 0/1/2) | **無人讀** | ⚠ 值像重力等級但零 xref |
| `_food_per_farmer_table` | 0x17D826 | `64 69 71 72 75 77 87 90 93 97 101 101`(單調遞增到 101) | `Get_A_Planet_Or_Star_Special_` | ❌ **名字掛錯**:這是 roll 1–100 的**累積骰表**,不是每農夫食物 |

**[HARD] 紀律:引用任何資料表的數值前,先查它的 xref,用「誰在讀它」確認語意。**
名字只是線索,`Get_A_Planet_Or_Star_Special_` 讀一張叫 food_per_farmer 的表 = 名字錯了,不是機制怪。
(同 `rulebook/62`「描述一段碼做什麼前先 grep 呼叫端」,資料表版。)

## 資料表有兩類:EXE 靜態 vs 執行期載入(BSS)

dump 時位址回報「未載入」= 該符號落在 **BSS**(未初始化區),代表**內容是執行期從 LBX 資料檔載入的**,
EXE 裡沒有靜態值。實測:`_ability_costs`(0x19B714)、各 `*_msg` 字串(0x1AA2xx)皆屬此類。

→ **對 remake 是好消息**:這些值在 LBX 檔裡,而專案已有完整 LBX 解碼器,不必靠反編。
→ 判別法:dump 回「未載入」就去 LBX 找,別在 EXE 裡繼續挖。

## 已可回答的問題(第一批)

### 原版畫面清單(直接對應「畫面差異」)

從符號表取得原版**完整畫面函式清單**(380 個 screen/window 相關符號),其中子畫面結構
硬證實了 2026-07-12 用 archive.org 截圖推得的結論:

```
Draw_History_Subscreen_        ← INFO 的「歷史圖表」是子畫面
Draw_Tech_Review_Subscreen_    ← 「科技總覽」是子畫面,不是跳去研究選擇
Draw_Race_Stats_Subscreen_
Draw_Turn_Summary_Subscreen_
Draw_Reference_Main_Subscreen_ / _Category_ / _How_To_
```

即 **INFO = 單一畫面 + 5 個子畫面**,remake 目前把 Tech Review 誤接成跳板(issue #5-2)
的判定由二進位證實,不再只是截圖推論。

其他原版畫面(remake 尚未建或簡化):
`Draw_Colony_Bombing_Screen_`、`Draw_Colony_Landing_Screen_`、`Draw_Colony_Combat_Screen_`、
`Draw_Hall_Of_Fame_Screen_`、`Draw_Hi_Score_Screen_`、`Draw_Main_Antaran_Room_Screen_`、
`Draw_Event_Screen_`、`Draw_Flag_Screen_`、`Draw_Racial_Option_Screen_`、`Draw_Mini_Main_Screen_`、
以及整組多人連線畫面(`Draw_Join_Net_Screen_`、`Draw_MP_Setup_Screen_`、`Draw_Hotseat_Screen_`…)。

### 各開放項對應的原版函式(待逐一反編驗證)

| 開放項 | 原版符號 |
|---|---|
| 母星開局態 | `Twiddle_Initial_Homeworld_`、`Init_Colony_`、`Init_Homeworld_Colony_`、`Init_Adv_Civ_Homeworlds_` |
| 人口產出公式 | `Colony_Food2_Per_Farmer_`、`Colony_Industry_Per_Worker_`、`Colony_Research_Per_Scientist_`、`_food_per_farmer_table` |
| 地面戰解算 | `Resolve_Ground_Combat_`、`Ground_Combat_Round_`、`Compute_Ground_Combat_Info_`、`Compute_Player_Ground_Combat_Bonuses_` |
| 指揮點數 | `Show_Command_Points_`、`_starting_command_points_msg` |
| 安塔蘭 | `Antaran_Invasion_Check_`、`_chance_for_antaran_invasion` |

## 工具

全部在 `/home/anr2/moo2-private-build/re/tools/`(repo 外):

| 檔案 | 用途 |
|---|---|
| `ida.sh` | docker headless wrapper(`analyze` / `idc` / `raw`) |
| `parse_syms.py` | 解析符號表 → TSV |
| `apply_names.idc` | 把符號寫回 IDA 資料庫 |
| `make_funcs.idc` | 在函式符號位址強制建函式(補 IDA 沒追到的) |

### 踩過的坑(下次直接避開)

1. **被中斷的 IDA 會留下未打包的 `.id0/.id1/.nam/.til`**,之後任何操作都靜默產出空檔。
   症狀跟「工具壞了」一樣——用**另一份 `.i64` 跑同一指令**做正對照即可分辨(kb 已記載,本次再次命中)。
   處置:刪掉那四個暫存檔,`.i64` 本身通常沒事。每次批次跑完順手清。
2. `idat` 對 `.i64` 一次只能一個行程,背景跑全量反編時不要同時跑別的查詢。
