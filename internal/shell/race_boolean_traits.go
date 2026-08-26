package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// race_boolean_traits.go:把原版特性陣列裡的**布林特性**接進遊戲。
//
// ============ 第 65 項(種族特性31格)留下的那半 ============
//
// 第 65 項(種族特性31格)挖出了原版 13 族的 31 格特性表,並把**數值型**那 9 格(農業/工業/科研/金錢/
// 人口/艦攻/艦防/地面戰/諜報)接了進去。布林那 21 格先做到「查得到」,再逐條接已有消費端。
//
// 這一項接的是「查得到之後怎麼用」。四條規則**全部早就寫好了**,缺的一直只是
// 「這一族到底有沒有這個特性」:
//
//	TRAIT_WARLORD(姆瑞森)      crewXPThresholdsWarlord / Ground*BarracksCap 的 warlord 參數
//	TRAIT_REPULSIVE(矽基)      gamedata.AssimilationTurns 的 repulsive 分支
//	TRAIT_TOLERANT(矽基)       engine.ColonyState.TolerantRace → PollutionCleanupCost
//	TRAIT_FANTASTIC_TRADERS(諾蘭姆) TradeGoodsIncome / IncomeFoodSurplusRevenue 的 fantasticTrader 參數
//
// 那四處的呼叫端先前全部硬傳 `false`(或欄位根本沒有人寫入)。`session.go` 的欄位註解
// 把理由寫得很清楚:「**目前沒有任何內建種族會設它**——十三經典種族的特質表還沒有特質欄位」。
// 那句話在寫下的當天是對的;第 65 項(種族特性31格)之後就不對了。
//
// 客製種族的布林選項另由 GameSession.CustomRaceTraits 保存；這不是原版 13 族的衍生副本,
// 而是玩家在點數畫面做出的選擇。已有公式仍統一走 raceHasTrait,因此內建與客製不會各有一套規則。
//
// ============ 為什麼內建種族是每回合重算,不是開局設一次 ============
//
// 比照第 59 項(成就科技效果)的成就同步:重算是冪等的,所以順序與呼叫次數都不影響結果,
// **而且舊存檔讀進來會自動補齊**——`RaceOrigIdx` 是新欄位,舊存檔解碼成 0(阿爾卡里),
// 但 `ApplyRace` 只在開局跑一次,不會再被呼叫。每回合同步就沒有這個問題。
//
// ⚠ 更前面一版曾經把種族編號與五個布林旗標**存起來**,那是錯的:舊存檔沒有那些欄位
// 會解出零值(種族編號 0 = 阿爾卡里,整個查錯族),存讀往返的狀態指紋也對不上。
// 內建種族一律由 `RaceIndex`(本來就存)算,見 raceOrigIdx；客製種族則由選項遮罩本身查詢。
//
// ============ 誠實留白 ============
//
//   - **跨維度/母星品質仍受現有星球生成模型限制。** 幸運、全知、匿蹤艦、心靈感應
//     已在事件、星圖偵測、外交／諜報與殖民地行動入口接線；跨維度與母星品質不在本輪。

// raceOrigIdx 回傳玩家種族在 gamedata 那張一手表上的編號;自訂種族回 −1。
//
// **算出來的,不是存起來的。** 早期版本把它存進 GameSession 與存檔,結果撞到三件事:
// 舊存檔沒有這個欄位會解出 0(= 阿爾卡里,整個查錯族)、存讀往返的狀態指紋對不上、
// 熱座席位換人時多一個要記得抄的欄位。`RaceIndex` 本來就存了,由它算就全部消失。
//
// 客製種族的選項不在原版 13 族表內,所以由 GameSession.CustomRaceTraits 保存；
// 這是選項本身的真實來源，不是把原版衍生特性複製成第二份。
func (s *GameSession) raceOrigIdx() int {
	if s.RaceIndex < 0 || s.RaceIndex >= len(Races) {
		return -1 // 自訂種族:走點數畫面自己組,不在原版 13 族表上
	}
	return Races[s.RaceIndex].OrigIdx
}

// RaceWarlord 回報是否為統帥種族(姆瑞森或客製選項)。
//
// 影響艦員經驗階梯(整條往上平移一格)與營房容量(加倍)。
func (s *GameSession) RaceWarlord() bool { return s.raceHasTrait(gamedata.TRAIT_WARLORD) }

// RaceRepulsive 回報是否為惹人厭種族(矽基或客製選項):同化速度減半。
func (s *GameSession) RaceRepulsive() bool { return s.raceHasTrait(gamedata.TRAIT_REPULSIVE) }

// RaceCharismatic 回報是否為魅力種族(人類或客製選項)。外交 +50% 與原版
// sub_E3456 證實的同化速率加倍均已接線。
func (s *GameSession) RaceCharismatic() bool { return s.raceHasTrait(gamedata.TRAIT_CHARISMATIC) }

// RaceTolerant 回報是否為寬容種族(矽基或客製選項):不必花產能清污染。
func (s *GameSession) RaceTolerant() bool { return s.raceHasTrait(gamedata.TRAIT_TOLERANT) }

// RaceFantasticTrader 回報是否為神級商人種族(諾蘭姆或客製選項)。
func (s *GameSession) RaceFantasticTrader() bool {
	return s.raceHasTrait(gamedata.TRAIT_FANTASTIC_TRADERS)
}

// RaceCybernetic 回報是否為半機械化種族(梅克拉或客製選項)。
// 殖民地半單位食物／生產帳本與每場戰鬥後完全修復均已接線；逐系統損傷的
// 10%／5% 回合修復仍需更細的艦艇損傷模型。
func (s *GameSession) RaceCybernetic() bool { return s.raceHasTrait(gamedata.TRAIT_CYBERNETIC) }

// RaceLithovore 回報是否為食岩種族(矽基或客製選項)。
// 引擎層以 ColonyState.Lithovore 實作每人口 0 食物消耗與免農業饑荒。
func (s *GameSession) RaceLithovore() bool { return s.raceHasTrait(gamedata.TRAIT_LITHOVORE) }

// RaceCreative 回報是否為富創造力種族。研究完成時會讓該領域保持「未明確
// 抉擇」狀態,由既有科技門檻解鎖該領域全部應用。
func (s *GameSession) RaceCreative() bool { return s.raceHasTrait(gamedata.TRAIT_CREATIVE) }

// RaceUncreative 回報是否為缺乏創造力種族。研究完成時會由研究亂數流自動擇一,
// 不顯示一般種族的玩家待決畫面。
func (s *GameSession) RaceUncreative() bool { return s.raceHasTrait(gamedata.TRAIT_UNCREATIVE) }

// RaceTelepathic 回報是否為心靈感應種族。除外交／諜報加成外，會解鎖
// cruiser 以上艦艇的殖民地心靈控制行動。
func (s *GameSession) RaceTelepathic() bool { return s.raceHasTrait(gamedata.TRAIT_TELEPATHIC) }

// RaceLucky 回報是否為幸運種族。一般壞事件選中 Lucky 時取消；另有逐回合累積的
// 額外好事件擲骰，見 events.go 與 docs/spec/lucky-events.md。
func (s *GameSession) RaceLucky() bool { return s.raceHasTrait(gamedata.TRAIT_LUCKY) }

// RaceOmniscience 回報是否為全知種族。星圖會在開局直接揭露所有星球資料，
// 並讓敵方艦隊的可見性不受匿蹤遮蔽。
func (s *GameSession) RaceOmniscience() bool { return s.raceHasTrait(gamedata.TRAIT_OMNISCIENCE) }

// RaceStealthyShips 回報是否為匿蹤艦種族。對目前的抽象 AI 艦隊模型，效果
// 先落在「玩家艦隊不成為 AI 偵測目標」的明確 API；戰術戰鬥本身不加成。
func (s *GameSession) RaceStealthyShips() bool { return s.raceHasTrait(gamedata.TRAIT_STEALTHY_SHIPS) }

func aiRaceHasTrait(a AIOpponent, t gamedata.RaceTrait) bool {
	idx := aiRaceIndex(a)
	if idx < 0 || idx >= len(Races) {
		return false
	}
	return gamedata.OrigRaceHasTrait(Races[idx].OrigIdx, t)
}

// aiRaceSpyBonus 回傳 AI 原版種族的諜報加成。AI 沒有客製種族遮罩，故只從
// RaceIndex／舊存檔名稱回填的 Races 表讀取；未知種族安全退回 0。
func aiRaceSpyBonus(a AIOpponent) int {
	idx := aiRaceIndex(a)
	if idx < 0 || idx >= len(Races) {
		return 0
	}
	return Races[idx].SpyBonus + boolTraitBonus(aiRaceHasTrait(a, gamedata.TRAIT_TELEPATHIC), gamedata.SpyTelepathicRaceBonus)
}

// aiSpyEmpireBonus 在 AI 種族／心靈感應共同值上加入官方五級 Spy Bonus。
// 官方表未拆攻守，remake 依共同帝國值同時供 AB／DB 消費；證據邊界見
// docs/re/ai-spy-difficulty-audit-20260826.md。
func aiSpyEmpireBonus(a AIOpponent, difficulty int) int {
	bonus := aiRaceSpyBonus(a)
	if d, ok := ai.AIDifficultyBonus(ai.Difficulty(difficulty)); ok {
		bonus += d.SpyBonus
	}
	return bonus
}

func boolTraitBonus(active bool, value int) int {
	if active {
		return value
	}
	return 0
}

func raceGravityForTraits(lowG, highG bool) gamedata.PlanetGravity {
	if highG { // sub_DDF2C 先判 +0x8AA High-G，再判 +0x8A9 Low-G。
		return gamedata.HEAVY_G
	}
	if lowG {
		return gamedata.LOW_G
	}
	return gamedata.NORMAL_G
}

type colonistProductionProfile struct {
	food, industry, research int
	gravity                  gamedata.PlanetGravity
	gravityImmune            bool
	aquatic                  bool
	cybernetic               bool
	lithovore                bool
	tolerant                 bool
	subterranean             bool
	growth                   int
}

func (s *GameSession) playerColonistProductionProfile() colonistProductionProfile {
	p := colonistProductionProfile{gravity: raceGravityForTraits(
		s.raceHasTrait(gamedata.TRAIT_LOW_G), s.raceHasTrait(gamedata.TRAIT_HIGH_G)),
		aquatic: s.raceHasTrait(gamedata.TRAIT_AQUATIC), tolerant: s.RaceTolerant(),
		cybernetic: s.RaceCybernetic(), lithovore: s.RaceLithovore(),
		subterranean: s.raceHasTrait(gamedata.TRAIT_SUBTERRANEAN), growth: s.raceGrowthPct}
	if s.RaceIndex >= 0 && s.RaceIndex < len(Races) {
		r := Races[s.RaceIndex]
		p.food, p.industry, p.research = r.FoodBonus, r.IndBonus, r.ResBonus
	} else {
		p.food = int(s.CustomRaceRuntimeTraits[gamedata.TRAIT_FARMING])
		p.industry = int(s.CustomRaceRuntimeTraits[gamedata.TRAIT_INDUSTRY])
		p.research = int(s.CustomRaceRuntimeTraits[gamedata.TRAIT_SCIENCE])
	}
	return p
}

func aiColonistProductionProfile(a AIOpponent) colonistProductionProfile {
	p := colonistProductionProfile{gravity: gamedata.NORMAL_G}
	idx := aiRaceIndex(a)
	if idx < 0 || idx >= len(Races) {
		return p
	}
	r := Races[idx]
	p.food, p.industry, p.research = r.FoodBonus, r.IndBonus, r.ResBonus
	p.gravity = raceGravityForTraits(aiRaceHasTrait(a, gamedata.TRAIT_LOW_G), aiRaceHasTrait(a, gamedata.TRAIT_HIGH_G))
	p.aquatic = aiRaceHasTrait(a, gamedata.TRAIT_AQUATIC)
	p.cybernetic = aiRaceHasTrait(a, gamedata.TRAIT_CYBERNETIC)
	p.lithovore = aiRaceHasTrait(a, gamedata.TRAIT_LITHOVORE)
	p.tolerant = aiRaceHasTrait(a, gamedata.TRAIT_TOLERANT)
	p.subterranean = aiRaceHasTrait(a, gamedata.TRAIT_SUBTERRANEAN)
	p.growth = r.GrowthPct
	return p
}

func syncOwnerPopulationGroup(c *engine.ColonyState, slot int, slotKnown bool, p colonistProductionProfile, cacheAlreadyIncludesProfile bool) {
	if c == nil {
		return
	}
	if c.OwnerRaceProfileKnown {
		c.FoodPerFarmer += p.food - c.OwnerFoodBonus
		c.FoodPerFarmer += gamedata.ClimateFoodPerFarmer(gamedata.RaceFoodClimate(c.Climate, p.aquatic)) -
			gamedata.ClimateFoodPerFarmer(gamedata.RaceFoodClimate(c.Climate, c.Aquatic))
		c.IndustryPerWorker += p.industry - c.OwnerIndustryBonus
		c.ResearchPerScientist += p.research - c.OwnerResearchBonus
	} else if !cacheAlreadyIncludesProfile {
		c.FoodPerFarmer += p.food
		c.IndustryPerWorker += p.industry
		c.ResearchPerScientist += p.research
	}
	c.OwnerFoodBonus, c.OwnerIndustryBonus, c.OwnerResearchBonus = p.food, p.industry, p.research
	c.OwnerRaceProfileKnown = true
	c.OwnerRaceSlot, c.OwnerRaceSlotKnown = slot, slotKnown
	if len(c.PopulationGroups) == 0 && c.Farmers+c.Workers+c.Scientists == c.Population {
		c.PopulationGroups = []engine.PopulationGroup{{
			RaceSlot: slot, RaceSlotKnown: slotKnown, Farmers: c.Farmers, Workers: c.Workers, Scientists: c.Scientists,
			PrisonerFarmers: c.UnassimilatedFarmers, PrisonerWorkers: c.UnassimilatedWorkers,
			PrisonerScientists: c.UnassimilatedScientists, FoodBonus: p.food, IndustryBonus: p.industry,
			ResearchBonus: p.research, Gravity: p.gravity, Aquatic: p.aquatic,
			Cybernetic: p.cybernetic, Lithovore: p.lithovore, ProfileKnown: true,
			GravityImmune: p.gravityImmune, Tolerant: p.tolerant, Subterranean: p.subterranean,
			GrowthBonusPercent: p.growth,
		}}
	}
	for i := range c.PopulationGroups {
		g := &c.PopulationGroups[i]
		if slotKnown && g.RaceSlotKnown && g.RaceSlot == slot {
			g.FoodBonus, g.IndustryBonus, g.ResearchBonus = p.food, p.industry, p.research
			g.Gravity, g.Aquatic, g.ProfileKnown = p.gravity, p.aquatic, true
			g.Cybernetic, g.Lithovore = p.cybernetic, p.lithovore
			g.GravityImmune, g.Tolerant, g.Subterranean = p.gravityImmune, p.tolerant, p.subterranean
			g.GrowthBonusPercent = p.growth
		}
	}
}

// raceSpyBonusForActions 是玩家每次間諜行動使用的完整種族池：既有的
// TRAIT_SPYING 數值加成加上心靈感應手冊明列的 +10%。
func (s *GameSession) raceSpyBonusForActions() int {
	return s.RaceSpyBonus + boolTraitBonus(s.RaceTelepathic(), gamedata.SpyTelepathicRaceBonus)
}

// syncRaceEngineFields 把種族布林特性推進引擎層(引擎不認識 shell 的種族表)。
//
// 每回合結算前呼叫,冪等。與 syncAchievementColonyFields(第 59 項(成就科技效果))同一種形狀。
// **只有跨層那兩個欄位需要同步**——shell 自己用的一律走上面那幾個方法,不留副本。
func (s *GameSession) syncRaceEngineFields() {
	s.Player.FantasticTrader = s.RaceFantasticTrader()
	// 偵察實驗室(第 68 項(元件盤點+飛彈防禦)):艦隊研究每回合重算——船會造、會沉、會被拆,
	// 存一份下來遲早不同步。與這裡的種族旗標同一個理由,所以搭同一班車。
	s.Player.FleetResearch = s.FleetResearchPoints()
	tolerant := s.RaceTolerant()
	lithovore := s.RaceLithovore()
	cybernetic := s.RaceCybernetic()
	aquatic := s.raceHasTrait(gamedata.TRAIT_AQUATIC)
	subterranean := s.raceHasTrait(gamedata.TRAIT_SUBTERRANEAN)
	raceGravity := raceGravityForTraits(s.raceHasTrait(gamedata.TRAIT_LOW_G), s.raceHasTrait(gamedata.TRAIT_HIGH_G))
	profile := s.playerColonistProductionProfile()
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].TolerantRace = tolerant
		s.PlayerColonies[i].Lithovore = lithovore
		s.PlayerColonies[i].Cybernetic = cybernetic
		s.PlayerColonies[i].Subterranean = subterranean
		s.PlayerColonies[i].RaceGravity = raceGravity
		s.PlayerColonies[i].RaceGravityKnown = true
		slot, known := 0, true
		if s.PlayerColonies[i].OwnerRaceSlotKnown {
			slot = s.PlayerColonies[i].OwnerRaceSlot
		}
		syncOwnerPopulationGroup(&s.PlayerColonies[i], slot, known, profile, true)
		s.PlayerColonies[i].Aquatic = aquatic
	}
}

// syncAIRaceEngineFields 把內建 AI 種族也同步到引擎層。
// AIOpponent 沒有客製種族遮罩,只需由 RaceIndex/舊存檔名稱回填的索引查原版特性表。
func (s *GameSession) syncAIRaceEngineFields(a *AIOpponent) {
	if a == nil {
		return
	}
	idx := aiRaceIndex(*a)
	tolerant, lithovore, cybernetic, aquatic, subterranean := false, false, false, false, false
	raceGravity := gamedata.NORMAL_G
	profile := aiColonistProductionProfile(*a)
	if idx >= 0 && idx < len(Races) {
		orig := Races[idx].OrigIdx
		tolerant = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_TOLERANT)
		lithovore = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_LITHOVORE)
		cybernetic = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_CYBERNETIC)
		aquatic = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_AQUATIC)
		subterranean = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_SUBTERRANEAN)
		raceGravity = raceGravityForTraits(
			gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_LOW_G),
			gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_HIGH_G))
	}
	for i := range a.Colonies {
		a.Colonies[i].TolerantRace = tolerant
		a.Colonies[i].Lithovore = lithovore
		a.Colonies[i].Cybernetic = cybernetic
		a.Colonies[i].Subterranean = subterranean
		a.Colonies[i].RaceGravity = raceGravity
		a.Colonies[i].RaceGravityKnown = true
		slot, known := a.PopulationRaceSlot, a.PopulationRaceSlotKnown
		if !known {
			for j := range s.AIPlayers {
				if &s.AIPlayers[j] == a {
					slot, known = j+1, true
					break
				}
			}
		}
		syncOwnerPopulationGroup(&a.Colonies[i], slot, known, profile, false)
		a.Colonies[i].Aquatic = aquatic
	}
}
