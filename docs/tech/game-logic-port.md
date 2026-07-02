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
| `spy.go` | 間諜 Roll Chance(分段封閉解)、派駐加成曲線、門檻、政府/種族/科技加成 | 手冊 p113-115 | `spy_test.go`(Opus 對手冊全等) |
| `combat.go` | 光束 to-hit:Range level、Range Penalty 0/0/10/20/30/40/55/70/85、`min(40+pen-PD,95)`、戰機防禦 | 手冊 p117-120 | `combat_test.go`(Opus 對手冊+worked example) |
| `missile.go` | 特殊防禦、飛彈閃避加成、jam 機率、AMR 命中、飛彈 Beam Defense | 手冊 p123-125,p117-120 | `missile_test.go`(含 AMR 反推手冊矛盾) |
| `production.go` | 污染容忍/清理、工廠/機器人/深核/回收加成 | GAME_MANUAL.pdf | `production_test.go`(假設處標註) |

### officer.go 公式重點
- **經驗等級**:門檻 `{60,150,300,500}` → 級 0..4。
- **技能加成**:`base = baseSkillValues[type][code]`;除 Megawealth(不隨等級倍增)外乘 `(expLevel+1)`;
  進階技能(tier 2)再 +50%;Navigator 用專屬值表。技能 id 位元編碼:`type=(id&0x30)>>4`、`code=id&0x0f`。

## 待移植(依 SA1 盤點,優先序)

1. ~~**研究樹拓撲**~~ ✅ `techtree.go`。
2. ~~**殖民地人口成長**~~ ✅ `colony.go`。~~**生產/污染**~~ ✅ `production.go`(GAME_MANUAL.pdf)。
3. ~~**光束命中 to-hit**~~ ✅ `combat.go`。~~**飛彈防禦/AMR**~~ ✅ `missile.go`。~~**間諜**~~ ✅ `spy.go`。
4. ~~**回合引擎編排**~~ 🟡 起步:`internal/engine`(殖民地經濟 `RunColonyTurn`:食物/工業/污染/研究/成長;
   研究進度 `RunResearchPhase`;帝國編排 `RunEmpireTurn`)。誠實界定:人口成長累積門檻、國庫 income 未驗證 → 只輸出不回寫。
5. **仍待**:傷害解算細節(dissipation/球形傷害 p126)、地面戰、外交關係演算、**AI 決策**、國庫 income 公式、
   人口成長累積尺度、save↔engine adapter。RNG 擲骰(命中/間諜/閃避)公式已給決定性機率/門檻,擲骰待建可重現 RNG。

> 全部已驗證公式彙整於 `docs/tech/moo2-formulas-reference.md`(8 系統 40 公式,附來源)。

> 已移植公式的共同保證:每條都對**權威來源**(openorion2 唯讀表 / patch 1.5 手冊)手算對照測試,Opus 逐條核實。
> 兩處手冊自相矛盾(AMR 命中率、飛彈速度公式 vs 表格)已在程式碼註解記錄推導/裁決,標待實機動態驗證。

> **關鍵轉折**:MOO2 patch 1.5 手冊 `moo2_patch1.5/MANUAL_150.html` 附錄(p110 起「The Algorithm」)
> **含具體公式與常數**(人口成長 FACTOR1=2000、間諜 Roll Chance、光束 Accuracy、飛彈 evasion…),
> 是**官方權威來源**——可據以移植(如 `colony.go`),非臆造。做法:逐公式從手冊抽全常數 → 手算對照測試 → 落地。
> openorion2 是渲染殼無這些邏輯(見記憶 [[openorion2-is-renderer-not-engine]]);1oom 是 MOO**1** 不適用。
> 已移植者:openorion2 唯讀正確資料表(科技樹、艦員/軍官/艦艇衍生值)+ 手冊權威公式(人口成長)。
