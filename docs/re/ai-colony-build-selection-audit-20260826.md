# AI 殖民地建造選擇靜態稽核

日期：2026-08-26

## 證據契約

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4，處理器 `metapc`
- 位址基準：IDA linear，DOS/4GW LE image
- 非破壞性匯出：`tools/ida/audit_ai_colony_build.py`

正式 `.i64` 唯讀掛載後複製到一次性容器的 `/tmp`；探針只匯出原始函式邊界、名稱、
指令、bytes、caller／callee 與資料參照，沒有保存資料庫或覆蓋原始名稱。

## 已證實

### 回合呼叫鏈是逐殖民地，而非全帝國單一造艦池

- `Next_Turn_Calc_ @ 0x136B3` 於 `0x13715` 呼叫 raw `sub_D6F67 @ 0xD6F67`
  （外部索引：`All_Colony_AI_`）。
- `sub_D6F67 @ 0xD6F67..0xD6FDA` 逐 AI 呼叫 raw `sub_D6ED4 @ 0xD6ED4`
  （`Colony_AI_`）。
- `sub_D6ED4 @ 0xD6ED4..0xD6F67` 先收集殖民地、更新科技／艦型，再於 `0xD6F4B`
  呼叫 raw `sub_D0D2F`，於 `0xD6F50` 呼叫 raw `sub_D6E1D`。
- `sub_D6E1D @ 0xD6E1D..0xD6ED4` 分別處理封鎖／未封鎖殖民地，於 `0xD6E75`
  呼叫 raw `sub_D10EE @ 0xD10EE`，接著逐殖民地呼叫 raw `sub_D2754 @ 0xD2754`。

因此原版產品與進度的權威單位是殖民地。remake 原先直接把 `EmpireOutput.TotalNetIndustry`
全數投入 `ShipBuildProgress`，沒有殖民地產品選擇，資料形狀已被直接反證。

### 空產品的建築候選與加權門

raw `sub_D2754 @ 0xD2754..0xD2783` 只在殖民地有效、有人口，且目前產品屬特定空／特殊
範圍時呼叫 raw `sub_D0B08 @ 0xD0B08`。後者：

1. `0xD0B32..0xD0B53` 掃 raw building ID `1..48`（上界 `0x31` 不含）。
2. 每項先呼叫 `Colony_Can_Build_Product_ @ 0xE11BC`；不可建者權重為 0。
3. 可建者呼叫 raw `sub_D0036 @ 0xD0036` 計分；最大分數至少夾到 1。
4. `0xD0BC5..0xD0BD9` 將 `(6-difficulty)×score < maxScore` 的候選歸零。
5. `0xD0C1B..0xD0C27` 以 51 槽權重陣列呼叫 raw `sub_FE92D` 抽選；結果 1..48
   透過 raw `sub_B206F` 轉成產品，49／50 分別寫入 `0xFFFD／0xFFFE`。

這證實了候選範圍、可建 gate、難度濾門與加權抽選控制流。

## 強推論、近似與未知

- **強推論**：`colony+0x115` 是目前產品欄；它被 `Colony_Product_Cost_`、
  `Apply_Production_` 與上述 AI 指派鏈共同讀寫。
- **已標註近似**：remake 已有 typed 建築分類、科技 gate 與殖民地輸出，但尚未一對一表示
  `sub_D0036` 使用的全部 raw 玩家／殖民地欄位；本輪只以 typed 類別與糧食／人口狀態重建分數。
- **已標註近似**：原版建築候選加權抽選的共用 PRNG 精確全局位置未閉合；remake 以回合與
  殖民地 key 決定性取樣，保存／讀取不會換產品。`sub_D0036` 內先前誤稱「國庫亂數」的
  `sub_134C92` 已由完整函式解碼證實是整數平方根，不屬這項 PRNG 留白。
- **未知**：`sub_D10EE` 中完整的帝國級建築、支援艦、戰鬥艦、改裝與購買配額；不可把
  本輪建築子鏈誤寫成完整原版 AI 生產 parity。

## 逐 case 跳表與第一批精確分數

IDA raw `jmp cs:jpt_D01BF[edx] @ 0xD01BF` 的資料表位於 `0xCFF62`，共 47 筆
little-endian dword；索引在 `0xD019F` 先減 1，因此 entry 0 對 raw building ID 1。
探針逐筆保存 entry 位址、四個原始 bytes 與 target，不依賴反編譯器的 case 排版。

第一批可由 remake 現有 typed 欄位完整表示的 case：

| raw ID | 建築 | table entry → target | 原始指令式 | typed 公式 | 等級 |
|---:|---|---|---|---|---|
| 4 | 太空大學 | `0xCFF6E a3060d00 → 0xD06A3` | `mov ebx,5` | `5` | 已證實 |
| 7 | 自動工廠 | `0xCFF7A d0060d00 → 0xD06D0` | `pop+13+2×var_2C` | `population + 13 + 2×[Honorable]` | 已證實 |
| 12 | 深層核心礦場 | `0xCFF8E 39070d00 → 0xD0739` | `pop+12+4×var_2C` | `population + 12 + 4×[Honorable]` | 已證實 |
| 34 | 機器人工廠 | `0xCFFE6 18090d00 → 0xD0918` | `12+2×var_2C` | `12 + 2×[Honorable]` | 已證實 |
| 36 | 機器人採礦廠 | `0xCFFEE 47090d00 → 0xD0947` | `pop+5+2×var_2C` | `population + 5 + 2×[Honorable]` | 已證實 |

`var_2C` 由 `0xD007A..0xD0082` 的 `player+0x28 == 4` 建立。專案既有
`ai.Personality` 已以 AIRACES／官方字串把 raw 4 對到 `PersonalityHonorable`，所以這五式
不需要猜測新的欄位語意。五條路徑最後都直接進共同正值／1000 上限，不會經
`0xD0414 add ebx,var_18` 的國庫／淨收入平方根因子。

## 第二批：四棟研究設施與共用優先建築 gate

同一份非破壞性探針現會全資料庫掃描原始運算元中的 `player+0x59D`，並匯出函式、原始名稱、
指令位址、bytes 與運算元。直接定位結果共 12 筆；唯一直接寫入是：

- `sub_DC288 @ 0xDC2E4..0xDC2EF`：映射後的研究 field 以 `cmp al,4Bh` 檢查，raw field
  `>=75` 時在 `0xDC2E8` 以 `C6 83 9D 05 00 00 01` 寫 `player+0x59D = 1`。
- 掃描沒有找到對 `player+0x59D` 的直接清零。整個 player record 的初始化仍可能間接清零，
  因此結論限定為「遊戲建立後一旦設立，未見直接撤銷端」，不把直接運算元掃描冒稱完整間接寫入證明。
- raw 75..82 與 typed `TOPIC_HYPER_*` 一對一；remake 可由目前 Hyper field、已完成 Hyper field
  或已保存的 Hyper level 重建這個持續狀態，不需新增猜測欄位。

探針另對共用 `ah` gate 的六個 raw offset 匯出全資料庫直接運算元與同函式上下文。
`sub_D58D4` 的成對科技／建築檢查提供可重算的雙基址對照：

| 科技 offset | `offset-0x117` | 科技 | 建築 offset | `offset-0x136` | 建築 |
|---:|---:|---|---:|---:|---|
| `player+0x12D` | 22 | Automated Factories | `colony+0x13D` | 7 | Automated Factory |
| `player+0x17E` | 103 | Marine Barracks | `colony+0x14C` | 22 | Marine Barracks |
| `player+0x125` | 14 | Armor Barracks | `colony+0x138` | 2 | Armor Barracks |

三個科技 ID 與三個建築 ID 均逐項吻合受版控 enum；`sub_D58D4` 還有
`player+0x1BB - 0x117 = TECH_SPACEPORT(164)` 對
`colony+0x15D - 0x136 = BUILDING_SPACEPORT(39)` 的獨立同形交叉驗證。因此這不是由名稱猜測 offset。

`sub_D0036 @ 0xD010D..0xD019A` 形成 `ah`：

1. 殖民地沒有 Automated Factory、帝國已知 Automated Factories，且 `planet+0x0A <= 2`
   （已證實的礦產等級 Ultra Poor／Poor／Abundant）時成立。
2. `signed(player+0x89F)/2 <= 1`，亦即已證實政府碼中的 Feudal／Confederation／
   Dictatorship／Imperium，且已知但尚未建造 Marine Barracks 或 Armor Barracks 時成立。

四棟研究設施的完整公式因此可閉合：

| raw ID | 建築 | consumer／正值指令 | typed 公式 |
|---:|---|---|---|
| 6 | 自動實驗室 | `0xD06AD..0xD06CB` | priority gate 或 late-tech 時 0，否則 `11+4×[Erratic]` |
| 19 | 銀河網路中心 | `0xD07CA..0xD07E4` | priority gate 或 late-tech 時 0，否則 `11` |
| 30 | 行星超級電腦 | `0xD08CD..0xD08E9` | priority gate 或 late-tech 時 0，否則 `8+3×[Erratic]` |
| 35 | 研究實驗室 | `0xD0925..0xD0942` | priority gate 或 late-tech 時 0，否則 `5+2×[Erratic]` |

`var_3C` 由 `player+0x28 == 3` 建立；既有 AIRACES 證據把 raw 3 對到
`PersonalityErratic`。四式不讀 `var_18`，所以不受國庫／淨收入平方根因子影響。

## 第三批：Biospheres

raw building ID 15 的 jump-table entry 位於 `0xCFF9A`，原始 bytes
`84 07 0D 00` 直接指向 `loc_D0784 @ 0xD0784`：

1. `test ah,ah @ 0xD0784`；共用 priority gate 成立時直接走共同零分出口。
2. 否則 `mov ebx,var_1C @ 0xD078C`，再 `add ebx,12h @ 0xD078F`。
3. `var_1C` 由 `player+0x28 == 5 @ 0xD006A..0xD0077` 建立；既有 AIRACES 對照把 raw 5
   釘為 `PersonalityPacifist`。

因此 Biospheres 的完整 typed 公式為：priority gate 時 `0`，否則
`18 + [Pacifist]`。這條路徑不讀 late-tech、人口或 `var_18`，沒有額外平方根因子或未解欄位。

## 第四批：Food Replicators

raw building ID 16 的 jump-table case 位於 `0xD0797..0xD07BD`。其完整資料鏈如下：

1. raw `sub_D3D34 @ 0xD489D..0xD4967` 建立每殖民地 7-byte AI cache。`cache+0` 由
   raw `sub_D2A08 @ 0xD2A08..0xD2AA9` 從 `colony+0x0C` 的 packed colonist 低 nibble
   計數，依原版多數／owner fallback 規則選出主要人口 player slot。
2. `memset @ 0xD489D..0xD48A4` 先令 cache 全零。`cmp primary+0x8B1,0 @ 0xD4946`
   的 `jz 0xD4963` 會直接執行 `mov [cache+2],1`；primary 為 Lithovore 時才繼續檢查
   owner，且只有 owner 也為 Lithovore 的 `jnz 0xD4967` 會跳過寫入。故 cache+2 的精確
   語意是 `!(primaryLithovore && ownerLithovore)`，不是先前只看落入寫入點而誤讀成的
   「主要人口為 Lithovore、owner 非 Lithovore」。此結論由 memset、兩個條件跳躍與唯一
   寫入端共同證實。
3. raw 16 先於 `0xD079A` 檢查 `cache+2`；只有主要人口與 owner 同為 Lithovore 時零分。
   其餘情況以 signed
   `player+0xB0` 選 `8`（負值）或 `4`（非負），最後加 `var_1C=[Pacifist]`。
4. `player+0xB0` 有兩條獨立直接寫入證據：raw `sub_DF8F0 @ 0xDFA34..0xDFA39`
   將逐殖民地食物產出減消耗的總差寫入；raw `sub_E2710 @ 0xE2A18..0xE2A1E`
   亦將聚合食物產出 `var_1C` 減聚合消耗 `var_10` 寫入。其符號可由 remake
   `EmpireOutput.TotalFoodHalf` 精確表示；半單位尺度不改變 `<0` 邊界。

因此完整 typed 公式為：主要人口／owner profile 不完整時維持未知並走明示 fallback；
資料完整且主要人口與 owner 同為 Lithovore 時 `0`；其餘為
`4 + 4×[帝國食物盈餘<0] + [Pacifist]`。這條路徑不讀 priority gate 或 late-tech。

## 第五批：Cloning Center 與 `var_18` 勘誤

raw building ID 10 的 jump-table entry 位於 `0xCFF86`，原始 bytes
`1c 07 0d 00` 指向 `0xD071C`。完整控制流為：

1. `test ah,ah @ 0xD071C`：priority gate 成立時直接零分。
2. `cmp byte ptr [edi+0x8A0],0 @ 0xD0724`：`+0x8A0 = +0x89F + trait index 1`；
   受版控 RACESTUF 轉換表與 `.GAM Traits[31]` 已證實 index 1 是人口成長 runtime 百分點。
   只有負成長時於 `0xD0731` 加 `[Pacifist]`。
3. 兩條路徑均進 `0xD05B0`，加入 `floor(var_18/2)`。

`var_18` 由函式前段 `0xD009B..0xD00BD` 建立：結算前國庫 `player+0x32 < 1500` 時為 0；
否則把 signed word `player+0xB2` 以朝零截斷除以 64，傳入 raw
`sub_134C92 @ 0x134C92..0x134D2D`，再於 `0xD0135..0xD0142` 夾到 10。
完整指令顯示 `sub_134C92` 是 unsigned 32-bit 整數平方根，不是 PRNG；負商會命中
`0x134CAD..0x134CBD` 的高值捷徑回傳 65535，之後同樣夾成 10。

`player+0xB2` 已由 `sub_E2710 @ 0xE2A4B..0xE2A64` 與維護費稽核閉合為本回合淨 BC，
`player+0x32` 是尚未加上該值的國庫。`Next_Turn_Calc_ @ 0x136B3` 於 `0x13715` 先呼叫
AI 殖民地建造，直到 `0x13742` 才呼叫 raw `sub_E4F49` 套用本回合淨 BC；因此 remake
以 `EmpireOutput.Player.BC-EmpireOutput.NetBC` 重建相同的結算前門檻。

完整 typed 公式為：priority gate 時 `0`；否則
`floor(budgetFactor/2) + [race growth < 0]×[Pacifist]`，其中
`budgetFactor = 0`（結算前國庫 `<1500`），否則
`min(10, isqrt32(uint32(trunc(int16(netBC)/64))))`。這裡的 `int16` 是原版 word 儲存契約，
不是 remake 任意縮窄。

## 第六批：Holo Simulator／Pleasure Dome

兩個 jump-table entry 與原始 bytes 為：

| raw ID | 建築 | entry → target | 固定正值 |
|---:|---|---|---:|
| 20 | Holo Simulator | `0xCFFAE 16 06 0d 00 → 0xD0616` | 10 |
| 31 | Pleasure Dome | `0xCFFDA 5c 06 0d 00 → 0xD065C` | 16 |

`0xD0616..0xD0657` 與 `0xD065C..0xD069E` 是同形控制流：

1. 讀 `player+0x89F` 的政府碼，以帶號、朝零除以 2；結果等於 3 時直接零分。
   既有 31-byte trait 表與政府 enum 已逐值對回 0..7，因此只排除 raw 6／7，也就是
   Unification／Galactic Unification。
2. `colony+0x0A >= 3` 時直接給固定正值。
3. 人口小於 3 時先要求 `var_18 > 0`，再要求人口至少 2；因此只有人口恰為 2 且
   `budgetFactor > 0` 仍給固定正值，人口 0／1 或因子 0 均為 0。
4. 兩條 case 都不讀共用 `ah` priority gate、late-tech 或 personality，也不把
   `budgetFactor` 加進分數；它在兩人口邊界只作布林 gate。

完整 typed 公式為：Unification 系政府時 `0`；否則若
`population >= 3 || (population == 2 && budgetFactor > 0)`，Holo Simulator 為 10、
Pleasure Dome 為 16；其餘為 0。

## 第七批：Gaia Transformation

2026-08-26 以同一份正式 `.i64` 的一次性副本重跑 IDA Pro 9.4／IDAPython；輸入與資料庫
SHA-256 仍分別為本文件證據契約所列的
`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` 與
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。非破壞性匯出
`/tmp/moo2-ai-gaia-v1.json` 再次確認：

- raw building ID 17 的 jump-table entry 是
  `0xCFFA2 c2 07 0d 00 → loc_D07C2 @ 0xD07C2`。
- `mov ebx,var_1C @ 0xD07C2` 先令分數等於 `[Pacifist]`；`jmp loc_D0414 @ 0xD07C5`
  隨後無條件執行 `add ebx,var_18 @ 0xD0414`，再走正值與 1000 上限出口。
- 此 case 不讀 priority gate、late-tech、人口、政府或星球氣候。氣候適用性屬
  `Colony_Can_Build_Product_` 的候選 gate，不是這個分數函式自行處理。

因此完整計分公式是 `budgetFactor + [Pacifist]`；`budgetFactor` 沿用第五批已閉合的
結算前國庫、signed word 淨 BC、朝零除以 64、unsigned 整數平方根及 10 上限契約。
remake 的候選層只在 typed 殖民地為 Terran 時提供 Gaia Transformation，完工後必須把
AI 殖民地與全局行星同步改成 Gaia；這個 typed 適用性 gate 是既有手冊／地形規則契約，
不冒稱本次 `sub_D0036` 計分反組譯本身證實了 `Colony_Can_Build_Product_` 的完整控制流。

## 第八批：Terraforming

同日以 IDA Pro 9.4 重跑擴充後的 `tools/ida/audit_ai_colony_build.py`；非破壞性匯出
`/tmp/moo2-ai-terraform-v1.json` 的輸入、資料庫 SHA-256 與位址基準均和本文件證據契約一致。
raw 44 外層 entry 是 `0xD000E 53 0a 0d 00 → 0xD0A53`。`0xD0A53` 先檢查共用
priority gate，成立即零分；否則讀已由事件與地形改造 consumer 證實為氣候的
`planet+8`，減 2 後以 `jmp cs:jpt_D0A67[eax*4] @ 0xD0A67` 分派。

內層表位於 `0xD001E`，六筆原始 dword 為：

| raw 氣候 | 氣候 | entry bytes → target | 非 Aquatic 基礎分 | Aquatic 基礎分 |
|---:|---|---|---:|---:|
| 2 | Barren | `0xD001E 6f 0a 0d 00 → 0xD0A6F` | 2 | 2 |
| 3 | Desert | `0xD0022 7f 0a 0d 00 → 0xD0A7F` | 1 | 1 |
| 4 | Tundra | `0xD0026 76 0a 0d 00 → 0xD0A76` | 0 | 1 |
| 5 | Ocean | `0xD002A 86 0a 0d 00 → 0xD0A86` | 4 | 0 |
| 6 | Swamp | `0xD002E 96 0a 0d 00 → 0xD0A96` | 6 | 0 |
| 7 | Arid | `0xD0032 7f 0a 0d 00 → 0xD0A7F` | 1 | 1 |

Aquatic 分支直接讀 `player+0x8AB @ 0xD0A76/0xD0A86/0xD0A96`；此位址等於 runtime
trait 基址 `+0x89F` 加 index 12，且 `sub_DE0C6` 等獨立食物 consumer 已證實 index 12
是 Aquatic。有效氣候 case 最後在 `0xD0AB0` 加 `3×[Pacifist]`，再經 `0xD0414` 加
`budgetFactor`。因此完整計分公式是：priority gate 時 0；否則
`climateBase[climate][Aquatic] + 3×[Pacifist] + budgetFactor`。

Toxic／Radiated／Terran／Gaia 不在這個內層表；remake 依既有 typed
`TerraformNextClimateOptions` 只讓具下一級的 Barren..Arid 成為候選。完工時採同一 helper
的第一個結果（Barren 的 Desert／Tundra 二選一仍是既有明示近似），並同步 AI 殖民地與
全局行星。候選適用性與 Barren 分歧並非本次計分函式的「已證實」內容。

## 第九批：Soil Enrichment

raw 37 的 entry 為 `0xCFFF2 56 09 0d 00 → 0xD0956`。完整控制流是：

1. `test ah,ah @ 0xD0956`：priority gate 成立時零分。
2. `cmp [cache+2],0 @ 0xD0961`：主要人口與 owner 同為 Lithovore 時零分；旗標完整
   寫入證據與勘誤見第四批。
3. `cmp byte ptr [colony+0xDD],0 @ 0xD096B`：目前每農夫食物產出不大於 0 時零分。
   `sub_13A3D @ 0x13F3D..0x13F4E` 由 colony 的 planet index 讀 `planet+0x0B` 後寫入此欄；
   正常重算端 `sub_E1D59 @ 0xE1D6F..0xE1D7B` 呼叫 raw `sub_DE03E @ 0xDE03E` 再回寫。
   `sub_DE03E` 將 planet 食物基值乘 2，加入 Weather Controller 的 4 個 half-unit 與
   Astro University 的 2 個 half-unit，因此 `+0xDD` 是目前每農夫食物的半單位快取。
   raw 37 只問正負，remake `FoodPerFarmer>0` 可保持同一邊界。
4. `player+0xB0 < 0` 時基礎分 5，否則 3；`0xD098E..0xD0993` 再加
   `2×[Pacifist]`。此 case 不加入 `budgetFactor` 或 late-tech。

因此完整公式為：priority gate、cache+2 為 0 或每農夫食物不大於 0 時為 0；否則
`3 + 2×[帝國食物盈餘<0] + 2×[Pacifist]`。remake 只在既有
`TerraformSoilEnrichmentWorks` 允許的氣候提供一次性候選；完工後每農夫食物 +1，且不寫入
`ColonyBuildings`。原版 planet `+0x0B` 沒有 remake 的獨立全局行星欄位，持久化權威是
`ColonyState.FoodPerFarmer`；這是資料模型對映，不宣稱逐欄保存原版 planet record。

其餘未封閉區域會讀 alien／outpost 狀態、政府／性格其他碼、其他殖民地 packed 人口用途、帝國建築數、
星球 owner／環境、事件與未解 player flags；在欄位寫入端與 typed 對映完成前維持
`unknown_pending_review`，由明示近似 fallback 處理。
