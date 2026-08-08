package shell

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

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
}

// MonsterAtStar 回傳守衛該星的怪獸;沒有回 nil。
func (s *GameSession) MonsterAtStar(starIdx int) *MonsterGuard {
	for i := range s.Monsters {
		if s.Monsters[i].StarIndex == starIdx {
			return &s.Monsters[i]
		}
	}
	return nil
}

// StarGuardedByMonster 回傳該星是否有怪獸把守(對應原版 `Star_Guarded_By_Monster_` @ 0x7A47A)。
func (s *GameSession) StarGuardedByMonster(starIdx int) bool {
	return s.MonsterAtStar(starIdx) != nil
}

// MonsterNameAtStar 回傳守衛該星的怪獸中文名;沒有回空字串(供 UI 顯示)。
func (s *GameSession) MonsterNameAtStar(starIdx int) string {
	if m := s.MonsterAtStar(starIdx); m != nil {
		return gamedata.MonsterNameZH(m.Kind)
	}
	return ""
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
		out = append(out, MonsterGuard{StarIndex: idx, Kind: kind, Structure: st.Structure})

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
	Ok        bool   // 是否真的發生戰鬥(false = 前置條件不足)
	Reason    string // Ok=false 時的原因
	Name      string // 怪獸中文名
	Won       bool   // 玩家是否清除了怪獸
	Damage    int    // 這次對怪獸造成的結構傷害
	Remaining int    // 怪獸剩餘結構(Won=true 時為 0)
	ShipsLost int    // 玩家損失艦艇數
	Message   string // 已填好數字的敘述
}

// AttackMonster 讓停在該星的玩家艦隊挑戰守衛怪獸。
//
// 解算沿用既有的艦隊戰鬥模型(手冊 p.30:怪獸「for tracking and combat purposes, they're
// treated as fleets」),但怪獸是**單一高結構目標**而不是一支艦隊:
//   - 玩家艦隊的總火力打在怪獸的結構上,打光就清除
//   - 怪獸每回合反擊一次,傷害取手冊 p.114 的該怪獸傷害範圍;必中的怪獸(海德拉/巨龍)
//     跳過命中判定
//   - 打不完的話怪獸留著剩餘結構,下次再打接續(不會回血——手冊沒提怪獸會恢復)
func (s *GameSession) AttackMonster(starIdx int) MonsterBattleResult {
	m := s.MonsterAtStar(starIdx)
	if m == nil {
		return MonsterBattleResult{Reason: "該星沒有怪獸"}
	}
	if s.Fleet().AtStar != starIdx || s.Fleet().ETA != 0 {
		return MonsterBattleResult{Reason: "艦隊尚未抵達該星"}
	}
	st, ok := gamedata.MonsterStatsFor(m.Kind)
	if !ok {
		return MonsterBattleResult{Reason: "怪獸資料不存在"}
	}
	pf, pfIdx := s.mkPlayerCombatantsIndexed()
	if len(pf) == 0 {
		return MonsterBattleResult{Reason: "艦隊沒有可戰鬥的艦艇"}
	}

	res := MonsterBattleResult{Ok: true, Name: st.NameZH}
	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + int64(starIdx)*7919 + 31))

	// 六回合上限,與 ResolveBattle 一致(避免無限纏鬥)。
	for round := 1; round <= 6 && m.Structure > 0 && len(pf) > 0; round++ {
		// 我方齊射:每艘船打一次,傷害取該船的 wmin..wmax。
		for _, c := range pf {
			dmg := c.wmin
			if c.wmax > c.wmin {
				dmg += rng.Intn(c.wmax - c.wmin + 1)
			}
			if dmg < 1 {
				dmg = 1
			}
			m.Structure -= dmg
			res.Damage += dmg
			if m.Structure <= 0 {
				break
			}
		}
		if m.Structure <= 0 {
			break
		}
		// 怪獸反擊:一發,打掉最弱的一艘(必中的怪獸不擲命中)。
		hit := st.AlwaysHits || rng.Intn(100) < 70 // 70% 是 remake 的一般命中率
		if hit {
			dmg := st.DamageMin
			if st.DamageMax > st.DamageMin {
				dmg += rng.Intn(st.DamageMax - st.DamageMin + 1)
			}
			// 傷害夠不夠打掉一艘:用最弱艦的 hp 當門檻(抽象結算,不逐艦追血)。
			weakest := 0
			for i := range pf {
				if pf[i].hp < pf[weakest].hp {
					weakest = i
				}
			}
			if dmg >= pf[weakest].hp {
				pf = append(pf[:weakest], pf[weakest+1:]...)
				pfIdx = append(pfIdx[:weakest], pfIdx[weakest+1:]...)
				res.ShipsLost++
			} else {
				pf[weakest].hp -= dmg
			}
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
	}
	// 倖存艦帶傷回港(見 repair.go);先寫損傷再移除陣亡艦,順序與 ResolveBattle 一致。
	s.applySurvivorDamage(pf, pfIdx)
	for i := 0; i < res.ShipsLost; i++ {
		s.removeWeakestShip()
	}
	s.repairAfterBattle(res.Won) // 自動修復/進階損害管制/工程師(手冊 p.80/p.82/p.136)

	if res.Won {
		res.Message = fmt.Sprintf("擊殺%s!該星系已可拓殖(我方損失 %d 艘)", res.Name, res.ShipsLost)
	} else {
		res.Message = fmt.Sprintf("%s仍盤據此地(已造成 %d 點傷害,剩餘 %d;我方損失 %d 艘)",
			res.Name, res.Damage, res.Remaining, res.ShipsLost)
	}
	return res
}

// removeMonsterAt 從清單中移除該星的怪獸。
func (s *GameSession) removeMonsterAt(starIdx int) {
	for i := range s.Monsters {
		if s.Monsters[i].StarIndex == starIdx {
			s.Monsters = append(s.Monsters[:i], s.Monsters[i+1:]...)
			return
		}
	}
}

// monsterBlockReason 回傳「因為怪獸而不能在這顆星動作」的理由;沒有怪獸回空字串。
// 手冊 p.62:殖民船只能在「all space monsters and enemy ships have been cleared from that
// planet's system」之後才能建殖民地。前哨站比照——手冊沒有單獨豁免它,而怪獸就在那裡。
func (s *GameSession) monsterBlockReason(starIdx int) string {
	if m := s.MonsterAtStar(starIdx); m != nil {
		return gamedata.MonsterNameZH(m.Kind) + "盤據此星系,必須先清除才能進駐"
	}
	return ""
}
