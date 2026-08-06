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
記錄 = [offset:4 LE][seg:2 LE][module:2 LE][type:1][namelen:1][name:namelen]
```

**位址欄在名字之前**。逐位元組驗證(`Generate_Climate_` 那筆):

```
… ab be 07 00 | 01 00 | 36 00 | 04 | 11 | "Generate_Climate_"
   offset       seg      module  type namelen=17 = 名字長度 ✓
```

- `type`:2 = 資料,4 = 函式(另有 3、5)
- `offset`:**映像相對位址**(非模組相對——逐模組檢查證實 offset 單調遞增:
  module 1 = 0x49B–0xCA6 → module 361 = 0x1506F8–0x1522BF)
- `module`:來源模組(.c 檔)索引,可用來把同一子系統的函式分群

### 位址映射

IDA 載入的段基址:`cseg01` org **0x10000**(程式碼+初始化資料)、`dseg02` org **0x178000**。

```
EA = 0x10000  + offset   (seg == 1)
EA = 0x178000 + offset   (seg == 2)
```

## ⚠⚠ 方法論教訓:統計正對照「通過了」,結論仍然全錯

第一版 parser 把格式讀成 `[type][len][name][offset][seg][module]`(位址在名字**之後**)。
這樣讀完一筆名字後抓到的 4 bytes 其實是**下一筆**的 offset,於是
**整份 8,589 個符號的 name↔addr 全部錯開一格**。

這個錯誤通過了當時做的兩道驗證:

| 當時的驗證 | 結果 | 為什麼騙過去 |
|---|---|---|
| 統計:4,013 個 type-4 符號落在 IDA 函式起點的比率 | 85.5%(對照組 0%) | **相鄰符號多半也是函式起點**。整體位移一格,命中率幾乎不變 |
| 語意:`Draw_Diplomacy_Screen_` 反編後呼叫外交動畫繪製 | 「吻合」 | 同一 `.c` 模組的函式在映像中相鄰,**錯開一格拿到的還是外交模組的函式** |

**兩道驗證都只證明了「位址落在合理的鄰域」,沒有證明「這個名字配這個位址」。**
偏移一格的系統性錯誤恰好是這類驗證的盲區。

### 真正戳破它的方法:內容能自我證明的錨點

拿一個**數值內容已由外部獨立來源確定**的表,看誰在讀它:

- `sub_8BFA3` 讀 `byte_17D81C[…]`,而 0x17D81C 起 10 bytes = `0 0 0 1 1 2 2 1 2 3`
  = 手冊 p.59 十個氣候的每農夫食物(Toxic0/Radiated0/Barren0/Desert1/Tundra1/Ocean2/Swamp2/Arid1/Terran2/Gaia3)。
  → 該函式必然是 `Generate_Food_Per_Farmer_`,而舊表把 0x8BFA3 標成 `Generate_Climate_`(前一格)。
- `sub_8BFE0` 讀 `byte_17D72A[5*size + x]`(5×5 表,值 0/1/2)→ 必然是重力等級(Low/Normal/Heavy),
  即 `Generate_Gravity_Class_` 讀 `_gravity_table`。舊表把該位址標成 `Generate_Food_Per_Farmer_`、
  把該表標成 `_class_to_mineral`(都是前一格)。

三個錨點各自獨立,同時指向「往前錯一格」,再回頭逐位元組看記錄佈局就找到了根因。

**[HARD] 紀律:驗證 name↔addr 這種「配對是否正確」的問題,統計命中率與「語意看起來合理」
都不算證據。必須找到一個內容能由外部來源獨立確定的錨點(已知數值表、已知字串),
用「誰在讀它」把名字釘死。**
(這也是為什麼先前 `Missile_Speed_` 反編出「視埠邊界檢查」、`Colony_Research_Per_Scientist_`
反編出「memset」——不是函式邊界壞掉,是**根本反編錯了函式**。)

## 產出(修正後)

| 項目 | 數量 |
|---|---|
| 解析出的符號 | 8,589 |
| type 4(函式) | 4,201 |
| type 2(資料) | 2,022 |
| type 3 / 5 | 928 / 1,438 |
| 成功寫回 IDA 資料庫 | 6,391 |

## Hex-Rays:32-bit 可用(修正舊斷言)

`~/.claude/knowledge-base/retro/ida-pro-9.4.md` 記載「Hex-Rays 不支援」——那是 **16-bit real mode**
的結論。本專案實測:

- **IDAPython 仍不可用**(`-S*.py` exit 1 無輸出,32-bit 也一樣)→ 腳本一律寫 IDC。
- **Hex-Rays 可用,走批次 `-O` 參數**(不需 Python):
  ```
  idat -A "-Ohexrays:/work/out.c:<函式名>" Orion2.exe.i64
  ```
  輸出開頭會標 `Detected compiler: Watcom C++`。`:ALL` 在此版產出空檔,只能逐一指名。

## ⚠ 反編結果出現 `JUMPOUT` 時

Watcom 產生的碼有共用尾碼/跨函式跳轉,IDA 有時把函式邊界劃錯,症狀是反編輸出裡出現
**`JUMPOUT(0x…)`** 或 `control flows out of bounds`。這種輸出不能直接當權威答案;
處理方式:①改讀原始組語(`.asm` 裡的絕對位址是硬的)②`set_func_end` 修邊界後重反編
③找呼叫端確認語意。

> 註:先前記在這裡的兩個「實例」(`Missile_Speed_` 反編出視埠檢查、
> `Colony_Research_Per_Scientist_` 反編出 memset)**歸因錯誤**——真正的原因是符號表
> parser 錯開一格,反編到了別的函式。已撤掉,免得下次照著去修邊界卻修錯地方。

## 資料表:先查 xref 再引用數值

名字對上位址之後,語意仍要靠 xref 釘死——同一張表可能被多處讀,索引方式決定它的維度。
實例:`_gravity_table` 名字看起來是一維查表,實際被 `Generate_Gravity_Class_` 用
`table[5*size + density]` 索引,是 **5×5 二維表**。只看名字會把維度弄錯。

### 已釘死的權威數值表(dseg02,EXE 內靜態)

以下全部經「內容 + 讀它的函式」雙重確認,可直接當 remake 的數值來源:

| 表 | 位址 | 長度 | 內容 | 語意 |
|---|---|---|---|---|
| `_gravity_table` | 0x17D72A | 25 | `0 0 0 1 1 / 0 0 1 1 1 / 0 1 1 1 2 / 1 1 1 2 2 / 1 1 2 2 2` | `[size][density]` → 重力等級 0/1/2 |
| `_climate_roll_table` | 0x17D77F | 40 | (見 gap report) | 氣候骰表 |
| `_normal_gal_climate_roll_table` | 0x17D7A7 | 40 | 同上 | 一般星系用 |
| `_old_gal_climate_roll_table` | 0x17D7CF | 40 | 同上 | 「古老星系」用 |
| `_planet_size_table` | 0x17D7F7 | 5 | `1 3 7 9 10` | d10 **累計**骰表 → 行星大小(Tiny 10%/Small 20%/Medium 40%/Large 20%/Huge 10%) |
| `_planet_max_farms` | 0x17D7FC | 5 | `2 4 5 7 10` | 各行星大小的農場上限 |
| `_planet_max_mines` | 0x17D801 | 5 | `2 4 6 9 12` | 各行星大小的礦場上限 |
| `_planet_max_population` | 0x17D806 | 5 | `5 10 15 20 25` | 各行星大小的基礎人口上限 |
| `_spectral_class_table` | 0x17D80B | 7 | `10 20 30 45 92 96 100` | d100 **累計**骰表 → 恆星光譜(藍10%/白10%/黃10%/橙15%/紅47%/棕矮4%/黑洞4%) |
| `_minerals_extracted_table` | 0x17D812 | 5 | `1 2 3 4 5` | 礦產豐度 → 開採等級 |
| `_climate_modifier_table` | 0x17D817 | 5 | `0 0 0 -10 -20` | 氣候修正(負值 = 懲罰) |
| `_food_per_farmer_table` | 0x17D81C | 10 | `0 0 0 1 1 2 2 1 2 3` | 十氣候每農夫食物,**與手冊 p.59 完全一致** |
| `_planet_special` | 0x17D826 | 12 | `64 69 71 72 75 77 87 90 93 97 101 101` | 特殊物產累計骰表 |
| `_planet_special_weighted_chance` | 0x17D832 | 12 | `64 5 3 2 3 2 9 4 3 0 5 0` | 同上的權重版(**總和 = 100**) |
| `_star_class_table` | 0x17D83E | 21 | `20 10 5 / 25 15 5 / 10 16 30 / 10 16 21 / 32 37 30 / 1 2 3 / 2 4 6` | 7 光譜 × 3 |
| `_ranged_to_hit_penalty` | 0x17D855 | 18 | words `0 0 -10 -20 -30 -40 -55 -70 -85` | 射程命中懲罰(%) |
| `_ranged_damage_penalty` | 0x17D867 | 18 | words `0 0 -10 -20 -30 -40 -50 -60 -65` | 射程傷害懲罰(%) |
| `_minerals_per_mine` | cseg01 0xDD4B5 | 5 | `1 2 3 5 8` | 礦產豐度五級每礦工工業,**與手冊一致** |

`_planet_max_population = 5 10 15 20 25` 獨立證實了 remake 既有的 `(size+1)*5` 公式;
`_food_per_farmer_table` 與 `_minerals_per_mine` 同樣與手冊逐格相符——
**手冊在這三項上沒有簡化或筆誤**,先前對它們的存疑可以撤銷。

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
| **`parse_syms2.py`** | **解析符號表 → TSV(格式正確版,一律用這支)** |
| ~~`parse_syms.py`~~ | ⚠ 格式讀反,產出的 name↔addr 全部錯開一格。留著只為記錄教訓,不要再用 |
| `apply_names.idc` | 把符號寫回 IDA 資料庫 |
| `dump_ea.idc` | 依位址 dump bytes(hex + signed + ASCII 三種解讀並列) |
| `xref_ea.idc` | 查某位址的 xref |
| `make_funcs.idc` | 在函式符號位址強制建函式(補 IDA 沒追到的) |

### 踩過的坑(下次直接避開)

1. **被中斷的 IDA 會留下未打包的 `.id0/.id1/.nam/.til`**,之後任何操作都靜默產出空檔。
   症狀跟「工具壞了」一樣——用**另一份 `.i64` 跑同一指令**做正對照即可分辨(kb 已記載,本次再次命中)。
   處置:刪掉那四個暫存檔,`.i64` 本身通常沒事。每次批次跑完順手清。
2. `idat` 對 `.i64` 一次只能一個行程,背景跑全量反編時不要同時跑別的查詢。
3. Hex-Rays batch 的函式名要**帶尾底線**(`Generate_Climate_`),與符號表原名一致;
   少一個底線就 rc=1 且不產檔,看起來像工具壞了。
4. **`.asm` 全檔匯出很值得留著**:它用 `sub_XXXXX` / `byte_XXXXX` 標絕對位址,
   反編輸出反而把位址藏在變數名後面。要確認「這段碼到底讀哪個位址」時,`.asm` 是最快的硬證據。
