package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// assimilation.go:**征服人口的同化**接進回合迴圈。
//
// 速率表(逐政體 2–20 回合、排斥種族 ×2、異族管理中心固定 2)全在
// `gamedata/assimilation.go`,這一檔負責:攻下殖民地時記下「這裡全是外族」、
// 每回合推進、以及把 remake 的政體索引對到那張表。
//
// ============ 為什麼這是「征服打法」的門檻 ============
//
// 民主 4 回合 vs 統一 20 回合——差五倍。一個統一政體的帝國打下一顆 8 人口的星,
// 要 160 回合才全部同化完;民主只要 32。這不是細節,是原版把「和平科技流」與
// 「征服流」分開的規則手段。異族管理中心(維護費 1 BC)把它壓成固定 2 回合,
// 所以那棟建築對統一政體等於十倍速。
//
// ============ 誠實留白 ============
//
//   - 未同化人口的兩個後果**都已接上**:20% 士氣懲罰走 `colonyMoralePercent`(第 98 項)、
//     叛亂走 `rebellion.go`(第 105 項)。所以同化不再只是一個會走完的計時器,
//     它是「征服打法要付的利息」——同化完之前每一回合都在擲叛亂骰。
//   - AI 攻下玩家殖民地那條路徑沒有同化(AI 沒有殖民地狀態的完整模型)。

// assimilationGovernment 把 remake 的政體對到同化速率表的政體。
//
// 進階政體(邦聯/帝國/聯邦/銀河統一)在原版是研究出來的升級版,所以這裡看對應科技:
// 有那個科技就走進階那一格(同化更快)。
func (s *GameSession) assimilationGovernment() gamedata.AssimilationGovernment {
	base := gamedata.AssimDictatorship
	var advTech gamedata.Technology
	switch s.Government {
	case gamedata.MoraleGovFeudalism:
		base, advTech = gamedata.AssimFeudal, gamedata.TECH_CONFEDERATION
	case gamedata.MoraleGovDictatorship:
		base, advTech = gamedata.AssimDictatorship, gamedata.TECH_IMPERIUM
	case gamedata.MoraleGovDemocracy:
		base, advTech = gamedata.AssimDemocracy, gamedata.TECH_FEDERATION
	case gamedata.MoraleGovUnification:
		base, advTech = gamedata.AssimUnification, gamedata.TECH_GALACTIC_UNIFICATION
	}
	// 進階政體的判定沿用 hasPoweredArmorFor 那組既有規則(主題完成 + 抉擇命中),
	// 不另立一套——四個進階政體都在 TOPIC_ADVANCED_GOVERNMENTS 這個四選一主題底下。
	if groundEquipTechOwned(s.Player, gamedata.TOPIC_ADVANCED_GOVERNMENTS, advTech) {
		return gamedata.AssimilationAdvancedForm(base)
	}
	return base
}

// AssimilationTurnsFor 回傳某個殖民地同化一單位人口要幾回合(供 UI 顯示)。
func (s *GameSession) AssimilationTurnsFor(colonyIdx int) int {
	hasCenter := colonyIdx >= 0 && colonyIdx < len(s.ColonyBuildings) &&
		s.ColonyBuildings[colonyIdx][alienManagementCenterName]
	return gamedata.AssimilationTurns(s.assimilationGovernment(), hasCenter, s.RaceRepulsive, s.RaceCharismatic)
}

// alienManagementCenterName 是建築表裡的中文名(gamedata.Buildings 的 NameZH)。
const alienManagementCenterName = "異族管理中心"

// markColonyConquered 把剛攻下的殖民地標成「全部是未同化的外族人口」,並記下**舊主是誰**
// ——叛亂成功時殖民地要還給它(手冊 p.165「the colony reverts back」,見 rebellion.go)。
func markColonyConquered(c *engine.ColonyState, fromAI int) {
	c.UnassimilatedPop = c.Population
	c.AssimilationProgress = 0
	c.ConqueredFrom, c.ConqueredFromKnown = fromAI, fromAI >= 0
}

// advanceAssimilation 每回合推進所有殖民地的同化。
//
// 一個殖民地累積滿「這個政體的回合數」就同化掉一單位人口,餘數留著繼續累
// ——不是「每 N 回合歸零重來」,那會在政體改變時吃掉玩家已經累積的進度。
func (s *GameSession) advanceAssimilation() {
	for i := range s.PlayerColonies {
		c := &s.PlayerColonies[i]
		if c.UnassimilatedPop <= 0 {
			c.AssimilationProgress = 0 // 沒有外族人口時不留殘值
			continue
		}
		need := s.AssimilationTurnsFor(i)
		c.AssimilationProgress++
		for c.AssimilationProgress >= need && c.UnassimilatedPop > 0 {
			c.AssimilationProgress -= need
			c.UnassimilatedPop--
		}
		// 人口被炸掉/餓死時未同化數不該超過總人口。
		if c.UnassimilatedPop > c.Population {
			c.UnassimilatedPop = c.Population
		}
		// 多種族士氣懲罰是「有沒有未同化人口」的函數,所以同化完最後一單位的那一刻
		// 懲罰就該消失——不重算的話玩家會一直被扣到下次蓋建築為止。
		s.recalcColonyMorale(i)
	}
}

// AssimilationRemainingTurns 回傳殖民地 i 把**剩下的**未同化人口全部同化完還要幾回合。
//
// ⚠ 這支的存在理由寫在 `gamedata.AssimilationProgressNeeded` 的檔頭上:
// 「UI 要顯示『這個殖民地還要幾回合才完全同化』——一個只在背景默默跑的機制對玩家等於不存在」。
// 那支函式抽出來之後**一直沒有呼叫端**,直到第 120 項。
//
// 已累積的進度要扣掉,否則玩家每一回合看到的數字都一樣,像是完全沒有在推進。
// ok=false 表示這個殖民地沒有未同化人口(不必顯示這一行)。
func (s *GameSession) AssimilationRemainingTurns(i int) (turns int, ok bool) {
	if i < 0 || i >= len(s.PlayerColonies) {
		return 0, false
	}
	c := s.PlayerColonies[i]
	if c.UnassimilatedPop <= 0 {
		return 0, false
	}
	need := gamedata.AssimilationProgressNeeded(c.UnassimilatedPop, s.AssimilationTurnsFor(i))
	need -= c.AssimilationProgress
	if need < 1 {
		need = 1 // 已經滿了但還沒結算——顯示「還有 1 回合」比顯示 0 誠實
	}
	return need, true
}
