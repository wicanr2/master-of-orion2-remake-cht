package gamedata

// 殖民地建造佇列裡的「Special」型別一次性行動(手冊《The Big List》型別標記,GAME_MANUAL.pdf
// p.75-76)。與 buildings.go 的 Building(常駐建物,建完持續提供效果、每回合收維護費)不同——
// Special 套用後改變星球狀態或觸發一次性效果,不是可維護的建物,依 docs/tech/colony-buildings.md
// 的既有結論**刻意不計入**該檔「40 項建築」的統計(見該檔「不列入本表的型別」說明)。
//
// 本檔收錄地形改造(Terraforming)/蓋亞轉化(Gaia Transformation)/土壤改良(Soil Enrichment)/
// 運輸艦隊(Freighter Fleet,2026-07-11 補實作 #4)四項。前三項是 terraform.go 已移植好唯讀
// 規則、但先前零呼叫端的部分,套用後改變星球狀態(氣候/PopMax 等)。運輸艦隊性質不同——它不改
// 殖民地狀態,是帝國整體效果(engine.PlayerState.ActiveFreighters 增量 + 一次性現金加成,見
// applySpecialAction 的「運輸艦隊」case),但共用同一個「Special 一次性行動、可重複建造、不記入
// ColonyBuildings dedup」框架,故沿用同一個 SpecialAction 型別收錄,不另開新型別。Colony Base
// 是另一個 Special 行動,但它走專屬的起始殖民/新殖民地流程(見 docs/tech/homeworld-init.md),
// 不透過這裡的一般建造佇列,故不收錄於本表。
//
// 前置研究主題(PrereqTopic)來源:openorion2/src/tech.cpp 的 research_choices[] 是以
// ResearchTopic 的整數值直接當陣列索引(`research_choices[_topic]`,非依 techtree[field][level]
// 順序排列),故只需比對 gamestate.h 的 TOPIC_ 列舉宣告順序取得每個索引對應的主題名稱。已用此
// 方法交叉驗證 gamedata/techtree.go 既有資料與 buildings.go 三十餘項既有 PrereqTopic 逐一相符
// (見開發過程紀錄,一致率 100%),故本檔三項數值高信心:
//   - Soil Enrichment:research_choices[1]={400, {TECH_CLONING_CENTER, TECH_DEATH_SPORES,
//     TECH_SOIL_ENRICHMENT}} → TOPIC_ADVANCED_BIOLOGY(index 1),與 buildings.go「複製中心」
//     PrereqTopic 相同分組,交叉驗證一致。
//   - Terraforming:research_choices[35]={1150, {TECH_TERRAFORMING}} → TOPIC_GENETIC_MUTATIONS
//     (index 35)。與 terraform.go 檔頭「移植自...『Genetic Mutations』章節下的 Terraforming...
//     小節」的手冊出處完全吻合。
//   - Gaia Transformation:research_choices[70]={7500, {TECH_BIOMORPHIC_FUNGI,
//     TECH_EVOLUTIONARY_MUTATION, TECH_GAIA_TRANSFORMATION}} → TOPIC_TRANS_GENETICS(index 70)。
//
// 建造成本(PP)缺口:手冊完全沒給任何 Special 行動的 PP 數字(見 terraform.go 檔頭「建造成本
// 缺口」大段說明),本檔比照 buildings.go 對其餘 34 項建築既有的估計做法——用同一個 RP 研究成本
// 量級,參照 buildings.go 同一 RP 區間內其他建築已經標好的 EstimatedCost 估計值當基準,一律標
// EstimatedCost=true,不是手冊/資料檔實據:
//   - 土壤改良 RP400,同區間裝甲營房/戰機基地估計 PP150,本檔取同一值。
//   - 地形改造 RP1150,同區間行星重力產生器/行星輻射護盾估計 PP220-260,本檔取 260。
//   - 蓋亞轉化 RP7500,同區間阿提米絲系統網估計 PP900,本檔取同一值。
//
// 手冊原文明講地形改造「每次套用成本會提高」("each application has an increased production
// cost"),但未給任何遞增公式或起始值——本檔的 ProductionCost 是固定值,**不模擬遞增**,誠實
// TODO:待未來取得存檔/資料檔(.LBX)實際數字後補上遞增公式,目前每次套用都收同一個估計成本,
// 保守簡化,不臆造遞增曲線。
//
// 運輸艦隊(Freighter Fleet,2026-07-11 補實作 #4)的 PrereqTopic 出處:
// openorion2/src/tech.cpp research_choices[55]={50, {TECH_FREIGHTERS, TECH_NUCLEAR_BOMB,
// TECH_NUCLEAR_DRIVE}} → TOPIC_NUCLEAR_FISSION(index 55,同一索引法見上段說明)。建造成本(PP)
// 手冊同樣沒給任何數字,比照上面的估計慣例:TOPIC_NUCLEAR_FISSION 研究成本 RP50,與
// TOPIC_ENGINEERING(同為 RP50)的「海軍陸戰隊營」估計 PP60 同一量級,本檔取同一值 60。
type SpecialAction struct {
	NameZH string
	NameEN string

	PrereqTopic ResearchTopic

	ProductionCost int
	EstimatedCost  bool
}

const (
	// TerraformActionName 地形改造(Terraforming)在殖民地建造佇列的中文顯示名稱。
	TerraformActionName = "地形改造"
	// GaiaTransformationActionName 蓋亞轉化(Gaia Transformation)在殖民地建造佇列的中文顯示名稱。
	GaiaTransformationActionName = "蓋亞轉化"
	// SoilEnrichmentActionName 土壤改良(Soil Enrichment)在殖民地建造佇列的中文顯示名稱。
	SoilEnrichmentActionName = "土壤改良"
	// FreighterFleetActionName 運輸艦隊(Freighter Fleet)在殖民地建造佇列的中文顯示名稱
	// (2026-07-11 補實作 #4:運輸艦淨現金版本差異,見檔頭說明)。
	FreighterFleetActionName = "運輸艦隊"
)

// SpecialActions 是本 remake 目前接線的全部 Special 一次性行動(見檔頭說明,收錄 4 項)。
var SpecialActions = []SpecialAction{
	{
		NameZH: TerraformActionName, NameEN: "Terraforming",
		PrereqTopic:    TOPIC_GENETIC_MUTATIONS,
		ProductionCost: 260, EstimatedCost: true, // 見檔頭「建造成本缺口」說明,非手冊實據
	},
	{
		NameZH: GaiaTransformationActionName, NameEN: "Gaia Transformation",
		PrereqTopic:    TOPIC_TRANS_GENETICS,
		ProductionCost: 900, EstimatedCost: true, // 見檔頭「建造成本缺口」說明,非手冊實據
	},
	{
		NameZH: SoilEnrichmentActionName, NameEN: "Soil Enrichment",
		PrereqTopic:    TOPIC_ADVANCED_BIOLOGY,
		ProductionCost: 150, EstimatedCost: true, // 見檔頭「建造成本缺口」說明,非手冊實據
	},
	{
		NameZH: FreighterFleetActionName, NameEN: "Freighter Fleet",
		PrereqTopic:    TOPIC_NUCLEAR_FISSION,
		ProductionCost: 60, EstimatedCost: true, // 見檔頭「運輸艦隊」段說明,非手冊實據
	},
}

// SpecialActionByNameZH 依中文名找 Special 行動資料;找不到回傳 ok=false(供呼叫端判斷某個
// 建造佇列名稱是否屬於「Special 一次性行動」,而非常駐建築,見 shell.advanceBuilds)。
func SpecialActionByNameZH(zh string) (SpecialAction, bool) {
	for _, a := range SpecialActions {
		if a.NameZH == zh {
			return a, true
		}
	}
	return SpecialAction{}, false
}

// AvailableSpecialActions 回傳「玩家已研究前置科技」才會出現的 Special 行動清單,依
// SpecialActions 原順序——與 AvailableBuildings(buildings.go)同款 gate 慣例。
func AvailableSpecialActions(completedTopics map[ResearchTopic]bool) []SpecialAction {
	out := make([]SpecialAction, 0, len(SpecialActions))
	for _, a := range SpecialActions {
		if completedTopics != nil && completedTopics[a.PrereqTopic] {
			out = append(out, a)
		}
	}
	return out
}
