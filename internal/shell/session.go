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
	Name            string
	Player          engine.PlayerState
	Colonies        []engine.ColonyState
	Decider         ai.Decider
	FleetStrength   int // 累積軍力(每回合由淨工業投資,好戰性格投更多)
	FleetInvestPool int // 造艦投資的餘數池(見 advanceAI):累積未達 invest 門檻的 NetIndustry,
	// 避免整數除法把小額淨工業直接捨去成 0(見 advanceAI 註解)。
	// Personality 是原版 AI 性格(AIRACES.CFG race_personality 0-6),開局依種族分布抽出。
	// 驅動關係演化、擴張積極度等行為差異——先前三個 AI 除了名字之外行為完全相同,
	// 因為所有性格相關的數字都是硬編的固定值(見 ai/personality_tables.go)。
	Personality ai.Personality
	Relation    int    // 對玩家的外交關係分數(驅動 17 級 RelationLevel 與態勢)
	StanceName  string // 目前對玩家態勢(中文;由 ai.DecideStance 推得)
	OwnedStars  int    // 已擴張佔領的星數(含母星)

	// Spies 是這個 AI 對手派來偷玩家科技的間諜數(見 spy.go advanceEspionage)。opt-in,
	// 新對局預設 0(Go 零值恰好是想要的預設值,無零值陷阱)。AI 目前用簡單週期政策自動增加
	// (見 advanceAI),不像玩家的 PlayerSpies 需要花 BC 呼叫 TrainSpy——AI 的訓練成本/BC
	// 限制未建模,是誠實簡化而非疏漏(見 spy.go 檔頭說明)。
	Spies int

	// LastRaidTurn 是這個 AI 上次對玩家發動突襲的回合(見 ai_attack.go),用來維持
	// aiRaidInterval 的最短間隔。0 = 從未突襲(Go 零值即想要的預設值:第一次突襲最早
	// 發生在 Turn >= aiRaidGraceTurns,不會因為零值提早)。
	LastRaidTurn int

	// WantsAudience 是「這位對手正在請求會談」(原版 `Humans_Requesting_Diplomacy_` 那個
	// 位元遮罩裡屬於它的那一位)。AudienceReason 是來意(宣戰/提議貿易/提議結盟)。
	// 見 audience.go。opt-in,零值 false = 沒有請求(新對局/舊存檔皆安全)。
	WantsAudience  bool
	AudienceReason string

	// ColonyStars 是 Colonies[i] 對應到 Stars 的索引(平行陣列),兩者長度須一致——aiExpand
	// append 新殖民地、InvadeColony 玩家攻陷 AI 殖民地時各自同步移除,見兩處函式。
	//
	// 2026-07-11 訂正:先前 aiExpand 每 5 回合佔領一顆無主星時只標記 Star.Owner=2、
	// OwnedStars++,並不會為那顆星建立真正的 engine.ColonyState,因此本欄位一度只有「開局
	// 母星」這一筆對映,其餘擴張出的星是「有旗標無殖民地模型」的版圖——地面入侵
	// (InvadeColony)打不到、AI 經濟(RunEmpireTurn)也不會因擴張成長。現在 aiExpand 改用
	// newColonyFromStar(colonization.go,與玩家 ColonizeStar 共用同一套建法)建立真殖民地並
	// 同步 append 進 Colonies/ColonyStars,擴張出的星此後都有實際殖民地模型,可被入侵、且
	// 會計入 AI 每回合的 TotalNetIndustry。
	ColonyStars []int

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

	// Leaders 是這個 AI 對手的領袖名單(欄位型別比照 GameSession.Leaders,同一個 shell.Leader
	// struct)。2026-07-11 新增,唯一消費端目前是 InvadeColony 的守方 Commando 加成
	// (commandoLeaderTier(aiPlayer.Leaders),見 ground_invasion.go)。
	//
	// ⚠ 誠實近似(#5,docs/tech/version-1.3-1.5-diff.md):原版領袖是從英雄池隨機雇用、可陣亡
	// 替換的動態資源;本 remake 沒有 AI 英雄雇用/成長系統,改用「開局依種族性格固定指派」當
	// 近似——布拉西人(手冊「體格強悍,地面戰加成」)給一名 Tier2 進階指揮官、姆瑞森人(好戰
	// 善攻)給一名 Tier1 一般指揮官、席隆人(重研究)不給指揮官。此清單 buildDemoAIOpponents
	// 開局建立後不隨遊戲成長變動(不模擬雇用/陣亡),與 commandoLeaderTier 對玩家 s.Leaders
	// 的「帝國全域清單當代理」同款近似紀律,非手冊逐字的隨機雇用機制。
	//
	// nil 安全:舊存檔沒有這個欄位時解碼為 nil,commandoLeaderTier(nil) 回傳 0(無 Commando
	// 加成),與加這個欄位之前的行為(TODO 留白、无加成)一致,不會 panic。
	Leaders []Leader
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
	Name     string
	Owner    int  // 0=無主 1=玩家 2=AI
	Explored bool // 艦隊是否曾抵達(已探索)
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
}

// Ship 是一艘艦艇(供艦隊畫面);Weapon/Armor/Shield/Special 為掛載的元件。
type Ship struct {
	Name                           string
	Class                          string // 艦體等級(護衛艦/巡洋艦/戰艦…)
	Weapon, Armor, Shield, Special string // 元件名
	WeaponAttack, BonusHP          int    // 武器攻擊加成、裝甲+護盾 HP 加成
	// Mods 是掛載在 Weapon 上的武器改造(gamedata.WeaponModCode 字串,如 "HV"/"PD"),
	// 只對 beam 武器生效(見 WeaponIsBeam / weapon_mods.go)。空切片/nil = 無改造(既有
	// 存檔沒有這個欄位,JSON 解碼會是 nil,行為與「無改造」完全一致,回歸安全)。
	Mods []string
	// Damage 是累積的結構損傷(0 = 完好)。remake 先前沒有艦艇損傷的概念——一艘船不是完好
	// 就是被擊沉——「自動修復」元件因此從加進來就沒有任何效果。見 repair.go。
	// 舊存檔沒有這個欄位,JSON 解碼為 0 = 完好,回歸安全。
	Damage int
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
		{"雷射", 20, 4, gamedata.TOPIC_PHYSICS, gamedata.TECH_LASER_CANNON},       // ResearchAll(早期)
		{"核飛彈", 30, 6, gamedata.TOPIC_CHEMISTRY, gamedata.TECH_NUCLEAR_MISSILE}, // ResearchAll(早期)
		{"質量投射器", 40, 8, gamedata.TOPIC_ADVANCED_MAGNETISM, gamedata.TECH_MASS_DRIVER},
		{"中子爆破槍", 60, 12, gamedata.TOPIC_NEUTRINO_PHYSICS, gamedata.TECH_NEUTRON_BLASTER}, // ✓ 值
		{"核融合光束", 80, 16, gamedata.TOPIC_FUSION_PHYSICS, gamedata.TECH_FUSION_BEAM},
		{"麥克萊特飛彈", 90, 17, gamedata.TOPIC_ADVANCED_CHEMISTRY, gamedata.TECH_MERCULITE_MISSILE},
		{"高斯砲", 120, 18, gamedata.TOPIC_SUBSPACE_FIELDS, gamedata.TECH_GAUSS_CANNON}, // ✓ 值 戰鬥最大
		{"相位砲", 160, 19, gamedata.TOPIC_MULTIPHASED_PHYSICS, gamedata.TECH_PHASOR},
		{"電漿砲", 200, 20, gamedata.TOPIC_PLASMA_PHYSICS, gamedata.TECH_PLASMA_CANNON}, // ✓ 值 1.50
		{"死光", 300, 25, gamedata.TOPIC_ARTIFICIAL_LIFE, 0},                           // 里程碑,proxy
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
	ArmorOptions = []Component{
		{"無裝甲", 0, 0, 0, 0},
		{"鈦裝甲", 30, 10, gamedata.TOPIC_CHEMISTRY, gamedata.TECH_TITANIUM_ARMOR}, // ResearchAll(早期)
		{"三鈦裝甲", 60, 20, gamedata.TOPIC_ADVANCED_METALLURGY, gamedata.TECH_TRITANIUM_ARMOR},
		{"佐特裝甲", 100, 35, gamedata.TOPIC_NANO_TECHNOLOGY, gamedata.TECH_ZORTRIUM_ARMOR},
		{"中子素裝甲", 160, 55, gamedata.TOPIC_MOLECULAR_MANIPULATION, gamedata.TECH_NEUTRONIUM_ARMOR},
		{"精金裝甲", 240, 80, gamedata.TOPIC_MOLECULAR_CONTROL, gamedata.TECH_ADAMANTIUM_ARMOR},
		{"氙素裝甲", 350, 120, gamedata.TOPIC_ARTIFICIAL_LIFE, 0}, // 里程碑,proxy
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
		// 硬化護盾:與隱形裝置同屬 TOPIC_DISTORTION_FIELDS(techtree.go 的三選一)。
		// 手冊兩處效果:每次攻擊額外減傷 3(gamedata.DamageHardShieldBonus,先前無元件載體
		// 等於死碼)、以及**星雲中護盾仍然可用**(見 nebula.go)。
		// ⚠ 成本 100 沿用同主題的隱形裝置,是 remake 值不是原版真值——本表其餘成本同樣是
		// remake 值(見表頭那幾筆的「proxy 待重設計」註記)。
		{"硬化護盾", 100, gamedata.DamageHardShieldBonus, gamedata.TOPIC_DISTORTION_FIELDS, gamedata.TECH_HARD_SHIELDS},
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

// shieldReduceByName 依護盾元件名回傳「每發減傷」(戰鬥用)。
// remake 由護盾階推導:無=0、第一級=2、第三級=4…第十級=10(精確 per-class 真值待逆向,
// 見 docs/tech/gameplay-systems-status.md);讓 DamageAfterShield 的護盾機制生效。
func shieldReduceByName(name string) int {
	for i, c := range ShieldOptions {
		if c.Name == name {
			return i * 2
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
func homeworldShips() []Ship {
	return []Ship{
		{Name: "拓荒號", Class: "殖民船", Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
		{Name: "先驅一號", Class: "偵察艦", Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
		{Name: "先驅二號", Class: "偵察艦", Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
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
	PlayerStart, EnemyStart   int // 開戰時雙方艦數
	PlayerWon                 bool
	PlayerLosses, EnemyLosses int
	Log                       []string // 逐回合戰報
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

// applyDamage 對艦隊(以各艦戰力當 HP)造成 dmg 傷害,由最弱艦起逐艘擊沉,回傳擊沉數。
// Difficulties 是難度選項(名稱 + 敵方戰力倍率),對應 NEW GAME 的 DIFFICULTY。
//
// **五級**是反組譯確認的:新遊戲畫面的難度選擇器建立時傳選項數 5
// (`sub_CCE2E` 裡 `push 5` + 變數 `word_1A1362`),而 `NEWGAME.LBX` 的難度圖正好也是 5 張
// (資產 4–8,五張手勢:伸手扶持 → 握手 → 比讚 → 握拳 → 雙拳相抵,難度遞增)。
// remake 先前只有四級,少的是最低的「教學」。
//
// ⚠ 這次補在**索引 0**,所以既有存檔的 Difficulty 索引整體位移一格(舊的「簡單」讀回來會變
// 「教學」,依此類推)。難度只影響敵方戰力倍率與 AI 關係漂移,舊存檔讀回來會變簡單一級,
// 不會壞掉。alpha 階段取「對齊原版」而非「保存索引」。
var Difficulties = []struct {
	Name string
	Mult float64
}{
	// 教學倍率 0.3:原版 Tutor 是「比 Easy 更寬鬆」的入門級,remake 沒有原版倍率表,
	// 取既有 Easy(0.6)的一半當保守值——這是 remake 調校值,不是原版數字。
	{"教學", 0.3}, {"簡單", 0.6}, {"普通", 1.0}, {"困難", 1.5}, {"不可能", 2.2},
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
// ⚠ 其餘效果仍未接。手冊已經給出可用的硬證,留給下一輪:
//   - 初始建築數上限:Pre-warp 3 / Average·Post-warp 5 / Advanced 9(不含 Capitol),
//     且 = min(⅔ 人口無條件進位, 上限)。remake 的母星建築是固定一組,還沒有依人口生成的機制。
//   - 開局已知科技領域數:Pre-warp 2 個(field 0 + Construction 第一個)、其餘 6 個。
//     remake 現在開局是 2 個(`newHomeworldPlayerState` 的 CompletedTopics)= **等同 Pre-warp**,
//     所以預設選 Average 時其實還沒拿到該有的 6 個領域。要補得先查出那 6 個是哪些
//     (手冊只說預設第一個是 field #29),沒有一手表之前不臆造。
//     (這一項與上面的 FTL 限制無關:FTL 只管「能不能出星系」,科技領域數管「開局會哪些技術」。)
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

// genEnemyFleet 依回合數 + 難度倍率生成敵方艦隊(戰力清單;越後期/越難越強)。
func genEnemyFleet(turn int, mult float64) []int {
	n := 2 + turn/3
	sizes := []int{2, 4, 8, 16}
	f := make([]int, 0, n)
	for i := 0; i < n; i++ {
		v := int(float64(sizes[i%len(sizes)]) * mult)
		if v < 1 {
			v = 1
		}
		f = append(f, v)
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
	kind    WeaponKind
	// mods 是攻方武器改造(gamedata.WeaponModCode 字串);只有 kind==WeaponKindBeam 時
	// battleVolley 會套用(見 WeaponIsBeam 判斷),敵方艦隊(genEnemyFleet)沒有個別武器
	// 設計,一律 nil(既有簡化,非本輪引入)。
	mods []string
}

// battleVolley 讓每個存活 attacker 對第一個存活 defender 射一發(固定近距 range=2),
// 依 attacker 的武器類型分流真戰鬥公式:beam 沿用 ResolveShot(不動,回歸測試見
// combat_weapon_kind_test.go);missile 改用 ResolveMissileShot(躲避/AMR 攔截);
// spherical 改用 ResolveSphericalShot(現行武器表暫無球形武器掛載,分支保留供未來串接)。
// 回傳本輪擊沉的 defender 數。移除陣亡艦。
func battleVolley(attackers []combatant, defenders *[]combatant, rng *rand.Rand) int {
	before := len(*defenders)
	for i := range attackers {
		ti := -1
		for j := range *defenders {
			if (*defenders)[j].hp > 0 {
				ti = j
				break
			}
		}
		if ti < 0 {
			break
		}
		d := &(*defenders)[ti]
		var shot ShotResult
		switch attackers[i].kind {
		case WeaponKindMissile:
			amrRoll := rng.Intn(100) + 1
			jamRoll := rng.Intn(100) + 1
			shot = ResolveMissileShot(false, 2, amrRoll, 0, 0, false, jamRoll,
				attackers[i].wmax, d.shield, d.armor, false)
		case WeaponKindSpherical:
			span := attackers[i].wmax - attackers[i].wmin
			r := 0
			if span > 0 {
				r = rng.Intn(span + 1)
			}
			aggD := gamedata.DamageSphericalRoll(attackers[i].wmin, r, 100)
			shot = ResolveSphericalShot(aggD, d.shield, d.armor, false, false)
		default:
			roll := rng.Intn(100) + 1
			net := attackers[i].atk - d.def
			shot = ResolveShotWithMods(net, attackers[i].wmin, attackers[i].wmax, 2, d.shield, d.armor, roll,
				false, weaponModCodes(attackers[i].mods))
		}
		if shot.Hit {
			d.armor = shot.RemainingArmorHP
			d.hp -= shot.DamageToStructure
		}
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
	for shipIdx, sh := range s.Fleet().Ships {
		if isSupportShipClass(sh.Class) {
			// 手冊 p.119 逐字:「Support ships … **do not fight**, but are destroyed if an enemy
			// military fleet chooses to engage them in combat.」——支援艦(殖民船/前哨船/運輸艦)
			// 不參與火力計算。先前它們會以最低戰力(shipStrength 回 1)混進戰列,等於帶著殖民船
			// 去打仗還能加一點傷害,並且在計算損失時第一個被打掉。
			//
			// 「被擊毀」那一半仍然成立:損失是由呼叫端的 removeWeakestShip 處理,支援艦戰力最低,
			// 打輸時本來就會先被移除,與手冊一致。
			continue
		}
		body := shipStrength(sh.Class)
		atk := body + sh.WeaponAttack
		atk += atk * s.RaceCombatPct / 100 // 種族戰鬥加成(姆瑞森+25、布拉西/阿爾卡里+15…)
		hp := body * 3
		// 戰機庫:出擊一隊戰機(手冊 GM p.127 出擊數:攔截機 4 / 重戰機 2),在艦級抽象結算中以
		// 母艦戰力加成承接整隊火力與血量(中隊規模手冊錨定、每架近似)。重戰機庫火力較強。
		switch sh.Special {
		case "戰機庫":
			fatk, fhp := gamedata.FighterBayCombatContribution()
			atk += fatk
			hp += fhp
		case "重戰機庫":
			fatk, fhp := gamedata.FighterHeavyBayCombatContribution()
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
		out = append(out, combatant{hp: hp, maxHP: shipMaxHP(sh), atk: atk, def: body, wmin: atk / 2, wmax: atk,
			shield: s.nebulaShield(shieldReduceByName(sh.Shield), shipHasHardShield(sh)),
			armor:  armorHPByName(sh.Armor),
			kind:   weaponKindByName(sh.Weapon), mods: sh.Mods,
			autoRepair: shipHasAutoRepair(sh), shipIdx: shipIdx})
		idx = append(idx, shipIdx)
	}
	return out, idx
}

// ResolveBattle 快速艦隊自動結算(無格子;供非互動戰鬥)。改用 gamedata 真戰鬥公式逐發解算,
// 與格子戰術戰鬥(tacticalScreen)一致:每回合雙方齊射,每發走命中判定→傷害→過盾→過甲。
func (s *GameSession) ResolveBattle(enemy string) BattleResult {
	mult := 1.0
	if s.Difficulty >= 0 && s.Difficulty < len(Difficulties) {
		mult = Difficulties[s.Difficulty].Mult
	}
	var ef []combatant
	for _, st := range genEnemyFleet(s.Turn, mult) {
		ef = append(ef, combatant{hp: st * 3, atk: st, def: st, wmin: st / 2, wmax: st, armor: st, shipIdx: -1})
	}
	pf, pfIdx := s.mkPlayerCombatantsIndexed()

	res := BattleResult{Enemy: enemy, PlayerStart: len(pf), EnemyStart: len(ef)}
	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + 12345)) // 依回合種子,可重現
	for round := 1; round <= 6 && len(pf) > 0 && len(ef) > 0; round++ {
		eDestroyed := battleVolley(pf, &ef, rng)
		pDestroyed := battleVolley(ef, &pf, rng)
		res.Log = append(res.Log, fmt.Sprintf("第 %d 回合:擊沉敵艦 %d ／ 我方損失 %d", round, eDestroyed, pDestroyed))
	}
	res.PlayerLosses = res.PlayerStart - len(pf)
	res.EnemyLosses = res.EnemyStart - len(ef)
	res.PlayerWon = len(ef) == 0 || len(pf) >= len(ef)
	// 倖存艦的剩餘血量寫回持久損傷(見 repair.go)。battleVolley 會就地移除陣亡艦,
	// 故 pf 剩下的是「倖存者、且保持原始相對順序」——與 pfIdx 的前綴對齊。
	s.applySurvivorDamage(pf, pfIdx)
	for i := 0; i < res.PlayerLosses; i++ {
		s.removeWeakestShip()
	}
	s.repairAfterBattle() // 自動修復元件/進階損害管制:戰後完全修復(手冊 p.80/p.82)
	s.LastBattle = &res
	return res
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
// raceDiploBonusPct 回傳目前種族的外交加成百分比。人類 Charismatic +50%(手冊 p.15「Humans
// gain a 50% bonus to their diplomatic efforts」)——這是移除人類假起始金後,其招牌特質首個
// 真正的消費端。first-version 只認人類(RaceIndex 0);自訂種族「魅力非凡」pick 尚未追蹤,回 0
// (誠實標)。
func (s *GameSession) raceDiploBonusPct() int {
	if s.RaceIndex == 0 { // 人類
		return 50
	}
	return 0
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

// DiplomacyResponse 處理玩家對某對手的外交動作(提議和平/貿易/威脅),**實際改變該 AI 的關係
// 分數**(先前只回文字不改狀態),並回一句回應。和平/貿易改善關係、威脅惡化;改善量受種族外交
// 加成放大(人類 Charismatic +50%,見 raceDiploBonusPct)。關係變化會反映到 races 畫面的態勢與
// 議會投票傾向(RelationLevelForScore/DecideStance)。
func (s *GameSession) DiplomacyResponse(action, enemy string) string {
	pPop, ePop := 0, 0
	for _, c := range s.PlayerColonies {
		pPop += c.Population
	}
	for _, a := range s.AIPlayers {
		for _, c := range a.Colonies {
			ePop += c.Population
		}
	}
	pFleet := 0
	for _, sh := range s.AllShips() { // 外交看的是**全帝國**軍力,不是眼前這一支
		pFleet += shipStrength(sh.Class)
	}
	ai := s.aiByDisplayName(enemy)
	diploPct := 100 + s.raceDiploBonusPct() // 人類 150、其餘 100
	switch action {
	case "peace":
		if ai != nil {
			ai.adjustRelation(15 * diploPct / 100) // 提議和平改善關係(Charismatic 放大)
		}
		if pPop >= ePop {
			return enemy + ":你們的實力我們敬佩,和平協議成立。(關係改善)"
		}
		return enemy + ":我們記下你的善意,但真正的和平言之過早。(關係略改善)"
	case "trade":
		if ai != nil {
			ai.adjustRelation(10 * diploPct / 100)
		}
		return enemy + ":貿易協定成立,願雙方繁榮昌盛。(關係改善)"
	case "threat":
		if ai != nil {
			ai.adjustRelation(-20)
		}
		if pFleet >= 10 {
			return enemy + ":……我們會記住這份侮辱。(關係惡化)"
		}
		return enemy + ":就憑你們這點艦隊?可笑!(關係惡化)"
	}
	return ""
}

// CombatShip 是格子戰術戰鬥中的一艘艦(有 HP + 格位 + 真戰鬥公式所需的攻防/傷害/盾甲)。
type CombatShip struct {
	Name      string
	HP, MaxHP int // 艦體結構 HP
	Attack    int // Beam Attack(BA,命中判定用)
	Col, Row  int // 格位(8 欄 × 6 列)
	// 以下供 ResolveShot 真戰鬥公式使用(remake 由艦艇設計推導,見 StartCombat 註記)。
	Defense         int        // 守方防禦(AF+BD),減 netAttack
	WeaponMin       int        // 單發最小傷害
	WeaponMax       int        // 單發最大傷害
	ShieldReduction int        // 護盾每發減傷
	HardShield      bool       // 硬化護盾(額外減傷 + 星雲中護盾仍可用,見 nebula.go)
	ArmorHP         int        // 裝甲 HP(結構外的緩衝,先耗盡才傷結構)
	Kind            WeaponKind // 武器戰鬥解算路徑(beam/missile/spherical,見 weapon_kind.go);
	// 敵方艦(genEnemyFleet)無個別武器設計資料,一律留零值 WeaponKindBeam(既有簡化)。
	Mods []string // 武器改造(gamedata.WeaponModCode 字串);只對 Kind==WeaponKindBeam 生效。
	// SpriteIdx 是 CMBTSHP.LBX 資產索引(含色塊偏移,45*色塊+艦級內索引),
	// 供戰術戰鬥畫面依艦級挑不同大小 sprite。見 docs/tech/cmbtshp-ship-sprites.md。
	SpriteIdx int
}

// CombatSpriteForClass 依艦體等級回傳 CMBTSHP 色塊內 sprite 索引(見 docs/tech/cmbtshp-ship-sprites.md)。
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

// StartCombat 依玩家艦隊 + 難度生成敵方,建立格子戰鬥雙方艦艇(HP=戰力×3、攻擊=戰力);
// 玩家艦置左欄、敵方置右欄,依序排列。
func (s *GameSession) StartCombat(enemy string) (player, enemyShips []CombatShip) {
	// 由艦艇設計推導真戰鬥公式所需數值(remake 近似;精確值需艦體空間格 + 元件佔格 + 軍官技能):
	//   結構 HP = 艦體×3;裝甲 HP = 設計 BonusHP;Beam Attack = 艦體 + 武器攻擊;
	//   防禦 = 艦體(小艦=低戰力=低防,趨勢近原版);單發傷害 min=max/2、max=Attack;
	//   護盾減傷暫 0(艦艇設計尚未把護盾與裝甲分離,見 gameplay-systems-status.md)。
	for i, sh := range s.Fleet().Ships {
		body := shipStrength(sh.Class)
		atk := body + sh.WeaponAttack
		player = append(player, CombatShip{
			Name: sh.Name, HP: body * 3, MaxHP: body * 3, Attack: atk, Col: 1, Row: i,
			Defense: body, WeaponMin: atk / 2, WeaponMax: atk,
			ShieldReduction: s.nebulaShield(shieldReduceByName(sh.Shield), shipHasHardShield(sh)),
			HardShield:      shipHasHardShield(sh),
			ArmorHP:         armorHPByName(sh.Armor),
			Kind:            weaponKindByName(sh.Weapon), Mods: sh.Mods,
			SpriteIdx: CombatSpriteForClass(sh.Class), // 色塊 0(玩家)
		})
	}
	mult := 1.0
	if s.Difficulty >= 0 && s.Difficulty < len(Difficulties) {
		mult = Difficulties[s.Difficulty].Mult
	}
	for i, st := range genEnemyFleet(s.Turn, mult) {
		enemyShips = append(enemyShips, CombatShip{
			Name: fmt.Sprintf("%s艦%d", enemy, i+1), HP: st * 3, MaxHP: st * 3, Attack: st, Col: 6, Row: i,
			Defense: st, WeaponMin: st / 2, WeaponMax: st, ShieldReduction: 0, ArmorHP: st,
			SpriteIdx: 45 + CombatSpriteForStrength(st), // 色塊 1(敵艦,與玩家色塊區隔)
		})
	}
	return
}

// ApplyCombatOutcome 依格子戰鬥後存活的玩家艦名,更新艦隊(移除陣亡艦)+ 記錄結果供結果畫面。
func (s *GameSession) ApplyCombatOutcome(enemy string, playerStart, enemyStart int, survivors map[string]bool, won bool) {
	f := s.Fleet() // 戰鬥打的是**參戰的那一支**艦隊
	kept := f.Ships[:0]
	for _, sh := range f.Ships {
		if survivors[sh.Name] {
			kept = append(kept, sh)
		}
	}
	f.Ships = kept
	s.LastBattle = &BattleResult{
		Enemy: enemy, PlayerStart: playerStart, EnemyStart: enemyStart, PlayerWon: won,
		PlayerLosses: playerStart - len(kept), EnemyLosses: enemyStart,
	}
}

// PrimaryEnemyName 回傳戰鬥/外交畫面顯示用的「主要對手」名稱。取第一個 AI 對手的種族名,
// 去掉 demoAIOpponentSetup 的「AI (…)」外殼(戰鬥標籤前綴接「艦N」時較自然,如「席隆人艦1」
// 而非「AI (席隆人)艦1」)。無 AI 對手時 fallback「敵軍」——避免舊硬編「賽隆人」(Races 表裡
// 根本不存在的錯字,見 demoAIOpponentSetup 註解)顯示在戰鬥畫面。
//
// 註:目前戰鬥為單一通用敵艦隊(genEnemyFleet),此名稱純顯示標籤,不綁定特定 AI 的實艦;
// 待「多 AI 目標選擇 UI」接上後改為玩家實際交戰的對手(見 remaining-work-roadmap 節 B)。
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
// (gamedata.WeaponSpaceWithMods)。mods 只在武器是 beam 路徑時生效(WeaponIsBeam)——
// 手冊的 HV/PD/AF/CO 明文只講 beam 武器,飛彈(核飛彈/麥克萊特飛彈)沒有這套 mod 掛鉤,
// 對非 beam 武器傳 mods 會被忽略,不誤加空間。
func ShipDesignSpaceUsedWithMods(class string, weapon, armor, shield, special int, mods []string) int {
	_ = armor // 見上方註解:手冊行為上裝甲不佔空間,顯式忽略以避免「未使用參數」誤解成疏漏
	_ = shield
	w := pick(WeaponOptions, weapon)
	sp := pick(SpecialOptions, special)
	classID, _ := shipClassFromName(class)
	hullSpace := gamedata.ShipHullSpace(classID)
	base := gamedata.WeaponSpaceByName[w.Name]
	used := base
	if len(mods) > 0 && WeaponIsBeam(w.Name) {
		used = gamedata.WeaponSpaceWithMods(base, weaponModCodes(mods))
	}
	used += gamedata.SpecialSpace(hullSpace, sp.Name != "" && sp.Name != "無")
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
	classID, _ := shipClassFromName(class)
	hullSpace := gamedata.ShipHullSpace(classID)
	return ShipDesignSpaceUsedWithMods(class, weapon, armor, shield, special, mods) <= hullSpace
}

// DesignCost 回傳一組元件選擇(艦體 + 武器/裝甲/護盾/特殊)的總生產成本(無武器改造)。
func DesignCost(class string, weapon, armor, shield, special int) int {
	return DesignCostWithMods(class, weapon, armor, shield, special, nil)
}

// DesignCostWithMods 同 DesignCost,套用武器改造對成本的影響(手冊「adds to the size AND
// cost」,與佔格用同一套百分比,見 gamedata.WeaponCostWithMods)。
func DesignCostWithMods(class string, weapon, armor, shield, special int, mods []string) int {
	w := pick(WeaponOptions, weapon)
	weaponCost := w.Cost
	if len(mods) > 0 && WeaponIsBeam(w.Name) {
		weaponCost = gamedata.WeaponCostWithMods(w.Cost, weaponModCodes(mods))
	}
	return ShipCost(class) + weaponCost + pick(ArmorOptions, armor).Cost +
		pick(ShieldOptions, shield).Cost + pick(SpecialOptions, special).Cost
}

// BuildShip 造一艘指定艦體 + 全元件(武器/裝甲/護盾/特殊)的艦:扣國庫總成本,加入艦隊。
// BC 不足回 false。武器加攻擊、裝甲+護盾加 HP、特殊「戰鬥電腦」再加攻擊。無武器改造。
func (s *GameSession) BuildShip(class string, weapon, armor, shield, special int) bool {
	return s.BuildShipWithMods(class, weapon, armor, shield, special, nil)
}

// BuildShipWithMods 同 BuildShip,額外把 mods(武器改造)存進造出的 Ship.Mods,並用
// DesignCostWithMods 算入改造增加/減少的成本。mods 對非 beam 武器無效(WeaponIsBeam),
// 但仍照玩家選擇存檔(不強制清空),避免玩家切換武器後 UI 狀態被意外抹除;佔格/傷害計算
// 端(ShipDesignSpaceUsedWithMods / battleVolley)各自已用 WeaponIsBeam 判斷是否套用。
func (s *GameSession) BuildShipWithMods(class string, weapon, armor, shield, special int, mods []string) bool {
	// 武器傷害(w.Value)吃這局遊戲的版本規則 profile(s.RuleProfile,見 BuildWeaponOptions 註解:
	// 電漿砲 1.3=30/1.5=20,其餘元件與套件級 WeaponOptions 逐一相同)——造艦時真正掛上版本相依
	// 傷害值,不再永遠是套件級硬編的 1.5 值。成本(DesignCostWithMods)不受影響:兩版電漿砲 Cost
	// 相同,差異只在 Value,見 ruleprofile.go RuleProfile.PlasmaCannonMaxDamage 註解。
	w, a, sh, sp := pick(BuildWeaponOptions(s.RuleProfile), weapon), pick(ArmorOptions, armor), pick(ShieldOptions, shield), pick(SpecialOptions, special)
	cost := DesignCostWithMods(class, weapon, armor, shield, special, mods)
	if s.Player.BC < cost {
		return false
	}
	s.Player.BC -= cost
	name := shipNamePool[s.ShipCount()%len(shipNamePool)] // 全帝國計數,不然分艦隊後會撞名
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
	f.Ships = append(f.Ships, Ship{Name: name, Class: class, Weapon: w.Name, Armor: a.Name, Shield: sh.Name,
		Special: sp.Name, WeaponAttack: atk, BonusHP: a.Value + sh.Value, Mods: modsCopy})
	return true
}

// ShiftColonyJob 在某殖民地把 1 名人口從 from 職務移到 to(f=農夫 w=工人 s=科學家);
// from 需有人。供殖民地人口重分配(影響下回合經濟)。
func (s *GameSession) ShiftColonyJob(idx int, from, to string) {
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
		*fp--
		*tp++
	}
}

// 銀河議會選舉勝利條件(成立門檻/2-3方候選/2/3多數/勝利判定)見 council.go
// ——2026-07-11 取代這裡原本「票數=人口、較高者當選」的簡化版(無成立門檻、無2/3多數、
// 未接勝利判定,對照 GAME_MANUAL.pdf p.183 手冊原文是錯誤示範,已移除)。

// ColonyBuild 是某殖民地目前的建造項目。
type ColonyBuild struct {
	Name     string
	Progress int
	Cost     int
}

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
	out = append(out, ColonyBuild{"", 0, 0})
	out = append(out, ColonyBuild{TradeGoodsBuildName, 0, 0})
	out = append(out, ColonyBuild{HousingBuildName, 0, 0})
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
	out := []ColonyBuild{{"", 0, 0}, {TradeGoodsBuildName, 0, 0}, {HousingBuildName, 0, 0}}
	for _, b := range gamedata.AvailableBuildings(completedTopics) {
		out = append(out, ColonyBuild{Name: b.NameZH, Progress: 0, Cost: b.ProductionCost})
	}
	for _, a := range gamedata.AvailableSpecialActions(completedTopics) {
		out = append(out, ColonyBuild{Name: a.NameZH, Progress: 0, Cost: a.ProductionCost})
	}
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
	if idx < 0 || idx >= len(s.Builds) {
		return
	}
	opts := availableBuildOptions(s.Player.CompletedTopics)
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
	case "研究實驗室": // Research Laboratory p.94:每科學家 +1 研究點 + 殖民地整體固定 +5 研究點。
		// 舊版 ResearchPerScientist+=5 把「固定 5 點」錯當成「每科學家 5 點」,訂正為 +1/科學家。
		c.ResearchPerScientist += 1
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
	case "行星超級電腦": // Planetary Supercomputer p.95:每科學家 +2 研究點(既有值相符,不動) + 固定 +10 研究點。
		c.ResearchPerScientist += 2
		c.FlatResearch += 10
	case "銀河網路中心": // Galactic Cybernet p.98:每科學家 +3 研究點(既有值相符,不動) + 固定 +15 研究點。
		c.ResearchPerScientist += 3
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
		// popGrowthThreshold(=300)這個 remake 尺度代表 1 人口單位,故 0.1 單位換算成
		// 這個尺度的 1/10;colonyGrowth 已依 Population<PopMax 判斷是否還要套用固定成長。
		c.FlatGrowth += popGrowthThreshold / 10
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
		//   - 異族管理中心:士氣計算路徑已預留(colonyMoralePercent 讀取此建築名),但因 remake
		//     無多種族人口追蹤,目前一律不套用多種族懲罰,故此建築在士氣上的效果暫不可見
		//     (詳見 colonyMoralePercent 註解)——不是假裝已完整建模,是誠實標記「架構已備、
		//     資料尚未跟上」。
		// 海軍陸戰隊營本來就有獨立的陸戰隊召兵系統(ground_invasion.go),現在額外貢獻
		// hasBarracks,兩個系統各自獨立生效,互不影響。
		//
		// 其餘 20 項(飛彈基地、戰機基地、地面砲台、再生反應爐、食物複製機、太空學院、自動實驗室、
		// 行星輻射/通量/屏障護盾、曲速力場干擾器、戰鬥站、星辰要塞、阿提米絲系統網、次元傳送門)
		// 手冊效果不對應 engine.ColonyState 既有欄位(艦艇駐防/軌道防禦等系統尚未建),暫不建模
		// ——僅由 advanceBuilds 記入 s.ColonyBuildings 為「已建」,顯示於畫面,不影響數值結算。
		// TODO:待對應遊戲系統(艦隊駐防/軌道防禦)建好後回頭補建模。
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
	if s.ColonyBuildings == nil {
		s.ColonyBuildings = make([]map[string]bool, len(s.PlayerColonies))
	}
	for i := range s.Builds {
		b := &s.Builds[i]
		if b.Name == "" || b.Cost == 0 {
			continue
		}
		ind := 0
		if i < len(s.LastPlayerOutput.Colonies) {
			// 稅金與建造搶同一份生產(手冊 GAME_MANUAL.pdf p.37:「Every rise in the tax rate
			// causes a corresponding drop in production」):稅率抽走 TaxRate% 的淨工業換 BC,剩下
			// (100-TaxRate)% 才用於建造。先前建造吃完整 NetIndustry、稅又另抽一次=稅金變免費錢
			// (非忠實),2026-07-12 校正扣掉稅金那份,使稅率成為真正的「更多錢 vs 更快建造」取捨。
			netInd := s.LastPlayerOutput.Colonies[i].NetIndustry
			ind = netInd * (100 - s.Player.TaxRate) / 100
		}
		b.Progress += ind
		if b.Progress >= b.Cost {
			if i < len(s.ColonyBuildings) {
				if _, isSpecial := gamedata.SpecialActionByNameZH(b.Name); isSpecial {
					// Special 一次性行動(地形改造/蓋亞轉化/土壤改良/運輸艦隊):刻意不記入
					// ColonyBuildings。手冊明講地形改造「可以套用好幾次」("You can terraform a
					// planet several times"),運輸艦隊同樣可反覆建造(每次都再 +5 艘),若記入
					// ColonyBuildings,第二次套用會被下面「已建過就不再套用效果」的 dedup 判斷
					// 擋下,不符手冊——見 gamedata/special_actions.go 檔頭說明。
					s.applySpecialAction(i, b.Name)
				} else {
					if s.ColonyBuildings[i] == nil {
						s.ColonyBuildings[i] = make(map[string]bool)
					}
					if !s.ColonyBuildings[i][b.Name] {
						s.ColonyBuildings[i][b.Name] = true
						s.applyBuildingEffect(i, b.Name) // 首次完工才套用長期效果
						s.recalcColonyMorale(i)          // 士氣建築(全息模擬艙/歡樂穹頂)或 Barracks 完工需重算士氣
					}
				}
			}
			s.LastBuilt = append(s.LastBuilt, fmt.Sprintf("殖民地 %d 完成建造:%s", i+1, b.Name))
			*b = ColonyBuild{} // 完成清空
			s.popNextBuild(i)  // 佇列有排隊項就自動接上(原版 7 格 BUILD QUEUE 行為)
		}
	}
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
		s.deliverNewShip(i, Ship{Name: s.nextSupportShipName(ColonyShipClass), Class: ColonyShipClass,
			Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"})

	case gamedata.OutpostShipActionName: // 前哨船完工 → 進玩家艦隊(見 outpost.go)
		s.deliverNewShip(i, Ship{Name: s.nextSupportShipName(OutpostShipClass), Class: OutpostShipClass,
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
	c := &s.PlayerColonies[i]
	old := c.Climate
	c.FoodPerFarmer += gamedata.ClimateFoodPerFarmer(next) - gamedata.ClimateFoodPerFarmer(old)
	c.PopMax = gamedata.TerraformPopMaxAfterClimateChange(c.PopMax, old, next)
	c.Climate = next
}

// Leader 是一名可雇用的軍官/領袖(供軍官列表)。
type Leader struct {
	Name  string
	Skill string // 專長(中文顯示標籤;技能效果透過 leaderSkillIDByName 對應到 gamedata 技能 id)
	Level int    // 顯示等級(1..5,對照 openorion2 MAX_LEADER_LEVELS=5 顯示慣例:1=最低、5=最高)。
	// 換算成 gamedata.LeaderSkillBonus 用的 expLevel(0..4)時用 leaderDisplayLevelToExpLevel(Level)。
	// 這是 demo 資料的既有欄位語意(非 HERODATA 真實經驗值,是直接指定的顯示等級)。
	Ship bool // true=艦艇軍官,false=殖民地領袖

	// Tier 技能階(0 無/1 一般/2 進階),對照 openorion2 Leader::hasSkill 的回傳值。demoLeaders
	// 皆為手動指定的示範資料(非 HERODATA 真實英雄),保守預設 1(一般技能),不臆造「進階」。
	Tier int
}

// demoLeaders 是示範領袖名單(固定;正式版由 HERODATA.LBX 真英雄資料填)。
func demoLeaders() []Leader {
	return []Leader{
		{"馮·諾伊曼", "科學家", 5, false, 1},
		{"洛克斐勒", "貿易家", 4, false, 1},
		{"漢尼拔", "指揮官", 6, true, 1},
		{"圖靈", "工程師", 3, true, 1},
	}
}

// leaderSkillIDByName 把 demoLeaders 的中文技能標籤對應到 gamedata 技能 id(openorion2
// gamestate.h enum LeaderSkills)。只收「名稱與技能語意清楚一對一」的項目:
//   - 科學家 → SKILL_RESEARCHER(common,officer.cpp skillFormatStrings row0 idx6 格式 "%+d",
//     為固定研究點數加成,非百分比)。
//   - 貿易家 → SKILL_TRADER(common,row0 idx9 格式 "%+d%%",收入百分比加成)。
//   - 工程師 → SKILL_ENGINEER(captain,row1 idx0 格式 "%+d%%",真實效果是艦艇維修速率加成——
//     remake 目前沒有艦艇維修系統,故技能 id 對應清楚但效果暫不生效,見 applyLeaderShipBonuses
//     的 TODO 註解)。
//
// 「指揮官」(漢尼拔)刻意不收錄在這張表:這張表只服務「殖民地經濟被動加成」
// (applyLeaderColonyBonuses,固定研究點數/收入百分比這種格式化數值)。2026-07-11 已確定
// 「指揮官」對應 gamedata.SKILL_COMMANDO(手冊 p.135 Commando,地面戰鬥系統),但它的消費端
// 是地面戰 force 加成而非殖民地經濟欄位,故改在 internal/shell/ground_invasion.go 用獨立的
// commandoLeaderTier(leaders []Leader) 依 l.Skill=="指揮官" 直接掃描、不透過本表——避免把
// 語意/單位都不同的兩套加成(經濟 vs 地面戰鬥)混進同一張映射表。SKILL_WEAPONRY 等其餘候選
// 已不採用(Commando 已定案)。
var leaderSkillIDByName = map[string]int{
	"科學家": int(gamedata.SKILL_RESEARCHER),
	"貿易家": int(gamedata.SKILL_TRADER),
	"工程師": int(gamedata.SKILL_ENGINEER),
}

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
// 呼叫端傳 &session.PlayerColonies[0])。只接「對應到 engine.ColonyState 既有欄位」的技能:
//   - SKILL_RESEARCHER(固定研究點數,格式 "%+d")→ FlatColony 整體固定加成 ColonyState.FlatResearch
//   - SKILL_TRADER(收入百分比,格式 "%+d%%")→ ColonyState.IncomeBonusPercent(與太空港/
//     證券交易所等建築的百分比加成同一欄位,可疊加)
//
// 其餘有對應 skill id 但 remake 尚無承接系統的技能(如 SKILL_SCIENCE_LEADER/
// SKILL_FINANCIAL_LEADER/SKILL_LABOR_LEADER 等 admin 技能──demoLeaders 目前沒有領袖標成這些
// 名稱,故不處理)一律略過,不臆造欄位。
func applyLeaderColonyBonuses(leaders []Leader, colony *engine.ColonyState) {
	for _, l := range leaders {
		if l.Ship {
			continue // 艦艇軍官不影響殖民地,見 applyLeaderShipBonuses
		}
		id, ok := leaderSkillIDByName[l.Skill]
		if !ok {
			continue // 無法對應的技能標籤(如「指揮官」),誠實跳過
		}
		expLevel := leaderDisplayLevelToExpLevel(l.Level)
		bonus := gamedata.LeaderSkillBonus(id, l.Tier, expLevel)
		switch id {
		case int(gamedata.SKILL_RESEARCHER):
			colony.FlatResearch += bonus
		case int(gamedata.SKILL_TRADER):
			colony.IncomeBonusPercent += bonus
		// SKILL_ENGINEER(工程師)等其餘已知 id 若指派給非 Ship 領袖,目前無殖民地承接欄位,
		// 略過不加總(不應該發生:工程師是 captain 技能,demoLeaders 裡標工程師的都是 Ship=true;
		// 這裡防禦性略過,避免未來資料改動時誤加到不相關欄位)。
		default:
		}
	}
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

	// SystemBodies 是**已淘汰的欄位**(2026-08-07,第 63 項)。
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
	// 不再與 `Stars` 平行(所有呼叫端已於第 62 項改走存取器)。
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
// ⚠ **已淘汰**:這一支看的是 `Planet.SystemBodies`,而那個欄位自 2026-08-07(第 63 項)
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
	Name         string // 中文名
	EnName       string // 英文名(對應 ai/original.go 種族性格)
	IndBonus     int
	ResBonus     int
	FoodBonus    int
	GrowthPct    int
	StartBC      int
	IncomePerPop int
	CombatPct    int
	Desc         string
	EnDesc       string // 英文模式的同一段說明(remake 自撰摘要的英譯,非原版文案)
}

// Races 是 MOO2 十三經典種族(招牌特性依手冊 p.15-16 校正,見 Race 型別註解)。索引 0 為人類(預設)。
var Races = []Race{
	{"人類", "Humans", 0, 0, 0, 0, 0, 0, 0, "外交手腕高明,雇用領袖較廉(民主政府)", "Skilled diplomats; leaders come cheap (Democracy)"},
	{"席隆", "Psilons", 0, 2, 0, 0, 0, 0, 0, "創造性研究,科學家產出高", "Creative researchers; high output per scientist"}, // ResBonus+2:手冊 p.614「2 more than galactic norm」,norm3+2=5,對齊 SAVE10.GAM Psilon 母星每科研=5
	{"薩克拉", "Sakkra", 0, 0, 1, 30, 0, 0, 0, "繁殖迅速,人口成長加成", "Prolific breeders; bonus to population growth"},
	{"克拉肯", "Klackons", 2, 0, 0, 0, 0, 0, 0, "團結勤奮,工業產出高", "Unified and industrious; high factory output"},
	{"姆瑞森", "Mrrshan", 0, 0, 0, 0, 0, 0, 25, "好戰善攻,艦艇攻擊加成", "Warlike marksmen; bonus to ship attack"},
	{"布拉西", "Bulrathi", 0, 0, 0, 0, 0, 0, 15, "體格強悍,地面與戰鬥加成", "Powerful build; bonus to ground and ship combat"},
	{"阿爾卡里", "Alkari", 0, 0, 0, 0, 0, 0, 15, "飛行天賦,艦艇迴避加成", "Born pilots; bonus to ship defense"},
	{"梅克拉", "Meklars", 1, 1, 0, 0, 0, 0, 0, "半機械,工業與研究兼具", "Cybernetic; strong in both industry and research"},
	{"達洛克", "Darloks", 0, 0, 0, 0, 0, 0, 0, "潛伏間諜,擅長滲透與隱形", "Shapeshifting spies; infiltration and stealth"},
	{"崔拉里安", "Trilarians", 0, 0, 1, 10, 0, 0, 0, "水棲民族,食物與成長加成", "Aquatic; bonus to food and population growth"},
	{"埃雷里安", "Elerians", 0, 1, 0, 0, 0, 0, 15, "心靈感應,研究與戰鬥", "Telepathic; strong in research and combat"},
	{"諾蘭姆", "Gnolams", 0, 0, 0, 0, 0, 2, 0, "幸運富商,每人口每回合額外進帳", "Lucky traders; extra income per unit of population"}, // IncomePerPop=2 半BC=+1 BC/人:手冊 p.16「each unit of Gnolam population generates an additional 1 BC per turn」(=money3 pick)
	{"矽基", "Silicoids", 1, 0, 0, -20, 0, 0, 0, "岩石生命,耐任何環境但成長慢", "Lithovore; immune to environment but slow to grow"},
}

// ApplyRace 把 Races[idx] 的起始加成套到玩家帝國:各殖民地每單位產出加成、額外起始國庫、
// 記錄成長/戰鬥百分點(供 advancePopulation/戰鬥使用)。只在新遊戲開局套一次。
func (s *GameSession) ApplyRace(idx int) {
	if idx < 0 || idx >= len(Races) {
		return
	}
	r := Races[idx]
	s.RaceIndex = idx
	s.raceGrowthPct = r.GrowthPct
	s.RaceCombatPct = r.CombatPct
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].IndustryPerWorker += r.IndBonus
		s.PlayerColonies[i].ResearchPerScientist += r.ResBonus
		s.PlayerColonies[i].FoodPerFarmer += r.FoodBonus
		s.PlayerColonies[i].IncomePerPop += r.IncomePerPop // 種族「錢」特質(諾蘭姆每人+1BC/回合)
	}
	s.Player.BC += r.StartBC
}

// ApplyCustomRaceBonuses 套用自訂種族(Custom Race)聚合出的數值加成。
// 加成來自 docs/tech/custom-race-picks.md 的官方 patch 1.5 點數值(生產/成長/戰鬥/國庫)。
// ⚠ 政府型態與特殊能力的深層效果(創造力科技解鎖、貿易奇才、心靈感應等)尚未模擬,
// 目前只套用可對應到 Race 欄位的數值部分;其餘由 Custom 畫面記錄待後續實作。
func (s *GameSession) ApplyCustomRaceBonuses(r Race) {
	s.raceGrowthPct = r.GrowthPct
	s.RaceCombatPct = r.CombatPct
	for i := range s.PlayerColonies {
		s.PlayerColonies[i].IndustryPerWorker += r.IndBonus
		s.PlayerColonies[i].ResearchPerScientist += r.ResBonus
		s.PlayerColonies[i].FoodPerFarmer += r.FoodBonus
		s.PlayerColonies[i].IncomePerPop += r.IncomePerPop
	}
	s.Player.BC += r.StartBC
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
	Name    string
	EnName  string
	R, G, B uint8
}{
	{"紅", "Red", 200, 95, 90},
	{"黃", "Yellow", 205, 185, 45},
	{"綠", "Green", 90, 165, 70},
	{"銀", "Silver", 195, 195, 205},
	{"藍", "Blue", 145, 185, 220},
	{"棕", "Brown", 200, 155, 110},
	{"紫", "Purple", 200, 145, 195},
	{"橙", "Orange", 225, 140, 45},
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
// (engine.ColonyState.MoralePercent 的來源;數值一律用 gamedata/morale.go 手冊常數,不自行
// 杜撰)。buildings 是該殖民地的 ColonyBuildings 項目(nil 視為尚無任何建築,map 讀取安全)。
//
// 已套用的來源:
//  1. gamedata.MoraleGovernmentBase(gov, hasBarracks)——hasBarracks 依手冊 p.76-79:
//     海軍陸戰隊營(Marine Barracks)或裝甲營房(Armor Barracks)其一即可解除
//     封建/獨裁/統一政府「無 Barracks -20%」的懲罰。
//  2. 全息模擬艙(Holo Simulator)已建 → +gamedata.MoraleHoloSimulatorBonus(+20,p.95-96)。
//  3. 歡樂穹頂(Pleasure Dome)已建 → +gamedata.MoralePleasureDomeBonus(+30,p.97-98)。
//
// 誠實列出「未套用」的手冊來源(不假裝精確,詳見呼叫端 ApplyGovernment/advanceBuilds 註解):
//   - Virtual Reality Network(全帝國 +20%,p.97-98):手冊定性為「成就」而非一般建築,不在
//     gamedata.Buildings 清單、remake 也無「成就」追蹤系統,無從得知是否擁有,故不套用。
//   - 多種族懲罰 gamedata.MoraleMultiRacialPenalty:remake 的 ColonyState 沒有「殖民地人口是否
//     含未同化外族血統」這個狀態(Population/Farmers/Workers/Scientists 只是職務數字,不分血統
//     來源),故無法判斷是否該套用,保守視為「單一種族」一律不套用——異族管理中心已建/未建在
//     此近似下暫無可見差異(與行星重力產生器在 demo session 暫不可見同一類「架構已備、資料
//     尚未跟上」情形,見 colony-buildings.md)。
//   - 首都淪陷懲罰 gamedata.MoraleCapitalCapturedPenalty:remake 沒有「首都被攻陷」這個狀態,
//     TODO 待地面入侵系統擴充到「可攻佔玩家母星」後補上,現在不加。
func colonyMoralePercent(gov gamedata.MoraleGovernmentType, buildings map[string]bool) int {
	hasBarracks := buildings["海軍陸戰隊營"] || buildings["裝甲營房"]
	pct := gamedata.MoraleGovernmentBase(gov, hasBarracks)
	if buildings["全息模擬艙"] {
		pct += gamedata.MoraleHoloSimulatorBonus
	}
	if buildings["歡樂穹頂"] {
		pct += gamedata.MoralePleasureDomeBonus
	}
	return pct
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
	s.PlayerColonies[i].MoralePercent = colonyMoralePercent(s.Government, s.buildingsFor(i))
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
	pct150 := func(v int) int { return (v*3 + 1) / 2 } // ×1.5 四捨五入
	if gov >= 0 && gov < len(moraleGovByIndex) {
		s.Government = moraleGovByIndex[gov]
	}
	for i := range s.PlayerColonies {
		switch gov {
		case 1: // 封建:研究減半
			s.PlayerColonies[i].ResearchPerScientist /= 2
		case 2: // 統一:食物 +50%、產能 +50%
			s.PlayerColonies[i].FoodPerFarmer = pct150(s.PlayerColonies[i].FoodPerFarmer)
			s.PlayerColonies[i].IndustryPerWorker = pct150(s.PlayerColonies[i].IndustryPerWorker)
		case 3: // 民主:研究 +50%
			s.PlayerColonies[i].ResearchPerScientist = pct150(s.PlayerColonies[i].ResearchPerScientist)
		}
	}
	s.recalcAllColonyMorale()
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
	galaxy, aiHomeStars := genGalaxy(stars, seed, numAI, s.galaxyAge())
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
	s.AIPlayers = buildDemoAIOpponents(aiHomeStars, s.Difficulty, seed)
	s.PlayerSpies = make([]int, len(s.AIPlayers)) // 平行 AIPlayers,重置為全新對手的間諜數(開局皆 0)
	s.PlayerColonyStars = []int{0}
	s.Fleet().AtStar = 0
	s.Fleet().DestStar = -1
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
func genGalaxy(n int, seed int64, aiHomes int, age gamedata.GalaxyAge) ([]Star, []int) {
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
	names := append([]string(nil), randomStarNamePool...)
	r.Shuffle(len(names), func(i, j int) { names[i], names[j] = names[j], names[i] })
	for gy := 0; gy < rows && idx < n; gy++ {
		for gx := 0; gx < cols && idx < n; gx++ {
			x := (float64(gx) + 0.15 + r.Float64()*0.7) / float64(cols)
			y := (float64(gy) + 0.15 + r.Float64()*0.7) / float64(rows)
			nm := names[idx%len(names)]
			if idx >= len(names) {
				nm = fmt.Sprintf("%s-%d", nm, idx/len(names)+1)
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
			stars = append(stars, Star{X: x, Y: y, Spectral: spectral, Size: r.Intn(4), Name: nm, Owner: owner, Wormhole: -1})
			idx++
		}
	}
	return stars, aiIdx
}

// GameSession 是一局進行中的遊戲狀態。玩家操作改變狀態,EndTurn 推進一回合(結算玩家 + 各 AI)。
type GameSession struct {
	Turn           int
	Player         engine.PlayerState
	PlayerColonies []engine.ColonyState
	AIPlayers      []AIOpponent
	// AIRelations 是 AI 對手彼此的外交關係矩陣(AIRelations[i][j] = AI i 對 AI j 的關係分數,
	// 夾 -40..40,同 AIOpponent.Relation 尺度)。先前 AI 只對玩家單向有關係;此矩陣讓多帝國彼此
	// 也有關係(依相對軍力漂移),使星系「活起來」並可支撐議會第三方搖擺票(見 advanceAIDiplomacy)。
	// 對稱維護(i↔j 各記一半視角)。長度 = len(AIPlayers)。
	AIRelations [][]int
	// History 是逐回合的全帝國國力快照(原版 module 122 Record_History_ 的對應物,
	// 供 INFO 的 History Graph 子畫面畫折線;見 history.go 檔頭)。
	History []HistoryTurn
	// LastEventReport 是上一回合觸發的隨機事件(結構化,供事件畫面用);nil = 本回合無事件。
	// 與 LastEvent(純文字,回合摘要用)同時寫入,見 events.go advanceEvents。
	// 比照 LastEvent 是「下回合會重算的顯示暫態」,刻意不存檔。
	LastEventReport *EventReport

	// LastDiscovery 是上一回合抵達星系時觸發的一次性發現(太空殘骸/失散殖民地…);
	// nil = 無。與 LastEventReport 分開:隨機事件是全銀河的新聞,發現是自家艦隊的回報。
	LastDiscovery    *SystemDiscovery
	LastPlayerOutput engine.EmpireOutput // 上一回合玩家結算(供畫面顯示)
	Stars            []Star              // 星系圖
	Nebulae          []Nebula            // 星雲(星圖地形,影響戰鬥護盾;見 nebula.go)
	// nebulaProbe 判定某個正規化座標是否落在星雲內。要讀星雲圖的遮罩,而本套件不碰資產,
	// 所以由 cmd/moo2 用 SetNebulaProbe 裝進來。**未匯出 = 不進存檔**,讀檔後要重裝。
	nebulaProbe       func(x, y float64) bool
	Planets           []Planet // 行星列表
	Leaders           []Leader // 已雇用的軍官/領袖名單(Leader Pool)
	MercPool          []Leader // 目前上門可雇用的傭兵領袖(手冊 p.134,見 advanceMercOffers/HireMerc)
	MercOfferedIdx    int      // 已釋出過的傭兵候選數(遞增指標,避免重複 offer 同一候選)
	MercCandidatePool []Leader // 傭兵候選池(cmd 層由 HERODATA.LBX 真英雄注入,見 SetMercCandidates);nil=用內建策展名單
	// Fleets 是玩家的艦隊清單,SelectedFleet 是目前操作中的那一支(見 fleet.go)。
	// 先前這裡是「Ships + FleetAtStar/DestStar/ETA/Marines/Tanks」一組欄位,
	// 也就是全帝國只有一支艦隊——那個限制卡住了星圖的遷移連線層與 F1/F2 循環。
	//
	// ⚠ 不變量:**永遠至少有一支艦隊**(可以沒有船)。由 ensureFleet 維持,
	// Fleet() 因此可以無條件回傳可寫指標,呼叫端不必逐處 nil 檢查。
	Fleets        []Fleet
	SelectedFleet int
	// ColonyRelocateTo 是各殖民地的集結點星索引(平行 PlayerColonies,見 relocation.go)。
	// ⚠ 預設必須是 −1 不是 Go 零值——0 是母星的索引。
	ColonyRelocateTo []int
	// ShowRelocationLines 是星圖遷移連線的顯示開關(原版 `byte_199BE4`,手冊那組設定裡的一項)。
	// 預設開:原版新開一局是畫的(`sub_127E1` 初始化時寫 1)。
	ShowRelocationLines bool
	LastBattle          *BattleResult // 上一場戰鬥結果(供戰鬥結果畫面)
	SelectedStar        int           // 星圖選中的星索引(-1=未選)
	Difficulty          int           // 難度索引(shell.Difficulties)
	Builds              []ColonyBuild // 各殖民地「當前建造中」的項目(對應 PlayerColonies;佇列見 BuildQueue)
	// BuildQueue[i] 是殖民地 i 的**後續**建造排隊項(不含 Builds[i] 那一格)。
	// 原版殖民地畫面的 BUILD QUEUE 是 7 格(反組譯 Add_Build_Queue_Fields_ 確認),
	// 完工自動接下一項;remake 先前只有一格。見 buildqueue.go 檔頭。
	BuildQueue [][]ColonyBuild
	LastBuilt  []string // 上回合完成的建造(供回合摘要)
	popAccum   []int    // 各殖民地人口成長累加值(達門檻則 +1 人口)

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

	// FleetTanks / PlayerColonyTanks / ArmorBarracksAge:裝甲營房(Armor Barracks)戰車營
	// 駐軍系統,與上面三個 Marine 對應欄位對稱(見 advanceArmor/LoadTanks,ground_invasion.go)。
	PlayerColonyTanks []int      // 各玩家殖民地 Armor Barracks 駐軍池(平行 PlayerColonies)
	ArmorBarracksAge  []int      // 各玩家殖民地 Armor Barracks 已運作回合數(平行 PlayerColonies)
	EventSeed         int64      // 隨機事件亂數種子(可重現;新遊戲遞增)
	LastEvent         string     // 本回合觸發的隨機事件描述(空=無事件;供回合摘要)
	DisableEvents     bool       // 關閉隨機事件(供確定性經濟測試隔離)
	eventRand         *rand.Rand // 事件亂數源(由 EventSeed 惰性建立)
	AntaresRaids      int        // 已發生的安塔蘭突襲次數(逐次升級強度)
	LastAntares       string     // 本回合安塔蘭突襲描述(空=無;供回合摘要)
	// Monsters 是星圖上守衛星系的太空怪獸(見 monster.go)。清空 = 全部已被清除。
	Monsters []MonsterGuard

	// CapturedPop 是玩家透過地面入侵俘虜來的人口單位總數(見 ground_invasion.go
	// InvadeColony)。原版計分對俘虜人口另有一份加分(手冊 p.184「You also get a premium
	// for captured population units」),見 score.go。
	CapturedPop int

	// PersistentEvents 是進行中的持續型隨機事件(超新星倒數/時空異象/超空間獸,
	// 見 events_persistent.go)。空 = 沒有任何持續事件。
	PersistentEvents []PersistentEvent

	// Outposts 是玩家的軍事前哨站(見 outpost.go)。與 PlayerColonies **完全分開**——
	// 手冊 p.85「produces nothing」,前哨站沒有人口也沒有產出,混進殖民地陣列會讓帝國經濟
	// 憑空多出一個殖民地。
	Outposts []Outpost

	// LastRaid / LastRaidReport 是本回合 AI 對玩家殖民地的突襲(見 ai_attack.go);
	// 空/nil = 無。與 LastAntares 分開:安塔蘭人是週期腳本,AI 突襲是外交/軍備的後果。
	LastRaid       string
	LastRaidReport *AIRaidReport
	RaceIndex      int    // 玩家選定的種族(shell.Races 索引)
	PlayerName     string // 玩家帝國/領袖名稱(新遊戲命名畫面設定)
	FlagColor      int    // 玩家旗幟顏色索引(shell.FlagColors)
	RaceCombatPct  int    // 種族戰鬥戰力百分點加成(供戰鬥使用)
	raceGrowthPct  int    // 種族人口成長百分點加成(供 advancePopulation)

	// Government 是玩家目前政府型態(2026-07-11 接線,供 colonyMoralePercent 士氣計算用)。
	// 由 ApplyGovernment 設定;新遊戲若從未呼叫 ApplyGovernment,預設見 NewDemoSession
	// (獨裁/Dictatorship,對應自訂種族 0 點基準)。
	//
	// Go 零值陷阱(比照 ColonyState.PlanetGravity 同款註解):gamedata.MoraleGovernmentType 的
	// 零值是 MoraleGovFeudalism(iota 從 0 開始,見 morale.go enum 順序),不是想要的預設政府
	// Dictatorship——任何建構 GameSession 卻沒有明確設定本欄位的呼叫端,會被誤判為封建政府,
	// 必須明確賦值,不能依賴零值。
	Government gamedata.MoraleGovernmentType

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
	Victory                VictoryState     // 遊戲是否已分出勝負(Over=true 後不再產生新的議會選舉)
	PendingCouncilElection *CouncilElection // 非玩家當選、等待玩家 RespondToCouncilElection 回應(手冊:議會無法強迫玩家接受)
	LastCouncil            string           // 本回合議會動態描述(空=無;供回合摘要)
	CouncilMeetings        int              // 已召開過的議會屆數
	lastCouncilTurn        int              // 上次召開議會的回合數(0=從未召開)

	// AntaranHomeworldConquered 是手冊三條勝利路徑之二「攻陷安塔蘭母星」的達成旗標(見
	// antaran_victory.go)。由 AssaultAntares 戰勝後設為 true;engine.CheckAntaranVictory 讀取
	// 這個布林旗標判定(該函式本身不追蹤戰鬥流程,見其註解),advanceAntaranVictory 依此設定
	// s.Victory。Go 零值(false)即想要的預設值,無零值陷阱。
	AntaranHomeworldConquered bool

	// --- 間諜(見 spy.go,最小可玩迴圈:只做偷科技 STEAL,見該檔檔頭說明) ---
	// PlayerSpies 是玩家派駐到 AIPlayers[i] 的間諜數(平行 AIPlayers)。opt-in,預設 0
	// (Go 零值即想要的預設值)。玩家經 TrainSpy(idx) 花 BC 增加;逐對手分配已經是這個陣列
	// 天然支援的結構,只是目前唯一一個 AI 對手時看不出差異。
	PlayerSpies []int

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
	LastEspionage []string
	spyRand       *rand.Rand // 間諜擲骰亂數源(由 EventSeed 惰性建立,比照 eventRand 慣例)
	discoveryRand *rand.Rand // 星系發現擲骰亂數源(見 discovery.go discoveryRoll,同上慣例)
}

// 安塔蘭人入侵參數:MOO2 的週期性終局威脅。前期寬限,之後每隔數回合一次突襲,強度隨次數升級。
const (
	antaresStartTurn = 20 // 前 20 回合寬限,不觸發
	antaresInterval  = 15 // 之後每 15 回合一次突襲
)

// advanceAntares 處理安塔蘭人週期性入侵:達排程回合觸發一次突襲,強度隨突襲次數升級,
// 對一殖民地造成人口損失 + 國庫掠奪(效果有界:人口不低於1、BC 不為負)。若玩家艦隊在母星
// 且有戰力,視為部分防禦、減半損失。結果記於 LastAntares(供回合摘要)。可用 DisableEvents 關閉。
func (s *GameSession) advanceAntares() {
	s.LastAntares = ""
	if s.DisableEvents {
		return
	}
	if s.Turn < antaresStartTurn || (s.Turn-antaresStartTurn)%antaresInterval != 0 {
		return
	}
	s.AntaresRaids++
	sev := s.AntaresRaids // 升級係數
	popLoss := 1 + sev/2
	bcLoss := 30 * sev

	// 母星防禦:艦隊在母星且有戰力則減半損失。
	// **任何一支**停在母星的艦隊有戰力就算防禦成功——不限玩家目前選中的那一支。
	defended := false
	for f := range s.Fleets {
		if s.Fleets[f].AtStar != 0 {
			continue
		}
		for _, sh := range s.Fleets[f].Ships {
			if shipStrength(sh.Class) > 0 {
				defended = true
				break
			}
		}
	}
	if defended {
		popLoss = (popLoss + 1) / 2
		bcLoss /= 2
	}

	// 夾在 [0, 現有 BC] 之間:忠實 yield 經濟(零緩衝)下 BC 可能已經是負值(見
	// docs/tech/colony-economy-maintenance.md),若只判斷「bcLoss > s.Player.BC」,
	// s.Player.BC 為負時 bcLoss 會被夾成負值,下面的 `BC -= bcLoss` 反而變成「損失負數」=
	// 白白多贈送 BC,和「入侵造成損失」的敘述矛盾。BC 本來就非正時,沒有更多可虧損,bcLoss=0。
	if s.Player.BC <= 0 {
		bcLoss = 0
	} else if bcLoss > s.Player.BC {
		bcLoss = s.Player.BC
	}
	s.Player.BC -= bcLoss
	if len(s.PlayerColonies) > 0 {
		c := &s.PlayerColonies[0] // 攻擊母星殖民地
		for ; popLoss > 0 && c.Population > 1; popLoss-- {
			c.Population--
			switch {
			case c.Workers > 0:
				c.Workers--
			case c.Farmers > 0:
				c.Farmers--
			case c.Scientists > 0:
				c.Scientists--
			}
		}
	}
	tag := ""
	if defended {
		tag = "(母星艦隊部分擊退)"
	}
	s.LastAntares = fmt.Sprintf("⚠ 安塔蘭人第 %d 次入侵%s:損失 %d BC + 母星人口", sev, tag, bcLoss)
}

// advanceEvents 每回合以固定機率觸發一個 MOO2 風格隨機事件並套用效果,結果記於 LastEvent
// (供回合摘要顯示)。單次效果本身有界(單一事件不會虧超過當下 BC、殖民地人口不低於 1)——但
// 這不保證 BC 全局不為負:忠實 yield 經濟(見 docs/tech/colony-economy-maintenance.md)緩衝
// 很薄,國庫仍可能因連續多回合建築維護費 > 收入而變負,事件本身不是這種情況的成因,只是不會
// 加重「已經非正的 BC」繼續被誤夾成負值虧損(見 case 1 太空海盜的夾值處理)。事件亂數由
// EventSeed 決定,可重現。事件與效果為 remake 設計(對齊 MOO2 事件定性:繁榮/瘟疫/海盜/礦脈/
// 突破/隕石)。
// SendFleet 派遣玩家艦隊前往 dest 星:依兩星歐氏距離換算航行回合數(ETA),每回合 EndTurn
// 遞減。dest 無效、與現址相同、或艦隊正航行中則忽略。回傳是否成功下令。
func (s *GameSession) SendFleet(dest int) bool {
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
	// 1 回合有手冊錨點:p.181 講隨機事件版的蟲洞時寫的是
	// 「moves that fleet to their destination in a single turn」,兩者共用這個語意。
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
		f.AtStar = f.DestStar
		f.DestStar = -1
		if f.AtStar < 0 || f.AtStar >= len(s.Stars) {
			continue
		}
		s.Stars[f.AtStar].Explored = true
		// 抵達星系當下結算一次性發現(原版 Do_System_Discoveries_At_Star_,見 discovery.go)。
		if d := s.discoverSystemSpecials(f.AtStar); d != nil {
			s.LastDiscovery = d
		}
	}
	s.mergeColocatedFleets()
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
// 查表加總——只算建築,艦艇/間諜/軍官維護費本專案尚無可推導的模型(未追蹤運輸艦數量),
// 不計入,見 newHomeworldPlayerState 同款 TODO 註記。s.ColonyBuildings 為 nil 或某殖民地
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
// 但被時空異象凍結的殖民地換成一份「人口歸零」的副本:人口 0 → 沒有產出、沒有食物消耗、
// 沒有人口成長、沒有建造進度,正好對上手冊 p.181 的
// 「unable to produce, grow, or move population … do not need food or cost maintenance either」。
//
// 為什麼用「換成零人口副本」而不是「從清單裡拿掉」:拿掉會破壞索引對齊,而那個對齊是
// 好幾個呼叫端共用的不變量。零人口副本讓凍結成為一個純資料上的效果,不動任何控制流。
func (s *GameSession) coloniesForTurn() []engine.ColonyState {
	frozen := false
	for i := range s.PlayerColonies {
		if s.ColonyInStasis(i) {
			frozen = true
			break
		}
	}
	if !frozen {
		return s.PlayerColonies // 常態:沒有凍結的殖民地,直接用原本的切片,零額外成本
	}
	out := make([]engine.ColonyState, len(s.PlayerColonies))
	copy(out, s.PlayerColonies)
	for i := range out {
		if !s.ColonyInStasis(i) {
			continue
		}
		out[i].Population = 0
		out[i].Farmers, out[i].Workers, out[i].Scientists = 0, 0, 0
		out[i].FlatFood, out[i].FlatIndustry, out[i].FlatResearch = 0, 0, 0
		out[i].SpecialIncome = 0
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
// 沒有這個機制,零緩衝的忠實 yield 經濟一旦被隨機事件/安塔蘭入侵把僅有的農夫扣到 0,
// engine/colony.go 的 colonyGrowth 會在饑荒時永久不套用成長公式,且本專案在此之前完全沒有
// 任何自動改派農夫的機制(ShiftColonyJob 只由玩家 UI 操作觸發)——殖民地會卡死在
// NetIndustry=0、TaxRevenue=0,而建築維護費仍每回合照扣,BC 保證單調流血至負值。這是
// docs/tech/colony-economy-maintenance.md §2.2 實測記錄的死結,本函式是解法之一(任務指示的
// 選項②):非饑荒(Farmers>0 或未 Starving)不動作,一次只搶救 1 人(避免一次饑荒就把整個
// 職務分配打亂),不改動 AI 殖民地(AI 目前的 Farmers/Workers 由 ApplyAIEconomy 每回合依
// decider 重新決定,不會卡在饑荒鎖死)。
// 2026-08-06 修正觸發條件(反組譯佐證):原本只在 Farmers==0 才搶救,但實測有更常見的
// 死結——殖民地農夫還有好幾個卻仍然缺糧(例:瘟疫事件的 losePop 從「人數最多的職務」扣人,
// 當農夫最多時就扣農夫,食物翻負後永遠回不來,30 回合探針 pop 10→…→6、食物卡在 -1)。
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
	s.Player.Maintenance = s.totalBuildingMaintenance()         // 依本回合結算前的實際已建建築重算(取代平坦常數)
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
	// Hyper-Advanced Lv1 研究成本。
	s.Player.HyperAdvancedResearchCost = gamedata.HyperAdvancedCost(s.RuleProfile)
	s.syncTradeGoodsFlag() // 依建造選單同步「貿易品」旗標,供 RunEmpireTurn 判斷是否換算收入
}

// EndTurn 推進一回合:先結算玩家帝國,再讓各 AI 對手自行決策並結算,回合數 +1。
func (s *GameSession) EndTurn() {
	s.prepPlayerDerived()
	s.LastPlayerOutput = engine.RunEmpireTurn(s.Player, s.coloniesForTurn())
	s.Player = s.LastPlayerOutput.Player
	s.recoverFromFamine() // 饑荒防死鎖:見函式註解;依本回合 Starving 結果修正下回合職務分配
	for i := range s.AIPlayers {
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
		ps, colonies := engine.ApplyAIEconomy(s.AIPlayers[i].Player, s.AIPlayers[i].Colonies, s.AIPlayers[i].Decider)
		s.AIPlayers[i].Colonies = colonies
		out := engine.RunEmpireTurn(ps, colonies)
		s.AIPlayers[i].Player = out.Player
		s.advanceAI(i, out) // AI 主動行為:造艦 / 擴張 / 外交態勢
	}
	// 間諜結算須排在玩家與所有 AI 本回合研究都跑完之後(用最新的 CompletedTopics/ChosenTech
	// 判定「對方已知、我方未知」的可偷科技清單),故緊接在上面的 AI 迴圈之後。
	s.advanceEspionage()  // 玩家 ↔ AI 間諜行動(最小迴圈:偷科技 STEAL,見 spy.go)
	s.advanceBuilds()     // 以本回合淨工業推進各殖民地建造
	s.advanceResearch()   // 目前研究主題完成則自動推進到下一個未完成的元件解鎖主題
	s.LastDiscovery = nil // 每回合先清掉上一回合的發現(與 advanceEvents 清 LastEvent 同一個節奏)
	s.advanceFleet()      // 推進艦隊星間航行(ETA 遞減,抵達則標記探索 + 結算一次性發現)
	s.advanceMarines()    // 各 Marine Barracks 殖民地依手冊公式補充陸戰隊駐軍(有上限)
	s.advanceArmor()      // 各 Armor Barracks 殖民地依手冊公式補充戰車營駐軍(有上限,見 ground_invasion.go)
	s.advancePopulation() // 累積各殖民地成長,達門檻則 +1 人口(回寫 Population)
	s.advanceEvents()     // 觸發 MOO2 風格隨機事件(繁榮/瘟疫/海盜…),記於 LastEvent
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
	s.Turn++
	s.advanceShipRepair()      // 停在自家據點的艦艇完全修復(原版 Repair_Ships_At_Colonies_)
	s.advanceAntares()         // 安塔蘭人週期性入侵(依 Turn 排程升級),記於 LastAntares
	s.advanceAIRaids()         // AI 對手突襲玩家殖民地(戰爭態勢 + 軍力領先才發動),記於 LastRaid
	s.advanceConquestVictory() // 對手是否已全滅(手冊三條勝利路徑之一:殲滅所有對手)
	s.advancePlayerDefeat()    // 玩家是否已無任何殖民地(超新星等事件可致,見該函式)
	s.advanceAntaranVictory()  // 是否已攻陷安塔蘭母星(手冊三條勝利路徑之二,見 antaran_victory.go)
	s.advanceAIDiplomacy()     // AI 對手彼此外交關係漂移(多帝國活星系;支撐議會第三方搖擺)
	s.advanceCouncil()         // 銀河議會選舉(手冊三條勝利路徑之一:2/3 多數當選銀河領袖),記於 LastCouncil
	s.advanceMercOffers()      // 傭兵領袖不定期上門(手冊 p.134),補進 MercPool 供玩家在軍官畫面雇用
	s.recordHistory()          // 全帝國國力快照(原版 module 122 Record_History_),供 INFO 歷史圖表
	// 熱座:其餘真人席位的帝國也要各自過完這一回合,否則他們的殖民地會被凍結
	// (見 hotseat.go advanceIdleSeats,含各席位不對稱之處的說明)。
	// 單人局 HotseatEnabled() 為 false,這裡是 no-op,既有行為不變。
	if s.HotseatEnabled() {
		s.advanceIdleSeats()
	}
}

// popGrowthThreshold 是「成長累加值 → +1 人口單位」的門檻。MOO2 手冊(MANUAL_150.html p111
// Growth Formula)給出每回合成長率 a=trunc[(2000·POPRACE·(POPMAX-POPAGG)/POPMAX)^0.5](典型
// ~90),並說明顯示人口為「累積成長率 + 功能人口單位」,但**未給累加→整格人口的明確門檻**;
// 存檔 pop_growth 欄位在單一種族殖民地此存檔為 0/~86,未能乾淨反推。故此門檻為 remake 調校值
// (取 300,使健康殖民地約每 3-4 回合 +1 人口),非 MOO2 精確值。詳見 docs/tech/component-values.md 同款 provenance 註記。
const popGrowthThreshold = 300

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
		grow := s.LastPlayerOutput.Colonies[i].PopGrowth
		grow += grow * s.raceGrowthPct / 100 // 種族成長加成(薩克拉+30、矽基-20…)
		s.popAccum[i] += grow
		for s.popAccum[i] >= popGrowthThreshold && s.PlayerColonies[i].Population < s.PlayerColonies[i].PopMax {
			s.popAccum[i] -= popGrowthThreshold
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
	if i < 0 || i >= len(s.PlayerColonies) {
		return
	}
	cand := s.PlayerColonies[i]
	cand.Workers++ // 原版預設:新人口進生產線
	if engine.RunColonyTurn(cand).FoodSurplus >= 0 {
		s.PlayerColonies[i] = cand
		return
	}
	cand = s.PlayerColonies[i]
	cand.Farmers++ // 會缺糧 → 改配農夫(Assign_Additional_Unblockaded_Farmers_)
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
//  1. 造艦:把部分淨工業投入軍力(好戰性格投更多),FleetStrength 累積。
//  2. 擴張:每隔數回合佔領一顆無主星(Owner=2,OwnedStars++)。
//  3. 外交態勢:依「AI 軍力 vs 玩家軍力 + 難度」漂移對玩家關係分數,經 ai.DecideStance
//     推得態勢(戰爭/敵視/中立/提議貿易/提議結盟),存中文 StanceName。
func (s *GameSession) advanceAI(i int, out engine.EmpireOutput) {
	a := &s.AIPlayers[i]
	prof := aiProfile(*a)

	// 1) 造艦:好戰(工業權重高)投資比例較高。
	//
	// FleetInvestPool 是餘數池,修正既有整數捨去 bug:直接算 TotalNetIndustry/invest 時,
	// 若 TotalNetIndustry(如忠實 yield 下常見的 3)小於 invest(Scientific 性格為 4),
	// 整數除法每回合都捨去成 0,FleetStrength 永久停滯(見 playerHomeworldColony 上方歷史記錄註解/
	// docs/tech/colony-economy-maintenance.md)。改成先把 NetIndustry 存進池子、池子夠 invest
	// 才兌現軍力、餘數留到下回合累積,小額淨工業也能跨回合逐步兌現,不會卡死。
	invest := 4 // 分母越小投資越多
	if prof.IndustryWeight > prof.ResearchWeight {
		invest = 2
	}
	if out.TotalNetIndustry > 0 {
		a.FleetInvestPool += out.TotalNetIndustry
		a.FleetStrength += a.FleetInvestPool / invest
		a.FleetInvestPool %= invest
	}

	// 2) 擴張:每 5 回合佔一顆最靠近既有版圖的無主星。
	if s.Turn%5 == 0 {
		s.aiExpand(i)
	}

	// 2.5) 間諜:AI 用最簡單的週期政策每 6 回合訓練 1 名間諜派來偷玩家科技(見 spy.go
	// advanceEspionage),上限比照手冊每對手 63 人(gamedata.SpySlotBonus 的夾範圍)。不像
	// 玩家 TrainSpy 需要花 BC——AI 訓練成本/BC 限制目前無資料可推導,誠實簡化為免費週期政策
	// (TODO:待有更細緻 AI 經濟模型後補上維護費/訓練成本)。
	if s.Turn%6 == 0 && a.Spies < 63 {
		a.Spies++
	}

	// 3) 外交態勢:AI 越強、難度越高,對玩家越敵對。
	diff := 1.0
	if s.Difficulty >= 0 && s.Difficulty < len(Difficulties) {
		diff = Difficulties[s.Difficulty].Mult
	}
	pm := s.playerMilitary()
	strengthGap := a.FleetStrength - pm // AI 領先越多越敢敵對

	// 關係往一個**平衡點**漂移,而不是每回合無止境累減。
	//
	// 先前是 `Relation -= strengthGap/20*diff`,純累加:只要 AI 軍力領先,關係必定一路
	// 觸底 -40,60 回合後三個 AI 一律「宣戰」,性格完全沒有體感(2026-08-06 探針實測)。
	//
	// 平衡點 = 性格的基礎關係傾向(原版 _personality_relation_modifiers,±70 尺度縮到
	// ±35 對齊 Relation 的 ±40)再被軍力差往敵對方向推。和平主義(+30)因此會停在友好側,
	// 除非玩家軍力遠遜於它;冷酷無情(-20)本來就偏敵對,軍力一領先就迅速宣戰。
	//
	// ⚠ 原版關係演化的確切公式尚未反編(module 27 的 Diplomacy_Growth_/Change_Relations_),
	// 「往平衡點漂移」是 remake 的模型;有硬證的只有平衡點取自哪張表。
	equilibrium := ai.PersonalityRelationModifier(a.Personality) / 2
	target := equilibrium - int(float64(strengthGap)/5*diff)
	switch {
	case a.Relation < target:
		a.Relation++
	case a.Relation > target:
		a.Relation--
	}
	if a.Relation > 40 {
		a.Relation = 40
	}
	if a.Relation < -40 {
		a.Relation = -40
	}
	prevStance := a.StanceName
	stance := ai.DecideStance(diplomacy.RelationLevelForScore(a.Relation), prof)
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

// advanceAIDiplomacy 每回合漂移 AI 對手彼此的外交關係:軍力領先者令鄰邦戒懼(關係下滑),尺度
// /夾值同 AIOpponent.Relation。使多帝國彼此形成敵友,讓星系「活起來」並支撐議會第三方搖擺票。
func (s *GameSession) advanceAIDiplomacy() {
	s.ensureAIRelations()
	n := len(s.AIPlayers)
	if n < 2 {
		return
	}
	diff := 1.0
	if s.Difficulty >= 0 && s.Difficulty < len(Difficulties) {
		diff = Difficulties[s.Difficulty].Mult
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			gap := s.AIPlayers[j].FleetStrength - s.AIPlayers[i].FleetStrength // j 領先 → i 戒懼 j
			r := s.AIRelations[i][j] - int(float64(gap)/20*diff)
			if r > 40 {
				r = 40
			}
			if r < -40 {
				r = -40
			}
			s.AIRelations[i][j] = r
		}
	}
}

// AIRelationName 回傳 AI i 對 AI j 的關係中文分級(供 races 畫面顯示);越界/無矩陣回「中立」。
// 依分數分五級,對齊 Chinese UI 語氣(RelationLevel.Name() 是英文,不直接用於顯示)。
func (s *GameSession) AIRelationName(i, j int) string {
	if i < 0 || j < 0 || i >= len(s.AIRelations) || j >= len(s.AIRelations[i]) {
		return "中立"
	}
	switch r := s.AIRelations[i][j]; {
	case r <= -25:
		return "敵對"
	case r <= -8:
		return "緊張"
	case r < 8:
		return "中立"
	case r < 25:
		return "友好"
	default:
		return "同盟"
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
// AIOpponent 的殖民地平行陣列,長度須恆等——不像玩家還有 Builds/PlayerColonyMarines/
// MarineBarracksAge/PlayerColonyTanks/ArmorBarracksAge/popAccum 那套逐殖民地建造/駐軍追蹤,
// 因為 EndTurn 對 AI 只呼叫 RunEmpireTurn 結算經濟,從不呼叫 advanceBuilds/advanceMarines/
// advanceArmor/advancePopulation 這些玩家專屬的逐殖民地流程,故無需同步那些陣列)。新殖民地的
// ColonyBuildings 項 append 空 map(手冊只保證母星有星基,新拓殖星沒有)。
//
// 2026-07-11 訂正:先前只設旗標、不建殖民地模型(見 AIOpponent.ColonyStars 欄位註解),
// 導致 AI 版圖擴張後經濟(EndTurn 的 RunEmpireTurn(ps, a.Colonies))永遠停在初始母星產出,
// 不會隨佔領星數成長。現在下回合 EndTurn 就會把新殖民地的淨工業算進
// out.TotalNetIndustry,advanceAI 的造艦投資(見上方)自然吃到更多產出,AI 才會隨擴張變強。
//
// gov 傳 gamedata.MoraleGovDictatorship(AIOpponent 沒有 Government 欄位,政府型態未建模,
// 見 newColonyFromStar 註解);種族加成傳 0(AI 無種族加成模型可查)。若該星行星資料不可殖民
// (climateColonizable 為 false——目前星系生成從不產生氣態巨星/小行星帶,見 colonization.go
// 檔頭,故實務上不會發生),保守 continue 找下一顆無主星,不 fallback 成只設旗標(避免旗標與
// 殖民地模型再度分裂)。找不到任何可擴張的無主星則整個 no-op。
func (s *GameSession) aiExpand(i int) {
	// 擴張積極度依性格(原版 _personality_expansion_chance:冷酷 100、和平主義只有 30)。
	// 先前所有 AI 一律每回合都嘗試擴張,擴張速度毫無性格差異。
	if chance := ai.PersonalityExpansionChance(s.AIPlayers[i].Personality); chance < 100 {
		if s.eventRand == nil {
			s.eventRand = rand.New(rand.NewSource(s.EventSeed*2654435761 + 1))
		}
		if s.eventRand.Intn(100) >= chance {
			return
		}
	}
	// 依原版行星估值挑最好的目標,而不是「掃到第一顆能殖民的就佔」。
	// 原版每回合替每個玩家把全星圖算一輪分數(Compute_Base_Planet_Values_),AI 據此挑目標;
	// remake 這裡即時算候選星的分數取最大值,結果等價而不必維護一張快取表。
	best, bestVal := -1, 0
	for idx := range s.Stars {
		if s.Stars[idx].Owner != 0 {
			continue
		}
		if v := s.aiPlanetValue(i, idx); v > bestVal {
			best, bestVal = idx, v
		}
	}
	// 全部候選都是 0 分(不宜居/無行星)時退回原本的順序掃描,確保不會因為估值而完全停止擴張。
	order := make([]int, 0, len(s.Stars))
	if best >= 0 {
		order = append(order, best)
	} else {
		for idx := range s.Stars {
			order = append(order, idx)
		}
	}
	for _, idx := range order {
		if s.Stars[idx].Owner != 0 {
			continue
		}
		if s.StarGuardedByMonster(idx) {
			continue // 怪獸盤據的星系 AI 也進不去(手冊 p.62 的清場條件對所有帝國一體適用)
		}
		colony, ok, _ := s.newColonyFromStar(idx, gamedata.MoraleGovDictatorship, 0, 0, 0)
		if !ok {
			continue
		}
		s.Stars[idx].Owner = 2
		s.AIPlayers[i].OwnedStars++
		s.AIPlayers[i].Colonies = append(s.AIPlayers[i].Colonies, colony)
		s.AIPlayers[i].ColonyStars = append(s.AIPlayers[i].ColonyStars, idx)
		// ColonyBuildings 同步 append 空 map,維持三個平行陣列等長(見 AIOpponent.ColonyBuildings
		// 欄位註解)——手冊只保證母星有星基,新拓殖星沒有,故新 AI 殖民地開局無建築可扣。
		s.AIPlayers[i].ColonyBuildings = append(s.AIPlayers[i].ColonyBuildings, map[string]bool{})
		s.consumeSpecialOnColonize(idx) // 原住民被 AI 併入人口後同樣從行星上消失(見 colonization.go)
		return
	}
}

// aiPlanetValue 算某顆星對 AI i 的價值(原版 Uncolonized_Planet_Worth_To_Player_)。
// 星索引越界、無行星資料或該行星不可殖民時回 0。
//
// AI 的目標傾向由性格決定:冷酷/排外偏礦產(工業=軍力),和平主義偏人口。
// 原版的目標傾向是獨立於性格的另一個維度(玩家結構偏移 2208),remake 沒有那一層,
// 用性格代打——這是 remake 的映射,不是原版對照。
func (s *GameSession) aiPlanetValue(aiIdx, starIdx int) int {
	if starIdx < 0 || starIdx >= len(s.Planets) {
		return 0
	}
	p, _ := s.PlanetDataAt(starIdx)
	if p.NoPlanet || p.Gen < planetGenVersion {
		return 0
	}
	if s.StarGuardedByMonster(starIdx) {
		return 0 // 打不下來的星不該排進拓殖目標
	}
	// 氣態巨星/小行星帶不能殖民(原版 AIPlanetValue 開頭也是 `type != HABITABLE → 0`),
	// AI 不該把它們排進拓殖目標。
	if p.TypeID != gamedata.HABITABLE {
		return 0
	}
	if !climateColonizable(p.ClimateID) {
		return 0
	}
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
	base := gamedata.AIPlanetValue(gamedata.AIPlanetValueInput{
		Habitable: true,
		MaxPop:    gamedata.PlanetBasePopMax(p.SizeID, p.ClimateID),
		Minerals:  p.MineralID,
		Climate:   p.ClimateID,
		Gravity:   p.GravityID,
		// FoodBase 對應原版 planet.foodbase;remake 的等價量是該氣候的每農夫食物。
		FoodBase: gamedata.ClimateFoodPerFarmer(p.ClimateID),
		Special:  int(p.SpecialID),
	}, obj)
	if base <= 0 {
		return 0
	}
	// 第二、三層:鄰近價值與協同效應(原版 Proximity_Worth_To_Player_ /
	// Compute_Contextual_Planet_Values_)。原版的「同一恆星系內的其他行星」在 remake
	// 對應成「鄰近的星」,因為 remake 是一星一行星(見 genPlanets 註解)。
	ctx := s.aiNeighborhood(aiIdx, starIdx, obj)
	ctx.Base = base + gamedata.AIProximityValue(ctx.distances)
	ctx.Size = p.SizeID
	ctx.Colonized = s.Stars[starIdx].Owner != 0
	return gamedata.AIContextualPlanetValue(ctx.AIContextualInput)
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
			if c.Tech != gamedata.TOPIC_STARTING_TECH && !seen[c.Tech] {
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
// 尚未完成的主題(全部完成則維持不變)。這讓「研究→解鎖元件→造艦」的迴圈跨回合持續流動,
// 而非卡在單一主題。玩家仍可透過研究選擇畫面(SetResearchTopic)手動改變當前主題。
func (s *GameSession) advanceResearch() {
	if s.Player.CompletedTopics == nil || !s.Player.CompletedTopics[s.Player.ResearchTopic] {
		return // 目前主題尚未完成,繼續累積
	}
	for _, t := range researchQueue() {
		if !s.Player.CompletedTopics[t] {
			s.Player.ResearchTopic = t
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
		MoralePercent:        colonyMoralePercent(gamedata.MoraleGovDictatorship, homeworldBuildings()),
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
// 這裡只是開局第一回合前的初始值。艦艇/間諜/軍官維護費目前無手冊可推導的模型(本專案未追蹤
// 運輸艦數量),暫不計入——TODO:待接上艦隊維護模型後補上,不臆造數字。
func newHomeworldPlayerState(researchTopic gamedata.ResearchTopic) engine.PlayerState {
	return engine.PlayerState{
		// TaxRate 15:2026-07-12 校正。手冊 p.37 工業稅是「臨時要現金才拉」的補充收入(原生 0-50%、
		// 10% 級距、預設偏低),主收入是人頭(每人 1 BC,見 gamedata.BaseIncomePerPopHalfBC)。先前
		// 寫死 40 把工業稅當唯一收入來源,導致低工業母星流血。使用者定 remake 起始預設為 15(介於
		// 原版慣用的 0 與舊值之間;原生級距為 10 的倍數,15 是 remake 起始值,玩家後續可調)。
		BC: 50, TaxRate: 15, Maintenance: gamedata.BuiltMaintenanceBC(homeworldBuildings()), ResearchTopic: researchTopic,
		// CommandPointsSupply 這裡刻意只填母星星基(homeworldBuildings 的"星基":true)貢獻的
		// 1 點建築供給,不含 gamedata.CommandPointsBase(帝國基礎值 5)——這只是「第一次 EndTurn
		// 前」的暫時值,玩家這欄會在 GameSession.EndTurn 用 totalCommandPointsSupply()(基礎值+
		// 建築供給)整個重算覆蓋掉(見該函式)。AI 這欄則從未被重算(AI 沒有逐殖民地/逐艦追蹤
		// 機制),但 AI 的 UsedCommandPoints 也刻意留 0(見下方),uncovered 永遠夾在 0,是否含
		// 基礎值不影響 AI 是否超支,故這裡不必為 AI 補上——避免誤導未來讀者以為 AI 有真的算過。
		// UsedCommandPoints 這裡刻意不填(留 0):本函式同時供玩家與 AI 共用,AI 沒有逐艦清單
		// (見 UsedCommandPoints 欄位註解),玩家的初始值改由 NewDemoSession 在 Ships 欄位就位後
		// 另外設定,避免在此對 AI 也套用玩家專屬的開局艦隊假設。
		CommandPointsSupply: gamedata.CommandPointsFromBuildings(homeworldBuildings()),
		CompletedTopics: map[gamedata.ResearchTopic]bool{
			gamedata.TOPIC_STARTING_TECH: true,
			gamedata.TOPIC_ENGINEERING:   true,
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
//   - Capitol 刻意不列入此 map:Capitol 不佔用建築格位、不計入 StartingBuildingCount 上限
//     (§3.2),且非玩家可建/可失去的一般建築,本專案的 ColonyBuildings 追蹤機制不收錄它,
//     視為首都固有(隱性)狀態。
//
// 建築數 2 遠低於 StartingBuildingCount(8, BuildingCapAverage)=5 的上限——這是符合手冊的
// (上限只是「至多」,實際只有這兩項的科技條件成立,見 §3.3)。
func homeworldBuildings() map[string]bool {
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
// commandoTier 欄位(2026-07-11 新增,#5 守方 Commando 加成):依種族性格指派開局固定的
// Commando 指揮官技能階(0 無/1 一般/2 進階),見 AIOpponent.Leaders 欄位註解的誠實近似說明——
// 布拉西人(手冊「體格強悍,地面戰加成」)給 Tier2、姆瑞森人(好戰善攻)給 Tier1、席隆人
// (重研究)給 0(無指揮官)。
//
// 2026-08-06 起 profile 不再手動指派:raceEn 對到原版 AIRACES.CFG 的種族性格分布
// (ai.ClassicRacePersonality),開局抽出性格後由 ai.ProfileForPersonality 推導經濟傾向。
// 那張分布表 remake 早就有,但一直是死碼——三個 AI 的行為差異全靠這裡手寫的 profile。
var demoAIOpponentSetup = []struct {
	name         string
	raceEn       string // AIRACES.CFG 的種族名(ai.ClassicRacePersonality 的 key)
	commandoTier int
}{
	{"AI (席隆人)", "Psilons", 0},
	{"AI (姆瑞森人)", "Mrrshan", 1},
	{"AI (布拉西人)", "Bulrathi", 2},
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

// aiCommandoLeader 依 commandoTier 建構一名指揮官技能領袖(Skill="指揮官"),tier<=0 時回傳
// nil(不指派領袖,對應「該 AI 無 Commando 守將」)。Name/Level/Ship 為示範性零值填法,比照
// demoLeaders 的既有欄位語意(非 HERODATA 真實英雄資料);消費端(commandoLeaderTier)只讀
// Skill/Tier,其餘欄位不影響守方加成計算。
func aiCommandoLeader(name string, tier int) []Leader {
	if tier <= 0 {
		return nil
	}
	return []Leader{{Name: name, Skill: "指揮官", Level: tier * 2, Ship: false, Tier: tier}}
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
func buildDemoAIOpponents(aiHomeStars []int, difficulty int, seed int64) []AIOpponent {
	aiPlayers := make([]AIOpponent, 0, len(aiHomeStars))
	// 性格抽樣用獨立的亂數流:同一個 seed 一定抽出同一組性格(存讀檔與重跑要可重現)。
	pr := rand.New(rand.NewSource(seed*31 + 17))
	for i := 0; i < len(aiHomeStars); i++ {
		setup := demoAIOpponentSetup[i%len(demoAIOpponentSetup)]
		pers := pickAIPersonality(setup.raceEn, difficulty, pr)
		aiPlayers = append(aiPlayers, AIOpponent{
			Name:        setup.name,
			Player:      newHomeworldPlayerState(1),
			Colonies:    []engine.ColonyState{playerHomeworldColony()}, // AI 同為 Average 起始單一母星,與玩家共用忠實 yield
			ColonyStars: []int{aiHomeStars[i]},                         // 唯一有實際殖民地模型的星(見 AIOpponent.ColonyStars 註解)
			// ColonyBuildings 母星比照玩家,開局已建成 homeworldBuildings()(海軍陸戰隊營+
			// 星基)——每個 AI 各自 cloneBuildings 一份獨立拷貝,不可共享同一個 map 參考(見
			// AIOpponent.ColonyBuildings 欄位註解)。
			ColonyBuildings: []map[string]bool{cloneBuildings(homeworldBuildings())},
			// Leaders 依種族性格指派開局固定 Commando 守將(#5,見 AIOpponent.Leaders 與
			// demoAIOpponentSetup.commandoTier 欄位註解的誠實近似說明)。
			Leaders:     aiCommandoLeader(setup.name, setup.commandoTier),
			Personality: pers,
			// 經濟傾向由性格推導(見 ai.ProfileForPersonality),不再手寫。
			Decider:    ai.NewRemakeDecider(ai.ProfileForPersonality(pers)),
			OwnedStars: 1,
			// 基礎關係傾向依性格起跳(原版 _personality_relation_modifiers):
			// 和平主義 +30、排外 -50……先前所有 AI 一律從 0 開始,性格毫無體感。
			Relation: clampRelation(ai.PersonalityRelationModifier(pers) / 2),
		})
	}
	return aiPlayers
}

func NewDemoSession() *GameSession {
	const galaxyStars = 24
	const numAIOpponents = 3
	galaxy, aiHomeStars := genGalaxy(galaxyStars, 42, numAIOpponents, galaxyAgeSetting) // 程序化星系(24 星,固定種子=可重現;正式版種子隨新遊戲)
	galaxy[0].Explored = true                                                           // 母星初始已探索
	// 蟲洞用獨立亂數流(45),與 SetupNewGame 同構——demo 局也要有,否則截圖廊/測試看不到這一層。
	genWormholes(galaxy, demoHomeStarSet(aiHomeStars), rand.New(rand.NewSource(45)).Intn)
	// 星雲同理:demo 局也要有,否則截圖廊/測試看不到這一層。
	demoNebulae := genNebulae(galaxySizeClass(len(galaxy)), demoHomeStarSet(aiHomeStars),
		galaxy, rand.New(rand.NewSource(46)))

	aiPlayers := buildDemoAIOpponents(aiHomeStars, 1, 42) // demo 固定難度 1 / seed 42(可重現)

	session := &GameSession{
		Turn:              1,
		Player:            newHomeworldPlayerState(gamedata.TOPIC_ADVANCED_CONSTRUCTION),
		PlayerColonies:    []engine.ColonyState{playerHomeworldColony()},
		ColonyBuildings:   []map[string]bool{homeworldBuildings()},
		PlayerColonyStars: []int{0},                       // 母星 = 星 0(見欄位註解)
		Government:        gamedata.MoraleGovDictatorship, // 預設獨裁(自訂種族 0 點基準),見欄位註解的零值陷阱說明
		AIPlayers:         aiPlayers,
		PlayerSpies:       make([]int, len(aiPlayers)), // 玩家對每個 AI 對手的間諜數,平行 AIPlayers,開局皆 0(見欄位/spy.go ensurePlayerSpies 註解)
		Stars:             galaxy,
		Nebulae:           demoNebulae,
		Planets:           genPlanets(galaxy, rand.New(rand.NewSource(43)), rand.New(rand.NewSource(47)), galaxyAgeSetting, demoHomeStarSet(aiHomeStars)),
		// Monsters 在下面 session 建好後補上——genMonsters 會就地修改 Planets(手冊 p.60:
		// 有怪獸的星系一定另有一個特殊物產),不能在同一個複合字面值裡引用尚未建立的欄位。
		// 開局領袖池為空(2026-07-12 手冊考據校正)。手冊 GAME_MANUAL.pdf p.47 + p.134「Mercenary
		// Leaders」:原版開局玩家**完全沒有任何領袖**,傭兵不定期上門、須花雇用費招入 Leader Pool
		// (上限殖民領袖 4 + 艦艇軍官 4)。先前 demoLeaders() 讓玩家開局自帶「馮·諾伊曼 科學家」並
		// 固定 +25 研究套進母星,是機制錯誤(那應是雇用並指派後才生效)。改為 nil = 忠實空池。
		// demoLeaders()/applyLeaderColonyBonuses 保留供未來「傭兵招募流程」實作後 seed 用(TODO)。
		Leaders: nil,
		// 開局一支艦隊,停在母星(星 0),沒有航行任務(見 fleet.go)。
		Fleets: []Fleet{{Ships: homeworldShips(), AtStar: 0, DestStar: -1}},
		// 母星開局預設建造「貿易品」:archive.org 線上原版實測(2026-07-12 oracle)讀到 Sol III
		// 的 BUILDING 欄就是「Trade Goods」,右下 Income +12 BC——remake 先前開局是「不建造」、
		// 收支 +0,是母星開局態沒對齊的一環(見 docs/tech/oracle-comparison-20260712.md)。
		// Cost 0 同「不建造」語意(見 TradeGoodsBuildName 註解:整包工業轉現金,不累積進度)。
		Builds:              []ColonyBuild{{Name: TradeGoodsBuildName, Progress: 0, Cost: 0}},
		SelectedStar:        -1,
		ShowRelocationLines: true, // 原版預設開(`sub_127E1` 初始化寫 1)
		EventSeed:           42,   // 隨機事件種子(可重現;正式新遊戲遞增)
		RuleProfile:         gamedata.Profile15(),
	}
	// 守衛怪獸(見 monster.go)。放在這裡而不是上面的複合字面值裡,因為 genMonsters 會就地
	// 修改 session.Planets(手冊 p.60:有怪獸的星系一定另有一個特殊物產)。
	session.Monsters = genMonsters(galaxy, session.Planets, rand.New(rand.NewSource(44)), demoHomeStarSet(aiHomeStars))
	session.Player.UsedCommandPoints = session.usedCommandPoints() // 依開局艦隊(homeworldShips)算實際需求,顯示與第一次 EndTurn 後一致
	// 領袖技能接線(2026-07-11):把 Ship=false 的殖民地領袖(科學家/貿易家)技能套到母星。
	// 2026-07-12 開局改為空領袖池(見上方 Leaders 註解),故此呼叫目前是 no-op;保留接線,待未來
	// 傭兵招募流程實作後,玩家雇用並指派殖民地領袖時即生效。
	applyLeaderColonyBonuses(session.Leaders, &session.PlayerColonies[0])
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
		{"馮·諾伊曼", "科學家", 2, false, 1}, // 殖民地領袖:固定研究加成
		{"洛克斐勒", "貿易家", 1, false, 1},  // 殖民地領袖:收入%加成
		{"漢尼拔", "指揮官", 2, true, 1},    // 艦艇軍官
		{"圖靈", "工程師", 1, true, 1},     // 艦艇軍官
	}
}

// advanceMercOffers 每 4 回合有一名傭兵上門(手冊 GAME_MANUAL.pdf p.134「odd individuals
// occasionally show up on your imperial doorstep, offering their services」),補進 MercPool
// (同時最多 3 個 offer)。依 MercOfferedIdx 順序釋出候選、不重複;確定性(依 Turn,不用亂數)
// 以維持存檔/探針可重現。候選釋出完即停(first-version 名單有限)。
func (s *GameSession) advanceMercOffers() {
	if s.Turn%4 != 0 || len(s.MercPool) >= 3 {
		return
	}
	cands := s.mercCandidates()
	if s.MercOfferedIdx >= len(cands) {
		return
	}
	s.MercPool = append(s.MercPool, cands[s.MercOfferedIdx])
	s.MercOfferedIdx++
}

// MercHireCost 回傳雇用某傭兵的一次性費用(gamedata.LeaderHireCost,依技能等級遞增)。
func (s *GameSession) MercHireCost(ld Leader) int {
	exp := leaderDisplayLevelToExpLevel(ld.Level)
	return gamedata.LeaderHireCost(5, exp, 0) // skillValue 基準 5,modifier 0(無 Charismatic 折扣建模)
}

// leaderSlotsFull 回傳該類領袖(殖民地 Ship=false / 艦艇 Ship=true)是否已達上限 4
// (手冊 p.134:「up to four Colony Leaders and four Ship Officers」)。
func (s *GameSession) leaderSlotsFull(ship bool) bool {
	n := 0
	for _, l := range s.Leaders {
		if l.Ship == ship {
			n++
		}
	}
	return n >= 4
}

// HireMerc 雇用 MercPool 的第一名傭兵(手冊 p.134:扣一次性雇用費、招入 Leader Pool)。BC 不足或
// 對應領袖類別已滿(各 4 名)則不雇用、傭兵留在池中。回傳是否成功。
func (s *GameSession) HireMerc() bool {
	if len(s.MercPool) == 0 {
		return false
	}
	ld := s.MercPool[0]
	if s.leaderSlotsFull(ld.Ship) {
		return false
	}
	newBC, ok := engine.HireLeader(s.Player.BC, s.MercHireCost(ld))
	if !ok {
		return false // BC 不足
	}
	s.Player.BC = newBC
	s.MercPool = s.MercPool[1:]
	s.Leaders = append(s.Leaders, ld)
	// 只套「新雇這一名」的殖民地加成——applyLeaderColonyBonuses 是 += 累加,不可對全名單重跑
	// (會重複計算既有領袖),故傳單元素 slice(見該函式註解)。
	if len(s.PlayerColonies) > 0 && !ld.Ship {
		applyLeaderColonyBonuses([]Leader{ld}, &s.PlayerColonies[0])
	}
	return true
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
