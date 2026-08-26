package main

import (
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// tacticalTurnAction 是格子戰術一個完整回合的單艦行動。ID 在本場唯一，故切片壓縮後
// 仍可重新定位；initiative 只作建立本回合排序用，不在回合中因狀態重算而重排。
type tacticalTurnAction struct {
	enemy      bool
	shipID     int
	initiative int
}

func (t *tacticalScreen) shipInitiativeEnabled() bool {
	return t != nil && t.b != nil && t.b.session.EffectiveGameSettings().ShipInitiative
}

func (t *tacticalScreen) ensureTacticalShipIDs() {
	next := 1
	used := map[int]bool{}
	for i := range t.player {
		if t.player[i].TacticalID > 0 && !used[t.player[i].TacticalID] {
			used[t.player[i].TacticalID] = true
			if t.player[i].TacticalID >= next {
				next = t.player[i].TacticalID + 1
			}
			continue
		}
		for used[next] {
			next++
		}
		t.player[i].TacticalID = next
		used[next] = true
		next++
	}
	for i := range t.enemy {
		if t.enemy[i].TacticalID > 0 && !used[t.enemy[i].TacticalID] {
			used[t.enemy[i].TacticalID] = true
			if t.enemy[i].TacticalID >= next {
				next = t.enemy[i].TacticalID + 1
			}
			continue
		}
		for used[next] {
			next++
		}
		t.enemy[i].TacticalID = next
		used[next] = true
		next++
	}
}

func combatShipIndexByTacticalID(ships []shell.CombatShip, id int) int {
	for i := range ships {
		if ships[i].TacticalID == id && ships[i].HP > 0 {
			return i
		}
	}
	return -1
}

func currentTacticalInitiative(ship shell.CombatShip) int {
	return gamedata.CombatInitiative(ship.Attack, shell.TacticalEffectiveSpeed(ship))
}

// resetInitiativeQueue 在回合交界以當下存活艦重建合併序列。先加入玩家、再加入敵方，
// 配合穩定排序形成規格所記的 remake 同分 tie-break。
func (t *tacticalScreen) resetInitiativeQueue() {
	if !t.shipInitiativeEnabled() {
		t.initiativeQueue = nil
		t.initiativePos = 0
		return
	}
	t.ensureTacticalShipIDs()
	t.initiativeQueue = t.initiativeQueue[:0]
	for i := range t.player {
		if t.player[i].HP > 0 {
			t.initiativeQueue = append(t.initiativeQueue, tacticalTurnAction{
				shipID: t.player[i].TacticalID, initiative: currentTacticalInitiative(t.player[i]),
			})
		}
	}
	for i := range t.enemy {
		if t.enemy[i].HP > 0 {
			t.initiativeQueue = append(t.initiativeQueue, tacticalTurnAction{
				enemy: true, shipID: t.enemy[i].TacticalID, initiative: currentTacticalInitiative(t.enemy[i]),
			})
		}
	}
	sort.SliceStable(t.initiativeQueue, func(i, j int) bool {
		return t.initiativeQueue[i].initiative > t.initiativeQueue[j].initiative
	})
	t.initiativePos = 0
}

func (t *tacticalScreen) currentInitiativePlayerIndex() int {
	if !t.shipInitiativeEnabled() || t.initiativePos < 0 || t.initiativePos >= len(t.initiativeQueue) {
		return -1
	}
	action := t.initiativeQueue[t.initiativePos]
	if action.enemy {
		return -1
	}
	return combatShipIndexByTacticalID(t.player, action.shipID)
}

func (t *tacticalScreen) weakestLivingPlayerIndex() int {
	best := -1
	for i := range t.player {
		if t.player[i].HP <= 0 {
			continue
		}
		if best < 0 || t.player[i].HP < t.player[best].HP {
			best = i
		}
	}
	return best
}

// compactPlayerCasualtiesByID 與敵艦壓縮不同，還要同步玩家行動、等待與移動預算。
// 所有平行狀態以 TacticalID 回填，避免前方艦艇陣亡造成錯位。
func (t *tacticalScreen) compactPlayerCasualtiesByID() {
	acted := map[int]bool{}
	waited := map[int]bool{}
	moveLeft := map[int]int{}
	for i := range t.player {
		id := t.player[i].TacticalID
		if i < len(t.acted) {
			acted[id] = t.acted[i]
		}
		if i < len(t.waited) {
			waited[id] = t.waited[i]
		}
		if i < len(t.moveLeft) {
			moveLeft[id] = t.moveLeft[i]
		}
	}
	alive := t.player[:0]
	for _, ship := range t.player {
		if ship.HP > 0 {
			alive = append(alive, ship)
		}
	}
	t.player = alive
	t.acted = make([]bool, len(t.player))
	t.waited = make([]bool, len(t.player))
	t.moveLeft = make([]int, len(t.player))
	for i := range t.player {
		id := t.player[i].TacticalID
		t.acted[i] = acted[id]
		t.waited[i] = waited[id]
		t.moveLeft[i] = moveLeft[id]
	}
}

// advanceInitiativeQueue 自動執行所有排在下一艘玩家艦之前的敵艦。敵方仍沿用既有
// typed 武器解算與最脆弱目標策略；本切片只忠實化行動順序，不自創 AI 移動。
func (t *tacticalScreen) advanceInitiativeQueue() {
	if !t.shipInitiativeEnabled() || t.over {
		return
	}
	for t.initiativePos < len(t.initiativeQueue) {
		action := t.initiativeQueue[t.initiativePos]
		if !action.enemy {
			idx := combatShipIndexByTacticalID(t.player, action.shipID)
			if idx >= 0 && (idx >= len(t.acted) || !t.acted[idx]) {
				t.sel = idx
				return
			}
			t.initiativePos++
			continue
		}
		enemyIndex := combatShipIndexByTacticalID(t.enemy, action.shipID)
		playerIndex := t.weakestLivingPlayerIndex()
		if enemyIndex >= 0 && playerIndex >= 0 {
			t.initiativeEnemyDamage += t.enemyRetaliationDamage(enemyIndex, playerIndex)
			t.compactPlayerCasualtiesByID()
		}
		t.initiativePos++
	}
	t.sel = -1
	t.finishRound(0, 0, false, false, 0)
}

func (t *tacticalScreen) completeInitiativePlayerAction(actor int) {
	if !t.shipInitiativeEnabled() || actor < 0 || actor >= len(t.player) {
		return
	}
	t.acted[actor] = true
	t.waited[actor] = false
	if t.initiativePos < len(t.initiativeQueue) {
		t.initiativePos++
	}
	t.advanceInitiativeQueue()
}

func (t *tacticalScreen) waitInitiativePlayerAction(actor int) bool {
	if !t.shipInitiativeEnabled() || actor < 0 || actor >= len(t.player) ||
		t.initiativePos >= len(t.initiativeQueue) {
		return false
	}
	action := t.initiativeQueue[t.initiativePos]
	if action.enemy || action.shipID != t.player[actor].TacticalID {
		return false
	}
	t.waited[actor] = true
	t.initiativeQueue = append(t.initiativeQueue, action)
	t.initiativePos++
	t.advanceInitiativeQueue()
	return true
}
