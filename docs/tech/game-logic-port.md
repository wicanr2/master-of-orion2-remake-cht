# 遊戲邏輯層移植進度

## 背景與方法

SA1 盤點(`rules-implementation-audit.md`)確認:**openorion2 是渲染殼,不含回合引擎**,遊戲邏輯要在 go 端從零建。
方法依記憶 [先參考再逆向]:**優先移植已驗證的現成公式**(openorion2 中 `⬜ 僅資料結構` 者仍含正確的唯讀公式表),
而非自行推導;逆向/實測為最後手段。每條公式落 `internal/gamedata/`,配手算對照的單元測試。

## 已移植(`internal/gamedata/`)

| 模組 | 內容 | 來源(openorion2) | 測試 |
|---|---|---|---|
| `formulas.go` | 行星基礎生產、電腦/引擎 HP、戰速、光束攻防、**艦員等級攻防加成**、軍官雇用費 | 各處 + `gamestate.cpp:162-167` | `formulas_test.go` |
| `officer.go` | 軍官經驗等級 `LeaderExpLevel`、技能加成 `LeaderSkillBonus` | `gamestate.cpp:607-701`(SA1 標記唯讀正確) | `officer_test.go`(exp 邊界 + 13 組技能手算) |
| `techtree.go` | **完整科技樹**:`researchChoices[83]`(成本/可選科技)、`techtree[8]`(領域→主題) | `tech.cpp:69-305`,重用 `enums.go` 的 212 Technology/83 ResearchTopic | `techtree_verify_test.go`(全 83 列 cost+choices 數對 C 基準,Opus 獨立驗證全等) |
| `enums.go` | 行星/星體/研究/殖民等列舉(212 Technology、83 ResearchTopic…) | `gamestate.h`(生成檔) | `enums_test.go` |

### officer.go 公式重點
- **經驗等級**:門檻 `{60,150,300,500}` → 級 0..4。
- **技能加成**:`base = baseSkillValues[type][code]`;除 Megawealth(不隨等級倍增)外乘 `(expLevel+1)`;
  進階技能(tier 2)再 +50%;Navigator 用專屬值表。技能 id 位元編碼:`type=(id&0x30)>>4`、`code=id&0x0f`。

## 待移植(依 SA1 盤點,優先序)

1. ~~**研究樹拓撲**~~ ✅ 已完成(`techtree.go`)。
2. ~~**殖民地人口成長**~~ ✅ 核心公式已完成(`colony.go`,手冊 p111)。生產/污染結算待。
3. **艦艇戰鬥解算** — 🟡 元件已備(BA=`BeamOffense`+`ShipCrewOffenseBonus`;BD=`BeamDefense`+`ShipCrewDefenseBonus`),
   **最終 to-hit% 組合待**:手冊 p117 給了三要素(BA/BD/Range Penalty)與 range→格對照表(每 3 格一級、PD 加倍、Hv 減半),
   但 base% 與各 range level 懲罰值尚未從手冊乾淨抽出 → 待精確查證後補,**不臆造**。需 RNG(命中擲骰)。
4. **外交/間諜/AI/回合引擎** — ❌ 全空白,最大工程,後期。手冊 p113 間諜、p123 飛彈閃避、p126 球形傷害均有公式。

> **關鍵轉折**:MOO2 patch 1.5 手冊 `moo2_patch1.5/MANUAL_150.html` 附錄(p110 起「The Algorithm」)
> **含具體公式與常數**(人口成長 FACTOR1=2000、間諜 Roll Chance、光束 Accuracy、飛彈 evasion…),
> 是**官方權威來源**——可據以移植(如 `colony.go`),非臆造。做法:逐公式從手冊抽全常數 → 手算對照測試 → 落地。
> openorion2 是渲染殼無這些邏輯(見記憶 [[openorion2-is-renderer-not-engine]]);1oom 是 MOO**1** 不適用。
> 已移植者:openorion2 唯讀正確資料表(科技樹、艦員/軍官/艦艇衍生值)+ 手冊權威公式(人口成長)。
