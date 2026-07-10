# MOO2 地面戰解算演算法(社群逆向 + 已知歧義)

> 目的:記錄 MOO2 地面戰(入侵殖民地)的**解算演算法**,供實作 `ResolveGroundBattle` 用。
> 日期:2026-07-10。方法:reference-before-reverse——手冊(GAME_MANUAL.pdf p.164)只給加成表與定性描述,**無解算式**;openorion2 只有 ground 的 tech-group/trait 引用(渲染殼,無解算)。故解算式取自**社群逆向**。
> ⚠ **鐵律**:社群公式的傷亡判定步驟記述較鬆散(運算子優先序/基礎值刻度有歧義),**尚未當精確真值**;實作前須對原版實測校準這些歧義點。已忠實的部分(加成表)在 `internal/gamedata/ground.go`。

## 已忠實(gamedata,手冊驗證)

`internal/gamedata/ground.go` 的加成表逐條轉寫自手冊(p.15-129):
- hits-to-kill:`GroundMarineHitsToKill`(基礎 + High-G + Powered Armor)、`GroundTankHitsToKill`。
- 戰力加成:`GroundArmorTechBonus`(Tritanium+10…Xentronium+30)、`GroundEquipmentTechBonus`(Powered Armor/Anti-Grav/Personal Shield)、`GroundRaceCombatBonus`(Bulrathi+10/Gnolam-10)、Low-G 懲罰、穴居防守+10、Battleoid+10、步槍(Laser/Fusion Rifle)。

## 解算式(社群逆向,Steam 討論串)

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

## 實作計畫(下輪,對原版校準後)

1. 建 `ResolveGroundBattle(atk, def GroundForce, rng)`:先用上式**結構**(每單位 1~1.5、總值對決、armor 攻+100/守+50、反覆至一方歸零)實作,歧義點取一種合理解讀並**標為待校準**。
2. 建入侵流程:運輸艦載陸戰隊 → 抵敵殖民地 → 觸發地面戰 → 勝則轉移殖民地(+同化/滅絕選擇,手冊 p.164)。
3. **校準**:用原版(DOSBox)跑數場已知兵力/科技的入侵,記錄勝負與傷亡,回頭調歧義點,直到趨勢吻合(這步需 oracle:原版實測)。

## 來源

- Steam《Master of Orion》社群〈Ground Combat Formula〉討論串(社群逆向)。<https://steamcommunity.com/app/298050/discussions/0/135509124605301259/>
- StrategyWiki《MOO2/Battle tactics》。
- 手冊 GAME_MANUAL.pdf p.164(定性 + 加成)、gamedata/ground.go(加成表,已驗證)。
