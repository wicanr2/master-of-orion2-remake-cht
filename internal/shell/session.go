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

// Star 是星系圖上的一顆星(供星圖渲染;正規化座標 0..1)。
type Star struct {
	X, Y     float64 // 0..1 正規化位置
	Spectral int     // 0=藍 1=白 2=黃 3=橙 4=紅 5=棕 6=黑洞
	Size     int     // 0=大 .. 3=小
	Name     string
	Owner    int // 0=無主 1=玩家 2=AI
}

// demoStars 是最小示範星系(固定佈局,供星圖視窗渲染;待接真星系生成 + STARNAME.LBX 真星名)。
func demoStars() []Star {
	return []Star{
		{0.12, 0.18, 2, 0, "獵戶", 1}, {0.30, 0.10, 1, 1, "天狼", 0},
		{0.48, 0.22, 3, 2, "南門", 0}, {0.68, 0.14, 0, 1, "參宿", 0},
		{0.86, 0.24, 4, 2, "畢宿", 2}, {0.18, 0.42, 4, 3, "織女", 0},
		{0.40, 0.48, 2, 1, "河鼓", 1}, {0.60, 0.40, 1, 0, "角宿", 0},
		{0.80, 0.50, 3, 2, "心宿", 0}, {0.10, 0.68, 0, 1, "北落", 0},
		{0.34, 0.72, 2, 2, "五車", 0}, {0.54, 0.66, 5, 3, "軒轅", 0},
		{0.72, 0.74, 4, 1, "太微", 2}, {0.90, 0.66, 1, 0, "天津", 0},
		{0.24, 0.88, 3, 2, "婁宿", 0}, {0.62, 0.86, 2, 1, "氐宿", 0},
	}
}

// GameSession 是一局進行中的遊戲狀態。玩家操作改變狀態,EndTurn 推進一回合(結算玩家 + 各 AI)。
type GameSession struct {
	Turn             int
	Player           engine.PlayerState
	PlayerColonies   []engine.ColonyState
	AIPlayers        []AIOpponent
	LastPlayerOutput engine.EmpireOutput // 上一回合玩家結算(供畫面顯示)
	Stars            []Star              // 星系圖
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
		Stars: demoStars(),
	}
}
