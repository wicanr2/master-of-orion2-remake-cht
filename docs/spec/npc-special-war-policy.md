# NPC 特殊宣戰規格

證據來源與未閉合邊界見
[`../re/npc-special-war-policy-audit-20260828.md`](../re/npc-special-war-policy-audit-20260828.md)。

## 狀態

- `AIOpponent.OriginalFoodDeficitTurns` 保存原版 `player+0x7EC`。
- 每次 AI 帝國經濟結算後，以 `TotalFoodHalf < 0` 遞增；非赤字立即歸零；
  signed word 正向範圍在 32767 飽和。
- 欄位透過既有 `AIOpponent` JSON 路徑存讀，舊存檔缺欄時從零開始。

## 候選順序

對每個目前沒有 AI 戰爭的來源，必須依下列順序各自掃描全部目標，不得改成
「一個目標跑完所有理由」：

1. raw government 3 的理由 20。
2. 輪值敵意門檻的理由 68。
3. 每個來源只擲一次的食物赤字理由 113。
4. 既有一般政策／國力理由 23。
5. 難度至少 3 的第三方戰爭目標 veto。
6. 從 type-1 候選中均勻抽一個，交給既有正式宣戰 consumer。

理由 113 即使赤字回合數為零或沒有合格目標，也必須消耗一次 `Random(100)`，
以維持後續外交亂數位置。理由 20、68 與 23 則只在各自前置守門通過後擲骰。

## 失敗邊界

非法難度、政府、signed relation、國力、赤字回合或亂數輸出採失敗即關閉
（fail-closed）。`+0x60E`、`+0x6FF` 與外層事件 gate 未有可靠 session producer
前不得以 personality、事件名稱或自編常數代替。
