package shell

// BuildNoticeKind 是回合摘要建造通知的穩定種類；顯示句子由 UI 文案目錄提供。
type BuildNoticeKind uint8

const (
	BuildNoticeCompleted BuildNoticeKind = iota
	BuildNoticeRefitCompleted
	BuildNoticeRefitCancelled
	BuildNoticeArtificialPlanet
)

// BuildNotice 只保存建造結果的具型別參數，不在規則層保存玩家句子。
type BuildNotice struct {
	Kind        BuildNoticeKind
	ColonyIndex int
	Name        string
}
