# NPC 對 NPC 外交談判規格

**狀態：CONFORMED（remake 可表示的 AI ordered-pair 切片）**

證據：[`../re/npc-treaty-negotiations-audit-20260827.md`](../re/npc-treaty-negotiations-audit-20260827.md)。

## 狀態

每個 AI ordered pair 保存：

- `policy`、`tradeActive`、`researchActive`；
- signed-byte `reputationRaw`（原版 `+0x6D7`）；
- signed-word `treatyBiasRaw`／`agreementBiasRaw`（`+0x68F／+0x69F`）；
- direction-only `tributeMode`（`+0x63F`）。

所有矩陣必須隨 JSON 存檔與熱座 AI 索引壓縮；舊存檔缺欄以原版初始化值 0
補入。

## 回合順序

1. 所有方向的 treaty／agreement bias 各 `+10`，正值夾回 0。
2. 依 outer 低到高、inner 低到高掃 ordered pair；self 跳過。
3. 每對先擲 `Random(250-40*difficulty)`，只有 1 才進行談判。
4. 依證據文件的 base、treaty、agreement、tribute 分數及原始門檻處理。
5. 通過頻率 gate 的 pair 不論是否建約，兩個 bias 最後各減 30。
6. 正式戰爭只由 policy／war state 消費；不得再用顯示關係 `-25` 宣戰或
   `+12` 自動停戰。

## 資料模型投影

- `+0x5EC` 暫以 `FleetStrength` 代入相同 ratio／cap／第三方戰爭除半公式。
- `+0x71F` 尚無可靠 writer，輸入固定初始化 0，不加第三方 bonus。
- AI 接觸／淘汰已有較高階可玩狀態但非原版方向矩陣；現行存活 AI pair 視為
  通過 `+0x584／+0x8B2` 外圈守門。
- 外交亂數使用可存檔確定性流，只保證 remake 重播一致。

## 驗收

- 純規則測試固定驗證政府表、難度 gate、同盟／研究、貿易、納貢與 bias 回復。
- shell 測試以注入 roller 驗證 raw 矩陣接線；高關係不得觸發舊 `+12` 停戰。
- 完整 `internal/gamedata`、`internal/shell` 回歸通過。
- 存檔與熱座測試涵蓋新增方向矩陣。
