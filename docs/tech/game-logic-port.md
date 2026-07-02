# 遊戲邏輯層移植進度

## 背景與方法

SA1 盤點(`rules-implementation-audit.md`)確認:**openorion2 是渲染殼,不含回合引擎**,遊戲邏輯要在 go 端從零建。
方法依記憶 [先參考再逆向]:**優先移植已驗證的現成公式**(openorion2 中 `⬜ 僅資料結構` 者仍含正確的唯讀公式表),
而非自行推導;逆向/實測為最後手段。每條公式落 `internal/gamedata/`,配手算對照的單元測試。

## 已移植(`internal/gamedata/`)

| 模組 | 內容 | 來源(openorion2) | 測試 |
|---|---|---|---|
| `formulas.go` | 行星基礎生產、電腦/引擎 HP、戰速、光束攻防、軍官雇用費 | 各處(前期 Phase) | `formulas_test.go` |
| `officer.go` | 軍官經驗等級 `LeaderExpLevel`、技能加成 `LeaderSkillBonus` | `gamestate.cpp:607-701`(SA1 標記唯讀正確) | `officer_test.go`(exp 邊界 + 13 組技能手算) |
| `enums.go` | 行星/星體/研究/殖民等列舉 | `gamestate.h` | `enums_test.go` |

### officer.go 公式重點
- **經驗等級**:門檻 `{60,150,300,500}` → 級 0..4。
- **技能加成**:`base = baseSkillValues[type][code]`;除 Megawealth(不隨等級倍增)外乘 `(expLevel+1)`;
  進階技能(tier 2)再 +50%;Navigator 用專屬值表。技能 id 位元編碼:`type=(id&0x30)>>4`、`code=id&0x0f`。

## 待移植(依 SA1 盤點,優先序)

1. **研究樹拓撲** `research_choices[]`(`tech.cpp:169`)— SA1 標記唯讀正確,可直接抄成 go 資料表。
2. **殖民地成長/生產結算** — 手冊有明確公式;openorion2 僅資料結構(`⬜`),需對照 `MANUAL_150.html`。
3. **艦艇戰鬥解算** — ❌ 未實作且需 RNG(命中率);openorion2 零 `rand`,須自建(參考 1oom/手冊)。
4. **外交/間諜/AI** — ❌ 全空白,最大工程,後期。

> 注意:戰鬥、殖民成長、AI 決策都需**隨機成分**,openorion2 全無 RNG 來源(見 SA1 §關鍵發現),
> 這部分無現成公式可抄,須依手冊機率 + 自建可重現的 RNG(存檔相容性另議)。手冊來源:
> `moo2_patch1.5/MANUAL_150.html`(完整規則);`original_game/…CD Manual.pdf` 僅 9 頁安裝指南。
