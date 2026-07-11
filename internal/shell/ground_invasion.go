package shell

import (
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// ground_invasion.go:地面戰入侵流程的「模型 + 流程」殼層(shell)——把 gamedata 已備妥的
// 解算式(ResolveGroundBattle)與加成表(GroundArmorTechBonus 等)接到活的對局狀態:
// 陸戰隊生成(Marine Barracks)→隨艦隊運送(LoadMarines)→抵達敵方殖民地觸發入侵
// (InvadeColony)→勝則佔領(星 Owner 轉移 + 殖民地過戶)。
//
// 本檔只碰資料/流程,不碰 UI(interactive.go)。

// --- 種族→地面戰 force 加成映射(手冊 p.15-16,僅 Bulrathi/Gnolam 有明確數字) ---

const (
	raceIdxBulrathi = 5  // shell.Races 索引:布拉西(Bulrathi)
	raceIdxGnolam   = 11 // shell.Races 索引:諾蘭姆(Gnolams)
)

// groundRaceFor 把玩家選定的種族索引(shell.Races)映射到 gamedata.GroundRace。
// 只有 Bulrathi/Gnolam 手冊給出明確地面戰數字,其餘一律 GroundRaceOther(加成 0)。
// ⚠ AI 對手目前未追蹤種族選擇(AIOpponent 無 RaceIndex 欄位),故本函式只用於玩家側,
// AI 的地面戰 force 一律不套種族加成(見 aiMarineForce)。
func groundRaceFor(raceIdx int) gamedata.GroundRace {
	switch raceIdx {
	case raceIdxBulrathi:
		return gamedata.GroundRaceBulrathi
	case raceIdxGnolam:
		return gamedata.GroundRaceGnolam
	default:
		return gamedata.GroundRaceOther
	}
}

// componentUnlockedFor 是 GameSession.ComponentUnlocked 的無 receiver 版本,供玩家與 AI
// 共用同一套元件解鎖規則(見該方法註解;規則本身完全相同,只是不綁定 s.Player)。
func componentUnlockedFor(ps engine.PlayerState, c Component) bool {
	if c.Tech == gamedata.TOPIC_STARTING_TECH {
		return true
	}
	if ps.CompletedTopics == nil || !ps.CompletedTopics[c.Tech] {
		return false
	}
	if c.UnlockTech == gamedata.TECH_NONE || ps.ExplicitChoice == nil || !ps.ExplicitChoice[c.Tech] {
		return true
	}
	return ps.ChosenTech != nil && ps.ChosenTech[c.Tech] == c.UnlockTech
}

// groundArmorBonusFor 依 ps 已解鎖的裝甲元件中最高階者,回傳 gamedata.GroundArmorTechBonus。
// 沿用既有 ArmorOptions/ComponentUnlocked 解鎖判定(尊重「已明確抉擇」語意),由高到低找
// 第一個 UnlockTech != TECH_NONE 且已解鎖的裝甲元件。氙素裝甲(Xentronium,proxy 元件,
// UnlockTech=0)是里程碑元件、無法對應手冊逐科技加成表,略過改抓次高階(見 session.go
// ArmorOptions 定義的註解)。
func groundArmorBonusFor(ps engine.PlayerState) int {
	for i := len(ArmorOptions) - 1; i >= 0; i-- {
		c := ArmorOptions[i]
		if c.UnlockTech == gamedata.TECH_NONE {
			continue
		}
		if componentUnlockedFor(ps, c) {
			return gamedata.GroundArmorTechBonus(c.UnlockTech)
		}
	}
	return 0
}

// groundEquipTechOwned 判定地面裝備科技(Powered Armor / Anti-Grav Harness / Personal
// Shield)是否已擁有。這三項手冊科技本 remake 的艦艇元件模型未收錄對應項(SpecialOptions
// 無 Powered Armor 等元件),故不透過 ComponentUnlocked/艦艇元件查,改直接查
// CompletedTopics/ExplicitChoice/ChosenTech——判定規則與 componentUnlockedFor 一致(主題
// 完成、未明確抉擇 → 視為解鎖;已明確抉擇 → 需選中該科技),只是省去「元件」這層。
func groundEquipTechOwned(ps engine.PlayerState, topic gamedata.ResearchTopic, tech gamedata.Technology) bool {
	if ps.CompletedTopics == nil || !ps.CompletedTopics[topic] {
		return false
	}
	if ps.ExplicitChoice == nil || !ps.ExplicitChoice[topic] {
		return true
	}
	return ps.ChosenTech != nil && ps.ChosenTech[topic] == tech
}

// hasPoweredArmorFor 回傳 ps 是否已擁有 Powered Armor(TOPIC_ROBOTICS 抉擇),供
// GroundMarineHitsToKill 的 poweredArmor 參數使用(手冊:多 1 hit 才會陣亡)。
func hasPoweredArmorFor(ps engine.PlayerState) bool {
	return groundEquipTechOwned(ps, gamedata.TOPIC_ROBOTICS, gamedata.TECH_POWERED_ARMOR)
}

// hasBattleoidsFor 回傳 ps 是否已研究 Battleoids(手冊 p.81,techtree.go TOPIC_ASTRO_CONSTRUCTION
// 三選一 TECH_BATTLEOIDS)。沿用 groundEquipTechOwned 的判定規則(主題完成 + 未明確抉擇時視為
// 解鎖、已明確抉擇時需選中該科技),與 hasPoweredArmorFor 同款,只是換主題/科技。
func hasBattleoidsFor(ps engine.PlayerState) bool {
	return groundEquipTechOwned(ps, gamedata.TOPIC_ASTRO_CONSTRUCTION, gamedata.TECH_BATTLEOIDS)
}

// tankHitsToKillFor 回傳 ps 這方戰車營單位的陣亡所需 hits。已研究 Battleoids 者手冊 p.81
// 明講「整批換成 Battleoid,固定 3 hits」(GroundBattleoidHitsToKill,不再套用 Heavy-G 修飾,
// 見 ground.go GroundTankHitsToKill 註解);未研究者沿用 GroundTankHitsToKill(highG 未建模,
// 理由同 playerMarineForce 對 Subterranean/High-G 的留白)。
func tankHitsToKillFor(ps engine.PlayerState) int {
	if hasBattleoidsFor(ps) {
		return gamedata.GroundBattleoidHitsToKill
	}
	return gamedata.GroundTankHitsToKill(false)
}

// commandoLeaderTier 掃描 leaders 找出擁有「指揮官」技能標籤(對應 gamedata.SKILL_COMMANDO,
// 手冊 p.135 Commando)的最高技能階(Tier:0 無/1 一般/2 進階)。找不到回傳 0(=無 Commando
// 領袖,無加成)。
//
// ⚠ 近似(2026-07-11,docs/tech/version-1.3-1.5-diff.md #5/#6):手冊描述 Commando 效果綁定
// 「同系統的殖民地領袖」或「同艦隊的艦艇軍官」,remake 沒有「領袖指派到某次入侵/某支艦隊」的
// 模型(shell.Leader/GameSession.Leaders 是帝國全域清單,無指派欄位)。故用「帝國是否擁有
// Commando 技能領袖」當代理條件,不論該領袖的 Ship 欄位、不論其實際位置——只要帝國內存在一名
// Tier>0 的指揮官技能領袖,任何一次入侵都套用其加成。此為誠實標記的近似,非精確的「該次入侵
// 指派了哪位領袖」模擬。
//
// 與 leaderSkillIDByName(session.go)刻意分開:那張表只服務「殖民地經濟被動加成」
// (applyLeaderColonyBonuses,科學家/貿易家/工程師),Commando 屬於地面戰鬥系統,語意/消費端
// 都不同,不應該混進同一張表。
func commandoLeaderTier(leaders []Leader) int {
	best := 0
	for _, l := range leaders {
		if l.Skill != "指揮官" {
			continue
		}
		if l.Tier > best {
			best = l.Tier
		}
	}
	return best
}

// tankForceBonusFor 回傳 tankCount 輛戰車若已升級 Battleoids,對整個地面戰 force 額外貢獻的
// 加成(手冊 p.81:「a ground combat rating 10 higher than a tank」,GroundBattleoidCombatBonus)。
// 只在 tankCount>0 時套用——0 輛戰車的一方不該白拿這個加成(該加成描述的是「戰車營升級
// Battleoid 後」的相對戰力,無戰車時無意義)。
func tankForceBonusFor(ps engine.PlayerState, tankCount int) int {
	if tankCount > 0 && hasBattleoidsFor(ps) {
		return gamedata.GroundBattleoidCombatBonus
	}
	return 0
}

// groundEquipmentBonusFor 加總 ps 已擁有的地面裝備科技加成。三項(Powered Armor/Anti-Grav
// Harness/Personal Shield)是各自獨立的裝備(非同一裝甲槽的互斥升級,不同於 ArmorOptions),
// 手冊逐條描述皆為各自加成,故加總而非取最高。
func groundEquipmentBonusFor(ps engine.PlayerState) int {
	bonus := 0
	if hasPoweredArmorFor(ps) {
		bonus += gamedata.GroundEquipmentTechBonus(gamedata.TECH_POWERED_ARMOR)
	}
	if groundEquipTechOwned(ps, gamedata.TOPIC_GRAVITIC_FIELDS, gamedata.TECH_ANTIGRAV_HARNESS) {
		bonus += gamedata.GroundEquipmentTechBonus(gamedata.TECH_ANTIGRAV_HARNESS)
	}
	if groundEquipTechOwned(ps, gamedata.TOPIC_ELECTROMAGNETIC_REFRACTION, gamedata.TECH_PERSONAL_SHIELD) {
		bonus += gamedata.GroundEquipmentTechBonus(gamedata.TECH_PERSONAL_SHIELD)
	}
	return bonus
}

// playerMarineForce 回傳玩家陸戰隊單位的 force 加成(裝甲科技 + 裝備科技 + 種族加成,
// Gnolam 另套 Low-G 10% 懲罰)。Subterranean 加成、High-G hits-to-kill 未套用:本 remake
// 未建模「特殊能力(Special Abilities)」選取(見 ApplyCustomRaceBonuses 註解),13 個標準
// 種族也沒有一個具備 Subterranean/High-G,故無從套用,誠實留白而非臆測。
func (s *GameSession) playerMarineForce() int {
	force := groundArmorBonusFor(s.Player) + groundEquipmentBonusFor(s.Player) + gamedata.GroundRaceCombatBonus(groundRaceFor(s.RaceIndex))
	if s.RaceIndex == raceIdxGnolam {
		force = gamedata.GroundApplyLowGPenalty(force)
	}
	return force
}

// hasPoweredArmor 回傳玩家是否已擁有 Powered Armor。
func (s *GameSession) hasPoweredArmor() bool {
	return hasPoweredArmorFor(s.Player)
}

// aiMarineForce 回傳某 AI 對手陸戰隊單位的 force 加成。⚠ 簡化:AIOpponent 目前未追蹤種族
// 選擇,故只計裝甲/裝備科技加成,不套種族/Low-G/Subterranean(這些本來就只有極少數種族有
// 明確數字,詳見 playerMarineForce)。
func aiMarineForce(a AIOpponent) int {
	return groundArmorBonusFor(a.Player) + groundEquipmentBonusFor(a.Player)
}

// --- Marine Barracks 生成(EndTurn 每回合補充,見 GameSession.EndTurn 呼叫 advanceMarines) ---

// marineBarracksBuildingName 是 gamedata.Buildings 對應「Marine Barracks」的中文譯名
// (session.go applyBuildingEffect/homeworldBuildings 已用同一字串當 key)。
const marineBarracksBuildingName = "海軍陸戰隊營"

// armorBarracksBuildingName 是 gamedata.Buildings 對應「Armor Barracks」的中文譯名(見
// gamedata/buildings.go NameZH:"裝甲營房";先前 session.go colonyMoralePercent 已用同一字面
// 字串當士氣判定的一部分,這裡補一個具名常數取代各處寫死字串,對稱 marineBarracksBuildingName)。
const armorBarracksBuildingName = "裝甲營房"

// advanceMarines 讓每個已建成 Marine Barracks 的玩家殖民地依手冊公式
// (gamedata.GroundMarineBarracksUnits)補充陸戰隊駐軍池,有上限(GroundMarineBarracksCap)。
// 只會成長,不會因為公式重算而倒退(用 max 寫回,而非直接覆蓋)——已消耗掉(見 LoadMarines)
// 的駐軍不會被本函式無中生有補回超過「理論上限」的量,只在殖民地公式支持的上限內回補。
//
// Warlord 特性(手冊 p.27,barracks 容量加倍)本 remake 未建模(無對應種族/特殊能力追蹤),
// 一律傳 false。
func (s *GameSession) advanceMarines() {
	if s.PlayerColonyMarines == nil {
		s.PlayerColonyMarines = make([]int, len(s.PlayerColonies))
	}
	if s.MarineBarracksAge == nil {
		s.MarineBarracksAge = make([]int, len(s.PlayerColonies))
	}
	for len(s.PlayerColonyMarines) < len(s.PlayerColonies) {
		s.PlayerColonyMarines = append(s.PlayerColonyMarines, 0)
	}
	for len(s.MarineBarracksAge) < len(s.PlayerColonies) {
		s.MarineBarracksAge = append(s.MarineBarracksAge, 0)
	}
	for i := range s.PlayerColonies {
		if i >= len(s.ColonyBuildings) || s.ColonyBuildings[i] == nil || !s.ColonyBuildings[i][marineBarracksBuildingName] {
			continue
		}
		age := s.MarineBarracksAge[i]
		c := s.PlayerColonies[i]
		n := gamedata.GroundMarineBarracksUnits(age, c.Population, c.PopMax, false)
		if n > s.PlayerColonyMarines[i] {
			s.PlayerColonyMarines[i] = n
		}
		s.MarineBarracksAge[i]++
	}
}

// advanceArmor 讓每個已建成 Armor Barracks 的玩家殖民地依手冊公式
// (gamedata.GroundArmorBarracksUnits)補充戰車營駐軍池,有上限(GroundArmorBarracksCap)。
// 邏輯與 advanceMarines 完全對稱(見該函式註解),只是換裝甲營房建築名/戰車駐軍池欄位。
//
// Warlord 特性同樣未建模(理由同 advanceMarines),一律傳 false。
func (s *GameSession) advanceArmor() {
	if s.PlayerColonyTanks == nil {
		s.PlayerColonyTanks = make([]int, len(s.PlayerColonies))
	}
	if s.ArmorBarracksAge == nil {
		s.ArmorBarracksAge = make([]int, len(s.PlayerColonies))
	}
	for len(s.PlayerColonyTanks) < len(s.PlayerColonies) {
		s.PlayerColonyTanks = append(s.PlayerColonyTanks, 0)
	}
	for len(s.ArmorBarracksAge) < len(s.PlayerColonies) {
		s.ArmorBarracksAge = append(s.ArmorBarracksAge, 0)
	}
	for i := range s.PlayerColonies {
		if i >= len(s.ColonyBuildings) || s.ColonyBuildings[i] == nil || !s.ColonyBuildings[i][armorBarracksBuildingName] {
			continue
		}
		age := s.ArmorBarracksAge[i]
		c := s.PlayerColonies[i]
		n := gamedata.GroundArmorBarracksUnits(age, c.Population, c.PopMax, false)
		if n > s.PlayerColonyTanks[i] {
			s.PlayerColonyTanks[i] = n
		}
		s.ArmorBarracksAge[i]++
	}
}

// --- 運送(陸戰隊隨艦隊出征) ---

// MarineTransportCapacity 估算玩家艦隊目前可載運的陸戰隊上限。
//
// ⚠ 簡化待精修:本 remake 尚無獨立的「運輸艦」船體類別(ShipCost/shipStrength 的 Class
// switch 沒有「運輸艦」這個 case),故無法像手冊那樣「每艘 Transport Ship 恰配 4 個 Marine
// 單位」精算。以「艦隊現有艦數 × gamedata.GroundTransportShipMarineCapacity(手冊每艘運輸艦
// 4 個單位的數字)」做為近似運力上限——不區分殖民船/偵察艦/戰鬥艦,所有艦一律視為「可搭載
// 陸戰隊艙位」。待補上真正的運輸艦船體類型後,應改為只計數該類型艦。
func (s *GameSession) MarineTransportCapacity() int {
	return len(s.Ships) * gamedata.GroundTransportShipMarineCapacity
}

// LoadMarines 把玩家殖民地 colonyIdx 的 Marine Barracks 駐軍池部隊,載上隨艦隊出征的
// FleetMarines,上限受 MarineTransportCapacity 節制(已載運的量不會被擠出)。
// 回傳實際載運數(0 表示無可載運空間或該殖民地無駐軍)。
func (s *GameSession) LoadMarines(colonyIdx int) int {
	if colonyIdx < 0 || colonyIdx >= len(s.PlayerColonyMarines) {
		return 0
	}
	room := s.MarineTransportCapacity() - s.FleetMarines
	if room <= 0 {
		return 0
	}
	n := s.PlayerColonyMarines[colonyIdx]
	if n > room {
		n = room
	}
	if n <= 0 {
		return 0
	}
	s.PlayerColonyMarines[colonyIdx] -= n
	s.FleetMarines += n
	return n
}

// LoadTanks 把玩家殖民地 colonyIdx 的 Armor Barracks 駐軍池戰車營,載上隨艦隊出征的
// FleetTanks。⚠ 簡化:remake 沒有獨立的「戰車運輸艙位」資料(手冊只明講 Transport Ship /
// Troop Pods 是針對 Marine),故與 FleetMarines 共用同一個 MarineTransportCapacity() 運力池
// (room 扣掉兩者已載運的量)——這是誠實的簡化,不是手冊原文規則,見 MarineTransportCapacity
// 註解。回傳實際載運數(0 表示無可載運空間或該殖民地無駐軍)。
func (s *GameSession) LoadTanks(colonyIdx int) int {
	if colonyIdx < 0 || colonyIdx >= len(s.PlayerColonyTanks) {
		return 0
	}
	room := s.MarineTransportCapacity() - s.FleetMarines - s.FleetTanks
	if room <= 0 {
		return 0
	}
	n := s.PlayerColonyTanks[colonyIdx]
	if n > room {
		n = room
	}
	if n <= 0 {
		return 0
	}
	s.PlayerColonyTanks[colonyIdx] -= n
	s.FleetTanks += n
	return n
}

// --- 入侵觸發 + 解算 ---

// findAIColonyByStar 尋找 starIdx 對應到哪個 AI 對手的哪個殖民地(依 AIOpponent.ColonyStars
// 對映)。找不到(ok=false)表示該星是「已佔領但未建模殖民地」的擴張版圖(見 aiExpand 與
// AIOpponent.ColonyStars 註解),目前不可入侵。
func (s *GameSession) findAIColonyByStar(starIdx int) (aiIdx, colonyIdx int, ok bool) {
	for ai := range s.AIPlayers {
		for ci, st := range s.AIPlayers[ai].ColonyStars {
			if st == starIdx {
				return ai, ci, true
			}
		}
	}
	return 0, 0, false
}

// PlayerOwnedStars 回傳玩家目前擁有的星數(即時依 Stars.Owner==1 計數,不另存計數器,
// 避免與 InvadeColony/aiExpand 等會改動 Owner 的流程手動同步出岔)。
func (s *GameSession) PlayerOwnedStars() int {
	n := 0
	for _, st := range s.Stars {
		if st.Owner == 1 {
			n++
		}
	}
	return n
}

// GroundInvasionResult 是一次入侵嘗試的結果(供 UI/測試檢視)。
type GroundInvasionResult struct {
	Ok                      bool   // 是否成功發動了一場入侵解算(false = 前置條件不足,未開打)
	Reason                  string // Ok=false 時的原因(供 UI 提示;Ok=true 時為空字串)
	AttackerWon             bool   // Ok=true 時才有意義
	AttackerSurvived        int    // 攻方存活總數(陸戰隊+戰車營,拆解見下兩欄)
	AttackerMarinesSurvived int    // 攻方存活的陸戰隊數(AttackerSurvived 的子集,見 InvadeColony 拆解說明)
	AttackerTanksSurvived   int    // 攻方存活的戰車營數(同上)
	DefenderSurvived        int
	Rounds                  int
	StarCaptured            bool // 攻方勝且完成佔領星 + 殖民地過戶
}

// InvadeColony 嘗試對 starIdx 這顆星發動地面入侵。前置條件:
//  1. 玩家艦隊已抵達該星(FleetAtStar==starIdx 且 FleetETA==0,航行中不能發動)。
//  2. 該星是敵方(Owner==2)且有「已建模」的殖民地(findAIColonyByStar 找得到)。
//  3. 玩家艦隊已載運地面部隊(FleetMarines>0 或 FleetTanks>0,由 LoadMarines/LoadTanks 載運)。
//
// 任一條件不足回傳 Ok=false + Reason,不消耗任何狀態、不呼叫 rng。
//
// 解算組雙方 gamedata.GroundForce:
//
//   - 攻方:FleetMarines 個陸戰隊單位 + FleetTanks 個戰車營單位混編。force 統一套用
//     playerMarineForce()(裝甲/裝備/種族加成,對整支部隊一視同仁——本 remake 的
//     GroundForce.Force 本來就是「side 級」單一加成,不分兵種,見 ground_battle.go 設計),
//     若持有 Battleoids 再疊加 tankForceBonusFor 的相對加成,再疊加 Commando 領袖加成(見
//     commandoLeaderTier + gamedata.GroundCommandoAttackerForceBonus,2026-07-11,#5/#6)。
//     hits-to-kill 陸戰隊/戰車營分開算(GroundMarineHitsToKill / tankHitsToKillFor)。
//
//     ⚠ 單位排序(無把握的接法,已選定但列出讓 L.CY 定案):合併後的 Units 陣列「陸戰隊在前、
//     戰車營在後」——這不是敘事上的「誰當前鋒」選擇(手冊未提供地面戰隊形資訊),而是技術上
//     唯一能在戰後把 res.AttackerSurvived(單一總數)準確拆回「陸戰隊存活數 / 戰車營存活數」
//     的排法:ResolveGroundBattle 的規則是「最前面存活單位先受創」,即單位嚴格按索引順序陣亡
//     (index 0 全滅後才輪到 index 1),故存活者必是原始順序的「後段」;把戰車營放在後段,
//     戰後只需 tanksSurvived=min(total存活,戰車營原始數量) 即可還原分兵種存活數,不需更動
//     gamedata 層的 GroundUnit/GroundForce 結構(該結構本身無兵種標記欄位)。若未來要精確
//     模擬「戰車在前掩護陸戰隊」的戰術隊形,需要先幫 GroundUnit 加兵種欄位,超出本輪死碼
//     串接範圍。
//
//   - 守方:兵力簡化為 gamedata.GroundMarineBarracksUnits(s.Turn, colony.Population,
//     colony.PopMax, false)——AI 未追蹤各殖民地 Marine Barracks 是否已建成/已運作幾回合
//     (AI 無對應 ColonyBuildings 追蹤機制),以「已運作 s.Turn 回合」做近似(AI 母星開局
//     即有 Marine Barracks,見 homeworldBuildings);force=aiMarineForce()。守方戰車營
//     TODO 未接:AI 開局 homeworldBuildings() 本就沒有裝甲營房(只有海軍陸戰隊營+星基),
//     且 AIOpponent 完全沒有 ColonyBuildings 追蹤機制可供判斷「AI 是否已建成裝甲營房」,
//     沒有資料可誠實推導守方戰車數,故不臆測補上——這與 marine 側的近似不同,marine 側
//     至少有「開局必有 Marine Barracks」這個已知事實撐腰,armor 側沒有對應事實。
//
// rng 依「回合數 + 星索引」種子化(同 ResolveBattle/ResolveGroundBattle 呼叫慣例),同一回合
// 對同一顆星重複輸入必得到相同結果,可重現。
//
// 攻方勝:星 Owner 轉 1;把該 AI 殖民地整筆過戶為玩家殖民地(PlayerColonies 新增一筆,
// Builds/ColonyBuildings/PlayerColonyMarines/MarineBarracksAge 同步補齊長度)、從
// AIOpponent.Colonies/ColonyStars 移除、雙方持有星數更新(AI.OwnedStars--;玩家由
// PlayerOwnedStars() 即時算,Owner 已轉 1 故自動反映)。過戶殖民地人口簡化為「地面戰守方
// 存活戰鬥單位數」(手冊 p.162-164 只有敘述性描述,無精確的「入侵後保留多少平民人口」公式,
// 至少保留 1 人口,標簡化待精修)。
//
// 攻方敗(含平手皆歸守方,見 ResolveGroundBattle):FleetMarines/FleetTanks 回寫為攻方存活數
// (戰損),Owner 不變、殖民地不轉移。
func (s *GameSession) InvadeColony(starIdx int) GroundInvasionResult {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return GroundInvasionResult{Reason: "無效的星索引"}
	}
	if s.FleetAtStar != starIdx || s.FleetETA != 0 {
		return GroundInvasionResult{Reason: "艦隊尚未抵達該星"}
	}
	star := &s.Stars[starIdx]
	if star.Owner != 2 {
		return GroundInvasionResult{Reason: "該星不是敵方殖民地"}
	}
	if s.FleetMarines <= 0 && s.FleetTanks <= 0 {
		return GroundInvasionResult{Reason: "艦隊未載運地面部隊"}
	}
	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		return GroundInvasionResult{Reason: "該星無可入侵的殖民地模型(簡化限制,見 AIOpponent.ColonyStars)"}
	}
	aiPlayer := &s.AIPlayers[aiIdx]
	colony := aiPlayer.Colonies[colonyIdx]

	tankCount := s.FleetTanks
	atkForce := s.playerMarineForce() + tankForceBonusFor(s.Player, tankCount)
	// Commando 領袖加成(#5/#6,2026-07-11,見 commandoLeaderTier 註解的近似說明):兩版攻方
	// 倍率相同(非差異項),不需要查 RuleProfile。
	atkForce += gamedata.GroundCommandoAttackerForceBonus(commandoLeaderTier(s.Leaders))
	marineHits := gamedata.GroundMarineHitsToKill(false, s.hasPoweredArmor())
	tankHits := tankHitsToKillFor(s.Player)
	// 合併陸戰隊+戰車營單位:Force 只借用 marineUnits/tankUnits 建構出來的 Units,side 級的
	// atkForce 已在上面算好,故建構單位時 force 參數傳 0(NewGroundForce 的 force 只是塞進
	// GroundForce.Force 欄位,這裡改在合併後的 atk struct 上設一次即可,避免混淆)。
	marineUnits := gamedata.NewGroundForce(s.FleetMarines, marineHits, 0, false).Units
	tankUnits := gamedata.NewGroundForce(tankCount, tankHits, 0, false).Units
	atkUnits := append(append([]gamedata.GroundUnit{}, marineUnits...), tankUnits...)
	atk := gamedata.GroundForce{Units: atkUnits, Force: atkForce, Defending: false}

	defCount := gamedata.GroundMarineBarracksUnits(s.Turn, colony.Population, colony.PopMax, false)
	defForce := aiMarineForce(*aiPlayer)
	// TODO 守方 Commando 加成(#5,ruleprofile.go RuleProfile.DefenderCommandoBonus)掛鉤點:
	// gamedata.GroundCommandoDefenderForceBonus(tier, s.RuleProfile.DefenderCommandoBonus) 已
	// 實作+測試,但 AIOpponent 沒有 Leaders 欄位(AI 完全沒有領袖資料模型,不同於陸戰隊/戰車那種
	// 「至少有開局必有 Marine Barracks」的可推導事實撐腰),沒有資料可誠實推導 AI 是否有 Commando
	// 領袖,故不臆測補上(不寫死 tier=0 假裝有接線——這裡就是誠實留白,見 RuleProfile 欄位註解)。
	defHits := gamedata.GroundMarineHitsToKill(false, hasPoweredArmorFor(aiPlayer.Player))
	def := gamedata.NewGroundForce(defCount, defHits, defForce, true)

	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + int64(starIdx)*97 + 555))
	res := gamedata.ResolveGroundBattle(atk, def, rng)

	// 拆回陸戰隊/戰車營各自存活數:見上方函式註解「單位排序」——戰車營排在合併陣列尾端,
	// 故戰後存活者必優先含括戰車營(死亡按原始順序發生),tanksSurvived 最多不超過原始戰車營
	// 數量,剩下的存活數才輪到陸戰隊。
	tanksSurvived := res.AttackerSurvived
	if tanksSurvived > tankCount {
		tanksSurvived = tankCount
	}
	marinesSurvived := res.AttackerSurvived - tanksSurvived

	out := GroundInvasionResult{
		Ok: true, AttackerWon: res.AttackerWon,
		AttackerSurvived: res.AttackerSurvived, DefenderSurvived: res.DefenderSurvived,
		AttackerMarinesSurvived: marinesSurvived, AttackerTanksSurvived: tanksSurvived,
		Rounds: res.Rounds,
	}
	s.FleetMarines = marinesSurvived
	s.FleetTanks = tanksSurvived

	if res.AttackerWon {
		star.Owner = 1

		captured := colony
		captured.Population = res.DefenderSurvived // 簡化近似,見函式註解
		if captured.Population < 1 {
			captured.Population = 1
		}
		s.PlayerColonies = append(s.PlayerColonies, captured)
		s.Builds = append(s.Builds, ColonyBuild{})
		for len(s.ColonyBuildings) < len(s.PlayerColonies) {
			s.ColonyBuildings = append(s.ColonyBuildings, nil)
		}
		for len(s.PlayerColonyMarines) < len(s.PlayerColonies) {
			s.PlayerColonyMarines = append(s.PlayerColonyMarines, 0)
		}
		for len(s.MarineBarracksAge) < len(s.PlayerColonies) {
			s.MarineBarracksAge = append(s.MarineBarracksAge, 0)
		}
		for len(s.PlayerColonyTanks) < len(s.PlayerColonies) {
			s.PlayerColonyTanks = append(s.PlayerColonyTanks, 0)
		}
		for len(s.ArmorBarracksAge) < len(s.PlayerColonies) {
			s.ArmorBarracksAge = append(s.ArmorBarracksAge, 0)
		}
		// popAccum(見 advancePopulation):該函式對 `i >= len(s.popAccum)` 是 break 而非
		// continue,若這裡不補齊長度,過戶的殖民地(以及任何排在它之後的殖民地)人口成長會被
		// 永久跳過,不是明顯的 crash,只是靜默停止成長——同步補齊,避免這個潛在缺口。
		for len(s.popAccum) < len(s.PlayerColonies) {
			s.popAccum = append(s.popAccum, 0)
		}
		// PlayerColonyStars(見 GameSession 欄位註解、colonization.go):過戶的殖民地所在星就是
		// starIdx 本身,同步補上,維持 len(PlayerColonyStars)==len(PlayerColonies) 不變量。
		for len(s.PlayerColonyStars) < len(s.PlayerColonies)-1 {
			s.PlayerColonyStars = append(s.PlayerColonyStars, -1)
		}
		s.PlayerColonyStars = append(s.PlayerColonyStars, starIdx)

		aiPlayer.Colonies = append(aiPlayer.Colonies[:colonyIdx], aiPlayer.Colonies[colonyIdx+1:]...)
		aiPlayer.ColonyStars = append(aiPlayer.ColonyStars[:colonyIdx], aiPlayer.ColonyStars[colonyIdx+1:]...)
		if aiPlayer.OwnedStars > 0 {
			aiPlayer.OwnedStars--
		}
		out.StarCaptured = true
		s.advanceConquestVictory() // 若這是該 AI 對手的最後一個殖民地,立即偵測「殲滅所有對手」勝利(見 council.go),不用等下個 EndTurn
	}
	return out
}
