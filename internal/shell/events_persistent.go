package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// events_persistent.go:**持續型**隨機事件——那些不是「發生一次扣點東西」而是「進入一個
// 會持續好幾回合的狀態,直到某個條件解除」的事件。
//
// 這一類先前全部卡在「缺前置子系統」:remake 只有「單次結算」的事件模型,沒有任何跨回合
// 的事件狀態。這個檔案補上那個模型,並把手冊 p.180-181 給的數字逐條落地——**那兩頁是逐事件
// 的規格書**,連回合門檻與機率都寫死了:
//
//	超新星     "This won't happen until at least 200 turns have passed … All Research Points
//	           the system generates are applied toward finding a solution. The supernova
//	           countdown takes from six to fourteen turns."
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
// ⚠ 手冊沒給的部分(標明,不臆造):每回合的觸發機率(除了「六回合後 5%」那兩條)、
// 怪獸移動到目標的速度、超新星「需要多少額外研究」的確切公式(手冊只說「倒數越短需要越多」)。

const (
	// 手冊 p.181 逐字:持續型狀態「六回合之後,每回合 5% 機率結束」。
	persistentEventMinTurns   = 6
	persistentEventEndPercent = 5

	// 超新星倒數 6-14 回合(手冊 p.181 逐字)。
	supernovaMinCountdown = 6
	supernovaMaxCountdown = 14
)

// PersistentEventKind 是持續型事件的種類。
type PersistentEventKind int

const (
	PersistentNone PersistentEventKind = iota
	// PersistentSupernova 超新星:某星系的恆星即將爆炸,靠研究點數搶救。
	PersistentSupernova
	// PersistentStasis 時空異象:某星系的殖民地凍結(不產出、不成長,也不吃食物、不繳維護費)。
	PersistentStasis
	// PersistentWarpBeast 超空間獸:航行中的艦隊有機率被拖走一艘船。
	PersistentWarpBeast
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
}

// advancePersistentEvents 每回合推進所有進行中的持續型事件,回傳要顯示的訊息。
func (s *GameSession) advancePersistentEvents() []string {
	if len(s.PersistentEvents) == 0 {
		return nil
	}
	var msgs []string
	kept := s.PersistentEvents[:0]
	for i := range s.PersistentEvents {
		e := &s.PersistentEvents[i]
		e.Turns++
		done, msg := s.stepPersistentEvent(e)
		if msg != "" {
			msgs = append(msgs, msg)
		}
		if !done {
			kept = append(kept, *e)
		}
	}
	s.PersistentEvents = kept
	return msgs
}

// stepPersistentEvent 推進單一持續型事件一回合,回傳 (是否結束, 訊息)。
func (s *GameSession) stepPersistentEvent(e *PersistentEvent) (bool, string) {
	switch e.Kind {
	case PersistentSupernova:
		return s.stepSupernova(e)
	case PersistentStasis:
		// 凍結本身的效果在 EndTurn 的產出結算裡(見 StarInStasis);這裡只管什麼時候結束。
		if s.persistentEventEnds(e) {
			return true, fmt.Sprintf("%s 星系的時空異象消散,殖民地恢復運作", s.starName(e.StarIndex))
		}
		return false, ""
	case PersistentWarpBeast:
		msg := s.warpBeastStrike()
		if s.persistentEventEnds(e) {
			end := "超空間獸遁入異次元,航道恢復安全"
			if msg != "" {
				return true, msg + ";" + end
			}
			return true, end
		}
		return false, msg
	}
	return true, ""
}

// persistentEventEnds 依手冊 p.181「六回合之後每回合 5% 機率結束」判定。
func (s *GameSession) persistentEventEnds(e *PersistentEvent) bool {
	if e.Turns < persistentEventMinTurns {
		return false
	}
	return s.eventRoll(100) <= persistentEventEndPercent
}

// stepSupernova 推進超新星倒數:系統的研究點投入搶救,倒數歸零時看是否救得回來。
//
// 手冊 p.181:「All Research Points the system generates are applied toward finding a
// solution… if the emperor doesn't accelerate the colony's research efforts, the colonies
// will discover the solution one turn too late.」——也就是**預設剛好差一點**,玩家必須
// 額外投入研究才救得回來。remake 據此把 ResearchNeeded 設成「該系統自然產出 × (倒數+1)」,
// 讓「什麼都不做就是差一回合」這個手冊描述的張力成立。
func (s *GameSession) stepSupernova(e *PersistentEvent) (bool, string) {
	// 該星系殖民地本回合的研究產出全部投入搶救。
	gained := 0
	for i, star := range s.PlayerColonyStars {
		if star != e.StarIndex || i >= len(s.LastPlayerOutput.Colonies) {
			continue
		}
		gained += s.LastPlayerOutput.Colonies[i].Research
	}
	e.ResearchDone += gained
	e.Countdown--

	if e.ResearchDone >= e.ResearchNeeded {
		return true, fmt.Sprintf("%s 星系的科學家搶在爆炸前穩定了恆星核心,超新星危機解除",
			s.starName(e.StarIndex))
	}
	if e.Countdown > 0 {
		return false, fmt.Sprintf("%s 星系恆星不穩定,倒數 %d 回合(搶救進度 %d/%d)",
			s.starName(e.StarIndex), e.Countdown, e.ResearchDone, e.ResearchNeeded)
	}

	// 倒數歸零且沒救回來:手冊「all of the system's inhabitants are killed and all colonies
	// are destroyed. All planets in the system become Radiated.」
	lost := s.destroyColoniesAtStar(e.StarIndex)
	if e.StarIndex >= 0 && e.StarIndex < len(s.Planets) {
		if p := s.PlanetOf(e.StarIndex); p != nil {
			p.ClimateID = gamedata.RADIATED
			p.Climate = climateDisplayName(gamedata.RADIATED)
		}
	}
	return true, fmt.Sprintf("%s 的恆星爆發為超新星,%d 座殖民地全滅,整個星系化為輻射廢土",
		s.starName(e.StarIndex), lost)
}

// warpBeastStrike 讓超空間獸對「正在航行的艦隊」出手:有機率拖走一艘船。
// 手冊只說「there is a random chance」沒給數字——這個機率是 remake 值。
func (s *GameSession) warpBeastStrike() string {
	if s.Fleet().ETA <= 0 || len(s.Fleet().Ships) == 0 {
		return "" // 沒有艦隊在星際航行中,野獸這回合抓不到東西
	}
	const warpBeastStrikePercent = 20 // ⚠ remake 值:手冊只說「random chance」
	if s.eventRoll(100) > warpBeastStrikePercent {
		return ""
	}
	lost := s.Fleet().Ships[len(s.Fleet().Ships)-1].Name
	s.removeWeakestShip()
	if lost == "" {
		lost = "一艘艦艇"
	}
	return fmt.Sprintf("航行中的艦隊被超空間獸扯進異次元,失去了「%s」", lost)
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

// ColonyInStasis 回傳玩家第 i 個殖民地是否被時空異象凍結。
func (s *GameSession) ColonyInStasis(i int) bool {
	star := s.PlayerColonyStarIndex(i)
	if star < 0 {
		return false
	}
	return s.StarInStasis(star)
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

// destroyColoniesAtStar 摧毀該星的所有玩家殖民地,回傳摧毀數量。
func (s *GameSession) destroyColoniesAtStar(starIdx int) int {
	n := 0
	for i := len(s.PlayerColonies) - 1; i >= 0; i-- {
		if s.PlayerColonyStarIndex(i) != starIdx {
			continue
		}
		s.removePlayerColony(i)
		n++
	}
	if starIdx >= 0 && starIdx < len(s.Stars) && n > 0 {
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
}

// starName 回傳星名(越界回「未知星系」)。
func (s *GameSession) starName(idx int) string {
	if idx < 0 || idx >= len(s.Stars) {
		return "未知星系"
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

// startSupernova 在某顆有玩家殖民地的星啟動超新星倒數。
// 手冊 p.181:200 回合之後才會發生。
func (s *GameSession) startSupernova() (string, bool) {
	if s.Turn < 200 {
		return "", false
	}
	i, ok := s.pickColony()
	if !ok {
		return "", false
	}
	star := s.PlayerColonyStarIndex(i)
	if star < 0 || s.hasPersistentEvent(PersistentSupernova, star) {
		return "", false
	}
	countdown := supernovaMinCountdown + s.eventRoll(supernovaMaxCountdown-supernovaMinCountdown+1) - 1

	// 需要的研究量:該系統目前的自然產出 × (倒數+1),讓「什麼都不做就差一回合」成立
	// (手冊描述的張力,見 stepSupernova 註解)。至少 1,避免零產出殖民地變成免費過關。
	perTurn := 0
	for j, st := range s.PlayerColonyStars {
		if st == star && j < len(s.LastPlayerOutput.Colonies) {
			perTurn += s.LastPlayerOutput.Colonies[j].Research
		}
	}
	need := perTurn * (countdown + 1)
	if need < 1 {
		need = 1
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{
		Kind: PersistentSupernova, StarIndex: star, Countdown: countdown, ResearchNeeded: need,
	})
	return fmt.Sprintf("%s 的恆星進入不穩定狀態,將在 %d 回合後爆發為超新星——全系統研究能量已投入搶救(需 %d RP)",
		s.starName(star), countdown, need), true
}

// startStasis 讓某顆有玩家殖民地的星進入時空異象凍結。
func (s *GameSession) startStasis() (string, bool) {
	i, ok := s.pickColony()
	if !ok {
		return "", false
	}
	star := s.PlayerColonyStarIndex(i)
	if star < 0 || s.hasPersistentEvent(PersistentStasis, star) {
		return "", false
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{Kind: PersistentStasis, StarIndex: star})
	return fmt.Sprintf("%s 星系陷入時空異象,當地殖民地凍結——不生產、不成長,但也不需要食物與維護費",
		s.starName(star)), true
}

// startWarpBeast 讓超空間獸進入銀河(全域效果,StarIndex = -1)。
func (s *GameSession) startWarpBeast() (string, bool) {
	if s.hasPersistentEvent(PersistentWarpBeast, -1) {
		return "", false
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{Kind: PersistentWarpBeast, StarIndex: -1})
	return "一頭半身沉在超空間裡的野獸開始在航道上遊蕩,星際航行變得危險", true
}

// applyWormhole 蟲洞(手冊 p.181:「moves that fleet to their destination in a single turn」)。
func (s *GameSession) applyWormhole() (string, bool) {
	if s.Fleet().ETA <= 1 || s.Fleet().DestStar < 0 {
		return "", false // 沒有正在長途航行的艦隊,這個好事無處可用
	}
	was := s.Fleet().ETA
	s.Fleet().ETA = 1
	return fmt.Sprintf("一端蟲洞突然出現在艦隊航道上,原本還要 %d 回合的航程縮短為 1 回合", was), true
}

// spawnInvadingMonster 讓一頭怪獸入侵銀河(事件 19-23)。
// 手冊 p.180-181 給了每種怪獸的**最早回合**,逐條照抄。
func (s *GameSession) spawnInvadingMonster(kind gamedata.SpaceMonster, minTurn int) (string, bool) {
	if s.Turn < minTurn {
		return "", false
	}
	st, ok := gamedata.MonsterStatsFor(kind)
	if !ok {
		return "", false
	}
	// 挑一顆玩家看得到、還沒有怪獸的星(優先無主星:怪獸「入侵」而不是憑空出現在殖民地上)。
	target := -1
	for i := range s.Stars {
		if p := s.PlanetOf(i); s.StarGuardedByMonster(i) || (p != nil && p.NoPlanet) {
			continue
		}
		if s.Stars[i].Owner == 0 {
			target = i
			break
		}
	}
	if target < 0 {
		return "", false
	}
	s.Monsters = append(s.Monsters, MonsterGuard{StarIndex: target, Kind: kind, Structure: st.Structure})
	return fmt.Sprintf("%s入侵銀河,盤據 %s 星系——不清除就無法在該系統落腳",
		st.NameZH, s.starName(target)), true
}
