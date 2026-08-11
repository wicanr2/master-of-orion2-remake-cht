package shell

import (
	"fmt"
	"math"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// ai_attack.go:AI 對玩家殖民地發動突襲。
//
// remake 先前的 AI 只會擴張與宣戰,宣完戰之後**什麼也不做**——關係掉到 -40、態勢寫著
// 「戰爭」,玩家卻毫髮無傷。整局遊戲唯一的軍事壓力來自安塔蘭人(週期性腳本),
// 三個 AI 對手實質上是背景裝飾。
//
// 原版有一整條「挑目標 → 派艦隊 → 打」的鏈,三個估值函式都已移植:
//
//	Colony_Worth_To_Player_       @ 0xD2CAE  已殖民星值多少(產出/人口/成長空間/氣候/重力/特殊物產)
//	Enemy_Colony_Worth_To_Player_ @ 0xD8D11  敵方殖民地作為**攻擊目標**值多少
//	Proximity_Worth_To_Player_    @ 0xD2AEA  距離加權
//
// 目標選擇因此是原版的:`AIEnemyColonyValue` 依外交狀態把「這顆星對**主人**有多值錢」與
// 「對**我**有多值錢」加權混合(權重偏向前者——打他最痛的地方,不是搶我最想要的地方),
// 再除以距離。
//
// **「什麼時候打、打贏了會怎樣」這一段仍是 remake 的模型**,原版對應的決策函式尚未反編。
// 下面每個常數都標明了這一點,別把它們當成考據結果。
//
// 設計上刻意保守(這是「讓遊戲有壓力」而不是「讓玩家開局被輾死」):
//   - 前 aiRaidGraceTurns 回合完全不打
//   - 只有態勢已是戰爭、且軍力真的領先的 AI 才會動手
//   - 損失有界:人口不低於 1、BC 不為負、一次最多毀一棟建築
//   - 玩家艦隊在場會實際參戰並可能擊退突襲

const (
	// aiRaidGraceTurns 是開局寬限:這之前 AI 絕不突襲(remake 的平衡值,非原版數字)。
	// 對照安塔蘭人的 antaresStartTurn=20——AI 突襲比安塔蘭人更早開始,因為它是可被外交
	// 與軍備影響的常態壓力,不是終局腳本事件。
	aiRaidGraceTurns = 12
	// aiRaidInterval 是同一個 AI 兩次突襲的最短間隔(回合)。remake 的節奏值,
	// 依 300 回合探針調過:間隔 6 時三個 AI 同時開戰會變成「平均每 3.3 回合被打一次」、
	// 連續兩百多回合,玩起來是磨而不是壓力;10 回合把頻率砍半,單次的痛感不變。
	aiRaidInterval = 10
	// aiRaidStrengthMargin 是發動突襲所需的軍力領先倍率(百分比)。125 = AI 軍力要有
	// 玩家的 1.25 倍才敢動手。remake 的門檻值。
	aiRaidStrengthMargin = 125
	// aiRaidDistanceUnit 與 aiDistanceUnit 同一把尺(見 session.go),把星圖歸一化座標
	// 換算成原版鄰近價值用的距離單位。
	aiRaidDistanceUnit = aiDistanceUnit
)

// AIRaidReport 是一次 AI 突襲的結果(供回合摘要顯示)。
type AIRaidReport struct {
	AIName     string // 發動突襲的 AI 顯示名
	AINameEN   string // 英文模式的 AI 名稱
	StarName   string // 被襲星名
	StarNameEN string // 英文模式的星名
	ColonyIdx  int    // 被襲殖民地索引
	Repelled   bool   // 玩家是否擊退
	PopLost    int    // 人口損失
	BCLost     int    // 國庫損失
	Building   string // 被摧毀的建築(空 = 無)
	FleetLost  int    // 玩家艦隊在防禦中損失的戰力(擊退時 AI 的損失)
	Message    string // 已填好數字的敘述
	MessageEN  string // 同一結果的英文敘述
}

// advanceAIRaids 每回合檢查所有 AI 對手是否對玩家發動突襲,結果寫進 LastRaid/LastRaidReport。
// DisableEvents 時整段停用(與 advanceAntares 一致,供確定性探針/測試使用)。
func (s *GameSession) advanceAIRaids() {
	s.LastRaid = ""
	s.LastRaidReport = nil
	if s.DisableEvents || s.Turn < aiRaidGraceTurns {
		return
	}
	for i := range s.AIPlayers {
		if rep := s.aiRaid(i); rep != nil {
			s.LastRaid = rep.Message
			s.LastRaidReport = rep
			return // 一回合最多一次突襲,避免多 AI 同時開火把玩家瞬間打爆
		}
	}
}

// aiRaidWilling 回傳第 i 個 AI 這回合是否願意且有能力發動突襲。
func (s *GameSession) aiRaidWilling(i int) bool {
	a := &s.AIPlayers[i]
	if a.Treaty.BlocksOffensive() {
		return false // 和平／互不侵犯／同盟均不得對玩家發動攻勢
	}
	if a.StanceName != stanceNames[ai.StanceWar] {
		return false // 只有已經進入戰爭態勢的 AI 才動手
	}
	if s.Turn-a.LastRaidTurn < aiRaidInterval {
		return false
	}
	// 軍力門檻:玩家軍力 0 時也要求 AI 至少有一點戰力,不讓 0 打 0。
	pm := s.playerMilitary()
	if a.FleetStrength <= 0 {
		return false
	}
	if pm > 0 && a.FleetStrength*100 < pm*aiRaidStrengthMargin {
		return false
	}
	// 性格:好戰/冷酷的 AI 更常動手。用「劣勢時的反應強度」表當積極度——
	// 那張表的語意(_personality_losing_ground_chance)最接近「多敢開打」,
	// 但**用在這個判斷點是 remake 的選擇**,原版在哪裡讀它還沒反編確認。
	chance := ai.PersonalityLosingGroundChance(a.Personality)
	if chance <= 0 {
		return false // 和平主義(0)從不主動突襲
	}
	return (s.Turn*7+i*13)%100 < chance // 確定性擬亂數,保持存檔/探針可重現
}

// aiRaidTarget 挑第 i 個 AI 最想打的玩家殖民地,回傳其索引;沒有可打的回 -1。
//
// 三層都用原版公式:
//  1. `AIColonyValue`(Colony_Worth_To_Player_)算這個殖民地本身值多少——分別站在
//     「主人(玩家)」與「攻擊者(AI)」兩個立場各算一次。
//  2. `AIEnemyColonyValue`(Enemy_Colony_Worth_To_Player_)依外交狀態把兩者加權混合。
//  3. 除以距離(Proximity_Worth_To_Player_ 的倒數加權方向)。
//
// ⚠ 立場差異在 remake 只反映在「目標傾向」這一項:AI 用自己性格對應的 AIObjective,
// 玩家用中性的 BalancedLow(remake 的玩家沒有 AI 目標傾向這個欄位)。原版兩個立場還會
// 差在種族人口上限與重力天賦,remake 沒有那兩層,故兩個估值目前只有權重不同。
func (s *GameSession) aiRaidTarget(i int) int {
	a := &s.AIPlayers[i]
	obj := aiObjectiveFor(a.Personality)
	policy := aiForeignPolicyFor(a)
	best, bestVal := -1, 0
	for ci := range s.PlayerColonies {
		star := s.PlayerColonyStarIndex(ci)
		ownerVal := s.aiColonyValue(ci, gamedata.AIObjectiveBalancedLow) // 玩家(主人)立場
		selfVal := s.aiColonyValue(ci, obj)                              // AI(攻擊者)立場
		v := gamedata.AIEnemyColonyValue(ownerVal, selfVal, policy, false)
		if v <= 0 {
			continue
		}
		// 距離折算:取該 AI 任一據點到目標星的最短距離。
		d := s.aiNearestOwnedDistance(i, star)
		if d < 1 {
			d = 1
		}
		v = v * gamedata.AIProximityOwnWeight / d
		if v > bestVal {
			best, bestVal = ci, v
		}
	}
	return best
}

// aiForeignPolicyFor 把 remake 的 AI 態勢對映到原版的 ForeignPolicy 編碼。
//
// remake 的態勢是 ai.DecideStance 由關係分數推得的五級(戰爭/敵視/中立/提議貿易/提議結盟),
// 原版是 ForeignPolicy 六級(無/互不侵犯/同盟/和平/有限戰爭/戰爭)。兩套不是同一個維度,
// 這裡取語意最接近的對映——**對映本身是 remake 的**,原版的態勢↔外交狀態關係另有機制。
func aiForeignPolicyFor(a *AIOpponent) gamedata.AIForeignPolicy {
	if a != nil {
		switch a.Treaty.FormalPolicy {
		case gamedata.DIPLO_NON_AGGRESSION:
			return gamedata.DiploNonAggression
		case gamedata.DIPLO_ALLIANCE:
			return gamedata.DiploAlliance
		case gamedata.DIPLO_PEACE:
			return gamedata.DiploPeace
		}
	}
	switch a.StanceName {
	case stanceNames[ai.StanceWar]:
		// 關係已經觸底(-30 以下)視為原版那一檔更極端的狀態:目標估值完全站在受害者立場。
		if a.Relation <= -30 {
			return gamedata.DiploTotalWar
		}
		return gamedata.DiploLimitedWar
	case stanceNames[ai.StanceProposeAlliance]:
		return gamedata.DiploAlliance
	case stanceNames[ai.StanceProposeTrade]:
		return gamedata.DiploPeace
	case stanceNames[ai.StanceHostile]:
		return gamedata.DiploNonAggression
	default:
		return gamedata.DiploNone
	}
}

// aiColonyValue 算玩家第 ci 個殖民地對 AI 的價值(原版 Colony_Worth_To_Player_)。
func (s *GameSession) aiColonyValue(ci int, obj gamedata.AIObjective) int {
	if ci < 0 || ci >= len(s.PlayerColonies) {
		return 0
	}
	c := s.PlayerColonies[ci]
	in := gamedata.AIColonyValueInput{
		Population: c.Population,
		// 原版取「主人的人口上限」與「評估者的人口上限」的平均;remake 沒有逐種族的人口上限
		// 模型(PopMax 已含該殖民地所有加成),兩者同值 → 平均就是 PopMax 本身。
		MaxPop:   c.PopMax,
		Food:     c.Farmers * c.FoodPerFarmer,
		Industry: c.Workers * c.IndustryPerWorker,
		Research: c.Scientists * c.ResearchPerScientist,
		Climate:  c.Climate,
		Gravity:  c.PlanetGravity,
	}
	if p := s.ColonyPlanet(ci); p != nil {
		in.Special = int(p.SpecialID)
	}
	return gamedata.AIColonyValue(in, obj)
}

// aiNearestOwnedDistance 回傳第 i 個 AI 的據點到 starIdx 的最短距離(aiRaidDistanceUnit 尺度)。
// AI 沒有任何據點時回一個大數(等於「打不到」)。
func (s *GameSession) aiNearestOwnedDistance(i, starIdx int) int {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return 1 << 20
	}
	target := s.Stars[starIdx]
	best := math.MaxFloat64
	for _, st := range s.AIPlayers[i].ColonyStars {
		if st < 0 || st >= len(s.Stars) {
			continue
		}
		if d := math.Hypot(s.Stars[st].X-target.X, s.Stars[st].Y-target.Y); d < best {
			best = d
		}
	}
	if best == math.MaxFloat64 {
		return 1 << 20
	}
	return int(best*aiRaidDistanceUnit) + 1
}

// aiObjectiveFor 依性格挑 AI 的目標傾向。與 aiPlanetValue 用的是同一個映射,
// 集中在這裡供兩處共用(原版目標傾向是獨立於性格的另一個維度,見 aiPlanetValue 註解)。
func aiObjectiveFor(p ai.Personality) gamedata.AIObjective {
	switch p {
	case ai.PersonalityRuthless, ai.PersonalityXenophobic:
		return gamedata.AIObjectiveMineral
	case ai.PersonalityPacifist:
		return gamedata.AIObjectivePopulation
	case ai.PersonalityAggressive:
		return gamedata.AIObjectiveBalancedHigh
	default:
		return gamedata.AIObjectiveBalancedLow
	}
}

// aiRaid 讓第 i 個 AI 對玩家最有價值的殖民地發動一次突襲。不動手時回 nil。
//
// 解算模型(remake 的,見檔頭):玩家的防禦力 = 艦隊戰力 + 該殖民地駐軍/坦克折算 +
// 星基類建築。防禦 >= 攻擊 → 擊退(AI 損失軍力);否則依戰力差造成人口/國庫/建築損失。
func (s *GameSession) aiRaid(i int) *AIRaidReport {
	if i < 0 || i >= len(s.AIPlayers) {
		return nil
	}
	// ⚠ 2026-08-08(第 47 項(AI艦隊移動))突襲的前提從「想打」改成「**打得到**」。
	//
	// 先前這裡是 `aiRaidWilling(i)` + `aiRaidTarget(i)`——想打誰就直接結算誰,
	// AI 艦隊憑空出現在目標上空。現在 AI 有位置了(見 ai_fleet.go),
	// 條件是「艦隊靜止,而且就停在某個玩家殖民地的星上」。
	//
	// 「要不要出兵」與「打誰」那兩個判斷沒有變,只是搬到了出發那一刻
	// (`aiLaunchRaidFleet`),中間隔著一段航程——**玩家因此看得到它來**。
	ci, ok := s.aiFleetAtPlayerColony(i)
	if !ok {
		return nil
	}
	a := &s.AIPlayers[i]
	// ⚠ 間隔守門**留在這裡**,不能跟著其他條件一起搬到出發那一刻。
	//
	// 其餘條件(戰爭態勢/軍力領先/性格)是「要不要出兵」,判一次就夠;
	// 間隔是「多久能打一次」,而艦隊抵達之後**會一直停在那裡**——少了這道守門,
	// 一支停在玩家殖民地上空的 AI 艦隊會每一回合都結算一次突襲。
	if s.Turn-a.LastRaidTurn < aiRaidInterval {
		return nil
	}
	a.LastRaidTurn = s.Turn

	starName := "未知星系"
	starNameEN := starName
	if star := s.PlayerColonyStarIndex(ci); star >= 0 && star < len(s.Stars) {
		starName = s.Stars[star].Name
		starNameEN = s.Stars[star].NameEN
		if starNameEN == "" {
			starNameEN = starName
		}
	}
	aiNameEN := a.Name
	if raceIdx := aiRaceIndex(*a); raceIdx >= 0 && raceIdx < len(Races) {
		aiNameEN = Races[raceIdx].EnName
	}
	rep := &AIRaidReport{AIName: a.Name, AINameEN: aiNameEN, StarName: starName,
		StarNameEN: starNameEN, ColonyIdx: ci}

	attack := a.FleetStrength
	defense := s.colonyDefense(ci)

	if defense >= attack {
		// 擊退:AI 損失部分軍力(比例取戰力差,夾在 10%~50%,remake 的模型)。
		loss := a.FleetStrength * 25 / 100
		if loss < 1 {
			loss = 1
		}
		a.FleetStrength -= loss
		if a.FleetStrength < 0 {
			a.FleetStrength = 0
		}
		rep.Repelled = true
		rep.FleetLost = loss
		rep.Message = fmt.Sprintf("⚔ %s 突襲 %s,遭防禦部隊擊退,對方損失 %d 戰力", a.Name, starName, loss)
		rep.MessageEN = fmt.Sprintf("⚔ %s raided %s but was repelled by the defense forces, losing %d fleet strength",
			aiNameEN, starNameEN, loss)
		s.LastRaid = rep.Message
		return rep
	}

	// 突破:戰力差越大損失越重(每超出 100% 防禦力損失 1 人口,上限 3)。
	margin := attack - defense
	// 就算沒擋下來,防禦力仍會對攻方造成消耗(打掉多少防禦,自己就折損多少的一半)。
	// 沒有這一段的話,「防禦蓋一半」等於完全白蓋——擋不住就毫無回報,玩家只能全有或全無。
	if defense > 0 {
		attrition := defense / 2
		if attrition > a.FleetStrength {
			attrition = a.FleetStrength
		}
		a.FleetStrength -= attrition
		rep.FleetLost = attrition
	}
	rep.PopLost = margin/100 + 1
	if rep.PopLost > 3 {
		rep.PopLost = 3
	}
	c := &s.PlayerColonies[ci]
	if rep.PopLost > c.Population-1 {
		rep.PopLost = c.Population - 1 // 突襲不會滅殖民地(那是地面入侵的事)
	}
	if rep.PopLost < 0 {
		rep.PopLost = 0
	}
	for k := 0; k < rep.PopLost; k++ {
		c.Population--
		switch {
		case c.Workers > 0:
			c.Workers--
		case c.Farmers > 0:
			c.Farmers--
		case c.Scientists > 0:
			c.Scientists--
		}
	}

	// 掠奪國庫:每點戰力差 1 BC,上限 60,且不讓 BC 變負。
	rep.BCLost = margin
	if rep.BCLost > 60 {
		rep.BCLost = 60
	}
	if rep.BCLost > s.Player.BC {
		rep.BCLost = s.Player.BC
	}
	if rep.BCLost < 0 {
		rep.BCLost = 0
	}
	s.Player.BC -= rep.BCLost

	// 摧毀一棟建築(戰力差 >= 200 才會發生)。
	if margin >= 200 {
		rep.Building = s.destroyColonyBuilding(ci)
	}

	parts := fmt.Sprintf("人口 -%d、國庫 -%d BC", rep.PopLost, rep.BCLost)
	if rep.Building != "" {
		parts += "、" + rep.Building + "被摧毀"
	}
	if rep.FleetLost > 0 {
		parts += fmt.Sprintf(";防禦部隊使對方折損 %d 戰力", rep.FleetLost)
	}
	rep.Message = fmt.Sprintf("⚔ %s 突襲 %s:%s", a.Name, starName, parts)
	partsEN := fmt.Sprintf("Population -%d, treasury -%d BC", rep.PopLost, rep.BCLost)
	if rep.Building != "" {
		partsEN += fmt.Sprintf("; %s was destroyed", rep.Building)
	}
	if rep.FleetLost > 0 {
		partsEN += fmt.Sprintf("; the defending force cost them %d fleet strength", rep.FleetLost)
	}
	rep.MessageEN = fmt.Sprintf("⚔ %s raided %s: %s", aiNameEN, starNameEN, partsEN)
	s.LastRaid = rep.Message
	return rep
}

// colonyDefense 回傳玩家第 ci 個殖民地的防禦力(remake 的模型,見檔頭)。
// 艦隊只在停泊於該星時計入——把艦隊擺對地方是玩家的決策,不是自動全域護盾。
func (s *GameSession) colonyDefense(ci int) int {
	def := 0
	star := s.PlayerColonyStarIndex(ci)
	if star >= 0 && s.Fleet().AtStar == star && s.Fleet().ETA == 0 {
		def += s.playerMilitary()
	}
	// 駐軍與坦克:折算成戰力(gamedata 無「陸戰隊 → 太空防禦」的換算,這是 remake 的簡化,
	// 語意是「地面部隊配合軌道防禦拖住突襲」)。
	if ci < len(s.PlayerColonyMarines) {
		def += s.PlayerColonyMarines[ci] * 2
	}
	if ci < len(s.PlayerColonyTanks) {
		def += s.PlayerColonyTanks[ci] * 3
	}
	// 防禦建築:與軌道轟炸的反擊**共用同一套推導**(`retaliationAttackers`)。
	//
	// ⚠ 這裡原本是 `CommandPointsFromBuildings(...) * 10` —— 一個自編的係數,而且只認
	// 星基/戰鬥星/星辰要塞三級,**飛彈基地與地面砲台完全不算**。那不是「模型還沒建好」:
	// `gamedata/satellite.go` 的 space 預算模型早就存在(飛彈基地 300 / 地面砲台 450 都是
	// 手冊 p.78 / p.81 的確認值),只是當初沒接到這條路徑上。
	//
	// 改用 `retaliationAttackers` 之後三件事一起對上:
	//   ① 反擊戰力隨**已解鎖的武器科技**成長,不再是寫死的 10/20/30
	//   ② 飛彈基地與地面砲台真的有用了
	//   ③ 1.3/1.5 的 beam arc-cost 差異(`RuleProfile`)自動吃到,不必再各寫一份
	if ci < len(s.ColonyBuildings) {
		for _, c := range retaliationAttackers(s.ColonyBuildings[ci], s.Player, s.RuleProfile) {
			def += c.atk
		}
		// 恆星轉換器(行星版)2026-08-07 起由 retaliationAttackers 一併回傳
		// ——先前它只算在這裡,結果同一棟建築擋得住 AI 來襲卻對軌道轟炸不反擊。
		// 統一到那一支之後這裡不再另外加,否則會雙重計算。
	}
	return def
}

// destroyColonyBuilding 摧毀第 ci 個殖民地的一棟已建建築(依名稱排序取第一棟,保持確定性),
// 回傳被摧毀的建築名;沒有建築可毀回空字串。
func (s *GameSession) destroyColonyBuilding(ci int) string {
	if ci < 0 || ci >= len(s.ColonyBuildings) || len(s.ColonyBuildings[ci]) == 0 {
		return ""
	}
	names := make([]string, 0, len(s.ColonyBuildings[ci]))
	for name, built := range s.ColonyBuildings[ci] {
		if built {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return ""
	}
	sortStrings(names)
	delete(s.ColonyBuildings[ci], names[0])
	return names[0]
}
