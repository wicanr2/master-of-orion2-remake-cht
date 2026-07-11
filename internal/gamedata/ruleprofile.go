// Package gamedata:版本規則 profile(patch 1.3 vs patch 1.5)。
//
// 依 docs/tech/version-1.3-1.5-diff.md §5 設計稿:逐條核對 CHANGELOG_150.TXT(1.50.0–1.50.26
// 全部版本)+ MANUAL_150.html 後,落在本專案「已實作系統」且兩版數字真的不同的項目只有 3 條
// (該文件 §2)。本檔刻意只收這 3 條,不要為了「看起來完整」塞進未確證/未接線的欄位——新差異
// 確證後才加欄位,見該文件 §5.3 擴充路徑。
//
// 2026-07-11 補實作 §1 全量表 #5/#7(地面戰/軌道轟炸的真正版本差異項)新增 2 個欄位——
// DefenderCommandoBonus(#5)、BombardmentBuildingBonusHits(#7)。同一批 #6/#8/#9/#11 已交叉
// 核對確認為「非差異項」(兩版預設相同),故不進本檔,只在 ground_version_diff.go 實作一次,
// 詳見 docs/tech/ground-combat-algorithm.md「2026-07-11 版本差異補實作」節。
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

	// 地面戰:防禦方 Commando 領袖加成倍率(套進 gamedata.GroundCommandoDefenderForceBonus 的
	// defenderCommandoBonus 參數)。來源:MANUAL_150.html「A defending commando gives 2.5x the
	// regular commando bonus to ground troops, just like an attacking commando already gives
	// in classic.」1.3=1.0(守方維持「regular commando bonus」基準值 2/3 不變、無額外加乘);
	// 1.5=2.5(守方追平攻方,也套用 2.5x 加乘)。攻方倍率(GroundCommandoAttackerForceBonus)
	// 兩版相同,非差異項,不進本欄位——見 ground_version_diff.go。
	//
	// ⚠ 現況(2026-07-11):gamedata 公式已實作+測試(見 ground_version_diff_test.go),但
	// internal/shell 尚無 AI 對手的領袖資料模型(AIOpponent 無 Leaders 欄位),故
	// InvadeColony 目前只接了攻方(玩家)commando 加成,守方這個欄位還沒有真正的呼叫端——見
	// internal/shell/ground_invasion.go InvadeColony 註解的 TODO 掛鉤點。「真做了 gamedata
	// 公式」與「shell 層已接線」是兩件事,誠實分開標記。
	DefenderCommandoBonus float64

	// 軌道轟炸:轟炸建築額外 +1 hit 的 bug 加成。來源:CHANGELOG_150.TXT 1.50.10「Undocumented
	// +1 hit bonus for civilian buildings during bombardment removed.」1.3=1(有這個未記錄的
	// bug 加成)、1.5=0(已移除)。
	//
	// ⚠ 現況(2026-07-11):本 remake 軌道轟炸目前只扣人口,不扣建築(AI 無 ColonyBuildings
	// 持久資料可扣,見 internal/shell/orbital_bombardment.go BombardColony「範圍限制」),故
	// 本欄位只是資料層佔位 + BombardColony 內留 TODO 掛鉤註解,尚未有任何函式真正讀取它——
	// 建築損傷模型建好前,這個欄位「有欄位、無行為」,不臆測套用到人口損傷上(人口不是建築,
	// 套用會是張冠李戴)。
	BombardmentBuildingBonusHits int
}

// Profile13 回傳官方最後正式 patch 1.31 的規則 profile。
func Profile13() RuleProfile {
	return RuleProfile{
		Version:                      VersionClassic13,
		HyperAdvancedLevel1Cost:      15000,
		PlasmaCannonMaxDamage:        30,
		BombardmentVolleys:           5,
		DefenderCommandoBonus:        1.0, // 守方 Commando 無額外加乘(維持基準值 2/3)
		BombardmentBuildingBonusHits: 1,   // 未記錄的 +1 hit bug(尚未接線,見欄位註解)
	}
}

// Profile15 回傳社群非官方 patch 1.50 的規則 profile。
// 三個值皆 = 本專案現行硬編值(techtree.go/session.go/orbital_bombardment.go 改用 profile 前的
// 常數),故以此 profile 為預設不改變任何現行行為。
func Profile15() RuleProfile {
	return RuleProfile{
		Version:                      VersionCommunity15,
		HyperAdvancedLevel1Cost:      25000, // = 現行 techtree.go 硬編值
		PlasmaCannonMaxDamage:        20,    // = 現行 session.go 硬編值
		BombardmentVolleys:           10,    // = 現行 orbital_bombardment.go 硬編值
		DefenderCommandoBonus:        2.5,   // 守方追平攻方 2.5x 加乘
		BombardmentBuildingBonusHits: 0,     // bug 已移除(尚未接線,見欄位註解)
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
