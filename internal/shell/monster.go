package shell

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

var monsterWeaponNames = map[int]string{
	23: "反物質彈", 26: "生物終結者", 40: "巨龍吐息", 41: "相位眼",
	42: "水晶射線", 43: "電漿吐息", 44: "電漿通量", 45: "腐蝕黏液",
}

func monsterTypedMods(raw int) []string {
	var out []string
	if raw&gamedata.MonsterWeaponModHeavyMount != 0 {
		out = append(out, string(gamedata.ModHeavyMount))
	}
	if raw&gamedata.MonsterWeaponModPointDefense != 0 {
		out = append(out, string(gamedata.ModPointDefense))
	}
	if raw&gamedata.MonsterWeaponModOverloadedTorpedo != 0 {
		out = append(out, string(gamedata.ModOverloadedTorpedo))
	}
	return out
}

func monsterCombatMounts(blueprint gamedata.MonsterBlueprint) []ShipWeaponMount {
	out := make([]ShipWeaponMount, 0, len(blueprint.Weapons))
	for _, raw := range blueprint.Weapons {
		_, hi, ok := gamedata.MonsterWeaponDamageRange(raw)
		if !ok {
			continue
		}
		w, _ := gamedata.OrigWeaponByID(raw.WeaponID)
		ammo := raw.Ammo
		if w.Cat == gamedata.WeaponCatMissile || w.Cat == gamedata.WeaponCatTorpedo {
			ammo = w.Ammo
		}
		out = append(out, ShipWeaponMount{RawType: raw.WeaponID, Name: monsterWeaponNames[raw.WeaponID],
			MaxCount: raw.Count, WorkingCount: raw.Count, Arc: gamedata.WeaponArc(raw.Arc),
			RawMods: uint16(raw.Mods), Mods: monsterTypedMods(raw.Mods), Ammo: ammo, Attack: hi})
	}
	return out
}

// StartMonsterCombat 從事件怪物精確藍圖建立正常格子戰術雙方。
func (s *GameSession) StartMonsterCombat(starIdx int) (player, monsters []CombatShip, reason MonsterCombatRefusalCode) {
	m := s.MonsterAtStar(starIdx)
	if m == nil {
		return nil, nil, MonsterCombatNoMonster
	}
	if s.Fleet().AtStar != starIdx || s.Fleet().ETA != 0 {
		return nil, nil, MonsterCombatFleetNotPresent
	}
	blueprint, ok := gamedata.MonsterBlueprintFor(m.Kind)
	if !ok {
		return nil, nil, MonsterCombatNoBlueprint
	}
	player, _ = s.StartCombat("\x00monster-adapter")
	if len(player) == 0 {
		return nil, nil, MonsterCombatNoCombatShip
	}
	count := monsterGroupCount(m)
	structure, armor := m.Structure, m.Armor
	for i := 0; i < count && structure > 0; i++ {
		hp := min(blueprint.Structure, structure)
		arm := min(blueprint.Armor, armor)
		structure, armor = structure-hp, armor-arm
		mounts := monsterCombatMounts(blueprint)
		firstName, firstMin, firstMax := "", 0, 0
		if len(mounts) > 0 {
			firstName, firstMax = mounts[0].Name, mounts[0].Attack
			firstMin, _, _ = gamedata.MonsterWeaponDamageRange(blueprint.Weapons[0])
		}
		name := fmt.Sprintf("%s %d", gamedata.MonsterNameZH(m.Kind), i+1)
		ship := CombatShip{Name: name, HP: hp, MaxHP: blueprint.Structure, ArmorHP: arm,
			Attack: gamedata.ComputerBonus(blueprint.Computer), Defense: blueprint.BaseCombatSpeed * 5,
			WeaponMin: firstMin, WeaponMax: firstMax, WeaponName: firstName,
			Kind: weaponKindByName(firstName), WeaponArc: gamedata.ARC_MONSTER_360, WeaponMounts: mounts,
			WeaponModes: NewTacticalWeaponModes(mounts), WeaponAmmo: func() int {
				if len(mounts) > 0 {
					return mounts[0].Ammo
				}
				return 0
			}(),
			DriveLevel: blueprint.Drive, CombatSpeed: blueprint.BaseCombatSpeed,
			SizeClass: gamedata.CombatShipClass(blueprint.Size), Col: 6, Row: i % TacticalGridRows,
			Facing: 8, Initiative: gamedata.CombatInitiative(gamedata.ComputerBonus(blueprint.Computer), blueprint.BaseCombatSpeed),
			SpriteIdx: blueprint.Picture}
		monsters = append(monsters, ship)
	}
	return player, monsters, ""
}

// ApplyMonsterTacticalOutcome 將格子戰術倖存者回寫玩家艦隊與怪物聚合雙血池。
func (s *GameSession) ApplyMonsterTacticalOutcome(starIdx, playerStart, monsterStart int,
	playerSurvivors map[string]bool, monsterSurvivors []CombatShip, won bool) {
	structure, armor := 0, 0
	for _, ship := range monsterSurvivors {
		if ship.HP > 0 {
			structure += ship.HP
			armor += max(0, ship.ArmorHP)
		}
	}
	s.applyMonsterTacticalOutcome(starIdx, playerStart, monsterStart, playerSurvivors,
		won, structure, armor, len(monsterSurvivors), true)
}

func (s *GameSession) applyMonsterTacticalOutcome(starIdx, playerStart, monsterStart int,
	playerSurvivors map[string]bool, won bool, structure, armor, survivorCount int, record bool) {
	m := s.MonsterAtStar(starIdx)
	if m == nil {
		return
	}
	if record {
		names := make([]string, 0, len(playerSurvivors))
		for name, alive := range playerSurvivors {
			if alive {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		s.recordPlayerCommand(PlayerCommand{Name: CmdMonsterCombatOutcome,
			Args: []int{starIdx, playerStart, monsterStart, boolInt(won), structure, armor, survivorCount},
			Text: strings.Join(names, "\x00")})
	}
	enemySurvivors := map[string]bool{}
	for i := 0; i < survivorCount; i++ {
		enemySurvivors[fmt.Sprintf("monster-%d", i)] = true
	}
	destroyedHullSum := 0
	if blueprint, ok := gamedata.MonsterBlueprintFor(m.Kind); ok {
		destroyedHullSum = max(0, monsterStart-survivorCount) * (blueprint.Size + 1)
	}
	s.applyCombatOutcome(gamedata.MonsterNameZH(m.Kind), playerStart, monsterStart,
		playerSurvivors, enemySurvivors, won, false, destroyedHullSum)
	if structure <= 0 || survivorCount == 0 {
		s.removeMonsterAt(starIdx)
		return
	}
	m = s.MonsterAtStar(starIdx)
	m.Structure, m.Armor = structure, armor
	m.Count = survivorCount
	if m.Kind == gamedata.MonsterEel {
		if len(m.EelAges) > survivorCount {
			m.EelAges = m.EelAges[:survivorCount]
		}
	}
}

// monster.go:守衛星系的太空怪獸。
//
// 這個系統其實一直被程式碼「引用」著卻不存在:colonization.go 檔頭抄的手冊原文就寫著殖民船
// 要「as long as all space monsters and enemy ships have been cleared from that planet's
// system」——但那個 gate 從來沒有東西可擋。現在有了。
//
// 種類與數值見 gamedata/space_monster.go(名字來自執行檔字串表、傷害來自手冊 p.114)。
// 這一層負責:①星圖生成時擺放 ②擋住拓殖/前哨站 ③艦隊抵達時的戰鬥。
//
// 手冊 p.30:「For tracking and combat purposes, they're treated as fleets.」——所以打怪獸
// 走的是既有的艦隊戰鬥解算,不另做一套。
//
// 手冊 p.60:「a system with a monster will always have another special — that's usually
// what drew the monster there in the first place.」——擺放怪獸時**強制該星另有一個特殊物產**,
// 這是手冊給的生成規則,不是 remake 的設計。

// MonsterGuard 是一顆星上的守衛怪獸。
type MonsterGuard struct {
	StarIndex int
	Kind      gamedata.SpaceMonster
	Structure int // 目前剩餘結構(打不完會留著,下次再打接續)
	Armor     int // 目前剩餘裝甲；舊存檔沒有此欄時維持 0，不把受傷狀態擅自補滿。
	// TransitETA 對應 owner 8 ship status 1/2 的剩餘航程。StarIndex 在航行中是目的星；
	// 只有 ETA==0 才是停泊守衛。舊 JSON 沒有此欄位，零值自然維持原有停泊語意。
	TransitETA int `json:"transitETA,omitempty"`
	// Count／EelAges 只供事件太空鰻：原版每個 type 13 ship 各有 +0x61 age，
	// 每 30 回合分裂且全銀河最多 4 艘。同星聚合仍須保存逐體錯開年齡。
	Count   int   `json:"count,omitempty"`
	EelAges []int `json:"eelAges,omitempty"`
}

func normalizeEelGroup(m *MonsterGuard) int {
	if m == nil || m.Kind != gamedata.MonsterEel {
		return 0
	}
	count := m.Count
	if count < 1 {
		count = len(m.EelAges)
	}
	if count < 1 {
		count = 1 // 舊 JSON 的單一太空鰻沒有 Count／EelAges。
	}
	if len(m.EelAges) > count {
		count = len(m.EelAges)
	}
	for len(m.EelAges) < count {
		m.EelAges = append(m.EelAges, 0)
	}
	m.Count = count
	return count
}

// advanceSpaceEelSplits 對應 sub_DB8D8 的 type 13 停泊分支；航行中怪物不增齡。
func (s *GameSession) advanceSpaceEelSplits() int {
	total := 0
	for i := range s.Monsters {
		total += normalizeEelGroup(&s.Monsters[i])
	}
	stats, ok := gamedata.MonsterStatsFor(gamedata.MonsterEel)
	if !ok {
		return 0
	}
	births := 0
	for i := range s.Monsters {
		m := &s.Monsters[i]
		if m.Kind != gamedata.MonsterEel || m.TransitETA > 0 {
			continue
		}
		existing := len(m.EelAges)
		groupBirths := 0
		for j := 0; j < existing; j++ {
			m.EelAges[j]++
			if m.EelAges[j] != 30 {
				continue
			}
			m.EelAges[j] = 0
			if total < 4 {
				total++
				births++
				groupBirths++
			}
		}
		for j := 0; j < groupBirths; j++ {
			m.EelAges = append(m.EelAges, 0)
			m.Structure += stats.Structure
			m.Armor += stats.Armor
		}
		m.Count = len(m.EelAges)
	}
	return births
}

// eventMonsterTransitETA 重建 sub_A1762 的外圍進場。原版視窗捲動 globals 沒有進入
// remake 狀態，因此出生邊界是明示近似；30 raw units/parsec、無條件進位及 owner 8
// 每回合 1 parsec 則是 sub_EBE79/sub_FF799 的已證實規則。
func (s *GameSession) eventMonsterTransitETA(target int) int {
	if target < 0 || target >= len(s.Stars) {
		return 0
	}
	s.eventRandForTest()
	dim := gamedata.GalaxyDims[s.GalaxySizeClass()]
	tx := s.Stars[target].X * float64(dim.Width)
	ty := s.Stars[target].Y * float64(dim.Height)
	offset := float64(s.eventRand.Intn(5)+1) * gamedata.ParsecUnits
	var rawDistance float64
	switch s.eventRand.Intn(4) {
	case 0:
		rawDistance = tx + offset
	case 1:
		rawDistance = float64(dim.Width) - tx + offset
	case 2:
		rawDistance = ty + offset
	default:
		rawDistance = float64(dim.Height) - ty + offset
	}
	eta := int(math.Ceil(rawDistance / gamedata.ParsecUnits))
	if eta < 1 {
		eta = 1
	}
	return eta
}

// advanceEventMonsterRoutes 推進 owner 8 事件怪物航程；停泊守衛與一般星圖怪物不動。
func (s *GameSession) advanceEventMonsterRoutes() []string {
	var messages []string
	var remove []int
	for i := range s.Monsters {
		m := &s.Monsters[i]
		if m.TransitETA <= 0 {
			continue
		}
		m.TransitETA--
		if m.TransitETA > 0 {
			continue
		}
		m.TransitETA = 0
		if m.Kind == gamedata.MonsterEel {
			normalizeEelGroup(m)
			for j := range m.EelAges {
				m.EelAges[j] = 0
			}
		}
		impact, attacked := s.resolveEventMonsterColonyAttack(m)
		if attacked {
			messages = append(messages, s.monsterColonyImpactMessage(m, impact))
			if impact.MonsterDestroyed {
				remove = append(remove, i)
			}
			continue
		}
		messages = append(messages, fmt.Sprintf("%s已抵達 %s 星系並開始盤據",
			gamedata.MonsterNameZH(m.Kind), s.starName(m.StarIndex)))
	}
	for i := len(remove) - 1; i >= 0; i-- {
		idx := remove[i]
		s.Monsters = append(s.Monsters[:idx], s.Monsters[idx+1:]...)
	}
	return messages
}

// MonsterGroupsAtStar 回傳該星所有已停泊的怪獸群組；航行中 record 不算守衛。
// 原版以 owner/type side bit 分組，同星不同種類不能互相覆蓋或合併。
func (s *GameSession) MonsterGroupsAtStar(starIdx int) []*MonsterGuard {
	groups := make([]*MonsterGuard, 0, 1)
	for i := range s.Monsters {
		if s.Monsters[i].StarIndex == starIdx && s.Monsters[i].TransitETA == 0 {
			groups = append(groups, &s.Monsters[i])
		}
	}
	return groups
}

// MonsterAtStar 回傳守衛該星的第一個怪獸群組；沒有回 nil。多群組時採穩定切片順序，
// 是玩家主動攻擊的可重播 adapter。原版 Search_For_Battles_ 會在全銀河自動戰鬥排程中
// 隨機排列並逐一消費所有 side；那不是單次「攻擊怪獸」指令可直接套用的選擇器。
func (s *GameSession) MonsterAtStar(starIdx int) *MonsterGuard {
	groups := s.MonsterGroupsAtStar(starIdx)
	if len(groups) > 0 {
		return groups[0]
	}
	return nil
}

// StarGuardedByMonster 回傳該星是否有怪獸把守(對應原版 `Star_Guarded_By_Monster_` @ 0x7A47A)。
func (s *GameSession) StarGuardedByMonster(starIdx int) bool {
	return s.MonsterAtStar(starIdx) != nil
}

// MonsterNameAtStar 回傳該星全部停泊怪獸種類；同種類多群只列一次，避免 UI 隱藏
// 第二個 owner/type side。沒有怪獸回空字串。
func (s *GameSession) MonsterNameAtStar(starIdx int) string {
	seen := map[gamedata.SpaceMonster]bool{}
	names := make([]string, 0, 1)
	for _, m := range s.MonsterGroupsAtStar(starIdx) {
		if seen[m.Kind] {
			continue
		}
		seen[m.Kind] = true
		names = append(names, gamedata.MonsterNameZH(m.Kind))
	}
	return strings.Join(names, "、")
}

// genMonsters 在星圖生成時擺放守衛怪獸(對應原版 `Make_System_Monsters_` @ 0x7CDC5)。
//
// 排除:母星(玩家與 AI)、黑洞/無行星的星系。手冊 p.60 的「有怪獸的星系一定另有一個特殊
// 物產」由本函式**主動補上**——被選中的星如果原本沒有特殊物產,就骰一個非「無」的給它。
func genMonsters(stars []Star, planets []Planet, r *rand.Rand, homeStars map[int]bool) []MonsterGuard {
	n := gamedata.GuardMonsterCountFor(len(stars))
	if n <= 0 {
		return nil
	}
	// 候選星:無主、非母星、有行星。
	var cands []int
	for i := range stars {
		if homeStars[i] || stars[i].Owner != 0 {
			continue
		}
		// ⚠ 不能寫 `planets[i]` ——`Planets` 自 2026-08-07(第 24 項(軌道資料層))起**不再與 Stars 平行**,
		// 一顆星有 1..5 個天體。要挑代表行星請走 representativePlanet(唯一那一份實作)。
		if p := representativePlanet(stars, planets, i); p >= 0 && planets[p].NoPlanet {
			continue
		}
		cands = append(cands, i)
	}
	if len(cands) == 0 {
		return nil
	}
	r.Shuffle(len(cands), func(a, b int) { cands[a], cands[b] = cands[b], cands[a] })
	if n > len(cands) {
		n = len(cands)
	}

	out := make([]MonsterGuard, 0, n)
	for k := 0; k < n; k++ {
		idx := cands[k]
		kind := gamedata.RollGuardMonster(r.Intn(len(gamedata.GuardStarMonsters)) + 1)
		st, ok := gamedata.MonsterStatsFor(kind)
		if !ok {
			continue
		}
		out = append(out, MonsterGuard{StarIndex: idx, Kind: kind, Structure: st.Structure, Armor: st.Armor})

		// 手冊 p.60:有怪獸的星系一定另有一個特殊物產。原本沒有的話補一個
		// (從權重表重骰,骰到「無」就再骰,最多幾次——骰不到就算了,不硬塞)。
		if p := representativePlanet(stars, planets, idx); p >= 0 && planets[p].SpecialID == gamedata.NoSpecial {
			for try := 0; try < 12; try++ {
				sp := gamedata.RollPlanetSpecial(r)
				if sp != gamedata.NoSpecial {
					planets[p].SpecialID = sp
					break
				}
			}
		}
	}
	return out
}

// MonsterBattleResult 是一次挑戰怪獸的結果。
type MonsterBattleResult struct {
	Ok             bool                     // 是否真的發生戰鬥(false = 前置條件不足)
	Reason         MonsterCombatRefusalCode // Ok=false 時的穩定原因碼
	MonsterKind    gamedata.SpaceMonster    // UI 如需顯示名稱，必須由外部 catalog 轉譯
	Won            bool                     // 玩家是否清除了怪獸
	Damage         int                      // 這次對怪獸造成的裝甲＋結構傷害
	Remaining      int                      // 怪獸剩餘結構(Won=true 時為 0)
	RemainingArmor int                      // 怪獸剩餘裝甲
	ShipsLost      int                      // 玩家損失艦艇數
}

type MonsterCombatRefusalCode string

const (
	MonsterCombatNoMonster       MonsterCombatRefusalCode = "no_monster"
	MonsterCombatFleetNotPresent MonsterCombatRefusalCode = "fleet_not_present"
	MonsterCombatNoBlueprint     MonsterCombatRefusalCode = "no_blueprint"
	MonsterCombatNoCombatShip    MonsterCombatRefusalCode = "no_combat_ship"
	MonsterCombatInvalidData     MonsterCombatRefusalCode = "invalid_data"
)

func (c MonsterCombatRefusalCode) String() string { return string(c) }

// AttackMonster 讓停在該星的玩家艦隊挑戰守衛怪獸。
//
// 解算沿用既有的艦隊戰鬥模型(手冊 p.30:怪獸「for tracking and combat purposes, they're
// treated as fleets」),但怪獸是**單一高結構目標**而不是一支艦隊:
//   - 玩家艦隊的總火力打在怪獸的結構上,打光就清除
//   - 怪獸每回合反擊一次,傷害取手冊 p.114 的該怪獸傷害範圍;必中的怪獸(海德拉/巨龍)
//     跳過命中判定
//   - 打不完的話怪獸留著剩餘結構,下次再打接續(不會回血——手冊沒提怪獸會恢復)
func (s *GameSession) AttackMonster(starIdx int) MonsterBattleResult {
	s.recordPlayerCommand(PlayerCommand{Name: CmdAttackMonster, Args: []int{starIdx}})
	m := s.MonsterAtStar(starIdx)
	if m == nil {
		return MonsterBattleResult{Reason: MonsterCombatNoMonster}
	}
	if s.Fleet().AtStar != starIdx || s.Fleet().ETA != 0 {
		return MonsterBattleResult{Reason: MonsterCombatFleetNotPresent}
	}
	st, ok := gamedata.MonsterStatsFor(m.Kind)
	if !ok {
		return MonsterBattleResult{Reason: MonsterCombatInvalidData}
	}
	pf, pfIdx := s.mkPlayerCombatantsIndexed()
	if len(pf) == 0 {
		return MonsterBattleResult{Reason: MonsterCombatNoCombatShip}
	}

	res := MonsterBattleResult{Ok: true, MonsterKind: m.Kind}
	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + int64(starIdx)*7919 + 31))
	galacticLoreBonus := s.galacticLoreCombatBonus()

	// 六回合上限,與 ResolveBattle 一致(避免無限纏鬥)。
	for round := 1; round <= 6 && m.Structure > 0 && len(pf) > 0; round++ {
		// 我方齊射:每艘船打一次,傷害取該船的 wmin..wmax。
		for _, c := range pf {
			dmg := c.wmin
			if c.wmax > c.wmin {
				dmg += rng.Intn(c.wmax - c.wmin + 1)
			}
			// Galactic Lore 的 +5 是怪獸戰鬥唯一可觀察的「combat」欄位；
			// 本 remake 的怪獸路徑沒有另擲 BA 命中骰，因此以每艘參戰艦
			// 的固定齊射加成承接，避免技能只在星圖顯示上生效。
			dmg += galacticLoreBonus
			if dmg < 1 {
				dmg = 1
			}
			dealt := dmg
			absorbed := min(m.Armor, dmg)
			m.Armor -= absorbed
			dmg -= absorbed
			m.Structure -= dmg
			res.Damage += dealt
			if m.Structure <= 0 {
				break
			}
		}
		if m.Structure <= 0 {
			break
		}
		// 怪獸反擊：事件怪物逐槽、逐門開火；Guardian 才保留舊的一發近似。
		if blueprint, exact := gamedata.MonsterBlueprintFor(m.Kind); exact {
			for monster := 0; monster < monsterGroupCount(m) && len(pf) > 0; monster++ {
				for _, mount := range blueprint.Weapons {
					weapon, known := gamedata.OrigWeaponByID(mount.WeaponID)
					if !known || weapon.Cat == gamedata.WeaponCatBomb {
						continue
					}
					lo, hi, damaging := gamedata.MonsterWeaponQuickDamageRange(mount)
					if !damaging {
						continue
					}
					if mount.WeaponID == 44 {
						// 快速結算沒有格位；保留原版「全範圍、雙方艦艇、尺寸分段」中的玩家
						// 艦隊消費端，並以中心距離承接全部抽象 combatant。一次 mount 聚合所有門。
						base := 0
						for shot := 0; shot < mount.Count; shot++ {
							rolled := lo
							if hi > lo {
								rolled += rng.Intn(hi - lo + 1)
							}
							base += rolled
						}
						for target := len(pf) - 1; target >= 0; target-- {
							dmg := PlasmaFluxSizeDamage(base, pf[target].sizeClass, func(limit int) int {
								return rng.Intn(limit) + 1
							})
							if dmg >= pf[target].hp {
								pf = append(pf[:target], pf[target+1:]...)
								pfIdx = append(pfIdx[:target], pfIdx[target+1:]...)
								res.ShipsLost++
							} else {
								pf[target].hp -= dmg
							}
						}
						continue
					}
					for shot := 0; shot < mount.Count && len(pf) > 0; shot++ {
						if !gamedata.MonsterWeaponAlwaysHits(mount.WeaponID) && rng.Intn(100) >= 70 {
							continue
						}
						dmg := lo
						if hi > lo {
							dmg += rng.Intn(hi - lo + 1)
						}
						if target := weakestMonsterTargetIndex(pf); target >= 0 {
							switch mount.WeaponID {
							case 45:
								// 快速結算沒有四面護盾與跨回合 +0x43；以同一擲值
								// 乘四承接原版四面包覆，屬規格明載的 remake 近似。
								dmg *= 4
							}
						}
						pf, pfIdx, res.ShipsLost = damageWeakestMonsterTarget(pf, pfIdx, dmg, res.ShipsLost)
					}
				}
			}
		} else if st.DamageMax > 0 && (st.AlwaysHits || rng.Intn(100) < 70) {
			dmg := st.DamageMin
			if st.DamageMax > st.DamageMin {
				dmg += rng.Intn(st.DamageMax - st.DamageMin + 1)
			}
			pf, pfIdx, res.ShipsLost = damageWeakestMonsterTarget(pf, pfIdx, dmg, res.ShipsLost)
		}
		// 自動修復元件:每回合修復 20% 的結構損傷(手冊 p.82)。
		for i := range pf {
			if !pf[i].autoRepair || pf[i].maxHP <= 0 {
				continue
			}
			if r := autoRepairInCombat(pf[i].maxHP - pf[i].hp); r > 0 {
				pf[i].hp += r
			}
		}
	}

	if m.Structure <= 0 {
		res.Won = true
		res.Remaining = 0
		s.removeMonsterAt(starIdx)
	} else {
		res.Remaining = m.Structure
		res.RemainingArmor = m.Armor
	}
	// 倖存艦帶傷回港(見 repair.go);先寫損傷再移除陣亡艦,順序與 ResolveBattle 一致。
	s.applySurvivorDamage(pf, pfIdx)
	for i := 0; i < res.ShipsLost; i++ {
		s.removeWeakestShip()
	}
	s.repairAfterBattle(res.Won) // 自動修復/進階損害管制/工程師(手冊 p.80/p.82/p.136)

	return res
}

func damageWeakestMonsterTarget(pf []combatant, pfIdx []int, damage, lost int) ([]combatant, []int, int) {
	if len(pf) == 0 || damage <= 0 {
		return pf, pfIdx, lost
	}
	weakest := weakestMonsterTargetIndex(pf)
	if damage >= pf[weakest].hp {
		pf = append(pf[:weakest], pf[weakest+1:]...)
		pfIdx = append(pfIdx[:weakest], pfIdx[weakest+1:]...)
		return pf, pfIdx, lost + 1
	}
	pf[weakest].hp -= damage
	return pf, pfIdx, lost
}

func weakestMonsterTargetIndex(pf []combatant) int {
	if len(pf) == 0 {
		return -1
	}
	weakest := 0
	for i := range pf {
		if pf[i].hp < pf[weakest].hp {
			weakest = i
		}
	}
	return weakest
}

// removeMonsterAt 從清單中移除該星的怪獸。
func (s *GameSession) removeMonsterAt(starIdx int) {
	for i := range s.Monsters {
		if s.Monsters[i].StarIndex == starIdx && s.Monsters[i].TransitETA == 0 {
			s.Monsters = append(s.Monsters[:i], s.Monsters[i+1:]...)
			return
		}
	}
}
