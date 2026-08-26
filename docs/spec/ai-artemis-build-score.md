# AI Artemis System Net 建造分數規格

## 原版錨點

- raw 3 jump entry：`0xCFF6A → 0xD054E..0xD05B8`。
- raw 45 星系 bit writer：`sub_E5296 @ 0xE5296..0xE53CD`。
- 證據：`docs/re/evidence/ai-artemis-build-score-ida.json` 與
  `docs/re/ai-colony-build-selection-audit-20260826.md`。

## Remake 契約

1. 缺戰略壓力或同星系 raw 45 狀態時走明示 fallback，不把零值當精確資料。
2. priority gate 成立且沒有 ETA9 來襲時分數為 0。
3. 同星系已有己方 raw 45 時只給 `floor(budgetFactor/2)`；不得改成檢查 raw 3 自身。
4. 尚無 raw 45 時分數為
   `10×ETA9 + treatyNear + noPolicyNear + 3×warNear + extended`；非零時加 Ruthless 1 點，
   再加 `floor(budgetFactor/2)`。
5. 正常科技候選完成後寫入唯一 `ColonyBuildings` 權威；玩家與 AI 艦隊抵達 consumer 都從
   raw ID 3 查詢，不保存內嵌中文建築名稱或第二份效果旗標。
6. 本切片不新增玩家文案；固定文字如有新增必須來自 JSON／YAML。

## 驗收

- 純規則測試覆蓋五係數、priority／ETA9、同星系 raw 45、Ruthless、budget 與缺 context。
- 唯一正常候選完成後，玩家艦隊與 AI 艦隊的既有 Artemis consumer 均能讀到建築。
- 既有水雷觸發、護盾減傷、沉船與決定性測試不得回歸。
