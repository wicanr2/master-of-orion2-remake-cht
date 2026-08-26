package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// ground_invasion.go:地面戰入侵流程的「模型 + 流程」殼層(shell)——把 gamedata 已備妥的
// 解算式(2026-08-07 起用原版的 ResolveGroundCombatOrig)與加成表(GroundArmorTechBonus 等)
// 接到活的對局狀態:
// 陸戰隊生成(Marine Barracks)→隨艦隊運送(LoadMarines)→抵達敵方殖民地觸發入侵
// (InvadeColony)→勝則佔領(星 Owner 轉移 + 殖民地過戶)。
//
// 本檔只碰資料/流程,不碰 UI(interactive.go)。

// componentUnlockedFor 是 GameSession.ComponentUnlocked 的無 receiver 版本,供玩家與 AI
// 共用同一套元件解鎖規則(見該方法註解;規則本身完全相同,只是不綁定 s.Player)。
func componentUnlockedFor(ps engine.PlayerState, c Component) bool {
	if c.Tech == gamedata.TOPIC_STARTING_TECH {
		return true
	}
	if c.UnlockTech == gamedata.TECH_NONE {
		return ps.CompletedTopics != nil && ps.CompletedTopics[c.Tech]
	}
	if playerStateKnowsTech(ps, c.Tech, c.UnlockTech) {
		return true
	}
	return false
}

// bestUnlockedWeaponValue 掃描 profile 版本規則下的武器清單(BuildWeaponOptions),在指定
// WeaponKind(beam/missile)裡挑出 ps 已解鎖(componentUnlockedFor)、Value 最大的一款,回傳
// 其 Value 與 gamedata.WeaponSpaceByName 查到的佔格。供 orbital_bombardment.go
// retaliationAttackers 用來推導軌道防禦基地「配備目前最好的武器」的戰力(#14)。
//
// 一款都沒解鎖時(ok=false):回退到該類「最基礎款」——beam 用雷射(Value=4,space=10)、
// missile 用核飛彈(Value=6,space=10),回退值就是該武器本身表列的 Value/Space(動態查表,
// 非另外硬編一組數字),與「玩家剛解鎖雷射/核飛彈當下」的數字完全相同,故 fallback 不會讓
// 早期 AI 的軌道防禦戰力算出不合理的 0——手冊對行星防禦的定性描述是「配備目前最好的武器」,
// 在「什麼都還沒解鎖」這個邊界情況下,合理的預設就是最基礎的雷射/核飛彈,不是 0 火力。ok 只
// 作誠實旗標供測試/未來 UI 判斷「這是不是真解鎖出來的值」,不影響 value/space 的計算結果。
//
// ok=false 只代表目前沒有已解鎖的同類武器；AI 研究會由 advanceAIResearch
// 在正常回合中推進 CompletedTopics，所以武器與軌道防禦戰力會隨對局進展提升。
func bestUnlockedWeaponValue(ps engine.PlayerState, profile gamedata.RuleProfile, kind WeaponKind) (value, space int, ok bool) {
	options := BuildWeaponOptions(profile)
	bestValue, bestSpace := -1, 0
	for _, c := range options {
		if c.Name == "無武裝" {
			continue
		}
		if weaponKindByName(c.Name) != kind {
			continue
		}
		if !componentUnlockedFor(ps, c) {
			continue
		}
		if c.Value > bestValue {
			bestValue = c.Value
			bestSpace = gamedata.WeaponSpaceByName[c.Name]
		}
	}
	if bestValue >= 0 {
		return bestValue, bestSpace, true
	}
	switch kind {
	case WeaponKindMissile:
		return 6, gamedata.WeaponSpaceByName["核飛彈"], false
	default:
		return 4, gamedata.WeaponSpaceByName["雷射"], false
	}
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

// groundRifleBonusFor 回傳玩家**最高階已知步槍**的地面戰加成。
//
// ⚠ 這一整條通道 remake 先前完全沒有(gap report 第 35 項(三張查表)):原版有
// `Player_Best_Rifle_` @ 0xDC416,走訪 `word_14A88` 那張表由高到低找第一個已知的。
// 上限差 30 點——後期科技全開時,先前的地面部隊比原版弱整整 30。
//
// 「取最高階」不是加總:原版那支函式一找到就回傳。
func groundRifleBonusFor(ps engine.PlayerState) int {
	ladder := gamedata.GroundRifleLadder()
	for i := len(ladder) - 1; i >= 0; i-- {
		tech := ladder[i]
		topic, ok := groundRifleTopic(tech)
		if !ok {
			continue
		}
		if groundEquipTechOwned(ps, topic, tech) {
			return gamedata.GroundRifleTechBonus(tech)
		}
	}
	return 0
}

// groundRifleTopic 把步槍科技對到它所屬的研究主題(techtree.go 的 Choices)。
//
// 脈衝步槍在**起始科技**(field 0)裡,所以一律已知——它的加成本來就是 0,
// 但保留在表上是為了讓「取最高階」的走訪有個底。
func groundRifleTopic(tech gamedata.Technology) (gamedata.ResearchTopic, bool) {
	switch tech {
	case gamedata.TECH_PULSE_RIFLE:
		return gamedata.TOPIC_STARTING_TECH, true
	case gamedata.TECH_LASER_RIFLE:
		return gamedata.TOPIC_PHYSICS, true
	case gamedata.TECH_FUSION_RIFLE:
		return gamedata.TOPIC_FUSION_PHYSICS, true
	case gamedata.TECH_PHASOR_RIFLE:
		return gamedata.TOPIC_MULTIPHASED_PHYSICS, true
	case gamedata.TECH_PLASMA_RIFLE:
		return gamedata.TOPIC_PLASMA_PHYSICS, true
	}
	return 0, false
}

// groundEquipTechOwned 判定地面裝備科技(Powered Armor / Anti-Grav Harness / Personal
// Shield)是否已擁有。這三項手冊科技本 remake 的艦艇元件模型未收錄對應項(SpecialOptions
// 無 Powered Armor 等元件),故不透過 ComponentUnlocked/艦艇元件查,改直接查
// CompletedTopics/ExplicitChoice/ChosenTech——判定規則與 componentUnlockedFor 一致(主題
// 完成、未明確抉擇 → 視為解鎖;已明確抉擇 → 需選中該科技),只是省去「元件」這層。
func groundEquipTechOwned(ps engine.PlayerState, topic gamedata.ResearchTopic, tech gamedata.Technology) bool {
	return playerStateKnowsTech(ps, topic, tech)
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
// 見 ground.go GroundTankHitsToKill 註解);未研究者沿用 GroundTankHitsToKill。
//
// highG 由呼叫端傳(玩家查 `s.raceHasTrait(TRAIT_HIGH_G)`;AI 一律 false——AIOpponent
// 沒有種族欄位,那一整層還不存在,見 aiMarineForce)。
func tankHitsToKillFor(ps engine.PlayerState, highG bool) int {
	if hasBattleoidsFor(ps) {
		return gamedata.GroundBattleoidHitsToKill
	}
	return gamedata.GroundTankHitsToKill(highG)
}

// commandoLeaderTier 掃描 leaders 找出擁有 Commando 技能(gamedata.SKILL_COMMANDO,
// 手冊 p.135)的最高技能階(0 無/1 一般/2 進階)。找不到回傳 0(=無 Commando 領袖,無加成)。
//
// ⚠ 2026-08-08(第 45 項(領袖技能))從比對中文標籤 `l.Skill == "指揮官"` 改成比對技能 id。
// 那個比法錯得很安靜:HERODATA 的**艦艇軍官一律被標成「指揮官」**(那是類別通稱,
// 不是 Commando 的譯名),於是每一位雇來的艦艇軍官都拿到了地面戰加成;
// 反過來在英文模式下標籤是 "Commander",一個都對不上,加成整個消失。
//
// ⚠ 近似(2026-07-11,docs/tech/version-1.3-1.5-diff.md #5/#6):手冊描述 Commando 效果綁定
// 「同系統的殖民地領袖」或「同艦隊的艦艇軍官」,remake 沒有「領袖指派到某次入侵/某支艦隊」的
// 模型(shell.Leader/GameSession.Leaders 是帝國全域清單,無指派欄位)。故用「帝國是否擁有
// Commando 技能領袖」當代理條件,不論該領袖的 Ship 欄位、不論其實際位置——只要帝國內存在一名
// Tier>0 的指揮官技能領袖,任何一次入侵都套用其加成。此為誠實標記的近似,非精確的「該次入侵
// 指派了哪位領袖」模擬。
//
// 與 applyLeaderColonyBonuses 分開的理由不變:那邊算的是殖民地經濟欄位,Commando 是地面戰
// force 加成,語意與單位都不同,不該混進同一段合成邏輯。分開的是**消費端**,識別鍵則是共用的
// 技能 id(leaderSkillTier)。
func commandoLeaderTier(leaders []Leader) int {
	best := 0
	for _, l := range leaders {
		if t := leaderSkillTier(l, int(gamedata.SKILL_COMMANDO)); t > best {
			best = t
		}
	}
	return best
}

// groundMarineOnlyBonusFor 是**只有陸戰隊**拿得到的裝備加成(原版 `[out+3]`,
// 只被類型 1 的 case 讀走)。目前只有動力裝甲。
func groundMarineOnlyBonusFor(ps engine.PlayerState) int {
	if hasPoweredArmorFor(ps) {
		return gamedata.GroundEquipmentTechBonus(gamedata.TECH_POWERED_ARMOR)
	}
	return 0
}

// groundTankOnlyBonusFor 是**只有裝甲/戰車營**拿得到的加成(原版 `[out+1]`,
// 只被類型 0 的 case 讀走)。目前只有 Battleoids。
//
// ⚠ 2026-08-07 由 `tankForceBonusFor`(加進整側的 force)改成分兵種。
// 先前戰車營一有 Battleoids,連陸戰隊都跟著 +10;現在只有戰車營拿。
// 「tankCount > 0 才給」那個守門也不需要了——它加在戰車那一格上,沒有戰車時那格本來就是空的。
func groundTankOnlyBonusFor(ps engine.PlayerState) int {
	if hasBattleoidsFor(ps) {
		return gamedata.GroundBattleoidCombatBonus
	}
	return 0
}

// groundEquipmentBonusFor 加總 ps 的**全兵種共用**地面裝備加成。
//
// ⚠ **2026-08-07 訂正:動力裝甲不在這裡了。** 原版把三項裝備寫進加成塊的不同欄位,
// 而那些欄位被不同的部隊類型讀走(gap report 第 35 項(三張查表)):
//
//	TECH_ANTIGRAV_HARNESS  → [out+0]        所有類型共用   ← 留在這個函式
//	TECH_PERSONAL_SHIELD   → [out+7]/[out+8] 所有類型共用   ← 留在這個函式(走「取最高階」通道)
//	TECH_POWERED_ARMOR     → [out+3]/[out+4] **只有陸戰隊**  ← 搬到 groundMarineOnlyBonusFor
//
// remake 先前把動力裝甲加給整支部隊,等於戰車營也白拿那 10 點。
func groundEquipmentBonusFor(ps engine.PlayerState) int {
	bonus := 0
	if groundEquipTechOwned(ps, gamedata.TOPIC_GRAVITIC_FIELDS, gamedata.TECH_ANTIGRAV_HARNESS) {
		bonus += gamedata.GroundEquipmentTechBonus(gamedata.TECH_ANTIGRAV_HARNESS)
	}
	if groundEquipTechOwned(ps, gamedata.TOPIC_ELECTROMAGNETIC_REFRACTION, gamedata.TECH_PERSONAL_SHIELD) {
		bonus += gamedata.GroundEquipmentTechBonus(gamedata.TECH_PERSONAL_SHIELD)
	}
	return bonus
}

// playerMarineForce 回傳玩家陸戰隊單位的 force 加成(裝甲科技 + 裝備科技 + 種族特性)。
//
// ⚠ **2026-08-08 兩處訂正(第 65 項(種族特性31格),種族特性表接線)。**
//
// ① 先前的檔頭寫著:
//
//	Subterranean 加成、High-G hits-to-kill 未套用:… 13 個標準種族也沒有一個具備
//	Subterranean/High-G,故無從套用,誠實留白而非臆測。
//
// **那句話可證為假。** `RACESTUF.LBX` asset 7 攤開來,薩克拉有 Subterranean、
// 布拉西有 High-G。當時查不到是因為 remake 沒有特性模型,不是因為原版沒有這兩族。
//
// ② **諾蘭姆的低重力懲罰先前扣了兩次**:`GroundRaceCombatBonus(GroundRaceGnolam)` 回 −10,
// 而下一行的 `GroundApplyLowGPenalty` 又扣一次 −10。反組譯只寫一次
// (`mov byte ptr [ecx+0Dh], 0F6h`),而諾蘭姆的 `TRAIT_GROUND_COMBAT` 本來就是 0
// ——它的懲罰完全來自 `TRAIT_LOW_G`。改由特性表驅動之後這個重複自然消失。
//
// defending 為真時另加 Subterranean 的守方加成(手冊 p.24 + 反組譯的呼叫端旗標,
// 見 gamedata.GroundSubterraneanBonus)。
func (s *GameSession) playerMarineForce() int {
	return s.marineForceFor(false)
}

// playerDefendingMarineForce 是守衛自家殖民地時的版本(Subterranean 只在這時生效)。
func (s *GameSession) playerDefendingMarineForce() int {
	return s.marineForceFor(true)
}

func (s *GameSession) marineForceFor(defending bool) int {
	force := groundArmorBonusFor(s.Player) + groundRifleBonusFor(s.Player) +
		groundEquipmentBonusFor(s.Player) + s.RaceGroundBonus
	if s.raceHasTrait(gamedata.TRAIT_LOW_G) {
		force = gamedata.GroundApplyLowGPenalty(force)
	}
	if defending && s.raceHasTrait(gamedata.TRAIT_SUBTERRANEAN) {
		force += gamedata.GroundSubterraneanBonus(true)
	}
	return force
}

// raceHasTrait 回報玩家的種族有沒有某項布林特性(見 race_boolean_traits.go)。
func (s *GameSession) raceHasTrait(t gamedata.RaceTrait) bool {
	if s.raceOrigIdx() < 0 {
		if t < gamedata.TRAIT_LOW_G || t > gamedata.TRAIT_POOR_HOMEWORLD {
			return false
		}
		return s.CustomRaceTraits&(uint32(1)<<uint(t)) != 0
	}
	return gamedata.OrigRaceHasTrait(s.raceOrigIdx(), t)
}

// hasPoweredArmor 回傳玩家是否已擁有 Powered Armor。
func (s *GameSession) hasPoweredArmor() bool {
	return hasPoweredArmorFor(s.Player)
}

// aiMarineForce 回傳 AI 自己的陸戰隊 force：科技、種族 Ground Combat 與 Low-G
// 都從該 AIOpponent 取得，不再沿用玩家數值。
func aiMarineForce(a AIOpponent) int {
	force := groundArmorBonusFor(a.Player) + groundRifleBonusFor(a.Player) +
		groundEquipmentBonusFor(a.Player)
	if idx := aiRaceIndex(a); idx >= 0 && idx < len(Races) {
		force += Races[idx].GroundCombatBonus
	}
	if aiRaceHasTrait(a, gamedata.TRAIT_LOW_G) {
		force = gamedata.GroundApplyLowGPenalty(force)
	}
	return force
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
		n := gamedata.GroundMarineBarracksUnits(age, c.Population, c.PopMax, s.RaceWarlord())
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
		n := gamedata.GroundArmorBarracksUnits(age, c.Population, c.PopMax, s.RaceWarlord())
		if n > s.PlayerColonyTanks[i] {
			s.PlayerColonyTanks[i] = n
		}
		s.ArmorBarracksAge[i]++
	}
}

// ensureAIGroundForceSlots 建立／修補 AI 殖民地地面部隊的四個平行陣列。
//
// 新欄位加入前的存檔沒有這些資料；對缺少的 slot 只以目前建築和 age=0 的原版
// 兵營初始公式建立一次，之後由 advanceAIGroundForces 以 max 回寫。這個「只對
// 缺欄位初始化」很重要：若每次入侵前都重算，玩家打掉的 AI 守軍會在下一次檢查
// 時復活，反而比原版更不忠實。
func ensureAIGroundForceSlots(a *AIOpponent) {
	n := len(a.Colonies)
	for len(a.ColonyMarines) < n {
		i := len(a.ColonyMarines)
		age := 0
		if i < len(a.MarineBarracksAge) {
			age = a.MarineBarracksAge[i]
		}
		n := 0
		if i < len(a.Colonies) && i < len(a.ColonyBuildings) && a.ColonyBuildings[i] != nil && a.ColonyBuildings[i][marineBarracksBuildingName] {
			c := a.Colonies[i]
			n = gamedata.GroundMarineBarracksUnits(age, c.Population, c.PopMax, aiRaceHasTrait(*a, gamedata.TRAIT_WARLORD))
		}
		a.ColonyMarines = append(a.ColonyMarines, n)
	}
	for len(a.ColonyTanks) < n {
		i := len(a.ColonyTanks)
		age := 0
		if i < len(a.ArmorBarracksAge) {
			age = a.ArmorBarracksAge[i]
		}
		n := 0
		if i < len(a.Colonies) && i < len(a.ColonyBuildings) && a.ColonyBuildings[i] != nil && a.ColonyBuildings[i][armorBarracksBuildingName] {
			c := a.Colonies[i]
			n = gamedata.GroundArmorBarracksUnits(age, c.Population, c.PopMax, aiRaceHasTrait(*a, gamedata.TRAIT_WARLORD))
		}
		a.ColonyTanks = append(a.ColonyTanks, n)
	}
	for len(a.MarineBarracksAge) < n {
		a.MarineBarracksAge = append(a.MarineBarracksAge, 0)
	}
	for len(a.ArmorBarracksAge) < n {
		a.ArmorBarracksAge = append(a.ArmorBarracksAge, 0)
	}
	if len(a.ColonyMarines) > n {
		a.ColonyMarines = a.ColonyMarines[:n]
	}
	if len(a.ColonyTanks) > n {
		a.ColonyTanks = a.ColonyTanks[:n]
	}
	if len(a.MarineBarracksAge) > n {
		a.MarineBarracksAge = a.MarineBarracksAge[:n]
	}
	if len(a.ArmorBarracksAge) > n {
		a.ArmorBarracksAge = a.ArmorBarracksAge[:n]
	}
}

// advanceAIGroundForces 讓 AI 的每座兵營依原版初始值／每五回合一單位／人口上限
// 公式補充駐軍。它在 AI 經濟結算之後呼叫，讓當回合人口回寫先影響容量，再進入
// 下一個玩家可觀察到的地面戰狀態。
func advanceAIGroundForces(a *AIOpponent) {
	ensureAIGroundForceSlots(a)
	warlord := aiRaceHasTrait(*a, gamedata.TRAIT_WARLORD)
	for i, c := range a.Colonies {
		if i < len(a.ColonyBuildings) && a.ColonyBuildings[i] != nil && a.ColonyBuildings[i][marineBarracksBuildingName] {
			n := gamedata.GroundMarineBarracksUnits(a.MarineBarracksAge[i], c.Population, c.PopMax, warlord)
			if n > a.ColonyMarines[i] {
				a.ColonyMarines[i] = n
			}
			a.MarineBarracksAge[i]++
		}
		if i < len(a.ColonyBuildings) && a.ColonyBuildings[i] != nil && a.ColonyBuildings[i][armorBarracksBuildingName] {
			n := gamedata.GroundArmorBarracksUnits(a.ArmorBarracksAge[i], c.Population, c.PopMax, warlord)
			if n > a.ColonyTanks[i] {
				a.ColonyTanks[i] = n
			}
			a.ArmorBarracksAge[i]++
		}
	}
}

// removeAIGroundForceSlot 與 AI 殖民地的其他平行陣列一起移除一筆駐軍資料。
func removeAIGroundForceSlot(a *AIOpponent, i int) {
	if i < 0 {
		return
	}
	cut := func(values *[]int) {
		if i < len(*values) {
			*values = append((*values)[:i], (*values)[i+1:]...)
		}
	}
	cut(&a.ColonyMarines)
	cut(&a.ColonyTanks)
	cut(&a.MarineBarracksAge)
	cut(&a.ArmorBarracksAge)
}

// --- 運送(陸戰隊隨艦隊出征) ---

// MarineTransportCapacity 回傳玩家艦隊目前可載運的陸戰隊上限,**逐艦依艦體等級**累加
// (手冊 p.121 的 Marines 欄,見 gamedata.ShipHullMarines)。
//
// ============ 原本的擋門理由問錯了問題 ============
//
// 這裡原本是 `len(Ships) * 4`,理由寫著:
//
//	⚠ 簡化待精修:本 remake 尚無獨立的「運輸艦」船體類別……故無法像手冊那樣
//	「每艘 Transport Ship 恰配 4 個 Marine 單位」精算。
//
// 那個理由假設運力來自**一種專門的運輸艦**。但手冊 p.121 的艦艇設計表**每一種艦體等級
// 都有一欄 Marines**(5/8/12/20/30/50)——運力是艦體本身的屬性,不是某個艦種的特權。
// 而 remake 一直有艦體等級(`shipClassFromName`,那支函式本身就是為了查同一張表寫的)。
//
// 所以先前的效果是:**一支偵察艦隊與一支末日之星艦隊的登陸能力完全相同**,
// 而手冊上兩者差 10 倍。
//
// ⚠ 誠實留白:偵察艦不在手冊那六個設計艦級裡(`shipClassFromName` 回 ok=false),
// 沿用它既有的「以護衛艦近似」慣例——那是既有決定,不是這一項新引入的。
func (s *GameSession) MarineTransportCapacity() int {
	total := 0
	for _, sh := range s.Fleet().Ships {
		class, _ := shipClassFromName(sh.Class)
		n := gamedata.ShipHullMarines(class)
		// 部隊艙(第 68 項(元件盤點+飛彈防禦)):手冊 p.79「doubling the number of Marines on board a ship」。
		// `gamedata.GroundTroopPodsMultiplier` 從寫出來就零生產端,因為元件不存在。
		// 乘在**逐艦**而不是總數上——只有裝了的那艘翻倍,這才是手冊講的「on board a ship」。
		if shipHasTroopPods(sh) {
			n *= gamedata.GroundTroopPodsMultiplier
		}
		total += n
	}
	return total
}

// LoadMarines 把玩家殖民地 colonyIdx 的 Marine Barracks 駐軍池部隊,載上隨艦隊出征的
// FleetMarines,上限受 MarineTransportCapacity 節制(已載運的量不會被擠出)。
// 回傳實際載運數(0 表示無可載運空間或該殖民地無駐軍)。
func (s *GameSession) LoadMarines(colonyIdx int) int {
	s.recordPlayerCommand(PlayerCommand{Name: CmdLoadMarines, Args: []int{colonyIdx}})
	if colonyIdx < 0 || colonyIdx >= len(s.PlayerColonyMarines) {
		return 0
	}
	room := s.MarineTransportCapacity() - s.Fleet().Marines
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
	s.Fleet().Marines += n
	return n
}

// LoadTanks 把玩家殖民地 colonyIdx 的 Armor Barracks 駐軍池戰車營,載上隨艦隊出征的
// FleetTanks。⚠ 簡化:remake 沒有獨立的「戰車運輸艙位」資料(手冊只明講 Transport Ship /
// Troop Pods 是針對 Marine),故與 FleetMarines 共用同一個 MarineTransportCapacity() 運力池
// (room 扣掉兩者已載運的量)——這是誠實的簡化,不是手冊原文規則,見 MarineTransportCapacity
// 註解。回傳實際載運數(0 表示無可載運空間或該殖民地無駐軍)。
func (s *GameSession) LoadTanks(colonyIdx int) int {
	s.recordPlayerCommand(PlayerCommand{Name: CmdLoadTanks, Args: []int{colonyIdx}})
	if colonyIdx < 0 || colonyIdx >= len(s.PlayerColonyTanks) {
		return 0
	}
	room := s.MarineTransportCapacity() - s.Fleet().Marines - s.Fleet().Tanks
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
	s.Fleet().Tanks += n
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
	AttackerMarinesStart    int    // 開打前攻方陸戰隊數(供地面戰畫面顯示戰前/戰後對比)
	AttackerTanksStart      int    // 開打前攻方戰車營數(同上)
	DefenderStart           int    // 開打前守方兵力(同上)
	ColonyName              string // 被入侵的星名(畫面標題用;engine.ColonyState 本身沒有名稱欄位)
	AttackerSurvived        int    // 攻方存活總數(陸戰隊+戰車營,拆解見下兩欄)
	AttackerMarinesSurvived int    // 攻方存活的陸戰隊數(AttackerSurvived 的子集,見 InvadeColony 拆解說明)
	AttackerTanksSurvived   int    // 攻方存活的戰車營數(同上)
	DefenderSurvived        int
	Rounds                  int
	StarCaptured            bool // 攻方勝且完成佔領星 + 殖民地過戶
	MindControlled          bool // 心靈感應直接接管殖民地，人口已全部同化
}

// MindControlColony 是心靈感應種族的替代入侵行動。手冊明列：至少一艘
// 巡洋艦以上艦艇在軌道上即可心靈控制殖民地，取得殖民地並立即同化全部人口；
// 不消耗陸戰隊／戰車，也不進入地面戰。
func (s *GameSession) MindControlColony(starIdx int) GroundInvasionResult {
	s.recordPlayerCommand(PlayerCommand{Name: CmdMindControl, Args: []int{starIdx}})
	if !s.RaceTelepathic() {
		return GroundInvasionResult{Reason: "只有心靈感應種族可以心靈控制殖民地"}
	}
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return GroundInvasionResult{Reason: "無效的星索引"}
	}
	if s.Fleet().AtStar != starIdx || s.Fleet().ETA != 0 {
		return GroundInvasionResult{Reason: "艦隊尚未抵達該星"}
	}
	if s.Stars[starIdx].Owner != 2 {
		return GroundInvasionResult{Reason: "該星不是敵方殖民地"}
	}
	cruiserOrLarger := false
	for _, ship := range s.Fleet().Ships {
		class, ok := shipClassFromName(ship.Class)
		if ok && class >= gamedata.SHIP_CRUISER {
			cruiserOrLarger = true
			break
		}
	}
	if !cruiserOrLarger {
		return GroundInvasionResult{Reason: "需要至少一艘巡洋艦以上艦艇才能心靈控制"}
	}
	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		return GroundInvasionResult{Reason: "該星無可心靈控制的殖民地模型"}
	}
	aiPlayer := &s.AIPlayers[aiIdx]
	ensureAIGroundForceSlots(aiPlayer)
	captured := aiPlayer.Colonies[colonyIdx]
	population := captured.Population
	if population < 1 {
		population = 1
	}
	captured.Population = population
	captured.UnassimilatedPop = 0
	captured.AssimilationProgress = 0
	captured.UnassimilatedFarmers, captured.UnassimilatedWorkers, captured.UnassimilatedScientists = 0, 0, 0
	engine.ClearPopulationGroupPrisoners(&captured)
	planet := -1
	if colonyIdx < len(aiPlayer.ColonyPlanets) {
		planet = aiPlayer.ColonyPlanets[colonyIdx]
	}
	if planet < 0 {
		planet = s.PlanetAt(starIdx)
	}
	transferredBuildings := s.prepareCapturedAIColony(aiIdx, colonyIdx, planet)
	s.PlayerColonies = append(s.PlayerColonies, captured)
	s.ensureColonyLeaderSlots()
	s.Builds = append(s.Builds, ColonyBuild{})
	s.ColonyBuildings = append(s.ColonyBuildings, transferredBuildings)
	s.PlayerColonyMarines = append(s.PlayerColonyMarines, 0)
	s.MarineBarracksAge = append(s.MarineBarracksAge, 0)
	s.PlayerColonyTanks = append(s.PlayerColonyTanks, 0)
	s.ArmorBarracksAge = append(s.ArmorBarracksAge, 0)
	s.popAccum = append(s.popAccum, 0)
	for len(s.PlayerColonyStars) < len(s.PlayerColonies)-1 {
		s.PlayerColonyStars = append(s.PlayerColonyStars, -1)
	}
	s.PlayerColonyStars = append(s.PlayerColonyStars, starIdx)
	for len(s.PlayerColonyPlanets) < len(s.PlayerColonies)-1 {
		s.PlayerColonyPlanets = append(s.PlayerColonyPlanets, -1)
	}
	s.PlayerColonyPlanets = append(s.PlayerColonyPlanets, planet)

	aiPlayer.Colonies = append(aiPlayer.Colonies[:colonyIdx], aiPlayer.Colonies[colonyIdx+1:]...)
	aiPlayer.ColonyStars = append(aiPlayer.ColonyStars[:colonyIdx], aiPlayer.ColonyStars[colonyIdx+1:]...)
	if colonyIdx < len(aiPlayer.ColonyPlanets) {
		aiPlayer.ColonyPlanets = append(aiPlayer.ColonyPlanets[:colonyIdx], aiPlayer.ColonyPlanets[colonyIdx+1:]...)
	}
	if colonyIdx < len(aiPlayer.ColonyBuildings) {
		aiPlayer.ColonyBuildings = append(aiPlayer.ColonyBuildings[:colonyIdx], aiPlayer.ColonyBuildings[colonyIdx+1:]...)
	}
	removeAIGroundForceSlot(aiPlayer, colonyIdx)
	s.recalcAIColonyMorale(aiIdx)
	s.recalcAllColonyMorale()
	stillEnemy := false
	for _, st := range aiPlayer.ColonyStars {
		if st == starIdx {
			stillEnemy = true
			break
		}
	}
	if !stillEnemy {
		s.Stars[starIdx].Owner = 1
		if aiPlayer.OwnedStars > 0 {
			aiPlayer.OwnedStars--
		}
	}
	s.advanceConquestVictory()
	return GroundInvasionResult{
		Ok: true, AttackerWon: true, StarCaptured: !stillEnemy, MindControlled: true,
		DefenderStart: population, DefenderSurvived: population, ColonyName: s.starName(starIdx),
	}
}

// InvadeColony 嘗試對 starIdx 這顆星發動地面入侵。前置條件:
//  1. 玩家艦隊已抵達該星(FleetAtStar==starIdx 且 FleetETA==0,航行中不能發動)。
//  2. 該星是敵方(Owner==2)且有「已建模」的殖民地(findAIColonyByStar 找得到)。
//  3. 玩家艦隊已載運地面部隊(FleetMarines>0 或 FleetTanks>0,由 LoadMarines/LoadTanks 載運)。
//
// 任一條件不足回傳 Ok=false + Reason,不消耗任何狀態、不呼叫 rng。
//
// 解算組雙方 gamedata.GroundSide(2026-08-07 換成**原版的資料結構**,見下)。
//
//   - 攻方:陸戰隊 = 類型 0、戰車營 = 類型 1。原版的一方就是「四種部隊,一種打完換下一種」
//     (`Ground_Combat_Round_` @ 0xEC4FE,見 gamedata/ground_battle_orig.go),
//     與先前「合併陣列、陸戰隊在前」的意圖相同,只是換成原版的形狀。
//     force 套用 playerMarineForce()(裝甲/裝備/種族加成),持有 Battleoids 再疊
//     tankForceBonusFor,再疊 Commando 領袖加成。
//     hits-to-kill 陸戰隊/戰車營分開算(GroundMarineHitsToKill / tankHitsToKillFor)。
//
//     ⚠ 先前這裡有一整段在解釋「為什麼把戰車營排在陣列尾端」——那個限制**已經消失**:
//     原版的結構逐類型記數量,戰後直接讀 `Count[類型]` 就是各兵種的真實存活數,
//     不必再用 `min(總存活, 戰車原始數)` 推算。那段說明連同它的 TODO 一併移除。
//
//     ⚠ 仍在的留白:原版**每種部隊各有一個攻擊力**(`[side + type*2 + 2]`),
//     那張表還沒追出來,所以兩種目前都填同一個 atkForce。填同值 = 維持現行數字,
//     而且把差異留在一個看得見的地方(見呼叫處的註解)。
//
//   - 守方:使用 AIOpponent 每座殖民地的駐軍池與兵營 age。AI 舊存檔缺少新欄位時,
//     ensureAIGroundForceSlots 只在第一次以現有建築的 age=0 初值補上；新局與每回合
//     則由 advanceAIGroundForces 依原版兵營公式補充。因此 Marine Barracks 與 Armor
//     Barracks 都是實際資料，不再以 s.Turn 猜測，也不會把戰損重算復原。
//
// rng 依「回合數 + 星索引」種子化(同 ResolveBattle/ResolveGroundBattle 呼叫慣例),同一回合
// 對同一顆星重複輸入必得到相同結果,可重現。
//
// 攻方勝:星 Owner 轉 1;把該 AI 殖民地整筆過戶為玩家殖民地(PlayerColonies 新增一筆,
// Builds/ColonyBuildings/PlayerColonyMarines/MarineBarracksAge 同步補齊長度；其他建築隨
// 殖民地過戶，Capitol 依原版 `sub_ECBF7` 單獨移除並觸發舊擁有者重指派；再從
// AIOpponent/Colony* 與 AI 駐軍平行陣列移除、
// 雙方持有星數更新(AI.OwnedStars--;玩家由 PlayerOwnedStars() 即時算,Owner 已轉 1 故自動反映)。
// 過戶殖民地保留原始 Population；原版靜態呼叫鏈 `Change_Colony_Ownership_ @ 0xED260`
// → `Resolve_Invasion_Troops_ @ 0xECECA` 寫回的是入侵部隊欄位，不是人口欄位。守方戰鬥
// 單位存活數只作結果摘要／守方駐軍回寫，不能當成俘虜人口。
//
// 攻方敗(含無法佔領的平手／同時歸零情況皆歸守方,見 ResolveGroundCombatOrig):FleetMarines/FleetTanks 回寫為攻方存活數
// (戰損),Owner 不變、殖民地不轉移。
func (s *GameSession) InvadeColony(starIdx int) GroundInvasionResult {
	s.recordPlayerCommand(PlayerCommand{Name: CmdInvadeColony, Args: []int{starIdx}})
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return GroundInvasionResult{Reason: "無效的星索引"}
	}
	if s.Fleet().AtStar != starIdx || s.Fleet().ETA != 0 {
		return GroundInvasionResult{Reason: "艦隊尚未抵達該星"}
	}
	star := &s.Stars[starIdx]
	if star.Owner != 2 {
		return GroundInvasionResult{Reason: "該星不是敵方殖民地"}
	}
	if s.Fleet().Marines <= 0 && s.Fleet().Tanks <= 0 {
		return GroundInvasionResult{Reason: "艦隊未載運地面部隊"}
	}
	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		return GroundInvasionResult{Reason: "該星無可入侵的殖民地模型(簡化限制,見 AIOpponent.ColonyStars)"}
	}
	aiPlayer := &s.AIPlayers[aiIdx]
	ensureAIGroundForceSlots(aiPlayer)
	colony := aiPlayer.Colonies[colonyIdx]
	// 被打下來的是**哪一顆行星**——AI 殖民地自 2026-08-07 起記得住(AIOpponent.ColonyPlanets)。
	// 要在移除那筆之前先抄下來。舊存檔沒記(−1)時退回該星的代表行星。
	capturedPlanet := -1
	if colonyIdx < len(aiPlayer.ColonyPlanets) {
		capturedPlanet = aiPlayer.ColonyPlanets[colonyIdx]
	}
	if capturedPlanet < 0 {
		capturedPlanet = s.PlanetAt(starIdx)
	}

	tankCount := s.Fleet().Tanks
	// atkForce 是**全兵種共用**的基礎;分兵種的加成在下面各自疊(見 groundMarineOnlyBonusFor /
	// groundTankOnlyBonusFor)。先前這裡把 Battleoids 的 +10 加進共用基礎,陸戰隊也跟著白拿。
	atkForce := s.playerMarineForce()
	// Commando 領袖加成(#5/#6,2026-07-11,見 commandoLeaderTier 註解的近似說明):兩版攻方
	// 倍率相同(非差異項),不需要查 RuleProfile。
	atkForce += gamedata.GroundCommandoAttackerForceBonus(commandoLeaderTier(s.Leaders))
	marineHits := gamedata.GroundMarineHitsToKill(s.raceHasTrait(gamedata.TRAIT_HIGH_G), s.hasPoweredArmor())
	tankHits := tankHitsToKillFor(s.Player, s.raceHasTrait(gamedata.TRAIT_HIGH_G))
	// 合併陸戰隊+戰車營單位:Force 只借用 marineUnits/tankUnits 建構出來的 Units,side 級的
	// atkForce 已在上面算好,故建構單位時 force 參數傳 0(NewGroundForce 的 force 只是塞進
	// GroundForce.Force 欄位,這裡改在合併後的 atk struct 上設一次即可,避免混淆)。
	// 攻方分兩種部隊:陸戰隊 = 類型 0、戰車營 = 類型 1(原版是「一種打完換下一種」,
	// 與先前「戰車營排在合併陣列尾端」的意圖相同,只是換成原版的資料結構)。
	//
	// 逐類型的攻擊力/耐受值 = side 級基礎 + 該類型的調整量。
	// 調整量是原版四個 case 的**純立即數**(裝甲 +10 攻擊 +1 耐受、陸戰隊基準、民兵 −10),
	// 見 gamedata.GroundTypeStrengthDelta / GroundTypeHitsDelta。
	//
	// ⚠ 原版那兩個「科技加成」欄位(加成塊 +1/+3、+2/+4)還沒對出意義,不含在調整量裡
	// ——回一個「差不多」的值會讓日後追出真值時看不出哪裡被污染過。
	var atkStrength, atkCounts, atkHits [gamedata.GroundUnitTypes]int
	atkStrength[groundTypeMarines] = atkForce +
		gamedata.GroundTypeStrengthDelta(groundTypeMarines) + groundMarineOnlyBonusFor(s.Player)
	atkCounts[groundTypeMarines] = s.Fleet().Marines
	atkHits[groundTypeMarines] = marineHits
	atkStrength[groundTypeTanks] = atkForce +
		gamedata.GroundTypeStrengthDelta(groundTypeTanks) + groundTankOnlyBonusFor(s.Player)
	atkCounts[groundTypeTanks] = tankCount
	// ⚠ 這裡**不再**另加 GroundTypeHitsDelta:`tankHitsToKillFor` 已經是手冊的成品值
	// (戰車 2、Battleoid 3),而那兩個數正好等於「基礎 1 + 類型 0 的 +1 (+ Battleoids 的 +1)」
	// ——再加一次就會變成 3 / 4。見 gap report 第 33 項(地面戰結構)的重建表。
	atkHits[groundTypeTanks] = tankHits

	defMarines := aiPlayer.ColonyMarines[colonyIdx]
	defTanks := aiPlayer.ColonyTanks[colonyIdx]
	defForce := aiMarineForce(*aiPlayer)
	// 守方 Commando 領袖加成(#5,2026-07-11 已接線;ruleprofile.go RuleProfile.DefenderCommandoBonus):
	// AIOpponent.Leaders(見該欄位註解)提供「AI 是否擁有 Commando 守將」的資料來源——
	// buildDemoAIOpponents 依種族性格開局固定指派(布拉西人 Tier2/姆瑞森人 Tier1/席隆人無),
	// 非手冊逐字的隨機雇用機制,是誠實標記的近似(與攻方 commandoLeaderTier(s.Leaders) 的
	// 「帝國全域清單當代理」同款近似紀律)。舊存檔 aiPlayer.Leaders 解碼為 nil 時
	// commandoLeaderTier(nil)=0,安全降級為無加成。
	defForce += gamedata.GroundCommandoDefenderForceBonus(commandoLeaderTier(aiPlayer.Leaders), s.RuleProfile.DefenderCommandoBonus)
	// 難度加成:原版**只給 AI**,人類玩家拿 0(`Compute_Player_Ground_Combat_Bonuses_`
	// @ 0xEC15C 的 `[player+0x28] == 100` 分支)。以「普通」為基準往兩邊偏。
	// 攻方是人類玩家,所以 atkForce 那邊不加——那不是漏掉,是原版就沒有。
	defForce += gamedata.GroundDifficultyBonus(s.Difficulty, gamedata.GroundAIEmpire)
	defHighG := aiRaceHasTrait(*aiPlayer, gamedata.TRAIT_HIGH_G)
	defHits := gamedata.GroundMarineHitsToKill(defHighG, hasPoweredArmorFor(aiPlayer.Player))
	// 守方:陸戰隊 + **民兵**。原版的殖民地填三格(裝甲 / 陸戰隊 / 民兵,見
	// `Compute_Colony_Ground_Combat_Info_` @ 0xED713 與手冊「your militia are also shown here」)。
	//
	// 民兵 = ⌊人口 / 5⌋(原版 `Colony_N_Militia_` @ 0xEC61E),攻擊力比陸戰隊低 10。
	// 裝甲營在類型 0，與原版的殖民地三格順序一致。
	var defStrength, defCounts, defHitsArr [gamedata.GroundUnitTypes]int
	defStrength[groundTypeTanks] = defForce + gamedata.GroundTypeStrengthDelta(groundTypeTanks) + groundTankOnlyBonusFor(aiPlayer.Player)
	defCounts[groundTypeTanks] = defTanks
	defHitsArr[groundTypeTanks] = tankHitsToKillFor(aiPlayer.Player, defHighG)
	defStrength[groundTypeMarines] = defForce + gamedata.GroundTypeStrengthDelta(groundTypeMarines)
	defCounts[groundTypeMarines] = defMarines
	defHitsArr[groundTypeMarines] = defHits
	militiaCount := gamedata.ColonyMilitiaUnits(colony.Population)
	defStrength[gamedata.GroundTypeMilitia] = defForce +
		gamedata.GroundTypeStrengthDelta(gamedata.GroundTypeMilitia)
	defCounts[gamedata.GroundTypeMilitia] = militiaCount
	defHitsArr[gamedata.GroundTypeMilitia] = defHits

	// 原版 Random_ 是 1..n 的 32-bit LCG；Ground_Combat_Round_ 只比較兩個
	// Random(100) 的大小，所以轉成 [0,100) 不改變命中／被命中的結果，但保留
	// 原版 rejection sampling 與後續抽樣的 state 序列。seed 仍是 remake 的
	// 回合／星索引穩定映射，尚未與原版全域 save seed 接軌，這個邊界在文件中保留。
	origRand := gamedata.NewOrigRand(uint32(int64(s.Turn)*2654435761 + int64(starIdx)*97 + 555))
	originalRoll := func(n int) int { return origRand.N(n) - 1 }
	// 換成原版的解算(`Ground_Combat_Round_` @ 0xEC4FE,見 gamedata/ground_battle_orig.go)。
	// 先前用的是一代 1oom 的結構,與二代有三處實質差異,最要緊的是**平手時雙方都挨打**。
	// 擲骰用 `originalRoll` 對應原版的 `Random_(100)`。
	atkSide := gamedata.NewGroundSide(atkStrength, atkCounts, atkHits)
	defSide := gamedata.NewGroundSide(defStrength, defCounts, defHitsArr)
	res := gamedata.ResolveGroundCombatOrig(atkSide, defSide, originalRoll, 0)

	// 拆回陸戰隊/戰車營各自存活數——現在是**逐類型的真實剩餘數**,
	// 不再是先前那個「戰車排在尾端所以先算給戰車」的推算。
	marinesSurvived := atkSide.Count[groundTypeMarines]
	tanksSurvived := atkSide.Count[groundTypeTanks]

	out := GroundInvasionResult{
		Ok: true, AttackerWon: res.AttackerWon,
		AttackerMarinesStart: s.Fleet().Marines, AttackerTanksStart: tankCount,
		DefenderStart: defTanks + defMarines + militiaCount, ColonyName: s.starName(starIdx),
		AttackerSurvived: res.AttackerSurvived, DefenderSurvived: res.DefenderSurvived,
		AttackerMarinesSurvived: marinesSurvived, AttackerTanksSurvived: tanksSurvived,
		Rounds: res.Rounds,
	}
	s.Fleet().Marines = marinesSurvived
	s.Fleet().Tanks = tanksSurvived
	// 原版戰後把守方各類型的剩餘部隊留在殖民地資料中；攻方失敗時下一次入侵
	// 必須看到這次戰損。人口不是這個欄位，不能用 DefenderSurvived 覆蓋。
	aiPlayer.ColonyMarines[colonyIdx] = defSide.Count[groundTypeMarines]
	aiPlayer.ColonyTanks[colonyIdx] = defSide.Count[groundTypeTanks]

	if res.AttackerWon {
		captured := colony
		transferredBuildings := s.prepareCapturedAIColony(aiIdx, colonyIdx, capturedPlanet)
		// `Resolve_Invasion_Troops @ 0xECECA` 寫回的是原始殖民地的入侵部隊欄位
		// (上限由 `Invade @ 0xED59D` 決定)，不是人口欄位。靜態證據沒有支持「守方
		// 戰鬥單位存活數 = 佔領後人口」，所以保留 captured.Population；markColonyConquered
		// 會把同一批原人口記成未同化人口。
		// 剛攻下來的殖民地整批是未同化的外族人口(手冊 p.21-24:依政體 2–20 回合
		// 同化一單位)。見 assimilation.go——這是「征服打法」在規則層的成本。
		markColonyConquered(&captured, aiIdx)
		s.PlayerColonies = append(s.PlayerColonies, captured)
		s.ensureColonyLeaderSlots()
		s.Builds = append(s.Builds, ColonyBuild{})
		for len(s.ColonyBuildings) < len(s.PlayerColonies) {
			s.ColonyBuildings = append(s.ColonyBuildings, nil)
		}
		s.ColonyBuildings[len(s.PlayerColonies)-1] = transferredBuildings
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
		// 行星索引同步(見 PlayerColonyPlanets 欄位註解):過戶的是**那一顆行星**上的殖民地,
		// 不是「該星系的代表行星」——同一個星系可能還有這個 AI 的另一個殖民地。
		for len(s.PlayerColonyPlanets) < len(s.PlayerColonies)-1 {
			s.PlayerColonyPlanets = append(s.PlayerColonyPlanets, -1)
		}
		s.PlayerColonyPlanets = append(s.PlayerColonyPlanets, capturedPlanet)
		// 俘虜人口:過戶過來的殖民地人口計入 CapturedPop(手冊 p.184 計分「You also get a
		// premium for captured population units」,見 score.go)。累計而非當下人口——之後這些
		// 人口自然成長或死亡都不影響「當初俘虜了多少」這個歷史數字。
		if idx := len(s.PlayerColonies) - 1; idx >= 0 && idx < len(s.PlayerColonies) {
			s.CapturedPop += s.PlayerColonies[idx].Population
		}

		aiPlayer.Colonies = append(aiPlayer.Colonies[:colonyIdx], aiPlayer.Colonies[colonyIdx+1:]...)
		aiPlayer.ColonyStars = append(aiPlayer.ColonyStars[:colonyIdx], aiPlayer.ColonyStars[colonyIdx+1:]...)
		if colonyIdx < len(aiPlayer.ColonyPlanets) {
			aiPlayer.ColonyPlanets = append(aiPlayer.ColonyPlanets[:colonyIdx], aiPlayer.ColonyPlanets[colonyIdx+1:]...)
		}
		// ColonyBuildings 同步移除對應項,維持三個平行陣列等長(見 AIOpponent.ColonyBuildings
		// 欄位註解)。colonyIdx 理論上恆在範圍內(與 Colonies/ColonyStars 同步維護),但仍防禦性
		// 檢查長度,避免舊存檔欄位缺失導致 panic。
		if colonyIdx < len(aiPlayer.ColonyBuildings) {
			aiPlayer.ColonyBuildings = append(aiPlayer.ColonyBuildings[:colonyIdx], aiPlayer.ColonyBuildings[colonyIdx+1:]...)
		}
		removeAIGroundForceSlot(aiPlayer, colonyIdx)
		s.recalcAIColonyMorale(aiIdx)
		s.recalcAllColonyMorale()
		// ⚠ 星的歸屬只在**這顆星上再也沒有敵方殖民地**時才翻面。同星系多殖民地打開之後
		// (第 24/24 項),一個星系可能有這個 AI 的兩個殖民地;打下一個就把整顆星判給玩家,
		// 會讓另一個殖民地變成「站在玩家星系裡的敵軍」,而且星圖顏色與可入侵性都對不上。
		stillEnemy := false
		for _, st := range aiPlayer.ColonyStars {
			if st == starIdx {
				stillEnemy = true
				break
			}
		}
		if !stillEnemy {
			star.Owner = 1
			if aiPlayer.OwnedStars > 0 {
				aiPlayer.OwnedStars--
			}
			out.StarCaptured = true
		}
		s.advanceConquestVictory() // 若這是該 AI 對手的最後一個殖民地,立即偵測「殲滅所有對手」勝利(見 council.go),不用等下個 EndTurn
	}
	return out
}

// HasBattleoids 回傳玩家是否已研究機動裝甲兵(手冊 p.101);地面戰畫面用來決定載具圖示。
func (s *GameSession) HasBattleoids() bool { return hasBattleoidsFor(s.Player) }
