package shell

import (
	"fmt"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// events_persistent.go:**持續型**隨機事件——那些不是「發生一次扣點東西」而是「進入一個
// 會持續好幾回合的狀態,直到某個條件解除」的事件。人口暴增的期間與效果另由
// IDA `sub_206A2`／`sub_E1839` 閉合，不沿用下方手冊的 6 回合／5% 通則。
//
// 這一類先前全部卡在「缺前置子系統」:remake 只有「單次結算」的事件模型,沒有任何跨回合
// 的事件狀態。這個檔案補上那個模型,並把手冊 p.180-181 給的數字逐條落地——**那兩頁是逐事件
// 的規格書**,連回合門檻與機率都寫死了:
//
//	超新星     手冊說至少 200 回合、全星系 RP 轉用；1.31 IDA 進一步證實倒數為
//	           Random(5)+10-difficulty、需求為建立時 system RP×倒數。
//	時空異象   "All colonies in the system are temporarily held in stasis, unable to produce,
//	           grow, or move population … (these colonies do not need food or cost maintenance
//	           either.) After it has been in effect for six turns, there is a five percent
//	           chance each turn that the anomaly will end."
//	超空間獸   "For any fleet traveling between systems, there is a random chance that one ship
//	           of that fleet will be yanked into another dimension and destroyed. After it's
//	           been around for six turns, there's a five percent chance each turn the beast
//	           will go away."
//	蟲洞       "The sudden appearance of one end of a wormhole in the path of a traveling fleet
//	           moves that fleet to their destination in a single turn."
//	太空怪獸   變形蟲 ≥100 回合、太空鰻 ≥150、太空水晶 ≥200、九頭蛇 ≥250、巨龍 ≥300;
//	           「Unless someone's fleet defeats it along the way, the monster will attack …」
//
// 事件 26 的 1.31 consumer 已由 IDA 證實：active 回合沒有額外攻擊機率，直接在指定帝國
// 的航行艦中 reservoir sampling。仍未知的是原版 raw 艦隊鏈首中間狀態、怪獸入侵移動速度，
// 以及 1.50 是否改過這些規則。

// PersistentEventKind 是持續型事件的種類。
type PersistentEventKind int

const (
	PersistentNone PersistentEventKind = iota
	// PersistentSupernova 超新星:某星系的恆星即將爆炸,靠研究點數搶救。
	PersistentSupernova
	// PersistentStasis 時空異象:某星系的殖民地凍結(不產出、不成長,也不吃食物、不繳維護費)。
	PersistentStasis
	// PersistentWarpBeast 超空間獸：逐回合從指定帝國的航行艦均勻拖走一艘。
	PersistentWarpBeast
	// PersistentPopulationBoom 人口暴增：指定殖民地的逐族人口成長加成 +100 百分點。
	PersistentPopulationBoom
	// PersistentPlague 瘟疫：指定殖民地成長 -200 百分點，研究進度完成後解除。
	PersistentPlague
	// PersistentComet 彗星：逐回合由目標星系停泊艦艇削減耐久，倒數歸零時撞擊。
	PersistentComet
	// PersistentPirateActivity 海盜活動：逐回合威脅運輸船，由同星停泊艦艇清剿。
	PersistentPirateActivity
	// PersistentHyperspaceFlux 超空間亂流：全銀河非跨維度艦隊停止航行。
	PersistentHyperspaceFlux
	// PersistentWarpFunnel 曲速漏斗：1.31 原版只保存目標艦索引與新聞生命週期，
	// 沒有停止艦隊移動的消費端。
	PersistentWarpFunnel
)

// PersistentEvent 是一個進行中的持續型事件。
type PersistentEvent struct {
	Kind      PersistentEventKind
	StarIndex int // 影響的星(超空間獸是全銀河,值為 -1)
	Turns     int // 已持續回合數
	// Countdown 是超新星的剩餘回合;其餘種類為 0。
	Countdown int
	// ResearchNeeded / ResearchDone 是超新星的搶救進度(手冊:系統產生的研究點全部投入)。
	ResearchNeeded int
	ResearchDone   int
	// PlanetIndex 是殖民地型事件的穩定目標；人口暴增使用。其他種類忽略。
	PlanetIndex int
	// Strength／InitialStrength 是彗星目前／初始耐久；其他事件忽略。
	Strength        int
	InitialStrength int
	// TargetKind／TargetIndex 保存事件 26 每回合指定的原版帝國槽投影。舊 JSON
	// 缺欄位時零值為一般玩家 0，與舊 remake 只攻擊玩家的行為安全相容。
	TargetKind  eventEmpireKind
	TargetIndex int
}

// advancePersistentEvents 每回合推進所有進行中的持續型事件,回傳要顯示的訊息。
func (s *GameSession) advancePersistentEvents() []string {
	s.LastPersistentEventEN = ""
	if len(s.PersistentEvents) == 0 {
		return nil
	}
	var msgs, msgsEN []string
	kept := s.PersistentEvents[:0]
	for i := range s.PersistentEvents {
		e := &s.PersistentEvents[i]
		e.Turns++
		done, msg, msgEN := s.stepPersistentEvent(e)
		if msg != "" {
			msgs = append(msgs, msg)
		}
		if msgEN != "" {
			msgsEN = append(msgsEN, msgEN)
		}
		if !done {
			kept = append(kept, *e)
		}
	}
	s.PersistentEvents = kept
	s.LastPersistentEventEN = strings.Join(msgsEN, "|")
	return msgs
}

// stepPersistentEvent 推進單一持續型事件一回合,回傳 (是否結束, 中文訊息, 英文訊息)。
func (s *GameSession) stepPersistentEvent(e *PersistentEvent) (bool, string, string) {
	switch e.Kind {
	case PersistentSupernova:
		return s.stepSupernova(e)
	case PersistentStasis:
		// 凍結本身的效果在 EndTurn 的產出結算裡(見 StarInStasis);這裡只管什麼時候結束。
		// 原版在第 0..4 歲不抽亂數；提早呼叫 eventRoll 會改變後續事件的
		// 決定性亂數序列，因此只在實際進入 1/20 結束判定時取樣。
		roll := 0
		if e.Turns > 4 {
			roll = s.eventRoll(20)
		}
		if gamedata.OriginalStasisEnds(e.Turns, roll) {
			return true,
				fmt.Sprintf("%s 星系的時空異象消散,殖民地恢復運作", s.starName(e.StarIndex)),
				fmt.Sprintf("The space-time anomaly in the %s system dissipated; colonies resume operations.", s.starNameEN(e.StarIndex))
		}
		return false, "", ""
	case PersistentWarpBeast:
		// sub_206A2 先判定解除；已解除的回合不再呼叫 sub_100618 刪艦。
		roll := 0
		if e.Turns > 4 {
			roll = s.eventRoll(20)
		}
		if gamedata.OriginalStasisEnds(e.Turns, roll) {
			return true, "超空間獸遁入異次元,航道恢復安全",
				"The warp beast slipped back into another dimension; the space lanes are safe again."
		}
		msg, msgEN := s.warpBeastStrikeReport(e)
		s.retargetWarpBeast(e)
		return false, msg, msgEN
	case PersistentPopulationBoom:
		if !s.populationBoomTargetExists(e.PlanetIndex) {
			return true, "人口暴增事件因目標殖民地不復存在而結束",
				"The population boom ended because its colony no longer exists."
		}
		// sub_206A2：前五個 active turn 不擲骰；第六回合起每回合 Random(20)==1
		// 結束，age > 20 時即使骰失敗也強制結束。Turns=-1 是建立當回合的公告期。
		if e.Turns >= 6 {
			ended := s.eventRoll(20) == 1
			if ended || e.Turns >= 22 {
				return true, "人口暴增潮已恢復正常，殖民地出生率回到長期水準",
					"The population boom subsided and the colony's birth rate returned to normal."
			}
		}
		return false, "", ""
	case PersistentPlague:
		if !s.populationBoomTargetExists(e.PlanetIndex) {
			return true, "瘟疫因目標殖民地不復存在而結束",
				"The plague ended because its colony no longer exists."
		}
		if e.ResearchNeeded > 0 && e.ResearchDone >= e.ResearchNeeded {
			return true, "殖民地研究團隊完成疫苗，瘟疫已受到控制",
				"The colony's researchers completed a cure and brought the plague under control."
		}
		return false, "", ""
	case PersistentComet:
		return s.stepComet(e)
	case PersistentPirateActivity:
		return s.stepPirateActivity(e)
	case PersistentHyperspaceFlux:
		return s.stepHyperspaceFlux(e)
	case PersistentWarpFunnel:
		// sub_206A2 @ 0x212BE：age > 4 後 Random(20)==1 結束；
		// age > 20 強制結束。Turns=-1 讓建立公告回合不算 active turn。
		if e.Turns >= 5 && (s.eventRoll(20) == 1 || e.Turns >= 21) {
			return true, "困於曲速漏斗的艦隊已平安脫困，所有船艦與船員均安然無恙",
				"The fleet trapped in the warp funnel broke free; every ship and crew member survived unharmed."
		}
		return false, "", ""
	}
	return true, "", ""
}

func (s *GameSession) populationBoomTargetExists(planetIndex int) bool {
	for i := range s.PlayerColonies {
		if s.ColonyPlanetIndex(i) == planetIndex {
			return true
		}
	}
	for i := range s.AIPlayers {
		for j := range s.AIPlayers[i].Colonies {
			if j < len(s.AIPlayers[i].ColonyPlanets) && s.AIPlayers[i].ColonyPlanets[j] == planetIndex {
				return true
			}
		}
	}
	return false
}

func (s *GameSession) planetPopulationBoomActive(planetIndex int) bool {
	for i := range s.PersistentEvents {
		e := &s.PersistentEvents[i]
		if e.Kind == PersistentPopulationBoom && e.PlanetIndex == planetIndex {
			return true
		}
	}
	return false
}

func (s *GameSession) planetPlagueActive(planetIndex int) bool {
	for i := range s.PersistentEvents {
		e := &s.PersistentEvents[i]
		if e.Kind == PersistentPlague && e.PlanetIndex == planetIndex {
			return true
		}
	}
	return false
}

func (s *GameSession) planetColonyEventActive(planetIndex int) bool {
	return s.planetPopulationBoomActive(planetIndex) || s.planetPlagueActive(planetIndex) || s.planetCometActive(planetIndex)
}

// stepSupernova 對應 sub_206A2 case 24：先扣該星本回合全部殖民地 RP，成功則立即結束；
// 尚未成功才遞減倒數，歸零時摧毀該星所有 owner 的殖民地。
func (s *GameSession) stepSupernova(e *PersistentEvent) (bool, string, string) {
	gained := s.supernovaSystemResearch(e.StarIndex)
	e.ResearchDone += gained

	if e.ResearchDone >= e.ResearchNeeded {
		return true,
			fmt.Sprintf("%s 星系的科學家搶在爆炸前穩定了恆星核心,超新星危機解除", s.starName(e.StarIndex)),
			fmt.Sprintf("Scientists in the %s system stabilized the stellar core before detonation; the supernova crisis is over.", s.starNameEN(e.StarIndex))
	}
	e.Countdown--
	if e.Countdown > 0 {
		return false,
			fmt.Sprintf("%s 星系恆星不穩定,倒數 %d 回合(搶救進度 %d/%d", s.starName(e.StarIndex), e.Countdown, e.ResearchDone, e.ResearchNeeded),
			fmt.Sprintf("The star in the %s system is unstable; %d turns remain (rescue progress %d/%d).", s.starNameEN(e.StarIndex), e.Countdown, e.ResearchDone, e.ResearchNeeded)
	}

	// 0x21110..0x2116B：逐 active colony 把其 planet+0x08 寫 1，再 sub_DCDAC(...,-1)。
	lost := s.destroyColoniesAtStar(e.StarIndex)
	return true,
		fmt.Sprintf("%s 的恆星爆發為超新星,%d 座殖民地全滅,整個星系化為輻射廢土", s.starName(e.StarIndex), lost),
		fmt.Sprintf("The star in the %s system went supernova; %d colonies were destroyed and the system became a radioactive wasteland.", s.starNameEN(e.StarIndex), lost)
}

// warpBeastStrike 是測試／相容入口；依目前玩家建立暫態 record，執行原版的一次選艦 consumer。
func (s *GameSession) warpBeastStrike() string {
	e := &PersistentEvent{Kind: PersistentWarpBeast, TargetKind: eventEmpirePlayer, TargetIndex: 0}
	msg, _ := s.warpBeastStrikeReport(e)
	return msg
}

func (s *GameSession) warpBeastFleetShip(fleets []Fleet) (int, int, bool) {
	s.eventRandForTest()
	fleetIndex, shipIndex, candidates := -1, -1, 0
	for fi := range fleets {
		if fleets[fi].ETA <= 0 {
			continue
		}
		for si := range fleets[fi].Ships {
			candidates++
			if s.eventRand.Intn(candidates) == 0 {
				fleetIndex, shipIndex = fi, si
			}
		}
	}
	return fleetIndex, shipIndex, fleetIndex >= 0
}

func removeFleetShipAt(fleets []Fleet, fleetIndex, shipIndex int) (string, bool) {
	if fleetIndex < 0 || fleetIndex >= len(fleets) || shipIndex < 0 || shipIndex >= len(fleets[fleetIndex].Ships) {
		return "", false
	}
	lost := fleets[fleetIndex].Ships[shipIndex].Name
	fleets[fleetIndex].Ships = append(fleets[fleetIndex].Ships[:shipIndex], fleets[fleetIndex].Ships[shipIndex+1:]...)
	return lost, true
}

func (s *GameSession) warpBeastStrikeReport(e *PersistentEvent) (string, string) {
	target := eventEmpireTarget{kind: e.TargetKind, index: e.TargetIndex}
	lost, ok := "", false
	switch target.kind {
	case eventEmpireSeat:
		if target.index < 0 || target.index >= len(s.Seats) {
			return "", ""
		}
		if target.index == s.ActiveSeat {
			fi, si, found := s.warpBeastFleetShip(s.Fleets)
			if found {
				lost, ok = removeFleetShipAt(s.Fleets, fi, si)
			}
		} else {
			fi, si, found := s.warpBeastFleetShip(s.Seats[target.index].Fleets)
			if found {
				lost, ok = removeFleetShipAt(s.Seats[target.index].Fleets, fi, si)
			}
		}
	case eventEmpireAI:
		if target.index < 0 || target.index >= len(s.AIPlayers) || s.AIPlayers[target.index].FleetETA <= 0 {
			return "", ""
		}
		a := &s.AIPlayers[target.index]
		s.eventRandForTest()
		if len(a.Ships) > 0 {
			si := s.eventRand.Intn(len(a.Ships))
			lost = a.Ships[si].Name
			a.Ships = append(a.Ships[:si], a.Ships[si+1:]...)
			s.syncAIShipStrength(target.index)
			ok = true
		}
	default:
		fi, si, found := s.warpBeastFleetShip(s.Fleets)
		if found {
			lost, ok = removeFleetShipAt(s.Fleets, fi, si)
		}
	}
	if !ok {
		return "", ""
	}
	if lost == "" {
		lost = "一艘艦艇"
	}
	name := s.eventEmpireTargetName(target)
	return fmt.Sprintf("%s 航行中的艦隊被超空間獸扯進異次元，失去了「%s」", name, lost),
		fmt.Sprintf("A ship of %s in transit was dragged into another dimension by the warp beast and destroyed.", name)
}

func (s *GameSession) retargetWarpBeast(e *PersistentEvent) {
	ev := gamedata.RandomEventByID(26)
	if ev == nil {
		e.TargetIndex = -1
		return
	}
	target, ok := s.chooseEventEmpireTargetByWeight(*ev)
	if !ok {
		e.TargetIndex = -1
		return
	}
	e.TargetKind, e.TargetIndex = target.kind, target.index
}

// StarInStasis 回傳該星是否處於時空異象的凍結狀態(供產出結算跳過)。
func (s *GameSession) StarInStasis(starIdx int) bool {
	for _, e := range s.PersistentEvents {
		if e.Kind == PersistentStasis && e.StarIndex == starIdx {
			return true
		}
	}
	return false
}

// StarUnderSupernova 回傳該星研究是否正被超新星搶救計畫全數轉用。
func (s *GameSession) StarUnderSupernova(starIdx int) bool {
	for _, e := range s.PersistentEvents {
		if e.Kind == PersistentSupernova && e.StarIndex == starIdx {
			return true
		}
	}
	return false
}

// ColonyInStasis 回傳玩家第 i 個殖民地是否被時空異象凍結。
func (s *GameSession) ColonyInStasis(i int) bool {
	star := s.PlayerColonyStarIndex(i)
	if star < 0 {
		return false
	}
	return s.StarInStasis(star)
}

func freezeColonyForStasis(c *engine.ColonyState) {
	if c == nil {
		return
	}
	c.Population = 0
	c.Farmers, c.Workers, c.Scientists = 0, 0, 0
	c.FlatFood, c.FlatIndustry, c.FlatResearch = 0, 0, 0
	c.SpecialIncome = 0
}

// hasPersistentEvent 回傳是否已有同種類、同一顆星的持續事件(避免重複疊加)。
func (s *GameSession) hasPersistentEvent(kind PersistentEventKind, starIdx int) bool {
	for _, e := range s.PersistentEvents {
		if e.Kind == kind && (starIdx < 0 || e.StarIndex == starIdx) {
			return true
		}
	}
	return false
}

func (s *GameSession) radiatePlanet(planet int) {
	if planet < 0 || planet >= len(s.Planets) {
		return
	}
	s.Planets[planet].ClimateID = gamedata.RADIATED
	s.Planets[planet].Climate = climateDisplayName(gamedata.RADIATED)
}

// destroyColoniesAtStar 對應原版五個殖民槽：摧毀該星玩家、熱座與 AI 殖民地。
func (s *GameSession) destroyColoniesAtStar(starIdx int) int {
	n := 0
	removeLoadedSeat := func() {
		for i := len(s.PlayerColonies) - 1; i >= 0; i-- {
			if s.PlayerColonyStarIndex(i) != starIdx {
				continue
			}
			s.radiatePlanet(s.ColonyPlanetIndex(i))
			s.removePlayerColony(i)
			n++
		}
	}
	if s.HotseatEnabled() {
		active := s.ActiveSeat
		if active >= 0 && active < len(s.Seats) {
			s.Seats[active] = s.saveSeat()
		}
		for i := range s.Seats {
			s.loadSeat(s.Seats[i])
			removeLoadedSeat()
			s.Seats[i] = s.saveSeat()
		}
		if active >= 0 && active < len(s.Seats) {
			s.loadSeat(s.Seats[active])
		}
	} else {
		removeLoadedSeat()
	}
	for ai := range s.AIPlayers {
		a := &s.AIPlayers[ai]
		for i := len(a.Colonies) - 1; i >= 0; i-- {
			if i >= len(a.ColonyStars) || a.ColonyStars[i] != starIdx {
				continue
			}
			if i < len(a.ColonyPlanets) {
				s.radiatePlanet(a.ColonyPlanets[i])
			}
			s.removeAIColonyAfterEvent(ai, i)
			n++
		}
	}
	if starIdx >= 0 && starIdx < len(s.Stars) && !s.galaxyHasActiveColonyAtStar(starIdx) {
		s.Stars[starIdx].Owner = 0
	}
	return n
}

// removePlayerColony 從 PlayerColonies 與所有平行陣列移除第 i 個殖民地。
// 平行陣列的長度不變量見 ColonizeStar/appendPlayerColony——這裡是它的反向操作。
func (s *GameSession) removePlayerColony(i int) {
	cut := func(n int) bool { return i >= 0 && i < n }
	if !cut(len(s.PlayerColonies)) {
		return
	}
	s.PlayerColonies = append(s.PlayerColonies[:i], s.PlayerColonies[i+1:]...)
	if cut(len(s.Builds)) {
		s.Builds = append(s.Builds[:i], s.Builds[i+1:]...)
	}
	if cut(len(s.BuildQueue)) {
		s.BuildQueue = append(s.BuildQueue[:i], s.BuildQueue[i+1:]...)
	}
	if cut(len(s.AutoBuild)) {
		s.AutoBuild = append(s.AutoBuild[:i], s.AutoBuild[i+1:]...)
	}
	if cut(len(s.RepeatBuild)) {
		s.RepeatBuild = append(s.RepeatBuild[:i], s.RepeatBuild[i+1:]...)
	}
	if cut(len(s.LastBuilt)) {
		s.LastBuilt = append(s.LastBuilt[:i], s.LastBuilt[i+1:]...)
	}
	if cut(len(s.ColonyBuildings)) {
		s.ColonyBuildings = append(s.ColonyBuildings[:i], s.ColonyBuildings[i+1:]...)
	}
	if cut(len(s.PlayerColonyMarines)) {
		s.PlayerColonyMarines = append(s.PlayerColonyMarines[:i], s.PlayerColonyMarines[i+1:]...)
	}
	if cut(len(s.MarineBarracksAge)) {
		s.MarineBarracksAge = append(s.MarineBarracksAge[:i], s.MarineBarracksAge[i+1:]...)
	}
	if cut(len(s.PlayerColonyTanks)) {
		s.PlayerColonyTanks = append(s.PlayerColonyTanks[:i], s.PlayerColonyTanks[i+1:]...)
	}
	if cut(len(s.ArmorBarracksAge)) {
		s.ArmorBarracksAge = append(s.ArmorBarracksAge[:i], s.ArmorBarracksAge[i+1:]...)
	}
	if cut(len(s.popAccum)) {
		s.popAccum = append(s.popAccum[:i], s.popAccum[i+1:]...)
	}
	if cut(len(s.PlayerColonyStars)) {
		s.PlayerColonyStars = append(s.PlayerColonyStars[:i], s.PlayerColonyStars[i+1:]...)
	}
	if cut(len(s.PlayerColonyPlanets)) {
		s.PlayerColonyPlanets = append(s.PlayerColonyPlanets[:i], s.PlayerColonyPlanets[i+1:]...)
	}
	if cut(len(s.ColonyLeaderNames)) {
		s.ColonyLeaderNames = append(s.ColonyLeaderNames[:i], s.ColonyLeaderNames[i+1:]...)
	}
}

// starName 回傳星名(越界回「未知星系」)。
func (s *GameSession) starName(idx int) string {
	if idx < 0 || idx >= len(s.Stars) {
		return "未知星系"
	}
	return s.Stars[idx].Name
}

func (s *GameSession) starNameEN(idx int) string {
	if idx < 0 || idx >= len(s.Stars) {
		return "an unknown system"
	}
	if s.Stars[idx].NameEN != "" {
		return s.Stars[idx].NameEN
	}
	return s.Stars[idx].Name
}

// eventRoll 回傳 1..n 的擲骰(對齊原版 Random_ 的 1..n 語意),用事件亂數流。
func (s *GameSession) eventRoll(n int) int {
	if n < 1 {
		return 1
	}
	s.eventRandForTest()
	return s.eventRand.Intn(n) + 1
}

// --- 事件觸發:把持續型事件掛上去 ---

func colonyResearchNow(c engine.ColonyState) int { return engine.RunColonyTurn(c).Research }

func (s *GameSession) galaxyHasActiveColonyAtStar(star int) bool {
	for i, candidate := range s.PlayerColonyStars {
		if candidate == star && i < len(s.PlayerColonies) && s.PlayerColonies[i].Population > 0 {
			return true
		}
	}
	if s.HotseatEnabled() {
		for seatIndex := range s.Seats {
			if seatIndex == s.ActiveSeat {
				continue
			}
			v := &s.Seats[seatIndex]
			for i, candidate := range v.PlayerColonyStars {
				if candidate == star && i < len(v.PlayerColonies) && v.PlayerColonies[i].Population > 0 {
					return true
				}
			}
		}
	}
	for ai := range s.AIPlayers {
		for i, candidate := range s.AIPlayers[ai].ColonyStars {
			if candidate == star && i < len(s.AIPlayers[ai].Colonies) && s.AIPlayers[ai].Colonies[i].Population > 0 {
				return true
			}
		}
	}
	return false
}

// galaxyHasEventColonyWithoutCapitolAtStar 對映 sub_23A5F 的候選星條件：至少一座
// active colony 的 raw building 9 槽為零。這只決定超新星能否選該星，不縮小成立後的效果範圍。
func (s *GameSession) galaxyHasEventColonyWithoutCapitolAtStar(star int) bool {
	for i, candidate := range s.PlayerColonyStars {
		if candidate == star && i < len(s.PlayerColonies) && s.PlayerColonies[i].Population > 0 &&
			!colonyHasCapitol(s.ColonyBuildings, i) {
			return true
		}
	}
	if s.HotseatEnabled() {
		for seatIndex := range s.Seats {
			if seatIndex == s.ActiveSeat {
				continue
			}
			v := &s.Seats[seatIndex]
			for i, candidate := range v.PlayerColonyStars {
				if candidate == star && i < len(v.PlayerColonies) && v.PlayerColonies[i].Population > 0 &&
					!colonyHasCapitol(v.ColonyBuildings, i) {
					return true
				}
			}
		}
	}
	for ai := range s.AIPlayers {
		a := &s.AIPlayers[ai]
		for i, candidate := range a.ColonyStars {
			if candidate == star && i < len(a.Colonies) && a.Colonies[i].Population > 0 &&
				!colonyHasCapitol(a.ColonyBuildings, i) {
				return true
			}
		}
	}
	return false
}

func (s *GameSession) supernovaSystemResearch(star int) int {
	total := 0
	for i, candidate := range s.PlayerColonyStars {
		if candidate == star && i < len(s.PlayerColonies) && s.PlayerColonies[i].Population > 0 {
			total += colonyResearchNow(s.PlayerColonies[i])
		}
	}
	if s.HotseatEnabled() {
		for seatIndex := range s.Seats {
			if seatIndex == s.ActiveSeat {
				continue
			}
			v := &s.Seats[seatIndex]
			for i, candidate := range v.PlayerColonyStars {
				if candidate == star && i < len(v.PlayerColonies) && v.PlayerColonies[i].Population > 0 {
					total += colonyResearchNow(v.PlayerColonies[i])
				}
			}
		}
	}
	for ai := range s.AIPlayers {
		a := &s.AIPlayers[ai]
		for i, candidate := range a.ColonyStars {
			if candidate == star && i < len(a.Colonies) && a.Colonies[i].Population > 0 {
				total += colonyResearchNow(a.Colonies[i])
			}
		}
	}
	return total
}

// pickSupernovaStar 對應 sub_23A5F 的全銀河 rejection sampling，最多 1,000 次。
func (s *GameSession) pickSupernovaStar() (int, bool) {
	if len(s.Stars) == 0 {
		return -1, false
	}
	for try := 0; try < 1000; try++ {
		star := s.eventRoll(len(s.Stars)) - 1
		if s.galaxyHasEventColonyWithoutCapitolAtStar(star) && !s.pirateActivityConflictAtStar(star) {
			return star, true
		}
	}
	return -1, false
}

// startSupernova 建立事件 24 的全銀河星系 record。
func (s *GameSession) startSupernova() (string, bool) {
	if s.Turn-1 < 200 {
		return "", false
	}
	star, ok := s.pickSupernovaStar()
	if !ok {
		return "", false
	}
	countdown := gamedata.OriginalSupernovaCountdown(s.Difficulty, s.eventRoll(5))
	need := gamedata.OriginalSupernovaResearchNeed(s.supernovaSystemResearch(star), countdown)
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{
		Kind: PersistentSupernova, StarIndex: star, Countdown: countdown, ResearchNeeded: need,
	})
	return fmt.Sprintf("%s 的恆星進入不穩定狀態,將在 %d 回合後爆發為超新星——全系統研究能量已投入搶救(需 %d RP)",
		s.starName(star), countdown, need), true
}

func (s *GameSession) pickStasisStar(colonies []engine.ColonyState, stars []int) (int, bool) {
	s.eventRandForTest()
	i, ok := pickEarthquakeColony(colonies, nil, s.eventRand.Intn)
	if !ok || i < 0 || i >= len(stars) {
		return -1, false
	}
	star := stars[i]
	if star < 0 || s.pirateActivityConflictAtStar(star) {
		return -1, false
	}
	return star, true
}

// startStasis 讓目前玩家／載入中的熱座帝國殖民星進入時空異象。
func (s *GameSession) startStasis() (string, bool) {
	star, ok := s.pickStasisStar(s.PlayerColonies, s.PlayerColonyStars)
	if !ok {
		return "", false
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{Kind: PersistentStasis, StarIndex: star, Turns: -1})
	return fmt.Sprintf("%s 星系陷入時空異象,當地殖民地凍結——不生產、不成長,但也不需要食物與維護費",
		s.starName(star)), true
}

func (s *GameSession) startAIStasis(aiIndex int) (string, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return "", false
	}
	a := &s.AIPlayers[aiIndex]
	star, ok := s.pickStasisStar(a.Colonies, a.ColonyStars)
	if !ok {
		return "", false
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{Kind: PersistentStasis, StarIndex: star, Turns: -1})
	return fmt.Sprintf("%s 星系陷入時空異象，當地殖民地完全凍結", s.starName(star)), true
}

// startWarpBeast 讓超空間獸進入銀河(全域效果,StarIndex = -1)。
func (s *GameSession) startWarpBeast(target eventEmpireTarget) (string, bool) {
	if s.hasPersistentEvent(PersistentWarpBeast, -1) {
		return "", false
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{
		Kind: PersistentWarpBeast, StarIndex: -1, Turns: -1,
		TargetKind: target.kind, TargetIndex: target.index,
	})
	return "一頭半身沉在超空間裡的野獸開始在航道上遊蕩,星際航行變得危險", true
}

// startWarpFunnel 建立事件 27 的報告型 persistent record。IDA 1.31 證據顯示
// 原版會挑有效艦艇作新聞目標，但不凍結艦隊 ETA；remake 不自行補造停航效果。
func (s *GameSession) startWarpFunnel() (string, bool) {
	ships := s.AllShips()
	if len(ships) == 0 || s.hasPersistentEvent(PersistentWarpFunnel, -1) {
		return "", false
	}
	s.eventRandForTest()
	_ = ships[s.eventRand.Intn(len(ships))]
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{
		Kind: PersistentWarpFunnel, StarIndex: -1, Turns: -1,
	})
	return "一支艦隊被曲速漏斗牢牢困在超空間中，能否平安脫困仍是未知數", true
}

// wormholeFleetCandidate 對應 sub_100519 的 reservoir sampling：候選單位是船，效果單位是
// 該船所屬整支艦隊，因此艦數較多的艦隊有等比例較高的中選機率。
func (s *GameSession) wormholeFleetCandidate(fleets []Fleet) (int, bool) {
	s.eventRandForTest()
	selected, candidates := -1, 0
	for fi := range fleets {
		if fleets[fi].ETA <= 0 || fleets[fi].DestStar < 0 || fleets[fi].DestStar >= len(s.Stars) {
			continue
		}
		for range fleets[fi].Ships {
			candidates++
			if s.eventRand.Intn(candidates) == 0 {
				selected = fi
			}
		}
	}
	return selected, selected >= 0
}

// applyWormhole 對應事件 28 sub_100519 → sub_FFDDA：抽船定艦隊後立即抵達。
func (s *GameSession) applyWormhole() (string, bool) {
	// 原版 ship+0x6D==1 不具資格；remake 以已接線的帝國跨維度 trait 對映該快取。
	if s.raceHasTrait(gamedata.TRAIT_TRANS_DIMENSIONAL) {
		return "", false
	}
	fi, ok := s.wormholeFleetCandidate(s.Fleets)
	if !ok {
		return "", false
	}
	f := &s.Fleets[fi]
	was, ships, dest := f.ETA, len(f.Ships), f.DestStar
	if !s.completePlayerFleetArrival(f) {
		return "", false
	}
	s.mergeColocatedFleets()
	return fmt.Sprintf("一端蟲洞突然出現在艦隊航道上，%d 艘艦艇立刻抵達 %s 星系（原航程尚餘 %d 回合）",
		ships, s.starName(dest), was), true
}

func (s *GameSession) applyAIWormhole(aiIndex int) (string, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return "", false
	}
	a := &s.AIPlayers[aiIndex]
	if aiRaceHasTrait(*a, gamedata.TRAIT_TRANS_DIMENSIONAL) || a.FleetETA <= 0 || len(a.Ships) == 0 ||
		a.FleetDestStar < 0 || a.FleetDestStar >= len(s.Stars) {
		return "", false
	}
	was, ships, dest := a.FleetETA, len(a.Ships), a.FleetDestStar
	targetAI := -1
	if a.FleetTargetAISet {
		targetAI = a.FleetTargetAI
	}
	a.FleetStar, a.FleetDestStar, a.FleetETA = dest, -1, 0
	a.FleetPosSet = true
	a.FleetTargetAI, a.FleetTargetAISet = -1, false
	s.queueOrionDiscoveryBroadcast(eventEmpireTarget{kind: eventEmpireAI, index: aiIndex, alive: true}, dest)
	s.applyArtemisMinesToAIFleet(aiIndex, dest)
	if s.EnableAIVsAI && targetAI >= 0 && targetAI < len(s.AIPlayers) && targetAI != aiIndex {
		s.LastAIAIBattle = s.resolveAIAIBattle(aiIndex, targetAI, dest)
	}
	return fmt.Sprintf("%s 的蟲洞航程讓 %d 艘艦艇立刻抵達 %s 星系（原航程尚餘 %d 回合）",
		stripAILabel(a.Name), ships, s.starName(dest), was), true
}

func eventMonsterKind(eventID int) (gamedata.SpaceMonster, bool) {
	switch eventID {
	case 19:
		return gamedata.MonsterAmoeba, true
	case 20:
		return gamedata.MonsterCrystal, true
	case 21:
		return gamedata.MonsterDragon, true
	case 22:
		return gamedata.MonsterEel, true
	case 23:
		return gamedata.MonsterHydra, true
	default:
		return gamedata.MonsterNone, false
	}
}

// reservoirMonsterStar 對應 sub_23BEC：每個有效殖民地都是一個候選，保存其星系索引。
func (s *GameSession) reservoirMonsterStar(stars []int, active func(int) bool) (int, bool) {
	s.eventRandForTest()
	target, candidates := -1, 0
	for i, star := range stars {
		if star < 0 || star >= len(s.Stars) || !active(i) || s.StarGuardedByMonster(star) {
			continue
		}
		candidates++
		if s.eventRand.Intn(candidates) == 0 {
			target = star
		}
	}
	return target, target >= 0
}

func (s *GameSession) placeInvadingMonster(kind gamedata.SpaceMonster, target int) (string, bool) {
	st, ok := gamedata.MonsterStatsFor(kind)
	if !ok || target < 0 || target >= len(s.Stars) {
		return "", false
	}
	monster := MonsterGuard{StarIndex: target, Kind: kind, Structure: st.Structure, Armor: st.Armor,
		TransitETA: s.eventMonsterTransitETA(target)}
	if kind == gamedata.MonsterEel {
		monster.Count, monster.EelAges = 1, []int{0}
	}
	s.Monsters = append(s.Monsters, monster)
	return fmt.Sprintf("%s入侵銀河，正朝 %s 星系前進（預計 %d 回合）",
		st.NameZH, s.starName(target), monster.TransitETA), true
}

// spawnInvadingMonster 讓怪獸落在目前玩家／目前載入熱座帝國的有效殖民星。
// 目標抽樣完成後建立 owner 8 航行 record；抵達前不算盤據該星。
func (s *GameSession) spawnInvadingMonster(kind gamedata.SpaceMonster, minTurn int) (string, bool) {
	if s.Turn-1 < minTurn {
		return "", false
	}
	stars := make([]int, len(s.PlayerColonies))
	for i := range stars {
		stars[i] = s.PlayerColonyStarIndex(i)
	}
	target, ok := s.reservoirMonsterStar(stars, func(i int) bool {
		return i >= 0 && i < len(s.PlayerColonies) && s.PlayerColonies[i].Population > 0
	})
	if !ok {
		return "", false
	}
	return s.placeInvadingMonster(kind, target)
}

func (s *GameSession) spawnInvadingMonsterForAI(kind gamedata.SpaceMonster, aiIndex int) (string, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return "", false
	}
	a := &s.AIPlayers[aiIndex]
	target, ok := s.reservoirMonsterStar(a.ColonyStars, func(i int) bool {
		return i >= 0 && i < len(a.Colonies) && a.Colonies[i].Population > 0
	})
	if !ok {
		return "", false
	}
	return s.placeInvadingMonster(kind, target)
}
