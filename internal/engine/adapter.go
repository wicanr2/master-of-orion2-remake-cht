package engine

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

// adapter.go 是 save(存檔二進位表示)↔ engine(乾淨回合狀態)的橋接。
// 核心階段(colony.go/empire.go)不 import save、維持可純測;只有本檔接觸 save 型別。
//
// v1 對映範圍:人口/工作分配、每單位產出率、人口上限、行星尺寸、污染。
// 尚未對映(需更多存檔解析,標明避免誤以為完整):建築旗標(污染處理器/大氣更新器/核心廢料場)、
// 種族特性(Tolerant)、成長獎金(科技/政體/AI)。這些在 ColonyState 對應欄位暫留零值。

// ColonyStateFromSave 把存檔的 Colony(需另給其所在 Planet 以取得尺寸)轉成引擎 ColonyState。
// 依 Colonist.Job(0 農夫/1 工人/2 科學家)統計前 Population 名殖民者的工作分配。
func ColonyStateFromSave(c *save.Colony, planet *save.Planet) ColonyState {
	pop := int(c.Population)
	if pop > len(c.Colonists) {
		pop = len(c.Colonists)
	}
	var farmers, workers, scientists int
	for i := 0; i < pop; i++ {
		switch gamedata.ColonistJob(c.Colonists[i].Job) {
		case gamedata.FARMER:
			farmers++
		case gamedata.WORKER:
			workers++
		case gamedata.SCIENTIST:
			scientists++
		}
	}
	return ColonyState{
		Population:           int(c.Population),
		PopMax:               int(c.MaxPopulation),
		Farmers:              farmers,
		Workers:              workers,
		Scientists:           scientists,
		FoodPerFarmer:        int(c.FoodPerFarmer),
		IndustryPerWorker:    int(c.IndustryPerWorker),
		ResearchPerScientist: int(c.ResearchPerScientist),
		PlanetSize:           gamedata.PlanetSize(planet.Size),
		// 建築旗標/Tolerant/成長獎金 v1 未對映(見檔頭說明),留零值。
	}
}

// PlayerStateFromSave 把存檔的 Player 轉成引擎 PlayerState。
// CompletedTopics 由 ResearchTopics 陣列的狀態推導留待後續(其編碼未查證),v1 留 nil
// (RunResearchPhase 對 nil map 安全)。
func PlayerStateFromSave(p *save.Player) PlayerState {
	return PlayerState{
		BC:               int(p.BC),
		TaxRate:          int(p.TaxRate),
		ResearchTopic:    gamedata.ResearchTopic(p.ResearchTopic),
		ResearchProgress: int(p.ResearchProgress),
	}
}
