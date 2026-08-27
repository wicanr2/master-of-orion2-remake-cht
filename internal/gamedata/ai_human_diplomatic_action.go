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
	SourceIncome         int
	TargetCredits        int
	CreditIntensityLimit int
	ColonyTarget         int
}

type OriginalHumanDiplomaticAction struct {
	Kind       OriginalHumanDiplomaticActionKind
	Credits    int
	Technology int
	Colony     int
	DirectTier int
}

// OriginalHumanDiplomaticActionSelect 對應 sub_4F93B @ 0x4F93B..0x4FD30
// 的候選 gate、RNG 順序與四種 payload。候選科技與殖民地由上游提供。
func OriginalHumanDiplomaticActionSelect(in OriginalHumanDiplomaticActionInput,
	roll func(int) int) (OriginalHumanDiplomaticAction, bool) {
	if in.Intensity < 0 || in.TechnologyRatioLimit < 0 || in.SourceIncome < 0 ||
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
	if intensity < 6 || in.ColonyTarget < 0 {
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
		if r-1 == 1 || (!direct && !tech && !credits) {
			kind = OriginalHumanDiplomaticActionColony
		}
	}
	if kind == OriginalHumanDiplomaticActionNone && tech {
		r := roll(3)
		if r < 1 || r > 3 {
			return OriginalHumanDiplomaticAction{}, false
		}
		if r-1 == 1 || (!direct && !credits) {
			kind = OriginalHumanDiplomaticActionTechnology
		}
	}
	if kind == OriginalHumanDiplomaticActionNone && direct {
		r := roll(2)
		if r < 1 || r > 2 {
			return OriginalHumanDiplomaticAction{}, false
		}
		if r-1 == 1 || !credits {
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
		income := in.SourceIncome
		if income < 10 {
			income = 10
		}
		amount := 10 * intensity * income
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
			(len(in.TechnologyCandidates)*intensity/in.TechnologyRatioLimit + (r - 1) - 2)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(in.TechnologyCandidates) {
			idx = len(in.TechnologyCandidates) - 1
		}
		out.Technology = in.TechnologyCandidates[idx]
	case OriginalHumanDiplomaticActionColony:
		out.Colony = in.ColonyTarget
	default:
		return OriginalHumanDiplomaticAction{}, true
	}
	return out, true
}
