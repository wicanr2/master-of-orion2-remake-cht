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
// **硬編成 1。** 它在 `cmd/moo2/tacticalfighter.go`,所以第 129 項那個
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
