# 星系封鎖規格

規格狀態：`.GAM` typed 保存 `CONFORMED`；正常回合 producer 與職務 consumer `DRAFT`。

證據見 [`../re/blockades-audit-20260828.md`](../re/blockades-audit-20260828.md)。

## Typed 契約

- `Star.BlockadedMask` 的 bit `p` 表示 player slot `p` 在該星遭封鎖。
- `Star.BlockadedBy[p]` 的 bit `q` 表示 fleet owner `q` 正在封鎖 player `p`。
- `.GAM` 匯入必須逐 byte 保存兩者，不得把非零值壓成 bool。
- 正常回合重算必須先整表清除，再只由有效、已抵達的艦隊重建。

## 尚未 READY

1. shell 中玩家、熱座與 AI 對原版 player slot 的完整映射，以及多帝國同星殖民 mask。
2. raw policy 6 的正式語意；判斷可保留 `4..6` raw 契約，不可假稱只有既有 enum 4／5。
3. owner≥8 艦隊在 remake 的 typed 表示。
4. 封鎖關係懲罰的共用 `sub_4E3B5` 輸入接線。
5. `sub_D61E7` 所需 `colony+0xE0` producer 與事件 filter 分支。
