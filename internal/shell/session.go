// Package shell 是「可玩遊戲殼」的純邏輯核心:活的對局狀態、輸入命中判定。
// 不 import ebiten(維持可純測);ebiten 的繪製與輸入輪詢在 cmd/moo2。
package shell

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

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

// Ship 是一艘艦艇(供艦隊畫面);Weapon/Armor/Shield/Special 為掛載的元件。
type Ship struct {
	Name                           string
	Class                          string // 艦體等級(護衛艦/巡洋艦/戰艦…)
	Weapon, Armor, Shield, Special string // 元件名
	WeaponAttack, BonusHP          int    // 武器攻擊加成、裝甲+護盾 HP 加成
}

// Component 是一個艦艇元件(名稱 + 成本 + 效果值)。
type Component struct {
	Name  string
	Cost  int
	Value int // 武器=攻擊、裝甲/護盾=HP、特殊=攻擊或視元件而定
}

// 元件清單(對齊 MOO2 早期元件概念:武器/裝甲/護盾/特殊裝備,各含成本與效果)。
var (
	WeaponOptions  = []Component{{"無武裝", 0, 0}, {"雷射", 20, 2}, {"質量投射器", 40, 4}, {"核飛彈", 60, 6}, {"離子砲", 100, 8}}
	ArmorOptions   = []Component{{"無裝甲", 0, 0}, {"鈦裝甲", 30, 10}, {"三鈦裝甲", 60, 25}, {"天龍鱗甲", 120, 50}}
	ShieldOptions  = []Component{{"無護盾", 0, 0}, {"I 級護盾", 40, 15}, {"II 級護盾", 80, 35}, {"III 級護盾", 150, 60}}
	SpecialOptions = []Component{{"無", 0, 0}, {"戰鬥電腦", 80, 3}, {"自動修復", 60, 0}, {"隱形裝置", 100, 0}}
)

// demoShips 是示範艦隊(固定;正式版由存檔/建造填)。
func demoShips() []Ship {
	return []Ship{
		{"探索號", "偵察艦", "無武裝", "無裝甲", "無護盾", "無", 0, 0},
		{"復仇號", "護衛艦", "雷射", "鈦裝甲", "無護盾", "無", 2, 10},
		{"雷霆號", "驅逐艦", "質量投射器", "三鈦裝甲", "I 級護盾", "無", 4, 40},
		{"守護號", "巡洋艦", "核飛彈", "三鈦裝甲", "II 級護盾", "戰鬥電腦", 9, 60},
	}
}

// shipStrength 依艦體等級給戰力點(供最小戰鬥解算;正式版由艦艇設計的武器/裝甲算)。
func shipStrength(class string) int {
	switch class {
	case "偵察艦":
		return 1
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
func applyDamage(fleet *[]int, dmg int) int {
	sort.Ints(*fleet)
	destroyed := 0
	for len(*fleet) > 0 && dmg >= (*fleet)[0] {
		dmg -= (*fleet)[0]
		*fleet = (*fleet)[1:]
		destroyed++
	}
	return destroyed
}

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
func (s *GameSession) ResolveBattle(enemy string) BattleResult {
	pFleet := make([]int, 0, len(s.Ships))
	for _, sh := range s.Ships {
		pFleet = append(pFleet, shipStrength(sh.Class))
	}
	mult := 1.0
	if s.Difficulty >= 0 && s.Difficulty < len(Difficulties) {
		mult = Difficulties[s.Difficulty].Mult
	}
	eFleet := genEnemyFleet(s.Turn, mult)
	res := BattleResult{Enemy: enemy, PlayerStart: len(pFleet), EnemyStart: len(eFleet)}
	for round := 1; round <= 6 && len(pFleet) > 0 && len(eFleet) > 0; round++ {
		pPower, ePower := 0, 0
		for _, v := range pFleet {
			pPower += v
		}
		for _, v := range eFleet {
			ePower += v
		}
		eDestroyed := applyDamage(&eFleet, pPower)
		pDestroyed := applyDamage(&pFleet, ePower)
		res.Log = append(res.Log, fmt.Sprintf("第 %d 回合:擊沉敵艦 %d ／ 我方損失 %d", round, eDestroyed, pDestroyed))
	}
	res.PlayerLosses = res.PlayerStart - len(pFleet)
	res.EnemyLosses = res.EnemyStart - len(eFleet)
	res.PlayerWon = len(eFleet) == 0 || len(pFleet) >= len(eFleet)
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

// CombatShip 是格子戰術戰鬥中的一艘艦(有 HP + 格位)。
type CombatShip struct {
	Name      string
	HP, MaxHP int
	Attack    int
	Col, Row  int // 格位(8 欄 × 6 列)
}

// StartCombat 依玩家艦隊 + 難度生成敵方,建立格子戰鬥雙方艦艇(HP=戰力×3、攻擊=戰力);
// 玩家艦置左欄、敵方置右欄,依序排列。
func (s *GameSession) StartCombat(enemy string) (player, enemyShips []CombatShip) {
	for i, sh := range s.Ships {
		hp := shipStrength(sh.Class)*3 + sh.BonusHP     // 艦體 + 裝甲/護盾 HP
		atk := shipStrength(sh.Class) + sh.WeaponAttack // 艦體 + 武器(含戰鬥電腦)攻擊
		player = append(player, CombatShip{Name: sh.Name, HP: hp, MaxHP: hp, Attack: atk, Col: 1, Row: i})
	}
	mult := 1.0
	if s.Difficulty >= 0 && s.Difficulty < len(Difficulties) {
		mult = Difficulties[s.Difficulty].Mult
	}
	for i, st := range genEnemyFleet(s.Turn, mult) {
		enemyShips = append(enemyShips, CombatShip{Name: fmt.Sprintf("%s艦%d", enemy, i+1), HP: st * 3, MaxHP: st * 3, Attack: st, Col: 6, Row: i})
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

// buildOptions 是可建造的項目(名稱 + 生產成本,對齊 MOO2 建築生產成本:
// 多數基礎建築 60、太空港 100、星基 300)。空字串為「不建造」。
var buildOptions = []ColonyBuild{
	{"", 0, 0}, {"自動工廠", 0, 60}, {"海軍陸戰隊營", 0, 60}, {"研究實驗室", 0, 60}, {"太空港", 0, 100}, {"星基", 0, 300},
}

// CycleColonyBuild 循環切換某殖民地的建造項目(進度歸零)。
func (s *GameSession) CycleColonyBuild(idx int) {
	if idx < 0 || idx >= len(s.Builds) {
		return
	}
	cur := 0
	for i, o := range buildOptions {
		if o.Name == s.Builds[idx].Name {
			cur = i
			break
		}
	}
	next := buildOptions[(cur+1)%len(buildOptions)]
	s.Builds[idx] = ColonyBuild{Name: next.Name, Progress: 0, Cost: next.Cost}
}

// advanceBuilds 以各殖民地淨工業推進建造;完成則清空並記錄(供回合摘要)。
func (s *GameSession) advanceBuilds() {
	s.LastBuilt = nil
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
}

// EndTurn 推進一回合:先結算玩家帝國,再讓各 AI 對手自行決策並結算,回合數 +1。
func (s *GameSession) EndTurn() {
	s.LastPlayerOutput = engine.RunEmpireTurn(s.Player, s.PlayerColonies)
	s.Player = s.LastPlayerOutput.Player
	for i := range s.AIPlayers {
		out := engine.RunAIEmpireTurn(s.AIPlayers[i].Player, s.AIPlayers[i].Colonies, s.AIPlayers[i].Decider)
		s.AIPlayers[i].Player = out.Player
	}
	s.advanceBuilds() // 以本回合淨工業推進各殖民地建造
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
	galaxy := genGalaxy(24, 42) // 程序化星系(24 星,固定種子=可重現;正式版種子隨新遊戲)
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
		Stars:        galaxy,
		Planets:      genPlanets(galaxy),
		Leaders:      demoLeaders(),
		Ships:        demoShips(),
		Builds:       make([]ColonyBuild, 2),
		SelectedStar: -1,
	}
}
