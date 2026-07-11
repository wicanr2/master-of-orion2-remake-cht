// Package gamedata:版本規則 profile(patch 1.3 vs patch 1.5)。
//
// 依 docs/tech/version-1.3-1.5-diff.md §5 設計稿:逐條核對 CHANGELOG_150.TXT(1.50.0–1.50.26
// 全部版本)+ MANUAL_150.html 後落在本專案「已實作系統」且兩版數字真的不同的項目逐輪確證增補
// (最初 3 條,見該文件 §2)。新差異確證後才加欄位,見該文件 §5.3 擴充路徑——⚠ 本段落只描述
// 「怎麼決定加不加欄位」這個方法,實際目前有幾個欄位以下方 RuleProfile struct 定義為準
// (dated 盤點數字會過期,見 rulebook 63)。
//
// 2026-07-11 補實作 §1 全量表 #5/#7(地面戰/軌道轟炸的真正版本差異項)新增 2 個欄位——
// DefenderCommandoBonus(#5)、BombardmentBuildingBonusHits(#7)。同一批 #6/#8/#9/#11 已交叉
// 核對確認為「非差異項」(兩版預設相同),故不進本檔,只在 ground_version_diff.go 實作一次,
// 詳見 docs/tech/ground-combat-algorithm.md「2026-07-11 版本差異補實作」節。
//
// 2026-07-11 再補實作 §1 全量表 #4(運輸艦淨現金版本差異)新增 FreightersCashBonus 欄位——
// 出處見該欄位註解。
package gamedata

// GameVersion 對應 CLAUDE.md「主選單選擇 1.3 或 1.5」的兩個選項。
type GameVersion int

const (
	VersionClassic13   GameVersion = iota // 官方最後正式 patch 1.31
	VersionCommunity15                    // 社群非官方 patch 1.50(本專案現行資料的預設來源)
)

// RuleProfile 收斂「已確證、且落在已實作系統上」的版本相依數值。
// ⚠ 刻意保持精簡——只收已逐條確證、且落在本專案已實作系統上的差異項,不要為了「看起來完整」
// 塞進未確證或未接線的欄位。新差異確證後才加欄位,見 docs/tech/version-1.3-1.5-diff.md §5.3 擴充路徑。
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
	// 現況(2026-07-11 已接線):AIOpponent 新增 ColonyBuildings 欄位後,
	// internal/shell/orbital_bombardment.go BombardColony 已把本欄位讀進「每棟建築消耗的
	// hits = GroundPlanetHitsPerBuilding + BombardmentBuildingBonusHits」,實際影響轟炸摧毀
	// 幾棟建築(1.3 每棟建築多 +1 hit 才摧毀、1.5 不加)。⚠ CHANGELOG 原句語意本身模糊(是建築
	// 多吸一擊、還是建築多受一擊才被摧毀),本 remake 採「每棟建築在 1.3 需多 +1 hit 才摧毀」的
	// 保守解讀,見 BombardColony 檔頭「建築吸收」段落的誠實標註,非手冊逐字驗證值。
	BombardmentBuildingBonusHits int

	// 軌道防禦(#14,2026-07-11 補實作):衛星光束武器的 arc-cost 百分比,套進
	// gamedata.SatelliteBeamSpaceWithArc,由 internal/shell 的 retaliationAttackers 用來把
	// 「星基/戰鬥站/星辰要塞」的 space 預算換算成能塞入的光束武器把數。來源:
	// CHANGELOG_150.TXT 1.50.7「Beam weapons on satellites... arc cost»、1.50.10「衛星
	// arc cost 由 +40% 修正為 +33.3%」——1.3 維持較舊的 +25%,1.5 最終值取整數 33
	// (33.3% 四捨五入)。
	SatelliteBeamArcCostPct int

	// 軌道防禦(#14):地面砲台(Ground Batteries,手冊 p.81「as many as fit in 450
	// space」)光束武器的 arc-cost 百分比。來源同上,CHANGELOG_150.TXT 1.50.7:1.3 地面砲台
	// arc-cost 為 +0%(無懲罰),1.5 統一改為 +50%。
	//
	// ⚠ 現況(2026-07-11):AI 開局 homeworldBuildings() 沒有「地面砲台」這項建築(見
	// session.go homeworldBuildings 註解),本欄位目前在 NewDemoSession 的自然對局流程走不到
	// ——只有 retaliationAttackers 對「地面砲台」存在時才會讀取,供未來 AI/玩家真的建出地面
	// 砲台後使用,不是遺漏,是誠實標註「已備妥但暫無呼叫端會觸發」。
	GroundBatteryBeamArcCostPct int

	// 經濟(#4,2026-07-11 補實作):建造「運輸艦隊」(Freighter Fleet)完工當下,套用進國庫的
	// 一次性現金加成(BC),對應手冊 config 參數 freighters_cash_bonus。來源:
	// MANUAL_150.html「Buildings & Freighters Free Cash Bug」表——1.31 版對運輸艦隊收取的實際
	// 維護費是 0-3 BC(依當時是否缺糧,0.5 BC/艘無條件捨去),但固定回饋(補償)5 BC,故淨得
	// 恆為 2-5 BC(手冊原文承認這是 1.31 就有的「一律有淨利」quirk,非 1.40 才出現);
	// CHANGELOG_150.TXT 1.50.8「Changed freighters_cash_bonus default from 5 to 0 BC so building
	// a freighter fleet no longer generates free cash.」——1.50 出廠預設把這個固定回饋改成 0。
	// 本 remake 簡化為只模擬「固定回饋」這一側(不模擬 0-3 BC 的建造當下維護費立即扣款,見
	// session.go applySpecialAction 對「運輸艦隊」case 的實作註解),故 1.3 設 5(對應手冊
	// 「固定回饋 5 BC」)、1.5 設 0(對應 1.50.8 起的預設值),兩版皆非官方逐 BC 精確淨額,是
	// 「哪個方向、量級對」的版本差異模擬,不是逐分錢重現。
	FreightersCashBonus int
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
		SatelliteBeamArcCostPct:      25,  // 衛星 arc-cost +25%(CHANGELOG_150.TXT 1.50.7 修正前的舊值)
		GroundBatteryBeamArcCostPct:  0,   // 地面砲台無 arc-cost 懲罰
		FreightersCashBonus:          5,   // 運輸艦隊完工固定回饋 5 BC(MANUAL_150.html Free Cash Bug 表)
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
		SatelliteBeamArcCostPct:      33,    // 衛星 arc-cost 最終值 +33.3%(1.50.10),取整數
		GroundBatteryBeamArcCostPct:  50,    // 地面砲台 arc-cost +50%(1.50.7)
		FreightersCashBonus:          0,     // freighters_cash_bonus 出廠預設(CHANGELOG_150.TXT 1.50.8)
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
// 現況(2026-07-11 已接線):engine.PlayerState 新增 HyperAdvancedResearchCost 欄位,
// engine.RunResearchPhase 對 Hyper 主題套用該覆寫值(見 internal/engine/research.go);
// shell.EndTurn 在呼叫 RunEmpireTurn 前,對玩家與每個 AI 對手都執行
// `Player.HyperAdvancedResearchCost = gamedata.HyperAdvancedCost(s.RuleProfile)` 注入這局的
// 版本規則值(見 internal/shell/session.go EndTurn),避免玩家用 1.3、AI 仍吃 1.5 硬編值的
// 規則不對稱。顯示層見 internal/shell/research.go GameSession.ResearchCostForDisplay。
func HyperAdvancedCost(p RuleProfile) int {
	return p.HyperAdvancedLevel1Cost
}
