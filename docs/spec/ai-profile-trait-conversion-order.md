# AI raw profile 種族特性轉換順序規格

狀態：CONFORMED

RE-TRACE: dos-orion2-1.31:0x589D6

CORRECTION-20260830-PROFILE-W4：IDA 完整函式匯出證實四候選初始權重為
`[2,1,2,1]`，不是現行 source 的 `[1,2,2,1]`。Ship Defense `20／40` 應加 raw ID 1，
Ship Attack `25／50` 應加 raw ID 0；現行 source 兩者 index 對調。因此「已有 profile
產生器與綠色總和測試」只能判為部分實作，尚未 CONFORMED。

## 原版契約

1. `sub_12983 @ 0x12983` 從 `RACESTUF.LBX` asset 7 複製 31-byte 選項等級到
   `player+0x89F` 與鏡像區 `player+0x8BE`。
2. 對索引 `1..9` 逐格以 `byte_17D1F9` 轉成 runtime 真值；索引 0 與 10..30
   保持 raw 值。
3. 所有玩家轉換完畢後，才在 `0x12C4B` 呼叫 `sub_589D6`。
4. `sub_589D6` 因此必須消費已展開的 runtime 特性，不可傳 RACESTUF 選項等級。
5. 原版對特性 6 的 20／40 比較保留原樣。現行 1.50 轉換表對該特性產生
   `-20/25/50`，所以正常內建種族的 `25/50` 不觸發 20／40 分支。remake 不得
   把比較值改成 25／50 來「修正」原版。

## remake 對應

- `OrigRaceTraits` 是已展開的 runtime 陣列，直接傳入 `RollOriginalAITechProfile`。
- 測試必須同時鎖住：合成 runtime 20 會觸發原始分支；真實表的 25／50
  不得被當成 20／40。
- 未來若加入其他 patch profile，須分開保存該版本的選項等級表與 runtime
  轉換表，不可混用 1.50 轉換結果。
- 三組初始權重固定為 `[1,2,1,2,2,1]`、`[2,1,2,1]`、`[2,2,1,2,1,2,3]`；
  trait 加權逐格表以 `ai-trait-profile-tech-homeworld-audit-20260830.md` 為唯一證據來源。
- 驗收不得只比較三組總和；至少逐格驗證四候選初值，以及 Ship Defense／Ship Attack
  對 raw ID 0／1 的相反加權。

證據見 [`../re/ai-starting-tech-profile-audit-20260825.md`](../re/ai-starting-tech-profile-audit-20260825.md)
與 [`../re/ai-trait-profile-tech-homeworld-audit-20260830.md`](../re/ai-trait-profile-tech-homeworld-audit-20260830.md)。
