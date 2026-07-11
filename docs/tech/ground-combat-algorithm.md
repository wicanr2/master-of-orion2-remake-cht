# MOO2 地面戰解算演算法(社群逆向 + 已知歧義)

> 目的:記錄 MOO2 地面戰(入侵殖民地)的**解算演算法**,供實作 `ResolveGroundBattle` 用。
> 日期:2026-07-10。方法:reference-before-reverse——手冊(GAME_MANUAL.pdf p.164)只給加成表與定性描述,**無解算式**;openorion2 只有 ground 的 tech-group/trait 引用(渲染殼,無解算)。故解算式取自**社群逆向**。
> ✅ **2026-07-10 定案**:社群公式歧義**不採用**;依使用者 directive 改用**一代(1oom)`game_ground_kill` 的 d100+force 解算**(明確、無歧義、不需 DOSBox 校準)+ 二代手冊加成表。詳見下方「解算式定案」。加成表(手冊驗證)在 `internal/gamedata/ground.go`。

## 已忠實(gamedata,手冊驗證)

`internal/gamedata/ground.go` 的加成表逐條轉寫自手冊(p.15-129):
- hits-to-kill:`GroundMarineHitsToKill`(基礎 + High-G + Powered Armor)、`GroundTankHitsToKill`。
- 戰力加成:`GroundArmorTechBonus`(Tritanium+10…Xentronium+30)、`GroundEquipmentTechBonus`(Powered Armor/Anti-Grav/Personal Shield)、`GroundRaceCombatBonus`(Bulrathi+10/Gnolam-10)、Low-G 懲罰、穴居防守+10、Battleoid+10、步槍(Laser/Fusion Rifle)。

## ★ 解算式定案(2026-07-10,依 directive 用一代公式補歧義)

> 使用者定案:手冊無 MOO2 解算式 → **沿用一代(1oom)公式**,不再等 DOSBox oracle。
> 一代來源:`~/master-of-orion/1oom/src/game/game_ground.c`(`game_ground_kill`,GPL 重製,逐位元組對齊原版)。

**採納的解算(1oom `game_ground_kill`,簡潔且無歧義)**:
```
每回合(game_ground_kill):
  v1 = rnd_1_n(100) + 攻方 force      // d100 + 戰力
  v2 = rnd_1_n(100) + 守方 force
  若 v1 <= v2 → 攻方損失一單位(平手歸守方)
  否則        → 守方損失一單位
反覆(game_turn_ground 的 while)至一方單位歸零;歸零方落敗。
```

**force 計算**:採 **MOO2 加成表**(已在 `internal/gamedata/ground.go`,手冊驗證),非一代的 tier×5——即「一代解算結構 + 二代數值表」:
- force = Σ(裝甲科技加成 `GroundArmorTechBonus` + 裝備 `GroundEquipmentTechBonus` + 種族 `GroundRaceCombatBonus` + 步槍 + Battleoid `GroundBattleoidCombatBonus`) − Low-G 懲罰;
- 守方另加 Subterranean 防禦 `GroundSubterraneanBonus(true)`(+10)。(一代守方 +5 之對應;MOO2 用穴居 +10 有手冊據,故採 MOO2。)

**hits-to-kill 疊加(MOO2 手冊 p.129,一代無)**:一代「損失一單位」= 直接 −1 pop;MOO2 單位需被命中 `GroundMarineHitsToKill`/`GroundTankHitsToKill`/`GroundBattleoidHitsToKill` 次才移除。故本專案:每回合敗方「最前單位」受 1 hit,累積達其 hits-to-kill 才移除(pop −1)。此為二代手冊表 + 一代解算的忠實合成。

**∴ 歧義全消解**:先前社群式的 x₁/x₂/x₃/x₄ 與傷亡運算子優先序歧義**不採用**;改用一代明確的 d100+force 對決。此定案不需原版實測即可實作(directive)。實作 `ResolveGroundBattle` 依此。

---

## (存查)社群逆向式(不採用,保留追溯)

**單位基礎戰力(combat rating)**:
- 步兵 = 0.5 + 步槍 + 艦裝甲科技 + 特殊裝備 + 種族
- 戰車 = 1 + 步槍 + 艦裝甲科技

**四常數**(攻擊方戰車 +100、防守方戰車 +50 是關鍵不對稱——armor 適合進攻):
- x₁ = (50 + 攻方步兵戰力) × (1 + 艦隊領袖加成) × 100
- x₂ = (100 + 攻方戰車戰力) × 100
- x₃ = (50 + 守方步兵戰力) × (1 + 殖民地領袖加成/100) × 100 + 5
- x₄ = (50 + 守方戰車戰力) × 100 + 5

**每回合解算**:
1. 每個攻方單位擲 rᵢ ∈ [1, 1.5]:步兵 aᵢ = x₁×rᵢ、戰車 bᵢ = x₂×rᵢ。守方同理用 x₃/x₄。
2. Total Attacker Value = Σaᵢ+Σbᵢ;Total Defender Value = 守方同。
3. **傷亡**(⚠ 歧義步驟):每單位再擲 vᵢ ∈ [0,1];攻方某單位若 `aᵢ < vᵢ × (Total Defender Value / 守方擲和)` 則陣亡。守方對稱。
4. 反覆至一方單位歸零。單位需被命中 hits-to-kill 次才真正移除(見 gamedata)。

**經驗法則**(社群):戰鬥評等每差 10-20 點,弱方需雙倍兵力才勝。

## 已知歧義(實作前須對原版校準,勿當精確)

1. **運算子優先序**:如 x₁「50 + 步兵戰力 × (1+領袖) × 100」——是 `(50+戰力)×(1+領袖)×100` 還是 `50 + 戰力×(1+領袖)×100`?社群記述未明。
2. **戰力刻度**:「基礎步兵 0.5 + 步槍」——步槍/裝甲加成在手冊是 +10/+20(combat rating),與 0.5 混算刻度不一致;可能社群把 rating 正規化過(÷某值),需釐清。
3. **傷亡式**中 `vᵢ × 對方總值/對方擲和` 的確切定義(vᵢ 值域、擲和是 r 和還是 v 和)社群描述含糊。
4. hits-to-kill 如何與「單位陣亡判定」結合(一次判定=一次 hit?)未明。

## 實作計畫(定案後,不需 oracle)

1. ✅ **已完成**(見 `internal/gamedata/ground_battle.go`):`ResolveGroundBattle(atk, def GroundForce, rng)`,一代解算 + 二代 hits-to-kill + 二代 force 加成表,確定性測試綠(`ground_battle_test.go`)。
2. ✅ **已完成**(2026-07-11,`internal/shell/ground_invasion.go` + `ground_invasion_test.go`):陸戰隊生成 → 運送 → 觸發入侵 → 勝則轉移殖民地的「模型 + 流程」shell 層接線,細節見下方「2026-07-11 shell 層接線」一節。**尚未做**:UI 繪製/操作介面(不碰 interactive.go)、同化/滅絕選擇(手冊 p.164,本輪只做「整批過戶」的簡化版)。
3. ✅ **已完成**:確定性單測驗「force 差/兵力比 → 勝率」符合社群經驗法則,見 `ground_battle_test.go` 與本輪新增的 `ground_invasion_test.go`(接了 shell 層模型後的端到端勝率測試)。
4. ✅ **已完成**(2026-07-11,`internal/shell/orbital_bombardment.go` + `ground_invasion.go` 補完):裝甲營房戰車生成/載運/納入攻方 `GroundForce`(Battleoids 升級切換 hits-to-kill/force 加成)+ 軌道轟炸(10 輪齊射模擬 → hits → 扣人口),細節見下方「2026-07-11 裝甲營房戰車 + 軌道轟炸接線」一節。**同日稍後補上 UI 操作介面**(`cmd/moo2/interactive.go` galaxy() 新增「軌道轟炸」按鈕,與「發動地面入侵」雙鈕共存)。**尚未做**:守方戰車(AI 無 `ColonyBuildings` 追蹤,無資料可推導)、轟炸扣建築/儲存生產/駐軍(AI 無對應持久資料)、光束/魚雷減半與電腦命中加成套用到轟炸(戰術戰鬥層本身尚無這兩項的獨立函式)。

## 2026-07-11 shell 層接線(InvadeColony 流程)

把上方已定案的解算式接進活的對局狀態(`internal/shell/ground_invasion.go`),流程:陸戰隊生成
(Marine Barracks)→ 載運(LoadMarines)→ 入侵解算(InvadeColony)→ 勝則佔領。

**忠實部分**(直接用手冊/已驗證加成表,無臆測):
- 陸戰隊生成公式 `GroundMarineBarracksUnits`(初始 4 + 每 5 回合 +1,上限 `GroundMarineBarracksCap`)逐回合接進 `EndTurn`(`advanceMarines`)。
- Force 計算重用既有「艦艇元件解鎖」判定(`ComponentUnlocked`/`ArmorOptions`):玩家裝甲科技加成取「已解鎖裝甲元件中最高階者」對應的 `GroundArmorTechBonus`,而非另建一套獨立判定,避免地面戰科技狀態與造艦科技狀態不同步。
- Powered Armor / Anti-Grav Harness / Personal Shield 三項裝備科技(本 remake 艦艇元件模型未收錄)直接查 `CompletedTopics`/`ChosenTech`,加總(非互斥升級,三者可並存,不同於裝甲槽)。
- 種族地面戰加成:僅 Bulrathi(+10)/Gnolam(−10 + Low-G 10% 懲罰)手冊有明確數字,套用 `groundRaceFor`;其餘種族與尚未建模的「特殊能力(Subterranean/High-G)」誠實留白,不臆測。

**簡化(標記待精修,依 83-completeness-over-roi 誠實揭露,不是藉口不做)**:
1. **運輸艦運力**:本 remake 尚無獨立「運輸艦」船體類別,`MarineTransportCapacity()` 用「艦隊現有艦數 × 手冊每艘 4 個單位」近似,不分船體類型。待補真正運輸艦船體後應改為只計數該類型。
2. **AI 守方兵力**:AIOpponent 沒有追蹤各殖民地 Marine Barracks 是否建成/已運作幾回合(無 AI 版 ColonyBuildings),用「已運作 `s.Turn` 回合」近似 `GroundMarineBarracksUnits` 的 turnsSinceBuilt 參數(AI 母星開局即有 Marine Barracks,近似合理但非精確追蹤)。
3. **AI 種族/特殊能力**:AIOpponent 無 RaceIndex,AI 側 force 只計裝甲/裝備科技加成,不套種族/Low-G/Subterranean。
4. **入侵後保留人口**:手冊 p.162-164 只有敘述性描述,無精確的「入侵後保留多少平民人口」公式;以「守方地面戰存活戰鬥單位數」近似戰後殖民地人口(至少 1),不做同化/滅絕的玩家抉擇(手冊有此選項,本輪未做)。
5. **可入侵範圍**:AI 每 5 回合 `aiExpand` 佔領的無主星只標記 `Owner=2`,不建立殖民地經濟模型(見 `AIOpponent.ColonyStars` 註解)。故本輪只有 AI 開局母星(唯一有真實 `ColonyState` 的星)可被入侵;其餘擴張版圖入侵時會回報「無可入侵的殖民地模型」。

**流程選擇**:入侵由玩家主動呼叫 `InvadeColony`(非艦隊一抵達就自動觸發),與既有架構一致——`SendFleet`/`BuildShip`/`ShiftColonyJob` 等所有玩家決策都是顯式呼叫,不是 `EndTurn` 自動觸發;也讓玩家能先觀察/多載陸戰隊再決定開打,貼近原版「艦隊指令選單」的操作語意。

## 2026-07-11 裝甲營房戰車 + 軌道轟炸接線

上一節「shell 層接線」只做了陸戰隊(Marine Barracks);`internal/gamedata/ground.go` 當時還有
幾個零呼叫端的死碼:`GroundArmorBarracksUnits`/`GroundArmorBarracksCap`(裝甲營房生成戰車)、
`GroundTankHitsToKill`(戰車耐久)、`GroundBombHitsFromDamage`/`GroundPlanetTotalHits`(軌道
轟炸)。本輪把這些接進去,只用已移植公式,不自編數字。

### 裝甲營房戰車生成

比照陸戰隊完全對稱的做法(`internal/shell/ground_invasion.go`):
- `GameSession` 新增 `FleetTanks`/`PlayerColonyTanks`/`ArmorBarracksAge` 三個欄位,對應
  `FleetMarines`/`PlayerColonyMarines`/`MarineBarracksAge`。
- `advanceArmor()`(接進 `EndTurn`,緊接在 `advanceMarines()` 之後):依
  `gamedata.GroundArmorBarracksUnits(age, pop, popMax, warlord)` 逐回合補充戰車營駐軍池,
  上限 `GroundArmorBarracksCap`,只認 `ColonyBuildings[i]["裝甲營房"]`(新增
  `armorBarracksBuildingName` 常數,取代先前只在 `colonyMoralePercent` 出現過的寫死字面字串)。
- `LoadTanks(colonyIdx)`:載運戰車上艦隊。⚠ 簡化:remake 沒有獨立的「戰車運輸艙位」資料
  (手冊只明講 Transport Ship/Troop Pods 是針對 Marine),與 `LoadMarines` **共用同一個**
  `MarineTransportCapacity()` 運力池(`room = 運力 - FleetMarines - FleetTanks`)。
- `InvadeColony` 攻方 `GroundForce` 混編:`FleetMarines` 個陸戰隊單位 + `FleetTanks` 個戰車
  單位。**排序決策(已選定,列出讓 L.CY 定案)**:合併陣列「陸戰隊在前、戰車在後」——這不是
  戰術隊形選擇(手冊未提供地面戰隊形資訊),而是**技術上唯一能把 `ResolveGroundBattle` 回傳的
  單一 `AttackerSurvived` 總數,精確拆回「陸戰隊存活數/戰車存活數」的排法**:該解算式的規則是
  「最前面存活單位先受創」,即單位嚴格按索引順序陣亡,故存活者必是原始順序的「後段」;把戰車
  放在後段,戰後只需 `tanksSurvived = min(總存活, 戰車原始數量)` 即可還原分兵種存活數,不需
  更動 `gamedata.GroundUnit`/`GroundForce` 結構(該結構本身無兵種標記欄位,加欄位是更大改動,
  超出本輪死碼串接範圍)。若未來要精確模擬「戰車在前掩護陸戰隊」這種戰術隊形,需要先幫
  `GroundUnit` 加兵種欄位。
- Battleoids 升級(手冊 p.81,`TOPIC_ASTRO_CONSTRUCTION` 三選一 `TECH_BATTLEOIDS`,真實可研究
  科技,非里程碑 proxy):已研究則戰車固定 3 hits(`GroundBattleoidHitsToKill`,不再套用
  Heavy-G 修飾)+ 額外 `GroundBattleoidCombatBonus`(+10)force 加成,**只在戰車數 > 0 時套用**
  (0 輛戰車不該白拿這個「戰車升級後」的加成)。
- **守方戰車:TODO 未接**,且理由與陸戰隊側不同——`InvadeColony` 對守方陸戰隊的近似(用
  `s.Turn` 當 `turnsSinceBuilt`)至少有「AI 母星開局必有 Marine Barracks」(`homeworldBuildings`)
  這個事實撐腰;但 `homeworldBuildings()` 本身就沒有裝甲營房,且 `AIOpponent` 完全沒有
  `ColonyBuildings` 追蹤機制可判斷「AI 是否已建成裝甲營房」,沒有資料可誠實推導守方戰車數,
  不臆測補上。

單測(`ground_invasion_test.go`):裝甲營房隨回合生成戰車受上限節制、無裝甲營房不生成、
`LoadTanks` 與 `LoadMarines` 共用運力池、Battleoid hits-to-kill/force 加成切換、
`InvadeColony` 只有戰車也能發動入侵、混編後存活數拆解一致性(30 場迴圈驗證恆等式)、
**加戰車確實提升攻方勝率**(3 陸戰隊+0 戰車 vs 3 陸戰隊+12 戰車,100 場對照組,證明戰車真的
被納入解算而非擺著沒用的死碼)。

### 軌道轟炸(Orbital Bombardment)

手冊 MANUAL_150.html p.129「Notes on Orbital Assault > Orbital Bombardment」與地面入侵是**兩個
獨立動作**:轟炸只削弱/殺人口(不佔領),佔領仍要靠 `InvadeColony` 的陸戰隊/戰車入侵。新增
`internal/shell/orbital_bombardment.go`:

- `fleetBombardDamage(rng)`:依手冊「All remaining ships fire all weapons 10 times... total
  damage is calculated from it」模擬 10 輪齊射,逐發解算重用既有戰術戰鬥公式(`ResolveShot`/
  `ResolveMissileShot`,依 `weaponKindByName` 分流,同 `battleVolley` 的既有分流邏輯),目標從
  「敵艦」換成「殖民地」,不模擬殖民地反擊(手冊本段只描述攻方輸出)。
- `BombardColony(starIdx)`:`fleetBombardDamage` 總傷害 → `gamedata.GroundBombHitsFromDamage`
  換算 hits → 依 Planet Hits 表「每整數人口 1 hit」直接扣減 `colony.Population`(夾在 0 以上)。
  另用 `gamedata.GroundPlanetTotalHits` 算一個純供顯示參考的 `PlanetHitsRequired`(對應手冊 UI
  「Estimated Bomb Hits」旁邊同時顯示的「Planet Hits」欄,讓玩家判斷這波轟炸夠不夠),**不**
  用它去扣建築/駐軍(見下方範圍限制)。

**範圍限制(誠實標註,非杜撰真值,是既有 remake 資料模型限制,非本輪引入)**:
1. **只扣人口,不扣建築/儲存生產/駐軍**——AI(`AIOpponent`)完全沒有 `ColonyBuildings`/儲存
   生產/駐軍的持久資料可扣,扣了會是憑空生資料,故不做。這與「只做攻方戰車」是同一種誠實邊界:
   有資料才接,沒資料不臆測。
2. **手冊「Damage of beams and torpedoes is halved just like in tactical combat」與「A better
   computer helps for beams here too」未套用**——不是轟炸專屬的遺漏,是戰術戰鬥層本身現在就
   還沒有獨立的「光束/魚雷減半」或「電腦命中加成接線」函式(`ground.go` 檔尾原有 TODO 已載明),
   本模擬只能沿用一般 `ResolveShot`,兩項都待戰術戰鬥層先補上才能真正對齊手冊轟炸公式。
3. **行星護盾未建模**——`damage.go` 的 `DamageAfterShield` 明講「本函式只處理艦對艦,行星護盾
   情境不適用」,remake 目前也沒有任何「行星防禦/護盾」資料欄位,故轟炸模擬視同殖民地護盾/
   裝甲恆為 0(無防禦)。
4. **人口歸零後的後續未定義**——手冊沒講「殖民地被轟炸到 0 人口」要不要摧毀殖民地/移除星系
   Owner,本函式讓 `Population` 停在 0、殖民地本身仍存在於 `aiPlayer.Colonies`,不臆測補上
   摧毀邏輯(TODO,留給未來確認手冊或 openorion2 行為後再接)。
5. **UI 已接**(2026-07-11 同日稍後補上)——`cmd/moo2/interactive.go` galaxy() 星系主畫面新增
   `"bombard"` 熱區/按鈕(敵殖民地星恆可用,不需陸戰隊),與既有 `"invade"` 熱區共存,分居
   y=402/424 兩列(轟炸恆在上排,入侵需 `FleetMarines>0` 才出現在下排)。按鈕點擊直接呼叫
   `BombardColony`,依 `GroundBombardResult.Ok`/`Reason`/`PopulationLost` 顯示結果訊息。

單測(`orbital_bombardment_test.go`):前置條件(無效星索引/艦隊未抵達/非敵方/無艦艇/無殖民地
模型)、rng 種子化可重現、**用保證命中+固定滿傷的艦隊(atk=101≥99,`CombatClassicToHit`/
`DamageForHit` 手冊「[2] BA+CO-AF-BD>=99 恆命中恆滿傷」分支)手算驗證整條換算鏈**(10 輪×1 艦
×101 傷害=1010 總傷害→hits=10→母星預設人口 8 全數扣光→`RemainingHits`=2)、人口不會扣成負數、
轟炸不佔領星/不增減 AI 殖民地筆數。

## 2026-07-11 版本差異補實作

補實作 `docs/tech/version-1.3-1.5-diff.md` §1 全量表 #5/#6/#7/#8/#9/#11(地面戰/軌道轟炸的
patch 1.3 vs 1.5 差異項)。新程式碼:`internal/gamedata/ground_version_diff.go`(公式)+
`internal/gamedata/ground_version_diff_test.go`(測試);接線:`internal/shell/ground_invasion.go`
(`commandoLeaderTier` + `InvadeColony` 攻方 force 疊加)、`internal/shell/orbital_bombardment.go`
(`BombardColony` 改用 `gamedata.GroundBombardPopulationLoss`);`internal/gamedata/ruleprofile.go`
新增 2 欄位(`DefenderCommandoBonus`、`BombardmentBuildingBonusHits`)。逐項誠實記錄:

| # | 項目 | 版本差異? | 實作程度 | 說明 |
|---|---|---|---|---|
| #6 | 指揮官(Commando)攻方倍率 | 否(兩版同) | **完整接線**(近似公式) | `GroundCommandoAttackerForceBonus(tier)`:tier1=5、tier≥2=7(2.5×3=7.5 捨去)。已接進 `InvadeColony` 的攻方 force(`commandoLeaderTier(s.Leaders)` 掃描帝國 Leaders 找 `Skill=="指揮官"` 的最高 Tier)。**近似**:①「regular commando bonus」基準值(2/3)手冊只給相對倍率,沒給獨立驗證的絕對數字,本檔直接當成最終加成點數;②remake 無「領袖指派到某次入侵」模型,用「帝國是否擁有 Commando 技能領袖」當代理條件,不論其 `Ship`/實際位置。單測:`TestGroundCommandoAttackerForceBonus`、`TestInvadeColony_CommandoLeaderImprovesWinRate`(實測無指揮官勝率 0.63→有指揮官 0.75,150 場)。 |
| #5 | 防禦方 Commando 2.5x | **是** | **完整接線(2026-07-11 補完,近似公式)** | `RuleProfile.DefenderCommandoBonus`(1.3=1.0、1.5=2.5)套進 `GroundCommandoDefenderForceBonus(tier, bonus)`,已接進 `InvadeColony` 守方 force(`commandoLeaderTier(aiPlayer.Leaders)`)。前置的 AI 領袖資料模型已補上:`AIOpponent.Leaders`(比照玩家 `GameSession.Leaders`),`buildDemoAIOpponents` 依種族性格開局固定指派(布拉西人 Tier2/姆瑞森人 Tier1/席隆人無指揮官)。**近似(誠實標註)**:原版領袖是從英雄池隨機雇用、可陣亡替換的動態資源;remake 用「開局依種族性格固定指派、不隨遊戲成長」當代理,非手冊逐字機制,與 #6 攻方 `commandoLeaderTier` 的「帝國全域清單當代理」同款近似紀律。`persist.go` `aiSnapshot.Leaders` 已比照 `ColonyBuildings` 雙向序列化,舊存檔解碼為 nil 時 `commandoLeaderTier(nil)=0` 安全降級。單測:`TestBuildDemoAIOpponents_CommandoLeadersByRace`、`TestInvadeColony_DefenderCommandoLowersAttackerWinRate`(實測有守將勝率 0.47 < 無守將 0.59,150 場)、`TestInvadeColony_DefenderCommandoVersionDifference`(1.3 攻方勝率 0.53 > 1.5 攻方勝率 0.47,150 場;公式數值 tier2:1.3→3、1.5→7)、`TestInvadeColony_NoDefenderCommandoForPsilon`(無守將回歸)、`TestInvadeColony_NilAIPlayerLeadersSafeDegrade`(舊存檔安全降級)。 |
| #7 | 轟炸建築 +1 hit(1.3 bug) | **是** | **僅 RuleProfile 欄位 + TODO 掛鉤** | `RuleProfile.BombardmentBuildingBonusHits`(1.3=1、1.5=0)。本 remake 軌道轟炸只扣人口不扣建築(AI 無 `ColonyBuildings` 持久資料可扣),無「建築 hit」概念可套用這個加成,故只加欄位 + `BombardColony` 內註解掛鉤點,無任何函式讀取它。 |
| #8 | civilian_armor 100hp | 否(兩版同) | **常數鎖定,無消費端** | `gamedata.GroundCivilianArmorHP = 100`(PARAMETERS.CFG:1778-1786 逐字數字)。與 #7 同屬未建的「建築損傷模型」,先鎖定數字供未來引用。單測:`TestGroundCivilianArmorHP_LockedValue`(純數值鎖定)。 |
| #9 | 地面防禦建築結構倍率 | 否(兩版同) | **常數鎖定,無消費端** | `gamedata.GroundDefenseArmorMultiplier = 100`(PARAMETERS.CFG:1772-1775)。remake 沒有「地面防禦建築」這個資料實體(`ColonyBuildings` 只是 `map[string]bool` 有/無旗標,無 HP/結構值欄位;AI 側連追蹤都沒有),無法真正套用,誠實標「掛鉤備妥、待防禦建築系統」。單測:`TestGroundDefenseArmorMultiplier_LockedValue`。 |
| #11 | 轟炸行星尺寸幾何 3-4-6-7-8 | 否(1.5 系列中途改過又於 1.50.11 修回 classic,對 1.3 vs 最終 1.5 不構成差異) | **完整接線**(近似公式) | `GroundPlanetSizeBombardCoefficient(size)` 對照手冊/CHANGELOG 數字(Tiny=3/Small=4/Medium=6/Large=7/Huge=8);`GroundBombardPopulationLoss(hits, size) = hits×6/coef`(以 Medium 為基準 1:1,大行星較耐轟)。已接進 `BombardColony` 取代原本的 `popLoss := hits` 直接相等。**近似**:手冊只給尺寸係數本身,沒給「係數如何代入人口損傷公式」的精確算式(原版可能牽涉地圖網格/區域數,remake 沒有這層模型),本檔採「與係數成反比、Medium 基準」的最簡近似。**behavior-preserving 巧合**:母星預設 `LARGE_PLANET`(係數 7),既有測試 `TestBombardColony_ReducesPopulationDeterministically` 的 hits=10 算出 popLoss=8,與換公式前的舊行為(`popLoss==hits` 直接相等,10 被人口 8 封頂為 8)結果剛好一致,測試未變紅。單測:`TestBombardmentPlanetSizeScaling`(Tiny 24 > Small 18 > Medium 12 > Large 10 > Huge 9,hits=12)。 |

**RuleProfile 新增欄位**(`internal/gamedata/ruleprofile.go`):
- `DefenderCommandoBonus float64`(Profile13=1.0 / Profile15=2.5)
- `BombardmentBuildingBonusHits int`(Profile13=1 / Profile15=0)

**驗證**:docker(`moo2-ebiten:latest`)內 `go build -buildvcs=false ./...`/`go vet ./...`/
`go test ./internal/...` 全綠(僅既有的 `internal/uifont` X11/GLFW headless 限制,與本輪無關,
既有已知環境限制);`go build -buildvcs=false -o /tmp/moo2-groundcombat-check ./cmd/moo2` 成功
(未寫回 repo 根目錄 `moo2` 二進位)。既有 `ground_invasion_test.go`/`orbital_bombardment_test.go`
全部維持綠燈,無需更新任何既有斷言數字。

## 2026-07-11 追加:防禦方 Commando 消費端接線收尾(#5 收尾)

上一輪(本節上方,同日較早批次)已做好 `gamedata` 公式與 `RuleProfile.DefenderCommandoBonus`,
但誠實留白——`AIOpponent` 沒有 `Leaders` 欄位,shell 層沒有資料可推導「AI 是否擁有 Commando
守將」。本輪補上這個最後一哩:

- `AIOpponent` 新增 `Leaders []Leader` 欄位(`internal/shell/session.go`),型別比照
  `GameSession.Leaders`。`buildDemoAIOpponents` 依種族性格開局固定指派(布拉西人「體格強悍,
  地面戰加成」→ Tier2 進階指揮官;姆瑞森人「好戰善攻」→ Tier1 一般指揮官;席隆人「重研究」→
  無指揮官),對應 `demoAIOpponentSetup` 新增的 `commandoTier` 欄位。
- `ground_invasion.go` `InvadeColony` 守方 force 加上一行:
  `defForce += gamedata.GroundCommandoDefenderForceBonus(commandoLeaderTier(aiPlayer.Leaders), s.RuleProfile.DefenderCommandoBonus)`,
  取代原本的 TODO 留白註解。
- `persist.go` `aiSnapshot` 比照 `ColonyBuildings` 加 `Leaders` 欄位雙向序列化,舊存檔解碼為 nil
  時 `commandoLeaderTier(nil)=0`,安全降級為無加成。

**誠實標註不變**:原版領袖是從英雄池隨機雇用、可陣亡替換的動態資源;remake 沒有 AI 英雄雇用/
成長系統,這裡的「依種族性格開局固定指派」是可觀察、確定性的近似(入侵布拉西最難、姆瑞森次之、
席隆無守方 commando 加成),不是手冊逐字的隨機雇用機制。AI 領袖清單建立後不隨遊戲成長變動。

**版本效果落地**:tier2 基準 3,1.3(`DefenderCommandoBonus=1.0`)→ `int(3*1.0)=3`;1.5
(`=2.5`)→ `int(3*2.5)=7`。150 場模擬:入侵有 Tier2 守將的布拉西人母星,1.3 攻方勝率 0.53、
1.5 攻方勝率 0.47——1.5 確實比 1.3 更難入侵,#5 的版本差異在遊戲迴圈裡真的生效,不再是只存在於
`gamedata` 單元測試裡的公式。

**驗證**(同一 docker image):`go build ./...`/`go vet ./internal/shell ./internal/gamedata`/
`go test ./internal/shell/ ./internal/gamedata/` 全綠;`moo2sim -turns 20` 經濟軌跡未變(本輪
只加 AI Leaders 資料 + `InvadeColony` 守方項,不進 `EndTurn`)。

## 來源

- Steam《Master of Orion》社群〈Ground Combat Formula〉討論串(社群逆向)。<https://steamcommunity.com/app/298050/discussions/0/135509124605301259/>
- StrategyWiki《MOO2/Battle tactics》。
- 手冊 GAME_MANUAL.pdf p.164(定性 + 加成)、gamedata/ground.go(加成表,已驗證)。
- `moo2_patch1.5/MOO2-1.50.26.zip` 內 `patch/150/docs/PARAMETERS.CFG:1772-1786,2740-2758`(地面
  防禦建築結構倍率/civilian_armor/Commando 倍率門檻,`(default, classic)` 逐條標註)。
- `moo2_patch1.5/MANUAL_150.html`(python 去標籤全文關鍵字搜尋 "Commando":Commando Leader 段落
  「A defending commando gives 2.5x the regular commando bonus...」)。
- `docs/tech/version-1.3-1.5-diff.md` §1 全量表 #5-#11(本節補實作的來源清單與逐條核對紀錄)。
