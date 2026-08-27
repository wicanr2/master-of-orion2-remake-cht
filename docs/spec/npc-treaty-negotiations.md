# NPC 對 NPC 外交談判規格

**狀態：CONFORMED（remake 可表示的 AI ordered-pair 切片）**

證據：[`../re/npc-treaty-negotiations-audit-20260827.md`](../re/npc-treaty-negotiations-audit-20260827.md)。

## 狀態

每個 AI ordered pair 保存：

- `policy`、`tradeActive`、`researchActive`；
- signed-byte `reputationRaw`（原版 `+0x6D7`）；
- signed-word `treatyBiasRaw`／`agreementBiasRaw`（`+0x68F／+0x69F`）；
- direction-only `tributeMode`（`+0x63F`）。
- 待處理外交 `reason／magnitude` 與重複事件記憶（`+0x64F／+0x65F／+0x71F`）。
- 單向永久違約記憶（`+0x727`）及締約／納貢對 `+0x72F` 的政府表增量。
- 戰爭計時與停戰冷卻另依
  [`npc-war-ceasefire.md`](npc-war-ceasefire.md) 保存 `+0x717／+0x72F`。

所有矩陣必須隨 JSON 存檔與熱座 AI 索引壓縮；舊存檔缺欄以原版初始化值 0
補入。

## 回合順序

1. 依 `sub_252D5` ordered pair 推進事件記憶，再把結果鏡射至反方向。
2. 所有方向的 treaty／agreement bias 各 `+10`，正值夾回 0。
3. 依 outer 低到高、inner 低到高掃 ordered pair；self 跳過。
4. 每對先擲 `Random(250-40*difficulty)`，只有 1 才進行談判。
5. 依證據文件的 base、treaty、agreement、tribute 分數及原始門檻處理。
6. 通過頻率 gate 的 pair 不論是否建約，兩個 bias 最後各減 30。
7. 正式戰爭只由 policy／war state 消費；不得再用顯示關係 `-25` 宣戰或
   `+12` 自動停戰。

## 資料模型投影

- `+0x5EC` 一般 AI 實艦方向國力已接；只有缺 raw 實艦的舊存檔使用明標非精確純量回退。
- `+0x64F／+0x65F／+0x71F` 已訂正為 pending reason／magnitude／重複事件記憶；一般政府
  更新、鏡射與第三方 +5 已接。一般宣戰的 `+0x727` 方向 writer、government 4 特殊門檻及
  締約／納貢 cooldown 表已接；其餘 reason／玩家回應 writer 尚未閉合。
- AI 接觸／淘汰已有較高階可玩狀態但非原版方向矩陣；現行存活 AI pair 視為
  通過 `+0x584／+0x8B2` 外圈守門。
- 外交亂數使用可存檔確定性流，只保證 remake 重播一致。

## 驗收

- 純規則測試固定驗證政府表、難度 gate、同盟／研究、貿易、納貢與 bias 回復。
- shell 測試以注入 roller 驗證 raw 矩陣接線；高關係不得觸發舊 `+12` 停戰。
- 完整 `internal/gamedata`、`internal/shell` 回歸通過。
- 存檔與熱座測試涵蓋新增方向矩陣。
