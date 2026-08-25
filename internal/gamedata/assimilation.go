package gamedata

// assimilation.go:**征服人口的同化**——手冊把整張表逐個政府寫死了。
//
// 攻下一個殖民地之後,那裡的人口不會馬上變成你的子民。原版每個政府同化一單位人口
// 要花不同的回合數,而且差距很大(民主 4 回合 vs 統一 20 回合)——
// 那是「征服打法」與「和平打法」在規則層的分野。
//
// ============ 逐政府的回合數(手冊 p.21–24 逐條)============
//
//	| 政府 | 同化一單位人口 | 手冊原文 |
//	|---|---|---|
//	| 封建 Feudal | 8 | It takes 8 turns for a Feudal government to assimilate a unit of population |
//	| 邦聯 Confederation | 4 | Assimilation of conquered colonists takes only 4 turns |
//	| 獨裁 Dictatorship | 8 | It takes 8 turns for a Dictatorship to assimilate a unit |
//	| 帝國 Imperium | 4 | Assimilation of conquered colonists takes only 4 turns |
//	| 民主 Democracy | 4 | It takes only 4 turns for a Democratic government |
//	| 聯邦 Federation | 2 | Assimilation of a unit of conquered population takes only 2 turns |
//	| 統一 Unification | 20 | It takes 20 turns for a Unified government |
//	| 銀河統一 Galactic Unification | 15 | Assimilation time is reduced to 15 turns |
//
// 四個進階政體(邦聯/帝國/聯邦/銀河統一)是基礎政體研究出來的升級版,
// 所以這張表是「4 個基礎 + 4 個進階」而不是八個獨立選項。
//
// ============ 兩個修正項（IDA Pro 已閉合）============
//
//	排斥 Repulsive        「assimilate conquered colonists … at only **half** the normal rate」
//	                       → 回合數 ×2
//	異族管理中心(建築)   「assimilates conquered populations at the rate of **1 per 2 turns**,
//	                       **regardless of government**」
//	                       → 直接蓋掉政府那一格,固定 2 回合
//
// 建築那條的「regardless of government」很關鍵:對統一政體(20 回合)來說,
// 蓋一座異族管理中心等於把同化速度變成十倍。一棟維護費 1 BC 的建築有這種效果,
// 是因為統一政體的懲罰本來就設計成「你不該靠征服玩」。
//
// `sub_E3456 @ 0xE3456` 證實原版保存的是 0..239 進度點：Charismatic 將每回合
// 進度加倍；只有它不存在時 Repulsive 才把進度減半。完整證據見
// docs/re/assimilation-race-traits-audit-20260825.md。
//
// 異族管理中心的叛亂機率減半已由 shell/rebellion.go 消費；多種族殖民地的
// 20% 士氣懲罰（建築可消除）也另走士氣鏈，不在本純公式檔重複計算。

// AssimilationGovernment 是同化速率表用的政體(含四個進階形式)。
//
// ============ 這個編號不是自己排的,執行檔驗證過 ============
//
// 下面的順序原本是照手冊表格排的(基本型 + 它的進階型交錯),**是 remake 自己的選擇**。
// 2026-08-08(第 54 項(三個寫入端))發現原版用的是同一組編號:`sub_E4204`(「取得某項科技」的
// 效果套用函式)對四項政府科技各寫一個立即數進 `[player+0x89F]`:
//
//	techIdx 42 Confederation         → [player+0x89F] = 1   (asm 327016)
//	techIdx 92 Imperium              → [player+0x89F] = 3   (asm 327006)
//	techIdx 65 Federation            → [player+0x89F] = 5   (asm 327021)
//	techIdx 77 Galactic Unification  → [player+0x89F] = 7   (asm 327011)
//
// 四項全中,一項不差。偶數是四個基本政體(封建/獨裁/民主/統一——正好是自訂種族
// 那一欄能選的四個),奇數是它們各自的科技升級版,而 `值/2` 就是「哪一族」。
//
// 這同時**推翻了 `docs/re/calc-tech-value.md` 把 `[player+0x89F]` 記成「種族特性相關(猜)」**
// ——那份文件當時寫「沒有查到寫入端」,而寫入端一直都在。
//
// 順帶:`lea esi, [eax+89Fh]` 與 `inc byte ptr [edx+eax+89Fh]` 顯示 0x89F 是**一段陣列的起點**
// (到 0x8BE),政府只是第 0 格。其餘各格是什麼仍未解。
//
// ⚠ 不要重排這個列舉:重排會讓上面那組對照失效,而它是目前唯一把 remake 的政體編號
// 釘在原版位元組上的東西(`government_orig_test.go`)。
type AssimilationGovernment int

const (
	AssimFeudal AssimilationGovernment = iota
	AssimConfederation
	AssimDictatorship
	AssimImperium
	AssimDemocracy
	AssimFederation
	AssimUnification
	AssimGalacticUnification
)

// assimilationTurns 是逐政體同化一單位人口所需的回合數(手冊逐條,見檔頭表)。
var assimilationRates = [8]int{
	30, 60, 30, 60, 60, 120, 12, 16,
}

// AssimilationProgressThreshold 是原版 `0xE353E cmp ...,0F0h` 的逐人口門檻。
const AssimilationProgressThreshold = 240

// AssimilationCenterTurns 是異族管理中心給的固定速率(手冊:1 per 2 turns,不分政府)。
const AssimilationCenterTurns = 2

// AssimilationRepulsiveMultiplier 是排斥種族的倍率(手冊:only half the normal rate → 回合數 ×2)。
const AssimilationRepulsiveMultiplier = 2

// AssimilationRate 回傳每回合加入殖民地 raw 同化進度的點數。
func AssimilationRate(gov AssimilationGovernment, hasCenter, repulsive, charismatic bool) int {
	if gov < 0 || int(gov) >= len(assimilationRates) {
		gov = AssimDictatorship
	}
	rate := assimilationRates[gov]
	if hasCenter {
		rate = AssimilationProgressThreshold / AssimilationCenterTurns
	}
	if charismatic {
		rate *= 2
	} else if repulsive {
		rate /= 2
	}
	if rate < 1 {
		return 1
	}
	return rate
}

// AssimilationTurns 回傳同化一單位征服人口需要幾回合。
//
// hasCenter 為 true 時直接用異族管理中心的固定值(手冊「regardless of government」),
// 政體那一格完全不看。
//
// ⚠ charismatic 目前**沒有效果**——手冊沒給數字(見檔頭誠實留白)。參數留著是為了
// 標明「這條規則存在但缺數字」,找到數字的人改這一個函式就好。
func AssimilationTurns(gov AssimilationGovernment, hasCenter, repulsive, charismatic bool) int {
	rate := AssimilationRate(gov, hasCenter, repulsive, charismatic)
	return (AssimilationProgressThreshold + rate - 1) / rate
}

// AssimilationAdvancedForm 回傳基礎政體在「已研究出進階形式」時對應的政體。
//
// 原版的四個進階政體是基礎政體研究出來的升級版,不是新遊戲的選項——
// 所以呼叫端傳「目前的基礎政體 + 有沒有那個科技」,而不是自己去對照。
func AssimilationAdvancedForm(base AssimilationGovernment) AssimilationGovernment {
	switch base {
	case AssimFeudal:
		return AssimConfederation
	case AssimDictatorship:
		return AssimImperium
	case AssimDemocracy:
		return AssimFederation
	case AssimUnification:
		return AssimGalacticUnification
	}
	return base
}

// AssimilationRemainingTurns 以原版 raw 進度計算全部剩餘人口 ETA。
func AssimilationRemainingTurns(unassimilated, progress, rate int) int {
	if unassimilated <= 0 || rate <= 0 {
		return 0
	}
	points := unassimilated*AssimilationProgressThreshold - progress
	if points <= 0 {
		return 1
	}
	return (points + rate - 1) / rate
}
