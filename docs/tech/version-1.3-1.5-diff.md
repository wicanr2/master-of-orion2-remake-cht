# patch 1.3(1.31)→ patch 1.5(1.50)規則/數值差異 + 版本 profile 設計建議

> 2026-07-11。目的:為 CLAUDE.md 核心需求「主選單選擇 1.3 或 1.5」打底——先確認**真正需要分版的值有多少**,
> 再提版本 profile 資料結構。純研究,不改任何程式/資料檔。
>
> 方法(rulebook 62/65):以 `moo2_patch1.5/CHANGELOG_150.TXT`(1730 行,1.50.0→1.50.26 全部版本)
> 為主軸逐條過,聚焦「玩家會感受到的規則/數值變化」,跳過純 bug/crash/UI 技術修正;數字交叉核對
> `MANUAL_150.html`(1.50 patch notes 版,含明確 "1.31" 對照段落)與 `PARAMETERS.CFG`(3574 行,many
> 參數註解 `(default, classic)`,可直接判定「1.5 出廠預設是否等於 1.3」)。凡引用數字一律標行號/段落
> 出處;查無出處者標「待查」,不編造。

## 0. 結論先講:能落地程式碼的真正數值差異其實很少

逐條過完整份 CHANGELOG 後,**落在本專案已實作系統上、且是真正「數值改變」(非只是新增可調參數、
非純 timing/bookkeeping)的項目只有 3 條**(§2),另有 2 條屬「已實作但尚未接線」的系統,分版意義較低
(§3)。CHANGELOG 裡另有大量新增 `xxx_config_parameter`——這些**多半是把既有經典行為暴露成可調參數,
預設值本來就等於 1.3 經典值**(`PARAMETERS.CFG` 逐一標註 `(default, classic)`),不代表 1.5 出廠預設
真的改了規則。這對版本 profile 是好消息:第一版分版範圍可以很小(§6)。

## 1. 差異清單(全量表)

依「是否落在本專案已實作系統」排序;類別欄對應 WORKLIST.md 的系統分類。

| # | 類別 | 1.3(1.31)行為 | 1.5(1.50)行為 | 落在已實作系統? | 來源 |
|---|---|---|---|---|---|
| 1 | 研究成本(Hyper-Advanced Lv1) | **實際成本 15,000 RP**(顯示卻是 25,000,官方承認的 bug) | 顯示與實際皆 **25,000 RP**(1.50.9 修正) | ✅ 是,`internal/gamedata/techtree.go` 8 個 `TOPIC_HYPER_*` | CHANGELOG_150.TXT 1.50.9「Fixed actual tech cost of first level Hyper-Advanced from 15k to 25k.」+ MANUAL_150.html「Hyper-Advanced Tech Cost Bug: Fixed the cost of level 1 hyper-advanced tech fields that were shown as 25k research points but had a real cost of 15k. Now both actual and displayed cost is 25k RP.」 |
| 2 | 戰鬥傷害(電漿砲) | Damage **6–30** | Damage **4–20** | ✅ 是,`internal/shell/session.go` `WeaponOptions`「電漿砲」 | MANUAL_150.html「Plasma Cannon min/max damage from 6/30 to 4/20」(已收錄於 `docs/tech/component-values.md`) |
| 3 | 軌道轟炸命中換算 | Bomb 武器 = **5 次攻擊**當量;戰機 = **0**(無當量) | Bomb 武器 = **10 次攻擊**當量;戰機 = **1** 次 | ✅ 是,`internal/shell/orbital_bombardment.go` `fleetBombardDamage`(現行 `for round:=0;round<10` 即 1.5 值) | CHANGELOG_150.TXT 1.50.9「Fixed bomb hits calculation for orbital bombardment: Bomb weapons now get bomb hits equivalent to 10 instead of 5 attacks. Fighters get the equivalence of 1 strike instead of 0.」 |
| 4 | 經濟(新造運輸艦淨現金) | 完工時**淨得 +2~5 BC**(0-3 BC 立即成本 + 固定 5 BC 補償,官方手冊原文承認此為 1.3 就有的「一律有淨利」quirk,非等到 1.40 才算 bug) | `freighters_cash_bonus` 出廠預設 **0**(1.50.8 起),淨得 0 BC | ⚠ 半實作——`gamedata.IncomeFreighterMaintenanceCost`(每回合 0.5 BC/艘持續維護費)已寫好,但「完工當下的一次性現金效果」整條機制(對應手冊 `freighters_cash_bonus`)本專案**完全沒有程式碼**(不追蹤運輸艦建造事件),不算「已實作」 | MANUAL_150.html「Buildings & Freighters Free Cash Bug」全段(1.31/1.40/1.50 三欄對照表)+ CHANGELOG 1.50.8「Changed freighters_cash_bonus default from 5 to 0 BC」 |
| 5 | 地面戰:防禦方指揮官加成 | 防禦方 Commando 技能領袖**無**額外加成(僅攻方有 2.5x) | 新增:防禦方 Commando 領袖也給 **2.5x** 加成(攻方不變) | ❌ 否——本專案 `ResolveGroundBattle`(1oom 公式+手冊加成表)完全沒有領袖加成項 | MANUAL_150.html「Commando Leader: A defending commando gives 2.5x the regular commando bonus to ground troops, just like an attacking commando already gives in classic.」 |
| 6 | 地面戰:commando 倍率門檻 | 攻方 5x/7.5x、守方 2x/3x(依技能等級) | **出廠預設不變**,只是新增 `ground_commando_attacker_x2`/`ground_commando_defender_x5` 讓玩家可調換 | ❌ 否(且屬確認非差異) | `PARAMETERS.CFG:2745-2753`「(default, classic)」逐條標註 |
| 7 | 轟炸:建築 hits 加成 | 未記錄文件的 **+1 hit** bonus(bug) | 移除該 +1 bug | ❌ 否——本專案軌道轟炸只扣人口,不扣建築(見 `docs/tech/ground-combat-algorithm.md`「範圍限制」),此差異對本專案模型無作用 | CHANGELOG_150.TXT 1.50.10「Undocumented +1 hit bonus for civilian buildings during bombardment removed.」 |
| 8 | 轟炸:建築/人口裝甲值 | `civilian_armor`(非防禦建築/人口單位)= **100 hp**(所有裝甲等級皆同) | 出廠預設不變,只是暴露成可調參數 | ❌ 否(且屬確認非差異) | `PARAMETERS.CFG:1778-1786`「Default is 100 hp regardless of armor (classic).」 |
| 9 | 地面戰:防禦建築結構倍率 | `ground_defense_armor_multiplier` = **100**(對應鈦裝甲等級 100 結構點) | 出廠預設不變 | ❌ 否(確認非差異) | `PARAMETERS.CFG:1772-1775`「Default is 100 ... (classic).」 |
| 10 | 研究:突破隨機性 | `fixed_research_cost=0`(有隨機突破機率) | 出廠預設不變 | ❌ 否(確認非差異;本專案研究系統本來就沒模擬「突破機率」這件事,只有固定 RP 成本) | `PARAMETERS.CFG:542-545`「(default, classic)」 |
| 11 | 轟炸:炸彈換算的行星大小分級 | 3-4-6-7-8(classic) | 中途版本(1.50.4 起)曾錯改,**1.50.11 已修回同一組 3-4-6-7-8** | ❌ 否——本專案不模擬行星尺寸對轟炸區域的幾何影響;且此差異在 1.5 系列內部自我修正,對 1.3 vs 最終 1.5 而言**不構成差異** | CHANGELOG_150.TXT 1.50.11「Restored planet sizes for bombardment to classic 3-4-6-7-8.」 |
| 12 | 開局:起始偵察艦戰鬥速度 | 手動設計與自動設計艦速度不一致 bug,起始 2 艘 Scout 戰鬥速度為 **10** | 修正後為 **12**(空間格利用速度加成一致套用) | ⚠ 邊緣——本專案 `CombatSpeed()` 是通用公式(見 `formulas.go`),未特別為「自動設計 vs 手動設計」分流,無法乾淨對應此 bug;可視為低優先 | CHANGELOG_150.TXT 引文見 `docs/tech/homeworld-init.md:111`(已收錄) |
| 13 | 掃描/通訊距離 | Tachyon/Neutron/Sensors 等偵測科技的顯示值與實際值有 1 格落差(如 Tachyon 顯示 3、實際 4) | 修正「顯示值=實際值」,多數偵測距離**預設值 +1** | ❌ 否——本專案完全未實作掃描/偵測範圍機制(WORKLIST 無此項) | MANUAL_150.html「Scanners and Communications Discrepancy」表(見下方 §4 附註,表格經 HTML 去標籤後欄位對不齊,**精確數字待用原始 HTML `<table>` 結構重新萃取,本表僅供方向性參考**) |
| 14 | 衛星/地面砲台佔格 | 光束武器在衛星 arc cost +25%、地面砲台 +0% | 修正為統一 arc cost,衛星/地面砲台空間分別 +40%/+50%(1.50.7),之後衛星再由 +40%→+33.3%(1.50.10) | ❌ 否——本專案 `ship-design-space.md` 只做 6 級標準艦體(Frigate…Doom Star),無獨立 Satellite/Missile Base/Ground Battery 船體類別 | CHANGELOG_150.TXT 1.50.7、1.50.10 |
| 15 | 新建築/新間諜維護費入帳時機 | 完工當回合立即扣費 + 補償(淨額 0) | 改為下回合扣費、取消補償(淨額同樣 0) | ❌ 否(且屬淨額非差異——只差「哪一回合帳上出現」,本專案 `BuiltMaintenanceBC` 是逐回合依「目前已建成清單」重算,不模擬「完工瞬間」這個時間點,模型顆粒度不到這一層) | MANUAL_150.html「Buildings & Freighters Free Cash Bug」表 |

## 2. 落在已實作系統上的差異(核心,共 3 條——見上表 #1–#3)

這 3 條全部滿足:(a) 官方文件白紙黑字給出具體數字,(b) 對應到本專案**目前正在跑的程式碼路徑**
(非死碼、非未來計畫),(c) 我們現有實作值 = 1.5 的值(因為 `techtree.go`/`session.go`/
`orbital_bombardment.go` 的資料來源本來就是 patch 1.5 隨附的 `GAME_MANUAL.pdf`/`MANUAL_150.html`)。
換言之:**現況 = 事實上的「1.5 profile」**,要支援 1.3 選項,只需要一組「1.3 覆寫值」:

| 值 | 現有程式碼(= 1.5) | 1.3 覆寫值 |
|---|---|---|
| Hyper-Advanced Lv1 研究成本(8 個 `TOPIC_HYPER_*` 主題共用) | 25000 | 15000 |
| 電漿砲 Damage(`WeaponOptions` 電漿砲 `Value`) | 20 | 30 |
| 軌道轟炸模擬齊射輪數(`fleetBombardDamage` for 迴圈上限) | 10 | 5 |
| 軌道轟炸戰機命中當量(目前本專案未區分戰機類型,無對應變數——見下方待辦) | — | — |

電漿砲的手冊數字其實是「min/max 各改一次」(6→4、30→20),但本專案 `Component` 結構只有單一
`Value`(代表最大傷害)欄位,沒有 `MinValue`——目前的 20/30 對照只能覆寫「最大傷害」這一項;
若要精確重現最小傷害(4 vs 6),需要先幫 `Component` 加一個欄位,這是比「分版」更大的既有資料模型
擴充,本檔誠實標記為「1.3 profile 只能做到最大傷害對齊,最小傷害維持現行單值模型的既有限制」。

軌道轟炸「戰機命中當量」目前本專案的 `fleetBombardDamage` 是逐艦逐輪算傷害,沒有「戰機 vs 非戰機
武器」的分流(不像 `weaponKindByName` 那樣分 beam/missile/spherical),故 1.3 的「戰機 0 當量」/
1.5 的「戰機 1 當量」這條差異目前**無法對應到任何變數**——這是「發現了差異,但我們連 1.5 值本身
都还没有为它建模」的情況,誠實標「待未來戰機武器分類任務接線後再談分版」。

## 3. 已實作但尚未接線的系統(§1 #4,經濟)

`gamedata.IncomeFreighterMaintenanceCost`(每回合持續維護費 0.5 BC/艘,手冊 p.169 有據)已寫好但
零呼叫端(`colony-economy-maintenance.md` §1 已記錄「本專案完全不追蹤運輸艦數量」)。手冊裡真正
1.3→1.5 有差的是**另一條機制**——「完工當下一次性補償」(`freighters_cash_bonus`,1.3 淨賺
2-5 BC、1.5 淨賺 0 BC)——這條機制本專案完全沒有程式碼,不是「差 0 vs 5」的分版問題,是「這個
機制本身还不存在」。若未來要做,設計時直接用 1.5 的行為(淨額 0,最簡單)當預設,1.3 覆寫值才有
意義;現在談分版為時過早。

## 4. shipped-game 預設值來源(PARAMETERS.CFG `(default, classic)` 交叉核對)

`PARAMETERS.CFG`(patch 1.5 內附,3574 行)對每個可調參數多附**「Default is X (classic)」**這類註解,
代表「1.5 出廠預設值就是經典 1.3 的值,這個參數只是把它暴露出來讓玩家/mod 調」。逐一核對後確認
以下項目 **1.5 出廠預設 = 1.3 經典行為,非真差異**(對應上表 #6/#8/#9/#10):

- `ground_commando_attacker_x2`/`ground_commando_defender_x5`(地面戰指揮官加成倍率)——`PARAMETERS.CFG:2745-2753`
- `civilian_armor`(轟炸建築/人口裝甲值 100hp)——`PARAMETERS.CFG:1778-1786`
- `ground_defense_armor_multiplier`(地面防禦建築結構倍率 100)——`PARAMETERS.CFG:1772-1775`
- `fixed_research_cost`(研究突破隨機性,預設仍隨機)——`PARAMETERS.CFG:542-545`

這對「盡量對照 1.5 預設 vs 1.3 原值」的任務要求(§3)而言,誠實的答案是:**這幾項刻意查過,結果
是「找到明確數字,但數字本身兩版相同」**——不是找不到,是差異不存在。詳細版本消歧陷阱(為何不能
直接盲抄 CFG,`##` 標記的是「improved」mod 值 vs classic,不是「1.3 vs 1.5」这条轴)已由既有文件
`docs/tech/patch15-cfg-data-source.md` 完整記錄,本檔沿用同一套紀律,不重複展開。

## 5. 版本 profile 資料結構設計建議

### 5.1 最小可行設計(對應§2 三個確證差異)

```go
// internal/gamedata/ruleprofile.go(建議新檔,尚未實作,設計稿)

package gamedata

// GameVersion 對應 CLAUDE.md「主選單選擇 1.3 或 1.5」的兩個選項。
type GameVersion int

const (
	VersionClassic13  GameVersion = iota // 官方最後正式 patch 1.31
	VersionCommunity15                    // 社群非官方 patch 1.50(本專案現行資料的預設來源)
)

// RuleProfile 收斂「已確證、且落在已實作系統上」的版本相依數值。
// ⚠ 刻意保持精簡——只收 docs/tech/version-1.3-1.5-diff.md §2 列出的 3 條,不要為了「看起來完整」
// 塞進未確證或未接線的欄位。新差異確證後才加欄位,見 §5.3 擴充路徑。
type RuleProfile struct {
	Version GameVersion

	// 研究:Hyper-Advanced 第一級科技(8 個 TOPIC_HYPER_* 主題共用同一個成本)。
	// 來源:MANUAL_150.html「Hyper-Advanced Tech Cost Bug」+ CHANGELOG_150.TXT 1.50.9。
	HyperAdvancedLevel1Cost int

	// 戰鬥:電漿砲最大傷害(Component.Value)。來源見 component-values.md。
	// 注意:手冊同時記載最小傷害 6→4,但 Component 結構目前只有單一 Value(最大傷害)欄位,
	// 無法表示最小值差異——這是既有資料模型限制,非本 profile 遺漏。
	PlasmaCannonMaxDamage int

	// 軌道轟炸:fleetBombardDamage 模擬齊射的輪數。來源:CHANGELOG_150.TXT 1.50.9。
	BombardmentVolleys int
}

func Profile13() RuleProfile {
	return RuleProfile{
		Version:                 VersionClassic13,
		HyperAdvancedLevel1Cost: 15000,
		PlasmaCannonMaxDamage:   30,
		BombardmentVolleys:      5,
	}
}

func Profile15() RuleProfile {
	return RuleProfile{
		Version:                 VersionCommunity15,
		HyperAdvancedLevel1Cost: 25000, // = 現行 techtree.go 硬編值
		PlasmaCannonMaxDamage:   20,    // = 現行 session.go 硬編值
		BombardmentVolleys:      10,    // = 現行 orbital_bombardment.go 硬編值
	}
}
```

### 5.2 接線點(現有程式碼要改的三處,設計層面,尚未實作)

| 現有程式碼 | 目前寫法 | 建議改法 |
|---|---|---|
| `internal/gamedata/techtree.go` 8 個 `TOPIC_HYPER_*` 條目 | `Cost: 25000` 硬編 | `researchChoices` 這 8 條的 `Cost` 改成讀 `activeProfile.HyperAdvancedLevel1Cost`(需把 `researchChoices` 從套件級 `var` 改成依 profile 產生的函式,或在 `ResearchChoiceFor` 內對這 8 個 topic 特判覆寫) |
| `internal/shell/session.go` `WeaponOptions` 電漿砲 | `{"電漿砲", 200, 20, ...}` 硬編 | 建構 `WeaponOptions` 時對電漿砲那一列的 `Value` 改讀 `activeProfile.PlasmaCannonMaxDamage`(`Component` 陣列從套件級 `var` 改成 `func BuildWeaponOptions(p RuleProfile) []Component`,呼叫端一次性建構,不必每次查表) |
| `internal/shell/orbital_bombardment.go` `fleetBombardDamage` | `for round := 0; round < 10; round++` 硬編 | 迴圈上限改讀 `s.RuleProfile.BombardmentVolleys`(`GameSession` 需新增 `RuleProfile` 欄位,由 `NewDemoSession`/新遊戲流程注入) |

`GameSession` 需要新增一個 `RuleProfile` 欄位(建構時由主選單選擇結果注入),這是唯一貫穿全專案的
新增狀態;`RuleProfile` 本身應視為**唯讀設定**,遊戲開始後不可變(避免中途切版本造成存檔/平衡
不一致——原版本身也是「一開局就決定規則集」,無 mid-game 切換)。

### 5.3 擴充路徑(未來新差異確證後怎麼加)

1. 新差異先進本檔 §1 表格(來源標行號/段落),分類「已實作系統」/「未實作」。
2. 只有「已實作系統」且「兩版數字真的不同」(§4 教訓:很多新增參數的 1.5 預設=1.3 經典值,不算差異)
   才加進 `RuleProfile` 欄位 + `Profile13()`/`Profile15()`。
3. 若差異牽涉到現有資料結構容不下的維度(如電漿砲最小傷害,`Component` 沒有 `MinValue` 欄位)——
   分兩步走:先評估「值不值得為此擴充資料結構」,再擴充,不要為了塞版本差異而扭曲既有結構。
4. `RuleProfile` 預期會隨著本專案更多系統忠實化(戰機分類、衛星船體、運輸艦建造事件等)持續變大,
   但**不要提前為未確證/未接線的系統占位欄位**——空欄位比缺欄位更容易誤導後續開發誤以為「已考慮
   版本差異」。

## 6. 第一版最小分版建議(可直接排入 WORKLIST)

**只做 §2 三個值**:`HyperAdvancedLevel1Cost`(15000/25000)、`PlasmaCannonMaxDamage`(30/20)、
`BombardmentVolleys`(5/10)。理由:

- 三者皆有官方文件逐字數字佐證(非社群逆向、非推測)。
- 三者皆落在本專案「已在跑」的程式碼路徑上,接線成本低(見 §5.2,三處都是把硬編常數換成讀
  profile 欄位,無需改資料流向或新增系統)。
- 三者對玩家可感知(研究成本、武器傷害、轟炸強度),符合任務「玩家會感受到的規則變化」門檻。
- **誠實揭露**:這是一個很小的第一版——多數 CHANGELOG 條目要嘛是純 bug fix(如衛星生成 bug)、
  要嘛是新增的可調參數但預設值等於經典值(§4)、要嘛落在本專案尚未實作的系統(衛星/戰機分類/
  領袖加成/運輸艦建造事件等,§1 表格 #4–#14)。**這不是研究不夠深入,是逐條核實後的真實結論**——
  1.3/1.5 對「本專案目前實作範圍」而言差異確實很小,之後每完成一個新系統(如衛星船體、戰機分類),
  應回頭比對 CHANGELOG 是否有該系統的版本差異,再擴充 `RuleProfile`。

## 7. 主選單「選版本」與 profile 的關係(範圍澄清)

CLAUDE.md 要求「主選單選擇版本 1.3 or 1.5」——本檔的 `RuleProfile` 解決的是**選版本後,遊戲規則
數值要跟著變**這一半;「主選單 UI 本身要有這個選項」是另一半(UI/流程層,不在本檔研究範圍,但
握手位置很單純:新遊戲流程在建立 `GameSession` 前先決定 `GameVersion`,傳入
`Profile13()`/`Profile15()` 其中之一)。兩者可分開排入 WORKLIST 的不同任務。

## 8. 來源清單

- `moo2_patch1.5/CHANGELOG_150.TXT`(1730 行,1.50.0–1.50.26 全部版本逐條核對)
- `moo2_patch1.5/MANUAL_150.html`(1.50 patch notes 版手冊,經 `python3 -re` 去 HTML 標籤後全文
  關鍵字比對;「Scanners and Communications Discrepancy」表格因去標籤後欄位錯位,標記待用原始
  `<table>` 結構重新萃取)
- `moo2_patch1.5/MOO2-1.50.26.zip` 內 `patch/150/docs/PARAMETERS.CFG`(3574 行,`(default, classic)`
  註解逐條核對 §4 四項)
- 既有專案文件(交叉引用,未重複研究):`docs/tech/component-values.md`(電漿砲差異原始發現)、
  `docs/tech/patch15-cfg-data-source.md`(CFG 版本消歧陷阱,本檔沿用其紀律)、
  `docs/tech/ship-design-space.md`(艦體/武器空間表,確認衛星非本專案範圍)、
  `docs/tech/ground-combat-algorithm.md`(地面戰現行公式,確認領袖加成未實作)、
  `docs/tech/research-system-status.md`(研究成本現行接線點)、
  `docs/tech/colony-economy-maintenance.md`(運輸艦維護費現行接線狀態)、
  `docs/tech/homeworld-init.md`(起始偵察艦速度差異原始發現)、
  `docs/tech/rules-implementation-audit.md`(已預先點名本任務為 Phase 7 待辦)
