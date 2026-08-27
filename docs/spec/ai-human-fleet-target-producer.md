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
7. AI→真人接觸計時對映 `player+0x88F`：只在接觸成立後每回合加一、封頂 250；未滿 10
   時不得建立真人目標。現行單人 adapter 尚未拆 contact bitset，因此把可外交 AI 視為已接觸。
8. personality score 表固定為 `[-10,-5,-3,0,20,20,-10]`。完整 score 傳入尾端後：
   - score≥0：`threshold=contactTurns/isqrt(score³+5)`；
   - score<0：`threshold=contactTurns*(-score)`；
   - RNG 順序固定為 `Random(3)`、`Random(100)`，通過後再依路徑消耗 `Random(16)` 與
     一至兩次 `Random(4)`；不得因 gate 結果重排。
9. `sub_4F93B` 外交 action selector 使用 kind 1=BC、2=科技、3=殖民地、4=直接要求；
   intensity 上限 10，候選 RNG 順序為 `3→3→2`，科技 payload 另消耗 `Random(3)`。
   BC 金額小於 100 時向下取整到 10，否則向下取整到 100，最後封頂 32000。
10. Honorable source（personality 4）若記錄 target 曾破壞正式條約 `+0x727==1`，base score
    必須改用 personality index 6 的 -10，不得仍使用 Honorable +20。玩家 `break_formal`
    是單向永久 writer；`break_trade`／`break_research` 不是。
11. `sub_4F0DC` 只把範圍 1..9 的 pending reason `+0x64F` 複製至 remembered reason
    `+0x6CF`。`sub_544A1` 在 `+0x71F>0` 且 `+0x6CF!=0` 時計算
    `-10*memory/word_180CF0[personality]`；表值固定為 `[1,2,3,3,4,5,2]`，相符的
    玩家可見原因碼為 `rememberedReason+70`。
12. 存活帝國少於三個且來源 AI 的 `player+0xA6` 人口嚴格高於其他存活帝國時，score
    加 `-10`；只剩來源一國也加 `-10`。三國以上不套此項。
13. 相對回合至少 100 後，比較雙方目前與 40 格前的 `+0xB9B` 人口歷史。只有真人人口
    成長大於來源 AI 時，加 `(sourceGrowth-targetGrowth)/2`，採 signed 整數向零截斷。

## DRAFT 邊界

`sub_544A1 @ 0x544A1..0x54CC0` 的四類尾端與 RNG 已形成純規則；尚缺的是 directional
incident memory 的 writer 門檻／正常玩家事件 producer、其餘國力／科技項，以及 `sub_4F93B`
的科技候選表／殖民地候選 producer。
這些欄位閉合前，remake 可用既有戰爭態勢決定是否呼叫原版目標估值，但必須標為 fallback，
不得把只含 relation/personality 的部分 score 冒充完整 producer。

## 驗證

- cooldown 2 必須連續阻擋兩次評估，第三次才可派艦。
- contact turns 由 249 增至 250 後封頂，並與 cooldown 一起通過存檔往返。
- 正負 score threshold、四類結果及每條路徑的 RNG 次數必須有純規則測試。
- 外交 action 四種 kind、候選 gate、RNG 順序、BC 取整及科技索引必須有純規則測試。
- 正式解約寫 +0x727、經存檔往返及 Honorable score 差 30；經濟協議解約須有負對照。
- 成功派艦後 cooldown 必須落在 20–39，並經 JSON snapshot 往返不變。
- 玩家軍力較高或 personality losing-ground chance 為 0，不得單獨形成固定 veto。
- 艦隊抵達後的 formal policy 5／6 writer 與 10 回合重複結算保護維持既有測試。
