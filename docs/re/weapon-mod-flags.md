# 武器改造與飛彈旗標證據索引

> 本檔是反組譯證據索引，不是把所有旗標都命名成已知規則。每筆結論都保留原始 raw
> 位址、來源檔、位址基準與證據等級；未達「已證實」的項目不得直接變成 UI 或戰鬥規則。

## 輸入與工具追溯

- 原始執行檔：`/home/anr2/moo2-private-build/re/Orion2.exe`；SHA256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 匯出的組語：`Orion2.exe.asm`；SHA256
  `76cac6231a60da0fdba705907a88a853a1d345ed7bb7c15788b280fdbb259a18`。
- IDA 資料庫：`Orion2.exe.i64`；SHA256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 交叉檢查的反編譯匯出：`orion2_all.c`；SHA256
  `c2b5c30701019c0cc58763eb29c2abddb55eb551e1e7e52f68070d629e694505`。
- 工具：IDA Pro 9.4；組語檔標頭記錄為 2026 產物。`0x……` 全部是 IDA 的 LE 32 位元
  OS/2／DOS 平面位址，不是檔案偏移；`.asm` 只作可重生的文字證據，交叉參照與函式邊界
  仍以 `.i64` 為主。

## 2026-08-09 函式名稱與消費端勘誤

`func_names.txt` 的歷史命名有一處容易造成誤接：完整光束改造函式不是 `sub_39434`，而是
`sub_394F7`；`sub_39434` 是它呼叫的射程／命中輔助函式。原始名稱、raw 位址與目前解讀如下：

| raw 位址 | 原始名稱／目前標籤 | 可直接回查的事實 | 等級 | 重製決策 |
|---|---|---|---|---|
| `0x39434` | `Get_Beam_Range_To_Hit_Bonus_` | 初始化範圍／命中輸出；改造旗標 `0x02` 使有效射程除 2，`0x04` 使射程乘 2（上限 8）；`0x20` 跳過射程衰減處理，靜態武器旗標也有免衰減分支。 | 已證實 | `ResolveBeamShot` 的 HV／PD／NR 接線引用此函式的語意，並保留完整改造 caller `0x394F7`。 |
| `0x394F7` | `Get_Beam_Weapon_Modifiers_` | 讀取 `word_17F815` 靜態武器欄位並合併到改造旗標，再呼叫 `0x39434`；旗標 `0x80` 走三發／每發 −20 命中路徑，旗標 `0x10` 加 25 戰鬥值，另一個高位旗標控制特殊四次迭代。 | 已證實（光束側） | 文件與程式註解把它列為完整光束改造消費端；不把這些光束旗標外推成 ARM/FST。 |
| `0x3CD21` | `func_names.txt` 原名 `Missile_Facing_`；舊匯出另稱 `Missile_Speed_` | 依 weapon kind、FTL／玩家旗標與呼叫旗標組合回傳速度樣式值；旗標 `0x10` 執行 `add edx,4`。`0x0E..0x11` 基礎值 12，`0x12/0x13` 為 20、`0x14` 為 24 且三者都不加 FTL，`0x1C..0x1F` 與 `0x28` 各有獨立分支。 | 已證實（分支）；名稱語意仍需保留 raw 定位 | `MissileSpeedOf`／`MissileBeamDefenseOf` 保留公式；FST 的速度與標準飛彈 Beam Defense 已進 remake 的 PD 垂直切片，完整原版攔截器仍未知。 |
| `0x3CE4F` | `func_names.txt` 原名 `Missile_Speed_` | 函式本體檢查飛彈座標是否位於戰鬥畫面範圍，回傳視窗內／外布林值；不是速度函式。 | 已證實（反駁舊名稱解讀） | 不讓此 raw 位址成為速度或 FST 證據；速度以 `0x3CD21` 為準。 |
| `0x35EAE` | `Fighter_Pilot_Bonus_` | 依 empire id 掃描仍有效的戰鬥艦記錄，對有軍官的艦呼叫 helper `0x35E6A`，保留最大值；helper 的普通／進階分支為 `5*(level+1)`／`15*(level+1)/2`。 | 已證實（迴圈／最大值／算術）；helper 舊標籤語意仍保留 | `fighterPilotBonusForCombat` 對選定參戰艦隊取 `SKILL_FIGHTER_PILOT` 最大值，並由母艦帶入戰機 Beam Defense；不把未參戰艦隊或敵方未知資料算進來。 |
| `0x3DFE0` | `Fighter_Ocv_` | 呼叫 `0x3CD21`，把結果乘 5；戰機類別路徑再合併 `0x35EAE` 的 Fighter Pilot 結果與 empire record `+2213` 的一個加成位元組。它是 OCV-like 下游證據，但沒有從此處追回敵方完整戰機消費鏈。 | 已證實（呼叫／算術／讀取）；`+2213` 欄位語意未知 | `MissileBeamDefenseOf`／PD 垂直切片使用已證實的速度分支；玩家戰機再接已證實的種族 Ship Defense 與 Fighter Pilot；不把 OCV-like 值宣稱成完整原版攔截命中率。 |
| `0x3E095` | `Missile_Dcv_` | 標準 kind `0x0E/0x0F/0x10/0x11` 回 `4/8/12/16`；`0x12..0x14` 回 5；`0x1C..0x1F` 讀 `Best_Computer` 後用 `6/10/8/4 × (bonus+100)/200`；high-byte bit `0x08` 使結果加倍。 | 已證實（分支／算術）；ARM 名稱是強推論 | `MissileInterceptionDurabilityForRawFlags` 保留 raw flag `0x0800` 與 Dcv；`Weapon_In_Range` 的攔截鏈已接到 remake PD，但完整原版參數／所有武器映射仍未知。 |
| `0x3A0B9` | `Weapon_In_Range_` | 讀 runtime 飛彈 kind／raw flags，呼叫 `0x3E095`；把本次攔截傷害先夾到一個 Dcv，再加 runtime `+0x0D` 餘數，以 quotient/remainder 算擊落枚數並寫回餘數。 | 已證實（資料流／算術） | `MissileWarheadsDestroyedByInterception` 複現上限、擊落數與餘數；快速結算／格子戰術的 PD 都走共用 `ResolvePointDefenseIntercept`。 |
| `0x3AD57` | `sub_3AD57`（外部索引名稱衝突） | `a3 != 0` 的攔截分支先計 OCV-like 值，再對飛彈呼叫 `Weapon_In_Range`；否則進入 `Fire_Beam_Weapon`。完整 fighter／beam 參數仍有未定欄位。 | 已證實（呼叫鏈）；完整武器映射未知 | remake 只接「PD 對標準飛彈」垂直切片，沒有把它宣稱成原版戰機／攔截機全模型。 |
| `0x3D2DF` | `Is_Missile_Target_Alive_` | 讀取飛彈 runtime record 的旗標與武器欄位；可觀察到旗標高位元組 bit 0 控制一次／四次套用、bit 1 影響數量、bit 2 使一個中間值減半、bit 6 參與特殊分支，低位元組 `0x20` 也改變特殊分支。 | 已證實（讀取）；旗標語意未知 | 保留 raw bits 與偏移，暫不批次命名為 ARM/FST/MIRV。 |
| `0x39F1D` | `Destroy_Ship_` | 讀取設計武器槽 `+0x52` 的 weapon ID、`+0x56` 的 16 位元旗標；`+56 & 4` 使一個值加倍，`+56 & 2` 使其減半，之後做武器／目標類別合法性判定。這不是已證實的飛彈生命值消費端。 | 已證實（欄位讀取）；ARM 語意未知 | 不將 `+56` 的位元直接命名為 ARM。 |
| `0x3A142` | `Apply_Damage_To_Missile_` | 將現有飛彈 runtime metadata 組成 bitmask；目前沒有看到 ARM 清除／攔截傷害加倍的直接寫入端。 | 已證實（流程）；ARM 消費端未知 | ARM 保持未接線。 |
| `0xABC5D` | `Anti_Missile_Rockets_` | 載入 `CMBTSFX.LBX`，計算位置並呼叫動畫／音效／繪圖流程；周邊路徑也處理飛彈數量，但本函式本身沒有可回查的 ARM 彈體耐久公式。 | 已證實（視覺流程） | remake 的 AMR 仍只使用已證實的距離命中率；不把 ARM/FST 硬接進去。 |

## 2026-08-09 戰機目標與速度下游盤點

這一輪重新使用同一份 `Orion2.exe.asm`／`.i64` 對照，補記戰機與一般飛彈共用的
runtime 記錄，避免把「有速度函式」誤寫成「敵方逐艦設計已解出」：

| raw 位址 | 原始定位／讀寫事實 | 等級 | 重製決策 |
|---|---|---|---|
| `0x3C892` | 建立 stride `0x1A` 的飛彈 runtime record；從設計槽 `+0x52` 複製 weapon ID、`+0x54` 複製數量，`+0x56` 的 raw flags 複製到 runtime `+0x11`，再呼叫 `0x3CD21`，把結果存到 runtime `+0x16` 並形成 `+0x13` 的座標／距離值。 | 已證實（欄位流向） | FST 只保留 raw／公式證據；目前 remake 的飛彈是即時解算，沒有可安全承接這組逐彈 runtime 記錄的移動器。 |
| `0x3CD21` | 讀取呼叫旗標的 `0x10` bit 時執行 `add edx,4`；該旗標在 `0x3C892` 的 runtime 初始化鏈中來自設計槽 raw flags。 | 已證實（速度分支）；旗標名稱未知 | 可支持 FST「速度 +4」；不單憑此宣稱完整攔截器 Beam Defense 已接。 |
| `0x3DFE0` | `Fighter_Ocv_` 先呼叫 `0x3CD21`，再把其結果乘 5；同一 raw 函式被 `0x2B7CC`（飛彈有效強度路徑）與 `sub_3AD57` 呼叫。 | 已證實（呼叫與算術）；完整參數語意強推論／未知 | 保留 `MissileBeamDefenseFast`／戰機公式資料；不把 OCV-like 值直接改名成單一 Beam Defense 欄位。 |
| `0x3E563`、`0x2A46A` | `0x3E563` 以戰鬥記錄 `+0x4A` 判斷 `Fighter_Combat_SFX`；`0x2A46A` 的 `Target_Ship_Value` 掃描候選時排除該類記錄，再呼叫 `0x2A239` 進行艦艇目標評估。 | 已證實（讀取／排除端） | remake 一般戰術畫面只在有逐艦資料的敵艦中選目標；不因這段證據虛構 `genEnemyFleet` 的敵方戰機。 |
| `0x3AD57` | `sub_3AD57` 的一個分支呼叫 `Retarget_Missile` 與 `Fighter_Ocv`，另一分支進入 `Fire_Beam_Weapon`；IDA 反編譯仍有未定參數與跨函式資料流。 | 已證實（呼叫）；完整武器／目標映射未知 | 不把相鄰 `sub_3AC20` 的直接傷害式誤套成這條命中公式。 |

手冊 p.157 已明定：主要目標仍有效時維持它，只有主要目標失效且仍有射擊才自動重選。
因此 `cmd/moo2/tacticalfighter.go` 現在讓 `FighterSquadron` 保存 `TargetName`，只在目標
失效時重選；這是目標狀態的已證實行為，不代表已解出原版候選優先分數。

手冊 help 第 402 筆另明定：若近迫防禦武器尚未開火，會在飛彈命中艦艇或戰機接戰前先開火。
重製因此新增 `PointDefenseCanFire`／`ResolvePointDefenseFighterShot`，讓格子戰術的戰機中隊
在接戰前消費同一艘艦的 PD 使用旗標；戰機命中判定使用 p.157 的
`5*Speed + racial Ship Defense + Fighter Pilot + Helmsman`；目前玩家呼叫端已由參戰艦隊資料
提供前兩項，Helmsman 明示傳 0。這是「順序＋公式＋玩家種族／戰機飛行員加成」的切片，
不是原版 `sub_3AC20`／`sub_3AD57` 的完整 raw record 參數、敵方戰機政策或逐彈 runtime 等價。

## 目前可安全接入的邊界

1. 光束 NR 的完整原版定位是 `0x394F7 → 0x39434`；重製光束路徑已用
   `DamageDissipationPenalty` 接入，魚雷仍沒有可證實的射程衰減模型。
2. FST 的 `+4` 分支與速度 runtime 下游已由 `0x3C892`／`0x3CD21`／`0x3DFE0` 串起；
   remake 以標準四種飛彈的 `MissileBeamDefenseOf` 接到 PD 垂直切片，但完整原版反飛彈武器
   的命中／行動順序仍未知；戰機接戰前的 PD 觸發順序已依 help 第 402 筆接入，但其
   `sub_3AC20`／`sub_3AD57` 的完整 raw 參數與敵方路徑仍未知。
3. ARM 的手冊效果「摧毀所需傷害 ×2」與 `0x3E095` high-byte bit `0x08`、`0x3A0B9`
   的 Dcv 消費鏈相符，列為強推論；remake 已以 raw `0x0800` 接到共用 PD 攔截，但不把
   這個垂直切片升格成原版所有攔截器／AMR／戰機路徑都已還原。
4. 未知 raw kind、魚雷與尚未追回的完整 runtime 讀寫端仍拒絕套用推測；魚雷 NR 仍是獨立缺口。

後續若要完成 ARM/FST 的原版等價，最低證據門檻仍是：保留上述 raw 位址與 runtime record
偏移，追回反飛彈光束／攔截機／AMR 對同一飛彈實體的完整傷害寫入端，再以原版與重製各自
的固定輸入做邊界值比對；目前測試證明的是 remake 垂直切片內部一致，不是原版全路徑等價。

## 2026-08-11 IDA Pro 位址基準勘誤

本文件早期表格使用的是 object-offset ledger；本輪以同一專案的 `Orion2.exe.i64` 直接讀取
IDA 線性位址，object #1 的 `0x3CD21` 對照為 `0x13CD21`。該資料庫在 `0x13CD21` 顯示的
原始函式是 `sub_13CD21`（間接呼叫／`putch_` 路徑），並未重現表格中所有 `Missile_Speed_`
語意；`symbols_fixed.tsv` 與 `symbols_full.tsv` 對同一 object-offset 也有不同名稱。

因此表格的歷史證據與既有 remake FST 垂直切片保留不動，但截至本批次，不能把這個 IDA 直接
對照當成新的速度 oracle。完整輸入雜湊、位址基準、raw 指令與未解項目見
[`oracle-static-ida-20260811.md`](oracle-static-ida-20260811.md)；後續若要升格，必須以同一
來源資料庫重新建立 object-offset ↔ IDA 線性位址對照，並保留本勘誤。

## 2026-08-11 remake 消費端更新

上面的「完整 `sub_3AC20`／`sub_3AD57` raw 參數未知」仍然成立；但同一 raw chain 已足以支持
不再使用固定近似傷害。`sub_3AD57 @ 0x3AD57` 對艦艇路徑的攻防修正、`<=95` 尾端
處理、40 命中門檻與 ID 31 第二組傷害範圍已落到 `internal/shell/fighter_attack.go`
與 `internal/gamedata/fighter_damage.go`；相鄰 `sub_3AC20 @ 0x3AC20` 的
`D=min+floor((floor(100/(2*S))+R)*S/100)` 也由 `ResolveFighterBomb` 獨立實作，
並接到 Bomber profile。兩條路徑都會逐架把結果送入最弱護盾／裝甲／結構下游；未解的是
完整 `sub_3DF8D` 攻方加成來源、raw flags 的正式名稱與 raw record 輸出指標，不是把
兩條相鄰公式混成一條。

### 2026-08-11 IDA Pro 深度勘誤：Fire 兩個相鄰 raw function

同一份 `Orion2.exe.i64`（SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`）、
IDA Pro 9.4／IDA 線性位址下，兩份外部索引對名稱互相錯位：`0x3AC20` 在
`symbols_fixed.tsv` 叫 `Fire_Fighter_Bomb_`，`0x3AD57` 則叫 `Fire_Fighter_Beam_`；
`func_names.txt` 恰好相反。因此本檔以下以 `sub_3AC20`／`sub_3AD57` 為主，不把
索引名稱升格成證據。

`sub_3AC20 @ 0x3AC20` 的 raw 公式是：從 `record + int16(EAX)*0x1A` 找設計與
`word_17F815/17/19[W*0x1C]` 的 raw/min/max，`R=sub_1247A0(100)` 為 1..100；
`S=max-min>0` 時 `D=min+floor((floor(100/(2*S))+R)*S/100)`，否則 `D=min`。
它把每次結果送 `sub_39985 @ 0x39985` 並依 runtime `+0x0F` 數量累加。

`sub_3AD57 @ 0x3AD57` 的艦艇路徑則是另一式：`R<=95` 才加
`OCV=sub_3DF8D(W)-target defense`（另一 branch 用 `sub_3DFE0`），只把結果上夾
100；`R+OCV<40` miss，命中時
`min+floor((M-40)*(max-min+1)/60)`，`DI==0` 再做 signed half。其
`var_18` 只由 raw `0x0100` 產生 `0x40`，故 `test var_18,4 @ 0x3AE4E` 不可達；
`RawFlags & 4` 不應被 remake 虛構成半傷害／+25 效果。`sub_39985` 與
`sub_3A0B9 @ 0x3A0B9` 的盾／甲／結構或 runtime record 消費分支已由 caller data flow
確認；完整 `sub_3DF8D` 欄位語意與兩份名稱索引仍標為**強推論／未知**。
