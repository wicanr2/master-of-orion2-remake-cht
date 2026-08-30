// Package shell 是「可玩遊戲殼」的純邏輯核心:活的對局狀態、輸入命中判定。
// 不 import ebiten(維持可純測);ebiten 的繪製與輸入輪詢在 cmd/moo2。
package shell

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/diplomacy"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// AIOpponent 是一個由 AI 操控的對手帝國。
type AIOpponent struct {
	Name string
	// Color 是原版玩家記錄 +0x26 的 CMBTSHP 色塊索引；未知／舊存檔沿用
	// demo 的相異色塊 fallback。ColorKnown 用來保留「raw color=0」這個合法值。
	Color      int  `json:"color,omitempty"`
	ColorKnown bool `json:"colorKnown,omitempty"`
	// RaceIndex 是這個 AI 接管成真人時的種族索引。舊存檔缺欄位時為 0
	// (人類),不影響原本只把 AI 當對手的路徑。
	RaceIndex int
	// LuckyEventCounter 對應原版 player+0xE73；事件排程是全帝國掃描，
	// 因此 AI Lucky 種族也必須持有自己的累積值，不能只替目前真人計數。
	LuckyEventCounter int `json:"luckyEventCounter,omitempty"`
	// PopulationRaceSlot 是原版 packed colonist 低四位所用的當局 player slot；
	// 不是 RaceIndex／OrigIdx。零為合法值，Known 區分舊 JSON。
	PopulationRaceSlot      int
	PopulationRaceSlotKnown bool
	// CapitolPlanet 對應原版 player+0x29；Known 區分合法行星 0 與舊存檔缺欄。
	CapitolPlanet          int
	CapitolPlanetKnown     bool
	CapitolRebuildRequired bool
	Player                 engine.PlayerState
	Colonies               []engine.ColonyState
	Decider                ai.Decider
	FleetStrength          int             // 由 Ships 的非支援艦艦體強度推導；舊存檔／舊測試純量相容欄位
	FleetInvestPool        int             // 舊存檔相容欄位；新造艦進度使用 ShipBuildProgress
	ShipDesigns            []ShipBlueprint `json:"shipDesigns,omitempty"`
	Ships                  []Ship          `json:"ships,omitempty"`
	ShipBuildProgress      int             `json:"shipBuildProgress,omitempty"`
	// ColonyBuilds 以星系索引保存 AI 各殖民地目前產品；原版 Colony_AI_ 是逐殖民地
	// 指派產品，不是把全帝國工業直接灌入單一造艦池。
	ColonyBuilds map[int]ColonyBuild `json:"colonyBuilds,omitempty"`
	// Personality 是原版 AI 性格(AIRACES.CFG race_personality 0-6),開局依種族分布抽出。
	// 驅動關係演化、擴張積極度等行為差異——先前三個 AI 除了名字之外行為完全相同,
	// 因為所有性格相關的數字都是硬編的固定值(見 ai/personality_tables.go)。
	Personality ai.Personality
	// OriginalTechProfile 是 sub_589D6 建立、供 sub_FC845／sub_FD335 常態研究選擇消費的
	// raw6／raw4／raw7 與 runtime 種族特性。Known 區分舊存檔的合法零值。
	OriginalTechProfile      gamedata.OriginalAITechProfile `json:"originalTechProfile,omitempty"`
	OriginalTechProfileKnown bool                           `json:"originalTechProfileKnown,omitempty"`
	// OriginalRaw28／Known 單獨保存 player+0x28。它也是 OriginalTechProfile.Raw6，
	// 但 GAM 目前尚未解析 +0x205／+0x206，不能為了其中一欄就把完整 profile 冒稱 known。
	OriginalRaw28      int  `json:"originalRaw28,omitempty"`
	OriginalRaw28Known bool `json:"originalRaw28Known,omitempty"`
	Relation           int  // 對玩家的 normalized 外交關係（-40..40）
	// OriginalRelationRaw 保存原版 player+0x617 signed byte 的 -100..100 餘數；
	// Known 區分合法 raw 0 與舊存檔缺欄。UI／AI 仍只消費 Relation。
	OriginalRelationRaw   int  `json:"originalRelationRaw,omitempty"`
	OriginalRelationKnown bool `json:"originalRelationKnown,omitempty"`
	// OriginalRelationTargetRaw 保存 sub_4D78E 由 byte_180ED4 初始化的
	// player+0x61F 目標。現有單向欄位表示「AI 觀察玩家」方向。
	OriginalRelationTargetRaw   int    `json:"originalRelationTargetRaw,omitempty"`
	OriginalRelationTargetKnown bool   `json:"originalRelationTargetKnown,omitempty"`
	StanceName                  string // 目前對玩家態勢(中文;由 ai.DecideStance 推得)
	// Treaty 是玩家與這個 AI 的正式外交／經濟協議狀態。正式狀態與貿易、研究
	// 旗標分開，對應原版 +0x627、+0x62F、+0x637 的資料形狀。
	Treaty     TreatyState
	OwnedStars int // 已擴張佔領的星數(含母星)
	// ExploredStars 對應原版逐星 star+0x33 中屬於此帝國的位元。母星建立、艦隊抵達
	// 與取得殖民地時設位；Known 區分新局的完整歷史與無法還原既往造訪的舊 JSON。
	ExploredStars      []bool `json:"exploredStars,omitempty"`
	ExploredStarsKnown bool   `json:"exploredStarsKnown,omitempty"`

	// --- AI 主力艦隊在星圖上的位置(2026-08-08 第 47 項(AI艦隊移動))---
	//
	// 先前 AI 沒有位置,突襲是瞬移的:玩家看不到它來、阿提米絲水雷也打不到它。
	// 現在它會從所在的星飛到目標,抵達才動手。見 ai_fleet.go。
	//
	// ⚠ **FleetStar 的零值 0 是合法的星索引**,所以另立 FleetPosSet 當「有沒有設過」的旗標
	// ——舊存檔解出來 FleetPosSet=false,advanceAIFleets 會把它初始化到母星。
	FleetStar     int
	FleetPosSet   bool
	FleetDestStar int // 只在 FleetETA > 0 時有意義
	FleetETA      int // 0 = 靜止
	// FleetTargetAI 是可選 AI 對 AI 戰爭的目標；FleetTargetAISet 避免舊存檔
	// 的零值 0 被誤讀成「攻擊第 0 個 AI」。沒有目標時以 -1 表示。
	FleetTargetAI    int
	FleetTargetAISet bool

	// Spies 是這個 AI 對手派來偷玩家科技的間諜數(見 spy.go advanceEspionage)。opt-in,
	// 新對局預設 0(Go 零值恰好是想要的預設值,無零值陷阱)。AI 目前用簡單週期政策自動增加
	// (見 advanceAI),不像玩家的 PlayerSpies 需要花 BC 呼叫 TrainSpy——AI 的訓練成本/BC
	// 限制未建模,是誠實簡化而非疏漏(見 spy.go 檔頭說明)。
	Spies int
	// DefensiveAgents 是駐守本帝國、對所有來犯間諜共用的 Agent 數量。與
	// Spies 分開，符合手冊 Spy(進攻)／Agent(防守) 的兩種 slot。
	DefensiveAgents int

	// LastRaidTurn 是這個 AI 上次對玩家發動突襲的回合(見 ai_attack.go)，只供
	// remake 的「停在同星不可每回合重複結算」adapter；出兵決策改讀原版 cooldown。
	LastRaidTurn int
	// OriginalHumanTargetDecisionCooldown 對應原版 player+0x816（decimal +2070）。
	// sub_53EDB 只在它為 0 時評估真人目標；類型 2 成功後寫 Random_(20)+20。
	OriginalHumanTargetDecisionCooldown int `json:"originalHumanTargetDecisionCooldown,omitempty"`
	// OriginalHumanContactTurns 對應 AI→目前真人方向的 player+0x88F；接觸後每回合遞增，封頂 250。
	OriginalHumanContactTurns int `json:"originalHumanContactTurns,omitempty"`

	// WantsAudience 是「這位對手正在請求會談」(原版 `Humans_Requesting_Diplomacy_` 那個
	// 位元遮罩裡屬於它的那一位)。AudienceReason 是來意(宣戰/提議貿易/提議結盟)。
	// 見 audience.go。opt-in,零值 false = 沒有請求(新對局/舊存檔皆安全)。
	WantsAudience  bool
	AudienceReason string
	// OriginalHumanDiplomaticRequest 保存 sub_53EDB outcome 1／3／4 與 sub_54CC0 payload。
	// nil 表示沒有該類原版請求；UI 尚未接受前不得先套用推測性的條約／資產效果。
	OriginalHumanDiplomaticRequest *gamedata.OriginalHumanDiplomaticRequest `json:"originalHumanDiplomaticRequest,omitempty"`
	// OriginalHumanDirectRequestTier 對映 sub_52049 的方向 direct-request tier 1／2。
	OriginalHumanDirectRequestTier int `json:"originalHumanDirectRequestTier,omitempty"`
	// OriginalHumanMilitaryCandidate* 對應 AI→目前真人方向 +0x837／+0x887；reason 106
	// 被拒時搬到 +0x7C7／+0x7C9。Known 區分合法 -1（無目標，轉為宣戰）與尚未刷新；
	// 現行 producer 是單主力艦隊的強推論近似，並不冒稱原版 sub_D94B3 多艦隊搜尋 exact。
	OriginalHumanMilitaryCandidateStar   int  `json:"originalHumanMilitaryCandidateStar,omitempty"`
	OriginalHumanMilitaryCandidateReason int  `json:"originalHumanMilitaryCandidateReason,omitempty"`
	OriginalHumanMilitaryCandidateKnown  bool `json:"originalHumanMilitaryCandidateKnown,omitempty"`
	OriginalHumanMilitaryTargetStar      int  `json:"originalHumanMilitaryTargetStar,omitempty"`
	OriginalHumanMilitaryTargetReason    int  `json:"originalHumanMilitaryTargetReason,omitempty"`
	OriginalHumanMilitaryTargetKnown     bool `json:"originalHumanMilitaryTargetKnown,omitempty"`

	// ColonyStars 是 Colonies[i] 對應到 Stars 的索引（平行陣列），兩者長度須一致。
	// aiExpand 透過 newColonyFromStar 建立真殖民地並同步 append；InvadeColony 攻陷時同步
	// 移除，因此 AI 後續殖民地可被入侵且會進入每回合帝國經濟。
	ColonyStars []int

	// ColonyPlanets 是 Colonies[i] 座落的**行星**索引(平行陣列,對 GameSession.Planets),
	// 語意同玩家的 PlayerColonyPlanets。−1 = 舊存檔沒記,只知道在哪顆星。
	//
	// ⚠ AI 目前一個星系仍然只會有一個殖民地(aiExpand 只找 `Owner == 0` 的星),不像玩家
	// 可以在自己的星系裡拓殖第二顆行星。這個欄位先把資料模型補齊,AI 的多殖民地擴張是
	// 另一件事(記在 docs/re/01-gap-report.md)。
	ColonyPlanets []int

	// ColonyBuildings 是 Colonies[i] 對應的已完工建築集合(平行陣列,比照 Colonies/ColonyStars
	// 兩者的長度不變量——三者長度須恆一致)。2026-07-11 新增:讓 AI 對手的殖民地也有建築資料
	// 可扣(見 orbital_bombardment.go BombardColony「軌道防禦建築吸收軌道轟炸」),補齊先前
	// 「AI 完全沒有建築欄位,轟炸只能扣人口」的資料模型缺口。
	//
	// 同步時機(逐一核對,勿遺漏):
	//   - buildDemoAIOpponents:每個 AI 母星初始化為 homeworldBuildings() 的獨立拷貝
	//     (cloneBuildings)——不可共享同一個 map 參考,否則轟炸掉一個 AI 的星基會連動到共用
	//     同一份 map 的其他 AI。
	//   - aiExpand:新殖民地 append 空 map(map[string]bool{}),不是 homeworldBuildings() 的
	//     拷貝——手冊只保證母星有星基,新拓殖星沒有,故新 AI 殖民地開局無建築。
	//   - InvadeColony:玩家攻陷 AI 殖民地移除該筆 Colonies/ColonyStars 時,同步移除對應的
	//     ColonyBuildings[colonyIdx](三者一起從陣列中刪除,維持等長)。
	//
	// nil 安全:舊存檔沒有這個欄位時解碼為 nil,BombardColony 對 nil/空 map 視為「無建築」,
	// 行為與加這個欄位之前逐位元一致(hits 全部進人口,不會 panic)。
	ColonyBuildings []map[string]bool

	// ColonyMarines／ColonyTanks 是 AI 每座殖民地的駐軍池；後兩個平行陣列記錄兵營
	// 已運作回合。它們對稱於玩家側的 PlayerColonyMarines／PlayerColonyTanks，讓
	// AI 守方的 Armor Barracks 不再以「只看目前回合」猜測。舊存檔沒有這四欄時，
	// 第一次使用會依現有建築以 age=0 建立安全的初始值；之後只依建築公式補充，
	// 不會把已在戰鬥中損失的部隊憑空重算回來。
	//
	// 平行陣列同步點：buildDemoAIOpponents、aiExpand、transferAIColony、
	// OfferStarGift、MindControlColony、InvadeColony 及 GAM 匯入／移除路徑。
	ColonyMarines     []int
	ColonyTanks       []int
	MarineBarracksAge []int
	ArmorBarracksAge  []int

	// Leaders／LeaderOffer／LeaderLastOfferTurn 對應原版全域英雄表與每帝國
	// +0x5A1/+0xE6A。AI 會在回合鏈中動態接受、拒絕與任命，不再依種族固定塞 Commando。
	// ColonyLeaderNames 與 Colonies 平行，保存管理領袖的 typed 任命結果。
	Leaders             []Leader
	LeaderOffer         *Leader
	LeaderLastOfferTurn int
	ColonyLeaderNames   []string
	// OriginalFoodDeficitTurns 對應原版 player+0x7EC：帝國食物結餘
	// player+0xB0 為負時逐回合累加，否則歸零；sub_25DF1 以此產生 reason 113。
	OriginalFoodDeficitTurns int `json:"originalFoodDeficitTurns,omitempty"`
	// OriginalWarFlag60ERaw 保存原版 Player+0x60E；consumer 已證實，producer 未知。
	// 只從原版 GAM 或 remake JSON 延續，不為新局自行推導。
	OriginalWarFlag60ERaw int `json:"originalWarFlag60ERaw,omitempty"`
	// OriginalBlockadeGrievanceRaw 對應 AI 看向目前真人的 player+0x6BF 方向值；
	// 人類艦隊在戰時封鎖 AI 殖民星時，由 Change_Relations_ reason raw 7 累加。
	OriginalBlockadeGrievanceRaw int `json:"originalBlockadeGrievanceRaw,omitempty"`
	// OriginalHumanBetrayalRaw 對應 AI 看向目前真人方向的 player+0x727；
	// 玩家破壞既有正式條約後永久設為 true。
	OriginalHumanBetrayalRaw bool `json:"originalHumanBetrayalRaw,omitempty"`
	// OriginalHumanTreatyGrievanceRaw／VictimRaw 對應 AI→目前真人方向的
	// signed +0x7EE／+0x7F6。Known 區分合法 victim slot 0 與舊 JSON 缺欄。
	OriginalHumanTreatyGrievanceRaw int  `json:"originalHumanTreatyGrievanceRaw,omitempty"`
	OriginalHumanTreatyVictimRaw    int  `json:"originalHumanTreatyVictimRaw,omitempty"`
	OriginalHumanTreatyVictimKnown  bool `json:"originalHumanTreatyVictimKnown,omitempty"`
	// OriginalHumanIncidentMemoryRaw／ReasonRaw 對應 AI→真人 +0x71F／+0x6CF。
	// Known=false 的 GAM／舊 JSON 不可把缺欄當成原版精確零。
	OriginalHumanIncidentMemoryRaw           int  `json:"originalHumanIncidentMemoryRaw,omitempty"`
	OriginalHumanIncidentReasonRaw           int  `json:"originalHumanIncidentReasonRaw,omitempty"`
	OriginalHumanIncidentKnown               bool `json:"originalHumanIncidentKnown,omitempty"`
	OriginalHumanIncidentPendingReasonRaw    int  `json:"originalHumanIncidentPendingReasonRaw,omitempty"`
	OriginalHumanIncidentPendingMagnitudeRaw int  `json:"originalHumanIncidentPendingMagnitudeRaw,omitempty"`
}

// cloneBuildings 回傳 m 的獨立拷貝(逐鍵複製),供需要「各自獨立、不共享底層 map」的初始化
// 情境使用(例如每個 AI 對手各自的 ColonyBuildings[0],若直接共用同一個 homeworldBuildings()
// 回傳值會導致轟炸掉一個 AI 的建築連動影響其他 AI——map 是參考型別,共享會出這種隱性 bug)。
// m 為 nil 時回傳 nil(不創建空 map,維持與「這個殖民地本來就沒有建築資料」一致的語意)。
func cloneBuildings(m map[string]bool) map[string]bool {
	if m == nil {
		return nil
	}
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// Star 是星系圖上的一顆星(供星圖渲染;正規化座標 0..1)。
type Star struct {
	X, Y     float64 // 0..1 正規化位置
	Spectral int     // 0=藍 1=白 2=黃 3=橙 4=紅 5=棕 6=黑洞
	Size     int     // 0=大 .. 3=小
	Name     string  // 目前語言的顯示名
	NameEN   string  // 原版英文星名；供英文動態報告與存檔相容回退
	Owner    int     // 0=無主 1=玩家 2=AI
	Explored bool    // 艦隊是否曾抵達(已探索)
	// Orbits 是這個星系 5 個軌道上的行星索引(OrbitEmpty = 空),對應原版
	// `word[星×0x71 + 0x4A + 軌道×2]`。見 orbit.go(含「5 個軌道」的三個來源)。
	//
	// ⚠ 舊存檔沒有這個欄位,零值是 5 個 0 —— 那會讓每顆星都宣稱軌道 0 上有行星 0。
	// 讀檔路徑一律走 `normalizeOrbits`(同 Wormhole 那個坑)。
	Orbits [StarOrbits]int
	// Wormhole 是蟲洞另一端的星索引,-1 = 沒有蟲洞。
	//
	// 對應原版星球結構 +0x29(int8,0xFF = 無)。**必須是雙向的**——openorion2
	// `gamestate.cpp:1946` 直接對單向蟲洞丟例外("One-way wormholes not allowed"),
	// 原版 `Draw_Wormhole_Links_` 也是兩端各畫一次線。
	//
	// ⚠ 舊存檔沒有這個欄位,零值是 0 而不是 -1 —— 那會讓每顆星都宣稱與星 0 有蟲洞。
	// 讀檔路徑一律走 `normalizeWormholes`(見 persist.go)。
	Wormhole int
	// InNebula 是「這顆星落在星雲內」。對應原版星球結構 +0x6F
	// (`Initialize_Star_In_Nebula_Info_` @ 0xEBA96 逐星寫入),也是 `internal/save`
	// 既有的 `Star.InNebula` 欄位。判定要讀星雲圖的遮罩,見 nebula.go。
	//
	// 零值 false 是安全的(= 不在星雲內),舊存檔不需要修補。
	InNebula bool
	// BlockadedMask 對應原版 star+0x2A：每一 bit 是被封鎖的 player slot。
	// BlockadedBy[slot] 對應 star+0x2B+slot：每一 bit 是封鎖該 slot 的艦隊 owner。
	// 兩者由原版每回合主鏈先消費舊值、移動艦隊後再整表重算；不是單一 bool。
	BlockadedMask uint8
	BlockadedBy   [8]uint8
}

// ShipWeaponMount 保存原版 8-byte weapon record 的 typed runtime 對應。
// RawType=-1 表示 remake 自建元件尚未對回原版 raw ID；其餘欄位均可 JSON 往返。
type ShipWeaponMount struct {
	RawType                int `json:"rawType"`
	Name                   string
	MaxCount, WorkingCount int
	Arc                    gamedata.WeaponArc
	RawMods                uint16
	Mods                   []string `json:"mods,omitempty"` // remake 已解碼改造；RawMods 仍原樣保存
	Ammo, Attack           int
}

func cloneWeaponMounts(in []ShipWeaponMount) []ShipWeaponMount {
	out := append([]ShipWeaponMount(nil), in...)
	for i := range out {
		out[i].Mods = append([]string(nil), out[i].Mods...)
	}
	return out
}

// Ship 是一艘艦艇(供艦隊畫面);Weapon/Armor/Shield/Special 為舊單槽相容欄位。
type Ship struct {
	Name                           string
	Class                          string // 艦體等級(護衛艦/巡洋艦/戰艦…)
	Weapon, Armor, Shield, Special string // 元件名
	// RawType／RawMission／ProductionCost 保存原版 129-byte Ship 紀錄中，破產強制拆船
	// 會直接消費的欄位。Known 旗標讓合法零值不會與舊 JSON 的缺欄混淆；remake 新造艦
	// 會寫入可確定的 type 與實際造價，舊存檔則由 Class／ShipCost 安全回退。
	RawType         gamedata.ShipType `json:"rawType,omitempty"`
	RawTypeKnown    bool              `json:"rawTypeKnown,omitempty"`
	RawMission      uint8             `json:"rawMission,omitempty"`
	RawMissionKnown bool              `json:"rawMissionKnown,omitempty"`
	ProductionCost  int               `json:"productionCost,omitempty"`
	// 原版 129-byte Ship／內嵌 ShipDesign 中，+0x10 computer 與 +0x6E..+0x7D
	// 的分離損傷／crew 欄位會被 +0x5EC 國力 producer 直接消費。Known 旗標區分
	// 合法 raw 0 與舊 JSON；Damage/CrewXP 仍是 remake 正常玩法的相容欄位。
	ComputerRaw          uint8    `json:"computerRaw,omitempty"`
	ComputerRawKnown     bool     `json:"computerRawKnown,omitempty"`
	DesignSizeRaw        uint8    `json:"designSizeRaw,omitempty"`
	DesignSizeRawKnown   bool     `json:"designSizeRawKnown,omitempty"`
	ArmorRaw             uint8    `json:"armorRaw,omitempty"`
	ArmorRawKnown        bool     `json:"armorRawKnown,omitempty"`
	ShieldRaw            uint8    `json:"shieldRaw,omitempty"`
	ShieldRawKnown       bool     `json:"shieldRawKnown,omitempty"`
	BaseCombatSpeedRaw   uint8    `json:"baseCombatSpeedRaw,omitempty"`
	BaseCombatSpeedKnown bool     `json:"baseCombatSpeedRawKnown,omitempty"`
	ShieldDamageRaw      uint8    `json:"shieldDamageRaw,omitempty"`
	DriveDamageRaw       uint8    `json:"driveDamageRaw,omitempty"`
	ComputerDamageRaw    uint8    `json:"computerDamageRaw,omitempty"`
	CrewLevelRaw         uint8    `json:"crewLevelRaw,omitempty"`
	CrewLevelRawKnown    bool     `json:"crewLevelRawKnown,omitempty"`
	ArmorDamageRaw       int      `json:"armorDamageRaw,omitempty"`
	StructureDamageRaw   int      `json:"structureDamageRaw,omitempty"`
	OriginalDamageKnown  bool     `json:"originalDamageKnown,omitempty"`
	DamagedSpecialsRaw   [5]uint8 `json:"damagedSpecialsRaw,omitempty"`
	// WeaponMounts 與 SpecialIDs 保存原版多槽 blueprint。快速與格子戰術已逐槽消費武器；
	// importer／建造仍把第一個有效槽同步到相容欄位，供舊存檔與其他顯示使用。
	WeaponMounts []ShipWeaponMount  `json:"weaponMounts,omitempty"`
	SpecialIDs   []int              `json:"specialIDs,omitempty"`
	Specials     []ShipSpecialMount `json:"specials,omitempty"`
	// CombatPicture 是原版艦艇記錄 +0xC4 的 raw picture 欄位。只有從 `.GAM`
	// 或其他原版設計資料讀到它時才把 CombatPictureKnown 設為 true；零是合法
	// picture，不能用零值本身判斷「未知」。
	CombatPicture      int  `json:"combatPicture,omitempty"`
	CombatPictureKnown bool `json:"combatPictureKnown,omitempty"`
	// Arc 是目前單一武器掛載的火線角。零值代表舊 JSON 沒有這欄，讀取／戰鬥
	// 邊界會依武器類型正規化；原版 save.ShipDesign.Weapons[i].Arc 是 uint8。
	Arc gamedata.WeaponArc `json:"arc,omitempty"`
	// OfficerName 是這艘船目前指派的艦艇軍官名稱;空字串=未指派。
	// 原版 save.Ship 對應欄位是 int16 Officer;remake 同時保存 OfficerID，對應
	// HERODATA／原版 `_leaders[]` 的 0..66 來源序號。名稱保留給舊 JSON 與人工編輯
	// 的回退；指派與技能查詢見 officer_assignment.go。
	OfficerName           string `json:"officer,omitempty"`
	OfficerID             int    `json:"officer_id,omitempty"`
	WeaponAttack, BonusHP int    // 武器攻擊加成、裝甲+護盾 HP 加成
	// Mods 是掛載在 Weapon 上的武器改造(gamedata.WeaponModCode 字串,如 "HV"/"PD"),
	// 只套用該武器類型支援的改造(見 WeaponModOptionsForWeapon)。空切片/nil = 無改造(既有
	// 存檔沒有這個欄位,JSON 解碼會是 nil,行為與「無改造」完全一致,回歸安全)。
	Mods []string
	// WeaponAmmo 是原版設計槽的 Ammo。零值代表舊 JSON；進入設計／戰鬥時由
	// NormalizeWeaponAmmo 依武器回退（標準飛彈 5、魚雷 2、光束 255）。
	WeaponAmmo int `json:"weapon_ammo,omitempty"`
	// Damage 是累積的結構損傷(0 = 完好)。remake 先前沒有艦艇損傷的概念——一艘船不是完好
	// 就是被擊沉——「自動修復」元件因此從加進來就沒有任何效果。見 repair.go。
	// 舊存檔沒有這個欄位,JSON 解碼為 0 = 完好,回歸安全。
	Damage int
	// CrewXP 是艦員累積經驗(手冊 p.121)。**等級不另存**——由 CrewXP 推導
	// (見 GameSession.shipCrewLevel)。兩個欄位遲早會不同步,一個不會。
	//
	// 太空學院「造出來的船起始等級 +1」就是把這個欄位設成那一級的門檻
	// (gamedata.CrewXPForLevel),不是另開一個「起始等級」欄位。
	// 舊存檔沒有這個欄位,JSON 解碼為 0 = 新兵,回歸安全。
	CrewXP int
}

// Component 是一個艦艇元件(名稱 + 成本 + 效果值 + 解鎖科技)。
type Component struct {
	Name  string
	Cost  int
	Value int                    // 武器=攻擊、裝甲/護盾=HP、特殊=攻擊或視元件而定
	Tech  gamedata.ResearchTopic // 解鎖所需研究主題(0=起始科技,一開始就有)
	// UnlockTech 是該元件真正對應的 MOO2 科技(0=TECH_NONE=未映射/里程碑/抽象元件,走主題層級)。
	// 校正依據見 docs/tech/component-tech-mapping.md。多選主題中,唯有玩家明確抉擇到此科技才解鎖。
	UnlockTech gamedata.Technology
}

// 元件清單(名稱取自 MOO2 真實科技譯名 tech.tsv;成本/效果依 MOO2 遞進,各標解鎖科技)。
// 涵蓋完整武器/裝甲/護盾/特殊進程,進階元件需先研究對應主題。
var (
	// Value = 單裝武器最大傷害。標 ✓ 者取自 patch 1.5 官方文件(MANUAL_150.html)確認值:
	// 中子爆破槍 12、高斯砲 18、電漿砲 20(1.50;1.31 為 30,版本相依)。其餘為依科技階遞增
	// 的單調估計,精確值待掃描版手冊武器附錄 OCR 交叉核對。詳見 docs/tech/component-values.md。
	// Tech/UnlockTech 經 docs/tech/component-tech-mapping.md 對真科技樹校正:
	// 掛正確主題 + 真 Technology。里程碑科技(死光/氙素裝甲)與抽象元件(戰鬥電腦/重生程序)
	// 真科技樹無單一 TOPIC 可掛,暫掛簡化 proxy 主題、UnlockTech=TECH_NONE(走主題層級,標註待重設計)。
	WeaponOptions = []Component{
		{"無武裝", 0, 0, 0, 0},
		// ⚠ 2026-08-08(第 64 項(武器傷害真表)):Value(最大傷害)全部換成**手冊 p.124-125 的真值**。
		// 先前是「依科技階遞增的單調估計」,而武器線本來就不是單調的——核融合光束在
		// 舊估計裡比中子爆破槍強(16 > 12),手冊上它比中子爆破槍**弱**(6 vs 12)。
		// 逐項出處與偏差幅度見 gamedata/weapon_damage.go。Cost 欄不動(那是 remake 的
		// 生產成本尺度,與手冊的 Cost 欄不是同一個單位)。
		{"雷射", 20, 4, gamedata.TOPIC_PHYSICS, gamedata.TECH_LASER_CANNON},                      // 手冊 1-4
		{"核飛彈", 30, 8, gamedata.TOPIC_CHEMISTRY, gamedata.TECH_NUCLEAR_MISSILE},                // 手冊 8(先前 6)
		{"質量投射器", 40, 6, gamedata.TOPIC_ADVANCED_MAGNETISM, gamedata.TECH_MASS_DRIVER},         // 手冊 6(先前 8)
		{"中子爆破槍", 60, 12, gamedata.TOPIC_NEUTRINO_PHYSICS, gamedata.TECH_NEUTRON_BLASTER},      // 手冊 3-12
		{"核融合光束", 80, 6, gamedata.TOPIC_FUSION_PHYSICS, gamedata.TECH_FUSION_BEAM},             // 手冊 2-6(先前 16)
		{"麥克萊特飛彈", 90, 14, gamedata.TOPIC_ADVANCED_CHEMISTRY, gamedata.TECH_MERCULITE_MISSILE}, // 手冊 14(先前 17)
		{"高斯砲", 120, 18, gamedata.TOPIC_SUBSPACE_FIELDS, gamedata.TECH_GAUSS_CANNON},           // 手冊 18
		{"相位砲", 160, 20, gamedata.TOPIC_MULTIPHASED_PHYSICS, gamedata.TECH_PHASOR},             // 手冊 5-20(先前 19)
		{"電漿砲", 200, 20, gamedata.TOPIC_PLASMA_PHYSICS, gamedata.TECH_PLASMA_CANNON},           // 手冊 4-20(1.50)
		// 死光:⚠ **`UnlockTech` 先前是 0,主題掛在人造生命** —— 兩個都不對。
		// 執行檔的武器表給 `TECH_DEATH_RAY`(47),而手冊把它放在「Xenon Technologies」那一節,
		// 與氙素裝甲同一節。這一格是英文顯示名的測試抓到的(UnlockTech=0 就推導不出英文名)。
		//
		// ⚠ **仍未解決的忠實度問題**:手冊那一節開頭寫著「only known to the enigmatic Antarans
		// (and Orions) and **cannot be discovered via the normal course of research**」——
		// 也就是死光與氙素裝甲在原版**不能靠研究拿到**,要從安塔蘭/獵戶座手上奪。
		// remake 現在把 TOPIC_XENON_TECHNOLOGY 當成一般可研究主題(氙素裝甲那一列也是),
		// 那是**既有的偏差**,不是這一行引入的。要修得動科技樹的可研究集合,列進待辦。
		{"死光", 300, 100, gamedata.TOPIC_XENON_TECHNOLOGY, gamedata.TECH_DEATH_RAY}, // 手冊 50-100

		// ⚠ 2026-08-08(第 64 項(武器傷害真表))補上手冊有、remake 沒有的八項。
		//
		// 傷害/佔格取自手冊 p.124-125;**研究主題取自執行檔**(`gamedata.OrigTechTopic`,
		// 第 49 項(安塔蘭防禦艦隊)那張 211/212 對得上的表)——不是照科技名猜的,有測試逐項核對。
		//
		// Cost 是 remake 的生產成本尺度(與手冊的 Cost 欄不同單位,見第 64 項(武器傷害真表)),
		// 依手冊成本的**相對名次**插在既有鄰居之間——**那是 remake 的選擇,不是手冊值**。
		{"離子脈衝砲", 100, 10, gamedata.TOPIC_ION_FISSION, gamedata.TECH_ION_PULSE_CANNON},
		{"引力波束", 140, 15, gamedata.TOPIC_ARTIFICIAL_GRAVITY, gamedata.TECH_GRAVITON_BEAM},
		{"質子魚雷", 150, 40, gamedata.TOPIC_HYPER_DIMENSIONAL_FISSION, gamedata.TECH_PROTON_TORPEDOES}, // 手冊 40
		{"電漿魚雷", 360, 120, gamedata.TOPIC_INTERPHASED_FISSION, gamedata.TECH_PLASMA_TORPEDOES},      // 手冊 120;每格 −5,NR 可取消
		{"脈衝飛彈", 170, 20, gamedata.TOPIC_MOLECULAR_COMPRESSION, gamedata.TECH_PULSON_MISSILE},
		{"氙素飛彈", 220, 30, gamedata.TOPIC_MOLECULAR_MANIPULATION, gamedata.TECH_ZEON_MISSILE},
		{"干擾者", 260, 40, gamedata.TOPIC_MULTIDIMENSIONAL_PHYSICS, gamedata.TECH_DISRUPTER_CANNON},
		{"粒子束", 280, 30, gamedata.TOPIC_XENON_TECHNOLOGY, gamedata.TECH_PARTICLE_BEAM},
		{"重錘裝置", 340, 100, gamedata.TOPIC_HYPER_DIMENSIONAL_PHYSICS, gamedata.TECH_MAULER_DEVICE},

		// 魚雷三件組(手冊 p.125；傷害與佔格由 weapon_table.go 交叉核對)。
		{"反物質魚雷", 130, 25, gamedata.TOPIC_ANTIMATTER_FISSION, gamedata.TECH_ANTIMATTER_TORPEDOES},

		// 炸彈(第 64 項(武器傷害真表),手冊 p.126 的 BOMB 表)。**只能打行星**——見 WeaponKindBomb。
		// 主題同樣取自執行檔;四項的執行檔 category 都是 19(炸彈),與手冊分類一致。
		{"核彈", 40, 12, gamedata.TOPIC_NUCLEAR_FISSION, gamedata.TECH_NUCLEAR_BOMB},
		{"融合彈", 90, 24, gamedata.TOPIC_ADVANCED_FUSION, gamedata.TECH_FUSION_BOMB},
		{"反物質彈", 180, 40, gamedata.TOPIC_ANTIMATTER_FISSION, gamedata.TECH_ANTIMATTER_BOMB},
		{"中子彈", 240, 60, gamedata.TOPIC_INTERPHASED_FISSION, gamedata.TECH_NEUTRONIUM_BOMB},

		// 球形武器(第 64 項(武器傷害真表),手冊 p.126 的球形清單 + p.127 的數值表)。
		// 先前 `WeaponKindSpherical` 這條解算分支**掛不到任何武器**——整段是死碼。
		{"脈衝星", 200, 24, gamedata.TOPIC_WARP_FIELDS, gamedata.TECH_PULSAR},
		{"空間壓縮器", 260, 32, gamedata.TOPIC_XENON_TECHNOLOGY, gamedata.TECH_SPATIAL_COMPRESSOR},
		// 陀螺去穩器(第 70 項(陀螺去穩器)):第 64 項(武器傷害真表)判定「資料齊但光束路徑沒有 per size class 乘數」
		// ——正解不是替光束加乘數,是認出它其實屬球形家族(依級數乘 + 豁免盾甲兩個特徵都有)。
		{"陀螺去穩器", 180, 4, gamedata.TOPIC_GRAVITIC_FIELDS, gamedata.TECH_GYRO_DESTABILIZER},
	}
)

// BuildWeaponOptions 依版本規則 profile 回傳一份武器元件清單:除「電漿砲」的 Value(最大傷害)
// 改讀 p.PlasmaCannonMaxDamage 外,其餘元件與套件級 WeaponOptions 逐一相同。
//
// 2026-07-11 接線:GameSession.BuildShipWithMods(session.go)已改用本函式(以 s.RuleProfile 為
// 參數)取代直接 pick(WeaponOptions, weapon) 算武器攻擊值,故 BuildShip/BuildShipWithMods 造出的
// 艦艇 WeaponAttack 現在真的隨版本 profile 變動(1.3 電漿砲=30、1.5=20)。
//
// ⚠ 仍未接線的呼叫端:ShipDesignSpaceUsedWithMods/DesignCostWithMods(佔格/造價,套件級純函式,
// 無 GameSession 可查)——但兩版電漿砲的 Cost 與佔格本身相同(見元件清單/component-values.md),
// 只有 Value 隨版本差異,故這兩者維持讀套件級 WeaponOptions 不影響正確性,非遺漏。
// battleVolley 等既有戰鬥計算讀的是已建成 Ship.WeaponAttack(本次接線後已含版本值),非重新
// pick(WeaponOptions, ...),故戰鬥傷害自然隨造艦時的版本值走,不需另外接線。
func BuildWeaponOptions(p gamedata.RuleProfile) []Component {
	out := make([]Component, len(WeaponOptions))
	copy(out, WeaponOptions)
	for i, c := range out {
		if c.Name == "電漿砲" {
			out[i].Value = p.PlasmaCannonMaxDamage
		}
	}
	return out
}

var (
	// ArmorOptions 的 Value(裝甲 HP)= 基準單位 × `gamedata.ArmorStructurePercent`。
	//
	// ⚠ 2026-08-08(第 67 項(裝甲科技倍率))由自編值 10/20/35/55/80/120 改成手冊階梯
	// 100/200/400/600/800/1000%(見 gamedata/armor_tech.go 的逐句出處與那則撤回)。
	// **階梯是一手的,基準單位 10 是 remake 值**——原版沒有獨立的「裝甲池」,
	// 裝甲科技決定的是艦艇結構點數,兩池是 remake 的抽象。
	//
	// 差別最大的三格:佐特 35→40、中子素 55→60、氙素 120→100(它是「10 倍」不是「+1100%」)。
	ArmorOptions = []Component{
		{"無裝甲", 0, 0, 0, 0},
		{"鈦裝甲", 30, 10, gamedata.TOPIC_CHEMISTRY, gamedata.TECH_TITANIUM_ARMOR}, // ResearchAll(早期)
		{"三鈦裝甲", 60, 20, gamedata.TOPIC_ADVANCED_METALLURGY, gamedata.TECH_TRITANIUM_ARMOR},
		{"佐特裝甲", 100, 40, gamedata.TOPIC_NANO_TECHNOLOGY, gamedata.TECH_ZORTRIUM_ARMOR},
		{"中子素裝甲", 160, 60, gamedata.TOPIC_MOLECULAR_MANIPULATION, gamedata.TECH_NEUTRONIUM_ARMOR},
		{"精金裝甲", 240, 80, gamedata.TOPIC_MOLECULAR_CONTROL, gamedata.TECH_ADAMANTIUM_ARMOR},
		// ⚠ 主題與解鎖科技一併訂正:先前掛在 TOPIC_ARTIFICIAL_LIFE 且 UnlockTech=0(標「里程碑,proxy」)。
		// 執行檔的 tech→topic 表說它屬 **Xenon Technology**(第 74 號),而 TECH_XENTRONIUM_ARMOR
		// 一直都在列舉裡(201)——「proxy」是當時沒查,不是查了查不到。
		{"氙素裝甲", 350, 100, gamedata.TOPIC_XENON_TECHNOLOGY, gamedata.TECH_XENTRONIUM_ARMOR},
	}
	ShieldOptions = []Component{
		{"無護盾", 0, 0, 0, 0},
		{"第一級護盾", 40, 15, gamedata.TOPIC_ADVANCED_MAGNETISM, gamedata.TECH_CLASS_I_SHIELD},
		{"第三級護盾", 90, 35, gamedata.TOPIC_MAGNETO_GRAVITICS, gamedata.TECH_CLASS_III_SHIELD},
		{"第五級護盾", 150, 60, gamedata.TOPIC_SUBSPACE_FIELDS, gamedata.TECH_CLASS_V_SHIELD},
		{"第七級護盾", 230, 90, gamedata.TOPIC_QUANTUM_FIELDS, gamedata.TECH_CLASS_VII_SHIELD},
		{"第十級護盾", 350, 140, gamedata.TOPIC_TEMPORAL_FIELDS, gamedata.TECH_CLASS_X_SHIELD},
	}
	SpecialOptions = []Component{
		{"無", 0, 0, 0, 0},
		{"戰鬥電腦", 80, 3, gamedata.TOPIC_ARTIFICIAL_INTELLIGENCE, 0}, // 抽象(電腦研究鏈),proxy 待重設計
		{"自動修復", 60, 0, gamedata.TOPIC_ADVANCED_MANUFACTURING, gamedata.TECH_AUTOMATED_REPAIR_UNIT},
		{"隱形裝置", 100, 0, gamedata.TOPIC_DISTORTION_FIELDS, gamedata.TECH_CLOAKING_DEVICE},
		{"重生程序", 150, 0, gamedata.TOPIC_ARTIFICIAL_LIFE, 0},                                         // 抽象(種族特性),proxy 待重設計
		{"戰機庫", 90, 0, gamedata.TOPIC_ADVANCED_ENGINEERING, gamedata.TECH_FIGHTER_BAYS},             // 攔截機隊出擊4(手冊 GM p.127),ResolveBattle 加母艦戰力
		{"重戰機庫", 160, 0, gamedata.TOPIC_SUPERSCALAR_CONSTRUCTION, gamedata.TECH_HEAVY_FIGHTER_BAYS}, // 重戰機隊出擊2、火力較強(手冊 GM p.127)
		// 轟炸機庫(第 70 項(陀螺去穩器)):`gamedata` 的戰機速度與射擊次數**兩組都是四型都齊的**,
		// 而 shell 只實作了兩型——連 FighterKind 的檔頭都寫著「手冊 p.127 的四種」。
		// 主題取自執行檔(進階機器人學)。
		{"轟炸機庫", 130, 0, gamedata.TOPIC_ADVANCED_ROBOTICS, gamedata.TECH_BOMBER_BAYS},
		// 硬化護盾:與隱形裝置同屬 TOPIC_DISTORTION_FIELDS(techtree.go 的三選一)。
		// 手冊兩處效果:每次攻擊額外減傷 3(gamedata.DamageHardShieldBonus,先前無元件載體
		// 等於死碼)、以及**星雲中護盾仍然可用**(見 nebula.go)。
		// ⚠ 成本 100 沿用同主題的隱形裝置,是 remake 值不是原版真值——本表其餘成本同樣是
		// remake 值(見表頭那幾筆的「proxy 待重設計」註記)。
		{"硬化護盾", 100, gamedata.DamageHardShieldBonus, gamedata.TOPIC_DISTORTION_FIELDS, gamedata.TECH_HARD_SHIELDS},
		// 反飛彈火箭(第 64 項(武器傷害真表)):`ResolveMissileShot` 的 hasAMR 參數從寫出來就恆傳 false,
		// 理由是「現行 remake 的 SpecialOptions 尚未提供這個可造艦元件」——那句話對,
		// 而這一行就是把它補上。研究主題取自執行檔(`OrigTechTopic` → 進階工程),
		// 執行檔的 category 也是 28(反飛彈/干擾),與手冊 p.127 的分類一致。
		// Value 留 0:AMR 不加攻防,它的效果是攔截(見 battleVolley 的 hasAMR 分支)。
		{"反飛彈火箭", 70, 0, gamedata.TOPIC_ADVANCED_ENGINEERING, gamedata.TECH_ANTIMISSILE_ROCKETS},
		// 高能聚焦(第 66 項(高能聚焦)):`gamedata.DamageMountBonusHEF = 50` 與 `DamageMountAdjustedValue`
		// 的 hefBonus 參數從寫出來就沒有生產端——因為 **HEF 在手冊裡是「系統」不是武器改造**
		// (`High Energy Focus (System)`),而系統要進這張表才裝得上,於是兩個呼叫端一路傳 0。
		//
		// 手冊逐字:「increasing the damage each of these weapons inflicts by 50%.
		// It does not improve the chances of hitting a target at a greater distance,
		// nor does it prevent the normal drop-off of damage over range.」
		// ——**只加傷害,不動命中、也不抵銷距離衰減**,三句各自對應程式裡的一個位置。
		//
		// 主題取自 techtree.go 的 TOPIC_HIGH_ENERGY_DISTRIBUTION 三選一
		// (能量吸收器 / 高能聚焦 / 超級導管),那張表本身是從執行檔抽的。
		// ⚠ 成本 90 是 **remake 值**:手冊行文沒給系統的建造成本,執行檔的元件表還沒挖到
		// (RACESTUF 那條路只有種族資料)。取值參考同屬中後期系統的硬化護盾(100)與
		// 反飛彈火箭(70)。與本表其餘元件同一種標記方式。
		{"高能聚焦", 90, 0, gamedata.TOPIC_HIGH_ENERGY_DISTRIBUTION, gamedata.TECH_HIGH_ENERGY_FOCUS},
		// 重裝甲(第 67 項(裝甲科技倍率)):與高能聚焦同一個形狀——手冊寫的是「Heavy Armor **(System)**」,
		// 系統要進這張表才裝得上,而它先前不在,於是兩個效果都沒有落點:
		//
		//	Installing Heavy Armor **triples** the amount of damage the ship's armor can
		//	sustain before damage gets through to the structure and internal systems.
		//	This system also **negates the Armor Piercing abilities** of enemy weapons.
		//
		// 第二句正是 `DamageApplyArmor` 那個從寫出來就恆傳 false 的 `apNegated` 參數。
		// 主題取自執行檔的 tech→topic 表(進階建設,與自動化工廠/行星飛彈基地三選一)。
		// ⚠ 成本 110 是 remake 值,理由同高能聚焦。
		{"重裝甲", 110, 0, gamedata.TOPIC_ADVANCED_CONSTRUCTION, gamedata.TECH_HEAVY_ARMOR},
		// --- 飛彈防禦系統家族(第 68 項(元件盤點+飛彈防禦))---
		//
		// `gamedata/missile.go` 把手冊 p.123「Special Defensive Systems」與「Missile Evasion」
		// 整段搬進來了:干擾器三型、慣性穩定/抵消、閃電場、位移裝置,**每一個都有精確數字**。
		// 而 `ResolveMissileShot` 的 `defenderEvasionBonus` 一直只吃得到艦員經驗與舵手技能
		// ——那七個常數一個生產端都沒有,因為**這些系統從來不在元件清單上**。
		//
		// 這是第 66/67 項同一個形狀的第三次:手冊的 System 沒進 SpecialOptions。
		// 這次不是一個一個撞到的——把手冊所有 `(System)`/`(Ship)` 條目對四張元件表做了
		// 完整性盤點,一次撈出來(盤點結果見 docs/re/01-gap-report.md 第 68 項(元件盤點+飛彈防禦))。
		//
		// 主題全部取自執行檔的 tech→topic 表(`OrigTechTopic`),不是猜的。
		// ⚠ 成本一律是 remake 值(理由同高能聚焦/重裝甲),依主題先後遞增排。
		// ⚠ **慣性抵消器與位移裝置同屬 Transwarp Fields(71),是三選一裡的兩個選項**
		// ——研究時只能挑一個,與微晶構築/奈米分解者同款(第 59 項(成就科技效果))。
		{"電子干擾器", 80, 0, gamedata.TOPIC_ADVANCED_MAGNETISM, gamedata.TECH_ECM_JAMMER},
		{"慣性穩定器", 100, 0, gamedata.TOPIC_GRAVITIC_FIELDS, gamedata.TECH_INERTIAL_STABILIZER},
		{"廣域干擾器", 130, 0, gamedata.TOPIC_QUANTUM_FIELDS, gamedata.TECH_WIDE_AREA_JAMMER},
		{"多波電子干擾器", 140, 0, gamedata.TOPIC_SUBSPACE_FIELDS, gamedata.TECH_MULTIWAVE_ECM_JAMMER},
		{"慣性抵消器", 150, 0, gamedata.TOPIC_TRANSWARP_FIELDS, gamedata.TECH_INERTIAL_NULLIFIER},
		{"位移裝置", 150, 0, gamedata.TOPIC_TRANSWARP_FIELDS, gamedata.TECH_DISPLACEMENT_DEVICE},
		{"閃電場", 160, 0, gamedata.TOPIC_WARP_FIELDS, gamedata.TECH_LIGHTNING_FIELD},
		// 部隊艙:`gamedata.GroundTroopPodsMultiplier = 2`(手冊 p.79「doubling the number of
		// Marines on board a ship」)同樣零生產端,同樣是因為元件不存在。
		{"部隊艙", 70, 0, gamedata.TOPIC_CAPSULE_CONSTRUCTION, gamedata.TECH_TROOP_PODS},
		// --- 第 68 項:第 68 項(元件盤點+飛彈防禦)盤點出來那 20 個裡,數字最硬、且 remake 已有承接位置的五個 ---
		// 主題同樣全部取自執行檔的 tech→topic 表;成本一律 remake 值。
		// 沒接的十幾個與各自的理由寫在 ship_systems.go 檔尾。
		{"偵察實驗室", 60, 0, gamedata.TOPIC_ARTIFICIAL_INTELLIGENCE, gamedata.TECH_SCOUT_LAB},
		{"強化船體", 90, 0, gamedata.TOPIC_ADVANCED_ENGINEERING, gamedata.TECH_REINFORCED_HULL},
		{"多相護盾", 170, 0, gamedata.TOPIC_MULTIPHASED_PHYSICS, gamedata.TECH_MULTIPHASED_SHIELDS},
		{"戰鬥掃描器", 120, 0, gamedata.TOPIC_TACHYON_PHYSICS, gamedata.TECH_BATTLE_SCANNER},
		// 第 68 項:傷害鏈收成具名結構之後,這兩個就接得進去了
		// (第 68 項(元件盤點+飛彈防禦)把它們列在「要先重構」那一格)。
		{"結構分析儀", 140, 0, gamedata.TOPIC_CYBERTRONICS, gamedata.TECH_STRUCTURAL_ANALYZER},
		{"阿基里斯瞄準器", 200, 0, gamedata.TOPIC_MOLECULATRONICS, gamedata.TECH_ACHILLES_TARGETING_UNIT},
		// 增強引擎(第 69 項(戰鬥速度與引擎階)):手冊「increase the combat speed of a ship by +5」。
		// 第 68 項(元件盤點+飛彈防禦)把它列在「機制不存在」那一格——戰鬥速度模型建起來之後就成立了。
		{"增強引擎", 100, 0, gamedata.TOPIC_ADVANCED_FUSION, gamedata.TECH_AUGMENTED_ENGINES},
		// 狀態類武器(第 69 項(戰鬥速度與引擎階)):第 64 項(武器傷害真表)判定它們「卡在機制」——牽引光束要戰鬥速度模型
		// (第 69 項(戰鬥速度與引擎階))、停滯力場要「這一輪不能動」的狀態。兩塊都建好了,現在接得上。
		{"牽引光束", 130, 0, gamedata.TOPIC_ARTIFICIAL_GRAVITY, gamedata.TECH_TRACTOR_BEAM},
		{"停滯力場", 190, 0, gamedata.TOPIC_DISTORTION_FIELDS, gamedata.TECH_STASIS_FIELD},
		// 行動次數家族(第 70 項(陀螺去穩器)):三個元件卡在同一個缺失機制「一回合開幾次火」,
		// 建一次就一次解三個。冷卻的讀法見 gamedata/shots_per_round.go
		// ——手冊的 unused 是「完全沒開火」,不是「沒有連射」。
		// 戰鬥艙:手冊「add equipment space without increasing the hull size」。
		// 舊擋門理由是「remake 沒有逐元件佔格的造艦模型」——那句**當初就寫錯**,
		// gamedata.ShipHullSpace + shell.ShipDesignSpaceUsed 一直都在。真正缺的是那個
		// 「加多少空間」的數字,而手冊從頭到尾沒給。
		//
		// 現在有了:原版特殊裝置表的空間欄是**負數**,而且剛好是艦體空間的一半
		// (docs/tech/special-device-table.md)。成本同樣走那張表(依艦級 20..2500),
		// 所以這一列的 Cost 欄留 0——它不會被 DesignCostWithMods 讀到。
		//
		// ⚠ remake 只有一個特殊系統槽,所以裝了戰鬥艙就裝不了別的。原版是多槽,
		// 戰鬥艙的用法是「騰出空間給其他系統」;在 remake 它換來的是**更大的武器**。
		// 這是單槽模型的必然後果,不是抄錯。
		{"戰鬥艙", 0, 0, gamedata.TOPIC_CAPSULE_CONSTRUCTION, gamedata.TECH_BATTLE_PODS},
		// --- 登艦戰家族(第 80 項(登艦戰),見 boarding.go)---
		// 手冊把登艦戰的解算方式直接指回地面戰,而那套解算器早就在了——缺的從來不是公式。
		// 傳送器需要格子戰術的護盾分面狀態；命中鏈在 interactive.go 接上該面減傷與容量。
		{"突擊艇", 0, 0, gamedata.TOPIC_ADVANCED_ENGINEERING, gamedata.TECH_ASSAULT_SHUTTLES},
		{"傳送器", 0, 0, gamedata.TOPIC_MATTER_ENERGY_CONVERSION, gamedata.TECH_TRANSPORTERS},
		{"保安站", 0, 0, gamedata.TOPIC_POSITRONICS, gamedata.TECH_SECURITY_STATIONS},
		// --- 匿蹤家族 + 測距瞄準器 + 能量吸收器(見 cloak.go / energy_absorber.go)---
		// 這四項先前都在「擋門理由已經過期或本來就錯」那一格,共通點是**手冊給了確切數字**,
		// 而缺的機制其實都已經建好了(Fired、不可被選為目標、真實格距離、跨回合狀態)。
		// 成本走原版特殊裝置表(依艦級),所以 Cost 欄留 0。
		//
		// ⚠ 隱形裝置本來就在這張表的最上面一段——那一列從加進來那天起就**沒有任何程式碼讀它**
		// (第 72 項(元件表有≠效果有接))。這一輪補的是效果,不是元件。
		{"能量吸收器", 0, 0, gamedata.TOPIC_HIGH_ENERGY_DISTRIBUTION, gamedata.TECH_ENERGY_ABSORBER},
		{"測距瞄準器", 0, 0, gamedata.TOPIC_ARTIFICIAL_CONSCIOUSNESS, gamedata.TECH_RANGEMASTER_UNIT},
		{"相位匿蹤", 0, 0, gamedata.TOPIC_TEMPORAL_FIELDS, gamedata.TECH_PHASING_CLOAK},
		{"快速飛彈架", 120, 0, gamedata.TOPIC_SERVO_MECHANICS, gamedata.TECH_FAST_MISSILE_RACKS},
		{"超載電容", 160, 0, gamedata.TOPIC_HYPER_DIMENSIONAL_FISSION, gamedata.TECH_HYPERX_CAPACITORS},
		{"時間扭曲加速器", 250, 0, gamedata.TOPIC_TEMPORAL_PHYSICS, gamedata.TECH_TIME_WARP_FACILITATOR},
	}
)

// armorHPByName 依裝甲元件名回傳其 HP 值(戰鬥用);查無回 0。
func armorHPByName(name string) int {
	for _, c := range ArmorOptions {
		if c.Name == name {
			return c.Value
		}
	}
	return 0
}

// shieldReduceByName 依護盾元件名回傳「每發減傷」(戰鬥用):無=0、第一級=1、第三級=3、
// 第五級=5、第七級=7、第十級=10(手冊值,見 gamedata.ShieldReductionForTech)。
//
// ⚠ 這段註解原本寫著「remake 由護盾階推導:…第一級=2、第三級=4…**精確 per-class 真值待逆向**」。
// 那句話有兩個問題,第 62 項(護盾減傷)一起修掉:推導出來的數字五級裡有四級是錯的,而**真值根本
// 不需要逆向**——`gamedata.DamageShieldReductionClass*` 從手冊抄進來之後一直躺在那裡沒人用。
func shieldReduceByName(name string) int {
	for _, c := range ShieldOptions {
		if c.Name == name {
			// ⚠ 2026-08-08(第 62 項(護盾減傷)):這裡原本回 `i * 2`(清單索引 × 2),
			// 五級裡有四級與手冊不符而且一律偏高(1/3/5/7/10 被算成 2/4/6/8/10)。
			// 手冊值改由 gamedata.ShieldReductionForTech 依**科技**查——名稱會被翻譯、
			// 清單順序一改索引就跑掉,那正是原本那個算法出問題的地方。
			return gamedata.ShieldReductionForTech(c.UnlockTech)
		}
	}
	return 0
}

// ComponentUnlocked 回傳某元件是否已解鎖。
//
// 解鎖規則(MOO2 每主題數科技間抉擇的非破壞式實作):
//   - 起始科技(Tech=0)一律解鎖。
//   - 主題未完成 → 未解鎖。
//   - 主題已完成、但元件未映射真科技(UnlockTech=TECH_NONE,如里程碑/抽象元件)→ 主題層級解鎖。
//   - 主題已完成、元件有映射科技、但玩家「未明確抉擇」該主題(AI/預設)→ 主題層級解鎖(不回歸)。
//   - 主題已完成、有映射科技、玩家「已明確抉擇」該主題 → 僅所選科技對應元件解鎖(忠實抉擇)。
func (s *GameSession) ComponentUnlocked(c Component) bool {
	// 規則本體抽成 componentUnlockedFor(ground_invasion.go),供玩家與 AI 共用同一套判定
	// (地面戰 force 加成需要對 AIOpponent.Player 套用相同規則)。
	return componentUnlockedFor(s.Player, c)
}

// NextUnlockedComponent 從 opts[cur] 起找下一個已解鎖元件的索引(循環;至少回 0=無)。
func (s *GameSession) NextUnlockedComponent(opts []Component, cur int) int {
	for step := 1; step <= len(opts); step++ {
		i := (cur + step) % len(opts)
		if s.ComponentUnlocked(opts[i]) {
			return i
		}
	}
	return 0
}

// homeworldShips 是「Average 起始文明等級」的忠實開局艦隊:1 艘殖民船 + 2 艘偵察艦。
// 依據 docs/tech/homeworld-init.md §4:
//   - 手冊 p.13 定性保證「small star fleet, including one Colony Ship」(高信心)。
//   - 「2 艘起始偵察艦」取自 patch 1.50 changelog「the two starting scouts will have 12
//     combat speed instead of 10」的間接證據(中信心;changelog 只改速度數值、隱含經典版
//     數量本就是 2,非正式列表)。
//   - 除此 3 艘外是否還有 Outpost Ship/護衛艦等其他艦,手冊未列完整清單(§4.3 待確認),
//     故 remake 目前只給這 3 艘,不臆測補齊。
//
// 三艘均為空武裝(殖民船/偵察艦在原版本就不具備武器容量,非本 remake 遺漏)。
//
// ⚠ **艦名存英文原文,中文由注入的翻譯器在建艦當下翻**(同第 84 項(名稱池雙語化)的星名/艦名池)。
// 先前這三個名字是硬編中文,所以英文模式的艦隊畫面上永遠掛著「拓荒號 / 先驅一二號」。
// 三個名字本身是 remake 自訂的(原版從艦名池抽,而這裡沒有 rng 可用),譯表在
// `assets/i18n/shipname.json` 尾端獨立標記。
func homeworldShips(tr func(string) string) []Ship {
	local := func(en string) string {
		if tr == nil {
			return en
		}
		return tr(en)
	}
	return []Ship{
		{Name: local("Pathfinder"), Class: "殖民船", Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
		{Name: local("Vanguard I"), Class: "偵察艦", Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
		{Name: local("Vanguard II"), Class: "偵察艦", Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
	}
}

// isSupportShipClass 回傳該艦體等級是否為支援艦(手冊 p.119 明列的三種:運輸艦、殖民船、
// 前哨船;運輸艦在 remake 尚無獨立艦種,陸戰隊由 FleetMarines 抽象承載)。
// 支援艦不參戰(見 mkPlayerCombatants)。
func isSupportShipClass(class string) bool {
	switch class {
	case ColonyShipClass, OutpostShipClass:
		return true
	}
	return false
}

// shipStrength 依艦體等級給戰力點(供最小戰鬥解算;正式版由艦艇設計的武器/裝甲算)。
func shipStrength(class string) int {
	switch class {
	case "偵察艦":
		return 1
	case "殖民船":
		return 1 // 非戰鬥艦(殖民/擴張用途),暫沿用最低戰力占位;remake 尚無獨立非戰鬥艦模型
	case "巡防艦", "護衛艦":
		return 2
	case "驅逐艦":
		return 4
	case "巡洋艦":
		return 8
	case "戰艦":
		return 16
	case "泰坦":
		return 32
	case "末日之星":
		return 64
	}
	return 1
}

// BattleResult 是一場戰鬥的結果(逐回合解算)。
type BattleResult struct {
	Enemy                     string
	EnemyKind                 BattleEnemyKind
	PlayerStart, EnemyStart   int // 開戰時雙方艦數
	PlayerWon                 bool
	PlayerLosses, EnemyLosses int
	Log                       []BattleRoundResult // 逐回合 typed 戰報
	// CrewXPGained 是這一仗每艘倖存艦拿到的艦員經驗(手冊 p.121:被擊沉敵艦艦體等級
	// 總和的一半,最少 1)。輸掉的仗是 0——手冊寫的是「Each battle **won**」。
	CrewXPGained int
}

type BattleEnemyKind string

const (
	BattleEnemyDynamic BattleEnemyKind = ""
	BattleEnemyAntaran BattleEnemyKind = "antaran_home_defense"
)

type BattleRoundResult struct {
	Round           int
	EnemyDestroyed  int
	PlayerDestroyed int
}

// removeWeakestShip 移除戰力最弱的一艘艦。
func (s *GameSession) removeWeakestShip() {
	f := s.Fleet() // 戰損打的是**參戰的那一支**艦隊
	if len(f.Ships) == 0 {
		return
	}
	wi := 0
	for i, sh := range f.Ships {
		if shipStrength(sh.Class) < shipStrength(f.Ships[wi].Class) {
			wi = i
		}
	}
	f.Ships = append(f.Ships[:wi], f.Ships[wi+1:]...)
}

// Difficulties 是 NEW GAME 的五個離散難度選項。
//
// **五級**是反組譯確認的:新遊戲畫面的難度選擇器建立時傳選項數 5
// (`sub_CCE2E` 裡 `push 5` + 變數 `word_1A1362`),而 `NEWGAME.LBX` 的難度圖正好也是 5 張
// (資產 4–8,五張手勢:伸手扶持 → 握手 → 比讚 → 握拳 → 雙拳相抵,難度遞增)。
// remake 先前只有四級,少的是最低的「教學」。
//
// 原版 consumer 各自直接讀 `difficulty 0..4`；這裡不放通用倍率。已證實效果
// 由 pickAIPersonality、GroundDifficultyBonus 等 typed helper 分別消費。
var Difficulties = []struct{ Name string }{
	{"教學"}, {"簡單"}, {"普通"}, {"困難"}, {"不可能"},
}

// GalaxyAges 是星系年齡選項,對應 NEW GAME 的 GALAXY AGE(反組譯:選項數 3,變數
// `word_1A1358`;`NEWGAME.LBX` 資產 1–3 是三張地景圖:熔岩 → 沙漠 → 綠地湖泊)。
//
// 效果在 `gamedata`(光譜分布 `StarClassWeights` 與氣候骰表 `OldGalaxyClimateWeights`)
// 早就實作了,先前只是被一個常數寫死成 Average,UI 完全選不到。
var GalaxyAges = []struct {
	Name string
	Age  gamedata.GalaxyAge
}{
	{"年輕", gamedata.GalaxyYouthful},
	{"普通", gamedata.GalaxyAverage},
	{"成熟", gamedata.GalaxyMature},
}

// TechLevels 是起始科技等級,對應 NEW GAME 的 TECH LEVEL(原版手冊稱 Starting Civilization)。
//
// **三級**同樣是反組譯確認的(選項數 3、變數 `word_1A1360`,`NEWGAME.LBX` 資產 20–22 是三張
// 城市圖,由樸素到未來)。patch 1.5 手冊寫的是「Pre-warp / Avg / Post-warp / Advanced」四級
// ——**那是 1.5 才加的**,1.5 的 CHANGELOG 明寫「Added pictures for cluster, random and
// post warp in newgame.lbx」。1.3 的 LBX 裡就是沒有第四張圖,兩邊沒有矛盾。
//
// gameplay 效果**已接一項**(2026-08-07):曲速前開局沒有 FTL,艦隊離不開本星系,
// 直到研究完 `FTLTopic`(見 `FleetHasFTL`)。手冊直引:「Exploring outside that system is
// impossible until faster than light (FTL) technologies are discovered.」
//
// **第二項效果 2026-08-07 接上**:開局送的研究主題依等級而不同。那張表先前拿不到
// (註解寫著「沒有一手表之前不臆造」),當天從 `Init_Player_Tech_` @ 0x5E55F 挖出來:
// 數量 `var_18` = 1 / 6 / 25,清單 `word_18111C` = 29, 55, 22, 57, 28, 23。
// 見 `gamedata.StartingTopicOrder` 與 `GameSession.applyStartingTech`。
// 手冊獨立說「預設的第一個是 field #29」,與表頭互證。
//
// 開局建築已依 3/5/9 上限與 `min(ceil(2/3 population), cap)` 產生；先進級固定六次後的
// 十九次選擇也已由 `sub_FD335` 的應用級單次抽選接線。兩者分別由
// `applyStartingBuildings` 與 `applyStartingRandomTech` 承接。
var TechLevels = []struct {
	Name string
	Desc string
}{
	{"曲速前", "只有母星,無 FTL(需研究核分裂才能離開本星系)"},
	{"一般", "母星 + 基本艦隊"},
	{"先進", "多項科技 + 跨星系艦隊"},
}

// OpponentCounts 是可選的帝國總數,對應 NEW GAME 的 PLAYERS。
//
// 反組譯:選項數 7、變數 `word_1A1366`,而且該變數是由 `byte_199CB1` **減 2** 得來
// (`movzx ax, byte_199CB1 / dec eax / dec eax`)——所以選項 0..6 對應帝國總數 **2..8**。
// `NEWGAME.LBX` 資產 13–19 正是七張數字圖「2 3 4 5 6 7 8」,兩邊互證。
//
// remake 的 `SetupNewGame(stars, seed, numAI)` 吃的是**對手數**,即帝國總數 − 1。
const (
	MinEmpires = 2
	MaxEmpires = 8
)

// genEnemyFleet 依回合數生成尚未有原版 runtime blueprint 的敵方艦隊代理清單。
// 難度不在這裡做通用倍率縮放：原版各子系統分別讀離散難度索引，
// 沒有證據支持把同一張浮點表直接乘進敵艦每項屬性。
func genEnemyFleet(turn int) []int {
	n := 2 + turn/3
	sizes := []int{2, 4, 8, 16}
	f := make([]int, 0, n)
	for i := 0; i < n; i++ {
		f = append(f, sizes[i%len(sizes)])
	}
	return f
}

// ResolveBattle 逐回合解算與某敵方的一場戰鬥:雙方艦隊每回合交火、逐艦擊沉,直到一方全滅
// 或滿 6 回合;套用玩家損失到艦隊。
// combatant 是快速艦隊結算用的單艦戰鬥屬性(與 StartCombat 同款由艦艇設計推導)。
// kind 依武器名分類戰鬥解算路徑(見 weapon_kind.go);敵方艦隊(genEnemyFleet)沒有個別
// 武器設計資料,一律留零值 WeaponKindBeam(既有簡化,非本輪引入)。
type combatant struct {
	hp, atk, def, wmin, wmax, shield, armor int
	// maxHP 是未受損時的血量,供戰鬥中的自動修復算「已受損多少」(見 repair.go)。
	// 0 = 未設(敵方艦隊不需要),此時視為等於 hp。
	maxHP int
	// autoRepair 標記這艘船裝有自動修復元件(手冊 p.82:每回合修復 20% 結構損傷)。
	autoRepair bool
	// shipIdx 是這個 combatant 對應 s.Ships 的索引(-1 = 敵方/無對應)。
	// battleVolley 過濾陣亡者時是整個 struct 複製,所以這個欄位會跟著倖存者走——
	// 這正是「戰後把剩餘血量寫回正確那艘船」需要的東西(先前用外部平行陣列會在有人陣亡後錯位)。
	shipIdx int
	// battleID 只在快速結算的一場戰鬥內識別行動者；合併主動權序列不能用切片索引，
	// 因為每次擊沉都會壓縮敵方切片。它不進存檔，也不改變原始艦艇索引。
	battleID int
	kind     WeaponKind
	// weaponName 是攻方武器的元件名——球形武器要靠它判「是不是空間壓縮器」
	// (只有那一項手冊明講豁免護盾與裝甲)。
	weaponName string
	// sizeClass 是這艘船的艦體等級。球形武器的傷害「per size class of target」要用它
	// ——見 battleVolley 的球形分支。
	sizeClass gamedata.CombatShipClass
	// hasAMR 是這艘船裝了反飛彈火箭(手冊 p.127:攔截來襲飛彈)。
	// 與 missileEvasion 一樣只在**它是防守方**時有意義。
	hasAMR bool
	// hasHEF 是這艘船裝了高能聚焦(手冊 p.87:光束傷害 +50%,不影響命中與距離衰減)。
	hasHEF bool
	// 攻方光束系統(第 68 項(元件盤點+飛彈防禦)):高能聚焦 / 結構分析儀 / 阿基里斯瞄準器。
	beamSystems BeamAttackerSystems
	// initiative 是手冊的主動權(Beam Attack + 10×戰鬥速度),決定齊射次序。
	initiative int
	// shots 是這一發齊射裡這艘船開幾次火(第 70 項(陀螺去穩器))。快速結算沒有跨回合狀態,
	// 所以充能一律視為滿——**這是刻意的簡化**:快速結算本來就沒有「上一回合」的概念。
	shots int
	// ammoSet 區分「舊測試／舊呼叫端未提供彈藥」與「本場已確實耗盡為 0」。
	// 不能只把 0 正規化成預設值，否則快速結算每輪都會把空彈架補滿。
	ammo    int
	ammoSet bool
	// apNegated 是這艘**被打的**船讓敵方穿甲失效(氙素裝甲 或 重裝甲系統,手冊各一句)。
	apNegated bool
	// 飛彈特殊防禦(手冊 p.123):裝了才擲骰,見 ResolveMissileShot 的 MissileDefenses。
	hasLightningField bool
	hasDisplacement   bool
	// hardShield 是這艘**被打的**船裝了硬化護盾。手冊那句的主詞是「each enemy attack」
	// ——不分武器類型,所以三條射擊路徑(光束/飛彈/球形)都要吃到。
	// 第 71 項(探針③內部函式)之前只有光束路徑接了,飛彈與球形一律傳 false。
	hardShield bool
	// scannerJamReduction 是這艘**開火的**船抵銷目標飛彈閃避的點數(迅子 20 / 中子 40)。
	// 帝國級(掃描科技無逐艦元件),敵方艦隊無科技資料故為 0。
	scannerJamReduction int
	// missileEvasion 是這艘船的飛彈閃避加成(手冊 ME 欄 + 舵手技能)。
	// 只有當**它是防守方**時才有意義;敵方艦隊無逐艦資料,一律 0(既有簡化)。
	missileEvasion int
	// missileFTLLevel 是這艘玩家艦在快速戰鬥中發射飛彈所用的引擎階，供 FST
	// 的 Beam Defense 消費端使用。敵方抽象艦隊沒有逐艦飛彈設計，留 0。
	missileFTLLevel int
	// pointDefenseSpent 是本次快速齊射中該艦的 PD 自動攔截是否已使用；目前每艘
	// 舊單槽艦由此欄相容；多槽艦另以 pointDefenseSpentSlots 逐槽追蹤。
	pointDefenseSpent bool
	// pointDefenseSpentSlots 與 weaponMounts 等長，只保存本輪快速齊射的自動開火狀態。
	pointDefenseSpentSlots []bool
	// pointDefenseInterceptionDamage 是原版 Weapon_In_Range 的攔截傷害餘數。
	// 它跨同一場快速戰鬥保留，不能在每次 PD 射擊後丟掉；ARM 只改變每枚
	// 飛彈所需的 durability，FST 則改變飛彈 Beam Defense，兩者都共用這條鏈。
	pointDefenseInterceptionDamage int
	// mods 是攻方武器改造(gamedata.WeaponModCode 字串);battleVolley 會依 weaponName
	// 過濾並套用 beam/missile/torpedo 改造,敵方艦隊(genEnemyFleet)沒有個別武器
	// 設計,一律 nil(既有簡化,非本輪引入)。
	mods []string
	// weaponMounts 是同一艘艦的逐槽武器資料。HP／護盾／特殊狀態仍只存在 combatant 一份；
	// battleVolley 輪到這艘艦時逐槽開火，不能把槽展開成多艘可被攻擊的假船。
	weaponMounts []ShipWeaponMount
	// --- 匿蹤與能量吸收器(見 cloak.go / energy_absorber.go)---
	//
	// ⚠ 快速結算沒有「回合」的概念(見 shots 欄的說明),所以匿蹤在這裡只有**一次性**的
	// 語意:開場隱形,一旦開火就永久失效——「停火一整回合重新隱形」需要跨回合狀態,
	// 那是格子戰術才有的東西。相位匿蹤的 10 回合降級同理,在這裡一律當成還沒降級。
	// 這是既有簡化的延伸,不是漏抄。
	cloakKind CloakKind
	cloaked   bool
	// energyAbsorber / storedEnergy:被打時轉存 1/4 潛在傷害,下一次開火自動命中射出。
	energyAbsorber bool
	storedEnergy   int
	// --- 登艦戰(見 boarding.go)---
	marines          int
	securityStations bool
	marineStrength   int
	marineHitsToKill int
	boardingBonus    int // 該艦艦員等級的 Bo
	commandoBonus    int // 同 owner 參戰艦軍官 Commando 最大值
	securityBonus    int // 同 owner 參戰艦軍官 Security 最大值；只由守方消費
	assaultShuttles  bool
	// boarded 是「這艘船這場戰鬥已經派過登艦隊了」。手冊:突擊艇是**一次性**的
	// (「unpiloted shuttles are set adrift to be picked up after the battle」),
	// 快速結算沒有回合,所以一場一次。
	boarded bool
}

// battleVolley 讓每個存活 attacker 對第一個存活 defender 射一發(固定近距 range=2),
// 依 attacker 的武器類型分流真戰鬥公式:beam 沿用 ResolveShot(不動,回歸測試見
// combat_weapon_kind_test.go);missile 改用 ResolveMissileShot(躲避/AMR 攔截);
// spherical 改用 ResolveSphericalShot(第 64 項(武器傷害真表)起真的有武器掛載:脈衝星/空間壓縮器)。
// 回傳本輪擊沉的 defender 數。移除陣亡艦。
func battleVolley(attackers []combatant, defenders *[]combatant, rng *rand.Rand) int {
	before := len(*defenders)
	for i := range *defenders {
		(*defenders)[i].pointDefenseSpent = false
		for j := range (*defenders)[i].pointDefenseSpentSlots {
			(*defenders)[i].pointDefenseSpentSlots[j] = false
		}
	}
	// 主動權排序(第 69 項(戰鬥速度與引擎階)):手冊「smaller ships should move before bigger, slower ones」。
	// 先前是艦隊清單順序,等於「先造的先打」——與速度完全無關。
	// 就地排序沒有問題:呼叫端每次都重新建構戰列(mkPlayerCombatantsIndexed / genEnemyFleet)。
	sortByInitiative(attackers)
	for i := range attackers {
		battleCombatantVolley(&attackers[i], defenders, rng)
	}
	alive := (*defenders)[:0]
	for _, c := range *defenders {
		if c.hp > 0 {
			alive = append(alive, c)
		}
	}
	*defenders = alive
	return before - len(*defenders)
}

// battleCombatantVolley 執行單一艦艇的一次回合行動。拆出此層是為了讓 Ship Initiative
// 開啟時能把敵我艦艇放進同一條行動序列，而不複製武器、彈藥與特殊系統公式。
func battleCombatantVolley(attacker *combatant, defenders *[]combatant, rng *rand.Rand) {
	if len(attacker.weaponMounts) > 0 {
		battleMountVolley(attacker, defenders, rng)
		return
	}
	nshots := attacker.shots
	if nshots < 1 {
		nshots = 1
	}
	for sh := 0; sh < nshots; sh++ {
		prepareCombatantAmmo(attacker)
		if attacker.kind == WeaponKindMissile && !weaponAmmoCanFire(attacker.ammo) {
			break
		}
		battleShot(attacker, defenders, rng)
	}
}

func compactLivingCombatants(items []combatant) []combatant {
	alive := items[:0]
	for _, item := range items {
		if item.hp > 0 {
			alive = append(alive, item)
		}
	}
	return alive
}

func resetQuickPointDefense(items []combatant) {
	for i := range items {
		items[i].pointDefenseSpent = false
		for j := range items[i].pointDefenseSpentSlots {
			items[i].pointDefenseSpentSlots[j] = false
		}
	}
}

type quickInitiativeAction struct {
	player     bool
	battleID   int
	initiative int
}

// resolveQuickCombatRound 只負責一個完整戰鬥回合的先後順序。關閉時沿用原版的
// 雙方分批行動；開啟時將兩側放入同一個穩定降冪序列，已先被擊沉的艦不會還擊。
func resolveQuickCombatRound(player, enemy *[]combatant, initiative bool, rng *rand.Rand) (enemyDestroyed, playerDestroyed int) {
	pBefore, eBefore := len(*player), len(*enemy)
	if !initiative {
		battleVolleyInFleetOrder(*player, enemy, rng)
		battleVolleyInFleetOrder(*enemy, player, rng)
		for i := range *player {
			(*player)[i].storedEnergy = 0
		}
		for i := range *enemy {
			(*enemy)[i].storedEnergy = 0
		}
		return eBefore - len(*enemy), pBefore - len(*player)
	}

	resetQuickPointDefense(*player)
	resetQuickPointDefense(*enemy)
	actions := make([]quickInitiativeAction, 0, len(*player)+len(*enemy))
	for i := range *player {
		actions = append(actions, quickInitiativeAction{player: true, battleID: (*player)[i].battleID, initiative: (*player)[i].initiative})
	}
	for i := range *enemy {
		actions = append(actions, quickInitiativeAction{battleID: (*enemy)[i].battleID, initiative: (*enemy)[i].initiative})
	}
	sort.SliceStable(actions, func(i, j int) bool { return actions[i].initiative > actions[j].initiative })
	for _, action := range actions {
		actors, targets := enemy, player
		if action.player {
			actors, targets = player, enemy
		}
		for i := range *actors {
			if (*actors)[i].battleID == action.battleID && (*actors)[i].hp > 0 {
				battleCombatantVolley(&(*actors)[i], targets, rng)
				*targets = compactLivingCombatants(*targets)
				break
			}
		}
	}
	return eBefore - len(*enemy), pBefore - len(*player)
}

func battleVolleyInFleetOrder(attackers []combatant, defenders *[]combatant, rng *rand.Rand) int {
	before := len(*defenders)
	resetQuickPointDefense(*defenders)
	for i := range attackers {
		battleCombatantVolley(&attackers[i], defenders, rng)
	}
	*defenders = compactLivingCombatants(*defenders)
	return before - len(*defenders)
}

// battleMountVolley 依原版 record 槽序消費同一艘艦的多個武器槽。它只暫時替換武器欄位；
// 儲能、匿蹤、登艦與其他艦級狀態留在原 combatant，故不會因槽數重置。
func battleMountVolley(atk *combatant, defenders *[]combatant, rng *rand.Rand) {
	origKind, origName := atk.kind, atk.weaponName
	origMin, origMax := atk.wmin, atk.wmax
	origAmmo, origAmmoSet := atk.ammo, atk.ammoSet
	origMods := atk.mods
	defer func() {
		atk.kind, atk.weaponName = origKind, origName
		atk.wmin, atk.wmax = origMin, origMax
		atk.ammo, atk.ammoSet, atk.mods = origAmmo, origAmmoSet, origMods
	}()

	shipShots := atk.shots
	if shipShots < 1 {
		shipShots = 1
	}
	for mi := range atk.weaponMounts {
		mount := &atk.weaponMounts[mi]
		if mount.Name == "" || mount.WorkingCount <= 0 {
			continue
		}
		atk.kind = weaponKindByName(mount.Name)
		atk.weaponName = mount.Name
		if mi == 0 {
			// 第一槽是舊相容欄位的來源；沿用既有 wmin/wmax，確保所有單槽存檔與
			// 戰鬥測試逐項不變。後續槽才使用各自保存的 Attack。
			atk.wmin, atk.wmax = origMin, origMax
		} else {
			atk.wmax = mount.Attack
			if atk.wmax < 1 {
				atk.wmax = origMax
			}
			atk.wmin = atk.wmax / 2
		}
		atk.ammo = NormalizeWeaponAmmo(mount.Name, mount.Ammo)
		atk.ammoSet = true
		// 已解碼改造逐槽消費；只有舊第一槽缺 typed Mods 時回退相容欄位。
		atk.mods = append([]string(nil), mount.Mods...)
		if mi == 0 && len(atk.mods) == 0 {
			atk.mods = origMods
		}
		for count := 0; count < mount.WorkingCount; count++ {
			for shot := 0; shot < shipShots; shot++ {
				if atk.kind == WeaponKindMissile && !weaponAmmoCanFire(atk.ammo) {
					break
				}
				battleShot(atk, defenders, rng)
			}
		}
		mount.Ammo = atk.ammo
	}
}

// battleShot 是 battleVolley 裡「一艘船打一發」的那一段(第 70 項(陀螺去穩器)從迴圈裡抽出來,
// 讓連射系統可以重複呼叫)。抽出來的過程沒有改任何判定,只是把 `attackers[i]` 換成參數。
func battleShot(atk *combatant, defenders *[]combatant, rng *rand.Rand) {
	ti := -1
	for j := range *defenders {
		// 相位匿蹤:手冊「While cloaked, the ship **cannot be attacked**」——不是難打中,
		// 是根本選不到。所以它排在存活判定旁邊,不是後面的命中判定裡。
		// (快速結算沒有回合,一律當成還沒過 10 回合的降級門檻,見 combatant.cloakKind。)
		if (*defenders)[j].hp > 0 && !((*defenders)[j].cloaked && (*defenders)[j].cloakKind == CloakPhasing) {
			ti = j
			break
		}
	}
	if ti < 0 {
		return // 敵方全滅(或全部躲在相位匿蹤裡),這一發沒有目標
	}
	d := &(*defenders)[ti]
	// 突擊艇(第 80 項(登艦戰)):手冊「Marines on Assault Shuttles **always try to capture**
	// the target ship」——所以快速結算這一側固定是奪船,沒有突襲那個選項。
	// 一場戰鬥一次(手冊:突擊艇是一次性的,放完人就漂在那裡)。
	// 沒裝就完全不動 RNG,既有戰鬥逐位元不變。
	if atk.assaultShuttles && !atk.boarded && d.marines > 0 {
		atk.boarded = true
		quickBoardingAttempt(atk, d, rng)
		if d.hp <= 0 {
			return
		}
	}
	// 儲能先射(手冊:自動命中,而且**射儲能不會解除隱形**,所以在 atk.cloaked 之前)。
	// 沒有儲能就完全不動 RNG,既有戰鬥逐位元不變。
	if atk.storedEnergy > 0 {
		releaseStoredEnergyQuick(atk, d, rng)
		if d.hp <= 0 {
			return
		}
	}
	var shot ShotResult
	switch atk.kind {
	case WeaponKindBomb:
		// 手冊 p.126:「Bombs installed in a ship are only useful against planetary
		// targets」——艦隊戰裡完全不開火。**不是 0 傷害的命中,是根本沒有這一發**,
		// 所以連骰子都不擲(擲了會位移後面每一發的隨機序列,讓決定性測試無故變動)。
		return
	case WeaponKindMissile:
		prepareCombatantAmmo(atk)
		if !weaponAmmoCanFire(atk.ammo) {
			return
		}
		atk.ammo = weaponAmmoSpend(atk.ammo)
		missileMods := WeaponModCodesForWeapon(atk.weaponName, atk.mods)
		warheads := gamedata.WeaponModMissileWarheadCount(missileMods)
		var mdef MissileDefenses
		// 多槽艦依原始槽序逐門自動迎擊；舊單槽艦仍只走一發相容路徑，
		// 因而沒有 PD 時及舊存檔的 RNG 消費不變。
		if len(d.weaponMounts) > 0 && len(d.pointDefenseSpentSlots) < len(d.weaponMounts) {
			d.pointDefenseSpentSlots = append(d.pointDefenseSpentSlots,
				make([]bool, len(d.weaponMounts)-len(d.pointDefenseSpentSlots))...)
		}
		pdMounts := pointDefenseMountsFor(d.weaponName, d.mods, d.wmax,
			d.weaponMounts, d.pointDefenseSpentSlots, d.pointDefenseSpent)
		for _, mount := range pdMounts {
			d.pointDefenseSpent = true
			if mount.Slot >= 0 {
				d.pointDefenseSpentSlots[mount.Slot] = true
			}
			for n := 0; n < mount.Count; n++ {
				if !PointDefenseCanEngage(mount.WeaponName, atk.weaponName, mount.BeamMods) {
					continue
				}
				pd := ResolvePointDefenseIntercept(PointDefenseShot{
					BeamWeaponName:            mount.WeaponName,
					BeamAttack:                d.atk,
					BeamDamageMax:             mount.BeamDamageMax,
					BeamRangeSquares:          0, // 同格自動攔截(手冊 p.117)
					BeamRoll:                  rng.Intn(100) + 1,
					BeamSystems:               d.beamSystems,
					BeamMods:                  mount.BeamMods,
					MissileWeaponName:         atk.weaponName,
					MissileFTLLevel:           atk.missileFTLLevel,
					MissileMods:               missileMods,
					CarriedInterceptionDamage: d.pointDefenseInterceptionDamage,
				})
				if pd.Fired {
					mdef.InterceptedWarheads += pd.DestroyedWarheads
					d.pointDefenseInterceptionDamage = pd.RemainingInterceptionDamage
				}
			}
		}
		amrRoll := rng.Intn(100) + 1
		jamRoll := rng.Intn(100) + 1
		// 特殊防禦裝置(第 68 項(元件盤點+飛彈防禦)):**裝了才擲骰**,沒裝就完全不動 RNG
		// ——這樣既有存檔/探針的戰鬥結果逐位元不變(同炸彈分支的處理)。
		if d.hasLightningField {
			mdef.HasLightningField, mdef.LightningRoll = true, rng.Intn(100)+1
		}
		if d.hasDisplacement {
			mdef.HasDisplacement, mdef.DisplacementRoll = true, rng.Intn(100)+1
		}
		// 匿蹤:手冊「missiles and torpedoes have a 50% chance to miss」。同款「裝了才擲骰」。
		if c := quickCloakMissChance(d); c > 0 {
			mdef.CloakMissChance, mdef.CloakRoll = c, rng.Intn(100)+1
		}
		if warheads > 1 {
			mdef.JamRolls = []int{jamRoll}
			if mdef.CloakMissChance > 0 {
				mdef.CloakRolls = []int{mdef.CloakRoll}
			}
			if mdef.HasDisplacement {
				mdef.DisplacementRolls = []int{mdef.DisplacementRoll}
			}
			for i := 1; i < warheads; i++ {
				mdef.JamRolls = append(mdef.JamRolls, rng.Intn(100)+1)
				if mdef.CloakMissChance > 0 {
					mdef.CloakRolls = append(mdef.CloakRolls, rng.Intn(100)+1)
				}
				if mdef.HasDisplacement {
					mdef.DisplacementRolls = append(mdef.DisplacementRolls, rng.Intn(100)+1)
				}
			}
		}
		// ⚠ 2026-08-08:上一版註解寫「那句話對 ECM 干擾器/慣性穩定器仍成立」
		// ——第 68 項(元件盤點+飛彈防禦)把那一整族補進 SpecialOptions 了,`missileEvasion` 現在吃得到。
		// ⚠ 2026-08-08(第 71 項(探針③內部函式)):倒數第二個引數先前恆為 false(硬化護盾),
		// 第五個引數恆為 0(攻方掃描器)。兩者現在都有真值來源——手冊各自寫得很清楚,
		// 只是這兩個參數位置從加進來那天起就沒有人回頭填。
		shot = ResolveMissileShotWithMods(d.hasAMR, 2, amrRoll, d.missileEvasion,
			atk.scannerJamReduction, false, jamRoll,
			atk.wmax, d.shield, d.armor, d.hardShield, mdef, atk.weaponName, missileMods)
	case WeaponKindSpherical:
		span := atk.wmax - atk.wmin
		r := 0
		if span > 0 {
			r = rng.Intn(span + 1)
		}
		aggD := gamedata.DamageSphericalRoll(atk.wmin, r, 100)
		// 手冊 p.127 把脈衝星與空間壓縮器的傷害寫成「**per size class of target**」
		// ——同一發打大船比打小船痛。
		//
		// ⚠ **「級數」取 index+1 是讀法,不是手冊字面。** 手冊沒列出級數的數字,只給了
		// 艦體名稱的順序;取 index 會讓護衛艦那一級乘 0(打護衛艦零傷害),那顯然不是
		// 規則,所以是 index+1(護衛艦 1 … 末日之星 6)。標在這裡,不假裝它是抄來的。
		aggD *= int(d.sizeClass) + 1
		// 空間壓縮器手冊明講「does all damage to structure only, ignoring shields
		// and armor」;脈衝星沒有那一句,所以只有前者豁免。
		shot = ResolveSphericalShot(aggD, d.shield, d.armor, d.hardShield,
			weaponBypassesShieldAndArmor(atk.weaponName))
	default:
		roll := rng.Intn(100) + 1
		// 隱形裝置:+80 光束防禦(手冊那句的主詞是 defense,所以加在守方而非扣攻方命中)。
		net := atk.atk - (d.def + quickCloakBeamDefense(d))
		shot = ResolveBeamShot(BeamShot{
			NetAttack: net, WeaponMin: atk.wmin, WeaponMax: atk.wmax,
			RangeSquares: 2, Roll: roll, Mods: WeaponModCodesForWeapon(atk.weaponName, atk.mods),
			Attacker: atk.beamSystems,
			Target: BeamTargetSystems{
				ShieldReduction: d.shield, ArmorHP: d.armor, APNegated: d.apNegated,
				// ⚠ 2026-08-08:HardShield 先前**沒有填**。第 71 項(探針③內部函式)補了飛彈與球形
				// 兩條路徑,而 combatant.hardShield 的註解卻寫著「三條路徑都要吃到,
				// 之前只有光束路徑接了」——事實相反:光束是唯一沒接的那一條。
				// 「Resolve* 有這個參數」不等於「呼叫端有填」,而那個結構欄位有預設零值,
				// 所以漏填不會編譯失敗、也不會有任何測試紅——直到有人逐欄看過去。
				HardShield: d.hardShield,
			},
		})
	}
	if shot.Hit {
		// 能量吸收器:轉存 1/4「抵達這艘船」的傷害。取扣盾前的值(見
		// gamedata.EnergyAbsorberStored 對 reaches / penetrates 兩個詞的說明);
		// 這裡拿得到的是扣完盾甲的結構傷害,所以用射前的潛在值反推不了——
		// **改用武器上限**,那正是「潛在傷害」在 remake 這條路徑上最接近的量。
		if d.energyAbsorber {
			d.storedEnergy += gamedata.EnergyAbsorberStored(atk.wmax)
		}
		d.armor = shot.RemainingArmorHP
		d.hp -= shot.DamageToStructure
	}
	// 開火即解除隱形(手冊:是開火當下,不是下一回合)。快速結算沒有回合,所以這是永久的
	// ——見 combatant.cloakKind 的說明。
	atk.cloaked = false
}

// quickBoardingAttempt 是快速結算這一側的登艦:一隊突擊艇(4 架 × 1 個陸戰隊單位)
// 飛過去奪船。
//
// ⚠ **奪到船在 remake 只表現成「那艘船退出戰鬥」**,不是真的換手。手冊說奪船之後還要
// 贏下整場戰鬥才留得住(心靈感應種族除外),而快速結算連「戰鬥結束時誰還在」都只用
// 存活數判定——真的把船搬到對面陣營要動戰後結算與艦隊清單兩處。這是**建模取捨**,
// 標在這裡:奪船的即時效果是正確的(那艘船不再開火),長期歸屬沒做。
func quickBoardingAttempt(atk *combatant, d *combatant, rng *rand.Rand) {
	party := BoardingParty{
		Intent:     BoardingCapture,
		Marines:    gamedata.FighterSquadronSize * gamedata.AssaultShuttleMarinesEach,
		Strength:   atk.marineStrength + atk.boardingBonus + atk.commandoBonus,
		HitsToKill: atk.marineHitsToKill,
	}
	def := BoardingDefense{
		Marines: d.marines, Strength: d.marineStrength + d.boardingBonus + d.commandoBonus,
		HitsToKill:       d.marineHitsToKill,
		StrengthBonus:    d.securityBonus,
		SecurityStations: d.securityStations,
	}
	res := ResolveBoarding(party, def, func(n int) int {
		if n <= 0 {
			return 0
		}
		return rng.Intn(n)
	})
	d.marines = res.DefenderSurvived
	if res.Captured {
		d.hp = 0 // 守軍全滅 → 這艘船這場戰鬥打完了
	}
}

// quickCloakBeamDefense / quickCloakMissChance 是快速結算這一側的匿蹤查詢。
//
// 與格子戰術共用同一組手冊常數,但**不共用函式**:那邊的 CloakBeamDefenseBonus 吃的是
// CombatShip 與回合數,而快速結算既沒有 CombatShip 也沒有回合。硬要共用得先造一個
// 中介型別,那比兩個三行函式貴。
func quickCloakBeamDefense(c *combatant) int {
	if !c.cloaked || c.cloakKind == CloakNone {
		return 0
	}
	return gamedata.ShipCloakingDeviceBeamDefense
}

func quickCloakMissChance(c *combatant) int {
	if !c.cloaked || c.cloakKind == CloakNone {
		return 0
	}
	return gamedata.ShipCloakingDeviceMissileMissChance
}

// releaseStoredEnergyQuick 是能量吸收器在快速結算這一側的發射:自動命中,除非目標有位移裝置;
// 傷害照光束的距離衰減表打折(快速結算固定 range=2,對應 level 1 → 衰減 0)。
func releaseStoredEnergyQuick(atk *combatant, d *combatant, rng *rand.Rand) {
	stored := atk.storedEnergy
	atk.storedEnergy = 0
	if d.hasDisplacement && rng.Intn(100)+1 <= gamedata.MissileDisplacementDeviceMissChance {
		return
	}
	level := gamedata.CombatRangeLevel(2)
	dmg := stored * (100 - gamedata.DamageDissipationPenalty(level)) / 100
	dmg = gamedata.DamageAfterShield(dmg, d.shield, d.hardShield, false)
	_, toStruct, remArmor := gamedata.DamageApplyArmor(dmg, d.armor, false, d.apNegated)
	d.armor = remArmor
	d.hp -= toStruct
}

// mkPlayerCombatants 把玩家目前艦隊(s.Ships)轉成 []combatant,供快速艦隊戰鬥解算共用——
// ResolveBattle 的一般艦隊交戰與 orbital_bombardment.go BombardColony 的軌道基地反擊都用
// 這一套換算規則(2026-07-11 從 ResolveBattle 內的匿名函式抽出,行為不變,純供多處重用,
// 避免兩處各自維護、日後改壞其中一邊卻沒同步)。
func (s *GameSession) mkPlayerCombatants() []combatant {
	out, _ := s.mkPlayerCombatantsIndexed()
	return out
}

// mkPlayerCombatantsIndexed 同 mkPlayerCombatants,另外回傳「第 k 個參戰艦對應 s.Ships 的
// 哪個索引」。戰後要把剩餘血量寫回持久損傷(見 repair.go applyBattleDamage)時需要這個對映
// ——支援艦不參戰(手冊 p.119),所以兩邊的索引不是一一對應。
func (s *GameSession) mkPlayerCombatantsIndexed() ([]combatant, []int) {
	var out []combatant
	var idx []int
	commando := fleetOfficerSkillMax(s.Leaders, s.Fleet().Ships, gamedata.SKILL_COMMANDO)
	security := fleetOfficerSkillMax(s.Leaders, s.Fleet().Ships, gamedata.SKILL_SECURITY)
	marineStrength := s.playerMarineForce()
	marineHits := gamedata.GroundMarineHitsToKill(s.raceHasTrait(gamedata.TRAIT_HIGH_G), s.hasPoweredArmor())
	for shipIdx, sh := range s.Fleet().Ships {
		if isSupportShipClass(sh.Class) {
			// 手冊 p.119 逐字:「Support ships … **do not fight**, but are destroyed if an enemy
			// military fleet chooses to engage them in combat.」——支援艦(殖民船/前哨船/運輸艦)
			// 不參與火力計算。先前它們會以最低戰力(shipStrength 回 1)混進戰列,等於帶著殖民船
			// 去打仗還能加一點傷害,並且在計算損失時第一個被打掉。
			//
			// 「被擊毀」由兩個戰後入口明確處理：快速結算的
			// removeDestroyedCombatParticipants 與戰術結算的 ApplyCombatOutcome 都移除支援艦，
			// 不再靠 removeWeakestShip 的排序副作用猜測。
			continue
		}
		body := shipStrength(sh.Class)
		atk := body + sh.WeaponAttack + s.shipOfficerSkillBonus(sh, gamedata.SKILL_WEAPONRY)
		ordnanceBonus := s.shipOfficerSkillBonus(sh, gamedata.SKILL_ORDNANCE)
		atk += atk * s.RaceCombatPct / 100 // 種族戰鬥加成(姆瑞森艦攻+50、埃雷里安+20…)
		// 戰鬥掃描器(第 68 項(元件盤點+飛彈防禦)):手冊「increases the ship's chance to hit with beam weapons
		// by 50」——**點數加成**,所以加在種族百分比之後(理由同下面的艦員加成)。
		atk += shipBeamOffenseBonus(sh)
		// 艦員經驗(手冊 p.121 的 BA/BD 兩欄):老手打得準也閃得掉。
		// 加在種族加成**之後**——那兩張表是直接的點數加成,不是百分比,所以不該被種族倍率放大。
		crew := s.shipCrewLevel(sh)
		atk += gamedata.ShipCrewOffenseBonus(crew)
		// ⚠ 2026-08-08(第 60 項(打得準也閃得掉)):上面那句註解說「打得準**也閃得掉**」,
		// 但先前只加了 BA(攻擊)那一欄,BD(防禦)那一欄從來沒加過——`def` 只有艦體值。
		// 手冊 p.121 的兩欄是分開的兩個加成,`engine.BeamDefense` 也是這樣算的
		// (openorion2 `Ship::beamDefense` 末項),只是 shell 這條路徑沒有走它。
		crewDef := gamedata.ShipCrewDefenseBonus(crew)
		// 種族艦艇**防禦**加成(阿爾卡里 +50、埃雷里安 +25)。與攻擊側對稱:
		// 套在艦體值上、在艦員點數加成之前——理由同上,那兩欄是點數不是百分比。
		defBody := body + body*s.RaceShipDefPct/100
		// 慣性穩定器/抵消器(第 68 項(元件盤點+飛彈防禦),補第 68 項(元件盤點+飛彈防禦)漏掉的那一半):手冊那一條同時給
		// 「+50 beam defense」與「+25 missile evasion」,先前只接了後者。
		defBody += shipBeamDefenseBonus(sh) + s.shipOfficerSkillBonus(sh, gamedata.SKILL_HELMSMAN)
		// 強化船體:手冊「triples the amount of structural damage a ship can sustain」。
		hp := body * 3 * shipStructureMultiplier(sh) / 100
		// 戰機庫:出擊一隊戰機(手冊:中隊一律 4 架;返航前射擊次數攔截機 4、重戰機 2),
		// 在艦級抽象結算中以母艦戰力加成承接整隊火力與血量。
		// ⚠ 2026-08-07 訂正:這裡原本寫「出擊數:攔截機 4 / 重戰機 2」——那一欄是 **Shots**
		// (每架返航前開幾次火),不是中隊人數。見 gamedata/combat.go 那一段的完整對照。
		if shipHasSpecial(sh, "戰機庫") {
			fatk, fhp := gamedata.FighterBayCombatContribution()
			atk += fatk
			hp += fhp
		}
		if shipHasSpecial(sh, "重戰機庫") {
			fatk, fhp := gamedata.FighterHeavyBayCombatContribution()
			atk += fatk
			hp += fhp
		}
		if shipHasSpecial(sh, "轟炸機庫") {
			fatk, fhp := gamedata.FighterBomberBayCombatContribution()
			atk += fatk
			hp += fhp
		}
		// 持久損傷(見 repair.go):受損的船帶著傷上陣,不再每場戰鬥都滿血。
		if d := ShipDamage(sh); d > 0 {
			hp -= d
			if hp < ShipDamageFloorHP {
				hp = ShipDamageFloorHP
			}
		}
		out = append(out, combatant{hp: hp, maxHP: shipMaxHP(sh), atk: atk, def: defBody + crewDef,
			wmin: atk / 2, wmax: atk + ordnanceBonus,
			shield: s.nebulaShield(shieldReduceByName(sh.Shield), shipHasHardShield(sh)) *
				shipShieldMultiplier(sh) / 100, // 多相護盾:吸收量 +50%
			armor: effectiveArmorHP(sh),
			kind:  weaponKindByName(sh.Weapon), weaponName: sh.Weapon, mods: sh.Mods,
			weaponMounts: cloneWeaponMounts(sh.WeaponMounts),
			sizeClass:    shipSizeClass(sh.Class),
			hasAMR:       shipHasSpecial(sh, antiMissileRocketName),
			hasHEF:       shipHasSpecial(sh, highEnergyFocusName),
			beamSystems:  shipBeamAttackerSystems(sh),
			initiative:   gamedata.CombatInitiative(atk, s.shipCombatSpeed(sh)),
			shots: gamedata.ShotsThisRound(shipShotsKind(sh),
				weaponKindByName(sh.Weapon) == WeaponKindBeam,
				weaponKindByName(sh.Weapon) == WeaponKindMissile, true),
			ammo:                NormalizeWeaponAmmo(sh.Weapon, sh.WeaponAmmo),
			ammoSet:             true,
			apNegated:           shipNegatesArmorPiercing(sh),
			hasLightningField:   shipHasLightningField(sh),
			hasDisplacement:     shipHasDisplacementDevice(sh),
			hardShield:          shipHasHardShield(sh),
			scannerJamReduction: bestPlayerScannerJamReduction(s.Player),
			missileEvasion: gamedata.ShipCrewMissileEvasionBonus(crew) + s.shipOfficerMissileEvasionBonus(sh) +
				shipMissileEvasionBonus(sh),
			missileFTLLevel: s.driveLevel(),
			marines:         ShipMarineComplement(sh), securityStations: shipHasSecurityStations(sh),
			marineStrength: marineStrength, marineHitsToKill: marineHits,
			boardingBonus: gamedata.ShipCrewBoardingBonus(crew), commandoBonus: commando,
			securityBonus:   security,
			assaultShuttles: shipHasAssaultShuttles(sh),
			cloakKind:       shipCloakKind(sh), cloaked: shipCloakKind(sh) != CloakNone,
			energyAbsorber: shipHasSpecial(sh, energyAbsorberName),
			autoRepair:     shipHasAutoRepair(sh), shipIdx: shipIdx})
		idx = append(idx, shipIdx)
	}
	return out, idx
}

// ResolveBattle 快速艦隊自動結算(無格子;供非互動戰鬥)。改用 gamedata 真戰鬥公式逐發解算,
// 與格子戰術戰鬥(tacticalScreen)一致:每回合雙方齊射,每發走命中判定→傷害→過盾→過甲。
func (s *GameSession) ResolveBattle(enemy string) BattleResult {
	var ef []combatant
	var enemyStartStrengths []int
	aiIdx := s.aiIndexByName(enemy)
	if aiIdx >= 0 {
		ef, enemyStartStrengths = s.aiCombatants(aiIdx)
	} else {
		for _, st := range genEnemyFleet(s.Turn) {
			ef = append(ef, combatant{hp: st * 3, atk: st, def: st, wmin: st / 2, wmax: st, armor: st, shipIdx: -1})
			enemyStartStrengths = append(enemyStartStrengths, st)
		}
	}
	pf, pfIdx := s.mkPlayerCombatantsIndexed()
	for i := range pf {
		pf[i].battleID = i + 1
	}
	for i := range ef {
		ef[i].battleID = len(pf) + i + 1
	}

	res := BattleResult{Enemy: enemy, PlayerStart: len(pf), EnemyStart: len(ef)}
	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + 12345)) // 依回合種子,可重現
	for round := 1; round <= 6 && len(pf) > 0 && len(ef) > 0; round++ {
		eDestroyed, pDestroyed := resolveQuickCombatRound(&pf, &ef, s.EffectiveGameSettings().ShipInitiative, rng)
		res.Log = append(res.Log, BattleRoundResult{Round: round, EnemyDestroyed: eDestroyed, PlayerDestroyed: pDestroyed})
	}
	res.PlayerLosses = res.PlayerStart - len(pf)
	res.EnemyLosses = res.EnemyStart - len(ef)
	res.PlayerWon = len(ef) == 0 || len(pf) >= len(ef)
	// 艦員經驗(手冊 p.121):打贏才給,而且是**被擊沉**敵艦艦體等級總和的一半。
	// 倖存的敵艦 atk 就是它的戰力值(genEnemyFleet 給的,戰鬥中不變),
	// 用「開打前的多重集合 − 結束時的多重集合」還原出被擊沉的是哪些。
	destroyedHullClassSum := hullClassSum(destroyedEnemySizeClasses(enemyStartStrengths, ef))
	// 倖存艦的剩餘血量寫回持久損傷(見 repair.go)。battleVolley 會就地移除陣亡艦,
	// 故 pf 剩下的是「倖存者、且保持原始相對順序」——與 pfIdx 的前綴對齊。
	s.applySurvivorDamage(pf, pfIdx)
	survivingParticipants := make(map[int]bool, len(pf))
	for _, survivor := range pf {
		if survivor.shipIdx >= 0 {
			survivingParticipants[survivor.shipIdx] = true
		}
	}
	if res.PlayerWon {
		res.CrewXPGained = s.awardBattleCrewXP(destroyedHullClassSum, survivingParticipants)
	}
	s.removeDestroyedCombatParticipants(pfIdx, survivingParticipants)
	if aiIdx >= 0 {
		s.applyAICombatantSurvivors(aiIdx, ef)
	}
	s.repairAfterBattle(res.PlayerWon) // 自動修復/進階損害管制/工程師(手冊 p.80/p.82/p.136)
	s.LastBattle = &res
	return res
}

// removeDestroyedCombatParticipants 依 battle combatant 的持久索引移除實際陣亡艦；
// 支援艦不在 combatant 陣列，但手冊 p.119 明載遭敵軍交戰時會被摧毀，因此一併移除。
// 這取代「按損失數反覆刪最弱艦」造成死艦留下、別艦代死的錯位。
func (s *GameSession) removeDestroyedCombatParticipants(participants []int, survivors map[int]bool) {
	participated := make(map[int]bool, len(participants))
	for _, idx := range participants {
		participated[idx] = true
	}
	f := s.Fleet()
	kept := f.Ships[:0]
	for idx, sh := range f.Ships {
		if isSupportShipClass(sh.Class) {
			continue
		}
		if participated[idx] && !survivors[idx] {
			continue
		}
		kept = append(kept, sh)
	}
	f.Ships = kept
}

// applySurvivorDamage 把戰鬥結束時倖存艦的剩餘血量寫回 Ship.Damage。
//
// 對映靠 combatant.shipIdx(過濾陣亡者時整個 struct 複製,索引跟著倖存者走),
// 不是靠外部平行陣列——後者在有人陣亡後就會錯位,把 A 船的傷記到 B 船上。
//
// idx 參數保留給呼叫端表達「這一批 combatant 原本對應哪些船」,目前只用於長度檢查。
func (s *GameSession) applySurvivorDamage(survivors []combatant, _ []int) {
	for _, c := range survivors {
		// maxHP <= 0 同時擋掉兩種不該寫回的來源:敵方艦隊,以及 antaran_victory.go /
		// orbital_bombardment.go 那些沒填 maxHP 的臨時 combatant(它們的 shipIdx 是 Go 零值 0,
		// 若不擋就會把傷害記到玩家的第一艘船上)。
		if c.shipIdx < 0 || c.shipIdx >= len(s.Fleet().Ships) || c.maxHP <= 0 {
			continue
		}
		s.applyBattleDamage([]int{c.shipIdx}, []int{c.hp})
	}
}

// DiplomacyResponse 依雙方相對實力回應一個外交提議(和平/貿易/威脅)。
// raceDiploBonusPct 回傳目前種族的外交加成百分比。魅力非凡 +50%(手冊 p.15「Humans
// gain a 50% bonus to their diplomatic efforts」)。
//
// 內建種族由 RaceIndex 查原版特性；自訂種族由 CustomRaceTraits 查點數畫面
// 寫入的選項。兩條路徑共用 RaceCharismatic/RaceTelepathic 消費端。
func (s *GameSession) raceDiploBonusPct() int {
	bonus := 0
	if s.RaceCharismatic() {
		bonus += 50
	}
	if s.RaceTelepathic() {
		bonus += 25
	}
	return bonus
}

// aiByDisplayName 依畫面顯示名找 AI 對手(模糊比對含 "AI (…)" 外殼);找不到退回主要對手
// (AIPlayers[0])。回傳可修改指標;無 AI 時回 nil。
func (s *GameSession) aiByDisplayName(enemy string) *AIOpponent {
	for i := range s.AIPlayers {
		n := s.AIPlayers[i].Name
		if n == enemy || strings.Contains(n, enemy) || strings.Contains(enemy, n) {
			return &s.AIPlayers[i]
		}
	}
	if len(s.AIPlayers) > 0 {
		return &s.AIPlayers[0]
	}
	return nil
}

// adjustRelation 調整某 AI 對玩家的關係分數,夾在 -40..40(同 advanceAI 慣例)。
// clampRelation 把關係分數夾在 -40..40(與 adjustRelation 同一個尺度)。
func clampRelation(v int) int {
	if v > 40 {
		return 40
	}
	if v < -40 {
		return -40
	}
	return v
}

func (a *AIOpponent) adjustRelation(delta int) {
	a.Relation += delta
	if a.Relation > 40 {
		a.Relation = 40
	}
	if a.Relation < -40 {
		a.Relation = -40
	}
}

// DiplomacyResponse 處理玩家對某對手的外交動作，並把正式條約／貿易／研究
// 協議寫入該 AI 的 Treaty 狀態。關係變化仍受種族外交加成放大(人類
// Charismatic +50%,見 raceDiploBonusPct)，但條約收益由回合結算推進。
func (s *GameSession) DiplomacyResponse(action, enemy string) DiplomacyResult {
	s.recordPlayerCommand(PlayerCommand{Name: CmdDiplomacy, Text: action + "\x00" + enemy})
	ai := s.aiByDisplayName(enemy)
	if ai == nil {
		return DiplomacyResult{}
	}
	pPop := populationOfColonies(s.PlayerColonies)
	ePop := populationOfColonies(ai.Colonies)
	minPop := pPop
	if ePop < minPop {
		minPop = ePop
	}
	pFleet := 0
	for _, sh := range s.AllShips() { // 外交看的是**全帝國**軍力,不是眼前這一支
		pFleet += shipStrength(sh.Class)
	}
	switch action {
	case "peace":
		if !ai.Treaty.startFormal(gamedata.DIPLO_PEACE) {
			return diplomacyResult(DiploResultFormalExists, enemy)
		}
		ai.adjustRelation(s.diplomacyRelationGain(15)) // 提議和平改善關係(Charismatic／Diplomat 放大)
		if pPop >= ePop {
			return diplomacyResult(DiploResultPeaceStrong, enemy)
		}
		return diplomacyResult(DiploResultPeaceWeak, enemy)
	case "trade":
		if !ai.Treaty.startTrade(minPop) {
			return diplomacyResult(DiploResultTradeExists, enemy)
		}
		ai.adjustRelation(s.diplomacyRelationGain(10))
		return diplomacyResult(DiploResultTradeStarted, enemy)
	case "research":
		if !ai.Treaty.startResearch(minPop) {
			return diplomacyResult(DiploResultResearchExists, enemy)
		}
		ai.adjustRelation(s.diplomacyRelationGain(10))
		return diplomacyResult(DiploResultResearchStarted, enemy)
	case "special_food":
		if !ai.Treaty.startSpecialTrade(SpecialTradeFoodForCredits, minPop) {
			return diplomacyResult(DiploResultSpecialExists, enemy)
		}
		ai.adjustRelation(s.diplomacyRelationGain(12))
		return diplomacyResult(DiploResultSpecialFoodStarted, enemy)
	case "special_research":
		if !ai.Treaty.startSpecialTrade(SpecialTradeResearchExchange, minPop) {
			return diplomacyResult(DiploResultSpecialExists, enemy)
		}
		ai.adjustRelation(s.diplomacyRelationGain(12))
		return diplomacyResult(DiploResultSpecialResStarted, enemy)
	case "nonaggression":
		if !ai.Treaty.startFormal(gamedata.DIPLO_NON_AGGRESSION) {
			return diplomacyResult(DiploResultFormalConflict, enemy)
		}
		ai.adjustRelation(s.diplomacyRelationGain(15))
		return diplomacyResult(DiploResultNAPStarted, enemy)
	case "alliance":
		if !ai.Treaty.startFormal(gamedata.DIPLO_ALLIANCE) {
			return diplomacyResult(DiploResultFormalConflict, enemy)
		}
		ai.adjustRelation(s.diplomacyRelationGain(30))
		return diplomacyResult(DiploResultAllianceStarted, enemy)
	case "tribute_5":
		if !ai.Treaty.startPlayerTribute(TributeFivePercent) {
			return diplomacyResult(DiploResultTributeExists, enemy)
		}
		return diplomacyResult(DiploResultTribute5Started, enemy)
	case "tribute_10":
		if !ai.Treaty.startPlayerTribute(TributeTenPercent) {
			return diplomacyResult(DiploResultTributeExists, enemy)
		}
		return diplomacyResult(DiploResultTribute10Started, enemy)
	case "gift_cash":
		return s.OfferCashGift(enemy, diplomacyCashGiftDefault)
	case "gift_tech":
		opts := spyStealOptions(ai.Player, s.Player)
		if len(opts) == 0 {
			return diplomacyResult(DiploResultNoGiftTech, enemy)
		}
		return s.OfferTechnologyGift(enemy, opts[0].Topic, opts[0].Tech)
	case "gift_star":
		for i, star := range s.PlayerColonyStars {
			if star > 0 && star < len(s.Stars) && s.Stars[star].Owner == 1 {
				return s.OfferStarGift(enemy, star)
			}
			_ = i
		}
		return diplomacyResult(DiploResultNoGiftStar, enemy)
	case "break_trade":
		if !ai.Treaty.endTrade() {
			return diplomacyResult(DiploResultNoTrade, enemy)
		}
		ai.adjustRelation(-30)
		return diplomacyResult(DiploResultTradeEnded, enemy)
	case "break_research":
		if !ai.Treaty.endResearch() {
			return diplomacyResult(DiploResultNoResearch, enemy)
		}
		ai.adjustRelation(-30)
		return diplomacyResult(DiploResultResearchEnded, enemy)
	case "break_formal":
		if !ai.Treaty.endFormal() {
			return diplomacyResult(DiploResultNoFormal, enemy)
		}
		ai.OriginalHumanBetrayalRaw = true
		penalty := -10
		if int(ai.Personality) == 4 {
			penalty = -20
		}
		ai.OriginalHumanTreatyGrievanceRaw = originalSignedByteAdd(ai.OriginalHumanTreatyGrievanceRaw, penalty)
		if ai.PopulationRaceSlotKnown {
			ai.OriginalHumanTreatyVictimRaw = ai.PopulationRaceSlot
			ai.OriginalHumanTreatyVictimKnown = true
		}
		ai.adjustRelation(-30)
		return diplomacyResult(DiploResultFormalEnded, enemy)
	case "break_tribute":
		if !ai.Treaty.endTribute() {
			return diplomacyResult(DiploResultNoTribute, enemy)
		}
		return diplomacyResult(DiploResultTributeEnded, enemy)
	case "break_special":
		if !ai.Treaty.endSpecialTrade() {
			return diplomacyResult(DiploResultNoSpecial, enemy)
		}
		ai.adjustRelation(-20)
		return diplomacyResult(DiploResultSpecialEnded, enemy)
	case "threat":
		ai.adjustRelation(-20)
		if pFleet >= 10 {
			return diplomacyResult(DiploResultThreatStrong, enemy)
		}
		return diplomacyResult(DiploResultThreatWeak, enemy)
	}
	return DiplomacyResult{}
}

// CombatShip 是格子戰術戰鬥中的一艘艦(有 HP + 格位 + 真戰鬥公式所需的攻防/傷害/盾甲)。
type CombatShip struct {
	// TacticalID 是單一格子戰術內的暫態識別碼。戰損會壓縮切片，合併主動權佇列
	// 必須以此重新定位艦艇；0 表示尚未由畫面指派，不進存檔或持久艦艇資料。
	TacticalID int
	Name       string
	HP, MaxHP  int // 艦體結構 HP
	Attack     int // Beam Attack(BA,命中判定用)
	Col, Row   int // 格位(8 欄 × 6 列)
	// Facing 是原版 combat record +0x23 的 16 向 heading。0=右、4=上、
	// 8=左、12=下；移動時由 tactical UI 依移動向量更新。
	Facing int
	// 以下供 ResolveShot 真戰鬥公式使用(remake 由艦艇設計推導,見 StartCombat 註記)。
	Defense         int        // 守方防禦(AF+BD),減 netAttack
	WeaponMin       int        // 單發最小傷害
	WeaponMax       int        // 單發最大傷害
	ShieldReduction int        // 護盾每發減傷
	HardShield      bool       // 硬化護盾(額外減傷 + 星雲中護盾仍可用,見 nebula.go)
	ArmorHP         int        // 裝甲 HP(結構外的緩衝,先耗盡才傷結構)
	Kind            WeaponKind // 武器戰鬥解算路徑(beam/missile/spherical,見 weapon_kind.go);
	// 敵方艦(genEnemyFleet)無個別武器設計資料,一律留零值 WeaponKindBeam(既有簡化)。
	WeaponName string             // 武器元件名；飛彈／魚雷改造需要用它區分適用性。
	Mods       []string           // 武器改造(gamedata.WeaponModCode 字串);依 WeaponName 過濾後生效。
	WeaponArc  gamedata.WeaponArc // 設計資料中的火線角；格子戰術依 Facing 套用方向遮罩。
	// WeaponMounts 保存同一艘戰術艦的逐槽武器；艦級 HP／護盾／狀態仍只有一份。
	// nil 代表舊存檔／敵方代理，繼續走上方單槽相容欄位。
	WeaponMounts []ShipWeaponMount
	// WeaponModes 是本場戰術戰鬥的逐槽可用／待命／關閉狀態，不寫回持久設計。
	WeaponModes []TacticalWeaponMode
	// SpriteIdx 是 CMBTSHP.LBX 資產索引。原版標準艦艇的精確公式是
	// 45*玩家顏色色塊 + raw picture(+0xC4)；未知 picture 才退回艦級尺寸代表值。
	// 見 docs/tech/cmbtshp-ship-sprites.md。
	SpriteIdx int
	// Bay / BayKind:這艘船帶不帶戰機庫,以及是哪一型(見 fighter.go)。
	// 格子戰術戰鬥用它決定「這艘船能不能派戰機出擊」;快速結算走的是另一條路
	// (母艦戰力加成,見本函式下方的 Special 分支)。
	Bay     bool
	BayKind FighterKind
	// Bays 保存多特殊裝置設計中的每一座不同機庫；nil 時回退 Bay/BayKind 舊存檔欄位。
	Bays []FighterKind
	// 以下三項是由這艘艦載出的戰機中隊所用的 Beam Defense 加成。
	// 種族／戰機飛行員的原版來源是帝國／參戰艦隊層級,因此由 StartCombat 先算好；
	// Helmsman 保留欄位供 1.50 公式接線,目前沒有把未證實的原版呼叫端硬填進來。
	FighterRacialDefenseBonus int
	FighterPilotBonus         int
	FighterHelmsmanBonus      int
	// HEF 是這艘船裝了高能聚焦(光束傷害 +50%,手冊 p.87)。與 Mods 一樣只對 beam 生效,
	// 但它是**系統**不是改造,所以另開一個欄位而不是塞進 Mods。
	//
	// 敵方艦(genEnemyFleet)沒有個別元件設計資料,一律 false——與 Mods/HardShield 同款簡化。
	HEF bool
	// APNegated 是這艘船讓敵方穿甲失效(氙素裝甲 或 重裝甲系統)。
	APNegated bool
	// 飛彈防禦(手冊 p.123)。先前格子戰術這一側一律傳 0/false,註解寫著
	// 「現行皆無對應可造艦元件」——第 64/68 項把那些元件都補上了。
	MissileEvasion    int  // 艦員經驗 + 舵手技能 + 干擾器/慣性穩定器
	HasAMR            bool // 反飛彈火箭:射程內攔截
	HasLightningField bool // 閃電場:每一枚來襲飛彈各 50% 直接摧毀
	HasDisplacement   bool // 位移裝置:一律 30% 完全未命中
	// PointDefenseSpent 是本回合已解碼 PD 自動攔截是否已使用；回合交界重置。
	// 尾槽 raw mods 尚未完整解碼，因此目前只會由可確認的 typed Mods 觸發。它不寫入存檔。
	PointDefenseSpent bool
	// PointDefenseSpentSlots 與 WeaponMounts 等長，保存本回合已自動開火的 typed PD 槽。
	// WeaponModes 只控制主動齊射；原版明定紅色關閉的 PD 仍會迎擊，因此兩者不得共用。
	PointDefenseSpentSlots []bool
	// PointDefenseInterceptionDamage 是 PD 攔截鏈跨回合保存的 quotient/remainder
	// 餘數。每回合只重置 PointDefenseSpent，不重置這個值。
	PointDefenseInterceptionDamage int
	// ScannerJamReduction 是這艘船**開火時**抵銷目標飛彈閃避的點數(迅子 20 / 中子 40,
	// 手冊那兩個條目的最後一句)。與 MissileEvasion 相對:一個是防守方的躲、一個是攻擊方的破。
	ScannerJamReduction int
	// BeamSystems 是這艘船的攻方光束系統(高能聚焦/結構分析儀/阿基里斯瞄準器)。
	BeamSystems BeamAttackerSystems
	// DriveLevel 是玩家目前的引擎階(1..6)。戰機速度與飛彈速度都吃它
	// ——先前 cmd/moo2 那一側硬編成 1(見 drive_level.go 檔頭)。
	DriveLevel int
	// ArmorLevelAboveTitanium 是裝甲比鈦裝甲高幾級(戰機 HP 用,手冊註腳
	// 「base hit points are modified by 2 times armor level above Titanium」)。
	ArmorLevelAboveTitanium int
	// CombatSpeed / Initiative 見 gamedata/combat_speed.go(執行檔一手表 + 手冊公式)。
	CombatSpeed int
	Initiative  int
	// SizeClass 是艦體級數(0=巡防艦 … 5=末日之星),牽引光束依它算需要幾束才定得住。
	SizeClass gamedata.CombatShipClass
	// --- 狀態類武器:投射能力(第 69 項(戰鬥速度與引擎階))---
	TractorBeams   int  // 這艘船投射幾束牽引光束(裝了 = 1)
	HasStasisField bool // 這艘船帶停滯力場產生器
	// --- 狀態類武器:承受中的效果(每回合重算,不存檔)---
	HeldByTractors int  // 身上有幾束敵方牽引光束
	InStasis       bool // 被停滯力場定住:不能動、不能開火、**也不能被打**
	// --- 行動次數(第 70 項(陀螺去穩器))---
	ShotsKind  gamedata.ShotsPerRoundKind // 超載電容 / 快速飛彈架 / 時間扭曲加速器
	Charged    bool                       // 上一回合完全沒開火 → 這一回合可以連射
	Fired      bool                       // 這一回合開過火(回合結束時決定下一回合的 Charged)
	WeaponAmmo int                        // 本場剩餘彈藥；255 表示不受限制
	// --- 匿蹤(見 cloak.go)---
	CloakKind CloakKind // 隱形裝置 / 相位匿蹤 / 無
	Cloaked   bool      // 此刻是否隱形(開場為真,開火即失效,停火一整回合恢復)
	// --- 能量吸收器(見 energy_absorber.go)---
	EnergyAbsorber bool // 這艘船帶能量吸收器:被打時轉存 1/4 潛在傷害
	StoredEnergy   int  // 目前存著的能量(下一次開火時自動命中射出)
	// --- 登艦戰(見 boarding.go)---
	Marines          int  // 艦上陸戰隊單位數(手冊 p.121 的 Marines 欄,部隊艙翻倍)
	SecurityStations bool // 保安站:守方陸戰隊 +20
	MarineStrength   int  // 所屬帝國的陸戰隊科技／種族 Strength
	MarineHitsToKill int  // 所屬帝國的 High-G／Powered Armor 耐受值
	BoardingBonus    int  // 該艦艦員等級的 Bo
	CommandoBonus    int  // 同 owner 參戰艦軍官 Commando 最大值
	SecurityBonus    int  // 同 owner 參戰艦軍官 Security 最大值；只由守方消費
	AssaultShuttles  bool // 突擊艇:可以派登艦隊出去
	// SystemsDisabled 是被突襲拆掉的內部系統數(手冊:突襲不傷結構,只拆系統)。
	// >0 的船特殊系統失效——remake 一艘船只有一個特殊系統槽,所以第一下就拆光了。
	SystemsDisabled int
	Transporters    bool // 傳送器:12 格內,但面向攻擊方的護盾失效且非硬化護盾
	// 護盾分面(手冊 per facing;原版艦艇記錄四個連續值)。快速結算沒有 CombatShip
	// 與格位，因此只由格子戰術消費。
	ShieldFacingHP           [4]int
	ShieldFacingsInitialized bool
	// CausticSlimeStrength 是 combat record +0x43 的單場暫態。原版每回合以目前
	// 強度依序攻擊四個護盾面，之後減 5；重複命中會累加。
	CausticSlimeStrength int
	// Captured 是「這艘船已經被對方奪走」。手冊:奪船之後**還要贏下這場戰鬥**才留得住
	// (除非是心靈感應種族),所以這裡只記狀態,歸屬要等戰鬥結束才定案。
	Captured bool
}

const (
	cmbtshpSpriteBlockSize = 45
	cmbtshpPaletteHolder   = 44
	cmbtshpMaxPicture      = 43
	CMBTSHPFrameCount      = 20
	// CMBTSHPFrameHoldTicks 與 CMBTSHPMotionFrameCount 是 remake 的顯示
	// adapter，不是原版 timer 常數。原版靜態碼已證實 20 幀與 16 向 heading，
	// 但沒有追回獨立 tick；這裡用固定 tick 讓移動動畫可重播，停止後不自行旋轉。
	CMBTSHPFrameHoldTicks      = 4
	CMBTSHPMotionFrameCount    = 4
	CMBTSHPMotionDurationTicks = CMBTSHPFrameHoldTicks * CMBTSHPMotionFrameCount
)

// CMBTSHPSpriteIndex 是原版 sub_30062 @ 0x30062 的標準艦艇索引公式：
// CMBTSHP.LBX[45*playerColor + rawShipPicture]。raw picture 44 是原版
// monster.lbx sentinel，不屬於 CMBTSHP 的 44 張艦艇 sprite，因此回傳 false。
// colorBlock 的有效範圍是原版玩家色 0..7；這裡保留 raw 範圍檢查，避免畫廊再次
// 讀到 palette-holder 或越界資產。
func CMBTSHPSpriteIndex(colorBlock, rawPicture int) (int, bool) {
	if colorBlock < 0 || colorBlock >= 8 || rawPicture < 0 || rawPicture > cmbtshpMaxPicture {
		return 0, false
	}
	return cmbtshpSpriteBlockSize*colorBlock + rawPicture, true
}

func normalizeCMBTSHPColor(colorBlock, fallback int) int {
	if colorBlock >= 0 && colorBlock < 8 {
		return colorBlock
	}
	if fallback >= 0 && fallback < 8 {
		return fallback
	}
	return 0
}

// CMBTSHPFrameForHeading 將原版 combat record +0x23 的 16 向 heading 映到
// CMBTSHP 每個資產的 20 個方向幀。原版 `Move_Ship @ 0x3F5F1`／
// `Get_Facing @ 0x3F628` 已證實 heading 是 0..15；LBX 解碼已證實每個
// CMBTSHP 資產有 20 幀。這個最近角度換算是 remake 的顯示 adapter，幀的
// 原版繪製停留時間與四個未使用中間幀仍未由靜態碼單獨證實，因此不把它命名
// 成原版 timer。
func CMBTSHPFrameForHeading(heading int) int {
	heading %= 16
	if heading < 0 {
		heading += 16
	}
	return (heading*CMBTSHPFrameCount + 8) / 16 % CMBTSHPFrameCount
}

// CMBTSHPFrameAtTick 是 CMBTSHP 的 remake 動畫消費端。
//
// moving=false 時維持原本的最近角度幀；moving=true 時以固定 hold tick 播放
// [0,1,2,1] 的短掃掠，讓「移動中」有可見回饋。elapsedTicks 應從本次移動開始
// 計算，負值會視為 0。這是可重播的近似，不把未知的原版 20-frame timer 冒充
// 成已證實公式。
func CMBTSHPFrameAtTick(heading, elapsedTicks int, moving bool) int {
	base := CMBTSHPFrameForHeading(heading)
	if !moving || elapsedTicks < 0 {
		return base
	}
	phase := (elapsedTicks / CMBTSHPFrameHoldTicks) % CMBTSHPMotionFrameCount
	offset := [...]int{0, 1, 2, 1}[phase]
	return (base + offset) % CMBTSHPFrameCount
}

// CombatSpriteForShip 先使用原版 raw picture；舊 JSON／程序化 demo 沒有 raw
// picture 時才使用可視覺辨識的艦級 fallback。這個 fallback 不是原版公式，僅是
// 缺少輸入資料時的顯示降級。
func CombatSpriteForShip(ship Ship, colorBlock int) int {
	if ship.CombatPictureKnown {
		if idx, ok := CMBTSHPSpriteIndex(colorBlock, ship.CombatPicture); ok {
			return idx
		}
	}
	if colorBlock < 0 || colorBlock >= 8 {
		colorBlock = 0
	}
	return cmbtshpSpriteBlockSize*colorBlock + CombatSpriteForClass(ship.Class)
}

// CombatSpriteForClass 是 raw picture 未知時的艦級 fallback(不是原版精確
// picture 對照)。保留它是為了舊 JSON、程序化 demo 與沒有 `.GAM` 設計欄位的
// 敵方抽象艦隊仍有可辨識畫面。
func CombatSpriteForClass(class string) int {
	switch class {
	case "驅逐艦":
		return 12
	case "巡洋艦":
		return 20
	case "戰艦":
		return 28
	case "泰坦":
		return 36
	case "末日之星":
		return 43
	default:
		return 3 // 巡防艦/護衛艦/偵察艦/殖民船等小艦
	}
}

// CombatSpriteForStrength 依 genEnemyFleet 的戰力值反推近似艦級 → sprite 索引
// (shipStrength:巡防2/驅逐4/巡洋8/戰艦16/泰坦32/末日64)。
func CombatSpriteForStrength(st int) int {
	switch {
	case st >= 64:
		return 43
	case st >= 32:
		return 36
	case st >= 16:
		return 28
	case st >= 8:
		return 20
	case st >= 4:
		return 12
	default:
		return 3
	}
}

// enemyFighterTechnology 是 remake 抽象敵艦的最佳可用戰機科技。原版敵方逐艦藍圖
// 尚未完整取回，因此只依回合穩定推進，確保敵方戰機速度／裝甲不會永遠停在起始階。
func enemyFighterTechnology(turn int) (driveLevel, armorAboveTitanium int) {
	driveLevel = 1 + turn/8
	if driveLevel > gamedata.CombatSpeedDriveLevels {
		driveLevel = gamedata.CombatSpeedDriveLevels
	}
	armorAboveTitanium = turn / 12
	if armorAboveTitanium > 4 {
		armorAboveTitanium = 4
	}
	return driveLevel, armorAboveTitanium
}

func (s *GameSession) enemyRaceShipDefenseBonus(enemy string) int {
	for _, ai := range s.AIPlayers {
		if stripAILabel(ai.Name) != enemy && ai.Name != enemy {
			continue
		}
		if ai.RaceIndex >= 0 && ai.RaceIndex < len(Races) {
			return Races[ai.RaceIndex].ShipDefPct
		}
	}
	return 0
}

// StartCombat 依玩家艦隊 + 難度生成敵方,建立格子戰鬥雙方艦艇(HP=戰力×3、攻擊=戰力);
// 玩家艦置左欄、敵方置右欄,依序排列。敵方艦的戰機庫由
// EnemyFighterProfileForStrength 建立，戰術畫面會用同一組 CombatShip 欄位出擊。
func (s *GameSession) StartCombat(enemy string) (player, enemyShips []CombatShip) {
	// 由艦艇設計推導真戰鬥公式所需數值(remake 近似;精確值需艦體空間格 + 元件佔格 + 軍官技能):
	//   結構 HP = 艦體×3;裝甲 HP = 設計 BonusHP;Beam Attack = 艦體 + 武器攻擊;
	//   防禦 = 艦體(小艦=低戰力=低防,趨勢近原版);單發傷害 min=max/2、max=Attack;
	//   護盾減傷與艦級分面容量由設計的護盾名稱推導,四面狀態另存於 CombatShip。
	fighterPilotBonus := s.fighterPilotBonusForCombat()
	commando := fleetOfficerSkillMax(s.Leaders, s.Fleet().Ships, gamedata.SKILL_COMMANDO)
	security := fleetOfficerSkillMax(s.Leaders, s.Fleet().Ships, gamedata.SKILL_SECURITY)
	marineStrength := s.playerMarineForce()
	marineHits := gamedata.GroundMarineHitsToKill(s.raceHasTrait(gamedata.TRAIT_HIGH_G), s.hasPoweredArmor())
	for i, sh := range s.Fleet().Ships {
		body := shipStrength(sh.Class)
		// 原版 GameState::shipBeamOffense 讀的是這艘船自己的 sptr->officer,
		// 不是帝國內任一位艦艇軍官。Weaponry 是逐艦命中加成。
		atk := body + sh.WeaponAttack + shipBeamOffenseBonus(sh) +
			s.shipOfficerSkillBonus(sh, gamedata.SKILL_WEAPONRY)
		ordnanceBonus := s.shipOfficerSkillBonus(sh, gamedata.SKILL_ORDNANCE)
		hullHP := body * 3 * shipStructureMultiplier(sh) / 100
		// 戰機庫 → 這艘船在格子戰場上能派中隊出擊(見 fighter.go)。同一個 Special 欄位
		// 在快速結算裡是母艦戰力加成,兩條路徑讀同一份設計資料,不會各說各話。
		bays := fighterBaysForShip(sh)
		bay, bayKind := len(bays) > 0, FighterInterceptor
		if bay {
			bayKind = bays[0]
		}
		player = append(player, CombatShip{
			Name: sh.Name, HP: hullHP, MaxHP: hullHP, Attack: atk, Col: 1, Row: i, Facing: 0,
			Defense: body + shipBeamDefenseBonus(sh) +
				s.shipOfficerSkillBonus(sh, gamedata.SKILL_HELMSMAN), WeaponMin: atk / 2, WeaponMax: atk + ordnanceBonus,
			ShieldReduction: s.nebulaShield(shieldReduceByName(sh.Shield), shipHasHardShield(sh)) *
				shipShieldMultiplier(sh) / 100,
			HardShield: shipHasHardShield(sh),
			ArmorHP:    effectiveArmorHP(sh),
			Kind:       weaponKindByName(sh.Weapon), WeaponName: sh.Weapon, Mods: sh.Mods,
			WeaponArc:    NormalizeWeaponArc(sh.Weapon, sh.Arc),
			WeaponMounts: cloneWeaponMounts(sh.WeaponMounts), WeaponModes: NewTacticalWeaponModes(sh.WeaponMounts),
			HEF:       shipHasSpecial(sh, highEnergyFocusName),
			APNegated: shipNegatesArmorPiercing(sh),
			MissileEvasion: gamedata.ShipCrewMissileEvasionBonus(s.shipCrewLevel(sh)) +
				s.shipOfficerMissileEvasionBonus(sh) + shipMissileEvasionBonus(sh),
			HasAMR:                    shipHasSpecial(sh, antiMissileRocketName),
			HasLightningField:         shipHasLightningField(sh),
			HasDisplacement:           shipHasDisplacementDevice(sh),
			BeamSystems:               shipBeamAttackerSystems(sh),
			DriveLevel:                s.driveLevel(),
			ArmorLevelAboveTitanium:   armorLevelAboveTitanium(sh.Armor),
			CombatSpeed:               s.shipCombatSpeed(sh),
			SizeClass:                 shipSizeClass(sh.Class),
			FighterRacialDefenseBonus: s.RaceShipDefPct,
			FighterPilotBonus:         fighterPilotBonus,
			TractorBeams:              boolToInt(shipHasSpecial(sh, tractorBeamName)),
			HasStasisField:            shipHasSpecial(sh, stasisFieldName),
			ScannerJamReduction:       bestPlayerScannerJamReduction(s.Player),
			ShotsKind:                 shipShotsKind(sh),
			WeaponAmmo:                NormalizeWeaponAmmo(sh.Weapon, sh.WeaponAmmo),
			CloakKind:                 shipCloakKind(sh),
			Cloaked:                   shipCloakKind(sh) != CloakNone, // 開場隱形(手冊沒有「要先充能」)
			EnergyAbsorber:            shipHasSpecial(sh, energyAbsorberName),
			Marines:                   ShipMarineComplement(sh),
			SecurityStations:          shipHasSecurityStations(sh),
			MarineStrength:            marineStrength,
			MarineHitsToKill:          marineHits,
			BoardingBonus:             gamedata.ShipCrewBoardingBonus(s.shipCrewLevel(sh)),
			CommandoBonus:             commando,
			SecurityBonus:             security,
			AssaultShuttles:           shipHasAssaultShuttles(sh),
			Transporters:              shipHasTransporters(sh),
			Charged:                   true, // 開場滿電(手冊沒有「第一回合不能連射」的限制)
			Initiative:                gamedata.CombatInitiative(atk, s.shipCombatSpeed(sh)),
			SpriteIdx:                 CombatSpriteForShip(sh, normalizeCMBTSHPColor(s.FlagColor, 0)),
			Bay:                       bay, BayKind: bayKind, Bays: append([]FighterKind(nil), bays...),
		})
		player[len(player)-1].ensureShieldFacings()
	}
	enemyShipDef := s.enemyRaceShipDefenseBonus(enemy)
	if aiIdx := s.aiIndexByName(enemy); aiIdx >= 0 {
		enemyShips = append(enemyShips, s.aiTacticalShips(aiIdx)...)
		return
	}
	enemyDrive, enemyArmor := enemyFighterTechnology(s.Turn)
	enemyColor := 1
	for _, ai := range s.AIPlayers {
		if stripAILabel(ai.Name) == enemy || ai.Name == enemy {
			if ai.ColorKnown {
				enemyColor = normalizeCMBTSHPColor(ai.Color, 1)
			}
			break
		}
	}
	for i, st := range genEnemyFleet(s.Turn) {
		bayKind, hasBay := EnemyFighterProfileForStrength(st)
		class := shipSizeClassFromStrength(st)
		sizeClass := gamedata.CombatShipClass(class - 1)
		combatSpeed := gamedata.ShipCombatSpeed(enemyDrive, sizeClass, false, false)
		ship := CombatShip{
			Name: fmt.Sprintf("%s艦%d", enemy, i+1), HP: st * 3, MaxHP: st * 3, Attack: st, Col: 6, Row: i % TacticalGridRows, Facing: 8,
			Defense: st + st*enemyShipDef/100, WeaponMin: st / 2, WeaponMax: st, ShieldReduction: 0, ArmorHP: st,
			Kind: WeaponKindBeam, WeaponName: "雷射", WeaponArc: gamedata.ARC_FWD,
			DriveLevel: enemyDrive, ArmorLevelAboveTitanium: enemyArmor,
			CombatSpeed: combatSpeed, Initiative: gamedata.CombatInitiative(st, combatSpeed), SizeClass: sizeClass,
			FighterRacialDefenseBonus: enemyShipDef,
			Bay:                       hasBay, BayKind: bayKind,
			SpriteIdx: cmbtshpSpriteBlockSize*enemyColor + CombatSpriteForStrength(st),
		}
		enemyShips = append(enemyShips, ship)
	}
	return
}

// ApplyCombatOutcome 是舊命令／測試相容入口；新戰術畫面改用
// ApplyCombatOutcomeWithEnemySurvivors 同時回寫 AI 實艦。
func (s *GameSession) ApplyCombatOutcome(enemy string, playerStart, enemyStart int, survivors map[string]bool, won bool, destroyedHullClassSum ...int) {
	s.applyCombatOutcome(enemy, playerStart, enemyStart, survivors, nil, won, true, destroyedHullClassSum...)
}

func (s *GameSession) ApplyCombatOutcomeWithEnemySurvivors(enemy string, playerStart, enemyStart int,
	survivors, enemySurvivors map[string]bool, won bool, destroyedHullClassSum ...int) {
	s.applyCombatOutcome(enemy, playerStart, enemyStart, survivors, enemySurvivors, won, true, destroyedHullClassSum...)
}

func (s *GameSession) applyCombatOutcome(enemy string, playerStart, enemyStart int,
	survivors, enemySurvivors map[string]bool, won, record bool, destroyedHullClassSum ...int) {
	sum := 0
	if len(destroyedHullClassSum) > 0 {
		sum = destroyedHullClassSum[0]
	} else if won {
		// 舊 command 沒有記錄總和；戰術勝利代表敵方序列已清空。無法還原歷史俘獲，
		// 故以當回合初始敵艦完整總和作向後相容近似。
		for _, strength := range genEnemyFleet(s.Turn) {
			sum += shipSizeClassFromStrength(strength)
		}
	}
	parts := make([]string, 0, len(survivors)+1)
	parts = append(parts, enemy)
	for name, alive := range survivors {
		if alive {
			parts = append(parts, name)
		}
	}
	sort.Strings(parts[1:])
	if enemySurvivors != nil {
		parts = append(parts, "\x01ENEMY")
		enemyNames := make([]string, 0, len(enemySurvivors))
		for name, alive := range enemySurvivors {
			if alive {
				enemyNames = append(enemyNames, name)
			}
		}
		sort.Strings(enemyNames)
		parts = append(parts, enemyNames...)
	}
	if record {
		s.recordPlayerCommand(PlayerCommand{
			Name: CmdCombatOutcome, Args: []int{playerStart, enemyStart, boolInt(won), sum}, Text: strings.Join(parts, "\x00"),
		})
	}
	f := s.Fleet() // 戰鬥打的是**參戰的那一支**艦隊
	kept := f.Ships[:0]
	eligible := map[int]bool{}
	combatSurvivors := 0
	for _, sh := range f.Ships {
		if survivors[sh.Name] {
			kept = append(kept, sh)
			if !isSupportShipClass(sh.Class) && survivors[sh.Name] {
				eligible[len(kept)-1] = true
				combatSurvivors++
			}
		}
	}
	f.Ships = kept
	if aiIdx := s.aiIndexByName(enemy); aiIdx >= 0 && (enemySurvivors != nil || won) {
		keptAI := s.AIPlayers[aiIdx].Ships[:0]
		for _, sh := range s.AIPlayers[aiIdx].Ships {
			if enemySurvivors != nil && enemySurvivors[sh.Name] {
				keptAI = append(keptAI, sh)
			}
		}
		s.AIPlayers[aiIdx].Ships = keptAI
		s.syncAIShipStrength(aiIdx)
	}
	crewXP := 0
	if won {
		crewXP = s.awardBattleCrewXP(sum, eligible)
	}
	enemyLosses := enemyStart
	if enemySurvivors != nil {
		enemyLosses = enemyStart - len(enemySurvivors)
	}
	s.LastBattle = &BattleResult{
		Enemy: enemy, PlayerStart: playerStart, EnemyStart: enemyStart, PlayerWon: won,
		PlayerLosses: playerStart - combatSurvivors, EnemyLosses: enemyLosses, CrewXPGained: crewXP,
	}
}

// PrimaryEnemyName 回傳戰鬥/外交畫面顯示用的「主要對手」名稱。取第一個 AI 對手的種族名,
// 去掉 demoAIOpponentSetup 的「AI (…)」外殼(戰鬥標籤前綴接「艦N」時較自然,如「席隆人艦1」
// 而非「AI (席隆人)艦1」)。無 AI 對手時 fallback「敵軍」——避免舊硬編「賽隆人」(Races 表裡
// 根本不存在的錯字,見 demoAIOpponentSetup 註解)顯示在戰鬥畫面。
//
// 一般戰鬥會以此名稱尋找對應 AI 的持久實艦；目前 UI 仍固定取第一位 AI，尚未提供多 AI
// 目標選擇。沒有對應帝國的特殊腳本敵人才退回 genEnemyFleet。
// localName 把原版的英文專有名詞(星名、艦名)翻成當前語言。
//
// ⚠ **翻譯發生在「取名的那一刻」,不是顯示的那一刻。** 兩者的差別在存檔:
// 前者讓存檔裡直接是玩家看到的那個字串,後者會讓存檔存英文而顯示時再翻。
// 選前者的理由有兩個——
//
//  1. **中文模式的輸出與這一輪之前逐位元相同**(畫廊 34 張比對是證據)。
//     如果改成顯示時翻,存檔內容會變,狀態指紋會變,而那個變動沒有任何遊戲意義。
//  2. 星名與艦名在遊戲裡會被玩家改(艦隊改名)、會進戰報字串、會進網路封包——
//     那些地方沒有一個知道「這個字串是不是可翻的專有名詞」。存成最終字串最單純。
//
// TranslateName 為 nil(預設)時是恆等函式,也就是**英文**。`internal/shell` 因此
// 不需要 import i18n:由 cmd/moo2 在建立對局時注入(見 sceneBuilder)。
// nameTranslator 回傳這一局實際生效的專有名詞翻譯器(逐局覆寫優先於行程層預設)。
func (s *GameSession) nameTranslator() func(string) string {
	if s.TranslateName != nil {
		return s.TranslateName
	}
	return defaultNameTranslator
}

func (s *GameSession) localName(en string) string {
	tr := s.nameTranslator()
	if tr == nil {
		return en
	}
	return tr(en)
}

// RelationToPlayer 依對手名回傳它對玩家的關係分數(−40..40);查無回 0。
//
// 供外交畫面挑背景音樂用(原版 `Start_Diplomacy_Music_` 依關係好壞切換 good/bad 兩首,
// 見 cmd/moo2/audiohook.go 的 playDiplomacyMusic)。
func (s *GameSession) RelationToPlayer(enemy string) int {
	for i := range s.AIPlayers {
		if s.AIPlayers[i].Name == enemy {
			return s.AIPlayers[i].Relation
		}
	}
	return 0
}

func (s *GameSession) PrimaryEnemyName() string {
	if len(s.AIPlayers) == 0 {
		return "敵軍"
	}
	return stripAILabel(s.AIPlayers[0].Name)
}

// stripAILabel 去掉 demoAIOpponentSetup 的「AI (…)」外殼,只留種族名。
// 兩個呼叫端:戰鬥標籤(PrimaryEnemyName)與熱座接管 AI 帝國時的席位命名(seatFromAI)
// ——後者若不去掉,交接畫面會寫「下一位:AI(布拉西人)」,但接手的是真人。
func stripAILabel(name string) string {
	if strings.HasPrefix(name, "AI (") && strings.HasSuffix(name, ")") {
		return name[len("AI (") : len(name)-len(")")]
	}
	return name
}

// ShipCost 造某艦體等級所需生產成本(MOO2 空殼艦體生產成本,每級約 ×3:
// 巡防18/驅逐60/巡洋180/戰艦540/泰坦1620/末日之星4860)。
func ShipCost(class string) int {
	switch class {
	case "巡防艦", "護衛艦":
		return 18
	case "驅逐艦":
		return 60
	case "巡洋艦":
		return 180
	case "戰艦":
		return 540
	case "泰坦":
		return 1620
	case "末日之星":
		return 4860
	case "偵察艦":
		return 10
	}
	return 18
}

func pick(opts []Component, i int) Component {
	if i >= 0 && i < len(opts) {
		return opts[i]
	}
	return opts[0]
}

// shipClassFromName 把艦體等級中文名對應到 gamedata.CombatShipClass(供空間驗證查表用)。
// 未知/不在手冊 6 個 Design 艦級表內的艦體(如「偵察艦」Scout,手冊 Ship Design 章節只列
// Frigate..Doom Star 六級,Scout 屬另建的支援艦,無獨立空間表)一律以 Frigate 空間近似
// (最保守、最小的手冊數值),並回傳 ok=false 供呼叫端知道這是近似對應。
func shipClassFromName(class string) (c gamedata.CombatShipClass, ok bool) {
	switch class {
	case "巡防艦", "護衛艦":
		return gamedata.SHIP_FRIGATE, true
	case "驅逐艦":
		return gamedata.SHIP_DESTROYER, true
	case "巡洋艦":
		return gamedata.SHIP_CRUISER, true
	case "戰艦":
		return gamedata.SHIP_BATTLESHIP, true
	case "泰坦":
		return gamedata.SHIP_TITAN, true
	case "末日之星":
		return gamedata.SHIP_DOOMSTAR, true
	case ColonyShipClass, OutpostShipClass:
		// 手冊 p.41 逐字:「Colony Ships, Transports and Outpost Ships count as Frigate class
		// ships」。這一條先前是靠下面的 fallback 碰巧算對,現在有手冊出處就明寫出來——
		// 它同時決定這兩種支援艦的指揮點數需求(手冊 p.85:各 1 點)。
		return gamedata.SHIP_FRIGATE, true
	}
	return gamedata.SHIP_FRIGATE, false // 例:偵察艦,近似值,非手冊確認
}

// ShipDesignSpaceUsed 回傳一組元件選擇(武器/裝甲/護盾/特殊)已用的艦體空間總和(無武器改造)。
// 委派 ShipDesignSpaceUsedWithMods(mods=nil),行為與加入 mods 系統前完全相同(回歸安全)。
//
// 依 GAME_MANUAL.pdf p.121-122(見 internal/gamedata/shipspace.go 檔頭 [HARD 誠實原則 2]):
// 裝甲(armor)與護盾(shield)在原版是「Automatics」,自動裝上目前科技最好的一套,不佔用
// Weapons/Specials 共用的空間預算——因此本函式的 armor/shield 參數目前一律不計入空間(回報 0),
// 只是為了與既有四下拉呼叫介面(DesignCost/BuildShip)保持一致的簽名,不是遺漏或臆造。
// 真正佔空間的是武器(gamedata.WeaponSpaceByName,手冊 Size 欄確認值)與特殊系統
// (gamedata.SpecialSpace,估計值,見該函式註解)。
func ShipDesignSpaceUsed(class string, weapon, armor, shield, special int) int {
	return ShipDesignSpaceUsedWithMods(class, weapon, armor, shield, special, nil)
}

// ShipDesignSpaceUsedWithMods 同 ShipDesignSpaceUsed,額外套用一組武器改造(mods,見
// gamedata.WeaponModCode / docs/tech/weapon-mods.md)對武器佔格的影響
// (gamedata.WeaponSpaceWithMods)。先依武器類型過濾改造；不支援的歷史字串不誤加空間。
func ShipDesignSpaceUsedWithMods(class string, weapon, armor, shield, special int, mods []string) int {
	return ShipDesignSpaceUsedWithModsAndArc(class, weapon, armor, shield, special, mods, gamedata.ARC_FWD)
}

// ShipDesignSpaceUsedWithModsAndArc 在武器改造後再套用火線角佔格。
// 舊入口固定用前向基準弧，保留既有呼叫端的數值回歸；艦艇設計畫面使用本入口。
func ShipDesignSpaceUsedWithModsAndArc(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc) int {
	w := pick(WeaponOptions, weapon)
	return ShipDesignSpaceUsedWithAmmo(class, weapon, armor, shield, special, mods, arc,
		NormalizeWeaponAmmo(w.Name, 0))
}

// ShipDesignSpaceUsedWithAmmo 是無玩家科技狀態、但含彈架容量的基礎入口。
func ShipDesignSpaceUsedWithAmmo(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc, ammo int) int {
	return shipDesignSpaceUsedWithMiniLevel(class, weapon, armor, shield, special, mods, arc, 0, 0, ammo)
}

func shipDesignSpaceUsedWithMiniLevel(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc, weaponMiniLevel, specialMiniLevel, ammo int) int {
	_ = armor // 見上方註解:手冊行為上裝甲不佔空間,顯式忽略以避免「未使用參數」誤解成疏漏
	_ = shield
	w := pick(WeaponOptions, weapon)
	sp := pick(SpecialOptions, special)
	classID, _ := shipClassFromName(class)
	base := gamedata.WeaponSpaceByName[w.Name]
	if WeaponUsesVariableMissileRack(w.Name) {
		base, _ = gamedata.MissileRackBaseValue(NormalizeWeaponAmmo(w.Name, ammo))
	}
	weaponSpace := base
	if len(mods) > 0 {
		weaponSpace = gamedata.WeaponSpaceWithMods(base, WeaponModCodesForWeapon(w.Name, mods))
	}
	weaponSpace = gamedata.WeaponSpaceAtMiniLevelForCategory(weaponSpace, weaponMiniLevel,
		gamedata.MiniaturizationSpaceCategoryForTech(w.UnlockTech))
	weaponSpace = gamedata.WeaponArcAdjustedValue(weaponSpace, NormalizeWeaponArc(w.Name, arc))
	used := weaponSpace
	// 特殊系統佔格改讀**原版表的真值**(依艦級,見 special_device_map.go)。
	// 先前走 gamedata.SpecialSpace 的 5% 估計——那個估計的註解自己就寫著「這不是手冊數字」。
	//
	// ⚠ 這裡可能是**負值**(戰鬥艙):手冊「add equipment space without increasing the
	// hull size」在原版就是做成負佔格。加上去讓總和變小是對的,不要在這裡夾成 0。
	specialSpace := specialDeviceSpaceFor(sp, classID)
	specialSpace = gamedata.WeaponSpaceAtMiniLevelForCategory(specialSpace, specialMiniLevel,
		gamedata.MiniaturizationSpaceCategoryForTech(sp.UnlockTech))
	used += specialSpace
	return used
}

// ShipDesignFits 回傳一組元件選擇是否能塞進指定艦體(已用空間 <= 艦體總空間,無武器改造)。
// 未知艦體等級(shipClassFromName 回傳 ok=false,如偵察艦)以 Frigate 空間近似判定,
// 保守地拒絕過大的設計;供 UI 判斷是否標記「不可建造」用。
func ShipDesignFits(class string, weapon, armor, shield, special int) bool {
	return ShipDesignFitsWithMods(class, weapon, armor, shield, special, nil)
}

// ShipDesignFitsWithMods 同 ShipDesignFits,套用武器改造的佔格變動(見
// ShipDesignSpaceUsedWithMods)。掛 Heavy Mount/Enveloping 等增加佔格的 mod 可能讓原本
// 塞得下的設計超格,藉此讓 UI/建造流程仍然擋下超格設計。
func ShipDesignFitsWithMods(class string, weapon, armor, shield, special int, mods []string) bool {
	return ShipDesignFitsWithModsAndArc(class, weapon, armor, shield, special, mods, gamedata.ARC_FWD)
}

// ShipDesignFitsWithModsAndArc 是含火線角的艦體空間判定。
func ShipDesignFitsWithModsAndArc(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc) bool {
	classID, _ := shipClassFromName(class)
	hullSpace := gamedata.ShipHullSpace(classID)
	return ShipDesignSpaceUsedWithModsAndArc(class, weapon, armor, shield, special, mods, arc) <= hullSpace
}

// DesignCost 回傳一組元件選擇(艦體 + 武器/裝甲/護盾/特殊)的總生產成本(無武器改造)。
func DesignCost(class string, weapon, armor, shield, special int) int {
	return DesignCostWithMods(class, weapon, armor, shield, special, nil)
}

// DesignCostWithMods 同 DesignCost,套用武器改造對成本的影響(手冊「adds to the size AND
// cost」,與佔格用同一套百分比,見 gamedata.WeaponCostWithMods)。
func DesignCostWithMods(class string, weapon, armor, shield, special int, mods []string) int {
	return DesignCostWithModsAndArc(class, weapon, armor, shield, special, mods, gamedata.ARC_FWD)
}

// DesignCostWithModsAndArc 是含火線角的艦艇總成本。
func DesignCostWithModsAndArc(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc) int {
	w := pick(WeaponOptions, weapon)
	return DesignCostWithAmmo(class, weapon, armor, shield, special, mods, arc,
		NormalizeWeaponAmmo(w.Name, 0))
}

// DesignCostWithAmmo 是無玩家科技狀態、但含彈架容量的基礎入口。
func DesignCostWithAmmo(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc, ammo int) int {
	return designCostWithMiniLevel(class, weapon, armor, shield, special, mods, arc, 0, 0, ammo)
}

func designCostWithMiniLevel(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc, weaponMiniLevel, specialMiniLevel, ammo int) int {
	w := pick(WeaponOptions, weapon)
	weaponCost := w.Cost
	if WeaponUsesVariableMissileRack(w.Name) {
		weaponCost, _ = gamedata.MissileRackBaseValue(NormalizeWeaponAmmo(w.Name, ammo))
	}
	if len(mods) > 0 {
		weaponCost = gamedata.WeaponCostWithMods(weaponCost, WeaponModCodesForWeapon(w.Name, mods))
	}
	weaponCost = gamedata.WeaponCostAtMiniLevel(weaponCost, weaponMiniLevel)
	weaponCost = gamedata.WeaponArcAdjustedValue(weaponCost, NormalizeWeaponArc(w.Name, arc))
	// 特殊系統成本改讀**原版表的真值**(依艦級,見 special_device_map.go)。原版的成本
	// 隨艦體等級變動——同一套系統裝在末日之星上比裝在巡防艦上貴一個數量級,
	// 而 Component.Cost 是單一數字。對不上原版表的幾項仍走 Component.Cost。
	classID, _ := shipClassFromName(class)
	specialCost := gamedata.WeaponCostAtMiniLevel(specialDeviceCostFor(pick(SpecialOptions, special), classID), specialMiniLevel)
	return ShipCost(class) + weaponCost + pick(ArmorOptions, armor).Cost +
		pick(ShieldOptions, shield).Cost +
		specialCost
}

// BuildShip 造一艘指定艦體 + 全元件(武器/裝甲/護盾/特殊)的艦:扣國庫總成本,加入艦隊。
// BC 不足回 false。武器加攻擊、裝甲+護盾加 HP、特殊「戰鬥電腦」再加攻擊。無武器改造。
func (s *GameSession) BuildShip(class string, weapon, armor, shield, special int) bool {
	return s.BuildShipWithMods(class, weapon, armor, shield, special, nil)
}

// BuildShipWithMods 同 BuildShip,額外把 mods(武器改造)存進造出的 Ship.Mods,並用
// DesignCostWithMods 算入改造增加/減少的成本。正常建造入口會依武器類型與微型化門檻再次
// 過濾，避免舊存檔或重播命令繞過 UI。
func (s *GameSession) BuildShipWithMods(class string, weapon, armor, shield, special int, mods []string) bool {
	return s.BuildShipWithModsAndArc(class, weapon, armor, shield, special, mods, gamedata.ARC_FWD)
}

// BuildShipWithModsAndArc 建造並保存一艘含火線角的艦艇。
func (s *GameSession) BuildShipWithModsAndArc(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc) bool {
	w := pick(BuildWeaponOptions(s.RuleProfile), weapon)
	return s.BuildShipWithLoadout(class, weapon, armor, shield, special, mods, arc,
		NormalizeWeaponAmmo(w.Name, 0))
}

// BuildShipWithLoadout 建造並保存一艘含火線角與彈架容量的艦艇。
func (s *GameSession) BuildShipWithLoadout(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc, ammo int) bool {
	// 武器傷害(w.Value)吃這局遊戲的版本規則 profile(s.RuleProfile,見 BuildWeaponOptions 註解:
	// 電漿砲 1.3=30/1.5=20,其餘元件與套件級 WeaponOptions 逐一相同)——造艦時真正掛上版本相依
	// 傷害值,不再永遠是套件級硬編的 1.5 值。成本(DesignCostWithMods)不受影響:兩版電漿砲 Cost
	// 相同,差異只在 Value,見 ruleprofile.go RuleProfile.PlasmaCannonMaxDamage 註解。
	w, a, sh, sp := pick(BuildWeaponOptions(s.RuleProfile), weapon), pick(ArmorOptions, armor), pick(ShieldOptions, shield), pick(SpecialOptions, special)
	arc = NormalizeWeaponArc(w.Name, arc)
	ammo = NormalizeWeaponAmmo(w.Name, ammo)
	level := weaponMiniaturizationLevel(w, s.Player.CompletedTopics, s.Player.HyperAdvancedLevels)
	mods = filterWeaponModsAtLevel(w.Name, mods, level)
	text := class + "\x00" + strings.Join(mods, "\x00")
	s.recordPlayerCommand(PlayerCommand{
		Name: CmdBuildShip, Args: []int{weapon, armor, shield, special, ammo, int(arc)}, Text: text,
	})
	specialLevel := weaponMiniaturizationLevel(sp, s.Player.CompletedTopics, s.Player.HyperAdvancedLevels)
	cost := designCostWithMiniLevel(class, weapon, armor, shield, special, mods, arc, level, specialLevel, ammo)
	if s.Player.BC < cost {
		return false
	}
	s.Player.BC -= cost
	// 全帝國計數,不然分艦隊後會撞名。池子存英文原文,這裡翻成當前語言
	// ——**翻譯發生在造艦當下**,所以存檔裡的艦名就是玩家看到的那個語言(見 localName)。
	name := s.localName(shipNamePool[s.ShipCount()%len(shipNamePool)])
	atk := w.Value
	if sp.Name == "戰鬥電腦" {
		atk += sp.Value
	}
	var modsCopy []string
	if len(mods) > 0 {
		modsCopy = append([]string(nil), mods...)
	}
	// 艦艇設計畫面直接花錢造的船,進**目前操作中**的艦隊(玩家按下建造時手上就是那一支)。
	f := s.Fleet()
	// 艦員起始經驗:這條路徑是「艦艇設計畫面直接花錢造」,沒有指定殖民地
	// ——太空學院是逐殖民地的建築,查不到是哪一座造的,所以走一般起始等級。
	// ⚠ 這是 remake 的路徑限制(設計畫面沒有「在哪造」的概念),不是規則如此;
	// 逐殖民地造艦那條路(deliverNewShip)有正確吃到學院加成。
	f.Ships = append(f.Ships, Ship{Name: name, Class: class, Weapon: w.Name, Armor: a.Name, Shield: sh.Name,
		RawType: gamedata.COMBAT_SHIP, RawTypeKnown: true, RawMissionKnown: true, ProductionCost: cost,
		Special: sp.Name, WeaponAttack: atk, BonusHP: a.Value + sh.Value, Mods: modsCopy,
		Arc: arc, WeaponAmmo: ammo,
		WeaponMounts: []ShipWeaponMount{{RawType: -1, Name: w.Name, MaxCount: 1, WorkingCount: 1,
			Arc: arc, Ammo: ammo, Attack: w.Value}},
		CrewXP: s.newShipCrewXP(-1)})
	return true
}

// ShiftColonyJob 在某殖民地把 1 名人口從 from 職務移到 to(f=農夫 w=工人 s=科學家);
// from 需有人。供殖民地人口重分配(影響下回合經濟)。
func (s *GameSession) ShiftColonyJob(idx int, from, to string) {
	s.recordPlayerCommand(PlayerCommand{Name: CmdShiftJob, Args: []int{idx}, Text: from + ">" + to})
	if idx < 0 || idx >= len(s.PlayerColonies) {
		return
	}
	c := &s.PlayerColonies[idx]
	get := func(j string) *int {
		switch j {
		case "f":
			return &c.Farmers
		case "w":
			return &c.Workers
		case "s":
			return &c.Scientists
		}
		return nil
	}
	fp, tp := get(from), get(to)
	if fp != nil && tp != nil && *fp > 0 {
		job := func(j string) gamedata.ColonistJob {
			switch j {
			case "w":
				return gamedata.WORKER
			case "s":
				return gamedata.SCIENTIST
			default:
				return gamedata.FARMER
			}
		}
		groupShifted := engine.ShiftPopulationGroupJob(c, job(from), job(to))
		getPrisoners := func(j string) *int {
			switch j {
			case "f":
				return &c.UnassimilatedFarmers
			case "w":
				return &c.UnassimilatedWorkers
			case "s":
				return &c.UnassimilatedScientists
			}
			return nil
		}
		fromPrisoners, toPrisoners := getPrisoners(from), getPrisoners(to)
		if !groupShifted && fromPrisoners != nil && toPrisoners != nil && *fromPrisoners > 0 &&
			c.UnassimilatedFarmers+c.UnassimilatedWorkers+c.UnassimilatedScientists == c.UnassimilatedPop {
			*fromPrisoners--
			*toPrisoners++
		}
		*fp--
		*tp++
	}
}

// 銀河議會選舉勝利條件(成立門檻/2-3方候選/2/3多數/勝利判定)見 council.go
// ——2026-07-11 取代這裡原本「票數=人口、較高者當選」的簡化版(無成立門檻、無2/3多數、
// 未接勝利判定,對照 GAME_MANUAL.pdf p.183 手冊原文是錯誤示範,已移除)。

// RefitJob 是一筆排入殖民地建造佇列的艦艇改裝工作。
//
// Source 保留在佇列裡而非艦隊中：這使艦艇在改裝期間不可出戰／移動，而且若玩家
// 從佇列移除這筆工作，就能遵守原版手冊所述的「取消改裝會毀掉該艦」。Target 是
// 已在排程當下凍結的目標設計，避免之後研究完成或調整設計畫面時讓同一存檔的
// 改裝結果漂移。ReturnStar 是完工後回到軌道的星系；若原艦隊已離開，會在該星
// 建立一支新艦隊而不是把船送到不相干的艦隊。
//
// 目標設計的選擇規則見 production_controls.go：原版允許從設計庫挑同艦體設計，
// remake 目前沒有可持久化的設計庫，故採「目前已解鎖的自動最佳模板」近似。
type RefitJob struct {
	Source     Ship
	Target     Ship
	ReturnStar int
}

// ColonyBuild 是某殖民地目前的建造項目。
type ColonyBuild struct {
	Name string
	// ProductKind 保存不應以本地化顯示文字判斷的特殊產品語意。空值代表歷史存檔的
	// 一般建築／特殊行動；穩定識別字不顯示給玩家。
	ProductKind  ColonyProductKind `json:"productKind,omitempty"`
	Progress     int
	ProgressHalf int // 半機械族建造進度的半單位餘數；舊存檔缺欄位時為 0
	Cost         int
	// Refit 非 nil 時本項是艦艇改裝，而非同名建築。保留在 ColonyBuild 裡讓既有
	// Builds / BuildQueue 的 JSON、熱座與網路指令都走同一條可保存佇列。
	Refit *RefitJob
}

type ColonyProductKind string

const (
	ColonyProductAIAgent ColonyProductKind = "ai_agent"
)

// TradeGoodsBuildName 是「貿易品」建造佇列選項的名稱。與空字串「不建造」同類——是佇列的
// 特殊選擇而非 gamedata.Buildings 裡的實體建築,恆可選、無前置科技 gate(手冊 GAME_MANUAL.pdf
// p.70:貿易品是把殖民地產能整包轉現金的建造選項)。Cost 固定 0,讓既有的
// advanceBuilds「b.Cost == 0 則不累積進度」判斷同時涵蓋「不建造」與「貿易品」兩種特殊選項,
// 不需要另外用名稱比對(見 advanceBuilds 註解)。
const TradeGoodsBuildName = "貿易品"

// HousingBuildName 是「住宅」建造佇列選項的名稱。與「貿易品」「不建造」同類——是佇列的
// 恆可選特殊項,不是 gamedata.Buildings 裡的實體建築,沒有前置科技 gate。
//
// 手冊把住宅與貿易品並列為「repetitive items」(patch1.5 changelog:"repetitive items
// (housing, trade goods or repeat builds) in the queue"),選它時該殖民地的淨工業不蓋建築、
// 改成加速人口成長(engine.ColonyState.Housing → gamedata.ColonyHousingBonus)。
//
// 為什麼補這一項:remake 的住房獎金公式與 ColonyState.Housing 欄位早就寫好,卻**從來沒有
// 任何地方設過那個欄位**——因為建造選單裡沒有「住宅」。開局玩家科技只解鎖 2 個建築、
// 而母星兩個都已蓋好,建造選單實際上只剩「貿易品」與「不建造」兩項可選(2026-08-06 實測)。
const HousingBuildName = "住宅"

// buildOptions 是「不看前置科技」的全部可建項目(名稱 + 生產成本),衍生自
// gamedata.Buildings(手冊全表 40 項:35 建築 + 5 衛星),空字串為「不建造」排第一個,
// 「貿易品」特殊選項排第二個。供將來「完整建築圖鑑」類 UI 參考;實際建造選單(有前置科技
// gate)請用 availableBuildOptions,CycleColonyBuild 已改用該函式。
var buildOptions = allBuildOptions()

// allBuildOptions 把 gamedata.Buildings + gamedata.SpecialActions 轉成 ColonyBuild 選項清單
// (含「不建造」「貿易品」兩個非建築特殊項於前兩位)。SpecialActions(地形改造/蓋亞轉化/
// 土壤改良/運輸艦隊)排在 Buildings 之後——它們是「Special」型別的一次性行動,不是常駐建築,但
// 同樣走殖民地建造佇列選單,見 gamedata/special_actions.go 檔頭說明。
func allBuildOptions() []ColonyBuild {
	out := make([]ColonyBuild, 0, len(gamedata.Buildings)+len(gamedata.SpecialActions)+2)
	out = append(out, ColonyBuild{Name: "", Progress: 0, Cost: 0})
	out = append(out, ColonyBuild{Name: TradeGoodsBuildName, Progress: 0, Cost: 0})
	out = append(out, ColonyBuild{Name: HousingBuildName, Progress: 0, Cost: 0})
	for _, b := range gamedata.Buildings {
		out = append(out, ColonyBuild{Name: b.NameZH, Progress: 0, Cost: b.ProductionCost})
	}
	for _, a := range gamedata.SpecialActions {
		out = append(out, ColonyBuild{Name: a.NameZH, Progress: 0, Cost: a.ProductionCost})
	}
	return out
}

// availableBuildOptions 回傳「玩家已研究前置科技」才會出現的建造選單(「不建造」「貿易品」
// 兩個特殊選項恆在,不受前置科技限制)。地形改造/蓋亞轉化/土壤改良/運輸艦隊比照建築同款前置
// 科技 gate(gamedata.AvailableSpecialActions),排在建築清單之後。
func availableBuildOptions(completedTopics map[gamedata.ResearchTopic]bool) []ColonyBuild {
	out := []ColonyBuild{
		{Name: "", Progress: 0, Cost: 0},
		{Name: TradeGoodsBuildName, Progress: 0, Cost: 0},
		{Name: HousingBuildName, Progress: 0, Cost: 0},
	}
	for _, b := range gamedata.AvailableBuildings(completedTopics) {
		out = append(out, ColonyBuild{Name: b.NameZH, Progress: 0, Cost: b.ProductionCost})
	}
	for _, a := range gamedata.AvailableSpecialActions(completedTopics) {
		out = append(out, ColonyBuild{Name: a.NameZH, Progress: 0, Cost: a.ProductionCost})
	}
	return out
}

func (s *GameSession) availableBuildOptionsForColony(colony int) []ColonyBuild {
	out := availableBuildOptions(s.Player.CompletedTopics)
	if !s.playerCanBuildCapitol(colony) {
		return out
	}
	// 原版 Capitol 是帝國固有 raw 9，不屬一般科技建築表；放在三個恆可選項之後。
	insert := 3
	if insert > len(out) {
		insert = len(out)
	}
	out = append(out, ColonyBuild{})
	copy(out[insert+1:], out[insert:])
	out[insert] = ColonyBuild{Name: CapitolBuildName, Cost: CapitolProductionCost}
	return out
}

// 起始文明等級的殖民地開局建築數上限(不含 Capitol),依 docs/tech/homeworld-init.md §2.2
// (MANUAL_150.html「Initial Buildings」段,一手來源):
// "The number of starting buildings on each colony is capped to 3 for Pre-warp,
// 5 for Average/Postwarp and 9 for Advanced game starts."
const (
	BuildingCapPreWarp  = 3
	BuildingCapAverage  = 5
	BuildingCapPostWarp = 5
	BuildingCapAdvanced = 9
)

// StartingBuildingCount 依手冊「Initial Buildings」公式算出某殖民地開局建築數(不含 Capitol):
// min(⅔ pop 無條件進位, 該起始等級上限)。手冊原文驗證範例(docs/tech/homeworld-init.md §3.5):
// 「a HW with 8 pop can have 6 buildings on Advanced Tech start, but only 5 on Average start
// due to the cap」——即 StartingBuildingCount(8, BuildingCapAdvanced)==6、
// StartingBuildingCount(8, BuildingCapAverage)==5,已寫進本套件單元測試。
// 注意:此函式只回傳「上限」,實際會生成哪些建築仍取決於 initial_buildings 優先清單與
// 已知科技(§3.3:Pre-warp/Average 僅 Marine Barracks + Star Base 兩項符合條件,即使
// 上限允許更多)。
func StartingBuildingCount(pop, cap int) int {
	if pop < 0 {
		pop = 0
	}
	n := (pop*2 + 2) / 3 // ⅔ pop 無條件進位
	if n > cap {
		return cap
	}
	return n
}

// CycleColonyBuild 循環切換某殖民地的建造項目(進度歸零)。選項依玩家目前已完成研究 gate
// (availableBuildOptions):尚未解鎖前置科技的建築不會出現在循環清單中。
func (s *GameSession) CycleColonyBuild(idx int) {
	s.recordPlayerCommand(PlayerCommand{Name: CmdCycleColonyBuild, Args: []int{idx}})
	if idx < 0 || idx >= len(s.Builds) {
		return
	}
	opts := s.availableBuildOptionsForColony(idx)
	if len(opts) == 0 {
		return
	}
	cur := 0
	for i, o := range opts {
		if o.Name == s.Builds[idx].Name {
			cur = i
			break
		}
	}
	next := opts[(cur+1)%len(opts)]
	s.Builds[idx] = ColonyBuild{Name: next.Name, Progress: 0, Cost: next.Cost}
}

// applyBuildingEffect 對殖民地 i 套用某已完工建築的長期產出效果(每殖民地每種建築只套一次)。
//
// 2026-07-11 忠實化訂正(詳見 docs/tech/colony-buildings.md 逐項頁碼):舊版把手冊「殖民地整體
// 固定加成」的建築(自動化工廠/機器人採礦廠/深層核心礦場/研究實驗室/行星超級電腦/銀河網路
// 中心/水耕農場/地底農場)近似揉進「每工人/科學家/農夫」per-worker 欄位裡湊數——這會讓小殖民地
// 過度受益、大殖民地受益不足。現在 engine.ColonyState 補上 FlatFood/FlatIndustry/FlatResearch/
// IncomeBonusPercent/PopMax(直接疊加)/FlatGrowth/NormalizeGravity,per-worker 與固定值分開
// 累加,per-worker 數字也一併訂正回手冊原值(不再為了湊固定效果而虛增)。
//
// 太空港(手冊:該殖民地所有來源 BC 收入 +50%)舊版誤植為「工業/工人 +1」,現改用
// IncomeBonusPercent,不再動 IndustryPerWorker。
//
// 氣候控制器(每農業人口食物產出 +2)本來就對應到 FoodPerFarmer 這個既有欄位、且數值正確,
// 維持不動。
func (s *GameSession) applyBuildingEffect(i int, name string) {
	if i < 0 || i >= len(s.PlayerColonies) {
		return
	}
	c := &s.PlayerColonies[i]
	switch name {
	case "自動工廠": // Automated Factories p.78:每工業人口 +1 產能 + 殖民地整體固定 +5 產能。
		// 舊版 IndustryPerWorker+=2 是「per-worker 值虛增以湊固定效果」的近似,訂正回手冊 +1。
		c.IndustryPerWorker += 1
		c.FlatIndustry += 5
	case "研究實驗室": // 現行 remake 只接 sub_DFF74 固定 +5；原版 DFDC6 每科學家 +1 待 READY spec。
		c.FlatResearch += 5
	case "太空港": // Spaceport p.79:該殖民地所有來源 BC 收入 +50%(手冊原文,不是工業加成)。
		c.IncomeBonusPercent += 50
	case "機器人採礦廠": // Robo Mining Plant p.80:每工業人口 +2 產能(既有值與手冊相符,不動) + 固定 +10 產能。
		c.IndustryPerWorker += 2
		c.FlatIndustry += 10
	case "深層核心礦場": // Deep Core Mine p.82:每工人 +3 產能(既有值與手冊相符,不動) + 固定 +15 產能。
		c.IndustryPerWorker += 3
		c.FlatIndustry += 15
	case "污染處理器": // Pollution Processor:對應 engine.ColonyState.PollutionProcessor 既有旗標
		c.PollutionProcessor = true
	case "大氣更新器": // Atmospheric Renewer:對應 engine.ColonyState.AtmosphericRenewer 既有旗標
		c.AtmosphericRenewer = true
	case "核心廢料場": // Core Waste Dumps:完全消除污染,對應 engine.ColonyState.CoreWasteDump 既有旗標
		c.CoreWasteDump = true
	case "行星超級電腦": // 現行 remake 只接 sub_DFF74 固定 +10；原版 DFDC6 每科學家 +2 待 READY spec。
		c.FlatResearch += 10
	case "銀河網路中心": // 現行 remake 只接 sub_DFF74 固定 +15；原版 DFDC6 每科學家 +3 待 READY spec。
		c.FlatResearch += 15
	case "水耕農場": // Hydroponic Farm p.99:殖民地食物整體固定 +2(手冊只有固定值,無 per-farmer 敘述)。
		// 舊版誤建模成 FoodPerFarmer+=1(每農夫 +1),訂正為純固定值、不再動 FoodPerFarmer。
		c.FlatFood += 2
	case "地底農場": // Subterranean Farms p.100:星球食物整體固定 +4(手冊只有固定值,無 per-farmer 敘述)。
		// 舊版誤建模成 FoodPerFarmer+=2,訂正為純固定值、不再動 FoodPerFarmer。
		c.FlatFood += 4
	case "氣候控制器": // Weather Controller p.100:每農業人口食物產出 +2(既有值正確,勿動)。
		c.FoodPerFarmer += 2
	case "行星證券交易所": // Planetary Stock Exchange p.93:該殖民地收入 +100%,與太空港同款累加。
		c.IncomeBonusPercent += 100
	case "太空大學": // Astro University p.93:每受教育人口(農/工/科)額外 +1 對應產出,per-worker 直接建模。
		c.FoodPerFarmer += 1
		c.IndustryPerWorker += 1
		c.ResearchPerScientist += 1
	case "生態圈": // Biospheres p.99:星球人口上限 +2 單位,直接疊加到 PopMax(見該欄位註解:
		// 不另立 PopMaxBonus 影子欄位,PopMax 本身就是成長/人口上限的唯一讀取點)。
		c.PopMax += 2
	case "複製中心": // Cloning Center p.99:人口成長 +0.1 單位/回合,直到達人口上限為止。
		// 官方 patch 1.50 手冊把標準值 100 明列為 +100k；一個人口單位為 1,000 點。
		// colonyGrowth 已依 Population<PopMax 判斷是否還要套用固定成長。
		c.FlatGrowth += gamedata.CloningCenterGrowthPoints
	case "自動實驗室": // Autolab p.96:「generating 30 research points per turn」——固定 30,不依賴人口。
		// 手冊那一句只有一個數字、沒有 per-scientist 敘述,所以只動 FlatResearch。
		c.FlatResearch += 30
	case "食物複製機": // Food Replicators p.85:饑荒時用產能 2:1 換食物,每單位再花 1 BC。
		// 「as needed」只補缺口,不換出盈餘——接法與那條規則都在 gamedata/food_replicators.go。
		c.FoodReplicators = true
	case "再生反應爐": // Recyclotron p.81:每單位人口 +1 產能(不分職業),且該產能不計入污染。
		// 不接 FlatIndustry ——那個欄位在污染縮減之前併入 gross,接錯地方會讓這份產能
		// 跟著產生污染,正好與手冊那句相反。接法見 engine.RunColonyTurn 的 recycled。
		c.Recyclotron = true
	case gamedata.BuildingPlanetaryRadiationShield, gamedata.BuildingPlanetaryFluxShield,
		gamedata.BuildingPlanetaryBarrierShield:
		// 三面護盾的共同效果:**Radiated 氣候轉 Barren**(手冊三句一致,見
		// gamedata/planetary_shield.go)。減傷那一半不在這裡——它由 orbital_bombardment.go
		// 在轟炸解算時讀 ColonyBuildings,不經 ColonyState。
		//
		// 走既有的 applyClimateChange(與地形改造同一支)才會連帶調整食物與人口上限;
		// 直接改 Climate 欄位會讓那兩個數字停在舊氣候上。
		if c.Climate == gamedata.RADIATED {
			s.applyClimateChange(i, gamedata.BARREN)
		}
	case "行星重力產生器": // Planetary Gravity Generator p.104:重力正常化,消除 Low-G/Heavy-G 負面效果。
		// 2026-07-11 已接線:engine.ColonyState.NormalizeGravity=true 時,colonyGravityPenaltyPercent
		// (colony.go)強制把重力懲罰歸零,不論 PlanetGravity 是什麼——此旗標現在真的有效,不再是
		// no-op。玩家母星固定 Normal-G(playerHomeworldColony)本來就無懲罰可消,故這棟建築在
		// demo session 目前看不出效果差異;要在 Low-G/Heavy-G 殖民地(例如存檔載入模式)上才看得出。
		c.NormalizeGravity = true
	case "機器人工廠": // Robotic Factory p.82:依礦產豐度固定加成(Ultra Poor+5/Poor+8/Abundant+10/
		// Rich+15/Ultra Rich+20)。
		//
		// 2026-07-11 已接線:engine.ColonyState 新增 MineralRichness 欄位(比照 PlanetGravity
		// 的接線手法,見該欄位註解的零值陷阱說明),獨立保留建立殖民地當下的原始豐度分類——
		// 不再從已經烘進 IndustryPerWorker 的靜態費率事後反推。gamedata.ProdRoboticFactoryBonus
		// (production.go)是既有查表函式(索引與 formulas.go mineralProductionTable 一致),
		// 直接依 c.MineralRichness 查出手冊固定值加進 FlatIndustry。
		// 注意:機器人工廠效果只有固定加成,不動 IndustryPerWorker——避免與建立殖民地當下已經
		// 烘進 IndustryPerWorker 的礦產費率(gamedata.MineralIndustryPerWorker)重複計算同一份
		// 礦產豐度效果。
		c.FlatIndustry += gamedata.ProdRoboticFactoryBonus(int(c.MineralRichness))
		//
		// 2026-07-11 已接線(移出下方 no-op 清單):全息模擬艙、歡樂穹頂、異族管理中心、裝甲營房。
		// 本 case 語句不直接改 MoralePercent——advanceBuilds 完工當下另外呼叫
		// s.recalcColonyMorale(i),該函式(colonyMoralePercent)讀 s.ColonyBuildings[i] 判斷這些
		// 建築是否存在,依手冊常數加總出淨士氣百分點:
		//   - 全息模擬艙 +20%、歡樂穹頂 +30%:確實會改變 MoralePercent,效果可見。
		//   - 裝甲營房:原本純 no-op,現貢獻 hasBarracks(與海軍陸戰隊營同等地位,解除政府
		//     「無 Barracks -20%」懲罰);裝甲營本身「產生裝甲營駐軍」的效果仍未建模(TODO,
		//     海軍陸戰隊營的駐軍生成系統見 ground_invasion.go,裝甲營房尚無對應版本)。
		//   - 異族管理中心:⚠ **這裡原本寫著「士氣計算路徑已預留(colonyMoralePercent 讀取
		//     此建築名)」——那句是假的。** 2026-08-07 查證:那個建築名在整個 repo 裡只出現在
		//     資料表與註解裡,`colonyMoralePercent` 從來沒讀過它。
		//     真正接上的是**同化速率**(第 40 項(同化系統),見 assimilation.go):手冊「1 per 2 turns,
		//     regardless of government」,對統一政體等於十倍速。
		//     後續已接同化、以 UnassimilatedPop 近似的多種族士氣及叛亂機率減半；
		//     2026-08-28 RE 又證實原版多種族判定按 race group、叛亂按 packed prisoner 與多舊主，
		//     現行資料模型仍是待 READY spec 修正的近似，不能再寫成「沒有叛亂系統」。
		// 海軍陸戰隊營本來就有獨立的陸戰隊召兵系統(ground_invasion.go),現在額外貢獻
		// hasBarracks,兩個系統各自獨立生效,互不影響。
		//
		// 2026-08-07(gap report 第 38/38 項)這段清單縮短了六棟:
		// **自動實驗室**(+30 研究點)、**再生反應爐**(每單位人口 +1 產能且不計污染)、
		// **食物複製機**(饑荒時 2:1 換食物 + 1 BC/食物)已在上面的 switch 建模;
		// **行星輻射/通量/屏障護盾**改由 orbital_bombardment.go 在轟炸解算時讀
		// s.ColonyBuildings(gamedata.PlanetaryShieldReduction),不經 ColonyState。
		//
		// 2026-08-07(第 38–40 項)又縮短三棟,而且都不經 ColonyState:
		// **阿提米絲系統網**(artemis.go,艦隊進入敵方星系時結算水雷)、
		// **太空學院**(crew.go,起始艦員等級 +1 與同星系每回合 +1 經驗)、
		// **異族管理中心**(assimilation.go,同化速率固定 2 回合不分政體)。
		//
		// 仍未建模的(飛彈基地、戰機基地、地面砲台、曲速力場干擾器、戰鬥站、星辰要塞、
		// 次元傳送門)**是真的缺子系統**,不是沒接線:艦隊駐防、軌道防禦火力、
		// 格子戰鬥的獨立戰機單位。
		// 這些仍只由 advanceBuilds 記入 s.ColonyBuildings 為「已建」,顯示於畫面,不影響數值結算。
	}
}

// advanceBuilds 以各殖民地淨工業推進建造;完成則套用建築長期效果、記錄(供回合摘要)並清空。
// 每殖民地每種建築只建/套用一次(ColonyBuildings 去重),重複建造會即時完成但不再疊加效果。
// 「不建造」()與「貿易品」(TradeGoodsBuildName)兩個特殊選項的 Cost 皆固定 0,故下方
// b.Cost == 0 判斷同時排除兩者,不累積建造進度——貿易品該殖民地的淨工業改由
// engine.RunEmpireTurn(依 syncTradeGoodsFlag 同步的 ColonyState.TradeGoods)換算成 BC,
// 不會、也不應該疊加到這裡的建造進度。
func (s *GameSession) advanceBuilds() {
	s.LastBuilt = nil
	s.ensureBuildQueue()
	if s.ColonyBuildings == nil {
		s.ColonyBuildings = make([]map[string]bool, len(s.PlayerColonies))
	}
	for i := range s.Builds {
		b := &s.Builds[i]
		if b.Name == "" {
			continue
		}
		if b.Cost == 0 {
			// 自動建造中的住宅在滿人口時，或貿易品在新設施解鎖後，會在這裡
			// 重新挑選；手動選擇時 AutoBuild 為 false，仍維持原本的持續模式。
			s.refreshAutoBuild(i)
			continue
		}
		// BUY 會在按鈕當下把進度標滿，效果仍在 EndTurn 才完成。這個早退
		// 不依賴 LastPlayerOutput，故即使剛讀檔還沒有上回合產出也不會卡住。
		if colonyBuildComplete(*b) {
			s.completeColonyBuild(i)
			continue
		}
		if i >= len(s.LastPlayerOutput.Colonies) {
			continue
		}
		// 稅金與建造搶同一份生產(手冊 GAME_MANUAL.pdf p.37:「Every rise in the tax rate
		// causes a corresponding drop in production」):稅率抽走 TaxRate% 的淨工業換 BC,剩下
		// (100-TaxRate)% 才用於建造。先前建造吃完整 NetIndustry、稅又另抽一次=稅金變免費錢
		// (非忠實),2026-07-12 校正扣掉稅金那份,使稅率成為真正的「更多錢 vs 更快建造」取捨。
		co := s.LastPlayerOutput.Colonies[i]
		ind := co.NetIndustry * (100 - s.Player.TaxRate) / 100
		complete := false
		if co.Cybernetic {
			// 半機械族每人口消耗半生產力；把該回合剩餘的奇數半單位
			// 累進，避免每回合先除 2 而遺失建造進度。
			indHalf := co.NetIndustryHalf * (100 - s.Player.TaxRate) / 100
			progressHalf := b.Progress*2 + b.ProgressHalf + indHalf
			b.Progress = progressHalf / 2
			b.ProgressHalf = progressHalf % 2
			complete = progressHalf >= b.Cost*2
		} else {
			b.Progress += ind
			b.ProgressHalf = 0
			complete = b.Progress >= b.Cost
		}
		if complete {
			s.completeColonyBuild(i)
		}
	}
}

func colonyBuildComplete(b ColonyBuild) bool {
	return b.Cost > 0 && b.Progress*2+b.ProgressHalf >= b.Cost*2
}

// completeColonyBuild 套用一項已完成的佇列工作，再決定是重複、接下一格或交給
// AUTO BUILD。改裝與一般建築共用這個出口，故 BUY、半機械族半單位與存檔回復
// 都不會各自走出不同的完成規則。
func (s *GameSession) completeColonyBuild(i int) {
	if i < 0 || i >= len(s.Builds) {
		return
	}
	b := s.Builds[i]
	if b.Name == "" {
		return
	}
	if b.Refit != nil {
		s.completeRefitJob(*b.Refit)
		s.LastBuilt = append(s.LastBuilt, BuildNotice{
			Kind: BuildNoticeRefitCompleted, ColonyIndex: i, Name: b.Refit.Target.Name,
		})
	} else {
		if i < len(s.ColonyBuildings) {
			if _, isSpecial := gamedata.SpecialActionByNameZH(b.Name); isSpecial {
				// Special 一次性行動(地形改造/蓋亞轉化/土壤改良/運輸艦隊):刻意不記入
				// ColonyBuildings。手冊明講地形改造可套用好幾次，運輸艦隊同樣可反覆
				// 建造；若記入集合，第二次會被一般建築的去重規則擋掉。
				s.applySpecialAction(i, b.Name)
			} else {
				if s.ColonyBuildings[i] == nil {
					s.ColonyBuildings[i] = make(map[string]bool)
				}
				if !s.ColonyBuildings[i][b.Name] {
					s.ColonyBuildings[i][b.Name] = true
					if b.Name == CapitolBuildName && i < len(s.PlayerColonyPlanets) {
						s.PlayerCapitolPlanet = s.PlayerColonyPlanets[i]
						s.PlayerCapitolPlanetKnown = true
						s.PlayerCapitolRebuildRequired = false
					}
					s.applyBuildingEffect(i, b.Name)
					if b.Name == CapitolBuildName {
						s.recalcAllColonyMorale()
					} else {
						s.recalcColonyMorale(i)
					}
				}
			}
		}
		s.LastBuilt = append(s.LastBuilt, BuildNotice{
			Kind: BuildNoticeCompleted, ColonyIndex: i, Name: b.Name,
		})
	}
	if repeat := s.RepeatBuildFor(i); sameRepeatBuild(repeat, b) {
		s.Builds[i] = ColonyBuild{Name: repeat.Name, Cost: repeat.Cost}
		return
	}
	s.Builds[i] = ColonyBuild{}
	s.popNextBuild(i)
}

// applySpecialAction 對殖民地 i 套用某個已完工的 Special 一次性行動(地形改造/蓋亞轉化/
// 土壤改良,見 gamedata/special_actions.go)。與 applyBuildingEffect 不同:呼叫端(advanceBuilds)
// 刻意不記入 ColonyBuildings,故本函式每次完工都會被呼叫一次,不是「只套一次」。
func (s *GameSession) applySpecialAction(i int, name string) {
	if i < 0 || i >= len(s.PlayerColonies) {
		return
	}
	c := &s.PlayerColonies[i]
	switch name {
	case gamedata.TerraformActionName: // Terraforming p.99-101:把氣候沿階梯往 Terran 方向推進一級。
		targets := gamedata.TerraformNextClimateOptions(c.Climate)
		if len(targets) == 0 {
			// 手冊未定義下一級(已到 Terran/Gaia 終點,或該氣候本來就不能地形改造,如
			// Toxic/Radiated——見 terraform.go terraformNextClimate 註解)。本次套用無效果,
			// PP 已消耗但不改變任何狀態;手冊沒有「退款/擋下建造」的規則,保守不擋。
			return
		}
		// 手冊對 Barren 的下一級給了兩個候選(Desert/Tundra)且未說明選擇條件(見 terraform.go
		// terraformNextClimate 註解),remake 保守固定選第一個候選,不臆造選擇規則。
		s.applyClimateChange(i, targets[0])
	case gamedata.GaiaTransformationActionName: // Gaia Transformation p.99-101:只能套用在 Terran。
		if !gamedata.GaiaTransformationCanApply(c.Climate) {
			return // 非 Terran 星球套用蓋亞轉化,手冊未給效果,保守視為無效果。
		}
		s.applyClimateChange(i, gamedata.GaiaTransformationResultClimate)
	case gamedata.SoilEnrichmentActionName: // Soil Enrichment p.99:每個農夫食物 +1。
		if !gamedata.TerraformSoilEnrichmentWorks(c.Climate) {
			// 手冊:Barren/Radiated/Toxic 星球的化學反應會抵銷肥沃化效果("undo the fertilization
			// as fast as it is done")——誠實模擬「套用了但沒有效果」,不是在建造選單擋下這個選項
			// 本身(手冊沒有明講遊戲介面是否允許排入這種星球的建造佇列,保守不擋選單)。
			return
		}
		c.FoodPerFarmer += gamedata.TerraformSoilEnrichmentFoodBonusPerFarmer
	case gamedata.ArtificialPlanetActionName:
		// 手冊:把同星系的氣態巨星或小行星帶組裝成一顆可殖民的世界(見 artificialplanet.go)。
		// 沒有材料時**誠實地什麼都不發生**——與土壤改良在錯誤氣候上的處理同一個立場:
		// 不在建造選單擋下選項(手冊沒說介面會擋),但套用時不硬塞效果。
		if newPlanet, ok := s.BuildArtificialPlanet(i); ok {
			s.LastBuilt = append(s.LastBuilt, BuildNotice{
				Kind: BuildNoticeArtificialPlanet, ColonyIndex: i, Name: s.Planets[newPlanet].Name,
			})
		}
	case gamedata.FreighterFleetActionName: // Freighter Fleet p.168:每次建成 +5 艘運輸艦 + 版本現金加成(#4)。
		// 帝國整體效果,不是這個殖民地本身的狀態,故不用 c(above 已宣告但本 case 用不到)。
		//
		// 維護費(每艘 0.5 BC/回合)不在這裡處理——ActiveFreighters 一旦變動,下回合
		// engine.RunEmpireTurn(EndTurn 既有呼叫)就會透過 gamedata.IncomeFreighterMaintenanceCost
		// 自動把維護費併入 NetBC,見 engine/empire.go 與該欄位註解,本檔不重複算一次。
		//
		// 現金加成:比照 s.Player.BC += r.StartBC(殖民/擴張既有直接寫 BC 的慣例,見本檔其他呼叫
		// 端),完工當下直接把 RuleProfile.FreightersCashBonus 加進國庫——這是「固定回饋」那一側
		// (見 ruleprofile.go FreightersCashBonus 註解),本 remake 刻意不模擬手冊同段講的「0-3 BC
		// 建造當下維護費立即扣款」那一側(該側本身已被 1.40+ 改成「下回合才扣」,且金額極小、
		// 對整體淨額影響有限,見 MANUAL_150.html Free Cash Bug 表),故 1.3/1.5 呈現的是簡化後的
		// 「淨現金效果方向與量級對」,不是逐 BC 精確重現。
		s.Player.ActiveFreighters += gamedata.FreighterFleetShipsPerBuild
		s.Player.BC += s.RuleProfile.FreightersCashBonus

	case gamedata.ColonyShipActionName: // 殖民船完工 → 進玩家艦隊(手冊 p.85 Frigate 級支援艦)
		// 新艦出現在**生產它的殖民地**;那顆星上剛好有艦隊就併進去,否則進第一支
		// (見 AddShipToHomeFleet ⚠:原版還會依遷移設定自動送往集結點,那要逐殖民地造艦才做得到)。
		action, _ := gamedata.SpecialActionByNameZH(gamedata.ColonyShipActionName)
		s.deliverNewShip(i, Ship{Name: s.nextSupportShipName(ColonyShipClass), Class: ColonyShipClass,
			RawType: gamedata.COLONY_SHIP, RawTypeKnown: true, RawMissionKnown: true, ProductionCost: action.ProductionCost,
			Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"})

	case gamedata.OutpostShipActionName: // 前哨船完工 → 進玩家艦隊(見 outpost.go)
		action, _ := gamedata.SpecialActionByNameZH(gamedata.OutpostShipActionName)
		s.deliverNewShip(i, Ship{Name: s.nextSupportShipName(OutpostShipClass), Class: OutpostShipClass,
			RawType: gamedata.OUTPOST_SHIP, RawTypeKnown: true, RawMissionKnown: true, ProductionCost: action.ProductionCost,
			Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"})
	}
}

// nextSupportShipName 給新造的支援艦一個不重複的名字(如「殖民船 2 號」)。
// 不用隨機艦名池:支援艦在原版的艦隊清單裡也是以用途辨識,編號比隨機名字好認。
func (s *GameSession) nextSupportShipName(class string) string {
	n := 1
	for _, sh := range s.AllShips() { // 全帝國編號,不然分艦隊後會出現兩艘「殖民船 1 號」
		if sh.Class == class {
			n++
		}
	}
	return class + " " + strconv.Itoa(n) + " 號"
}

// applyClimateChange 把殖民地 i 的氣候推進到 next,同步重算 FoodPerFarmer(手冊給的每氣候絕對
// 食物值,前後差值疊加,保留既有建築加成不被覆蓋)與 PopMax(gamedata.TerraformPopMaxAfterClimateChange
// 依 pop_climate 百分比係數等比例縮放,近似值,理由見該函式註解)。
func (s *GameSession) applyClimateChange(i int, next gamedata.PlanetClimate) {
	applyClimateChangeToColony(&s.PlayerColonies[i], next,
		s.raceHasTrait(gamedata.TRAIT_AQUATIC), s.raceHasTrait(gamedata.TRAIT_TOLERANT))
}

func applyClimateChangeToColony(c *engine.ColonyState, next gamedata.PlanetClimate, aquatic, tolerant bool) {
	if c == nil {
		return
	}
	old := c.Climate
	oldFoodClimate := raceFoodClimate(old, aquatic)
	nextFoodClimate := raceFoodClimate(next, aquatic)
	c.FoodPerFarmer += gamedata.ClimateFoodPerFarmer(nextFoodClimate) - gamedata.ClimateFoodPerFarmer(oldFoodClimate)
	oldPopClimate := racePopulationClimate(old, aquatic, tolerant)
	nextPopClimate := racePopulationClimate(next, aquatic, tolerant)
	c.PopMax = gamedata.TerraformPopMaxAfterClimateChange(c.PopMax, oldPopClimate, nextPopClimate)
	c.Climate = next
}

// LeaderSkill 是一位領袖**實際擁有的一項技能**:技能 id + 技能階。
//
// id 用 gamedata 的 LeaderSkills(openorion2 gamestate.h enum),Tier 是
// `Leader::hasSkill` 的回傳值(1 一般 / 2 進階)。
type LeaderSkill struct {
	ID   int `json:"id"`
	Tier int `json:"tier"`
}

// Leader 是一名可雇用的軍官/領袖(供軍官列表)。
type Leader struct {
	// ID 是 HERODATA.LBX／原版 `_leaders[]` 的固定來源序號。0 也是有效 ID；JSON
	// 不省略此欄，避免把第一位領袖誤當成「沒有 ID」。
	ID    int `json:"id"`
	Name  string
	Skill string // 專長的**顯示**標籤(會隨語言翻譯,不可拿來當識別鍵——見 Skills)
	Level int    // 顯示等級(1..5,對照 openorion2 MAX_LEADER_LEVELS=5 顯示慣例:1=最低、5=最高)。
	// 換算成 gamedata.LeaderSkillBonus 用的 expLevel(0..4)時用 leaderDisplayLevelToExpLevel(Level)。
	// 這是 demo 資料的既有欄位語意(非 HERODATA 真實經驗值,是直接指定的顯示等級)。
	Ship bool // true=艦艇軍官,false=殖民地領袖

	// Tier 是 Skill 那一項技能的階(0 無/1 一般/2 進階)。**只在 Skills 為空時才有意義**
	// (demo 領袖與既有測試走這條舊路徑);真英雄的每項技能各自帶自己的階。
	Tier int

	// Skills 是這位領袖的**完整技能清單**,由 HERODATA 的技能位元解出來(2 bit/技能)。
	//
	// ⚠ 這一欄存在的理由是 `Skill` 那個字串**不能當識別鍵**:
	//   - 它會被翻譯(英文模式下 "Scientist" 查不到任何加成,而畫面上看不出異狀);
	//   - 一位真英雄本來就可能同時有好幾項技能,一個字串只放得下一個。
	//
	// 為空時退回舊路徑(用 Skill 標籤反查單一技能),見 leaderSkills。
	Skills []LeaderSkill `json:"skills,omitempty"`

	// RawETA／RawStatus／RawLocation／RawPlayerIndex 是從原版 `.GAM` 59-byte
	// 領袖記錄保留下來的原始欄位。它們不是由 remake 自己推導的「任期」：
	// RawStatus=4 時，原版 `Deassign_Officer @ 0x934CF` 會每回合遞增 raw +0x37，
	// 達 30 後交給 `Check_Officer_Fields @ 0x933F2` 清除；RawStatus=1 則每回合
	// 遞減 raw +0x37，降到 0 且 RawLocation=1 時呼叫 `Colony_Calculation`。
	// 最後一個回呼的完整 raw 重算仍未解出；remake 在 ETA 由 1→0 且 location=1
	// 時，以 `applyLeaderETACallback` 重整已指派殖民地的衍生欄位／士氣，保留
	// 領袖與任職，不宣稱這是原版所有 colony raw 欄位的逐值重建。
	// RawExperience 是原版記錄 +0x24 的 u16。只有 GAM 匯入時才標記已知；舊
	// JSON／demo 沒有這個欄位時，外交特殊貿易只使用顯示等級作保守 fallback。
	RawExperience      int  `json:"rawExperience,omitempty"`
	RawExperienceKnown bool `json:"rawExperienceKnown,omitempty"`
	RawETA             int  `json:"rawEta,omitempty"`
	RawStatus          int  `json:"rawStatus,omitempty"`
	RawLocation        int  `json:"rawLocation,omitempty"`
	RawPlayerIndex     int  `json:"rawPlayerIndex,omitempty"`
}

// demoLeaders 是示範領袖名單(固定;正式版由 HERODATA.LBX 真英雄資料填)。
func demoLeaders() []Leader {
	return []Leader{
		{Name: "馮·諾伊曼", Skill: "科學家", Level: 5, Ship: false, Tier: 1},
		{Name: "洛克斐勒", Skill: "貿易家", Level: 4, Ship: false, Tier: 1},
		{Name: "漢尼拔", Skill: "指揮官", Level: 6, Ship: true, Tier: 1},
		{Name: "圖靈", Skill: "工程師", Level: 3, Ship: true, Tier: 1},
	}
}

// leaderSkills 回傳這位領袖**生效的技能清單**。
//
// 真英雄(HERODATA)帶 Skills;demo 領袖與既有測試只有一個中文 `Skill` 標籤 + `Tier`,
// 退回用 `gamedata.LeaderSkillIDByZH` 反查成單一技能。標籤查不到時回 nil(誠實跳過,
// 不臆造技能)。
func leaderSkills(l Leader) []LeaderSkill {
	if len(l.Skills) > 0 {
		return l.Skills
	}
	id, ok := gamedata.LeaderSkillIDByZH(l.Skill)
	if !ok {
		return nil
	}
	tier := l.Tier
	if tier < 1 {
		tier = 1 // demoLeaders 的既有慣例:標籤有寫就是有這項技能,保守給一般階
	}
	return []LeaderSkill{{ID: id, Tier: tier}}
}

// leaderSkillTier 回傳這位領袖某項技能的階,沒有該技能回 0。
func leaderSkillTier(l Leader, skillID int) int {
	for _, sk := range leaderSkills(l) {
		if sk.ID == skillID {
			return sk.Tier
		}
	}
	return 0
}

// 技能標籤 → id 的對照表已搬到 `gamedata.LeaderSkillIDByZH`(27 個技能全收,名字來自
// GAME_MANUAL.pdf p.135-137)。這裡原本有一張只收 10 項的 `leaderSkillIDByName`,
// 2026-08-08(第 45 項(領袖技能))拿掉,理由有三個:
//
//  1. **標籤不能當識別鍵。** 它會被翻譯——英文模式下 `Skill` 存的是 "Scientist",
//     查表查不到,**所有領袖加成當場全部失效**,而畫面上看不出任何異狀。
//     現在識別鍵是 `Leader.Skills` 裡的技能 id,標籤只負責顯示。
//  2. **「這個技能存在」與「remake 有沒有接」是兩件事。** 舊表用「收不收進表裡」
//     兼表兩者,於是那段註解要一直維護一份「沒收的技能與理由」清單——而它已經過期了
//     (農業官/勞工官/科學官在第 45 項(領袖技能)就接上了,註解卻還寫著沒收)。
//     現在前者查 gamedata 的全表,後者看 `applyLeaderColonyBonuses` 的 switch 有沒有 case。
//  3. 一位真英雄本來就可能同時有好幾項技能,一個字串只放得下一個。
//
// 各技能落在哪個欄位、單位是點數還是百分比,見 `gamedata/leader_skill_apply.go` 的格式字串
// 對照表與下面 switch 的逐條註解。

// leaderDisplayLevelToExpLevel 把 Leader.Level(demo 資料的 1..5 顯示等級)換算成
// gamedata.LeaderSkillBonus 用的 expLevel(0..4)。openorion2 Leader::rank()把
// expLevel(0..4)直接轉成 5 種官階顯示字串,顯示慣例是「數字愈大階級愈高」,故這裡採
// Level-1 對應 expLevel、並夾在 [0,4](demoLeaders 目前有 Level=6 的示範值,略高於官方
// MAX_LEADER_LEVELS=5 上限,夾在 4 是保守處理,不是新規則)。
func leaderDisplayLevelToExpLevel(level int) int {
	exp := level - 1
	if exp < 0 {
		return 0
	}
	if exp > 4 {
		return 4
	}
	return exp
}

// applyLeaderColonyBonuses 把殖民地領袖(Ship=false)的技能加成套到指定殖民地(demo 只有母星,
// 呼叫端傳 &session.PlayerColonies[0])。
//
// **哪些技能真的生效,看下面 switch 有沒有對應的 case**——沒有 case 的技能一律略過,
// 不臆造欄位。目前殖民地欄位接 9 項(科學家/貿易家/財務官/心靈導師/醫官/農業官/勞工官/科學官/環保官);
// 教官、工程師、指揮官、領航員，以及其餘 captain/common 技能落在其他消費端，見下方註解與
// `leader_effects.go`。
//
// ⚠ 2026-08-07(第 45 項(領袖技能))修掉一個一直在的錯:**加成不是每個領袖都疊一份**。
//
// 手冊 p.137「Applicability」:「The effects of the **Megawealth and Researcher** abilities
// are **cumulative**, but **the rest are not** … the leader with the **best applicable
// bonus**」。remake 先前是無條件 `+=`——兩個貿易家就加兩份,而原版只算最強的那一個。
// 合成規則收在 `gamedata.LeaderSkillCombine`。
//
// ⚠ 2026-08-08(第 45 項(領袖技能))改成逐**技能**跑而不是逐領袖跑一次:一位真英雄可能同時有
// 好幾項技能(HERODATA 的技能欄位是每技能 2 bit 的 tier,不是一個人一項技能)。
func applyLeaderColonyBonuses(leaders []Leader, colony *engine.ColonyState) {
	// 先依技能分組收集,再依「累加 vs 取最佳」合成——不能邊走邊加。
	bySkill := map[int][]int{}
	for _, l := range leaders {
		if l.Ship {
			continue // 艦艇軍官不影響殖民地(它們的技能落在戰鬥/航行,見 repair.go、starlane.go)
		}
		expLevel := leaderDisplayLevelToExpLevel(l.Level)
		for _, sk := range leaderSkills(l) {
			if b := gamedata.LeaderSkillBonus(sk.ID, sk.Tier, expLevel); b != 0 {
				bySkill[sk.ID] = append(bySkill[sk.ID], b)
			}
		}
	}
	for id, list := range bySkill {
		bonus := gamedata.LeaderSkillCombine(id, list)
		switch id {
		case int(gamedata.SKILL_RESEARCHER):
			colony.FlatResearch += bonus // 固定點數(格式 "%+d")
		case int(gamedata.SKILL_TRADER):
			colony.IncomeBonusPercent += bonus // 百分比(格式 "%+d%%")
		case int(gamedata.SKILL_FINANCIAL_LEADER):
			colony.IncomeBonusPercent += bonus // 百分比,與貿易家/太空港同一欄位可疊
		case int(gamedata.SKILL_SPIRITUAL_LEADER):
			colony.MoralePercent += bonus // 百分點,與建築/政府士氣同一把尺
		case int(gamedata.SKILL_MEDICINE):
			colony.GrowthBonusSum += bonus // 成長百分點,與種族/科技加成同一把尺
		case int(gamedata.SKILL_FARMING_LEADER):
			colony.FoodBonusPercent += bonus // 分項百分比(2026-08-07 第 45 項(領袖技能)加的欄位)
		case int(gamedata.SKILL_LABOR_LEADER):
			colony.IndustryBonusPercent += bonus
		case int(gamedata.SKILL_SCIENCE_LEADER):
			colony.ResearchBonusPercent += bonus
		case int(gamedata.SKILL_ENVIRONMENTALIST):
			// 環保官的 skillBonus 是**負值**(base −10,降低會致污染的產能),
			// 而 ColonyState 那一欄存的是正的「減幅」——負負得正,所以是 `-=`。
			// 這樣消費端讀起來就是「(100 − 減幅)」,公式裡不必再處理負號。
			colony.PollutionReductionPercent -= bonus
		// 落在別的系統、不是殖民地欄位的技能(所以這裡沒有 case,不是漏掉):
		//   SKILL_INSTRUCTOR → 艦員每回合經驗(帝國層,leaderInstructorXPBonus / crew.go)
		//   SKILL_ENGINEER   → 戰後完全修復(repair.go engineerLeaderTier)
		//   SKILL_COMMANDO   → 地面戰 force(ground_invasion.go commandoLeaderTier)
		//   SKILL_NAVIGATOR  → 艦隊航速與星雲/黑洞豁免(starlane.go FleetHasNavigator)
		default:
		}
	}
}

// leaderInstructorXPBonus 回傳教官技能給的**每回合額外艦員經驗**(手冊 p.137
// 「Boosts the number of experience points earned each turn by all ship crews in your
// empire」;單位是固定點數而非百分比,見 gamedata/leader_skill_apply.go 的格式字串對照)。
//
// 這是**帝國層**的技能(手冊原文 "in your empire"),所以不分艦隊、不分殖民地。
// 依手冊的 Applicability,教官不是累加型 → 多位教官取最強的那一位。
func leaderInstructorXPBonus(leaders []Leader) int {
	var list []int
	for _, l := range leaders {
		tier := leaderSkillTier(l, int(gamedata.SKILL_INSTRUCTOR))
		if tier <= 0 {
			continue
		}
		if b := gamedata.LeaderSkillBonus(int(gamedata.SKILL_INSTRUCTOR), tier,
			leaderDisplayLevelToExpLevel(l.Level)); b != 0 {
			list = append(list, b)
		}
	}
	return gamedata.LeaderSkillCombine(int(gamedata.SKILL_INSTRUCTOR), list)
}

// Planet 是一顆行星的資料(供行星列表與拓殖)。
//
// 字串欄位是顯示用;`*ID` 欄位是原版骰表產生的真值。兩者並存的理由是相容既有存檔:
// 2026-08-06 之前的存檔只有字串,載入時由 restorePlanetIDs 從字串回填(見 persist.go)。
// 新程式碼一律讀 `*ID`,不要再從顯示字串反解語意。
type Planet struct {
	Name    string // 星名 + 羅馬數字
	Climate string
	Gravity string
	Mineral string
	Size    string

	// Gen 是生成器版本:0 = 2026-08-06 之前的舊存檔(只有字串,ID 欄位無意義),
	// 1 = 原版骰表生成(gamedata/galaxygen.go)。用顯式版本號而不是「ID 是否為零值」判斷,
	// 因為 TOXIC/LOW_G/ULTRA_POOR/TINY 的 enum 值都恰好是 0,零值無法區分「未設」與「真的是它」。
	Gen       int
	ClimateID gamedata.PlanetClimate
	GravityID gamedata.PlanetGravity
	MineralID gamedata.PlanetMinerals
	SizeID    gamedata.PlanetSize
	Orbit     int  // 該行星所在軌道(0..4),決定溫度帶進而決定氣候
	NoPlanet  bool // 該恆星沒有行星(黑洞;原版 Generate_Number_Of_Satellites_ 回 0)
	// SpecialID 是行星特殊物產(金礦/寶石礦/原住民…),依原版權重表生成,
	// 見 gamedata/planet_special.go。零值 = 無(原版 64% 的行星都是無)。
	SpecialID gamedata.PlanetSpecial
	// SpecialSeen 表示「抵達星系時的一次性發現」已經結算過(見 discovery.go)。
	// 原版是把結算後的 Star.special 覆寫成訊息碼來達成同樣的「只觸發一次」;remake 沒有
	// Star.special,改用這個旗標,語意一致。
	SpecialSeen bool

	// TypeID 是這顆代表行星的類別(原版 `Generate_Satellite_Type_` + `_orbit_to_satellite_type`,
	// 見 gamedata.RollSatelliteType)。氣態巨星/小行星帶不能直接殖民——手冊 p.61 只允許
	// 「solid planet」建殖民地,那兩類要蓋前哨站。Gen < 2 的舊存檔沒有這個欄位,
	// restorePlanetIDs 一律回填 HABITABLE(舊生成器本來就只產一般行星)。
	TypeID gamedata.PlanetType

	// SystemBodies 是**已淘汰的欄位**(2026-08-07,第 24 項(軌道資料層))。
	//
	// 它原本是一星一行星時代的折衷:代表行星只有一顆,其餘天體記在這裡供顯示用。
	// 而它自己的註解就在擔心「兩份資料要同步的老問題」——現在同系天體是**真正的
	// `Planet` 條目**,各佔一條軌道(見 orbit.go),摘要從軌道表算,只有一份資料。
	//
	// **新程式碼不要用它,產生器也不再填它。** 留著只為了讀得回舊存檔的顯示
	// (舊檔的 SystemBodies 仍在 JSON 裡)。
	SystemBodies []SystemBody
}

// SystemBody 是恆星系裡的一個非代表天體(見 Planet.SystemBodies)。
// 只帶顯示與前哨站判斷需要的最小資訊,不是完整的 Planet——它沒有殖民地、沒有特殊物產結算。
type SystemBody struct {
	Orbit  int                 // 軌道 0..4(由內而外)
	Type   gamedata.PlanetType // 小行星帶 / 氣態巨星 / 一般行星
	Name   string              // 顯示名(恆星名 + 羅馬數字)
	SizeID gamedata.PlanetSize // 一般行星才有意義;氣態巨星/小行星帶留零值
}

// planetGenVersion 是目前生成器版本,寫進 Planet.Gen。
//
//	1 = 原版骰表(光譜/大小/氣候/礦產/重力/行星數)
//	2 = 再加上行星類別(氣態巨星/小行星帶,`_orbit_to_satellite_type`)與同系其他天體
const planetGenVersion = 2

// galaxyAgeSetting 是星系年齡的**預設值**(沒設定時用)。
//
// 2026-08-07:這裡原本是「remake 的新遊戲流程還沒有這個選項,先固定 Average」——已不成立。
// NEW GAME 畫面的 GALAXY AGE 欄接上了(反組譯確認原版就有這個欄位,見 shell.GalaxyAges),
// 實際值改讀 `GameSession.GalaxyAge`,本常數只當零值時的退路。
const galaxyAgeSetting = gamedata.GalaxyAverage

// galaxyAge 回傳這一局的星系年齡。未設定(nil GalaxyAgeSet)時用預設值。
func (s *GameSession) galaxyAge() gamedata.GalaxyAge {
	if !s.GalaxyAgeSet {
		return galaxyAgeSetting
	}
	return s.GalaxyAge
}

// demoHomeStarSet 把「玩家母星 + AI 母星」的星索引收成集合,供 genPlanets 強制生成宜居行星。
func demoHomeStarSet(aiHomeStars []int) map[int]bool {
	m := map[int]bool{0: true} // 星 0 恆為玩家母星,見 PlayerColonyStars 欄位註解
	for _, idx := range aiHomeStars {
		m[idx] = true
	}
	return m
}

// genPlanets 依原版骰表生成每顆恆星的代表行星。
//
// 生成鏈與原版一致(見 gamedata/galaxygen.go 的函式對照):
// 恆星光譜 → 該恆星的行星數(黑洞為 0)→ 隨機取一條軌道 → [光譜][軌道] 得溫度帶
// → 溫度帶加權骰出氣候;行星大小/礦產各自骰表,重力由「礦產(密度)× 大小(體積)」查表。
//
// ⚠ 已知簡化:原版每顆恆星有 1–5 顆行星各佔一條軌道,remake 目前是「一星一行星」
// (s.Planets 與 s.Stars 索引一一對應,UI/拓殖/AI 全部依賴這個對齊)。這裡的做法是
// **在該恆星的軌道中隨機取一條**當代表行星,讓氣候的邊際分布與原版一致;
// 「一星多行星」是獨立的一項改造,見 docs/re/01-gap-report.md。
//
// homeStars 內的索引(玩家與 AI 母星)強制生成宜居行星:原版母星恆為可農作世界,
// 交給機率骰有機會生出 Toxic 母星。
func genPlanets(stars []Star, r, bodyRand *rand.Rand, age gamedata.GalaxyAge, homeStars map[int]bool) []Planet {
	roman := []string{"I", "II", "III", "IV", "V"}
	out := make([]Planet, 0, len(stars))
	// 軌道表(見 orbit.go):每顆星 5 個軌道,預設全空。
	//
	// **同系的每一個天體都是真正的 `Planet` 條目**,各佔一條軌道——`Planets` 因此
	// 不再與 `Stars` 平行(所有呼叫端已於第 24 項(軌道資料層)改走存取器)。
	//
	// ⚠ 非代表天體用**獨立的亂數流** `bodyRand`,理由與 genPlanets/genMonsters/genWormholes
	// 各自一條流完全一樣:多骰幾次不能讓**後面每一顆星**的代表行星跟著漂掉。
	// 沒有這條隔離,同一個 seed 的星系會整個換掉。
	for i := range stars {
		stars[i].Orbits = emptyOrbits()
	}
	setOrbit := func(star, orbit, planet int) {
		if orbit < 0 || orbit >= StarOrbits {
			orbit = 0
		}
		stars[star].Orbits[orbit] = planet
	}
	for i, s := range stars {
		sc := gamedata.SpectralClass(s.Spectral)
		nPlanets := gamedata.RollNumSatellites(r.Intn(10)+1, sc)
		p := Planet{Gen: planetGenVersion}

		if homeStars[i] {
			p.TypeID = gamedata.HABITABLE
			// 母星不骰:直接給與 playerHomeworldColony 一致的行星資料。
			//
			// 為什麼要特判:母星的殖民地狀態(playerHomeworldColony,硬編 Terran/Medium/Abundant,
			// 依 archive.org 原版實測 Sol III)與星圖上那顆星的 Planets 資料是**兩份獨立資料**。
			// 交給骰表決定會出現「殖民地總覽說類地、殖民地畫面說凍原」這種自打嘴巴
			// (2026-08-06 實機截圖抓到)。原版母星本來就是固定的,不是生成出來的。
			// AI 母星比照辦理——AI 殖民地同樣以宜居母星起家。
			p.Orbit = 2 // 宜居帶(見 gamedata.ClassToGroup:黃星第 3 軌道 = 溫度帶 2)
			p.ClimateID = gamedata.TERRAN
			p.SizeID = gamedata.MEDIUM_PLANET
			p.MineralID = gamedata.ABUNDANT
			p.GravityID = gamedata.NORMAL_G
			p.Name = s.Name + " " + roman[p.Orbit]
			p.Climate = climateDisplayName(p.ClimateID)
			p.Gravity = gravityDisplayName(p.GravityID)
			p.Mineral = mineralDisplayName(p.MineralID)
			p.Size = sizeDisplayName(p.SizeID)
			setOrbit(i, p.Orbit, len(out))
			out = append(out, p)
			continue
		}

		if nPlanets <= 0 {
			p.Name = s.Name
			p.NoPlanet = true
			p.Climate, p.Gravity, p.Mineral, p.Size = "無行星", "—", "—", "—"
			// ⚠ 「沒有行星」的星仍然在軌道 0 掛一筆 NoPlanet 條目。
			// 那是**現況的忠實表示**:remake 的 Planets 目前與 Stars 平行,每顆星都有一筆,
			// 靠 NoPlanet 旗標區分。等 SystemBodies 升格之後,這種星應該是「五個軌道全空」。
			setOrbit(i, 0, len(out))
			out = append(out, p)
			continue
		}

		// 依原版 `_orbit_to_satellite_type` 骰出這個恆星系每條軌道上的天體類別,
		// 再挑一顆「代表行星」:優先一般行星(可殖民),整組都不宜居時才退而取第一個天體。
		// 這樣每顆星的**可殖民與否**與原版一致(原版一個系統只要有一顆一般行星就能殖民),
		// 而不是把單一天體的類別直接當成整顆星的命運。
		if nPlanets > 5 {
			nPlanets = 5
		}
		types := make([]gamedata.PlanetType, nPlanets)
		for o := 0; o < nPlanets; o++ {
			types[o], _ = gamedata.RollSatelliteType(r.Intn(10)+1, o, r.Intn(100)+1)
		}
		orbit := 0
		for o := 0; o < nPlanets; o++ {
			if types[o] == gamedata.HABITABLE {
				orbit = o
				break
			}
		}
		group := gamedata.PlanetOrbitGroup(sc, orbit)

		p.TypeID = types[orbit]
		p.Orbit = orbit
		p.SizeID = gamedata.RollPlanetSize(r.Intn(10) + 1)
		p.MineralID = gamedata.RollMineralClass(r.Intn(10)+1, sc)
		p.GravityID = gamedata.PlanetGravityFor(p.MineralID, p.SizeID)
		p.ClimateID = gamedata.RollClimate(r, group, age, homeStars[i])
		// 特殊物產(原版 _planet_special_weighted_chance:64% 無、寶石礦只有 2%)。
		p.SpecialID = gamedata.RollPlanetSpecial(r)

		p.Name = s.Name + " " + roman[orbit]
		p.Climate = climateDisplayName(p.ClimateID)
		p.Gravity = gravityDisplayName(p.GravityID)
		p.Mineral = mineralDisplayName(p.MineralID)
		p.Size = sizeDisplayName(p.SizeID)
		if p.TypeID != gamedata.HABITABLE {
			// 氣態巨星/小行星帶沒有氣候與農業,顯示字串換掉——沿用一般行星的「凍原/海洋」
			// 會讓玩家以為那是可以殖民的星。礦產/重力/大小仍有意義(前哨站升級後會用到)。
			p.Climate = planetTypeDisplayName(p.TypeID)
		}
		setOrbit(i, p.Orbit, len(out))
		out = append(out, p)
		// 同系其他天體:**每一個都是完整的行星條目**,各佔一條軌道。
		// 用 bodyRand(獨立亂數流),不影響代表行星與後續星系的骰序。
		for o := 0; o < nPlanets; o++ {
			if o == orbit {
				continue
			}
			b := Planet{Gen: planetGenVersion, TypeID: types[o], Orbit: o,
				Name: s.Name + " " + roman[o]}
			bg := gamedata.PlanetOrbitGroup(sc, o)
			b.SizeID = gamedata.RollPlanetSize(bodyRand.Intn(10) + 1)
			b.MineralID = gamedata.RollMineralClass(bodyRand.Intn(10)+1, sc)
			b.GravityID = gamedata.PlanetGravityFor(b.MineralID, b.SizeID)
			b.ClimateID = gamedata.RollClimate(bodyRand, bg, age, false)
			b.SpecialID = gamedata.RollPlanetSpecial(bodyRand)
			b.Climate = climateDisplayName(b.ClimateID)
			b.Gravity = gravityDisplayName(b.GravityID)
			b.Mineral = mineralDisplayName(b.MineralID)
			b.Size = sizeDisplayName(b.SizeID)
			if b.TypeID != gamedata.HABITABLE {
				// 氣態巨星/小行星帶沒有氣候與農業(同代表行星那一段的理由)。
				b.Climate = planetTypeDisplayName(b.TypeID)
			}
			setOrbit(i, o, len(out))
			out = append(out, b)
		}
	}
	return out
}

// SystemCompositionText 回傳「同一恆星系裡除了代表行星以外還有什麼」的摘要字串
// (如「同系:氣態巨星×2、小行星帶」);沒有其他天體回空字串。供星系資訊面板顯示。
// ⚠ **已淘汰**:這一支看的是 `Planet.SystemBodies`,而那個欄位自 2026-08-07(第 24 項(軌道資料層))
// 起不再被填——同系天體現在是真正的 `Planet` 條目,掛在軌道表上。
// 新程式碼請用 `GameSession.SystemCompositionText(星)`。
// 這一支留著只為了讀得回舊存檔的顯示(舊檔的 SystemBodies 仍在 JSON 裡)。
func (p Planet) SystemCompositionText() string {
	if len(p.SystemBodies) == 0 {
		return ""
	}
	order := []gamedata.PlanetType{gamedata.HABITABLE, gamedata.GAS_GIANT, gamedata.ASTEROIDS}
	n := map[gamedata.PlanetType]int{}
	for _, b := range p.SystemBodies {
		n[b.Type]++
	}
	out := ""
	for _, t := range order {
		if n[t] == 0 {
			continue
		}
		if out != "" {
			out += "、"
		}
		out += planetTypeDisplayName(t)
		if n[t] > 1 {
			out += "×" + strconv.Itoa(n[t])
		}
	}
	if out == "" {
		return ""
	}
	return "同系:" + out
}

// SystemBodyCountText 回傳「同系還有幾個天體」的極短字串(如「另有 3 天體」),
// 供欄位很窄的行星列表用;沒有其他天體回空字串。完整組成用 SystemCompositionText。
func (p Planet) SystemBodyCountText() string {
	if len(p.SystemBodies) == 0 {
		return ""
	}
	return "另有 " + strconv.Itoa(len(p.SystemBodies)) + " 天體"
}

// planetTypeDisplayName 回傳行星類別的中文顯示名。
func planetTypeDisplayName(t gamedata.PlanetType) string {
	switch t {
	case gamedata.ASTEROIDS:
		return "小行星帶"
	case gamedata.GAS_GIANT:
		return "氣態巨星"
	case gamedata.HABITABLE:
		return "一般行星"
	}
	return "未知"
}

// Race 是可選種族(名稱 + 起始加成)。加成對齊 MOO2 各族招牌特性(remake 調校值,非自訂點數精算):
// 工業/研究/食物為每單位產出加成、GrowthPct 為人口成長百分點、StartBC 為額外起始國庫、
// IncomePerPop 為每人口每回合額外 BC(種族「錢」特質)、CombatPct 為戰鬥戰力百分點。
// Desc 為特性摘要(供顯示)。
//
// ⚠ 2026-07-12 手冊考據校正(GAME_MANUAL.pdf p.15-16 種族章 + SAVE10.GAM):**原版沒有任何
// 種族靠「一次性起始國庫」取得優勢**——五個種族存檔開局 BC 全=50,種族「錢」優勢一律是「每回合
// 按人口的收入加成」。故 StartBC 的種族差異全數移除(人類 60/諾蘭姆 120/達洛克 30 皆為先前捏造),
// 諾蘭姆改用手冊逐字公式 IncomePerPop=2 半BC(=+1 BC/人/回合,「each unit of Gnolam population
// generates an additional 1 BC per turn」);人類真實特質是外交 +50%/易同化/雇用領袖便宜(尚未建模,誠實留白);
// 達洛克是間諜 +20/隱形(對應間諜系統,無錢加成)。StartBC 欄位保留供自訂種族 money pick 用。
type Race struct {
	Name   string // 中文名
	EnName string // 英文名(對應 ai/original.go 種族性格)
	// OrigIdx 是原版種族編號(字母序,也是 RACESEL 肖像順序)。
	//
	// 有了它才查得到 `gamedata.OrigRaceTraits` 那 31 格特性——**布林特性只能從那裡拿**
	// (水棲、地底、食岩、寬容、戰帥、神級商人、高/低重力…),下面這幾個數值欄位裝不下。
	// 自訂種族填 -1。
	OrigIdx      int
	IndBonus     int
	ResBonus     int
	FoodBonus    int
	GrowthPct    int
	StartBC      int
	IncomePerPop int
	// CombatPct 是**艦艇攻擊**加成(原版 TRAIT_SHIP_ATTACK),不是通用戰鬥加成。
	//
	// ⚠ 2026-08-08 訂正:原版把攻擊與防禦分成兩個獨立特性(姆瑞森只有攻 +50、
	// 阿爾卡里只有防 +50、埃雷里安是防 +25 攻 +20),先前 remake 壓成一個
	// 「CombatPct」再填自編值(25/15/15),等於讓阿爾卡里拿到它沒有的攻擊加成。
	CombatPct int
	// ShipDefPct 是艦艇防禦加成(原版 TRAIT_SHIP_DEFENSE)。
	ShipDefPct int
	// GroundCombatBonus 是地面戰加成(原版 TRAIT_GROUND_COMBAT,布拉西 +10)。
	//
	// **定值不是百分比**:反組譯把低重力懲罰寫成 `mov byte ptr [ecx+0Dh], 0F6h`(有號 −10),
	// 與其他地面加成一起加進攻擊力;手冊行文的「10%」是隨手寫法
	// (見 gamedata.GroundLowGCombatPenalty 的完整交代)。
	GroundCombatBonus int
	// SpyBonus 是諜報加成(原版 TRAIT_SPYING,達洛克 +20、薩克拉 −10)。
	// 同樣是定值,與 spyTechBonusFor 那幾項科技加成同一個池子。
	SpyBonus int
}

// Races 是 MOO2 十三經典種族。索引 0 為人類(預設)。
//
// ⚠ **數值全部來自 `gamedata.OrigRaceTraits`(RACESTUF.LBX + 執行檔換算表 + SAVE10 三方核對),
// 不是估計值。** 2026-08-08 之前這裡是七個自編數字,對照下來錯了不少:
//
//	克拉肯   工業+2       → 實為 農業+2、工業+1
//	阿爾卡里 通用戰鬥+15   → 實為 艦艇**防禦**+50(它沒有攻擊加成)
//	姆瑞森   通用戰鬥+25   → 實為 艦艇**攻擊**+50
//	薩克拉   成長+30、食物+1 → 實為 成長+100、農業+2、間諜-10
//	崔拉里安 食物+1、成長+10 → 實為 兩者皆 0(它的特性是水棲與跨維度)
//	埃雷里安 科研+1        → 實為 0(它的加成在艦艇攻防)
//	矽基     工業+1、成長-20 → 實為 工業 0、成長-50
//	達洛克   全 0          → 實為 間諜+20
//
// `race_traits_wiring_test.go` 逐族釘住這張表與 gamedata 一致,改動任一格都會紅。
// 欄位序:名稱 / 英文名 / 原版編號 / 工業 / 科研 / 農業 / 成長% / 起始BC / 每人BC(半單位) /
//
//	艦攻% / 艦防% / 地面戰% / 諜報%。玩家名稱與摘要由外部 JSON 提供。
var Races = []Race{
	{"人類", "Humans", 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{"席隆", "Psilons", 9, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0},
	{"薩克拉", "Sakkra", 10, 0, 0, 2, 100, 0, 0, 0, 0, 0, -10},
	{"克拉肯", "Klackons", 6, 1, 0, 2, 0, 0, 0, 0, 0, 0, 0},
	{"姆瑞森", "Mrrshan", 8, 0, 0, 0, 0, 0, 0, 50, 0, 0, 0},
	{"布拉西", "Bulrathi", 1, 0, 0, 0, 0, 0, 0, 20, 0, 10, 0},
	{"阿爾卡里", "Alkari", 0, 0, 0, 0, 0, 0, 0, 0, 50, 0, 0},
	{"梅克拉", "Meklars", 7, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{"達洛克", "Darloks", 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20},
	{"崔拉里安", "Trilarians", 12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{"埃雷里安", "Elerians", 3, 0, 0, 0, 0, 0, 0, 20, 25, 0, 0},
	{"諾蘭姆", "Gnolams", 4, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0},
	{"矽基", "Silicoids", 11, 0, 0, 0, -50, 0, 0, 0, 0, 0, 0},
}

// ApplyRace 把 Races[idx] 的起始加成套到玩家帝國:各殖民地每單位產出加成、額外起始國庫、
// 記錄成長/戰鬥百分點(供 advancePopulation/戰鬥使用)。只在新遊戲開局套一次。
func (s *GameSession) ApplyRace(idx int) {
	if idx < 0 || idx >= len(Races) {
		return
	}
	r := Races[idx]
	s.RaceIndex = idx
	s.CustomRaceTraits = 0
	s.raceGrowthPct = r.GrowthPct
	s.RaceCombatPct = r.CombatPct
	s.RaceShipDefPct = r.ShipDefPct
	s.RaceGroundBonus = r.GroundCombatBonus
	s.RaceSpyBonus = r.SpyBonus
	s.CustomRaceRuntimeTraits = [gamedata.RaceTraitCount]int8{}
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].IndustryPerWorker += r.IndBonus
		s.PlayerColonies[i].ResearchPerScientist += r.ResBonus
		s.PlayerColonies[i].FoodPerFarmer += r.FoodBonus
		s.PlayerColonies[i].OwnerFoodBonus = r.FoodBonus
		s.PlayerColonies[i].OwnerIndustryBonus = r.IndBonus
		s.PlayerColonies[i].OwnerResearchBonus = r.ResBonus
		s.PlayerColonies[i].OwnerRaceProfileKnown = true
		s.PlayerColonies[i].IncomePerPop += r.IncomePerPop // 種族「錢」特質(諾蘭姆每人+1BC/回合)
	}
	s.applyPlayerHomeworldRaceTraits(r)
	s.Player.BC += r.StartBC
	// 內建種族的政體就是特性 0；先前所有種族都永遠沿用 demo 獨裁。
	if traits, ok := gamedata.OrigRaceTraits(r.OrigIdx); ok {
		gov := gamedata.MoraleGovernmentType(traits[gamedata.TRAIT_GOVERNMENT])
		if choice, ok := governmentChoiceForType(gov); ok {
			s.ApplyGovernment(choice)
		} else {
			s.applyGovernmentType(gov)
		}
	}
	s.finalizeStartingTechForRace()
	s.syncRaceEngineFields()
}

// ApplyCustomRaceBonuses 套用自訂種族(Custom Race)聚合出的數值加成與特殊能力。
// 加成來自 docs/tech/custom-race-picks.md 的官方 patch 1.5 點數值(生產/成長/戰鬥/國庫)。
// traits 保存客製畫面選到的布林能力；目前已有引擎公式的能力會直接由 raceHasTrait 讀取。
// 尚未建模的能力也保留在同一個遮罩中，避免玩家存檔後失去選項語意。
func (s *GameSession) ApplyCustomRaceBonuses(r Race, traits ...gamedata.RaceTrait) {
	// ⚠ 自訂種族不在原版 13 族表上,RaceIndex 必須標成 −1。
	// 不標的話它會停在預設的 0(人類),於是自訂種族會憑空拿到「魅力非凡」
	// ——布林特性是由 RaceIndex 查的(見 raceOrigIdx),而 0 是一個合法索引。
	s.RaceIndex = -1
	s.CustomRaceTraits = 0
	s.CustomRaceRuntimeTraits = [gamedata.RaceTraitCount]int8{}
	for _, t := range traits {
		if t >= gamedata.TRAIT_LOW_G && t <= gamedata.TRAIT_POOR_HOMEWORLD {
			s.CustomRaceTraits |= uint32(1) << uint(t)
			if int(t) < gamedata.RaceTraitCount {
				s.CustomRaceRuntimeTraits[t] = 1
			}
		}
	}
	s.CustomRaceRuntimeTraits[gamedata.TRAIT_POPULATION] = int8(r.GrowthPct)
	s.CustomRaceRuntimeTraits[gamedata.TRAIT_FARMING] = int8(r.FoodBonus)
	s.CustomRaceRuntimeTraits[gamedata.TRAIT_INDUSTRY] = int8(r.IndBonus)
	s.CustomRaceRuntimeTraits[gamedata.TRAIT_SCIENCE] = int8(r.ResBonus)
	s.CustomRaceRuntimeTraits[gamedata.TRAIT_MONEY] = int8(r.IncomePerPop)
	s.CustomRaceRuntimeTraits[gamedata.TRAIT_SHIP_DEFENSE] = int8(r.ShipDefPct)
	s.CustomRaceRuntimeTraits[gamedata.TRAIT_SHIP_ATTACK] = int8(r.CombatPct)
	s.CustomRaceRuntimeTraits[gamedata.TRAIT_GROUND_COMBAT] = int8(r.GroundCombatBonus)
	s.CustomRaceRuntimeTraits[gamedata.TRAIT_SPYING] = int8(r.SpyBonus)
	s.raceGrowthPct = r.GrowthPct
	s.RaceCombatPct = r.CombatPct
	s.RaceShipDefPct = r.ShipDefPct
	s.RaceGroundBonus = r.GroundCombatBonus
	s.RaceSpyBonus = r.SpyBonus
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].IndustryPerWorker += r.IndBonus
		s.PlayerColonies[i].ResearchPerScientist += r.ResBonus
		s.PlayerColonies[i].FoodPerFarmer += r.FoodBonus
		s.PlayerColonies[i].OwnerFoodBonus = r.FoodBonus
		s.PlayerColonies[i].OwnerIndustryBonus = r.IndBonus
		s.PlayerColonies[i].OwnerResearchBonus = r.ResBonus
		s.PlayerColonies[i].OwnerRaceProfileKnown = true
		s.PlayerColonies[i].IncomePerPop += r.IncomePerPop
	}
	s.applyPlayerHomeworldRaceTraits(r)
	s.Player.BC += r.StartBC
	s.syncRaceEngineFields()
}

// applyPlayerHomeworldRaceTraits 套用種族只影響母星的環境能力。
// 呼叫點位於 ApplyRace/ApplyCustomRaceBonuses 的最後,此時 RaceIndex/CustomRaceTraits
// 已完成,而政府乘數尚未套用,正好能用基礎值重新組合後交給 ApplyGovernment。
func (s *GameSession) applyPlayerHomeworldRaceTraits(r Race) {
	if len(s.PlayerColonies) == 0 {
		return
	}
	var planet *Planet
	if len(s.PlayerColonyPlanets) > 0 {
		idx := s.PlayerColonyPlanets[0]
		if idx >= 0 && idx < len(s.Planets) {
			planet = &s.Planets[idx]
		}
	}
	if planet == nil && len(s.Stars) > 0 {
		if idx := s.PlanetAt(0); idx >= 0 && idx < len(s.Planets) {
			planet = &s.Planets[idx]
		}
	}
	applyHomeworldRaceTraits(&s.PlayerColonies[0], planet, r, s.raceHasTrait)
}

// FlagColors 是玩家旗幟顏色選項(原版新遊戲命名畫面選旗色)。
//
// **順序就是原版的旗色索引**(存在玩家結構 +0x26),不能重排——星圖上的艦隊圖示是
// `BUFFER0.LBX 資產 205 + 旗色×4 + 縮放`(見 `Get_Ship_Icon_Pict_Seg_` @ 0xA0D78),
// 順序錯了選紅色會開出白色的艦隊。
//
// ⚠ 2026-08-07 修正:先前是「紅/黃/綠/**藍/白/紫/橙/棕**」,後五個全錯位。
// 兩個獨立來源對上才改的:
//
//	① 把 BUFFER0.LBX 205/209/213/217/221/225/229/233(每組縮放 0 那張)渲染出來量代表色 →
//	   (192,101,96) 紅 /(193,173,28) 黃 /(81,155,61) 綠 /(190,190,198) 銀 /
//	   (143,184,216) 藍 /(198,154,111) 棕 /(196,141,193) 紫 /(223,139,45) 橙
//	② openorion2 `src/gfx.h` 的 `FONT_COLOR_PLAYER_*`:
//	   RED → YELLOW → GREEN → **SILVER** → BLUE → BROWN → PURPLE → ORANGE
//
// 一個量自原版美術、一個抄自別人的重製專案,兩邊逐項相同。
// 原版第 4 色叫 **SILVER**(銀)不是 White——量到的 (190,190,198) 也確實偏灰。
//
// RGB 取自 ① 的量測值(小圖的高光像素,略提亮以便在 UI 上辨識)。
var FlagColors = []struct {
	Key     string
	R, G, B uint8
}{
	{"red", 200, 95, 90},
	{"yellow", 205, 185, 45},
	{"green", 90, 165, 70},
	{"silver", 195, 195, 205},
	{"blue", 145, 185, 220},
	{"brown", 200, 155, 110},
	{"purple", 200, 145, 195},
	{"orange", 225, 140, 45},
}

// Governments 是自訂種族可選的政府型態(順序對應 customrace 政府型態循環選項)。
var Governments = []string{"獨裁", "封建", "統一", "民主"}

// moraleGovByIndex 把 Governments(自訂種族政府循環選項,索引 0-3)映射到
// gamedata.MoraleGovernmentType(士氣查表用的政府 enum,見 internal/gamedata/morale.go)。
//
// MOO2 原版政府其實分基礎型/進階型兩層(Feudalism→Confederation、Dictatorship→Imperium、
// Democracy→Federation、Unification→Galactic Unification),但 Governments 這個 remake 選單
// 只給四個基礎型,故一律映射到對應基礎型,不區分進階版(進階政府的差異——如 Imperium 額外
// +20% 士氣、Command Rating+50%——remake 尚未實作「政府升級」機制,見
// docs/tech/custom-race-picks.md 附錄)。
var moraleGovByIndex = []gamedata.MoraleGovernmentType{
	gamedata.MoraleGovDictatorship, // 0 獨裁
	gamedata.MoraleGovFeudalism,    // 1 封建
	gamedata.MoraleGovUnification,  // 2 統一
	gamedata.MoraleGovDemocracy,    // 3 民主
}

// colonyMoralePercent 依政府基礎值 + 該殖民地已建士氣相關建築,算出淨士氣百分點
// (engine.ColonyState.MoralePercent 的來源；常數已由原版 sub_DDB25 raw 表與手冊交叉驗證)。
// buildings 是該殖民地的 ColonyBuildings 項目(nil 視為尚無任何建築,map 讀取安全)。
//
// 已套用的來源:
//  1. gamedata.MoraleGovernmentBase(gov, hasBarracks)——hasBarracks 依手冊 p.76-79:
//     海軍陸戰隊營(Marine Barracks)或裝甲營房(Armor Barracks)其一即可解除
//     封建/獨裁/統一政府「無 Barracks -20%」的懲罰。
//  2. 全息模擬艙(Holo Simulator)已建 → +gamedata.MoraleHoloSimulatorBonus(+20,p.95-96)。
//  3. 歡樂穹頂(Pleasure Dome)已建 → +gamedata.MoralePleasureDomeBonus(+30,p.97-98)。
//
// 誠實列出「未套用」的手冊來源(不假裝精確,詳見呼叫端 ApplyGovernment/advanceBuilds 註解):
//   - ~~Virtual Reality Network(全帝國 +20%,p.97-98):remake 無「成就」追蹤系統~~
//     ⚠ **2026-08-08(第 59 項(成就科技效果))已接。** 那個理由不成立:「成就」在 MOO2 就是科技,
//     而「有沒有研究出來」一直查得到。與心靈學(+10%,依政體)一起走 achievements.go。
//   - 原版 sub_DDAD4 依 packed population 的不同正常 race slot 判定多種族；remake 尚未保存
//     完整逐人口 race group，目前以 UnassimilatedPop>0 近似。此偏差已登記於
//     docs/re/colony-morale-audit-20260828.md，待 RE gate 後依 READY spec 修正。
//   - 首都淪陷懲罰由 PlayerCapitolRebuildRequired 接入；只有攻陷／移除鏈能設置，
//     Capitol 完工後清除，不從缺少建築鍵值猜測事件。
//
// multiRacial 為 true 時套用手冊的多種族殖民地 20% 士氣懲罰
// (`gamedata.MoraleMultiRacialPenalty`,異族管理中心可消除)。
//
// 2026-08-07 接線:那支函式先前是死碼——remake 沒有「這個殖民地有沒有外族人口」可判斷。
// 第 40 項(同化系統)加上 `ColonyState.UnassimilatedPop` 後可提供玩家可見近似，但 IDA 已證實
// 「未同化人口 > 0」不等同原版 multi-racial；同化完成不會自動讓不同 race slot 消失。
func colonyMoralePercent(gov gamedata.MoraleGovernmentType, buildings map[string]bool, multiRacial bool,
	achievementPct int) int {
	hasBarracks := buildings["海軍陸戰隊營"] || buildings["裝甲營房"]
	pct := gamedata.MoraleGovernmentBase(gov, hasBarracks) + achievementPct
	if buildings["全息模擬艙"] {
		pct += gamedata.MoraleHoloSimulatorBonus
	}
	if buildings["歡樂穹頂"] {
		pct += gamedata.MoralePleasureDomeBonus
	}
	if multiRacial {
		pct += gamedata.MoraleMultiRacialPenalty(buildings[alienManagementCenterName])
	}
	return pct
}

func setColonyGovernmentOutput(c *engine.ColonyState, gov gamedata.MoraleGovernmentType) {
	c.GovernmentFoodBonusPercent = gamedata.GovernmentJobProductionBonus(gov, 0)
	c.GovernmentIndustryBonusPercent = gamedata.GovernmentJobProductionBonus(gov, 1)
	c.GovernmentResearchBonusPercent = gamedata.GovernmentJobProductionBonus(gov, 2)
	if gov == gamedata.MoraleGovUnification || gov == gamedata.MoraleGovGalacticUnification {
		c.MoralePercent = 0 // sub_DE280 對統一系跳過 colony+7 一般士氣項。
	}
}

func effectiveAIGovernment(a *AIOpponent) gamedata.MoraleGovernmentType {
	traits, ok := gamedata.OrigRaceTraits(a.RaceIndex)
	if !ok {
		return gamedata.MoraleGovDictatorship
	}
	base := gamedata.MoraleGovernmentType(traits[gamedata.TRAIT_GOVERNMENT])
	advanced := gamedata.Technology(0)
	switch base {
	case gamedata.MoraleGovFeudalism:
		advanced = gamedata.TECH_CONFEDERATION
	case gamedata.MoraleGovDictatorship:
		advanced = gamedata.TECH_IMPERIUM
	case gamedata.MoraleGovDemocracy:
		advanced = gamedata.TECH_FEDERATION
	case gamedata.MoraleGovUnification:
		advanced = gamedata.TECH_GALACTIC_UNIFICATION
	}
	if advanced != 0 && groundEquipTechOwned(a.Player, gamedata.TOPIC_ADVANCED_GOVERNMENTS, advanced) {
		return gamedata.MoraleGovernmentType(gamedata.AssimilationAdvancedForm(gamedata.AssimilationGovernment(base)))
	}
	return base
}

func syncAIColonyGovernmentOutput(a *AIOpponent) {
	gov := effectiveAIGovernment(a)
	for i := range a.Colonies {
		setColonyGovernmentOutput(&a.Colonies[i], gov)
	}
}

// buildingsFor 回傳殖民地 i 已完工建築集合。s.ColonyBuildings 是延遲配置的(見 advanceBuilds
// 註解),索引越界或該殖民地尚無記錄一律視為「尚無建築」(nil map 讀取回傳零值,不 panic)。
func (s *GameSession) buildingsFor(i int) map[string]bool {
	if i < 0 || i >= len(s.ColonyBuildings) {
		return nil
	}
	return s.ColonyBuildings[i]
}

// recalcColonyMorale 依目前政府(s.Government)+ 殖民地 i 已建士氣建築,重算
// PlayerColonies[i].MoralePercent(見 colonyMoralePercent)。呼叫時機:政府變更
// (ApplyGovernment)、建築完工(advanceBuilds)。
func (s *GameSession) recalcColonyMorale(i int) {
	if i < 0 || i >= len(s.PlayerColonies) {
		return
	}
	// ⚠ 用**生效政體**不是 s.Government:手冊的士氣表對基本型與進階型給不同的值
	// (帝國比獨裁多 +20%),傳基本型會讓研究出帝國的獨裁玩家永遠拿獨裁那一格。
	gov := s.effectiveGovernment()
	s.PlayerColonies[i].MoralePercent = colonyMoralePercent(gov, s.buildingsFor(i),
		s.PlayerColonies[i].UnassimilatedPop > 0, achievementMoralePercent(s.Player, gov))
	if s.playerCapitolMissing() {
		s.PlayerColonies[i].MoralePercent += gamedata.MoraleCapitalCapturedPenalty(gov)
	}
	setColonyGovernmentOutput(&s.PlayerColonies[i], gov)
}

// recalcAllColonyMorale 對所有玩家殖民地重算士氣(見 recalcColonyMorale)。
func (s *GameSession) recalcAllColonyMorale() {
	for i := range s.PlayerColonies {
		s.recalcColonyMorale(i)
	}
}

// ApplyGovernment 套用政府型態:①「本 remake 已建模資源」的乘數效果(手冊 p.20–23 明列百分比)
// ②(2026-07-11 接線)記錄選定政府並重算所有殖民地士氣(colonyMoralePercent)。
//   - 封建(1):研究減半。
//   - 統一(2):食物 +50%、產能 +50%。
//   - 民主(3):研究 +50%。
//   - 獨裁(0):基準,無資源乘數。
//
// ⚠ 誠實標註:政府在原版還有征服同化回合、間諜/防禦加成、造艦成本等系統本 remake 尚未建模,
// 故**未模擬**(不自編近似)——但士氣已從「未建模」升級為「已建模」(見 colonyMoralePercent),
// 上面這條舊聲明已不再涵蓋士氣。詳見 docs/tech/custom-race-picks.md 政府效果附錄與缺口說明。
// gov 索引對應 Governments。
func (s *GameSession) ApplyGovernment(gov int) {
	if gov >= 0 && gov < len(moraleGovByIndex) {
		s.applyGovernmentType(moraleGovByIndex[gov])
	}
	s.syncStartingCapitolForGovernment()
	s.recalcAllColonyMorale()
	s.finalizeStartingTechForRace()
}

func (s *GameSession) applyGovernmentType(gov gamedata.MoraleGovernmentType) {
	s.Government = gov
	if s.RaceIndex < 0 {
		s.CustomRaceRuntimeTraits[gamedata.TRAIT_GOVERNMENT] = int8(gov)
	}
}

func governmentChoiceForType(gov gamedata.MoraleGovernmentType) (int, bool) {
	for choice, candidate := range moraleGovByIndex {
		if candidate == gov {
			return choice, true
		}
	}
	return 0, false
}

// GalaxySizes 是星系大小選項(名稱 + 星數),對應 NEW GAME 的 GALAXY SIZE。
var GalaxySizes = []struct {
	Name  string
	Stars int
}{
	// 星數是原版 `Galaxy_Size_From_N_Stars_` @ 0x798D2 的四個門檻(20/36/54/72)——
	// 那支就是原版自己的「星數 → 銀河大小」判定,門檻值即各檔的星數。
	// ⚠ 先前是 12/24/36/48(remake 自訂),與原版對不上;而星雲數、銀河跨距(秒差距)
	// 這些表都是**以檔位為索引**的,對不上就整串偏掉。見 internal/gamedata/starlane.go。
	{"小型", 20}, {"中型", 36}, {"大型", 54}, {"巨型", 72},
}

// RegenGalaxy 依指定星數重生星系(+ 對應行星);舊介面,保留供其餘不需要重建 AI 對手的呼叫端
// 使用。2026-07-11 訂正:cmd/moo2 的新遊戲流程(customrace.go/raceselect.go)先前呼叫的正是
// 本函式,但本函式只重生星系、完全不重建 s.AIPlayers——結果 NewDemoSession 建的 3 個 AI 的
// ColonyStars/Colonies 仍指向舊(demo)星系的星索引,新星系裡卻只有 1 顆星被標成 AI 母星
// (aiHomes 寫死 1),資料與畫面對不上,正式開局形同沒有正確對手。全 repo grep 只有
// customrace.go/raceselect.go 兩處呼叫端(見 SetupNewGame 呼叫),兩者都已改呼叫 SetupNewGame,
// 不再經過本函式;本函式轉呼叫 SetupNewGame(n, seed, 1) 保留「只需 1 AI」語意的相容出口,
// 供將來其他呼叫端(如測試)需要單純重生星系但不在意 AI 正確性時使用。
func (s *GameSession) RegenGalaxy(n int, seed int64) {
	s.SetupNewGame(n, seed, 1)
}

// SetupNewGame 重生星系並依 numAI 重建 AI 對手,取代舊版只重生星系的 RegenGalaxy——正式新遊戲
// 流程(customrace.go/raceselect.go 的 applyAndStart)用本函式開局,確保重生後的星系與
// s.AIPlayers 的 ColonyStars 對得上號(都指向同一份新星系),不再殘留舊 demo 星系的 stale 索引。
//
// 只重建「星系與 AI」,不動玩家的種族加成/政府/殖民地——那些由呼叫端在 SetupNewGame 之後各自
// ApplyRace/ApplyCustomRaceBonuses/ApplyGovernment(順序與現行一致)。玩家母星/起始殖民地
// (PlayerColonies/PlayerColonyStars)已由 NewDemoSession 建好(cmd/moo2 的 sceneBuilder 一律以
// shell.NewDemoSession() 起始 session,見 interactive.go newInteractive),新遊戲流程只是「換一個
// 星系與 AI 陣容」,玩家殖民地本身維持不動(母星固定星 0,見 PlayerColonyStars 欄位註解)。
//
// numAI<=0 時 buildDemoAIOpponents 收到空的 aiHomeStars 會回傳空 slice,退化為無 AI;呼叫端應傳
// >=1。
func (s *GameSession) SetupNewGame(stars int, seed int64, numAI int) {
	s.EventLastTurn = 0
	s.EventAttemptCounter = 0
	s.LuckyEventCounter = 0
	s.eventRand = nil
	galaxy, aiHomeStars := genGalaxy(stars, seed, numAI, s.galaxyAge(), s.nameTranslator())
	galaxy[0].Explored = true // 母星初始已探索(與 NewDemoSession 一致)
	s.Stars = galaxy
	// 行星生成用獨立的亂數流(seed+1),讓「同一 seed 的星圖佈局」不受行星骰表的抽取次數影響——
	// 骰表每顆星抽的次數依光譜而異,共用一條流會讓佈局跟著漂。
	s.Planets = genPlanets(galaxy, rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+5)), s.galaxyAge(), demoHomeStarSet(aiHomeStars))
	// 守衛怪獸也用獨立亂數流(seed+2),理由同上:不讓它的抽取次數影響星圖與行星的骰序。
	// 它會就地修改 s.Planets(手冊 p.60:有怪獸的星系一定另有一個特殊物產),故在 genPlanets 之後。
	s.Monsters = genMonsters(galaxy, s.Planets, rand.New(rand.NewSource(seed+2)), demoHomeStarSet(aiHomeStars))
	// 蟲洞也用獨立亂數流(seed+3),理由同上——不讓它的抽取次數影響星圖/行星/怪獸的骰序。
	// 母星(玩家的星 0 + 各 AI 的)不可當端點,與原版一致(見 wormhole.go)。
	whRand := rand.New(rand.NewSource(seed + 3))
	genWormholes(s.Stars, demoHomeStarSet(aiHomeStars), whRand.Intn)
	// 星雲同樣用獨立亂數流(seed+4),理由同上。銀河大小 → 星雲數的對應表見 nebula.go;
	// remake 的星數不是原版的「小/中/大/巨大」四檔,這裡用星數換算過去。
	s.Nebulae = genNebulae(galaxySizeClass(len(s.Stars)), demoHomeStarSet(aiHomeStars),
		s.Stars, rand.New(rand.NewSource(seed+4)))
	s.SelectedStar = -1
	s.AIPlayers = buildDemoAIOpponents(aiHomeStars, len(s.Stars), s.Difficulty, seed)
	s.syncAIColonyPlanets() // 行星索引要等 Planets 生完才補得起來(見該函式)
	s.PlayerColonyPlanets = []int{s.PlanetAt(0)}
	s.PlayerCapitolPlanetKnown = false
	for i := range s.AIPlayers {
		s.AIPlayers[i].CapitolPlanetKnown = false
	}
	s.ensureCapitolState()
	s.PlayerSpies = make([]int, len(s.AIPlayers))              // 平行 AIPlayers,重置為全新對手的間諜數(開局皆 0)
	s.PlayerSpyMissions = make([]SpyMission, len(s.AIPlayers)) // 零值 STEAL
	s.PlayerColonyStars = []int{0}
	s.Fleet().AtStar = 0
	s.Fleet().DestStar = -1
	s.newGameRacePending = true
	s.applyStartingTech()
	for i := range s.AIPlayers {
		s.ensureAIShipDesigns(i)
	}
}

func resetOpeningResearch(ps *engine.PlayerState) {
	ps.ResearchProgress = 0
	ps.ResearchApplication = 0
	ps.HasResearchApplication = false
	ps.CompletedTopics = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_STARTING_TECH: true}
	ps.HyperAdvancedLevels = nil
	ps.ChosenTech = map[gamedata.ResearchTopic]gamedata.Technology{
		gamedata.TOPIC_ENGINEERING: gamedata.TECH_COLONY_BASE,
	}
	ps.ExplicitChoice = nil
	ps.HasPendingChoice = false
	ps.PendingChoice = 0
}

// finalizeStartingTechForRace 只在正式新局的種族／政府套用完畢時重建開局科技。
// SetupNewGame 的預發供直接 API caller 保持相容；UI 正常路徑會在這裡以真實種族重抽。
func (s *GameSession) finalizeStartingTechForRace() {
	if !s.newGameRacePending {
		return
	}
	s.newGameRacePending = false
	resetOpeningResearch(&s.Player)
	for i := range s.AIPlayers {
		resetOpeningResearch(&s.AIPlayers[i].Player)
	}
	s.applyStartingTech()
	s.applyAdvancedCivilizationStartingBC()
	s.applyAdvancedCivilizationColonies()
	normalizeCurrent := func(ps *engine.PlayerState) {
		if !ps.CompletedTopics[ps.ResearchTopic] {
			return
		}
		if available := gamedata.AvailableTopics(ps.CompletedTopics); len(available) > 0 {
			ps.ResearchTopic = available[0]
			ps.ResearchProgress = 0
		}
	}
	normalizeCurrent(&s.Player)
	for i := range s.AIPlayers {
		normalizeCurrent(&s.AIPlayers[i].Player)
	}
}

// applyAdvancedCivilizationStartingBC 在正式新遊戲的種族特性已確定後套用原版先進文明國庫。
// 一般／曲速前開局維持 newHomeworldPlayerState 的 50 BC；只有 TechLevel=2 會改寫。
// 玩家與 AI 必須在同一時點處理，否則先進開局會讓真人憑空比 AI 少 150 BC。
func (s *GameSession) applyAdvancedCivilizationStartingBC() {
	if s.techLevel() != 2 {
		return
	}
	if traits, ok := s.playerStartingRuntimeTraits(); ok {
		s.Player.BC = gamedata.AdvancedCivilizationStartingBC(int(traits[gamedata.TRAIT_MONEY]))
	}
	for i := range s.AIPlayers {
		origRace := -1
		if raceIdx := aiRaceIndex(s.AIPlayers[i]); raceIdx >= 0 && raceIdx < len(Races) {
			origRace = Races[raceIdx].OrigIdx
		}
		if traits, ok := gamedata.OrigRaceTraits(origRace); ok {
			s.AIPlayers[i].Player.BC = gamedata.AdvancedCivilizationStartingBC(int(traits[gamedata.TRAIT_MONEY]))
		}
	}
}

func (s *GameSession) playerStartingRuntimeTraits() ([gamedata.RaceTraitCount]int8, bool) {
	if idx := s.raceOrigIdx(); idx >= 0 {
		traits, ok := gamedata.OrigRaceTraits(idx)
		if ok {
			traits[gamedata.TRAIT_GOVERNMENT] = int8(s.Government)
		}
		return traits, ok
	}
	traits := s.CustomRaceRuntimeTraits
	traits[gamedata.TRAIT_GOVERNMENT] = int8(s.Government)
	return traits, true
}

// applyStartingTech 依 TECH LEVEL 把開局該給的研究主題發下去(玩家 + 所有 AI)。
//
// 擺在 `SetupNewGame` 的最後而不是 `newHomeworldPlayerState` 裡,是因為 TECH LEVEL 是
// NEW GAME 畫面設的,而 PlayerState 在那之前就建好了(`applyNewGameSettings` → `SetupNewGame`
// 的順序見 cmd/moo2)。在這裡補發,順序上一定拿得到設定,舊存檔也不受影響。
//
// **AI 一起發**:原版的 `Init_Player_Tech_` 是逐玩家跑的(第一個參數就是玩家編號),
// 不是玩家專屬。只發給玩家等於把 AI 永遠留在曲速前。
func (s *GameSession) applyStartingTech() {
	topics := gamedata.StartingTopics(s.techLevel())
	granted := map[gamedata.ResearchTopic]bool{}
	for _, t := range topics {
		granted[t] = true
	}
	grant := func(ps *engine.PlayerState) {
		if ps.CompletedTopics == nil {
			ps.CompletedTopics = map[gamedata.ResearchTopic]bool{}
		}
		// 先把**這張固定表裡、這一級不該有的**清掉,再發該有的。
		//
		// 不是多餘的一步:`NewDemoSession` 會先用預設等級發一輪,測試/正式流程再改成別的等級
		// 重新 setup。只加不減的話,曲速前會留著「一般」發過的核分裂——那正好是
		// 「曲速前不該有 FTL」這條規則的反例,而且不會有任何錯誤訊息。
		//
		// 只動這張表裡的主題:玩家真的研究出來的東西不歸這裡管(setup 當下也還沒有)。
		for _, t := range gamedata.StartingTopicOrder {
			if !granted[t] {
				delete(ps.CompletedTopics, t)
			}
		}
		for _, t := range topics {
			ps.CompletedTopics[t] = true
		}
	}
	grant(&s.Player)
	for i := range s.AIPlayers {
		grant(&s.AIPlayers[i].Player)
	}
	s.applyStartingRandomTech()
	s.applyStartingBuildings()
}

// applyStartingRandomTech 發先進級開局多出來的那 19 個**隨機**主題
// (gap report 第 30 項(TECH LEVEL 第二效果)留下的缺口,結構見 gamedata/starting_random_tech.go)。
//
// 原版 `Init_Player_Tech_` 的主迴圈跑 1 / 6 / 25 次:前 6 次取固定表,
// **第 7 次起改由 `sub_FD335` 隨機挑**。remake 先前只發那六個。
//
// 每挑一個就重算一次候選清單——原版每完成一個主題就解鎖它的後繼,所以這 19 個是
// **沿著樹往上走**而不是從同一池子裡抽 19 次。`gamedata.AvailableTopics` 正好是
// 「每個領域第一個尚未完成的主題」,與原版的狀態 2 同義(見第 37 項(研究樹一手驗證)的雙向驗證)。
//
// 2026-08-25 由 IDA 再確認 `Choose_Tech_Application_ @ 0xFD335` 是直接對 212 個科技應用
// 評分並只做一次加權抽選；不再先抽主題、再用第二顆亂數抽應用。人類共同估值鏈已接；
// raw4（原版 player+0x205；英文符號語意仍不作事實）已由 `sub_589D6` 權重表初始化，
// 並送入真人／AI 共用後段；不再略過該四組分支。
func (s *GameSession) applyStartingRandomTech() {
	extra := gamedata.StartingTopicRandomExtras(s.techLevel())
	if extra <= 0 {
		return
	}
	// 每回合研究點:開局那一刻還沒有回合跑過,取母星的初始研究產出——
	// 那是最接近原版當下狀態的值(原版讀 [player+0xAC])。
	rp := gamedata.ResearchPerScientistNorm
	if len(s.PlayerColonies) > 0 {
		if n := s.PlayerColonies[0].Scientists * s.PlayerColonies[0].ResearchPerScientist; n > 0 {
			rp = n
		}
	}
	knownTechs := func(ps engine.PlayerState) map[gamedata.Technology]bool {
		known := map[gamedata.Technology]bool{}
		for tech, granted := range ps.GrantedTechs {
			if granted {
				known[tech] = true
			}
		}
		for topic := range ps.CompletedTopics {
			choices := gamedata.ResearchChoicesForTopic(topic)
			if len(choices) == 0 {
				continue
			}
			if ps.ExplicitChoice != nil && ps.ExplicitChoice[topic] {
				if tech, ok := ps.ChosenTech[topic]; ok {
					known[tech] = true
				}
				continue
			}
			for _, tech := range choices {
				known[tech] = true
			}
		}
		return known
	}
	grantRandom := func(ps *engine.PlayerState, human bool, origRace int,
		profileTraits *[gamedata.RaceTraitCount]int8, profileOverride *gamedata.OriginalAITechProfile, seed int64,
		opponents func() []map[gamedata.Technology]bool) {
		if ps.CompletedTopics == nil {
			return
		}
		rng := newRandStream(seed)
		var aiProfile gamedata.OriginalAITechProfile
		aiProfileKnown := false
		if profileOverride != nil {
			aiProfile = *profileOverride
			aiProfileKnown = true
		} else if profileTraits != nil {
			raw27 := 0
			if origRace >= 0 {
				raw27 = gamedata.RollOriginalAIRaw27(origRace, s.Difficulty, rng.Intn)
			}
			aiProfile = gamedata.RollOriginalAITechProfile(*profileTraits, s.Difficulty, raw27, rng.Intn)
			aiProfileKnown = true
		}
		for i := 0; i < extra; i++ {
			avail := gamedata.AvailableTopics(ps.CompletedTopics)
			state := gamedata.OriginalStartingValueState{
				Human: human, Difficulty: s.Difficulty, InitialSixKnown: true,
				RelativeTurn: s.Turn,
				AIProfile:    aiProfile, AIProfileKnown: aiProfileKnown,
				Raw4: aiProfile.Raw4, Raw4Known: aiProfileKnown,
				Known: knownTechs(*ps), Opponents: opponents(),
			}
			t, tech, ok := gamedata.StartingOriginalApplicationPick(avail, rp, state, rng.Intn)
			if !ok {
				return // 沒有候選了(整棵樹研究完),不硬塞
			}
			ps.CompletedTopics[t] = true
			// ⚠ 2026-08-08(第 49 項(安塔蘭防禦艦隊))補上**粒度**:原版 `Choose_Tech_Application_`
			// 挑的是一個**科技應用**,不是整個主題(見 gamedata/starting_random_tech.go
			// 檔尾那段的反組譯證據)。只標 CompletedTopics 而不做抉擇,
			// `componentUnlockedFor` 會把那個主題底下的抉擇**全部**解鎖
			// ——先進級開局因此拿到原版的兩到三倍。
			//
			// ResearchAll 主題完成後取得全部應用，不寫 ExplicitChoice；其他主題保存這次
			// 原版應用級抽選的唯一結果。
			if !gamedata.ResearchTopicGrantsAll(t) {
				if ps.ChosenTech == nil {
					ps.ChosenTech = map[gamedata.ResearchTopic]gamedata.Technology{}
				}
				if ps.ExplicitChoice == nil {
					ps.ExplicitChoice = map[gamedata.ResearchTopic]bool{}
				}
				ps.ChosenTech[t] = tech
				ps.ExplicitChoice[t] = true
			}
			applyResearchTopicGrantCallbacks(ps, t)
		}
	}
	playerOpponents := func() []map[gamedata.Technology]bool {
		out := make([]map[gamedata.Technology]bool, 0, len(s.AIPlayers))
		for _, ai := range s.AIPlayers {
			out = append(out, knownTechs(ai.Player))
		}
		return out
	}
	playerTraits, playerTraitsKnown := s.playerStartingRuntimeTraits()
	var playerTraitsPtr *[gamedata.RaceTraitCount]int8
	if playerTraitsKnown {
		playerTraitsPtr = &playerTraits
	}
	grantRandom(&s.Player, true, s.raceOrigIdx(), playerTraitsPtr, nil,
		s.EventSeed*6364136223846793005+11, playerOpponents)
	for i := range s.AIPlayers {
		aiIdx := i
		opponents := func() []map[gamedata.Technology]bool {
			out := []map[gamedata.Technology]bool{knownTechs(s.Player)}
			for j, other := range s.AIPlayers {
				if j != aiIdx {
					out = append(out, knownTechs(other.Player))
				}
			}
			return out
		}
		origRace := -1
		if raceIdx := aiRaceIndex(s.AIPlayers[i]); raceIdx >= 0 && raceIdx < len(Races) {
			origRace = Races[raceIdx].OrigIdx
		}
		var aiTraitsPtr *[gamedata.RaceTraitCount]int8
		if traits, ok := gamedata.OrigRaceTraits(origRace); ok {
			aiTraitsPtr = &traits
		}
		var profileOverride *gamedata.OriginalAITechProfile
		if s.AIPlayers[i].OriginalTechProfileKnown {
			profileOverride = &s.AIPlayers[i].OriginalTechProfile
		}
		grantRandom(&s.AIPlayers[i].Player, false, origRace, aiTraitsPtr, profileOverride,
			s.EventSeed*6364136223846793005+int64(i)*7919+101, opponents)
	}
}

// applyStartingBuildings 依 TECH LEVEL 重發母星的開局建築。
//
// 與 `applyStartingTech` 同一個理由擺在這裡:建築內容取決於「開局知道哪些科技」,
// 而那由 NEW GAME 的 TECH LEVEL 決定、比 PlayerState 建好的時間晚。
// 呼叫點就接在發完科技之後——**順序不能反**,建築要看科技。
//
// 只動母星(索引 0):其他殖民地是玩家自己拓的,不歸開局規則管。
func (s *GameSession) applyStartingBuildings() {
	// 讀**玩家實際知道的主題**而不是再算一次固定表:先進級的 19 個隨機主題
	// (第 43 項(先進級開局主題))也要算進去,否則母星永遠只蓋得出兩棟(見 homeworldBuildingsForKnown)。
	known := s.Player.CompletedTopics
	if known == nil {
		known = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_STARTING_TECH: true}
	}
	b := homeworldBuildingsForKnown(s.techLevel(), homeworldStartPop, known)
	if !isUnifiedGovernment(s.effectiveGovernment()) {
		b[CapitolBuildName] = true
	}
	if len(s.ColonyBuildings) > 0 {
		s.ColonyBuildings[0] = cloneBuildings(b)
	}
	for i := range s.AIPlayers {
		if len(s.AIPlayers[i].ColonyBuildings) > 0 {
			aiBuildings := cloneBuildings(b)
			if isUnifiedGovernment(effectiveAIGovernment(&s.AIPlayers[i])) {
				delete(aiBuildings, CapitolBuildName)
			}
			s.AIPlayers[i].ColonyBuildings[0] = aiBuildings
		}
	}
}

// aiHomeStarIndices 依「星數 n、AI 對手數 aiHomes」算出 aiHomes 個彼此不同、且都不是星 0
// (玩家母星)的星索引,供 genGalaxy 標記 AI 母星用。分佈公式 idx_k = n*k/(aiHomes+1)
// (k=1..aiHomes)把 AI 母星在星圖索引上盡量平均攤開,不擠在同一角落;aiHomes=1 時
// idx_1 = n*1/2 = n/2,與 genGalaxy 舊版「唯一 AI 母星固定在 n/2」逐位元相同,故
// RegenGalaxy(仍只需 1 個 AI 母星索引的呼叫端)行為完全不變,不是新的星系配置。
// 若算出的索引撞到已佔用的(理論上只在 n 遠小於 aiHomes 時發生,目前呼叫端 n>=8),
// 用「向後掃描找下一個未佔用索引、繞回 idx=1 續找」補位,最多嘗試 n 次避免死迴圈——
// 極端小 n 高 aiHomes 下不保證完全不撞,只保證函式一定終止。
func aiHomeStarIndices(n, aiHomes int) []int {
	if aiHomes <= 0 || n <= 1 {
		return nil
	}
	used := map[int]bool{0: true} // 星 0 保留給玩家母星
	out := make([]int, 0, aiHomes)
	for k := 1; k <= aiHomes; k++ {
		idx := n * k / (aiHomes + 1)
		if idx <= 0 {
			idx = 1
		}
		if idx >= n {
			idx = n - 1
		}
		for tries := 0; used[idx] && tries < n; tries++ {
			idx++
			if idx >= n {
				idx = 1
			}
		}
		used[idx] = true
		out = append(out, idx)
	}
	return out
}

// genGalaxy 程序化生成星系:以種子亂數在抖動網格上佈星,隨機光譜/大小/星名;
// 第 0 星為玩家母星、aiHomes 個星(見 aiHomeStarIndices)為各 AI 對手母星。
// n=星數(對應星系大小),回傳值第二項是各 AI 母星依序(對應 AIPlayers[0]、[1]、…)的星索引,
// 供呼叫端(NewDemoSession)直接拿來設 AIOpponent.ColonyStars,不必在呼叫端重算一次索引公式
// (先前 1 AI 版本 NewDemoSession 用 `aiHomeStar := galaxyStars/2` 手動重算,靠註解說明「與
// genGalaxy 內部規則一致」维持同步——兩處各算一次同一個公式是有漂移風險的重複邏輯,這裡改成
// 單一權威來源直接回傳)。
// 星名取自原版 STARNAME.LBX asset1 的 829 條隨機星名池(randomStarNamePool,見
// internal/shell/starnames.go),829 遠大於任何星系大小上限(最大 48 星),不需 fallback。
// nameTr 把星名池的英文原文翻成當前語言(nil = 不翻,即英文);見 GameSession.localName。
func genGalaxy(n int, seed int64, aiHomes int, age gamedata.GalaxyAge, nameTr func(string) string) ([]Star, []int) {
	r := rand.New(rand.NewSource(seed))
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	rows := (n + cols - 1) / cols
	stars := make([]Star, 0, n)
	aiIdx := aiHomeStarIndices(n, aiHomes)
	aiSet := make(map[int]bool, len(aiIdx))
	for _, x := range aiIdx {
		aiSet[x] = true
	}
	idx := 0
	// 先洗英文原文，再由同一個原文產生目前語言的顯示名。這樣中文星名
	// 不會覆蓋英文來源，動態事件報告才能在英文模式重建正確的星名。
	namesEN := make([]string, len(randomStarNamePool))
	copy(namesEN, randomStarNamePool)
	r.Shuffle(len(namesEN), func(i, j int) { namesEN[i], namesEN[j] = namesEN[j], namesEN[i] })
	for gy := 0; gy < rows && idx < n; gy++ {
		for gx := 0; gx < cols && idx < n; gx++ {
			x := (float64(gx) + 0.15 + r.Float64()*0.7) / float64(cols)
			y := (float64(gy) + 0.15 + r.Float64()*0.7) / float64(rows)
			nmEN := namesEN[idx%len(namesEN)]
			if idx >= len(namesEN) {
				nmEN = fmt.Sprintf("%s-%d", nmEN, idx/len(namesEN)+1)
			}
			nm := nmEN
			if nameTr != nil {
				nm = nameTr(nmEN)
			}
			owner := 0
			if idx == 0 {
				owner = 1 // 玩家母星
			} else if aiSet[idx] {
				owner = 2 // AI 母星(不分哪個 AI——個別歸屬見 AIOpponent.ColonyStars,Star.Owner
				// 只是粗粒度「有主/無主/玩家/AI」旗標,見型別註解)
			}
			// 光譜依原版 _star_class_table 加權(紅矮星近半、藍白星稀少),不是均勻亂數——
			// 均勻擲是 remake 星圖與原版最明顯的差異來源(見 gamedata.StarClassWeights)。
			// 母星所在的星強制為黃星:原版玩家母星恆在黃色恆星系,而加權骰有 37% 機率骰到紅矮星,
			// 那會讓母星落在生不出宜居行星的溫度帶(見 gamedata.ClassToGroup 紅矮列)。
			spectral := int(gamedata.RollSpectralClass(r, age))
			if owner != 0 {
				spectral = int(gamedata.Yellow)
			}
			// Star.Size 是星圖上的**視覺**大小(0=大..3=小),與行星大小無關,維持原本的均勻亂數。
			stars = append(stars, Star{X: x, Y: y, Spectral: spectral, Size: r.Intn(4), Name: nm, NameEN: nmEN, Owner: owner, Wormhole: -1})
			idx++
		}
	}
	return stars, aiIdx
}

// GameSession 是一局進行中的遊戲狀態。玩家操作改變狀態,EndTurn 推進一回合(結算玩家 + 各 AI)。
type GameSession struct {
	// TranslateName 把原版英文專有名詞(星名/艦名)翻成當前語言;nil = 不翻(英文)。
	//
	// **不進存檔**(小寫欄位會被 json 略過?不會——所以這裡靠 save 那一側的白名單,
	// 見 internal/save)。它是行程層的注入點,由 cmd/moo2 在建立對局時設定,
	// 讓 internal/shell 不必 import i18n。用法見 localName。
	TranslateName func(string) string `json:"-"`

	Turn int
	// 1 表示 ColonyState.AssimilationProgress 使用原版 0..239 點；零值是舊 JSON
	// 的「已累積回合」，由 restore／首次消費一次性轉換。
	AssimilationProgressVersion int
	Player                      engine.PlayerState
	PlayerColonies              []engine.ColonyState
	// PlayerCapitolPlanet 對應原版 player+0x29；Known 區分合法行星 0 與舊 JSON。
	PlayerCapitolPlanet          int
	PlayerCapitolPlanetKnown     bool
	PlayerCapitolRebuildRequired bool
	AIPlayers                    []AIOpponent
	// AIRelations 是 AI 對手彼此的 -40..40 顯示／消費投影。原版 raw current
	// 另存在 AIRelationsRaw；Diplomacy_Growth_ 每對完成後依高槽→低槽方向鏡射。
	AIRelations         [][]int
	AIRelationsRaw      [][]int
	AIRelationsRawKnown [][]bool
	// NPC_To_NPC_Treaty_Negotiations_ 的方向性 raw 狀態：+0x6D7 聲望／積怨、
	// +0x68F 條約記憶、+0x69F 協議記憶及 +0x63F 納貢模式。
	AIReputationRaw    [][]int
	AITreatyBiasRaw    [][]int
	AIAgreementBiasRaw [][]int
	AITributeModes     [][]int
	// AIIncidentReasonRaw／MagnitudeRaw／MemoryRaw 對應方向性 +0x64F／+0x65F／+0x71F。
	// 前兩者保存 Change_Relations_ 待處理的最強事件；MemoryRaw 是談判第三方 +5 的來源。
	AIIncidentReasonRaw    [][]int
	AIIncidentMagnitudeRaw [][]int
	AIIncidentMemoryRaw    [][]int
	AIIncidentBetrayalRaw  [][]bool
	// AIWarDurationRaw／AIDiplomacyCooldownRaw 對應方向性 +0x717／+0x72F。
	AIWarDurationRaw       [][]int
	AIDiplomacyCooldownRaw [][]int
	// EnableAIVsAI 是 remake 的可選強化開關。新示範／新局開啟；舊存檔缺欄位
	// 解為 false，避免在沒有快照資料時突然改變既有對局。
	EnableAIVsAI bool
	// AIWars / AIPolicies / AITrade / AIResearch 是 AI 彼此的正式外交狀態。
	// 這是 remake 的抽象矩陣，不是原版存檔欄位；只在 EnableAIVsAI 時消費。
	AIWars     [][]bool
	AIPolicies [][]gamedata.ForeignPolicy
	AITrade    [][]bool
	AIResearch [][]bool
	// LastAIAIBattle 是上一回合 AI 對 AI 抽象戰鬥的報告，不進存檔。
	LastAIAIBattle *AIAIBattleReport `json:"-"`
	// History 是逐回合的全帝國國力快照(原版 module 122 Record_History_ 的對應物,
	// 供 INFO 的 History Graph 子畫面畫折線;見 history.go 檔頭)。
	History       []HistoryTurn
	HistoryScales [4]int
	// LastEventReport 是上一回合觸發的隨機事件(結構化,供事件畫面用);nil = 本回合無事件。
	// 與 LastEvent(純文字,回合摘要用)同時寫入,見 events.go advanceEvents。
	// 比照 LastEvent 是「下回合會重算的顯示暫態」,刻意不存檔。
	LastEventReport *EventReport
	// StatusBroadcast 保存 GNN 29..35 的待播佇列與去重狀態；與上面的顯示暫態不同，
	// 必須進存檔，否則讀檔會重播帝國滅亡／Orion 或遺失同回合第二則新聞。
	StatusBroadcast StatusBroadcastState
	// PendingSurrenders 對應原版 player+0xE72：事件 34 建立後，下一個 consumer 才吸收資產。
	PendingSurrenders []EmpireSurrender

	// LastDiscovery 是上一回合抵達星系時觸發的一次性發現(太空殘骸/失散殖民地…);
	// nil = 無。與 LastEventReport 分開:隨機事件是全銀河的新聞,發現是自家艦隊的回報。
	LastDiscovery *SystemDiscovery
	// LastArtemis 是上一回合艦隊進入敵方星系時踩到的水雷網結算(阿提米絲系統網,見
	// artemis.go);nil = 沒踩到或該星沒有水雷網。與 LastDiscovery 分開:發現是好事、
	// 這個是損失,混在同一個欄位會讓 UI 分不出該用哪種語氣。
	LastArtemis      *ArtemisStrike
	LastPlayerOutput engine.EmpireOutput // 上一回合玩家結算(供畫面顯示)
	Stars            []Star              // 星系圖
	Nebulae          []Nebula            // 星雲(星圖地形,影響戰鬥護盾;見 nebula.go)
	// nebulaProbe 判定某個正規化座標是否落在星雲內。要讀星雲圖的遮罩,而本套件不碰資產,
	// 所以由 cmd/moo2 用 SetNebulaProbe 裝進來。**未匯出 = 不進存檔**,讀檔後要重裝。
	nebulaProbe func(x, y float64) bool
	Planets     []Planet // 行星列表
	Leaders     []Leader // 已雇用的軍官/領袖名單(Leader Pool)
	// ColonyLeaderNames 與 PlayerColonies 平行，保存殖民地領袖的穩定名稱識別。
	// 空字串表示未指派；完整技能仍由 Leaders 查回，避免存檔複製衍生資料。
	ColonyLeaderNames []string
	MercPool          []Leader    // 目前上門可雇用的傭兵領袖(手冊 p.134,見 advanceMercOffers/HireMerc)
	MercOfferedIdx    int         // 舊存檔／荒難領袖事件相容欄位；正常招募不再以它循序選人
	MercLastOfferTurn int         // 上次正常領袖 offer 回合；0 表示尚未出現
	MercCandidatePool []Leader    // 傭兵候選池(cmd 層由 HERODATA.LBX 真英雄注入,見 SetMercCandidates);nil=用內建策展名單
	OfficerCooldowns  map[int]int // AI 拒絕候選的剩餘 limbo 回合；key 是 HERODATA ID
	// Fleets 是玩家的艦隊清單,SelectedFleet 是目前操作中的那一支(見 fleet.go)。
	// 先前這裡是「Ships + FleetAtStar/DestStar/ETA/Marines/Tanks」一組欄位,
	// 也就是全帝國只有一支艦隊——那個限制卡住了星圖的遷移連線層與 F1/F2 循環。
	//
	// ⚠ 不變量:**永遠至少有一支艦隊**(可以沒有船)。由 ensureFleet 維持,
	// Fleet() 因此可以無條件回傳可寫指標,呼叫端不必逐處 nil 檢查。
	Fleets        []Fleet
	SelectedFleet int
	// ShipDesigns 是目前玩家的六艦體持久 blueprint；索引 0..5 與原版每筆
	// 99-byte 設計庫一致。舊存檔由 EnsureShipDesigns 依當前科技補齊。
	ShipDesigns []ShipBlueprint
	// ColonyRelocateTo 是各殖民地的集結點星索引(平行 PlayerColonies,見 relocation.go)。
	// ⚠ 預設必須是 −1 不是 Go 零值——0 是母星的索引。
	ColonyRelocateTo []int
	// ShowRelocationLines 是星圖遷移連線的顯示開關(原版 `byte_199BE4`,手冊那組設定裡的一項)。
	// 預設開:原版新開一局是畫的(`sub_127E1` 初始化時寫 1)。
	ShowRelocationLines bool
	// GameSettings 保存原版 SETTINGS 分頁的完整偏好；ShowRelocationLines 暫留作舊存檔與
	// 現有星圖消費端的相容鏡像，所有新寫入應經 ApplyGameSettings。
	GameSettings GameSettings
	LastBattle   *BattleResult // 上一場戰鬥結果(供戰鬥結果畫面)
	// LastLeaderUpkeep 是本回合實際扣掉的領袖維護費(見 leader_upkeep.go)。
	// 不進存檔:它是「這一回合發生了什麼」的展示值,下一次 EndTurn 會重算。
	LastLeaderUpkeep int `json:"-"`
	// LastBankruptcy 是本回合負國庫時的自動資產處分報告，不進存檔。
	LastBankruptcy []BankruptcyAction `json:"-"`

	// LastRebellions 是上一回合的叛亂檢定結果(供回合摘要;沒有事情發生時是 nil)。
	// 不進存檔:它是「這一回合發生了什麼」的展示資料,重載存檔時本來就沒有上一回合。
	LastRebellions []RebellionResult `json:"-"`

	// LastAIArrivals 是上一回合抵達某顆星的 AI 艦隊(供回合摘要 / 水雷回報)。
	// 同樣不進存檔:它是展示資料,不是狀態。
	LastAIArrivals []AIFleetArrival `json:"-"`
	SelectedStar   int              // 星圖選中的星索引(-1=未選)
	Difficulty     int              // 難度索引(shell.Difficulties)
	Builds         []ColonyBuild    // 各殖民地「當前建造中」的項目(對應 PlayerColonies;佇列見 BuildQueue)
	// BuildQueue[i] 是殖民地 i 的**後續**建造排隊項(不含 Builds[i] 那一格)。
	// 原版殖民地畫面的 BUILD QUEUE 是 7 格(反組譯 Add_Build_Queue_Fields_ 確認),
	// 完工自動接下一項;remake 先前只有一格。見 buildqueue.go 檔頭。
	BuildQueue [][]ColonyBuild
	// AutoBuild / RepeatBuild 與 PlayerColonies 平行。AutoBuild 是 AUTO BUILD 的
	// 開關；RepeatBuild 的零值表示未指定重複項目，非零值只允許可重複的 Special。
	// 兩者由 ensureBuildQueue 與殖民地數對齊，因此舊存檔零值安全。
	AutoBuild   []bool
	RepeatBuild []ColonyBuild
	LastBuilt   []BuildNotice `json:"-"` // 上回合建造結果；暫態 typed UI 資料
	popAccum    []int         // 各殖民地人口成長累加值(達門檻則 +1 人口)

	// --- 地面戰入侵(見 ground_invasion.go) ---
	PlayerColonyMarines []int             // 各玩家殖民地 Marine Barracks 駐軍池(平行 PlayerColonies)
	MarineBarracksAge   []int             // 各玩家殖民地 Marine Barracks 已運作回合數(平行 PlayerColonies)
	ColonyBuildings     []map[string]bool // 各殖民地已完工建築(去重,避免重複套用長期效果)

	// PlayerColonyStars 是 PlayerColonies[i] 對應到 Stars 的索引(平行陣列),與
	// AIOpponent.ColonyStars 同一設計(見該欄位註解)。開局只有一筆(母星→星 0,
	// NewDemoSession 設定)。ColonizeStar(colonization.go)每建一個新殖民地都會同步 append;
	// InvadeColony(ground_invasion.go)過戶敵方殖民地時也會同步 append 被佔領的星索引。
	// 舊存檔/舊呼叫端若未同步到這個欄位而導致長度落後 PlayerColonies,兩處寫入前都會先
	// padding 補 -1(語意「星索引未知」)再 append 真正值,維持 len(PlayerColonyStars)==
	// len(PlayerColonies) 的不變量。
	PlayerColonyStars []int
	// PlayerColonyPlanets 是 PlayerColonies[i] 座落的**行星**索引(平行陣列,對 s.Planets)。
	//
	// 為什麼在有了 PlayerColonyStars 之後還要這個:**一個星系可以有多個殖民地**
	// (手冊 p.61 的殖民地是建在行星上,而一個星系有 1..5 個天體)。只記星索引時,
	// 「同星系的第二顆殖民地」會與第一顆共用同一筆行星資料,氣候/重力/物產全部串在一起。
	// 維護慣例同 PlayerColonyStars(padding −1 = 行星索引未知,通常是舊存檔)。
	PlayerColonyPlanets []int

	// FleetTanks / PlayerColonyTanks / ArmorBarracksAge:裝甲營房(Armor Barracks)戰車營
	// 駐軍系統,與上面三個 Marine 對應欄位對稱(見 advanceArmor/LoadTanks,ground_invasion.go)。
	PlayerColonyTanks     []int       // 各玩家殖民地 Armor Barracks 駐軍池(平行 PlayerColonies)
	ArmorBarracksAge      []int       // 各玩家殖民地 Armor Barracks 已運作回合數(平行 PlayerColonies)
	EventSeed             int64       // 隨機事件亂數種子(可重現;新遊戲遞增)
	LuckyEventCounter     int         // 原版 Lucky 額外好事件的逐帝國累積計數器
	EventLastTurn         int         // 原版全局一般事件最後成功回合（相對星曆，Turn-1）
	EventAttemptCounter   int         // 原版一般事件前五次保護檢查計數器（0..5）
	LastEvent             string      // 本回合觸發的隨機事件描述(空=無事件;供回合摘要)
	LastPersistentEventEN string      `json:"-"` // 持續事件英文播報(與 LastEvent 同回合,不進存檔)
	DisableEvents         bool        // 關閉隨機事件(供確定性經濟測試隔離)
	eventRand             *randStream // 事件亂數源(由 EventSeed 惰性建立;抽取次數會進存檔,見 randstream.go)
	researchRand          *randStream // 研究選項亂數源(缺乏創造力;抽取次數會進存檔)
	// commandRecorder 只記錄目前真人席位在本回合的玩家操作；AI 與世界結算時
	// 由 commandReplayDepth 暫停記錄。它不進存檔，因為是傳輸層的行程狀態。
	commandRecorder    func(PlayerCommand)
	commandReplayDepth int
	AntaresRaids       int                  // 已成功部署的安塔蘭攻擊艦隊數
	AntaranInvasion    AntaranInvasionState // 原版全局資源／建艦／出征狀態
	LastAntaranNotice  *AntaranNotice       // 本回合安塔蘭突襲的型別化結果；玩家句型由 UI 提供
	// Monsters 是星圖上守衛星系的太空怪獸(見 monster.go)。清空 = 全部已被清除。
	Monsters []MonsterGuard

	// CapturedPop 是玩家透過地面入侵俘虜來的人口單位總數(見 ground_invasion.go
	// InvadeColony)。原版計分對俘虜人口另有一份加分(手冊 p.184「You also get a premium
	// for captured population units」),見 score.go。
	CapturedPop int
	// ScoreBaseMultiplierPercent 是客製種族畫面依未使用 Picks 固定的基礎倍率。
	// 內建種族與舊存檔為 100；Evolutionary Mutation 的未消費 4 Picks 在計分時另加。
	ScoreBaseMultiplierPercent int

	// PersistentEvents 是進行中的持續型隨機事件(超新星倒數/時空異象/超空間獸/人口暴增/瘟疫,
	// 見 events_persistent.go)。空 = 沒有任何持續事件。
	PersistentEvents []PersistentEvent

	// Outposts 是玩家的軍事前哨站(見 outpost.go)。與 PlayerColonies **完全分開**——
	// 手冊 p.85「produces nothing」,前哨站沒有人口也沒有產出,混進殖民地陣列會讓帝國經濟
	// 憑空多出一個殖民地。
	Outposts []Outpost

	// LastRaidReport 是本回合 AI 對玩家殖民地的型別化突襲結果(見 ai_attack.go);
	// 空/nil = 無。與 LastAntaranNotice 分開:安塔蘭人是週期腳本,AI 突襲是外交/軍備的後果。
	LastRaidReport *AIRaidReport
	RaceIndex      int // 玩家選定的種族(shell.Races 索引)
	// CustomRaceTraits 是客製種族畫面所選的布林能力位元遮罩。零值代表沒有特殊能力；
	// 舊存檔沒有此欄位時自然退回此語意。它與 RaceIndex=-1 一起表示客製種族。
	CustomRaceTraits uint32
	// CustomRaceRuntimeTraits 是客製種族已展開的 player+0x89F 31 格。
	// 數值 1..9 與布林 10..30 共用同一份真相，供原版 AI／科技估值消費。
	CustomRaceRuntimeTraits [gamedata.RaceTraitCount]int8
	PlayerName              string // 玩家帝國/領袖名稱(新遊戲命名畫面設定)
	FlagColor               int    // 玩家旗幟顏色索引(shell.FlagColors)
	RaceCombatPct           int    // 種族**艦艇攻擊**百分點加成(原版 TRAIT_SHIP_ATTACK)
	RaceShipDefPct          int    // 種族艦艇防禦百分點加成(原版 TRAIT_SHIP_DEFENSE)
	RaceGroundBonus         int    // 種族地面戰定值加成(原版 TRAIT_GROUND_COMBAT)
	RaceSpyBonus            int    // 種族諜報定值加成(原版 TRAIT_SPYING)

	raceGrowthPct int // 種族人口成長百分點加成(供 advancePopulation)

	// Government 是玩家目前政府型態(2026-07-11 接線,供 colonyMoralePercent 士氣計算用)。
	// 由 ApplyGovernment 設定;新遊戲若從未呼叫 ApplyGovernment,預設見 NewDemoSession
	// (獨裁/Dictatorship,對應自訂種族 0 點基準)。
	//
	// Go 零值陷阱(比照 ColonyState.PlanetGravity 同款註解):gamedata.MoraleGovernmentType 的
	// 零值是 MoraleGovFeudalism(iota 從 0 開始,見 morale.go enum 順序),不是想要的預設政府
	// Dictatorship——任何建構 GameSession 卻沒有明確設定本欄位的呼叫端,會被誤判為封建政府,
	// 必須明確賦值,不能依賴零值。
	Government gamedata.MoraleGovernmentType
	// newGameRacePending 只在 SetupNewGame 到種族／政府套用完畢之間為 true。
	// 它不存檔；用來防止遊戲中途 ApplyGovernment 重置研究。
	newGameRacePending bool

	// RuleProfile 是這局遊戲的版本規則 profile(patch 1.3 vs 1.5,見
	// gamedata.RuleProfile/docs/tech/version-1.3-1.5-diff.md)。唯讀設定,開局決定、遊戲中不可變
	// (原版本身也是「一開局就決定規則集」,無 mid-game 切換)。
	//
	// Go 零值陷阱:gamedata.RuleProfile{} 的零值三個欄位皆為 0(不是任何一版的真值),任何建構
	// GameSession 卻沒有明確設定本欄位的呼叫端,會導致轟炸輪數/研究成本/武器傷害查詢異常——
	// NewDemoSession 已明確設為 gamedata.Profile15()(=現行硬編值,no-op),新的建構路徑
	// (未來 UI 選 1.3)須呼叫 SetRuleProfile 或直接賦值,不能依賴零值。
	RuleProfile gamedata.RuleProfile

	// --- 勝利條件(見 council.go)---
	Victory                VictoryState        // 遊戲是否已分出勝負(Over=true 後不再產生新的議會選舉)
	PendingCouncilElection *CouncilElection    // 非玩家當選、等待玩家 RespondToCouncilElection 回應(手冊:議會無法強迫玩家接受)
	PendingCouncilVote     *CouncilVotePending // AI 已投票、等待真人選候選人或棄權
	LastCouncilNotice      *CouncilNotice      // 本回合議會型別化通知(nil=無；供回合摘要)
	CouncilMeetings        int                 // 已召開過的議會屆數
	// OriginalCouncilDiplomacyState 對應 word_19A0E2：0=尚未開會、1=已開會但無勝者、
	// 2=真人當選、3=其他帝國當選。Known 區分舊 JSON／GAM 缺 raw 與合法 0。
	OriginalCouncilDiplomacyState      int
	OriginalCouncilDiplomacyStateKnown bool
	lastCouncilTurn                    int         // 上次召開議會的回合數(0=從未召開)
	CouncilLastVotes                   []int       // 帝國順序為玩家、AIPlayers；-2=上屆棄權／未知，-1=玩家
	councilRand                        *randStream // 議會 Vote_Check_ 的獨立 1..200 可重播亂數流

	// AntaranHomeworldConquered 是手冊三條勝利路徑之二「攻陷安塔蘭母星」的達成旗標(見
	// antaran_victory.go)。由 AssaultAntares 戰勝後設為 true;engine.CheckAntaranVictory 讀取
	// 這個布林旗標判定(該函式本身不追蹤戰鬥流程,見其註解),advanceAntaranVictory 依此設定
	// s.Victory。Go 零值(false)即想要的預設值,無零值陷阱。
	AntaranHomeworldConquered bool

	// --- 間諜(見 spy.go / spy_mission.go) ---
	// PlayerSpies 是玩家派駐到 AIPlayers[i] 的間諜數(平行 AIPlayers)。opt-in,預設 0
	// (Go 零值即想要的預設值)。玩家經 TrainSpy(idx) 花 BC 增加;逐對手分配已經是這個陣列
	// 天然支援的結構。PlayerSpyMissions 同樣平行保存 STEAL/SABOTAGE/HIDE 任務；
	// SABOTAGE 的最小建築破壞結算已接入 spy.go。
	PlayerSpies       []int
	PlayerSpyMissions []SpyMission
	// DefensiveAgents 是玩家帝國共用的防守 Agent 數量，與 PlayerSpies 的逐對手
	// 進攻配置分開保存。
	DefensiveAgents int

	// Seats / ActiveSeat 是熱座多人(見 hotseat.go)。單人局 Seats 為 nil、ActiveSeat 為 0,
	// 所有既有邏輯逐位元不變——熱座只在席位數 > 1 時才會動到任何東西。
	Seats      []seat
	ActiveSeat int

	// GalaxyAge / TechLevel 是 NEW GAME 畫面的兩個設定(見 shell.GalaxyAges / TechLevels)。
	//
	// Go 零值陷阱:`gamedata.GalaxyAge` 的零值是 `GalaxyYouthful`(不是想要的 Average),
	// 所以另用 `GalaxyAgeSet` 標記「有沒有真的設過」——舊存檔與沒走設定畫面的建構路徑
	// 解出來 GalaxyAgeSet=false,`galaxyAge()` 退回 Average,與加這個欄位之前的行為一致。
	// TechLevel 同款零值陷阱,而且**更危險**:零值 0 = 「曲速前」,而曲速前的艦隊
	// 離不開本星系(見 FleetHasFTL)。沒有標記的話,舊存檔與任何沒設過這欄的建構路徑
	// 會被靜默判成曲速前、艦隊整個凍住。所以同樣用 `TechLevelSet` 標記,
	// 未設過一律當「一般」(techLevel() 的退路)。
	// ⚠ 2026-08-07 就是在這裡踩到:接上 FTL 限制後 TestFleetInterstellarMovement 立刻
	// 紅燈,因為 NewDemoSession 沒設過這欄。
	GalaxyAge    gamedata.GalaxyAge
	GalaxyAgeSet bool
	TechLevel    int
	TechLevelSet bool
	// LastEspionage 是本回合諜報結算的訊息(供回合摘要顯示;每回合開頭清空)。
	LastEspionage       []string
	spyRand             *randStream // 間諜擲骰亂數源(由 EventSeed 惰性建立,比照 eventRand 慣例)
	discoveryRand       *randStream // 星系發現擲骰亂數源(見 discovery.go discoveryRoll,同上慣例)
	populationRand      *randStream // 負成長刪人口 reservoir sampling；抽取次數隨存檔保存
	agreementRand       *randStream // 貿易／研究協議 goal%5 的逐方向補點擲骰；抽取次數隨存檔保存
	diplomacyGrowthRand *randStream // Diplomacy_Growth_ 條約關係增益；獨立流保證存讀／鎖步可重播
	officerRand         *randStream // 隨機領袖招募的百分比與候選抽取；抽取次數隨存檔保存
}

// advanceEvents 的實作位於 events.go：排程、候選 ID、帝國目標與已閉合的效果逐項依原版
// IDA 證據接線；尚未閉合的事件仍消耗候選但不播空訊息。事件亂數由 EventSeed 與可存檔
// randStream 決定，禁止再以固定機率或六種 remake 自編事件描述這條路徑。
// SendFleet 派遣玩家艦隊前往 dest 星:依兩星歐氏距離換算航行回合數(ETA),每回合 EndTurn
// 遞減。dest 無效、與現址相同、或艦隊正航行中則忽略。回傳是否成功下令。
func (s *GameSession) SendFleet(dest int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdSendFleet, Args: []int{dest}})
	if s.hyperspaceFluxActive() && !s.raceHasTrait(gamedata.TRAIT_TRANS_DIMENSIONAL) {
		return false
	}
	if dest < 0 || dest >= len(s.Stars) || dest == s.Fleet().AtStar || s.Fleet().ETA > 0 {
		return false
	}
	if !s.FleetHasFTL() {
		return false // 曲速前開局:沒有 FTL 就出不了本星系(見 FleetHasFTL)
	}
	// 黑洞:手冊「No ship can safely pass within 2 parsecs of a black hole
	// (unless the ship contains an officer with the Navigator skill)」。蟲洞不走實空間,不受限。
	if !s.WormholeBetween(s.Fleet().AtStar, dest) && s.RouteBlockedByBlackHole(s.Fleet().AtStar, dest) {
		return false
	}
	eta := 1
	// 蟲洞:兩端直通,不看距離。這是 MOO2 蟲洞的**遊戲機制**價值——把銀河兩頭接起來,
	// 讓遠端的殖民地不再是「派兵要十回合」的孤島(見 internal/shell/wormhole.go)。
	// 固定星圖蟲洞的航行尺度仍是 1 回合；隨機事件 28 則已由 1.31 IDA 證實在事件
	// consumer 內立即寫成抵達，兩者不可再共用同一個 ETA=1 實作。
	if !s.WormholeBetween(s.Fleet().AtStar, dest) {
		// 秒差距模型(見 internal/shell/starlane.go):距離換成整數秒差距、除以艦隊航速。
		// 先前是 `ceil(正規化距離 × 8)` 這種沒有速度概念的固定換算,手冊裡以「秒差距/回合」
		// 表述的規則(星雲降速、Navigator、干擾場)因此全都無處可掛。
		eta = s.FleetETATo(s.Fleet().AtStar, dest)
		if eta < 1 {
			eta = 1
		}
	}
	s.Fleet().DestStar = dest
	s.Fleet().ETA = eta
	return true
}

// FTLTopic 是「解鎖星際航行」的研究主題。
//
// remake 的科技樹裡 `TECH_NUCLEAR_DRIVE`(核融合引擎,MOO2 的入門 FTL 引擎)屬於
// `TOPIC_NUCLEAR_FISSION`(techtree.go 第 55 列,Cost 50、ResearchAll)——不在開局就給的
// `TOPIC_STARTING_TECH` 裡。所以「有沒有 FTL」在 remake 就是「這個主題研究完了沒」。
const FTLTopic = gamedata.TOPIC_NUCLEAR_FISSION

// FleetHasFTL 回傳艦隊能不能離開本星系。
//
// 手冊直引(Pre-warp 起始文明,見 docs/tech/homeworld-init.md 的引文):
//
//	"Every race has one colony — their home star system. Exploring outside that system is
//	 impossible until faster than light (FTL) technologies are discovered."
//
// **只有「曲速前」(TechLevel 0)這一級受限**:一般 / 先進開局本來就配了 FTL 引擎,
// 手冊描述它們開局就有星際艦。TechLevel 的其餘效果仍未接,見 shell.TechLevels 註解。
//
// ⚠ 這條在 2026-08-07 之前完全沒有實作:NEW GAME 的 TECH LEVEL 是「只存設定、不影響
// gameplay」,選了曲速前照樣能全圖亂飛。這是接上的第一個真效果。
func (s *GameSession) FleetHasFTL() bool {
	if s.techLevel() != TechLevelPrewarp {
		return true
	}
	return s.Player.CompletedTopics != nil && s.Player.CompletedTopics[FTLTopic]
}

// TechLevelPrewarp / TechLevelDefault 是 shell.TechLevels 裡的索引。
const (
	TechLevelPrewarp = 0 // 曲速前
	TechLevelDefault = 1 // 一般(未設定時的退路)
)

// techLevel 回傳這一局的起始科技等級。未設定(舊存檔、demo session)時退回「一般」——
// **不能直接讀 TechLevel**,零值 0 是「曲速前」,見該欄位註解。
func (s *GameSession) techLevel() int {
	if !s.TechLevelSet {
		return TechLevelDefault
	}
	return s.TechLevel
}

// advanceFleet 推進艦隊航行:ETA 遞減,歸零則抵達(FleetAtStar=目的),並將該星標記為已探索。
func (s *GameSession) advanceFleet() {
	if s.hyperspaceFluxActive() && !s.raceHasTrait(gamedata.TRAIT_TRANS_DIMENSIONAL) {
		return
	}
	// **逐艦隊推進**:多艦隊之後每一支各自航行(自動遷移的新艦也在其中,見 relocation.go)。
	for i := range s.Fleets {
		f := &s.Fleets[i]
		if f.ETA <= 0 || f.DestStar < 0 {
			continue
		}
		f.ETA--
		if f.ETA != 0 {
			continue
		}
		s.completePlayerFleetArrival(f)
	}
	s.mergeColocatedFleets()
}

// completePlayerFleetArrival 是一般航行與事件 28 共用的立即抵達 consumer。
// 呼叫時 DestStar 必須仍保留，讓事件不會只改座標而漏掉探索、水雷與發現效果。
func (s *GameSession) completePlayerFleetArrival(f *Fleet) bool {
	if f == nil || f.DestStar < 0 || f.DestStar >= len(s.Stars) {
		return false
	}
	f.AtStar = f.DestStar
	f.DestStar = -1
	f.ETA = 0
	s.Stars[f.AtStar].Explored = true
	target := eventEmpireTarget{kind: eventEmpirePlayer, index: 0, alive: true}
	if s.HotseatEnabled() {
		target = eventEmpireTarget{kind: eventEmpireSeat, index: s.ActiveSeat, alive: true}
	}
	s.queueOrionDiscoveryBroadcast(target, f.AtStar)
	// 阿提米絲系統網:手冊寫的是「any enemy ship **entering** that system」,
	// 是**進入的那一刻**而不是停留每回合,所以結算在這裡(見 artemis.go)。
	if a := s.applyArtemisMines(f, f.AtStar); a != nil {
		s.LastArtemis = a
	}
	// 抵達星系當下結算一次性發現(原版 Do_System_Discoveries_At_Star_)。
	if d := s.discoverSystemSpecials(f.AtStar); d != nil {
		s.LastDiscovery = d
	}
	return true
}

// mergeColocatedFleets 把停在同一顆星、都沒有航行任務的艦隊併成一支。
//
// 為什麼要併:自動遷移每回合可能生出新的一支,不併的話艦隊清單會無限長大,
// 而玩家看到的是「同一顆星上一堆各有一兩艘船的艦隊」——那不是原版的樣子。
// 原版的艦隊本來就是「在同一個地方的船的集合」。
//
// ⚠ 只併**都靜止**的:還在航行的艦隊各有各的目的地,併了會弄丟任務。
func (s *GameSession) mergeColocatedFleets() {
	for i := 0; i < len(s.Fleets); i++ {
		if s.Fleets[i].DestStar >= 0 {
			continue
		}
		for j := len(s.Fleets) - 1; j > i; j-- {
			if s.Fleets[j].DestStar >= 0 || s.Fleets[j].AtStar != s.Fleets[i].AtStar {
				continue
			}
			s.Fleets[i].Ships = append(s.Fleets[i].Ships, s.Fleets[j].Ships...)
			s.Fleets[i].Marines += s.Fleets[j].Marines
			s.Fleets[i].Tanks += s.Fleets[j].Tanks
			s.Fleets = append(s.Fleets[:j], s.Fleets[j+1:]...)
			if s.SelectedFleet == j {
				s.SelectedFleet = i // 被併掉的那支正被選著 → 焦點跟到合併後的那支
			} else if s.SelectedFleet > j {
				s.SelectedFleet--
			}
		}
	}
	s.ensureFleet()
}

// totalBuildingMaintenance 加總玩家目前所有殖民地「已建成」建築的維護費(BC/回合),取代
// 先前 Player.Maintenance 平坦寫死 5 的 placeholder。逐殖民地用 gamedata.BuiltMaintenanceBC
// 查表加總。本函式刻意只負責建築；運輸艦、指揮點超支、間諜與軍官由各自分項計算，
// 最後在 RunEmpireTurn 一次扣除。s.ColonyBuildings 為 nil 或某殖民地
// 尚無記錄時,該殖民地視為 0(尚未建成任何建築,非漏算)。
func (s *GameSession) totalBuildingMaintenance() int {
	total := 0
	for i, built := range s.ColonyBuildings {
		if s.ColonyInStasis(i) {
			// 手冊 p.181:被時空異象凍結的殖民地「do not need food or cost maintenance either」。
			continue
		}
		total += gamedata.BuiltMaintenanceBC(built)
	}
	return total
}

// coloniesForTurn 回傳這回合要交給 engine.RunEmpireTurn 結算的殖民地清單。
//
// 與 s.PlayerColonies **索引一一對應**(呼叫端 advanceBuilds/advancePopulation 都靠這個對齊),
// 事件需要修改回合輸入時才建立副本：被時空異象凍結的殖民地以人口 0 停止產出、食物、成長
// 與建造；超新星星系則只標記 ResearchDiverted，保留 ColonyOutput.Research 給搶救進度，卻不
// 讓同一批 RP 進入一般研究。
//
// 為什麼用副本而不是「從清單裡拿掉」:拿掉會破壞索引對齊,而那個對齊是
// 好幾個呼叫端共用的不變量。零人口副本讓凍結成為一個純資料上的效果,不動任何控制流。
func (s *GameSession) coloniesForTurn() []engine.ColonyState {
	needsCopy := false
	for i := range s.PlayerColonies {
		star := s.PlayerColonyStarIndex(i)
		planet := s.ColonyPlanetIndex(i)
		if s.ColonyInStasis(i) || s.StarUnderSupernova(star) ||
			s.planetPopulationBoomActive(planet) || s.planetPlagueActive(planet) {
			needsCopy = true
			break
		}
	}
	if !needsCopy {
		return s.PlayerColonies // 常態:沒有凍結的殖民地,直接用原本的切片,零額外成本
	}
	out := make([]engine.ColonyState, len(s.PlayerColonies))
	copy(out, s.PlayerColonies)
	for i := range out {
		star := s.PlayerColonyStarIndex(i)
		if s.planetPopulationBoomActive(s.ColonyPlanetIndex(i)) {
			out[i].GrowthBonusSum += 100
		}
		if s.planetPlagueActive(s.ColonyPlanetIndex(i)) {
			out[i].GrowthBonusSum -= 200
		}
		if s.StarUnderSupernova(star) {
			out[i].ResearchDiverted = true
		}
		if !s.ColonyInStasis(i) {
			continue
		}
		freezeColonyForStasis(&out[i])
	}
	return out
}

// totalCommandPointsSupply 加總玩家目前的指揮評等(Command Rating)供給:帝國基礎值
// gamedata.CommandPointsBase(每帝國加一次,非逐殖民地——見該常數註解的 oracle 依據與
// flat-vs-per-colony 不確定性 TODO)加上所有殖民地「已建成」軌道衛星(星基/戰鬥站/星辰要塞)
// 供給的總和(逐殖民地用 gamedata.CommandPointsFromBuildings 查表,三者取代關係已在該函式內
// 處理,不會重複疊加)。與 totalBuildingMaintenance 同款模式:s.ColonyBuildings 為 nil 或
// 某殖民地尚無記錄時,該殖民地視為 0。
//
// 2026-07-11 修正:先前這裡只加總建築供給,漏了基礎值,導致開局供給只有星基 1 點 < 需求 3 點
// (殖民船+2偵察艦),每回合 -20 BC 死亡螺旋(SAVE10.GAM oracle 反推證實基礎應為 5,見
// gamedata.CommandPointsBase 註解)。修正後開局供給 = 5(基礎)+1(星基)= 6 ≥ 3,不再超支。
// CommandPointsSupplyNow / CommandPointsUsedNow 是**當下現算**的指揮點數供給與需求。
//
// ⚠ 為什麼需要它們:`Player.CommandPointsSupply` / `Player.UsedCommandPoints` 是**快取欄位**,
// 只在 `EndTurn` 結算時更新。開局第一回合、或剛蓋好一座星基還沒結算時,那兩個欄位是舊值——
// 畫面直接讀它們會顯示「起始 5 + 建築 1,總計卻是 1」這種自相矛盾的組合(指揮點數視窗
// 第一版就是這樣被抓到的)。要顯示給玩家看的地方一律走這兩支。
func (s *GameSession) CommandPointsSupplyNow() int { return s.totalCommandPointsSupply() }

// CommandPointsUsedNow 見 CommandPointsSupplyNow。
func (s *GameSession) CommandPointsUsedNow() int { return s.usedCommandPoints() }

func (s *GameSession) totalCommandPointsSupply() int {
	total := gamedata.CommandPointsBase
	for _, built := range s.ColonyBuildings {
		total += gamedata.CommandPointsFromBuildings(built)
	}
	// Operations 是帝國層的 Command Rating 加成；與殖民地建築同樣只取
	// 最佳的一位 common 領袖，不把同一技能重複疊加。
	total += leaderEmpireSkillBonus(s.Leaders, gamedata.SKILL_OPERATIONS)
	return total
}

// usedCommandPoints 加總玩家目前所有艦艇(s.Ships)消耗的指揮評等(Command Rating)點數
// (GAME_MANUAL.pdf p.169,gamedata.ShipCommandCost)。s.Ships 目前的艦體種類只有殖民船/
// 偵察艦/六級戰鬥艦體(見 shipClassFromName),不含貨運艦隊(Freighter Fleet)——本專案未把
// 貨運艦隊塑模成 Ship 條目(IncomeFreighterMaintenanceCost 走獨立的「使用中運輸艦數量」參數,
// 與 s.Ships 無關),故不會誤把手冊明文排除的貨運艦隊算進指揮評等需求。偵察艦不在手冊 Ship
// Design 六級表內,shipClassFromName 近似當 Frigate(=1 點)處理,與其他空間/戰力計算一致。
func (s *GameSession) usedCommandPoints() int {
	total := 0
	for _, sh := range s.AllShips() { // 手冊 p.169 明文是**全帝國**艦艇,不是單一艦隊
		class, _ := shipClassFromName(sh.Class)
		total += gamedata.ShipCommandCost(class)
	}
	return total
}

// recoverFromFamine 饑荒防死鎖:若某玩家殖民地上回合結算後 Farmers=0 且 Starving(食物盈餘
// <0),但仍有人口(Population>0),自動把 1 個非農夫單位(優先 Worker,其次 Scientist)
// 改派回農業,近似「玩家發現饑荒會手動 ShiftColonyJob 自救」的行為。
//
// 沒有這個機制,零緩衝經濟一旦被隨機事件/安塔蘭入侵把僅有的農夫扣到 0，現在已閉合的
// 原版 signed 成長還會依短缺累積死亡率；若玩家一直不手動改派，殖民地會在最後一人保護線前
// 持續減員並卡在 NetIndustry=0、TaxRevenue=0。這是
// docs/tech/colony-economy-maintenance.md §2.2 實測記錄的死結,本函式是解法之一(任務指示的
// 選項②):非饑荒(Farmers>0 或未 Starving)不動作,一次只搶救 1 人(避免一次饑荒就把整個
// 職務分配打亂),不改動 AI 殖民地(AI 目前的 Farmers/Workers 由 ApplyAIEconomy 每回合依
// decider 重新決定,不會卡在饑荒鎖死)。
// 2026-08-06 修正觸發條件(反組譯佐證):原本只在 Farmers==0 才搶救,但實測有更常見的
// 死結——殖民地農夫還有好幾個卻仍然缺糧。早期一次性瘟疫 `losePopulation` 曾重現此問題；
// 現行事件 16 已改走原版 signed growth，但其他人口傷亡路徑仍可能留下相同職務失衡。
// 原版 Orion2.exe 有 Assign_Additional_Unblockaded_Farmers_(A 級硬證,見 docs/re/01-gap-report.md),
// 語意是「缺糧時加派『額外的』農夫」,不是「農夫歸零才救」。故改成:只要 Starving 就每回合
// 補 1 名農夫,不論目前農夫數——任何成因(事件扣人、人口成長、氣候改變)造成的赤字都會自癒。
func (s *GameSession) recoverFromFamine() {
	for i := range s.PlayerColonies {
		c := &s.PlayerColonies[i]
		if c.Population <= 0 {
			continue
		}
		if i >= len(s.LastPlayerOutput.Colonies) || !s.LastPlayerOutput.Colonies[i].Starving {
			continue
		}
		switch { // 一次只搶救 1 人,避免一次饑荒把整個職務分配打亂
		case c.Workers > 0:
			c.Workers--
			c.Farmers++
		case c.Scientists > 0:
			c.Scientists--
			c.Farmers++
		}
	}
}

// syncTradeGoodsFlag 依 s.Builds(建造選單,UI 側狀態,對應 PlayerColonies——見該欄位註解)
// 目前選擇,同步各殖民地 engine.ColonyState.TradeGoods 旗標。玩家把某殖民地的建造項切到
// 「貿易品」(TradeGoodsBuildName)時,此處把旗標同步到 engine 層,結算時
// (engine.RunEmpireTurn)才知道該殖民地當回合要把淨工業換 BC 而非蓋建築。以 s.Builds 為
// 單一真相來源、只在結算前同步一次,不在 CycleColonyBuild 額外維護第二份旗標,避免兩處
// 狀態不同步。
func (s *GameSession) syncTradeGoodsFlag() {
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].TradeGoods = i < len(s.Builds) && s.Builds[i].Name == TradeGoodsBuildName
		// 住宅同理:選它時該殖民地進入「住房」產能配置,啟用 gamedata.ColonyHousingBonus 的
		// 成長加成(engine.colonyGrowth 早已支援,先前無人設過這個旗標)。
		s.PlayerColonies[i].Housing = i < len(s.Builds) && s.Builds[i].Name == HousingBuildName
	}
}

// prepPlayerDerived 把「結算前要先重算的玩家衍生值」算好。
//
// 抽成獨立函式是為了讓熱座的其餘席位(hotseat.go advanceSeatEmpire)能跑同一段,
// 不必複製一份會漂移的副本。內容與順序原樣搬自 EndTurn,無行為變更。
func (s *GameSession) prepPlayerDerived() {
	// 舊存檔可能已知 Battleoids，卻早於 GrantedTechs 欄位而缺少原版連帶授予。
	applyResearchTopicGrantCallbacks(&s.Player, gamedata.TOPIC_ASTRO_CONSTRUCTION)
	s.Player.Maintenance = s.totalBuildingMaintenance() // 依本回合結算前的實際已建建築重算(取代平坦常數)
	s.Player.SpyMaintenance = s.totalSpyMaintenance()
	s.Player.OfficerMaintenance = s.LeaderUpkeepTotal()
	s.Player.CommandPointsSupply = s.totalCommandPointsSupply() // 指揮評等供給:實際已建成的星基/戰鬥站/星辰要塞
	s.Player.UsedCommandPoints = s.usedCommandPoints()          // 指揮評等需求:玩家目前所有艦艇加總
	// GovtBonusMoneyPercent 依目前政府型態算好傳給 RunEmpireTurn(engine 層不關心政府列舉本身,
	// 見 engine.PlayerState.GovtBonusMoneyPercent 註解)。demo 預設 Dictatorship → 0,no-op。
	// ActiveFreighters 這裡不需要顯式賦值——它不是每回合重算的衍生值(不同於 Maintenance/
	// CommandPointsSupply),而是由 advanceBuilds()(本函式下方)在「運輸艦隊」完工當下直接
	// 累加的持久欄位(2026-07-11(#4)接線,見 engine.PlayerState.ActiveFreighters 註解、
	// gamedata.FreighterFleetActionName)。此處呼叫 RunEmpireTurn 時吃的是「上回合累積到現在」
	// 的值,故新完工的運輸艦要下回合才開始計維護費——此處的零值陷阱只在「玩家從未建過運輸艦隊」
	// 時才恆為 0,不是本欄位的固定行為。
	s.Player.GovtBonusMoneyPercent = gamedata.IncomeGovtMoneyBonusPercent(s.Government)
	// 銀河貨幣交易所是 **Achievement 科技**(手冊:研究完成即生效,不必建造),所以在這裡由
	// 科技擁有狀況同步給 engine,而不是走建築表。判定規則與其他多選一主題一致。
	s.Player.HasGalacticCurrencyExchange = driveTechOwned(s.Player,
		gamedata.GalacticCurrencyExchangeTech.Topic, gamedata.GalacticCurrencyExchangeTech.Tech)
	// HyperAdvancedResearchCost 依這局遊戲的版本規則 profile 算好傳給 RunResearchPhase(engine 層
	// 不關心 RuleProfile 本身,只吃算好的數字,見 engine.PlayerState.HyperAdvancedResearchCost
	// 註解)。Profile15(現行預設)= 25000 = 套件級硬編值,no-op;Profile13 = 15000,真的改變
	// Hyper-Advanced 第一級 profile 基礎成本；RunResearchPhase 再依原版公式加 level×10000。
	s.Player.HyperAdvancedResearchCost = gamedata.HyperAdvancedCost(s.RuleProfile)
	s.syncTradeGoodsFlag()          // 依建造選單同步「貿易品」旗標,供 RunEmpireTurn 判斷是否換算收入
	s.syncRaceEngineFields()        // 種族布林特性 → 引擎層欄位,見 race_boolean_traits.go
	s.syncAchievementColonyFields() // 成就科技的全帝國效果(污染容忍、每工人產能),見 achievements.go
	s.recalcAllColonyMorale()       // 成就也影響士氣(VR 網路/心靈學),要在經濟結算之前重算
}

// EndTurn 推進一回合:先結算玩家帝國,再讓各 AI 對手自行決策並結算,回合數 +1。
func (s *GameSession) EndTurn() {
	// 狀態播報 29 要比較「本回合改變前／後」；第一次進入舊存檔時只建立基線，不補播歷史。
	s.ensureStatusEmpireBaseline()
	// 原版 sub_E4DC9 在後續回合消費 player+0xE72；必須早於本回合帝國經濟結算。
	s.resolvePendingSurrenders()
	// 協議狀態每個世界回合只推進一次；同一份結果稍後分別餵給玩家與對應 AI，
	// 避免在兩個帝國的經濟結算中把協議進度推進兩次。
	treatyYields := s.advanceTreaties()
	playerTreatyBC, playerTreatyResearch := 0, 0
	for _, y := range treatyYields {
		playerTreatyBC += y.PlayerBC
		playerTreatyResearch += y.PlayerResearch
	}
	s.Player.TreatyIncomeBC = playerTreatyBC
	s.Player.TreatyResearch = playerTreatyResearch
	s.preparePlayerResearchApplication()
	s.prepPlayerDerived()
	s.LastPlayerOutput = engine.RunEmpireTurnWithResearchRoller(s.Player, s.coloniesForTurn(), s.researchBreakthroughRoll)
	s.recordPlayerPlagueResearch(s.LastPlayerOutput)
	s.LastPlayerOutput.Player.TreatyIncomeBC = 0
	s.LastPlayerOutput.Player.TreatyResearch = 0
	s.Player = s.LastPlayerOutput.Player
	s.resolvePlayerBankruptcy()
	s.applyPlayerResearchRaceTrait(s.LastPlayerOutput.ResearchDone)
	if s.LastPlayerOutput.ResearchDone {
		applyResearchTopicGrantCallbacks(&s.Player, s.Player.ResearchTopic)
		s.UpdatePlayerShipDesignsAfterTech()
	}
	s.recoverFromFamine() // 饑荒防死鎖:見函式註解;依本回合 Starving 結果修正下回合職務分配
	// 原版先完成整張條約關係 pair loop，再完成整張目標漂移 loop；在 AI
	// 經濟／行動前集中執行，避免逐 AI 交錯而改變擲骰順序。
	s.advanceOriginalDiplomacyGrowth()
	aiGross := make([]int, len(s.AIPlayers))
	for i := range s.AIPlayers {
		applyResearchTopicGrantCallbacks(&s.AIPlayers[i].Player, gamedata.TOPIC_ASTRO_CONSTRUCTION)
		s.prepareAIResearchApplication(&s.AIPlayers[i])
		s.syncAIRaceEngineFields(&s.AIPlayers[i])
		syncAIColonyGovernmentOutput(&s.AIPlayers[i])
		s.syncAICommandPoints(i)
		// 分兩步而非直接呼叫 engine.RunAIEmpireTurn:ApplyAIEconomy 回傳的 colonies(職務
		// 重新分配後的結果)必須寫回 s.AIPlayers[i].Colonies——先前直接用 RunAIEmpireTurn 時,
		// 這個回傳值只在函式內部傳給 RunEmpireTurn 算完當回合經濟就丟棄,從未寫回存檔用的
		// AIOpponent.Colonies,導致存檔/未來 UI 若讀取 AI 殖民地職務分配會看到「從未更新」的
		// 初始值(雖然目前無 UI 讀取此欄位,經濟結算本身不受影響——因為每回合都是從同一組
		// 靜態 Population/FoodPerFarmer 重新算,但欄位本身是錯的,發現後順手修正)。
		// AI 對手與玩家共用同一局的版本規則 profile(RuleProfile 是整個 GameSession 唯讀設定,
		// 見該欄位註解),Hyper-Advanced 研究成本覆寫同樣要套用,否則 1.3 局裡 AI 仍會用 1.5 的
		// 25000 成本研究,造成玩家/AI 規則不對稱。
		s.AIPlayers[i].Player.HyperAdvancedResearchCost = gamedata.HyperAdvancedCost(s.RuleProfile)
		if i < len(treatyYields) {
			s.AIPlayers[i].Player.TreatyIncomeBC = treatyYields[i].AIBC
			s.AIPlayers[i].Player.TreatyResearch = treatyYields[i].AIResearch
		}
		ps := s.AIPlayers[i].Player
		jobCtx := engine.OriginalAIJobContext{
			Personality:         s.AIPlayers[i].Personality,
			LateTech:            aiOriginalLateTechReached(ps),
			ColonyFoodHalf:      make([]int, len(s.AIPlayers[i].Colonies)),
			ColonyFoodHalfKnown: make([]bool, len(s.AIPlayers[i].Colonies)),
			ColonyBlockaded:     make([]bool, len(s.AIPlayers[i].Colonies)),
		}
		knownTech := knownTechnologyApplications(ps)
		for colony := range s.AIPlayers[i].Colonies {
			var built map[string]bool
			if colony < len(s.AIPlayers[i].ColonyBuildings) {
				built = s.AIPlayers[i].ColonyBuildings[colony]
			}
			jobCtx.ColonyFoodHalf[colony], jobCtx.ColonyFoodHalfKnown[colony] =
				originalAIColonyFoodHalf(s.AIPlayers[i].Colonies[colony], built, knownTech)
			if colony < len(s.AIPlayers[i].ColonyStars) {
				star, slot := s.AIPlayers[i].ColonyStars[colony], s.AIPlayers[i].PopulationRaceSlot
				if star >= 0 && star < len(s.Stars) && s.AIPlayers[i].PopulationRaceSlotKnown && slot >= 0 && slot < 8 {
					jobCtx.ColonyBlockaded[colony] = s.Stars[star].BlockadedMask&(1<<slot) != 0
				}
			}
		}
		// sub_D66B3 在 sub_E2D72 重算後讀本回合帝國產出；因此職務選擇必須和
		// 最終結算看見同一份難度／事件暫態加成。回傳後只合併職務欄，避免把
		// GrowthBonusSum 等暫態值永久寫回並在下回合重複疊加。
		jobInput := s.aiColoniesForTurn(i, s.AIPlayers[i].Colonies)
		assigned, freighterPressure, exactJobs := engine.ApplyOriginalAIJobsWithTransport(ps, jobInput, jobCtx)
		colonies := mergeAIJobAssignments(s.AIPlayers[i].Colonies, assigned)
		if !exactJobs {
			// 舊 JSON 缺逐種族 profile 或 +0xDD 無法建立時，保留既有可玩 fallback；
			// 不把這條路徑宣稱為原版忠實。原版可執行檔沒有 AI 回合寫入 player+0x31
			// 稅率的 producer，因此 fallback 只能代理職務，不能順便套用 remake 的國庫門檻調稅。
			originalTaxRate := ps.TaxRate
			ps, colonies = engine.ApplyAIEconomy(ps, s.AIPlayers[i].Colonies, s.AIPlayers[i].Decider)
			ps.TaxRate = originalTaxRate
		} else if freighterPressure {
			// sub_D6AD4 只有在運輸壓力旗標成立時才呼叫 Random(10)；不可在無壓力時
			// 預先求值 eventRoll，否則會改變後續事件的確定性亂數序列。
			if gain, ok := gamedata.OriginalAIFreighterFleetGain(
				true, s.Difficulty, s.eventRoll(10)); ok {
				// 原版不是把這批貨運艦排入一般建築產品表；新增後從下一回合重算運輸。
				ps.ActiveFreighters += gain
			}
		}
		ps = applyAIDifficultyPlayerInputs(ps, s.Difficulty)
		s.AIPlayers[i].Colonies = colonies
		out := engine.RunEmpireTurnWithResearchRoller(ps, s.aiColoniesForTurn(i, colonies), s.researchBreakthroughRoll)
		if turns, ok := gamedata.OriginalNPCFoodDeficitTurns(s.AIPlayers[i].OriginalFoodDeficitTurns, out.TotalFoodHalf); ok {
			s.AIPlayers[i].OriginalFoodDeficitTurns = turns
		}
		s.recordAIPlagueResearch(i, out)
		out.Player.TreatyIncomeBC = 0
		out.Player.TreatyResearch = 0
		s.AIPlayers[i].Player = out.Player
		advanceAIColonyPopulation(&s.AIPlayers[i], out, s.activePopulationRaceSlots(), s.populationRandForTurn())
		aiGross[i] = empireGrossBC(out)
		s.applyAIResearchRaceTrait(&s.AIPlayers[i], out.ResearchDone)
		if out.ResearchDone {
			applyResearchTopicGrantCallbacks(&s.AIPlayers[i].Player, s.AIPlayers[i].Player.ResearchTopic)
			s.updateAIShipDesignsAfterTech(i)
		}
		// AI 人口／職務在 ApplyAIEconomy 後才是本回合的新值；兵營容量與每五回合
		// 產出要吃這個回寫後的人口，再交給 advanceAI 進行擴張等行動。
		advanceAIGroundForces(&s.AIPlayers[i])
		s.advanceAI(i, out) // AI 主動行為:造艦 / 擴張 / 外交態勢
	}
	// 原版把納貢成本納入帝國經濟後才供外交／摘要讀取；這裡在雙方
	// RunEmpireTurn 完成後一次移轉，避免玩家與 AI 看到半回合狀態。
	// AI 的下一步決策從下一回合開始使用收到的國庫，維持現有回合順序。
	s.applyTributeTransfers(empireGrossBC(s.LastPlayerOutput), aiGross)
	// 間諜結算須排在玩家與所有 AI 本回合研究都跑完之後(用最新的 CompletedTopics/ChosenTech
	// 判定「對方已知、我方未知」的可偷科技清單),故緊接在上面的 AI 迴圈之後。
	bcBeforeLeaderEffects := s.Player.BC
	s.advanceLeaderLimbo() // 原版 status=4 閒置記數達 30 後清除(見 leader_tenure.go)
	s.LastLeaderUpkeep = s.LastPlayerOutput.OfficerMaintenanceCost
	if wealth := leaderMegawealthBC(s.Leaders); wealth != 0 {
		s.Player.BC += wealth
	}
	// 領袖維護費與 Megawealth 是帝國層現金變化；回填摘要，讓 NetBC 與
	// 實際國庫保持一致。沒有領袖時差值為 0，維持舊開局序列。
	s.LastPlayerOutput.NetBC += s.Player.BC - bcBeforeLeaderEffects
	s.LastPlayerOutput.Player.BC = s.Player.BC
	bcBeforeEspionage := s.Player.BC
	s.advanceEspionage() // 玩家 ↔ AI 間諜行動(最小迴圈:偷科技 STEAL,見 spy.go)
	s.LastPlayerOutput.NetBC += s.Player.BC - bcBeforeEspionage
	s.LastPlayerOutput.Player.BC = s.Player.BC
	s.advanceBuilds()         // 以本回合淨工業推進各殖民地建造
	s.advanceResearch()       // 目前研究主題完成則自動推進到下一個未完成的元件解鎖主題
	s.LastDiscovery = nil     // 每回合先清掉上一回合的發現(與 advanceEvents 清 LastEvent 同一個節奏)
	s.advanceFleet()          // 推進艦隊星間航行(ETA 遞減,抵達則標記探索 + 結算一次性發現)
	s.advanceCrewExperience() // 艦員經驗:每回合 +1,停泊星系每有一座太空學院再 +1(見 crew.go)
	s.advanceAssimilation()   // 征服人口同化:依政體 2–20 回合同化一單位(見 assimilation.go)
	// 叛亂檢定接在同化**之後**:同化先扣掉這一回合該同化的人口,剩下的才是有機會起事的。
	// 反過來的話,一個「這回合剛好同化完最後一單位」的殖民地還會多擲一次骰。
	s.LastRebellions = s.advanceRebellions() // 未同化人口叛亂(見 rebellion.go)
	s.queueRebellionBroadcasts(s.LastRebellions)
	// AI 艦隊航行:先推進位置,advanceAIRaids 才判「有沒有停在玩家殖民地上」。
	// 順序不能反——反了會讓艦隊在抵達的同一回合就開打,少掉一回合的預警。
	s.LastAIArrivals = s.advanceAIFleets() // 見 ai_fleet.go
	s.advanceMarines()                     // 各 Marine Barracks 殖民地依手冊公式補充陸戰隊駐軍(有上限)
	s.advanceArmor()                       // 各 Armor Barracks 殖民地依手冊公式補充戰車營駐軍(有上限,見 ground_invasion.go)
	s.advancePopulation()                  // 累積各殖民地成長,達門檻則 +1 人口(回寫 Population)
	s.advanceEvents()                      // 觸發 MOO2 風格隨機事件(繁榮/瘟疫/海盜…),記於 LastEvent
	// 持續型事件(超新星倒數/時空異象/超空間獸)每回合推進一次,見 events_persistent.go。
	// 它們的訊息接在 LastEvent 後面——一回合可能同時有「新抽到的事件」與「持續中的狀態」。
	if msgs := s.advancePersistentEvents(); len(msgs) > 0 {
		for _, m := range msgs {
			if s.LastEvent == "" {
				s.LastEvent = m
			} else {
				s.LastEvent += "|" + m
			}
		}
	}
	if births := s.advanceSpaceEelSplits(); births > 0 {
		message := fmt.Sprintf("太空鰻完成分裂，銀河中新增 %d 艘太空鰻", births)
		if s.LastEvent == "" {
			s.LastEvent = message
		} else {
			s.LastEvent += "|" + message
		}
	}
	// 原版先推進停泊怪物 consumer，再由 Move_All_Ships_ 推進航行；因此剛抵達的太空鰻
	// age 維持 0，不會在抵達同回合又加一。
	for _, message := range s.advanceEventMonsterRoutes() {
		if s.LastEvent == "" {
			s.LastEvent = message
		} else {
			s.LastEvent += "|" + message
		}
	}
	s.Turn++
	s.advanceShipRepair()          // 停在自家據點的艦艇完全修復(原版 Repair_Ships_At_Colonies_)
	s.advanceAntares()             // 安塔蘭人週期性入侵(依 Turn 排程升級),記於 LastAntaranNotice
	s.advanceAIRaids()             // AI 對手突襲玩家殖民地(戰爭態勢 + 軍力領先才發動),記於 LastRaidReport
	s.recomputeOriginalBlockades() // 本回合艦隊移動／戰鬥後重建，供下一回合殖民地 AI 消費。
	s.advanceConquestVictory()     // 對手是否已全滅(手冊三條勝利路徑之一:殲滅所有對手)
	s.advancePlayerDefeat()        // 玩家是否已無任何殖民地(超新星等事件可致,見該函式)
	s.advanceAntaranVictory()      // 是否已攻陷安塔蘭母星(手冊三條勝利路徑之二,見 antaran_victory.go)
	s.detectEmpireEliminationBroadcasts()
	s.publishNextStatusBroadcast()
	s.advanceAIDiplomacy() // 由本回合原版關係結果推進可選 AI↔AI 政策／戰爭 consumer
	s.advanceAISurrenders()
	s.publishNextStatusBroadcast()
	s.advanceCouncil()    // 銀河議會選舉(手冊三條勝利路徑之一:2/3 多數當選銀河領袖),記於 LastCouncilNotice
	s.advanceMercOffers() // 傭兵領袖不定期上門(手冊 p.134),補進 MercPool 供玩家在軍官畫面雇用
	s.recordHistory()     // 全帝國國力快照(原版 module 122 Record_History_),供 INFO 歷史圖表
	// 熱座:其餘真人席位的帝國也要各自過完這一回合,否則他們的殖民地會被凍結
	// (見 hotseat.go advanceIdleSeats,含各席位不對稱之處的說明)。
	// 單人局 HotseatEnabled() 為 false,這裡是 no-op,既有行為不變。
	if s.HotseatEnabled() {
		s.advanceIdleSeats()
	}
}

func mergeAIJobAssignments(base, assigned []engine.ColonyState) []engine.ColonyState {
	if len(base) != len(assigned) {
		return base
	}
	out := append([]engine.ColonyState(nil), base...)
	for i := range out {
		out[i].Farmers = assigned[i].Farmers
		out[i].Workers = assigned[i].Workers
		out[i].Scientists = assigned[i].Scientists
		out[i].PopulationGroups = append([]engine.PopulationGroup(nil), assigned[i].PopulationGroups...)
	}
	return out
}

func advanceAIColonyPopulation(a *AIOpponent, out engine.EmpireOutput, activeSlots int, rng *randStream) {
	if a == nil {
		return
	}
	for ci := range a.Colonies {
		if ci >= len(out.Colonies) || out.Colonies[ci].PopulationGroupGrowthCount != len(a.Colonies[ci].PopulationGroups) ||
			!engine.PopulationGroupsComplete(a.Colonies[ci]) {
			continue
		}
		for gi := range a.Colonies[ci].PopulationGroups {
			a.Colonies[ci].PopulationGroups[gi].GrowthPoints += out.Colonies[ci].PopulationGroupGrowth[gi]
		}
		protected := protectedPopulationGroup(a.Colonies[ci])
		// 原版先完成所有負池刪除，再進 owner-first 的正池 pass。
		for slot := 0; slot < gamedata.PopulationRaceSlots; slot++ {
			gi := populationGroupIndexBySlot(a.Colonies[ci], slot)
			if gi < 0 {
				continue
			}
			minimum := 0
			if gi == protected {
				minimum = 1
			}
			for a.Colonies[ci].PopulationGroups[gi].GrowthPoints < 0 {
				if populationGroupUnits(a.Colonies[ci].PopulationGroups[gi]) <= minimum ||
					!removeNegativeGrowthColonist(&a.Colonies[ci], gi, rng) {
					a.Colonies[ci].PopulationGroups[gi].GrowthPoints = 0
					break
				}
				a.Colonies[ci].PopulationGroups[gi].GrowthPoints += gamedata.PopulationGrowthPointsPerUnit
			}
		}
		owner := -1
		if a.Colonies[ci].OwnerRaceSlotKnown {
			owner = a.Colonies[ci].OwnerRaceSlot
		}
		for _, slot := range positivePopulationSlotOrder(owner, activeSlots, rng) {
			gi := populationGroupIndexBySlot(a.Colonies[ci], slot)
			if gi < 0 {
				continue
			}
			limit := engine.PopulationGroupLimit(a.Colonies[ci], gi)
			for a.Colonies[ci].PopulationGroups[gi].GrowthPoints >= gamedata.PopulationGrowthPointsPerUnit &&
				a.Colonies[ci].Population < limit {
				a.Colonies[ci].PopulationGroups[gi].GrowthPoints -= gamedata.PopulationGrowthPointsPerUnit
				a.Colonies[ci].Population++
				cand := a.Colonies[ci]
				cand.Workers++
				engine.AddPopulationGroupUnit(&cand, gi, gamedata.WORKER)
				if engine.RunColonyTurn(cand).FoodSurplus < 0 {
					cand = a.Colonies[ci]
					cand.Farmers++
					engine.AddPopulationGroupUnit(&cand, gi, gamedata.FARMER)
				}
				a.Colonies[ci] = cand
			}
			if a.Colonies[ci].PopulationGroups[gi].GrowthPoints >= gamedata.PopulationGrowthPointsPerUnit {
				a.Colonies[ci].PopulationGroups[gi].GrowthPoints = gamedata.PopulationGrowthPointsPerUnit - 1
			}
		}
	}
}

// advancePopulation 把各殖民地本回合成長率(LastPlayerOutput.Colonies[i].PopGrowth)累加到
// popAccum,達門檻則 +1 人口(回寫 Population,新單位預設為工人),受 PopMax 上限。
func (s *GameSession) advancePopulation() {
	if s.popAccum == nil {
		s.popAccum = make([]int, len(s.PlayerColonies))
	}
	for i := range s.PlayerColonies {
		if i >= len(s.LastPlayerOutput.Colonies) || i >= len(s.popAccum) {
			break
		}
		out := s.LastPlayerOutput.Colonies[i]
		if out.PopulationGroupGrowthCount == len(s.PlayerColonies[i].PopulationGroups) &&
			engine.PopulationGroupsComplete(s.PlayerColonies[i]) {
			for gi := range s.PlayerColonies[i].PopulationGroups {
				s.PlayerColonies[i].PopulationGroups[gi].GrowthPoints += out.PopulationGroupGrowth[gi]
			}
			protected := protectedPopulationGroup(s.PlayerColonies[i])
			for slot := 0; slot < gamedata.PopulationRaceSlots; slot++ {
				gi := populationGroupIndexBySlot(s.PlayerColonies[i], slot)
				if gi < 0 {
					continue
				}
				minimum := 0
				if gi == protected {
					minimum = 1
				}
				for s.PlayerColonies[i].PopulationGroups[gi].GrowthPoints < 0 {
					if populationGroupUnits(s.PlayerColonies[i].PopulationGroups[gi]) <= minimum ||
						!removeNegativeGrowthColonist(&s.PlayerColonies[i], gi, s.populationRandForTurn()) {
						s.PlayerColonies[i].PopulationGroups[gi].GrowthPoints = 0
						break
					}
					s.PlayerColonies[i].PopulationGroups[gi].GrowthPoints += gamedata.PopulationGrowthPointsPerUnit
				}
			}
			owner := -1
			if s.PlayerColonies[i].OwnerRaceSlotKnown {
				owner = s.PlayerColonies[i].OwnerRaceSlot
			}
			for _, slot := range positivePopulationSlotOrder(owner, s.activePopulationRaceSlots(), s.populationRandForTurn()) {
				gi := populationGroupIndexBySlot(s.PlayerColonies[i], slot)
				if gi < 0 {
					continue
				}
				limit := engine.PopulationGroupLimit(s.PlayerColonies[i], gi)
				for s.PlayerColonies[i].PopulationGroups[gi].GrowthPoints >= gamedata.PopulationGrowthPointsPerUnit &&
					s.PlayerColonies[i].Population < limit {
					s.PlayerColonies[i].PopulationGroups[gi].GrowthPoints -= gamedata.PopulationGrowthPointsPerUnit
					s.PlayerColonies[i].Population++
					s.assignNewColonistFromGroup(i, gi)
				}
				if s.PlayerColonies[i].PopulationGroups[gi].GrowthPoints >= gamedata.PopulationGrowthPointsPerUnit {
					s.PlayerColonies[i].PopulationGroups[gi].GrowthPoints = gamedata.PopulationGrowthPointsPerUnit - 1
				}
			}
			continue
		}
		grow := out.PopGrowth
		grow += grow * s.raceGrowthPct / 100 // 舊 JSON fallback：群組不完整時沿用 owner 種族加成。
		s.popAccum[i] += grow
		for s.popAccum[i] >= gamedata.PopulationGrowthPointsPerUnit && s.PlayerColonies[i].Population < s.PlayerColonies[i].PopMax {
			s.popAccum[i] -= gamedata.PopulationGrowthPointsPerUnit
			s.PlayerColonies[i].Population++
			s.assignNewColonist(i)
		}
	}
}

// assignNewColonist 決定新增人口的職務(殖民地 i 的 Population 已 +1,尚未配職)。
//
// 原版行為依據(A 級硬證):Orion2.exe 除錯符號表有 Make_First_Unassigned_Into_Farmer_ 與
// Assign_Additional_Unblockaded_Farmers_(見 docs/re/01-gap-report.md),即原版會在需要時
// 把新增/未指派人口指派成農夫。remake 先前一律 Workers++,導致人口一成長就缺糧 →
// 餓死螺旋(30 回合探針:pop 8→9→…→5、食物 -1)。
//
// ⚠ 精確的原版判斷式尚未反編(函式邊界問題,見 docs/re/00-orion2-symbols.md),此處採
// 函式名直接蘊含的最小忠實規則:**先試工人,若會造成食物赤字就改配農夫**。
// 這不是臆造機制(機制存在是硬證),只是判斷門檻取「不讓殖民地挨餓」這個保守值。
func (s *GameSession) assignNewColonist(i int) {
	group := -1
	if i >= 0 && i < len(s.PlayerColonies) && s.PlayerColonies[i].OwnerRaceSlotKnown {
		for gi, g := range s.PlayerColonies[i].PopulationGroups {
			if g.RaceSlotKnown && g.RaceSlot == s.PlayerColonies[i].OwnerRaceSlot {
				group = gi
				break
			}
		}
	}
	s.assignNewColonistFromGroup(i, group)
}

func (s *GameSession) assignNewColonistFromGroup(i, group int) {
	if i < 0 || i >= len(s.PlayerColonies) {
		return
	}
	cand := s.PlayerColonies[i]
	cand.Workers++ // 原版預設:新人口進生產線
	if group >= 0 {
		engine.AddPopulationGroupUnit(&cand, group, gamedata.WORKER)
	} else {
		engine.AddOwnerPopulationGroupUnit(&cand, gamedata.WORKER)
	}
	if engine.RunColonyTurn(cand).FoodSurplus >= 0 {
		s.PlayerColonies[i] = cand
		return
	}
	cand = s.PlayerColonies[i]
	cand.Farmers++ // 會缺糧 → 改配農夫(Assign_Additional_Unblockaded_Farmers_)
	if group >= 0 {
		engine.AddPopulationGroupUnit(&cand, group, gamedata.FARMER)
	} else {
		engine.AddOwnerPopulationGroupUnit(&cand, gamedata.FARMER)
	}
	s.PlayerColonies[i] = cand
}

// aiProfile 取出 AI 對手的性格(從 RemakeDecider);非該型別則回平衡型。
func aiProfile(a AIOpponent) ai.Profile {
	if rd, ok := a.Decider.(*ai.RemakeDecider); ok {
		return rd.Profile
	}
	return ai.ProfileBalanced
}

// playerMilitary 回傳玩家目前艦隊總戰力(供 AI 態勢比較)。
func (s *GameSession) playerMilitary() int {
	m := 0
	for _, sh := range s.AllShips() { // 國力是**全帝國**的
		m += shipStrength(sh.Class)
	}
	return m
}

// advanceAI 推進第 i 個 AI 對手的主動行為(每回合,經濟結算後):
//  1. 生產:逐殖民地先推進自己的建築產品；沒有可建建築的產能才進造艦轉接層。
//  2. 擴張:每隔數回合佔領一顆無主星(Owner=2,OwnedStars++)。
//  3. 研究:替 AI 處理待決的科技抉擇,並在目前主題完成時挑下一個(見 ai_research.go)。
//  4. 外交態勢:消費本回合已由原版 Diplomacy_Growth 規則更新的關係分數，
//     經 ai.DecideStance 推得態勢並保存顯示名稱。
func (s *GameSession) advanceAI(i int, out engine.EmpireOutput) {
	a := &s.AIPlayers[i]
	prof := aiProfile(*a)

	// 1) 造艦：工業投入持久造艦進度，完成後交付引用 AI 自己藍圖的實艦。
	//
	// FleetInvestPool 是餘數池,修正既有整數捨去 bug:直接算 TotalNetIndustry/invest 時,
	// 若 TotalNetIndustry(如忠實 yield 下常見的 3)小於 invest(Scientific 性格為 4),
	// 整數除法每回合都捨去成 0,FleetStrength 永久停滯(見 playerHomeworldColony 上方歷史記錄註解/
	// docs/tech/colony-economy-maintenance.md)。改成先把 NetIndustry 存進池子、池子夠 invest
	// 才兌現軍力、餘數留到下回合累積,小額淨工業也能跨回合逐步兌現,不會卡死。
	shipProduction := s.advanceAIColonyBuilds(i, out)
	if shipProduction > 0 {
		// EmpireOutput 的 NetIndustry 已是扣維護後可投入生產的點數；藍圖成本同樣以
		// 生產點計價，不再先縮成抽象軍力單位。偏工業性格額外投入一倍，屬既有 AI
		// 性格權重的 remake 轉接，不是原版精確 build selector。
		production := shipProduction
		if prof.IndustryWeight > prof.ResearchWeight {
			production *= 2
		}
		s.advanceAIShipProduction(i, production)
	}

	// 2) 殖民：All_AI_Colonize_ 每回合掃描已抵達來源；只在殖民船目前所在星系
	// 建立殖民地。跨星系航程由 advanceAIFleets 的單主力艦隊 adapter 規劃。
	s.aiExpand(i)

	// 2.7) 研究:處理待決的科技抉擇 + 目前主題完成時挑下一個(見 ai_research.go)。
	//
	// 先前完全沒有這一步——AI 每回合把研究點投進一個早就完成的主題,無限重複完成同一項,
	// 科技線整條靠間諜從玩家那裡偷。
	s.advanceAIResearch(i, out.TotalResearch)

	// 2.5) 間諜:AI 用最簡單的週期政策每 6 回合訓練 1 名間諜派來偷玩家科技(見 spy.go
	// advanceEspionage),上限比照手冊每對手 63 人(gamedata.SpySlotBonus 的夾範圍)。不像
	// 玩家 TrainSpy 需要花 BC——AI 訓練成本/BC 限制目前無資料可推導,誠實簡化為免費週期政策
	// (TODO:待有更細緻 AI 經濟模型後補上維護費/訓練成本)。
	if s.Turn%6 == 0 && a.Spies < spyMaxSlots {
		a.Spies++
	}
	// 防守 Agent 不再由免費週期計時器生成；原版 raw -7 是殖民地 100 PP 產品，
	// 由 advanceAIColonyBuilds 保存進度並在完工後加入 self Agent pool。

	// 3) 外交態勢：關係已在本回合 AI loop 前依 Diplomacy_Growth_ 更新。
	prevStance := a.StanceName
	stance := ai.DecideStance(diplomacy.RelationLevelForScore(a.Relation), prof)
	// 正式戰爭欄位是原版權威狀態；-90 raw 投影為 -36，不能因 remake
	// 的 -40..40 顯示尺而降級成「敵視」。
	if a.Treaty.FormalPolicy >= gamedata.DIPLO_LIMITED_WAR {
		stance = ai.StanceWar
	}
	a.StanceName = stanceNames[stance]
	// 態勢改變 = 這位對手有話要說(宣戰要通知、提議要開口)。見 audience.go。
	a.noteStanceChange(prevStance)
}

// ensureAIRelations 確保 AIRelations 矩陣尺寸 = len(AIPlayers)(懶初始化;新建/讀舊檔皆補齊,
// 保留已有值)。
func (s *GameSession) ensureAIRelations() {
	n := len(s.AIPlayers)
	if len(s.AIRelations) == n {
		return
	}
	m := make([][]int, n)
	for i := range m {
		m[i] = make([]int, n)
		if i < len(s.AIRelations) {
			copy(m[i], s.AIRelations[i])
		}
	}
	s.AIRelations = m
}

// advanceAIDiplomacy 只把本回合已由原版 Diplomacy_Growth_ 更新的 AI↔AI
// 關係交給現有可選政策／戰爭模型。舊軍力差關係漂移已移除。
func (s *GameSession) advanceAIDiplomacy() {
	s.ensureAIRelations()
	if s.EnableAIVsAI {
		s.advanceAIAIDiplomacy()
	}
}

// stanceNames 是 ai.Stance 的中文顯示。
var stanceNames = map[ai.Stance]string{
	ai.StanceWar:             "宣戰",
	ai.StanceHostile:         "敵視",
	ai.StanceNeutral:         "中立",
	ai.StanceProposeTrade:    "提議貿易",
	ai.StanceProposeAlliance: "提議結盟",
}

// aiExpand 讓第 i 個 AI 佔領一顆無主星:標 Star.Owner=2、OwnedStars++,並用
// newColonyFromStar(colonization.go,與玩家 ColonizeStar 共用同一套建法)建立真正的
// engine.ColonyState,append 進 AIOpponent.Colonies + ColonyStars + ColonyBuildings(三者是
// AIOpponent 的殖民地平行陣列,長度須恆等。AI 目前產品以 ColonyBuilds[star] 保存，故不另加
// 一條會因殖民地移除而錯位的平行陣列；駐軍與人口仍各有既有同步欄位。新殖民地的
// ColonyBuildings 項 append 空 map(手冊只保證母星有星基,新拓殖星沒有)。
//
// 2026-07-11 訂正:先前只設旗標、不建殖民地模型(見 AIOpponent.ColonyStars 欄位註解),
// 導致 AI 版圖擴張後經濟(EndTurn 的 RunEmpireTurn(ps, a.Colonies))永遠停在初始母星產出,
// 不會隨佔領星數成長。現在下回合 EndTurn 會把新殖民地的淨工業交給逐殖民地產品；有可建
// 設施時先累積建築，沒有候選才交給造艦轉接層，AI 仍會隨擴張增加總生產投入。
//
// gov 傳 gamedata.MoraleGovDictatorship(AIOpponent 沒有 Government 欄位,政府型態未建模,
// 見 newColonyFromPlanet 註解)。AI 已由 RaceIndex 補套布林種族特性，但數值型食物／工業／
// 研究加成仍未傳入共用建構器。若該星系沒有可殖民的
// 天體(全是氣態巨星/小行星帶,或氣候不合)就 continue 找下一顆無主星,不 fallback 成只設旗標
// (避免旗標與殖民地模型再度分裂)。找不到任何可擴張的無主星則整個 no-op。
func (s *GameSession) aiExpand(i int) {
	if i < 0 || i >= len(s.AIPlayers) {
		return
	}
	currentStar := aiFleetStar(s.AIPlayers[i])
	colonyBase := -1
	for colony, star := range s.AIPlayers[i].ColonyStars {
		if star == currentStar && colony < len(s.AIPlayers[i].ColonyBuildings) &&
			s.AIPlayers[i].ColonyBuildings[colony][ColonyBaseBuildName] {
			colonyBase = colony
			break
		}
	}
	colonyShip := -1
	for ship := range s.AIPlayers[i].Ships {
		candidate := s.AIPlayers[i].Ships[ship]
		if (candidate.RawTypeKnown && candidate.RawType == gamedata.COLONY_SHIP) || candidate.Class == ColonyShipClass {
			colonyShip = ship
			break
		}
	}
	// All_AI_Colonize_ → sub_E65F8 只處理實際位於 AI 艦隊／軍官記錄中的殖民船；
	// 沒有來源時不會依固定週期免費建立殖民地。
	if colonyBase < 0 && colonyShip < 0 {
		return
	}
	if s.AIPlayers[i].FleetETA > 0 {
		return
	}
	if !s.aiCanExpandInto(i, currentStar) {
		return
	}
	ensureAIGroundForceSlots(&s.AIPlayers[i])
	// sub_E65F8 只在來源所在星系的五個行星槽中挑目標，不會讓殖民船瞬移到全圖最佳星。
	order := []int{currentStar}
	for _, idx := range order {
		if !s.aiCanExpandInto(i, idx) {
			continue
		}
		if s.StarGuardedByMonster(idx) {
			continue // 怪獸盤據的星系 AI 也進不去(手冊 p.62 的清場條件對所有帝國一體適用)
		}
		// AI 也是「殖民到行星上」——挑該星系第一顆還沒被任何帝國佔走的可殖民天體
		// (同玩家的 ColonizeStar)。
		planetIdx, _ := s.bestAIColonizablePlanet(i, idx)
		if planetIdx < 0 {
			continue
		}
		aiRace := aiRaceIndex(s.AIPlayers[i])
		foodBonus, indBonus, resBonus := 0, 0, 0
		if aiRace >= 0 && aiRace < len(Races) {
			foodBonus, indBonus, resBonus = Races[aiRace].FoodBonus, Races[aiRace].IndBonus, Races[aiRace].ResBonus
		}
		colony, ok, _ := s.newColonyFromPlanet(planetIdx, gamedata.MoraleGovDictatorship, foodBonus, indBonus, resBonus)
		if !ok {
			continue
		}
		// newColonyFromPlanet 同時服務玩家與 AI;此處把共用建構器先帶入的玩家特性
		// 改成該 AI 的原版種族特性,避免 AI 擴張殖民地誤套玩家種族。
		if aiRace >= 0 && aiRace < len(Races) {
			orig := Races[aiRace].OrigIdx
			colony.TolerantRace = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_TOLERANT)
			colony.Lithovore = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_LITHOVORE)
			colony.Aquatic = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_AQUATIC)
			colony.Subterranean = gamedata.OrigRaceHasTrait(orig, gamedata.TRAIT_SUBTERRANEAN)
			colony.FoodPerFarmer += gamedata.ClimateFoodPerFarmer(raceFoodClimate(colony.Climate, colony.Aquatic)) -
				gamedata.ClimateFoodPerFarmer(colony.Climate)
			colony.PopMax = racePopulationMax(colony.PlanetSize, colony.Climate, colony.Aquatic, colony.TolerantRace, colony.Subterranean)
			if colony.PopMax < colony.Population {
				colony.PopMax = colony.Population
			}
		} else {
			colony.TolerantRace = false
			colony.Lithovore = false
			colony.Aquatic = false
			colony.Subterranean = false
		}
		// 共用建構器先建立玩家 slot 0 群組；AI 擴張要由自己的 slot/profile 重建。
		colony.PopulationGroups = nil
		colony.OwnerRaceSlotKnown = false
		if s.Stars[idx].Owner == 0 {
			// 只有「本來無主」才算多佔一顆星。在自己已有的星系裡再殖民一顆行星不會讓
			// 版圖變大——OwnedStars 若跟著加,征服勝利的判定與外交評分都會被灌水。
			s.Stars[idx].Owner = 2
			s.AIPlayers[i].OwnedStars++
		}
		s.AIPlayers[i].Colonies = append(s.AIPlayers[i].Colonies, colony)
		s.AIPlayers[i].ColonyStars = append(s.AIPlayers[i].ColonyStars, idx)
		// sub_5E55F @ 0x5E766 在建立新殖民地時寫 colony+0x141=1；raw 11 Colony Base
		// 是可由後續同星系殖民消耗的一次性來源，不是空建築 map。
		s.AIPlayers[i].ColonyBuildings = append(s.AIPlayers[i].ColonyBuildings,
			map[string]bool{ColonyBaseBuildName: true})
		s.AIPlayers[i].ColonyPlanets = append(s.AIPlayers[i].ColonyPlanets, planetIdx)
		s.AIPlayers[i].ColonyMarines = append(s.AIPlayers[i].ColonyMarines, 0)
		s.AIPlayers[i].ColonyTanks = append(s.AIPlayers[i].ColonyTanks, 0)
		s.AIPlayers[i].MarineBarracksAge = append(s.AIPlayers[i].MarineBarracksAge, 0)
		s.AIPlayers[i].ArmorBarracksAge = append(s.AIPlayers[i].ArmorBarracksAge, 0)
		if colonyBase >= 0 {
			delete(s.AIPlayers[i].ColonyBuildings[colonyBase], ColonyBaseBuildName)
		} else {
			s.AIPlayers[i].Ships = append(s.AIPlayers[i].Ships[:colonyShip], s.AIPlayers[i].Ships[colonyShip+1:]...)
			s.syncAIShipStrength(i)
		}
		s.syncAIRaceEngineFields(&s.AIPlayers[i])
		s.consumeSpecialOnColonize(planetIdx) // 原住民被 AI 併入人口後同樣從行星上消失(見 colonization.go)
		return
	}
}

// aiCanExpandInto 回傳第 i 個 AI 能不能往 starIdx 這顆星拓殖:無主的可以,
// **自己已經有殖民地的星系也可以**(同星系多殖民地,手冊 p.61 的條件是那顆行星沒被殖民)。
// 別人的星系(玩家的、或另一個 AI 的)不行——那要打下來。
func (s *GameSession) aiCanExpandInto(i, starIdx int) bool {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return false
	}
	if s.Stars[starIdx].Owner == 0 {
		return true
	}
	if s.Stars[starIdx].Owner != 2 {
		return false // 玩家的星系
	}
	// Star.Owner 只分「無主/玩家/AI」,分不出是哪一個 AI——要查這個 AI 自己的殖民地清單。
	for _, st := range s.AIPlayers[i].ColonyStars {
		if st == starIdx {
			return true
		}
	}
	return false
}

// aiExpansionCandidates 列出第 i 個 AI 這一輪可以考慮的拓殖目標星。
//
// 自己的星系排在無主星之後:同樣分數時先往外擴,與原版 AI「先搶地盤」的行為一致
// (而且回內部補行星的機會永遠都在,無主星被別人搶走就沒了)。
func (s *GameSession) aiExpansionCandidates(i int) []int {
	out := make([]int, 0, len(s.Stars))
	for idx := range s.Stars {
		if s.Stars[idx].Owner == 0 {
			out = append(out, idx)
		}
	}
	seen := make(map[int]bool, len(s.AIPlayers[i].ColonyStars))
	for _, st := range s.AIPlayers[i].ColonyStars {
		if st < 0 || st >= len(s.Stars) || seen[st] {
			continue
		}
		seen[st] = true
		out = append(out, st)
	}
	return out
}

// syncAIColonyPlanets 把每個 AI 殖民地的行星索引補齊(見 AIOpponent.ColonyPlanets)。
//
// buildDemoAIOpponents 建 AI 時手上沒有 Planets(它只拿到母星的星索引),所以行星索引在
// 星系與行星都生完之後才補。舊存檔載入時同樣走這裡把 nil 補成真值。
func (s *GameSession) syncAIColonyPlanets() {
	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		for len(a.ColonyPlanets) < len(a.ColonyStars) {
			a.ColonyPlanets = append(a.ColonyPlanets, -1)
		}
		for j, star := range a.ColonyStars {
			if j < len(a.ColonyPlanets) && a.ColonyPlanets[j] >= 0 {
				continue
			}
			a.ColonyPlanets[j] = s.PlanetAt(star)
		}
	}
}

// aiPlanetValue 算某顆星對 AI i 的跨星航線 contextual 價值；逐行星基礎公式由
// aiBasePlanetValueAt 對映 Uncolonized_Planet_Worth_To_Player_，兩層不可混用。
//
// AI 的目標傾向由性格決定:冷酷/排外偏礦產(工業=軍力),和平主義偏人口。
// 原版的目標傾向是獨立於性格的另一個維度(玩家結構偏移 2208),remake 沒有那一層,
// 用性格代打——這是 remake 的映射,不是原版對照。
func (s *GameSession) aiPlanetValue(aiIdx, starIdx int) int {
	if starIdx < 0 || starIdx >= len(s.Stars) || s.StarGuardedByMonster(starIdx) {
		return 0
	}
	best := 0
	for _, planet := range s.PlanetsAt(starIdx) {
		if s.PlanetColonized(planet) {
			continue
		}
		if value := s.aiPlanetValueAt(aiIdx, starIdx, planet); value > best {
			best = value
		}
	}
	return best
}

// bestAIColonizablePlanet 對應 sub_E65F8 從軌道 4 反掃到 0、只在嚴格較高時替換；
// 同分因此保留較高軌道，不再把 FirstColonizablePlanet 當成「最佳」行星。
func (s *GameSession) bestAIColonizablePlanet(aiIdx, starIdx int) (int, int) {
	if starIdx < 0 || starIdx >= len(s.Stars) || s.StarGuardedByMonster(starIdx) {
		return -1, 0
	}
	best, bestValue := -1, 0
	planets := s.PlanetsAt(starIdx)
	for orbit := len(planets) - 1; orbit >= 0; orbit-- {
		planet := planets[orbit]
		if s.PlanetColonized(planet) {
			continue
		}
		if value := s.aiBasePlanetValueAt(aiIdx, planet); value > bestValue {
			best, bestValue = planet, value
		}
	}
	return best, bestValue
}

func (s *GameSession) aiPlanetValueAt(aiIdx, starIdx, planetIdx int) int {
	if starIdx < 0 || starIdx >= len(s.Stars) || planetIdx < 0 || planetIdx >= len(s.Planets) {
		return 0
	}
	p := s.Planets[planetIdx]
	base := s.aiBasePlanetValueAt(aiIdx, planetIdx)
	if base <= 0 {
		return 0
	}
	obj := s.aiPlanetObjective(aiIdx)
	// 第二、三層仍供跨星航線規劃；sub_E65F8 的同星系五軌道選擇只讀上方 base。
	ctx := s.aiNeighborhood(aiIdx, starIdx, obj)
	ctx.Base = base + gamedata.AIProximityValue(ctx.distances)
	ctx.Size = p.SizeID
	ctx.Colonized = s.Stars[starIdx].Owner != 0
	return gamedata.AIContextualPlanetValue(ctx.AIContextualInput)
}

func (s *GameSession) aiPlanetObjective(aiIdx int) gamedata.AIObjective {
	obj := gamedata.AIObjectiveBalancedLow
	if aiIdx >= 0 && aiIdx < len(s.AIPlayers) {
		switch s.AIPlayers[aiIdx].Personality {
		case ai.PersonalityRuthless, ai.PersonalityXenophobic:
			obj = gamedata.AIObjectiveMineral
		case ai.PersonalityPacifist, ai.PersonalityHonorable:
			obj = gamedata.AIObjectivePopulation
		case ai.PersonalityAggressive:
			obj = gamedata.AIObjectiveBalancedHigh
		}
	}
	return obj
}

func (s *GameSession) aiBasePlanetValueAt(aiIdx, planetIdx int) int {
	if planetIdx < 0 || planetIdx >= len(s.Planets) {
		return 0
	}
	p := s.Planets[planetIdx]
	if p.NoPlanet || p.Gen < planetGenVersion || p.TypeID != gamedata.HABITABLE || !climateColonizable(p.ClimateID) {
		return 0
	}
	return gamedata.AIPlanetValue(gamedata.AIPlanetValueInput{
		Habitable: true,
		MaxPop:    gamedata.PlanetBasePopMax(p.SizeID, p.ClimateID),
		Minerals:  p.MineralID,
		Climate:   p.ClimateID,
		Gravity:   p.GravityID,
		// FoodBase 對應原版 planet.foodbase;remake 的等價量是該氣候的每農夫食物。
		FoodBase: gamedata.ClimateFoodPerFarmer(p.ClimateID),
		Special:  int(p.SpecialID),
	}, s.aiPlanetObjective(aiIdx))
}

// aiNeighborhoodResult 是某顆候選星的鄰近狀況(供 aiPlanetValue 的第二、三層用)。
type aiNeighborhoodResult struct {
	gamedata.AIContextualInput
	distances []int // 到每一顆我方已佔星的距離(整數化,見 aiNeighborRadius)
}

// aiNeighborRadius 是「鄰近」的判定半徑(星圖座標為 0..1 正規化,故用比例)。
//
// 原版判定的是「同一個恆星系內的其他行星」——那是很緊的範圍(同一顆恆星的其他軌道)。
// remake 一星一行星,只能用距離代替;半徑本身是 remake 的選擇,不是原版值。
// 取 0.15:24 星星圖的平均間距約 1/√24 ≈ 0.2,所以這大致等於「緊鄰的那一兩顆星」。
// 先前取 0.28(涵蓋約四分之一星圖)太寬——母星周圍一大片都會被當成「同系」,
// 全部套上「同系已有我方殖民地 → ×(size+1)/10」的邊際價值折扣,近處變得幾乎不值得殖民。
const aiNeighborRadius = 0.15

// aiDistanceUnit 把 0..1 的星圖距離換成原版距離表那種「小整數」尺度,
// 讓 AIProximityValue 的 120/distance 落在與第一層估值相稱的範圍
// (實測第一層是 1..80,取 40 讓鄰近加成落在 5..15;取 12 會讓近星拿到 60 分,
// 壓過行星本身的好壞——那不是原版的意思,原版的距離表單位另有尺度)。
const aiDistanceUnit = 40.0

// aiMaxNeighbors 是計入協同效應的鄰居數上限。
//
// 原版的「鄰居」是**同一恆星系內的其他行星**,一個星系最多 5 個軌道,所以天然上限是 4。
// remake 一星一行星、鄰近用半徑判定,鄰居數沒有天花板——不設限的話,
// 星圖密集處的空星會靠「鄰居 base 總和 / 8」堆出比行星本身還高的分數
// (實測:本體 3 分的貧瘠星最終 27 分,壓過本體 76 分的大型乾旱星)。
// 取最近的 4 顆,對齊原版的天然上限。
const aiMaxNeighbors = 4

// aiNeighborhood 掃候選星周圍,統計我方/敵方/無主鄰居,並收集到我方星的距離。
func (s *GameSession) aiNeighborhood(aiIdx, starIdx int, obj gamedata.AIObjective) aiNeighborhoodResult {
	var out aiNeighborhoodResult
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return out
	}
	me := s.Stars[starIdx]
	ownStars := map[int]bool{}
	if aiIdx >= 0 && aiIdx < len(s.AIPlayers) {
		for _, idx := range s.AIPlayers[aiIdx].ColonyStars {
			ownStars[idx] = true
		}
	}
	type neighbor struct {
		idx int
		d   float64
	}
	var near []neighbor
	for j := range s.Stars {
		if j == starIdx {
			continue
		}
		d := math.Hypot(s.Stars[j].X-me.X, s.Stars[j].Y-me.Y)
		if ownStars[j] {
			out.distances = append(out.distances, int(d*aiDistanceUnit))
		}
		if d <= aiNeighborRadius {
			near = append(near, neighbor{j, d})
		}
	}
	// 只取最近的 aiMaxNeighbors 顆(見該常數註解)。
	sort.Slice(near, func(a, b int) bool { return near[a].d < near[b].d })
	if len(near) > aiMaxNeighbors {
		near = near[:aiMaxNeighbors]
	}
	for _, n := range near {
		switch {
		case ownStars[n.idx]:
			out.NeighborOwnN++
			out.NeighborOwn += s.aiPlanetBaseValue(n.idx, obj)
		case s.Stars[n.idx].Owner != 0:
			out.NeighborEnemyN++
		default:
			out.NeighborEmpty += s.aiPlanetBaseValue(n.idx, obj)
		}
	}
	return out
}

// aiPlanetBaseValue 只算第一層(行星本身),供鄰居統計用——不能直接呼叫 aiPlanetValue,
// 否則會遞迴(鄰居的估值又要掃它自己的鄰居)。
func (s *GameSession) aiPlanetBaseValue(starIdx int, obj gamedata.AIObjective) int {
	if starIdx < 0 || starIdx >= len(s.Planets) {
		return 0
	}
	p, _ := s.PlanetDataAt(starIdx)
	if p.NoPlanet || p.Gen < planetGenVersion || p.TypeID != gamedata.HABITABLE ||
		!climateColonizable(p.ClimateID) {
		return 0
	}
	return gamedata.AIPlanetValue(gamedata.AIPlanetValueInput{
		Habitable: true,
		MaxPop:    gamedata.PlanetBasePopMax(p.SizeID, p.ClimateID),
		Minerals:  p.MineralID,
		Climate:   p.ClimateID,
		Gravity:   p.GravityID,
		FoodBase:  gamedata.ClimateFoodPerFarmer(p.ClimateID),
		Special:   int(p.SpecialID),
	}, obj)
}

// researchQueue 回傳「所有元件解鎖主題」依研究成本遞增去重排序的序列。作為研究自動推進的
// 順序:玩數回合累積研究點,便會由低階到高階逐步完成主題、逐步解鎖艦艇設計的進階元件。
func researchQueue() []gamedata.ResearchTopic {
	seen := map[gamedata.ResearchTopic]bool{}
	var q []gamedata.ResearchTopic
	for _, opts := range [][]Component{WeaponOptions, ArmorOptions, ShieldOptions, SpecialOptions} {
		for _, c := range opts {
			if gamedata.IsResearchableTopic(c.Tech) && !seen[c.Tech] {
				seen[c.Tech] = true
				q = append(q, c.Tech)
			}
		}
	}
	sort.Slice(q, func(i, j int) bool {
		return gamedata.ResearchChoiceFor(q[i]).Cost < gamedata.ResearchChoiceFor(q[j]).Cost
	})
	return q
}

// advanceResearch 在玩家目前研究主題完成後,把 ResearchTopic 推進到 researchQueue 中下一個
// 尚未完成的主題；全部一般主題完成時維持目前 Hyper 主題，讓它可重複累積。這讓
// 「研究→解鎖元件→造艦」的迴圈跨回合持續流動,
// 而非卡在單一主題。玩家仍可透過研究選擇畫面(SetResearchTopic)手動改變當前主題。
func (s *GameSession) advanceResearch() {
	if s.Player.CompletedTopics == nil || !s.Player.CompletedTopics[s.Player.ResearchTopic] {
		return // 目前主題尚未完成,繼續累積
	}
	for _, t := range researchQueue() {
		if !s.Player.CompletedTopics[t] {
			s.Player.ResearchTopic = t
			s.Player.ResearchApplication = 0
			s.Player.HasResearchApplication = false
			s.preparePlayerResearchApplication()
			return
		}
	}
}

// (歷史記錄)AI 母星原本用一個獨立的 averageHomeworldColony(FoodPerFarmer/IndustryPerWorker
// 維持 remake placeholder 4/6,不接查表值),因為 advanceAI 的造艦投資曾有整數捨去 bug——
// `FleetStrength += TotalNetIndustry/invest`,TotalNetIndustry 小於 invest(Scientific 性格為
// 4)時直接捨去成 0,FleetStrength 永久停滯。忠實 yield 下 AI NetIndustry 會穩定落在 3
// (3/4=0),接上去會讓 AI 軍力完全停止成長。該 bug 現已用 FleetInvestPool 餘數池修好(見
// advanceAI 註解:小額淨工業累積到池子裡跨回合兌現,不再捨去歸零),AI 母星於是與玩家共用
// 下方 playerHomeworldColony() 的忠實 yield,經濟對稱完整。詳見
// docs/tech/colony-economy-maintenance.md。

// playerHomeworldColony 建母星殖民地(玩家與 AI 共用):起始文明等級/PopMax/PlanetSize 設定,
// FoodPerFarmer/IndustryPerWorker 接
// gamedata.ClimateFoodPerFarmer(TERRAN)=2、gamedata.MineralIndustryPerWorker(ABUNDANT)=3
// ——母星氣候/礦產基準 Terran/Abundant(docs/tech/homeworld-init.md),手冊 GAME_MANUAL.pdf
// p.58-59/p.56-57 實據(見 planet_yield.go 逐頁引註)。
//
// Farmers=4/Workers=3(對調自 averageHomeworldColony 的 3/4)是機械必要的人口分配調整:
// Population=8、新 FoodPerFarmer=2 時,沿用舊的 Farmers=3 只夠 3×2=6 食物,餵不飽 8 人口
// (結構性饑荒);調成 Farmers=4 才能與消耗打平(4×2=8=8×1 消耗),這是把「農夫該配置多少人」
// 這個機械限制忠實反映出來,不是為了湊測試反推(該推導過程與數字見
// docs/tech/colony-economy-maintenance.md §2.1)。
//
// 接上忠實 yield 後開局第一回合是「零緩衝打平」(FoodSurplus=0),需要搭配兩項機制才不會被
// 隨機事件/安塔蘭入侵的人口損失推入永久饑荒鎖死:①EndTurn 的 recoverFromFamine(饑荒防死鎖,
// 見該函式)②engine.RunEmpireTurn 新接上的 gamedata.IncomeFoodSurplusRevenue(食物盈餘→BC,
// 見 empire.go),讓殖民地在食物盈餘轉正的回合能多存一點 BC 緩衝,吸收下次事件衝擊。
// 詳見 docs/tech/colony-economy-maintenance.md 本輪最新記錄(含 300 回合實測數字)。
//
// 2026-07-11 士氣接線訂正:MoralePercent 先前硬編 +10(無手冊依據的 remake placeholder),
// 現改用 colonyMoralePercent(獨裁/Dictatorship 基準 + homeworldBuildings() 已建建築)算出忠實值。
// **這個值是 0,不是原本想像的 +10**:獨裁政府「無 Barracks -20%」(手冊 p.21-22/p.165-167)在
// homeworldBuildings() 已含海軍陸戰隊營(Marine Barracks)時被解除、淨額歸零,且母星起始未建
// 全息模擬艙/歡樂穹頂,故無額外正面加成——0% 士氣即無 bonus/無 penalty 的中性起點,是手冊算出來
// 的忠實值,不是「退步」。這會讓新遊戲第一回合的食物/工業/研究產出比先前的 demo(+10% 灌水)少
// 一成,玩家會感覺到差異,回報/HONEST-STATUS.md 需誠實標明。政府基準選獨裁(索引 0)理由見
// GameSession.Government 欄位註解(自訂種族 0 點基準)。
func playerHomeworldColony() engine.ColonyState {
	return engine.ColonyState{
		// 職務分配 農4/工1/科3(2026-07-12 校正,SAVE10.GAM oracle:5 顆原版 turn-1 母星
		// 分配全部滿足「工≤2、科≥2」不變式,先前 農4工3科1 兩處違反——工3 超原版上限 2、
		// 科1 低於原版下限 2,母星科研被嚴重壓縮)。食物中性種族每農夫 2 食物 × 農4 = 8,剛好
		// 餵飽 pop8;餘 4 人依原版母星偏科研傾向配 工1/科3(SAVE10 Terran 種族多為工0-2/科2-4)。
		// ⚠ SAVE10 五名皆 AI 種族無 Human 樣本,精確三數為中信心重建(不變式本身高信心);
		// 分配漣漪到工業/稅收已用 moo2sim 開局軌跡驗證無死亡螺旋。見 original-gameplay-reference.md §7.0.1。
		// 2026-08-06 再校正為 農4/工2/科2:archive.org 線上原版實測(2026-07-12 oracle,
		// docs/tech/oracle-comparison-20260712.md)直接讀到 Sol III 母星就是 FARMERS 4 /
		// WORKERS 2 / SCIENTISTS 2。這是**直接觀察原版**(oracle 優先序高於 SAVE10 不變式重建),
		// 且同時滿足 SAVE10 的「工≤2、科≥2」兩條不變式。先前 4/1/3 是無 Human 樣本下的重建值。
		// PopMax 由 gamedata 依「大小×氣候」推導,不再硬編:archive.org 原版實測讀到母星是
		// **Medium Terran**(先前 remake 硬編 LARGE/上限 20 偏大)。
		// PlanetBasePopMax(MEDIUM,TERRAN) = (2+1)*5 * 80% = 12,與 oracle 筆記記的「約 12」相符。
		Population: 8, PopMax: gamedata.PlanetBasePopMax(gamedata.MEDIUM_PLANET, gamedata.TERRAN),
		Farmers: 4, Workers: 2, Scientists: 2,
		FoodPerFarmer:     gamedata.ClimateFoodPerFarmer(gamedata.TERRAN),
		IndustryPerWorker: gamedata.MineralIndustryPerWorker(gamedata.ABUNDANT),
		// 研究每科學家=銀河基準 3(手冊 p.949「usual 3」+ Psilon +2 邏輯 + SAVE10.GAM 驗證,
		// 見 gamedata.ResearchPerScientistNorm)。先前硬編 30 約 10x 過高,2026-07-12 校正。
		ResearchPerScientist: gamedata.ResearchPerScientistNorm,
		PlanetSize:           gamedata.MEDIUM_PLANET, // 原版 Sol III = Medium Terran(archive.org oracle)
		// 母星是自己的人,沒有未同化的外族人口 → 不套多種族懲罰。
		// 開局沒有任何成就科技,所以成就加成傳 0(見 achievements.go)。
		MoralePercent: colonyMoralePercent(gamedata.MoraleGovDictatorship, homeworldBuildings(), false, 0),
		// PlanetGravity 母星固定 Normal-G(手冊/homeworld-init.md 慣例基準,與 Terran/Abundant
		// 同一組母星設定),無重力懲罰。engine.ColonyState.PlanetGravity 的 Go 零值恰好是
		// gamedata.LOW_G(ordinal 0),必須明確賦值,不能依賴零值(見該欄位註解)。
		PlanetGravity: gamedata.NORMAL_G,
		// MineralRichness 母星固定 Abundant,與上面 IndustryPerWorker 用的
		// gamedata.MineralIndustryPerWorker(gamedata.ABUNDANT) 同一組母星礦產設定
		// (docs/tech/homeworld-init.md 慣例基準)。engine.ColonyState.MineralRichness 的
		// Go 零值恰好是 gamedata.ULTRA_POOR(ordinal 0),必須明確賦值(見該欄位註解)。
		MineralRichness: gamedata.ABUNDANT,
		// Climate 母星固定 Terran,與上面 FoodPerFarmer 用的
		// gamedata.ClimateFoodPerFarmer(gamedata.TERRAN) 同一組母星氣候設定。
		// engine.ColonyState.Climate 的 Go 零值恰好是 gamedata.TOXIC(ordinal 0),必須明確賦值
		// (見該欄位註解)——否則地形改造/蓋亞轉化(見 applySpecialAction)會誤判母星氣候。
		Climate: gamedata.TERRAN,
	}
}

// newHomeworldPlayerState 建立「Average 起始文明等級」的忠實起始 PlayerState:標記兩項
// 恆真起始科技已完成,依 docs/tech/homeworld-init.md §3.1/§5.1(MANUAL_150.html 一手來源,
// 與 openorion2 tech.cpp:170/212 交叉驗證,高信心):
//   - Tech field 0(TOPIC_STARTING_TECH):Capitol/Spy Network/Pulse Rifle 一律已知
//     (cost 0、無子項清單,ResearchTopic 層級本身即效果)。
//   - Tech field Engineering(TOPIC_ENGINEERING):Colony Base/Star Base/Marine Barracks
//     一律已知(ResearchAll=true)。ChosenTech 記入 Choices[0](TECH_COLONY_BASE)代表「全解」,
//     語意與 engine.recordCompletion 對 ResearchAll 主題的既有記錄慣例一致。
//
// BC 國庫 50(2026-07-12 校正,SAVE10.GAM oracle:5 名玩家開局 BC 全=50,humbe.no
// 攻略獨立記「~50 BC」交叉一致)。先前沿用 remake 預設 100 為未確認佔位值,已訂正。
//
// Maintenance 不再是無據 placeholder(先前寫死 5):改由 gamedata.BuiltMaintenanceBC 加總
// 母星起始已建成建築(homeworldBuildings:海軍陸戰隊營 1 BC + 星基 2 BC = 3 BC/回合,兩個
// 數字都是手冊 MaintenanceBC 實據,見 buildings.go)算出。玩家後續每回合的 Maintenance 由
// EndTurn 依 s.ColonyBuildings 實際清單重算(見 GameSession.totalBuildingMaintenance),
// 這裡只是開局第一回合前的初始建築值；運輸艦、指揮點超支、間諜與軍官會在
// prepPlayerDerived／RunEmpireTurn 依現況重算。
// ⚠ 這裡**只給 `TOPIC_STARTING_TECH`**(field 0,原版 `Init_Player_Tech_` 開頭那八個
// 無條件寫入的欄位,與 TECH LEVEL 無關)。其餘主題由 `applyStartingTech` 依 TECH LEVEL 發,
// 因為 TECH LEVEL 是 NEW GAME 畫面設的、比這裡晚。
// 先前這裡直接寫死 `TOPIC_ENGINEERING`,等於不論選哪一級都拿曲速前的科技。
func newHomeworldPlayerState(researchTopic gamedata.ResearchTopic) engine.PlayerState {
	return engine.PlayerState{
		// TaxRate 15:2026-07-12 校正。手冊 p.37 工業稅是「臨時要現金才拉」的補充收入(原生 0-50%、
		// 10% 級距、預設偏低),主收入是人頭(每人 1 BC,見 gamedata.BaseIncomePerPopHalfBC)。先前
		// 寫死 40 把工業稅當唯一收入來源,導致低工業母星流血。使用者定 remake 起始預設為 15(介於
		// 原版慣用的 0 與舊值之間;原生級距為 10 的倍數,15 是 remake 起始值,玩家後續可調)。
		BC: 50, TaxRate: 15, Maintenance: gamedata.BuiltMaintenanceBC(homeworldBuildings()), ResearchTopic: researchTopic,
		// CommandPointsSupply 這裡刻意只填母星星基(homeworldBuildings 的"星基":true)貢獻的
		// 1 點建築供給,不含 gamedata.CommandPointsBase(帝國基礎值 5)——這只是「第一次 EndTurn
		// 前」的暫時值；玩家會在 EndTurn 用 totalCommandPointsSupply 重算，AI 則在自己的
		// 經濟結算前由 syncAICommandPoints 依 ColonyBuildings 與實艦重算。
		// UsedCommandPoints 這裡刻意不填：艦隊尚未建立，玩家與 AI 都在艦艇就位後計算。
		CommandPointsSupply: gamedata.CommandPointsFromBuildings(homeworldBuildings()),
		CompletedTopics: map[gamedata.ResearchTopic]bool{
			gamedata.TOPIC_STARTING_TECH: true, // field 0,無條件(見本函式註解)
		},
		ChosenTech: map[gamedata.ResearchTopic]gamedata.Technology{
			gamedata.TOPIC_ENGINEERING: gamedata.TECH_COLONY_BASE, // ResearchAll 代表值(全解語意)
		},
	}
}

// homeworldBuildings 是 Average 起始文明等級母星「已建成」的常駐建築標記,依
// docs/tech/homeworld-init.md §3.2/§3.3(MANUAL_150.html 一手來源,高信心):
//   - Marine Barracks + Star Base:唯二出現在預設 initial_buildings 清單且技術已知的項目
//     ("Pre-warp and Average Tech games only start with Marine Barracks and a Star Base")。
//   - Colony Base 刻意不列入:它是一次性殖民行動,非常駐建築(§3.3)。
//   - Capitol 不計入 StartingBuildingCount 上限，但會以 `CapitolBuildName` 存進
//     ColonyBuildings；正常開局自動給予，失都後只在指定行星重建。
//
// 建築數 2 遠低於 StartingBuildingCount(8, BuildingCapAverage)=5 的上限——這是符合手冊的
// (上限只是「至多」,實際只有這兩項的科技條件成立,見 §3.3)。
// ⚠ 2026-08-07:這一組先前是**寫死的兩棟**。現在改成從原版的優先清單算出來
// (`gamedata.InitialBuildings`,清單來自 `Init_Homeworld_Colony2_` @ 0x13A3D 的 `word_17D8AC`)。
// 一般等級算出來仍然**正好是這兩棟**——那不是巧合,是手冊那句話的機器版驗證
// (「no other techs are Known that are also in the default initial buildings list」)。
// 先進等級才會多出東西,而先進等級先前完全沒有差別。
func homeworldBuildings() map[string]bool {
	return homeworldBuildingsFor(TechLevelDefault, homeworldStartPop)
}

// homeworldStartPop 是母星開局人口(oracle:SAVE10.GAM 8 pop,見 docs/tech/homeworld-init.md)。
//
// 只有「開局建築數」用得到它:數量 = min(⌈⅔ pop⌉, 等級上限)。
const homeworldStartPop = 8

// homeworldBuildingsFor 依 TECH LEVEL 與人口算出開局已建成的建築。
//
// 兩層限制,兩層都是真值:
//
//	數量 = min(⌈⅔ 人口⌉, `gamedata.InitialBuildingCap`)   ; 上限表 = 原版 byte_13A3A = 3/5/9
//	內容 = 優先清單裡**科技條件成立**的,照原版順序                ; 清單 = 原版 word_17D8AC
//
// 科技用的是 `gamedata.StartingTopics(techLevel)` + field 0 —— 與 `applyStartingTech`
// 發下去的那一組**同一個來源**。兩邊各算一次的話,「有這棟建築但沒有它的科技」這種矛盾
// 會靜靜地出現。
func homeworldBuildingsFor(techLevel, pop int) map[string]bool {
	known := map[gamedata.ResearchTopic]bool{gamedata.TOPIC_STARTING_TECH: true}
	for _, t := range gamedata.StartingTopics(techLevel) {
		known[t] = true
	}
	return homeworldBuildingsForKnown(techLevel, pop, known)
}

// homeworldBuildingsForKnown 同上,但直接吃**這位玩家實際知道的主題集合**。
//
// 2026-08-07(第 44 項(下游讀真值)):第 43 項(先進級開局主題)把先進級的 19 個隨機主題發出去之後,
// `homeworldBuildingsFor` 那條路就不夠用了——它從固定表現算科技集合,
// **看不到那 19 個**,所以先進級的母星仍然只蓋得出兩棟(手冊說先進級上限是 9 棟)。
//
// 這正是第 31 項(開局建築清單)寫的那條依賴鏈:開局建築取決於開局知道哪些科技。
// 上游補齊了,下游就得跟著讀真正的集合,而不是再算一次固定表。
func homeworldBuildingsForKnown(techLevel, pop int, known map[gamedata.ResearchTopic]bool) map[string]bool {
	n := StartingBuildingCount(pop, gamedata.InitialBuildingCap(techLevel))
	out := map[string]bool{}
	for _, name := range gamedata.InitialBuildings(known, n) {
		out[name] = true
	}
	return out
}

func homeworldBuildingsLegacy() map[string]bool {
	return map[string]bool{
		"海軍陸戰隊營": true, // Marine Barracks
		"星基":     true, // Star Base
	}
}

// demoAIOpponentSetup 是 NewDemoSession 建立各 AI 對手時的固定名稱/性格配置(順序對應
// AIPlayers[0]/[1]/[2])。三個都取自 Races(session.go 上方 13 經典種族表)裡實際存在的
// 種族名,對應手冊描述的招牌性格,搭配 ai.Profile 的造艦/研究權重:
//   - 席隆人(Psilons):手冊「創造性研究,科學家產出高」→ ai.ProfileScientific(重研究)。
//   - 姆瑞森人(Mrrshan):手冊「好戰善攻,艦艇攻擊加成」→ ai.ProfileAggressive(重工業造艦)。
//   - 布拉西人(Bulrathi):手冊「體格強悍,地面與戰鬥加成」→ ai.ProfileExpansionist(偏工業,
//     這裡取「擴張」而非「好戰」對應,避免兩個 AI 都是同一種造艦優先權重、行為趨同看不出差異;
//     手冊沒有描述 Bulrathi 特別擅長殖民擴張,這點是 remake 為了讓 3 個 AI 行為可辨識的選擇,
//     非手冊逐字對應)。
//
// 舊版單 AI demo 用的名稱是「AI (賽隆人)」——「賽隆人」四字實際上不在 Races 表裡(疑似「席隆人」
// 的手誤,且 cmd/moo2 的 diplomatRaceIndex 早已把它當「舊字串相容」映射到薩克拉肖像,可見
// 命名本身從一開始就不精確)。這裡順手訂正為 Races 表裡真實存在的「席隆人」,不延續錯字——沒有
// 任何測試字串比對這個名稱(已查證,見 grep AIPlayers[0].Name),訂正不影響既有測試。
//
// 2026-08-06 起 profile 不再手動指派:raceEn 對到原版 AIRACES.CFG 的種族性格分布
// (ai.ClassicRacePersonality),開局抽出性格後由 ai.ProfileForPersonality 推導經濟傾向。
// 那張分布表 remake 早就有,但一直是死碼——三個 AI 的行為差異全靠這裡手寫的 profile。
var demoAIOpponentSetup = []struct {
	name   string
	raceEn string // AIRACES.CFG 的種族名(ai.ClassicRacePersonality 的 key)
}{
	{"AI (席隆人)", "Psilons"},
	{"AI (姆瑞森人)", "Mrrshan"},
	{"AI (布拉西人)", "Bulrathi"},
}

// pickAIPersonality 依原版公式從種族的性格分布抽一個性格。
//
// 原版(docs/tech/original-ai-re.md §1.3):
//
//	column := Random(10) + 1 - difficulty_byte(0-4)   // clamp 到 1..10
//	personality := race_personality[race][column-1]
//
// 難度越高,欄位越往前偏 → 抽到分布表前段(通常是比較兇的性格)的機率越大。
// 查無此族時退回「反覆無常」這個中性性格,不臆造。
func pickAIPersonality(raceEn string, difficulty int, r *rand.Rand) ai.Personality {
	dist, ok := ai.ClassicRacePersonality(raceEn)
	if !ok {
		return ai.PersonalityErratic
	}
	col := r.Intn(10) + 1 - difficulty
	if col < 1 {
		col = 1
	}
	if col > 10 {
		col = 10
	}
	return dist[col-1]
}

// NewDemoSession 建一個最小可玩對局:玩家 + 3 個性格互異的 AI 對手(多帝國競爭骨架,見
// demoAIOpponentSetup),各自持 Average 起始文明等級的單一母星(docs/tech/homeworld-init.md,
// 取代先前程序生成的假殖民地)。玩家與各 AI 母星 yield 皆接忠實 Terran/Abundant 查表值
// (playerHomeworldColony)——AI 原本因 advanceAI 造艦投資的整數捨去 bug 而暫維持 placeholder
// yield,該 bug 已用 FleetInvestPool 餘數池修好(見 advanceAI 註解),經濟對稱完整。
//
// 2026-07-11 由 1 AI 擴為 3 AI(激活真議會,見 council.go councilEligible/advanceCouncil):
// 資料模型(AIPlayers 平行陣列、PlayerSpies 平行陣列、council 的 extantRaceCount/
// aiPopulationTotal 迴圈)先前就已是「對任意個數 AI 迴圈處理」的寫法,只是從未真的建過 >1 個
// AI 去驗證——見各處「天然支援,只是 1 AI 看不出差異」的既有註解。這裡是把資料模型的既有
// N-ready 設計第一次接上實際的 N=3。
//
// 供「最小可玩迴圈」骨架用;正式新遊戲流程(選種族/星系生成/起始文明等級選擇,含真正的多 AI
// 建構)為後續工作——cmd/moo2 的 RegenGalaxy 呼叫端(customrace.go/raceselect.go)目前完全
// 不建立 AIPlayers,是既有落差,不在本輪範圍內。
// buildDemoAIOpponents 依各 AI 母星索引(aiHomeStars,通常來自 genGalaxy 第二個回傳值)建立
// 一組 AIOpponent:名稱/性格依序取自 demoAIOpponentSetup(席隆人/姆瑞森人/布拉西人…,索引超出
// 表長度則循環使用),各自持 Average 起始文明等級的單一母星殖民地(playerHomeworldColony,與
// 玩家共用忠實 yield)。NewDemoSession 與 SetupNewGame 共用此函式,確保「新遊戲開局怎麼建 AI」
// 只有一個權威實作,不會兩處各自維護一份、逐漸漂移不一致。
func buildDemoAIOpponents(aiHomeStars []int, starCount, difficulty int, seed int64) []AIOpponent {
	aiPlayers := make([]AIOpponent, 0, len(aiHomeStars))
	// 性格抽樣用獨立的亂數流:同一個 seed 一定抽出同一組性格(存讀檔與重跑要可重現)。
	pr := rand.New(rand.NewSource(seed*31 + 17))
	// sub_589D6 的 raw profile 只在建立帝國時抽一次。使用與性格分離的穩定 stream，
	// 避免新增研究資料改變既有 personality 序列；原版全域 PRNG 位元序列仍不宣稱一致。
	tr := rand.New(rand.NewSource(seed*6364136223846793005 + 101))
	for i := 0; i < len(aiHomeStars); i++ {
		setup := demoAIOpponentSetup[i%len(demoAIOpponentSetup)]
		pers := pickAIPersonality(setup.raceEn, difficulty, pr)
		raceIndex := raceIndexForEnglishName(setup.raceEn)
		var techProfile gamedata.OriginalAITechProfile
		techProfileKnown := false
		if raceIndex >= 0 && raceIndex < len(Races) {
			origRace := Races[raceIndex].OrigIdx
			if traits, ok := gamedata.OrigRaceTraits(origRace); ok {
				raw27 := gamedata.RollOriginalAIRaw27(origRace, difficulty, tr.Intn)
				techProfile = gamedata.RollOriginalAITechProfile(traits, difficulty, raw27, tr.Intn)
				techProfileKnown = true
			}
		}
		player := newHomeworldPlayerState(1)
		// 原版 player+0x31 稅率 byte 的直接寫入只存在於真人稅率 UI；AI 回合沒有
		// 自動調稅 producer。新建 AI 因此使用清零初始化的 0%，其後回合保持現值。
		// 匯入 .GAM 的 AI 若已有其他值，正常回合也不會在上方被擅自覆蓋。
		player.TaxRate = gamedata.TaxRateMinPercent
		startingColonyShips := homeworldShips(defaultNameTranslator)[:1]
		startingColonyShips[0].RawType = gamedata.COLONY_SHIP
		startingColonyShips[0].RawTypeKnown = true
		startingColonyShips[0].RawMissionKnown = true
		aiPlayers = append(aiPlayers, AIOpponent{
			Name:               setup.name,
			Color:              (i + 1) % 8,
			ColorKnown:         true,
			RaceIndex:          raceIndex,
			PopulationRaceSlot: i + 1, PopulationRaceSlotKnown: true,
			Player:      player,
			Colonies:    []engine.ColonyState{playerHomeworldColony()}, // AI 同為 Average 起始單一母星,與玩家共用忠實 yield
			ColonyStars: []int{aiHomeStars[i]},                         // 唯一有實際殖民地模型的星(見 AIOpponent.ColonyStars 註解)
			// ColonyBuildings 母星比照玩家,開局已建成 homeworldBuildings()(海軍陸戰隊營+
			// 星基)——每個 AI 各自 cloneBuildings 一份獨立拷貝,不可共享同一個 map 參考(見
			// AIOpponent.ColonyBuildings 欄位註解)。
			ColonyBuildings: []map[string]bool{cloneBuildings(homeworldBuildings())},
			// Average 開局的實際殖民船由 All_AI_Colonize_ 消費；不再讓每五回合的
			// remake 排程無限免費拓殖。偵察艦仍待 AI 多艦隊開局模型閉合。
			Ships: startingColonyShips,
			// 原版開局後由全域英雄池逐回合產生 offer；不再依種族固定贈送 Commando。
			Leaders:                    nil,
			Personality:                pers,
			OriginalTechProfile:        techProfile,
			OriginalTechProfileKnown:   techProfileKnown,
			OriginalRaw28:              techProfile.Raw6,
			OriginalRaw28Known:         techProfileKnown,
			OriginalHumanIncidentKnown: true,
			// 經濟傾向由性格推導(見 ai.ProfileForPersonality),不再手寫。
			Decider:    ai.NewRemakeDecider(ai.ProfileForPersonality(pers)),
			OwnedStars: 1,
			ExploredStars: func() []bool {
				visited := make([]bool, starCount)
				visited[aiHomeStars[i]] = true
				return visited
			}(),
			ExploredStarsKnown: true,
			// 基礎關係傾向依性格起跳(原版 _personality_relation_modifiers):
			// 和平主義 +30、排外 -50……先前所有 AI 一律從 0 開始,性格毫無體感。
			Relation: clampRelation(ai.PersonalityRelationModifier(pers) / 2),
		})
		if raceIndex >= 0 && raceIndex < len(Races) {
			orig := Races[raceIndex].OrigIdx
			has := func(trait gamedata.RaceTrait) bool { return gamedata.OrigRaceHasTrait(orig, trait) }
			applyHomeworldRaceTraits(&aiPlayers[len(aiPlayers)-1].Colonies[0], nil, Races[raceIndex], has)
			// 重建 owner population group，讓母星環境與數值 picks 在第一回合職務排序前生效。
			(&GameSession{}).syncAIRaceEngineFields(&aiPlayers[len(aiPlayers)-1])
		}
		// Marine Barracks 是 AI 母星的開局建築；以原版「初建立即最多 4 單位」
		// 建立駐軍，而不是等第一次入侵時才用目前回合數倒推。其餘殖民地由
		// aiExpand 追加空的平行 slot。
		ensureAIGroundForceSlots(&aiPlayers[len(aiPlayers)-1])
	}
	return aiPlayers
}

// raceIndexForEnglishName 將 AI 建構表的英文種族鍵轉回玩家種族表索引。
// 這是資料模型接線,不是新的種族判定來源;AIRACES.CFG 的性格表仍由 raceEn 驅動。
func raceIndexForEnglishName(name string) int {
	for i, r := range Races {
		if r.EnName == name {
			return i
		}
	}
	return -1
}

// defaultNameTranslator 是行程層的專有名詞翻譯器(星名/艦名),由 cmd/moo2 在啟動時
// 用 SetNameTranslator 注入。nil = 不翻(英文)。
//
// ⚠ **為什麼是套件級變數而不是參數**:`NewDemoSession` 有十來個呼叫端(測試、探針、
// 截圖廊、網路對局),而語言在整個行程裡只設定一次。改簽名的成本遠高於它換來的東西。
// 每一局仍可用 `GameSession.TranslateName` 覆寫。
var defaultNameTranslator func(string) string

// SetNameTranslator 設定行程層的專有名詞翻譯器。cmd/moo2 在載入譯表之後呼叫一次。
func SetNameTranslator(fn func(string) string) { defaultNameTranslator = fn }

func NewDemoSession() *GameSession {
	const galaxyStars = 24
	const numAIOpponents = 3
	// 程序化星系(24 星,固定種子=可重現;正式版種子隨新遊戲)
	galaxy, aiHomeStars := genGalaxy(galaxyStars, 42, numAIOpponents, galaxyAgeSetting, defaultNameTranslator)
	galaxy[0].Explored = true // 母星初始已探索
	// 蟲洞用獨立亂數流(45),與 SetupNewGame 同構——demo 局也要有,否則截圖廊/測試看不到這一層。
	genWormholes(galaxy, demoHomeStarSet(aiHomeStars), rand.New(rand.NewSource(45)).Intn)
	// 星雲同理:demo 局也要有,否則截圖廊/測試看不到這一層。
	demoNebulae := genNebulae(galaxySizeClass(len(galaxy)), demoHomeStarSet(aiHomeStars),
		galaxy, rand.New(rand.NewSource(46)))

	aiPlayers := buildDemoAIOpponents(aiHomeStars, len(galaxy), 1, 42) // demo 固定難度 1 / seed 42(可重現)

	session := &GameSession{
		Turn:                               1,
		Difficulty:                         1, // 與 buildDemoAIOpponents(..., 1, ...) 同一份示範局難度，避免零值誤套 Tutor。
		Player:                             newHomeworldPlayerState(gamedata.TOPIC_ADVANCED_CONSTRUCTION),
		PlayerColonies:                     []engine.ColonyState{playerHomeworldColony()},
		ColonyBuildings:                    []map[string]bool{homeworldBuildings()},
		PlayerColonyStars:                  []int{0},                       // 母星 = 星 0(見欄位註解)
		Government:                         gamedata.MoraleGovDictatorship, // 預設獨裁(自訂種族 0 點基準),見欄位註解的零值陷阱說明
		AIPlayers:                          aiPlayers,
		EnableAIVsAI:                       true, // 新示範／新局啟用；舊存檔缺欄位仍維持關閉
		OriginalCouncilDiplomacyStateKnown: true,
		PlayerSpies:                        make([]int, len(aiPlayers)),        // 玩家對每個 AI 對手的間諜數,平行 AIPlayers,開局皆 0(見欄位/spy.go ensurePlayerSpies 註解)
		PlayerSpyMissions:                  make([]SpyMission, len(aiPlayers)), // 零值 STEAL,與原本最小迴圈相容
		Stars:                              galaxy,
		Nebulae:                            demoNebulae,
		Planets:                            genPlanets(galaxy, rand.New(rand.NewSource(43)), rand.New(rand.NewSource(47)), galaxyAgeSetting, demoHomeStarSet(aiHomeStars)),
		// Monsters 在下面 session 建好後補上——genMonsters 會就地修改 Planets(手冊 p.60:
		// 有怪獸的星系一定另有一個特殊物產),不能在同一個複合字面值裡引用尚未建立的欄位。
		// 開局領袖池為空(2026-07-12 手冊考據校正)。手冊 GAME_MANUAL.pdf p.47 + p.134「Mercenary
		// Leaders」:原版開局玩家**完全沒有任何領袖**,傭兵不定期上門、須花雇用費招入 Leader Pool
		// (上限殖民領袖 4 + 艦艇軍官 4)。先前 demoLeaders() 讓玩家開局自帶「馮·諾伊曼 科學家」並
		// 固定 +25 研究套進母星,是機制錯誤(那應是雇用並指派後才生效)。改為 nil = 忠實空池。
		// demoLeaders()/applyLeaderColonyBonuses 保留供未來「傭兵招募流程」實作後 seed 用(TODO)。
		Leaders: nil,
		// 開局一支艦隊,停在母星(星 0),沒有航行任務(見 fleet.go)。
		Fleets: []Fleet{{Ships: homeworldShips(defaultNameTranslator), AtStar: 0, DestStar: -1}},
		// 母星開局預設建造「貿易品」:archive.org 線上原版實測(2026-07-12 oracle)讀到 Sol III
		// 的 BUILDING 欄就是「Trade Goods」,右下 Income +12 BC——remake 先前開局是「不建造」、
		// 收支 +0,是母星開局態沒對齊的一環(見 docs/tech/oracle-comparison-20260712.md)。
		// Cost 0 同「不建造」語意(見 TradeGoodsBuildName 註解:整包工業轉現金,不累積進度)。
		Builds:              []ColonyBuild{{Name: TradeGoodsBuildName, Progress: 0, Cost: 0}},
		SelectedStar:        -1,
		ShowRelocationLines: true, // 原版預設開(`sub_127E1` 初始化寫 1)
		GameSettings:        DefaultGameSettings(),
		EventSeed:           42, // 隨機事件種子(可重現;正式新遊戲遞增)
		RuleProfile:         gamedata.Profile15(),
	}
	// 守衛怪獸(見 monster.go)。放在這裡而不是上面的複合字面值裡,因為 genMonsters 會就地
	// 修改 session.Planets(手冊 p.60:有怪獸的星系一定另有一個特殊物產)。
	session.Monsters = genMonsters(galaxy, session.Planets, rand.New(rand.NewSource(44)), demoHomeStarSet(aiHomeStars))
	session.ensureAIRelations()
	session.ensureAIAIState()
	session.syncAIColonyPlanets()                                  // AI 殖民地的行星索引(見該函式)
	session.Player.UsedCommandPoints = session.usedCommandPoints() // 依開局艦隊(homeworldShips)算實際需求,顯示與第一次 EndTurn 後一致
	// 領袖技能接線(2026-07-11):把 Ship=false 的殖民地領袖(科學家/貿易家)技能套到母星。
	// 2026-07-12 開局改為空領袖池(見上方 Leaders 註解),故此呼叫目前是 no-op;保留接線,待未來
	// 傭兵招募流程實作後,玩家雇用並指派殖民地領袖時即生效。
	applyLeaderColonyBonuses(session.Leaders, &session.PlayerColonies[0])
	// 玩家母星座落的行星(見 PlayerColonyPlanets 欄位註解)。星 0 恆為母星,
	// demoHomeStarSet 保證那裡有可殖民天體。
	session.PlayerColonyPlanets = []int{session.PlanetAt(0)}
	session.ensureCapitolState()
	// 開局研究主題(依 TECH LEVEL;demo 局沒設過,`techLevel()` 退回「一般」)。
	// 與 SetupNewGame 走同一條路,免得兩邊漂開。
	session.applyStartingTech()
	for i := range session.AIPlayers {
		session.ensureAIShipDesigns(i)
	}
	return session
}

// SetRuleProfile 設定這局遊戲的版本規則 profile(gamedata.Profile13()/Profile15())。
//
// 最小掛勾:供未來主選單「選 1.3/1.5」的新遊戲流程呼叫(建立 GameSession 後、EndTurn 前設定
// 一次),本任務不接 UI,只確保注入路徑存在。SetupNewGame(重開新局)刻意不重置 RuleProfile——
// 版本規則由更上層的「選版本」流程決定,不屬於「重新產生星系/AI」的 SetupNewGame 職責範圍。
func (s *GameSession) SetRuleProfile(p gamedata.RuleProfile) {
	s.RuleProfile = p
}

// CycleTaxRate 循環切換帝國工業稅率(手冊 GAME_MANUAL.pdf p.37:點國庫框調整,0-50%、
// 10% 級距)。每次呼叫進到下一個 10% 級距,超過 50% 繞回 0%;非標準起始值(如 remake
// 預設 15)會進位到下一個 10 的倍數。稅率影響 advanceBuilds(建造吃 (100-稅率)% 工業)與
// RunEmpireTurn(稅率% 工業換 BC),是「更多錢 vs 更快建造」的取捨。
func (s *GameSession) CycleTaxRate() {
	s.recordPlayerCommand(PlayerCommand{Name: CmdCycleTaxRate})
	next := (s.Player.TaxRate/gamedata.TaxRateStepPercent + 1) * gamedata.TaxRateStepPercent
	if next > gamedata.TaxRateMaxPercent {
		next = gamedata.TaxRateMinPercent
	}
	s.Player.TaxRate = next
}

// SetMercCandidates 注入傭兵候選池(cmd 層從 HERODATA.LBX 真英雄解析後傳入,見
// cmd/moo2 loadHerodataMercs)。傳空則維持內建策展名單。建議依等級升冪排序,使開局先遇到低階
// (便宜、雇得起)傭兵,對齊攻略「開局只有最低階領袖」。
func (s *GameSession) SetMercCandidates(pool []Leader) { s.MercCandidatePool = pool }

// mercCandidates 回傳傭兵候選名單:優先用注入的 HERODATA 真英雄池(MercCandidatePool);未注入
// (無 LBX / headless)時退回內建策展名單,**低階起步**(手冊/攻略:「Only the lowest level
// leaders are available at the start of the game」),Level 1-2、雇用費親民。
func (s *GameSession) mercCandidates() []Leader {
	if len(s.MercCandidatePool) > 0 {
		return s.MercCandidatePool
	}
	return []Leader{
		{ID: 0, Name: "馮·諾伊曼", Skill: "科學家", Level: 2, Ship: false, Tier: 1},
		{ID: 1, Name: "洛克斐勒", Skill: "貿易家", Level: 1, Ship: false, Tier: 1},
		{ID: 2, Name: "漢尼拔", Skill: "指揮官", Level: 2, Ship: true, Tier: 1},
		{ID: 3, Name: "圖靈", Skill: "工程師", Level: 1, Ship: true, Tier: 1},
	}
}

func (s *GameSession) officerRandForTurn() *randStream {
	if s.officerRand == nil {
		s.officerRand = newRandStream(s.EventSeed*2654435761 + 31)
	}
	return s.officerRand
}

func (s *GameSession) officerRecruitChance() int {
	return officerRecruitChanceFor(s.Turn, s.MercLastOfferTurn, s.Leaders,
		s.RaceCharismatic(), s.RaceRepulsive())
}

func officerRecruitChanceFor(turn, lastOffer int, leaders []Leader, charismatic, repulsive bool) int {
	if turn < 5 || len(leaders) >= 8 {
		return 0
	}
	elapsed := turn
	if lastOffer > 0 {
		elapsed = turn - lastOffer
	}
	chance := elapsed + 1
	if charismatic {
		chance += 5
	}
	if repulsive {
		chance -= 10
	}
	for _, leader := range leaders {
		level := leader.Level
		if level < 1 {
			level = 1
		}
		switch leaderSkillTier(leader, int(gamedata.SKILL_FAMOUS)) {
		case 1:
			chance += level
		case 2:
			chance += 15 * level / 10
		}
	}
	chance /= max(1, len(leaders)+1)
	return max(0, chance)
}

func sameLeader(a, b Leader) bool {
	if a.ID != 0 || b.ID != 0 {
		return a.ID == b.ID
	}
	return a.Name == b.Name
}

func (s *GameSession) leaderAlreadyVisible(candidate Leader) bool {
	for _, list := range [][]Leader{s.Leaders, s.MercPool} {
		for _, leader := range list {
			if sameLeader(candidate, leader) {
				return true
			}
		}
	}
	for i := range s.AIPlayers {
		for _, leader := range s.AIPlayers[i].Leaders {
			if sameLeader(candidate, leader) {
				return true
			}
		}
		if s.AIPlayers[i].LeaderOffer != nil && sameLeader(candidate, *s.AIPlayers[i].LeaderOffer) {
			return true
		}
	}
	if s.OfficerCooldowns != nil && s.OfficerCooldowns[candidate.ID] > 0 {
		return true
	}
	return false
}

// advanceMercOffers 依原版 sub_97A66/sub_9781D/sub_97B2D 逐回合擲招募機率並從
// 隨星曆開放的候選前綴隨機選人。remake 的獨立亂數流可存檔，但不宣稱原版 PRNG 位元一致。
func (s *GameSession) advanceMercOffers() {
	s.advanceAILeaderOffers()
	s.advancePlayerMercOffer()
	s.generateAILeaderOffers()
}

func (s *GameSession) advancePlayerMercOffer() {
	chance := s.officerRecruitChance()
	if chance > 0 && s.officerRandForTurn().Intn(100) < chance {
		if candidate, ok := s.pickOfficerCandidate(s.Leaders, s.RaceCharismatic(), s.RaceRepulsive()); ok {
			s.MercPool = []Leader{candidate}
			s.MercLastOfferTurn = s.Turn
		}
	}
}

func (s *GameSession) pickOfficerCandidate(leaders []Leader, charismatic, repulsive bool) (Leader, bool) {
	cands := s.mercCandidates()
	// word_199998 是本局玩家槽數，不是目前席位；Next_Turn_Calc 在 0x137D1..0x137E7
	// 也用它作 Random_Officer_Check 的迴圈上限。
	prefix := s.Turn/5 + 10 + 1 + len(s.AIPlayers)
	if charismatic {
		prefix += 10
	} else if repulsive {
		prefix /= 2
	}
	prefix = min(max(prefix, 0), len(cands))
	if prefix == 0 {
		return Leader{}, false
	}
	for tries := 0; tries < 100; tries++ {
		candidate := cands[s.officerRandForTurn().Intn(prefix)]
		if leaderSlotsFullFor(leaders, candidate.Ship) || s.leaderAlreadyVisible(candidate) {
			continue
		}
		return candidate, true
	}
	return Leader{}, false
}

// MercHireCost 回傳雇用某傭兵的一次性費用(gamedata.LeaderHireCost,依技能等級遞增)。
func (s *GameSession) MercHireCost(ld Leader) int {
	exp := leaderDisplayLevelToExpLevel(ld.Level)
	return gamedata.LeaderHireCost(5, exp, leaderFamousHireModifier(s.Leaders))
}

// leaderSlotsFull 回傳該類領袖(殖民地 Ship=false / 艦艇 Ship=true)是否已達上限 4
// (手冊 p.134:「up to four Colony Leaders and four Ship Officers」)。
func (s *GameSession) leaderSlotsFull(ship bool) bool {
	return leaderSlotsFullFor(s.Leaders, ship)
}

func leaderSlotsFullFor(leaders []Leader, ship bool) bool {
	n := 0
	for _, l := range leaders {
		if l.Ship == ship {
			n++
		}
	}
	return n >= 4
}

// HireMercAt 雇用 MercPool 指定位置的傭兵(手冊 p.134:扣一次性雇用費、招入 Leader Pool)。
// BC 不足或對應領袖類別已滿(各 4 名)則不雇用、傭兵留在池中。回傳是否成功。
//
// 索引而非名稱是這個畫面的暫時識別方式：同名英雄不應在 UI 層被猜成同一人，
// 且 MercPool 的候選順序本身就是「待僱佇列」的可見順序。
func (s *GameSession) hireMercAt(index int) bool {
	if index < 0 || index >= len(s.MercPool) {
		return false
	}
	ld := s.MercPool[index]
	if s.leaderSlotsFull(ld.Ship) {
		return false
	}
	newBC, ok := engine.HireLeader(s.Player.BC, s.MercHireCost(ld))
	if !ok {
		return false // BC 不足
	}
	s.Player.BC = newBC
	copy(s.MercPool[index:], s.MercPool[index+1:])
	s.MercPool = s.MercPool[:len(s.MercPool)-1]
	s.Leaders = append(s.Leaders, ld)
	// 只套「新雇這一名」的殖民地加成——applyLeaderColonyBonuses 是 += 累加,不可對全名單重跑
	// (會重複計算既有領袖),故傳單元素 slice(見該函式註解)。
	if len(s.PlayerColonies) > 0 && !ld.Ship {
		applyLeaderColonyBonuses([]Leader{ld}, &s.PlayerColonies[0])
	}
	return true
}

func (s *GameSession) HireMercAt(index int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdHireMercAt, Args: []int{index}})
	return s.hireMercAt(index)
}

// HireMerc 保留舊的「雇用佇列首名」入口，供指令層與舊回放相容；新畫面用
// HireMercAt 讓玩家可以依手冊的 HIRE 模式挑選指定候選。
func (s *GameSession) HireMerc() bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdHireMerc})
	return s.hireMercAt(0)
}

// SystemCompositionText 回傳「這個星系除了代表行星以外還有什麼」的摘要
// (如「同系:氣態巨星×2、小行星帶」);沒有其他天體回空字串。
//
// **從軌道表算**,不是讀 `Planet.SystemBodies` —— 那個欄位是一星一行星時代的折衷,
// 它自己的註解就在擔心「兩份資料要同步」。現在同系天體是真正的行星條目,
// 摘要從那裡數出來,只有一份資料。
func (s *GameSession) SystemCompositionText(star int) string {
	rep := s.PlanetAt(star)
	order := []gamedata.PlanetType{gamedata.HABITABLE, gamedata.GAS_GIANT, gamedata.ASTEROIDS}
	n := map[gamedata.PlanetType]int{}
	for _, p := range s.PlanetsAt(star) {
		if p == rep {
			continue // 代表行星本身不算「同系其他天體」
		}
		n[s.Planets[p].TypeID]++
	}
	out := ""
	for _, t := range order {
		if n[t] == 0 {
			continue
		}
		if out != "" {
			out += "、"
		}
		out += planetTypeDisplayName(t)
		if n[t] > 1 {
			out += "×" + strconv.Itoa(n[t])
		}
	}
	if out == "" {
		return ""
	}
	return "同系:" + out
}

// SystemBodyCountText 回傳「同系還有幾個天體」的極短字串(如「另有 3 天體」),
// 供欄位很窄的行星列表用;沒有其他天體回空字串。完整組成用 SystemCompositionText。
func (s *GameSession) SystemBodyCountText(star int) string {
	n := len(s.PlanetsAt(star)) - 1 // 扣掉代表行星本身
	if n <= 0 {
		return ""
	}
	return "另有 " + strconv.Itoa(n) + " 天體"
}
