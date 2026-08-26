# AI 常態研究選擇稽核（2026-08-26）

## 證據身分

- 輸入、資料庫、工具、位址基準與非破壞性方法同
  [`antaran-marines-caller-audit-20260826.md`](antaran-marines-caller-audit-20260826.md)。
- 本輪直接追查：`sub_DCA69 @ 0xDCA69`、`sub_DC288 @ 0xDC288`、
  `sub_FD335 @ 0xFD335` 與 `sub_FC845 @ 0xFC845`。
- `sub_FD335`／`sub_FC845` 的完整估值分支沿用既有
  [`ai-starting-tech-profile-audit-20260825.md`](ai-starting-tech-profile-audit-20260825.md)；
  本輪補的是常態回合 caller 與觸發條件。

## 已證實的常態回合鏈

1. `Next_Turn_Calc_ @ 0x136B3` 在 `0x13724` 呼叫 `sub_DCA69`。
2. `sub_DCA69` 先以 stack 上的 0xA8-byte block 設定全域暫存指標，呼叫 `sub_DC47C` 建立
   AI 資料；之後由最高 player slot 往下掃。
3. inactive player 與 `player+0x28 == 100` 的真人 slot 被跳過；每個其餘 AI 都在
   `0xDCAAF` 呼叫一次 `sub_DC288`。
4. `sub_DC288` 讀目前 field `player+0x321`。該 field 的狀態已完成，或
   `player+0x323 != 0` 時，清除 change flag 並呼叫 `sub_FD335`。
5. `sub_FD335` 回傳的是 application ID；`0xDC2CC` 寫 `player+0x322`，接著由 application
   table 反查 field，於 `0xDC2DE` 寫回 `player+0x321`。
6. field `>=75` 時另設 `player+0x59D`，屬 Hyper-Advanced／晚期科技鏈。

所以原版常態 AI 的原子決策是「從目前可用 application 中作一次估值加權抽選，選中的
application 同時決定 field」，不是 remake 目前的「依 Profile 從八個 field 隊首挑成本最低／
最高，再按 category 最高分挑 application」。

## 可直接重用的既有證據

`StartingOriginalApplicationPick` 已轉寫 `sub_FD335` 的 application 級候選、研究回合視野、
成本反比、難度／raw6 篩選與一次加權抽選；`OriginalAITechValueKnownSlice` 已轉寫
`sub_FC845` 的 raw6／raw4／raw7、種族特性、類別與可表示的共用後段。這些規則不是開局
專用；開局只是它的其中一個 caller。

## 仍需明示的近似

- 原版全域亂數流尚未與 remake 對齊；使用可存檔的研究亂數流，只保證 remake 重播一致。
- `sub_DC47C` 的 0xA8-byte AI data 尚未逐欄 typed；remake 使用本回合實際 `TotalResearch`
  作研究速度輸入。
- 舊存檔沒有 raw6／raw4／raw7 時不能捏造 profile，應保留既有設計性選擇作安全 fallback。
- Uncreative 原版會先把各 field 的可用 application 限縮成一項；現行 typed tree 沒保存整張
  application status byte 表，本輪不宣稱該候選集合逐位元一致。

