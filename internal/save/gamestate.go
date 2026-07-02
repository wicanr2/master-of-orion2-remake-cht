package save

import "fmt"

// 存檔佈局常數,對照 openorion2 gamestate.cpp / gamestate.h。
const (
	saveVersion       = 0xe0   // GameConfig.version 必為此值
	saveGameNameSize  = 37     // SAVE_GAME_NAME_SIZE
	colonyCountOffset = 0x25b  // COLONY_COUNT_OFFSET
	galaxyOffset      = 0x31be4 // GameState::load 對 galaxy 的 seek

	maxNebulas  = 4
	maxColonies = 250
	maxPlanets  = 72 * 5 // MAX_STARS * MAX_ORBITS
	maxStars    = 72
	leaderCount = 67
	maxPlayers  = 8
	maxShips    = 500
)

// GameConfig 是存檔開頭的設定區(59 bytes),對照 GameConfig::load。
type GameConfig struct {
	Version      uint32
	SaveGameName string
	Stardate     uint32

	Multiplayer                bool
	EndOfTurnSummary           bool
	EndOfTurnWait              bool
	RandomEvents               bool
	EnemyMoves                 bool
	ExpandingHelp              bool
	AutoSelectShips            bool
	Animations                 bool
	AutoSelectColony           bool
	ShowRelocationLines        bool
	ShowGNNReport              bool
	AutoDeleteTradeGoodHousing bool
	ShowOnlySeriousTurnSummary bool
	ShipInitiative             bool
}

func (c *GameConfig) load(r *reader) error {
	if err := r.need(59); err != nil {
		return err
	}
	c.Version = r.u32()
	if c.Version != saveVersion {
		return fmt.Errorf("save: 版本非 %#x(得 %#x)", saveVersion, c.Version)
	}
	c.SaveGameName = r.cstr(saveGameNameSize)
	c.Stardate = r.u32()
	b := func() bool { return r.u8() != 0 }
	c.Multiplayer = b()
	c.EndOfTurnSummary = b()
	c.EndOfTurnWait = b()
	c.RandomEvents = b()
	c.EnemyMoves = b()
	c.ExpandingHelp = b()
	c.AutoSelectShips = b()
	c.Animations = b()
	c.AutoSelectColony = b()
	c.ShowRelocationLines = b()
	c.ShowGNNReport = b()
	c.AutoDeleteTradeGoodHousing = b()
	c.ShowOnlySeriousTurnSummary = b()
	c.ShipInitiative = b()
	return nil
}

// Nebula 是星雲(5 bytes),對照 Nebula::load。
type Nebula struct {
	X, Y uint16
	Type uint8
}

// Galaxy 是星系設定(32 bytes),對照 Galaxy::load。
type Galaxy struct {
	SizeFactor    uint8
	Width, Height uint16
	Nebulas       [maxNebulas]Nebula
	NebulaCount   uint8
}

func (g *Galaxy) load(r *reader) error {
	if err := r.need(32); err != nil {
		return err
	}
	g.SizeFactor = r.u8()
	r.skip(4) // 未知
	g.Width = r.u16()
	g.Height = r.u16()
	r.skip(2) // 未知
	for i := 0; i < maxNebulas; i++ {
		g.Nebulas[i] = Nebula{X: r.u16(), Y: r.u16(), Type: r.u8()}
	}
	g.NebulaCount = r.u8()
	return nil
}

// GameState 是解析後的存檔。目前已實作:GameConfig、Galaxy、各區段計數。
// 各實體陣列(colonies/planets/stars/leaders/players/ships)的完整欄位解析為後續工作,
// 尚未填入(見 WORKLIST Phase 1)。
type GameState struct {
	Config GameConfig
	Galaxy Galaxy

	ColonyCount int
	// 以下計數需完成對應實體結構解析後才可靠取得,目前為 -1(未解析)。
	PlanetCount int
	StarCount   int
	PlayerCount int
	ShipCount   int
}

// Load 解析一份存檔位元組。
func Load(data []byte) (*GameState, error) {
	r := newReader(data)
	gs := &GameState{PlanetCount: -1, StarCount: -1, PlayerCount: -1, ShipCount: -1}

	if err := gs.Config.load(r); err != nil {
		return nil, err
	}
	if err := r.seek(galaxyOffset); err != nil {
		return nil, fmt.Errorf("save: 定位 galaxy 失敗: %w", err)
	}
	if err := gs.Galaxy.load(r); err != nil {
		return nil, fmt.Errorf("save: 解析 galaxy 失敗: %w", err)
	}
	if err := r.seek(colonyCountOffset); err != nil {
		return nil, fmt.Errorf("save: 定位 colonyCount 失敗: %w", err)
	}
	if err := r.need(2); err != nil {
		return nil, err
	}
	gs.ColonyCount = int(r.u16())
	if gs.ColonyCount > maxColonies {
		return nil, fmt.Errorf("save: colonyCount %d 超過上限 %d", gs.ColonyCount, maxColonies)
	}
	return gs, nil
}
