// 殖民地建築全表:手冊《The Big List》(GAME_MANUAL.pdf p.75-112)萃取的 35 建築 + 5 衛星
// (共 40 項),資料權威見 docs/tech/colony-buildings.md(已對 openorion2/src/tech.cpp 的前置
// 研究主題逐一交叉驗證)。
//
// 誠實聲明(見權威文件「五、待原版確認清單」):手冊只給維護費(BC/turn)與研究成本(RP),
// **建造成本(PP)幾乎沒有手冊來源**——唯一例外是 Armor Barracks(MANUAL_150.html modding
// 範例明載 150 PP)。本檔其餘各項 PP 成本一律是「依 RP 研究成本量級外推」的 remake 估計值,
// 用 EstimatedCost=true 標記,不可當作已驗證數值使用;待未來取得存檔/資料檔(.LBX)實際數字
// 再回填。維護費(MaintenanceBC)、前置研究(PrereqTopic)則是手冊/tech.cpp 逐項核對過的數值。
package gamedata

// BuildingCategory 是建築的功能分類(供建造選單分組顯示用,取手冊「分類」欄的主要類別;
// 部分建築手冊分類為複合,如「防禦/陸戰」,取第一個作為主分類)。
type BuildingCategory int

const (
	CategoryProduction  BuildingCategory = iota // 生產(工業產能)
	CategoryResearch                            // 科研
	CategoryFood                                // 食物
	CategoryHousing                             // 居住(人口上限)
	CategoryDefense                             // 防禦/陸戰/地形防禦
	CategoryTrade                               // 貿易(BC 收入)
	CategoryMorale                              // 士氣
	CategoryEnvironment                         // 環保(污染處理)
	CategorySociety                             // 社會(異族同化)
	CategoryMilitary                            // 軍事(艦艇訓練)
	CategorySatellite                           // 軌道衛星(Star Base 系列 + Artemis + Dimensional Portal)
)

// Building 是一棟可建造的殖民地建築(含軌道衛星)的結構化資料。
type Building struct {
	NameZH string // 建議中譯(既有 5 棟沿用 shipped 既有字串,避免破壞既有存檔/效果比對)
	NameEN string // 手冊原文英文名

	Category BuildingCategory

	// MaintenanceBC 是手冊給的維護費(BC/turn)。手冊 40 項全數有給,無 -1 情形。
	MaintenanceBC int

	// ProductionCost 是建造成本(PP)。手冊幾乎沒給(見檔頭誠實聲明);EstimatedCost=true
	// 者一律為 remake 外推估計值,不是手冊/資料檔實據。
	ProductionCost int
	EstimatedCost  bool

	// PrereqTopic 是前置研究主題(ResearchTopic),對照 docs/tech/colony-buildings.md 與
	// openorion2/src/tech.cpp 的 techtree 欄位名稱,已逐一核對相符。
	PrereqTopic ResearchTopic

	// Effect 是手冊效果敘述(中文摘要,供 UI 顯示/未來數值建模參考;不是程式碼可直接消費的
	// 結構化效果——目前只有部分建築已在 internal/shell/session.go 的 applyBuildingEffect
	// 建模,其餘標 TODO 未建模)。
	Effect string
}

// Buildings 是手冊全表 40 項(35 建築 + 5 衛星),依 docs/tech/colony-buildings.md 表列順序。
// 注意:Stellar Converter(行星版,Building/System 混合型別)依權威文件「不計入 40 項」,本表
// 不收錄。
var Buildings = []Building{
	// ---- 二、殖民地建築(35 項) ----
	{
		NameZH: "海軍陸戰隊營", NameEN: "Marine Barracks",
		Category: CategoryDefense, MaintenanceBC: 1,
		ProductionCost: 60, EstimatedCost: true, // 既有 shipped 值(session.go 原 buildOptions),非手冊實據
		PrereqTopic: TOPIC_ENGINEERING,
		Effect:      "建成立即產生最多 4 個陸戰隊單位;之後每 5 回合 +1,上限為「現有人口/2」與「星球人口上限/2」取較小值;特定政府下可消除士氣懲罰",
	},
	{
		NameZH: "自動工廠", NameEN: "Automated Factories",
		Category: CategoryProduction, MaintenanceBC: 1,
		ProductionCost: 60, EstimatedCost: true, // 既有 shipped 值,非手冊實據
		PrereqTopic: TOPIC_ADVANCED_CONSTRUCTION,
		Effect:      "每個工業人口 +1 產能/回合,殖民地整體 +5 產能",
	},
	{
		NameZH: "飛彈基地", NameEN: "Missile Base",
		Category: CategoryDefense, MaintenanceBC: 2,
		ProductionCost: 90, EstimatedCost: true, // 估計:依 RP 150 量級外推,無手冊實據
		PrereqTopic: TOPIC_ADVANCED_CONSTRUCTION,
		Effect:      "配備最佳飛彈(佔用 300 空間內盡量多),自動防禦來襲艦隊;只能被軌道轟炸摧毀",
	},
	{
		NameZH: "裝甲營房", NameEN: "Armor Barracks",
		Category: CategoryDefense, MaintenanceBC: 2,
		ProductionCost: 150, EstimatedCost: false, // 手冊實據:MANUAL_150.html modding 範例
		PrereqTopic: TOPIC_ASTRO_ENGINEERING,
		Effect:      "建成立即產生最多 2 個裝甲營,之後每 5 回合 +1,上限為「現有人口/4」與「星球人口上限/4」取較小值;特定政府下可消除士氣懲罰",
	},
	{
		NameZH: "太空港", NameEN: "Spaceport",
		Category: CategoryTrade, MaintenanceBC: 1,
		ProductionCost: 100, EstimatedCost: true, // 既有 shipped 值,非手冊實據
		PrereqTopic: TOPIC_ASTRO_ENGINEERING,
		Effect:      "該殖民地所有來源的 BC 收入 +50%",
	},
	{
		NameZH: "戰機基地", NameEN: "Fighter Garrison",
		Category: CategoryDefense, MaintenanceBC: 2,
		ProductionCost: 150, EstimatedCost: true, // 估計:依 RP 400 量級外推(與 Armor Barracks 同組)
		PrereqTopic: TOPIC_ASTRO_ENGINEERING,
		Effect:      "依已解鎖的最高階戰機科技,可駐留 10 個攔截機中隊 / 6 個轟炸機中隊 / 4 個重型戰機中隊;每 10 回合全數整補;只能被軌道轟炸摧毀",
	},
	{
		NameZH: "機器人採礦廠", NameEN: "Robo Mining Plant",
		Category: CategoryProduction, MaintenanceBC: 2,
		ProductionCost: 200, EstimatedCost: true, // 估計:依 RP 650 量級外推
		PrereqTopic: TOPIC_ROBOTICS,
		Effect:      "每個工業人口 +2 產能,殖民地整體 +10 產能",
	},
	{
		NameZH: "地面砲台", NameEN: "Ground Batteries",
		Category: CategoryDefense, MaintenanceBC: 2,
		ProductionCost: 260, EstimatedCost: true, // 估計:依 RP 1150 量級外推
		PrereqTopic: TOPIC_ASTRO_CONSTRUCTION,
		Effect:      "配備最佳光束武器的 Heavy Mount 與 Point Defense 版本(佔用 450 空間內盡量多);只能被軌道轟炸摧毀",
	},
	{
		NameZH: "再生反應爐", NameEN: "Recyclotron",
		Category: CategoryProduction, MaintenanceBC: 3,
		ProductionCost: 320, EstimatedCost: true, // 估計:依 RP 1500 量級外推
		PrereqTopic: TOPIC_ADVANCED_MANUFACTURING,
		Effect:      "每單位人口(不論職業)額外產生 1 產能,且此產能不計入污染;1.50i 起可將 Toxic 星球轉為 Barren",
	},
	{
		NameZH: "機器人工廠", NameEN: "Robotic Factory",
		Category: CategoryProduction, MaintenanceBC: 3,
		ProductionCost: 380, EstimatedCost: true, // 估計:依 RP 2000 量級外推
		PrereqTopic: TOPIC_ADVANCED_ROBOTICS,
		Effect:      "依礦產豐度加成:Ultra Poor +5、Poor +8、Abundant +10、Rich +15、Ultra Rich +20",
	},
	{
		NameZH: "深層核心礦場", NameEN: "Deep Core Mine",
		Category: CategoryProduction, MaintenanceBC: 3,
		ProductionCost: 550, EstimatedCost: true, // 估計:依 RP 3500 量級外推
		PrereqTopic: TOPIC_TECTONIC_ENGINEERING,
		Effect:      "每個工人 +3 產能,殖民地整體 +15 產能",
	},
	{
		NameZH: "核心廢料場", NameEN: "Core Waste Dumps",
		Category: CategoryEnvironment, MaintenanceBC: 8,
		ProductionCost: 550, EstimatedCost: true, // 估計:依 RP 3500 量級外推
		PrereqTopic: TOPIC_TECTONIC_ENGINEERING,
		Effect:      "完全消除星球污染;建成時取代 Pollution Processor 與 Atmospheric Renewer(以全額建造成本回收出售)",
	},
	{
		NameZH: "食物複製機", NameEN: "Food Replicators",
		Category: CategoryFood, MaintenanceBC: 10,
		ProductionCost: 460, EstimatedCost: true, // 估計:依 RP 2750 量級外推
		PrereqTopic: TOPIC_MATTER_ENERGY_CONVERSION,
		Effect:      "可依需求將工業產能以 2:1 轉換成食物,每單位食物花費 1 BC",
	},
	{
		NameZH: "污染處理器", NameEN: "Pollution Processor",
		Category: CategoryEnvironment, MaintenanceBC: 1,
		ProductionCost: 200, EstimatedCost: true, // 估計:依 RP 650 量級外推
		PrereqTopic: TOPIC_ADVANCED_CHEMISTRY,
		Effect:      "可處理殖民地一半產能對應的污染,按比例降低污染量;與 Atmospheric Renewer 疊加(合計 1/8 產能致污染);被 Core Waste Dump 取代",
	},
	{
		NameZH: "大氣更新器", NameEN: "Atmospheric Renewer",
		Category: CategoryEnvironment, MaintenanceBC: 3,
		ProductionCost: 260, EstimatedCost: true, // 估計:依 RP 1150 量級外推
		PrereqTopic: TOPIC_MOLECULAR_COMPRESSION,
		Effect:      "消除四分之三工業產能造成的污染;與 Pollution Processor 同時存在時合計只剩 1/8 產能致污染;被 Core Waste Dump 取代",
	},
	{
		NameZH: "太空學院", NameEN: "Space Academy",
		Category: CategoryMilitary, MaintenanceBC: 2,
		ProductionCost: 80, EstimatedCost: true, // 估計:依 RP 150 量級外推
		PrereqTopic: TOPIC_MILITARY_TACTICS,
		Effect:      "此殖民地建造的艦艇船員起始等級 +1 級;同星系內每有一座 Space Academy,所有駐留艦艇船員每回合額外 +1 經驗",
	},
	{
		NameZH: "異族管理中心", NameEN: "Alien Management Center",
		Category: CategorySociety, MaintenanceBC: 1,
		ProductionCost: 200, EstimatedCost: true, // 估計:依 RP 650 量級外推
		PrereqTopic: TOPIC_XENO_RELATIONS,
		Effect:      "每 2 回合同化 1 單位被征服人口(不論政府);消除多種族殖民地 -20% 士氣懲罰,並使未同化人口叛亂機率減半",
	},
	{
		NameZH: "行星證券交易所", NameEN: "Planetary Stock Exchange",
		Category: CategoryTrade, MaintenanceBC: 2,
		ProductionCost: 260, EstimatedCost: true, // 估計:依 RP 1150 量級外推
		PrereqTopic: TOPIC_MACRO_ECONOMICS,
		Effect:      "該殖民地收入 +100%",
	},
	{
		NameZH: "太空大學", NameEN: "Astro University",
		Category: CategoryResearch, MaintenanceBC: 4,
		ProductionCost: 380, EstimatedCost: true, // 估計:依 RP 2000 量級外推
		PrereqTopic: TOPIC_TEACHING_METHODS,
		Effect:      "每單位受教育人口(農/工/科)額外 +1 對應產出(食物/產能/研究皆適用)",
	},
	{
		NameZH: "研究實驗室", NameEN: "Research Laboratory",
		Category: CategoryResearch, MaintenanceBC: 1,
		ProductionCost: 60, EstimatedCost: true, // 既有 shipped 值,非手冊實據
		PrereqTopic: TOPIC_OPTRONICS,
		Effect:      "每個科學家人口 +1 研究點;另自動產生 5 研究點",
	},
	{
		NameZH: "行星超級電腦", NameEN: "Planetary Supercomputer",
		Category: CategoryResearch, MaintenanceBC: 2,
		ProductionCost: 220, EstimatedCost: true, // 估計:依 RP 900 量級外推
		PrereqTopic: TOPIC_POSITRONICS,
		Effect:      "每個科學家人口 +2 研究點,殖民地整體 +10 研究點",
	},
	{
		NameZH: "全息模擬艙", NameEN: "Holo Simulator",
		Category: CategoryMorale, MaintenanceBC: 1,
		ProductionCost: 220, EstimatedCost: true, // 估計:依 RP 900 量級外推
		PrereqTopic: TOPIC_POSITRONICS,
		Effect:      "殖民地士氣 +20%",
	},
	{
		NameZH: "自動實驗室", NameEN: "Autolab",
		Category: CategoryResearch, MaintenanceBC: 3,
		ProductionCost: 460, EstimatedCost: true, // 估計:依 RP 2750 量級外推
		PrereqTopic: TOPIC_CYBERTRONICS,
		Effect:      "全自動產生 30 研究點/回合(不依賴人口)",
	},
	{
		NameZH: "銀河網路中心", NameEN: "Galactic Cybernet",
		Category: CategoryResearch, MaintenanceBC: 3,
		ProductionCost: 650, EstimatedCost: true, // 估計:依 RP 4500 量級外推
		PrereqTopic: TOPIC_GALACTIC_NETWORKING,
		Effect:      "每個科學家人口 +3 研究點,殖民地整體 +15 研究點",
	},
	{
		NameZH: "歡樂穹頂", NameEN: "Pleasure Dome",
		Category: CategoryMorale, MaintenanceBC: 3,
		ProductionCost: 800, EstimatedCost: true, // 估計:依 RP 6000 量級外推
		PrereqTopic: TOPIC_MOLECULATRONICS,
		Effect:      "殖民地士氣 +30%",
	},
	{
		NameZH: "水耕農場", NameEN: "Hydroponic Farm",
		Category: CategoryFood, MaintenanceBC: 2,
		ProductionCost: 50, EstimatedCost: true, // 估計:依 RP 80 量級外推
		PrereqTopic: TOPIC_ASTRO_BIOLOGY,
		Effect:      "殖民地食物產出 +2",
	},
	{
		NameZH: "生態圈", NameEN: "Biospheres",
		Category: CategoryHousing, MaintenanceBC: 1,
		ProductionCost: 60, EstimatedCost: true, // 社群來源 60 PP(低可信度),仍標估計
		PrereqTopic: TOPIC_ASTRO_BIOLOGY,
		Effect:      "星球人口上限 +2 單位",
	},
	{
		NameZH: "複製中心", NameEN: "Cloning Center",
		Category: CategoryFood, MaintenanceBC: 2,
		ProductionCost: 100, EstimatedCost: true, // 社群來源 100 PP(低可信度),仍標估計
		PrereqTopic: TOPIC_ADVANCED_BIOLOGY,
		Effect:      "人口成長 +100,000(0.1 單位)/回合,直到達星球人口上限為止",
	},
	{
		NameZH: "地底農場", NameEN: "Subterranean Farms",
		Category: CategoryFood, MaintenanceBC: 4,
		ProductionCost: 320, EstimatedCost: true, // 估計:依 RP 1500 量級外推
		PrereqTopic: TOPIC_MACRO_GENETICS,
		Effect:      "星球食物產出 +4",
	},
	{
		NameZH: "氣候控制器", NameEN: "Weather Controller",
		Category: CategoryFood, MaintenanceBC: 3,
		ProductionCost: 320, EstimatedCost: true, // 估計:依 RP 1500 量級外推
		PrereqTopic: TOPIC_MACRO_GENETICS,
		Effect:      "每個農業人口食物產出 +2",
	},
	{
		NameZH: "行星重力產生器", NameEN: "Planetary Gravity Generator",
		Category: CategoryHousing, MaintenanceBC: 2,
		ProductionCost: 260, EstimatedCost: true, // 估計:依 RP 1150 量級外推
		PrereqTopic: TOPIC_ARTIFICIAL_GRAVITY,
		Effect:      "將星球重力正常化至 Normal-G,消除 Low-G/Heavy-G 的負面效果",
	},
	{
		NameZH: "行星輻射護盾", NameEN: "Planetary Radiation Shield",
		Category: CategoryDefense, MaintenanceBC: 1,
		ProductionCost: 220, EstimatedCost: true, // 估計:依 RP 900 量級外推
		PrereqTopic: TOPIC_MAGNETO_GRAVITICS,
		Effect:      "Radiated 氣候星球維持 Barren 狀態;軌道轟炸傷害 -5;被 Planetary Flux Shield 取代",
	},
	{
		NameZH: "曲速力場干擾器", NameEN: "Warp Field Interdictor",
		Category: CategoryDefense, MaintenanceBC: 3,
		ProductionCost: 380, EstimatedCost: true, // 估計:依 RP 2000 量級外推
		PrereqTopic: TOPIC_WARP_FIELDS,
		Effect:      "星系內半徑 3 秒差距範圍,使敵方艦艇移動速度降為 1 秒差距/回合",
	},
	{
		NameZH: "行星通量護盾", NameEN: "Planetary Flux Shield",
		Category: CategoryDefense, MaintenanceBC: 3,
		ProductionCost: 650, EstimatedCost: true, // 估計:依 RP 4500 量級外推
		PrereqTopic: TOPIC_QUANTUM_FIELDS,
		Effect:      "Radiated 氣候轉 Barren;軌道轟炸傷害 -10;取代已存在的 Planetary Radiation Shield;被 Planetary Barrier Shield 取代",
	},
	{
		NameZH: "行星屏障護盾", NameEN: "Planetary Barrier Shield",
		Category: CategoryDefense, MaintenanceBC: 5,
		ProductionCost: 1200, EstimatedCost: true, // 估計:依 RP 15000 量級外推
		PrereqTopic: TOPIC_TEMPORAL_FIELDS,
		Effect:      "Radiated 氣候轉 Barren;軌道轟炸傷害 -20;生物武器無法進入大氣層;取代 Planetary Flux Shield",
	},

	// ---- 三、軌道衛星(5 項) ----
	{
		NameZH: "星基", NameEN: "Star Base",
		Category: CategorySatellite, MaintenanceBC: 2,
		ProductionCost: 300, EstimatedCost: true, // 既有 shipped 值,非手冊實據
		PrereqTopic: TOPIC_ENGINEERING,
		Effect:      "配備最新式武器的軌道平台;星球掃描範圍 +2 秒差距;沒有 Star Base 的星球無法建造超過驅逐艦等級的船艦;+1 指揮評等",
	},
	{
		NameZH: "戰鬥站", NameEN: "Battlestation",
		Category: CategorySatellite, MaintenanceBC: 3,
		ProductionCost: 200, EstimatedCost: true, // 估計:依 RP 650 量級外推
		PrereqTopic: TOPIC_ROBOTICS,
		Effect:      "比 Star Base 火力更強;掃描範圍 +4 秒差距;為己方艦隊 +10 光束攻擊;+2 指揮評等;取代同軌道的 Star Base",
	},
	{
		NameZH: "星辰要塞", NameEN: "Star Fortress",
		Category: CategorySatellite, MaintenanceBC: 4,
		ProductionCost: 800, EstimatedCost: true, // 估計:依 RP 6000 量級外推
		PrereqTopic: TOPIC_SUPERSCALAR_CONSTRUCTION,
		Effect:      "比 Battlestation 更強;掃描範圍 +6 秒差距;為己方艦隊 +20 光束攻擊;+3 指揮評等;取代同軌道的 Battlestation 或 Star Base",
	},
	{
		NameZH: "阿提米絲系統網", NameEN: "Artemis System Net",
		Category: CategorySatellite, MaintenanceBC: 5,
		ProductionCost: 900, EstimatedCost: true, // 估計:依 RP 7500 量級外推
		PrereqTopic: TOPIC_PLANETOID_CONSTRUCTION,
		Effect:      "環繞整個星系的巨型水雷網;敵艦進入時依船體等級觸發機率(Frigate 20%~Doom Star 100%);每次觸發 8-28 枚水雷命中,每枚造成 20 點傷害減去目標護盾等級",
	},
	{
		NameZH: "次元傳送門", NameEN: "Dimensional Portal",
		Category: CategorySatellite, MaintenanceBC: 2,
		ProductionCost: 650, EstimatedCost: true, // 估計:依 RP 4500 量級外推
		PrereqTopic: TOPIC_MULTIDIMENSIONAL_PHYSICS,
		Effect:      "同系統內的艦隊可跨越次元,對安塔蘭人發動攻擊(終局戰觸發點)",
	},
}

// BuildingByNameZH 依中文名找建築資料。
func BuildingByNameZH(zh string) (Building, bool) {
	for _, b := range Buildings {
		if b.NameZH == zh {
			return b, true
		}
	}
	return Building{}, false
}

// BuiltMaintenanceBC 加總一組「已建成」建築(以中文名為 key,值為 true 代表已建成)的每回合
// 維護費(BC)。每項 MaintenanceBC 皆為手冊實據(見檔頭誠實聲明,ProductionCost 才有
// EstimatedCost 的區分,MaintenanceBC 40 項全數手冊有給,無估計值)。
//
// 用途:取代 remake 先前「每回合固定 Maintenance=5」的無據 placeholder——維護費現在由
// 玩家(或 AI)實際已建成的建築清單加總而來,而非常數。找不到對應建築資料(理論上不會
// 發生,呼叫端的 built 只會記錄本檔 Buildings 表中存在的中文名)的 key 直接跳過不計,
// 不臆造數字。
func BuiltMaintenanceBC(built map[string]bool) int {
	total := 0
	for name, ok := range built {
		if !ok {
			continue
		}
		if b, found := BuildingByNameZH(name); found {
			total += b.MaintenanceBC
		}
	}
	return total
}

// CommandPointsFromBuildings 回傳單一殖民地「軌道衛星」建築提供的指揮評等(Command Rating)供給。
//
// 手冊原文逐項出處(GAME_MANUAL.pdf):
//   - p.79「A Star Base requires 2 BC per turn to maintain and adds 1 to your Command Rating.」
//   - p.82「A Battlestation costs 3 BC in maintenance each turn and adds 2 to your Command
//     Rating.... It replaces any Star Base in orbit around the same planet.」
//   - p.83「A Star Fortress costs 4 BC in maintenance each turn and adds 3 to your Command
//     Rating.... The fortress replaces any Battlestation or Star Base in orbit around the same
//     planet.」
//
// 三者是「取代關係」(replaces),不是可疊加的獨立加成——同一殖民地最多只有一座軌道衛星在軌
// (完工時舊的一併出售,見 buildings.go 三項 Effect 欄「取代」敘述)。本函式故意不用加總,
// 而是依「built 中最高階者」直接回傳其固定值,天然滿足互斥語意,即使呼叫端的 built map 因為
// 某種原因(如存檔相容或未來 UI)同時記錄了兩者也不會被誤算成疊加。
//
// built 為 nil(尚無任何建築)回傳 0,非漏算。
func CommandPointsFromBuildings(built map[string]bool) int {
	switch {
	case built["星辰要塞"]: // Star Fortress,取代 Battlestation/Star Base
		return 3
	case built["戰鬥站"]: // Battlestation,取代 Star Base
		return 2
	case built["星基"]: // Star Base
		return 1
	}
	return 0
}

// AvailableBuildings 回傳「前置研究已完成」的建築清單,依 Buildings 原順序。
// completedTopics 為 nil 時視為尚無任何研究完成(只回傳前置為 TOPIC_STARTING_TECH 的項目——
// 目前表中無此類建築,故回傳空清單)。
func AvailableBuildings(completedTopics map[ResearchTopic]bool) []Building {
	out := make([]Building, 0, len(Buildings))
	for _, b := range Buildings {
		if completedTopics != nil && completedTopics[b.PrereqTopic] {
			out = append(out, b)
		}
	}
	return out
}
