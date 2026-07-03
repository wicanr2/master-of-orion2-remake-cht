// Package shell 是「可玩遊戲殼」的純邏輯核心:活的對局狀態、輸入命中判定。
// 不 import ebiten(維持可純測);ebiten 的繪製與輸入輪詢在 cmd/moo2。
package shell

import (
	"fmt"

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

// Ship 是一艘艦艇(供艦隊畫面)。
type Ship struct {
	Name  string
	Class string // 艦體等級(護衛艦/驅逐艦/巡洋艦/戰艦…)
}

// demoShips 是示範艦隊(固定;正式版由存檔/建造填)。
func demoShips() []Ship {
	return []Ship{
		{"探索號", "偵察艦"},
		{"復仇號", "護衛艦"},
		{"雷霆號", "驅逐艦"},
		{"守護號", "巡洋艦"},
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

// BattleResult 是一場戰鬥的結果。
type BattleResult struct {
	Enemy                         string
	PlayerStrength, EnemyStrength int
	PlayerWon                     bool
	PlayerLosses, EnemyLosses     int
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
		_ = i
	}
	s.Ships = append(s.Ships[:wi], s.Ships[wi+1:]...)
}

// ResolveBattle 解算與某敵方的一場戰鬥:比較雙方艦隊總戰力,套用損失,回傳結果。
// 敵方戰力隨回合數增強(示範規則;正式版由敵方實際艦隊算)。
func (s *GameSession) ResolveBattle(enemy string) BattleResult {
	ps := 0
	for _, sh := range s.Ships {
		ps += shipStrength(sh.Class)
	}
	es := 8 + s.Turn*3
	res := BattleResult{Enemy: enemy, PlayerStrength: ps, EnemyStrength: es}
	if ps >= es {
		res.PlayerWon = true
		res.EnemyLosses = es
		if ps < es*3/2 && len(s.Ships) > 0 { // 慘勝小損
			s.removeWeakestShip()
			res.PlayerLosses = 1
		}
	} else {
		for len(s.Ships) > 0 && res.PlayerLosses < 2 {
			s.removeWeakestShip()
			res.PlayerLosses++
		}
	}
	s.LastBattle = &res
	return res
}

// shipNamePool 供新造艦命名(依序循環)。
var shipNamePool = []string{"先鋒號", "勝利號", "無畏號", "蒼穹號", "星辰號", "破曉號", "遠征號", "不朽號", "疾風號", "曙光號"}

// ShipCost 造某艦體等級所需國庫 BC(依戰力點)。
func ShipCost(class string) int { return shipStrength(class) * 20 }

// BuildShip 造一艘指定艦體等級的艦:扣國庫 BC,加入艦隊。BC 不足回 false 不造。
func (s *GameSession) BuildShip(class string) bool {
	cost := ShipCost(class)
	if s.Player.BC < cost {
		return false
	}
	s.Player.BC -= cost
	name := shipNamePool[len(s.Ships)%len(shipNamePool)]
	s.Ships = append(s.Ships, Ship{Name: name, Class: class})
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

// buildOptions 是可建造的項目(名稱 + 工業成本)。空字串為「不建造」。
var buildOptions = []ColonyBuild{
	{"", 0, 0}, {"住宅", 0, 30}, {"工廠", 0, 60}, {"研究實驗室", 0, 80}, {"星港", 0, 120},
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
	Planets          []Planet            // 行星列表
	Leaders          []Leader            // 軍官/領袖名單
	Ships            []Ship              // 艦隊
	LastBattle       *BattleResult       // 上一場戰鬥結果(供戰鬥結果畫面)
	SelectedStar     int                 // 星圖選中的星索引(-1=未選)
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
		Stars:        demoStars(),
		Planets:      genPlanets(demoStars()),
		Leaders:      demoLeaders(),
		Ships:        demoShips(),
		Builds:       make([]ColonyBuild, 2),
		SelectedStar: -1,
	}
}
