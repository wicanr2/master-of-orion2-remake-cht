package ai

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/diplomacy"

// diplomacy.go:AI 對某一勢力的外交姿態決策【設計性重建,非原版】。

// Stance 是 AI 決定對某勢力採取的外交姿態。
type Stance int

const (
	StanceWar             Stance = iota // 開戰/維持戰爭
	StanceHostile                       // 敵視(不主動開戰但冷淡)
	StanceNeutral                       // 中立
	StanceProposeTrade                  // 提議貿易/研究條約
	StanceProposeAlliance               // 提議結盟
)

// DecideStance 依目前關係等級與 AI 性格決定外交姿態【設計啟發式】:
//   - 敵對關係(WARY 以下):好戰性格傾向開戰,其餘僅敵視。
//   - 友好關係(AFFABLE 以上):傾向提議條約;非好戰且極友好(UNITY 以上)則提議結盟。
//   - 中立區間:非好戰性格提議貿易,好戰性格保持中立(伺機而動)。
func DecideStance(level diplomacy.RelationLevel, p Profile) Stance {
	warlike := p.IndustryWeight > p.ResearchWeight // 設計:重工業=偏軍事傾向

	switch {
	case level.IsHostile():
		if warlike {
			return StanceWar
		}
		return StanceHostile
	case level.IsFriendly():
		if !warlike && level >= diplomacy.RelationUnity {
			return StanceProposeAlliance
		}
		return StanceProposeTrade
	default: // 中立區間
		if !warlike {
			return StanceProposeTrade
		}
		return StanceNeutral
	}
}
