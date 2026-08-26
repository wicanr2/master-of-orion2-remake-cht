# AI 曲速場干擾器建造分數規格

## 原版錨點

- `Colony_Building_Score_` raw 45：`0xD0012 → 0xD05BD..0xD0614`。
- 星系 raw 45 bit writer：`sub_E5296 @ 0xE5296..0xE53CD`。
- 證據與分級見 `docs/re/ai-colony-build-selection-audit-20260826.md` 及
  `docs/re/evidence/ai-warp-interdictor-score-ida.json`。

## Remake 契約

1. 缺星系建築狀態或戰略壓力 context 時回到既有明示 fallback，不把零值冒充精確輸入。
2. priority gate 成立且沒有 ETA9 來襲時分數為 0。
3. 同星系已有己方 raw 45 時，分數只有 `floor(budgetFactor/2)`。
4. 尚無 raw 45 時，分數為
   `5×ETA9 + 2×treatyNear + 3×noPolicyNear + 4×warNear + extended`；非零時加
   Ruthless 1 點，再加 `floor(budgetFactor/2)`。
5. AI 從正常候選完成建築後，既有航線 consumer 必須立即讀到該建築；不得另存第二份效果旗標。
6. 本切片不新增玩家文案；未來若需提示，固定文字必須由 JSON／YAML 提供。

## 驗收

- 純規則測試覆蓋五個係數、priority／ETA9、既有同星系建築、Ruthless 與缺 context。
- 唯一候選測試從科技解鎖進入建造、完成並寫入 `ColonyBuildings`。
- 正常航線 consumer 抽樣確認三秒差距內敵艦降速；既有 route 測試仍須通過。
