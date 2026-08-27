# NPC 特殊宣戰規格

規格狀態：`CONFORMED`（理由 20／68／113、超空間亂流 gate 與明列的 raw
輸入範圍；`+0x60E` runtime producer 與理由 22 不在本規格已完成範圍）。

證據來源與未閉合邊界見
[`../re/npc-special-war-policy-audit-20260828.md`](../re/npc-special-war-policy-audit-20260828.md)。

## 狀態

- `AIOpponent.OriginalFoodDeficitTurns` 保存原版 `player+0x7EC`。
- 每次 AI 帝國經濟結算後，以 `TotalFoodHalf < 0` 執行原版 `inc word`；
  非赤字立即歸零，32767 依 signed 16-bit 二補數回繞為 -32768。
- `AIOpponent.OriginalWarFlag60ERaw` 無損保留原版 `.GAM` Player
  `+0x60E`；消費端已證實，runtime producer 仍未知，新局固定為零。
- 欄位透過既有 `AIOpponent` JSON 路徑存讀，舊存檔缺欄時從零開始。

## 候選順序

對每個目前沒有 AI 戰爭的來源，必須依下列順序各自掃描全部目標，不得改成
「一個目標跑完所有理由」：

1. 超空間亂流 active 且來源不是跨維度種族時，跳過整條候選鏈。
2. raw government 3 的理由 20。
3. 輪值敵意門檻的理由 68。
4. `.GAM` `+0x60E == 1` 的無亂數理由 113。
5. 每個來源只擲一次的食物赤字理由 113。
6. 既有一般政策／國力理由 23。
7. 難度至少 3 的第三方戰爭目標 veto。
8. 從 type-1 候選中均勻抽一個，交給既有正式宣戰 consumer。

理由 113 即使赤字回合數為零或沒有合格目標，也必須消耗一次 `Random(100)`，
以維持後續外交亂數位置。理由 20、68 與 23 則只在各自前置守門通過後擲骰。

## 失敗邊界

非法難度、政府、signed relation、國力、赤字回合或亂數輸出採失敗即關閉
（fail-closed）。`+0x60E` 只可來自原版存檔或 remake 快照；`+0x6FF`
未有可靠 session producer 前不得以 personality 或自編常數代替；外層事件 gate
必須重用既有 `PersistentHyperspaceFlux` 與跨維度 trait。
