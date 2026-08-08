package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// achievements.go:手冊講的「**成就**」(Achievements)——全帝國生效的科技效果。
//
// ============ 擋門理由是「沒有成就追蹤系統」,而那句話不成立 ============
//
// 四條手冊規則寫好了、有數字、有出處,四條都**零呼叫端**:
//
//	MoraleVirtualRealityNetworkBonus         全帝國士氣 +20%      (p.97-98)
//	MoralePsionicsGovernmentBonus            特定政府下士氣 +10%  (p.100-101)
//	ProdMicroliteConstructionPerWorkerBonus  每個工業工人 +1 產能 (p.?)
//	PollutionToleranceWithNanoDisassemblers  行星污染容忍值 ×2
//
// `colonyMoralePercent` 的檔頭寫了為什麼:
//
//	Virtual Reality Network(全帝國 +20%,p.97-98):手冊定性為「成就」而非一般建築,
//	不在 gamedata.Buildings 清單、**remake 也無「成就」追蹤系統,無從得知是否擁有**。
//
// **「成就」在 MOO2 就是科技。** 它們不進建造清單是因為研究出來就自動生效,不是因為它們
// 是另一種東西。而「有沒有研究出來」remake 一直查得到——`groundEquipTechOwned` 已經是
// 四個系統共用的判定(生物武器、地面裝備、進階政體、間諜科技加成)。
//
// 這是第 58 項(擋門理由過期三個月)同一個形狀:**擋門理由當時成立,之後沒有人回頭看**。
//
// ============ 誠實留白 ============
//
//   - **只接查得到出處的四條。** 手冊還有別的成就(如恆星轉換器那類),沒有數字的不接。
//   - **AI 對手也吃這些效果**,因為判定只看 `engine.PlayerState`,AI 也有一份。
//     這是對的:成就是科技,AI 研究得出來就該生效(AI 從第 55 項(AI 科技先前靠偷)起才真的會研究)。

// achievementTechs 是本檔會查的成就科技(供測試列舉,順序即檢查順序)。
var achievementTechs = []gamedata.Technology{
	gamedata.TECH_VIRTUAL_REALITY_NETWORK,
	gamedata.TECH_PSIONICS,
	gamedata.TECH_MICROLITE_CONSTRUCTION,
	gamedata.TECH_NANO_DISASSEMBLERS,
}

// hasAchievement 回報這一方有沒有研究出某項成就科技。
//
// 走 `groundEquipTechOwned` 那條既有判定(主題完成 + 未明確抉擇視為擁有 / 已抉擇需選中),
// 與生物武器、地面裝備、間諜科技加成同一套規則——不另立一套「成就專用」的判定。
func hasAchievement(ps engine.PlayerState, tech gamedata.Technology) bool {
	topic, ok := gamedata.OrigTechTopic(tech)
	if !ok {
		return false
	}
	return groundEquipTechOwned(ps, topic, tech)
}

// achievementMoralePercent 回傳成就給的全帝國士氣加成(百分點)。
//
// gov 要傳**進階形式**(見 GameSession.effectiveGovernment):手冊的心靈學那一條列的是
// 「Dictatorship, Imperium, Feudalism, or Confederation」——基本型與進階型分別列名,
// 傳基本型會讓研究出帝國的獨裁玩家查錯格。
func achievementMoralePercent(ps engine.PlayerState, gov gamedata.MoraleGovernmentType) int {
	pct := 0
	if hasAchievement(ps, gamedata.TECH_VIRTUAL_REALITY_NETWORK) {
		pct += gamedata.MoraleVirtualRealityNetworkBonus
	}
	if hasAchievement(ps, gamedata.TECH_PSIONICS) {
		pct += gamedata.MoralePsionicsGovernmentBonus(gov)
	}
	return pct
}

// achievementIndustryPerWorkerBonus 回傳微晶構築給每個工業工人的額外產能。
func achievementIndustryPerWorkerBonus(ps engine.PlayerState) int {
	if hasAchievement(ps, gamedata.TECH_MICROLITE_CONSTRUCTION) {
		return gamedata.ProdMicroliteConstructionPerWorkerBonus
	}
	return 0
}

// hasNanoDisassemblers 回報是否已研究奈米分解者(污染容忍值加倍)。
func hasNanoDisassemblers(ps engine.PlayerState) bool {
	return hasAchievement(ps, gamedata.TECH_NANO_DISASSEMBLERS)
}

// effectiveGovernment 回傳玩家目前**生效**的政體(含研究出來的進階形式)。
//
// `s.Government` 只存基本型。手冊對基本型與進階型給的是**不同的值**(帝國多 +20% 士氣、
// 銀河統一的產出加成是統一的兩倍),所以凡是查政體表的地方都該用這一支。
//
// 與 `assimilationGovernment()` 是同一組編號的兩種型別——原版只有一個 `[player+0x89F]`
// (第 54 項(三個寫入端)),Go 這邊分成三個列舉是歷史,轉型是安全的(`spy_bonus_test.go` 釘住)。
func (s *GameSession) effectiveGovernment() gamedata.MoraleGovernmentType {
	return gamedata.MoraleGovernmentType(s.assimilationGovernment())
}

// syncAchievementColonyFields 把成就效果同步到各殖民地的引擎欄位。
//
// **每回合重算,不是完工時設一次。** 建築可以「蓋好就永遠有」,成就不行:科技會被偷、
// 被交換,而 `groundEquipTechOwned` 的判定還吃「有沒有做過明確抉擇」——那也會變。
// 重算是冪等的,所以順序與呼叫次數都不影響結果,舊存檔讀進來也會自動補齊。
func (s *GameSession) syncAchievementColonyFields() {
	nano := hasNanoDisassemblers(s.Player)
	perWorker := achievementIndustryPerWorkerBonus(s.Player)
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].NanoDisassemblers = nano
		s.PlayerColonies[i].IndustryPerWorkerBonus = perWorker
	}
}
