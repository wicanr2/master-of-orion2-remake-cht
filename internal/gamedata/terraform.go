package gamedata

// 地形改造(Terraforming)/蓋亞轉化(Gaia Transformation)/土壤改良(Soil Enrichment)相關唯讀規則,
// 移植自 GAME_MANUAL.pdf(moo2_patch1.5 隨附的完整遊戲手冊)「Genetic Mutations」章節下的
// Terraforming 與 Gaia Transformation 小節(約 p.99-101),以及「Macro Genetics」章節前的
// Soil Enrichment 小節(約 p.99)。
//
// moo2_patch1.5/MANUAL_150.html(1.50 patch 說明書)補充兩處:
//  1. 「Toxic Planet Terraforming」說明:1.50i 用 Recyclotron 把 Toxic 星球轉成 Barren(見
//     TerraformToxicNextClimate),並明講一般的 Terraforming/Gaia Transformation 兩個特殊建設
//     都不能建在 Toxic 星球上。
//  2. Modding 附錄的 pop_climate 參數(見 TerraformClimatePopFactor),量化地形改造沿氣候鏈
//     推進到 Terran/Gaia 能換來多少人口容量成長。
//
// 建造成本(production cost)缺口:兩份文件都**沒有**列出地形改造/蓋亞轉化/土壤改良本身的
// 產能建造成本數字。GAME_MANUAL.pdf 原文只說 Terraforming "each application has an increased
// production cost"(每次套用成本會提高),未給任何公式、係數或起始值;MANUAL_150.html 的
// Buildings 小節只舉了 Armor Barracks(150 PP)當 modding 範例,未提到 Terraforming/Gaia
// Transformation/Soil Enrichment 這三個「Special」類型的建造成本。因此本檔案不移植、也不猜測
// 任何建造成本公式 —— TODO 待查證:地形改造逐次遞增的產能成本數字/公式,需要從遊戲資料檔
// (非手冊文字)取得。
//
// 命名前綴:Terraform = 地形改造,Gaia = 蓋亞轉化。
//
// TerraformClimatePopFactor 的順序額外說明:MANUAL_150.html 只給數值陣列本身
// ("pop_climate = 25 25 25 25 25 25 40 60 80 100"),沒有在文字中標註各欄位對應的氣候名稱
// (該手冊註明對照表只在讀不到的 Excel 手冊截圖裡)。本檔案採用專案既有 enums.go 的
// PlanetClimate 0-based 順序(TOXIC,RADIATED,BARREN,DESERT,TUNDRA,OCEAN,SWAMP,ARID,TERRAN,
// GAIA),並以 openorion2/src/gamestate.cpp 的 climatePopFactors 陣列(逐項附氣候名稱註解,
// 數值與順序跟 MANUAL_150.html 完全一致)交叉驗證,不是憑空排序。

const (
	// TerraformSoilEnrichmentFoodBonusPerFarmer 土壤改良(Soil Enrichment)使每個農業人口單位
	// 額外 +1 食物產出(GAME_MANUAL.pdf:"This 'fertilization' process increases the food output
	// of each farming unit of population by 1.")。
	TerraformSoilEnrichmentFoodBonusPerFarmer = 1
)

// terraformSoilEnrichmentBlockedClimates 土壤改良在這些氣候上無效(GAME_MANUAL.pdf:"Soil
// Enrichment does not work in hostile climates. Barren worlds have no topsoil to work on, while
// ongoing chemical processes in the soils of Radiated and Toxic planets undo the fertilization
// as fast as it is done.")。
var terraformSoilEnrichmentBlockedClimates = map[PlanetClimate]bool{
	BARREN:   true,
	RADIATED: true,
	TOXIC:    true,
}

// TerraformSoilEnrichmentWorks 回傳土壤改良在指定氣候下是否有效。
func TerraformSoilEnrichmentWorks(climate PlanetClimate) bool {
	return !terraformSoilEnrichmentBlockedClimates[climate]
}

// terraformNextClimate 地形改造(Terraforming 特殊建設)把星球氣候往 Terran 方向推進一級
// (GAME_MANUAL.pdf:"Terraforming will only work on planets that have hospitable environments
// already. Barren worlds become Desert or Tundra, Desert environments become Arid, Tundra
// planets become Swamp worlds, and Ocean, Arid, and Swamp become Terran. You can terraform a
// planet several times, but each application has an increased production cost.")。
//
// 手冊原文對 Barren 的下一級給了兩個選項(Desert 或 Tundra),未說明選擇條件,故 BARREN 對應
// 兩個候選值;其餘氣候手冊只給單一結果。TERRAN(手冊未提及 Terraforming 可再把 Terran 往前推,
// 那是 Gaia Transformation 的範圍,見下方)、TOXIC(見 TerraformToxicNextClimate)、RADIATED、
// GAIA 手冊未列入這條 Terraforming 轉換鏈,故不在此表中。
var terraformNextClimate = map[PlanetClimate][]PlanetClimate{
	BARREN: {DESERT, TUNDRA},
	DESERT: {ARID},
	TUNDRA: {SWAMP},
	OCEAN:  {TERRAN},
	ARID:   {TERRAN},
	SWAMP:  {TERRAN},
}

// TerraformNextClimateOptions 回傳地形改造後可能的下一級氣候候選清單;若手冊未定義該氣候的
// 地形改造規則(例如 TERRAN、GAIA、RADIATED,或該氣候已是鏈中終點),回傳空 slice。
func TerraformNextClimateOptions(climate PlanetClimate) []PlanetClimate {
	return terraformNextClimate[climate]
}

// TerraformToxicNextClimate 1.50i 修改版新增規則:Toxic 星球可透過建造 Recyclotron(再生反應爐)
// 轉換成 Barren 氣候(MANUAL_150.html:"In 150i, Toxic worlds can be transformed to Barren
// climates by building a Recyclotron." 及 Modding 附錄:"the building can be specified that will
// cause a Toxic climate to convert to a Barren one. The 150 improved mod, specifies Recyclotron.
// Note: Specifying Terraforming or Gaia Transformation will not work for this parameter, since
// both cannot be built on Toxic planets.")。此規則走 BUILDING_RECYCLOTRON,不是一般的
// Terraforming/Gaia Transformation 特殊建設,故獨立於 terraformNextClimate 表。
const TerraformToxicNextClimate = BARREN

const (
	// GaiaTransformationSourceClimate 蓋亞轉化(Gaia Transformation)只能套用在 Terran 星球上
	// (GAME_MANUAL.pdf:"The transformation can only be applied to Terran environments.")。
	GaiaTransformationSourceClimate = TERRAN
	// GaiaTransformationResultClimate 蓋亞轉化完成後星球變成 Gaia 級(GAME_MANUAL.pdf:
	// "Afterward, the planet becomes a Gaia class world.")。
	GaiaTransformationResultClimate = GAIA
)

// GaiaTransformationCanApply 回傳蓋亞轉化是否可套用於指定氣候(僅 Terran)。
func GaiaTransformationCanApply(climate PlanetClimate) bool {
	return climate == GaiaTransformationSourceClimate
}

// terraformClimatePopFactor 星球氣候對人口容量的百分比係數(0-100)。量化地形改造沿氣候鏈
// (terraformNextClimate)推進到 Terran、再靠 Gaia Transformation 推進到 Gaia 時,人口容量
// 同步提升多少 —— 這是玩家投入地形改造成本的主要回報。數值來自 MANUAL_150.html 的 Modding
// 附錄("Population Capacities":"pop_climate = 25 25 25 25 25 25 40 60 80 100"),順序見本檔
// 開頭的交叉驗證說明。
var terraformClimatePopFactor = [10]int{
	25,  // TOXIC
	25,  // RADIATED
	25,  // BARREN
	25,  // DESERT
	25,  // TUNDRA
	25,  // OCEAN
	40,  // SWAMP
	60,  // ARID
	80,  // TERRAN
	100, // GAIA
}

// TerraformClimatePopFactorPercent 回傳指定氣候的人口容量係數(0-100);climate 超出範圍回 0。
func TerraformClimatePopFactorPercent(climate PlanetClimate) int {
	if climate < 0 || int(climate) >= len(terraformClimatePopFactor) {
		return 0
	}
	return terraformClimatePopFactor[climate]
}

// TerraformPopMaxAfterClimateChange 依氣候變動前後的 pop_climate 百分比係數(見
// terraformClimatePopFactor),對目前 PopMax 做等比例縮放,回傳地形改造/蓋亞轉化套用後的新 PopMax。
//
// 誠實近似聲明:MANUAL_150.html modding 附錄把 pop_climate 定義成「星球人口容量的百分比係數」,
// 字面意思應該是「行星尺寸決定的基礎人口容量 × pop_climate% = PopMax」。但本 remake 的
// engine.ColonyState 沒有獨立追蹤「行星尺寸 → 基礎人口容量」這個中介值——PopMax 是直接烘進的
// 整數,可能已經疊加了生態圈(Biospheres p.99 +2)等其他建築的固定加成(見
// docs/tech/colony-buildings.md §6.1)。因此本函式改用「目前 PopMax 整體乘上新舊係數比例」近似,
// 而不是精確重算「基礎值 × 新係數」——這代表已疊加的固定加成也會跟著等比例縮放,不是官方精確
// 公式,是誠實記錄的近似值。TODO:待補「行星尺寸→基礎人口容量」對映表(手冊/資料檔尚未查到)後,
// 可回頭重算成精確版本。
//
// oldClimate 對應係數 <=0(超出 PlanetClimate 合法範圍)時視為無法換算,直接回傳原 PopMax 不動,
// 避免除以零、也避免把 PopMax 錯誤歸零。
func TerraformPopMaxAfterClimateChange(currentPopMax int, oldClimate, newClimate PlanetClimate) int {
	oldFactor := TerraformClimatePopFactorPercent(oldClimate)
	if oldFactor <= 0 {
		return currentPopMax
	}
	newFactor := TerraformClimatePopFactorPercent(newClimate)
	return currentPopMax * newFactor / oldFactor
}
