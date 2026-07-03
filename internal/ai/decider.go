package ai

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/diplomacy"

// decider.go:AI 決策模式的抽象介面,支援「remake / original」兩種 AI 並存,由玩家選擇
//(如同主選單可選 1.3/1.5 版本,亦可選 AI 模式:remake 或 original)。
//
//   - ModeRemake:本專案的【設計性重建】AI(internal/ai 的設計啟發式,非原版)。
//   - ModeOriginal:盡量還原原版 MOO2 AI 行為(逆向考據,見 docs/tech/original-ai-re.md)。
//     ⚠ 尚待 RE 完成;目前 NewDecider(ModeOriginal, …) 暫以 remake 啟發式代理,待實作後替換。

// AIMode 是 AI 決策模式。
type AIMode int

const (
	ModeRemake   AIMode = iota // 設計性重建 AI
	ModeOriginal               // 逆向原版 AI(待 RE)
)

// Name 回傳模式名稱。
func (m AIMode) Name() string {
	if m == ModeOriginal {
		return "original"
	}
	return "remake"
}

// Decider 是 AI 決策介面。remake 與 original 兩種實作以此統一,供上層依玩家選擇注入。
type Decider interface {
	// ColonyJobs 決定殖民地工作分配(農夫/工人/科學家)。
	ColonyJobs(population, foodPerFarmer int) (farmers, workers, scientists int)
	// TaxRate 依國庫決定稅率(%)。
	TaxRate(treasuryBC int) int
	// ResearchTopic 從候選中選研究主題(回 TopicID,無候選 -1)。
	ResearchTopic(candidates []ResearchCandidate) int
	// BuildPriority 決定殖民地生產優先。
	BuildPriority(threatenedByEnemy, hasColonizableTarget, infrastructureComplete bool) BuildPriority
	// Stance 依關係等級決定外交姿態。
	Stance(level diplomacy.RelationLevel) Stance
	// Mode 回傳此決策器的實際模式。
	Mode() AIMode
}

// RemakeDecider 是設計性重建 AI:包裝 internal/ai 的設計啟發式函式。
type RemakeDecider struct {
	Profile     Profile
	TreasuryLow int
	TreasuryHi  int
}

// NewRemakeDecider 以指定性格建立 remake 決策器(國庫門檻用預設設計值)。
func NewRemakeDecider(p Profile) *RemakeDecider {
	return &RemakeDecider{Profile: p, TreasuryLow: 50, TreasuryHi: 300}
}

func (d *RemakeDecider) ColonyJobs(pop, fpf int) (int, int, int) {
	return DecideColonyJobs(pop, fpf, d.Profile)
}
func (d *RemakeDecider) TaxRate(bc int) int {
	return DecideTaxRate(bc, d.TreasuryLow, d.TreasuryHi)
}
func (d *RemakeDecider) ResearchTopic(c []ResearchCandidate) int {
	return DecideResearchTopic(c, d.Profile)
}
func (d *RemakeDecider) BuildPriority(threat, colonizable, infra bool) BuildPriority {
	return DecideBuildPriority(d.Profile, threat, colonizable, infra)
}
func (d *RemakeDecider) Stance(level diplomacy.RelationLevel) Stance {
	return DecideStance(level, d.Profile)
}
func (d *RemakeDecider) Mode() AIMode { return ModeRemake }

// NewDecider 依模式與性格建立 AI 決策器。
//
// ⚠ ModeOriginal 的逆向原版 AI 尚待實作(見 docs/tech/original-ai-re.md);在完成前,
// 本工廠對 original 模式暫回傳 remake 決策器(行為為設計重建,非原版)。ok 回傳是否已提供
// 所請求的模式:false 表示 fallback 到 remake。上層可據此提示玩家。
func NewDecider(mode AIMode, p Profile) (d Decider, ok bool) {
	switch mode {
	case ModeOriginal:
		// TODO: original AI 實作後改為回傳 OriginalDecider。
		return NewRemakeDecider(p), false
	default:
		return NewRemakeDecider(p), true
	}
}
