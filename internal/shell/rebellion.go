package shell

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// rebellion.go:**被征服人口的叛亂**——同化計時器的另一半。
//
// 第 41 項接了同化,但那個檔的檔頭自己寫著:
//
//	未同化人口目前沒有負面效果 … 叛亂系統根本不存在。所以現在同化只是一個會走完的計時器
//	——**機制在、後果還沒接**。
//
// 這一檔就是那個後果。機率規則(每單位 1%、難度修正、異族管理中心減半、滅絕加倍)
// 全在 `gamedata/rebellion.go`,是從 `Check_Rebellion_` @ 0xED260 逐指令讀出來的;
// 這裡負責每回合檢定、打那一場地面戰、以及**輸了要把殖民地還給舊主**。
//
// ============ 叛軍是第四種地面部隊 ============
//
// `gamedata.GroundTypeFourth` 一直掛著「⚠ 未定名(−20,且基礎值取自另一方),
// 而殖民地防守方根本不填它」。這一輪知道它是什麼了:
//
// `Get_Rebellion_Info_` @ 0xEC65A 把守方的三種部隊填進 `[+0x0A]`(裝甲)、
// `[+0x0C]`(陸戰隊)、`[+0x0E]`(民兵),而**叛軍的數量填進 `[+0x10]`**
// ——正好是同一個陣列的第四格。所以類型 3 = **叛軍**。
//
// 這也解釋了原本那句「殖民地防守方根本不填它」:叛軍永遠是攻方,守方那一格當然是空的。
// 常數已改名為 `GroundTypeRebels`。
//
// ============ 叛軍人數不是全部未同化人口 ============
//
// 原版在決定叛亂之後又擲了一次骰:
//
//	mov  eax, ecx        ; 未同化人口單位數
//	call sub_1247A0      ; rand(1..n)
//	...
//	movsx edx, word ptr [var_8]
//	call sub_EC65A       ; → [叛軍結構+0x10] = 這個數
//
// 所以起事的是 **rand(1..未同化人口數)**。手冊完全沒提這件事。
//
// ============ 誠實留白 ============
//
//   - **只檢定玩家的殖民地。** AI 打下玩家殖民地那條路徑本來就沒有同化模型
//     (見 assimilation.go 檔頭),沒有未同化人口可以叛亂。
//   - **沒有滅絕政策。** remake 沒有這個選項,所以「×2」那一路目前不會發生
//     (`gamedata.RebellionChancePermille` 仍收這個參數,等 UI 有那個選項時直接接上)。
//   - **守方的裝甲那一格用 `PlayerColonyTanks`**,與 `InvadeColony` 守方留 0 的處理不同
//     ——那邊留 0 是因為 AI 沒有建築追蹤,這邊是玩家自己的殖民地,資料是有的。

// RebellionResult 是一次叛亂檢定的結果(供回合摘要顯示)。
type RebellionResult struct {
	ColonyName     string
	StarIndex      int
	ChancePermille int  // 這一回合的叛亂機率(千分之一)
	Rebels         int  // 起事的叛軍單位數
	Triggered      bool // 是否真的爆發
	DefenderWon    bool
	ColonyLost     bool // 守方輸 → 殖民地還給舊主
	RevertedToAI   int  // ColonyLost 時才有意義
	Message        string
}

// rebelGroundForce 回傳叛軍那一側的攻擊力基礎。
//
// 叛軍屬於**舊主那個帝國**的人口,所以用該 AI 的地面戰基礎 + 難度加成
// (`GroundAIEmpire`,原版只給 AI,人類拿 0)。舊主索引無效時退回玩家自己的基礎,
// 不讓資料缺口變成 0 攻擊力的假勝利。
func (s *GameSession) rebelGroundForce(aiIdx int) int {
	if aiIdx < 0 || aiIdx >= len(s.AIPlayers) {
		return s.playerMarineForce()
	}
	return aiMarineForce(s.AIPlayers[aiIdx]) +
		gamedata.GroundDifficultyBonus(s.Difficulty, gamedata.GroundAIEmpire)
}

// resolveRebellionCombat 打叛軍 vs 殖民地守軍那一場,回傳守方是否守住。
//
// 攻方只有一種部隊(類型 3 叛軍);守方是陸戰隊 + 裝甲 + 民兵,與 InvadeColony 的守方
// 同一套建構,差別在這裡的殖民地是玩家自己的,所以資料查得到。
func (s *GameSession) resolveRebellionCombat(colonyIdx, rebels int, rebelForce int) bool {
	colony := s.PlayerColonies[colonyIdx]

	var atkStrength, atkCounts, atkHits [gamedata.GroundUnitTypes]int
	atkStrength[gamedata.GroundTypeRebels] = rebelForce +
		gamedata.GroundTypeStrengthDelta(gamedata.GroundTypeRebels)
	atkCounts[gamedata.GroundTypeRebels] = rebels
	atkHits[gamedata.GroundTypeRebels] = gamedata.GroundMarineHitsToKill(false, false)

	// 守方:玩家自己的殖民地,難度加成 0(原版只給 AI)。
	// 用**守方版**——地底種族(薩克拉)只有在守自家殖民地時才拿那 +10,
	// 而這裡正是手冊與反組譯所說的那個情境(見 gamedata.GroundSubterraneanBonus)。
	defForce := s.playerDefendingMarineForce() +
		gamedata.GroundCommandoDefenderForceBonus(commandoLeaderTier(s.Leaders), s.RuleProfile.DefenderCommandoBonus)
	defHits := gamedata.GroundMarineHitsToKill(s.raceHasTrait(gamedata.TRAIT_HIGH_G), s.hasPoweredArmor())

	var defStrength, defCounts, defHitsArr [gamedata.GroundUnitTypes]int
	defStrength[groundTypeMarines] = defForce + gamedata.GroundTypeStrengthDelta(groundTypeMarines)
	defCounts[groundTypeMarines] = s.colonyMarinesAt(colonyIdx)
	defHitsArr[groundTypeMarines] = defHits
	defStrength[groundTypeTanks] = defForce + gamedata.GroundTypeStrengthDelta(groundTypeTanks)
	defCounts[groundTypeTanks] = s.colonyTanksAt(colonyIdx)
	defHitsArr[groundTypeTanks] = tankHitsToKillFor(s.Player, s.raceHasTrait(gamedata.TRAIT_HIGH_G))
	defStrength[gamedata.GroundTypeMilitia] = defForce +
		gamedata.GroundTypeStrengthDelta(gamedata.GroundTypeMilitia)
	defCounts[gamedata.GroundTypeMilitia] = gamedata.ColonyMilitiaUnits(colony.Population)
	defHitsArr[gamedata.GroundTypeMilitia] = defHits

	// 確定性亂數:同一回合、同一個殖民地永遠打出同一場(存檔/探針可重現),
	// 種子式樣沿用 InvadeColony 的既有寫法,只換常數避免與它撞流。
	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + int64(colonyIdx)*97 + 8081))
	atkSide := gamedata.NewGroundSide(atkStrength, atkCounts, atkHits)
	defSide := gamedata.NewGroundSide(defStrength, defCounts, defHitsArr)
	res := gamedata.ResolveGroundCombatOrig(atkSide, defSide, rng.Intn, 0)
	return !res.AttackerWon
}

// colonyMarinesAt / colonyTanksAt 讀平行陣列,越界回 0(平行陣列是延遲配置的,
// 見 GameSession 欄位註解)。
func (s *GameSession) colonyMarinesAt(i int) int {
	if i < 0 || i >= len(s.PlayerColonyMarines) {
		return 0
	}
	return s.PlayerColonyMarines[i]
}

func (s *GameSession) colonyTanksAt(i int) int {
	if i < 0 || i >= len(s.PlayerColonyTanks) {
		return 0
	}
	return s.PlayerColonyTanks[i]
}

// advanceRebellions 每回合對所有玩家殖民地做一次叛亂檢定。
//
// **由後往前掃**:叛亂成功會移除殖民地,由前往後掃會讓後面的索引全部位移。
func (s *GameSession) advanceRebellions() []RebellionResult {
	var out []RebellionResult
	for i := len(s.PlayerColonies) - 1; i >= 0; i-- {
		if r, ok := s.checkRebellionAt(i); ok {
			out = append(out, r)
		}
	}
	// 由後往前掃出來的順序是反的,翻回殖民地順序再回報。
	for l, r := 0, len(out)-1; l < r; l, r = l+1, r-1 {
		out[l], out[r] = out[r], out[l]
	}
	return out
}

// checkRebellionAt 對單一殖民地做檢定。第二個回傳值是「有沒有事情發生」
// ——沒有未同化人口、或擲骰沒中,都回 false(不要在回合摘要洗版)。
func (s *GameSession) checkRebellionAt(i int) (RebellionResult, bool) {
	c := s.PlayerColonies[i]
	if c.UnassimilatedPop <= 0 {
		return RebellionResult{}, false
	}
	// 原版還有一個「這個殖民地被征服過嗎」的獨立閘門(`[colony+0x12F] == 4` 直接返回)。
	// remake 不需要另立旗標:自己拓殖的殖民地 UnassimilatedPop 恆為 0,
	// 上面那一行就等價於同一個閘門。
	chance := gamedata.RebellionChancePermille(
		c.UnassimilatedPop, s.Difficulty,
		true, // 殖民地主人是人類玩家
		true, // 叛軍屬於 AI 帝國(舊主)
		s.buildingsFor(i)[alienManagementCenterName],
		false, // remake 沒有滅絕政策,見檔頭
	)
	roll := s.eventRoll(gamedata.RebellionRollMax)
	res := RebellionResult{
		ColonyName:     s.starName(s.PlayerColonyStarIndex(i)),
		StarIndex:      s.PlayerColonyStarIndex(i),
		ChancePermille: chance,
	}
	if !gamedata.RebellionTriggers(chance, roll) {
		return res, false
	}
	res.Triggered = true
	res.Rebels = gamedata.RebellionRebelUnits(c.UnassimilatedPop, s.eventRoll)

	oldRuler := -1
	if c.ConqueredFromKnown {
		oldRuler = c.ConqueredFrom
	}
	res.DefenderWon = s.resolveRebellionCombat(i, res.Rebels, s.rebelGroundForce(oldRuler))

	if res.DefenderWon {
		// 鎮壓成功:起事的那些人口單位沒了(它們就是叛軍)。
		s.suppressRebellion(i, res.Rebels)
		res.Message = fmt.Sprintf("%s:%d 單位未同化人口起事,已遭鎮壓",
			res.ColonyName, res.Rebels)
		return res, true
	}
	res.ColonyLost = true
	res.RevertedToAI = oldRuler
	if s.revertColonyToOldRuler(i, oldRuler) {
		res.Message = fmt.Sprintf("%s:叛亂成功,殖民地已回到原主人手中", res.ColonyName)
	} else {
		// 舊主已經不存在(被滅了/舊存檔沒記)——殖民地脫離帝國但沒有人接手。
		res.RevertedToAI = -1
		res.Message = fmt.Sprintf("%s:叛亂成功,殖民地已脫離帝國", res.ColonyName)
	}
	return res, true
}

// suppressRebellion 鎮壓成功後的後果:起事的人口單位被消滅。
//
// 手冊沒有逐字寫「鎮壓後死多少人」,原版那段在 `Resolve_Rebellion_Troops_` 裡
// (沒有逐指令抄)。這裡採「起事者全滅」——**這是 remake 的建模選擇,不是手冊或
// 反組譯的逐字依據**,寫明白。人口下限 1,免得殖民地在鎮壓中自我消滅。
func (s *GameSession) suppressRebellion(i, rebels int) {
	c := &s.PlayerColonies[i]
	lost := rebels
	if lost > c.UnassimilatedPop {
		lost = c.UnassimilatedPop
	}
	if c.Population-lost < 1 {
		lost = c.Population - 1
	}
	if lost < 0 {
		lost = 0
	}
	c.Population -= lost
	c.UnassimilatedPop -= lost
	if c.UnassimilatedPop < 0 {
		c.UnassimilatedPop = 0
	}
	s.recalcColonyMorale(i)
}

// revertColonyToOldRuler 把殖民地還給舊主(手冊 p.165「the colony reverts back」)。
//
// 回傳 false = 舊主不存在,呼叫端改走「脫離帝國」那條。
//
// 移除玩家這一側用既有的 `removePlayerColony`(它負責所有平行陣列的長度不變量);
// 這裡只負責把資料補到 AI 那一側,並在該星再也沒有玩家殖民地時把星的歸屬翻回去。
func (s *GameSession) revertColonyToOldRuler(i, aiIdx int) bool {
	if aiIdx < 0 || aiIdx >= len(s.AIPlayers) {
		s.removePlayerColony(i)
		s.refreshStarOwnerAfterLoss(i, -1)
		return false
	}
	colony := s.PlayerColonies[i]
	starIdx := s.PlayerColonyStarIndex(i)
	planetIdx := -1
	if i < len(s.PlayerColonyPlanets) {
		planetIdx = s.PlayerColonyPlanets[i]
	}
	buildings := s.buildingsFor(i)

	// 還回去的是「已經同化完的那部分人口」——起事的那些人成了新主人的自己人。
	colony.UnassimilatedPop = 0
	colony.AssimilationProgress = 0
	colony.ConqueredFrom, colony.ConqueredFromKnown = 0, false

	ai := &s.AIPlayers[aiIdx]
	ai.Colonies = append(ai.Colonies, colony)
	ai.ColonyStars = append(ai.ColonyStars, starIdx)
	ai.ColonyPlanets = append(ai.ColonyPlanets, planetIdx)
	ai.ColonyBuildings = append(ai.ColonyBuildings, buildings)

	s.removePlayerColony(i)
	s.refreshStarOwnerAfterLoss(i, starIdx)
	return true
}

// refreshStarOwnerAfterLoss 在玩家失去一個殖民地之後更新那顆星的歸屬。
//
// **只在這顆星上再也沒有玩家殖民地時才翻面**——同星系多殖民地打開之後
// (第 25/25 項),一個星系可能還有玩家的另一個殖民地。與 InvadeColony 的同款判斷一致。
//
// 呼叫時 removePlayerColony 已經跑過,所以掃的是移除後的陣列。
func (s *GameSession) refreshStarOwnerAfterLoss(_ int, starIdx int) {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return
	}
	for j := range s.PlayerColonies {
		if s.PlayerColonyStarIndex(j) == starIdx {
			return // 玩家在這顆星還有別的殖民地
		}
	}
	// 這顆星還有沒有 AI 的殖民地?有就判給敵方,沒有就變無主。
	for a := range s.AIPlayers {
		for _, st := range s.AIPlayers[a].ColonyStars {
			if st == starIdx {
				s.Stars[starIdx].Owner = 2
				return
			}
		}
	}
	s.Stars[starIdx].Owner = 0
}
