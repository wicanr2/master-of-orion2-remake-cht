package shell

import (
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// drive_level.go:引擎階模型——remake 先前**根本沒有**這個東西。
//
// ============ 怎麼發現的 ============
//
// `gamedata` 裡有一整組吃 `ftlLevel` 的公式:飛彈速度、飛彈的光束閃避、戰機速度……
// 註解都寫著「速度 = 基礎 + 2×(FTL−1)」。而唯一的生產端呼叫長這樣:
//
//	shell.NewFighterSquadron(s.BayKind, false, idx, s.Col, s.Row, **1**, 0)
//
// **硬編成 1。** 它在 `cmd/moo2/tacticalfighter.go`,所以第 66 項那個
// 「哪些 gamedata 參數被餵固定值」的掃描器(只掃 `gamedata.X(...)`)看不到它
// ——**掃描器的盲區也要記帳**,不然「已經掃過了」會變成另一句過期的擋門理由。
//
// ============ 為什麼是六階 ============
//
// 執行檔的戰鬥速度表(見 gamedata/combat_speed.go)有 **6 列**有效資料,而 MOO2 剛好
// 有六種 FTL 引擎:核融合裂變 → 融合 → 離子 → 反物質 → 超空間 → 相位。兩邊對得上,
// 不是湊出來的。

// driveLevel 回傳玩家目前的引擎階(1..6);一項都沒有時回 1(核子引擎是起始科技)。
//
// 取**最高**已研究的那一階:引擎在 MOO2 是自動升級的(手冊 Fusion Drive:「This drive is
// added to all your ships that are not in hyperspace **as soon as you complete your
// research**」),不是逐艦選裝。
func (s *GameSession) driveLevel() int {
	best := 1
	for level := gamedata.CombatSpeedDriveLevels; level > 1; level-- {
		tech := gamedata.DriveTechForLevel(level)
		topic, ok := gamedata.OrigTechTopic(tech)
		if !ok {
			continue
		}
		if groundEquipTechOwned(s.Player, topic, tech) {
			best = level
			break
		}
	}
	return best
}

// shipCombatSpeed 回傳這艘船的戰鬥速度(引擎階 + 艦體 + 增強引擎 + 跨維度種族)。
func (s *GameSession) shipCombatSpeed(sh Ship) int {
	class, ok := shipClassFromName(sh.Class)
	if !ok {
		return 0
	}
	return gamedata.ShipCombatSpeed(s.driveLevel(), class,
		sh.Special == augmentedEnginesName,
		s.raceHasTrait(gamedata.TRAIT_TRANS_DIMENSIONAL))
}

// augmentedEnginesName 是增強引擎的元件名(手冊「Augmented Engines」:戰鬥速度 +5)。
const augmentedEnginesName = "增強引擎"

// sortByInitiative 依主動權由高到低排序戰列。
//
// 手冊逐字:「A ship's initiative is equal to its current Beam Attack plus 10 times its
// current combat speed. Thus, **smaller ships should move before bigger, slower ones**.」
//
// ⚠ **用穩定排序**:主動權相同的船必須維持原本的相對次序,否則同一場戰鬥在不同機器上
// (或 Go 版本更新後)可能打出不同結果——那會讓存檔與探針的可重現性整個失效。
func sortByInitiative(cs []combatant) {
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].initiative > cs[j].initiative })
}

// --- 戰術棋盤比例尺(第 70 項)---
//
// remake 的戰術棋盤是 **8 × 6**,原版是 **81 × 68**(見 gamedata.CombatGridColumns,
// 從 `Assign_Combat_Grids_` 的清空迴圈界限挖出來)。兩邊差約 **10 倍**:
//
//	81 / 8 ≈ 10.1      68 / 6 ≈ 11.3
//
// 所以原版的速度值(13..30)換算到 remake 盤面是 **1..3 格**。
//
// ⚠ **「remake 用 8×6」本身是 remake 的簡化**,不是原版的東西;所以這個比例尺是一個
// **明說的 remake 決定**,不是轉寫。但它至少是**從一手尺寸推出來的**決定,而不是
// 「覺得走 2 格差不多」——差別在於:原版尺寸改了(或以後把棋盤放大),這個換算會跟著對。
//
// 保留原版的相對關係才是重點:小船走得比大船遠、引擎升級走得更遠、增強引擎再多一點。
// 下限 1 格是刻意的——速度再慢的船也要能動,否則末日之星在低引擎階會完全無法移動,
// 那不是手冊講的東西(手冊只說大船比較慢)。

// TacticalGridColumns / TacticalGridRows 是 remake 戰術棋盤的尺寸。
//
// 與 `cmd/moo2` 的 `gcCols`/`gcRows` 必須一致——`drive_level_test.go` 沒辦法從 shell
// 檢查 cmd 的常數,所以那一側自己有一條測試釘住兩邊相同。
const (
	TacticalGridColumns = 8
	TacticalGridRows    = 6
)

// TacticalMoveSquares 把原版的戰鬥速度換算成 remake 棋盤上一回合可走的格數。
//
// 用外圈(欄)的比例:兩軸的比例分別是 10.1 與 11.3,取較小的那個會讓縱向移動偏慢,
// 而 remake 的戰列是橫向排開的(玩家在左欄、敵方在右欄),橫向距離才是主要的。
// 下限 1(見上方說明)。
func TacticalMoveSquares(combatSpeed int) int {
	if combatSpeed <= 0 {
		return 0 // 沒有引擎的船不能動
	}
	n := combatSpeed * TacticalGridColumns / gamedata.CombatGridColumns
	if n < 1 {
		n = 1
	}
	return n
}

// TacticalRangeSquares 把原版棋盤的格數換算成 remake 盤面的格數(第 70 項)。
//
// 與 TacticalMoveSquares 同一個比例尺(1:10),但**向上取整**:手冊的短射程
// (停滯力場 3 格)向下取整會變成 0,那等於這個武器不存在。向上取整之後
// 牽引光束 12 格 → 2、停滯力場 3 格 → 1,**「牽引比停滯遠」這個相對關係保住了**
// ——那才是手冊在講的事,絕對值在 8 欄的盤面上本來就表達不出來。
//
// ⚠ **既有的 `fireRange = 4` 不走這條路。** 它是 remake 自訂的盤面射程(不是換算來的),
// 改它會動到現有的戰鬥平衡。兩套射程並存是一個已知的不一致,記在這裡而不是假裝沒有。
func TacticalRangeSquares(origSquares int) int {
	if origSquares <= 0 {
		return 0
	}
	n := (origSquares*TacticalGridColumns + gamedata.CombatGridColumns - 1) / gamedata.CombatGridColumns
	if n < 1 {
		n = 1
	}
	return n
}

// ApplyTacticalStatusEffects 依雙方位置重算「誰被牽引、誰被定住」(第 70 項)。
//
// **每回合重算,不累積。** 手冊把兩者都描述成**持續的場**而不是打出去的一發:
//
//	停滯力場:「it remains in effect **as long as the ship generating the field remains
//	          undestroyed and in combat**」
//	牽引光束:「**up to the maximum range of** 12 squares away」——離開射程就鬆開
//
// 所以每回合從零重算才是對的:產生源被打掉、或目標飛出射程,效果就該消失。
// 累加的話會出現「產生源早就沒了,目標還定在那裡」。
//
// dist 用曼哈頓距離,與射程/移動同一個度量(理由見第 70 項)。
func ApplyTacticalStatusEffects(a, b []CombatShip) {
	apply := func(targets, sources []CombatShip) {
		for i := range targets {
			t := &targets[i]
			t.HeldByTractors, t.InStasis = 0, false
			for _, s := range sources {
				if s.HP <= 0 {
					continue // 產生源已被擊毀 → 效果消失
				}
				d := absInt(s.Col-t.Col) + absInt(s.Row-t.Row)
				if s.TractorBeams > 0 && d <= TacticalRangeSquares(gamedata.TractorBeamRangeSquares) {
					t.HeldByTractors += s.TractorBeams
				}
				if s.HasStasisField && d <= TacticalRangeSquares(gamedata.StasisFieldRangeSquares) {
					t.InStasis = true
				}
			}
		}
	}
	apply(a, b)
	apply(b, a)
}

// TacticalEffectiveSpeed 回傳這艘船在目前狀態下的實際戰鬥速度。
func TacticalEffectiveSpeed(sh CombatShip) int {
	if sh.InStasis {
		return 0 // 手冊:「the ship cannot move」
	}
	return gamedata.TractorSlowedSpeed(sh.CombatSpeed, sh.SizeClass, sh.HeldByTractors)
}

// TacticalEffectiveDefense 回傳這艘船在目前狀態下的實際防禦。
//
// 手冊:「An immobilized ship receives an **additional −20 Ship Defense** penalty」。
// 只有**完全定住**才扣——被拖慢但還能動的不扣(手冊那句的主詞是 immobilized)。
func TacticalEffectiveDefense(sh CombatShip) int {
	def := sh.Defense
	if gamedata.TractorIsImmobilized(sh.SizeClass, sh.HeldByTractors) {
		def += gamedata.TractorImmobilizedDefensePenalty
	}
	if def < 0 {
		def = 0
	}
	return def
}
