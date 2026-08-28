# 戰略轟炸殖民地傷亡回寫靜態稽核（2026-08-24）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫副本：`moo2-Orion2-consumer-work.i64`，SHA-256
  `6562313be340a6bb80d43f25478446ba0bae24285ac86f0419b4f7de02a14fd0`。
- 工具：IDA Pro 9.4／IDAPython，SDK 940；本文位址均為 IDA linear address。
- 探針：`tools/ida/audit_bombardment_writeback.py`。探針只匯出原始函式名、位址、
  bytes、operand、交叉參照與反編譯導覽文字，不改名、不套型別、不寫回資料庫。
- `sub_E87D2` 因區域配置失敗無法取得可靠 Hex-Rays 偽碼；以下結論以原始指令、
  caller／callee 與欄位讀寫為證據，不把導覽文字當證據。

## 呼叫鏈與結果池

**已證實**：`sub_E87D2 @ 0xE87D2` 的唯一直接 caller 是 `sub_E938C @ 0xE938C`
內的 `0xE9859`。它多次呼叫 `sub_4267B @ 0x4267B` 建立轟炸結果，再呼叫
`sub_DD2F2 @ 0xDD2F2` 消費結果。

戰略分支由 `byte_199CB4 == 1` 選出。`sub_4267B` 呼叫 `sub_4257E` 後：

- `0x42720`：結果 record `word +0x06 = 0`。
- `0x42726`：把 `sub_4257E` 回傳的 `/40` 傷害點數寫入 `word +0x03`。

`sub_DD2F2` 在 `word +0x06 + word +0x03 > 0` 時迴圈消費；`+0x06` 非零走
`sub_DD13E @ 0xDD13E`，否則把剩餘 `+0x03` 傳給 `sub_DCEBD @ 0xDCEBD`，
再扣掉 helper 回傳的成本。helper 回傳零時直接停止消費。因此戰略轟炸的玩家可見
傷亡分配走 `sub_DCEBD`；`sub_DD13E` 屬另一結果池，不是本輪 remake 的戰略路徑。

## `sub_DCEBD` 候選池

以下原始殖民地偏移均由已驗證 `.GAM` `Colony` layout 對回，證據等級為**已證實**：

- `+0x0A`：人口。
- `+0xB4 + 2*race`：各種族人口的百分之一人口點數。
- `+0x125`：目前建造進度。
- `+0x130`：士兵／陸戰隊。
- `+0x132`：戰車。
- `+0x136 + rawBuildingID`：49 個建築旗標。

helper 依下列順序建立等機率候選：

1. raw building ID 由 48 遞減到 1；旗標有效且 ID 不在
   `{8, 9, 26, 27, 40, 41, 42, 47}` 才加入。
2. 每一名士兵各加入一項。
3. 每一輛戰車各加入一項。
4. `BuildProgress != 0` 時加入一項。
5. 人口不等於 1 時，每單位人口各加入一項。

排除的建築依序是 Battlestation、Capitol、Missile Base、Ground Batteries、
Star Base、Star Fortress、Stellar Converter、Fighter Garrison。這份排除表只適用
`sub_DCEBD` 的殖民地內部傷亡候選；不可拿來改寫 `Get_Colony_Hits_` 的快速戰鬥耐久表。

候選數大於零時呼叫 `sub_1247A0(candidateCount) - 1` 選取索引。其回傳範圍可由
減一後作陣列索引的消費端判定為 1..N；「各候選等機率」為**強推論**，尚未在本輪
深入該亂數 helper 的內部實作。

## 各類候選的寫回與成本

### 建築

**已證實**：呼叫 `sub_145EA(colonyIndex, rawBuildingID)` 移除建築，並把結果 record
`byte +0x0E+rawBuildingID` 設為 1；回傳成本 1。

### 士兵與戰車

**已證實**：士兵基礎成本 1，特定玩家欄位可加 1、另一玩家欄位再加 1；成本不超過
剩餘傷害才扣 `colony +0x130` 並增加結果 `word +0x08`。戰車同理，基礎成本 2，
扣 `colony +0x132` 並增加結果 `word +0x0A`。

玩家欄位與既有地面戰公式對齊為 Heavy-G／Powered Armor（士兵）及
Heavy-G／Battleoids（戰車）是**強推論**：成本形狀與已驗證的地面單位 hit 規則一致，
但本輪沒有替 `Player +0x1A7`、`+0x12F`、`+0x8AA` 各自完成獨立寫入端追蹤。
remake 因此沿用已有的 typed 科技／種族 helper，不把 raw offset 名稱升格成已證實語意。

### 建造進度與人口

**已證實**：候選池只在建造進度非零時加入一項，但選中尾端後，原始程式仍拿尾端索引
與實際 `word [colony+0x125]` 比較。索引小於進度時會把全部建造進度清零，並把原值寫到
結果 `+0x43`；其他情況則移除一名人口。這是原始控制流可見的資料形狀不對稱，remake
保留它，不擅自修成「每點只扣一點進度」。

人口大於一時，被抽中的殖民者由最後一名有效殖民者覆蓋，再把人口減一。人口等於一時
不加入一般候選；只有其他候選全空才以 100 點傷害扣該殖民者種族的 `RacePopulation`。
剩餘點數大於 100 時只扣 100；不大於 100 時人口歸零並設定殖民地毀滅相關旗標／計數。

## remake 邊界

- 本輪精確重製 `sub_DCEBD` 的候選集合、順序、成本不足即停止、建造進度尾端分支、
  最後人口 100 點模型與欄位回寫。
- 1.3 的 `BombardmentBuildingBonusHits` 來自 CHANGELOG，並非本份 1.50 executable；
  它保留為版本相依近似，不能宣稱由本次 IDA 證實。
- **2026-08-28 勘誤**：排除不代表八棟各有獨立摧毀分支。Capitol raw 9 沒有 combatant；
  三層軌道基地 raw 40／8／41 互斥地建立一個額外快速 record；raw 26／27／42／47 只接進
  行星 record 的武器／旗標。`sub_4267B` 把特殊池 `+0x06` 清零，故正常戰略轟炸不逐棟移除
  這七種防禦。完整證據見
  [`strategic-bombardment-full-audit-20260828.md`](strategic-bombardment-full-audit-20260828.md)。
- 合法資料不接近 16-bit 溢位；remake 使用有界 `int`，不模擬溢位。
