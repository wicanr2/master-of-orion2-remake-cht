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
// ============ 兩個修正項 ============
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
// ============ 誠實留白 ============
//
//   - **魅力 Charismatic 沒有數字。** 手冊只說「assimilate conquered colonists **easily**」,
//     patch 1.5 的手冊也沒有補。所以這裡**不給它任何效果**,而不是憑感覺塞一個 ×0.5。
//     `AssimilationTurns` 收了 charismatic 參數但目前不用它——留參數是為了讓找到數字的人
//     知道該改哪裡,而不是讓人以為這條規則不存在。
//   - 手冊還說異族管理中心「decreases the unrest of the unassimilated populations,
//     **halving the chance of revolt**」。remake 沒有叛亂系統,**這條沒接**。
//   - 多種族殖民地的 **20% 士氣懲罰**(建築可消除)另外走士氣那條路,不在這一檔。

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
var assimilationTurns = [8]int{
	8,  // 封建
	4,  // 邦聯
	8,  // 獨裁
	4,  // 帝國
	4,  // 民主
	2,  // 聯邦
	20, // 統一
	15, // 銀河統一
}

// AssimilationCenterTurns 是異族管理中心給的固定速率(手冊:1 per 2 turns,不分政府)。
const AssimilationCenterTurns = 2

// AssimilationRepulsiveMultiplier 是排斥種族的倍率(手冊:only half the normal rate → 回合數 ×2)。
const AssimilationRepulsiveMultiplier = 2

// AssimilationTurns 回傳同化一單位征服人口需要幾回合。
//
// hasCenter 為 true 時直接用異族管理中心的固定值(手冊「regardless of government」),
// 政體那一格完全不看。
//
// ⚠ charismatic 目前**沒有效果**——手冊沒給數字(見檔頭誠實留白)。參數留著是為了
// 標明「這條規則存在但缺數字」,找到數字的人改這一個函式就好。
func AssimilationTurns(gov AssimilationGovernment, hasCenter, repulsive, charismatic bool) int {
	turns := AssimilationCenterTurns
	if !hasCenter {
		if gov < 0 || int(gov) >= len(assimilationTurns) {
			gov = AssimDictatorship // 未知政體退回獨裁(remake 的預設政體),不是回 0
		}
		turns = assimilationTurns[gov]
	}
	if repulsive {
		turns *= AssimilationRepulsiveMultiplier
	}
	_ = charismatic // 見檔頭:手冊沒給數字,不臆造
	if turns < 1 {
		turns = 1
	}
	return turns
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

// AssimilationProgressNeeded 回傳同化 n 單位人口所需的總回合數。
//
// 抽出來是因為 UI 要顯示「這個殖民地還要幾回合才完全同化」——
// 一個只在背景默默跑的機制對玩家等於不存在。
func AssimilationProgressNeeded(unassimilated, turnsPerUnit int) int {
	if unassimilated <= 0 || turnsPerUnit <= 0 {
		return 0
	}
	return unassimilated * turnsPerUnit
}
