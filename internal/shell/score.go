package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// score.go:把 gamedata 的原版計分公式接到這局遊戲的實際狀態上(Hi-Score / Hall of Fame)。
//
// 係數全部來自反組譯(見 gamedata/score.go 檔頭的八個 `Get_*_Score_` 位址);
// 這一層只負責「從 GameSession 撈出每個輸入欄位」。

// galaxySizeIndex 回傳目前星圖大小在 GalaxySizes 裡的索引(找不到就用最接近的)。
// 原版的時間分與俘虜人口分都用「這個索引 + 1」。
func (s *GameSession) galaxySizeIndex() int {
	best, bestDiff := 0, 1<<30
	for i, g := range GalaxySizes {
		d := g.Stars - len(s.Stars)
		if d < 0 {
			d = -d
		}
		if d < bestDiff {
			best, bestDiff = i, d
		}
	}
	return best
}

// playerPopulation 回傳玩家所有殖民地的人口總和(原版 `Get_Population_Score_` 的掃法)。
func (s *GameSession) playerPopulation() int {
	n := 0
	for _, c := range s.PlayerColonies {
		n += c.Population
	}
	return n
}

// researchTopicsKnown 回傳玩家已完成的研究主題數。
//
// 原版掃的是玩家結構 +0xC4 起、長度 0x53(83)的研究主題陣列,數狀態 == RSTATE_KNOWN 的個數;
// remake 的對應資料是 Player.CompletedTopics,語意相同。
func (s *GameSession) researchTopicsKnown() int {
	n := 0
	for _, done := range s.Player.CompletedTopics {
		if done {
			n++
		}
	}
	return n
}

// hyperAdvancedLevels 回傳玩家已達成的 Hyper-Advanced 等級數。
// 原版掃玩家結構 +0x21C 起的 8 個位元組並加總；remake 直接加總同一八 topic 的
// HyperAdvancedLevels。舊存檔缺 map 時，以 CompletedTopics 的每個 Hyper 布林補一級。
func (s *GameSession) hyperAdvancedLevels() int {
	n := 0
	if s.Player.HyperAdvancedLevels != nil {
		for topic, level := range s.Player.HyperAdvancedLevels {
			if gamedata.IsHyperAdvancedTopic(topic) && level > 0 {
				n += level
			}
		}
		return n
	}
	for topic, done := range s.Player.CompletedTopics {
		if done && gamedata.IsHyperAdvancedTopic(topic) {
			n++
		}
	}
	return n
}

// racesEliminated 回傳玩家殲滅掉的種族數(以「原本有幾個對手、現在還剩幾個有殖民地」推得)。
//
// ⚠ 這是 remake 的推法:原版有一個「這個玩家滅了誰」的逐玩家陣列(player+0x1F2),
// remake 沒有追蹤「是誰滅的」——AI 也可能互相殲滅。目前把所有已滅亡的對手都算給玩家,
// 這在單人對局裡多數情況成立,但 AI 互滅時會高估。標明,不假裝精確。
func (s *GameSession) racesEliminated() int {
	n := 0
	for _, a := range s.AIPlayers {
		if len(a.Colonies) == 0 {
			n++
		}
	}
	return n
}

// FinalScore 算出玩家目前的最終得分(逐項 + 總分)。
// 對局尚未結束時一樣可以呼叫——原版的 Hi-Score 畫面也是隨時算得出來的。
func (s *GameSession) FinalScore() gamedata.ScoreBreakdown {
	multiplier := s.ScoreBaseMultiplierPercent
	if multiplier <= 0 {
		multiplier = 100
	}
	// sub_58F4A 對 Evolutionary Mutation 的 known-state 額外加入 4 Picks；remake 尚無
	// mutation 再選種族能力畫面，因此這四點目前全數保留。
	if playerStateKnowsTech(s.Player, gamedata.TOPIC_TRANS_GENETICS, gamedata.TECH_EVOLUTIONARY_MUTATION) {
		multiplier += 40
	}
	return gamedata.ComputeScore(gamedata.ScoreInput{
		GalaxySizeIndex: s.galaxySizeIndex(),
		// 種族數 = 玩家 + AI 對手(手冊:「the number of races involved in the struggle for
		// galactic domination (not including the Antarans)」)。
		RaceCount: 1 + len(s.AIPlayers),
		// 已過回合數:原版用「目前星曆 − 起始星曆 3500」算,remake 的 Turn 從 1 起算,
		// 兩者都是「玩了幾回合」,語意相同。
		TurnsElapsed:        s.Turn - 1,
		Population:          s.playerPopulation(),
		CapturedPopulation:  s.CapturedPop,
		ResearchTopicsKnown: s.researchTopicsKnown(),
		HyperAdvancedLevels: s.hyperAdvancedLevels(),
		RacesEliminated:     s.racesEliminated(),
		// ⚠ remake 目前沒有獵戶座星系(第三梯項目),此項恆為 0——不是漏接,是那個系統還沒做。
		OrionCaptured:     false,
		CouncilVictory:    s.Victory.Over && s.Victory.Reason == engine.VictoryHighCouncil && s.Victory.Winner == "player",
		AntaranVictory:    s.AntaranHomeworldConquered,
		MultiplierPercent: multiplier,
	})
}

// SetCustomRaceUnusedPicks 保存自訂種族確認時尚未使用的 Picks，供原版最終分數倍率使用。
func (s *GameSession) SetCustomRaceUnusedPicks(picks int) {
	if picks < 0 {
		picks = 0
	}
	s.ScoreBaseMultiplierPercent = 100 + picks*10
}

// ScoreLine 是 Hi-Score 畫面的一列(項目名 + 分數)。
type ScoreLine struct {
	Label string
	Value int
}

// ScoreLines 把逐項得分整理成顯示用的列表(順序比照原版 `Draw_*_Score_` 的分項)。
// 分數為 0 的項目照樣列出——原版畫面也是逐項固定顯示,不是有分才畫。
func (s *GameSession) ScoreLines() []ScoreLine {
	b := s.FinalScore()
	return []ScoreLine{
		{"時間 / 星圖 / 種族數", b.Time},
		{"人口", b.Population},
		{"俘虜人口", b.Captured},
		{"科技", b.Technology},
		{"殲滅種族", b.Elimination},
		{"攻下獵戶座", b.Orion},
		{"議會勝利", b.Council},
		{"擊敗安塔蘭", b.Antares},
		{"總分", b.Total},
	}
}
