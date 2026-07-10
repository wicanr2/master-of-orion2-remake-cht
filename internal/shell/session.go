// Package shell 是「可玩遊戲殼」的純邏輯核心:活的對局狀態、輸入命中判定。
// 不 import ebiten(維持可純測);ebiten 的繪製與輸入輪詢在 cmd/moo2。
package shell

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/diplomacy"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// AIOpponent 是一個由 AI 操控的對手帝國。
type AIOpponent struct {
	Name          string
	Player        engine.PlayerState
	Colonies      []engine.ColonyState
	Decider       ai.Decider
	FleetStrength int    // 累積軍力(每回合由淨工業投資,好戰性格投更多)
	Relation      int    // 對玩家的外交關係分數(驅動 17 級 RelationLevel 與態勢)
	StanceName    string // 目前對玩家態勢(中文;由 ai.DecideStance 推得)
	OwnedStars    int    // 已擴張佔領的星數(含母星)
}

// Star 是星系圖上的一顆星(供星圖渲染;正規化座標 0..1)。
type Star struct {
	X, Y     float64 // 0..1 正規化位置
	Spectral int     // 0=藍 1=白 2=黃 3=橙 4=紅 5=棕 6=黑洞
	Size     int     // 0=大 .. 3=小
	Name     string
	Owner    int  // 0=無主 1=玩家 2=AI
	Explored bool // 艦隊是否曾抵達(已探索)
}

// Ship 是一艘艦艇(供艦隊畫面);Weapon/Armor/Shield/Special 為掛載的元件。
type Ship struct {
	Name                           string
	Class                          string // 艦體等級(護衛艦/巡洋艦/戰艦…)
	Weapon, Armor, Shield, Special string // 元件名
	WeaponAttack, BonusHP          int    // 武器攻擊加成、裝甲+護盾 HP 加成
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
		{"重生程序", 150, 0, gamedata.TOPIC_ARTIFICIAL_LIFE, 0}, // 抽象(種族特性),proxy 待重設計
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
	if c.Tech == gamedata.TOPIC_STARTING_TECH {
		return true
	}
	if s.Player.CompletedTopics == nil || !s.Player.CompletedTopics[c.Tech] {
		return false
	}
	// 主題已完成:未映射科技或未明確抉擇 → 主題層級(維持既有行為)。
	if c.UnlockTech == gamedata.TECH_NONE || s.Player.ExplicitChoice == nil || !s.Player.ExplicitChoice[c.Tech] {
		return true
	}
	// 已明確抉擇該主題:僅所選科技對應元件解鎖。
	return s.Player.ChosenTech != nil && s.Player.ChosenTech[c.Tech] == c.UnlockTech
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
		{"拓荒號", "殖民船", "無武裝", "無裝甲", "無護盾", "無", 0, 0},
		{"先驅一號", "偵察艦", "無武裝", "無裝甲", "無護盾", "無", 0, 0},
		{"先驅二號", "偵察艦", "無武裝", "無裝甲", "無護盾", "無", 0, 0},
	}
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
	if len(s.Ships) == 0 {
		return
	}
	wi := 0
	for i, sh := range s.Ships {
		if shipStrength(sh.Class) < shipStrength(s.Ships[wi].Class) {
			wi = i
		}
	}
	s.Ships = append(s.Ships[:wi], s.Ships[wi+1:]...)
}

// applyDamage 對艦隊(以各艦戰力當 HP)造成 dmg 傷害,由最弱艦起逐艘擊沉,回傳擊沉數。
// Difficulties 是難度選項(名稱 + 敵方戰力倍率),對應 NEW GAME 的 DIFFICULTY。
var Difficulties = []struct {
	Name string
	Mult float64
}{
	{"簡單", 0.6}, {"普通", 1.0}, {"困難", 1.5}, {"不可能", 2.2},
}

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
type combatant struct{ hp, atk, def, wmin, wmax, shield, armor int }

// battleVolley 讓每個存活 attacker 對第一個存活 defender 射一發(真戰鬥公式,固定近距 range=2)。
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
		roll := rng.Intn(100) + 1
		net := attackers[i].atk - d.def
		shot := ResolveShot(net, attackers[i].wmin, attackers[i].wmax, 2, d.shield, d.armor, roll, false, false)
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

// ResolveBattle 快速艦隊自動結算(無格子;供非互動戰鬥)。改用 gamedata 真戰鬥公式逐發解算,
// 與格子戰術戰鬥(tacticalScreen)一致:每回合雙方齊射,每發走命中判定→傷害→過盾→過甲。
func (s *GameSession) ResolveBattle(enemy string) BattleResult {
	mkPlayer := func() []combatant {
		var out []combatant
		for _, sh := range s.Ships {
			body := shipStrength(sh.Class)
			atk := body + sh.WeaponAttack
			atk += atk * s.RaceCombatPct / 100 // 種族戰鬥加成(姆瑞森+25、布拉西/阿爾卡里+15…)
			out = append(out, combatant{hp: body * 3, atk: atk, def: body, wmin: atk / 2, wmax: atk,
				shield: shieldReduceByName(sh.Shield), armor: armorHPByName(sh.Armor)})
		}
		return out
	}
	mult := 1.0
	if s.Difficulty >= 0 && s.Difficulty < len(Difficulties) {
		mult = Difficulties[s.Difficulty].Mult
	}
	var ef []combatant
	for _, st := range genEnemyFleet(s.Turn, mult) {
		ef = append(ef, combatant{hp: st * 3, atk: st, def: st, wmin: st / 2, wmax: st, armor: st})
	}
	pf := mkPlayer()

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
	for i := 0; i < res.PlayerLosses; i++ {
		s.removeWeakestShip()
	}
	s.LastBattle = &res
	return res
}

// DiplomacyResponse 依雙方相對實力回應一個外交提議(和平/貿易/威脅)。
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
	for _, sh := range s.Ships {
		pFleet += shipStrength(sh.Class)
	}
	switch action {
	case "peace":
		if pPop >= ePop {
			return enemy + ":你們的實力我們敬佩,和平協議成立。"
		}
		return enemy + ":哼,弱者不配談和。"
	case "trade":
		return enemy + ":貿易協定成立,願雙方繁榮昌盛。"
	case "threat":
		if pFleet >= 10 {
			return enemy + ":……我們會記住這份侮辱。(關係惡化)"
		}
		return enemy + ":就憑你們這點艦隊?可笑!"
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
	Defense         int // 守方防禦(AF+BD),減 netAttack
	WeaponMin       int // 單發最小傷害
	WeaponMax       int // 單發最大傷害
	ShieldReduction int // 護盾每發減傷
	ArmorHP         int // 裝甲 HP(結構外的緩衝,先耗盡才傷結構)
}

// StartCombat 依玩家艦隊 + 難度生成敵方,建立格子戰鬥雙方艦艇(HP=戰力×3、攻擊=戰力);
// 玩家艦置左欄、敵方置右欄,依序排列。
func (s *GameSession) StartCombat(enemy string) (player, enemyShips []CombatShip) {
	// 由艦艇設計推導真戰鬥公式所需數值(remake 近似;精確值需艦體空間格 + 元件佔格 + 軍官技能):
	//   結構 HP = 艦體×3;裝甲 HP = 設計 BonusHP;Beam Attack = 艦體 + 武器攻擊;
	//   防禦 = 艦體(小艦=低戰力=低防,趨勢近原版);單發傷害 min=max/2、max=Attack;
	//   護盾減傷暫 0(艦艇設計尚未把護盾與裝甲分離,見 gameplay-systems-status.md)。
	for i, sh := range s.Ships {
		body := shipStrength(sh.Class)
		atk := body + sh.WeaponAttack
		player = append(player, CombatShip{
			Name: sh.Name, HP: body * 3, MaxHP: body * 3, Attack: atk, Col: 1, Row: i,
			Defense: body, WeaponMin: atk / 2, WeaponMax: atk,
			ShieldReduction: shieldReduceByName(sh.Shield), ArmorHP: armorHPByName(sh.Armor),
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
		})
	}
	return
}

// ApplyCombatOutcome 依格子戰鬥後存活的玩家艦名,更新艦隊(移除陣亡艦)+ 記錄結果供結果畫面。
func (s *GameSession) ApplyCombatOutcome(enemy string, playerStart, enemyStart int, survivors map[string]bool, won bool) {
	kept := s.Ships[:0]
	for _, sh := range s.Ships {
		if survivors[sh.Name] {
			kept = append(kept, sh)
		}
	}
	s.Ships = kept
	s.LastBattle = &BattleResult{
		Enemy: enemy, PlayerStart: playerStart, EnemyStart: enemyStart, PlayerWon: won,
		PlayerLosses: playerStart - len(kept), EnemyLosses: enemyStart,
	}
}

// shipNamePool 供新造艦命名(依序循環)。
var shipNamePool = []string{"先鋒號", "勝利號", "無畏號", "蒼穹號", "星辰號", "破曉號", "遠征號", "不朽號", "疾風號", "曙光號"}

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

// DesignCost 回傳一組元件選擇(艦體 + 武器/裝甲/護盾/特殊)的總生產成本。
func DesignCost(class string, weapon, armor, shield, special int) int {
	return ShipCost(class) + pick(WeaponOptions, weapon).Cost + pick(ArmorOptions, armor).Cost +
		pick(ShieldOptions, shield).Cost + pick(SpecialOptions, special).Cost
}

// BuildShip 造一艘指定艦體 + 全元件(武器/裝甲/護盾/特殊)的艦:扣國庫總成本,加入艦隊。
// BC 不足回 false。武器加攻擊、裝甲+護盾加 HP、特殊「戰鬥電腦」再加攻擊。
func (s *GameSession) BuildShip(class string, weapon, armor, shield, special int) bool {
	w, a, sh, sp := pick(WeaponOptions, weapon), pick(ArmorOptions, armor), pick(ShieldOptions, shield), pick(SpecialOptions, special)
	cost := ShipCost(class) + w.Cost + a.Cost + sh.Cost + sp.Cost
	if s.Player.BC < cost {
		return false
	}
	s.Player.BC -= cost
	name := shipNamePool[len(s.Ships)%len(shipNamePool)]
	atk := w.Value
	if sp.Name == "戰鬥電腦" {
		atk += sp.Value
	}
	s.Ships = append(s.Ships, Ship{Name: name, Class: class, Weapon: w.Name, Armor: a.Name, Shield: sh.Name,
		Special: sp.Name, WeaponAttack: atk, BonusHP: a.Value + sh.Value})
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

// VoteResult 是一屆銀河議會投票結果(票數依人口)。
type VoteResult struct {
	PlayerVotes, EnemyVotes int
	PlayerWon               bool
}

// CouncilVote 解算一屆銀河議會投票:雙方票數 = 各自帝國總人口,較高者當選領袖。
func (s *GameSession) CouncilVote() VoteResult {
	pv := 0
	for _, c := range s.PlayerColonies {
		pv += c.Population
	}
	ev := 0
	for _, a := range s.AIPlayers {
		for _, c := range a.Colonies {
			ev += c.Population
		}
	}
	return VoteResult{PlayerVotes: pv, EnemyVotes: ev, PlayerWon: pv >= ev}
}

// ColonyBuild 是某殖民地目前的建造項目。
type ColonyBuild struct {
	Name     string
	Progress int
	Cost     int
}

// buildOptions 是「不看前置科技」的全部可建項目(名稱 + 生產成本),衍生自
// gamedata.Buildings(手冊全表 40 項:35 建築 + 5 衛星),空字串為「不建造」排第一個。
// 供將來「完整建築圖鑑」類 UI 參考;實際建造選單(有前置科技 gate)請用
// availableBuildOptions,CycleColonyBuild 已改用該函式。
var buildOptions = allBuildOptions()

// allBuildOptions 把 gamedata.Buildings 轉成 ColonyBuild 選項清單(含「不建造」空項於首位)。
func allBuildOptions() []ColonyBuild {
	out := make([]ColonyBuild, 0, len(gamedata.Buildings)+1)
	out = append(out, ColonyBuild{"", 0, 0})
	for _, b := range gamedata.Buildings {
		out = append(out, ColonyBuild{Name: b.NameZH, Progress: 0, Cost: b.ProductionCost})
	}
	return out
}

// availableBuildOptions 回傳「玩家已研究前置科技」才會出現的建造選單(空字串「不建造」恆在)。
func availableBuildOptions(completedTopics map[gamedata.ResearchTopic]bool) []ColonyBuild {
	out := []ColonyBuild{{"", 0, 0}}
	for _, b := range gamedata.AvailableBuildings(completedTopics) {
		out = append(out, ColonyBuild{Name: b.NameZH, Progress: 0, Cost: b.ProductionCost})
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
// 既有 5 棟(不可壞,對齊既有測試/remake 調校值):
// 自動工廠→工業/工人 +2、研究實驗室→研究/科學家 +5、太空港→工業/工人 +1。
// 海軍陸戰隊營/星基屬防禦設施,現階段無直接產出建模(仍記錄為已建)。
//
// 新增建築:手冊有明確「每單位人口 +N 產出」且對應到 engine.ColonyState 既有欄位
// (IndustryPerWorker/ResearchPerScientist/FoodPerFarmer 每工人-單位產出率;或
// PollutionProcessor/AtmosphericRenewer/CoreWasteDump 既有污染布林旗標)者才建模;
// 其餘(殖民地整體「+N 固定值」、「收入 +N%」、「士氣 +N%」等)因 engine.ColonyState
// 目前無對應欄位,暫不建模,只記錄為已建——TODO:待 engine 補上殖民地固定加成/百分比
// 欄位後回填(不在本次任務範圍,詳見 docs/tech/colony-buildings.md)。
func (s *GameSession) applyBuildingEffect(i int, name string) {
	if i < 0 || i >= len(s.PlayerColonies) {
		return
	}
	c := &s.PlayerColonies[i]
	switch name {
	case "自動工廠": // Automated Factories(既有,不可壞)
		c.IndustryPerWorker += 2
	case "研究實驗室": // Research Laboratory(既有,不可壞;手冊另有 +5 固定研究點,engine 無固定欄位,TODO)
		c.ResearchPerScientist += 5
	case "太空港": // Spaceport(既有,不可壞;手冊實為「BC 收入 +50%」,engine 無收入百分比欄位,以工業近似)
		c.IndustryPerWorker += 1
	case "機器人採礦廠": // Robo Mining Plant:每工業人口 +2 產能(手冊另有 +10 固定值,TODO)
		c.IndustryPerWorker += 2
	case "深層核心礦場": // Deep Core Mine:每工人 +3 產能(手冊另有 +15 固定值,TODO)
		c.IndustryPerWorker += 3
	case "污染處理器": // Pollution Processor:對應 engine.ColonyState.PollutionProcessor 既有旗標
		c.PollutionProcessor = true
	case "大氣更新器": // Atmospheric Renewer:對應 engine.ColonyState.AtmosphericRenewer 既有旗標
		c.AtmosphericRenewer = true
	case "核心廢料場": // Core Waste Dumps:完全消除污染,對應 engine.ColonyState.CoreWasteDump 既有旗標
		c.CoreWasteDump = true
	case "行星超級電腦": // Planetary Supercomputer:每科學家 +2 研究點(手冊另有 +10 固定值,TODO)
		c.ResearchPerScientist += 2
	case "銀河網路中心": // Galactic Cybernet:每科學家 +3 研究點(手冊另有 +15 固定值,TODO)
		c.ResearchPerScientist += 3
	case "水耕農場": // Hydroponic Farm:手冊為殖民地整體 +2 固定食物,engine 無固定欄位,以每農夫 +1 近似
		c.FoodPerFarmer += 1
	case "地底農場": // Subterranean Farms:手冊為殖民地整體 +4 固定食物,engine 無固定欄位,以每農夫 +2 近似
		c.FoodPerFarmer += 2
	case "氣候控制器": // Weather Controller:每農業人口食物產出 +2(手冊數值,對應欄位存在,直接建模)
		c.FoodPerFarmer += 2
		// 其餘 25 項(飛彈基地、裝甲營房、戰機基地、地面砲台、再生反應爐、機器人工廠、
		// 食物複製機、太空學院、異族管理中心、行星證券交易所、太空大學、全息模擬艙、
		// 自動實驗室、歡樂穹頂、生態圈、複製中心、行星重力產生器、行星輻射/通量/屏障護盾、
		// 曲速力場干擾器、戰鬥站、星辰要塞、阿提米絲系統網、次元傳送門)手冊效果不對應
		// engine.ColonyState 既有欄位(陸戰隊生成/艦艇駐防/百分比收入/百分比士氣/人口上限/
		// 軌道防禦等),暫不建模——僅由 advanceBuilds 記入 s.ColonyBuildings 為「已建」,
		// 顯示於畫面,不影響數值結算。TODO:待對應遊戲系統(陸戰隊/艦隊駐防/士氣/國庫/
		// 人口上限)建好後回頭補建模。
	}
}

// advanceBuilds 以各殖民地淨工業推進建造;完成則套用建築長期效果、記錄(供回合摘要)並清空。
// 每殖民地每種建築只建/套用一次(ColonyBuildings 去重),重複建造會即時完成但不再疊加效果。
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
			ind = s.LastPlayerOutput.Colonies[i].NetIndustry
		}
		b.Progress += ind
		if b.Progress >= b.Cost {
			if i < len(s.ColonyBuildings) {
				if s.ColonyBuildings[i] == nil {
					s.ColonyBuildings[i] = make(map[string]bool)
				}
				if !s.ColonyBuildings[i][b.Name] {
					s.ColonyBuildings[i][b.Name] = true
					s.applyBuildingEffect(i, b.Name) // 首次完工才套用長期效果
				}
			}
			s.LastBuilt = append(s.LastBuilt, fmt.Sprintf("殖民地 %d 完成建造:%s", i+1, b.Name))
			*b = ColonyBuild{} // 完成清空
		}
	}
}

// Leader 是一名可雇用的軍官/領袖(供軍官列表)。
type Leader struct {
	Name  string
	Skill string // 專長
	Level int    // 等級
	Ship  bool   // true=艦艇軍官,false=殖民地領袖
}

// demoLeaders 是示範領袖名單(固定;正式版由 HERODATA.LBX 真英雄資料填)。
func demoLeaders() []Leader {
	return []Leader{
		{"馮·諾伊曼", "科學家", 5, false},
		{"洛克斐勒", "貿易家", 4, false},
		{"漢尼拔", "指揮官", 6, true},
		{"圖靈", "工程師", 3, true},
	}
}

// Planet 是一顆行星的顯示資料(供行星列表;正式版由存檔/星系生成填真值)。
type Planet struct {
	Name    string // 星名 + 羅馬數字
	Climate string
	Gravity string
	Mineral string
	Size    string
}

// genPlanets 依星系生成每星一顆行星(氣候由光譜、大小由星體衍生;固定規則,不用亂數)。
func genPlanets(stars []Star) []Planet {
	climates := []string{"放射", "貧瘠", "海洋", "沙漠", "凍原", "有毒", "地獄"}
	sizes := []string{"巨大", "大型", "中型", "小型"}
	minerals := []string{"貧瘠", "一般", "豐富", "富饒"}
	gravs := []string{"低", "常態", "高"}
	roman := []string{"I", "II", "III"}
	out := make([]Planet, 0, len(stars))
	for i, s := range stars {
		cl := "地獄"
		if s.Spectral >= 0 && s.Spectral < len(climates) {
			cl = climates[s.Spectral]
		}
		sz := "中型"
		if s.Size >= 0 && s.Size < len(sizes) {
			sz = sizes[s.Size]
		}
		out = append(out, Planet{
			Name:    s.Name + " " + roman[i%len(roman)],
			Climate: cl,
			Gravity: gravs[i%len(gravs)],
			Mineral: minerals[i%len(minerals)],
			Size:    sz,
		})
	}
	return out
}

// starNamePool 是星名池(二十八宿 + 常見星名;程序生成時依序取用)。
var starNamePool = []string{
	"獵戶", "天狼", "南門", "參宿", "畢宿", "織女", "河鼓", "角宿", "心宿", "北落",
	"五車", "軒轅", "太微", "天津", "婁宿", "氐宿", "房宿", "尾宿", "箕宿", "斗宿",
	"牛宿", "女宿", "虛宿", "危宿", "室宿", "壁宿", "奎宿", "胃宿", "昴宿", "觜宿",
	"井宿", "鬼宿", "柳宿", "星宿", "張宿", "翼宿", "軫宿", "亢宿",
}

// Race 是可選種族(名稱 + 起始加成)。加成對齊 MOO2 各族招牌特性(remake 調校值,非自訂點數精算):
// 工業/研究/食物為每單位產出加成、GrowthPct 為人口成長百分點、StartBC 為額外起始國庫、
// CombatPct 為戰鬥戰力百分點。Desc 為特性摘要(供顯示)。
type Race struct {
	Name      string // 中文名
	EnName    string // 英文名(對應 ai/original.go 種族性格)
	IndBonus  int
	ResBonus  int
	FoodBonus int
	GrowthPct int
	StartBC   int
	CombatPct int
	Desc      string
}

// Races 是 MOO2 十三經典種族,各帶招牌起始加成(remake 調校)。索引 0 為人類(預設)。
var Races = []Race{
	{"人類", "Humans", 0, 0, 0, 0, 60, 0, "外交貿易見長,起始國庫充裕"},
	{"席隆", "Psilons", 0, 4, 0, 0, 0, 0, "創造性研究,科學家產出高"},
	{"薩克拉", "Sakkra", 0, 0, 1, 30, 0, 0, "繁殖迅速,人口成長加成"},
	{"克拉肯", "Klackons", 2, 0, 0, 0, 0, 0, "團結勤奮,工業產出高"},
	{"姆瑞森", "Mrrshan", 0, 0, 0, 0, 0, 25, "好戰善攻,艦艇攻擊加成"},
	{"布拉西", "Bulrathi", 0, 0, 0, 0, 0, 15, "體格強悍,地面與戰鬥加成"},
	{"阿爾卡里", "Alkari", 0, 0, 0, 0, 0, 15, "飛行天賦,艦艇迴避加成"},
	{"梅克拉", "Meklars", 1, 1, 0, 0, 0, 0, "半機械,工業與研究兼具"},
	{"達洛克", "Darloks", 0, 0, 0, 0, 30, 0, "潛伏間諜,擅長滲透"},
	{"崔拉里安", "Trilarians", 0, 0, 1, 10, 0, 0, "水棲民族,食物與成長加成"},
	{"埃雷里安", "Elerians", 0, 1, 0, 0, 0, 15, "心靈感應,研究與戰鬥"},
	{"諾蘭姆", "Gnolams", 0, 0, 0, 0, 120, 0, "幸運富商,起始國庫豐厚"},
	{"矽基", "Silicoids", 1, 0, 0, -20, 0, 0, "岩石生命,耐任何環境但成長慢"},
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
	}
	s.Player.BC += r.StartBC
}

// FlagColors 是玩家旗幟顏色選項(原版新遊戲命名畫面選旗色)。RGB 為近似值,供 UI 呈現。
var FlagColors = []struct {
	Name    string
	R, G, B uint8
}{
	{"紅", 210, 60, 50},
	{"黃", 230, 210, 80},
	{"綠", 80, 190, 90},
	{"藍", 70, 130, 220},
	{"白", 220, 220, 230},
	{"紫", 170, 90, 200},
	{"橙", 230, 140, 60},
	{"棕", 150, 110, 70},
}

// Governments 是自訂種族可選的政府型態(順序對應 customrace 政府型態循環選項)。
var Governments = []string{"獨裁", "封建", "統一", "民主"}

// ApplyGovernment 套用政府型態對「本 remake 已建模資源」的效果(手冊 p.20–23 明列百分比):
//   - 封建(1):研究減半。
//   - 統一(2):食物 +50%、產能 +50%。
//   - 民主(3):研究 +50%。
//   - 獨裁(0):基準,無資源乘數。
//
// ⚠ 誠實標註:政府在原版還有士氣、征服同化回合、間諜/防禦加成、造艦成本、首都陷落效應等,
// 這些系統本 remake 尚未建模,故**未模擬**(不自編近似)。詳見 docs/tech/custom-race-picks.md
// 政府效果附錄與缺口說明。gov 索引對應 Governments。
func (s *GameSession) ApplyGovernment(gov int) {
	pct150 := func(v int) int { return (v*3 + 1) / 2 } // ×1.5 四捨五入
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
}

// GalaxySizes 是星系大小選項(名稱 + 星數),對應 NEW GAME 的 GALAXY SIZE。
var GalaxySizes = []struct {
	Name  string
	Stars int
}{
	{"小型", 12}, {"中型", 24}, {"大型", 36}, {"巨型", 48},
}

// RegenGalaxy 依指定星數重生星系(+ 對應行星);供 NEW GAME 依星系大小生成。
func (s *GameSession) RegenGalaxy(n int, seed int64) {
	s.Stars = genGalaxy(n, seed)
	s.Planets = genPlanets(s.Stars)
	s.SelectedStar = -1
}

// genGalaxy 程序化生成星系:以種子亂數在抖動網格上佈星,隨機光譜/大小/星名;
// 第 0 星為玩家母星、約中段一星為 AI 母星。n=星數(對應星系大小)。
func genGalaxy(n int, seed int64) []Star {
	r := rand.New(rand.NewSource(seed))
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	rows := (n + cols - 1) / cols
	stars := make([]Star, 0, n)
	idx := 0
	names := append([]string(nil), starNamePool...)
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
			} else if idx == n/2 {
				owner = 2 // AI 母星
			}
			stars = append(stars, Star{X: x, Y: y, Spectral: r.Intn(7), Size: r.Intn(4), Name: nm, Owner: owner})
			idx++
		}
	}
	return stars
}

// GameSession 是一局進行中的遊戲狀態。玩家操作改變狀態,EndTurn 推進一回合(結算玩家 + 各 AI)。
type GameSession struct {
	Turn             int
	Player           engine.PlayerState
	PlayerColonies   []engine.ColonyState
	AIPlayers        []AIOpponent
	LastPlayerOutput engine.EmpireOutput // 上一回合玩家結算(供畫面顯示)
	Stars            []Star              // 星系圖
	Planets          []Planet            // 行星列表
	Leaders          []Leader            // 軍官/領袖名單
	Ships            []Ship              // 艦隊
	LastBattle       *BattleResult       // 上一場戰鬥結果(供戰鬥結果畫面)
	SelectedStar     int                 // 星圖選中的星索引(-1=未選)
	Difficulty       int                 // 難度索引(shell.Difficulties)
	Builds           []ColonyBuild       // 各殖民地建造項目(對應 PlayerColonies)
	LastBuilt        []string            // 上回合完成的建造(供回合摘要)
	FleetAtStar      int                 // 玩家艦隊所在星索引(初始=母星 0)
	FleetDestStar    int                 // 艦隊目的星索引(-1=無航行任務)
	FleetETA         int                 // 抵達目的星尚需回合數(0=已抵達/靜止)
	popAccum         []int               // 各殖民地人口成長累加值(達門檻則 +1 人口)
	ColonyBuildings  []map[string]bool   // 各殖民地已完工建築(去重,避免重複套用長期效果)
	EventSeed        int64               // 隨機事件亂數種子(可重現;新遊戲遞增)
	LastEvent        string              // 本回合觸發的隨機事件描述(空=無事件;供回合摘要)
	DisableEvents    bool                // 關閉隨機事件(供確定性經濟測試隔離)
	eventRand        *rand.Rand          // 事件亂數源(由 EventSeed 惰性建立)
	AntaresRaids     int                 // 已發生的安塔蘭突襲次數(逐次升級強度)
	LastAntares      string              // 本回合安塔蘭突襲描述(空=無;供回合摘要)
	RaceIndex        int                 // 玩家選定的種族(shell.Races 索引)
	PlayerName       string              // 玩家帝國/領袖名稱(新遊戲命名畫面設定)
	FlagColor        int                 // 玩家旗幟顏色索引(shell.FlagColors)
	RaceCombatPct    int                 // 種族戰鬥戰力百分點加成(供戰鬥使用)
	raceGrowthPct    int                 // 種族人口成長百分點加成(供 advancePopulation)
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
	defended := false
	if s.FleetAtStar == 0 {
		for _, sh := range s.Ships {
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

	if bcLoss > s.Player.BC {
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
// (供回合摘要顯示)。效果皆有界(BC 不為負、殖民地人口不低於 1)。事件亂數由 EventSeed 決定,
// 可重現。事件與效果為 remake 設計(對齊 MOO2 事件定性:繁榮/瘟疫/海盜/礦脈/突破/隕石)。
func (s *GameSession) advanceEvents() {
	s.LastEvent = ""
	if s.DisableEvents {
		return
	}
	if s.eventRand == nil {
		s.eventRand = rand.New(rand.NewSource(s.EventSeed*2654435761 + 1))
	}
	if s.eventRand.Float64() >= 0.30 { // 每回合 30% 機率有事件
		return
	}
	colony := func() *engine.ColonyState {
		if len(s.PlayerColonies) == 0 {
			return nil
		}
		return &s.PlayerColonies[s.eventRand.Intn(len(s.PlayerColonies))]
	}
	losePop := func(c *engine.ColonyState, n int) {
		for ; n > 0 && c.Population > 1; n-- {
			c.Population--
			switch { // 由最多的職務扣人
			case c.Workers >= c.Farmers && c.Workers >= c.Scientists && c.Workers > 0:
				c.Workers--
			case c.Farmers >= c.Scientists && c.Farmers > 0:
				c.Farmers--
			case c.Scientists > 0:
				c.Scientists--
			}
		}
	}
	switch s.eventRand.Intn(6) {
	case 0: // 經濟繁榮
		gain := 50 + s.Turn
		s.Player.BC += gain
		s.LastEvent = fmt.Sprintf("經濟繁榮:國庫獲得 %d BC", gain)
	case 1: // 太空海盜
		loss := 40
		if loss > s.Player.BC {
			loss = s.Player.BC
		}
		s.Player.BC -= loss
		s.LastEvent = fmt.Sprintf("太空海盜劫掠:損失 %d BC", loss)
	case 2: // 富礦脈
		if c := colony(); c != nil {
			c.IndustryPerWorker++
			s.LastEvent = "發現富礦脈:一殖民地工業/工人 +1"
		}
	case 3: // 瘟疫
		if c := colony(); c != nil && c.Population > 1 {
			losePop(c, 2)
			s.LastEvent = "瘟疫爆發:一殖民地人口減少"
		}
	case 4: // 科學突破
		s.Player.ResearchProgress += 150
		s.LastEvent = "科學突破:研究進度 +150 RP"
	case 5: // 隕石撞擊
		if c := colony(); c != nil && c.Population > 1 {
			losePop(c, 1)
			s.LastEvent = "隕石撞擊:一殖民地人口減少"
		}
	}
}

// SendFleet 派遣玩家艦隊前往 dest 星:依兩星歐氏距離換算航行回合數(ETA),每回合 EndTurn
// 遞減。dest 無效、與現址相同、或艦隊正航行中則忽略。回傳是否成功下令。
func (s *GameSession) SendFleet(dest int) bool {
	if dest < 0 || dest >= len(s.Stars) || dest == s.FleetAtStar || s.FleetETA > 0 {
		return false
	}
	a, b := s.Stars[s.FleetAtStar], s.Stars[dest]
	dist := math.Hypot(a.X-b.X, a.Y-b.Y)
	eta := int(math.Ceil(dist * 8)) // 8 = 星系跨度→回合的換算(全跨約 8-11 回合)
	if eta < 1 {
		eta = 1
	}
	s.FleetDestStar = dest
	s.FleetETA = eta
	return true
}

// advanceFleet 推進艦隊航行:ETA 遞減,歸零則抵達(FleetAtStar=目的),並將該星標記為已探索。
func (s *GameSession) advanceFleet() {
	if s.FleetETA <= 0 || s.FleetDestStar < 0 {
		return
	}
	s.FleetETA--
	if s.FleetETA == 0 {
		s.FleetAtStar = s.FleetDestStar
		s.FleetDestStar = -1
		if s.FleetAtStar < len(s.Stars) {
			s.Stars[s.FleetAtStar].Explored = true
		}
	}
}

// EndTurn 推進一回合:先結算玩家帝國,再讓各 AI 對手自行決策並結算,回合數 +1。
func (s *GameSession) EndTurn() {
	s.LastPlayerOutput = engine.RunEmpireTurn(s.Player, s.PlayerColonies)
	s.Player = s.LastPlayerOutput.Player
	for i := range s.AIPlayers {
		out := engine.RunAIEmpireTurn(s.AIPlayers[i].Player, s.AIPlayers[i].Colonies, s.AIPlayers[i].Decider)
		s.AIPlayers[i].Player = out.Player
		s.advanceAI(i, out) // AI 主動行為:造艦 / 擴張 / 外交態勢
	}
	s.advanceBuilds()     // 以本回合淨工業推進各殖民地建造
	s.advanceResearch()   // 目前研究主題完成則自動推進到下一個未完成的元件解鎖主題
	s.advanceFleet()      // 推進艦隊星間航行(ETA 遞減,抵達則標記探索)
	s.advancePopulation() // 累積各殖民地成長,達門檻則 +1 人口(回寫 Population)
	s.advanceEvents()     // 觸發 MOO2 風格隨機事件(繁榮/瘟疫/海盜…),記於 LastEvent
	s.Turn++
	s.advanceAntares() // 安塔蘭人週期性入侵(依 Turn 排程升級),記於 LastAntares
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
			s.PlayerColonies[i].Workers++ // 新人口預設分配為工人
		}
	}
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
	for _, sh := range s.Ships {
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
	invest := 4 // 分母越小投資越多
	if prof.IndustryWeight > prof.ResearchWeight {
		invest = 2
	}
	if out.TotalNetIndustry > 0 {
		a.FleetStrength += out.TotalNetIndustry / invest
	}

	// 2) 擴張:每 5 回合佔一顆最靠近既有版圖的無主星。
	if s.Turn%5 == 0 {
		s.aiExpand(i)
	}

	// 3) 外交態勢:AI 越強、難度越高,對玩家越敵對。
	diff := 1.0
	if s.Difficulty >= 0 && s.Difficulty < len(Difficulties) {
		diff = Difficulties[s.Difficulty].Mult
	}
	pm := s.playerMilitary()
	strengthGap := a.FleetStrength - pm // AI 領先越多越敢敵對
	a.Relation -= int(float64(strengthGap)/20*diff) + 0
	if a.Relation > 40 {
		a.Relation = 40
	}
	if a.Relation < -40 {
		a.Relation = -40
	}
	stance := ai.DecideStance(diplomacy.RelationLevelForScore(a.Relation), prof)
	a.StanceName = stanceNames[stance]
}

// stanceNames 是 ai.Stance 的中文顯示。
var stanceNames = map[ai.Stance]string{
	ai.StanceWar:             "宣戰",
	ai.StanceHostile:         "敵視",
	ai.StanceNeutral:         "中立",
	ai.StanceProposeTrade:    "提議貿易",
	ai.StanceProposeAlliance: "提議結盟",
}

// aiExpand 讓第 i 個 AI 佔領一顆無主星(標 Owner=2),OwnedStars++。找不到無主星則不動作。
func (s *GameSession) aiExpand(i int) {
	for idx := range s.Stars {
		if s.Stars[idx].Owner == 0 {
			s.Stars[idx].Owner = 2
			s.AIPlayers[i].OwnedStars++
			return
		}
	}
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

// averageHomeworldColony 建一個「Average 起始文明等級」的忠實母星殖民地,依
// docs/tech/homeworld-init.md:單一母星(§1)、PlanetSize=Large(母星通常為大型)、
// PopMax=20 對齊 gamedata `pop_max` 表(Large 星球容量,§8 交叉驗證高信心)。
//
// ⚠ Population=8、Farmers/Workers/Scientists 分配:手冊全文搜尋「starting population」
// 零命中(§2.1),此為手冊 §3.5 建築數公式 worked example 用的同一 pop 值(8),沿用作合理
// 預設,人口分配比例（§2.3 手冊亦未給）延續既有 remake 預設,兩者皆待 DOSBox 原版存檔確認。
func averageHomeworldColony() engine.ColonyState {
	return engine.ColonyState{
		Population: 8, PopMax: 20, Farmers: 3, Workers: 4, Scientists: 1,
		FoodPerFarmer: 4, IndustryPerWorker: 6, ResearchPerScientist: 30,
		PlanetSize: gamedata.LARGE_PLANET, MoralePercent: 10,
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
// BC 國庫沿用既有 remake 預設值 100——手冊未給開局 BC 數字(§6.1),待確認。
func newHomeworldPlayerState(researchTopic gamedata.ResearchTopic) engine.PlayerState {
	return engine.PlayerState{
		BC: 100, TaxRate: 40, Maintenance: 5, ResearchTopic: researchTopic,
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

// NewDemoSession 建一個最小可玩對局:玩家 + 1 個科學傾向 AI 對手,雙方各持 Average 起始
// 文明等級的忠實單一母星(docs/tech/homeworld-init.md,取代先前程序生成的 2 假殖民地)。
// 供「最小可玩迴圈」骨架用;正式新遊戲流程(選種族/星系生成/起始文明等級選擇)為後續工作。
func NewDemoSession() *GameSession {
	galaxy := genGalaxy(24, 42) // 程序化星系(24 星,固定種子=可重現;正式版種子隨新遊戲)
	galaxy[0].Explored = true   // 母星初始已探索
	return &GameSession{
		Turn:            1,
		Player:          newHomeworldPlayerState(gamedata.TOPIC_ADVANCED_CONSTRUCTION),
		PlayerColonies:  []engine.ColonyState{averageHomeworldColony()},
		ColonyBuildings: []map[string]bool{homeworldBuildings()},
		AIPlayers: []AIOpponent{{
			Name:     "AI (賽隆人)",
			Player:   newHomeworldPlayerState(1),
			Colonies: []engine.ColonyState{averageHomeworldColony()}, // 對稱:AI 同樣忠實單一母星
			Decider:  ai.NewRemakeDecider(ai.ProfileScientific),
		}},
		Stars:         galaxy,
		Planets:       genPlanets(galaxy),
		Leaders:       demoLeaders(),
		Ships:         homeworldShips(),
		Builds:        make([]ColonyBuild, 1),
		SelectedStar:  -1,
		FleetAtStar:   0,  // 母星
		FleetDestStar: -1, // 無航行任務
		EventSeed:     42, // 隨機事件種子(可重現;正式新遊戲遞增)
	}
}
