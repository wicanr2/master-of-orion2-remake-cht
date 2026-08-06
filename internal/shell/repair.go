package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// repair.go:艦艇損傷與修復。
//
// remake 先前**完全沒有艦艇損傷這個概念**——一艘船不是完好就是被擊沉,打完一場慘勝的仗
// 之後,倖存的船跟出港時一模一樣。於是「自動修復」這個元件(SpecialOptions 裡的
// `{"自動修復", 60, 0, TOPIC_ADVANCED_MANUFACTURING, TECH_AUTOMATED_REPAIR_UNIT}`)
// 從加進來的那天起就沒有任何效果:沒有損傷可修。
//
// --- 手冊逐字(GAME_MANUAL.pdf)---
//
//	p.82 Automated Repair Unit:"Each combat round, this system takes a number of points equal
//	     to 20% of the ship's armor and structural damage and restores that number of points,
//	     first to the ship's structure, then any leftover is applied to the armor. The unit also
//	     repairs 10% of the damage to the ship's internal systems … In addition, any ship
//	     equipped with an Automated Repair Unit is completely repaired after every battle."
//	p.80 Advanced Damage Control:"The ADC unit repairs a ship completely after every battle."
//	p.25 Cybernetic:"During combat … fix armor and structural damage at 10% per round and
//	     systems damage at 5% per round. And after any combat, they repair their ships completely."
//
// --- 反組譯 ---
//
//	Repair_Ships_At_Colonies_ @ 0x580F5 → 對符合條件的船呼叫 `Repair_Ship_Full_` @ 0x581F3。
//	**是「完全修復」而不是逐回合慢慢修**;條件包含「船停在某顆星、而那顆星上有自己的殖民地」
//	(那段 `imul edx, 71h`(恆星 stride)+ `sub_1276F0` 的查詢)。
//
// --- 交叉驗證:openorion2 讀存檔的 `struct Ship`(gamestate.h:1268)---
//
//	uint8_t  shieldDamage, driveDamage;   // percent
//	uint8_t  computerDamage, crewLevel;
//	uint8_t  damagedSpecials[(MAX_SHIP_SPECIALS+7)/8];
//	uint16_t armorDamage, structureDamage;
//
// 這是原版存檔的真實佈局,把「remake 沒做哪些」逐欄位講白了:原版每艘船記**六份**損傷
// (護盾/引擎/電腦/逐元件旗標/裝甲/結構),remake 只有 `structureDamage` 這一份。
// ships.cpp:1060 的 `isSpecialDamaged(i)` 決定艦艇資訊面板把該元件名字畫成損壞色,
// 那是逐元件損傷唯一的 UI 出口——remake 沒有逐元件狀態,自然也沒有那個顯示。
//
// ⚠ remake 沒有建模、誠實留白(不是漏做,是抽象層級不同):
//   - 「內部系統損傷」(引擎/武器/護盾/電腦/各元件各自的損壞度)——remake 的戰鬥是艦級抽象,
//     沒有逐系統狀態,所以手冊那句「systems damage 10%/5% per round」無處可套。
//     原版對應的 `Apply_Internal_Damage_` @ 0x35251 是依傷害類型分十幾條分支、分別去打
//     艦艇結構 +0x29/+0xC2/+0x134 等欄位的大函式,要接它得先有逐系統模型。
//   - 裝甲與結構分離:手冊說先修結構再修裝甲,remake 只有單一「損傷」值,兩者合一。

// ShipDamageFloorHP 是受損艦在戰鬥中的最低有效血量——損傷再重也留 1 點,
// 避免「還沒開打就已經是 0 血」這種不合理狀態(該沉的船在上一場就已經沉了)。
const ShipDamageFloorHP = 1

// AutoRepairPercentPerRound 是自動修復元件每回合修復的結構損傷比例(手冊 p.82:20%)。
const AutoRepairPercentPerRound = 20

// CyberneticRepairPercentPerRound 是機械化種族每回合修復的結構損傷比例(手冊 p.25:10%)。
// ⚠ remake 目前沒有「機械化」種族特質欄位,這個常數尚無呼叫端——留著是為了讓手冊數字有地方
// 記錄,並在種族特質系統補上時直接可用。
const CyberneticRepairPercentPerRound = 10

// shipMaxHP 回傳一艘船未受損時的戰鬥血量(與 mkPlayerCombatants 的算法一致)。
func shipMaxHP(sh Ship) int {
	hp := shipStrength(sh.Class) * 3
	switch sh.Special {
	case "戰機庫":
		_, fhp := gamedata.FighterBayCombatContribution()
		hp += fhp
	case "重戰機庫":
		_, fhp := gamedata.FighterHeavyBayCombatContribution()
		hp += fhp
	}
	return hp + sh.BonusHP
}

// shipHasAutoRepair 回傳該船是否裝有自動修復元件(手冊 p.82)。
func shipHasAutoRepair(sh Ship) bool { return sh.Special == "自動修復" }

// playerHasAdvancedDamageControl 回傳玩家是否已研究進階損害管制(手冊 p.80:戰後完全修復)。
//
// ⚠ remake 的科技樹目前沒有獨立的「進階損害管制」主題,故恆為 false——標明是缺科技,
// 不是這條規則沒接。等該主題進科技樹時,這裡是它唯一的掛勾點。
func (s *GameSession) playerHasAdvancedDamageControl() bool { return false }

// ShipDamage 回傳一艘船目前的結構損傷(0 = 完好)。
func ShipDamage(sh Ship) int {
	if sh.Damage < 0 {
		return 0
	}
	if max := shipMaxHP(sh); sh.Damage > max-ShipDamageFloorHP {
		return max - ShipDamageFloorHP
	}
	return sh.Damage
}

// ShipDamagePercent 回傳一艘船的結構損傷百分比(0 = 完好)。供 UI 顯示——原版的艦隊列表
// 也是把損傷畫在艦艇資訊面板上(ships.cpp:1060 用損壞色標出壞掉的元件),打完仗有傷卻
// 沒地方看得到的話,這個系統對玩家等於不存在。
func ShipDamagePercent(sh Ship) int {
	max := shipMaxHP(sh)
	if max <= 0 {
		return 0
	}
	return ShipDamage(sh) * 100 / max
}

// repairShipFull 把一艘船修到完好(原版 `Repair_Ship_Full_`)。
func repairShipFull(sh *Ship) { sh.Damage = 0 }

// advanceShipRepair 每回合結束時修復艦艇(原版 `Repair_Ships_At_Colonies_`)。
//
// 規則:艦隊停在「自己有殖民地或前哨站的星」就**完全修復**——原版就是直接呼叫
// Repair_Ship_Full_,不是逐回合慢慢修。remake 的艦隊是單一集合(沒有多支艦隊各自位置),
// 故條件簡化為「艦隊目前所在星是自己的據點」。
//
// 回傳被修復的艦艇數(供回合摘要顯示;0 = 這回合沒修到)。
func (s *GameSession) advanceShipRepair() int {
	if s.FleetETA > 0 || s.FleetAtStar < 0 {
		return 0 // 航行中沒有港口可修
	}
	if !s.starIsPlayerBase(s.FleetAtStar) {
		return 0
	}
	n := 0
	for i := range s.Ships {
		if s.Ships[i].Damage > 0 {
			repairShipFull(&s.Ships[i])
			n++
		}
	}
	return n
}

// starIsPlayerBase 回傳該星是否為玩家的據點(有殖民地或前哨站)。
func (s *GameSession) starIsPlayerBase(starIdx int) bool {
	for _, st := range s.PlayerColonyStars {
		if st == starIdx {
			return true
		}
	}
	return s.HasOutpostAt(starIdx)
}

// repairAfterBattle 依手冊 p.80/p.82 做戰後修復:裝有自動修復元件的船、或玩家已研究
// 進階損害管制時的所有船,戰鬥結束後完全修復。
func (s *GameSession) repairAfterBattle() {
	adc := s.playerHasAdvancedDamageControl()
	for i := range s.Ships {
		if adc || shipHasAutoRepair(s.Ships[i]) {
			repairShipFull(&s.Ships[i])
		}
	}
}

// applyBattleDamage 把一場戰鬥後的剩餘血量寫回艦艇的持久損傷。
//
// combatIdx 是「第 k 個參戰艦」對應到 s.Ships 的索引(mkPlayerCombatantsIndexed 產生);
// remaining 是戰鬥結束時該艦剩下的血量,-1 表示已被擊沉(呼叫端另行移除)。
func (s *GameSession) applyBattleDamage(combatIdx []int, remaining []int) {
	for k, shipIdx := range combatIdx {
		if k >= len(remaining) || shipIdx < 0 || shipIdx >= len(s.Ships) {
			continue
		}
		r := remaining[k]
		if r < 0 {
			continue // 已陣亡,不用記損傷
		}
		max := shipMaxHP(s.Ships[shipIdx])
		d := max - r
		if d < 0 {
			d = 0
		}
		if d > max-ShipDamageFloorHP {
			d = max - ShipDamageFloorHP
		}
		s.Ships[shipIdx].Damage = d
	}
}

// autoRepairInCombat 是自動修復元件的**戰鬥中**修復(手冊 p.82:每回合修復結構損傷的 20%)。
// 回傳修復的點數。
func autoRepairInCombat(damage int) int {
	if damage <= 0 {
		return 0
	}
	r := damage * AutoRepairPercentPerRound / 100
	if r < 1 {
		r = 1 // 手冊沒講捨去到 0 的情形;至少修 1 點,免得小損傷永遠修不完
	}
	if r > damage {
		r = damage
	}
	return r
}
