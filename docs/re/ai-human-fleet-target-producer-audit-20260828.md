# AI 對真人艦隊目標 producer 稽核

日期：2026-08-28

## 證據契約

- `Orion2.exe` SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- `.i64` SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- IDA Pro 9.4；DOS/4GW LE image、IDA linear EA；只分析正式資料庫的可寫副本。
- raw instructions／bytes／xrefs：
  [`evidence/ai-human-fleet-target-ida-20260828.json`](evidence/ai-human-fleet-target-ida-20260828.json)。
- Hex-Rays 輸出只供導覽；變數名與型別不是證據。

## 已證實

1. `sub_53EDB @ 0x53EDB..0x544A1` 是正常每回合 AI 對真人決策 producer；由主回合鏈
   `sub_136B3 @ 0x136B3..0x13822` 的 `0x13706` 呼叫 `sub_252A7`，再於 `0x252A7`
   呼叫 `sub_53EDB`。它不是 `sub_DB257` 的軍事接戰階段。
2. producer 排除 `player+0x28==100` 的 source，只處理未淘汰 AI；僅在 `player+0x7C7==-1`
   且沒有任何 human target 的 formal policy `>=4` 時掃描真人候選。
3. 真人候選必須仍有效、未進 formal war、具備接觸／可互動狀態，且不能處於三種已占用的
   `player+0x74F` 狀態。精確欄位語意仍有部分未知，故不以推測名稱取代 raw offset。
4. `sub_544A1 @ 0x544A1..0x54CC0` 以 directional relation `player+0x617`、personality class、
   外交事件記憶、接觸狀態、雙方排名／科技趨勢、難度及多次 `Random_` 決定四類結果。
   舊 `aiRaidGraceTurns=12`、`aiRaidStrengthMargin=125` 與
   `PersonalityLosingGroundChance` 擬亂數不在此原版鏈中。
5. 類型 2 成功時，`sub_53EDB` 只有在 personality class 不等於 4、候選 target 存在、
   `player+0x857` 估值高於門檻且 `player+0x847>0` 時，才把 target 寫入
   `player+0x7C7`、任務碼寫入 `player+0x7C9`，並把 `player+0x816` 寫成
   `Random_(20)+20`，即 20–39 回合。
6. `sub_DB47E @ 0xDB47E..0xDB659` 是之後的 AI 軍事階段；它依序準備目標表、呼叫
   `sub_D7764`、再呼叫 `sub_DB257`。`sub_DB257` 讀 `+0x7C7/+0x7CA`，在目標矩陣 entry
   非零時呼叫 `sub_51078`，屬接戰／宣戰 consumer，不是目標 producer。
7. `sub_4DAB2 @ 0x4DAB2..0x4DD6B` 是三個 gate 的每回合 producer：`player+0x816`
   非零就減一；AI→真人方向在 `player+0x584` 接觸 bit 成立時，`player+0x88F` 每回合加一、
   封頂 250。`player+0x74F` 則每回合由方向 formal policy `+0x627` 鏡射，不是 personality。
8. `word_181080 @ 0x181080` 七個 signed word 為 `-10/-5/-3/0/20/20/-10`，索引是
   source personality class。`sub_E5E09 @ 0xE5E09..0xE5EB3` 掃 target 的有效領袖，取
   Diplomat 一般 `10*(level+1)` 或進階 `15*(level+1)` 的最大值。
9. `0x54A27..0x54A87` 的敵意機率 threshold 已由 raw instructions 閉合：score 非負時為
   `contactTurns/isqrt(score³+5)`，score 負時為 `contactTurns*(-score)`。尾端先消耗
   `Random_(3)` 建立 action count，再消耗 `Random_(100)`；即使接觸未滿 10，這兩次 RNG
   仍已發生。通過後依序選回傳 2、4、3 或 1；Repulsive gate 位於 type 2／4 之後。
10. `0x5499F..0x549CF` 先讀 source personality；只有 personality 4（Honorable）且
    AI→真人方向永久違約旗標 `+0x727==1` 時，改以 personality index 6（Dishonored）的
    `word_181080=-10` 取代 index 4 的 `+20`。原版 `sub_5138E` 在 actor 破壞既有正式條約時
    寫入對方看向 actor 的 `+0x727`；普通貿易／研究終止不屬此 writer。
11. `sub_4F0DC @ 0x4F0DC..0x4F59B` 是 AI 對真人方向 pending 關係事件的消費端。
    `0x4F333..0x4F34B` 只在 `+0x64F` 為 1..9 時複製至 `+0x6CF`；全庫 operand 掃描顯示
    `+0x6CF` 的唯一 runtime writer 正是 `0x4F34B`。`sub_544A1 @ 0x54524..0x5457A`
    在 signed `+0x71F>0` 且 `+0x6CF!=0` 時計算 `-10*memory/divisor`，除數重用
    `word_180CF0 @ 0x180CF0 = [1,2,3,3,4,5,2]`；若該負分成為目前最低項，原因碼寫成
    unsigned `+0x6CF + 70`。上述資料流與表 bytes 已證實；Hex-Rays 名稱僅供導覽。
12. `0x547DE..0x54889` 掃所有未淘汰帝國的 `player+0xA6`：存活數少於三且來源人口
    嚴格大於其他存活帝國，或只餘一國時，score 加 `-10`、原因碼候選為 178。
    `0x5488B..0x54949` 自相對回合 100 起讀雙方 `+0xB9B` 的目前格與 40 格前；真人成長
    嚴格較高時，加入 `(sourceGrowth-targetGrowth)/2`，原因碼候選為 117。350 格之後原版
    以 309 作 40 格前的環回索引；remake 的時間排序 350 格環保存玩家可見等價資料。
13. `sub_500CF @ 0x500CF..0x50164` 讀雙向 `player+0x5EC`，計算
    `100*(sourceToTarget+1)/(targetToSource+1)`、夾至 800，來源對每個非雙方的正式戰爭
    再右移一位。這與既有 `OriginalNPCPowerRatio` 完全相同。`0x54768..0x547DC` 在 ratio
    至少 300 且來源政府 raw !=5 時加入 `-ratio/40`、把行動上限設為 150。ratio<300
    後的 raw 人口比較是 `2*sourcePopulation < 3*sourcePopulation` 時跳過，正常正人口下
    不形成第二條 gate；不得把 Hex-Rays 的同值比較誤寫成雙方人口公式。
14. `sub_E5B17 @ 0xE5B17..0xE5B69` 掃 target 所有非 outpost 殖民地，呼叫
    `sub_E0C1D` 加總種族人口容量；它不是艦力 helper。`0x54593..0x545CE` 在 source
    `+0x60E==1` 且 source 類型 raw 2，或來源總人口 `+0xA6` 大於 target 容量一半時，
    直接把 score 設為 -150、原因碼 114。
15. `sub_DCB47 @ 0xDCB47..0xDCBB0` 掃殖民地記錄的 player mask：同時含 source／target
    的殖民地加 5，只有 target 且 `sub_FF666` 對 source 可達則加 1。government raw 0
    分支以此總數作 `Random(400)` 門檻；資料形狀已證實，殖民地可達語意尚未 typed。
16. `0x545CE..0x5470C` 的其餘特殊分支已由 raw comparison 閉合：government 3 使用
    `Random(200)<=difficulty+1` 並覆寫 score=-150／原因 109／行動上限 100；所有政府都
    無條件先消耗 `Random(100)`，只有結果嚴格小於 signed `+0x7EC` 食物赤字回合時覆寫
    -150／原因 119。government 1 在 `sub_500CF>=100`、方向 `+0x857>=200` 且對應
    `+0x837!=-1` 時加入 `-value/20`／原因 115。government 0 則在
    `Random(400)<=sub_DCB47` 時覆寫 -150／原因 121。
17. 全庫 `+0x7EE／+0x7F6` operand 掃描顯示初始化之外，runtime writer 只在
    `sub_5138E @ 0x5138E..0x515F8` 的正式違約鏈。受害者及已接觸第三方看向 actor 的
    signed `+0x7EE` 一般減 10，observer personality raw 4 則減 20；`+0x7F6` 記錄受害
    player slot。`sub_544A1` 加入 `3*signed(+0x7EE)/5`，受害 slot 等於 source 時原因 177，
    否則 176。普通貿易／研究終止沒有這個 writer。
18. `player+0x28` 已由既有 `sub_589D6` 證據保存為 `OriginalAITechProfile.Raw6`；此次只
    增加獨立 known 狀態。GAM parser 的 `save.Player.Objective` 正是該 raw byte，但因
    `+0x205／+0x206` 未解析，仍不把整份 profile 標為 known。
19. `sub_DCB47 @ 0xDCB47..0xDCBB0` 的 raw prologue 先保存 `eax` source，再由 `dl` 建 target
    mask；因此 Hex-Rays 顯示的未初始化 `v6` 是反編譯假象。`sub_FF666 @ 0xFF666..0xFF68A`
    讀 source `player+0x324` 後呼叫 `sub_FF5F8 @ 0xFF5F8..0xFF666`；後者先檢查 source，
    再檢查 source 對其正式聯盟 raw 2 的其他 player。
20. `sub_FF4E9 @ 0xFF4E9..0xFF593` 對含指定 player 人口 mask 的殖民地，以原版星圖座標
    判斷 `dx²+dy² <= 900*fuelRange²`；`sub_FF593 @ 0xFF593..0xFF5F8` 另有 target 蟲洞
    `+0x29` 與 partner star mask `+0x33` 支線。無蟲洞距離契約已證實；蟲洞 mask 的完整
    語意仍標為 unknown，remake 不以 `Star.Owner` 猜測替代。
21. `sub_4F93B @ 0x4FB53／0x4FB7A／0x4FBA6` 的三次 raw comparison 都是
    `call sub_1247A0; cmp eax,1`。現行 selector 原先把它寫成 `r-1==1`，實際錯移為命中 2；
    `0x4FCDD..0x4FD17` 的科技索引也證實常數是 `Random(3)-2`。兩項已依 raw 指令勘誤。
22. `sub_4F93B @ 0x4FBFB..0x4FC23` 讀 source `player+0xB4`；獨立
    `sub_E2000 @ 0xE2000` 證據已證實這是建築、運輸艦、指揮赤字、間諜、納貢、軍官六項
    維護費總額，不是收入。remake 已改名 `SourceMaintenance`；有 AI→玩家納貢而缺本回合
    raw 分項時失敗即關閉。
23. `sub_27094 @ 0x27094..0x2720F` 掃 target 已知且 source 未知的科技，排除
    `Calc_Tech_Value_==0` 後依估值排序。`0x4F9C0..0x4FA78` 分別取雙方最高科技 level，
    target 較低時算 `10*target/source`，否則 10，再夾下限 1。typed producer 已接來源
    `OriginalTechProfile`、雙方已知 application 與既有 `OriginalAITechValueKnownSlice`。
24. `sub_E5CD4 @ 0xE5CD4..0xE5DE8` 建立真人殖民星候選；`sub_E5BE3 @
    0xE5BE3..0xE5CD4` 加總同星真人殖民地 `+0x0A` population byte，並用
    `sub_E5BD8 @ 0xE5BD8..0xE5BE3` 的 `uint8(a[1])-uint8(b[1])` 升冪排序。它排除
    首都建築 raw 9、不可達星及真人唯一殖民星；候選數的 `(n+1)/2` 是抽選半部大小。
    intensity<=6 取低人口半部，intensity>6 取高人口半部。上述 producer 與 payload 已接；
    `star+0x67` 暫時外交佔用槽在 remake 尚無獨立狀態，正常無 pending request 路徑等價全 -1。
25. outcome 拆段後確認原版 RNG 邊界：`powerRatio/40+Random_(3)-2` 先產生 intensity，
    `sub_4F93B` 再消耗 kind／payload，之後 `0x54B30` 才消耗 `Random_(100)`。尾端
    strongest 讀的是來源 `player+0xA6` 人口；`word_19A0E2` 已知在議會流程寫 1、後續外交
    流程可寫 2／3，因此只改稱 `CouncilStateIs1`，不以「開過議會」猜造 producer。
26. `sub_53EDB` 的 outcome 1／3／4 分別寫方向 reason 106／105／124，之後都呼叫
    `sub_54CC0 @ 0x54CC0`。後者將 `word_19B580／word_19B582／byte_19B584／byte_19B587`
    鏡射到雙方方向記錄；remake 現以 `OriginalHumanDiplomaticRequest` 保存 outcome、raw reason
    與 typed action，並與會談旗標及 JSON snapshot 同步。原版接受／拒絕 callback 尚未閉合，
    因此不從 request payload 推測後續資產／條約 mutation。
27. `word_19A0E2 @ 0x19A0E2` 的完整直接 xref 已閉合：`sub_4D78E @ 0x4DA90`
    初始化寫 0；`sub_15239 @ 0x152BA` 每次議會開始寫 1；`sub_15DF8 @ 0x15E37／0x15E42`
    在達成 2/3 多數時，真人當選寫 2、其他帝國當選寫 3，沒有勝者不覆寫而維持 1。
    remake 現以 known+raw 保存這個三態，接入既有議會開始／計票結果並通過 snapshot 往返；
    舊 JSON／GAM 缺 raw 時仍保守 unknown。

## Remake 對映與限制

- 已接：新增可持久化 `OriginalHumanTargetDecisionCooldown` 對映 `player+0x816`；大於 0
  時每世界回合遞減且不啟動新真人目標，成功派艦後依原版範圍寫 20–39。
- 已接：新增可持久化 `OriginalHumanContactTurns` 對映 `player+0x88F`；現行 remake 尚未拆出
  獨立接觸 bitset，因此可進外交畫面的 AI 視為已接觸，每回合遞增並封頂 250，未滿 10 不派艦。
- 已接純規則：七欄 personality score、可表示的 relation／formal policy／Charismatic／Diplomat
  基礎分、正負 threshold、原版 RNG 消耗順序與 0–4 結果分類。這些規則已有 fail-closed 測試，
  但尚未用不完整 score 取代 shell fallback。
- 已接方向 writer：玩家成功執行 `break_formal` 後，AI→玩家的
  `OriginalHumanBetrayalRaw` 永久設為 true 並通過存檔；Honorable base score 隨即從 +20 改讀
  Dishonored -10。只終止貿易的反例不寫此旗標。
- 已接純規則：`OriginalHumanTargetIncidentScore` 依原版共用表消費 `+0x71F／+0x6CF`，
  並回傳相符的原因碼；非法 signed-byte／reason／personality 失敗即關閉。`sub_4F0DC`
  上游完整門檻及各正常玩家事件 reason 尚未 typed，因此此輪不偽造 writer。
- 已接純規則：存活帝國人口優勢與 40 回合人口成長差；兩者直接消費原版已存在的
  `+0xA6／+0xB9B` typed 對映，不把反編譯器暫名當證據。
- 已接純規則：`sub_500CF` 國力比重用既有 `OriginalNPCPowerRatio`，另補 ratio>=300 的
  score／行動上限 consumer。玩家↔AI 的逐艦方向 `+0x5EC` shell producer 已泛化：雙方
  皆依 owner 艦艇與 observer 科技／引擎／種族防禦計算，任一 raw 缺欄就失敗即關閉。
- 已接純規則：上述 `+0x60E`、government 3、食物赤字、government 1 與 government 0
  分支的 gate／算式。`sub_DCB47` 無蟲洞 producer 已接 typed 人口群、殖民星、聯盟與燃料
  距離；人口群／燃料未知或目標需要尚未閉合的蟲洞 mask 時失敗即關閉。
- 已接垂直鏈：玩家 `break_formal` 除既有永久 `+0x727` 外，現在也依受害 AI personality
  寫 signed `+0x7EE` 與受害 slot `+0x7F6`，通過 JSON 往返；純規則依原版 `3*x/5`
  與 176／177 原因碼消費。經濟協議解約仍有負對照。
- 已接 composer：`OriginalHumanTargetScore` 依原始先後處理關係、incident、三個 -150
  覆寫、government 1、grievance、國力、人口、歷史、條約、personality、target raw
  modifier、Charismatic、Diplomat 與難度，保留 worst term／reason／action limit。
- 已接 shell typed producer：新局可由雙向逐艦國力、人口、歷史與持久 raw 組合上述 score；
  government 0 蟲洞支線、GAM／舊 JSON 缺 incident、歷史不足時失敗即關閉。完整 producer
  成功時已接管正常回合；上述 unknown 狀態才保留 stance fallback。
- 已接 `sub_4F93B` 四種候選 producer：科技比例與候選排序、無納貢時 source 六項維護費、
  真人可要求殖民星排序與高低半部 payload 均可 typed 產生；kind RNG 與科技 payload 的兩個
  off-by-one 已修正。任一必要資料未知時整體失敗即關閉，避免部分 availability 改變 RNG。
- 已接正常回合 orchestration：同一 RNG stream 依序走完整 score、intensity、action 與 outcome；
  outcome 1／3／4 建立可持久化 typed request，outcome 2 交給既有單主力航程 adapter，0 不動作。
  任一 producer unknown 才回到明示 stance fallback。來源人口最強且 intensity 可能大於 3 時，
  新局直接消費議會生命週期保存的 `word_19A0E2` 三態；只有舊 JSON／GAM unknown 且可能改變
  type 4 時，才在消耗本輪 RNG 前失敗即關閉。
- 已移除：producer 的固定 12 回合寬限、1.25 倍軍力門檻與 losing-ground personality
  擬亂數。10 回合 `LastRaidTurn` 只留作 remake 單一主力艦隊停在同星時避免每回合重複
  結算，不能稱作原版 target cooldown。
- 尚未閉合：`sub_544A1` 所需的完整 directional incident writer，以及 outcome 1／3／4
  接受／拒絕 callback。只有這些 typed 輸入 unknown 時，
  願戰來源才保留明示的 `DecideStance` 相容 fallback；不以部分 score 升格整條決策。

## 勘誤：`sub_4F93B`

`sub_4F93B @ 0x4F93B..0x4FD30` 不是軍事目標 availability。它以四個候選旗標、action count、
雙方科技／BC／殖民地資料選擇外交行動，並寫 `word_19B580／word_19B582／byte_19B584／
byte_19B587` payload；回傳 0 才使 `sub_544A1` 把 action count 歸零。後續文件與 typed API
一律稱 `DiplomaticActionAvailable`，不再沿用錯誤軍事名稱。

四種內部 kind 已由 raw branch 閉合：1 是 BC payload、2 是科技候選 payload、3 是殖民地
payload、4 是一至二級直接要求。selector 先把 intensity 夾至 10；intensity>6 關閉 kind 4、
科技／BC 分別受可供候選比例與支付能力上限、殖民地要求 intensity≥6。候選選擇依序消耗
`Random_(3)`、`Random_(3)`、`Random_(2)`；kind 1 金額按 10／100 向下取整並封頂 32000，
kind 2 再消耗 `Random_(3)` 由候選尾端反向取索引。這一核心已實作為
`OriginalHumanDiplomaticActionSelect`；科技候選表與可要求殖民地仍由尚未閉合的上游建立。
