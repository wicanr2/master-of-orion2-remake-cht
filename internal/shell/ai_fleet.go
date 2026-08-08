package shell

import (
	"math"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// ai_fleet.go:**AI 的艦隊會在星圖上移動**。
//
// ============ 先前的 AI 突襲是瞬移的 ============
//
// `ai_attack.go` 的突襲鏈是「挑目標 → 直接結算」——AI 沒有位置,艦隊憑空出現在玩家的
// 殖民地上空。這造成三件事:
//
//  1. **玩家看不到它來。** 突襲無法被預警、無法被攔截,只能事後看回合摘要。
//  2. **阿提米絲系統網打不到 AI。** 那棟建築的手冊效果是「任何**進入**該星系的敵艦」,
//     而 AI 從來沒有「進入」這個動作——它直接就在那裡了。這條缺口從第 38 項(行星護盾等三棟)記到現在。
//  3. AI 艦隊的移動不吃星雲/黑洞/干擾場,因為它根本不移動。
//
// 這一檔給 AI 一個**位置**和一條**航線**,突襲改成「抵達目標才打」。
//
// ============ 誠實留白 ============
//
//   - **一個 AI 只有一支艦隊。** 玩家可以分/合多支(第 19 項(AI請求會談)),AI 沒有——它的軍力
//     是單一的 `FleetStrength` 整數,不是艦艇清單。要給 AI 多艦隊得先給它艦艇模型。
//   - **AI 艦隊沒有逐艦資料**,所以:艦員經驗(第 39 項(艦員經驗))拿不到、水雷的逐艦觸發率
//     (依艦體等級 20–100%)也套不上——水雷對 AI 改成對**艦隊戰力**的整體折損,
//     見 `applyArtemisMinesToAIFleet` 的說明。**這是近似,不是原版行為。**
//   - **航線不判星雲/黑洞/干擾場。** 玩家那條路徑模型(第 16/17 項)吃這些懲罰,
//     AI 目前只算直線秒差距 ÷ 速度。要接得先讓 AI 走同一套 `fleetSpeedForTrip`,
//     而那支函式綁在 `s.Player` 上。
//   - **只打玩家。** AI 之間不互相出兵——那需要 AI-vs-AI 的戰鬥解算,remake 沒有。

// aiFleetStar 回傳這個 AI 的主力艦隊目前所在的星。
//
// `FleetPosSet` 為 false 時(新對局尚未初始化、或舊存檔)退回它的母星
// ——**不能直接讀 FleetStar**,零值 0 是合法的星索引。
func aiFleetStar(a AIOpponent) int {
	if a.FleetPosSet {
		return a.FleetStar
	}
	if len(a.ColonyStars) > 0 {
		return a.ColonyStars[0]
	}
	return -1
}

// aiFleetSpeedParsecs 回傳這個 AI 的艦隊每回合秒差距。
//
// 與玩家的 `FleetSpeedParsecs` 同一套引擎階查表,只是讀的是 AI 自己的科技狀態。
// 查不到任何引擎時退回**核融引擎**(階 1)——理由與玩家那邊相同:手冊寫得很清楚
// 「Nuclear Drive … is the slowest of the FTL propulsion systems」,能出星系就至少有它。
// 少了這個下界,AI 的航速會是 0,ETA 被夾成 1,整個移動模型形同虛設而且畫面上看不出來。
func aiFleetSpeedParsecs(a AIOpponent) int {
	tier := gamedata.DriveTierFromTechs(func(topic gamedata.ResearchTopic, tech gamedata.Technology) bool {
		return driveTechOwned(a.Player, topic, tech)
	})
	if tier < 1 {
		tier = 1
	}
	sp := gamedata.FleetSpeedForDrive(tier)
	if sp < 1 {
		sp = 1
	}
	return sp
}

// aiFleetETATo 回傳 AI 艦隊從 from 飛到 to 要幾回合(至少 1)。
//
// 只算直線秒差距 ÷ 速度,不含星雲/黑洞/干擾場懲罰(見檔頭留白)。
func (s *GameSession) aiFleetETATo(a AIOpponent, from, to int) int {
	if from < 0 || to < 0 || from >= len(s.Stars) || to >= len(s.Stars) {
		return 0
	}
	if from == to {
		return 0
	}
	sp := aiFleetSpeedParsecs(a)
	pc := s.ParsecsBetweenStars(from, to)
	eta := int(math.Ceil(float64(pc) / float64(sp)))
	if eta < 1 {
		eta = 1
	}
	return eta
}

// AIFleetArrival 是一支 AI 艦隊抵達某顆星的紀錄(供回合摘要 / 水雷回報)。
type AIFleetArrival struct {
	AIName   string
	StarName string
	StarIdx  int
	Mines    *ArtemisStrike // 挨了水雷時才有
}

// advanceAIFleets 推進所有 AI 艦隊的航行,回傳這一回合抵達的那幾支。
//
// 順序:先讓在途的往前走一格,再讓閒置的決定要不要出兵。**同一回合不會又出發又抵達**
// ——出發那一回合 ETA 至少是 1,要下一回合才會遞減到 0。
func (s *GameSession) advanceAIFleets() []AIFleetArrival {
	var arrivals []AIFleetArrival
	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		if !a.FleetPosSet {
			// 第一次跑到這裡就把位置定在母星(新對局與舊存檔共用這條初始化)。
			if home := aiFleetStar(*a); home >= 0 {
				a.FleetStar, a.FleetPosSet = home, true
			}
		}
		if a.FleetETA > 0 {
			a.FleetETA--
			if a.FleetETA > 0 {
				continue // 還在路上
			}
			a.FleetStar = a.FleetDestStar
			a.FleetDestStar = -1
			arr := AIFleetArrival{AIName: a.Name, StarName: s.starName(a.FleetStar), StarIdx: a.FleetStar}
			// 阿提米絲系統網:手冊「any enemy ship **entering** that system」
			// ——進門的那一刻結算,與玩家那條同一個時點(見 session.go advanceFleet)。
			arr.Mines = s.applyArtemisMinesToAIFleet(i, a.FleetStar)
			arrivals = append(arrivals, arr)
			continue
		}
		s.aiLaunchRaidFleet(i)
	}
	return arrivals
}

// aiLaunchRaidFleet 讓閒置的 AI 艦隊決定要不要朝某個玩家殖民地出發。
//
// 出兵條件沿用 `aiRaidWilling`(戰爭態勢 + 軍力領先 + 最短間隔 + 性格積極度),
// 目標沿用 `aiRaidTarget`(原版的三層估值)——**這一輪沒有新增任何決策規則**,
// 只是把「決定打誰」與「實際打到」之間插進了一段航程。
func (s *GameSession) aiLaunchRaidFleet(i int) {
	if s.DisableEvents || s.Turn < aiRaidGraceTurns {
		return
	}
	a := &s.AIPlayers[i]
	if !s.aiRaidWilling(i) {
		return
	}
	ci := s.aiRaidTarget(i)
	if ci < 0 {
		return
	}
	dest := s.PlayerColonyStarIndex(ci)
	from := aiFleetStar(*a)
	if dest < 0 || from < 0 {
		return
	}
	if dest == from {
		return // 已經在目標上空,交給 advanceAIRaids 結算
	}
	eta := s.aiFleetETATo(*a, from, dest)
	if eta <= 0 {
		return
	}
	a.FleetDestStar = dest
	a.FleetETA = eta
}

// aiFleetAtPlayerColony 回傳這個 AI 的艦隊是否**停在**某個玩家殖民地上空
// (靜止且位置吻合)。突襲的前提從「想打」變成「打得到」。
func (s *GameSession) aiFleetAtPlayerColony(i int) (colonyIdx int, ok bool) {
	a := s.AIPlayers[i]
	if a.FleetETA > 0 {
		return -1, false // 還在路上
	}
	star := aiFleetStar(a)
	if star < 0 {
		return -1, false
	}
	for ci := range s.PlayerColonies {
		if s.PlayerColonyStarIndex(ci) == star {
			return ci, true
		}
	}
	return -1, false
}

// applyArtemisMinesToAIFleet 對剛抵達的 AI 艦隊結算阿提米絲水雷網。
//
// 這條缺口從第 38 項(行星護盾等三棟)記到現在:「⚠ 只對玩家艦隊生效,AI 無艦隊移動模型」。
// 現在 AI 會移動了,所以雷區對它生效。
//
// ⚠ **不是原版行為的逐艦模擬。** 原版是逐艦擲觸發率(依艦體等級 20–100%)、
// 逐艦扣血;AI 在 remake 沒有艦艇清單,只有一個 `FleetStrength` 整數。所以這裡改成:
//
//	雷數 = ArtemisMineCount(擲骰)           ← 與玩家那條同一支函式,同一個範圍
//	每雷傷害 = ArtemisMineDamage(護盾等級 0)  ← AI 沒有護盾資料,一律當無護盾(對 AI 不利)
//	戰力折損 = 雷數 × 每雷傷害 / aiMineStrengthDivisor
//
// 除數把「傷害點數」換算成「艦隊戰力」——那是兩把不同的尺,而 remake 沒有換算依據。
// 取 10 是**刻意保守的建模選擇**(讓水雷對 AI 有感但不會一次抹平一支艦隊),
// 不是考據值。寫在這裡,不假裝它是真值。
func (s *GameSession) applyArtemisMinesToAIFleet(i, starIdx int) *ArtemisStrike {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return nil
	}
	ci := -1
	for k := range s.PlayerColonies {
		if s.PlayerColonyStarIndex(k) == starIdx {
			ci = k
			break
		}
	}
	if ci < 0 || !s.buildingsFor(ci)[artemisBuildingName] {
		return nil
	}
	a := &s.AIPlayers[i]
	if a.FleetStrength <= 0 {
		return nil
	}
	// ArtemisMineCount 收的是 0..span 的**偏移**(不是 1..n 的擲骰),
	// 所以 eventRoll(span+1) 回 1..span+1 之後要減 1。
	span := gamedata.ArtemisMinesMax - gamedata.ArtemisMinesMin
	mines := gamedata.ArtemisMineCount(s.eventRoll(span+1) - 1)
	dmg := gamedata.ArtemisMineDamage(0) // AI 無護盾資料,一律當無護盾
	lost := mines * dmg / aiMineStrengthDivisor
	if lost < 1 {
		lost = 1 // 進了雷區總要付出點什麼
	}
	if lost > a.FleetStrength {
		lost = a.FleetStrength
	}
	a.FleetStrength -= lost
	return &ArtemisStrike{
		StarName:    s.starName(starIdx),
		ShipsHit:    1, // AI 只有「一支艦隊」這個粒度
		TotalDamage: lost,
	}
}

// aiMineStrengthDivisor 把水雷的傷害點數折算成 AI 的艦隊戰力損失。
//
// **這不是考據值**,是 remake 的建模選擇(見 applyArtemisMinesToAIFleet)——
// 兩把尺之間沒有原版依據可循,取 10 讓一次雷區大約削掉一支中等艦隊的個位數戰力。
const aiMineStrengthDivisor = 10
