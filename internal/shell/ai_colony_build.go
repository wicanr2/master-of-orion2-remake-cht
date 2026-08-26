package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// aiColonyBuildKey 以星系索引保存佇列；舊存檔若缺 ColonyStars，使用不會和合法星系
// 索引衝突的負值，待對映恢復後自然建立正式 key。
func aiColonyBuildKey(a *AIOpponent, colony int) int {
	if colony >= 0 && colony < len(a.ColonyStars) && a.ColonyStars[colony] >= 0 {
		return a.ColonyStars[colony]
	}
	return -colony - 1
}

// aiBuildingScore 是原版 Colony_Building_Score_ 的 typed 轉接層。原版 47 個 case 的
// 完整欄位語意尚未全部映射；此處只使用 remake 已有的殖民地輸出與建築分類。加權抽選、
// 難度濾門、候選範圍與逐殖民地產品資料形狀則直接依 IDA 證據實作。
func aiBuildingScore(b gamedata.Building, colony engine.ColonyState, out engine.ColonyOutput) int {
	score := 20
	switch b.Category {
	case gamedata.CategoryProduction, gamedata.CategoryEnvironment:
		score += 80
	case gamedata.CategoryFood:
		score += 45
		if out.FoodSurplusHalf <= 0 {
			score += 80
		}
	case gamedata.CategoryResearch:
		score += 55
	case gamedata.CategoryDefense, gamedata.CategorySatellite, gamedata.CategoryMilitary:
		score += 40
	case gamedata.CategoryHousing:
		if colony.Population >= colony.PopMax-2 {
			score += 70
		}
	case gamedata.CategoryTrade, gamedata.CategoryMorale, gamedata.CategorySociety:
		score += 30
	}
	if score > 1000 {
		return 1000
	}
	return score
}

func chooseAIColonyBuilding(a *AIOpponent, colony int, out engine.ColonyOutput, difficulty, turn int) (ColonyBuild, bool) {
	if colony < 0 || colony >= len(a.Colonies) {
		return ColonyBuild{}, false
	}
	built := map[string]bool(nil)
	if colony < len(a.ColonyBuildings) {
		built = a.ColonyBuildings[colony]
	}
	type candidate struct {
		build ColonyBuild
		score int
	}
	var candidates []candidate
	maxScore := 1 // raw Assign_Colony_New_Building_ 也把最大分數下限夾到 1。
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		if built[b.NameZH] {
			continue
		}
		score := aiBuildingScore(b, a.Colonies[colony], out)
		if score > maxScore {
			maxScore = score
		}
		candidates = append(candidates, candidate{ColonyBuild{Name: b.NameZH, Cost: b.ProductionCost}, score})
	}
	if len(candidates) == 0 {
		return ColonyBuild{}, false
	}
	// 0xD0BC5..0xD0BD9：(6-difficulty)*score < maxScore 的候選歸零。
	if difficulty < 0 {
		difficulty = 0
	}
	if difficulty > 4 {
		difficulty = 4
	}
	total := 0
	for i := range candidates {
		if (6-difficulty)*candidates[i].score < maxScore {
			candidates[i].score = 0
		}
		total += candidates[i].score
	}
	if total <= 0 {
		return ColonyBuild{}, false
	}
	// 原版呼叫共用 PRNG；remake 尚未逐位元重現該全局亂數流，故以可存檔狀態導出的
	// 決定性取樣作明示近似，避免讀檔後換產品。
	key := aiColonyBuildKey(a, colony)
	pick := int((uint64(turn+1)*0x9E3779B185EBCA87 + uint64(key+257)*0xC2B2AE3D27D4EB4F) % uint64(total))
	for _, c := range candidates {
		if pick < c.score {
			return c.build, true
		}
		pick -= c.score
	}
	return candidates[len(candidates)-1].build, true
}

// applyAICompletedBuilding 把已完成建築接到 AI 殖民地的主要經濟欄位。建築 map 仍是
// 防禦、維護與駐軍等其他消費端的權威；這裡只處理 ColonyState 內的累積型產出欄位。
func applyAICompletedBuilding(c *engine.ColonyState, name string) {
	switch name {
	case "自動工廠":
		c.IndustryPerWorker++
		c.FlatIndustry += 5
	case "研究實驗室":
		c.FlatResearch += 5
	case "太空港":
		c.IncomeBonusPercent += 50
	case "機器人採礦廠":
		c.IndustryPerWorker += 2
		c.FlatIndustry += 10
	case "深層核心礦場":
		c.IndustryPerWorker += 3
		c.FlatIndustry += 15
	case "污染處理器":
		c.PollutionProcessor = true
	case "大氣更新器":
		c.AtmosphericRenewer = true
	case "核心廢料場":
		c.CoreWasteDump = true
	case "行星超級電腦":
		c.FlatResearch += 10
	case "銀河網路中心":
		c.FlatResearch += 15
	case "水耕農場":
		c.FlatFood += 2
	case "地底農場":
		c.FlatFood += 4
	case "氣候控制器":
		c.FoodPerFarmer += 2
	case "行星證券交易所":
		c.IncomeBonusPercent += 100
	case "太空大學":
		c.FoodPerFarmer++
		c.IndustryPerWorker++
		c.ResearchPerScientist++
	case "生態圈":
		c.PopMax += 2
	case "複製中心":
		c.FlatGrowth += gamedata.CloningCenterGrowthPoints
	case "自動實驗室":
		c.FlatResearch += 30
	case "食物複製機":
		c.FoodReplicators = true
	case "再生反應爐":
		c.Recyclotron = true
	case "行星重力產生器":
		c.NormalizeGravity = true
	case "機器人工廠":
		c.FlatIndustry += gamedata.ProdRoboticFactoryBonus(int(c.MineralRichness))
	}
}

// advanceAIColonyBuilds 逐殖民地消費本回合淨工業，回傳沒有可建建築之殖民地的產能，
// 供尚待進一步 RE 的艦艇產品轉接層使用。同一份產能不會同時蓋建築又造艦。
func (s *GameSession) advanceAIColonyBuilds(aiIndex int, out engine.EmpireOutput) int {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return 0
	}
	a := &s.AIPlayers[aiIndex]
	if a.ColonyBuilds == nil {
		a.ColonyBuilds = make(map[int]ColonyBuild)
	}
	shipProduction := 0
	for i := range a.Colonies {
		if i >= len(out.Colonies) || out.Colonies[i].NetIndustry <= 0 {
			continue
		}
		key := aiColonyBuildKey(a, i)
		build := a.ColonyBuilds[key]
		if build.Name == "" {
			var ok bool
			build, ok = chooseAIColonyBuilding(a, i, out.Colonies[i], s.Difficulty, s.Turn)
			if !ok {
				shipProduction += out.Colonies[i].NetIndustry
				continue
			}
		}
		build.Progress += out.Colonies[i].NetIndustry
		if build.Progress < build.Cost {
			a.ColonyBuilds[key] = build
			continue
		}
		for len(a.ColonyBuildings) <= i {
			a.ColonyBuildings = append(a.ColonyBuildings, nil)
		}
		if a.ColonyBuildings[i] == nil {
			a.ColonyBuildings[i] = make(map[string]bool)
		}
		if !a.ColonyBuildings[i][build.Name] {
			a.ColonyBuildings[i][build.Name] = true
			applyAICompletedBuilding(&a.Colonies[i], build.Name)
		}
		delete(a.ColonyBuilds, key)
	}
	return shipProduction
}
