# NPC 特殊宣戰候選逆向稽核（2026-08-28）

## 證據身分

- 輸入：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；
  資料庫內記錄的原始 `Orion2.exe` SHA-256 為
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4／IDAPython；IDA 線性位址空間。
- 非破壞性匯出：
  [`evidence/npc-treaty-negotiations-ida-20260827.json`](evidence/npc-treaty-negotiations-ida-20260827.json)。
- 原始名稱與定位一律保留；以下語意名稱只供導覽，不取代位址與原始運算元。

## 已證實

`sub_25DF1 @ 0x25DF1` 先建立每個目標的候選類型與理由，再依固定順序掃描各
分支；type 2 立即宣戰，type 1 只在來源目前沒有戰爭／任務時均勻抽一個目標。
難度至少 3 時，已和其他 AI 交戰的目標會在候選完成後被清除。

### 理由 20

- 來源 raw government 必須為 3。
- 目標須通過存活、接觸／探索、冷卻與方向性國力比例至少 100 的守門。
- `Random(30 * (difficulty*difficulty + 1)) == 1` 時建立 type 1、理由 20。

### 理由 68

- 只檢查 `turn % playerCount` 的輪值目標。
- `0x25F7B..0x26057` 以 signed relation `player+0x617` 計算：
  `threshold = (-relation - 5) / (2*difficulty + 1)`。
- 呼叫 `Random(100)`，結果不大於門檻時建立 type 1、理由 68。

### 理由 113 的食物赤字分支

- `sub_4DAB2 @ 0x4DAB2` 讀 signed word `player+0xB0`；負值時遞增
  `player+0x7EC`，否則把 `+0x7EC` 清為零。`+0xB0` 已由 AI 殖民地建造稽核
  的直接 writer 證實為帝國食物結餘。
- `sub_25DF1 @ 0x2611A..0x261D6` 每個來源先呼叫一次 `Random(100)`；結果小於
  `player+0x7EC` 時，把所有尚無候選且通過接觸、冷卻與國力守門的目標設為
  type 1、理由 113。
- `sub_544A1 @ 0x54623` 也讀 `+0x7EC`，但該外交評分用途不是本切片的完成證據。
- `inc word ptr [player+0x7EC]` 沒有飽和分支；因此 remake 依 signed
  16-bit 二補數語意由 32767 回繞為 -32768。

### 理由 113 的 `+0x60E` 分支

- **已證實 consumer**：`sub_25DF1` 只在 `player+0x60E == 1`、方向性國力
  比例至少 100 且 cooldown 非正時，無亂數建立理由 113 候選。
- **已證實資料鏈**：Player record stride 為 3753 bytes；parser 現保留 raw
  `+0x60E`，`.GAM` importer 投影到 `AIOpponent.OriginalWarFlag60ERaw`，JSON
  快照可往返，宣戰 consumer 已接線。
- **抽樣實體值**：`SAVE10.GAM` 以種族字串錨定各 Player base 後，八個
  `+0x60E` 皆為 0；這只證明該存檔的值，不證明新局以外永遠為零。

## 外層超空間亂流 gate

- **已證實**：`sub_233FA @ 0x233FA` 在事件 9 固定 record 狀態為 2／4／6 時
  回傳 true；完整事件 record 與 caller 鏈見
  [`random-event-hyperspace-flux-audit-20260825.md`](random-event-hyperspace-flux-audit-20260825.md)。
- `sub_25DF1` 在該 gate 成立時跳過一般來源，只有 `player+0x8BC == 1` 可繼續。
  `+0x8BC` 由手冊語意與多個航行 caller 強推論為跨維度能力；remake 使用既有 typed
  `TRAIT_TRANS_DIMENSIONAL` 投影。

## 尚未閉合

- **未知 producer**：`player+0x60E` 目前只有 `sub_25DF1` 與 `sub_544A1`
  的可靠讀取證據，尚未找到 runtime writer；因此 remake 新局不臆造該值。
- **已證實資料形狀、未接 AI↔AI producer**：理由 22 讀方向性
  `player+0x6FF > 0`。`sub_DCEBD`、`sub_DD13E`、`sub_ECBF7`、`sub_ECF41`
  會在殖民地人口／建築／單位遭破壞或奪取時累加，宣戰或停戰時清除；現有抽象
  AI 艦隊只能在正式開戰後攻擊殖民地，不能憑空製造戰前怨值。
上述 producer 未知與 `+0x6FF` 不阻止已有 `.GAM` raw 的消費端或理由
20、68 與食物赤字理由 113 的可達垂直切片，但不得宣稱
`sub_25DF1` 已逐分支完整還原。
