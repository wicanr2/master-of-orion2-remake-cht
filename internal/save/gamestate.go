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

// GameState 是完整解析後的存檔。對照 openorion2 GameState::load(gamestate.cpp:1844)。
// 各實體陣列固定為 MAX_* 長度(存檔一律含滿),Count 欄為存檔標示的有效數。
type GameState struct {
	Config GameConfig
	Galaxy Galaxy

	ColonyCount int
	Colonies    []Colony // len maxColonies
	PlanetCount int
	Planets     []Planet // len maxPlanets
	StarCount   int
	Stars       []Star // len maxStars
	Leaders     []Leader // len leaderCount(無獨立計數)
	PlayerCount int
	Players     []Player // len maxPlayers
	ShipCount   int
	Ships       []Ship // len maxShips

	// SeqEnd 是順序資料區(colonies…ships)解析後的結尾 offset,供驗證用。
	SeqEnd int
}

// Load 解析一份存檔位元組。讀取序列對照 GameState::load:config → galaxy(檔尾)→
// colonyCount → colonies → planetCount → planets → starCount → stars → leaders →
// playerCount → players → shipCount → ships。
func Load(data []byte) (*GameState, error) {
	r := newReader(data)
	gs := &GameState{}

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

	// 順序區從 colonyCountOffset 起。每讀一個計數前先確保還有位元組。
	readCount := func(name string, max int) (int, error) {
		if err := r.need(2); err != nil {
			return 0, fmt.Errorf("save: 讀取 %s 失敗: %w", name, err)
		}
		n := int(r.u16())
		if n > max {
			return 0, fmt.Errorf("save: %s=%d 超過上限 %d(schema 可能對不齊)", name, n, max)
		}
		return n, nil
	}

	var err error
	if gs.ColonyCount, err = readCount("colonyCount", maxColonies); err != nil {
		return nil, err
	}
	gs.Colonies = make([]Colony, maxColonies)
	for i := range gs.Colonies {
		if err := r.need(1); err != nil {
			return nil, fmt.Errorf("save: colony %d: %w", i, err)
		}
		gs.Colonies[i].load(r)
	}

	if gs.PlanetCount, err = readCount("planetCount", maxPlanets); err != nil {
		return nil, err
	}
	gs.Planets = make([]Planet, maxPlanets)
	for i := range gs.Planets {
		gs.Planets[i].load(r)
	}

	if gs.StarCount, err = readCount("starCount", maxStars); err != nil {
		return nil, err
	}
	gs.Stars = make([]Star, maxStars)
	for i := range gs.Stars {
		gs.Stars[i].load(r)
	}

	// leaders 無獨立計數,固定 LEADER_COUNT 個。
	gs.Leaders = make([]Leader, leaderCount)
	for i := range gs.Leaders {
		gs.Leaders[i].load(r)
	}

	if gs.PlayerCount, err = readCount("playerCount", maxPlayers); err != nil {
		return nil, err
	}
	gs.Players = make([]Player, maxPlayers)
	for i := range gs.Players {
		gs.Players[i].load(r)
	}

	if gs.ShipCount, err = readCount("shipCount", maxShips); err != nil {
		return nil, err
	}
	gs.Ships = make([]Ship, maxShips)
	for i := range gs.Ships {
		gs.Ships[i].load(r)
	}

	if r.at() > len(data) {
		return nil, fmt.Errorf("save: 順序區讀取越界(pos=%d len=%d)", r.at(), len(data))
	}
	gs.SeqEnd = r.at()
	return gs, nil
}
