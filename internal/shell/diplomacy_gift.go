package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// diplomacyCashGiftDefault 是外交畫面目前提供的固定現金餽贈額度。
//
// 原版的 Diplomacy_Offer_Money @0x1D565 允許玩家輸入指定金額；現有
// remake 的外交畫面尚未有可重用的數字輸入框，因此先提供一個可玩的
// 10 BC 垂直切片。OfferCashGift 本身接受任意正整數，之後接入輸入框時
// 不需要改動國庫轉移與測試邊界。
const diplomacyCashGiftDefault = 10

// diplomacyCashGiftRelationDelta 是現金餽贈切片的關係改善量。
//
// 已證實：原版 Get_Gift_Response @0x539D9 的 cash gift 分支(a3=5)
// 會對外交評估分數套用 -50；該分數與 remake 的 -40..40 Relation
// 並非同一欄位，不能直接把 50 寫成關係分數。這裡以固定 +5 作為
// 「改善關係」的保守正規化，並在 OfferCashGift 的呼叫端套用已接好的
// 魅力外交倍率；數值對映屬強推論，完整接受門檻仍待原版 v9/v20 的
// 未解資料流與 runtime oracle。
func diplomacyCashGiftRelationDelta() int {
	return 5
}

// OfferCashGift 執行玩家對指定 AI 的一次性現金餽贈。
//
// 原版直接交易的方向已證實：An @0x1DEF8 的外交 case 13 對玩家國庫
// 減去 word_19A192，並對對方國庫加上同額。本函式保留該資料流，並以
// 玩家目前 BC 作為失敗即不變的邊界；關係接受／拒絕的完整原版判定仍
// 未映射，這個最小切片視為玩家送出後成功。
func (s *GameSession) OfferCashGift(enemy string, amount int) string {
	s.recordPlayerCommand(PlayerCommand{Name: CmdOfferCashGift, Args: []int{amount}, Text: enemy})
	ai := s.aiByDisplayName(enemy)
	if ai == nil {
		return ""
	}
	if amount <= 0 {
		return enemy + ":餽贈金額必須大於 0 BC。"
	}
	if s.Player.BC < amount {
		return fmt.Sprintf("%s:國庫只有 %d BC，無法餽贈 %d BC。", enemy, s.Player.BC, amount)
	}

	s.Player.BC -= amount
	ai.Player.BC += amount

	delta := s.diplomacyRelationGain(diplomacyCashGiftRelationDelta())
	if delta < 1 {
		delta = 1
	}
	ai.adjustRelation(delta)
	return fmt.Sprintf("%s:我們接受你贈送的 %d BC。(關係改善)", enemy, amount)
}

// OfferTechnologyGift 把玩家已知、對方未知的一項科技直接贈送給 AI。
// 科技的主題／明確選擇狀態沿用間諜偷竊同一個 applyTechTheft 入口，避免
// 「研究、偷竊、贈送」三條路徑留下不同的解鎖語意。
func (s *GameSession) OfferTechnologyGift(enemy string, topic gamedata.ResearchTopic, tech gamedata.Technology) string {
	s.recordPlayerCommand(PlayerCommand{Name: CmdOfferTechGift, Args: []int{int(topic), int(tech)}, Text: enemy})
	ai := s.aiByDisplayName(enemy)
	if ai == nil {
		return ""
	}
	if !psKnowsTech(s.Player, topic, tech) {
		return enemy + ":我方尚未掌握這項科技。"
	}
	if psKnowsTech(ai.Player, topic, tech) {
		return enemy + ":對方已經知道這項科技。"
	}
	applyTechTheft(&ai.Player, spyStealOption{Topic: topic, Tech: tech})
	delta := s.diplomacyRelationGain(8)
	if delta < 1 {
		delta = 1
	}
	ai.adjustRelation(delta)
	return fmt.Sprintf("%s:我們接受你贈送的科技「%s」。(關係改善)", enemy, gamedata.TechnologyName(tech))
}

// OfferStarGift 把玩家的一座非母星殖民地完整移交給指定 AI。這是一次性
// 外交餽贈，不把人口標成征服人口；平行殖民地陣列沿用 removePlayerColony
// 的維護契約，AI 端同步 ColonyStars／ColonyPlanets／ColonyBuildings。
func (s *GameSession) OfferStarGift(enemy string, starIdx int) string {
	s.recordPlayerCommand(PlayerCommand{Name: CmdOfferStarGift, Args: []int{starIdx}, Text: enemy})
	ai := s.aiByDisplayName(enemy)
	if ai == nil {
		return ""
	}
	if starIdx <= 0 || starIdx >= len(s.Stars) {
		return enemy + ":母星或無效星系不可餽贈。"
	}
	if len(s.PlayerColonies) <= 1 {
		return enemy + ":至少要保留一座殖民地。"
	}
	colonyIdx := -1
	for i := range s.PlayerColonies {
		if s.PlayerColonyStarIndex(i) == starIdx {
			colonyIdx = i
			break
		}
	}
	if colonyIdx < 0 || s.Stars[starIdx].Owner != 1 {
		return enemy + ":這顆星不是我方可餽贈的殖民地。"
	}
	for _, st := range ai.ColonyStars {
		if st == starIdx {
			return enemy + ":對方已經擁有這顆星的殖民地。"
		}
	}
	captured := s.PlayerColonies[colonyIdx]
	planet := s.ColonyPlanetIndex(colonyIdx)
	buildings := map[string]bool(nil)
	if colonyIdx < len(s.ColonyBuildings) {
		buildings = cloneBuildings(s.ColonyBuildings[colonyIdx])
	}
	marines, tanks, marineAge, armorAge := 0, 0, 0, 0
	if colonyIdx < len(s.PlayerColonyMarines) {
		marines = s.PlayerColonyMarines[colonyIdx]
	}
	if colonyIdx < len(s.PlayerColonyTanks) {
		tanks = s.PlayerColonyTanks[colonyIdx]
	}
	if colonyIdx < len(s.MarineBarracksAge) {
		marineAge = s.MarineBarracksAge[colonyIdx]
	}
	if colonyIdx < len(s.ArmorBarracksAge) {
		armorAge = s.ArmorBarracksAge[colonyIdx]
	}
	s.removePlayerColony(colonyIdx)
	ensureAIGroundForceSlots(ai)
	ai.Colonies = append(ai.Colonies, captured)
	ai.ColonyStars = append(ai.ColonyStars, starIdx)
	ai.ColonyPlanets = append(ai.ColonyPlanets, planet)
	ai.ColonyBuildings = append(ai.ColonyBuildings, buildings)
	ai.ColonyMarines = append(ai.ColonyMarines, marines)
	ai.ColonyTanks = append(ai.ColonyTanks, tanks)
	ai.MarineBarracksAge = append(ai.MarineBarracksAge, marineAge)
	ai.ArmorBarracksAge = append(ai.ArmorBarracksAge, armorAge)
	s.Stars[starIdx].Owner = 2
	ai.OwnedStars++
	ai.adjustRelation(s.diplomacyRelationGain(12))
	return fmt.Sprintf("%s:我方已餽贈殖民地「%s」。(關係改善)", enemy, s.starName(starIdx))
}
