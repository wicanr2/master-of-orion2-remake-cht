# 新遊戲開局規則規格

> 狀態：本規格範圍已閉合，2026-08-25。原版 AI 常態回合的難度分支屬
> `Compute_AI_Data_` 等 AI／經濟規格，不再混入開局規格。

## 證據輸入

- 原版執行檔：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具與位址：IDA Pro 9.4；DOS/4GW LE object #1 的 IDA 線性位址。
- 開局科技：`Init_Player_Tech_ @ 0x5E55F`。
- 母星建築：`Init_Homeworld_Colony2_ @ 0x13A3D`，上限表 `byte_13A3A`，
  優先表 `word_17D8AC`。
- 難度與開局等級全域：`byte_199CB0`（難度 `0..4`）與
  `byte_199CB5`（開局文明等級 `0..2`）。
- 詳細原始指令與工程日誌：[`docs/re/01-gap-report.md`](../re/01-gap-report.md)、
  [`docs/re/00-orion2-symbols.md`](../re/00-orion2-symbols.md)。

2026-08-24 的 `audit_newgame_rules.py` 初次執行因舊映像授權失敗，未作證據；2026-08-25
已以可用的 IDA Pro 9.4／IDAPython 映像完成後續唯讀稽核。新增證據分別見
[`starting-tech-application-audit-20260825.md`](../re/starting-tech-application-audit-20260825.md)、
[`ai-starting-tech-profile-audit-20260825.md`](../re/ai-starting-tech-profile-audit-20260825.md) 與
[`human-starting-tech-profile-audit-20260825.md`](../re/human-starting-tech-profile-audit-20260825.md)。

## 已證實契約

### 五級難度是離散設定

1. 新遊戲選擇器有五個難度項目，對應索引 `0..4`。
2. 原版子系統直接讀難度索引並各自分支。已取回的例子包含：
   - AI 性格欄位：`Random(10)+1-difficulty`，夾在 `1..10`。
   - 地面戰：以普通為基準的整數偏移，且只在原版指定陣營套用。
   - 安塔蘭：難度影響資源累積率、特定觸發門檻與單艦裝甲，不改變艦數上限表。
3. 沒有證據支持把難度映射成單一浮點倍率，再同時乘進敵艦強度與外交
   關係。因此 remake 不得保留 `0.3/0.6/1.0/1.5/2.2` 這張調校表。

證據等級：五級索引與上述各已追回 consumer 為**已證實**。所有常態 AI 難度效果並非
本規格完成條件，仍由 AI／經濟矩陣追蹤。

### 開局科技

1. 三級開局分別執行 `1 / 6 / 25` 次科技發放。
2. 前六次由固定表取值；先進級的後十九次每次都從當下可研究集合重新選擇。
3. 隨機選擇的粒度是科技應用，不是把同一主題的所有互斥應用全部解鎖。
4. remake 必須以開局種子提供決定性亂數；玩家與各 AI 使用分離亂數流。

證據等級：迴圈次數、固定／隨機分流、可研究門、應用粒度、`sub_589D6` raw 6／4／7、
`sub_FC845` 真人／AI／共用估值鏈與 `sub_FD335` 單次加權抽選均為**已證實**。原版符號中的
personality／objective／theme 無法證明三個 raw 欄位的英文語意，因此只保存 raw 值；這不阻擋行為實作。

### 開局母星建築

1. 依 `word_17D8AC` 的原始優先順序逐項檢查，只收已知科技允許的建築。
2. 名額上限為曲速前 `3`、一般 `5`、先進 `9`。
3. 另套用 `min(ceil(2/3 * population), cap)`。開局人口 `8` 時的人口上限為 `6`。
4. 曲速前與一般級因已知科技只能通過星基與海軍陸戰隊營，實際均為兩棟。

證據等級：優先表、上限表、科技門與人口夾限為**已證實**。

## 版本規則

- patch 1.31 與 1.50 共用本規格的五級難度索引、建築優先表與開局迴圈結構。
- 開局選擇若觸及版本相異的科技成本，必須由 `RuleProfile` 取值；目前十九次開局選擇
  不應抵達 Hyper-Advanced 主題，因此兩版在現行已證閉開局範圍應產生相同結構。
- 測試必須分別以 `Profile13()` 與 `Profile15()` 建立開局，不以選單可切換代替規則驗證。

## 實作契約

1. `Difficulties` 只保留五個顯示項目，不提供通用 `Mult`。
2. 敵艦 runtime blueprint 與外交關係不再乘上自編難度倍率。
3. 難度只能進入已有原版證據的 typed helper，例如 `pickAIPersonality`、
   `GroundDifficultyBonus` 與 `RebellionDifficultyAdjust`。
4. AI 常態經濟、建造、研究與外交難度分支由各自規格追蹤，不以新的通用倍率補洞。

## 驗收

- 五個難度顯示項保持固定索引，且沒有 `Mult` 欄位或浮點倍率 consumer。
- 同回合敵艦建構不因難度直接縮放；AI 性格仍依原版離散公式受難度影響。
- 三級開局科技數為 `1/6/25`，先進級同種子可重現，不同帝國使用分離亂數流。
- 曲速前／一般母星實際只有兩棟合法建築；科技全解的人口 `8` 母星最多六棟。
- 1.31 與 1.50 各自完成開局存檔往返，`Version`、`Difficulty`、`TechLevel`、
  `CompletedTopics`、`ChosenTech`、`ExplicitChoice` 與殖民地建築均不丟失。
