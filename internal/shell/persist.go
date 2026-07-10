package shell

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// saveFormatVersion 標記存檔格式版本(未來欄位變動時據以相容/拒絕)。
const saveFormatVersion = 1

// aiSnapshot 是一個 AI 對手的可序列化快照。Decider 為介面不能直接序列化,故只存其性格
// (ai.Profile,純 struct),讀檔時以 ai.NewRemakeDecider 重建。
type aiSnapshot struct {
	Name            string               `json:"name"`
	Player          engine.PlayerState   `json:"player"`
	Colonies        []engine.ColonyState `json:"colonies"`
	Profile         ai.Profile           `json:"profile"`
	FleetStrength   int                  `json:"fleetStrength"`
	FleetInvestPool int                  `json:"fleetInvestPool"` // 造艦投資餘數池(見 session.go advanceAI)
	Relation        int                  `json:"relation"`
	StanceName      string               `json:"stanceName"`
	OwnedStars      int                  `json:"ownedStars"`
	ColonyStars     []int                `json:"colonyStars"` // 見 shell.AIOpponent.ColonyStars 註解
}

// sessionSnapshot 是 GameSession 的完整可序列化狀態(排除純顯示的暫態:LastEvent/LastAntares
// /LastBattle/LastPlayerOutput,它們下一回合會重算)。含未匯出的遊戲狀態(popAccum/raceGrowthPct)。
type sessionSnapshot struct {
	Version        int                  `json:"version"`
	Turn           int                  `json:"turn"`
	Player         engine.PlayerState   `json:"player"`
	PlayerColonies []engine.ColonyState `json:"playerColonies"`
	AIPlayers      []aiSnapshot         `json:"aiPlayers"`
	Stars          []Star               `json:"stars"`
	Planets        []Planet             `json:"planets"`
	Leaders        []Leader             `json:"leaders"`
	Ships          []Ship               `json:"ships"`
	SelectedStar   int                  `json:"selectedStar"`
	Difficulty     int                  `json:"difficulty"`
	Builds         []ColonyBuild        `json:"builds"`
	FleetAtStar    int                  `json:"fleetAtStar"`
	FleetDestStar  int                  `json:"fleetDestStar"`
	FleetETA       int                  `json:"fleetETA"`
	PopAccum       []int                `json:"popAccum"`
	ColonyBuild    []map[string]bool    `json:"colonyBuildings"`
	EventSeed      int64                `json:"eventSeed"`
	AntaresRaids   int                  `json:"antaresRaids"`
	RaceIndex      int                  `json:"raceIndex"`
	RaceCombatPct  int                  `json:"raceCombatPct"`
	RaceGrowthPct  int                  `json:"raceGrowthPct"`

	// Government 是玩家政府型態(2026-07-11 士氣接線;見 GameSession.Government 欄位註解)。
	// 底層是 gamedata.MoraleGovernmentType(int-based enum),json 直接序列化成數字。
	Government gamedata.MoraleGovernmentType `json:"government"`

	// --- 地面戰入侵(見 ground_invasion.go) ---
	FleetMarines        int   `json:"fleetMarines"`
	PlayerColonyMarines []int `json:"playerColonyMarines"`
	MarineBarracksAge   []int `json:"marineBarracksAge"`
}

// snapshot 擷取 GameSession 目前狀態成可序列化快照。
func (s *GameSession) snapshot() sessionSnapshot {
	ais := make([]aiSnapshot, len(s.AIPlayers))
	for i, a := range s.AIPlayers {
		prof := ai.ProfileBalanced
		if rd, ok := a.Decider.(*ai.RemakeDecider); ok {
			prof = rd.Profile
		}
		ais[i] = aiSnapshot{Name: a.Name, Player: a.Player, Colonies: a.Colonies, Profile: prof,
			FleetStrength: a.FleetStrength, FleetInvestPool: a.FleetInvestPool,
			Relation: a.Relation, StanceName: a.StanceName, OwnedStars: a.OwnedStars,
			ColonyStars: a.ColonyStars}
	}
	return sessionSnapshot{
		Version: saveFormatVersion, Turn: s.Turn, Player: s.Player,
		PlayerColonies: s.PlayerColonies, AIPlayers: ais,
		Stars: s.Stars, Planets: s.Planets, Leaders: s.Leaders, Ships: s.Ships,
		SelectedStar: s.SelectedStar, Difficulty: s.Difficulty, Builds: s.Builds,
		FleetAtStar: s.FleetAtStar, FleetDestStar: s.FleetDestStar, FleetETA: s.FleetETA,
		PopAccum: s.popAccum, ColonyBuild: s.ColonyBuildings, EventSeed: s.EventSeed,
		AntaresRaids: s.AntaresRaids, RaceIndex: s.RaceIndex,
		RaceCombatPct: s.RaceCombatPct, RaceGrowthPct: s.raceGrowthPct,
		FleetMarines: s.FleetMarines, PlayerColonyMarines: s.PlayerColonyMarines,
		MarineBarracksAge: s.MarineBarracksAge, Government: s.Government,
	}
}

// restore 由快照重建一個 GameSession(重建 AI Decider;eventRand 由 EventSeed 惰性重建)。
func (snap sessionSnapshot) restore() *GameSession {
	ais := make([]AIOpponent, len(snap.AIPlayers))
	for i, a := range snap.AIPlayers {
		ais[i] = AIOpponent{
			Name: a.Name, Player: a.Player, Colonies: a.Colonies,
			Decider:         ai.NewRemakeDecider(a.Profile), // 由性格重建決策器
			FleetStrength:   a.FleetStrength,
			FleetInvestPool: a.FleetInvestPool,
			Relation:        a.Relation, StanceName: a.StanceName, OwnedStars: a.OwnedStars,
			ColonyStars: a.ColonyStars,
		}
	}
	return &GameSession{
		Turn: snap.Turn, Player: snap.Player, PlayerColonies: snap.PlayerColonies,
		AIPlayers: ais, Stars: snap.Stars, Planets: snap.Planets, Leaders: snap.Leaders,
		Ships: snap.Ships, SelectedStar: snap.SelectedStar, Difficulty: snap.Difficulty,
		Builds: snap.Builds, FleetAtStar: snap.FleetAtStar, FleetDestStar: snap.FleetDestStar,
		FleetETA: snap.FleetETA, popAccum: snap.PopAccum, ColonyBuildings: snap.ColonyBuild,
		EventSeed: snap.EventSeed, AntaresRaids: snap.AntaresRaids, RaceIndex: snap.RaceIndex,
		RaceCombatPct: snap.RaceCombatPct, raceGrowthPct: snap.RaceGrowthPct,
		FleetMarines: snap.FleetMarines, PlayerColonyMarines: snap.PlayerColonyMarines,
		MarineBarracksAge: snap.MarineBarracksAge, Government: snap.Government,
	}
}

// Save 把目前對局狀態寫入 path(JSON)。這是 remake 自身的存檔格式(非原版 .GAM;原版格式
// 由 internal/save 唯讀解析)。
func (s *GameSession) Save(path string) error {
	data, err := json.MarshalIndent(s.snapshot(), "", "  ")
	if err != nil {
		return fmt.Errorf("序列化存檔: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("寫入存檔 %s: %w", path, err)
	}
	return nil
}

// LoadSession 從 path 讀取 remake 存檔,回傳重建的對局。
func LoadSession(path string) (*GameSession, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取存檔 %s: %w", path, err)
	}
	var snap sessionSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("解析存檔: %w", err)
	}
	if snap.Version != saveFormatVersion {
		return nil, fmt.Errorf("存檔格式版本 %d 不相容(需 %d)", snap.Version, saveFormatVersion)
	}
	return snap.restore(), nil
}

// SaveExists 回傳 path 是否存在可讀存檔。
func SaveExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
