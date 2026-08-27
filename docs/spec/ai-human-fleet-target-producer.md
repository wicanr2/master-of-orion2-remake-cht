# AI 對真人艦隊目標 producer 規格

狀態：部分 CONFORMED；完整決策分數為 DRAFT。

## 已閉合契約

1. 每個 AI 保存可序列化的真人目標 decision cooldown；舊存檔缺欄視為 0。
2. cooldown 大於 0 的世界回合只遞減，不建立新真人目標；歸零後下一回合才重新評估。
3. 新真人目標成功派艦後寫入 `Random_(20)+20`，範圍固定為 20–39。
4. 已有正式和平／互不侵犯／同盟時不得進攻。`sub_53EDB` 本身排除已與真人正式開戰者；
   remake 單主力模型尚未閉合另一條戰時軍事 producer，因此已開戰派艦暫由同一 adapter 承接。
5. 固定開局寬限、固定軍力倍率與 losing-ground personality 表不得再作本 producer 的原版門檻。
6. `LastRaidTurn` 的 10 回合限制只屬單艦隊結算 adapter，不得混同 `player+0x816`。

## DRAFT 邊界

`sub_544A1 @ 0x544A1..0x54CC0` 的 directional incident memory、排名／科技趨勢、四類結果與
多次 RNG 尚未形成完整 typed input。這些欄位閉合前，remake 可用既有戰爭態勢決定是否呼叫
原版目標估值，但必須標為 fallback，不得增添新的常數門檻。

## 驗證

- cooldown 2 必須連續阻擋兩次評估，第三次才可派艦。
- 成功派艦後 cooldown 必須落在 20–39，並經 JSON snapshot 往返不變。
- 玩家軍力較高或 personality losing-ground chance 為 0，不得單獨形成固定 veto。
- 艦隊抵達後的 formal policy 5／6 writer 與 10 回合重複結算保護維持既有測試。
