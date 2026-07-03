// moo2sim 是 headless 回合模擬器:用 internal/engine 連跑數回合並印報告,
// 展示回合引擎(殖民經濟 + 研究 + 國庫)端到端運作。純 Go,不依賴 ebiten。
//
// 用法:
//
//	moo2sim [-turns 10]
//	moo2sim -save save1.gam
//	moo2sim -ai aggressive [-turns 10]
//
// 未給 -save/-ai 時用內建合成帝國(2 殖民地、固定工作分配)連跑 -turns 回合示範;
// 給 -save 時載入真實存檔,只跑一回合並逐玩家印報告(不回寫存檔);
// 給 -ai <profile> 時用內建合成帝國連跑 -turns 回合,但工作分配與稅率交給
// internal/ai(設計性重建的 AI 決策層)每回合自行決定,展示 AI 對手如何管理經濟。
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

func main() {
	turns := flag.Int("turns", 10, "模擬回合數(僅合成帝國模式適用)")
	savePath := flag.String("save", "", "載入真實 MOO2 存檔(.GAM)並跑一回合,逐玩家印報告")
	aiName := flag.String("ai", "", "由 AI 自行管理合成帝國經濟,值:aggressive/scientific/balanced/expansionist(留空=不啟用,維持固定工作分配)")
	flag.Parse()

	if *savePath != "" {
		runFromSave(*savePath)
		return
	}
	if *aiName != "" {
		runAI(*turns, *aiName)
		return
	}
	runSynthetic(*turns)
}

// resolveAIProfile 把 -ai 旗標的名稱字串對映到 internal/ai 的 Profile。
func resolveAIProfile(name string) (ai.Profile, bool) {
	switch name {
	case "aggressive":
		return ai.ProfileAggressive, true
	case "scientific":
		return ai.ProfileScientific, true
	case "balanced":
		return ai.ProfileBalanced, true
	case "expansionist":
		return ai.ProfileExpansionist, true
	default:
		return ai.Profile{}, false
	}
}

// runAI 用內建合成帝國(2 殖民地,工作分配留給 AI)連跑 turns 回合,
// 每回合先用 engine.ApplyAIEconomy 讓 AI(依 profile)分配農夫/工人/科學家與稅率,
// 再用 engine.RunEmpireTurn 結算,印出報告與每回合 AI 的工作分配,展示 AI 對手如何運作。
func runAI(turns int, name string) {
	profile, ok := resolveAIProfile(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "未知 AI 性格:%q(可用值:aggressive/scientific/balanced/expansionist)\n", name)
		os.Exit(1)
	}

	// 合成帝國:兩個殖民地(一大一小),不預設工作分配(Farmers/Workers/Scientists=0),
	// 交給 AI 每回合自行決定;人口/產出率設定同 runSynthetic 以便對照。
	colonies := []engine.ColonyState{
		{Population: 8, PopMax: 20,
			FoodPerFarmer: 4, IndustryPerWorker: 6, ResearchPerScientist: 30,
			PlanetSize: gamedata.LARGE_PLANET, MoralePercent: 10},
		{Population: 4, PopMax: 12,
			FoodPerFarmer: 4, IndustryPerWorker: 5, ResearchPerScientist: 20,
			PlanetSize: gamedata.SMALL_PLANET},
	}
	ps := engine.PlayerState{
		BC: 100, TaxRate: 40, Maintenance: 5,
		ResearchTopic: gamedata.ResearchTopic(1),
	}

	fmt.Printf("=== MOO2 AI 帝國模擬(%d 回合)===\n", turns)
	fmt.Printf("AI 性格:%s(工業權重=%d 研究權重=%d)\n\n", profile.Name, profile.IndustryWeight, profile.ResearchWeight)
	fmt.Printf("%-4s %-8s %-8s %-8s %-10s %-8s %-6s\n", "回合", "食物", "淨工業", "研究", "研究進度", "國庫BC", "稅率%")
	for t := 1; t <= turns; t++ {
		decidedPS, decidedColonies := engine.ApplyAIEconomy(ps, colonies, profile)
		out := engine.RunEmpireTurn(decidedPS, decidedColonies)
		ps = out.Player
		done := ""
		if out.ResearchDone {
			done = "  ← 研究完成!"
		}
		fmt.Printf("%-4d %-8d %-8d %-8d %-10d %-8d %-6d%s\n",
			t, out.TotalFood, out.TotalNetIndustry, out.TotalResearch,
			ps.ResearchProgress, ps.BC, decidedPS.TaxRate, done)
		for i, cs := range decidedColonies {
			fmt.Printf("      殖民地%d: 人口=%d 農夫=%d 工人=%d 科學家=%d\n",
				i, cs.Population, cs.Farmers, cs.Workers, cs.Scientists)
		}
	}
	fmt.Printf("\n完成研究主題數:%d\n", len(ps.CompletedTopics))
}

// runFromSave 載入一份真實存檔,用 engine.RunGameTurn 跑一回合並逐玩家印報告。
func runFromSave(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "讀取存檔失敗: %v\n", err)
		os.Exit(1)
	}
	gs, err := save.Load(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析存檔失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== MOO2 存檔模擬:%q(stardate %d)===\n", gs.Config.SaveGameName, gs.Config.Stardate)
	fmt.Printf("玩家數=%d 殖民地數=%d 星系數=%d\n\n", gs.PlayerCount, gs.ColonyCount, gs.StarCount)

	result := engine.RunGameTurn(gs)

	fmt.Printf("%-4s %-8s %-8s %-8s %-8s %-8s %-8s\n",
		"玩家", "殖民地", "總食物", "淨工業", "研究", "稅收", "研究完成")
	for p := 0; p < gs.PlayerCount; p++ {
		out, ok := result.PlayerOutputs[p]
		if !ok {
			continue
		}
		done := ""
		if out.ResearchDone {
			done = "是"
		}
		fmt.Printf("%-4d %-8d %-8d %-8d %-8d %-8d %-8s\n",
			p, len(out.Colonies), out.TotalFood, out.TotalNetIndustry,
			out.TotalResearch, out.TaxRevenue, done)
	}
}

// runSynthetic 用內建合成帝國(2 殖民地)連跑 turns 回合示範。
func runSynthetic(turns int) {
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

	fmt.Printf("=== MOO2 回合模擬(%d 回合)===\n", turns)
	fmt.Printf("%-4s %-8s %-8s %-8s %-10s %-8s\n", "回合", "食物", "淨工業", "研究", "研究進度", "國庫BC")
	for t := 1; t <= turns; t++ {
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
