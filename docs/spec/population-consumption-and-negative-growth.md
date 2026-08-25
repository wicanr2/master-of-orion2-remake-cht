# 逐族人口消耗與負成長規格

## 證據範圍

本規格對應 `sub_DEB4B @ 0xDEB4B`、`sub_DF546 @ 0xDF546`、
`sub_E1839 @ 0xE1839` 與 `sub_E2DCA @ 0xE2DCA`。原始位址、指令與證據等級見
`docs/re/mixed-race-colonist-production-audit-20260825.md`。

## 群組 profile

每個 `PopulationGroup` 除既有產出、重力與成長 trait 外，必須保存 `Cybernetic` 與
`Lithovore`。玩家、AI、熱座、`.GAM`、新殖民地及 JSON 往返都必須注入這兩個欄位。
Android／Natives 使用固定特殊 profile，不把 owner 的兩個布林值套給它們。

## 半單位消耗

- 一般人口每人吃 2 半食物；Cybernetic 吃 1；Lithovore 吃 0。若資料同時帶兩項，依原版
  分支順序 Cybernetic 優先，消耗 1。
- Android 不吃食物、每人消耗 2 半工業；Natives 每人吃 2 半食物、不消耗工業。
- 一般 Cybernetic 每人消耗 1 半工業；其餘一般人口不消耗工業。
- owner 人口先歸 owner 類；非 owner prisoner 歸 prisoner 類，其餘歸外族類。逐 slot 原始
  需求仍另行保存，供外族短缺比例分配。
- 食物供應依 owner、外族、prisoner、Natives 順序扣除；工業供應依 owner、Android、外族、
  prisoner 順序扣除。Android 只有在尚有完整 2 半工業時才供應一人。

## signed 逐槽成長率

1. owner 未滿足的食物與工業需求每半單位 `-25`；Natives 未滿足食物同樣每半單位 `-25`。
2. Android 未滿足工業需求每半單位 `-500`。
3. 外族／prisoner 的剩餘短缺合計乘 `-25`，再以各非 owner 一般 slot 的原始食物＋工業需求
   除以該類全部原始需求分配；採 Go 有號整數除法朝零截斷，對應 x86 `idiv`。
4. 負項先建立，再加入既有逐 slot 平方根正成長、種族成長、住房與固定成長；不得因殖民地
   `Starving` 就把整槽直接歸零。
5. Android／Natives 只保留短缺負項，不加入自然平方根正成長；原版正成長 pass 只走一般
   player slots。

## 低於零時刪人口

1. 每回合先決定唯一保護槽：有 owner 人口則保護 owner；否則取一般 player slot 人口最多者，
   同數取較高 slot；再回退 Android、Natives。保護槽最低留一人，其餘槽可降至零。
2. 各槽 `GrowthPoints += signed rate`。低於 0 且人口高於保護量時，刪一人並加回 1000；可在
   同回合重複。已到最低量時將負池夾為 0，避免日後人口加入立即兌現舊死亡債。
3. 候選只取該槽；若有工人／科學家便不選農夫。候選集合內用原版 reservoir sampling 形式
   均勻抽選，並把 prisoner 身分一起納入候選。
4. 刪除後同步 `Population`、三職務總數、群組職務、群組 prisoner、`Unassimilated*` 與
   `UnassimilatedPop`。空群組可保留 profile 與成長池，不因人口為零刪除。
5. 玩家與 AI 使用同一規則；亂數流必須可存檔。聚合群組沒有原版 packed record 排列，故只
   宣稱候選集合與機率分布對齊，不宣稱逐位元 PRNG 或同一 record 對齊。
6. 回合順序固定為：先把所有 signed rates 加入各槽 → 依 slot 0..9 完成全部負池刪除 → owner
   第一、其餘一般 player slots 以原版 Fisher–Yates 形式洗牌後處理正池新增。Android／Natives
   不進正池 pass。

## 驗收

- 混合 Cybernetic／Lithovore／Android／Natives 的半食物、半工業消耗符合逐群 profile。
- 部分供應時短缺依原版優先序落到正確類別，signed rate 可與正成長抵銷。
- 負池 `-1` 會刪一人並變 999；保護槽最後一人不死且池歸零。
- 有非農夫時永不先刪農夫；固定 seed 的 reservoir sampling 與存讀檔結果可重播。
- prisoner 死亡同步所有未同化計數；玩家與 AI 各有一條垂直測試。
