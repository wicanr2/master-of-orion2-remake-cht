package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

// originalAITurnData 是 Compute_AI_Data_ 玩家可表示切片的單回合 typed cache。
//
// 它刻意不是 GameSession／AIOpponent 欄位，也不進 sessionSnapshot：原版在同一世界回合
// 建立並釋放 cache。PersistentColonies 是存檔權威狀態；TurnColonies 則包含難度與事件的
// 暫態投影，職務排序與最終帝國結算必須消費同一份投影，避免同回合重建時漂移。
type originalAITurnData struct {
	AIIndex            int
	Player             engine.PlayerState
	PersistentColonies []engine.ColonyState
	TurnColonies       []engine.ColonyState
	Jobs               engine.OriginalAIJobContext
}

// buildOriginalAITurnData 在 AI 外交成長完成、殖民地職務與帝國結算之前建立本回合 cache。
// 呼叫端須先同步科技 grant、種族、政府、指揮點與協議收入；本函式只快照 consumer 輸入，
// 不偷偷修補未知欄位。無效 AI index 失敗即關閉。
func (s *GameSession) buildOriginalAITurnData(aiIndex int) (originalAITurnData, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return originalAITurnData{}, false
	}
	a := &s.AIPlayers[aiIndex]
	persistent := append([]engine.ColonyState(nil), a.Colonies...)
	data := originalAITurnData{
		AIIndex:            aiIndex,
		Player:             a.Player,
		PersistentColonies: persistent,
		TurnColonies:       s.aiColoniesForTurn(aiIndex, persistent),
		Jobs: engine.OriginalAIJobContext{
			Personality:         a.Personality,
			LateTech:            aiOriginalLateTechReached(a.Player),
			ColonyFoodHalf:      make([]int, len(a.Colonies)),
			ColonyFoodHalfKnown: make([]bool, len(a.Colonies)),
			ColonyBlockaded:     make([]bool, len(a.Colonies)),
		},
	}
	knownTech := knownTechnologyApplications(a.Player)
	for colony := range a.Colonies {
		var built map[string]bool
		if colony < len(a.ColonyBuildings) {
			built = a.ColonyBuildings[colony]
		}
		data.Jobs.ColonyFoodHalf[colony], data.Jobs.ColonyFoodHalfKnown[colony] =
			originalAIColonyFoodHalf(a.Colonies[colony], built, knownTech)
		if colony >= len(a.ColonyStars) {
			continue
		}
		star, slot := a.ColonyStars[colony], a.PopulationRaceSlot
		if star >= 0 && star < len(s.Stars) && a.PopulationRaceSlotKnown && slot >= 0 && slot < 8 {
			data.Jobs.ColonyBlockaded[colony] = s.Stars[star].BlockadedMask&(1<<slot) != 0
		}
	}
	return data, true
}

// coloniesAfterJobs 把職務結果合回權威殖民地，但不保存 TurnColonies 上的事件／難度暫態值。
// 若原版 typed 輸入不完整，保留既有可玩 fallback，且不讓 fallback 偷改原版沒有 writer 的稅率。
func (d *originalAITurnData) applyJobs(decider ai.Decider) (engine.PlayerState, []engine.ColonyState, bool, bool) {
	assigned, freighterPressure, exact := engine.ApplyOriginalAIJobsWithTransport(d.Player, d.TurnColonies, d.Jobs)
	colonies := mergeAIJobAssignments(d.PersistentColonies, assigned)
	if exact {
		return d.Player, colonies, freighterPressure, true
	}
	originalTaxRate := d.Player.TaxRate
	player, fallback := engine.ApplyAIEconomy(d.Player, d.PersistentColonies, decider)
	player.TaxRate = originalTaxRate
	return player, fallback, false, false
}

// economyColonies 以職務合併後的權威狀態重建一次暫態投影。這不是第二份獨立 cache：
// 它沿用同一筆 data 的生命週期，且只反映職務變更後必須重算的 derived colony output。
func (d *originalAITurnData) economyColonies(s *GameSession, colonies []engine.ColonyState) []engine.ColonyState {
	d.PersistentColonies = colonies
	d.TurnColonies = s.aiColoniesForTurn(d.AIIndex, colonies)
	return d.TurnColonies
}
