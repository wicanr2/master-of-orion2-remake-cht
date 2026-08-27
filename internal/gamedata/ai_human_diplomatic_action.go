package gamedata

// OriginalHumanDiplomaticActionKind 對應 sub_4F93B 內部 v14 的 1..4。
type OriginalHumanDiplomaticActionKind int

const (
	OriginalHumanDiplomaticActionNone OriginalHumanDiplomaticActionKind = iota
	OriginalHumanDiplomaticActionCredits
	OriginalHumanDiplomaticActionTechnology
	OriginalHumanDiplomaticActionColony
	OriginalHumanDiplomaticActionDirect
)

type OriginalHumanDiplomaticActionInput struct {
	Intensity            int
	DirectEnabled        bool
	TechnologyEnabled    bool
	CreditsEnabled       bool
	ColonyEnabled        bool
	TechnologyCandidates []int
	TechnologyRatioLimit int
	SourceMaintenance    int
	TargetCredits        int
	CreditIntensityLimit int
	ColonyCandidates     []int
}

type OriginalHumanDiplomaticAction struct {
	Kind       OriginalHumanDiplomaticActionKind
	Credits    int
	Technology int
	Colony     int
	DirectTier int
}

// OriginalHumanDiplomaticRequest 保存 sub_53EDB outcome 1／3／4 交給 sub_54CC0
// 前後的玩家可見請求資料。ReasonCode 是方向記錄 +0x657 的 raw 105／106／124；Action
// 是 sub_4F93B 寫入四個暫存 payload 後，由 sub_54CC0 鏡射到雙方方向記錄的 typed 表示。
type OriginalHumanDiplomaticRequest struct {
	Outcome    int
	ReasonCode int
	Action     OriginalHumanDiplomaticAction
}

// OriginalHumanDiplomaticRequestForOutcome 對應 sub_53EDB 三個外交 outcome 的 raw reason。
// outcome 2 是軍事目標，不建立會談請求；0 是無動作。
func OriginalHumanDiplomaticRequestForOutcome(outcome int,
	action OriginalHumanDiplomaticAction) (OriginalHumanDiplomaticRequest, bool) {
	reason := 0
	switch outcome {
	case 1:
		reason = 106
	case 3:
		reason = 105
	case 4:
		reason = 124
	default:
		return OriginalHumanDiplomaticRequest{}, false
	}
	if action.Kind == OriginalHumanDiplomaticActionNone {
		return OriginalHumanDiplomaticRequest{}, false
	}
	return OriginalHumanDiplomaticRequest{Outcome: outcome, ReasonCode: reason, Action: action}, true
}

// OriginalHumanDiplomaticActionSelect 對應 sub_4F93B @ 0x4F93B..0x4FD30
// 的候選 gate、RNG 順序與四種 payload。候選科技與殖民地由上游提供。
func OriginalHumanDiplomaticActionSelect(in OriginalHumanDiplomaticActionInput,
	roll func(int) int) (OriginalHumanDiplomaticAction, bool) {
	if in.Intensity < 0 || in.TechnologyRatioLimit < 0 || in.SourceMaintenance < 0 ||
		in.TargetCredits < 0 || in.CreditIntensityLimit < 0 || roll == nil {
		return OriginalHumanDiplomaticAction{}, false
	}
	intensity := in.Intensity
	if intensity > 10 {
		intensity = 10
	}
	if intensity == 0 {
		return OriginalHumanDiplomaticAction{}, true
	}
	direct, tech, credits, colony := in.DirectEnabled, in.TechnologyEnabled, in.CreditsEnabled, in.ColonyEnabled
	if intensity > 6 {
		direct = false
	}
	if intensity > in.TechnologyRatioLimit || len(in.TechnologyCandidates) == 0 {
		tech = false
	}
	if intensity > in.CreditIntensityLimit {
		credits = false
	}
	if intensity < 6 || len(in.ColonyCandidates) == 0 {
		colony = false
	}
	if !direct && !tech && !credits && !colony {
		return OriginalHumanDiplomaticAction{}, true
	}

	kind := OriginalHumanDiplomaticActionNone
	if colony {
		r := roll(3)
		if r < 1 || r > 3 {
			return OriginalHumanDiplomaticAction{}, false
		}
		if r == 1 || (!direct && !tech && !credits) {
			kind = OriginalHumanDiplomaticActionColony
		}
	}
	if kind == OriginalHumanDiplomaticActionNone && tech {
		r := roll(3)
		if r < 1 || r > 3 {
			return OriginalHumanDiplomaticAction{}, false
		}
		if r == 1 || (!direct && !credits) {
			kind = OriginalHumanDiplomaticActionTechnology
		}
	}
	if kind == OriginalHumanDiplomaticActionNone && direct {
		r := roll(2)
		if r < 1 || r > 2 {
			return OriginalHumanDiplomaticAction{}, false
		}
		if r == 1 || !credits {
			kind = OriginalHumanDiplomaticActionDirect
		}
	}
	if kind == OriginalHumanDiplomaticActionNone && credits {
		kind = OriginalHumanDiplomaticActionCredits
	}

	out := OriginalHumanDiplomaticAction{Kind: kind}
	switch kind {
	case OriginalHumanDiplomaticActionDirect:
		out.DirectTier = 1
		if intensity >= 3 {
			out.DirectTier = 2
		}
	case OriginalHumanDiplomaticActionCredits:
		maintenance := in.SourceMaintenance
		if maintenance < 10 {
			maintenance = 10
		}
		amount := 10 * intensity * maintenance
		if amount > in.TargetCredits {
			amount = in.TargetCredits
		}
		if amount < 100 {
			amount = 10 * (amount / 10)
		} else {
			amount = 100 * (amount / 100)
		}
		if amount <= 0 {
			return OriginalHumanDiplomaticAction{}, true
		}
		if amount > 32000 {
			amount = 32000
		}
		out.Credits = amount
	case OriginalHumanDiplomaticActionTechnology:
		if in.TechnologyRatioLimit < 1 {
			return OriginalHumanDiplomaticAction{}, false
		}
		r := roll(3)
		if r < 1 || r > 3 {
			return OriginalHumanDiplomaticAction{}, false
		}
		idx := len(in.TechnologyCandidates) - 1 -
			(len(in.TechnologyCandidates)*intensity/in.TechnologyRatioLimit + r - 2)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(in.TechnologyCandidates) {
			idx = len(in.TechnologyCandidates) - 1
		}
		out.Technology = in.TechnologyCandidates[idx]
	case OriginalHumanDiplomaticActionColony:
		half := (len(in.ColonyCandidates) + 1) / 2
		r := roll(half)
		if r < 1 || r > half {
			return OriginalHumanDiplomaticAction{}, false
		}
		index := r - 1
		if intensity > 6 {
			index = len(in.ColonyCandidates) - r
		}
		out.Colony = in.ColonyCandidates[index]
	default:
		return OriginalHumanDiplomaticAction{}, true
	}
	return out, true
}
