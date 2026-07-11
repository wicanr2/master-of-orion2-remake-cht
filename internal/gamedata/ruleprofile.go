// Package gamedata:版本規則 profile(patch 1.3 vs patch 1.5)。
//
// 依 docs/tech/version-1.3-1.5-diff.md §5 設計稿:逐條核對 CHANGELOG_150.TXT(1.50.0–1.50.26
// 全部版本)+ MANUAL_150.html 後,落在本專案「已實作系統」且兩版數字真的不同的項目只有 3 條
// (該文件 §2)。本檔刻意只收這 3 條,不要為了「看起來完整」塞進未確證/未接線的欄位——新差異
// 確證後才加欄位,見該文件 §5.3 擴充路徑。
package gamedata

// GameVersion 對應 CLAUDE.md「主選單選擇 1.3 或 1.5」的兩個選項。
type GameVersion int

const (
	VersionClassic13   GameVersion = iota // 官方最後正式 patch 1.31
	VersionCommunity15                    // 社群非官方 patch 1.50(本專案現行資料的預設來源)
)

// RuleProfile 收斂「已確證、且落在已實作系統上」的版本相依數值。
// ⚠ 刻意保持精簡——只收 docs/tech/version-1.3-1.5-diff.md §2 列出的 3 條,不要為了「看起來完整」
// 塞進未確證或未接線的欄位。新差異確證後才加欄位,見 §5.3 擴充路徑。
//
// RuleProfile 應視為唯讀設定,遊戲開始後不可變(避免中途切版本造成存檔/平衡不一致——原版
// 本身也是「一開局就決定規則集」,無 mid-game 切換)。
type RuleProfile struct {
	Version GameVersion

	// 研究:Hyper-Advanced 第一級科技(8 個 TOPIC_HYPER_* 主題共用同一個成本)。
	// 來源:MANUAL_150.html「Hyper-Advanced Tech Cost Bug」+ CHANGELOG_150.TXT 1.50.9。
	HyperAdvancedLevel1Cost int

	// 戰鬥:電漿砲最大傷害(Component.Value)。來源見 component-values.md。
	// 注意:手冊同時記載最小傷害 6→4,但 Component 結構目前只有單一 Value(最大傷害)欄位,
	// 無法表示最小值差異——這是既有資料模型限制,非本 profile 遺漏。
	PlasmaCannonMaxDamage int

	// 軌道轟炸:fleetBombardDamage 模擬齊射的輪數。來源:CHANGELOG_150.TXT 1.50.9。
	BombardmentVolleys int
}

// Profile13 回傳官方最後正式 patch 1.31 的規則 profile。
func Profile13() RuleProfile {
	return RuleProfile{
		Version:                 VersionClassic13,
		HyperAdvancedLevel1Cost: 15000,
		PlasmaCannonMaxDamage:   30,
		BombardmentVolleys:      5,
	}
}

// Profile15 回傳社群非官方 patch 1.50 的規則 profile。
// 三個值皆 = 本專案現行硬編值(techtree.go/session.go/orbital_bombardment.go 改用 profile 前的
// 常數),故以此 profile 為預設不改變任何現行行為。
func Profile15() RuleProfile {
	return RuleProfile{
		Version:                 VersionCommunity15,
		HyperAdvancedLevel1Cost: 25000, // = 現行 techtree.go 硬編值
		PlasmaCannonMaxDamage:   20,    // = 現行 session.go 硬編值
		BombardmentVolleys:      10,    // = 現行 orbital_bombardment.go 硬編值
	}
}

// hyperAdvancedTopics 是 techtree.go 裡 8 個共用「Hyper-Advanced Lv1 研究成本」硬編 25000 的
// 主題(researchChoices 表中的 TOPIC_HYPER_* 8 列)。
var hyperAdvancedTopics = map[ResearchTopic]bool{
	TOPIC_HYPER_BIOLOGY:      true,
	TOPIC_HYPER_POWER:        true,
	TOPIC_HYPER_PHYSICS:      true,
	TOPIC_HYPER_CONSTRUCTION: true,
	TOPIC_HYPER_FIELDS:       true,
	TOPIC_HYPER_CHEMISTRY:    true,
	TOPIC_HYPER_COMPUTERS:    true,
	TOPIC_HYPER_SOCIOLOGY:    true,
}

// IsHyperAdvancedTopic 回傳 topic 是否為上述 8 個共用 Hyper-Advanced Lv1 研究成本的主題之一。
func IsHyperAdvancedTopic(topic ResearchTopic) bool {
	return hyperAdvancedTopics[topic]
}

// HyperAdvancedCost 回傳指定版本 profile 的 Hyper-Advanced Lv1 研究成本覆寫值(15000=1.3 /
// 25000=1.5)。這是「查詢時覆寫」掛勾:researchChoices 表本身(techtree.go)刻意不動,仍是
// 1.5 的硬編值,呼叫端要套用版本差異時,對 IsHyperAdvancedTopic(topic)==true 的主題改用
// HyperAdvancedCost(profile) 取代 ResearchChoiceFor(topic).Cost。
//
// ⚠ 現況(2026-07-11):本函式尚未接進 engine.RunResearchPhase——該函式簽名只吃
// engine.PlayerState(見 internal/engine/research.go),不含 RuleProfile,也沒有 GameSession
// 可查;要接線需先讓研究流程知道玩家所在的版本 profile,這會動到 engine/shell 邊界的既有資料流
// (RunResearchPhase 呼叫鏈:engine.RunEmpireTurn → shell.EndTurn),屬於比本任務(資料層 profile
// 本身)更大的重構,誠實留待「主選單真的接 1.3 選項」的任務再一併處理。本函式已測試、隨時可用,
// 亦可先供 UI 顯示「若選 1.3,這項科技要花多少 RP」的預覽用途。
func HyperAdvancedCost(p RuleProfile) int {
	return p.HyperAdvancedLevel1Cost
}
