# AI 開局科技 raw profile 與估值稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4／IDAPython，映像 `fd2-ida-authorized-local:latest`
- 位址基準：IDA 線性位址，DOS/4GW LE object #1
- 非破壞性匯出：`tools/ida/audit_calc_tech_value.py`

## 已證實：`sub_589D6 @ 0x589D6..0x58F1E`

1. 先以 `byte_181090[origRace*10+column]` 寫入 `[player+0x27]`；`column = clamp(Random(10)+1-difficulty, 0, 9)`。該 13×10 bytes 與受版控的經典種族十格表逐字節一致。
2. 接著建立三組權重並抽選，寫入 `[player+0x28]`、`[player+0x205]`、`[player+0x206]`；候選數為 6／4／7，基礎權重分別是 `1,2,1,2,2,1`、`1,2,2,1`、`2,2,1,2,1,2,3`。
3. 三組權重直接消費 `[player+0x8A0..0x8BD]`，即 `+0x89F` 31-byte 種族特性陣列的索引 1..30；難度 3／4 另增加六項表的 raw 4／2／3。
4. 任一組總權重超過 1000 時，原版會對整組反覆做有號除 2，直到總和不超過 1000 才抽選。
5. `[player+0x27] == 0` 時，七項表 raw 2 權重加 3。人類玩家只保留 `[+0x28]=100`，四項與七項表仍會抽選。

## 已證實：`sub_FC845 @ 0xFC845..0xFD199`

- AI 先以 category 靜態值開始，再依 raw 4、raw 7、raw 6 三個 switch 覆寫為 `1/5/10/20/50/100`。
- category `0/1/2/3/4/6/0x0C/0x10/0x12/0x19/0x1B/0x1C/0x25/0x28` 還會依種族特性或政體覆寫。tech `5`、`0x83` 有獲立的特性門。
- 完成 AI 分支後，會進入與人類共用的等級基礎值、同類科技邊際修正、其他玩家類別上限與早期遊戲分支。
- `sub_FD335` 在難度大於 0 時依 raw 6 再篩候選：raw 1／2 排除小於最高分 `1/6`；raw 3／4 排除小於 `4/5`；raw 5 排除小於 `1/2`。

## 保留的不確定性

- 函式符號名列出 Personalities／Objectives／Themes，但名稱無法單獨證明 `+0x28/+0x205/+0x206` 各自對應哪個英文名詞。remake 因此只保存 raw 6／4／7，不把導覽名當事實。

## 2026-08-25 轉換時序勘誤

- **已證實**：`sub_589D6` 唯一 caller 是 `sub_12983 @ 0x12983`。後者在
  `0x12A8E..0x12AC5` 先複製兩份 31-byte RACESTUF 陣列，再在
  `0x12AE4..0x12B53` 以 `byte_17D1F9` 轉換索引 1..9，最後才於
  `0x12C4B` 呼叫 `sub_589D6`。
- **已證實**：`sub_589D6` 讀的是轉換後 runtime 真值。特性 6 的 20／40 立即數
  並非轉換前等級，也不否定 `OrigRaceTraits` 的 25／50；在 1.50 現行表下，
  正常艦防 25／50 不觸發這兩條權重分支。
- 上一版文件把它寫成「轉換順序尚未追到的矛盾」已過期；現行解釋是
  原版保留了對 20／40 的比較，但現行轉換表不產生這兩個正值。

## remake 對應

- `RollOriginalAIRaw27` 與 `RollOriginalAITechProfile` 保留 raw 抽選與千分比縮放。
- `OriginalAITechValueKnownSlice` 實作 AI 類別／特性分支並進入共用後段。
- 正常先進級開局以 AI 原版種族索引建立 profile，三個 AI 使用獨立可重播亂數流。

## 2026-08-29 Stealthy Ships 補證

第一組六候選權重的 stack 順序已追回；`player+0x8BB == 1` 在 `0x58D5D` 對候選 1 加 100，
最後由同組抽選寫入 `player+0x28`。`Calc_Tech_Value_` 則只在非人類 AI 的 category `0x25`
把表定 multiplier 5 覆寫為 1；該 category 的四項 tech ID 為 38／53／126／172，即
Cloaking Device、Displacement Device、Phasing Cloak、Stealth Field。這不是單一 Stealth Field
特例，也不是最終科技價值固定為 1。證據見
`evidence/stealthy-profile-tech-ida-20260829.json`。
