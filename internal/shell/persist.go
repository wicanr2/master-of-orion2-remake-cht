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
	// 會談請求(見 shell/audience.go)。舊存檔缺欄位解成 false/"" —— 正是「沒有請求」,
	// 沒有零值陷阱。
	WantsAudience  bool   `json:"wantsAudience,omitempty"`
	AudienceReason string `json:"audienceReason,omitempty"`
	ColonyStars    []int  `json:"colonyStars"` // 見 shell.AIOpponent.ColonyStars 註解
	// ColonyPlanets 見 shell.AIOpponent.ColonyPlanets。舊存檔沒有 → nil,
	// ColonyPlanetIndexOfAI 退回該星的代表行星,行為與加欄位前一致。
	ColonyPlanets []int `json:"colonyPlanets,omitempty"`
	Spies         int   `json:"spies"` // AI 派來偷玩家科技的間諜數,見 spy.go
	// Personality 是 AI 性格(見 shell.AIOpponent.Personality)。omitempty 不適用:
	// 0 是合法值(排外),舊存檔缺欄位會解成 0——那與「排外」無法區分,屬已知的相容性折衷,
	// 影響只是舊存檔的 AI 性格會一律變成排外,不會壞掉。
	Personality ai.Personality `json:"personality"`
	// LastRaidTurn 是這個 AI 上次突襲玩家的回合(見 ai_attack.go)。不存的話讀檔後
	// 每個 AI 的間隔計時器都歸零,存檔當回合可能立刻又被突襲一次。
	LastRaidTurn int `json:"last_raid_turn"`

	// ColonyBuildings 見 shell.AIOpponent.ColonyBuildings 註解。舊存檔(本欄位加入前存的檔)
	// 解碼時這裡是 nil——BombardColony 對 nil/空 map 視為「無建築」,回歸行為與加欄位前一致,
	// 不會 panic。
	ColonyBuildings []map[string]bool `json:"colonyBuildings"`

	// Leaders 見 shell.AIOpponent.Leaders 註解(#5 守方 Commando 加成)。舊存檔(本欄位加入前
	// 存的檔)解碼時這裡是 nil——commandoLeaderTier(nil) 回傳 0(無加成),回歸行為與加欄位前
	// 一致(TODO 留白時的行為),不會 panic。
	Leaders []Leader `json:"leaders"`

	// AI 主力艦隊在星圖上的位置(見 ai_fleet.go)。**不存的話讀檔後艦隊會瞬移回母星**
	// ——一支飛了八回合快到玩家家門口的艦隊,存一次檔就回去了。
	//
	// 舊存檔沒有這四個欄位:解碼出 FleetPosSet=false,advanceAIFleets 下一回合會把位置
	// 初始化到母星,行為與加欄位前一致(那時 AI 本來就沒有位置)。
	FleetStar     int  `json:"fleetStar"`
	FleetPosSet   bool `json:"fleetPosSet"`
	FleetDestStar int  `json:"fleetDestStar"`
	FleetETA      int  `json:"fleetETA"`
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
	// Fleets / SelectedFleet 是**多艦隊模型**(見 fleet.go)。
	// omitempty:2026-08-07 之前的存檔沒有這兩個欄位,解碼成 nil,由 restore 從下面那組
	// 舊欄位(Ships / FleetAtStar / …)組出唯一的一支艦隊。
	Fleets        []Fleet `json:"fleets,omitempty"`
	SelectedFleet int     `json:"selectedFleet,omitempty"`
	// ColonyRelocateTo / ShowRelocationLines 是集結點(見 relocation.go)。
	// omitempty:舊存檔沒有 → nil = 每個殖民地都沒設定,語意正確。
	// ⚠ ShowRelocationLines 反過來:舊存檔解出 false,但原版預設是**開**,
	// 所以 restore 端要把「舊檔」補成 true(見 restore)。
	ColonyRelocateTo    []int `json:"colonyRelocateTo,omitempty"`
	ShowRelocationLines bool  `json:"showRelocationLines,omitempty"`
	// ⚠ 以下到 FleetETA 為止是**舊格式,只讀不寫**:單艦隊時代的欄位。
	// 新存檔一律寫 Fleets;留著是為了讀得回 2026-08-07 之前存的檔。
	Ships        []Ship        `json:"ships,omitempty"`
	SelectedStar int           `json:"selectedStar"`
	Difficulty   int           `json:"difficulty"`
	Builds       []ColonyBuild `json:"builds"`
	// BuildQueue 是各殖民地的後續建造排隊項(原版 7 格 BUILD QUEUE,見 buildqueue.go)。
	// omitempty:2026-08-06 之前的存檔沒有這個欄位,解碼成 nil = 「佇列是空的」,語意正確。
	BuildQueue [][]ColonyBuild `json:"build_queue,omitempty"`
	// Outposts 是玩家的軍事前哨站(見 outpost.go)。omitempty:2026-08-06 之前的存檔沒有
	// 這個欄位,解碼成 nil = 「沒有前哨站」,語意正確。
	Outposts []Outpost `json:"outposts,omitempty"`
	// Monsters 是星圖上的守衛怪獸(見 monster.go)。omitempty:舊存檔沒有這個欄位,
	// 解碼成 nil = 「星圖上沒有怪獸」,與加這個系統之前的行為逐位元一致。
	Monsters []MonsterGuard `json:"monsters,omitempty"`
	// PersistentEvents 是進行中的持續型事件(見 events_persistent.go)。omitempty:
	// 舊存檔沒有這個欄位,解碼成 nil = 「沒有進行中的持續事件」,與加這個系統之前一致。
	PersistentEvents []PersistentEvent `json:"persistent_events,omitempty"`
	// CapturedPop 是累計俘虜人口(供計分,見 score.go)。omitempty:舊存檔沒有這個欄位,
	// 解碼成 0 = 「沒俘虜過」,與加這個欄位之前一致。
	CapturedPop   int               `json:"captured_pop,omitempty"`
	FleetAtStar   int               `json:"fleetAtStar"`
	FleetDestStar int               `json:"fleetDestStar"`
	FleetETA      int               `json:"fleetETA"`
	PopAccum      []int             `json:"popAccum"`
	ColonyBuild   []map[string]bool `json:"colonyBuildings"`
	EventSeed     int64             `json:"eventSeed"`
	// RuleVersion 是這局的規則版本(1.3 / 1.5)。存**版本**不存整個 RuleProfile:
	// profile 是由版本推導出來的衍生資料,存衍生資料會在新增欄位時悄悄留下舊值。
	//
	// ⚠ 2026-08-07 補:先前**完全沒存**,於是讀檔後 `s.RuleProfile` 是零值——
	// 那既不是 1.3 也不是 1.5,而是「Version=1.3 但所有數值欄位都是 0」的混種:
	// Hyper-Advanced 研究成本、電漿砲傷害、轟炸輪數、守方 Commando 加成、感測器加成、
	// 貨運現金加成全部歸零。主選單選的版本因此撐不過一次存讀檔。
	// 舊存檔沒有這個欄位 → 0 = VersionClassic13,由 restore 重建成完整的 Profile13()。
	RuleVersion gamedata.GameVersion `json:"ruleVersion"`
	// 三條長壽命亂數流已經抽了幾次(見 randstream.go)。舊存檔沒有這三個欄位 → 0 →
	// 讀回來的流從頭開始,行為與加欄位前一致。
	EventDraws     int64 `json:"eventDraws,omitempty"`
	DiscoveryDraws int64 `json:"discoveryDraws,omitempty"`
	SpyDraws       int64 `json:"spyDraws,omitempty"`
	AntaresRaids   int   `json:"antaresRaids"`
	RaceIndex      int   `json:"raceIndex"`
	// PlayerName / FlagColor 是新遊戲「命名旗色」畫面設定的帝國名與旗色。
	// ⚠ 這兩個欄位先前**完全沒有進存檔**——玩家取的帝國名與選的旗色一讀檔就消失,
	// 換回預設值。2026-08-07 補上(存檔槽列表要顯示帝國名時才發現)。舊存檔沒有這兩個
	// 欄位,解出來是零值,restore 會退回預設(見該處註解),不會壞。
	PlayerName    string `json:"playerName,omitempty"`
	FlagColor     int    `json:"flagColor,omitempty"`
	RaceCombatPct int    `json:"raceCombatPct"`
	RaceGrowthPct int    `json:"raceGrowthPct"`

	// Government 是玩家政府型態(2026-07-11 士氣接線;見 GameSession.Government 欄位註解)。
	// 底層是 gamedata.MoraleGovernmentType(int-based enum),json 直接序列化成數字。
	Government gamedata.MoraleGovernmentType `json:"government"`

	// --- 地面戰入侵(見 ground_invasion.go) ---
	FleetMarines        int   `json:"fleetMarines"`
	PlayerColonyMarines []int `json:"playerColonyMarines"`
	MarineBarracksAge   []int `json:"marineBarracksAge"`

	// PlayerColonyStars 見 GameSession 欄位註解(colonization.go/ground_invasion.go 同步維護)。
	PlayerColonyStars []int `json:"playerColonyStars"`
	// PlayerColonyPlanets 見 GameSession 欄位註解。舊存檔沒有 → nil,
	// ColonyPlanetIndex 退回該星的代表行星,行為與加欄位前一致。
	PlayerColonyPlanets []int `json:"playerColonyPlanets,omitempty"`

	// --- 勝利條件(見 council.go / antaran_victory.go)---
	Victory                   VictoryState     `json:"victory"`
	PendingCouncilElection    *CouncilElection `json:"pendingCouncilElection,omitempty"`
	CouncilMeetings           int              `json:"councilMeetings"`
	LastCouncilTurn           int              `json:"lastCouncilTurn"`
	AntaranHomeworldConquered bool             `json:"antaranHomeworldConquered,omitempty"`

	// PlayerSpies 是玩家派駐到各 AI 對手的間諜數(平行 AIPlayers),見 spy.go。
	// LastEspionage(本回合諜報結算訊息)比照 LastEvent/LastAntares/LastBattle,是下回合會
	// 重算的純顯示暫態,刻意不存檔。
	PlayerSpies []int `json:"playerSpies"`

	// MercPool/MercOfferedIdx 是傭兵領袖招募狀態(見 session.go advanceMercOffers/HireMerc)。
	// omitempty:舊存檔無此欄位時解為零值(空池),讀檔後由 advanceMercOffers 自然補回,不破壞相容。
	MercPool       []Leader `json:"mercPool,omitempty"`
	MercOfferedIdx int      `json:"mercOfferedIdx,omitempty"`

	// AIRelations 是 AI 對手彼此關係矩陣(見 GameSession.AIRelations)。omitempty:舊存檔無此欄位
	// 解為 nil,ensureAIRelations 讀檔後自然補回,不破壞相容。
	AIRelations [][]int `json:"aiRelations,omitempty"`

	// History 是逐回合國力快照(見 shell/history.go)。omitempty:舊存檔無此欄位解為 nil,
	// 之後每回合自然累積,不破壞相容(折線圖在累積足夠回合前只顯示提示)。
	History []HistoryTurn `json:"history,omitempty"`

	// Seats / ActiveSeat 是熱座多人的席位快照(見 shell/hotseat.go)。原版也把遊戲模式
	// 寫進存檔(`byte_199F3A` 在 save/load 各有一次 1 byte 的 fread/fwrite),不存的話
	// 熱座局讀回來會變成單人局、其餘真人的帝國直接消失。
	// omitempty:單人局與舊存檔無此欄位,解為 nil → HotseatEnabled() 為 false,行為不變。
	Seats      []seat `json:"seats,omitempty"`
	ActiveSeat int    `json:"activeSeat,omitempty"`

	// GalaxyAge / TechLevel 是 NEW GAME 畫面的兩個設定(見 shell.GalaxyAges / TechLevels)。
	// 星系年齡在生成完之後就不再影響任何事(星圖已經存進 Stars/Planets),存它是為了讓
	// 「這一局是什麼設定」可以回頭查、也讓未來要用到時不必再改存檔格式。
	// omitempty:舊存檔沒有這兩個欄位,GalaxyAgeSet 解成 false → galaxyAge() 退回預設,
	// 與加欄位之前的行為一致。
	GalaxyAge    gamedata.GalaxyAge `json:"galaxyAge,omitempty"`
	GalaxyAgeSet bool               `json:"galaxyAgeSet,omitempty"`
	TechLevel    int                `json:"techLevel,omitempty"`
	TechLevelSet bool               `json:"techLevelSet,omitempty"`
}

// snapshot 擷取 GameSession 目前狀態成可序列化快照。
func (s *GameSession) snapshot() sessionSnapshot {
	// 熱座:目前這一席的「活的」狀態放在頂層欄位(Player/PlayerColonies/…),
	// Seats[ActiveSeat] 停在上次換人時的快照。存檔前先同步回去,否則讀檔會退回上一輪。
	seats := s.Seats
	if s.HotseatEnabled() {
		seats = append([]seat(nil), s.Seats...) // 複製,不就地改動活的 session
		seats[s.ActiveSeat] = s.saveSeat()
	}
	ais := make([]aiSnapshot, len(s.AIPlayers))
	for i, a := range s.AIPlayers {
		prof := ai.ProfileBalanced
		if rd, ok := a.Decider.(*ai.RemakeDecider); ok {
			prof = rd.Profile
		}
		ais[i] = aiSnapshot{Name: a.Name, Player: a.Player, Colonies: a.Colonies, Profile: prof,
			FleetStrength: a.FleetStrength, FleetInvestPool: a.FleetInvestPool,
			Relation: a.Relation, StanceName: a.StanceName, OwnedStars: a.OwnedStars,
			ColonyStars: a.ColonyStars, ColonyPlanets: a.ColonyPlanets,
			Spies: a.Spies, ColonyBuildings: a.ColonyBuildings,
			Leaders: a.Leaders, Personality: a.Personality, LastRaidTurn: a.LastRaidTurn,
			WantsAudience: a.WantsAudience, AudienceReason: a.AudienceReason,
			FleetStar: a.FleetStar, FleetPosSet: a.FleetPosSet,
			FleetDestStar: a.FleetDestStar, FleetETA: a.FleetETA}
	}
	return sessionSnapshot{
		Version: saveFormatVersion, Turn: s.Turn, Player: s.Player,
		PlayerColonies: s.PlayerColonies, AIPlayers: ais,
		Stars: s.Stars, Planets: s.Planets, Leaders: s.Leaders,
		Fleets: s.Fleets, SelectedFleet: s.SelectedFleet,
		ColonyRelocateTo: s.ColonyRelocateTo, ShowRelocationLines: s.ShowRelocationLines,
		SelectedStar: s.SelectedStar, Difficulty: s.Difficulty, Builds: s.Builds,
		BuildQueue:       s.BuildQueue,
		Outposts:         s.Outposts,
		Monsters:         s.Monsters,
		PersistentEvents: s.PersistentEvents,
		CapturedPop:      s.CapturedPop,
		PopAccum:         s.popAccum, ColonyBuild: s.ColonyBuildings, EventSeed: s.EventSeed,
		RuleVersion: s.RuleProfile.Version,
		EventDraws:  s.eventRand.Draws(), DiscoveryDraws: s.discoveryRand.Draws(),
		SpyDraws:     s.spyRand.Draws(),
		AntaresRaids: s.AntaresRaids, RaceIndex: s.RaceIndex,
		PlayerName: s.PlayerName, FlagColor: s.FlagColor,
		RaceCombatPct: s.RaceCombatPct, RaceGrowthPct: s.raceGrowthPct,
		PlayerColonyMarines: s.PlayerColonyMarines,
		MarineBarracksAge:   s.MarineBarracksAge, Government: s.Government,
		PlayerColonyStars: s.PlayerColonyStars, PlayerColonyPlanets: s.PlayerColonyPlanets,
		Victory: s.Victory, PendingCouncilElection: s.PendingCouncilElection,
		CouncilMeetings: s.CouncilMeetings, LastCouncilTurn: s.lastCouncilTurn,
		AntaranHomeworldConquered: s.AntaranHomeworldConquered,
		PlayerSpies:               s.PlayerSpies,
		MercPool:                  s.MercPool,
		MercOfferedIdx:            s.MercOfferedIdx,
		AIRelations:               s.AIRelations,
		History:                   s.History,
		Seats:                     seats,
		ActiveSeat:                s.ActiveSeat,
		GalaxyAge:                 s.GalaxyAge,
		GalaxyAgeSet:              s.GalaxyAgeSet,
		TechLevel:                 s.TechLevel,
		TechLevelSet:              s.TechLevelSet,
	}
}

// restore 由快照重建一個 GameSession(重建 AI Decider;三條亂數流由 EventSeed + 抽取次數
// 快轉回原位,見 randstream.go —— 不快轉的話讀檔會把事件序列從頭重播一次)。
func (snap sessionSnapshot) restore() *GameSession {
	ais := make([]AIOpponent, len(snap.AIPlayers))
	for i, a := range snap.AIPlayers {
		ais[i] = AIOpponent{
			Name: a.Name, Player: a.Player, Colonies: a.Colonies,
			Decider:         ai.NewRemakeDecider(a.Profile), // 由性格重建決策器
			FleetStrength:   a.FleetStrength,
			FleetInvestPool: a.FleetInvestPool,
			Relation:        a.Relation, StanceName: a.StanceName, OwnedStars: a.OwnedStars,
			ColonyStars: a.ColonyStars, ColonyPlanets: a.ColonyPlanets,
			Spies: a.Spies, ColonyBuildings: a.ColonyBuildings,
			Leaders: a.Leaders, Personality: a.Personality, LastRaidTurn: a.LastRaidTurn,
			WantsAudience: a.WantsAudience, AudienceReason: a.AudienceReason,
			FleetStar: a.FleetStar, FleetPosSet: a.FleetPosSet,
			FleetDestStar: a.FleetDestStar, FleetETA: a.FleetETA,
		}
	}
	restorePlanetIDs(snap.Planets)
	// ⚠ 舊存檔沒有 Star.Wormhole,JSON 解出來是零值 **0** —— 那會讓每顆星都宣稱與星 0
	// 有蟲洞(星圖畫滿放射狀連線、艦隊到處一回合直達)。normalizeWormholes 把不合法的
	// (越界 / 自己連自己 / 單向)一律清成 -1,見 wormhole.go。
	normalizeWormholes(snap.Stars)
	// ⚠ 舊存檔沒有 Star.Orbits,解出來是 5 個零值 **0** —— 那會讓每顆星都宣稱軌道 0 上有
	// 行星 0(同蟲洞那個坑)。normalizeOrbits 依「整張表都是 0」判定是舊檔,重建成
	// 一星一行星時代的形狀(軌道 0 放同索引的行星),與該版本的實際狀態逐位元一致。
	normalizeOrbits(snap.Stars, len(snap.Planets))
	// ⚠ 存檔裡的 Seats[ActiveSeat] 與頂層的 Player/Colonies/… 是同一份資料的兩個副本
	// (snapshot 存檔前才剛同步過)。restore 之後兩邊仍一致,直到下一次 AdvanceSeat。
	out := &GameSession{
		Turn: snap.Turn, Player: snap.Player, PlayerColonies: snap.PlayerColonies,
		AIPlayers: ais, Stars: snap.Stars, Planets: snap.Planets, Leaders: snap.Leaders,
		Fleets: snap.restoredFleets(), SelectedFleet: snap.SelectedFleet,
		ColonyRelocateTo: snap.ColonyRelocateTo,
		// ⚠ 舊存檔沒有這個欄位 → 解出 false,但原版預設是**開**。
		// 用「有沒有 Fleets 欄位」判斷是不是舊檔(同 restoredFleets 的判準):
		// 舊檔一律補成開,新檔照存的值。
		ShowRelocationLines: len(snap.Fleets) == 0 || snap.ShowRelocationLines,
		SelectedStar:        snap.SelectedStar, Difficulty: snap.Difficulty,
		Builds: snap.Builds, BuildQueue: snap.BuildQueue, Outposts: snap.Outposts, Monsters: snap.Monsters,
		PersistentEvents: snap.PersistentEvents, CapturedPop: snap.CapturedPop,
		popAccum: snap.PopAccum, ColonyBuildings: snap.ColonyBuild,
		EventSeed: snap.EventSeed, AntaresRaids: snap.AntaresRaids, RaceIndex: snap.RaceIndex,
		// 舊存檔沒有這兩個欄位 → 零值:PlayerName 空字串由 UI 自行退回預設顯示,
		// FlagColor 0 正好是 FlagColors 的第一個顏色,兩者都是安全的降級。
		PlayerName: snap.PlayerName, FlagColor: snap.FlagColor,
		RaceCombatPct: snap.RaceCombatPct, raceGrowthPct: snap.RaceGrowthPct,
		PlayerColonyMarines: snap.PlayerColonyMarines,
		MarineBarracksAge:   snap.MarineBarracksAge, Government: snap.Government,
		PlayerColonyStars: snap.PlayerColonyStars, PlayerColonyPlanets: snap.PlayerColonyPlanets,
		Victory: snap.Victory, PendingCouncilElection: snap.PendingCouncilElection,
		CouncilMeetings: snap.CouncilMeetings, lastCouncilTurn: snap.LastCouncilTurn,
		AntaranHomeworldConquered: snap.AntaranHomeworldConquered,
		PlayerSpies:               snap.PlayerSpies,
		MercPool:                  snap.MercPool,
		MercOfferedIdx:            snap.MercOfferedIdx,
		AIRelations:               snap.AIRelations,
		History:                   snap.History,
		Seats:                     snap.Seats,
		ActiveSeat:                snap.ActiveSeat,
		GalaxyAge:                 snap.GalaxyAge,
		GalaxyAgeSet:              snap.GalaxyAgeSet,
		TechLevel:                 snap.TechLevel,
		TechLevelSet:              snap.TechLevelSet,
	}
	// 規則版本 → 完整 profile(存的是版本,profile 是衍生資料,見 RuleVersion 欄位註解)。
	out.RuleProfile = gamedata.Profile15()
	if snap.RuleVersion == gamedata.VersionClassic13 {
		out.RuleProfile = gamedata.Profile13()
	}
	// 三條長壽命亂數流:快轉到存檔當下的位置。不快轉的話讀檔會把事件序列從頭重播一次
	// (存檔洗事件變得毫無成本,而且網路對戰時中途讀檔的那台會與其他人分岔)。
	// 種子公式必須與各自的惰性建立處一致——寫在這裡是為了讓「存了什麼、還原什麼」看得見。
	out.eventRand = restoreRandStream(snap.EventSeed*2654435761+1, snap.EventDraws)
	out.discoveryRand = restoreRandStream(
		snap.EventSeed*6364136223846793005+1442695040888963407, snap.DiscoveryDraws)
	out.spyRand = restoreRandStream(snap.EventSeed*2654435761+7, snap.SpyDraws)
	return out
}

// restorePlanetIDs 把 2026-08-06 之前存檔裡「只有顯示字串」的行星回填成 enum 欄位(就地修改)。
// 那批存檔的 Planet.Gen 是 0(JSON 缺欄位 → 零值),回填後標成 planetGenVersion,
// 讓後續程式碼一律走 enum 路徑,不必到處判斷存檔世代。
//
// 舊生成器產不出 Swamp/Arid/Terran/Gaia/Tiny/Ultra Poor,所以回填不會遇到查不到的詞;
// 真的查不到(手改過的存檔)就落在各自的保守預設,與舊版 newColonyFromStar 的行為一致。
func restorePlanetIDs(planets []Planet) {
	for i := range planets {
		p := &planets[i]
		if p.Gen >= planetGenVersion {
			continue
		}
		// Gen 1 的存檔已經有全部 ID 欄位,只缺 Gen 2 才有的行星類別——舊生成器只產一般行星,
		// 一律回填 HABITABLE 就是它當時的實際語意,不必重解字串。
		if p.Gen >= 1 {
			p.TypeID = gamedata.HABITABLE
			p.Gen = planetGenVersion
			continue
		}
		p.TypeID = gamedata.HABITABLE
		if c, ok := climateFromDisplay(p.Climate); ok {
			p.ClimateID = c
		} else {
			p.ClimateID = gamedata.TOXIC
		}
		if g, ok := gravityFromDisplay(p.Gravity); ok {
			p.GravityID = g
		} else {
			p.GravityID = gamedata.NORMAL_G
		}
		if m, ok := mineralFromDisplay(p.Mineral); ok {
			p.MineralID = m
		} else {
			p.MineralID = gamedata.POOR
		}
		if s, ok := sizeFromDisplay(p.Size); ok {
			p.SizeID = s
		} else {
			p.SizeID = gamedata.MEDIUM_PLANET
		}
		p.Gen = planetGenVersion
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

// restoredFleets 把存檔還原成艦隊清單,並吃下**舊格式**。
//
// 2026-08-07 之前的存檔是單艦隊時代的形狀:頂層一組 Ships + FleetAtStar / FleetDestStar /
// FleetETA / FleetMarines / FleetTanks。那些欄位在新存檔裡不再寫出(omitempty),
// 讀到舊檔時就在這裡組成唯一的一支艦隊。
//
// ⚠ 判斷「是不是舊檔」用的是 `len(Fleets) == 0` 而不是版本號:版本號會被別的改動一起往上帶,
// 而**這個欄位在不在**才是這件事真正的判準。
func (snap sessionSnapshot) restoredFleets() []Fleet {
	if len(snap.Fleets) > 0 {
		return snap.Fleets
	}
	// ⚠ 舊格式**沒有存戰車營**:單艦隊時代的 snapshot 有 fleetMarines 卻沒有對應的
	// fleetTanks 欄位(2026-08-07 改多艦隊時才發現)。也就是說**舊存檔讀回來戰車營一律歸零**
	// ——那是舊格式本身的漏欄,不是這次遷移弄丟的。新格式把整個 Fleet 序列化,
	// Tanks 跟著進去,這個洞就補起來了。
	return []Fleet{{
		Ships: snap.Ships, AtStar: snap.FleetAtStar, DestStar: snap.FleetDestStar,
		ETA: snap.FleetETA, Marines: snap.FleetMarines,
	}}
}
