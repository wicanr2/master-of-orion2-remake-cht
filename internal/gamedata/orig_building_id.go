package gamedata

// orig_building_id.go:remake 建築名 → **原版內部建築編號**的對照表。
//
// 2026-08-07 從 `cmd/moo2/colonysurface.go` 搬過來:原本只有畫地表 sprite 用得到,
// 現在 `initial_buildings.go` 的優先清單也要用它把原版編號換回建築。
// **兩份表要靠同一個對照連起來**,各自抄一份遲早會漂開。
// cmd/moo2 那邊改成引用這裡,既有的 9 條回歸測試照樣跑。

// OrigBuildingID 是 remake 的建築英文名(`Building.NameEN`,取自手冊 The Big List)
// → **原版內部建築編號**(1-based,見檔頭的兩個獨立來源)。
//
// 為什麼需要一張手寫對照:手冊的用字和遊戲內部字串不一樣——手冊寫 "Automated Factories"、
// 遊戲內部是 "Automated Factory";手冊 "Alien Management Center"、內部 "Alien Control Center";
// 手冊 "Planetary Stock Exchange"、內部 "Stock Exchange"。照字串比對會漏一半。
// 右欄註記的是原版 TECHNAME.LBX 裡的原文,方便日後核對。
var OrigBuildingID = map[string]int{
	"Alien Management Center":     1,  // Alien Control Center
	"Armor Barracks":              2,  // Armor Barracks
	"Artemis System Net":          3,  // Artemis System Net
	"Astro University":            4,  // Astro University
	"Atmospheric Renewer":         5,  // Atmosphere Renewer
	"Autolab":                     6,  // Autolab
	"Automated Factories":         7,  // Automated Factory
	"Battlestation":               8,  // Battlestation
	"Cloning Center":              10, // Cloning Center
	"Deep Core Mine":              12, // Deep Core Mine
	"Core Waste Dumps":            13, // Core Waste Dump
	"Dimensional Portal":          14, // Dimensional Portal
	"Biospheres":                  15, // Biospheres
	"Food Replicators":            16, // Food Replicators
	"Gaia Transformation":         17, // Gaia Transformation (Special)
	"Galactic Cybernet":           19, // Galactic Cybernet
	"Holo Simulator":              20, // Holo Simulator
	"Hydroponic Farm":             21, // Hydroponic Farm
	"Marine Barracks":             22, // Marine Barracks
	"Planetary Barrier Shield":    23, // Barrier Shield
	"Planetary Flux Shield":       24, // Flux Shield
	"Planetary Gravity Generator": 25, // Gravity Generator
	"Missile Base":                26, // Missile Base
	"Ground Batteries":            27, // Ground Batteries
	"Planetary Radiation Shield":  28, // Radiation Shield
	"Planetary Stock Exchange":    29, // Stock Exchange
	"Planetary Supercomputer":     30, // Supercomputer
	"Pleasure Dome":               31, // Pleasure Dome
	"Pollution Processor":         32, // Pollution Processor
	"Recyclotron":                 33, // Recyclotron
	"Robotic Factory":             34, // Robotic Factory
	"Research Laboratory":         35, // Research Lab
	"Robo Mining Plant":           36, // Robo Miner Plant
	"Space Academy":               38, // Space Academy
	"Spaceport":                   39, // Spaceport
	"Star Base":                   40, // Star Base
	"Star Fortress":               41, // Star Fortress
	"Stellar Converter":           42, // Stellar Converter(行星版,分類 0 → 畫在地表格點上)
	"Subterranean Farms":          43, // Subterranean Farms
	"Terraforming":                44, // Terraforming (Special)
	"Warp Field Interdictor":      45, // Warp Interdictor
	"Weather Controller":          46, // Weather Controller
	"Fighter Garrison":            47, // Fighter Garrison
}

// OriginalBuildingIDForName 將 remake 常駐建築／Special 使用的中文名，或既有英文手冊名，
// 對回原版 raw building ID。無法映射時回傳 ok=false，不以表位置或字串相似度猜測。
func OriginalBuildingIDForName(name string) (id int, ok bool) {
	if id, ok = OrigBuildingID[name]; ok {
		return id, true
	}
	building, ok := BuildingByNameZH(name)
	if ok {
		id, ok = OrigBuildingID[building.NameEN]
		return id, ok
	}
	action, ok := SpecialActionByNameZH(name)
	if ok {
		id, ok = OrigBuildingID[action.NameEN]
		return id, ok
	}
	return 0, false
}
