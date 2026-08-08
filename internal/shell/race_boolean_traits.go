package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// race_boolean_traits.go:把原版特性陣列裡的**布林特性**接進遊戲。
//
// ============ 第 66 項留下的那半 ============
//
// 第 66 項挖出了原版 13 族的 31 格特性表,並把**數值型**那 9 格(農業/工業/科研/金錢/
// 人口/艦攻/艦防/地面戰/諜報)接了進去。布林那 21 格當時只做到「查得到」。
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
// 那句話在寫下的當天是對的;第 66 項之後就不對了。
//
// ============ 為什麼是每回合重算,不是開局設一次 ============
//
// 比照第 60 項的成就同步:重算是冪等的,所以順序與呼叫次數都不影響結果,
// **而且舊存檔讀進來會自動補齊**——`RaceOrigIdx` 是新欄位,舊存檔解碼成 0(阿爾卡里),
// 但 `ApplyRace` 只在開局跑一次,不會再被呼叫。每回合同步就沒有這個問題。
//
// ⚠ 更前面一版曾經把種族編號與五個布林旗標**存起來**,那是錯的:舊存檔沒有那些欄位
// 會解出零值(種族編號 0 = 阿爾卡里,整個查錯族),存讀往返的狀態指紋也對不上。
// 現在一律由 `RaceIndex`(本來就存)算,見 raceOrigIdx。
//
// ============ 誠實留白 ============
//
//   - **魅力(TRAIT_CHARISMATIC)仍然不生效。** 現在查得到人類有它了,但手冊只寫
//     「assimilate conquered colonists **easily**」,沒給數字。查得到 ≠ 知道加多少。
//   - **其餘布林特性(水棲/食岩/半機械/創造力/幸運/全知/匿蹤艦/跨維度/母星品質)未接。**
//     它們要的是 remake 還沒有的機制(星球適居度模型、科技樹分支、偵測模型、母星生成),
//     不是「忘了接」。逐項狀態見 docs/re/01-gap-report.md 第 66 項。

// raceOrigIdx 回傳玩家種族在 gamedata 那張一手表上的編號;自訂種族回 −1。
//
// **算出來的,不是存起來的。** 早期版本把它存進 GameSession 與存檔,結果撞到三件事:
// 舊存檔沒有這個欄位會解出 0(= 阿爾卡里,整個查錯族)、存讀往返的狀態指紋對不上、
// 熱座席位換人時多一個要記得抄的欄位。`RaceIndex` 本來就存了,由它算就全部消失。
//
// 這也是 `rulebook/70-deep-modules` 的一般原則:**衍生狀態不該有第二個真相來源。**
func (s *GameSession) raceOrigIdx() int {
	if s.RaceIndex < 0 || s.RaceIndex >= len(Races) {
		return -1 // 自訂種族:走點數畫面自己組,不在原版 13 族表上
	}
	return Races[s.RaceIndex].OrigIdx
}

// RaceWarlord 回報是否為統帥種族(姆瑞森)。
//
// 影響艦員經驗階梯(整條往上平移一格)與營房容量(加倍)。
func (s *GameSession) RaceWarlord() bool { return s.raceHasTrait(gamedata.TRAIT_WARLORD) }

// RaceRepulsive 回報是否為惹人厭種族(矽基):同化速度減半。
func (s *GameSession) RaceRepulsive() bool { return s.raceHasTrait(gamedata.TRAIT_REPULSIVE) }

// RaceCharismatic 回報是否為魅力種族(人類)。
//
// ⚠ **查得到但還不生效**:手冊只說「assimilate conquered colonists **easily**」,沒給數字。
// 見 gamedata/assimilation.go 的誠實留白——「查得到這一族有沒有」與「有了要加多少」
// 是兩個問題,第 66/66 項解掉的是前者。
func (s *GameSession) RaceCharismatic() bool { return s.raceHasTrait(gamedata.TRAIT_CHARISMATIC) }

// RaceTolerant 回報是否為寬容種族(矽基):不必花產能清污染。
func (s *GameSession) RaceTolerant() bool { return s.raceHasTrait(gamedata.TRAIT_TOLERANT) }

// RaceFantasticTrader 回報是否為神級商人種族(諾蘭姆)。
func (s *GameSession) RaceFantasticTrader() bool {
	return s.raceHasTrait(gamedata.TRAIT_FANTASTIC_TRADERS)
}

// syncRaceEngineFields 把種族布林特性推進引擎層(引擎不認識 shell 的種族表)。
//
// 每回合結算前呼叫,冪等。與 syncAchievementColonyFields(第 60 項)同一種形狀。
// **只有跨層那兩個欄位需要同步**——shell 自己用的一律走上面那幾個方法,不留副本。
func (s *GameSession) syncRaceEngineFields() {
	s.Player.FantasticTrader = s.RaceFantasticTrader()
	// 偵察實驗室(第 69 項):艦隊研究每回合重算——船會造、會沉、會被拆,
	// 存一份下來遲早不同步。與這裡的種族旗標同一個理由,所以搭同一班車。
	s.Player.FleetResearch = s.FleetResearchPoints()
	tolerant := s.RaceTolerant()
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].TolerantRace = tolerant
	}
}
