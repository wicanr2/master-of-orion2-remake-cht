package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 帝國歷史記錄對應 Record_History_ @ 0x10208A；證據見 docs/spec/history-ring.md。
type HistorySample struct {
	Fleet      int `json:"fleet"`
	Technology int `json:"technology"`
	Population int `json:"pop"`
	Buildings  int `json:"buildings"`
	// 舊欄位只用來辨識不相容存檔；新資料不再寫入。
	BC       int `json:"bc,omitempty"`
	Research int `json:"research,omitempty"`
}
type HistoryTurn struct {
	Turn    int             `json:"turn"`
	Empires []HistorySample `json:"empires"`
}

const historyMaxTurns = 350
const historyMaxValue = 250

func (s *GameSession) recordHistory() {
	raw := []HistorySample{historyRawSample(s.AllShips(), s.Player.CompletedTopics, s.Player.HyperAdvancedLevels, s.PlayerColonies, s.ColonyBuildings)}
	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		raw = append(raw, historyRawSample(a.Ships, a.Player.CompletedTopics, a.Player.HyperAdvancedLevels, a.Colonies, a.ColonyBuildings))
	}
	s.ensureHistoryScales()
	for metric := HistoryFleet; metric < historyMetricCount; metric++ {
		old := s.HistoryScales[metric]
		for historyRawMax(raw, metric)/s.HistoryScales[metric] > historyMaxValue {
			s.HistoryScales[metric]++
		}
		if old != s.HistoryScales[metric] {
			for i := range s.History {
				for j := range s.History[i].Empires {
					v := historyValue(&s.History[i].Empires[j], metric)
					setHistoryValue(&s.History[i].Empires[j], metric, v*old/s.HistoryScales[metric])
				}
			}
		}
	}
	emp := make([]HistorySample, len(raw))
	for i := range raw {
		for metric := HistoryFleet; metric < historyMetricCount; metric++ {
			setHistoryValue(&emp[i], metric, historyValue(&raw[i], metric)/s.HistoryScales[metric])
		}
	}
	s.History = append(s.History, HistoryTurn{Turn: s.Turn, Empires: emp})
	if len(s.History) > historyMaxTurns {
		s.History = s.History[len(s.History)-historyMaxTurns:]
	}
}
func (s *GameSession) ensureHistoryScales() {
	for i := range s.HistoryScales {
		if s.HistoryScales[i] < 1 {
			s.HistoryScales[i] = 1
		}
	}
}
func historyRawSample(ships []Ship, completed map[gamedata.ResearchTopic]bool, hyper map[gamedata.ResearchTopic]int, colonies []engine.ColonyState, buildings []map[string]bool) HistorySample {
	v := HistorySample{}
	for _, sh := range ships {
		class, _ := shipClassFromName(sh.Class)
		v.Fleet += 10 << (int(class) + 1)
	}
	for topic, done := range completed {
		if done && int(topic) >= 0 && int(topic) < len(gamedata.OrigTopicCost) && !gamedata.IsHyperAdvancedTopic(topic) {
			v.Technology += gamedata.OrigTopicCost[int(topic)]
		}
	}
	for _, levels := range hyper {
		for level := 0; level < levels; level++ {
			v.Technology += gamedata.HyperAdvancedCost(gamedata.Profile13()) + level*10000
		}
	}
	for _, c := range colonies {
		v.Population += c.Population
	}
	for _, colony := range buildings {
		for name, built := range colony {
			if built {
				if id, ok := gamedata.OriginalBuildingIDForName(name); ok {
					if cost, ok := gamedata.OriginalBuildingProductionCost(id); ok {
						v.Buildings += cost
					}
				}
			}
		}
	}
	return v
}
func historyRawMax(samples []HistorySample, m HistoryMetric) int {
	max := 0
	for i := range samples {
		if v := historyValue(&samples[i], m); v > max {
			max = v
		}
	}
	return max
}

func (s *GameSession) playerFleetStrength() int {
	n := 0
	for _, sh := range s.AllShips() {
		n += sh.WeaponAttack + sh.BonusHP/2
	}
	return n
}
func (s *GameSession) HistoryEmpireNames() []string {
	names := []string{"你"}
	for i := range s.AIPlayers {
		names = append(names, s.AIPlayers[i].Name)
	}
	return names
}

type HistoryMetric int

const (
	HistoryFleet HistoryMetric = iota
	HistoryTechnology
	HistoryPopulation
	HistoryBuildings
	historyMetricCount
)

func HistoryMetricName(m HistoryMetric) string {
	switch m {
	case HistoryFleet:
		return "艦隊"
	case HistoryTechnology:
		return "科技"
	case HistoryBuildings:
		return "建築"
	default:
		return "人口"
	}
}
func (s *GameSession) HistorySeries(m HistoryMetric) (series [][]int, turns []int) {
	if len(s.History) == 0 {
		return nil, nil
	}
	n := 0
	for _, h := range s.History {
		if len(h.Empires) > n {
			n = len(h.Empires)
		}
	}
	series = make([][]int, n)
	for i := range series {
		series[i] = make([]int, 0, len(s.History))
	}
	for _, h := range s.History {
		turns = append(turns, h.Turn)
		for i := 0; i < n; i++ {
			v := 0
			if i < len(h.Empires) {
				v = historyValue(&h.Empires[i], m)
			}
			series[i] = append(series[i], v)
		}
	}
	return series, turns
}
func historyValue(s *HistorySample, m HistoryMetric) int {
	switch m {
	case HistoryFleet:
		return s.Fleet
	case HistoryTechnology:
		return s.Technology
	case HistoryBuildings:
		return s.Buildings
	default:
		return s.Population
	}
}
func setHistoryValue(s *HistorySample, m HistoryMetric, v int) {
	switch m {
	case HistoryFleet:
		s.Fleet = v
	case HistoryTechnology:
		s.Technology = v
	case HistoryBuildings:
		s.Buildings = v
	default:
		s.Population = v
	}
}
func (s *GameSession) PlayerFleetStrengthForUI() int { return s.playerFleetStrength() }
