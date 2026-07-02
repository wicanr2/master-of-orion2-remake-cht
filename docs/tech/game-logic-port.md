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
2. **殖民地成長/生產結算** — openorion2 **無公式**(只從存檔讀 `pop_growth` 欄位,SA1 標 ⬜),
   須從手冊 `MANUAL_150.html` 推導,屬**須查證**工作,不可盲抄。
3. **艦艇戰鬥解算** — ❌ 未實作且需 RNG(命中率);openorion2 零 `rand`,無現成公式,須依手冊機率自建。
4. **外交/間諜/AI** — ❌ 全空白,最大工程,後期。

> **重要範圍界定(避免臆造規則)**:2-4 在 openorion2 內**沒有可移植的驗證公式**(SA1 已證其為渲染殼)。
> 1oom 是 Master of Orion **1**,機制與 MOO2 不同,**不是** MOO2 公式來源(見記憶 [[openorion2-is-renderer-not-engine]])。
> 因此這三項須以手冊機率為據**逐條查證後實作 + 自建可重現 RNG**,不能像科技樹那樣逐字轉寫。
> 手冊來源:`moo2_patch1.5/MANUAL_150.html`(完整規則);`original_game/…CD Manual.pdf` 僅 9 頁安裝指南。
> 目前已移植者皆為 openorion2 中「資料結構雖 ⬜、但唯讀公式/資料表正確」的部分(科技樹、艦員/軍官/艦艇衍生值)。
