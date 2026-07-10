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

1. 建 `ResolveGroundBattle(atk, def GroundForce, rng)`:用**一代解算**(每回合雙方 d100+force,低者敗損 1 hit;平手歸守方)+ **二代 hits-to-kill**(單位累積達 hits-to-kill 才 −1 pop)+ **二代 force 加成表**(gamedata/ground.go)。反覆至一方歸零。可寫確定性測試(固定 seed 驗勝負趨勢:force 高方勝率高、雙倍兵力優勢)。
2. 建入侵流程:運輸艦載陸戰隊 → 抵敵殖民地 → 觸發地面戰 → 勝則轉移殖民地(+同化/滅絕選擇,手冊 p.164)。
3. 驗證:確定性單測(seed 化)驗「force 差 → 勝率」「兵力比 → 勝率」符合社群經驗法則(每差 10-20 點需雙倍兵力);**不需原版實測**(一代公式即定案)。

## 來源

- Steam《Master of Orion》社群〈Ground Combat Formula〉討論串(社群逆向)。<https://steamcommunity.com/app/298050/discussions/0/135509124605301259/>
- StrategyWiki《MOO2/Battle tactics》。
- 手冊 GAME_MANUAL.pdf p.164(定性 + 加成)、gamedata/ground.go(加成表,已驗證)。
