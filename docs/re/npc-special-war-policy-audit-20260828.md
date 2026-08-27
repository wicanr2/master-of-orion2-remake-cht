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

## 尚未閉合

- **未知**：`player+0x60E == 1` 會走另一條無亂數的理由 113 分支；目前只有
  `sub_25DF1` 與 `sub_544A1` 讀取，尚未找到可回查的玩家玩法 producer。
- **已證實資料形狀、未接 AI↔AI producer**：理由 22 讀方向性
  `player+0x6FF > 0`。`sub_DCEBD`、`sub_DD13E`、`sub_ECBF7`、`sub_ECF41`
  會在殖民地人口／建築／單位遭破壞或奪取時累加，宣戰或停戰時清除；現有抽象
  AI 艦隊只能在正式開戰後攻擊殖民地，不能憑空製造戰前怨值。
- **未知**：外層 `sub_233FA` 在全域值 2／4／6 時阻止一般來源執行，只有
  `player+0x8BC` 的特定旗標可穿透；這些全域值與現有事件型別尚未完成精確對映。

上述三項不阻止理由 20、68 與食物赤字理由 113 的可達垂直切片，但不得宣稱
`sub_25DF1` 已逐分支完整還原。
