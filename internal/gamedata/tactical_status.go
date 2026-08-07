package gamedata

// tactical_status.go:戰術戰鬥的**狀態類武器**——牽引光束與停滯力場。
//
// ============ 為什麼它們拖到第 138 項才做 ============
//
// 第 128 項盤點手冊 p.127 的特殊武器時,把這兩個判定成「卡在機制」:
//
//	牽引光束    reduces target speed  → 戰鬥速度模型
//	停滯力場    suspended animation   → 「這一輪不能動」的狀態
//
// 第 136 項建了戰鬥速度、第 137 項建了移動預算。**缺的最後一塊是「讓對方動不了」**
// ——那是逐艦狀態,不是自己的能力。這一項補上。
//
// ============ 手冊給的規則比想像中精確 ============
//
// 牽引光束:
//
//	Each beam can **trap a Frigate class ship or slow a larger one — in proportion to its
//	size** — up to the maximum range of **12 squares** away. The effect of multiple Tractor
//	Beams on a single target is **cumulative**. (Thus, for example, **6 beams would
//	immobilize a Doom Star**.) A slowed or trapped ship can move or turn only according to
//	its new speed… An immobilized ship receives an additional **−20 Ship Defense** penalty
//	and can be boarded…
//
// 那個括號裡的例子把公式釘死了:末日之星是第 5 級,而 6 = 5+1。所以
// **困住一艘船需要「級數 + 1」束**,每束扣掉 1/(級數+1) 的速度。
// 這與第 127 項球形武器的「級數 = index+1」是同一個讀法,兩處互相支持。
//
// 停滯力場:
//
//	While suspended, the ship cannot move, fire, recharge any of its weapons or shields,
//	cloak, retreat, **or be affected by any weapon**. It is effectively **removed from
//	battle entirely**. The field has a range of only **3 squares**…
//
// 「or be affected by any weapon」很重要:被定住的船**不能打也不能被打**。
// 只做「不能動」會讓它變成活靶,那是相反的效果。

const (
	// TractorBeamRangeSquares 手冊:「up to the maximum range of 12 squares away」。
	// 這是**原版棋盤**的格數(81×68),換算到 remake 盤面見 shell.TacticalRangeSquares。
	TractorBeamRangeSquares = 12
	// StasisFieldRangeSquares 手冊:「The field has a range of only 3 squares」。
	StasisFieldRangeSquares = 3
	// TractorImmobilizedDefensePenalty 手冊:「An immobilized ship receives an additional
	// −20 Ship Defense penalty」。
	TractorImmobilizedDefensePenalty = -20
)

// TractorBeamsToImmobilize 回傳完全定住某級艦體需要幾束牽引光束。
//
// 手冊只給了一個例子(末日之星 = 6 束),而末日之星是第 5 級——所以是「級數 + 1」。
// 巡防艦(第 0 級)= 1 束,正好對上「Each beam can **trap a Frigate class ship**」。
// **兩端都對上,中間就是線性內插**,不是猜的。
func TractorBeamsToImmobilize(class CombatShipClass) int {
	if int(class) < 0 {
		return 1
	}
	return int(class) + 1
}

// TractorSlowedSpeed 回傳被 beams 束牽引光束拉住之後的速度。
//
// 手冊:「slow a larger one — **in proportion to its size**」+「cumulative」。
// 每束扣掉 1/(級數+1),扣滿就是 0(完全定住)。
func TractorSlowedSpeed(speed int, class CombatShipClass, beams int) int {
	need := TractorBeamsToImmobilize(class)
	if beams <= 0 {
		return speed
	}
	if beams >= need {
		return 0
	}
	return speed * (need - beams) / need
}

// TractorIsImmobilized 回報這艘船是否已被完全定住。
func TractorIsImmobilized(class CombatShipClass, beams int) bool {
	return beams >= TractorBeamsToImmobilize(class)
}
