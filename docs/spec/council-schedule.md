# 銀河議會召開排程規格

## 規則

1. 議會只有在以下條件同時成立時才能召開：
   - `Turn >= gamedata.CouncilFirstMeetingTurn`（25）；
   - 存續帝國至少 3；
   - 已殖民星球數至少為正整數 `totalStars / 2` 向下取整；
   - 沒有待玩家回應的上一屆結果，遊戲尚未結束。
2. 第一次符合條件時立即召開；不是在「剛達半數殖民」的任意早期回合立即召開。
3. 召開後記錄 `lastCouncilTurn = Turn`；下一屆最早在
   `Turn - lastCouncilTurn >= gamedata.CouncilMeetingInterval`（25）召開。
4. `CouncilMeetings`、`lastCouncilTurn` 與待回應結果沿用既有存檔欄位；存讀檔不得重置間隔。
5. 本規格不定義兩位候選人、外交投票、棄權或接受結果的精確原版公式。

## 邊界驗收

- 即使殖民與帝國數達標，Turn 24 不開會、Turn 25 開第一次。
- 第一次在 Turn 25 召開後，Turn 49 不開、Turn 50 開第二次。
- 5 顆星時 1 顆已殖民不成立、2 顆成立；24 顆星時仍須 12 顆。
- 快照還原後保留上次召開回合，不能提早開下一屆。

逆向證據見 `docs/re/council-schedule-audit-20260824.md`。
