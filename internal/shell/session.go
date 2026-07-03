// Package shell 是「可玩遊戲殼」的純邏輯核心:活的對局狀態、輸入命中判定。
// 不 import ebiten(維持可純測);ebiten 的繪製與輸入輪詢在 cmd/moo2。
package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

// AIOpponent 是一個由 AI 操控的對手帝國。
type AIOpponent struct {
	Name     string
	Player   engine.PlayerState
	Colonies []engine.ColonyState
	Decider  ai.Decider
}

// GameSession 是一局進行中的遊戲狀態。玩家操作改變狀態,EndTurn 推進一回合(結算玩家 + 各 AI)。
type GameSession struct {
	Turn             int
	Player           engine.PlayerState
	PlayerColonies   []engine.ColonyState
	AIPlayers        []AIOpponent
	LastPlayerOutput engine.EmpireOutput // 上一回合玩家結算(供畫面顯示)
}

// EndTurn 推進一回合:先結算玩家帝國,再讓各 AI 對手自行決策並結算,回合數 +1。
func (s *GameSession) EndTurn() {
	s.LastPlayerOutput = engine.RunEmpireTurn(s.Player, s.PlayerColonies)
	s.Player = s.LastPlayerOutput.Player
	for i := range s.AIPlayers {
		out := engine.RunAIEmpireTurn(s.AIPlayers[i].Player, s.AIPlayers[i].Colonies, s.AIPlayers[i].Decider)
		s.AIPlayers[i].Player = out.Player
	}
	s.Turn++
}

// NewDemoSession 建一個最小可玩對局:玩家 2 殖民地 + 1 個科學傾向 AI 對手。
// 供「最小可玩迴圈」骨架用;正式新遊戲流程(選種族/星系生成)為後續工作。
func NewDemoSession() *GameSession {
	mkColonies := func() []engine.ColonyState {
		return []engine.ColonyState{
			{Population: 8, PopMax: 20, Farmers: 3, Workers: 4, Scientists: 1,
				FoodPerFarmer: 4, IndustryPerWorker: 6, ResearchPerScientist: 30,
				PlanetSize: 3 /*LARGE*/, MoralePercent: 10},
			{Population: 4, PopMax: 12, Farmers: 2, Workers: 1, Scientists: 1,
				FoodPerFarmer: 4, IndustryPerWorker: 5, ResearchPerScientist: 20,
				PlanetSize: 1 /*SMALL*/},
		}
	}
	return &GameSession{
		Turn:           1,
		Player:         engine.PlayerState{BC: 100, TaxRate: 40, Maintenance: 5, ResearchTopic: 1},
		PlayerColonies: mkColonies(),
		AIPlayers: []AIOpponent{{
			Name:     "AI (賽隆人)",
			Player:   engine.PlayerState{BC: 100, TaxRate: 40, Maintenance: 5, ResearchTopic: 1},
			Colonies: mkColonies(),
			Decider:  ai.NewRemakeDecider(ai.ProfileScientific),
		}},
	}
}
