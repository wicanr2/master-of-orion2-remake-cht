package diplomacy

// events.go:關係分數的事件調整與每回合演進。【全部為設計值,非原版 MOO2 數字】。

// RelationEvent 是會影響關係分數的事件類型。
type RelationEvent int

const (
	EventDeclareWar         RelationEvent = iota // 宣戰
	EventAttackColony                            // 攻擊殖民地
	EventAttackFleet                             // 攻擊艦隊
	EventSignPeace                               // 簽和平
	EventSignTradeTreaty                         // 簽貿易條約
	EventSignResearchTreaty                      // 簽研究條約
	EventSignAlliance                            // 結盟
	EventBreakTreaty                             // 撕毀條約
	EventGiveGift                                // 贈禮
	EventGiveTributeTurn                         // 進貢(每回合)
	EventSpyCaught                               // 間諜被抓
	EventSharedWarTurn                           // 有共同敵人(每回合)
	EventTradeTurn                               // 貿易往來(每回合)
	EventBorderTensionTurn                       // 邊境擴張壓力(每回合)
)

// relationEventDelta 各事件對關係分數的調整值【設計值】。
var relationEventDelta = map[RelationEvent]int{
	EventDeclareWar:         -40,
	EventAttackColony:       -25,
	EventAttackFleet:        -15,
	EventSignPeace:          +20,
	EventSignTradeTreaty:    +10,
	EventSignResearchTreaty: +10,
	EventSignAlliance:       +30,
	EventBreakTreaty:        -30,
	EventGiveGift:           +15,
	EventGiveTributeTurn:    +10,
	EventSpyCaught:          -20,
	EventSharedWarTurn:      +5,
	EventTradeTurn:          +1,
	EventBorderTensionTurn:  -2,
}

// EventDelta 回傳事件的關係分數調整值(未知事件回 0)。
func EventDelta(e RelationEvent) int { return relationEventDelta[e] }

// RelationState 是一對勢力之間的關係狀態(對稱;各方視角可各持一份)。
type RelationState struct {
	Score int // [-RelationScoreMax, +RelationScoreMax]
}

// clampScore 夾限分數到合法範圍。
func clampScore(s int) int {
	if s < -RelationScoreMax {
		return -RelationScoreMax
	}
	if s > RelationScoreMax {
		return RelationScoreMax
	}
	return s
}

// Apply 套用一個事件,更新並回傳新分數。
func (r *RelationState) Apply(e RelationEvent) int {
	r.Score = clampScore(r.Score + EventDelta(e))
	return r.Score
}

// Level 回傳目前分數對應的關係等級。
func (r *RelationState) Level() RelationLevel { return RelationLevelForScore(r.Score) }

// naturalDriftPerTurn 是每回合關係自然回歸中立的速率【設計值】。
const naturalDriftPerTurn = 1

// AdvanceTurn 每回合演進:無事件時關係往中立(0)漂移 naturalDriftPerTurn。
// 有持續性事件(貿易/共同敵人/邊境壓力/進貢)由呼叫端另行 Apply。
func (r *RelationState) AdvanceTurn() {
	switch {
	case r.Score > 0:
		r.Score = clampScore(r.Score - naturalDriftPerTurn)
	case r.Score < 0:
		r.Score = clampScore(r.Score + naturalDriftPerTurn)
	}
}
