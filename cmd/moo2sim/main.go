// moo2sim 是 headless 回合模擬器:用 internal/engine 連跑數回合並印報告,
// 展示回合引擎(殖民經濟 + 研究 + 國庫)端到端運作。純 Go,不依賴 ebiten。
//
// 用法:
//
//	moo2sim [-turns 10]
//
// 目前用內建合成帝國(2 殖民地)示範;未來可加 -save 載入真實存檔。
package main

import (
	"flag"
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func main() {
	turns := flag.Int("turns", 10, "模擬回合數")
	flag.Parse()

	// 合成帝國:兩個殖民地(一大一小),稅率 40%,起始研究 TOPIC 1(成本 400)。
	colonies := []engine.ColonyState{
		{Population: 8, PopMax: 20, Farmers: 3, Workers: 4, Scientists: 1,
			FoodPerFarmer: 4, IndustryPerWorker: 6, ResearchPerScientist: 30,
			PlanetSize: gamedata.LARGE_PLANET, MoralePercent: 10},
		{Population: 4, PopMax: 12, Farmers: 2, Workers: 1, Scientists: 1,
			FoodPerFarmer: 4, IndustryPerWorker: 5, ResearchPerScientist: 20,
			PlanetSize: gamedata.SMALL_PLANET},
	}
	ps := engine.PlayerState{
		BC: 100, TaxRate: 40, Maintenance: 5,
		ResearchTopic: gamedata.ResearchTopic(1),
	}

	fmt.Printf("=== MOO2 回合模擬(%d 回合)===\n", *turns)
	fmt.Printf("%-4s %-8s %-8s %-8s %-10s %-8s\n", "回合", "食物", "淨工業", "研究", "研究進度", "國庫BC")
	for t := 1; t <= *turns; t++ {
		out := engine.RunEmpireTurn(ps, colonies)
		ps = out.Player
		done := ""
		if out.ResearchDone {
			done = "  ← 研究完成!"
		}
		fmt.Printf("%-4d %-8d %-8d %-8d %-10d %-8d%s\n",
			t, out.TotalFood, out.TotalNetIndustry, out.TotalResearch,
			ps.ResearchProgress, ps.BC, done)
	}
	fmt.Printf("\n完成研究主題數:%d\n", len(ps.CompletedTopics))
}
